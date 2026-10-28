package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Say(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user := actor.Controller().Account(); user != nil && user.Muted {
		Tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	isDrunk := actor.Character.HasBuffFlag("drunk")

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
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
		FarText: `You hear someone talking.`,
	})

	events.AddToQueue(events.Communication{
		SourceActorId: actor.Character.Id,
		CommType:      `say`,
		Name:          actor.Character.Name,
		Message:       rest,
	})

	return true, nil
}

func drunkify(sentence string) string {

	var drunkSentence strings.Builder
	isStartOfWord := true
	sentenceLength := len(sentence)
	insertedHiccup := false

	for i, char := range sentence {
		// Randomly decide whether to modify the character
		if util.Rand(10) < 3 || (!insertedHiccup || i == sentenceLength-1) {
			switch char {
			case 's':
				if isStartOfWord {
					drunkSentence.WriteString("sss")
				} else {
					drunkSentence.WriteString("sh")
				}
			case 'S':
				drunkSentence.WriteString("Sh")
			default:
				drunkSentence.WriteRune(char)
			}

			// Insert a hiccup in the middle of the sentence
			if !insertedHiccup && i >= sentenceLength/2 {
				drunkSentence.WriteString(" *hiccup* ")
				insertedHiccup = true
			}
		} else {
			drunkSentence.WriteRune(char)
		}

		// Update isStartOfWord based on spaces and punctuation
		if char == ' ' || char == '.' || char == '!' || char == '?' || char == ',' {
			isStartOfWord = true
		} else {
			isStartOfWord = false
		}
	}

	return drunkSentence.String()
}
