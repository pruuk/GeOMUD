// Package messaging delivers messages (Phase 4b spec section 2): it finds
// who a message reaches from its range and audience, asks CanPerceive for
// each receiver, picks the receiver's role and text with tokens resolved
// for them, and hands the result to the receiver's controller. The
// vocabulary is in internal/perception.
package messaging

import (
	"sort"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// receiver is one body a message reaches, with where it stands.
type receiver struct {
	actor        *actors.Actor
	room         *rooms.Room
	far          bool
	direction    string
	inTargetRoom bool // in the target's room, which is not the actor's room
}

// Send delivers m to everyone it reaches. It must run on the world
// goroutine: room maps are not locked. Delivery is synchronous, so a
// Perceive that calls Send re-enters it.
func Send(m perception.Message) {
	actor := actors.Get(m.Actor)
	target := actors.Get(m.Target)

	if m.Kind == perception.WorldEvent && m.Audience != perception.ActorOnly && m.Senses == 0 {
		mudlog.Error("messaging", "error", "message has no senses", "actorId", m.Actor)
	}

	if m.Kind == perception.Interface || m.Audience == perception.ActorOnly {
		if actor != nil && !excluded(m, actor.Id) {
			deliver(m, actor, target, receiver{actor: actor, room: roomOfActor(actor)}, perception.Actor, 0)
		}
		return
	}

	for _, rc := range receiversOf(m, actor, target) {
		if excluded(m, rc.actor.Id) {
			continue
		}
		role := perception.Observer
		switch {
		case actor != nil && rc.actor.Id == actor.Id:
			deliver(m, actor, target, rc, perception.Actor, 0)
			continue
		case target != nil && rc.actor.Id == target.Id:
			role = perception.Target
		}
		if m.Audience == perception.Named && !named(m, rc.actor.Id) {
			continue
		}
		if role == perception.Observer && rc.inTargetRoom && !rc.far && m.TargetRoomText == `` && m.NoTargetRoomText {
			continue
		}
		sense := perceivedBy(rc.actor, rc.room, m)
		if sense == 0 {
			continue
		}
		deliver(m, actor, target, rc, role, sense)
	}
}

func deliver(m perception.Message, actor, target *actors.Actor, rc receiver, role perception.Role, sense perception.Sense) {
	p := perception.Perception{
		Role:       role,
		Message:    m,
		SenseUsed:  sense,
		ActorSeen:  canSee(rc.actor, rc.room, actor, rc.far),
		TargetSeen: canSee(rc.actor, rc.room, target, rc.far),
		Far:        rc.far,
		Direction:  rc.direction,
	}
	if m.Kind == perception.Interface {
		// Interface text is for the actor alone and is written as it is:
		// admin help that shows "{{ .X }}" or a literal "{actor}" is untouched.
		p.Text = textFor(m, role, rc)
	} else {
		p.Text = resolveTokens(textFor(m, role, rc),
			nameFor(actor, p.ActorSeen), nameFor(target, p.TargetSeen), rc.direction)
	}
	if ctrl := rc.actor.Controller(); ctrl != nil {
		ctrl.Perceive(p)
	}
}

// textFor picks the text for a role (decision 6 of the slice 3 plan).
func textFor(m perception.Message, role perception.Role, rc receiver) string {
	switch {
	case role == perception.Actor:
		return m.ActorText
	case role == perception.Target && m.TargetText != ``:
		return m.TargetText
	case role == perception.Observer && rc.inTargetRoom && !rc.far && m.TargetRoomText != ``:
		return m.TargetRoomText
	case rc.far && m.FarText != ``:
		return m.FarText
	}
	return m.ObserverText
}

func named(m perception.Message, id ids.ID) bool {
	for _, n := range m.Named {
		if n == id {
			return true
		}
	}
	return false
}

// excluded reports whether id is in m.Exclude.
func excluded(m perception.Message, id ids.ID) bool {
	for _, x := range m.Exclude {
		if x == id {
			return true
		}
	}
	return false
}

func roomOfActor(a *actors.Actor) *rooms.Room {
	if a == nil || a.Character == nil || !rooms.IsRoomLoaded(a.Character.RoomId) {
		return nil
	}
	return rooms.LoadRoom(a.Character.RoomId)
}

// sourceRooms are the actor's room and the target's room when it differs,
// or the message's own room for a world event without an actor.
func sourceRooms(m perception.Message, actor, target *actors.Actor) []*rooms.Room {
	out := []*rooms.Room{}
	add := func(r *rooms.Room) {
		if r == nil {
			return
		}
		for _, have := range out {
			if have.RoomId == r.RoomId {
				return
			}
		}
		out = append(out, r)
	}
	if actor != nil {
		add(roomOfActor(actor))
	} else if m.Room != 0 && rooms.IsRoomLoaded(m.Room) {
		add(rooms.LoadRoom(m.Room))
	}
	add(roomOfActor(target))
	return out
}

// receiversOf lists who m's range reaches, source rooms first, each body once.
// The actor is included even when it is in no room.
func receiversOf(m perception.Message, actor, target *actors.Actor) []receiver {
	out := []receiver{}
	seen := map[ids.ID]bool{}
	add := func(a *actors.Actor, r *rooms.Room, far bool, direction string, inTargetRoom bool) {
		if a == nil || seen[a.Id] {
			return
		}
		seen[a.Id] = true
		out = append(out, receiver{actor: a, room: r, far: far, direction: direction, inTargetRoom: inTargetRoom})
	}

	src := sourceRooms(m, actor, target)
	inSource := map[int]bool{}
	targetRoom := roomOfActor(target)
	for _, r := range src {
		inSource[r.RoomId] = true
		// the first source room is the actor's (or the world event's)
		inTarget := targetRoom != nil && r.RoomId == targetRoom.RoomId && r.RoomId != src[0].RoomId
		for _, id := range r.GetBodies() {
			add(actors.Get(id), r, false, ``, inTarget)
		}
	}
	if actor != nil {
		add(actor, roomOfActor(actor), false, ``, false)
	}

	switch m.Range {
	case perception.Nearby:
		for _, r := range src {
			for _, nId := range neighbourIds(r) {
				if inSource[nId] || !rooms.IsRoomLoaded(nId) {
					continue
				}
				n := rooms.LoadRoom(nId)
				back := n.FindExitTo(r.RoomId)
				if back == `` {
					continue
				}
				for _, id := range n.GetBodies() {
					add(actors.Get(id), n, true, back, false)
				}
			}
		}
	case perception.Zone, perception.World:
		zones := map[string]bool{}
		for _, r := range src {
			zones[r.Zone] = true
		}
		for _, a := range actors.All() {
			r := roomOfActor(a)
			if r == nil || inSource[r.RoomId] {
				continue
			}
			if m.Range == perception.Zone && !zones[r.Zone] {
				continue
			}
			add(a, r, true, ``, false)
		}
	}
	return out
}

// neighbourIds are the rooms r leads to through its exits, temporary exits
// and active mutator exits, in room ID order.
func neighbourIds(r *rooms.Room) []int {
	set := map[int]bool{}
	for _, e := range r.Exits {
		set[e.RoomId] = true
	}
	for _, e := range r.ExitsTemp {
		set[e.RoomId] = true
	}
	for mut := range r.ActiveMutators {
		for _, e := range mut.GetSpec().Exits {
			set[e.RoomId] = true
		}
	}
	out := make([]int, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}
