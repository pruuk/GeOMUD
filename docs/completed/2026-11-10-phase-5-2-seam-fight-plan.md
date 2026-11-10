# Phase 5 slice 2: the seam and the fight (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved spec `2026-11-10-phase-5-2-seam-fight-spec.md`: `internal/rules`,
`internal/combat` moved verbatim to `internal/rules/dikubasic`, the engine calling `rules.Get()`, an
import guard, stub-ruleset tests, and `shoot` through `CanAttack` with `AggroChanged`.

**Architecture:** See the spec ("Design"). This plan adds no design decision; any question found while
executing goes back to the spec.

**Tech stack:** Go, testify, combat and reads goldens in `internal/hooks`, `internal/guards` (Go syntax
tree), `cmd/smoke`.

Branch `strip/phase-5-2-seam-fight` (exists; holds the spec). Tasks run one after another on the
branch: they share files, so no parallel worktrees.

---

## Line numbers this plan relies on (verified 2026-11-10 at `c1ca8e7f`; the spec's facts table has the rest)

| What | Where |
|---|---|
| `combat.Attack`, `accountOf`, `CreditAccount` | `internal/combat/combat.go:22-50`, `:52-58`, `:60-65` |
| `calculateCombat` return and init | `combat.go:182`, `:184` |
| `DoCombat` loop; `fightersWithAccountsFirst` | `NewRound_DoCombat.go:33`; `:45-62` |
| `accountOf` (hooks) | `NewRound_DoCombat.go:64-70` |
| `lookForWeapon` call; roll | `:156-160`; `:305-309` |
| wait lines; `combat.Attack` | `:167`; `:182` |
| group hatred block | `:211-220` |
| flee chance block | `:406-417` |
| `consider` lookups | `consider.go:39-53` |
| `attack` NPC branch; player branch with `CanPvp` | `attack.go:188-190`; `:232-237` |
| `shoot` NPC branch; player branch | `shoot.go:93-97`; `:109-122` |
| guard controls | `guards/combat_test.go:41`, `:57-60`, `:69`, `:100` |
| `main.go` listeners | `main.go:206` |
| Test mains | `hooks/actionpoints_test.go:22`, `usercommands/usercommands_test.go:25` |

If a line has moved when you get there, find the code by its text and say so in the report.

## Agent rules (brief every subagent with these)

- Edit files with the Edit and Write tools. Never write a repo file from the shell (no `>`, `>>`,
  `sed -i`, Python open-write); a hook blocks it. `git mv` and `git rm` are fine.
- Read files in chunks of at most 400 lines; state "read N of N files" in the report.
- `git add <named paths>` only, never `-A` or `.`; commit messages through a quoted heredoc ending
  with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never `--no-verify`.
- Never push. Never touch `upstream`.
- Run every test command in the foreground with an explicit timeout of 300000 ms; never wait on a
  background run.
- A grep that proves absence must also show it can match (run the same pattern on a symbol that
  exists).
- `go test` always with `-count=1`. **Never run `-update` in this slice.** A golden diff is a failure
  to report, not to re-record.
- If a step's code does not compile or a prescribed failure does not happen, report it; do not paper
  over it.
- Run `gofmt -w` on every Go file you created or changed before the per-task check (import order and
  comment alignment in the plan's code blocks are not guaranteed gofmt-clean).

## Per-task check (end of every task)

```bash
go build ./...
go vet ./... 2>&1 | head
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'   # only the 3 known Windows failures (geomud-verifying)
gofmt -l internal cmd *.go                                    # prints nothing (covers new, untracked files)
git status --short -- internal/hooks/testdata/                # prints nothing
```

Report: build result, the failing-test lines, and the testdata `git status` output (must be empty).

---

### Task 0: Baseline (controller)

- [ ] On `strip/phase-5-2-seam-fight`, run the per-task check; save the output in the scratchpad as
  the baseline (expected: the 3 known Windows failures only).
- [ ] Commit this plan once Calabe approves it: `git add 2026-11-10-phase-5-2-seam-fight-plan.md`,
  message `Phase 5.2: plan (the seam and the fight)`.

### Task 1: `internal/rules` (Sonnet)

**Files:** create `internal/rules/rules.go`, `internal/rules/rules_test.go`.

- [ ] **Step 1: Write the failing test** `internal/rules/rules_test.go`:

```go
package rules

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake is a Ruleset whose methods are never called; it only has to exist.
type fake struct{ name string }

func (fake) TurnOrder(f []*actors.Actor) []*actors.Actor       { return f }
func (fake) ResolveAttack(a, d *actors.Actor) AttackResult     { return AttackResult{} }
func (fake) WaitLines(a, d *actors.Actor) []perception.Message { return nil }
func (fake) HatredRounds(a *actors.Actor) int                  { return 0 }
func (fake) FindsWeapon(npc *actors.Actor) bool                { return false }
func (fake) FleeEscapes(fleer, blocker *actors.Actor) bool     { return true }
func (fake) Consider(a, d *actors.Actor) float64               { return 1 }
func (fake) CanAttack(a, d *actors.Actor) error                { return nil }

func TestGetPanicsWithNoRuleset(t *testing.T) {
	restore := SetForTest(nil)
	defer restore()
	assert.PanicsWithValue(t, "rules: no ruleset is set; main must call rules.Set at boot", func() { Get() })
}

func TestSetForTestRestoresThePreviousRuleset(t *testing.T) {
	restoreOuter := SetForTest(fake{name: "outer"})
	defer restoreOuter()

	restore := SetForTest(fake{name: "inner"})
	require.Equal(t, fake{name: "inner"}, Get())
	restore()
	require.Equal(t, fake{name: "outer"}, Get())
}
```

- [ ] **Step 2: Run it; it fails** to compile (`undefined: SetForTest`, `AttackResult`).
  `go test ./internal/rules -count=1`
- [ ] **Step 3: Write `internal/rules/rules.go`:**

```go
// Package rules is the seam between the engine and a game's rules. The engine
// calls the active Ruleset for every game rule and never imports a ruleset;
// main.go chooses one at boot with Set. A ruleset never sends text: it
// returns messages, and the engine sends them.
package rules

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/perception"
)

// AttackResult is what one round of an attacker's attacks did. The ruleset
// works it out; the engine applies DamageToTarget, queues BuffTarget and
// sends Messages in order.
type AttackResult struct {
	Hit            bool  // at least one attack connected
	Crit           bool  // at least one attack was a critical hit
	BuffTarget     []int // buffs a crit puts on the defender
	DamageToTarget int   // total damage dealt to the defender

	// Messages are the round's lines, one per swing (a dodge included), in
	// the order they happened; messaging resolves {actor} and {target} per
	// receiver.
	Messages []perception.Message
}

// Ruleset is everything the engine asks a game's rules. A method that needs
// a random draw makes it itself. Phase 5 slice 2 holds the fight; later
// slices add the round, death and display.
type Ruleset interface {
	// TurnOrder orders the actors that may take a combat turn this round.
	// It receives every actor with a body and returns them in turn order.
	TurnOrder(fighters []*actors.Actor) []*actors.Actor

	// ResolveAttack works out one round of attacker's attacks on defender.
	// It changes no health: the engine applies the result.
	ResolveAttack(attacker, defender *actors.Actor) AttackResult

	// WaitLines are the lines for a round attacker spends waiting to swing.
	WaitLines(attacker, defender *actors.Actor) []perception.Message

	// HatredRounds is how many rounds an NPC's groups hate the account of
	// an attacker who just attacked it.
	HatredRounds(attacker *actors.Actor) int

	// FindsWeapon reports whether an unarmed NPC carrying items looks for a
	// weapon to equip this round.
	FindsWeapon(npc *actors.Actor) bool

	// FleeEscapes reports whether fleer gets past blocker, one of the
	// actors fighting it.
	FleeEscapes(fleer, blocker *actors.Actor) bool

	// Consider is attacker's odds against defender: above 1 favours the
	// attacker, 1 is even, 0 or less is hopeless.
	Consider(attacker, defender *actors.Actor) float64

	// CanAttack is nil when attacker may start a fight with defender, or
	// the refusal to show the attacker.
	CanAttack(attacker, defender *actors.Actor) error
}

// active is set once at boot, before the game loop starts, and swapped only
// by tests, which do not run in parallel with the loop.
var active Ruleset

// Set makes r the active ruleset. main.go calls it once at boot.
func Set(r Ruleset) { active = r }

// Get is the active ruleset. It panics when none is set: a boot that forgot
// to choose one must fail loudly, not fight by default rules.
func Get() Ruleset {
	if active == nil {
		panic("rules: no ruleset is set; main must call rules.Set at boot")
	}
	return active
}

// SetForTest makes r the active ruleset and returns the func that restores
// the previous one, for t.Cleanup.
func SetForTest(r Ruleset) (restore func()) {
	prev := active
	active = r
	return func() { active = prev }
}
```

- [ ] **Step 4: Run it; it passes.** `go test ./internal/rules -count=1` prints `ok`.
- [ ] **Step 5:** Per-task check. Commit `internal/rules/rules.go internal/rules/rules_test.go`:
  `Phase 5.2: internal/rules, the Ruleset seam (fight methods)`.

### Task 2: Move `internal/combat` to `internal/rules/dikubasic`, verbatim (Sonnet)

A pure move so git records renames and the reviewer sees only the package line and import paths.

**Files:** `git mv` the 10 `.go` files in `internal/combat/` to `internal/rules/dikubasic/`; `git rm
internal/combat/AGENTS.md internal/combat/CLAUDE.md`. Modify the five importers
(`hooks/NewRound_DoCombat.go`, `usercommands/consider.go`, `usercommands/suicide_npc.go`,
`web/api_v1_items.go`, `web/api_v1_mobs.go`), `internal/guards/combat_test.go`,
`internal/guards/guards_test.go` (:202), `internal/rules/dikubasic/weapon_rank_test.go` (:17) and the
comment at `cmd/smoke/fight.go:123`.

- [ ] **Step 1:** `mkdir -p internal/rules/dikubasic`, then `git mv internal/combat/<f>.go
  internal/rules/dikubasic/<f>.go` for each of: `armor_rank`, `attackresult`, `calculations`,
  `calculations_test`, `combat`, `mob_rank`, `simulate`, `simulate_test`, `weapon_rank`,
  `weapon_rank_test`. `git rm` the two `.md` files. `ls internal/combat` then fails.
- [ ] **Step 2:** In each moved file change `package combat` to `package dikubasic`. Nothing else.
- [ ] **Step 3:** In each of the five importers change the import
  `github.com/GoMudEngine/GoMud/internal/combat` to `github.com/GoMudEngine/GoMud/internal/rules/dikubasic`
  and every `combat.` selector to `dikubasic.`. Before and after, `grep -n '\bcombat\b' <file>`: only
  comments and strings may still say "combat". Then `gofmt -w` the five files (the new path sorts
  differently among the imports).
- [ ] **Step 4:** In `internal/guards/combat_test.go`: control path `"internal/combat/combat.go"` (:41)
  and the control message naming it (:50) become `internal/rules/dikubasic/combat.go`; the scan
  prefix `"internal/combat/"` (:69) becomes `"internal/rules/"`; the comment at :57 and the control
  message at :101 name `internal/rules`. (Task 3 rewrites the `Attack` control and the messages.) In `internal/guards/guards_test.go:202` the key
  `"internal/combat/simulate.go|newSimMob|InstantiateItems"` becomes
  `"internal/rules/dikubasic/simulate.go|newSimMob|InstantiateItems"`; `gofmt -w` the file (the map
  alignment changes). In `internal/rules/dikubasic/weapon_rank_test.go:17`
  `filepath.Join(filepath.Dir(thisFile), "..", "..")` becomes
  `filepath.Join(filepath.Dir(thisFile), "..", "..", "..")` (one folder deeper). In the comment at
  `cmd/smoke/fight.go:123` `internal/combat/combat.go` becomes `internal/rules/dikubasic/combat.go`.
- [ ] **Step 5: Prove absence:**
  `grep -rn --exclude-dir=.claude --exclude-dir=.git 'internal/combat' --include=*.go .` prints
  nothing (no trailing quote: the guards key has `/simulate.go` after it; `.claude` holds agent
  worktrees) (control: the same command with `internal/rooms` piped to `head -1` prints a hit).
- [ ] **Step 6:** Per-task check; after staging the edits, `git status --short` shows `R` lines for
  the 10 files (`RM` before the package-line edits are staged).
- [ ] **Step 7:** Commit the moves, the two removals, the five importers, the two guard files, the
  weapon rank test and `cmd/smoke/fight.go` (name only paths that exist now: `git mv` and `git rm`
  already staged the `internal/combat/...` side, and naming those in `git add` fails):
  `Phase 5.2: move internal/combat to internal/rules/dikubasic (verbatim, package rename only)`.

### Task 3: `dikubasic.Ruleset`; the engine calls `rules.Get()` (Sonnet; the controller reads the diff line by line)

Every random draw stays where it is. Copy the code blocks exactly.

**Files:** create `internal/rules/dikubasic/ruleset.go`; `git rm internal/rules/dikubasic/attackresult.go`;
modify `internal/rules/dikubasic/combat.go`, `internal/hooks/NewRound_DoCombat.go`,
`internal/usercommands/consider.go`, `internal/usercommands/attack.go`, `main.go`,
`internal/hooks/actionpoints_test.go`, `internal/usercommands/usercommands_test.go`,
`internal/guards/combat_test.go`; create `internal/rules/dikubasic/ruleset_test.go`,
`internal/usercommands/pvp_test.go`, `internal/hooks/attack_engine_test.go`.

- [ ] **Step 1: `ruleset.go`:**

```go
package dikubasic

import (
	"errors"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Ruleset is the reference ruleset: GoMud's rules, moved behind the seam
// unchanged (Phase 5, decision 1). Its methods keep every random draw where
// the engine made it before, so the combat goldens hold.
type Ruleset struct{}

var _ rules.Ruleset = Ruleset{}

// New is the reference ruleset, for rules.Set.
func New() Ruleset { return Ruleset{} }

// TurnOrder puts actors with an account first, then NPCs, each group in the
// order given (players always swing before NPCs).
func (Ruleset) TurnOrder(fighters []*actors.Actor) []*actors.Actor {
	ordered := make([]*actors.Actor, 0, len(fighters))
	for _, a := range fighters {
		if accountOf(a) != nil {
			ordered = append(ordered, a)
		}
	}
	for _, a := range fighters {
		if accountOf(a) == nil {
			ordered = append(ordered, a)
		}
	}
	return ordered
}

// ResolveAttack is one round of attacker's attacks on defender.
func (Ruleset) ResolveAttack(attacker, defender *actors.Actor) rules.AttackResult {
	return calculateCombat(*attacker.Character, *defender.Character)
}

// WaitLines are the lines for a round the attacker spends waiting.
func (Ruleset) WaitLines(attacker, defender *actors.Actor) []perception.Message {
	return GetWaitMessages(items.Wait, attacker.Character, defender.Character)
}

// HatredRounds is two minutes of rounds less the attacker's Perception.
func (Ruleset) HatredRounds(attacker *actors.Actor) int {
	return configs.GetConfig().Timing.MinutesToRounds(2) - attacker.Character.Stats.Perception.ValueAdj
}

// FindsWeapon rolls d100 under the NPC's Perception.
func (Ruleset) FindsWeapon(npc *actors.Actor) bool {
	roll := util.Rand(100)

	util.LogRoll(`Look for weapon`, roll, npc.Character.Stats.Perception.ValueAdj)

	return roll < npc.Character.Stats.Perception.ValueAdj
}

// FleeEscapes: 30% flat, plus up to 70% from Speed against the blocker's.
func (Ruleset) FleeEscapes(fleer, blocker *actors.Actor) bool {
	// Stat comparison accounts for up to 70% of chance to flee.
	chanceIn100 := int(float64(fleer.Character.Stats.Speed.ValueAdj) / (float64(fleer.Character.Stats.Speed.ValueAdj) + float64(blocker.Character.Stats.Speed.ValueAdj)) * 70)
	chanceIn100 += 30

	roll := util.Rand(100)

	util.LogRoll(`Flee`, roll, chanceIn100)

	return roll < chanceIn100
}

// Consider is CombatOdds: the defender's rounds to kill over the attacker's.
func (Ruleset) Consider(attacker, defender *actors.Actor) float64 {
	return CombatOdds(*attacker.Character, *defender.Character)
}

// CanAttack allows any NPC target. An account target needs PvP allowed in
// the attacker's room and, when it differs, the defender's room (spec
// ruling 1: no shooting out of or into a safe room).
func (Ruleset) CanAttack(attacker, defender *actors.Actor) error {
	if accountOf(defender) == nil {
		return nil
	}
	// CanPvp checks only the attacker's room against the recovery room; a
	// safe room protects whoever stands in it.
	if defender.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) {
		return errors.New(`Fighting is not allowed here.`)
	}
	r := rooms.LoadRoom(attacker.Character.RoomId)
	if r == nil {
		return errors.New(`Fighting is not allowed here.`)
	}
	if err := r.CanPvp(attacker.Character, defender.Character); err != nil {
		return err
	}
	if defender.Character.RoomId != attacker.Character.RoomId {
		dr := rooms.LoadRoom(defender.Character.RoomId)
		if dr == nil {
			return errors.New(`Fighting is not allowed here.`)
		}
		if err := dr.CanPvp(attacker.Character, defender.Character); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 2: `combat.go`:** delete `Attack` with its comment (:22-50) and `CreditAccount` with its
  comment (:60-65); keep `accountOf`. `git rm internal/rules/dikubasic/attackresult.go`. In
  `calculateCombat` change the return type (:182) and `attackResult := AttackResult{}` (:184) to
  `rules.AttackResult`; add the `rules` import.

- [ ] **Step 3: `NewRound_DoCombat.go`.** Replace the `dikubasic` import with
  `github.com/GoMudEngine/GoMud/internal/rules`. Then:

  (a) :33 becomes `for _, a := range rules.Get().TurnOrder(fighters()) {` and :45-62
  (`fightersWithAccountsFirst` with its comment) becomes:

```go
// fighters lists every actor with a body that may take a combat turn: those
// with an account, then NPCs, each from its own ID source (map order). The
// ruleset puts them in turn order.
func fighters() []*actors.Actor {
	out := []*actors.Actor{}
	for _, userId := range accounts.GetOnlineUserIds() {
		if a := actors.ByUserId(userId); a != nil {
			out = append(out, a)
		}
	}
	for _, mobId := range mobs.GetAllMobIds() {
		if a := actors.Get(mobId); a != nil {
			out = append(out, a)
		}
	}
	return out
}
```

  (b) After `accountOf` (:64-70) add:

```go
// attack runs one round of a's attacks on d: the ruleset works it out, the
// engine applies the damage, credits a's account on every attack, hit or
// miss, and runs the defender's wimpy check when it took damage.
func attack(a, d *actors.Actor) rules.AttackResult {
	result := rules.Get().ResolveAttack(a, d)

	d.Character.ApplyHealthChange(result.DamageToTarget * -1)

	// Remember who has hit him
	if acct := accountOf(a); acct != nil {
		d.Character.TrackPlayerDamage(acct.UserId, result.DamageToTarget)
	}

	if result.DamageToTarget != 0 {
		if acct := accountOf(d); acct != nil {
			acct.WimpyCheck()
		}
	}

	return result
}
```

  (c) The call `lookForWeapon(mob)` at :159 (inside the `if mob := mobs.GetInstance(a.Id); mob != nil {`
  block, which stays) becomes `lookForWeapon(a, mob)`. In `lookForWeapon` the signature becomes
  `func lookForWeapon(a *actors.Actor, mob *mobs.Mob)` and :305-309 (from `roll := util.Rand(100)`
  through `if roll < mob.Character.Stats.Perception.ValueAdj {`) becomes
  `if rules.Get().FindsWeapon(a) {`. The weapon pick `util.Rand(len(possibleWeapons))` stays.

  (d) :167 `dikubasic.GetWaitMessages(items.Wait, a.Character, d.Character)` becomes
  `rules.Get().WaitLines(a, d)`.

  (e) :182 `dikubasic.Attack(a, d)` becomes `attack(a, d)`.

  (f) :211-220 (the group hatred block) becomes:

```go
	// Group hatred follows kill credit (plan Decision 8): the attacker's
	// account. How long is the ruleset's.
	if defMob != nil {
		if acct := accountOf(a); acct != nil {
			hatred := rules.Get().HatredRounds(a)
			for _, groupName := range defMob.Groups {
				mobs.MakeHostile(groupName, acct.UserId, hatred)
			}
		}
	}
```

  (g) In `flee`, :406-417 (from `// Stat comparison accounts` through the closing brace of
  `if roll >= chanceIn100 { ... }`) becomes:

```go
		if !rules.Get().FleeEscapes(a, b) {
			blocker = b
			break
		}
```

  `offhandBreakCheck` is not touched (spec ruling 5). Remove only the imports the compiler reports
  unused.

- [ ] **Step 4: `consider.go`** :39-53 (from `var odds float64` through the closing brace of the
  `if playerId > 0 { ... } else { ... }`) becomes:

```go
	var target *actors.Actor
	considerType := "mob"
	if playerId > 0 {
		target = actors.ByUserId(playerId)
		considerType = "user"
	} else {
		target = actors.Get(mobId)
	}

	if target == nil || target.Character == nil {
		Tell(actor, `<ansi fg="mobname">`+lookAt+`</ansi> can't be found.`)
		return true, nil
	}

	odds := rules.Get().Consider(actor, target)
	considerName := target.Character.Name
```

  Replace the `dikubasic` import with `rules`; drop `accounts` and `mobs` if the compiler says unused.

- [ ] **Step 5: `attack.go`.** NPC target: :188-190 (`m := mobs.GetInstance(attackMobId)`, blank
  line, `if m != nil {`) becomes:

```go
		m := mobs.GetInstance(attackMobId)
		d := actors.Get(attackMobId)

		if m != nil && d != nil {
			if err := rules.Get().CanAttack(actor, d); err != nil {
				Tell(actor, err.Error())
				return true, nil
			}
```

  Player target: :232-237 (`if p := accounts.GetByUserId(attackPlayerId); p != nil {` through the
  closing brace of the `CanPvp` check) becomes:

```go
		p := accounts.GetByUserId(attackPlayerId)
		d := actors.ByUserId(attackPlayerId)
		if p != nil && d != nil {

			if err := rules.Get().CanAttack(actor, d); err != nil {
				Tell(actor, err.Error())
				return true, nil
			}
```

  The rest of the block is unchanged. Add the `rules` import. (A name match with no actor is skipped
  silently: spec, decisions table.)

- [ ] **Step 6: Boot and test mains.** `main.go`, immediately before `hooks.RegisterListeners()` (:206):

```go
	// The reference ruleset. This is the one place the engine names a ruleset.
	rules.Set(dikubasic.New())
```

  with imports `github.com/GoMudEngine/GoMud/internal/rules` and
  `github.com/GoMudEngine/GoMud/internal/rules/dikubasic`. In `TestMain` of
  `internal/hooks/actionpoints_test.go` and `internal/usercommands/usercommands_test.go` add
  `rules.Set(dikubasic.New())` after `messaging.Init()`, with the same two imports in each file. If another package's tests panic with "no
  ruleset is set", add the same line to its `TestMain` (create one modelled on
  `usercommands_test.go:25` if it has none) and list it in the report.

- [ ] **Step 7: Guard controls** in `internal/guards/combat_test.go`. In `TestOneCombatPath` replace
  the comment above it and the positive-control block (`// Positive control:` through the end of the
  function) with:

```go
// TestOneCombatPath enforces the phase 4b slice 5 ruling, as the Phase 5 seam
// keeps it: one attack (the engine's attack, calling the ruleset's
// ResolveAttack) and one round (DoCombat). The per-kind attack variants, the
// two per-kind round handlers and the room-message fields are gone and stay
// gone.
```

```go
	// Positive control: the one attack, the ruleset's resolution and the one
	// round exist, so the walk reached the files the ban is about.
	foundAttack, foundResolve, foundRound := false, false, false
	for path, f := range goFiles(t) {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if path == "internal/hooks/NewRound_DoCombat.go" && fd.Name.Name == "attack" && fd.Recv == nil {
				foundAttack = true
			}
			if path == "internal/rules/dikubasic/ruleset.go" && fd.Name.Name == "ResolveAttack" && fd.Recv != nil {
				foundResolve = true
			}
			if path == "internal/hooks/NewRound_DoCombat.go" && fd.Name.Name == "DoCombat" {
				foundRound = true
			}
		}
	}
	if !foundAttack {
		t.Error("control: func attack not found in internal/hooks/NewRound_DoCombat.go; the guard may be scanning the wrong place")
	}
	if !foundResolve {
		t.Error("control: method ResolveAttack not found in internal/rules/dikubasic/ruleset.go; the guard may be scanning the wrong place")
	}
	if !foundRound {
		t.Error("control: func DoCombat not found in internal/hooks/NewRound_DoCombat.go; the guard may be scanning the wrong place")
	}
}
```

  In the same test's ban loop, the error message `"%s: uses %s; combat has one Attack and one round (slice 5)"`
  becomes `"%s: uses %s; combat has one attack and one round (slice 5, kept by the Phase 5 seam)"`.
  Rename `TestCombatPackageDoesNotSend` to `TestRulesetDoesNotSend`, with the comment:
  `// TestRulesetDoesNotSend enforces: internal/rules returns messages and the` /
  `// engine sends them. Nothing under it sends text or builds an events.Message.`
  In its body, the four messages `"...; combat returns messages, the round sends them"` become
  `"...; a ruleset returns messages, the engine sends them"`, and the comment
  `// Positive control: the combat package was parsed.` becomes
  `// Positive control: internal/rules was parsed.` (Task 2 already changed its scan prefix and the
  `scanned` message.)

- [ ] **Step 7b: Unit tests for the moved methods** (no golden reaches `FindsWeapon`, and goldens do
  not print hatred length). Create `internal/rules/dikubasic/ruleset_test.go`:

```go
package dikubasic

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init() // NPC instances get actors through its hooks
	os.Exit(m.Run())
}

// npc is a test NPC actor with the given Perception and Speed.
func npc(t *testing.T, perception, speed int) *actors.Actor {
	t.Helper()
	m := mobs.NewTestInstance(mobs.Mob{})
	id := m.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	a := actors.Get(id)
	require.NotNil(t, a, "actors.Init wires NPC instances to actors")
	a.Character.Stats.Perception.ValueAdj = perception
	a.Character.Stats.Speed.ValueAdj = speed
	return a
}

// player is a test actor with an account.
func player(t *testing.T, userId int) *actors.Actor {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	u := &accounts.Account{UserId: userId, Character: c, Password: "$2a$test"}
	a := actors.Spawn(c, actors.NewConnectionController(u))
	t.Cleanup(func() { actors.Remove(a.Id) })
	return a
}

func TestTurnOrderPutsAccountsFirstKeepingOrder(t *testing.T) {
	n1, p1, n2, p2 := npc(t, 0, 0), player(t, 9601), npc(t, 0, 0), player(t, 9602)
	got := New().TurnOrder([]*actors.Actor{n1, p1, n2, p2})
	assert.Equal(t, []*actors.Actor{p1, p2, n1, n2}, got)
}

// matchesOneRoll checks, over many seeds, that roll() answers exactly
// util.Rand(100) < chance from the same seed and makes exactly one draw (the
// next draw after it is the second draw of the stream).
func matchesOneRoll(t *testing.T, chance int, roll func() bool) {
	t.Helper()
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) })
	for seed := int64(1); seed <= 1000; seed++ {
		util.SeedRandForTest(seed)
		want := util.Rand(100) < chance
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		require.Equal(t, want, roll(), "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: exactly one draw", seed)
	}
}

func TestFindsWeaponRollsUnderPerception(t *testing.T) {
	a := npc(t, 50, 0)
	matchesOneRoll(t, 50, func() bool { return New().FindsWeapon(a) })
}

func TestFleeEscapesOnSpeedShareOf70Plus30(t *testing.T) {
	// Unequal Speed, so swapping fleer and blocker changes the chance:
	// int(10/(10+30)*70) = 17, plus 30.
	fleer, blocker := npc(t, 0, 10), npc(t, 0, 30)
	matchesOneRoll(t, 47, func() bool { return New().FleeEscapes(fleer, blocker) })
}

func TestHatredRoundsIsTwoMinutesLessPerception(t *testing.T) {
	a := player(t, 9603)
	a.Character.Stats.Perception.ValueAdj = 7
	assert.Equal(t, configs.GetConfig().Timing.MinutesToRounds(2)-7, New().HatredRounds(a))
}

func TestCanAttackAllowsAnyNPCTarget(t *testing.T) {
	assert.NoError(t, New().CanAttack(player(t, 9604), npc(t, 0, 0)))
}

// A safe room protects whoever stands in it, and that check comes before
// the PvP settings. The attacker stands in a loaded room and the recovery
// room is a loaded, non-zero room (room 0 aliases to the start room; new
// bodies start in unloaded room -1). PvP is disabled, so a missing check, or
// one moved after CanPvp, reads "PVP is disabled." instead.
func TestCanAttackRefusesATargetInTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94001, 94002
	// RestoreOverrides lays keys over the live config and resets nothing
	// else, so the cleanup names each key's old value (as setConfig does).
	old := map[string]any{
		"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		"GamePlay.PVP.Enabled":           string(configs.GetPVPConfig().Enabled),
		"GamePlay.PVP.MinimumLevel":      int(configs.GetPVPConfig().MinimumLevel),
	}
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["SpecialRooms.DeathRecoveryRoom"] = recovery
	next["GamePlay.PVP.Enabled"] = "disabled"
	next["GamePlay.PVP.MinimumLevel"] = 0
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		for k, v := range old {
			back[k] = v
		}
		require.NoError(t, configs.RestoreOverrides(back))
		require.Equal(t, old["SpecialRooms.DeathRecoveryRoom"], int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "config restored")
	})
	require.Equal(t, recovery, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom))

	for _, id := range []int{here, recovery} {
		r := rooms.NewEmptyRoom()
		r.RoomId, r.Zone = id, "rulesetzone"
		rooms.AddTestRoom(r)
		id := id
		t.Cleanup(func() { rooms.RemoveTestRoom(id) })
	}
	attacker, target := player(t, 9607), player(t, 9608)
	attacker.Character.RoomId, target.Character.RoomId = here, recovery
	assert.EqualError(t, New().CanAttack(attacker, target), "Fighting is not allowed here.")
}
```

  Add `time`, `github.com/GoMudEngine/GoMud/internal/rooms` and
  `github.com/GoMudEngine/GoMud/internal/util` to the imports. Run
  `go test ./internal/rules/dikubasic -count=1`; all pass. Show each test can fail, one mutation at
  a time (Edit, run, see the matching test fail, restore): `HatredRounds` `-` to `+`;
  `FindsWeapon` `<` to `<=`; the `30` in `FleeEscapes` to `29`; an extra `util.Rand(100)` before the
  roll in `FindsWeapon`; swap the two loops in `TurnOrder`; delete the recovery-room check in
  `CanAttack`; move that check after the attacker-room `CanPvp`; in `FleeEscapes` use the blocker's
  Speed in the numerator. Paste the eight failure lines.

- [ ] **Step 7c: Command tests: `consider`, and `attack`'s PvP refusal.** Create
  `internal/usercommands/pvp_test.go` (shared helpers; Task 6 uses them too):

```go
package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setPvP sets GamePlay.PVP.Enabled for one test, with no minimum level (the
// shipped config has 15). AddOverlayOverrides will not replace a key that is
// already set, and RestoreOverrides resets nothing it is not given, so this
// names each key's old value on the way back, as setConfig in
// internal/hooks does.
func setPvP(t *testing.T, setting string) {
	t.Helper()
	oldEnabled := string(configs.GetPVPConfig().Enabled)
	oldMin := int(configs.GetPVPConfig().MinimumLevel)
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["GamePlay.PVP.Enabled"] = setting
	next["GamePlay.PVP.MinimumLevel"] = 0
	require.NoError(t, configs.RestoreOverrides(next))
	require.Equal(t, setting, string(configs.GetPVPConfig().Enabled), "the override took")
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["GamePlay.PVP.Enabled"] = oldEnabled
		back["GamePlay.PVP.MinimumLevel"] = oldMin
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

// addFoundPlayer is addPlayer, registered so FindByName and GetByUserId see it.
func addFoundPlayer(t *testing.T, userId int, name string, room *rooms.Room) *actors.Actor {
	t.Helper()
	a := addPlayer(t, userId, name, room)
	accounts.SetTestUser(a.Controller().Account())
	t.Cleanup(func() { accounts.RemoveTestUser(userId) })
	return a
}

func TestAttackAPlayerRefusedWhenPvPDisabled(t *testing.T) {
	_, player, room := goSetup(t)
	addFoundPlayer(t, 9203, "Target", room)
	setPvP(t, configs.PVPDisabled)

	told := playerTexts(t, player, func() {
		_, err := Attack("Target", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"PVP is disabled.\n"}, told)
	assert.Nil(t, player.Character.Aggro)
}

// PvP settings do not touch fighting an NPC. This pins CanAttack's argument
// order: swapped, the player becomes the defender with an account and PvP
// rules would refuse the rat.
func TestAttackAnNPCAllowedWhenPvPDisabled(t *testing.T) {
	_, player, room := goSetup(t) // the NPC is "Rat"
	setPvP(t, configs.PVPDisabled)

	told := playerTexts(t, player, func() {
		_, err := Attack("Rat", player, room, 0)
		require.NoError(t, err)
	})
	require.NotEmpty(t, told)
	assert.Contains(t, told[0], "You prepare to fight")
	assert.NotNil(t, player.Character.Aggro)
}
```

  Run `go test ./internal/usercommands -run 'PvPDisabled' -count=1 -v`; both pass. One probe at a
  time (Edit, run, see it fail, restore), pasting each failure: delete the `CanAttack` call in
  `attack.go`'s player branch (the first test fails); swap the arguments of the `CanAttack` call in
  `attack.go`'s NPC branch to `(d, actor)` (the second test fails).

- [ ] **Step 7d: The engine's `attack` keeps the wimpy check.** Create
  `internal/hooks/attack_engine_test.go`:

```go
package hooks

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/stretchr/testify/require"
)

// The defender's wimpy check runs after damage: with wimpy above any health
// percentage, Alice queues a flee once the guard hurts her.
func TestAttackRunsTheDefendersWimpyCheck(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	alice.SetConfigOption(`wimpy`, 101)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)

	require.Less(t, alice.Character.Health, 12, "the fixture must damage Alice")
	require.Contains(t, w.out.String(), `Input Alice: "flee"`)
}

// harmless is the active ruleset where every attack connects for 0.
type harmless struct{ rules.Ruleset }

func (harmless) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
	return rules.AttackResult{Hit: true}
}

// The wimpy check runs only on damage: a 0-damage round queues no flee.
func TestAttackSkipsTheWimpyCheckWithoutDamage(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	alice.SetConfigOption(`wimpy`, 101)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(harmless{rules.Get()}))
	w.start()
	w.rounds(3)

	require.NotContains(t, w.out.String(), `"flee"`)
}
```

  Add `github.com/GoMudEngine/GoMud/internal/actors` and `github.com/GoMudEngine/GoMud/internal/rules`
  to the imports. Changing `if result.DamageToTarget != 0 {` in `attack` to `if true {` makes the
  second test fail (restore after); paste it.

  Before relying on the last line, check how `Account.Command(cmd, -1)` (`userrecord.go:620`) reaches
  the record (`grep -n 'func (u \*Account) Command' -A 15 internal/accounts/userrecord.go`); if a
  negative wait sends it some other way than an `events.Input` the recorder sees, assert on what does
  appear and say so. Run it; it passes. Delete the `acct.WimpyCheck()` call in `attack` (Edit), run,
  see it fail, restore. Paste the failure.

- [ ] **Step 8: Prove the engine no longer names the ruleset:**
  `grep -rln --exclude-dir=.claude --exclude-dir=.git 'rules/dikubasic"' --include=*.go . | grep -v _test.go | grep -v '^./internal/rules/' | sort`
  prints exactly `./internal/usercommands/suicide_npc.go`, `./internal/web/api_v1_items.go`,
  `./internal/web/api_v1_mobs.go`, `./main.go`.

- [ ] **Step 9:** Per-task check. **All goldens unchanged** (testdata `git status` empty). Controller
  note: three rewired lines are pinned only by Task 5's stub tests, not by anything in this commit:
  `HatredRounds(a)`, the `MakeHostile` rounds argument, and `range rules.Get().TurnOrder(fighters())`.
  Read those three by eye in this task's diff. If one
  changes, stop: report the first differing lines
  (`go test ./internal/hooks -run 'CombatGolden|TestReads' -count=1 -v 2>&1 | head -80`); do not
  re-record.
- [ ] **Step 10:** Smoke: `go run ./cmd/smoke` and `go run ./cmd/smoke -bots 5`; paste the summary lines.
- [ ] **Step 11:** Commit every file above, including the three new test files (not
  `attackresult.go`: `git rm` already staged its removal, and naming it in `git add` fails):
  `Phase 5.2: dikubasic.Ruleset; DoCombat, consider and attack call rules.Get() (goldens unchanged)`.

### Task 4: Import guard (Sonnet)

**Files:** modify `internal/guards/combat_test.go` (append).

- [ ] **Step 1: Write the guard:**

```go
// TestOnlyBootNamesTheRuleset enforces the Phase 5 seam: the engine reaches
// the rules through internal/rules (rules.Get()) and never imports the
// reference ruleset; main.go picks it at boot. Each other importer is listed
// with the slice that removes it. An entry that no longer imports the ruleset
// fails, so the list cannot go stale, and main.go's entry is the positive
// control that the walk reached the files.
func TestOnlyBootNamesTheRuleset(t *testing.T) {
	const ruleset = "github.com/GoMudEngine/GoMud/internal/rules/dikubasic"
	allowed := map[string]string{
		"main.go":                              "boot: rules.Set(dikubasic.New())",
		"internal/usercommands/suicide_npc.go": "AlignmentChange; slice 3 moves death behind the ruleset",
		"internal/web/api_v1_items.go":         "weapon and armour rankings; slice 4 shrink",
		"internal/web/api_v1_mobs.go":          "mob rankings; slice 4 shrink",
	}
	seen := map[string]bool{}
	for path, f := range goFiles(t) {
		if strings.HasPrefix(path, "internal/rules/dikubasic/") {
			continue
		}
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, `"`) != ruleset {
				continue
			}
			if _, ok := allowed[path]; ok {
				seen[path] = true
				continue
			}
			t.Errorf("%s imports %s; the engine reaches the rules through internal/rules (rules.Get())", path, ruleset)
		}
	}
	for path, why := range allowed {
		if !seen[path] {
			t.Errorf("stale allowlist: %s (%s) no longer imports %s; remove it from the list", path, why, ruleset)
		}
	}
}
```

- [ ] **Step 2: Run; it passes.** `go test ./internal/guards -run TestOnlyBootNamesTheRuleset -count=1`.
- [ ] **Step 3: Show it can fail, twice.** (i) Add
  `_ "github.com/GoMudEngine/GoMud/internal/rules/dikubasic"` to the imports of `internal/hooks/hooks.go`
  (Edit); run; paste the failure line; remove it (Edit); rerun; `ok`. (ii) Add the allowlist entry
  `"internal/hooks/hooks.go": "probe",`; run; paste the stale-allowlist failure; remove it; rerun; `ok`.
  `git diff --stat` then shows only `internal/guards/combat_test.go`.
- [ ] **Step 4:** Per-task check. Commit `internal/guards/combat_test.go`:
  `Phase 5.2: guard: only main.go (and three listed stragglers) import dikubasic`.

### Task 5: Stub rulesets change the fight with no core edit (Sonnet)

**Files:** create `internal/hooks/ruleset_stub_test.go`, `internal/usercommands/ruleset_stub_test.go`.
No non-test file changes in this task.

- [ ] **Step 1:** (none; the turn-order test asserts on the order of `ResolveAttack` calls, not on
  swing text, which varies with the seed.)

- [ ] **Step 2: `internal/hooks/ruleset_stub_test.go`:**

```go
package hooks

import (
	"fmt"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/stretchr/testify/require"
)

// oneShot is the active ruleset with one rule changed: every attack hits for
// 1000.
type oneShot struct{ rules.Ruleset }

func (oneShot) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
	return rules.AttackResult{Hit: true, DamageToTarget: 1000, Messages: []perception.Message{{
		Kind: perception.WorldEvent, Actor: a.Id, Target: d.Id,
		Senses: perception.Sight, Range: perception.Room, ActorText: "STUB HIT",
	}}}
}

// The pvm_death fixture under a ruleset where every hit kills: the rat dies
// in round 1 instead of round 4, with no change to the engine.
func TestStubRulesetChangesTheFight(t *testing.T) {
	w := newCombatWorld(t, 111)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "stub-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(oneShot{rules.Get()}))
	w.start()
	w.rounds(20)

	out := w.out.String()
	require.Contains(t, out, "== round 1")
	require.NotContains(t, out, "== round 2", "the reference ruleset takes 4 rounds (golden pvm_death)")
	require.Contains(t, out, `Input rat: "suicide"`)
	require.Contains(t, out, "STUB HIT")
}

// reversed is the active ruleset with the turn order reversed (NPCs
// first). It also records who resolves an attack, in order, and otherwise
// delegates, so the random draws are the reference ruleset's.
type reversed struct {
	rules.Ruleset
	attackers *[]ids.ID
}

func (r reversed) TurnOrder(f []*actors.Actor) []*actors.Actor {
	in := r.Ruleset.TurnOrder(f)
	out := make([]*actors.Actor, 0, len(in))
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return out
}

func (r reversed) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
	*r.attackers = append(*r.attackers, a.Id)
	return r.Ruleset.ResolveAttack(a, d)
}

// Turn order is the ruleset's: reversed, the guard attacks before Alice in
// round 1 (the reference ruleset has Alice first).
func TestStubRulesetChangesTurnOrder(t *testing.T) {
	w := newCombatWorld(t, 202)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 50)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 50)
	alice.Character.SetAggro(guard.Character.Id, characters.DefaultAttack)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	var attackers []ids.ID
	t.Cleanup(rules.SetForTest(reversed{Ruleset: rules.Get(), attackers: &attackers}))
	w.start()
	w.rounds(1)

	require.Len(t, attackers, 2)
	require.Equal(t, []ids.ID{guard.Character.Id, alice.Character.Id}, attackers)
}

// armed is the active ruleset with the weapon search forced to an answer.
// It records whom it is asked about.
type armed struct {
	rules.Ruleset
	finds bool
	asked *[]ids.ID
}

func (r armed) FindsWeapon(npc *actors.Actor) bool {
	*r.asked = append(*r.asked, npc.Id)
	return r.finds
}

// The weapon search is the ruleset's: an unarmed guard carrying a weapon
// queues "equip" when the ruleset says it finds one, and not otherwise. Each
// case gives the guard the Perception that makes the reference ruleset
// answer the opposite, so each fails if the engine ignores the stub.
func TestStubRulesetDecidesTheWeaponSearch(t *testing.T) {
	for _, finds := range []bool{true, false} {
		t.Run(fmt.Sprint(finds), func(t *testing.T) {
			w := newCombatWorld(t, 303)
			alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 50)
			guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 50)
			guard.Character.Items = append(guard.Character.Items, weapon(93106, "testclub", items.Bludgeoning, 1, 4, 0))
			if finds {
				guard.Character.Stats.Perception.ValueAdj = 0 // the reference never finds
			} else {
				guard.Character.Stats.Perception.ValueAdj = 100 // the reference always finds
			}
			guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
			var asked []ids.ID
			t.Cleanup(rules.SetForTest(armed{Ruleset: rules.Get(), finds: finds, asked: &asked}))
			w.start()
			w.rounds(1)

			if finds {
				require.Contains(t, w.out.String(), `Input guard: "equip testclub"`)
			} else {
				require.NotContains(t, w.out.String(), `"equip `)
			}
			require.Equal(t, []ids.ID{guard.Character.Id}, asked, "asked about the NPC, not its target")
		})
	}
}
```

  Before relying on the `equip` line: queued commands are recorded as `Input <name>: "<text>"` (the
  `pvm_death` golden shows `Input rat: "suicide"`); check what `DisplayName()` returns for the
  `weapon` helper's item. If the record shows it differently, assert on what it shows and say so.
  Nothing between `addNPC` and the weapon search recomputes `ValueAdj` (no `Validate` in the round
  before it); confirm the reference answers by running each case once without the `SetForTest` line
  and without the final `asked` require (the reference queues `equip` in the `false` case and not in
  the `true` case).

  Then add, in the same file:

```go
// hateful is the active ruleset with group hatred lasting 2 rounds. It
// records whom it is asked about.
type hateful struct {
	rules.Ruleset
	asked *[]ids.ID
}

func (h hateful) HatredRounds(a *actors.Actor) int {
	*h.asked = append(*h.asked, a.Id)
	return 2
}

// Hatred length is the ruleset's, asked about the attacker: the rat's group
// hates Alice for exactly two passes of ReduceHostility.
func TestStubRulesetSetsHatredLength(t *testing.T) {
	w := newCombatWorld(t, 111)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 500, "hate-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	var asked []ids.ID
	t.Cleanup(rules.SetForTest(hateful{Ruleset: rules.Get(), asked: &asked}))
	w.start()
	w.rounds(1)

	require.Equal(t, []ids.ID{alice.Character.Id}, asked)
	group := rat.Groups[0] // addNPC suffixes the name per run
	require.True(t, mobs.IsHostile(group, alice.UserId))
	mobs.ReduceHostility()
	require.True(t, mobs.IsHostile(group, alice.UserId), "1 round left")
	mobs.ReduceHostility()
	require.False(t, mobs.IsHostile(group, alice.UserId), "2 rounds are up")
}
```

  Add `github.com/GoMudEngine/GoMud/internal/mobs` to the imports. The rat has 500 health so round 1
  cannot kill it.

- [ ] **Step 3: `internal/usercommands/ruleset_stub_test.go`:**

```go
package usercommands

import (
	"errors"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pacifist is the active ruleset refusing every fight.
type pacifist struct{ rules.Ruleset }

func (pacifist) CanAttack(a, d *actors.Actor) error { return errors.New("STUB REFUSES") }

// attack asks the ruleset before any fight, NPC targets included.
func TestStubRulesetRefusesAttack(t *testing.T) {
	_, player, room := goSetup(t) // the NPC is "Rat"
	t.Cleanup(rules.SetForTest(pacifist{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := Attack("Rat", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"STUB REFUSES\n"}, told)
	assert.Nil(t, player.Character.Aggro)
}

// optimist is the active ruleset rating a fight 5 to 1 when the attacker is
// the one with an account, and 0.1 the other way round, so swapped
// arguments read another label.
type optimist struct{ rules.Ruleset }

func (optimist) Consider(a, d *actors.Actor) float64 {
	if a.Controller() != nil && a.Controller().Account() != nil {
		return 5
	}
	return 0.1
}

// consider reads the ruleset's odds, asked with the player as attacker.
func TestStubRulesetDecidesConsider(t *testing.T) {
	_, player, room := goSetup(t) // the NPC is "Rat"
	t.Cleanup(rules.SetForTest(optimist{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := Consider("Rat", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, told, 2)
	assert.Contains(t, told[1], "Very Favorable")
}

// Considering a player names them in the player colour.
func TestConsiderAPlayerUsesThePlayerColour(t *testing.T) {
	_, player, room := goSetup(t)
	addFoundPlayer(t, 9203, "Target", room)
	t.Cleanup(rules.SetForTest(optimist{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := Consider("Target", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, told, 2)
	assert.Contains(t, told[0], `<ansi fg="username">Target</ansi>`)
}
```

  (`addFoundPlayer` is in `pvp_test.go`, Task 3.) Probe: delete `considerType = "user"` in
  `consider.go`; this test fails.

- [ ] **Step 4:** Run: `go test ./internal/hooks ./internal/usercommands -run Stub -count=1 -v`; all pass.
- [ ] **Step 5: Prove they test the seam**, one probe at a time (Edit, run, see it fail, restore,
  rerun, pass), pasting each failure line:
  - one-shot, pacifist, hatred, consider: remove the test's `t.Cleanup(rules.SetForTest(...))` line.
    (Consider with the reference panics on a nil race in `GetDefaultDiceRoll`: race data is not
    loaded in `usercommands` tests. That panic is the expected failure.) Also for consider: swap the
    arguments in `consider.go`'s `rules.Get().Consider(actor, target)`; the test fails on the label.
  - weapon search: in `lookForWeapon` change `if rules.Get().FindsWeapon(a) {` to
    `if _ = rules.Get().FindsWeapon(a); true {` (the `false` subtest fails on `equip`), then to
    `...; false {` (the `true` subtest fails). (Removing `SetForTest` fails only on the `asked`
    recorder, which shows the stub was installed, not that the engine obeys it.) Also change
    `lookForWeapon(a, mob)` to `lookForWeapon(d, mob)`: the `asked` assert fails.
  - turn order: in `DoCombat` change `range rules.Get().TurnOrder(fighters())` to `range fighters()`;
    the test fails with the two IDs swapped. (Removing `SetForTest` here only shows the recorder is
    gone.)
  - hatred: in `NewRound_DoCombat.go` change `HatredRounds(a)` to `HatredRounds(d)`, then the
    `MakeHostile` rounds argument to `hatred*0`; each fails. (The hatred test's `SetForTest`
    removal fails on the `asked` recorder; `hatred*0` is the probe that tests the length.)
- [ ] **Step 6:** Per-task check. Commit the two test files:
  `Phase 5.2: stub rulesets change the fight, turn order, weapon search, hatred, consider and attack check with no core edit`.
  `git show --stat HEAD` lists only the two `_test.go` files.

### Task 6: `shoot` through `CanAttack`, with `AggroChanged` (Sonnet)

Player-visible (spec ruling 1): shooting a player obeys the PvP settings in both rooms.

**Files:** modify `internal/usercommands/shoot.go`; create `internal/usercommands/shoot_test.go`.

- [ ] **Step 1: Write the failing tests** `internal/usercommands/shoot_test.go`:

```go
package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setPvP and addFoundPlayer are in pvp_test.go (Task 3).

// giveBow puts a shooting weapon in a's hands.
func giveBow(a *actors.Actor) {
	spec := items.ItemSpec{ItemId: 92101, Name: "testbow", NameSimple: "testbow",
		Type: items.Weapon, Subtype: items.Shooting, Hands: items.OneHanded,
		Damage: items.Damage{Attacks: 1, DiceCount: 1, SideCount: 6}}
	a.Character.Equipment.Weapon = items.Item{ItemId: 92101, Id: ids.New(ids.NoCell), Spec: &spec}
}

// shootSetup is goSetup with a bow on the player and a second player,
// "Target", in the room north.
func shootSetup(t *testing.T) (shooter, target *actors.Actor, here, north *rooms.Room) {
	t.Helper()
	_, shooter, here = goSetup(t)
	giveBow(shooter)
	north = rooms.LoadRoom(testNextRoomId)
	target = addFoundPlayer(t, 9203, "Target", north)
	return shooter, target, here, north
}

func TestShootAtAPlayerRefusedWhenPvPDisabled(t *testing.T) {
	shooter, _, here, _ := shootSetup(t)
	setPvP(t, configs.PVPDisabled)

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"PVP is disabled.\n"}, told)
	assert.Nil(t, shooter.Character.Aggro, "a refused shot starts no fight")
}

// Spec ruling 1: with PvP limited, both rooms must be PvP rooms.
func TestShootIntoASafeRoomIsRefused(t *testing.T) {
	shooter, _, here, north := shootSetup(t)
	setPvP(t, configs.PVPLimited)
	here.Pvp, north.Pvp = true, false
	t.Cleanup(func() { here.Pvp, north.Pvp = false, false })

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"This is not a PVP area.\n"}, told)
	assert.Nil(t, shooter.Character.Aggro)
}

// Spec ruling 1, the other half: no shooting out of a safe room either.
func TestShootOutOfASafeRoomIsRefused(t *testing.T) {
	shooter, _, here, north := shootSetup(t)
	setPvP(t, configs.PVPLimited)
	here.Pvp, north.Pvp = false, true
	t.Cleanup(func() { here.Pvp, north.Pvp = false, false })

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"This is not a PVP area.\n"}, told)
	assert.Nil(t, shooter.Character.Aggro)
}

func TestShootAtAPlayerQueuesAggroChanged(t *testing.T) {
	shooter, target, here, _ := shootSetup(t)
	setPvP(t, configs.PVPEnabled)

	got := capture[events.AggroChanged](t, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	require.Len(t, got, 1)
	assert.Equal(t, shooter.Id, got[0].ActorId)
	assert.Equal(t, here.RoomId, got[0].RoomId, "GMCP picks the room to update from this")
	require.NotNil(t, shooter.Character.Aggro)
	assert.Equal(t, target.Id, shooter.Character.Aggro.Target)
	assert.Equal(t, "north", shooter.Character.Aggro.ExitName)
}

// Shooting an NPC queues AggroChanged, and PvP settings do not refuse it
// (disabled here, to pin CanAttack's argument order in shoot's NPC branch).
func TestShootAtAnNPCQueuesAggroChanged(t *testing.T) {
	npc, shooter, here := goSetup(t) // "Rat" starts here: move it north
	giveBow(shooter)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	setPvP(t, configs.PVPDisabled)

	got := capture[events.AggroChanged](t, func() {
		_, err := Shoot("Rat north", shooter, here, 0)
		require.NoError(t, err)
	})
	require.Len(t, got, 1)
	assert.Equal(t, shooter.Id, got[0].ActorId)
	assert.Equal(t, here.RoomId, got[0].RoomId)
}

// An NPC shooting a player obeys the PvP settings too.
func TestNPCShootingAPlayerRefusedWhenPvPDisabled(t *testing.T) {
	npc, _, here := goSetup(t) // "Rat" is in this room
	giveBow(npc)
	addFoundPlayer(t, 9203, "Target", rooms.LoadRoom(testNextRoomId))
	setPvP(t, configs.PVPDisabled)

	_, err := Shoot("Target north", npc, here, 0)
	require.NoError(t, err)
	assert.Nil(t, npc.Character.Aggro, "a refused NPC shot starts no fight")
}

// shoot asks the ruleset before shooting an NPC too. (pacifist is in
// ruleset_stub_test.go, Task 5).
func TestShootAtAnNPCAsksTheRuleset(t *testing.T) {
	npc, shooter, here := goSetup(t)
	giveBow(shooter)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	t.Cleanup(rules.SetForTest(pacifist{rules.Get()}))

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Rat north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"STUB REFUSES\n"}, told)
	assert.Nil(t, shooter.Character.Aggro)
}
```

  Add `github.com/GoMudEngine/GoMud/internal/rules` to this file's imports.

  Check before running: `Room.Pvp` is the field `IsPvp` reads (`rooms.go:2047`); `rooms.MoveToRoom`
  is `(ids.ID, int, ...bool) error` (`roommanager.go:281`). If `MoveToRoom` needs more setup in this
  test world (it may expect a zone or fire a look), move the NPC the way `go_test.go` does and say so.

- [ ] **Step 2: Run; they fail.** `go test ./internal/usercommands -run Shoot -count=1 -v`: the
  three refusal tests fail (aggro set, no refusal line), both `AggroChanged` tests fail (`Len 0`), the NPC
  shooter test fails (aggro set), and the pacifist NPC test fails (aggro set, no refusal). If the PvP-enabled player test fails for any
  other reason, stop and report.
- [ ] **Step 3: Change `shoot.go`.** NPC branch: :93-95 (`m := mobs.GetInstance(attackMobId)`, blank
  line, `if m != nil {`) becomes:

```go
		m := mobs.GetInstance(attackMobId)
		d := actors.Get(attackMobId)

		if m != nil && d != nil {

			if err := rules.Get().CanAttack(actor, d); err != nil {
				Tell(actor, err.Error())
				return true, nil
			}
```

  and directly after `actor.Character.SetAggroRemote(exitName, attackMobId, characters.Shooting)` (:97):

```go
			events.AddToQueue(events.AggroChanged{ActorId: actor.Character.Id, RoomId: actor.Character.RoomId})
```

  Player branch: :109-111 (`p := accounts.GetByUserId(attackPlayerId)` and `if p != nil {`) becomes:

```go
		p := accounts.GetByUserId(attackPlayerId)
		d := actors.ByUserId(attackPlayerId)

		if p != nil && d != nil {

			if err := rules.Get().CanAttack(actor, d); err != nil {
				Tell(actor, err.Error())
				return true, nil
			}
```

  (before the party check, as in `attack`), and directly after
  `actor.Character.SetAggroRemote(exitName, p.Character.Id, characters.Shooting)` (:122) the same
  `events.AddToQueue(events.AggroChanged{...})` line. Add the `rules` import.
- [ ] **Step 4: Run; they pass.** Same command. Then, one probe at a time (Edit, run, see it fail,
  restore), paste each failure: in `dikubasic.CanAttack` delete the defender-room `CanPvp` call
  (`TestShootIntoASafeRoomIsRefused` fails); change `rooms.LoadRoom(attacker.Character.RoomId)` to
  `rooms.LoadRoom(defender.Character.RoomId)` (`TestShootOutOfASafeRoomIsRefused` fails); replace the `SetForTest` line in
  `TestShootAtAnNPCAsksTheRuleset` with `_ = rules.Get` (deleting it outright leaves `rules` unused
  and the file stops compiling) (the test fails); swap the arguments of `shoot.go`'s NPC-branch
  `CanAttack` to `(d, actor)` (`TestShootAtAnNPCQueuesAggroChanged` fails).
- [ ] **Step 5:** Per-task check (`shoot_exit` sets its aggro directly, so goldens stay). Commit
  `internal/usercommands/shoot.go internal/usercommands/shoot_test.go`:
  `Phase 5.2: shoot asks CanAttack (PvP in both rooms) and fires AggroChanged, as attack does`.

### Task 7: Verify, review, docs, PR (controller)

- [ ] Full check from a clean tree: `go build ./...`, `go test ./... -count=1` (3 known failures only),
  `gofmt -l .` minus vendored, `go run ./cmd/smoke` and `-bots 5`; paste the summary lines.
- [ ] `git diff master --stat -- internal/hooks/testdata/` prints nothing.
- [ ] Opus whole-branch review (correctness, random-draw order, missed references), briefed with the
  agent rules, the spec and this plan. Fixes in their own commits.
- [ ] Phase design "Notes for later slices": the weapon delay moves with swing speed (spec ruling 3);
  the three allowlisted importers and their slices; slice 4 deletes offhand breakage (spec ruling 5:
  `offhandBreakCheck`, broken item 20, `BreakChance`/`BreakTest` if unread; `pvp` golden changes);
  `BackStab` is never set (slice 4 deletes the branch; its reset acts on a copy). Remove from the
  slice 1 note what this slice settled (players-first turn order is `dikubasic.TurnOrder`).
- [ ] `PATCH_NOTES.md` line (player-visible: shooting a player obeys PvP settings in both rooms, for
  player and NPC shooters; a player in the recovery room cannot be attacked from next door; GMCP
  clients see a shooter's aggro the same round);
  `status.md` (`geomud-verifying`'s ok-package count becomes 40: minus `combat`, plus `rules` and
  `rules/dikubasic`; add under Watch: `hooks` `TestCombatGoldenFleeSuccess` fails under `go test -shuffle=on`,
  on the baseline too; order-dependent, not caused by this slice); `geomud-verifying` package count (`internal/rules` and `internal/rules/dikubasic` are
  tested packages; `internal/combat` is gone).
- [ ] Commit the docs: `git add PATCH_NOTES.md status.md 2026-11-06-phase-5-rules-seam-design.md
  .claude/skills/geomud-verifying/SKILL.md`, message `Phase 5.2 done: patch notes, status, notes for
  later slices`.
- [ ] Push `strip/phase-5-2-seam-fight`, open the PR on `pruuk/GeOMUD` (`gh pr create --repo pruuk/GeOMUD`),
  merge once CI is green (memory `claude-merges-prs`); move the spec and this plan to `docs/completed/`.
