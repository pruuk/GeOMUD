package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Checks whether their level is too high for a guide
func Message_SendMessage(e events.Event) events.ListenerReturn {

	message, typeOk := e.(events.Message)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Message", "Actual Type", e.Type())
		return events.Continue
	}

	if message.UserId > 0 {

		if user := accounts.GetByUserId(message.UserId); user != nil {

			textOut := templates.AnsiParse(message.Text)
			if user.ScreenReader {
				textOut = util.StripCharsForScreenReaders(textOut)
			}
			connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOut), user.ConnectionId())

			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)

		}
	}

	if message.RoomId > 0 {

		room := rooms.LoadRoom(message.RoomId)
		if room == nil {
			return events.Continue
		}

		for _, userId := range room.GetPlayers() {
			skip := false

			if message.UserId == userId {
				continue
			}

			exLen := len(message.ExcludeUserIds)
			if exLen > 0 {
				for _, excludeId := range message.ExcludeUserIds {
					if excludeId == userId {
						skip = true
						break
					}
				}
			}

			if skip {
				continue
			}

			if user := accounts.GetByUserId(userId); user != nil {

				// If this is a quiet message, make sure the player can hear it
				if message.IsQuiet {
					if !user.Character.HasBuffFlag("superhearing") {
						continue
					}
				}

				textOut := templates.AnsiParse(message.Text)
				if user.ScreenReader {
					textOut = util.StripCharsForScreenReaders(textOut)
				}

				connections.SendTo([]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOut), user.ConnectionId())

				events.AddToQueue(events.RedrawPrompt{UserId: user.UserId}, 100)

			}
		}

	}
	return events.Continue

}
