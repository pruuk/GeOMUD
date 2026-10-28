package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
)

// ActionPoints gives every live body one action point per turn, up to its
// maximum (Phase 4b slice 4 decision 1: one mechanism for players and NPCs).
func ActionPoints(e events.Event) events.ListenerReturn {
	for _, a := range actors.All() {
		c := a.Character
		if c == nil {
			continue
		}
		c.ActionPoints += 1
		if c.ActionPoints > c.ActionPointsMax.Value {
			c.ActionPoints = c.ActionPointsMax.Value
		}
	}

	return events.Continue
}
