package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Give(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		Tell(actor, `Give what? To whom? (<ansi fg="command">give {object-name} {receiver-name}</ansi>)`)
		return true, nil
	}

	var giveWho string = args[len(args)-1]
	args = args[:len(args)-1]
	var giveWhat string = strings.Join(args, " ")

	var giveItem items.Item = items.Item{}
	var giveGoldAmount int = 0

	if len(giveWhat) > 4 && giveWhat[len(giveWhat)-4:] == "gold" {

		g, _ := strconv.ParseInt(giveWhat[0:len(giveWhat)-5], 10, 32)
		giveGoldAmount = int(g)

		if giveGoldAmount < 0 {
			Tell(actor, "You can't give a negative amount of gold.")
			return true, nil
		}

		if giveGoldAmount > actor.Character.Gold {
			Tell(actor, "You don't have that much gold to give.")
			return true, nil
		}

	} else {

		var found bool = false

		// Check whether the user has an item in their inventory that matches
		giveItem, found = actor.Character.FindInBackpack(giveWhat)

		if !found {
			Tell(actor, fmt.Sprintf(`You don't have a %s to give. (<ansi fg="command">give {object-name} {receiver-name}</ansi>)`, giveWhat))
			return true, nil
		}

	}

	playerId, mobId := room.FindByName(giveWho)

	if playerId > 0 {

		actor.Character.CancelBuffsWithFlag("hidden")

		targetUser := accounts.GetByUserId(playerId)
		if targetUser == nil {
			return true, nil
		}

		// Swap the item location
		if giveItem.ItemId > 0 {
			targetUser.Character.StoreItem(giveItem)
			actor.Character.RemoveItem(giveItem)

			Tell(actor,
				fmt.Sprintf(`You give the <ansi fg="item">%s</ansi> to <ansi fg="username">%s</ansi>.`, giveItem.DisplayName(), targetUser.Character.Name),
			)
			targetUser.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> gives you their <ansi fg="item">%s</ansi>.`, NameColour(actor), actor.Character.Name, giveItem.DisplayName()),
			)
			room.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> gives <ansi fg="username">%s</ansi> a <ansi fg="itemname">%s</ansi>.`, NameColour(actor), actor.Character.Name, targetUser.Character.Name, giveItem.NameSimple()),
				ExcludeId(actor),
				targetUser.UserId)

			events.AddToQueue(events.ItemOwnership{
				ActorId: actor.Character.Id,
				Item:    giveItem,
				Gained:  false,
			})

			events.AddToQueue(events.ItemOwnership{
				ActorId: targetUser.Character.Id,
				Item:    giveItem,
				Gained:  true,
			})

		} else if giveGoldAmount > 0 {

			if targetUser.UserId == ExcludeId(actor) {

				Tell(actor,
					fmt.Sprintf(`You count out <ansi fg="gold">%d gold</ansi> and put it back in your pocket.`, giveGoldAmount),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> counts out some <ansi fg="gold">gold</ansi> and put it back in their pocket.`, NameColour(actor), actor.Character.Name),
					ExcludeId(actor))

			} else {
				targetUser.Character.Gold += giveGoldAmount
				actor.Character.Gold -= giveGoldAmount

				events.AddToQueue(events.EquipmentChange{
					ActorId:    targetUser.Character.Id,
					GoldChange: giveGoldAmount,
				})

				events.AddToQueue(events.EquipmentChange{
					ActorId:    actor.Character.Id,
					GoldChange: -giveGoldAmount,
				})

				Tell(actor,
					fmt.Sprintf(`You give <ansi fg="gold">%d gold</ansi> to <ansi fg="username">%s</ansi>.`, giveGoldAmount, targetUser.Character.Name),
				)
				targetUser.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> gives you <ansi fg="gold">%d gold</ansi>.`, NameColour(actor), actor.Character.Name, giveGoldAmount),
				)
				room.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> gives <ansi fg="username">%s</ansi> some <ansi fg="gold">gold</ansi>.`, NameColour(actor), actor.Character.Name, targetUser.Character.Name),
					ExcludeId(actor),
					targetUser.UserId)
			}
		} else {
			Tell(actor, "Something went wrong.")
		}

		return true, nil

	}

	//
	// Look for an NPC
	//
	if !mobId.IsZero() {

		actor.Character.CancelBuffsWithFlag("hidden")

		m := mobs.GetInstance(mobId)

		if m != nil {

			// Swap the item location
			if giveItem.ItemId > 0 || giveGoldAmount > 0 {

				if giveGoldAmount > 0 {
					m.Character.Gold += giveGoldAmount
					actor.Character.Gold -= giveGoldAmount

					events.AddToQueue(events.EquipmentChange{
						ActorId:    actor.Character.Id,
						GoldChange: -giveGoldAmount,
					})

					Tell(actor,
						fmt.Sprintf(`You give <ansi fg="gold">%d gold</ansi> to <ansi fg="username">%s</ansi>.`, giveGoldAmount, m.Character.Name),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="%s">%s</ansi> gave some gold to <ansi fg="mobname">%s</ansi>.`, NameColour(actor), actor.Character.Name, m.Character.Name),
						ExcludeId(actor),
					)
				} else {

					m.Character.StoreItem(giveItem)
					actor.Character.RemoveItem(giveItem)

					Tell(actor,
						fmt.Sprintf(`You give the <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, giveItem.DisplayName(), m.Character.Name),
					)
					room.SendText(
						fmt.Sprintf(`<ansi fg="%s">%s</ansi> gave their <ansi fg="item">%s</ansi> to <ansi fg="mobname">%s</ansi>.`, NameColour(actor), actor.Character.Name, giveItem.DisplayName(), m.Character.Name),
						ExcludeId(actor),
					)

					events.AddToQueue(events.ItemOwnership{
						ActorId: actor.Character.Id,
						Item:    giveItem,
						Gained:  false,
					})

					events.AddToQueue(events.ItemOwnership{
						ActorId: m.Character.Id,
						Item:    giveItem,
						Gained:  true,
					})

				}

				if giveGoldAmount > 0 {
					m.Command(`emote counts his gold coins and chuckles a bit.`)
				} else {
					m.Command(fmt.Sprintf(`emote considers the <ansi fg="itemname">%s</ansi> for a moment.`, giveItem.DisplayName()))
				}
			} else {
				Tell(actor, "Something went wrong.")
			}

		}

		return true, nil
	}

	Tell(actor, `Who??? (<ansi fg="command">give {object-name} {receiver-name}</ansi>)`)

	return true, nil
}
