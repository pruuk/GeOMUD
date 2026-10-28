package hooks

import (
	"time"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// ActorTick runs every actor's controller for the round, after combat and
// healing (slice 6): an NPC's behaviour (idle, wander, its path) runs from
// BehaviourController.Tick; a player's controller does nothing. Actors go in
// ID order. actors.All is a copy, so a tick may rest or end an NPC.
func ActorTick(e events.Event) events.ListenerReturn {
	tStart := time.Now()
	for _, a := range actors.All() {
		if actors.Get(a.Id) == nil {
			continue // ended earlier in this pass
		}
		a.Controller().Tick()
	}
	util.TrackTime(`ActorTick()`, time.Since(tStart).Seconds())
	return events.Continue
}
