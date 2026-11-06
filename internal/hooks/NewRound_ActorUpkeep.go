package hooks

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
)

// ActorUpkeep is every actor's round upkeep, before combat (slice 6): the
// same cooldowns, buff triggers and stat refresh for players
// and NPCs. An account also decays its alignment and reports changed vitals;
// an NPC at 0 health dies. Actors go in ID order.
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

		if triggeredBuffs := char.Buffs.Trigger(); len(triggeredBuffs) > 0 {
			triggeredBuffIds := []int{}
			for _, buff := range triggeredBuffs {
				triggeredBuffIds = append(triggeredBuffIds, buff.BuffId)
			}
			events.AddToQueue(events.BuffsTriggered{ActorId: a.Id, BuffIds: triggeredBuffIds})
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
		messaging.Tell(user.Character.Id, fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, before, after))
		events.AddToQueue(events.AlignmentChanged{
			UserId:       user.UserId,
			AlignmentOld: alignmentBefore,
			AlignmentNew: alignmentAfter,
		})
	}
}
