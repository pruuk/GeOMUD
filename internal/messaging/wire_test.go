package messaging

import (
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/stretchr/testify/assert"
)

// The byte tests for the four writers moved here from internal/hooks (plan
// Task 2 step 2). Every expectation is written from the old hook code, read
// at c08ee9e4:
//   - lineStart is the cursor-column + erase-line prefix the text writers put
//     before each text (Message_SendMessages.go:31 and :80,
//     Broadcast_SendToAll.go:46 and :63); ESC [ G and ESC [ 2 K are
//     term.AnsiMoveCursorColumn and term.AnsiEraseLine (term.go:176, :141).
//   - text goes through templates.AnsiParse (Message_SendMessages.go:27, :75;
//     Broadcast_SendToAll.go:23, :28), so a coloured line is expected as
//     AnsiParse of its markup;
//   - a screen-reader account's text has box characters, "[" and "]" turned
//     into spaces (Message_SendMessages.go:28-30, util.StripCharsForScreenReaders);
//   - a Broadcast gives a screen-reader account its screen-reader text when
//     there is one (Broadcast_SendToAll.go:35-54), unstripped, and Text
//     otherwise (:56-66); SkipLineRefresh drops the prefix (:39-43, :56-60);
//   - each text write queues a RedrawPrompt at priority 100
//     (Message_SendMessages.go:33, :82; Broadcast_SendToAll.go:33), so the
//     prompt follows the text, once per user (RedrawPrompt is a unique event);
//     the prompt writer sends AnsiParse(GetCommandPrompt()) with no prefix of
//     its own (RedrawPrompt_SendRedraw.go:38-39), and GetCommandPrompt starts
//     with the same prefix (accounts/userrecord.prompt.go, GetCommandPrompt's return);
//   - OnlyIfChanged skips an unchanged prompt (RedrawPrompt_SendRedraw.go:24-36);
//   - a RoomId message reaches the room's players minus UserId and
//     ExcludeUserIds (Message_SendMessages.go:38-64);
//   - a web client command reaches a websocket only
//     (WebClientCommand_SendWebClientCommand.go:18-20).
//
// The connection itself turns "\n" into "\r\n" (ConnectionDetails.Write).

const (
	lineStart = "\x1b[G\x1b[2K"
	prompt    = lineStart + "P> "
)

// fakeConn is a net.Conn that keeps what is written to it.
type fakeConn struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (c *fakeConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.Write(b)
}
func (c *fakeConn) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}
func (c *fakeConn) Read(b []byte) (int, error)         { return 0, nil }
func (c *fakeConn) Close() error                       { return nil }
func (c *fakeConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *fakeConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *fakeConn) SetDeadline(t time.Time) error      { return nil }
func (c *fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *fakeConn) SetWriteDeadline(t time.Time) error { return nil }

// writers registers the four writers for one test, as hooks.RegisterListeners does.
func writers(t *testing.T) {
	t.Helper()
	events.ProcessEvents() // nothing left over from another test
	m := events.RegisterListener(events.Message{}, WriteMessage)
	p := events.RegisterListener(events.RedrawPrompt{}, WritePrompt)
	w := events.RegisterListener(events.WebClientCommand{}, WriteWebClientCommand)
	b := events.RegisterListener(events.Broadcast{}, WriteBroadcast)
	accounts.ResetActiveUsers()
	actors.Reset()
	t.Cleanup(func() {
		actors.Reset()
		events.UnregisterListener(events.Message{}, m)
		events.UnregisterListener(events.RedrawPrompt{}, p)
		events.UnregisterListener(events.WebClientCommand{}, w)
		events.UnregisterListener(events.Broadcast{}, b)
		accounts.ResetActiveUsers()
	})
}

// wiredAccount is an active account on a fake connection, with the prompt
// "P> ", and no actor.
func wiredAccount(t *testing.T, userId int, screenReader bool) (*accounts.Account, *fakeConn) {
	t.Helper()
	conn := &fakeConn{}
	cd := connections.Add(conn, nil)
	t.Cleanup(func() { connections.Remove(cd.ConnectionId()) })
	u := accounts.NewAccount(userId, cd.ConnectionId())
	u.ScreenReader = screenReader
	u.ConfigOptions[`prompt-compiled`] = `P> `
	accounts.SetTestUser(u)
	accounts.SetTestConnection(cd.ConnectionId(), userId)
	return u, conn
}

// wiredUser is wiredAccount with an actor, so a room can list the player.
func wiredUser(t *testing.T, userId int, screenReader bool) (*accounts.Account, *fakeConn) {
	t.Helper()
	u, conn := wiredAccount(t, userId, screenReader)
	u.Character.Id = ids.New(ids.NoCell)
	actors.Spawn(u.Character, actors.NewConnectionController(u))
	return u, conn
}

func crlf(s string) string { return strings.ReplaceAll(s, "\n", "\r\n") }

func TestWriteMessageToAUser(t *testing.T) {
	writers(t)
	u, conn := wiredUser(t, 9501, false)
	coloured := `<ansi fg="red">red</ansi> text` + "\n"

	events.AddToQueue(events.Message{UserId: u.UserId, Text: "hello there\n"})
	events.AddToQueue(events.Message{UserId: u.UserId, Text: coloured})
	events.ProcessEvents()

	assert.Equal(t,
		lineStart+"hello there\r\n"+
			lineStart+crlf(templates.AnsiParse(coloured))+
			prompt,
		conn.String())
}

func TestWriteMessageStripsForAScreenReader(t *testing.T) {
	writers(t)
	u, conn := wiredUser(t, 9502, true)

	events.AddToQueue(events.Message{UserId: u.UserId, Text: "┌─┐ box [x]\n"})
	events.ProcessEvents()

	assert.Equal(t, lineStart+"    box  x \r\n"+prompt, conn.String())
}

func TestWriteMessageToARoom(t *testing.T) {
	writers(t)
	r := rooms.NewEmptyRoom()
	r.RoomId = 95100
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(r.RoomId) })
	a, connA := wiredUser(t, 9503, false)
	b, connB := wiredUser(t, 9504, false)
	c, connC := wiredUser(t, 9505, false)
	d, connD := wiredUser(t, 9506, false)
	for _, u := range []*accounts.Account{a, b, c, d} {
		r.AddPlayer(u.UserId)
	}

	events.AddToQueue(events.Message{UserId: a.UserId, RoomId: r.RoomId, ExcludeUserIds: []int{b.UserId}, Text: "room\n"})
	events.ProcessEvents()

	assert.Equal(t, lineStart+"room\r\n"+prompt, connA.String(), "the UserId gets the text once")
	assert.Equal(t, "", connB.String(), "excluded")
	assert.Equal(t, lineStart+"room\r\n"+prompt, connC.String())
	assert.Equal(t, lineStart+"room\r\n"+prompt, connD.String())
}

func TestWriteBroadcast(t *testing.T) {
	writers(t)
	_, plain := wiredUser(t, 9507, false)
	_, reader := wiredUser(t, 9508, true)

	events.AddToQueue(events.Broadcast{Text: "all [x]\n", TextScreenReader: "for readers\n"})
	events.ProcessEvents()

	assert.Equal(t, lineStart+"all [x]\r\n"+prompt, plain.String())
	assert.Equal(t, lineStart+"for readers\r\n"+prompt, reader.String())
}

func TestWriteBroadcastSkipLineRefreshAndNoScreenReaderText(t *testing.T) {
	writers(t)
	_, plain := wiredUser(t, 9509, false)
	_, reader := wiredUser(t, 9510, true)

	events.AddToQueue(events.Broadcast{Text: "saved [x]\n", SkipLineRefresh: true})
	events.ProcessEvents()

	// No screen-reader text: the reader gets Text, unstripped (Broadcast_SendToAll.go:56-60).
	assert.Equal(t, "saved [x]\r\n"+prompt, plain.String())
	assert.Equal(t, "saved [x]\r\n"+prompt, reader.String())
}

func TestWritePromptOnlyIfChanged(t *testing.T) {
	writers(t)
	u, conn := wiredUser(t, 9511, false)

	events.AddToQueue(events.RedrawPrompt{UserId: u.UserId, OnlyIfChanged: true})
	events.ProcessEvents()
	events.AddToQueue(events.RedrawPrompt{UserId: u.UserId, OnlyIfChanged: true})
	events.ProcessEvents()
	events.AddToQueue(events.RedrawPrompt{UserId: u.UserId})
	events.ProcessEvents()

	assert.Equal(t, prompt+prompt, conn.String(), "the unchanged prompt is skipped once, then forced")
}

func TestWriteWebClientCommandSkipsATelnetConnection(t *testing.T) {
	writers(t)
	u, conn := wiredUser(t, 9512, false)

	events.AddToQueue(events.WebClientCommand{ConnectionId: u.ConnectionId(), Text: "TEXTMASK:true"})
	events.ProcessEvents()

	assert.Equal(t, "", conn.String())
}
