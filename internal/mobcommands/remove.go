package mobcommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Remove(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	if mob.Character.HasBuffFlag("perma-gear") {
		mob.Command(`emote struggles with their gear for a while, then gives up.`)
		return true, nil
	}

	if rest == "all" {
		removedItems := []items.Item{}
		for _, item := range mob.Character.Equipment.GetAllItems() {
			Remove(item.Name(), mob, room)
			removedItems = append(removedItems, item)
		}

		events.AddToQueue(events.EquipmentChange{
			MobInstanceId: mob.Character.Id,
			ItemsRemoved:  removedItems,
		})

		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := mob.Character.FindOnBody(rest)

	if found && matchItem.ItemId > 0 {

		if mob.Character.RemoveFromBody(matchItem) {

			mob.Character.CancelBuffsWithFlag("hidden")

			room.SendText(
				fmt.Sprintf(`<ansi fg="mobname">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, mob.Character.Name, matchItem.DisplayName()),
			)

			mob.Character.StoreItem(matchItem)
		}

		mob.Character.Validate()

		events.AddToQueue(events.EquipmentChange{
			MobInstanceId: mob.Character.Id,
			ItemsRemoved:  []items.Item{matchItem},
		})

	}

	return true, nil
}
