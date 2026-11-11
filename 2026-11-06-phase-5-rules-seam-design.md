# Phase 5: rules seam, design

Date: 2026-11-06. Written at `master` `9131eac2`; one branch and PR per slice
(`strip/phase-5-<n>-<name>`). Inputs:
`gomud_strip_plan.md` (Phase 5 row; decisions 1-9), `2026-11-06-phase-5-inventory.md` (M, N, O),
`2026-11-06-phase-5-rulings.md`, `2026-10-27-phase-4b-npcs-are-players-design.md` ("Notes for later
slices").

**Goal.** A `Ruleset` interface the engine calls for everything that is a game rule, with a bare
reference ruleset behind it, so an Earthdawn ruleset can be layered on later with no core edit, and
a fork of the base is still a working MUD.

## Rulings this design implements (Calabe, 2026-11-06; strip plan decisions 5-9)

- The reference ruleset ends bare: health, a hit roll with weapon damage, health regen, death at 0
  with a respawn, races as names only, and gold. Levels, XP, the stat block, mana and racial stat
  differences go.
- The rules move behind the seam first with parity (goldens unchanged), then shrink (goldens change
  on purpose), per decision 1.
- The engine knows health and nothing else of the rules.
- One death path for every actor; the ruleset says whether an actor is dead, so a later ruleset can
  add a downed state.
- Dead rules systems are deleted (spells, cast, charm, dead skill code, professions, `set
  shopstyle`, unused buff flags, references to missing buffs). Quit gets a plain message.
- Earlier rulings that still bind: no JS or Lua (behaviour is Go); one actor model, no player-only or
  NPC-only path for what both do; no escaping a fight except by fleeing; every swing earns credit; a
  hidden actor reads "Someone" to anyone who can't see it.

## Facts verified against source (2026-11-06, `9131eac2`)

Grepped or printed by the controller today, not copied from the inventories.

| Fact | Value | Where |
|---|---|---|
| Rules packages, non-test lines | combat 1,843; buffs 1,020; skills 530; races 264; spells 251; statmods 101; stats 87 | `cat internal/<pkg>/*.go` minus tests, `wc -l` |
| Importers (non-test packages) | stats 5, statmods 5, combat 3, races 7, skills 7, spells 4, buffs 11, characters 12 | `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...` |
| `Character` | `character.go` 2,128 lines; `GrantXP` :438, `SetAggro` :1104, `ApplyHealthChange` :1219, `XPTL` :1266, `LevelUp` :1295, `Heal` :1331, `RecalculateStats` :1383, `Validate` :1515 | grep `func (c \*Character)` |
| `Heal` bug | `c.Mana += hp` (mana argument unused) | `character.go:1339` |
| Unknown race | becomes race 1 | `character.go:1521-1523` |
| Combat round | `NewRound_DoCombat.go` 521 lines; `fightersWithAccountsFirst` :49; `combat.Attack(a, d)` :182; group hatred names `Stats.Perception` :217; `handleAffected` :477 | grep |
| Bleed-out | `Health < 1` bleeds 1 a round, "Somebody needs to provide aid!", dies at `<= -10` | `NewRound_AutoHeal.go:44-59` |
| Death | all deaths run `suicide`; `suicide.go:28-30` branches to `npcDeath` (`suicide_npc.go`, 443 lines) | read |
| Hard-coded zones in death | `Shadow Realm` (`suicide.go:41`), `Training` (`suicide_npc.go:119,242`) | grep |
| `revive-on-death` | read only on the NPC path | `suicide_npc.go:31,43` |
| NPC aggression | a hostile NPC prints "notices you as you enter!" and nothing more | `usercommands/go.go:322` |
| PvP gate | `Room.CanPvp` :2327 (min level :2343); `attack` calls it (`attack.go:239`), `shoot` does not | grep `CanPvp` |
| Action points | spent only by `go`: 10, or 50 when items exceed `CarryCapacity()` | `go.go:51-58`; `DeductActionPoints` has one caller |
| Prompt | default `HP:{hp}/{HP} MP:{mp}/{MP}` | `config.textformats.go:18`, `_datafiles/config.yaml:453` |
| GMCP stats | fixed fields such as `strength` | `modules/gmcp/gmcp.Char.go:880` |
| Combat goldens | 9 files in `internal/hooks/testdata/combat-golden/` (pvm_death, pvp, mvp, mvm_charmed, flee_blocked, flee_success, shoot_exit, observer, hidden_attacker) | `ls` |
| Guards | `internal/guards`: `combat_test.go`, `commands_test.go`, `output_test.go`, `tick_test.go`, `guards_test.go` | `ls` |
| World rules folders | `buffs` 3, `buffs-flags` 17, `spells` 8, `skills` 16, `professions` 10, `races` 3 (inventory O section 6; folder list checked with `ls _datafiles/world/empty`) | `ls` |

## Architecture

### Packages

- `internal/rules`: the seam. The `Ruleset` interface, the value types it returns, and the active
  ruleset (`rules.Get()`, set once at boot; `rules.SetForTest` returns a restore func for
  `t.Cleanup`). Engine packages import this and never a ruleset.
- `internal/rules/dikubasic`: the reference ruleset. `internal/combat` moves here first, verbatim.
  `stats`, `statmods`, `races` data and `skills` follow it in, then most of them are deleted in the
  shrink.
- Root `main.go` is the one place that names `dikubasic` (`rules.Set(dikubasic.New())`).

### What the engine keeps

The round loop and listener order; who fights whom (aggro, targeting, range, flee movement, the
hidden-target skip); messaging; kill credit; `KilledBy`; wimpy; corpses and item and gold moves;
the recovery-room move; timed affects (apply, tick, expire, persist, cancel by flag); health and max
health; the one death routine; events.

### What the ruleset answers (first cut; each slice plan fixes the exact signatures)

Grounded in the call sites in inventories M, N and O. The ruleset never sends text: it returns
messages for the engine to send (the principle of `TestCombatPackageDoesNotSend`).

- Fight: `TurnOrder`, `AttackDelay`, `ResolveAttack` (today's `calculateCombat`, moved verbatim so
  the random draw order holds), `HatredRounds`, `BreaksOnHit`, `FleeEscapes`, `Consider`,
  `CanAttack` (one check for `attack` and `shoot`).
- Condition: `Condition(actor) {Up, Downed, Dead}` replacing the five scattered health thresholds;
  `ClampDamage` for floors.
- Upkeep: `RoundUpkeep(actor)`, `Regen(actor)`.
- Death: `OnDeath(victim, credited, killer) DeathOutcome` (revive or not, revive health, drops,
  awards, lines, whether to go to the recovery room).
- Bodies: `NewBody` (starting values), `Refresh` (derived values; today `RecalculateStats`).
- Display: `Summary(actor)` for `who`, party and GMCP; `PromptToken(actor, token)` for prompt
  tokens other than health; `Kinds()` for the start flow and `help races`.
- Affects: `AffectEffects()` (Go `OnApply`/`OnTick`/`OnEnd` per effect key), `AffectFlags()` (rules
  flags), `TraitKeys()` (validates affect and item mods).
- Senses: `CanPerceive` (today's `messaging.CanPerceive` body as the default).

### Per-body ruleset values

The engine gives each `Character` a keyed store the ruleset owns (proposal: `Values map[string]int`
saved under one YAML key, plus a ruleset-owned blob if a ruleset needs structure). Gold lives there
under a key the ruleset names (`rules.Get().Currency()` gives the key and the display word), so the
kept get, drop, give and put commands move "the currency" without naming gold. A load-time
migration moves today's `gold:` into the store (`accounts/migration.go:25-92` already runs user
migrations at boot). Health and max health stay engine fields.

### Guard test

`TestCoreNamesNoRulesetValue` in `internal/guards`, built on the Go syntax tree like the existing
guards (inventory O section 8):
1. Import ban: no core file imports `internal/rules/dikubasic`.
2. Selector and exact-string-literal bans on every value the reference ruleset registers (its
   currency key, trait keys, flag names), read from `dikubasic` by the test so the list cannot
   drift. Import resolution keeps `perception.Perception` from tripping it.
3. In-memory positive controls, and stale-allowlist failure, as `output_test.go` does.

The existing `combat_test.go` controls name `internal/combat/combat.go`; they move in the same
commit as the code.

## Slices

Each slice gets its own plan with a verified facts table, parity or golden evidence, an Opus review,
and a PR. Calabe sees only player-visible rulings.

1. **Delete dead systems** (decision 8). Spells package, world folder, `help spell` path, admin
   pages and API; charm plumbing (start, countdown, follow, party listing, persistence; nothing can
   start one); spellbook, form change, cast chance, barter, tame, `map` skill and the other dead
   `Character` and `mobs` methods M lists; professions folder and loader; `set shopstyle`; unused
   flags (`poison`, `drunk`, `tripping`, `accuracy`, `blink`, `see-nouns`); `wildfire`'s buff 22 and
   the potion's buff 5; loot goblin config and its admin branch; `PricePerLife`; `ChanceToTame`;
   `GetRoomWithMostItems`; `RememberRoom` and `roomHistory`; `TrapBuffIds`; `ShopItem.BuffId`; the
   tracking block with `RemoveBuff(26)`. Quit gets a plain message back. Combat goldens unchanged
   (charm fixture `mvm_charmed` is rewritten or dropped, named in the plan); reads goldens change
   only where quit, help or `set` text changes.
2. **The seam and the fight** (decisions 1, 5). `internal/rules` and `dikubasic`; `internal/combat`
   moves verbatim; `DoCombat`, `consider`, `attack`, `shoot` call the ruleset. All 9 combat goldens
   byte-identical with no `-update`. A stub-ruleset test changes the fight with no core edit. In its
   own commit: `shoot` goes through `CanAttack` (PvP settings apply to it) and fires `AggroChanged`.
3. **Round, regen and death through the seam** (decisions 5, 7). `Condition`, `Regen`,
   `RoundUpkeep`, one engine death routine calling `OnDeath` for every actor (replaces the
   `suicide.go`/`suicide_npc.go` branch; `revive-on-death` then works for anyone). Before moving
   death, add a combat golden that runs the queued `suicide`. Parity: goldens unchanged. The
   hard-coded `Shadow Realm` and `Training` names go (the first becomes "already in the recovery
   room"; `Training` exists nowhere in `world/empty`, checked in the plan).
4. **Shrink the reference ruleset** (decisions 5, 7, 9). Delete levels, XP, the stat block, mana,
   alignment, skills, permadeath and extra lives, racial stats and the race
   fields beyond the name; die at 0 (no downed state,
   no bleed); simple hit roll and weapon damage; regen for every actor (one-actor ruling). World
   data, help templates, config knobs, admin pages, prompt tokens and GMCP fields that served them go
   with them. Goldens are re-recorded and each diff is read and summarised in the PR.
5. **The engine knows health only** (decision 6). The per-body value store; gold moves into it with
   a save migration; prompt, `who`, party and GMCP get rules values from the ruleset; the guard test
   lands.
6. **Timed affects** (strip plan Phase 5 row). `buffs` becomes the engine's timed affects: one
   upkeep pass for every actor per round (replaces `PruneBuffs`' player and NPC split); engine flags
   become Go constants; buff YAML names a Go effect by `effect:` key; mutators get one buff list for
   every actor plus the native list; the source is stored so GMCP stops reading "unknown"; nil check
   in `HasFlag`; `reload buffs`.
   Rename `buffs` to `affects` as a separate mechanical commit at the end.
7. **Close Phase 5.** Smoke runs a fight through the ruleset; a stub ruleset swapped in at boot
   changes the outcome with no core edit; docs sweep (`help` pages, known issues, the spec notes
   this phase settles).

Slices 1-3 suit the 4b process: Opus plan, parallel Sonnet groups in agent worktrees, Opus review.
Slice 4 needs Opus throughout (every golden changes; each change must be read).

## Consequences of the rulings (answered 2026-11-06)

These followed from decisions 6 and 9 but change what a player sees, so they were asked. Calabe
took the recommendation on both: (1) is decision 10, (2) is decision 11 in `gomud_strip_plan.md`.
Slice 4 implements both.

1. **Movement cost and encumbrance.** Action points are spent only by `go` (10 a move, 50 when
   carrying more than `5 + Strength/3` items). With no stats and the engine knowing health only,
   action points become a ruleset value. Options: (a) the bare ruleset drops them: moving is free
   and carrying has no limit; (b) it keeps a flat pool and a fixed carry limit (no stats), so
   encumbrance still slows a player. Recommendation: (a). It is the bare floor, and Earthdawn brings
   its own movement and carrying rules.
2. **What a death costs.** Today a player above level 5 may drop equipment (25% each), gold, and
   leave a corpse; XP penalty is off; permadeath is off. With no levels, the protection band and XP
   penalty go regardless. Options: (a) nothing is lost: respawn in the recovery room with
   everything; (b) carried items and gold drop into a corpse, worn gear stays; (c) keep today's
   chances (25% per worn item, gold, corpse). NPC deaths drop their loot in every option.
   Recommendation: (b). A fork gets a working corpse run with no tuning knobs, and the corpse code is
   engine code that stays anyway.

## Notes for later slices

Each slice plan reads these first; a note leaves once a slice settles it.

- **From slice 1 (2026-11-06).** Slice 4 deletes `statmods.Casting`/`CastingPrefix` (dead since
  spells went), the race field `KnowsFirstAid` (shown only in `races.template`), and the stat help
  templates that still mention spells (`help/mysticism`, `smarts`, `stats`); `skills/cast.yaml` goes
  with skills. Upstream `AGENTS.md`/`README.md` notes in `characters`, `gamelock`, `rooms`, `mobs`,
  `buffs`, `skills` still describe charm, spells, trap buffs and visitor tracking as live; they go
  with their packages or are trimmed in slice 7's docs sweep. `AggroType` values shifted when
  `SpellCast` went (`Flee` 4 to 3); NPC sidecars save the int, so a pre-slice-1 local sidecar saved
  mid-flee loads as a plain attack (no shipped data). The spec's 4b note "Slice 5 inherits" is stale
  (inventory N section 6): only the -10 branch, NPCs never attacking, accounts-only alignment decay
  and the killer quirks remain, and slices 3-4 settle them (players-first turn order is now
  `dikubasic.TurnOrder`, slice 2; the pet hatred quirk went with charm in slice 1).
- **From slice 2 (2026-11-10; spec `docs/completed/2026-11-10-phase-5-2-seam-fight-spec.md`).**
  - The weapon delay (`characters.SetAggro`, weapon plus offhand `WaitRounds`) is still engine code:
    `characters` cannot import `rules` (cycle via `actors`). It moves behind the ruleset with swing
    speed (Calabe 2026-11-10: weapon speed plus stats and skill set swings per round, possibly 0
    with carried progress; out of scope until then).
  - Two files still import `dikubasic`, allowlisted in `TestOnlyBootNamesTheRuleset`:
    `web/api_v1_items.go` and `web/api_v1_mobs.go` (rankings, slice 4). Remove each from the
    allowlist when it goes.
  - Slice 4 deletes offhand breakage (Calabe 2026-11-10): `offhandBreakCheck` in `DoCombat`, the
    broken item 20, and `BreakChance`/`BreakTest` if nothing else reads them; the `pvp` golden
    (shield break chance 90) changes then.
  - Slice 4 deletes the `BackStab` branch in `calculateCombat` (nothing sets `BackStab`; its reset
    acts on a copy of the character).
  - Better targeting and aggro tracking than the single aggro flag is wanted later (Calabe
    2026-11-10), outside the slim base.
  - Test config helpers: `setConfig` (hooks), `setPvP` (usercommands) and an inline copy in
    `dikubasic/ruleset_test.go` repeat one pattern, and each leaves restored keys pinned as explicit
    overrides (a later `AddOverlayOverrides` on them in the same test binary is skipped). Nothing
    trips it today; one `configs` test helper that removes the keys it added would replace all three.
- **From slice 3 (2026-11-10).** Slice 4 takes what slice 3 left in the engine:
  - NPC regen (one-actor rule).
  - The -5/-10 floors in `ApplyHealthChange`.
  - XP, levels and alignment, alignment decay included (still in upkeep).
  - The four `characters` health checks and `rooms.FindDowned`.
  - Offhand breakage and `BackStab` (see slice 2 above).
  - `actors.BodyReplaced`, which has no production caller since permadeath went.

## Out of scope

NPC aggression, initiative, `aid`, NPC levelling, party XP rules, possession, the geo work. Each
waits for a ruleset or a later phase.
