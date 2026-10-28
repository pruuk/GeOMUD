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
		Tell(actor, infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := accounts.GetByUserId(targetUserId); u != nil {

			u.Muted = true

			Tell(actor, fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-5">MUTED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	Tell(actor, "Could not find user.")
	return true, nil
}

func UnMute(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.mute", nil, user.UserId)
		Tell(actor, infoOutput)
		return true, nil
	}

	targetUserId, _ := room.FindByName(rest)

	if targetUserId > 0 {

		if u := accounts.GetByUserId(targetUserId); u != nil {

			u.Muted = false

			Tell(actor, fmt.Sprintf(`<ansi fg="username">%s</ansi> (<ansi fg="username">%s</ansi>) has been <ansi fg="alert-1">UNMUTED</ansi>`, u.Username, u.Character.Name))

			return true, nil
		}

	}

	Tell(actor, "Could not find user.")
	return true, nil
}
