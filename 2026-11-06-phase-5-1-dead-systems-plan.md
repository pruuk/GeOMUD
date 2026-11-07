# Phase 5 slice 1: delete dead rules systems (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Delete the rules systems that no player path reaches (strip plan decision 8), so the seam
in slices 2-6 has less to carry. Quit gets a plain message back. No fight, move or death changes.

**Architecture:** Five deletion groups, run one after another on the branch (they all edit
`internal/characters/character.go`, so no parallel worktrees). Each group deletes, follows the
compiler, proves zero references remain with a grep that can match, keeps the combat goldens
byte-identical (except the one charm fixture), names every reads-golden diff, and commits.

**Tech stack:** Go, testify, the combat and reads goldens in `internal/hooks`, `cmd/smoke`.

Spec: `2026-11-06-phase-5-rules-seam-design.md` (slice 1). Inventory: `2026-11-06-phase-5-inventory.md`
(M section 5, N section 6 item 9, O sections 3-7). Branch `strip/phase-5-1-dead-systems` (one
branch and PR per slice, as in Phase 4b).

---

## Facts verified against source (2026-11-06, `9131eac2`)

Grepped by the controller today. "Non-test" counts are lines from `grep -rn <pattern> --include=*.go . | grep -v _test.go`.

| Fact | Value | Where |
|---|---|---|
| `internal/spells` importers (files) | `main.go` (loader :1507, comment :1453), `characters/character.go` (:181, `GetBaseCastSuccessChance`), `usercommands/help.go` (:168-179, `help spell`), `web/api_v1_yaml.go` (:138), `web/api_v2_spells.go` | grep `internal/spells"` |
| Spell surfaces | routes `api_routes.go:179-185` (v2 spells), `admin_routes.go:56-57`; `admin_items.go:36` (permission map), `:195-200` (handlers); permissions `api_v1_permissions.go:63-65` (`spell*` command keys), `:90` (`spells.write`); pages `_datafiles/html/admin/spells.html`, `spells-api.html` | grep |
| `SpellBook` | field `character.go:68`; admin JSON `api_v1_users.go:99,118,126-127`, `api_v1_mobs.go:110-113`; methods tested in `character_test.go:700-1135` | grep |
| World spells | `_datafiles/world/empty/spells/`: 8 files (aidskill, curepoison, heal, healall, illum, mm, sparks, tameskill) | `ls` |
| Charm, non-test call sites | 45 lines in 18 files: accounts `users.go` 5; behaviour `behaviour.go` 1, `tick.go` 1; combat `combat.go` 1; hooks `NewRound_ActorUpkeep.go` 9, `NewRound_DoCombat.go` 2, `PlayerDespawn_HandleLeave.go` 3; mobs `saved.go` 3; rooms `roomdetails.go` 1, `rooms.go` 7; usercommands `admin.teleport.go` 2, `attack.go` 1, `go.go` 3, `party.go` 1, `shoot.go` 1, `suicide_npc.go` 2, `usercommands.go` 1; gmcp `gmcp.Party.go` 1 (plus inside `characters`) | grep `IsCharmed\|GetCharmIds\|TrackCharmed\|RemoveCharm\|GetCharmedBy\|\.Charmed\b\|CharmedMobs\|CharmInfo\|CharmPermanent\|CharmExpired\|FindCharmed\|\.Charm\(` |
| Charm tests | `accounts/charmed_test.go`, `behaviour/behaviour_test.go`, `characters/character_test.go`, `characters/charminfo_test.go`, `hooks/charmexpiry_test.go`, `hooks/combat_golden_test.go` (`TestCombatGoldenMvMCharmed` :438), `mobs/saved_test.go` | grep |
| Professions | `skills/profession.go`, loader `main.go:1514`, reload `admin.reload.go:44`; `GetProfession` in `accounts/userrecord.go:609` (online info), `templates/templatesfunctions.go:150`, `gmcp.Char.go:430` (`Char.Info.Class`), `GetProfessionRanks` `gmcp.Char.go:657` (`Char.Jobs`); `OnlineInfo.Profession` `accounts/onlineinfo.go:8`, `online.go:21,68` (column), `gmcp.Game.go:65,79`; web `api_v1_professions.go`, `api_v1_yaml.go:177-204`, routes `api_routes.go:158-163`, `admin_routes.go:38`, `admin_items.go:167`, `admin_nav.go:95`, page `skills-professions.html`; world `professions/` 10 files | grep |
| Flags to delete, readers | `poison` `character.go:642`; `drunk` `say.go:22`, `shout.go:21`; `tripping` `roomdetails.go:453-458` (also a colour pattern of that name); `accuracy` `combat/calculations.go:204`; `blink` `calculations.go:205` (`emote.go:18` "blink" is an emote word, keep); `see-nouns` `roomdetails.go:127`, `look.go:338`. YAML in `_datafiles/world/empty/buffs-flags/` | grep each flag in quotes or backticks |
| Other dead items, non-test lines | `LootGoblin` 9; `PricePerLife` 3 (+1 test file); `ChanceToTame` 2 (+1 test); `GetRoomWithMostItems` 1; `RememberRoom` 2; `roomHistory` 6; `TrapBuffIds` 8; `tracking-mob` 1; `AddWornBuff` 1 (+1 test); `shopstyle` 7; `MobMastery` 6; `FormRaceId` 10; `TrackBuffStarted` 1; `SearchBuffs` 1 (+1 test); `GetBuffIdsWithFlag` 2 | grep counts |
| Missing buffs | `mutators/wildfire.yaml:12-13` (`playerbuffids: [22]`, `mobbuffids: [22]`); `items/consumables-30000/30001-small_red_potion.yaml:8-9` (`buffids: [5]`); shipped buffs are 0, 1, 24 | grep, `ls buffs` |
| `ShopItem.BuffId` | `characters/shop.go:21`, `:158`; `web/api_v1_mobs.go:179` | grep |
| Quit today | `quit.go:9-23`: refuses in a fight, else `user.AddBuff(0, "quitting")`; silent | read |
| Old quit text (deleted JS, `git show 57255ee4^:_datafiles/world/empty/buffs/0-meditating.js`) | onStart: "You sit down and begin your meditation." / "Your meditation must complete without interruption to quit gracefully." / room: "<name> sits down a begins to meditate." onTrigger: "You continue your meditation. *N rounds left*" | git show |
| Send shape | `messaging.Send(perception.Message{Kind: perception.WorldEvent, Actor: actor.Id, Senses: perception.Sight \| perception.Sound, Range: perception.Room, ActorText: ..., ObserverText: "{actor} ..."})` | `usercommands/emote.go:101-108` |
| Goldens touched | `G-character-info.txt:39` (shopstyle line), `:112` (online table Profession column); `H-outside.txt` (tame lines); combat `mvm_charmed.txt` | grep |
| Golden update flag | `-update` rewrites combat and reads goldens (`reads_golden_test.go:76`); on Windows it rewrites every golden's line endings: restore untouched ones before committing | test source; handoff note |
| Test baseline (Windows) | 39 packages ok, 3 known failures (`geomud-verifying`) | skill |

## Decisions (plan defaults; none is player-visible beyond what the rulings already decided)

1. **Charm goes entirely**, including companions in party lists, follow-on-move, flee-together,
   charmer kill credit, the `charmed` adjective, saved `charmedmobs` and NPC `Charmed`, and login
   pruning. YAML decode of old saves with a `charmedmobs:` key: verify in Task 3 whether the loader
   is strict (`yaml.UnmarshalStrict`); if strict, keep the field name in a `yaml:"charmedmobs,omitempty"`
   ignored stub or add a migration, whichever the loader allows with less code.
2. **`mvm_charmed` combat fixture** is rewritten as `mvm` (two NPCs fighting, aggro set directly, no
   charm) and re-recorded. It is the only combat golden that may change. Its diff is summarised in
   the commit.
3. **Professions go**; skills stay until slice 4 (the `dual-wield`, `peep`, `enchant` checks are
   live). The `online` table loses its Profession column; GMCP `Char.Info.Class` and `Char.Jobs` go
   (fields removed, not sent empty). The `profession` template function goes.
4. **Tame mastery goes** (`MobMastery`, the `Rand(1000)` roll in `suicide_npc.go`, `ChanceToTame`
   and helpers): nothing tames. Removing the roll shifts later random draws in reads golden H; the
   diff is named in the commit.
5. **Dead `Character`, `mobs`, `races`, `stats` methods** from inventory M section 5 go, each
   proved by a grep that also matches a live method in the same file.
6. **Cooldowns stay.** They are a generic engine framework, not a rules system, and not in decision 8.
7. **Quit message:** the actor reads "You sit down and begin your meditation." then "Your
   meditation must complete without interruption to quit gracefully."; the room reads "{actor} sits
   down and begins to meditate." (upstream's typo "a begins" fixed). No per-round countdown.
8. **Flag readers are deleted with the flag**, not left reading a flag nothing grants. The `tripping`
   colour pattern goes if nothing else names it.
9. **`HasFlag` nil check, `reload buffs`, `Source` storage** wait for slice 6 (design).

## Agent rules (brief every subagent with these)

- Edit files with the Edit and Write tools. Never write a repo file from the shell (no `>`, `>>`,
  `sed -i`, Python open-write); a hook blocks it. Deleting whole files with `git rm <path>` is fine.
- Read files in chunks of at most 400 lines; state "read N of N files" in the report.
- `git add <named paths>` only, never `-A` or `.`; commit messages through a quoted heredoc
  ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never `--no-verify`.
- Never push. Never touch `upstream`.
- A grep that proves absence must also show it can match (run the same pattern on a symbol that
  exists).
- `go test` always with `-count=1`. Never run `-update` on all goldens; to re-record one, run
  `go test ./internal/hooks -run <TestName> -count=1 -update`, then `git diff --stat` and
  `git checkout` any golden file you did not mean to change (Windows rewrites line endings).

## Per-group check (end of every task)

```bash
go build ./...
go vet ./internal/... 2>&1 | head
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'   # only the 3 known Windows failures
gofmt -l $(git diff --name-only -- '*.go')                    # prints nothing
git diff --stat -- internal/hooks/testdata/                    # only the files this task names
```

Report: build result, the failing-test lines, every changed golden with a one-line reason.

---

### Task 0: Baseline and docs commit (controller)

- [ ] Run the per-group check on the clean branch; record the output in the scratchpad.
- [ ] Commit the Phase 5 docs: `gomud_strip_plan.md`, `status.md`, `2026-11-06-phase-5-inventory.md`,
  `2026-11-06-phase-5-rulings.md`, `2026-11-06-phase-5-rules-seam-design.md`, this plan.
  Message: `Phase 5: inventories, rulings (decisions 5-11), design and slice 1 plan`.

### Task 1: Spells (Sonnet)

**Files:** delete `internal/spells/` (all files), `internal/web/api_v2_spells.go`,
`_datafiles/html/admin/spells.html`, `spells-api.html`, `_datafiles/world/empty/spells/` (8 files).
Modify `main.go:1453,1507`; `internal/usercommands/help.go:168-179`; `internal/web/api_v1_yaml.go`
(spell YAML handler near :138 and its route); `api_routes.go:179-185`; `admin_routes.go:56-57`;
`admin_items.go:36,195-200`; `api_v1_permissions.go:63-65,90`; admin nav entry for spells (grep
`spells` in `admin_nav.go`); `internal/characters/character.go` (`SpellBook` field :68,
`GetBaseCastSuccessChance` :179, spellbook methods from `GetSpells` :381 through `UnLearnSpell` :430, `SpellCast` aggro type and
`SpellAggroInfo` in `aggro.go:11-19` if nothing else reads them); `api_v1_users.go:99,118,126-127`;
`api_v1_mobs.go:110-113`; `character_test.go` spellbook tests.

- [ ] Delete and follow the compiler until `go build ./...` passes.
- [ ] Prove absence: `grep -rnE 'internal/spells"|SpellBook|SpellCast|spells\.write|/admin/spells|LoadSpellFiles' --include=*.go --include=*.html --include=*.js .` prints nothing (control: the same command with `LoadDataFiles` prints hits).
- [ ] Check the help text: `grep -rn -i "help spell\|spells" _datafiles/world/empty/templates/help/*.md _datafiles/world/empty/keywords.yaml`; remove lines that point players at spells or `cast` (stat help templates that mention spells stay until slice 4; list them in the report).
- [ ] Per-group check. Expected golden change: none.
- [ ] Commit: `Phase 5.1: delete spells (package, data, help spell, admin pages and API)`.

### Task 2: Charm and companions (Sonnet)

**Files:** delete `internal/characters/charminfo.go`, `charminfo_test.go`,
`internal/accounts/charmed_test.go`, `internal/hooks/charmexpiry_test.go`. Modify every file in the
charm row of the facts table, plus `characters/character.go` (`Charmed`, `CharmedMobs` :69-70,
methods :465-526, `GetMaxCharmedCreatures` :974), `characters/formattedname.go` (`charmed`
adjective), `behaviour/behaviour_test.go`, `mobs/saved_test.go`, `character_test.go`.

Call-site rules:
- Party and GMCP party lists: list only members (drop the companion rows).
- `go.go` follow-on-move, `NewRound_DoCombat.go` flee-together, `admin.teleport.go` party
  teleport: delete the companion branch.
- `combat/combat.go` credit to the charmer: credit goes to the attacker's account only (an NPC
  attacker credits nobody, as an uncharmed NPC does today).
- `attack.go`/`shoot.go` "your friend" refusal: delete.
- `behaviour` wander suppression for charmed NPCs: delete.
- `NewRound_ActorUpkeep.go` charm countdown and expiry: delete.
- `accounts/users.go` login pruning and the newbie-guide special case: delete.
- `mobs/saved.go` `Charmed`: delete the field; check the decoder (decision 1).
- Combat golden: rename `TestCombatGoldenMvMCharmed` to `TestCombatGoldenMvM`, drop the `Charm`
  call, keep the seed (404) and fixture, set the two NPCs' aggro directly; delete
  `testdata/combat-golden/mvm_charmed.txt` and record `mvm.txt` with `-update` on that test only.

- [ ] Delete and follow the compiler.
- [ ] Prove absence: `grep -rnE 'Charm|charmed' --include=*.go .` prints only unrelated hits; list each one in the report with why it is unrelated (control: the pattern hits before the change).
- [ ] Decoder check (decision 1): write a test in `internal/mobs` or `internal/accounts` that decodes a YAML fixture containing `charmedmobs: [5]` (users) and `charmed:` (NPC sidecar) and asserts no error. It must pass.
- [ ] Per-group check. Expected golden changes: `mvm_charmed.txt` deleted, `mvm.txt` added; reads goldens unchanged (if a party or look line changes, name it).
- [ ] Commit: `Phase 5.1: delete charm and companions`.

### Task 3: Professions, tame mastery and dead character methods (Sonnet)

**Files:** delete `internal/skills/profession.go`, `internal/web/api_v1_professions.go`,
`_datafiles/html/admin/skills-professions.html`, `_datafiles/world/empty/professions/` (10 files),
`_datafiles/world/empty/templates/character/skillup.template` (no Go reference; verify with grep
`skillup` first). Modify per the professions row of the facts table; `skills.go` (`GetProfession`,
`GetProfessionRanks`, `ProfessionRank`, `GetExperienceLevel` if unused after); `admin.go` profession
CRUD; `templatesfunctions.go:147-151` (`profession`) and `:111` (`tnl` only if it has no template
user: grep `tnl` over `_datafiles`).

Tame: `characters/mobmastery.go` (delete), `MobMastery` field and readers, `suicide_npc.go:174-204,
271-302` (mastery roll), `combat/calculations.go` `ChanceToTame` :365 with `chanceToTame`,
`tameSizeModifier`, `tameHealthBonus` and their tests; `mobs.IsTameable` and race `Tameable` only if
nothing else reads them after.

Dead methods (inventory M section 5; delete each only after its grep shows no non-test caller):
`SetSkill`, `TrainSkill`, `GetSkillLevelCost`, `GetMapSprawlCapacity`, `GetMemoryCapacity`,
`RememberRoom` with `roomHistory` and its call at `rooms/roommanager.go:417`, `KnowsFirstAid`
(keep the race field until slice 4; the template shows it), form change (`FormRaceId`,
`ApplyFormChange`, `RevertFormChange`, `IsFormChanged`, `RaceSize`, the form branch in
`GetDescription` and `GetAdjectives`), `ApplyManaChange`, `BarterPrice`, `MovementCost`,
`races.FindRace`, `stats.SetMod`, `mobs.HatesRace`, `HatesMob`, `ConsidersAnAlly`, `GetSellPrice`,
`Despawns`.

- [ ] Delete and follow the compiler.
- [ ] Prove absence per symbol: `grep -rnw <Symbol> --include=*.go .` empty for each (control: `grep -rnw GetSkillLevel` hits).
- [ ] Per-group check. Expected golden changes: `G-character-info.txt:112` (online table loses the Profession column); `H-outside.txt` (tame lines go; later random draws may shift: describe each changed line).
- [ ] Commit: `Phase 5.1: delete professions, tame mastery and dead character methods`.

### Task 4: Dead flags, buff appliers and missing buff references (Sonnet)

**Files:** delete `_datafiles/world/empty/buffs-flags/{poison,drunk,tripping,accuracy,blink,see-nouns}.yaml`.
Modify the readers in the flags row (delete the branch, keep the surrounding code); `combat/calculations.go:200-235`
(crit chance without the two flag inputs; the random draws must not change); the `tripping` colour
pattern in `_datafiles/world/empty/color-patterns.yaml` if grep shows no other user;
`mutators/wildfire.yaml:12-13` (drop the two buff lines); `30001-small_red_potion.yaml:8-9` (drop
`buffids`); `gamelock.Lock.TrapBuffIds` and its 8 sites; `characters/shop.go:21,158` and
`web/api_v1_mobs.go:179` (`ShopItem.BuffId`); the `tracking-mob`/`tracking-user` block with
`RemoveBuff(26)` in `rooms/roomdetails.go:366-440`; `items.AddWornBuff`; buffs dead code
(`TrackBuffStarted`, `OnStartWaiting`, `Started`, `BuffMessage`/`BuffMessages`, the unused
`buffId...` parameter of `Trigger`, `TriggersLeft` method, `GetBuffIdsWithFlag`, `SearchBuffs`)
with their tests.

- [ ] Delete and follow the compiler.
- [ ] Prove absence: each flag name in quotes or backticks over Go and YAML (control: `"hidden"` hits); each symbol with `grep -rnw`.
- [ ] Boot check: `go run ./cmd/smoke` prints no "unknown buff flag" or missing-buff warning in the server log (`-server-log <scratchpad file>`; grep it).
- [ ] Per-group check. Expected golden changes: none (no shipped buff carries these flags).
- [ ] Commit: `Phase 5.1: delete unused buff flags, dead buff appliers and missing buff references`.

### Task 5: Config leftovers, set shopstyle, quit message (Sonnet)

**Files:** `internal/configs` (`LootGoblin` block and `PricePerLife` with validation and test),
`_datafiles/config.yaml` (the `LootGoblin` block at :112 and `PricePerLife`), `usercommands/admin.mob.go`
(`mob loot goblin` branch, :129-132), `rooms/roommanager.go` (`GetRoomWithMostItems` :441 and any
loot-goblin helper), `usercommands/set.go:82-87,297-330` (shopstyle), the account option it writes,
`_datafiles/world/empty/templates/help/set.md:35-36`, `usercommands/quit.go`.

Quit (decision 7): after the fight check and before `AddBuff`:

```go
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        actor.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ActorText:    "You sit down and begin your meditation.\nYour meditation must complete without interruption to quit gracefully.",
		ObserverText: "{actor} sits down and begins to meditate.",
	})
```

Check first how other multi-line `ActorText` values are written (grep `ActorText:.*\\n`); if none
use `\n`, send the second line with `Tell(actor, ...)` instead. Add a test in
`internal/usercommands` beside the existing command tests (grep `func TestQuit` first; if none, model
it on the nearest `messaging` capture test there) asserting the actor and an observer read their
lines and that a fighter gets "You're too busy to quit right now!" and no meditation line.

- [ ] Write the quit test; run it; it fails (no meditation line).
- [ ] Add the send; run it; it passes.
- [ ] Delete the config and shopstyle items; follow the compiler.
- [ ] Prove absence: `grep -rnwE 'LootGoblin|PricePerLife|GetRoomWithMostItems|shopstyle|ShopStyle' --include=*.go --include=*.yaml --include=*.md --include=*.template .` (control: `grep -rnw AlignmentDecayRounds` hits).
- [ ] Per-group check. Expected golden changes: `G-character-info.txt:39` (shopstyle line gone; check the `set` overview lines around it). If a reads golden covers quit, name its new lines.
- [ ] Commit: `Phase 5.1: delete loot goblin, PricePerLife and set shopstyle; quit says it is meditating`.

### Carried from task reviews

- Task 1 review: `ansi-aliases.yaml:12-13` (`spellname`, `spell-text`) unused: delete in Task 4.
  `statmods.Casting`/`CastingPrefix` (`statmods.go:20-21`) dead since spells went: slice 4 deletes
  them with statmods (add to the design's slice 4 notes in Task 6). Deleting `SpellCast` shifted
  `AggroType` values (`Flee` 4 to 3); NPC sidecars save `Type` as an int (`mobs/saved.go:20`), so a
  local dev sidecar saved mid-flee loads as a plain attack. Unlaunched, no data in the repo: no fix.
  No loader in the repo decodes YAML strictly, so removed keys in old saves load fine.
- Task 2 (`93e400f0`): a second combat golden changed. `flee_success` charmed its `dog` bystander,
  so `flee_success.txt` loses one line (`Input dog: "flee"`); the review confirmed no other line or
  random draw moved. `mvm.txt` differs from `mvm_charmed.txt` only in the rat's `playerdamage`
  (charmer credit gone) and three group-hatred lines. Follow-up commit removes code the charm
  deletion left dead (`behaviour.Rest`, the accounts function exporter); `Room.RestNPC` moved to
  `rooms/export_test.go` because the resting-store tests seed state with it.
- Task 3 (`bb96c182`): reads goldens `G-character-info.txt` (online table loses Profession) and
  `H-outside.txt` (tame line gone; title string; `RoomRoundTick x21` to `x22` because removing the
  `Rand(1000)` tame roll shifts the seeded stream the idle-message loop draws from). Review
  confirmed both. Follow-up removes `GetAllSkillRanks`, the form split in `RecalculateStats`,
  `statmods.Tame`, the `formset` permission, the `tame` skill file, `Mob.Hates` (no reader since
  `HatesRace`/`HatesMob` went), and unused CSS and JS leftovers.
- Task 4 (`8cd1bdac`): `critChance` lost the accuracy/blink inputs and a second min clamp; review
  confirmed both clamps compared against the same `minChance` with only the flag branches between,
  so no value changes; `Crits` draws the same. Follow-up: `users.html` buff shop entries (missed),
  dead `RoomView.TrackingString`, the room visitor exit finders and `ApplyCharacterWarping`, the
  `poisoned` adjective style (`5231e9c5`). Left for later: `Room.userVisitors`/`mobVisitors` are
  still written but read only by a `len(...) > 0` check (`rooms.go` ~:1352); clean up with the
  room code when that check's purpose is settled.

- Whole-branch review (Opus): Task 5 passed; follow-ups `bcc1582c` (room visitor tracking removed;
  `MarkVisited` keeps the `lastVisited` bump the unloader reads; `TurnsPerSecond` gone) and
  `ffb7eb6d` (five functions orphaned by the slice, GMCP party help, known issues). The controller
  found the rewritten `move_test` asserts vacuous (`AddTestRoom` already sets `lastVisited`), zeroed
  it before the move, and showed the asserts fail with `MarkVisited` stubbed.

### Task 6: Verify, review, PR (controller)

- [ ] Full check from a clean tree: `go build ./...`, `go test ./... -count=1` (3 known failures only), `gofmt -l .` minus vendored, smoke `go run ./cmd/smoke` and `-bots 5`; paste the summary lines.
- [ ] `git diff master --stat -- internal/hooks/testdata/` lists only: `mvm_charmed.txt` (deleted), `mvm.txt` (added), `G-character-info.txt`, `H-outside.txt`, and any quit golden named in Task 5.
- [ ] Opus whole-branch review of slice 1 (correctness and missed references; brief with the agent rules and this plan). Fix findings in their own commits.
- [ ] `PATCH_NOTES.md` line, `status.md` update, design "Notes for later slices" if a finding belongs to a later slice; the `geomud-verifying` package count if a package with tests was deleted (`internal/spells` has no tests: check).
- [ ] Push `strip/phase-5-1-dead-systems`, open the PR on `pruuk/GeOMUD` (`gh pr create --repo pruuk/GeOMUD`), merge once CI is green (memory `claude-merges-prs`).
