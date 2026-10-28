package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

func Who(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// The room details are built from the account's settings; text for an NPC goes nowhere.
	user := actor.Controller().Account()
	if user == nil {
		return true, nil
	}

	details := rooms.GetDetails(room, user)

	whoTxt, _ := templates.Process("descriptions/who", details, user.UserId)
	user.SendText(whoTxt)

	return true, nil
}
