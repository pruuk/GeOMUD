# Phase 4b slice 3: messaging core (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** one package delivers every world event and interface message to the actors who can
perceive it, in the form their controller takes (text for a connection, a structured perception for
an NPC behaviour). This slice builds and tests that package; slices 5 and 7 move today's output onto
it.

**Architecture:** a leaf package `internal/perception` holds the vocabulary (`Message`, `Perception`,
kinds, roles, senses, ranges, audiences), so `actors.Controller` can name it. `internal/messaging`
resolves receivers from range and audience, runs the `CanPerceive` check per receiver, picks the
role, renders the text with tokens resolved for that receiver, and calls `Controller.Perceive`. The
connection controller sends the text through the account's existing output queue; the behaviour
controller fires a hook with the structured perception. A guard test fences direct connection
writes to an allowlist that may only shrink.

**Tech stack:** Go, testify. No new dependency.

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md`, decisions 5-8, section 2, section 4
slice 3. Branch: `strip/phase-4b-3-messaging-core` off `master` once PR #22 (slice 2c) has merged;
if it has not, off `strip/phase-4b-2c-item-ids`, rebased when #22 merges. The two branches share no
files (this plan puts its guard in a new file).

---

## Facts verified against source (2026-10-28, master `82b5f53b`)

| Fact | Value | Where |
|---|---|---|
| Controller today | `actors.Controller` has one method, `Account() *accounts.Account`; its comment says slice 3 adds `Perceive`. `ConnectionController{account}`, `BehaviourController{mob}` | `internal/actors/controller.go:12-40` |
| Actor | `Actor{Id ActorId, Character *characters.Character, controller}`; `ActorId = ids.ID` (the body's ID). `actors.Get(id)`, `actors.All()` (sorted by ID string), `actors.Spawn(char, ctrl)` (refuses a zero ID), `actors.Reset()` for tests | `actor.go`, `registry.go:22-103, 230-236` |
| Import graph | `actors` imports `accounts characters ids mobs mudlog`; `rooms` imports `actors` (and `accounts characters events exit gametime ids mobs mutators ...`). So `messaging` (needs actors and rooms) cannot be imported by `actors`, and the type `Perceive` takes must live in a package below `actors` | `go list -f '{{.Imports}}'` |
| How text reaches a player today | `Account.SendText(txt)` queues `events.Message{UserId, Text: txt+"\n"}` (`userrecord.go:334-340`); `Room.SendText` queues `events.Message{RoomId, ExcludeUserIds}` (`rooms.go:273-282`). The hook `Message_SendMessage` (`hooks/Message_SendMessages.go`) ANSI-parses, strips for screen readers, writes with `connections.SendTo` and queues `RedrawPrompt`. For room messages it skips `IsQuiet` text unless the player has buff flag `superhearing` (`:68-73`). Mobs receive no room text | read |
| Communication today | `say`: a hidden speaker reads `someone says, ...` to the room, else `<ansi fg="username">Name</ansi> says, ...`; the speaker gets `You say, ...`; adjacent rooms get the quiet `You hear someone talking.` (`usercommands/say.go:20-36`). `shout`: adjacent rooms through `Exits`, `ExitsTemp` and active mutator exits get `Someone shouts from the <ansi fg="exit">dir</ansi> direction, ...` where `dir = otherRoom.FindExitTo(room.RoomId)`, skipped when empty (`shout.go:30-59`) | read |
| Names in text | Today's room text writes the plain name: `<ansi fg="username">%s</ansi>` for players, `mobname` for NPCs. `Character.GetPlayerName/GetMobName(viewer)` add adjectives and colour suffixes and are used for listings, not sentences | `say.go:31`, `character.go:589-595, 657-690` |
| Hidden | Buff flag `hidden` (sneaking). Commands check it to write `someone` (`say.go:20,29`, 9 mob commands) | `grep '"hidden"'` |
| Light | `Room.GetVisibility()` returns 0-2 from night, biome, mutator `LightMod` and light carriers (`rooms.go:157-205`). `look` refuses when it is below 1 unless the viewer has buff flag `nightvision` (`look.go:25-31`) | read |
| Room bodies and neighbours | `Room.GetBodies() []ids.ID` (arrival order, `rooms.go:802`); `Room.Exits map[string]exit.RoomExit`, `ExitsTemp map[string]exit.TemporaryRoomExit` (both have `RoomId`), `Room.ActiveMutators` is a range-over-func iterator of `mutators.Mutator` whose `GetSpec().Exits` is `map[string]exit.RoomExit`; `Room.FindExitTo(roomId) string` checks all three (`rooms.go:1580-1603`). `Room.Zone` is exported. `rooms.IsRoomLoaded(id)`, `rooms.LoadRoom(id)` (loads from disk when not in memory) | read |
| Putting rooms in memory | `addRoomToMemory` is unexported (`roommanager.go:557`); `room.bodies` is unexported. Other packages' tests have no way in today. Precedent for exported test helpers in a non-test file: `accounts/testhelpers.go`, `mobs/testhelpers.go` | read |
| Hooks | `util.Hook[T]` with `Register(func(T) T)` and `Fire(T) T` | `internal/util/hook.go:26-50` |
| Direct connection writes | `connections.SendTo` is called outside `internal/connections` in 16 non-test files: `hooks/Broadcast_SendToAll.go`, `hooks/Log_FollowLogs.go`, `hooks/Message_SendMessages.go`, `hooks/PlayerDespawn_HandleLeave.go`, `hooks/RedrawPrompt_SendRedraw.go`, `hooks/WebClientCommand_SendWebClientCommand.go`, `inputhandlers/cleanser.go`, `echo.go`, `login.go`, `login_prompt_handler.go`, `systemcommands.go`, `term_ansi.go`, `term_iac.go`, `main.go`, `modules/gmcp/gmcp.go`, `world.go`. `connections.Broadcast` only in `main.go`; `connections.SendRawTo` nowhere. `main.go:1160,1162,1209` also write "server full" notices straight to a `net.Conn` before a connection is registered | `grep -l` |
| Guard tests | `internal/guards/guards_test.go` parses every non-test `.go` file (`goFiles(t)`, keys are slash paths from the repo root) | read |
| Test baseline | 35 packages ok on master; slice 2c (PR #22) makes it 34. 3 known Windows failures | `geomud-verifying` |

## Decisions this plan applies

From the spec (Calabe, 2026-10-27): one messaging package for all output (decision 5); ranges room,
nearby, zone, world (decision 6); presentation, GMCP and comms channels are separate (decision 7);
senses stay minimal (decision 8); section 2's message shape, delivery, `CanPerceive` hook, interface
text and guard.

Defaults set by the controller (recorded so they are not re-asked; Calabe reviews them with this
plan):

1. **Two packages.** `internal/perception` is the vocabulary and imports only `ids`;
   `internal/messaging` does delivery. `actors.Controller.Perceive` takes a `perception.Perception`,
   which a type in `messaging` could not be (import cycle, facts). Both count as "messaging" for the
   guard; neither writes to a connection in this slice.
2. **`messaging` renders; controllers deliver.** `messaging` knows the receiver, the light and the
   names, so it resolves tokens and puts the receiver's text on the perception. The connection
   controller sends that text with `Account.SendText`, so it joins the same output queue as every
   other line today and ordering is unchanged; the socket write stays in `Message_SendMessage` until
   slice 7. The behaviour controller ignores the text and fires `actors.OnBehaviourPerceive` with the
   structured perception (slice 6 attaches behaviour to it).
3. **Tokens.** `{actor}` and `{target}` become the plain coloured name used in room text today
   (`<ansi fg="username">Name</ansi>` for an actor with an account, `mobname` otherwise), or
   `someone` when the receiver cannot see that body; `Someone` when the token opens the text.
   `{direction}` becomes `<ansi fg="exit">dir</ansi>`. Names are plain (no adjectives) so slices 5
   and 7 can match today's text.
4. **Who is unseen.** A receiver sees a body unless the body is `hidden` (today's sneaking rule), the
   receiver's sight is blocked in its own room, or the receiver is outside the source rooms (a far
   receiver). Every body sees itself.
5. **`CanPerceive` in 4b.** Sight: the receiver's room has visibility 1 or more, or the receiver has
   `nightvision` (look's rule). Sound: passes, except a `Faint` message needs `superhearing`
   (today's `IsQuiet`). Touch and Smell pass. It is a replaceable function variable
   (`messaging.CanPerceive`) that Phase 5's ruleset replaces. A receiver perceives a message when any
   of its senses passes; senses are tried in the order Sight, Sound, Touch, Smell, and the first that
   passes is `SenseUsed`.
6. **Roles and text.** The actor always receives `ActorText`, with no sense check. The target
   receives `TargetText`, or `ObserverText` when that is empty. A far receiver gets `FarText`, or
   `ObserverText` when that is empty. Every receiver who perceives the message gets `Perceive`; the
   connection controller drops a perception whose text is empty.
7. **Ranges.** Source rooms are the actor's room and, when the target is in another room, the
   target's room (spec section 2); an actorless world event names its room in `Message.Room`.
   `Room`: bodies in the source rooms. `Nearby`: plus the bodies of each loaded room a source room
   reaches through `Exits`, `ExitsTemp` or active mutator exits, when that room has an exit back
   (`FindExitTo`, as `shout` does); `{direction}` is that exit. `Zone`: every live actor whose loaded
   room has a source room's zone. `World`: every live actor. Rooms are never loaded to deliver a
   message (an unloaded room has no live bodies). Receivers are deduplicated, source rooms first.
8. **Audiences.** `InRange`: everyone the range reaches. `ActorOnly`: the actor alone (interface
   text, which also skips the sense check and goes to the actor alone whatever its range).
   `Named`: the actor plus the named bodies, when the range reaches them.
9. **Interface text before an actor exists** (login screens) is not in this slice; slice 7 adds the
   connection-level function with the login conversion.
10. **Guard.** A new test file fences `connections.SendTo`, `SendRawTo` and `Broadcast` to
    `internal/connections`, `internal/messaging`, `internal/perception`, the permanent GMCP exception
    (`modules/gmcp/gmcp.go`), and an allowlist of today's other 15 files. A listed file that no longer
    calls them fails the test, so the list only shrinks. `main.go`'s three raw `net.Conn` writes are
    named in the test's comment and in the spec's notes for slice 7; the guard does not see them.
11. **No production caller yet.** Nothing calls `messaging.Send` outside tests in this slice (slice 5
    moves combat onto it, slice 7 the rest). Smoke is unchanged and must pass.

## File structure

| File | Change |
|---|---|
| `internal/perception/perception.go` (new) | the vocabulary |
| `internal/perception/perception_test.go` (new) | sense set test |
| `internal/actors/controller.go` | `Perceive` on the interface and both controllers; `OnBehaviourPerceive` |
| `internal/actors/controller_test.go` (new) | both controllers' `Perceive` |
| `internal/rooms/testhelpers.go` (new) | `AddTestRoom`, `RemoveTestRoom`, `(*Room).AddTestBody` |
| `internal/messaging/messaging.go` (new) | `Send`, receivers, roles |
| `internal/messaging/perceive.go` (new) | `CanPerceive`, `canSee`, the flag and light seams |
| `internal/messaging/render.go` (new) | token rendering |
| `internal/messaging/messaging_test.go`, `render_test.go` (new) | the slice's Done-when tests |
| `internal/guards/messaging_test.go` (new) | connection-write guard |
| `PATCH_NOTES.md`, `status.md`, spec, `geomud-verifying` | Task 7 |

---

### Task 0: Branch

- [ ] **Step 1:** If PR #22 has merged: `git switch master && git pull --ff-only && git switch -c strip/phase-4b-3-messaging-core`.
  Otherwise: `git switch strip/phase-4b-2c-item-ids && git switch -c strip/phase-4b-3-messaging-core`.
  Record which in the Task 7 notes.

---

### Task 1: `internal/perception`

**Files:** Create `internal/perception/perception.go`, `internal/perception/perception_test.go`.

- [ ] **Step 1: Write the failing test** `perception_test.go`:

```go
package perception

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSenseSet(t *testing.T) {
	s := Sight | Sound
	assert.True(t, s.Has(Sight))
	assert.True(t, s.Has(Sound))
	assert.False(t, s.Has(Touch))
	assert.Equal(t, []Sense{Sight, Sound}, s.Each())
	assert.Equal(t, []Sense{Sight, Sound, Touch, Smell}, (Smell | Touch | Sound | Sight).Each(), "always in sense order")
	assert.Empty(t, Sense(0).Each())
}
```

- [ ] **Step 2:** Run `go test ./internal/perception/ -count=1`. Expected: FAIL (package has no
  non-test files / undefined names).

- [ ] **Step 3: Write `perception.go`:**

```go
// Package perception is the vocabulary of Phase 4b messaging (spec section
// 2): what a message is, and what reaches a receiver. It imports nothing of
// the world, so actors.Controller can name a Perception while
// internal/messaging, which delivers them, imports actors.
package perception

import "github.com/GoMudEngine/GoMud/internal/ids"

// Kind says whether a message is something happening in the world or text
// for the actor alone (help, score, prompts, errors).
type Kind uint8

const (
	WorldEvent Kind = iota // perceived by whoever can perceive it
	Interface              // for the actor only; no sense check; behaviours drop it
)

// Sense is a set of senses that can carry a message.
type Sense uint8

const (
	Sight Sense = 1 << iota
	Sound
	Touch
	Smell
)

// senseOrder is the order senses are tried in, and SenseUsed picks the first.
var senseOrder = []Sense{Sight, Sound, Touch, Smell}

// Has reports whether s includes every sense in o.
func (s Sense) Has(o Sense) bool { return s&o == o }

// Each lists the senses in s in sense order.
func (s Sense) Each() []Sense {
	out := []Sense{}
	for _, one := range senseOrder {
		if s.Has(one) {
			out = append(out, one)
		}
	}
	return out
}

// Range is how far a message carries.
type Range uint8

const (
	Room   Range = iota // the source rooms
	Nearby              // the source rooms and the rooms next to them
	Zone                // every room in the source rooms' zones
	World               // everywhere
)

// Audience says who in range receives a message.
type Audience uint8

const (
	InRange   Audience = iota // everyone in range
	ActorOnly                 // the actor alone
	Named                     // the actor and the bodies in Message.Named
)

// Message is one thing to be perceived. Text fields may hold the tokens
// {actor}, {target} and {direction}, resolved per receiver by messaging.
type Message struct {
	Kind     Kind
	Actor    ids.ID // the body doing it; zero for a world event (weather, a door creaking)
	Target   ids.ID // optional
	Room     int    // the source room of a world event with no actor
	Senses   Sense  // which senses can carry it
	Faint    bool   // needs keen senses (today's IsQuiet)
	Range    Range
	Audience Audience
	Named    []ids.ID // for Audience Named

	ActorText    string // what the actor reads
	TargetText   string // what the target reads; ObserverText when empty
	ObserverText string // what everyone else in the source rooms reads
	FarText      string // what receivers outside the source rooms read; ObserverText when empty
}

// Role is how a receiver stands to a message.
type Role uint8

const (
	Actor Role = iota
	Target
	Observer
)

// Perception is what reaches one receiver: the message, how it reached
// them, and the text for them with tokens resolved. A connection controller
// writes Text; a behaviour reads the rest.
type Perception struct {
	Role       Role
	Message    Message
	SenseUsed  Sense  // zero for the actor's own text and interface text
	ActorSeen  bool   // the receiver could see the actor (true when there is none)
	TargetSeen bool   // the receiver could see the target (true when there is none)
	Far        bool   // the receiver is outside the source rooms
	Direction  string // for a far receiver next to a source room: the exit back toward it
	Text       string // the receiver's text; empty when there is none for its role
}
```

- [ ] **Step 4:** Run the test. Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/perception/perception.go internal/perception/perception_test.go
git commit -F - <<'EOF'
perception: the messaging vocabulary (Phase 4b slice 3)

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 2: Controllers perceive

**Files:** Modify `internal/actors/controller.go`; create `internal/actors/controller_test.go`.

- [ ] **Step 1: Write the failing test** `controller_test.go`:

```go
package actors

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBehaviourControllerFiresTheStructuredPerception(t *testing.T) {
	m := bodyMob()
	var got []BehaviourPerception
	OnBehaviourPerceive.Register(func(bp BehaviourPerception) BehaviourPerception {
		if bp.Mob == m {
			got = append(got, bp)
		}
		return bp
	})

	p := perception.Perception{Role: perception.Observer, SenseUsed: perception.Sound,
		Message: perception.Message{Kind: perception.WorldEvent, Senses: perception.Sound}}
	NewBehaviourController(m).Perceive(p)
	require.Len(t, got, 1)
	assert.Equal(t, p, got[0].Perception)

	NewBehaviourController(m).Perceive(perception.Perception{Message: perception.Message{Kind: perception.Interface}})
	assert.Len(t, got, 1, "a behaviour drops interface text")
}

func TestConnectionControllerIgnoresEmptyText(t *testing.T) {
	acc := &accounts.Account{UserId: 9}
	assert.NotPanics(t, func() {
		NewConnectionController(acc).Perceive(perception.Perception{})
	})
}
```

  (`bodyMob` is in `registry_test.go`.) A test that the connection controller's text reaches the
  output queue belongs to the messaging tests, which use a recording controller; `Account.SendText`
  only queues an event.

- [ ] **Step 2:** Run `go test ./internal/actors/ -count=1`. Expected: FAIL (`Perceive` undefined).

- [ ] **Step 3: Implement** in `controller.go`. Imports: add
  `"github.com/GoMudEngine/GoMud/internal/perception"` and `"github.com/GoMudEngine/GoMud/internal/util"`.
  Replace the interface and its comment:

```go
// Controller is what drives an actor: a connection with its account, or an
// NPC behaviour. Code outside this package asks the controller questions and
// never asserts its concrete type (internal/guards enforces this). Slice 6
// adds Tick.
type Controller interface {
	// Account is the account driving the actor, or nil when none does.
	Account() *accounts.Account
	// Perceive is called by internal/messaging for each message that
	// reaches the actor.
	Perceive(p perception.Perception)
}
```

  After `ConnectionController.Account`:

```go
// Perceive writes the receiver's text to the account's output, in order
// with everything else it is sent. A perception with no text for this
// receiver writes nothing.
func (c *ConnectionController) Perceive(p perception.Perception) {
	if p.Text == "" {
		return
	}
	c.account.SendText(p.Text)
}
```

  After `BehaviourController.Account`:

```go
// BehaviourPerception is one perception reaching an NPC.
type BehaviourPerception struct {
	Mob        *mobs.Mob
	Perception perception.Perception
}

// OnBehaviourPerceive fires for each world event an NPC perceives. Slice 6
// attaches behaviour to it.
var OnBehaviourPerceive util.Hook[BehaviourPerception]

// Perceive hands a world event to the NPC's behaviour as a structured
// perception. Interface text is not for behaviours and is dropped.
func (b *BehaviourController) Perceive(p perception.Perception) {
	if p.Message.Kind == perception.Interface {
		return
	}
	OnBehaviourPerceive.Fire(BehaviourPerception{Mob: b.mob, Perception: p})
}
```

  Check `go build ./...`: any other type that implements `actors.Controller` (grep
  `Account() \*accounts.Account` across the repo, tests included) now needs `Perceive`; add a no-op
  to test fakes and report each.

- [ ] **Step 4:** Run `go build ./... && go test ./internal/actors/ ./internal/guards/ -count=1`.
  Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/actors/controller.go internal/actors/controller_test.go
git commit -F - <<'EOF'
actors: controllers perceive (text to the account, structured to a behaviour hook)

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

  (add any test fake files Step 3 touched to the `git add`).

---

### Task 3: Rooms test helpers

**Files:** Create `internal/rooms/testhelpers.go`.

- [ ] **Step 1: Write it:**

```go
package rooms

import "github.com/GoMudEngine/GoMud/internal/ids"

// AddTestRoom puts r in memory as LoadRoom finds it, for other packages'
// tests. RemoveTestRoom takes it out.
func AddTestRoom(r *Room) { addRoomToMemory(r, true) }

// RemoveTestRoom takes a test room out of memory.
func RemoveTestRoom(roomId int) { delete(roomManager.rooms, roomId) }

// AddTestBody puts a body in r without the move and its events, for other
// packages' tests.
func (r *Room) AddTestBody(id ids.ID) { r.bodies = append(r.bodies, id) }
```

  Read `addRoomToMemory` (`roommanager.go:557`) and `ClearRoomCache` first: if the forced overwrite
  does more than replace the memory entry (saves, unloads NPCs, fires events), use the unforced form
  and have `RemoveTestRoom` clear what `addRoomToMemory` set (`roomManager.rooms`, and the
  `roomIdToFileCache` entry if it was new). Say which you did.

- [ ] **Step 2:** `go build ./... && go vet ./internal/rooms/`. Expected: clean. (The helpers are
  exercised by Task 4's tests.)

- [ ] **Step 3: Commit**

```bash
git add internal/rooms/testhelpers.go
git commit -F - <<'EOF'
rooms: exported test helpers to put rooms and bodies in memory

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 4: `messaging.Send`

**Files:** Create `internal/messaging/messaging.go`, `perceive.go`, `render.go`,
`messaging_test.go`, `render_test.go`.

- [ ] **Step 1: Write the test scaffolding and the failing tests** in `messaging_test.go`:

```go
package messaging

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recorder is a controller that keeps what it perceives. With an account it
// stands for a player (names render as username), without one for an NPC.
type recorder struct {
	acc *accounts.Account
	got []perception.Perception
}

func (r *recorder) Account() *accounts.Account          { return r.acc }
func (r *recorder) Perceive(p perception.Perception) { r.got = append(r.got, p) }

// world builds rooms in memory and bodies in them for one test.
type world struct {
	t     *testing.T
	rooms []int
	flags map[ids.ID]map[string]bool
	light map[int]int
}

func newWorld(t *testing.T) *world {
	t.Helper()
	actors.Reset()
	w := &world{t: t, flags: map[ids.ID]map[string]bool{}, light: map[int]int{}}
	oldFlag, oldLight := hasFlag, roomVisibility
	hasFlag = func(c *characters.Character, flag string) bool { return w.flags[c.Id][flag] }
	roomVisibility = func(r *rooms.Room) int {
		if v, ok := w.light[r.RoomId]; ok {
			return v
		}
		return 2
	}
	t.Cleanup(func() {
		hasFlag, roomVisibility = oldFlag, oldLight
		for _, id := range w.rooms {
			rooms.RemoveTestRoom(id)
		}
		actors.Reset()
	})
	return w
}

// room adds a room in zone; exits maps an exit name to the room it leads to.
func (w *world) room(id int, zone string, exits map[string]int) *rooms.Room {
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = id, zone
	r.Exits = map[string]exit.RoomExit{}
	for name, to := range exits {
		r.Exits[name] = exit.RoomExit{RoomId: to}
	}
	rooms.AddTestRoom(r)
	w.rooms = append(w.rooms, id)
	return r
}

// body puts a named body in r. player says whether an account drives it.
func (w *world) body(r *rooms.Room, name string, player bool) (ids.ID, *recorder) {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = name
	c.RoomId = r.RoomId
	rec := &recorder{}
	if player {
		rec.acc = &accounts.Account{UserId: len(w.flags) + 100, Character: c}
	}
	require.NotNil(w.t, actors.Spawn(c, rec))
	r.AddTestBody(c.Id)
	w.flags[c.Id] = map[string]bool{}
	return c.Id, rec
}

func texts(r *recorder) []string {
	out := []string{}
	for _, p := range r.got {
		out = append(out, p.Text)
	}
	return out
}

func say(actor ids.ID) perception.Message {
	return perception.Message{
		Actor: actor, Senses: perception.Sound, Range: perception.Room,
		ActorText:    `You say, "hi"`,
		ObserverText: `{actor} says, "hi"`,
	}
}

func TestRoomRangeRolesAndNames(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, fredRec := w.body(square, "Fred", true)
	steve, steveRec := w.body(square, "Steve", true)
	_, ratRec := w.body(square, "rat", false)

	Send(perception.Message{
		Actor: fred, Target: steve, Senses: perception.Sight | perception.Sound, Range: perception.Room,
		ActorText:    `You attack {target}!`,
		TargetText:   `{actor} attacks you!`,
		ObserverText: `{actor} attacks {target}!`,
	})

	assert.Equal(t, []string{`You attack <ansi fg="username">Steve</ansi>!`}, texts(fredRec))
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> attacks you!`}, texts(steveRec))
	require.Len(t, ratRec.got, 1)
	assert.Equal(t, perception.Observer, ratRec.got[0].Role)
	assert.Equal(t, perception.Sight, ratRec.got[0].SenseUsed, "sight is tried first")
	assert.Equal(t, perception.Actor, fredRec.got[0].Role)
	assert.Equal(t, perception.Target, steveRec.got[0].Role)
}

func TestNPCNamesRenderAsMobname(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	rat, _ := w.body(square, "rat", false)
	_, fredRec := w.body(square, "Fred", true)

	Send(say(rat))
	assert.Equal(t, []string{`<ansi fg="mobname">rat</ansi> says, "hi"`}, texts(fredRec))
}

func TestTargetFallsBackToObserverText(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, _ := w.body(square, "Fred", true)
	steve, steveRec := w.body(square, "Steve", true)

	m := say(fred)
	m.Target = steve
	Send(m)
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> says, "hi"`}, texts(steveRec))
}

func TestDarkRoomBlocksSight(t *testing.T) {
	w := newWorld(t)
	cave := w.room(1, "caves", nil)
	w.light[1] = 0
	fred, _ := w.body(cave, "Fred", true)
	_, steveRec := w.body(cave, "Steve", true)
	owl, owlRec := w.body(cave, "Owl", true)
	w.flags[owl]["nightvision"] = true

	Send(perception.Message{Actor: fred, Senses: perception.Sight, Range: perception.Room,
		ActorText: `You wave.`, ObserverText: `{actor} waves.`})

	assert.Empty(t, steveRec.got, "a sight-only message does not reach a receiver in the dark")
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> waves.`}, texts(owlRec), "nightvision sees")
}

func TestSomeoneWhenTheActorIsUnseen(t *testing.T) {
	w := newWorld(t)
	cave := w.room(1, "caves", nil)
	w.light[1] = 0
	fred, fredRec := w.body(cave, "Fred", true)
	_, steveRec := w.body(cave, "Steve", true)

	Send(say(fred)) // heard in the dark, speaker unseen
	assert.Equal(t, []string{`Someone says, "hi"`}, texts(steveRec))
	assert.False(t, steveRec.got[0].ActorSeen)
	assert.Equal(t, perception.Sound, steveRec.got[0].SenseUsed)
	assert.Equal(t, []string{`You say, "hi"`}, texts(fredRec), "the actor always gets its text")

	square := w.room(2, "town", nil)
	sneak, _ := w.body(square, "Sneak", true)
	w.flags[sneak]["hidden"] = true
	_, bobRec := w.body(square, "Bob", true)
	m := say(sneak)
	m.ObserverText = `You hear {actor} whisper.`
	Send(m)
	assert.Equal(t, []string{`You hear someone whisper.`}, texts(bobRec), "lower case inside a sentence")
}

func TestFaintNeedsSuperhearing(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, _ := w.body(square, "Fred", true)
	_, steveRec := w.body(square, "Steve", true)
	keen, keenRec := w.body(square, "Keen", true)
	w.flags[keen]["superhearing"] = true

	m := say(fred)
	m.Faint = true
	Send(m)
	assert.Empty(t, steveRec.got)
	assert.Len(t, keenRec.got, 1)
}

func TestNearbyReachesNeighboursWithTheirDirection(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", map[string]int{"north": 2, "east": 3})
	w.room(2, "town", map[string]int{"south": 1})
	w.room(3, "town", nil) // no exit back: not reached, as shout today
	far := w.room(4, "town", map[string]int{"west": 1}) // an exit to the square, but the square has none to it
	fred, _ := w.body(square, "Fred", true)
	_, nRec := w.body(roomOf(t, 2), "Nora", true)
	_, eRec := w.body(roomOf(t, 3), "Ed", true)
	_, fRec := w.body(far, "Fay", true)
	_, sqRec := w.body(square, "Sam", true)

	Send(perception.Message{Actor: fred, Senses: perception.Sound, Range: perception.Nearby,
		ActorText:    `You shout, "HI"`,
		ObserverText: `{actor} shouts, "HI"`,
		FarText:      `{actor} shouts from the {direction} direction, "HI"`})

	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> shouts, "HI"`}, texts(sqRec))
	assert.Equal(t, []string{`Someone shouts from the <ansi fg="exit">south</ansi> direction, "HI"`}, texts(nRec))
	assert.True(t, nRec.got[0].Far)
	assert.Equal(t, "south", nRec.got[0].Direction)
	assert.Empty(t, eRec.got)
	assert.Empty(t, fRec.got, "neighbours are the rooms the source room leads to")
}

func TestRoomRangeWithATargetElsewhereCoversBothRooms(t *testing.T) {
	w := newWorld(t)
	a := w.room(1, "town", map[string]int{"north": 2})
	b := w.room(2, "town", map[string]int{"south": 1})
	archer, _ := w.body(a, "Archer", true)
	_, aRec := w.body(a, "Ann", true)
	deer, deerRec := w.body(b, "deer", false)
	_, bRec := w.body(b, "Bea", true)

	Send(perception.Message{Actor: archer, Target: deer, Senses: perception.Sight, Range: perception.Room,
		ActorText: `You shoot {target}.`, TargetText: `An arrow hits you.`, ObserverText: `{actor} shoots {target}.`})

	assert.Equal(t, []string{`<ansi fg="username">Archer</ansi> shoots <ansi fg="mobname">deer</ansi>.`}, texts(aRec))
	assert.Equal(t, []string{`<ansi fg="username">Archer</ansi> shoots <ansi fg="mobname">deer</ansi>.`}, texts(bRec),
		"the target's room is a source room, so its observers are not far")
	assert.Equal(t, []string{`An arrow hits you.`}, texts(deerRec))
}

func TestZoneAndWorldRanges(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	gate := w.room(2, "town", nil)
	cave := w.room(3, "caves", nil)
	crier, _ := w.body(square, "Crier", true)
	_, gateRec := w.body(gate, "Gus", true)
	_, caveRec := w.body(cave, "Cass", true)

	m := perception.Message{Actor: crier, Senses: perception.Sound, Range: perception.Zone,
		ActorText: `You call out.`, ObserverText: `{actor} calls out.`, FarText: `A call echoes through town.`}
	Send(m)
	assert.Equal(t, []string{`A call echoes through town.`}, texts(gateRec))
	assert.Empty(t, caveRec.got)

	m.Range = perception.World
	m.FarText = ``
	Send(m)
	assert.Equal(t, []string{`Someone calls out.`}, texts(caveRec), "far receivers fall back to observer text and do not see the actor")
}

func TestActorOnlyAndInterface(t *testing.T) {
	w := newWorld(t)
	cave := w.room(1, "caves", nil)
	w.light[1] = 0
	fred, fredRec := w.body(cave, "Fred", true)
	_, steveRec := w.body(cave, "Steve", true)

	Send(perception.Message{Kind: perception.Interface, Actor: fred, Audience: perception.ActorOnly,
		Range: perception.World, ActorText: `Your score: 3`})
	assert.Equal(t, []string{`Your score: 3`}, texts(fredRec), "interface text needs no sense and no light")
	assert.Empty(t, steveRec.got)
}

func TestNamedAudience(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	other := w.room(2, "town", nil)
	fred, fredRec := w.body(square, "Fred", true)
	steve, steveRec := w.body(square, "Steve", true)
	_, samRec := w.body(square, "Sam", true)
	away, awayRec := w.body(other, "Away", true)

	Send(perception.Message{Actor: fred, Senses: perception.Sound, Range: perception.Room,
		Audience: perception.Named, Named: []ids.ID{steve, away},
		ActorText: `You whisper.`, ObserverText: `{actor} whispers to you.`})

	assert.Len(t, fredRec.got, 1)
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> whispers to you.`}, texts(steveRec))
	assert.Empty(t, samRec.got, "not named")
	assert.Empty(t, awayRec.got, "named but out of range")
}

func TestWorldEventWithoutAnActor(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	_, fredRec := w.body(square, "Fred", true)

	Send(perception.Message{Room: 1, Senses: perception.Sound, Range: perception.Room,
		ObserverText: `A door creaks.`})
	require.Len(t, fredRec.got, 1)
	assert.Equal(t, `A door creaks.`, fredRec.got[0].Text)
	assert.True(t, fredRec.got[0].ActorSeen)
}

func TestBehaviourReceivesTheStructuredPerception(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, _ := w.body(square, "Fred", true)
	_, ratRec := w.body(square, "rat", false)

	Send(say(fred))
	require.Len(t, ratRec.got, 1)
	p := ratRec.got[0]
	assert.Equal(t, perception.Observer, p.Role)
	assert.Equal(t, fred, p.Message.Actor)
	assert.Equal(t, perception.Sound, p.SenseUsed)
	assert.True(t, p.ActorSeen)
}

func TestReceiversAreDeduplicated(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", map[string]int{"north": 2})
	w.room(2, "town", map[string]int{"south": 1})
	fred, _ := w.body(square, "Fred", true)
	_, samRec := w.body(square, "Sam", true)

	Send(perception.Message{Actor: fred, Senses: perception.Sound, Range: perception.World,
		ActorText: `x`, ObserverText: `y`})
	assert.Len(t, samRec.got, 1)
}

func roomOf(t *testing.T, id int) *rooms.Room {
	t.Helper()
	r := rooms.LoadRoom(id)
	require.NotNil(t, r)
	return r
}
```

  In `render_test.go`:

```go
package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapitalisedSomeoneOnlyAtTheStart(t *testing.T) {
	assert.Equal(t, `Someone waves at someone.`, resolveTokens(`{actor} waves at {target}.`, `someone`, `someone`, ``))
	assert.Equal(t, `<ansi fg="exit">north</ansi>`, resolveTokens(`{direction}`, ``, ``, `north`))
	assert.Equal(t, `no tokens`, resolveTokens(`no tokens`, `a`, `b`, `c`))
}
```

- [ ] **Step 2:** Run `go test ./internal/messaging/ -count=1`. Expected: FAIL (undefined `Send`,
  `hasFlag`, `roomVisibility`, `resolveTokens`).

- [ ] **Step 3: Write `perceive.go`:**

```go
package messaging

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// CanPerceive reports whether receiver, in room, takes in m through sense.
// It is the hook Phase 5's ruleset replaces (spec section 2). In 4b: sight
// needs light or nightvision (look's rule); a faint sound needs
// superhearing (today's quiet messages); touch and smell pass.
var CanPerceive = func(receiver *actors.Actor, room *rooms.Room, m perception.Message, sense perception.Sense) bool {
	switch sense {
	case perception.Sight:
		return canSeeIn(receiver, room)
	case perception.Sound:
		return !m.Faint || hasFlag(receiver.Character, "superhearing")
	default:
		return true
	}
}

// hasFlag and roomVisibility are seams for tests: buff flags need loaded buff
// specs, and visibility reads the game clock and biome.
var (
	hasFlag        = func(c *characters.Character, flag string) bool { return c.HasBuffFlag(flag) }
	roomVisibility = func(r *rooms.Room) int { return r.GetVisibility() }
)

// canSeeIn reports whether receiver can see in room.
func canSeeIn(receiver *actors.Actor, room *rooms.Room) bool {
	if room == nil {
		return false
	}
	return roomVisibility(room) >= 1 || hasFlag(receiver.Character, "nightvision")
}

// canSee reports whether receiver sees the body subject: a body always sees
// itself; a hidden body is unseen; a far receiver, or one whose sight is
// blocked in its own room, sees no one.
func canSee(receiver *actors.Actor, room *rooms.Room, subject *actors.Actor, far bool) bool {
	if subject == nil {
		return true
	}
	if subject.Id == receiver.Id {
		return true
	}
	if far || hasFlag(subject.Character, "hidden") {
		return false
	}
	return canSeeIn(receiver, room)
}

// perceivedBy returns the first sense in m.Senses that carries m to
// receiver, or zero when none does.
func perceivedBy(receiver *actors.Actor, room *rooms.Room, m perception.Message) perception.Sense {
	for _, sense := range m.Senses.Each() {
		if CanPerceive(receiver, room, m, sense) {
			return sense
		}
	}
	return 0
}
```

- [ ] **Step 4: Write `render.go`:**

```go
package messaging

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
)

// nameFor is how receiver reads subject's name: the plain coloured name
// today's room text uses, or "someone" when unseen.
func nameFor(subject *actors.Actor, seen bool) string {
	if subject == nil {
		return ``
	}
	if !seen {
		return `someone`
	}
	style := `mobname`
	if ctrl := subject.Controller(); ctrl != nil && ctrl.Account() != nil {
		style = `username`
	}
	return fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, style, subject.Character.Name)
}

// resolveTokens fills {actor}, {target} and {direction}. A "someone" that
// opens the text is capitalised.
func resolveTokens(text, actorName, targetName, direction string) string {
	if direction != `` {
		direction = fmt.Sprintf(`<ansi fg="exit">%s</ansi>`, direction)
	}
	for _, tok := range []struct{ token, value string }{
		{`{actor}`, actorName}, {`{target}`, targetName}, {`{direction}`, direction},
	} {
		if strings.HasPrefix(text, tok.token) && strings.HasPrefix(tok.value, `someone`) {
			text = `Someone` + tok.value[len(`someone`):] + text[len(tok.token):]
		}
		text = strings.ReplaceAll(text, tok.token, tok.value)
	}
	return text
}
```

- [ ] **Step 5: Write `messaging.go`:**

```go
// Package messaging delivers messages (Phase 4b spec section 2): it finds
// who a message reaches from its range and audience, asks CanPerceive for
// each receiver, picks the receiver's role and text with tokens resolved
// for them, and hands the result to the receiver's controller. The
// vocabulary is in internal/perception.
package messaging

import (
	"sort"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// receiver is one body a message reaches, with where it stands.
type receiver struct {
	actor     *actors.Actor
	room      *rooms.Room
	far       bool
	direction string
}

// Send delivers m to everyone it reaches.
func Send(m perception.Message) {
	actor := actors.Get(m.Actor)
	target := actors.Get(m.Target)

	if m.Kind == perception.Interface || m.Audience == perception.ActorOnly {
		if actor != nil {
			deliver(m, actor, target, receiver{actor: actor, room: roomOfActor(actor)}, perception.Actor, 0)
		}
		return
	}

	for _, rc := range receiversOf(m, actor, target) {
		role := perception.Observer
		switch {
		case actor != nil && rc.actor.Id == actor.Id:
			deliver(m, actor, target, rc, perception.Actor, 0)
			continue
		case target != nil && rc.actor.Id == target.Id:
			role = perception.Target
		}
		if m.Audience == perception.Named && !named(m, rc.actor.Id) {
			continue
		}
		sense := perceivedBy(rc.actor, rc.room, m)
		if sense == 0 {
			continue
		}
		deliver(m, actor, target, rc, role, sense)
	}
}

func deliver(m perception.Message, actor, target *actors.Actor, rc receiver, role perception.Role, sense perception.Sense) {
	p := perception.Perception{
		Role:       role,
		Message:    m,
		SenseUsed:  sense,
		ActorSeen:  canSee(rc.actor, rc.room, actor, rc.far),
		TargetSeen: canSee(rc.actor, rc.room, target, rc.far),
		Far:        rc.far,
		Direction:  rc.direction,
	}
	p.Text = resolveTokens(textFor(m, role, rc.far),
		nameFor(actor, p.ActorSeen), nameFor(target, p.TargetSeen), rc.direction)
	if ctrl := rc.actor.Controller(); ctrl != nil {
		ctrl.Perceive(p)
	}
}

// textFor picks the text for a role (decision 6 of the slice 3 plan).
func textFor(m perception.Message, role perception.Role, far bool) string {
	switch {
	case role == perception.Actor:
		return m.ActorText
	case role == perception.Target && m.TargetText != ``:
		return m.TargetText
	case far && m.FarText != ``:
		return m.FarText
	}
	return m.ObserverText
}

func named(m perception.Message, id ids.ID) bool {
	for _, n := range m.Named {
		if n == id {
			return true
		}
	}
	return false
}

func roomOfActor(a *actors.Actor) *rooms.Room {
	if a == nil || a.Character == nil || !rooms.IsRoomLoaded(a.Character.RoomId) {
		return nil
	}
	return rooms.LoadRoom(a.Character.RoomId)
}

// sourceRooms are the actor's room and the target's room when it differs,
// or the message's own room for a world event without an actor.
func sourceRooms(m perception.Message, actor, target *actors.Actor) []*rooms.Room {
	out := []*rooms.Room{}
	add := func(r *rooms.Room) {
		if r == nil {
			return
		}
		for _, have := range out {
			if have.RoomId == r.RoomId {
				return
			}
		}
		out = append(out, r)
	}
	if actor != nil {
		add(roomOfActor(actor))
	} else if m.Room != 0 && rooms.IsRoomLoaded(m.Room) {
		add(rooms.LoadRoom(m.Room))
	}
	add(roomOfActor(target))
	return out
}

// receiversOf lists who m's range reaches, source rooms first, each body once.
// The actor is included even when it is in no room.
func receiversOf(m perception.Message, actor, target *actors.Actor) []receiver {
	out := []receiver{}
	seen := map[ids.ID]bool{}
	add := func(a *actors.Actor, r *rooms.Room, far bool, direction string) {
		if a == nil || seen[a.Id] {
			return
		}
		seen[a.Id] = true
		out = append(out, receiver{actor: a, room: r, far: far, direction: direction})
	}

	src := sourceRooms(m, actor, target)
	inSource := map[int]bool{}
	for _, r := range src {
		inSource[r.RoomId] = true
		for _, id := range r.GetBodies() {
			add(actors.Get(id), r, false, ``)
		}
	}
	if actor != nil {
		add(actor, roomOfActor(actor), false, ``)
	}

	switch m.Range {
	case perception.Nearby:
		for _, r := range src {
			for _, nId := range neighbourIds(r) {
				if inSource[nId] || !rooms.IsRoomLoaded(nId) {
					continue
				}
				n := rooms.LoadRoom(nId)
				back := n.FindExitTo(r.RoomId)
				if back == `` {
					continue
				}
				for _, id := range n.GetBodies() {
					add(actors.Get(id), n, true, back)
				}
			}
		}
	case perception.Zone, perception.World:
		zones := map[string]bool{}
		for _, r := range src {
			zones[r.Zone] = true
		}
		for _, a := range actors.All() {
			r := roomOfActor(a)
			if r == nil || inSource[r.RoomId] {
				continue
			}
			if m.Range == perception.Zone && !zones[r.Zone] {
				continue
			}
			add(a, r, true, ``)
		}
	}
	return out
}

// neighbourIds are the rooms r leads to through its exits, temporary exits
// and active mutator exits, in room ID order.
func neighbourIds(r *rooms.Room) []int {
	set := map[int]bool{}
	for _, e := range r.Exits {
		set[e.RoomId] = true
	}
	for _, e := range r.ExitsTemp {
		set[e.RoomId] = true
	}
	for mut := range r.ActiveMutators {
		for _, e := range mut.GetSpec().Exits {
			set[e.RoomId] = true
		}
	}
	out := make([]int, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}
```

  Notes for the implementer:
  - `actors.All()` and `actors.Get` take the registry lock; `Send` must not be called with it held
    (nothing in this slice does).
  - If `mut.GetSpec()` cannot be called on the range variable (pointer receiver on a value), copy it
    to a local first (`m := mut; m.GetSpec()`), as the code in `rooms` does; check how
    `rooms.go:1594-1600` compiles and follow it.
  - In `TestNearbyReachesNeighboursWithTheirDirection` the body `Fay` is in room 4, which has an exit
    to the square but the square has none to it: it must get nothing.

- [ ] **Step 6:** Run `go test ./internal/messaging/ -count=1 -v`. Expected: all pass. A failing
  assertion is a finding: read it against the decisions above before changing code or test, and
  report any test you changed and why.

- [ ] **Step 7:** `go build ./... && go vet ./internal/messaging/ ./internal/perception/ ./internal/actors/ ./internal/rooms/`
  and `gofmt -l internal/messaging internal/perception internal/actors internal/rooms`. Expected:
  clean.

- [ ] **Step 8: Commit**

```bash
git add internal/messaging
git commit -F - <<'EOF'
messaging: Send resolves receivers, senses, roles and tokens (Phase 4b slice 3)

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 5: Connection-write guard

**Files:** Create `internal/guards/messaging_test.go`.

- [ ] **Step 1: Write it:**

```go
package guards

import (
	"go/ast"
	"strings"
	"testing"
)

// TestOnlyMessagingWritesToConnections enforces the slice 3 guard (spec
// section 2): text reaches a connection only through internal/messaging.
// connectionWriters lists today's other writers; slice 7 moves them and the
// list may only shrink (an entry that no longer writes fails the test). GMCP
// keeps its own path for good. main.go also writes "server full" notices
// straight to a net.Conn before a connection is registered (main.go,
// "Server is full"); this guard does not see those, and slice 7 moves them.
func TestOnlyMessagingWritesToConnections(t *testing.T) {
	writers := map[string]bool{"SendTo": true, "SendRawTo": true, "Broadcast": true}
	permanent := map[string]string{
		"modules/gmcp/gmcp.go": "GMCP keeps its own path (spec section 2)",
	}
	connectionWriters := map[string]string{
		"internal/hooks/Broadcast_SendToAll.go":                   "slice 7",
		"internal/hooks/Log_FollowLogs.go":                        "slice 7",
		"internal/hooks/Message_SendMessages.go":                  "slice 7: the text write behind Account.SendText",
		"internal/hooks/PlayerDespawn_HandleLeave.go":             "slice 7",
		"internal/hooks/RedrawPrompt_SendRedraw.go":               "slice 7",
		"internal/hooks/WebClientCommand_SendWebClientCommand.go": "slice 7",
		"internal/inputhandlers/cleanser.go":                      "slice 7",
		"internal/inputhandlers/echo.go":                          "slice 7",
		"internal/inputhandlers/login.go":                         "slice 7: login screens, before an actor exists",
		"internal/inputhandlers/login_prompt_handler.go":          "slice 7: login screens, before an actor exists",
		"internal/inputhandlers/systemcommands.go":                "slice 7",
		"internal/inputhandlers/term_ansi.go":                     "slice 7: terminal negotiation",
		"internal/inputhandlers/term_iac.go":                      "slice 7: telnet negotiation",
		"main.go":                                                 "slice 7",
		"world.go":                                                "slice 7",
	}
	inside := func(path string) bool {
		return strings.HasPrefix(path, "internal/connections/") ||
			strings.HasPrefix(path, "internal/messaging/") ||
			strings.HasPrefix(path, "internal/perception/")
	}

	used := map[string]bool{}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || !writers[sel.Sel.Name] {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); !ok || x.Name != "connections" {
				return true
			}
			switch {
			case inside(path):
			case permanent[path] != "":
			case connectionWriters[path] != "":
				used[path] = true
			default:
				t.Errorf("%s: writes to a connection (connections.%s) outside internal/messaging", path, sel.Sel.Name)
			}
			return true
		})
	}
	for path := range connectionWriters {
		if !used[path] {
			t.Errorf("%s no longer writes to a connection: remove it from connectionWriters", path)
		}
	}
}
```

  Check with `grep -rn 'connections\.\(SendTo\|SendRawTo\|Broadcast\)' --include=*.go .` that the
  list matches the facts table; if a file's import alias for `connections` is not `connections`,
  report it (the guard would not see it).

- [ ] **Step 2: Prove it bites.** Add `connections.SendTo(nil)` to a function in
  `internal/usercommands/say.go` with the Edit tool (and the import if needed), run
  `go test ./internal/guards/ -run TestOnlyMessagingWritesToConnections -count=1`, see it fail naming
  `say.go`; remove the edit with Edit and confirm `git diff internal/usercommands/say.go` is empty.
  Then rename one allowlist key (for example `world.go` to `worldX.go`), run again, see both the
  "no longer writes" failure and the `world.go` failure, and restore it.

- [ ] **Step 3:** Run `go test ./internal/guards/ -count=1`. Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/guards/messaging_test.go
git commit -F - <<'EOF'
guards: connection writes only in messaging (allowlist of today's writers may only shrink)

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 6: Verify (controller)

- [ ] `go build ./...`; `go test ./... -count=1` (expected: the branch's base baseline plus 2 new
  packages, `perception` and `messaging`; the 3 known Windows failures only); `gofmt -l` on changed
  Go files; `go run ./cmd/smoke` (passes; nothing in production calls `messaging.Send` yet).

### Task 7: Records (controller)

- [ ] `PATCH_NOTES.md`: a line for slice 3 (perception vocabulary, messaging delivery with roles,
  senses, ranges, audiences, tokens and the `CanPerceive` hook; controllers perceive; connection
  write guard). `status.md`: slice 3 built; next slice 4. Spec: section 4 "Built so far" adds 3;
  "Notes for later slices" gains: slice 7 moves `main.go`'s raw `net.Conn` "server full" writes
  (unguarded) and the telnet and terminal negotiation writes (`term_iac.go`, `term_ansi.go`, `echo.go`),
  which are protocol bytes, not text; decide there whether they are a second named exception like
  GMCP; slice 7 adds interface text before an actor exists; slices 5 and 7 write texts with
  `{actor}` at the start when today's line starts with the name (so "Someone" capitalises).
  `geomud-verifying`: new package count. Commit with named paths. Move this plan to
  `docs/completed/` when the PR merges.

## Done when

`messaging` and `perception` exist with the tests above passing (each role, range and audience; a dark
room blocks Sight; "Someone" when the actor is unseen; a behaviour receives the structured
perception); `actors.Controller` has `Perceive` on both controllers; the connection-write guard passes
with the allowlist of today's writers; build, the full test suite at the baseline, and smoke pass.
