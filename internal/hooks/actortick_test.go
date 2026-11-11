package hooks

import (
	"strconv"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/behaviour"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collect runs f and returns the messages sent and the inputs queued for mob.
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

func TestActorUpkeepTicksEveryActorsCooldowns(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Cooldowns = characters.Cooldowns{"kick": 3}
	actors.Spawn(c, actors.NewConnectionController(&accounts.Account{UserId: 61, Character: c}))
	m := mobs.NewTestInstance(mobs.Mob{})
	t.Cleanup(func() { mobs.Rest(m.Character.Id) })
	m.Character.Cooldowns = characters.Cooldowns{"bite": 3}
	m.Character.Health = 10 // a bare test NPC is at 0, which is dead: upkeep would end it

	ActorUpkeep(events.NewRound{RoundNumber: 1})

	assert.Equal(t, 2, c.Cooldowns["kick"], "a player's cooldowns tick")
	assert.Equal(t, 2, m.Character.Cooldowns["bite"], "an NPC's cooldowns tick")
}

func TestActorTickRunsEachNPCsBehaviourOnce(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	actors.Spawn(c, actors.NewConnectionController(&accounts.Account{UserId: 62, Character: c}))
	m := mobs.NewTestInstance(mobs.Mob{})
	t.Cleanup(func() { mobs.Rest(m.Character.Id) })
	require.NotNil(t, actors.Get(m.Character.Id))

	ticks := 0
	actors.OnBehaviourTick.Register(func(x *mobs.Mob) *mobs.Mob {
		if x == m {
			ticks++
		}
		return x
	})

	ActorTick(events.NewRound{RoundNumber: 1})

	assert.Equal(t, 1, ticks)
}

// The behaviour package's Tick reaches an NPC through the round: main calls
// behaviour.Init, and ActorTick runs each NPC's controller. An NPC on a path
// takes its first step.
func TestActorTickWalksAnNPCsPathThroughBehaviour(t *testing.T) {
	behaviour.Init()
	const from, to = 9301, 9302
	zone := "Wired Zone"
	a := rooms.NewEmptyRoom()
	a.RoomId, a.Zone = from, zone
	a.Exits["north"] = exit.RoomExit{RoomId: to}
	rooms.AddTestRoom(a)
	t.Cleanup(func() { rooms.RemoveTestRoom(from) })
	b := rooms.NewEmptyRoom()
	b.RoomId, b.Zone = to, zone
	b.Exits["south"] = exit.RoomExit{RoomId: from}
	rooms.AddTestRoom(b)
	t.Cleanup(func() { rooms.RemoveTestRoom(to) })

	actors.Reset()
	t.Cleanup(actors.Reset)
	m := mobs.Mob{MobId: 9, HomeRoomId: from}
	m.Character = *characters.New()
	m.Character.Name = "rat"
	m.Character.RoomId = from
	m.Character.Zone = zone
	mob := mobs.NewTestInstance(m)
	t.Cleanup(func() { mobs.Rest(mob.Character.Id) })

	behaviour.PathTo(mob, strconv.Itoa(to))
	require.Equal(t, 1, mob.Path.Len(), "the path was planned")

	_, inputs := collect(t, mob, func() {
		ActorTick(events.NewRound{RoundNumber: 1})
	})

	assert.Contains(t, inputs, "north")
}
