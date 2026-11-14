package rules_test

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/rules/dikubasic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false) // a body's names ask buffs, which log
	os.Exit(m.Run())
}

// downedAtFive is the reference rules with a downed state: from 5 health
// down to 1 a body is downed (not up, not dead).
type downedAtFive struct{ rules.Ruleset }

func (downedAtFive) Condition(a *actors.Actor) rules.Condition {
	switch {
	case a.Character.Health <= 0:
		return rules.Dead
	case a.Character.Health <= 5:
		return rules.Downed
	}
	return rules.Up
}

// woundedBody is a body at 5 of 10 health with a killer on record, wearing
// a ring that can never be removed.
func woundedBody() (*characters.Character, items.Item) {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = "Ann"
	c.HealthMax.Value, c.Health = 10, 5
	c.KilledBy = ids.New(ids.NoCell)
	ring := items.Item{ItemId: 9, Id: ids.New(ids.NoCell), CanNeverBeRemoved: true,
		Spec: &items.ItemSpec{ItemId: 9, Name: "ring", Type: items.Ring}}
	c.Equipment.Ring = ring
	return c, ring
}

// Phase 5 slice 4.1: characters asks the ruleset whether a body is up
// (characters.IsUp), for the downed display, the killer record and bound
// gear, whether or not a live actor holds the body.
func TestCharactersAskTheRulesetWhetherABodyIsUp(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rules rules.Ruleset
		up    bool
	}{
		{"downed at 5", downedAtFive{dikubasic.New()}, false},
		{"reference rules at 5", dikubasic.New(), true},
	} {
		for _, live := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: ", bare body", true: ", live actor"}[live], func(t *testing.T) {
				t.Cleanup(rules.SetForTest(tc.rules))
				c, ring := woundedBody()
				if live {
					acc := &accounts.Account{UserId: 9701, Character: c}
					require.NotNil(t, actors.Spawn(c, actors.NewConnectionController(acc)))
					t.Cleanup(func() { actors.Remove(c.Id) })
				}

				name := c.GetPlayerName(ids.ID{}, characters.RenderHealth)
				c.ForgetKillerIfUp()
				removed := c.RemoveFromBody(ring)

				if tc.up {
					assert.Empty(t, name.Suffix)
					assert.Contains(t, name.Adjectives, "50%", "health shows while up")
					assert.NotContains(t, name.Adjectives, "downed")
					assert.True(t, c.KilledBy.IsZero(), "up: the killer record goes")
					assert.False(t, removed, "up: bound gear stays on")
				} else {
					assert.Equal(t, "downed", name.Suffix)
					assert.Contains(t, name.Adjectives, "downed")
					assert.NotContains(t, name.Adjectives, "50%", "no health while downed")
					assert.False(t, c.KilledBy.IsZero(), "downed: the killer record stays")
					assert.True(t, removed, "downed: bound gear comes off")
				}
			})
		}
	}
}

// A copy of a live body (a corpse holds one) is asked about itself, not
// about the live body with the same ID.
func TestIsUpAsksAboutTheBodyItIsGiven(t *testing.T) {
	t.Cleanup(rules.SetForTest(downedAtFive{dikubasic.New()}))
	c, _ := woundedBody()
	c.Health = 10
	acc := &accounts.Account{UserId: 9702, Character: c}
	require.NotNil(t, actors.Spawn(c, actors.NewConnectionController(acc)))
	t.Cleanup(func() { actors.Remove(c.Id) })

	copied := *c
	copied.Health = 3
	assert.True(t, characters.IsUp(c), "the live body is up")
	assert.False(t, characters.IsUp(&copied), "its copy at 3 health is downed")
}

// rules.Set and rules.SetForTest wire characters.IsUp.
func TestSetWiresCharactersIsUp(t *testing.T) {
	saved := characters.IsUp
	t.Cleanup(func() { characters.IsUp = saved })
	t.Cleanup(rules.SetForTest(nil)) // Set below changes the active ruleset; this puts the previous one back

	characters.IsUp = nil
	rules.Set(dikubasic.New())
	require.NotNil(t, characters.IsUp, "Set")

	characters.IsUp = nil
	rules.SetForTest(dikubasic.New())()
	require.NotNil(t, characters.IsUp, "SetForTest")
}
