package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/scripting"
)

//
// Runs the item's onFound/onLost script when a user gains or loses it
//

func ItemScriptEvents(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.ItemOwnership)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "ItemOwnership", "Actual Type", e.Type())
		return events.Cancel
	}

	// Only care about users for this stuff
	if evt.UserId == 0 {
		return events.Continue
	}

	if evt.Gained {

		scripting.TryItemScriptEvent(`onFound`, evt.Item, evt.UserId)

	} else {

		scripting.TryItemScriptEvent(`onLost`, evt.Item, evt.UserId)

	}

	return events.Continue
}
