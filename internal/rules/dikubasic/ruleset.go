package dikubasic

import (
	"errors"
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Ruleset is the reference ruleset: GoMud's rules, moved behind the seam
// unchanged (Phase 5, decision 1). Its methods keep every random draw where
// the engine made it before, so the combat goldens hold.
type Ruleset struct{}

var _ rules.Ruleset = Ruleset{}

// New is the reference ruleset, for rules.Set.
func New() Ruleset { return Ruleset{} }

// TurnOrder puts actors with an account first, then NPCs, each group in the
// order given (players always swing before NPCs).
func (Ruleset) TurnOrder(fighters []*actors.Actor) []*actors.Actor {
	ordered := make([]*actors.Actor, 0, len(fighters))
	for _, a := range fighters {
		if accountOf(a) != nil {
			ordered = append(ordered, a)
		}
	}
	for _, a := range fighters {
		if accountOf(a) == nil {
			ordered = append(ordered, a)
		}
	}
	return ordered
}

// ResolveAttack is one round of attacker's attacks on defender.
func (Ruleset) ResolveAttack(attacker, defender *actors.Actor) rules.AttackResult {
	return calculateCombat(*attacker.Character, *defender.Character)
}

// WaitLines are the lines for a round the attacker spends waiting.
func (Ruleset) WaitLines(attacker, defender *actors.Actor) []perception.Message {
	return GetWaitMessages(items.Wait, attacker.Character, defender.Character)
}

// HatredRounds is two minutes of rounds less the attacker's Perception.
func (Ruleset) HatredRounds(attacker *actors.Actor) int {
	return configs.GetConfig().Timing.MinutesToRounds(2) - attacker.Character.Stats.Perception.ValueAdj
}

// FindsWeapon rolls d100 under the NPC's Perception.
func (Ruleset) FindsWeapon(npc *actors.Actor) bool {
	roll := util.Rand(100)

	util.LogRoll(`Look for weapon`, roll, npc.Character.Stats.Perception.ValueAdj)

	return roll < npc.Character.Stats.Perception.ValueAdj
}

// FleeEscapes: 30% flat, plus up to 70% from Speed against the blocker's.
func (Ruleset) FleeEscapes(fleer, blocker *actors.Actor) bool {
	// Stat comparison accounts for up to 70% of chance to flee.
	chanceIn100 := int(float64(fleer.Character.Stats.Speed.ValueAdj) / (float64(fleer.Character.Stats.Speed.ValueAdj) + float64(blocker.Character.Stats.Speed.ValueAdj)) * 70)
	chanceIn100 += 30

	roll := util.Rand(100)

	util.LogRoll(`Flee`, roll, chanceIn100)

	return roll < chanceIn100
}

// Consider is CombatOdds: the defender's rounds to kill over the attacker's.
func (Ruleset) Consider(attacker, defender *actors.Actor) float64 {
	return CombatOdds(*attacker.Character, *defender.Character)
}

// CanAttack allows any NPC target. An account target needs PvP allowed in
// the attacker's room and, when it differs, the defender's room (spec
// ruling 1: no shooting out of or into a safe room).
func (Ruleset) CanAttack(attacker, defender *actors.Actor) error {
	if accountOf(defender) == nil {
		return nil
	}
	// CanPvp checks only the attacker's room against the recovery room; a
	// safe room protects whoever stands in it, its private copies included.
	if rooms.IsDeathRecoveryRoom(defender.Character.RoomId) {
		return errors.New(`Fighting is not allowed here.`)
	}
	r := rooms.LoadRoom(attacker.Character.RoomId)
	if r == nil {
		return errors.New(`Fighting is not allowed here.`)
	}
	if err := r.CanPvp(attacker.Character, defender.Character); err != nil {
		return err
	}
	if defender.Character.RoomId != attacker.Character.RoomId {
		dr := rooms.LoadRoom(defender.Character.RoomId)
		if dr == nil {
			return errors.New(`Fighting is not allowed here.`)
		}
		if err := dr.CanPvp(attacker.Character, defender.Character); err != nil {
			return err
		}
	}
	return nil
}

// Condition: dead at 0 health or below, else up. The reference rules have
// no downed state (decision 7).
func (Ruleset) Condition(a *actors.Actor) rules.Condition {
	if a.Character.Health <= 0 {
		return rules.Dead
	}
	return rules.Up
}

// Regen in the death recovery room (or a private copy of it) is fast:
// recoveryRegen. Elsewhere it is GoMud's AutoHeal, moved behind the seam
// unchanged: an account, every 3rd round, out of a fight and not downed,
// regains HealthPerRound and ManaPerRound. NPCs do not regenerate there
// (slice 4 gives every actor the one rule).
func (Ruleset) Regen(a *actors.Actor, round uint64) rules.RegenResult {
	if rooms.IsDeathRecoveryRoom(a.Character.RoomId) {
		return recoveryRegen(a)
	}
	if accountOf(a) == nil || round%3 != 0 || a.Character.Aggro != nil {
		return rules.RegenResult{}
	}
	if a.Character.Health <= 0 {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: a.Character.HealthPerRound(), Mana: a.Character.ManaPerRound()}
}

// recoveryRegen is the recovery room's fast regen (spec ruling 2): upstream's
// deleted Death Recovery buff (24-death_recovery.js), 1d10 health and 1d10
// mana every round, health drawn first, with its lines. The amounts are what
// Heal will apply (capped at the maximums), so the lines say what happened.
func recoveryRegen(a *actors.Actor) rules.RegenResult {
	c := a.Character
	health := min(util.RollDice(1, 10), max(c.HealthMax.Value-c.Health, 0))
	mana := min(util.RollDice(1, 10), max(c.ManaMax.Value-c.Mana, 0))

	text := ``
	switch {
	case health > 0 && mana > 0:
		text = fmt.Sprintf(`The shadow realm heals you for <ansi fg="healing">%d damage</ansi> and restores <ansi fg="mana-100">%d mana</ansi>!`, health, mana)
	case health > 0:
		text = fmt.Sprintf(`The shadow realm heals you for <ansi fg="healing">%d damage</ansi>!`, health)
	case mana > 0:
		text = fmt.Sprintf(`The shadow realm restores <ansi fg="mana-100">%d mana</ansi>!`, mana)
	}
	if text == `` {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: health, Mana: mana, Messages: []perception.Message{
		{Kind: perception.Interface, Actor: a.Id, Audience: perception.ActorOnly, ActorText: text},
		{
			Kind:         perception.WorldEvent,
			Actor:        a.Id,
			Range:        perception.Room,
			Senses:       perception.Sight,
			ObserverText: `{actor} is recovering from a recent death.`,
		},
	}}
}
