package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Drop(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		Tell(actor, `Drop what?`)

		return true, nil
	}

	if args[0] == "all" {

		iCopies := []items.Item{}

		if actor.Character.Gold > 0 {
			Drop(fmt.Sprintf("%d gold", actor.Character.Gold), actor, room, flags)
		}

		// Gold has dropped; stuck gear refuses the items once, not per item.
		if permaGearRefused(actor) {
			return true, nil
		}

		iCopies = append(iCopies, actor.Character.Items...)

		for _, item := range iCopies {
			Drop(item.Name(), actor, room, flags)
		}

		return true, nil
	}

	// Drop 10 gold
	if len(args) >= 2 && args[1] == "gold" {
		g, _ := strconv.ParseInt(args[0], 10, 32)
		dropAmt := int(g)
		if dropAmt < 1 {
			Tell(actor, "Oops!")
			return true, nil
		}

		if dropAmt > actor.Character.Gold {
			Tell(actor, fmt.Sprintf("You don't have a %d gold to drop.", dropAmt))
			return true, nil
		}

		actor.Character.CancelBuffsWithFlag("hidden")

		room.Gold += dropAmt
		actor.Character.Gold -= dropAmt

		events.AddToQueue(events.EquipmentChange{
			ActorId:    actor.Character.Id,
			GoldChange: -dropAmt,
		})

		itemRoomEvent(actor,
			fmt.Sprintf(`You drop <ansi fg="gold">%d gold</ansi> on the floor.`, dropAmt),
			fmt.Sprintf(`{actor} drops <ansi fg="gold">%d gold</ansi>.`, dropAmt),
		)

		return true, nil
	}

	if permaGearRefused(actor) {
		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := actor.Character.FindInBackpack(rest)

	if !found {
		Tell(actor, fmt.Sprintf("You don't have a %s to drop.", rest))
	} else {

		actor.Character.CancelBuffsWithFlag("hidden")

		iSpec := matchItem.GetSpec()

		// Check floor item limit before dropping
		floorLimit := int(configs.GetGamePlayConfig().FloorItemCountMax)
		if floorLimit > 0 && len(room.Items) >= floorLimit {
			itemRoomEvent(actor,
				fmt.Sprintf(`There are too many items on the floor! The <ansi fg="item">%s</ansi> re-appears in your hands!`, matchItem.DisplayName()),
				fmt.Sprintf(`There are too many items on the floor! The <ansi fg="item">%s</ansi> re-appears in {actor}'s hands!`, matchItem.DisplayName()),
			)
			return true, nil
		}

		// Swap the item location
		actor.Character.RemoveItem(matchItem)

		events.AddToQueue(events.ItemOwnership{
			ActorId: actor.Character.Id,
			Item:    matchItem,
			Gained:  false,
		})

		itemRoomEvent(actor,
			fmt.Sprintf(`You drop the <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
			fmt.Sprintf(`{actor} drops their <ansi fg="item">%s</ansi>...`, matchItem.DisplayName()),
		)

		// If grenades are dropped, they explode and affect everyone in the room!
		if iSpec.Type == items.Grenade {

			Tell(actor, `Todo. Grenades disabled for now.`)

		}

		room.AddItem(matchItem, false)

	}

	return true, nil
}
