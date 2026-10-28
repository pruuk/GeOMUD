package usercommands

import (
	"fmt"
	"sort"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testNextRoomId = 92002
	observerId     = 9202 // a player watching in the test room
	farId          = 9203 // a player in the next room
)

// addReader puts a player who reads text in roomId: an account among the
// active users and its body in the room, which is what the writers need to
// deliver to it.
func addReader(t *testing.T, userId int, roomId int) *accounts.Account {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = fmt.Sprintf("Reader%d", userId)
	c.RoomId = roomId
	u := &accounts.Account{UserId: userId, Character: c, Password: "$2a$test"}
	actors.Spawn(c, actors.NewConnectionController(u))
	reading(t, u)
	return u
}

// reading makes an account a reader: active, its body in its room.
func reading(t *testing.T, u *accounts.Account) {
	t.Helper()
	accounts.SetTestUser(u)
	room := rooms.LoadRoom(u.Character.RoomId)
	require.NotNil(t, room)
	room.AddPlayer(u.UserId)
	t.Cleanup(func() { accounts.RemoveTestUser(u.UserId) })
}

// captureReads records what each player read while fn ran, by UserId. It
// expands the queued text events the way the writers deliver them
// (hooks.Message_SendMessage, hooks.Broadcast_SendToAll): a UserId message
// to that user; a RoomId message to the room's players at drain time minus
// the UserId and ExcludeUserIds, honouring IsQuiet; a Broadcast to every
// active account. After slice 7 every line arrives as a UserId message or a
// Broadcast, so the same capture keeps working.
func captureReads(t *testing.T, fn func()) map[int][]string {
	t.Helper()
	got := map[int][]string{}
	read := func(userId int, text string) {
		if accounts.GetByUserId(userId) != nil {
			got[userId] = append(got[userId], text)
		}
	}
	msgId := events.RegisterListener(events.Message{}, func(e events.Event) events.ListenerReturn {
		m := e.(events.Message)
		if m.UserId > 0 {
			read(m.UserId, m.Text)
		}
		if m.RoomId > 0 {
			room := rooms.LoadRoom(m.RoomId)
			if room == nil {
				return events.Continue
			}
		players:
			for _, userId := range room.GetPlayers() {
				if userId == m.UserId {
					continue
				}
				for _, ex := range m.ExcludeUserIds {
					if ex == userId {
						continue players
					}
				}
				if u := accounts.GetByUserId(userId); u != nil && m.IsQuiet && !u.Character.HasBuffFlag("superhearing") {
					continue
				}
				read(userId, m.Text)
			}
		}
		return events.Continue
	})
	bcId := events.RegisterListener(events.Broadcast{}, func(e events.Event) events.ListenerReturn {
		b := e.(events.Broadcast)
		all := accounts.GetAllActiveUsers()
		sort.Slice(all, func(i, j int) bool { return all[i].UserId < all[j].UserId })
		for _, u := range all {
			if u.ScreenReader && len(b.TextScreenReader) > 0 {
				read(u.UserId, b.TextScreenReader)
				continue
			}
			read(u.UserId, b.Text)
		}
		return events.Continue
	})
	defer events.UnregisterListener(events.Message{}, msgId)
	defer events.UnregisterListener(events.Broadcast{}, bcId)
	events.ProcessEvents() // drop anything queued earlier
	clear(got)
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

func TestNameColour(t *testing.T) {
	npc, player, _ := setup(t)
	assert.Equal(t, "mobname", NameColour(npc))
	assert.Equal(t, "username", NameColour(player))
}

func TestNPCSayReachesTheRoomWithMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"
	addReader(t, observerId, testRoomId)

	reads := captureReads(t, func() {
		_, err := Say("hello there", npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> says, "<ansi fg="saytext">hello there</ansi>"` + "\n"},
		reads[observerId])
}

func TestPlayerSayTextUnchanged(t *testing.T) {
	_, player, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	player.Character.Name = "Ann"
	addReader(t, observerId, testRoomId)

	reads := captureReads(t, func() {
		_, err := Say("hello there", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="username">Ann</ansi> says, "<ansi fg="saytext">hello there</ansi>"` + "\n"},
		reads[observerId])
}

func TestNPCShoutReachesTheNextRoomUpperCased(t *testing.T) {
	npc, _, _ := setup(t)
	joinNextRoom(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"
	addReader(t, observerId, testRoomId)
	addReader(t, farId, testNextRoomId)

	reads := captureReads(t, func() {
		_, err := Shout("it's time to die!", npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> shouts, "<ansi fg="yellow">IT'S TIME TO DIE!</ansi>"` + "\n"},
		reads[observerId])
	assert.Equal(t,
		[]string{`Someone shouts from the <ansi fg="exit">south</ansi> direction, "<ansi fg="yellow">IT'S TIME TO DIE!</ansi>"` + "\n"},
		reads[farId])
}

func TestNPCEmoteUsesMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"
	addReader(t, observerId, testRoomId)

	reads := captureReads(t, func() {
		_, err := Emote("wiggles its nose", npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> <ansi fg="20">wiggles its nose</ansi>` + "\n"},
		reads[observerId])
}

func TestNPCBroadcastUsesMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"
	addReader(t, observerId, testRoomId)

	reads := captureReads(t, func() {
		_, err := Broadcast("squeak", npc, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, reads[observerId], 1)
	assert.Contains(t, reads[observerId][0], `<ansi fg="mobname">Rat</ansi>: <ansi fg="broadcast-body">squeak</ansi>`)
}
