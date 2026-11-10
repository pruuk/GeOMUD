# Phase 5 slice 2: the seam and the fight (spec)

Date: 2026-11-10. Written at `master` `c1ca8e7f`. Branch `strip/phase-5-2-seam-fight`.
Builds on `2026-11-06-phase-5-rules-seam-design.md` (slice 2 row, "Notes for later slices"),
`2026-11-06-phase-5-rulings.md` and `2026-11-06-phase-5-inventory.md` N. Inventory N was written at
`9131eac2`, before slice 1; every fact below was re-read at `c1ca8e7f`.

**Goal.** A `Ruleset` interface between the engine and the fight, with today's combat code moved
behind it unchanged as the reference ruleset `dikubasic`. Proof: all 9 combat goldens and every reads
golden byte-identical with no `-update`, and stub rulesets that change the fight with no engine edit.
In its own commit, `shoot` asks the same "may I attack?" question as `attack`.

## Rulings this spec implements

From earlier (still binding): decision 1 (move with parity, then shrink), decision 5 (bare reference
ruleset later), ruling 1A (diku-basic keeps GoMud's quirks until the real ruleset), the 2026-11-06
default "`shoot` skips `CanPvp` and sends no aggro event: fix to match `attack`", one actor model,
rulesets return messages and never send.

Calabe, 2026-11-10, in brainstorming:

1. **Shooting across rooms needs PvP allowed in both rooms** (option A). You cannot shoot out of a
   safe room or into one.
2. **No ruleset set means fail loudly.** `rules.Get()` panics; there is no silent fallback to
   `dikubasic`.
3. **The weapon delay stays for now.** Calabe would rather have swing speed: weapon speed plus stats
   and skill set swings per weapon per round, possibly 0 with progress carried to the next round (a
   heavy crossbow that must be wound). Out of scope; recorded below.
4. **Targeting and aggro tracking** better than the single aggro flag: wanted later, out of scope for
   the slim base.

## Facts verified against source (2026-11-10, `c1ca8e7f`)

Read or grepped by the controller today.

| Fact | Value | Where |
|---|---|---|
| `internal/combat` files | `armor_rank.go` 266, `attackresult.go` 16, `calculations.go` 485, `combat.go` 339, `mob_rank.go` 236, `simulate.go` 188, `weapon_rank.go` 220; tests `calculations_test.go` 358, `simulate_test.go` 93, `weapon_rank_test.go` 68; upstream `AGENTS.md`, `CLAUDE.md` | `wc -l`, `ls -a` |
| `combat` imports | accounts, actors, buffs, characters, configs, ids, items, mobs, mudlog, perception, races, rooms, statmods, stats, util | `go list` |
| `combat` importers | `hooks/NewRound_DoCombat.go` (`GetWaitMessages` :167, `Attack` :182, `CreditAccount` :214); `usercommands/consider.go` (`CombatOdds` :45, :50); `usercommands/suicide_npc.go` (`AlignmentChange` :149, :214); `web/api_v1_items.go` (`RankWeapons` :18, `RankArmor` :34); `web/api_v1_mobs.go` (`RankMobs` :24) | `grep -rln 'internal/combat"'`, `grep -n 'combat\.'` |
| `combat.Attack` | `combat.go:29-50`: `calculateCombat(*a.Character, *d.Character)`; `ApplyHealthChange(-dmg)`; `TrackPlayerDamage` when `CreditAccount(a)` is non-nil; `WimpyCheck` when damage != 0 and the defender has an account. `CreditAccount` (:63) just returns `accountOf` (:53) | read |
| `AttackResult` | `attackresult.go:6-16`: `Hit`, `Crit`, `BuffTarget []int`, `DamageToTarget`, `Messages []perception.Message` | read |
| Import graph | `actors` imports accounts, characters, ids, mobs, mudlog, perception, util. `characters` imports neither `actors` nor `rooms`. `rooms` imports `actors` | `go list` per package |
| Consequence | `internal/rules` can import `actors` and `perception`. `characters` cannot call `rules.Get()` (cycle `rules` → `actors` → `characters`), so the weapon delay set in `SetAggro` (`character.go:788-812`, weapon plus offhand `WaitRounds`) cannot move behind the ruleset without moving aggro setting out of `characters` | derived from the row above |
| `DoCombat` | `NewRound_DoCombat.go`: `fightersWithAccountsFirst()` :33, defined :49-62 (online user IDs, then all mob IDs, each in map order); `combatTurn` :80; `lookForWeapon` call :158-160, roll :305-309; wait lines :162-172; hidden skip :175-178; `combat.Attack` :182; `KilledBy` :196-207; group hatred :213-220 (`MinutesToRounds(2) - a.Character.Stats.Perception.ValueAdj` :217); offhand break :223-225, rule part :333-340; flee chance :406-417; `handleAffected` :467-511 | read |
| Signatures | `Timing.MinutesToRounds(int) int` (`config.timing.go:45`); `mobs.MakeHostile(string, int, int)` (`mobs.go:434`); `rooms.MoveToRoom(ids.ID, int, ...bool) error` (`roommanager.go:281`) | grep |
| `consider` | `consider.go:43-53`: `accounts.GetByUserId` or `mobs.GetInstance` with no nil check, then `combat.CombatOdds(*actor.Character, ...)`; labels :58-74 | read |
| `attack` | NPC target :186-228, no PvP check; player target :230-279, `room.CanPvp(actor.Character, p.Character)` :234, before the party check :239 | read |
| `shoot` | NPC target `SetAggroRemote` :97; player target party check :113-120, `SetAggroRemote` :122; no `CanPvp`, no `AggroChanged` | read |
| `CanPvp` | `rooms.go:2061-2089`, one caller (`attack.go:234`). Refusals: "Fighting is not allowed here." (attacker in room -1 or the recovery room, :2064), "PVP is disabled." (:2074), "Players must be at least level %d to PVP." (:2078), "This is not a PVP area." (limited and `!r.IsPvp()`, :2085) | read |
| PvP config | key `GamePlay.PVP.Enabled`; values `enabled`, `disabled`, `off`, `limited` (`configs.go:17-20`); invalid or blank validates to `enabled`, `MinimumLevel` floor 0 (`config.gameplay.go:153-162`) | read |
| Guards naming combat | `guards/combat_test.go`: `TestOneCombatPath` control on `internal/combat/combat.go` func `Attack` (:41); `TestCombatPackageDoesNotSend` scans `internal/combat/` (:69), control `scanned < 5` (:100). `goFiles` parses non-test files only (`guards_test.go:30`) | read |
| Combat goldens | 9: `flee_blocked`, `flee_success`, `hidden_attacker`, `mvm`, `mvp`, `observer`, `pvm_death`, `pvp`, `shoot_exit`. `pvm_death` takes 4 rounds and queues `Input rat: "suicide"` in round 4 | `ls`, grep |
| Golden harness | `combat_golden_test.go`: `newCombatWorld(t, seed)` :65, `addPlayer` :197, `addNPC` :216, `weapon` :240, `start` :344, `rounds` :367; the record is `w.out`, per-round reads `w.reads`. `shoot_exit` sets its aggro directly, not through the command | read |
| Test setup | `TestMain` in `hooks/actionpoints_test.go:22` and `usercommands/usercommands_test.go:25` (8 other packages have one; none runs combat). Helpers `goSetup` (`go_test.go:18`; room 92001, exit `north` to 92002), `addPlayer` (`combat_text_test.go:18`), `playerTexts` (`go_test.go:34`), `capture[T]` (`items_test.go:48`), `accounts.SetTestUser` | read |
| Boot | `main.go:206` `hooks.RegisterListeners()`; smoke builds and runs the real binary (`cmd/smoke/server.go:88`) | read |
| Smoke fight strings | "rat has died.", "You gained ... experience points ... (combat)" (`cmd/smoke/fight.go:88-99`) | grep |
| `BackStab` | declared `characters/aggro.go:11`, read only at `combat.go:213`; nothing sets it | grep |

## Design

### Packages

**`internal/rules`** (new; imports only `actors` and `perception`):

```go
type AttackResult struct { // moved from combat, unchanged
	Hit            bool
	Crit           bool
	BuffTarget     []int
	DamageToTarget int
	Messages       []perception.Message
}

type Ruleset interface {
	TurnOrder(fighters []*actors.Actor) []*actors.Actor
	ResolveAttack(attacker, defender *actors.Actor) AttackResult
	WaitLines(attacker, defender *actors.Actor) []perception.Message
	HatredRounds(attacker *actors.Actor) int
	FindsWeapon(npc *actors.Actor) bool
	OffhandBreaks(defender *actors.Actor, crit bool) bool
	FleeEscapes(fleer, blocker *actors.Actor) bool
	Consider(attacker, defender *actors.Actor) float64
	CanAttack(attacker, defender *actors.Actor) error
}

func Set(r Ruleset)
func Get() Ruleset                          // panics when none is set (ruling 2)
func SetForTest(r Ruleset) (restore func()) // for t.Cleanup
```

| Method | Contract |
|---|---|
| `TurnOrder` | Gets every actor with a body, in the engine's collection order; returns them in turn order |
| `ResolveAttack` | One round of the attacker's attacks: hits, damage, crit buffs, one line per swing. Changes no health |
| `WaitLines` | Lines for a round the attacker spends waiting to swing |
| `HatredRounds` | Rounds an NPC's groups hate the attacker's account after it attacks the NPC |
| `FindsWeapon` | Whether an unarmed NPC carrying items looks for a weapon this round |
| `OffhandBreaks` | Whether a hit breaks the defender's offhand item; called only when there is one |
| `FleeEscapes` | Whether the fleer gets past one actor fighting it |
| `Consider` | Odds: above 1 favours the attacker, 1 is even, 0 or less is hopeless |
| `CanAttack` | nil, or the refusal the attacker reads |

A method that needs a random draw makes it itself, at the point in the sequence where the engine
makes it today.

**`internal/rules/dikubasic`.** `internal/combat` moves here by `git mv`; only the package line and
import paths change. Upstream's `AGENTS.md` and `CLAUDE.md` in that folder are deleted. A new
`ruleset.go` holds `type Ruleset struct{}`, `New()`, and the methods:

- `ResolveAttack` is `calculateCombat`; `WaitLines` is `GetWaitMessages(items.Wait, ...)`;
  `Consider` is `CombatOdds`.
- `HatredRounds`, `FindsWeapon`, `OffhandBreaks` and `FleeEscapes` hold the lines now inline in
  `DoCombat` (facts table), moved unchanged, `util.LogRoll` calls included.
- `TurnOrder` is a stable partition: accounts first, then NPCs, each in the given order. On today's
  input (already accounts first) it returns the input unchanged.
- `CanAttack` allows any NPC target. For an account target it runs `CanPvp` with the attacker's room,
  then, when the defender is in another room, with the defender's room too (ruling 1). A missing
  attacker room refuses with "Fighting is not allowed here."
- `combat.Attack` and `CreditAccount` are deleted (their work moves to the engine); `accountOf`
  stays for `TurnOrder` and `CanAttack`.

**`main.go`** calls `rules.Set(dikubasic.New())` immediately before `hooks.RegisterListeners()`.

### What stays in the engine

Applying damage, kill credit on every swing, the wimpy check, `KilledBy`, aggro and its events,
target lookup and range, the hidden-target skip, flee movement, item and message handling for a
broken offhand, NPC retaliation movement, every send, and (until slice 3) the health thresholds and
`handleAffected`.

### The round after the change

1. `fighters()` collects every actor with a body exactly as `fightersWithAccountsFirst` does today
   (renamed: ordering is now the ruleset's), and `rules.Get().TurnOrder(fighters())` orders it.
2. Per actor, unchanged: aggro and `no-combat` checks, room check, `cancel-on-combat`, NPC combat
   commands, flee, target lookup, "Your rage subsides".
3. An unarmed NPC with items: `FindsWeapon(a)`; if true the engine picks the weapon and queues
   `equip` as today.
4. Wait rounds: count one down and send `WaitLines(a, d)`.
5. Hidden target skip, unchanged.
6. The engine's new `attack(a, d)`: `ResolveAttack`, apply `-DamageToTarget`, credit the attacker's
   account, wimpy check on damage. Same statements, same order, as `combat.Attack` today.
7. Crit buffs queued, lines sent in order, `KilledBy` recorded, unchanged.
8. NPC defender and an attacker with an account: `HatredRounds(a)` once, applied to each group.
9. On a hit with an offhand item present: `OffhandBreaks(d, crit)`; the engine moves the item and
   sends the lines.
10. Vitals event, NPC retaliation, end or renew aggro, `handleAffected`: unchanged.

**Flee:** for each body fighting the fleer, `FleeEscapes(fleer, blocker)`; the first `false` blocks.

**`consider`:** resolves the target as an actor (`actors.ByUserId` or `actors.Get`), refuses a nil
one with the existing "`<name>` can't be found." line, calls `Consider`, maps the odds to the same
seven labels.

**`attack`:** `CanAttack` for every target. NPC target: first thing once the NPC is found, before the
party auto-attack. Player target: in place of `room.CanPvp`, at the same spot before the party check.
The refusal is told to the attacker and the command returns, as today.

**`shoot`** (own commit): `CanAttack` placed as in `attack`, and `AggroChanged` queued after each
`SetAggroRemote`.

### Guards

- `TestOneCombatPath`: positive controls become the engine's `attack` in `NewRound_DoCombat.go` and
  the method `ResolveAttack` in `internal/rules/dikubasic/ruleset.go`.
- `TestCombatPackageDoesNotSend` becomes `TestRulesetDoesNotSend` and scans `internal/rules/`.
- New `TestOnlyBootNamesTheRuleset`: no non-test file outside `internal/rules/dikubasic/` imports
  `dikubasic`, except an allowlist where each entry names the slice that removes it:
  `main.go` (boot), `internal/usercommands/suicide_npc.go` (`AlignmentChange`; slice 3),
  `internal/web/api_v1_items.go` and `internal/web/api_v1_mobs.go` (rankings; slice 4). An entry
  that no longer imports `dikubasic` fails (no stale list), and `main.go`'s entry is the positive
  control.

## Testing and parity evidence

- **Parity:** all 9 combat goldens and every reads golden byte-identical, no `-update` anywhere in
  the slice; a golden diff stops the work for diagnosis. The moved unit tests run unchanged in
  `dikubasic`. Smoke (`go run ./cmd/smoke`, `-bots 5`) passes its fight strings.
- **Stub rulesets** (each test file is the only change it needs):
  - one-shot: the `pvm_death` fixture under `ResolveAttack` = 1000 damage ends in round 1, not 4,
    with `suicide` queued and the stub's line read;
  - reversed turn order: NPCs first, so Alice reads the guard's swing before her own in round 1;
  - pacifist: `CanAttack` always refuses; `attack Rat` gives the refusal and no aggro.

  Each is shown to fail once with its `SetForTest` line removed.
- **`shoot`**, test first: PvP disabled gives "PVP is disabled." and no aggro; PvP limited with the
  shooter in a PvP room and the target in a non-PvP room gives "This is not a PVP area." (the ruling 1
  case); PvP enabled queues exactly one `AggroChanged` with target and exit set; shooting an NPC
  queues `AggroChanged`.
- **Guards:** the import guard is shown to fail twice (an unlisted importer added; a stale entry
  added).
- **`rules`:** `Get` panics when unset; `SetForTest` restores the previous ruleset.

## Commits, in order

1. `internal/rules` with its tests.
2. The move (`git mv`, package line, import paths, the two `.md` files). Goldens unchanged.
3. The rewire: `dikubasic.Ruleset`, the engine's `attack`, `DoCombat`, flee, `consider` and `attack`
   on `rules.Get()`, `main.go`, `rules.Set` in the affected `TestMain`s, guard controls. Goldens
   unchanged.
4. The import guard.
5. The stub-ruleset tests (test files only).
6. `shoot` through `CanAttack` with `AggroChanged`, tests first. Player-visible: shooting a player
   now obeys the PvP settings.
7. Docs: `PATCH_NOTES.md`, `status.md`, the phase design's "Notes for later slices".

## What this locks in

- The engine reaches the rules only through `internal/rules`. A new rule question becomes a new
  interface method, never an import of a ruleset.
- Rulesets return messages and never send (today's no-send guard, widened to every ruleset).
- Rule A is the reference ruleset's PvP rule for any fight across rooms.
- The interface takes `*actors.Actor`, so a ruleset sees the whole body (character, controller,
  account). Convenient, and wide; narrowing it later touches every method.

## Out of scope

- **Swing speed** (ruling 3): swings per weapon per round from weapon speed, stats and skill,
  possibly 0 with carried progress. When it comes, the weapon delay moves behind the ruleset with it.
  The seam leaves room: `ResolveAttack` already decides the attack count, and carried progress needs
  the per-body value store planned for slice 5.
- **Targeting and aggro tracking** beyond the single aggro flag (ruling 4).
- Slice 3: health thresholds, regen, death, round upkeep.
- NPC aggression and initiative (ruling 1A: they wait for the real ruleset).
- Splitting `ResolveAttack` finer (not before parity holds).

## Decisions taken without asking

| Decision | Why | If wrong |
|---|---|---|
| The interface is the fight only (the nine methods above); `Condition`, `Regen`, `RoundUpkeep`, `OnDeath` wait for slice 3 | The phase design gives slice 3 the round, regen and death | Slice 3 adds them anyway; nothing here blocks it |
| Two methods beyond the phase design's list: `WaitLines` and `FindsWeapon` | Both are rule code (the wait lines share the swing-line builder and draw RNG; the weapon search rolls against Perception); leaving them would keep the engine importing `dikubasic` or naming a stat | Fold or drop them in the slice 4 shrink |
| The design's `BreaksOnHit` is named `OffhandBreaks`; `AttackDelay` is not in the interface | Only the offhand can break today; the delay cannot move this slice (import cycle, facts table) and ruling 3 replaces it later | Rename is one identifier; `AttackDelay` comes with swing speed |
| Each rule method makes its own random draw, at today's point in the sequence | The goldens pin the draw order; a ruleset that does not roll should not be made to | Methods return chances and the engine rolls; same goldens |
| `TurnOrder` receives every fighter in today's collection order and `dikubasic` partitions accounts first | Identity on today's input, so parity holds, while a stub can reorder | Engine sorts nothing either way; only the method body changes |
| `CanAttack` is called for NPC targets too, and `dikubasic` allows them | One check for every attack (design: "one check for `attack` and `shoot`"); the pacifist stub proves NPC targets reach it | Call it for player targets only; the pacifist test changes |
| An import guard with a temporary allowlist lands now, though the design puts the guard test in slice 5 | The import ban is cheap and stops new importers in the meantime; the slice 5 guard adds the value-name bans | Delete it; slice 5 adds its own |
| Three files keep importing `dikubasic` until slices 3 and 4 | Moving them now drags death and the admin rankings into this slice | Move them now, at the cost of a larger slice |
| `consider` resolves its target as an actor and nil-checks it | The method takes actors; the nil check fixes the inventory's crash path for free | Keep the old lookups and add only the nil check |
| `fightersWithAccountsFirst` is renamed `fighters` | Ordering is the ruleset's now, so the old name would lie | Keep the name; cosmetic |
| Upstream `AGENTS.md`/`CLAUDE.md` in `internal/combat` are deleted in the move | Project `CLAUDE.md`: they go as their packages are stripped | Keep them in `dikubasic`; they describe GoMud, not us |
| `BackStab` (never set) is left alone; slice 4 deletes the branch | Parity slice; the branch never runs | Delete now; no golden would change |
