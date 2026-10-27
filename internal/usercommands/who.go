package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

func Who(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	details := rooms.GetDetails(room, user)

	whoTxt, _ := templates.Process("descriptions/who", details, user.UserId)
	user.SendText(whoTxt)

	return true, nil
}
