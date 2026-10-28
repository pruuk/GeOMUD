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
9. **One spec, several slices.** The stages become explicit slices (section 4), each with its own
   plan, branch and PR.

## Decisions (Calabe, 2026-10-27, slice 2 planning)

10. **IDs are `<cell>:<random>` in hex, Google-feature-ID style**: the S2 cell (the grid's cell
    key) of the room or region where the thing was created, then a random part. Format
    `0x<16 hex>:0x<16 hex>`; things created in a room with no geography use cell `0x0`. People
    never type IDs: they name things by name, and GMCP carries IDs as opaque tokens. `#N` and
    `@N` go.
11. **An ID is assigned once, at creation or instantiation, and kept for good.** It survives
    copyover, restart and anything else. Today's instances: a player's body (assigned at character
    creation, saved with the character), a player's companions, an existing NPC, an item, a room.
    Only creating or instantiating something new makes a new ID; a new instantiation of an NPC
    template is a new thing with a new ID. This replaces decision 3's "every live actor has a
    runtime `ActorId`" and section 1's "`ActorId` is never saved".

## Decisions (Calabe, 2026-10-27, slice 2b design)

12. **An existing NPC survives anything that is not its death**: copyover, restart and room unload.
    It is saved with the room it is in when that room unloads or the server stops, and comes back
    when the room loads. Live NPC instances are saved world state.
13. **A restored NPC keeps what it is, what it has and who it is tangled with; what it was doing
    resets.** Kept: template, level, elite flag, home room, applied spawn overrides, health and
    mana, inventory, gold, equipment, buffs, adjectives, aggro target, who charmed it. Reset: path,
    wander count, boredom, last idle command, temporary data. It comes back as saved, whatever the
    elapsed time; change over elapsed time belongs to the geo time-and-deltas work.
14. **A spawn slot keeps its NPC until that NPC dies.** Slots save which NPC they own and its death
    time. Unloading is not death; being killed, or despawned by an admin or cleanup, is. A slot
    respawns only after a recorded death and its respawn time. An index records where each saved
    NPC is; a slot whose NPC is neither live nor in the index logs an error and respawns.
15. **Leaving memory is never an ending.** Boredom (the memory valve) and charm expiry unload an NPC:
    it is saved with the room it is in and comes back later, the same as an NPC in a room that
    unloads. Only death, and despawn by an admin, end an NPC. (Replaces decision 14's "despawned by
    cleanup is death" for boredom and charm expiry.)
16. **A resting NPC comes back only when its room is next loaded from disk**, whatever rested it
    (room unload, boredom, charm expiry, or a later cause). It never returns into a room that stayed
    in memory, so a farewell is not undone a round later. (Calabe, 2026-10-27. Nothing charms an NPC
    today and boredom needs more than 100 live NPCs; the rule is set now so later paths share it.)
17. **A spawn slot does not add permanent buffs to the NPC it spawns; a variant is its own
    template.** `SpawnInfo.BuffIds` goes (it never took effect on a fresh spawn and no content
    uses it). Permanent buffs come from the mob template (and race and worn items), so a restored
    NPC carries exactly what a fresh one does. Room-side effects applied to whoever is in a room
    (for example regeneration) are a separate mechanism and stay. (Calabe, 2026-10-27.)

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
                        (As built in slice 1: that state lives in accounts.Session, which the
                        Account points to and the controller reaches through it, because
                        accounts cannot import actors. Moving the controller still moves it.)
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

## 4. Slices

This spec is built in seven slices. **Each slice gets its own implementation plan, branch and PR.**
A slice's plan is written when that slice starts, with its own facts table read from the code as the
earlier slices left it, and goes to Calabe for review before work begins. Each slice deletes the old
code it replaces, so no slice leaves two live paths behind.

| # | Slice | Needs | Work | Deletes | Done when |
|---|---|---|---|---|---|
| 1 | Actors | none | `actors`: `Actor`, `ActorId`, `Controller` (both implementations), registry. `users` renamed `accounts`; connection state moves from the account to `ConnectionController`. Logged-in players and spawned mobs become actors. Old IDs still route, through registry lookups. Smoke gains a transcript recording, and the pre-4b transcript is recorded from master `c4314b7f` | connection fields on the account | Unit tests: registry; swapping a controller keeps the `ActorId`. Guard test: no `Controller` type switch outside `actors`. Smoke passes |
| 2 | One ID space | 1 | The `cell:random` ID type (decision 10) in a leaf package; a body's ID lives on its `Character`; a player's is assigned at character creation and saved, an NPC's at instantiation; every reference to a live body (rooms, aggro, events, charm, killer, GMCP, the room window) holds it. Names replace `#N`/`@N` for people | `players`/`mobs` room lists, `Aggro.MobInstanceId`/`UserId`, `Input.MobInstanceId`, `instanceCounter`, `#N`/`@N` | `MobInstanceId` absent from non-test code; a guard allows new IDs only at creation sites; a player keeps their ID across logout, restart and copyover. Smoke passes |
| 2b | NPC and companion persistence | 2 | Live NPC instances and players' companions are saved and restored, with their IDs, on copyover and restart (decision 11). Its plan settles what state survives and how spawn tables count a restored NPC | respawn-on-copyover for existing NPCs; companion pruning at login | An NPC and a companion keep their IDs and state across copyover and restart |
| 2c | Item IDs | 2 | Items move from upstream's UUID to the `cell:random` ID, keeping each saved item's identity | `internal/uuid` if nothing else uses it | Every item ID is `cell:random`; saved items load with their identity kept |
| 3 | Messaging core | 1 | `messaging`: kinds, roles, senses, ranges, audiences, tokens, `CanPerceive`, delivery to both controllers. Guard test runs against an allowlist of today's offending files that may only shrink | none | Unit tests for each role, range and audience; a dark room blocks Sight; "Someone" when the actor is unseen; a behaviour receives the structured event |
| 4 | One command set and input path | 2 | `Command(rest, actor, room, flags)`; mob commands folded in or turned into behaviour actions; `processInput` and `processMobInput` merged | `mobcommands` package, `processMobInput` | One registry. Smoke passes |
| 5 | One combat path | 2, 3 | `combat.Attack` and one combat round, with every outcome difference from the handler diff kept; combat output goes through `messaging` | the four `Attack*` variants, `handlePlayerCombat`/`handleMobCombat` split, `AttackResult` room-message fields | Smoke fight: admin spawns a rat, a bot fights it to an end; with `-bots 2` the second bot receives the observer text and not the attacker's |
| 6 | One tick and behaviour | 4 | One actor round tick calling `Controller.Tick`; `BehaviourController` runs idle and wander | `UserRoundTick`/`MobRoundTick` split, `IdleMobs`, `HandleIdleMobs`, `MobIdle` event | Seeded test: the rat idles and wanders within `maxwander` at its `activitylevel` odds. Smoke passes |
| 7 | Output conversion | 3, 4 | Every remaining output site moves to `messaging`, split by command group across subagents | direct `SendText` and `connections.SendTo` outside `messaging` | Guard allowlist holds only the GMCP exception. Text parity holds. The Phase 4b Done-when line in `gomud_strip_plan.md` holds |

Order: 1, 2, 2b, 3, 4, 5, 6, 7 (2c fits anywhere after 2; rooms take the ID format with the geo
room key). Built so far: 1, 2, 2b, 2c, 3, 4, 5. Slice 3 needs only slice 1, so its plan can be written while slice 2 is
in review.

Plan files sit at the repo root as `<date>-phase-4b-<n>-<slug>-plan.md` (for example
`2026-10-27-phase-4b-1-actors-plan.md`) and move to `docs/completed/` when their slice merges.

**Every slice** ends with `go build ./...`, `go test ./... -count=1` at the Windows baseline
(`geomud-verifying`), and `go run ./cmd/smoke`.

**Text parity.** Slices 5 and 7 change how output is produced, and each compares a smoke transcript
against the pre-4b recording from slice 1, with random rolls masked. Before relying on it, prove the
comparison catches a difference with a deliberate change. Intended differences: sight-only messages
stop reaching receivers in a dark room (none exist in `empty`), and messages from an actor the
receiver cannot see say "Someone".

## Out of scope

Possession (`switch`), link-dead stand-in behaviours, named or composable behaviours, behaviour
trees, the presentation package, the GMCP package rework, the comms-channel package, hearing and
touch rules beyond `Faint`, light and darkness beyond today's visibility, the ruleset (Phase 5), and
lasting NPC identity (geo work).

## Notes for later slices (from reviews, 2026-10-27)

Each slice plan reads these first; a note leaves once a slice settles it.

- **Slices 4-5, reaching the mob behind an NPC actor.** Nothing outside `actors` can, without a
  type assertion (which `internal/guards` forbids). Put the body on the `Actor` (for example
  `Actor.Mob`, nil for players) rather than asking the controller.
- **Possession (later).** Controller-keyed indexes break under possession: an account must remember
  its own body's actor, and logout returns control to it instead of removing it. One
  `MoveController(from, to)` under the registry lock is cleaner than two `SetController` calls.
- **Actor fields and locks.** `Actor.Character` is read without the registry lock while a rebind can
  write it from a login goroutine: lock it behind a getter, or move account hooks onto the world
  loop. Login goroutines (`LoginUser`, `CopyoverReconnectUser`) change accounts without the world's
  `util.LockMud`; web admin edits run under `RunWithMUDLocked`.
- **Tests.** `accounts.SetTestUser` skips `OnActivate`, so tests that use it get no actor.
- **Slices 5 and 7, text parity** against `cmd/smoke/testdata/pre-4b/bot0.txt`: mask the run tag
  (`smk<tag>u0`), game-clock times, HP/MP prompts and wandering-NPC lines; strip `\r` or pin
  `cmd/smoke/testdata/** text eol=lf` in `.gitattributes`.
- **Slice 5 inherits** the combat bugs kept on purpose in slice 2 (`NewRound_DoCombat.go`, marked
  in comments): the shadowed `userId` in the flee check, `affectedPlayerIds` getting 0 in a mob
  fight, `user` dereferenced before its nil check, a spell fizzle clearing aggro without an event;
  and `KilledBy` is zero for a PvP kill.
- **Permadeath reroll** (`usercommands/suicide.go`): body re-keying and event order are unverified
  live (permadeath is off in `empty`). Check with permadeath on and `LivesStart: 0`. A reroll
  orphans the old body's charmed mobs until the charm expires.
- **Slice 5 (built 2026-10-28) leaves for later** (plan `2026-10-28-phase-4b-5-one-combat-path-plan.md`,
  decisions 1-14; inventory beside it). Calabe's rulings: a hidden attacker reads "Someone"
  everywhere; anyone fighting a fleer may block it and a blocked fleer resumes its pre-flee target
  (one with none ends with no aggro); every swing earns credit. The parity tool for slice 7 is the
  golden combat test (`internal/hooks/combat_golden_test.go`, `util.SeedRandForTest`): it records
  what each player reads; the old pre-4b transcript has no combat (`cmd/smoke/testdata/pre-5/` is
  the fight baseline for reading). Slice 7: the player-death `Broadcast` (screen-reader and refresh
  fields), party XP and per-member sends still use the old path (`// slice 7:` comments). Phase 5:
  turn order is still accounts first, then NPCs (initiative is a rule); NPCs still never start or
  join fights; the `-10` suicide threshold in the round is unreachable (bleed-out death is in
  AutoHeal); a pet's group-hatred duration uses the pet's Perception, and a pet that downs a player
  is recorded as the killer. Messaging gained `Message.NoTargetRoomText` for a target-room line
  that must be silent (remote wait lines). In a dark room combat lines carry by sound and every
  name reads "someone" (slice 3's unseen rule), the attacker's view included.
- **Slice 4 (built 2026-10-28) leaves for later** (plan `2026-10-28-phase-4b-4-one-command-set-plan.md`,
  decisions 1-17; inventory beside it). The facts table above predates it: the mob registry had 21
  commands (`rest` came in slice 2b), plus 3 plugin mob commands, and `events.Input` already carried
  `ActorId`. Calabe's rulings (2026-10-28): action points for every actor; NPCs use the player
  wording; an NPC's unknown command is logged; nobody walks or quits out of a fight (the principle:
  no escaping a fight except by fleeing). Slice 5: the NPC half of the combat round only clears a
  fled NPC's aggro; resolving `flee` for NPCs (they leave) is slice 5's one round; NPC death is the
  `suicide` command's template branch (`usercommands/suicide_npc.go`) until the death rule; an
  NPC's `attack` goes through `CanPvp` against players. Slice 6: input ordering is still per kind
  (no account: strict FIFO; account: a ready command may overtake a delayed one), keyed on
  `Account()`; the idle tick calls `behaviour.RunIdle`, `PathTo`, `Rest`; the waypoint `noop` delay
  lives in the path walking. Slice 7: commands write through `usercommands.Tell` and exclude by
  `ExcludeId` (point them at messaging); the cleanup and follow modules use the same helpers.
  NPCs now receive a room's mutator `PlayerBuffIds` through `MoveToRoom`; rooms are still prepared
  only for players (an NPC's move or look prepares nothing).
- **Slice 3 (built 2026-10-28) leaves for later** (plan `2026-10-28-phase-4b-3-messaging-core-plan.md`,
  decisions 1-11): `Message.TargetRoomText` was added after review so ranged combat can give the
  attacker's and the target's rooms different observer text, as `combat.go:113-137` does. Slice 5
  rules on a hidden attacker: today its target and room messages are suppressed (`combat.go:124`);
  through messaging they would read "someone". Under the Named audience a target that is not named
  gets nothing. Delivery is synchronous and re-entrant: a `Perceive` that calls `Send` nests, and
  `util.Hook.Fire` holds its lock while handlers run, so slice 6 queues NPC reactions rather than
  sending from inside `OnBehaviourPerceive` (two NPCs reacting to each other would recurse).
  Text that varies per receiver beyond the three tokens (adjectives from `GetPlayerName(viewer)`,
  language) has no form yet; slice 7 may need a per-receiver render hook. `Message.Room` is ignored
  when there is an actor. Slices 5 and 7 start a text with `{actor}` when today's line starts with
  the name, so "Someone" capitalises. Slice 7 also moves `main.go`'s raw `net.Conn` "server full"
  writes (the guard does not see them), decides whether telnet and terminal negotiation
  (`term_iac.go`, `term_ansi.go`, `echo.go`) is a second named exception like GMCP, and adds
  interface text before an actor exists.
- **Slice 2c (built 2026-10-28) leaves for later:** item IDs now survive saves, so any path that
  puts one item in two saved places gives two items one ID. The whole-branch review found none on
  normal play paths; the known ones (corpse looting with `CorpseItems` on, admin room API calls) are
  in `docs/known-issues.md`. A saved-room NPC that walks into an ephemeral room leaves its old copy
  in the first room's sidecar, and the ephemeral room never saves, so a restart restores the old
  copy with items it has since given away (reachability unverified: no content found with such an
  exit); crash windows between the user and room saves do the same. A duplicate-ID check logged at
  save time would show these. The admin user page still sends items by `ItemId` only, so of two
  identical items the first is kept when one is removed.
- **Slice 2b gaps:** a fresh spawn reaches disk only at its room's next save; a crash between a slot
  settle's two writes leaves an unread ledger entry; NPCs live in a deleted room or zone are saved
  nowhere (upstream behaviour); a room delete can lose an NPC in a crash if the deleted room held an
  old copy of an NPC now in an ephemeral or unloaded room. The admin rooms page and mapper lost the
  spawn "Permanent Buffs" picker (decision 17): passes JSHint, not checked in a browser.
