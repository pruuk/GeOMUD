package rooms_test

// behaviour.Rest needs the room package's NPC store, whose test seams are
// unexported, and rooms cannot import behaviour (behaviour imports rooms). So
// Rest is tested here, from the external test package, through export_test.go.

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/behaviour"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restTestNPC(t *testing.T, roomId int) *mobs.Mob {
	t.Helper()
	m := mobs.Mob{MobId: 5, HomeRoomId: 4100}
	m.Character = *characters.New()
	m.Character.Name = "rat"
	m.Character.RoomId = roomId
	inst := mobs.NewTestInstance(m)
	id := inst.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	return inst
}

func TestRestTakesAnNPCOutOfMemoryIntoItsRoomsRestingList(t *testing.T) {
	rooms.PersistSetup(t)
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = 4100, "Rest Zone"
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(4100) })
	mob := restTestNPC(t, 4100)
	id := mob.Character.Id
	r.AddTestBody(id)

	behaviour.Rest(mob, "depression 3/3")

	assert.Nil(t, mobs.GetInstance(id), "the NPC is out of memory")
	assert.True(t, rooms.IsResting(id), "and resting")
	assert.Empty(t, r.GetMobs(), "and not in the room")
}

func TestRestInAnEphemeralRoomEndsTheNPC(t *testing.T) {
	rooms.PersistSetup(t)
	roomId := rooms.EphemeralRoomIdMinimum() + 9
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = roomId, "Rest Zone"
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(roomId) })
	require.True(t, r.IsEphemeral())
	mob := restTestNPC(t, roomId)
	id := mob.Character.Id
	r.AddTestBody(id)

	behaviour.Rest(mob, "charm expired")

	assert.Nil(t, mobs.GetInstance(id))
	assert.False(t, rooms.IsResting(id), "nothing is saved in an ephemeral room, so it cannot rest")
	assert.Empty(t, r.GetMobs())
}
