# Phase 5 slice 3: round, regen and death through the seam (spec)

Date: 2026-11-10. Written at `master` `a06e0f07` (slice 2 merged as #31). Branch
`strip/phase-5-3-round-death`. Builds on `2026-11-06-phase-5-rules-seam-design.md` (slice 3 row,
"Notes for later slices"), `gomud_strip_plan.md` decisions 1 and 5-11, and slice 2's spec
(`docs/completed/2026-11-10-phase-5-2-seam-fight-spec.md`). **Approved by Calabe 2026-11-10**, with
ruling 5 (help files) added at review.

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
5. **Help files say that a corpse's contents are destroyed when it rots** (Calabe, at spec review).
   Verified: `Room.UpdateCorpses` (`rooms.go:231-258`) removes a decayed corpse with its `Items`
   and `Gold`, and nothing moves them out first.

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
| Help pages on death | `templates/help/death.template` (bleed out, die at -10, gear and XP loss, shadow realm portal) and `templates/help/health.template` (0 HP bleed, -10 death, "all gear you posess drops to the ground"); keyword `death` (`keywords.yaml:18`); no help page mentions corpses | `grep -rli 'corpse\|death' templates/help keywords.yaml`, read |
| Corpse decay | `UpdateCorpses` (`rooms.go:231-258`) prunes a corpse once `CorpseDecayTime` (game time, shipped `1 hour`) has passed, announcing "A <name> corpse crumbles to dust."; its items and gold go with it | read |
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
   (killed by ...)", whose killer list each reader gets cut to the killers it perceives (none left:
   no "(killed by ...)" part; plan review round 5), and the event log; KD bookkeeping (PvP or mob death, per-killer kills, elite
   kills and deaths; the `Training` exemption goes); `PlayerDeath` (without `Permanent`) or
   `MobDeath`; carry out `outcome.Awards`, each with its lines (Q1: no separate `Messages`).
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
- It stays `AllowedWhenDowned`, so a downed actor under a downed-state ruleset can still end it. The
  in-fight refusal comes first, so that actor can end it only once nobody is fighting it.

**`Heal` is fixed** (`c.Mana += mana`) in its own commit.

### Corpses (Calabe, 2026-11-11, plan review round 2)

- **Both ordinal forms, for every name:** `2.rat` and `rat#2` are the second rat, for items,
  characters and corpses alike, read by the one parser `util.GetMatchNumber`.
- **One numbering per name: the living first, then corpses newest first.** With a live rat and two
  dead rats, 1 is the live rat, 2 the rat killed most recently, 3 the rat that died first. The
  living are numbered as today (`FindByName`, unchanged); corpses count past them, newest first by
  `RoundCreated`, players' and NPCs' corpses alike. A dead character's name reaches its corpse, and
  an exact `<name> corpse` wins over a prefix. The person typing is not in their own name's living
  count for a corpse: Bob's `get all bob` and `look bob` reach his newest corpse. A corpse command
  (`get ... <corpse>`, `bury`) whose ordinal lands on someone living refuses ("The rat is still
  alive."; a player's name takes no "The": "Bob is still alive.") and does not skip ahead.
- **One visibility rule, `rooms.Perceives`:** a hidden body is perceived only by itself, an admin, or
  a viewer with `see-hidden`. The room view, every lookup of a room's bodies (`FindByName` and its 12
  callers), the corpse count, messages (`canSee`, which keeps its light and distance rules), GMCP's
  room contents and combat all ask it. A hidden body the viewer does not perceive takes no ordinal,
  cannot be named and no reply reveals it. The corpse lookup
  reuses the existing matcher (`util.FindMatchIn`'s loop, exposed by position as
  `util.FindMatchIndexIn`) and serves every caller (`get`, `look`, `bury`).
- **`bury` refuses a player's corpse that holds any item or gold,** for every actor, its owner
  included: "You can't bury the <name> corpse while it still holds something." An empty player
  corpse and every NPC corpse (with its loot) bury as before. Principle: no single command lets
  anyone destroy a player's belongings.
- **`look <corpse>` lists what the corpse holds** (the panel's fallback is fixed; it showed an empty
  box).

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

### Help pages (ruling 5)

`help death` is rewritten for the new rules:

- At 0 health you die.
- What you drop: the items you carry and your gold go into your corpse; what you wear stays on you.
- You come back in the recovery room at 1 health and heal fast there, then leave by the portal.
- Your corpse holds your things until it crumbles to dust after a while. **When it crumbles,
  everything still in it is destroyed**, so go back for it.
- An NPC's corpse holds its loot the same way, and its contents are destroyed the same way.

`help health` loses its bleed-out and -10 lines and points to `help death`. `help get`
(`templates/help/get.md`) mentions taking things from a corpse, and that its contents are destroyed
when it crumbles. `help get` and `help bury` (the cleanup module's `bury.template`) name `rat#2` for
the second of two same-named corpses, and `help bury` says a player's corpse that holds anything
cannot be buried (round 2 rulings).

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
- Ordinal parser: `2.rat`, `rat#2`, `0.rat`, `rat`, `1.5`, `.rat`, `2.`.
- Corpse numbering: corpses newest first in both forms, one list for players and NPCs, exact over
  prefix; with the living present, 1 is the living (the player when a player and an NPC both match),
  corpses start after them, an ordinal on the living is refused (`get`, `bury`; a player's name
  without "The"), and the one typing is not counted in their own name (`look bob` by Bob shows his
  corpse).
- Hidden: a hidden body cannot be named or reached by ID, takes no corpse ordinal and is never in a
  refusal; `see-hidden` and an admin perceive it; a body perceives itself. Messages: a hidden speaker
  reads by name to an admin and to `see-hidden`, "Someone" to others, and "Someone" to an admin in
  the dark. Combat: a hidden defender stops an attacker who does not perceive it, not an admin.
  The room view, GMCP's room contents, `Char.Enemies`, a say's GMCP sender, GMCP's leave notice and
  refreshes, untargeted `attack`, the walk-in notice and the follow lines each give a hidden body
  only to a viewer who perceives it (round 3).
- `bury`: refused for a player's corpse holding gold (its owner asking) or an item (another asking);
  an empty player corpse and an NPC corpse with loot are buried.
- `look <corpse>` lists the corpse's items.

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
5a. One visibility rule (`rooms.Perceives`: hidden, `see-hidden`, admins) for the room view, name
    lookups, messages, GMCP room contents and combat. Goldens unchanged.
5b. Both ordinal forms in `util.GetMatchNumber`; one numbering per name (the living, then corpses
    newest first); `bury` buries the corpse it names. Goldens unchanged.
5c. `bury` refuses a player's corpse that holds anything. Goldens unchanged.
6. One death routine (with the corpse panel's fix): `internal/death`, `OnDeath`, XP and alignment into `dikubasic`; the deletions
   above; bleed-out and the -10 branch. Goldens change; each diff summarised.
7. `suicide` refuses in a fight and in the recovery room.
8. Guards.
9. Stub-ruleset tests.
10. Docs: help pages (ruling 5), patch notes, status, design notes, strip plan decision 12,
    `geomud-verifying` count. The help pages are player-facing and could ride with commit 6
    instead; the plan decides, and the reads goldens show any help text they cover.

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

## Settled while planning (controller, 2026-11-10, after approval)

The plan draft found these by reading the code; the controller took each recommendation. None
changes the design above except Q1's outcome shape. Details, with line numbers, are in the plan's
"Questions settled" section.

- **Q1. Awards, not award lines.** `GrantXP` sends its own line, so a ruleset that granted XP would
  send text and put "You gained N experience points!" before "has died." (today it comes after).
  `DeathOutcome` carries `Awards []rules.Award{UserId, XP, Messages}`: `dikubasic` works out each
  share and applies the alignment shift in today's order (`UpdateAlignment` may draw a random number);
  the engine, after announcing the death, credits the kill, calls `GrantXP` and sends the award's
  lines. Today's reading order and draw order hold. `DeathOutcome.Messages` above becomes this.
- **Q2. A corpse holds exactly the drops** (player-visible). A corpse stores a copy of the body, and
  `get` and the corpse panel also hand out or show that copy's worn gear. Unreachable today
  (`CorpseItems: false`); with `CorpseItems` removed, anyone could take copies of a player's kept gear
  or an NPC's unrolled worn items. `death.Die` stores the body copy with nothing carried, worn or held.
- **Q3. Three more engine health checks go through `Condition`:** the downed gate in `TryCommand`
  (`usercommands.go:281`, `IsDisabled`) and in `go` (`go.go:46`), and `remove.go:51`. Same answers
  under `dikubasic`; without this a stub's `Downed` would not stop the actor acting, and the new guard
  would fail on `remove.go`.
- **Q4. Tests get a recovery room.** Package tests load no config, so `DeathRecoveryRoom` is 0, and
  moving a player to room 0 panics in `CreateEphemeralRoomIds` (`ephemeral.go:144-146,188`; read, not
  run). Fixtures that reach a player death use a test-only recovery room. The engine panic stays (the
  shipped room is 75) and goes on the status Watch list.
- **Q5. Smoke bot names are zero-padded** (`Smk%sc%03d`, `cmd/smoke/main.go:124,131`): the bot takes
  its ring back from its corpse by name, and corpse lookup matches by prefix, so `c1` would match
  `c10 corpse`.
- **Q6. More goldens change in commit 6, all decided:** `mvp`, `pvp`, `hidden_attacker` (a downed
  player now dies in the round), `mvm` (the rat's death now runs), and reads golden H's follow section
  (the rat's gold is in its corpse, not on the floor). Four tests change with them
  (`ruleset_stub_test.go`'s queued-`suicide` assert, `TestKilledByOutlivesTheKiller` moves under a
  downed-state stub, the wimpy test and the upkeep test).
- **Q7. Smaller choices the plan makes:** upkeep and the immediate buff trigger kill any actor the
  ruleset calls dead; an account always respawns; respawn keeps `Mana = 0`; with corpses off, drops
  use today's NPC floor lines for everyone; `MobItemDrop` fires for every NPC drop; reviving an account
  sends vitals; `PlayerDrop` for accounts only; recovery regen applies to any actor there and always
  rolls; commit 3 keeps the raw recovery compare (parity) and commit 5 switches to the helper;
  `suicide` in the recovery room keeps its error return; help pages ride in commit 10;
  `get <item> <word>` always tries a corpse; `actors.BodyReplaced` loses its caller and waits for
  slice 4; the ID guard's permadeath allowlist entry goes.

**Plan review round 1 (controller decisions):**

- **`get all <corpse>` works** (player-visible). Today `get.go:39` returns early whenever the first
  word is `all`, so `get all <corpse>` and `get all from <corpse>` silently do nothing. With every
  death's drops in a corpse, that is the main way back to your things. When the word after `all`
  (or after `all from`) names a corpse in the room, `get` takes every item and the gold from that
  corpse; otherwise `get all ...` keeps today's room-wide behaviour exactly (so `get all ground`,
  `get all stash` and `get all <container>` are unchanged). Tested; the help pages describe it.
- **Death drops queue `ItemOwnership{Gained: false}`** for each item that leaves the body, as the old
  `drop` did and as every other item move does (no listener today).
- **`IsDeathRecoveryRoom` is false when no recovery room is set** (0 or less): with 0 it would match
  room 0 and every empty ephemeral room (`ephemeral.go:92` records 0 as their original).
- **The `Shadow Realm` check was live, not broken:** the shipped recovery zone is named "Shadow Realm"
  (`rooms/shadow_realm/zone-config.yaml`, `75.yaml`), so `suicide.go:41` already covered the copies.
  The helper replaces it for the reason of not naming a zone, not to fix it.
- **Upstream's inert Death Recovery buff goes:** room 75's mutator still applies buff 24 (`24-death_recovery.yaml`),
  which does nothing since its script was deleted. Ruling 2's regen rule replaces it, so buff 24 and
  the mutator's reference to it are deleted in commit 5.
- **Event order (Q1):** `AlignmentChanged` now queues before `MobDeath` as well as before
  `GainExperience`. Both have only GMCP listeners; accepted.
- **A ruleset may queue an event** (`dikubasic`'s alignment shift queues `AlignmentChanged`). The
  no-send guard bans sends, not events; slice 4 deletes alignment, so this is not generalised.
- **`suicide` with `revive-on-death`** revives the player at full health (the flag works for anyone
  now). No shipped buff carries the flag; accepted.
- **Help and docs** promise `get all <corpse>` only once the fix above lands; the help pages ride
  in the docs commit after it.

**Plan review round 2 (2026-11-10): every item applied to the plan 2026-11-11** (the plan's "Plan
review round 2: what this revision changed" lists each with its task and step). The round ran Tasks
1-9 on a scratch copy: every task builds, 40 ok plus the 3 known failures, Task 6 changed exactly the
nine predicted goldens, and draw order held. Items:

- **Blocker, controller decision (bug fix):** `look <corpse>` shows an empty box.
  `buildCorpseInventoryPanel` (`look.panels.go:112-121`) falls back to a panel named `inv` but writes
  every row to `equip`, so the rows are dropped. Unreachable while `CorpseItems` was false; fix the
  fallback to create `equip` (as `buildInventoryLookPanel`'s fallback does at `:57-59`; the name
  `buildEquipmentPanel` first written here was wrong) in commit 6. The fallback always runs: no data
  folder has the `character/equipment-look` layout (`panel-layouts/` holds only
  `character/inventory.yaml`). **Applied:** plan Task 6 Step 7; its test is
  `TestLookAtACorpseShowsWhatItHolds` (Step 11b), with a probe. Confirmed on a scratch worktree:
  empty box and `no panel with id "equip"` before, the item listed after.
- **For Calabe: `bury` on a full corpse.** `modules/cleanup/cleanup.go:141-160` removes a corpse with
  no check on its contents, and anyone may bury anyone's corpse. Once a player's corpse holds their
  carried items and gold, a stranger can destroy them in one command (reads golden H already buries a
  rat corpse holding 25 gold). Options: refuse to bury a corpse that holds anything; spill the
  contents to the floor first; only the owner may bury a player's corpse.
  **Ruled by Calabe 2026-11-11: refuse if a player corpse.** `bury` refuses a player's corpse that
  holds any item or gold, for every actor including its owner; an empty player corpse and every NPC
  corpse bury as today (an NPC corpse's loot is destroyed with it). Principle: no single command lets
  anyone destroy a player's belongings. (Calabe meant "a player's corpse is never buried"; he
  ruled both readings acceptable and kept this one, so an empty player corpse may be buried.)
  Enforced by a test in the plan (refused with contents, allowed
  when empty, NPC corpse with contents still buried). **Applied:** plan Task 5c (commit 5c, before
  the death commit), `TestBuryRefusesAPlayersCorpseThatHoldsAnything`; refusal text in "Decisions
  taken without asking"; help `bury` and the patch notes say it (Task 10).
- **For Calabe: two corpses with the same name.** `FindCorpseByRef` (`rooms.go:945-990`) lists each
  name once and the first corpse wins, so ordinals (`2.rat`) do not work and a bare name takes the
  first prefix match. Two rats killed in one room: the second rat's loot is unreachable until the
  first corpse crumbles; a player who dies twice in one room cannot reach the newer corpse. Options:
  fix the lookup (every corpse listed, ordinals, an exact "`<word>` corpse" match first), or accept
  and document it. **Ruled by Calabe 2026-11-11: fix the lookup in this slice.** Every live corpse
  is a candidate (no de-duplication by name), and the existing ordinal syntax works: `rat#2`
  (`util.GetMatchNumber`, `util.go:298`; the `2.rat` above was wrong, the codebase has no dot
  ordinal). `util.FindMatchIn` (`util.go:316`) already does the ordinal and exact-before-prefix; the
  de-duplication in `FindCorpseByRef` is what defeats it, so the fix reuses `FindMatchIn` and adds no
  matcher of its own. Player corpses still rank before NPC corpses. The plan fixes it for every caller
  of `FindCorpseByRef`/`FindCorpse`, with tests for two same-name corpses (bare name, `#1`, `#2`,
  exact over prefix). **Applied:** plan Task 5b (commit 5b). `FindMatchIn` returns the matched text,
  which cannot tell two corpses of one name apart, so its loop moves unchanged into
  `util.FindMatchIndexIn` (positions) and `FindMatchIn` wraps it: one matcher still. `FindCorpse`
  delegates to `FindCorpseByRef`; `look.go`'s unused third copy of the name lists goes. Callers
  checked (`grep -rn 'FindCorpse' --include=*.go .`: `get.go:92`, `get.go:422`, `look.go:394`,
  `cleanup.go:141`): only `bury` needed an edit, because `RemoveCorpse` matches owner, name and round
  and would remove the first of two rats killed in one round; it now removes the corpse it found
  (`RemoveCorpseRef`). Tests: `rooms` (bare, `#1`, `#2`, `#3` absent, exact over a prefix that comes
  first, players first, a crumbled corpse skipped), `util` (`FindMatchIndexIn` over repeats), `bury
  rat#2` (Task 5b) and `get all rat#2` (Task 6). **Goldens: none change** (the only golden lookup is
  reads golden H's `bury rat` with one corpse in the room); Task 6 still changes exactly nine. Code,
  tests and three probes were run on a scratch worktree at `65bbc144`: pass, and fail under the
  probes. **Open, for Calabe:** see "#N counts within one kind" below. (Superseded the same day:
  the two-list numbering and its player-first ranking were replaced by the rulings below, and Task
  5b was reworked to them.)
- **Major (plan):** two "prints ok" steps run `-count=16` over packages with the known
  `TestCombatGoldenFleeSuccess` repeat flake; scope them to the death tests or skip that test.
  **Applied:** plan Task 6 Step 11b(c) runs `-run 'Death|CombatGolden(MvP|PvP|HiddenAttacker)$|Wimpy|Respawn'`
  in `hooks` (checked with `-list`: no flee test) and `-run 'Corpse'` in `usercommands`.
- **Minors (plan):** `remove.go` probes that do not compile; Task 6 Step 8's grep runs before the
  steps that clear its hits, and its control count is 129 lines in 19 files; Task 5's grep list misses
  its own new comment; mutations that survive (respawn ignoring `RespawnHealth`, ignoring
  `Respawn: false` for NPCs, the drop gold clamp, the "item not held" guard, `PlayerDeath` killer
  fields, the death event-log line, `CleanupMobSpawns`, NPC regen in the recovery room, container
  precedence in `namesACorpse`, the recovery line's `Range`/`Senses`); the death guard's scope
  prefixes need a per-prefix reach check; two fixture comments need new text; `RegenResult.Changes()`
  and the `IsDisabled` ban belong in "Decisions taken without asking"; `suicide` staying
  `AllowedWhenDowned` next to "refused in a fight" needs one sentence.
  **Applied:** `remove.go` probes keep a `rules.` call (Task 3 Step 9b); the absence grep moved to
  Task 6 Step 13 with a new Step 0 control (129 lines in 19 files, rechecked: 117 in 18 at
  `65bbc144` plus Task 1's 12); Task 5 Step 5 names `ruleset.go`; each surviving mutation now has a
  test and probe (Task 5: NPC regen, `Senses`/`Range` [round 5: the `Range` probe as written could
  not fail; see the round 5 section]; Task 6 Step 11b: `PlayerDeath` fields and the
  event log in `TestRespawnStartsClean`, `CleanupMobSpawns` in `TestDeathFreesTheNPCsSpawnSlot`,
  container precedence in `TestGetAllWithoutACorpseTakesTheFloor`; Task 9: respawn health and NPC
  respawn in `TestStubRulesetDecidesTheRespawn`, the gold clamp and the "item not held" guard in
  `TestStubRulesetCannotDropWhatTheBodyLacks`); Task 8 checks each scope prefix reaches a file; Task
  6 Step 11 gives the two fixture comments; the decisions table holds the two calls; the design's
  `suicide` section and Task 7 Step 3b say the sentence.

**Asked of Calabe (found applying round 2; answered by the ruling that follows): "#N" counts within one kind.** The lookup matches player
corpses and NPC corpses as two lists, players first (today's ranking, kept by the ruling), so `#N`
counts within a kind: with Ratty's corpse (a player) and two rat corpses, `rat` and `rat#1` find
Ratty's, `rat#2` finds the second rat (not the first), and the first rat is reached as `rat corpse`.
One combined list (players first) would make `#N` count across both, but would also change today's
order when only a word inside a name matches (`contains` fallback). The plan implements the two
lists. Options: keep it (rare: it needs a player corpse and NPC corpses sharing a prefix in one room),
or count across one list.

**Ruled by Calabe 2026-11-11 (replaces both options): one numbering, living first, then corpses
newest first.** Principle: a name counts every living character in the room before any corpse, and
corpses count from the most recent death back. With a live rat and two dead rats: 1 is the live rat,
2 is the rat killed most recently, 3 is the rat that died first. The name of the dead character
works as the name of its corpse (`rat` reaches rat corpses). Because the living come first, every
lookup of a living target numbers exactly as today; only corpse lookups change, counting past the
living matches. Recency is `Corpse.RoundCreated` (`corpse.go:13`), newest first. Applied by the
controller as written below; the player-before-NPC ranking of corpses is dropped (see the decisions
table). Calabe confirmed the controller's two calls the same day: players and NPCs share the rule
(no player-first ranking), and a corpse command whose ordinal lands on a living character refuses
("The rat is still alive.") instead of skipping to the next corpse.

**Applied (2026-11-11):** plan Task 5b, reworked. **The living count** is read from today's code:
`FindByName` (`rooms.go:1334-1341`) numbers players (`findPlayerByName`, in the order they came into
the room) and NPCs (`findMobByName`, likewise; its internal `name#k` only keys a map) apart, and all
12 callers take the player when both are found. So ordinal n reaches someone living exactly when n
is at most the matching players or at most the matching NPCs: the living count is the larger of the
two, and for an ordinal both reach, the player is the one named. `livingReach` asks `FindByName`
itself, so living lookups cannot drift. A hidden NPC counts, as `FindByName` lists it for `look`
today. Corpses then count from the newest (`RoundCreated` descending; of two from one round, the
later added). The refusal: `get <item> <word>`, `get all <word>` and `bury` say "The <name> is still
alive." with the living character's name; `look` needs none (it shows the living first, as today).
Tests: `rooms` (newest first, both forms, one list, exact over prefix, crumbled skipped,
`RemoveCorpseRef`), `hooks` (`TestCorpseNumbersCountPastTheLiving`: a live player Ratty and a live
rat share 1, corpses at 2 and 3, Ann's own corpse at `2.ann`; `bury rat` refused, `bury 2.rat`),
`usercommands` (`get all rat` refused, `2.rat`, `rat#3`; Task 6). Code, tests and four probes were run
on a scratch worktree: pass, probes fail, full suite at the 3 known Windows failures. Found while
running: the hooks test file must sort after `combat_golden_test.go`, because a reads-world test
that runs before `TestCombatGoldenFleeSuccess` sets off that test's known order flake. (Superseded
the same day by the two rulings below: a hidden NPC no longer counts, and Ann's own corpse is `ann`,
not `2.ann`.)

**Ruled by Calabe 2026-11-11: both ordinal forms, everywhere.** `2.rat` and `rat#2` both work, for
every command that takes a name (items, characters, corpses). One parser: `util.GetMatchNumber`
accepts both forms; no second parser.

**Applied (2026-11-11):** plan Task 5b Step 3. Its two callers are the two name matchers
(`util.FindMatchIn`, `util.go:322`; `items.FindMatchIn`, `items.go:604`), so every name lookup gets
both forms; `findMobByName`'s `#`/`@` guard needs no dot twin (a nameless `2.` stays as typed), and
ID tokens (`0x...:0x...`) never read as the dot form. World data: no name, title, zone or map key
starts with digits and a dot (two greps over the 100 YAML files under `_datafiles` and `modules`, 0
lines each; both shown able to match a planted line), and character names cannot hold a dot
(`NameRejectRegex` `^[a-zA-Z0-9_]+$`). Parser tests: `2.rat`, `rat#2`, `0.rat`, `rat`, `1.5`, `.rat`,
`2.`, and three more. `help item-names` gains the dot form (Task 10).

**Ruled by Calabe 2026-11-11: the person typing is not in their own name's living count** when
they loot or look at a corpse. `get all bob` typed by Bob reaches Bob's newest corpse; other
players' numbering is unchanged. A player's name never takes "The" in the refusal line.

**Ruled by Calabe 2026-11-11: a hidden character is not perceivable until revealed to the viewer.**
Principle: what a viewer cannot perceive does not exist in the list of things that viewer's names
are matched against. It takes no ordinal, cannot be named, and no reply (refusal or otherwise)
reveals it. This applies to every name lookup, living and corpse, not only corpses.

**Applied (2026-11-11), the person typing:** plan Task 5b. The corpse lookup takes the viewer and
leaves it out of its own living count, for `get`, `bury` and `look`'s corpse branch; `FindByName`
itself keeps the viewer, so no living number moves for any command. Today `look bob` typed by Bob
matches Bob himself (`findPlayerByName` lists every player, and `look` has no self check), so he
looks at himself; now he does only when no corpse answers to his name, and otherwise sees his newest
corpse. `consider` already drops a self match and `attack` refuses one; `give` and `show` have no
self check and keep none. The refusal names a player bare ("Bob is still alive.") and an NPC with
"The". The smoke bot takes its ring back with its own name again.

**Applied (2026-11-11), hidden:** plan Task 5a (new; before 5b). Hidden is the buff flag `hidden`;
reveal to a viewer is the viewer's `see-hidden`; the room view applies both per viewer, inline
(`roomdetails.go:252-256`, `:268-272`). The living lookups do not filter today (`FindByName` passes
`FindAll`, and `GetMobs`/`GetPlayers` then return every body), so `look`, `attack` and the rest reach
a hidden body. The room view's check becomes `rooms.Perceives(viewer, subject)` (a body perceives
itself; a hidden one only a viewer with `see-hidden` perceives), used by the room view, by
`FindByName` (now taking the viewer; all 12 callers pass the body typing, by name and by ID token),
and by the corpse count. Size: the 12 callers in 10 files, `rooms.go`, `roomdetails.go`, 4 test
calls: within the coordinator's bound. Every caller has a viewer (each is a command run by a body);
the changes per caller are listed in Task 5a, and the two calls that needed a choice are in the
decisions table (admins get no bypass; GMCP and messaging keep their own stricter checks).

**Goldens after both:** none change in 5a, 5b or 5c; Task 6 still changes exactly nine. The goldens'
hidden bodies are only ever the ones typing (a body perceives itself), and no golden command holds a
digits-dot word. Run on a scratch worktree: every golden byte-identical, with and without the hidden
filter; full suite at the 3 known Windows failures.

**Ruled by Calabe 2026-11-11: admins always perceive hidden characters.** The bypass lives in the
one shared check (`rooms.Perceives`), so an admin's room view, name lookups and messages all show
hidden characters; no per-command bypass.

**Ruled by Calabe 2026-11-11: one visibility rule everywhere.** GMCP room contents and messaging's
`canSee` move onto the shared check in this slice, so `see-hidden` and admins reveal hidden
characters there too. Principle: every place that decides whether a viewer perceives a character
asks the same function.

**Applied (2026-11-11), admins and one rule:** plan Task 5a, widened. The shared check is
`rooms.PerceivesWith(flag, viewer, subject)` (a body perceives itself; an unhidden body is
perceived; a hidden one only by an admin or a viewer with `see-hidden`; a nil viewer by no one), with
`rooms.Perceives` reading the bodies' own buff flags. The admin test is the codebase's own: an
account whose `Role` is `accounts.RoleAdmin` (`userrecord.go:27`), the role `HasRolePermission` lets
through first (`userrecord.go:386-389`), found from the viewer by `actors.AccountOf`
(`registry.go:70`). Every place that decides whether a viewer perceives a character was grepped and
listed in Task 5a with file:line. **Moved onto the rule:** the room view (`roomdetails.go:252`,
`:268`), the name lookups (`FindByName`), the corpse count, GMCP's room contents
(`gmcp.Room.go:346`, `:374`), messaging's `canSee` hidden clause (`perceive.go:51`) and combat's
"Can't see them, can't fight them" (`NewRound_DoCombat.go:196`). **Left alone, with reasons** (also in
the decisions table): the sneaking actors' own silences (`go.go`, `look.go`, `read.go`,
`cleanup.go`: whether a hidden actor's line is sent at all; sending it would show everyone else
"Someone ...", a widening); darkness (`look.go:28-31`, `:239`; `canSeeIn`), a rule of the viewer's
light that `canSee` keeps, so a hidden speaker in the dark still reads "Someone" to an admin; the
`RoomChange.Unseen` field (no reader); the "hidden" adjective and the `{h}` prompt (drawing, not
deciding). **Messaging's stricter parts kept:** a far receiver and a receiver in the dark see no
one, as before; only the hidden clause gains the reveal. **No import cycle:** `rooms` imports none of
`messaging`, `hooks`, `usercommands` or `modules/gmcp`, which already import `rooms`; the admin test
uses `actors` and `accounts`, already imported by `rooms`; messaging passes its own test seam
(`hasFlag`) to `PerceivesWith`, so its tests keep faking flags. Size: 4 production files and 2 test
files beyond the first hidden ruling's set. **Goldens:** none change: reads goldens A to C make Ann
an admin but hide no one; D, E, F and H hide Ann with no admin or `see-hidden` in the room; the
combat golden `hidden_attacker` hides the attacker, not the defender; GMCP is in no golden. Run on a
scratch worktree: identical, full suite at the 3 known Windows failures.

**Ruled by Calabe 2026-11-11: combat follows the one rule.** An admin or a `see-hidden` attacker
may fight a hidden defender they perceive (`NewRound_DoCombat.go:196` moves onto `Perceives`).

**Ruled by Calabe 2026-11-11: sneaking actors' silences are deferred.** The lines a hidden actor's
move, look, read and trash suppress (`go.go`, `look.go`, `read.go`, `cleanup.go`'s `trashCommand`,
`:89-100`; "bury" was written here first, wrongly) stay silent to
everyone, admins included, in this slice. Follow-up: GitHub issue pruuk/GeOMUD#33 (a messaging
feature that sends a line only to viewers who perceive the actor).

Both round 2 questions answered 2026-11-11 (above), and the rulings after them applied.

**Plan review round 3 (2026-11-11): every finding applied to the plan** (the plan's "Plan review
round 3: what this revision changed" lists each with its task and step). The root cause of the two
majors: Task 5a's site list came from grepping hidden-related words, which cannot find a path that
never checked hidden. The search was redone over every place that enumerates a room's bodies, every
GMCP module, every `events.Communication` producer and consumer, every untargeted target pick and
every line that names one body to another; Task 5a's table lists each site with file:line, moved or
left alone with a reason.

- **M1** `Char.Enemies` (`gmcp.Char.go:455-466`) sent a hidden enemy's Id and Name: filtered by
  `Perceives`; tested.
- **M2** a say's `Comm.Channel` (`gmcp.Comm.go`, from `say.go:40`) named a hidden speaker the text
  shows as "Someone": per receiver, the name only when `Perceives` passes, else `Someone` with no
  player or mob `Source`; tested. The other channels (party, broadcast, whisper) are left alone: their
  speaker addresses the audience by name.
- **Found by the redone search and moved under the same rulings:** GMCP's `Room.RemovePlayer` on
  leaving the game (`gmcp.Room.go` `despawnHandler`) and its room refreshes on a move or a fight
  change; the walk-in "<NPC> notices you as you enter!" line (`go.go:280-300`); the follow lines
  (`follow.go:203-222`, `:465-475`, `:501`; see the follow ruling below). Each tested, with a probe.
- **m1** tests for the room view (`roomdetails.go:252`, `:268`) and GMCP's room contents
  (`gmcp.Room.go:346`, `:374`), with probes. **m2** untargeted `attack` (`attack.go:48-150`) skips a
  body the attacker does not perceive; tested. **m3** the look test's comment no longer names a deleted
  knob. **m4** see the decisions table (corpses with neither ID). **m5** Task 5b's sort probe
  prediction fixed. **m6** Task 1's predictions: `hp=1/7`, no Remove lines in `death_player`. **m7**
  the silences ruling names `trash`. **m8** `get.md` and the `item-names` anchor fixed.

Code and tests for all of it were run on a scratch worktree: pass; each probe named in Task 5a Step 7
fails; `go test ./internal/hooks -count=1` passes twice; the full suite shows the 3 known Windows
failures and 41 ok (`modules/gmcp` gains its first test file). **Goldens:** none change (no golden
hides a body from a viewer who reads GMCP, follows, attacks untargeted or walks into a hidden hostile;
reads golden H's follow lines name an unhidden Bob exactly as before). Next: a fresh blind round 4.

**Ruled by Calabe 2026-11-11: follow does not cancel hiding.** Shadowing someone is classic
skullduggery. Principle (his hidden principle, applied to lines): what a viewer cannot perceive does
not exist for them, and no reply reveals it, so **a line whose only content is a hidden body's
presence is dropped for a viewer who does not perceive that body; a line that carries other news the
viewer gets anyway keeps "Someone"**. Applied to each substitution round 3 had added (plan Task 5a):

| Line | Its content | Now |
|---|---|---|
| "X is following you." (to the target) | the follower's presence only | **dropped** for a target who does not perceive the follower |
| "X stopped following you." (to the target, on `follow stop` or when a follow breaks) | the follower's presence only | **dropped** likewise |
| "You are no longer following X." (to the follower) | the follower's own follow ended, which it learns anyway | **kept**, "someone" for a leader it does not perceive |
| A say's GMCP `Comm.Channel` `Sender` | the said text, which the receiver reads anyway ("Someone says ...") | **kept** as `Someone`, `Source` empty |

Every other round 3 site already drops rather than substitutes (`Char.Enemies`, GMCP's leave notice
and refreshes, the walk-in notice, untargeted `attack` picks). Messaging's "Someone" stays for lines
that carry an act the viewer gets anyway (a say's words, an attack's blow, a death's corpse); a
messaging line that only announces a hidden body's presence is dropped for viewers who do not
perceive it (round 4: leaving the game, starting to quit, appearing by teleport, fleeing, dropping,
reviving, recovering). Tested by `TestHiddenFollowerIsNotAnnounced` (a hidden Bob follows and stops:
Ann reads nothing; made an admin, she reads both lines), with a probe on each target line; run on a
scratch worktree: pass, both probes fail, `internal/hooks` passes twice, reads golden H unchanged.

**Plan review round 4 (2026-11-11): every finding applied to the plan** (all follow from existing
rulings; no new question):

- **M1** `follow lose` (`follow.go:436-452`, also run by a followed player's teleport, `:296`) named a
  leader the follower no longer perceives: the follower's own line now uses `seenName` ("someone");
  `TestFollowLoseNamesOnlyAPerceivedLeader`, with a probe.
- **M2** the `player-despawn` template line (`PlayerDespawn_HandleLeave.go:53-60`) read "drag someone
  away" to viewers who could not perceive the leaver: it is only presence, so it is **dropped** for
  them, as GMCP's `Room.RemovePlayer` already is. Mechanism: `rooms.NotPerceiving(body)` lists the
  bodies in its room that do not perceive it, passed as the message's `Exclude`. The sweep of every
  room-scoped line naming an actor for arrival, departure, login, logout, spawn, death, recovery or
  wake (plan Task 5a's "Presence-only lines" table): **dropped** for unperceiving viewers: leaving
  the game, starting to quit (`quit.go`), appearing by teleport (`admin.teleport.go`, two lines),
  fleeing (`NewRound_DoCombat.go`), dropping downed (`PlayerDrop_HandlePlayerDrop.go`), the recovery
  room's line (`recoveryRegen`), the revive (`death.Die`); **keeps "Someone"**: "has died." (the
  corpse is news the viewer gets anyway); **already silenced**: walking in, out and into walls (the
  deferred sneaking silences); **left alone**: a new character's first arrival (`start.go`, it cannot
  be hidden yet) and the global death notice; **none exists**: login, spawn and wake lines. Six
  production sites plus the revive in Task 6's new code. `TestPresenceLinesReachOnlyThoseWhoPerceive`
  (`quit`, leaving the game, `NotPerceiving` itself), with a probe on each line.
- **Minors:** Task 5b's `RemoveCorpseRef` test looks up `rat` (the newest, gold 2), so the probe that
  puts back `RemoveCorpse` now kills it (reasoned: `RemoveCorpse` takes the first, gold 1; not re-run,
  as the test needs Task 5b's lookup); Task 5a Step 7's always-true probe is now predicted to change
  `hidden_attacker.txt` line 11 and `F-comms-party.txt` line 22 (both seen on the scratch run); tests
  and probes for `aggroChangedHandler`, `roomChangeHandler`'s old-room loop and the follow lines for
  an NPC follower and an NPC leader; `seenName`'s comment sits above its own function;
  `TestPerceivesAnUnhiddenBodyAndItself` gives its two bodies distinct IDs.

Run on a scratch worktree (Task 5a with the round 3 and round 4 changes): every new test passes, each
named probe fails, `go test ./internal/hooks -count=1` passes twice, `go test ./... -count=1` shows
the 3 known Windows failures and 41 ok. **Goldens:** none change (no golden hides a body that quits,
leaves the game, flees, teleports, drops or follows; reads golden E's flee is by an unhidden Ann).
Next: a fresh blind round 5.

**Before round 5 (2026-11-11): every moved presence line has its own test and probe** (a site that
forgets its `Exclude` would have passed the round 4 tests). `TestEachPresenceLineReachesOnlyThoseWhoPerceive`
has one subtest per site (teleport, the teleport party loop, flee, drop); deleting each site's
`Exclude` fails its own subtest only (run). The recovery line gets
`TestRecoveryLineExcludesThoseWhoDoNotPerceive` in Task 5a (run against a stand-in, since Tasks 3
and 5 were not applied; its `Regen` form is unrun), and the revive gets
`TestReviveOfAHiddenBodyReachesOnlyThoseWhoPerceive` in Task 6 (unrun: the `death` package is Task
6's). The `RemoveCorpseRef` minor above is now run, not reasoned: with Task 5b applied the test
passes, and putting back `RemoveCorpse` fails it ("the one named went, the other stayed").

**Plan review round 5 (2026-11-11): the whole plan ran, Tasks 1 to 10;** everything builds and the
nine predicted goldens hold. M1 is **the controller's application of Calabe's principle** ("what a
viewer cannot perceive does not exist for them; no reply reveals it"), **not a Calabe ruling**;
Calabe may overrule it (its rows are in "Decisions taken without asking").

- **M1, the death notice named a hidden killer.** "*** Bob has DIED! (killed by Alice) ***" went to
  everyone while Bob and Carol only ever read "Someone stabs you" (the `hidden_attacker` golden
  pinned it). Now each reader's "(killed by ...)" lists only the killers that reader perceives
  (`rooms.Perceives`); with none left the part is dropped, never "someone", which would tell of an
  unseen killer. Perceiving a hidden body takes being in its room: a reader elsewhere never perceives
  a hidden killer, except an admin, who perceives hidden bodies anywhere. An unhidden killer is named
  to all, as today. The notice was one `events.Broadcast` string; it becomes one `UserId` message per
  active account (`messaging.NotifyEach`), written with the same line refresh and prompt redraw.
  `hidden_attacker`: only Alice reads "(killed by Alice)"; Bob and Carol read the notice without it.
  The PK records (`KD.PlayerDeaths`, "userid:name") still hold the hidden killer's name; nothing shows
  them to a player (GMCP shows a player only their own kills).
- **Minor 8, the fight refusals and a hidden attacker.** Not applied: a fight's victim already
  perceives the fight through the attacker's hit lines, so the refusal reveals nothing; counting only
  perceived attackers would let quit escape an unseen attacker.
- **Minors:** Task 5's `Range` probe could not fail (`Room` is the zero value, `perception.go:49`):
  it now widens the range; the round 2 note above is marked. `death_player` at seed 303 removes and
  drops the testsword (`ItemOwnership Alice item=93101`): Tasks 1 and 6 now say so. `get.md`'s stash
  line needs two trailing spaces or `internal/markdown` joins it to the next (rendered to check).
  Two probes that did not compile now keep their names in use (`_ =`). `bury.template`'s new lines go
  after line 7. Three site-table rows pasted into the presence table moved back.
  `DeathOutcome.Messages` is dropped: Q1 made awards carry the lines, and nothing set or tested it.

Run on a scratch copy: today's `suicide.go` sending its notice through `NotifyEach` passes the notice
test written for the reads world, three probes fail as stated, and `go test ./... -count=1` shows only
the 3 known Windows failures, every golden unchanged. Task 6's form of the notice test is unrun (it
needs the `death` package).

**Plan review round 6 (2026-11-11): nothing material.** The whole plan ran, Tasks 1 to 10, on one
scratch copy: everything builds, every named test passes, exactly the nine predicted goldens change,
and the tests earlier rounds could not run (recovery line in its `Regen` form, revive, death notice,
`suicide`) pass with their probes failing. Shuffled and repeated runs are clean apart from the known
`TestCombatGoldenFleeSuccess` flake. Four text slips fixed by the controller: the Extra Lives range is
`users.html:366-370` (its `</div>` is on 370); two probes in Task 6 Step 11b(b) and Task 7 Step 4
needed `_ =` lines to compile; Task 5's cap probes fail first at the d10 test's full-health assert.
One unrecorded choice added to the decisions table (`perceivesBody`). Review stops here: six rounds.

**Calabe at plan review (2026-11-11):** asked about two burials of one corpse and burying a corpse
that just rotted. Commands and the round tick run one at a time (`MainWorker` under
`util.LockMud()`, `world.go:316-380`), so the second command resolves its name afresh, and a crumbled
corpse is never a candidate. Task 5c Step 4b pins this with a test (added after round 6, not
reviewed blind). **Plan approved by Calabe 2026-11-11** with that test added.

**Facts corrected while planning:** help pages are under `_datafiles/world/empty/templates/help/`;
more health checks exist than the table lists (`NewRound_AutoHeal.go:68`, `remove.go:51`, and in
`characters` `:464`, `:1204` and the remove-lock checks); `save_and_load.go:437` is another raw
recovery compare (room loading only; left alone); the admin pages' ranges are wider and
`users.html:366-370,592,1089` (Extra Lives) is missing; `configs_bench_test.go:28,174-176` name deleted
knobs; `decayAlignment` is `:71-91`.

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
| `RegenResult.Changes()` (in `internal/rules`) holds the regen pass's "anything to apply" check | The regen pass is in a guarded package; written there as `r.Health != 0 \|\| r.Mana != 0` it compares `.Health` with a number and fails `TestEngineAsksTheRulesetWhoIsDead` | Inline the check and let the guard skip comparisons against 0 (weakens the guard) |
| The death guard also bans a body's `IsDisabled` in the scanned packages (`X.Character.IsDisabled()`, or any `IsDisabled` call in a file not on a two-file allowlist of `items.Item.IsDisabled` callers, each with why; a stale entry fails) | `IsDisabled` is `Health <= 0`: a health check by another name that would bypass the ruleset's `Condition` (Q3) | Drop the ban; a later edit could reintroduce a body-health check the guard does not see |
| `bury`'s refusal reads "You can't bury the <name> corpse while it still holds something." (told to the actor only; the room sees nothing) | Calabe asked the plan to pick it; it says why without urging anyone to loot another player's corpse | Reword in `cleanup.go` and its test |
| The corpse lookup reuses the matcher by moving `util.FindMatchIn`'s loop, unchanged, into `util.FindMatchIndexIn` (positions); `FindMatchIn` wraps it | Calabe ruled "reuse `FindMatchIn`, no new matcher", but its text result cannot tell two corpses of one name apart; a position can. Every other caller sees the same results (`TestFindMatchIn` unchanged) | Map text back to a corpse some other way, which means re-implementing part of the match |
| `FindCorpse` delegates to `FindCorpseByRef`, and `look.go`'s unused copy of the name lists goes | Three copies of one lookup is how the de-duplication bug lived on | none |
| `bury` removes the corpse it found (`Room.RemoveCorpseRef`, by address); `RemoveCorpse` stays for its tests | `RemoveCorpse` matches owner, name and round, so with two rats killed in one round `bury rat#2` would bury the first | none |
| The lookup fix and the `bury` refusal are commits 5b and 5c, before the death commit | The death commit then lands on a world where corpses hold belongings safely and every corpse is reachable | Move them after commit 6; no code changes |
| The dot form is digits, a dot, then a name not starting with a digit; `#N` is read first; `0.rat` is 1 (as `rat#0`); `1.5`, `.rat` and `2.` stay as typed | Keeps decimals and stray dots literal; one rule shape with `#N`'s clamp. World data has no digits-dot name to misread | Accept `1.5` as the fifth "5", or refuse `0.` |
| The living count is the larger of the matching players and the matching NPCs, asked of `FindByName` itself (`livingReach`); for an ordinal both reach, the player is the one named | That is the set of ordinals a living lookup reaches today (two lists, callers prefer the player), so living lookups number exactly as before | Sum them (players then NPCs as one list): living ordinals past the first kind would then differ from today's |
| ~~A hidden NPC counts as living~~ (replaced 2026-11-11 by Calabe's hidden ruling: a hidden body is in no viewer's list) | | |
| The refusal applies only when some corpse answers to the name; otherwise the lookup finds nothing and the command goes on as for any word | Keeps round 1's ruling that `get all <word>` with no corpse of that name stays room-wide, and `get <item> <word>` unchanged when no corpse is involved | Refuse whenever the ordinal lands on the living |
| The refusal names the living character it landed on, as the viewer sees it: "The Rat is still alive." for an NPC named Rat, "Bob is still alive." for a player (Calabe ruled no "The" for players); told to the actor only | The ruling's text with the real name; the player when both kinds match | Use the typed word instead |
| The person typing is left out of their own name's count only in corpse lookups; `FindByName` keeps them in its list | Calabe's ruling names looting and looking at a corpse; keeping living lookups as they are moves no living number for anyone | Leave the typer out of every lookup: `consider bob#2` by Bob and the like would renumber |
| `look <own name>` shows the newest own corpse when one answers to the name, else the looker as today | The ruling ("`look bob` reaches his newest corpse") with today's self-look kept when there is no corpse | Always show the corpse branch's "not found" when no corpse |
| `give` and `show` keep having no self check | Not in the rulings; unchanged behaviour | Add "You can't give to yourself" in a later slice |
| The visibility check is the room view's (`hidden` unless the viewer has `see-hidden`), moved into `rooms.Perceives` and used by the room view, `FindByName` and the corpse count | Calabe asked to reuse the room view's check; one function keeps them from drifting | none |
| ~~GMCP's room contents and messaging's `canSee` keep their own checks~~ (replaced 2026-11-11 by Calabe's one-rule ruling: both ask `Perceives`) | | |
| ~~No admin bypass~~ (replaced 2026-11-11 by Calabe's ruling: admins always perceive hidden characters, in `Perceives`) | | |
| The admin test is `actors.AccountOf(viewer.Id).Role == accounts.RoleAdmin`, the role `HasRolePermission` lets through first; mods are not included (a mod with `see-hidden` perceives as any viewer) | The ruling names admins; it is the codebase's own admin test | Include mods (`RoleMod`) |
| The rule is `rooms.PerceivesWith(flag, viewer, subject)`; `Perceives` passes `HasBuffFlag`, messaging passes its `hasFlag` seam | Messaging's tests fake buff flags through that seam (`messaging_test.go:42`); one body of code, two flag readers | Change messaging's tests to real buffs and drop the seam from the call |
| Messaging's `canSee` keeps its light and distance rules; only its hidden clause moves | They are rules of the viewer's senses, stricter than hidden; moving them would change what players read beyond the reveal | Fold light into `Perceives` (would also gate name lookups and GMCP by light: a narrowing) |
| Combat's "Can't see them, can't fight them" asks `Perceives`: an admin or `see-hidden` attacker can now fight a hidden defender (**player-visible, flagged**) | It is a perception decision, and the ruling is one rule everywhere | Keep combat blocking every hidden defender (a second rule) |
| The sneaking actors' own silences (`go.go:197`, `:214`, `:273`, `:330`; `look.go:56`, `:91`, `:122`, `:256`, `:293`, `:332`; `read.go:39`; `cleanup.go:100`) are left alone: a hidden actor's line is still not sent, so admins and `see-hidden` do not read it either (**flagged**) | Each decides whether a line exists, not who perceives the actor; sending it through messaging would show everyone else "Someone ...", a widening the ruling forbids | Add a message flag that drops the line for receivers who do not perceive the actor, then send it (new messaging behaviour) |
| Darkness stays out of `Perceives` (room view's `look.go:28-31`, `:239`; messaging's `canSeeIn`); GMCP room contents and name lookups still ignore light, as today | Light is a separate rule of the viewer's room; adding it would narrow what clients and commands see today | Make light part of the one rule |
| `RoomChange.Unseen` (`roommanager.go:429`, `rooms.go:691`) is left as is | Nothing reads it and it has no viewer | Delete it in a later slice |
| Lookups of online players anywhere by name (`whisper`, `party invite` when not `SameRoomOnly`, admin `locate`, `teleport`, `visit`) are unchanged | They do not look up a room's bodies; hidden is room presence | Apply the rule when the target is in the viewer's room |
| A nil viewer perceives no hidden body; only tests pass nil (every production caller has a body typing) | No viewer is no one in particular, who cannot see hidden | Treat nil as all-seeing |
| A hidden body is unreachable by ID token too | "Cannot be named" includes the client's ID form; after Task 5a GMCP sends a hidden body's ID (room contents, `Char.Enemies`) only to a viewer who perceives it. (This row first said GMCP never lists a hidden body's ID; `Char.Enemies` did, round 3 M1.) | Allow IDs (a client holding an old ID could reach it) |
| Of two corpses from one round, the one added later is newer | `RoundCreated` is a round number; the later `AddCorpse` is the later death | none |
| `look` gets no refusal | It looks up the living first (`look.go:76`), so an ordinal on the living shows them, as today | none |
| The smoke bot takes its ring with `get copper ring <its name>` | The one typing is not in its own name's count (Calabe, 2026-11-11), so its name reaches its corpse | none |
| The new hooks test file is `reads_corpse_test.go`, sorting after `combat_golden_test.go` | A reads-world test before `TestCombatGoldenFleeSuccess` sets off its known order flake | Fix the flake (status Watch), then the name is free |
| The corpse lookup lists every corpse that has not crumbled, without today's `UserId > 0` / `MobId > 0` test (`rooms.go:909-922`), so a corpse with neither ID is findable (round 3, m4) | The engine never makes one (`death.Die` sets `UserId` for an account, `MobId` for an NPC; a test NPC with template 0 is the only way), and a name is a name | Keep the test: a corpse with neither ID is unreachable by name, as today |
| An unperceived say on GMCP reads `Sender: "Someone"` and an empty `Source` | The rulings: the receiver gets what messaging shows ("Someone says"); `Source` "player" or "mob" would tell it what kind of body hid | Keep `Source` |
| GMCP sends a viewer no room refresh when a body it does not perceive moves or changes its fight, and no `Room.RemovePlayer` for one that leaves the game | A refresh or a leave notice is presence; the contents it would carry already leave the body out | Keep the refreshes (contents unchanged, only the timing tells) |
| ~~Follow lines read "Someone"~~ (replaced 2026-11-11 by Calabe's follow ruling: the target's lines are dropped; the follower's own line keeps "someone") | | |
| `party invite` and `whisper` still name a hidden sender | The sender addresses the receiver by name on purpose, as on the party and whisper channels; not room presence | Show "Someone" there too |
| `give` and `show` are left as they are | Each cancels the giver's hidden before naming it | none |
| The death notice's "(killed by ...)" names, per reader, only the killers that reader perceives; with none it is dropped, never "someone" (the controller's application of Calabe's principle, round 5 M1; not a Calabe ruling) | "No reply reveals" a body the reader does not perceive; "someone" would reveal an unseen killer's presence | Name every killer to everyone, as today, or write "someone" |
| A hidden killer in another room is named only to an admin; a `see-hidden` reader elsewhere does not read it (round 5 M1) | `see-hidden` reveals what is in the reader's room; admins perceive hidden bodies everywhere (Calabe's admin ruling) | Let `see-hidden` reveal a hidden killer anywhere |
| The per-reader notice is one `UserId` message per active account (`messaging.NotifyEach`), not a broadcast with a receiver list | Same bytes, line refresh and prompt redraw; no change to `events.Broadcast` or its writer; the goldens record it the same way. A screen-reader account now gets the notice stripped, as every message is | Add a receiver list to `events.Broadcast` |
| `DeathOutcome.Messages` is dropped | Q1: awards carry the death's lines; nothing set or tested it | Restore it with a test |
| GMCP treats a body no longer in the registry as perceived (`perceivesBody`) (round 6) | A body that has left has nothing left to hide; the refresh reveals nothing new | Treat it as not perceived and skip the refresh |
| A floor item named by the whole phrase wins over a corpse named by its last word (whole-branch review M1) | It restores today's floor pickup: `get iron key` beside a Keymaster corpse took the key before corpses were findable by any word's prefix, and now it does again, living Keymaster or not | Let the last word name the corpse first: `get iron key` then searches the Keymaster corpse (or is refused while a Keymaster lives), and the key is reachable only as `get iron` or `get iron key from ground` |
