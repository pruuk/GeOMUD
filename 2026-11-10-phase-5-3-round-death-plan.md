# Phase 5 slice 3: round, regen and death through the seam (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved spec `2026-11-10-phase-5-3-round-death-spec.md`: the ruleset says who is
up, downed or dead (`Condition`), how much a body regenerates (`Regen`) and what a death means
(`OnDeath`); every actor dies through one engine routine, `death.Die`; round and regen move with
parity; death goes straight to the bare rules, with a golden of today's death recorded first.

**Architecture:** See the spec ("Design", and "Settled while planning"). This plan adds no design
decision of its own. The questions found while planning are below; the controller accepted every
recommendation, and the spec records each. Code in the tasks that follows one is marked "(Q*n*)".

**Tech stack:** Go, testify, combat and reads goldens in `internal/hooks`, `internal/guards` (Go syntax
tree), `cmd/smoke`.

Branch `strip/phase-5-3-round-death` (exists; holds the spec). Tasks run one after another on the
branch: they share files, so no parallel worktrees.

---

## Questions settled by the controller (recorded in the spec)

Each was found while planning, by reading the code. All recommendations accepted (2026-11-10).

**Q1. The XP line would come before "has died."** `Account.GrantXP` sends its own line
(`accounts/userrecord.go:175-190`, `u.tell(...)`). The spec has the engine call `OnDeath` first and
announce the death second; if `dikubasic.OnDeath` granted XP, the killer would read "You gained N
experience points!" before "rat has died." (today it reads them the other way round: reads golden
H, lines 4-6). It would also put a send inside the ruleset, against "rulesets never send".
*Recommend:* `DeathOutcome` carries `Awards []rules.Award{UserId, XP, Messages}`. The ruleset works
out each share and applies the alignment shift (in today's order, because
`Character.UpdateAlignment` can draw a random number, `characters/character.go:1315`); the engine,
after announcing the death, credits the kill, calls `GrantXP(xp, "combat")` and sends the award's
lines, per award in order. `GrantXP` draws nothing (`LevelUp`, `Validate` have no `util.Rand`), so
moving it after the drop rolls keeps the draw order. Kill credit (`KD.AddMobKill`/`AddEliteKill`)
goes to every award recipient, which is exactly today's set (non-party killers, and every member of
a killer's party). Side effect: `AlignmentChanged` is now queued before the `MobDeath` event and
before the `GainExperience` event (today it comes after both: `suicide_npc.go:66` queues `MobDeath`,
then `:145` grants XP, then `:160` queues `AlignmentChanged`; the party loop likewise, `:210`, `:225`). No golden records `AlignmentChanged`
or `GainExperience`; their listeners are GMCP's (spec, plan review round 1: accepted).

**Q2. A corpse would let anyone copy what the dead kept.** A corpse stores a copy of the whole body
(`Corpse.Character`), and `get all <corpse>`, `get <item> <corpse>` and the corpse panel also take or
show that copy's worn gear (`usercommands/get.go:152-173`, `:207-216`, `:240`;
`look.panels.go:125-140`). Today this is unreachable (shipped `CorpseItems: false` turns corpse
looting off). With `CorpseItems` removed, a player's corpse would hand out copies of the gear the
player kept (decision 11 keeps it on the player), and an NPC's corpse would hand out the worn items
its `ItemDropChance` rolls kept (ruling 3).
*Recommend:* `death.Die` stores a copy of the body with nothing carried, worn or held
(`corpseBody`), so a corpse holds exactly the drops (`Corpse.Items`, `Corpse.Gold`). `get` and the
panel are left as they are (their worn-gear loops then find nothing). *Alternative:* delete those
loops from `get` and the panel instead.

**Q3. Three engine health checks the spec does not list.** (a) The downed gate in `TryCommand`
(`usercommands.go:281`, `userDisabled = actor.Character.IsDisabled()`) and in `Go` (`go.go:46`) ask
`characters.IsDisabled` (`Health <= 0`), not the ruleset, so a ruleset's `Downed` at health above 0
would not stop the actor acting, which the spec's stub test requires ("the actor cannot act").
(b) `usercommands/remove.go:51` compares `actor.Character.Health > 0`, inside a package the new guard
scans; the guard would fail on it. *Recommend:* in Task 3 (parity), (a) both become
`rules.Get().Condition(actor) != rules.Up` and (b) becomes `rules.Get().Condition(actor) == rules.Up`.
Under `dikubasic` each gives the same answer as today (`Dead` exactly when `Health <= 0`).

**Q4. Package tests ship no recovery room, and a player death there panics.** Test binaries load no
`config.yaml`, so `SpecialRooms.DeathRecoveryRoom` is 0 (and `ProtectionLevels` 0,
`EquipmentDropChance` 0, `CorpsesEnabled` false). `MoveToRoom(id, 0)` then calls
`CreateEphemeralRoomIds(0)`, which skips room 0 and indexes an empty slice in its log line
(`rooms/ephemeral.go:144-146`, `:188`): a panic (read, not run). Today no combat golden runs a player
death. After Task 6 the `mvp`, `pvp` and `hidden_attacker` goldens, `TestAttackRunsTheDefendersWimpyCheck`
and the new death fixtures do. *Recommend:* a test helper `recoveryRoom()` (Task 1) gives those
fixtures a recovery room (room C, 93003, unconnected) whose template sits in a temporary data folder;
it also keeps a level-up's `SaveUser` out of the repo's data folder. The engine panic is left alone
(the shipped room is 75) and goes on the status Watch list.

**Q5. Smoke: corpse names collide from 11 bots up.** The bot takes its ring back with
`get copper ring <its name>`. `FindCorpseByRef` matches by prefix and returns the first corpse whose
name starts with the word (`rooms.go:945-990`, `util.go:373-397`): `Smk<tag>c1` matches
`Smk<tag>c10 corpse`, so with `-bots 50` one bot can empty another's corpse and the other then fails.
*Recommend:* pad the bot number to three digits (`cmd/smoke/main.go:124`, `:131`:
`Smk%sc%03d`), so no bot name is a prefix of another.

**Q6. The spec's list of expected golden changes is short.** It names the `suicide` input lines, the
death fixtures and reads golden H's NPC- and player-death sections. Also changing in Task 6, all
"decided" (decision 7: die at 0; decision 11; the `CorpseItems` removal): `mvp`, `pvp` and
`hidden_attacker` (a downed player now dies in the round, no `PlayerDrop`), `mvm` (the rat's death
now runs), and H's follow section (Ann's look loses "On the Ground: 25 gold", the rat's gold now being
in its corpse). Test code that must change with them: `ruleset_stub_test.go:43` (asserts the queued
`suicide`), `TestKilledByOutlivesTheKiller` (its scenario, a killer destroyed between the downing and
the death, cannot happen when death is immediate; it moves under a downed-state stub),
`TestAttackRunsTheDefendersWimpyCheck` (Alice now dies: needs the recovery room),
`TestActorUpkeepTicksEveryActorsCooldowns` (its bare NPC is at 0 health and upkeep would now end it),
and `TestCombatGoldenPvP` (`combat_golden_test.go:421`) and `TestCombatGoldenMvP` (`:434`), whose
`require.Less(health, 1, ...)` fails once the dead player respawns at 1: each asserts the death
instead (the player is in a private copy of the recovery room). `TestCombatGoldenHiddenAttacker`
(`:512-523`) has no health assert; it gains the same death assert, so the fixture says what it
reaches.
*Recommend:* accept as decided; Task 6 lists each.

**Q7. Engine details the spec leaves open; the plan's choice in each case:**

- a. Round upkeep and the immediate buff trigger call `Die` for any actor the ruleset calls dead
  (today they queue `suicide` for NPCs only). Spec: "every actor dies through one engine routine".
- b. A ruleset that returns `Respawn: false` for a body with no NPC instance (an account): the engine
  respawns it anyway. Ending needs an NPC instance, and there is no permadeath.
- c. Respawn keeps today's `Mana = 0` (`suicide.go:305`); the recovery room restores it.
- d. Corpses off: drops go to the floor with today's NPC lines ("<item> drops to the ground.",
  "<n> gold drops to the ground."), for players too (players' floor drops used the `drop` command's
  lines).
- e. `MobItemDrop` is queued for every NPC drop, corpse or floor (today only for floor drops).
- f. Reviving an account queues `CharacterVitalsChanged` (an account never revived before).
- g. `Downed`: only an account gets `PlayerDrop` (the event takes a `UserId`); a downed NPC reads
  nothing.
- h. The recovery-room regen applies to any actor in that room (outside it, accounts only, for
  parity); its two dice are drawn even at full health, as upstream's buff did.
- i. Task 3's `Regen` keeps today's raw-ID recovery compare (parity: today a player in the private
  copy regenerates normally); Task 5 switches to `rooms.IsDeathRecoveryRoom` with the fast regen.
- j. `suicide` in the recovery room keeps today's error return (`already dead`); the in-fight refusal
  returns none, as `quit` does.
- k. The help pages ride with Task 10 (commit 10): no golden covers `help death`, `help health` or
  `help get` (reads golden G records `help say` only).
- l. With `CorpseItems` gone, `get <item> <word>` always tries a corpse named by the last word, as it
  did whenever `CorpseItems` was on.
- m. `actors.BodyReplaced` (`actors/registry.go:180-`) loses its only production caller (the
  permadeath reroll). Left in place with its test, for the slice 4 shrink.
- n. `TestIdsAreMadeOnlyAtCreation`'s allowlist entry for `suicide.go` ("permadeath reroll",
  `guards_test.go:167`) is removed with the reroll. That guard has no stale-entry check, so nothing
  forces it; leaving it would let `suicide.go` make IDs.

## Plan review round 1: what this revision changed

The spec's "Plan review round 1 (controller decisions)" are implemented here:
- `get all <corpse>` and `get all from <corpse>` take everything from that corpse; every other
  `get all ...` (`get all`, `get all ground`, `get all stash`, `get all <container>`, any other word)
  keeps today's room-wide behaviour exactly. The room-wide branch at `get.go:40` is skipped only when
  `namesACorpse` finds a corpse (Task 6 Step 7b; tests both ways in Step 11b).
- Death drops queue `ItemOwnership{Gained: false}` for each item that leaves the body, player or NPC
  (`takeDrops`, Task 6 Step 3; asserted in `TestDeathWithCorpsesOffDropsOnTheFloor` and in the death
  goldens, Step 16).
- `IsDeathRecoveryRoom` is false with no recovery room (Task 2).
- Buff 24 and room 75's mutator reference are deleted (Task 5).
- The event-order note (Q1). The review's other findings are
applied in the tasks; the calls this revision made while applying them, none of which changes what a
player sees:

- **The combat record trims a trailing `\r`** (Task 1 Step 6): the player-death notice ends in
  `term.CRLFStr`, and `compareGolden` strips `\r\n` from the golden side only.
- **`RegenResult.Changes()`** (Task 3) holds the regen pass's "is there anything to apply" check in
  `internal/rules`, outside the guarded packages, so the regen pass compares no `.Health` with a
  number.
- **`wasUp` is not added** (Task 3 Step 5): the rage check (`NewRound_DoCombat.go:169`) returns
  before any attack on a defender who is not `Up`, and nothing between it and the attack changes the
  defender's health, so "the defender was up before the attack" is always true there. A variable
  holding it could never be shown to matter (its mutant is equivalent).
- **The respawn does not clear the killer record a second time** (Task 6 Step 3): `recordPlayerDeath`
  clears it (as `suicide.go:99-109` did) and books an elite killer first; a second clear in `respawn`
  made each clear impossible to test on its own.
- **Every test death's recovery copy is freed** (the recovery-room helpers): each copy takes one of
  100 ephemeral chunks for good, so `-count=16` ran out after about 14 repeats.
- **The death fixtures record more** (Task 1): the kill and death books (`kd=`) in their state lines,
  `MobItemDrop` events, an elite rat and an elite guard, and the rat's ending (its ledger entry and
  its leaving the room), all on today's code first, so Task 6 shows each holding.
- **The `.IsDisabled()` ban** (Task 8) bans a body's `IsDisabled` (`X.Character.IsDisabled()`, or any
  `IsDisabled` call in a file not on a short allowlist). The two files on the allowlist call
  `items.Item.IsDisabled` (a disabled equipment slot), which is not a body's condition.

## Where the spec's facts table is wrong or short (checked at `225c08ba`)

- **Help pages** live at `_datafiles/world/empty/templates/help/` (`death.template`, `health.template`,
  `get.md`), not `templates/help/`. `keywords.yaml` is `_datafiles/world/empty/keywords.yaml` (`:18`
  correct). `death.template` and `config.yaml` have CRLF line ends; `health.template` and `get.md` LF.
- **Health thresholds:** missing `hooks/NewRound_AutoHeal.go:68` (`Health > 0`) and
  `usercommands/remove.go:51` (`Health > 0`, inside the guarded set; Q3). The `characters` list also
  misses `:464` (`Health > 0`, health percentage) and `:1204` (`< -10`) and the remove-lock checks
  (`:1522-1648`); all outside the scanned set.
- **Downed gate:** `TryCommand` (`usercommands.go:281`) and `go.go:46` read `IsDisabled`, not a health
  literal (Q3).
- **Raw recovery-room compares:** also `rooms/save_and_load.go:437` (room loading; template rooms only,
  never a copy; left alone).
- **Admin pages:** `config-wizard.html` entries are `:1110-1130` (lives), `:1161-1200`
  (`EquipmentDropChance` through `PermaDeath`), `:1214-1219` (`CorpseItems`), the Lives group
  `:1543-1551` and the Death group's `desc`/`fields` `:1555`, `:1559`. `config.html` is `:228-238`, not
  `:232-238`. Missing: `users.html:366-369`, `:592`, `:1089` (the Extra Lives field in the user editor).
- **Other readers of deleted knobs:** `configs/configs_bench_test.go:28` sets `Death.XPPenalty` (stops
  compiling when the field goes); `:174-176` name three deleted keys in a map. `config.yaml:196-237`
  holds the deleted entries.
- **`GrantXP` sends** and **`UpdateAlignment` draws**: not in the table; both matter (Q1).
- **Corpse looting reads the corpse's body copy:** not in the table (Q2).
- **Tests run with config zero values** and a player death there panics (Q4).
- **Expected golden changes** are short of the real list (Q6).
- `decayAlignment` is `NewRound_ActorUpkeep.go:71-91` (the spec says `:69-91`).

---

## Line numbers this plan relies on (verified 2026-11-10 at `225c08ba`, read in chunks of 400 lines or less)

| What | Where |
|---|---|
| `Ruleset` interface; `fake` in the rules test | `internal/rules/rules.go:30-61`; `rules_test.go:12-22` |
| `dikubasic` methods; `CanAttack` raw recovery compare | `ruleset.go:27-113`; `:91-95` |
| `dikubasic` tests: helpers, recovery-room test | `ruleset_test.go:21-49`; `:101-142` |
| `Suicide` (player), `npcDeath` (NPC) | `usercommands/suicide.go:27-313`; `suicide_npc.go:26-371` |
| XP draw; per-user loop; party loop; worn-item roll | `suicide_npc.go:88`; `:97-176`; `:181-238`; `:269-281` |
| Registry entry and comment | `usercommands.go:186-188` |
| Downed gate | `usercommands.go:281` (set), `:314-317` (check; the refusal `Tell` is `:315`) |
| Root package `TestMain` (sets no ruleset) | `world_test.go:66-86`; `TestNPCInputReachesTheSharedRegistry` `:132` |
| `get`: the early `all` return; corpse source | `get.go:40-54`; `:90-100` |
| `Go` downed check; `inFight` | `go.go:46`; `go.go:353-370` |
| Cursed remove check | `remove.go:51` |
| `quit` refusal | `quit.go:18-21` |
| Corpse source in `get`; worn loops | `get.go:90-100`; `:152-173`, `:207-216`, `:240` |
| Corpse panel call; panel | `look.go:412-414`; `look.panels.go:110-157` |
| `Corpse` fields | `rooms/corpse.go:9-17` |
| `CanPvp` | `rooms/rooms.go:2061-2089` (check `:2063`) |
| `UpdateCorpses`; `FindCorpseByRef` | `rooms.go:231-258`; `:945-990` |
| `GetOriginalRoom`; `IsEphemeralRoomId`; empty-copy log | `rooms/ephemeral.go:284-289`; `:212-214`; `:185-189` |
| Ephemeral chunks: limit 100; `FindEphemeralRoomIds`; `TryEphemeralCleanup` (frees a chunk only when no player is in it) | `ephemeral.go:14`; `:55-65`; `:216-262` |
| Room player bookkeeping (`PlayerCt`, `GetPlayers`, `RemovePlayer`) | `rooms.go:1054`, `:1158`, `:1807-1815` |
| `NPCEnded` (writes `rooms.instances/npcs.endings.yaml` under the data folder when `SpawnSlot > 0`); `LoadEndings` | `rooms/npcstore.go:841-853`; `:771-795` |
| Buff maps: `buffs.New`; a bare test NPC gets them as in `givePermaGear` | `buffs/buffs.go:43-49` (the nil-map write `:187`); `usercommands/items_test.go:42` |
| `KDStats` fields | `characters/kdstats.go:8-20` |
| Upstream's inert buff 24 and its mutator; room 75's reference | `_datafiles/world/empty/buffs/24-death_recovery.yaml`; `mutators/death_recovery.yaml:16`; `rooms/shadow_realm/75.yaml:18-19` |
| Recovery copy on move | `rooms/roommanager.go:298-306` |
| `DoCombat`; `combatTurn` rage; downing; aggro end | `hooks/NewRound_DoCombat.go:28-43`; `:169-173`; `:201`, `:213-228`; `:269-275` |
| `fighters`; `accountOf` | `NewRound_DoCombat.go:48-61`; `:64-69` |
| `handleAffected` | `NewRound_DoCombat.go:474-520` |
| `AutoHeal` | `hooks/NewRound_AutoHeal.go:1-89` |
| Upkeep NPC death | `NewRound_ActorUpkeep.go:62-65` |
| Buff trigger death | `Buff_ApplyBuffs.go:62-72` |
| `PlayerDrop` handler | `PlayerDrop_HandlePlayerDrop.go:18-48` |
| Listener order (NewRound) | `hooks/hooks.go:18-32` |
| `Heal`; `c.Mana += hp` | `characters/character.go:988-1000`; `:996` |
| `IsDisabled`; `ExtraLives` field | `character.go:818-820`; `:80` |
| `UpdateAlignment` random draw | `character.go:1302-1320` (draw `:1315`) |
| `RemoveItem`; `RemoveFromBody` | `character.go:550-558`; `:1644-1657` |
| `Account.GrantXP`; lives in it | `accounts/userrecord.go:175-257`; `:202`, `:204`, `:216`, `:221-223`, `:245-249` |
| `NewAccount` lives | `userrecord.go:53`, `:69-71` |
| `start` lives | `usercommands/start.go:182` |
| `permadeath` template func | `templates/templatesfunctions.go:193-195` |
| `LevelUp.LivesGained`; `PlayerDeath.Permanent` | `events/eventtypes.go:208`; `:225` |
| Death config struct; validation | `configs/config.gameplay.go:17-19`, `:76-85`; `:92-131`; imports `:3-6` |
| Combat golden: consts, world, listeners | `hooks/combat_golden_test.go:39-43`; `:51-63`; `:100-147` |
| Combat golden: `collect` (trims only `\n`, `:173`); `compareGolden` strips `\r\n` from the golden side only | `combat_golden_test.go:167-182`; `reads_golden_test.go:200-226` (`:210`) |
| Combat golden health asserts that a respawn breaks | PvP `:421`, MvP `:434` |
| Player death notice ends in `term.CRLFStr` | `usercommands/suicide.go:88-90` |
| Combat golden: `state`, `writeState`, `start`, `rounds` | `:296-315`; `:317-341`; `:344-349`; `:367-388` |
| Combat golden fixtures | PvM `:397-407`, PvP `:411-422`, MvP `:426-435`, MvM `:438-448`, hidden `:512-523`, killer `:542-559` |
| Reads harness: recorder Input, `setConfig`, `bodyIdPattern` | `reads_golden_test.go:98-102`; `:566-587`; `:421` |
| Reads golden H | `reads_golden_test.go:830-906` |
| Stub test asserting `suicide` | `hooks/ruleset_stub_test.go:43` |
| Wimpy test | `hooks/attack_engine_test.go:14-24` |
| Upkeep cooldown test (bare NPC) | `hooks/actortick_test.go:43-58` |
| Guards: `goFiles`; import allowlist; ID allowlist | `guards/guards_test.go:18-48`; `combat_test.go:120-125`; `guards_test.go:163-169` |
| Smoke death step; bot names | `cmd/smoke/script.go:221-229`; `cmd/smoke/main.go:124`, `:131` |
| Test mains | `hooks/actionpoints_test.go:24-38`; `usercommands/usercommands_test.go:27-41`; `rooms/rooms_test.go:40-55` |
| Test helpers | `goSetup` (`go_test.go:18-31`), `playerTexts` (`:34-43`), `capture[T]` (`items_test.go:48-63`), `setup` (`usercommands_test.go:45-83`) |
| Upstream recovery buff | `git show 57255ee4^:_datafiles/world/empty/buffs/24-death_recovery.js` |

If a line has moved when you get there, find the code by its text and say so in the report.

## Agent rules (brief every subagent with these)

- Edit files with the Edit and Write tools. Never write a repo file from the shell (no `>`, `>>`,
  `sed -i`, Python open-write); a hook blocks it. `git mv` and `git rm` are fine.
- Read files in chunks of at most 400 lines; state "read N of N files" with the list in the report.
- `git add <named paths>` only, never `-A` or `.`; commit messages through a quoted heredoc ending
  with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never `--no-verify`.
- Never push. Never touch `upstream`.
- Run every test command in the foreground with an explicit timeout of 300000 ms; never wait on a
  background run.
- A grep that proves absence must also show it can match (run the same pattern on a symbol that
  exists).
- `go test` always with `-count=1`. **`-update` runs only in Task 1 and Task 6, only with the exact
  `-run` pattern the task gives.** Anywhere else a golden diff is a failure to report, not to
  re-record.
- If a step's code does not compile or a prescribed failure does not happen, report it; do not paper
  over it.
- Run `gofmt -w` on every Go file you created or changed before the per-task check.
- A mutation probe is: Edit, run, see the named test fail, paste the failure line, undo with Edit,
  rerun, see `ok`. `git diff --stat` after the probes shows only the task's files.

## Per-task check (end of every task)

```bash
go build ./...
go vet ./... 2>&1 | head
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'   # only the 3 known Windows failures (geomud-verifying)
gofmt -l internal cmd *.go                                    # prints nothing (covers new, untracked files)
git status --short -- internal/hooks/testdata/ _datafiles/    # prints nothing (Tasks 1, 5, 6, 10: after their commit)
```

Report: build result, the failing-test lines, and the `git status` output (must be empty).

---

### Task 0: Baseline (controller)

- [ ] On `strip/phase-5-3-round-death`, run the per-task check; save the output in the scratchpad as
  the baseline (expected: the 3 known Windows failures only).
- [ ] Settle Q1-Q7 (Calabe for Q1's read order, Q2 and Q5 if he wants them; record the answers in the
  spec's "Decisions taken without asking").
- [ ] Commit this plan once Calabe approves it: `git add 2026-11-10-phase-5-3-round-death-plan.md`,
  message `Phase 5.3: plan (round, regen and death through the seam)`.

### Task 1: Death goldens on today's code (Opus)

Test files and testdata only. The combat golden harness gains a mode that runs the inputs a round
queues, a regen-pass step for the bleed-out, a recovery room in a temporary data folder (whose
private copies are freed after each test), a `MobItemDrop` note, a record line that keeps the
player-death notice's line end out of the text, and three fixtures recorded once on today's code.
The 9 existing combat goldens and every reads golden stay byte-identical.

**Files:** modify `internal/hooks/combat_golden_test.go`; create
`internal/hooks/testdata/combat-golden/death_npc.txt`, `death_player.txt`, `death_pvp.txt` (by
`-update`).

- [ ] **Step 1: Header comment.** In the file comment, the last two lines (`:13-14`)
  `// colour change shows. Queued commands (events.Input) are recorded, not run:` /
  `// a "suicide" in the record means the round queued it.` become:

```go
// colour change shows. Queued commands (events.Input) are recorded, not run:
// a "suicide" in the record means the round queued it. The death fixtures
// (runInputs, Phase 5 slice 3) also run them after recording, as the game
// loop would, and a private copy of a room reads copy-of-<room>.
```

- [ ] **Step 2: Imports.** Add `"os"`, `"path/filepath"`, `github.com/GoMudEngine/GoMud/internal/configs`
  and `github.com/GoMudEngine/GoMud/internal/usercommands`.

- [ ] **Step 3: Room C and the world's new fields.** The const block (`:39-43`) becomes:

```go
const (
	goldenRoomA     = 93001
	goldenRoomB     = 93002
	goldenRoomC     = 93003 // the death recovery room (recoveryRoom); no exit leads to it
	goldenHiddenBuf = 93901
)
```

  In `combatWorld` (`:51-63`), after `events []string` add:

```go
	// deaths turns on the death fixtures' extras (runInputs): the queued
	// inputs are run, and state lines show gold, experience, worn items and
	// the kill and death books.
	deaths  bool
	pending []events.Input
```

- [ ] **Step 4: Masked room IDs in the event notes.** In `newCombatWorld`, the `AggroChanged` note
  `note(fmt.Sprintf("AggroChanged %s room=%d", w.name(ev.ActorId), ev.RoomId))` becomes
  `note(fmt.Sprintf("AggroChanged %s room=%s", w.name(ev.ActorId), roomName(ev.RoomId)))`, and the
  `PlayerDrop` note `note(fmt.Sprintf("PlayerDrop %s room=%d", w.user(ev.UserId), ev.RoomId))` becomes
  `note(fmt.Sprintf("PlayerDrop %s room=%s", w.user(ev.UserId), roomName(ev.RoomId)))`.

  Then a `MobItemDrop` note (an NPC's drop; no listener in the game handles it, so only a record can
  show it is still queued). After the `PlayerDeath` listener (the last in the `append`, `:135-139`)
  add:

```go
		events.RegisterListener(events.MobItemDrop{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.MobItemDrop)
			note(fmt.Sprintf("MobItemDrop mob=%d room=%s zone=%s item=%d", ev.MobId, roomName(ev.RoomId), ev.Zone, ev.ItemId))
			return events.Continue
		}),
```

  and in the cleanup's `zeros` list (`:142-143`) add `events.MobItemDrop{}` after `events.PlayerDeath{}`
  (the list and the listeners must stay in the same order). No existing golden queues a
  `MobItemDrop` (today only a run `suicide` does, and no existing fixture runs one), so the 9 stay
  byte-identical.

- [ ] **Step 5: `state` and `writeState`.** In `state` (`:296-315`), the first line becomes
  `s := fmt.Sprintf("room=%s hp=%d/%d aggro=[%s] playerdamage=%s", roomName(c.RoomId), c.Health, c.HealthMax.Value, w.aggro(c), w.damage(c))`,
  and before `return s` add:

```go
	if w.deaths {
		s += fmt.Sprintf(" gold=%d xp=%d worn=[%s] kd=%s", c.Gold, c.Experience, itemIds(c.GetAllWornItems()), kd(c.KD))
	}
```

  In `writeState` (`:317-341`), the rooms loop (from `for _, rid := range []int{goldenRoomA, goldenRoomB} {`
  through its closing brace) becomes:

```go
	for _, rid := range []int{goldenRoomA, goldenRoomB} {
		r := rooms.LoadRoom(rid)
		if r == nil {
			continue
		}
		if len(r.Items) > 0 {
			held := []string{}
			for _, it := range r.Items {
				held = append(held, fmt.Sprint(it.ItemId))
			}
			fmt.Fprintf(&w.out, "  room %d floor: %s\n", rid, strings.Join(held, ","))
		}
		if r.Gold > 0 {
			fmt.Fprintf(&w.out, "  room %d floor gold: %d\n", rid, r.Gold)
		}
		for _, cp := range r.Corpses {
			fmt.Fprintf(&w.out, "  room %d corpse %s: items=[%s] gold=%d worn=[%s]\n",
				rid, cp.Character.Name, itemIds(cp.Items), cp.Gold, itemIds(cp.Character.GetAllWornItems()))
		}
	}
```

- [ ] **Step 6: `collect`, `start` and `rounds`.** In `collect` (`:167-182`), the line
  `text := strings.TrimSuffix(e.text, "\n")` (`:173`) becomes:

```go
		// A world notice ends in CRLF (the player-death notice, suicide.go:88-90);
		// the golden side drops its \r (compareGolden), so the record does too.
		text := strings.TrimSuffix(strings.TrimSuffix(e.text, "\n"), "\r")
```

  (The 9 stay byte-identical: their files hold no `\r` (`grep -c $'\r'` prints 0 for each) and they
  pass today, so no read they record ends in one; the multi-line count is taken after the trim and
  counts `\n` only. The reads goldens do not use `collect`.)

  In `start`, after `w.rec.take()` add `w.pending = nil`.
  `rounds` (`:366-388`, with its comment) becomes:

```go
// rounds runs combat rounds until no fixture actor has an aggro, at most max.
func (w *combatWorld) rounds(max int) {
	for i := 1; i <= max && w.fighting(); i++ {
		w.reads = map[int][]string{}
		w.events = nil
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		w.drainInputs()
		w.collect()
		w.record(fmt.Sprintf("round %d", i))
	}
}

// record writes what the step just collected under "== <title>": each
// player's reads, the events, then the state.
func (w *combatWorld) record(title string) {
	fmt.Fprintf(&w.out, "== %s\n", title)
	for _, u := range w.players {
		fmt.Fprintf(&w.out, "-- %s reads\n", u.Character.Name)
		for _, line := range w.reads[u.UserId] {
			fmt.Fprintf(&w.out, "  %s\n", line)
		}
	}
	fmt.Fprintf(&w.out, "-- events\n")
	for _, e := range w.events {
		fmt.Fprintf(&w.out, "  %s\n", e)
	}
	w.writeState()
}
```

- [ ] **Step 7: The death helpers.** Add after `check` (`:390-394`):

```go
// roomName is a room ID for the record. A private copy (a death moves the
// body into its own copy of the recovery room) reads copy-of-<room>: its ID
// depends on which copies earlier tests made.
func roomName(id int) string {
	if rooms.IsEphemeralRoomId(id) {
		return fmt.Sprintf("copy-of-%d", rooms.GetOriginalRoom(id))
	}
	return fmt.Sprint(id)
}

// itemIds lists the item IDs, comma separated.
func itemIds(list []items.Item) string {
	out := []string{}
	for _, it := range list {
		out = append(out, fmt.Sprint(it.ItemId))
	}
	return strings.Join(out, ",")
}

// kd is the kill and death books for the record: mob kills/mob deaths/PvP
// kills/PvP deaths/elite kills/elite deaths, then whom the PvP kills and
// deaths were against (userid:name=count, sorted).
func kd(k characters.KDStats) string {
	sum := func(m map[string]int) int {
		n := 0
		for _, v := range m {
			n += v
		}
		return n
	}
	counts := func(m map[string]int) string {
		keys := []string{}
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := []string{}
		for _, key := range keys {
			out = append(out, fmt.Sprintf("%s=%d", key, m[key]))
		}
		return strings.Join(out, ",")
	}
	return fmt.Sprintf("%d/%d/%d/%d/%d/%d pk=[%s] pd=[%s]", k.TotalKills, k.TotalDeaths, k.TotalPvpKills,
		k.TotalPvpDeaths, sum(k.EliteKills), sum(k.EliteDeaths), counts(k.PlayerKills), counts(k.PlayerDeaths))
}

// freeCopies frees every private copy of roomId. Each copy a death makes
// holds one of the 100 ephemeral chunks (rooms/ephemeral.go) until it is
// freed, and a chunk is freed only when no player is in it, so the players
// leave first; without this, repeated runs (-count=N) run out of chunks.
func freeCopies(roomId int) {
	for _, id := range rooms.FindEphemeralRoomIds(roomId) {
		if r := rooms.LoadRoom(id); r != nil {
			for _, userId := range r.GetPlayers() {
				r.RemovePlayer(userId)
			}
		}
		rooms.TryEphemeralCleanup(id)
	}
}

// trinket is an item with its own spec, like weapon: a ring, a gem.
func trinket(itemId int, name string, typ items.ItemType, subtype items.ItemSubType) items.Item {
	spec := items.ItemSpec{ItemId: itemId, Name: name, NameSimple: name, Type: typ, Subtype: subtype}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

// runInputs makes the record run what a step queues as input (events.Input)
// as a command, after recording it, as the game loop would. The death
// fixtures use it: until Phase 5 slice 3 a death is a queued "suicide". It
// also turns on the state lines' gold, experience, worn items and books. Call it
// before start.
func (w *combatWorld) runInputs() {
	w.t.Helper()
	w.deaths = true
	id := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		w.pending = append(w.pending, e.(events.Input))
		return events.Continue
	})
	w.t.Cleanup(func() { events.UnregisterListener(events.Input{}, id) })
}

// drainInputs runs the inputs runInputs collected, in order, each followed
// by the events it queues (which may queue more). Without runInputs nothing
// is collected and this does nothing.
func (w *combatWorld) drainInputs() {
	for len(w.pending) > 0 {
		in := w.pending[0]
		w.pending = w.pending[1:]
		w.rec.note(fmt.Sprintf("Ran %s: %q", w.name(in.ActorId), in.InputText))
		verb, rest, _ := strings.Cut(in.InputText, " ")
		if _, err := usercommands.TryCommand(verb, rest, in.ActorId, in.Flags); err != nil {
			w.rec.note(fmt.Sprintf("(error: %s)", bodyIdPattern.ReplaceAllString(err.Error(), "<id>")))
		}
		events.ProcessEvents()
	}
}

// heal runs the regen pass (AutoHeal) as on rounds 3, 6, 9 and so on (the
// old pass acts only on every third round), recording each as "== heal N",
// until done reports true, at most max times. done is asked before each
// pass, so a fixture whose death needs no pass records none.
func (w *combatWorld) heal(max int, done func() bool) {
	for i := 1; i <= max && !done(); i++ {
		w.reads = map[int][]string{}
		w.events = nil
		AutoHeal(events.NewRound{RoundNumber: uint64(3 * i)})
		events.ProcessEvents()
		w.drainInputs()
		w.collect()
		w.record(fmt.Sprintf("heal %d", i))
	}
}

// recoveryRoom makes room C the death recovery room for this fixture. A
// death moves the body into a private copy made from the room's template
// file, so the template is saved into a temporary data folder; the repo's
// data folder is never written (a level up saves the account there too).
// Every copy of the room a death made is freed afterwards (freeCopies).
// Call it before start.
func (w *combatWorld) recoveryRoom() {
	w.t.Helper()
	c := rooms.NewEmptyRoom()
	c.RoomId, c.Zone, c.Title = goldenRoomC, "goldenzone", "Room C"
	rooms.AddTestRoom(c)
	w.t.Cleanup(func() { rooms.RemoveTestRoom(goldenRoomC) })
	w.t.Cleanup(func() { freeCopies(goldenRoomC) })
	tmp := w.t.TempDir()
	setConfig(w.t,
		map[string]any{"FilePaths.DataFiles": tmp, "SpecialRooms.DeathRecoveryRoom": goldenRoomC},
		map[string]any{
			"FilePaths.DataFiles":            configs.GetFilePathsConfig().DataFiles.String(),
			"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		})
	require.Equal(w.t, goldenRoomC, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "the override took")
	require.NoError(w.t, os.MkdirAll(filepath.Join(tmp, "rooms", "goldenzone"), 0o755))
	require.NoError(w.t, rooms.SaveRoomTemplate(*rooms.LoadRoom(goldenRoomC)))
}

// deathConfig sets today's death rules for one fixture with settings that
// make each rule show (the test binary loads no config.yaml, so all are
// zero otherwise): corpses on, drops on the floor beside them (CorpseItems
// off, as shipped), half of a player's gear dropped by chance, a 10% XP
// penalty, and levels up to 1 protected.
func deathConfig(t *testing.T) {
	t.Helper()
	gp := configs.GetGamePlayConfig()
	setConfig(t,
		map[string]any{
			"GamePlay.Death.CorpsesEnabled":      true,
			"GamePlay.Death.CorpseItems":         false,
			"GamePlay.Death.EquipmentDropChance": 0.5,
			"GamePlay.Death.XPPenalty":           "10%",
			"GamePlay.Death.ProtectionLevels":    1,
		},
		map[string]any{
			"GamePlay.Death.CorpsesEnabled":      bool(gp.Death.CorpsesEnabled),
			"GamePlay.Death.CorpseItems":         bool(gp.Death.CorpseItems),
			"GamePlay.Death.EquipmentDropChance": float64(gp.Death.EquipmentDropChance),
			"GamePlay.Death.XPPenalty":           string(gp.Death.XPPenalty),
			"GamePlay.Death.ProtectionLevels":    int(gp.Death.ProtectionLevels),
		})
	require.True(t, bool(configs.GetGamePlayConfig().Death.CorpsesEnabled), "the override took")
	require.Equal(t, "10%", string(configs.GetGamePlayConfig().Death.XPPenalty), "the override took")
}
```

  Before relying on `trinket`: `items.Ring`, `items.Gemstone`, `items.Wearable`, `items.Mundane` are
  `itemspec.go:106`, `:121`, `:127`, `:132`. Before relying on `setConfig` and `bodyIdPattern`: both
  are in `reads_golden_test.go` (same package, `:566`, `:421`). Before relying on `freeCopies`:
  `TryEphemeralCleanup` (`ephemeral.go:216-262`) frees the chunk holding the given copy only when
  `PlayerCt()` is 0 for every room in it; `PlayerCt` and `GetPlayers` count bodies an account drives
  (`rooms.go:726-740`), so the players are removed first while their actors still exist (this cleanup
  is registered after `newCombatWorld`'s `actors.Reset`, so it runs before it). `kd` uses `sort`,
  which the file already imports (`:18`, for `damage`).

- [ ] **Step 8: The three fixtures.** Add after `TestCombatGoldenHiddenAttacker` (`:512-523`):

```go
// A player kills an elite NPC that carries an item and gold and wears two
// (Phase 5 slice 3, recorded on the death before it moved): XP with the
// elite bonus, alignment, the kill on the books, drops by the NPC's
// ItemDropChance (MobItemDrop each), corpse, and the NPC's ending (it leaves
// the room; its slot's ending is written to the endings ledger).
func TestCombatGoldenDeathNPC(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: a level up saves the account
	deathConfig(t)
	w.runInputs()
	dataFolder := configs.GetFilePathsConfig().DataFiles.String()
	t.Cleanup(func() { // registered after recoveryRoom's cleanups, so it runs first, in the temporary folder
		require.NoError(t, os.RemoveAll(filepath.Join(dataFolder, "rooms.instances")))
		require.NoError(t, rooms.LoadEndings()) // no ledger file: the ledger in memory is empty again
	})
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	alice.Character.Alignment = 19 // neutral; killing the evil rat shifts her to lawful
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "golden-rats-death")
	rat.ItemDropChance = 50
	rat.IsElite = true // the elite XP bonus and the elite kill
	rat.SpawnSlot = 1  // a slot NPC: rooms.NPCEnded writes its ending to the ledger
	rat.Character.Alignment = -100
	rat.Character.Equipment.Weapon = weapon(93106, "testclub", items.Bludgeoning, 1, 4, 0)
	rat.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	rat.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	rat.Character.Gold = 30
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("death_npc")
	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the fixture must reach the rat's death")
	require.NotContains(t, rooms.LoadRoom(goldenRoomA).GetMobs(), rat.Character.Id, "the rat left the room")
	ledger, err := os.ReadFile(filepath.Join(dataFolder, "rooms.instances", "npcs.endings.yaml"))
	require.NoError(t, err, "the rat's ending was recorded (rooms.NPCEnded)")
	require.Contains(t, string(ledger), rat.Character.Id.String())
}

// An elite NPC kills a player above the protection band (Phase 5 slice 3,
// recorded on the death before it moved). Today she goes down, bleeds out
// over the regen passes, and her death drops gear by chance, costs XP,
// books an elite death and leaves a corpse; she wakes in the recovery room.
func TestCombatGoldenDeathPlayer(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	deathConfig(t)
	w.runInputs()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	alice.Character.Level, alice.Character.Experience = 2, 1000 // above ProtectionLevels 1
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	alice.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	alice.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.IsElite = true // her death is booked as an elite death
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	dead := func() bool { return rooms.IsEphemeralRoomId(alice.Character.RoomId) }
	w.heal(20, dead)
	w.check("death_player")
	require.True(t, dead(), "the fixture must reach Alice's death")
}

// A player kills a player (Phase 5 slice 3, recorded on the death before it
// moved): the kill credit and "killed by" notice, drops by chance, XP
// penalty, corpse, recovery room.
func TestCombatGoldenDeathPvP(t *testing.T) {
	w := newCombatWorld(t, 202)
	w.recoveryRoom()
	deathConfig(t)
	w.runInputs()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 15)
	bob.Character.Level, bob.Character.Experience = 2, 1000 // above ProtectionLevels 1
	bob.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	bob.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	dead := func() bool { return rooms.IsEphemeralRoomId(bob.Character.RoomId) }
	w.heal(20, dead)
	w.check("death_pvp")
	require.True(t, dead(), "the fixture must reach Bob's death")
}
```

- [ ] **Step 9: The 9 existing combat goldens and the reads goldens are unchanged.**
  `go test ./internal/hooks -run 'CombatGolden|TestReads|TestShootWait|TestKilledBy' -count=1` fails only
  in the three new tests (no golden file yet). Any other failure: stop and report its first
  differing line.

- [ ] **Step 10: Record the three, once.**
  `go test ./internal/hooks -run 'TestCombatGoldenDeath(NPC|Player|PvP)$' -count=1 -update`, then
  `git status --short -- internal/hooks/testdata/` prints exactly three `??` lines
  (`death_npc.txt`, `death_player.txt`, `death_pvp.txt`) and nothing else.

- [ ] **Step 11: Stable.** `go test ./internal/hooks -count=1` twice: `ok` both times (the full
  package, so other tests' recovery copies run first and the `copy-of` mask is exercised). Then the
  copies are freed: `go test ./internal/hooks -run 'TestCombatGoldenDeath(Player|PvP)$' -count=60`
  prints `ok` (120 deaths, more than the 100 ephemeral chunks). Probe: delete the
  `w.t.Cleanup(func() { freeCopies(goldenRoomC) })` line and rerun the same command: it fails once
  the chunks run out (the move then lands in room C itself, so `dead()` is false or the record reads
  `room=93003`); paste the line; restore.

- [ ] **Step 12: Read the three records in full** (Read tool, chunks of 400 lines) and list in the
  report where each shows today's rule: `death_npc`: `Input rat: "suicide"` then `Ran rat: "suicide"`,
  "rat has died.", Alice's XP line (with the elite bonus), her alignment line, the floor drops
  ("drops to the ground", the `room 93001 floor:` and `floor gold:` lines), one
  `MobItemDrop mob=93201 room=93001 zone=goldenzone item=N` note per dropped item, the corpse line,
  `MobDeath rat`, and Alice's `kd=1/0/0/0/1/0 pk=[] pd=[]` (one mob kill, one elite kill).
  `death_player`: the `PlayerDrop`, the `== heal N` bleed-out sections, `Ran Alice: "suicide"`,
  "has DIED!" (with no `\r` at the line end), the Remove and Drop lines, "You lost 100 experience
  points", the corpse, `room=copy-of-93003 hp=1/12`, `kd=0/1/0/0/0/1` (a mob death, an elite death).
  `death_pvp`: the same for Bob with "(killed by Alice)", Bob's `kd=0/0/0/1/0/0 pk=[] pd=[9301:Alice=1]`
  and Alice's `kd=0/0/1/0/0/0 pk=[9302:Bob=1] pd=[]`. If a `kd=` differs from these, report the
  record's value and why (read `suicide.go:56-109` and `suicide_npc.go:97-118`); do not change the
  fixture. If neither player fixture dropped any item by chance (all draws missed), say so: the
  controller may change that fixture's seed and re-record.

- [ ] **Step 13: Show they can fail**, one probe at a time:
  - delete `w.drainInputs()` from `rounds`: `TestCombatGoldenDeathNPC` fails (the rat never dies);
  - delete it from `heal`: `TestCombatGoldenDeathPlayer` fails (Alice never dies);
  - in `collect`, put back `text := strings.TrimSuffix(e.text, "\n")`: `TestCombatGoldenDeathPvP`
    fails at the "has DIED!" line (the record keeps the notice's `\r`, the golden side drops it);
  - in `writeState`, delete the corpse loop (`for _, cp := range r.Corpses {` through its closing
    brace): `TestCombatGoldenDeathPvP` fails at its first corpse line;
  - in `TestCombatGoldenDeathNPC`, delete `rat.SpawnSlot = 1`: the ledger assert fails (no ending is
    written for an NPC without a slot, which shows the assert reads what `NPCEnded` wrote).

  Paste the five failure lines.

- [ ] **Step 14:** Per-task check (the testdata line now lists the three files until they are
  committed). Commit `internal/hooks/combat_golden_test.go` and the three golden files:
  `Phase 5.3: death goldens on today's code (the harness runs queued inputs for them)`.
  `git show --stat HEAD` lists only those four files.

### Task 2: `rooms.IsDeathRecoveryRoom` in `CanPvp` and `CanAttack` (Sonnet)

**Files:** modify `internal/rooms/rooms.go`, `internal/rules/dikubasic/ruleset.go`,
`internal/rules/dikubasic/ruleset_test.go`; create `internal/rooms/recovery_test.go`.

- [ ] **Step 1: Write the failing tests.** Create `internal/rooms/recovery_test.go`:

```go
package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryAt makes roomId the death recovery room for one test and returns
// the ID of a private copy of it. The copy is recorded the way
// CreateEphemeralRoomIds records one (originalRoomIdLookups), at the last
// copy ID, which no test makes.
func recoveryAt(t *testing.T, roomId int) (copyId int) {
	t.Helper()
	old := int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["SpecialRooms.DeathRecoveryRoom"] = roomId
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["SpecialRooms.DeathRecoveryRoom"] = old
		require.NoError(t, configs.RestoreOverrides(back))
	})
	require.Equal(t, roomId, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom))

	copyId = ephemeralRoomIdMinimum + ephemeralChunksLimit*ephemeralChunkSize - 1
	originalRoomIdLookups[copyId] = roomId
	t.Cleanup(func() { delete(originalRoomIdLookups, copyId) })
	require.True(t, IsEphemeralRoomId(copyId))
	return copyId
}

func TestIsDeathRecoveryRoomKnowsTheRoomAndItsCopies(t *testing.T) {
	const recovery, other = 94101, 94102
	copyId := recoveryAt(t, recovery)
	assert.True(t, IsDeathRecoveryRoom(recovery), "the room")
	assert.True(t, IsDeathRecoveryRoom(copyId), "a private copy of it")
	assert.False(t, IsDeathRecoveryRoom(other), "another room")
}

// With no recovery room set (0, as in a test binary), no room is it: not
// room 0, and not an empty ephemeral room, whose original is recorded as 0
// (ephemeral.go:92).
func TestIsDeathRecoveryRoomIsNoRoomWhenNoneIsSet(t *testing.T) {
	emptyCopy := recoveryAt(t, 0) // a copy whose original reads 0
	assert.False(t, IsDeathRecoveryRoom(0), "room 0")
	assert.False(t, IsDeathRecoveryRoom(emptyCopy), "an empty ephemeral room")
	assert.False(t, IsDeathRecoveryRoom(94102), "another room")
}

// An attacker standing in its own copy of the recovery room may not fight
// (the raw ID compare missed the copy every respawn makes).
func TestCanPvpRefusesAnAttackerInACopyOfTheRecoveryRoom(t *testing.T) {
	const recovery, other = 94101, 94102
	copyId := recoveryAt(t, recovery)
	att, def := characters.New(), characters.New()
	att.RoomId, def.RoomId = copyId, other
	assert.EqualError(t, NewEmptyRoom().CanPvp(att, def), "Fighting is not allowed here.")
}
```

  In `internal/rules/dikubasic/ruleset_test.go` add `"path/filepath"` to the imports (the `rules`
  import comes with Task 3's tests; adding it now would not compile), and append:

```go
// setConfig sets config keys for one test and puts back the old values
// named in old (RestoreOverrides resets nothing it is not given; as
// setConfig in internal/hooks).
func setConfig(t *testing.T, set, old map[string]any) {
	t.Helper()
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	for k, v := range set {
		next[k] = v
	}
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
	})
}

// testRoom puts an empty room in memory for one test.
func testRoom(t *testing.T, id int, zone string) {
	t.Helper()
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = id, zone
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(id) })
}

// recoveryCopy makes roomId (a room testRoom put in memory, in zone) the
// death recovery room, saves its template into a temporary data folder and
// makes a private copy of it, as a death does. It returns the copy's ID.
func recoveryCopy(t *testing.T, roomId int, zone string) int {
	t.Helper()
	tmp := t.TempDir()
	setConfig(t,
		map[string]any{"FilePaths.DataFiles": tmp, "SpecialRooms.DeathRecoveryRoom": roomId},
		map[string]any{
			"FilePaths.DataFiles":            configs.GetFilePathsConfig().DataFiles.String(),
			"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		})
	require.Equal(t, roomId, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "the override took")
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "rooms", zone), 0o755))
	require.NoError(t, rooms.SaveRoomTemplate(*rooms.LoadRoom(roomId)))
	copies, err := rooms.CreateEphemeralRoomIds(roomId)
	require.NoError(t, err)
	id := copies[roomId]
	require.True(t, rooms.IsEphemeralRoomId(id))
	// Free the copy's ephemeral chunk (one of 100; ephemeral.go:14): removing
	// only the room would hold the chunk for the rest of the run.
	t.Cleanup(func() { rooms.TryEphemeralCleanup(id) })
	return id
}

// A target in a private copy of the recovery room is protected as one in
// the room itself. PvP is disabled, so a missed copy reads "PVP is
// disabled." instead.
func TestCanAttackRefusesATargetInACopyOfTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94003, 94004
	testRoom(t, here, "rulesetzone")
	testRoom(t, recovery, "rulesetzone")
	copyId := recoveryCopy(t, recovery, "rulesetzone")
	setConfig(t,
		map[string]any{"GamePlay.PVP.Enabled": "disabled", "GamePlay.PVP.MinimumLevel": 0},
		map[string]any{
			"GamePlay.PVP.Enabled":      string(configs.GetPVPConfig().Enabled),
			"GamePlay.PVP.MinimumLevel": int(configs.GetPVPConfig().MinimumLevel),
		})
	attacker, target := player(t, 9609), player(t, 9610)
	attacker.Character.RoomId, target.Character.RoomId = here, copyId
	assert.EqualError(t, New().CanAttack(attacker, target), "Fighting is not allowed here.")
}
```

- [ ] **Step 2: Run; they fail.** `go test ./internal/rooms -run 'DeathRecoveryRoom|CopyOfTheRecoveryRoom' -count=1`
  fails to compile (`undefined: IsDeathRecoveryRoom`). Comment out
  `TestIsDeathRecoveryRoomKnowsTheRoomAndItsCopies` and `TestIsDeathRecoveryRoomIsNoRoomWhenNoneIsSet`
  for a moment: `TestCanPvpRefusesAnAttackerInACopyOfTheRecoveryRoom` fails (nil error); restore them. `go test ./internal/rules/dikubasic -run CopyOfTheRecoveryRoom -count=1`
  fails with `PVP is disabled.`.

- [ ] **Step 3: The helper.** In `internal/rooms/rooms.go`, directly above
  `// Returns an error with a reason why they cannot PVP, or nil` (the comment of `CanPvp`, `:2060`):

```go
// IsDeathRecoveryRoom reports whether roomId is the death recovery room or
// a private copy of it: a death moves the body into its own copy
// (MoveToRoom), so a raw ID compare misses where the dead stand. With no
// recovery room set (0 or less) no room is it: 0 would match room 0 and
// every empty ephemeral room, whose original is recorded as 0.
func IsDeathRecoveryRoom(roomId int) bool {
	recovery := int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)
	return recovery > 0 && GetOriginalRoom(roomId) == recovery
}
```

  In `CanPvp` (`:2063`), `if attChar.RoomId == -1 || attChar.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) {`
  becomes `if attChar.RoomId == -1 || IsDeathRecoveryRoom(attChar.RoomId) {`.

- [ ] **Step 4: `CanAttack`.** In `internal/rules/dikubasic/ruleset.go` (`:91-95`), the comment and
  the check become:

```go
	// CanPvp checks only the attacker's room against the recovery room; a
	// safe room protects whoever stands in it, its private copies included.
	if rooms.IsDeathRecoveryRoom(defender.Character.RoomId) {
		return errors.New(`Fighting is not allowed here.`)
	}
```

  Drop the `configs` import only if the compiler says it is unused (`HatredRounds` still uses it).

- [ ] **Step 5: Run; they pass.** Both packages' commands from Step 2 print `ok`; so does
  `go test ./internal/rules/dikubasic -run CanAttack -count=1` (the existing recovery-room test).
  Probes: in `IsDeathRecoveryRoom` return `recovery > 0 && roomId == recovery` (both copy tests and
  the copy assert fail); return `GetOriginalRoom(roomId) == recovery` (dropping `recovery > 0 &&`:
  `TestIsDeathRecoveryRoomIsNoRoomWhenNoneIsSet` fails on room 0 and the empty copy); in `CanAttack`
  put back the raw compare `defender.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)`
  (`TestCanAttackRefusesATargetInACopyOfTheRecoveryRoom` fails). Paste the failure lines.

- [ ] **Step 6:** Per-task check (goldens unchanged). Commit `internal/rooms/rooms.go
  internal/rooms/recovery_test.go internal/rules/dikubasic/ruleset.go internal/rules/dikubasic/ruleset_test.go`:
  `Phase 5.3: rooms.IsDeathRecoveryRoom; CanPvp and CanAttack see the recovery room's copies`.

### Task 3: `Condition` and `Regen` behind the seam, with parity (Opus; the controller reads the diff line by line)

Every death check and regen decision asks the ruleset; behaviour is unchanged under `dikubasic`. The
bleed-out and the -10 `suicide` stay engine code until Task 6. Q3 applies.

**Files:** modify `internal/rules/rules.go`, `internal/rules/rules_test.go`,
`internal/rules/dikubasic/ruleset.go`, `internal/rules/dikubasic/ruleset_test.go`,
`internal/hooks/NewRound_DoCombat.go`, `internal/hooks/NewRound_AutoHeal.go`,
`internal/hooks/NewRound_ActorUpkeep.go`, `internal/hooks/Buff_ApplyBuffs.go`,
`internal/hooks/attack_engine_test.go`,
`internal/usercommands/usercommands.go`, `internal/usercommands/go.go`, `internal/usercommands/remove.go`,
`world_test.go` (repo root); create `internal/usercommands/remove_test.go`.

- [ ] **Step 1: `rules.go`.** After `AttackResult` (`:12-25`) add:

```go
// Condition is what the ruleset says an actor's body is.
type Condition int

const (
	Up     Condition = iota // can act
	Downed                  // cannot act, not dead
	Dead                    // the engine runs its death (death.Die)
)

// RegenResult is what one round regenerates for an actor. The engine applies
// Health and Mana (either may be negative: a bleed) and sends Messages.
type RegenResult struct {
	Health, Mana int
	Messages     []perception.Message
}

// Changes reports whether r changes health or mana. It lives here, outside
// the engine packages, so the regen pass compares no health with a number
// (TestEngineAsksTheRulesetWhoIsDead).
func (r RegenResult) Changes() bool { return r.Health != 0 || r.Mana != 0 }
```

  In the `Ruleset` comment, `Phase 5 slice 2 holds the fight; later` / `// slices add the round, death and display.`
  becomes `Phase 5 slice 2 holds the fight and slice 3 the` / `// round and death; later slices add display.`.
  After `CanAttack` in the interface add:

```go

	// Condition is whether a is up, downed or dead. The engine compares no
	// health against a number: it asks this.
	Condition(a *actors.Actor) Condition

	// Regen is what a regenerates this round. It changes no health or mana:
	// the engine applies the result.
	Regen(a *actors.Actor, round uint64) RegenResult
```

- [ ] **Step 2: `rules_test.go`.** After the `CanAttack` line of `fake` add:

```go
func (fake) Condition(a *actors.Actor) Condition               { return Up }
func (fake) Regen(a *actors.Actor, round uint64) RegenResult   { return RegenResult{} }
```

  and append:

```go
func TestRegenResultChangesWhenHealthOrManaDoes(t *testing.T) {
	assert.False(t, RegenResult{}.Changes())
	assert.False(t, RegenResult{Messages: []perception.Message{{ActorText: "a line"}}}.Changes(), "a line alone changes nothing")
	assert.True(t, RegenResult{Health: 1}.Changes())
	assert.True(t, RegenResult{Health: -1}.Changes(), "a bleed")
	assert.True(t, RegenResult{Mana: 2}.Changes())
}
```

  Run `go test ./internal/rules -run Changes -count=1`: `ok`. Probes: `r.Health != 0 || r.Mana != 0` to
  `r.Health > 0 || r.Mana != 0` (the bleed assert fails); drop `|| r.Mana != 0` (the mana assert
  fails). Paste the two lines.

- [ ] **Step 3: `dikubasic`.** Append to `ruleset.go`:

```go
// Condition: dead at 0 health or below, else up. The reference rules have
// no downed state (decision 7).
func (Ruleset) Condition(a *actors.Actor) rules.Condition {
	if a.Character.Health <= 0 {
		return rules.Dead
	}
	return rules.Up
}

// Regen is GoMud's AutoHeal, moved behind the seam unchanged: an account,
// every 3rd round, out of a fight, outside the death recovery room and not
// downed, regains HealthPerRound and ManaPerRound. NPCs do not regenerate
// (slice 4 gives every actor the one rule).
func (Ruleset) Regen(a *actors.Actor, round uint64) rules.RegenResult {
	if accountOf(a) == nil || round%3 != 0 || a.Character.Aggro != nil {
		return rules.RegenResult{}
	}
	if a.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) {
		return rules.RegenResult{}
	}
	if a.Character.Health <= 0 {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: a.Character.HealthPerRound(), Mana: a.Character.ManaPerRound()}
}
```

  (The raw-ID compare is today's, kept for parity: Q7i. Task 5 replaces it.)

- [ ] **Step 4: `dikubasic` unit tests.** Append to `ruleset_test.go` (imports: add
  `github.com/GoMudEngine/GoMud/internal/rules`):

```go
func TestConditionIsDeadAtZeroOrBelow(t *testing.T) {
	a := player(t, 9620)
	for health, want := range map[int]rules.Condition{1: rules.Up, 0: rules.Dead, -1: rules.Dead} {
		a.Character.Health = health
		assert.Equal(t, want, New().Condition(a), "health %d", health)
	}
}

// Parity with the old AutoHeal (NewRound_AutoHeal.go before slice 3).
func TestRegenIsTheOldAutoHeal(t *testing.T) {
	const here, recovery = 94005, 94006
	setConfig(t,
		map[string]any{"SpecialRooms.DeathRecoveryRoom": recovery},
		map[string]any{"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)})
	none := rules.RegenResult{}

	a := player(t, 9621)
	a.Character.RoomId, a.Character.Health = here, 5
	full := rules.RegenResult{Health: a.Character.HealthPerRound(), Mana: a.Character.ManaPerRound()}
	require.NotZero(t, full.Health)
	assert.Equal(t, full, New().Regen(a, 3), "every 3rd round")
	assert.Equal(t, full, New().Regen(a, 6))
	assert.Equal(t, none, New().Regen(a, 4), "only every 3rd round")
	assert.Equal(t, none, New().Regen(a, 5))

	a.Character.Aggro = &characters.Aggro{Type: characters.DefaultAttack}
	assert.Equal(t, none, New().Regen(a, 3), "never in a fight")
	a.Character.Aggro = nil

	a.Character.RoomId = recovery
	assert.Equal(t, none, New().Regen(a, 3), "not in the recovery room")
	a.Character.RoomId = here

	a.Character.Health = 0
	assert.Equal(t, none, New().Regen(a, 3), "never while down")

	n := npc(t, 0, 0)
	n.Character.RoomId, n.Character.Health = here, 5
	assert.Equal(t, none, New().Regen(n, 3), "accounts only")
}
```

  Run `go test ./internal/rules/dikubasic -run 'Condition|Regen' -count=1`; both pass. Probes, one at a
  time: in `Condition` `<= 0` to `< 0`; in `Regen` `round%3` to `round%2`; delete `|| a.Character.Aggro != nil`;
  delete `accountOf(a) == nil ||`; delete the recovery-room `if`; delete the `Health <= 0` `if`. Each
  makes a named assert fail; paste the six lines.

- [ ] **Step 5: `NewRound_DoCombat.go`.** Five edits, no other change:

  (a) `:169` `if d.Character.Health < 1 {` becomes `if rules.Get().Condition(d) != rules.Up {`.

  (b) `:201` `healthBefore := d.Character.Health` and the blank line after it: delete.

  (c) `:215-217` (the comment `// Combat damage downed a player: remember who did it (plan Decision 7),` /
  `// for the death that follows.` and `if defAcct != nil && healthBefore >= 1 && d.Character.Health < 1 {`)
  becomes:

```go
	// Combat took a player out of Up: remember who did it (plan Decision 7),
	// for the death that follows. The defender was up before this attack:
	// the rage check above returns for one who is not.
	if defAcct != nil && rules.Get().Condition(d) != rules.Up {
```

  Parity, for the controller's read of (b) and (c): today's `healthBefore >= 1` is always true at
  `:217`, because `:169` returned for a defender under 1 and nothing between `:169` and `attack`
  (`:203`) changes the defender's health (the weapon search is the attacker's; the wait and hidden
  branches return). Keeping a `wasUp` variable would keep a check no test could show matters.

  (d) `:269` `if a.Character.Health <= 0 || d.Character.Health <= 0 {` becomes
  `if rules.Get().Condition(a) != rules.Up || rules.Get().Condition(d) != rules.Up {`.

  (e) `handleAffected` (`:474-520`, with its comment) becomes:

```go
// handleAffected settles everyone the round touched, actors with an account
// first, then NPCs, each once: the ruleset says who is down or dead.
func handleAffected(affected []ids.ID) {

	handled := map[ids.ID]struct{}{}
	npcs := []*mobs.Mob{}

	for _, id := range affected {
		if _, ok := handled[id]; ok {
			continue
		}
		handled[id] = struct{}{}

		a := actors.Get(id)
		if a == nil {
			continue
		}

		if user := accountOf(a); user != nil {

			if user.Character.Health <= -10 {

				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

			} else if rules.Get().Condition(a) != rules.Up {

				events.AddToQueue(events.PlayerDrop{UserId: user.UserId, RoomId: user.Character.RoomId})

			}

			continue
		}

		if mob := mobs.GetInstance(id); mob != nil {
			npcs = append(npcs, mob)
		}
	}

	for _, mob := range npcs {
		if a := actors.Get(mob.Character.Id); a != nil && rules.Get().Condition(a) == rules.Dead {

			mob.Command(`suicide`)

		}
	}

}
```

  (The `-10` literal goes in Task 6 with the bleed-out.)

- [ ] **Step 6: `NewRound_AutoHeal.go`** becomes, in full:

```go
package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
)

// AutoHeal is the engine's regen pass (Phase 5 slice 3): every round, for
// every actor with a body, the ruleset says what it regenerates; the engine
// applies it, sends the ruleset's lines and reports an account's changed
// vitals. The bleed-out stays engine code until slice 3's death commit
// removes it.
func AutoHeal(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	for _, a := range fighters() {
		user := accountOf(a)

		healthStart := a.Character.Health
		manaStart := a.Character.Mana

		if user != nil && bleeds(a, evt.RoundNumber) {

			if a.Character.Health <= -10 {

				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

			} else {
				a.Character.Health--
				if room := rooms.LoadRoom(a.Character.RoomId); room != nil {
					messaging.Send(perception.Message{
						Kind:         perception.WorldEvent,
						Actor:        a.Character.Id,
						Senses:       perception.Sight | perception.Sound,
						Range:        perception.Room,
						ActorText:    `<ansi fg="red">you are bleeding out!</ansi>`,
						ObserverText: `{actor} is <ansi fg="red">bleeding out</ansi>! Somebody needs to provide aid!`,
					})
				} else {
					messaging.Tell(a.Character.Id, `<ansi fg="red">you are bleeding out!</ansi>`)
				}
			}

		} else {

			r := rules.Get().Regen(a, evt.RoundNumber)
			if r.Changes() {
				a.Character.Heal(r.Health, r.Mana)
			}
			for _, m := range r.Messages {
				messaging.Send(m)
			}
		}

		if user == nil {
			continue
		}

		// If it has changed, send an update
		if a.Character.Health != healthStart || a.Character.Mana != manaStart {

			// Trigger a redraw, but only if the users prompt has changed.
			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId, OnlyIfChanged: true}, 100)

			events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})

		}
	}

	return events.Continue
}

// bleeds reports whether a downed account bleeds this round: the old
// AutoHeal's gates (every 3rd round, out of a fight, outside the death
// recovery room by raw ID) and a body the ruleset does not call up.
func bleeds(a *actors.Actor, round uint64) bool {
	return round%3 == 0 &&
		a.Character.Aggro == nil &&
		a.Character.RoomId != int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) &&
		rules.Get().Condition(a) != rules.Up
}
```

  Parity, for the controller's read: `fighters()` (`NewRound_DoCombat.go:48-61`) walks the online
  user IDs in the order `AutoHeal` did, then NPCs, which `dikubasic.Regen` leaves alone. An account
  either bleeds (same gates as before) or gets `Regen`, which is non-zero exactly when the old code
  called `Heal` with health above 0. `Heal(0, 0)` is no longer called when regen is zero; it changed
  nothing then either (health is already at or under the maximum, and the killer record is clear
  while health is up).

- [ ] **Step 7: Upkeep and the buff trigger.** In `NewRound_ActorUpkeep.go` (`:62`)
  `if mob != nil && char.Health <= 0 {` becomes `if mob != nil && rules.Get().Condition(a) == rules.Dead {`;
  add the `rules` import. In `Buff_ApplyBuffs.go` (`:64`) `if isMob && targetChar.Health <= 0 {` becomes
  `if a := actors.Get(evt.ActorId); isMob && a != nil && rules.Get().Condition(a) == rules.Dead {`; add
  the `rules` import.

- [ ] **Step 8: `usercommands` (Q3).** `usercommands.go:281`
  `userDisabled = actor.Character.IsDisabled()` becomes
  `userDisabled = rules.Get().Condition(actor) != rules.Up`; `go.go:46`
  `if actor.Character.IsDisabled() {` becomes `if rules.Get().Condition(actor) != rules.Up {`;
  `remove.go:51` `if matchItem.IsCursed() && actor.Character.Health > 0 {` becomes
  `if matchItem.IsCursed() && rules.Get().Condition(actor) == rules.Up {`. Add the
  `github.com/GoMudEngine/GoMud/internal/rules` import to each of the three files (none has it:
  `grep -c 'internal/rules"'` printed 0 for each).

  `TryCommand` now calls `rules.Get()`, which panics with no ruleset set. The root package's
  `TestMain` (`world_test.go:66-86`) sets none, and `TestNPCInputReachesTheSharedRegistry`
  (`world_test.go:132`) reaches `TryCommand` through `HandleInputEvents`. In `world_test.go`, after
  `actors.Init()` (`:68`) add `rules.Set(dikubasic.New()) // TryCommand asks the ruleset whether the actor is up`,
  and add `github.com/GoMudEngine/GoMud/internal/rules` and
  `github.com/GoMudEngine/GoMud/internal/rules/dikubasic` to its imports, as the `hooks` and
  `usercommands` test mains do (`actionpoints_test.go:28`, `usercommands_test.go:31`). The ruleset
  guard `TestOnlyBootNamesTheRuleset` reads no `_test.go` file (`goFiles`, `guards_test.go:31`), so
  this import needs no allowlist entry. Check: `go test . -count=1` prints `ok`; with the `rules.Set`
  line deleted it panics with `rules: no ruleset is set` (paste the line, restore).

- [ ] **Step 9: Tests for the rewired checks.** Two of the rewired checks are reached by no test
  today. Both tests pass on today's code too (the rewiring keeps behaviour), so their probes are what
  shows them working.

  (a) The rage check. Append to `internal/hooks/attack_engine_test.go` (it already imports `actors`,
  `characters`, `rules` and `require`; add `"slices"`; `oneShot` is in `ruleset_stub_test.go`):

```go
// "Your rage subsides" is the ruleset's call: an attacker whose target is no
// longer up when its turn comes stops. Under oneShot the first of the two to
// swing takes the rat out before the other's turn. Which goes first is map
// order (fighters, NewRound_DoCombat.go:47-60), so the test finds out.
func TestRageSubsidesWhenTheTargetIsNotUp(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: the kill's XP may level the killer up and save her
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 30, 30, 20, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "rage-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	bob.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(oneShot{rules.Get()}))
	w.start()
	w.rounds(1)

	first, second := alice, bob
	if !slices.Contains(w.reads[alice.UserId], "STUB HIT") {
		first, second = bob, alice
	}
	require.Contains(t, w.reads[first.UserId], "STUB HIT", "one of them swung first")
	require.Contains(t, w.reads[second.UserId], "Your rage subsides.")
	require.NotContains(t, w.reads[second.UserId], "STUB HIT", "the other did not swing at a body that is not up")
}
```

  Probe: at `:169` replace `rules.Get().Condition(d) != rules.Up` with `false` (Bob swings: the rage
  assert fails). Paste the line, restore.

  (b) The cursed-item gate. Create `internal/usercommands/remove_test.go`:

```go
package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A curse holds only a body the ruleset calls up; a body that is not up can
// shed a cursed item (the old player death removed worn gear this way).
func TestCursedItemComesOffOnlyABodyThatIsNotUp(t *testing.T) {
	_, player, room := goSetup(t)
	player.Character.Equipment.Ring = testItem(92720, "testring", items.ItemSpec{Type: items.Ring, Subtype: items.Wearable, Cursed: true})

	told := playerTexts(t, player, func() {
		_, err := Remove("testring", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, told, 1)
	assert.Contains(t, told[0], "CURSED!")
	assert.Equal(t, 92720, player.Character.Equipment.Ring.ItemId, "up: the curse holds")

	player.Character.Health = 0 // dead under the reference ruleset
	_, err := Remove("testring", player, room, 0)
	require.NoError(t, err)
	assert.Zero(t, player.Character.Equipment.Ring.ItemId, "not up: it comes off")
}
```

  Check before relying on it: the curse refusal needs `GetSkillLevel("enchant")` under
  `skills.MaxSkillLevel("enchant")`, which is `DefaultMaxLevel` 4 with no skill data loaded
  (`skills/skill.go:22`, `:116-121`); a new character has no skills. Probes: at `remove.go:51` replace
  `rules.Get().Condition(actor) == rules.Up` with `true` (the not-up assert fails), then with `false`
  (the up assert fails). Paste the two lines, restore.

- [ ] **Step 10: Prove the engine packages ask the ruleset.**
  `grep -nE '\.Health *(<=|>=|<|>|==|!=) *-?[0-9]' internal/hooks/*.go internal/usercommands/*.go | grep -v _test.go`
  prints exactly two lines, the `<= -10` checks in `NewRound_DoCombat.go` (`handleAffected`) and
  `NewRound_AutoHeal.go` (both go in Task 6). The regen pass's "anything to apply" check is
  `r.Changes()`; written as `r.Health != 0 || r.Mana != 0` it would be a third line, and Task 8's guard
  would fail on it. Control: the same pattern on `internal/characters/character.go` prints hits.

- [ ] **Step 11:** Per-task check. **All goldens unchanged** (the three death fixtures run the bleed
  through the new pass). If one changes, stop and report the first differing line
  (`go test ./internal/hooks -run 'CombatGolden|TestReads' -count=1 -v 2>&1 | grep -B2 -A3 differs`).
  Commit the fourteen files in **Files**:
  `Phase 5.3: Condition and Regen behind the seam; round checks, upkeep, buff trigger and AutoHeal ask the ruleset (goldens unchanged)`.

### Task 4: The `Heal` fix (Sonnet)

**Files:** modify `internal/characters/character.go`, `internal/characters/character_test.go`.

- [ ] **Step 1: Write the failing test.** Append to `character_test.go`:

```go
// Heal's second argument is mana: it restores mana by that amount, not by
// the health amount (slice 3 fix).
func TestHealRestoresManaByTheManaAmount(t *testing.T) {
	c := New()
	c.HealthMax.Value, c.ManaMax.Value = 20, 20
	c.Health, c.Mana = 10, 10
	dh, dm := c.Heal(1, 5)
	assert.Equal(t, 11, c.Health)
	assert.Equal(t, 15, c.Mana)
	assert.Equal(t, 1, dh)
	assert.Equal(t, 5, dm)
}
```

- [ ] **Step 2: Run; it fails** (`Mana` 11, not 15). `go test ./internal/characters -run TestHealRestoresMana -count=1`.
- [ ] **Step 3:** `character.go:996` `c.Mana += hp` becomes `c.Mana += mana`.
- [ ] **Step 4: Run; it passes.** The undo of Step 3 is the probe (Step 2 showed the failure).
- [ ] **Step 5:** Per-task check. Goldens unchanged: the only regen in a golden is `death_pvp`'s heal
  passes, where Alice's `HealthPerRound` and `ManaPerRound` are both 1 (no race in
  `_datafiles/world/empty/races/` carries a recovery statmod: `grep -il recovery` prints nothing). If a
  golden changes, stop and report. Commit both files: `Phase 5.3: Heal restores mana by its mana argument`.

### Task 5: Recovery-room fast regen (Sonnet)

Spec ruling 2: upstream's deleted Death Recovery buff, as a reference-ruleset regen rule. The
buff's inert data file and room 75's mutator that applies it go in the same commit (spec, plan
review round 1).

**Files:** modify `internal/rules/dikubasic/ruleset.go`, `internal/rules/dikubasic/ruleset_test.go`,
`_datafiles/world/empty/rooms/shadow_realm/75.yaml`; `git rm _datafiles/world/empty/buffs/24-death_recovery.yaml
_datafiles/world/empty/mutators/death_recovery.yaml`.

- [ ] **Step 1: Write the failing tests.** In `TestRegenIsTheOldAutoHeal` delete the three lines from
  `a.Character.RoomId = recovery` through `a.Character.RoomId = here` and the blank line after them
  (the recovery room now has its own rule, tested below), and rename the test
  `TestRegenOutsideTheRecoveryRoomIsTheOldAutoHeal`. Append:

```go
// reseed puts the shared random source back on a clock seed after a test.
func reseed(t *testing.T) {
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) })
}

// The recovery room (and each private copy of it) heals 1d10 health and
// 1d10 mana every round, health first, drawn from the shared source.
func TestRegenInTheRecoveryRoomRollsD10HealthAndMana(t *testing.T) {
	const recovery = 94007
	testRoom(t, recovery, "regenzone")
	copyId := recoveryCopy(t, recovery, "regenzone")
	a := player(t, 9622)
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	reseed(t)
	for _, roomId := range []int{recovery, copyId} {
		a.Character.RoomId = roomId
		for seed := int64(1); seed <= 200; seed++ {
			a.Character.Health, a.Character.Mana = 1, 0
			util.SeedRandForTest(seed)
			h := util.Rand(10) + 1
			m := util.Rand(10) + 1
			next := util.Rand(1000)

			util.SeedRandForTest(seed)
			got := New().Regen(a, uint64(seed)) // every round, not only every 3rd
			require.Equal(t, h, got.Health, "room %d seed %d", roomId, seed)
			require.Equal(t, m, got.Mana, "room %d seed %d", roomId, seed)
			require.Equal(t, next, util.Rand(1000), "room %d seed %d: two draws", roomId, seed)
			require.Equal(t, 1, a.Character.Health, "the ruleset changes no health")

			// At full health and mana nothing heals, and the two dice are
			// still drawn, as upstream's buff drew them (Q7h).
			a.Character.Health, a.Character.Mana = 100, 100
			util.SeedRandForTest(seed)
			require.Equal(t, rules.RegenResult{}, New().Regen(a, uint64(seed)), "room %d seed %d: full", roomId, seed)
			require.Equal(t, next, util.Rand(1000), "room %d seed %d: full, still two draws", roomId, seed)
		}
	}
}

// The amounts are what Heal will apply (capped at the maximums), and the
// lines are upstream's buff's: one for the actor, one for the room.
func TestRegenInTheRecoveryRoomSaysWhatItHealed(t *testing.T) {
	const recovery = 94008
	testRoom(t, recovery, "regenzone")
	recoveryCopy(t, recovery, "regenzone")
	a := player(t, 9623)
	a.Character.RoomId = recovery
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	reseed(t)
	util.SeedRandForTest(1)

	a.Character.Health, a.Character.Mana = 1, 0
	got := New().Regen(a, 1)
	require.Len(t, got.Messages, 2)
	assert.Equal(t, perception.Interface, got.Messages[0].Kind)
	assert.Equal(t, perception.ActorOnly, got.Messages[0].Audience)
	assert.Equal(t, a.Id, got.Messages[0].Actor)
	assert.Equal(t, fmt.Sprintf(`The shadow realm heals you for <ansi fg="healing">%d damage</ansi> and restores <ansi fg="mana-100">%d mana</ansi>!`, got.Health, got.Mana), got.Messages[0].ActorText)
	assert.Equal(t, perception.WorldEvent, got.Messages[1].Kind)
	assert.Equal(t, a.Id, got.Messages[1].Actor)
	assert.Equal(t, `{actor} is recovering from a recent death.`, got.Messages[1].ObserverText)

	// Capped at the maximums: one point short heals exactly one point, and
	// says so, whatever the dice (over 20 seeds, most of the d10s roll above 1).
	for seed := int64(1); seed <= 20; seed++ {
		util.SeedRandForTest(seed)
		a.Character.Health, a.Character.Mana = 99, 100
		got = New().Regen(a, 2)
		require.Equal(t, 1, got.Health, "seed %d: capped at the health maximum", seed)
		require.Zero(t, got.Mana, "seed %d: mana is full", seed)
		require.NotEmpty(t, got.Messages)
		require.Equal(t, `The shadow realm heals you for <ansi fg="healing">1 damage</ansi>!`, got.Messages[0].ActorText)

		util.SeedRandForTest(seed)
		a.Character.Health, a.Character.Mana = 100, 99
		got = New().Regen(a, 3)
		require.Zero(t, got.Health, "seed %d: health is full", seed)
		require.Equal(t, 1, got.Mana, "seed %d: capped at the mana maximum", seed)
		require.NotEmpty(t, got.Messages)
		require.Equal(t, `The shadow realm restores <ansi fg="mana-100">1 mana</ansi>!`, got.Messages[0].ActorText)
	}

	a.Character.Health, a.Character.Mana = 100, 100
	got = New().Regen(a, 4)
	assert.Equal(t, rules.RegenResult{}, got, "nothing to heal: no lines")
}
```

  Add `"fmt"` and `github.com/GoMudEngine/GoMud/internal/perception` to the imports.

- [ ] **Step 2: Run; they fail.** `go test ./internal/rules/dikubasic -run Regen -count=1`: both new tests
  fail (zero regen); the renamed parity test passes.

- [ ] **Step 3: The rule.** In `ruleset.go`, `Regen`'s comment and the raw recovery-room `if` become:

```go
// Regen in the death recovery room (or a private copy of it) is fast:
// recoveryRegen. Elsewhere it is GoMud's AutoHeal, moved behind the seam
// unchanged: an account, every 3rd round, out of a fight and not downed,
// regains HealthPerRound and ManaPerRound. NPCs do not regenerate there
// (slice 4 gives every actor the one rule).
func (Ruleset) Regen(a *actors.Actor, round uint64) rules.RegenResult {
	if rooms.IsDeathRecoveryRoom(a.Character.RoomId) {
		return recoveryRegen(a)
	}
	if accountOf(a) == nil || round%3 != 0 || a.Character.Aggro != nil {
		return rules.RegenResult{}
	}
	if a.Character.Health <= 0 {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: a.Character.HealthPerRound(), Mana: a.Character.ManaPerRound()}
}

// recoveryRegen is the recovery room's fast regen (spec ruling 2): upstream's
// deleted Death Recovery buff (24-death_recovery.js), 1d10 health and 1d10
// mana every round, health drawn first, with its lines. The amounts are what
// Heal will apply (capped at the maximums), so the lines say what happened.
func recoveryRegen(a *actors.Actor) rules.RegenResult {
	c := a.Character
	health := min(util.RollDice(1, 10), max(c.HealthMax.Value-c.Health, 0))
	mana := min(util.RollDice(1, 10), max(c.ManaMax.Value-c.Mana, 0))

	text := ``
	switch {
	case health > 0 && mana > 0:
		text = fmt.Sprintf(`The shadow realm heals you for <ansi fg="healing">%d damage</ansi> and restores <ansi fg="mana-100">%d mana</ansi>!`, health, mana)
	case health > 0:
		text = fmt.Sprintf(`The shadow realm heals you for <ansi fg="healing">%d damage</ansi>!`, health)
	case mana > 0:
		text = fmt.Sprintf(`The shadow realm restores <ansi fg="mana-100">%d mana</ansi>!`, mana)
	}
	if text == `` {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: health, Mana: mana, Messages: []perception.Message{
		{Kind: perception.Interface, Actor: a.Id, Audience: perception.ActorOnly, ActorText: text},
		{
			Kind:         perception.WorldEvent,
			Actor:        a.Id,
			Senses:       perception.Sight,
			Range:        perception.Room,
			ObserverText: `{actor} is recovering from a recent death.`,
		},
	}}
}
```

  Add `"fmt"` to `ruleset.go`'s imports. (`util.RollDice(1, 10)` is `util.Rand(10) + 1`,
  `util.go:543-565`; the two statements run in order, health first.)

- [ ] **Step 4: Run; they pass.** Probes, each compiled and run on its own:
  - swap the two `RollDice` lines (the d10 test fails: health and mana swapped);
  - in `IsDeathRecoveryRoom`'s place use the raw compare
    `a.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)` (the copy half of
    the d10 test fails);
  - drop the health cap only: `health := util.RollDice(1, 10)` (the mana line still uses `c`, so it
    compiles; the health-cap assert fails on the first seed whose d10 is above 1);
  - drop the mana cap only: `mana := util.RollDice(1, 10)` (the health line still uses `c`; the
    mana-cap assert fails the same way);
  - add `if c.Health >= c.HealthMax.Value && c.Mana >= c.ManaMax.Value { return rules.RegenResult{} }`
    as the first statement after `c := a.Character` (the full-health "still two draws" assert fails);
  - change `{actor} is recovering` to `{actor} recovers` (the line test fails).

  Paste the six lines.

- [ ] **Step 5: Upstream's inert buff and its mutator go** (spec, plan review round 1). Buff 24
  (`_datafiles/world/empty/buffs/24-death_recovery.yaml`) is a name, a description and a trigger rate
  with no flags or stat mods: since its script was deleted it does nothing. The mutator
  `death-recovery` (`_datafiles/world/empty/mutators/death_recovery.yaml`) has one active line besides
  its ID, `playerbuffids: [24] # death-recovery` (`:16`); everything else is commented out. With buff 24
  gone it would do nothing at all, so it goes too, and room 75 stops naming it.
  - `git rm _datafiles/world/empty/buffs/24-death_recovery.yaml _datafiles/world/empty/mutators/death_recovery.yaml`.
  - `_datafiles/world/empty/rooms/shadow_realm/75.yaml` (CRLF, as the file): delete the two lines
    `mutators:` and `- mutatorid: death-recovery` (`:18-19`), so `exits:` (`:15-17`) is followed by
    `mapx: 0`. Check the line ends after the edit: `file _datafiles/world/empty/rooms/shadow_realm/75.yaml`
    still says CRLF, and `git diff -- _datafiles/world/empty/rooms/shadow_realm/75.yaml` shows exactly
    the two removed lines.
  - Prove nothing else names them: `grep -rn 'death-recovery\|death_recovery\|Death Recovery\|24-death' --exclude-dir=.git --exclude-dir=.claude .`
    prints only: the historical documents (`2026-11-06-phase-5-inventory.md`,
    `2026-11-06-phase-5-rulings.md`, `docs/completed/2026-10-26-phase-4-scripting-surface.md`,
    `docs/completed/2026-10-28-phase-4b-5-inventory.md`; left as written), the spec and this plan,
    `2026-11-06-phase-5-rules-seam-design.md:153` (its slice 6 sentence, removed in Task 10),
    `internal/usercommands/suicide.go:302` (a comment Task 6 deletes) and
    `_datafiles/html/admin/config-wizard.html:1398` (`'Death Recovery Room ID'`, the room setting, which
    stays). `_datafiles/html/admin/mutators-api.html:78`, `:83` use `[24]` as an example buff ID in an
    API sample; it names no file and stays. Control: the same grep with `pvp-enabled` prints hits.
  - The world still loads room 75: `go run ./cmd/smoke` passes (its death step looks at the
    recovery room); paste the summary line.

- [ ] **Step 6:** Per-task check. Goldens unchanged: the death fixtures stop their heal passes at the
  death, before anyone is in the recovery room; reads golden H never runs the regen pass. If a golden
  changes, stop and report. Commit the two Go files, `75.yaml` and the two removals (`git rm` staged
  them):
  `Phase 5.3: the recovery room heals 1d10 health and mana a round (upstream's Death Recovery, as a ruleset rule); its inert buff and mutator go`.

### Task 6: One death routine (Opus; the controller reads the diff and every golden hunk)

`internal/death`, `OnDeath`, XP and alignment into `dikubasic`, the deletions, and the goldens'
deliberate changes. Q1, Q2, Q4, Q5, Q6 and Q7 apply.

**Files:** create `internal/death/death.go`, `internal/rules/dikubasic/death.go`; `git rm
internal/usercommands/suicide_npc.go _datafiles/world/empty/templates/character/permadeath.template`;
modify `internal/rules/rules.go`, `internal/rules/rules_test.go`, `internal/rules/dikubasic/ruleset_test.go`,
`internal/usercommands/suicide.go`, `internal/usercommands/usercommands.go`, `internal/usercommands/get.go`,
`internal/usercommands/look.go`, `internal/usercommands/look.panels.go`, `internal/usercommands/start.go`,
`internal/rooms/corpse.go`, `internal/hooks/NewRound_DoCombat.go`, `internal/hooks/NewRound_AutoHeal.go`,
`internal/hooks/NewRound_ActorUpkeep.go`, `internal/hooks/Buff_ApplyBuffs.go`,
`internal/configs/config.gameplay.go`, `internal/configs/configs_bench_test.go`,
`internal/characters/character.go`, `internal/accounts/userrecord.go`, `internal/events/eventtypes.go`,
`internal/templates/templatesfunctions.go`, `internal/guards/combat_test.go`, `internal/guards/guards_test.go`,
`_datafiles/config.yaml`, `_datafiles/html/admin/config-wizard.html`, `_datafiles/html/admin/config.html`,
`_datafiles/html/admin/users.html`, `cmd/smoke/script.go`, `cmd/smoke/main.go`,
`internal/hooks/combat_golden_test.go`, `internal/hooks/attack_engine_test.go`,
`internal/hooks/ruleset_stub_test.go`, `internal/hooks/actortick_test.go`, and the 9 goldens in Step 15;
create `internal/hooks/death_engine_test.go` and `internal/usercommands/death_test.go` (Step 11b).
`internal/usercommands/suicide_npc_test.go` stays as it is: it runs `Suicide` on an NPC, which now
goes through `death.Die`, and still passes.

- [ ] **Step 1: `rules.go`.** Add `github.com/GoMudEngine/GoMud/internal/items` to the imports. After
  `RegenResult` add:

```go
// Award is one account's share of a kill (Q1). After announcing the death
// the engine credits the kill (an NPC victim's), grants XP through the
// account (source "combat") and sends Messages, award by award.
type Award struct {
	UserId   int
	XP       int
	Messages []perception.Message
}

// DeathOutcome is what a death means. The engine carries it out
// (internal/death): it announces the death, moves the drops and respawns
// or ends the body.
type DeathOutcome struct {
	Revive        bool // the victim gets up instead of dying
	ReviveHealth  int  // health on revive
	Respawn       bool // true: back to life in the recovery room; false: the body ends
	RespawnHealth int  // health on respawn

	// DropItems leave the body, carried or worn, for its corpse (or the
	// floor when corpses are off); DropGold likewise.
	DropItems []items.Item
	DropGold  int

	Awards   []Award              // the killers' shares, in order
	Messages []perception.Message // further lines, sent after the awards
}
```

  After `Regen` in the interface add:

```go

	// OnDeath is what victim's death means. Its awards may change the
	// killers' characters (alignment); the victim's body is the engine's.
	OnDeath(victim *actors.Actor) DeathOutcome
```

  In `rules_test.go` add to `fake`: `func (fake) OnDeath(v *actors.Actor) DeathOutcome { return DeathOutcome{} }`.

- [ ] **Step 2: `internal/rules/dikubasic/death.go`** (new):

```go
package dikubasic

import (
	"fmt"
	"math"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// OnDeath is a death under the reference rules (decisions 7 and 11, slice 3
// rulings 1 and 3). A body with revive-on-death gets up at full health. An
// NPC's killers share its XP and shift alignment, as GoMud did. An account
// comes back in the recovery room at 1 health; an NPC's body ends. Drops: an
// account drops what it carries and its gold and keeps what it wears; an NPC
// drops what it carries, its gold, and each worn item at its ItemDropChance;
// nothing under perma-gear.
//
// Random draws, in the order usercommands/suicide_npc.go made them: the XP
// variation, any alignment floors, then one d100 per worn NPC item.
func (Ruleset) OnDeath(victim *actors.Actor) rules.DeathOutcome {
	c := victim.Character

	if c.HasBuffFlag("revive-on-death") {
		return rules.DeathOutcome{Revive: true, ReviveHealth: c.HealthMax.Value}
	}

	mob := mobs.GetInstance(victim.Id)
	if mob == nil {
		out := rules.DeathOutcome{Respawn: true, RespawnHealth: 1}
		if !c.HasBuffFlag("perma-gear") {
			out.DropItems = append([]items.Item{}, c.Items...)
			out.DropGold = c.Gold
		}
		return out
	}

	out := rules.DeathOutcome{Awards: killAwards(mob)}
	if c.HasBuffFlag("perma-gear") {
		return out
	}
	out.DropItems = append([]items.Item{}, c.Items...)
	for _, item := range c.Equipment.GetAllItems() {
		if item.IsRemoveLocked() {
			continue
		}

		roll := util.Rand(100)

		util.LogRoll(`Drop Item`, roll, mob.ItemDropChance)

		if roll >= mob.ItemDropChance {
			continue
		}
		out.DropItems = append(out.DropItems, item)
	}
	out.DropGold = c.Gold
	return out
}

// killAwards is the NPC's XP and alignment for its killers, moved from
// usercommands/suicide_npc.go (slice 3) with the engine's parts left out:
// the kill credit, the XP grant and every send are internal/death's.
// Alignment is applied here, in the order it was, because UpdateAlignment
// can draw a random number.
func killAwards(mob *mobs.Mob) []rules.Award {

	mobXP := mob.Character.XPTL(mob.Character.Level - 1)

	xpVal := mobXP / 90

	xpVariation := xpVal / 100
	if xpVariation < 1 {
		xpVariation = 1
	}

	awards := []rules.Award{}
	partyTracker := map[int]int{} // key is party leader ID, value is how much will be shared.

	if len(mob.Character.PlayerDamage) > 0 {

		xpVal = xpVal / len(mob.Character.PlayerDamage) // Div by number of players that beat him up
		xpVal += ((util.Rand(3) - 1) * xpVariation)     // a little bit of variation

		totalPlayerLevels := 0
		for uId := range mob.Character.PlayerDamage {
			if user := accounts.GetByUserId(uId); user != nil {
				totalPlayerLevels += user.Character.Level
			}
		}

		for uId := range mob.Character.PlayerDamage {
			user := accounts.GetByUserId(uId)
			if user == nil {
				continue
			}

			// In a party: the party shares it, below.
			if p := parties.Get(user.UserId); p != nil {
				partyTracker[p.LeaderUserId] += xpVal
				continue
			}

			xpScaler := 1.0

			// If there's a level delta of more than 5, apply a scaler
			if math.Abs(float64(mob.Character.Level)-float64(totalPlayerLevels)) > 5 {

				xpScaler = float64(mob.Character.Level) / float64(totalPlayerLevels) // How much of the mobs level is the player?
				if xpScaler > 1.5 {
					xpScaler = 1.5
				} else if xpScaler < 0.25 {
					xpScaler = 0.25
				}

			}

			finalXPVal := int(math.Ceil(float64(xpVal) * xpScaler))

			if mob.IsElite {
				finalXPVal = finalXPVal + int(math.Ceil(float64(finalXPVal)*float64(eliteXPBonus())/100.0))
			}

			mudlog.Debug("XP Calculation", "MobLevel", mob.Character.Level, "XPBase", mobXP, "xpVal", xpVal, "xpVariation", xpVariation, "xpScaler", xpScaler, "finalXPVal", finalXPVal)

			awards = append(awards, rules.Award{UserId: user.UserId, XP: finalXPVal, Messages: shiftAlignment(user, mob)})
		}
	}

	// Party XP: an equal share for every member.
	for leaderId, xp := range partyTracker {
		p := parties.Get(leaderId)
		if p == nil {
			continue
		}

		allMembers := p.GetMembers()
		xpSplit := xp / len(allMembers)

		if mob.IsElite {
			xpSplit = xpSplit + int(math.Ceil(float64(xpSplit)*float64(eliteXPBonus())/100.0))
		}

		mudlog.Info(`Party XP`, `totalXP`, xp, `splitXP`, xpSplit, `memberCt`, len(allMembers))

		for _, memberId := range allMembers {
			if user := accounts.GetByUserId(memberId); user != nil {
				awards = append(awards, rules.Award{UserId: user.UserId, XP: xpSplit, Messages: shiftAlignment(user, mob)})
			}
		}
	}

	return awards
}

// eliteXPBonus is the percent XP bonus for killing an elite NPC (10 when
// unset).
func eliteXPBonus() int {
	eliteBonus := int(configs.GetGamePlayConfig().EliteXPBonus)
	if eliteBonus <= 0 {
		eliteBonus = 10
	}
	return eliteBonus
}

// shiftAlignment applies the alignment change for user killing mob and
// returns the line that tells the player, if the alignment's name changed.
func shiftAlignment(user *accounts.Account, mob *mobs.Mob) []perception.Message {
	alignmentBefore := user.Character.AlignmentName()
	alignmentAdj := AlignmentChange(user.Character.Alignment, mob.Character.Alignment)
	user.Character.UpdateAlignment(alignmentAdj)
	alignmentAfter := user.Character.AlignmentName()

	mudlog.Debug("Alignment", "user Alignment", user.Character.Alignment, "mob Alignment", mob.Character.Alignment, `alignmentAdj`, alignmentAdj, `alignmentBefore`, alignmentBefore, `alignmentAfter`, alignmentAfter)

	if alignmentBefore == alignmentAfter {
		return nil
	}

	events.AddToQueue(events.AlignmentChanged{
		UserId:       user.UserId,
		AlignmentOld: alignmentBefore,
		AlignmentNew: alignmentAfter,
	})

	before := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
	after := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
	return []perception.Message{{
		Kind:      perception.Interface,
		Actor:     user.Character.Id,
		Audience:  perception.ActorOnly,
		ActorText: fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, before, after),
	}}
}
```

  For the controller's read, the moved arithmetic is `suicide_npc.go:59-238` less: the `MobDeath`
  event (`:61-74`), the aggro clear (`:100-105`), the `Training` kill tracking (`:112-117`, `:203-208`;
  now the engine's, for every award recipient), `user.GrantXP` (`:145`, `:210`) and `messaging.Tell`
  (`:159`, `:224`; now `Award.Messages`). `perception.Interface` with `ActorOnly` is what
  `messaging.Tell` builds (`messaging/interface.go:37-39`). Both elite blocks read the same config as
  before, through `eliteXPBonus`.

- [ ] **Step 3: `internal/death/death.go`** (new):

```go
// Package death is the engine's one death routine (Phase 5 slice 3): every
// actor dies through Die, player or NPC, and what a death means is the
// ruleset's (rules.Ruleset.OnDeath). The engine announces the death, keeps
// the books, moves the drops and respawns or ends the body. hooks and
// usercommands import it; it imports neither.
package death

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Die is victim's death, whatever caused it: a combat round, round upkeep,
// a buff, or the suicide command.
func Die(victim *actors.Actor) {
	if victim == nil || victim.Character == nil {
		return
	}
	c := victim.Character

	outcome := rules.Get().OnDeath(victim)

	if outcome.Revive {
		revive(victim, outcome.ReviveHealth)
		return
	}

	user := accountOf(victim)
	mob := mobs.GetInstance(victim.Id)
	room := rooms.LoadRoom(c.RoomId)

	if mob != nil {
		mudlog.Debug(`Mob Death`, `name`, c.Name)
	}

	// Send a death msg to everyone in the room.
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        victim.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ObserverText: `{actor} has died.`,
	})

	if user != nil {
		recordPlayerDeath(user)
	}
	if mob != nil {
		recordMobDeath(mob)
	}

	for _, aw := range outcome.Awards {
		award(aw, mob)
	}
	for _, m := range outcome.Messages {
		messaging.Send(m)
	}

	clearKillersAggro(victim)

	// The drops leave the body. With no room to put them in, they stay on it.
	var drops []items.Item
	gold := 0
	corpsesOn := bool(configs.GetGamePlayConfig().Death.CorpsesEnabled)
	if room != nil {
		drops, gold = takeDrops(c, outcome)
		placeDrops(room, mob, c.Zone, drops, gold, corpsesOn)
	}

	corpse := rooms.Corpse{Character: corpseBody(c), RoundCreated: util.GetRoundCount(), Items: drops, Gold: gold}

	if !outcome.Respawn && mob != nil {
		end(mob, room)
		if corpsesOn && room != nil {
			corpse.MobId = int(mob.MobId)
			room.AddCorpse(corpse)
		}
		return
	}

	// Every body with an account comes back: there is no permadeath
	// (decision 11), and ending a body needs an NPC instance (Q7b).
	if corpsesOn && room != nil {
		if user != nil {
			corpse.UserId = user.UserId
		} else if mob != nil {
			corpse.MobId = int(mob.MobId)
		}
		room.AddCorpse(corpse)
	}
	respawn(victim, user, outcome.RespawnHealth)
}

// accountOf is the account driving a, or nil for an NPC.
func accountOf(a *actors.Actor) *accounts.Account {
	if ctrl := a.Controller(); ctrl != nil {
		return ctrl.Account()
	}
	return nil
}

// revive gets the victim up instead of dying.
func revive(victim *actors.Actor, health int) {
	victim.Character.Health = health

	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        victim.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ObserverText: `{actor} is suddenly revived in a shower of sparks!`,
	})

	victim.Character.CancelBuffsWithFlag("revive-on-death")

	if user := accountOf(victim); user != nil {
		events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
	}
}

// recordPlayerDeath keeps an account's books: the kill and death counts, the
// world notice naming who killed it, the elite death, the PlayerDeath event
// and the event log (suicide.go before slice 3, less permadeath).
func recordPlayerDeath(user *accounts.Account) {
	c := user.Character

	i := 0
	dmgCt := len(c.PlayerDamage)

	if dmgCt > 0 {
		c.KD.AddPvpDeath()
	} else {
		c.KD.AddMobDeath()
	}

	killedByUserIds := []int{}
	killedBy := ``
	for uid := range c.PlayerDamage {

		if u := accounts.GetByUserId(uid); u != nil {

			// Update PK stats
			c.KD.AddPlayerDeath(u.UserId, u.Character.Name)
			u.Character.KD.AddPlayerKill(user.UserId, c.Name)

			if i > 0 {
				if i < dmgCt-1 {
					killedBy += `, `
				} else {
					killedBy += ` and `
				}
			}
			killedBy += `<ansi fg="username">` + u.Character.Name + `</ansi>`
			i++
		}

		killedByUserIds = append(killedByUserIds, uid)
	}

	msg := fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, c.Name, term.CRLFStr)
	if killedBy != `` {
		msg = fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> (killed by %s) <ansi fg="magenta-bold">***</ansi>%s`, c.Name, killedBy, term.CRLFStr)
	}

	messaging.Notify(messaging.Notice{
		Text: msg,
	})

	killerMobId := 0
	if !c.KilledBy.IsZero() {
		killerMobId = c.KillerMobId
		if c.KillerMobIsElite {
			c.KD.AddEliteDeath(killerMobId, c.KillerMobName)
		}
		c.KilledBy = ids.ID{}
		c.KillerMobIsElite = false
		c.KillerMobName = ``
		c.KillerMobId = 0
	}

	events.AddToQueue(events.PlayerDeath{
		UserId:        user.UserId,
		RoomId:        c.RoomId,
		Username:      user.Username,
		CharacterName: c.Name,
		KilledByUsers: killedByUserIds,
		KillerMobId:   killerMobId,
	})

	user.EventLog.Add(`death`, fmt.Sprintf(`<ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED</ansi>`, c.Name))
}

// recordMobDeath queues the MobDeath event.
func recordMobDeath(mob *mobs.Mob) {
	c := &mob.Character
	killedByUsers := make([]int, 0, len(c.PlayerDamage))
	for uId := range c.PlayerDamage {
		killedByUsers = append(killedByUsers, uId)
	}

	events.AddToQueue(events.MobDeath{
		MobId:         int(mob.MobId),
		ActorId:       c.Id,
		RoomId:        c.RoomId,
		CharacterName: c.Name,
		Level:         c.Level,
		PlayerDamage:  c.PlayerDamage,
		KilledByUsers: killedByUsers,
	})
}

// award carries out one share of a kill (Q1): the kill credit when the
// victim is an NPC, the XP, then the award's lines.
func award(aw rules.Award, mob *mobs.Mob) {
	user := accounts.GetByUserId(aw.UserId)
	if user == nil {
		return
	}
	if mob != nil {
		user.Character.KD.AddMobKill(int(mob.MobId))
		if mob.IsElite {
			user.Character.KD.AddEliteKill(int(mob.MobId), mob.Character.Name)
		}
	}
	user.GrantXP(aw.XP, `combat`)
	for _, m := range aw.Messages {
		messaging.Send(m)
	}
}

// clearKillersAggro ends the fight of every account credited with damage to
// the victim whose body still targets it.
func clearKillersAggro(victim *actors.Actor) {
	for uId := range victim.Character.PlayerDamage {
		u := accounts.GetByUserId(uId)
		if u == nil || u.Character.Aggro == nil || u.Character.Aggro.Target != victim.Id {
			continue
		}
		u.Character.Aggro = nil
		events.AddToQueue(events.AggroChanged{ActorId: u.Character.Id, RoomId: u.Character.RoomId})
	}
}

// takeDrops removes the outcome's drops from the body and returns what left
// it: each item carried or worn (an item the body does not hold is not
// dropped), and the gold, up to what the body has. Each item that leaves
// queues ItemOwnership (lost), as every other item move does (spec, plan
// review round 1).
func takeDrops(c *characters.Character, outcome rules.DeathOutcome) ([]items.Item, int) {
	drops := make([]items.Item, 0, len(outcome.DropItems))
	for _, itm := range outcome.DropItems {
		if c.RemoveItem(itm) || c.RemoveFromBody(itm) {
			drops = append(drops, itm)
			events.AddToQueue(events.ItemOwnership{ActorId: c.Id, Item: itm, Gained: false})
		}
	}
	gold := min(max(outcome.DropGold, 0), c.Gold)
	c.Gold -= gold
	return drops, gold
}

// placeDrops puts the drops on the floor when corpses are off, with the
// lines an NPC's floor drops have always had (Q7d); otherwise the corpse
// holds them. MobItemDrop is queued for each NPC drop either way (Q7e).
func placeDrops(room *rooms.Room, mob *mobs.Mob, zone string, drops []items.Item, gold int, corpsesOn bool) {
	for _, itm := range drops {
		if !corpsesOn {
			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Room:         room.RoomId,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ObserverText: fmt.Sprintf(`<ansi fg="item">%s</ansi> drops to the ground.`, itm.DisplayName()),
			})
		}
		if mob != nil {
			events.AddToQueue(events.MobItemDrop{
				MobId:  int(mob.MobId),
				RoomId: room.RoomId,
				Zone:   zone,
				ItemId: itm.ItemId,
			})
		}
		if !corpsesOn {
			room.AddItem(itm, false)
		}
	}

	if !corpsesOn && gold > 0 {
		messaging.Send(perception.Message{
			Kind:         perception.WorldEvent,
			Room:         room.RoomId,
			Senses:       perception.Sight | perception.Sound,
			Range:        perception.Room,
			ObserverText: fmt.Sprintf(`<ansi fg="yellow-bold">%d gold</ansi> drops to the ground.`, gold),
		})
		room.Gold += gold
	}
}

// corpseBody is the dead body for its corpse, carrying, wearing and holding
// nothing: the corpse holds the drops (Corpse.Items, Corpse.Gold), so get
// and look find nothing the dead kept (Q2).
func corpseBody(c *characters.Character) characters.Character {
	body := *c
	body.Items = nil
	body.Equipment = characters.Worn{}
	body.Gold = 0
	return body
}

// end is an NPC's ending (decision 15): its slot waits its respawn rate.
func end(mob *mobs.Mob, room *rooms.Room) {
	rooms.NPCEnded(mob, false)

	// Destroy any record of this mob.
	mobs.DestroyInstance(mob.Character.Id)

	// Clean up mob from room...
	if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
		r.CleanupMobSpawns()
	}

	if room != nil {
		room.RemoveMob(mob.Character.Id)
	}
}

// respawn brings the body back in the death recovery room (its own private
// copy, MoveToRoom) with no buffs, the given health and no mana (Q7c), its
// damage record cleared. recordPlayerDeath has already cleared the killer
// record (only an account has one: NewRound_DoCombat records it for a
// defender with an account).
func respawn(victim *actors.Actor, user *accounts.Account, health int) {
	c := victim.Character

	c.CancelBuffsWithFlag(buffs.All)

	c.Health = health
	c.Mana = 0
	if user != nil {
		events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
	}

	clear(c.PlayerDamage)

	if err := rooms.MoveToRoom(victim.Id, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)); err != nil {
		mudlog.Error("death", "error", err)
	}
}
```

  Check before relying on it: `go list -deps ./internal/death | grep -E 'internal/(hooks|usercommands)$'`
  prints nothing.

- [ ] **Step 4: `suicide.go`** becomes, in full:

```go
package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// Suicide is a death by the actor's own hand, through the one death routine
// (internal/death), for players and NPCs alike.
func Suicide(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	death.Die(actor)
	return true, nil
}
```

  `git rm internal/usercommands/suicide_npc.go`. In `usercommands.go` (`:186-187`) the comment
  `// Kept until Phase 5: it is player and NPC death. NewRound_AutoHeal, NewRound_DoCombat` /
  `// (players and mobs), and Buff_ApplyBuffs dispatch it.` becomes
  `// Death by one's own hand (death.Die). Allowed when downed, so a downed actor` /
  `// can end it under a ruleset with a downed state.`.

- [ ] **Step 5: The engine's death checks call `Die`.** In `NewRound_DoCombat.go`, `handleAffected`
  (Task 3's version, with its comment) becomes:

```go
// handleAffected settles everyone the round touched, actors with an account
// first, then NPCs, each once: the ruleset says who is dead (death.Die) or,
// under a ruleset with a downed state, down (an account's PlayerDrop).
func handleAffected(affected []ids.ID) {

	handled := map[ids.ID]struct{}{}
	npcs := []*actors.Actor{}

	for _, id := range affected {
		if _, ok := handled[id]; ok {
			continue
		}
		handled[id] = struct{}{}

		a := actors.Get(id)
		if a == nil {
			continue
		}

		if user := accountOf(a); user != nil {
			switch rules.Get().Condition(a) {
			case rules.Dead:
				death.Die(a)
			case rules.Downed:
				events.AddToQueue(events.PlayerDrop{UserId: user.UserId, RoomId: user.Character.RoomId})
			}
			continue
		}

		if mobs.GetInstance(id) != nil {
			npcs = append(npcs, a)
		}
	}

	for _, a := range npcs {
		if rules.Get().Condition(a) == rules.Dead {
			death.Die(a)
		}
	}

}
```

  Add the `github.com/GoMudEngine/GoMud/internal/death` import.

  In `NewRound_ActorUpkeep.go`, the comment's `an NPC at 0 health dies` becomes
  `an actor the ruleset calls dead dies (death.Die)`; delete
  `mob := mobs.GetInstance(a.Id) // nil for a player`; and `:55-65` (from `if user != nil {` through the
  closing brace of the `if mob != nil && rules.Get().Condition(a) == rules.Dead {` block) becomes:

```go
		if user != nil {
			if char.Health != hBefore || char.Mana != mBefore {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
			}
		}

		if rules.Get().Condition(a) == rules.Dead {
			death.Die(a)
		}
```

  Add the `death` import. (Q7a: any actor.)

  In `Buff_ApplyBuffs.go`, the `TriggerNow` block (`:62-72`) becomes:

```go
	if buffInfo.TriggerNow {

		if a := actors.Get(evt.ActorId); a != nil && rules.Get().Condition(a) == rules.Dead {
			death.Die(a)
		}
	}
```

  Add the `death` import. (Q7a: any actor.)

- [ ] **Step 6: The regen pass loses the bleed-out.** `NewRound_AutoHeal.go` becomes, in full:

```go
package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/rules"
)

// AutoHeal is the engine's regen pass (Phase 5 slice 3): every round, for
// every actor with a body, the ruleset says what it regenerates; the engine
// applies it, sends the ruleset's lines and reports an account's changed
// vitals. A ruleset that wants bleeding returns a negative Regen.
func AutoHeal(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	for _, a := range fighters() {
		healthStart := a.Character.Health
		manaStart := a.Character.Mana

		r := rules.Get().Regen(a, evt.RoundNumber)
		if r.Changes() {
			a.Character.Heal(r.Health, r.Mana)
		}
		for _, m := range r.Messages {
			messaging.Send(m)
		}

		user := accountOf(a)
		if user == nil {
			continue
		}

		// If it has changed, send an update
		if a.Character.Health != healthStart || a.Character.Mana != manaStart {

			// Trigger a redraw, but only if the users prompt has changed.
			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId, OnlyIfChanged: true}, 100)

			events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})

		}
	}

	return events.Continue
}
```

- [ ] **Step 7: `CorpseItems` readers.** `get.go:90-91` (`// Look for a corpse as the source when CorpseItems is enabled`
  and `if containerName == `` && configs.GetGamePlayConfig().Death.CorpseItems {`) become
  `// Look for a corpse as the source` and ``if containerName == `` {``. `look.go:412-414` (the `if`
  around `Tell(actor, buildCorpseInventoryPanel(&corpse))`) becomes the bare
  `Tell(actor, buildCorpseInventoryPanel(&corpse))`. Remove the `configs` import from both files (each
  had only that use). `look.panels.go:110` comment becomes
  `// buildCorpseInventoryPanel renders the lootable contents of a corpse.`. `rooms/corpse.go:15-16`
  comments become `// What the body dropped (rules.DeathOutcome)` and `// The gold it dropped`.

- [ ] **Step 7b: `get all <corpse>` works** (spec, plan review round 1; player-visible). Today
  `get.go:40` (`if args[0] == "all" {`) takes everything on the floor and returns whenever the first
  word is `all`, so `get all bob` and `get all from bob` never reach the corpse lookup (`:90-100`) or
  its `rest == "all"` branch (`:112-178`), which already takes the gold and every item. The
  controller's ruling (spec, plan review round 1): when the word after `all` (or after `all from`)
  names a corpse in the room, `get` takes everything from that corpse; any other `get all ...` keeps
  today's room-wide behaviour exactly (`get all`, `get all ground`, `get all stash`,
  `get all <container>`, `get all <anything else>`). So the room-wide branch is not narrowed: it is
  skipped only for a corpse. `get.go:40` (it has no comment) becomes:

```go
	// "get all <corpse>" and "get all from <corpse>" take everything from
	// that corpse (the corpse branch below). Every other "get all ..." takes
	// everything on the floor, as it always has.
	if args[0] == "all" && !namesACorpse(room, args) {
```

  and after `Get` (after its closing brace, `:435`) add:

```go
// namesACorpse reports whether args is "all <word>" or "all from <word>"
// with <word> naming a corpse in the room. A container of the same name
// wins, as it does for every get (the container lookup comes first,
// :80-88), so then this is false and "get all" keeps its room-wide meaning.
func namesACorpse(room *rooms.Room, args []string) bool {
	if len(args) != 2 && !(len(args) == 3 && args[1] == "from") {
		return false
	}
	word := args[len(args)-1]
	if room.FindContainerByName(word) != `` {
		return false
	}
	_, ok := room.FindCorpseByRef(word)
	return ok
}
```

  Nothing else in `get.go` changes: for a corpse word the lookup below finds the same corpse
  (`:90-100`, now without the `CorpseItems` gate), strips an optional `from` (`:94-98`) and leaves
  `rest == "all"` for the corpse branch. Every other `get all ...` takes the same branch as today (reads
  golden D runs plain `get all`, `reads_golden_test.go:705`, and must stay byte-identical).
  `FindCorpseByRef` is `rooms.go:945-990` (skips a prunable corpse); `FindContainerByName` is
  `rooms.go:1464`. The tests are in Step 11b (`TestGetFromAPlayersCorpse`,
  `TestGetAllWithoutACorpseTakesTheFloor`).

- [ ] **Step 8: Permadeath, lives and the deleted knobs.**
  - `configs/config.gameplay.go`: delete the `LivesStart`, `LivesMax`, `LivesOnLevelUp` fields
    (`:17-19`); in `GameplayDeath` delete `EquipmentDropChance`, `AlwaysDropBackpack`, `XPPenalty`,
    `ProtectionLevels`, `PermaDeath`, `CorpseItems`, keeping `CorpsesEnabled` and `CorpseDecayTime`; in
    `Validate`, `:93-131` (from `// Ignore AllowItemBuffRemoval` through the closing brace of the
    `LivesOnLevelUp` check) becomes:

```go
	// Ignore AllowItemBuffRemoval
	// Ignore CorpsesEnabled
```

    and delete the `import ( "strconv" "strings" )` block (`:3-6`; their only uses were in the deleted
    lines). `configs_bench_test.go`: delete `:28` (`c.GamePlay.Death.XPPenalty = "none"`) and the three
    map lines `"XPPenalty":`, `"EquipmentDropChance":`, `"ProtectionLevels":` (`:174-176`).
  - `characters/character.go:80`: delete the `ExtraLives` field. In the `forgetKillerIfUp` comment
    (`:703-708`), `suicide) is not credited` stays and `on a level gain. suicide.go's revive clears the record itself.`
    becomes `on a level gain. death.Die's respawn clears the record itself.`.
  - `accounts/userrecord.go`: in `NewAccount` delete `c := configs.GetGamePlayConfig()` (`:53`) and the
    `if c.Death.PermaDeath { ... }` block (`:69-71`). In `GrantXP` delete `c := configs.GetGamePlayConfig()`
    (`:202`), `livesBefore := u.Character.ExtraLives` (`:204`), `LivesGained:    0,` (`:216`), the
    `if c.Death.PermaDeath && c.LivesOnLevelUp > 0 { ... }` block (`:221-223`), the
    `if u.Character.ExtraLives > int(c.LivesMax) { ... }` block (`:245-247`) and
    `levelUpEvent.LivesGained = u.Character.ExtraLives - livesBefore` (`:249`).
  - `usercommands/start.go:182`: delete `user.Character.ExtraLives = int(configs.GetGamePlayConfig().LivesStart)`
    (and the blank line after it). `configs` stays imported (`:141`, `:192`).
  - `templates/templatesfunctions.go:193-195`: delete the `"permadeath"` func. `git rm
    _datafiles/world/empty/templates/character/permadeath.template`.
  - `events/eventtypes.go`: delete `LivesGained    int` (`:208`) and `Permanent     bool` (`:225`).
  - `_datafiles/config.yaml`: `:194-237` (from `  # Death settings` through `  LivesOnLevelUp: 1`)
    becomes (CRLF, as the file):

```yaml
  # Death settings
  Death:
    # - CorpsesEnabled -
    #   If set to true, corpses are left behind by players or mobs that are killed.
    #   A corpse holds what its owner dropped: a player's carried items and gold,
    #   an NPC's carried items, gold and the worn items its ItemDropChance drops.
    #   Players take them with "get <item> <corpse>", or everything with
    #   "get all <corpse>" or "get all from <corpse>".
    CorpsesEnabled: true
    # - CorpseDecayTime -
    #   How long until corpses crumble to dust (Go away). Everything still in a
    #   corpse when it crumbles is destroyed.
    #   See ShopRestockRate comments for time format.
    CorpseDecayTime: 1 hour
```

  - `_datafiles/html/admin/config-wizard.html`: delete `:1110-1130` (`'GamePlay.LivesStart': {` through
    the `},` closing `'GamePlay.LivesOnLevelUp'`), `:1161-1200` (`'GamePlay.Death.EquipmentDropChance': {`
    through the `},` closing `'GamePlay.Death.PermaDeath'`), `:1214-1219` (`'GamePlay.Death.CorpseItems'`
    through its `},`) and the Lives group `:1543-1551` (the `{` line before `key: 'GamePlay.Lives',`
    through its `},`). In the Death group, `desc` (`:1555`) becomes
    `desc: 'Corpses, and how long they last before crumbling.',` and `fields` (`:1559`) becomes
    `fields: ['GamePlay.Death.CorpsesEnabled','GamePlay.Death.CorpseDecayTime'],`. Do these edits
    bottom-up so earlier line numbers hold.
  - `_datafiles/html/admin/config.html`: delete `:228-232`, `:234` and `:236-238` (the
    `EquipmentDropChance`, `AlwaysDropBackpack`, `XPPenalty`, `ProtectionLevels`, `PermaDeath`,
    `CorpseItems`, `LivesStart`, `LivesMax`, `LivesOnLevelUp` entries), keeping `CorpsesEnabled` and
    `CorpseDecayTime`.
  - `_datafiles/html/admin/users.html`: delete the Extra Lives field (`:366-369`, the
    `<div class="field">` holding `<label>Extra Lives</label>` through its `</div>`), `:592`
    (`setVal('f-extralives', ...)`) and `:1089` (`ExtraLives: ...`).
  - Prove absence, whole words only (`-w`, so `get.go:134`'s `allCorpseItems`, a variable that
    stays, does not match):
    `grep -rnwE 'ExtraLives|LivesStart|LivesMax|LivesOnLevelUp|PermaDeath|ProtectionLevels|XPPenalty|EquipmentDropChance|AlwaysDropBackpack|CorpseItems|LivesGained|f-extralives' --exclude-dir=.git --exclude-dir=.claude --include=*.go --include=*.html --include=*.yaml --include=*.template .`
    prints nothing. Controls: on the code before this task the same command printed 117 lines in 18
    files (counted at `8b424308`, `get.go:90-91` among them and `get.go:134` not); after it, the same
    command with `CorpsesEnabled` in place of the list prints hits.
    The lower-case word is checked apart, because comments may still say it:
    `grep -rnw 'permadeath' --exclude-dir=.git --exclude-dir=.claude --include=*.go --include=*.html --include=*.yaml --include=*.template .`
    prints exactly these comment lines and nothing else: `internal/actors/registry.go:169` and `:180`
    ("a permadeath reroll", Q7m), and in the new `internal/death/death.go` the `recordPlayerDeath`
    comment (`suicide.go before slice 3, less permadeath`) and the `Die` comment (`there is no
    permadeath (decision 11)`). A hit in code (a template func, a config key, a template name) is a
    miss: fix it.

- [ ] **Step 9: Guards' allowlists.** `guards/combat_test.go:122` (the `suicide_npc.go` entry of
  `TestOnlyBootNamesTheRuleset`): delete. `guards/guards_test.go:167` (`"internal/usercommands/suicide.go":
  "permadeath reroll: a new character",`): delete (Q7n). `gofmt -w` both (alignment).

- [ ] **Step 10: `dikubasic` unit tests.** Append to `ruleset_test.go` (imports: add `"math"`,
  `github.com/GoMudEngine/GoMud/internal/buffs`, `github.com/GoMudEngine/GoMud/internal/events`,
  `github.com/GoMudEngine/GoMud/internal/items` and `github.com/GoMudEngine/GoMud/internal/parties`):

```go
// trinket is an item with its own spec (no item file needed).
func trinket(itemId int, name string, typ items.ItemType) items.Item {
	spec := items.ItemSpec{ItemId: itemId, Name: name, NameSimple: name, Type: typ}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

func idsOf(list []items.Item) []int {
	out := []int{}
	for _, it := range list {
		out = append(out, it.ItemId)
	}
	return out
}

// onlinePlayer is player, registered so accounts.GetByUserId finds it.
func onlinePlayer(t *testing.T, userId int) *actors.Actor {
	t.Helper()
	a := player(t, userId)
	accounts.SetTestUser(a.Controller().Account())
	t.Cleanup(func() { accounts.RemoveTestUser(userId) })
	return a
}

// flagged gives c a test buff carrying one flag. A bare test NPC
// (mobs.NewTestInstance(mobs.Mob{})) has a zero Buffs whose maps are nil,
// and AddBuff writes into them (buffs.go:187), so it gets buffs.New() first,
// as givePermaGear does (usercommands/items_test.go:42).
func flagged(t *testing.T, c *characters.Character, buffId int, flag string) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000, Flags: []string{flag},
	}))
	if c.Buffs.List == nil { // buffs.New() makes List non-nil (buffs.go:43-49)
		c.Buffs = buffs.New()
	}
	require.NoError(t, c.AddBuff(buffId, true))
	require.True(t, c.HasBuffFlag(flag))
}

// Decision 11: a player drops what it carries and its gold, keeps what it
// wears, and comes back at 1 health.
func TestOnDeathAPlayerDropsWhatItCarriesAndKeepsWhatItWears(t *testing.T) {
	a := player(t, 9630)
	a.Character.Items = []items.Item{trinket(96101, "testgem", items.Gemstone)}
	a.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)
	a.Character.Gold = 25

	out := New().OnDeath(a)
	assert.Equal(t, []int{96101}, idsOf(out.DropItems))
	assert.Equal(t, 25, out.DropGold)
	assert.True(t, out.Respawn)
	assert.Equal(t, 1, out.RespawnHealth)
	assert.False(t, out.Revive)
	assert.Empty(t, out.Awards)
}

// Ruling 3: an NPC drops what it carries and its gold always, and each worn
// item at its ItemDropChance, one d100 per item in slot order (Neck before
// Ring, items.AllEquipSlots).
func TestOnDeathAnNPCDropsEachWornItemAtItsDropChance(t *testing.T) {
	n := npc(t, 0, 0)
	mobs.GetInstance(n.Id).ItemDropChance = 50
	n.Character.Items = []items.Item{trinket(96101, "testgem", items.Gemstone)}
	n.Character.Equipment.Neck = trinket(96103, "testamulet", items.Neck)
	n.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)
	n.Character.Gold = 30
	reseed(t)
	for seed := int64(1); seed <= 500; seed++ {
		util.SeedRandForTest(seed)
		want := []int{96101}
		if util.Rand(100) < 50 {
			want = append(want, 96103)
		}
		if util.Rand(100) < 50 {
			want = append(want, 96102)
		}
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		out := New().OnDeath(n)
		require.Equal(t, want, idsOf(out.DropItems), "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: one draw per worn item", seed)
		require.Equal(t, 30, out.DropGold)
		require.False(t, out.Respawn, "an NPC's body ends")
	}
}

// The draw order of the moved code: the XP variation (one d3) before the
// worn-item rolls. Both alignments are 0, so UpdateAlignment draws nothing.
func TestOnDeathDrawsTheXPVariationBeforeTheDrops(t *testing.T) {
	onlinePlayer(t, 9631)
	n := npc(t, 0, 0)
	mobs.GetInstance(n.Id).ItemDropChance = 50
	n.Character.PlayerDamage = map[int]int{9631: 5}
	n.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)

	// One killer within 5 levels of the NPC: no scaler, not elite.
	base := n.Character.XPTL(n.Character.Level-1) / 90
	variation := base / 100
	if variation < 1 {
		variation = 1
	}
	reseed(t)
	for seed := int64(1); seed <= 300; seed++ {
		util.SeedRandForTest(seed)
		xp := base + (util.Rand(3)-1)*variation
		ring := util.Rand(100) < 50
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		out := New().OnDeath(n)
		require.Len(t, out.Awards, 1)
		require.Equal(t, 9631, out.Awards[0].UserId)
		require.Equal(t, xp, out.Awards[0].XP, "seed %d", seed)
		require.Equal(t, ring, len(out.DropItems) == 1, "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: two draws", seed)
	}
}

// An alignment shift is applied to the killer, returned as its line, and
// queued as AlignmentChanged (the ruleset may queue an event: spec, plan
// review round 1).
func TestOnDeathShiftsTheKillersAlignment(t *testing.T) {
	killer := onlinePlayer(t, 9632)
	killer.Character.Alignment = 19 // neutral; killing an evil NPC makes her lawful (reads golden H)
	n := npc(t, 0, 0)
	n.Character.Alignment = -100
	n.Character.PlayerDamage = map[int]int{9632: 5}
	var changed []events.AlignmentChanged
	id := events.RegisterListener(events.AlignmentChanged{}, func(e events.Event) events.ListenerReturn {
		changed = append(changed, e.(events.AlignmentChanged))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.AlignmentChanged{}, id) })
	events.ProcessEvents() // drop what setup queued

	out := New().OnDeath(n)
	events.ProcessEvents()
	require.Len(t, out.Awards, 1)
	require.Len(t, out.Awards[0].Messages, 1)
	assert.Contains(t, out.Awards[0].Messages[0].ActorText, "Your alignment has shifted from")
	assert.Equal(t, killer.Id, out.Awards[0].Messages[0].Actor)
	assert.Equal(t, "lawful", killer.Character.AlignmentName())
	require.Len(t, changed, 1)
	assert.Equal(t, events.AlignmentChanged{UserId: 9632, AlignmentOld: "neutral", AlignmentNew: "lawful"}, changed[0])
}

// Party XP, moved as is: a killer in a party puts its share in the party's
// pot, which is split equally among the members, each awarded (a member who
// dealt no damage too), with the elite bonus (10% when EliteXPBonus is
// unset) added to each split for an elite NPC. Both alignments are 0, so
// UpdateAlignment draws nothing.
func TestOnDeathSplitsAPartysShare(t *testing.T) {
	onlinePlayer(t, 9640)
	onlinePlayer(t, 9641)
	p := parties.New(9640)
	require.NotNil(t, p)
	require.True(t, p.InvitePlayer(9641))
	require.True(t, p.AcceptInvite(9641))
	t.Cleanup(p.Disband)
	n := npc(t, 0, 0)
	n.Character.Level = 5
	n.Character.TNLScale = 1 // a bare test NPC has 0, which makes its XP 0 (XPTL, character.go:923-933)
	n.Character.PlayerDamage = map[int]int{9640: 5}

	base := n.Character.XPTL(n.Character.Level-1) / 90
	variation := max(base/100, 1)
	reseed(t)
	for _, elite := range []bool{false, true} {
		mobs.GetInstance(n.Id).IsElite = elite
		util.SeedRandForTest(7)
		split := (base + (util.Rand(3)-1)*variation) / 2
		if elite {
			split += int(math.Ceil(float64(split) * 10 / 100.0))
		}
		require.Positive(t, split, "the fixture needs XP to share")

		util.SeedRandForTest(7)
		out := New().OnDeath(n)
		require.Len(t, out.Awards, 2, "elite %v: one award per member", elite)
		assert.Equal(t, []int{9640, 9641}, []int{out.Awards[0].UserId, out.Awards[1].UserId}, "elite %v: member order", elite)
		assert.Equal(t, split, out.Awards[0].XP, "elite %v", elite)
		assert.Equal(t, split, out.Awards[1].XP, "elite %v", elite)
	}
}

func TestOnDeathPermaGearDropsNothing(t *testing.T) {
	p, n := player(t, 9633), npc(t, 0, 0)
	mobs.GetInstance(n.Id).ItemDropChance = 100
	for _, a := range []*actors.Actor{p, n} {
		flagged(t, a.Character, 96201, "perma-gear")
		a.Character.Items = []items.Item{trinket(96101, "testgem", items.Gemstone)}
		a.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)
		a.Character.Gold = 25
		out := New().OnDeath(a)
		assert.Empty(t, out.DropItems)
		assert.Zero(t, out.DropGold)
	}
}

// revive-on-death revives any actor at full health (an NPC's rule before
// slice 3, now everyone's).
func TestOnDeathReviveOnDeathRevivesAnyActorAtFullHealth(t *testing.T) {
	for name, a := range map[string]*actors.Actor{"account": player(t, 9634), "NPC": npc(t, 0, 0)} {
		flagged(t, a.Character, 96202, "revive-on-death")
		a.Character.HealthMax.Value = 33 // after AddBuff's Validate
		a.Character.Gold = 25
		out := New().OnDeath(a)
		assert.True(t, out.Revive, name)
		assert.Equal(t, 33, out.ReviveHealth, name)
		assert.Empty(t, out.DropItems, name)
		assert.Zero(t, out.DropGold, name)
	}
}
```

  Check before relying on the alignment test: reads golden H shows neutral 19 against -100 moving to
  lawful (`reads_golden_test.go:840-849`; H-outside.txt line 6). Run
  `go test ./internal/rules/dikubasic -run OnDeath -count=1 -v`; all pass. Probes, one at a time: add
  `out.DropItems = append(out.DropItems, c.Equipment.GetAllItems()...)` to the account branch (the
  player test fails); `roll >= mob.ItemDropChance` to `roll > mob.ItemDropChance` (the NPC drop test
  fails on a seed rolling 50); move `out := rules.DeathOutcome{Awards: killAwards(mob)}` below the
  worn-item loop, building `Awards` after the drops (the draw-order test fails); delete the
  `revive-on-death` branch (the revive test fails); in `shiftAlignment` return `nil` always (the
  alignment test fails); delete the `events.AddToQueue(events.AlignmentChanged{...})` call (the
  `changed` assert fails); `xpSplit := xp / len(allMembers)` to `xpSplit := xp` (the party test fails);
  delete the party loop's `if mob.IsElite { ... }` block (the party test's elite case fails); delete
  the party loop's `awards = append(...)` (the party test's `Len` fails). Paste the nine lines. Also
  delete `c.Buffs = buffs.New()` from `flagged` and run the perma-gear and revive tests: they panic
  in `buffs.(*Buffs).AddBuff` (`assignment to entry in nil map`), which shows why the line is there;
  restore.

- [ ] **Step 11: Test code the change reaches (Q4, Q6).**
  - `combat_golden_test.go`: add `w.recoveryRoom()` as the first line after `newCombatWorld` in
    `TestCombatGoldenPvMToDeath`, `TestCombatGoldenPvP`, `TestCombatGoldenMvP`,
    `TestCombatGoldenHiddenAttacker`. The dead player now respawns at 1 health, so the health
    asserts that said "went down" assert the death (Q6):
    - `TestCombatGoldenPvP` (`:421`): `require.Less(t, bob.Character.Health, 1, "the fixture must reach Bob going down")`
      becomes `require.True(t, rooms.IsEphemeralRoomId(bob.Character.RoomId), "the fixture must reach Bob's death")`;
    - `TestCombatGoldenMvP` (`:434`): `require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")`
      becomes `require.True(t, rooms.IsEphemeralRoomId(alice.Character.RoomId), "the fixture must reach Alice's death")`;
    - `TestCombatGoldenHiddenAttacker` (no health assert today): after `w.check("hidden_attacker")` add
      `require.True(t, rooms.IsEphemeralRoomId(bob.Character.RoomId), "the fixture must reach Bob's death")`.

    The rat asserts in `TestCombatGoldenPvMToDeath` (`:406`) and `TestCombatGoldenMvM` (`:447`) stay:
    an NPC's body is not revived, and `rat` still points at it after the instance ends.
    `deathConfig` loses the deleted keys and its comment; it becomes:

```go
// deathConfig turns corpses on for one fixture (the test binary loads no
// config.yaml, so they are off otherwise).
func deathConfig(t *testing.T) {
	t.Helper()
	gp := configs.GetGamePlayConfig()
	setConfig(t,
		map[string]any{"GamePlay.Death.CorpsesEnabled": true},
		map[string]any{"GamePlay.Death.CorpsesEnabled": bool(gp.Death.CorpsesEnabled)})
	require.True(t, bool(configs.GetGamePlayConfig().Death.CorpsesEnabled), "the override took")
}
```

    The comments of `TestCombatGoldenDeathPlayer` and `TestCombatGoldenDeathPvP` lose "above the
    protection band" / "drops by chance, XP penalty" and the `// above ProtectionLevels 1` comments
    (the `Level`/`Experience` lines stay, so the record shows no XP is lost). `TestKilledByOutlivesTheKiller`
    (`:540-559`, with its comment) becomes:

```go
// downedAtZero is the active ruleset with a downed state: at 0 health or
// below an actor is downed, never dead.
type downedAtZero struct{ rules.Ruleset }

func (downedAtZero) Condition(a *actors.Actor) rules.Condition {
	if a.Character.Health <= 0 {
		return rules.Downed
	}
	return rules.Up
}

// The killer is recorded when combat takes a player out of Up, so under a
// ruleset with a downed state, destroying the killer before the death does
// not lose its MobId; healing clears it. (Under the reference rules death
// is immediate and clears the record itself.)
func TestKilledByOutlivesTheKiller(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(downedAtZero{rules.Get()}))
	w.start()
	w.rounds(20)
	require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "downed, not dead")

	mobs.DestroyInstance(guard.Character.Id)
	require.Nil(t, mobs.GetInstance(guard.Character.Id))
	require.False(t, alice.Character.KilledBy.IsZero())
	require.Equal(t, 93202, alice.Character.KillerMobId, "the MobId was recorded while the guard existed")

	alice.Character.Heal(alice.Character.HealthMax.Value, 0)
	require.True(t, alice.Character.KilledBy.IsZero(), "back on her feet: no stale killer")
	require.Zero(t, alice.Character.KillerMobId)
}
```

    Add `github.com/GoMudEngine/GoMud/internal/rules` to the file's imports.
  - `attack_engine_test.go` `TestAttackRunsTheDefendersWimpyCheck`: add `w.recoveryRoom()` after
    `newCombatWorld` (Alice now dies; her health after respawn is 1, still under 12).
  - `ruleset_stub_test.go` `TestStubRulesetChangesTheFight`: add `w.recoveryRoom()` after
    `newCombatWorld` (Alice's XP may level her up and save her); line `:43` (the `require.Contains`
    that asserts the queued `Input rat: "suicide"`) becomes `require.Contains(t, out, "MobDeath rat")`.
  - `actortick_test.go` `TestActorUpkeepTicksEveryActorsCooldowns`: after
    `m.Character.Cooldowns = characters.Cooldowns{"bite": 3}` add
    `m.Character.Health = 10 // a bare test NPC is at 0, which is dead: upkeep would end it`.

- [ ] **Step 11b: Tests that pin the death path.** The goldens pin the round's deaths, the awards,
  the kill and death books, the drops into a corpse, `MobItemDrop` and the NPC's ending (Task 1). These
  tests pin the rest: the other two places that run a death, the corpses-off floor, revive and
  respawn, and taking things back from a corpse.

  (a) Create `internal/hooks/death_engine_test.go`:

```go
package hooks

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/require"
)

// taken splits what the recorder holds into the event notes and each
// player's reads.
func (w *combatWorld) taken() (notes []string, reads map[int][]string) {
	reads = map[int][]string{}
	for _, e := range w.rec.take() {
		if e.userId == 0 {
			notes = append(notes, e.text)
			continue
		}
		reads[e.userId] = append(reads[e.userId], e.text)
	}
	return notes, reads
}

// flag gives c a test buff carrying one flag and keeps its health (AddBuff
// runs Validate, which would reset it, as in hide).
func (w *combatWorld) flag(c *characters.Character, buffId int, flag string) {
	w.t.Helper()
	w.t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000, Flags: []string{flag},
	}))
	hpMax, hp := c.HealthMax.Value, c.Health
	require.NoError(w.t, c.AddBuff(buffId, true))
	c.HealthMax.Value, c.Health = hpMax, hp
	require.True(w.t, c.HasBuffFlag(flag))
}

// Round upkeep runs the death of any actor the ruleset calls dead (Q7a): an
// NPC ends, a player respawns.
func TestUpkeepRunsTheDeathOfWhoeverTheRulesetCallsDead(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "upkeep-rats")
	w.start()
	alice.Character.Health, rat.Character.Health = 0, 0

	ActorUpkeep(events.NewRound{RoundNumber: 1})
	events.ProcessEvents()

	notes, _ := w.taken()
	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the rat ended")
	require.Contains(t, notes, "MobDeath rat")
	require.Contains(t, notes, "PlayerDeath Alice")
	require.Equal(t, goldenRoomC, rooms.GetOriginalRoom(alice.Character.RoomId), "Alice is in a copy of the recovery room")
	require.Equal(t, 1, alice.Character.Health)
}

// An immediate buff trigger runs the death of a body the ruleset calls dead
// (Q7a). The death ends the fight of every credited account still targeting
// the body (clearKillersAggro), credits the kill and grants the XP (award).
func TestBuffTriggerDeathEndsTheKillersFightAndPaysTheKill(t *testing.T) {
	const triggerBuff = 93902
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: a level up saves the account
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "buff-rats")
	rat.Character.Level = 5 // enough XP that the grant shows
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	rat.Character.PlayerDamage = map[int]int{alice.UserId: 12}
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: triggerBuff, Name: "test trigger", TriggerNow: true, TriggerRate: "1 round", TriggerCount: 1,
	}))
	w.start()
	xpBefore := alice.Character.Experience
	rat.Character.Health = 0

	ApplyBuffs(events.Buff{ActorId: rat.Character.Id, BuffId: triggerBuff})
	events.ProcessEvents()

	notes, reads := w.taken()
	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the rat died")
	require.Contains(t, notes, "MobDeath rat")
	require.Nil(t, alice.Character.Aggro, "her fight with the rat ended")
	require.Contains(t, notes, "AggroChanged Alice room=93001")
	require.Equal(t, 1, alice.Character.KD.GetMobKills(93201), "the kill is credited")
	require.Greater(t, alice.Character.Experience, xpBefore, "the XP is granted")
	require.Contains(t, strings.Join(reads[alice.UserId], "\n"), "experience points")
}

// With corpses off the drops go on the floor with the floor lines (Q7d), and
// nothing is lost.
func TestDeathWithCorpsesOffDropsOnTheFloor(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	setConfig(t,
		map[string]any{"GamePlay.Death.CorpsesEnabled": false},
		map[string]any{"GamePlay.Death.CorpsesEnabled": bool(configs.GetGamePlayConfig().Death.CorpsesEnabled)})
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	alice.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	require.Equal(t, 25, alice.Character.Gold, "a new character's gold")
	w.start()

	death.Die(actors.Get(alice.Character.Id))
	events.ProcessEvents()

	notes, reads := w.taken()
	require.Contains(t, notes, "ItemOwnership Alice item=93108 gained=false", "the gem left her (spec, plan review round 1)")
	room := rooms.LoadRoom(goldenRoomA)
	require.Equal(t, "93108", itemIds(room.Items), "the gem is on the floor")
	require.Equal(t, 25, room.Gold, "her gold is on the floor")
	require.Empty(t, room.Corpses)
	require.Empty(t, alice.Character.Items)
	require.Zero(t, alice.Character.Gold)
	bobRead := strings.Join(reads[bob.UserId], "\n")
	require.Contains(t, bobRead, "testgem")
	require.Contains(t, bobRead, `<ansi fg="yellow-bold">25 gold</ansi> drops to the ground.`)
}

// revive-on-death gets a body up once: the flag is used up, so the next
// death is a death; an account's vitals are reported (Q7f).
func TestReviveOnDeathWorksOnce(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	w.flag(alice.Character, 93903, "revive-on-death")
	w.start()
	a := actors.Get(alice.Character.Id)

	alice.Character.Health = 0
	death.Die(a)
	events.ProcessEvents()
	notes, _ := w.taken()
	require.Positive(t, alice.Character.Health, "revived")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "not moved")
	require.False(t, alice.Character.HasBuffFlag("revive-on-death"), "the flag is used up")
	require.Contains(t, notes, "CharacterVitalsChanged Alice")
	require.NotContains(t, notes, "PlayerDeath Alice")

	alice.Character.Health = 0
	death.Die(a)
	events.ProcessEvents()
	notes, _ = w.taken()
	require.Contains(t, notes, "PlayerDeath Alice", "the second death is a death")
}

// A respawn starts clean: no buffs, no mana (Q7c), no damage record, no
// killer record; an elite killer is booked as an elite death first.
func TestRespawnStartsClean(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	w.flag(alice.Character, 93904, "testmark") // any flagged buff: buffs.All cancels flagged buffs
	w.start()
	c := alice.Character
	require.Positive(t, c.ManaMax.Value, "the fixture needs room for mana")
	c.Health, c.Mana = 0, c.ManaMax.Value
	c.PlayerDamage = map[int]int{9302: 3}
	c.KilledBy, c.KillerMobId, c.KillerMobName, c.KillerMobIsElite = guard.Character.Id, 93202, "guard", true

	death.Die(actors.Get(c.Id))
	events.ProcessEvents()

	require.Equal(t, goldenRoomC, rooms.GetOriginalRoom(c.RoomId))
	require.Equal(t, 1, c.Health)
	require.Zero(t, c.Mana, "no mana (Q7c)")
	require.False(t, c.HasBuffFlag("testmark"), "every buff is cancelled")
	require.Empty(t, c.PlayerDamage)
	require.True(t, c.KilledBy.IsZero())
	require.Zero(t, c.KillerMobId)
	require.Equal(t, 1, c.KD.EliteDeaths["93202:guard"], "an elite killer's death is booked")
}
```

  Check before relying on them: `buffs.All` cancels only buffs that carry a flag
  (`buffs.go:104-139`, the loop over `bSpec.Flags`), hence `testmark`; an unknown flag only logs a
  warning there. `KDStats.EliteDeaths` keys are `<mobId>:<name>` (`kdstats.go:104-110`). Run
  `go test ./internal/hooks -run 'Upkeep|BuffTrigger|CorpsesOff|ReviveOnDeath|RespawnStartsClean' -count=1 -v`;
  all pass. If `ManaMax` is 0 for the fixture, report it (do not drop the mana assert).

  Probes, one at a time (each named test fails; paste the line):
  - upkeep: delete the `if rules.Get().Condition(a) == rules.Dead { death.Die(a) }` block (the upkeep
    test fails); make it NPC-only, `if mobs.GetInstance(a.Id) != nil && rules.Get().Condition(a) == rules.Dead {`
    (its Alice asserts fail);
  - buff trigger: delete `death.Die(a)` in `ApplyBuffs` (the buff test fails); in `death.Die` delete
    `clearKillersAggro(victim)` (the aggro asserts fail); in `award` delete
    `user.Character.KD.AddMobKill(int(mob.MobId))` (the kill assert fails); delete
    `user.GrantXP(aw.XP, `combat`)` (the XP assert fails);
  - corpses off: in `placeDrops` delete `room.AddItem(itm, false)` (the floor-items assert fails);
    delete `room.Gold += gold` (the floor-gold assert fails); in `takeDrops` delete the
    `events.AddToQueue(events.ItemOwnership{...})` call (the `ItemOwnership` assert fails; the death
    goldens fail too, Step 16);
  - revive: in `revive` delete `victim.Character.CancelBuffsWithFlag("revive-on-death")` (the flag
    assert fails); delete the `CharacterVitalsChanged` queue (the vitals assert fails);
  - respawn: in `respawn` delete `c.CancelBuffsWithFlag(buffs.All)` (the buff assert fails); delete
    `c.Mana = 0` (the mana assert fails); delete `clear(c.PlayerDamage)` (the damage assert fails); in
    `recordPlayerDeath` delete `c.KilledBy = ids.ID{}` (the killer assert fails); delete
    `c.KD.AddEliteDeath(killerMobId, c.KillerMobName)` (the elite assert fails).

  (b) Create `internal/usercommands/death_test.go`. It holds the recovery-room helper Task 7's
  `suicide` tests also use:

```go
package usercommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryRoom makes roomId the death recovery room for one test, with its
// template in a temporary data folder (a death moves the body into a
// private copy made from it), and returns the ID of a copy of it. Every
// copy of the room is freed afterwards (freeCopies).
func recoveryRoom(t *testing.T, roomId int) (copyId int) {
	t.Helper()
	oldData := configs.GetFilePathsConfig().DataFiles.String()
	oldRecovery := int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)
	tmp := t.TempDir()
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["FilePaths.DataFiles"] = tmp
	next["SpecialRooms.DeathRecoveryRoom"] = roomId
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["FilePaths.DataFiles"] = oldData
		back["SpecialRooms.DeathRecoveryRoom"] = oldRecovery
		require.NoError(t, configs.RestoreOverrides(back))
	})
	r := rooms.LoadRoom(roomId)
	require.NotNil(t, r)
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "rooms", filepath.FromSlash(rooms.ZoneToFolder(r.Zone))), 0o755))
	require.NoError(t, rooms.SaveRoomTemplate(*r))
	copies, err := rooms.CreateEphemeralRoomIds(roomId)
	require.NoError(t, err)
	copyId = copies[roomId]
	require.True(t, rooms.IsEphemeralRoomId(copyId))
	t.Cleanup(func() { freeCopies(roomId) })
	return copyId
}

// freeCopies frees every private copy of roomId. Each copy holds one of the
// 100 ephemeral chunks (rooms/ephemeral.go) until freed, and a chunk is
// freed only when no player is in it, so the players leave first.
func freeCopies(roomId int) {
	for _, id := range rooms.FindEphemeralRoomIds(roomId) {
		if r := rooms.LoadRoom(id); r != nil {
			for _, userId := range r.GetPlayers() {
				r.RemovePlayer(userId)
			}
		}
		rooms.TryEphemeralCleanup(id)
	}
}

// corpsesOn turns corpses on for one test (the test binary loads no
// config.yaml, so they are off), as setPvP sets PvP.
func corpsesOn(t *testing.T) {
	t.Helper()
	old := bool(configs.GetGamePlayConfig().Death.CorpsesEnabled)
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["GamePlay.Death.CorpsesEnabled"] = true
	require.NoError(t, configs.RestoreOverrides(next))
	require.True(t, bool(configs.GetGamePlayConfig().Death.CorpsesEnabled), "the override took")
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["GamePlay.Death.CorpsesEnabled"] = old
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

func itemIdsOf(list []items.Item) []int {
	out := []int{}
	for _, it := range list {
		out = append(out, it.ItemId)
	}
	return out
}

// A player's death leaves a corpse found by the player's name (it records
// the account: Corpse.UserId) holding what the player carried and its gold.
// "get all <name>" and "get all from <name>" take it all (spec, plan review
// round 1); "get <item> <name>" takes one item.
func TestGetFromAPlayersCorpse(t *testing.T) {
	for _, tc := range []struct {
		command   string
		wantItems []int
		wantGold  int
	}{
		{"all bob", []int{92730}, 25},
		{"all from bob", []int{92730}, 25},
		{"testgem bob", []int{92730}, 0},
	} {
		t.Run(tc.command, func(t *testing.T) {
			npc, player, room := goSetup(t)
			recoveryRoom(t, testNextRoomId)
			corpsesOn(t)
			player.Character.Name = "Bob"
			player.Character.Items = []items.Item{testItem(92730, "testgem", items.ItemSpec{Type: items.Gemstone})}
			require.Equal(t, 25, player.Character.Gold, "a new character's gold")
			death.Die(player)
			require.Len(t, room.Corpses, 1, "Bob's corpse is in the room")
			require.Equal(t, 9201, room.Corpses[0].UserId, "a player's corpse records the account")

			_, err := Get(tc.command, npc, room, 0)
			require.NoError(t, err)
			assert.Equal(t, tc.wantItems, itemIdsOf(npc.Character.Items), "what the taker carries")
			assert.Equal(t, tc.wantGold, npc.Character.Gold, "the taker's gold")
			assert.Empty(t, room.Corpses[0].Items, "the item left the corpse")
		})
	}
}

// Any other "get all ..." keeps today's room-wide meaning (spec, plan review
// round 1): a word that names no corpse, or more words than "all [from]
// <corpse>" ("all junk bob" names Bob's corpse last), takes everything on
// the floor and leaves Bob's corpse alone.
func TestGetAllWithoutACorpseTakesTheFloor(t *testing.T) {
	for _, command := range []string{"all ground", "all junk", "all from junk", "all junk bob"} {
		t.Run(command, func(t *testing.T) {
			npc, player, room := goSetup(t)
			recoveryRoom(t, testNextRoomId)
			corpsesOn(t)
			player.Character.Name = "Bob"
			player.Character.Items = []items.Item{testItem(92730, "testgem", items.ItemSpec{Type: items.Gemstone})}
			death.Die(player)
			require.Len(t, room.Corpses, 1, "Bob's corpse is in the room")
			room.AddItem(testItem(92732, "teststick", items.ItemSpec{Type: items.Object}), false)
			room.Gold = 5

			_, err := Get(command, npc, room, 0)
			require.NoError(t, err)
			assert.Equal(t, []int{92732}, itemIdsOf(npc.Character.Items), "the floor's item, not the corpse's")
			assert.Equal(t, 5, npc.Character.Gold, "the floor's gold, not the corpse's")
			assert.Empty(t, room.Items, "the floor is empty")
			assert.Equal(t, []int{92730}, itemIdsOf(room.Corpses[0].Items), "the corpse is untouched")
			assert.Equal(t, 25, room.Corpses[0].Gold)
		})
	}
}

// The corpse panel shows what a corpse holds (the CorpseItems gate is gone).
func TestLookAtACorpseShowsWhatItHolds(t *testing.T) {
	npc, player, room := goSetup(t)
	corpsesOn(t)
	mobs.GetInstance(npc.Id).MobId = 92801 // an NPC corpse is found by name only with a MobId (rooms.go:945-990)
	npc.Character.Items = []items.Item{testItem(92731, "testpebble", items.ItemSpec{Type: items.Gemstone})}
	death.Die(npc)
	require.Len(t, room.Corpses, 1)

	told := playerTexts(t, player, func() {
		_, err := Look("rat corpse", player, room, 0)
		require.NoError(t, err)
	})
	assert.Contains(t, strings.Join(told, ""), "testpebble", "the panel lists the corpse's item")
}
```

  Check before relying on them: `goSetup`'s NPC is a bare test NPC with 0 gold and no items, so what
  it holds afterwards came from the corpse; `StoreItem` has no capacity check
  (`character.go:538-548`); corpse lookup is case-insensitive (`util.go` `stringMatch`); the look test
  keeps the shipped data folder (no `recoveryRoom`, so the panel templates load). Run
  `go test ./internal/usercommands -run 'FromAPlayersCorpse|WithoutACorpse|LookAtACorpse' -count=1 -v`;
  all pass. (`get all ground`: today's code strips `ground` and runs the room-wide branch only because
  `get.go:40` returns first; the new condition keeps that.)

  Probes, one at a time (paste each failure line):
  - put back `if args[0] == "all" {` at `get.go:40`: the two `all` subtests of
    `TestGetFromAPlayersCorpse` fail (the room-wide branch takes nothing from the corpse);
  - make `namesACorpse` return `len(args) >= 2` (any word after `all` leaves the room-wide branch):
    the four `TestGetAllWithoutACorpseTakesTheFloor` subtests fail (the taker gets nothing);
  - in `namesACorpse` delete the length check (`if len(args) != 2 && ... { return false }`): the
    `all junk bob` subtest fails (the corpse branch then reads `rest` as `all junk` and takes nothing);
  - in `get.go` put back the gate as ``if containerName == `` && false {``: all three subtests fail;
  - in `death.Die` delete `corpse.UserId = user.UserId`: the `UserId` assert fails (the corpse could
    not be found by name);
  - in `look.go` delete `Tell(actor, buildCorpseInventoryPanel(&corpse))`: the look test fails.

  (c) The recovery copies are freed: `go test ./internal/hooks ./internal/usercommands -count=16`
  prints `ok` twice (every test death takes an ephemeral chunk; about ten player deaths a run in
  `hooks` would exhaust the 100 chunks by the eleventh run without `freeCopies`).

  (d) The parts of `death.Die` the death goldens pin. Run these **after Step 15** (the goldens hold
  the new records then); each makes `go test ./internal/hooks -run 'TestCombatGoldenDeath' -count=1`
  fail (paste the line), never with `-update`:
  - in `recordPlayerDeath` delete `c.KD.AddPvpDeath()` (`death_pvp`, Bob's `kd=`); delete
    `u.Character.KD.AddPlayerKill(user.UserId, c.Name)` (`death_pvp`, Alice's `pk=`); delete
    `c.KD.AddMobDeath()` (`death_player`);
  - in `award` delete `user.Character.KD.AddEliteKill(...)` (`death_npc`, Alice's `kd=`);
  - in `killAwards` delete the solo `if mob.IsElite { ... }` block (`death_npc`, the XP line);
  - in `placeDrops` delete the `events.AddToQueue(events.MobItemDrop{...})` call (`death_npc`, the
    `MobItemDrop` notes);
  - in `end` delete `rooms.NPCEnded(mob, false)` (`TestCombatGoldenDeathNPC`'s ledger assert); delete
    `room.RemoveMob(mob.Character.Id)` (its "the rat left the room" assert).

- [ ] **Step 12: Smoke (Q5).** `cmd/smoke/main.go:124` `fmt.Sprintf("Smk%sc0", tag)` becomes
  `fmt.Sprintf("Smk%sc000", tag)`; `:131` `char: fmt.Sprintf("Smk%sc%d", tag, i),` becomes
  `char: fmt.Sprintf("Smk%sc%03d", tag, i), // fixed width: no name is a prefix of another (corpse lookup)`.
  In `cmd/smoke/script.go`, `:221-229` (the comment from `// Death, with the Phase 4 stand-in` through
  `s.field("death: still holds the item", ...)`) becomes:

```go
	// Death (Phase 5 slice 3, decision 11): suicide drops what the bot carries
	// (the ring and its gold) into its corpse in the start room and brings it
	// back at 1 health in a private copy of the recovery room (75, "Waiting
	// room"), whose portal leads back to the start room. Walking out proves
	// the bot is not downed (go is refused while downed); taking the ring
	// from its own corpse proves the drop went into the corpse.
	s.do("death: suicide", "suicide", b.char+" has DIED!")
	s.do("death: in the recovery room", "look", recoveryRoomTitle)
	s.do("death: walk out", "portal", startRoomTitle)
	s.do("death: take the item from the corpse", "get "+smokeItem+" "+b.char,
		"You take the "+smokeItem+" from the "+b.char+" corpse")
	s.field("death: holds the item again", "inventory", "Carrying:", smokeItem)
```

  Check before relying on it: the ring is carried, never worn (`script.go:183-188`, `:219`: `get`, then
  "Carrying:"); `get <item> <word>` finds the corpse by the last word (`get.go:90-100`, now without the
  `CorpseItems` gate) and says `You take the <item> from the <name> corpse.` (`get.go:241-250`; markup
  between the words is stripped by the smoke client, as for "You pick up the"). The shipped config has
  `CorpsesEnabled: true`. If the get wording differs, use what `get.go` sends and say so.

- [ ] **Step 13: Build.** `go build ./...`; `gofmt -w` every changed Go file. Prove the old death code
  is gone (the zone names were backtick literals):

```bash
grep -rnE 'npcDeath|`Shadow Realm`|`Training`' --include=*.go internal   # prints nothing
grep -rn 'func Die' --include=*.go internal                              # control: internal/death/death.go
```

  Then:
  `grep -rn 'suicide' --include=*.go internal | grep -v _test.go` prints only `usercommands.go`'s
  registry line (`suicide`: `{Suicide, true, false}`) and comments; no `Command(` and no `InputText:`
  (control: `grep -rn 'equip %s' --include=*.go internal/hooks` prints `NewRound_DoCombat.go`). Task 8's
  guard makes this permanent.

- [ ] **Step 14: Which tests fail, before any `-update`.** `go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'`:
  besides the 3 known Windows failures, exactly these fail, each on a golden difference:
  `TestCombatGoldenPvMToDeath`, `TestCombatGoldenMvM`, `TestCombatGoldenMvP`, `TestCombatGoldenPvP`,
  `TestCombatGoldenHiddenAttacker`, `TestCombatGoldenDeathNPC`, `TestCombatGoldenDeathPlayer`,
  `TestCombatGoldenDeathPvP`, `TestReadsH_Outside`. Any other failure, or one of these failing on
  something other than a golden line: stop and report.

- [ ] **Step 15: Re-record the nine, once.**
  `go test ./internal/hooks -run 'TestCombatGolden(PvMToDeath|MvM|MvP|PvP|HiddenAttacker|DeathNPC|DeathPlayer|DeathPvP)$|TestReadsH_Outside$' -count=1 -update`.
  Then `git diff --stat -- internal/hooks/testdata/` lists exactly `combat-golden/death_npc.txt`,
  `death_player.txt`, `death_pvp.txt`, `hidden_attacker.txt`, `mvm.txt`, `mvp.txt`, `pvm_death.txt`,
  `pvp.txt` and `reads-golden/H-outside.txt`. Then `go test ./internal/hooks -count=1` twice: `ok`.

- [ ] **Step 16: Read every diff and sort every hunk.** For each file, `git diff -U3 -- <file>` (read
  it all). Each hunk is **decided** only if it is one of the changes listed for that file below;
  anything else is **unexpected** and stops the work (report it; do not commit). Write the summary
  (file, decided hunks in one line each, unexpected: none) into the commit message body; the controller
  copies it into the PR.

  Expected, all decided:
  - Every combat golden: rounds before the death round are byte-identical (same swings, same draws).
    A changed line before the death round is unexpected.
  - `pvm_death.txt`, `mvm.txt` (decision 7; the one death routine): in the death round the event
    `Input rat: "suicide"` becomes `MobDeath rat`; Alice reads `rat has died.` and
    `25 gold drops to the ground.` (corpses are off in these fixtures), in `pvm_death` also her
    `You gained N experience points! (combat)` line between them; a new state line
    `room 93001 floor gold: 25`.
  - `mvp.txt`, `pvp.txt`, `hidden_attacker.txt` (decision 7: die at 0, no downed state): in the round
    the player went down (`mvp` round 6, `pvp` round 4, `hidden_attacker` round 6 today), the event
    `PlayerDrop <name> room=93001` becomes `PlayerDeath <name>` followed by
    `CharacterVitalsChanged <name>` (the respawn's); every player, the dead one included, reads the
    world notice `*** <name> has DIED! ***` (`(killed by Alice)` in `pvp` and `hidden_attacker`, which
    name the attacker as today's notice did), with no `\r` at the line end; the other players in room A
    read `<name> has died.` (the dead player does not: the line has no actor text); **every** player
    in room A, the dead one included, reads `<ansi fg="yellow-bold">25 gold</ansi> drops to the ground.`
    (corpses are off in these fixtures, and the line is sent to the room before the respawn moves the
    body, `placeDrops`); the dead player's state reads `room=copy-of-93003 hp=1/<max>` (`1/12` in
    `mvp`, `1/15` in `pvp`, `1/10` in `hidden_attacker`) with `playerdamage={}` and no `killedby`;
    worn items stay on the player; the room gains `room 93001 floor gold: 25` (in `pvp` after the
    existing `room 93001 floor: 0`, the broken shield). Nothing before the death round changes.
  - `death_npc.txt` (Q2, the `CorpseItems` removal, ruling 3): the `Input rat: "suicide"` and
    `Ran rat: "suicide"` lines go, `MobDeath rat` stays, and the `MobItemDrop mob=93201 ...` notes keep
    their item IDs and order (the same worn-item decisions; Q7e: they are now queued for corpse drops
    too); new `ItemOwnership rat item=N gained=false` notes appear, one per dropped item, in the same
    IDs and order as the `MobItemDrop` notes and before them (`takeDrops` runs before `placeDrops`;
    today's NPC death queued none: spec, plan review round 1); Alice's XP and alignment lines are **unchanged** (this is the draw-order evidence: a different
    XP number is unexpected), and so is her `kd=1/0/0/0/1/0` (the kill and the elite kill, now
    credited by `death.award`); the `drops to the ground` lines and the `room 93001 floor:` and
    `floor gold:` lines go; the corpse line gains exactly the items that were on the floor (same IDs)
    and `gold=30`, and its `worn=` is empty (Q2: today it lists the rat's worn items). The rat's own
    state line changes too: today the old code left everything on the rat's body, so it reads
    `items=93108` and `gold=30` with the worn items unchanged; now `items=` is gone (the gem left the
    body), it reads `gold=0`, and `worn=` (and `weapon=` if the club dropped) loses exactly the worn
    items that went into the corpse (`worn=[]` if both dropped).
  - `death_player.txt`, `death_pvp.txt` (decisions 7 and 11, ruling 1): the player dies in the
    combat round where it went down (no `PlayerDrop`; `PlayerDeath` and the respawn's
    `CharacterVitalsChanged` instead); every `== heal N` section goes (no bleed-out); the `Ran ...
    "suicide"` line, the Remove and Drop lines and `You lost 100 experience points` go, and so do the
    `room 93001 floor:` and `floor gold:` lines. The `ItemOwnership <name> item=N gained=false` notes
    the old `drop` queued (`drop.go:108`, one per item its rolls dropped, in the heal section that ran
    `suicide`) are replaced by exactly one, `ItemOwnership <name> item=93108 gained=false` (the carried
    gem, queued by `takeDrops`), in the death round after `PlayerDeath <name>` and before the respawn's
    `CharacterVitalsChanged <name>`; no worn item's note appears (worn gear stays). In `death_pvp` Alice reads
    `Bob has died.` in that round rather than in a heal section. The dead player's state line: no
    `items=` (the gem is in the corpse, whatever its 50% roll did today), `gold=0` (in the corpse,
    not on the floor), `xp=1000` (no penalty), every worn item kept (`weapon=93101` stays in
    `death_player`, `worn=` as at the start), the same `kd=` as today (`0/1/0/0/0/1` for Alice in
    `death_player`, the elite death; `0/0/0/1/0/0 pk=[] pd=[9301:Alice=1]` for Bob and
    `0/0/1/0/0/0 pk=[9302:Bob=1] pd=[]` for Alice in `death_pvp`), and
    `room=copy-of-93003 hp=1/<max>`. The corpse line holds the carried gem and `gold=25`, `worn=`
    empty. The player reads the notice (no `\r`).
  - `H-outside.txt` (the `CorpseItems` removal): in "NPC death" the two
    `<ansi fg="yellow-bold">25 gold</ansi> drops to the ground.` lines go (the gold is in the rat's
    corpse); in "follow module: Bob stops following", Ann's look loses the two lines
    `Ann: ...On the Ground: ...25 gold...` and the `Ann|` after it; in "player death" the lines
    `Bob: You drop ... 25 gold ... on the floor.` and `Ann: ... Bob ... drops ... 25 gold ...` go.
    Every other line is identical.

- [ ] **Step 17: Smoke.** `go run ./cmd/smoke` and `go run ./cmd/smoke -bots 5`; paste the summary
  lines. Then `go run ./cmd/smoke -bots 12` (two-digit numbers, the Q5 case); paste its summary line.

- [ ] **Step 18:** Per-task check (after the commit, the testdata and `_datafiles` lines print
  nothing). Commit every file in **Files** that exists now (name only paths that exist: `git rm` staged
  the two removals):
  `Phase 5.3: one death routine (internal/death) and OnDeath; XP and alignment move to dikubasic; permadeath, lives, penalties, CorpseItems and the bleed-out go (goldens change: summary below)`,
  with Step 16's summary as the body.

### Task 7: `suicide` refuses in a fight and in the recovery room (Sonnet)

**Files:** modify `internal/usercommands/suicide.go`; create `internal/usercommands/suicide_test.go`.

- [ ] **Step 1: Write the failing tests** `internal/usercommands/suicide_test.go`:

```go
package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryRoom (the recovery room for one test, its copies freed after) is
// in death_test.go (Task 6).

// Nobody escapes a fight except by fleeing; dying by one's own hand is an
// escape.
func TestSuicideRefusedInAFight(t *testing.T) {
	npc, player, room := goSetup(t)
	recoveryRoom(t, testNextRoomId)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	health := player.Character.Health

	var deaths []events.PlayerDeath
	told := playerTexts(t, player, func() {
		deaths = capture[events.PlayerDeath](t, func() {
			_, err := Suicide("", player, room, 0)
			require.NoError(t, err)
		})
	})
	assert.Equal(t, []string{"You're too busy to die right now!\n"}, told)
	assert.Empty(t, deaths)
	assert.Equal(t, health, player.Character.Health)
	assert.Equal(t, testRoomId, player.Character.RoomId)
}

func TestSuicideRefusedInTheRecoveryRoom(t *testing.T) {
	_, player, room := goSetup(t)
	recoveryRoom(t, testRoomId)

	told := playerTexts(t, player, func() {
		_, err := Suicide("", player, room, 0)
		assert.EqualError(t, err, "already dead")
	})
	assert.Equal(t, []string{"You're already dead!\n"}, told)
}

// The private copy a death puts the body in is the recovery room too.
func TestSuicideRefusedInACopyOfTheRecoveryRoom(t *testing.T) {
	_, player, _ := goSetup(t)
	copyId := recoveryRoom(t, testNextRoomId)
	player.Character.RoomId = copyId

	told := playerTexts(t, player, func() {
		_, err := Suicide("", player, rooms.LoadRoom(copyId), 0)
		assert.EqualError(t, err, "already dead")
	})
	assert.Equal(t, []string{"You're already dead!\n"}, told)
}

func TestSuicideDiesOtherwise(t *testing.T) {
	_, player, room := goSetup(t)
	recoveryRoom(t, testNextRoomId)

	deaths := capture[events.PlayerDeath](t, func() {
		_, err := Suicide("", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, deaths, 1)
	assert.Equal(t, 1, player.Character.Health)
	assert.True(t, rooms.IsEphemeralRoomId(player.Character.RoomId))
	assert.Equal(t, testNextRoomId, rooms.GetOriginalRoom(player.Character.RoomId), "in a copy of the recovery room")
}
```

  Check before relying on it: `capture[T]` (`items_test.go:48-63`) drains the queue before and after
  `fn`; nested inside `playerTexts` it still sees the `PlayerDeath` (both register listeners before
  `fn`). If the nesting misbehaves, capture the two in separate calls and say so.

- [ ] **Step 2: Run; they fail.** `go test ./internal/usercommands -run Suicide -count=1 -v`: the three
  refusal tests fail (the player dies); `TestSuicideDiesOtherwise` passes.
- [ ] **Step 3: `suicide.go`** becomes:

```go
package usercommands

import (
	"errors"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// Suicide is a death by the actor's own hand, through the one death routine
// (internal/death), for players and NPCs alike. It is refused in a fight
// (nobody escapes one except by fleeing, as quit is refused) and in the
// recovery room or a copy of it.
func Suicide(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	if inFight(actor, room) {
		Tell(actor, "You're too busy to die right now!")
		return true, nil
	}

	if rooms.IsDeathRecoveryRoom(actor.Character.RoomId) {
		Tell(actor, `You're already dead!`)
		return true, errors.New(`already dead`)
	}

	death.Die(actor)
	return true, nil
}
```

- [ ] **Step 4: Run; they pass.** Probes: delete the `inFight` check (the fight test fails); replace
  `rooms.IsDeathRecoveryRoom(actor.Character.RoomId)` with
  `actor.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)` (add the `configs`
  import for the probe; the copy test fails, the room test passes); delete the recovery check (both
  recovery tests fail); delete `death.Die(actor)` (`TestSuicideDiesOtherwise` fails). Paste the four
  lines.
- [ ] **Step 5:** Per-task check (goldens unchanged: no golden runs `suicide` in a fight or the recovery
  room). `go test ./internal/usercommands -run Suicide -count=16` prints `ok` (each test's copies are
  freed). Commit both files: `Phase 5.3: suicide is refused in a fight and in the recovery room (and its copies)`.

### Task 8: Guards (Sonnet)

**Files:** create `internal/guards/death_test.go`.

- [ ] **Step 1: Write the guards:**

```go
package guards

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// deathScanned is the engine's round and command code, where the slice 3
// death guards apply. characters, rooms and the rulesets are outside it
// (slice 4 moves the characters checks; the import cycle keeps them there
// for now).
var deathScanned = []string{"internal/hooks/", "internal/usercommands/", "internal/death/"}

// itemIsDisabled lists the scanned files that call items.Item.IsDisabled
// (a disabled equipment slot, not a body's condition), with why. A file on
// it may still not call .Character.IsDisabled().
var itemIsDisabled = map[string]string{
	"internal/usercommands/look.panels.go":      "the equipment and corpse panels skip a disabled slot",
	"internal/usercommands/inventory.panels.go": "the inventory panel skips a disabled slot",
}

// plantedDeath is the death guards' own control: source with exactly what
// each guard must catch. It is parsed, never compiled.
const plantedDeath = `package planted

func f(a, c, m, u, itm any) bool {
	m.Command("suicide")
	u.CommandFlagged(` + "`suicide`" + `, 0)
	events.AddToQueue(events.Input{ActorId: a, InputText: "suicide"})
	_ = itm.IsDisabled()
	return a.Character.IsDisabled() || c.Health <= -10 || 1 > c.Health || c.Health == (0) || c.Health > c.HealthMax
}
`

func parsePlanted(t *testing.T) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "planted.go", plantedDeath, 0)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func inDeathScope(path string) bool {
	for _, p := range deathScanned {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// TestEngineAsksTheRulesetWhoIsDead enforces Phase 5 slice 3: the engine
// compares no health with a number and never asks a body whether it is
// disabled; it asks rules.Get().Condition.
func TestEngineAsksTheRulesetWhoIsDead(t *testing.T) {
	files, healthCompares := 0, 0
	itemCalls := map[string]bool{}
	for path, f := range goFiles(t) {
		if !inDeathScope(path) {
			continue
		}
		files++
		literal, compares := healthLiteralCompares(f)
		healthCompares += compares
		for range literal {
			t.Errorf("%s: compares .Health with a number; ask rules.Get().Condition (slice 3)", path)
		}
		for _, onCharacter := range isDisabledCalls(f) {
			if onCharacter {
				t.Errorf("%s: calls .Character.IsDisabled(); ask rules.Get().Condition (slice 3)", path)
				continue
			}
			if _, ok := itemIsDisabled[path]; !ok {
				t.Errorf("%s: calls .IsDisabled(); a body's: ask rules.Get().Condition (slice 3); an item's: add the file to itemIsDisabled with why", path)
				continue
			}
			itemCalls[path] = true
		}
	}
	for path, why := range itemIsDisabled {
		if !itemCalls[path] {
			t.Errorf("stale allowlist: %s (%s) no longer calls IsDisabled; remove it from itemIsDisabled", path, why)
		}
	}
	// Positive controls: the walk reached the packages and saw a .Health
	// comparison (with a variable: the regen pass's change check) ...
	if files < 50 {
		t.Errorf("control: scanned only %d files under %v; the guard may be scanning the wrong place", files, deathScanned)
	}
	if healthCompares == 0 {
		t.Error("control: no .Health comparison seen in the scanned packages; the guard may not be matching")
	}
	// ... and the matchers find exactly what is planted: three literal
	// compares (signed, reversed, in brackets) and not the compare with a
	// variable; one body IsDisabled call and one other.
	planted := parsePlanted(t)
	if literal, compares := healthLiteralCompares(planted); len(literal) != 3 || compares != 4 {
		t.Errorf("control: found %d literal health compares of 3 and %d health compares of 4 in the planted source; the literal check is broken", len(literal), compares)
	}
	onCharacter, other := 0, 0
	for _, c := range isDisabledCalls(planted) {
		if c {
			onCharacter++
		} else {
			other++
		}
	}
	if onCharacter != 1 || other != 1 {
		t.Errorf("control: found %d .Character.IsDisabled calls of 1 and %d other IsDisabled calls of 1 in the planted source", onCharacter, other)
	}
}

// TestNothingQueuesSuicide enforces Phase 5 slice 3: a death runs
// death.Die; no code queues the suicide command as input.
func TestNothingQueuesSuicide(t *testing.T) {
	inputs, commands := 0, 0
	for path, f := range goFiles(t) {
		found, in, cmd := suicideQueues(f)
		inputs += in
		commands += cmd
		for range found {
			t.Errorf("%s: queues the suicide command (an events.Input or a Command call); a death runs death.Die", path)
		}
	}
	// Positive controls: inputs and command calls exist (Account.Command
	// builds an events.Input; DoCombat queues "equip") ...
	if inputs == 0 {
		t.Error("control: no events.Input literal seen; the guard may not be matching")
	}
	if commands == 0 {
		t.Error("control: no Command call seen; the guard may not be matching")
	}
	// ... and the matcher finds the three planted suicides (Command,
	// CommandFlagged with a raw string, an events.Input).
	if found, _, _ := suicideQueues(parsePlanted(t)); len(found) != 3 {
		t.Errorf("control: found %d of the 3 planted suicides; the suicide check is broken", len(found))
	}
}

// healthLiteralCompares returns where f compares a .Health selector with a
// number literal, and how many .Health comparisons it has in all.
func healthLiteralCompares(f *ast.File) (literal []token.Pos, compares int) {
	ast.Inspect(f, func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok || !isComparison(be.Op) {
			return true
		}
		l, r := isHealth(be.X), isHealth(be.Y)
		if !l && !r {
			return true
		}
		compares++
		if (l && isNumber(be.Y)) || (r && isNumber(be.X)) {
			literal = append(literal, be.Pos())
		}
		return true
	})
	return literal, compares
}

// isDisabledCalls returns one entry per X.IsDisabled() call in f: true when
// X is a .Character selector (a body's IsDisabled, characters.Character).
// The guard sees syntax, not types: a body reached through another name
// (char := a.Character; char.IsDisabled()) reads as "other", which is
// allowed only in the itemIsDisabled files.
func isDisabledCalls(f *ast.File) []bool {
	out := []bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "IsDisabled" {
			return true
		}
		recv, isSel := sel.X.(*ast.SelectorExpr)
		out = append(out, isSel && recv.Sel.Name == "Character")
		return true
	})
	return out
}

// suicideQueues returns where f queues the suicide command, and how many
// events.Input literals and Command calls it has in all.
func suicideQueues(f *ast.File) (found []token.Pos, inputs, commands int) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			sel, ok := x.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Input" {
				return true
			}
			if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "events" {
				return true
			}
			inputs++
			for _, el := range x.Elts {
				kv, ok := el.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "InputText" && isSuicide(kv.Value) {
					found = append(found, x.Pos())
				}
			}
		case *ast.CallExpr:
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok || (sel.Sel.Name != "Command" && sel.Sel.Name != "CommandFlagged") || len(x.Args) == 0 {
				return true
			}
			commands++
			if isSuicide(x.Args[0]) {
				found = append(found, x.Pos())
			}
		}
		return true
	})
	return found, inputs, commands
}

func isComparison(op token.Token) bool {
	switch op {
	case token.LSS, token.LEQ, token.GTR, token.GEQ, token.EQL, token.NEQ:
		return true
	}
	return false
}

// isHealth reports whether e is a .Health selector (a body's health).
func isHealth(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Health"
}

// isNumber reports whether e is a number literal, signed or in brackets.
func isNumber(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.BasicLit:
		return x.Kind == token.INT || x.Kind == token.FLOAT
	case *ast.UnaryExpr:
		return (x.Op == token.SUB || x.Op == token.ADD) && isNumber(x.X)
	case *ast.ParenExpr:
		return isNumber(x.X)
	}
	return false
}

// isSuicide reports whether e is the string literal suicide.
func isSuicide(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	s, err := strconv.Unquote(lit.Value)
	return err == nil && strings.TrimSpace(s) == "suicide"
}
```

- [ ] **Step 2: Run; they pass.** `go test ./internal/guards -run 'WhoIsDead|QueuesSuicide' -count=1`.
- [ ] **Step 3: Show they can fail.** Create `internal/hooks/zz_probe.go` (Write):

```go
package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
)

func zzProbe(a *actors.Actor, c *characters.Character, m *mobs.Mob) bool {
	m.Command(`suicide`)
	events.AddToQueue(events.Input{InputText: "suicide"})
	return a.Character.IsDisabled() || c.Health <= -10 || 1 > c.Health
}
```

  Run; paste the failure lines: three from `TestEngineAsksTheRulesetWhoIsDead` (both literal sides and
  the `.Character.IsDisabled()` call) and two from `TestNothingQueuesSuicide`. Delete the probe file
  (`rm internal/hooks/zz_probe.go`; it is untracked), rerun; `ok`. Then the controls, one at a time
  (each fails; paste the line; restore):
  - `isNumber` returns `false` always (its body becomes `return false`): the planted-source control
    of `TestEngineAsksTheRulesetWhoIsDead` fails ("found 0 literal health compares of 3");
  - `isSuicide` returns `false` always: the planted control of `TestNothingQueuesSuicide` fails
    ("found 0 of the 3 planted suicides");
  - in `isDisabledCalls`, `recv.Sel.Name == "Character"` to `recv.Sel.Name == "Body"`: the guard then
    reads the planted body call as an item's, and the IsDisabled control fails;
  - add `"internal/usercommands/go.go": "probe",` to `itemIsDisabled`: the stale-allowlist check fails
    (`go.go` calls no `IsDisabled` since Task 3);
  - `files < 50` to `files < 5000`: the walk control fails.
- [ ] **Step 4:** Per-task check; `git status --short` shows only `internal/guards/death_test.go`.
  Commit it: `Phase 5.3: guards: the engine asks the ruleset who is dead (no health literal, no body IsDisabled); nothing queues suicide`.

### Task 9: Stub-ruleset tests (Sonnet)

Test files only: a ruleset can add a downed state, revive, fix the drops and set regen with no engine
edit.

**Files:** modify `internal/hooks/ruleset_stub_test.go`, `internal/usercommands/ruleset_stub_test.go`.

- [ ] **Step 1: `internal/hooks/ruleset_stub_test.go`.** Add `github.com/GoMudEngine/GoMud/internal/events`
  and `github.com/GoMudEngine/GoMud/internal/rooms` to the imports, and append:

```go
// Downed is generic engine handling: under a ruleset with a downed state
// (downedAtZero, combat_golden_test.go) the round downs Alice instead of
// killing her, the room reads her drop, and nobody dies.
func TestStubRulesetDownedStateIsNotDeath(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	id := events.RegisterListener(events.PlayerDrop{}, HandlePlayerDrop)
	t.Cleanup(func() { events.UnregisterListener(events.PlayerDrop{}, id) })
	t.Cleanup(rules.SetForTest(downedAtZero{rules.Get()}))
	w.start()
	w.rounds(20)

	out := w.out.String()
	require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "not moved to the recovery room")
	require.Contains(t, out, "PlayerDrop Alice")
	require.Contains(t, out, `you drop to the ground!`)
	require.NotContains(t, out, "PlayerDeath")
}

// reviver is the active ruleset where every death is a revive at 7 health.
type reviver struct{ rules.Ruleset }

func (reviver) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{Revive: true, ReviveHealth: 7}
}

func TestStubRulesetRevivesInsteadOfDeath(t *testing.T) {
	w := newCombatWorld(t, 111)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "revive-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(reviver{rules.Get()}))
	w.start()
	w.rounds(20)

	out := w.out.String()
	require.Equal(t, 7, rat.Character.Health)
	require.NotNil(t, mobs.GetInstance(rat.Character.Id), "the rat lives on")
	require.Contains(t, out, "is suddenly revived in a shower of sparks!")
	require.NotContains(t, out, "MobDeath")
}

// fixedDrops is the active ruleset where a death drops exactly drop and 3
// gold, and the body ends.
type fixedDrops struct {
	rules.Ruleset
	drop items.Item
}

func (f fixedDrops) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{DropItems: []items.Item{f.drop}, DropGold: 3}
}

// The corpse holds exactly the ruleset's drops, and its body nothing (Q2):
// the rat keeps a second carried item, its ring and 27 gold, and none of
// them may show on the corpse's body copy.
func TestStubRulesetDecidesTheDrops(t *testing.T) {
	w := newCombatWorld(t, 111)
	deathConfig(t) // corpses on
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "drop-rats")
	gem := trinket(93108, "testgem", items.Gemstone, items.Mundane)
	pebble := trinket(93109, "testpebble", items.Gemstone, items.Mundane) // carried, not dropped
	rat.Character.Items = []items.Item{gem, pebble}
	rat.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	rat.Character.Gold = 30
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(fixedDrops{Ruleset: rules.Get(), drop: gem}))
	w.start()
	w.rounds(20)

	room := rooms.LoadRoom(goldenRoomA)
	require.Len(t, room.Corpses, 1)
	c := room.Corpses[0]
	require.Equal(t, "93108", itemIds(c.Items))
	require.Equal(t, 3, c.Gold)
	require.Equal(t, "93109", itemIds(rat.Character.Items), "the rat's body kept the pebble (the fixture makes the next asserts mean something)")
	require.Equal(t, 27, rat.Character.Gold, "and 27 gold")
	require.Empty(t, c.Character.GetAllWornItems(), "the corpse's body wears nothing")
	require.Empty(t, c.Character.Items, "the corpse's body carries nothing")
	require.Zero(t, c.Character.Gold, "the corpse's body holds no gold")
	require.Empty(t, room.Items, "nothing on the floor")
	require.Zero(t, room.Gold)
}

// plusFive is the active ruleset where an account regenerates 5 health
// every round, with a line.
type plusFive struct{ rules.Ruleset }

func (plusFive) Regen(a *actors.Actor, round uint64) rules.RegenResult {
	if a.Controller() == nil || a.Controller().Account() == nil {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: 5, Messages: []perception.Message{{
		Kind: perception.Interface, Actor: a.Id, Audience: perception.ActorOnly, ActorText: "STUB REGEN",
	}}}
}

// The regen pass applies the ruleset's amounts and sends its lines.
func TestStubRulesetDecidesRegen(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 20)
	alice.Character.Health = 3
	t.Cleanup(rules.SetForTest(plusFive{rules.Get()}))
	w.start()

	AutoHeal(events.NewRound{RoundNumber: 1}) // not a 3rd round: the reference ruleset gives nothing
	events.ProcessEvents()

	require.Equal(t, 8, alice.Character.Health)
	var reads, notes []string
	for _, e := range w.rec.take() {
		if e.userId == 0 {
			notes = append(notes, e.text)
		} else {
			reads = append(reads, e.text)
		}
	}
	require.Contains(t, strings.Join(reads, ""), "STUB REGEN")
	require.Contains(t, notes, "CharacterVitalsChanged Alice")
}
```

  Add `"strings"` to the imports. (`downedAtZero` is in `combat_golden_test.go`, Task 6; `trinket`,
  `itemIds`, `deathConfig` and `recoveryRoom` in the same file, Task 1.)

- [ ] **Step 2: `internal/usercommands/ruleset_stub_test.go`.** Append:

```go
// downedAlways is the active ruleset with every actor downed.
type downedAlways struct{ rules.Ruleset }

func (downedAlways) Condition(a *actors.Actor) rules.Condition { return rules.Downed }

// A downed actor cannot act: commands not allowed when downed are refused,
// though the reference ruleset calls the player (10 health) up.
func TestStubRulesetDownedActorCannotAct(t *testing.T) {
	_, player, _ := goSetup(t)
	t.Cleanup(rules.SetForTest(downedAlways{rules.Get()}))

	told := playerTexts(t, player, func() {
		_, err := TryCommand("get", "stick", player.Id, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"You are unable to do that while downed.\n"}, told)

	told = playerTexts(t, player, func() {
		_, err := TryCommand("north", "", player.Id, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"You are unable to do that while downed.\n"}, told)
	assert.Equal(t, testRoomId, player.Character.RoomId)
}
```

- [ ] **Step 3: Run.** `go test ./internal/hooks ./internal/usercommands -run Stub -count=1 -v`; all pass.
- [ ] **Step 4: Prove they test the seam**, one probe at a time (each fails, paste the line):
  - downed (hooks): delete the `case rules.Downed:` branch in `handleAffected` (no `PlayerDrop`);
    replace the test's `SetForTest` line with `_ = downedAtZero{}` (Alice dies).
  - revive: in `death.revive` change `victim.Character.Health = health` to `= 1`; replace the
    `SetForTest` line with `_ = reviver{}`.
  - drops: in `death.Die` change `drops, gold = takeDrops(c, outcome)` to
    `drops, gold = append([]items.Item{}, c.Items...), c.Gold` (the corpse holds the gem, the pebble
    and 30); in `corpseBody` delete `body.Equipment = characters.Worn{}` (the body-wears-nothing
    assert fails); delete `body.Items = nil` (the body-carries-nothing assert fails: the copy holds
    the pebble); delete `body.Gold = 0` (the body-gold assert fails: the copy holds 27).
  - regen: in `AutoHeal` change `a.Character.Heal(r.Health, r.Mana)` to `a.Character.Heal(0, r.Mana)`;
    delete the `for _, m := range r.Messages` loop (the line assert fails); replace the `SetForTest`
    line with `_ = plusFive{}`.
  - downed actor (usercommands): `usercommands.go:281` to
    `userDisabled = actor.Character.IsDisabled() && rules.Get() != nil` (asks the body again; the
    `rules.Get()` call keeps the file's only `rules` use, so it compiles; the `get` assert fails);
    `go.go:46` to `if actor.Character.IsDisabled() && rules.Get() != nil {` (the same; the `north` assert
    fails).
- [ ] **Step 5:** Per-task check. Commit the two test files:
  `Phase 5.3: stub rulesets add a downed state, revive, fix the drops and set regen with no engine edit`.
  `git show --stat HEAD` lists only the two `_test.go` files.

### Task 10: Verify, review, docs, PR (controller)

- [ ] Full check from a clean tree: `go build ./...`, `go test ./... -count=1` (3 known failures only),
  `gofmt -l .` minus vendored, `go run ./cmd/smoke`, `-bots 5` and `-bots 12`; paste the summary lines.
- [ ] `git diff master --stat -- internal/hooks/testdata/` lists the three new goldens and the Task 6
  set only.
- [ ] Opus whole-branch review (correctness, random-draw order, the death path for players and NPCs,
  missed references to the deleted knobs), briefed with the agent rules, the spec and this plan. Fixes
  in their own commits.
- [ ] **Help pages (ruling 5).** `_datafiles/world/empty/templates/help/death.template` becomes (CRLF,
  as the file):

```text
<ansi fg="black-bold">.:</ansi> <ansi fg="magenta">Help for </ansi><ansi fg="command">death</ansi>

When your health falls to 0, you die.

You drop everything you carry and all your gold into your corpse, which stays
where you fell. What you wear stays on you.

You come back in the shadow realm with 1 health, and you heal fast while you
are there. Leave through the <ansi fg="exit">portal</ansi> when you are ready.

Your corpse holds your things until it crumbles to dust after a while. When it
crumbles, everything still in it is destroyed, so go back for it. Type
<ansi fg="command">get all</ansi> and your name to take it all back (see
<ansi fg="command">help get</ansi>).

A creature's corpse holds its loot the same way, and its contents are
destroyed the same way when it crumbles.
```

  In `health.template`, the two paragraphs `At <ansi fg="red">0 HP</ansi>, you will drop to the floor and slowly bleed out.`
  and `At <ansi fg="red">-10 HP</ansi>, you will die. When you die, all gear you posess drops to the ground.`
  become one: `At <ansi fg="red">0 HP</ansi>, you die. See <ansi fg="command">help death</ansi>.`.
  In `get.md`, after the `~get stick from stash~` usage pair add:

```text
  ~get all bob~  
  This would take everything from Bob's corpse (~get all from bob~ works too).  
  ~get stick bob~  
  This would take a stick from Bob's corpse.

A corpse holds what its owner dropped until it crumbles to dust. When it crumbles, everything still in it is destroyed.
```

  (These lines describe Task 6 Step 7b: `get all <corpse>`, `get all from <corpse>` and
  `get <item> <corpse>` are each tested in `TestGetFromAPlayersCorpse`; plain `get all` stays the
  room-wide take (reads golden D), and so do the other forms (`TestGetAllWithoutACorpseTakesTheFloor`). `death.template` above says
  the same in words; the `config.yaml` comment, Task 6 Step 8, names the same three forms.)

  Check: `go test ./internal/hooks -run TestReads -count=1` still `ok` (no reads golden covers these
  pages).
- [ ] Strip plan decision 12 (`gomud_strip_plan.md`, with the other decisions): "Move with parity,
  then shrink" (decision 1) covers code that survives the shrink. Code the shrink deletes goes straight
  to its bare form, with a golden recorded first so the change is visible (Calabe, 2026-11-10, slice 3).
- [ ] Phase design "Notes for later slices": slice 4 takes NPC regen (one-actor rule), the -5/-10
  floors in `ApplyHealthChange`, XP, levels and alignment (decay included, still in upkeep), the four
  `characters` health checks and `FindDowned`, offhand breakage and `BackStab`, and `actors.BodyReplaced`
  (no production caller since permadeath went). Remove what this slice settled (round, regen, death),
  including slice 6's sentence "The Death Recovery affect (heal in the recovery room) is rebuilt in Go
  here." (`:153-154`): ruling 2 made it a regen rule and Task 5 deleted the buff and its mutator.
- [ ] `PATCH_NOTES.md` (player-visible): you die at 0 health (no bleeding out); a death drops what you
  carry and your gold into your corpse and you keep what you wear, at any level; no XP loss, no lives,
  no permadeath; you come back at 1 health and the shadow realm heals 1d10 health and mana a round;
  corpses hold creatures' loot too, and anything left in a corpse is destroyed when it crumbles;
  `get all <corpse>` (and `get all from <corpse>`) takes everything from a corpse; you
  cannot `suicide` in a fight or in the recovery room; nobody can attack from inside the recovery room's
  copy; a creature with revive-on-death and a player with it both get up at full health; mana regen
  uses the mana amount. Admin: the death, lives and corpse-items settings and the Extra Lives field are
  gone.
- [ ] `status.md`: slice 3 done; Watch: `CreateEphemeralRoomIds` panics when the recovery room is 0
  (Q4; unreachable with the shipped 75). `geomud-verifying`: the smoke flow's step 2 (death now drops
  the ring into the bot's corpse; the bot takes it back after the portal; bot names are
  `Smk<tag>c000`...), and the ok-package count (`internal/death` has no test file, so it stays 40 unless
  the run says otherwise: count it).
- [ ] Commit the docs and help pages: `git add` the three help files, `PATCH_NOTES.md`, `status.md`,
  `gomud_strip_plan.md`, `2026-11-06-phase-5-rules-seam-design.md`, `.claude/skills/geomud-verifying/SKILL.md`;
  message `Phase 5.3 done: help pages, patch notes, status, decision 12, notes for later slices`.
- [ ] Push `strip/phase-5-3-round-death`, open the PR on `pruuk/GeOMUD` (`gh pr create --repo pruuk/GeOMUD`)
  with Task 6's golden summary, merge once CI is green (memory `claude-merges-prs`); move the spec and
  this plan to `docs/completed/`.
