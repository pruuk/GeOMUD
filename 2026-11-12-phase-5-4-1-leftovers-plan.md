# Phase 5 slice 4.1: round and death leftovers (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved spec `2026-11-12-phase-5-4-1-leftovers-spec.md`: `characters` asks the
ruleset whether a body is up through one func variable (`characters.IsUp`); no health floors in the
engine; every actor regenerates by one rule; `Character.IsDisabled`, `rooms.FindDowned`,
`actors.BodyReplaced`, offhand breakage and `BackStab` are deleted; the death guard covers
`characters` and `rooms`; and (rulings 6-9, Task 6a) a body that is attacked fights back, and a body
whose target is gone turns to its longest attacker.

**Architecture:** See the spec ("Design"). Every design decision this plan implements is in the spec:
its rulings, its "Retaliation and target switching" section, and its "Decisions taken without asking"
rows, including those the plan reviews raised (marked "plan Q*n*" or "plan review round *n*"), such
as the per-turn `CanAttack` re-check on the current target, `Aggro.OutOfReach` saved in NPC
sidecars, `shootingExit`'s exit preference, unseen targets counting toward the wait, and the
recovery room refusing attackers in it. Where the spec leaves a choice to the plan (its five open
items), the evidence and the choice are under "Settled from the code". The six questions found while planning (Q1 to Q6) were answered by the controller on
2026-11-12 and the spec records each ("Settled while planning" rows of its "Decisions taken without
asking"); the steps that implement an answer cite it as "(Q*n*)". Task 6a (rulings 6-9, added with
the spec's amendment) raised Q7 to Q15, answered by the controller the same day and built into
Task 6a, and then Q16 and Q17, answered too (spec ruling 10 and the "plan Q17" Decisions row).
Q18 (a smoke timeout that predates this slice) is answered: no code for it; ruling 11's pinned rats
remove it in launch mode; Task 7 adds a status Watch line for `-addr` mode.

**Tech stack:** Go, testify, the combat and reads goldens in `internal/hooks`, `internal/guards` (Go
syntax tree), `cmd/smoke`.

Branch `strip/phase-5-4-1-leftovers` (exists; holds the spec at `1c8e607f`). Tasks run one after
another on the branch: they share files. No parallel worktrees.

Every task below was applied, in order, to a scratch copy of `1c8e607f` (`git archive`) while this
plan was written: the builds, every golden diff stated below, every new test and every mutation
probe were run there (2026-11-12). The repo was not touched.

---

## Facts re-verified (2026-11-12, at `1c8e607f`, beyond the spec's table)

| Fact | Value | Where |
|---|---|---|
| Guard walk skips test files | `goFiles` returns only non-`_test.go` files | `guards/guards_test.go:30` |
| Guard matches selectors only | `isHealth` is true only for a `.Health` selector; a local such as `newHealth` in `ApplyHealthChange` (`character.go:898-907`) is never matched | `guards/death_test.go:246-250` |
| Literal `.Health` compares left in `characters` after the spec's changes | one: `Validate`'s `c.Health < -10` (`character.go:1205`). `Heal`'s and `Validate`'s `c.Health > c.HealthMax.Value` (`:994`, `:1201`) compare with a variable and pass | grep; guard run on the scratch copy |
| `IsDisabled` calls in `characters` that are an item's | 16 in `character.go` (`:1254`, `:1264`, `:1267`, `:1401`, and 12 in `Wear` `:1516`-`:1628`); the guard reads every one as "other" | grep; guard run (16 errors) |
| `IsDisabled` in `rooms` | none | grep |
| Bound-gear checks | 12 in `Wear` (`:1523`, `:1534`, `:1552`, `:1559`, `:1568`, `:1577`, `:1586`, `:1595`, `:1604`, `:1613`, `:1622`, `:1631`; `Wear` starts `:1465`) and 1 in `RemoveFromBody` (`:1649`) | grep |
| `FindAll` | literal `0b111111111` (`rooms.go:62`); deleting the `FindDowned` constant (`:55`) leaves it unchanged | read |
| `IsUp` reach in package tests | with `IsUp` panicking when unset and wired by `rules.Set`/`SetForTest`, the full suite panics in exactly two packages: `internal/characters` (`TestRecoveringHealthClearsTheKiller`) and `modules/gmcp` (`TestRoomContentsListOnlyWhatTheViewerPerceives`, via `gmcp.Room.go:377`) | scratch run |
| Boot order | `rules.Set` at `main.go:209`, before `loadAllDataFiles` at `:222` and `copyover.Restore` at `:275` | read |
| Other binaries | `cmd/reset-admin-pw` links `characters` but only loads and saves an admin account (no name, wear, heal or damage path); the other `cmd/*` and `scripts` do not link it | `go list -deps`; read `cmd/reset-admin-pw/main.go` |
| Registry lock | `actors.Get` takes `mu.RLock`; the only caller holding the registry lock while running caller code is `actors.Each` in `NewTurn_ActionPoints.go:11`, whose callback touches action points only | grep, read |
| `death.revive` | calls `ForgetKillerIfUp` on the live victim (`death/death.go:108-110`) | read |
| `BreakTest` draws a random number | `util.Rand(100)` (`items.go:105`) on every hit on a defender whose offhand has `BreakChance >= 1` | read |
| Crit flag | `isCrit = attackResult.Crit \|\| Crits(...)` (`combat.go:245`); `Crits` draws `util.Rand(100)` (`calculations.go:193`); `attackResult` is one per round (`:148`) | read |
| Unknown aggro type | after `BackStab` goes, type 3 is neither `DefaultAttack` (0) nor `Flee` (2): `combatTurn` (`NewRound_DoCombat.go:124`, `:130`) treats it as an attack on `Aggro.Target`; a saved flee has a zero target, so the turn tells "Your target can't be found." and ends the aggro (`:156-160`). Run on the scratch copy: one round, `AggroChanged rat`, aggro nil after | scratch test |
| Unknown aggro type on load | `SavedNPC.Mob` copies `Aggro` as saved, no check (`mobs/saved.go:54`); `Character.Aggro` is `yaml:"-"` (`character.go:73`), so only NPC sidecars carry it | read |
| Other `BackStab` text | `items.CanBackstab` (`itemspec.go:364-369`, no caller); `cmd/smoke/fight.go:122-133` strips a `*[BACKSTAB]*` prefix (tested in `fight_test.go:44-45`); admin tooltip `items-rank-weapons.html:208`; sample response `keywords-api.html:57` | grep `-i backstab` |
| Smoke crit handling | the only leading mark left is `*** ` (`combat.go:267-273`); `fight_test.go:43` "crit line" already covers it at line start; with the strip reduced to one `TrimPrefix`, both crit rows pass and a no-op strip fails both (run) | read, scratch run |
| Stale comments the deletions leave | `combat_golden_test.go:82` ("the broken-item spec"), `:603-604` ("the defender's shield can break"), `character.go:702-709` (`ForgetKillerIfUp`: "once its health is 1 or more"), `ruleset.go:125-129` (`Regen`), `registry_test.go:112` ("no BodyReplaced call") | grep |
| Combat golden helpers | `heal` (`combat_golden_test.go:537-551`), fixtures end at `:823`; last test `TestKilledByOutlivesTheKiller` ends at `:859` (end of file) | read |
| Golden line endings | goldens are `i/lf w/crlf` with `core.autocrlf=true`; `-update` writes LF and `git status` stays clean | `git ls-files --eol` |
| Package count | the scratch run after Task 6: 41 packages ok plus the 3 known Windows failures (unchanged; `geomud-verifying` needs no edit) | scratch run |
| `combatTurn` at `1c8e607f` (re-verified, plan review round 2) | `:102` the function; `:104-106` no aggro, no turn; `:124` combat commands for a `DefaultAttack` NPC; `:130` flee; `:143` the target lookup; `:156-160` not found ("Your target can't be found.", aggro ends); `:170-174` not up ("Your rage subsides."); `:184-194` weapon wait rounds; `:197-200` not perceived ("You can't seem to find your target."); `:202` `attack`; `:254` the NPC "stays put / heads that way" block; `:269-275` the end of the swing | read |
| Task 6a, ruling 12: NPC give-up today | `behaviour.loseLostQuarry` (`behaviour/tick.go:71-84`), called from `Tick`'s aggro branch (`:53-58`): a player target offline or in another room ends the NPC's aggro with `emote mumbles about losing their quarry.`; an NPC target never (`mobs.GetInstance(t) != nil` returns early). No test names it (grep `quarry`: `tick.go` only) | read, grep |
| Task 6a, ruling 12: config home | `GamePlay` (`configs/config.gameplay.go:3-27`), defaults in `GamePlay.Validate` (`:73-134`); `GetGamePlayConfig` validates before returning (`:203-211`); keys also listed in `_datafiles/config.yaml` (`:275-286`), `_datafiles/html/admin/config.html` (`:273-275`) and `config-wizard.html` (`:1089-1109`, fields `:1474`); no Go test enumerates keys | read, grep |
| Task 6a: who can reach the switch rule | `hooks` imports `death` (`NewRound_DoCombat.go:12`); `death` imports `rules`, `rooms`, `actors`, `messaging` (`death/death.go:11-26`) | read |
| Task 6a: round count | `util.GetRoundCount`/`IncrementRoundCount`/`SetRoundCount` (`util/util.go:104-124`); the game loop increments it once a round and sends it as `NewRound.RoundNumber` (`world.go:389-391`); the combat goldens' `DoCombat(events.NewRound{})` never advanced it | grep |
| Task 6a: `fighters()` order | map order on both sides (`accounts/users.go:140-149`, `mobs/mobs.go:218-225`); body IDs from `crypto/rand` (`ids/ids.go:28-36`) | read |
| Task 6a: `go` in a fight | refused when the body has an aggro or a body in its room targets it (`usercommands/go.go:25-28`, `:355-375`) | read |
| Task 6a: `CanAttack` | PvP rules whenever the target is an account (`dikubasic/ruleset.go:85-114`, `rooms.go:2139-2168`); shipped `limited`, minimum level 15 (`_datafiles/config.yaml:267-270`) | read |
| Task 6a: messages | an `Interface` message is written without token resolution (`messaging/interface.go:37-39`, `messaging.go:81-82`); an `ActorOnly` `WorldEvent` resolves `{target}` | read, scratch run |
| `docs/known-issues.md` | names none of this slice's items | grep `-i` for break, backstab, downed, floor, regenerat |
| Task 6a (plan review round 5): how `shoot` finds an exit | `room.FindExitByName(direction)` for the stored name and the room it leads to, then `room.GetExitInfo(exitName).Lock.IsLocked()` (`usercommands/shoot.go:62-66`). `FindExitByName` (`rooms.go:1688`) searches `Exits`, `ExitsTemp` and active mutators' `Exits`, and returns a temporary exit's `Title`; `GetExitInfo` (`rooms.go:805-820`) checks active mutators' exits before `Exits` and never `ExitsTemp` (a temporary exit has no lock). `Room.ActiveMutators` (`rooms.go:2073`) yields nothing when the zone has no config | read |
| Task 6a (plan review round 5): mutator specs in tests | `allMutators` (`mutators/mutators.go`), `MutatorSpec.Exits` (`:70`), `Mutator.GetSpec` (`:201-203`), `MutatorList.Add` refuses an unknown id (`:113-132`); no test helper existed (`ls internal/mutators`: `admin.go`, `mutators.go`, `plugin.go` plus GoMud's notes); `admin.go`'s `SaveMutatorSpec` writes a file | read |

## Settled from the code (the spec's five open items)

**1. How `IsUp` finds the actor, and where the wiring lives.** `internal/rules` gets
`isUp(c) = Get().Condition(actorOf(c)) == Up`. `actorOf(c)` is `actors.Get(c.Id)` when that actor's
`Character` is `c` itself; otherwise (a bare test character, an account not yet bound, a corpse's copy
of a body, which shares the live body's ID) a detached `&actors.Actor{Id: c.Id, Character: c}` with
no controller. The identity check matters: a corpse's copy must be judged on its own health, not the
respawned body's (test `TestIsUpAsksAboutTheBodyItIsGiven`). The wiring is `characters.IsUp = isUp` in
`rules.Set` and in `rules.SetForTest`: `main.go:209` runs `Set` before any data loads, and every
package test that already sets a ruleset gets it. `characters` keeps an unexported `isUp(c)` that
panics with `characters: IsUp is not set; rules.Set wires it at boot` when nothing set it. Package
tests: `hooks`, `usercommands` and the root package already call `rules.Set` in `TestMain`;
`dikubasic` and `rooms` tests never reach an `IsUp` path (full scratch run); `characters` cannot import
`rules`, so its `TestMain` sets `IsUp` to the reference rule (`Health > 0`); `modules/gmcp`'s
`TestMain` gains `rules.Set(dikubasic.New())`.

**2. An aggro `Type` that names nothing (old `Flee` = 3).** Loading copies it as saved. The round
treats it as an attack on `Aggro.Target`; a saved flee has a zero target, so the NPC's next turn
tells it "Your target can't be found." (it reads nothing) and ends its aggro (`AggroChanged`). It
never attacks and nothing crashes. Verified on the scratch copy. After Task 6a the same zero target
takes the lookup's step 1: the NPC turns to whoever in its room attacks it, or its aggro ends with
the same line. No code change; Task 6's commit
message records it.

**3. The commit 1 fixture.** `TestCombatGoldenNPCRegen`, appended at the end of
`combat_golden_test.go`: seed 1010, Alice (speed 30, strength 30, perception 20, 10 health, the
`testsword`) attacks a 40-health rat (race 10, no groups) for 2 rounds; both aggros are cleared
(`EndAggro`); then `heal(4, never)` runs the regen pass on rounds 3, 6, 9 and 12. On today's code the
rat reads 30/40 after round 2 and in all four heal passes (the exact file is in Task 1). It avoids
`TestCombatGoldenFleeSuccess`'s trouble: it uses no flee, no groups (so no hatred), no recovery room
and no runInputs; it sits last in the file, so in the default order it runs after
`FleeSuccess`. Measured on the scratch copy: `go test ./internal/hooks/ -count=2` fails
`FleeSuccess` alone with or without the new test (`-skip TestCombatGoldenNPCRegen`); the new test
passes `-count=3` alone and never failed in six `-shuffle` seeds (1 to 6).

**4. Where the widened death guard fires.** On the post-change code: nowhere in `rooms`. In
`characters`, two things: (a) `Validate`'s `c.Health < -10` (`:1205`) until Task 3 removes it, and
(b) the 16 item `IsDisabled` calls in `character.go`, which the guard counts as non-body calls and
which need `character.go` on the `itemIsDisabled` allowlist (Q2). `ApplyHealthChange`'s
`newHealth < 0`, `newHealth < -5`, `newHealth <= -10` compare a local, which the guard does not match
(it matches `.Health` selectors only). After Task 3 the guard passes; planting `c.Health < 1` in
`characters/isup.go` and `mob.Character.Health < 1` in `rooms.go` each fails it (run).

**5. Tests naming deleted symbols.**

| Symbol | Test | What happens |
|---|---|---|
| `Character.IsDisabled` | `characters/character_test.go:1195-1236` `TestCharacter_IsDisabled` | deleted (Task 2) |
| `Character.IsDisabled` | `guards/death_test.go:35` planted `a.Character.IsDisabled()`, `:77-81` the body branch, `:111-126` its control, `:177-198` `isDisabledCalls` | the body ban goes; the matcher counts every `IsDisabled` call (Task 3) |
| `items.Item.IsDisabled` | `items/items_test.go:161` `TestItem_IsDisabled` | stays (the item method stays) |
| `FindDowned` | none | |
| `BodyReplaced` | `actors/registry_test.go:185-195` `TestBodyReplacedRekeysThePlayer` | deleted (Task 2) |
| `BodyReplaced` | `actors/registry_test.go:112` comment "no BodyReplaced call" | comment reworded (Task 2) |
| `BreakTest`, `BreakChance` | `hooks/combat_golden_test.go:270-276` `shield(..., breakChance)`, `:611` `shield(93102, "testshield", 90)` | parameter and field dropped (Task 5) |
| `BackStab` | `hooks/combat_golden_test.go:295` the aggro name map | entry dropped (Task 6) |
| `BackStab` (text) | `cmd/smoke/fight_test.go:44-45`, the two `*[BACKSTAB]*` rows | deleted; one "crit line after prompt" row takes the prompt case (Task 6, Q5) |

## Questions settled by the controller (2026-11-12, recorded in the spec)

Each was found while planning, by reading or running the code.

- **Q1. The guard widens in commit 3, with the floors** (Task 3). `Validate`'s `if c.Health < -10`
  (`character.go:1205`) is a literal `.Health` compare, so a guard widened in commit 2 fails (run:
  `internal/characters/character.go: compares .Health with a number`). The `ApplyHealthChange` floors
  compare a local and do not trip it.
- **Q2. `character.go` joins the guard's item `IsDisabled` allowlist** (Task 3 Step 6): its 16 calls
  are item slot checks (race-disabled slots in `Validate`, `BestUpgrades` and `Wear`).
- **Q3. The `pvp` golden's shifted draws are decided.** `BreakTest` drew `util.Rand(100)` on every hit
  on Bob, so removing it shifts every draw from round 2 on; the full expected diff is in Task 5 Step 7.
- **Q4. The crit cascade stays; its comment is corrected** (Task 6 Step 3). Once a hit crits,
  `attackResult.Crit` is true and every later hit that round short-circuits
  `attackResult.Crit || Crits(...)` (a crit, no roll); the comment at `combat.go:243-244` said the
  opposite. Task 7 adds a 4.3 note naming the cascade to the phase design.
- **Q5. The backstab leftovers go, all in Task 6:** `items.CanBackstab` (`itemspec.go:364`, no caller),
  the smoke client's `*[BACKSTAB]*` strip (`cmd/smoke/fight.go:122-133`, the function and its comment)
  and its two test rows (`fight_test.go:44-45`), and the word "backstab" in the admin weapon-rank
  tooltip (`items-rank-weapons.html:208`). `_datafiles/html/admin/keywords-api.html` (a sample API
  response naming the `backstab` help topic) stays for slice 7's docs sweep.
- **Q6. Spec line numbers corrected** in the spec's facts table. No design effect. (The plan's own
  first draft of Q6 said the nil check was at `rooms.go:264`; that was read from the scratch copy after
  `FindDowned`'s constant had gone. At `1c8e607f` it is `:265`, as the spec says.)
- **`Wear`'s 12 bound-gear checks have no unit test a mutation can kill** (`Wear` needs race data);
  accepted, with the death guard covering them (Task 2 Step 1).

## Questions settled by the controller for Task 6a (2026-11-12, recorded in the spec as "plan Q7" to "plan Q15")

- **Q7:** `dikubasic.CanAttack` applies the PvP rules only when both sides are players; the
  recovery room (and its copies) protects a player from every attacker (Task 6a Step 5a).
- **Q8:** `inFight` counts the body's own aggro only when its target is in the body's room; a body
  in its room being attacked is still in a fight (Step 5b).
- **Q9:** turn order sorted (players by UserId, NPCs by MobId then body ID) (Step 4).
- **Q10:** `death_npc`'s Alice has 30 health (Step 7).
- **Q11:** a switch inside a turn swings at the new opponent that turn, unless the weapon's delay
  makes it wait first: the switch re-arms through `SetAggro` like `attack <name>` (plan review round
  3; Step 5, `TestSwitchKeepsTheWeaponDelay`).
- **Q12:** lookup order: gone, switch; not up, switch; other room and can shoot, shoot (kept even if
  someone in the room attacks); other room and attacked here, switch; else wait (Step 5).
- **Q13:** the waiting line is "Your target is out of reach." (once) (Step 5).
- **Q14:** the switch rule lives in `internal/death/switch.go` (Step 3).
- **Q15:** a failed flee keeps the `Since` the fleer's aggro had (Step 5c).
- **Q16 (spec ruling 10):** PvP settings apply only between two players; NPCs may attack, shoot and
  retaliate against players anywhere except the recovery room. The slice 2 test flips to
  `TestNPCShootingAPlayerAllowedWhenPvPDisabled`, and the commit says it reverses slice 2's
  NPC-shooter line by ruling 10 (Step 5a, Step 13).
- **Q17 (spec Decisions "plan Q17"):** `inFight`'s new meaning applies to `go`, `quit` and `suicide`; a
  pending flee with nobody attacking is not a fight. A `quit` test covers a remote-only target (Step 5b).

## Question 18 (answered)

**Q18. The smoke fight step can time out against a level 4 or 5 rat, before and after this slice.**
Found while settling the smoke risk (Task 6a Step 12). Bot 0's fight length depends only on its own
hits, so retaliation does not change it. Simulated, about 1.7% (level 4) and 2.4% (level 5) of fights
run past 60 rounds, the 4-minute step timeout at 4-second rounds; rats spawn at a random level 1 to 5,
so roughly 0.8% of smoke runs on that alone (unverified as a run rate: it assumes each level equally
likely and one fight per run). The step's comment measured "up to about 110 s, 27 rounds". Options:
the smoke world copy also pins the rat's spawn level to 1 (`level: 1` on its spawn entries in
`rooms/startland/2.yaml`), which also shortens the step; or raise `fightStepTimeout`; or leave it.
*Answered:* no code for the flake itself; ruling 11's pinned rats (Task 6a Step 12) remove it in
launch mode; `-addr` mode keeps it; Task 7 records it in the status Watch list.

## Held with the owner, now answered (plan review round 2; ruling 12)

- How long a waiting aggro lasts, and `behaviour.loseLostQuarry` (`internal/behaviour/tick.go:55-84`),
  which `ActorTick` runs every round and which ends an NPC's aggro on a player who is offline or in
  another room. *Answered by ruling 12:* a wait ends after `GamePlay.OutOfReachRounds` (default 5)
  consecutive turns out of reach, for every body; `loseLostQuarry` and its call are deleted (Task 6a
  Steps 5 and 5d).
- NPC against NPC: an NPC whose NPC target leaves the room keeps a waiting aggro with no end
  (`loseLostQuarry` skips NPC targets). *Answered by ruling 12:* the same limit ends it
  (`TestNPCWaitsEndForOfflinePlayersAndNPCs`).

The record of each question as asked follows, for the evidence.

## Questions as asked (Task 6a, round 2; answered above)

**Q16. Q7's answer reverses a slice 2 rule for NPC `shoot` and `attack`.** Slice 2 made shooting a
player obey the PvP settings "for player and NPC shooters alike" (`docs/completed/2026-11-10-phase-5-2-seam-fight-spec.md:254-257`,
patch notes 2026-11-10), tested by `TestNPCShootingAPlayerRefusedWhenPvPDisabled`
(`usercommands/shoot_test.go:113-123`). With PvP applying only between players, that test fails (run:
the NPC's shot sets an aggro with PvP disabled), and an NPC's `attack` or `shoot` on a player (NPC
combat commands) now works under `limited` below level 15 and outside PvP rooms. Task 6a Step 5a
replaces the test with `TestNPCShootingAPlayerAllowedWhenPvPDisabled` (the NPC shoots; its aggro
targets the player). Confirm that the reversal is intended; if not, retaliation needs its own check
instead of `CanAttack`.

**Q17. `inFight` also governs `quit` and `suicide`.** `quit` (`usercommands/quit.go:18`) and `suicide`
(`suicide.go:17`) refuse through the same `inFight`. After Q8's change a body whose aggro waits on a
target in another room, or that shoots into another room, can also quit and suicide (it can walk away,
so quitting is no worse). And a body whose only aggro is a pending `Flee` (target zero) with nobody in
the room targeting it no longer counts as in a fight. Confirm, or limit the change to `go`.

## Questions as asked (Task 6a, round 1; answered above)

**Q7. `CanAttack(defender, attacker)` stops NPCs fighting back against players under the shipped
config.** `dikubasic.CanAttack` applies the PvP rules whenever the target (here the retaliation's
target, the player) is an account, whoever attacks (`ruleset.go:85-114`); `Room.CanPvp` then refuses
below `MinimumLevel` and outside PvP rooms (`rooms.go:2139-2168` at `1c8e607f`). Shipped: `Enabled: limited`,
`MinimumLevel: 15` (`_datafiles/config.yaml:267-270`). So an NPC hit by a level 1 player never
retaliates. Smoke on the scratch copy shows it: across 1, 5 and 12 bots the server logged only the
bots' attacks on the rat (`calculateCombat ... Target=rat`), never the rat's, over fights of up to 3
rounds. The goldens do not show it: test binaries load no config, so `CanPvp` passes. Player against
player is symmetric (both must meet the same rule), so only NPC retaliation is affected. Options:
(a) `dikubasic.CanAttack` applies the PvP rules only when both sides are accounts (an NPC hitting a
player is not PvP); this also changes `attack` and `shoot` by an NPC (NPC combat commands);
(b) retaliation does not ask `CanAttack` when the defender is an NPC; (c) keep as written.
*Recommend (a):* it is what "PvP" means, and it keeps "may it attack back" the ruleset's.

**Q8. A waiting aggro blocks walking.** `go` refuses any body with an aggro, or with a body in its
room targeting it (`usercommands/go.go:25-28`, `inFight` `:355-375`). (a) The design's "an NPC attacked
from another room still walks toward its attacker ... and swings once it arrives" does not happen: the
NPC queues `go <exit>` before retaliation gives it an aggro, the command runs after and is refused.
Shown by the `shot_from_another_room` fixture (`Ran rat: "go south"`, the rat stays in room B).
(b) A player shot from another room without a bow gets a waiting aggro with no time limit, and can then
leave only by `flee` (a random exit) until the shooter dies or leaves the game. Options: `inFight`
ignores the body's own aggro while its target is in another room; or retaliation from another room
gives no aggro unless the defender can shoot back; or keep as written (the NPC never comes, the player
must flee). *Recommend:* `inFight` ignores an own aggro whose target is not in the room, which makes
both (a) and (b) work as the design describes and keeps "nobody walks out of a fight" for fights in
the room.

**Q9. Turn order within a kind is map order.** `fighters()` lists accounts from
`accounts.GetOnlineUserIds()` and NPCs from `mobs.GetAllMobIds()`, both map iterations
(`users.go:140-149`, `mobs.go:218-225`); body IDs are crypto-random (`ids.go:28-36`). Once defenders
fight back, two fighting accounts or two fighting NPCs act in a random order each round: after
`-update`, `pvp`, `death_pvp`, `mvm` and `hidden_attacker` differed on 8 of 8 reruns, and the
switch's "ties go to turn order" is random in play. Task 6a Step 4 sorts accounts by `UserId` and
NPCs by `MobId`, then body ID text; with it every golden was stable over 6 runs. Player-visible: among
players, the lower UserId (the older account) acts first each round. Accept, or name another key.

**Q10. `TestCombatGoldenDeathNPC` no longer reaches its NPC death.** With retaliation the elite rat
(club) kills Alice in round 5 and survives at 1/12 (scratch run); the fixture exists to record an NPC
death (drops, XP, ledger). Recommended: Alice's health 10 becomes 30 in that test; the run then has the
rat die in round 6 with Alice at 20/30 (Task 6a Step 7).

**Q11. Does a switch during a turn also swing?** The spec says the aggro moves with a line; it does
not say whether the turn then swings. Task 6a swings at the new opponent in the same turn when the
lookup switches (steps 1-3), so a body loses no round (`target_flees_switch`, round 2: the line, then
the swing). A switch from a death happens after all turns (`handleAffected`), so it swings next
round. Confirm, or make the lookup switch take the turn.

**Q12. Step 3 before step 4 can replace a chosen shot target.** A body shooting a target it chose
(`shoot Z`, through an exit) switches to anyone in its own room who attacks it, since step 3 comes
before step 4. Ruling 7 says a chosen target is never replaced while it is reachable, and a shot target
is reachable. The aggro does not record whether it was chosen or taken in retaliation. Options: keep
the step order; or step 4 before step 3 (a body that can shoot its target keeps shooting it).

**Q13. The waiting line for a target that was never here.** A body shot from another room, with no
bow, reads "Your target is no longer here." on its first turn (`shot_from_another_room`, Carol), though
the target was never in its room. Reword (for example "Your target is out of reach."), or suppress it
when the aggro began with the target elsewhere.

**Q14. Where the switch rule lives.** `hooks` imports `death`, so the rule that both the round and
`Die` use is in `internal/death/switch.go` (`death.NextOpponent`, `death.Switch`). The spec places all
of this in the round and `internal/death`; confirm, or a new engine package (for example
`internal/fight`) both import.

**Q15. A failed flee restarts `Since`.** `flee` replaces the aggro with a `Flee` aggro whose target is
zero (`usercommands/flee.go:22`), and a blocked flee re-arms the old target (`resumeAfterFlee`), which
is a target change, so `Since` restarts. A body that tries to flee drops to the back of its
opponent's queue. Accept, or carry `Since` through the flee aggro.

## Agent rules (brief every subagent with these)

- Edit files with the Edit and Write tools. Never write a repo file from the shell (no `>`, `>>`,
  `sed -i`, Python open-write); a hook blocks it. `git rm` is fine.
- Read files in chunks of at most 400 lines; state "read N of N files" with the list in the report.
- `git add <named paths>` only, never `-A` or `.`; commit messages through a quoted heredoc ending
  with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never `--no-verify`. Never push.
  Never touch `upstream`.
- Run every test command in the foreground with an explicit timeout of 300000 ms (600000 for the
  full suite).
- `go test` always with `-count=1`. **`-update` runs only in Tasks 1, 3, 4, 5 and 6a, only with the
  exact `-run` pattern the task gives.** Anywhere else a golden diff is a failure to report, not to
  re-record.
- Throwaway output goes in the session scratchpad the controller names, never `C:/tmp` or `$TMP`.
- If a step's code does not compile, a prescribed failure does not happen, or a golden diff differs
  from the one stated, stop and report it; do not paper over it.
- Run `gofmt -w` on every Go file you created or changed before the per-task check.
- A mutation probe is: Edit, run the named test, see it fail, paste the failure line, undo with Edit,
  rerun, see `ok`. `git diff --stat` after the probes shows only the task's files.
- Line numbers below are at `1c8e607f`. Earlier tasks shift them; find each edit by its `old_string`.

## Per-task check (end of every task)

```bash
go build ./...
go vet ./... 2>&1 | head
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'   # only the 3 known Windows failures (geomud-verifying)
gofmt -l internal modules cmd *.go                          # prints nothing
git status --short                                           # only this task's files before the commit; nothing after (Task 0 leaves the tree clean)
```

The 3 known failures: `cmd/reset-admin-pw` `TestPromptForPasswordMismatch`, `internal/mobs`
`TestLoadPluginMobs_MergesValidMob`, `scripts` `TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride`.
Report: build result, the failing-test lines, the `git status` output.

---

### Task 0: Baseline (controller)

- [ ] On `strip/phase-5-4-1-leftovers`, run the per-task check except its last line; save the output
  in the scratchpad as the baseline (expected: 41 packages ok and the 3 known failures). Before this
  commit `git status --short` shows exactly ` M 2026-11-12-phase-5-4-1-leftovers-spec.md` (the
  controller's settled-while-planning edits since `1c8e607f`) and
  `?? 2026-11-12-phase-5-4-1-leftovers-plan.md`.
- [ ] Once Calabe approves the plan, commit the spec and the plan together:

```bash
git add 2026-11-12-phase-5-4-1-leftovers-spec.md 2026-11-12-phase-5-4-1-leftovers-plan.md
git commit -F - <<'EOF'
Phase 5.4.1: plan, and the spec's settled-while-planning answers

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
git status --short   # prints nothing: every task from here starts on a clean tree
```

### Task 1: The regen golden on today's code (Sonnet)

Test files and testdata only (decision 12). Every existing combat and reads golden stays
byte-identical.

**Files:** modify `internal/hooks/combat_golden_test.go`; create
`internal/hooks/testdata/combat-golden/npc_regen.txt` (by `-update`).

- [ ] **Step 1: The fixture.** Append at the end of `internal/hooks/combat_golden_test.go` (after
  `TestKilledByOutlivesTheKiller`, which ends at `:859`):

```go

// A player wounds an NPC, then the fight ends (both aggros are cleared, as
// a "break" would leave them) and the regen pass runs four times (Phase 5
// slice 4.1, recorded on today's code first: an NPC does not regenerate).
// The rat has no groups, so the fixture leaves no hatred behind.
func TestCombatGoldenNPCRegen(t *testing.T) {
	w := newCombatWorld(t, 1010)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 40)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(2)
	alice.Character.EndAggro()
	rat.Character.EndAggro()
	wounded := rat.Character.Health
	w.heal(4, func() bool { return false })
	w.check("npc_regen")
	require.Less(t, wounded, rat.Character.HealthMax.Value, "the fixture must wound the rat")
	require.Greater(t, wounded, 0, "the rat must survive the fight")
}
```

- [ ] **Step 2: Record it.**

```bash
go test ./internal/hooks/ -run 'TestCombatGoldenNPCRegen$' -count=1 -update
```

  Expected: `ok`. The new file must read exactly:

```
seed 1010
== start
-- state
  Alice: room=93001 hp=10/10 aggro=[rat DefaultAttack wait=0] playerdamage={} weapon=93101
  rat: room=93001 hp=40/40 aggro=[none] playerdamage={}
== round 1
-- Alice reads
  Your <ansi fg="item">testsword</ansi> cuts into <ansi fg="mobname">rat</ansi>, dealing <ansi fg="damage">3 damage</ansi>!
-- events
  AggroChanged Alice room=93001
-- state
  Alice: room=93001 hp=10/10 aggro=[rat DefaultAttack wait=0] playerdamage={} weapon=93101
  rat: room=93001 hp=37/40 aggro=[none] playerdamage={Alice:3}
== round 2
-- Alice reads
  In a sweeping motion, you slash <ansi fg="mobname">rat</ansi> with your <ansi fg="item">testsword</ansi>, doing an impressive <ansi fg="damage">7 damage</ansi>!
-- events
  AggroChanged Alice room=93001
-- state
  Alice: room=93001 hp=10/10 aggro=[rat DefaultAttack wait=0] playerdamage={} weapon=93101
  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
== heal 1
-- Alice reads
-- events
-- state
  Alice: room=93001 hp=10/10 aggro=[none] playerdamage={} weapon=93101
  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
== heal 2
-- Alice reads
-- events
-- state
  Alice: room=93001 hp=10/10 aggro=[none] playerdamage={} weapon=93101
  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
== heal 3
-- Alice reads
-- events
-- state
  Alice: room=93001 hp=10/10 aggro=[none] playerdamage={} weapon=93101
  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
== heal 4
-- Alice reads
-- events
-- state
  Alice: room=93001 hp=10/10 aggro=[none] playerdamage={} weapon=93101
  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
```

  (The rat never takes an aggro of its own: the record shows `aggro=[none]` throughout.)

- [ ] **Step 3: Nothing else moved.**

```bash
go test ./internal/hooks/ -count=1
git status --short -- internal/hooks/testdata/
```

  Expected: `ok`; status shows only `?? internal/hooks/testdata/combat-golden/npc_regen.txt`.

- [ ] **Step 4: Order check.** `go test ./internal/hooks/ -run 'TestCombatGoldenNPCRegen$' -count=3`
  prints `ok` (the fixture repeats cleanly). Report it.

- [ ] **Step 5: Per-task check, then commit.**

```bash
git add internal/hooks/combat_golden_test.go internal/hooks/testdata/combat-golden/npc_regen.txt
git commit -F - <<'EOF'
Phase 5.4.1: regen golden on today's code

A player wounds a rat, the fight ends and the regen pass runs four
times; today the rat stays at 30/40 (an NPC never regenerates).
Recorded before the change (decision 12). Every other golden unchanged.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 2: `characters.IsUp`; the checks through it; three deletions (Opus; the controller reads the diff)

Goldens unchanged. The death guard is widened in Task 3, with the floors, not here (Q1).

**Files:** create `internal/characters/isup.go`, `internal/rules/isup_test.go`; modify
`internal/characters/character.go`, `internal/characters/character_test.go`, `internal/rules/rules.go`,
`internal/rooms/rooms.go`, `internal/actors/registry.go`, `internal/actors/registry_test.go`,
`modules/gmcp/visibility_test.go`.

- [ ] **Step 1: The stub-ruleset test, failing first.** Create `internal/rules/isup_test.go`:

```go
package rules_test

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/rules/dikubasic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false) // a body's names ask buffs, which log
	os.Exit(m.Run())
}

// downedAtFive is the reference rules with a downed state: from 5 health
// down to 1 a body is downed (not up, not dead).
type downedAtFive struct{ rules.Ruleset }

func (downedAtFive) Condition(a *actors.Actor) rules.Condition {
	switch {
	case a.Character.Health <= 0:
		return rules.Dead
	case a.Character.Health <= 5:
		return rules.Downed
	}
	return rules.Up
}

// woundedBody is a body at 5 of 10 health with a killer on record, wearing
// a ring that can never be removed.
func woundedBody() (*characters.Character, items.Item) {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = "Ann"
	c.HealthMax.Value, c.Health = 10, 5
	c.KilledBy = ids.New(ids.NoCell)
	ring := items.Item{ItemId: 9, Id: ids.New(ids.NoCell), CanNeverBeRemoved: true,
		Spec: &items.ItemSpec{ItemId: 9, Name: "ring", Type: items.Ring}}
	c.Equipment.Ring = ring
	return c, ring
}

// Phase 5 slice 4.1: characters asks the ruleset whether a body is up
// (characters.IsUp), for the downed display, the killer record and bound
// gear, whether or not a live actor holds the body.
func TestCharactersAskTheRulesetWhetherABodyIsUp(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rules rules.Ruleset
		up    bool
	}{
		{"downed at 5", downedAtFive{dikubasic.New()}, false},
		{"reference rules at 5", dikubasic.New(), true},
	} {
		for _, live := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: ", bare body", true: ", live actor"}[live], func(t *testing.T) {
				t.Cleanup(rules.SetForTest(tc.rules))
				c, ring := woundedBody()
				if live {
					acc := &accounts.Account{UserId: 9701, Character: c}
					require.NotNil(t, actors.Spawn(c, actors.NewConnectionController(acc)))
					t.Cleanup(func() { actors.Remove(c.Id) })
				}

				name := c.GetPlayerName(ids.ID{}, characters.RenderHealth)
				c.ForgetKillerIfUp()
				removed := c.RemoveFromBody(ring)

				if tc.up {
					assert.Empty(t, name.Suffix)
					assert.Contains(t, name.Adjectives, "50%", "health shows while up")
					assert.NotContains(t, name.Adjectives, "downed")
					assert.True(t, c.KilledBy.IsZero(), "up: the killer record goes")
					assert.False(t, removed, "up: bound gear stays on")
				} else {
					assert.Equal(t, "downed", name.Suffix)
					assert.Contains(t, name.Adjectives, "downed")
					assert.NotContains(t, name.Adjectives, "50%", "no health while downed")
					assert.False(t, c.KilledBy.IsZero(), "downed: the killer record stays")
					assert.True(t, removed, "downed: bound gear comes off")
				}
			})
		}
	}
}

// A copy of a live body (a corpse holds one) is asked about itself, not
// about the live body with the same ID.
func TestIsUpAsksAboutTheBodyItIsGiven(t *testing.T) {
	t.Cleanup(rules.SetForTest(downedAtFive{dikubasic.New()}))
	c, _ := woundedBody()
	c.Health = 10
	acc := &accounts.Account{UserId: 9702, Character: c}
	require.NotNil(t, actors.Spawn(c, actors.NewConnectionController(acc)))
	t.Cleanup(func() { actors.Remove(c.Id) })

	copied := *c
	copied.Health = 3
	assert.True(t, characters.IsUp(c), "the live body is up")
	assert.False(t, characters.IsUp(&copied), "its copy at 3 health is downed")
}

// rules.Set and rules.SetForTest wire characters.IsUp.
func TestSetWiresCharactersIsUp(t *testing.T) {
	saved := characters.IsUp
	t.Cleanup(func() { characters.IsUp = saved })
	t.Cleanup(rules.SetForTest(nil)) // Set below changes the active ruleset; this puts the previous one back

	characters.IsUp = nil
	rules.Set(dikubasic.New())
	require.NotNil(t, characters.IsUp, "Set")

	characters.IsUp = nil
	rules.SetForTest(dikubasic.New())()
	require.NotNil(t, characters.IsUp, "SetForTest")
}
```

  Run `go test ./internal/rules/ -count=1`. Expected: build failure, `undefined: characters.IsUp`.

  (`Wear`'s 12 bound-gear checks are not exercised here: `Wear` needs race data
  (`HandsRequired` reads `races.GetRace`, nil without data files; tried, it panics). They change by
  the same expression as `RemoveFromBody`'s, and the death guard (widened in Task 3) fails on any
  health compare there. The controller accepted this gap on that basis, 2026-11-12.)

- [ ] **Step 2: `characters.IsUp`.** Create `internal/characters/isup.go`:

```go
package characters

// IsUp reports whether the ruleset counts c as up (rules.Condition Up).
// internal/rules sets it (rules.Set, rules.SetForTest); characters cannot
// import rules, which imports actors, which imports characters. It is the
// one route by which characters asks the ruleset about a body (Phase 5
// slice 4.1); it compares no health with a number itself.
var IsUp func(c *Character) bool

// isUp is IsUp, failing loudly when nothing set it, as rules.Get does: a
// boot or a test that chose no ruleset must not guess a body's condition.
func isUp(c *Character) bool {
	if IsUp == nil {
		panic("characters: IsUp is not set; rules.Set wires it at boot")
	}
	return IsUp(c)
}
```

- [ ] **Step 3: The wiring.** In `internal/rules/rules.go`, the import block (`:7-11`) gains
  `"github.com/GoMudEngine/GoMud/internal/characters"` after the `actors` line. `Set` (`:127-128`):

```go
// Set makes r the active ruleset and wires characters.IsUp to it. main.go
// calls it once at boot.
func Set(r Ruleset) {
	active = r
	characters.IsUp = isUp
}
```

  `SetForTest` (`:139-145`) becomes, with the two new functions after it:

```go
// SetForTest makes r the active ruleset and returns the func that restores
// the previous one, for t.Cleanup.
func SetForTest(r Ruleset) (restore func()) {
	prev := active
	active = r
	characters.IsUp = isUp
	return func() { active = prev }
}

// isUp is characters.IsUp: whether the active ruleset calls c's actor Up.
func isUp(c *characters.Character) bool {
	return Get().Condition(actorOf(c)) == Up
}

// actorOf is the live actor whose body is c. A body no live actor holds (a
// corpse's copy of a body, a test's bare character, an account not yet
// bound) gets a detached actor: that body, its ID, no controller.
func actorOf(c *characters.Character) *actors.Actor {
	if a := actors.Get(c.Id); a != nil && a.Character == c {
		return a
	}
	return &actors.Actor{Id: c.Id, Character: c}
}
```

- [ ] **Step 4: The five checks in `character.go`.** Each by Edit:
  - `GetAdjectives` (`:422`): `if c.Health < 1 {` becomes `if !isUp(c) {`.
  - `getFormattedName` (`:462-463`): the comment and condition become
    `// Health shows only while up: a body that is not up shows downed instead.` /
    `if includeHealth && isUp(c) {`.
  - `getFormattedName` (`:470`): `if c.Health < 1 {` becomes `if !isUp(c) {`.
  - `ForgetKillerIfUp` (`:702-712`): the comment's first two lines
    `// ForgetKillerIfUp drops the record of who downed this character once its` /
    `// health is 1 or more, so a later death from another cause (poison, a`
    become
    `// ForgetKillerIfUp drops the record of who downed this character once the` /
    `// ruleset calls it up again (IsUp), so a later death from another cause (poison, a`;
    and `if c.Health < 1 {` (`:711`) becomes `if !isUp(c) {`.
  - The 13 bound-gear checks (`:1523`, `:1534`, `:1552`, `:1559`, `:1568`, `:1577`, `:1586`,
    `:1595`, `:1604`, `:1613`, `:1622`, `:1631`, `:1649`): each `IsRemoveLocked() && c.Health > 0`
    becomes `IsRemoveLocked() && isUp(c)`. Use Edit with `replace_all: true` on the exact text
    `IsRemoveLocked() && c.Health > 0`, then confirm:

```bash
grep -c "IsRemoveLocked() && isUp(c)" internal/characters/character.go   # 13
grep -n "c.Health > 0\|c.Health < 1" internal/characters/character.go     # nothing (exit 1); run it on its own
```

- [ ] **Step 5: Delete `Character.IsDisabled`** (`character.go:819-821` and the blank line after):

```go
func (c *Character) IsDisabled() bool {
	return c.Health <= 0
}
```

  and its test, `character_test.go:1195-1236` (`func TestCharacter_IsDisabled` through its closing
  `}`; the next line is `func TestCharacter_EndAggro`).

- [ ] **Step 6: The `characters` test binary sets `IsUp`.** In `character_test.go`, `TestMain`
  (`:55-70`): after `skills.SetTestData(seedSkills)` and its blank line, before `os.Exit(m.Run())`,
  add:

```go
	// The reference rules' Condition (rules.Set wires IsUp in the game;
	// this package cannot import rules): up above 0 health.
	IsUp = func(c *Character) bool { return c.Health > 0 }

```

  And append the loud-failure test at the end of the file:

```go

// An unset IsUp fails loudly (Phase 5 slice 4.1), as rules.Get does.
func TestUnsetIsUpFailsLoudly(t *testing.T) {
	saved := IsUp
	t.Cleanup(func() { IsUp = saved })
	IsUp = nil
	assert.PanicsWithValue(t, "characters: IsUp is not set; rules.Set wires it at boot", func() { isUp(New()) })
}
```

- [ ] **Step 7: `modules/gmcp` sets the ruleset.** In `modules/gmcp/visibility_test.go`, the imports
  (`:3-18`) gain `"github.com/GoMudEngine/GoMud/internal/rules"` and
  `"github.com/GoMudEngine/GoMud/internal/rules/dikubasic"` after the `rooms` line; `TestMain`
  (`:20-24`) becomes:

```go
func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init()
	rules.Set(dikubasic.New()) // the room panel's names ask characters.IsUp
	os.Exit(m.Run())
}
```

  (Without it `TestRoomContentsListOnlyWhatTheViewerPerceives` panics with the `IsUp` message, via
  `gmcp.Room.go:377`.)

- [ ] **Step 8: Delete `rooms.FindDowned`.** In `internal/rooms/rooms.go` delete the constant line
  (`:55`) `	FindDowned         FindFlag = 0b00001000000 // hp < 1` and both reads with the blank line
  after each:

```go
		if typeFlag&FindDowned == FindDowned && mob.Character.Health < 1 {
			mobMatches = append(mobMatches, mobId)
			continue
		}

```

  (`:1144-1148`) and

```go
		if typeFlag&FindDowned == FindDowned && user.Character.Health < 1 {
			playerMatches = append(playerMatches, userId)
			continue
		}

```

  (`:1219-1223`). `FindAll` (`:62`) keeps its literal.

- [ ] **Step 9: Delete `actors.BodyReplaced`.** `registry.go:180-193` (the comment from
  `// BodyReplaced re-keys a player` through the function's closing `}`) and the blank line after.
  `registry_test.go:185-195` (`func TestBodyReplacedRekeysThePlayer` through its closing `}`) and the
  blank line after. `registry_test.go:112`: the comment `// no BodyReplaced call` becomes
  `// a new body ID, seen at the next login`.

- [ ] **Step 10: Run the new tests.**

```bash
go test ./internal/rules/ ./internal/characters/ ./internal/rooms/ ./internal/actors/ ./modules/gmcp/ -count=1
grep -rn --exclude-dir=.claude "FindDowned\|BodyReplaced\|TestCharacter_IsDisabled" --include=*.go .   # nothing; run on its own
grep -rn "func (c \*Character) IsDisabled" internal/                               # nothing; run on its own
grep -rn "func (i Item) IsDisabled" internal/items/items.go                        # one line: the pattern can match
```

  Expected: all `ok`.

- [ ] **Step 11: Mutation probes** (each must fail the named test, then be undone):

| Mutation | Test | Failure seen on the scratch copy |
|---|---|---|
| `GetAdjectives`: `if !isUp(c) {` to `if c.Health < 1 {` | `go test ./internal/rules/ -run TestCharactersAsk` | `[]string{} does not contain "downed"` |
| health %: `if includeHealth && isUp(c) {` to `if includeHealth && c.Health > 0 {` | same | `should not contain "50%"`, "no health while downed" |
| suffix: the `if !isUp(c) {` before `f.Suffix = \`downed\`` to `if c.Health < 1 {` | same | `Not equal` (suffix) |
| `ForgetKillerIfUp`: `if !isUp(c) {` to `if c.Health < 1 {` | same | `Should be false`, "downed: the killer record stays" |
| `RemoveFromBody`: `i.IsRemoveLocked() && isUp(c)` to `i.IsRemoveLocked() && c.Health > 0` | same | `Should be true`, "downed: bound gear comes off" |
| `actorOf`: `a != nil && a.Character == c {` to `a != nil {` | `-run TestIsUpAsksAbout` | `Should be false`, "its copy at 3 health is downed" |
| `Set`: delete its `characters.IsUp = isUp` | `-run TestSetWires` | `Expected value not to be nil`, "Set" |
| `SetForTest`: delete its `characters.IsUp = isUp` | `-run TestSetWires` | `Expected value not to be nil`, "SetForTest" |
| `isup.go`: the `panic(...)` line to `return false` | `go test ./internal/characters/ -run TestUnsetIsUp` | `func (assert.PanicTestFunc)(...) should panic` (the test calls `isUp` directly, so nothing else in the body runs; calling `GetAdjectives` instead would hit the package's nil logger before Task 3 and fail as "should panic with value" for the wrong reason: plan review round 3) |

- [ ] **Step 12: Goldens unchanged.** `go test ./internal/hooks/ -count=1` prints `ok` (no
  `-update`). Then the per-task check; `git status` shows only this task's files.

- [ ] **Step 13: Commit.**

```bash
git add internal/characters/isup.go internal/characters/character.go internal/characters/character_test.go \
  internal/rules/rules.go internal/rules/isup_test.go internal/rooms/rooms.go \
  internal/actors/registry.go internal/actors/registry_test.go modules/gmcp/visibility_test.go
git commit -F - <<'EOF'
Phase 5.4.1: characters asks the ruleset whether a body is up

characters.IsUp, set by rules.Set and rules.SetForTest, is the active
ruleset's Condition == Up for the body's live actor (a detached actor
for a body no live actor holds, such as a corpse's copy). The downed
adjective and suffix, the health percentage, the killer record and the
13 bound-gear checks go through it; an unset IsUp panics.
Character.IsDisabled, rooms.FindDowned and actors.BodyReplaced are
deleted with their tests. Goldens unchanged.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 3: No health floors; the death guard widened (Sonnet)

`mvm` changes. The guard is widened here, after the floors go (Q1); Step 6 adds `character.go` to the
item allowlist (Q2).

**Files:** modify `internal/characters/character.go`, `internal/characters/character_test.go`,
`internal/guards/death_test.go`, `internal/hooks/testdata/combat-golden/mvm.txt` (by `-update`).

- [ ] **Step 1: Tests, failing first.** Append to `internal/characters/character_test.go`:

```go

// Phase 5 slice 4.1: the engine applies damage as given. No floor below 0;
// the cap at the maximum stays.
func TestApplyHealthChangeHasNoFloor(t *testing.T) {
	c := New()
	c.HealthMax.Value, c.Health = 10, 10
	assert.Equal(t, -30, c.ApplyHealthChange(-30))
	assert.Equal(t, -20, c.Health, "a 30 hit on 10 health")
	assert.Equal(t, -10, c.ApplyHealthChange(-10))
	assert.Equal(t, -30, c.Health, "a hit on a body already below 0")

	c.Health = 8
	assert.Equal(t, 2, c.ApplyHealthChange(50))
	assert.Equal(t, 10, c.Health, "a heal above the maximum stops there")
}

// Phase 5 slice 4.1: a saved body's health below 0 loads as saved.
func TestValidateKeepsHealthBelowZero(t *testing.T) {
	c := New()
	c.Health = -50
	require.NoError(t, c.Validate())
	assert.Equal(t, -50, c.Health)
}
```

  `ApplyHealthChange` below 0 calls `CancelBuffsWithFlag`, whose `buffs.HasFlag` logs, and the
  package's tests set up no logger (run: nil-pointer panic in `slog`). In `TestMain` (`:55`), first
  line, add `	mudlog.SetupLogger(nil, "", "", false) // ApplyHealthChange below 0 asks buffs, which log`,
  and add `"github.com/GoMudEngine/GoMud/internal/mudlog"` to the imports after the `ids` line
  (`:12`).

  Run `go test ./internal/characters/ -run 'TestApplyHealthChangeHasNoFloor|TestValidateKeepsHealthBelowZero' -count=1`.
  Expected: both FAIL, with these mismatches (run at `1c8e607f`):

```
--- FAIL: TestApplyHealthChangeHasNoFloor
    expected: -30  actual: -15        (the return value of the 30 hit: capped at -5 from 10)
    expected: -20  actual: -5         "a 30 hit on 10 health"
    expected: -10  actual: -5         (the return value of the next 10 hit: floored at -10 from -5)
    expected: -30  actual: -10        "a hit on a body already below 0"
--- FAIL: TestValidateKeepsHealthBelowZero
    expected: -50  actual: -10
```

- [ ] **Step 2: `ApplyHealthChange`** (`character.go:896-916`). The block

```go
	if newHealth < 0 {
		c.CancelBuffsWithFlag("cancel-on-combat")

		// If they haven't dropped yet, require a drop before going straight to death.
		// Don't allow players to drop under -5 in a single hit.
		if newHealth < -5 && oldHealth > 0 {
			newHealth = -5
		} else if newHealth <= -10 {
			newHealth = -10
		}
	} else if newHealth > c.HealthMax.Value {
```

  becomes

```go
	if newHealth < 0 {
		c.CancelBuffsWithFlag("cancel-on-combat")
	} else if newHealth > c.HealthMax.Value {
```

- [ ] **Step 3: `Validate`** (`:1205-1208`): delete

```go
	if c.Health < -10 {
		c.Health = -10
	}

```

- [ ] **Step 4: Tests pass; mutation probes.**

```bash
go test ./internal/characters/ -count=1
```

| Mutation | Test | Failure seen on the scratch copy |
|---|---|---|
| `newHealth := c.Health + healthChange` to `newHealth := max(c.Health+healthChange, -10)` | `-run TestApplyHealthChangeHasNoFloor` | `Not equal` |
| `} else if newHealth > c.HealthMax.Value {` to `} else if newHealth > c.HealthMax.Value+100 {` | same | `Not equal` |
| before `if c.Mana < 0 {` in `Validate`, insert `if c.Health < -10 { c.Health = -10 }` | `-run TestValidateKeepsHealthBelowZero` | `Not equal` |

- [ ] **Step 5: The guard's scope.** In `internal/guards/death_test.go`, the comment and variable
  (`:12-16`) become:

```go
// deathScanned is the engine code where the death guards apply: the round
// and command code (slice 3), and the bodies and rooms (slice 4.1, where
// characters asks the ruleset through characters.IsUp). The rulesets are
// outside it: comparing health is their job.
var deathScanned = []string{"internal/hooks/", "internal/usercommands/", "internal/death/", "internal/characters/", "internal/rooms/"}
```

- [ ] **Step 6: The `IsDisabled` check (Q2).** The allowlist (`:18-24`) becomes:

```go
// itemIsDisabled lists the scanned files that call items.Item.IsDisabled
// (a disabled equipment slot, not a body's condition), with why. A body has
// no IsDisabled since slice 4.1; any other IsDisabled call in a scanned file
// must be listed here.
var itemIsDisabled = map[string]string{
	"internal/characters/character.go":          "Validate, BestUpgrades and Wear handle a slot the race cannot use",
	"internal/usercommands/look.panels.go":      "the equipment and corpse panels skip a disabled slot",
	"internal/usercommands/inventory.panels.go": "the inventory panel skips a disabled slot",
}
```

  The planted source's last two lines (`:34-35`) become (the body call goes):

```go
	_ = itm.IsDisabled()
	return c.Health <= -10 || 1 > c.Health || c.Health == (0) || c.Health > c.HealthMax
```

  In `TestEngineAsksTheRulesetWhoIsDead`, the literal-compare report (`:74-76`)

```go
		for range literal {
			t.Errorf("%s: compares .Health with a number; ask rules.Get().Condition (slice 3)", path)
		}
```

  becomes (`characters` cannot import `rules`, so its files are pointed at `characters.IsUp`):

```go
		ask := "rules.Get().Condition"
		if p == "internal/characters/" {
			ask = "characters.IsUp (characters cannot import rules)"
		}
		for range literal {
			t.Errorf("%s: compares .Health with a number; ask %s", path, ask)
		}
```

  (`p` is the `deathPrefix(path)` the loop already holds, `:66`.) Then the loop (`:77-87`)

```go
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
```

  becomes

```go
		if isDisabledCalls(f) > 0 {
			if _, ok := itemIsDisabled[path]; !ok {
				t.Errorf("%s: calls .IsDisabled(); an item's: add the file to itemIsDisabled with why (a body has none: ask rules.Get().Condition)", path)
				continue
			}
			itemCalls[path] = true
		}
```

  The control (`:109-126`, from `// ... and the matchers find exactly what is planted` through the
  `onCharacter != 1 || other != 1` block) becomes:

```go
	// ... and the matchers find exactly what is planted: three literal
	// compares (signed, reversed, in brackets) and not the compare with a
	// variable; one IsDisabled call.
	planted := parsePlanted(t)
	if literal, compares := healthLiteralCompares(planted); len(literal) != 3 || compares != 4 {
		t.Errorf("control: found %d literal health compares of 3 and %d health compares of 4 in the planted source; the literal check is broken", len(literal), compares)
	}
	if n := isDisabledCalls(planted); n != 1 {
		t.Errorf("control: found %d IsDisabled calls of 1 in the planted source; the IsDisabled check is broken", n)
	}
```

  And `isDisabledCalls` with its comment (`:177-198`) becomes:

```go
// isDisabledCalls counts the X.IsDisabled() calls in f. The guard sees
// syntax, not types, so every one counts; only the itemIsDisabled files may
// have any. It no longer tells a body's call from an item's, which is sound
// while Character.IsDisabled does not exist (deleted in slice 4.1); a body
// method of that name would need the distinction back.
func isDisabledCalls(f *ast.File) int {
	n := 0
	ast.Inspect(f, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "IsDisabled" {
			n++
		}
		return true
	})
	return n
}
```

- [ ] **Step 7: The guard passes, and fails on a plant in each new package.**

```bash
go test ./internal/guards/ -count=1
```

  Expected: `ok`. Probes (Edit, run `go test ./internal/guards/ -run TestEngineAsksTheRulesetWhoIsDead -count=1`, undo):

| Plant | Failure seen on the scratch copy |
|---|---|
| `characters/isup.go`: `return IsUp(c)` to `return c.Health < 1 \|\| IsUp(c)` | `internal/characters/isup.go: compares .Health with a number; ask characters.IsUp (characters cannot import rules)` |
| `rooms/rooms.go`, `GetMobs`: `typeFlag&FindBuffed == FindBuffed && len(mob.Character.Buffs.List) > 0 {` to `typeFlag&FindBuffed == FindBuffed && mob.Character.Health < 1 {` | `internal/rooms/rooms.go: compares .Health with a number; ask rules.Get().Condition` |
| `deathScanned` back to the three slice 3 prefixes | `stale allowlist: internal/characters/character.go (...) no longer calls IsDisabled` |

- [ ] **Step 8: The `mvm` golden.**

```bash
go test ./internal/hooks/ -count=1                                    # FAIL: mvm.txt differs at line 37
go test ./internal/hooks/ -run 'TestCombatGoldenMvM$' -count=1 -update
git diff -- internal/hooks/testdata/
```

  Expected diff, exactly one line (decided: the raw value, no -5 cap):

```diff
-  rat: room=93001 hp=-5/12 aggro=[none] playerdamage={}
+  rat: room=93001 hp=-9/12 aggro=[none] playerdamage={}
```

  Then `go test ./internal/hooks/ -count=1` prints `ok`. Any other golden line changing is
  unexpected: stop.

- [ ] **Step 9: Per-task check, then commit.**

```bash
git add internal/characters/character.go internal/characters/character_test.go \
  internal/guards/death_test.go internal/hooks/testdata/combat-golden/mvm.txt
git commit -F - <<'EOF'
Phase 5.4.1: no health floors in the engine; death guard covers characters and rooms

ApplyHealthChange loses the -5 one-hit cap and the -10 floor, and
Validate its -10 clamp: health goes where the damage puts it (the cap at
the maximum and the cancel-on-combat cancel below 0 stay). The death
guard now scans internal/characters and internal/rooms; a body has no
IsDisabled, so every IsDisabled call there is an item's and the file is
allowlisted.

Golden: mvm, the dead rat reads hp=-9/12 (was -5/12, the cap). Decided.
Every other golden unchanged.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 4: NPC regen (Sonnet)

The commit 1 fixture changes. The gate stays "out of a fight" = "has no aggro of its own" (the spec's
regen section): this task only drops the account gate. From Task 6a on, a body that is attacked has
an aggro (retaliation), and a body whose aggro waits on a target in another room keeps it, so neither
regenerates; nothing in this task changes for that.

**Files:** modify `internal/rules/dikubasic/ruleset.go`, `internal/rules/dikubasic/ruleset_test.go`,
`internal/hooks/combat_golden_test.go` (a comment), `internal/hooks/testdata/combat-golden/npc_regen.txt`
(by `-update`).

- [ ] **Step 1: The test, failing first.** In `ruleset_test.go`, replace
  `TestRegenOutsideTheRecoveryRoomIsTheOldAutoHeal` with its comment (`:240-267`, from
  `// Parity with the old AutoHeal` through the closing `}` before `// reseed puts`) by:

```go
// Outside the recovery room every actor regenerates by one rule (Phase 5
// slice 4.1): every 3rd round, out of a fight and above 0 health, it regains
// HealthPerRound and ManaPerRound, an NPC as a player does.
func TestRegenOutsideTheRecoveryRoomIsOneRuleForEveryActor(t *testing.T) {
	const here, recovery = 94005, 94006
	setConfig(t,
		map[string]any{"SpecialRooms.DeathRecoveryRoom": recovery},
		map[string]any{"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)})
	none := rules.RegenResult{}

	for name, a := range map[string]*actors.Actor{"player": player(t, 9621), "npc": npc(t, 0, 0)} {
		a.Character.RoomId, a.Character.Health = here, 5
		full := rules.RegenResult{Health: a.Character.HealthPerRound(), Mana: a.Character.ManaPerRound()}
		require.NotZero(t, full.Health, name)
		assert.Equal(t, full, New().Regen(a, 3), "%s: every 3rd round", name)
		assert.Equal(t, full, New().Regen(a, 6), name)
		assert.Equal(t, none, New().Regen(a, 4), "%s: only every 3rd round", name)
		assert.Equal(t, none, New().Regen(a, 5), name)

		a.Character.Aggro = &characters.Aggro{Type: characters.DefaultAttack}
		assert.Equal(t, none, New().Regen(a, 3), "%s: never in a fight", name)
		a.Character.Aggro = nil

		a.Character.Health = 0
		assert.Equal(t, none, New().Regen(a, 3), "%s: never at 0 health", name)
	}
}
```

  Run `go test ./internal/rules/dikubasic/ -run TestRegenOutsideTheRecoveryRoomIsOneRule -count=1`.
  Expected: FAIL, `npc: every 3rd round` (`expected: rules.RegenResult{Health:1, Mana:1, ...}`).

- [ ] **Step 2: `Regen`** (`ruleset.go:125-141`). The comment (`:125-129`) becomes:

```go
// Regen in the death recovery room (or a private copy of it) is fast:
// recoveryRegen. Elsewhere every actor, player or NPC, regenerates by one
// rule (Phase 5 slice 4.1): every 3rd round, out of a fight and above 0
// health, it regains HealthPerRound and ManaPerRound.
```

  and `:134` `	if accountOf(a) == nil || round%3 != 0 || a.Character.Aggro != nil {` becomes
  `	if round%3 != 0 || a.Character.Aggro != nil {`. (`accountOf` stays: `combat.go` uses it.)

  In `internal/hooks/combat_golden_test.go`, `TestCombatGoldenNPCRegen`'s comment (Task 1) lines
  `// a "break" would leave them) and the regen pass runs four times (Phase 5` /
  `// slice 4.1, recorded on today's code first: an NPC does not regenerate).` become
  `// a "break" would leave them) and the regen pass runs four times (Phase 5` /
  `// slice 4.1: recorded before NPC regen; since then the rat heals each pass).`

- [ ] **Step 3: Pass; mutation probes.** `go test ./internal/rules/dikubasic/ -count=1` prints `ok`.

| Mutation (in `Regen`) | Failure seen on the scratch copy |
|---|---|
| back to `if accountOf(a) == nil \|\| round%3 != 0 \|\| a.Character.Aggro != nil {` | `npc: every 3rd round`, `Not equal` |
| `if round%3 != 0 \|\| a.Character.Aggro != nil {` to `if round%3 != 0 {` | `never in a fight`, `Not equal` |
| `Regen`'s `if a.Character.Health <= 0 {`, the one right after `if round%3 != 0 \|\| a.Character.Aggro != nil {` and its `return rules.RegenResult{}` (`Condition` has the same text: leave that one), to `if a.Character.Health < 0 {` | `never at 0 health`, `Not equal` |

- [ ] **Step 4: The golden.**

```bash
go test ./internal/hooks/ -count=1                                       # FAIL: npc_regen.txt differs at line 27
go test ./internal/hooks/ -run 'TestCombatGoldenNPCRegen$' -count=1 -update
git diff -- internal/hooks/testdata/
```

  Expected diff, exactly four lines (decided: the rat regains 1 a pass, `HealthPerRound` with no
  stat mods):

```diff
@@ heal 1
-  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
+  rat: room=93001 hp=31/40 aggro=[none] playerdamage={Alice:10}
@@ heal 2
-  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
+  rat: room=93001 hp=32/40 aggro=[none] playerdamage={Alice:10}
@@ heal 3
-  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
+  rat: room=93001 hp=33/40 aggro=[none] playerdamage={Alice:10}
@@ heal 4
-  rat: room=93001 hp=30/40 aggro=[none] playerdamage={Alice:10}
+  rat: room=93001 hp=34/40 aggro=[none] playerdamage={Alice:10}
```

  (file lines 27, 33, 39, 45). No other golden changes: the death fixtures' `heal` passes record none
  (death is immediate), and no other fixture runs a heal pass. Then `go test ./internal/hooks/ -count=1`
  prints `ok`.

- [ ] **Step 5: Per-task check, then commit.**

```bash
git add internal/rules/dikubasic/ruleset.go internal/rules/dikubasic/ruleset_test.go \
  internal/hooks/combat_golden_test.go internal/hooks/testdata/combat-golden/npc_regen.txt
git commit -F - <<'EOF'
Phase 5.4.1: every actor regenerates by one rule

dikubasic.Regen drops its account gate: outside the recovery room any
body out of a fight and above 0 health regains HealthPerRound and
ManaPerRound every 3rd round. An NPC wounded and left alone now heals.

Golden: npc_regen, the rat goes 30, 31, 32, 33, 34 over the four heal
passes (was 30 throughout). Decided. Every other golden unchanged.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 5: Offhand breakage gone (Opus; the controller reads the golden diff)

`pvp` changes, with the shifted draws decided (Q3).

**Files:** modify `internal/hooks/NewRound_DoCombat.go`, `internal/items/items.go`,
`internal/items/itemspec.go`, `internal/items/items_test.go`, `_datafiles/html/admin/items.html`,
`internal/hooks/combat_golden_test.go`, `internal/hooks/testdata/combat-golden/pvp.txt` (by `-update`).

- [ ] **Step 1: The load test, passing before and after.** In `internal/items/items_test.go` the
  imports (`:3-10`) become:

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)
```

  and append:

```go

// An item file from before Phase 5 slice 4.1 may still have a breakchance
// key: the game's item loader (LoadAllFlatFiles, as LoadDataFiles calls it)
// loads it, and the keys after it too.
func TestOldItemFileWithBreakKeyLoads(t *testing.T) {
	dir := t.TempDir()
	spec := ItemSpec{ItemId: 1, Name: "old shield"}
	path := filepath.Join(dir, filepath.FromSlash(spec.Filepath()))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	text := "itemid: 1\nname: old shield\ntype: offhand\nbreakchance: 50\ndamagereduction: 10\n"
	require.NoError(t, os.WriteFile(path, []byte(text), 0o644))

	loaded, err := fileloader.LoadAllFlatFiles[int, *ItemSpec](dir)
	require.NoError(t, err)
	require.Contains(t, loaded, 1)
	assert.Equal(t, "old shield", loaded[1].Name)
	assert.Equal(t, Offhand, loaded[1].Type)
	assert.Equal(t, 10, loaded[1].DamageReduction)
}
```

  Run `go test ./internal/items/ -run TestOldItemFileWithBreakKeyLoads -count=1`: `ok` today (the
  field exists). It must still pass after Step 3, when the key names nothing. (The name avoids the
  strings Step 6's empty grep looks for; the YAML key `breakchance` is lower case and does not match
  them.)

- [ ] **Step 2: The call and the function.** In `NewRound_DoCombat.go` delete (`:243-247`):

```go
	// If the attack connected, check for damage to equipment.
	if roundResult.Hit {
		offhandBreakCheck(d, defRoom, roundResult.Crit)
	}

```

  and the whole of `offhandBreakCheck` with its comment (`:341-390`: from
  `// offhandBreakCheck may break the defender's offhand item after a hit. A` through its closing `}`)
  plus the blank line after, so `// resumeAfterFlee puts` follows the previous function's blank line.
  Every import stays in use (built on the scratch copy).

- [ ] **Step 3: The fields.** In `internal/items/items.go` delete `BreakTest` with its comment
  (`:98-114`, from `// performs a break test and returns true if the item breaks` through the
  closing `}`) and the blank line after. In `internal/items/itemspec.go` delete the field line
  (`:212`)
  ``	BreakChance     uint8             `yaml:"breakchance,omitempty"` // Chance in 100 that the item will break ...``
  and the clamp (`:488-491`):

```go
	if i.BreakChance > 100 {
		i.BreakChance = 100
	}

```

  `gofmt -w internal/items/itemspec.go` realigns the three tag comments below the deleted field; that
  is expected.

- [ ] **Step 4: The admin editor** (`_datafiles/html/admin/items.html`). Delete the field block
  (`:293-296`):

```html
                <div class="field">
                    <label>Break Chance %</label>
                    <input type="number" id="f-breakchance" min="0" max="100" placeholder="0" />
                </div>
```

  the load line (`:508`) `        setVal('f-breakchance', item.BreakChance || 0);` and the save line
  (`:648`) `            BreakChance:     parseInt(document.getElementById('f-breakchance').value, 10) || 0,`.
  Then `grep -n -i "break" _datafiles/html/admin/items.html` prints nothing (run on its own).

- [ ] **Step 5: The golden fixture.** In `combat_golden_test.go`:
  - `:82` `// Attack message text, the broken-item spec and race dice come from the` becomes
    `// Attack message text and race dice come from the`.
  - `shield` (`:270-276`) becomes:

```go
func shield(itemId int, name string) items.Item {
	spec := items.ItemSpec{
		ItemId: itemId, Name: name, NameSimple: name, Type: items.Offhand, Subtype: items.Wearable,
		DamageReduction: 10,
	}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}
```

  - `TestCombatGoldenPvP`'s comment (`:603-604`) becomes
    `// A player attacks a player who does not fight back; the defender's shield` /
    `// blocks some of the damage.` and `:611` becomes
    `	bob.Character.Equipment.Offhand = shield(93102, "testshield")`.

- [ ] **Step 6: Build and the load test.**

```bash
go build ./...
go test ./internal/items/ -count=1
grep -rn --exclude-dir=.claude "BreakChance\|BreakTest\|offhandBreakCheck\|breakChance" --include=*.go --include=*.html . # nothing; run on its own
grep -rn "DamageReduction" internal/items/itemspec.go | head -1                                     # one line: the pattern can match
```

  Mutation probe on the loader the game uses for items (`LoadAllFlatFiles`, called by
  `itemspec.go:530`): in `internal/fileloader/fileloader.go:181` (the same text is also at `:76`, so
  Edit with the line above it for context: `		var loaded T` / blank / `		err = yaml.Unmarshal(bytes, &loaded)`,
  two tabs; `:76` has one tab and no `var loaded T` above it),
  `err = yaml.Unmarshal(bytes, &loaded)` to `err = yaml.UnmarshalStrict(bytes, &loaded)`;
  `go test ./internal/items/ -run TestOldItemFileWithBreakKeyLoads -count=1` fails with
  `Received unexpected error: yaml: unmarshal errors: line 4: field breakchance not found in type items.ItemSpec`
  (run on the scratch copy). Undo. (The same change at `:76`, in `LoadFlatFile`, leaves the test
  passing: it does not take that path.)

- [ ] **Step 7: The `pvp` golden.**

```bash
go test ./internal/hooks/ -count=1                                     # FAIL: pvp.txt differs at line 19
go test ./internal/hooks/ -run 'TestCombatGoldenPvP$' -count=1 -update
git diff -- internal/hooks/testdata/
```

  Expected diff (run on the scratch copy). Why it is larger than two lines: `BreakTest` drew one
  random number on every hit on Bob; without it, every draw from round 2 on shifts, so the attack
  lines and damage change; Bob still dies in round 4 and now keeps his shield:

```diff
@@ -16,9 +16,9 @@
   Bob: room=93001 hp=10/15 aggro=[none] playerdamage={Alice:5} offhand=93102
 == round 2
 -- Alice reads
-  You wave your <ansi fg="item">testsword</ansi> at <ansi fg="username">Bob</ansi>, to no avail!
+  You slash at <ansi fg="username">Bob</ansi> but fail to connect!
 -- Bob reads
-  <ansi fg="username">Alice</ansi> attempts to slash you but misses!
+  <ansi fg="username">Alice</ansi> waves their <ansi fg="item">testsword</ansi> at you, but does nothing.
 -- events
   AggroChanged Alice room=93001
 -- state
@@ -26,32 +26,26 @@
   Bob: room=93001 hp=10/15 aggro=[none] playerdamage={Alice:5} offhand=93102
 == round 3
 -- Alice reads
-  You deliver a powerful blow to <ansi fg="username">Bob</ansi>, causing an impressive <ansi fg="damage">6 damage</ansi>!
+  You slash <ansi fg="username">Bob</ansi> with your <ansi fg="item">testsword</ansi>, causing <ansi fg="damage">4 damage</ansi>!
 -- Bob reads
-  In a sweeping motion, <ansi fg="username">Alice</ansi> slashes you with their <ansi fg="item">testsword</ansi>, doing an impressive <ansi fg="damage">6 damage</ansi>!
+  <ansi fg="username">Alice</ansi> slashes you with their <ansi fg="item">testsword</ansi>, causing <ansi fg="damage">4 damage</ansi>! <ansi fg="red">[you blocked 1]</ansi>
 -- events
   CharacterVitalsChanged Bob
   AggroChanged Alice room=93001
 -- state
   Alice: room=93001 hp=10/10 aggro=[Bob DefaultAttack wait=0] playerdamage={} weapon=93101
-  Bob: room=93001 hp=4/15 aggro=[none] playerdamage={Alice:11} offhand=93102
+  Bob: room=93001 hp=6/15 aggro=[none] playerdamage={Alice:9} offhand=93102
 == round 4
 -- Alice reads
-  Your <ansi fg="item">testsword</ansi> cuts deeply into <ansi fg="username">Bob</ansi>, dealing significant damage of <ansi fg="damage">7</ansi>!
-  <ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">testshield</ansi> <ansi fg="username">Bob</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>
+  In a sweeping motion, you slash <ansi fg="username">Bob</ansi> with your <ansi fg="item">testsword</ansi>, doing an impressive <ansi fg="damage">7 damage</ansi>!
   <ansi fg="username">Bob</ansi> has died.
   <ansi fg="magenta-bold">***</ansi> <ansi fg="username">Bob</ansi> has <ansi fg="red-bold">DIED!</ansi> (killed by <ansi fg="username">Alice</ansi>) <ansi fg="magenta-bold">***</ansi>
   <ansi fg="yellow-bold">25 gold</ansi> drops to the ground.
 -- Bob reads
-  In a sweeping motion, <ansi fg="username">Alice</ansi> slashes you with their <ansi fg="item">testsword</ansi>, doing an impressive <ansi fg="damage">7 damage</ansi>! <ansi fg="red">[you blocked 1]</ansi>
-  <ansi fg="202">***</ansi>
-  <ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">testshield</ansi> breaks! <ansi fg="202">***</ansi></ansi>
-  <ansi fg="202">***</ansi>
+  You are hit hard as <ansi fg="username">Alice</ansi> cuts deeply into you for <ansi fg="damage">7 damage</ansi>!
   <ansi fg="magenta-bold">***</ansi> <ansi fg="username">Bob</ansi> has <ansi fg="red-bold">DIED!</ansi> (killed by <ansi fg="username">Alice</ansi>) <ansi fg="magenta-bold">***</ansi>
   <ansi fg="yellow-bold">25 gold</ansi> drops to the ground.
 -- events
-  ItemOwnership Bob item=93102 gained=false
-  ItemOwnership Bob item=0 gained=true
   CharacterVitalsChanged Bob
   AggroChanged Alice room=93001
   AggroChanged Bob room=93001
@@ -59,6 +53,5 @@
   CharacterVitalsChanged Bob
 -- state
   Alice: room=93001 hp=10/10 aggro=[none] playerdamage={} weapon=93101
-  Bob: room=copy-of-93003 hp=1/15 aggro=[none] playerdamage={}
-  room 93001 floor: 0
+  Bob: room=copy-of-93003 hp=1/15 aggro=[none] playerdamage={} offhand=93102
   room 93001 floor gold: 25
```

  Round 1 is identical (the first draw for the break came after Alice's round 1 hit). Then
  `go test ./internal/hooks/ -count=1` prints `ok`. Any other golden file changing is unexpected:
  stop.

- [ ] **Step 8: Per-task check, then commit.**

```bash
git add internal/hooks/NewRound_DoCombat.go internal/items/items.go internal/items/itemspec.go \
  internal/items/items_test.go _datafiles/html/admin/items.html \
  internal/hooks/combat_golden_test.go internal/hooks/testdata/combat-golden/pvp.txt
git commit -F - <<'EOF'
Phase 5.4.1: offhand breakage gone

offhandBreakCheck and its call, Item.BreakTest, ItemSpec.BreakChance
and its clamp, the admin item editor's break-chance field and the
golden fixture's breakChance parameter are deleted. An item file that
still has breakchance: loads as before (the key is ignored; tested).

Golden: pvp. Bob keeps his shield and the two break lines and their
ItemOwnership events are gone. BreakTest drew a random number on every
hit on Bob, so from round 2 every later draw shifts: rounds 2-4 read
different attack lines, Bob is at 6/15 after round 3 (was 4/15), and he
still dies in round 4. Decided (Q3). Every other golden unchanged.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 6: `BackStab` gone (Sonnet)

Goldens unchanged. Step 3 corrects the crit comment and keeps the behaviour (Q4); Steps 5 to 7
delete the backstab leftovers (Q5).

**Files:** modify `internal/characters/aggro.go`, `internal/rules/dikubasic/combat.go`,
`internal/hooks/combat_golden_test.go`, `internal/items/itemspec.go`, `cmd/smoke/fight.go`,
`cmd/smoke/fight_test.go`, `_datafiles/html/admin/items-rank-weapons.html`.

- [ ] **Step 1: The value.** `aggro.go:10-12`: delete the line `	BackStab`, so the constants read
  `DefaultAttack` (0), `Shooting` (1), `Flee` (2).

- [ ] **Step 2: The branch.** In `dikubasic/combat.go` delete (`:175-183`, with the blank line after):

```go
		attackMessagePrefix := ``
		// If they are backstabbing it's a free crit
		if sourceChar.Aggro.Type == characters.BackStab {
			attackResult.Crit = true
			attackMessagePrefix = `<ansi fg="magenta-bold">*[BACKSTAB]*</ansi> `
			// Failover to the default attack
			sourceChar.SetAggro(sourceChar.Aggro.Target, characters.DefaultAttack)
		}

```

  and the prefix block (`:276-284`, with the blank line after):

```go
				if len(attackMessagePrefix) > 0 {
					toAttackerMsg = attackMessagePrefix + toAttackerMsg
					toDefenderMsg = attackMessagePrefix + toDefenderMsg
					toAttackerRoomMsg = attackMessagePrefix + toAttackerRoomMsg
					if len(toDefenderRoomMsg) > 0 {
						toDefenderRoomMsg = attackMessagePrefix + toDefenderRoomMsg
					}
				}

```

- [ ] **Step 3: The crit flag's comment (Q4: behaviour kept, comment corrected).** `:243-246`

```go
					// Backstab sets attackResult.Crit for the first hit only; subsequent
					// hits use a fresh per-attack roll so crits don't cascade.
					isCrit = attackResult.Crit || Crits(sourceChar, targetChar)
					attackResult.Crit = false // consume the backstab flag after one use
```

  becomes

```go
					// Once a hit this round is a crit, every later hit this round
					// is one too, with no new crit roll (kept as it was; slice
					// 4.3's hit roll replaces this).
					isCrit = attackResult.Crit || Crits(sourceChar, targetChar)
```

  The comment fix is the two lines at `:243-244`; line `:246`, `attackResult.Crit = false // consume
  the backstab flag after one use`, goes too because it names the deleted backstab and does nothing:
  when `attackResult.Crit` was true `isCrit` is true and the next line sets it true again; when false
  it stays false. The `characters` import stays in use (`characters.Character` in the signatures).

- [ ] **Step 4: The golden fixture's name map** (`combat_golden_test.go:294-296`):
  `characters.Shooting: "Shooting", characters.BackStab: "BackStab",` becomes
  `characters.Shooting: "Shooting",`. Run `gofmt -w` on the file.

- [ ] **Step 5: `items.CanBackstab`** (`internal/items/itemspec.go:364-369`, no caller): delete it and
  the blank line after:

```go
func CanBackstab(iSubType ItemSubType) bool {
	if iSubType == Cleaving || iSubType == Stabbing || iSubType == Slashing || iSubType == Claws {
		return true
	}
	return false
}

```

- [ ] **Step 6: The smoke client.** In `cmd/smoke/fight.go`, `stripCombatMarks` with its comment
  (`:122-133`)

```go
// stripCombatMarks removes the leading marks combat puts on a line (from
// internal/rules/dikubasic/combat.go, ansi stripped): "*[BACKSTAB]* " on a backstab
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
```

  becomes

```go
// stripCombatMarks removes the leading mark combat puts on a crit line (from
// internal/rules/dikubasic/combat.go, ansi stripped): "*** ".
func stripCombatMarks(line string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, "***"))
}
```

  (The crit form `*** <line> ***` is the only leading mark left: `combat.go:267-273` wraps every crit
  line in `<ansi fg="yellow-bold">***</ansi> ... ***`. One trim is enough; the loop existed for the
  backstab-then-crit order. The caller, `checkObserverText` `:144`, strips the prompt first.)

  In `cmd/smoke/fight_test.go`, the two rows (`:44-45`)

```go
		{"backstab crit line", "*[BACKSTAB]* *** Smkabc0 hits rat with their fists. ***\n", false},
		{"backstab line after prompt", "[☀️ 4:01AM HP:6/6 MP:5/5]:*[BACKSTAB]* Smkabc0 hits rat.\n", false},
```

  become one row that keeps the after-prompt case for the crit form (the row `"crit line"` at `:43`
  already covers a crit line on its own):

```go
		{"crit line after prompt", "[☀️ 4:01AM HP:6/6 MP:5/5]:*** Smkabc0 hits rat with their fists. ***\n", false},
```

  Mutation probe: `stripCombatMarks`'s body to `return line`; `go test ./cmd/smoke/ -run TestCheckObserverText -count=1`
  fails on both crit rows (run on the scratch copy:
  `crit line: err = the observer read no line naming Smkabc0 and the rat (other than "prepares to fight"), want error false`
  and the same for `crit line after prompt`). Undo.

- [ ] **Step 7: The admin tooltip** (`_datafiles/html/admin/items-rank-weapons.html:208`): in the
  `Subtype` header's `data-tip`, the text
  `Determines which attack messages are used and which skills apply (e.g. backstab requires cleaving, stabbing, slashing, or claws).`
  becomes `Determines which attack messages are used and which skills apply.` Leave
  `_datafiles/html/admin/keywords-api.html` alone (a sample API response; slice 7's docs sweep).

- [ ] **Step 8: Verify.**

```bash
go build ./...
grep -rn --exclude-dir=.claude -i "backstab\|attackMessagePrefix" --include=*.go --include=*.html . # only keywords-api.html:57; run on its own
grep -rn "characters.Flee" --include=*.go internal | head -2   # lines: the pattern can match
go test ./internal/hooks/ ./internal/rules/... ./internal/usercommands/ ./internal/items/ ./cmd/smoke/ -count=1
```

  Expected: the grep finds only `_datafiles/html/admin/keywords-api.html:57`; all tests `ok`, no
  `-update`: every combat and reads golden byte-identical (run on the scratch copy).

- [ ] **Step 9: Per-task check, then commit.**

```bash
git add internal/characters/aggro.go internal/rules/dikubasic/combat.go internal/hooks/combat_golden_test.go \
  internal/items/itemspec.go cmd/smoke/fight.go cmd/smoke/fight_test.go \
  _datafiles/html/admin/items-rank-weapons.html
git commit -F - <<'EOF'
Phase 5.4.1: BackStab gone

The AggroType value and the dikubasic branch are deleted; Flee moves
from 3 to 2. A sidecar saved mid-flee before this holds type 3, which
now names nothing: it loads as saved, and on the NPC's next turn the
round treats it as an attack on its target, finds none (a flee has no
target), and ends the aggro. No shipped data holds one.

Also gone: items.CanBackstab (no caller), the smoke client's
*[BACKSTAB]* strip and its two test rows (a crit-after-prompt row keeps
that case), and backstab in the admin weapon-rank tooltip. The crit
flag's comment now says what the code does: a crit makes every later
hit that round a crit, with no roll (kept; 4.3 decides crits). Goldens
unchanged.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 6a: Retaliation and target switching (Opus; the controller reads every golden hunk)

Spec "Retaliation and target switching (rulings 6-9)". Every step here was applied to the scratch
copy after Tasks 1-6 and run (2026-11-12), first as asked and again after the controller's answers
to Q7-Q17 (the code below is the last run's). Combat goldens change; the expected
change per fixture is in Step 9, written before the run, with what the run showed.

**Files:** modify `internal/characters/aggro.go`, `internal/characters/character.go`,
`internal/characters/character_test.go`, `internal/death/death.go`, `internal/hooks/NewRound_DoCombat.go`,
`internal/hooks/combat_golden_test.go`, `internal/mobs/saved_test.go`,
`internal/rules/dikubasic/ruleset.go`, `internal/rules/dikubasic/ruleset_test.go`,
`internal/usercommands/go.go`, `internal/usercommands/go_test.go`, `internal/usercommands/flee.go`,
`internal/usercommands/shoot_test.go`, `internal/usercommands/quit_test.go`,
`internal/hooks/death_engine_test.go` (a comment), `cmd/smoke/server.go`, `cmd/smoke/fight.go`,
`cmd/smoke/fight_test.go`, `cmd/smoke/main.go`, `cmd/smoke/script.go` (a comment),
`internal/configs/config.gameplay.go`, `internal/behaviour/tick.go`, `internal/behaviour/tick_test.go`,
`_datafiles/config.yaml`, `_datafiles/html/admin/config.html`, `_datafiles/html/admin/config-wizard.html`,
every `internal/hooks/testdata/combat-golden/*.txt` and
`internal/hooks/testdata/reads-golden/E-move-look.txt` (by `-update`); create
`internal/death/switch.go`, `internal/hooks/retaliation_test.go`,
`internal/hooks/retaliation_golden_test.go`, `internal/configs/config.gameplay_test.go`,
`internal/mutators/testhelpers.go`,
`cmd/smoke/fightrats_test.go` and eight golden files (by `-update`).

- [ ] **Step 1: `Since` and `OutOfReach` on the aggro.** `internal/characters/aggro.go`, after the
  `FleeFrom` field (`:20-22`), inside the struct:

```go
	// Since is the round this aggro began on its target: set when the target
	// changes, kept when the same target is re-armed (SetAggro). The longest
	// attacker is the one a body turns to next (Phase 5 slice 4.1, ruling 7).
	// An NPC sidecar saved before 4.1 loads 0, which reads as longest.
	Since uint64
	// OutOfReach counts this aggro's consecutive turns that found its target
	// out of reach (ruling 9's wait); a turn that finds it reachable, in the
	// room or through an exit, resets it. At the configured limit the aggro
	// ends (ruling 12, GamePlay.OutOfReachRounds). NPC sidecars save it with
	// the aggro, so a restored NPC resumes its count; a sidecar saved before
	// 4.1 loads 0.
	OutOfReach int
	// OutOfReachTold records that this wait has told the body "Your target
	// is out of reach.", so the line comes once per wait (plan review round
	// 6). Only a turn that finds the target reachable and perceived clears
	// it, not being attacked by the target (which resets OutOfReach). Saved
	// in NPC sidecars with OutOfReach.
	OutOfReachTold bool
```

  (NPC sidecars save the aggro by plain YAML, `mobs/saved.go:20`, so `since:`, `outofreach:` and
  `outofreachtold:` are saved; a player's aggro is `yaml:"-"`, `character.go:73`. Saving the count
  and the mark is the plain-YAML default for every aggro field and keeps a mid-wait NPC's wait as it
  was across a rest or restart, rather than giving it a fresh 5 rounds and a second line. The mark is
  separate from the count because being attacked by the target resets the count (Step 5, plan
  review round 6) and must not repeat the line: with the line tied to count 1, steady fire from the
  next room told the body every round (12 lines in 12 rounds, scratch run). Re-arming with
  `SetAggro` makes a new record, so the count and the mark start at 0 and false after every swing.) In `SetAggro` (`character.go:789-813`)
  the final assignment becomes:

```go
	since := util.GetRoundCount()
	if c.Aggro != nil && c.Aggro.Target == target {
		since = c.Aggro.Since // the same fight, re-armed
	}

	c.Aggro = &Aggro{
		Target:        target,
		Type:          aggroType,
		RoundsWaiting: combatAddlWaitRounds,
		Since:         since,
	}
```

  (`util` is already imported by `character.go`.) A failed flee re-arms the fled target through a
  `Flee` aggro whose target is zero; Step 5c carries the `Since` through it (Q15).

- [ ] **Step 2: Unit tests for `Since`, failing first.** Append to `character_test.go` (and add
  `"github.com/GoMudEngine/GoMud/internal/util"` to its imports):

```go

// Phase 5 slice 4.1, ruling 7: an aggro records the round it began on its
// target; re-arming the same target keeps it, a new target sets it.
func TestSetAggroSinceIsSetOnChangeAndKeptOnReArm(t *testing.T) {
	prev := util.GetRoundCount()
	t.Cleanup(func() { util.SetRoundCount(prev) })
	c := New()
	first, second := ids.New(ids.NoCell), ids.New(ids.NoCell)

	util.SetRoundCount(500)
	c.SetAggro(first, DefaultAttack)
	assert.Equal(t, uint64(500), c.Aggro.Since)

	util.SetRoundCount(510)
	c.SetAggro(first, DefaultAttack)
	assert.Equal(t, uint64(500), c.Aggro.Since, "the same target, re-armed")

	c.SetAggro(second, DefaultAttack)
	assert.Equal(t, uint64(510), c.Aggro.Since, "a new target")
}
```

  Append to `internal/mobs/saved_test.go`:

```go

// Phase 5 slice 4.1: a sidecar keeps the aggro's Since, and its wait
// (OutOfReach and OutOfReachTold). (One saved before 4.1 has no since key
// and loads 0, YAML's zero value for a missing key; nothing in this slice
// can change that, so no test claims it.)
func TestSavedAggroKeepsSince(t *testing.T) {
	withTemplate(t, 7, nil)
	m := Mob{MobId: 7, HomeRoomId: 2}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.Aggro = &characters.Aggro{Target: ids.New(ids.NoCell), Since: 77, OutOfReach: 3, OutOfReachTold: true}

	b, err := yaml.Marshal(Snapshot(&m))
	require.NoError(t, err)
	require.Contains(t, string(b), "since: 77")
	require.Contains(t, string(b), "outofreachtold: true")
	var back SavedNPC
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, uint64(77), back.Mob().Character.Aggro.Since)
	assert.Equal(t, 3, back.Mob().Character.Aggro.OutOfReach, "a restored NPC resumes its wait count")
	assert.True(t, back.Mob().Character.Aggro.OutOfReachTold, "and does not repeat the out-of-reach line")
}
```

  (Plan review round 4: an earlier half of this test stripped `since:` and checked 0 was loaded; no
  code change could make it fail, so it is gone, and the plan no longer claims a test for it.)

  Order: write these two tests before Step 1's code, run
  `go test ./internal/characters/ ./internal/mobs/ -run 'TestSetAggroSince|TestSavedAggroKeepsSince' -count=1`
  (build failure: `unknown field Since`), then Step 1, then `ok`.

- [ ] **Step 3: The switch rule** (Q14). Create `internal/death/switch.go`
  (`hooks` imports `death`, so the rule lives where both the round and `Die` reach it):

```go
package death

import (
	"sort"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
)

// NextOpponent is the body a turns to when its target is gone (Phase 5
// slice 4.1, ruling 7): of the bodies in a's room that are up, that a
// perceives, that the ruleset lets a attack (rules.CanAttack), and whose
// aggro targets a, the one whose aggro began earliest (Aggro.Since); ties go
// to turn order. Nil when there is none.
func NextOpponent(a *actors.Actor) *actors.Actor {
	r := rooms.LoadRoom(a.Character.RoomId)
	if r == nil {
		return nil
	}
	candidates := []*actors.Actor{}
	for _, id := range r.GetBodies() {
		b := actors.Get(id)
		if b == nil || b.Id == a.Id || b.Character == nil || b.Character.Aggro == nil {
			continue
		}
		if b.Character.Aggro.Target != a.Id {
			continue // the room's bodies are all in a's room; no RoomId check (review round 5)
		}
		if rules.Get().Condition(b) != rules.Up || !rooms.Perceives(a.Character, b.Character) {
			continue
		}
		if rules.Get().CanAttack(a, b) != nil {
			continue // the ruleset refuses it (two players where PvP is off)
		}
		candidates = append(candidates, b)
	}
	var next *actors.Actor
	for _, b := range rules.Get().TurnOrder(SortBodies(candidates)) {
		if next == nil || b.Character.Aggro.Since < next.Character.Aggro.Since {
			next = b
		}
	}
	return next
}

// SortBodies orders bodies as the round lists them before the ruleset's
// turn order (Phase 5 slice 4.1): players by UserId, then NPCs by MobId,
// then any other body; ties by body ID text. It sorts in place and returns
// the slice.
func SortBodies(bodies []*actors.Actor) []*actors.Actor {
	type keyed struct {
		body    *actors.Actor
		kind, n int
		id      string
	}
	list := make([]keyed, len(bodies))
	for i, a := range bodies {
		k := keyed{body: a, kind: 2, id: a.Id.String()}
		if acct := actors.AccountOf(a.Id); acct != nil {
			k.kind, k.n = 0, acct.UserId
		} else if m := mobs.GetInstance(a.Id); m != nil {
			k.kind, k.n = 1, int(m.MobId)
		}
		list[i] = k
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].kind != list[j].kind {
			return list[i].kind < list[j].kind
		}
		if list[i].n != list[j].n {
			return list[i].n < list[j].n
		}
		return list[i].id < list[j].id
	})
	for i := range list {
		bodies[i] = list[i].body
	}
	return bodies
}

// Switch turns a to its next opponent (NextOpponent), with the line naming
// it, and reports whether there was one. With none, a's aggro is left as it
// is: the caller ends it with the line for its case.
func Switch(a *actors.Actor) bool {
	b := NextOpponent(a)
	if b == nil {
		return false
	}
	a.Character.SetAggro(b.Id, characters.DefaultAttack)
	messaging.Send(perception.Message{
		Kind:      perception.WorldEvent,
		Actor:     a.Id,
		Target:    b.Id,
		Audience:  perception.ActorOnly,
		ActorText: `You turn to fight {target}!`,
	})
	events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
	return true
}

// releaseAttackers ends the fight of every body, player or NPC, whose aggro
// targets the victim, and turns each to its next opponent at once. Bodies
// in the victim's room go first, in the room's order, then any elsewhere
// (a shooter, a body waiting on the victim) in actor order.
func releaseAttackers(victim *actors.Actor) {
	seen := map[actors.ActorId]bool{}
	order := []*actors.Actor{}
	if r := rooms.LoadRoom(victim.Character.RoomId); r != nil {
		for _, id := range r.GetBodies() {
			if b := actors.Get(id); b != nil {
				seen[b.Id] = true
				order = append(order, b)
			}
		}
	}
	for _, b := range actors.All() {
		if !seen[b.Id] {
			order = append(order, b)
		}
	}
	for _, b := range order {
		if b.Id == victim.Id || b.Character == nil || b.Character.Aggro == nil || b.Character.Aggro.Target != victim.Id {
			continue
		}
		if Switch(b) {
			continue
		}
		b.Character.EndAggro()
		events.AddToQueue(events.AggroChanged{ActorId: b.Id, RoomId: b.Character.RoomId})
	}
}
```

  (The line is a `WorldEvent` to the actor alone, so `{target}` resolves per reader: an `Interface`
  message is written unresolved; tried, it printed `{target}`.) `NextOpponent` has no `RoomId` test:
  its candidates are `r.GetBodies()` of a's room, and `rooms.MoveToRoom` keeps a body's `RoomId` and
  its room's lists in step (plan review round 5; its unit test moves the rat with `MoveToRoom`, so
  the "another room" assertion holds consistent state). In `internal/death/death.go`, `Die`'s
  `clearKillersAggro(victim)` (`:72`) becomes `releaseAttackers(victim)`, and `clearKillersAggro` with
  its comment (`:264-275`) and the blank line after are deleted. A victim that revives returns from
  `Die` before `releaseAttackers` (`death.go:39-42`, unchanged), so its attackers keep fighting it:
  accepted (it got up, and the fight goes on; plan review round 2). In
  `internal/hooks/death_engine_test.go`, the comment at `:70-72`, `... The death ends the fight of every
  credited account still targeting` / `// the body (clearKillersAggro), credits the kill and grants the
  XP (award).`, becomes `... The death ends the fight of every body still targeting` / `// the body
  (releaseAttackers), credits the kill and grants the XP (award).` (the test's assertions still hold:
  run).

- [ ] **Step 4: The round's turn order is stable (Q9).** `fighters()`
  (`NewRound_DoCombat.go:46-62`) becomes:

```go
// fighters lists every actor with a body that may take a combat turn, in
// death.SortBodies order (players by UserId, then NPCs by MobId, then body
// ID). The ruleset puts them in turn order.
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
	return death.SortBodies(out)
}
```

  (one sort, shared with `NextOpponent`'s tie-break, Step 3). Without it, `pvp`, `death_pvp`, `mvm` and `hidden_attacker`
  recorded a different file on 8 of 8 reruns after `-update` (two accounts, or two NPCs, both
  fighting: map order). With it, every golden was stable over 6 full runs.

- [ ] **Step 4a: The wait limit's config field (ruling 12), before Step 5 reads it.**
  - **Config home**: `GamePlay`, beside `AlignmentDecayRounds` and `EliteXPBonus`
    (`internal/configs/config.gameplay.go:3-27`): an engine game-play value, read through
    `configs.GetGamePlayConfig()` (which validates first, `config.gameplay.go:203-211`). It does not go
    in `GamePlay.Combat` (`CombatConfig`, `:30-54`): that block holds the reference ruleset's combat
    bounds, and the wait limit is the engine round's. After `EliteXPBonus` (`:26`):

```go
	// OutOfReachRounds is how many rounds in a row a fight waits on a target
	// out of reach before the body loses track of it (0 or less = 5).
	OutOfReachRounds ConfigInt `yaml:"OutOfReachRounds"`
```

    and in `Validate`, after the `EliteXPBonus` default (`:126-128`):

```go

	if g.OutOfReachRounds <= 0 {
		g.OutOfReachRounds = 5
	}
```

    Create `internal/configs/config.gameplay_test.go`:

```go
package configs

import "testing"

// Phase 5 slice 4.1 (ruling 12): the out-of-reach wait limit defaults to 5
// rounds when it is 0 or less.
func TestOutOfReachRoundsDefaultsToFive(t *testing.T) {
	for set, want := range map[ConfigInt]ConfigInt{0: 5, -2: 5, 1: 1, 3: 3} {
		g := GamePlay{OutOfReachRounds: set}
		g.Validate()
		if g.OutOfReachRounds != want {
			t.Errorf("OutOfReachRounds %d validates to %d, want %d", set, g.OutOfReachRounds, want)
		}
	}
}
```

    Mutation (run, killed): `<= 0` to `< 0`. The config keys in the data and admin files are Step 5d.

- [ ] **Step 5: The target lookup, retaliation and the end of a swing.** Do Step 4a first (this step
  reads `GamePlay.OutOfReachRounds`). In `NewRound_DoCombat.go`'s imports add `"sort"` (`shootingExit`
  below sorts exit names; after Step 4 nothing else in the file uses it). In `combatTurn`
  (`:102-174` at `1c8e607f`), everything from the no-aggro check through the `"Your rage subsides."`
  block, that is

```go
	if a == nil || a.Character == nil || a.Character.Aggro == nil {
		return nil
	}

	// If has a buff that prevents combat, skip the actor
	if a.Character.HasBuffFlag("no-combat") {
		return nil
	}
```

  down to the end of

```go
	if rules.Get().Condition(d) != rules.Up {
		usercommands.Tell(a, `Your rage subsides.`)
		endAggro(a)
		return affected
	}
```

  becomes (every body with an aggro settles its target every round before `no-combat` or an NPC's
  combat command may take its turn; only the swing is skipped: spec, plan review round 3):

```go
	if a == nil || a.Character == nil || a.Character.Aggro == nil {
		return nil
	}

	aRoom := rooms.LoadRoom(a.Character.RoomId)
	if aRoom == nil {
		endAggro(a)
		return nil
	}

	// A buff that prevents combat takes the turn (the swing below, or a flee).
	noCombat := a.Character.HasBuffFlag("no-combat")

	if a.Character.Aggro.Type == characters.Flee {
		if noCombat {
			return nil
		}
		a.Character.CancelBuffsWithFlag("cancel-on-combat")
		flee(a, aRoom)
		return nil
	}

	/**************************
	*
	* START HANDLING PHYSICAL COMBAT
	*
	**************************/

	// Every body with an aggro settles its target every round, before
	// anything may take its turn (no-combat, an NPC's combat command): the
	// lookup runs, a switch or a lost track happens, the wait count advances
	// (Phase 5 slice 4.1, ruling 12). Only the swing is skipped.
	d := settleTarget(a, aRoom)
	if d == nil {
		return []ids.ID{a.Id}
	}

	if noCombat {
		return nil
	}

	// Disable any buffs that are cancelled by combat
	a.Character.CancelBuffsWithFlag("cancel-on-combat")

	// An NPC's combat commands (behaviour data on its template) may take its
	// turn instead of the attack.
	if mob := mobs.GetInstance(a.Id); mob != nil && a.Character.Aggro.Type == characters.DefaultAttack {
		if combatCommand(mob) {
			return nil
		}
	}

	affected = append(affected, a.Id)

	defRoom := rooms.LoadRoom(d.Character.RoomId)
	if defRoom == nil {
		endAggro(a)
		return affected
	}

	d.Character.CancelBuffsWithFlag("cancel-on-combat")
```

  (The flee path keeps today's order: `no-combat` skips it, otherwise the attacker's
  `cancel-on-combat` buffs end and it flees. For a swing the body's own `cancel-on-combat` buffs now
  end after the lookup instead of before it; the lookup reads perception of others, not the body's own
  buffs. Today's affected list held the body whenever it reached the lookup; it still does, and a
  `no-combat` body is still not in it. Run on the scratch copy: every golden byte-identical with this
  reorder. A `no-combat` body whose lookup switches reads "You turn to fight <name>!" and does not
  swing: accepted, plan review round 4.)

  The NPC block's comment (`// An NPC that is attacked and not fighting stays put; attacked from` /
  `// another room, it heads that way.`) ends `it heads that way (and retaliates below).` The block
  stays before retaliation: it tests `d.Character.Aggro == nil`. Then the end of the swing
  (at `1c8e607f` `:269-275`)

```go
	if rules.Get().Condition(a) != rules.Up || rules.Get().Condition(d) != rules.Up {
		endAggro(a)
		endAggro(d)
	} else {
		a.Character.SetAggro(d.Id, characters.DefaultAttack)
		events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
	}
```

  becomes

```go
	// A defender whose aggro targets its attacker is still in that fight:
	// the attack, hit or miss, resets its wait count (plan review round 6).
	if d.Character.Aggro != nil && d.Character.Aggro.Target == a.Id {
		d.Character.Aggro.OutOfReach = 0
	}
	retaliate(d, a)

	// A defender taken out of Up leaves its own fight. The attacker keeps
	// its aggro: the death that follows (death.Die) turns it to its next
	// opponent, or ends it; under a ruleset with a downed state its next
	// turn does (settleTarget).
	if rules.Get().Condition(d) != rules.Up {
		endAggro(d)
	} else if exitName := a.Character.Aggro.ExitName; exitName != `` && d.Character.RoomId != a.Character.RoomId {
		// A shot through an exit re-arms through the same exit, so the next
		// turn's lookup finds the aggro unchanged.
		a.Character.SetAggroRemote(exitName, d.Id, characters.Shooting)
		events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
	} else {
		a.Character.SetAggro(d.Id, characters.DefaultAttack)
		events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
	}
```

  (An attacker's own attack never takes it out of Up, so the old `Condition(a)` half had no case.
  Re-arming a shot with plain `SetAggro` would drop the exit, and the next turn's step 3 would put it
  back with a second `AggroChanged` each round: plan review round 4.) The reset before `retaliate`
  is the spec's lookup step 5 "and so does being attacked by the target" (Decisions row "Being
  attacked by its target resets a body's wait count"): this point is reached after every attack,
  hit or miss, the same place retaliation is decided.
  Add before `// combatCommand may run one of an NPC's combat commands`:

```go
// settleTarget finds the target a swings at this turn (Phase 5 slice 4.1,
// rulings 7-9), or returns nil when the turn ends here:
//  1. the target no longer exists: switch to the next opponent, or end
//     ("Your target can't be found.");
//  2. the target is not up: switch, or end ("Your rage subsides.");
//  3. the target is in another room, a wields a shooting weapon and an
//     unlocked exit leads there: shoot through that exit (a target a can
//     shoot is reachable, so it is kept even if someone here attacks a);
//  4. the target is in another room, or here but a cannot perceive it, and
//     someone here whom a perceives attacks a: switch;
//  5. otherwise the aggro waits, telling a once "Your target is out of
//     reach."; a swings again once they share a room. After
//     GamePlay.OutOfReachRounds such turns in a row (5 by default) a loses
//     track of the target and the aggro ends (ruling 12). A target in reach
//     that a cannot perceive counts the same way (inReach). Being attacked
//     by the target resets the count but not the told mark, so the line
//     comes once per wait.
//
// A target it would attack (in the room, or through an exit) must still be
// one the ruleset lets it attack (rules.CanAttack: the fight may have moved
// into a room where PvP is off); if not, it switches as in step 4 or ends,
// told the refusal. The switch picks only opponents the ruleset lets it
// attack (death.NextOpponent). A switch re-arms the aggro through SetAggro,
// as attack <name> does, and the turn goes on to the new opponent: it swings
// this turn, unless its weapon's delay makes it wait first.
func settleTarget(a *actors.Actor, aRoom *rooms.Room) *actors.Actor {
	d := actors.Get(a.Character.Aggro.Target)
	if d == nil || d.Character == nil {
		return switchOrEnd(a, `Your target can't be found.`)
	}
	if rules.Get().Condition(d) != rules.Up {
		return switchOrEnd(a, `Your rage subsides.`)
	}
	if d.Character.RoomId == a.Character.RoomId {
		if err := rules.Get().CanAttack(a, d); err != nil {
			return switchOrEnd(a, err.Error())
		}
		// Step 4 for a target here that a cannot perceive: a perceived
		// attacker in the room comes first.
		if !rooms.Perceives(a.Character, d.Character) && death.Switch(a) {
			return actors.Get(a.Character.Aggro.Target)
		}
		return inReach(a, d)
	}
	if exitName := shootingExit(a, aRoom, d.Character.RoomId); exitName != `` {
		if err := rules.Get().CanAttack(a, d); err != nil {
			return switchOrEnd(a, err.Error())
		}
		if a.Character.Aggro.Type != characters.Shooting || a.Character.Aggro.ExitName != exitName {
			a.Character.Aggro.Type = characters.Shooting
			a.Character.Aggro.ExitName = exitName
			events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
		}
		return inReach(a, d)
	}
	if death.NextOpponent(a) != nil {
		death.Switch(a)
		return actors.Get(a.Character.Aggro.Target)
	}
	a.Character.Aggro.OutOfReach++
	if !a.Character.Aggro.OutOfReachTold {
		a.Character.Aggro.OutOfReachTold = true
		usercommands.Tell(a, `Your target is out of reach.`)
	}
	if a.Character.Aggro.OutOfReach >= outOfReachRounds() {
		loseTrack(a, d)
	}
	return nil
}

// inReach settles a target a can reach (in its room, or through an exit).
// One a perceives resets the wait count and the told mark (OutOfReachTold). One a cannot perceive counts
// toward the wait limit like a target out of reach (ruling 12; the turn
// then says "You can't seem to find your target."), and a loses track of it
// at the limit.
func inReach(a, d *actors.Actor) *actors.Actor {
	if rooms.Perceives(a.Character, d.Character) {
		a.Character.Aggro.OutOfReach = 0
		a.Character.Aggro.OutOfReachTold = false
		return d
	}
	a.Character.Aggro.OutOfReach++
	if a.Character.Aggro.OutOfReach >= outOfReachRounds() {
		loseTrack(a, d)
		return nil
	}
	return d
}

// outOfReachRounds is how many consecutive turns an aggro may wait on a
// target out of reach (ruling 12): GamePlay.OutOfReachRounds, which
// GamePlay.Validate makes 5 when it is 0 or less (GetGamePlayConfig always
// returns a validated config).
func outOfReachRounds() int {
	return int(configs.GetGamePlayConfig().OutOfReachRounds)
}

// loseTrack ends a's wait on d (ruling 12): a is told it lost track of d,
// and those in its room who perceive it see it give up.
func loseTrack(a, d *actors.Actor) {
	messaging.Send(perception.Message{
		Kind:      perception.WorldEvent,
		Actor:     a.Id,
		Target:    d.Id,
		Audience:  perception.ActorOnly,
		ActorText: `You lose track of {target}.`,
	})
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        a.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ObserverText: `{actor} mumbles about losing their quarry.`,
		Exclude:      rooms.NotPerceiving(a.Character),
	})
	endAggro(a)
}

// switchOrEnd turns a to its next opponent and returns it, or ends a's
// aggro with line and returns nil.
func switchOrEnd(a *actors.Actor, line string) *actors.Actor {
	if death.Switch(a) {
		return actors.Get(a.Character.Aggro.Target)
	}
	usercommands.Tell(a, line)
	endAggro(a)
	return nil
}

// shootingExit is the exit from aRoom, unlocked and leading to roomId, that a
// can shoot through: none unless a wields a shooting weapon. Exits are looked
// up as the shoot command does (room.FindExitByName for where an exit leads,
// room.GetExitInfo for its lock), so temporary and mutator exits and mutator
// locks count. The aggro's own exit is preferred; otherwise the first by
// name of the room's exits, temporary exits and active mutators' exits.
func shootingExit(a *actors.Actor, aRoom *rooms.Room, roomId int) string {
	if a.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		return ``
	}
	names := []string{}
	if a.Character.Aggro != nil && a.Character.Aggro.ExitName != `` {
		names = append(names, a.Character.Aggro.ExitName)
	}
	all := []string{}
	for name := range aRoom.Exits {
		all = append(all, name)
	}
	for name := range aRoom.ExitsTemp {
		all = append(all, name)
	}
	for mut := range aRoom.ActiveMutators {
		if spec := mut.GetSpec(); spec != nil {
			for name := range spec.Exits {
				all = append(all, name)
			}
		}
	}
	sort.Strings(all)
	names = append(names, all...)
	for _, name := range names {
		exitName, toRoom := aRoom.FindExitByName(name)
		if toRoom != roomId {
			continue
		}
		if exitInfo, ok := aRoom.GetExitInfo(exitName); ok && exitInfo.Lock.IsLocked() {
			continue
		}
		return exitName // the name shoot <direction> would store
	}
	return ``
}

// retaliate gives a defender that is not fighting an aggro on its attacker
// (Phase 5 slice 4.1, rulings 6, 8 and 9), after every attack, hit or miss,
// when it is up and the ruleset lets it attack back: a plain attack when they
// share a room; through an exit when the attacker is in another room and the
// defender can shoot there; otherwise an aggro that waits. An attacker the
// defender cannot perceive is taken all the same (its turn then says "You
// can't seem to find your target.").
func retaliate(d, a *actors.Actor) {
	if d.Character.Aggro != nil || rules.Get().Condition(d) != rules.Up {
		return
	}
	if rules.Get().CanAttack(d, a) != nil {
		return
	}
	if a.Character.RoomId == d.Character.RoomId {
		d.Character.SetAggro(a.Id, characters.DefaultAttack)
	} else if dRoom := rooms.LoadRoom(d.Character.RoomId); dRoom != nil && shootingExit(d, dRoom, a.Character.RoomId) != `` {
		d.Character.SetAggroRemote(shootingExit(d, dRoom, a.Character.RoomId), a.Id, characters.Shooting)
	} else {
		d.Character.SetAggro(a.Id, characters.DefaultAttack)
	}
	events.AddToQueue(events.AggroChanged{ActorId: d.Id, RoomId: d.Character.RoomId})
}
```

  This carries Q11 (a switch inside a turn goes on to the new opponent: it swings, unless the
  weapon's delay, which `SetAggro` re-arms, makes it wait first), Q12 (shoot before switch)
  and Q13 (the waiting line). `go build ./...` and `gofmt -l internal` print nothing.

- [ ] **Step 5a: NPCs attacking players are not PvP (Q7; ruling 10).** In
  `internal/rules/dikubasic/ruleset.go`, `CanAttack`'s comment and its first checks (`:85-96`)

```go
// CanAttack allows any NPC target. An account target needs PvP allowed in
// the attacker's room and, when it differs, the defender's room (spec
// ruling 1: no shooting out of or into a safe room).
func (Ruleset) CanAttack(attacker, defender *actors.Actor) error {
	if accountOf(defender) == nil {
		return nil
	}
	// CanPvp checks only the attacker's room against the recovery room; a
	// safe room protects whoever stands in it, its private copies included.
	if rooms.IsDeathRecoveryRoom(defender.Character.RoomId) {
		return errors.New(`Fighting is not allowed here.`)
	}
```

  become

```go
// CanAttack refuses any attacker standing in the recovery room. It allows
// any other NPC target. An account target in the recovery room
// (or a private copy of it) is safe from every attacker. The PvP rules apply
// only when both sides are accounts (Phase 5 slice 4.1: an NPC fighting a
// player is not PvP): PvP allowed in the attacker's room and, when it
// differs, the defender's room (spec ruling 1: no shooting out of or into a
// safe room).
func (Ruleset) CanAttack(attacker, defender *actors.Actor) error {
	// The recovery room (or a private copy of it) is safe in both
	// directions: nobody fights out of it either (plan review round 4).
	if rooms.IsDeathRecoveryRoom(attacker.Character.RoomId) {
		return errors.New(`Fighting is not allowed here.`)
	}
	if accountOf(defender) == nil {
		return nil
	}
	// CanPvp checks only the attacker's room against the recovery room; a
	// safe room protects whoever stands in it, its private copies included.
	if rooms.IsDeathRecoveryRoom(defender.Character.RoomId) {
		return errors.New(`Fighting is not allowed here.`)
	}
	if accountOf(attacker) == nil {
		return nil
	}
```

  (the rest, from `r := rooms.LoadRoom(attacker.Character.RoomId)`, stays). The attacker check is the
  spec's "safe in both directions" row (plan review round 4): a player respawned there cannot shoot
  out through its portal at an NPC that could never fight back. The smoke death path does nothing in
  the recovery room but `look` and `portal` (`script.go:229-230`, and the fight step's own recovery,
  Step 12), neither of which asks `CanAttack`. Append to `ruleset_test.go`:

```go

// Phase 5 slice 4.1 (plan Q7): an NPC attacking a player is not PvP. Under
// the shipped settings (limited, minimum level 15), in a room that is not a
// PvP room, an NPC may attack a level 1 player; a player may not; and the
// recovery room, and a private copy of it, is safe from both.
func TestCanAttackAppliesPvPOnlyBetweenPlayers(t *testing.T) {
	const here, recovery = 94011, 94012
	testRoom(t, here, "rulesetzone")
	testRoom(t, recovery, "rulesetzone")
	copyId := recoveryCopy(t, recovery, "rulesetzone")
	setConfig(t,
		map[string]any{"GamePlay.PVP.Enabled": "limited", "GamePlay.PVP.MinimumLevel": 15},
		map[string]any{
			"GamePlay.PVP.Enabled":      string(configs.GetPVPConfig().Enabled),
			"GamePlay.PVP.MinimumLevel": int(configs.GetPVPConfig().MinimumLevel),
		})
	target := player(t, 9631)
	target.Character.RoomId = here
	require.Equal(t, 1, target.Character.Level)

	n := npc(t, 0, 0)
	n.Character.RoomId = here
	assert.NoError(t, New().CanAttack(n, target), "an NPC may attack a level 1 player outside a PvP room")

	p := player(t, 9632)
	p.Character.RoomId = here
	assert.EqualError(t, New().CanAttack(p, target), "Players must be at least level 15 to PVP.", "player against player: the PvP rules, unchanged")

	for _, roomId := range []int{recovery, copyId} {
		target.Character.RoomId = roomId // the NPC stays outside, in room here
		assert.EqualError(t, New().CanAttack(n, target), "Fighting is not allowed here.", "room %d: safe from an NPC too", roomId)
	}
}

// Phase 5 slice 4.1 (plan review round 4): the recovery room is safe in
// both directions; a body standing in it, or in a private copy of it,
// attacks nobody (for example an NPC through the room's portal).
func TestCanAttackRefusesAnAttackerInTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94013, 94014
	testRoom(t, here, "rulesetzone")
	testRoom(t, recovery, "rulesetzone")
	copyId := recoveryCopy(t, recovery, "rulesetzone")
	n := npc(t, 0, 0)
	n.Character.RoomId = here
	p := player(t, 9633)
	for _, roomId := range []int{recovery, copyId} {
		p.Character.RoomId = roomId
		assert.EqualError(t, New().CanAttack(p, n), "Fighting is not allowed here.", "room %d", roomId)
	}
	p.Character.RoomId = here
	assert.NoError(t, New().CanAttack(p, n), "control: out of it, the NPC may be attacked")
}
```

  Mutation for the new test (run, killed): the attacker check to `if false {` fails it ("room
  94014"). (`TestCanAttackAppliesPvPOnlyBetweenPlayers` keeps its NPC outside the recovery room, so the
  attacker check does not mask the target check it tests.)

  Mutations (run, both killed): delete the new `if accountOf(attacker) == nil { return nil }` (fails
  "an NPC may attack a level 1 player outside a PvP room"); move that check above the recovery-room
  check (fails "room 94012: safe from an NPC too" and the same for the copy). The existing
  `TestCanAttackRefusesATargetInTheRecoveryRoom` and `...InACopyOfTheRecoveryRoom` (player attackers)
  still pass.

  `internal/usercommands/shoot_test.go:113-123`, `TestNPCShootingAPlayerRefusedWhenPvPDisabled`,
  fails with this change (run: `a refused NPC shot starts no fight`, an aggro was set). By ruling 10
  it becomes:

```go
// An NPC shooting a player is not PvP (Phase 5 slice 4.1, plan Q7): it
// shoots even with PvP disabled.
func TestNPCShootingAPlayerAllowedWhenPvPDisabled(t *testing.T) {
	npc, _, here := goSetup(t) // "Rat" is in this room
	giveBow(npc)
	target := addFoundPlayer(t, 9203, "Target", rooms.LoadRoom(testNextRoomId))
	setPvP(t, configs.PVPDisabled)

	_, err := Shoot("Target north", npc, here, 0)
	require.NoError(t, err)
	require.NotNil(t, npc.Character.Aggro, "an NPC shot is not PvP")
	assert.Equal(t, target.Id, npc.Character.Aggro.Target)
}
```

  The following `TestNPCShootingAPlayerAllowedWhenPvPEnabled` still passes and is now the same case
  with PvP enabled; its comment (`shoot_test.go:125-126`), `// Control for the test above: with PvP
  enabled the same NPC shot sets the` / `// aggro, so a refusal there is the PvP rule, not a failed
  lookup.`, goes stale and becomes `// The test above with PvP enabled: the NPC's shot sets the aggro
  either way` / `// (an NPC shooting a player is not PvP, ruling 10).`

- [ ] **Step 5b: A waiting aggro does not block walking, quitting or suicide (Q8, Q17).** In
  `internal/usercommands/go.go`, `inFight` (`:355-375`), the first check

```go
	if actor.Character.Aggro != nil {
		return true
	}
```

  becomes

```go
	// Its own aggro is a fight only while its target is here: a body whose
	// aggro waits on a target elsewhere, or that shoots into another room,
	// may walk (Phase 5 slice 4.1, ruling 8).
	if aggro := actor.Character.Aggro; aggro != nil {
		if target := actors.Get(aggro.Target); target != nil && target.Character != nil && target.Character.RoomId == actor.Character.RoomId {
			return true
		}
	}
```

  and the function's comment (`:355-357`) gains: `Its own aggro counts only while its target is in its
  room (slice 4.1).` `inFight` also gates `quit` (`quit.go:18`) and `suicide` (`suicide.go:17`), so
  those follow (Q17). Append to `go_test.go`:

```go

// Phase 5 slice 4.1 (plan Q8): a body's own aggro is a fight only while its
// target is in its room; one whose target is elsewhere may walk.
func TestBodyWhoseTargetIsInAnotherRoomMayGo(t *testing.T) {
	npc, player, room := goSetup(t)
	player.Character.SetAggro(npc.Id, characters.DefaultAttack)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))

	handled, err := Go("north", player, room, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, testNextRoomId, player.Character.RoomId, "her target is not here: she may walk")
}
```

  Mutations (run, both killed): the room test to `true {` (fails "her target is not here: she may
  walk"); to `false {` (fails the existing `TestNPCWithAggroTargetIsRefusedGo`). The existing
  `TestPlayerTargetedByNPCInRoomIsRefusedGo` (attacked in the room) still passes.

  Append to `internal/usercommands/quit_test.go` (and add
  `"github.com/GoMudEngine/GoMud/internal/rooms"` to its imports):

```go

// Phase 5 slice 4.1 (plan Q17): a player whose only fight is with a target
// in another room is not in a fight: she may quit.
func TestPlayerWhoseTargetIsInAnotherRoomMayQuit(t *testing.T) {
	npc, player, room := goSetup(t)
	player.Character.SetAggro(npc.Id, characters.DefaultAttack)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	reading(t, player.Controller().Account())
	quitterId := player.Controller().Account().UserId

	reads := captureReads(t, func() {
		handled, err := Quit("", player, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})

	quitter := strings.Join(reads[quitterId], "")
	assert.Contains(t, quitter, "You sit down and begin your meditation.", "her target is not here: she may quit")
	assert.NotContains(t, quitter, "You're too busy to quit right now!")
}
```

  Mutation (run, killed): `inFight`'s room test to `true {` fails it with
  `"You're too busy to quit right now!\n" does not contain "You sit down and begin your meditation."`.
  The existing `TestPlayerTargetedByNPCInRoomIsRefusedQuit` (attacked in the room) still passes.

- [ ] **Step 5c: A failed flee keeps its place (Q15).** In `internal/usercommands/flee.go` (`:15-22`):

```go
		// Keep the target fled from, and when that fight began: a blocked or
		// exitless flee resumes it in its place (Phase 5 slice 4.1).
		fleeFrom := ids.ID{}
		var since uint64
		if actor.Character.Aggro != nil {
			fleeFrom = actor.Character.Aggro.Target
			since = actor.Character.Aggro.Since
		}

		actor.Character.Aggro = &characters.Aggro{Type: characters.Flee, FleeFrom: fleeFrom, Since: since}
```

  In `NewRound_DoCombat.go`, `flee` reads `since := a.Character.Aggro.Since` after
  `fleeFrom := a.Character.Aggro.FleeFrom`; both `resumeAfterFlee(a, fleeFrom)` calls become
  `resumeAfterFlee(a, fleeFrom, since)`; and `resumeAfterFlee` becomes:

```go
func resumeAfterFlee(a *actors.Actor, fleeFrom ids.ID, since uint64) {
	if fleeFrom.IsZero() {
		a.Character.EndAggro()
	} else {
		a.Character.SetAggro(fleeFrom, characters.DefaultAttack)
		a.Character.Aggro.Since = since // the fight resumes in its place
	}
	events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
}
```

  The test is in Step 8 (`TestFailedFleeKeepsSince`).

- [ ] **Step 5d: The wait limit's config keys elsewhere, and `loseLostQuarry` deleted (ruling 12).**
  (The `GamePlay.OutOfReachRounds` field, its default and its test are Step 4a, before Step 5.)
  - **Where the keys are listed**: `_datafiles/config.yaml`, after `EliteXPBonus: 10` (`:286`; the
    file is CRLF in a Windows checkout, keep its line ends):

```yaml
  # - OutOfReachRounds -
  #   How many rounds in a row a fight waits on a target out of reach (in
  #   another room, not shootable) before the fighter loses track of it.
  #   0 or less means 5.
  OutOfReachRounds: 5
```

    `_datafiles/html/admin/config.html`, after the `'GamePlay.EliteXPBonus': ...` description
    (`:275`):
    `        'GamePlay.OutOfReachRounds': 'Rounds in a row a fight waits on a target out of reach (another room, not shootable) before the fighter loses track of it. 0 or less = 5.',`.
    `_datafiles/html/admin/config-wizard.html`: after the `'GamePlay.EliteXPBonus': { ... },` entry
    (`:1103-1109`) a new entry

```js
    'GamePlay.OutOfReachRounds': {
        label: 'Out-of-Reach Rounds',
        description: 'How many rounds in a row a fight waits on a target that is out of reach (in another room and not shootable) before the fighter loses track of it.',
        implications: 'Higher values keep fighters waiting longer for a fleeing target to come back; lower values end such fights sooner. 0 or less means 5.',
        inputType: 'number',
        min: 0, max: 100, step: 1,
    },
```

    and `'GamePlay.OutOfReachRounds'` joins the Gameplay section's `fields` list (`:1474`), after
    `'GamePlay.EliteXPBonus'`. No Go test enumerates the config keys (grep `config-wizard\|config.html`
    and the key names over `*_test.go`: none).
  - **`loseLostQuarry` deleted**: in `internal/behaviour/tick.go`, `Tick`'s aggro branch (`:53-58`)

```go
	// If they are doing some sort of combat thing,
	// Don't do idle actions
	if mob.Character.Aggro != nil {
		loseLostQuarry(mob)
		return
	}
```

    becomes

```go
	// If they are doing some sort of combat thing,
	// Don't do idle actions. A fight whose target is out of reach ends in the
	// combat round (ruling 12), for every body.
	if mob.Character.Aggro != nil {
		return
	}
```

    and `loseLostQuarry` with its comment (`:71-84`) and the blank line after are deleted; the
    `accounts` and `events` imports go (nothing else in the file uses them; built). No test names
    `loseLostQuarry` or its line (grep `quarry` over the repo: `tick.go` only). Append to
    `internal/behaviour/tick_test.go` (and add `"github.com/GoMudEngine/GoMud/internal/ids"` to its
    imports):

```go

// Phase 5 slice 4.1 (ruling 12): an NPC's fight is the combat round's to
// end, by one wait limit for every body; its behaviour tick leaves the aggro
// alone even when the target is nowhere to be found.
func TestTickLeavesAFightToTheCombatRound(t *testing.T) {
	testRoom(t, 9206, testZone, nil)
	mob := busyMob(t, 9206, 9206, 3)
	mob.Character.Aggro = &characters.Aggro{Target: ids.New(ids.NoCell)} // nobody online has this body

	assert.Empty(t, tickRounds(t, mob, 5), "no idle, no emote")
	require.NotNil(t, mob.Character.Aggro, "the tick does not end the fight")
}
```

    Mutation (run, killed): the aggro branch also sets `mob.Character.Aggro = nil` (as
    `loseLostQuarry` did for an absent player) fails first on "no idle, no emote" (the later ticks,
    with no aggro, emote `hums.` four times), then "the tick does not end the fight". The existing
    `TestTickWhileFightingDoesNotIdle` fails under it too (`Should be empty, but was [emote hums. ...]`),
    for the same reason; without the mutation both pass.
  - **What `loseLostQuarry`'s lines become** (spec Decisions row "A target that leaves the game ends
    the aggro with today's ..."): a target that left the game is settled by step 1, which switches or
    ends with today's "Your target can't be found." and sends no lose-track line, for every body (an
    NPC whose player target was offline, or whose target had vanished, emoted before:
    `tick.go:74-84` at `1c8e607f`). The lose-track lines come only from step 5 and
    `inReach` (`loseTrack`, Step 5), and the observer line there is a plain `WorldEvent` room line,
    `{actor} mumbles about losing their quarry.`, without the emote colour `loseLostQuarry`'s `emote`
    command gave it.
  - **The NPC walk toward a remote attacker** is not cut short: the NPC queues `go <exit>` in the
    attacker's turn, the command runs before the NPC's next turn, and the next turn finds the target
    in its room (count back to 0); in `shot_from_another_room` the rat's count never passed 1. An NPC
    with a bow shot from another room does both: the walk block (it tests `d.Character.Aggro == nil`)
    runs first and queues `go <exit>` and an angry command, as today, then `retaliate` gives it a
    shooting aggro through the exit, so it may shoot back once before the walk lands it in the
    shooter's room. Intended: the walk is today's behaviour (spec "What stays as today"), the shot is
    ruling 9.

- [ ] **Step 6: The golden harness advances the round count.** In `combat_golden_test.go`
  `newCombatWorld`, before `util.SeedRandForTest(seed)` (`:175`):

```go
	// The round count starts at 1000 for every fixture and each recorded
	// round advances it, as the game loop does (an aggro records the round
	// it began, Aggro.Since).
	prevRound := util.GetRoundCount()
	util.SetRoundCount(1000)
	t.Cleanup(func() { util.SetRoundCount(prevRound) })
```

  and in `rounds`, `DoCombat(events.NewRound{})` becomes
  `DoCombat(events.NewRound{RoundNumber: util.IncrementRoundCount()})`. Applied alone, before Step 5,
  this left every golden byte-identical (run). Fixture comments that the new behaviour falsifies
  (line numbers at `1c8e607f`, where Tasks 1-6 leave these comments; the six-line round-count block
  above them in this step moves each down by 6): `TestCombatGoldenPvP` (`:603-604`)
  `// A player attacks a player, who fights back; the` / `// defender's shield blocks some of the
  damage.`; `TestCombatGoldenFleeSuccess` (`:662-663`) `// A player's flee succeeds: they leave through
  the only exit and the NPC` / `// that fought them waits on its target in the room they left.`;
  `TestCombatGoldenShootThroughExit` (`:677-678`) `// A player shoots an NPC in the next room through
  an exit; a player in that` / `// room watches. The shooter keeps shooting; the NPC, with no bow,
  waits.`; `TestCombatGoldenPvMToDeath` (`:589`) `// A player kills an NPC that fights back.`

- [ ] **Step 7: `TestCombatGoldenDeathNPC` keeps its NPC death (Q10).** With retaliation the
  elite rat (club) kills Alice in round 5 and survives at 1/12, so `require.Nil(t, mobs.GetInstance(rat...))`
  fails (run). Recommended fix, run on the scratch copy: Alice's health 10 becomes 30
  (`w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 30)` in that test only); the rat then dies in
  round 6 with Alice at 20/30, its drops in its corpse.

- [ ] **Step 8: The unit tests.** Create `internal/hooks/retaliation_test.go`:

```go
package hooks

import (
	"errors"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/gamelock"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mutators"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// told is what userId read since the last call.
func (w *combatWorld) told(userId int) []string {
	events.ProcessEvents()
	out := []string{}
	for _, e := range w.rec.take() {
		if e.userId == userId {
			out = append(out, strings.TrimSuffix(e.text, "\n"))
		}
	}
	return out
}

// aggroOn gives c an aggro on target that began on round since.
func aggroOn(c *characters.Character, target ids.ID, since uint64) {
	c.SetAggro(target, characters.DefaultAttack)
	c.Aggro.Since = since
}

// Phase 5 slice 4.1, ruling 7: the next opponent is the one in the room,
// up and perceived, whose aggro on the body began earliest; ties go to turn
// order (accounts first, then NPCs, each in room order).
func TestNextOpponentIsTheLongestAttacker(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	rat := w.addNPC(93205, "rat", 10, goldenRoomA, 5, 5, 5, 30)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30) // last in the room, first in turn order
	a := actors.Get(alice.Character.Id)
	next := func() string {
		if n := death.NextOpponent(a); n != nil {
			return n.Character.Name
		}
		return "none"
	}

	assert.Equal(t, "none", next(), "nobody attacks her")

	aggroOn(&cat.Character, alice.Character.Id, 9)
	aggroOn(&rat.Character, alice.Character.Id, 5)
	assert.Equal(t, "rat", next(), "earliest Since, though the cat is first in turn order")

	rat.Character.Aggro.Since = 9
	assert.Equal(t, "cat", next(), "a tie goes to turn order")

	aggroOn(bob.Character, alice.Character.Id, 9)
	assert.Equal(t, "Bob", next(), "accounts come first in turn order")

	bob.Character.Health = 0
	assert.Equal(t, "cat", next(), "a body that is not up is skipped")

	w.hide(&cat.Character)
	assert.Equal(t, "rat", next(), "a body she cannot perceive is skipped")

	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomB))
	assert.Equal(t, "none", next(), "a body in another room is skipped")
}

// Phase 5 slice 4.1, rulings 7 and 9: settleTarget's five steps.
func TestSettleTargetSteps(t *testing.T) {
	setup := func(t *testing.T) (*combatWorld, *actors.Actor, *rooms.Room) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		return w, actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	}

	t.Run("1 gone, nobody else: ends with today's line", func(t *testing.T) {
		w, a, room := setup(t)
		a.Character.SetAggro(ids.New(ids.NoCell), characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Your target can't be found.`}, w.told(9301))
	})

	t.Run("1 gone, an attacker here: switches", func(t *testing.T) {
		w, a, room := setup(t)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.SetAggro(ids.New(ids.NoCell), characters.DefaultAttack)
		w.told(9301)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id, "the turn goes on to the new opponent")
		assert.Equal(t, cat.Character.Id, a.Character.Aggro.Target)
		assert.Equal(t, []string{`You turn to fight <ansi fg="mobname">cat</ansi>!`}, w.told(9301))
	})

	t.Run("2 not up, nobody else: rage subsides", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
		rat.Character.Health = 0
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Your rage subsides.`}, w.told(9301))
	})

	t.Run("3 in another room, a bow, an attacker here: keeps shooting", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id, "a target it can shoot is reachable: kept")
	})

	t.Run("4 in another room, an attacker here: switches", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id)
	})

	t.Run("3 in another room, a bow and an unlocked exit: shoots", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		events.ProcessEvents()
		w.rec.take()
		aggroChanged := func() int {
			events.ProcessEvents()
			n := 0
			for _, e := range w.rec.take() {
				if e.userId == 0 && strings.HasPrefix(e.text, "AggroChanged Alice ") {
					n++
				}
			}
			return n
		}
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id)
		assert.Equal(t, characters.Shooting, a.Character.Aggro.Type)
		assert.Equal(t, "north", a.Character.Aggro.ExitName)
		assert.Equal(t, 1, aggroChanged(), "turning to shoot queues one AggroChanged")
		require.NotNil(t, settleTarget(a, room))
		assert.Equal(t, 0, aggroChanged(), "already shooting through that exit: none")
	})

	t.Run("3 the exit is locked: waits", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		room.Exits["north"] = exit.RoomExit{RoomId: goldenRoomB, Lock: gamelock.Lock{Difficulty: 5}}
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		assert.Nil(t, settleTarget(a, room))
		require.NotNil(t, a.Character.Aggro)
	})

	t.Run("5 waits, says so once, swings when they share a room", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, settleTarget(a, room))
		assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "told once")
		require.NotNil(t, a.Character.Aggro, "the aggro waits")
		rat.Character.RoomId = goldenRoomA
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id)
	})
}

// Phase 5 slice 4.1, rulings 6, 8 and 9: a defender with no aggro takes one
// on its attacker after an attack, hit or miss.
func TestRetaliate(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	a := actors.Get(alice.Character.Id)

	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
	retaliate(actors.Get(rat.Character.Id), a)
	require.NotNil(t, rat.Character.Aggro, "same room: a plain attack")
	assert.Equal(t, alice.Character.Id, rat.Character.Aggro.Target)
	assert.Empty(t, rat.Character.Aggro.ExitName)

	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	aggroOn(&cat.Character, rat.Character.Id, 1)
	retaliate(actors.Get(cat.Character.Id), a)
	assert.Equal(t, rat.Character.Id, cat.Character.Aggro.Target, "a body already fighting keeps its target")

	archer := w.addNPC(93206, "archer", 1, goldenRoomB, 5, 5, 5, 30)
	archer.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	retaliate(actors.Get(archer.Character.Id), a)
	require.NotNil(t, archer.Character.Aggro)
	assert.Equal(t, characters.Shooting, archer.Character.Aggro.Type, "another room, a bow: shoots back")
	assert.Equal(t, "south", archer.Character.Aggro.ExitName)

	dog := w.addNPC(93203, "dog", 1, goldenRoomB, 5, 5, 5, 30)
	retaliate(actors.Get(dog.Character.Id), a)
	require.NotNil(t, dog.Character.Aggro, "another room, no bow: an aggro that waits")
	assert.Empty(t, dog.Character.Aggro.ExitName)

	down := w.addNPC(93207, "mouse", 10, goldenRoomA, 5, 5, 5, 30)
	down.Character.Health = 0
	retaliate(actors.Get(down.Character.Id), a)
	assert.Nil(t, down.Character.Aggro, "a body that is not up does not retaliate")

	t.Cleanup(rules.SetForTest(refusesAll{rules.Get()}))
	owl := w.addNPC(93208, "owl", 10, goldenRoomA, 5, 5, 5, 30)
	retaliate(actors.Get(owl.Character.Id), a)
	assert.Nil(t, owl.Character.Aggro, "the ruleset's CanAttack refuses")
}

// refusesAll is the active ruleset with every attack refused.
type refusesAll struct{ rules.Ruleset }

func (refusesAll) CanAttack(attacker, defender *actors.Actor) error {
	return errors.New("Fighting is not allowed here.")
}

// neverEscapes is the active ruleset with every flee blocked.
type neverEscapes struct{ rules.Ruleset }

func (neverEscapes) FleeEscapes(fleer, blocker *actors.Actor) bool { return false }

// Phase 5 slice 4.1 (plan Q15): a blocked flee resumes the fight with the
// Since it had, so the fleer keeps its place.
func TestFailedFleeKeepsSince(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	aggroOn(alice.Character, guard.Character.Id, 5)
	aggroOn(&guard.Character, alice.Character.Id, 5)
	t.Cleanup(rules.SetForTest(neverEscapes{rules.Get()}))

	_, err := usercommands.Flee("", a, rooms.LoadRoom(goldenRoomA), 0)
	require.NoError(t, err)
	require.Equal(t, characters.Flee, alice.Character.Aggro.Type)
	flee(a, rooms.LoadRoom(goldenRoomA))

	require.NotNil(t, alice.Character.Aggro)
	assert.Equal(t, guard.Character.Id, alice.Character.Aggro.Target, "the fight resumes")
	assert.Equal(t, uint64(5), alice.Character.Aggro.Since, "in its place")
}

// Phase 5 slice 4.1 (plan review round 2): before the turn attacks its
// target, in the room or through an exit, the ruleset must still allow it
// (the fight may have moved into a room where PvP is off). Refused, the body
// switches to an attacker in its room, or its aggro ends with the refusal.
func TestSettleTargetAsksCanAttackBeforeAttacking(t *testing.T) {
	setup := func(t *testing.T) (*combatWorld, *actors.Actor, *rooms.Room) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		t.Cleanup(rules.SetForTest(refusesAll{rules.Get()}))
		return w, actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	}

	t.Run("in the room, refused, nobody else: ends with the refusal", func(t *testing.T) {
		w, a, room := setup(t)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		a.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Fighting is not allowed here.`}, w.told(9301))
	})

	t.Run("through an exit, refused: ends with the refusal", func(t *testing.T) {
		w, a, room := setup(t)
		bob := w.addPlayer(9302, "Bob", goldenRoomB, 10, 10, 10, 30)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Fighting is not allowed here.`}, w.told(9301))
	})

	t.Run("in the room, refused, an attacker it may fight here: switches", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		t.Cleanup(rules.SetForTest(refuses{Ruleset: rules.Get(), target: bob.Character.Id}))
		a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id)
	})
}

// refuses is the active ruleset refusing every attack on one body.
type refuses struct {
	rules.Ruleset
	target ids.ID
}

func (r refuses) CanAttack(attacker, defender *actors.Actor) error {
	if defender.Id == r.target {
		return errors.New("Fighting is not allowed here.")
	}
	return r.Ruleset.CanAttack(attacker, defender)
}

// Phase 5 slice 4.1 (plan review round 3, ruling 10): two players carrying
// aggros on each other into a room where PvP is off stop fighting; neither
// switches to the other, since the switch also asks CanAttack.
func TestPlayersWithMutualAggrosStopWherePvPIsOff(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
	setConfig(t,
		map[string]any{"GamePlay.PVP.Enabled": "limited", "GamePlay.PVP.MinimumLevel": 0},
		map[string]any{
			"GamePlay.PVP.Enabled":      string(configs.GetPVPConfig().Enabled),
			"GamePlay.PVP.MinimumLevel": int(configs.GetPVPConfig().MinimumLevel),
		})
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	aggroOn(alice.Character, bob.Character.Id, 1)
	aggroOn(bob.Character, alice.Character.Id, 1)
	w.told(9301)
	assert.Nil(t, settleTarget(a, room), "no swing at Bob, and no switch to him")
	assert.Nil(t, a.Character.Aggro)
	assert.Equal(t, []string{`This is not a PVP area.`}, w.told(9301))
}

// noCombatBuff is a test buff whose flag takes a body's swing.
const noCombatBuff = 93903

// Phase 5 slice 4.1 (plan review round 3): a body whose swing is taken
// (no-combat, an NPC's combat command) still settles its target every
// round, so its wait runs out at the limit.
func TestTakenTurnsStillCountTheWait(t *testing.T) {
	t.Run("no-combat", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
			BuffId: noCombatBuff, Name: "test calm", TriggerRate: "1 round", TriggerCount: 1000,
			Flags: []string{"no-combat"},
		}))
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		require.NoError(t, alice.Character.AddBuff(noCombatBuff, true))
		require.True(t, alice.Character.HasBuffFlag("no-combat"))
		a := actors.Get(alice.Character.Id)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		for turn := 1; turn <= 5; turn++ {
			require.NotNil(t, a.Character.Aggro, "turn %d", turn)
			combatTurn(a)
		}
		assert.Nil(t, a.Character.Aggro, "a no-combat body loses track at the limit")
	})

	t.Run("an NPC's combat command", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 30)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		dog.CombatCommands = []string{"emote growls."}
		dog.ActivityLevel = 100 // the combat command takes every turn it can
		d := actors.Get(dog.Character.Id)
		d.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		for turn := 1; turn <= 5; turn++ {
			require.NotNil(t, d.Character.Aggro, "turn %d", turn)
			combatTurn(d)
		}
		assert.Nil(t, d.Character.Aggro, "an NPC whose combat commands take its turns still loses track")
	})
}

// Phase 5 slice 4.1 (plan review round 3): a switch re-arms the aggro as
// attack <name> does, so a weapon with a delay waits its rounds before
// swinging at the new opponent.
func TestSwitchKeepsTheWeaponDelay(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	alice.Character.Equipment.Weapon = weapon(93104, "testmace", items.Bludgeoning, 1, 4, 1)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	aggroOn(&cat.Character, a.Id, 1)
	a.Character.SetAggro(ids.New(ids.NoCell), characters.DefaultAttack) // a target that is gone

	combatTurn(a)
	require.NotNil(t, a.Character.Aggro)
	assert.Equal(t, cat.Character.Id, a.Character.Aggro.Target, "switched")
	assert.Equal(t, 0, a.Character.Aggro.RoundsWaiting, "the mace's one round of delay was spent this turn")
	_, swung := cat.Character.PlayerDamage[9301]
	assert.False(t, swung, "no swing at the cat this turn")
}

// Phase 5 slice 4.1 (plan review round 2): a tie on Since goes to turn
// order as the round builds it (players by UserId, then NPCs by MobId),
// not to the order the bodies came into the room.
func TestNextOpponentTieGoesToTurnOrderNotRoomOrder(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93205, "rat", 10, goldenRoomA, 5, 5, 5, 30) // in the room before the cat
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	aggroOn(&rat.Character, alice.Character.Id, 9)
	aggroOn(&cat.Character, alice.Character.Id, 9)
	require.Equal(t, "cat", death.NextOpponent(a).Character.Name, "NPCs by MobId")

	carol := w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 30) // in the room before Bob
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
	aggroOn(carol.Character, alice.Character.Id, 9)
	aggroOn(bob.Character, alice.Character.Id, 9)
	assert.Equal(t, "Bob", death.NextOpponent(a).Character.Name, "players by UserId, before NPCs")
}

// Phase 5 slice 4.1 (ruling 12): a wait ends on its 5th turn in a row out of
// reach (the default limit); the body loses track of its target and those
// in its room see it give up.
func TestWaitRunsOutAtTheLimit(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.told(9301)

	for turn := 1; turn <= 4; turn++ {
		assert.Nil(t, settleTarget(a, room))
		require.NotNil(t, a.Character.Aggro, "turn %d: still waiting", turn)
	}
	assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "told once")
	assert.Nil(t, settleTarget(a, room))
	assert.Nil(t, a.Character.Aggro, "the 5th turn out of reach ends the wait")
	events.ProcessEvents()
	reads := map[int][]string{}
	for _, e := range w.rec.take() {
		reads[e.userId] = append(reads[e.userId], strings.TrimSuffix(e.text, "\n"))
	}
	assert.Equal(t, []string{`You lose track of <ansi fg="mobname">rat</ansi>.`}, reads[9301])
	assert.Equal(t, []string{`<ansi fg="username">Alice</ansi> mumbles about losing their quarry.`}, reads[9303])
}

// Phase 5 slice 4.1 (ruling 12): any turn that finds the target reachable,
// in the room or through an exit (shooting), resets the count.
func TestWaitCountResetsWhenTheTargetIsReachable(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	wait := func(turns int) {
		for i := 0; i < turns; i++ {
			require.NotNil(t, a.Character.Aggro, "the wait has not run out")
			require.Nil(t, settleTarget(a, room))
		}
	}

	wait(4)
	rat.Character.RoomId = goldenRoomA
	require.NotNil(t, settleTarget(a, room), "in the room: reachable")
	rat.Character.RoomId = goldenRoomB
	wait(4)
	require.NotNil(t, a.Character.Aggro, "the count started again: 4 of 5")

	a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	require.NotNil(t, settleTarget(a, room), "through the exit: reachable by shooting")
	a.Character.Equipment.Weapon = items.Item{}
	wait(4)
	assert.NotNil(t, a.Character.Aggro, "shooting reset the count too: 4 of 5")
}

// Phase 5 slice 4.1 (ruling 12, replacing behaviour.loseLostQuarry): an
// NPC whose player target left the game ends its fight at once (step 1);
// one waiting on an NPC in another room ends at the limit like anyone.
func TestNPCWaitsEndForOfflinePlayersAndNPCs(t *testing.T) {
	w := newCombatWorld(t, 1)
	bob := w.addPlayer(9302, "Bob", goldenRoomB, 10, 10, 10, 30)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 30)
	d, room := actors.Get(dog.Character.Id), rooms.LoadRoom(goldenRoomA)

	d.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	actors.Remove(bob.Character.Id) // Bob leaves the game
	assert.Nil(t, settleTarget(d, room))
	assert.Nil(t, d.Character.Aggro, "an offline player target ends the fight at once")

	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	d.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	for turn := 1; turn <= 4; turn++ {
		require.Nil(t, settleTarget(d, room))
		require.NotNil(t, d.Character.Aggro, "turn %d", turn)
	}
	assert.Nil(t, settleTarget(d, room))
	assert.Nil(t, d.Character.Aggro, "an NPC target out of reach ends at the limit too")
}

// Phase 5 slice 4.1 (plan review round 4): a shot through an exit re-arms
// through the same exit, so each shooting round queues one AggroChanged for
// the shooter, not a second one when the next lookup restores the exit.
func TestShootingRoundQueuesOneAggroChanged(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 200)
	alice.Character.SetAggroRemote("north", rat.Character.Id, characters.Shooting)
	events.ProcessEvents()
	w.rec.take()
	for round := 1; round <= 3; round++ {
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		n := 0
		for _, e := range w.rec.take() {
			if e.userId == 0 && strings.HasPrefix(e.text, "AggroChanged Alice ") {
				n++
			}
		}
		assert.Equal(t, 1, n, "round %d: one AggroChanged for the shooter", round)
		require.NotNil(t, alice.Character.Aggro)
		assert.Equal(t, "north", alice.Character.Aggro.ExitName, "round %d: the exit is kept", round)
	}
}

// Phase 5 slice 4.1 (plan review round 4): a target in reach that the body
// cannot perceive counts toward the wait limit. A hidden attacker who stops
// attacking is lost at the limit; one who keeps hitting resets the count with
// each attack (plan review round 6), so the body never loses track of it.
func TestUnseenTargetCountsTowardTheWait(t *testing.T) {
	t.Run("the hidden attacker stops", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 60)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 60)
		w.hide(alice.Character)
		b := actors.Get(bob.Character.Id)
		b.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
		for turn := 1; turn <= 5; turn++ {
			require.NotNil(t, b.Character.Aggro, "turn %d", turn)
			combatTurn(b)
		}
		assert.Nil(t, b.Character.Aggro, "Bob loses track of hidden Alice at the limit")
	})

	t.Run("the hidden attacker keeps hitting", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 1, 20, 60)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 500)
		w.hide(alice.Character)
		alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		w.told(9302)
		for round := 1; round <= 12; round++ {
			DoCombat(events.NewRound{})
			events.ProcessEvents()
			require.NotNil(t, bob.Character.Aggro, "round %d: each attack resets Bob's count", round)
			assert.Equal(t, alice.Character.Id, bob.Character.Aggro.Target)
		}
		for _, line := range w.told(9302) {
			assert.NotContains(t, line, "You lose track of", "Bob never loses track while she attacks")
		}
	})
}

// Phase 5 slice 4.1 (plan review round 6): being attacked by its target, hit
// or miss, resets a body's wait count. A body under steady fire from the next
// room keeps its aggro; once the fire stops it loses track at the limit.
func TestBeingAttackedByItsTargetResetsTheWait(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 1, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	dan := w.addPlayer(9304, "Dan", goldenRoomB, 10, 10, 10, 500)
	alice.Character.SetAggroRemote("north", dan.Character.Id, characters.Shooting)
	w.told(9304)
	for round := 1; round <= 12; round++ {
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		require.NotNil(t, dan.Character.Aggro, "round %d: Dan, shot every round, keeps his aggro", round)
		assert.Equal(t, alice.Character.Id, dan.Character.Aggro.Target)
	}
	outOfReach := 0
	for _, line := range w.told(9304) {
		assert.NotContains(t, line, "You lose track of", "steady fire never runs the wait out")
		if line == `Your target is out of reach.` {
			outOfReach++
		}
	}
	assert.Equal(t, 1, outOfReach, "one out-of-reach line for the whole wait, though each shot resets the count")

	// Alice's fire stops; Bob, whom Dan is not fighting, shoots him from
	// Alice's room instead, which does not reset Dan's wait on Alice. Dan's
	// turn in the round of Alice's last shot was his 1st out of reach (she
	// acts first), so the 4th round after it is his 5th.
	alice.Character.Aggro = nil
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 30, 1, 20, 60)
	bob.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	bob.Character.SetAggroRemote("north", dan.Character.Id, characters.Shooting)
	for round := 1; round <= 3; round++ {
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		require.NotNil(t, dan.Character.Aggro, "round %d after the fire stopped: still waiting", round)
	}
	DoCombat(events.NewRound{})
	events.ProcessEvents()
	assert.Nil(t, dan.Character.Aggro, "his 5th turn out of reach ends the wait")
	assert.Contains(t, w.told(9304), `You lose track of <ansi fg="username">Alice</ansi>.`)
}

// Phase 5 slice 4.1 (plan review round 6): "Your target is out of reach."
// comes once per wait. A turn that finds the target reachable ends the wait,
// so the next time it is out of reach the body is told again.
func TestOutOfReachLineOncePerWait(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	room := rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.told(9301)

	assert.Nil(t, settleTarget(a, room))
	assert.Nil(t, settleTarget(a, room))
	assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "the first wait: told once")

	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomA))
	require.NotNil(t, settleTarget(a, room), "reachable: the wait ends")
	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomB))
	assert.Nil(t, settleTarget(a, room))
	assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "a new wait: told again")
}

// Phase 5 slice 4.1 (plan review round 5): a shot goes through any exit the
// shoot command could use, temporary and mutator exits included, and a
// mutator's lock on an exit stops it.
func TestShootingExitUsesEveryExitTheRoomHas(t *testing.T) {
	const mutId = "test shooting exit"
	setup := func(t *testing.T) (*actors.Actor, *rooms.Room) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a := actors.Get(alice.Character.Id)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		return a, rooms.LoadRoom(goldenRoomA)
	}
	mutator := func(t *testing.T, room *rooms.Room, exits map[string]exit.RoomExit) {
		t.Cleanup(mutators.SetTestSpec(&mutators.MutatorSpec{MutatorId: mutId, Exits: exits}))
		require.True(t, room.Mutators.Add(mutId))
	}

	t.Run("a temporary exit", func(t *testing.T) {
		a, room := setup(t)
		delete(room.Exits, "north")
		room.ExitsTemp = map[string]exit.TemporaryRoomExit{
			"shimmering portal": {RoomId: goldenRoomB, Title: "shimmering portal"},
		}
		require.NotNil(t, settleTarget(a, room), "shoots through the portal")
		assert.Equal(t, "shimmering portal", a.Character.Aggro.ExitName)
	})

	t.Run("a mutator's exit", func(t *testing.T) {
		a, room := setup(t)
		delete(room.Exits, "north")
		mutator(t, room, map[string]exit.RoomExit{"trapdoor": {RoomId: goldenRoomB}})
		require.NotNil(t, settleTarget(a, room), "shoots through the trapdoor")
		assert.Equal(t, "trapdoor", a.Character.Aggro.ExitName)
	})

	t.Run("a mutator locks the exit", func(t *testing.T) {
		a, room := setup(t)
		mutator(t, room, map[string]exit.RoomExit{
			"north": {RoomId: goldenRoomB, Lock: gamelock.Lock{Difficulty: 5}},
		})
		assert.Nil(t, settleTarget(a, room), "the locked exit stops the shot: waits")
		require.NotNil(t, a.Character.Aggro)
		assert.Empty(t, a.Character.Aggro.ExitName)
	})
}

// Phase 5 slice 4.1 (plan review round 5, step 4): a body whose target is in
// the room but unseen switches to an attacker here it perceives; with no
// such attacker it keeps the target and the turn counts toward the wait.
func TestUnseenTargetGivesWayToASeenAttacker(t *testing.T) {
	setup := func(t *testing.T) (*combatWorld, *actors.Actor, *mobs.Mob) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		w.hide(alice.Character)
		b := actors.Get(bob.Character.Id)
		b.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
		rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
		return w, b, rat
	}

	t.Run("a perceived attacker here: switches", func(t *testing.T) {
		_, b, rat := setup(t)
		aggroOn(&rat.Character, b.Id, 1)
		d := settleTarget(b, rooms.LoadRoom(goldenRoomA))
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id, "the turn swings at the rat")
		assert.Equal(t, rat.Character.Id, b.Character.Aggro.Target)
	})

	t.Run("a seen target keeps the turn over another attacker", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		b := actors.Get(bob.Character.Id)
		b.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
		rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&rat.Character, b.Id, 1)
		d := settleTarget(b, rooms.LoadRoom(goldenRoomA))
		require.NotNil(t, d)
		assert.Equal(t, alice.Character.Id, d.Id, "no switch from a target in reach")
	})

	t.Run("nobody attacks: keeps the target and waits", func(t *testing.T) {
		_, b, _ := setup(t)
		target := b.Character.Aggro.Target
		d := settleTarget(b, rooms.LoadRoom(goldenRoomA))
		require.NotNil(t, d, "the turn goes at the unseen target")
		assert.Equal(t, target, d.Id)
		require.NotNil(t, b.Character.Aggro)
		assert.Equal(t, target, b.Character.Aggro.Target)
		assert.Equal(t, 1, b.Character.Aggro.OutOfReach, "the unseen turn counts")
	})
}
```

  `TestShootingExitUsesEveryExitTheRoomHas` needs a test-only way to register a mutator spec without
  writing a file (`mutators.SaveMutatorSpec` in `admin.go` writes one). Create
  `internal/mutators/testhelpers.go`, modelled on `internal/buffs/testhelpers.go`'s `SetTestSpec`:

```go
package mutators

// SetTestSpec adds a mutator spec in memory, for other packages' tests, and
// returns a func that takes it out again. Nothing is saved to disk.
func SetTestSpec(spec *MutatorSpec) (remove func()) {
	old, had := allMutators[spec.MutatorId]
	allMutators[spec.MutatorId] = spec
	return func() {
		if had {
			allMutators[spec.MutatorId] = old
			return
		}
		delete(allMutators, spec.MutatorId)
	}
}
```

  (The golden test rooms' zone `goldenzone` has a zone config, made by `rooms.AddTestRoom`
  (`testhelpers.go:24-29`); `Room.ActiveMutators` yields nothing without one (`rooms.go:2073-2078` at
  `1c8e607f`). A mutator in the room's `Mutators` list with `DespawnedRound` 0 is live
  (`mutators.go:183-185`). `shootingExit` resolves each name as `shoot <direction>` does,
  `usercommands/shoot.go:62-66`: `FindExitByName` for the room it leads to and the name to store,
  `GetExitInfo` for the lock, which checks active mutators' exits before the room's own
  (`rooms.go:805-820`). A temporary exit has no lock.)

  (These run with the test binary's config, which `GetGamePlayConfig` validates, so the limit is the
  default 5.)

  (The plan review round 2 probe: PvP `limited`, room A a PvP room and room B not; Alice with a bow
  kept shooting Bob after he moved to B, and a melee attacker followed its target into a non-PvP
  room and kept swinging. The `CanAttack` checks above end both.)

  Mutation probes (each run on the scratch copy; each killed):

| Mutation | Test | Failure |
|---|---|---|
| `SetAggro`: `since = c.Aggro.Since` dropped | `TestSetAggroSince...` | "the same target, re-armed" |
| `SetAggro`: `c.Aggro != nil && c.Aggro.Target == target` to `c.Aggro != nil` | same | "a new target" |
| `aggro.go`: `Since uint64` to ``Since uint64 `yaml:"-"` `` | `TestSavedAggroKeepsSince` | `"mob:..." does not contain "since: 77"` |
| `aggro.go`: `OutOfReachTold bool` to ``OutOfReachTold bool `yaml:"-"` `` | same | `"mob:..." does not contain "outofreachtold: true"` |
| `aggro.go`: `OutOfReach int` to ``OutOfReach int `yaml:"-"` `` | same | "a restored NPC resumes its wait count" |
| `NextOpponent`: `<` to `<=` | `TestNextOpponentIsTheLongestAttacker` | "a tie goes to turn order" |
| `NextOpponent`: drop the `Since` compare (`if next == nil {`) | same | "earliest Since, though the cat is first in turn order" |
| `NextOpponent`: range `candidates` instead of `TurnOrder(candidates)` | same | "accounts come first in turn order" |
| `NextOpponent`: drop `!rooms.Perceives(...)` | same | "a body she cannot perceive is skipped" |
| `NextOpponent`: drop the `Condition(b) != Up` test | same | "a body that is not up is skipped" |
| step 1: `switchOrEnd(...)` to `endAggro(a); return nil` | `TestSettleTargetSteps` | "1 gone, an attacker here" |
| step 4: `if death.NextOpponent(a) != nil {` to `if false {` | same | step 4 `Expected value not to be nil` |
| step 3: the `Shooting` branch returns nil | same | step 3 `Expected value not to be nil` |
| steps 3 and 4 swapped back (switch before shoot) | same | "a target it can shoot is reachable: kept" |
| `shootingExit`: the lock check `if exitInfo, ok := aRoom.GetExitInfo(exitName); ok && exitInfo.Lock.IsLocked() {` to `if false {` | same, and `TestShootingExitUsesEveryExitTheRoomHas` | "the exit is locked: waits" (`Expected nil, but got: &actors.Actor{...}`) |
| `shootingExit`: the lock read from `aRoom.Exits[exitName]` instead of `GetExitInfo` | `TestShootingExitUsesEveryExitTheRoomHas` | "the locked exit stops the shot: waits" |
| `shootingExit`: `FindExitByName(name)` replaced by `name, aRoom.Exits[name].RoomId` | same | "shoots through the portal" |
| `shootingExit`: the `ExitsTemp` loop ranges an empty map | same | "shoots through the portal" |
| `shootingExit`: the mutator loop ranges `(&rooms.Room{}).ActiveMutators` (none) | same | "shoots through the trapdoor" |
| step 4 for an unseen target here: `!rooms.Perceives(...) && death.Switch(a)` to `false && death.Switch(a)` | `TestUnseenTargetGivesWayToASeenAttacker` | "the turn swings at the rat" |
| the same, with the `!rooms.Perceives(...)` test dropped (a seen target given up for another attacker) | same | "no switch from a target in reach" |
| step 5: `if !a.Character.Aggro.OutOfReachTold {` (the one "out of reach" line) to `if true {` | same | "told once" |
| step 5: the mark `a.Character.Aggro.OutOfReachTold = true` not set | same, and `TestOutOfReachLineOncePerWait` | "told once" |
| step 5: the line keyed on the count again (`if a.Character.Aggro.OutOfReach == 1 {`) | `TestBeingAttackedByItsTargetResetsTheWait` | "one out-of-reach line for the whole wait, though each shot resets the count" |
| `inReach`: `a.Character.Aggro.OutOfReachTold = false` dropped | `TestOutOfReachLineOncePerWait` | "a new wait: told again" |
| step 5: `OutOfReach >= outOfReachRounds()` to `>` | `TestWaitRunsOutAtTheLimit` | "the 5th turn out of reach ends the wait" |
| `inReach`: the reset `OutOfReach = 0` dropped (one line serves a target in the room and one shot through an exit) | `TestWaitCountResetsWhenTheTargetIsReachable` | "the wait has not run out" |
| the swing end: the attacked defender's reset `d.Character.Aggro.OutOfReach = 0` dropped | `TestBeingAttackedByItsTargetResetsTheWait` (and `TestUnseenTargetCountsTowardTheWait`, "the hidden attacker keeps hitting") | "round 5: Dan, shot every round, keeps his aggro" (and "round 5: each attack resets Bob's count") |
| the swing end: the reset's `d.Character.Aggro.Target == a.Id` test dropped (any attacker resets the wait) | `TestBeingAttackedByItsTargetResetsTheWait` | "his 5th turn out of reach ends the wait" (`Expected nil, but got: &characters.Aggro{... OutOfReach:1}`) |
| step 5: `loseTrack(a, d)` not called | `TestNPCWaitsEndForOfflinePlayersAndNPCs` | "an NPC target out of reach ends at the limit too" |
| `settleTarget`: the same-room `CanAttack` check removed | `TestSettleTargetAsksCanAttackBeforeAttacking` | `Expected nil, but got: &actors.Actor{...}` (in the room) |
| `settleTarget`: the through-an-exit `CanAttack` check removed | same | `Expected nil, but got: &actors.Actor{...}` (through the exit; the next assertion then reports `&characters.Aggro{Type:1, ... ExitName:"north"...}`) |
| `NextOpponent`: `TurnOrder(SortBodies(candidates))` to `TurnOrder(candidates)` | `TestNextOpponentTieGoesToTurnOrderNotRoomOrder` | "NPCs by MobId" |
| `NextOpponent`: the `CanAttack(a, b)` filter removed | `TestPlayersWithMutualAggrosStopWherePvPIsOff` | "no swing at Bob, and no switch to him" (`Expected nil, but got: &actors.Actor{...}`) |
| `combatTurn`: the `no-combat` return moved back above `settleTarget` | `TestTakenTurnsStillCountTheWait` | "a no-combat body loses track at the limit" |
| `combatTurn`: the NPC combat-command block moved back above `settleTarget` | same | "an NPC whose combat commands take its turns still loses track" |
| `Switch`: `a.Character.Aggro.RoundsWaiting = 0` after `SetAggro` (a switch that skips the weapon delay) | `TestSwitchKeepsTheWeaponDelay` | "the mace's one round of delay was spent this turn" |
| step 3: the `AggroChanged` it queues removed | `TestSettleTargetSteps` (step 3, "a bow and an unlocked exit") | "turning to shoot queues one AggroChanged" |
| step 3: the "only if changed" guard to `if true {` (an `AggroChanged` every shooting turn) | same | "already shooting through that exit: none" |
| `resumeAfterFlee`: `a.Character.Aggro.Since = since` dropped | `TestFailedFleeKeepsSince` | "in its place" |
| `flee.go`: `since = actor.Character.Aggro.Since` dropped | same | "in its place" |
| `retaliate`: drop `Condition(d) != Up` | `TestRetaliate` (the only test that kills it: under `dikubasic` a defender that is not up is dead and its death ends everything in the same round, so no golden can see the difference) | "a body that is not up does not retaliate" |
| the swing end: the `SetAggroRemote` re-arm branch disabled (`if false {`), so a shot re-arms with plain `SetAggro` | `TestShootingRoundQueuesOneAggroChanged` | "round 1: the exit is kept" (and two `AggroChanged` from round 2) |
| `inReach`: the unseen branch returns `d` without counting | `TestUnseenTargetCountsTowardTheWait` | "Bob loses track of hidden Alice at the limit" (the "keeps hitting" subtest passes under it: since round 6, it checks that Bob never loses track) |
| `retaliate`: drop `d.Character.Aggro != nil` | same | "a body already fighting keeps its target" |
| `retaliate`: the shooting branch `if false` | same | `Not equal` (`Shooting`) |
| `retaliate`: `CanAttack` check `if false` | same | "the ruleset's CanAttack refuses" |
| the call `retaliate(d, a)` removed | `TestCombatGoldenPvMToDeath` | `pvm_death.txt differs at line 9` |
| step 2: `switchOrEnd(a, "Your rage subsides.")` to `endAggro(a); return nil` | `TestCombatGoldenDeathReleasesAttackers` | `differs at line 15` |
| the swing end ends `a` too (`endAggro(a)` before `endAggro(d)`) | `TestCombatGoldenSwitchToLongestAttacker` | `switch_longest.txt differs at line 50` |
| `Die`: `releaseAttackers(victim)` removed | `TestCombatGoldenDeathReleasesAttackers` | `differs at line 19` |

- [ ] **Step 9: The existing goldens.** Run
  `go test ./internal/hooks/ -run 'Golden|Reads' -count=1 -update`, then `git diff --stat` and read
  every hunk. Expected change in kind, written before the run, and what the scratch run showed:

| Fixture | Expected (before the run) | Run on the scratch copy |
|---|---|---|
| `pvm_death` | the rat swings back every round from round 1; the rat still dies, maybe later; Alice survives | as expected: the rat dies in round 6 (was 4); Alice ends at 5/10 |
| `pvp` | Bob fights back; Bob still dies; Alice takes damage | Bob dies in round 3 (was 4); Bob's swings all miss, Alice ends at 10/10 |
| `mvp` | Alice swings from round 2; she still dies about round 6; the guard takes damage | as expected: Alice dies in round 6; the guard ends at 5/10 |
| `mvm` | the rat swings back; the rat still dies about round 3 | as expected: the rat dies in round 4 (was 3) |
| `flee_blocked` | her blocked flee ends her aggro; the guard's hit gives her one; she swings from round 2 | as expected: she swings in rounds 2-3 (misses); the guard ends at 10/10 |
| `flee_success` | the guard waits on Alice (no "can't be found", no end); a round 2 is recorded with nothing read | as expected: 2 rounds (was 1); the guard ends `aggro=[Alice DefaultAttack wait=0]` |
| `shoot_exit` | the rat takes a waiting aggro and still queues `go south` (recorded, not run); in round 2 Alice shoots again (was "Your target can't be found.") | as expected; with the round 4 re-arm through the exit, Alice's state lines read `aggro=[rat Shooting wait=0 exit=north]` after rounds 1 and 2, and round 2 has one `AggroChanged Alice` (the re-arm's) |
| `observer` | the rat swings at Alice from the round Alice first hits; Carol reads its lines | as expected: the rat ends at 1/12, Alice at 8/10, both still fighting after 6 rounds |
| `hidden_attacker` | Bob takes an aggro on Alice and reads "You can't seem to find your target." each turn; Bob still dies | as expected: four such lines; Bob dies in round 6 |
| `death_npc` | the rat (club) swings back; the rat probably still dies | **surprise: Alice dies in round 5; the rat survives at 1/12 and the test's `require` fails (Q10)** |
| `death_player` | Alice swings back from round 2; outcome uncertain | Alice still dies in round 6; the guard ends at 2/10 |
| `death_pvp` | like `pvp` | Bob dies in round 4 (was 3); Alice ends at 7/10 |
| `npc_regen` | the rat swings in rounds 1-2; the numbers shift; heals raise the rat | the rat ends the fight at 32/40 and heals to 36; Alice takes no damage |
| reads `E-move-look` | the rat's lines appear where Ann attacks it | two lines added after line 130: Bob reads `rat's teeth and nails connects with Ann.`, Ann reads `rat strikes you with their teeth and nails, causing 1 damage.` |

  Every other reads golden byte-identical. Then rerun without `-update` three times
  (`go test ./internal/hooks/ -count=1`): `ok` each time (with Step 4; without it, four goldens
  differ run to run).

- [ ] **Step 10: The new fixtures.** Create `internal/hooks/retaliation_golden_test.go`:

```go
package hooks

// Combat golden fixtures for retaliation and target switching (Phase 5
// slice 4.1, rulings 6-9). Same harness as combat_golden_test.go.

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/require"
)

// Two NPCs attack a player in turn while she fights a third she chose; she
// kills it and turns to the one that attacked first (the rat, Since 1000),
// not the one first in turn order (the cat, MobId 93204, Since 1001).
func TestCombatGoldenSwitchToLongestAttacker(t *testing.T) {
	w := newCombatWorld(t, 1101)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 8)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	rat := w.addNPC(93205, "rat", 10, goldenRoomA, 5, 5, 5, 30)
	alice.Character.SetAggro(dog.Character.Id, characters.DefaultAttack)
	rat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack) // Since 1000
	util.IncrementRoundCount()
	cat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack) // Since 1001
	w.start()
	w.rounds(8)
	w.check("switch_longest")
	require.Equal(t, rat.Character.Id, alice.Character.Aggro.Target, "she turned to the rat")
}

// A player chooses her target; an NPC attacking her dies at another
// player's hand, and her target does not move.
func TestCombatGoldenChosenTargetStays(t *testing.T) {
	w := newCombatWorld(t, 1102)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	carol := w.addPlayer(9303, "Carol", goldenRoomA, 30, 30, 20, 60)
	carol.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 4)
	alice.Character.SetAggro(dog.Character.Id, characters.DefaultAttack)
	rat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	carol.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(4)
	w.check("chosen_target_stays")
	require.Less(t, rat.Character.Health, 1, "the fixture must reach the rat's death")
	require.Equal(t, dog.Character.Id, alice.Character.Aggro.Target, "her chosen target stays")
}

// The target flees and another attacker is in the room: the body switches.
func TestCombatGoldenTargetFleesSwitch(t *testing.T) {
	w := newCombatWorld(t, 1103)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 100, 5, 5, 30)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	cat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	fleeAggro(&rat.Character)
	w.start()
	w.rounds(3)
	w.check("target_flees_switch")
	require.Equal(t, goldenRoomB, rat.Character.RoomId, "the rat fled")
	require.Equal(t, cat.Character.Id, alice.Character.Aggro.Target, "she turned to the cat")
}

// The target flees and nobody else attacks: the aggro waits, saying so
// once, and swings when the target comes back.
func TestCombatGoldenTargetFleesWait(t *testing.T) {
	w := newCombatWorld(t, 1104)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 100, 5, 5, 30)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	fleeAggro(&rat.Character)
	w.start()
	w.rounds(3)
	require.Equal(t, goldenRoomB, rat.Character.RoomId, "the rat fled")
	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomA))
	w.rounds(1)
	w.check("target_flees_wait")
	require.Equal(t, rat.Character.Id, alice.Character.Aggro.Target)
}

// Shot from another room: with a bow the body shoots back through the
// exit; without, it waits ("Your target is out of reach."); an NPC walks
// over (its queued "go" is run) and swings on arrival.
func TestCombatGoldenShotFromAnotherRoom(t *testing.T) {
	w := newCombatWorld(t, 1105)
	w.runInputs()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomB, 30, 30, 20, 60)
	bob.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	carol := w.addPlayer(9303, "Carol", goldenRoomB, 30, 30, 20, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 10, 5, 5, 60)
	rat.Character.ActionPoints = 100 // a walk costs action points; the harness runs no turn that refills them
	alice.Character.SetAggroRemote("north", bob.Character.Id, characters.Shooting)
	w.start()
	w.rounds(2)
	alice.Character.SetAggroRemote("north", carol.Character.Id, characters.Shooting)
	w.rounds(2)
	alice.Character.SetAggroRemote("north", rat.Character.Id, characters.Shooting)
	w.rounds(3)
	w.check("shot_from_another_room")
}

// A hidden attacker: the defender takes an aggro on it and cannot find it.
func TestCombatGoldenHiddenAttackerRetaliation(t *testing.T) {
	w := newCombatWorld(t, 1106)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 5, 20, 60)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 60)
	w.hide(alice.Character)
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(2)
	w.check("hidden_attacker_retaliation")
	require.Equal(t, alice.Character.Id, bob.Character.Aggro.Target)
}

// Three bodies target the rat when the round starts. Carol kills it; the
// dog's own turn then finds it not up and turns to Bob, who attacks the dog;
// the rat's death turns Alice to the cat that attacks her and ends Carol's
// fight (she has nobody to turn to).
func TestCombatGoldenDeathReleasesAttackers(t *testing.T) {
	w := newCombatWorld(t, 1107)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 4, 6, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 5, 5, 5, 60)
	carol := w.addPlayer(9303, "Carol", goldenRoomA, 5, 5, 5, 60)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 60)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 1, 1, 1, 1)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	carol.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	dog.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	cat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	bob.Character.SetAggro(dog.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(1)
	w.check("death_releases_attackers")
	require.Less(t, rat.Character.Health, 1, "the fixture must reach the rat's death")
	require.Equal(t, cat.Character.Id, alice.Character.Aggro.Target)
	require.Nil(t, carol.Character.Aggro)
	require.Equal(t, bob.Character.Id, dog.Character.Aggro.Target)
}

// A wait runs out (ruling 12): the rat flees, nobody else fights Alice, and
// after 5 turns out of reach she loses track of it; Carol sees her give up.
func TestCombatGoldenWaitRunsOut(t *testing.T) {
	w := newCombatWorld(t, 1108)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 60)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 100, 5, 5, 30)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	fleeAggro(&rat.Character)
	w.start()
	w.rounds(10)
	w.check("wait_runs_out")
	require.Equal(t, goldenRoomB, rat.Character.RoomId, "the rat fled")
	require.Nil(t, alice.Character.Aggro, "she lost track of it")
}
```

  Record: `go test ./internal/hooks/ -run 'TestCombatGolden(SwitchToLongestAttacker|ChosenTargetStays|TargetFleesSwitch|TargetFleesWait|ShotFromAnotherRoom|HiddenAttackerRetaliation|DeathReleasesAttackers|WaitRunsOut)$' -count=1 -update`.
  What each recorded on the scratch copy (read each file against it):
  - `switch_longest`: round 3 Alice kills the dog and reads `You turn to fight <ansi fg="mobname">rat</ansi>!`
    (after the XP line, before the gold line); from round 4 she swings at the rat.
  - `chosen_target_stays`: Carol kills the rat in round 2; Alice's aggro stays on the dog to round 4.
  - `target_flees_switch`: round 1 Alice hits the rat, the rat flees north; round 2 she reads
    `You turn to fight <ansi fg="mobname">cat</ansi>!` and swings at the cat in the same turn (fists:
    no weapon delay).
  - `target_flees_wait`: round 2 Alice reads `Your target is out of reach.`; round 3 nothing; after
    the rat is moved back, her next turn swings (the second `rounds` call numbers from 1 again).
  - `shot_from_another_room`: Bob shoots back through `south` in round 1; Carol takes a waiting aggro
    and reads `Your target is out of reach.` on her first turn; in the third part the rat queues
    `go south` and an angry emote, the harness runs them, and the rat arrives in room A (Alice reads
    `rat enters from the north`, Bob and Carol `rat leaves towards the south exit`); from the next
    round it swings at Alice in her room and she hits it there. Since plan review round 6, Alice's
    second shot at Carol resets Carol's wait count (not her told mark, so no second "out of reach"
    line), and her count is 4 when the fixture ends (part 3's round 3), so she still waits on Alice
    (`aggro=[Alice DefaultAttack wait=0]`). Before round 6 her 5th turn fell in
    part 3's round 3 and she lost track (the fourth run below).
  - `wait_runs_out`: round 1 Alice misses and the rat flees north; round 2 Alice reads `Your target is
    out of reach.`; rounds 3 to 5 nothing; round 6 (her 5th turn out of reach) she reads `You lose track
    of rat.`, Carol reads `Alice mumbles about losing their quarry.`, `AggroChanged Alice`, and the
    rounds stop (nobody fights).

  Fourth run (ruling 12; expectations written first: only `shot_from_another_room` changes, Carol
  losing track of Alice in part 3 round 3; `flee_success`, `shoot_exit` and `target_flees_wait` wait 2
  turns or fewer and stay as they were). Result: as expected, four lines in `shot_from_another_room`
  (Bob's mumble line, Carol's lose-track line, `AggroChanged Carol`, her state `aggro=[none]`), and the
  new `wait_runs_out` file.

  Fifth run (plan review round 4; expectations written first: the re-arm through the exit takes away
  the seven extra `AggroChanged` lines in `shot_from_another_room` and one in `shoot_exit`; unseen
  targets counting toward the wait change `hidden_attacker`, Bob losing track of hidden Alice in round
  5 and retaliating again in round 6; the recovery-room attacker check changes nothing). Result: the
  `AggroChanged` lines went as expected, and with them came a change not predicted: every state line of
  a shooter now shows its kept exit (`aggro=[Bob Shooting wait=0 exit=north]`, `[Alice Shooting wait=0
  exit=south]`, `[rat Shooting wait=0 exit=north]`), since the harness prints `ExitName` and the aggro
  now keeps it. Decided: it is the same fact the fix is about. `hidden_attacker` did not change:
  Alice's dagger waits its first round, so Bob's aggro begins in round 2 and his 5th unseen turn would
  be in round 6, but Alice kills him first in round 6 (she acts first). The miss was in the
  expectation, not the code; `TestUnseenTargetCountsTowardTheWait` covers the rule.

  Sixth run (plan review round 5; expectations written first: no golden changes. No fixture has a
  temporary or mutator exit, so `shootingExit` finds the same exits; in `hidden_attacker` and
  `hidden_attacker_retaliation` nobody Bob perceives attacks him while his target is hidden, so
  step 4 for an unseen target never switches; dropping `NextOpponent`'s `RoomId` test changes no
  candidate list). Result: as expected, every golden byte-identical after `-update`
  (`diff -rq --strip-trailing-cr` against a copy taken before).

  Seventh run (plan review round 6, being attacked by its target resets the wait count;
  expectations written first: only `shot_from_another_room` changes. Carol is shot in both of
  part 2's rounds; the second shot resets her count, so she reads `Your
  target is out of reach.` a second time, and part 3's three rounds take her only to 4: her
  lose-track line, Bob's mumble line, `AggroChanged Carol` and her `aggro=[none]` go, her state
  keeps `aggro=[Alice DefaultAttack wait=0]`. `shoot_exit`, `hidden_attacker_retaliation` and
  `hidden_attacker` show no change: an NPC's tells are not recorded, state lines do not print
  `OutOfReach`, and Bob dies before his limit). Result: exactly that, five hunks in
  `shot_from_another_room` (one line added at 78, three removed at 140-147, Carol's state line at
  152), every other golden byte-identical. The repeated line showed the count could not also carry
  "told once"; the controller chose a separate mark (option (b)).

  Eighth run (plan review round 6, the told mark `OutOfReachTold`; expectations written first: only
  Carol's second `Your target is out of reach.` in `shot_from_another_room` goes, the rest of the
  seventh run's change stays; no other fixture has a waiting body found reachable and then out of
  reach again, and state lines do not print the mark). Result: exactly that, one line removed
  (line 78), every other golden byte-identical.

  Each new fixture's mutation (run on the scratch copy; each fails its test):

| Fixture | Mutation | Failure |
|---|---|---|
| `TestCombatGoldenSwitchToLongestAttacker` | `NextOpponent`: `if next == nil \|\| b.Character.Aggro.Since < ... {` to `if next == nil {` | `switch_longest.txt differs from the golden at line 50` |
| `TestCombatGoldenChosenTargetStays` | `settleTarget`: switch to `NextOpponent` before the same-room check (a chosen target replaced while here) | `chosen_target_stays.txt differs from the golden at line 10` |
| `TestCombatGoldenTargetFleesSwitch` | `settleTarget`: step 4 (switch when attacked here) removed | `target_flees_switch.txt differs from the golden at line 23` |
| `TestCombatGoldenTargetFleesWait` | `settleTarget`: `if !a.Character.Aggro.OutOfReachTold {` to `if true {` (the line every turn) | `target_flees_wait.txt differs from the golden at line 25` |
| `TestCombatGoldenWaitRunsOut` | `loseTrack`: the observer text emptied | `wait_runs_out.txt differs from the golden at line 58` |
| `TestCombatGoldenShotFromAnotherRoom` | `inFight`: the room test to `true {` (a waiting aggro blocks `go`) | `shot_from_another_room.txt differs from the golden at line 91` |
| `TestCombatGoldenHiddenAttackerRetaliation` | `retaliate`: also return when `!rooms.Perceives(d.Character, a.Character)` | `hidden_attacker_retaliation.txt differs from the golden at line 11` |
| `TestCombatGoldenDeathReleasesAttackers` | `Die`: `releaseAttackers(victim)` removed | `death_releases_attackers.txt differs at line 19` |

  And for Step 5a's flipped test, `TestNPCShootingAPlayerAllowedWhenPvPDisabled`: delete `CanAttack`'s
  new `if accountOf(attacker) == nil { return nil }`; it fails with `Expected value not to be nil`,
  "an NPC shot is not PvP" (run).

  Third run (plan review round 2 fixes; expectations written first: the `CanAttack` re-check and the
  sorted tie-break change no golden; step 3's `AggroChanged` adds one line per remote shooting turn
  whose aggro lost its exit at the previous re-arm). Result, as expected: `shoot_exit` gains
  `AggroChanged Alice room=93001` in round 2 (line 29); `shot_from_another_room` gains seven
  `AggroChanged` lines (Alice in the first two parts' second rounds, Bob in every round after his
  first); nothing else changed.

  First run, before the answers to Q7-Q15 (kept as evidence): the rat stayed in room B (`go` refused a
  body with an aggro; Q8) and the waiting line was "Your target is no longer here." (Q13). Second run,
  expectations written first: only `target_flees_wait` and `shot_from_another_room` change (the line
  text; the rat walks); everything else byte-identical. Result: as expected, with one surprise, fixed in
  the fixture: the rat did not walk until it had action points (an NPC's `go` costs them and is refused
  silently without, `go_test.go:85`; the harness runs no turn that refills them), hence
  `rat.Character.ActionPoints = 100`.
  - `hidden_attacker_retaliation`: Bob reads `You can't seem to find your target.` in rounds 1 and 2.
  - `death_releases_attackers`: round 1: Carol kills the rat; the dog's turn finds the rat not up and
    turns to Bob (step 2); `Die` turns Alice to the cat (she reads `You turn to fight ... cat!` after the
    XP line) and ends Carol's aggro silently.

- [ ] **Step 11: Order dependence.** Before this task: `go test ./internal/hooks/ -count=2` fails
  only `TestCombatGoldenFleeSuccess` (once), and `-shuffle=1` to `-shuffle=6` fail only it. After:
  the same (scratch, after the answers to Q7-Q15: `-count=2` fails only `FleeSuccess`, once;
  `-shuffle=1` to `6` fail only `FleeSuccess`; the full suite shows only the 3 known Windows failures;
  three plain reruns of `internal/hooks` print `ok`). Run both and report the failing names; any new
  name is a regression.

- [ ] **Step 12: Smoke fights a real rat and checks both endings (ruling 11).** With ruling 10 the
  rat now swings back at bot 0 in the fight step. The numbers behind ruling 11, from the smoke world
  (the shipped config and `_datafiles/world/empty`, which launch mode copies):
  - **Bot 0**: a new human (`cmd/smoke/script.go:176-177`), level 1: health 6 (the prompt reads
    `HP:6/6`, `fight_test.go:38`), Speed, Strength and Perception 0 after `Validate`.
  - **The rat** (`mobs/startland/1-rat.yaml`, race 10 `rodent`): spawned in Startland at a random level
    1 to 5 (`rooms/startland/zone-config.yaml` `autoscale: 1..5`; `rooms.go:522-547`), stats from
    `AutoTrain`. Built through `mobs.NewMobById(1, 1, level)`: level 1: health 1; level 2: 4; level 3:
    5; level 4: 15; level 5: 16. Its attack: one, `1d2` (race `damage: diceroll: 1d2`), strength and
    perception bonuses 0 at these stats; the most it did in one round in the simulation below was 3 at
    level 1 and 6 at levels 2 to 5 (crits, whose bonus grows with the Perception difference).
  - **Turn order**: players act first (`dikubasic.TurnOrder`), so bot 0 swings first each round, and
    the rat, given its aggro by retaliation during bot 0's turn, swings later in the same round.
  - **Simulated** (scratch test, `calculateCombat` both ways, bot first, from full health; 40,000
    seeded fights a level): bot 0 dies first in 833 of 40,000 fights against a level 1 rat (2.1%;
    200,000 fights: 4,207), 35,606 (89%) at level 2, 37,434 (94%) at level 3, and every fight at
    levels 4 and 5 (bot 0 never kills a 15-health rat before it takes 6).

  So with the shipped world's random rat the step would fail at random. Ruling 11: launch mode's
  world copy pins two rats, bot 0 fights the level 1 rat and then the level 5 rat, and each fight
  accepts either ending and checks it. The rodent's damage stays as shipped.

  **Lengths, re-simulated for the two pinned rats** (40,000 seeded fights each, to either ending, bot
  first each round):

| Rat | Bot 0 starts at | Rat dies | Bot 0 dies | Longest fight |
|---|---|---|---|---|
| level 1 (health 1) | 6/6 | 39,167 | 833 | 17 rounds |
| level 1 | 1/6 (just respawned) | 26,096 | 13,904 | 10 rounds |
| level 5 (health 16) | 6/6 | 0 | 40,000 | 10 rounds |
| level 5 | 1/6 | 0 | 40,000 | 4 rounds |

  So fight 1 ends in the rat's death about 98% of the time and fight 2 always in bot 0's; no fight
  ran past 17 rounds (68 s at 4-second rounds), well inside `fightStepTimeout` (4 minutes per fight,
  `fight.go:11-14`). Pinning the levels also removes the Q18 timeout flake in launch mode: the long
  fights came from level 4 and 5 rats that bot 0 had to kill.

  **Targeting.** The two rats get distinct names in the world copy only: "small rat" (level 1) and
  "large rat" (level 5), through the room spawn's `name` override (`rooms/spawninfo.go:12`, applied at
  `rooms.go:560-561`). Bot 0 types `attack small`, then `attack large`. `attack` resolves its target
  with `room.FindByName` (`usercommands/attack.go:176`; `rooms.go:1392`), whose `findMobByName`
  lists the room's NPCs as `<name>#<n>` and matches the typed word as a prefix (`rooms.go:1452-1492`,
  `util.FindMatchIndexIn`, `stringMatch` `util.go:407-427`): "small" matches only "small rat#n" and
  "large" only "large rat#n", whatever order they spawned in. Town Square also holds the plain "rat"
  the admin step spawns to speak (`cmd/smoke/admin.go:58-71`, `mob spawn rat`), which may wander:
  "small" and "large" are neither prefixes of "rat#n" nor of any word in it (the contains fallback,
  `util.go:412-421`, matches word prefixes), so `attack` never picks it, and the fight and observer
  checks wait for "small rat has died." / "large rat has died.", which its death would not print.
  Both names keep the word "rat", so the
  death lines ("small rat has died.") and the observer's `\brat\b` check still read them.

  **The recovery route.** A player killed in Town Square respawns in a private copy of room 75
  ("Waiting room", `world/empty/rooms/shadow_realm/75.yaml`), whose only exit, `portal`, leads to room
  1, Town Square, the fight room. There is no wait: the existing death step walks out at once
  (`script.go:229-230`), and the scratch runs below did the same after a rat killed bot 0. Its death
  puts the ring it carries into its corpse in Town Square, so after walking back bot 0 takes it back
  (`get copper ring <its name>`, as the death step does), and the later steps (relogin "still holds
  the item", suicide, the corpse) run unchanged: a name reaches the newest of its corpses first, so the
  suicide step's `get` finds the suicide corpse.

  **World copy.** Removing End of the Line's three rats from the copy and giving the two fight rats a
  24-hour respawn are recorded in the spec's Decisions table (row "The smoke world copy also removes
  End of the Line's three rats and gives the two fight rats a 24-hour respawn"). In
  `cmd/smoke/server.go`, `launch`, after the world copy (`os.CopyFS(world, ...)`, `:77-80`):

```go
	if err := fightRats(world); err != nil {
		return s, err
	}
```

  the imports gain `"gopkg.in/yaml.v2"` (already a module dependency, `go.mod:17`), and before
  `// serverBinary is where launch mode builds the server.`:

```go
// fightRats sets up launch mode's world copy for the smoke fight (Phase 5
// slice 4.1, ruling 11). Town Square (startland room 1) gets exactly two
// rats: a "small rat" pinned at level 1 and a "large rat" pinned at level 5
// (launchRats), neither wandering (an idle list of one empty command) and
// neither respawning within a run. End of the Line (room 2), where the
// shipped world's three rats live and from where they wander, gets none.
// The admin step still spawns one plain "rat" in Town Square (npcSpeaks),
// which may wander; bot 0 attacks "small" and "large", which never match
// it, and the fight checks wait for "small rat has died." and "large rat
// has died.". The rodent race's damage is unchanged:
// either rat fights back, and the fight step checks whichever ending comes.
// Only the temporary copy changes.
func fightRats(world string) error {
	dir := filepath.Join(world, "rooms", "startland")
	if err := editSpawns(filepath.Join(dir, "2.yaml"), func(spawns []any) ([]any, error) {
		for _, sp := range spawns {
			if mobIdOf(sp) != ratMobId {
				return nil, fmt.Errorf("room 2 spawns more than rats: update fightRats")
			}
		}
		return nil, nil
	}); err != nil {
		return err
	}
	return editSpawns(filepath.Join(dir, "1.yaml"), func(spawns []any) ([]any, error) {
		for _, sp := range spawns {
			if mobIdOf(sp) == ratMobId {
				return nil, fmt.Errorf("room 1 already spawns a rat: update fightRats")
			}
		}
		for i, level := range []int{1, 5} {
			spawns = append(spawns, yaml.MapSlice{
				{Key: "mobid", Value: ratMobId},
				{Key: "name", Value: launchRats[i].name},
				{Key: "level", Value: level},
				{Key: "idlecommands", Value: []string{""}},
				{Key: "respawnrate", Value: "24 real hours"},
			})
		}
		return spawns, nil
	})
}

// ratMobId is the rat's mob template (world/empty/mobs/startland/1-rat.yaml).
const ratMobId = 1

// editSpawns rewrites a room file's spawninfo list through edit; an empty
// result removes the key.
func editSpawns(path string, edit func([]any) ([]any, error)) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("fight rats: %w", err)
	}
	var room yaml.MapSlice
	if err := yaml.Unmarshal(b, &room); err != nil {
		return fmt.Errorf("fight rats: %s: %w", path, err)
	}
	out := yaml.MapSlice{}
	var spawns []any
	for _, item := range room {
		if item.Key == "spawninfo" {
			list, _ := item.Value.([]any)
			spawns = list
			continue
		}
		out = append(out, item)
	}
	spawns, err = edit(spawns)
	if err != nil {
		return fmt.Errorf("fight rats: %s: %w", path, err)
	}
	if len(spawns) > 0 {
		out = append(out, yaml.MapItem{Key: "spawninfo", Value: spawns})
	}
	b, err = yaml.Marshal(out)
	if err != nil {
		return fmt.Errorf("fight rats: %s: %w", path, err)
	}
	return os.WriteFile(path, b, 0o644)
}

// mobIdOf is a spawn entry's mobid, or 0.
func mobIdOf(spawn any) int {
	if m, ok := spawn.(yaml.MapSlice); ok {
		for _, item := range m {
			if item.Key == "mobid" {
				if n, ok := item.Value.(int); ok {
					return n
				}
			}
		}
	}
	return 0
}
```

  (An empty idle command is a no-op; a spawn's `idlecommands` replaces the template's, whose
  `'wander'` would take the rat out of the room, `rooms.go:555-556`. `respawnrate` uses the same
  `<qty> real <unit>` form the shipped spawns use; parsed by `gametime.parsePeriod`.) Create
  `cmd/smoke/fightrats_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

// fightRats (ruling 11) gives the world copy's Town Square exactly two rats,
// small at level 1 and large at level 5, neither wandering, and takes the
// shipped rats out of End of the Line; the shipped files are not touched.
func TestFightRatsPinsTwoRatsInTownSquare(t *testing.T) {
	shipped := filepath.Join("..", "..", "_datafiles", "world", "empty", "rooms", "startland")
	world := t.TempDir()
	dir := filepath.Join(world, "rooms", "startland")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	originals := map[string][]byte{}
	for _, f := range []string{"1.yaml", "2.yaml"} {
		b, err := os.ReadFile(filepath.Join(shipped, f))
		if err != nil {
			t.Fatal(err)
		}
		originals[f] = b
		if err := os.WriteFile(filepath.Join(dir, f), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := fightRats(world); err != nil {
		t.Fatal(err)
	}

	type spawn struct {
		MobId        int      `yaml:"mobid"`
		Name         string   `yaml:"name"`
		Level        int      `yaml:"level"`
		IdleCommands []string `yaml:"idlecommands"`
	}
	type room struct {
		RoomId    int     `yaml:"roomid"`
		Title     string  `yaml:"title"`
		SpawnInfo []spawn `yaml:"spawninfo"`
	}
	read := func(f string) room {
		var r room
		b, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			t.Fatal(err)
		}
		if err := yaml.Unmarshal(b, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}

	square := read("1.yaml")
	if square.RoomId != 1 || square.Title != startRoomTitle {
		t.Errorf("room 1 lost its fields: %+v", square)
	}
	rats := []spawn{}
	for _, sp := range square.SpawnInfo {
		if sp.MobId == ratMobId {
			rats = append(rats, sp)
		}
	}
	if len(rats) != 2 {
		t.Fatalf("Town Square spawns %d rats, want 2: %+v", len(rats), square.SpawnInfo)
	}
	for i, want := range []struct {
		name  string
		level int
	}{{"small rat", 1}, {"large rat", 5}} {
		got := rats[i]
		if got.Name != want.name || got.Level != want.level || len(got.IdleCommands) != 1 || got.IdleCommands[0] != "" {
			t.Errorf("rat %d is %+v, want %s at level %d with one empty idle command", i, got, want.name, want.level)
		}
		if launchRats[i].name != want.name {
			t.Errorf("launchRats[%d] is %q, want %q (the fight step attacks what fightRats spawns)", i, launchRats[i].name, want.name)
		}
	}
	if n := len(read("2.yaml").SpawnInfo); n != 0 {
		t.Errorf("End of the Line still spawns %d mobs, want 0", n)
	}
	for f, b := range originals {
		again, _ := os.ReadFile(filepath.Join(shipped, f))
		if string(again) != string(b) {
			t.Errorf("the shipped %s changed", f)
		}
	}
}
```

  Mutations (run, each killed): the levels `[]int{1, 5}` to `[]int{2, 5}`; room 2's edit returning
  `spawns` instead of `nil` (its rats kept); the `idlecommands` entry removed; `launchRats` reordered
  (large first).

  **The fight step.** In `cmd/smoke/fight.go`, everything from the file's start through the end of
  `fight` (`:1-113`) is replaced by:

```go
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
```

  In `observe` (`fight.go:163` at `1c8e607f`; Task 6 shortened the file above it), the wait

```go
	s.wait("fight observer: sees the rat die", "rat has died.")
	if s.err != nil {
		return
	}
```

  becomes

```go
	// Each fight's ending, whichever it was: the rat's death line or bot 0's.
	// Each ending must be read once per fight it ended, so bot 0 dying in
	// both fights needs its death line twice.
	s.err = s.c.until(s.ctx, s.c.timeout, "fight observer: sees each fight end", func(got string, _ bool) (bool, string) {
		return sawEndings(got, fc.endings)
	})
	if s.err != nil {
		return
	}
```

  and before `// observe is bot 1's step:` add:

```go
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
```

  (Plan review round 5: one `s.wait` per ending let the first death line satisfy the second wait
  when bot 0 died in both fights, because the text mark moves only on a send and the observer sends
  nothing after its look. The observer's screen runs from that look, so it holds every ending once
  per fight.) Append to `cmd/smoke/fight_test.go`:

```go

// Plan review round 5: each fight's ending is read once per fight, so when
// bot 0 dies in both fights the first death line does not satisfy the
// second wait.
func TestSawEndingsCountsEachFight(t *testing.T) {
	died := "Smkabc0 has DIED!"
	twice := []string{died, died}
	if ok, _ := sawEndings("...\n"+died+"\n...", twice); ok {
		t.Error("one death line satisfied two fights")
	}
	if ok, missing := sawEndings(died+"\n...\n"+died+"\n", twice); !ok {
		t.Errorf("two death lines for two fights: %s", missing)
	}
	if ok, missing := sawEndings("small rat has died.\n"+died+"\n", []string{"small rat has died.", died}); !ok {
		t.Errorf("one of each: %s", missing)
	}
	if ok, _ := sawEndings("small rat has died.\n", []string{"small rat has died.", died}); ok {
		t.Error("a missing ending was not reported")
	}
}
```

  Mutation (run, killed): `n < want[e]` to `n < 1` (the old per-line wait) fails "one death line
  satisfied two fights".

  (`checkObserverText` is unchanged: bot 0's swing lines name it and "small rat"/"large rat", which
  match its `\brat\b`, and a rat's lines and the death notices are third person to the observer.) In
  `cmd/smoke/main.go`, `runBots` (`:124`), `fc := newFightCheck(n, fmt.Sprintf("Smk%sc000", tag))`
  becomes:

```go
	rats := addrRats // a server someone else started keeps its own rats
	if srv != nil {
		rats = launchRats // the world copy's two pinned rats (fightRats)
	}
	fc := newFightCheck(n, fmt.Sprintf("Smk%sc000", tag), rats)
```

  and in `cmd/smoke/fight_test.go:12`, `newFightCheck(2, "Smkabc0")` becomes
  `newFightCheck(2, "Smkabc0", launchRats)`. In `script.go:195` the comment `// The smoke fight:
  bot 0 attacks the rat in Town Square, bot 1 watches.` becomes `// The smoke fight: bot 0 fights the
  two rats in Town Square (ruling 11), bot 1 watches.`

  **`-addr` mode** (a server someone else started, such as `provisioning/compose.smoke.yml`) gets no
  world patch: bot 0 fights one rat, whichever `attack rat` finds, at whatever level it spawned, and
  accepts either ending. A level 4 or 5 rat there can still run the fight past the 4-minute timeout
  (Q18's flake); CI runs launch mode.

  **Copyover** (Linux launch mode only, `script.go:238-255`): every bot waits online while the admin
  restarts the server, then checks its room and item. It runs after the fights, so bot 0 is back in
  Town Square with its ring whichever way its fights ended, and the pinned rats are templates the
  restarted server reads from the same world copy. Unverified locally: copyover does not run on
  Windows; CI's Linux run is its check.

  **Runs on the scratch copy** (ruling 11's world copy, real rat damage):
  - 1 bot: `PASS 1/1`. Fight 1: `small rat has died.` and the XP line. Fight 2: `Smk...c000 has DIED!`,
    the Waiting room, `portal` back, the ring taken back from its corpse. The later suicide step then
    took the ring from its newest corpse as before.
  - 5 bots: `PASS 5/5`. Fight 1: `small rat has died.`; fight 2: bot 0 died, Waiting room, back,
    ring retrieved. Other bots' later suicide notices (`Smk...c002 has DIED!` and so on) reached bot 0's
    screen during the run; the check names bot 0 (`c000`, fixed-width names), so they did not count.
  - 12 bots: `PASS 12/12`. Same endings: the small rat died, then the large rat killed bot 0.
  In all three runs bot 1 read both endings and its observer check passed. Every fight 1 ended with
  the rat's death and every fight 2 with bot 0's, as the simulation predicts (bot 0 beats the level 1
  rat about 98% of the time, the level 5 rat never). Re-run after plan review round 5 (fresh scratch
  copy, the observer's count-based wait): `PASS 1/1`, `PASS 5/5`, `PASS 12/12`; in each, the small
  rat died and the large rat killed bot 0, who walked back and took the ring. Re-run after plan
  review round 6 (the wait reset, then the told mark): the same each time, `PASS 1/1`, `PASS 5/5`,
  `PASS 12/12`.

- [ ] **Step 13: Per-task check, then commit.**

```bash
git add internal/characters/aggro.go internal/characters/character.go internal/characters/character_test.go \
  internal/death/death.go internal/death/switch.go internal/hooks/NewRound_DoCombat.go \
  internal/hooks/combat_golden_test.go internal/hooks/retaliation_test.go internal/hooks/retaliation_golden_test.go \
  internal/mutators/testhelpers.go \
  internal/mobs/saved_test.go internal/rules/dikubasic/ruleset.go internal/rules/dikubasic/ruleset_test.go \
  internal/usercommands/go.go internal/usercommands/go_test.go internal/usercommands/flee.go \
  internal/usercommands/quit_test.go internal/usercommands/shoot_test.go \
  cmd/smoke/server.go cmd/smoke/fightrats_test.go cmd/smoke/fight.go cmd/smoke/fight_test.go \
  cmd/smoke/main.go cmd/smoke/script.go internal/hooks/death_engine_test.go \
  internal/configs/config.gameplay.go internal/configs/config.gameplay_test.go \
  internal/behaviour/tick.go internal/behaviour/tick_test.go \
  _datafiles/config.yaml _datafiles/html/admin/config.html _datafiles/html/admin/config-wizard.html \
  internal/hooks/testdata/combat-golden/ internal/hooks/testdata/reads-golden/E-move-look.txt
git commit -F - <<'EOF'
Phase 5.4.1: retaliation and target switching (rulings 6-9, 12)

A body attacked while not fighting takes an aggro on its attacker
(after every attack, hit or miss, when it is up and CanAttack allows
it): in the same room a plain attack; from another room a shot back
through the exit with a shooting weapon, otherwise an aggro that waits.
Each turn settles the target: gone or not up, switch to the longest
attacker in the room (or end with today's line), swinging at once; in
another room, shoot if possible (kept even if attacked here), else
switch if attacked here, else wait ("Your target is out of reach."
once). Before it attacks, in the room or through an exit, the ruleset
must still allow it (CanAttack); refused, it switches or ends with the
refusal; the switch picks only opponents CanAttack allows, so two
players with aggros on each other in a room without PvP stop. A switch
re-arms like attack <name>: a slow weapon waits its delay first. A
shot re-arms through the same exit (SetAggroRemote), one AggroChanged
a round. A shot goes through any exit shoot <direction> could use:
the room's, a temporary one or an active mutator's, resolved with
FindExitByName and locked by GetExitInfo (mutators.SetTestSpec added
for tests). A target in reach that the body cannot perceive gives way
to an attacker here it perceives (step 4), else counts toward the
wait limit. Being attacked by its target, hit or miss, resets a
body's wait count, so a hidden attacker or a shooter in the next room
that keeps attacking is never lost track of. Every
body with an aggro settles its target every round, even when no-combat
or an NPC combat command takes its swing. Ties on Since go to turn
order (death.SortBodies, which the round's fighters() uses too). A revived victim's attackers keep fighting
it. The aggro records the round it began (Since, saved in NPC
sidecars; a failed flee keeps it). Death ends and switches the aggro of
every body targeting the victim. Turn order within accounts and within
NPCs is now stable (UserId; MobId, then body ID). A body's own aggro
blocks go, quit and suicide only while its target is in its room; a
pending flee with nobody attacking is not a fight.

Ruling 12: a wait ends after GamePlay.OutOfReachRounds (new; 5 by
default, 0 or less means 5) consecutive turns out of reach, for every
body, NPC targets included; any turn that finds the target reachable
(in the room, or by shooting) resets the count (Aggro.OutOfReach,
saved in NPC sidecars). "Your target is out of reach." comes once per
wait (Aggro.OutOfReachTold, saved with it; cleared only when the
target is found reachable, not by being attacked). The body reads
"You lose track of <name>.",
its room "<name> mumbles about losing their quarry.".
behaviour.loseLostQuarry and its call in Tick are deleted: step 1
covers its offline-player case, the limit its other-room case.

Ruling 10: the PvP settings apply only between two players
(dikubasic.CanAttack). NPCs may attack, shoot and retaliate against
players anywhere except the recovery room, which is safe in both
directions: CanAttack also refuses an attacker standing in it or in a
private copy of it. This reverses slice 2's
"shooting a player obeys the PvP settings for player and NPC shooters
alike" for NPC shooters: TestNPCShootingAPlayerRefusedWhenPvPDisabled
becomes TestNPCShootingAPlayerAllowedWhenPvPDisabled.

Smoke (ruling 11): launch mode's world copy (fightRats) spawns exactly
two rats in Town Square, a "small rat" pinned at level 1 and a "large
rat" pinned at level 5, neither wandering, and none in End of the Line;
rat damage is unchanged. Bot 0 fights the small rat, then the large
one; each fight accepts either ending and checks it (the rat's death
and the XP line, or bot 0's death, the recovery room, the walk back
and the ring taken back from its corpse). Bot 1's observer check waits
for each ending. -addr mode fights one rat, either ending.

Goldens: all 13 combat goldens and reads E-move-look change in kind as
the plan's Task 6a table states; eight new fixture files (the spec's
six, its flee case split into switch and wait, and ruling 12's
wait_runs_out). Decided.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 7: Docs, smoke, review (controller)

**Files:** modify `PATCH_NOTES.md`, `status.md`, `2026-11-06-phase-5-rules-seam-design.md`,
`internal/guards/combat_test.go` (two allowlist reasons);
`docs/known-issues.md` only if the grep below finds something.

- [ ] **Step 1: Whole-branch verification.** The per-task check; `go test ./internal/hooks/ -count=1`
  with no `-update`; and smoke with 1, 5 and 12 bots:

```bash
go run ./cmd/smoke -bots 1
go run ./cmd/smoke -bots 5
go run ./cmd/smoke -bots 12
```

  Report each summary line (unverified at plan time).

- [ ] **Step 2: `PATCH_NOTES.md`.** Under a heading for the execution date, one line:

  `- Phase 5 slice 4.1: round and death leftovers. The engine compares no body's health with a number, in characters and rooms too (guarded): characters asks the ruleset whether a body is up (characters.IsUp, wired by rules.Set). Player-visible: creatures fight back when attacked, as players now do too (a body with no fight of its own takes one on whoever attacks it, hit or miss); when your target dies, flees or leaves, you turn to whoever in the room has been attacking you longest ("You turn to fight <name>!"), and a target you chose yourself stays chosen while it is around; shot from another room, you shoot back through the exit (a portal or other temporary exit too, unless locked) if you hold a bow or other shooting weapon (and keep shooting even if someone in your room attacks you), otherwise you wait ("Your target is out of reach.") and swing once you share a room; a wait lasts 5 rounds out of reach in a row (admin setting GamePlay.OutOfReachRounds; each attack your target makes on you starts the count again), then you lose track of your target ("You lose track of <name>."), and creatures follow the same rule (they used to give up on a player in another room at once, and never on a creature); while you wait you can walk (into the shooter's room, or away), and a creature shot from another room walks over to fight; a creature attacking a player is not PvP, so creatures fight back against players of any level in any room, and can attack or shoot players under any PvP setting, except in the recovery room; NPCs still never start a fight on their own. A creature wounded and left alone, or once its fight ends, heals over time as players do (at least 1 health and 1 mana every 3rd round out of a fight); a shield no longer breaks. Internal: the engine keeps no health floor (damage below 0 is applied as dealt; nothing shows it before the death settles). Removed: offhand breakage (and the admin item editor's Break Chance field; old item files with breakchance still load), the unused backstab aggro, and dead code (Character.IsDisabled, rooms.FindDowned, actors.BodyReplaced)`

- [ ] **Step 3: `status.md`.** "Now": nothing in progress, slice 4.1 merged (PR number). "Next" item 1
  becomes slice 4.2 (progression: levels, XP, alignment, training and stat points, the PvP minimum
  level and their displays) then 4.3 (the body: the stat block, `statmods`, mana, race fields, action
  points and carry capacity, skills, the simple hit roll with weapon damage, which also decides the
  crit cascade). "Watch": the last-run line with today's numbers; the `FleeSuccess` line stays; add (Q18):
  `- Smoke -addr mode (a server someone else started) can time out its fight step: against that
  world's own rats, spawned at level 1 to 5, about 1.7% (level 4) and 2.4% (level 5) of fights run past
  60 rounds, the 4-minute fightStepTimeout at 4-second rounds (simulated). Launch mode, which CI runs,
  no longer has this: since slice 4.1 its world copy pins the rats at levels 1 and 5 (ruling 11), and
  its fights end within 17 rounds. Pre-existing. Fix candidate for -addr mode: raise the timeout`.

- [ ] **Step 4: The phase design.** In `2026-11-06-phase-5-rules-seam-design.md`: slice row 4
  (`:140-145`) gains the cut: "Cut into 4.1 (round and death leftovers, done), 4.2 (progression) and
  4.3 (the body); Calabe 2026-11-12." In "Notes for later slices" (`:182-222`): from the slice 2
  notes, delete the offhand breakage and `BackStab` bullets (`:205-209`), and reword the "Better
  targeting and aggro tracking than the single aggro flag is wanted later (Calabe 2026-11-10), outside
  the slim base." bullet (`:210-211`) to what is left: "Slice 4.1 gave the single aggro record its
  start round (`Since`), retaliation, switching to the longest attacker and a wait limit (rulings 6-9,
  12). Still wanted later, outside the slim base: more than one target per body, threat or aggro
  tracking beyond the longest attacker (Calabe 2026-11-10)."; from the slice 3 notes,
  delete NPC regen, the floors, the four `characters` health checks and `FindDowned`, offhand and
  `BackStab`, and `BodyReplaced` (`:217`, `:218`, `:220`, `:221`, `:222`), keeping XP, levels and
  alignment (`:219`) under a new line naming 4.2; add a "From slice 4.1" note with the 4.2 and 4.3
  lists (spec ruling 1). The 4.3 list names the crit cascade (Q4): "In `calculateCombat`
  (`dikubasic/combat.go`), once a hit crits every later hit that round crits with no roll
  (`isCrit = attackResult.Crit || Crits(...)`); kept through 4.1 with its comment corrected; 4.3's
  hit roll decides crits." `keywords-api.html`'s sample `backstab` help entry stays on slice 7's docs
  sweep.

  Relabel the notes still addressed to "slice 4" that now belong to a sub-slice:
  - `:186` (from slice 1): "Slice 4 deletes `statmods.Casting`/`CastingPrefix` ..., the race field
    `KnowsFirstAid` ..., and the stat help templates" becomes "Slice 4.3 deletes ..." (the stat block,
    `statmods`, race fields and skills are 4.3's, spec ruling 1).
  - `:203` (from slice 2): "(rankings, slice 4)" becomes "(rankings, slice 4.3)". Why 4.3: the three
    ranking functions are built on the body rules 4.3 replaces. `RankWeapons` scores against a
    `baselineCharacter` with six stats at 50 (`dikubasic/weapon_rank.go:57-74`); `RankArmor` scores
    `statmods` by stat (`armor_rank.go:11`, `:112-141`); `RankMobs` builds a stat baseline from the
    level (six stats at `level * 2`, kept within 1 to 100) and runs `newSimMob` (`mob_rank.go:52-62`, `:132`). Level is an input to `RankMobs` (4.2),
    but the rankings stand or fall with the stat block and the hit roll, so they go or change in 4.3.
  - `internal/guards/combat_test.go:122-123`, the allowlist reasons
    `"weapon and armour rankings; slice 4 shrink"` and `"mob rankings; slice 4 shrink"`, become
    `"weapon and armour rankings; slice 4.3 shrink"` and `"mob rankings; slice 4.3 shrink"` (same
    reason as above).

  Add a note for later slices: "NPC aggression (NPCs starting fights) is still out of scope after
  slice 4.1: NPCs now fight back when attacked (rulings 6-9) but never attack first; it comes with a
  ruleset or a later phase (spec 'What this locks in')."

  Add to the 4.2 list (kill stats): "Since slice 4.1 a player's retaliation that misses still credits
  a 0 in the attacker's `PlayerDamage` (`attack` tracks every swing, `NewRound_DoCombat.go:80-83`), so
  when that attacker later dies to an NPC, `recordPlayerDeath` books it as a PvP death because its
  `PlayerDamage` is not empty (`death/death.go:136-137`; plan review round 2). Slice 4.1 keeps this on
  purpose (spec, Decisions row "Retaliation records the retaliator's damage like any attack"); settle
  it with the kill and death books."

  Add to the 4.3 list: "`rules.actorOf` (slice 4.1) gives any body that is not a live actor's own
  pointer a detached actor with no controller: a corpse's copy of a body, and a by-value copy such as
  the template copies `SimulateCombat` fights (`dikubasic/simulate.go:103-114`, which call
  `ApplyHealthChange`, so `ForgetKillerIfUp`, so `IsUp`). Under `dikubasic` this gives the same answer
  (its `Condition` reads only health). A ruleset whose `Condition` reads the controller would misjudge
  those copies. `calculateCombat` also works on by-value copies (`combat.go:146`) and passes
  `&sourceChar`/`&targetChar` to `buildCombatMessages`, but no call there reaches `IsUp` today (its
  names come from message tokens); a later change that formats a name from those copies would."

- [ ] **Step 5: Known issues.** `grep -n -i "break\|backstab\|downed\|floor\|regenerat\|IsDisabled\|FindDowned\|BodyReplaced" docs/known-issues.md`.
  At `1c8e607f` it finds only line 23 ("Live floor items"), unrelated: no change expected.

- [ ] **Step 6: Commit** the docs and the relabelled guard reasons
  (`git add PATCH_NOTES.md status.md 2026-11-06-phase-5-rules-seam-design.md internal/guards/combat_test.go`;
  run `go test ./internal/guards/ -count=1` first, `ok`), message
  `Phase 5.4.1: docs (patch notes, status, phase design, guard reasons)`, ending with the
  `Co-Authored-By` line. Then the whole-branch review and the PR (`geomud-git`). The spec and plan move to
  `docs/completed/` when the PR merges, as slice 3's did.
