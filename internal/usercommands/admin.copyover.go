package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

/*
* Role Permissions:
* copyover			(All)
 */
func Copyover(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}
	Tell(actor, `Initiating copyover...`)
	if err := triggerCopyoverFunc(); err != nil {
		mudlog.Error("copyover command", "error", err)
		Tell(actor, `Copyover failed: `+err.Error())
	}
	return true, nil
}

// triggerCopyoverFunc is set by main at startup to point to the platform-specific
// triggerCopyover helper.
var triggerCopyoverFunc func() error = func() error { return nil }

// SetCopyoverFunc registers the platform-specific copyover trigger with the usercommands package.
func SetCopyoverFunc(f func() error) {
	triggerCopyoverFunc = f
}
