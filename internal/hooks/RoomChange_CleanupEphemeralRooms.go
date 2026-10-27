package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

//
// RoomChangeHandler waits for RoomChange events
//

func CleanupEphemeralRooms(e events.Event) events.ListenerReturn {

	evt := e.(events.RoomChange)

	// If this isn't a user changing rooms, just pass it along.
	if actors.AccountOf(evt.ActorId) == nil {
		return events.Continue
	}

	if rooms.IsEphemeralRoomId(evt.FromRoomId) {
		rooms.TryEphemeralCleanup(evt.FromRoomId)
	}

	return events.Continue
}
