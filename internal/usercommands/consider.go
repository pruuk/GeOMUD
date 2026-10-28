package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/combat"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Consider(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		return true, nil
	}

	lookAt := args[0]

	playerId, mobId := room.FindByName(lookAt)
	if playerId == UserIdOf(actor) {
		playerId = 0
	}

	if playerId == 0 && mobId.IsZero() {
		if lookAt == `` {
			Tell(actor, `<ansi fg="command">consider</ansi> whom??`)
		} else {
			Tell(actor, `<ansi fg="mobname">`+lookAt+`</ansi> can't be found.`)
		}
		return true, nil
	}

	var odds float64
	considerType := "mob"
	considerName := "nobody"

	if playerId > 0 {
		u := accounts.GetByUserId(playerId)
		odds = combat.CombatOdds(*actor.Character, *u.Character)
		considerType = "user"
		considerName = u.Character.Name
	} else {
		m := mobs.GetInstance(mobId)
		odds = combat.CombatOdds(*actor.Character, m.Character)
		considerType = "mob"
		considerName = m.Character.Name
	}

	// odds is defRoundsToKill / atkRoundsToKill.
	// > 1.0 means the player kills the target faster than the target kills the player.
	// 1.0 is a dead-even fight.
	prediction := `Unknown`
	switch {
	case odds >= 3.0:
		prediction = `<ansi fg="blue-bold">Very Favorable</ansi>`
	case odds >= 1.75:
		prediction = `<ansi fg="green">Favorable</ansi>`
	case odds >= 1.1:
		prediction = `<ansi fg="green">Good</ansi>`
	case odds >= 0.9:
		prediction = `<ansi fg="yellow">Even</ansi>`
	case odds >= 0.5:
		prediction = `<ansi fg="red-bold">Bad</ansi>`
	case odds > 0:
		prediction = `<ansi fg="red-bold">Very Bad</ansi>`
	default:
		prediction = `<ansi fg="red-bold">YOU WILL DIE</ansi>`
	}

	Tell(actor,
		fmt.Sprintf(`You consider <ansi fg="%sname">%s</ansi>...`, considerType, considerName),
	)
	Tell(actor,
		fmt.Sprintf(`It is estimated that your chances to kill <ansi fg="%sname">%s</ansi> are %s`, considerType, considerName, prediction),
	)

	return true, nil
}
