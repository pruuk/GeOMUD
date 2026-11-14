# Phase 5 slice 4.1: round and death leftovers (spec)

Date: 2026-11-12. Written at `master` `12e4848a` (slice 3 merged as #34). Branch
`strip/phase-5-4-1-leftovers`. Builds on `2026-11-06-phase-5-rules-seam-design.md` (slice 4 row,
"Notes for later slices"), `gomud_strip_plan.md` decisions 1, 5, 7 and 12, and slice 3's spec
(`docs/completed/2026-11-10-phase-5-3-round-death-spec.md`).
**Approved by Calabe 2026-11-12; amended the same day with rulings 6-9 (retaliation and target
switching), which came out of plan review round 1. Amendment approved by Calabe 2026-11-12.
Rulings 10-12 added the same day from Calabe's answers during plan review; ruling 13 the same day,
its detail settled by Calabe 2026-11-14.**

**Goal.** Clear what slices 2 and 3 left in the engine for slice 4: the health floors, the health
checks in `characters` and `rooms`, NPCs that never regenerate, offhand breakage, `BackStab` and
`actors.BodyReplaced`. After this slice the only code that compares a body's health with a number is
the ruleset's, and the death guard covers `characters` and `rooms` too. And Diku-style fighting:
being attacked makes you fight back, and when your target is gone you turn to the next one.

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
6. **Being attacked makes you retaliate** (2026-11-12, after plan review round 1 found that an NPC
   being beaten would regenerate mid-fight because nobody ever fights back). A body that is attacked
   and is not already fighting fights its attacker. A body already fighting keeps its target;
   `attack <name>` switches, as today. "Requiring the player to type `attack <name>` when they are
   being hit" is not Diku-style. Under the one-actor ruling this holds for NPCs too: NPCs fight back.
7. **When your target is gone, you turn to the next opponent** (2026-11-12): dead, fled, or left the
   room. The next opponent is whoever in your room has been attacking you longest. A target you
   chose yourself (`attack Z`) is never replaced while it is reachable.
8. **An attacker you cannot reach still makes you fight** (2026-11-12): attacked from another room
   (shot through an exit) or by an attacker you cannot perceive, you take an aggro on it. You swing
   only once you can reach it.
9. **A target in another room: shoot back if you can, otherwise wait** (2026-11-12). With a shooting
   weapon and the target in a room one of your exits leads to, you shoot it. Otherwise your aggro
   waits on it, and you swing as soon as you share a room.
10. **PvP settings apply only between two players** (2026-11-12, plan question Q16). NPCs may
    attack, shoot and retaliate against players of any level in any room except the recovery room,
    which is safe from everyone. This reverses slice 2's line that an NPC shooting a player obeys
    the PvP settings (`docs/completed/2026-11-10-phase-5-2-seam-fight-spec.md`, commit 6), and its
    test `TestNPCShootingAPlayerRefusedWhenPvPDisabled` becomes "allowed".
11. **Smoke fights a real rat and checks both endings on every run** (2026-11-12, after plan review
    round 2 proposed making the smoke world's rats deal no damage, which hid retaliation from the
    only end-to-end test). The smoke world copy spawns two rats in the fight room, one at level 1 and
    one at level 5, with the rodent race's damage unchanged. Bot 0 fights the level 1 rat first, then
    the level 5 rat. Each fight may end either way, and the step checks whichever ending happens: the
    rat's death ("rat has died." and the XP line), or bot 0's (its death line, the recovery room, and
    the walk back to the fight room for the steps after it). In the plan's simulation bot 0 beats the
    level 1 rat about 98% of the time and never beats the level 5 rat, so almost every run covers
    both endings; pinning the levels also removes the fight-timeout flake from high-level rats.
12. **A waiting aggro ends after a set number of rounds out of reach, for every body** (2026-11-12,
    after plan review round 2 found players waited forever without regen while NPCs gave up within
    a round). Default 5 rounds, a config value. One rule for players and NPCs replaces the NPC-only
    `behaviour.loseLostQuarry`.
13. **Shooting is one shot per action** (2026-11-12, after plan review round 6 found the plan made
    `shoot` keep firing every round). `shoot` fires once, as today, and shooting back under ruling 9
    is one shot too; neither repeats each round. Its detail (2026-11-14, Calabe): after the shot the
    aggro **waits** (lookup step 5: it swings if the target comes into its room, and loses track
    after the wait limit), and **each typed `shoot` earns one shot back**, while a shot back earns
    nothing, so two shooters cannot trade shots forever without anyone typing.

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
| Health checks in `characters` | `GetAdjectives` `Health < 1` adds `downed` (`:422`); `getFormattedName` shows health % only when `Health > 0` (`:463`) and sets the `downed` suffix at `Health < 1` (`:470`); `ForgetKillerIfUp` returns early at `Health < 1` (`:710-711`); `IsDisabled` is `Health <= 0` (`:819-821`); 13 bound-gear checks `IsRemoveLocked() && c.Health > 0`: 12 in `Wear` (`:1465`; checks at `:1523`-`:1631`) and 1 in `RemoveFromBody` (`:1649`) | grep `Health` in `character.go` |
| Item `IsDisabled` in `characters` | `character.go` calls `items.Item.IsDisabled` 16 times (slot checks), so a widened death guard needs it on its item allowlist | plan agent, re-grepped by the controller |
| `ForgetKillerIfUp` callers | `ApplyHealthChange` (`:914`), `LevelUp` (`:984`), `Heal` (`:1001`), `death.go:110` (revive) | grep |
| `Character.IsDisabled` callers | tests only (`character_test.go:1195-1236`); the slice 3 death guard bans it (`guards/death_test.go:20-35`, `:79`) | grep |
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
| Backstab leftovers | `items.CanBackstab` (`itemspec.go:364`), no callers; the smoke client strips a `*[BACKSTAB]*` prefix (`cmd/smoke/fight.go:122-133`, comment and function) with two test rows (`fight_test.go:44-45`); the admin weapon-rank tooltip names backstab (`items-rank-weapons.html:208`) | grep `-i backstab` |
| Break draws | `BreakTest` draws `util.Rand(100)` whenever the item's break chance is above 0 (`items.go:100-106`); the `pvp` fixture's shield has 90, so every hit on Bob draws once and removing it shifts later draws | read |
| Crit cascade | `combat.go:245` `isCrit = attackResult.Crit \|\| Crits(...)`, and `:248` sets `attackResult.Crit = true` after a crit, so every later hit in that round crits with no roll; the comment at `:243-244` says the opposite | read |
| Saved aggro | NPC sidecars save `Aggro` (`mobs/saved.go:20`), its `Type` as an int; no shipped YAML holds an `aggro:` key | grep `_datafiles` |
| `BodyReplaced` | `actors/registry.go:180-193`; callers: its own test only (`registry_test.go:185-195`) | grep |
| Overkill in goldens | `mvm.txt:37` `rat: ... hp=-5/12` | grep |

Facts for rulings 6-9, read by the controller 2026-11-12 at `1c8e607f`:

| Fact | Value | Where |
|---|---|---|
| Who gets an aggro | only `attack` (`attack.go:230`, `:284`; party members are told to `attack` too, `:220`, `:276`), `shoot` (`SetAggroRemote`, `shoot.go:104`, `:136`), and the round re-arming the attacker (`NewRound_DoCombat.go:273`) and flee (a blocked or exitless flee resumes the fight, `:398`; every flee first resets the aggro to a plain attack, `:411`). Nothing gives a defender an aggro | grep `SetAggro(` |
| Defenders today | the `pvp` golden's Bob is hit three rounds with `aggro=[none]` (`pvp.txt:16`, `:26`, `:37`); `pvm_death`'s rat likewise (`pvm_death.txt:5`, `:13`) | read |
| The turn | `combatTurn` (`NewRound_DoCombat.go:102`): no aggro, no turn (`:104-106`); `no-combat` skips (`:109`); flee (`:130`); target looked up by ID, in the room or through the aggro's `ExitName` (`:143-154`); not found: "Your target can't be found." and the aggro ends (`:156-160`); target not up: "Your rage subsides." and the aggro ends (`:170-174`); weapon wait rounds (`:184-194`); target not perceived: "You can't seem to find your target." and the turn ends, aggro kept (`:197-200`); `attack` (`:202`) (corrected by plan review round 2) | read |
| NPC give-up | `behaviour.loseLostQuarry` (`behaviour/tick.go:74-84`), run by `ActorTick` every round after combat and regen (`hooks.go:32`): an NPC whose target is a player who is offline or in another room emotes "mumbles about losing their quarry." and its aggro ends; a vanished NPC target is treated the same, but an NPC target in another room is never lost (found by plan review round 2; corrected in round 3) | read |
| After the swing | an NPC defender with no aggro stops idling and, attacked from another room, walks there and runs one of its angry commands (`:252-266`); if either side is not up both aggros end, else the attacker's aggro is re-armed (`:269-275`) | read |
| Turn order | `dikubasic.TurnOrder`: accounts first, then NPCs, each in the order given (`ruleset.go:26-41`); the round takes the order once, before any turn (`NewRound_DoCombat.go:34-36`) | read |
| Shooting | needs a wielded weapon of subtype `Shooting` (`shoot.go:40`) and the target in the room an unlocked exit leads to (`:59-81`); `SetAggroRemote` sets the aggro and its `ExitName` (`character.go:784-787`); `SetAggro` turns a `DefaultAttack` into `Shooting` when the weapon shoots (`:801-803`) | read |
| Aggro record | `Aggro{Type, Target, ExitName, RoundsWaiting, FleeFrom}` (`aggro.go:15-23`); no record of when it began. Players do not save it (`character.go:73`, `yaml:"-"`); NPC sidecars do (`mobs/saved.go:20`) | read |
| A shot today | `shoot` sets a shooting aggro with its exit (`SetAggroRemote`); after the swing the round re-arms through `SetAggro` (`NewRound_DoCombat.go:273`), which builds a new `Aggro` without `ExitName` (`character.go:807-811`), so the next turn finds the target in another room with no exit, reads "Your target can't be found." and ends the aggro (`:146-160`). One shot per `shoot` (read 2026-11-14 at `f8e96841`) | read |
| Aggro on death | `death.clearKillersAggro` ends the aggro of each account credited with damage that still targets the victim (`death.go:264-275`); NPCs targeting the victim are not cleared there | read |

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

"Out of a fight" stays "has no aggro of its own". With retaliation (ruling 6) every body that is
attacked has one, so nobody regenerates while being fought, NPCs included. A body whose aggro waits
on a target in another room (ruling 9) does not regenerate either: it is still in that fight.

### Retaliation and target switching (rulings 6-9)

All of this is engine code in the round (`NewRound_DoCombat.go`) and `internal/death`: who fights
whom is the engine's (phase design, "What the engine keeps"). It applies to every actor.

**When an aggro begins.** The aggro record gains the round it began on its target (`Since`, a round
number), set whenever the target changes and kept when the same target is re-armed. NPC sidecars
save it with the rest of the aggro; an old sidecar loads with 0, which reads as "longest".

**Retaliation.** After every attack, hit or miss (`combatTurn`, after `attack`), a defender with no
aggro that is up, and that the ruleset lets attack back (`CanAttack(defender, attacker)`), takes an
aggro on its attacker:
- the attacker in the defender's room: a plain attack;
- the attacker in another room (it shot through an exit): an aggro on it. When the shot came from
  a typed `shoot`, the defender wields a shooting weapon and an unlocked exit leads there, the
  aggro holds one shot back (ruling 13); otherwise it waits (below);
- an attacker the defender cannot perceive: the aggro is taken all the same; the turn's existing
  "You can't seem to find your target." applies until it can.

The defender's turn may come later in the same round (turn order is fixed before the round), so a
retaliating NPC can swing back in the round it was first hit. An NPC attacked from another room
still walks toward its attacker and runs an angry command, as today, and swings once it arrives.

**Target lookup each turn.** Where the turn now ends the aggro on a missing target (`:156-160`),
it settles the target instead, in this order:
1. The target no longer exists (dead NPC, a player who left the game): switch (below).
2. The target is not up: switch (below). Today's "Your rage subsides." stays for when there is
   nobody to switch to.
3. The target is in another room, the aggro holds a shot, the body wields a shooting weapon, and
   an unlocked exit leads to the target's room: the turn shoots through that exit, once, and the
   shot is spent (ruling 13). A target the body is about to shoot is reachable, so it is kept even
   when someone in the body's room attacks it (ruling 7). A shot that cannot be fired (no shooting
   weapon, no unlocked exit) is dropped, and the turn goes on to step 4.

   **Shots.** An aggro holds a shot in two cases only: `shoot` arms one, and a body shot by a typed
   `shoot` (hit or miss, as for retaliation) earns one back on its aggro on that shooter, whether
   the shot gave it a new aggro (retaliation) or it was already fighting that shooter. A shot back
   earns nothing, and neither does a shot from anyone the body is not fighting. A body never holds more than one shot. After its
   shot an aggro waits on its target like any other (step 5).
4. The target is in another room, or in this room but not perceived by the body, and someone in
   this room whom the body perceives is attacking it: switch.
5. Otherwise the aggro waits: no swing, and one line the first time the turn finds the target out
   of reach: "Your target is out of reach." (not after a shot: the body just shot it there) When the target and the body share a room again, the
   next turn swings. A wait lasts at most the configured number of rounds (ruling 12, default 5),
   counted in consecutive turns that end here; any turn that finds the target reachable resets the
   count, and so does being attacked by the target (a shooter in the next room that keeps typing
   `shoot` is still in the fight). The turn that fires a shot counts as reachable. When the count runs out the aggro ends, the body is told "You lose track of <name>." and
   observers who perceive it see "<name> mumbles about losing their quarry." (the line
   `loseLostQuarry` used for NPCs). `behaviour.loseLostQuarry` (`behaviour/tick.go:71-84`) and its
   call in `Tick` are deleted: step 1 covers its offline-player case, and this limit covers its
   other-room case for every body, NPC targets included.

A switch inside a turn is followed by the swing at the new opponent in that same turn.

**Walking while an aggro waits** (ruling 8: "you have to move into that room"). `go` refuses a body
in a fight (`usercommands/go.go:358-375`). A body's own aggro now counts as a fight only when its
target is in the body's room; a body that someone in its room is attacking is still in a fight, as
today. So a body whose aggro waits, or that is shooting into another room, can walk; flee stays the
only way out of a fight in the room. An NPC attacked from another room can now take the walk it
queues.

**NPCs attacking players are not PvP.** `dikubasic.CanAttack` applies the PvP rules to any attack
on a player (`ruleset.go:88-113`), so an NPC could never fight back against a player below the PvP
minimum level or outside a PvP room. The PvP checks now apply only when both sides are players. The
recovery room stays safe from every attacker.

**Turn order is fixed by ID.** `fighters()` builds its lists from map iteration
(`NewRound_DoCombat.go:46-58`), which was harmless while defenders stood still; with defenders
swinging back, who acts first changes outcomes from run to run. It now lists players by user ID and
NPCs by NPC ID, then body ID; `dikubasic.TurnOrder` still puts players first.

**Switch.** The body's next opponent is the body in its room, up, perceived by it and one the
ruleset lets it attack (`CanAttack(body, candidate)`), whose aggro targets it with the earliest
`Since`; ties go to turn order. So two players carrying aggros on each other into a room where PvP
is not allowed stop fighting (ruling 10). The aggro moves to that opponent
(`Since` set to now) with a line naming it ("You turn to fight <name>!"). With nobody to switch
to, the aggro ends as today, with today's line for the case.

A target the body chose stays chosen: the switch runs only when the target is gone (steps 1, 2 and 4), so
`attack Z` while others are attacking keeps Z until Z is gone.

**Death.** `death.Die` ends the aggro of every body, player or NPC, whose aggro targets the victim
(today `clearKillersAggro` does it only for credited accounts), and each of them switches at once
by the rule above. A dead player respawned in the recovery room is never waited on.

The round's "either side not up, both aggros end" (`:269-271`) no longer ends the attacker's: the
death that follows switches it, so a killer with other opponents turns to the next one at once.

**What stays as today:** `attack` and `shoot` themselves; party members joining an `attack`; flee
and wimpy; a body with `no-combat` takes no turn (it still gets the aggro); the attacker's own
re-arm after a swing. A shooter's re-arm keeps the exit it shot through, with its shot spent, so the
aggro waits on the target rather than ending with "Your target can't be found." as today.

**Every body with an aggro settles its target every round,** before anything can take its turn
(`no-combat`, an NPC's combat command): the lookup runs, a switch or a lost track happens, the wait
count advances. Only the swing is skipped. Otherwise a `no-combat` body could wait forever without
regenerating (found by plan review round 3).

### Deleted

- `Character.IsDisabled` and its test (the death guard's ban on it goes with it, since nothing named
  `IsDisabled` on a body remains; the item allowlist stays).
- `rooms.FindDowned` and its two reads.
- `actors.BodyReplaced` and its test.
- Offhand breakage: `offhandBreakCheck` and its call, `Item.BreakTest`, `ItemSpec.BreakChance` and
  its clamp, the admin item editor's break-chance field, and the golden fixture's `breakChance`
  parameter. An item file that still has `breakchance:` loads as before (the key is ignored).
- `BackStab`: the `AggroType` value and the `dikubasic` branch; `items.CanBackstab`; the smoke
  client's `*[BACKSTAB]*` strip and its two test rows; the word "backstab" in the admin weapon-rank
  tooltip. The crit line's comment (`combat.go:243-244`) is corrected to say what the code does
  (crits cascade within a round); the behaviour is left for 4.3. `Flee` moves from 3 to 2. A sidecar
  saved mid-flee before this slice holds type 3, which no longer names anything; no shipped data
  holds one (facts: saved aggro), as in slice 1. The plan checks what the round does with an unknown
  type and records it.

### Guards

The death guard's `deathScanned` adds `internal/characters/` and `internal/rooms/`, so no body
health comparison returns there. Its comment changes to match. `character.go` joins the item
`IsDisabled` allowlist (its 16 slot checks are on items). The guard is shown to fail with a planted
`c.Health < 1` in each new package. The widening lands in commit 3, with the floors: `Validate`'s
-10 clamp would fail it in commit 2.

## Testing and evidence

**Commit 1 records a golden before anything changes** (decision 12): a fixture where a player wounds
an NPC, the fight ends (the player flees or the NPC's aggro is cleared), and the regen pass runs a few
times. On today's code the NPC's health stays put.

**Goldens that change, each diff read and summarised in its commit and the PR** as "decided" or
"unexpected" (an unexpected change stops the work):

- floors: `mvm` (`hp=-5/12` becomes the raw value) and any other dump line below 0;
- regen: the commit 1 fixture (the NPC's health rises every heal pass), and any fixture whose NPC
  sits out of a fight under a heal pass;
- breakage: `pvp` loses its two break lines and keeps its shield. Every hit on Bob drew one random
  number for the break test, so from round 2 on the draws shift: the attack lines of rounds 2-4
  differ and Bob's health after round 3 differs (6/15 against 4/15). He still dies in round 4. The
  plan states the full expected diff;
- up to and including the `BackStab` commit, every other combat golden and every reads golden
  byte-identical;
- retaliation (its own commit, last before docs): nearly every combat golden changes, since
  defenders now fight back (Bob in `pvp` trades blows; NPCs swing back in `pvm_death`, `observer`,
  the flee and death fixtures; `shoot_exit`'s target queues its walk toward the shooter (the
  harness runs no queued input; `shot_from_another_room` shows the walk and the swing on arrival);
  `hidden_attacker`'s defender takes an aggro and cannot find its target). The plan states
  each fixture's expected change in kind (who swings, who dies, in which round) before the run, and
  each diff is read against it. Reads goldens that run a fight change the same way.

**New combat golden fixtures for rulings 6-9,** recorded in the retaliation commit:

- two NPCs attack a player in turn; the player kills its target and turns to the one that attacked
  first, with the switch line;
- a player chooses a third target with `attack`; its attackers' deaths do not move it;
- a target flees: with another attacker in the room the body switches; alone, its aggro waits
  ("Your target is out of reach." once), and it swings when the target returns;
- shot from another room: with a shooting weapon the body shoots back through the exit once and
  then waits; a second typed `shoot` earns a second shot back; neither side fires in the rounds
  between; without a shooting weapon it waits; an NPC walks over and swings;
- a hidden attacker: the defender takes an aggro and reads "You can't seem to find your target.";
- a death with three bodies targeting the victim: each switches or ends.

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

**Smoke:** the fight step becomes ruling 11's two fights; bot 1 still observes and its check still
holds. The fighter must be able to choose the level 1 rat over the level 5 one deterministically
(how is the plan's). Every other step is unchanged. Passes with 1, 5 and 12 bots.

## Commits, in order

1. Regen golden on today's code (test files and testdata only).
2. `characters.IsUp` and its wiring; the characters checks through it; `IsDisabled`, `FindDowned`
   and `BodyReplaced` deleted. Goldens unchanged.
3. Floors gone; the death guard widened. `mvm` changes.
4. NPC regen. The commit 1 fixture changes.
5. Offhand breakage gone. `pvp` changes.
6. `BackStab` gone. Goldens unchanged.
6a. Retaliation and target switching (rulings 6-9), with the new fixtures and unit tests (the
    switch's choice by `Since` and turn order; the lookup's five steps; `Since` kept on a re-arm,
    set on a change, saved in a sidecar; ruling 13's shot: armed by `shoot` and by a typed shot,
    not by a shot back, spent by firing, dropped when it cannot fire, saved in a sidecar).
    Combat goldens change; each diff summarised.
7. Docs: patch notes, status, the phase design's slice 4 row (cut into 4.1-4.3) and its "Notes for
   later slices" (this slice's items leave; 4.2's and 4.3's lists are written there),
   `docs/known-issues.md` if it names any of these.

## What this locks in

- **No engine code compares a body's health with a number,** `characters` and `rooms` included
  (guarded), except `ApplyHealthChange`'s `newHealth < 0` test for the `cancel-on-combat` cancel,
  kept on purpose until slice 6 (a local variable, which the guard does not see). The ruleset's `Condition` is the only line between up, downed and dead.
- **`characters` asks the ruleset through one func variable.** Any later body rule `characters`
  needs uses the same route rather than a second one.
- **Health below 0 is the ruleset's business.** The engine applies damage as given.
- **Being attacked puts every body in the fight** (rulings 6-9). NPCs fight back. NPCs still
  never start a fight, for now: NPC aggression is out of scope (phase design) and comes with a
  ruleset or a later phase.
- **An aggro remembers when it began,** and the next opponent is the longest attacker.
- **Shooting is one shot per action** (ruling 13): no body fires twice without a typed `shoot`
  behind each shot, its own or its target's.

## Out of scope

- 4.2 and 4.3 (ruling 1).
- Weapon delay (`characters.SetAggro`) behind the ruleset: waits for swing speed (slice 2 notes).
- The `cancel-on-combat` cancel only firing below 0 health: unchanged (it is a buff rule; slice 6).
- `dikubasic/simulate.go`'s own health checks: ruleset code, outside the guard.
- The crit cascade (facts): 4.3 replaces the hit roll; its spec decides crits. Recorded in the phase
  design's notes for 4.3.
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
| The death guard widens in commit 3, not commit 2 (settled while planning) | `Validate`'s -10 clamp, removed in commit 3, fails the widened guard | Remove the clamp in commit 2 instead |
| `character.go` joins the guard's item `IsDisabled` allowlist (settled while planning) | Its 16 calls are item slot checks, which the allowlist exists for | none |
| The `pvp` golden's shifted draws are accepted as decided (settled while planning) | They follow from removing the break test's draw; the fight's outcome (Bob dies in round 4, keeps his shield) holds | Keep a dummy draw to hold the goldens (dead code for parity's sake) |
| The backstab leftovers outside `aggro.go` and `combat.go` go too (settled while planning) | Ruling 3 deletes backstab; each is dead or names it | Leave them for slice 7's docs sweep |
| The crit cascade is documented now and fixed in 4.3, not here (settled while planning) | 4.3 replaces the hit roll; a fix now changes goldens that 4.3 changes again | Fix it here, in its own commit with its golden diff |
| Retaliation is engine code (round and death), not a ruleset method | Who fights whom is the engine's (phase design); only "may it attack back" is the ruleset's, through the existing `CanAttack` | Add a `Retaliates` ruleset method so a ruleset can opt out |
| Retaliation also follows a miss | "Being attacked" (ruling 6); every swing earns credit, so a miss is an attack | Only a hit |
| "Longest attacker" is kept as the round the aggro began (`Since`) on the attacker's aggro | The attacker's aggro is the only record of who attacks whom; a list on the defender would be a second record to keep in step | A per-defender attacker list |
| Ties in `Since` go to turn order | Deterministic, and turn order is the ruleset's | Random |
| Switch candidates must be in the body's room, up and perceived | Ruling 7 says "in your room"; a body cannot turn to fight what it cannot perceive (the visibility rule) | Include remote shooters |
| Switch line "You turn to fight <name>!", the waiting line "Your target is out of reach." (told to the body only) | New lines the rulings imply; wording is mine. "No longer here" read oddly for a shooter that was never in the room (plan Q13) | Reword |
| The PvP checks apply only when both sides are players; the recovery room protects from everyone (plan Q7) | Without it, an NPC cannot fight back against a player below PvP level 15 or outside a PvP room under the shipped `limited` config, which defeats ruling 6 | Keep PvP rules for NPC attackers (NPCs then mostly cannot fight back) |
| A body's own aggro blocks `go` only when its target is in its room (plan Q8) | Ruling 8 says you move into the attacker's room; a waiting aggro must not freeze a body. Being attacked in the room still blocks `go` | Keep the block; a waiting body could only flee |
| Turn order is sorted by ID within players and within NPCs (plan Q9) | Map order made outcomes vary between runs once defenders swing back; lower user IDs act first among players | Random order each round, on purpose, with goldens that pin it |
| A switch inside a turn swings at the new opponent that turn (plan Q11), except that a weapon's own delay still applies: a switch re-arms through `SetAggro` like `attack <name>`, so a slow weapon waits its rounds first (plan review round 3) | A turn lost to switching would punish the body for its target's death; weapon delay is the weapon's rule, the same for any new target | Spend the turn switching; or skip the delay on a switch |
| Switch candidates must be ones the ruleset lets the body attack (plan review round 3) | Without it a switch could fight a player the PvP rules protect (ruling 10) | none |
| Each turn re-checks `CanAttack` on the current target, in the room or through an exit; when refused, the body switches (to a candidate the ruleset allows) or the aggro ends with the ruleset's refusal text ("This is not a PVP area.") (plan review round 2; recorded here in round 4) | Ruling 10 has to hold for a fight already going: two players fighting who walk into a room without PvP, or an admin who turns PvP off, must stop, as `attack` would refuse them | Check only when the aggro is taken; such fights continue |
| A turn that finds its target in reach but cannot perceive it ("You can't seem to find your target.") counts toward the wait limit (plan review round 4) | Ruling 12 is for every body; otherwise a body that retaliated against a hidden attacker waits forever without regen. A hidden attacker that keeps hitting resets the count each time it attacks (round 6 row below), so the body loses track only once the attacks stop | Reset the count whenever the target is in reach |
| The recovery room is safe in both directions: `CanAttack` also refuses an attacker standing in it (plan review round 4) | Ruling 10 calls it safe from everyone; otherwise a player there could shoot through its exit at an NPC that could never fight back (ruling 6) | Refuse only targets in it, as today |
| `Aggro.OutOfReach` is saved in NPC sidecars; old sidecars load 0 | An NPC restored mid-wait resumes its count rather than getting a fresh limit | Do not save it |
| "Your target is out of reach." is told once per wait: the aggro carries a told mark, cleared only when the target is found reachable and perceived, not by being attacked; saved in NPC sidecars with the count (plan review round 6) | Keeps step 5's "one line the first time" true under steady fire, where the count keeps resetting | Tell it each time the count starts from 0 |
| Being attacked by its target resets a body's wait count (plan review round 6) | Without it a body under steady fire from the next room loses track of its shooter every 5 rounds and its room reads the mumble line each time, then retaliation re-arms it; ruling 12 is about a target that has gone, and one still shooting has not. The same holds for a hidden attacker that keeps hitting, so that row's lose-and-re-arm cycle no longer happens | Accept the cycle |
| A target in the body's room that the body cannot perceive is given up for a perceived attacker in the room (lookup step 4) (plan review round 5) | Ruling 7: a chosen target is kept while reachable, and a target you cannot find is not one you can fight; otherwise a body swings at nothing for up to the wait limit while a visible attacker hits it | Keep the unseen target until the limit runs out |
| Shooting looks exits up the way `shoot` does (`room.FindExitByName` and the room's exit info, so temporary and mutator exits and mutator locks count) (plan review round 5) | "`shoot` stays as today": a shot through a temporary exit worked before this slice | none |
| A target that leaves the game ends the aggro with today's "Your target can't be found." (or a switch), not the lose-track line; the observer line is a plain room line without the old emote colour (plan review round 5) | `loseLostQuarry`'s offline case emoted; step 1 is the same rule for every body | Emit the lose-track lines on step 1 too |
| The smoke world copy also removes End of the Line's three rats and gives the two fight rats a 24-hour respawn (plan review round 5) | Ruling 11 needs exactly those two rats to fight; a wandering third or a respawn mid-run would change which rat bot 0 meets | Leave the shipped rats; filter by name only |
| When shooting back, the exit is the aggro's own exit if it still leads to the target, otherwise the first exit by name that does | Deterministic; most rooms have one exit to a given room | Another order |
| Retaliation records the retaliator's damage like any attack, so a player who fought back against a player has that player in `PlayerDamage`, and a later death can be booked as PvP (`death.go:136-137`) | Every swing earns credit (same as a miss counting as an attack); the kill-credit rules are 4.2's (progression), noted there | Exclude retaliation from `PlayerDamage` here |
| Every body with an aggro settles its target each round even when its turn is taken (`no-combat`, an NPC combat command) (plan review round 3) | Ruling 12's limit is for every body; otherwise those bodies wait forever | Count only turns that could swing |
| A target the body is about to shoot (its aggro holds a shot) counts as reachable and is kept (lookup step 3 before step 4; plan Q12; narrowed by ruling 13, 2026-11-14). Once the shot is spent, a perceived attacker in the body's room takes over under step 4 | Ruling 7: a chosen target is never replaced while reachable. A target the body will not fire at again is out of reach, and waiting on it while someone hits you in your room is the thing ruling 7 forbids | Keep the remote target until the wait limit runs out, while the attacker in the room hits for free |
| How a shot is held: one mark on the aggro, set by `shoot` and by being shot by a typed `shoot` from the body's own target or (retaliation) from a new one; spent by firing, dropped when it cannot be fired; never more than one. The mark says which it is (typed or earned), and only a fired typed shot earns a shot back. An NPC running `shoot` as a combat command types it like a player (ruling 13) | "Each typed `shoot` earns one shot back"; a mark on the shot (typed or not) stops two shooters answering each other every round, which is continued fire again. One shot at most: five unanswered typed shots in one round still earn one, since the body fires once per turn | A count of owed shots |
| A `no-combat` body holding a shot keeps it, and its wait count stays at 0, until the buff ends and the shot fires (found by the whole-branch review, 2026-11-14) | Lookup step 3 finds the target reachable (the body would shoot it); only the shot is skipped, as every swing is under `no-combat`. The wait lasts as long as the buff that carries the flag; no shipped buff carries it (`buffs-flags/no-combat.yaml` defines the flag only) | Count the held shot's turns toward the wait limit |
| A held shot is spent by any attack the turn makes, including a swing when the target has come into the body's room (settled while planning, 2026-11-14) | One attack per action: the shot was the body's next attack, and the swing took its place. The plain re-arm after a swing makes a new record, as today | Carry the shot through a swing, like `Since`, so the body fires it if the target steps out again |
| A shot earned by a body with no shooting weapon is recorded and dropped on its next turn (step 3), not refused when earned (settled while planning, 2026-11-14) | One place decides whether a shot can fire; the body may wield a bow by its next turn | Refuse it when earned (same result unless the body arms between) |
| The shot mark is saved in NPC sidecars; old sidecars load with none | Sidecars save the rest of the aggro; an NPC restored with a shot owed fires it | Do not save it |
| A hit from a typed `shoot` by someone the body is not fighting earns no shot (ruling 6: a body already fighting keeps its target) | One target at a time; switching to every new shooter is what ruling 7 rules out | Earn a shot at the new shooter and switch |
| "Your target is out of reach." is not told on the turns after a shot (the shot sets the told mark) | The body just shot the target where it stands; telling it the target is out of reach reads wrong | Tell it as for any wait |
| A body attacked from another room by a typed shot, with no shooting weapon, waits (ruling 9); an NPC also takes today's walk | Unchanged by ruling 13 | none |
| "In a fight" (`inFight`, which gates `go`, `quit` and `suicide`) means someone in the body's room fights it, or it fights someone in its room (plan Q17) | One meaning for all three commands; a body shot from the next room could already quit today; flee stays the only way out of a fight in the room. A pending flee with nobody attacking is no longer a fight | Narrow the change to `go` only |
| A failed flee keeps the `Since` the fleer's aggro had (plan Q15) | The fight resumes with the same opponent; it should not lose its place | Reset it |
| The switch rule lives in `internal/death` (plan Q14) | The round imports `death`, not the reverse; death needs the rule | A new package both import |
| The `death_npc` fixture gives Alice 30 health (plan Q10) | With the rat fighting back it killed her first, and the fixture exists to record an NPC death | Add a second fixture instead |
| The wait count is consecutive turns at lookup step 5, reset by any turn that finds the target reachable (shooting counts as reachable) | Ruling 12 says "rounds out of reach"; a target that steps back in and out should not run down a count it earned while away | Count total rounds since the target was last in the room |
| The wait limit sits in the engine round, not in NPC behaviour, and `loseLostQuarry` is deleted | One rule for every body (ruling 12, one-actor ruling); the combat golden harness does not run `ActorTick`, so a behaviour-side rule would also be invisible to the goldens | Keep `loseLostQuarry` for NPCs alongside the new limit |
| An NPC waiting on an NPC target in another room gets the same limit | `loseLostQuarry` skipped NPC targets, so such an aggro never ended; one rule covers it | none |
| Lose-track lines: "You lose track of <name>." to the body; "<name> mumbles about losing their quarry." to observers, for players and NPCs alike | The observer line is the one NPCs already used; the body's line is new wording, mine | Reword, or keep the observer line for NPCs only |
| A limit of 0 or less in config means the default, 5 | A wait that ends before it starts would make every remote fight end at once (option 3, not chosen) | Treat 0 as "no limit" |
| A body whose aggro waits does not regenerate | It is still in that fight; one gate (own aggro) | Let a waiting body regenerate |
| An NPC attacked from another room keeps today's walk toward the attacker and angry command | Existing NPC behaviour; with retaliation it now swings on arrival | Have NPCs wait like players |
| The round no longer ends the killer's aggro when the defender goes down; death switches it | One place decides what follows a target's loss | Keep the round's end and let the killer re-acquire on the next hit |
| `Since` is saved in NPC sidecars; an old sidecar's 0 counts as longest | Sidecars already save the aggro | Do not save it (a reloaded NPC's fights restart their order) |
| Sub-slice names are 4.1, 4.2, 4.3 and the branch `strip/phase-5-4-1-leftovers` | Avoids "4b", already the name of the NPCs-are-players phase | Rename |
