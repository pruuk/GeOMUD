# Phase 4b: NPCs are players (design)

Date: 2026-10-27. Parent: `gomud_strip_plan.md` Phase 4b ("One actor model (decision 4): an NPC is a
player-shaped actor with no connection, running the same commands, combat path, round tick and ID
space as players, with NPC behaviour layered on top. `mobcommands` and the player/mob combat variants
go. Done when: one command set and one combat path serve players and NPCs; smoke passes, including a
fight against an NPC"). Branch `strip/phase-4b-npcs-are-players`.

## Facts verified against source (2026-10-27, master `c4314b7f`)

Counts are non-test Go files under `internal`, `modules` and the repo root unless a row says otherwise.

| Fact | Value | Where / how |
|---|---|---|
| Shared character core | Players and mobs both hold `characters.Character` | `UserRecord.Character` (`internal/users/userrecord.go:40`); `Mob.Character` (`internal/mobs/mobs.go:40-66`) |
| Player wrapper | `users.UserRecord`: account fields plus connection state (`connectionId`, `unsentText`, `activePrompt`, `isLinkDead`, `inputBlocked`, ...) | `internal/users/userrecord.go:31` |
| Mob wrapper | `mobs.Mob`: YAML spawn data (`MobId`, `Zone`, `ActivityLevel`, `Hostile`, `Groups`, `Hates`, `IdleCommands`, `AngryCommands`, `CombatCommands`, `MaxWander`, `BuffIds`, `EliteChance`) plus runtime state (`InstanceId`, `HomeRoomId`, `LastIdleCommand`, `BoredomCounter`, `WanderCount`, `PreventIdle`, `IsElite`, `Path`, `lastCommandTurn`, `playersAttacked`) | `internal/mobs/mobs.go:40` |
| Mob IDs | Own counter `instanceCounter`, map `mobInstances` | `internal/mobs/mobs.go:22-26, 142-148` |
| Package sizes | `users` 15 files / 2,931 lines; `mobs` 6 / 1,031; `characters` 11 / 3,085; `usercommands` 68 / 11,617; `mobcommands` 20 / 2,160; `combat` 7 / 1,913 | `wc -l` |
| Command sets | `usercommands` registers 65 commands, `UserCommand func(rest, *users.UserRecord, *rooms.Room, events.EventFlag)`; `mobcommands` registers 20 (broadcast, despawn, drop, emote, equip, get, give, go, look, noop, pathto, put, remove, replyto, say, sayto, saytoonly, shout, suicide, wander), `MobCommand func(rest, *mobs.Mob, *rooms.Room)` | `internal/usercommands/usercommands.go:30`; `internal/mobcommands/mobcommands.go:16-40` |
| Dead mob command | `mobcommands/unlock.go` defines `Unlock` but it is not registered (the user `unlock` is) | `grep` |
| Direct and wide comms | No `tell`; `whisper` is the direct command. `broadcast` exists in both sets | `usercommands.go` registry |
| Input | `events.Input{UserId, MobInstanceId, ...}`; world dispatches to `processInput` or `processMobInput` | `internal/events/eventtypes.go:66`; `world.go:89-113, 487, 598` |
| Combat variants | `AttackPlayerVsMob`, `AttackPlayerVsPlayer`, `AttackMobVsPlayer`, `AttackMobVsMob`, each called once from `NewRound_DoCombat.go` (lines 348, 517, 808, 938) | `internal/combat/combat.go:28-79` |
| Combat round | `NewRound_DoCombat.go`, 1,073 lines: `handlePlayerCombat` (39), `handleMobCombat` (597), `handleAffected` (1025) | `internal/hooks/` |
| Ticks | `UserRoundTick` (132 lines), `MobRoundTick` (76), `IdleMobs` (132), `HandleIdleMobs` on `MobIdle` (52) | `internal/hooks/hooks.go:20-32` |
| Room lists | `players []int`, `mobs []int` | `internal/rooms/rooms.go:93-94` |
| Aggro | `Aggro{Type, MobInstanceId, UserId, ...}` | `internal/characters/aggro.go:21` |
| Saved IDs | Only template and account keys are saved: `KD.Kills` by `MobId`, `PlayerKills`/`PlayerDeaths` by `userid:username`. Actor-to-actor references are transient: `Charmed`, `CharmedMobs` (`yaml:"-"`), `KillerMobInstanceId` (`yaml:"-"`), aggro, room lists | `internal/characters/character.go:67-91`; `kdstats.go` |
| Alt characters | Module stripped; `internal/users/charindex.go` still indexes names | `ls modules` |
| ID references | `MobInstanceId`: 148 uses, 37 files. `users.GetByUserId(`: 114. `mobs.GetInstance(`: 61 | `grep` |
| Output sites | `SendText(` (users and rooms): 961 in about 95 files, of which `room.SendText(` 145 in 43; `connections.SendTo(` 86 in 16; `events.Broadcast{` 14 in 8; `events.Message{` 6 in 4; `SendTextCommunication(` 9 in 4; `SendTextToExits(` 5 in 5 | `grep` |
| Combat messages | `AttackResult` has `MessagesToSource`, `MessagesToTarget`, `MessagesToSourceRoom`, `MessagesToTargetRoom`, `MessagesToRoomOld` | `internal/combat/attackresult.go:3` |
| Senses today | `events.Message.IsQuiet` ("only heard by superior hearing"); darkness affects vision only: `Room.GetVisibility` (`rooms.go:155-192`), `look` ("too dark", `look.go:228`). Messages ignore darkness | source |
| Test world | `empty` mobs: rat (`1-rat.yaml`: idle emote and `wander`, `maxwander: 8`, `activitylevel: 10`, group `rats`) and guard (`2-guard.yaml`: `maxwander: 20`, `activitylevel: 20`, `hates: rats`). Biomes: 3 `city`, 1 `default`, none dark | `_datafiles/world/empty/` |
| Mob-started fights | None since Phase 2; mobs fight back when attacked | `PATCH_NOTES.md` 2026-10-26 |
| Admin mob spawn | `mob spawn`, permission `mob.spawn` | `internal/usercommands/admin.mob.go:50` |
| Upstream prior art | `upstream/morq-combat-refactor` makes combat pluggable but keeps all four pairings (`internal/combat/delegator.go`); not used | `git show` |

## Decisions (Calabe, 2026-10-27)

1. **The controller is a swappable slot.** Nothing outside the controller assumes "player" or "NPC".
   Possession (`switch`) and link-dead stand-ins are later work that this design must allow without
   rework.
2. **Behaviour parity.** Today's mob behaviour (idle commands, wander, activity level) moves onto the
   new plumbing unchanged, as one behaviour per NPC. Mob YAML is unchanged. Named, composable
   behaviours and behaviour trees come later (DOGMud's behaviour-tree concepts are a source then).
3. **Account, character and actor are separate.** Every live actor has a runtime `ActorId`.
4. **Approach A:** a new `actors` package, converted in stages, each stage building and passing tests
   and smoke.
5. **One messaging package for all output, done fully in this phase.** It understands roles
   (actor, target, observer), senses, and whether the receiver takes text or a structured event.
6. **Ranges are room, nearby (adjacent rooms), zone and world.**
7. **Neighbouring packages are separate and out of scope:** presentation (wrapping, formatting, menus,
   forms, splash screens) downstream of messaging; GMCP fed directly by messaging; comms channels
   feeding into and fed by messaging.
8. **Senses stay minimal in 4b.** Hearing, touch and fuller light and darkness rules come later
   (DOGMud's light and darkness concepts are a source then).

## 1. Actor model

```
accounts.Account        was users.UserRecord, slimmed to the account:
                        UserId, Username, Password, Role, Permissions, Macros, Aliases,
                        ConfigOptions, ScreenReader, IsAI, EmailAddress, TipsComplete, EventLog
                        Saved as today in users/<UserId>.yaml, character included.

actors.Actor            one per live body
  Id         ActorId                 runtime only, one counter for every actor
  Character  *characters.Character
  Controller Controller              the slot

actors.Controller       interface
  Perceive(Perception)               a world event reached this actor (section 2)
  Account() *accounts.Account        nil when no account drives the actor
  Tick(actor, round)                 per-round hook

  ConnectionController  connectionId, unsent text, active prompt, link-dead and input-blocked
                        state, last input round, and the Account. Renders perceptions as text.
  BehaviourController   the mob runtime state listed in the facts table and its spawn data.
                        Receives perceptions as structured events. Tick runs idle and wander.

mobs.Template           the YAML half of today's mobs.Mob. Spawning a template creates an Actor
                        with a BehaviourController.
```

Rules:

- **Ask the actor, never its type.** No type switch or type assertion on `Controller` outside
  `actors`. Code that needs an account calls `Account()` and handles nil.
- **Connection state lives on the controller.** Moving a `ConnectionController` to another actor is
  the whole of possession.
- **`ActorId` is never saved.** `UserId` stays the account and save key; `MobId` stays the template
  key. A player's `ActorId` differs between logins. Admin commands that take an actor ID refer to live
  actors only. A lasting NPC identity for geo deltas is a separate key added by the geo work.
- **Account-only commands** (role, permissions, aliases, macros, settings, prompts; about 18 of the 68
  `usercommands` files by a name grep, to be recounted in the plan) call `Account()` and refuse when
  it is nil, as admin commands refuse non-admins today.
- **Package rename:** `users` becomes `accounts`.

## 2. Messaging

`internal/messaging` is the only producer of output.

```
Message
  Kind      Perception | Interface
  Actor     ActorId           none for world events (weather, a door creaking)
  Target    ActorId           optional
  Senses    Sight | Sound | Touch | Smell ...     which senses can carry it
  Faint     bool              needs keen senses (today's IsQuiet)
  Range     Room | Nearby | Zone | World
  Audience  InRange | ActorOnly | Named(ids)
  Text      ActorText, TargetText, ObserverText, optional FarText for Nearby and wider
            Tokens resolve per receiver: "{actor} attacks {target}!" reads
            "Fred attacks Steve!" or "Someone attacks Steve!"; {direction} in FarText
            is the exit the message arrived from

Perception  what a controller receives: Role (Actor, Target, Observer), the Message, SenseUsed
```

- **Delivery.** `messaging` resolves receivers from `Audience` and `Range`, asks
  `CanPerceive(receiver, message)` for each, picks the role, and calls `Controller.Perceive`. A
  `ConnectionController` renders the role's text with tokens resolved for that receiver. A
  `BehaviourController` receives the structured `Perception`.
- **`Room` range with a target elsewhere** (an attack through an exit) covers the rooms of both the
  actor and the target. This replaces `MessagesToSourceRoom` and `MessagesToTargetRoom`.
- **`CanPerceive` is a hook.** In 4b it applies today's room visibility (`Room.GetVisibility`) to
  Sight and passes every other sense. Phase 5's ruleset supplies the real check (blindness, keen
  hearing, light).
- **Interface text** (help, `score`, prompts, errors) is `Kind: Interface`, `Audience: ActorOnly`, no
  sense check. A `BehaviourController` drops it.
- **Before an actor exists** (login screens), interface text goes to the connection through a
  messaging function for that case, so the socket write still lives in `messaging`.
- **Formatting stays where it is.** Wrapping, `ansi` tags and templates are untouched in 4b; they move
  to the presentation package later. Only the sending moves.
- **Communication** in 4b: `say` and `emote` are Sound or Sight perceptions in `Room` range; `shout`
  uses `Nearby`, with `FarText` "Someone shouts from the {direction} direction, ..." as today
  (`usercommands/shout.go:30-59`); `whisper` uses `Named`; `broadcast` uses `World`. The
  comms-channel package takes these over later.
- **GMCP** keeps its own path in 4b and is allowed by the guard test as a named exception.
- **Guard test:** no package outside `messaging` calls `connections.SendTo` or writes text to a
  connection, apart from the named GMCP exception.

## 3. Commands, combat, ticks, rooms, death

- **One command set.** `Command func(rest string, actor *actors.Actor, room *rooms.Room, flags
  events.EventFlag) (bool, error)`. The 20 mob commands fold into the user set where a user command
  exists. `wander`, `pathto`, `despawn` and `noop` become behaviour actions: `wander` and `pathto`
  pick an exit and issue `go <exit>`; `despawn` is an engine call; `noop` is the behaviour doing
  nothing. Mob variants that differ from the user command (for example `go`, `say`, `sayto`,
  `replyto`, `saytoonly`) are diffed in the plan and each difference kept or justified. `unlock.go`
  (unregistered) is deleted.
- **One input path.** `events.Input` carries an `ActorId`; `processInput` and `processMobInput`
  merge.
- **One attack path.** `combat.Attack(attacker, defender *actors.Actor) AttackResult` replaces the
  four variants; its messages are one `messaging.Message`.
- **One combat round.** `handlePlayerCombat` and `handleMobCombat` merge into one loop over actors
  with aggro.
- **Outcomes stay identical.** The plan diffs the four handlers and lists every difference (kill
  credit to `KD.Kills` or `PlayerKills`, PvP checks, alignment change, mob mastery, loot,
  `playersAttacked`, charm). The single path keeps each by asking the actor for the data the rule
  needs (its template `MobId`, its `Account()`).
- **One tick.** `UserRoundTick`, `MobRoundTick`, `IdleMobs` and the `MobIdle` handler become one
  actor round tick that calls `Controller.Tick`. The `BehaviourController` runs idle and wander at
  today's `ActivityLevel` odds and `MaxWander` limit.
- **Rooms** hold `actors []ActorId`. Room output reads as today.
- **Aggro** holds one `ActorId` target.
- **Death.** Players keep the Phase 4 stand-in (revive at 1 health, portal out of room 75). NPCs
  keep dying, dropping loot and respawning from their template. The branch is the question "does
  this actor return from a template", which carries over to Phase 5's death rule.

## 4. Stages and testing

| Stage | Work | Checked by |
|---|---|---|
| 1 | `actors`: `Actor`, `ActorId`, `Controller`, registry. `users` renamed `accounts`; account split from connection state. Logged-in players and spawned mobs become actors | Unit tests: registry; swapping a controller keeps the `ActorId` |
| 2 | `messaging` core: kinds, roles, senses, ranges, audiences, tokens, `CanPerceive`, delivery to both controllers | Unit tests for each role, range and audience; dark room blocks Sight; "Someone" when the actor is unseen; behaviour receives the structured event |
| 3 | One input path and one command set; mob commands folded or turned into behaviour actions | Build, tests, smoke |
| 4 | Convert all output sites, split by command group across subagents | Guard test; smoke transcript unchanged |
| 5 | One attack path and combat round with every listed outcome kept | Tests, smoke fight |
| 6 | One tick; `BehaviourController` runs idle and wander | Seeded test: rat idles and wanders within `maxwander` at its `activitylevel` odds |
| 7 | Delete `mobcommands`, the four `Attack*` variants and `MobInstanceId` | Guard tests: no `Controller` type switch outside `actors`; old names absent |

Every stage ends with `go build ./...`, `go test ./... -count=1` at the Windows baseline
(`geomud-verifying`), and `go run ./cmd/smoke`.

**Smoke additions (stage 5):** the admin spawns a rat in the start room; a bot attacks it and the
fight runs to an end (the rat dies, or the bot dies and the death stand-in runs). With `-bots 2`, the
second bot must receive the observer text and not the attacker's.

**Text parity (stage 4):** record a smoke transcript before the conversion and compare after, with
random rolls masked. Prove the comparison catches a difference with a deliberate change first.
Intended changes: sight-only messages stop reaching receivers in a dark room (none exist in
`empty`), and messages from an actor the receiver cannot see say "Someone".

## Out of scope

Possession (`switch`), link-dead stand-in behaviours, named or composable behaviours, behaviour
trees, the presentation package, the GMCP package rework, the comms-channel package, hearing and
touch rules beyond `Faint`, light and darkness beyond today's visibility, the ruleset (Phase 5), and
lasting NPC identity (geo work).
