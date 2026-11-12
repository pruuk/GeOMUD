# Phase 5 slice 4.1: round and death leftovers (spec)

Date: 2026-11-12. Written at `master` `12e4848a` (slice 3 merged as #34). Branch
`strip/phase-5-4-1-leftovers`. Builds on `2026-11-06-phase-5-rules-seam-design.md` (slice 4 row,
"Notes for later slices"), `gomud_strip_plan.md` decisions 1, 5, 7 and 12, and slice 3's spec
(`docs/completed/2026-11-10-phase-5-3-round-death-spec.md`).

**Goal.** Clear what slices 2 and 3 left in the engine for slice 4: the health floors, the health
checks in `characters` and `rooms`, NPCs that never regenerate, offhand breakage, `BackStab` and
`actors.BodyReplaced`. After this slice the only code that compares a body's health with a number is
the ruleset's, and the death guard covers `characters` and `rooms` too.

## Rulings this spec implements

Earlier, still binding: decision 7 (die at 0, no downed state; the engine asks the ruleset whether
an actor is dead), decision 12 (code the shrink deletes goes straight to its bare form, with a golden
recorded first), one actor model, rulesets return messages and never send.

Calabe:

1. **Slice 4 is cut into three sub-slices, each with its own spec, plan and PR** (2026-11-12):
   4.1 this spec; 4.2 progression (levels, XP, alignment, training and stat points, the PvP minimum
   level and their displays); 4.3 the body (the stat block, `statmods`, mana, race fields, action
   points and carry capacity, skills, the simple hit roll with weapon damage).
2. **Offhand breakage is deleted** (2026-11-10, slice 2 notes).
3. **The `BackStab` branch is deleted** (2026-11-10, slice 2 notes).
4. **Every actor regenerates by one rule** (the one-actor ruling, decision 4; named for slice 4 in
   the phase design's slice 4 row).
5. **The design for this slice** (2026-11-12, brainstorm): no health floors in the engine; the
   `downed` display and the killer record follow the ruleset; dead code goes.

## Facts verified against source (2026-11-12, `12e4848a`)

Read or grepped by the controller today.

| Fact | Value | Where |
|---|---|---|
| Import cycle | `characters` cannot import `rules` (`rules` → `actors` → `characters`) | slice 2 spec; `go list` importers of `internal/rules` |
| Prior art for the cycle | `rooms.SendMessage` is a func variable that `messaging.Init` sets at boot (`rooms.go:37-40`, nil-checked at `:265`; `messaging/interface.go:25-32`) | read |
| `rules.Get` | panics when no ruleset is set (`rules.go:130-137`); `SetForTest` returns a restore func (`:139-145`) | read |
| `Condition` | `dikubasic`: `Dead` at `Health <= 0`, else `Up` (`ruleset.go:116-123`) | read |
| Death timing | `DoCombat` applies damage in `attack` (`NewRound_DoCombat.go:73-91`) and runs `death.Die` for the dead in `handleAffected` (`:476-515`), in the same listener call. Prompt redraws and vitals are queued events (`messaging/wire.go:7`, `:42`, `:91`, `:124`; `NewRound_DoCombat.go:249`) and read the body when processed (`WritePrompt`, `wire.go:165-175`), after death has respawned it at 1 or ended it | read |
| Damage floors | `ApplyHealthChange` (`character.go:896-916`): below 0 it cancels `cancel-on-combat` buffs (`:899-900`); from above 0 a hit stops at -5 (`:904-905`); otherwise floor -10 (`:906-907`) | read |
| Load floor | `Validate` (`character.go:1163`) clamps health to -10 (`:1205-1207`) | read |
| Health checks in `characters` | `GetAdjectives` `Health < 1` adds `downed` (`:422`); `getFormattedName` shows health % only when `Health > 0` (`:463`) and sets the `downed` suffix at `Health < 1` (`:470`); `ForgetKillerIfUp` returns early at `Health < 1` (`:710-711`); `IsDisabled` is `Health <= 0` (`:819-821`); 13 bound-gear checks `IsRemoveLocked() && c.Health > 0` in `Wear` (`:1465`) and `RemoveFromBody` (`:1645-1649`) | grep `Health` in `character.go` |
| `ForgetKillerIfUp` callers | `ApplyHealthChange` (`:914`), `LevelUp` (`:984`), `Heal` (`:1001`), `death.go:110` (revive) | grep |
| `Character.IsDisabled` callers | tests only (`character_test.go:1195-1232`); the slice 3 death guard bans it (`guards/death_test.go:20-35`, `:79`) | grep |
| `downed` style | `formattedname.go:26` (`☠downed`, `☠`, red) | read |
| Health checks in `rooms` | `FindDowned` flag (`rooms.go:55`) read at `:1144` and `:1219`; nothing passes it | grep `FindDowned` over the repo |
| Death guard scope | `deathScanned` = `hooks`, `usercommands`, `death` (`guards/death_test.go:12-16`; comment says slice 4 moves the `characters` checks) | read |
| Regen gate | `dikubasic.Regen` returns nothing for a body without an account (`ruleset.go:134`); comment `:125-129`; test `TestRegenOutsideTheRecoveryRoomIsTheOldAutoHeal` (`ruleset_test.go:241`) | read |
| Regen pass | `AutoHeal` runs `Regen` for every body in `fighters()` (accounts, then NPCs) every round (`NewRound_AutoHeal.go:17-44`) | read |
| Regen amounts | `HealthPerRound` = `1 + StatMod(HealthRecovery)` (`character.go:1006`); `ManaPerRound` = `1 + StatMod(ManaRecovery)` (`:1017`) | read |
| Golden regen | the combat golden's `heal` helper runs `AutoHeal` on rounds 3, 6, 9 ... (`combat_golden_test.go:537-551`); used at `:781`, `:805` | read |
| Offhand breakage | `offhandBreakCheck` (`NewRound_DoCombat.go:341-390`), called at `:245`; on a break it makes `items.New(20)`, which no world file defines (`world/empty/items/other-0` holds only `1-note.yaml`) | read, `ls` |
| Break fields | `Item.BreakTest` (`items.go:100`); `ItemSpec.BreakChance` (`itemspec.go:212`, clamp `:488-490`); admin item editor `items.html:295`, `:508`, `:648`; golden fixture `shield(..., breakChance)` (`combat_golden_test.go:270-273`, used `:611` with 90) | grep `BreakChance\|BreakTest\|breakchance` |
| World data | no `breakchance:` key in `_datafiles`; item YAML loads by plain `yaml.Unmarshal` (`fileloader.go:76`, `:181`), which ignores unknown keys | grep |
| Break in goldens | `pvp.txt:41`, `:48` | grep `breaks` |
| `BackStab` | `AggroType` = `DefaultAttack`, `Shooting`, `BackStab`, `Flee` (`aggro.go:7-13`); read only at `dikubasic/combat.go:177-182`; nothing sets it | grep over the repo |
| Saved aggro | NPC sidecars save `Aggro` (`mobs/saved.go:20`), its `Type` as an int; no shipped YAML holds an `aggro:` key | grep `_datafiles` |
| `BodyReplaced` | `actors/registry.go:180-190`; callers: its own test only (`registry_test.go:185-192`) | grep |
| Overkill in goldens | `mvm.txt:37` `rat: ... hp=-5/12` | grep |

## Design

### The ruleset says whether a body is up, inside `characters`

`characters` gets one func variable, wired at boot the way `rooms.SendMessage` is:

```go
// IsUp reports whether the ruleset counts c as up (rules.Condition Up).
// internal/rules sets it; characters cannot import rules.
var IsUp func(c *Character) bool
```

`internal/rules` sets it so that `IsUp(c)` is `Get().Condition(a) == Up` for the actor whose body is
`c`. Every body check in `characters` goes through it:

| Today | After |
|---|---|
| `GetAdjectives`: `Health < 1` adds `downed` | `!IsUp(c)` adds `downed` |
| `getFormattedName`: health % when `Health > 0` | health % when `IsUp(c)` |
| `getFormattedName`: `downed` suffix at `Health < 1` | `downed` suffix when `!IsUp(c)` |
| `ForgetKillerIfUp`: keep the record at `Health < 1` | keep the record when `!IsUp(c)` |
| 13 bound-gear checks: removable when `Health <= 0` | removable when `!IsUp(c)` |

Under `dikubasic` every row behaves as today: "not up" is "dead", and death follows in the same
round.

**Who sees a difference.** Nobody under `dikubasic`: a body is never "not up" while anyone can see
it, since death runs in the same round, before any prompt or look. The change is for a later ruleset
whose downed state starts above 0 health: today `characters` would call that body up until its
health fell below 1; after this slice it follows the ruleset.

The plan fixes how `IsUp` finds the actor for a `*Character` (by `c.Id` in the actor registry, and
what an unregistered body such as a bare test character gets) and where the wiring call lives (in
`rules.Set` and `rules.SetForTest`, or an `Init` beside `messaging.Init`). An unset `IsUp` fails
loudly, as `rules.Get()` does.

### No health floors in the engine

- `ApplyHealthChange` loses the -5 one-hit cap and the -10 floor. Health goes where the damage puts
  it; the cap at `HealthMax` stays. The `cancel-on-combat` cancel below 0 stays as it is.
- `Validate` loses its -10 clamp. A save with health at or below 0 still dies at the next upkeep, as
  today.
- Nothing reads health below 0 before death settles it (facts: death timing). A ruleset with a
  downed state owns its own floor; `ClampDamage` stays unadded, as slice 3 decided, until one does.

### Every actor regenerates by one rule

`dikubasic.Regen` drops its account gate. Any body out of a fight, with health above 0, regains
`HealthPerRound()` health and `ManaPerRound()` mana every 3rd round, capped at its maximums by
`Heal`. The recovery-room rule is unchanged.

Player-visible: an NPC wounded and then left alone heals back over time. Today it stays wounded
until it dies or the zone resets it. Regen lines are unchanged: outside the recovery room, regen
sends none.

### Deleted

- `Character.IsDisabled` and its test (the death guard's ban on it goes with it, since nothing named
  `IsDisabled` on a body remains; the item allowlist stays).
- `rooms.FindDowned` and its two reads.
- `actors.BodyReplaced` and its test.
- Offhand breakage: `offhandBreakCheck` and its call, `Item.BreakTest`, `ItemSpec.BreakChance` and
  its clamp, the admin item editor's break-chance field, and the golden fixture's `breakChance`
  parameter. An item file that still has `breakchance:` loads as before (the key is ignored).
- `BackStab`: the `AggroType` value and the `dikubasic` branch. `Flee` moves from 3 to 2. A sidecar
  saved mid-flee before this slice holds type 3, which no longer names anything; no shipped data
  holds one (facts: saved aggro), as in slice 1. The plan checks what the round does with an unknown
  type and records it.

### Guards

The death guard's `deathScanned` adds `internal/characters/` and `internal/rooms/`, so no body
health comparison returns there. Its comment changes to match. The guard is shown to fail with a
planted `c.Health < 1` in each new package.

## Testing and evidence

**Commit 1 records a golden before anything changes** (decision 12): a fixture where a player wounds
an NPC, the fight ends (the player flees or the NPC's aggro is cleared), and the regen pass runs a few
times. On today's code the NPC's health stays put.

**Goldens that change, each diff read and summarised in its commit and the PR** as "decided" or
"unexpected" (an unexpected change stops the work):

- floors: `mvm` (`hp=-5/12` becomes the raw value) and any other dump line below 0;
- regen: the commit 1 fixture (the NPC's health rises every heal pass), and any fixture whose NPC
  sits out of a fight under a heal pass;
- breakage: `pvp` loses its two break lines and keeps its shield;
- every other combat golden and every reads golden byte-identical.

**Unit tests:**

- `ApplyHealthChange`: a 30 hit on 10 health leaves -20; a heal above max caps at max.
- `Validate`: health -50 stays -50.
- `dikubasic.Regen`: an NPC regenerates every 3rd round out of a fight; not in a fight; not at 0
  health. The accounts-only parity test is replaced.
- `IsUp` under a stub ruleset whose `Condition` returns `Downed` at 5 health: the formatted name
  shows `downed` and no health %; `ForgetKillerIfUp` keeps the killer; a bound item can be removed.
  Under `dikubasic` at 5 health: none of those.
- An item YAML with `breakchance: 50` loads.

**Every new test shown to fail under a prescribed mutation,** as in slices 2 and 3.

**Smoke:** unchanged steps pass with 1, 5 and 12 bots.

## Commits, in order

1. Regen golden on today's code (test files and testdata only).
2. `characters.IsUp` and its wiring; the characters checks through it; `IsDisabled`, `FindDowned`
   and `BodyReplaced` deleted; the death guard widened. Goldens unchanged.
3. Floors gone. `mvm` changes.
4. NPC regen. The commit 1 fixture changes.
5. Offhand breakage gone. `pvp` changes.
6. `BackStab` gone. Goldens unchanged.
7. Docs: patch notes, status, the phase design's slice 4 row (cut into 4.1-4.3) and its "Notes for
   later slices" (this slice's items leave; 4.2's and 4.3's lists are written there),
   `docs/known-issues.md` if it names any of these.

## What this locks in

- **No engine code compares a body's health with a number,** `characters` and `rooms` included
  (guarded). The ruleset's `Condition` is the only line between up, downed and dead.
- **`characters` asks the ruleset through one func variable.** Any later body rule `characters`
  needs uses the same route rather than a second one.
- **Health below 0 is the ruleset's business.** The engine applies damage as given.

## Out of scope

- 4.2 and 4.3 (ruling 1).
- Weapon delay (`characters.SetAggro`) behind the ruleset: waits for swing speed (slice 2 notes).
- The `cancel-on-combat` cancel only firing below 0 health: unchanged (it is a buff rule; slice 6).
- `dikubasic/simulate.go`'s own health checks: ruleset code, outside the guard.
- The test config helpers' pinned overrides (slice 2 notes): nothing trips them.

## Decisions taken without asking

| Decision | Why | If wrong |
|---|---|---|
| `characters` asks the ruleset through a func variable set at boot | Same pattern as `rooms.SendMessage`; the import cycle forbids a direct call | Move each check out to its callers in engine packages that can import `rules`: five call paths, and the next `characters` check would face the same choice |
| The 13 bound-gear checks join this slice instead of waiting for slice 5 | Widening the guard to `characters` would fail on them; they are health checks like the other four | Allowlist them in the guard until slice 5 |
| `IsUp` returns a bool, not `rules.Condition` | `characters` cannot name a `rules` type; every check here only asks "up or not" | Move `Condition` into a package both can import |
| The health % in a formatted name shows only when up | Today's `Health > 0` meant "not downed"; a downed body shows `downed` instead | Show it whenever health is above 0 |
| No floor at all, rather than a 0 floor | Death follows the hit in the same round, so a floor protects nothing; a 0 floor would block a ruleset that lets health run negative | Floor at 0 in the engine |
| The `cancel-on-combat` cancel stays in `ApplyHealthChange` below 0 | Out of this slice's subject; a buff rule | Move it with slice 6 |
| NPC regen uses the same amounts and every-3rd-round timing as players | One rule; the amounts change with 4.3 when the stat mods go | none |
| `ItemSpec.BreakChance` is deleted, not left inert | Nothing would read it; old item files still load | Keep the field |
| The admin item editor loses its break-chance field | It would edit nothing | none |
| Commit 1 records a new regen fixture rather than relying on existing goldens | Decision 12 asks for the change to be visible; no existing fixture leaves a wounded NPC out of a fight | none |
| Sub-slice names are 4.1, 4.2, 4.3 and the branch `strip/phase-5-4-1-leftovers` | Avoids "4b", already the name of the NPCs-are-players phase | Rename |
