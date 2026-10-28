package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadedRoom brings room roomId into memory from disk, as LoadRoom does.
func loadedRoom(roomId int) *Room {
	r := memRoom(roomId)
	loadRoomNPCs(r)
	return r
}

func TestRestNPCInALoadedRoomStaysRestingWhileTheRoomStays(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 1, 0)
	other := liveTestNPC(t, 6, 1, 0)
	r := loadedRoom(4100)
	r.bodies = append(r.bodies, id, other)
	r.Prepare(false) // a player arrived: the room's first prepare after its load

	require.True(t, r.RestNPC(mobs.GetInstance(id)))

	assert.Nil(t, mobs.GetInstance(id))
	assert.Nil(t, actors.Get(id))
	assert.True(t, IsResting(id))
	assert.Equal(t, []ids.ID{other}, r.GetMobs(), "only that NPC left the room")
	assert.ElementsMatch(t, []ids.ID{id, other}, restingIds(readRoomNPCFile(persistZone, 4100)),
		"the sidecar has it as resting and the other as live, once each")

	// Rounds pass: prepares (as HandleRespawns and arrivals run them) and saves.
	for i := 0; i < 5; i++ {
		r.Prepare(false)
		r.CleanupMobSpawns()
		require.NoError(t, saveRoomNPCs(r))
	}
	assert.Nil(t, mobs.GetInstance(id), "it does not come back into a room that stayed in memory")
	assert.True(t, IsResting(id))
	assert.Equal(t, []ids.ID{other}, r.GetMobs())
	assert.ElementsMatch(t, []ids.ID{id, other}, restingIds(readRoomNPCFile(persistZone, 4100)),
		"and later saves keep it resting")
}

func TestRestNPCComesBackOnceAfterItsRoomUnloadsAndLoads(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 1, 0)
	other := liveTestNPC(t, 6, 1, 0)
	r := loadedRoom(4100)
	r.bodies = append(r.bodies, id, other)
	r.Prepare(false)
	require.True(t, r.RestNPC(mobs.GetInstance(id)))

	removeRoomFromMemory(r)
	assert.Nil(t, mobs.GetInstance(id))
	assert.Nil(t, mobs.GetInstance(other))
	assert.ElementsMatch(t, []ids.ID{id, other}, restingIds(readRoomNPCFile(persistZone, 4100)),
		"the unload keeps it resting, once")

	r2 := loadedRoom(4100)
	r2.Prepare(false)
	assert.NotNil(t, mobs.GetInstance(id), "the first prepare after the load restores it")
	assert.NotNil(t, mobs.GetInstance(other))
	assert.False(t, IsResting(id))
	assert.ElementsMatch(t, []ids.ID{id, other}, r2.GetMobs())

	r2.Prepare(false)
	assert.ElementsMatch(t, []ids.ID{id, other}, r2.GetMobs(), "once")
}

func TestRestNPCIntoALoadedButUnpreparedRoomWaitsForTheNextLoad(t *testing.T) {
	persistSetup(t)
	before := liveTestNPC(t, 5, 1, 0)
	r := memRoom(4100)
	r.bodies = append(r.bodies, before)
	removeRoomFromMemory(r) // before rests on disk

	r2 := loadedRoom(4100)
	late := liveTestNPC(t, 6, 1, 0)
	r2.bodies = append(r2.bodies, late)
	require.True(t, r2.RestNPC(mobs.GetInstance(late))) // rested before the room's first prepare

	r2.Prepare(false)
	assert.NotNil(t, mobs.GetInstance(before), "what the load read comes back")
	assert.Nil(t, mobs.GetInstance(late), "what rested after the load does not")
	assert.True(t, IsResting(late))
	assert.ElementsMatch(t, []ids.ID{before, late}, restingIds(readRoomNPCFile(persistZone, 4100)))
}

func TestNPCEndedRecordsOnlySlotNPCsAndLeavesItsSidecar(t *testing.T) {
	persistSetup(t)
	slotted := liveTestNPC(t, 5, 1, 0)
	slotless := liveTestNPC(t, 6, 1, 0)
	mobs.GetInstance(slotted).SpawnSlot = 1
	r := memRoom(4100)
	r.bodies = append(r.bodies, slotted, slotless)
	require.NoError(t, saveRoomNPCs(r)) // an autosave wrote both

	NPCEnded(mobs.GetInstance(slotted), false)
	NPCEnded(mobs.GetInstance(slotless), true)

	assert.Empty(t, restingIds(readRoomNPCFile(persistZone, 4100)), "an ended NPC leaves its sidecar")
	e, ok := takeEnding(slotted)
	assert.True(t, ok)
	assert.False(t, e.Immediate)
	_, ok = takeEnding(slotless)
	assert.False(t, ok, "an NPC with no slot ends without a ledger entry")
}

func TestDeleteRoomEndsOnlySlotNPCsInTheLedger(t *testing.T) {
	persistSetup(t)
	slotted := liveTestNPC(t, 5, 1, 0)
	slotless := liveTestNPC(t, 6, 1, 0)
	mobs.GetInstance(slotted).SpawnSlot = 1
	r := memRoom(4100)
	r.bodies = append(r.bodies, slotted, slotless)
	removeRoomFromMemory(r)
	roomManager.roomSummaries[4100] = RoomSummaryInfo{Zone: persistZone}

	require.NoError(t, DeleteRoomForAdmin(4100))
	assert.False(t, IsResting(slotted))
	assert.False(t, IsResting(slotless))
	_, ok := takeEnding(slotted)
	assert.True(t, ok)
	_, ok = takeEnding(slotless)
	assert.False(t, ok)
}
