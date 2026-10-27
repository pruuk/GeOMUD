package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// This is a no-op, does nothing
func Noop(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {
	return true, nil
}
