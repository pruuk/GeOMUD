# Phase 4b slice 2b: fact inventory (NPC persistence)

Date: 2026-10-27. Source: branch `docs/phase-4b-2b-rulings` at `2784d918`, whose only change against
`master` `9f7666aa` is the design doc (`git diff --stat 9f7666aa HEAD`: 1 file, the spec). Every line
number below is from that tree. Read-only inventory; nothing in the repo was changed.

Read 51 of 51 files:

| # | File | Read |
|---|---|---|
| 1 | `2026-10-27-phase-4b-npcs-are-players-design.md` | partly, 1-135 (facts table and all three Decisions sections; sections 1-4 beyond line 135 not needed) |
| 2 | `internal/rooms/save_and_load.go` | full (454) |
| 3 | `internal/rooms/spawninfo.go` | full |
| 4 | `internal/rooms/rooms.go` | partly: 1-135, 415-489, 515-819, 1084-1247, 1878-1925, 2030-2289 (struct, buffs-to-mobs, Prepare, CleanupMobSpawns, AddMob/RemoveMob, visitors, GetMobs, PruneVisitors, RoundTick, AddPlayer, Validate). Other 1,100 lines are finders, nouns, signs, exits text |
| 5 | `internal/rooms/roommanager.go` | full (989) |
| 6 | `internal/rooms/ephemeral.go` | full |
| 7 | `internal/mobs/mobs.go` | full (698) |
| 8 | `internal/mobs/mobs_path.go` | full |
| 9 | `internal/mobs/memory.go` | full |
| 10 | `internal/characters/character.go` | partly: 1-400, 460-529, 1066-1149, 1345-1619, 2015-2090 (struct, New, charm, aggro/buffs, RecalculateStats, AutoTrain, Validate, permabuffs) |
| 11 | `internal/characters/aggro.go` | full |
| 12 | `internal/characters/charminfo.go` | full |
| 13 | `internal/buffs/buffs.go` | partly, 1-80 (Buff and Buffs structs, Validate) |
| 14 | `internal/items/items.go` | partly, 37-166 (Item struct, New, Validate) |
| 15 | `internal/ids/ids.go` | full |
| 16 | `internal/actors/wire.go` | full |
| 17 | `internal/actors/registry.go` | partly, 1-62 (Spawn) |
| 18 | `internal/guards/guards_test.go` | partly, 140-199 (ID-creation guard) |
| 19 | `internal/mobcommands/suicide.go` | full (434) |
| 20 | `internal/mobcommands/despawn.go` | full |
| 21 | `internal/mobcommands/go.go` | full |
| 22 | `internal/mobcommands/wander.go` | full |
| 23 | `internal/mobcommands/pathto.go` | full |
| 24 | `internal/hooks/NewRound_MobRoundTick.go` | full |
| 25 | `internal/hooks/NewRound_IdleMobs.go` | full |
| 26 | `internal/hooks/MobIdle_HandleIdleMobs.go` | full |
| 27 | `internal/hooks/PlayerDespawn_HandleLeave.go` | full |
| 28 | `internal/hooks/NewRound_HandleRespawns.go` | full |
| 29 | `internal/hooks/NewTurn_AutoSave.go` | full |
| 30 | `internal/hooks/Buff_ApplyBuffs.go` | full |
| 31 | `internal/hooks/NewRound_DoCombat.go` | partly, 890-934 (death dispatch) |
| 32 | `internal/hooks/NewRound_UserRoundTick.go` | partly, 1-40 (RoundTick caller) |
| 33 | `internal/hooks/PlayerSpawn_HandleJoin.go` | full |
| 34 | `modules/cleanup/cleanup.go` | full |
| 35 | `modules/follow/follow.go` | partly, 70-169 (follow state) |
| 36 | `internal/usercommands/admin.teleport.go` | full |
| 37 | `internal/usercommands/admin.mob.go` | partly, 40-169 (`mob spawn`) |
| 38 | `internal/usercommands/go.go` | partly, 200-259 (charmed mobs follow) |
| 39 | `internal/accounts/users.go` | partly, 225-504 (login charm prune, link-dead, LoadUser) |
| 40 | `internal/accounts/index.go` | partly, 1-400 of 520 (format, load, create, add; the rest is the same pattern for remove/format) |
| 41 | `internal/accounts/copyover.go` | full |
| 42 | `internal/configs/config.memory.go` | full |
| 43 | `internal/configs/config.timing.go` | partly, 1-50 |
| 44 | `_datafiles/config.yaml` | partly, 88-127 (Memory block) |
| 45 | `copyover.go` (repo root) | full |
| 46 | `main.go` | partly, 60-452 and 1430-1507 (startup, shutdown, resume, loadAllDataFiles) |
| 47 | `world.go` | partly, 280-409 (enterWorld, MainWorker shutdown and maintenance) |
| 48 | `_datafiles/world/empty/rooms/startland/1.yaml` | full |
| 49 | `_datafiles/world/empty/rooms/startland/2.yaml` | full |
| 50 | `_datafiles/world/empty/mobs/startland/1-rat.yaml` and `2-guard.yaml` | full (counted as one row; both read) |
| 51 | `_datafiles/world/empty/users/1.yaml` | full |

Not read: `internal/copyover/copyover.go` (the exec/fd mechanism; only its callers matter here),
`fileloader` (whether `LoadFlatFile` calls `Validate()`: open question 10.4). Caller lists below come from
repo-wide `grep -rn --include=*.go` runs that did return hits, so each search was capable of matching.

---

## 1. Room instance persistence

**Summary:** a room instance file is a reflection diff of exported, non-`yaml:"-"`, non-`instance:"skip"`
Room fields against the template, written to `rooms.instances/<zone>/<id>.yaml` and unmarshalled over the
template on load. Spawn slots are excluded from the diff, and nothing about mobs is saved. There is no
load hook; `LoadRoom` is the single choke point.

**Write (`SaveRoomInstance`, `save_and_load.go:242-310`):**
- Refuses ephemeral rooms (244). Loads the template (248).
- Iterates `reflect.TypeOf(r)` fields (259): skips `field.PkgPath != ""` (unexported, 261), `yaml:"-"`
  (265-268), `instance:"skip"` (270), and any field `reflect.DeepEqual` to the template (277). Key is the
  yaml tag name or the Go field name (281-285).
- Path: `<DataFiles>/rooms.instances/<zone folder>/<RoomId>.yaml` (291-293; folder from `ZoneToFolder`,
  lowercased with `_`, `roommanager.go:598-612`). Empty diff deletes the file (295-298). Format: YAML
  `map[string]interface{}` via `gopkg.in/yaml.v2` (300).
- Only top-level fields are diffed; a changed field is written whole.

**Room fields and whether the instance save sees them (`rooms.go:59-96`):**

| Saved in instance diff | Not saved |
|---|---|
| `Containers`, `Exits` (locks), `Nouns`, `Items`, `Stash`, `Gold`, `SkillTraining`, `Signs`, `IdleMessages`, `LongTermDataStore`, `Mutators`, `Pvp`, `Tags`, map coords, plus title/description etc. if changed | `Biome` and `SpawnInfo` (`instance:"skip"`, 68, 77); `ExitsTemp`, `Corpses`, `LastIdleMessage` (`yaml:"-"`, 71, 75, 81); `bodies`, `userVisitors`, `mobVisitors`, `lastVisited`, `tempDataStore` (unexported, 91-95) |

`SpawnInfo.InstanceId` and `SpawnInfo.DespawnedRound` are also `yaml:"-"` (`spawninfo.go:7, 20`), so even
without `instance:"skip"` the slot state would not be written.

**When it runs (`SaveAllRooms`, `save_and_load.go:324-346`, iterates only rooms in memory, skips
ephemeral):**

| Caller | Where | Order |
|---|---|---|
| Shutdown | `world.go:358` in `MainWorker` on `shutdown` | rooms, then `accounts.SaveAllUsers()` (361) |
| Autosave | `NewTurn_AutoSave.go:53` | users (41), rooms (53), plugins (65); every `RoundsPerAutoSave` rounds, default 900 (`config.timing.go:27-28`) |
| Copyover | `copyover.go:20` (`triggerCopyover`) | rooms (20), users (21), plugins (22), then `copyover.Execute` |
| Room unload | `removeRoomFromMemory` calls `SaveRoomInstance` directly (`roommanager.go:517`), not `SaveAllRooms` | after destroying mobs (501-515) |
| Template edit | `SaveRoomTemplate` calls `SaveRoomInstance` (`save_and_load.go:236`) | after re-adding the room |

**Load (`LoadRoom`, 79-98):** memory cache (88), else `LoadRoomInstance` (92) then `addRoomToMemory` (93).
`LoadRoomInstance` (101-123) loads the template then `yaml.Unmarshal(bytes, room)` over it (119), error
ignored. `addRoomToMemory` (`roommanager.go:527-569`) sets `lastVisited = now` and caches. `LoadRoom` does
not call `Prepare`; spawning happens later in `Prepare` (callers in section 4).

**Hook points:** none exists for room load or unload. `util.Hook` instances in the repo are
`OnActivate`/`OnDeactivate` (accounts), `OnBuildPrompt`, `OnGetFormattedName`, `OnInstanceCreated`/
`OnInstanceDestroyed` (mobs), `OnRoomLook`, `OnGetHelpContents` (grep `util.Hook\[`). Natural insertion
points: between `LoadRoomInstance` and `addRoomToMemory` in `LoadRoom` (`save_and_load.go:92-93`) for
restore, and before `mobs.DestroyInstance` in `removeRoomFromMemory` (`roommanager.go:501`) plus in
`SaveAllRooms` for save. An exported Room field (say `NPCs`) would ride the existing diff for free: the
template never has one, so `DeepEqual` fails and it is written; `yaml.Unmarshal` would fill it on load.

**Unexported/`yaml:"-"` mechanism confirmed:** explicit `PkgPath` and tag checks at `save_and_load.go:261-272`.
Inside a written value, `yaml.v2` itself skips unexported fields and honours `yaml:"-"`.

## 2. What a mob instance carries

**Summary:** a mob is `mobs.Mob` (template fields plus runtime) wrapping `characters.Character`. YAML of the
whole `Mob` would already carry most of decision 13's "kept" list; the gaps are `HomeRoomId`, `IsElite`,
`Charmed`, `Aggro` (all `yaml:"-"`), the spawn-slot permabuff list (unexported), and item UUIDs.

### 2a. `mobs.Mob` (`internal/mobs/mobs.go:57-81`)

| Field | YAML today | Decision 13 class | Note |
|---|---|---|---|
| `MobId` | yes, no tag, key `mobid` | keep (template key) | |
| `Zone` | yes | template-only | template zone; current zone is `Character.Zone` |
| `ItemDropChance`, `ActivityLevel` | yes | template-only | clamped by `Mob.Validate` (554-577) |
| `HomeRoomId` | **no** (`yaml:"-"`, 62) | keep | the spawn room; `FindNative` and `pathto home` use it |
| `Hostile` | yes | keep (spawn override `ForceHostile`) | |
| `LastIdleCommand`, `BoredomCounter` | no (`yaml:"-"`) | reset | |
| `Groups`, `Hates`, `AngryCommands`, `CombatCommands` | yes | template-only | |
| `IdleCommands` | yes | keep (spawn override) | |
| `Character` | yes, no tag, key `character` | see 2b | |
| `MaxWander` | yes | keep (spawn override) | |
| `WanderCount`, `PreventIdle` | no | reset | |
| `BuffIds` | yes | template-only | |
| `EliteChance` | yes | template-only | |
| `IsElite` | **no** (`yaml:"-"`, 77) | keep | rolled at spawn, cannot recompute |
| `Path` | no (`yaml:"-"`); `PathQueue` has only unexported fields holding `PathRoom` interfaces (`mobs_path.go:9-12`) | reset | would not marshal anyway |
| `tempDataStore` | no (unexported) | reset | holds `home-impossible` (`pathto.go:18, 49, 56`) |
| `lastCommandTurn` | no (unexported) | reset | |

### 2b. `characters.Character` (`internal/characters/character.go:46-98`)

| Field | YAML today | Class | Note |
|---|---|---|---|
| `Id` | yes (`id`) | keep | `ids.ID` marshals via `MarshalText` as `0x..:0x..` (`ids.go:75-88`); seen in `users/1.yaml:8` |
| `Name`, `Description` | yes (no tag) | keep (`Name` may be a spawn override) | |
| `Adjectives` | yes | keep | includes `elite`, `charmed`, `lost` |
| `RoomId`, `Zone` | yes | keep | where it is now |
| `RoomIdOnReset` | yes | derived/unused for mobs | |
| `RaceId`, `FormRaceId` | yes | keep | |
| `Stats` | yes | keep (training) | Base/Mods/Value recomputed by `RecalculateStats` (1349-1449) |
| `Level`, `Experience`, `TrainingPoints`, `StatPoints` | yes | keep | level cannot be recomputed (random zone scale, elite roll) |
| `Health`, `Mana`, `ActionPoints`, `Alignment`, `Gold`, `Bank` | yes | keep | |
| `Shop` | yes | keep | merchant stock |
| `SpellBook`, `Skills`, `Cooldowns`, `Settings`, `KeyRing`, `KD`, `MiscData`, `ExtraLives`, `MobMastery`, `Created`, `Timers`, `ZonesVisited` | yes | keep (mostly empty for mobs) | |
| `Charmed *CharmInfo` | **no** (`yaml:"-"`, 69) | keep ("who charmed it") | `CharmInfo{By ids.ID, RoundsRemaining, ExpiredCommand}` all exported (`charminfo.go:13-17`), would marshal |
| `CharmedMobs []ids.ID` | **no** (`yaml:"-"`, 70) | keep (for players, see 7) | |
| `Items` | yes | keep | `items.Item.UUID` is `yaml:"-"` (`items.go:39`): item IDs are regenerated by `Validate` on load (155-157) |
| `Buffs` | yes | keep | `Buffs.List []*Buff`, exported fields; index maps unexported, rebuilt by `Buffs.Validate` (`buffs.go:38-80`) |
| `Equipment` | yes | keep | |
| `TNLScale`, `HealthMax`, `ManaMax`, `ActionPointsMax` | no | derived | `RecalculateStats` |
| `Aggro *Aggro` | **no** (`yaml:"-"`, 78) | keep ("aggro target") | `Aggro{Type, Target ids.ID, SpellInfo, ExitName, RoundsWaiting}`, all exported (`aggro.go:21-27`) |
| `roomHistory` | no (unexported) | reset | |
| `PlayerDamage`, `LastPlayerDamage` | no | reset (XP split on death) | |
| `KilledBy`, `KillerMobIsElite`, `KillerMobName` | no | reset (player-only) | |
| `permaBuffIds` | **no** (unexported, 96) | keep (see risk 9.3) | set by `SetPermaBuffs` (2020-2022) |
| `userId` | no (unexported) | n/a for mobs | |

Marshalling caveats: `MiscData map[string]any` and `Timers map[string]gametime.RoundTimer` round-trip
for players today. `Path`, `tempDataStore` and `Aggro`'s lack of tag are the only non-clean pieces, and all
but `Aggro` and `Charmed` are reset anyway.

### 2c. Instance vs template after `NewMobById` (`mobs.go:153-230`)

| Step | Line | Recomputable on restore? |
|---|---|---|
| Shallow copy `mob := *m` | 157 | n/a. Slices and maps (`Character.Items`, `SpellBook`, `Shop`...) share backing storage with the template (risk 9.6) |
| `HomeRoomId`, `Character.RoomId` = spawn room; `PlayerDamage` new map | 159-161 | HomeRoomId no (not saved) |
| `Level` = forceLevel | 164-166 | no (zone autoscale is random, `rooms.go:576-578`) |
| Elite roll, level bonus | 169-180 | no (random) |
| `StatPoints` from level, `Experience = XPTNL` of level-1, `AutoTrain`, full `Health`/`Mana` | 182-198 | derived from level, but saved Stats training and current HP/MP must be kept |
| `SetAdjective("elite")` | 200-202 | saved in `Adjectives` |
| `SetPermaBuffs(mob.BuffIds)`, then `Buffs = buffs.New()` | 204-206 | permabuffs reapplied by `Validate(true)` at 225 |
| Item and equipment `Validate` (UUIDs) | 208-222 | UUIDs regenerate |
| Alignment from race if zero | 212-218 | saved |
| `mob.Validate()`, `Character.Validate(true)` | 224-225 | |
| `addInstance`: **`Character.Id = ids.New(ids.NoCell)`**, store, fire `OnInstanceCreated` | 42-48 | must not be repeated on restore (guard, risk 9.2) |

### 2d. Spawn overrides in `Room.Prepare` (`rooms.go:565-628`)

| Override | Line | Saved by mob YAML? |
|---|---|---|
| `Level` / zone autoscale + `LevelMod` | 567-588 (passed to `NewMobById` at 590) | `Character.Level` yes |
| Shop `Restock()` | 593-595 | `Shop` yes |
| `SpawnInfo.BuffIds` via `SetPermaBuffs` (replaces template list) | 597-599 | **no** (`permaBuffIds` unexported) |
| `IdleCommands` | 602-604 | yes |
| `Name` | 607-609 | yes |
| `ForceHostile` | 611-613 | `Hostile` yes |
| `MaxWander` | 615-617 | yes |
| `Character.Zone = r.Zone`, `mob.Validate()` | 619-620 | yes |
| Appended straight to `r.bodies` (not `AddMob`: no RoomChange event, no visit mark) | 622 | n/a |

Must be saved because not recomputable: level, elite flag, home room, the override results (which mob YAML
already carries except the spawn permabuff list), HP/MP, inventory, gold, equipment, buffs, aggro, charm.

## 3. Every way a mob instance stops existing

**Summary:** `mobs.DestroyInstance` (`mobs.go:257-261`: delete from `mobInstances`, fire
`OnInstanceDestroyed`, which unbinds the actor, `wire.go:54-58`) is the only removal from `mobInstances`
(grep `delete(mobInstances`). It has 6 call sites: 3 in `suicide.go`, 1 in `despawn.go`, 2 in
`removeRoomFromMemory`. `modules/cleanup` does not remove mobs (only `trash`/`bury` commands, 127-204).
Nothing issues `suicide vanish` today (grep `vanish`: only `suicide.go:24, 47-48`).

| Path | Trigger | Class (decision 14) | Slot `DespawnedRound` |
|---|---|---|---|
| `Suicide` vanish branch, `suicide.go:48-62` | `suicide vanish` (no caller) | admin/cleanup despawn | `LoadRoom(HomeRoomId).CleanupMobSpawns(false)` (54-56) sets `roundNow` (`rooms.go:727`) |
| `Suicide` normal, perma-gear off, `suicide.go:324-411` | `mob.Command("suicide")` from `NewRound_MobRoundTick.go:68-70`, `NewRound_DoCombat.go:924-927`; `Input suicide` from `Buff_ApplyBuffs.go:64-69` | death | `CleanupMobSpawns(false)` (391-393), `roundNow`. Drops items/gold or corpse |
| `Suicide` normal, perma-gear on, `suicide.go:414-433` | same | death | `CleanupMobSpawns(false)` (418-420) |
| `Despawn`, `despawn.go:9-25` | boredom: `NewRound_IdleMobs.go:44-47` (`despawn depression`), only when live mob count exceeds `MobUnloadThreshold` (29-32; config 100, `config.yaml:102`) and `BoredomCounter >= MaxMobBoredom` (250); charm expiry: `CharmExpiredDespawn` command `...;despawn charmed mob expired` (`charminfo.go:7`) run by `NewRound_MobRoundTick.go:47-62` | cleanup despawn | `CleanupMobSpawns(true)` (18): **`DespawnedRound = 0`, respawn immediately** (`rooms.go:724-725`) |
| `removeRoomFromMemory`, `roommanager.go:489-520` | `RoomMaintenance` (229) and `TryEphemeralCleanup` (`ephemeral.go:254`) | unload | **none**: room struct is dropped (519), slot state lost |

`Suicide` before any branch: `revive-on-death` buff cancels the death (24-33); charm removed and the
charmer's `CharmedMobs` updated (41-45); XP, KD, alignment, tame chance per `PlayerDamage` (79-322).
`MobDeath` event (86-94) carries `ActorId`.

Fallback slot resets (instance missing when the room next looks): `Prepare` (`rooms.go:541-548`) and
`RoundTick` (2062-2067) both set `DespawnedRound = roundNow`. `CleanupMobSpawns` and these run on the home
room only, found by `mob.HomeRoomId`.

## 4. Spawn slots

**Summary:** a slot is "filled" while `InstanceId` names a live instance anywhere in the world, "despawned"
when `InstanceId` is zero and `DespawnedRound > 0`, and "ready" when zero/zero or the respawn period has
passed. Slot state lives only in memory.

`SpawnInfo` (`spawninfo.go:5-22`): `MobId`, `InstanceId ids.ID yaml:"-"`, `Container`, `ItemId`, `Gold`,
`Message`, `Name`, `ForceHostile`, `MaxWander`, `IdleCommands`, `BuffIds`, `Level`, `LevelMod`,
`DespawnedRound uint64 yaml:"-"`, `RespawnRate`. `RespawnRate` defaults to `15 real minutes` in
`Room.Validate` (`rooms.go:2261-2263`).

Every read/write of `InstanceId` / `DespawnedRound` (grep in `internal/rooms`):

| Site | Rule |
|---|---|
| `Prepare` 541-550 | `InstanceId` set and `mobs.GetInstance` nil: clear it, `DespawnedRound = roundNow`, skip. Set and live: skip (**filled, wherever the mob is**) |
| `Prepare` 553-558 | `DespawnedRound > 0` and `roundNow < GetDate(DespawnedRound).AddPeriod(RespawnRate)`: skip |
| `Prepare` 624-627 | after spawning: `InstanceId = mob.Character.Id`, `DespawnedRound = 0` |
| `Prepare` 653, 676 | item/gold slots reuse `DespawnedRound` as last-spawn round |
| `CleanupMobSpawns` 719-733 | instance gone: clear, `DespawnedRound = 0` if `noCooldown` else `roundNow` |
| `RoundTick` 2062-2067 | instance gone: clear, `roundNow` |
| `removeRoomFromMemory` 505-515 | slot instance live and `Character.RoomId == room.RoomId`: destroy (redundant with 501-503) |

`Prepare` callers: `MoveToRoom` when the room has no players (`roommanager.go:312-314`, `Prepare(true)`, which
also prepares adjacent rooms without players, `rooms.go:686-708`); `HandleRespawns` for rooms with players
(`NewRound_HandleRespawns.go:14-23`); user `look` (`usercommands/look.go:408`); mob `look`
(`mobcommands/look.go:173`); `admin.prepare.go:28`; loot goblin (`admin.mob.go:128`). (Call sites from grep.)

Wandered mob: the slot stays filled while the mob is live anywhere (541-549 only checks `GetInstance`).
`HomeRoomId` is the room whose `Prepare` spawned it (`NewMobById(..., r.RoomId, ...)`, 590 into
`mobs.go:159`); death and despawn find the slot only through it (`suicide.go:54, 391, 418`;
`despawn.go:17`). A slot does not record which mob is its own beyond `InstanceId`, and a mob does not
record which slot index it came from (a room can have several slots for the same `MobId`: room 2 has three
rat slots, `2.yaml:10-19`). `CleanupMobSpawns` handles this by sweeping every slot for missing instances.
Admin `mob spawn` creates a mob with `HomeRoomId = room` and no slot (`admin.mob.go:141-142`).

## 5. Wandering and room membership

**Summary:** mobs move only through `Room.RemoveMob` + `Room.AddMob`; `AddMob` sets `Character.RoomId` and
`Zone`. A mob can be live in memory while its home room is unloaded; a mob cannot be live in a room that is
unloaded, because unload destroys every mob listed in it.

- `AddMob` (`rooms.go:737-759`): `MarkMobVisited`, queue `RoomChange`, set `Character.RoomId` and `Zone`
  (753-754), append to `bodies`, update `roomsWithMobs`. `RemoveMob` (761-770): mark visited, drop body.
- Callers (grep): `mobcommands/go.go:42-43` (forced move to a room id with no exit) and `:110-111` (exit);
  `admin.mob.go:142`; `admin.teleport.go:135-136` (charmed mobs with party members). Death/despawn call
  `RemoveMob` only.
- `wander.go`: refuses when charmed (14-16); over `MaxWander` from home: `pathto home` (19-28); picks an exit
  in the same zone, `WanderCount++`, queues `go <exit>` (65-84).
- `pathto.go`: `mapper.GetPath` to `HomeRoomId` or waypoints (28-46), sets `Path` (68); unreachable home
  sets temp `home-impossible` and adjective `lost` (48-58), then drains 10% HP per call (17-25).
- `IdleMobs` (`NewRound_IdleMobs.go:80-123`) steps the path; `HandleIdleMobs` (`MobIdle_HandleIdleMobs.go:27-42`)
  sends displaced or over-wandered uncharmed mobs home.
- Unload: `RoomMaintenance` (`roommanager.go:160-235`) unloads rooms whose `lastVisited` is older than
  `RoomUnloadRounds` (450), only while the loaded-room count exceeds `RoomUnloadThreshold` (200)
  (`config.yaml:106, 109`; `roommanager.go:173-176, 217-222`). `removeRoomFromMemory` refuses rooms with
  players (497-499) and is the only non-ephemeral unload (`ClearRoomCache` at `admin.go:264` and via
  `addRoomToMemory(force)` drops a room without destroying mobs; not followed further).
- `lastVisited` is set by `MarkUserVisited`/`MarkMobVisited` (1107, 1117) and `addRoomToMemory` (538).
  `PruneVisitors` refreshes visitor maps but not `lastVisited` (1880-1922). A room holding resting mobs and no
  traffic therefore ages out and unloads, destroying them.
- Home room unloaded while its mob wanders: the mob survives with `HomeRoomId` pointing at an unloaded room.
  When the room reloads, `SpawnInfo` comes from template YAML with zero `InstanceId`, so `Prepare` spawns a
  second mob for that slot. This duplicate exists today.
- In the `empty` world there are 4 rooms and 4 mobs, below both thresholds: neither room unload nor boredom
  despawn can happen in smoke without changing config.

## 6. Startup and copyover order

**Summary:** rooms are cached as summaries at boot and loaded lazily; no mob exists until a room is
`Prepare`d by a player entering. Copyover saves rooms first, but saved instance files carry no mobs, so every
NPC after copyover or restart is a fresh instantiation with a new ID.

`main.go` order:

| Step | Line |
|---|---|
| `actors.Init()` (hooks into mobs and accounts) | 105 |
| copyover contributors registered: connections, accounts, util, gametime, token | 108-112 |
| `configs.ReloadConfig()` | 124 |
| `os.Mkdir rooms.instances` | 171 |
| `hooks.RegisterListeners()` | 203 |
| `loadAllDataFiles(false)`: `rooms.LoadDataFiles` (zones and room summaries only, `roommanager.go:978-989`, `save_and_load.go:362-454`), then buffs, items, races, skills, `mobs.LoadDataFiles` (templates) ... | 214; 1494-1501 |
| `mapper.PreCacheMaps()` | 220 |
| not copyover: `InitUserIndex`, rebuild if stale, `DoBodyIdMigrationV1` | 226-245 |
| `GetCharacterIndex().Rebuild()` | 247 |
| not copyover: `LoadRoundCount` | 251-255 |
| copyover: `copyover.Restore(fd)` (accounts contributor reloads each user from disk, `accounts/copyover.go:56-88`) | 257-263 |
| `plugins.Load` | 274 |
| listeners, `InputWorker`, `MainWorker` (world start) | 297-344 |
| copyover: `resumeRestoredConnection` per user, which `SendEnterWorld` (445) to `enterWorld` then `MoveToRoom` (`world.go:291-307`), so `Prepare(true)` spawns fresh mobs | 349-364 |
| shutdown: `MainWorker` saves rooms then users (`world.go:358-361`); `main` saves round count (382) and plugins (403) | |

`triggerCopyover` (`copyover.go:12-25`): `serverAlive=false`, `SaveAllRooms`, `SaveAllUsers`, `plugins.Save`,
`copyover.Execute`. No room is in memory in the new process until a player arrives.

NPC index load/save candidates: load after `loadAllDataFiles` (mob templates and room summaries exist) and
before `copyover.Restore`/listeners, beside `InitUserIndex` but on both the copyover and cold paths (the
user index is skipped on copyover, 228). Save alongside `SaveAllRooms` in all three places it is called
(`world.go:358`, `NewTurn_AutoSave.go:53`, `copyover.go:20`) and on each `removeRoomFromMemory`.

## 7. Companions

**Summary:** no code path charms a mob today, so companions only exist in principle. Charm state is not
saved on either side, and login prunes a player's list against live instances.

- `Character.Charmed` and `CharmedMobs` are `yaml:"-"` (`character.go:69-70`). API: `TrackCharmed`,
  `GetCharmIds`, `Charm`, `IsCharmed`, `RemoveCharm` (464-525). `Charm` has no caller (grep `\.Charm(`: none
  outside `character.go`).
- Login prune: three sites drop IDs where `!mobs.MobInstanceExists` (`users.go:246-250` reconnect,
  `290-294` link-dead reconnect, `325-329` fresh login). Because `CharmedMobs` is not saved, a fresh login
  starts empty anyway; the prune matters only for in-memory accounts.
- Link-dead: newbie-guide mobs get `RoundsRemaining = 0` (`users.go:348-362`).
- Logout: `HandleLeave` expires the charm of charmed mobs in the player's current room only
  (`PlayerDespawn_HandleLeave.go:49-55`); `MobRoundTick` then runs `ExpiredCommand` (`NewRound_MobRoundTick.go:47-62`),
  which for `CharmExpiredDespawn` ends in `despawn` (slot cooldown zero).
- Charm countdown: `MobRoundTick` 32-34; players' own `Charmed` 103-104 of `NewRound_UserRoundTick.go` (grep).
- Follow: user `go` makes charmed mobs in the origin room repeat the move (`usercommands/go.go:227-239`).
  `modules/follow` keeps `followed`/`followers`/`followLimits` maps in module memory only (`follow.go:83-90`),
  not saved: reset under decision 13.
- Teleport: only when the teleported user is a party leader, each party member in the room is moved and
  that member's charmed mobs go with them (`admin.teleport.go:108-142`). The leader's own charmed mobs are not
  moved. `MoveToRoom` itself moves no mobs (`roommanager.go:264-392`).
- Death of a charmed mob updates the charmer (`suicide.go:41-45`).

## 8. Existing world data

**Summary:** no room instance files exist; the two zone folders are empty and gitignored. The user index is
the only binary index and is a reasonable model.

- `_datafiles/world/empty/rooms.instances/` holds empty `shadow_realm/` and `startland/` (made by
  `loadAllRoomZones`, `save_and_load.go:380-383`). Saves would land at `rooms.instances/startland/1.yaml`,
  `.../2.yaml`, `rooms.instances/shadow_realm/75.yaml` (and `-1.yaml`). `.gitignore:11` ignores
  `**/rooms.instances`; `.gitignore:4` ignores `_datafiles/**/users/*` (yet `users/1.yaml` is tracked).
- Spawn data: room 1 one guard slot (`1.yaml:12-22`, idle overrides, `levelmod: 10`, `5 real minutes`);
  room 2 three rat slots (`2.yaml:10-19`).
- `users.idx` (`internal/accounts/index.go`): fixed-width binary file at `<DataFiles>/users/users.idx` (66);
  100-byte text header `VERSION=..,RECORDCOUNT=..,RECORDSIZE=..,CHECKSUM=..` (31, 361); 89-byte records
  (80-byte name, 8-byte LE id, newline, 30, 391-394); loaded fully into maps (95-136); append-only `AddUser`
  (370-400); FNV-64 checksum of the directory to detect staleness and `Rebuild` from source files
  (176-263). It is a derived cache rebuilt from the YAML it indexes. An NPC index keyed by `ids.ID` with a
  room id per entry would be small; the rebuild-from-source and checksum pattern is the part worth copying,
  the fixed-width name layout is not.

## 9. Risks and surprises

1. **Spawn slots are not saved at all.** `SpawnInfo` is `instance:"skip"` and its two state fields are
   `yaml:"-"`. Decision 14's "slots save which NPC they own and its death time" needs a new saved shape
   (slot state keyed by slot index, since the template `spawninfo` list is the identity) and a merge onto
   template `SpawnInfo` on load. Slot identity by list index breaks if a builder reorders `spawninfo`.
2. **Restore cannot go through `NewMobById`/`addInstance`.** `addInstance` always makes a new ID
   (`mobs.go:43`), and `TestIdsAreMadeOnlyAtCreation` (`guards_test.go:162-187`) allows `ids.New` in
   `mobs.go` only as "a new NPC instantiation". A restore entry must insert into `mobInstances` with the
   saved ID and still fire `OnInstanceCreated` so `actors.Spawn` binds it (`wire.go:46-52`); `Spawn` refuses
   a duplicate ID for a different body (`registry.go:44-52`), which protects against double restore.
3. **Spawn permabuffs are invisible to YAML.** `permaBuffIds` is unexported. On restore, any
   `Validate(true)` (`character.go:1607-1609`, called from `CancelBuffsWithFlag` 1117, form change 1649/1656,
   `NewMobById` 225) runs `reapplyPermabuffs`, which removes `PermaBuff` buffs not backed by
   `permaBuffIds`, race or worn items (2042-2048, 2068-2071). Restore must re-derive the list (template
   `BuffIds`, or the slot's `BuffIds` when the slot set one, since `SetPermaBuffs` replaces rather than
   merges, `rooms.go:597-599`). Separately, `Prepare` sets the slot's list after `NewMobById` and then calls
   `mob.Validate()` without `true` (620), so slot permabuffs appear not to be applied at spawn today until
   some later `Validate(true)`; worth a test before relying on it.
4. **`HomeRoomId`, `IsElite`, `Charmed`, `Aggro` are `yaml:"-"`.** Saving a `Mob` by plain YAML loses four
   "kept" items. Changing the tags also changes `Mob.Save()` (`mobs.go:593-610`, template save), which would
   then write runtime fields into template files; a separate save struct or a save-time copy avoids that.
5. **Unload already duplicates wandering mobs** (section 5). Decision 12 fixes it only if the wanderer,
   not the room, carries persistence: an NPC is "saved with the room it is in", but its slot lives in
   another room that may unload first. Unloading the home room must keep the slot pointing at a live NPC
   elsewhere, and unloading the NPC's current room must save it there while the slot (maybe still loaded)
   still says filled. Hence the index in decision 14; every slot check (`Prepare`, `RoundTick`,
   `CleanupMobSpawns`) uses `mobs.GetInstance` today and must also consult the index.
6. **Template sharing through the shallow copy.** `mob := *m` (`mobs.go:157`) shares `Items`, maps and
   `Shop` with the template; `Items[idx].Validate()` (208-210) writes a UUID into the template's array.
   A restored mob decoded from YAML gets fresh storage, so behaviour differs from a fresh spawn; the
   existing sharing is a latent bug either way.
7. **Item IDs are not persistent.** `Item.UUID` is `yaml:"-"` (`items.go:39`); decision 11 lists items as
   keeping their ID for good. Out of scope for 2b, but a restored NPC's inventory will have new UUIDs.
8. **Boredom despawn is a cleanup despawn with zero cooldown** (`despawn.go:18`). Under decision 14 it
   counts as death, so it respawns at once; with persistence it will keep culling and replacing NPCs in busy
   worlds. It is gated by `MobUnloadThreshold` (100). Calabe may want boredom to become "unload" instead.
9. **Elapsed time and rounds.** `DespawnedRound` is a round count; the round counter persists across restart
   (`util.LoadRoundCount`, `main.go:251-255`) and copyover (util contributor), so a saved death round stays
   meaningful. Charm `RoundsRemaining` and buff counters also restore as saved, matching decision 13's
   "whatever the elapsed time".
10. **Ephemeral rooms** are copies of templates with `SpawnInfo` (`ephemeral.go:148-161`) and spawn mobs;
    `SaveRoomInstance` refuses them. Their NPCs should stay transient, or they need their own rule.
11. **Template edits drop slot state.** `SaveRoomTemplate` replaces the room with the new template and copies
    items, gold, signs and bodies but not `SpawnInfo` state (`save_and_load.go:186-236`), so editing a room
    respawns its slots today. A saved slot state would fix this only if the merge runs there too.
12. **Smoke cannot reach unload.** The `empty` world is below `RoomUnloadThreshold` and `MobUnloadThreshold`;
    tests for unload persistence need config overrides or unit tests on `removeRoomFromMemory`.

## 10. Open questions

1. Whether `yaml.v2` round-trips `ids.ID` as a map key and inside `Aggro`/`CharmInfo` pointers: `MarshalText`
   is implemented and the player save shows the scalar case, but the key and pointer cases were not run.
2. Whether boredom despawn and charm-expiry despawn should count as death (decision 14 says "despawned by
   cleanup" is death) or as unload. A ruling, not a reading.
3. What a restored NPC's `HomeRoomId` should mean if its template room no longer has a matching slot (slot
   removed by a builder): log and keep, or treat as admin-spawned (no slot).
4. Whether `fileloader.LoadFlatFile` calls `Room.Validate()` (it would set the `RespawnRate` default);
   `fileloader` was not read.
5. Whether `ClearRoomCache` (via `admin.go:264` and forced `addRoomToMemory`) can orphan mobs: it drops a room
   without `removeRoomFromMemory`, so listed mobs stay in `mobInstances` with no room. Not traced.
6. Whether a restored NPC with `Aggro` against a player who is offline should keep it: `IdleMobs` already
   clears aggro on a lost target (`NewRound_IdleMobs.go:65-75`), so keeping it is cheap, but it is a choice.
