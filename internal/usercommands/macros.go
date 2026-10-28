package usercommands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Macros(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if len(user.Macros) == 0 {
		Tell(actor,"You have no macros set.")
		return true, nil
	}

	sortedKeys := make([]string, 0, len(user.Macros))

	for k := range user.Macros {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	Tell(actor,`<ansi fg="226">Your macros:</ansi>`)
	for _, macro := range sortedKeys {
		macroCommand := user.Macros[macro]
		Tell(actor,``)

		Tell(actor,fmt.Sprintf(`  <ansi fg="228">%s</ansi>:`, macro))

		commandParts := strings.Split(macroCommand, `;`)

		for i, cmd := range commandParts {

			cmdParts := strings.Split(cmd, ` `)
			cmdAlone := cmdParts[0]
			cmdRest := ``
			if len(cmdParts) > 1 {
				cmdRest = strings.Join(cmdParts[1:], ` `)
			}

			Tell(actor,fmt.Sprintf(`      %s) <ansi fg="command">%s</ansi> %s`, string(rune(97+i)), cmdAlone, cmdRest))
		}
	}
	Tell(actor,``)
	Tell(actor,`To use a macro, type <ansi fg="command">={num}</ansi>.`)
	Tell(actor,`Some terminals support pressing the associated F-Key (<ansi fg="228">F1</ansi>, <ansi fg="228">F2</ansi>, etc.)`)

	return true, nil
}
