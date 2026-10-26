package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var signupMu sync.Mutex

// bot is one player account the script creates and plays.
type bot struct {
	user, pass, char string
}

// script runs steps in order and stops at the first failure: each method
// is a no-op once err is set.
type script struct {
	ctx context.Context
	c   *client
	err error

	gmcp map[string]int // packages received on connections already quit
}

func (s *script) do(step, cmd string, wants ...string) {
	if s.err == nil {
		s.err = s.c.do(s.ctx, step, cmd, wants...)
	}
}

// doFor is do with its own timeout, for steps that take several rounds.
func (s *script) doFor(step string, timeout time.Duration, cmd string, wants ...string) {
	if s.err == nil {
		s.c.send(cmd)
		s.err = s.c.waitFor(s.ctx, timeout, step, wants...)
	}
}

func (s *script) wait(step string, wants ...string) {
	if s.err == nil {
		s.err = s.c.wait(s.ctx, step, wants...)
	}
}

// field sends cmd and checks that its "label" line lists item.
func (s *script) field(step, cmd, label, item string) {
	if s.err == nil {
		s.c.send(cmd)
		s.err = s.c.waitField(s.ctx, step, label, item)
	}
}

// dropItem drops smokeItem. The server refuses a drop while the room floor
// holds Gameplay.FloorItemCountMax items (10 in the stock config) and hands the
// item back with "too many items on the floor". The admin put one item per bot
// on the floor, so with more bots than that limit, early droppers are refused
// until enough bots have picked theirs up. Every bot's first get does not
// depend on any drop, so the floor empties and the retries end. A refused drop
// is retried once a second until timeout.
func (s *script) dropItem(step string, timeout time.Duration) {
	if s.err != nil {
		return
	}
	deadline := time.Now().Add(timeout)
	for {
		s.c.send("drop " + smokeItem)
		refused := false
		s.err = s.c.until(s.ctx, s.c.timeout, step, func(got string, _ bool) (bool, string) {
			switch {
			case strings.Contains(got, "You drop the "+smokeItem):
				return true, ""
			case strings.Contains(got, "too many items on the floor"):
				refused = true
				return true, ""
			}
			return false, fmt.Sprintf("no %q", "You drop the "+smokeItem)
		})
		if s.err != nil || !refused {
			return
		}
		if time.Now().After(deadline) {
			s.err = &stepError{step, "the floor stayed full (too many items) for " + timeout.String(), s.c.screen()}
			return
		}
		select {
		case <-time.After(time.Second):
		case <-s.ctx.Done():
			s.err = &stepError{step, "interrupted", s.c.screen()}
			return
		}
	}
}

// quit logs out the polite way: quit applies the meditating buff (buff 0) and
// the server hangs up when it expires. Its "begin your meditation" messages
// were the buff's JS script, removed in Phase 4, so quit is silent until the
// disconnect.
func (s *script) quit(step string) {
	if s.err == nil {
		s.c.send("quit")
		s.err = s.c.waitClosed(s.ctx, time.Minute, step+": disconnect")
	}
	// The connection is closed (or the step failed): nothing more will arrive
	// that the check could count, so fold this connection's packages in.
	if s.gmcp == nil {
		s.gmcp = map[string]int{}
	}
	for n, k := range s.c.gmcpNames() {
		s.gmcp[n] += k
	}
}

// play runs the whole smoke script for one bot against addr, then stays
// logged in for hold before quitting.
//
// The admin has already put one smokeItem per bot on the floor of the start
// room (prepareItems). The item counts need no lock: every bot gets once
// before it drops, a bot wants at most one item at a time, and the floor
// always holds one item for every bot that is not carrying one. So a bot's
// first get, and its get after its own drop, finds an item whatever order
// the bots run in. (The floor item limit can delay a drop; see dropItem.)
//
// It also returns the GMCP package names the bot received (both connections
// together), and fails the bot unless they include Char, Room and Comm packages.
func play(ctx context.Context, addr string, b bot, timeout, hold time.Duration) (map[string]int, error) {
	c, err := dial(ctx, addr, timeout)
	if err != nil {
		return nil, &stepError{"connect", err.Error(), ""}
	}
	defer c.Close()
	s := &script{ctx: ctx, c: c, gmcp: map[string]int{}}

	s.wait("login prompt", `username (or "new")`)
	s.do("account: new", "new", "Choose your username")
	s.do("account: username", b.user, "Choose your password")
	s.do("account: password", b.pass, "Repeat your password")
	s.do("account: repeat password", b.pass, "Email Address")
	s.do("account: email", "", "screen reader")
	s.do("account: screen reader", "n", "create a new user/login named "+b.user)
	// One signup at a time: users.CreateUser reads the highest user ID and
	// adds the new one without a lock, so two signups at once can get the
	// same ID and then share each other's output. Known engine bug, not
	// what this test is for.
	signupMu.Lock()
	s.do("account: create", "y", "Type start to begin playing")
	signupMu.Unlock()

	s.do("start: race prompt", "start", "Which race will you be?")
	s.do("start: race", "human", "known as (name)?")
	s.do("start: name", b.char, "Choose the name "+b.char+"?")
	// There is no tutorial question: the world has no tutorial rooms, so a
	// new character goes straight to the start room.
	s.do("start: confirm name", "yes", startRoomTitle)

	s.do("item: get", "get "+smokeItem, "You pick up the "+smokeItem)
	s.field("item: carried after get", "inventory", "Carrying:", smokeItem)
	s.dropItem("item: drop", 3*time.Minute)
	s.field("item: on the ground after drop", "look", "On the Ground:", smokeItem)
	s.do("item: get again", "get "+smokeItem, "You pick up the "+smokeItem)
	s.field("item: carried after get again", "inventory", "Carrying:", smokeItem)

	// The start zone (startland) has two rooms: Town Square (1) and End of
	// the Line (2), joined by north and south.
	s.do("walk: north", "north", "End of the Line")
	s.do("walk: south", "south", startRoomTitle)
	s.do("say", "say hello from "+b.char, `You say, "hello from `+b.char+`"`)
	s.quit("quit")
	if s.err != nil {
		return s.gmcp, s.err
	}

	c, err = dial(ctx, addr, timeout)
	if err != nil {
		return s.gmcp, &stepError{"reconnect", err.Error(), ""}
	}
	defer c.Close()
	s.c = c

	s.wait("relogin: prompt", `username (or "new")`)
	s.do("relogin: username", b.user, "password:")
	s.do("relogin: password", b.pass, "Welcome to the Mud!", startRoomTitle)
	s.do("relogin: character", "online", b.char)
	s.field("relogin: still holds the item", "inventory", "Carrying:", smokeItem)

	// Death, with the Phase 4 stand-in (docs/plans/2026-10-26-phase-4-scripting-surface.md):
	// suicide revives the player at 1 health in a private copy of the recovery
	// room (75, "Waiting room"), which has a fixed portal back to the start room.
	// Walking out proves the player is not downed (go is refused while downed).
	// Bots are level 1, under Death.ProtectionLevels, so nothing is dropped.
	s.do("death: suicide", "suicide", b.char+" has DIED!")
	s.do("death: in the recovery room", "look", recoveryRoomTitle)
	s.do("death: walk out", "portal", startRoomTitle)
	s.field("death: still holds the item", "inventory", "Carrying:", smokeItem)
	if s.err != nil {
		return s.gmcp, s.err
	}

	if hold > 0 {
		select {
		case <-time.After(hold):
		case <-ctx.Done():
		}
	}
	s.quit("final quit")
	if s.err == nil {
		s.err = checkGMCP(s.gmcp)
	}
	return s.gmcp, s.err
}

// requiredGMCP are the package families a Mudlet-style client must receive
// during the script (the Phase 4 plan's Mudlet check).
var requiredGMCP = []string{"Char.", "Room.", "Comm."}

// checkGMCP fails unless got holds at least one package under each required family.
func checkGMCP(got map[string]int) error {
	var missing []string
	for _, prefix := range requiredGMCP {
		found := false
		for name := range got {
			if strings.HasPrefix(name, prefix) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, prefix+"*")
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return &stepError{"gmcp", fmt.Sprintf("no %s package received; got: %s", strings.Join(missing, ", "), gmcpList(got)), ""}
}

// gmcpList formats package names with counts, sorted, for messages.
func gmcpList(m map[string]int) string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	for i, n := range names {
		names[i] = fmt.Sprintf("%s(%d)", n, m[n])
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, " ")
}
