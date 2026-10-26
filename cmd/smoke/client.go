package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/smoke/telnet"
)

// client is one telnet connection. A reader goroutine collects the server's
// text (ANSI stripped) and answers option negotiation the way Mudlet does:
// it accepts GMCP and refuses everything else.
type client struct {
	conn     net.Conn
	timeout  time.Duration
	lastSend time.Time // only the script goroutine sends

	writeMu sync.Mutex

	mu      sync.Mutex
	text    strings.Builder
	mark    int            // waits search the text after this offset; send moves it
	closed  bool           // the server closed the connection
	gmcp    map[string]int // GMCP package names received, with counts
	changed chan struct{}
}

// gmcpNames returns the GMCP package names this connection has received.
func (c *client) gmcpNames() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int, len(c.gmcp))
	for k, v := range c.gmcp {
		out[k] = v
	}
	return out
}

func dial(ctx context.Context, addr string, timeout time.Duration) (*client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return newClient(conn, timeout), nil
}

func newClient(conn net.Conn, timeout time.Duration) *client {
	// lastSend starts at connect: the server's drop window also opens when it
	// sends the first prompt (main.go handleTelnetConnection sets lastInput
	// after the prompt), so an instant first reply is lost too.
	c := &client{conn: conn, timeout: timeout, changed: make(chan struct{}, 1), gmcp: map[string]int{}, lastSend: time.Now()}
	go c.read()
	return c
}

func (c *client) Close() error { return c.conn.Close() }

func (c *client) write(b []byte) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.conn.Write(b)
}

func (c *client) read() {
	parser := telnet.NewParser()
	buf := make([]byte, 4096)
	for {
		n, err := c.conn.Read(buf)
		for _, tok := range parser.Feed(buf[:n]) {
			switch tok.Kind {
			case telnet.TokenText:
				clean := strings.ReplaceAll(string(telnet.StripAnsi(tok.Text)), "\r", "")
				c.mu.Lock()
				c.text.WriteString(clean)
				c.mu.Unlock()
				c.notify()
			case telnet.TokenIAC:
				c.negotiate(tok.Command, tok.Option)
			case telnet.TokenGMCP:
				c.mu.Lock()
				c.gmcp[tok.GMCPPackage]++
				c.mu.Unlock()
			}
		}
		if err != nil {
			c.mu.Lock()
			c.closed = true
			c.mu.Unlock()
			c.notify()
			return
		}
	}
}

func (c *client) negotiate(cmd, opt byte) {
	switch {
	case cmd == telnet.WILL && opt == telnet.GMCP:
		c.write(telnet.DoGMCP())
		hello, _ := json.Marshal(map[string]string{"client": "geomud-smoke", "version": "1"})
		c.write(telnet.FrameGMCP("Core.Hello", string(hello)))
		sup, _ := json.Marshal(telnet.SupportedPackages)
		c.write(telnet.FrameGMCP("Core.Supports.Set", string(sup)))
	case cmd == telnet.WILL:
		c.write([]byte{telnet.IAC, telnet.DONT, opt})
	case cmd == telnet.DO:
		c.write([]byte{telnet.IAC, telnet.WONT, opt})
	}
}

func (c *client) notify() {
	select {
	case c.changed <- struct{}{}:
	default:
	}
}

// sendGap spaces out commands. The server silently drops a line that arrives
// within Timing.TurnMs (50 ms in the stock config) of the previous one
// (main.go, the EnterPressed branch), so a bot that answers a prompt the
// instant it appears loses its answer.
const sendGap = 200 * time.Millisecond

// send writes one line. Later waits only look at text that arrives after it.
func (c *client) send(line string) {
	if d := time.Until(c.lastSend.Add(sendGap)); d > 0 {
		time.Sleep(d)
	}
	c.lastSend = time.Now()
	c.mu.Lock()
	c.mark = c.text.Len()
	c.mu.Unlock()
	c.write([]byte(line + "\r\n"))
}

// screen returns the text received since the last send.
func (c *client) screen() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.text.String()[c.mark:]
}

// wait blocks until every want has appeared since the last send.
func (c *client) wait(ctx context.Context, step string, wants ...string) error {
	return c.waitFor(ctx, c.timeout, step, wants...)
}

func (c *client) waitFor(ctx context.Context, timeout time.Duration, step string, wants ...string) error {
	return c.until(ctx, timeout, step, func(got string, _ bool) (bool, string) {
		for _, w := range wants {
			if !strings.Contains(got, w) {
				return false, fmt.Sprintf("no %q", w)
			}
		}
		return true, ""
	})
}

// waitField waits for a "label ..." line whose value contains item, such as
// "Carrying: newbie kit (1), graduation cap".
func (c *client) waitField(ctx context.Context, step, label, item string) error {
	return c.until(ctx, c.timeout, step, func(got string, _ bool) (bool, string) {
		val, ok := field(got, label)
		if !ok {
			return false, fmt.Sprintf("no %q line", label)
		}
		if !strings.Contains(val, item) {
			return false, fmt.Sprintf("%q line is %q, want %q in it", label, val, item)
		}
		return true, ""
	})
}

// waitClosed waits for the server to hang up.
func (c *client) waitClosed(ctx context.Context, timeout time.Duration, step string) error {
	return c.until(ctx, timeout, step, func(_ string, closed bool) (bool, string) {
		return closed, "connection still open"
	})
}

// until polls ok against the text since the last send until it reports done.
// ok returns a description of what is missing for the failure message.
func (c *client) until(ctx context.Context, timeout time.Duration, step string, ok func(got string, closed bool) (bool, string)) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		c.mu.Lock()
		got := c.text.String()[c.mark:]
		closed := c.closed
		c.mu.Unlock()

		done, missing := ok(got, closed)
		switch {
		case done:
			return nil
		case closed:
			return &stepError{step, "connection closed: " + missing, got}
		}

		select {
		case <-c.changed:
		case <-timer.C:
			return &stepError{step, fmt.Sprintf("%s after %s", missing, timeout), got}
		case <-ctx.Done():
			return &stepError{step, "interrupted: " + missing, got}
		}
	}
}

// do sends a command and waits for every want in the reply.
func (c *client) do(ctx context.Context, step, cmd string, wants ...string) error {
	c.send(cmd)
	return c.wait(ctx, step, wants...)
}

// stepError names the failed step and carries the last screen of text.
type stepError struct {
	step, problem, screen string
}

func (e *stepError) Error() string {
	return fmt.Sprintf("step %q: %s\n--- last screen ---\n%s\n---", e.step, e.problem, lastLines(e.screen, 30))
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n "), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// field returns the rest of the last line that starts with label (after
// leading spaces), trimmed, e.g. field(inv, "Carrying:") = "newbie kit (1)".
func field(screen, label string) (string, bool) {
	val, found := "", false
	for _, line := range strings.Split(screen, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimLeft(line, " │"), label); ok {
			val, found = strings.TrimSpace(rest), true
		}
	}
	return val, found
}
