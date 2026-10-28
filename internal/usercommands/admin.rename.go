package usercommands

import (
	"fmt"
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
* rename 				(All)
 */
func Rename(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 2 {
		// send some sort of help info?
		infoOutput, _ := templates.Process("admincommands/help/command.rename", nil, user.UserId)
		Tell(actor, infoOutput)
		return true, nil
	}

	// Check whether the user has an item in their inventory that matches
	matchItem, found := user.Character.FindInBackpack(args[0])
	rest = strings.Join(args[1:], " ")

	if !found {
		Tell(actor, fmt.Sprintf("You don't have a %s to rename.", rest))
	} else {
		// Swap the item location
		user.Character.RemoveItem(matchItem)
		oldNameSimple := matchItem.DisplayName()
		oldName := matchItem.DisplayName()

		if len(args) > 2 {
			matchItem.Rename(strings.TrimSpace(args[1]), strings.TrimSpace(args[2]))
		} else {
			matchItem.Rename(strings.TrimSpace(args[1]))
		}

		matchItem.Validate()

		user.Character.StoreItem(matchItem)

		messaging.Send(perception.Message{
			Kind:         perception.WorldEvent,
			Actor:        actor.Id,
			Senses:       perception.Sight | perception.Sound,
			Range:        perception.Room,
			ActorText:    fmt.Sprintf(`You chant softly and wave your hand over the <ansi fg="item">%s</ansi>. Success! It's now a <ansi fg="item">%s</ansi>`, oldNameSimple, matchItem.DisplayName()),
			ObserverText: fmt.Sprintf(`{actor} chants softly and waves their hand over <ansi fg="item">%s</ansi>, causing it to glow briefly.`, oldName),
		})
	}

	return true, nil
}
