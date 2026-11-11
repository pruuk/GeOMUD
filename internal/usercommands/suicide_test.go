package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryRoom (the recovery room for one test, its copies freed after) is
// in death_test.go (Task 6).

// Nobody escapes a fight except by fleeing; dying by one's own hand is an
// escape.
func TestSuicideRefusedInAFight(t *testing.T) {
	npc, player, room := goSetup(t)
	recoveryRoom(t, testNextRoomId)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	health := player.Character.Health

	var deaths []events.PlayerDeath
	told := playerTexts(t, player, func() {
		deaths = capture[events.PlayerDeath](t, func() {
			_, err := Suicide("", player, room, 0)
			require.NoError(t, err)
		})
	})
	assert.Equal(t, []string{"You're too busy to die right now!\n"}, told)
	assert.Empty(t, deaths)
	assert.Equal(t, health, player.Character.Health)
	assert.Equal(t, testRoomId, player.Character.RoomId)
}

func TestSuicideRefusedInTheRecoveryRoom(t *testing.T) {
	_, player, room := goSetup(t)
	recoveryRoom(t, testRoomId)

	told := playerTexts(t, player, func() {
		_, err := Suicide("", player, room, 0)
		assert.EqualError(t, err, "already dead")
	})
	assert.Equal(t, []string{"You're already dead!\n"}, told)
}

// The private copy a death puts the body in is the recovery room too.
func TestSuicideRefusedInACopyOfTheRecoveryRoom(t *testing.T) {
	_, player, _ := goSetup(t)
	copyId := recoveryRoom(t, testNextRoomId)
	player.Character.RoomId = copyId

	told := playerTexts(t, player, func() {
		_, err := Suicide("", player, rooms.LoadRoom(copyId), 0)
		assert.EqualError(t, err, "already dead")
	})
	assert.Equal(t, []string{"You're already dead!\n"}, told)
}

func TestSuicideDiesOtherwise(t *testing.T) {
	_, player, room := goSetup(t)
	recoveryRoom(t, testNextRoomId)

	deaths := capture[events.PlayerDeath](t, func() {
		_, err := Suicide("", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, deaths, 1)
	assert.Equal(t, 1, player.Character.Health)
	assert.True(t, rooms.IsEphemeralRoomId(player.Character.RoomId))
	assert.Equal(t, testNextRoomId, rooms.GetOriginalRoom(player.Character.RoomId), "in a copy of the recovery room")
}
