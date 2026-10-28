# Phase 4b slice 6 inventory (2026-10-28, master a1d51725)

Three read-only inventories written for the slice 6 plan (`2026-10-28-phase-4b-6-one-tick-plan.md`): G, the round tick; H, idle and behaviour; I, the world loop, input ordering and the blast radius. Each lists the files it read. Their open questions were settled by Calabe or by the plan's defaults (plan, "Decisions"). This file moves to `docs/completed/` with the plan.

# Inventory G: the round tick (Phase 4b slice 6)

Branch `strip/phase-4b-6-one-tick`, HEAD `3d7f7e47`, read 2026-10-28. Read-only; no repo file changed.

Read 44 of 44 files (all printed with `cat -n`/`sed -n`, every chunk under 400 lines):

| # | File | Lines read |
|---|---|---|
| 1 | `2026-10-27-phase-4b-npcs-are-players-design.md` | full (1-345) |
| 2 | `internal/hooks/hooks.go` | full (1-73) |
| 3 | `internal/hooks/NewRound_UserRoundTick.go` | full (1-132) |
| 4 | `internal/hooks/NewRound_MobRoundTick.go` | full (1-111) |
| 5 | `internal/hooks/NewRound_IdleMobs.go` | full (1-140) |
| 6 | `internal/hooks/MobIdle_HandleIdleMobs.go` | full (1-53) |
| 7 | `internal/hooks/NewRound_InactivePlayers.go` | full (1-62) |
| 8 | `internal/hooks/NewRound_UpdateZoneMutators.go` | full (1-24) |
| 9 | `internal/hooks/NewRound_CheckNewDay.go` | full (1-34) |
| 10 | `internal/hooks/NewRound_HandleRespawns.go` | full (1-26) |
| 11 | `internal/hooks/NewRound_AutoHeal.go` | full (1-81) |
| 12 | `internal/hooks/NewRound_DoCombat.go` | partial: 40-80, 205-240, plus a grep of `actors.`/`mobs.GetInstance` |
| 13 | `internal/hooks/NewTurn_ActionPoints.go` | full (1-23) |
| 14 | `internal/hooks/NewTurn_PruneBuffs.go` | full (1-78) |
| 15 | `internal/hooks/Buff_ApplyBuffs.go` | 14-81 (imports skipped) |
| 16 | `internal/hooks/charmexpiry_test.go` | full (1-90) |
| 17 | `internal/actors/controller.go` | full (1-74) |
| 18 | `internal/actors/actor.go` | full (1-28) |
| 19 | `internal/actors/registry.go` | full (1-247) |
| 20 | `internal/actors/wire.go` | full (1-58) |
| 21 | `internal/accounts/userrecord.go` | 440-500 |
| 22 | `internal/accounts/users.go` | 200-250, 345-370 |
| 23 | `internal/characters/cooldowns.go` | full (1-44) |
| 24 | `internal/characters/character.go` | 500-535, 1515-1600 |
| 25 | `internal/rooms/rooms.go` | 868-898, 2020-2120 |
| 26 | `internal/rooms/roommanager.go` | 475-505 |
| 27 | `internal/events/eventtypes.go` | 1-20, 120-160 |
| 28 | `internal/events/listeners.go` | 30-192 |
| 29 | `internal/events/events.go` | 55-234 |
| 30 | `world.go` | 330-430 |
| 31 | `internal/behaviour/behaviour.go` | full (1-160) |
| 32 | `internal/mobs/mobs.go` | 1-80, 210-340, 460-500, 560-620 |
| 33 | `internal/buffs/buffs.go` | 227-275 |
| 34 | `internal/buffs/buffspec.go` | 28-58 |
| 35 | `internal/util/util.go` | 170-200 |
| 36 | `internal/util/hook.go` | full (1-49) |
| 37 | `internal/guards/commands_test.go` | 56-131 |
| 38 | `internal/guards/guards_test.go` | grep hits only (lines 53-57, 77-106) |
| 39 | `internal/inputhandlers/login.go` | 66-110 |
| 40 | `modules/follow/follow.go` | 30-330 (of 525) |
| 41 | `modules/gmcp/gmcp.Gametime.go` | 1-60 |
| 42 | `modules/gmcp/gmcp.Char.go` | 101-116 (grep context) |
| 43 | `modules/all-modules.go` | full |
| 44 | `_datafiles/world/empty/mobs/startland/1-rat.yaml`, `2-guard.yaml` | full (counted as one entry); `_datafiles/config.yaml` grep hits only |

**Skipped on purpose:** the rest of `NewRound_DoCombat.go` (slice 5's, out of this scope); the bodies of
`actors/*_test.go`, `behaviour/behaviour_test.go`, `hooks/actionpoints_test.go`, `hooks/mobflee_test.go`,
`rooms/npcslots_test.go`, `rooms/behaviour_rest_test.go` (found by grep, names and hit lines only);
`follow.go` 331-525; `guards_test.go` outside the grep hits. Not traced: what an account sees between
`LoginUser` and `enterWorld` (see section 9, item 2).

---

## 1. Where NewRound comes from and who hears it

- **Fired** once per round from the turn timer: `world.go:386-391`
  (`turnCt % TurnsPerRound() == 0` then `events.AddToQueue(events.NewRound{RoundNumber, TimeNow})`),
  under `util.LockMud`. `TurnMs: 50`, `RoundSeconds: 4` (`_datafiles/config.yaml:160,164`).
- **Type:** `events.NewRound{RoundNumber uint64; TimeNow time.Time}` (`internal/events/eventtypes.go:120-125`).
- **Dispatch:** `events.ProcessEvents` (`events.go:176-233`) pops the priority heap (lower number first,
  then insertion order, `events.go:65-70`) and calls `DoListeners` (`listeners.go:151-192`), which runs
  the type's listeners in list order and stops at the first non-`Continue`. Events a listener queues
  at priority 0 run later in the same `ProcessEvents` call (the loop runs until the heap is empty).
- **Listener order** (`listeners.go:56-109`: a plain registration appends before any `Last` listener;
  `First` prepends). Module `init()`s run before `main` (`main.go:52` imports `modules`, which imports
  `cleanup`, `follow`, `gmcp`, `webhelp`: `modules/all-modules.go`); `hooks.RegisterListeners()` runs at
  `main.go:204`. So the real NewRound order is:

| # | Listener | Registered | Scope |
|---|---|---|---|
| 1 | `follow.onNewRound` (follow-limit expiry, texts) | `modules/follow/follow.go:70` (init) | accounts and NPCs (follow keys) |
| 2 | `gmcp.Gametime.newRoundHandler` | `modules/gmcp/gmcp.Gametime.go:25` (init) | accounts (GMCP) |
| 3 | `InactivePlayers` | `hooks.go:17` | account-only (idle kick) |
| 4 | `UpdateZoneMutators` | `hooks.go:18` | world (zones) |
| 5 | `CheckNewDay` | `hooks.go:19` | world (queues `DayNightCycle`) |
| 6 | **`UserRoundTick`** | `hooks.go:20` | rooms with players + accounts in them |
| 7 | **`MobRoundTick`** | `hooks.go:21` | every NPC instance |
| 8 | `HandleRespawns` | `hooks.go:22` | rooms with players (`room.Prepare(false)`) |
| 9 | `DoCombat` | `hooks.go:26` | actors (accounts first, then NPCs: `NewRound_DoCombat.go:49-62`) |
| 10 | `AutoHeal` | `hooks.go:30` | account-only (every 3 rounds) |
| 11 | **`IdleMobs`** | `hooks.go:31` | every NPC instance; queues `MobIdle` |
| later | **`MobIdle` listeners**: `follow.idleMobHandler` (`First`, `follow.go:68`), then `HandleIdleMobs` (`hooks.go:32`) | | per NPC |

  Note: the `hooks.go:16-31` comment layout suggests hooks run first; in fact the two module listeners
  run before all of them. Tests that call `events.ClearListeners()` (`internal/rooms/move_test.go:26,38`)
  are the only callers of `ClearListeners`.

- **Other direct calls of the listeners:** `DoCombat(events.NewRound{})` in
  `internal/hooks/combat_golden_test.go:423`, `mobflee_test.go:83,105`. No test calls `UserRoundTick`,
  `MobRoundTick`, `IdleMobs` or `HandleIdleMobs` (grep of the four names over `--include=*.go .`
  returned only the definitions and `hooks.go`; the same grep found `modules/follow/follow.go`, so it
  reaches modules).

## 2. `UserRoundTick` (`internal/hooks/NewRound_UserRoundTick.go`, 132 lines)

Loops `rooms.GetRoomsWithPlayers()` (`:23`; `roommanager.go:482-505` iterates the map
`roomManager.roomsWithUsers`, so **room order is Go map order**), then `room.GetPlayers()` (`:77`).

| Line | Step | Class |
|---|---|---|
| 26-27 | `room.RoundTick()` for each room with a player (see 2.1) | room (not actor) |
| 29-75 | Room idle message: 5% (20% if `RoomId == -1`) to print one of `room.IdleMessages`, else the zone's; a numeric first entry means "use that room's messages"; repeat-suppression via `room.LastIdleMessage`; `room.SendText(msg)` | room (not actor) |
| 79-82 | `accounts.GetByUserId(uId)`, skip nil | (a) iteration |
| 85 | `user.Character.Cooldowns.RoundTick()` | (a) any actor |
| 88-101 | Alignment decay every `AlignmentDecayRounds` (config 100, `config.yaml:353`): `DecayAlignment()`; if the alignment *name* changed, `user.SendText("Your alignment has shifted from X to Y!")` and queue `events.AlignmentChanged{UserId,...}` | (b) account-only today; the decay itself is character-level (see G-3) |
| 103-105 | `Charmed.RoundsRemaining--` when > 0 | (a) any actor; but players never expire (see 4) |
| 107-115 | `Buffs.Trigger()`; queue `BuffsTriggered{ActorId: user.Character.Id, BuffIds}` | (a) any actor |
| 118-123 | `Character.Validate()`; if Health or Mana changed, queue `CharacterVitalsChanged{UserId}` | Validate (a); the event (b) (keyed by `UserId`, only listener `modules/gmcp/gmcp.Char.go:45`) |

### 2.1 `Room.RoundTick` (`internal/rooms/rooms.go:2038-2080`)

Called only from `UserRoundTick:27` (grep `\.RoundTick()`; tests `rooms/npcslots_test.go:37,48,81`). So
**it runs only for rooms with a player**:
mutator update and area buffs to players/mobs/native mobs (`:2047-2054`); spawn-slot settle and save
(`:2060-2064`, decision 14); `mob.BoredomCounter = 0` for every NPC in a room with a player
(`:2068-2074`); corpse decay (`:2079`). None of this is per actor; it is the room's tick.

### 2.2 `Account.RoundTick` (`internal/accounts/userrecord.go:462-464`)

Empty body. **No caller** (grep `RoundTick(` over all `.go` files: definitions at `userrecord.go:462`,
`cooldowns.go:9`, `rooms.go:2038`, `NewRound_MobRoundTick.go:54`, `NewRound_UserRoundTick.go:19`; calls
only at `UserRoundTick:27,85`, `MobRoundTick:73`, `npcslots_test.go`, `cooldowns_test.go:43`). Dead code.

### 2.3 `Cooldowns.RoundTick` (`internal/characters/cooldowns.go:9-13`)

Decrements every tag; pruning happens in `Prune` (`:15-21`), called from `Try` and from
`Character.Validate` (`character.go:1569`). Callers: `UserRoundTick:85`, `MobRoundTick:73`,
`cooldowns_test.go:43`. Identical for both kinds.

## 3. `MobRoundTick` (`internal/hooks/NewRound_MobRoundTick.go`, 111 lines)

| Line | Step | Class |
|---|---|---|
| 59 | `mobs.ReduceHostility()` once per round: decrements `mobsHatePlayers[group][userId]` (`mobs.go:568-583`). Set by `DoCombat:217` (`MakeHostile`), read by `usercommands/go.go:274` and `combat_golden_test.go:380` | world step (group-hatred table, keyed by account `UserId`) |
| 64-70 | Loop `mobs.GetAllMobIds()` (`mobs.go:228-235`: **map order**), `GetInstance`, skip nil | (c) iteration |
| 73 | `Cooldowns.RoundTick()` | (a) |
| 75-77 | `Charmed.RoundsRemaining--` when > 0 | (a) |
| 79-87 | `Buffs.Trigger()`; queue `BuffsTriggered{ActorId: mobId}` | (a) (same event; `mobId == Character.Id`) |
| 90-98 | Charm expiry: when charmed and `RoundsRemaining == 0`: `RemoveCharm()`, charmer's `TrackCharmed(id,false)` via `accounts.GetByBodyId`, then `runCharmExpiredCommand` (`:20-52`: a `rest` part calls `behaviour.Rest` and stops; other parts run now through `usercommands.TryCommand` when resting, else queue as input via `mob.Command`) | (c) NPC-only today |
| 101 | `Character.Validate()` (no vitals event) | (a) |
| 103-106 | `Health <= 0` then `mob.Command("suicide")` (NPC death path, `usercommands.go:187-189`) | (c) NPC-only (the death branch, Phase 5) |

`runCharmExpiredCommand` callers: `MobRoundTick:97`, `charmexpiry_test.go:70,84`.

## 4. User vs mob: same step, different behaviour

| Step | Accounts (`UserRoundTick`) | NPCs (`MobRoundTick`) |
|---|---|---|
| Who is ticked | only accounts listed in a room in `roomsWithUsers` | every instance in `mobInstances`, in any room |
| Cooldowns | same | same |
| Alignment decay | yes, with text and `AlignmentChanged` | **never** |
| Charm countdown | decrements to 0, then nothing: a charmed account stays charmed forever | decrements; at 0 the charm is removed and `ExpiredCommand` runs |
| Buff trigger | same; `ActorId: user.Character.Id` | same; `ActorId: mobId` (identical value) |
| Validate | yes, plus `CharacterVitalsChanged` on HP/MP change | yes, no event (the event is keyed by `UserId`, so it cannot name an NPC) |
| Health <= 0 | nothing here (bleed-out and the `-10` suicide are in `AutoHeal:44-56`, account-only) | queues `suicide` |
| Group hatred decay | n/a | `ReduceHostility` once per round (world) |
| Buff prune (per turn, `NewTurn_PruneBuffs.go`) | accounts in rooms with players: prune, Validate, vitals, `BuffsTriggered`, `BuffsRemoved`, buff 0 logs off | every NPC: prune, Validate only |
| Regen (`AutoHeal`, every 3 rounds) | yes | **never** (NPCs do not regenerate) |

**Upstream bugs and drift found**

- **G-1. Boredom never accrues.** `BoredomCounter` is only reset (`rooms.go:2071`, `IdleMobs:53`,
  `mobs/saved.go:58`); nothing increments it (grep `BoredomCounter` over all `.go`). Upstream
  incremented it in `mobcommands/lookfortrouble.go:160-161` (`git grep` on `upstream/master`), stripped
  in Phase 2 (`gomud_strip_plan.md:100` lists `lookfortrouble`). So `IdleMobs:46-57` (the memory valve,
  decision 15) is dead today. The grep did find `upstream/master` hits, so it could find an increment.
- **G-2. A charm-expiry rest is followed by work on the gone NPC.** After `runCharmExpiredCommand` rests
  the NPC (`MobRoundTick:97`), the same iteration still runs `Validate` (`:101`) and, at `Health <= 0`,
  queues `suicide` for an ID that is no longer an instance (`:103-106`). Harmless today (the input
  finds no actor), but a merged tick should stop the iteration after a rest.
- **G-3. Players' charm never ends** (`UserRoundTick:103-105` counts down with no expiry). Nothing calls
  `Character.Charm` outside tests (grep `\.Charm(`: `behaviour_test.go:139`, `character_test.go:31`,
  `combat_golden_test.go:518,550`), so charm is reachable today only through restored saves; no player
  is ever charmed.
- **G-4. `Buffs.Trigger(buffId...)` filter is a no-op** (`buffs.go:233-238`: `continue` inside the inner
  loop). Both ticks call it with no argument, so it does not bite here. A triggered buff has no effect
  beyond `RoundCounter`/`TriggersLeft` and the `BuffsTriggered` event, whose only listener is GMCP
  (`gmcp.Char.go:52`, returns early for NPCs via `actors.AccountOf`, `:108-111`).
- **G-5. Order is map order, twice.** Rooms (`roomsWithUsers`) and mob IDs (`mobInstances`) are both
  ranged as maps, so the order in which NPCs tick, idle and draw from `util.Rand` changes run to run.
  `Room.GetRandomExit` also picks by index over a map (`rooms.go:870-898`), so with two or more exits a
  seeded `util.SeedRandForTest` (`util.go:181-185`) does **not** fix which exit a wander takes.
- **G-6. `MaxWander` counts moves, not distance, and allows one extra.** `Wander` increments
  `WanderCount` and the checks are `WanderCount > MaxWander` (`behaviour.go:33,53`;
  `HandleIdleMobs:38`). The rat (`maxwander: 8`) can make 9 wander moves before it heads home.
  `WanderCount` resets only when a path ends in the home room (`IdleMobs:127-129`), or on restore
  (`saved.go:58`); `follow.go:310` also increments it. `MaxWander: -1` wanders forever; `0` never.
- **G-7. `GetIdleCommand` has a 1-in-100 "do nothing"** before choosing (`mobs.go:486-488`), and the rat
  lists `''` as one of three idle commands (`1-rat.yaml`). Expected per-round odds for an idle rat at
  home: `10/100 * 99/100 * 1/3` each for emote, wander, nothing-listed. `LastIdleCommand` is never
  read or written outside saves (grep: `mobs.go:62`, `saved.go:58`, `saved_test.go`), so repeats are
  not suppressed (unlike room idle messages).
- **G-8. `MaxMobBoredom` is cast to `uint8`** (`IdleMobs:27`); a config above 255 wraps. Config ships 250.

## 5. `IdleMobs` and `HandleIdleMobs` (all NPC-only, class c)

`IdleMobs` (`NewRound_IdleMobs.go`), after `DoCombat` and `AutoHeal`:

| Line | Step |
|---|---|
| 25-34 | `allowedUnloadCt = len(instances) - MobUnloadThreshold` (config 100, `config.yaml:102`) |
| 38-44 | loop `GetAllMobIds()` (map order); nil decrements the allowance |
| 46-57 | boredom valve: at `BoredomCounter >= maxBoredom`, `behaviour.Rest(mob, "depression n/m")` if `Despawns()` (not a shopkeeper, `mobs.go:327-332`), else reset counter. Dead today (G-1) |
| 59-63 | `PreventIdle`: skip this round and clear. Set only by `DoCombat:234` (an attacked NPC with no aggro) |
| 65-78 | in a fight: if the aggro target is neither an NPC instance nor an account body in the same room, `emote mumbles about losing their quarry.`, clear aggro, queue `AggroChanged`. Always `continue` (no idle while aggro is set) |
| 80-131 | path walking: if the current step is not the NPC's room, re-plan to the remaining waypoints or home (`behaviour.PathTo`); else issue `mob.Command(exitName)` for the next step, plus `mob.Command("noop", 1)` at a waypoint (the slice-4 note's "waypoint noop delay"); a broken step clears the path; arriving home resets `WanderCount` |
| 133 | `events.AddToQueue(events.MobIdle{ActorId})`: only reached when no path step was taken this round |
| 137 | `util.TrackTime("IdleMobs()")` |

`MobIdle` (`eventtypes.go:135-140`) listeners: `follow.idleMobHandler` (`follow.go:68`, `First`; returns
`Cancel` when the NPC is following someone, `:246-254`), then `HandleIdleMobs`:

| Line | Step |
|---|---|
| 18-21 | `GetInstance`, nil cancels |
| 28-32 | `MaxWander == 0` and away from home and not charmed: `PathTo(home)` |
| 34-36 | charmed: stop |
| 38-43 | `MaxWander > -1 && WanderCount > MaxWander`: `PathTo(home)`, stop |
| 46-50 | `util.Rand(100) < ActivityLevel`: `GetIdleCommand()`, then `behaviour.RunIdle` (`wander` and `pathto` are behaviour actions; anything else is queued input, `behaviour.go:150-160`) |

**`MobIdle` is a documented module extension point**: `_datafiles/html/admin/docs-modules.html:1123,
1557-1583` tells module authors to register `events.First` on `MobIdle` and `Cancel` to replace the
built-in AI. Deleting the event deletes that contract and the doc rows.

## 6. The actors package today

- `Controller` (`controller.go:14-20`): `Account() *accounts.Account`, `Perceive(perception.Perception)`.
  **No `Tick`** (comment `:12-13`: "Slice 6 adds Tick").
- `ConnectionController` (`:25-43`): holds the account; `Perceive` writes text.
- `BehaviourController` (`:47-74`): holds `mob *mobs.Mob` (private); `Perceive` fires
  `OnBehaviourPerceive` (`util.Hook[BehaviourPerception]`, `:65`). Registered only by tests
  (`controller_test.go:16`, `messaging_test.go:465`): **no NPC reacts to perceptions in production**.
  The slice-3 note asks slice 6 to queue reactions rather than send from inside the hook
  (`util.Hook.Fire` holds its read lock while handlers run, `hook.go:42-49`).
- `Actor` (`actor.go:16-21`): `Id`, `Character`, private `controller`. **No `Actor.Mob`** field (the
  spec note at line 266 suggests one); code reaches the mob with `mobs.GetInstance(a.Id)`
  (`NewRound_DoCombat.go:102,158,202,209`).
- Registry (`registry.go`): `All()` returns a sorted snapshot (`:93-103`, ID text order); `Each(fn)`
  runs under the read lock and **must not reach a write path** (`:105-115`).
  `mobs.DestroyInstance` fires `OnInstanceDestroyed` (`mobs.go:237-241`), which `unbindMob` handles
  with `mu.Lock` (`wire.go:23,54-58`; `registry.go:216-226`). Rest (`rooms/npcstore.go:933`) and NPC
  death (`usercommands/suicide_npc.go:381,411`) call `DestroyInstance`. So **a tick under
  `actors.Each` that rests an NPC deadlocks**; the tick must loop a snapshot (`All()` or an ID list).
- Wiring (`wire.go:18-25`): accounts become actors on `accounts.OnActivate` (fired from `LoginUser`
  and `CopyoverReconnectUser`, `users.go:244,297,324`; copyover `copyover.go:84`); NPCs on
  `mobs.OnInstanceCreated`.
- **Import graph** (`go list`): `actors` imports `accounts characters ids mobs mudlog perception util`;
  `behaviour` imports `mapper mobs mudlog rooms util`; `rooms` imports `actors`. So **`actors` cannot
  import `behaviour`, `rooms` or `usercommands`** (cycle through `rooms`). A `BehaviourController.Tick`
  written inside `actors` cannot call `behaviour.Wander/PathTo/Rest/RunIdle` or
  `usercommands.TryCommand` directly.
- Guards: `TestNoControllerTypeAssertions` (`guards_test.go:80-110`) matches selectors
  `actors.ConnectionController`/`actors.BehaviourController` outside `internal/actors/`; a controller
  type defined in another package would not be covered without editing the guard.
  `TestNoDroppedVerbsQueued` (`commands_test.go:56-109`) forbids queuing `wander`, `pathto`, `rest`,
  `despawn` as command text, so the tick must call the behaviour functions, as today.
- Prior art for one loop over actors: `NewTurn_ActionPoints.go:10-23` (`actors.Each`, character-only,
  safe because it never destroys) and `fightersWithAccountsFirst` (`NewRound_DoCombat.go:49-62`:
  accounts from `GetOnlineUserIds`, then NPCs from `GetAllMobIds`, map order kept on purpose, plan
  decision 5). Prior art for a controller-side hook out of `actors`: `OnBehaviourPerceive`.

## 7. Every reference to what slice 6 deletes or moves

| Symbol | References (all `.go` unless noted) |
|---|---|
| `UserRoundTick` | def `NewRound_UserRoundTick.go:19`; `hooks.go:20` |
| `MobRoundTick` | def `NewRound_MobRoundTick.go:54`; `hooks.go:21`; comment `usercommands/usercommands.go:188` |
| `runCharmExpiredCommand` | `NewRound_MobRoundTick.go:20,97`; `charmexpiry_test.go:70,84` |
| `IdleMobs` | def `NewRound_IdleMobs.go:23`; `hooks.go:31`; `util.TrackTime` label `:137` |
| `HandleIdleMobs` | def `MobIdle_HandleIdleMobs.go:14`; `hooks.go:32` |
| `events.MobIdle` | def `eventtypes.go:136-140`; `NewRound_IdleMobs.go:133`; `hooks.go:32`; `modules/follow/follow.go:68,247`; docs `_datafiles/html/admin/docs-modules.html:1123,1557,1563,1583` |
| `Account.RoundTick` | def `userrecord.go:462` only |
| `Room.RoundTick` | `UserRoundTick:27`; `rooms/npcslots_test.go:37,48,81` |
| `Cooldowns.RoundTick` | `UserRoundTick:85`; `MobRoundTick:73`; `cooldowns_test.go:43` |
| `mobs.ReduceHostility` | `MobRoundTick:59` only |
| `BoredomCounter` | `IdleMobs:46,49,53`; `rooms.go:2071`; `mobs.go:63`; `saved.go:58`; `saved_test.go:33,70` |
| `PreventIdle` | `DoCombat:234`; `IdleMobs:60-61`; `mobs.go:72`; `saved.go:60`; `saved_test.go:75,87` |
| `WanderCount` | `behaviour.go:33,53`; `behaviour_test.go:111,122,150`; `HandleIdleMobs:38`; `IdleMobs:128`; `mobs.go:71`; `saved.go:58`; `saved_test.go`; `usercommands/admin.locate.go:153`; `follow.go:310` |
| `GetIdleCommand` | `HandleIdleMobs:47`; def `mobs.go:480` |
| `behaviour.*` from the ticks | `HandleIdleMobs:30,41,48`; `IdleMobs:49,96,101`; `MobRoundTick:42`; also `rooms/behaviour_rest_test.go:40,59` |
| `GetRoomsWithPlayers` | `UserRoundTick:23`; `HandleRespawns:14`; `NewTurn_PruneBuffs:25`; def `roommanager.go:482` |
| `OnBehaviourPerceive` | `controller.go:63-73`; `controller_test.go:16`; `messaging/messaging_test.go:465` |
| smoke | `cmd/smoke/admin.go:58-66` and `fight.go` rely on the rat existing and wandering (the fight retries when it "wandered off", `fight.go:16-17`); nothing in `cmd/smoke/*.go` asserts idle text (grep `wander\|idle\|wiggles` hit only comments) |

No hit in `internal/guards` for any of the deleted names (grep over `.` covered it; the same grep found
`internal/guards/commands_test.go` hits for `wander`, so the directory is reached).

## 8. Options for one actor round tick

The pieces sort into four kinds: **world** steps (`ReduceHostility`, zone mutators, new day), **room**
steps (`Room.RoundTick`, idle messages, respawns), **actor** steps every body needs (cooldowns, charm
countdown, buff trigger, Validate), and **controller** steps (account: alignment text, vitals event;
behaviour: charm expiry, health-0 death, lost quarry, path walking, idle and wander). Slice 6's
deletions only require that the actor and controller kinds become one loop calling `Controller.Tick`.

**Shape 1: one actor tick at `UserRoundTick`'s slot; Tick runs everything, idle included.**
`RoomRoundTick` (rooms with players: `Room.RoundTick` + idle messages) keeps its slot; then
`ActorRoundTick` loops a snapshot, runs the shared actor steps, then `ctrl.Tick(a, round)`.
Behaviour's Tick does charm expiry, death, lost quarry, path, idle.
Effect: NPCs idle and wander **before** combat instead of after; `PreventIdle` set by combat in round N
is consumed in round N+1 (an attacked NPC gets one idle-free round one round late, and could
emote or wander in the round it is first attacked, before the swing lands: the swing still lands,
since `go` is queued input, but the text order changes). A fresh respawn from `HandleRespawns` idles
from the next round instead of the same one.

**Shape 2: maintenance before combat, behaviour after (closest to today).**
`ActorRoundTick` at `MobRoundTick`'s slot does the shared actor steps and the account's Tick part;
the behaviour part (`Tick`) runs from a second listener at `IdleMobs`' slot (after `DoCombat` and
`AutoHeal`). Two options for "Tick" here: `Tick(actor, round)` called only in the late pass (account
Tick does nothing there), or a phase argument. Effect: order the same as today except accounts and
NPCs share the early loop.

**Shape 3: one listener after combat.** Move the whole actor tick to `IdleMobs`' slot. Effect: cooldowns
and buff triggers tick after the round's swings instead of before, so a cooldown or buff that ends
this round no longer applies to this round's combat (visible as one round's difference in
`Cooldowns.Try`-gated actions and timed buffs).

**How `BehaviourController.Tick` reaches `behaviour`** (import graph, section 6):
- (i) A hook in `actors`, as `OnBehaviourPerceive`: `Tick` fires `OnBehaviourTick`; `hooks` (or
  `behaviour` at init) registers the body. Fits the existing pattern; the body lives outside the
  controller, which the spec's "Tick runs idle and wander" reads loosely.
- (ii) Inject one function at `actors.Init` (`Init(tickNPC func(*mobs.Mob, uint64))`). Same reach,
  explicit wiring, no hook lock held during the body.
- (iii) Move `BehaviourController` into `behaviour` (which may import `actors`), and inject its
  constructor into `wire.go:mobSpawned`. Puts the runtime state and Tick together as the spec draws
  it; the guard's concrete-type list must name the new package.
- (iv) Move the idle state onto the controller (spec section 1: "the mob runtime state ... spawn data"):
  `WanderCount`, `PreventIdle`, `Path`, `BoredomCounter`, `LastIdleCommand`. Touches `saved.go`,
  `admin.locate.go`, `follow.go:310`, `DoCombat:234`, `rooms.go:2071`. Optional for "done when"; can be
  a later step.

**Iteration** (any shape): loop `actors.All()` or an ID list, never `actors.Each` (deadlock, section 6).
Re-check `actors.Get(id)` before each actor (an earlier actor's Tick can rest or kill one). Order
options: keep accounts first, then NPCs, map order (as slice 5's `fightersWithAccountsFirst`), or
`All()`'s sorted ID order (deterministic, helps the seeded test).

**Replacing `MobIdle` for `follow`**: (a) a veto hook (`OnBehaviourIdle util.Hook[...]` with a `Skip`
field) that follow registers; (b) follow sets `PreventIdle` each round for a following NPC (it already
writes `WanderCount`); (c) behaviour asks a "following?" query that follow exports. Each deletes the
`docs-modules.html` `MobIdle` rows; (a) keeps a module extension point.

**Seeded test** (done-when): the rat's per-round odds are `ActivityLevel/100`, then `99/100`, then
1/3 per entry (G-7); wander moves up to `MaxWander + 1` times (G-6). To be seedable: one rat, a room
with exactly one exit per room (G-5), deterministic actor order, `util.SeedRandForTest` as
`combat_golden_test.go:167-168` does, and `events.ProcessEvents()` between rounds (wander is queued
input, `behaviour.go:54`, `mobs.go:253-284`, with `ReadyTurn`). `behaviour_test.go` already builds
rooms and mobs this way (`testRoom`, `testMob`, `queuedInput`, lines 60-101).

## 9. Player-visible changes a merge would cause

1. **Ordering** (any shape that changes slots or mixes accounts and NPCs): the text order of a charm
   farewell, an NPC death, a lost-quarry emote and a player's alignment line within a round. Shape 1
   moves NPC idle before combat (section 8).
2. **Who is ticked:** today an account is ticked only while listed in a room with players; an actor
   exists from `LoginUser` (`users.go:316-324` via `activated`) until logout. A registry loop would
   also tick an account between login and `enterWorld` (`world.go:224`), if such a window exists
   (unverified: what the connection does between those two points was not traced). Cooldowns and buffs would count down there; an alignment line could be sent.
3. **NPC alignment decay** if decay becomes an actor step: the guard (alignment 30) drifts to neutral;
   alignment drives `HatesAlignment` (`mobs.go:411`) and the alignment name players see. A rule.
4. **Player charm expiry** if expiry becomes an actor step: a charmed player would lose the charm and
   run `ExpiredCommand` as input. Unreachable today (G-3).
5. **Health-0 death in the tick** if made an actor step: a player at 0 HP would die at once instead of
   bleeding out in `AutoHeal`. Must stay on the "returns from a template" branch (spec section 3).
6. **NPC vitals event:** none (keyed by `UserId`).
7. **Follow:** a following NPC must still not idle; any replacement that misses this makes followers
   wander off mid-follow.
8. **Deleting dead code** (`Account.RoundTick`, boredom valve if dropped): no visible change.

## 10. Rules (Phase 5), not engine

- Alignment decay and who it applies to (`UserRoundTick:88-101`).
- NPC death at health 0 in the tick vs players' bleed-out at `AutoHeal` (the death rule).
- Regeneration only for accounts (`AutoHeal`), never NPCs.
- Group hatred (`mobsHatePlayers`, `ReduceHostility`), keyed by account `UserId`.
- Charm duration and what expiry does (`CharmExpiredRest`/`Revert`, `charminfo.go:6-8`).
- The idle odds themselves (`ActivityLevel`, the 1-in-100 nothing) are behaviour data the spec keeps
  unchanged (decision 2), so they are engine for 4b.

## Open questions (decisions for Calabe)

1. **Where NPC idle runs in the round** (Shape 1 vs 2 vs 3). Player-visible: with Shape 1 an NPC can
   emote or start to wander in the round it is first attacked, before the swing text; with Shape 3
   cooldowns and timed buffs end one round later relative to combat; Shape 2 keeps today's order.
2. **Alignment decay for NPCs.** Keep it accounts-only (today) or apply to every actor? Player-visible:
   NPC alignment colours and alignment-based hatred drift to neutral over time. Likely defer to Phase 5.
3. **Revive or delete the boredom valve** (G-1). It is dead today. Keep the code with no trigger,
   restore an increment (rounds without a player seen), or delete it. Player-visible only when more
   than 100 NPCs are live: NPCs nobody has seen for 250 rounds would rest out of memory and come back
   when their room next loads from disk (decision 16).
4. **The module hook that replaces `MobIdle`.** A veto hook keeps the documented extension point for
   module AI; a narrower follow-only check removes it. Player-visible: none today (only `follow`
   uses it); it decides whether a module can drive NPCs.
5. **Wander limit off-by-one** (G-6): keep "`maxwander` 8 allows 9 moves" for parity (decision 2) or fix
   to 8. Player-visible: the rat strays one room less.
6. **Actor tick order:** accounts first then NPCs in map order (today), or sorted ID order.
   Player-visible: only the order of same-round lines; sorted order makes seeded tests stable.

# Inventory H: idle and behaviour (Phase 4b slice 6)

Branch `strip/phase-4b-6-one-tick`, read 2026-10-28. Read-only; no repo file changed.

Read 46 of 46 files (each printed in chunks of at most 400 lines):

| # | File | Lines read |
|---|---|---|
| 1 | `2026-10-27-phase-4b-npcs-are-players-design.md` | 100-344 (sections 1, 2, 3, 4, Out of scope, Notes for later slices). **Skipped 1-99** (facts table and decisions), not in my brief |
| 2 | `internal/hooks/NewRound_IdleMobs.go` | full 1-140 |
| 3 | `internal/hooks/MobIdle_HandleIdleMobs.go` | full 1-53 |
| 4 | `internal/behaviour/behaviour.go` | full 1-160 |
| 5 | `internal/behaviour/behaviour_test.go` | full 1-200 |
| 6 | `internal/actors/controller.go` | full 1-74 |
| 7 | `internal/mobs/mobs_path.go` | full 1-55 |
| 8 | `internal/mobs/mobs.go` | full 1-654 (two chunks) |
| 9 | `internal/hooks/hooks.go` | full 1-73 |
| 10 | `internal/hooks/NewRound_MobRoundTick.go` | full 1-111 |
| 11 | `internal/hooks/NewRound_UserRoundTick.go` | full 1-132 |
| 12 | `internal/events/eventtypes.go` | 125-145 |
| 13 | `internal/events/listeners.go` | 150-192, plus grep 36-46 |
| 14 | `internal/events/events.go` | 60-126, 176-234 |
| 15 | `internal/util/util.go` | 165-200 |
| 16 | `internal/util/hook.go` | full 1-49 |
| 17 | `internal/rooms/rooms.go` | 580-625, 735-748, 868-906, 2038-2080 |
| 18 | `internal/rooms/roommanager.go` | 398-410 |
| 19 | `internal/mapper/mapper.path.go` | 100-240 |
| 20 | `modules/follow/follow.go` | 55-75, 182-212, 240-330 |
| 21 | `internal/hooks/NewRound_DoCombat.go` | 220-300 |
| 22 | `internal/mobs/saved.go` | 45-70 |
| 23 | `internal/mobs/testhelpers.go` | full 1-4 |
| 24 | `internal/mobs/templatefile_test.go` | 1-40 |
| 25 | `internal/actors/wire.go` | full 1-58 |
| 26 | `internal/actors/actor.go` | full 1-28 |
| 27 | `internal/actors/registry.go` | 90-118 (plus `^func` list) |
| 28 | `internal/actors/controller_test.go` | 1-45 |
| 29 | `internal/messaging/messaging.go` | 60-85 |
| 30 | `internal/messaging/messaging_test.go` | 450-500 |
| 31 | `internal/hooks/combat_golden_test.go` | 1-120, 269-292, plus grep of 120-330 and 398-424 |
| 32 | `internal/hooks/actionpoints_test.go` | 21-35 |
| 33 | `internal/rooms/behaviour_rest_test.go` | 1-40 |
| 34 | `internal/combat/simulate_test.go` | 1-30 |
| 35 | `internal/configs/config.memory.go` | full 1-39 |
| 36 | `internal/usercommands/usercommands.go` | 219-330, plus grep 140-175 |
| 37 | `internal/usercommands/go.go` | 40-75 |
| 38 | `internal/usercommands/noop.go` | 10-12 (grep) |
| 39 | `internal/usercommands/admin.mob.go` | 140-165, 355-372 |
| 40 | `internal/usercommands/admin.locate.go` | 140-160 |
| 41 | `internal/characters/character.go` | 1528-1536, grep line 53 |
| 42 | `world.go` | 90-140, grep 66, 390 |
| 43 | `cmd/smoke/admin.go` | 55-110 |
| 44 | `_datafiles/world/empty/mobs/startland/1-rat.yaml` and `2-guard.yaml` | full (counted as one entry) |
| 45 | `_datafiles/world/empty/rooms/startland/1.yaml`, `2.yaml`, `zone-config.yaml`; `_datafiles/config.yaml` 88-103, 150-167, 191 | full / partial (counted as one entry) |
| 46 | `internal/ids/ids.go` (grep only); `git show 764b845e` (hooks diff) | grep / diff |

Partial reads are partial on purpose; anything not in the ranges above is not claimed below.

---

## 1. Where idle runs today (two listeners, one event)

Registration, `internal/hooks/hooks.go`:

- `:20` `UserRoundTick`, `:21` `MobRoundTick`, `:22` `HandleRespawns`, `:26` `DoCombat`, `:30` `AutoHeal`, `:31` **`IdleMobs`** (all on `events.NewRound`), `:32` **`HandleIdleMobs`** on `events.MobIdle`.
- `NewRound` is queued by `world.go:390`. Listeners run in registration order (`events/listeners.go:174-180`), so idle runs **after combat and AutoHeal** in the same round, and `MobRoundTick` (cooldowns, charm countdown, buff triggers, health-zero suicide) runs **before** combat.
- `IdleMobs` queues one `events.MobIdle{ActorId}` per NPC (`NewRound_IdleMobs.go:133`). The queue drains in the same `ProcessEvents` pass (`events.go:209-230`: loop until empty; `qLock` released around `DoListeners`, `:221-223`), so `HandleIdleMobs` runs later in the same pass, after any events queued earlier.
- `MobIdle` event type: `internal/events/eventtypes.go:136-140` (`ActorId ids.ID`).
- Other `MobIdle` subscriber: `modules/follow/follow.go:68` registers `idleMobHandler` with `events.First`; it returns `events.Cancel` when the NPC is following (`follow.go:246-254`), which stops `HandleIdleMobs` (`listeners.go:177-179` returns on non-Continue). Admin docs page documents `MobIdle` for module authors: `_datafiles/html/admin/docs-modules.html:1123, 1557-1583`.

Grep `MobIdle|IdleMobs|HandleIdleMobs` over all non-md files found exactly these sites (the grep also found `hooks.go:31-32` and `eventtypes.go:136`, so it could find them).

## 2. The per-NPC decision, in order, with formulas

### 2a. `IdleMobs` (`internal/hooks/NewRound_IdleMobs.go`), per NPC in `mobs.GetAllMobIds()` (map iteration, `mobs.go:228-235`: random order every round)

1. `maxBoredom = uint8(mc.MaxMobBoredom)` (`:27`); `allowedUnloadCt = len(all) - MobUnloadThreshold`, floored at 0 (`:31-34`). Config: `MaxMobBoredom: 250`, `MobUnloadThreshold: 100` (`_datafiles/config.yaml:99,102`); `config.memory.go:13-14` turns `<1` into 150 (so the yaml comment "set this to 0" to never despawn is false).
2. **Boredom** (`:46-57`): if `allowedUnloadCt > 0 && BoredomCounter >= maxBoredom`: `Despawns()` (`mobs.go:327-332`: false only for shopkeepers) then `behaviour.Rest(mob, "depression N/M")`, else reset counter. `continue` either way.
   - **Dead code today.** `BoredomCounter` is never incremented anywhere. Grep `Boredom` over `*.go`: only the declaration `mobs.go:63`, resets at `rooms.go:2071` (room tick, when a player is in the room), `saved.go:58`, and the two reads above. The increment (`if mob.BoredomCounter < 255 { mob.BoredomCounter++ }`) lived in the `lookfortrouble` mob command, deleted in Phase 2.3 (`git show 764b845e`, hunk at diff line 1619-1630). So no NPC rests from boredom.
3. **PreventIdle** (`:60-63`): one-round skip, then cleared. Set only by `NewRound_DoCombat.go:232-234`: an NPC that is attacked while it has no aggro. Cleared on restore (`saved.go:60`).
4. **Aggro** (`:67-78`): any aggro skips idle. If the target is not a live NPC and is not a player in the same room, the NPC queues `emote mumbles about losing their quarry.`, clears aggro and queues `AggroChanged` (player-visible, combat territory).
5. **Path walking** (`:82-131`), runs only if `Path.Current() != nil || Path.Len() > 0`:
   - Displaced (`currentStep.RoomId() != Character.RoomId`, `:88`): re-plan through remaining waypoints (`PathTo(mob, "<wp> <wp>")`, `:90-97`) or else `PathTo(mob, "home")` (`:101`); `continue`.
   - Next step (`:107-121`): if the room's exit `nextStep.ExitName()` leads to `nextStep.RoomId()`, `mob.Command(<exitName>)` (bare exit name; `TryCommand` turns an exit name into `go`, `usercommands.go:236-240`). If the step is a waypoint, `mob.Command("noop", 1)` (`:113-117`): this queues a no-op whose ReadyTurn is 1 s later (`mobs.go:267-268`: `turnDelay = SecondsToTurns(1) * 1`), and because an NPC's input is strict FIFO (`world.go:90-113`, slice 4 decision 13) it holds back everything the NPC queues after it. `noop` is a registered command that does nothing (`usercommands.go:163`, `noop.go:10-12`). `continue` (no idle that round).
   - Otherwise (path exhausted or exit gone): `Path.Clear()` (`:125`), and **only here** `if HomeRoomId == RoomId { WanderCount = 0 }` (`:127-129`). Falls through to idle in the same round.
   - Path walking does not check charm or follow (follow's Cancel is on `MobIdle`, which is queued after this).
   - Waypoints: `mapper.findPath` marks the last room of each segment as the waypoint (`mapper.path.go:161-162`); `PathQueue.Waypoints()` (`mobs_path.go:37-50`).
6. `events.AddToQueue(events.MobIdle{ActorId: mobId})` (`:133`).

### 2b. `HandleIdleMobs` (`internal/hooks/MobIdle_HandleIdleMobs.go`)

1. Missing instance: `Cancel` (`:18-21`).
2. `MaxWander == 0 && RoomId != HomeRoomId && !charmed` then `PathTo(mob, "home")` (`:28-32`), and **continues to the idle roll** (no return).
3. Charmed: return (`:34-36`). Charmed NPCs never run idle commands.
4. `MaxWander > -1 && WanderCount > MaxWander`: `PathTo(mob, "home")`, return (`:38-43`). This does not check the room.
5. **Activity roll:** `if util.Rand(100) < mob.ActivityLevel` (`:46`), i.e. P = ActivityLevel/100.
6. `GetIdleCommand()` (`mobs.go:480-496`): `util.Rand(100) == 0` gives `""` (1 in 100); else `IdleCommands[util.Rand(len)]` uniformly; none gives `""`.
7. Non-empty: `behaviour.RunIdle(mob, cmd)` (`:47-49`).

`ActivityLevel` is clamped by `Mob.Validate()` (`mobs.go:531-535`): `<1` becomes **10**, `>100` becomes 100. A YAML `activitylevel: 0` therefore idles at 10%.

### 2c. `behaviour` package (`internal/behaviour/behaviour.go`)

- `RunIdle` (`:150-160`): `wander` runs `Wander`; `pathto <args>` runs `PathTo`; anything else is `mob.Command(command)` (queued input; `;` splits into several, `mobs.go:271-282`).
- `Wander` (`:23-57`): charmed returns (`:26-28`). Not home and `MaxWander > -1 && WanderCount > MaxWander`: `PathTo(home)`, return (`:31-38`). `MaxWander == 0`: return (`:41-43`). `room.GetRandomExit()`; if the destination room's `Zone == Character.Zone`, `WanderCount++` and `mob.Command("go <exit>")` (`:50-56`). `WanderCount` counts **queued moves**, not distance and not successful moves.
- `Room.GetRandomExit` (`rooms.go:868-906`): non-secret, unlocked exits plus active mutator exits, collected into a **map**; `util.Rand(len)` then walks the map. Map order is random per run, so with 2+ exits the chosen exit is **not reproducible under a seed**.
- `PathTo` (`:63-119`): `home` when `home-impossible` temp flag is set costs `ceil(HealthMax/10)` health and returns (`:66-75`). Targets: `home` becomes `HomeRoomId`, ints parsed (`:79-89`). No targets or first target is the current room: return without a path (`:91-93`). `mapper.GetPath` error: log WARN `pathto`, and for `home` set `home-impossible` and adjective `lost` (`:95-103`). Empty path home: same marking (`:105-109`). Else `Path.SetPath` (`:111-118`). It never moves the NPC.
- `Rest` (`:127-146`): ephemeral room ends the NPC (`EndEphemeralNPC`, refills the home slot); otherwise `room.RestNPC`. Callers: boredom (dead, above), charm expiry (`NewRound_MobRoundTick.go:41-43`).
- `mapper.findPath` iterates `node.Exits` (a map, `mapper.path.go:169`): with two equal-cost routes the route picked is not reproducible.

### 2d. Who writes the idle fields (grep over `*.go`, which also found the declarations, so it could find writers)

| Field | Decl | Written |
|---|---|---|
| `ActivityLevel` | `mobs.go:58` yaml `activitylevel` | `Validate` `mobs.go:531-535` |
| `MaxWander` | `mobs.go:70` yaml `maxwander` | spawn override `rooms.go:614-616` (only if non-zero); admin wizard `admin.mob.go:369` |
| `WanderCount` | `mobs.go:71` | `behaviour.go:53` (++), `follow.go:310` (++ per follow step), `NewRound_IdleMobs.go:128` (=0), `saved.go:58` (=0) |
| `BoredomCounter` | `mobs.go:63` | `rooms.go:2071` (=0), `NewRound_IdleMobs.go:53` (=0), `saved.go:58`. **No increment** |
| `LastIdleCommand` | `mobs.go:62` | `saved.go:58` only. **Read nowhere**: unused |
| `PreventIdle` | `mobs.go:72` | `NewRound_DoCombat.go:234` (true), `NewRound_IdleMobs.go:61`, `saved.go:60` |
| `IdleCommands` | `mobs.go:66` | spawn override `rooms.go:601-603` |
| `Path` | `mobs.go:76` | `behaviour.go:118`, `NewRound_IdleMobs.go:125`, `saved.go:59` |
| `HomeRoomId` | `mobs.go:59` | `NewMobById` `mobs.go:138` |
| `lastCommandTurn` | `mobs.go:78` | `Command` `mobs.go:261-274`; `saved.go:61` |
| `tempDataStore` (`home-impossible`) | `mobs.go:77` | `behaviour.go:99,106`; `saved.go:62` |

`ActivityLevel` has a second consumer: `combatCommand` (`NewRound_DoCombat.go:267-275`), the chance an NPC uses a combat command instead of attacking.

### 2e. What suppresses idle (complete list from the code above)

Boredom rest (dead), `PreventIdle` (one round after being attacked), any aggro, walking a path, following (follow module Cancel; path walking is not suppressed), charm (idle commands and `Wander`; path walking still runs), `WanderCount > MaxWander` (go home instead), `MaxWander == 0 && displaced` (path home, but the idle roll still runs).

**Not** suppressing idle: players in the room or zone (no player check anywhere in idle; players only reset the dead boredom counter). Room unload: an NPC only exists while loaded.

## 3. The rat

Live world for server, smoke and tests: `_datafiles/world/empty` (`_datafiles/config.yaml:191`; default in `configs/config.filepaths.go:26`; `hooks` tests point at it in `actionpoints_test.go:25-31`). `ls _datafiles/world` shows only `empty`.

`_datafiles/world/empty/mobs/startland/1-rat.yaml`:

```yaml
mobid:  1
zone: Startland
maxwander: 8
idlecommands:
  - 'emote wiggles its nose'
  - 'wander'
  - ''
activitylevel: 10
character: { name: rat, level: 1, raceid: 10, ... }
```

Spawns: room 2 "End of the Line" (`rooms/startland/2.yaml:11-19`), three slots of mob 1, no idle or maxwander override, respawn 2/3/4 real minutes. Room 2 exits: `south` to 1. Room 1 "Town Square" exits: `north` to 2 (`1.yaml:9-11`); its guard (mob 2, `maxwander: 20`, `activitylevel: 20`) has a spawn idle override with `wander` (`1.yaml:15-20`). Zone Startland, two rooms. Smoke also admin-spawns a rat in the admin's room (`cmd/smoke/admin.go:58-70`, `admin.mob.go:147`: `NewMobById(mobId, room.RoomId)`, so its home is that room).

**Rat odds per round** (idle reached, at home or `WanderCount <= 8`, not charmed, no aggro/path/follow):

- activity pass: 10/100
- given a pass: `""` 1/100; else one of three, 1/3 each
- P(wander) = 0.10 x 0.99 x 1/3 = **0.0330**; P(emote) = **0.0330**; P(nothing) = 0.934
- one exit per room, so `GetRandomExit` is deterministic here (`util.Rand(1)` still consumes a draw: `util.go:187-194` draws for any `maxInt >= 1`).
- 4 s rounds (`config.yaml:164`): about one wander per 30 rounds (2 minutes). The 9th wander (`WanderCount` 9 > 8) comes after about 273 rounds (18 minutes) on average; the rat bounces 2 to 1 to 2, so after 9 moves it is in room 1, plans home, walks `north`, and at home the path end resets `WanderCount` to 0.

## 4. Seeded-RNG facilities and test seams

- `util.SeedRandForTest(seed)` (`util.go:181-185`) replaces the shared source; `util.Rand` (`:187-194`) draws from it under a mutex. Every RNG call on the idle path goes through `util.Rand`: `MobIdle_HandleIdleMobs.go:46`, `mobs.go:486, 492`, `rooms.go:896`. The golden test seeds and restores a clock seed in cleanup (`combat_golden_test.go` lines 167-168 of the file, grep line 48-49 of the 120-330 slice).
- `math/rand` importers outside tests: `internal/events/events.go` (`:181` sampling `rand.Intn(100)` on the global source, does not touch util's source), `internal/gametime/gametimedata.go`, `internal/procedural/gridmaze.go`, `internal/util/util.go`. None is on the idle path.
- **Non-RNG nondeterminism that breaks seeding:**
  1. NPC order: `mobs.GetAllMobIds` is map order (`mobs.go:231`). `actors.All()` sorts by ID text (`registry.go:93-103`), but IDs come from `crypto/rand` (`ids.go:10,31`), so order still differs per run. With two NPCs the draws interleave differently. A seeded test must have one NPC, or sort by something stable.
  2. Exit choice with 2+ exits: map walk in `GetRandomExit` (`rooms.go:872-903`).
  3. Equal-cost paths: map walk in `findPath` (`mapper.path.go:169`).
  4. Room idle messages draw from `util.Rand` for rooms with players (`NewRound_UserRoundTick.go:47,59`). If the merged tick runs room ticks, a player in the test room adds draws.
- **Building an NPC in a test:** `mobs.NewTestInstance(m)` (`testhelpers.go:4`) calls `addInstance`, which fires `OnInstanceCreated`; `actors.Init()` (`wire.go:18-25`) makes that spawn an actor with a `BehaviourController` (`wire.go:46-52`). Pattern: `behaviour_test.go:51-55, 60-83` (`actors.Init`, `rooms.AddTestRoom`, `actors.Reset`), `combat_golden_test.go:269-290` (`NewTestInstance`, `rooms.LoadRoom(id).AddMob`), `messaging_test.go:452-462` (`actors.Spawn` + `AddTestBody`). `NewTestInstance` skips `Validate`, so `ActivityLevel` stays 0 and the roll `Rand(100) < 0` never passes: a test must set it or call `Validate`. `NewMobById` does not set `Character.Zone`; `AddMob` does (`rooms.go:742-743`).
- Loading the real rat: the `hooks` package `TestMain` already points `FilePaths.DataFiles` at the empty world (`actionpoints_test.go:25-31`), so `mobs.LoadDataFiles()` plus `races.LoadDataFiles()` would load mob 1. No current test loads mob templates (`combat/simulate_test.go:9` is inside a comment block).
- `Input` events are run only by `world.go:66` (package main). In package tests a queued `go north` is not executed; tests record it (`behaviour_test.go:86-101`, `combat_golden_test.go:112-116`) or a test listener could run `usercommands.TryCommand` (`usercommands.go:219`). `go` costs 10 action points (`go.go:49-66`); a fresh `characters.New()` body needs AP set.

**Sketch of the slice 6 done-when test** (package `hooks`, or wherever the tick lands):

1. Seed `util.SeedRandForTest(<fixed>)`, restore in cleanup.
2. Load mob 1 from the empty world (or copy `ActivityLevel 10, MaxWander 8, IdleCommands` from `GetMobSpec(1)` and call `Validate`), one NPC only.
3. Build a one-way ring of 12 test rooms in one zone, **one exit each** (room i `next` to i+1, last to first), so exit choice and paths are unique. Home = ring room 0.
4. Register an `Input` listener that runs the queued command through `TryCommand` (or moves with `AddMob`), with AP kept full.
5. Run N rounds (e.g. 3,000) of the new tick plus `events.ProcessEvents()`.
6. Assert: the count of rounds with an idle command is within a band around 0.10 x 0.99 x 2/3 x N (or equal to a recorded count, since the draw sequence is fixed with one NPC and one-exit rooms); every move keeps ring distance from home `<= MaxWander + 1` and `WanderCount <= MaxWander + 1`; after `WanderCount` passes 8 the rat paths home and `WanderCount` returns to 0 at home; the emote and `go` inputs appear with the 1:1 ratio.

## 5. `OnBehaviourPerceive`

- Declared `actors/controller.go:65` (`util.Hook[BehaviourPerception]`), fired from `BehaviourController.Perceive` (`:69-74`) for every non-Interface perception. `messaging.deliver` calls `ctrl.Perceive` synchronously (`messaging.go:68-83`), holding no messaging lock.
- **Subscribers: tests only.** `actors/controller_test.go:16` and `messaging/messaging_test.go:465`. Grep `OnBehaviourPerceive` over `*.go` (it found the declaration, so it could find more). No production reaction exists.
- `util.Hook.Fire` holds `RLock` while handlers run (`hook.go:42-49`); there is no `Unregister`, so test registrations stay for the life of the test binary. A handler that sends a message nests `Send`, then `Perceive`, then `Fire` (re-entrant `RLock`; deadlocks only if a `Register` is waiting at that moment, Go `RWMutex` semantics). Two reacting NPCs recurse without bound.
- **Where a reaction can be queued without recursion**, both existing: (a) `mob.Command(...)` queues an `events.Input` (`mobs.go:253-284`), run later by `world.HandleInputEvents`, strict FIFO per NPC; (b) a pending list on the `BehaviourController` drained by its `Tick`. (a) already exists and needs no new state; (b) puts reactions on the round clock. Option (a) also runs inside `DoListeners`, which holds `listenerLock` (`listeners.go:153-154`): a reaction must not register listeners.
- **Import cycle constraint:** `rooms` imports `actors`, `behaviour` imports `rooms` (`go list` output). `actors` cannot import `behaviour`, so `BehaviourController.Tick` cannot call `behaviour.RunIdle` directly. It needs a hook or func variable in `actors` that `behaviour`/`hooks` fills (the pattern `OnBehaviourPerceive` and `wire.go` already use), or the tick logic lives in `hooks`.
- **Registry lock constraint:** `actors.Each` holds the registry read lock (`registry.go:105-116`) and forbids write paths. Idle can `Rest` an NPC (ephemeral end destroys the instance, firing `OnInstanceDestroyed` to `unbindMob`, a write) and health loss can lead to removal. A tick loop must iterate `actors.All()` (a copy), not `Each`.

## 6. Player-visible changes a move into `BehaviourController.Tick` would cause

Engine (slice 6 territory):

1. **Timing against combat.** Idle now runs after `DoCombat`/`AutoHeal` (`hooks.go:26,30,31`). If the one actor tick sits where `UserRoundTick`/`MobRoundTick` are (`:20-21`, before combat), an attacked NPC may idle or wander once before its first defence, `PreventIdle` set by combat in round N suppresses round N+1 instead of round N, and new aggro is seen one round later.
2. **Same-round order.** `MobIdle` is queued, so idle commands are issued after every other event already queued that pass; inline calls in `Tick` issue them earlier. Visible only as line order within one round.
3. **Following NPCs.** The follow module stops idle by cancelling `MobIdle` (`follow.go:68, 246-254`). Deleting the event without a replacement makes a following NPC also emote and wander away from its leader.
4. **Module API.** `MobIdle` is documented for module authors (`docs-modules.html:1123, 1557-1583`); the docs page needs the replacement.
5. **Boredom.** Dead today (no increment since Phase 2.3). Keeping the branch changes nothing; restoring an increment would make NPCs out of players' sight rest after 250 rounds (about 17 minutes) once more than 100 NPCs are loaded.
6. **NPC order.** Both today's loop and `actors.All()` give a per-run random order; no change players can see.
7. **Pre-existing bugs that move with the code** (decide whether slice 6 keeps them):
   - **Stuck silent at home.** `WanderCount > MaxWander` while at home: `HandleIdleMobs:38-42` calls `PathTo(home)`, which returns with no path because the target is the current room (`behaviour.go:91-93`), and the reset (`NewRound_IdleMobs.go:127-129`) only runs when a path ends. The NPC never idles again until restart or copyover restore. Reachable when a queued `go` fails (out of AP, `go.go:56-65`; the count rises on queueing, `behaviour.go:53`) or when a follower is led back home (`follow.go:310`).
   - `MaxWander` counts moves, not rooms from home (comment `mobs.go:70` says "Max rooms to wander from home"). Admin wizard prompt says `-1 = none. 0 = unlimted` (`admin.mob.go:364`); the code means `0` never wanders and `-1` wanders forever (`behaviour.go:32,41`).
   - `MaxMobBoredom` yaml says 0 means never; `config.memory.go:13-14` makes it 150.
   - `LastIdleCommand` is unused (written only by `saved.go:58`).

Rules, not engine (Phase 5; flag, do not design in slice 6):

- `ActivityLevel` as the idle chance and as the combat-command chance (`NewRound_DoCombat.go:273`).
- "mumbles about losing their quarry" and dropping aggro when the target leaves (`NewRound_IdleMobs.go:67-78`): a perception/pursuit rule.
- `PreventIdle` on being attacked, and turning toward a remote attacker (`NewRound_DoCombat.go:230-245`): a reaction rule.
- `home-impossible` health drain of 10% per call (`behaviour.go:66-74`) and the `lost` adjective: an NPC with `MaxWander 0` that cannot reach home loses health each idle round and dies about ten rounds later (`MobRoundTick` queues `suicide` at health <= 0, `NewRound_MobRoundTick.go:103-106`).
- Charm stopping idle and wander (`MobIdle_HandleIdleMobs.go:34-36`, `behaviour.go:26-28`), and follow counting as wander (`follow.go:308-311`).
- The 1-in-100 empty idle command (`mobs.go:486-488`).

## Open questions (decisions for Calabe)

1. **Where the one tick runs in the round.** Before combat (where the two round ticks are now) or after (where idle is now). Effect: before means an attacked NPC can wander off or emote once before it fights back, and "losing quarry" fires a round later; after keeps today's feel. Splitting the tick into two calls keeps today's order but is not "one tick".
2. **How a following NPC stops idling once `MobIdle` is gone.** A hook the follow module registers on the behaviour tick, or a "following" question the tick asks. Effect if missed: NPC followers wander off their leader.
3. **Dead boredom despawn.** Delete `BoredomCounter`, `MaxMobBoredom`, `MobUnloadThreshold` and the branch, or restore the increment. Effect: delete changes nothing players see today; restore brings back NPCs vanishing after about 17 minutes unwatched when over 100 are loaded.
4. **Fix the stuck-at-home bug in slice 6 or carry it.** Fix (for example reset `WanderCount` when `PathTo(home)` finds the NPC already home). Effect: NPCs that hit their wander limit at home keep emoting and wandering instead of going silent for good.
5. **Where NPC reactions to perceptions are queued.** Through the NPC's input queue (`mob.Command`), or a pending list drained by `Tick`. Effect: input queue reacts within the same few turns (fast, ordered with the NPC's other commands); `Tick` reacts at the next round (up to 4 s later). No reactions exist yet, so this only sets the pattern.
6. **`MaxWander` meaning.** Keep "moves since last home" or make it "rooms from home". Effect: with moves, the rat in the two-room Startland turns home after 9 moves however close it is; with distance, a rat bouncing between two rooms never hits a limit of 8. Also whether to fix the admin wizard prompt that has `-1`/`0` backwards.

# Phase 4b slice 6, inventory I: the world loop, input ordering, the blast radius

Branch `strip/phase-4b-6-one-tick` at `3d7f7e47` (working tree clean at session start). Read-only. Every
fact below was read from source in this session; file:line refers to that tree.

## Read 40 of 40 files

Full reads (every line, chunks of at most 400 lines):

1. `world.go` 1-400, 400-663 (663 lines)
2. `world_test.go` 1-262
3. `internal/events/events.go` 1-244
4. `internal/events/queue.go` 1-113
5. `internal/events/const.go` 1-27
6. `internal/events/listeners.go` 1-192
7. `internal/events/eventtypes.go` 1-361
8. `internal/hooks/hooks.go` 1-73
9. `internal/hooks/NewRound_UserRoundTick.go` 1-132
10. `internal/hooks/NewRound_MobRoundTick.go` 1-111
11. `internal/hooks/NewRound_IdleMobs.go` 1-140
12. `internal/hooks/MobIdle_HandleIdleMobs.go` 1-53
13. `internal/hooks/NewTurn_ActionPoints.go` 1-23
14. `internal/behaviour/behaviour.go` 1-160
15. `internal/behaviour/behaviour_test.go` 1-200
16. `internal/actors/controller.go` 1-74
17. `internal/actors/actor.go` 1-28
18. `internal/actors/registry.go` 1-247 (in three reads: 1-90, 90-129, 128-247)
19. `internal/actors/wire.go` 1-58
20. `modules/follow/follow.go` 1-400, 400-525
21. `internal/mobs/saved.go` 1-93
22. `internal/usercommands/usercommands.go` 1-394
23. `internal/usercommands/admin.command.go` 1-71
24. `internal/guards/guards_test.go` 1-274
25. `internal/guards/commands_test.go` 1-210
26. `internal/guards/combat_test.go` 1-103
27. `internal/guards/messaging_test.go` 1-70
28. `cmd/smoke/main.go` 1-203
29. `cmd/smoke/script.go` 1-305
30. `cmd/smoke/fight.go` 1-189
31. `cmd/smoke/admin.go` 1-139
32. `cmd/smoke/copyover.go` 1-162
33. `cmd/smoke/recorder.go` 1-41
34. `cmd/smoke/server.go` 1-232
35. `cmd/smoke/client.go` 1-251
36. `internal/util/copyover.go` (whole file, printed with cat)
37. `_datafiles/world/empty/mobs/startland/1-rat.yaml` (whole file, cat)

Partial reads (only these ranges):

38. Spec `2026-10-27-phase-4b-npcs-are-players-design.md` 185-345 (sections 3, 4 and all of "Notes for later
    slices"). Lines 1-184 (facts table, decisions, sections 1-2) NOT read.
39. `internal/mobs/mobs.go` 40-339 and 470-549. Other lines not read.
40. Excerpts printed with sed/grep -A (only these lines read): `internal/accounts/userrecord.go` 277-341,
    462-464, 612-625; `internal/hooks/Buff_ApplyBuffs.go` 58-75; `internal/hooks/PlayerSpawn_HandleJoin.go`
    70-82; `internal/usercommands/go.go` 215-250; `internal/hooks/NewRound_DoCombat.go` 225-245, 280-295;
    `internal/util/util.go` 95-125, 170-200; `internal/rooms/rooms.go` 868-900, 2030-2099;
    `modules/gmcp/gmcp.Gametime.go` 20-45; `internal/hooks/NewRound_HandleRespawns.go` 12-26;
    `_datafiles/config.yaml` 150-175; `internal/configs/config.timing.go` 32-35, 39-41, 63-65; grep lines
    only from `NewRound_AutoHeal.go`, `NewRound_InactivePlayers.go`, `internal/copyover/copyover.go`,
    `main.go`, `cmd/smoke/*_test.go` (test names only), `cmd/smoke/testdata/pre-5/bot0.txt` (rat lines).

Skipped explicitly: `internal/events/queue_test.go`, `eventlogger.go`, `memory.go`, `const_test.go` (only grep
lines); `cmd/smoke/client_test.go`, `copyover_test.go`, `fight_test.go` bodies; `internal/copyover/*` bodies;
`internal/guards/doc.go`.

Note: the brief named only `guards_test.go`; `internal/guards/` also holds `commands_test.go`,
`combat_test.go`, `messaging_test.go` (all read).

---

## 1. The world loop (`world.go`)

| What | Where | Fact |
|---|---|---|
| Timers | `world.go:308-312` | `roomUpdateTimer` 3 s, `ansiAliasTimer` 4 s, `eventLoopTimer` 1 ms, `turnTimer` `Timing.TurnMs`, `statsTimer` 10 s |
| Event pass | `world.go:368-374` | every 1 ms: `util.LockMud(); w.EventLoop(); util.UnlockMud()` |
| `EventLoop` | `world.go:658-663` | `events.ProcessEvents()` then `clear(w.inputTracker)`. `ignoreInput` is NOT cleared here |
| Turn tick | `world.go:376-393` | under `LockMud`: `turnCt := util.IncrementTurnCount()`; queue `events.NewTurn{TurnNumber, TimeNow}` (383) |
| Round tick | `world.go:386-391` | `if turnCt%uint64(c.Timing.TurnsPerRound()) == 0` then `util.IncrementRoundCount()` and queue `events.NewRound{RoundNumber, TimeNow}` (390). This is the only producer of `NewRound` outside tests |
| Player input | `world.go:440-471` | `InputWorker` goroutine: `actors.ByUserId(wi.FromId)` then `events.AddToQueue(events.Input{ActorId, InputText, ReadyTurn: util.GetTurnCount()})` (463-467). It does NOT take `LockMud`, so player Input lands in the heap between or during passes |
| Input listener | `world.go:66` | `events.RegisterListener(events.Input{}, w.HandleInputEvents)` registered in `NewWorld` (main.go:68, before `hooks.RegisterListeners` at main.go:204) |
| `processInput` | `world.go:477-599` | one path for every actor; account-only steps guarded by `user != nil`. Macros queue `Input` with `readyTurn` incremented per part (523-536). NPC unknown command is logged (566-577) |

Timing (stock config, `_datafiles/config.yaml:150-153`): `TurnMs: 50`, `RoundSeconds: 4`. `turnsPerRound =
RoundSeconds*1000/TurnMs` (`config.timing.go:32`) = 80 turns per round; `SecondsToTurns(1)` = 20 turns
(`config.timing.go:63-65`).

### Event queue semantics (`internal/events`)

- One global min-heap ordered by `(priority, order)`, lower first; `order` is a global insert counter
  (`events.go:65-70`, `111-122`). Default priority 0; no caller in this inventory passes a priority.
- `ProcessEvents` (`events.go:176-233`): first re-adds every requeued event from the previous pass with a
  fresh `order` (`197-200`), then pops until the heap is empty, so events queued by listeners during the
  pass run in the same pass (`203-230`). A listener returning `CancelAndRequeue` sends the event to
  `requeues`, which only re-enter at the top of the NEXT pass (`224-226`, `159-171`).
- Consequence: an event queued between passes (player input from `InputWorker`, `NewTurn`/`NewRound` from
  the turn timer) has a lower `order` than every requeued event re-added at the top of the next pass, so it
  pops first.
- Listener order (`listeners.go:56-119`): `First` prepends, `Last` appends as final, else appended before
  finals. `DoListeners` stops at the first non-`Continue` result (`174-181`). Return codes `Continue`,
  `Cancel`, `CancelAndRequeue` (`listeners.go:42-46`).
- Unique events (`UniqueID()`) dedupe in the queue (`events.go:100-109`); `Input`, `NewRound`, `MobIdle`
  are not unique (`eventtypes.go:51-58`, `120-125`, `136-140`).
- Event types: `Input{ActorId ids.ID; InputText string; ReadyTurn uint64; Flags EventFlag}`
  (`eventtypes.go:51-58`); `NewRound{RoundNumber uint64; TimeNow time.Time}` (`120-125`); `NewTurn`
  (`128-133`); `MobIdle{ActorId ids.ID}` (`136-140`).
- Flags (`const.go:21-26`): `CmdSecretly`, `CmdIsRequeue`, `CmdBlockInput`, `CmdUnBlockInput`,
  `CmdBlockInputUntilComplete` = both.

## 2. Input ordering: the two modes, keyed on `Account()`

All in `World.HandleInputEvents`, `world.go:73-146`. The key is `user := actor.Controller().Account()`
(`world.go:94`). State: `w.inputTracker map[ids.ID]struct{}` (`world.go:48`, cleared after each pass at
`world.go:662`) and `w.ignoreInput map[ids.ID]uint64` (`world.go:40`).

An input whose actor is gone is dropped with a debug log (`world.go:81-87`).

### Mode 1: no account (NPC), strict FIFO (`world.go:95-114`, "Decision 13" comment at 90-93)

1. If the actor is already in `inputTracker` this pass: `CancelAndRequeue` (`98-100`). Checked first, so
   even a `ReadyTurn == 0` input waits behind a pending delayed one (comment `96-97`).
2. `ReadyTurn == 0`: run now, do not mark the tracker (`102-105`).
3. Otherwise mark the tracker BEFORE the ready check (`108`), then if `ReadyTurn > turnCt` requeue
   (`109-111`), else run (`112-113`).

Effect: the first NPC input popped in a pass either runs or blocks every later input for that NPC until
the next pass; a not-yet-ready input re-blocks them every pass until it runs. Tested by
`TestNPCDelayedCommandHoldsBackTheNext` (`world_test.go:195-215`).

Why FIFO holds across passes for NPCs (derived from the queue rules above, not tested directly): requeued
inputs keep their relative order (the `requeues` slice is replayed in order, `events.go:197-200`). Every NPC
`Input` producer found runs inside a pass (listeners and commands; see section 4), so a new NPC input always
gets a higher `order` than the NPC's requeued ones. An NPC `Input` queued from outside a pass (another
goroutine) would pop before the requeued ones and could overtake. None found: no `events.Input{` in
`internal/web` (repo-wide grep of `events\.Input\{` listed every producer, section 4).

### Mode 2: account (player), ready commands may overtake (`world.go:116-145`)

1. `ReadyTurn == 0`: unblock if flagged, run now, bypass and do not count toward the limit (`117-121`).
2. Already in the tracker this pass: `CancelAndRequeue` (`124-126`). One queued command per actor per
   event-loop pass (not per turn; comment `92-93`).
3. Not ready (`ReadyTurn > turnCt`): the tracker is NOT set, so a later ready command for the same actor
   can run in this pass (comment `128-129`). If the input has `CmdBlockInput`, record the turn the wait
   began in `ignoreInput[actor.Id]` (`132-136`). Requeue (`137`).
4. Ready: `unblockIfDone` (`141`), run (`142`), mark the tracker (`143`).

`unblockIfDone` (`world.go:150-158`): on a `CmdUnBlockInput` input whose block was recorded, delete the
`ignoreInput` entry and `user.UnblockInput()`. Blocked input is enforced in `TryCommand`: a command not
`AllowedWhenDowned` returns handled-and-silent while `user.InputBlocked()` (`usercommands.go:320-323`).
Tested by `TestPlayerReadyCommandOvertakesADelayedOne` (`world_test.go:217-231`) and
`TestBlockInputUntilCompleteUnblocksAfterAOneSecondWait` (`world_test.go:238-262`).

### Command delays: how `ReadyTurn` is set

| Producer | Where | ReadyTurn |
|---|---|---|
| `Mob.Command(text, waitSeconds...)` | `mobs.go:253-284` | per-NPC monotonic: starts at `max(now, m.lastCommandTurn)` (`255-265`), adds `SecondsToTurns(1)*wait` (`267-269`), each `;` part one turn later (`271-281`), stores back in `m.lastCommandTurn` (`274`, unexported field `mobs.go:78`) |
| `Account.Command(text, wait...)` | `userrecord.go:277-290` | `now + SecondsToTurns(1)*wait`; no per-account monotonic counter |
| `Account.CommandFlagged` | `userrecord.go:304-322` | as above; calls `BlockInput()` first when `CmdBlockInput` |
| `usercommands.queue` / `queueFlagged` | `usercommands.go:76-109` | as `Account.Command`; any actor; `queueFlagged` blocks an account's input when flagged |
| `InputWorker` | `world.go:463-467` | `now` (not 0, so it counts toward the per-pass limit) |
| macros | `world.go:523-536` | `now`, `now+1`, ... |
| admin `command <target> a;b` | `admin.command.go:46-68` | `now`, `now+1`, ...; bypasses `Mob.lastCommandTurn` |
| `Buff_ApplyBuffs` TriggerNow death | `Buff_ApplyBuffs.go:64-70` | 0 (`suicide`) |
| `OnLoginCommands` | `PlayerSpawn_HandleJoin.go:76-80` | 0 |

The waypoint `noop` delay: `NewRound_IdleMobs.go:112-117` issues `mob.Command(exitName)` then, at a
waypoint, `mob.Command("noop", 1)`; because `Mob.Command` is monotonic, `noop` is scheduled 20 turns after the
move and, under mode 1, holds back every later NPC command until it runs. `noop` is a registered command
that does nothing (`usercommands.go:163`, `noop.go:10-12`).

How a delayed command requeues: the listener returns `CancelAndRequeue`, `ProcessEvents` puts it in
`requeues` (`events.go:224-226`), and it is re-added with a new `order` at the start of the next pass
(~1 ms later, `world.go:370`). There is no per-actor queue anywhere; an actor's pending commands are just
`Input` events circulating through the global heap.

Unverified oddity, outside slice 6: `WimpyCheck` calls `u.Command("flee", -1)` (`userrecord.go:621`), so
`uint64(float64(20) * -1)` is added to `readyTurn` (`userrecord.go:281`). Converting a negative float to
`uint64` is implementation-defined in Go; not tested here.

## 3. The router (`internal/usercommands/usercommands.go`) and ticks/turns/delays

- `TryCommand(cmd, rest, actorId, flags)` (`219-369`) has no turn or round logic of its own. It resolves
  exit names to `go` (`236-239`), personal aliases for accounts only (`242-245`), keyword aliases
  (`264-280`), cancels `cancel-on-action` buffs (`287`), enforces downed and input-blocked
  (`312-324`), admin role (`327`, NPCs never admin), then emotes (`352-355`), then a `go` attempt
  (`363-365`).
- Delays enter only through `queue`/`queueFlagged` (`76-109`). The one flagged producer: `go.go:152-154`
  queues the move again with `CmdIsRequeue|CmdBlockInputUntilComplete` and a 1 s wait when the exit has an
  `ExitMessage`.
- `go.go:229` (party members follow the leader with `partyUser.Command(rest)`) and `go.go:247` (charmed
  mobs follow with `mob.Command(rest)`) queue moves for other actors.

## 4. Blast radius (non-`.md` files; every hit)

Grep: `UserRoundTick|MobRoundTick|IdleMobs|HandleIdleMobs|MobIdle|events\.NewRound|RoundTick|BehaviourController|AccountController|OnBehaviourPerceive|PreventIdle|LastIdleCommand|BoredomCounter|WanderCount`
over the whole repo, glob `!*.md`. The same grep found hits in every target file, so it could find them.

### 4.1 The four things slice 6 deletes

| Hit | What it does |
|---|---|
| `hooks.go:20` | registers `UserRoundTick` on `NewRound` |
| `hooks.go:21` | registers `MobRoundTick` on `NewRound` |
| `hooks.go:31` | registers `IdleMobs` on `NewRound` (after `AutoHeal`) |
| `hooks.go:32` | registers `HandleIdleMobs` on `MobIdle` |
| `NewRound_UserRoundTick.go:19-132` | for each room with players (`rooms.GetRoomsWithPlayers`, 23): `room.RoundTick()` (27); room/zone idle messages, 5% (20% in room -1) chance, no immediate repeat (29-75); then for each player in the room: `Cooldowns.RoundTick()` (85), alignment decay every `AlignmentDecayRounds` with text and `AlignmentChanged` (88-101), charm `RoundsRemaining--` (103-105), `Buffs.Trigger()` then `BuffsTriggered` (107-115), `Validate()` and `CharacterVitalsChanged` if HP/MP moved (118-123) |
| `NewRound_MobRoundTick.go:54-111` | `mobs.ReduceHostility()` (59, global); for every mob instance (`mobs.GetAllMobIds`, 64): `Cooldowns.RoundTick()` (73), charm `RoundsRemaining--` (75-77), `Buffs.Trigger()` then `BuffsTriggered` (79-87), charm expiry: `RemoveCharm`, charmer's `TrackCharmed(false)`, `runCharmExpiredCommand` (90-98), `Validate()` (101), `Health <= 0` then `mob.Command("suicide")` (103-106) |
| `NewRound_MobRoundTick.go:20-52` | `runCharmExpiredCommand`: `rest` part calls `behaviour.Rest` and stops; other parts run via `TryCommand` now if resting, else `mob.Command` (tested: `charmexpiry_test.go:70,84`) |
| `NewRound_IdleMobs.go:23-140` | per mob: boredom unload when over `MobUnloadThreshold` and `BoredomCounter >= MaxMobBoredom` (`behaviour.Rest`, 46-57); `PreventIdle` one-round skip (60-63); aggro: lost-quarry emote and clear aggro when target gone or elsewhere, else skip (67-78); path walking: re-plan if displaced (82-105), step via `mob.Command(exit)` plus waypoint `noop` 1 s (107-118), clear path and reset `WanderCount` at home (125-129); otherwise queue `events.MobIdle{ActorId}` (133); timing via `util.TrackTime("IdleMobs()")` (137) |
| `MobIdle_HandleIdleMobs.go:14-53` | mob gone: `Cancel` (18-21); `MaxWander == 0` and away from home and not charmed: `PathTo("home")` (28-32); charmed: stop (34-36); `MaxWander > -1 && WanderCount > MaxWander`: `PathTo("home")` (38-43); `util.Rand(100) < ActivityLevel` then `GetIdleCommand()` then `behaviour.RunIdle` (46-50) |
| `eventtypes.go:135-140` | the `MobIdle` event type |

### 4.2 Other listeners on `MobIdle` / `NewRound`

| Hit | What it does |
|---|---|
| `modules/follow/follow.go:68` | `RegisterListener(events.MobIdle{}, f.idleMobHandler, events.First)` |
| `follow.go:246-254` | `idleMobHandler`: returns `Cancel` when the NPC is following someone, which suppresses ALL of `HandleIdleMobs` for it (no idle command, no wander, no path-home). Slice 6 must give follow another way to veto idle; no `MobIdle` means this listener has nothing to hang on |
| `follow.go:70`, `182-229` | `onNewRound`: expires follows at their cutoff round (`followLimits`) and tells the people involved. Not an actor tick; unaffected unless `NewRound` itself changes |
| `follow.go:305-311` | NPC follower moves via `mob.Command(exit, .25)` (5 turns) and `mob.WanderCount++` ("Count follows as wandering") |
| `follow.go:489-492` | `follow` on an NPC clears `mob.Path` |
| `modules/gmcp/gmcp.Gametime.go:25`, `35` | GMCP game-time push on each `NewRound`; unaffected |
| `hooks.go:17-30` | other `NewRound` listeners: `InactivePlayers` (players only), `UpdateZoneMutators`, `CheckNewDay`, `HandleRespawns` (`room.Prepare(false)` for rooms with players, `NewRound_HandleRespawns.go:14-23`), `DoCombat`, `AutoHeal` (online accounts only, `NewRound_AutoHeal.go:28-29`) |
| `NewRound_UpdateZoneMutators.go:13`, `NewRound_InactivePlayers.go:16`, `NewRound_CheckNewDay.go:15`, `NewRound_AutoHeal.go:19` | each casts `e.(events.NewRound)` |

Listener order on `NewRound` today: module `init()` listeners register before `main.go:204`
`hooks.RegisterListeners()`, so `follow.onNewRound` and the GMCP handler run first, then `InactivePlayers`,
`UpdateZoneMutators`, `CheckNewDay`, `UserRoundTick`, `MobRoundTick`, `HandleRespawns`, `DoCombat`,
`AutoHeal`, `IdleMobs` (`hooks.go:17-31`). Ordering facts slice 6 must keep or rule on:
- `DoCombat` sets `defMob.PreventIdle = true` (`NewRound_DoCombat.go:234`) before `IdleMobs` reads and
  clears it in the same round (`NewRound_IdleMobs.go:60-63`).
- `MobRoundTick` (Health <= 0 then suicide) runs before `IdleMobs`, so a `PathTo("home")` health loss
  (`behaviour.go:66-74`, 10% per call when home is impossible) kills on the next round.
- `MobIdle` events queued by `IdleMobs` run later in the same pass (queued inside a pass).

### 4.3 State fields on `Mob`

| Field | Decl | Writes / reads |
|---|---|---|
| `ActivityLevel` | `mobs.go:58` (yaml `activitylevel`, "1-100%") | read `MobIdle_HandleIdleMobs.go:46`; clamped by `Mob.Validate`: `<1` becomes 10, `>100` becomes 100 (`mobs.go:531-535`) |
| `MaxWander` | `mobs.go:70` (yaml `maxwander`) | read `behaviour.go:32-43`, `MobIdle_HandleIdleMobs.go:28,38`, `admin.locate.go:153` |
| `WanderCount` | `mobs.go:71` (`yaml:"-"`) | `++` `behaviour.go:53`, `follow.go:310`; reset `NewRound_IdleMobs.go:128`, `saved.go:58`; read `behaviour.go:33`, `MobIdle_HandleIdleMobs.go:38`, `admin.locate.go:153` |
| `PreventIdle` | `mobs.go:72` | set `NewRound_DoCombat.go:234`; cleared `NewRound_IdleMobs.go:61`, `saved.go:60` |
| `BoredomCounter` | `mobs.go:63` | reset `rooms.go:2071` (room with players, in `Room.RoundTick`), `NewRound_IdleMobs.go:53`, `saved.go:58`; read `NewRound_IdleMobs.go:46,49`. **Never incremented anywhere**: no `++`/`+=` in the repo (grep `Boredom` over all non-md files). Upstream incremented it in `internal/mobcommands/lookfortrouble.go:160-161` (`git grep` on `upstream/master`), which this fork stripped; master `c4314b7f` already had no increment. So boredom unload is dead code today |
| `LastIdleCommand` | `mobs.go:62` | only reset (`saved.go:58`) and in `saved_test.go`; never read, never set otherwise (`git grep HEAD`). Dead field |
| `Path` | `mobs.go:76` | walked by `IdleMobs`; set by `behaviour.PathTo`; cleared by `follow.go:491` |
| `lastCommandTurn` | `mobs.go:78` | `Mob.Command` only; reset on restore `saved.go:61` |

### 4.4 `Controller` / `BehaviourController` / `OnBehaviourPerceive`

| Hit | What it does |
|---|---|
| `controller.go:14-20` | `Controller` interface: `Account()`, `Perceive(p)`. Comment line 12-13: "Slice 6 adds Tick." |
| `controller.go:45-55` | `BehaviourController{mob *mobs.Mob}`; comment: "slice 6 moves the mob's runtime behaviour state here and runs it from Tick" |
| `controller.go:63-74` | `OnBehaviourPerceive util.Hook[BehaviourPerception]` fired from `Perceive`; no production listener (grep: only `controller_test.go:16` and `messaging_test.go:465`) |
| `wire.go:46-52` | every mob instance gets `Spawn(&m.Character, NewBehaviourController(m))` via `mobs.OnInstanceCreated` |
| `registry.go:29-34` | `kindOf` keys player vs NPC on `Account()` |
| `guards_test.go:81` | `TestNoControllerTypeAssertions` names `ConnectionController`, `BehaviourController` |
| `messaging_test.go:25-26` | test fake `recorder` implements `Controller`; adding `Tick` to the interface breaks it until it gains a `Tick` |
| `controller_test.go:13-29`, `registry_test.go:47-155`, `rooms_test.go:69`, `messaging_test.go:452-461` | tests that construct `NewBehaviourController` |
| `AccountController` | no hits anywhere. The grep pattern matched `BehaviourController` in the same files, so it was capable; the name does not exist (the account side is `ConnectionController`, `controller.go:25`) |
| `Controller()` call sites | 67 non-test call sites in 57 files (`grep "Controller()"`). They call `Account()`; adding a method to the interface does not touch them |

### 4.5 `RoundTick` methods (name collision, not the hooks)

| Hit | What it does |
|---|---|
| `rooms.go:2038-2080` `Room.RoundTick` | mutator update and buffs (2047-2054), spawn-slot settle and save (2060-2064), boredom reset when players present (2068-2074), corpse decay (2079). Called only from `NewRound_UserRoundTick.go:27`, so it runs only for rooms with players. It is a room tick; one actor tick does not absorb it, but deleting `UserRoundTick` needs a new home for it |
| `characters/cooldowns.go:9` `Cooldowns.RoundTick` | called `NewRound_UserRoundTick.go:85`, `NewRound_MobRoundTick.go:73`; tested `cooldowns_test.go` |
| `accounts/userrecord.go:462-464` `Account.RoundTick` | empty body; no callers (grep `\.RoundTick\(\)` found the room and cooldown calls, none on an account) |
| `rooms/npcslots_test.go:37,48,81` | tests call `home.RoundTick()` |

### 4.6 `events.NewRound` in tests

`mobflee_test.go:83,105` and `combat_golden_test.go:423` call `DoCombat(events.NewRound{})` directly. No test
calls `UserRoundTick`, `MobRoundTick`, `IdleMobs` or `HandleIdleMobs` (grep over `*_test.go`): the code
slice 6 replaces has no direct test coverage today.

### 4.7 Non-Go hits

- `_datafiles/html/admin/docs-modules.html:1123` documents `MobIdle` ("Use `events.First` to intercept
  before built-in AI"); `:1557-1583` shows the follow module's `MobIdle` pattern as the way modules drive
  mob behaviour. Stale once `MobIdle` goes.
- `_datafiles/config.yaml:94-102` and admin `config.html:211-212`, `config-wizard.html:1344-1352,1652`:
  `MaxMobBoredom` / `MobUnloadThreshold` (dead while boredom never increments).
- `.md` hits are historical plans/inventories in `docs/completed/` plus the spec; not live docs.

### 4.8 Every `Input` producer (who feeds the queue the tick will feed)

Grep `events\.Input\{|\.Command\(|CommandFlagged\(|queueFlagged\(|\bqueue\(` over non-test Go:
`world.go:463,529,582`; `behaviour.go:54,158`; `follow.go:297,307,314,322,326`;
`userrecord.go:284,315,621`; `Buff_ApplyBuffs.go:66`; `NewRound_AutoHeal.go:48`;
`NewRound_DoCombat.go:238,240,288,291,319,464,497,516`; `NewRound_IdleMobs.go:72,112,116`;
`NewRound_MobRoundTick.go:49,105`; `PlayerSpawn_HandleJoin.go:76,93`; `mobs.go:276`;
`admin.build.go:51,155`; `admin.command.go:50,59`; `attack.go:206,260`; `default.go:17`; `give.go:205,207`;
`go.go:154,229,247`; `usercommands.go:70,82,103` (plus `exec.Command` false positives in
`cmd/smoke/server.go:110`, `internal/copyover/copyover.go:152`). All except `world.go:463` (`InputWorker`)
run inside an event pass.

## 5. Structural constraints for `Controller.Tick` (found while reading; affect the design)

1. **Import cycle.** `internal/actors` imports `accounts, characters, ids, mobs, mudlog, perception, util`;
   `internal/behaviour` imports `mapper, mobs, rooms`; `internal/rooms` imports `internal/actors`
   (`go list`, and `rooms.go:2089` calls `actors.ByUserId`). So `BehaviourController.Tick` in package
   `actors` cannot call `behaviour.RunIdle`/`PathTo`/`Rest` directly (actors -> behaviour -> rooms ->
   actors). It needs a hook or an injected function, as `OnBehaviourPerceive` already is
   (`controller.go:63-65`).
2. **Registry lock.** `actors.Each` holds the registry read lock while it calls `fn` and forbids write paths
   (`registry.go:105-115`). A tick that rests or kills an NPC reaches `mobs.Rest`/`DestroyInstance` then
   `OnInstanceDestroyed` then `unbindMob`, which takes `mu.Lock()` (`registry.go:216-218`, `wire.go:54-58`):
   deadlock under `Each`. `actors.All()` returns a sorted copy (`registry.go:94-103`) and is safe to iterate
   while removing. `Actor.Controller()` also takes the read lock (`actor.go:24-28`).
3. **Determinism for the seeded test.** `util.SeedRandForTest` makes `util.Rand` repeatable
   (`util.go:181-194`), but `mobs.GetAllMobIds` ranges a map (`mobs.go:228-235`) and
   `Room.GetRandomExit` builds and ranges a map (`rooms.go:868-900+`), so with more than one NPC, or more
   than one exit, draw order or exit choice is not repeatable from the seed alone. `actors.All()` is sorted
   by ID text, which would give a fixed order.
4. **Odds today, for the rat.** `1-rat.yaml`: `activitylevel: 10`, `maxwander: 8`, idle commands
   `emote wiggles its nose`, `wander`, `''`. Per round: `Rand(100) < 10` (10%), then `GetIdleCommand`
   returns `''` 1 time in 100 (`mobs.go:486-488`) else one of the three uniformly (`mobs.go:491-493`). So
   about 3.3% per round to wander, 3.3% to emote. `Wander` only counts a move inside the NPC's zone
   (`behaviour.go:50-55`); past `MaxWander` it plans home (`behaviour.go:31-37`, and
   `MobIdle_HandleIdleMobs.go:38-43`). The rats spawn in startland room 2 (`rooms/startland/2.yaml:11-18`).

## 6. Guards (`internal/guards`)

Existing guards slice 6 must keep green:
- `TestCharacterPointerIsNotReassigned` (`guards_test.go:53-75`): a tick must copy into `*Character`,
  never assign `.Character` (allowlist: three files).
- `TestNoControllerTypeAssertions` (`guards_test.go:80-113`): no assertion or type switch on
  `actors.ConnectionController`/`BehaviourController` outside `internal/actors`. A tick dispatched through
  the interface satisfies it.
- `TestIdsAreMadeOnlyAtCreation`, `TestNoMobInstanceIds`, `TestAccountsGoOnlineOneWay`,
  `TestItemsAreInstantiatedOnlyFromTemplates`, `TestNoNumericShorthandIds` (`guards_test.go`): not touched
  by tick code unless it formats IDs or makes them.
- `TestNoSecondCommandSet`, `TestNoAddMobCommand` (`commands_test.go:15-49`).
- `TestNoDroppedVerbsQueued` (`commands_test.go:56-109`): no `.Command(`/`queue(` literal may start with
  `wander`, `pathto`, `rest`, `despawn`, `sayto`, `saytoonly`, `replyto`. `noop` is allowed (waypoint delay).
- `TestOneCombatPath` (`combat_test.go:12-55`) has a positive control on `func DoCombat` in
  `internal/hooks/NewRound_DoCombat.go`; renaming or moving that file breaks it.
- `TestOnlyMessagingWritesToConnections` (`messaging_test.go:16-70`): `world.go` is on the shrinking
  allowlist (`:36`); any new file that calls `connections.SendTo` fails. `room.SendText` (idle messages,
  `NewRound_UserRoundTick.go:69`) is not a `connections.*` call and is not caught.

Candidates to add (pattern: banned identifiers plus a positive control, as `TestOneCombatPath` does):
- Ban the identifiers `UserRoundTick`, `MobRoundTick`, `IdleMobs`, `HandleIdleMobs`, `MobIdle` in
  non-test Go; positive control: `Controller` has a `Tick` method and the one round tick hook exists.
- Optionally ban `PreventIdle`/`BoredomCounter`/`LastIdleCommand`/`WanderCount` on `mobs.Mob` if slice 6
  moves that state into `BehaviourController` (controller.go:45-46 says it will).

## 7. Smoke (`cmd/smoke`)

What the run does: admin logs in, spawns one copper ring per bot in Town Square (`admin.go:28-56`), spawns a
rat and makes it `say` via the admin `command` command (`admin.go:67-71`; the NPC line goes through the
input path with `ReadyTurn` now, `admin.command.go:46-55`); each bot signs up, gets/drops/gets the item,
walks north and south between Town Square (1) and End of the Line (2), says hello (`script.go:160-194`);
bot 0 fights a rat, bot 1 observes (`script.go:197-202`); quit, relogin, suicide and portal back
(`script.go:208-229`); copyover on Linux launch mode only (`script.go:237-251`, `main.go:112-122`).

`-bots` modes (`main.go:38`, `124-155`): `-bots 1` (default) runs only bot 0, which fights alone;
`-bots 2` adds bot 1 as the fight observer (`newFightCheck`, `fight.go:41-48`); bots 2+ run the plain
script. `-hold` keeps bots online after the script. `-transcript DIR` writes `bot<N>.txt`
(`main.go:138-148`, `recorder.go`). `-copyover` default true, skipped on Windows (`main.go:117-118`).

Where NPC idle and wander touch it:
- The fight retries when the rat is not in the room: on "You attack the darkness!" or "Your target can't be
  found." it waits `fightRetryGap` 2 s and attacks again, up to `fightStepTimeout` 4 min (`fight.go:11-18`,
  `79-113`). Comment: the rat "wandered off". A rat that wandered less would retry less; one that wandered
  far more could make the step slower but it ends when any rat dies.
- `npcSpeaks` relies on a rat in the admin's room answering within `-timeout` (30 s default, `main.go:40`).
  Under the NPC FIFO rule a rat that holds a pending delayed command (waypoint `noop`, 1 s) delays its
  `say` by up to that delay; well inside 30 s.
- The smoke script itself masks nothing; it checks with `strings.Contains` on the screen since the last
  send (`client.go:156-165`), so a wandering rat's "enters"/"leaves" lines do not fail steps.
  `checkObserverText` ignores lines that do not name the fighter (`fight.go:139-158`).
- Masking for text parity is a slice 5/7 transcript concern (spec notes, line 277-279): mask the run tag,
  game-clock times, HP/MP prompts and wandering-NPC lines. The pre-5 baseline has such lines, e.g.
  `cmd/smoke/testdata/pre-5/bot0.txt:412` `>>> rat enters from the north.` and the varying
  `Also here: ...` lines (113, 160, 203, 223). Any change in idle timing or odds changes these lines, which
  the parity mask already expects to drop.
- Copyover check (`copyover.go:41-116`): the new process must log `event=spawned kind=npc` only for IDs the
  old one had. A tick change does not touch this unless it spawns or rests NPCs during the 5 s window.

## 8. Copyover and tick/idle state

- Contributors registered: connections, accounts, util, gametime, reconnect tokens (`main.go:109-113`).
  `util` saves `RoundCount` and `TurnCount` (`internal/util/copyover.go:5-31`), so turn and round numbers
  survive copyover and `ReadyTurn` values stay meaningful.
- No contributor saves the event queue: pending `Input` events (delayed NPC and player commands) and
  `requeues` are lost at copyover. `World.inputTracker` and `World.ignoreInput` are not saved either.
  (Not checked: whether `accounts` copyover restores a session's `inputBlocked`; if it does, a player
  blocked mid-wait could stay blocked because `ignoreInput` is gone. Unverified.)
- NPCs survive through the room sidecars (`SavedNPC`, `saved.go:15-22`). Its doc comment and `Mob()`
  reset "what the NPC was doing": `WanderCount`, `BoredomCounter`, `LastIdleCommand` = 0, `Path` empty,
  `PreventIdle` false, `lastCommandTurn` 0, temp data nil (incl. `home-impossible`) (`saved.go:58-62`).
  Tested: `saved_test.go:33,69-71,75,87-88`. Player-visible consequence today: an NPC away from home
  restarts its wander count after copyover, so it can wander up to `MaxWander` more rooms before heading
  home. If slice 6 moves this state into `BehaviourController`, the same reset applies by construction
  (the controller is rebuilt on `OnInstanceCreated`, `wire.go:46-47`), unless the plan chooses to save it.

## 9. Input ordering in slice 6: options (not decisions)

Today's split is the only per-kind input rule left in `world.go` (`world.go:94-145`). Options:

**A. Keep both rules, keyed on `Account()` (status quo).** No player-visible change. The rule stays
"who drives the body", so a possessed NPC (later) would get the player rule. Cost: two code paths, and the
"strict FIFO" guarantee for NPCs depends on every NPC input being queued inside a pass (section 2), which
nothing enforces.

**B. Strict FIFO for every actor.** Player-visible: a command typed while an earlier delayed one is pending
waits for it. Where players have delayed commands today: an exit with an `ExitMessage` (1 s,
`go.go:152-154`, already blocks non-downed commands via `CmdBlockInput`, but `look`, `say` and other
`AllowedWhenDowned` commands currently run during the wait and would now wait); following someone (each
step `user.Command(exit, .25)`, `follow.go:322`, so anything typed in that 0.25 s waits); macros (one part
per turn, `world.go:523-536`). `ReadyTurn 0` inputs (login commands, `PlayerSpawn_HandleJoin.go:76-80`)
would also wait behind a pending delay. `ignoreInput`/`CmdBlockInput` could then be simpler, since order
alone holds later commands.

**C. Overtaking for every actor (the player rule).** Player-visible: NPCs could act out of order. Because
`Mob.Command` schedules monotonically (`mobs.go:258-274`), most NPC sequences already have rising
`ReadyTurn` and keep their order. What changes: the waypoint `noop` stops holding back later commands
(e.g. a follow step or angry command queued right after a waypoint runs before the 1 s pause ends);
admin `command rat a;b` and other non-`Mob.Command` producers (`admin.command.go`, `Buff_ApplyBuffs.go:66`)
could run before a pending delayed command. The "waypoint `noop` delay lives in the path walking" note
would then need the path walker to own the pause (e.g. not stepping until a turn), not the input queue.

**D. One rule, owned by the actor.** Each actor (or its controller) keeps its own ordered input queue and
`Controller.Tick` or the turn tick drains it; the global heap carries only "actor has input". Player-visible:
depends on which rule (B or C) the queue applies; this option is about where the rule lives. Larger change:
touches every producer in section 4.8 and the copyover gap in section 8.

## Open questions (human decisions, each with the player-visible effect)

1. **Input ordering in slice 6: keep per kind (A), FIFO for all (B), overtake for all (C), or defer to a
   later slice?** Effect: A changes nothing. B makes a player's typed `look`/`say` wait behind a pending
   0.25 s follow step or a 1 s exit-message walk. C lets NPCs skip their waypoint pause and lets admin
   `command` lines jump an NPC's pending delayed command.
2. **Boredom unload is dead (counter never increments since `lookfortrouble` was stripped). Restore it
   in the behaviour tick, or delete it with its config keys `MaxMobBoredom`/`MobUnloadThreshold`?** Effect:
   restoring it makes NPCs in rooms no player has visited for `MaxMobBoredom` rounds (250 in stock config)
   rest out of memory once more than `MobUnloadThreshold` (100) NPCs are loaded; deleting keeps today's
   behaviour (they never unload this way) and removes two admin config fields.
3. **Does a following NPC still skip all idle behaviour?** Today follow's `MobIdle` veto suppresses idle
   emotes, wandering and the path home while following. With `MobIdle` gone, slice 6 must either keep the
   veto (a behaviour query or hook the follow module answers) or change it. Effect if dropped: a following
   NPC also emotes, may wander off between its leader's moves, and walks home when past `MaxWander`
   (follow steps count as wandering, `follow.go:310`).
4. **Where does `Room.RoundTick` go once `UserRoundTick` is deleted, and does it keep running only for
   rooms with players?** Effect: if it stays player-rooms-only, mutator buffs, spawn-slot settling and corpse
   decay in empty rooms stay paused as today; running it for every loaded room would make corpses decay and
   spawn slots refill with nobody watching.
5. **Should the behaviour state (`WanderCount`, `Path`, `PreventIdle`, boredom) survive copyover and restart
   once it moves into the controller?** Effect: today an NPC away from home gets a fresh wander allowance
   and forgets its path after copyover; saving it would keep it heading home.
(Not a decision, for the plan: the dead `LastIdleCommand` field and the empty, uncalled
`Account.RoundTick` can go with slice 6; `LastIdleCommand` is `yaml:"-"`, so the saved NPC form does not
change.)
