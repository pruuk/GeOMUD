package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Shout(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user := actor.Controller().Account(); user != nil && user.Muted {
		tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	isSneaking := actor.Character.HasBuffFlag("hidden")
	isDrunk := actor.Character.HasBuffFlag("drunk")

	rest = strings.ToUpper(rest)

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}

	if isSneaking {
		room.SendTextCommunication(fmt.Sprintf(`someone shouts, "<ansi fg="yellow">%s</ansi>"`, rest), excludeId(actor))
	} else {
		room.SendTextCommunication(fmt.Sprintf(`<ansi fg="username">%s</ansi> shouts, "<ansi fg="yellow">%s</ansi>"`, actor.Character.Name, rest), excludeId(actor))
	}

	for _, roomInfo := range room.Exits {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendTextCommunication(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), excludeId(actor))
			}
		}
	}

	for _, roomInfo := range room.ExitsTemp {
		if otherRoom := rooms.LoadRoom(roomInfo.RoomId); otherRoom != nil {
			if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
				otherRoom.SendTextCommunication(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), excludeId(actor))
			}
		}
	}

	for mut := range room.ActiveMutators {
		spec := mut.GetSpec()
		if len(spec.Exits) == 0 {
			continue
		}
		for _, exitInfo := range spec.Exits {
			if otherRoom := rooms.LoadRoom(exitInfo.RoomId); otherRoom != nil {
				if sourceExit := otherRoom.FindExitTo(room.RoomId); sourceExit != `` {
					otherRoom.SendTextCommunication(fmt.Sprintf(`Someone shouts from the <ansi fg="exit">%s</ansi> direction, "<ansi fg="yellow">%s</ansi>"`, sourceExit, rest), excludeId(actor))
				}
			}
		}
	}

	tell(actor, fmt.Sprintf(`You shout, "<ansi fg="yellow">%s</ansi>"`, rest))

	return true, nil
}
