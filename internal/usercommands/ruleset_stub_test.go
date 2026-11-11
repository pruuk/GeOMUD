package usercommands

import (
	"errors"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pacifist is the active ruleset refusing every fight.
type pacifist struct{ rules.Ruleset }

func (pacifist) CanAttack(a, d *actors.Actor) error { return errors.New("STUB REFUSES") }

// attack asks the ruleset before any fight, NPC targets included.
func TestStubRulesetRefusesAttack(t *testing.T) {
	_, player, room := goSetup(t) // the NPC is "Rat"
	t.Cleanup(rules.SetForTest(pacifist{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := Attack("Rat", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"STUB REFUSES\n"}, told)
	assert.Nil(t, player.Character.Aggro)
}

// optimist is the active ruleset rating a fight 5 to 1 when the attacker is
// the one with an account, and 0.1 the other way round, so swapped
// arguments read another label.
type optimist struct{ rules.Ruleset }

func (optimist) Consider(a, d *actors.Actor) float64 {
	if a.Controller() != nil && a.Controller().Account() != nil {
		return 5
	}
	return 0.1
}

// consider reads the ruleset's odds, asked with the player as attacker.
func TestStubRulesetDecidesConsider(t *testing.T) {
	_, player, room := goSetup(t) // the NPC is "Rat"
	t.Cleanup(rules.SetForTest(optimist{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := Consider("Rat", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, told, 2)
	assert.Contains(t, told[1], "Very Favorable")
}

// Considering a player names them in the player colour.
func TestConsiderAPlayerUsesThePlayerColour(t *testing.T) {
	_, player, room := goSetup(t)
	addFoundPlayer(t, 9203, "Target", room)
	t.Cleanup(rules.SetForTest(optimist{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := Consider("Target", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, told, 2)
	assert.Contains(t, told[0], `<ansi fg="username">Target</ansi>`)
}

// downedAlways is the active ruleset with every actor downed.
type downedAlways struct{ rules.Ruleset }

func (downedAlways) Condition(a *actors.Actor) rules.Condition { return rules.Downed }

// A downed actor cannot act: commands not allowed when downed are refused,
// though the reference ruleset calls the player (10 health) up.
func TestStubRulesetDownedActorCannotAct(t *testing.T) {
	_, player, _ := goSetup(t)
	t.Cleanup(rules.SetForTest(downedAlways{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := TryCommand("get", "stick", player.Id, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"You are unable to do that while downed.\n"}, told)

	told = playerTexts(t, player, func() {
		_, err := TryCommand("north", "", player.Id, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"You are unable to do that while downed.\n"}, told)
	assert.Equal(t, testRoomId, player.Character.RoomId)
}
