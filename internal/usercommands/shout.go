package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Shout(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user := actor.Controller().Account(); user != nil && user.Muted {
		Tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	isDrunk := actor.Character.HasBuffFlag("drunk")

	rest = strings.ToUpper(rest)

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}

	// {actor} reads "Someone" for a receiver who cannot see a hidden shouter.
	// Rooms next to this one (exits, temporary exits, mutator exits) read the
	// far text with the direction back toward this room.
	typed := messaging.Literal(rest)
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        actor.Id,
		Senses:       perception.Sound,
		Range:        perception.Nearby,
		ActorText:    fmt.Sprintf(`You shout, "<ansi fg="yellow">%s</ansi>"`, typed),
		ObserverText: fmt.Sprintf(`{actor} shouts, "<ansi fg="yellow">%s</ansi>"`, typed),
		FarText:      fmt.Sprintf(`Someone shouts from the {direction} direction, "<ansi fg="yellow">%s</ansi>"`, typed),
	})

	return true, nil
}
