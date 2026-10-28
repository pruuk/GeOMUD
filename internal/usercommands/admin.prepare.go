package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

/*
* Role Permissions:
* prepare 				(All)
 */
func Prepare(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.prepare", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	allRoomIds := rooms.GetAllRoomIds()
	for _, roomId := range allRoomIds {
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}
		room.Prepare(false) // we are preparing all rooms, no need to check adjacent rooms
	}

	user.SendText(
		"All rooms have been Prepare()'ed",
	)

	return true, nil
}
