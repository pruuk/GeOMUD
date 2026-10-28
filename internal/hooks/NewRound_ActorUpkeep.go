package hooks

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/behaviour"
	"github.com/GoMudEngine/GoMud/internal/configs"
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

// ActorUpkeep is every actor's round upkeep, before combat (slice 6): the
// same cooldowns, charm countdown, buff triggers and stat refresh for players
// and NPCs. An account also decays its alignment and reports changed vitals;
// an NPC's charm can expire, and an NPC at 0 health dies. Actors go in ID
// order.
func ActorUpkeep(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	//
	// Reduce existing hostility (if any)
	//
	mobs.ReduceHostility()

	for _, a := range actors.All() {
		if actors.Get(a.Id) == nil {
			continue // ended earlier in this pass
		}

		char := a.Character
		user := a.Controller().Account()
		mob := mobs.GetInstance(a.Id) // nil for a player

		// Roundtick any cooldowns
		char.Cooldowns.RoundTick()

		if user != nil {
			decayAlignment(user, evt.RoundNumber)
		}

		if char.Charmed != nil && char.Charmed.RoundsRemaining > 0 {
			char.Charmed.RoundsRemaining--
		}

		if triggeredBuffs := char.Buffs.Trigger(); len(triggeredBuffs) > 0 {
			triggeredBuffIds := []int{}
			for _, buff := range triggeredBuffs {
				triggeredBuffIds = append(triggeredBuffIds, buff.BuffId)
			}
			events.AddToQueue(events.BuffsTriggered{ActorId: a.Id, BuffIds: triggeredBuffIds})
		}

		// Do charm cleanup: an NPC's charm expires; a player's only counts down.
		if user == nil && mob != nil && char.IsCharmed() && char.Charmed.RoundsRemaining == 0 {
			cmd := char.Charmed.ExpiredCommand
			if charmedBy := char.RemoveCharm(); !charmedBy.IsZero() {
				if charmedUser := accounts.GetByBodyId(charmedBy); charmedUser != nil {
					charmedUser.Character.TrackCharmed(a.Id, false)
				}
			}
			runCharmExpiredCommand(mob, cmd)
			if mobs.GetInstance(a.Id) == nil {
				continue // rested or ended: nothing left to refresh
			}
		}

		// Recalculate all stats at the end of the round tick
		hBefore, mBefore := char.Health, char.Mana
		char.Validate()

		if user != nil {
			if char.Health != hBefore || char.Mana != mBefore {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
			}
			continue
		}

		if mob != nil && char.Health <= 0 {
			// Mob died
			mob.Command(`suicide`)
		}
	}

	return events.Continue
}

// decayAlignment moves an account's alignment toward neutral every
// AlignmentDecayRounds rounds and tells the player when its name changes.
// Accounts only: NPC alignment is a Phase 5 rule.
func decayAlignment(user *accounts.Account, round uint64) {
	decayRounds := int(configs.GetGamePlayConfig().AlignmentDecayRounds)
	if decayRounds <= 0 || round%uint64(decayRounds) != 0 {
		return
	}
	alignmentBefore := user.Character.AlignmentName()
	user.Character.DecayAlignment()
	if alignmentAfter := user.Character.AlignmentName(); alignmentAfter != alignmentBefore {
		before := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
		after := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
		user.SendText(fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, before, after))
		events.AddToQueue(events.AlignmentChanged{
			UserId:       user.UserId,
			AlignmentOld: alignmentBefore,
			AlignmentNew: alignmentAfter,
		})
	}
}
