package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// Suicide is a death by the actor's own hand, through the one death routine
// (internal/death), for players and NPCs alike.
func Suicide(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	death.Die(actor)
	return true, nil
}
