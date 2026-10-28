package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// fightStepTimeout is how long bot 0 waits for the rat to die. Rounds are 4 s
// and `attack rat` may pick a tougher rat (measured: up to about 110 s, 27
// rounds), so 4 minutes leaves room on a loaded CI runner.
const fightStepTimeout = 4 * time.Minute

// fightRetryGap is how long bot 0 waits before attacking again after the rat
// was not there (it wandered off) or the target was lost.
const fightRetryGap = 2 * time.Second

// observerWaitTimeout bounds how long bot 0 waits for the observer to reach
// Town Square. The observer is another bot running the same script
// concurrently, so this covers the signup lock and a full floor.
const observerWaitTimeout = 5 * time.Minute

// fightCheck syncs the smoke fight. Bot 0 fights the rat in Town Square; bot 1
// (when there is one) stands in the room and reads what an observer sees.
// ready closes when bot 1 is in the room (or has ended, on any path), and
// done closes when bot 0's fight step has ended (on any path). ok is set
// before done closes and tells bot 1 that the rat died, so there is
// something to check.
type fightCheck struct {
	observer  bool // a bot 1 exists
	ready     chan struct{}
	done      chan struct{}
	readyOnce sync.Once
	doneOnce  sync.Once
	ok        bool
	fighter   string // bot 0's character name, for the observer's checks
}

func newFightCheck(bots int, fighterChar string) *fightCheck {
	return &fightCheck{
		observer: bots >= 2,
		ready:    make(chan struct{}),
		done:     make(chan struct{}),
		fighter:  fighterChar,
	}
}

func (fc *fightCheck) markReady() { fc.readyOnce.Do(func() { close(fc.ready) }) }
func (fc *fightCheck) markDone()  { fc.doneOnce.Do(func() { close(fc.done) }) }

// fight is bot 0's step: wait for the observer, attack the rat and wait for
// its death line and the XP line. The fight is over, and the bot's aggro
// cleared, by the time the death line arrives (npcDeath clears it before it
// grants XP), so the later steps start with no fight running.
func (s *script) fight(fc *fightCheck) {
	if fc == nil {
		return
	}
	defer fc.markDone()
	if s.err != nil {
		return
	}
	if fc.observer {
		select {
		case <-fc.ready:
		case <-s.ctx.Done():
			s.err = &stepError{"fight: wait for the observer", "interrupted", s.c.screen()}
			return
		case <-time.After(observerWaitTimeout):
			s.err = &stepError{"fight: wait for the observer", "bot 1 never reached Town Square in " + observerWaitTimeout.String(), s.c.screen()}
			return
		}
	}

	const step = "fight: attack rat until it dies"
	deadline := time.Now().Add(fightStepTimeout)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			s.err = &stepError{step, "the rat did not die in " + fightStepTimeout.String(), s.c.screen()}
			return
		}
		s.c.send("attack rat")
		retry := false
		s.err = s.c.until(s.ctx, left, step, func(got string, _ bool) (bool, string) {
			if strings.Contains(got, "rat has died.") {
				// Any rat dying ends the step; the XP line follows in the same round.
				if strings.Contains(got, "You gained") && strings.Contains(got, "experience points") && strings.Contains(got, "(combat)") {
					return true, ""
				}
				return false, `no "You gained N experience points ... (combat)" line after "rat has died."`
			}
			if strings.Contains(got, "You attack the darkness!") || strings.Contains(got, "Your target can't be found.") {
				retry = true
				return true, ""
			}
			return false, `no "rat has died."`
		})
		if s.err != nil || !retry {
			if s.err == nil {
				fc.ok = true
			}
			return
		}
		select {
		case <-time.After(fightRetryGap):
		case <-s.ctx.Done():
			s.err = &stepError{step, "interrupted", s.c.screen()}
			return
		}
	}
}

// ratWord matches the rat as a whole word.
var ratWord = regexp.MustCompile(`\brat\b`)

// promptPrefix matches the status prompt, e.g. "[☀️ 4:01AM HP:6/6 MP:5/5]:".
var promptPrefix = regexp.MustCompile(`^\[[^\]]*\]:`)

// stripCombatMarks removes the leading marks combat puts on a line (from
// internal/combat/combat.go, ansi stripped): "*[BACKSTAB]* " on a backstab
// and "*** " on a crit, in that order when both apply.
func stripCombatMarks(line string) string {
	for {
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "*[BACKSTAB]*"), "***"))
		if trimmed == line {
			return line
		}
		line = trimmed
	}
}

// checkObserverText checks what a bystander read during bot fighter's fight:
// at least one line naming the fighter and the rat besides the "prepares to
// fight" line (the observer form of a round's swing line), and no attacker-only
// text (a line that starts "You"/"Your" and names the rat, or the XP line).
func checkObserverText(got, fighter string) error {
	swing := false
	for _, line := range strings.Split(got, "\n") {
		// A line that arrives right after the status prompt starts with it.
		line = strings.TrimSpace(promptPrefix.ReplaceAllString(strings.TrimSpace(line), ""))
		line = stripCombatMarks(line)
		// The name may be possessive: "Smk..c0's fists connects with rat."
		if (strings.HasPrefix(line, fighter+" ") || strings.HasPrefix(line, fighter+"'")) && ratWord.MatchString(line) && !strings.Contains(line, "prepares to fight") {
			swing = true
		}
		you := strings.HasPrefix(line, "You ") || strings.HasPrefix(line, "Your ")
		if you && (ratWord.MatchString(line) || strings.Contains(line, "experience points")) {
			return fmt.Errorf("the observer read attacker-only text: %q", line)
		}
	}
	if !swing {
		return fmt.Errorf("the observer read no line naming %s and the rat (other than \"prepares to fight\")", fighter)
	}
	return nil
}

// observe is bot 1's step: confirm it stands in Town Square, tell bot 0 it can
// attack, wait for the fight to end and check the text it read meanwhile. The
// look sets the text mark, so the screen holds everything since it.
func (s *script) observe(fc *fightCheck) {
	if fc == nil {
		return
	}
	defer fc.markReady()
	s.do("fight observer: in Town Square", "look", startRoomTitle)
	if s.err != nil {
		return
	}
	fc.markReady()
	select {
	case <-fc.done:
	case <-s.ctx.Done():
		s.err = &stepError{"fight observer: wait for the fight", "interrupted", s.c.screen()}
		return
	}
	if !fc.ok {
		return // bot 0 failed or was cut short: its error is the one that counts
	}
	s.wait("fight observer: sees the rat die", "rat has died.")
	if s.err != nil {
		return
	}
	if err := checkObserverText(s.c.screen(), fc.fighter); err != nil {
		s.err = &stepError{"fight observer: observer lines", err.Error(), s.c.screen()}
	}
}
