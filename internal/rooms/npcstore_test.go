package rooms

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// useTempNPCStore points the store at a temp dir and clears its state.
func useTempNPCStore(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := npcStoreRoot
	npcStoreRoot = func() string { return dir }
	npcStoreMu.Lock()
	restingIndex = map[ids.ID]int{}
	endingsLedger = map[ids.ID]ending{}
	npcSidecars = map[ids.ID][]string{}
	sidecarNPCs = map[string][]ids.ID{}
	npcStoreMu.Unlock()
	t.Cleanup(func() { npcStoreRoot = old })
	return dir
}

// restartStore is what a restart or copyover leaves: the store's memory is
// gone and rebuilt from disk the way main.go rebuilds it.
func restartStore(t *testing.T) {
	t.Helper()
	npcStoreMu.Lock()
	restingIndex = map[ids.ID]int{}
	endingsLedger = map[ids.ID]ending{}
	npcSidecars = map[ids.ID][]string{}
	sidecarNPCs = map[string][]ids.ID{}
	npcStoreMu.Unlock()
	resetRoomManager()
	require.NoError(t, BuildRestingIndex())
	require.NoError(t, LoadEndings())
}

func savedWithID(id ids.ID, mobId int) mobs.SavedNPC {
	s := mobs.SavedNPC{HomeRoomId: 1}
	s.Template.MobId = mobs.MobId(mobId)
	s.Template.Character = characters.Character{Id: id}
	return s
}

func TestNPCStore_BuildRestingIndex(t *testing.T) {
	dir := useTempNPCStore(t)
	a, b, c := ids.New(0), ids.New(0), ids.New(0)

	require.NoError(t, writeRoomNPCFile("Frostfang", 100, roomNPCFile{
		Resting: []mobs.SavedNPC{savedWithID(a, 1), savedWithID(b, 2)},
		Slots:   []slotState{{Index: 0, MobId: 1, NPC: a}},
	}))
	require.NoError(t, writeRoomNPCFile("Other Zone", 200, roomNPCFile{
		Resting: []mobs.SavedNPC{savedWithID(c, 3)},
	}))
	// a file that does not parse is skipped
	bad := filepath.Join(dir, "frostfang", "999.npcs.yaml")
	require.NoError(t, os.WriteFile(bad, []byte("resting: [: nope"), 0644))

	require.NoError(t, BuildRestingIndex())

	assert.True(t, IsResting(a))
	assert.True(t, IsResting(b))
	assert.True(t, IsResting(c))
	assert.False(t, IsResting(ids.New(0)))
	assert.Equal(t, 100, restingIndex[a])
	assert.Equal(t, 100, restingIndex[b])
	assert.Equal(t, 200, restingIndex[c])

	got := readRoomNPCFile("Frostfang", 100)
	require.Len(t, got.Resting, 2)
	assert.Equal(t, a, got.Resting[0].Template.Character.Id)
	require.Len(t, got.Slots, 1)
	assert.Equal(t, slotState{Index: 0, MobId: 1, NPC: a}, got.Slots[0])
}

func TestNPCStore_EmptyFileDeletes(t *testing.T) {
	useTempNPCStore(t)
	require.NoError(t, writeRoomNPCFile("Z", 5, roomNPCFile{Slots: []slotState{{Index: 0, MobId: 1, NPC: ids.New(0)}}}))
	p := roomNPCPath("Z", 5)
	_, err := os.Stat(p)
	require.NoError(t, err)

	require.NoError(t, writeRoomNPCFile("Z", 5, roomNPCFile{}))
	_, err = os.Stat(p)
	assert.True(t, os.IsNotExist(err))

	// empty write with no file is fine
	require.NoError(t, writeRoomNPCFile("Z", 5, roomNPCFile{}))
	assert.Empty(t, readRoomNPCFile("Z", 5).Resting)
}

func TestNPCStore_EndingsRoundTripAndTake(t *testing.T) {
	useTempNPCStore(t)
	a, b := ids.New(0), ids.New(0)
	RecordEnding(a, 10, false)
	RecordEnding(b, 20, true)

	npcStoreMu.Lock()
	endingsLedger = map[ids.ID]ending{}
	npcStoreMu.Unlock()
	require.NoError(t, LoadEndings())

	e, ok := takeEnding(a)
	assert.True(t, ok)
	assert.Equal(t, ending{Round: 10}, e)
	_, ok = takeEnding(a)
	assert.False(t, ok)

	// the take was saved: a fresh load has only b
	npcStoreMu.Lock()
	endingsLedger = map[ids.ID]ending{}
	npcStoreMu.Unlock()
	require.NoError(t, LoadEndings())
	_, ok = takeEnding(a)
	assert.False(t, ok)
	e, ok = takeEnding(b)
	assert.True(t, ok)
	assert.Equal(t, ending{Round: 20, Immediate: true}, e)
}

func TestNPCStore_LoadEndingsMissingFile(t *testing.T) {
	useTempNPCStore(t)
	assert.NoError(t, LoadEndings())
}

func TestNPCStore_IDYAMLRoundTrips(t *testing.T) {
	a, b, c := ids.New(0), ids.New(7), ids.New(0)

	// inside slotState
	out, err := yaml.Marshal(slotState{Index: 2, MobId: 3, NPC: a})
	require.NoError(t, err)
	var ss slotState
	require.NoError(t, yaml.Unmarshal(out, &ss))
	assert.Equal(t, slotState{Index: 2, MobId: 3, NPC: a}, ss)

	// inside SavedNPC.Aggro
	s := savedWithID(b, 4)
	s.Aggro = &characters.Aggro{Target: c}
	out, err = yaml.Marshal(s)
	require.NoError(t, err)
	var s2 mobs.SavedNPC
	require.NoError(t, yaml.Unmarshal(out, &s2))
	require.NotNil(t, s2.Aggro)
	assert.Equal(t, c, s2.Aggro.Target)
	assert.Equal(t, b, s2.Template.Character.Id)

	// as a map key
	m := map[ids.ID]ending{a: {Round: 1}, b: {Round: 2, Immediate: true}}
	out, err = yaml.Marshal(m)
	require.NoError(t, err)
	m2 := map[ids.ID]ending{}
	require.NoError(t, yaml.Unmarshal(out, &m2))
	assert.Equal(t, m, m2)
}
