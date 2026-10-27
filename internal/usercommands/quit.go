package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Quit(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user.Character.Aggro != nil {
		user.SendText("You're too busy to quit right now!")
		return true, nil
	}
	user.AddBuff(0, `quitting`)

	return true, nil
}
