package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

func Exits(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	details := rooms.GetDetails(room, user)

	exitTxt, _ := templates.Process("descriptions/exits", details, user.UserId)
	user.SendText(exitTxt)

	return true, nil
}
