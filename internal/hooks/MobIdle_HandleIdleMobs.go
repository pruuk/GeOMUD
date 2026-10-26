package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/scripting"
	"github.com/GoMudEngine/GoMud/internal/util"
)

//
// Handles default mob idle behavior
//

func HandleIdleMobs(e events.Event) events.ListenerReturn {

	evt := e.(events.MobIdle)

	mob := mobs.GetInstance(evt.MobInstanceId)
	if mob == nil {
		return events.Cancel
	}

	isCharmed := mob.Character.IsCharmed()

	// if a mob shouldn't be allowed to leave their area (via wandering)
	// but has somehow been displaced, such as pulling through combat, spells, or otherwise
	// tell them to path back home
	if mob.MaxWander == 0 && mob.Character.RoomId != mob.HomeRoomId {
		if !isCharmed {
			mob.Command("pathto home")
		}
	}

	// If they have idle commands, maybe do one of them?
	handled, _ := scripting.TryMobScriptEvent("onIdle", mob.InstanceId, 0, ``, nil)
	if !handled {

		if isCharmed {
			return events.Continue
		}

		if mob.MaxWander > -1 && mob.WanderCount > mob.MaxWander {
			// Not charmed and far from home, and should never leave home.
			// So go home.
			mob.Command(`pathto home`)
			return events.Continue
		}

		if util.Rand(100) < mob.ActivityLevel {
			if idleCmd := mob.GetIdleCommand(); idleCmd != `` {
				mob.Command(idleCmd)
			}
		}

	}

	return events.Continue
}
