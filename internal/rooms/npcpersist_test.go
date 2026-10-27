package rooms

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const persistZone = "Persist Zone"

// liveTestNPC adds a live NPC instance with some state worth keeping: hurt by
// damage below its maximum health, and carrying gold.
func liveTestNPC(t *testing.T, mobId int, damage, gold int) ids.ID {
	t.Helper()
	m := mobs.Mob{MobId: mobs.MobId(mobId), HomeRoomId: 4100}
	m.Character = *characters.New()
	m.Character.HealthMax.Base = 0 // as a mob loaded from YAML has (HealthMax is not saved)
	m.Character.Name = "rat"
	m.Character.Level = 20
	m.Character.Gold = gold
	inst := mobs.NewTestInstance(m)
	inst.Character.Validate()
	require.Greater(t, inst.Character.HealthMax.Value, damage)
	inst.Character.Health = inst.Character.HealthMax.Value - damage
	id := inst.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	return id
}

// memRoom puts an empty room in memory, as LoadRoom would.
func memRoom(roomId int) *Room {
	r := NewEmptyRoom()
	r.RoomId = roomId
	r.Zone = persistZone
	roomManager.rooms[roomId] = r
	return r
}

func restingIds(f roomNPCFile) []ids.ID {
	out := []ids.ID{}
	for _, s := range f.Resting {
		out = append(out, s.Template.Character.Id)
	}
	return out
}

func persistSetup(t *testing.T) string {
	t.Helper()
	dir := useTempNPCStore(t)
	resetRoomManager()
	actors.Reset()
	t.Cleanup(actors.Reset)
	return dir
}

func TestUnloadRestsNPCsAndPrepareRestoresThem(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 2, 13)
	health := mobs.GetInstance(id).Character.Health

	r := memRoom(4100)
	r.bodies = append(r.bodies, id)

	removeRoomFromMemory(r)

	assert.Nil(t, mobs.GetInstance(id), "an unloaded room's NPC is out of memory")
	assert.Nil(t, actors.Get(id), "and its actor is gone")
	assert.True(t, IsResting(id), "and in the resting index")
	assert.Equal(t, []ids.ID{id}, restingIds(readRoomNPCFile(persistZone, 4100)), "and in the sidecar once")
	assert.Nil(t, getRoomFromMemory(4100))

	// The room comes back from disk and is prepared for a player.
	r2 := NewEmptyRoom()
	r2.RoomId, r2.Zone = 4100, persistZone
	loadRoomNPCs(r2)
	require.Len(t, r2.resting, 1)
	addRoomToMemory(r2)
	r2.Prepare(false)

	m := mobs.GetInstance(id)
	require.NotNil(t, m, "prepare restores the NPC under its saved ID")
	assert.NotNil(t, actors.Get(id), "with its actor")
	assert.Equal(t, health, m.Character.Health)
	assert.Equal(t, 13, m.Character.Gold)
	assert.Equal(t, 4100, m.Character.RoomId)
	assert.Equal(t, []ids.ID{id}, r2.GetMobs())
	assert.False(t, IsResting(id), "a restored NPC is not resting")
	assert.Empty(t, r2.resting)
	assert.Equal(t, []ids.ID{id}, restingIds(readRoomNPCFile(persistZone, 4100)),
		"the sidecar holds it once, as the room's live NPC")

	r2.Prepare(false)
	assert.Equal(t, []ids.ID{id}, r2.GetMobs(), "a second prepare does not restore it twice")
}

func TestSaveWritesLiveNPCsOnceAndSlots(t *testing.T) {
	persistSetup(t)
	a := liveTestNPC(t, 5, 3, 0)
	b := liveTestNPC(t, 6, 3, 0)

	r := memRoom(4100)
	r.bodies = append(r.bodies, a)
	r.resting = []mobs.SavedNPC{mobs.Snapshot(mobs.GetInstance(a))} // as unload builds it
	r.SpawnInfo = []SpawnInfo{{MobId: 5, InstanceId: a}, {MobId: 6}, {MobId: 6, InstanceId: b}}

	require.NoError(t, saveRoomNPCs(r))

	f := readRoomNPCFile(persistZone, 4100)
	assert.Equal(t, []ids.ID{a}, restingIds(f), "an NPC both resting and live is written once")
	assert.Equal(t, []slotState{{Index: 0, MobId: 5, NPC: a}, {Index: 2, MobId: 6, NPC: b}}, f.Slots)

	// load merges slot state by index and MobId
	r2 := NewEmptyRoom()
	r2.RoomId, r2.Zone = 4100, persistZone
	r2.SpawnInfo = []SpawnInfo{{MobId: 5}, {MobId: 6}, {MobId: 99}}
	loadRoomNPCs(r2)
	assert.Equal(t, a, r2.SpawnInfo[0].InstanceId)
	assert.True(t, r2.SpawnInfo[1].InstanceId.IsZero())
	assert.True(t, r2.SpawnInfo[2].InstanceId.IsZero(), "a slot whose MobId changed starts fresh")
}

func TestNPCIsInOneSidecarAtATime(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 3, 0)

	a := memRoom(4100)
	b := memRoom(4101)
	a.bodies = append(a.bodies, id)
	require.NoError(t, saveRoomNPCs(a)) // an autosave while it is in A
	require.Equal(t, []ids.ID{id}, restingIds(readRoomNPCFile(persistZone, 4100)))

	// It wanders to B and B unloads before A saves again.
	a.removeBody(id)
	b.bodies = append(b.bodies, id)
	removeRoomFromMemory(b)

	assert.Equal(t, []ids.ID{id}, restingIds(readRoomNPCFile(persistZone, 4101)))
	assert.Empty(t, restingIds(readRoomNPCFile(persistZone, 4100)), "writing it to B took it out of A")

	// A restart rebuilds the index from disk: one room has it.
	require.NoError(t, BuildRestingIndex())
	assert.Equal(t, 4101, restingIndex[id])
}

func TestEphemeralUnloadDestroysAndWritesNothing(t *testing.T) {
	dir := persistSetup(t)
	id := liveTestNPC(t, 5, 3, 0)

	r := memRoom(ephemeralRoomIdMinimum + 5)
	r.bodies = append(r.bodies, id)
	removeRoomFromMemory(r)

	assert.Nil(t, mobs.GetInstance(id))
	assert.False(t, IsResting(id), "ephemeral NPCs do not rest")
	entries, err := os.ReadDir(filepath.Join(dir, ZoneNameSanitize(persistZone)))
	if err == nil {
		assert.Empty(t, entries)
	}
}

func TestTemplateEditCarriesRestingAndSlots(t *testing.T) {
	persistSetup(t)
	id := ids.New(ids.NoCell)
	old := memRoom(4100)
	old.resting = []mobs.SavedNPC{savedWithID(id, 5)}
	old.SpawnInfo = []SpawnInfo{{MobId: 5, InstanceId: id}, {MobId: 6, InstanceId: ids.New(ids.NoCell)}, {MobId: 8, DespawnedRound: 42}}

	tpl := Room{RoomId: 4100, Zone: persistZone, SpawnInfo: []SpawnInfo{{MobId: 5}, {MobId: 7}, {MobId: 8}}}
	carryNPCState(old, &tpl)

	require.Len(t, tpl.resting, 1)
	assert.Equal(t, id, tpl.resting[0].Template.Character.Id)
	assert.Equal(t, id, tpl.SpawnInfo[0].InstanceId)
	assert.True(t, tpl.SpawnInfo[1].InstanceId.IsZero(), "a changed slot starts fresh")
	assert.Equal(t, uint64(42), tpl.SpawnInfo[2].DespawnedRound, "a slot waiting after a death keeps its wait")
}
