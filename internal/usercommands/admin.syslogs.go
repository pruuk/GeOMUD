package usercommands

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"
)

/*
* Role Permissions:
* syslogs 				(All)
 */
func SysLogs(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.syslogs", nil, user.UserId)
		Tell(actor, infoOutput)
		return true, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if args[0] == "off" {

		events.AddToQueue(events.Log{FollowRemove: user.ConnectionId()})
		Tell(actor, `Log follow disabled.`)

		return true, nil
	}

	events.AddToQueue(events.Log{FollowAdd: user.ConnectionId(), Level: strings.ToUpper(args[0])})
	Tell(actor, `Log follow enabled. Use <ansi fg="command">syslogs off</ansi> to turn it off.`)

	return true, nil
}
