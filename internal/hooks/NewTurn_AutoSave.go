// Round ticks for players
package hooks

import (
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/util"
)

//
// Autosaves users/rooms every so often
//

func AutoSave(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.NewTurn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.TurnNumber%uint64(configs.GetTimingConfig().TurnsPerAutoSave()) == 0 {

		totalTimeStart := time.Now()
		defer func() {
			util.TrackTime(`AutoSave`, time.Since(totalTimeStart).Seconds())
		}()

		//////////////////////////////////////////
		// SAVE ALL USERS
		//////////////////////////////////////////
		messaging.Notify(messaging.Notice{Text: `Saving users...`})

		accounts.SaveAllUsers(true)

		messaging.Notify(messaging.Notice{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		//////////////////////////////////////////
		// SAVE ALL ROOMS
		//////////////////////////////////////////
		messaging.Notify(messaging.Notice{Text: `Saving rooms...`})

		rooms.SaveAllRooms()

		messaging.Notify(messaging.Notice{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

		//////////////////////////////////////////
		// SAVE ALL PLUGINS
		//////////////////////////////////////////
		messaging.Notify(messaging.Notice{Text: `Saving other...`})
		// Save plugin states if applicable
		plugins.Save()

		messaging.Notify(messaging.Notice{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

	}

	return events.Continue
}
