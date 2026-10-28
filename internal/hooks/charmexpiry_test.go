package hooks

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An ephemeral room, so that resting ends the NPC without writing the NPC
// store (the store seams are in package rooms' tests).
const charmTestRoomId = 2_000_000_000

func charmTestNPC(t *testing.T) (*mobs.Mob, *rooms.Room) {
	t.Helper()
	actors.Reset()
	t.Cleanup(actors.Reset)
	t.Cleanup(rooms.AddTestDefaultBiome()) // messaging asks the room's light
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = charmTestRoomId, "Charm Zone"
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(charmTestRoomId) })
	require.True(t, r.IsEphemeral())

	m := mobs.Mob{MobId: 5, HomeRoomId: charmTestRoomId}
	m.Character = *characters.New()
	m.Character.Name = "rat"
	m.Character.RoomId = charmTestRoomId
	m.Character.ActionPoints = 100
	m.Character.ActionPointsMax.Value = 100
	inst := mobs.NewTestInstance(m)
	r.AddTestBody(inst.Character.Id)
	t.Cleanup(func() { mobs.Rest(inst.Character.Id) })
	return inst, r
}

func collect(t *testing.T, mob *mobs.Mob, f func()) (messages []string, inputs []string) {
	t.Helper()
	mid := events.RegisterListener(events.Message{}, func(e events.Event) events.ListenerReturn {
		messages = append(messages, e.(events.Message).Text)
		return events.Continue
	})
	iid := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		if in := e.(events.Input); in.ActorId == mob.Character.Id {
			inputs = append(inputs, in.InputText)
		}
		return events.Continue
	})
	t.Cleanup(func() {
		events.UnregisterListener(events.Message{}, mid)
		events.UnregisterListener(events.Input{}, iid)
	})
	events.ProcessEvents()
	messages, inputs = nil, nil
	f()
	events.ProcessEvents()
	return messages, inputs
}

func TestCharmExpiredRestEmotesBeforeTheNPCLeaves(t *testing.T) {
	mob, r := charmTestNPC(t)
	id := mob.Character.Id

	// The farewell is perceived, not written to the room: a player in the
	// room reads it.
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = "Watcher"
	c.RoomId = charmTestRoomId
	u := &accounts.Account{UserId: 9101, Character: c, Password: "$2a$test"}
	actors.Spawn(c, actors.NewConnectionController(u))
	accounts.SetTestUser(u)
	r.AddPlayer(u.UserId)
	t.Cleanup(func() { accounts.RemoveTestUser(u.UserId) })

	messages, inputs := collect(t, mob, func() {
		runCharmExpiredCommand(mob, characters.CharmExpiredRest)
	})

	require.NotEmpty(t, messages, "the farewell emote reached the room")
	assert.True(t, strings.Contains(strings.Join(messages, ""), "bids you farewell"), "%q", messages)
	assert.Empty(t, inputs, "nothing is left queued for an NPC that is gone")
	assert.Nil(t, mobs.GetInstance(id), "and then the NPC rests out of memory")
	assert.Empty(t, r.GetMobs())
}

// An NPC at 0 health whose charm expires into a rest is gone, so the upkeep
// must not go on to queue its suicide.
func TestUpkeepDoesNotKillAnNPCThatRestedOnCharmExpiry(t *testing.T) {
	mob, _ := charmTestNPC(t)
	id := mob.Character.Id
	mob.Character.Charm(ids.ID{}, 0, characters.CharmExpiredRest)
	mob.Character.Health = 0
	require.True(t, mob.Character.IsCharmed())

	_, inputs := collect(t, mob, func() {
		ActorUpkeep(events.NewRound{RoundNumber: 1})
	})

	assert.Nil(t, mobs.GetInstance(id), "the NPC rested out of memory")
	assert.NotContains(t, inputs, "suicide")
}

func TestCharmExpiredWithoutRestQueuesTheCommands(t *testing.T) {
	mob, _ := charmTestNPC(t)

	messages, inputs := collect(t, mob, func() {
		runCharmExpiredCommand(mob, characters.CharmExpiredRevert)
	})

	assert.Empty(t, messages)
	assert.Equal(t, []string{"emote reverts to its old ways"}, inputs)
	assert.NotNil(t, mobs.GetInstance(mob.Character.Id))
}
