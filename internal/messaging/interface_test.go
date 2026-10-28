package messaging

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queued records the text events the interface paths queue, in order.
func queued(t *testing.T) *[]events.Event {
	t.Helper()
	events.ProcessEvents() // nothing left over from another test
	got := &[]events.Event{}
	keep := func(e events.Event) events.ListenerReturn { *got = append(*got, e); return events.Continue }
	m := events.RegisterListener(events.Message{}, keep)
	b := events.RegisterListener(events.Broadcast{}, keep)
	w := events.RegisterListener(events.WebClientCommand{}, keep)
	actors.Reset()
	t.Cleanup(func() {
		events.UnregisterListener(events.Message{}, m)
		events.UnregisterListener(events.Broadcast{}, b)
		events.UnregisterListener(events.WebClientCommand{}, w)
		actors.Reset()
	})
	return got
}

// player is an actor driven by an account (UserId userId), in no room.
func player(userId int) (ids.ID, *accounts.Account) {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = "Ann"
	u := &accounts.Account{UserId: userId, Character: c}
	actors.Spawn(c, actors.NewConnectionController(u))
	return c.Id, u
}

// npc is an actor driven by a behaviour.
func npc() ids.ID {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = "rat"
	actors.Spawn(c, actors.NewBehaviourController(nil))
	return c.Id
}

func TestTellWritesTheLineToAPlayer(t *testing.T) {
	got := queued(t)
	ann, _ := player(9601)

	Tell(ann, "hello")
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.Message{UserId: 9601, Text: "hello\n"}}, *got)
}

func TestTellWritesNothingToAnNPC(t *testing.T) {
	got := queued(t)
	rat := npc()

	Tell(rat, "hello")
	TellRaw(rat, "hello")
	events.ProcessEvents()

	assert.Empty(t, *got)
}

func TestTellEmptyWritesABlankLine(t *testing.T) {
	got := queued(t)
	ann, _ := player(9602)

	Tell(ann, "")
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.Message{UserId: 9602, Text: "\n"}}, *got)
}

func TestTellIsNotTokenResolved(t *testing.T) {
	got := queued(t)
	ann, _ := player(9603)

	Tell(ann, "{actor} {{x}} {{ .Name }} {target} {direction}")
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.Message{UserId: 9603, Text: "{actor} {{x}} {{ .Name }} {target} {direction}\n"}}, *got)
}

func TestTellRawHasNoNewline(t *testing.T) {
	got := queued(t)
	ann, _ := player(9604)

	TellRaw(ann, "hello")
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.Message{UserId: 9604, Text: "hello"}}, *got)
}

func TestToAccountWorksWithNoActor(t *testing.T) {
	got := queued(t)
	u := &accounts.Account{UserId: 9605}

	ToAccount(u, "welcome")
	ToAccount(nil, "nobody")
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.Message{UserId: 9605, Text: "welcome\n"}}, *got)
}

func TestNotifyQueuesABroadcast(t *testing.T) {
	got := queued(t)

	Notify(Notice{Text: "saving", ScreenReaderText: "saving now", SkipLineRefresh: true})
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.Broadcast{Text: "saving", TextScreenReader: "saving now", SkipLineRefresh: true}}, *got)
}

func TestNotifyReachesAnAccountWithNoActor(t *testing.T) {
	writers(t)
	_, plain := wiredAccount(t, 9606, false)
	_, reader := wiredAccount(t, 9607, true)

	Notify(Notice{Text: "shutting down [x]\n", ScreenReaderText: "shutting down\n"})
	events.ProcessEvents()

	assert.Equal(t, lineStart+"shutting down [x]\r\n"+prompt, plain.String())
	assert.Equal(t, lineStart+"shutting down\r\n"+prompt, reader.String(), "the screen-reader text")
}

func TestRawWritesToAConnection(t *testing.T) {
	writers(t)
	u, conn := wiredAccount(t, 9608, false)

	Raw(u.ConnectionId(), []byte("login: "))
	events.ProcessEvents()

	assert.Equal(t, "login: ", conn.String(), "no prefix, no prompt")
}

func TestRawAllWritesToEveryConnectionPastLogin(t *testing.T) {
	writers(t)
	u, conn := wiredAccount(t, 9612, false)
	connections.Get(u.ConnectionId()).SetState(connections.LoggedIn)
	_, loginConn := wiredAccount(t, 9613, false) // still at the login screens

	RawAll([]byte("\r\nCopyover complete.\r\n"))
	events.ProcessEvents()

	// The connection turns every "\n" into "\r\n", as it did for
	// connections.Broadcast: the bytes on the wire are unchanged.
	assert.Equal(t, crlf("\r\nCopyover complete.\r\n"), conn.String(), "no prefix, no prompt")
	assert.Equal(t, "", loginConn.String(), "a connection at the login screens is skipped")
}

func TestRawConnWritesToAnUnregisteredConn(t *testing.T) {
	conn := &fakeConn{}
	RawConn(conn, []byte("Server is full.\n"))
	assert.Equal(t, "Server is full.\n", conn.String(), "as it is: the connection layer is not involved")
}

func TestWebClientQueuesACommand(t *testing.T) {
	got := queued(t)

	WebClient(connections.ConnectionId(7), "TEXTMASK:true")
	events.ProcessEvents()

	assert.Equal(t, []events.Event{events.WebClientCommand{ConnectionId: 7, Text: "TEXTMASK:true"}}, *got)
}

func TestUserIdOf(t *testing.T) {
	queued(t)
	ann, _ := player(9609)
	assert.Equal(t, 9609, UserIdOf(ann))
	assert.Equal(t, 0, UserIdOf(npc()))
	assert.Equal(t, 0, UserIdOf(ids.New(ids.NoCell)), "no such actor")
}

func TestInitSetsTheFunctionVariables(t *testing.T) {
	Init()
	require.NotNil(t, rooms.SendMessage)
	require.NotNil(t, accounts.Tell)

	got := queued(t)
	accounts.Tell(&accounts.Account{UserId: 9610}, "hi")
	ann, _ := player(9611)
	rooms.SendMessage(perception.Message{Kind: perception.Interface, Actor: ann, Audience: perception.ActorOnly, ActorText: "there"})
	events.ProcessEvents()

	assert.Equal(t, []events.Event{
		events.Message{UserId: 9610, Text: "hi\n"},
		events.Message{UserId: 9611, Text: "there\n"},
	}, *got)
}
