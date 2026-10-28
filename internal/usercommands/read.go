package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Read(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// Check whether the user has an item in their inventory that matches

	foundItemName := ""
	foundItemLongName := ""
	foundItemDescription := ""
	// Search for an exact match first
	if readItem, found := actor.Character.FindInBackpack(rest); found {
		iSpec := readItem.GetSpec()
		if iSpec.Type == items.Readable {
			foundItemName = readItem.DisplayName()
			foundItemLongName = readItem.DisplayName()
			foundItemDescription = string(readItem.GetBlob())
			if len(foundItemDescription) == 0 {
				foundItemDescription = iSpec.Description
			}
		}
	}

	isSneaking := actor.Character.HasBuffFlag("hidden")

	if len(foundItemName) == 0 {
		Tell(actor, fmt.Sprintf(`You don't have a "%s" that can be read.`, rest))
	} else {
		youLine := fmt.Sprintf(`You look at <ansi fg="item">%s</ansi>...`, foundItemLongName)

		if isSneaking {
			Tell(actor, youLine)
		} else {
			itemRoomEvent(actor,
				youLine,
				fmt.Sprintf(`{actor} looks at their <ansi fg="item">%s</ansi>...`, foundItemName),
			)
		}

		Tell(actor, foundItemDescription)
	}

	return true, nil
}
