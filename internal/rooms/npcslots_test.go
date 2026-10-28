package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// slotRoom is a room in memory with one spawn slot owning npc. Its MobId has
// no template, so a respawn attempt spawns nothing.
func slotRoom(roomId int, npc ids.ID) *Room {
	r := memRoom(roomId)
	r.SpawnInfo = []SpawnInfo{{MobId: 987654, InstanceId: npc, RespawnRate: "15 real minutes"}}
	return r
}

func TestSlotWhoseNPCRestsElsewhereStaysFilled(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 1, 0)

	home := slotRoom(4100, id)
	away := memRoom(4101)
	away.bodies = append(away.bodies, id)
	removeRoomFromMemory(away)
	require.True(t, IsResting(id))

	home.Prepare(false)
	assert.Equal(t, id, home.SpawnInfo[0].InstanceId, "the slot's NPC rests elsewhere; the slot stays filled")
	assert.Zero(t, home.SpawnInfo[0].DespawnedRound)

	home.CleanupMobSpawns()
	assert.Equal(t, id, home.SpawnInfo[0].InstanceId, "CleanupMobSpawns keeps it too")

	home.RoundTick()
	assert.Equal(t, id, home.SpawnInfo[0].InstanceId, "and RoundTick")
}

func TestSlotWhoseNPCIsLiveElsewhereStaysFilled(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 1, 0)
	home := slotRoom(4100, id)

	home.Prepare(false)
	home.CleanupMobSpawns()
	home.RoundTick()
	assert.Equal(t, id, home.SpawnInfo[0].InstanceId)
}

func TestSlotWhoseNPCDiedWaitsTheRespawnRate(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell)
	RecordEnding(id, 777, false)
	home := slotRoom(4100, id)

	home.Prepare(false)
	assert.True(t, home.SpawnInfo[0].InstanceId.IsZero(), "an ended NPC frees its slot")
	assert.Equal(t, uint64(777), home.SpawnInfo[0].DespawnedRound, "a death waits the respawn rate from its round")
	_, ok := takeEnding(id)
	assert.False(t, ok, "the slot consumed the ending")
}

func TestSlotWhoseNPCWasDespawnedRespawnsAtOnce(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell)
	RecordEnding(id, 777, true)
	home := slotRoom(4100, id)

	home.CleanupMobSpawns()
	assert.True(t, home.SpawnInfo[0].InstanceId.IsZero())
	assert.Zero(t, home.SpawnInfo[0].DespawnedRound, "an admin's despawn has no cooldown")
}

func TestSlotWhoseNPCIsLostLogsAndRespawnsAfterTheRate(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell) // not live, not resting, not ended
	home := slotRoom(4100, id)

	home.RoundTick()
	assert.True(t, home.SpawnInfo[0].InstanceId.IsZero())
	assert.Equal(t, util.GetRoundCount(), home.SpawnInfo[0].DespawnedRound)
}

func TestSlotThatEndedIsSavedSoARestartDoesNotCallItLost(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell)
	round := util.GetRoundCount()
	RecordEnding(id, round, false)
	home := slotRoom(4100, id)
	require.NoError(t, saveRoomNPCs(home))
	require.Len(t, readRoomNPCFile(persistZone, 4100).Slots, 1)

	home.Prepare(false)
	assert.Equal(t, []slotState{{Index: 0, MobId: 987654, Despawned: round}}, readRoomNPCFile(persistZone, 4100).Slots,
		"the ending was consumed and saved: the slot is saved free, with its wait")
}

func TestDeathWaitSurvivesARestart(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell)
	round := util.GetRoundCount()
	RecordEnding(id, round, false) // a death this round
	home := slotRoom(4100, id)
	home.Prepare(false)
	require.Equal(t, round, home.SpawnInfo[0].DespawnedRound)

	restartStore(t)

	back := memRoom(4100)
	back.SpawnInfo = []SpawnInfo{{MobId: 987654, RespawnRate: "15 real minutes"}}
	loadRoomNPCs(back)
	assert.True(t, back.SpawnInfo[0].InstanceId.IsZero())
	assert.Equal(t, round, back.SpawnInfo[0].DespawnedRound, "the slot still waits its respawn rate")
}
