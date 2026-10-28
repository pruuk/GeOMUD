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

func Put(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		tell(actor, "Place what where?")
		return true, nil
	}

	containerName := ``
	nameSearch := ``
	for i := len(args) - 1; i >= 1; i-- {
		if len(nameSearch) > 0 {
			nameSearch = ` ` + nameSearch
		}
		nameSearch = args[i] + nameSearch

		containerName = room.FindContainerByName(nameSearch)
		if containerName != `` {
			args = args[:i]
			break
		}
	}

	if containerName == `` {
		tell(actor, `No container found by that name`)
		return true, nil
	}

	container := room.Containers[containerName]

	if container.Lock.IsLocked() {
		tell(actor, ``)
		tell(actor, fmt.Sprintf(`The <ansi fg="container">%s</ansi> is locked.`, containerName))
		tell(actor, ``)
		return true, nil
	}

	if len(args) < 1 {
		tell(actor, "Place what where?")
		return true, nil
	}

	var item items.Item
	var itemFound bool
	goldAmt := 0

	if len(args) >= 2 && args[1] == `gold` {

		g, _ := strconv.ParseInt(args[0], 10, 32)
		goldAmt = int(g)
		if goldAmt < 0 {
			goldAmt = -1 * goldAmt
		}

	} else {

		item, itemFound = actor.Character.FindInBackpack(strings.Join(args, ` `))
		if !itemFound && len(args) > 1 {
			item, itemFound = actor.Character.FindInBackpack(args[0])
		}

	}

	if !itemFound && goldAmt == 0 {
		tell(actor, `You don't seem to be carrying that.`)
		return true, nil
	}

	if goldAmt > actor.Character.Gold {
		tell(actor, `You don't have that much gold.`)
		return true, nil
	}

	if goldAmt > 0 {
		actor.Character.Gold -= goldAmt

		events.AddToQueue(events.EquipmentChange{
			ActorId:    actor.Character.Id,
			GoldChange: goldAmt,
		})

		container.Gold += goldAmt
		tell(actor, fmt.Sprintf(`You place <ansi fg="gold">%d gold</ansi> into the <ansi fg="container">%s</ansi>`, goldAmt, containerName))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> places some <ansi fg="gold">gold</ansi> into the <ansi fg="container">%s</ansi>`, actor.Character.Name, containerName), excludeId(actor))
	}

	if itemFound {

		container.AddItem(item)
		actor.Character.RemoveItem(item)

		events.AddToQueue(events.ItemOwnership{
			ActorId: actor.Character.Id,
			Item:    item,
			Gained:  false,
		})

		tell(actor, fmt.Sprintf(`You place your <ansi fg="itemname">%s</ansi> into the <ansi fg="container">%s</ansi>`, item.DisplayName(), containerName))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> places their <ansi fg="itemname">%s</ansi> into the <ansi fg="container">%s</ansi>`, actor.Character.Name, item.DisplayName(), containerName), excludeId(actor))

		// Enforce container size limits

		if len(container.Items) > int(configs.GetGamePlayConfig().ContainerSizeMax) {

			randItemToRemove := util.Rand(len(container.Items))
			oopsItem := container.Items[randItemToRemove]

			// get all items that spawn in chests
			for _, spn := range room.SpawnInfo {
				if spn.Container == containerName && oopsItem.ItemId == spn.ItemId {
					// Don't let this one pop out
					oopsItem = item
					break
				}
			}

			container.RemoveItem(oopsItem)
			room.SendText(fmt.Sprintf(`The <ansi fg="container">%s</ansi> is too full and a <ansi fg="itemname">%s</ansi> falls out and onto the floor.`, containerName, oopsItem.DisplayName()))
			room.AddItem(oopsItem, false)
		}
	}

	room.Containers[containerName] = container

	return true, nil
}
