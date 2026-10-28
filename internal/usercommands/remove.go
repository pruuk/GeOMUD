package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
)

func Remove(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if rest == "all" {
		removedItems := []items.Item{}
		for _, item := range actor.Character.Equipment.GetAllItems() {
			if item.IsRemoveLocked() {
				continue
			}
			Remove(item.Name(), actor, room, flags)
			removedItems = append(removedItems, item)
		}

		events.AddToQueue(events.EquipmentChange{
			ActorId:      actor.Character.Id,
			ItemsRemoved: removedItems,
		})

		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := actor.Character.FindOnBody(rest)

	if !found || matchItem.ItemId < 1 {
		tell(actor, fmt.Sprintf(`You don't appear to be using a "%s".`, rest))
	} else {

		if matchItem.IsRemoveLocked() {
			tell(actor,
				fmt.Sprintf(`Your <ansi fg="item">%s</ansi> is bound to you and <ansi fg="red-bold">cannot be removed</ansi>.`, matchItem.DisplayName()),
			)
			return true, nil
		}

		if matchItem.IsCursed() && actor.Character.Health > 0 {
			if actor.Character.GetSkillLevel(`enchant`) < skills.MaxSkillLevel(`enchant`) {
				tell(actor,
					fmt.Sprintf(`You can't seem to remove your <ansi fg="item">%s</ansi>... It's <ansi fg="red-bold">CURSED!</ansi>`, matchItem.DisplayName()),
				)

				return true, nil
			} else {
				tell(actor,
					`It's <ansi fg="red-bold">CURSED</ansi> but luckily your <ansi fg="skillname">enchant</ansi> skill level allows you to remove it.`,
				)
			}
		}

		actor.Character.CancelBuffsWithFlag("hidden")

		statsBefore := captureStatSnapshot(actor.Character)

		if actor.Character.RemoveFromBody(matchItem) {

			actor.Character.Validate()

			statDiff := formatStatChanges(statsBefore, captureStatSnapshot(actor.Character))

			tell(actor,
				fmt.Sprintf(`You remove your <ansi fg="item">%s</ansi> and return it to your backpack.%s`, matchItem.DisplayName(), statDiff),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="username">%s</ansi> removes their <ansi fg="item">%s</ansi> and stores it away.`, actor.Character.Name, matchItem.DisplayName()),
				excludeId(actor),
			)

			actor.Character.StoreItem(matchItem)

			events.AddToQueue(events.EquipmentChange{
				ActorId:      actor.Character.Id,
				ItemsRemoved: []items.Item{matchItem},
			})

		} else {
			tell(actor,
				fmt.Sprintf(`You can't seem to remove your <ansi fg="item">%s</ansi>.`, matchItem.DisplayName()),
			)
		}

	}

	return true, nil
}
