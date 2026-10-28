# Phase 4b slice 4: one command set and one input path (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** players and NPCs run the same commands through the same input path. A command takes the
actor that runs it; whatever needs an account asks the actor for one. NPCs move, pay action points,
speak and handle items by the same rules as players.

**Architecture:** one registry of `usercommands.Command func(rest string, actor *actors.Actor, room
*rooms.Room, flags events.EventFlag) (bool, error)`. The mob command package goes: its commands fold
into the user versions (one behaviour per command, differences settled below), and the four that are
not actions an actor takes become engine calls or behaviour actions in a new `internal/behaviour`
package. One `rooms.MoveToRoom(actorId, ...)` moves any body. One router in `world.go` handles every
`events.Input` and runs one `processInput`; the steps that only make sense with an account run when
the actor has one. Action points regenerate for every actor.

**Tech stack:** Go, testify. No new dependency.

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md`, decisions 1-4, section 1 rules,
section 3 ("One command set", "One input path"), section 4 slice 4. Inventory:
`2026-10-28-phase-4b-4-inventory.md` (A: mob vs user commands, B: user command surface, C: input
paths), referred to below as Inv A, Inv B, Inv C with their section numbers. Branch:
`strip/phase-4b-4-one-command-set` (created; the first commit moved the slice 3 plan to
`docs/completed/`).

---

## Facts verified against source (2026-10-28, master `48a7d92b`)

The inventories hold the full tables with file:line; these rows are the ones the tasks build on,
re-checked by the controller.

| Fact | Value | Where |
|---|---|---|
| Registries | User: 65 built-in commands in `usercommands.go:52-121` plus 6 plugin commands (`trash`, `bury`, `follow`, `mudletmap`, `mudletui`, `checkclient`). Mob: 21 in `mobcommands.go:24-46` (the spec says 20; `rest` came in slice 2b) plus 3 plugin mob commands (`trash`, `bury`, `follow`). `mobcommands/unlock.go` is unregistered | Inv A facts, Inv B facts |
| Overlap | 14 names in both (`broadcast drop emote equip get give go look noop put remove say shout suicide`); mob-only: `despawn pathto replyto rest sayto saytoonly wander` | Inv C 3 |
| Files | `internal/usercommands` has 68 non-test `.go` files and no tests | Inv B |
| Account use | 31 files need an account (21 `admin.*`, plus `alias history macros party password quit save set start suicide`); 3 mixed (`attack`, `go`, `look`); the rest work for any actor | Inv B 1 |
| Output | 688 `user.SendText(` in usercommands; 69 room sends excluding the speaker by `user.UserId`; NPCs never receive room text today (`hooks/Message_SendMessages.go:45`) | Inv B 3 |
| Router | `World.HandleInputEvents` (`world.go:78-190`) is the only `events.Input` listener; it classifies by `actors.AccountOf(id) == nil`; NPC commands are strict FIFO (tracker set before the ready check, `world.go:101-114`), player commands are not (`world.go:128-187`) | read |
| Paths | `processInput(userId, text, flags)` (`world.go:502-611`), `processMobInput(mobId, text)` (`world.go:613-653`); account-only steps listed in Inv C 2 | read |
| Movement | `rooms.MoveToRoom(userId, toRoomId, isSpawn...)` (`roommanager.go:269-396`, 12 callers) works on `user.Character` apart from `MarkUserVisited(userId)`, `RemovePlayer/AddPlayer(userId)`, the `roomsWithUsers` count, `Prepare(true)` when the first player arrives, and `user.AddBuff` for mutator `PlayerBuffIds` (an `events.Buff` for the body). Mob moves use `Room.RemoveMob`/`AddMob` (`rooms.go:725-757`) | read |
| Action points | `hooks/NewTurn_ActionPoints.go` adds 1 AP per turn to `accounts.GetAllActiveUsers()` only; `ActionPointsMax` is computed for every character, at least 50 (`character.go:1418-1432`); `DeductActionPoints` (`character.go:236`) | read |
| Plugins | `Plugin.AddUserCommand(name, usercommands.UserCommand, allowWhenDowned, isAdminOnly)` (`plugins.go:297`), `AddMobCommand` (`:311`); callers `modules/cleanup/cleanup.go:55-59`, `modules/follow/follow.go:61-62`, `modules/gmcp/gmcp.Mudlet.go:73-75` | grep |
| Behaviour producers | `wander` from data idle commands (`mobs/startland/1-rat.yaml:10`, spawn override `rooms/startland/1.yaml:20`); `pathto` from `hooks/MobIdle_HandleIdleMobs.go:29,40`, `NewRound_IdleMobs.go:94,99`, `mobcommands/wander.go:23`, `mobcommands/go.go:69`; `rest` from `NewRound_IdleMobs.go:47` and the charm-expiry constant (`characters/charminfo.go:7`, nothing sets a charm today); `despawn` from nothing but admin `command`; `suicide` to NPCs from `NewRound_MobRoundTick.go:70`, `NewRound_DoCombat.go:927`, `Buff_ApplyBuffs.go:66-69` | Inv A 2, 3 |
| Recent deaths | `mobs.TrackRecentDeath` (one caller, `mobcommands/suicide.go:36`) and `mobs.RecentlyDied` (one reader, `world.go:618`, to silence a log) | grep |
| Baseline | 36 packages ok, 3 known Windows failures | `geomud-verifying` |

## Decisions

Calabe, 2026-10-28 (slice 4 planning):

1. **One mechanism for action points.** Action points regenerate for every actor and every actor
   pays the same costs. (Today only players regenerate; upstream's per-turn hook loops over online
   users.)
2. **NPCs use the player wording.** Folding an NPC command into the user command gives NPCs the
   user command's text (their names keep the `mobname` colour). Example: an NPC shout is upper case
   and carries its words to adjacent rooms, as a player's does.
3. **An NPC's unknown command is logged, not shown.** A warning in the server log names the NPC and
   the command; the room sees nothing (today the room reads `looks a little confused (cmd rest).`).
4. **Nobody walks out of a fight; they flee.** An actor in a fight cannot `go`, player or NPC: it
   must `flee`. In a fight means the actor has an aggro target, or a body in its room has the actor
   as its aggro target. (Today `go` checks only the actor's own target, `go.go:19-22`, so a player
   an NPC attacks, who has not attacked back, can walk away; NPCs have no check at all.) Being shot
   at from another room does not hold an actor in its room.

Defaults set by the controller from the rulings above, decision 2 of the spec (behaviour parity)
and "one rule for every actor" (Calabe reviews them with this plan):

5. **One move.** `rooms.MoveToRoom` takes the body's ID and moves any actor. What only makes sense
   for a player stays keyed on the account: the user visitor trail (`MarkUserVisited`, which
   `findUserExit` reads), the `roomsWithUsers` count, and `Prepare(true)` when the first player
   enters (spec decision 16: rooms are prepared for players). NPCs get the rest, including a room's
   mutator `PlayerBuffIds`, death-room instancing and reset rooms.
6. **What decision 4 removes.** The pursuit block in `go` (`go.go:245-273`: NPCs fighting a player
   roll to follow when the player walks away) can no longer run, because walking away from someone
   fighting you is refused; it is deleted. Fleeing escapes pursuers, as today (the flee code makes
   no NPC follow). A charmed NPC fighting beside a player who flees no longer walks after them
   (`NewRound_DoCombat.go:148-155`): it is in the fight, so it flees on its own roll (queue `flee`
   for it there). An NPC attacked from another room still walks toward its attacker
   (`NewRound_DoCombat.go:473-480`): it fights no one in its own room yet. Nothing charms an NPC
   today, so the charmed change has no live effect. `flee` only sets the intent; the combat round
   resolves it, and only the player half of the round does today (`NewRound_DoCombat.go:68`; the NPC
   half handles `DefaultAttack` only, `:556`). So an NPC that flees stops attacking but does not leave
   until slice 5's one combat round resolves flee for every actor; Task 10 records that for slice 5.
7. **NPCs follow the player rules for locks, keys and sneaking** (they unlock with keys they carry;
   a hidden NPC moves silently; the far side of a locked exit is unlocked as for players).
8. **Commands nothing issues are dropped**, not ported: `sayto`, `saytoonly`, `replyto`,
   `equip all` / `equip random`, `go <roomId>` (teleport), `go home`, the `vanish` death form, the
   MobId 38 ("The Guide") death case, the `secretly` look prefix, `despawn`, `wander loot` /
   `wander players` arguments, and `mobs.Sleep` (no caller). Admin `command <mob> despawn` stops
   working (admins end an NPC from the web admin's NPC page, slice 2b).
9. **Not commands any more.** `rest` (the boredom and charm-expiry unload, slice 2b decision 15)
   becomes `behaviour.Rest(mob, reason)`. `wander` and `pathto` become `behaviour.Wander(mob)` and
   `behaviour.PathTo(mob, args)`: they choose or plan, and the body moves by issuing `go <exit>`
   through the input queue as today. An idle command reading `wander` or `pathto ...` in data runs
   the behaviour action. `noop` stays one shared command (its point is the delay the input queue
   gives it).
10. **Death keeps one command until slice 5.** `suicide` stays the command combat, buffs and the
   round tick issue. It branches on the spec's death question, "does this actor return from a
   template" (`mobs.GetInstance(actor.Id) != nil`): the NPC branch is today's mob death, moved as
   is; the player branch is today's death stand-in. Slice 5 replaces both with the death rule.
11. **Command output in this slice.** Commands write to their actor through one helper,
    `tell(actor, text)`: the account's `SendText` when the actor has an account, nothing otherwise.
    Players read exactly today's text. Slice 7 points `tell` at `messaging`. Room sends keep their
    exclusion list as user IDs, through `excludeId(actor)` (the account's UserId, or 0).
12. **Dispatch order** is the user order for everyone: exit name, personal alias (account),
    keyword alias, `cancel-on-action` buffs, `look` self-keywords, plaintext-password gate
    (account), registry (downed check with its message; input-blocked check (account); admin-only
    needs an account with the permission, so an NPC never runs one), emote alias, `go` fallback.
13. **Ordering stays per kind for now** (spec decision 2): an actor without an account is strict
    FIFO; an actor with one keeps today's rules (a ready command may overtake a delayed one;
    `ReadyTurn 0` runs at once). The router asks `Account()`, not the controller's type. Slice 6
    revisits it with the one tick.
14. **Account-only input steps** (prompt answering, macros, `SetLastInputRound`, the backtick and
    dot shortcuts stay for all, bad-input tracking, the "not recognized" text and confused emote,
    the prompt redraw) run when the actor has an account. Input for a body with no live actor is
    dropped with a debug log; `mobs.TrackRecentDeath`/`RecentlyDied` go.
15. **Keep from the mob versions:** the `perma-gear` buff flag refuses equip, remove and item drops
    for any actor; re-equipping the item already worn is a silent no-op for any actor; the nil check
    in `give`; clearing an NPC follower's `Path` when it starts following. **Fixed by taking the
    user version:** mob `put` minted gold and ignored locks; mob `get` destroyed an item a full NPC
    could not carry; mob `remove` ignored curses and bound items.
16. **Plugin API.** `AddUserCommand` takes a `usercommands.Command`; `AddMobCommand` goes. The
    cleanup and follow modules keep one handler per command; account-only parts (XP grant, parties)
    ask for the account.
17. **Small fixes in code this slice rewrites:** `EventFlag.Add` clears the flag instead of setting
    it (`events/const.go:9-11`); `input.Flags.Remove` on a local copy (`world.go:163`); dead
    `World.eventRequeue` and `WorldInput.ReadyTurn`; the "per turn" comment on the router (the
    limit is per event-loop pass). The stuck `inputBlocked` edge for a sub-turn
    `CmdBlockInputUntilComplete` (Inv C 1c) gets a test; fix it only if the test shows it.

## File structure

| File | Change |
|---|---|
| `internal/hooks/NewTurn_ActionPoints.go` (+ test) | regenerate AP for every actor |
| `internal/mobs/mobs.go` | a new NPC starts with full AP; `Sleep`, `TrackRecentDeath`, `RecentlyDied` go |
| `internal/rooms/roommanager.go` (+ test) | `MoveToRoom(actorId ids.ID, ...)` for any actor; 12 callers updated |
| `internal/events/const.go` | `EventFlag.Add` fixed |
| `internal/usercommands/usercommands.go` | `Command`, `CommandAccess`, `TryCommand(cmd, rest, actorId, flags)`, `tell`, `excludeId` |
| `internal/usercommands/*.go` (68 files) | the actor signature |
| `internal/usercommands/suicide.go`, new `suicide_npc.go` | death branches |
| `internal/usercommands/emote.go`, `say.go`, `shout.go`, `broadcast.go`, `drop.go`, `get.go`, `give.go`, `put.go`, `equip.go`, `remove.go`, `look.go`, `go.go` | fold differences (Decision 15) |
| `internal/behaviour/behaviour.go` (new, + tests) | `Wander`, `PathTo`, `Rest`, `RunIdle` |
| `internal/hooks/MobIdle_HandleIdleMobs.go`, `NewRound_IdleMobs.go`, `NewRound_MobRoundTick.go`, `NewRound_DoCombat.go` | call behaviour actions; charmed NPCs flee with their master |
| `world.go` | one router and one `processInput` |
| `internal/mobcommands/` | deleted |
| `internal/plugins/plugins.go`, `plugincallbacks.go` | one command registration |
| `modules/cleanup`, `modules/follow`, `modules/gmcp/gmcp.Mudlet.go` | the actor signature |
| `internal/usercommands/admin.command.go`, `_datafiles/.../templates/admincommands/help/command.command.template`, `_datafiles/html/admin/docs-modules.html` | no mob command list |
| `internal/guards/commands_test.go` (new) | guards for this slice |
| `cmd/smoke/admin.go` | an NPC runs a command through the shared path |
| records | Task 10 |

---

### Task 0: Inventory in the repo

- [ ] **Step 1:** Commit the inventory beside this plan (the controller wrote
  `2026-10-28-phase-4b-4-inventory.md`) and this plan:

```bash
git add 2026-10-28-phase-4b-4-inventory.md 2026-10-28-phase-4b-4-one-command-set-plan.md
git commit -F - <<'EOF'
Phase 4b slice 4 plan and inventory: one command set and input path

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 1: Action points for every actor (Decision 1)

**Files:** Modify `internal/hooks/NewTurn_ActionPoints.go`, `internal/mobs/mobs.go`; test
`internal/hooks/actionpoints_test.go` (new).

- [ ] **Step 1: Failing test.** Spawn an NPC actor (as `actors` tests do: a `mobs.Mob` through
  `mobs.NewTestInstance`, which fires `OnInstanceCreated` and spawns its actor when the `actors` wire
  is registered; check `internal/actors/wire_test.go` for the setup it needs, e.g. `actors.Wire()`),
  set `Character.ActionPoints = 0` and `ActionPointsMax.Value = 100`, call
  `ActionPoints(events.NewTurn{})`, and assert the NPC's AP is 1; call it 200 times and assert 100.
  Run `go test ./internal/hooks/ -run ActionPoints -count=1`: FAIL (NPC stays at 0).

- [ ] **Step 2: Implement.**

```go
// ActionPoints gives every live body one action point per turn, up to its
// maximum (Phase 4b slice 4 decision 1: one mechanism for players and NPCs).
func ActionPoints(e events.Event) events.ListenerReturn {
	for _, a := range actors.All() {
		c := a.Character
		if c == nil {
			continue
		}
		c.ActionPoints += 1
		if c.ActionPoints > c.ActionPointsMax.Value {
			c.ActionPoints = c.ActionPointsMax.Value
		}
	}
	return events.Continue
}
```

  Drop the unused `accounts` import and the commented-out block. In `mobs.NewMobById`, where it
  sets `Health` and `Mana` to their maximums, add
  `mob.Character.ActionPoints = mob.Character.ActionPointsMax.Value` (a new body starts rested,
  as it starts at full health). Check `actors.All()` takes the registry read lock and that the
  NewTurn handler runs on the world loop (it does: Inv C 5).

- [ ] **Step 3:** Run the test and `go build ./...`. Expected: PASS. Note in the report how many
  actors `All()` returns on a loaded world is unverified until smoke.

- [ ] **Step 4: Commit** `hooks: action points regenerate for every actor; a new NPC starts rested`.

---

### Task 2: One move for any actor (Decision 5)

**Files:** Modify `internal/rooms/roommanager.go` (`MoveToRoom`), its 12 callers; test
`internal/rooms/move_test.go` (new).

- [ ] **Step 1: Failing tests** (rooms package, internal test): put two rooms in memory with
  `AddTestRoom` and an exit between them; spawn an NPC actor and a player actor (an
  `accounts.Account` with a `Character` and an ID, bound as `actors` tests do); call
  `MoveToRoom(npc.Id, 2)` and assert the NPC's `Character.RoomId == 2`, it is in room 2's
  `GetBodies()`, not in room 1's, an `events.RoomChange` with its ID was queued (register a
  listener and `events.ProcessEvents()`), and room 2 was not `Prepare`d (no player arrived). Same
  for the player, asserting the user visitor trail is set on room 1 and room 2 was prepared. Run:
  FAIL (signature).

- [ ] **Step 2: Implement.** Change the signature to
  `func MoveToRoom(actorId ids.ID, toRoomId int, isSpawn ...bool) error`. At the top:

```go
	a := actors.Get(actorId)
	if a == nil || a.Character == nil {
		return fmt.Errorf("actor %s not found", actorId)
	}
	char := a.Character
	acc := a.Controller().Account() // nil for an NPC
```

  Replace every `user.Character` with `char`. Keep account-keyed only (decision 4):
  `MarkUserVisited`, `RemovePlayer`/`AddPlayer` and the `roomsWithUsers` bookkeeping, and the
  `Prepare(true)` on first player, all under `if acc != nil`. For an NPC use `MarkMobVisited`,
  `RemoveMob(actorId)` and `AddMob(actorId)` (which already queues `RoomChange` and sets RoomId and
  Zone; do not queue a second `RoomChange`, so for NPCs skip the function's own `RoomChange` and its
  `RoomId`/`Zone` assignments, or restructure so exactly one `RoomChange` is queued per move; read
  `AddMob` and say which). Mutator `PlayerBuffIds`: replace `user.AddBuff(buffId, "area")` with
  the same event for the body: read `Account.AddBuff` (`userrecord.go:318-331`) and queue the same
  `events.Buff{ActorId: char.Id, BuffId, Source}` directly. `RememberRoom`/`MarkVisitedRoom` stay for
  every actor.

  Callers: `grep -rn 'MoveToRoom(' --include=*.go .` lists 12 non-test calls; change each
  `rooms.MoveToRoom(user.UserId, ...)` to `rooms.MoveToRoom(user.Character.Id, ...)` (the commands
  are converted to actors in Task 4, which then passes `actor.Id`).

- [ ] **Step 3:** Run the tests, `go build ./...`, `go test ./internal/rooms/ ./internal/usercommands/ ./internal/hooks/ -count=1`. Expected: PASS.

- [ ] **Step 4: Commit** `rooms: MoveToRoom moves any actor (account-keyed parts kept for players)`.

---

### Task 3: The `Command` type, actor dispatch, `tell` (Decisions 11, 12)

**Files:** Modify `internal/usercommands/usercommands.go`, `world.go:580` (the one external
`TryCommand` caller); test `internal/usercommands/usercommands_test.go` (new).

- [ ] **Step 1:** Read `usercommands.go` in full and Inv B 2.

- [ ] **Step 2: Types.** Add (keep `UserCommand` for now; Task 4 removes it):

```go
// Command is one command any actor can run (Phase 4b slice 4): the rest of
// the input line, the actor, its room and the input flags.
type Command func(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error)

// accountCommand adapts a command not yet converted to the actor signature:
// it runs only for an actor with an account. Removed when Task 4 ends.
func accountCommand(f UserCommand) Command {
	return func(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
		user := actor.Controller().Account()
		if user == nil {
			return false, nil
		}
		return f(rest, user, room, flags)
	}
}

// tell writes text to the actor: to its account when it has one, nowhere
// otherwise (Decision 11; slice 7 points this at messaging).
func tell(actor *actors.Actor, text string) {
	if user := actor.Controller().Account(); user != nil {
		user.SendText(text)
	}
}

// excludeId is the ID a room send excludes for actor: its account's UserId,
// or 0 for an NPC, which never receives room text (Decision 11).
func excludeId(actor *actors.Actor) int {
	if user := actor.Controller().Account(); user != nil {
		return user.UserId
	}
	return 0
}
```

  `CommandAccess.Func` becomes `Command`; wrap every registry entry as
  `accountCommand(Attack)` and so on, so the build stays green.

- [ ] **Step 3: `TryCommand(cmd, rest string, actorId ids.ID, flags events.EventFlag) (bool, error)`.**
  Look up `actors.Get(actorId)` (nil: error), `user := actor.Controller().Account()` (may be nil),
  the room from `actor.Character.RoomId`. Keep today's steps in today's order (Inv B 2), with each
  account step under `if user != nil`: personal aliases, the plaintext-password gate, the
  input-blocked check. Admin-only: refuse when `user == nil` (return `true, nil` silently, as a
  refused non-admin is handled) or the permission check fails as today. Downed: `actor.Character`.
  The downed message goes through `tell`. Calls into commands pass `actor`.
  `world.go:580` becomes `usercommands.TryCommand(command, remains, user.Character.Id, flags)`.
  Check `RegisterCommand` (plugins path) and any other function in the file that takes a
  `UserCommand`; adapt it the same way and list each.

- [ ] **Step 4: Tests** in `usercommands_test.go` (needs `mudlog.SetupLogger(nil, "", "", false)`
  in a `TestMain` if any path logs): register a test command in the registry for the test
  (`userCommands["zz_test"] = CommandAccess{Func: f}` with cleanup), and check: an NPC actor runs a
  non-admin command; an NPC actor is refused an admin-only command; a player actor without the
  permission is refused; `tell` to an NPC actor does nothing and does not panic; `excludeId` is the
  UserId or 0. Build NPC and player actors as in Task 2. Run: PASS.

- [ ] **Step 5:** `go build ./... && go test ./internal/usercommands/ ./internal/guards/ -count=1`.

- [ ] **Step 6: Commit** `usercommands: one Command type dispatched by actor (adapter for unconverted commands)`.

---

### Task 4: Convert every command to the actor signature (Decision 11)

Split across subagents by group; each group is its own commit. The conversion rules, the same for
every group:

- Signature: `func X(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error)`.
- **A command that needs an account** (Inv B 1 classification "needs an account"): first line
  `user := actor.Controller().Account(); if user == nil { return false, nil }`, then the body
  unchanged. Nothing else changes in that file.
- **A command that works for any actor:** `user.Character` becomes `actor.Character`;
  `user.SendText(` becomes `tell(actor, `; `user.UserId` in a room send's exclusion list becomes
  `excludeId(actor)`; any other use of `user` (Inv B 1 lists each with its line) goes inside
  `if user := actor.Controller().Account(); user != nil { ... }`, keeping the account-only behaviour
  for players. `user.Command(...)` / `CommandFlagged(...)` (queue input for the body) become a call
  on a small helper in `usercommands.go`, `queue(actor, text, wait...)`, that queues
  `events.Input{ActorId: actor.Id, ...}` exactly as `Account.Command` does (read `userrecord.go:277-322`;
  for `CommandFlagged`'s `BlockInput`, keep it under the account check).
- **Mixed** (`attack`, `go`, `look`): as "any actor", with the listed account parts under the check.
  `go`: the move is `rooms.MoveToRoom(actor.Id, ...)`. `look`: `rooms.GetDetails` takes an account;
  pass it when there is one, and for an NPC skip the room view (it would only be text for nobody).
- Remove the `accountCommand(...)` wrap from the registry entry of each converted command. The group
  that converts the last command deletes `accountCommand` and `UserCommand`.
- No behaviour change for players: same text, same order, same events. Each group runs
  `go build ./...` and `go test ./internal/usercommands/ -count=1` and greps its files for
  `user\.` uses outside an account check.

Groups (file lists from Inv B 1; each subagent reads its files in full, in chunks of at most 400
lines, and reports "read N of N files" with the list):

- [ ] **4a: the 21 `admin.*` files** (all "needs an account"). Commit `usercommands: admin commands take the actor`.
- [ ] **4b: the 10 account files** `alias history macros party password quit save set start suicide`
  ("needs an account"; `suicide` gets its NPC branch in Task 5, so here it refuses NPCs). Commit
  `usercommands: account commands take the actor`.
- [ ] **4c: every other file except the fold targets of Task 5** (`broadcast drop emote equip get give
  go look noop put remove say shout`), including `attack`. Commit `usercommands: actor commands`.
- [ ] **4d: the 13 fold targets**, converted with the rules only (no fold differences yet). Delete
  `accountCommand` and `UserCommand`. Commit `usercommands: fold targets take the actor; adapter gone`.

After 4d, `grep -n 'accountCommand\|UserCommand' internal/usercommands` prints nothing outside
comments, and `internal/plugins`, `modules/*` still build (Task 7 converts them; until then they
need a temporary wrap: if `go build` fails there, add the same adapter inside `plugins.go` for
`AddUserCommand` and note it for Task 7).

---

### Task 5: Fold the mob commands into the user commands (Decisions 2, 4-10, 15)

One subagent per cluster, in order, each its own commit. Each reads the mob file and the user file
in full and Inv A section 1's row, applies the decisions, and lists every behavioural difference with
what it did. Text for players is unchanged unless a row below says so.

- [ ] **5a: speech** (`say`, `shout`, `emote`, `broadcast`). Take the user versions for every actor
  (Decision 2): no "no players in room" early return, mute and drunk checks stay account or
  buff-based as they are, self echo through `tell`. NPC names in room text use the `mobname`
  colour: where a user command writes `<ansi fg="username">%s</ansi>` with the actor's name, use a
  helper `nameColour(actor)` returning `username` or `mobname` by whether the actor has an account.
  `saytext-mob`: check `_datafiles/.../ansi-aliases.yaml`; if it differs from `saytext`, keep it for
  NPC say text through the same helper and say so, else drop it. Emote alias tables: one table (they
  are byte-identical, Inv A facts). Tests: an NPC actor's `say` reaches a player in the room with the
  NPC name in `mobname`; an NPC `shout` reaches the next room with its words.
- [ ] **5b: items** (`drop`, `get`, `give`, `put`, `equip`, `remove`). User versions for all; keep
  Decision 15's mob rules (`perma-gear` refusals with the user command's refusal text for players and
  the emote `struggles with their gear for a while, then gives up.` for any actor, matching today's
  NPC line; same-item re-equip is a silent no-op; nil check in `give`). Drop `equip all`/`random`.
  Tests: an NPC with a full backpack does not destroy an item it cannot carry; an NPC `put` of gold
  it lacks is refused; `perma-gear` blocks `drop` for a player.
- [ ] **5c: `look`.** User version; an NPC `look` produces nothing (Task 4 rule) but still runs the
  side effects the user version has (check: the `CmdSecretly` path, noun discovery). Drop the
  `secretly` prefix.
- [ ] **5d: `go`** (Decisions 4-7, Inv A row `go`). User version for every actor, with:
  - **In a fight** (Decision 4): replace the `Aggro != nil` check with a helper in `usercommands`,
    `inFight(actor, room) bool`: true when `actor.Character.Aggro != nil`, or when any body in
    `room.GetBodies()` (other than the actor) has `Aggro != nil && Aggro.Target == actor.Id`. The
    refusal text for players stays `You can't do that! You are in combat!`; check whether any other
    command uses the same `Aggro != nil` test to mean "in combat" for leaving the room (grep
    `Aggro != nil` in `usercommands`; `flee` itself must still work in a fight) and list them in the
    report without changing them.
  - The AP cost applies to every actor (Task 1).
  - The mob waypoint `noop` delay moves to the behaviour that issued the step
    (`NewRound_IdleMobs.go` path walking: after the last step it queues `noop` with the 1 s wait,
    as `mobcommands/go.go:130-135` does).
  - Delete the pursuit block (`go.go:245-273`, Decision 6). Keep the charmed-follow block
    (`go.go:225-237`) for an actor that walks out of no fight (a charmed NPC follows its master's
    exit; a charmed NPC in a fight is refused by its own `go`, which is the rule). Party follow and
    hostile notice apply to whoever moved.
  - `NewRound_DoCombat.go:148-155` (charmed NPCs follow a fleeing player by the exit name): queue
    `flee` for each such NPC instead (Decision 6). `NewRound_DoCombat.go:473-480` (an NPC attacked
    from another room walks toward its attacker) stays: `inFight` is false for it.
  - Tests: an NPC with an aggro target is refused `go`; a player who is an NPC's aggro target in
    the same room, with no target of their own, is refused `go` (the new player behaviour); a
    player whose NPC attacker is in another room may `go`; an NPC with too few AP is refused (`go`
    prints nothing for it).
- [ ] **5e: `suicide`** (Decision 10). Move `mobcommands/suicide.go`'s body into
  `usercommands/suicide_npc.go` as `npcDeath(actor *actors.Actor, mob *mobs.Mob, room *rooms.Room)`,
  dropping `vanish` and the MobId 38 case and `TrackRecentDeath`; `Suicide` calls it when
  `mobs.GetInstance(actor.Id) != nil`, else runs today's player branch (with Task 4b's account
  check). Test: an NPC actor's `suicide` destroys the instance and fires `events.MobDeath`.

---

### Task 6: Behaviour actions (Decisions 8, 9)

**Files:** Create `internal/behaviour/behaviour.go`, `behaviour_test.go`. Modify
`hooks/MobIdle_HandleIdleMobs.go`, `hooks/NewRound_IdleMobs.go`, `hooks/NewRound_MobRoundTick.go`.

- [ ] **Step 1:** Read `mobcommands/wander.go`, `pathto.go`, `rest.go`, `despawn.go`, the three hooks,
  and Inv A 2. `behaviour` imports `mobs`, `rooms`, `mapper` (check `mapper` does not import a
  package that would cycle; `go list -deps`), `mudlog`; it issues movement with `mob.Command("go "+exit)`.
- [ ] **Step 2: API** (each with a doc comment saying it replaces the mob command of the same name):

```go
// Wander picks an exit for mob as the wander command did and issues
// "go <exit>"; past MaxWander it plans the way home instead.
func Wander(mob *mobs.Mob)

// PathTo plans mob's path to the listed rooms, or "home", as the pathto
// command did. It does not move: the idle round tick walks the path.
// A path that cannot be found is logged (Decision 3) and marks the NPC lost,
// as before.
func PathTo(mob *mobs.Mob, args string)

// Rest takes mob out of memory without ending it (slice 2b decision 15), as
// the rest command did: an NPC in an ephemeral room ends there; any other
// rests in its room.
func Rest(mob *mobs.Mob, reason string)

// RunIdle runs one idle command from data: "wander" and "pathto ..." are
// behaviour actions; anything else is queued as the NPC's input.
func RunIdle(mob *mobs.Mob, command string)
```

  Bodies: today's `Wander`, `Pathto`, `Rest` moved as is, minus `wander loot`/`players` and with
  `mob.Command("pathto home")` calls replaced by `PathTo(mob, "home")`. `Rest` needs room functions
  (`room.RestNPC`, `EndEphemeralNPC`, `CleanupMobSpawns`); they are in `rooms`, which `behaviour` may
  import.
- [ ] **Step 3: Producers.** `MobIdle_HandleIdleMobs.go`: `pathto home` calls become
  `behaviour.PathTo(mob, "home")`; the idle command call becomes `behaviour.RunIdle(mob, cmd)`.
  `NewRound_IdleMobs.go`: boredom `rest depression` becomes `behaviour.Rest(mob, "depression ...")`;
  `pathto` calls become `behaviour.PathTo`. `NewRound_MobRoundTick.go:59`: the charm expired command
  is a `;`-separated command string; when it is `characters.CharmExpiredRest`, run its emote part
  as the NPC's input and then call `behaviour.Rest` (read the constant; keep the emote visible before
  the NPC leaves, e.g. by running the emote synchronously through `usercommands.TryCommand` before
  `Rest`). Nothing sets a charm today, so test it directly.
- [ ] **Step 4: Tests** (`behaviour_test.go`): `RunIdle(mob, "wander")` on an NPC in a room with one
  exit in its zone queues `go <that exit>` (register an `events.Input` listener); `MaxWander` 0 queues
  nothing; `PathTo` with an unknown room logs and marks `lost`; `Rest` takes the NPC out of memory and
  into its room's resting list (reuse the rooms persist setup patterns; if `Rest` cannot be tested
  from outside `rooms`, test it through `rooms` helpers and say so).
- [ ] **Step 5:** `go build ./...`, tests. Commit `behaviour: wander, pathto and rest are behaviour actions`.

---

### Task 7: One input path; mob commands deleted (Decisions 3, 13, 14, 16, 17)

**Files:** Modify `world.go`, `internal/plugins/plugins.go`, `plugincallbacks.go`,
`modules/cleanup/cleanup.go`, `modules/follow/follow.go`, `modules/gmcp/gmcp.Mudlet.go`,
`internal/usercommands/admin.command.go`, the admin `command` help template,
`_datafiles/html/admin/docs-modules.html`, `internal/mobs/mobs.go`, `internal/events/const.go`.
Delete `internal/mobcommands/`.

- [ ] **Step 1: Router.** Replace `HandleInputEvents` with one path:

```go
func (w *World) HandleInputEvents(e events.Event) events.ListenerReturn {
	input, typeOk := e.(events.Input)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Input", "Actual Type", e.Type())
		return events.Continue
	}

	actor := actors.Get(input.ActorId)
	if actor == nil {
		// A body with no live actor: a player who logged out, an NPC that
		// died, a rerolled body (Decision 14).
		mudlog.Debug("input", "event", "dropped", "actorId", input.ActorId.String(), "reason", "no live actor")
		return events.Continue
	}
	turnCt := util.GetTurnCount()

	// Decision 13: an actor without an account runs its commands strictly in
	// order, so a delayed command holds back the ones after it; an actor with
	// one keeps the player rules. The limit is one queued command per actor
	// per event-loop pass, not per turn.
	if actor.Controller().Account() == nil {
		if _, ok := w.inputTracker[actor.Id]; ok {
			return events.CancelAndRequeue
		}
		if input.ReadyTurn == 0 {
			w.processInput(actor, input.InputText, input.Flags)
			return events.Continue
		}
		w.inputTracker[actor.Id] = struct{}{}
		if input.ReadyTurn > turnCt {
			return events.CancelAndRequeue
		}
		w.processInput(actor, input.InputText, input.Flags)
		return events.Continue
	}

	// ... the player branch as today (world.go:125-189), keyed by actor.Id
	// in w.inputTracker and w.ignoreInput instead of the UserId, calling
	// w.processInput(actor, ...).
}
```

  Write the player branch out in full from `world.go:125-189` with those substitutions (the
  `CmdUnBlockInput` handling, the requeue, `ignoreInput`, the tracker set after running). The two
  trackers become one `inputTracker map[ids.ID]struct{}`; `ignoreInput` is keyed by `ids.ID`. Clear
  `inputTracker` where `world.go:722-723` clears the two today. Fix `input.Flags.Remove` on the local
  copy (`world.go:163`): it never reached the requeued event; check whether removing it changes
  anything (it did nothing) and delete it, saying so.
- [ ] **Step 2: `processInput(actor *actors.Actor, inputText string, flags events.EventFlag)`.** One
  function from today's `processInput` (`world.go:502-611`) with `user := actor.Controller().Account()`:
  the prompt, macro, `SetLastInputRound`, empty-input prompt redraw, bad-input tracking, the
  "not recognized" text and `emote @looks a little confused`, and the final prompt redraw run when
  `user != nil`. `TrimSpace`, the shortcut split and the command split run for everyone. Unknown
  command with no account: `mudlog.Warn("input", "event", "unknown command", "actorId", ..., "name", actor.Character.Name, "command", command, "rest", remains)` and nothing else (Decision 3). Dispatch:
  `usercommands.TryCommand(command, remains, actor.Id, flags)`. Delete `processMobInput`.
- [ ] **Step 3: Delete the mob command package** (`git rm -r internal/mobcommands`). Fix every
  reference the build reports: `world.go`, `admin.command.go` (its help lists mob commands: list the
  registry's commands that an NPC can run instead, or drop the list; update
  `templates/admincommands/help/command.command.template` lines 6-8 and 13-16 to say NPCs run the same
  commands as players), `plugins.go`, `plugincallbacks.go`.
- [ ] **Step 4: Plugins and modules** (Decision 16). `AddUserCommand(command string, handlerFunc
  usercommands.Command, allowWhenDowned bool, isAdminOnly bool)`; delete `AddMobCommand` and the
  `mobCommands` callback map. `modules/cleanup`: one `trash` and one `bury` handler on the actor
  signature; the XP grant (`user.GrantXP`) under the account check; NPC use keeps the mob handler's
  behaviour (silent removal plus `ItemOwnership`) where it differs only by texts going through
  `tell`. `modules/follow`: one `follow` handler; keep the NPC `Path.Clear()` when an NPC starts
  following (`mobs.GetInstance(actor.Id)`); fix the "stopped following you" name bug
  (`follow.go:457, 584`, Inv A plugin table) with a test if the module has tests, else say so.
  `modules/gmcp/gmcp.Mudlet.go`: the three commands need an account (`ConnectionId`): first-line
  account check. Update `_datafiles/html/admin/docs-modules.html` (the "Mob Commands" table at
  about 631-643, the follow example at about 1554, prose at about 367 and 1505) to the one API.
- [ ] **Step 5: Leftovers.** Delete `mobs.Sleep`, `TrackRecentDeath`, `RecentlyDied` and their data
  (check nothing else reads them). Delete `World.eventRequeue` and `WorldInput.ReadyTurn` if still
  unused. Fix `EventFlag.Add` (`events/const.go:9-11`, it clears instead of setting) with a test.
- [ ] **Step 6: Tests.** In a `world_test.go` (root package; check whether one exists and how it
  builds a `World`): an NPC actor's `events.Input` reaches `TryCommand` (register a test command);
  input for an unknown actor ID is dropped without panic; an NPC's unknown command logs and sends
  nothing (assert no `events.Message` queued); an NPC's delayed command holds back its next one
  (FIFO); a player's ready command runs while a delayed one waits. If building a `World` in a test is
  impractical, extract the router's decision into a function that takes the tracker maps and test
  that; say which.
- [ ] **Step 7:** `go build ./...`, `go vet ./...`, `go test ./... -count=1`. Commit
  `world: one input path for every actor; mobcommands deleted`.

---

### Task 8: Guards

**Files:** Create `internal/guards/commands_test.go`.

- [ ] Guards (AST, as the other guard files do):
  1. No non-test file imports `internal/mobcommands` (it is deleted; this stops it coming back) and
     none defines a function named `processMobInput`.
  2. No function in `internal/usercommands` has a parameter of type `*accounts.Account` named `user`
     whose function also matches the `Command` shape (rest string, ..., *rooms.Room,
     events.EventFlag): every command takes the actor.
  3. No `.AddMobCommand(` call anywhere.
  Prove each bites with a temporary edit (Edit tool), then remove it and confirm `git diff` is clean.
  Commit `guards: one command set`.

---

### Task 9: Smoke: an NPC runs a command through the shared path

**Files:** `cmd/smoke/admin.go` (and its step list).

- [ ] Read `cmd/smoke/admin.go` and the admin commands for spawning an NPC and making it act
  (`admin.mob.go`, `admin.command.go`; find the exact syntax, e.g. `mob spawn <id|name>` and
  `command <name> <text>`). After the admin spawns the rings, it spawns a `rat` in the start room
  and runs `command rat say smoke-<tag>`; the admin's connection must receive the line
  `rat says, "smoke-<tag>"` (with ANSI stripped, the client already strips it). Then the admin
  ends the rat the way a test admin can (if no command exists, leave it: the start room's own rat
  wanders anyway; say so). Run `go run ./cmd/smoke`: PASS. Commit
  `smoke: an NPC speaks through the shared command path`.

---

### Task 10: Verify and record (controller)

- [ ] `go build ./...`; `go test ./... -count=1` (36 packages less `mobcommands` if it had tests
  (it has none: check) plus `behaviour`: expect 37; the 3 known Windows failures only); `gofmt -l`
  on changed Go files; `go run ./cmd/smoke`; `go run ./cmd/smoke -bots 10` (load, with NPC moves now
  paying AP).
- [ ] Records: `PATCH_NOTES.md` (one command set and input path; NPCs pay AP and use player
  wording; nobody walks out of a fight, players included, and NPCs no longer chase a player who
  walks off, since none can; unknown NPC commands logged; bugs fixed in the fold); `status.md`; spec section 4 "Built
  so far" adds 4 and "Notes for later slices" gets what this slice leaves (ordering per kind for
  slice 6; `tell` and `excludeId` for slice 7; NPC death in `suicide` for slice 5; slice 5's one
  combat round resolves `flee` for NPCs (decision 6); mutator buffs now
  reach NPCs; the dropped commands list); `geomud-verifying` package count; the spec facts that
  Inv A and C corrected (21 mob commands; the `Input` row). Move this plan and the inventory to
  `docs/completed/` after merge.

## Done when

One registry (`usercommands`) serves players and NPCs; `internal/mobcommands` and `processMobInput`
are gone; every command takes the actor; `HandleInputEvents` has one path; NPCs regenerate and pay AP;
nobody in a fight walks out of it; the guards pass; build, the full test suite at the new baseline,
and smoke (including the NPC step) pass.
