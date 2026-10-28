package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Shoot(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if actor.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		tell(actor, `You don't have a shooting weapon.`)
		return true, nil
	}

	attackPlayerId := 0
	attackMobId := ids.ID{}

	// It's possible that they are shooting in a direction, so check whether multiple words were provided
	// And whether the last word is a direction.
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		tell(actor, `Syntax: <ansi fg="command">shoot [target] [exit]</ansi>`)
		return true, nil
	}

	direction := args[len(args)-1]
	args = args[:len(args)-1]

	// Only shooting weapons can target adjacent rooms
	// "attack goblin east"
	exitName, attackRoomId := room.FindExitByName(direction)
	if exitName != `` {

		exitInfo, _ := room.GetExitInfo(exitName)
		if exitInfo.Lock.IsLocked() {
			tell(actor, fmt.Sprintf("The %s exit is locked.", exitName))
			return true, nil
		}

		if adjacentRoom := rooms.LoadRoom(attackRoomId); adjacentRoom != nil {
			attackPlayerId, attackMobId = adjacentRoom.FindByName(strings.Join(args, ` `))
		}
	} else {
		tell(actor, `Could not find where you wanted to shoot`)
		return true, nil
	}

	if attackPlayerId == 0 && attackMobId.IsZero() {
		tell(actor, `Could not find your target.`)
		return true, nil
	}

	isSneaking := actor.Character.HasBuffFlag("hidden")

	/*
		combatAddlWaitRounds := actor.Character.Equipment.Weapon.GetSpec().WaitRounds + actor.Character.Equipment.Weapon.GetSpec().WaitRounds
		attkType := characters.DefaultAttack
		if actor.Character.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			attkType = characters.Shooting
		}
	*/

	if !attackMobId.IsZero() {

		m := mobs.GetInstance(attackMobId)

		if m != nil {

			if m.Character.IsCharmed(actor.Character.Id) {
				tell(actor, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
				return true, nil
			}

			actor.Character.SetAggroRemote(exitName, attackMobId, characters.Shooting)

			tell(actor,
				fmt.Sprintf(`You prepare to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, m.Character.Name, exitName),
			)

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to shoot at <ansi fg="mobname">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, actor.Character.Name, m.Character.Name, exitName),
					excludeId(actor),
				)
			}

		}

	} else if attackPlayerId > 0 {

		p := accounts.GetByUserId(attackPlayerId)

		if p != nil {

			if user := actor.Controller().Account(); user != nil {
				if partyInfo := parties.Get(user.UserId); partyInfo != nil {
					if partyInfo.IsMember(attackPlayerId) {
						tell(actor, fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
						return true, nil
					}
				}
			}

			actor.Character.SetAggroRemote(exitName, p.Character.Id, characters.Shooting)

			tell(actor,
				fmt.Sprintf(`You prepare to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, p.Character.Name, exitName),
			)

			if !isSneaking {

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to shoot at <ansi fg="username">%s</ansi> through the <ansi fg="exit">%s</ansi> exit.`, actor.Character.Name, p.Character.Name, exitName),
					excludeId(actor),
				)

			}

		}

	}

	return true, nil
}
