package main

import (
	"context"
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

// quit logs out the polite way (meditate, then the server hangs up).
func (s *script) quit(step string) {
	s.do(step, "quit", "begin your meditation")
	if s.err == nil {
		s.err = s.c.waitClosed(s.ctx, time.Minute, step+": disconnect")
	}
}

// play runs the whole smoke script for one bot against addr, then stays
// logged in for hold before quitting.
//
// New characters go through the tutorial (rooms 900-903) because it is the
// only place the stock world hands each player an item of their own: the
// graduation cap in room 903. Shared floor items would not survive 50 bots.
func play(ctx context.Context, addr string, b bot, timeout, hold time.Duration) error {
	c, err := dial(ctx, addr, timeout)
	if err != nil {
		return &stepError{"connect", err.Error(), ""}
	}
	defer c.Close()
	s := &script{ctx: ctx, c: c}

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
	s.do("start: confirm name", "yes", "skip the tutorial?")
	s.do("start: enter tutorial", "no", "Learning to Look")

	// Room 900. The tutorial script only accepts its lessons in this order.
	s.do("tutorial: look", "look", "Learning to Look", "Exits: east")
	s.do("tutorial: look orb", "look orb", "Orb of Vision is in perfect health")
	s.do("tutorial: look again", "look", "Learning to Look")
	s.do("tutorial: look east", "look east", "You peer toward the east", "Learning about You")
	s.do("walk: east", "east", "You head towards the east exit", "Learning about You")

	// Room 901.
	s.do("tutorial: status", "status", b.char+" the", "Gold:")
	s.field("tutorial: inventory", "inventory", "Carrying:", "sharp stick")
	s.do("tutorial: experience", "experience", "XP:")
	s.do("tutorial: conditions", "conditions", "Conditions")
	s.do("walk: south", "south", "You head towards the south exit", "Learning to Fight")

	// Room 902.
	s.do("tutorial: equip stick", "equip stick", "You wield your sharp stick")
	// One 1d2 hit per 4 s round, and the dummy sometimes dodges.
	s.doFor("tutorial: attack dummy", 3*time.Minute, "attack dummy", "training dummy has died")
	s.do("walk: west", "west", "You head towards the west exit", "Training Complete")

	// Room 903: the orb drops a cap on the floor of this player's own copy
	// of the room. Equipping it is the lesson that frees every command.
	s.wait("item: cap on the floor", "drops their graduation cap")
	s.do("item: get", "get cap", "You pick up the graduation cap")
	s.field("item: carried after get", "inventory", "Carrying:", "graduation cap")
	s.do("tutorial: equip cap", "equip cap", "You wear your graduation cap")
	s.do("item: remove", "remove cap", "You remove your graduation cap")
	s.do("item: drop", "drop cap", "You drop the graduation cap")
	s.field("item: on the ground after drop", "look", "On the Ground:", "graduation cap")
	s.do("item: get again", "get cap", "You pick up the graduation cap")
	s.field("item: carried after get again", "inventory", "Carrying:", "graduation cap")
	s.do("walk: portal", "portal", "Town Square")

	s.do("walk: north", "north", "Cobblestone Way")
	s.do("walk: south", "south", "Town Square")
	s.do("say", "say hello from "+b.char, `You say, "hello from `+b.char+`"`)
	s.quit("quit")
	if s.err != nil {
		return s.err
	}

	c, err = dial(ctx, addr, timeout)
	if err != nil {
		return &stepError{"reconnect", err.Error(), ""}
	}
	defer c.Close()
	s.c = c

	s.wait("relogin: prompt", `username (or "new")`)
	s.do("relogin: username", b.user, "password:")
	s.do("relogin: password", b.pass, "Welcome to the Mud!", "Town Square")
	s.do("relogin: character", "status", b.char+" the")
	s.field("relogin: still holds the item", "inventory", "Carrying:", "graduation cap")
	if s.err != nil {
		return s.err
	}

	if hold > 0 {
		select {
		case <-time.After(hold):
		case <-ctx.Done():
		}
	}
	s.quit("final quit")
	return s.err
}
