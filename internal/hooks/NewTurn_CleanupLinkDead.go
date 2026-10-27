// Round ticks for players
package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

//
// Cleans up link-dead users
// Link-dead users are users who have disconnected but their user/character is still in game.
//

func CleanupLinkDead(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.NewTurn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return events.Cancel
	}

	et := configs.GetTimingConfig()
	gp := configs.GetNetworkConfig()

	expTurns := uint64(et.SecondsToTurns(int(gp.LinkDeadSeconds)))

	if expTurns < evt.TurnNumber {

		expired := accounts.GetExpiredLinkDeadUsers(evt.TurnNumber - expTurns)

		if len(expired) > 0 {

			mudlog.Info("Expired Link-Dead Users", "count", len(expired))

			for _, userId := range expired {
				events.AddToQueue(events.System{
					Command:     `leaveworld`,
					Data:        userId,
					Description: `Link-Dead Expired`,
				})
			}

		}
	}

	return events.Continue
}
