package rooms

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// restIn makes a live slot NPC and rests it in roomId by unloading that room.
func restIn(t *testing.T, roomId int) ids.ID {
	t.Helper()
	id := liveTestNPC(t, 5, 1, 0)
	mobs.GetInstance(id).SpawnSlot = 1
	r := memRoom(roomId)
	r.bodies = append(r.bodies, id)
	removeRoomFromMemory(r)
	require.True(t, IsResting(id))
	return id
}

func sidecarExists(roomId int) bool {
	_, err := os.Stat(roomNPCPath(persistZone, roomId))
	return err == nil
}

func TestDeleteRoomEndsItsRestingNPCsAndKeepsLiveOnes(t *testing.T) {
	persistSetup(t)
	resting := restIn(t, 4100)

	// The room is back in memory, not yet prepared, with a live NPC too.
	r := memRoom(4100)
	loadRoomNPCs(r)
	live := liveTestNPC(t, 6, 1, 0)
	r.bodies = append(r.bodies, live)
	roomManager.roomSummaries[4100] = RoomSummaryInfo{Zone: persistZone}

	home := slotRoom(4101, resting)

	require.NoError(t, DeleteRoomForAdmin(4100))

	assert.Nil(t, getRoomFromMemory(4100))
	assert.False(t, sidecarExists(4100), "the deleted room's sidecar goes")
	assert.False(t, IsResting(resting), "its resting NPC leaves the index")
	assert.NotNil(t, mobs.GetInstance(live), "a live NPC stays live, as a delete always left it")

	home.Prepare(false)
	assert.True(t, home.SpawnInfo[0].InstanceId.IsZero(), "the slot that owned the resting NPC frees")
	assert.Equal(t, util.GetRoundCount(), home.SpawnInfo[0].DespawnedRound, "after its respawn rate, as a death")
}

func TestDeleteZoneEndsRestingNPCsOfLoadedAndUnloadedRooms(t *testing.T) {
	persistSetup(t)
	a := restIn(t, 4100)
	b := restIn(t, 4102)
	memRoom(4100) // loaded again; 4102 stays on disk only
	roomManager.zones[persistZone] = &ZoneConfig{Name: persistZone, RoomId: 4100,
		RoomIds: map[int]struct{}{4100: {}, 4102: {}}}

	require.NoError(t, DeleteZoneForAdmin(persistZone))

	assert.False(t, IsResting(a))
	assert.False(t, IsResting(b))
	assert.False(t, sidecarExists(4100))
	assert.False(t, sidecarExists(4102))
	for _, id := range []ids.ID{a, b} {
		e, ok := takeEnding(id)
		assert.True(t, ok)
		assert.False(t, e.Immediate)
	}
}

func TestMoveRoomNPCFileKeepsOneSidecarPerNPC(t *testing.T) {
	persistSetup(t)
	id := restIn(t, 4100)

	require.NoError(t, moveRoomNPCFile(persistZone, "Other Zone", 4100))
	assert.False(t, sidecarExists(4100))
	assert.Equal(t, []ids.ID{id}, restingIds(readRoomNPCFile("Other Zone", 4100)))

	// Writing it to another room still takes it out of the moved file.
	b := memRoom(4101)
	b.resting = []mobs.SavedNPC{readRoomNPCFile("Other Zone", 4100).Resting[0]}
	require.NoError(t, saveRoomNPCs(b))
	assert.Empty(t, readRoomNPCFile("Other Zone", 4100).Resting)
}

func TestRenameZoneNPCFilesKeepsOneSidecarPerNPC(t *testing.T) {
	dir := persistSetup(t)
	id := restIn(t, 4100)

	require.NoError(t, os.Rename(filepath.Join(dir, ZoneNameSanitize(persistZone)), filepath.Join(dir, ZoneNameSanitize("New Name"))))
	renameZoneNPCFiles(persistZone, "New Name")

	b := memRoom(4101)
	b.resting = []mobs.SavedNPC{readRoomNPCFile("New Name", 4100).Resting[0]}
	require.Equal(t, id, b.resting[0].Template.Character.Id)
	require.NoError(t, saveRoomNPCs(b))
	assert.Empty(t, readRoomNPCFile("New Name", 4100).Resting, "the renamed sidecar is still tracked")
}
