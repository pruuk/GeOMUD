package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryAt makes roomId the death recovery room for one test and returns
// the ID of a private copy of it. The copy is recorded the way
// CreateEphemeralRoomIds records one (originalRoomIdLookups), at the last
// copy ID, which no test makes.
func recoveryAt(t *testing.T, roomId int) (copyId int) {
	t.Helper()
	old := int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["SpecialRooms.DeathRecoveryRoom"] = roomId
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["SpecialRooms.DeathRecoveryRoom"] = old
		require.NoError(t, configs.RestoreOverrides(back))
	})
	require.Equal(t, roomId, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom))

	copyId = ephemeralRoomIdMinimum + ephemeralChunksLimit*ephemeralChunkSize - 1
	originalRoomIdLookups[copyId] = roomId
	t.Cleanup(func() { delete(originalRoomIdLookups, copyId) })
	require.True(t, IsEphemeralRoomId(copyId))
	return copyId
}

func TestIsDeathRecoveryRoomKnowsTheRoomAndItsCopies(t *testing.T) {
	const recovery, other = 94101, 94102
	copyId := recoveryAt(t, recovery)
	assert.True(t, IsDeathRecoveryRoom(recovery), "the room")
	assert.True(t, IsDeathRecoveryRoom(copyId), "a private copy of it")
	assert.False(t, IsDeathRecoveryRoom(other), "another room")
}

// With no recovery room set (0, as in a test binary), no room is it: not
// room 0, and not an empty ephemeral room, whose original is recorded as 0
// (ephemeral.go:92).
func TestIsDeathRecoveryRoomIsNoRoomWhenNoneIsSet(t *testing.T) {
	emptyCopy := recoveryAt(t, 0) // a copy whose original reads 0
	assert.False(t, IsDeathRecoveryRoom(0), "room 0")
	assert.False(t, IsDeathRecoveryRoom(emptyCopy), "an empty ephemeral room")
	assert.False(t, IsDeathRecoveryRoom(94102), "another room")
}

// An attacker standing in its own copy of the recovery room may not fight
// (the raw ID compare missed the copy every respawn makes).
func TestCanPvpRefusesAnAttackerInACopyOfTheRecoveryRoom(t *testing.T) {
	const recovery, other = 94101, 94102
	copyId := recoveryAt(t, recovery)
	att, def := characters.New(), characters.New()
	att.RoomId, def.RoomId = copyId, other
	assert.EqualError(t, NewEmptyRoom().CanPvp(att, def), "Fighting is not allowed here.")
}
