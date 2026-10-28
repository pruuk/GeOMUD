package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"
)

/*
* Role Permissions:
* spawn				(All)
 */
func Spawn(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.spawn", nil, user.UserId)
		Tell(actor, infoOutput)
		return true, nil
	}

	spawnType := args[0]
	args = args[1:]

	spawnTarget := ``
	if len(args) == 1 {
		spawnTarget = args[0]
		args = args[1:]
	} else {
		spawnTarget = strings.Join(args, ` `)
		args = []string{}

	}

	if len(spawnTarget) > 0 {

		if spawnType == `container` {

			containerName := room.SpawnTempContainer(spawnTarget, "3 rounds", 0)

			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Actor:        actor.Id,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ActorText:    fmt.Sprintf(`You wave your hands around and <ansi fg="container">%s</ansi> appears from thin air and falls to the ground.`, containerName),
				ObserverText: fmt.Sprintf(`{actor} waves their hands around and <ansi fg="container">%s</ansi> appears from thin air and falls to the ground.`, containerName),
			})

			return true, nil
		}

		if spawnType == `gold` || spawnTarget == `gold` {

			goldAmt := 0
			if spawnType == `gold` {
				goldAmt, _ = strconv.Atoi(spawnTarget)
			} else {
				goldAmt, _ = strconv.Atoi(spawnType)
			}

			if goldAmt < 1 {
				goldAmt = 1
			}

			room.Gold += goldAmt

			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Actor:        actor.Id,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ActorText:    fmt.Sprintf(`You wave your hands around and <ansi fg="gold">%d gold</ansi> appears from thin air and falls to the ground.`, goldAmt),
				ObserverText: fmt.Sprintf(`{actor} waves their hands around and <ansi fg="gold">%d gold</ansi> appears from thin air and falls to the ground.`, goldAmt),
			})

			return true, nil
		}

	}

	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        actor.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ActorText:    "You wave your hands around pathetically.",
		ObserverText: "{actor} waves their hands around pathetically.",
	})

	return true, nil
}
