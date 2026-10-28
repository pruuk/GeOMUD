package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Show(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		tell(actor, "Show what? To whom?")
		return true, nil
	}

	var showItem items.Item = items.Item{}
	var found bool = false

	var targetName string = args[len(args)-1]
	args = args[:len(args)-1]
	var objectName string = strings.Join(args, " ")

	// Check whether the user has an item in their inventory that matches
	showItem, found = actor.Character.FindInBackpack(objectName)

	if !found {
		tell(actor, fmt.Sprintf("You don't have a %s to show.", objectName))
		return true, nil
	}

	playerId, mobId := room.FindByName(targetName)

	if playerId > 0 {

		actor.Character.CancelBuffsWithFlag("hidden")

		targetUser := accounts.GetByUserId(playerId)

		// Swap the item location
		if showItem.ItemId > 0 {

			// Tell the shower
			tell(actor,
				fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, showItem.DisplayName(), targetUser.Character.Name),
			)

			// Tell the Showee
			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> shows you their <ansi fg="item">%s</ansi>.`, actor.Character.Name, showItem.DisplayName()),
			)

			targetUser.SendText(
				"\n" + showItem.GetLongDescription() + "\n",
			)

			// Tell the rest of the room
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, actor.Character.Name, showItem.DisplayName(), targetUser.Character.Name),
				targetUser.UserId,
				excludeId(actor))

		} else {
			tell(actor, "Something went wrong.")
		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if !mobId.IsZero() {

		actor.Character.CancelBuffsWithFlag("hidden")

		targetMob := mobs.GetInstance(mobId)

		if targetMob != nil {

			if showItem.ItemId > 0 {

				tell(actor,
					fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, showItem.DisplayName(), targetMob.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> shows their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, actor.Character.Name, showItem.DisplayName(), targetMob.Character.Name),
					excludeId(actor),
				)

			} else {
				tell(actor, "Something went wrong.")
			}

		}

		return true, nil
	}

	tell(actor, "Who???")

	return true, nil
}
