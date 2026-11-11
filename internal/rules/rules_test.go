package rules

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake is a Ruleset whose methods are never called; it only has to exist.
type fake struct{ name string }

func (fake) TurnOrder(f []*actors.Actor) []*actors.Actor       { return f }
func (fake) ResolveAttack(a, d *actors.Actor) AttackResult     { return AttackResult{} }
func (fake) WaitLines(a, d *actors.Actor) []perception.Message { return nil }
func (fake) HatredRounds(a *actors.Actor) int                  { return 0 }
func (fake) FindsWeapon(npc *actors.Actor) bool                { return false }
func (fake) FleeEscapes(fleer, blocker *actors.Actor) bool     { return true }
func (fake) Consider(a, d *actors.Actor) float64               { return 1 }
func (fake) CanAttack(a, d *actors.Actor) error                { return nil }
func (fake) Condition(a *actors.Actor) Condition               { return Up }
func (fake) Regen(a *actors.Actor, round uint64) RegenResult   { return RegenResult{} }

func TestGetPanicsWithNoRuleset(t *testing.T) {
	restore := SetForTest(nil)
	defer restore()
	assert.PanicsWithValue(t, "rules: no ruleset is set; main must call rules.Set at boot", func() { Get() })
}

func TestSetForTestRestoresThePreviousRuleset(t *testing.T) {
	restoreOuter := SetForTest(fake{name: "outer"})
	defer restoreOuter()

	restore := SetForTest(fake{name: "inner"})
	require.Equal(t, fake{name: "inner"}, Get())
	restore()
	require.Equal(t, fake{name: "outer"}, Get())
}

func TestRegenResultChangesWhenHealthOrManaDoes(t *testing.T) {
	assert.False(t, RegenResult{}.Changes())
	assert.False(t, RegenResult{Messages: []perception.Message{{ActorText: "a line"}}}.Changes(), "a line alone changes nothing")
	assert.True(t, RegenResult{Health: 1}.Changes())
	assert.True(t, RegenResult{Health: -1}.Changes(), "a bleed")
	assert.True(t, RegenResult{Mana: 2}.Changes())
}
