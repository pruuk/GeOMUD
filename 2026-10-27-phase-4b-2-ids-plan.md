# Phase 4b slice 2: one ID space (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** every live body (player or NPC) has one ID of the form `0x<cell>:0x<random>`, made once
when the body is created or instantiated and kept for good; every reference to a live body holds
that ID; `#N`/`@N` give way to ID tokens that only clients send.

**Architecture:** a new leaf package `internal/ids` defines `ids.ID` (cell plus random, hex text
form). The ID lives on `characters.Character` (`Character.Id`), so a player's is saved in the user
file and a mob instance's is its identity (the mob instance counter goes). `actors.ActorId` becomes
an alias of `ids.ID` and the registry keys on the body's ID. Events, aggro, charm, killer, room
lists, visitor tracking, GMCP and `FindByName` all hold `ids.ID` for bodies; account-scoped data
(parties, kill credit, mob hatred, saved records) keeps `UserId`.

**Tech stack:** Go 1.27.3, `crypto/rand`, testify. No new dependency: cells are `uint64` and every
room uses `ids.NoCell` until the geo room key gives rooms a cell (the S2 library arrives with it).

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md`, decisions 10 and 11, section 4
slice 2.

---

## Decisions this plan applies (Calabe, 2026-10-27)

- IDs are `0x<16 hex>:0x<16 hex>`: the cell key of the room where the thing was created, then a
  random 64-bit part. Rooms without geography (all of them, today) use cell `0x0`.
- An ID is made once, at creation or instantiation, and kept through logout, copyover, restart and
  anything else. A player's body ID is assigned at character creation and saved. A new NPC
  instantiation gets a new ID. NPC and companion persistence across copyover and restart is slice
  2b; item IDs are slice 2c.
- People name things by name; GMCP carries IDs as opaque tokens; `#N` and `@N` go. The name forms
  `FindByName` accepts today (including its positional `name#2`) are unchanged in this slice.

Defaults set by the controller from those rulings (recorded here so they are not re-asked):

- Account-scoped data keeps `UserId`: parties, `PlayerDamage` and kill credit, `KD`,
  `mobsHatePlayers` (a mob's hostility follows the account across logins), `Sign.VisibleUserId`,
  `Item.StashedBy`, connection-addressed GMCP, `Message.UserId`, `System` events.
- Room message exclude lists (`Message.ExcludeUserIds`, `room.SendText(txt, userIds...)`) keep
  user IDs; slice 7 replaces all output. `Room.GetPlayers()` keeps returning user IDs, derived from
  the single body list.
- Two pre-existing bugs that unifying IDs would silently change are fixed here: `follow`'s mob-death
  handler uses the template `MobId` as an instance ID, and `findUserExit` resolves mob visitors as
  user IDs. The other bugs in the inventory (E8c-g) sit in `NewRound_DoCombat.go` and go to the
  slice 5 plan unchanged.
- Dead code that would otherwise be converted is deleted: `Character.SetCast` (no callers) with
  `SpellAggroInfo.TargetUserIds`/`TargetMobInstanceIds` and the SpellCast branches that read them;
  `Mob.PlayerAttacked`/`HasAttackedPlayer`/`playersAttacked` (no callers); event types with no
  producer and no consumer (`RoomAction`, `SkillUsed`, `CLIRequest`, `Purchase`, `Party`).
- The web admin JSON field names for runtime state change (`Aggro`, `Charmed`, `CharmedMobs`,
  `KillerMobInstanceId`, `Mob.InstanceId`); nothing under `_datafiles/html` reads them except the
  docs page, which this slice edits. `UpdateOnlineUser` keeps the live `Character.Id`, so an admin
  edit cannot change a body's ID.

## Facts verified against source (2026-10-27, master `df1b04c6`)

Rows marked (inv) come from the slice 2 inventory, which read 86 files (79 in full, 7 in the parts
the patterns touched) on 2026-10-27; the controller re-checked the rows marked (ctl).

| Fact | Value | Where |
|---|---|---|
| `MobInstanceId` | 157 non-test uses in 37 files; 59 of them in files slices 4-6 delete or rewrite (mobcommands 15, `NewRound_DoCombat.go` 38, `NewRound_IdleMobs.go` 3, `NewRound_MobRoundTick.go` 2, `MobIdle_HandleIdleMobs.go` 1) | `grep` (ctl, inv) |
| Other patterns | `.InstanceId` 85, `mobs.GetInstance(` 61, `GetAllMobInstanceIds` 6, `KillerMobInstanceId` 5, `CharmedMobs` 7, `GetCharmIds` 6, `TrackCharmed` 6, `Charmed.UserId` 4, `IsCharmed` 15, `Aggro.UserId` 27, `Aggro.MobInstanceId` 19, `SetAggro` 12, `SetAggroRemote` 3, `IsAggro` 2, `GetPlayers` 25, `GetMobs` 19, `FindByName` 19 | (inv) Table A |
| Character body identity | `Character` has no ID field; new bodies come from `characters.New()` at `accounts/userrecord.go:63` (`NewAccount`), the permadeath reroll `usercommands/suicide.go:133`, and mob instantiation `mobs.addInstance` (`mobs.go:42-49`) | (ctl) |
| Mob instance IDs | `instanceCounter` (`mobs.go:23`), `mobInstances map[int]*Mob`, `addInstance` is the only add path, `DestroyInstance` the only remove; `OnInstanceDestroyed util.Hook[int]` | (ctl, inv D1) |
| Actor IDs today | `actors.lastId` counter separate from `instanceCounter`, so a mob's `ActorId` differs from its instance ID; `byMobInstanceId` index | `internal/actors/registry.go` (inv E1) |
| Import graph | `events` imports `connections items mudlog stats util`; `characters` imports no `accounts`/`actors`; `mobs` imports `characters events`; `accounts` imports `mobs events`; `actors` imports `accounts characters mobs mudlog`; nothing imports `actors` but the root package. So `ids` must be a leaf, and `rooms`, `hooks`, `usercommands`, `combat`, `modules/*` may import `actors` | `go list` (ctl) |
| Client round trip | GMCP `Room.Info.Contents.Npcs[].id` = `Mob.ShorthandId()` `#<InstanceId>` (`gmcp.Room.go:376`), `Players[].id` = `Account.ShorthandId()` `@<UserId>` (`:348`), `Char.Enemies[].id` `#<InstanceId>` (`gmcp.Char.go:452`). `window-room.js:494-521` sends `look <id>`, `attack <id>`, `list <id>` with the id as an opaque string. `Room.FindByName` parses `#N`/`@N` (`rooms.go:1396-1469`) | (ctl) |
| Generated commands | Party auto-attack sends `attack #%d` and `attack @%d` (`usercommands/attack.go:179,227`) | (ctl) |
| `ShorthandId` | `Account` (`userrecord.go:161`), `Mob` (`mobs.go:266`), `Item` (`items.go:478`, stays until 2c) | (ctl) |
| Persistence | Room lists, visitors, `SpawnInfo.InstanceId`, aggro, charm and killer are runtime only; copyover carries `{UserId, ConnectionId, IsLinkDead, LinkDeadTurn}` and reloads the user file; mobs respawn | (inv C, D) |
| YAML and text marshalling | Items' `uuid.UUID` saves through `MarshalText`/`UnmarshalText` with `gopkg.in/yaml.v2` (`internal/uuid/uuid.go:106-125`) | (ctl) |
| User migrations | `accounts.DoUserMigrations()` runs from `main.go` when the user index does not exist; `SearchOfflineUsers` walks every user file (`accounts/migration.go`, `users.go:481`) | (ctl) |
| Docs page | `_datafiles/html/admin/docs-modules.html:1124-1228` documents `MobInstanceId` on events and `InstanceId` on `MobDeath` | (inv D6) |
| Event classification | Body fields to convert: `Buff`, `BuffsTriggered`, `BuffsAdded`, `BuffsRemoved`, `Input`, `RoomChange`, `MobIdle`, `EquipmentChange`, `ItemOwnership`, `AggroChanged` (`UserId`/`MobInstanceId` pairs), `MobDeath.InstanceId`, `Communication.SourceUserId`/`SourceMobInstanceId`. Account fields stay. Dead: `RoomAction`, `SkillUsed`, `CLIRequest`, `Purchase`, `Party` | (inv Table B) |

The inventory is in the repo beside this plan: `2026-10-27-phase-4b-2-inventory.md` (Tables A and
B, room internals in C, ids leaving the process in D, risks in E). It moves to `docs/completed/`
with this plan.

## Rules for every task

- Edit files with the Edit and Write tools only. No `sed -i`, no shell redirects into the repo, no
  Python write-opens, no `gofmt -w` or `gopls -w` in the repo. A generated file is written to the
  session scratchpad and landed with `cp`. `~/.claude/hooks/guard.py` enforces it.
- `git add <named paths>` only. Commit messages through a quoted heredoc ending with
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Every task ends with `go build ./...`, the targeted tests named in it, and
  `git diff --name-only -- '*.go' | xargs -r gofmt -l` printing nothing. Tasks 4, 6, 7 and 11 also
  run the full suite (`go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'`, only the 3
  Windows failures in `geomud-verifying` allowed) and `go run ./cmd/smoke`.
- `<session scratchpad>` means the scratchpad directory named in your system prompt, as a full
  path. Never `C:/tmp`, `$TMP` or `/tmp`.
- A step whose expected output does not match: stop and report it. Do not change the expectation.
- Conversion tasks are compiler-driven: change the type, then fix each compile error by the task's
  rule table. A site the rule table does not cover: stop and report it with file:line.

## File structure

| File | Responsibility |
|---|---|
| `internal/ids/ids.go`, `ids_test.go` (new) | `ids.ID`: make, format, parse, text marshalling |
| `cmd/newid/main.go` (new) | Prints a fresh ID for content a builder creates by hand |
| `internal/characters/character.go` | `Character.Id`; aggro, charm and killer hold `ids.ID` |
| `internal/characters/aggro.go`, `charminfo.go` | `Aggro.Target`, `CharmInfo.By` |
| `internal/accounts/` | Body ID at creation and in the startup migration; `UpdateOnlineUser` keeps it |
| `internal/mobs/mobs.go` | Instance identity is `Character.Id`; `instanceCounter` goes |
| `internal/actors/` | `ActorId = ids.ID`; registry keyed by body ID; `ByMobInstanceId` goes; `BodyReplaced` |
| `internal/events/eventtypes.go` | Body fields become `ActorId ids.ID`; dead types go |
| `internal/rooms/` | One body list; `GetPlayers` (user IDs) and `GetMobs` (IDs) derived from it; split visitors; `FindByName` takes ID tokens |
| `modules/gmcp`, `modules/follow`, `modules/cleanup` | IDs in payloads and keys |
| `internal/guards/guards_test.go` | No `MobInstanceId`; IDs made only at creation sites; no `#%d`/`@%d` |
| `_datafiles/html/admin/docs-modules.html` | Event field docs |
| `_datafiles/world/empty/users/1.yaml` | The stock admin's character gets its saved ID |

---

### Task 0: Branch

- [ ] **Step 1:** the spec update and this plan merge to `master` as their own PR first, then:

```bash
cd "C:/Users/Calabe Davis/workspace/new_geo_mudbase"
git switch master && git pull --ff-only origin master
git switch -c strip/phase-4b-2-ids
```

---

### Task 1: The `ids` package

**Files:** Create `internal/ids/ids.go`, `internal/ids/ids_test.go`

- [ ] **Step 1: Write the failing test** (`internal/ids/ids_test.go`)

```go
package ids

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var format = regexp.MustCompile(`^0x[0-9a-f]{16}:0x[0-9a-f]{16}$`)

func TestNewIsUniqueAndKeepsTheCell(t *testing.T) {
	a, b := New(NoCell), New(0x89c25a31ec0f0e5f)
	assert.False(t, a.IsZero())
	assert.NotEqual(t, a, b)
	assert.NotEqual(t, New(NoCell), New(NoCell))
	assert.Equal(t, uint64(0), a.Cell())
	assert.Equal(t, uint64(0x89c25a31ec0f0e5f), b.Cell())
	assert.Regexp(t, format, a.String())
	assert.Equal(t, "0x89c25a31ec0f0e5f:", b.String()[:19])
}

func TestZero(t *testing.T) {
	var z ID
	assert.True(t, z.IsZero())
	assert.Equal(t, "", z.String())
}

func TestParseRoundTrip(t *testing.T) {
	id := New(0x1234)
	got, err := Parse(id.String())
	require.NoError(t, err)
	assert.Equal(t, id, got)

	for _, bad := range []string{"", "0x1:0x2", "1234567890123456:1234567890123456", "0x89c25a31ec0f0e5f-0xd134e199a405a163",
		"0x89c25a31ec0f0e5f:0xD134E199A405A163", "0x89c25a31ec0f0e5f:0x0000000000000000", "#12", "@5"} {
		_, err := Parse(bad)
		assert.Error(t, err, bad)
	}
}

type holder struct {
	Id    ID `yaml:"id,omitempty"`
	Other int
}

func TestTextMarshalling(t *testing.T) {
	id := New(0x42)

	j, err := json.Marshal(holder{Id: id})
	require.NoError(t, err)
	var jh holder
	require.NoError(t, json.Unmarshal(j, &jh))
	assert.Equal(t, id, jh.Id)

	y, err := yaml.Marshal(holder{Id: id})
	require.NoError(t, err)
	assert.Contains(t, string(y), "id: "+id.String())
	var yh holder
	require.NoError(t, yaml.Unmarshal(y, &yh))
	assert.Equal(t, id, yh.Id)

	empty, err := yaml.Marshal(holder{Other: 1})
	require.NoError(t, err)
	assert.NotContains(t, string(empty), "id:", "a zero ID is omitted")
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./internal/ids/ -count=1`
Expected: FAIL, `undefined: New` (or no non-test files).

- [ ] **Step 3: Create `internal/ids/ids.go`**

```go
// Package ids is the engine's one ID scheme (Phase 4b decisions 10 and 11).
// An ID is the cell key of the room or region where a thing was created and a
// random 64-bit part, written 0x<cell>:0x<random> in lowercase hex, in the
// style of Google's feature IDs. An ID is made once, when a thing is created
// or instantiated, and kept for good; internal/guards lists the places allowed
// to make one.
package ids

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
)

// NoCell is the cell of a thing created where there is no geography yet.
// Every room uses it until the geo room key gives rooms a cell.
const NoCell uint64 = 0

// ID identifies one thing for good. The zero ID means none.
type ID struct {
	cell uint64
	rand uint64
}

// New makes the ID for a thing created now in cell. The random part is never
// zero, so a made ID is never the zero ID.
func New(cell uint64) ID {
	var b [8]byte
	for {
		rand.Read(b[:]) // crypto/rand.Read never returns an error (Go 1.24+)
		if r := binary.BigEndian.Uint64(b[:]); r != 0 {
			return ID{cell: cell, rand: r}
		}
	}
}

// IsZero reports whether id is the zero ID (no thing).
func (id ID) IsZero() bool { return id == ID{} }

// Cell is the cell key the thing was created in.
func (id ID) Cell() uint64 { return id.cell }

// String is the text form, 0x<16 hex>:0x<16 hex>, or "" for the zero ID.
func (id ID) String() string {
	if id.IsZero() {
		return ""
	}
	return fmt.Sprintf("0x%016x:0x%016x", id.cell, id.rand)
}

// Parse reads the text form exactly: both halves 0x-prefixed, 16 lowercase
// hex digits each, and a non-zero random half.
func Parse(s string) (ID, error) {
	if len(s) != 37 || s[0:2] != "0x" || s[18] != ':' || s[19:21] != "0x" {
		return ID{}, fmt.Errorf("ids: %q is not 0x<16 hex>:0x<16 hex>", s)
	}
	cell, err1 := parseHex(s[2:18])
	r, err2 := parseHex(s[21:37])
	if err1 != nil || err2 != nil || r == 0 {
		return ID{}, fmt.Errorf("ids: %q is not 0x<16 hex>:0x<16 hex>", s)
	}
	return ID{cell: cell, rand: r}, nil
}

func parseHex(s string) (uint64, error) {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return 0, fmt.Errorf("not lowercase hex")
		}
	}
	return strconv.ParseUint(s, 16, 64)
}

func (id ID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *ID) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*id = ID{}
		return nil
	}
	got, err := Parse(string(b))
	if err != nil {
		return err
	}
	*id = got
	return nil
}
```

- [ ] **Step 4: Run, expect PASS.** `go test ./internal/ids/ -count=1 -v` and `go vet ./internal/ids/`.

- [ ] **Step 5: Commit**

```bash
git add internal/ids/ids.go internal/ids/ids_test.go
git commit -F - <<'EOF'
ids: one ID scheme, 0x<cell>:0x<random>

Phase 4b slice 2, decision 10. A leaf package, so events, characters
and mobs can all hold IDs.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 2: A player's body ID, made at creation and saved

**Files:**
- Modify: `internal/characters/character.go` (field), `internal/accounts/users.go` (`CreateUser`, `UpdateOnlineUser`), `internal/accounts/migration.go`, `internal/usercommands/suicide.go:133`, `main.go` (call the migration)
- Test: `internal/accounts/bodyid_test.go` (new)

- [ ] **Step 1: Add the field.** First field of `characters.Character`:

```go
	Id ids.ID `yaml:"id,omitempty"` // The body's ID, made once at creation and kept for good (Phase 4b decision 11)
```

- [ ] **Step 2: Write the failing tests** (`internal/accounts/bodyid_test.go`)

```go
package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestGiveBodyIdOnlyWhenMissing(t *testing.T) {
	u := NewAccount(81, 0)
	require.True(t, u.Character.Id.IsZero())
	assert.True(t, giveBodyId(u))
	first := u.Character.Id
	assert.False(t, first.IsZero())
	assert.False(t, giveBodyId(u), "an existing ID is never replaced")
	assert.Equal(t, first, u.Character.Id)
}

func TestBodyIdSurvivesTheUserFile(t *testing.T) {
	u := NewAccount(82, 0)
	giveBodyId(u)
	b, err := yaml.Marshal(u)
	require.NoError(t, err)
	var back Account
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, u.Character.Id, back.Character.Id)
}

func TestUpdateOnlineUserKeepsTheBodyId(t *testing.T) {
	resetUserManager()
	live := NewAccount(83, 5)
	giveBodyId(live)
	userManager.Users[83] = live
	edited := *live
	ec := *live.Character
	ec.Id = ids.New(ids.NoCell) // an admin edit trying to change the ID
	edited.Character = &ec
	UpdateOnlineUser(edited)
	assert.Equal(t, live.Character.Id, userManager.Users[83].Character.Id)
	assert.NotEqual(t, ec.Id, userManager.Users[83].Character.Id)
}
```

- [ ] **Step 3: Run, expect a build failure** (`undefined: giveBodyId`).

- [ ] **Step 4: Implement.**

In `internal/accounts/migration.go` add:

```go
// giveBodyId makes u's character its body ID if it has none, and reports
// whether it did. A body ID is made once and never replaced (Phase 4b
// decision 11), so this is only for a new character and for characters saved
// before body IDs existed.
func giveBodyId(u *Account) bool {
	if u.Character == nil || !u.Character.Id.IsZero() {
		return false
	}
	u.Character.Id = ids.New(ids.NoCell)
	return true
}

// DoBodyIdMigrationV1 gives every saved character without a body ID one and
// saves it. main runs it at startup before anyone can log in, so no two loads
// of one account can race to give it two IDs. It is idempotent.
func DoBodyIdMigrationV1() error {
	var errorResult error
	SearchOfflineUsers(func(u *Account) bool {
		if giveBodyId(u) {
			if err := SaveUser(*u); err != nil {
				errorResult = err
				return false
			}
		}
		return true
	})
	return errorResult
}
```

(Read `SearchOfflineUsers` first: if it passes a pointer to a loop-local copy, `SaveUser(*u)` saves
that copy, which is what is wanted. If `SaveUser` needs index or config state that is not ready at
the point `main` calls it, stop and report.)

`CreateUser` (`users.go`): before its `SaveUser(*u)`, add `giveBodyId(u)`.

`UpdateOnlineUser`: inside the `if updated.Character != nil && u.Character != nil` block, before
`*u.Character = *updated.Character`, add
`updated.Character.Id = u.Character.Id // a body's ID never changes (decision 11)`.

`usercommands/suicide.go:133` (permadeath): after `*user.Character = *characters.New()` add
`user.Character.Id = ids.New(ids.NoCell) // a rerolled character is a new body` and leave the
actor re-keying to Task 5 (add a `// Task 5: actors.BodyReplaced` marker comment only if the build
needs nothing else; remove it in Task 5).

`main.go`: call `accounts.DoBodyIdMigrationV1()` right after the user-index block
(`if !isCopyover { ... InitUserIndex ... }`), inside that same `!isCopyover` guard, logging an error
with `mudlog.Error("DoBodyIdMigrationV1", "error", err)` if it returns one. Copyover skips it: the
process before the copyover already ran it.

- [ ] **Step 5: `cmd/newid`, and the stock admin's saved ID.** Builders hand-writing content need
fresh IDs, and the stock admin in `_datafiles/world/empty/users/1.yaml` should ship with one rather
than get a different one on every fresh copy. Create `cmd/newid/main.go`:

```go
// Command newid prints a new ID (Phase 4b decision 10) for a thing a builder
// is creating by hand, such as a character or record written into a world's
// data files. Each run is a new thing: never reuse an ID.
package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/ids"
)

func main() {
	cell := flag.String("cell", "0x0", "cell key of the room the thing is created in, hex (0x0 while rooms have no geography)")
	flag.Parse()
	c, err := strconv.ParseUint(*cell, 0, 64)
	if err != nil {
		fmt.Println("newid: -cell must be a number such as 0x0 or 0x89c25a31ec0f0e5f")
		return
	}
	fmt.Println(ids.New(c))
}
```

Run `go run ./cmd/newid` once and, with the Edit tool, add the printed value to
`_datafiles/world/empty/users/1.yaml` as the first key under `character:` (two-space indent:
`  id: 0x0000000000000000:0x...`). Task 8's guard lists `cmd/newid/main.go` as a creation site.

- [ ] **Step 6: Run the tests, expect PASS.** `go build ./... && go test ./internal/accounts/ ./internal/characters/ ./internal/usercommands/ -count=1`, then `go run ./cmd/smoke 2>&1 | tail -1` (PASS).

- [ ] **Step 7: Commit** the changed Go files, the test, `cmd/newid/main.go` and `_datafiles/world/empty/users/1.yaml`:

```bash
git commit -F - <<'EOF'
Body IDs for players: made at creation, saved, never replaced

Character.Id holds the body's ids.ID. CreateUser and the permadeath
reroll make one; a startup migration gives every saved character
without one an ID before anyone can log in; admin edits cannot change
it.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 3: A mob instance's identity is its body ID

**Files:** `internal/mobs/mobs.go`, `internal/mobs/lifecycle_test.go`, and every compile error that
follows (Table A lists the files: about 85 `.InstanceId` and 61 `mobs.GetInstance(` sites).

This is the large compiler-driven sweep. Implementer: Opus.

- [ ] **Step 1: Change the core** (`internal/mobs/mobs.go`):
  - delete `instanceCounter` and the `InstanceId int` field of `Mob`;
  - `mobInstances map[ids.ID]*Mob`; `recentlyDied map[ids.ID]int`;
  - `addInstance(mob Mob) *Mob` sets `mob.Character.Id = ids.New(ids.NoCell) // rooms carry no cell until the geo room key`, stores under it, fires `OnInstanceCreated`;
  - `GetInstance(id ids.ID)`, `MobInstanceExists(id ids.ID)`, `GetAllMobInstanceIds() []ids.ID`, `DestroyInstance(id ids.ID)`, `OnInstanceDestroyed util.Hook[ids.ID]`;
  - delete `Mob.ShorthandId`, `PlayerAttacked`, `HasAttackedPlayer`, `playersAttacked` (no callers);
  - `SpawnInfo.InstanceId` (`internal/rooms/spawninfo.go`) becomes `ids.ID`.
  - Update `internal/mobs/lifecycle_test.go` to the new types (it must still prove: an added instance gets a non-zero ID, two get different IDs, `GetInstance` finds it, both hooks fire).

- [ ] **Step 2: Fix every compile error by this table.** Run `go build ./... 2>&1 | head -50`, fix,
repeat.

| Old | New |
|---|---|
| `m.InstanceId` / `mob.InstanceId` | `m.Character.Id` |
| a variable, parameter, map key or slice of mob instance IDs typed `int` | `ids.ID` (rename `*MobInstanceId*` names to `*MobId*` only where `MobId` is not already the template ID in scope; otherwise `*Mob*Body*`; never leave the string `MobInstanceId` in an identifier) |
| `== 0`, `> 0`, `!= 0` on a mob instance ID | `.IsZero()` / `!x.IsZero()` |
| `0` passed as "no mob" | `ids.ID{}` |
| `%d` formatting of a mob instance ID | `%s` |
| `strconv.Itoa(id)` / `Atoi` of a mob instance ID | `id.String()` / `ids.Parse` |
| `mob.ShorthandId()` | `mob.Character.Id.String()` |
| event fields `MobInstanceId` | leave for Task 6; to compile now, assign `ids.ID` to them by changing only those fields' types to `ids.ID` in `eventtypes.go` (Task 6 renames and merges them) |
| `Aggro.MobInstanceId`, `SetAggro(userId, mobInstanceId, ...)` and friends | change their mob half to `ids.ID` now; Task 4 merges the halves |
| `CharmedMobs []int`, `TrackCharmed(mobId int, ...)`, `GetCharmIds() []int`, `KillerMobInstanceId int` | `[]ids.ID` / `ids.ID` now; Task 4 renames `KillerMobInstanceId` |
| `actors.ByMobInstanceId(id)` | `actors.Get(id)` once Task 5 lands; for now change the registry's `byMobInstanceId` key type to `ids.ID` |

Files that slices 4-6 delete (`internal/mobcommands/*`, `NewRound_IdleMobs.go`,
`NewRound_MobRoundTick.go`, `MobIdle_HandleIdleMobs.go`) and rewrite (`NewRound_DoCombat.go`) get
the same mechanical treatment: no redesign there.

`#N` parsing in `rooms.findMobByName` (`rooms.go:1455-1468`) stops compiling: replace that branch
with `ids.Parse(searchName)` (Task 7 finishes `FindByName`).

- [ ] **Step 3: Verify**

```bash
go build ./... && go vet ./internal/... ./modules/... .
grep -rn 'InstanceId\b' --include=*.go . | grep -v _test | grep -v 'SpawnInfo\|spawnInfo\|ids\.'   # read every hit; mob instance IDs must be gone
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'
go run ./cmd/smoke 2>&1 | tail -1
go run ./cmd/smoke -bots 3 2>&1 | tail -1
```

Expected: clean build and vet; the grep lists only `SpawnInfo.InstanceId` sites (now `ids.ID`) and
event fields awaiting Task 6; only the 3 known test failures; both smoke runs PASS.

- [ ] **Step 4: Commit** every changed file by name (`git diff --name-only` lists them; check
`git status --short` shows nothing untracked or unexpected):

```
Mob instance identity is the body's ids.ID

instanceCounter goes: addInstance gives each instantiation a new ids.ID
on its Character. Every mob instance reference follows. Dead
PlayerAttacked/HasAttackedPlayer and Mob.ShorthandId go.
```

---

### Task 4: Aggro, charm and killer hold one body ID

**Files:** `internal/characters/aggro.go`, `charminfo.go`, `character.go`, and the compile errors
that follow (`Aggro.UserId` 27 and `Aggro.MobInstanceId` 19 sites; `SetAggro` 12, `SetAggroRemote`
3, `IsAggro` 2, `Charmed.UserId` 4, `IsCharmed` 15, `KillerMobInstanceId` 5).

- [ ] **Step 1: Change the core.**
  - `Aggro{Type, Target ids.ID, SpellInfo, ExitName, RoundsWaiting}` (delete `MobInstanceId`, `UserId`).
  - `SetAggro(target ids.ID, aggroType AggroType, roundsWaitTime ...int)`, `SetAggroRemote(exitName string, target ids.ID, aggroType AggroType, ...)`, `IsAggro(target ids.ID) bool`.
  - Delete `SetCast` (no callers), `SpellAggroInfo.TargetUserIds`/`TargetMobInstanceIds`, and the `IsAggro` SpellCast branch that reads them.
  - `CharmInfo{By ids.ID, RoundsRemaining, ExpiredCommand}` (`By` is the charmer's body); `NewCharm(by ids.ID, ...)`; `Character.Charm(by ids.ID, ...)`; `GetCharmedBy() ids.ID` replaces `GetCharmedUserId`; `IsCharmed(by ...ids.ID)`.
  - `KillerMobInstanceId int` becomes `KilledBy ids.ID` (`yaml:"-"`, the body that delivered the killing blow).
  - `GetMobName`/`GetPlayerName(viewer ids.ID)` compare `Aggro.Target == viewer`; callers that passed `0` pass `ids.ID{}`.

- [ ] **Step 2: Fix compile errors by this table.**

| Old | New |
|---|---|
| `x.Aggro.UserId > 0` / `x.Aggro.MobInstanceId > 0` (which kind is the target) | `t := actors.Get(x.Aggro.Target); t != nil && t.Controller().Account() != nil` for "target is a player"; the negation for "target is an NPC". Packages that cannot import `actors` (`characters`) must not need this; stop and report if one does |
| `accounts.GetByUserId(x.Aggro.UserId)` | `actors.Get(x.Aggro.Target)` then `.Controller().Account()` |
| `mobs.GetInstance(x.Aggro.MobInstanceId)` | `mobs.GetInstance(x.Aggro.Target)` |
| `SetAggro(user.UserId, 0, t)` / `SetAggro(0, mob.Character.Id, t)` | `SetAggro(user.Character.Id, t)` / `SetAggro(mob.Character.Id, t)` |
| `Charmed.UserId` used as a body (aggro clear, follow, assist, `IsCharmed(user.UserId)`) | `Charmed.By` / `IsCharmed(user.Character.Id)` |
| `Charmed.UserId` used as an account (`TrackPlayerDamage`, `SetTempData` on the account, `TrackCharmed` on the charmer's Character) | resolve the account: `actors.Get(c.Charmed.By)`, `.Controller().Account()`, then its `UserId` / its Character |
| `KillerMobInstanceId` | `KilledBy` |

`rooms.FindFightingPlayer`/`FindFightingMob` (`rooms.go:1114-1123, 1216-1225`) keep their meaning
by asking the target's kind as in the first row.

- [ ] **Step 3: Test.** Add to `internal/characters/character_test.go`:

```go
func TestAggroHoldsOneTarget(t *testing.T) {
	c := New()
	target := ids.New(ids.NoCell)
	c.SetAggro(target, DefaultAttack)
	require.NotNil(t, c.Aggro)
	assert.Equal(t, target, c.Aggro.Target)
	assert.True(t, c.IsAggro(target))
	assert.False(t, c.IsAggro(ids.New(ids.NoCell)))
}

func TestCharmRemembersTheCharmersBody(t *testing.T) {
	c := New()
	by := ids.New(ids.NoCell)
	c.Charm(by, 5, "")
	assert.Equal(t, by, c.GetCharmedBy())
	assert.True(t, c.IsCharmed(by))
	assert.False(t, c.IsCharmed(ids.New(ids.NoCell)))
}
```

- [ ] **Step 4: Verify** as Task 3 Step 3, plus `grep -rn 'Aggro\.UserId\|Aggro\.MobInstanceId\|Charmed\.UserId\|KillerMobInstanceId' --include=*.go .` prints nothing.

- [ ] **Step 5: Commit** (`Aggro, charm and killer hold one body ID`).

---

### Task 5: Actors key on the body ID

**Files:** `internal/actors/actor.go`, `registry.go`, `wire.go`, `registry_test.go`, `wire_test.go`, `internal/usercommands/suicide.go`

- [ ] **Step 1: Change the registry.**
  - `type ActorId = ids.ID` (alias). `Actor.Id` is the body's `Character.Id`.
  - `Spawn(char, ctrl)` uses `char.Id`; if it is zero, log `mudlog.Error("actor", "event", "no body id", ...)` and return nil (a body without an ID is a bug: Task 2 and Task 3 give every body one). `mobSpawned`, `accountOnline` and `bindAccount` handle the nil (log, no actor) instead of dereferencing it.
  - Delete `lastId`, `byMobInstanceId`, `ByMobInstanceId`, `unbindMob`'s index use: `mobDestroyed(id)` removes `byId[id]` if it is an NPC actor. `index`/`unindex` keep only `byUserId`.
  - `bindAccount` keys on `u.Character.Id`. If an account comes online whose body ID differs from its existing actor's (only the permadeath reroll does this), rebind replaces the actor under the new ID.
  - Add `BodyReplaced(u *accounts.Account, old ActorId)`: under one lock, remove the actor with ID `old` if `u` drives it, then bind `u`'s new body; log `event=replaced`.
  - `usercommands/suicide.go` permadeath: after the Task 2 line, `actors.BodyReplaced(user, oldId)` where `oldId` was read before the reroll.

- [ ] **Step 2: Tests.** Convert `registry_test.go` and `wire_test.go` to `ids.ID` (bodies made with `ids.New(ids.NoCell)` on the character before spawning). Keep every existing assertion's meaning; replace `ByMobInstanceId(x)` with `Get(x)`. Add:

```go
func TestNPCActorIdIsItsBodyId(t *testing.T) {
	Reset()
	m := &mobs.Mob{}
	m.Character.Id = ids.New(ids.NoCell)
	mobs.OnInstanceCreated.Fire(m)
	a := Get(m.Character.Id)
	require.NotNil(t, a)
	assert.Equal(t, m.Character.Id, a.Id)
}

func TestPlayerKeepsItsActorIdAcrossLogins(t *testing.T) {
	Reset()
	body := ids.New(ids.NoCell)
	first := &accounts.Account{UserId: 91, Character: characters.New()}
	first.Character.Id = body
	accounts.OnActivate.Fire(first)
	accounts.OnDeactivate.Fire(first)
	again := &accounts.Account{UserId: 91, Character: characters.New()} // the next login loads the saved file
	again.Character.Id = body
	accounts.OnActivate.Fire(again)
	assert.Equal(t, body, ByUserId(91).Id)
}

func TestBodyReplacedRekeysThePlayer(t *testing.T) {
	Reset()
	u := &accounts.Account{UserId: 92, Character: characters.New()}
	u.Character.Id = ids.New(ids.NoCell)
	accounts.OnActivate.Fire(u)
	old := u.Character.Id
	u.Character.Id = ids.New(ids.NoCell)
	BodyReplaced(u, old)
	assert.Nil(t, Get(old))
	assert.Same(t, u, Get(u.Character.Id).Controller().Account())
	assert.Len(t, All(), 1)
}

func TestSpawnRefusesABodyWithoutAnId(t *testing.T) {
	Reset()
	assert.Nil(t, Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 93})))
	assert.Empty(t, All())
}
```

- [ ] **Step 3: Verify** (`go build ./...`, `go test ./internal/actors/ ./internal/accounts/ -count=1`, smoke PASS) and **commit** (`actors: the actor ID is the body's ID`).

---

### Task 6: Events name bodies by ID

**Files:** `internal/events/eventtypes.go` and every producer and consumer in inventory Table B;
`_datafiles/html/admin/docs-modules.html:1124-1228`.

- [ ] **Step 1: Change the event types.**

| Event | Before | After |
|---|---|---|
| `Buff`, `BuffsTriggered`, `BuffsAdded`, `BuffsRemoved`, `RoomChange`, `EquipmentChange`, `ItemOwnership`, `AggroChanged` | `UserId int`, `MobInstanceId` | `ActorId ids.ID` (the body) |
| `Input` | `UserId int`, `MobInstanceId` | `ActorId ids.ID` |
| `MobIdle` | `MobInstanceId` | `ActorId ids.ID` |
| `MobDeath` | `InstanceId` | `ActorId ids.ID` (keep `MobId`, the template) |
| `Communication` | `SourceUserId int`, `SourceMobInstanceId` | `SourceActorId ids.ID` (keep `TargetUserId`: the whisper recipient's connection) |
| `RoomAction`, `SkillUsed`, `CLIRequest`, `Purchase`, `Party` | dead types | deleted |
| every other event | unchanged (account or connection fields) | |

- [ ] **Step 2: Producers and consumers.** Producers set `ActorId` from the body
(`user.Character.Id`, `mob.Character.Id`). Consumers that need the account or connection resolve it:
`actors.Get(evt.ActorId)`, then `.Controller().Account()`, skipping when nil (an NPC has no
connection to send GMCP to). Consumers that branch on kind (`RoomChange` in `gmcp.Room.go`,
`CleanupEphemeralRooms`, `gmcp.Party.go`, `AggroChanged` in `gmcp.Room.go`) ask the actor the same
way. `world.go`'s input loop resolves `Input.ActorId` to the actor: an account means the player
path with that account's `UserId` (input blocking and the per-turn tracker stay keyed by `UserId`);
otherwise the mob path with the mob's ID. `world.go`'s `InputWorker` builds `Input` from the
connection's user: `actors.ByUserId(userId).Id` (log and drop the line if the player has no actor).
`Account.Command`/`CommandFlagged` and `Mob.Command` set `ActorId` from their character's ID.

`modules/follow/follow.go`: `followId` keys and the `MobDeath` handler use `evt.ActorId`. This fixes
the pre-existing bug where it used the template `evt.MobId` as an instance ID.

- [ ] **Step 3: Docs page.** In `docs-modules.html`, update the documented fields of `MobIdle`,
`RoomChange`, `AggroChanged`, `Buff*`, `Input` and `MobDeath` to `ActorId` (an `ids.ID`, written
`0x<cell>:0x<random>`), and remove the deleted types if listed.

- [ ] **Step 4: Verify** as Task 3 Step 3, plus `grep -rn 'MobInstanceId' --include=*.go .` lists only
test files. Commit (`Events name bodies by ID`).

---

### Task 7: Rooms hold one body list; `FindByName` and GMCP use ID tokens

**Files:** `internal/rooms/rooms.go`, `roommanager.go`, `roomdetails.go`, `save_and_load.go`,
`internal/accounts/userrecord.go` (`ShorthandId`), `internal/usercommands/attack.go:179,227`,
`modules/gmcp/gmcp.Room.go:348,376`, `gmcp.Char.go:452`, and compile errors that follow.

- [ ] **Step 1: One list.** Replace `players []int` and `mobs []int` with `bodies []ids.ID`.
  - `AddPlayer(userId)`/`RemovePlayer(userId)` keep their signatures and store or remove the account's `Character.Id`.
  - `AddMob(id ids.ID)`/`RemoveMob(id ids.ID)`; `Prepare`'s direct append and the room-replace copy use `bodies`.
  - `GetPlayers(flags...) []int` returns the `UserId` of each body whose actor has an account, filtered as today. `GetMobs(flags...) []ids.ID` returns each body whose actor has no account. Add `GetBodies() []ids.ID`.
  - `PlayerCt()`/`MobCt()` count the two kinds. `roomsWithUsers`/`roomsWithMobs` keep counting per kind, as today.
  - Fix `GetDetails`' mutation of the list while ranging over it (collect, then remove).
- [ ] **Step 2: Visitors.** Replace `visitors map[VisitorType]map[int]uint64` with
  `userVisitors map[int]uint64` (account, so tracking a player by name works across logins) and
  `mobVisitors map[ids.ID]uint64`. `MarkVisited` splits into `MarkUserVisited(userId, ...)` and
  `MarkMobVisited(id, ...)`; `Visitors(vType)` into `UserVisitors()` and `MobVisitors()`.
  `findUserExit` reads `UserVisitors()`. This fixes the pre-existing bug where it resolved mob
  visitor IDs as user IDs.
- [ ] **Step 3: `FindByName`.** Delete the `#N` and `@N` branches. When `ids.Parse(searchName)`
  succeeds, look up the actor: an account returns that account's `UserId` as the player half;
  otherwise the mob half. Every name form is unchanged.
- [ ] **Step 4: IDs out.** `Account.ShorthandId` is deleted. GMCP `Npcs[].id`, `Players[].id` and
  `Char.Enemies[].id` are the body's `Character.Id.String()`. Party auto-attack sends
  `fmt.Sprintf("attack %s", target.Character.Id)` for both kinds (`attack.go:179,227`).
  `window-room.js` needs no change: it echoes `id` as an opaque string.
- [ ] **Step 5: Tests** (`internal/rooms/rooms_test.go`, new or appended): with `actors.Init()` and
  `mudlog.SetupLogger(nil, "", "", false)` in `TestMain`, spawn one account actor and one mob actor
  (bodies with `ids.New(ids.NoCell)`), add both to a `NewEmptyRoom`, and assert `GetPlayers()` is
  `[]int{userId}`, `GetMobs()` is `[]ids.ID{mobId}`, `PlayerCt()==1`, and that `FindByName` with
  each ID's string returns the right half, with `"#1"` and `"@1"` finding nothing.
- [ ] **Step 6: Verify** as Task 3 Step 3, plus:

```bash
grep -rn '`#%d`\|"#%d\|`@%d`\|"@%d' --include=*.go . | grep -v _test   # nothing
```

Commit (`Rooms hold one body list; IDs replace #N and @N`).

---

### Task 8: Guards for the ID rulings

**Files:** `internal/guards/guards_test.go`

- [ ] **Step 1: Add three guards.**

```go
// TestNoMobInstanceIds: one ID space means the old mob instance ID is gone,
// in names and fields alike.
func TestNoMobInstanceIds(t *testing.T) {
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && strings.Contains(id.Name, "MobInstanceId") {
				t.Errorf("%s: %s; a body is named by its ids.ID", path, id.Name)
			}
			return true
		})
	}
}

// TestIdsAreMadeOnlyAtCreation enforces decision 11: an ID is made once, when
// a thing is created or instantiated. Each allowed file names its creation site.
func TestIdsAreMadeOnlyAtCreation(t *testing.T) {
	allowed := map[string]string{
		"internal/accounts/migration.go":   "giveBodyId: a new character, or one saved before body IDs existed",
		"internal/mobs/mobs.go":            "addInstance: a new NPC instantiation",
		"internal/usercommands/suicide.go": "permadeath reroll: a new character",
		"cmd/newid/main.go":                "a builder creating a thing by hand",
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "New" {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "ids" {
				if _, ok := allowed[path]; !ok {
					t.Errorf("%s: makes an ids.ID outside a creation site", path)
				}
			}
			return true
		})
	}
}

// TestNoNumericShorthandIds: people name things by name; clients send ID
// tokens. The #N and @N forms are gone.
func TestNoNumericShorthandIds(t *testing.T) {
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING &&
				(strings.Contains(lit.Value, "#%d") || strings.Contains(lit.Value, "@%d")) {
				t.Errorf("%s: formats a #N or @N id", path)
			}
			return true
		})
	}
}
```

- [ ] **Step 2: Prove each can fail.** For each guard, add a probe in a non-test file (an identifier
named `probeMobInstanceId`; a call `ids.New(ids.NoCell)` in `main.go`; a string `"#%d"`), run
`go test ./internal/guards/ -count=1` to see it FAIL naming the probe's file, remove the probe with
Edit, see PASS.
- [ ] **Step 3: Commit** (`guards: one ID space; IDs made only at creation; no #N/@N`).

---

### Task 9: Copyover keeps a player's body ID (round trip on every platform)

**Files:** `internal/accounts/copyover_test.go`

- [ ] **Step 1:** extend `TestCopyoverRoundTripKeepsOnlineAccounts` so each account gets a body ID
(`giveBodyId`) before the save, and assert after `RestoreFrom` that `GetByUserId(61).Character.Id`
and `GetByUserId(62).Character.Id` equal the originals. (The test's `restoreLoad` fake returns a copy
of the stored account, like the user file does.)
- [ ] **Step 2:** run `go test ./internal/accounts/ -count=1`, expect PASS. Then prove it can fail:
temporarily make the fake clear `c.Character.Id` (a copy of the character with `Id` zeroed), see
FAIL, revert with Edit, see PASS.
- [ ] **Step 3: Commit** (`copyover: a player's body ID survives the round trip`).

---

### Task 10: Live evidence: a player's ID survives logout and relogin

No code. Run the smoke test with the server log and read the actor lines:

```bash
LOG="<session scratchpad>/slice2-server.log"
go run ./cmd/smoke -server-log "$LOG" 2>&1 | tail -1
grep 'msg="actor ' "$LOG" | grep 'kind=player'
```

Expected: PASS; the bot (userId 2) spawns twice (signup, relogin) with the **same** `actorId`
`0x0000000000000000:0x...`, and the admin (userId 1) with the ID now saved in
`_datafiles/world/empty/users/1.yaml` (the smoke test copies that world). Record the lines for the
PR. If the bot's two `actorId`s differ, stop: the body ID is not being saved or loaded.

---

### Task 11: Verify the slice, record it, open the PR

- [ ] **Step 1:** full suite, gofmt on changed files, `go run ./cmd/smoke`, `-bots 3`,
`-copyover=false`; expected as the rules say.
- [ ] **Step 2:** `PATCH_NOTES.md` (one line: the ID scheme, body IDs saved, `#N`/`@N` replaced by
ID tokens, the two bug fixes), `status.md` (slice 2 done; next is slice 2b's plan),
`geomud-verifying`'s package count (one more: `internal/ids`). Move this plan with `git mv` to
`docs/completed/` and repoint references.
- [ ] **Step 3:** commit, push, open the PR (body lists what landed, verification output, Task 10's
actor lines, and the player-visible changes below).
- [ ] **Step 4: CI evidence.** All checks pass and the Smoke Test log shows the copyover step ran
(`smoke: copyover step ran ...`) and `PASS`: copyover now restores players whose actors key on
saved body IDs.

## Player-visible changes in this slice

- The room window's click menus send ID tokens instead of `#12`/`@5`; they work as before.
- Typing `#12` or `@5` no longer finds anything; names work as before.
- Nothing else a player can see changes. NPCs still respawn on copyover (slice 2b keeps them).

## Self-review against the spec

| Spec item (slice 2 row, decisions 10-11) | Task |
|---|---|
| `cell:random` ID type in a leaf package | 1 |
| A body's ID lives on its `Character`; a player's assigned at creation and saved | 2 |
| An NPC's at instantiation; `instanceCounter` goes | 3 |
| Rooms, aggro, events, charm, killer, GMCP, the room window hold it | 3, 4, 6, 7 |
| Names replace `#N`/`@N` for people | 7 |
| Deletes: room lists, `Aggro.MobInstanceId`/`UserId`, `Input.MobInstanceId`, `instanceCounter`, `#N`/`@N` | 3, 4, 6, 7 |
| Done when: `MobInstanceId` absent; a guard allows new IDs only at creation sites | 8 |
| Done when: a player keeps their ID across logout, restart and copyover | 2 (saved; restart loads the file), 9 (copyover), 10 (logout and relogin, live) |
| Done when: smoke passes | every task, 11 |
