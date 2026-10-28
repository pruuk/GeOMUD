package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/term"
)

// Global chat room
func Broadcast(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// An actor without an account is never muted.
	if user := actor.Controller().Account(); user != nil && user.Muted {
		Tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	msg := fmt.Sprintf(`<ansi fg="broadcast-prefix">(broadcast)</ansi> <ansi fg="%s">%s</ansi>: <ansi fg="broadcast-body">%s</ansi>`, NameColour(actor), actor.Character.Name, rest)

	messaging.Notify(messaging.Notice{
		Text: msg + term.CRLFStr,
	})

	events.AddToQueue(events.Communication{
		SourceActorId: actor.Character.Id,
		CommType:      `broadcast`,
		Name:          actor.Character.Name,
		Message:       rest,
	})

	return true, nil
}
