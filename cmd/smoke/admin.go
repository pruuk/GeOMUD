package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// smokeItem is what the admin spawns for the bots: one copper ring (item
// 20005 in the empty world) per bot. The name is unique in the empty world,
// so "get copper ring" can only match the spawned ones.
const smokeItem = "copper ring"

// startRoomTitle is the title of the start room (room 1, startland).
const startRoomTitle = "Town Square"

// prepareItems logs in as the admin and drops n copies of smokeItem on the
// floor of the start room, then quits. The bots take them from there.
//
// The stock admin account has a plaintext password, which GoMud makes the
// player change before any other command works. The change is handled in
// changePasswordIfForced.
func prepareItems(ctx context.Context, addr, user, pass string, n int, timeout time.Duration) error {
	c, err := dial(ctx, addr, timeout)
	if err != nil {
		return &stepError{"admin: connect", err.Error(), ""}
	}
	defer c.Close()
	s := &script{ctx: ctx, c: c}

	s.wait("admin: login prompt", `username (or "new")`)
	s.do("admin: username", user, "password:")
	// A forced-change account gets the "must change your password" notice
	// instead of the message of the day.
	if s.err == nil {
		s.c.send(pass)
		s.err = s.c.until(ctx, timeout, "admin: password", func(got string, _ bool) (bool, string) {
			if strings.Contains(got, "Welcome to the Mud!") || strings.Contains(got, "You must change your password") {
				return true, ""
			}
			return false, "neither the welcome nor the password-change notice"
		})
	}
	changePasswordIfForced(s, pass)

	// Stand in the start room (the stock admin character already does).
	if s.err == nil {
		s.c.send("look")
		s.err = s.c.wait(ctx, "admin: look", "Exits:")
	}
	if s.err == nil && !strings.Contains(s.c.screen(), startRoomTitle) {
		s.do("admin: go to the start room", "teleport 1", startRoomTitle)
	}

	for i := 1; i <= n; i++ {
		step := fmt.Sprintf("admin: spawn %d of %d", i, n)
		s.do(step, "item spawn "+smokeItem, "falls to the ground")
		if s.err != nil {
			break
		}
		// Check it landed, not only that the spawn message printed.
		s.field(step+": on the ground", "look", "On the Ground:", smokeItem)
	}
	s.quit("admin: quit")
	return s.err
}

// changePasswordIfForced answers GoMud's forced first-login password change.
// It sends `password` and sees which question comes back:
//
//   - "What new password would you like?": the account still has a plaintext
//     password, so there is no current-password question. The new password is
//     the one the admin just logged in with. It is now stored hashed, and
//     reruns with the same -admin-pass keep working.
//   - "What is your current password?": the password was changed on an
//     earlier run or by hand (the server is a long-lived target). Nothing is
//     forced, so a deliberately wrong answer cancels the prompt and leaves the
//     password alone.
func changePasswordIfForced(s *script, pass string) {
	if s.err != nil {
		return
	}
	s.c.send("password")
	var forced bool
	s.err = s.c.until(s.ctx, s.c.timeout, "admin: password question", func(got string, _ bool) (bool, string) {
		switch {
		case strings.Contains(got, "What new password would you like?"):
			forced = true
			return true, ""
		case strings.Contains(got, "What is your current password?"):
			return true, ""
		}
		return false, "neither password question"
	})
	if s.err != nil {
		return
	}
	if !forced {
		s.do("admin: password already changed", "not-the-password", "your password was incorrect")
		return
	}
	s.do("admin: new password", pass, "Confirm the change")
	s.do("admin: confirm new password", pass, "Your password has been changed!")
}
