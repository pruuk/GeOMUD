package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setPvP sets GamePlay.PVP.Enabled for one test, with no minimum level (the
// shipped config has 15). AddOverlayOverrides will not replace a key that is
// already set, and RestoreOverrides resets nothing it is not given, so this
// names each key's old value on the way back, as setConfig in
// internal/hooks does.
func setPvP(t *testing.T, setting string) {
	t.Helper()
	oldEnabled := string(configs.GetPVPConfig().Enabled)
	oldMin := int(configs.GetPVPConfig().MinimumLevel)
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["GamePlay.PVP.Enabled"] = setting
	next["GamePlay.PVP.MinimumLevel"] = 0
	require.NoError(t, configs.RestoreOverrides(next))
	require.Equal(t, setting, string(configs.GetPVPConfig().Enabled), "the override took")
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["GamePlay.PVP.Enabled"] = oldEnabled
		back["GamePlay.PVP.MinimumLevel"] = oldMin
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

// addFoundPlayer is addPlayer, registered so FindByName and GetByUserId see it.
func addFoundPlayer(t *testing.T, userId int, name string, room *rooms.Room) *actors.Actor {
	t.Helper()
	a := addPlayer(t, userId, name, room)
	accounts.SetTestUser(a.Controller().Account())
	t.Cleanup(func() { accounts.RemoveTestUser(userId) })
	return a
}

func TestAttackAPlayerRefusedWhenPvPDisabled(t *testing.T) {
	_, player, room := goSetup(t)
	addFoundPlayer(t, 9203, "Target", room)
	setPvP(t, configs.PVPDisabled)

	told := playerTexts(t, player, func() {
		_, err := Attack("Target", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"PVP is disabled.\n"}, told)
	assert.Nil(t, player.Character.Aggro)
}

// PvP settings do not touch fighting an NPC. This pins CanAttack's argument
// order: swapped, the player becomes the defender with an account and PvP
// rules would refuse the rat.
func TestAttackAnNPCAllowedWhenPvPDisabled(t *testing.T) {
	_, player, room := goSetup(t) // the NPC is "Rat"
	setPvP(t, configs.PVPDisabled)

	told := playerTexts(t, player, func() {
		_, err := Attack("Rat", player, room, 0)
		require.NoError(t, err)
	})
	require.NotEmpty(t, told)
	assert.Contains(t, told[0], "You prepare to fight")
	assert.NotNil(t, player.Character.Aggro)
}
