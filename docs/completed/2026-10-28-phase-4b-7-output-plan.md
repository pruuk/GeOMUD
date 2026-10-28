# Phase 4b slice 7: output conversion (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every byte of game text reaches a connection through `internal/messaging`; GMCP is the only other writer. Room and world text becomes `messaging.Send` with roles, senses and tokens; actor-only text goes through one interface path. Text parity holds, apart from the changes Calabe ruled.

**Architecture:** Four stages. (1) A per-receiver "reads" golden harness records what every player reads for a set of scenarios, before anything changes. (2) Messaging gains the writers (moved verbatim from `internal/hooks`), the interface, notice and raw paths, an exclude list, literal text and blank lines; `usercommands.Tell` points at it. (3) Call sites convert by group, groups in parallel worktrees, each keeping the goldens unchanged except the ruled differences. (4) The old primitives are deleted and the guard narrows to the GMCP exception.

**Tech stack:** Go, `internal/messaging`, `internal/perception`, `internal/events` queue, testify, `go/ast` guards.

Spec: `2026-10-27-phase-4b-npcs-are-players-design.md` (section 2, section 4 row 7, "Notes for later slices"). Inventory: `2026-10-28-phase-4b-7-inventory.md` (J primitives, K usercommands, L outside). Branch `strip/phase-4b-7-output`.

---

## Facts verified against source (2026-10-28, branch at `62cc22ac`)

| Fact | Value | Where |
|---|---|---|
| Messaging entry point | `messaging.Send(perception.Message)`, synchronous, world goroutine only; Interface or ActorOnly goes to the actor alone; others by `receiversOf`, sense check, role text, tokens | `internal/messaging/messaging.go:30-63` |
| Tokens | `{actor}`, `{target}`, `{direction}` replaced in one pass; an opening "someone" capitalised; unseen reads "someone"; seen name `<ansi fg="username|mobname">Name</ansi>` | `internal/messaging/render.go:11-49` |
| Senses | Sight needs room visibility ≥ 1 or nightvision; Sound passes unless `Faint` without superhearing; hidden subject is unseen | `internal/messaging/perceive.go:13-66` |
| Message fields | `Kind, Actor, Target, Room, Senses, Faint, Range, Audience, Named, ActorText, TargetText, ObserverText, FarText, TargetRoomText, NoTargetRoomText` | `internal/perception/perception.go:69-93` |
| `Message.Room` | used only when there is no actor | `messaging.go:130-134` |
| World range | skips actors with no loaded room | `messaging.go:178-183` |
| Connection delivery | `ConnectionController.Perceive` drops empty text, else `account.SendText(p.Text)` | `internal/actors/controller.go:38-44` |
| `Account.SendText` | queues `events.Message{UserId, Text: txt + "\n"}` | `internal/accounts/userrecord.go:334-341` |
| `Room.SendText`, `SendTextCommunication` | queue `events.Message{RoomId, Text+"\n", ExcludeUserIds}` | `internal/rooms/rooms.go:262-283` |
| `Room.SendTextToExits` | per exit and temp exit, a `RoomId` message (quiet optional) | `rooms.go:285-315` |
| Text writer | `Message_SendMessage`: `AnsiParse`, screen-reader strip, cursor-column + erase-line prefix, `connections.SendTo`, then `RedrawPrompt` at priority 100; RoomId messages resolve players at drain time, skip `UserId` and `ExcludeUserIds`, `IsQuiet` needs superhearing | `internal/hooks/Message_SendMessages.go:15-90` |
| Broadcast writer | every `accounts.GetAllActiveUsers()`, redraw queued, screen-reader text and `SkipLineRefresh` honoured | `internal/hooks/Broadcast_SendToAll.go:13-70` |
| Prompt writer | `RedrawPrompt_SendRedraw`, `OnlyIfChanged` compares `cmdprompt` temp data | `internal/hooks/RedrawPrompt_SendRedraw.go:12-44` |
| Web client writer | `WebClientCommand_SendWebClientCommand`, websocket only | `internal/hooks/WebClientCommand_SendWebClientCommand.go` |
| Event structs | `Message{UserId, ExcludeUserIds, RoomId, Text, IsQuiet, IsCommunication}`; `Broadcast{Text, TextScreenReader, IsCommunication, SourceIsMod, SkipLineRefresh}`; `WebClientCommand{ConnectionId, Text}`; `RedrawPrompt{UserId, OnlyIfChanged}` | `internal/events/eventtypes.go:61-98, 314-319` |
| `Tell` | `usercommands.Tell(actor, text)`: account `SendText` or nothing | `internal/usercommands/usercommands.go:35-41` |
| `ExcludeId` | account `UserId` or 0; also used as a user ID for templates and the mapper | `usercommands.go:43-50`; inventory K §4e |
| Existing connection guard | `TestOnlyMessagingWritesToConnections`: bans `connections.SendTo/SendRawTo/Broadcast` outside `internal/connections`, `internal/messaging`, `internal/perception`; GMCP permanent; 15 files on a shrinking list | `internal/guards/messaging_test.go` |
| Raw `net.Conn` writes | "server full" and "AI pool full" | `main.go:1162, 1164, 1211` |
| `connections.SendTo` callers | 15 files: 6 hooks, 7 inputhandlers, `main.go` (21), `world.go` (3); plus `modules/gmcp/gmcp.go` | grep (inventory L §1) |
| usercommands sends | 804 sites in 68 files: 514 account `SendText`, 67 room `SendText`, 8 `SendTextCommunication`, 2 `SendTextToExits`, 204 `Tell`, 1 `events.Message`, 2 `events.Broadcast`, 6 `messaging.Send` | inventory K §1 |
| Outside usercommands | about 65 `SendText`/`Tell` sites in hooks, accounts, rooms, `world.go`, modules cleanup, follow, gmcp (Mudlet command text, 20 lines) | inventory L §1 |
| Import cycles | `accounts`, `rooms`, `actors`, `connections` cannot import `messaging`; `usercommands`, `hooks`, `inputhandlers`, `world.go`, `main.go`, modules can | inventory J §3, L §0 |
| Tests that capture room messages | `speech_test.go`, `look_test.go`, `items_test.go` capture `RoomId` messages only; two negative checks (`items_test.go:259`, `look_test.go:28`) would pass vacuously after conversion | inventory K §5 |
| Combat golden | records per receiver from `events.Message` (markup, before `AnsiParse`) | `internal/hooks/combat_golden_test.go:1-16, 100-120` |
| Smoke | asserts substrings; transcripts compared by hand; two runs of one commit differ by 33 masked lines | inventory K §5, L §5 |
| Test baseline (Windows) | 39 packages ok, 3 known failures | `geomud-verifying` |

## Decisions

Calabe (2026-10-28):

1. **Converted lines follow the messaging rules everywhere.** A hidden actor reads "Someone" in every line others get (emote, lock, unlock, put, get from a corpse, look, speech); "someone says" becomes "Someone says". The upstream oddities are fixed: the looking player no longer reads "X is looking at Y" about itself; admin teleport's "appears in a flash of light" goes to the arrival room. This extends the slice 5 ruling (a hidden attacker reads "Someone" everywhere) to every action.

Plan defaults (each keeps today's text and order unless stated):

2. **Writes stay queued.** `events.Message`, `events.Broadcast`, `events.WebClientCommand` stay the transport and `RedrawPrompt` stays at priority 100, so line and prompt order is as today. The four writer listeners move into `messaging` verbatim; `hooks.go` registers them. Only `messaging` (and `Account.SendText`'s two callers, below) construct these events.
3. **`Account.SendText` stays as the transport primitive**, called only by `ConnectionController.Perceive` and `internal/messaging`. "Deletes direct `SendText`" means no other caller. `Room.SendText`, `SendTextCommunication` and `SendTextToExits` are deleted.
4. **Interface text** (`messaging.Tell`) always writes, and empty text writes a blank line (80 spacing sites keep their spacing). A behaviour drops it, as today.
5. **Typed text is literal.** `messaging.Literal` escapes braces so `say {target}` prints as typed.
6. **Notices to every active account** (save, reload, shutdown, `broadcast`, the death line) go through `messaging.Notify`, the old `Broadcast` semantics: no sense check, roomless accounts included, screen-reader text and `SkipLineRefresh` kept. NPCs get none.
7. **Raw bytes** (login screens before an account exists, telnet and terminal negotiation, echo, the splash screen, server-full, the log tee) go through `messaging.Raw` and `messaging.RawConn`. Telnet negotiation is not a second exception: messaging owns every socket byte but GMCP's. `internal/connections` keeps its own transport writes (copyover text in `fd_unix.go`), inside the guard's allowed prefix.
8. **Senses keep today's reach.** A converted room line uses `Sight|Sound` (a dark room still gets it, with names as "someone" when unseen, decision 1); speech uses `Sound`; a quiet line uses `Sound` with `Faint`. Phase 5's ruleset decides real senses.
9. **Text to a named player elsewhere** (party notices, admin notices, the tame and alignment lines) is interface text to that player's actor: `messaging.Tell(theirId, ...)`, always delivered as today. `whisper` is a `Named` Sound perception in `World` range (spec section 2); non-faint sound always passes today, so it always arrives.
10. **Packages below messaging** (`rooms`, `accounts`) send through function variables that `messaging.Init()` sets: `rooms.SendMessage func(perception.Message)` and `accounts.Tell func(u *Account, text string)`. `main.go` calls `messaging.Init()` beside `actors.Init()`.
11. **An exclude list:** `Message.Exclude []ids.ID`; excluded bodies receive nothing, actor included. Needed by go's "You hear someone moving around" (origin room excluded).
12. **Parity tool:** per-receiver "reads" goldens at the markup level (Task 1), recorded before any conversion; each group task re-records only its own scenarios and lists every changed line against decision 1. The moved writers get a byte-level test on a fake connection (Task 2). Smoke stays the end-to-end check.
13. **`print`** keeps writing without a newline (`messaging.TellRaw`).
14. **`ExcludeId` goes.** Where it served as a user ID for templates and the mapper, it becomes `UserIdOf(actor)` (same body).
15. **Groups run in parallel worktrees** and are merged by cherry-pick in group order. Each group edits only its files and its own golden files.

## Files

- Create `internal/hooks/reads_golden_test.go`, `internal/hooks/testdata/reads-golden/*.txt` (Task 1).
- Modify `internal/hooks/combat_golden_test.go` (shares the recorder), `internal/usercommands/speech_test.go`, `look_test.go`, `items_test.go` (per-receiver capture).
- Create `internal/messaging/wire.go`, `internal/messaging/interface.go`, `internal/messaging/wire_test.go`, `internal/messaging/interface_test.go` (Task 2).
- Modify `internal/messaging/messaging.go`, `render.go`, `internal/perception/perception.go`, `internal/actors/controller.go`, `internal/hooks/hooks.go`, `internal/usercommands/usercommands.go`, `internal/rooms/rooms.go`, `internal/accounts/userrecord.go`, `main.go`.
- Delete `internal/hooks/Message_SendMessages.go`, `Broadcast_SendToAll.go`, `RedrawPrompt_SendRedraw.go`, `WebClientCommand_SendWebClientCommand.go` (bodies move to messaging).
- Group tasks: the files listed in each.
- Create `internal/guards/output_test.go`; modify `internal/guards/messaging_test.go`.

Every subagent: edit files only with the Edit and Write tools (no `sed -i`, no shell redirection into repo files, no Python read-modify-write). `git add` named paths only; commit messages through a quoted heredoc ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never push, never touch `upstream` or DOGMud.

---

### Task 1: Reads goldens (before any conversion)

**Model:** Opus. **Files:** create `internal/hooks/reads_golden_test.go` and `internal/hooks/testdata/reads-golden/`; modify `internal/hooks/combat_golden_test.go`, `internal/usercommands/speech_test.go`, `look_test.go`, `items_test.go`.

The recorder is the combat golden's: it listens to `events.Message`, `events.Broadcast` and `events.Input`, and expands each event to the players who read it exactly as the writer delivers it (a `UserId` message to that user; a `RoomId` message to the room's players at drain time minus `UserId` and `ExcludeUserIds`, honouring `IsQuiet`; a Broadcast to every active account, screen-reader text for a screen-reader user). Text is recorded as markup (before `AnsiParse`). After conversion every line arrives as a `UserId` message or a Broadcast, so the same recorder keeps working.

- [ ] **Step 1:** Read `internal/hooks/combat_golden_test.go` in full (chunks of at most 400 lines). Move its recorder (the `onMessage` expansion, the Broadcast and Input listeners, name maps, `-update` flag handling) into `reads_golden_test.go` as a reusable `readsRecorder` with: `newReadsWorld(t, seed)` that builds rooms A (`93101`, exits north to B and east to C) and B (`93102`, south to A) and C (`93103`, west to A), zone `readszone`; `addPlayer(name, room)` (account, character, actor via `actors.Spawn`, active user, settings default) and `addNPC(name, room)`; `run(actor, line)` that calls `usercommands.TryCommand(verb, rest, actorId, 0)` then drains `events.ProcessEvents()` until empty; `section(title)`; and `compare(name)` against `testdata/reads-golden/<name>.txt` (`-update` rewrites). Each line of the record is `<reader>: <markup>`; queued inputs are `Input <name>: "<text>"`. Keep `combat_golden_test.go` using the shared recorder; its goldens must not change (`go test ./internal/hooks/ -run Golden -count=1` passes without `-update`).
- [ ] **Step 2: Scenarios.** One golden file per group, named by group letter, each a sequence of sections. Every scenario has three players (`Ann` actor, `Bob` observer in the same room, `Cid` in the next room) and one NPC (`rat`) unless the section says otherwise, so actor, target, observer and far text all show. Include at least:
  - `A-admin-room.txt`: `room info`, `room exits`, one `room set` reply (run as an admin account: set `Role` the way `usercommands_test.go` does).
  - `B-admin-item-mob.txt`: `item spawn` of an item that exists in the empty world, `mob spawn rat`, `server` with no args.
  - `C-admin-other.txt`: `teleport` Ann with Bob in her party to C (shows the "appears" line), `locate Bob`, `modify` (a role change notice to Bob), `mute`, `grant`, `spawn`, `syslogs` with no args.
  - `D-items.txt`: get, drop, put into a container, give to Bob, give to the rat, equip, remove, read, lock and unlock a door exit, show to Bob, `get all`, `drop all`, each once as a visible Ann and once with Ann hidden (give Ann the hidden buff flag the way `combat_golden_test.go`'s hidden fixture does).
  - `E-move-look.txt`: `north` (leave, arrive, "You hear someone moving around" in C), `look`, `look Bob`, `look rat`, `look north` (look through an exit), `look` at an item, `consider rat`, `attack rat` one round of `DoCombat`, `flee`, Ann hidden for `look Bob`.
  - `F-comms-party.txt`: `say hi`, `say {target}` (literal braces), `shout hi` (reaches C as far text), `emote waves`, `emote` hidden, `whisper Bob psst`, `broadcast hello`, `party invite Bob`, `party accept` (as Bob), `party chat hi`, `party leave`.
  - `G-character-info.txt`: `set`, `help`, `who`, `online`, `inventory`, `history`, `alias`, `macros`, `visited`, `keyring`, `save`, `motd`, `biome`, `print hello`.
  - `H-outside.txt`: NPC death through `suicide` on the rat after damage by Ann (tame and alignment lines), a corpse crumbling (`room.UpdateCorpses` with a corpse past its decay), a room idle message (`RoomRoundTick` with a seeded room idle message), alignment decay (`ActorUpkeep` on a round divisible by `AlignmentDecayRounds`), `GrantXP` level text, the follow module's lines (`follow Ann` as Bob, then Ann moves), the cleanup module's lines if it has commands, the player-death line (Bob downed and `suicide`).
  If a command needs data the empty world lacks, pick the nearest that works and note it in the section title; do not skip a group. Seed the RNG per file (`util.SeedRandForTest`).
- [ ] **Step 3:** Record: `go test ./internal/hooks/ -run Reads -count=1 -update`, then run without `-update` twice: identical (no map-order or clock noise; mask a clock time only if it appears, by replacing digits in that one field and saying so in the file header).
- [ ] **Step 4: Per-receiver unit tests.** Rewrite the capture in `speech_test.go`, `look_test.go`, `items_test.go` to record per receiver (use a small helper in each test package that does what the recorder does; usercommands tests cannot import hooks test code). The negative checks at `items_test.go:259` and `look_test.go:28` must assert "nobody read anything". Prove they bite: temporarily add `room.SendText("x")` at the start of the code path each negative check covers, see it fail, remove it.
- [ ] **Step 5:** `go test ./internal/hooks/ ./internal/usercommands/ -count=1` → ok. Commit (`git add` the test files and `internal/hooks/testdata/reads-golden/`), message `tests: per-receiver reads goldens for every slice 7 group, recorded before conversion; usercommands speech/look/items capture per receiver`.

### Task 2: Messaging owns the writers and gains the interface, notice and raw paths

**Model:** Opus. **Files:** as listed under Files for Task 2.

- [ ] **Step 1: Move the writers.** Create `internal/messaging/wire.go` holding, verbatim apart from the package and names, `WriteMessage` (from `Message_SendMessage`), `WriteBroadcast` (from `Broadcast_SendToAll`), `WritePrompt` (from `RedrawPrompt_SendRedraw`) and `WriteWebClientCommand` (from `WebClientCommand_SendWebClientCommand`), each an `events.Listener`. `git rm` the four hook files; in `internal/hooks/hooks.go` register `messaging.WriteMessage` etc. in the same slots. Check messaging may import `templates`, `term`, `util`, `connections`, `events`, `accounts` without a cycle (`go build ./...`).
- [ ] **Step 2: Byte test.** `internal/messaging/wire_test.go`: register a fake `net.Conn` with `connections.Add`, an account bound to it, queue `events.Message{UserId}` and `events.Broadcast{}` (one with screen-reader text and a screen-reader account), drain, and assert the exact bytes: cursor-column + erase-line prefix, `AnsiParse` output, screen-reader stripping, and the prompt after. Write the expected bytes from the current code's behaviour (read it), not from the new code's output.
- [ ] **Step 3: Perception additions** in `internal/perception/perception.go`: `Exclude []ids.ID` ("bodies that receive nothing, the actor included"). In `messaging.Send` skip excluded receivers (and an excluded actor in the Interface/ActorOnly branch). Test: an excluded observer and an excluded far receiver get nothing.
- [ ] **Step 4: Literal text** in `render.go`: `func Literal(s string) string` returns `s` with every `{` doubled; `resolveTokens` adds the pair `{{` → `{` first in its replacer, so escaped braces print as typed and are never read as tokens. Test: `"{actor} says, \"" + Literal("{target} {{x}") + "\""` resolves to the name then `"{target} {{x}"`.
- [ ] **Step 5: Interface paths** in `internal/messaging/interface.go`:

```go
// Tell writes interface text to one actor: no sense check, the actor's
// account only (a behaviour drops it). Empty text writes a blank line.
func Tell(actorId ids.ID, text string) {
	Send(perception.Message{Kind: perception.Interface, Actor: actorId, Audience: perception.ActorOnly, ActorText: text})
}

// TellRaw is Tell without the closing newline (the print command).
func TellRaw(actorId ids.ID, text string)

// ToAccount writes interface text to an account whether or not it has an
// actor yet (login, character creation, logout).
func ToAccount(u *accounts.Account, text string)

// Notice is text for every active account, with no sense check: saves,
// reloads, shutdown, the broadcast command and death notices.
type Notice struct {
	Text             string
	ScreenReaderText string // optional; a screen-reader account reads it instead
	SkipLineRefresh  bool
}

// Notify sends n to every active account, roomless ones included. NPCs get none.
func Notify(n Notice)

// Raw writes bytes to a connection as they are: screens before an account
// exists, telnet and terminal negotiation, echo.
func Raw(connId connections.ConnectionId, b []byte)

// RawConn writes to a connection that is not registered yet (server full).
func RawConn(c net.Conn, b []byte)

// WebClient sends a web client command (password masking).
func WebClient(connId connections.ConnectionId, text string)

// UserIdOf is the account UserId behind an actor, or 0.
```

Interface text is never token-resolved: in `deliver`, a `perception.Interface` message keeps its text as written, so admin help that shows `{{ .X }}` or a literal `{actor}` is untouched (Interface text is for the actor alone; callers write names themselves). `Literal` and the `{{` pair matter only for world events. Test: `Tell(id, "{actor} {{x}}")` writes exactly that. `TellRaw` and `ToAccount` queue `events.Message{UserId}` (TellRaw without `"\n"`; ToAccount with). Blank lines: `ConnectionController.Perceive` writes when `p.Text != ""` **or** `p.Message.Kind == perception.Interface`. `Notify` queues `events.Broadcast`. `Raw` calls `connections.SendTo`; `RawConn` calls `c.Write`. Use whatever `ConnectionId` type `connections.SendTo` takes. Tests for each: `Tell` to a player writes the line, to an NPC writes nothing, empty writes `"\n"`; `TellRaw` has no newline; `ToAccount` works with no actor; `Notify` reaches an account with no actor and uses screen-reader text.

- [ ] **Step 6: Re-point.** `usercommands.Tell(actor, text)` becomes `messaging.Tell(actor.Id, text)` (keep the function; groups may call either). Add `func UserIdOf(actor *actors.Actor) int` in usercommands (same body as `ExcludeId`) for template and mapper IDs. Set `rooms.SendMessage` and `accounts.Tell` in `messaging.Init()` (declare `var SendMessage func(perception.Message)` in rooms and `var Tell func(u *Account, text string)` in accounts, each with a doc comment saying messaging sets it at startup and why: the import cycle). `main.go` calls `messaging.Init()` right after `actors.Init()`. Tests that need them call `messaging.Init()` in `TestMain`.
- [ ] **Step 7:** `go build ./... && go vet ./... && go test ./internal/... -count=1`: baseline, and the reads goldens unchanged (no `-update`). Commit: `messaging: owns the text, broadcast, prompt and web client writers; Tell, ToAccount, Notify, Raw, Literal, Exclude; blank interface lines; usercommands.Tell points at messaging`.

### Conversion rules (Tasks 3-11)

Each group converts every output call in its files. Read each file in full (chunks of at most 400 lines) before editing it; report "converted N of N sites" per file with the grep before and after.

| Today | Becomes |
|---|---|
| `Tell(actor, t)` | stays (it is messaging now) |
| `user.SendText(t)` where `user` is the actor's account | `Tell(actor, t)` |
| `user.SendText(t)` for another player | `messaging.Tell(theirActorId, t)`; with only an account and maybe no actor, `messaging.ToAccount(u, t)` |
| an admin helper that takes `*accounts.Account` | `messaging.ToAccount(user, t)` (A and B use this; do not thread the actor through) |
| `Tell(actor, "You X")` + `room.SendText("<name> Xs", ExcludeId(actor))` | one `messaging.Send(perception.Message{Kind: perception.WorldEvent, Actor: actor.Id, Senses: perception.Sight \| perception.Sound, Range: perception.Room, ActorText: "You X", ObserverText: "{actor} Xs"})` |
| with a target excluded too | add `Target`, `TargetText` |
| a name written as `<ansi fg="%s">%s</ansi>` with `NameColour(actor)` or `"username"` | `{actor}` / `{target}` (same markup for a seen body; "someone" when unseen; `mobname` for an NPC) |
| a room line with no exclude (the actor reads it too) | `Send` with `Actor`, `ActorText` equal to the observer line written for the actor, or, if the line is not about an actor, an actorless world event: `Actor` zero, `Room: room.RoomId` |
| `room.SendTextCommunication` (say, emote) | `Send` with `Senses: perception.Sound` (say) or `Sight\|Sound` (emote); typed text wrapped in `messaging.Literal(...)` |
| a hand-written "someone says" branch for a hidden speaker | delete the branch; `{actor}` does it (decision 1) |
| `otherRoom.SendTextCommunication` per exit (shout) | `Range: perception.Nearby, FarText: "...{direction}..."` |
| `room.SendTextToExits(t, quiet, ex...)` | a second `Send` with `Range: perception.Nearby`, `FarText: t`, empty `ActorText`/`ObserverText`, `Faint: quiet`, `Senses: perception.Sound`, `Exclude` for the excluded bodies |
| a line to a player in another room (party, notices) | `messaging.Tell(theirId, t)` (decision 9) |
| `whisper` | `Send{Audience: perception.Named, Named: []ids.ID{target}, Range: perception.World, Senses: perception.Sound, Target: target, ...}` |
| `events.Broadcast{...}` | `messaging.Notify(messaging.Notice{...})` (same fields) |
| `events.Message{...}` built by hand | the matching messaging call |
| `connections.SendTo(b, id)` | `messaging.Raw(id, b)` |
| `ExcludeId(actor)` used as a user ID (templates, mapper) | `UserIdOf(actor)` |
| `user.SendText("")` / `Tell(actor, "")` | `Tell(actor, "")` (writes a blank line) |

Templates, wrapping and `ansi` tags stay as they are (spec: only the sending moves). A send that must happen before a move (go's "leaves" line) moves before `MoveToRoom`. Each group runs `go test ./internal/hooks/ -run Reads -count=1` after converting; if its own golden changes, it re-records with `-update` and reports every changed line, each tied to decision 1 (hidden or unseen "Someone", the two fixed oddities, `mobname` for an NPC named in `username` today) or to a listed exception; any other change is a bug to fix, not to record. No group touches another group's golden.

### Task 3-9: usercommands groups (parallel, Sonnet, one worktree each)

| Task | Group | Files (`internal/usercommands/`) | Golden |
|---|---|---|---|
| 3 | A admin room and zone | `admin.room.go`, `admin.zone.go` | `A-admin-room.txt` |
| 4 | B admin item, mob, server, build | `admin.item.go`, `admin.mob.go`, `admin.server.go`, `admin.build.go` | `B-admin-item-mob.txt` |
| 5 | C admin other | `admin.badcommands.go`, `admin.command.go`, `admin.copyover.go`, `admin.grant.go`, `admin.locate.go`, `admin.modify.go`, `admin.mute.go`, `admin.prepare.go`, `admin.redescribe.go`, `admin.reload.go`, `admin.rename.go`, `admin.spawn.go`, `admin.syslogs.go`, `admin.teleport.go`, `admin.visit.go` | `C-admin-other.txt` |
| 6 | D items | `get.go`, `drop.go`, `put.go`, `give.go`, `equip.go`, `remove.go`, `read.go`, `lock.go`, `unlock.go`, `show.go` | `D-items.txt` |
| 7 | E movement, look, combat | `go.go`, `look.go`, `look.panels.go`, `descriptions.panels.go`, `default.go`, `flee.go`, `attack.go`, `shoot.go`, `consider.go`, `suicide.go`, `suicide_npc.go` | `E-move-look.txt` |
| 8 | F communication and party | `say.go`, `shout.go`, `emote.go`, `whisper.go`, `broadcast.go`, `party.go` | `F-comms-party.txt` |
| 9 | G character and info | `set.go`, `start.go`, `alias.go`, `macros.go`, `password.go`, `history.go`, `who.go`, `save.go`, `quit.go`, `motd.go`, `online.go`, `help.go`, `visited.go`, `keyring.go`, `inventory.go`, `inventory.panels.go`, `biome.go`, `print.go`, `noop.go`, `usercommands.go` (its sends other than `Tell` itself, e.g. the plaintext-password nag) | `G-character-info.txt` |

Before dispatch, the controller checks with `ls internal/usercommands/*.go` that every non-test file is in exactly one group or is `usercommands.go`/`prompt.go`-style infrastructure with no sends; a file with sends in no group is added to the nearest group.

Each task's steps:
- [ ] **Step 1:** Read the conversion rules, decisions 1-15, inventory K's rows for your files (§1, §3, §4 entries naming them) and your golden file.
- [ ] **Step 2:** Convert file by file; `go build ./...` after each file.
- [ ] **Step 3:** `go vet ./internal/usercommands/ && go test ./internal/usercommands/ ./internal/hooks/ -count=1`. Re-record your golden only if it changed, with the changed-line report.
- [ ] **Step 4:** `grep -n 'SendText\|SendTextCommunication\|SendTextToExits\|events.Message{\|events.Broadcast{\|ExcludeId(' <your files>` prints nothing (a known-hit check first: run it on `master`'s copy of one of your files with `git show`).
- [ ] **Step 5:** One commit per group: `usercommands: <group> output through messaging`.

Group-specific notes:
- **C:** `admin.teleport.go:133`'s "appears in a flash of light" goes to the arrival room (decision 1); `:102` tells the admin about the player it moved, not "Moved to room" about itself, only if that is the same fix; otherwise keep the text.
- **D:** `give.go:171` names a mob in `username`: `{target}` gives `mobname`. `show.go:63` sends the long description to the target: it goes in `TargetText`.
- **E:** `go.go:198` "leaves" moves before `MoveToRoom`; `go.go:211` uses `Exclude` with the origin room's bodies; `look.go:91` stops sending the looker's own line (decision 1); `look.go:113` keeps the baked mob name (`GetMobName`) in the text rather than `{target}`, so adjectives stay; `look.go:429,434` (a line into the room being looked into) is an actorless world event with `Room` set to that room and the looker's name written as `{actor}` is not possible there, so keep the baked name and say so. `suicide.go:96`'s death line becomes `messaging.Notify` (drop the trailing CRLF only if the golden shows a doubled blank line, and say which).
- **F:** say/shout typed text goes through `messaging.Literal`. `party.go` lines to members are `messaging.Tell(memberId, ...)`.
- **G:** `print.go` uses `messaging.TellRaw`. `start.go` runs before the character is placed: use `ToAccount` where there is no actor.

### Task 10: Outside usercommands (hooks, rooms, accounts, world, modules)

**Model:** Sonnet. **Files:** every non-test file outside `internal/usercommands` and `internal/messaging` that inventory L §1 lists for `internal/hooks`, `internal/accounts`, `internal/rooms`, `internal/actors` (only `controller.go` may keep `SendText`), `world.go`, `modules/cleanup`, `modules/follow`, `modules/gmcp/gmcp.Mudlet.go` (its command text, not GMCP packets). Golden `H-outside.txt`.

- [ ] Steps as for Tasks 3-9. Rooms and accounts use `rooms.SendMessage(...)` and `accounts.Tell(u, ...)` (decision 10). `world.go`'s prompt writes (`:563`, `:592`) become `messaging.Raw` with the same bytes. Hooks that queue `events.Message` or `events.Broadcast` use the messaging calls. Commit: `hooks, rooms, accounts, world, modules: output through messaging`.

### Task 11: Raw paths (login, protocol, main)

**Model:** Sonnet. **Files:** `internal/inputhandlers/*.go`, `main.go`, `internal/hooks/Log_FollowLogs.go`, `internal/hooks/PlayerDespawn_HandleLeave.go`.

- [ ] Every `connections.SendTo(b, id)` becomes `messaging.Raw(id, b)`; the three `conn.Write` calls in `main.go` become `messaging.RawConn(conn, b)`; `SendWebClientCommand` callers use `messaging.WebClient`. Bytes unchanged. Smoke `-bots 1` passes (login, password prompt and its masking, start). Commit: `inputhandlers, main, log tee: raw writes through messaging`.

### Task 12: Delete the old primitives; the guard

**Model:** Sonnet, then the controller checks.

- [ ] **Step 1:** Delete `Room.SendText`, `Room.SendTextCommunication`, `Room.SendTextToExits`, `usercommands.ExcludeId`, `Account.SendWebClientCommand` (no callers) and `connections.SendRawTo` if it has no callers. `go build ./...` must pass; a caller left is a group's miss: report it, convert it, and say which group.
- [ ] **Step 2: Guard** `internal/guards/output_test.go`, `TestOutputGoesThroughMessaging`: outside `internal/messaging`, `internal/connections` and `internal/perception`, no call to a method named `SendText` except in `internal/actors/controller.go`; no composite literal of `events.Message`, `events.Broadcast` or `events.WebClientCommand`; no `connections.SendTo/SendRawTo/Broadcast` except `modules/gmcp/gmcp.go`; no `.Write(` on a `net.Conn` (match `conn.Write` and any selector `Write` whose receiver's type, by `go/types` if simple, else by the identifier names used in `main.go`, is a connection). Positive controls: `messaging.WriteMessage` exists in `internal/messaging/wire.go`, and `ConnectionController.Perceive` calls `SendText`. Narrow `TestOnlyMessagingWritesToConnections`'s `connectionWriters` to empty and delete the map (the test keeps the GMCP permanent entry). Prove the guard bites: add `room.SendText` is impossible now, so add `events.Message{}` in `internal/hooks/hooks.go`, see it fail, remove it.
- [ ] **Step 3:** The Phase 4b Done-when line in `gomud_strip_plan.md`: read it and check each clause with evidence (one command set, one combat path, smoke with an NPC fight); report.
- [ ] **Step 4:** Commit: `Delete Room.SendText and the other direct writers; guard: output goes through messaging, GMCP the only exception`.

### Task 13: Verify, docs, review, PR

- [ ] `geomud-verifying`: build, vet, full tests (39 ok, 3 known), all reads goldens and combat goldens pass without `-update`; smoke `-bots 1`, `-bots 2`, `-bots 10`.
- [ ] Opus whole-branch review against this plan, decision 1's list of text changes, and the goldens' diffs since Task 1 (`git diff <task1-sha>..HEAD -- internal/hooks/testdata/reads-golden/`): every changed golden line must be justified by a decision. Fix confirmed findings.
- [ ] Docs: `PATCH_NOTES.md` slice 7 line (list every player-visible text change from the golden diffs); the spec's notes (remove the slice 7 items they settled, add what slice 7 leaves: per-receiver render hook still none, `look.go` baked names, telnet negotiation now under messaging); `status.md` (Phase 4b done if the Done-when check holds; next Phase 5); the `geomud-verifying` baseline if the package count changed.
- [ ] Push, PR (`gh pr create --repo pruuk/GeOMUD --base master`), merge when CI is green. Leave this plan and the inventory at the repo root (the next phase's first commit moves them).
