package usercommands

import (
	"errors"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// Suicide is a death by the actor's own hand, through the one death routine
// (internal/death), for players and NPCs alike. It is refused in a fight
// (nobody escapes one except by fleeing, as quit is refused) and in the
// recovery room or a copy of it.
func Suicide(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	if inFight(actor, room) {
		Tell(actor, "You're too busy to die right now!")
		return true, nil
	}

	if rooms.IsDeathRecoveryRoom(actor.Character.RoomId) {
		Tell(actor, `You're already dead!`)
		return true, errors.New(`already dead`)
	}

	death.Die(actor)
	return true, nil
}
