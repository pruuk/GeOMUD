package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Say(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user := actor.Controller().Account(); user != nil && user.Muted {
		Tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	// {actor} reads "Someone" for a receiver who cannot see a hidden speaker.
	typed := messaging.Literal(rest)
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        actor.Id,
		Senses:       perception.Sound,
		Range:        perception.Room,
		ActorText:    fmt.Sprintf(`You say, "<ansi fg="saytext">%s</ansi>"`, typed),
		ObserverText: fmt.Sprintf(`{actor} says, "<ansi fg="saytext">%s</ansi>"`, typed),
	})

	messaging.Send(perception.Message{
		Kind:    perception.WorldEvent,
		Actor:   actor.Id,
		Senses:  perception.Sound,
		Faint:   true,
		Range:   perception.Nearby,
		FarText: `(From {direction}) You hear someone talking.`,
	})

	events.AddToQueue(events.Communication{
		SourceActorId: actor.Character.Id,
		CommType:      `say`,
		Name:          actor.Character.Name,
		Message:       rest,
	})

	return true, nil
}
