# Phase 4b slice 2b: NPC and companion persistence (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** an existing NPC survives everything except its death or an admin's despawn: copyover,
restart, room unload, boredom and charm expiry. It comes back with its ID, what it is, what it has
and who it is tangled with; spawn slots keep their NPC until it ends.

**Architecture:** one idea, the *resting NPC*: an NPC saved to disk and out of memory, belonging to
the room it was in. A room unload rests every NPC in the room; boredom and charm expiry rest one
NPC in its current room. Resting NPCs live in a per-room sidecar file beside the room's instance
save, are listed in an in-memory index (rebuilt from the sidecars at startup), and come back when
their room is next prepared for a player, the moment spawning would happen. Copyover and restart
need nothing extra: every room save writes its live NPCs into its sidecar, so after either, every
NPC is resting and returns on its room's next prepare. Endings (death, admin despawn) go to a saved
ledger that spawn slots read.

**Tech stack:** Go 1.27.3, `gopkg.in/yaml.v2`, testify.

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md` decisions 12-15, section 4 slice 2b.
Facts: `2026-10-27-phase-4b-2b-inventory.md` (read 51 files, 2026-10-27), cited below as (inv §N).

---

## Decisions this plan applies (Calabe, 2026-10-27)

- 12: an NPC survives anything that is not its ending, saved with the room it is in.
- 13: kept: template, level, elite flag, home room, applied spawn overrides (including spawn
  permabuffs), health and mana, inventory, gold, equipment, buffs, adjectives, aggro target, who
  charmed it. Reset: path, wander count, boredom, last idle command, temporary data. Restored as
  saved, whatever the elapsed time.
- 14: a spawn slot keeps its NPC until it ends; slots save their NPC and its ending; an index says
  where each saved NPC is; a slot whose NPC is not live, not resting and not ended logs an error and
  respawns.
- 15: leaving memory is never an ending. Boredom and charm expiry rest an NPC. Only death and an
  admin's despawn end one.

Defaults set by the controller from those rulings:

- An admin's despawn keeps today's zero cooldown (its slot respawns at once); a death keeps the
  slot's respawn rate. Both are parity with today.
- Item UUIDs are not saved yet (`items.Item.UUID` is `yaml:"-"`): a restored NPC's items get new
  UUIDs. Slice 2c gives items saved IDs.
- Ephemeral rooms are not saved today, so their NPCs stay transient.
- A spawn slot's saved state is matched to the template's slot by list index and `MobId`; if a
  builder changed the slot, the saved state is dropped with a log line and the slot starts fresh.
- A restored NPC whose home room no longer has its slot keeps its home room and lives on as an
  NPC with no slot (like an admin-spawned one).
- Restored aggro is kept; `IdleMobs` already drops aggro on a target who is gone.
- Companions: `Character.CharmedMobs` is saved with the player, and the login prune drops only IDs
  that are neither live nor resting. Nothing charms a mob today, so companion persistence is proven
  by tests only.

## Facts verified against source (2026-10-27, master `9f7666aa`)

| Fact | Value | Where |
|---|---|---|
| Room instance save | Reflection diff of exported, non-`yaml:"-"`, non-`instance:"skip"` Room fields vs the template, to `<DataFiles>/rooms.instances/<zone folder>/<RoomId>.yaml`; empty diff deletes the file | `rooms/save_and_load.go:242-310` (inv §1) |
| Room saves happen at | shutdown `world.go:358`, autosave `hooks/NewTurn_AutoSave.go:53`, copyover `copyover.go:20`, unload `roommanager.go:517`, template edit `save_and_load.go:236` | (inv §1) |
| Room load | `LoadRoom` (`save_and_load.go:79-98`): cache, else `LoadRoomInstance` then `addRoomToMemory`; no load hook | (inv §1) |
| Prepare | spawning happens in `Room.Prepare` (`rooms.go:522-708`), called on player arrival (`MoveToRoom`), `HandleRespawns`, `look`, admin | (inv §4) |
| Not saved today | `Mob.HomeRoomId`, `IsElite`, `Character.Charmed`, `Aggro`, `CharmedMobs` (`yaml:"-"`), `Character.permaBuffIds` (unexported), `SpawnInfo.InstanceId`/`DespawnedRound` (`yaml:"-"`, and `SpawnInfo` is `instance:"skip"`) | (inv §2, §4) |
| Template save | `Mob.Save()` writes template YAML (`mobs.go:593-610`), so mob tags cannot change without leaking runtime fields into templates | (inv §9.4) |
| Mob instance entry | `addInstance` (`mobs.go:42-48`) is the only add path and always makes a new ID; the guard allows `ids.New` in `mobs.go` only for a new instantiation | (inv §2c, §9.2) |
| Mob removal | `DestroyInstance` (`mobs.go:257-261`) is the only removal; callers: `mobcommands/suicide.go` (3, death), `mobcommands/despawn.go` (1), `roommanager.go` unload (2) | (inv §3) |
| Despawn | `despawn` is a mob command; boredom queues `despawn depression ...` (`hooks/NewRound_IdleMobs.go:44-47`, only above `MobUnloadThreshold` 100 live mobs and `MaxMobBoredom` 250); charm expiry runs `CharmExpiredDespawn` (`characters/charminfo.go:7`, `...;despawn charmed mob expired`); admins reach it with `command`; `CleanupMobSpawns(true)` gives zero cooldown | (ctl, inv §3) |
| Unload | `removeRoomFromMemory` (`roommanager.go:489-520`): refuses rooms with players, destroys every mob listed, saves the instance, drops the room; only above `RoomUnloadThreshold` 200 loaded rooms after `RoomUnloadRounds` 450 | (inv §5) |
| Existing duplicate bug | a mob that wandered away outlives its home room's unload; on reload the slot looks empty and spawns a second copy | (inv §5) |
| Spawn permabuffs | `Room.Prepare` sets the slot's `BuffIds` via `SetPermaBuffs` (replaces the template list) and then calls `mob.Validate()` without `true` (`rooms.go:597-599, 620`); `Validate(true)` strips permabuffs not in `permaBuffIds` | (inv §9.3) |
| Round counter | persists across restart (`util.LoadRoundCount`) and copyover, so saved rounds stay meaningful | (inv §9.9) |
| Test world | 4 rooms, 4 mobs: below both unload thresholds, so smoke never unloads a room without a config override | (inv §5, §9.12) |
| Copyover order | `triggerCopyover`: `SaveAllRooms`, `SaveAllUsers`, `plugins.Save`, `copyover.Execute`; the new process loads rooms lazily and players re-enter through `MoveToRoom`, which prepares their room and adjacent rooms | (inv §6) |
| Gitignore | `**/rooms.instances` is ignored | `.gitignore:11` (inv §8) |

## Rules for every task

- Edit files with the Edit and Write tools only. No `sed -i`, no shell redirects into the repo, no
  Python write-opens, no `gofmt -w`/`gopls -w` in the repo (landing gofmt output from the scratchpad
  with `cp` is allowed). `~/.claude/hooks/guard.py` enforces it.
- `git add <named paths>` only; commit messages through a quoted heredoc ending with
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Every task ends with `go build ./...`, its targeted tests, and gofmt clean on changed files.
  Tasks 6, 8 and 10 also run the full suite (`go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'`;
  only the 3 Windows failures in `geomud-verifying` allowed) and `go run ./cmd/smoke` and `-bots 3`.
- `<session scratchpad>` means the scratchpad directory in your system prompt, as a full path.
- TDD where a task gives tests: see them fail first. A step whose expected output does not match:
  stop and report. A site the task does not cover: stop and report with file:line.
- `internal/guards` must keep passing (no `ids.New` outside the creation sites; restore uses saved IDs).

## File structure

| File | Responsibility |
|---|---|
| `internal/mobs/saved.go` (new) | `SavedNPC`: the saved form of a live NPC; `Snapshot`, `Restore`, `Rest` |
| `internal/mobs/testhelpers.go` (new) | `NewTestInstance` for other packages' tests |
| `internal/characters/character.go` | `PermaBuffIds()` getter; `CharmedMobs` saved |
| `internal/rooms/npcstore.go` (new) | The sidecar file, the resting index, the endings ledger, slot state |
| `internal/rooms/rooms.go`, `roommanager.go`, `save_and_load.go` | Resting on unload, restoring on prepare, slot decisions, sidecar saves |
| `internal/mobcommands/rest.go` (new), `despawn.go`, `suicide.go` | `rest` command; endings recorded |
| `internal/hooks/NewRound_IdleMobs.go`, `internal/characters/charminfo.go` | Boredom and charm expiry rest instead of despawn |
| `internal/accounts/users.go` | Companion prune keeps resting NPCs |
| `main.go` | Build the resting index and load the ledger at startup |
| `cmd/smoke/copyover.go` | Copyover step checks NPC IDs survive (Linux CI) |

---

### Task 0: Branch

- [ ] The rulings and this plan merge to `master` as their own PR; then
  `git switch master && git pull --ff-only origin master && git switch -c strip/phase-4b-2b-npc-persistence`.

---

### Task 1: The saved form of an NPC

**Files:** Create `internal/mobs/saved.go`, `internal/mobs/saved_test.go`; modify `internal/characters/character.go` (getter).

- [ ] **Step 1: Write the failing test** (`internal/mobs/saved_test.go`)

```go
package mobs

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestSnapshotRoundTripsWhatDecision13Keeps(t *testing.T) {
	m := Mob{MobId: 7, HomeRoomId: 2, IsElite: true, MaxWander: 3, Hostile: true,
		IdleCommands: []string{"emote sniffs"}, WanderCount: 5, BoredomCounter: 9, LastIdleCommand: 2}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.Name = "rat"
	m.Character.Level = 4
	m.Character.Health = 3
	m.Character.Gold = 11
	m.Character.SetPermaBuffs([]int{5, 6})
	target := ids.New(ids.NoCell)
	m.Character.Aggro = &characters.Aggro{Target: target}
	by := ids.New(ids.NoCell)
	m.Character.Charmed = characters.NewCharm(by, 12, "")

	b, err := yaml.Marshal(Snapshot(&m))
	require.NoError(t, err)
	var back SavedNPC
	require.NoError(t, yaml.Unmarshal(b, &back))

	got := back.Mob()
	assert.Equal(t, m.Character.Id, got.Character.Id)
	assert.Equal(t, MobId(7), got.MobId)
	assert.Equal(t, 2, got.HomeRoomId)
	assert.True(t, got.IsElite)
	assert.Equal(t, 3, got.MaxWander)
	assert.True(t, got.Hostile)
	assert.Equal(t, []string{"emote sniffs"}, got.IdleCommands)
	assert.Equal(t, "rat", got.Character.Name)
	assert.Equal(t, 4, got.Character.Level)
	assert.Equal(t, 3, got.Character.Health)
	assert.Equal(t, 11, got.Character.Gold)
	assert.Equal(t, []int{5, 6}, got.Character.PermaBuffIds())
	require.NotNil(t, got.Character.Aggro)
	assert.Equal(t, target, got.Character.Aggro.Target)
	require.NotNil(t, got.Character.Charmed)
	assert.Equal(t, by, got.Character.Charmed.By)
	// what it was doing resets
	assert.Zero(t, got.WanderCount)
	assert.Zero(t, got.BoredomCounter)
	assert.Zero(t, got.LastIdleCommand)
}
```

- [ ] **Step 2:** run `go test ./internal/mobs/ -run Snapshot -count=1`; expect a build failure.

- [ ] **Step 3: Implement.** In `characters/character.go` add beside `SetPermaBuffs`:

```go
// PermaBuffIds is the spawn or template permabuff list SetPermaBuffs set.
func (c *Character) PermaBuffIds() []int { return append([]int{}, c.permaBuffIds...) }
```

`internal/mobs/saved.go`:

```go
package mobs

import "github.com/GoMudEngine/GoMud/internal/characters"

// SavedNPC is a live NPC as saved to disk (Phase 4b decision 13): what it is,
// has and is tangled with. It is separate from Mob so template saves
// (Mob.Save) never gain runtime fields. What the NPC was doing (path, wander
// count, boredom, last idle command, temp data) is not saved.
type SavedNPC struct {
	Template     Mob                   `yaml:"mob"`                    // the instance's exported fields, including Character
	HomeRoomId   int                   `yaml:"homeroomid"`
	IsElite      bool                  `yaml:"elite,omitempty"`
	PermaBuffIds []int                 `yaml:"permabuffids,omitempty"` // spawn overrides replace the template list
	Aggro        *characters.Aggro     `yaml:"aggro,omitempty"`
	Charmed      *characters.CharmInfo `yaml:"charmed,omitempty"`
}

// Snapshot saves m as it is now.
func Snapshot(m *Mob) SavedNPC {
	return SavedNPC{
		Template:     *m,
		HomeRoomId:   m.HomeRoomId,
		IsElite:      m.IsElite,
		PermaBuffIds: m.Character.PermaBuffIds(),
		Aggro:        m.Character.Aggro,
		Charmed:      m.Character.Charmed,
	}
}

// Mob rebuilds the NPC from its saved form, with what it was doing reset.
func (s SavedNPC) Mob() Mob {
	m := s.Template
	m.HomeRoomId = s.HomeRoomId
	m.IsElite = s.IsElite
	m.Character.SetPermaBuffs(s.PermaBuffIds)
	m.Character.Aggro = s.Aggro
	m.Character.Charmed = s.Charmed
	m.WanderCount, m.BoredomCounter, m.LastIdleCommand = 0, 0, 0
	m.Path = PathQueue{}
	return m
}
```

(If `Mob` has fields named differently from those used here, read `mobs.go:57-81` and use the real
names; if `PathQueue{}` is not a valid zero value, use the reset the type provides. Keep the
test's assertions.)

- [ ] **Step 4:** test passes; also run `go test ./internal/mobs/ ./internal/characters/ -count=1`.
- [ ] **Step 5: Commit** `mobs: SavedNPC, the saved form of a live NPC`.

---

### Task 2: Restore and rest in the mob registry

**Files:** `internal/mobs/saved.go`, `internal/mobs/testhelpers.go` (new), `internal/mobs/saved_test.go`.

- [ ] **Step 1: Tests** (append to `saved_test.go`):

```go
func TestRestoreKeepsTheIdAndFires(t *testing.T) {
	var created []*Mob
	OnInstanceCreated.Register(func(m *Mob) *Mob { created = append(created, m); return m })
	m := Mob{MobId: 8}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	r, err := Restore(Snapshot(&m))
	require.NoError(t, err)
	assert.Equal(t, m.Character.Id, r.Character.Id)
	assert.Same(t, r, GetInstance(m.Character.Id))
	assert.Contains(t, created, r)
	_, err = Restore(Snapshot(&m))
	assert.Error(t, err, "restoring an NPC that is already live is refused")
	Rest(m.Character.Id)
	assert.Nil(t, GetInstance(m.Character.Id))
}

func TestRestoreRefusesAZeroId(t *testing.T) {
	m := Mob{MobId: 9}
	m.Character = *characters.New()
	_, err := Restore(Snapshot(&m))
	assert.Error(t, err)
}
```

- [ ] **Step 2: Implement** in `saved.go`:

```go
// Restore puts a saved NPC back in memory under its saved ID and fires
// OnInstanceCreated. It never makes an ID (decision 11).
func Restore(s SavedNPC) (*Mob, error) {
	m := s.Mob()
	if m.Character.Id.IsZero() {
		return nil, fmt.Errorf("mobs: saved NPC %d has no body ID", m.MobId)
	}
	if _, live := mobInstances[m.Character.Id]; live {
		return nil, fmt.Errorf("mobs: NPC %s is already live", m.Character.Id)
	}
	m.Validate()
	m.Character.Validate(true) // permaBuffIds is set, so permabuffs survive
	mobInstances[m.Character.Id] = &m
	stored := mobInstances[m.Character.Id]
	OnInstanceCreated.Fire(stored)
	return stored, nil
}

// Rest takes a live NPC out of memory without ending it (decision 15). The
// caller has saved it first. Actors drop it through OnInstanceDestroyed.
func Rest(id ids.ID) {
	if _, ok := mobInstances[id]; !ok {
		return
	}
	delete(mobInstances, id)
	OnInstanceDestroyed.Fire(id)
}
```

`testhelpers.go`: `// NewTestInstance adds m as a new instantiation, for other packages' tests.` `func NewTestInstance(m Mob) *Mob { return addInstance(m) }`.

- [ ] **Step 3:** tests pass; guards pass (`go test ./internal/guards/ -count=1`). Commit `mobs: Restore and Rest`.

---

### Task 3: The NPC store (sidecars, resting index, endings ledger)

**Files:** Create `internal/rooms/npcstore.go`, `internal/rooms/npcstore_test.go`.

The store owns three things:

```go
// roomNPCFile is a room's sidecar, rooms.instances/<zone>/<id>.npcs.yaml:
// the NPCs resting in the room and the room's spawn slot state.
type roomNPCFile struct {
	Resting []mobs.SavedNPC `yaml:"resting,omitempty"`
	Slots   []slotState     `yaml:"slots,omitempty"`
}

// slotState is one spawn slot's memory: which NPC it owns. Matched to the
// template's SpawnInfo by Index and MobId.
type slotState struct {
	Index int    `yaml:"index"`
	MobId int    `yaml:"mobid"`
	NPC   ids.ID `yaml:"npc"`
}

// ending is an NPC that ended (decision 15): its round, and whether its slot
// may respawn at once (an admin's despawn) or waits its respawn rate (death).
type ending struct {
	Round     uint64 `yaml:"round"`
	Immediate bool   `yaml:"immediate,omitempty"`
}
```

and the API (all on the world goroutine; guard with one `sync.Mutex` because autosave runs there
too and copyover saves from a command):

| Function | Does |
|---|---|
| `restingIndex map[ids.ID]int` | resting NPC ID to its room ID |
| `BuildRestingIndex() error` | scan every `*.npcs.yaml` under `rooms.instances`, fill the index; log and skip a file that does not parse |
| `LoadEndings() error` / `saveEndings() error` | the ledger at `rooms.instances/npcs.endings.yaml`, `map[ids.ID]ending` |
| `RecordEnding(id ids.ID, round uint64, immediate bool)` | add to the ledger and save it |
| `takeEnding(id) (ending, bool)` | read and remove (a slot consumes it) |
| `readRoomNPCFile(r *Room) roomNPCFile` / `writeRoomNPCFile(r *Room, f roomNPCFile) error` | path from the room's zone and ID as `SaveRoomInstance` builds it; write deletes the file when it is empty |
| `IsResting(id ids.ID) bool` | in the index |

- [ ] **Step 1: Tests** (`npcstore_test.go`), using a temp directory for the data path. Read
`save_and_load.go:291-293` for how the instance path is built and give the store a package
variable `npcStoreRoot = func() string { ... }` that tests replace with `t.TempDir()`. Test: write
a room file with two resting NPCs and a slot, `BuildRestingIndex` finds both with the right room;
`RecordEnding` then `LoadEndings` in a fresh map round-trips; `takeEnding` removes; an empty room
file deletes the file; YAML of `ids.ID` inside `slotState`, `SavedNPC.Aggro` and as a map key
round-trips (inventory open question 1).
- [ ] **Step 2:** implement; tests pass. Commit `rooms: the NPC store (sidecars, resting index, endings)`.

---

### Task 4: Rooms rest and restore NPCs

**Files:** `internal/rooms/rooms.go`, `roommanager.go`, `save_and_load.go`, `npcstore.go`, tests.

- [ ] **Step 1: Room state.** Add unexported `resting []mobs.SavedNPC` to `Room`. `LoadRoom`,
between `LoadRoomInstance` and `addRoomToMemory`, reads the sidecar: `resting` from it, and slot
state merged into `SpawnInfo` (for each saved slot, if `Index < len(SpawnInfo)` and
`SpawnInfo[Index].MobId == MobId`, set `SpawnInfo[Index].InstanceId = NPC`; else log
`npc store: slot state dropped` with the room and index).
- [ ] **Step 2: Saving.** `SaveRoomInstance` also writes the sidecar: `Resting` = the room's
`resting` plus a `mobs.Snapshot` of every live NPC body in the room (`GetMobs()` that resolve with
`mobs.GetInstance`), and `Slots` = every slot with a non-zero `InstanceId`. This makes copyover,
autosave and shutdown save live NPCs with no other change. Skip ephemeral rooms (as the instance
save does). Template edits (`SaveRoomTemplate`, `save_and_load.go:186-236`) copy `resting` and the
slot `InstanceId`s from the old room to the new one before saving.
- [ ] **Step 3: Unload rests.** In `removeRoomFromMemory`, replace the two destroy loops: for each
live NPC body in the room, append `mobs.Snapshot(m)` to `resting`, then save the instance (which
writes the sidecar), then `mobs.Rest(id)` each and add each ID to the resting index with this room.
No ending is recorded.
- [ ] **Step 4: Prepare restores.** At the top of `Room.Prepare`, before the spawn loop: for each
`resting` NPC, `mobs.Restore`; on success append its ID to `bodies` (as `Prepare` does for
spawns) and remove it from the resting index; on error log and drop it. Clear `resting`, and write
the sidecar so the restored NPCs are no longer on disk as resting.
- [ ] **Step 5: Tests** (`internal/rooms/npcpersist_test.go`): with the temp store root, a room in
memory with one live NPC (made with `mobs.NewTestInstance`, added with `AddMob` or by appending to
`bodies`): unload rests it (not in `mobs.GetInstance`, in the index, in the sidecar); loading and
preparing the room restores it with the same ID, health and gold; a second prepare does not restore
it twice.
- [ ] **Step 6:** full suite and smoke. Commit `rooms: unload rests NPCs; prepare restores them`.

---

### Task 5: Spawn slots keep their NPC until it ends

**Files:** `internal/rooms/rooms.go` (`Prepare`, `CleanupMobSpawns`, `RoundTick`), `npcstore.go`, tests.

- [ ] **Step 1: One decision function** in `npcstore.go`:

```go
// slotNPC says what a slot's NPC is doing (decision 14): live anywhere,
// resting anywhere, or ended; an NPC that is none of these is lost.
func slotNPC(id ids.ID) (state string, end ending) {
	if mobs.GetInstance(id) != nil {
		return "live", ending{}
	}
	if IsResting(id) {
		return "resting", ending{}
	}
	if e, ok := takeEnding(id); ok {
		return "ended", e
	}
	return "lost", ending{}
}
```

- [ ] **Step 2: Use it** at the three sites that today treat "not live" as "gone"
(`Prepare` 541-550, `CleanupMobSpawns` 719-733, `RoundTick` 2062-2067):
  - `live`, `resting`: the slot stays filled; do nothing.
  - `ended`: clear `InstanceId`; `DespawnedRound = 0` if `Immediate`, else the ending's `Round`.
  - `lost`: `mudlog.Error("npc store: slot lost its NPC", "room", r.RoomId, "npc", id)`; clear and
    set `DespawnedRound = roundNow` (respawn after the rate, as a vanished NPC does today).
  `CleanupMobSpawns(noCooldown)` keeps its parameter only if a caller still needs it after Task 6;
  remove it if not.
- [ ] **Step 3: Tests:** a slot whose NPC is resting in another room stays filled through
`Prepare`; one whose NPC ended by death waits the respawn rate; by admin despawn respawns at once;
a lost one logs and respawns after the rate. Use the temp store root.
- [ ] **Step 4:** commit `rooms: spawn slots keep their NPC until it ends`.

---

### Task 6: Endings and rests at their sources

**Files:** `internal/mobcommands/suicide.go`, `despawn.go`, `rest.go` (new), `mobcommands.go`,
`internal/hooks/NewRound_IdleMobs.go:44-47`, `internal/characters/charminfo.go:7`.

- [ ] Death: in `suicide.go`'s death branches (before `DestroyInstance`), `rooms.RecordEnding(mob.Character.Id, util.GetRoundCount(), false)`. The vanish branch (no caller) records an immediate ending.
- [ ] Admin despawn: `despawn.go` records `RecordEnding(id, round, true)` before `DestroyInstance`.
- [ ] New mob command `rest` (`rest.go`, registered in `mobcommands.go`): logs the reason; finds the mob's current room; `room.RestNPC(mob)` (new in `rooms.go`: snapshot into `resting`, write the sidecar, `RemoveMob`, `mobs.Rest`, add to the resting index). No ending.
- [ ] Boredom: `NewRound_IdleMobs.go` queues `rest depression %d/%d` instead of `despawn ...`.
- [ ] Charm expiry: `CharmExpiredDespawn` ends in `;rest charmed mob expired` instead of `;despawn ...` (rename the constant to `CharmExpiredRest` and update its users).
- [ ] Test: `RestNPC` on a loaded room rests the NPC (gone from memory, in the index and sidecar), and the room's next `Prepare` restores it.
- [ ] Full suite and smoke. Commit `NPC endings recorded; boredom and charm expiry rest instead of ending`.

---

### Task 7: Startup

**Files:** `main.go`.

- [ ] After `loadAllDataFiles(false)` and before `copyover.Restore` (both the cold and copyover
paths), call `rooms.BuildRestingIndex()` and `rooms.LoadEndings()`, logging errors. Comment that
after a restart or copyover every NPC is resting and comes back on its room's next prepare.
- [ ] Commit `Startup builds the resting index and loads NPC endings`.

---

### Task 8: Companions

**Files:** `internal/characters/character.go` (`CharmedMobs` tag), `internal/accounts/users.go`
(the three login prune sites, inventory §7).

- [ ] `CharmedMobs []ids.ID` becomes `yaml:"charmedmobs,omitempty"` (saved with the player).
- [ ] `accounts` cannot import `rooms`. Give `accounts` a hook the startup wiring fills:
`var NPCExists = func(id ids.ID) bool { return mobs.MobInstanceExists(id) }`, and set it in
`main.go` after Task 7 to `func(id ids.ID) bool { return mobs.MobInstanceExists(id) || rooms.IsResting(id) }`.
The three prune sites use `NPCExists`.
- [ ] Test (`accounts`): an account whose `CharmedMobs` names a resting NPC keeps it through the
prune; one naming an unknown ID drops it. Save and load a user YAML and see `CharmedMobs` survive.
- [ ] Full suite and smoke. Commit `Companions are saved with their player`.

---

### Task 9: Copyover brings every NPC back with its ID (Linux CI)

**Files:** `cmd/smoke/copyover.go`, `cmd/smoke/copyover_test.go`.

- [ ] After the copyover step succeeds and every bot has checked its room, read the server log:
the set of NPC actor IDs with `event=spawned kind=npc` logged by the first process (before the
second `Starting server` line) must equal the set logged by the new process for rooms the bots are
in or next to. Simplest robust form: every NPC ID the new process logs must have been logged by the
old one (no fresh NPCs replaced restored ones), and at least one was. Fail the copyover step
otherwise, naming the IDs. Unit-test the log parsing with a fixture string.
- [ ] Windows skips the step as today; CI on the PR is the evidence. Commit
`smoke: copyover keeps every NPC's ID`.

---

### Task 10: Verify, record, PR

- [ ] Full suite, gofmt, smoke 1 bot, 3 bots, `-copyover=false`.
- [ ] Unload, live: run the server once with a `CONFIG_PATH` override that sets
`Memory.RoomUnloadThreshold: 0` and `Memory.RoomUnloadRounds: 2` (read `config.yaml` for the real
key names), walk a probe client from Town Square to End of the Line and wait for room 1 to
unload, then walk back and check in the server log that room 1's NPC came back with the same
`actorId` (`event=spawned kind=npc` lines). The smoke `server.go` override pattern shows how to
run an isolated server; stop it by its PID. Record the log lines.
- [ ] `PATCH_NOTES.md`, `status.md` (2b done; next is slice 2c or 3), `geomud-verifying` count if it
changed. Move this plan and the inventory to `docs/completed/`.
- [ ] Commit, push, PR; CI must show the copyover step ran and the NPC ID check passed.

## Player-visible changes

- A wounded or looted NPC is still wounded and still carries its loot after a copyover, restart or
  room unload. The same rat, the same ID.
- Spawn slots no longer duplicate a wandering NPC when its home room reloads.
- In a busy world, a bored NPC now rests and comes back when a player next enters its room,
  instead of being replaced.

## Self-review against the spec

| Ruling | Task |
|---|---|
| 12: survives copyover, restart, room unload; saved with its room | 4 (unload, save), 7 (startup), 9 (copyover evidence), 10 (unload live) |
| 13: kept vs reset fields | 1 |
| 14: slot keeps its NPC until it ends; index; lost logs and respawns | 3, 5 |
| 15: boredom and charm expiry rest; only death and admin despawn end | 6 |
| Companions keep their IDs | 8 |
| IDs never made on restore | 2 (guard) |
