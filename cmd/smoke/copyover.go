package main

import (
	"context"
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
		co.ran = true
	}
	s.quit("copyover: admin quit")
	co.err = s.err
}
