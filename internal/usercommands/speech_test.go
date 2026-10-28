package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNextRoomId = 92002

// captureMessages records every room Message queued while fn runs.
func captureMessages(t *testing.T, fn func()) []events.Message {
	t.Helper()
	var got []events.Message
	id := events.RegisterListener(events.Message{}, func(e events.Event) events.ListenerReturn {
		if m, ok := e.(events.Message); ok && m.RoomId != 0 {
			got = append(got, m)
		}
		return events.Continue
	})
	defer events.UnregisterListener(events.Message{}, id)
	events.ProcessEvents() // drop anything queued earlier
	got = nil
	fn()
	events.ProcessEvents()
	return got
}

// joinNextRoom adds a room north of the test room, with an exit back south.
func joinNextRoom(t *testing.T) {
	t.Helper()
	here := rooms.LoadRoom(testRoomId)
	require.NotNil(t, here)
	here.Exits["north"] = exit.RoomExit{RoomId: testNextRoomId}

	next := rooms.NewEmptyRoom()
	next.RoomId, next.Zone = testNextRoomId, "cmdzone"
	next.Exits["south"] = exit.RoomExit{RoomId: testRoomId}
	rooms.AddTestRoom(next)
	t.Cleanup(func() {
		delete(here.Exits, "north")
		rooms.RemoveTestRoom(testNextRoomId)
	})
}

func messagesFor(msgs []events.Message, roomId int) []string {
	out := []string{}
	for _, m := range msgs {
		if m.RoomId == roomId {
			out = append(out, m.Text)
		}
	}
	return out
}

func TestNameColour(t *testing.T) {
	npc, player, _ := setup(t)
	assert.Equal(t, "mobname", NameColour(npc))
	assert.Equal(t, "username", NameColour(player))
}

func TestNPCSayReachesTheRoomWithMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"

	msgs := captureMessages(t, func() {
		_, err := Say("hello there", npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> says, "<ansi fg="saytext">hello there</ansi>"` + "\n"},
		messagesFor(msgs, testRoomId))
}

func TestPlayerSayTextUnchanged(t *testing.T) {
	_, player, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	player.Character.Name = "Ann"

	msgs := captureMessages(t, func() {
		_, err := Say("hello there", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="username">Ann</ansi> says, "<ansi fg="saytext">hello there</ansi>"` + "\n"},
		messagesFor(msgs, testRoomId))
}

func TestNPCShoutReachesTheNextRoomUpperCased(t *testing.T) {
	npc, _, _ := setup(t)
	joinNextRoom(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"

	msgs := captureMessages(t, func() {
		_, err := Shout("it's time to die!", npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> shouts, "<ansi fg="yellow">IT'S TIME TO DIE!</ansi>"` + "\n"},
		messagesFor(msgs, testRoomId))
	assert.Equal(t,
		[]string{`Someone shouts from the <ansi fg="exit">south</ansi> direction, "<ansi fg="yellow">IT'S TIME TO DIE!</ansi>"` + "\n"},
		messagesFor(msgs, testNextRoomId))
}

func TestNPCEmoteUsesMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"

	msgs := captureMessages(t, func() {
		_, err := Emote("wiggles its nose", npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> <ansi fg="20">wiggles its nose</ansi>` + "\n"},
		messagesFor(msgs, testRoomId))
}

func TestNPCBroadcastUsesMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"

	var got []events.Broadcast
	id := events.RegisterListener(events.Broadcast{}, func(e events.Event) events.ListenerReturn {
		if b, ok := e.(events.Broadcast); ok {
			got = append(got, b)
		}
		return events.Continue
	})
	defer events.UnregisterListener(events.Broadcast{}, id)

	_, err := Broadcast("squeak", npc, room, 0)
	require.NoError(t, err)
	events.ProcessEvents()
	require.Len(t, got, 1)
	assert.Contains(t, got[0].Text, `<ansi fg="mobname">Rat</ansi>: <ansi fg="broadcast-body">squeak</ansi>`)
}
