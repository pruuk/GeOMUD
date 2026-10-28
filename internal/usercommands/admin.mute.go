package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
)

/*
* Role Permissions:
* mute 				(All)
 */
func Mute(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.mute", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := accounts.GetByUserId(targetUserId); u != nil {

			u.Muted = true

			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-5">MUTED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	user.SendText("Could not find user.")
	return true, nil
}

func UnMute(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.mute", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := accounts.GetByUserId(targetUserId); u != nil {

			u.Muted = false

			user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-1">UNMUTED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	user.SendText("Could not find user.")
	return true, nil
}
