package usercommands

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerTargetedByNPCInRoomIsRefusedQuit(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	require.Nil(t, player.Character.Aggro, "the player has not attacked back")

	var handled bool
	told := playerTexts(t, player, func() {
		var err error
		handled, err = Quit("", player, room, 0)
		require.NoError(t, err)
	})
	assert.True(t, handled)
	assert.Equal(t, []string{"You're too busy to quit right now!\n"}, told)
}

func TestQuitInPeaceTellsTheQuitterAndRoomAboutTheMeditation(t *testing.T) {
	_, player, room := goSetup(t)
	player.Character.Name = "Ann"
	reading(t, player.Controller().Account())
	addReader(t, observerId, testRoomId)
	quitterId := player.Controller().Account().UserId

	reads := captureReads(t, func() {
		handled, err := Quit("", player, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})

	quitter := strings.Join(reads[quitterId], "")
	assert.Contains(t, quitter, "You sit down and begin your meditation.")
	assert.Contains(t, quitter, "Your meditation must complete without interruption to quit gracefully.")
	assert.Contains(t, strings.Join(reads[observerId], ""), "Ann</ansi> sits down and begins to meditate.")
}

func TestQuitInAFightSaysNothingOfMeditation(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	reading(t, player.Controller().Account())
	addReader(t, observerId, testRoomId)
	quitterId := player.Controller().Account().UserId

	reads := captureReads(t, func() {
		handled, err := Quit("", player, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})

	quitter := strings.Join(reads[quitterId], "")
	assert.Contains(t, quitter, "You're too busy to quit right now!")
	assert.NotContains(t, quitter, "meditat")
	assert.NotContains(t, strings.Join(reads[observerId], ""), "meditate")
}

// Phase 5 slice 4.1 (plan Q17): a player whose only fight is with a target
// in another room is not in a fight: she may quit.
func TestPlayerWhoseTargetIsInAnotherRoomMayQuit(t *testing.T) {
	npc, player, room := goSetup(t)
	player.Character.SetAggro(npc.Id, characters.DefaultAttack)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	reading(t, player.Controller().Account())
	quitterId := player.Controller().Account().UserId

	reads := captureReads(t, func() {
		handled, err := Quit("", player, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})

	quitter := strings.Join(reads[quitterId], "")
	assert.Contains(t, quitter, "You sit down and begin your meditation.", "her target is not here: she may quit")
	assert.NotContains(t, quitter, "You're too busy to quit right now!")
}
