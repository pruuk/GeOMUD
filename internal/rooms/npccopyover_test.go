package rooms

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// homeSlots is room 4102's template spawn list: two NPC slots whose MobId has
// no template, so a prepare never spawns into them.
func homeSlots() []SpawnInfo {
	return []SpawnInfo{
		{MobId: 987654, RespawnRate: "15 real minutes"},
		{MobId: 987655, RespawnRate: "15 real minutes"},
	}
}

// TestCopyoverBringsEveryNPCBackOnce is copyover's shape: every loaded room
// saves, the process goes (memory empties), the store is rebuilt from disk
// as main.go does, and rooms load and are prepared as players reach them.
func TestCopyoverBringsEveryNPCBackOnce(t *testing.T) {
	dir := persistSetup(t)

	x := liveTestNPC(t, 5, 1, 0) // live in A, owned by home slot 0
	y := liveTestNPC(t, 6, 1, 0) // live in B, no slot
	z := liveTestNPC(t, 7, 1, 0) // rested in A
	w := liveTestNPC(t, 8, 1, 0) // owned by home slot 1, rested in B
	mobs.GetInstance(x).SpawnSlot = 1
	mobs.GetInstance(w).SpawnSlot = 2

	a := memRoom(4100)
	b := memRoom(4101)
	home := memRoom(4102)
	home.SpawnInfo = homeSlots()
	home.SpawnInfo[0].InstanceId = x
	home.SpawnInfo[1].InstanceId = w
	a.bodies = append(a.bodies, x, z)
	b.bodies = append(b.bodies, y, w)
	require.True(t, a.RestNPC(mobs.GetInstance(z)))
	require.True(t, b.RestNPC(mobs.GetInstance(w)))

	require.NoError(t, SaveAllRooms())

	// The old process is gone.
	for _, id := range []ids.ID{x, y} {
		mobs.Rest(id)
	}
	restartStore(t)
	for _, id := range []ids.ID{x, y, z, w} {
		require.True(t, IsResting(id), "every NPC is resting after the rebuild")
	}

	load := func(roomId int, spawns []SpawnInfo) *Room {
		r := memRoom(roomId)
		r.SpawnInfo = spawns
		loadRoomNPCs(r)
		return r
	}
	a2 := load(4100, nil)
	home2 := load(4102, homeSlots())
	b2 := load(4101, nil)
	require.Equal(t, x, home2.SpawnInfo[0].InstanceId)
	require.Equal(t, w, home2.SpawnInfo[1].InstanceId)

	a2.Prepare(false)
	home2.Prepare(false)
	assert.Equal(t, x, home2.SpawnInfo[0].InstanceId, "its NPC is live in A")
	assert.Equal(t, w, home2.SpawnInfo[1].InstanceId, "its NPC rests in B, not yet prepared: the slot stays filled")
	assert.Zero(t, home2.SpawnInfo[1].DespawnedRound)
	b2.Prepare(false)

	inRooms := map[ids.ID]int{}
	for _, r := range []*Room{a2, b2, home2} {
		for _, id := range r.GetMobs() {
			inRooms[id]++
		}
	}
	for _, id := range []ids.ID{x, y, z, w} {
		assert.NotNil(t, mobs.GetInstance(id), "back under its saved ID")
		assert.Equal(t, 1, inRooms[id], "in exactly one room")
		assert.False(t, IsResting(id))
	}
	assert.ElementsMatch(t, []ids.ID{x, z}, a2.GetMobs())
	assert.ElementsMatch(t, []ids.ID{y, w}, b2.GetMobs())
	assert.ElementsMatch(t, []ids.ID{x, y, z, w}, allRestingOnDisk(t, dir), "on disk once each, as live NPCs")
	assert.Equal(t, x, home2.SpawnInfo[0].InstanceId)
	assert.Equal(t, w, home2.SpawnInfo[1].InstanceId)
}

// saveTemplateSetup gives SaveRoomTemplate the zone it looks up.
func saveTemplateSetup(t *testing.T) {
	t.Helper()
	persistSetup(t)
	roomManager.zones[persistZone] = &ZoneConfig{Name: persistZone, RoomId: 4100, RoomIds: map[int]struct{}{}}
	roomsDir := filepath.Join(configs.GetFilePathsConfig().DataFiles.String(), "rooms")
	require.NoError(t, os.MkdirAll(filepath.Join(roomsDir, filepath.FromSlash(ZoneToFolder(persistZone))), 0777))
	t.Cleanup(func() { os.RemoveAll(roomsDir) })
}

func TestSaveRoomTemplateWithTheRoomInMemoryKeepsItsNPCs(t *testing.T) {
	saveTemplateSetup(t)
	live := liveTestNPC(t, 5, 1, 0)
	rest := liveTestNPC(t, 6, 1, 0)
	round := util.GetRoundCount()

	r := memRoom(4101)
	r.SpawnInfo = homeSlots()
	r.SpawnInfo[0].InstanceId = live
	r.SpawnInfo[1].DespawnedRound = round // waiting after a death
	r.bodies = append(r.bodies, live, rest)
	require.True(t, r.RestNPC(mobs.GetInstance(rest)))

	tpl := Room{RoomId: 4101, Zone: persistZone, Title: "edited", SpawnInfo: homeSlots()}
	require.NoError(t, SaveRoomTemplate(tpl))

	got := getRoomFromMemory(4101)
	require.NotNil(t, got)
	assert.Equal(t, "edited", got.Title)
	assert.Equal(t, []ids.ID{live}, got.GetMobs())
	require.Len(t, got.resting, 1)
	assert.Equal(t, rest, got.resting[0].Template.Character.Id)
	assert.Equal(t, live, got.SpawnInfo[0].InstanceId)
	assert.Equal(t, round, got.SpawnInfo[1].DespawnedRound)

	f := readRoomNPCFile(persistZone, 4101)
	assert.ElementsMatch(t, []ids.ID{live, rest}, restingIds(f))
	assert.Equal(t, []slotState{{Index: 0, MobId: 987654, NPC: live}, {Index: 1, MobId: 987655, Despawned: round}}, f.Slots)
	assert.True(t, IsResting(rest))
}

func TestSaveRoomTemplateWithTheRoomNotInMemoryLoadsItsNPCs(t *testing.T) {
	saveTemplateSetup(t)
	owner := liveTestNPC(t, 5, 1, 0)
	r := memRoom(4101)
	r.SpawnInfo = homeSlots()
	r.SpawnInfo[0].InstanceId = owner
	r.bodies = append(r.bodies, owner)
	removeRoomFromMemory(r) // it rests on disk
	require.Nil(t, getRoomFromMemory(4101))

	tpl := Room{RoomId: 4101, Zone: persistZone, Title: "edited", SpawnInfo: homeSlots()}
	require.NoError(t, SaveRoomTemplate(tpl))

	got := getRoomFromMemory(4101)
	require.NotNil(t, got)
	require.Len(t, got.resting, 1)
	assert.Equal(t, owner, got.resting[0].Template.Character.Id)
	assert.Equal(t, owner, got.SpawnInfo[0].InstanceId)
	assert.True(t, IsResting(owner))

	got.Prepare(false)
	assert.NotNil(t, mobs.GetInstance(owner))
	assert.Equal(t, []ids.ID{owner}, got.GetMobs())
	assert.Equal(t, owner, got.SpawnInfo[0].InstanceId)
}
