package main

import (
	"context"
	"fmt"
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

const (
	startLine = "time=1 msg=\"Starting server         \" name=GeOMUD pid=%d\n"
	npcLine   = "time=2 msg=\"actor                   \" event=spawned kind=npc actorId=0x0:%s\n"
	mobLine   = "time=2 msg=\"actor                   \" event=spawned kind=player actorId=0x0:%s\n"
)

func fixture(before, after []string) string {
	s := fmt.Sprintf(startLine, 100)
	for _, id := range before {
		s += fmt.Sprintf(npcLine, id)
	}
	s += fmt.Sprintf(startLine, 200)
	for _, id := range after {
		s += fmt.Sprintf(npcLine, id)
	}
	return s
}

func TestCheckNpcIDsKept(t *testing.T) {
	assert.NoError(t, checkNpcIDsKept(fixture([]string{"a", "b", "c"}, []string{"a", "b"})))
}

func TestCheckNpcIDsKeptIgnoresNonNpc(t *testing.T) {
	log := fixture([]string{"a"}, []string{"a"}) + fmt.Sprintf(mobLine, "zzz")
	assert.NoError(t, checkNpcIDsKept(log))
}

func TestCheckNpcIDsKeptFreshID(t *testing.T) {
	err := checkNpcIDsKept(fixture([]string{"a", "b"}, []string{"a", "fresh1"}))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "fresh1")
		assert.NotContains(t, err.Error(), "0x0:a")
	}
}

func TestCheckNpcIDsKeptNoneAfter(t *testing.T) {
	err := checkNpcIDsKept(fixture([]string{"a"}, nil))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no NPC")
	}
}

func TestCheckNpcIDsKeptOneStart(t *testing.T) {
	err := checkNpcIDsKept(fmt.Sprintf(startLine, 100) + fmt.Sprintf(npcLine, "a"))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "second")
	}
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
