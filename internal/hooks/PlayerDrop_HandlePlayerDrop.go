package hooks

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

//
// Some clean up
//

func HandlePlayerDrop(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.PlayerDrop)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerDrop", "Actual Type", e.Type())
		return events.Cancel
	}

	user := accounts.GetByUserId(evt.UserId)
	if user == nil {
		mudlog.Error("HandlePlayerDrop", "error", fmt.Sprintf(`user %d not found`, evt.UserId))
		return events.Cancel
	}

	room := rooms.LoadRoom(evt.RoomId)
	if room == nil {
		messaging.Tell(user.Character.Id, `<ansi fg="red">you drop to the ground!</ansi>`)
		return events.Continue
	}

	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        user.Character.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ActorText:    `<ansi fg="red">you drop to the ground!</ansi>`,
		ObserverText: `{actor} <ansi fg="red">drops to the ground!</ansi>`,
	})

	return events.Continue
}
