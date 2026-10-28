package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each test lets only the first disk write of a pair through, then rebuilds
// the store from disk: a crash between the two writes must never bring an
// ended NPC back, nor leave a freed slot owning an NPC that is nowhere.

func TestCrashInsideNPCEndedNeverBringsItBack(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 1, 0)
	mobs.GetInstance(id).SpawnSlot = 1
	r := memRoom(4100)
	r.bodies = append(r.bodies, id)
	require.NoError(t, saveRoomNPCs(r)) // an autosave wrote it

	crashAfterWrites(t, 1)
	NPCEnded(mobs.GetInstance(id), false)
	mobs.DestroyInstance(id)
	diskWritesBack()

	restartStore(t)
	assert.False(t, IsResting(id), "the ended NPC is not on disk to come back")
}

func TestCrashInsideSlotSettleKeepsTheFreedSlotAndItsWait(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell)
	round := util.GetRoundCount()
	RecordEnding(id, round, false)
	home := slotRoom(4100, id)
	require.NoError(t, saveRoomNPCs(home)) // the slot owns it on disk

	crashAfterWrites(t, 1)
	home.Prepare(false)
	diskWritesBack()

	restartStore(t)
	back := memRoom(4100)
	back.SpawnInfo = []SpawnInfo{{MobId: 987654, RespawnRate: "15 real minutes"}}
	loadRoomNPCs(back)
	assert.True(t, back.SpawnInfo[0].InstanceId.IsZero(), "the slot is saved free, not owning an NPC that is nowhere")
	assert.Equal(t, round, back.SpawnInfo[0].DespawnedRound, "with its wait")
}

func TestCrashInsideRoomDeleteNeverBringsItsNPCsBack(t *testing.T) {
	persistSetup(t)
	id := restIn(t, 4100)
	memRoom(4100)
	loadRoomNPCs(roomManager.rooms[4100])
	roomManager.roomSummaries[4100] = RoomSummaryInfo{Zone: persistZone}

	crashAfterWrites(t, 1)
	require.NoError(t, DeleteRoomForAdmin(4100))
	diskWritesBack()

	restartStore(t)
	assert.False(t, IsResting(id), "the deleted room's NPC is not on disk to come back")
}
