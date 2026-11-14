package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Flee(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if actor.Character.Aggro == nil || actor.Character.Aggro.Type != characters.Flee {
		Tell(actor, `You attempt to flee...`)

		// Keep the target fled from, and when that fight began: a blocked or
		// exitless flee resumes it in its place (Phase 5 slice 4.1).
		fleeFrom := ids.ID{}
		var since uint64
		if actor.Character.Aggro != nil {
			fleeFrom = actor.Character.Aggro.Target
			since = actor.Character.Aggro.Since
		}

		actor.Character.Aggro = &characters.Aggro{Type: characters.Flee, FleeFrom: fleeFrom, Since: since}
		events.AddToQueue(events.AggroChanged{ActorId: actor.Character.Id, RoomId: actor.Character.RoomId})
	}

	return true, nil
}
