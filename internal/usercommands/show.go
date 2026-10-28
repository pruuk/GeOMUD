package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Show(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	rest = util.StripPrepositions(rest)

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 2 {
		Tell(actor, "Show what? To whom?")
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
		Tell(actor, fmt.Sprintf("You don't have a %s to show.", objectName))
		return true, nil
	}

	playerId, mobId := room.FindByName(targetName)

	if playerId > 0 {

		actor.Character.CancelBuffsWithFlag("hidden")

		targetUser := accounts.GetByUserId(playerId)

		// Swap the item location
		if showItem.ItemId > 0 {

			// The shower, the showee (with the long description that used to be
			// a second send) and the rest of the room.
			messaging.Send(perception.Message{
				Kind:      perception.WorldEvent,
				Actor:     actor.Id,
				Target:    targetUser.Character.Id,
				Senses:    perception.Sight | perception.Sound,
				Range:     perception.Room,
				ActorText: fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to {target}.`, showItem.DisplayName()),
				TargetText: fmt.Sprintf(`{actor} shows you their <ansi fg="item">%s</ansi>.`, showItem.DisplayName()) +
					"\n\n" + messaging.Literal(showItem.GetLongDescription()) + "\n",
				ObserverText: fmt.Sprintf(`{actor} shows their <ansi fg="item">%s</ansi> to {target}.`, showItem.DisplayName()),
			})

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

		targetMob := mobs.GetInstance(mobId)

		if targetMob != nil {

			if showItem.ItemId > 0 {

				messaging.Send(perception.Message{
					Kind:         perception.WorldEvent,
					Actor:        actor.Id,
					Target:       mobId,
					Senses:       perception.Sight | perception.Sound,
					Range:        perception.Room,
					ActorText:    fmt.Sprintf(`You show the <ansi fg="item">%s</ansi> to {target}.`, showItem.DisplayName()),
					ObserverText: fmt.Sprintf(`{actor} shows their <ansi fg="item">%s</ansi> to {target}.`, showItem.DisplayName()),
				})

			} else {
				Tell(actor, "Something went wrong.")
			}

		}

		return true, nil
	}

	Tell(actor, "Who???")

	return true, nil
}
