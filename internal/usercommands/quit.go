package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Quit(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	// Nobody escapes a fight except by fleeing (Decision 4); quitting is an escape.
	if inFight(actor, room) {
		user.SendText("You're too busy to quit right now!")
		return true, nil
	}
	user.AddBuff(0, `quitting`)

	return true, nil
}
