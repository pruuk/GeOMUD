package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/term"
)

// Global chat room
func Broadcast(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// An actor without an account is never muted and never a moderator.
	sourceIsMod := false
	if user := actor.Controller().Account(); user != nil {
		if user.Muted {
			tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
			return true, nil
		}
		sourceIsMod = user.Role != accounts.RoleUser
	}

	msg := fmt.Sprintf(`<ansi fg="broadcast-prefix">(broadcast)</ansi> <ansi fg="%s">%s</ansi>: <ansi fg="broadcast-body">%s</ansi>`, nameColour(actor), actor.Character.Name, rest)

	events.AddToQueue(events.Broadcast{
		Text:            msg + term.CRLFStr,
		IsCommunication: true,
		SourceIsMod:     sourceIsMod,
	})

	events.AddToQueue(events.Communication{
		SourceActorId: actor.Character.Id,
		CommType:      `broadcast`,
		Name:          actor.Character.Name,
		Message:       rest,
	})

	return true, nil
}
