package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNPCSuicideDestroysTheInstanceAndQueuesMobDeath(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	require.NotNil(t, mobs.GetInstance(npc.Id))

	var deaths []events.MobDeath
	id := events.RegisterListener(events.MobDeath{}, func(e events.Event) events.ListenerReturn {
		if d, ok := e.(events.MobDeath); ok {
			deaths = append(deaths, d)
		}
		return events.Continue
	})
	defer events.UnregisterListener(events.MobDeath{}, id)
	events.ProcessEvents()

	handled, err := Suicide("", npc, room, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	events.ProcessEvents()

	assert.Nil(t, mobs.GetInstance(npc.Id))
	require.Len(t, deaths, 1)
	assert.Equal(t, npc.Id, deaths[0].ActorId)
	assert.Equal(t, testRoomId, deaths[0].RoomId)
}
