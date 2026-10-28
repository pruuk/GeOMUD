package usercommands

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobcommands"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

/*
* Role Permissions:
* command 				(All)
 */
func AdminCommand(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	// args should look like one of the following:
	// target buffId - put buff on target if in the room
	// buffId - put buff on self
	// search searchTerm - search for buff by name, display results
	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		mobCommands := mobcommands.GetAllMobCommands()

		infoOutput, _ := templates.Process("admincommands/help/command.command", mobCommands, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	searchName := args[0]
	args = args[1:]
	cmd := strings.TrimPrefix(rest, searchName+` `)
	//cmd := strings.Join(args, ` `)

	playerId, mobId := room.FindByName(searchName)

	// Use the index for how many turns to defer the extra commands
	readyTurn := util.GetTurnCount()
	for _, oneCmd := range strings.Split(cmd, `;`) {
		if !mobId.IsZero() {

			events.AddToQueue(events.Input{
				ActorId:   mobId,
				InputText: oneCmd,
				ReadyTurn: readyTurn,
			})

		} else if playerId > 0 {

			if targetUser := accounts.GetByUserId(playerId); targetUser != nil {
				events.AddToQueue(events.Input{
					ActorId:   targetUser.Character.Id,
					InputText: oneCmd,
					ReadyTurn: readyTurn,
				})
			}

		}
		readyTurn++
	}

	return true, nil
}
