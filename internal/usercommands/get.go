package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Get(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		tell(actor, "Get what?")
		return true, nil
	}

	if args[0] == "all" {
		if room.Gold > 0 {
			Get(`gold`, actor, room, flags)
		}

		if len(room.Items) > 0 {
			iCopies := append([]items.Item{}, room.Items...)

			for _, item := range iCopies {
				Get(item.Name(), actor, room, flags)
			}
		}

		return true, nil
	}

	getFromStash := false
	containerName := ``
	var corpseRef *rooms.Corpse

	if len(args) >= 2 {
		// Detect "stash" or "from stash" at end and remove it
		if args[len(args)-1] == "stash" {
			getFromStash = true
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

		if args[len(args)-1] == "ground" {
			getFromStash = false
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

		containerName = room.FindContainerByName(args[len(args)-1])
		if containerName != `` {
			getFromStash = false
			if args[len(args)-2] == "from" {
				rest = strings.Join(args[0:len(args)-2], " ")
			} else {
				rest = strings.Join(args[0:len(args)-1], " ")
			}
		}

		// Look for a corpse as the source when CorpseItems is enabled
		if containerName == `` && configs.GetGamePlayConfig().Death.CorpseItems {
			if c, ok := room.FindCorpseByRef(args[len(args)-1]); ok {
				corpseRef = c
				if len(args) >= 2 && args[len(args)-2] == "from" {
					rest = strings.Join(args[0:len(args)-2], " ")
				} else {
					rest = strings.Join(args[0:len(args)-1], " ")
				}
			}
		}
	}

	// Handle getting from a corpse
	if corpseRef != nil {

		corpseColor := `mob-corpse`
		if corpseRef.UserId > 0 {
			corpseColor = `user-corpse`
		}
		corpseName := fmt.Sprintf(`<ansi fg="%s">%s corpse</ansi>`, corpseColor, corpseRef.Character.Name)

		// "get all <corpse>"
		if rest == "all" {
			tookSomething := false

			if corpseRef.Gold > 0 {
				goldAmt := corpseRef.Gold
				actor.Character.Gold += goldAmt
				corpseRef.Gold = 0

				events.AddToQueue(events.EquipmentChange{
					ActorId:    actor.Character.Id,
					GoldChange: -goldAmt,
				})

				tell(actor, fmt.Sprintf(`You take <ansi fg="gold">%d gold</ansi> from the %s.`, goldAmt, corpseName))
				room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> takes some <ansi fg="gold">gold</ansi> from the %s.`, actor.Character.Name, corpseName), excludeId(actor))
				tookSomething = true
			}

			// Backpack items on corpse
			allCorpseItems := append([]items.Item{}, corpseRef.Items...)
			for _, item := range allCorpseItems {
				if actor.Character.StoreItem(item) {
					corpseRef.RemoveItem(item)
					events.AddToQueue(events.ItemOwnership{
						ActorId: actor.Character.Id,
						Item:    item,
						Gained:  true,
					})
					tell(actor, fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the %s.`, item.DisplayName(), corpseName))
					room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> takes the <ansi fg="itemname">%s</ansi> from the %s.`, actor.Character.Name, item.DisplayName(), corpseName), excludeId(actor))
					tookSomething = true
				} else {
					tell(actor, fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, item.DisplayName()))
				}
			}

			// Worn items on corpse character
			allWorn := corpseRef.Character.GetAllWornItems()
			for _, item := range allWorn {
				if actor.Character.StoreItem(item) {
					corpseRef.Character.RemoveFromBody(item)
					events.AddToQueue(events.ItemOwnership{
						ActorId: actor.Character.Id,
						Item:    item,
						Gained:  true,
					})
					tell(actor, fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the %s.`, item.DisplayName(), corpseName))
					room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> takes the <ansi fg="itemname">%s</ansi> from the %s.`, actor.Character.Name, item.DisplayName(), corpseName), excludeId(actor))
					tookSomething = true
				} else {
					tell(actor, fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, item.DisplayName()))
				}
			}

			if !tookSomething {
				tell(actor, fmt.Sprintf(`There is nothing to take from the %s.`, corpseName))
			}

			return true, nil
		}

		// "get gold <corpse>"
		goldName := `gold`
		if args[0] == goldName || (len(args[0]) >= 1 && len(args[0]) < 5 && goldName[0:len(args[0])] == args[0]) {
			if corpseRef.Gold < 1 {
				tell(actor, fmt.Sprintf(`There is no gold on the %s.`, corpseName))
			} else {
				actor.Character.CancelBuffsWithFlag("hidden")

				goldAmt := corpseRef.Gold
				actor.Character.Gold += goldAmt
				corpseRef.Gold = 0

				events.AddToQueue(events.EquipmentChange{
					ActorId:    actor.Character.Id,
					GoldChange: -goldAmt,
				})

				tell(actor, fmt.Sprintf(`You take <ansi fg="gold">%d gold</ansi> from the %s.`, goldAmt, corpseName))
				room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> takes some <ansi fg="gold">gold</ansi> from the %s.`, actor.Character.Name, corpseName), excludeId(actor))
			}
			return true, nil
		}

		// "get <item> <corpse>" - search backpack items first, then worn items
		matchItem, found := corpseRef.FindItem(rest)
		if !found {
			// Also search worn equipment on the corpse character
			wornItems := corpseRef.Character.GetAllWornItems()
			closeMatch, exactMatch := items.FindMatchIn(rest, wornItems...)
			if exactMatch.ItemId != 0 {
				matchItem = exactMatch
				found = true
			} else if closeMatch.ItemId != 0 {
				matchItem = closeMatch
				found = true
			}
		}

		if !found {
			tell(actor, fmt.Sprintf(`You don't see a %s on the %s.`, rest, corpseName))
			return true, nil
		}

		actor.Character.CancelBuffsWithFlag("hidden")

		if !actor.Character.StoreItem(matchItem) {
			tell(actor, fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()))
			return true, nil
		}

		events.AddToQueue(events.ItemOwnership{
			ActorId: actor.Character.Id,
			Item:    matchItem,
			Gained:  true,
		})

		// Remove from whichever location it came from
		corpseRef.RemoveItem(matchItem)
		corpseRef.Character.RemoveFromBody(matchItem)

		tell(actor, fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the %s.`, matchItem.DisplayName(), corpseName))
		room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> takes the <ansi fg="itemname">%s</ansi> from the %s.`, actor.Character.Name, matchItem.DisplayName(), corpseName), excludeId(actor))

		return true, nil
	}

	if containerName != `` {
		container := room.Containers[containerName]

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) >= 1 && len(args[0]) < 5 && goldName[0:len(args[0])] == args[0]) {

			if container.Gold < 1 {
				tell(actor, "There's no gold to grab.")
			} else {

				actor.Character.CancelBuffsWithFlag("hidden") // No longer sneaking

				goldAmt := container.Gold
				actor.Character.Gold += goldAmt
				container.Gold -= goldAmt
				room.Containers[containerName] = container

				events.AddToQueue(events.EquipmentChange{
					ActorId:    actor.Character.Id,
					GoldChange: -goldAmt,
				})

				tell(actor,
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi> from the <ansi fg="container">%s</ansi>.`, goldAmt, containerName),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi> from the <ansi fg="container">%s</ansi>.`, actor.Character.Name, containerName),
					excludeId(actor),
				)
			}

			return true, nil
		}

		matchItem, found := container.FindItem(rest)

		if !found {
			tell(actor, fmt.Sprintf(`You don't see a %s in the <ansi fg="container">%s</ansi>.`, rest, containerName))
		} else {

			actor.Character.CancelBuffsWithFlag("hidden") // No longer sneaking

			// Trigger onFound event
			if actor.Character.StoreItem(matchItem) {

				events.AddToQueue(events.ItemOwnership{
					ActorId: actor.Character.Id,
					Item:    matchItem,
					Gained:  true,
				})

				// Swap the item location
				container.RemoveItem(matchItem)
				room.Containers[containerName] = container

				tell(actor,
					fmt.Sprintf(`You take the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>.`, matchItem.DisplayName(), containerName),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi> from the <ansi fg="container">%s</ansi>...`, actor.Character.Name, matchItem.DisplayName(), containerName),
					excludeId(actor),
				)

				return true, nil

			} else {
				tell(actor,
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
				)
			}

		}

	} else {

		goldName := `gold`
		if args[0] == goldName || (len(args[0]) >= 1 && len(args[0]) < 5 && goldName[0:len(args[0])] == args[0]) {

			if room.Gold < 1 {
				tell(actor, "There's no gold to grab.")
			} else {

				actor.Character.CancelBuffsWithFlag("hidden") // No longer sneaking

				goldAmt := room.Gold
				actor.Character.Gold += goldAmt
				room.Gold -= goldAmt

				events.AddToQueue(events.EquipmentChange{
					ActorId:    actor.Character.Id,
					GoldChange: -goldAmt,
				})

				tell(actor,
					fmt.Sprintf(`You pick up <ansi fg="gold">%d gold</ansi>.`, goldAmt),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up some <ansi fg="gold">gold</ansi>.`, actor.Character.Name),
					excludeId(actor),
				)
			}

			return true, nil
		}

		// Check whether the user has an item in their inventory that matches
		matchItem, found := room.FindOnFloor(rest, getFromStash)

		// Check if user is specifying an item they stashed
		if !found && !getFromStash {
			stashItemMatch, stashFound := room.FindOnFloor(rest, true)
			// Stash recovery is keyed by account: an actor without one reclaims nothing.
			if user := actor.Controller().Account(); stashFound && user != nil && stashItemMatch.StashedBy == user.UserId {
				found = true
				getFromStash = true
				matchItem = stashItemMatch
			}
		}

		if found {

			if matchItem.HasAdjective(`exploding`) {
				tell(actor, `You can't pick that up, it's about to explode!`)
				return true, nil
			}

			actor.Character.CancelBuffsWithFlag("hidden") // No longer sneaking

			// If it was in the stash, remove the stash owner tag
			if getFromStash {
				matchItem.StashedBy = 0
			}

			if actor.Character.StoreItem(matchItem) {

				// Swap the item location
				room.RemoveItem(matchItem, getFromStash)

				events.AddToQueue(events.ItemOwnership{
					ActorId: actor.Character.Id,
					Item:    matchItem,
					Gained:  true,
				})

				if getFromStash {
					tell(actor,
						fmt.Sprintf(`You dig out the <ansi fg="itemname">%s</ansi> from where it was stashed.`, matchItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> digs around in the area and picks something up...`, actor.Character.Name),
						excludeId(actor),
					)
				} else {
					tell(actor,
						fmt.Sprintf(`You pick up the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> picks up the <ansi fg="itemname">%s</ansi>...`, actor.Character.Name, matchItem.DisplayName()),
						excludeId(actor),
					)
				}

			} else {
				tell(actor,
					fmt.Sprintf(`You can't carry the <ansi fg="itemname">%s</ansi>.`, matchItem.DisplayName()),
				)
			}

			return true, nil
		}

		//
		// Look for any nouns in the room info
		//
		foundNoun, _ := room.FindNoun(rest)
		if len(foundNoun) > 0 {

			tell(actor, fmt.Sprintf(`You can't get the <ansi fg="noun">%s</ansi>`, foundNoun))
			room.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is grasping at the air.`, actor.Character.Name), excludeId(actor))

			return true, nil
		}

	}

	if _, corpseFound := room.FindCorpse(rest); corpseFound {
		tell(actor, `You can't pick up corpses. What would people think?`)
		return true, nil
	}

	containerName = room.FindContainerByName(rest)
	if containerName != `` {
		tell(actor, fmt.Sprintf(`You can't pick up the <ansi fg="container">%s</ansi>. Try looking at it.`, containerName))
	} else {
		tell(actor, fmt.Sprintf("You don't see a %s around.", rest))
	}

	return true, nil
}
