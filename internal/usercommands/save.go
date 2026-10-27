package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Save(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	user.SendText("Saving...")
	accounts.SaveUser(*user)
	user.SendText("done.")

	return true, nil
}
