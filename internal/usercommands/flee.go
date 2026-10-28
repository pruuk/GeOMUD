package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Flee(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if actor.Character.Aggro == nil || actor.Character.Aggro.Type != characters.Flee {
		tell(actor, `You attempt to flee...`)

		actor.Character.Aggro = &characters.Aggro{}
		actor.Character.Aggro.Type = characters.Flee
		events.AddToQueue(events.AggroChanged{ActorId: actor.Character.Id, RoomId: actor.Character.RoomId})
	}

	return true, nil
}
