package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// fightStepTimeout bounds each of bot 0's fights. Rounds are 4 s. Simulated
// (Phase 5 slice 4.1 plan, Task 6a Step 12), a fight against the level 1 or
// the level 5 rat of launch mode's world copy ends either way within 17
// rounds (68 s); 4 minutes leaves room on a loaded CI runner and for the
// walk back after a death.
const fightStepTimeout = 4 * time.Minute

// fightRetryGap is how long bot 0 waits before attacking again after the rat
// was not there or the target was lost.
const fightRetryGap = 2 * time.Second

// observerWaitTimeout bounds how long bot 0 waits for the observer to reach
// Town Square. The observer is another bot running the same script
// concurrently, so this covers the signup lock and a full floor.
const observerWaitTimeout = 5 * time.Minute

// ratTarget is one fight of the smoke: the word bot 0 attacks and the rat's
// full name, as its death line reads it ("<name> has died.").
type ratTarget struct{ word, name string }

// launchRats are launch mode's two fights (Phase 5 slice 4.1, ruling 11):
// the world copy (fightRats) spawns exactly these two in Town Square, the
// small one at level 1 and the large one at level 5, and neither wanders.
// Bot 0 fights the small rat first. Each fight may end either way.
var launchRats = []ratTarget{{"small", "small rat"}, {"large", "large rat"}}

// addrRats is -addr mode's one fight: a server someone else started keeps
// its own rats (any level, wandering), so bot 0 attacks whichever rat
// `attack rat` finds; that fight may end either way too.
var addrRats = []ratTarget{{"rat", "rat"}}

// fightCheck syncs the smoke fight. Bot 0 fights the rats in Town Square;
// bot 1 (when there is one) stands in the room and reads what an observer
// sees. ready closes when bot 1 is in the room (or has ended, on any path),
// and done closes when bot 0's fight step has ended (on any path). ok is set
// before done closes and tells bot 1 that every fight ended, so there is
// something to check; endings are the line each fight ended with, as bot 1
// reads it.
type fightCheck struct {
	observer  bool // a bot 1 exists
	ready     chan struct{}
	done      chan struct{}
	readyOnce sync.Once
	doneOnce  sync.Once
	ok        bool
	fighter   string      // bot 0's character name, for the observer's checks
	rats      []ratTarget // the fights, in order
	endings   []string    // per fight: "<rat name> has died." or "<fighter> has DIED!"
}

func newFightCheck(bots int, fighterChar string, rats []ratTarget) *fightCheck {
	return &fightCheck{
		observer: bots >= 2,
		ready:    make(chan struct{}),
		done:     make(chan struct{}),
		fighter:  fighterChar,
		rats:     rats,
	}
}

func (fc *fightCheck) markReady() { fc.readyOnce.Do(func() { close(fc.ready) }) }
func (fc *fightCheck) markDone()  { fc.doneOnce.Do(func() { close(fc.done) }) }

// fight is bot 0's step: wait for the observer, then fight each rat in turn.
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
	for _, rat := range fc.rats {
		ending := s.fightOne(fc.fighter, rat)
		if s.err != nil {
			return
		}
		fc.endings = append(fc.endings, ending)
	}
	fc.ok = true
}

// fightOne attacks rat until one of them dies and checks that ending
// (ruling 11): the rat's death ("<name> has died." and the XP line), or bot
// 0's ("<fighter> has DIED!", then the recovery room, the walk back to Town
// Square through its portal, and the ring taken back from its corpse, so the
// steps after the fight run as before). It returns the ending's line as an
// observer reads it.
func (s *script) fightOne(fighter string, rat ratTarget) string {
	step := "fight: attack " + rat.name
	ratDied, botDied := rat.name+" has died.", fighter+" has DIED!"
	deadline := time.Now().Add(fightStepTimeout)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			s.err = &stepError{step, "the fight did not end in " + fightStepTimeout.String(), s.c.screen()}
			return ""
		}
		s.c.send("attack " + rat.word)
		retry, ending := false, ""
		s.err = s.c.until(s.ctx, left, step, func(got string, _ bool) (bool, string) {
			if strings.Contains(got, botDied) {
				ending = botDied
				return true, ""
			}
			if strings.Contains(got, ratDied) {
				// The XP line follows the death line in the same round.
				if strings.Contains(got, "You gained") && strings.Contains(got, "experience points") && strings.Contains(got, "(combat)") {
					ending = ratDied
					return true, ""
				}
				return false, fmt.Sprintf(`no "You gained N experience points ... (combat)" line after %q`, ratDied)
			}
			if strings.Contains(got, "You attack the darkness!") || strings.Contains(got, "Your target can't be found.") {
				retry = true
				return true, ""
			}
			return false, fmt.Sprintf("no %q or %q", ratDied, botDied)
		})
		if s.err != nil {
			return ""
		}
		if !retry {
			if ending == botDied {
				s.do(step+": in the recovery room", "look", recoveryRoomTitle)
				s.do(step+": walk back", "portal", startRoomTitle)
				s.do(step+": take the item from the corpse", "get "+smokeItem+" "+fighter,
					"You take the "+smokeItem+" from the "+fighter+" corpse")
			}
			return ending
		}
		select {
		case <-time.After(fightRetryGap):
		case <-s.ctx.Done():
			s.err = &stepError{step, "interrupted", s.c.screen()}
			return ""
		}
	}
}

// ratWord matches the rat as a whole word.
var ratWord = regexp.MustCompile(`\brat\b`)

// promptPrefix matches the status prompt, e.g. "[☀️ 4:01AM HP:6/6 MP:5/5]:".
var promptPrefix = regexp.MustCompile(`^\[[^\]]*\]:`)

// stripCombatMarks removes the leading mark combat puts on a crit line (from
// internal/rules/dikubasic/combat.go, ansi stripped): "*** ".
func stripCombatMarks(line string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, "***"))
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

// sawEndings reports whether got holds every fight's ending, each line as
// many times as fights ended with it, and if not, which line is short.
func sawEndings(got string, endings []string) (bool, string) {
	want := map[string]int{}
	for _, e := range endings {
		want[e]++
	}
	for _, e := range endings {
		if n := strings.Count(got, e); n < want[e] {
			return false, fmt.Sprintf("%q read %d of %d times", e, n, want[e])
		}
	}
	return true, ""
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
	// Each fight's ending, whichever it was: the rat's death line or bot 0's.
	// Each ending must be read once per fight it ended, so bot 0 dying in
	// both fights needs its death line twice.
	s.err = s.c.until(s.ctx, s.c.timeout, "fight observer: sees each fight end", func(got string, _ bool) (bool, string) {
		return sawEndings(got, fc.endings)
	})
	if s.err != nil {
		return
	}
	if err := checkObserverText(s.c.screen(), fc.fighter); err != nil {
		s.err = &stepError{"fight observer: observer lines", err.Error(), s.c.screen()}
	}
}
