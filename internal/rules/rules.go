// Package rules is the seam between the engine and a game's rules. The engine
// calls the active Ruleset for every game rule and never imports a ruleset;
// main.go chooses one at boot with Set. A ruleset never sends text: it
// returns messages, and the engine sends them.
package rules

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/perception"
)

// AttackResult is what one round of an attacker's attacks did. The ruleset
// works it out; the engine applies DamageToTarget, queues BuffTarget and
// sends Messages in order.
type AttackResult struct {
	Hit            bool  // at least one attack connected
	Crit           bool  // at least one attack was a critical hit
	BuffTarget     []int // buffs a crit puts on the defender
	DamageToTarget int   // total damage dealt to the defender

	// Messages are the round's lines, one per swing (a dodge included), in
	// the order they happened; messaging resolves {actor} and {target} per
	// receiver.
	Messages []perception.Message
}

// Condition is what the ruleset says an actor's body is.
type Condition int

const (
	Up     Condition = iota // can act
	Downed                  // cannot act, not dead
	Dead                    // the engine runs its death (death.Die)
)

// RegenResult is what one round regenerates for an actor. The engine applies
// Health and Mana (either may be negative: a bleed) and sends Messages.
type RegenResult struct {
	Health, Mana int
	Messages     []perception.Message
}

// Changes reports whether r changes health or mana. It lives here, outside
// the engine packages, so the regen pass compares no health with a number
// (TestEngineAsksTheRulesetWhoIsDead).
func (r RegenResult) Changes() bool { return r.Health != 0 || r.Mana != 0 }

// Award is one account's share of a kill (Q1). After announcing the death
// the engine credits the kill (an NPC victim's), grants XP through the
// account (source "combat") and sends Messages, award by award.
type Award struct {
	UserId   int
	XP       int
	Messages []perception.Message
}

// DeathOutcome is what a death means. The engine carries it out
// (internal/death): it announces the death, moves the drops and respawns
// or ends the body.
type DeathOutcome struct {
	Revive        bool // the victim gets up instead of dying
	ReviveHealth  int  // health on revive
	Respawn       bool // true: back to life in the recovery room; false: the body ends
	RespawnHealth int  // health on respawn

	// DropItems leave the body, carried or worn, for its corpse (or the
	// floor when corpses are off); DropGold likewise.
	DropItems []items.Item
	DropGold  int

	Awards []Award // the killers' shares, in order; their Messages are the outcome's lines
}

// Ruleset is everything the engine asks a game's rules. A method that needs
// a random draw makes it itself. Phase 5 slice 2 holds the fight and slice 3 the
// round and death; later slices add display.
type Ruleset interface {
	// TurnOrder orders the actors that may take a combat turn this round.
	// It receives every actor with a body and returns them in turn order.
	TurnOrder(fighters []*actors.Actor) []*actors.Actor

	// ResolveAttack works out one round of attacker's attacks on defender.
	// It changes no health: the engine applies the result.
	ResolveAttack(attacker, defender *actors.Actor) AttackResult

	// WaitLines are the lines for a round attacker spends waiting to swing.
	WaitLines(attacker, defender *actors.Actor) []perception.Message

	// HatredRounds is how many rounds an NPC's groups hate the account of
	// an attacker who just attacked it.
	HatredRounds(attacker *actors.Actor) int

	// FindsWeapon reports whether an unarmed NPC carrying items looks for a
	// weapon to equip this round.
	FindsWeapon(npc *actors.Actor) bool

	// FleeEscapes reports whether fleer gets past blocker, one of the
	// actors fighting it.
	FleeEscapes(fleer, blocker *actors.Actor) bool

	// Consider is attacker's odds against defender: above 1 favours the
	// attacker, 1 is even, 0 or less is hopeless.
	Consider(attacker, defender *actors.Actor) float64

	// CanAttack is nil when attacker may start a fight with defender, or
	// the refusal to show the attacker.
	CanAttack(attacker, defender *actors.Actor) error

	// Condition is whether a is up, downed or dead. The engine compares no
	// health against a number: it asks this.
	Condition(a *actors.Actor) Condition

	// Regen is what a regenerates this round. It changes no health or mana:
	// the engine applies the result.
	Regen(a *actors.Actor, round uint64) RegenResult

	// OnDeath is what victim's death means. Its awards may change the
	// killers' characters (alignment); the victim's body is the engine's.
	OnDeath(victim *actors.Actor) DeathOutcome
}

// active is set once at boot, before the game loop starts, and swapped only
// by tests, which do not run in parallel with the loop.
var active Ruleset

// Set makes r the active ruleset. main.go calls it once at boot.
func Set(r Ruleset) { active = r }

// Get is the active ruleset. It panics when none is set: a boot that forgot
// to choose one must fail loudly, not fight by default rules.
func Get() Ruleset {
	if active == nil {
		panic("rules: no ruleset is set; main must call rules.Set at boot")
	}
	return active
}

// SetForTest makes r the active ruleset and returns the func that restores
// the previous one, for t.Cleanup.
func SetForTest(r Ruleset) (restore func()) {
	prev := active
	active = r
	return func() { active = prev }
}
