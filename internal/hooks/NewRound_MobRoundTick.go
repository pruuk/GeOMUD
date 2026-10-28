package hooks

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/behaviour"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
)

// runCharmExpiredCommand runs a charm's ExpiredCommand, a `;`-separated list.
// `rest <reason>` is the behaviour action (it is no longer a command). When the
// list rests the NPC, the other parts run now through the shared command path,
// in order, so that an emote is seen before the NPC leaves memory (a queued
// command would run after it was gone). Without a rest, parts are queued as the
// NPC's input as before.
func runCharmExpiredCommand(mob *mobs.Mob, cmd string) {
	if cmd == `` {
		return
	}

	parts := []string{}
	resting := false
	for _, part := range strings.Split(cmd, `;`) {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		parts = append(parts, part)
		if verb, _, _ := strings.Cut(part, ` `); strings.EqualFold(verb, `rest`) {
			resting = true
		}
	}

	for _, part := range parts {
		verb, args, _ := strings.Cut(part, ` `)
		switch {
		case strings.EqualFold(verb, `rest`):
			behaviour.Rest(mob, strings.TrimSpace(args))
			return // the NPC is out of memory (or ended): nothing after it can run
		case resting:
			if _, err := usercommands.TryCommand(strings.ToLower(verb), strings.TrimSpace(args), mob.Character.Id, 0); err != nil {
				mudlog.Warn("charm expired", "mobname", mob.Character.Name, "command", part, "error", err)
			}
		default:
			mob.Command(part)
		}
	}
}

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
			runCharmExpiredCommand(mob, cmd)
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
