package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLastStartPid(t *testing.T) {
	log := "time=1 msg=\"Starting server         \" name=GeOMUD pid=100\n" +
		"time=2 msg=other pid=5\n" +
		"time=3 msg=\"Starting server         \" name=GeOMUD pid=200\n"
	assert.Equal(t, 200, lastStartPid(log))
	assert.Equal(t, 0, lastStartPid("time=1 msg=\"Starting server         \" name=GeOMUD\n"))
}

func TestCopyoverCheckWaitsForEveryBot(t *testing.T) {
	co := newCopyoverCheck(2)
	started := make(chan struct{})
	go func() {
		co.arrived.Wait()
		close(started)
		co.ran = true
		close(co.done)
	}()
	co.arrived.Done()
	select {
	case <-started:
		t.Fatal("copyover started before every bot arrived")
	case <-time.After(50 * time.Millisecond):
	}
	co.arrived.Done()
	assert.True(t, co.wait(context.Background()))
}
