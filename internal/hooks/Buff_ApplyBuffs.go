package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

//
// Applies triggered buffs
//

func ApplyBuffs(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.Buff)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Buff", "Actual Type", e.Type())
		return events.Cancel
	}

	//mudlog.Debug(`Event`, `type`, evt.Type(), `ActorId`, evt.ActorId, `BuffId`, evt.BuffId)

	buffInfo := buffs.GetBuffSpec(evt.BuffId)
	if buffInfo == nil {
		return events.Cancel
	}

	var targetChar *characters.Character

	// An account means a player's body; otherwise the body is a mob.
	buffUser := actors.AccountOf(evt.ActorId)
	isMob := buffUser == nil

	if isMob {

		buffMob := mobs.GetInstance(evt.ActorId)
		if buffMob == nil {
			return events.Cancel
		}

		targetChar = &buffMob.Character

	} else {

		targetChar = buffUser.Character
	}

	if evt.BuffId < 0 {
		targetChar.RemoveBuff(buffInfo.BuffId * -1)
		return events.Continue
	}

	// Apply the buff
	targetChar.AddBuff(evt.BuffId, false)

	//
	// If the buff calls for an immediate triggering
	//
	if buffInfo.TriggerNow {

		if isMob && targetChar.Health <= 0 {
			// Mob died
			events.AddToQueue(events.Input{
				ActorId:   evt.ActorId,
				InputText: `suicide`,
			})

		}
	}

	events.AddToQueue(events.BuffsTriggered{ActorId: evt.ActorId, BuffIds: []int{evt.BuffId}})

	if !isMob {
		events.AddToQueue(events.BuffsAdded{ActorId: evt.ActorId, BuffIds: []int{evt.BuffId}})
	}

	return events.Continue
}
