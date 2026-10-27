package hooks

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
)

func MobRoundTick(e events.Event) events.ListenerReturn {

	//
	// Reduce existing hostility (if any)
	//
	mobs.ReduceHostility()

	//
	// Do mob round maintenance
	//
	for _, mobId := range mobs.GetAllMobIds() {

		mob := mobs.GetInstance(mobId)

		if mob == nil {
			continue
		}

		// Roundtick any cooldowns
		mob.Character.Cooldowns.RoundTick()

		if mob.Character.Charmed != nil && mob.Character.Charmed.RoundsRemaining > 0 {
			mob.Character.Charmed.RoundsRemaining--
		}

		if triggeredBuffs := mob.Character.Buffs.Trigger(); len(triggeredBuffs) > 0 {

			triggeredBuffIds := []int{}
			for _, buff := range triggeredBuffs {
				triggeredBuffIds = append(triggeredBuffIds, buff.BuffId)
			}

			events.AddToQueue(events.BuffsTriggered{ActorId: mobId, BuffIds: triggeredBuffIds})
		}

		// Do charm cleanup
		if mob.Character.IsCharmed() && mob.Character.Charmed.RoundsRemaining == 0 {
			cmd := mob.Character.Charmed.ExpiredCommand
			if charmedBy := mob.Character.RemoveCharm(); !charmedBy.IsZero() {
				if charmedUser := accounts.GetByBodyId(charmedBy); charmedUser != nil {
					charmedUser.Character.TrackCharmed(mob.Character.Id, false)
				}
			}
			if cmd != `` {
				cmds := strings.Split(cmd, `;`)
				for _, cmd := range cmds {
					cmd = strings.TrimSpace(cmd)
					if len(cmd) > 0 {
						mob.Command(cmd)
					}
				}
			}
		}

		// Recalculate all stats at the end of the round tick
		mob.Character.Validate()

		if mob.Character.Health <= 0 {
			// Mob died
			mob.Command(`suicide`)
		}

	}

	return events.Continue
}
