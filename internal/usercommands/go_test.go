package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const inCombat = "You can't do that! You are in combat!\n"

// goSetup is setup with both bodies placed in the test room, a room to the
// north, and AP to spare for both.
func goSetup(t *testing.T) (npc, player *actors.Actor, room *rooms.Room) {
	t.Helper()
	npc, player, _ = setup(t)
	t.Cleanup(rooms.AddTestDefaultBiome()) // a move ends in a secret look
	joinNextRoom(t)
	room = rooms.LoadRoom(testRoomId)
	room.AddMob(npc.Id)
	room.AddPlayer(player.Controller().Account().UserId)
	npc.Character.Name = "Rat"
	npc.Character.Health = 10 // a bare test NPC is at 0, which is downed
	npc.Character.ActionPoints = 100
	player.Character.ActionPoints = 100
	return npc, player, room
}

// playerTexts is what the player was told while fn ran.
func playerTexts(t *testing.T, player *actors.Actor, fn func()) []string {
	t.Helper()
	out := []string{}
	for _, m := range capture[events.Message](t, fn) {
		if m.RoomId == 0 && m.UserId == player.Controller().Account().UserId {
			out = append(out, m.Text)
		}
	}
	return out
}

func TestNPCWithAggroTargetIsRefusedGo(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)

	handled, err := Go("north", npc, room, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, testRoomId, npc.Character.RoomId)
	assert.Equal(t, 100, npc.Character.ActionPoints, "a refused move costs nothing")
}

func TestPlayerTargetedByNPCInRoomIsRefusedGo(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	require.Nil(t, player.Character.Aggro, "the player has not attacked back")

	var handled bool
	told := playerTexts(t, player, func() {
		var err error
		handled, err = Go("north", player, room, 0)
		require.NoError(t, err)
	})
	assert.True(t, handled)
	assert.Equal(t, testRoomId, player.Character.RoomId)
	assert.Equal(t, []string{inCombat}, told)
}

func TestPlayerWhoseAttackerIsInAnotherRoomMayGo(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	// The attacker shoots from the next room.
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	require.NotContains(t, room.GetBodies(), npc.Id)

	handled, err := Go("north", player, room, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, testNextRoomId, player.Character.RoomId)
}

func TestNPCWithTooFewAPIsRefusedSilently(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.ActionPoints = 5

	var handled bool
	msgs := capture[events.Message](t, func() {
		var err error
		handled, err = Go("north", npc, room, 0)
		require.NoError(t, err)
	})
	assert.True(t, handled)
	assert.Equal(t, testRoomId, npc.Character.RoomId)
	assert.Equal(t, 5, npc.Character.ActionPoints)
	assert.Empty(t, msgs, "go prints nothing for an NPC it refuses")
	_ = player
}

// A move prepares the destination only for a player: rooms are prepared only
// for players (plan Decision 5).
func TestNPCMoveDoesNotPrepareTheDestination(t *testing.T) {
	npc, _, room := goSetup(t)
	dest := rooms.LoadRoom(testNextRoomId)
	dest.Containers = map[string]rooms.Container{"chest": {DespawnRound: 1}}

	handled, err := Go("north", npc, room, 0)
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, testNextRoomId, npc.Character.RoomId)
	assert.Contains(t, dest.Containers, "chest", "an NPC arriving prepares nothing")
}

func TestPlayerMovePreparesTheDestination(t *testing.T) {
	_, player, room := goSetup(t)
	dest := rooms.LoadRoom(testNextRoomId)
	dest.Containers = map[string]rooms.Container{"chest": {DespawnRound: 1}}

	handled, err := Go("north", player, room, 0)
	require.NoError(t, err)
	require.True(t, handled)
	assert.NotContains(t, dest.Containers, "chest", "a player arriving prepares the room")
}

func TestNPCWithAPMovesAndPaysTheCost(t *testing.T) {
	npc, _, room := goSetup(t)
	addReader(t, observerId, testRoomId)

	reads := captureReads(t, func() {
		handled, err := Go("north", npc, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})
	assert.Equal(t, testNextRoomId, npc.Character.RoomId)
	assert.Contains(t, rooms.LoadRoom(testNextRoomId).GetBodies(), npc.Id)
	assert.NotContains(t, room.GetBodies(), npc.Id)
	assert.Equal(t, 90, npc.Character.ActionPoints)

	// The player left behind in the room reads the leave line.
	left := reads[observerId]
	require.NotEmpty(t, left)
	assert.Contains(t, left[0], `<ansi fg="mobname">Rat</ansi> leaves towards`)
}
