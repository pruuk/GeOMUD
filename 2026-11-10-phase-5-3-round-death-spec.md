# Phase 5 slice 3: round, regen and death through the seam (spec)

Date: 2026-11-10. Written at `master` `a06e0f07` (slice 2 merged as #31). Branch
`strip/phase-5-3-round-death`. Builds on `2026-11-06-phase-5-rules-seam-design.md` (slice 3 row,
"Notes for later slices"), `gomud_strip_plan.md` decisions 1 and 5-11, and slice 2's spec
(`docs/completed/2026-11-10-phase-5-2-seam-fight-spec.md`). Not yet reviewed by Calabe.

**Goal.** The engine asks the ruleset whether an actor is up, downed or dead, how much it
regenerates, and what a death means; every actor dies through one engine routine. Round and regen
move with parity. Death goes straight to the bare rules (decisions 7 and 11) instead of moving
GoMud's death behind the seam first.

## Rulings this spec implements

Earlier, still binding: decision 7 (die at 0, no downed state; the engine asks the ruleset whether an
actor is dead so a later ruleset can add one), decision 11 (a death drops carried items and gold
into a corpse; worn gear stays; no protection band, XP penalty, drop chances or permadeath; NPCs
drop their loot), one actor model, rulesets return messages and never send, no escaping a fight
except by fleeing.

Calabe, 2026-11-10, in brainstorming:

1. **Death goes straight to the bare rules (option B).** Round, regen and the dead-or-alive check
   move with parity, because they survive the shrink. Death does not: moving GoMud's death
   (permadeath, lives, XP penalty, drop chances, bleed-out) behind the seam only for slice 4 to
   delete it is throwaway work. A golden that runs today's death is recorded first, so every change
   is visible. This narrows decision 1; recorded as strip plan decision 12 (below).
2. **A player respawns at 1 health, and regen in the recovery room is fast.** "Fast" is upstream's
   deleted Death Recovery buff (`git show 57255ee4^:_datafiles/world/empty/buffs/24-death_recovery.js`):
   1d10 health and 1d10 mana every round. It becomes a reference-ruleset regen rule, not a rebuilt
   buff, which takes "rebuild Death Recovery" off slice 6.
3. **An NPC drops its carried items and gold always, and each worn item at its own
   `ItemDropChance`, as today (option B).** This keeps one drop chance in the bare ruleset, by
   choice: builders keep a way to make rare drops.
4. **`suicide` stays a player command, refused in a fight and in the recovery room (option A).**

## Facts verified against source (2026-11-10, `a06e0f07`)

Read or grepped by the controller today.

| Fact | Value | Where |
|---|---|---|
| Death code | `usercommands/suicide.go` 313 lines (player death; branches to `npcDeath` at :29-31); `suicide_npc.go` 371 lines | `wc -l`, read |
| Health thresholds in engine code | `NewRound_DoCombat.go:169` (`< 1`, rage subsides), `:217` (`>= 1` then `< 1`, `KilledBy`), `:269` (`<= 0`, end aggro), `:494` (`<= -10`, queue `suicide`), `:498` (`< 1`, `PlayerDrop`), `:513` (NPC `< 1`, queue `suicide`); `NewRound_ActorUpkeep.go:62` (NPC `<= 0`, queue `suicide`); `Buff_ApplyBuffs.go:64` (NPC `<= 0`, queue `suicide` input); `NewRound_AutoHeal.go:44,46` (`< 1` bleed, `<= -10` suicide) | `grep -rn 'Health <= -10\|Health < 1\|Health <= 0'` |
| Thresholds outside engine packages | `characters/character.go:423` (`downed` adjective), `:471` (`downed` name suffix), `:710` (`forgetKillerIfUp`), `:819` (`IsDisabled`); `rooms/rooms.go:1144,1219` (`FindDowned`); `rules/dikubasic/simulate.go:138,157` | same grep |
| Import cycle | `characters` cannot import `rules` (`rules` → `actors` → `characters`); slice 2 spec facts | slice 2 |
| Damage floors | `ApplyHealthChange` (`character.go:895-916`): from above 0 a hit stops at -5; otherwise floor -10 | read |
| AutoHeal | `NewRound_AutoHeal.go:17-89`: every 3rd round (:22); accounts only (`GetOnlineUserIds`); skips anyone with an aggro (:33-35) and the recovery room by raw ID (:26, :37-39); bleed 1 a round with "you are bleeding out!" / "Somebody needs to provide aid!" (:44-64); else `Heal(HealthPerRound(), ManaPerRound())` (:68-73); vitals and prompt events on change (:77-84) | read |
| Regen amounts | `HealthPerRound` = `1 + StatMod(HealthRecovery)` (`character.go:1005`); `ManaPerRound` = `1 + StatMod(ManaRecovery)` (:1016) | read |
| `Heal` bug | `character.go:988-1000`: `c.Mana += hp` (the mana argument is unused); one caller, `NewRound_AutoHeal.go:69` | read, grep `\.Heal(` |
| Recovery room is a private copy | `MoveToRoom` to `DeathRecoveryRoom` creates an ephemeral copy (`roommanager.go:300-307`); `rooms.GetOriginalRoom` (`ephemeral.go:284`) maps a copy back; `IsEphemeralRoomId` (:212). Raw-ID compares that miss the copy: `NewRound_AutoHeal.go:37`, `rooms.go:2063` (`CanPvp`), `dikubasic/ruleset.go:93` (`CanAttack`) | read, grep `DeathRecoveryRoom` |
| Zones | world zones are `shadow_realm` and `startland`; no `Training` zone ships | `ls _datafiles/world/empty/rooms` |
| Hard-coded zone names | `suicide.go:41` (`Shadow Realm`, "You're already dead!"); `suicide_npc.go:112,203` (`Training`, skip kill tracking) | grep |
| Player death today | room "has died." (:47-53); KD PvP or mob death (:56-62); per-killer KD and "killed by" list (:64-86); world notice "*** X has DIED! ..." (:88-95); protection band `Level > ProtectionLevels` (:97); elite death and killer clear (:99-109); `PlayerDeath` event with `Permanent` (:111-119); permadeath and lives (:122-170); unless `perma-gear`: worn items at `EquipmentDropChance`, gold, backpack (`AlwaysDropBackpack` or chance), XP penalty, corpse (:177-286); protected: corpse only (:288-296); cancel all buffs (:300); revive `Health = 1`, `Mana = 0`, vitals event (:302-306); clear `PlayerDamage` (:308); move to recovery room (:310) | read |
| NPC death today | `revive-on-death`: full health, "is suddenly revived in a shower of sparks!", cancel flag (`suicide_npc.go:31-46`); "has died." (:50-57); XP pool `XPTL(level-1)/90`, `MobDeath` event, per-credited-account XP with level scaler, elite bonus, alignment, aggro clear, party XP (:59-235); unless `perma-gear`: carried items always, worn at `ItemDropChance` (`Rand(100)` per item, :277-279), gold; `MobItemDrop` events (:257, :293); floor or corpse by `CorpseItems` (:240-345); `NPCEnded`, `DestroyInstance`, leave room, corpse (:318-366) | read |
| Death config | `Death.EquipmentDropChance` 0.25, `AlwaysDropBackpack` false, `XPPenalty` none, `ProtectionLevels` 5, `PermaDeath` false, `CorpsesEnabled` true, `CorpseDecayTime` 1 hour, `CorpseItems` false; `LivesStart` 3, `LivesMax` 3, `LivesOnLevelUp` 1 | `_datafiles/config.yaml:196-237`; `config.gameplay.go:17-19, 77-84, 94-150` |
| Lives and permadeath readers | `ExtraLives` field `character.go:80`; set `userrecord.go:69-70` and `start.go:182`; level-up grant and cap `userrecord.go:204-249` (`LevelUp.LivesGained`); template func `permadeath` (`templatesfunctions.go:193-195`), template `character/permadeath.template`; admin pages `config-wizard.html:1111-1126,1195-1198,1546-1559`, `config.html:232-238` | grep |
| `CorpseItems` readers | `get.go:90-91` (loot from corpse), `look.go:412`, `look.panels.go:110` (corpse contents panel), `rooms/corpse.go:15-16` (comment) | grep |
| `MobItemDrop` | emitted only by `suicide_npc.go`; no listener | grep |
| `suicide` command | registry `usercommands.go:188` (`AllowedWhenDowned: true`), with a comment that it is the death path until Phase 5 (:186-187) | read |
| Fight check | `inFight(actor, room)` (`go.go:353-370`): own aggro, or anyone in the room targeting the actor; `quit` refuses with "You're too busy to quit right now!" (`quit.go:19`) | read |
| Downed handling | `PlayerDrop` event (`eventtypes.go:213`), queued only at `NewRound_DoCombat.go:500`, handled by `PlayerDrop_HandlePlayerDrop.go` ("you drop to the ground!" / "{actor} drops to the ground!"); commands refused when downed unless `AllowedWhenDowned` (`usercommands.go:311`) | grep, read |
| Upkeep | `NewRound_ActorUpkeep.go`: alignment decay for accounts (:40, `decayAlignment` :69-91); buff triggers; `Validate`; NPC `<= 0` queues `suicide` (:62) | read |
| Smoke death step | `cmd/smoke/script.go:220-228`: `suicide`, "has DIED!", look shows the recovery room, `portal` to start, still holds `smokeItem` ("Bots are level 1, under Death.ProtectionLevels, so nothing is dropped") | read |
| Reads golden H | `reads_golden_test.go:830-`: rat `suicide` after damage (alignment lines), corpse bury and crumble, alignment decay, `GrantXP`, Bob downed then `suicide` with the recovery room set to room B | read |
| Combat golden | records queued `events.Input` (e.g. `Input rat: "suicide"`), never runs it; 11 tests | slice 2 spec; `grep -c 'func Test'` |
| Free package name | `internal/death` does not exist | `ls` |
| `parties` imports | only `util` | `go list` |

## Design

### Ruleset methods added

```go
type Condition int

const (
	Up Condition = iota
	Downed
	Dead
)

type RegenResult struct {
	Health, Mana int
	Messages     []perception.Message // sent by the engine
}

type DeathOutcome struct {
	Revive        bool                  // the victim gets up instead of dying
	ReviveHealth  int                   // health on revive
	Respawn       bool                  // true: back to life in the recovery room; false: the body ends
	RespawnHealth int                   // health on respawn
	DropItems     []items.Item          // moved into the corpse (or onto the floor)
	DropGold      int
	Messages      []perception.Message  // award lines (XP, alignment), sent by the engine
}

// on Ruleset:
Condition(a *actors.Actor) Condition
Regen(a *actors.Actor, round uint64) RegenResult
OnDeath(victim *actors.Actor) DeathOutcome
```

The exact field set is the plan's to fix against the code; the shape above is the design.

**`dikubasic`:**

- `Condition`: `Dead` at health 0 or below, else `Up`. Never `Downed`.
- `Regen`:
  - **In the recovery room** (the room or its private copy, through `GetOriginalRoom`): 1d10
    health and 1d10 mana every round, with upstream's lines ("The shadow realm heals you for N
    damage!" and the mana variants; the room reads "{actor} is recovering from a recent death.").
  - **Otherwise, parity:** accounts only, every 3rd round, nothing with an aggro, `HealthPerRound`
    and `ManaPerRound`.
- `OnDeath`:
  - **Revive:** a victim with `revive-on-death` revives at full health (today's NPC rule, now for
    anyone).
  - **Awards:** for an NPC victim, today's XP, alignment and party XP for the credited accounts,
    moved as is (levels exist until slice 4), applied to the killers' characters, lines returned.
  - **Respawn or end:** `Respawn` and `RespawnHealth` 1 for an account; `Respawn` false for an
    NPC.
  - **Drops:** for a player, carried items and gold; worn gear stays (decision 11). For an NPC,
    carried items and gold always, each worn item at `ItemDropChance` (ruling 3). Nothing when
    the victim has `perma-gear` (an engine flag the ruleset reads).

### What stays engine

`internal/death` (new; imported by `hooks` and `usercommands`, imports neither): `death.Die(actor)`
for every actor:

1. `outcome := rules.Get().OnDeath(victim)`.
2. **Revive:** set health, cancel `revive-on-death`, room "{actor} is suddenly revived in a shower
   of sparks!", stop.
3. **Announce:** room "{actor} has died."; for an account, the world notice "*** X has DIED!
   (killed by ...)" and the event log; KD bookkeeping (PvP or mob death, per-killer kills, elite
   kills and deaths; the `Training` exemption goes); `PlayerDeath` (without `Permanent`) or
   `MobDeath`; send `outcome.Messages`.
4. **Clear aggro** of every credited account's body targeting the victim (today NPC deaths only).
5. **Drops:** `outcome.DropItems` and `DropGold` leave the body. If corpses are enabled they go
   into the corpse; if not, onto the floor. `MobItemDrop` is still queued per NPC item.
6. **Ending:**
   - **Respawn:** cancel all buffs, health `RespawnHealth`, clear `PlayerDamage` and the killer
     record, move to the recovery room. The corpse holds the drops.
   - **End:** `NPCEnded`, `DestroyInstance`, leave the room, corpse.

**Death checks call `Die` directly** where `Condition` returns `Dead`: the end of a combat round
(`handleAffected`), round upkeep, an immediate buff trigger. Nothing queues `suicide` as input.

**Downed stays generic engine handling** for a ruleset that returns it: the actor cannot act
(commands not `AllowedWhenDowned` are refused), the room reads "drops to the ground" once on
becoming downed, nobody dies of it. `dikubasic` never returns it, so only a stub-ruleset test
reaches it. Bleeding, if a ruleset wants it, is a negative `Regen`.

**The round's other health checks go through `Condition`:** "Your rage subsides" when the target is
not `Up`; `KilledBy` recorded when the defender leaves `Up`; aggro ends when either side is not
`Up`.

**AutoHeal becomes the engine's regen pass:** every actor, every round, `Regen` decides; the engine
applies `Health`/`Mana` through `Heal`, sends the lines, and raises the vitals and prompt events on
change. The bleed-out and its "Somebody needs to provide aid!" line go.

**One recovery-room test:** an engine helper `rooms.IsDeathRecoveryRoom(roomId)` (through
`GetOriginalRoom`) replaces the raw-ID compares in regen, `suicide`, `CanPvp` and
`dikubasic.CanAttack`.

**`suicide` command:**

- Refuses in a fight (`inFight`): "You're too busy to die right now!"
- Refuses in the recovery room: "You're already dead!" (today's text, now recognising the copy).
- Otherwise calls `death.Die`.
- It stays `AllowedWhenDowned`, so a downed actor under a downed-state ruleset can still end it.

**`Heal` is fixed** (`c.Mana += mana`) in its own commit.

### Deleted

- **Player-only death rules:** permadeath, `ExtraLives` and its level-up grant and cap, the lives
  config (`LivesStart`, `LivesMax`, `LivesOnLevelUp`), the `permadeath` template function and
  template, `PlayerDeath.Permanent`, `LevelUp.LivesGained`; `ProtectionLevels`, `XPPenalty`,
  `EquipmentDropChance`, `AlwaysDropBackpack`; `CorpseItems` (drops always go into a corpse when
  corpses are on, so its readers in `get`, `look` and the corpse panel lose the check).
- **The old death code:** `npcDeath` and the body of `Suicide`; the bleed-out; the -10 branch in
  `handleAffected`.
- **Hard-coded zone names:** `Shadow Realm` and `Training`.
- **Admin config pages:** the deleted knobs' entries.
- **The import guard's allowlist** loses `usercommands/suicide_npc.go` (deleted; `AlignmentChange`
  is used only inside `dikubasic`).

### Guards

- **New AST guard `TestEngineAsksTheRulesetWhoIsDead`:** in `internal/hooks`, `internal/usercommands`
  and `internal/death`, no comparison of a `.Health` selector with a literal. `characters`, `rooms`
  and `dikubasic` are outside the scanned set. Positive control: a planted comparison fails it.
- **No engine code builds an `events.Input` whose text is `suicide`**, nor calls
  `.Command("suicide")`.

### Strip plan decision 12 (added in the docs commit)

"Move with parity, then shrink" (decision 1) covers code that survives the shrink. Code the shrink
deletes goes straight to its bare form, with a golden recorded first so the change is visible
(Calabe, 2026-11-10, slice 3).

## Testing and evidence

**Commit 1 records today's death before anything moves.** The combat golden harness gains a mode
that runs queued inputs. New fixtures:

- a player kills an NPC that carries and wears items and gold: XP, alignment, drops, corpse;
- an NPC kills a player above the protection band: drops by chance, corpse, recovery room;
- a PvP death.

Recorded once on today's code.

**Parity commits** (`Condition` and `Regen` for the round and upkeep; the recovery-room helper):
all 9 combat goldens, the new death fixtures and every reads golden byte-identical. The bleed-out
goes in the death commit, not here.

**The death commit changes goldens on purpose.** Each diff is read and summarised in the commit and
the PR as "decided" (decisions 7 and 11, rulings 1-4) or "unexpected"; an unexpected change stops
the work. Expected:

- `Input <name>: "suicide"` lines become the death's own record;
- the death fixtures lose the protection band, chance drops and XP penalty, and gain corpse
  contents;
- reads golden H changes in its NPC-death and player-death sections.

**Stub rulesets, no engine edit:**

- `Condition` returning `Downed`: the actor cannot act, the room reads "drops to the ground", nobody
  dies;
- `OnDeath` returning `Revive`: the victim gets up;
- `OnDeath` with fixed drops: exactly those land in the corpse;
- `Regen` returning +5: the engine applies 5.

**`dikubasic` unit tests:**

- `Condition` at 1, 0, -1.
- `Regen`:
  - every 3rd round only; never with an aggro; accounts only (parity);
  - recovery room: every round, the room and its copy, 1d10 each drawn under a seed and matched
    draw for draw.
- `OnDeath`:
  - player drops carried and gold, keeps worn;
  - NPC worn items follow `ItemDropChance` (seeded);
  - `perma-gear` drops nothing;
  - `revive-on-death` revives an account and an NPC.

**Command and engine tests:**

- `suicide`: refused in a fight, refused in the recovery room and its copy, dies otherwise.
- `Heal`: mana regen uses the mana amount.
- `IsDeathRecoveryRoom`: the room, a copy, another room.

**Guards:** the health-comparison guard and the no-`suicide`-input guard each shown to fail with a
planted violation.

**Every new test shown to fail under a prescribed mutation,** as in slice 2.

**Smoke:** the death step still passes. Its comment changes, since bots now drop carried items into
the corpse. The smoke item is carried, so the "still holds the item" check changes to finding it in
the corpse; the plan fixes the exact steps against `script.go`.

## Commits, in order

1. Death goldens on today's code (test files and testdata only).
2. `rooms.IsDeathRecoveryRoom` and its uses in `CanPvp` and `CanAttack`. Goldens unchanged.
3. `Condition` and `Regen` behind the seam with parity (round checks, upkeep, buff trigger, AutoHeal
   less the bleed-out). Goldens unchanged.
4. The `Heal` fix.
5. Recovery-room fast regen.
6. One death routine: `internal/death`, `OnDeath`, XP and alignment into `dikubasic`; the deletions
   above; bleed-out and the -10 branch. Goldens change; each diff summarised.
7. `suicide` refuses in a fight and in the recovery room.
8. Guards.
9. Stub-ruleset tests.
10. Docs: patch notes, status, design notes, strip plan decision 12, `geomud-verifying` count.

## What this locks in

- **One death path.** Every actor dies through `death.Die`; what a death means is `OnDeath`. A new
  kind of death is a ruleset change, never a second path.
- **The ruleset says who is dead.** The engine compares no health against a number (guarded).
- **A ruleset can add a downed state with no engine edit,** shown only by a stub test until one does.
- **Drops go into a corpse whenever corpses are on,** for players and NPCs alike.
- **Decision 12.**

## Out of scope

- Slice 4:
  - NPC regen (the one-actor ruling; slice 3 keeps accounts-only regen for parity);
  - the -5/-10 floors in `ApplyHealthChange`;
  - XP, levels and alignment (including decay, which stays in upkeep);
  - the four `characters` health checks and `FindDowned` (the import cycle; recorded in the
    design's notes);
  - offhand breakage and `BackStab`.
- An `aid` command.
- Corpse looting rules (who may loot a player's corpse): unchanged.
- `RoundUpkeep` and `ClampDamage`: no rule needs them yet.

## Decisions taken without asking

| Decision | Why | If wrong |
|---|---|---|
| Drops go **inside** the corpse whenever corpses are on, for players and NPCs; `CorpseItems` is removed. Today's shipped `false` drops loot on the floor beside an empty corpse | Decision 11 says "into a corpse"; one rule for every actor | Keep floor drops for NPCs only: a split death path for one kind |
| New engine package `internal/death` | Both `hooks` and `usercommands` call it; `usercommands` cannot import `hooks` | Put it in `usercommands` and have `hooks` call through it |
| `Regen` and `OnDeath` return lines in their result; the engine sends them | Rulesets never send (slice 2) | none |
| `ClampDamage` and `RoundUpkeep` not added | Health below 0 after the killing blow is read by nothing (death follows at once, respawn sets 1); alignment decay is slice 4's to delete | Add them when a rule needs one |
| XP, alignment and party XP move into `dikubasic`'s `OnDeath` as they are | Levels exist until slice 4; the slice 4 shrink deletes them there | none |
| Recovery-room lines reuse upstream's buff text ("The shadow realm heals you for N damage!") | Prior art; the shipped recovery zone is `shadow_realm` | Reword |
| `suicide` in a fight reads "You're too busy to die right now!" | Matches `quit`'s refusal | Reword |
| The recovery-room compare moves to one helper used by regen, `suicide`, `CanPvp` and `CanAttack` | The raw-ID compares miss the private copy every respawn creates | none |
| Downed stays generic engine handling, reachable only by a stub ruleset | Decision 7 asks for the seam, not a downed `dikubasic` | Delete the downed handling; a later ruleset rebuilds it |
| `PlayerDeath.Permanent` and `LevelUp.LivesGained` fields go with permadeath | No permadeath | Keep the fields, always false or zero |
| The admin config pages lose the deleted knobs' entries | They would edit nothing | none |
