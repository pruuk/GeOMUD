package messaging

// The four socket writers (plan Task 2, decision 2). They were the
// internal/hooks listeners Message_SendMessage, Broadcast_SendToAll,
// RedrawPrompt_SendRedraw and WebClientCommand_SendWebClientCommand, moved
// here verbatim; hooks.RegisterListeners registers them in the same slots.
// Writes stay queued: events.Message, events.Broadcast, events.RedrawPrompt
// and events.WebClientCommand are the transport.

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

// WriteMessage writes an events.Message: to its UserId, and to the players
// of its RoomId at drain time minus UserId and ExcludeUserIds (a quiet
// message needs superhearing). Each write queues a prompt redraw.
func WriteMessage(e events.Event) events.ListenerReturn {

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

// WriteBroadcast writes an events.Broadcast to every active account, the
// screen-reader text to a screen-reader account when there is one, and
// queues a prompt redraw for each.
func WriteBroadcast(e events.Event) events.ListenerReturn {

	broadcast, typeOk := e.(events.Broadcast)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Broadcast", "Actual Type", e.Type())
		return events.Continue
	}

	textOut := ``
	if len(broadcast.Text) > 0 {
		textOut = templates.AnsiParse(broadcast.Text)
	}

	textOutSR := ``
	if len(broadcast.TextScreenReader) > 0 {
		textOutSR = templates.AnsiParse(broadcast.TextScreenReader)
	}

	for _, u := range accounts.GetAllActiveUsers() {

		events.AddToQueue(events.RedrawPrompt{UserId: u.UserId}, 100)

		if u.ScreenReader {

			if len(textOutSR) > 0 {

				if broadcast.SkipLineRefresh {
					connections.SendTo(
						[]byte(textOutSR),
						u.ConnectionId(),
					)
				} else {
					connections.SendTo(
						[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOutSR),
						u.ConnectionId(),
					)
				}

				continue
			}

		}

		if broadcast.SkipLineRefresh {
			connections.SendTo(
				[]byte(textOut),
				u.ConnectionId(),
			)
		} else {
			connections.SendTo(
				[]byte(term.AnsiMoveCursorColumn.String()+term.AnsiEraseLine.String()+textOut),
				u.ConnectionId(),
			)
		}
	}

	return events.Continue
}

// WritePrompt writes an account's command prompt; with OnlyIfChanged, only
// when it differs from the last one written that way.
func WritePrompt(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.RedrawPrompt)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "RedrawPrompt", "Actual Type", e.Type())
		return events.Cancel
	}

	if user := accounts.GetByUserId(evt.UserId); user != nil {

		newCmdPrompt := user.GetCommandPrompt()

		if evt.OnlyIfChanged {

			oldCmdPrompt := user.GetTempData(`cmdprompt`)

			// If the prompt hasn't changed, skip redrawing
			if oldCmdPrompt != nil && oldCmdPrompt.(string) == newCmdPrompt {
				return events.Continue
			}

			// save the new prompt for next time we want to check
			user.SetTempData(`cmdprompt`, newCmdPrompt)

		}

		pTxt := templates.AnsiParse(newCmdPrompt)
		connections.SendTo([]byte(pTxt), user.ConnectionId())

	}

	return events.Continue
}

// WriteWebClientCommand writes a web client command to a websocket
// connection; any other connection gets nothing.
func WriteWebClientCommand(e events.Event) events.ListenerReturn {

	cmd, typeOk := e.(events.WebClientCommand)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "WebClientCommand", "Actual Type", e.Type())
		return events.Continue
	}

	if !connections.IsWebsocket(cmd.ConnectionId) {
		return events.Cancel
	}

	connections.SendTo([]byte(cmd.Text), cmd.ConnectionId)

	return events.Continue
}
