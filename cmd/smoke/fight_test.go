package main

import (
	"context"
	"testing"
	"time"
)

// A bot that fails to connect must still release the other bot's wait on it.
func TestPlayConnectFailureReleasesFightWaiters(t *testing.T) {
	for idx, name := range []string{"fighter (done)", "observer (ready)"} {
		fc := newFightCheck(2, "Smkabc0")
		_, err := play(context.Background(), "127.0.0.1:1", bot{idx: idx, user: "u", pass: "p", char: "c"},
			500*time.Millisecond, 0, nil, nil, fc)
		if err == nil {
			t.Fatalf("%s: expected a connect error", name)
		}
		ch := fc.done
		if idx == 1 {
			ch = fc.ready
		}
		select {
		case <-ch:
		default:
			t.Errorf("%s: channel still open after a connect failure", name)
		}
	}
}

func TestCheckObserverText(t *testing.T) {
	const f = "Smkabc0"
	cases := []struct {
		name, got string
		wantErr   bool
	}{
		{"observer lines", "Smkabc0 prepares to fight rat.\n[☀️ 4:01AM HP:6/6 MP:5/5]:Smkabc0 hits rat with their fists.\nrat has died.\n", false},
		{"swing after a blank prompt line", "[☀️ 4:01AM HP:6/6 MP:5/5]:\nSmkabc0 swings at rat, but misses.\nrat has died.\n", false},
		{"possessive name", "[☀️ 4:01AM HP:6/6 MP:5/5]:Smkabc0 prepares to fight rat.\n[☀️ 4:01AM HP:6/6 MP:5/5]:Smkabc0's fists connects with rat.\nrat has died.\n", false},
		{"only the prepare line", "Smkabc0 prepares to fight rat.\nrat has died.\n", true},
		{"attacker line", "Smkabc0 hits rat with their fists.\n[☀️ 4:01AM HP:6/6 MP:5/5]:You miss with your fists.\nYou swing your fists at rat.\n", true},
		{"attacker prepare", "Smkabc0 hits rat.\nYou prepare to fight rat!\n", true},
		{"xp line", "Smkabc0 hits rat.\nYou gained 44 experience points! (combat)\n", true},
		{"crit line", "Smkabc0 prepares to fight rat.\n*** Smkabc0 hits rat with their fists. ***\nrat has died.\n", false},
		{"backstab crit line", "*[BACKSTAB]* *** Smkabc0 hits rat with their fists. ***\n", false},
		{"backstab line after prompt", "[☀️ 4:01AM HP:6/6 MP:5/5]:*[BACKSTAB]* Smkabc0 hits rat.\n", false},
		{"crit attacker line", "Smkabc0 prepares to fight rat.\n*** You hit rat with your fists. ***\n", true},
		{"unrelated you line", "Smkabc0 hits rat.\nYou say, \"hi\"\n", false},
	}
	for _, c := range cases {
		err := checkObserverText(c.got, f)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, want error %v", c.name, err, c.wantErr)
		}
	}
}
