# Phase 4b slice 7 inventory (2026-10-28, branch at 62cc22ac)

Three read-only inventories written for the slice 7 plan (`2026-10-28-phase-4b-7-output-plan.md`): J, messaging and the send primitives; K, the output sites in `internal/usercommands`; L, the output sites outside it, the guard and the parity tools. Each lists the files it read. Their open questions were settled by Calabe or by the plan's defaults (plan, "Decisions"). This file moves to `docs/completed/` with the plan.

# Inventory J: delivery layer and send primitives (Phase 4b slice 7)

Branch `strip/phase-4b-7-output`, read 2026-10-28. Read-only: no repo file changed. Every fact below
was read from source in this session; file:line is given for each.

Read 45 of 45 files (listed below). Full = every line printed (chunks of at most 400 lines).
Partial = only the range given; nothing outside it is claimed.

| # | File | Lines read |
|---|---|---|
| 1 | `2026-10-27-phase-4b-npcs-are-players-design.md` | 1-361 (full) |
| 2 | `internal/messaging/messaging.go` | 1-225 (full) |
| 3 | `internal/messaging/perceive.go` | 1-66 (full) |
| 4 | `internal/messaging/render.go` | 1-49 (full) |
| 5 | `internal/messaging/messaging_test.go` | 1-538 (full) |
| 6 | `internal/messaging/main_test.go` | 1-13 (full) |
| 7 | `internal/messaging/render_test.go` | 1-18 (full) |
| 8 | `internal/perception/perception.go` | 1-114 (full) |
| 9 | `internal/actors/controller.go` | 1-88 (full) |
| 10 | `internal/actors/actor.go` | 1-28 (full) |
| 11 | `internal/actors/registry.go` | 1-247 (full) |
| 12 | `internal/actors/wire.go` | 1-58 (full) |
| 13 | `internal/accounts/userrecord.go` | 1-620 (full, two chunks) |
| 14 | `internal/accounts/session.go` | 1-32 (full) |
| 15 | `internal/accounts/userrecord.prompt.go` | 1-313 (full) |
| 16 | `internal/accounts/users.go` | 126-185 (partial); `LoginUser` and `CreateUser` grepped for `activate` only |
| 17 | `internal/hooks/Message_SendMessages.go` | 1-90 (full) |
| 18 | `internal/hooks/Broadcast_SendToAll.go` | 1-70 (full) |
| 19 | `internal/hooks/RedrawPrompt_SendRedraw.go` | 1-44 (full) |
| 20 | `internal/hooks/WebClientCommand_SendWebClientCommand.go` | 1-25 (full) |
| 21 | `internal/hooks/hooks.go` | 1-72 (full) |
| 22 | `internal/hooks/Log_FollowLogs.go`, `PlayerDespawn_HandleLeave.go` | grep excerpts around the SendTo lines only (53-59, 60-66) |
| 23 | `internal/events/events.go` | 85-244 (partial) |
| 24 | `internal/events/eventtypes.go` | 51-107, 314-330 (partial) |
| 25 | `internal/connections/connections.go` | 1-352 (full) |
| 26 | `internal/connections/connectiondetails.go` | 1-457 (full, two chunks) |
| 27 | `internal/connections/heartbeat.go`, `copyover.go`, `clientsettings.go`, `memory.go`, `fd_windows.go` | full |
| 28 | `internal/connections/fd_unix.go` | 1-120 (full) |
| 29 | `internal/inputhandlers/login.go` | 1-288 (full) |
| 30 | `internal/inputhandlers/login_prompt_handler.go` | 1-435 (full, two chunks) |
| 31 | `internal/inputhandlers/echo.go`, `cleanser.go`, `term_iac.go`, `term_ansi.go`, `systemcommands.go`, `signals.go`, `term_textprefix.go`, `inputhistory.go` | full |
| 32 | `internal/guards/doc.go`, `messaging_test.go`, `guards_test.go` | full |
| 33 | `internal/guards/combat_test.go` | 55-103 (partial); `commands_test.go`, `tick_test.go` grepped for output names only |
| 34 | `main.go` | 360-1491 (partial; 1-359 grepped for output calls only) |
| 35 | `world.go` | 150-663 (partial; 1-149 grepped for output calls only) |
| 36 | `internal/rooms/rooms.go` | 255-324, 1240-1279 (partial) + import block |
| 37 | `internal/rooms/roommanager.go` | 185-220 (partial) |
| 38 | `internal/usercommands/usercommands.go` | 1-80 (partial) |
| 39 | `internal/usercommands/print.go`, `broadcast.go` | full |
| 40 | `internal/usercommands/suicide.go` | 80-115 (partial) |
| 41 | `modules/gmcp/gmcp.go` | 146-239, 520-608 (partial) |
| 42 | `internal/web/web.go` | 467-511 (partial) |
| 43 | `internal/templates/templates.go` | 140-175, 434-455 (partial) |
| 44 | `internal/characters/character.go` | 540-640 (partial) |
| 45 | `internal/term/term.go` 1-60, `internal/util/hook.go` 42-49, `internal/plugins/plugincallbacks.go` 25-33 | partial |

Skipped on purpose: the bodies of usercommands (inventory K) and of non-usercommands call sites
(inventory L); test files other than messaging's; `cmd/smoke`.

## Facts verified against source

| Fact | Value | Where / how |
|---|---|---|
| `.SendText(` calls, non-test | 631 in 63 files | grep |
| ... on `*accounts.Account` | 554, plus 1 on a value `Account` (usercommands) = 555 | type-resolved count (go/packages program in scratchpad, `sendcount/`) |
| ... by package | usercommands 514, gmcp 20, hooks 7, follow 7, rooms 3, accounts 2, actors 1, root (`world.go`) 1 | same |
| ... on `*rooms.Room` | 76 in 27 files: usercommands 67, hooks 4, rooms 3, cleanup 2 | same; 555 + 76 = 631 reconciles with grep |
| `SendTextCommunication(` | 8 calls in 3 files (emote, say, shout) | type-resolved and grep agree |
| `SendTextToExits(` | 2 calls (`go.go:211`, `say.go:36`) | grep |
| `connections.SendTo(` | 86 in 16 files: main.go 20, login.go 21, login_prompt_handler.go 12, gmcp.go 7, term_ansi.go 6, Broadcast_SendToAll 4, world.go 3, cleanser 2, echo 2, term_iac 2, Message_SendMessages 2, and 1 each in Log_FollowLogs, PlayerDespawn_HandleLeave, RedrawPrompt_SendRedraw, WebClientCommand_SendWebClientCommand, systemcommands | `grep -c`, sums to 86 |
| `connections.SendRawTo` | defined (`connections.go:237`), 0 callers | grep |
| `connections.Broadcast(` | 1 caller, `main.go:376` "Copyover complete." | grep |
| `usercommands.Tell(` / `Tell(` | 226 call lines in 33 files | grep (excludes the definition) |
| `ExcludeId(` | 78 calls in 22 files | grep |
| `events.Broadcast{...}` constructions | 12 in 7 files (AutoSave 6, world.go 2, systemcommands, broadcast.go, suicide.go, main.go). The brief's 13 includes the listener registration `events.Broadcast{}` at `hooks.go:54` | grep |
| `events.Message{...}` constructions | 5 in 3 files: `userrecord.go:336`, `rooms.go:264,276,307`, `print.go:17`. The brief's 6 includes the registration at `hooks.go:44` | grep |
| `.Write([]byte` | 12 in 7 files; only 6 write to a client: `main.go:1162,1164,1211` (raw `net.Conn`), `fd_unix.go:32,35,38` (inside connections). The rest are hashers (`accounts/index.go:198`, `userrecord.go:496`, `procedural/gridmaze.go:92`, `util/util.go:409,416`) and a comment (`term/term.go:27`) | grep, each line read |
| messaging call sites today | hooks/NewRound_DoCombat.go 5, usercommands attack.go 2, shoot.go 1, suicide.go 1, suicide_npc.go 2 | grep `messaging\.` |
| messaging imports `connections`/`templates` | no (0 matches in messaging.go, perceive.go, render.go; the same grep finds them elsewhere) | grep -c |
| Handlers on `actors.OnBehaviourPerceive` | none registered (only the declaration and Fire in `actors/controller.go:70,78`) | grep |
| `events.Message.IsCommunication`, `events.Broadcast.IsCommunication`, `SourceIsMod` | set but never read | grep for readers |
| `events.Broadcast.TextScreenReader` | read by `Broadcast_SendToAll.go:27` but never set by any sender | grep |
| `Account.SendWebClientCommand` | defined `userrecord.go:343`, 0 callers | grep |

## 1. `internal/messaging` today

Files: `messaging.go` (delivery), `perceive.go` (sense checks), `render.go` (names and tokens). The
vocabulary is in `internal/perception/perception.go`, which imports only `ids` so that
`actors.Controller` can name a `Perception` (`perception.go:1-5`).

**API.**
- `messaging.Send(m perception.Message)` is the only exported function besides the `CanPerceive` hook
  variable (`messaging.go:30`, `perceive.go:14`). It "must run on the world goroutine: room maps are
  not locked" and is synchronous and re-entrant (`messaging.go:27-29`).
- `perception.Message` (`perception.go:66-91`): `Kind` (`WorldEvent` | `Interface`), `Actor`,
  `Target` (both `ids.ID`), `Room` (int, source room of an actor-less world event), `Senses`
  (bitset Sight, Sound, Touch, Smell), `Faint`, `Range` (`Room`, `Nearby`, `Zone`, `World`),
  `Audience` (`InRange`, `ActorOnly`, `Named`), `Named []ids.ID`, and the texts `ActorText`,
  `TargetText`, `ObserverText`, `FarText`, `TargetRoomText`, plus `NoTargetRoomText bool`.
- `perception.Perception` (`perception.go:105-114`): `Role`, `Message`, `SenseUsed`, `ActorSeen`,
  `TargetSeen`, `Far`, `Direction`, `Text` (resolved for this receiver).

**Delivery** (`messaging.go:30-83`):
1. `actor := actors.Get(m.Actor)`, `target := actors.Get(m.Target)`.
2. A `WorldEvent` that is not `ActorOnly` and has `Senses == 0` logs an error (`:34-36`) but still
   delivers (test `TestWorldEventWithNoSensesStillReachesTheActor`).
3. `Kind == Interface` or `Audience == ActorOnly`: deliver to the actor only, role Actor, no sense
   check; if the actor is not in the registry nothing is delivered and nothing is logged (`:38-43`).
4. Otherwise `receiversOf` (`:142-202`) lists bodies: every body in the source rooms (actor's room,
   or `m.Room` when there is no actor, plus the target's room when different), then the actor even if
   it is in no room, then by range: `Nearby` adds bodies in neighbour rooms that have an exit back
   (`FindExitTo`), marked far with that exit as `direction`; `Zone`/`World` walk `actors.All()` and
   add every actor in a loaded room outside the source rooms (Zone filters by zone), marked far.
   Actors in no loaded room are skipped (`:191-194`).
5. Per receiver: the actor gets `ActorText` with no sense check; `Named` drops receivers not named;
   `NoTargetRoomText` silences observers in the target's other room; `perceivedBy` tries senses in
   order Sight, Sound, Touch, Smell (`perception.go:29`) through `CanPerceive`; zero means dropped.
6. `deliver` (`:68-83`) builds the `Perception`, picks the text by role (`textFor`, `:86-98`),
   resolves tokens, then calls `rc.actor.Controller().Perceive(p)`.

**`CanPerceive`** (`perceive.go:14-23`): Sight needs `Room.GetVisibility() >= 1` or the
`nightvision` flag; Sound passes unless `Faint` without `superhearing`; Touch and Smell pass.
`canSee` (`:44-55`): self always; a `hidden` subject or a far receiver sees no one; otherwise light.

**Tokens** (`render.go`): `{actor}`, `{target}`, `{direction}`. `nameFor` renders
`<ansi fg="username">Name</ansi>` when the subject's controller has an account, `mobname` otherwise,
or `someone` when unseen (`render.go:12-24`). An opening `someone` is capitalised; replacement is
one pass with `strings.NewReplacer`, so a name holding a token is not substituted again
(`render.go:28-49`). **Every** text passed to `Send` is token-substituted; there is no raw mode. A
user-typed `{target}` with no target becomes the empty string (`nameFor(nil, …)` returns ``,
`render.go:13-15`).

**How delivery reaches the socket.** `ConnectionController.Perceive` (`actors/controller.go:40-45`)
returns on empty `Text` or nil account, else calls `c.account.SendText(p.Text)`. So every messaging
delivery ends in the same queued path as a direct `Account.SendText` (section 2a). The Kind is
ignored by the connection controller. `BehaviourController.Perceive` (`controller.go:74-79`) drops
`Interface` and fires `OnBehaviourPerceive`, which has no handlers today.

**Interface text.** `Kind: Interface` with `Actor` set reaches only that actor (test
`TestActorOnlyAndInterface`, `messaging_test.go:267-278`), needs no light, and is dropped by an NPC
(`TestBehaviourControllerEndToEnd`, `:478-479`). It needs a registered actor: text for an account
with no actor (login, logout, a record not bound) is silently dropped.

**Before an actor exists.** Nothing. messaging has no connection-level function, no function that
takes a connection ID or a `net.Conn`, and does not import `connections`, `templates` or `events`.
The spec asks for one (section 2, "Before an actor exists"); slice 3's note says slice 7 adds it.

**Other gaps** (read from the code, each needed by some primitive in section 2):
- No exclusion list (`Room.SendText(txt, excludeUserIds...)` has no equivalent; `Named` is the
  inverse).
- No "every account" audience without a sense check: `World` range applies senses and reaches NPCs;
  `Interface` reaches only the actor. System notices (save, shutdown, reload, goodbye) have no form.
- No raw-text flag (section 1, Tokens).
- No screen-reader alternative text and no "skip line refresh".
- No prompt handling.

## 2. Send primitives outside messaging

### 2a. `Account.SendText` (`internal/accounts/userrecord.go:334-341`)

Chain:
1. `events.AddToQueue(events.Message{UserId: u.UserId, Text: txt + "\n"})`, default priority 0
   (`events.go:90-123`). The queue is a heap ordered by priority, then FIFO by an order counter
   (`events.go:44-69`).
2. `World.MainWorker` drains it every millisecond under `util.LockMud()` via `EventLoop` →
   `events.ProcessEvents()` (`world.go:368-374, 658-663`). `ProcessEvents` pops until empty, so
   events queued by handlers run in the same drain (`events.go:176-233`).
3. `hooks.Message_SendMessage` (registered `hooks.go:44`; body `Message_SendMessages.go:15-90`): for
   `UserId > 0`, `accounts.GetByUserId`, `templates.AnsiParse(text)`, `util.StripCharsForScreenReaders`
   when `user.ScreenReader`, then `connections.SendTo([]byte(AnsiMoveCursorColumn + AnsiEraseLine +
   text), user.ConnectionId())` (`:27-31`), then `AddToQueue(events.RedrawPrompt{UserId}, 100)`.
4. `connections.SendTo` (`connections.go:259-292`): under the package lock, `cd.Write(b)` per ID; a
   failed write removes the connection after unlock.
5. `ConnectionDetails.Write` (`connectiondetails.go:263-307`): no-op if `outputSuppressed`; `\n` →
   `\r\n`; ANSI SGR stripped when `stripAnsi` (AI port, `main.go:1173-1175`) unless the first byte is
   IAC; then SSH channel write, or websocket `WriteMessage(TextMessage)` under `wsLock` (IAC-first
   payloads refused with an error log, `:294-297`), or `net.Conn.Write`.

Ordering: queued. Text is written in `AddToQueue` order among priority-0 events, after whatever is
already queued at priority 0. The prompt redraw is priority 100 and unique per user
(`eventtypes.go:320`), so one redraw follows the whole drain. Each text write first erases the
current line (the prompt and the player's half-typed input), which the redraw restores with
`GetUnsentText()` (`userrecord.prompt.go:112`).

Callers: 555 calls (table above). `Account.SendWebClientCommand` (`userrecord.go:343-350`) queues
`events.WebClientCommand`; no callers.

### 2b. `Room.SendText`, `SendTextCommunication`, `SendTextToExits` (`internal/rooms/rooms.go:262-317`)

- `SendText(txt, excludeUserIds...)` queues `events.Message{RoomId, Text: txt+"\n", ExcludeUserIds}`
  (`:274-283`). `SendTextCommunication` is the same with `IsCommunication: true` (`:262-272`); that
  field is never read, so the two are identical on the wire.
- `SendTextToExits(txt, isQuiet, exclude...)` (`:285-317`): for every exit and temporary exit of
  `r`, loads the target room and, for each of its exits that leads back to `r`, queues an
  `events.Message` for that room with text `(From <ansi fg="exit">EXIT</ansi>) ` + txt and
  `IsQuiet`.
- `Message_SendMessage` room branch (`Message_SendMessages.go:38-87`): loads the room **at drain
  time**, loops `room.GetPlayers()` (accounts only, so NPCs never get room text), skips
  `message.UserId` and the exclude list, skips quiet messages for players without `superhearing`,
  then the same AnsiParse, screen-reader strip, erase-line write and `RedrawPrompt` as 2a.

Ordering: queued, FIFO with 2a. Receivers are resolved when the event is processed, not when it is
sent. `messaging.Send` resolves receivers synchronously at the call.

Callers: `SendText` 76 (usercommands 67, hooks 4, rooms 3, cleanup 2), `SendTextCommunication` 8,
`SendTextToExits` 2. Inside `rooms` itself: `rooms.go:248,251,534` (`r.SendText`) and
`roommanager.go:197,201,216` (`u.SendText` on accounts, in room maintenance).

### 2c. `connections` writers (`internal/connections/connections.go`)

| Writer | Line | Behaviour | Callers |
|---|---|---|---|
| `SendTo(b, ids...)` | 259 | synchronous `cd.Write` under the package lock | 86 (table above) |
| `SendRawTo(b, ids...)` | 237 | `cd.WriteRaw`: no CRLF, no strip, refused on websocket | none |
| `Broadcast(b, skip...)` | 172 | writes every connection not in `Login` state | `main.go:376` |
| `cd.Write` direct | `fd_unix.go:32,35,38` | copyover save writes `RELOGTKN:<token>` or "Server is rebooting" to websockets (Unix only; `fd_windows.go` writes nothing) | inside connections |
| `Kick`, `Remove` | 127, 150 | close, no text | many |

Ordering: synchronous on the calling goroutine. These bypass the event queue, so a `SendTo` can reach
the client before text queued earlier by `SendText`.

### 2d. `usercommands.Tell` and `ExcludeId` (`internal/usercommands/usercommands.go:35-50`)

- `Tell(actor, text)`: `actor.Controller().Account()`; if non-nil, `user.SendText(text)`; NPCs get
  nothing. Chain and ordering as 2a. 226 call lines in 33 files (30 usercommands files, `hooks/NewRound_DoCombat.go`
  6, `modules/follow` 11, `modules/cleanup` 5).
- `ExcludeId(actor)`: the account's `UserId`, or 0 for an NPC; used only to build
  `Room.SendText` exclude lists. 78 calls in 22 files.
- The comment at `:36` already says "slice 7 points this at messaging".

### 2e. `events.Message` and `Message_SendMessage`

Covered in 2a and 2b. Direct constructions outside the two primitives: `print.go:17` (the `print`
command queues `events.Message{UserId, Text: rest}` with **no** trailing newline, unlike
`SendText`).

### 2f. `events.Broadcast` and `Broadcast_SendToAll` (`internal/hooks/Broadcast_SendToAll.go:13-70`)

Fields `Text`, `TextScreenReader`, `IsCommunication`, `SourceIsMod`, `SkipLineRefresh`
(`eventtypes.go:61-67`). The handler loops `accounts.GetAllActiveUsers()` (non-link-dead accounts,
`users.go:126-138`), queues `RedrawPrompt` (priority 100) first, then writes the screen-reader text
if set and the user has `ScreenReader`, else `Text`, with or without the erase-line prefix by
`SkipLineRefresh`. No newline is added (senders add `term.CRLFStr` themselves). No ScreenReader
character strip here, unlike 2a.

Senders: `NewTurn_AutoSave.go:39,43,51,55,63,67` (the "Saving users... Done." pairs use
`SkipLineRefresh`), `world.go:171,177` (reload), `systemcommands.go:172` (shutdown countdown, from
a goroutine), `usercommands/broadcast.go:28`, `usercommands/suicide.go:96` (player death, kept on
this path on purpose: comment `suicide.go:93-95`), `main.go:390` (goodbye on server stop).

### 2g. Prompt redraw

- `events.RedrawPrompt{UserId, OnlyIfChanged}` (`eventtypes.go:314-320`), unique per user.
- Handler `RedrawPrompt_SendRedraw` (`RedrawPrompt_SendRedraw.go:12-44`): `GetCommandPrompt()`;
  with `OnlyIfChanged`, compares with temp data `cmdprompt`; `AnsiParse`; `connections.SendTo`.
- Queued at priority 100 by 2a, 2b, 2f and `NewRound_AutoHeal.go:72` (OnlyIfChanged).
- `GetCommandPrompt` (`userrecord.prompt.go:65-114`) returns `AnsiMoveCursorColumn + AnsiEraseLine +
  prompt + GetUnsentText() + IAC GA (if SendTelnetGoAhead)`; an active question prompt replaces the
  status prompt.
- Synchronous prompt writes bypassing the queue: `world.go:563` (blank input) and `world.go:592`
  (after a command when an input prompt was or is active), both `connections.SendTo(AnsiParse(
  GetCommandPrompt()))` on the world goroutine inside `processInput`. They reach the client before
  the command's own queued text.

### 2h. `WebClientCommand`

- `events.WebClientCommand{ConnectionId, Text}` (`eventtypes.go:93-98`); handler
  `WebClientCommand_SendWebClientCommand.go:10-25` cancels for non-websocket connections, else
  `connections.SendTo([]byte(cmd.Text))`.
- Senders: `main.go:1067` (`TEXTMASK:false` after websocket login). `login_prompt_handler.go:385-391`
  sends `TEXTMASK:true|false` synchronously with `SendTo` instead, on purpose (comment `:379-384`:
  the queued version raced the prompt).

### 2i. Raw `net.Conn` writes (`main.go`)

`TelnetListenOnPort`: `conn.Write` "AI connection pool is full" (`:1162`) or "Server is full (%d
connections)" (`:1164`), then `conn.Close()`, before `connections.Add`. `SSHListenOnPort`:
`conn.Write` "Server is full" (`:1211`) on the raw TCP conn before the SSH handshake. Accept
goroutines; no `ConnectionDetails` exists. The current guard does not see these (comment
`guards/messaging_test.go:16-18`).

### 2j. Telnet and terminal negotiation (`internal/inputhandlers`, `main.go`)

- `main.go` `handleTelnetConnection` (`:637-689`): WILL SUPPRESS-GO-AHEAD, WONT LINEMODE, DO NAWS,
  CHARSET request, SuppressGoAhead, WILL EOR, ANSI resolution request; `detectClientType` sends
  NEW-ENVIRON request (`:567`); echo baseline WONT/WILL ECHO (`:725,728`).
- `term_iac.go:81` CHARSET UTF-8 reply; `:131` NEW-ENVIRON SEND request. Module IAC handlers run first
  (`:53-67`; GMCP registers one).
- `term_ansi.go:111-113, 145-147`: backspace/space/backspace runs to redraw the input line for
  history up/down.
- `echo.go:13,22`: echoes typed bytes and CRLF (in the post-login chain on telnet, SSH and websocket:
  `main.go:843, 1022, 1094, 1406`).
- `cleanser.go:31`: backspace erase sequence.
- `login_prompt_handler.go:332,398`: WONT/WILL ECHO around masked steps for Mudlet.
All synchronous on the connection's goroutine. `echo.go` and `echo` of typed input sit outside the
world loop entirely.

### 2k. Input handlers, login and other pre-actor output (`internal/inputhandlers`)

- `login.go` `FinalizeLoginOrCreate` (`:18-157`): 21 `SendTo`s: goodbye template to a kicked
  duplicate session (`:45`), login failure texts, `LoginUser`'s message (`:73,82`), creation errors,
  each followed by `term.CRLF`, then `connections.Remove`.
- `login_prompt_handler.go`: per-byte echo or mask for raw telnet (`:235,240`); websocket echo of the
  submitted line or mask (`:263,266`); CRLF (`:275`); validation errors (`:291-292`); "Too many
  mistakes." (`:302`); TEXTMASK (`:390`); Mudlet ECHO (`:332,398`); the step's prompt template
  (`:401`).
- `systemcommands.go:129`: `/quit` goodbye template, then `Kick`; `:172` shutdown countdown
  `Broadcast` from a goroutine.
- `main.go`: connect splash for telnet, websocket, SSH (`:709, 965, 1326`); "This port is for AI
  clients" (`:712`); AI/human port warnings after login (`:860, 862`); AI rate-limit drop (`:904`);
  websocket copyover reconnect message (`:1011, 1017`); "Copyover complete." (`:376`).
- `hooks/PlayerDespawn_HandleLeave.go:63`: goodbye template to the connection at logout, then
  `LogOutUserByConnectionId` (the actor is going away).
- `hooks/Log_FollowLogs.go:56`: log lines to admins following logs, by connection ID.
All of these except the last two run on connection or accept goroutines, not the world goroutine
that `messaging.Send` requires.

### 2l. Web client websocket write path

`web.Listen` upgrades `/ws` and hands the `*websocket.Conn` to `main.HandleWebSocketConnection`
(`web.go:490-500`, `main.go:311, 940`). It is registered with `connections.Add(nil, conn)` (`:943`)
and every write then goes through `ConnectionDetails.Write` (2a step 5), plus the heartbeat ping
`WriteControl` under the same `wsLock` (`heartbeat.go:89-106`). `internal/web` itself has no
`SendText`, `connections.` or event-message calls (grep returned nothing; the same pattern matches in
other packages).

### 2m. GMCP (`modules/gmcp`)

`GMCPOut` events → `dispatchGMCP` writes `GmcpPayload` or `GmcpWebPayload` with `connections.SendTo`
(`gmcp.go:551-602`); `sendGMCPEnableRequest` (`:202-207`) on net connect and on `PlayerSpawn`. Seven
`SendTo`s, all in `gmcp.go`, which the guard allows permanently. Separately, `gmcp.Mudlet.go` has 20
`user.SendText` calls (the `mudletui` command and the Mudlet-detected notice), which are ordinary
text and not GMCP.

Also exposed but unused: `plugins.NetConnection` has `Write` (`plugincallbacks.go:28`); no module
calls it (grep of `modules` for `.Write(` found none).

## 3. Import graph

`go list` on this branch:

- **messaging imports** `actors`, `characters`, `ids`, `mudlog`, `perception`, `rooms`.
- **messaging's transitive internal deps** (cannot import messaging): `mudlog copyover term util
  configs fileloader gametime statmods buffs colorpatterns connections ids items stats events races
  skills spells characters mobs prompt accounts perception actors gamelock exit keywords mutators
  rooms`.
- **Who imports messaging today:** `hooks`, `usercommands`. Packages whose deps include it: root,
  `hooks`, `plugins`, `usercommands`, `modules/*` (cleanup, follow, gmcp, webhelp).
- **Who imports `accounts`:** root, `cmd/reset-admin-pw`, `actors`, `behaviour`, `combat`, `hooks`,
  `inputhandlers`, `mapper`, `rooms`, `templates`, `usercommands`, `web`, `modules/follow`,
  `modules/gmcp`.
- **Who imports `rooms`:** root, `behaviour`, `combat`, `hooks`, `mapper`, `messaging`, `procedural`,
  `usercommands`, `web`, `modules/cleanup`, `modules/follow`, `modules/gmcp`.
- **Who imports `usercommands`:** root, `hooks`, `plugins`, `modules/cleanup`, `follow`, `gmcp`,
  `webhelp`.
- **Who imports `hooks`:** root only.
- **`events` imports** `connections`, `ids`, `items`, `mudlog`, `stats`, `util`.
- **`templates`' deps** do not include `rooms`, `actors` or `messaging`, so messaging can import
  `templates`, `connections`, `term`, `events`, `util` without a cycle.
- **`inputhandlers` imports** `accounts configs connections events language mudlog templates term`;
  none depends on messaging, so `inputhandlers` can import messaging.

What can be re-pointed in place (same signature, body calls messaging):

| Primitive | In place? | Why |
|---|---|---|
| `usercommands.Tell` | yes | usercommands already imports messaging |
| `hooks.Message_SendMessage`, `Broadcast_SendToAll`, `RedrawPrompt_SendRedraw`, `WebClientCommand_…` | yes, or move the bodies into messaging | messaging may import `events`, `connections`, `templates`, `term` (all already in its dep set or cycle-free); main or hooks registers them |
| `world.go` prompt writes, `main.go` writes | yes | root imports everything |
| `inputhandlers` writes | yes | no cycle (above) |
| `Account.SendText` | **no** | `accounts` is in messaging's deps (messaging → actors → accounts). Possible only through a function variable or hook set by messaging at init (the `actors.OnBehaviourTick` pattern, `controller.go:81-85`), and then `ConnectionController.Perceive` (which calls `Account.SendText`) would recurse unless it uses a separate low-level write |
| `Room.SendText`, `SendTextCommunication`, `SendTextToExits` | **no** | messaging imports `rooms`; same function-variable workaround only |
| `rooms` package's own sends (`rooms.go:248,251,534`, `roommanager.go:197,201,216`) | **no** | same; they would need a hook, or to return what was pruned to a caller outside `rooms` |
| `accounts.GrantXP` text (`userrecord.go:181,187`) | **no** | same |
| `ConnectionController.Perceive` (`actors/controller.go:44`) | **no** | `actors` is in messaging's deps; it is the delivery endpoint and must reach the socket through something `actors` can import (accounts or connections) or an injected writer |

## 4. Guards in `internal/guards`

- `TestOnlyMessagingWritesToConnections` (`messaging_test.go:19-73`): finds `connections.SendTo`,
  `SendRawTo` and `Broadcast` selectors whose receiver identifier is literally `connections`.
  Allowed: anything under `internal/connections/`, `internal/messaging/`, `internal/perception/`;
  `modules/gmcp/gmcp.go` permanently; a shrinking list of 15 files (`:24-40`), each of which must
  still write or the test fails (`:68-72`). Passes today (`go test ./internal/guards -run
  'TestOnlyMessagingWritesToConnections|TestCombatPackageDoesNotSend' -count=1`: PASS).
- What it does not see: `Account.SendText`, `Room.SendText*`, `Tell`, `events.Message{}`,
  `events.Broadcast{}`, `events.RedrawPrompt{}`, `events.WebClientCommand{}` constructions, raw
  `net.Conn.Write` in `main.go`, `cd.Write` on a `*ConnectionDetails`, `plugins.NetConnection.Write`,
  and a `connections` package imported under another name.
- `TestCombatPackageDoesNotSend` (`combat_test.go:60-103`) is the ready pattern for a broader ban:
  call names `SendText`, `Tell`, `SendTextToExits`, `SendTextCommunication`, `messaging.Send`, and
  composite literal `events.Message`, with a positive control on files scanned. It is scoped to
  `internal/combat/`.
- `goFiles` (`guards_test.go:18-47`) parses every non-test `.go` file outside `.git` and
  `_datafiles` and fails below 100 files.

A slice 7 guard needs, at least:
1. The existing connection-writer check with the allowlist reduced to GMCP (plus any second
   exception Calabe rules on, section "Open questions").
2. A ban outside `internal/messaging` on calls named `SendText`, `SendTextCommunication`,
   `SendTextToExits`, `Tell`, `SendWebClientCommand`, and on composite literals `events.Message`,
   `events.Broadcast`, `events.RedrawPrompt`, `events.WebClientCommand` (or the event types deleted).
   A name ban catches every receiver type, which matters because `SendText` exists on both
   `*Account` and `*Room`.
3. A ban on `.Write(` with a `[]byte` argument whose receiver is a `net.Conn`, `*ConnectionDetails`
   or `plugins.NetConnection` outside `connections` and `messaging`. A syntactic guard cannot tell a
   hasher's `Write` from a socket's; it needs either type info (`go/types`, as the scratchpad counter
   does) or a file allowlist for the hashers (`accounts/index.go`, `accounts/userrecord.go`,
   `procedural/gridmaze.go`, `util/util.go`).
4. A positive control per check (a known count of files scanned, and a deliberate violation in a test
   fixture if wanted).

## 5. Login and pre-actor output

Listed in 2i, 2j and 2k. Summary of when an actor does and does not exist:
- No account, no actor: accept-time "server full" (raw `net.Conn`), connect splash, telnet
  negotiation, the whole login and creation prompt sequence, login errors, `LoginUser` messages.
- Account and actor exist, but off the world goroutine: `LoginUser` and `CreateUser` fire
  `OnActivate` (which binds the actor, `actors/wire.go:20`) before they return
  (`accounts/users.go` `activated(user)` in both). So `login.go:82` (`LoginUser`'s success message) and
  `main.go:860,862` (port warnings) already have an actor, but run on the connection goroutine,
  where `messaging.Send` must not run.
- Actor going away: `PlayerDespawn_HandleLeave.go:63` goodbye, `/quit` goodbye (`systemcommands.go:129`),
  duplicate-login goodbye to the old session (`login.go:45`).
- Copyover: `fd_unix.go:32-38` (inside connections), `main.go:376`, `main.go:1011,1017`.

No messaging function exists for any of these. Two constraints for the one the spec asks for: it must
be safe off the world goroutine (it runs on connection and accept goroutines), and "server full" has
no connection ID yet (it writes a bare `net.Conn`).

## 6. Per-receiver variance

| Variance | Where | Today |
|---|---|---|
| Screen reader, text form | `Message_SendMessages.go:28,76` strip characters for `ScreenReader` accounts | in the socket writer, so it applies to messaging output too (messaging → `SendText` → this handler) |
| Screen reader, alternate text | `Broadcast.TextScreenReader` (`Broadcast_SendToAll.go:26-52`) | no sender sets it |
| Screen reader, templates | `templates.Process(name, data, receivingUserId)` tries `*.screenreader.template` first (`templates.go:146-176`) | chosen when the text is built, by the caller |
| Screen reader in commands | `look.go:455`, `set.go:31,334,339` | caller-side (inventory K) |
| Viewer-dependent names | `Character.GetPlayerName/GetMobName(viewer)` (`character.go:590-596`); the viewer only adds an aggro-on-you adjective (`getFormattedName`, `:658-685`). Callers: `rooms/roomdetails.go:263,283`, `usercommands/look.go:113` | all build text for one viewer (room descriptions, look), so they are interface text for that viewer; messaging's `nameFor` has no adjectives |
| Colour | `templates.AnsiParse` uses one global flag (`templates.go:434-455`); per connection only `stripAnsi` on the AI port (`connectiondetails.go:275-279`) | no per-account colour setting found (grep for colour config options returned nothing) |
| Language | `language.T` used in `inputhandlers/login.go` and `usercommands/online.go` only | not per receiver |
| Erase-line and prompt | 2a, 2f (`SkipLineRefresh`) | per write, not per receiver |

`ConnectionController.Perceive` renders by passing `p.Text` (tokens already resolved per receiver by
`deliver`) to `Account.SendText`; AnsiParse and the screen-reader strip happen later in
`Message_SendMessage`. So the three tokens are the only per-receiver rendering inside messaging; the
rest happens downstream per account, and stays correct as long as messaging output keeps ending in
that writer.

## 7. Conversion strategies

### (a) Every call site calls messaging directly

Each `Tell` becomes `messaging.Send(Interface, ActorOnly)`; each `Room.SendText*` pairing with its
`Tell` becomes one `WorldEvent` with roles and tokens; system broadcasts, prompts and pre-actor text
call new messaging functions. `Account.SendText`, `Room.SendText*`, `Tell`, `ExcludeId`,
`events.Message`, `events.Broadcast` and their hooks are deleted; the socket writers move into
messaging (no cycle, section 3).
- Messaging additions needed: raw-text flag, exclusion or "everyone but" audience, an "all
  accounts" interface audience with no sense check, a connection-level write (goroutine-safe), a raw
  `net.Conn` write, prompt redraw, `SkipLineRefresh`.
- The packages that cannot import messaging (`rooms` 6 sends, `accounts` 2, `actors` 1) need their
  text moved to callers or injected writers.
- `ConnectionController.Perceive` needs a low-level write it can reach (an unexported account write
  or an injected function).
- Guard: strongest. Ban every name in section 4 item 2 outside messaging; connection writers only in
  messaging and GMCP.
- Churn: every one of 555 + 76 + 10 + 86 + 226 sites; about 60 files; this is the split by command
  group the slice row describes.
- Player-visible: receivers resolve at send time, not at drain; NPCs start perceiving room text
  (no handler yet); names in converted lines may read "Someone" in dark rooms or for hidden actors
  (the spec's intended differences); user text needs the raw flag or it is token-substituted.

### (b) Re-point the primitives, keep their signatures

`Tell` calls messaging in place. `Account.SendText` and `Room.SendText*` become thin wrappers that
call messaging through a function variable set at init (they cannot import it). The socket writers
move into messaging. Call sites do not change.
- Requires: a split between "send text to this account" (which would go through messaging and its
  actor) and the controller's own write (to avoid `Perceive → SendText → Send → Perceive`
  recursion); an exclusion audience for `Room.SendText(…, exclude...)`; a sense choice for room text
  with no actor (today there is no sense check at all, so any sense other than "none" drops text in
  dark rooms).
- Guard: weak on intent. It can ban connection writers and `events.Message{}` outside messaging, but
  `SendText` calls stay legal everywhere, so the slice row's "Deletes: direct `SendText`" is met only
  if a re-pointed `SendText` counts as not direct.
- Churn: small (tens of lines plus messaging additions).
- Player-visible: near parity, but room text carries no roles or actor, so NPC behaviours get
  unstructured text and "Someone" never applies to it.

### (c) Mix: bottom layer first, then call sites by group

1. Move the socket writers into messaging (`Message_SendMessage`, `Broadcast_SendToAll`,
   `RedrawPrompt_SendRedraw`, WebClientCommand, the `world.go` prompt writes, `PlayerDespawn`
   goodbye, log follow) and add the pre-actor and raw-conn functions; point `inputhandlers` and
   `main.go` at them. This alone shrinks the connection-writer allowlist to GMCP (plus any ruled
   exception).
2. Re-point `Tell` in place (one body, 226 sites unchanged), with a raw-text flag so user text is
   not token-substituted.
3. Convert `Room.SendText*` + paired `Tell` sites by command group (inventories K and L) to
   `WorldEvent` messages with roles; delete `Room.SendText*`, `ExcludeId`, `events.Message`'s room
   branch when the last caller goes.
4. Move the `rooms`/`accounts` own sends to callers or a hook.
- Guard: grows per step with a shrinking allowlist per name, the same shape as today's.
- Churn: same end state as (a), but each step builds and passes on its own, and steps 1-2 are small.
- Player-visible: as (a) once step 3 lands.

## Open questions (decisions for Calabe)

1. **Telnet and terminal negotiation: second named exception, or messaging's?** (`term_iac.go`,
   `term_ansi.go`, `echo.go`, `cleanser.go`, `main.go` negotiation and echo, Mudlet ECHO toggles.)
   Player-visible effect: none either way; it decides whether input echo and IAC bytes count as
   "output" that must live in messaging.
2. **WebClientCommand (`TEXTMASK`): with GMCP as an exception, or in messaging?** Effect: none
   visible; the password field masking must stay synchronous before the prompt
   (`login_prompt_handler.go:379-391`), whichever owns it.
3. **Receivers resolved at send time instead of drain time** for room text. Effect: in the same
   millisecond tick, a player who walks in or out between the send and the drain may now see (or
   miss) a line they missed (or saw) before. Accept, or keep late resolution?
4. **Prompt order.** `world.go:563,592` write the prompt immediately, before the command's queued
   text, then the redraw repeats it after. Keep the synchronous write, or queue it so the prompt
   always follows the text? Effect: a prompt line that today can flash before command output.
5. **NPCs perceiving converted room text and broadcasts.** Today NPCs never receive room text or
   broadcasts. Through messaging, `InRange` world events reach every NPC in range (no handler reacts
   yet). Should system notices (save, shutdown, reload, the player-death broadcast) reach NPCs at
   all? Effect: none until a behaviour reacts; then NPCs could react to "Saving users...".
6. **System notices with no sense check.** Save, reload, shutdown and goodbye broadcasts have no
   messaging form: as `World` world events they are dropped in dark rooms (Sight) or need Sound.
   Add an "every account, interface" audience? Effect: whether a player in a dark room still sees
   "Saving users...".
7. **Raw text.** User-typed text with `{actor}`, `{target}` or `{direction}` is substituted today if
   it passes through `messaging.Send` (a `say {target}` would print an empty name). Add a raw flag,
   or escape braces in user text? Effect: what players see when they type braces.
8. **Interface text for an account with no bound actor** (login window, logout) is dropped silently
   by `Send`. Should messaging's interface path accept an account or connection when no actor exists?
   Effect: lines like the logout goodbye or post-login warnings could vanish if routed through
   `Send` as it is.
9. **Strategy (a), (b) or (c),** and whether a re-pointed `SendText` wrapper meets "Deletes: direct
   `SendText`". Effect: (b) keeps today's text exactly but leaves room lines without roles, so
   "Someone" never applies to them and NPC behaviours get no actor.
10. **`print` keeps sending without a newline** (`print.go:17`, raw `events.Message`) while every
    other path appends one. Keep that quirk when it moves? Effect: whether `print` output runs into
    the next line.

# Inventory K: output sites in `internal/usercommands` (Phase 4b slice 7)

Branch `strip/phase-4b-7-output` at `62cc22ac`, read 2026-10-28. Read-only; no repo file changed.

**Read 68 of 68 files** (all non-test `.go` files in `internal/usercommands`). Every file was also counted by grep (script below).

- **Full (49):** noop, save, motd, biome, quit, who, default, flee, broadcast, history, read, macros, whisper, shout, say, emote, print, password, consider, alias, visited, online, inventory, keyring, help, show, lock, drop, put, remove, equip, give, get, go, look (1-615, two chunks), attack, shoot, suicide, suicide_npc, party (1-655, two chunks), set (1-414, two chunks), start, usercommands.go, admin.spawn, admin.teleport, admin.rename, admin.redescribe, admin.mute. `unlock.go`: read as a full `diff lock.go unlock.go`; lock.go read in full, so every unlock line is accounted for.
- **Partial (4):** admin.locate.go 30-60 plus grep; admin.modify.go 55-80 plus grep; admin.item.go 115-145 plus grep; admin.mob.go 135-165 plus grep.
- **Skimmed by grep only (15):** admin.room.go (1712 lines), admin.zone.go, admin.server.go, admin.build.go, admin.visit.go, admin.copyover.go, admin.prepare.go, admin.syslogs.go, admin.badcommands.go, admin.reload.go, admin.command.go, admin.grant.go, descriptions.panels.go, inventory.panels.go, look.panels.go. For these I read only the lines matching `SendText|Tell(|templates.Process|Account()|StartPrompt|events.(Broadcast|Message)`. The receiver grep shows that every send in them goes to `user` (the admin), apart from the lines named in the table. The three panel files have zero send matches: they build strings, and their callers send them.
- **Not read:** the bodies of admin.room.go, admin.zone.go, admin.server.go and admin.build.go beyond their send lines. They need a full read before a subagent converts their prompt menus (section 4g).

Counting script: `count.sh` in the scratchpad (per-file `grep -cE`). The receivers were listed with `grep -ohE "[A-Za-z_.()]+\.SendText(Communication|ToExits)?\("`.

---

## 0. Facts verified against source

| Fact | Source |
|---|---|
| `Tell(actor, text)` writes to `actor.Controller().Account()` via `user.SendText`; it writes nothing for an NPC. Comment: "slice 7 points this at messaging" | `usercommands.go:35-41` |
| `ExcludeId(actor)` = account `UserId`, or 0 for an NPC; `NameColour(actor)` = `username`/`mobname` by account | `usercommands.go:43-59` |
| `Account.SendText` queues `events.Message{UserId, Text: txt+"\n"}` | `accounts/userrecord.go:334-341` |
| `Room.SendText` / `SendTextCommunication` queue `events.Message{RoomId, Text: txt+"\n", ExcludeUserIds}`; `IsCommunication` is set and never read anywhere (grep: only eventtypes.go:64,77, rooms.go:269, broadcast.go:30) | `rooms/rooms.go:262-283` |
| `Room.SendTextToExits(txt, isQuiet, exclude...)` walks `Exits`+`ExitsTemp` (no mutator exits), and queues `(From <exit>) txt` for each back-exit found | `rooms/rooms.go:285-315` |
| The `RoomId` message is resolved **at delivery**: `room.GetPlayers()` minus `UserId`/`ExcludeUserIds`, with `IsQuiet` needing `superhearing` | `hooks/Message_SendMessages.go:38-85` |
| `SendTextCommunication` and `SendTextToExits` have no non-test caller outside `usercommands` (grep of internal+modules) | grep |
| `messaging.Send`: Interface or ActorOnly reaches only the actor (`messaging.go:38-43`); others are resolved **at send time** from the actor's room, the target's room, or `m.Room` when there is no actor (`sourceRooms`, `messaging.go:117-139`) | `messaging.go` |
| Far receivers (Nearby, Zone, World) never see the actor: `canSee` returns false when `far`, so `{actor}` reads "someone" | `messaging/perceive.go:44-55` |
| Zone and World ranges skip any actor with no loaded room | `messaging.go:178-185` |
| `ConnectionController.Perceive` drops empty text; otherwise it calls `account.SendText(p.Text)` (so the queued `events.Message{UserId}` path and its "\n" stay) | `actors/controller.go:40-45` |
| `{actor}` renders `<ansi fg="username|mobname">Name</ansi>`, the same markup as `NameColour`; an opening "someone" is capitalised | `messaging/render.go:11-48` |
| `connections.SendTo`: 0 calls in usercommands; `internal/connections` not imported (grep exit 1; the same pattern finds 2 hits in Message_SendMessages.go, so the search works) | grep |
| Totals: `.SendText(` 581 (514 to an account, 67 to a room), `SendTextCommunication` 8, `SendTextToExits` 2, `Tell(` 204 (excluding `func Tell`), `events.Message{` 1, `events.Broadcast{` 2, `messaging.` 6, `events.Communication{` 4. Sum 804 = group totals in section 2 | count.sh |
| `Tell`/`ExcludeId`/`NameColour` outside the package: 28 lines, in `hooks/NewRound_DoCombat.go` (6), `modules/cleanup/cleanup.go` (8), `modules/follow/follow.go` (14) | grep |
| Blank-line sends: `Tell(actor, `` )` 15 (look 13, put 2); `.SendText(`` )` 65 (admin.room 23, set 9, admin.mob 8, admin.server 8, admin.item 6, admin.zone 4, start 4, macros 2, alias 1). Also `look.go:130-131` `Tell(actor, statusTxt/invTxt)`, which are always "" | grep |
| Smoke never compares transcripts in code: there is no Go reference to `pre-4b`/`pre-5`/`testdata` under cmd, internal or modules | grep |

---

## 1. Per-file table

Columns: **Acct** = `.SendText(` on an account (`user`, `u`, `targetUser`, `partyUser`, `toUser`, `locateUser`, `invitedUser`); **Room** = `room`/`destRoom`/`gotoRoom`/`otherRoom` `.SendText(`; **Comm** = `SendTextCommunication`; **Exits** = `SendTextToExits`; **Tell**; **Msg/Bc** = `events.Message{` / `events.Broadcast{`; **MSG** = `messaging.` calls. `connections.SendTo` is 0 in every file and is left out.

Classes: **i** = interface text to the actor only; **ii** = room or observer text; **iii** = text to a named other actor; **iv** = world or broadcast.
Flags: **PR** = text built per receiver; **EX** = excludes someone; **AD** = admin-only; **NPC** = can run with `Account()` nil, so output may be written while the actor is an NPC; **G** = returns early when `Account()` is nil (an NPC gets nothing); **HID** = how hidden is handled.

| File | Read | Acct | Room | Comm | Exits | Tell | Msg/Bc | MSG | Classes | Flags and notes |
|---|---|---|---|---|---|---|---|---|---|---|
| admin.badcommands.go | skim | 1 | | | | | | | i | AD G; table template |
| admin.build.go | skim | 16 | | | | | | | i | AD G |
| admin.command.go | skim | 1 | | | | | | | i | AD G; makes an NPC run a command (smoke uses it, admin.go:70) |
| admin.copyover.go | skim | 2 | | | | | | | i | AD G |
| admin.grant.go | skim | 5 | | | | | | | i | AD G; the XP text the target reads comes from `GrantXP` (outside the package) |
| admin.item.go | partial | 35 | 1 | | | | | | i, ii | AD G; `:133` room line excludes `user.UserId`; prompt `item create` `:163`; helpers take `*accounts.Account`, not the actor (`:76,121,150`) |
| admin.locate.go | partial | 6 | | | | | | | i, iii | AD G; `:45` `locateUser` anywhere in the world: "You get the feeling someone is looking for you..." |
| admin.mob.go | partial | 38 | 1 | | | | | | i, ii | AD G; `:153` room line excludes `user.UserId`; prompt `mob create` `:192`; helpers take `*accounts.Account` (`:80,125`) |
| admin.modify.go | partial | 9 | | | | | | | i, iii | AD G; `:74` the target account: "Your role has been set to" |
| admin.mute.go | full | 6 | | | | | | | i | AD G |
| admin.prepare.go | skim | 2 | | | | | | | i | AD G |
| admin.redescribe.go | full | 3 | 1 | | | | | | i, ii | AD G; `:54` EX admin |
| admin.reload.go | skim | 8 | | | | | | | i | AD G |
| admin.rename.go | full | 3 | 1 | | | | | | i, ii | AD G; `:58` EX admin |
| admin.room.go | skim | 137 | | | | | | | i | AD G; prompts `:609`, `:1206`, `:1589`; 23 blank lines; helpers take `*accounts.Account` (`:566,1146,1168,1582`) |
| admin.server.go | skim | 27 | | | | | | | i | AD G; prompt `server config` `:322`; 8 blank lines |
| admin.spawn.go | full | 4 | 3 | | | | | | i, ii | AD G; `:56,:82,:95` EX admin |
| admin.syslogs.go | skim | 3 | | | | | | | i | AD G; log-follow output itself is written outside the package |
| admin.teleport.go | full | 9 | 2 | | | | | | i, ii, iii | AD G; `:108` destination room EX the moved player (who may not be the admin); `:132` partyUser; `:133` "appears" sent to the **origin** room inside a loop over `room.GetPlayers()` (`:119`) |
| admin.visit.go | skim | 9 | | | | | | | i | AD G |
| admin.zone.go | skim | 23 | | | | | | | i | AD G; prompt `zone edit` `:147` |
| alias.go | full | 6 | | | | | | | i | G; tables |
| attack.go | full | | | | | 6 | | 2 | i, ii (+target) | NPC; `:224`, `:274` already `messaging.Send` (Tell for the actor, Send with ObserverText/TargetText); HID by messaging |
| biome.go | full | | | | | 2 | | | i | NPC |
| broadcast.go | full | | | | | 1 | Bc 1 | | i, iv | NPC (an NPC can broadcast); `:28` Broadcast, `:34` `events.Communication` (GMCP) |
| consider.go | full | | | | | 4 | | | i | NPC |
| default.go | full | | | | | | | | none | calls `Look` `:22`, queues `storage` `:17` |
| descriptions.panels.go | skim | | | | | | | | none | string builders |
| drop.go | full | | 3 | | | 8 | | | i, ii | NPC EX; HID cancelled before room lines (`:62`, `:94`); `drop all` recurses (`:31`, `:42`) |
| emote.go | full | | 2 | 1 | | 4 | | | i, ii | NPC EX; **no hidden check** (today a hidden emoter is named); `@` prefix suppresses the actor echo (`:123-127`); actor echo hardcodes `username` (`:110`, `:126`) |
| equip.go | full | | 3 | | | 7 | | | i, ii | NPC EX; HID cancelled `:58`; re-wear is silent (`:53-56`) |
| flee.go | full | | | | | 1 | | | i | NPC |
| get.go | full | | 11 | | | 27 | | | i, ii | NPC EX; **hidden not cancelled** on the corpse "get all" path (`:97-155`) or the noun path (`:398-403`); `get all` recurses (`:26`, `:33`) |
| give.go | full | 2 | 5 | | | 12 | | | i, ii, iii | NPC EX (actor + target `:86-89`, `:134-137`); targetUser `:83`, `:131`; mob target reacts by `m.Command(emote)` `:205,:207`; `:171` names a mob in `username` colour |
| go.go | full | 1 | 6 | | 1 | 15 | | | i, ii, iii | NPC EX; `:198` leave line sent to the **old room after `MoveToRoom` (`:181`)**; `:211` exits line excludes all origin-room players; `:228` partyUser in a loop; HID: no leave/enter/exits lines when sneaking (`:186`); GMCP `:141`, `:307` |
| help.go | full | | | | | 2 | | | i | NPC; template |
| history.go | full | 1 | | | | | | | i | G; table |
| inventory.go | full | | | | | 1 | | | i | NPC; panel |
| inventory.panels.go | skim | | | | | | | | none | builders |
| keyring.go | full | | | | | 1 | | | i | NPC; template rendered for `ExcludeId(actor)` `:164` |
| lock.go | full | | 4 | | | 10 | | | i, ii | NPC EX; **no hidden check**; room lines hardcode `username` (`:49,69,101,121`) |
| look.go | full | 1 | 9 | | | 43 | | | i, ii, iii | NPC EX PR; `:87` target player; `:91-93` excludes **only the target** (the actor reads its own "X is looking at Y"); `:113` `GetMobName(ids.ID{})`; `:429`, `:434` sent to **another room** (lookRoom through an exit); HID checks at `:54,86,112,242,273,309`, none at `:379`; 13 blank lines plus `:130-131` |
| look.panels.go | skim | | | | | | | | none | builders |
| macros.go | full | 8 | | | | | | | i | G |
| motd.go | full | | | | | 1 | | | i | NPC; template |
| noop.go | full | | | | | | | | none | |
| online.go | full | | | | | 1 | | | i | NPC; staff columns by role |
| party.go | full | 51 | | | | | | | i, iii | G; 38 lines to the actor, 13 to other accounts anywhere (`:124,173,204,440,442,480,498,503,563,568,615,621,640`); `:646` Communication; `:640` party chat line has no closing quote |
| password.go | full | 4 | | | | | | | i | G; prompt `:17` |
| print.go | full | | | | | 1 | Msg 1 | | i | `:17` raw `events.Message{UserId}` with **no "\n" appended**; `printline` `:59` |
| put.go | full | | 3 | | | 10 | | | i, ii | NPC EX; **no hidden check**; `:134` overflow line has **no exclude** (the actor reads it too) |
| quit.go | full | 1 | | | | | | | i | G |
| read.go | full | | 1 | | | 3 | | | i, ii | NPC EX; HID-gated `:41`; hardcodes `username` `:43` |
| remove.go | full | | 1 | | | 6 | | | i, ii | NPC EX; HID cancelled `:65`; `remove all` recurses `:25` |
| save.go | full | 2 | | | | | | | i | G |
| say.go | full | | | 2 | 1 | 2 | | | i, ii | NPC EX; hidden reads lowercase "someone says" (`:29`); `:36` quiet exits line, no exclude; `:38` Communication |
| set.go | full | 51 | | | | | | | i | G; 9 blank lines |
| shoot.go | full | | | | | 9 | | 1 | i, ii | NPC; `:27` already messaging (no Target, on purpose: `:21-25`) |
| shout.go | full | | | 5 | | 2 | | | i, ii | NPC EX; hidden "someone shouts" (`:30`); neighbour rooms through Exits, ExitsTemp and mutator exits (`:35-63`) |
| show.go | full | 2 | 2 | | | 7 | | | i, ii, iii | NPC EX; targetUser `:59`, `:63` (item long description as a second send); `:68` EX target + actor; `:98` mob branch tells the mob nothing |
| start.go | full | 21 | 1 | | | | | | i, ii | G; prompt `:31`; `:210` destination room after `MoveToRoom`, EX user; 4 blank lines |
| suicide.go | full | 4 | | | | | Bc 1 | 1 | i, iv, ii | player death; `:47` messaging; `:96` Broadcast kept on purpose (`:93-95` "slice 7:" comment); `:137` permadeath template; calls `Remove`/`Drop` (`:144,147,193,202,215,226,237`), which write their own lines |
| suicide_npc.go | full | 6 | 3 | | | | | 2 | iii, ii | NPC-only path; all 6 account sends go to **other** accounts (damagers and party members, possibly in other rooms): `:166,197,199,265,297,299`; `:325,356,372` actorless drop lines with no exclude; `:219-221` "slice 7:" comment |
| unlock.go | full (diff) | | 4 | | | 10 | | | i, ii | as lock.go |
| usercommands.go | full | 4 | | | | 2 | | | i | router: `:305-306` plaintext-password nag, `:316` downed (Tell), `:347` no permission; `:39` is Tell's own body; `:68` permaGearRefused |
| visited.go | full | | | | | 1 | | | i | NPC; template rendered for `ExcludeId(actor)` |
| who.go | full | 1 | | | | | | | i | G; template |
| whisper.go | full | 1 | | | | 5 | | | i, iii | NPC; `:44` toUser anywhere; `:48` Communication |

GMCP calls (allowed exception): `go.go:139-142` and `go.go:305-308`, through `GetExportedFunction("SendGMCPEvent")`.

---

## 2. Command groups

**Step 0, by the controller before the groups start:** repoint `Tell` (usercommands.go:37) at `messaging.Send{Kind: Interface, Actor: actor.Id, ActorText: text}`. That converts all 204 in-package `Tell` sites and the 28 outside callers without touching them. It needs the blank-line ruling first (Open question 1). Also decide the shared helpers the groups will call, so they are written once: an Interface send to an account that has no actor argument (for the admin helpers that take `*accounts.Account`), a far-only send for "You hear someone ...", and an exclude list.

| Group | Files | Lines | Sends | Acct / Room+Comm+Exits / Tell |
|---|---|---|---|---|
| A. Admin room and zone editors | admin.room.go, admin.zone.go | 1998 | 160 | 160 / 0 / 0 |
| B. Admin item, mob, server, build | admin.item.go, admin.mob.go, admin.server.go, admin.build.go | 1585 | 118 | 116 / 2 / 0 |
| C. Admin other (cross-actor sends) | admin.badcommands, command, copyover, grant, locate, modify, mute, prepare, redescribe, reload, rename, spawn, syslogs, teleport, visit (15) | 1322 | 78 | 71 / 7 / 0 |
| D. Items | get, drop, put, give, equip, remove, read, lock, unlock, show (10) | 1638 | 141 | 4 / 37 / 100 |
| E. Movement, look and combat | go, look, look.panels, descriptions.panels, default, flee, attack, shoot, consider, suicide, suicide_npc (11) | 2570 | 116 | 12 / 19 / 78 |
| F. Communication and party | say, shout, emote, whisper, broadcast, party (6) | 1046 | 78 | 52 / 11 / 14 |
| G. Character and info | set, start, alias, macros, password, history, who, save, quit, motd, online, help, visited, keyring, inventory, inventory.panels, biome, print, noop, usercommands.go (20) | 2283 | 113 | 99 / 1 / 12 |
| **Total** | 68 | 12442 | **804** | |

Once step 0 is done, the Tell column needs no per-site work, so the real load is Acct + Room. Groups D and E are lighter than they look (37 and 31 non-Tell sites), and A, B and G carry the account sends.

Shared helpers and cross-calls that set the boundaries:
- `lookRoom` and `Look` (E) are called from go.go:292, default.go:22, start.go:217/259 (G) and admin.teleport.go:136/155 (C). Look's body stays in E.
- `Drop`/`Remove` (D) are called from suicide.go (E). Only calls cross the boundary; no edits do.
- `permaGearRefused` and `queue` (usercommands.go, G) are used by drop, equip and remove (D); `permaGearRefused` sends through Tell plus an emote queue, so it needs no change.
- `Help` (G) is called by start.go:80 (G).
- `drunkify` (say.go) is used by shout.go: both in F.
- `othersOf` (attack.go) and `sendShootOpening` (shoot.go): both in E.
- The admin helpers that take `*accounts.Account` (admin.item, admin.mob, admin.room, admin.server, admin.zone, admin.visit) need either the actor threaded through or an account-keyed Interface send. A and B must agree on which.

---

## 3. Patterns

| # | Pattern | Example | Becomes |
|---|---|---|---|
| P1 | Actor-only error or info | `Tell(actor, "Get what?")` get.go:20; `user.SendText("Saving...")` save.go:16 | Repointed `Tell` (Interface, ActorOnly). Account-gated `user.SendText` sites become `Tell(actor, ...)` |
| P2 | "You X" plus "{actor} Xs" to the room, excluding the actor | drop.go:120-126 (`Tell` + `room.SendText(..., ExcludeId(actor))`) | One `Send{WorldEvent, Actor, Range: Room, Senses: ?, ActorText: "You drop ...", ObserverText: "{actor} drops their ..."}`. `<ansi fg="%s">%s</ansi>` with `NameColour` becomes `{actor}` (same markup, render.go:22) |
| P3 | Actor, target and room, excluding both | give.go:80-89; show.go:54-71 | `Send` with `Target`, `TargetText`, `ObserverText`. For show.go the target's second send (long description, `:63`) goes into TargetText |
| P4 | Room line with no exclude (the actor reads it too) | put.go:134; suicide_npc.go:325,356,372 | Actorless `Send{WorldEvent, Room: room.RoomId, ObserverText}` (message.Room is used only when there is no actor, messaging.go:130-134); or Actor with ActorText = ObserverText |
| P5 | Speech with a hand-written "someone" when hidden | say.go:28-32; shout.go:29-33 | One `Send{Senses: Sound, ObserverText: "{actor} says, ..."}`; messaging writes "Someone" (capitalised, a text change from "someone says") |
| P6 | Neighbour rooms | shout.go:35-63 (`otherRoom.SendTextCommunication` per exit) | `Range: Nearby, FarText: "Someone shouts from the {direction} direction, ..."` (the spec's own example) |
| P7 | Faint line to neighbours only | say.go:36, go.go:211 `SendTextToExits(..., true, ...)` | No clean form: see section 4a. A separate Nearby + Faint Send with empty Actor/Observer text, or a far-only audience |
| P8 | Text to a named other account anywhere | whisper.go:44; party.go:173; suicide_npc.go:166; admin.locate.go:45 | `Audience: Named, Range: World` with the name written into the text (far receivers read `{actor}` as "someone"); or a new Named Interface form. See 4c |
| P9 | Templated multi-line block to the actor | who.go:20-21; look.go:528-613; alias.go:70-71 | Interface (ActorOnly) with the rendered text. Templates stay (spec section 2, "Formatting stays where it is") |
| P10 | Blank line | `Tell(actor, ``)` look.go:146; `user.SendText(``)` set.go:28 | **Not expressible**: `ConnectionController.Perceive` drops "" (controller.go:41). 80 sites plus look.go:130-131 |
| P11 | Raw text without a newline | print.go:17-20 | **Not expressible**: every messaging path ends in `SendText`, which appends "\n" |
| P12 | World broadcast | broadcast.go:28; suicide.go:96 | `Range: World`, with the name in the text. Broadcast also reaches roomless users and has `TextScreenReader`/`SkipLineRefresh`; `SourceIsMod` and `IsCommunication` are never read (grep). See 4d |
| P13 | Already converted (prior art) | attack.go:220-231 (`Tell` then `Send` with ObserverText only); shoot.go:26-35 | Fold the `Tell` into `ActorText`, or keep it as the slice-5 pattern |

What messaging can't express yet: blank lines (P10), raw no-newline text (P11), an exclude list (go.go:211), a far-only message (P7), Interface text to a named other actor (P8), a source room other than the actor's current one (4b), receivers with no room (4d), and per-receiver text beyond the three tokens (4a).

---

## 4. Hard cases

### 4a. Per-receiver text beyond `{actor}`/`{target}`
- look.go:113: `m.Character.GetMobName(ids.ID{}).String()`. The mob's formatted name (adjectives) is rendered for no viewer and baked into the room text. Converting it to `{target}` changes the markup (`nameFor` gives the plain `mobname` name, render.go:11-23). Keeping the baked string keeps parity.
- Templates rendered per receiver: `templates.Process(..., userId)` reads that user's ScreenReader setting (templates.go:142-157). Every such site is actor-only (P9), so no conflict.
- Hardcoded `username` colour for an actor that may be an NPC: lock.go/unlock.go:49,69,101,121; read.go:43; show.go:69,99; emote.go:110,126 (actor echo); give.go:171 (mob named in `username`). `{actor}` gives `mobname` for an NPC, a visible change for NPC actors only, in line with `NameColour` and the slice-4 ruling.
- go.go:211, `SendTextToExits` with `room.GetPlayers(rooms.FindAll)...`: today players in the origin room are excluded from the "You hear someone moving around" line. Under Nearby from the destination room, the origin room is a neighbour and its players would get the FarText. **Needs an exclude list.**
- Differences between `SendTextToExits` and Nearby: SendTextToExits uses Exits+ExitsTemp and sends once per back-exit (rooms.go:285-315); `neighbourIds` adds mutator exits and sends once per neighbour (messaging.go:198-216).

### 4b. Sends to an account or room that is not the actor's room
- go.go:198: the "leaves towards" line goes to the old room **after** `MoveToRoom` (`:181`). Through messaging the actor's room is already the destination, and `Message.Room` is ignored when there is an actor. The Send has to move before `MoveToRoom`, or messaging needs a source-room override. Moving it changes ordering against the queued move events.
- look.go:429,434: `lookRoom` tells **another room** "X is looking into the room from the Y exit" (look through an exit, `:246`). The actor is not in that room.
- admin.teleport.go:108 (destination room, the moved player may not be the actor) and `:133` (origin room, in a loop).
- start.go:210: destination room after the move; fine through messaging, since the actor is now there.
- whisper.go:44; party.go 13 lines; admin.locate.go:45; admin.modify.go:74; admin.teleport.go:132; suicide_npc.go 6 lines (damagers can be in another room by shooting through an exit).
- Today a `RoomId` message resolves its receivers when the event is delivered (Message_SendMessages.go:45); messaging resolves at send time. Anyone who enters or leaves between the two now sees a different set of lines.

### 4c. Loops over room players or account lists
- go.go:222-231: party members in the room get "You follow the party leader." and their input is queued.
- admin.teleport.go:119-147: loop over `room.GetPlayers()`, which moves each party member, sends to it, and sends an "appears" line to the **origin** room (a bug: it should go to the destination).
- party.go:168-175, 436-445, 475-482, 493-505, 566-570, 618-624, 635-642: loops over party member and invitee lists, in any room.
- suicide_npc.go:104-215 and 224-310: loops over `PlayerDamage` and party members (alignment and tame lines).
- drop.go:41-43, get.go:32-34, remove.go:21-27: `all` recursion, one Send per item (the order must stay).
- look.go:172-195: recipe list, one `Tell` per line.

### 4d. Output that must stay out of an NPC's perception, or reaches the wrong set
- All Interface text: a BehaviourController drops it (perception.go:15). The account-gated commands (G flag) never write for an NPC.
- New perceptions that NPCs do not get today: show.go:98 (mob target), look.go:112-118 (looked-at mob), give.go:185 (given-to mob). With `Target` set, the NPC receives a Target perception. That is harmless while no NPC reacts (spec notes: reactions must queue, never send inside `OnBehaviourPerceive`).
- Broadcast (broadcast.go:28, suicide.go:96) reaches `accounts.GetAllActiveUsers()` (Broadcast_SendToAll.go:31). World range skips actors with no loaded room (messaging.go:180-183), such as a character in the void during `start`.
- look.go:91-93 excludes only the target, so the looking player reads "Fred is looking at Steve." about itself. Messaging gives the actor ActorText instead.

### 4e. `Tell` callers and what Tell is for
`Tell` is the actor-only (Interface) send that is safe for an NPC: it writes to the account, or nowhere (usercommands.go:35-41, slice 4 decision 11). Callers: 204 sites in 30 files of this package (Tell column), plus `hooks/NewRound_DoCombat.go:135,149,176,344,353,438` and 22 lines in `modules/cleanup` and `modules/follow`. `ExcludeId` serves both as an exclude (room sends) and as a stand-in user ID for templates and mapper (visited.go:63, online.go:102, keyring.go:164, help.go:79, look.go:474,528-612). When room sends go away, the template uses still need an account ID.

### 4f. Templated, multi-line output
who.go:20; history.go:46; alias.go:70,83; online.go:102; keyring.go:164; help.go:79,181; visited.go:63; motd.go:14; party.go:341; suicide.go:136 (permadeath, then `colorpatterns`); look.go:528-612 (room title, panel, signs, who, ground, exits); start.go:58,111; admin `tables/*` and `admincommands/help/*` (see the grep list in section 1). All are actor-only. show.go:63 sends a multi-line item description **to the target**.

### 4g. Prompt and input-blocking interactions
- `cmdPrompt` flows: password.go:17; start.go:31; admin.item.go:163; admin.mob.go:192; admin.room.go:609,1206,1589; admin.server.go:322; admin.zone.go:147. The question text is not sent in usercommands: `Ask` only records it (prompt.go:58-91), and world.go:484-493 drives and redraws it. Inventory L owns that path. In these files only the replies and menus are sends.
- Input blocking: go.go:154 (`CmdBlockInputUntilComplete` requeue after an exit message), usercommands.go:97-100 (`BlockInput` for an account), usercommands.go:321 (`InputBlocked` check). No output, but the ordering of the exit message (`go.go:153`) against the requeue matters.
- usercommands.go:304-307: the plaintext-password nag pre-empts every command.

---

## 5. Parity

**What `cmd/smoke` exercises** (commands sent; script.go:166-243, admin.go:37-138, fight.go:85,168, copyover.go:129):
`look`, `teleport 1`, `item spawn`, `mob spawn`, `command <npc> say ...`, `password` (prompt), `start` (prompt: race, name, confirm), `get`, `inventory`, `drop`, `north`/`south` and `portal` (go), `say`, `attack rat` (bot 0) with bot 1 observing, `quit`, `online`, `suicide`, `copyover` (Linux only).

What it asserts: substrings such as "You pick up the", the room title, `You say, "hello from ...`, `<npc> says, "..."` (an NPC's room line read by the admin), "falls to the ground", "appears in the air", "has DIED!" (the Broadcast), "Your password has been changed!", "Carrying:"/"On the Ground:" fields. On fight days, `fight.go:146-155` checks the observer read a line naming the fighter and the rat (other than "prepares to fight") and no "You" line about the rat.

The transcripts (`-transcript`, recorder.go) are compared by hand against `cmd/smoke/testdata/pre-4b/bot0.txt` and `pre-5/bot0.txt`/`bot1.txt`. No code diffs them (grep), and the spec's masks (run tag, clock, HP/MP prompts, wandering NPC lines) are applied by hand. With one bot, bot 0 sees no other player's room text, so transcripts show only actor text plus NPC and world lines.

**Unit tests:**
- `internal/hooks/combat_golden_test.go` does not run usercommands: it records queued Input, so a queued "suicide" never runs (`:14-16`). suicide.go and suicide_npc.go text is outside the goldens.
- `usercommands/combat_text_test.go` covers attack.go, shoot.go and the NPC death line per receiver.
- `speech_test.go`, `look_test.go` and `items_test.go` capture **`RoomId` messages only** (`captureMessages`, speech_test.go:16-31; `messagesFor`). Converted room lines arrive as per-receiver `UserId` messages, so:
  - positive assertions fail loudly (speech_test.go:66-145, look_test.go:12-25, items_test.go:114,283);
  - **negative assertions pass with no check at all**: items_test.go:259 "no room text" and look_test.go:28 `TestNPCSecretLookQueuesNoMessage`. These must be rewritten to capture per receiver before conversion, or they stop testing anything.

**Where a text change shows / doesn't:**
- Shows: smoke substrings above; combat_text_test; speech/look/items tests (once rewritten); the golden combat files (combat lines only).
- Invisible to every test: all admin output except spawn, teleport and password; party; whisper; broadcast text (only "has DIED!" is checked); give/show target text; lock/unlock/put/remove/read room lines; look variants (exit, item, noun, corpse, container); blank-line spacing; set/alias/macros/help/online/visited/keyring; go's leave/enter/exits lines (a single bot never reads them; slice 4 tests check movement, not text); the hidden-actor and dark-room "Someone" changes outside combat.

---

## Open questions

1. **Blank lines.** 80 sites (plus look.go:130-131) send an empty line for spacing; messaging drops empty text (controller.go:41). Player effect: if blank lines are dropped, look (containers, nouns, items), `set`, `start` and every admin menu lose their spacing. Options: messaging writes a blank line for empty Interface text (keeps the screen as it is), or the spacing goes. Recommend keeping it.
2. **Senses for silent actions.** Room lines for drop, get, put, equip, remove, look, lock, give and emote need `Senses`. Sight only: in a dark room other players get nothing. Sight plus Sound: they read "Someone drops ...". Combat was ruled Sight|Sound (slice 5). Player effect: whether players in the dark notice item handling and looking. No `empty` room is dark today.
3. **Party, whisper and admin notices to players elsewhere.** Party chat and notices, `whisper`, `locate`'s "someone is looking for you", `modify`'s role notice, and the tame/alignment lines in suicide_npc reach accounts in other rooms. Treat them as perceptions (Named, World range, so senses apply: a whisper unheard by the deaf, later) or as channel/interface text that always arrives? Player effect: today they always arrive. The spec puts whisper under `Named`; party is not mentioned.
4. **Broadcasts to players with no room, and the death Broadcast.** Today `broadcast` and the "has DIED!" line reach every active account, including one creating a character in the void; World range would skip them. Keep `events.Broadcast` as a second allowed path, or make World include roomless actors? Player effect: whether a new player at the race prompt sees broadcasts and death notices.
5. **Upstream oddities that change when converted.** (a) look.go:91: the looking player reads "Fred is looking at Steve." about itself. (b) admin.teleport.go:133: "appears in a flash of light" goes to the room the party left; `:102` tells the admin "Moved to room" when it teleported someone else. (c) Hidden actors named today in emote, lock/unlock, put, go's lock lines, look-at-corpse, and the get-from-corpse and noun paths would read "Someone"; say/shout's "someone says" becomes "Someone says". Fix all of these to the messaging rule (recommend) or preserve each?
6. **`print`.** It writes raw text with no newline (print.go:17). There are no Go callers (grep). Player effect: if it goes through messaging, its output gains a line break. Accept, or keep a raw form?

# Phase 4b slice 7, inventory L: output sites outside `internal/usercommands` and `internal/messaging`

Branch `strip/phase-4b-7-output`, HEAD `62cc22ac`, read 2026-10-28. Read-only: no repo file edited.

Read 52 of 52 list entries (entries 51 and 52 each group several files, listed inline; full unless a range is given; "grep" means only grep hits were read):

1. `2026-10-27-phase-4b-npcs-are-players-design.md` 140-361 (section 2 Messaging, section 4 Slices incl. row 7, Out of scope, all of "Notes for later slices"); headings by grep
2. `internal/messaging/messaging.go` 1-225
3. `internal/messaging/perceive.go` 1-66
4. `internal/messaging/render.go` 1-49
5. `internal/actors/controller.go` 1-88
6. `internal/connections/connections.go` 1-352
7. `internal/connections/connectiondetails.go` 1-457
8. `internal/connections/heartbeat.go` 1-113
9. `internal/connections/copyover.go` 1-26
10. `internal/connections/fd_unix.go` 1-120
11. `internal/connections/fd_windows.go` 1-17
12. `internal/connections/clientsettings.go` 1-41
13. `internal/connections/memory.go` 1-18
14. `internal/guards/messaging_test.go` 1-70
15. `internal/guards/doc.go` 1-3
16. `internal/guards/guards_test.go` 16-45 (goFiles); test names by grep
17. `internal/guards/combat_test.go` 55-103
18. `internal/hooks/Message_SendMessages.go` 1-90
19. `internal/hooks/Broadcast_SendToAll.go` 1-70
20. `internal/hooks/RedrawPrompt_SendRedraw.go` 1-44
21. `internal/hooks/WebClientCommand_SendWebClientCommand.go` 1-25
22. `internal/hooks/Log_FollowLogs.go` 1-88
23. `internal/hooks/PlayerDespawn_HandleLeave.go` 1-83
24. `internal/hooks/NewRound_AutoHeal.go` 20-81
25. `internal/hooks/NewRound_InactivePlayers.go` 20-62
26. `internal/hooks/NewRound_RoomRoundTick.go` 20-72
27. `internal/hooks/PlayerDrop_HandlePlayerDrop.go` 20-42
28. `internal/hooks/PlayerSpawn_HandleJoin.go` 20-97
29. `internal/hooks/NewRound_DoCombat.go` 125-195, 335-360, 415-460
30. `internal/hooks/NewRound_ActorUpkeep.go` 130-153
31. `internal/hooks/NewTurn_AutoSave.go` 30-75
32. `internal/hooks/combat_golden_test.go` 1-637
33. `internal/accounts/userrecord.go` 170-195, 325-360
34. `internal/accounts/userrecord.prompt.go` 60-115
35. `internal/rooms/rooms.go` 240-330, 525-540
36. `internal/rooms/roommanager.go` 180-222
37. `internal/events/eventtypes.go` 55-110, 310-322
38. `main.go` 340-1521 (1-339 by grep for Send/Write/Broadcast/Print/websocket: no output there)
39. `world.go` 150-195, 440-605
40. `internal/inputhandlers/echo.go` 1-25
41. `internal/inputhandlers/cleanser.go` 1-93
42. `internal/inputhandlers/login.go` 1-288
43. `internal/inputhandlers/login_prompt_handler.go` 200-435 (1-199 grep: no write)
44. `internal/inputhandlers/systemcommands.go` 100-187
45. `internal/inputhandlers/term_ansi.go` 95-155
46. `internal/inputhandlers/term_iac.go` 60-140
47. `modules/cleanup/cleanup.go` 75-155
48. `modules/follow/follow.go` 180-228, 430-520
49. `modules/gmcp/gmcp.go` 140-215, 530-608
50. `modules/gmcp/gmcp.Mudlet.go` grep of every `func`/`SendText` line, 340-363
51. `cmd/smoke/main.go` 1-203, `recorder.go` 1-41, `script.go` 1-305, `client.go` 1-251, `fight.go` 1-189 (copyover.go, admin.go by grep; server.go not read) (one entry: the smoke tool)
52. `cmd/smoke/testdata/pre-4b/bot0.txt` 1-399; `.gitattributes` 1-4; `internal/smoke/telnet/ansi.go` 1-16; `internal/plugins/plugincallbacks.go` 20-60; `internal/plugins/plugins.go` 538-560; `internal/term/term.go` 15-35; `internal/web/web.go` 465-520; `internal/usercommands/suicide.go` 80-115, `suicide_npc.go` 205-270, `usercommands.go` 1-90, `print.go` 1-30; `gomud_strip_plan.md` 113-126 (one entry: spot reads)

Skipped on purpose: `cmd/smoke/server.go`, `pre-5/*.txt` (diffed by script, not read line by line), the rest of `internal/web` (grep only), the rest of `modules/gmcp` (grep only).

Live evidence gathered (scratchpad, not repo): four `go run ./cmd/smoke` runs at HEAD (two with `-bots 2`, two with `-bots 1`, all PASS), transcripts diffed by `scratchpad/mask.py` with the spec's masks. Results in section 5.

---

## 0. Facts that shape the whole slice

| Fact | Source |
|---|---|
| Messaging delivery to a player is still a queued `events.Message`: `ConnectionController.Perceive` calls `account.SendText(p.Text)`, which queues `events.Message{UserId, Text: txt+"\n"}`; the hook `Message_SendMessage` does the socket write | `actors/controller.go:40-45`, `accounts/userrecord.go:334-341`, `hooks/Message_SendMessages.go:23-36` |
| So the one text write behind messaging lives in `internal/hooks`, not in `internal/messaging` | guard entry `messaging_test.go:24` says so |
| `messaging` imports, transitively: mudlog copyover term util configs fileloader gametime statmods buffs colorpatterns **connections** ids items stats **events** races skills spells **characters mobs** prompt **accounts** perception **actors** gamelock exit keywords mutators **rooms** | `go list -deps ./internal/messaging` |
| Hence `accounts`, `rooms`, `connections`, `events`, `characters`, `mobs`, `items`, `actors` **cannot call `messaging`** (import cycle). Sends that originate there (`Account.GrantXP`, room decay text, sign and exit pruning, connections' copyover text) need a seam (a hook or func var set by messaging, the `CanPerceive`/`OnBehaviourTick` pattern) or must move to a caller | same |
| `inputhandlers`, `web` do not import messaging today; `messaging` does not import them, so they can | `go list -deps` |
| `templates` (owner of `AnsiParse`, `templates.go:434`) imports accounts but not rooms/actors/messaging, so messaging may import templates | `go list` |
| Connection write functions: `connections.SendTo` (`connections.go:259`), `SendRawTo` (`:237`, **zero callers** anywhere), `Broadcast` (`:172`, skips `Login` state), `(*ConnectionDetails).Write` (`connectiondetails.go:263`: drops output when suppressed, `\n`->`\r\n`, ANSI strip for AI, websocket `WriteMessage`, refuses IAC on websocket), `WriteRaw` (`:250`), websocket ping `WriteControl` (`heartbeat.go:96`) | read |
| The current guard only sees selectors `connections.SendTo/SendRawTo/Broadcast`; it allows all of `internal/connections/`, `internal/messaging/`, `internal/perception/`; it does not see `.Write` on a `ConnectionDetails` or a raw `net.Conn` | `guards/messaging_test.go:17-63` |
| Output events and their only listeners: `events.Message` -> `Message_SendMessage`, `events.Broadcast` -> `Broadcast_SendToAll`, `events.RedrawPrompt` -> `RedrawPrompt_SendRedraw`, `events.WebClientCommand` -> `WebClientCommand_SendWebClientCommand` | `hooks/hooks.go:44,46,52,54` |
| Tests that observe output by listening to `events.Message`/`Broadcast`: `actors/controller_test.go:36`, `hooks/charmexpiry_test.go:45`, `hooks/combat_golden_test.go:111`, `usercommands/speech_test.go:19,132`, `world_test.go:157`. Removing the events breaks their capture point | grep `RegisterListener(events.(Message|Broadcast...` |
| `Account.SendWebClientCommand` has **no callers** | grep, `userrecord.go:343` only |
| `events.Broadcast.IsCommunication` and `SourceIsMod` are set (`usercommands/broadcast.go:30-31`, `rooms.go:269` for Message) but **read by nothing** | grep |

Send-site counts outside usercommands (non-test, `\.SendText\(|SendTextToExits\(|SendTextCommunication\(|usercommands\.Tell\(` per file): accounts/userrecord.go 2, actors/controller.go 1, hooks: ActorUpkeep 1, AutoHeal 2, DoCombat 6, InactivePlayers 1, RoomRoundTick 1, PlayerDespawn 1, PlayerDrop 2, PlayerSpawn 3, rooms/roommanager.go 3, rooms/rooms.go 3 calls (+2 definitions), modules/cleanup 7, modules/follow 18, modules/gmcp/gmcp.Mudlet.go 20, world.go 1. For scale: `.SendText(` is 631 sites repo-wide, 581 in usercommands; `Tell(` 227, 205 in usercommands; `ExcludeId(` 79, 77 in usercommands.

---

## 1. Exhaustive grep, by package

Grep (non-test, all of repo except `internal/usercommands/`, `internal/messaging/`, `cmd/smoke/`; 313 files): `\.SendText\(|SendTextCommunication|SendTextToExits|connections\.(SendTo|SendRawTo|Broadcast)\(|\.WriteRaw\(|\.Write\(|Fprint`, then `events\.(Message|Broadcast|WebClientCommand|RedrawPrompt|GMCPOut)\{`, `\bTell\(|ExcludeId\(`, and `SendText[A-Za-z]*\b` without a call. The grep found hits in accounts, rooms, hooks etc., so it could succeed; packages with no hit (below) are a real absence.

Kinds: **I** interface (to one actor/account), **R** room or observer, **N** named receivers, **W** world/broadcast, **P** protocol or negotiation bytes, **G** GMCP, **X** not game output.

### internal/hooks (the write layer and its callers)

| file:line | receivers | kind | what |
|---|---|---|---|
| Message_SendMessages.go:31 | `message.UserId`'s connection | I | writes `AnsiMoveCursorColumn+EraseLine+AnsiParse(text)` (screen-reader stripped), then queues `RedrawPrompt` (:33) |
| Message_SendMessages.go:80 | every **player** in `RoomId` minus UserId and ExcludeUserIds, quiet needs `superhearing` | R | same write per player, queues `RedrawPrompt` (:82). NPCs never receive |
| Broadcast_SendToAll.go:40,45,57,62 | `accounts.GetAllActiveUsers()` | W | screen-reader text when set (:27-52), plain otherwise; `SkipLineRefresh` omits the cursor/erase prefix; queues RedrawPrompt per user first (:33) |
| RedrawPrompt_SendRedraw.go:39 | one account | I | writes `AnsiParse(GetCommandPrompt())`; `OnlyIfChanged` compares with temp data `cmdprompt` (:24-35) |
| WebClientCommand_SendWebClientCommand.go:22 | one websocket conn | P | writes raw `cmd.Text` (e.g. `TEXTMASK:false`); cancels for non-websocket (:18) |
| Log_FollowLogs.go:56 | admin connections following logs at a level | I (admin) | tees `fmt.Sprintln(evt.Data[1:]...)`; lists set by `usercommands/admin.syslogs.go:33,39` |
| PlayerDespawn_HandleLeave.go:59 | room (`room.SendText(tplTxt)` "player-despawn" template), no exclude | R | leaving player's name |
| PlayerDespawn_HandleLeave.go:63 | leaving connection | I (no actor after) | "goodbye" template, `connections.SendTo`, then logout and `connections.Remove` (:65-68) |
| NewRound_AutoHeal.go:52 | user | I | "you are bleeding out!" |
| NewRound_AutoHeal.go:54 | room minus user | R | "<name> is bleeding out! Somebody needs to provide aid!" |
| NewRound_AutoHeal.go:72 | user | prompt | `RedrawPrompt{OnlyIfChanged:true}` |
| NewRound_InactivePlayers.go:47 | user | I | idle-kick warning |
| NewRound_RoomRoundTick.go:64 | room, no actor | R (world event) | room/zone idle message (random 5%/20%) |
| PlayerDrop_HandlePlayerDrop.go:30 / :37-39 | user / room minus user | I / R | "you drop to the ground!" / "<name> drops to the ground!" |
| PlayerSpawn_HandleJoin.go:68-69, :90 | user | I | plaintext-password warning (2 lines), portal reset line |
| NewRound_ActorUpkeep.go:146 | account | I | alignment shift |
| NewRound_DoCombat.go:135,149,176,438 | actor via `usercommands.Tell` | I | target lost, rage subsides, can't find target, no exit to flee |
| NewRound_DoCombat.go:344,353 | defender via `Tell` | I | the `***` frame lines around the shield-break `messaging.Send` (:345-352) |
| NewRound_DoCombat.go:168,189,345,421,444 | already `messaging.Send` | R | done in slice 5 |
| NewTurn_AutoSave.go:39,43,51,55,63,67 | all | W | "Saving users/rooms/other..." and "Done." (`SkipLineRefresh`) |

### internal/accounts

| file:line | receivers | kind | what |
|---|---|---|---|
| userrecord.go:181,187 | self (`u.SendText`) | I | XP gained line, called from `GrantXP` (cannot import messaging) |
| userrecord.go:334-341 | self | I | `SendText`: queues `events.Message{UserId, Text+"\n"}`. The funnel for every `.SendText` |
| userrecord.go:343-350 | self | P | `SendWebClientCommand`: queues `events.WebClientCommand`; **no callers** |
| userrecord.prompt.go:112 | (builder) | prompt | `GetCommandPrompt` returns `AnsiMoveCursorColumn+EraseLine+prompt+unsent text+TelnetGoAhead(if client wants GA)`: protocol bytes inside the prompt text |
| index.go:166..503, userrecord.go:496 | files, hashers | X | not output |

### internal/rooms

| file:line | receivers | kind | what |
|---|---|---|---|
| rooms.go:248,251 | room, no actor | R (world event) | corpse crumbles to dust (mob / user corpse colour) |
| rooms.go:534 | room, no actor | R (world event) | container crumbles (`Prepare`) |
| rooms.go:262-272 | `SendTextCommunication`: room, exclude ids, `IsCommunication:true` | R | definition; callers only in usercommands (8) |
| rooms.go:274-283 | `SendText`: room, exclude ids | R | definition; queues `events.Message{RoomId}` |
| rooms.go:285-317 | `SendTextToExits`: each neighbour room with a back exit, "(From <exit>) " prefix, quiet flag | R (far) | definition; callers only in usercommands (2) |
| roommanager.go:197,201 | each player in room, per sign; rune only to its owner | I per player | "A sign crumbles to dust." / "The rune you had enscribed here has faded away." (text carries its own `\n`, SendText adds another) |
| roommanager.go:216 | each player in room | I per player | "The <exit> vanishes." (own `\n` too) |

### internal/actors

| controller.go:44 | the actor's account | I/R | `ConnectionController.Perceive` -> `account.SendText(p.Text)`. Slice 7's real write point |

### internal/inputhandlers (pre-actor and protocol)

| file:line | receivers | kind | what |
|---|---|---|---|
| echo.go:13, :22 | own conn | P | echo typed bytes; CRLF on Enter |
| cleanser.go:31 | own conn | P | BS SP BS on backspace (:28 is a comment) |
| term_ansi.go:111-113, 145-147 | own conn | P | erase the input line on history up/down |
| term_iac.go:81-84 | own conn | P | telnet CHARSET reply " UTF-8" |
| term_iac.go:131 | own conn | P | NEW-ENVIRON SEND request |
| login.go:45 | the **other**, kicked connection | I (pre-actor) | "goodbye" template before `connections.Kick` |
| login.go:58-59, 65-66, 73-74, 82-83, 89-90, 106-107, 116-117, 129-130, 136-137, 145-146 | own conn | I (pre-actor) | login and creation errors (`language.T("Error.LoginFailedGeneric")`, "Nope. Bye!", `LoginUser` msg, "Invalid login.", "Okay, bye!", "that user already exists", `Error.UserCreationFailed`, `err.Error()`), each followed by CRLF |
| login_prompt_handler.go:235, 240 | own conn | P | per-byte mask char / echo (raw telnet only, `:212-215`) |
| login_prompt_handler.go:263, 266 | own websocket conn | P/I | mask chars or the submitted buffer, once, for scrollback |
| login_prompt_handler.go:275 | own conn (not Mudlet) | P | CRLF after Enter |
| login_prompt_handler.go:291-292, 302 | own conn | I (pre-actor) | validator error + CRLF; "Too many mistakes." then `Remove` |
| login_prompt_handler.go:332, 398 | own conn (Mudlet) | P | WONT ECHO / WILL ECHO around masked steps |
| login_prompt_handler.go:390 | own websocket | P | `TEXTMASK:true/false`, sent synchronously on purpose (:379-384) |
| login_prompt_handler.go:401 | own conn | I (pre-actor) | the step's prompt template |
| systemcommands.go:129 | own conn | I | "goodbye" on `/quit`, then `Kick` |
| systemcommands.go:172 | all | W | shutdown countdown `events.Broadcast` from a goroutine (:148-180) |

### repo root

| file:line | receivers | kind | what |
|---|---|---|---|
| main.go:376 | `connections.Broadcast` (all not in Login) | W | "\r\nCopyover complete.\r\n" (smoke waits for it, `cmd/smoke/copyover.go:136`, `script.go:240`) |
| main.go:390 | all | W | "goodbye" template as `events.Broadcast` at shutdown |
| main.go:567 | own conn | P | NEW-ENVIRON request (client detection) |
| main.go:637-689 | own conn | P | WILL SGA, WONT LINEMODE, DO NAWS, CHARSET request, SGA payload, WILL EOR, ANSI request-resolution |
| main.go:709, 965, 1326 | own conn (telnet, websocket, SSH) | I (pre-actor) | connect splash template |
| main.go:712 | own conn (AI port) | I (pre-actor) | "This port is for AI clients..." |
| main.go:725, 728 | own conn | P | baseline WONT/WILL ECHO by client type |
| main.go:860, 862 | own conn | I | AI-flag/port mismatch warnings, after `SendEnterWorld` |
| main.go:904-907 | own conn (AI) | I | "Command dropped, AI rate limit" |
| main.go:1011, 1017 | own websocket | I | `CopyoverReconnectUser` msg |
| main.go:1067-1070 | own websocket | P | `events.WebClientCommand{TEXTMASK:false}` after login |
| main.go:1162, 1164, 1211 | raw `net.Conn` before `connections.Add` | I (pre-connection) | "AI connection pool is full", "Server is full (%d connections)", SSH "Server is full" (the guard cannot see these) |
| world.go:171, 177-180 | all | W | "Reloading flat files..." / "Done." (`SkipLineRefresh`) on `/reload` |
| world.go:563, 592 | user's conn | prompt | prompt written straight to the connection (bypasses RedrawPrompt) after prompt answers |
| world.go:581 | user | I | "<cmd> not recognized. Type help for commands." |
| world.go:597 | (comment) | | |
| copyover.go, signals_*.go | none | | grep found no write |

### internal/connections (inside the guard's allowed prefix today)

| file:line | kind | what |
|---|---|---|
| connections.go:201, 244, 271 | transport | the Broadcast/SendRawTo/SendTo loops |
| connectiondetails.go:255-306 | transport | `WriteRaw`, `Write` |
| heartbeat.go:96 | P | websocket ping |
| fd_unix.go:32 | P | `RELOGTKN:<token>` to each websocket at copyover save |
| fd_unix.go:35, 38 | I | "\r\nServer is rebooting. Please reconnect.\r\n" to websockets (game text, inside connections) |

### modules

| file:line | receivers | kind | what |
|---|---|---|---|
| cleanup/cleanup.go:82, 97-98, 130, 143, 152 | actor (`usercommands.Tell`) | I | trash/bury replies |
| cleanup/cleanup.go:101-103, 144 | room minus `ExcludeId(actor)` (0 for an NPC, so an NPC's own line reaches nobody extra) | R | "<name> destroys <item>...", "<name> buries the <corpse>." with `NameColour(actor)` (no "Someone" handling; trash skips the room line when hidden, :100) |
| cleanup/cleanup.go:116 | account | I | `user.GrantXP` -> accounts XP line |
| follow/follow.go:202-203, 209, 221 | both followers/targets by account | N (two accounts) | follow expiry lines on `NewRound` |
| follow/follow.go:445 | each lost follower account | N | "You are no longer following X." |
| follow/follow.go:464-465 | target account + actor | N | stop following, both sides |
| follow/follow.go:500 + 498 | target account + actor | N | start following, both sides |
| follow/follow.go:382, 388, 429, 451, 474, 480, 510, 514, 520 | actor (`Tell`) | I | command replies |
| gmcp/gmcp.go:203 | own conn | P/G | IAC WILL GMCP request (on connect and on join) |
| gmcp/gmcp.go:551, 553, 573, 575, 600, 602 | user's conn | G | GMCP payload (telnet frame or `!!GMCP` web frame) from `GMCPOut` events |
| gmcp/gmcp.go:541-542 etc. | stdout | X | debug `fmt.Print` when `CONSOLE_GMCP_OUTPUT=1` |
| gmcp/gmcp.Mudlet.go:246, 248, 262, 280-289, 298, 302, 306, 312, 318, 321, 359 | the account (`user.SendText`) | I | **game text** from the `mudletui`/`checkclient` commands. Not GMCP payload: the GMCP exception should not cover these |

### Other packages: no output found

`internal/web` (no `events.`, `connections.`, `usercommands` reference; only `RunWithMUDLocked` matched, proving the grep ran), `internal/plugins` (the `NetConnection` interface exposes `Write`, `plugincallbacks.go:28`; its only user `gmcp.onNetConnect` `gmcp.go:146-160` never calls it), `internal/combat` (only `fmt.Fprintf(&sb)` into report strings for admin ranking pages, `armor_rank.go:241-250`, `mob_rank.go:210-219`, `weapon_rank.go:199-204`: X; `TestCombatPackageDoesNotSend` guards it), `internal/copyover` (binary encoder writes: X), `internal/util`, `internal/procedural` (hashers: X), `internal/term/term.go:27` (inside a `/* */` comment), mobs, characters, buffs, items, parties, spells, mutators, behaviour, events (types only), `cmd/generate`, `cmd/reset-admin-pw` (stderr/stdout of a CLI: X), `cmd/newid`.

---

## 2. The slice notes' items

**`// slice 7:` comments.** Grep `slice 7` over all `*.go`: the guard's own entries (`guards/messaging_test.go:11,15,22-36`) plus three in usercommands (inventory U owns the code, listed here because the notes name them):
- `usercommands/suicide.go:93-98`: the player-death line stays an `events.Broadcast{Text: msg}`; reasons given in the comment: "messaging adds a newline after the CRLF the text carries" (`msg` ends in `term.CRLFStr`, :88-90, and `Account.SendText` appends `\n`), Broadcast has screen-reader and refresh fields messaging lacks, and the "killed by" list is text, not a role. Note: this call sets **neither** `TextScreenReader` nor `SkipLineRefresh`; only `NewTurn_AutoSave.go:45,57,69` and `world.go:179` set `SkipLineRefresh`, and **nothing sets `TextScreenReader`** (grep). The only reader of both is `Broadcast_SendToAll.go:27-62`.
- `usercommands/suicide_npc.go:219-221`: party XP and per-member sends stay as they are. The loop is `:222-270+`: per member `user.GrantXP(xpSplit, "combat")` (-> `accounts/userrecord.go:181/187`), alignment line `user.SendText(updateTxt)` (:265). Item and gold drop lines also stay (name no actor).
- `usercommands/usercommands.go:35-41`: `Tell` writes `user.SendText(text)` when the actor has an account; "slice 7 points this at messaging". `ExcludeId` (:45-50) returns the account's UserId or 0.

No `// slice 7:` comment exists outside usercommands and guards (the grep covered every package).

**Party XP and per-member sends**: above. The XP text itself is produced inside `accounts` (`GrantXP`), which cannot import messaging. Callers of `GrantXP` outside usercommands: `modules/cleanup/cleanup.go:116`.

**`main.go` raw `net.Conn` writes**: `main.go:1162`, `:1164` (telnet listeners, AI and human pools, before `connections.Add` at :1172), `:1211` (SSH listener, before the handshake). No `ConnectionDetails` exists yet, so a messaging function for this case takes the `net.Conn`.

**Telnet and terminal negotiation**: `term_iac.go:81,131`, `term_ansi.go:111-113,145-147`, `echo.go:13,22`, `cleanser.go:31`, `login_prompt_handler.go:235,240,263,266,275,332,390,398`, `main.go:567,637-689,725,728,1067`, `WebClientCommand` hook, and the GA inside the prompt (`userrecord.prompt.go:77-79,112`). GMCP's enable request `gmcp.go:203` is negotiation inside the GMCP file.

**Cleanup and follow modules' send helpers**: they call `usercommands.Tell`, `usercommands.ExcludeId`, `usercommands.NameColour`, `room.SendText`, and `Account.SendText` directly (tables above). When `Tell`/`ExcludeId` move to messaging these modules follow; follow's two-account lines are `Named` audience candidates.

**Interface text before an actor exists**: login and creation (`inputhandlers/login.go`, `login_prompt_handler.go`), splash and AI port lines (`main.go:709,712,965,1326`), goodbye on kick/quit/despawn (`login.go:45`, `systemcommands.go:129`, `PlayerDespawn_HandleLeave.go:63`, which runs after the actor is gone), copyover reconnect (`main.go:1011,1017`; websocket reboot text `fd_unix.go:32-38`), server full (`main.go:1162-1211`). Character creation after login is commands (`start`, race, name) in usercommands, with an actor (the transcript shows `start` before the race prompt).

---

## 3. Web and admin paths (out of scope; for the guard's exclusions)

- Websocket game traffic goes through `connections`: `main.go:943` `connections.Add(nil, conn)` then `ConnectionDetails.Write` -> `wsConn.WriteMessage` (`connectiondetails.go:286-304`). The upgrade is `web/web.go:490-500`, which hands the conn to `main.HandleWebSocketConnection` and writes nothing itself. Pings: `heartbeat.go:96`.
- Admin web does not reach game connections: `internal/web` has no `events.`, `connections.`, or `usercommands` reference (grep). Admin handlers run under `RunWithMUDLocked` (`admin_routes.go:13-...`).
- Non-game server output to exclude: HTTP responses in `internal/web` (16 `w.Write`/`Fprint(w`/`http.Error`/`WriteHeader` hits; e.g. `web.go:855`, `web.go:911`), `https_mode.go:184` (stderr), `cmd/*` CLI stdout/stderr, `gmcp.go` debug prints, combat rank report builders, log output via `mudlog`.

## 4. Templates

Outside usercommands, `templates.Process` output is sent at: `PlayerDespawn_HandleLeave.go:58` (room, via `room.SendText`) and `:62-63` (`connections.SendTo(AnsiParse(...))`); `login.go:44-45`; `login_prompt_handler.go:222-226` (mask) and `:371-401` (prompt, `AnsiParse` then `SendTo`); `systemcommands.go:127-129` and `:151-172` (countdown, `fmt.Sprintf(tplTxt, i)` into a Broadcast, parsed by the hook); `main.go:385-391` (goodbye, parsed before the Broadcast, then parsed again by the hook), `:708-709`, `:964-965`, `:1325-1326`. Two patterns: (a) text with `<ansi>` markup handed to `SendText`/`room.SendText`/`Broadcast`, parsed by the hook (`AnsiParse` in `Message_SendMessages.go:27,75`, `Broadcast_SendToAll.go:23,28`); (b) `AnsiParse` at the call site then a raw `connections.SendTo`. Inside usercommands 31 files call `templates.Process` (count only; inventory U). Spec: formatting stays, only the sending moves, so messaging needs an entry that takes already-parsed bytes (pattern b) and one that takes markup (pattern a).

---

## 5. Parity tools

**`cmd/smoke`.** Launches a server on a temp copy of the empty world, an admin spawns one copper ring per bot (and an NPC that speaks, `admin.go:68-70`), bots sign up, create a character, get/drop/get, walk north/south, say, bot 0 fights a rat while bot 1 observes (`fight.go`), quit, relogin, suicide, walk out of the recovery room, copyover (Linux only), final quit; checks GMCP Char/Room/Comm families. Pass/fail is by substring waits (`client.go:151-216`); `checkObserverText` (`fight.go:139-158`) checks role text. `-transcript DIR` writes `bot<N>.txt`: server text with ANSI CSI stripped (`telnet/ansi.go:8`, regex `\x1b\[[0-9;?]*[ -/]*[@-~]`) and `\r` removed (`client.go:79`), plus "> line" for each line sent (`recorder.go:28-35`). No masking exists in the repo; nothing compares transcripts automatically.

Baselines: `pre-4b/bot0.txt` (399 lines, master `c4314b7f`, 1 bot, no fight), `pre-5/bot0.txt` (431), `pre-5/bot1.txt` (440). `git ls-files --eol`: all three are `i/lf w/crlf attr/`; `core.autocrlf=true`; `.gitattributes` pins only `*.go`, `go.mod`, `go.sum`. So on this machine the checked-out baselines are CRLF while a fresh transcript is LF: a byte compare fails on every line unless `\r` is stripped or `cmd/smoke/testdata/** text eol=lf` is added.

**Measured noise (HEAD `62cc22ac`, scratchpad runs).** Masks applied: run tag, `nameless-N`, game clock, HP/MP, the Online column, `>>> X leaves towards/enters from` lines, repeated prompts on one line. Masked diff line counts:
- two 1-bot runs at the same commit: **33** lines differ;
- two 2-bot runs, bot0: **46**, bot1: **64**;
- 1-bot run vs `pre-4b/bot0.txt`: 29; 2-bot run vs `pre-5`: bot0 22, bot1 40.

What still varies after the spec's masks: "Also here:" NPC lists (rats wander; `nameless-N` digit count also changes the `who` table width), NPC idle lines ("guard says...", "A craftsman gives the sign...", "rat wiggles its nose"), "rat notices you as you enter!", the whole combat block (rolls, round count, XP amount), blank lines and prompt placement around async lines (a line arrives before or after a prompt redraw), and with 2 bots the interleaving (who logs in or dies first, online-table row order). After masking, the remaining legitimate content (login screens, room text, inventory, say, death line, goodbye art) is stable.

What the transcript cannot see at all: ANSI colour (stripped), the cursor-column and erase-line prefix and so whether a line refreshed the prompt, `\r` vs `\n`, telnet bytes, GMCP payloads (counted only), screen-reader output, websocket and SSH paths, admin-only text, and any message no bot triggers (bleeding out, idle kick, sign crumbles, autosave, reload, shutdown, follow, cleanup, mudletui).

**`internal/hooks/combat_golden_test.go`.** Seeds `util.SeedRandForTest(seed)`, builds two test rooms, runs `DoCombat` per round, and records per player "what each player read" by **listening to `events.Message`** (:111, :191-235) and re-expanding RoomId messages the way `Message_SendMessage` does; text kept as markup before `AnsiParse`; multi-line texts recorded as a line count; plus state and events. Compared with `testdata/combat-golden/*.txt` (9 files), `\r\n` normalised (:453), rewritten with `-update`. It only covers the combat round. If slice 7 stops routing player text through `events.Message`, this test (and the five other tests listening to `events.Message`/`Broadcast`, section 0) record nothing and pass or fail for the wrong reason: the capture point must move with the write.

**Options for "text parity holds".**

| Option | Effort | Catches | Misses |
|---|---|---|---|
| A. Smoke transcript diff, spec masks plus the extra masks above (drop "Also here", NPC-origin lines, the fight block, collapse blank lines, sort table rows), 1 bot vs `pre-4b`, `\r` stripped | small: a Go test or script in `cmd/smoke`; baseline already exists | wording and order of login, room, inventory, say, death, quit text for one player | colour, cursor/erase prefixes, prompt redraw count, protocol bytes, all observer text, every message the script does not trigger; masks hide real NPC-line regressions |
| B. As A with a raw recorder mode (keep ANSI and control bytes, mask only clock/HP/tag) | small-medium | adds colour and cursor-control parity for the scripted path | still noisy around async lines; still only scripted messages |
| C. In-process golden for output, extending the combat golden pattern: seeded rand, fixed rooms, real accounts with a fake `net.Conn` per account registered through `connections.Add(conn, nil)` (`connections.go:47`), so capture is the **bytes on the wire** and works the same before and after conversion; fixtures drive each converted site (commands through `usercommands.TryCommand`, hooks by firing their events, login via the prompt handler with a fake conn, broadcasts) then `events.ProcessEvents()` | medium-large: harness plus one fixture per site group; record goldens on the pre-conversion tree first | exact bytes per receiver (colour, CRLF, erase-line, prompt redraws, screen-reader variant), observers, NPC-excluded receivers, non-scripted messages; deterministic | NPC-wander timing; cross-goroutine ordering in the live server; GMCP unless captured too |
| D. Re-point the existing combat golden to the same wire capture as C | small once C's capture exists | keeps slice 5's combat parity honest after the write moves | only combat |

Recommendation for the plan to weigh: C (with D) as the proof, A as a coarse smoke cross-check; record C's goldens on this branch before the first conversion commit, then prove the comparison catches a deliberate one-word change (spec 4, "Text parity"). Pin `cmd/smoke/testdata/** text eol=lf` (and the new golden dir) in `.gitattributes` either way.

---

## 6. The guard

Existing output guards: `TestOnlyMessagingWritesToConnections` (`guards/messaging_test.go:16-70`; selectors `connections.SendTo/SendRawTo/Broadcast`; `permanent`: `modules/gmcp/gmcp.go`; shrinking `connectionWriters`: 15 files, each must still write or the test fails; `inside`: `internal/connections/`, `internal/messaging/`, `internal/perception/`) and `TestCombatPackageDoesNotSend` (`combat_test.go:60-103`; bans `SendText`, `Tell`, `SendTextToExits`, `SendTextCommunication`, `messaging.Send`, `events.Message{}` inside `internal/combat/`, with a positive control). `goFiles` (`guards_test.go:18-45`) parses every non-test `.go` file outside `.git` and `_datafiles`, fails if fewer than 100.

Gaps today: `.Write`/`.WriteRaw` on a `ConnectionDetails` or `NetConnection` outside connections (none exist now, nothing stops one); raw `net.Conn.Write` (`main.go:1162-1211`); game text written inside `internal/connections` (`fd_unix.go:35,38`); text that reaches the socket through `Account.SendText`/`Room.SendText`/`events.Message`/`events.Broadcast` from any package (the guard only checks the final writer, which is the hook); `gmcp.Mudlet.go` text is not covered by the GMCP exception today only because the exception is by file and that is another file.

The final guard ("allowlist holds only the GMCP exception") would need:
1. Writers: `connections.SendTo`, `SendRawTo` (or delete it: zero callers), `Broadcast`, and any method call named `Write`, `WriteRaw`, `WriteMessage`, `WriteControl` whose receiver is not in `internal/connections/` or `internal/messaging/` (AST name match on the selector; a type-checked scan via `go/types` would be exact). Flag any `.Write(` in `main.go` and `world.go` specifically, since the only legitimate ones today are the raw `net.Conn` writes slice 7 moves.
2. Bypass routes: calls named `SendText`, `SendTextToExits`, `SendTextCommunication`, `SendWebClientCommand`, `Tell` outside messaging (or have the compiler enforce it by deleting them), and composite literals `events.Message{}`, `events.Broadcast{}`, `events.WebClientCommand{}`, `events.RedrawPrompt{}` outside messaging (if those events survive as messaging internals).
3. Allowed: `internal/messaging/**`; `internal/connections/**` restricted to the transport (`Write`/`WriteRaw` implementations, the SendTo/Broadcast loops, `heartbeat.go` ping), with `fd_unix.go`'s text moved or named; `modules/gmcp/gmcp.go` (payload and enable request) as the one permanent entry. If Calabe makes negotiation a second exception, its files (`term_iac.go`, `term_ansi.go`, `echo.go`, `cleanser.go`, the negotiation block of `main.go`, the mask/echo bytes in `login_prompt_handler.go`) cannot be file-level entries without also exempting their text writes; better a named messaging function (`messaging.Protocol(connId, bytes)`) that the guard allows by name, or split those writes into their own file.
4. A positive control (as in `combat_test.go:100-102`): e.g. assert the scan found `messaging.go` and at least one transport write in connections.
5. Exclusions: `cmd/smoke/**` and `internal/smoke/**` (clients), `internal/web/**` HTTP writers, `cmd/*` CLIs, hashers/encoders (`.Write` on `hash.Hash`, `*os.File`, `bytes.Buffer`, `strings.Builder`, gzip, copyover encoder). A name-only `Write` ban needs these by path; a `go/types` scan can ban by receiver type (`*connections.ConnectionDetails`, `net.Conn`, `plugins.NetConnection`, `*websocket.Conn`, `ssh.Channel`) and needs no exclusions.

---

## 7. Proposed task split (non-usercommands work)

Order matters: T0 first, T8 last, T1-T7 can run in parallel subagents after T0 (they touch disjoint files).

- **T0 Foundation (one agent, serial).** Move the socket write into `internal/messaging`: messaging owns delivery to a connection (the bodies of `Message_SendMessage`, `Broadcast_SendToAll`, `RedrawPrompt_SendRedraw`, `WebClientCommand_...`), keeping cursor/erase prefix, screen-reader stripping, `SkipLineRefresh`, prompt redraw and the queued ordering unless Calabe rules otherwise. Add the API the other tasks need: interface text to a connection id (pre-actor), to a raw `net.Conn` (server full), world broadcast with screen-reader text, prompt redraw, admin log tee, and (by ruling) protocol bytes. Add a seam for packages messaging imports (`accounts`, `rooms`, `connections`): a func var or `util.Hook` they call and messaging sets at init. Build the parity harness (option C) and the wire capture, record goldens on the unconverted tree, re-point `combat_golden_test.go` and the five tests that listen to `events.Message`/`Broadcast`.
- **T1 Login and pre-actor**: `inputhandlers/login.go`, `login_prompt_handler.go` (text writes), `systemcommands.go:129`, `main.go:709,712,860,862,904,965,1011,1017,1162,1164,1211,1326`, `PlayerDespawn_HandleLeave.go:63`, `connections/fd_unix.go:32-38`.
- **T2 Protocol and negotiation** (per ruling Q1): `echo.go`, `cleanser.go`, `term_ansi.go`, `term_iac.go`, `login_prompt_handler.go` echo/mask/ECHO/TEXTMASK bytes, `main.go:567,637-689,725,728,1067`, GA in the prompt.
- **T3 World events from rooms**: `rooms.go:248,251,534`, `roommanager.go:197,201,216`, `NewRound_RoomRoundTick.go:64` (messages with `Room` set and no actor; senses per ruling Q3); delete `Room.SendText`, `SendTextCommunication`, `SendTextToExits` once inventory U's callers are gone.
- **T4 Account and actor text in hooks and world**: `accounts.GrantXP` (via seam), `NewRound_AutoHeal.go:52,54`, `InactivePlayers.go:47`, `ActorUpkeep.go:146`, `PlayerDrop.go:30,37`, `PlayerSpawn.go:68,69,90`, `PlayerDespawn.go:59`, `DoCombat.go` six `Tell`s, `world.go:563,581,592`.
- **T5 Broadcasts**: `NewTurn_AutoSave.go` (6), `world.go:171,177`, `systemcommands.go:172`, `main.go:376,390`, and the death line `usercommands/suicide.go:96` (coordinate with inventory U; ruling Q6).
- **T6 Admin log tee**: `Log_FollowLogs.go:56` (ruling Q7).
- **T7 Modules**: `modules/cleanup` (7), `modules/follow` (18, Named audience for two-account lines), `modules/gmcp/gmcp.Mudlet.go` (20 `SendText`).
- **T8 Delete and guard (after U's tasks)**: delete `Account.SendText`, `SendWebClientCommand`, `connections.SendRawTo`, the four output hooks and (if unused) `events.Message`/`Broadcast`/`WebClientCommand`/`RedrawPrompt`; `ConnectionController.Perceive` writes through messaging's connection path; final guard (section 6); parity proof (deliberate change caught, then reverted); smoke; Done-when check against `gomud_strip_plan.md` 4b row ("One command set and one combat path serve players and NPCs; smoke passes, including a fight against an NPC").

---

## Open questions (decisions for Calabe)

1. **Telnet and terminal negotiation: second named exception, or through messaging?** Player effect: none either way (same bytes). Exception keeps echo/backspace/IAC code where it is; routing through a `messaging` protocol function means the guard needs no second entry and one place owns every socket byte.
2. **Keep socket writes queued, or write synchronously from messaging?** Today every text line is queued (`events.Message`) and written by a hook later in the same event pass; prompts are queued at priority 100. Queued keeps today's line order exactly. Synchronous is simpler but can move a line before or after a prompt redraw or a broadcast that was queued earlier: players could see lines and prompts in a different order.
3. **Senses for room world events with no actor** (corpse and container crumble, idle messages, sign crumbles, temp exit vanishes, bleeding out, drop to the ground). As Sight only, a player in a dark room stops reading "A rat corpse crumbles to dust." (spec allows sight-only loss in darkness as an intended difference, and none exist in `empty`); as Sight|Sound they read as today. NPCs start receiving these as perceptions either way (no visible effect today).
4. **Parity proof**: option C (in-process wire-byte goldens, medium-large effort, catches colour, prompts, observers and unscripted messages) or option A (masked smoke transcript, small effort, misses all of that and needs extra masks because two runs of the same commit already differ by 33 masked lines). Player effect: none; it decides what regressions slice 7 can miss.
5. **Copyover text inside `internal/connections`** (`fd_unix.go:35,38` "Server is rebooting. Please reconnect." and `RELOGTKN`): move behind messaging, or leave as transport. Player effect: web-client players at a copyover; none if moved verbatim.
6. **The player-death line** (`suicide.go:96`): stay an `events.Broadcast` (then it is a permanent exception and the slice note's reason stands) or become a `World` interface message. Player effect: through today's `SendText` path it would gain a blank line after the CRLF it carries unless the CRLF is dropped; `TextScreenReader` and `SkipLineRefresh` are not set on this call, so screen-reader players see no difference.
7. **Admin log tee** (`Log_FollowLogs.go:56`, admin `syslogs`): move to a messaging function or allow as an admin-only exception. Player effect: admins following logs only; none if moved verbatim.
