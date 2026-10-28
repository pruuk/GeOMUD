package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// Default is a special command that tries to contextually pick a default action for a room.
// The failover is to "look"
func Default(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// If a storage location, "storage"
	if room.HasTag(`storage`) {
		// Dispatch via the event queue so we do not create an init cycle
		// with the storage module's RegisterCommand call.
		queue(actor, `storage`)
		return true, nil
	}

	// Default to "look"
	Look(``, actor, room, flags)

	return true, nil
}
