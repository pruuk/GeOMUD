package main

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/smoke/telnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeServer returns a client wired to the far end of a pipe.
func fakeServer(t *testing.T) (*client, net.Conn) {
	t.Helper()
	near, far := net.Pipe()
	c := newClient(near, time.Second)
	t.Cleanup(func() { c.Close(); far.Close() })
	return c, far
}

func TestField(t *testing.T) {
	inv := " ┌─ .:Equipment ──┐\n │ Weapon:   sharp stick │\n └────┘\n Carrying: newbie kit (1), graduation cap\n\n[HP:6/6]:"
	val, ok := field(inv, "Carrying:")
	assert.True(t, ok)
	assert.Equal(t, "newbie kit (1), graduation cap", val)

	val, ok = field(inv, "Weapon:")
	assert.True(t, ok)
	assert.Equal(t, "sharp stick │", val)

	_, ok = field(inv, "On the Ground:")
	assert.False(t, ok)
}

func TestWaitSeesOnlyTextAfterSend(t *testing.T) {
	c, far := fakeServer(t)
	far.Write([]byte("old reply: You pick up the graduation cap\r\n"))
	require.NoError(t, c.wait(context.Background(), "old", "graduation cap"))

	go func() {
		buf := make([]byte, 64)
		far.Read(buf) // the command
		far.Write([]byte("\x1b[33mYou drop the graduation cap.\x1b[0m\r\n"))
	}()
	require.NoError(t, c.do(context.Background(), "drop", "drop cap", "You drop the graduation cap."))

	// The pick-up line came before the last send, so it no longer counts.
	c.timeout = 50 * time.Millisecond
	err := c.wait(context.Background(), "get", "You pick up")
	var se *stepError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, "get", se.step)
	assert.Contains(t, err.Error(), `no "You pick up"`)
	assert.Contains(t, err.Error(), "You drop the graduation cap.")
}

func TestWaitField(t *testing.T) {
	c, far := fakeServer(t)
	go far.Write([]byte(" Carrying: newbie kit (1)\r\n"))
	c.timeout = 100 * time.Millisecond
	err := c.waitField(context.Background(), "inv", "Carrying:", "graduation cap")
	assert.ErrorContains(t, err, `"Carrying:" line is "newbie kit (1)", want "graduation cap" in it`)
}

func TestWaitFailsWhenServerHangsUp(t *testing.T) {
	c, far := fakeServer(t)
	far.Write([]byte("Goodbye\r\n"))
	far.Close()
	err := c.wait(context.Background(), "quit", "never")
	assert.ErrorContains(t, err, "connection closed")
	require.NoError(t, c.waitClosed(context.Background(), time.Second, "closed"))
}

func TestNegotiationAcceptsGMCPOnly(t *testing.T) {
	_, far := fakeServer(t)
	go far.Write([]byte{telnet.IAC, telnet.WILL, 1, telnet.IAC, telnet.WILL, telnet.GMCP})

	got := make([]byte, 0, 256)
	buf := make([]byte, 256)
	far.SetReadDeadline(time.Now().Add(time.Second))
	for !strings.Contains(string(got), "Core.Supports.Set") {
		n, err := far.Read(buf)
		require.NoError(t, err)
		got = append(got, buf[:n]...)
	}
	assert.Equal(t, []byte{telnet.IAC, telnet.DONT, 1, telnet.IAC, telnet.DO, telnet.GMCP}, got[:6])
}

func TestClientRecordsGMCPPackages(t *testing.T) {
	c, far := fakeServer(t)
	go func() {
		far.Write(telnet.FrameGMCP("Room.Info", `{}`))
		far.Write(telnet.FrameGMCP("Room.Info", `{}`))
		far.Write(telnet.FrameGMCP("Comm.Channel", `{}`))
	}()
	require.Eventually(t, func() bool { return c.gmcpNames()["Comm.Channel"] == 1 }, time.Second, 5*time.Millisecond)
	assert.Equal(t, map[string]int{"Room.Info": 2, "Comm.Channel": 1}, c.gmcpNames())
}

func TestCheckGMCP(t *testing.T) {
	assert.NoError(t, checkGMCP(map[string]int{"Char.Vitals": 1, "Room.Info": 2, "Comm.Channel": 1}))

	err := checkGMCP(map[string]int{"Char.Vitals": 1, "Room.Info": 2, "Party": 1})
	assert.ErrorContains(t, err, "no Comm.* package received")
	assert.ErrorContains(t, err, "Char.Vitals(1) Party(1) Room.Info(2)")

	err = checkGMCP(nil)
	assert.ErrorContains(t, err, "no Char.*, Room.*, Comm.* package received; got: none")
}

func TestOverridesOnlyLoopback(t *testing.T) {
	got := overrides(filepath.Join("tmp", "w"), 4242)
	assert.Contains(t, got, `FilePaths.DataFiles: "tmp/w"`)
	assert.Contains(t, got, "Network.TelnetPort: []\n")
	assert.Contains(t, got, "Network.LocalPort: 4242\n")
	for _, k := range []string{"HttpPort", "HttpsPort", "SSHPort", "AI.Port"} {
		assert.Contains(t, got, "Network."+k+": 0\n")
	}
}
