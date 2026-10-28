# Phase 4b slice 5: one combat path (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** one attack (`combat.Attack(attacker, defender *actors.Actor)`) and one combat round over
every actor in a fight, with outcomes kept and combat text delivered through `internal/messaging`.
NPCs flee by the same rules as players.

**Architecture:** first pin today's behaviour: a seedable RNG for tests, a golden combat test that
records, roll for roll, what every receiver reads and what state changes in each kind of fight, and
a smoke fight. Then replace the four attack variants with `combat.Attack`, whose result carries one
`perception.Message` per swing instead of room-message fields; merge `handlePlayerCombat` and
`handleMobCombat` into one pass over actors with aggro; move combat command and death room text to
`messaging.Send`. The golden test must match afterwards, except for a listed set of intended text
changes.

**Tech stack:** Go, testify. No new dependency.

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md`, section 2, section 3 ("One attack
path", "One combat round", "Outcomes stay identical"), section 4 slice 5, and its "Notes for later
slices" entries for slices 3, 4 and 5. Inventory: `2026-10-28-phase-4b-5-inventory.md` (D: combat
package; E: the round; F: output and parity), cited below as Inv D, Inv E, Inv F with section
numbers. Branch: `strip/phase-4b-5-one-combat-path` (created; first commit moved the slice 4 plan).

---

## Facts verified against source (2026-10-28, master `dc94a7f7`)

The inventories hold the full tables with file:line; these rows are the ones the tasks build on.

| Fact | Value | Where |
|---|---|---|
| Variants | `AttackPlayerVsMob`, `AttackPlayerVsPlayer`, `AttackMobVsPlayer`, `AttackMobVsMob` all call one `calculateCombat(src, tgt, srcType, tgtType)` (`combat.go:31,49,67,82`); rolls, damage, crits and message choice are shared. They differ in: `WimpyCheck` (account option), damage credit (`TrackPlayerDamage`: PvM always, PvP on hit only in the hook at `NewRound_DoCombat.go:296`, MvM to the charmer, MvP none), `ApplyHealthChange(0)` on a miss (PvM, MvM), and name rendering (`GetMobName` for NPCs) | Inv D 2 |
| Variant call sites | `NewRound_DoCombat.go:265, 442, 681, 812` (the spec's facts table has older numbers) | Inv D |
| `AttackResult` | room-message fields `MessagesToSourceRoom`, `MessagesToTargetRoom`, unused `MessagesToRoomOld`/`SendToRoomOld`; dead `DamageToSource`, `BuffSource` (never set); `DamageToTargetReduction` never read | Inv D 3 |
| Round | `DoCombat` runs `handlePlayerCombat` (players, `:40-520`), then `handleMobCombat` (NPCs, `:522-897`), then `handleAffected` (`:899-947`); four branches PvP, PvM, MvP, MvM; flee resolved for players only (`:69-168`); within each kind the order is Go map order | Inv E 1, 5 |
| Kill credit | the round only fills `PlayerDamage` and the affected lists and calls `MakeHostile` (PvM only, `:468-470`); kills, XP, alignment, mastery and loot are in `suicide`/`suicide_npc.go` | Inv E 2 |
| Kept bugs | E8c (`:100`, no player ever blocks a flee), E8e (`:437`, appends user ID 0), E8f (`:51` before `:55`, nil deref order), `KilledBy` never set by combat (`-10` branch unreachable); E8g and `playersAttacked` are gone | Inv E 1 |
| Drift | NPC offhands never break under player hits; broken items land in the attacker's room; PvP end and re-target fire no `AggroChanged`; PvM fires none when the target moved; `defRoom` unchecked in PvM and MvM; a hidden target freezes `RoundsWaiting` only in MvP | Inv E 2 |
| Hidden | `combat.go:124` hides only wait/prepare lines from target and room; `attack.go:222,270`, `shoot.go:92,122` hide opening lines; every swing names a hidden attacker; no buff in `empty` grants `hidden` | Inv D 2, Inv F 2 |
| Message data | attack lines name actors as `<ansi fg="{sourcetype}">{source}</ansi>` and `<ansi fg="{targettype}">{target}</ansi>`, plus `{itemname}`, `{damage}`, `{exitname}`, `{entrancename}`; 34 lines put the possessive `'s` inside the tag; `cleaving.yaml:69` has a bare `{target}` | Inv D 4, Inv F 1 |
| RNG | `util.Rand` is the global `math/rand.Intn` (`util.go:173-178`); no seed hook | Inv E 5, Inv F 3 |
| Parity | `cmd/smoke/testdata/pre-4b/bot0.txt` has no combat; no test exercises `calculateCombat` or the variants; the round's only test is `mobflee_test.go` | Inv D 6, Inv F 3 |
| NPCs never start fights | nothing sets an NPC's aggro onto a player ("mobs no longer start or join fights until Phase 5", `PATCH_NOTES.md`), so MvP and MvM run only if something makes an NPC `attack` | Inv F 0 |
| Messaging | `messaging.Send` has no production caller yet; receivers are resolved at send time; tokens `{actor}`, `{target}`, `{direction}`; texts `ActorText`, `TargetText`, `ObserverText`, `TargetRoomText`, `FarText` | `internal/messaging` |

## Decisions

Calabe, 2026-10-28 (slice 5 planning):

1. **A hidden attacker reads "Someone" everywhere.** Every line a hidden attacker produces for
   others (opening, wait, swing, dodge) reads "Someone ...": others know they are attacked, not by
   whom. Messaging's unseen-actor rule does this; the per-line hiding goes.
2. **Anyone fighting the fleer may block a flee, and a blocked fleer keeps fighting.** Every actor
   (player or NPC) whose target is the fleer gets the block roll (fixes E8c). A blocked or exitless
   flee restores the target the fleer had before fleeing (today it is lost and the fleer stops
   swinging, stuck "in a fight" with no target).
3. **Every swing earns kill credit.** The attacker's account, or its charmer's account for a
   charmed NPC, is credited in the target's `PlayerDamage` on every attack, hit or miss, whatever
   the target. (Today: PvM on every swing, PvP on hits only, a charmed NPC only against NPCs.)

Defaults set by the controller (from "outcomes stay identical", "one rule for every actor", and
the rulings above; Calabe reviews them with this plan):

4. **Pin before changing.** A seedable RNG for tests, a golden combat test recorded on today's code
   for every branch (PvM, PvP, MvP, MvM, flee, death), and a smoke fight, all committed before any
   combat code changes. Afterwards the golden test must match except for the intended changes listed
   in Task 7, each named in the commit that updates the goldens.
5. **Turn order stays: actors with an account act first, then NPCs** (`Account() != nil`), each
   group in today's map order. It decides close fights; initiative belongs to the Phase 5 ruleset.
6. **Attacks through an exit** (`Aggro.ExitName`, set by `shoot`) work for every actor, still for
   one round (the end-of-round `SetAggro` reset stays).
7. **Drift is fixed to the majority behaviour:** an NPC's offhand can break under a player's hit; a
   broken item lands in the defender's room; every aggro end and re-target fires `AggroChanged`; a
   missing target room clears the aggro with an event; the wait countdown runs before the hidden
   target check in every case (a hidden target no longer freezes it). E8e and E8f go with the
   merge. `KilledBy`: set the killer when combat damage downs a player (the unreachable `-10`
   branch goes), so `PlayerDeath.KillerMobId` and elite deaths work.
8. **Group hatred** (`MakeHostile`) follows the credit rule: the attacker's account, or its
   charmer's, when the target is an NPC with groups.
9. **Wimpy** stays an account option: `WimpyCheck` runs for a defender (and attacker, where today)
   with an account.
10. **Names in combat text are messaging tokens.** The YAML attack lines move to `{actor}` and
    `{target}` (the `<ansi fg="{sourcetype}">...</ansi>` wrappers go: messaging colours names);
    `{itemname}`, `{damage}`, `{exitname}`, `{entrancename}` are filled before sending. Intended
    text changes: NPC names lose the adjectives `GetMobName` adds in combat lines; a possessive
    `'s` falls outside the name's colour; the dodge line's name gains its colour; after a crit or
    backstab prefix "someone" stays lower case. Listed in Task 7.
11. **One message per swing.** `combat.Attack` returns the swing's lines as `[]perception.Message`
    (one per swing, plus wait lines), sent in order by the round. Defender-room text uses
    `TargetRoomText`.
12. **Combat output through messaging** in this slice: the round, `combat`, and the combat commands
    `attack`, `shoot`, `flee`, `consider`'s world text, and the room lines of NPC and player death
    ("has died"). The player-death `Broadcast` (World range, with screen-reader and refresh fields
    messaging lacks), party XP lines and other per-member sends stay for slice 7, as do interface
    lines that only `Tell` the actor (they already reach only the actor).
13. **NPC flee moves.** A successful NPC flee moves it through `rooms.MoveToRoom` (as a player's
    does), with no `Look`; the slice 4 stand-in that only cleared a fled NPC's aggro goes, and its
    test becomes "a fled NPC leaves".
14. **The smoke fight:** bot 0 attacks the rat in Town Square until it dies; with `-bots 2`, bot 1
    is in the room and must receive the observer lines and none of the attacker's (Done-when). The
    rat does not fight back (NPCs do not start fights until Phase 5).

## File structure

| File | Change |
|---|---|
| `internal/util/util.go` (+ test) | `SeedRandForTest(seed)` |
| `internal/hooks/combat_golden_test.go`, `internal/hooks/testdata/combat-golden/*.txt` (new) | golden combat test |
| `cmd/smoke/script.go`, `main.go` | the fight step, observer check |
| `internal/combat/combat.go`, `attackresult.go` | `Attack`; variants and room fields gone |
| `_datafiles/world/empty/items/attack-messages/*.yaml` (path: check) | tokens |
| `internal/hooks/NewRound_DoCombat.go` | one round |
| `internal/usercommands/attack.go`, `shoot.go`, `flee.go`, `consider.go`, `suicide.go`, `suicide_npc.go` | world text through messaging |
| `internal/guards/combat_test.go` (new) | guards |
| records | Task 9 |

---

### Task 0: Commit the plan and inventory

- [ ] `git add 2026-10-28-phase-4b-5-inventory.md 2026-10-28-phase-4b-5-one-combat-path-plan.md` and
  commit `Phase 4b slice 5 plan and inventory: one combat path` (quoted heredoc, Co-Authored-By line).

---

### Task 1: A seedable RNG for tests

**Files:** `internal/util/util.go`, `internal/util/util_test.go` (or a new `rand_test.go`).

- [ ] **Step 1: Failing test:** `SeedRandForTest(42)` then ten `Rand(1000)` values; `SeedRandForTest(42)`
  again gives the same ten; a different seed gives a different sequence. Run: FAIL (undefined).
- [ ] **Step 2: Implement.** Read `util.go:160-194`. Keep today's behaviour by default (a source
  seeded from the clock, as the global `math/rand` is), behind a package variable guarded by a mutex
  (the global was safe for concurrent use; keep that):

```go
var (
	randMu  sync.Mutex
	randSrc = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// SeedRandForTest makes Rand and every other helper in this file that draws
// from the shared source repeatable. Tests only: production keeps the clock
// seed.
func SeedRandForTest(seed int64) {
	randMu.Lock()
	defer randMu.Unlock()
	randSrc = rand.New(rand.NewSource(seed))
}
```

  and change `Rand` (and every other function in `util.go` that calls `rand.` directly: grep
  `rand\.` in `internal/util` and list them) to draw from `randSrc` under `randMu`. Check the rest
  of the repo for direct `math/rand` use in the combat path (`grep -rn '"math/rand"' --include=*.go
  internal`): list any in `combat`, `characters` (`RollDice`?), `items` (`BreakTest`), `mobs`; route
  those through `util` if they decide combat outcomes, or report them (a golden test with an
  unseeded roll in it is flaky).
- [ ] **Step 3:** Run the test, `go build ./...`, `go test ./internal/util/ -count=1`. Commit
  `util: seedable random source for tests`.

---

### Task 2: Golden combat test on today's code (Decision 4)

**Files:** `internal/hooks/combat_golden_test.go`, `internal/hooks/testdata/combat-golden/`.

- [ ] **Step 1: The harness.** In package `hooks` (internal test), build each fixture from scratch:
  rooms with `rooms.AddTestRoom` (and `rooms.AddTestDefaultBiome()`), players as accounts plus actors
  (as `internal/usercommands/*_test.go` and `internal/rooms/move_test.go` build them), NPCs with
  `mobs.NewTestInstance`, weapons with an inline `Spec` (a `Shooting` subtype where race data would
  be read), fixed stats and health. Register an `events.Message` listener and turn every message
  into per-receiver lines the way `hooks/Message_SendMessages.go` delivers them (a `UserId` message
  to that user; a `RoomId` message to every player in the room minus `ExcludeUserIds`, honouring
  `IsQuiet`), so the record is "what each player read". Also record, after each round: each actor's
  health, aggro target and type, room, `PlayerDamage`, and the input commands queued
  (`events.Input`, e.g. `suicide`, `flee`, `equip`). `SeedRandForTest(seed)` before each fixture.
- [ ] **Step 2: Fixtures** (each runs `DoCombat(events.NewRound{})` for N rounds, N chosen so the
  fight shows hits, misses and an end): player attacks NPC to its death; player attacks player;
  NPC attacks player (set the NPC's aggro directly; nothing in data does it); NPC attacks NPC
  (a charmed NPC, for the charmer credit); a player flee blocked by an NPC; a successful player flee;
  a player shooting through an exit (`SetAggroRemote`); a player fighting with a second player
  watching (observer lines); a hidden attacker (set the buff flag through `buffs.SetTestSpec` as
  `internal/usercommands/items_test.go` does). Write the record to
  `testdata/combat-golden/<fixture>.txt` when run with `-update` (a test flag), compare otherwise.
- [ ] **Step 3: Prove it catches a change.** Record the goldens; change one combat text in the data
  or one roll threshold with a temporary Edit; see the test fail naming the fixture and line;
  revert; `git diff` clean apart from the new files. Run the test twice more to prove it is stable
  (no unseeded randomness): if a fixture differs between runs, find the unseeded roll (Task 1 Step 2)
  and fix the source, not the test.
- [ ] **Step 4:** Commit `hooks: golden combat test recorded on the pre-slice-5 code` with the goldens.

---

### Task 3: The smoke fight (Decision 14)

**Files:** `cmd/smoke/script.go`, and what the step needs.

- [ ] Read Inv F 4 and `cmd/smoke/script.go`. After the walk and `say` steps, bot 0 runs
  `attack rat` and waits (a step timeout long enough for the fight: measure, at most 2 minutes) for
  the rat's death line (find today's exact text in `suicide_npc.go` and the round, e.g.
  `rat has died`), then checks its XP line. With more than one bot, bot 1 waits in the room for
  bot 0's fight: it must see the observer form of a swing line naming bot 0, and must not see any
  attacker-only text (`You ` lines from the attack set). Use a sync point so bot 1 is in Town Square
  before bot 0 attacks (read how the smoke bots share state, e.g. the copyover `WaitGroup`). Other
  bots do nothing new. Several rats may be in the room (the spawn and the admin's from slice 4):
  any rat dying ends the step.
- [ ] Run `go run ./cmd/smoke`, `-bots 2` and `-bots 10` on today's code: PASS, three lines. Record
  `-bots 2 -transcript` into `cmd/smoke/testdata/pre-5/` as the slice's text baseline (for reading,
  not for byte comparison: rolls differ per run). Commit
  `smoke: a bot fights the rat; a second bot sees the observer lines`.

---

### Task 4: `combat.Attack` (Decisions 3, 9, 10, 11)

**Files:** `internal/combat/combat.go`, `attackresult.go`, the attack message YAML, callers.

- [ ] **Step 1:** Read `combat.go`, `attackresult.go`, `calculations.go` and Inv D 2-4 in full.
- [ ] **Step 2: Signature.** `func Attack(attacker, defender *actors.Actor) AttackResult`, one body
  replacing the four variants (`calculateCombat` stays the shared core; it keeps taking characters by
  value as today). Per-kind differences become actor questions:
  - credit (Decision 3): `creditAccount(attacker)` = the attacker's account, else the account of its
    charmer (`GetCharmedBy`), else none; `defender.Character.TrackPlayerDamage(acct.UserId, damage)`
    on every attack when there is one.
  - `ApplyHealthChange` always (the PvM/MvM behaviour; on a miss it clamps and is otherwise a no-op;
    check the PvP/MvP "only if nonzero" against the golden test and keep whichever the goldens show
    matters, noting it).
  - `WimpyCheck` for whichever side has an account, where today's variants call it.
  - Remove `DamageToSource`, `BuffSource`, `DamageToTargetReduction` if unused (grep), and the
    room-message fields.
- [ ] **Step 3: Messages.** `AttackResult.Messages []perception.Message`: one per swing (and the
  wait/prepare lines from `GetWaitMessages`, which becomes part of the same result or a sibling
  function returning `[]perception.Message`). Each message: `Kind: WorldEvent`, `Actor` the
  attacker, `Target` the defender, `Senses: Sight | Sound` (check: a dark room today still shows
  combat lines to players; if so, Sound must carry them, and say so), `Range: Room` (the target's
  room is a source room when it differs), `ActorText` = the to-attacker line, `TargetText` = the
  to-defender line, `ObserverText` = the attacker-room line, `TargetRoomText` = the defender-room
  line. Tokens per Decision 10: convert the attack message YAML (find the folder: `grep -rl
  sourcetype _datafiles`) from `<ansi fg="{sourcetype}">{source}</ansi>` to `{actor}` and the target
  form to `{target}`, fix the bare `{target}` in `cleaving.yaml`; fill `{itemname}`, `{damage}`,
  `{exitname}`, `{entrancename}` with `strings.Replacer` before building the message. The token
  rename touches every attack message file: list counts per file. The name colour per receiver now
  comes from messaging.
- [ ] **Step 4:** Keep the hook compiling: the round temporarily sends `AttackResult.Messages`
  through `messaging.Send` at each old call site (Task 5 rewrites the round). Update `simulate.go`.
- [ ] **Step 5:** `go build ./...`; run the golden test: it fails only on the intended text changes
  of Decision 10 and the credit rule (Decision 3) in the PvP fixture; list every failing line and
  classify it. Do not update the goldens yet (Task 7 does, once). Commit
  `combat: one Attack for every actor; messages are perception messages`.

---

### Task 5: One combat round (Decisions 2, 5-8, 13)

**Files:** `internal/hooks/NewRound_DoCombat.go`, `internal/hooks/mobflee_test.go`.

- [ ] **Step 1:** Read `NewRound_DoCombat.go` in full and Inv E 1-5.
- [ ] **Step 2: Structure.**

```go
func DoCombat(e events.Event) events.ListenerReturn {
	// Decision 5: actors with an account act first, then NPCs.
	fighters := fightersWithAccountsFirst() // actors.All() filtered to Aggro != nil, split by Account() != nil, each group in today's map order (see Step 3)
	var affected []*actors.Actor
	for _, a := range fighters {
		affected = append(affected, combatTurn(a)...)
	}
	handleAffected(affected)
	return events.Continue
}
```

  `combatTurn(a)` runs, for any actor: the `no-combat` skip and nil checks (E8f fixed by order);
  `cancel-on-combat`; NPC `CombatCommands` (when `mobs.GetInstance(a.Id) != nil`, as today, before
  the attack); flee (Step 4); target finding (same room, or through `Aggro.ExitName` for any actor,
  Decision 6); target gone, downed, hidden (wait countdown first, Decision 7); NPC weapon retry (as
  today, MvP's behaviour, now for any NPC attacker); `combat.Attack`; send its messages in order;
  credit and `MakeHostile` (Decision 8); offhand break on a hit for any defender (Decision 7),
  broken item to the defender's room; `CharacterVitalsChanged` for an account on either side as
  today; defender reaction (`PreventIdle`, move toward a remote attacker, angry command) when the
  defender is an NPC with no aggro; end of fight and re-target with `AggroChanged` always.
  `handleAffected`: players with `Health < 1` from combat damage get `KilledBy` set from the killer
  (Decision 7) before `PlayerDrop`/`suicide` as today; NPCs `suicide` as today. Every "player vs
  NPC" test becomes `Account() != nil` or `mobs.GetInstance(id) != nil`; list each in the report.
- [ ] **Step 3: Order.** Today players are iterated from `accounts.GetOnlineUserIds()` and NPCs from
  `mobs.GetAllMobIds()` (both map order). Keep those two sources for the two groups (map order is
  already random per round; `actors.All()` sorts by ID, which would make the order fixed, a change):
  `fightersWithAccountsFirst` maps those IDs to actors. Check the golden fixtures with two actors of a
  kind do not depend on map order (single-fighter fixtures do not); if one does, sort that fixture's
  expectation, not the code.
- [ ] **Step 4: Flee for every actor** (Decisions 2, 13). Store the pre-flee target when `flee` runs
  (`flee.go`: keep the old `Aggro.Target` in a new `Aggro.FleeFrom ids.ID` or similar, read
  `characters.Aggro`); blockers are every actor in the room whose aggro target is the fleer
  (players and NPCs, same formula, `GetBodies` order); blocked or no exit restores the pre-flee
  target (`SetAggro(prev, DefaultAttack)`) with `AggroChanged`; success: texts through messaging,
  aggro cleared with event, `rooms.MoveToRoom(a.Id, exitRoomId)`, charmed followers of the fleer
  get `flee` queued, and `Look` only for an actor with an account. Delete the slice 4 NPC stand-in
  (clearing a fled NPC's aggro) and change `mobflee_test.go` to "a fled NPC leaves its room".
- [ ] **Step 5:** `go build ./...`; `go test ./internal/hooks/ -count=1`: the golden test fails only
  on intended changes (Decisions 2, 3, 7, 10, 13): list and classify every failing line. Commit
  `hooks: one combat round for every actor`.

---

### Task 6: Combat command and death text through messaging (Decisions 1, 12)

**Files:** `internal/usercommands/attack.go`, `shoot.go`, `flee.go`, `consider.go`, `suicide.go`,
`suicide_npc.go`.

- [ ] Read Inv F 1c, 1d, 2. Each world-text site (room sends naming the actor) becomes one
  `messaging.Send` with roles; the per-line `hidden` checks go (Decision 1: messaging renders a
  hidden actor as "Someone" to others). Sites Inv F flags as not fitting keep today's form with a
  comment naming slice 7: the flee-blocked-by-player line if `TargetText` fallback would show it to
  the blocker (give the blocker its own `TargetText`, or keep the send), the `shoot` opening line
  (keep it to the shooter's room: a message with no `Target`, or `Room` range with an explicit room),
  the death `Broadcast`, party XP. Interface lines that only `Tell` the actor stay. Tests: a hidden
  attacker's `attack` opening line reads "Someone ..." for an observer; the NPC death room line
  reaches a player in the room.
- [ ] `go build ./...`, tests, smoke. Commit `usercommands: combat world text goes through messaging`.

---

### Task 7: Goldens and the intended changes (Decision 4)

- [ ] Run the golden test; for every differing line, confirm it is one of the intended changes and
  nothing else:
  1. Names: NPC names without adjectives; possessive outside the colour tag; dodge-line name
     coloured; "someone" lower case after a crit/backstab prefix (Decision 10).
  2. A hidden attacker reads "Someone" in every line others see (Decision 1).
  3. Credit: PvP misses credit the attacker; a charmed NPC credits its charmer against a player
     (Decision 3); group hatred follows (Decision 8).
  4. Flee: a player can block a flee; a blocked fleer keeps fighting; an NPC that flees leaves
     (Decisions 2, 13).
  5. Drift fixes (Decision 7): `AggroChanged` events, offhand break, broken item room, `KilledBy`,
     wait countdown with a hidden target.
  6. Delivery: receivers resolved when messaging sends (NPCs receive structured perceptions; no
     player-visible effect beyond ordering within one swing, which must be unchanged: check).
  Anything else is a bug to fix before updating. Then run with `-update`, review the golden diff
  line by line, and commit `hooks: combat goldens after slice 5 (intended changes listed)` with the
  list in the commit message body.

---

### Task 8: Guards

**Files:** `internal/guards/combat_test.go` (new).

- [ ] No `AttackPlayerVsMob|AttackPlayerVsPlayer|AttackMobVsPlayer|AttackMobVsMob`,
  `handlePlayerCombat|handleMobCombat`, `MessagesToSourceRoom|MessagesToTargetRoom` identifier in
  non-test code; `internal/combat` makes no `SendText`/`Tell`/`events.Message` send (it returns
  messages). Bite-test each (temporary Edit, revert, `git diff` clean). Shrink the slice 3
  connection-writer allowlist if any combat file left it (it should not have been on it).
  Commit `guards: one combat path`.

---

### Task 9: Verify and record (controller)

- [ ] `go build ./...`, `go vet ./...`, `go test ./... -count=1` (expect 39 plus any new test
  packages; the 3 known Windows failures only), `gofmt -l`, `go run ./cmd/smoke`, `-bots 2`,
  `-bots 10`.
- [ ] Records: `PATCH_NOTES.md` (one combat path; the three rulings; drift fixes; NPC flee moves;
  text changes in combat lines); `status.md`; spec section 4 "Built so far" adds 5 and "Notes for
  later slices" (what slice 5 leaves: the death `Broadcast` and party XP for slice 7; per-kind turn
  order for Phase 5's initiative; NPCs still never start fights until Phase 5; the golden test as the
  parity tool for slice 7); `geomud-verifying` count. Move this plan and the inventory to
  `docs/completed/` with the next branch.

## Done when

`combat.Attack` is the one attack and `DoCombat` one round for every actor in a fight; the four
variants, the two handlers and the room-message fields are gone; combat text reaches players through
messaging; the golden test matches with only the listed intended changes; the smoke fight passes,
and with `-bots 2` the second bot receives the observer lines and not the attacker's; the guards
pass; build, the full suite at the baseline, and smoke pass.
