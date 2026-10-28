package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Save(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	Tell(actor, "Saving...")
	accounts.SaveUser(*user)
	Tell(actor, "done.")

	return true, nil
}
