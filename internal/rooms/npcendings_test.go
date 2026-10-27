package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestNPCRestsOneNPCAndPrepareRestoresIt(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 1, 0)
	other := liveTestNPC(t, 6, 1, 0)
	r := memRoom(4100)
	r.bodies = append(r.bodies, id, other)

	require.True(t, r.RestNPC(mobs.GetInstance(id)))

	assert.Nil(t, mobs.GetInstance(id))
	assert.Nil(t, actors.Get(id))
	assert.True(t, IsResting(id))
	assert.Equal(t, []ids.ID{other}, r.GetMobs(), "only that NPC left the room")
	assert.ElementsMatch(t, []ids.ID{id, other}, restingIds(readRoomNPCFile(persistZone, 4100)),
		"the sidecar has it as resting and the other as live, once each")

	r.Prepare(false)
	assert.NotNil(t, mobs.GetInstance(id))
	assert.False(t, IsResting(id))
	assert.ElementsMatch(t, []ids.ID{id, other}, r.GetMobs())
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
