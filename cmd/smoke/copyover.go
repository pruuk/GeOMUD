package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// copyoverCheck runs one copyover while every bot is online. Each bot calls
// arrived.Done when it finishes its script or fails; the admin waits for all
// of them, runs copyover and closes done. ran and err are set before done
// closes.
type copyoverCheck struct {
	arrived sync.WaitGroup
	done    chan struct{}
	ran     bool
	err     error
}

func newCopyoverCheck(bots int) *copyoverCheck {
	co := &copyoverCheck{done: make(chan struct{})}
	co.arrived.Add(bots)
	return co
}

// wait blocks until the step has ended and reports whether copyover ran.
func (co *copyoverCheck) wait(ctx context.Context) bool {
	select {
	case <-co.done:
		return co.ran
	case <-ctx.Done():
		return false
	}
}

var errNoNpcAfter = errors.New("copyover: the new process spawned no NPC")

// npcIDsByProcess splits the server log at its second "Starting server" line
// and returns the actor IDs of the "event=spawned kind=npc" lines before and
// after it. It errors when the log holds fewer than two starts.
func npcIDsByProcess(log string) (before, after map[string]bool, err error) {
	before, after = map[string]bool{}, map[string]bool{}
	starts := 0
	for _, line := range strings.Split(log, "\n") {
		if strings.Contains(line, `msg="Starting server`) {
			starts++
			continue
		}
		if !strings.Contains(line, "event=spawned") || !strings.Contains(line, "kind=npc") {
			continue
		}
		i := strings.Index(line, "actorId=")
		if i < 0 {
			continue
		}
		f := strings.Fields(line[i+len("actorId="):])
		if len(f) == 0 {
			continue
		}
		if starts >= 2 {
			after[f[0]] = true
		} else {
			before[f[0]] = true
		}
	}
	if starts < 2 {
		return nil, nil, fmt.Errorf("copyover: log has %d \"Starting server\" lines, no second process to check", starts)
	}
	return before, after, nil
}

// checkNpcIDsKept fails when the new process spawned an NPC ID the old one
// never had (a fresh NPC replaced a restored one), or spawned none.
func checkNpcIDsKept(log string) error {
	before, after, err := npcIDsByProcess(log)
	if err != nil {
		return err
	}
	if len(after) == 0 {
		return errNoNpcAfter
	}
	var fresh []string
	for id := range after {
		if !before[id] {
			fresh = append(fresh, id)
		}
	}
	if len(fresh) > 0 {
		sort.Strings(fresh)
		return fmt.Errorf("copyover: new process spawned NPC IDs the old one never had: %s", strings.Join(fresh, ", "))
	}
	return nil
}

// checkNpcsAfterCopyover polls the server log until the restored NPCs appear
// (or a fresh ID does), up to timeout.
func checkNpcsAfterCopyover(ctx context.Context, srv *server, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		b, _ := os.ReadFile(srv.logPath())
		err := checkNpcIDsKept(string(b))
		if err == nil || !errors.Is(err, errNoNpcAfter) || time.Now().After(deadline) {
			return err
		}
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// run waits for every bot, then logs in as the admin and runs copyover. The
// server process this program started exits; srv adopts the new one.
func (co *copyoverCheck) run(ctx context.Context, addr string, adm adminLogin, timeout time.Duration, srv *server) {
	defer close(co.done)
	co.arrived.Wait()
	s, err := loginAdmin(ctx, addr, adm.user, adm.pass, timeout)
	if err != nil {
		co.err = err
		return
	}
	defer s.c.Close()
	s.c.send("copyover")
	failed := ""
	s.err = s.c.until(ctx, 2*time.Minute, "copyover: run", func(got string, _ bool) (bool, string) {
		if i := strings.Index(got, "Copyover failed:"); i >= 0 {
			failed = strings.TrimSpace(strings.SplitN(got[i:], "\n", 2)[0])
			return true, ""
		}
		if strings.Contains(got, "Copyover complete.") {
			return true, ""
		}
		return false, `no "Copyover complete."`
	})
	if s.err == nil && failed != "" {
		s.err = &stepError{"copyover: run", failed, s.c.screen()}
	}
	if s.err == nil {
		// The resumed connection drops a line within Timing.TurnMs of its resume (main.go resumeRestoredConnection sets lastInput), as a new connection does.
		s.c.lastSend = time.Now()
	}
	if s.err == nil {
		s.err = srv.adoptCopyoverChild(ctx, time.Minute)
	}
	if s.err == nil {
		// The bots wait in the start room, so the new process restores its NPCs and logs them.
		if err := checkNpcsAfterCopyover(ctx, srv, 5*time.Second); err != nil {
			s.err = &stepError{"copyover: npc ids", err.Error(), s.c.screen()}
		}
	}
	if s.err == nil {
		co.ran = true
	}
	s.quit("copyover: admin quit")
	co.err = s.err
}
