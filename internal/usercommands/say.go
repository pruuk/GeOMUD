package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Say(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user := actor.Controller().Account(); user != nil && user.Muted {
		Tell(actor, `You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	isSneaking := actor.Character.HasBuffFlag("hidden")
	isDrunk := actor.Character.HasBuffFlag("drunk")

	if isDrunk {
		// modify the text to look like it's the speech of a drunk person
		rest = drunkify(rest)
	}

	if isSneaking {
		room.SendTextCommunication(fmt.Sprintf(`someone says, "<ansi fg="saytext">%s</ansi>"`, rest), ExcludeId(actor))
	} else {
		room.SendTextCommunication(fmt.Sprintf(`<ansi fg="%s">%s</ansi> says, "<ansi fg="saytext">%s</ansi>"`, NameColour(actor), actor.Character.Name, rest), ExcludeId(actor))
	}

	Tell(actor, fmt.Sprintf(`You say, "<ansi fg="saytext">%s</ansi>"`, rest))

	room.SendTextToExits(`You hear someone talking.`, true)

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
