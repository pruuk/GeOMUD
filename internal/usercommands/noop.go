package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// This is a no-op, does nothing
func Noop(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	return true, nil
}
