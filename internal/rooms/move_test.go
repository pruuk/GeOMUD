package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	moveRoomA = 91001
	moveRoomB = 91002
)

// moveSetup puts two rooms in memory. Room B holds a crumbling container so a
// Prepare of that room is visible.
func moveSetup(t *testing.T) (a, b *Room, changes *[]events.RoomChange) {
	t.Helper()
	actors.Reset()
	events.ClearListeners()
	a = NewEmptyRoom()
	a.RoomId, a.Zone = moveRoomA, "movezone"
	b = NewEmptyRoom()
	b.RoomId, b.Zone = moveRoomB, "movezone"
	b.Containers = map[string]Container{"chest": {DespawnRound: 1}}
	AddTestRoom(a)
	AddTestRoom(b)
	t.Cleanup(func() {
		RemoveTestRoom(moveRoomA)
		RemoveTestRoom(moveRoomB)
		actors.Reset()
		events.ClearListeners()
	})

	got := []events.RoomChange{}
	events.RegisterListener(events.RoomChange{}, func(e events.Event) events.ListenerReturn {
		got = append(got, e.(events.RoomChange))
		return events.Continue
	})
	return a, b, &got
}

func TestMoveToRoomMovesAnNPC(t *testing.T) {
	a, b, changes := moveSetup(t)
	m := mobs.NewTestInstance(mobs.Mob{})
	id := m.Character.Id
	m.Character.RoomId = moveRoomA
	a.AddMob(id)
	events.ProcessEvents()
	*changes = nil

	require.NoError(t, MoveToRoom(id, moveRoomB))
	events.ProcessEvents()

	assert.Equal(t, moveRoomB, m.Character.RoomId)
	assert.Equal(t, "movezone", m.Character.Zone)
	assert.Contains(t, b.GetBodies(), id)
	assert.NotContains(t, a.GetBodies(), id)
	require.Len(t, *changes, 1, "exactly one RoomChange per move")
	assert.Equal(t, events.RoomChange{ActorId: id, FromRoomId: moveRoomA, ToRoomId: moveRoomB}, (*changes)[0])
	assert.Contains(t, b.Containers, "chest", "no player arrived, so the room was not prepared")
	assert.Empty(t, b.userVisitors, "an NPC leaves no user trail")
	assert.Contains(t, a.mobVisitors, id)
	assert.Equal(t, 1, roomManager.roomsWithMobs[moveRoomB])
	assert.NotContains(t, roomManager.roomsWithUsers, moveRoomB)
}

func TestMoveToRoomMovesAPlayer(t *testing.T) {
	a, b, changes := moveSetup(t)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.RoomId = moveRoomA
	u := &accounts.Account{UserId: 9101, Character: c}
	actors.Spawn(c, actors.NewConnectionController(u))
	a.AddPlayer(u.UserId)

	require.NoError(t, MoveToRoom(c.Id, moveRoomB))
	events.ProcessEvents()

	assert.Equal(t, moveRoomB, c.RoomId)
	assert.Contains(t, b.GetBodies(), c.Id)
	assert.NotContains(t, a.GetBodies(), c.Id)
	require.Len(t, *changes, 1)
	assert.Equal(t, events.RoomChange{ActorId: c.Id, FromRoomId: moveRoomA, ToRoomId: moveRoomB}, (*changes)[0])
	assert.Contains(t, a.userVisitors, u.UserId, "the trail is left in the old room")
	assert.Contains(t, b.userVisitors, u.UserId)
	assert.NotContains(t, b.Containers, "chest", "the first player arriving prepares the room")
	assert.Equal(t, 1, roomManager.roomsWithUsers[moveRoomB])
}

func TestMoveToRoomUnknownActor(t *testing.T) {
	moveSetup(t)
	assert.Error(t, MoveToRoom(ids.New(ids.NoCell), moveRoomB))
}
