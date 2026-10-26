package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/scripting"
)

//
// RoomChangeHandler waits for RoomChange events
//

func CleanupEphemeralRooms(e events.Event) events.ListenerReturn {

	evt := e.(events.RoomChange)

	// If this isn't a user changing rooms, just pass it along.
	if evt.UserId == 0 {
		return events.Continue
	}

	if rooms.IsEphemeralRoomId(evt.FromRoomId) {
		removedRoomIds := rooms.TryEphemeralCleanup(evt.FromRoomId)
		if len(removedRoomIds) > 0 {
			scripting.PruneRoomVMs(removedRoomIds...)
		}
	}

	return events.Continue
}
