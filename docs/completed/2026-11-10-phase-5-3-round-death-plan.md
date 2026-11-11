# Phase 5 slice 3: round, regen and death through the seam (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved spec `2026-11-10-phase-5-3-round-death-spec.md`: the ruleset says who is
up, downed or dead (`Condition`), how much a body regenerates (`Regen`) and what a death means
(`OnDeath`); every actor dies through one engine routine, `death.Die`; round and regen move with
parity; death goes straight to the bare rules, with a golden of today's death recorded first.

**Architecture:** See the spec ("Design", and "Settled while planning"). This plan adds no design
decision of its own: the questions found while planning are below, the controller accepted every
recommendation, and the spec records each; the calls made while applying review round 2 are in the
spec's "Decisions taken without asking". Code in the tasks that follows one is marked "(Q*n*)".

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
`Smk%sc%03d`), so no bot name is a prefix of another. (Still needed after Task 5b: a name still
reaches every corpse whose name it begins, so `Smk<tag>c1` would count `Smk<tag>c10`'s corpse and
body too. After Task 5b the bot still names its corpse by its own name: the one typing is not in its
own name's count; Task 6 Step 12.)

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

## Plan review round 2: what this revision changed

The spec's "Plan review round 2" items, each applied here (the spec marks each with this place):

- **Blocker, the corpse panel drops its rows** (`look.panels.go:112-121`; no
  `panel-layouts/character/equipment-look.yaml` exists, so the fallback always runs): Task 6 Step 7
  makes the fallback create `equip`; `TestLookAtACorpseShowsWhatItHolds` (Step 11b) is its failing
  test and gains a probe.
- **Calabe's ruling, `bury` refuses a player's corpse that holds anything**: new **Task 5c**.
- **Calabe's rulings on corpse names** (fix the lookup; then, replacing the two-list version, one
  numbering per name with the living first and corpses newest first; and both ordinal forms, `2.rat`
  and `rat#2`, in the one parser `util.GetMatchNumber`): new **Task 5b** (the parser,
  `util.FindMatchIndexIn`, `FindCorpseByRef` with `livingReach`, `FindCorpse` delegating, the
  "still alive" refusal in `get` and `bury`, `bury` removing the corpse it found); the `get` caller
  test rides in Task 6 Step 11b, where `get` first reaches corpses. Then two more rulings: the one
  typing is not in their own name's living count for a corpse (Task 5b, with `look bob` by Bob
  reaching his corpse), and a player's refusal line takes no "The"; and a hidden body is in no
  viewer's name list, for every name lookup; then admins always perceive the hidden, and one
  visibility rule everywhere (new **Task 5a**: `rooms.Perceives`, the room view's check plus the
  admin reveal, now asked by the room view, `FindByName` for all 12 callers, the corpse count,
  messaging's `canSee` through `PerceivesWith`, GMCP's room contents and combat's "can't see them";
  the sneaking actors' own silences, darkness and the unread `Unseen` field are left alone, with the
  reasons in Task 5a and the spec). Goldens: none change in 5a, 5b or 5c (see each task); Task 6
  still changes exactly nine.
- **Major, `-count=16` over the `TestCombatGoldenFleeSuccess` repeat flake**: Task 6 Step 11b(c) now
  runs only the tests that make deaths, one package at a time.
- **Minor, `remove.go` probes**: Task 3 Step 9(b) probes keep a `rules` use so they compile.
- **Minor, Task 6 Step 8's absence grep**: moved to Step 13 (after Steps 11 and 12 clear their hits);
  its control is now a new Step 0 (129 lines in 19 files before the task).
- **Minor, Task 5's grep list**: names `ruleset.go`'s own `recoveryRegen` comment.
- **Minor, surviving mutations**, each now killed by a named test and probe: respawn health and NPC
  respawn (`TestStubRulesetDecidesTheRespawn`, Task 9); the gold clamp and the "item not held" guard
  (`TestStubRulesetCannotDropWhatTheBodyLacks`, Task 9); `PlayerDeath`'s killer fields and the death
  event-log line (`TestRespawnStartsClean`, Task 6 Step 11b); `CleanupMobSpawns`
  (`TestDeathFreesTheNPCsSpawnSlot`, Task 6 Step 11b); NPC regen in the recovery room and the
  recovery line's `Senses`/`Range` (Task 5); container precedence in `namesACorpse`
  (`TestGetAllWithoutACorpseTakesTheFloor`, Task 6 Step 11b).
- **Minor, the death guard's scope**: Task 8 checks each prefix reaches a file.
- **Minor, two fixture comments**: Task 6 Step 11 gives their new text.
- **Minor, `suicide` staying `AllowedWhenDowned`**: Task 7 Step 3b words the registry comment; the
  spec's design says it in one sentence.
- **Minor, `RegenResult.Changes()` and the `IsDisabled` ban**: in the spec's "Decisions taken without
  asking".

## Plan review round 3: what this revision changed

- **M1, `Char.Enemies`** and **M2, a say's `Comm.Channel` sender**: Task 5a's table rows, Step 6's
  round 3 hunks, tests in `modules/gmcp/visibility_test.go`.
- **The site search redone** (every path that hands a viewer a body's name, ID or presence, not only
  the places that named hidden): Task 5a's table now lists each with file:line, moved or left alone
  with a reason. Besides M1 and M2 it moved GMCP's leave notice and room refreshes, untargeted
  `attack`, the walk-in "notices you" line and the follow lines.
- **Calabe's follow ruling** (following does not cancel hiding; a line whose only content is a hidden
  body's presence is dropped for a viewer who does not perceive it, while a line that carries news the
  viewer gets anyway keeps "Someone"): the target's "is following you" and "stopped following you"
  are dropped (`TestHiddenFollowerIsNotAnnounced`, two probes); the follower's own "You are no longer
  following someone" is kept; a say's GMCP `Sender: "Someone"` is kept (the say text arrives anyway,
  as "Someone says"). Every other round 3 site already drops (`Char.Enemies`, the leave notice and
  refreshes, the walk-in notice, untargeted picks).
- **m1** (no test reached the room view or GMCP's room contents): `TestRoomViewListsOnlyWhatTheViewerPerceives`
  and `TestRoomContentsListOnlyWhatTheViewerPerceives`, each with a probe (Task 5a Step 7).
- **m2** (untargeted `attack` picks a hidden body): Task 5a, `TestUntargetedAttackPicksOnlyWhatTheAttackerPerceives`.
- **m3** (Task 6 Step 13's knob grep hits a comment Step 11b writes): the look test's comment reworded.
- **m4** (the corpse lookup no longer needs `UserId > 0` or `MobId > 0`): the look test's stale comment
  reworded; the call is in the spec's decisions table.
- **m5** (the sort probe cannot kill the same-round bury test): Task 5b Step 9's prediction fixed.
- **m6** (Task 1 Step 12's predictions): `hp=1/7` in `death_player` and `death_pvp`, no Remove lines in
  `death_player`.
- **m7** (the deferred silences ruling said `bury`; it is `trash`): fixed in the spec.
- **m8** (help text): `get.md` says what a name reaches, Bobby and the no-corpse case included; the
  `item-names.template` anchor names its line.

## Plan review round 4: what this revision changed

- **M1** `follow lose` uses `seenName`: Task 5a table and Step 6 hunk, `TestFollowLoseNamesOnlyAPerceivedLeader`.
- **M2** presence-only lines dropped for viewers who do not perceive the body (`rooms.NotPerceiving`,
  Task 5a Step 3, passed as `Exclude`): Task 5a's "Presence-only lines" table lists each line and its
  verdict; Step 6 hunks and the recovery-line bullet; Task 6 Step 3's revive;
  `TestPresenceLinesReachOnlyThoseWhoPerceive`.
- **Minors:** Task 5b's `RemoveCorpseRef` test looks up `rat`; Task 5a Step 7's always-true probe
  names the two goldens it changes; tests and probes for `aggroChangedHandler`, the old-room loop and
  the NPC-follower and NPC-leader follow lines (`TestFollowExpiryFollowsTheRule`); `seenName`'s
  comment placed above it; distinct IDs in `TestPerceivesAnUnhiddenBodyAndItself`.

## Before plan review round 5: what this revision changed

- **Every moved presence line has its own test and probe.** Task 5a Step 1:
  `TestEachPresenceLineReachesOnlyThoseWhoPerceive`, one subtest per call site (teleport, the
  teleport party loop, flee, drop), and `TestRecoveryLineExcludesThoseWhoDoNotPerceive` in
  `dikubasic`; Step 7 probes delete each site's `Exclude` in turn. Task 6 Step 11b:
  `TestReviveOfAHiddenBodyReachesOnlyThoseWhoPerceive`, with its probe in the revive bullet.
- **Run:** the four hooks subtests pass on a scratch worktree with Task 5a applied, and each probe
  fails its own subtest only. Task 5b's `RemoveCorpseRef` test passes with 5b applied and fails
  under its probe. The recovery test ran against a stand-in (Tasks 3 and 5 not applied); the revive
  test is unrun (the `death` package is Task 6's).

## Plan review round 5: what this revision changed

Round 5 ran the whole plan, Tasks 1 to 10: everything builds and the nine predicted goldens hold.
M1 is the controller's application of Calabe's principle ("what a viewer cannot perceive does not
exist for them; no reply reveals it"), not a Calabe ruling; he may overrule.

- **M1** the world death notice named a hidden killer to everyone. Task 6 Step 3's
  `recordPlayerDeath` now names, per reader, only the killers that reader perceives (`namesKiller`:
  `rooms.Perceives`, and a hidden killer in another room only to an admin); with none left the
  `(killed by ...)` part goes, never "someone". The notice was one `events.Broadcast` string; Step 3a
  adds `messaging.NotifyEach`, one `UserId` message per active account. Test
  `TestDeathNoticeNamesOnlyTheKillersEachReaderPerceives` (Task 6 Step 11b) with three probes; the
  `hidden_attacker` prediction (Step 16) now has only Alice reading `(killed by Alice)`.
- **Minor 8 not applied:** a fight's victim already perceives the fight through the attacker's hit
  lines, so the refusal reveals nothing; counting only perceived attackers would let quit escape an
  unseen attacker. `inFight` counts every attacker, as today (site table row).
- **Minors:** Task 5 Step 4's `Range` probe widens the range (deleting the line cannot fail:
  `Room` is the zero value); `death_player`'s testsword remove, drop and `ItemOwnership Alice item=93101`
  in Task 1 Step 12 and Task 6 Step 16; `get.md`'s stash line gains two trailing spaces (rendered
  through `markdown.NewParser`); the Task 3 Step 8 and Task 6 Step 11b(b) probes compile (`_ =`);
  `bury.template`'s lines go after line 7; the presence table's three stray site rows moved back to
  the site table; `DeathOutcome.Messages` dropped (Q1: awards carry the lines).
- **Run on a scratch copy at `65bbc144`:** today's `suicide.go` sending the notice through
  `NotifyEach` with `namesKiller` passes the notice test written for the reads world, its three
  probes fail as stated, and `go test ./... -count=1` shows only the 3 known Windows failures
  (goldens unchanged). The `get.md` render was run. Task 6's form of the notice test is unrun (the
  `death` package is Task 6's).

## Where the spec's facts table is wrong or short (checked at `225c08ba`)

- **Help pages** live at `_datafiles/world/empty/templates/help/` (`death.template`, `health.template`,
  `get.md`), not `templates/help/`. `keywords.yaml` is `_datafiles/world/empty/keywords.yaml` (`:18`
  correct). Line ends (rechecked at `65bbc144`, `git ls-files --eol`): `death.template`,
  `health.template`, `get.md`, `config.yaml`, `75.yaml` and `bury.template` are all LF in git and CRLF
  in this checkout (`core.autocrlf=true`), so the working-tree edits below keep CRLF and git stores
  LF. (The earlier "`health.template` and `get.md` LF" was wrong.)
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
  `:232-238`. Missing: `users.html:366-370`, `:592`, `:1089` (the Extra Lives field in the user editor).
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
| `UpdateCorpses`; `FindCorpseByRef` | `rooms.go:231-258`; `:945-993` (before Task 5b rewrites it) |
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

Added for plan review round 2 (verified 2026-11-11 at `65bbc144`):

| What | Where |
|---|---|
| `GetMatchNumber`; `FindMatchIn`; `stringMatch` | `internal/util/util.go:298-314`; `:316-368`; `:373-397` |
| `FindMatchIn` tests | `util_test.go:533-561` (`TestFindMatchIn`) |
| `RemoveCorpse` (matches owner, name and round; first wins) | `rooms.go:209-229` |
| `FindCorpse`; `FindCorpseByRef` (each de-duplicates names); `FindContainerByName` | `rooms.go:893-943`; `:945-993`; `:1464-1482` |
| Callers of the two lookups (all of them: `grep -rn 'FindCorpse' --include=*.go .`) | `get.go:92` (`ByRef`), `get.go:422`, `look.go:394`, `modules/cleanup/cleanup.go:141`; tests `rooms_test.go:178-214` |
| Living lookup: `FindByName`; `findPlayerByName`; `findMobByName` (its `#`/`@` guard; its `name#k` list) | `rooms.go:1334-1341`; `:1343-1387`; `:1389-1435` (`:1391-1397`; `:1417`) |
| `FindByName` callers (12, `grep -rn '\.FindByName(' --include=*.go . \| grep -v _test`); each takes the player when both come back | e.g. `look.go:76-84`, `consider.go:23-40`, `attack.go:164` |
| `GetMatchNumber` callers | `util.go:322` (in `FindMatchIn`), `items/items.go:604`; tests `util_test.go:509-531` |
| `Corpse.RoundCreated` | `rooms/corpse.go:13` (`uint64`) |
| Character names: shipped `NameRejectRegex` | `_datafiles/config.yaml:578` (`^[a-zA-Z0-9_]+$`), read by `accounts.ValidateName` (`users.go:517`) |
| Help on ordinals | `_datafiles/world/empty/templates/help/item-names.template:13-15` (the `#N` form) |
| Hidden and its reveal: buff flags `hidden`, `see-hidden` | `_datafiles/world/empty/buffs-flags/hidden.yaml`, `see-hidden.yaml`; the only `see-hidden` readers are `rooms/roomdetails.go:253`, `:269` |
| Room view hides a body per viewer | `roomdetails.go:252-256` (players), `:268-272` (NPCs) |
| Other hidden checks with no reveal (moved onto `Perceives` in Task 5a) | `modules/gmcp/gmcp.Room.go:346`, `:374`; `messaging/perceive.go:43-55` (`canSee`); `hooks/NewRound_DoCombat.go:195-199` |
| Admin role and the permission test that lets it through | `accounts/userrecord.go:27` (`RoleAdmin`), `:386-389` (`HasRolePermission`); admin commands gated at `usercommands.go:326` |
| Messaging's flag seam (its tests fake `hidden`) | `messaging/perceive.go:28-31`; `messaging_test.go:42` |
| `FindAll` returns every body unfiltered | `GetMobs` `rooms.go:1083-1085`; `GetPlayers` `:1175-1177` |
| Self checks after `FindByName` | `consider.go:24-26`, `attack.go:167-172`; none in `look`, `give`, `show` |
| Global name lookups (not of a room's bodies) | `whisper.go:34`, `party.go:104`, `admin.locate.go:34`, `admin.teleport.go:47`, `:77`, `admin.visit.go:60` (`accounts.GetByCharacterName`) |
| `look.go`'s unused third copy of the de-duplicated name lists | `look.go:367-392` (built, never read) |
| `bury` | `modules/cleanup/cleanup.go:132-168` (lookup `:141`, `RemoveCorpse` `:143`); registered `:58`; help `modules/cleanup/files/datafiles/templates/help/bury.template` (8 lines) |
| The reads harness runs module commands | `reads_golden_test.go:72` (imports `modules/cleanup`); `newReadsWorld` `:256`, `trio` `:403`, `run` `:456` |
| Corpse panel fallback names `inv`, rows go to `equip`; the equipment panel's fallback names `equip` | `look.panels.go:112-121`; `:55-61` (`buildInventoryLookPanel`). `_datafiles/world/empty/panel-layouts/` holds only `character/inventory.yaml`, so both fallbacks always run |
| NPC ending: home-room slot settling | `rooms.go:668-676` (`CleanupMobSpawns`); `rooms/npcstore.go:492-558` (`slotNPC`, `settleSlot`) |
| `PlayerDeath` fields; `UserLog` | `events/eventtypes.go:220-228`; `accounts/userlog.go:14-21` (`Category`, `What`) |
| `TestCombatGoldenFleeSuccess` repeat flake | `combat_golden_test.go:468`; `status.md:24-25` (Watch) |

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
  "has DIED!" (with no `\r` at the line end), the Drop lines (a Remove line comes before each worn
  item a roll drops; at seed 303 `death_player` has one: `You remove your <ansi fg="item">testsword</ansi> and return it to your backpack.`
  then `You drop the <ansi fg="item">testsword</ansi>.` and the note `ItemOwnership Alice item=93101`,
  plan review round 5's run; round 3's "has none" was wrong), "You lost 100 experience points", the corpse, `room=copy-of-93003 hp=1/7` (round
  3's run: the respawned level 2 body's maximum is 7, not the 12 `addPlayer` set: report the value
  the record holds and do not change the fixture), `kd=0/1/0/0/0/1` (a mob death, an elite death).
  `death_pvp`: the same for Bob (`hp=1/7` too, round 3's run) with "(killed by Alice)", Bob's `kd=0/0/0/1/0/0 pk=[] pd=[9301:Alice=1]`
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
  line replaced by `_, _ = rules.Get, dikubasic.New // probe: no ruleset set` (deleting it alone
  leaves both imports unused, and the package does not compile) it panics with
  `rules: no ruleset is set` (paste the line, restore).

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
  (`skills/skill.go:22`, `:116-121`); a new character has no skills. Probes (each keeps a `rules.`
  call, the file's only one, so the import stays used and the probe compiles): at `remove.go:51`
  replace `rules.Get().Condition(actor) == rules.Up` with `rules.Get() != nil` (always true: the
  not-up assert fails), then with `rules.Get() == nil` (always false: the up assert fails). Paste the
  two lines, restore.

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

	// Any actor there heals, an NPC too (Q7h; outside it, accounts only).
	n := npc(t, 0, 0)
	n.Character.RoomId = copyId
	n.Character.HealthMax.Value, n.Character.ManaMax.Value = 100, 100
	n.Character.Health, n.Character.Mana = 1, 0
	util.SeedRandForTest(1)
	require.True(t, New().Regen(n, 1).Changes(), "an NPC in the recovery room heals")
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
	assert.Equal(t, perception.Sight, got.Messages[1].Senses, "the room sees it")
	assert.Equal(t, perception.Room, got.Messages[1].Range, "the room, not further")
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
    compiles; the d10 test fails, first at its full-health assert, `ruleset_test.go:298`);
  - drop the mana cap only: `mana := util.RollDice(1, 10)` (the health line still uses `c`; the
    d10 test fails the same way, first at its full-health assert);
  - add `if c.Health >= c.HealthMax.Value && c.Mana >= c.ManaMax.Value { return rules.RegenResult{} }`
    as the first statement after `c := a.Character` (the full-health "still two draws" assert fails);
  - change `{actor} is recovering` to `{actor} recovers` (the line test fails);
  - make the recovery rule accounts only, `if accountOf(a) != nil && rooms.IsDeathRecoveryRoom(a.Character.RoomId) {`
    (the d10 test's NPC assert fails);
  - delete the room line's `Senses:` line (the `Senses` assert fails); set its `Range:` to
    `perception.Nearby` (the `Range` assert fails). Deleting the `Range:` line is no probe: `Room` is
    `Range`'s zero value (`perception.go:49`), so the line could go and every test still passes
    (plan review round 5).

  Paste the nine lines.

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
    `internal/rules/dikubasic/ruleset.go` (this task's own `recoveryRegen` comment, "deleted Death
    Recovery buff (24-death_recovery.js)", which names the deleted script on purpose),
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

### Task 5a: One visibility rule (`rooms.Perceives`) for the room view, name lookups, messages, GMCP and combat (Sonnet)

Calabe's rulings (spec, plan review round 2, 2026-11-11):

- **A hidden character is not perceivable until revealed to the viewer:** it is not in the list the
  viewer's names are matched against, takes no ordinal, cannot be named, and no reply reveals it.
  Every name lookup, living and corpse.
- **Admins always perceive hidden characters,** in the one shared check, not per command.
- **One visibility rule everywhere:** every place that decides whether a viewer perceives a
  character asks the same function, so `see-hidden` and admins reveal hidden characters in GMCP and
  in messages too.

**How hidden, reveal and admin work today (read at `65bbc144`).** Hidden is the buff flag `hidden`
on the body (`_datafiles/world/empty/buffs-flags/hidden.yaml`). Reveal to a viewer is the viewer's
buff flag `see-hidden` (`buffs-flags/see-hidden.yaml`); its only readers are `roomdetails.go:253`
and `:269`. There is no per-viewer reveal list. An admin is an account whose `Role` is
`accounts.RoleAdmin` (`accounts/userrecord.go:27`, `RoleAdmin string = "admin"`); the codebase's own
permission test lets it through first, `if u.Role == RoleAdmin { return true }`
(`HasRolePermission`, `userrecord.go:386-389`), and admin commands are gated through that
(`usercommands.go:326`). A body's account is `actors.AccountOf(id)` (`actors/registry.go:70`), which
`rooms` already uses (`findPlayerByName`, `rooms.go:1353`).

**Every path that hands a viewer a character's name, ID or presence** (plan review round 3: a grep
for hidden-related words cannot find a path that never checked hidden, so the search also covers
every place that enumerates a room's bodies, every GMCP module and payload, every
`events.Communication` producer and consumer, every untargeted target pick and every line that
names one body to another). The searches, at `65bbc144`:
`grep -rn '"hidden"\|`hidden`\|see-hidden\|nightvision\|GetVisibility()\|Unseen\|invisib' --include=*.go internal modules | grep -v _test`
(nothing named invisible or blind exists);
`grep -rln 'GetPlayers(\|GetMobs(\|GetBodies(\|playerIds()\|mobIds()' --include=*.go internal modules *.go | grep -v _test`
(15 files, each read);
`grep -n 'GetMobs\|GetPlayers\|\.Name\b\|Id.String()\|GetAllActiveUsers\|GetOnlineUserIds' modules/gmcp/*.go`;
`grep -rn 'events.Communication{' --include=*.go internal modules` (producers `say.go:40`,
`broadcast.go:28`, `party.go:609`, `whisper.go:60`; one consumer, `gmcp.Comm.go:32`);
`grep -rn 'messaging.Tell(\|messaging.ToAccount(' --include=*.go internal modules` and
`grep -rnE '(ObserverText|TargetText|TargetRoomText|FarText):.*Character\.Name' --include=*.go internal modules`
(lines that name a body to someone other than the actor).

| Where | What it checks today | This task |
|---|---|---|
| `rooms/roomdetails.go:252-256`, `:268-272` (room view) | hidden unless the viewer has `see-hidden` | calls `Perceives`; gains the admin reveal |
| `rooms.go:1334-1435` (`FindByName`, `findPlayerByName`, `findMobByName`) | nothing: every body, hidden or not, by name or ID | filters by `Perceives(viewer, ·)` |
| `modules/gmcp/gmcp.Room.go:346` (players), `:374` (NPCs) in `GetRoomNode(user, ...)` (`:267`) | hidden from everyone, no reveal | calls `Perceives(user.Character, ·)`; gains `see-hidden` and admin |
| `messaging/perceive.go:43-55` (`canSee`: whether a receiver reads an actor's or target's name or "someone") | self always; far: no one; hidden: no one, no reveal; else light (`canSeeIn`: room visibility or `nightvision`) | the hidden clause calls `rooms.PerceivesWith(hasFlag, ...)`; far and light stay as they are |
| `hooks/NewRound_DoCombat.go:195-199` ("Can't see them, can't fight them") | a hidden defender cannot be fought by anyone | calls `Perceives(attacker, defender)`; an admin or `see-hidden` attacker can fight one |
| Sneaking actors' own lines: `go.go:197`, `:214`, `:273`, `:330`; `look.go:56`, `:91`, `:122`, `:256`, `:293`, `:332`; `read.go:39`; `modules/cleanup/cleanup.go:100` | a hidden actor's action is not announced to the room at all (no message is sent) | **left alone** (spec, decisions): this decides whether a line exists, not who perceives the actor; sending it through messaging would show every other viewer "Someone ...", which widens what players perceive |
| Darkness: `look.go:28-31`, `:239` (the looker's own room view and looking through an exit); `messaging/perceive.go:33-39` (`canSeeIn`) | light or `nightvision` | **left alone**: a rule about the viewer's light, not about the character; `Perceives` is the hidden rule only, and `canSee` keeps light and distance (stricter, sense-based) |
| `RoomChange.Unseen` set at `rooms/roommanager.go:429`, `rooms.go:691` | hidden mover | **left alone**: no reader anywhere (`grep -rn '\.Unseen'` prints nothing) and no viewer |
| `characters/character.go:435`, `formattedname.go:28` (the "hidden" adjective), `accounts/userrecord.prompt.go:244` (`{h}` in one's own prompt) | how a hidden body is drawn, once perceived; one's own state | **left alone**: not a perception decision |
| GMCP room contents and name lookups in a dark room | list or match bodies regardless of light | **left alone**: adding light would narrow what clients and commands see today, beyond these rulings |
| `modules/gmcp/gmcp.Char.go:455-466` (`Char.Enemies`: every NPC fighting in the viewer's room, with Id and Name) | nothing | **moved** (round 3, M1): a hidden enemy is listed only to a viewer who perceives it |
| `modules/gmcp/gmcp.Comm.go:48-50`, `:80-83`, `:105-117` (`Comm.Channel` for a `say`, sent to every player in the speaker's room with `Sender` = the speaker's name and `Source` player or mob) | nothing: names a hidden speaker the text shows as "Someone" | **moved** (round 3, M2): per receiver, `Sender` is the name only when the receiver perceives the speaker; otherwise `Someone`, as messaging shows it, and `Source` is empty |
| `gmcp.Comm.go` for `party`, `broadcast`, `whisper` | the channel's own audience | **left alone**: not room presence; the speaker addresses the audience by name on purpose (the text names them too: `party.go:603`, `whisper.go:57`, broadcast) |
| `modules/gmcp/gmcp.Room.go:112-152` (`despawnHandler`: `Room.RemovePlayer "<name>"` to every player in the room when a player leaves the game; produced at `world.go:188`, `:203`, `:406`) | nothing | **moved**: only to a viewer who perceives the leaver |
| `gmcp.Room.go:154-215` (`roomChangeHandler`) and `:85-110` (`aggroChangedHandler`): a contents refresh to every player in the room when a body moves or changes its fight | nothing: the refresh itself tells a client something moved | **moved**: no refresh to a viewer who does not perceive the body (`perceivesBody`) |
| `gmcp.Room.go:57-83` (`itemOwnershipHandler`) | an items refresh | **left alone**: about items, names no body |
| `modules/gmcp/gmcp.Game.go:69-83` (`Game.Who`: every active player) and `usercommands/online.go:45` | global online lists | **left alone**: not scoped to a room; hidden is presence in a room |
| `modules/gmcp/gmcp.Party.go:235-274` | party members' names and vitals | **left alone**: party membership, not room presence |
| `usercommands/attack.go:48-150` (untargeted `attack`: whoever is fighting you or your party, `*`, `*mob`, `*user`) | picks hidden bodies; the attacker then reads "You can't seem to find your target." | **moved** (round 3, m2): every pick skips a body the attacker does not perceive (`othersOf`, `perceivedPlayers`) |
| `usercommands/go.go:280-300` (walking in: "<NPC> notices you as you enter!" for each hostile NPC) | names a hidden hostile NPC | **moved**: only an NPC the mover perceives |
| `modules/follow/follow.go:203-222`, `:465-475`, `:501` (follow lines: "X is following you", "X stopped following you", "You are no longer following X") | name a hidden follower or leader | **moved** (Calabe, 2026-11-11: following does not cancel hiding): the target's lines, "X is following you" and "X stopped following you", carry nothing but the follower's presence, so a target who does not perceive the follower gets **no line**; the follower's own "You are no longer following X" carries news it gets anyway (its follow ended), so it is kept, with "someone" for a leader it does not perceive (`seenName`) |
| `usercommands/give.go:70`, `show.go:48` (lines naming the giver to the receiver) | the giver's name | **left alone**: both cancel the giver's hidden first, so the giver is seen |
| `usercommands/party.go:123` (`party invite`: "X invited you"), `whisper.go:57` | the sender's name | **left alone**: the sender addresses the receiver by name on purpose, as on the party and whisper channels |
| `modules/follow/follow.go:436-452` (`follow lose`: "You are no longer following X" to each follower; also run by a followed player's teleport, `:296`) | names the leader with its real name | **moved** (round 4, M1): the follower's own line, kept, with "someone" for a leader it does not perceive (`seenName`) |
| `usercommands/who.go:18` | `GetDetails` | covered by the room view row |
| `usercommands/go.go:353-370` (`inFight`: the actor has an aggro target, or a body in its room targets it), read by `go.go:25` ("You can't do that! You are in combat!"), `quit.go:18` ("You're too busy to quit right now!") and, from Task 7, `suicide` ("You're too busy to die right now!") | counts every attacker in the room, hidden or not | **left alone** (round 5, minor 8 not applied): a fight's victim already perceives the fight through the attacker's hit lines ("Someone stabs you..."), so the refusal reveals nothing; counting only perceived attackers would let `quit` escape an unseen attacker |
| `hooks/NewRound_DoCombat.go:414` (flee blockers), `rooms.go:1236-1270` (`ArePlayersAttacking`, `AreMobsAttacking`) | whether a body fights the actor | **left alone**: game mechanics, no name or ID reaches anyone; lines they cause go through messaging |
| `messaging/messaging.go`, `wire.go:54` (who receives a message) | receivers | **left alone**: delivery; what each receiver reads is `canSee`'s, moved above |
| `rooms.go:194`, `:382`, `:406`, `:430`, `:1880`; `roommanager.go:198`, `:218`, `:495`, `:527`; `NewTurn_PruneBuffs.go:32`; `admin.teleport.go:122`; `web/api_v1_mobs.go` | light, buffs, spawns, cleanup, admin tools | **left alone**: no viewer is told of a body |

**Presence-only lines** (round 4, M2: every room-scoped line or template that names an actor for an
arrival, departure, login, logout, spawn, death, recovery or wake, tested against Calabe's
principle: a line whose only content is a hidden body's presence is dropped for a viewer who does
not perceive it, `Exclude: rooms.NotPerceiving(body)`; a line that carries news the viewer gets
anyway keeps messaging's "Someone"). Searched with
`grep -rn 'ObserverText' --include=*.go internal modules | grep -v _test` (46 in `usercommands`, each
read) and the templates under `_datafiles/world/empty/templates/` (`player-despawn.template` is the
only room line among them):

| Line | Verdict |
|---|---|
| `hooks/PlayerDespawn_HandleLeave.go:53-60`, template `player-despawn` ("... drag {actor} away!"), on leaving the game | **drop** |
| `usercommands/quit.go:23-30` ("{actor} sits down and begins to meditate."), logging out | **drop** |
| `usercommands/admin.teleport.go:109-115`, `:136-142` ("{actor} appears in a flash of light!") | **drop** |
| `hooks/NewRound_DoCombat.go:452-459` ("{actor} flees to the X exit!"), a departure | **drop** |
| `hooks/PlayerDrop_HandlePlayerDrop.go:37-45` ("{actor} drops to the ground!"), becoming downed | **drop** |
| `dikubasic` `recoveryRegen` (Task 5: "{actor} is recovering from a recent death.") | **drop** (Step 6) |
| `death.Die`'s revive line (Task 6: "{actor} is suddenly revived in a shower of sparks!") | **drop** (Task 6 Step 3) |
| `death.Die`'s "{actor} has died." (Task 6) | **keep "Someone"**: the corpse it leaves (named after the dead) is news the viewer gets anyway |
| `go.go:203`, `:231`, `:337` (leave, arrive, bump) | **already silenced** for a hidden mover (the deferred sneaking silences) |
| `start.go:211-219` ("{actor} enters from somewhere."), a new character's first arrival | **left alone**: a character just made carries no buffs, so it cannot be hidden |
| Login (`PlayerSpawn_HandleJoin.go`, `world.go` `enterWorld`), spawns (`SpawnInfo.Message` is read nowhere), wake (no sleep or wake command exists) | **no room line exists** |
| The world notice "*** X has DIED! ***" (Task 6 `recordPlayerDeath`) | **sent to all, as today**: a global notice of the dead by name, not room presence. Its "(killed by ...)" part names, per receiver, only the killers that receiver perceives (round 5, M1; Task 6 Step 3) |

Six sites change in production code (the five files above and the recovery line) plus the revive in
Task 6's new code: within the round's bound of about 8.

**No import cycle.** `rooms` imports neither `messaging`, `hooks`, `usercommands` nor `modules/gmcp`
(`go list -deps ./internal/rooms` lists none of them), while `messaging` (`perceive.go:7`), `hooks`
(`NewRound_DoCombat.go:19`) and `modules/gmcp` (`gmcp.Room.go:15`) already import `rooms`. The admin
test needs `actors` and `accounts`, both already imported by `rooms.go` (`:10-11`). So every caller
reaches the one function by an import it already has.

**Messaging's test seam.** `messaging` reads buff flags through a package seam, `hasFlag`
(`perceive.go:28-31`), which its tests replace (`messaging_test.go:42`) to fake `hidden` without
buff specs. So the rule is `rooms.PerceivesWith(flag, viewer, subject)`, taking the flag reader;
`rooms.Perceives` is it with the bodies' own `HasBuffFlag`, and `canSee` passes `hasFlag`. One rule,
one body of code.

**Living lookups do not filter today.** `FindByName` (`rooms.go:1334-1341`) passes `FindAll` to
`GetPlayers` (`:1158`) and `GetMobs` (`:1065`), which then return every body in the room
(`:1083-1085`, `:1175-1177`), hidden or not, and `findPlayerByName` / `findMobByName`
(`:1343-1387`, `:1389-1435`) match every name, and every ID token, they are given. So today
`look rat`, `attack rat` and the rest reach a hidden rat.

**What changes for each `FindByName` caller** (all 12 have a viewer: every one is a command run by a
body; none is viewer-less; `grep -rn '\.FindByName(' --include=*.go . | grep -v _test` lists them):

| Caller | Today | After (for a viewer who is neither admin nor `see-hidden`) |
|---|---|---|
| `look.go:76` (look at a body) | finds a hidden body | a hidden body is not found: the look goes on to items, exits and corpses, as for any unmatched word |
| `consider.go:23` | likewise | "<name> can't be found." |
| `attack.go:164` | attacks a hidden body by name | "You attack the darkness!" |
| `shoot.go:72` (the adjacent room) | likewise | likewise, judged by the shooter's own perception |
| `give.go:66`, `show.go:44` | give or show to a hidden body | not found (their existing not-found replies) |
| `party.go:102` (`SameRoomOnly`) | invites a hidden player | not found ("<name> not found.") |
| `modules/follow/follow.go:419` | follows a hidden body | not found (follow's existing reply) |
| `admin.command.go:43`, `admin.grant.go:49`, `admin.mute.go:30`, `:61` | an admin reaches a hidden body | unchanged: these run only for admins (or mods granted them), and an admin perceives every hidden body; a mod without `see-hidden` does not |

An NPC running a command (its own idle commands or an admin's `command`) is the viewer too: without
`see-hidden` it cannot name a hidden player. Lookups that are not of a room's bodies are outside the
rule and unchanged: `whisper` (`whisper.go:34`), `party invite` when not `SameRoomOnly`
(`party.go:104`) and the admin `locate`, `teleport` and `visit` (`admin.locate.go:34`,
`admin.teleport.go:47`, `:77`, `admin.visit.go:60`) all find an online player by name anywhere
(`accounts.GetByCharacterName`). Target-less picks (`attack` with no name, `attack.go:48-150`) are
not name lookups, but they hand the attacker a body: they skip a body it does not perceive (the
table's attack row).

**Size:** `rooms.go`, `roomdetails.go`, the 12 call sites in 10 files, `rooms_test.go` (4 calls),
then for the two newer rulings `messaging/perceive.go`, `modules/gmcp/gmcp.Room.go`,
`hooks/NewRound_DoCombat.go`, and for round 3 `modules/gmcp/gmcp.Char.go`, `gmcp.Comm.go` and
`usercommands/go.go` (`attack.go` and `follow.go` are already in the set); tests: 5 new test files.
Round 3 adds 3 production files and 2 test files; round 4 adds `PlayerDespawn_HandleLeave.go`,
`PlayerDrop_HandlePlayerDrop.go`, `quit.go`, `admin.teleport.go` and `dikubasic/ruleset.go` (the
presence-only lines) and no test file.

**Files:** modify `internal/rooms/rooms.go`, `internal/rooms/roomdetails.go`,
`internal/rooms/rooms_test.go`, `internal/messaging/perceive.go`, `modules/gmcp/gmcp.Room.go`,
`modules/gmcp/gmcp.Char.go`, `modules/gmcp/gmcp.Comm.go`, `internal/hooks/NewRound_DoCombat.go`,
`internal/hooks/PlayerDespawn_HandleLeave.go`, `internal/hooks/PlayerDrop_HandlePlayerDrop.go`,
`internal/rules/dikubasic/ruleset.go`, `internal/rules/dikubasic/ruleset_test.go`,
`internal/usercommands/admin.command.go`, `admin.grant.go`, `admin.mute.go`, `admin.teleport.go`,
`attack.go`, `consider.go`, `give.go`, `go.go`, `look.go`, `party.go`, `quit.go`, `shoot.go`,
`show.go`, `modules/follow/follow.go`; create `internal/hooks/reads_corpse_test.go` (its name must sort after
`combat_golden_test.go`: see Task 5b), `internal/hooks/perceive_combat_test.go`,
`internal/messaging/reveal_test.go`, `internal/usercommands/perceive_test.go`,
`modules/gmcp/visibility_test.go` (the first test file in `modules/gmcp`: the ok-package count
becomes 41, Task 10).

- [ ] **Step 1: Write the failing tests.** Create `internal/hooks/reads_corpse_test.go`:

```go
package hooks

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
)

// readsSeeHiddenBuf is the test buff carrying see-hidden.
const readsSeeHiddenBuf = 93910

// A hidden body is not in the list a viewer's names are matched against
// (Calabe, 2026-11-11): no ordinal, no name, not by ID either. A viewer
// with see-hidden perceives it, as does an admin; a body always perceives
// itself.
func TestHiddenBodiesAreNotInTheViewersNames(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, cid, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	w.hide(&rat.Character)
	w.hide(bob.Character)

	playerId, mobId := room.FindByName(ann.Character, "rat")
	assert.Zero(t, playerId)
	assert.True(t, mobId.IsZero(), "the hidden rat cannot be named")
	playerId, _ = room.FindByName(ann.Character, "bob")
	assert.Zero(t, playerId, "nor the hidden Bob")
	_, mobId = room.FindByName(ann.Character, rat.Character.Id.String())
	assert.True(t, mobId.IsZero(), "not by its ID either")
	playerId, _ = room.FindByName(bob.Character, "bob")
	assert.Equal(t, bob.UserId, playerId, "Bob perceives himself")

	w.admin(cid) // an admin perceives the hidden (reads_golden_test.go:381)
	_, mobId = room.FindByName(cid.Character, "rat")
	assert.Equal(t, rat.Character.Id, mobId, "an admin perceives it")

	w.flag(ann.Character, readsSeeHiddenBuf, "see-hidden")
	_, mobId = room.FindByName(ann.Character, "rat")
	assert.Equal(t, rat.Character.Id, mobId, "see-hidden perceives it, as the room view does")
}
```

  Append to the same file (add `"strings"` to its imports): the follow lines and the room view.

```go
// Following does not cancel hiding (Calabe, 2026-11-11): a hidden follower
// the target does not perceive is not announced, not even as "Someone",
// starting or stopping; the follower still reads its own lines. An admin
// target perceives it and reads them as before (reads golden H's follow
// section shows the unhidden lines).
func TestHiddenFollowerIsNotAnnounced(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, _ := w.trio()
	w.hide(bob.Character)

	w.run("Bob", "follow Ann")
	w.run("Bob", "follow stop")
	out := w.out.String()
	assert.NotContains(t, out, `Ann: <ansi fg="username">Bob</ansi> is following you.`)
	assert.NotContains(t, out, `following you.`, "no line to Ann at all, not even as someone")
	assert.Contains(t, out, `Bob: You start following <ansi fg="username">Ann</ansi>.`)
	assert.Contains(t, out, `Bob: You are no longer following <ansi fg="username">Ann</ansi>.`)

	w.admin(ann)
	w.run("Bob", "follow Ann")
	w.run("Bob", "follow stop")
	out = w.out.String()
	assert.Contains(t, out, `Ann: <ansi fg="username">Bob</ansi> is following you.`, "an admin perceives Bob")
	assert.Contains(t, out, `Ann: <ansi fg="username">Bob</ansi> stopped following you.`)
}

// The room view lists a hidden player or NPC only to a viewer who perceives
// it (rooms.Perceives): see-hidden or an admin.
func TestRoomViewListsOnlyWhatTheViewerPerceives(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	w.hide(bob.Character)
	w.hide(&rat.Character)

	d := rooms.GetDetails(room, ann)
	assert.NotContains(t, strings.Join(d.VisiblePlayers, ","), "Bob", "Ann does not perceive hidden Bob")
	assert.NotContains(t, strings.Join(d.VisibleMobs, ","), "rat", "nor the hidden rat")

	w.admin(ann)
	d = rooms.GetDetails(room, ann)
	assert.Contains(t, strings.Join(d.VisiblePlayers, ","), "Bob", "an admin does")
	assert.Contains(t, strings.Join(d.VisibleMobs, ","), "rat")
}
```

  Append to the same file (add `github.com/GoMudEngine/GoMud/internal/events` and
  `github.com/GoMudEngine/GoMud/internal/ids` to its imports): `follow lose`, a follow running out
  (an NPC follower, an NPC leader), and the presence-only lines (round 4).

```go
// follow lose (also run by a leader's teleport) tells each follower its
// follow ended: a leader the follower does not perceive is "someone".
func TestFollowLoseNamesOnlyAPerceivedLeader(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, _, _, _ := w.trio()
	w.run("Bob", "follow Ann")
	w.hide(ann.Character)
	w.run("Ann", "follow lose")
	out := w.out.String()
	assert.Contains(t, out, `Bob: You are no longer following someone.`)
	assert.NotContains(t, out, `Bob: You are no longer following <ansi fg="username">Ann</ansi>.`)
}

// When a follow runs out: an NPC follower's "stopped following you" is only
// its presence, so a target who does not perceive it reads nothing; a player
// following an NPC it does not perceive reads "someone".
func TestFollowExpiryFollowsTheRule(t *testing.T) {
	w := newReadsWorld(t, 1)
	_, _, _, rat := w.trio()
	w.run("rat", "follow Ann")
	w.run("Bob", "follow rat")
	w.hide(&rat.Character)
	w.do("NewRound far in the future", func() { events.AddToQueue(events.NewRound{RoundNumber: 1 << 40}) })
	out := w.out.String()
	assert.NotContains(t, out, `Ann: <ansi fg="mobname">rat</ansi> stopped following you.`, "Ann does not perceive the hidden rat")
	assert.NotContains(t, out, `stopped following you`)
	assert.Contains(t, out, `Bob: You are no longer following someone.`)
}

// A line that is only a body's presence (starting to quit, leaving the
// game) reaches only viewers who perceive it (rooms.NotPerceiving).
func TestPresenceLinesReachOnlyThoseWhoPerceive(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, rat := w.trio()
	dan := w.addPlayer("Dan", readsRoomA)
	w.admin(bob)
	w.hide(ann.Character)

	assert.ElementsMatch(t, []ids.ID{dan.Character.Id, rat.Character.Id}, rooms.NotPerceiving(ann.Character),
		"Dan and the rat do not perceive Ann; Bob, an admin, does")

	w.run("Ann", "quit")
	out := w.out.String()
	assert.Contains(t, out, `Bob: <ansi fg="username">Ann</ansi> sits down and begins to meditate.`)
	assert.NotContains(t, out, `Dan: `, "nothing to Dan, not even someone")

	w.do("Ann leaves the game", func() { HandleLeave(events.PlayerDespawn{UserId: ann.UserId, RoomId: readsRoomA}) })
	out = w.out.String()
	assert.Contains(t, out, `Bob: Suddenly and without warning`)
	assert.NotContains(t, out, `Dan: `, "still nothing to Dan")
}
```

  Append to the same file (add `github.com/GoMudEngine/GoMud/internal/actors`,
  `github.com/GoMudEngine/GoMud/internal/characters`, `github.com/GoMudEngine/GoMud/internal/parties`
  and `github.com/stretchr/testify/require` to its imports): one case per remaining presence-line
  call site in this task's hooks and commands, so a site that forgets its `Exclude` fails its own
  case (plan review round 5). The recovery line has its own test (`dikubasic`, below) and the revive
  line its own in Task 6.

```go
// linesTo is what the record shows the named player reading.
func linesTo(out, who string) string {
	got := []string{}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, who+": ") {
			got = append(got, line)
		}
	}
	return strings.Join(got, "\n")
}

// Each presence-only line, one call site per case (round 5): a hidden body's
// line reaches an admin (Eve) and never a viewer who does not perceive it
// (Dan), not even as "someone".
func TestEachPresenceLineReachesOnlyThoseWhoPerceive(t *testing.T) {
	for _, tc := range []struct {
		name, line string
		run        func(w *readsWorld, ann, zed *characters.Character)
	}{
		{"teleport", "appears in a flash of light", func(w *readsWorld, ann, _ *characters.Character) {
			w.run("Ann", "teleport 93102") // Ann, an admin, into room B
		}},
		{"teleport party", "appears in a flash of light", func(w *readsWorld, _, _ *characters.Character) {
			w.run("Ann", "teleport 93102") // the leader moves Zed, the hidden member, too
		}},
		{"flee", "flees to the", func(w *readsWorld, ann, _ *characters.Character) {
			w.do("Ann flees", func() { flee(actors.Get(ann.Id), rooms.LoadRoom(readsRoomA)) })
		}},
		{"drop", `<ansi fg="red">drops to the ground!</ansi>`, func(w *readsWorld, ann, _ *characters.Character) {
			w.do("Ann drops", func() { HandlePlayerDrop(events.PlayerDrop{UserId: w.players["Ann"].UserId, RoomId: readsRoomA}) })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newReadsWorld(t, 1)
			ann, _, _, rat := w.trio()
			zed := w.addPlayer("Zed", readsRoomA)
			w.admin(ann)
			hidden := ann.Character
			watchRoom := readsRoomA
			switch tc.name {
			case "teleport", "teleport party":
				watchRoom = readsRoomB
			case "flee":
				ann.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
			}
			if tc.name == "teleport party" {
				p := parties.New(ann.UserId)
				require.NotNil(t, p)
				require.True(t, p.InvitePlayer(zed.UserId))
				require.True(t, p.AcceptInvite(zed.UserId))
				t.Cleanup(p.Disband)
				hidden = zed.Character
			}
			dan := w.addPlayer("Dan", watchRoom)
			eve := w.addPlayer("Eve", watchRoom)
			w.admin(eve)
			_, _ = dan, eve
			w.hide(hidden)

			tc.run(w, ann.Character, zed.Character)
			out := w.out.String()
			assert.Contains(t, linesTo(out, "Eve"), hidden.Name+`</ansi> `+tc.line, "an admin perceives the hidden body")
			for _, line := range strings.Split(linesTo(out, "Dan"), "\n") {
				if !strings.Contains(line, tc.line) {
					continue
				}
				assert.NotContains(t, line, hidden.Name, "Dan does not read the hidden body's line")
				assert.NotContains(t, line, "Someone", "not even as someone")
			}
		})
	}
}
```

  Append to `internal/rules/dikubasic/ruleset_test.go` (add `github.com/GoMudEngine/GoMud/internal/buffs`
  to its imports; `accounts`, `actors`, `ids` and `rooms` are there, and `testRoom` and
  `recoveryCopy` are Task 2's): the recovery room's line, through `Regen`, Task 5's code.

```go
// The recovery room's line is only the body's presence (round 5): a viewer
// there who does not perceive the hidden body is excluded from it; an admin
// is not, nor is anyone while the body is not hidden.
func TestRecoveryLineExcludesThoseWhoDoNotPerceive(t *testing.T) {
	const recovery = 94009
	testRoom(t, recovery, "regenzone")
	recoveryCopy(t, recovery, "regenzone")
	r := rooms.LoadRoom(recovery)
	a, dan, eve := player(t, 9651), player(t, 9652), player(t, 9653)
	for userId, p := range map[int]*actors.Actor{9651: a, 9652: dan, 9653: eve} {
		p.Character.RoomId = recovery
		r.AddPlayer(userId)
	}
	actors.AccountOf(eve.Id).Role = accounts.RoleAdmin
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	reseed(t)

	a.Character.Health, a.Character.Mana = 1, 0
	got := New().Regen(a, 1)
	require.Len(t, got.Messages, 2)
	assert.Empty(t, got.Messages[1].Exclude, "unhidden: everyone perceives it")

	const hiddenBuff = 96209
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: hiddenBuff, Name: "test hidden", TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{"hidden"},
	}))
	require.NoError(t, a.Character.AddBuff(hiddenBuff, true))
	require.True(t, a.Character.HasBuffFlag("hidden"))

	// AddBuff runs Validate, which recomputes the maximums: set them again.
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	a.Character.Health, a.Character.Mana = 1, 0
	got = New().Regen(a, 2)
	require.Len(t, got.Messages, 2)
	assert.Equal(t, []ids.ID{dan.Id}, got.Messages[1].Exclude, "Dan does not perceive the hidden body; Eve, an admin, does")
}
```

  Create `internal/usercommands/perceive_test.go` (untargeted `attack` and walking in):

```go
package usercommands

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hiddenTestBuffId = 92910

// hideBody gives a the hidden flag through a test buff.
func hideBody(t *testing.T, a *actors.Actor) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: hiddenTestBuffId, Name: "test hidden", TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{"hidden"},
	}))
	if a.Character.Buffs.List == nil { // a bare test NPC has no buff maps yet
		a.Character.Buffs = buffs.New()
	}
	require.NoError(t, a.Character.AddBuff(hiddenTestBuffId, true))
	require.True(t, a.Character.HasBuffFlag("hidden"))
}

// Walking in, a player is told only of the hostile NPCs it perceives: a
// hidden one is not named ("notices you"), unless the player is an admin.
func TestEnteringNamesOnlyTheHostilesTheMoverPerceives(t *testing.T) {
	_, player, here := goSetup(t)
	lurker := mobs.NewTestInstance(mobs.Mob{Hostile: true})
	lurker.Character.Name = "Lurker"
	lurker.Character.Health = 10
	id := lurker.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	rooms.LoadRoom(testNextRoomId).AddMob(id)
	hideBody(t, actors.Get(id))

	walk := func(admin bool) string {
		player.Controller().Account().Role = accounts.RoleUser
		if admin {
			player.Controller().Account().Role = accounts.RoleAdmin
		}
		told := playerTexts(t, player, func() {
			_, err := Go("north", player, here, 0)
			require.NoError(t, err)
		})
		_, err := Go("south", player, rooms.LoadRoom(testNextRoomId), 0)
		require.NoError(t, err)
		player.Character.ActionPoints = 100
		return strings.Join(told, "")
	}
	assert.NotContains(t, walk(false), "Lurker", "the hidden Lurker is not named")
	assert.Contains(t, walk(true), "Lurker", "an admin perceives it (the control: the line exists)")
}

// An untargeted attack never picks a body the attacker does not perceive
// (the one visibility rule): the hidden Rat fighting the player is not
// picked, nor picked at random; an admin picks it.
func TestUntargetedAttackPicksOnlyWhatTheAttackerPerceives(t *testing.T) {
	npc, player, room := goSetup(t)
	user := player.Controller().Account()
	accounts.SetTestUser(user) // the room's fight lists find the player's account by body
	t.Cleanup(func() { accounts.RemoveTestUser(user.UserId) })
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	hideBody(t, npc)

	for _, rest := range []string{``, `*`, `*mob`} {
		told := playerTexts(t, player, func() {
			_, err := Attack(rest, player, room, 0)
			require.NoError(t, err)
		})
		assert.Nil(t, player.Character.Aggro, "%q: no target picked", rest)
		if rest == `` {
			assert.Equal(t, []string{"You attack the darkness!\n"}, told)
		}
	}

	// The control: the same pick with the Rat perceived (an admin perceives
	// it) does attack it, so the refusals above are the rule's.
	user.Role = accounts.RoleAdmin
	_, err := Attack(``, player, room, 0)
	require.NoError(t, err)
	require.NotNil(t, player.Character.Aggro, "an admin perceives the Rat and attacks it")
	assert.Equal(t, npc.Id, player.Character.Aggro.Target)
}
```

  Create `modules/gmcp/visibility_test.go`, the package's first test file (its `TestMain` sets up
  the logger and the actor registry, as `usercommands_test.go:27-41` does):

```go
package gmcp

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init()
	os.Exit(m.Run())
}

const (
	visRoom      = 95001
	visHiddenBuf = 95901
	visSeeBuf    = 95902
)

// visWorld is one room for the GMCP visibility tests; events queued while
// setting it up are dropped.
func visWorld(t *testing.T) *rooms.Room {
	t.Helper()
	actors.Reset()
	accounts.ResetActiveUsers()
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone, r.Title = visRoom, "gmcpzone", "GMCP Room"
	rooms.AddTestRoom(r)
	t.Cleanup(func() {
		rooms.RemoveTestRoom(visRoom)
		accounts.ResetActiveUsers()
		actors.Reset()
	})
	return r
}

func visPlayer(t *testing.T, r *rooms.Room, userId int, name string) *accounts.Account {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name, c.RoomId, c.Zone = name, r.RoomId, r.Zone
	u := &accounts.Account{UserId: userId, Username: name, Character: c, Role: accounts.RoleUser}
	require.NotNil(t, actors.Spawn(c, actors.NewConnectionController(u)))
	accounts.SetTestUser(u)
	r.AddPlayer(userId)
	return u
}

func visMob(t *testing.T, r *rooms.Room, name string) *mobs.Mob {
	t.Helper()
	m := mobs.Mob{MobId: 95201}
	m.Character = *characters.New()
	m.Character.Name = name
	inst := mobs.NewTestInstance(m)
	id := inst.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	r.AddMob(id)
	return inst
}

// visFlag gives c a test buff carrying one flag.
func visFlag(t *testing.T, c *characters.Character, buffId int, flag string) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000, Flags: []string{flag},
	}))
	require.NoError(t, c.AddBuff(buffId, true))
	require.True(t, c.HasBuffFlag(flag))
}

// outs collects the GMCPOut and GMCPRoomUpdate events fn queues.
func outs(t *testing.T, fn func()) (sent []GMCPOut, updates []GMCPRoomUpdate) {
	t.Helper()
	events.ProcessEvents()
	a := events.RegisterListener(GMCPOut{}, func(e events.Event) events.ListenerReturn {
		sent = append(sent, e.(GMCPOut))
		return events.Cancel // not on to a connection
	}, events.First)
	b := events.RegisterListener(GMCPRoomUpdate{}, func(e events.Event) events.ListenerReturn {
		updates = append(updates, e.(GMCPRoomUpdate))
		return events.Cancel
	}, events.First)
	defer events.UnregisterListener(GMCPOut{}, a)
	defer events.UnregisterListener(GMCPRoomUpdate{}, b)
	fn()
	events.ProcessEvents()
	return sent, updates
}

func names(list []GMCPRoomModule_Payload_Contents_Character) []string {
	out := []string{}
	for _, c := range list {
		out = append(out, c.Name)
	}
	return out
}

// The room contents list a hidden body only to a viewer who perceives it
// (rooms.Perceives): see-hidden, an admin, never anyone else.
func TestRoomContentsListOnlyWhatTheViewerPerceives(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	bob := visPlayer(t, r, 9502, "Bob")
	ada := visPlayer(t, r, 9503, "Ada")
	ada.Role = accounts.RoleAdmin
	rat := visMob(t, r, "rat")
	visFlag(t, bob.Character, visHiddenBuf, "hidden")
	visFlag(t, &rat.Character, visHiddenBuf, "hidden")
	g := &GMCPRoomModule{}

	players, _ := g.GetRoomNode(ann, `Room.Info.Contents.Players`)
	npcs, _ := g.GetRoomNode(ann, `Room.Info.Contents.Npcs`)
	assert.Equal(t, []string{"Ada"}, names(players.([]GMCPRoomModule_Payload_Contents_Character)), "Ann does not perceive hidden Bob")
	assert.Empty(t, npcs.([]GMCPRoomModule_Payload_Contents_Character), "nor the hidden rat")

	players, _ = g.GetRoomNode(ada, `Room.Info.Contents.Players`)
	npcs, _ = g.GetRoomNode(ada, `Room.Info.Contents.Npcs`)
	assert.Equal(t, []string{"Ann", "Bob"}, names(players.([]GMCPRoomModule_Payload_Contents_Character)), "an admin perceives Bob")
	assert.Equal(t, []string{"rat"}, names(npcs.([]GMCPRoomModule_Payload_Contents_Character)), "and the rat")

	visFlag(t, ann.Character, visSeeBuf, "see-hidden")
	npcs, _ = g.GetRoomNode(ann, `Room.Info.Contents.Npcs`)
	assert.Equal(t, []string{"rat"}, names(npcs.([]GMCPRoomModule_Payload_Contents_Character)), "see-hidden perceives it")
}

// Char.Enemies lists a hidden enemy only to a viewer who perceives it.
func TestEnemiesListOnlyWhatTheViewerPerceives(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	rat := visMob(t, r, "rat")
	rat.Character.SetAggro(ann.Character.Id, characters.DefaultAttack)
	visFlag(t, &rat.Character, visHiddenBuf, "hidden")
	g := &GMCPCharModule{}

	enemies, _ := g.GetCharNode(ann, `Char.Enemies`)
	assert.Empty(t, enemies.([]GMCPCharModule_Enemy), "Ann does not perceive the hidden rat fighting her")

	ann.Role = accounts.RoleAdmin
	enemies, _ = g.GetCharNode(ann, `Char.Enemies`)
	require.Len(t, enemies.([]GMCPCharModule_Enemy), 1, "an admin does")
	assert.Equal(t, "rat", enemies.([]GMCPCharModule_Enemy)[0].Name)
}

// A say names its speaker on Comm.Channel only to a receiver who perceives
// it; to anyone else the sender reads "Someone", as the text does.
func TestSayNamesItsSpeakerOnlyToThoseWhoPerceiveIt(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	sneak := visPlayer(t, r, 9502, "Sneak")
	ada := visPlayer(t, r, 9503, "Ada")
	ada.Role = accounts.RoleAdmin
	visFlag(t, sneak.Character, visHiddenBuf, "hidden")

	sent, _ := outs(t, func() {
		(&GMCPCommModule{}).onComm(events.Communication{SourceActorId: sneak.Character.Id, CommType: `say`, Name: "Sneak", Message: "hi"})
	})
	senders := map[int]GMCPCommModule_Payload{}
	for _, o := range sent {
		senders[o.UserId] = o.Payload.(GMCPCommModule_Payload)
	}
	require.Len(t, senders, 3)
	assert.Equal(t, "Someone", senders[ann.UserId].Sender, "Ann does not perceive Sneak")
	assert.Empty(t, senders[ann.UserId].Source, "nor learns whether it is a player")
	assert.Equal(t, "Sneak", senders[ada.UserId].Sender, "an admin does")
	assert.Equal(t, "player", senders[ada.UserId].Source)
	assert.Equal(t, "Sneak", senders[sneak.UserId].Sender, "Sneak perceives itself")
}

// Leaving the game and moving name or announce a body only to viewers who
// perceive it.
func TestRoomEventsReachOnlyThoseWhoPerceiveTheBody(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	ada := visPlayer(t, r, 9502, "Ada")
	ada.Role = accounts.RoleAdmin
	sneak := visPlayer(t, r, 9503, "Sneak")
	visFlag(t, sneak.Character, visHiddenBuf, "hidden")
	g := &GMCPRoomModule{}

	sent, _ := outs(t, func() {
		g.despawnHandler(events.PlayerDespawn{UserId: sneak.UserId, RoomId: r.RoomId, CharacterName: "Sneak"})
	})
	to := []int{}
	for _, o := range sent {
		to = append(to, o.UserId)
	}
	assert.Equal(t, []int{ada.UserId}, to, "only the admin reads Room.RemovePlayer for the hidden Sneak")

	_, updates := outs(t, func() {
		g.roomChangeHandler(events.RoomChange{ActorId: sneak.Character.Id, ToRoomId: r.RoomId})
	})
	to = []int{}
	for _, u := range updates {
		if u.Identifier == `Room.Info.Contents.Players` {
			to = append(to, u.UserId)
		}
	}
	assert.Equal(t, []int{ada.UserId}, to, "only the admin's room list is refreshed for the hidden mover")
	assert.NotContains(t, to, ann.UserId)

	// The room left behind: the same rule (its own loop).
	_, updates = outs(t, func() {
		g.roomChangeHandler(events.RoomChange{ActorId: sneak.Character.Id, FromRoomId: r.RoomId})
	})
	to = []int{}
	for _, u := range updates {
		if u.Identifier == `Room.Info.Contents.Players` { // the mover gets its own Room.Info
			to = append(to, u.UserId)
		}
	}
	assert.Equal(t, []int{ada.UserId}, to, "only the admin's list is refreshed in the room left")

	// A fight change: the same rule.
	_, updates = outs(t, func() {
		g.aggroChangedHandler(events.AggroChanged{ActorId: sneak.Character.Id, RoomId: r.RoomId})
	})
	to = []int{}
	for _, u := range updates {
		to = append(to, u.UserId)
	}
	assert.ElementsMatch(t, []int{ada.UserId, sneak.UserId}, to, "the admin and Sneak itself, not Ann")
}
```

  Create `internal/hooks/perceive_combat_test.go` (a combat-world test; its name also sorts after
  `combat_golden_test.go`):

```go
package hooks

import (
	"fmt"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/stretchr/testify/require"
)

// A hidden defender cannot be fought by an attacker who does not perceive
// it; an admin perceives it (the one visibility rule, Calabe 2026-11-11).
func TestCombatAsksTheOneVisibilityRule(t *testing.T) {
	for _, admin := range []bool{false, true} {
		t.Run(fmt.Sprintf("admin=%v", admin), func(t *testing.T) {
			w := newCombatWorld(t, 111)
			alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
			rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "seen-rats")
			w.hide(&rat.Character)
			if admin {
				alice.Role = accounts.RoleAdmin
			}
			alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
			w.start()
			w.rounds(1)
			if admin {
				require.NotContains(t, w.reads[alice.UserId], "You can't seem to find your target.", "an admin finds it")
			} else {
				require.Contains(t, w.reads[alice.UserId], "You can't seem to find your target.")
			}
		})
	}
}
```

  Create `internal/messaging/reveal_test.go` (the package's own world: `hasFlag` fakes the flags,
  `messaging_test.go:37-56`):

```go
package messaging

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/stretchr/testify/assert"
)

// One visibility rule (Calabe, 2026-11-11): a hidden speaker reads by name
// to an admin and to a receiver with see-hidden, and as "Someone" to anyone
// else. Darkness still hides it from them all: the reveal is of hidden, not
// of light (canSee keeps its light and distance rules).
func TestHiddenIsRevealedToAdminsAndSeeHidden(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	sneak, _ := w.body(square, "Sneak", true)
	w.flags[sneak]["hidden"] = true
	_, bobRec := w.body(square, "Bob", true)
	_, adminRec := w.body(square, "Ada", true)
	adminRec.acc.Role = accounts.RoleAdmin
	seer, seerRec := w.body(square, "Seer", true)
	w.flags[seer]["see-hidden"] = true

	Send(say(sneak))
	assert.Equal(t, []string{`Someone says, "hi"`}, texts(bobRec))
	assert.Equal(t, []string{`<ansi fg="username">Sneak</ansi> says, "hi"`}, texts(adminRec), "an admin perceives the hidden")
	assert.Equal(t, []string{`<ansi fg="username">Sneak</ansi> says, "hi"`}, texts(seerRec), "see-hidden perceives the hidden")

	cave := w.room(2, "caves", nil)
	w.light[2] = 0
	dark, _ := w.body(cave, "Dark", true)
	w.flags[dark]["hidden"] = true
	_, caveAdminRec := w.body(cave, "Cad", true)
	caveAdminRec.acc.Role = accounts.RoleAdmin
	Send(say(dark))
	assert.Equal(t, []string{`Someone says, "hi"`}, texts(caveAdminRec), "darkness still hides it from an admin")
}
```

  and append to `internal/rooms/rooms_test.go` (the hidden cases need loaded buff specs, so they
  are in the hooks test above):

```go
// Perceives: a body perceives an unhidden body, and itself; so does a nil
// viewer. The two bodies have distinct IDs, so self and other differ.
func TestPerceivesAnUnhiddenBodyAndItself(t *testing.T) {
	viewer, other := characters.New(), characters.New()
	viewer.Id, other.Id = ids.New(ids.NoCell), ids.New(ids.NoCell)
	require.NotEqual(t, viewer.Id, other.Id)
	assert.True(t, Perceives(viewer, other))
	assert.True(t, Perceives(nil, other))
	assert.True(t, Perceives(viewer, viewer))
}
```

- [ ] **Step 2: Run; they fail.** The hooks and rooms tests fail to compile (`FindByName` takes no
  viewer; `undefined: Perceives`). `go test ./internal/messaging -run HiddenIsRevealed -count=1`
  fails on the admin and see-hidden asserts (today's `canSee` reveals the hidden to no one);
  `go test ./internal/hooks -run OneVisibility -count=1` (with the reads test commented out for a
  moment) fails on `admin=true`. `go test ./modules/gmcp ./internal/usercommands -run 'Perceive|Speaker|Untargeted|Entering' -count=1`
  fails in each test (today each site either names or picks the hidden body, or, for GMCP's room
  list, hides it from the admin too). `go test ./internal/rules/dikubasic -run RecoveryLine -count=1`
  compiles (the test names nothing new) and fails on the hidden assert: Dan is not excluded.

- [ ] **Step 3: `rooms.go`.** `FindByName` (`:1334-1341`, the whole function) becomes:

```go
// PerceivesWith is the one visibility rule (Calabe, 2026-11-11): whether
// viewer perceives subject, a body, with buff flags read through flag. A
// body always perceives itself; an unhidden body is perceived; a hidden one
// only by an admin (accounts.RoleAdmin, the role HasRolePermission lets
// through) or a viewer with see-hidden. A nil viewer perceives no hidden
// body. Light and distance are not part of it: messaging's canSee keeps
// them. Perceives is this with the bodies' own buff flags; messaging passes
// its test seam.
func PerceivesWith(flag func(c *characters.Character, flag string) bool, viewer, subject *characters.Character) bool {
	if viewer != nil && viewer.Id == subject.Id {
		return true
	}
	if !flag(subject, "hidden") {
		return true
	}
	if viewer == nil {
		return false
	}
	if acc := actors.AccountOf(viewer.Id); acc != nil && acc.Role == accounts.RoleAdmin {
		return true
	}
	return flag(viewer, "see-hidden")
}

// Perceives is PerceivesWith reading the bodies' own buff flags. Every
// place that decides whether a viewer perceives a body asks it: the room
// view, the name lookups, GMCP's room contents and combat; messaging asks
// PerceivesWith.
func Perceives(viewer, subject *characters.Character) bool {
	return PerceivesWith(func(c *characters.Character, flag string) bool { return c.HasBuffFlag(flag) }, viewer, subject)
}

// NotPerceiving lists the bodies in subject's room that do not perceive
// subject (Perceives). A line whose only content is subject's presence
// (arriving, leaving the game, starting to meditate, dropping, reviving)
// excludes them: no reply reveals a hidden body (Calabe, 2026-11-11).
func NotPerceiving(subject *characters.Character) []ids.ID {
	r := LoadRoom(subject.RoomId)
	if r == nil {
		return nil
	}
	out := []ids.ID{}
	for _, id := range r.GetBodies() {
		if id == subject.Id {
			continue
		}
		if a := actors.Get(id); a != nil && a.Character != nil && !Perceives(a.Character, subject) {
			out = append(out, id)
		}
	}
	return out
}

// FindByName finds the player and the NPC in the room that searchName
// names for viewer: a body viewer does not perceive (Perceives) is not in
// the lists, so it takes no ordinal and cannot be named, by name or ID.
// Players and NPCs are numbered apart; a caller takes the player when both
// come back.
func (r *Room) FindByName(viewer *characters.Character, searchName string, findTypes ...FindFlag) (playerId int, mobId ids.ID) {
	return r.findByName(func(c *characters.Character) bool { return Perceives(viewer, c) }, searchName, findTypes...)
}

// findByName is FindByName with the bodies it may match given by sees.
func (r *Room) findByName(sees func(*characters.Character) bool, searchName string, findTypes ...FindFlag) (playerId int, mobId ids.ID) {
	if len(findTypes) < 1 {
		findTypes = []FindFlag{FindAll}
	}
	mobId, _ = r.findMobByName(sees, searchName, findTypes...)
	playerId, _ = r.findPlayerByName(sees, searchName, findTypes...)
	return playerId, mobId
}
```

  In `findPlayerByName` (`:1343`) and `findMobByName` (`:1389`) add the first parameter
  `sees func(*characters.Character) bool`. In `findPlayerByName`: the ID-token branch's
  `if acc := actors.AccountOf(idMatch); acc != nil {` becomes
  `if acc := actors.AccountOf(idMatch); acc != nil && sees(acc.Character) {`, and in the name loop
  `if u == nil {` becomes `if u == nil || !sees(u.Character) {`. In `findMobByName`: the ID-token
  loop body `if mId == idMatch { return mId, nil }` becomes:

```go
			if mId != idMatch {
				continue
			}
			if m := mobs.GetInstance(mId); m == nil || sees(&m.Character) {
				return mId, nil
			}
```

  (a body with no instance is returned as today), and after the name loop's
  `continue // body with no instance (disconnect race)` block add:

```go
		if !sees(&m.Character) {
			continue // the viewer does not perceive it: no ordinal, no name
		}
```

- [ ] **Step 4: `roomdetails.go`.** The two inline checks become the shared one:
  `:252-256` (`if player.Character.HasBuffFlag("hidden") { ... see-hidden ... }`) becomes
  `if !Perceives(user.Character, player.Character) { // Don't show them if they are sneaking` /
  `continue` / `}`, and `:268-272` likewise with `&mob.Character`. Same rule plus the admin reveal.

- [ ] **Step 5: The `FindByName` callers.** In each of the 12 calls insert `actor.Character, ` as the
  first argument (`room.FindByName(rest)` becomes `room.FindByName(actor.Character, rest)`;
  ``adjacentRoom.FindByName(strings.Join(args, ` `))`` likewise). Every one runs inside a command
  whose body is `actor`. In `rooms_test.go` the four calls (`:81`, `:85`, `:90`, `:112`) take `nil, `.

- [ ] **Step 6: Messages, GMCP and combat ask the same rule.**
  - `messaging/perceive.go` `canSee` (`:40-55`): its comment's `a hidden body is unseen;` becomes
    `a body the receiver does not perceive (rooms.PerceivesWith, the one visibility rule: hidden
    unless the receiver is an admin or has see-hidden) is unseen;`, and
    `if far || hasFlag(subject.Character, "hidden") {` becomes
    `if far || !rooms.PerceivesWith(hasFlag, receiver.Character, subject.Character) {`. The self
    check before it, and the light check after it (`canSeeIn`), stay: a hidden speaker in the dark
    still reads "Someone" to an admin.
  - `modules/gmcp/gmcp.Room.go`: `:346` `if u.Character.HasBuffFlag("hidden") {` becomes
    `if !rooms.Perceives(user.Character, u.Character) {`, and `:374`
    `if mob.Character.HasBuffFlag("hidden") {` becomes `if !rooms.Perceives(user.Character, &mob.Character) {`
    (`user` is `GetRoomNode`'s viewer, `:267`).
  - `hooks/NewRound_DoCombat.go:195-196`: `// Can't see them, can't fight them.` /
    `if d.Character.HasBuffFlag("hidden") {` become
    `// Can't see them, can't fight them (the one visibility rule).` /
    `if !rooms.Perceives(a.Character, d.Character) {`.

  - **The round 3 and round 4 sites** (the table's **moved** rows: `Char.Enemies`, the say on
    `Comm.Channel`, the leave, move and fight refreshes, untargeted `attack`, the walk-in notice, the
    follow lines including `follow lose`, and the presence-only lines: leaving the game, starting to
    quit, appearing by teleport, fleeing, dropping). Each hunk below is the exact change, made with
    Edit (match the context lines, not the line numbers: Steps 3 to 5 have moved some). The two
    `GetRoomNode` hunks at the end of `gmcp.Room.go` are this step's second bullet above, shown again
    in place; the `NewRound_DoCombat.go` hunk is only the flee line (the "Can't see them" change is the
    third bullet above). Every file here already imports `rooms`; `follow.go` gains `characters`.
  - **The recovery room's line** (Task 5's `recoveryRegen`, `internal/rules/dikubasic/ruleset.go`):
    the room line `{actor} is recovering from a recent death.` is only the body's presence, so after
    its `ObserverText:` line add `Exclude: rooms.NotPerceiving(c),` (`c` is `a.Character` there;
    `rooms` is imported). The copy is private, so another viewer is rare, but the room itself and an
    admin's visit are not.

```diff
diff --git a/internal/hooks/NewRound_DoCombat.go b/internal/hooks/NewRound_DoCombat.go
index 4173b48f..bf94e0be 100644
--- a/internal/hooks/NewRound_DoCombat.go
+++ b/internal/hooks/NewRound_DoCombat.go
@@ -457,6 +457,7 @@ func flee(a *actors.Actor, aRoom *rooms.Room) {
 		Range:        perception.Room,
 		ActorText:    fmt.Sprintf(`You flee to the <ansi fg="exit">%s</ansi> exit!`, exitName),
 		ObserverText: fmt.Sprintf(`{actor} flees to the <ansi fg="exit">%s</ansi> exit!`, exitName),
+		Exclude:      rooms.NotPerceiving(a.Character), // a leaving: only its presence
 	})
 
 	endAggro(a)
diff --git a/internal/hooks/PlayerDespawn_HandleLeave.go b/internal/hooks/PlayerDespawn_HandleLeave.go
index 2c36a656..b191a467 100644
--- a/internal/hooks/PlayerDespawn_HandleLeave.go
+++ b/internal/hooks/PlayerDespawn_HandleLeave.go
@@ -49,7 +49,8 @@ func HandleLeave(e events.Event) events.ListenerReturn {
 
 	if _, ok := room.RemovePlayer(evt.UserId); ok {
 		// The leaving body is already out of the room's list; the template
-		// names it with {actor}, so an observer who cannot see reads "someone".
+		// names it with {actor}. The line is only its presence, so a viewer
+		// who does not perceive it reads nothing (rooms.NotPerceiving).
 		tplTxt, _ := templates.Process("player-despawn", nil)
 		messaging.Send(perception.Message{
 			Kind:         perception.WorldEvent,
@@ -57,6 +58,7 @@ func HandleLeave(e events.Event) events.ListenerReturn {
 			Senses:       perception.Sight | perception.Sound,
 			Range:        perception.Room,
 			ObserverText: tplTxt,
+			Exclude:      rooms.NotPerceiving(user.Character),
 		})
 	}
 
diff --git a/internal/hooks/PlayerDrop_HandlePlayerDrop.go b/internal/hooks/PlayerDrop_HandlePlayerDrop.go
index 52420e4f..0f02a0af 100644
--- a/internal/hooks/PlayerDrop_HandlePlayerDrop.go
+++ b/internal/hooks/PlayerDrop_HandlePlayerDrop.go
@@ -42,6 +42,7 @@ func HandlePlayerDrop(e events.Event) events.ListenerReturn {
 		Range:        perception.Room,
 		ActorText:    `<ansi fg="red">you drop to the ground!</ansi>`,
 		ObserverText: `{actor} <ansi fg="red">drops to the ground!</ansi>`,
+		Exclude:      rooms.NotPerceiving(user.Character), // only its presence
 	})
 
 	return events.Continue
diff --git a/internal/usercommands/admin.teleport.go b/internal/usercommands/admin.teleport.go
index bcc1c23b..34569935 100644
--- a/internal/usercommands/admin.teleport.go
+++ b/internal/usercommands/admin.teleport.go
@@ -112,6 +112,7 @@ func Teleport(rest string, actor *actors.Actor, room *rooms.Room, flags events.E
 				Senses:       perception.Sight | perception.Sound,
 				Range:        perception.Room,
 				ObserverText: `{actor} appears in a flash of light!`,
+				Exclude:      rooms.NotPerceiving(targetUser.Character), // only its presence
 			})
 
 			if party := parties.Get(targetUser.UserId); party != nil {
@@ -139,6 +140,7 @@ func Teleport(rest string, actor *actors.Actor, room *rooms.Room, flags events.E
 								Senses:       perception.Sight | perception.Sound,
 								Range:        perception.Room,
 								ObserverText: `{actor} appears in a flash of light!`,
+								Exclude:      rooms.NotPerceiving(partyUser.Character), // only its presence
 							})
 
 							if partyActor := actors.ByUserId(partyUser.UserId); partyActor != nil {
diff --git a/internal/usercommands/attack.go b/internal/usercommands/attack.go
index 3e9d2805..6e9d24fe 100644
--- a/internal/usercommands/attack.go
+++ b/internal/usercommands/attack.go
@@ -22,9 +22,30 @@ import (
 func othersOf(actor *actors.Actor, bodyIds []ids.ID) []ids.ID {
 	out := make([]ids.ID, 0, len(bodyIds))
 	for _, id := range bodyIds {
-		if id != actor.Id {
-			out = append(out, id)
+		if id == actor.Id {
+			continue
 		}
+		// Never a body the attacker does not perceive (rooms.Perceives).
+		if b := actors.Get(id); b != nil && b.Character != nil && !rooms.Perceives(actor.Character, b.Character) {
+			continue
+		}
+		out = append(out, id)
+	}
+	return out
+}
+
+// perceivedPlayers is the room's players less the attacker and less every
+// player it does not perceive (rooms.Perceives).
+func perceivedPlayers(actor *actors.Actor, room *rooms.Room, selfUserId int) []int {
+	out := []int{}
+	for _, userId := range room.GetPlayers() {
+		if userId == selfUserId {
+			continue
+		}
+		if u := accounts.GetByUserId(userId); u == nil || !rooms.Perceives(actor.Character, u.Character) {
+			continue
+		}
+		out = append(out, userId)
 	}
 	return out
 }
@@ -47,7 +68,7 @@ func Attack(rest string, actor *actors.Actor, room *rooms.Room, flags events.Eve
 		// If no argument supplied, attack whoever is attacking the player currently.
 		for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
 			m := mobs.GetInstance(mId)
-			if m.Character.Aggro == nil {
+			if m == nil || m.Character.Aggro == nil || !rooms.Perceives(actor.Character, &m.Character) {
 				continue
 			}
 
@@ -67,7 +88,7 @@ func Attack(rest string, actor *actors.Actor, room *rooms.Room, flags events.Eve
 		if attackMobId.IsZero() {
 			for _, uId := range room.GetPlayers(rooms.FindFightingPlayer) {
 				u := accounts.GetByUserId(uId)
-				if u == nil || u.Character.Aggro == nil {
+				if u == nil || u.Character.Aggro == nil || !rooms.Perceives(actor.Character, u.Character) {
 					continue
 				}
 
@@ -96,12 +117,15 @@ func Attack(rest string, actor *actors.Actor, room *rooms.Room, flags events.Eve
 
 						target := partyUser.Character.Aggro.Target
 
-						if mobs.GetInstance(target) != nil {
-							attackMobId = target
-							break
+						if tm := mobs.GetInstance(target); tm != nil {
+							if rooms.Perceives(actor.Character, &tm.Character) {
+								attackMobId = target
+								break
+							}
+							continue
 						}
 
-						if tu := accounts.GetByBodyId(target); tu != nil {
+						if tu := accounts.GetByBodyId(target); tu != nil && rooms.Perceives(actor.Character, tu.Character) {
 							attackPlayerId = tu.UserId
 							break
 						}
@@ -116,13 +140,7 @@ func Attack(rest string, actor *actors.Actor, room *rooms.Room, flags events.Eve
 		if rest == `*` { // * ANYONE
 
 			allMobs := othersOf(actor, room.GetMobs())
-			allPlayers := []int{}
-			for _, userId := range room.GetPlayers() {
-				if userId == selfUserId {
-					continue
-				}
-				allPlayers = append(allPlayers, userId)
-			}
+			allPlayers := perceivedPlayers(actor, room, selfUserId)
 
 			totalTargets := len(allMobs) + len(allPlayers)
 			if totalTargets == 0 {
@@ -146,13 +164,7 @@ func Attack(rest string, actor *actors.Actor, room *rooms.Room, flags events.Eve
 
 		} else { // *user etc. ANY PLAYER
 
-			allPlayers := []int{}
-			for _, userId := range room.GetPlayers() {
-				if userId == selfUserId {
-					continue
-				}
-				allPlayers = append(allPlayers, userId)
-			}
+			allPlayers := perceivedPlayers(actor, room, selfUserId)
 
 			if len(allPlayers) > 0 {
 				attackPlayerId = allPlayers[util.Rand(len(allPlayers))]
diff --git a/internal/usercommands/go.go b/internal/usercommands/go.go
index 34c90c09..4ba1c00c 100644
--- a/internal/usercommands/go.go
+++ b/internal/usercommands/go.go
@@ -286,6 +286,10 @@ func Go(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFl
 					if mob.Character.Aggro != nil {
 						continue
 					}
+					// The line names the NPC: only one the mover perceives.
+					if !rooms.Perceives(actor.Character, &mob.Character) {
+						continue
+					}
 					isHostile := mob.Hostile // Is it automatically hostile?
 					if !isHostile {
 						for _, groupName := range mob.Groups {
diff --git a/internal/usercommands/quit.go b/internal/usercommands/quit.go
index 097b5abb..9d07e608 100644
--- a/internal/usercommands/quit.go
+++ b/internal/usercommands/quit.go
@@ -27,6 +27,8 @@ func Quit(rest string, actor *actors.Actor, room *rooms.Room, flags events.Event
 		Range:        perception.Room,
 		ActorText:    "You sit down and begin your meditation.",
 		ObserverText: "{actor} sits down and begins to meditate.",
+		// Only the quitter's presence: none for a viewer who does not perceive it.
+		Exclude: rooms.NotPerceiving(actor.Character),
 	})
 	Tell(actor, "Your meditation must complete without interruption to quit gracefully.")
 	user.AddBuff(0, `quitting`)
diff --git a/modules/follow/follow.go b/modules/follow/follow.go
index e3ecdc5a..d50bb08e 100644
--- a/modules/follow/follow.go
+++ b/modules/follow/follow.go
@@ -8,6 +8,7 @@ import (
 	"github.com/GoMudEngine/GoMud/internal/accounts"
 	"github.com/GoMudEngine/GoMud/internal/actors"
 	"github.com/GoMudEngine/GoMud/internal/behaviour"
+	"github.com/GoMudEngine/GoMud/internal/characters"
 	"github.com/GoMudEngine/GoMud/internal/events"
 	"github.com/GoMudEngine/GoMud/internal/gametime"
 	"github.com/GoMudEngine/GoMud/internal/ids"
@@ -200,14 +201,20 @@ func (f *FollowModule) onNewRound(e events.Event) events.ListenerReturn {
 
 			// user doing the following? Tell both users
 			if followSourceUser != nil {
-				messaging.Tell(followTargetUser.Character.Id, fmt.Sprintf(`<ansi fg="username">%s</ansi> stopped following you.`, followSourceUser.Character.Name))
-				messaging.Tell(followSourceUser.Character.Id, fmt.Sprintf(`You are no longer following <ansi fg="username">%s</ansi>.`, followTargetUser.Character.Name))
+				// A line only about the follower's presence: none for a target
+				// who does not perceive it (Calabe, 2026-11-11).
+				if rooms.Perceives(followTargetUser.Character, followSourceUser.Character) {
+					messaging.Tell(followTargetUser.Character.Id, fmt.Sprintf(`<ansi fg="username">%s</ansi> stopped following you.`, followSourceUser.Character.Name))
+				}
+				messaging.Tell(followSourceUser.Character.Id, fmt.Sprintf(`You are no longer following %s.`, seenName(followSourceUser.Character, followTargetUser.Character, `username`)))
 				continue
 			}
 
 			// mob doing the following? tell the target user
 			if followSourceMob != nil {
-				messaging.Tell(followTargetUser.Character.Id, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> stopped following you.`, followSourceMob.Character.Name))
+				if rooms.Perceives(followTargetUser.Character, &followSourceMob.Character) {
+					messaging.Tell(followTargetUser.Character.Id, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> stopped following you.`, followSourceMob.Character.Name))
+				}
 				continue
 			}
 
@@ -219,7 +226,7 @@ func (f *FollowModule) onNewRound(e events.Event) events.ListenerReturn {
 
 			// user doing the following? Tell the following user
 			if followSourceUser != nil {
-				messaging.Tell(followSourceUser.Character.Id, fmt.Sprintf(`You are no longer following <ansi fg="mobname">%s</ansi>.`, followTargetMob.Character.Name))
+				messaging.Tell(followSourceUser.Character.Id, fmt.Sprintf(`You are no longer following %s.`, seenName(followSourceUser.Character, &followTargetMob.Character, `mobname`)))
 			}
 
 			continue
@@ -373,6 +380,18 @@ func (f *FollowModule) onPlayerDeath(e events.Event) events.ListenerReturn {
 // Commands
 //
 
+// seenName is subject's name, in colour, as viewer reads it in a follow
+// line that tells viewer about its own follow: the name when viewer
+// perceives subject (rooms.Perceives, the one visibility rule), else
+// "someone", as messaging shows an unseen body. The line carries news the
+// viewer gets anyway (its follow ended), so it is kept.
+func seenName(viewer, subject *characters.Character, colour string) string {
+	if rooms.Perceives(viewer, subject) {
+		return fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, colour, subject.Name)
+	}
+	return `someone`
+}
+
 // followCommand is the one follow command for players and NPCs. A follow key
 // is the account's UserId for a player and the body's ID for an NPC.
 func (f *FollowModule) followCommand(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
@@ -443,7 +462,10 @@ func (f *FollowModule) followCommand(rest string, actor *actors.Actor, room *roo
 				}
 
 				if followerUser := accounts.GetByUserId(fId.userId); followerUser != nil {
-					messaging.Tell(followerUser.Character.Id, fmt.Sprintf(`You are no longer following <ansi fg="%s">%s</ansi>.`, usercommands.NameColour(actor), actor.Character.Name))
+					// The follower's own line (its follow ended): "someone" for
+					// a leader it does not perceive. Also run by a leader's
+					// teleport (roomChangeHandler's "follow lose").
+					messaging.Tell(followerUser.Character.Id, fmt.Sprintf(`You are no longer following %s.`, seenName(followerUser.Character, actor.Character, usercommands.NameColour(actor))))
 				}
 			}
 
@@ -462,8 +484,10 @@ func (f *FollowModule) followCommand(rest string, actor *actors.Actor, room *roo
 		if wasFollowing.userId > 0 {
 
 			if followUser := accounts.GetByUserId(wasFollowing.userId); followUser != nil {
-				messaging.Tell(followUser.Character.Id, fmt.Sprintf(`<ansi fg="%s">%s</ansi> stopped following you.`, usercommands.NameColour(actor), actor.Character.Name))
-				usercommands.Tell(actor, fmt.Sprintf(`You are no longer following <ansi fg="username">%s</ansi>.`, followUser.Character.Name))
+				if rooms.Perceives(followUser.Character, actor.Character) {
+					messaging.Tell(followUser.Character.Id, fmt.Sprintf(`<ansi fg="%s">%s</ansi> stopped following you.`, usercommands.NameColour(actor), actor.Character.Name))
+				}
+				usercommands.Tell(actor, fmt.Sprintf(`You are no longer following %s.`, seenName(actor.Character, followUser.Character, `username`)))
 				return true, nil
 			}
 
@@ -472,7 +496,7 @@ func (f *FollowModule) followCommand(rest string, actor *actors.Actor, room *roo
 		if !wasFollowing.mobId.IsZero() {
 
 			if followMob := mobs.GetInstance(wasFollowing.mobId); followMob != nil {
-				usercommands.Tell(actor, fmt.Sprintf(`You are no longer following <ansi fg="mobname">%s</ansi>.`, followMob.Character.Name))
+				usercommands.Tell(actor, fmt.Sprintf(`You are no longer following %s.`, seenName(actor.Character, &followMob.Character, `mobname`)))
 				return true, nil
 			}
 
@@ -498,7 +522,11 @@ func (f *FollowModule) followCommand(rest string, actor *actors.Actor, room *roo
 
 		usercommands.Tell(actor, fmt.Sprintf(`You start following <ansi fg="username">%s</ansi>.`, targetUser.Character.Name))
 
-		messaging.Tell(targetUser.Character.Id, fmt.Sprintf(`<ansi fg="%s">%s</ansi> is following you.`, usercommands.NameColour(actor), actor.Character.Name))
+		// Shadowing: a follower the target does not perceive is not announced
+		// (following does not cancel hiding; Calabe, 2026-11-11).
+		if rooms.Perceives(targetUser.Character, actor.Character) {
+			messaging.Tell(targetUser.Character.Id, fmt.Sprintf(`<ansi fg="%s">%s</ansi> is following you.`, usercommands.NameColour(actor), actor.Character.Name))
+		}
 
 		return true, nil
 	}
diff --git a/modules/gmcp/gmcp.Char.go b/modules/gmcp/gmcp.Char.go
index d9025721..46d279ab 100644
--- a/modules/gmcp/gmcp.Char.go
+++ b/modules/gmcp/gmcp.Char.go
@@ -457,6 +457,10 @@ func (g *GMCPCharModule) GetCharNode(user *accounts.Account, gmcpModule string)
 				if mob == nil {
 					continue
 				}
+				// Only an enemy the viewer perceives (the one visibility rule).
+				if !rooms.Perceives(user.Character, &mob.Character) {
+					continue
+				}
 
 				e := GMCPCharModule_Enemy{
 					Id:      mob.Character.Id.String(),
diff --git a/modules/gmcp/gmcp.Comm.go b/modules/gmcp/gmcp.Comm.go
index f6b5f034..b18a126e 100644
--- a/modules/gmcp/gmcp.Comm.go
+++ b/modules/gmcp/gmcp.Comm.go
@@ -104,6 +104,14 @@ func (g *GMCPCommModule) onComm(e events.Event) events.ListenerReturn {
 
 	}
 
+	// A say names its speaker only to a receiver who perceives it (the one
+	// visibility rule); to anyone else it is "Someone", as the text reads
+	// (messaging's canSee), with no player or mob source.
+	var speaker *actors.Actor
+	if evt.CommType == `say` {
+		speaker = actors.Get(evt.SourceActorId)
+	}
+
 	for _, userId := range sendToUserIds {
 
 		// Exclude user from receiving their own messages?
@@ -111,10 +119,18 @@ func (g *GMCPCommModule) onComm(e events.Event) events.ListenerReturn {
 		//continue
 		//}
 
+		out := payload
+		if speaker != nil && speaker.Character != nil {
+			if receiver := accounts.GetByUserId(userId); receiver == nil || !rooms.Perceives(receiver.Character, speaker.Character) {
+				out.Sender = `Someone`
+				out.Source = ``
+			}
+		}
+
 		events.AddToQueue(GMCPOut{
 			UserId:  userId,
 			Module:  `Comm.Channel`,
-			Payload: payload,
+			Payload: out,
 		})
 
 	}
diff --git a/modules/gmcp/gmcp.Room.go b/modules/gmcp/gmcp.Room.go
index 32b475a0..e16971d5 100644
--- a/modules/gmcp/gmcp.Room.go
+++ b/modules/gmcp/gmcp.Room.go
@@ -82,6 +82,19 @@ func (g *GMCPRoomModule) itemOwnershipHandler(e events.Event) events.ListenerRet
 	return events.Continue
 }
 
+// perceivesBody reports whether the account's body perceives the body
+// bodyId (rooms.Perceives, the one visibility rule). A body no longer in the
+// registry has nothing to hide, so it is perceived.
+func perceivesBody(viewer *accounts.Account, bodyId actors.ActorId) bool {
+	if viewer == nil {
+		return false
+	}
+	if a := actors.Get(bodyId); a != nil && a.Character != nil {
+		return rooms.Perceives(viewer.Character, a.Character)
+	}
+	return true
+}
+
 func (g *GMCPRoomModule) aggroChangedHandler(e events.Event) events.ListenerReturn {
 
 	evt, typeOk := e.(events.AggroChanged)
@@ -100,6 +113,10 @@ func (g *GMCPRoomModule) aggroChangedHandler(e events.Event) events.ListenerRetu
 	}
 
 	for _, uId := range room.GetPlayers() {
+		// A viewer who does not perceive the body gets no update about it.
+		if !perceivesBody(accounts.GetByUserId(uId), evt.ActorId) {
+			continue
+		}
 		events.AddToQueue(GMCPRoomUpdate{
 			UserId:     uId,
 			Identifier: identifier,
@@ -127,6 +144,8 @@ func (g *GMCPRoomModule) despawnHandler(e events.Event) events.ListenerReturn {
 		return events.Continue
 	}
 
+	leaver := accounts.GetByUserId(evt.UserId)
+
 	//
 	// Send GMCP Updates for players leaving
 	//
@@ -141,6 +160,11 @@ func (g *GMCPRoomModule) despawnHandler(e events.Event) events.ListenerReturn {
 			continue
 		}
 
+		// Only a viewer who perceives the leaver reads its name.
+		if leaver != nil && !rooms.Perceives(u.Character, leaver.Character) {
+			continue
+		}
+
 		events.AddToQueue(GMCPOut{
 			UserId:  uid,
 			Payload: fmt.Sprintf(`Room.RemovePlayer "%s"`, evt.CharacterName),
@@ -175,7 +199,7 @@ func (g *GMCPRoomModule) roomChangeHandler(e events.Event) events.ListenerReturn
 	if evt.FromRoomId != 0 {
 		if oldRoom := rooms.LoadRoom(evt.FromRoomId); oldRoom != nil {
 			for _, uId := range oldRoom.GetPlayers() {
-				if uId == moverUserId {
+				if uId == moverUserId || !perceivesBody(accounts.GetByUserId(uId), evt.ActorId) {
 					continue
 				}
 				events.AddToQueue(GMCPRoomUpdate{
@@ -189,7 +213,7 @@ func (g *GMCPRoomModule) roomChangeHandler(e events.Event) events.ListenerReturn
 	if evt.ToRoomId != 0 {
 		if newRoom := rooms.LoadRoom(evt.ToRoomId); newRoom != nil {
 			for _, uId := range newRoom.GetPlayers() {
-				if uId == moverUserId {
+				if uId == moverUserId || !perceivesBody(accounts.GetByUserId(uId), evt.ActorId) {
 					continue
 				}
 				events.AddToQueue(GMCPRoomUpdate{
@@ -343,7 +367,7 @@ func (g *GMCPRoomModule) GetRoomNode(user *accounts.Account, gmcpModule string)
 				continue
 			}
 
-			if u.Character.HasBuffFlag("hidden") {
+			if !rooms.Perceives(user.Character, u.Character) {
 				continue
 			}
 
@@ -371,7 +395,7 @@ func (g *GMCPRoomModule) GetRoomNode(user *accounts.Account, gmcpModule string)
 				continue
 			}
 
-			if mob.Character.HasBuffFlag("hidden") {
+			if !rooms.Perceives(user.Character, &mob.Character) {
 				continue
 			}
 
```

  After this step `grep -rn 'HasBuffFlag("hidden")\|see-hidden' --include=*.go internal modules | grep -v _test`
  prints only the lines the table above leaves alone (the sneaking actors' own checks, `go.go:38`,
  `:330`, `look.go:37`, `read.go:32`, `cleanup.go:89`; the adjective, `character.go:435`; the prompt,
  `userrecord.prompt.go:244`; the two `Unseen` fields) and the comments and `"see-hidden"` of
  `PerceivesWith` and `canSee` (12 lines on the scratch worktree). Control: the same command
  before this task also prints `roomdetails.go:252`, `:268`, `gmcp.Room.go:346`, `:374` and
  `NewRound_DoCombat.go:196`.

- [ ] **Step 7: Run; they pass.** `go build ./...`; `go test ./internal/rooms ./internal/usercommands ./internal/messaging ./modules/gmcp ./internal/rules/dikubasic -count=1`
  and `go test ./internal/hooks -count=1` (twice) print `ok`. Probes (paste each failure line,
  restore):
  - room view (`roomdetails.go:268`): put back `mob.Character.HasBuffFlag("hidden") && !user.Character.HasBuffFlag("see-hidden")`
    as the skip condition: `TestRoomViewListsOnlyWhatTheViewerPerceives` fails (the admin does not
    see the rat); likewise `:252` for players (the admin does not see Bob);
  - GMCP room contents (`gmcp.Room.go`, the NPC list): put back `if mob.Character.HasBuffFlag("hidden") {`:
    `TestRoomContentsListOnlyWhatTheViewerPerceives` fails ("and the rat", "see-hidden perceives
    it"); likewise the players' list (`if u.Character.HasBuffFlag("hidden") {`: the admin does not
    see Bob);
  - `Char.Enemies`: make its new check `if false && ...`: `TestEnemiesListOnlyWhatTheViewerPerceives`
    fails;
  - `Comm.Channel`: make its receiver check `false && (...)`: `TestSayNamesItsSpeakerOnlyToThoseWhoPerceiveIt`
    fails ("Ann does not perceive Sneak");
  - `despawnHandler`: make its check `false && ...`: `TestRoomEventsReachOnlyThoseWhoPerceiveTheBody`
    fails on `Room.RemovePlayer`; drop `|| !perceivesBody(...)` from a `roomChangeHandler` loop: the
    same test fails on the refresh;
  - untargeted `attack`: drop `|| !rooms.Perceives(actor.Character, &m.Character)` from the first
    loop: `TestUntargetedAttackPicksOnlyWhatTheAttackerPerceives` fails (`"": no target picked`);
  - walk-in notice (`go.go`): make its check `if false && ...`: `TestEnteringNamesOnlyTheHostilesTheMoverPerceives`
    fails ("the hidden Lurker is not named");
  - follow: make the "is following you" check `if true || rooms.Perceives(...)`:
    `TestHiddenFollowerIsNotAnnounced` fails ("no line to Ann at all, not even as someone"); the same
    for the `follow stop` check (`if rooms.Perceives(followUser.Character, actor.Character)`);
  - make `PerceivesWith` return `true` after the self check: the hidden lookup test fails on its
    first four asserts, and so do two goldens, because messaging's `canSee` asks the same rule:
    `hidden_attacker.txt` line 11 (want "Someone watches you intently, waiting to strike.", got
    "Alice watches ...") and `F-comms-party.txt` line 22 (want "Bob: Someone waves", got "Bob: Ann
    waves"). Never `-update` on a probe (plan review round 4 corrected the earlier "no golden
    changes");
  - `follow lose` (`follow.go`, the `lostFollowers` loop): put back the real name,
    ``fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, usercommands.NameColour(actor), actor.Character.Name)`` in
    place of `seenName(...)`: `TestFollowLoseNamesOnlyAPerceivedLeader` fails;
  - a follow running out: make the NPC follower's check `if true || rooms.Perceives(...)`:
    `TestFollowExpiryFollowsTheRule` fails ("Ann does not perceive the hidden rat");
  - presence lines: set `quit.go`'s `Exclude:` to `nil`: `TestPresenceLinesReachOnlyThoseWhoPerceive`
    fails ("nothing to Dan"); restore, then set `HandleLeave`'s `Exclude:` to `nil`: it fails ("still
    nothing to Dan");
  - each other presence line, one at a time (round 5): delete the `Exclude:` line of
    `admin.teleport.go`'s first "appears in a flash of light" message: `TestEachPresenceLineReachesOnlyThoseWhoPerceive/teleport`
    fails, and only it; of its party loop's message: `/teleport_party` fails; of `flee`'s message in
    `NewRound_DoCombat.go`: `/flee` fails; of `HandlePlayerDrop`'s message: `/drop` fails (each
    "not even as someone"); of `recoveryRegen`'s room line: `TestRecoveryLineExcludesThoseWhoDoNotPerceive`
    fails (`go test ./internal/rules/dikubasic -run RecoveryLine -count=1`);  - GMCP refreshes: make `aggroChangedHandler`'s check `if false && ...`: `TestRoomEventsReachOnlyThoseWhoPerceiveTheBody`
    fails ("the admin and Sneak itself, not Ann"); drop `|| !perceivesBody(...)` from
    `roomChangeHandler`'s old-room loop: it fails ("only the admin's list is refreshed in the room
    left");
  - make the admin `if` false (`if acc := ...; false && ...`): the messaging admin assert, the combat
    `admin=true` subtest and the lookup test's admin assert fail;
  - in `canSee` put back `hasFlag(subject.Character, "hidden")`: the messaging admin and see-hidden
    asserts fail;
  - drop `|| !sees(u.Character)`: the "hidden Bob" assert fails; drop the ID-token `sees` check in
    `findMobByName`: the "by its ID" assert fails.

  (Run on a scratch worktree at `65bbc144`: the tests pass, and the admin, `canSee`, room view (NPC),
  GMCP NPC list, `Char.Enemies`, `Comm.Channel`, `despawnHandler`, untargeted `attack`, walk-in, both
  follow probes, and round 4's always-true, `follow lose`, follow expiry, `quit`, `HandleLeave`,
  `aggroChangedHandler` and old-room probes fail as stated; round 5's four per-site probes (both
  teleport lines, flee, drop) each fail their own subtest only. The recovery-line test needs Tasks 3
  and 5, so it ran against a stand-in carrying the same room line and `Exclude`: it passed, and
  failed with the `Exclude` deleted; its `Regen` form is unrun; `go test ./internal/hooks -count=1` prints `ok` twice; `go test ./... -count=1`
  shows only the 3 known Windows failures, 41 ok, `modules/gmcp` now among them.)

- [ ] **Step 8:** Per-task check. **Goldens: none change.** Reads goldens A, B and C make Ann an
  admin but hide no one; D, E, F and H hide Ann, and no one there is an admin or has `see-hidden`,
  so every observer's line still reads "Someone" where it did; a body always perceives itself, and
  Ann looks at, shows to and emotes with visible bodies only. The combat golden `hidden_attacker`
  hides the attacker, not the defender, and no observer is an admin. GMCP output is in no golden.
  Reads golden H's follow section hides no one: Ann perceives Bob, so her lines are sent as before,
  and `seenName` writes a perceived name exactly as the old lines did (`<ansi fg="username">Ann</ansi>`). No golden runs an untargeted `attack` on a hidden
  body or walks into a room holding a hidden hostile NPC. (Run on the scratch worktree: every golden
  byte-identical, with the filter on and switched off.) Commit the files in **Files**:
  `Phase 5.3: one visibility rule (rooms.Perceives): hidden, see-hidden and admins, for the room view, name lookups, messages, GMCP, follow, untargeted attack and combat`.

### Task 5b: Both ordinal forms, and one numbering per name: the living, then corpses newest first (Opus)

Calabe's rulings (spec, plan review round 2, 2026-11-11):

- **Both ordinal forms, everywhere, one parser.** `2.rat` and `rat#2` are the second rat for every
  command that takes a name. `util.GetMatchNumber` (`util.go:298-314`) learns the dot form; no
  second parser. Its two callers are the two name matchers, so every name lookup gets it:
  `util.FindMatchIn` (`util.go:322`; rooms' players, NPCs, containers, exits, corpses, party members,
  zone names, NPC templates) and `items.FindMatchIn` (`items.go:604`; every item lookup).
  `grep -rn 'GetMatchNumber' --include=*.go .` lists exactly those two, its definition (`util.go:298`),
  `util_test.go` and a comment in `findMobByName` (`rooms.go:1393`).
- **One numbering per name: the living first, then corpses newest first.** With a live rat and two
  dead rats, 1 is the live rat, 2 the rat killed most recently, 3 the rat that died first. Recency
  is `Corpse.RoundCreated`, newest first (of two from one round, the one added later). The dead
  character's name reaches its corpse (`rat` reaches `rat corpse`). Players and NPCs follow the same
  rule (no player-first ranking of corpses). Living lookups number exactly as today; only corpse
  lookups change, counting past the living. A corpse command whose ordinal lands on a living
  character refuses and does not skip to the next corpse: "The rat is still alive." for an NPC,
  "Bob is still alive." for a player (a player's name never takes "The").
- **The person typing is not in their own name's living count** when they loot or look at a corpse:
  Bob's `get all bob` or `look bob` reaches Bob's newest corpse; other players' numbering is
  unchanged (to Ann, Bob is still 1).
- **The living counted are those the viewer perceives** (Task 5a): a hidden body takes no ordinal,
  so a corpse lookup counts past it unseen, and no refusal ever names it.

**How the living are numbered today (read at `65bbc144`), which the corpse numbering counts past.**
`FindByName` (`rooms.go:1334-1341`) asks two lookups apart and returns both results:
`findPlayerByName` (`:1343-1387`) matches the names of the room's players in the order they came
into the room (`GetPlayers`, `:1158`, `r.playerIds()`); `findMobByName` (`:1389-1435`) matches the
room's NPCs in the order they came in (`GetMobs`, `:1065`, `r.mobIds()`), listing each as `name#k`
with k its position among all the room's NPCs (`:1417`), which only makes the map keys unique:
`GetMatchNumber` strips the ordinal from the search, never from the list, so the search's ordinal
counts matching NPCs, and an NPC is never an exact match (its listed name holds `#k`). Both refuse a
search starting with `#` or `@` (`:1345-1351`, `:1391-1397`: a bare `#2` would match every name by
position); the dot form needs no such guard, because the parser leaves a nameless `2.` as typed.
Both take an ID token first (`ids.Parse`, `0x<16 hex>:0x<16 hex>`, `ids.go:54-64`; it has no dot,
and `0x` is not digits-then-dot, so the dot form never reads one). Every caller of
`FindByName` (12, `grep -rn '\.FindByName(' --include=*.go . | grep -v _test`) takes the player when
both come back (for example `look.go:76-84`, `consider.go:23-40`). So **ordinal n reaches someone
living exactly when n is at most the number of matching players or at most the number of matching
NPCs**: the living count is the larger of the two, players and NPCs each numbered from 1, and for an
ordinal both reach, the player is the one it names. The corpse lookup asks the same lookup
(`findByName`, Task 5a) for this count (`livingReach`), so it cannot drift from it, with the bodies
it may match narrowed to those the viewer perceives, less the viewer itself.

**`look bob` typed by Bob, today:** `findPlayerByName` lists every player in the room, Bob
included, and `look` has no self check (`look.go:76-84`), so Bob looks at himself. After this task
Bob looks at himself only when no corpse answers to his name; with one, `look bob` shows his newest
corpse (Step 7). `consider` already drops a self match (`consider.go:24-26`) and `attack` refuses
one (`attack.go:167-169`); `give` and `show` have no self check today and keep none. Living lookups
for other commands are unchanged: the viewer stays in their own list there, so no living number
moves.

**What else Task 5b does (from the earlier round 2 revision, kept):** `util.FindMatchIn` returns the
matched text, which cannot tell two corpses of one name apart, so its loop moves unchanged into
`util.FindMatchIndexIn` (positions) and `FindMatchIn` wraps it: one matcher still. `FindCorpse` (a
copy) delegates to `FindCorpseByRef`; `look.go`'s unused third copy of the name lists goes; `bury`
removes the corpse it found (`RemoveCorpseRef`; `RemoveCorpse` matches owner, name and round, so of
two rats killed in one round it would remove the first). Comes before Task 6 so the death commit
lands on the new lookup.

**World data holds no name the dot form would misread** (2026-11-11, `65bbc144`):
`grep -rniE "^\s*-?\s*(name|namesimple|title|zonename|zone|displayname)\s*:\s*[\"']?[0-9]+\." _datafiles modules --include=*.yaml`
(names, titles, zones) and `grep -rnE "^\s*-?\s*[\"']?[0-9]+\.[^:]*[\"']?\s*:" _datafiles modules --include=*.yaml`
(map keys: nouns, exits, containers) each print 0 lines over the 100 YAML files. Controls: both
patterns print the planted line from a scratch file holding `  2.rat: a rat` and `name: 3.sword`;
the first with `\w+` in place of the key list prints 19 lines (`config.yaml:199`
`EquipmentDropChance: 0.25` among them). Character names cannot hold a dot: the shipped
`NameRejectRegex` is `^[a-zA-Z0-9_]+$` (`config.yaml:578`, read by `accounts.ValidateName`,
`users.go:517`).

**Files:** modify `internal/util/util.go`, `internal/rooms/rooms.go`, `internal/rooms/rooms_test.go`
(`TestRoom_FindCorpse`'s four calls take a nil viewer), `internal/usercommands/get.go`,
`internal/usercommands/look.go`, `modules/cleanup/cleanup.go`, `internal/hooks/reads_corpse_test.go`
(Task 5a's); create `internal/util/matchnumber_test.go`, `internal/util/findmatchindex_test.go`,
`internal/rooms/corpselookup_test.go`.

**The hooks test file's name must sort after `combat_golden_test.go`.** Go runs a package's tests in
file order, and any reads-world test that runs before `TestCombatGoldenFleeSuccess` sets off its known
order flake (`status.md` Watch): on the scratch worktree a file named `bury_test.go` made
`go test ./internal/hooks -count=1` fail on that test, and the same tests in `reads_corpse_test.go`
pass twice in a row. Do not rename it to sort earlier.

- [ ] **Step 1: Write the failing tests.** Create `internal/util/matchnumber_test.go`:

```go
package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Both ordinal forms, one parser (Calabe, 2026-11-11): "2.rat" and "rat#2"
// are the second rat. A dot counts only between digits and a name that does
// not start with a digit; anything else stays as typed.
func TestGetMatchNumberReadsBothOrdinalForms(t *testing.T) {
	for _, tt := range []struct {
		input string
		name  string
		num   int
	}{
		{"2.rat", "rat", 2},
		{"rat#2", "rat", 2},
		{"0.rat", "rat", 1}, // as rat#0: below 1 is 1
		{"rat", "rat", 1},
		{"1.5", "1.5", 1},   // a number, not the fifth "5"
		{".rat", ".rat", 1}, // no digits before the dot
		{"2.", "2.", 1},     // no name after the dot
		{"12.Rat Corpse", "rat corpse", 12},
		{"2.rat#3", "2.rat", 3}, // "#N" is read first
		{"a2.rat", "a2.rat", 1}, // not all digits before the dot
	} {
		name, num := GetMatchNumber(tt.input)
		assert.Equal(t, tt.name, name, "%q: name", tt.input)
		assert.Equal(t, tt.num, num, "%q: number", tt.input)
	}
}
```

  Create `internal/util/findmatchindex_test.go`:

```go
package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FindMatchIndexIn tells repeated items apart by position (two corpses of
// one name): "#N" picks the Nth, and an exact match wins over a prefix that
// comes before it.
func TestFindMatchIndexInTellsRepeatsApart(t *testing.T) {
	list := []string{"rat corpsewing corpse", "rat corpse", "rat corpse"}
	for _, tt := range []struct {
		search       string
		match, close int
	}{
		{"rat", -1, 0},
		{"rat#1", -1, 0},
		{"rat#2", -1, 1},
		{"rat#3", -1, 2},
		{"rat#4", -1, -1},
		{"rat corpse", 1, 0},
		{"rat corpse#2", 2, 1},
		{"", -1, -1},
	} {
		match, close := FindMatchIndexIn(tt.search, list...)
		assert.Equal(t, tt.match, match, "%q: match", tt.search)
		assert.Equal(t, tt.close, close, "%q: close match", tt.search)
	}
}
```

  Create `internal/rooms/corpselookup_test.go` (no one living: a bare `Room` has no bodies; the
  viewer is nil):

```go
package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// corpseOf is a corpse for the lookup tests; gold tells them apart.
func corpseOf(name string, userId, mobId int, round uint64, gold int) Corpse {
	return Corpse{UserId: userId, MobId: mobId, Character: characters.Character{Name: name}, RoundCreated: round, Gold: gold}
}

// With nobody living in the room, a name numbers its corpses newest first
// (RoundCreated; of two from one round, the later added), crumbled ones
// left out, and both ordinal forms work (Calabe, 2026-11-11). FindCorpse
// finds the same corpse.
func TestFindCorpseByRefNumbersCorpsesNewestFirst(t *testing.T) {
	r := &Room{}
	crumbled := corpseOf("rat", 0, 7, 9, 9)
	crumbled.Prunable = true
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 1))
	r.AddCorpse(corpseOf("rat", 0, 7, 8, 3)) // the newest
	r.AddCorpse(crumbled)
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 2)) // same round as gold 1, added later

	for search, gold := range map[string]int{
		"rat": 3, "rat#1": 3, "1.rat": 3,
		"rat#2": 2, "2.rat": 2,
		"rat#3": 1, "3.rat": 1,
		"rat corpse": 3, "rat corpse#2": 2, "2.rat corpse": 2,
	} {
		c, refusal := r.FindCorpseByRef(nil, search)
		require.NotNil(t, c, search)
		assert.Empty(t, refusal, search)
		assert.Equal(t, gold, c.Gold, search)
		v, ok := r.FindCorpse(nil, search)
		require.True(t, ok, search)
		assert.Equal(t, gold, v.Gold, "FindCorpse %s", search)
	}
	c, refusal := r.FindCorpseByRef(nil, "4.rat")
	assert.Nil(t, c, "three live rat corpses, not four (the crumbled one is not a candidate)")
	assert.Empty(t, refusal)
}

// Players' and NPCs' corpses share one numbering (no player-first ranking),
// and a corpse named exactly wins over one the name only begins.
func TestFindCorpseByRefOneListExactFirst(t *testing.T) {
	r := &Room{}
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 1))
	r.AddCorpse(corpseOf("Ratty", 9, 0, 6, 2))          // a player's, newer
	r.AddCorpse(corpseOf("rat corpsewing", 0, 8, 7, 3)) // newest; "rat corpsewing corpse" begins with "rat corpse"

	c, _ := r.FindCorpseByRef(nil, "rat")
	require.NotNil(t, c)
	assert.Equal(t, 3, c.Gold, "the newest corpse the name begins")
	c, _ = r.FindCorpseByRef(nil, "2.rat")
	require.NotNil(t, c)
	assert.Equal(t, 2, c.Gold, "the player's corpse is the second: one list")
	c, _ = r.FindCorpseByRef(nil, "rat corpse")
	require.NotNil(t, c)
	assert.Equal(t, 1, c.Gold, "exact over a prefix that comes first")
}

// RemoveCorpseRef removes the very corpse it is given, though another has
// the same owner, name and round (RemoveCorpse takes the first of those).
func TestRemoveCorpseRefRemovesTheCorpseFound(t *testing.T) {
	r := &Room{}
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 1))
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 2))

	// "rat" is the newest: of two from one round, the one added later.
	// RemoveCorpse, matching owner, name and round, would take the first.
	c, _ := r.FindCorpseByRef(nil, "rat")
	require.NotNil(t, c)
	require.Equal(t, 2, c.Gold)
	require.True(t, r.RemoveCorpseRef(c))
	require.Len(t, r.Corpses, 1)
	assert.Equal(t, 1, r.Corpses[0].Gold, "the one named went, the other stayed")
	assert.False(t, r.RemoveCorpseRef(&Corpse{}), "a corpse not in the room")
}
```

  Append to `internal/hooks/reads_corpse_test.go` (Task 5a's file; add
  `github.com/GoMudEngine/GoMud/internal/characters` and `github.com/stretchr/testify/require` to its
  imports). The reads harness runs module commands (it imports `modules/cleanup`,
  `reads_golden_test.go:72`); `trio` puts Ann, Bob and the live rat in room A and Cid in C
  (`:403-412`):

```go
// ratCorpses puts two rat corpses in room A: the older (round 5) holds 1
// gold, the newer (round 7) 2.
func ratCorpses(room *rooms.Room) {
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 1})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 7, Gold: 2})
}

// One numbering per name, the living first (Calabe, 2026-11-11). Room A
// holds the live rat, a live player Ratty and two rat corpses. Players and
// NPCs are numbered apart (FindByName) and a command takes the player when
// both match, so 1 reaches the living (Ratty, then the rat) and the corpses
// start at 2: 2 is the newest, 3 the older. A player's name takes no "The".
// The person typing is not in their own name's count: Ann's "ann" is her
// newest corpse; Bob's "ann" is Ann.
func TestCorpseNumbersCountPastTheLiving(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, _ := w.trio()
	w.addPlayer("Ratty", readsRoomA)
	room := rooms.LoadRoom(readsRoomA)
	ratCorpses(room)
	room.AddCorpse(rooms.Corpse{UserId: ann.UserId, Character: characters.Character{Name: "Ann"}, RoundCreated: 6, Gold: 3})

	c, refusal := room.FindCorpseByRef(ann.Character, "rat")
	assert.Nil(t, c)
	assert.Equal(t, "Ratty is still alive.", refusal, "1 is the living; the player when a player and an NPC both match")
	for search, gold := range map[string]int{"2.rat": 2, "rat#2": 2, "3.rat": 1, "rat corpse": 2, "ann": 3} {
		c, refusal := room.FindCorpseByRef(ann.Character, search)
		require.NotNil(t, c, search)
		assert.Empty(t, refusal, search)
		assert.Equal(t, gold, c.Gold, search)
	}
	c, refusal = room.FindCorpseByRef(ann.Character, "4.rat")
	assert.Nil(t, c, "two rat corpses only")
	assert.Empty(t, refusal)

	c, refusal = room.FindCorpseByRef(bob.Character, "ann")
	assert.Nil(t, c)
	assert.Equal(t, "Ann is still alive.", refusal, "to Bob, Ann is alive and first")
	c, _ = room.FindCorpseByRef(bob.Character, "2.ann")
	require.NotNil(t, c)
	assert.Equal(t, 3, c.Gold)

	c, refusal = room.FindCorpseByRef(ann.Character, "bob")
	assert.Nil(t, c, "no corpse answers to bob: nothing, not a refusal")
	assert.Empty(t, refusal)
}

// A hidden body takes no ordinal in a corpse lookup either, and nothing
// says it is there; a viewer with see-hidden counts it.
func TestHiddenBodiesTakeNoCorpseOrdinal(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, _, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	ratCorpses(room)
	w.hide(&rat.Character)

	c, refusal := room.FindCorpseByRef(ann.Character, "rat")
	require.NotNil(t, c, "the hidden rat takes no ordinal: 1 is the newest corpse")
	assert.Empty(t, refusal, "and nothing says it is there")
	assert.Equal(t, 2, c.Gold)

	w.flag(ann.Character, readsSeeHiddenBuf, "see-hidden")
	c, refusal = room.FindCorpseByRef(ann.Character, "rat")
	assert.Nil(t, c)
	assert.Equal(t, "The rat is still alive.", refusal)
}

// bury numbers as the lookup does: an ordinal on the living is refused, not
// skipped; "2.rat" buries the newest rat corpse, here the later of two from
// the same round (which RemoveCorpse, matching owner, name and round, would
// mistake for the first).
func TestBuryBuriesTheCorpseItNames(t *testing.T) {
	w := newReadsWorld(t, 1)
	w.trio() // the rat is alive in room A
	room := rooms.LoadRoom(readsRoomA)
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 1})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 2})

	w.run("Ann", "bury rat")
	require.Len(t, room.Corpses, 2, "the live rat is not a corpse")
	w.run("Ann", "bury 2.rat")

	require.Len(t, room.Corpses, 1)
	require.Equal(t, 1, room.Corpses[0].Gold, "the newest rat corpse was buried, the older stays")
	out := w.out.String()
	require.Contains(t, out, `Ann: The rat is still alive.`)
	require.Contains(t, out, `Ann: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`)
}

// Bob looking at "bob" sees his newest corpse when one is here (the person
// typing is not in their own name's count for a corpse); with none, he looks
// at himself, as before.
func TestLookAtYourOwnNameFindsYourCorpse(t *testing.T) {
	w := newReadsWorld(t, 1)
	_, bob, _, _ := w.trio()
	room := rooms.LoadRoom(readsRoomA)

	w.run("Bob", "look bob")
	before := w.out.String()
	require.NotContains(t, before, "Bob: You look at the", "no corpse yet: Bob looks at himself")

	room.AddCorpse(rooms.Corpse{UserId: bob.UserId, Character: characters.Character{Name: "Bob"}, RoundCreated: 5})
	w.run("Bob", "look bob")
	require.Contains(t, w.out.String()[len(before):], `Bob: You look at the <ansi fg="user-corpse">Bob corpse</ansi>.`)
}
```

- [ ] **Step 2: Run; they fail.** `go test ./internal/util ./internal/rooms -count=1` and
  `go test ./internal/hooks -run 'CorpseNumbers|CorpseOrdinal|Bury|LookAtYour' -count=1` fail to
  compile (`undefined: FindMatchIndexIn`, `undefined: RemoveCorpseRef`, and `FindCorpseByRef` takes
  no viewer and returns `bool`). Comment out the new hooks tests' bodies and the util index test for
  a moment:
  `go test ./internal/util -run GetMatchNumberReadsBoth -count=1` fails at `2.rat` (no dot form
  today). Restore.

- [ ] **Step 3: `util.go`, the parser.** `GetMatchNumber` (`:296-314`, from its comment
  `// accepts an input and splits it along a # if any.` through its closing brace) becomes:

```go
// GetMatchNumber splits an ordinal off a name: "rat#2" and "2.rat" are both
// the second rat (Calabe, 2026-11-11: both forms, for every name, in this one
// parser). A "#N" suffix is read first. Otherwise a prefix of digits and a dot
// counts when a name follows that does not start with a digit, so "1.5" stays
// a number and ".rat" and "2." stay as typed. An ordinal below 1 is 1; with
// none, the number is 1. The name comes back trimmed and lower case.
func GetMatchNumber(input string) (string, int) {
	// Clean up the input
	input = strings.TrimSpace(strings.ToLower(input))

	// "rat#2": the left is the name, the right the number
	if strings.Contains(input, "#") {
		parts := strings.Split(input, "#")
		input = parts[0]
		inputNumber, _ := strconv.Atoi(strings.Join(parts[1:], "#"))
		if inputNumber < 1 {
			inputNumber = 1
		}
		return input, inputNumber
	}

	// "2.rat": digits, a dot, then a name that does not start with a digit
	if dot := strings.IndexByte(input, '.'); dot > 0 {
		digits, name := input[:dot], strings.TrimSpace(input[dot+1:])
		if strings.Trim(digits, "0123456789") == `` && name != `` && (name[0] < '0' || name[0] > '9') {
			inputNumber, _ := strconv.Atoi(digits)
			if inputNumber < 1 {
				inputNumber = 1
			}
			return name, inputNumber
		}
	}

	return input, 1
}
```

  The existing `TestGetMatchNumber` (`util_test.go:509-531`) still passes unchanged.

- [ ] **Step 4: `util.go`, the matcher by position.** `FindMatchIn` (`:316-368` before Step 3; find
  it by text) becomes the two functions below. The loop is today's, moved: positions instead of
  text, and the contains-fallback condition `len(closeMatch) == 0` becomes `closeMatch < 0` (the same
  for every non-empty item).

```go
func FindMatchIn(searchName string, items ...string) (match string, closeMatch string) {

	matchIdx, closeMatchIdx := FindMatchIndexIn(searchName, items...)
	if matchIdx >= 0 {
		match = items[matchIdx]
	}
	if closeMatchIdx >= 0 {
		closeMatch = items[closeMatchIdx]
	}
	return match, closeMatch
}

// FindMatchIndexIn is FindMatchIn by position: it returns the index in items
// of the match and of the close match, -1 for none. Items may repeat (two
// corpses of one name, rooms.FindCorpseByRef); the index tells them apart
// where the text cannot. FindMatchIn is this, returning the items.
func FindMatchIndexIn(searchName string, items ...string) (match int, closeMatch int) {

	match, closeMatch = -1, -1

	if searchName == `` {
		return match, closeMatch // No match
	}

	searchName, searchNumber := GetMatchNumber(searchName)

	var matchCt int = 0
	var closeMatchCt int = 0

	for idx, i := range items {

		part, full := stringMatch(searchName, i, false)

		if part {
			closeMatchCt++
			if closeMatchCt == searchNumber {
				closeMatch = idx
			}
		}

		if full {
			matchCt++
			if matchCt == searchNumber {
				match = idx
				break
			}
		}

	}

	// If no "starts with" or "exact" matches are found, try and find the first item that contain the supplied name
	// Note: Can't have an exact match if there was never a close match
	if closeMatch < 0 {
		closeMatchCt = 0
		for idx, i := range items {
			part, _ := stringMatch(searchName, i, true)

			if part {
				closeMatchCt++
				if closeMatchCt == searchNumber {
					closeMatch = idx
					break
				}
			}

		}

	}

	return match, closeMatch
}
```

- [ ] **Step 5: `rooms.go`.** Add `"cmp"` and `"slices"` to the imports (neither is there today).
  `FindCorpse` and `FindCorpseByRef` (`:893-993` at `65bbc144`, both whole functions; Task 5a moved
  nothing above them) become the four functions below. Both lookups now take the viewer, the body
  typing; `livingReach` and the refusal ask Task 5a's `findByName`.

```go
// FindCorpse is FindCorpseByRef's corpse, as a copy. An ordinal that lands
// on someone living finds none.
func (r *Room) FindCorpse(viewer *characters.Character, searchName string) (Corpse, bool) {
	if c, _ := r.FindCorpseByRef(viewer, searchName); c != nil {
		return *c, true
	}
	return Corpse{}, false
}

// FindCorpseByRef finds the corpse searchName names for viewer, the body
// typing it. A name has one numbering (Calabe, 2026-11-11): first the
// living viewer perceives (Perceives), numbered as FindByName numbers them
// but without viewer itself, then the corpses that have not crumbled,
// newest first (RoundCreated; of two from one round, the later added).
// Players and NPCs follow the same rule. "rat" reaches a "rat corpse", and
// util.FindMatchIndexIn matches "<name> corpse" exactly before by a prefix.
// When the ordinal lands on someone living, corpse is nil and refusal is
// the line to tell viewer ("Bob is still alive." for a player, "The rat is
// still alive." for an NPC); the caller does not skip ahead. When no corpse
// answers to the name at all, both are empty, and the caller goes on as for
// any word that names no corpse.
func (r *Room) FindCorpseByRef(viewer *characters.Character, searchName string) (corpse *Corpse, refusal string) {

	name, number := util.GetMatchNumber(searchName)
	if name == `` {
		return nil, ``
	}

	order := []int{}
	for idx := len(r.Corpses) - 1; idx >= 0; idx-- {
		if !r.Corpses[idx].Prunable {
			order = append(order, idx)
		}
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return cmp.Compare(r.Corpses[b].RoundCreated, r.Corpses[a].RoundCreated)
	})
	names := make([]string, len(order))
	for i, idx := range order {
		names[i] = r.Corpses[idx].Character.Name + ` corpse`
	}

	if match, closeMatch := util.FindMatchIndexIn(name, names...); match < 0 && closeMatch < 0 {
		return nil, ``
	}

	// The living viewer perceives, less viewer itself (Calabe, 2026-11-11:
	// the person typing is not in their own name's living count).
	sees := func(c *characters.Character) bool {
		return (viewer == nil || c.Id != viewer.Id) && Perceives(viewer, c)
	}
	living := r.livingReach(sees, name)
	if number <= living {
		playerId, mobId := r.findByName(sees, fmt.Sprintf(`%s#%d`, name, number))
		if u := accounts.GetByUserId(playerId); playerId > 0 && u != nil {
			return nil, fmt.Sprintf(`%s is still alive.`, u.Character.Name)
		}
		if m := mobs.GetInstance(mobId); m != nil {
			return nil, fmt.Sprintf(`The %s is still alive.`, m.Character.Name)
		}
		return nil, fmt.Sprintf(`The %s is still alive.`, name)
	}

	match, closeMatch := util.FindMatchIndexIn(fmt.Sprintf(`%s#%d`, name, number-living), names...)
	if match >= 0 {
		return &r.Corpses[order[match]], ``
	}
	if closeMatch >= 0 {
		return &r.Corpses[order[closeMatch]], ``
	}
	return nil, ``
}

// livingReach is how many ordinals of name reach a living body sees allows,
// as FindByName numbers them: players and NPCs apart, each in the order
// they came into the room, a caller taking the player when both match. So
// ordinals 1 up to the larger of the two counts reach someone living.
// Asked through the same lookup, so it cannot drift from it.
func (r *Room) livingReach(sees func(*characters.Character) bool, name string) int {
	for n := 1; ; n++ {
		playerId, mobId := r.findByName(sees, fmt.Sprintf(`%s#%d`, name, n))
		if playerId == 0 && mobId.IsZero() {
			return n - 1
		}
	}
}

// RemoveCorpseRef removes the corpse c points at (one FindCorpseByRef
// returned). RemoveCorpse compares owner, name and round, so of two corpses
// alike in those (two rats killed in one round) it would take the first.
func (r *Room) RemoveCorpseRef(c *Corpse) bool {
	for idx := range r.Corpses {
		if &r.Corpses[idx] == c {
			r.Corpses = append(r.Corpses[:idx], r.Corpses[idx+1:]...)
			return true
		}
	}
	return false
}
```

  Notes for the controller's read. `livingReach` ends: ordinal n reaches someone only while n is at
  most the count of names that contain a word starting with the search (the matcher's widest pass),
  which is finite and grows with n never. The lookup re-encodes the ordinal as `name#k`; `#` is read
  before the dot, so a literal name like `1.5` survives the round trip (`1.5#1`). The viewer is left
  out only here, in the corpse count; `FindByName` keeps the viewer in its own list, so no living
  number moves for any command. `RemoveCorpse` stays (its tests use it); it has no production caller
  after Step 8. In `rooms_test.go`, `TestRoom_FindCorpse`'s four `r.FindCorpse("...")` calls
  (`:197`, `:202`, `:207`, `:212`) take `nil, ` first; they pass unchanged otherwise.

- [ ] **Step 6: `get.go`.** Its corpse source (`:91-93`) reads the new result. From
  `if c, ok := room.FindCorpseByRef(args[len(args)-1]); ok {` through `corpseRef = c` becomes:

```go
			c, refusal := room.FindCorpseByRef(actor.Character, args[len(args)-1])
			if refusal != `` {
				// The ordinal names someone living, not a corpse (one
				// numbering, the living first): refuse, do not skip ahead.
				Tell(actor, refusal)
				return true, nil
			}
			if c != nil {
				corpseRef = c
```

  (The rest of that block is unchanged. It stays behind the `CorpseItems` gate until Task 6; Task 6
  Step 11b tests it through `get`.) `get.go:422`'s `room.FindCorpse(rest)` becomes
  `room.FindCorpse(actor.Character, rest)`.

- [ ] **Step 7: `look.go`.** Two edits.
  - Inside `if len(room.Corpses) > 0 {` (`:365`), delete `:367-393` (from
    `mobCorpseLookup := map[string]int{}` through the blank line after the loop's closing brace), so
    the block opens with `if corpse, corpseFound := room.FindCorpse(actor.Character, rest); corpseFound {`
    (the viewer added). The deleted lists were built and never read
    (`grep -n 'mobCorpses\|playerCorpses\|CorpseLookup' internal/usercommands/look.go` prints only
    `:367-389` today, all assignments).
  - After `playerId, mobId := room.FindByName(actor.Character, lookAt)` (`:76`, Task 5a) add:

```go
	// Your own name reaches your corpse when one answers to it: the person
	// typing is not in their own name's living count when they look at a
	// corpse (Calabe, 2026-11-11). With none, you look at yourself, as
	// before.
	if playerId > 0 && playerId == UserIdOf(actor) {
		if _, ok := room.FindCorpse(actor.Character, lookAt); ok {
			playerId, mobId = 0, ids.ID{}
		}
	}
```

    (`UserIdOf` is the helper `consider.go:24` uses; `ids` is imported, `look.go:12`.) The look then
    goes on as for an unmatched word and reaches the corpse block. `look` needs no "still alive"
    refusal: it shows the living first, as today.

- [ ] **Step 8: `bury`.** In `modules/cleanup/cleanup.go` (`:141-163`), from
  `if corpse, corpseFound := room.FindCorpse(rest); corpseFound {` through that `if`'s closing brace
  becomes:

```go
	found, refusal := room.FindCorpseByRef(actor.Character, rest)
	if refusal != `` {
		// The ordinal names someone living, not a corpse (one numbering,
		// the living first): refuse, do not skip ahead.
		usercommands.Tell(actor, refusal)
		return true, nil
	}

	if found != nil {

		// Removing the corpse shifts the room's list, so what is said about
		// it is read first.
		corpse := *found

		// The very corpse found (of two alike, the one named; RemoveCorpse
		// would take the first).
		if room.RemoveCorpseRef(found) {

			corpseColor := `mob-corpse`
			if corpse.UserId > 0 {
				corpseColor = `user-corpse`
			}

			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Actor:        actor.Id,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ActorText:    fmt.Sprintf(`You bury the <ansi fg="%s">%s corpse</ansi>.`, corpseColor, corpse.Character.Name),
				ObserverText: fmt.Sprintf(`{actor} buries the <ansi fg="%s">%s corpse</ansi>.`, corpseColor, corpse.Character.Name),
			})
			return true, nil

		}

		return true, nil
	}
```

  All callers of the two lookups, checked: `grep -rn 'FindCorpse' --include=*.go . | grep -v _test`
  prints `get.go:92`, `get.go:422`, `look.go:394`, `cleanup.go:141` and the two definitions; each
  now passes `actor.Character`.

- [ ] **Step 9: Run; they pass.** `go test ./internal/util ./internal/rooms ./internal/items -count=1`
  and `go test ./internal/hooks -count=1` (twice: the whole package, so the file-order note above is
  exercised) print `ok` (`TestFindMatchIn`, `TestGetMatchNumber`, `TestRoom_FindCorpse` and the items
  `FindMatchIn` tests unchanged and passing; every golden byte-identical). Probes, one at a time
  (paste each failure line, restore):
  - in `GetMatchNumber` delete the dot-form `if` block: the parser test fails at `2.rat`;
  - in the dot-form condition drop `&& (name[0] < '0' || name[0] > '9')`: it fails at `1.5`;
  - in `FindCorpseByRef`'s sort swap `a` and `b` (oldest first): the two rooms numbering tests and
    `TestCorpseNumbersCountPastTheLiving` fail. `TestBuryBuriesTheCorpseItNames` does not: its two
    corpses share round 5, so the stable sort keeps them later-added first either way (that test is
    there for the `RemoveCorpse` probe below, which needs one round);
  - after `living := r.livingReach(sees, name)` add `if number <= living { number = living + 1 }`
    (skip ahead): the refusal asserts of `TestCorpseNumbersCountPastTheLiving` and the bury test's
    "the live rat is not a corpse" fail;
  - in `FindCorpseByRef`'s `sees`, drop `(viewer == nil || c.Id != viewer.Id) &&` (count the
    viewer): `TestCorpseNumbersCountPastTheLiving` fails at `ann` and `TestLookAtYourOwnNameFindsYourCorpse`
    fails (and, with Task 5c, its own-corpse refusal assert);
  - in `look.go` make the new self check `if false && ...`: `TestLookAtYourOwnNameFindsYourCorpse`
    fails;
  - in the refusal, `u.Character.Name` to `"The "+u.Character.Name` in the player line: the
    "Ratty is still alive." assert fails;
  - in `FindMatchIndexIn`, `if matchCt == searchNumber {` to `if matchCt == 1 {`: the util index test
    fails at `rat corpse#2`;
  - in `bury`, `room.RemoveCorpseRef(found)` to `room.RemoveCorpse(corpse)`: `TestBuryBuriesTheCorpseItNames`
    fails on the gold assert (its two corpses share a round);
  - `RemoveCorpseRef`'s body to `return r.RemoveCorpse(*c)`: `TestRemoveCorpseRefRemovesTheCorpseFound`
    fails on the gold assert.

  (Code and tests were run on a scratch worktree at `65bbc144` with Tasks 5a and 5c applied: all
  pass; the sort, skip-ahead, count-the-viewer, `look` self-check and `bury` probes fail as stated;
  `go test ./internal/hooks -count=1` prints `ok` twice; `go test ./... -count=1` shows only the 3
  known Windows failures, 40 ok.)

- [ ] **Step 10:** Per-task check. **Goldens: none change.** No golden or smoke input holds a
  digits-dot word (`grep` of every `w.run(` line in `reads_golden_test.go` for a dot, and of
  `cmd/smoke/script.go` for `[0-9]\.`, prints nothing), so the parser changes no recorded read. The
  only corpse lookup any golden runs is reads golden H's `bury rat`, after the rat's death has removed
  it from the room (no one living answers to `rat`) with one corpse there: the same corpse; Ann, who
  types it, is not a rat. Living lookups are `FindByName`, untouched by this task (Task 5a's viewer
  filter is the only change to them). If one changes, stop and report. Commit the ten files in
  **Files**:
  `Phase 5.3: both ordinal forms (2.rat, rat#2); one numbering per name, the living then corpses newest first, less the one typing; bury buries the corpse it names`.

### Task 5c: `bury` refuses a player's corpse that holds anything (Sonnet)

Calabe's ruling (spec, plan review round 2, 2026-11-11): `bury` refuses a player's corpse that holds
any item or gold, for every actor including its owner (NPCs run `bury` too:
`cleanup-about.html:25`); an empty player corpse and every NPC corpse bury as today. Principle: no
single command lets anyone destroy a player's belongings. The refusal text is the spec's (Decisions
taken without asking).

**Files:** modify `modules/cleanup/cleanup.go`, `internal/hooks/reads_corpse_test.go`.

- [ ] **Step 1: Write the failing test.** Append to `internal/hooks/reads_corpse_test.go` (add
  `github.com/GoMudEngine/GoMud/internal/items` to its imports; item 1 is the empty world's note,
  as reads golden H uses it). The live rat leaves the room first, so `rat` names its corpse; Ann
  names her own corpse `ann`, since she is not in her own name's count (Task 5b):

```go
// bury refuses a player's corpse that holds any item or gold, whoever asks,
// its owner too: no single command destroys a player's belongings (Calabe,
// 2026-11-11). An empty player corpse, and an NPC corpse with its loot,
// bury as before.
func TestBuryRefusesAPlayersCorpseThatHoldsAnything(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, _, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	room.RemoveMob(rat.Character.Id) // no live rat: "rat" is the corpse
	room.AddCorpse(rooms.Corpse{UserId: ann.UserId, Character: characters.Character{Name: "Ann"}, Gold: 5})
	room.AddCorpse(rooms.Corpse{UserId: 9398, Character: characters.Character{Name: "Zed"}, Items: []items.Item{items.New(1)}})
	room.AddCorpse(rooms.Corpse{UserId: 9399, Character: characters.Character{Name: "Yan"}})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, Gold: 25})

	w.run("Ann", "bury ann") // her own name reaches her corpse (Task 5b)
	w.run("Ann", "bury zed")
	w.run("Ann", "bury yan")
	w.run("Ann", "bury rat")

	out := w.out.String()
	require.Contains(t, out, `Ann: You can't bury the <ansi fg="user-corpse">Ann corpse</ansi> while it still holds something.`, "her own, holding gold")
	require.Contains(t, out, `Ann: You can't bury the <ansi fg="user-corpse">Zed corpse</ansi> while it still holds something.`, "another's, holding an item")
	require.Contains(t, out, `Ann: You bury the <ansi fg="user-corpse">Yan corpse</ansi>.`, "an empty player corpse")
	require.Contains(t, out, `Ann: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`, "an NPC corpse, loot and all")
	left := []string{}
	for _, c := range room.Corpses {
		left = append(left, c.Character.Name)
	}
	require.Equal(t, []string{"Ann", "Zed"}, left, "the two holding something are still there")
}
```

- [ ] **Step 2: Run; it fails.** `go test ./internal/hooks -run BuryRefuses -count=1`: the first
  refusal assert fails (Ann's corpse is buried).

- [ ] **Step 3: The refusal.** In `buryCommand`, after `corpse := *found` (Task 5b) and before the
  `// The very corpse found` comment, add:

```go
		// No single command destroys a player's belongings (Calabe,
		// 2026-11-11): a player's corpse that holds any item or gold is not
		// buried, whoever asks, its owner too. An NPC corpse's loot goes
		// with it.
		if corpse.UserId > 0 && (len(corpse.Items) > 0 || corpse.Gold > 0) {
			usercommands.Tell(actor, fmt.Sprintf(`You can't bury the <ansi fg="user-corpse">%s corpse</ansi> while it still holds something.`, corpse.Character.Name))
			return true, nil
		}
```

- [ ] **Step 4: Run; it passes**, with the Task 5b tests and the whole package
  (`go test ./internal/hooks -count=1`). Probes, one at a time (paste each failure line, restore):
  delete the `if` (the Ann and Zed asserts fail); `corpse.UserId > 0` alone (the Yan assert fails);
  drop `|| corpse.Gold > 0` (the Ann assert fails); drop `corpse.UserId > 0 &&` (the rat assert
  fails). (Code and test were run on the scratch worktree: both pass.)

- [ ] **Step 4b: Back-to-back burials (Calabe, plan approval 2026-11-11).** Commands and the round
  tick run one at a time on `MainWorker` under `util.LockMud()` (`world.go:316-380`), so two `bury`
  commands never interleave: the second resolves its name afresh against what is left. Append to
  `internal/hooks/reads_corpse_test.go`:

```go
// Two players bury in turn: the second bury resolves its name against what
// is left, so it takes the next match or reads the not-found line. A corpse
// that has crumbled (Prunable) is never found. Commands and the round tick
// run one at a time (MainWorker, util.LockMud), so this order is the only one.
func TestBuryAfterTheCorpseIsGone(t *testing.T) {
	w := newReadsWorld(t, 1)
	_, _, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	room.RemoveMob(rat.Character.Id) // no live rat: "rat" is a corpse
	room.AddCorpse(rooms.Corpse{UserId: 9399, Character: characters.Character{Name: "Yan"}})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 3})
	room.AddCorpse(rooms.Corpse{MobId: 8, Character: characters.Character{Name: "rat"}, RoundCreated: 5})
	room.AddCorpse(rooms.Corpse{UserId: 9398, Character: characters.Character{Name: "Xan"}, Prunable: true})

	w.run("Ann", "bury yan")
	w.run("Bob", "bury yan") // already gone
	w.run("Ann", "bury rat") // the newest rat (round 5)
	w.run("Bob", "bury rat") // the other rat (round 3)
	w.run("Ann", "bury rat") // none left
	w.run("Bob", "bury xan") // crumbled, never a candidate

	out := w.out.String()
	require.Contains(t, out, `Ann: You bury the <ansi fg="user-corpse">Yan corpse</ansi>.`)
	require.Contains(t, out, `Bob: You don't see a yan around for burying.`, "the second bury finds it gone")
	require.Contains(t, out, `Ann: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`)
	require.Contains(t, out, `Bob: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`, "the next rat, resolved afresh")
	require.Contains(t, out, `Ann: You don't see a rat around for burying.`, "no rat left")
	require.Contains(t, out, `Bob: You don't see a xan around for burying.`, "a crumbled corpse is not found")
	left := []string{}
	for _, c := range room.Corpses {
		left = append(left, c.Character.Name)
	}
	require.Equal(t, []string{"Xan"}, left, "only the crumbled corpse remains, for UpdateCorpses to remove")
}
```

  Run `go test ./internal/hooks -run 'BuryAfterTheCorpseIsGone|BuryRefuses' -count=1`: both pass
  (no code change; this pins existing behaviour). Probes, one at a time (paste each failure line,
  restore): in `bury`, skip `RemoveCorpseRef` (the "already gone" assert fails); in
  `FindCorpseByRef`, drop the `Prunable` skip (the crumbled assert fails). If a not-found line's
  wording differs from `cleanup.go`'s `You don't see a %s around for burying.`, stop and report.

- [ ] **Step 5:** Per-task check. **Goldens: none change**: reads golden H buries an NPC corpse, which
  still buries (after Task 6 it holds the rat's 25 gold and still buries). If one changes, stop and
  report. Commit both files:
  `Phase 5.3: bury refuses a player's corpse that holds anything (Calabe's ruling)`.

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
`internal/templates/templatesfunctions.go`, `internal/messaging/interface.go` (Step 3a),
`internal/guards/combat_test.go`, `internal/guards/guards_test.go`,
`_datafiles/config.yaml`, `_datafiles/html/admin/config-wizard.html`, `_datafiles/html/admin/config.html`,
`_datafiles/html/admin/users.html`, `cmd/smoke/script.go`, `cmd/smoke/main.go`,
`internal/hooks/combat_golden_test.go`, `internal/hooks/attack_engine_test.go`,
`internal/hooks/ruleset_stub_test.go`, `internal/hooks/actortick_test.go`, and the 9 goldens in Step 15;
create `internal/hooks/death_engine_test.go` and `internal/usercommands/death_test.go` (Step 11b).
`internal/usercommands/suicide_npc_test.go` stays as it is: it runs `Suicide` on an NPC, which now
goes through `death.Die`, and still passes.

- [ ] **Step 0: The knob grep's control, before any edit.** Run the deleted-knob grep of Step 13 on the
  tree as Task 5c left it, and count:
  `grep -rnwE 'ExtraLives|LivesStart|LivesMax|LivesOnLevelUp|PermaDeath|ProtectionLevels|XPPenalty|EquipmentDropChance|AlwaysDropBackpack|CorpseItems|LivesGained|f-extralives' --exclude-dir=.git --exclude-dir=.claude --include=*.go --include=*.html --include=*.yaml --include=*.template . | wc -l`
  prints 129, in 19 files (`... | cut -d: -f1 | sort -u | wc -l`): 117 lines in 18 files at
  `65bbc144` (counted 2026-11-11), plus Task 1's 12 in `combat_golden_test.go` (`deathConfig`'s
  comment, set and restore maps and `XPPenalty` assert, and the two `// above ProtectionLevels 1`
  comments). Another count: stop and report (the grep that later proves absence must be shown able to
  find these).

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

	Awards []Award // the killers' shares, in order; their Messages are the outcome's lines
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
		// Only the body's presence: none for a viewer who does not perceive
		// it (Task 5a, rooms.NotPerceiving).
		Exclude: rooms.NotPerceiving(victim.Character),
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

	if len(c.PlayerDamage) > 0 {
		c.KD.AddPvpDeath()
	} else {
		c.KD.AddMobDeath()
	}

	killedByUserIds := []int{}
	killers := []*accounts.Account{}
	for uid := range c.PlayerDamage {

		if u := accounts.GetByUserId(uid); u != nil {

			// Update PK stats
			c.KD.AddPlayerDeath(u.UserId, u.Character.Name)
			u.Character.KD.AddPlayerKill(user.UserId, c.Name)
			killers = append(killers, u)
		}

		killedByUserIds = append(killedByUserIds, uid)
	}

	// The notice names, to each reader, only the killers that reader
	// perceives; with none left it names no killer at all ("someone"
	// would tell of an unseen killer). Plan review round 5, M1.
	messaging.NotifyEach(func(reader *accounts.Account) string {
		named := []string{}
		for _, k := range killers {
			if namesKiller(reader, k.Character) {
				named = append(named, `<ansi fg="username">`+k.Character.Name+`</ansi>`)
			}
		}
		return deathNotice(c.Name, named)
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

// namesKiller reports whether the death notice names killer to reader: the
// one visibility rule (rooms.Perceives), from where the reader is. Seeing a
// hidden body takes being in its room, unless the reader is an admin, who
// perceives hidden bodies anywhere; an unhidden killer is named to all, as
// today.
func namesKiller(reader *accounts.Account, killer *characters.Character) bool {
	if reader.Character.RoomId != killer.RoomId && killer.HasBuffFlag("hidden") {
		return reader.Role == accounts.RoleAdmin
	}
	return rooms.Perceives(reader.Character, killer)
}

// deathNotice is the world notice of name's death, naming the killers given
// ("A", "A and B", "A, B and C"); with none it names no killer.
func deathNotice(name string, killers []string) string {
	if len(killers) == 0 {
		return fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, name, term.CRLFStr)
	}
	killedBy := ``
	for i, k := range killers {
		if i > 0 {
			if i < len(killers)-1 {
				killedBy += `, `
			} else {
				killedBy += ` and `
			}
		}
		killedBy += k
	}
	return fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> (killed by %s) <ansi fg="magenta-bold">***</ansi>%s`, name, killedBy, term.CRLFStr)
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

- [ ] **Step 3a: `messaging.NotifyEach`** (plan review round 5, M1: the death notice is read per
  receiver). Today the notice is one `messaging.Notify`, an `events.Broadcast` of one string that
  `WriteBroadcast` (`wire.go:104-161`) writes to every active account; it cannot vary by reader. Each
  reader instead gets its own `events.Message` (`UserId`), which `WriteMessage` (`wire.go:24-45`)
  writes with the same line refresh (`AnsiMoveCursorColumn` and `AnsiEraseLine`) and prompt redraw
  (priority 100) as the broadcast; a screen-reader account gets the text stripped
  (`util.StripCharsForScreenReaders`), as every message is (the broadcast sent it unstripped when no
  screen-reader text was given). In `internal/messaging/interface.go` add `"sort"` to the imports and,
  after `Notify` (`:66-72`):

```go
// NotifyEach is a notice whose text depends on who reads it (the death
// notice names only the killers each reader perceives). Every active
// account, in UserId order, gets text(account) as its own message: the same
// line refresh and prompt redraw as Notify's broadcast, and stripped for a
// screen reader as every message is. An empty text sends that account
// nothing.
func NotifyEach(text func(u *accounts.Account) string) {
	all := accounts.GetAllActiveUsers()
	sort.Slice(all, func(i, j int) bool { return all[i].UserId < all[j].UserId })
	for _, u := range all {
		if t := text(u); t != `` {
			events.AddToQueue(events.Message{UserId: u.UserId, Text: t})
		}
	}
}
```

  The goldens' recorder writes a `UserId` message and a broadcast alike (`r.read`, `reads_golden_test.go:143-196`),
  and both expand in `UserId` order, so a notice that names the same killers to everyone records
  exactly as before (run on a scratch copy with today's `suicide.go` sending its notice this way:
  `go test ./... -count=1` showed only the 3 known Windows failures; reads golden H's notice lines
  unchanged).

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
  (After Task 5b the `look.go` lines are `:385-387`: find them by their text.)

  **The corpse panel's fallback (spec, plan review round 2, blocker).** `buildCorpseInventoryPanel`
  loads `character/equipment-look`, which no data folder has (`_datafiles/world/empty/panel-layouts/`
  holds only `character/inventory.yaml`), so its fallback always runs; the fallback builds a panel
  named `inv` while every row is written to `equip`, and `Panel("equip")` then logs
  `no panel with id "equip"` and returns a throwaway panel: `look <corpse>` shows an empty box.
  Unreachable while `CorpseItems` was off. In `look.panels.go:115-116` the two lines
  `layout.AddPanelsToSlot(layout.AddSlot(), "inv")` and
  ``layout.Panel("inv").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Inventory</ansi> `).SetWidth(78)``
  become (as `buildInventoryLookPanel`'s fallback, `:57-59`):

```go
		layout.AddPanelsToSlot(layout.AddSlot(), "equip")
		layout.Panel("equip").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Inventory</ansi> `).SetWidth(78)
```

  `TestLookAtACorpseShowsWhatItHolds` (Step 11b) is its test. (Run on the scratch worktree with
  `CorpseItems` forced on: before the edit the panel is an empty box and the log reads
  `no panel with id "equip"`; after it the item is listed.)

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
	if args[0] == "all" && !namesACorpse(actor, room, args) {
```

  and after `Get` (after its closing brace, `:435`) add:

```go
// namesACorpse reports whether args is "all <word>" or "all from <word>"
// with <word> a corpse's name for actor: it finds a corpse, or its ordinal
// lands on someone living whom a corpse shares a name with (the corpse
// branch below then refuses, "The rat is still alive."; Task 5b). A
// container of the same name wins, as it does for every get (the container
// lookup comes first, :80-88), so then this is false and "get all" keeps
// its room-wide meaning.
func namesACorpse(actor *actors.Actor, room *rooms.Room, args []string) bool {
	if len(args) != 2 && !(len(args) == 3 && args[1] == "from") {
		return false
	}
	word := args[len(args)-1]
	if room.FindContainerByName(word) != `` {
		return false
	}
	c, refusal := room.FindCorpseByRef(actor.Character, word)
	return c != nil || refusal != ``
}
```

  Nothing else in `get.go` changes: for a corpse word the lookup below finds the same corpse, or
  refuses for the living (Task 5b's block, now without the `CorpseItems` gate), strips an optional
  `from` and leaves `rest == "all"` for the corpse branch. A word no corpse answers to returns
  nothing from the lookup, so every other `get all ...` takes the same branch as today (reads golden D
  runs plain `get all`, `reads_golden_test.go:705`, and must stay byte-identical). `FindCorpseByRef`
  is Task 5b's (one numbering per name: the living first, then corpses newest first);
  `FindContainerByName` follows it in `rooms.go` (`:1458` after Task 5b). The tests are in Step 11b
  (`TestGetFromAPlayersCorpse`, `TestGetAllWithoutACorpseTakesTheFloor`,
  `TestGetAllCountsCorpsesPastTheLiving`).

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
  - `_datafiles/html/admin/users.html`: delete the Extra Lives field (`:366-370`, the
    `<div class="field">` holding `<label>Extra Lives</label>` through its `</div>`), `:592`
    (`setVal('f-extralives', ...)`) and `:1089` (`ExtraLives: ...`).
  - The absence proof for these names runs in Step 13, after Step 11 (`deathConfig` and the death
    fixtures' comments in `combat_golden_test.go`) and Step 12 (the smoke comment naming
    `ProtectionLevels`) clear the last hits; run here it would still find them.

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

    The two player fixtures' comments describe the new death. `TestCombatGoldenDeathPlayer`'s
    comment (from `// An elite NPC kills a player above the protection band` through
    `// books an elite death and leaves a corpse; she wakes in the recovery room.`) becomes:

```go
// An elite NPC kills a player (Phase 5 slice 3; recorded on the old death
// first, Task 1). She dies in the round her health reaches 0; the gem she
// carries and her gold go into her corpse; she keeps what she wears and her
// XP; the elite death is booked; she wakes in the recovery room at 1 health.
```

    `TestCombatGoldenDeathPvP`'s comment (from `// A player kills a player (Phase 5 slice 3, recorded on the death before it`
    through `// penalty, corpse, recovery room.`) becomes:

```go
// A player kills a player (Phase 5 slice 3; recorded on the old death first,
// Task 1): the kill credit and the "killed by" notice; the gem Bob carries
// and his gold go into his corpse; he keeps what he wears and his XP; he
// wakes in the recovery room at 1 health.
```

    In both, the line comment `// above ProtectionLevels 1` on the `Level, Experience = 2, 1000` line
    becomes `// the record shows no XP is lost` (the line stays). `TestKilledByOutlivesTheKiller`
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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
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

// The revive line is only the body's presence (Task 5a's rule, plan review
// round 5): a viewer who does not perceive the hidden body reads nothing,
// not even "someone"; an admin reads it.
func TestReviveOfAHiddenBodyReachesOnlyThoseWhoPerceive(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	eve := w.addPlayer(9303, "Eve", goldenRoomA, 10, 10, 10, 12)
	eve.Role = accounts.RoleAdmin
	w.flag(alice.Character, 93903, "revive-on-death")
	w.hide(alice.Character)
	w.start()

	alice.Character.Health = 0
	death.Die(actors.Get(alice.Character.Id))
	events.ProcessEvents()
	_, reads := w.taken()
	require.Positive(t, alice.Character.Health, "revived")
	require.Contains(t, strings.Join(reads[eve.UserId], "\n"), "is suddenly revived in a shower of sparks!", "an admin perceives the hidden body")
	require.NotContains(t, strings.Join(reads[bob.UserId], "\n"), "is suddenly revived", "Bob does not perceive Alice: no line, not even as someone")
}

// The world death notice names, to each reader, only the killers that reader
// perceives (plan review round 5, M1): a hidden killer (Alice) only to an
// admin anywhere (Eve), a see-hidden reader in her room (Dan) and herself;
// never "someone". An unhidden killer (Gus, in room B) is named to all.
func TestDeathNoticeNamesOnlyTheKillersEachReaderPerceives(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	carol := w.addPlayer(9303, "Carol", goldenRoomB, 10, 10, 10, 12)
	dan := w.addPlayer(9304, "Dan", goldenRoomA, 10, 10, 10, 12)
	eve := w.addPlayer(9305, "Eve", goldenRoomB, 10, 10, 10, 12)
	fay := w.addPlayer(9306, "Fay", goldenRoomB, 10, 10, 10, 12)
	gus := w.addPlayer(9307, "Gus", goldenRoomB, 10, 10, 10, 12)
	eve.Role = accounts.RoleAdmin
	w.flag(dan.Character, 93905, "see-hidden")
	w.flag(fay.Character, 93906, "see-hidden")
	w.hide(alice.Character)
	w.start()

	die := func(u *accounts.Account, damage map[int]int) map[int][]string {
		u.Character.PlayerDamage = damage
		u.Character.Health = 0
		death.Die(actors.Get(u.Character.Id))
		events.ProcessEvents()
		_, reads := w.taken()
		return reads
	}
	notice := func(reads map[int][]string, u *accounts.Account) string {
		for _, r := range reads[u.UserId] {
			if strings.Contains(r, "DIED!") {
				return r
			}
		}
		require.Failf(t, "no notice", "%s read no death notice", u.Character.Name)
		return ""
	}
	aliceName, gusName := `<ansi fg="username">Alice</ansi>`, `<ansi fg="username">Gus</ansi>`

	reads := die(bob, map[int]int{alice.UserId: 5, gus.UserId: 3})
	for _, u := range []*accounts.Account{alice, dan, eve} {
		n := notice(reads, u)
		require.Contains(t, n, aliceName, "%s perceives the hidden Alice", u.Character.Name)
		require.Contains(t, n, gusName, u.Character.Name)
	}
	for _, u := range []*accounts.Account{bob, carol, fay, gus} {
		n := notice(reads, u)
		require.NotContains(t, n, aliceName, "%s does not perceive the hidden Alice", u.Character.Name)
		require.Contains(t, n, "(killed by "+gusName+")", "%s: only the killer it perceives", u.Character.Name)
		require.NotContains(t, strings.ToLower(n), "someone", u.Character.Name)
	}

	// With the only killer hidden, a reader who does not perceive her reads
	// no "(killed by ...)" at all; an admin reads her name.
	reads = die(gus, map[int]int{alice.UserId: 5})
	require.NotContains(t, notice(reads, carol), "(killed by", "nothing to name, so no killed by")
	require.Contains(t, notice(reads, eve), "(killed by "+aliceName+")")
}

// A respawn starts clean: no buffs, no mana (Q7c), no damage record, no
// killer record; an elite killer is booked as an elite death first. The
// PlayerDeath event names who killed her, and her event log records it.
func TestRespawnStartsClean(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	w.flag(alice.Character, 93904, "testmark") // any flagged buff: buffs.All cancels flagged buffs
	w.start()
	var deaths []events.PlayerDeath
	id := events.RegisterListener(events.PlayerDeath{}, func(e events.Event) events.ListenerReturn {
		deaths = append(deaths, e.(events.PlayerDeath))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.PlayerDeath{}, id) })
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

	require.Len(t, deaths, 1)
	require.Equal(t, 93202, deaths[0].KillerMobId, "the killer NPC")
	require.Equal(t, []int{9302}, deaths[0].KilledByUsers, "the accounts credited with damage")
	require.Equal(t, goldenRoomA, deaths[0].RoomId, "where she died, not where she woke")
	require.NotEmpty(t, alice.EventLog)
	last := alice.EventLog[len(alice.EventLog)-1]
	require.Equal(t, "death", last.Category)
	require.Contains(t, last.What, "DIED", "the event log records the death")
}

// An NPC's ending frees its spawn slot in its home room (CleanupMobSpawns),
// so the slot can spawn again.
func TestDeathFreesTheNPCsSpawnSlot(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: the ending ledger and the room's NPC file are written there
	dataFolder := configs.GetFilePathsConfig().DataFiles.String()
	t.Cleanup(func() { // registered after recoveryRoom's cleanups, so it runs first, in the temporary folder
		require.NoError(t, os.RemoveAll(filepath.Join(dataFolder, "rooms.instances")))
		require.NoError(t, rooms.LoadEndings())
	})
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "slot-rats")
	rat.SpawnSlot = 1 // a slot NPC: rooms.NPCEnded records its ending
	home := rooms.LoadRoom(goldenRoomA) // addNPC makes the room it is put in its home (HomeRoomId)
	home.SpawnInfo = []rooms.SpawnInfo{{MobId: 93201, InstanceId: rat.Character.Id}}
	w.start()

	death.Die(actors.Get(rat.Character.Id))
	events.ProcessEvents()

	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the rat ended")
	require.True(t, home.SpawnInfo[0].InstanceId.IsZero(), "its slot is free (CleanupMobSpawns)")
}
```

  Check before relying on them: `buffs.All` cancels only buffs that carry a flag
  (`buffs.go:104-139`, the loop over `bSpec.Flags`), hence `testmark`; an unknown flag only logs a
  warning there. `KDStats.EliteDeaths` keys are `<mobId>:<name>` (`kdstats.go:104-110`). Run
  `go test ./internal/hooks -run 'Upkeep|BuffTrigger|CorpsesOff|ReviveOnDeath|RespawnStartsClean|SpawnSlot' -count=1 -v`;
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
    assert fails); delete the `CharacterVitalsChanged` queue (the vitals assert fails); delete the
    `Exclude: rooms.NotPerceiving(victim.Character),` line (`TestReviveOfAHiddenBodyReachesOnlyThoseWhoPerceive`
    fails on Bob, "not even as someone");
  - death notice (round 5, M1): make `namesKiller` return `true` first (`TestDeathNoticeNamesOnlyTheKillersEachReaderPerceives`
    fails: Bob, Carol and Fay read Alice); replace its far condition with `false` (it fails on Fay,
    a see-hidden reader in another room); make the far branch `return false` (it fails on Eve, the
    admin). Each of the three ran against today's `suicide.go` sending the notice this way, on a
    scratch copy, with the same test written for the reads world: each failed as stated;
  - respawn: in `respawn` delete `c.CancelBuffsWithFlag(buffs.All)` (the buff assert fails); delete
    `c.Mana = 0` (the mana assert fails); delete `clear(c.PlayerDamage)` (the damage assert fails); in
    `recordPlayerDeath` delete `c.KilledBy = ids.ID{}` (the killer assert fails); delete
    `c.KD.AddEliteDeath(killerMobId, c.KillerMobName)` (the elite assert fails); in its
    `events.PlayerDeath{...}` delete `KillerMobId:   killerMobId,` (the killer NPC assert fails), then
    `KilledByUsers: killedByUserIds,` (the credited accounts assert fails); delete the
    `user.EventLog.Add(...)` line (the event-log asserts fail);
  - spawn slot: in `end` delete the `if r := rooms.LoadRoom(mob.HomeRoomId); r != nil { r.CleanupMobSpawns() }`
    block (`TestDeathFreesTheNPCsSpawnSlot`'s slot assert fails).

  (b) Create `internal/usercommands/death_test.go`. It holds the recovery-room helper Task 7's
  `suicide` tests also use:

```go
package usercommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
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
// the floor and leaves Bob's corpse alone. So does a word that also names a
// container ("bob" begins "bobchest"): the container wins, as for every get.
func TestGetAllWithoutACorpseTakesTheFloor(t *testing.T) {
	for _, tc := range []struct{ command, container string }{
		{"all ground", ""}, {"all junk", ""}, {"all from junk", ""}, {"all junk bob", ""},
		{"all bob", "bobchest"},
	} {
		t.Run(tc.command+" "+tc.container, func(t *testing.T) {
			npc, player, room := goSetup(t)
			recoveryRoom(t, testNextRoomId)
			corpsesOn(t)
			player.Character.Name = "Bob"
			player.Character.Items = []items.Item{testItem(92730, "testgem", items.ItemSpec{Type: items.Gemstone})}
			death.Die(player)
			require.Len(t, room.Corpses, 1, "Bob's corpse is in the room")
			room.AddItem(testItem(92732, "teststick", items.ItemSpec{Type: items.Object}), false)
			room.Gold = 5
			if tc.container != "" {
				room.Containers = map[string]rooms.Container{tc.container: {}}
				t.Cleanup(func() { room.Containers = nil })
			}

			_, err := Get(tc.command, npc, room, 0)
			require.NoError(t, err)
			assert.Equal(t, []int{92732}, itemIdsOf(npc.Character.Items), "the floor's item, not the corpse's")
			assert.Equal(t, 5, npc.Character.Gold, "the floor's gold, not the corpse's")
			assert.Empty(t, room.Items, "the floor is empty")
			assert.Equal(t, []int{92730}, itemIdsOf(room.Corpses[0].Items), "the corpse is untouched")
			assert.Equal(t, 25, room.Corpses[0].Gold)
		})
	}
}

// The corpse panel shows what a corpse holds (look no longer gates it on a
// config setting).
func TestLookAtACorpseShowsWhatItHolds(t *testing.T) {
	npc, player, room := goSetup(t)
	corpsesOn(t)
	mobs.GetInstance(npc.Id).MobId = 92801 // a real NPC's template ID, which its corpse records (death.Die)
	npc.Character.Items = []items.Item{testItem(92731, "testpebble", items.ItemSpec{Type: items.Gemstone})}
	death.Die(npc)
	require.Len(t, room.Corpses, 1)

	told := playerTexts(t, player, func() {
		_, err := Look("rat corpse", player, room, 0)
		require.NoError(t, err)
	})
	assert.Contains(t, strings.Join(told, ""), "testpebble", "the panel lists the corpse's item")
}

// One numbering per name through get (Task 5b): goSetup's live NPC "Rat"
// is 1, so "get all rat" is refused rather than skipped ahead; 2 is the
// newest rat corpse and 3 the older, in both ordinal forms.
func TestGetAllCountsCorpsesPastTheLiving(t *testing.T) {
	_, player, room := goSetup(t)
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 1})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 7, Gold: 2})
	gold := player.Character.Gold

	told := playerTexts(t, player, func() {
		_, err := Get("all rat", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"The Rat is still alive.\n"}, told, "1 is the living Rat")
	assert.Equal(t, gold, player.Character.Gold, "nothing taken")

	_, err := Get("all 2.rat", player, room, 0)
	require.NoError(t, err)
	assert.Equal(t, gold+2, player.Character.Gold, "2 is the newest corpse")
	_, err = Get("all rat#3", player, room, 0)
	require.NoError(t, err)
	assert.Equal(t, gold+3, player.Character.Gold, "3 is the older")
}
```

  Check before relying on them: `goSetup`'s NPC is a bare test NPC with 0 gold and no items, so what
  it holds afterwards came from the corpse; `StoreItem` has no capacity check
  (`character.go:538-548`); corpse lookup is case-insensitive (`util.go` `stringMatch`). The look test
  exercises the panel's fallback: no data folder has the `character/equipment-look` layout, so the
  fallback always runs (Step 7's fix). Run
  `go test ./internal/usercommands -run 'Corpse' -count=1 -v`; all four pass. (`get all ground`: today's code strips `ground` and runs the room-wide branch only because
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
  - in `look.go` delete `Tell(actor, buildCorpseInventoryPanel(&corpse))`: the look test fails;
  - in `look.panels.go` put back the fallback's `"inv"` (both lines, Step 7): the look test fails;
  - in `namesACorpse` delete the container check (``if room.FindContainerByName(word) != `` { return false }``):
    the `all bob bobchest` subtest fails (the taker gets nothing from the floor);
  - in `get.go` replace Task 5b's ``if refusal != `` { ... }`` block with `_ = refusal` (deleting it
    alone leaves `refusal` unused, and `get.go` does not compile): `TestGetAllCountsCorpsesPastTheLiving`
    fails on its refusal assert (no "still alive" line);
  - in `namesACorpse` return `c != nil` only (write `_ = refusal; return c != nil` so it compiles):
    the same test's refusal assert fails (`get all rat`
    takes the room-wide branch instead).

  (`TestGetAllCountsCorpsesPastTheLiving` was not run while writing this revision: it needs Task 6's
  `get` changes. The lookup it calls was run, Task 5b.)

  (c) The recovery copies are freed, scoped to the tests that make deaths (the whole `hooks` package
  would run `TestCombatGoldenFleeSuccess`, which fails on repeats on the baseline too: `status.md`
  Watch):
  `go test ./internal/hooks -run 'Death|CombatGolden(MvP|PvP|HiddenAttacker)$|Wimpy|Respawn' -count=16`
  prints `ok` (about ten player deaths a run: `MvP`, `PvP`, `HiddenAttacker`, `DeathPlayer`, `DeathPvP`,
  the wimpy test, upkeep, corpses off, revive, respawn; 16 runs would exhaust the 100 ephemeral chunks
  without `freeCopies`), and `go test ./internal/usercommands -run 'Corpse' -count=16` prints `ok`.
  Check the pattern first: `go test ./internal/hooks -list '<the same pattern>'` lists no
  `TestCombatGoldenFleeSuccess` (`-list` takes the pattern itself; it ignores `-run`). At `65bbc144`
  it lists the two wimpy tests, `PvMToDeath`, `PvP`, `MvP` and `HiddenAttacker`; by this step it also
  lists the three death goldens (Task 1) and this task's death engine tests.

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
	// from its own corpse proves the drop went into the corpse (its own name
	// reaches its corpse: the one typing is not in its own name's count).
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
  `CorpsesEnabled: true`. If the get wording differs, use what `get.go` sends and say so. The bot
  stands in the start room when it takes the ring (it walked out through the portal); its own name
  skips itself (Task 5b) and reaches its newest corpse. Other bots in the room do not answer to it:
  names are fixed width, so none is a prefix of another. A bot that is hidden is not counted either
  (Task 5a); the smoke bots never hide.

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

  Then the deleted knobs are gone (moved here from Step 8, which ran before Steps 11 and 12 cleared
  their hits). Whole words only (`-w`, so `get.go`'s `allCorpseItems`, a variable that stays, does not
  match):
  `grep -rnwE 'ExtraLives|LivesStart|LivesMax|LivesOnLevelUp|PermaDeath|ProtectionLevels|XPPenalty|EquipmentDropChance|AlwaysDropBackpack|CorpseItems|LivesGained|f-extralives' --exclude-dir=.git --exclude-dir=.claude --include=*.go --include=*.html --include=*.yaml --include=*.template .`
  prints nothing. Controls: Step 0 showed the same command printing 129 lines in 19 files before this
  task; now the same command with `CorpsesEnabled` in place of the list prints hits.
  The lower-case word is checked apart, because comments may still say it:
  `grep -rnw 'permadeath' --exclude-dir=.git --exclude-dir=.claude --include=*.go --include=*.html --include=*.yaml --include=*.template .`
  prints exactly these comment lines and nothing else: `internal/actors/registry.go:169` and `:180`
  ("a permadeath reroll", Q7m), and in the new `internal/death/death.go` the `recordPlayerDeath`
  comment (`suicide.go before slice 3, less permadeath`) and the `Die` comment (`there is no
  permadeath (decision 11)`). A hit in code (a template func, a config key, a template name) is a
  miss: fix it.

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
    world notice `*** <name> has DIED! ***`, with no `\r` at the line end. In `pvp` every player's
    notice reads `(killed by Alice)`, as today's did. In `hidden_attacker` Alice is hidden and no one
    there is an admin or has `see-hidden`, so only Alice's own notice reads `(killed by Alice)`; Bob's
    and Carol's read `*** Bob has DIED! ***` with no `(killed by ...)` part and no "someone" (plan
    review round 5, M1; the notice is three `UserId` messages there, recorded as before); the other players in room A
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
    "suicide"` line, the Drop lines and the Remove lines (`death_player`: `You remove your <ansi fg="item">testsword</ansi> and return it to your backpack.`
    and `You drop the <ansi fg="item">testsword</ansi>.`) and `You lost 100 experience points` go, and so do the
    `room 93001 floor:` and `floor gold:` lines. The `ItemOwnership <name> item=N gained=false` notes
    the old `drop` queued (`drop.go:108`, one per item its rolls dropped, in the heal section that ran
    `suicide`; `death_player`'s is `ItemOwnership Alice item=93101`, the testsword) are replaced by exactly one, `ItemOwnership <name> item=93108 gained=false` (the carried
    gem, queued by `takeDrops`), in the death round after `PlayerDeath <name>` and before the respawn's
    `CharacterVitalsChanged <name>`; no worn item's note appears (worn gear stays). In `death_pvp` Alice reads
    `Bob has died.` in that round rather than in a heal section. The dead player's state line: no
    `items=` (the gem is in the corpse, whatever its 50% roll did today), `gold=0` (in the corpse,
    not on the floor), `xp=1000` (no penalty), every worn item kept (`weapon=93101` is back in
    `death_player`, where today's record lost it to the drop; `worn=` as at the start), the same `kd=` as today (`0/1/0/0/0/1` for Alice in
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

**Files:** modify `internal/usercommands/suicide.go`, `internal/usercommands/usercommands.go` (the
registry comment, Step 3b); create `internal/usercommands/suicide_test.go`.

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

- [ ] **Step 3b: The registry comment says how the two rules meet** (spec, plan review round 2).
  `suicide` stays `AllowedWhenDowned` (`{Suicide, true, false}`), and the in-fight refusal runs
  before anything else, so under a ruleset with a downed state a downed actor can end it only once
  nobody is fighting it. In `usercommands.go`, the comment Task 6 Step 4 wrote above the `suicide`
  entry (`// Death by one's own hand (death.Die). Allowed when downed, so a downed actor` /
  `// can end it under a ruleset with a downed state.`) becomes:

```go
	// Death by one's own hand (death.Die). Allowed when downed, so under a
	// ruleset with a downed state a downed actor can end it, though only once
	// nobody is fighting it: Suicide refuses in a fight first.
```

- [ ] **Step 4: Run; they pass.** Probes: delete the `inFight` check (the fight test fails); replace
  `rooms.IsDeathRecoveryRoom(actor.Character.RoomId)` with
  `actor.Character.RoomId == int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)` (add the `configs`
  import for the probe; the copy test fails, the room test passes); delete the recovery check, adding
  `_ = errors.New` so the import stays used (both recovery tests fail); delete `death.Die(actor)`,
  adding `_ = death.Die` (`TestSuicideDiesOtherwise` fails). Paste the four
  lines.
- [ ] **Step 5:** Per-task check (goldens unchanged: no golden runs `suicide` in a fight or the recovery
  room). `go test ./internal/usercommands -run Suicide -count=16` prints `ok` (each test's copies are
  freed). Commit the three files in **Files**: `Phase 5.3: suicide is refused in a fight and in the recovery room (and its copies)`.

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

// deathPrefix is the deathScanned prefix path is under, or "".
func deathPrefix(path string) string {
	for _, p := range deathScanned {
		if strings.HasPrefix(path, p) {
			return p
		}
	}
	return ""
}

// TestEngineAsksTheRulesetWhoIsDead enforces Phase 5 slice 3: the engine
// compares no health with a number and never asks a body whether it is
// disabled; it asks rules.Get().Condition.
func TestEngineAsksTheRulesetWhoIsDead(t *testing.T) {
	files, healthCompares := 0, 0
	reached := map[string]int{}
	itemCalls := map[string]bool{}
	for path, f := range goFiles(t) {
		p := deathPrefix(path)
		if p == "" {
			continue
		}
		reached[p]++
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
	// ... every prefix reached a file (a typo in one, or a package that
	// moved, would otherwise go unseen behind the other two) ...
	for _, p := range deathScanned {
		if reached[p] == 0 {
			t.Errorf("control: no file under %s; a scope prefix reaches nothing", p)
		}
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
  - `files < 50` to `files < 5000`: the walk control fails;
  - in `deathScanned`, `"internal/death/"` to `"internal/deaths/"`: the per-prefix control fails
    (`no file under internal/deaths/`) though the walk control still passes (`hooks` has 22 non-test
    files and `usercommands` 68 at `65bbc144`).
- [ ] **Step 4:** Per-task check; `git status --short` shows only `internal/guards/death_test.go`.
  Commit it: `Phase 5.3: guards: the engine asks the ruleset who is dead (no health literal, no body IsDisabled); nothing queues suicide`.

### Task 9: Stub-ruleset tests (Sonnet)

Test files only: a ruleset can add a downed state, revive, fix the drops, set the respawn and set
regen with no engine edit, and cannot drop what the body lacks.

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

// respawner is the active ruleset where every death respawns the body at 4
// health.
type respawner struct{ rules.Ruleset }

func (respawner) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{Respawn: true, RespawnHealth: 4}
}

// The engine respawns at the ruleset's health (not a fixed 1), and respawns
// an NPC too when the ruleset says so: its body does not end.
func TestStubRulesetDecidesTheRespawn(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "respawn-rats")
	t.Cleanup(rules.SetForTest(respawner{rules.Get()}))
	w.start()

	for _, c := range []*characters.Character{alice.Character, &rat.Character} {
		c.Health = 0
		death.Die(actors.Get(c.Id))
		events.ProcessEvents()
		require.Equal(t, 4, c.Health, "%s respawns at the ruleset's health", c.Name)
		require.Equal(t, goldenRoomC, rooms.GetOriginalRoom(c.RoomId), "%s is in a copy of the recovery room", c.Name)
	}
	require.NotNil(t, mobs.GetInstance(rat.Character.Id), "the rat's body did not end")
}

// overDrops is the active ruleset whose drops ask for more than the body
// has: an item it does not hold, and gold.
type overDrops struct {
	rules.Ruleset
	drops []items.Item
	gold  int
}

func (o overDrops) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{DropItems: o.drops, DropGold: o.gold}
}

// The engine drops only what the body has: an item it does not hold is not
// made from nothing, and the gold is at most the body's and never below 0.
func TestStubRulesetCannotDropWhatTheBodyLacks(t *testing.T) {
	gem := trinket(93108, "testgem", items.Gemstone, items.Mundane)
	phantom := trinket(93109, "testpebble", items.Gemstone, items.Mundane) // never held
	for _, tc := range []struct{ gold, wantCorpse, wantBody int }{
		{1000, 30, 0}, // more than it has: all of it
		{-5, 0, 30},   // below 0: none
	} {
		t.Run(fmt.Sprint(tc.gold), func(t *testing.T) {
			w := newCombatWorld(t, 111)
			deathConfig(t) // corpses on
			rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "overdrop-rats")
			rat.Character.Items = []items.Item{gem}
			rat.Character.Gold = 30
			t.Cleanup(rules.SetForTest(overDrops{Ruleset: rules.Get(), drops: []items.Item{gem, phantom}, gold: tc.gold}))
			w.start()

			death.Die(actors.Get(rat.Character.Id))
			events.ProcessEvents()

			room := rooms.LoadRoom(goldenRoomA)
			require.Len(t, room.Corpses, 1)
			require.Equal(t, "93108", itemIds(room.Corpses[0].Items), "the held gem, not the phantom")
			require.Equal(t, tc.wantCorpse, room.Corpses[0].Gold)
			require.Equal(t, tc.wantBody, rat.Character.Gold)
		})
	}
}
```

  Add `"strings"` and `github.com/GoMudEngine/GoMud/internal/death` to the imports (`fmt`, `actors`,
  `characters`, `items`, `mobs` and `rules` are there already: `ruleset_stub_test.go:3-15`). If the
  respawned rat (an NPC in a recovery copy with no player, freed by `freeCopies`) upsets a later
  test, report it; do not drop the NPC half. (`downedAtZero` is in `combat_golden_test.go`, Task 6; `trinket`,
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
  - respawn: in `death.respawn` change `c.Health = health` to `c.Health = 1` (Alice's health assert
    fails); in `death.Die` change `if !outcome.Respawn && mob != nil {` to `if mob != nil {` (the rat
    ends: its health and instance asserts fail).
  - what the body lacks: in `takeDrops` change `gold := min(max(outcome.DropGold, 0), c.Gold)` to
    `gold := outcome.DropGold` (both subtests fail), then to `gold := min(outcome.DropGold, c.Gold)`
    (the `-5` subtest fails); change `if c.RemoveItem(itm) || c.RemoveFromBody(itm) {` to
    `if c.RemoveItem(itm) || c.RemoveFromBody(itm) || itm.ItemId > 0 {` (the phantom lands in the
    corpse: the item assert fails).
  - downed actor (usercommands): `usercommands.go:281` to
    `userDisabled = actor.Character.IsDisabled() && rules.Get() != nil` (asks the body again; the
    `rules.Get()` call keeps the file's only `rules` use, so it compiles; the `get` assert fails);
    `go.go:46` to `if actor.Character.IsDisabled() && rules.Get() != nil {` (the same; the `north` assert
    fails).
- [ ] **Step 5:** Per-task check. Commit the two test files:
  `Phase 5.3: stub rulesets add a downed state, revive, fix the drops, set the respawn and set regen with no engine edit`.
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
  In `get.md`, end the stash line (`  This would get a stick stashed in the area and put it in your backpack.`)
  with two spaces, as every other usage line ends: `internal/markdown` joins a line to the next unless
  it ends in two spaces, so without them the stash line and `~get all bob~` render as one line
  (rendered through `markdown.NewParser` on a scratch copy, plan review round 5). Then, after that
  usage pair, add:

```text
  ~get all bob~  
  This would take everything from Bob's corpse (~get all from bob~ works too).  
  ~get stick bob~  
  This would take a stick from Bob's corpse.

A corpse holds what its owner dropped until it crumbles to dust. When it crumbles, everything still in it is destroyed.
```

  and, after that paragraph, a blank line and then this one:

```text
A name counts the living first, then corpses from the newest. With a live rat and two rat corpses here, ~get all 2.rat~ takes everything from the newest corpse and ~get all 3.rat~ from the older (~rat#2~ works too); ~get all rat~ is refused, since the rat is alive. Your own name skips you, so ~get all~ and your name finds your newest corpse, unless someone here has a name that starts with yours (Bobby, for Bob): they come first, and your corpse is then ~get all 2.bob~. With no corpse of that name here, ~get all~ and a name takes everything on the floor, as plain ~get all~ does. Someone you cannot see is not counted.
```

  (These lines describe Task 6 Step 7b: `get all <corpse>`, `get all from <corpse>` and
  `get <item> <corpse>` are each tested in `TestGetFromAPlayersCorpse`; plain `get all` stays the
  room-wide take (reads golden D), and so do the other forms (`TestGetAllWithoutACorpseTakesTheFloor`). `death.template` above says
  the same in words; the `config.yaml` comment, Task 6 Step 8, names the same three forms. The
  numbering line is `TestGetAllCountsCorpsesPastTheLiving` and Task 5b's
  `TestCorpseNumbersCountPastTheLiving` (`ann` typed by Ann is her own corpse; `rat` refused while
  Ratty, whose name starts with it, is alive: the Bobby case); the floor case is
  `TestGetAllWithoutACorpseTakesTheFloor`; the hidden line is `TestHiddenBodiesTakeNoCorpseOrdinal`.)

  `_datafiles/world/empty/templates/help/item-names.template` (CRLF in this checkout, LF in git; 16
  lines, the last of them blank): after line 15 (`example: <ansi fg="command">get heavy#2</ansi> would
  get the second match: ...`) and before that final blank line, add (Task 5b's parser):

```text

You can also put the number first, with a dot: <ansi fg="command">get 2.heavy</ansi> is the same.
Both forms work for any name: items, characters and corpses.
```

  so the file still ends with its one blank line.

  `modules/cleanup/files/datafiles/templates/help/bury.template` (CRLF in this checkout, LF in git),
  after line 7 (`  <ansi fg="command">bury rat corpse</ansi>`) and before line 8, the file's final
  blank line (8 lines, the last blank), add the lines below, so the file still ends with its one
  blank line:

```text
  <ansi fg="command">bury 2.rat</ansi>

A name counts the living first, then corpses from the newest: with a live rat
here, <ansi fg="command">bury 2.rat</ansi> buries the newest rat corpse
(<ansi fg="command">bury rat#2</ansi> is the same), and
<ansi fg="command">bury rat</ansi> is refused, since the rat is alive.

A player's corpse that still holds anything cannot be buried, by anyone, its
owner included. A creature's corpse is buried with whatever it holds.
```

  (Task 5b's `TestBuryBuriesTheCorpseItNames` and Task 5c's `TestBuryRefusesAPlayersCorpseThatHoldsAnything`.)

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
  `get all <corpse>` (and `get all from <corpse>`) takes everything from a corpse, and `look <corpse>`
  shows what it holds; every name takes a number either way (`2.rat` or `rat#2`); a name counts the
  living first, then corpses from the newest (with a live rat, `get all 2.rat` empties the newest rat
  corpse, and `get all rat` or `bury rat` says "The rat is still alive."; your own name, typed by
  you, reaches your corpse, for `get` and `look` alike); you can no longer name, target or count a
  hidden character you cannot see (`look`, `attack`, `consider`, `give`, `show`, `follow`,
  `party invite`, `shoot`); seeing hidden (`see-hidden`) now also shows hidden characters by name in
  messages and in the GMCP room list, and lets you fight one; admins always see hidden characters,
  everywhere; following does not cancel hiding: a hidden character can shadow you, and you are not
  told it started or stopped following unless you can see it; a hidden character you cannot see is
  "Someone" as the sender of a say on GMCP (as in the text), is left out of GMCP's enemy list, sends no GMCP room refresh or leave notice, is never
  picked by an untargeted `attack`, and a hidden hostile NPC no longer announces that it notices you;
  a hidden character you cannot see is not named in the "has DIED! (killed by ...)" notice (with no
  killer you can see, the notice names none); a hidden character
  you cannot see leaving the game, starting to quit, appearing by teleport, fleeing, dropping,
  recovering or reviving shows you no line at all;
  nobody can `bury` a player's corpse that still holds anything (its owner included); you
  cannot `suicide` in a fight or in the recovery room; nobody can attack from inside the recovery room's
  copy; a creature with revive-on-death and a player with it both get up at full health; mana regen
  uses the mana amount. Admin: the death, lives and corpse-items settings and the Extra Lives field are
  gone.
- [ ] `status.md`: slice 3 done; Watch: `CreateEphemeralRoomIds` panics when the recovery room is 0
  (Q4; unreachable with the shipped 75). `geomud-verifying`: the smoke flow's step 2 (death now drops
  the ring into the bot's corpse; the bot takes it back after the portal; bot names are
  `Smk<tag>c000`...), and the ok-package count: 41 (`modules/gmcp` gains its first test file in
  Task 5a; `internal/death` has none), unless the run says otherwise: count it.
- [ ] Commit the docs and help pages: `git add` the three help files,
  `_datafiles/world/empty/templates/help/item-names.template`,
  `modules/cleanup/files/datafiles/templates/help/bury.template`, `PATCH_NOTES.md`, `status.md`,
  `gomud_strip_plan.md`, `2026-11-06-phase-5-rules-seam-design.md`, `.claude/skills/geomud-verifying/SKILL.md`;
  message `Phase 5.3 done: help pages, patch notes, status, decision 12, notes for later slices`.
- [ ] Push `strip/phase-5-3-round-death`, open the PR on `pruuk/GeOMUD` (`gh pr create --repo pruuk/GeOMUD`)
  with Task 6's golden summary, merge once CI is green (memory `claude-merges-prs`); move the spec and
  this plan to `docs/completed/`.
