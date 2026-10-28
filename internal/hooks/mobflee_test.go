package hooks

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fleeTestRoomA = 93501
	fleeTestRoomB = 93502
)

// fleeTestNPC puts an NPC called name in room roomId.
func fleeTestNPC(t *testing.T, mobId int, name string, roomId int) *mobs.Mob {
	t.Helper()
	m := mobs.Mob{MobId: mobs.MobId(mobId), HomeRoomId: roomId}
	m.Character = *characters.New()
	m.Character.Name = name
	m.Character.RoomId = roomId
	inst := mobs.NewTestInstance(m)
	rooms.LoadRoom(roomId).AddMob(inst.Character.Id)
	t.Cleanup(func() { mobs.Rest(inst.Character.Id) })
	return inst
}

// fleeTestRooms adds room A and, when linked, room B north of it.
func fleeTestRooms(t *testing.T, linked bool) {
	t.Helper()
	actors.Reset()
	t.Cleanup(actors.Reset)
	t.Cleanup(rooms.AddTestDefaultBiome()) // messaging asks the room's light
	a := rooms.NewEmptyRoom()
	a.RoomId, a.Zone = fleeTestRoomA, "fleezone"
	rooms.AddTestRoom(a)
	t.Cleanup(func() { rooms.RemoveTestRoom(fleeTestRoomA) })
	if linked {
		b := rooms.NewEmptyRoom()
		b.RoomId, b.Zone = fleeTestRoomB, "fleezone"
		rooms.AddTestRoom(b)
		t.Cleanup(func() { rooms.RemoveTestRoom(fleeTestRoomB) })
		a.Exits = map[string]exit.RoomExit{"north": {RoomId: fleeTestRoomB}}
	}
}

func recordAggroChanged(t *testing.T) *[]events.AggroChanged {
	t.Helper()
	changed := []events.AggroChanged{}
	id := events.RegisterListener(events.AggroChanged{}, func(e events.Event) events.ListenerReturn {
		changed = append(changed, e.(events.AggroChanged))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.AggroChanged{}, id) })
	events.ProcessEvents()
	changed = changed[:0]
	return &changed
}

// An NPC that flees with nobody blocking leaves its room by an exit, as a
// player does (plan Decision 13), and stops fighting.
func TestFledNPCLeavesItsRoom(t *testing.T) {
	fleeTestRooms(t, true)
	mob := fleeTestNPC(t, 5, "rat", fleeTestRoomA)
	other := fleeTestNPC(t, 6, "cat", fleeTestRoomA)
	mob.Character.SetAggro(other.Character.Id, characters.DefaultAttack)

	a := actors.Get(mob.Character.Id)
	require.NotNil(t, a)
	_, err := usercommands.Flee(``, a, rooms.LoadRoom(fleeTestRoomA), events.CmdNone)
	require.NoError(t, err)
	require.Equal(t, other.Character.Id, mob.Character.Aggro.FleeFrom, "flee keeps the target fled from")

	changed := recordAggroChanged(t)

	DoCombat(events.NewRound{})
	events.ProcessEvents()

	assert.Equal(t, fleeTestRoomB, mob.Character.RoomId, "a fled NPC leaves its room")
	assert.Nil(t, mob.Character.Aggro, "a fled NPC stops fighting")
	require.Len(t, *changed, 1)
	assert.Equal(t, mob.Character.Id, (*changed)[0].ActorId)
}

// A flee with no exit puts the fleer back in the fight it fled (plan
// Decision 2), instead of leaving it fighting nobody.
func TestExitlessFleeResumesTheFight(t *testing.T) {
	fleeTestRooms(t, false)
	mob := fleeTestNPC(t, 5, "rat", fleeTestRoomA)
	other := fleeTestNPC(t, 6, "cat", fleeTestRoomA)
	mob.Character.SetAggro(other.Character.Id, characters.DefaultAttack)

	_, err := usercommands.Flee(``, actors.Get(mob.Character.Id), rooms.LoadRoom(fleeTestRoomA), events.CmdNone)
	require.NoError(t, err)

	changed := recordAggroChanged(t)

	DoCombat(events.NewRound{})
	events.ProcessEvents()

	assert.Equal(t, fleeTestRoomA, mob.Character.RoomId)
	require.NotNil(t, mob.Character.Aggro)
	assert.Equal(t, other.Character.Id, mob.Character.Aggro.Target, "the fleer resumes its pre-flee target")
	assert.Equal(t, characters.DefaultAttack, mob.Character.Aggro.Type)
	require.Len(t, *changed, 1)
	assert.Equal(t, mob.Character.Id, (*changed)[0].ActorId)
}
