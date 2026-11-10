// Package rules is the seam between the engine and a game's rules. The engine
// calls the active Ruleset for every game rule and never imports a ruleset;
// main.go chooses one at boot with Set. A ruleset never sends text: it
// returns messages, and the engine sends them.
package rules

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
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

// Ruleset is everything the engine asks a game's rules. A method that needs
// a random draw makes it itself. Phase 5 slice 2 holds the fight; later
// slices add the round, death and display.
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
