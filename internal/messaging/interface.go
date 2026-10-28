package messaging

// The interface, notice and raw paths (plan Task 2 step 5): text for one
// actor or account, notices for every active account, and bytes for a
// connection. Everything here ends in the writers in wire.go, apart from
// RawConn, which writes to a connection not registered yet.

import (
	"net"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

var initOnce sync.Once

// Init sets the function variables the packages below messaging send
// through (decision 10): rooms.SendMessage and accounts.Tell. main calls it
// once at startup, beside actors.Init; rooms and accounts cannot import this
// package, which imports them.
func Init() {
	initOnce.Do(func() {
		rooms.SendMessage = Send
		accounts.Tell = ToAccount
	})
}

// Tell writes interface text to one actor: no sense check, the actor's
// account only (a behaviour drops it). Empty text writes a blank line.
// Interface text is written as it is, never token-resolved.
func Tell(actorId ids.ID, text string) {
	Send(perception.Message{Kind: perception.Interface, Actor: actorId, Audience: perception.ActorOnly, ActorText: text})
}

// TellRaw is Tell without the closing newline (the print command).
func TellRaw(actorId ids.ID, text string) {
	if u := accountOf(actorId); u != nil {
		events.AddToQueue(events.Message{UserId: u.UserId, Text: text})
	}
}

// ToAccount writes interface text to an account whether or not it has an
// actor yet (login, character creation, logout).
func ToAccount(u *accounts.Account, text string) {
	if u == nil {
		return
	}
	u.SendText(text)
}

// Notice is text for every active account, with no sense check: saves,
// reloads, shutdown, the broadcast command and death notices.
type Notice struct {
	Text             string
	ScreenReaderText string // optional; a screen-reader account reads it instead
	SkipLineRefresh  bool
}

// Notify sends n to every active account, roomless ones included. NPCs get none.
func Notify(n Notice) {
	events.AddToQueue(events.Broadcast{
		Text:             n.Text,
		TextScreenReader: n.ScreenReaderText,
		SkipLineRefresh:  n.SkipLineRefresh,
	})
}

// Raw writes bytes to a connection as they are: screens before an account
// exists, telnet and terminal negotiation, echo.
func Raw(connId connections.ConnectionId, b []byte) {
	connections.SendTo(b, connId)
}

// RawAll writes bytes as they are to every connection past the login
// screens (connections.Broadcast skips those): the copyover notice, sent
// before any account is active again.
func RawAll(b []byte) {
	connections.Broadcast(b)
}

// RawConn writes to a connection that is not registered yet (server full).
func RawConn(c net.Conn, b []byte) {
	c.Write(b)
}

// WebClient sends a web client command (password masking).
func WebClient(connId connections.ConnectionId, text string) {
	events.AddToQueue(events.WebClientCommand{ConnectionId: connId, Text: text})
}

// UserIdOf is the account UserId behind an actor, or 0.
func UserIdOf(actorId ids.ID) int {
	if u := accountOf(actorId); u != nil {
		return u.UserId
	}
	return 0
}

// accountOf is the account driving the actor, or nil.
func accountOf(actorId ids.ID) *accounts.Account {
	a := actors.Get(actorId)
	if a == nil {
		return nil
	}
	if ctrl := a.Controller(); ctrl != nil {
		return ctrl.Account()
	}
	return nil
}
