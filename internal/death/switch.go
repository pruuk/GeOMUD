package death

import (
	"sort"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
)

// NextOpponent is the body a turns to when its target is gone (Phase 5
// slice 4.1, ruling 7): of the bodies in a's room that are up, that a
// perceives, that the ruleset lets a attack (rules.CanAttack), and whose
// aggro targets a, the one whose aggro began earliest (Aggro.Since); ties go
// to turn order. Nil when there is none.
func NextOpponent(a *actors.Actor) *actors.Actor {
	r := rooms.LoadRoom(a.Character.RoomId)
	if r == nil {
		return nil
	}
	candidates := []*actors.Actor{}
	for _, id := range r.GetBodies() {
		b := actors.Get(id)
		if b == nil || b.Id == a.Id || b.Character == nil || b.Character.Aggro == nil {
			continue
		}
		if b.Character.Aggro.Target != a.Id {
			continue // the room's bodies are all in a's room; no RoomId check (review round 5)
		}
		if rules.Get().Condition(b) != rules.Up || !rooms.Perceives(a.Character, b.Character) {
			continue
		}
		if rules.Get().CanAttack(a, b) != nil {
			continue // the ruleset refuses it (two players where PvP is off)
		}
		candidates = append(candidates, b)
	}
	var next *actors.Actor
	for _, b := range rules.Get().TurnOrder(SortBodies(candidates)) {
		if next == nil || b.Character.Aggro.Since < next.Character.Aggro.Since {
			next = b
		}
	}
	return next
}

// SortBodies orders bodies as the round lists them before the ruleset's
// turn order (Phase 5 slice 4.1): players by UserId, then NPCs by MobId,
// then any other body; ties by body ID text. It sorts in place and returns
// the slice.
func SortBodies(bodies []*actors.Actor) []*actors.Actor {
	type keyed struct {
		body    *actors.Actor
		kind, n int
		id      string
	}
	list := make([]keyed, len(bodies))
	for i, a := range bodies {
		k := keyed{body: a, kind: 2, id: a.Id.String()}
		if acct := actors.AccountOf(a.Id); acct != nil {
			k.kind, k.n = 0, acct.UserId
		} else if m := mobs.GetInstance(a.Id); m != nil {
			k.kind, k.n = 1, int(m.MobId)
		}
		list[i] = k
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].kind != list[j].kind {
			return list[i].kind < list[j].kind
		}
		if list[i].n != list[j].n {
			return list[i].n < list[j].n
		}
		return list[i].id < list[j].id
	})
	for i := range list {
		bodies[i] = list[i].body
	}
	return bodies
}

// Switch turns a to its next opponent (NextOpponent), with the line naming
// it, and reports whether there was one. With none, a's aggro is left as it
// is: the caller ends it with the line for its case.
func Switch(a *actors.Actor) bool {
	b := NextOpponent(a)
	if b == nil {
		return false
	}
	a.Character.SetAggro(b.Id, characters.DefaultAttack)
	messaging.Send(perception.Message{
		Kind:      perception.WorldEvent,
		Actor:     a.Id,
		Target:    b.Id,
		Audience:  perception.ActorOnly,
		ActorText: `You turn to fight {target}!`,
	})
	events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
	return true
}

// releaseAttackers ends the fight of every body, player or NPC, whose aggro
// targets the victim, and turns each to its next opponent at once. Bodies
// in the victim's room go first, in the room's order, then any elsewhere
// (a shooter, a body waiting on the victim) in actor order.
func releaseAttackers(victim *actors.Actor) {
	seen := map[actors.ActorId]bool{}
	order := []*actors.Actor{}
	if r := rooms.LoadRoom(victim.Character.RoomId); r != nil {
		for _, id := range r.GetBodies() {
			if b := actors.Get(id); b != nil {
				seen[b.Id] = true
				order = append(order, b)
			}
		}
	}
	for _, b := range actors.All() {
		if !seen[b.Id] {
			order = append(order, b)
		}
	}
	for _, b := range order {
		if b.Id == victim.Id || b.Character == nil || b.Character.Aggro == nil || b.Character.Aggro.Target != victim.Id {
			continue
		}
		if Switch(b) {
			continue
		}
		b.Character.EndAggro()
		events.AddToQueue(events.AggroChanged{ActorId: b.Id, RoomId: b.Character.RoomId})
	}
}
