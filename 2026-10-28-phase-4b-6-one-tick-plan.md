# Phase 4b slice 6: one tick and behaviour (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every actor, player or NPC, gets the same round upkeep and one round tick through `Controller.Tick`; the NPC behaviour (idle, wander, walking a path) runs from `BehaviourController.Tick`. The per-kind round ticks, `IdleMobs`, `HandleIdleMobs` and the `MobIdle` event go.

**Architecture:** Two passes over `actors.All()` (sorted by ID), at the two places in the round where today's per-kind code runs, so that nothing moves relative to combat: `ActorUpkeep` before combat (cooldowns, charm countdown, buff triggers, stat refresh for every actor; alignment decay for accounts; charm expiry and the health-0 death for NPCs) and `ActorTick` after healing (calls `Controller.Tick` for every actor). `BehaviourController.Tick` fires `actors.OnBehaviourTick`, which `internal/behaviour` fills at startup (`behaviour.Init`), because `actors` cannot import `behaviour` (cycle through `rooms`). The behaviour state stays on the `Mob`.

**Tech stack:** Go, `internal/events` listeners, `util.Hook`, testify.

Spec: `2026-10-27-phase-4b-npcs-are-players-design.md` (section 4 row 6, "Notes for later slices"). Inventory: `2026-10-28-phase-4b-6-inventory.md` (G round tick, H idle and behaviour, I world loop and blast radius). Branch `strip/phase-4b-6-one-tick`.

---

## Facts verified against source (2026-10-28, branch at `3d7f7e47`)

| Fact | Value | Where |
|---|---|---|
| NewRound listener order | `InactivePlayers`, `UpdateZoneMutators`, `CheckNewDay`, `UserRoundTick`, `MobRoundTick`, `HandleRespawns`, `DoCombat`, `AutoHeal`, `IdleMobs`; `MobIdle` → `HandleIdleMobs` | `internal/hooks/hooks.go:17-32` |
| Controller interface | `Account() *accounts.Account`, `Perceive(perception.Perception)`; no `Tick` | `internal/actors/controller.go:14-20` |
| Controller implementations | `ConnectionController`, `BehaviourController` (both `internal/actors/controller.go`), test fake `recorder` | `internal/messaging/messaging_test.go:20-26` |
| NPC perception hook pattern | `var OnBehaviourPerceive util.Hook[BehaviourPerception]`, fired by `BehaviourController.Perceive` | `controller.go:63-74` |
| `util.Hook` | `Register(func(T) T)`, `Fire(T) T`; holds its read lock while handlers run; no unregister | `internal/util/hook.go` (grep `type Hook\[`) |
| `actors.All` / `actors.Each` | `All` copies and sorts by ID text, no lock held after return; `Each` holds the read lock (a rest or death inside would deadlock on `unbindMob`) | `internal/actors/registry.go:93-116` |
| Import cycle | `behaviour` imports `rooms`; `rooms` imports `actors`; so `actors` cannot import `behaviour` | `internal/behaviour/behaviour.go:8-19`; inventory G §2 |
| `actors.Init` | called at `main.go:106`; tests call it in `TestMain` (`behaviour_test.go:53`, `actionpoints_test.go:23`, `rooms_test.go:42`, `usercommands_test.go:26`, `world_test.go:68`) | grep `actors.Init()` |
| `UserRoundTick` | per room with players: `room.RoundTick()`, room idle message (5 in 100, 20 in room -1); per player: cooldowns, alignment decay (+text, `AlignmentChanged`), charm countdown, buff triggers (`BuffsTriggered`), `Validate`, `CharacterVitalsChanged` on change | `internal/hooks/NewRound_UserRoundTick.go:19-132` |
| `MobRoundTick` | `mobs.ReduceHostility()`; per mob: cooldowns, charm countdown, buff triggers, charm expiry (`RemoveCharm`, `TrackCharmed`, `runCharmExpiredCommand`), `Validate`, `suicide` at health ≤ 0 | `internal/hooks/NewRound_MobRoundTick.go:54-111` |
| Charm-expiry drift | after `runCharmExpiredCommand` rests the NPC, the same pass still validates it and can queue `suicide` | `NewRound_MobRoundTick.go:92-108` |
| `runCharmExpiredCommand` tests | `internal/hooks/charmexpiry_test.go:70,84` call it by name | grep |
| `IdleMobs` | boredom unload branch (`:46-56`), `PreventIdle` one-round skip (`:59-62`), aggro: "losing quarry" (`:66-79`), path walking with waypoint `noop` 1 s (`:83-130`), then queues `MobIdle` (`:133`); `util.TrackTime("IdleMobs()")` (`:137`) | `internal/hooks/NewRound_IdleMobs.go` |
| `HandleIdleMobs` | MaxWander 0 displaced → `PathTo home`; charmed → stop; `MaxWander > -1 && WanderCount > MaxWander` → `PathTo home`; `util.Rand(100) < ActivityLevel` → `RunIdle(GetIdleCommand())` | `internal/hooks/MobIdle_HandleIdleMobs.go:14-53` |
| `Wander` limit | `MaxWander > -1 && WanderCount > MaxWander` → `PathTo home` (so `maxwander: 8` allows 9 moves) | `internal/behaviour/behaviour.go:31-37` |
| Stuck at home | `WanderCount > MaxWander` at home: `PathTo(home)` returns at once (`toRoomIds[0] == RoomId`, `behaviour.go:87`), count never resets (`IdleMobs:127-129` only on path end) | inventory H §4 |
| `PreventIdle` writer | `defMob.PreventIdle = true` when an NPC is attacked and not fighting | `internal/hooks/NewRound_DoCombat.go:234` |
| Follow veto | `events.RegisterListener(events.MobIdle{}, f.idleMobHandler, events.First)`; handler cancels when the NPC follows | `modules/follow/follow.go:68, 246-254` |
| `MobIdle` event | `type MobIdle struct{ ActorId ids.ID }` | `internal/events/eventtypes.go:135-140` |
| `MobIdle` docs | event table row and follow example and callout | `_datafiles/html/admin/docs-modules.html:1123, 1557-1563, 1583` |
| Boredom | `BoredomCounter` never incremented (only reset: `rooms.go:2066-2073`, `saved.go:58`); config `MaxMobBoredom`, `MobUnloadThreshold` | `internal/mobs/mobs.go:63`; `internal/configs/config.memory.go:5-6,13-19`; `_datafiles/config.yaml:94-102`; `config.html:211-212`; `config-wizard.html:1344-1357, 1652` |
| Dead fields | `LastIdleCommand` never read (`mobs.go:62`, `saved.go:58`); `Account.RoundTick` empty, no caller (`internal/accounts/userrecord.go:462-464`) | grep |
| Saved NPC test | `saved_test.go:32-33, 69-71` set and assert `BoredomCounter`, `LastIdleCommand` | grep |
| Config loading | `yaml.Unmarshal` (non-strict): a removed key left in a config file is ignored | `internal/configs/configs.go:98, 403, 435` |
| Admin wander prompt | "(-1 = none. 0 = unlimted)" is backwards: code treats -1 as unlimited, 0 as never | `internal/usercommands/admin.mob.go:364`; `behaviour.go:32` |
| The rat | `activitylevel: 10`, `maxwander: 8`, `idlecommands: ['emote wiggles its nose', 'wander', '']` | `_datafiles/world/empty/mobs/startland/1-rat.yaml` |
| Idle command pick | 1 in 100 empty, else uniform over `IdleCommands` | `internal/mobs/mobs.go:480-496` |
| RNG | idle path uses `util.Rand` only; `util.SeedRandForTest(seed)` | `internal/util/util.go:181-194` |
| Map-order randomness | `Room.GetRandomExit` builds a map (`rooms.go:868`): a room with two or more exits picks in map order | inventory H §5 |
| Mob YAML library | `gopkg.in/yaml.v2` | `internal/mobs/mobs.go:13` |
| Path queue | `Len`, `Current`, `Next`, `Clear`, `Waypoints`, `SetPath` | `internal/mobs/mobs_path.go` |
| Guards walk | `goFiles` skips `_test.go` and `_datafiles` | `internal/guards/guards_test.go:18-45` |
| `TestOneCombatPath` | pins `DoCombat` in `internal/hooks/NewRound_DoCombat.go`; this plan does not move it | `internal/guards/combat_test.go:43` |
| Input ordering | per kind on `Account()`: no account strict FIFO, account may overtake | `world.go:90-145` |
| Windows test baseline | 39 packages ok, 3 known failures | `geomud-verifying` skill |

## Decisions

Calabe (2026-10-28):

1. **Boredom unload is deleted**, with `BoredomCounter`, `MaxMobBoredom` and `MobUnloadThreshold`. Nothing players see changes (it has been dead since Phase 2). NPC population limits are for the geo work.
2. **Input ordering stays per kind** (NPC strict FIFO, account may overtake). NPCs chain commands through `mob.Command` and rely on strict order.

Plan defaults (no player-visible change unless stated):

3. **Tick placement keeps today's round order:** upkeep before combat, behaviour after healing. Two passes over all actors, both in sorted ID order (today: accounts in room order, then NPCs in map order; only the order of same-round lines can differ).
4. **Upkeep is not a controller method.** Character upkeep is the same for every actor; account-only (alignment decay, vitals event) and NPC-only (charm expiry, health-0 death) parts are keyed on `Account()` and `mobs.GetInstance`, as rules data. `Controller.Tick` is what the driver does by itself: nothing for a connection, behaviour for an NPC.
5. **`BehaviourController` stays in `actors`** and fires `actors.OnBehaviourTick`; `behaviour.Init()` registers `behaviour.Tick`. Moving the controller into `behaviour` would need `actors` tests to import `behaviour` (cycle).
6. **Behaviour state stays on the `Mob`** (`WanderCount`, `Path`, `PreventIdle`): combat (`NewRound_DoCombat.go:234`), follow (`follow.go:310`), admin `locate` and the saved-NPC form use it. The controller-state move in the spec comment is left for possession.
7. **The `MobIdle` veto becomes `behaviour.OnIdle`** (`util.Hook[IdleCheck]`, a handler sets `Skip`), fired at the point `MobIdle` was queued. Following NPCs still never idle.
8. **`maxwander` means wander moves since last home, and allows exactly that many** (8, not 9). The admin prompt is corrected.
9. **Stuck-at-home fixed:** an NPC at home with `WanderCount >= MaxWander` resets the count and idles.
10. **Charm-expiry drift fixed:** an NPC rested by charm expiry is not validated or killed in the same pass.
11. **Kept as today:** alignment decay is accounts-only (a Phase 5 rule); `Room.RoundTick` and room idle messages run only for rooms with players; wander state resets on copyover and restore; no NPC reactions are added (the pattern for later: queue through `mob.Command`, never send inside `OnBehaviourPerceive`).
12. **Dead code goes:** `LastIdleCommand`, `Account.RoundTick`.

## Files

- Modify `internal/actors/controller.go`: `Tick()` on the interface and both controllers; `OnBehaviourTick`.
- Modify `internal/actors/controller_test.go`, `internal/messaging/messaging_test.go` (fake gains `Tick`).
- Create `internal/behaviour/tick.go`: `Init`, `IdleCheck`, `OnIdle`, `Tick`, `loseLostQuarry`, `walkPath`, `idle`.
- Modify `internal/behaviour/behaviour.go`: `Wander` limit.
- Create `internal/behaviour/tick_test.go`, `internal/behaviour/rat_test.go`.
- Create `internal/hooks/NewRound_RoomRoundTick.go`, `internal/hooks/NewRound_ActorUpkeep.go`, `internal/hooks/NewRound_ActorTick.go`, `internal/hooks/actortick_test.go`.
- Delete `internal/hooks/NewRound_UserRoundTick.go`, `NewRound_MobRoundTick.go`, `NewRound_IdleMobs.go`, `MobIdle_HandleIdleMobs.go`.
- Modify `internal/hooks/hooks.go`, `main.go`, `modules/follow/follow.go`, `internal/events/eventtypes.go`, `_datafiles/html/admin/docs-modules.html`.
- Modify (boredom and dead fields) `internal/mobs/mobs.go`, `internal/mobs/saved.go`, `internal/mobs/saved_test.go`, `internal/rooms/rooms.go`, `internal/configs/config.memory.go`, `_datafiles/config.yaml`, `_datafiles/html/admin/config.html`, `_datafiles/html/admin/config-wizard.html`, `internal/accounts/userrecord.go`, `internal/usercommands/admin.mob.go`.
- Create `internal/guards/tick_test.go`.
- Docs: `PATCH_NOTES.md`, `status.md`, the spec's "Notes for later slices".

Every subagent: edit files only with the Edit and Write tools (no `sed -i`, no shell redirection into repo files, no Python read-modify-write). `git add` with named paths only; commit messages through a quoted heredoc, ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never push or touch upstream.

---

### Task 1: `Controller.Tick` and `OnBehaviourTick`

**Files:** Modify `internal/actors/controller.go`, `internal/actors/controller_test.go`, `internal/messaging/messaging_test.go:25-26`.

- [ ] **Step 1: Write the failing test** (append to `internal/actors/controller_test.go`)

```go
// bodyMob is the helper in registry_test.go; add the mobs import to this file.
func TestBehaviourControllerTickFiresTheBehaviourHook(t *testing.T) {
	m := bodyMob()
	var got []*mobs.Mob
	OnBehaviourTick.Register(func(x *mobs.Mob) *mobs.Mob {
		if x == m {
			got = append(got, x)
		}
		return x
	})

	NewBehaviourController(m).Tick()

	assert.Equal(t, []*mobs.Mob{m}, got)
}

func TestConnectionControllerTickDoesNothing(t *testing.T) {
	assert.NotPanics(t, func() { NewConnectionController(nil).Tick() })
}
```

- [ ] **Step 2: Run it and see it fail**

Run: `go test ./internal/actors/ -run 'Tick' -count=1`
Expected: build failure, `OnBehaviourTick` undefined and `Tick` undefined.

- [ ] **Step 3: Implement** in `internal/actors/controller.go`

Replace the interface and its doc comment with:

```go
// Controller is what drives an actor: a connection with its account, or an
// NPC behaviour. Code outside this package asks the controller questions and
// never asserts its concrete type (internal/guards enforces this).
type Controller interface {
	// Account is the account driving the actor, or nil when none does.
	Account() *accounts.Account
	// Perceive is called by internal/messaging for each message that
	// reaches the actor.
	Perceive(p perception.Perception)
	// Tick is called once a round for every live actor, after combat and
	// healing (internal/hooks ActorTick): what the actor does by itself.
	Tick()
}
```

After `ConnectionController.Perceive` add:

```go
// Tick does nothing: a player acts by typing.
func (c *ConnectionController) Tick() {}
```

Replace the `BehaviourController` doc comment with:

```go
// BehaviourController drives an NPC. Its round of behaviour runs from Tick
// through OnBehaviourTick; the behaviour state stays on the Mob (slice 6).
```

After `BehaviourController.Perceive` add:

```go
// OnBehaviourTick fires once a round for each NPC, from BehaviourController.Tick.
// internal/behaviour attaches idle, wander and path walking to it
// (behaviour.Init): this package cannot import behaviour, which imports rooms,
// which imports this package.
var OnBehaviourTick util.Hook[*mobs.Mob]

// Tick runs the NPC's behaviour for the round.
func (b *BehaviourController) Tick() { OnBehaviourTick.Fire(b.mob) }
```

In `internal/messaging/messaging_test.go`, after the `recorder.Perceive` line add:

```go
func (r *recorder) Tick()                              {}
```

- [ ] **Step 4: Run the tests**

Run: `go build ./... && go test ./internal/actors/ ./internal/messaging/ -count=1`
Expected: ok for both.

- [ ] **Step 5: Commit**

```bash
git add internal/actors/controller.go internal/actors/controller_test.go internal/messaging/messaging_test.go
git commit -F - <<'EOF'
actors: Controller.Tick; BehaviourController.Tick fires OnBehaviourTick

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 2: `behaviour.Tick` (idle, wander, path walking) and the wander limit

**Files:** Create `internal/behaviour/tick.go`, `internal/behaviour/tick_test.go`; modify `internal/behaviour/behaviour.go:31-37` and its test file. The hooks that call the old code stay until Task 4, so the build stays green.

- [ ] **Step 1: Write the failing tests** (`internal/behaviour/tick_test.go`)

```go
package behaviour

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// busyMob is an NPC that idles every round it may: activity 100, one emote.
func busyMob(t *testing.T, roomId, homeId, maxWander int) *mobs.Mob {
	t.Helper()
	m := testMob(t, roomId, homeId, maxWander)
	m.ActivityLevel = 100
	m.IdleCommands = []string{"emote hums."}
	return m
}

// tickRounds runs n rounds of Tick and returns the input queued.
func tickRounds(t *testing.T, mob *mobs.Mob, n int) []string {
	t.Helper()
	return queuedInput(t, mob, func() {
		for i := 0; i < n; i++ {
			Tick(mob)
		}
	})
}

func TestTickIdlesAtItsActivityLevel(t *testing.T) {
	testRoom(t, 9201, testZone, nil)
	mob := busyMob(t, 9201, 9201, 3)
	util.SeedRandForTest(1)

	assert.Contains(t, tickRounds(t, mob, 3), "emote hums.")
}

func TestTickWithActivityZeroNeverIdles(t *testing.T) {
	testRoom(t, 9202, testZone, nil)
	mob := busyMob(t, 9202, 9202, 3)
	mob.ActivityLevel = 0

	assert.Empty(t, tickRounds(t, mob, 50))
}

func TestTickPreventIdleSkipsOneRound(t *testing.T) {
	testRoom(t, 9203, testZone, nil)
	mob := busyMob(t, 9203, 9203, 3)
	mob.PreventIdle = true

	assert.Empty(t, tickRounds(t, mob, 1))
	assert.False(t, mob.PreventIdle, "the hold lasts one round")
}

func TestTickWhileFightingDoesNotIdle(t *testing.T) {
	testRoom(t, 9204, testZone, nil)
	mob := busyMob(t, 9204, 9204, 3)
	mob.Character.Aggro = &characters.Aggro{}

	assert.Empty(t, tickRounds(t, mob, 5))
}

func TestTickOnIdleSkipStopsTheIdle(t *testing.T) {
	testRoom(t, 9205, testZone, nil)
	mob := busyMob(t, 9205, 9205, 3)
	OnIdle.Register(func(c IdleCheck) IdleCheck {
		if c.Mob == mob {
			c.Skip = true
		}
		return c
	})

	assert.Empty(t, tickRounds(t, mob, 5))
}

func TestTickAtHomeOverTheWanderLimitStartsAgain(t *testing.T) {
	testRoom(t, 9206, testZone, nil)
	mob := busyMob(t, 9206, 9206, 2)
	mob.ActivityLevel = 0
	mob.WanderCount = 5

	tickRounds(t, mob, 1)

	assert.Equal(t, 0, mob.WanderCount, "an NPC home again may wander again")
}

func TestTickAwayAtTheWanderLimitPlansTheWayHome(t *testing.T) {
	testRoom(t, 9207, testZone, map[string]int{"north": 9208})
	testRoom(t, 9208, testZone, map[string]int{"south": 9207})
	mob := busyMob(t, 9208, 9207, 2)
	mob.WanderCount = 2

	got := tickRounds(t, mob, 1)

	assert.Empty(t, got, "it plans; it does not idle")
	assert.Equal(t, 1, mob.Path.Len())
}

func TestTickWalksItsPath(t *testing.T) {
	testRoom(t, 9209, testZone, map[string]int{"east": 9210})
	testRoom(t, 9210, testZone, nil)
	mob := busyMob(t, 9209, 9209, 3)
	PathTo(mob, "9210")

	got := tickRounds(t, mob, 1)

	require.NotEmpty(t, got)
	assert.Equal(t, "east", got[0])
}
```

Add the `mobs` import (`github.com/GoMudEngine/GoMud/internal/mobs`). In `internal/behaviour/behaviour_test.go`, after `TestWanderPastMaxWanderPlansTheWayHome`, add:

```go
func TestWanderAtMaxWanderPlansTheWayHome(t *testing.T) {
	testRoom(t, 9143, testZone, map[string]int{"north": 9144})
	testRoom(t, 9144, testZone, map[string]int{"south": 9143})
	mob := testMob(t, 9144, 9143, 2)
	mob.WanderCount = 2

	got := queuedInput(t, mob, func() { Wander(mob) })

	assert.Empty(t, got, "maxwander 2 allows two moves, not three")
	assert.Equal(t, 1, mob.Path.Len())
}
```

- [ ] **Step 2: Run them and see them fail**

Run: `go test ./internal/behaviour/ -count=1`
Expected: build failure (`Tick`, `OnIdle`, `IdleCheck` undefined).

- [ ] **Step 3: Implement** `internal/behaviour/tick.go`

```go
package behaviour

import (
	"strconv"
	"strings"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var initOnce sync.Once

// Init attaches Tick to every NPC's controller (actors.OnBehaviourTick).
// main calls it once at startup, after actors.Init.
func Init() {
	initOnce.Do(func() {
		actors.OnBehaviourTick.Register(func(m *mobs.Mob) *mobs.Mob { Tick(m); return m })
	})
}

// IdleCheck asks whether an NPC may idle this round. A handler sets Skip to
// keep it from idling.
type IdleCheck struct {
	Mob  *mobs.Mob
	Skip bool
}

// OnIdle fires each round for an NPC that is about to idle: it is not held
// back, not fighting and not walking a path. A handler that sets Skip stops
// this round's idle, including the way home past its wander limit; the follow
// module uses it so a following NPC stays with its leader. It replaces the
// MobIdle event (slice 6).
var OnIdle util.Hook[IdleCheck]

// Tick is one NPC's round of behaviour, run after combat and healing through
// BehaviourController.Tick.
func Tick(mob *mobs.Mob) {
	if mob == nil {
		return
	}

	// If idle prevented, it's a one round interrupt (until another comes along)
	if mob.PreventIdle {
		mob.PreventIdle = false
		return
	}

	// If they are doing some sort of combat thing,
	// Don't do idle actions
	if mob.Character.Aggro != nil {
		loseLostQuarry(mob)
		return
	}

	if walkPath(mob) {
		return
	}

	if OnIdle.Fire(IdleCheck{Mob: mob}).Skip {
		return
	}

	idle(mob)
}

// loseLostQuarry ends the fight of an NPC whose target is gone: a player
// target who is offline or in another room is lost. A vanished NPC target is
// treated the same (accepted trade-off).
func loseLostQuarry(mob *mobs.Mob) {
	t := mob.Character.Aggro.Target
	if t.IsZero() || mobs.GetInstance(t) != nil {
		return
	}
	if user := accounts.GetByBodyId(t); user == nil || user.Character.RoomId != mob.Character.RoomId {
		mob.Command(`emote mumbles about losing their quarry.`)
		mob.Character.Aggro = nil
		events.AddToQueue(events.AggroChanged{ActorId: mob.Character.Id, RoomId: mob.Character.RoomId})
	}
}

// walkPath takes the NPC's next step on its path, and reports whether the
// round went on the path. A path that ends or breaks is cleared, and the
// round goes on to idle.
func walkPath(mob *mobs.Mob) bool {
	currentStep := mob.Path.Current()
	if currentStep == nil && mob.Path.Len() == 0 {
		return false
	}

	if currentStep != nil {
		// If their currentStep isn't actually the room they are in
		// They've somehow been moved. Reclaculate a new path.
		if currentStep.RoomId() != mob.Character.RoomId {
			reDoWaypoints := mob.Path.Waypoints()
			if len(reDoWaypoints) > 0 {
				newArgs := make([]string, 0, len(reDoWaypoints))
				for _, wpInt := range reDoWaypoints {
					newArgs = append(newArgs, strconv.Itoa(wpInt))
				}
				PathTo(mob, strings.Join(newArgs, ` `))
				return true
			}
			// if we were unable to come up with a new path, send them home.
			PathTo(mob, "home")
			return true
		}
	}

	if nextStep := mob.Path.Next(); nextStep != nil {
		if room := rooms.LoadRoom(mob.Character.RoomId); room != nil {
			if exitInfo, ok := room.Exits[nextStep.ExitName()]; ok {
				if exitInfo.RoomId == nextStep.RoomId() {
					mob.Command(nextStep.ExitName())
					if nextStep.Waypoint() {
						// Anytime a mob reaches a waypoint, introduce a 1 second delay before they can perform any additional commands.
						// This gives a more natural feel to mob behavior, and gives those following a moment to catch up before the mob does something.
						mob.Command("noop", 1)
					}
					return true
				}
			}
		}
	}

	mob.Path.Clear()
	if mob.HomeRoomId == mob.Character.RoomId {
		mob.WanderCount = 0
	}
	return false
}

// idle is the NPC's free round: home if it has strayed, else maybe one of its
// idle commands at its ActivityLevel odds.
func idle(mob *mobs.Mob) {
	isCharmed := mob.Character.IsCharmed()

	// if a mob shouldn't be allowed to leave their area (via wandering)
	// but has somehow been displaced, such as pulling through combat, spells, or otherwise
	// tell them to path back home
	if mob.MaxWander == 0 && mob.Character.RoomId != mob.HomeRoomId && !isCharmed {
		PathTo(mob, "home")
	}

	if isCharmed {
		return
	}

	// MaxWander moves used up: away from home it heads home; at home it may
	// wander again (it used to stay over the limit and never idle again).
	if mob.MaxWander > 0 && mob.WanderCount >= mob.MaxWander {
		if mob.Character.RoomId != mob.HomeRoomId {
			PathTo(mob, "home")
			return
		}
		mob.WanderCount = 0
	}

	// If they have idle commands, maybe do one of them?
	if util.Rand(100) < mob.ActivityLevel {
		if idleCmd := mob.GetIdleCommand(); idleCmd != `` {
			RunIdle(mob, idleCmd)
		}
	}
}
```

In `internal/behaviour/behaviour.go`, replace lines 30-37:

```go
	// If they aren't home and need to go home, do it.
	if mob.Character.RoomId != mob.HomeRoomId {
		if mob.MaxWander > -1 { // -1 means they can wander forever and never go home. 0 means they never wander.
			if mob.WanderCount > mob.MaxWander {
				PathTo(mob, `home`)
				return
			}
		}
	}
```

with:

```go
	// If they aren't home and have used up their MaxWander moves, go home.
	// -1 means they can wander forever and never go home. 0 means they never wander.
	if mob.Character.RoomId != mob.HomeRoomId {
		if mob.MaxWander > 0 && mob.WanderCount >= mob.MaxWander {
			PathTo(mob, `home`)
			return
		}
	}
```

Also update the `PathTo` doc comment "the idle round tick walks the path" to "Tick walks the path".

- [ ] **Step 4: Run the tests**

Run: `go build ./... && go test ./internal/behaviour/ -count=1`
Expected: ok. If `TestTickIdlesAtItsActivityLevel` fails on the 1-in-100 empty pick for seed 1, raise the round count to 5; do not change the code.

- [ ] **Step 5: Commit**

```bash
git add internal/behaviour/tick.go internal/behaviour/tick_test.go internal/behaviour/behaviour.go internal/behaviour/behaviour_test.go
git commit -F - <<'EOF'
behaviour: Tick runs an NPC's idle, wander and path; OnIdle veto; maxwander allows exactly that many moves; an NPC home over its limit idles again

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 3: The seeded rat test (slice 6's done test)

**Files:** Create `internal/behaviour/rat_test.go`.

- [ ] **Step 1: Write the test**

```go
package behaviour

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestTheRatIdlesAndWandersWithinMaxWander is slice 6's done test: the rat,
// with its own data, idles at its activitylevel odds and never makes more
// than maxwander wander moves before it walks home.
//
// The rooms are a ring with one exit each, so the only random choices are the
// rat's own (a room with two or more exits picks in map order, which a seed
// does not fix). The ring is longer than maxwander, so the rat never wanders
// home by chance: every trip uses up its moves and paths home.
func TestTheRatIdlesAndWandersWithinMaxWander(t *testing.T) {
	b, err := os.ReadFile("../../_datafiles/world/empty/mobs/startland/1-rat.yaml")
	require.NoError(t, err)
	var rat mobs.Mob
	require.NoError(t, yaml.Unmarshal(b, &rat))
	require.Equal(t, 10, rat.ActivityLevel)
	require.Equal(t, 8, rat.MaxWander)
	require.Equal(t, []string{"emote wiggles its nose", "wander", ""}, rat.IdleCommands)

	const home, ringSize = 9301, 12
	next := func(id int) int { return home + (id-home+1)%ringSize }
	for id := home; id < home+ringSize; id++ {
		testRoom(t, id, testZone, map[string]int{"east": next(id)})
	}
	mob := testMob(t, home, home, rat.MaxWander)
	mob.ActivityLevel = rat.ActivityLevel
	mob.IdleCommands = rat.IdleCommands

	// Rounds in which the rat rolls its activity odds.
	rolls := 0
	OnIdle.Register(func(c IdleCheck) IdleCheck {
		m := c.Mob
		if m == mob && !(m.MaxWander > 0 && m.WanderCount >= m.MaxWander && m.Character.RoomId != m.HomeRoomId) {
			rolls++
		}
		return c
	})

	var inputs []string
	lid := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		if in := e.(events.Input); in.ActorId == mob.Character.Id {
			inputs = append(inputs, in.InputText)
		}
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.Input{}, lid) })
	events.ProcessEvents()
	inputs = nil

	util.SeedRandForTest(6)
	emotes, wanders, trips, wandersSinceHome, mostWanders := 0, 0, 0, 0, 0
	for round := 0; round < 20000; round++ {
		Tick(mob)
		events.ProcessEvents()
		for _, in := range inputs {
			switch in {
			case "go east": // a wander move
				wanders++
				wandersSinceHome++
				mob.Character.RoomId = next(mob.Character.RoomId)
			case "east": // a step on the path home
				mob.Character.RoomId = next(mob.Character.RoomId)
				if mob.Character.RoomId == home {
					trips++
				}
			case "emote wiggles its nose":
				emotes++
			case "noop":
			default:
				t.Fatalf("round %d: unexpected input %q", round, in)
			}
		}
		inputs = inputs[:0]
		if mob.Character.RoomId == home {
			wandersSinceHome = 0
		}
		if wandersSinceHome > mostWanders {
			mostWanders = wandersSinceHome
		}
		require.LessOrEqual(t, wandersSinceHome, rat.MaxWander, "round %d: wandered past maxwander", round)
	}

	assert.Equal(t, rat.MaxWander, mostWanders, "the rat uses all of its maxwander moves")
	assert.Greater(t, trips, 10, "the rat walks home after its moves")

	// activitylevel 10, then 99 in 100 an idle command, one in three of each:
	// 3.3% of rolls emote and 3.3% wander.
	require.Greater(t, rolls, 15000)
	emoteRate := float64(emotes) / float64(rolls)
	wanderRate := float64(wanders) / float64(rolls)
	assert.InDelta(t, 0.033, emoteRate, 0.008, "emotes %d of %d rolls", emotes, rolls)
	assert.InDelta(t, 0.033, wanderRate, 0.008, "wanders %d of %d rolls", wanders, rolls)
}
```

- [ ] **Step 2: Run it**

Run: `go test ./internal/behaviour/ -run TestTheRat -count=1 -v`
Expected: PASS. Report the emote and wander counts, `rolls`, `trips` (add a `t.Logf` while checking, remove it before commit).

- [ ] **Step 3: Show it catches the old limit**

Temporarily change `>=` to `>` in `idle` (`tick.go`) and `Wander` (`behaviour.go`), run again, expect FAIL with "wandered past maxwander", then undo exactly those two characters. Report both runs. Do not commit the temporary change.

- [ ] **Step 4: Commit**

```bash
git add internal/behaviour/rat_test.go
git commit -F - <<'EOF'
behaviour: seeded test, the rat idles at its activitylevel odds and wanders within maxwander

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

### Task 4: One actor upkeep and one actor tick in the round

**Files:** Create `internal/hooks/NewRound_RoomRoundTick.go`, `NewRound_ActorUpkeep.go`, `NewRound_ActorTick.go`, `actortick_test.go`. Delete `NewRound_UserRoundTick.go`, `NewRound_MobRoundTick.go`, `NewRound_IdleMobs.go`, `MobIdle_HandleIdleMobs.go`. Modify `internal/hooks/hooks.go:20-32`, `main.go:106`, `modules/follow/follow.go:68, 246-254`.

- [ ] **Step 1: Write the failing tests** (`internal/hooks/actortick_test.go`)

```go
package hooks

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActorUpkeepTicksEveryActorsCooldowns(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Cooldowns = characters.Cooldowns{"kick": 3}
	actors.Spawn(c, actors.NewConnectionController(&accounts.Account{UserId: 61, Character: c}))
	m := mobs.NewTestInstance(mobs.Mob{})
	t.Cleanup(func() { mobs.Rest(m.Character.Id) })
	m.Character.Cooldowns = characters.Cooldowns{"bite": 3}

	ActorUpkeep(events.NewRound{RoundNumber: 1})

	assert.Equal(t, 2, c.Cooldowns["kick"], "a player's cooldowns tick")
	assert.Equal(t, 2, m.Character.Cooldowns["bite"], "an NPC's cooldowns tick")
}

func TestActorTickRunsEachNPCsBehaviourOnce(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	actors.Spawn(c, actors.NewConnectionController(&accounts.Account{UserId: 62, Character: c}))
	m := mobs.NewTestInstance(mobs.Mob{})
	t.Cleanup(func() { mobs.Rest(m.Character.Id) })
	require.NotNil(t, actors.Get(m.Character.Id))

	ticks := 0
	actors.OnBehaviourTick.Register(func(x *mobs.Mob) *mobs.Mob {
		if x == m {
			ticks++
		}
		return x
	})

	ActorTick(events.NewRound{RoundNumber: 1})

	assert.Equal(t, 1, ticks)
}
```

If `characters.New()` leaves `Cooldowns` nil, the literal assignment above covers it. If `Validate` on a bare `characters.New()` panics for missing race data, set `c.RaceId` to the rat's race (10) as `charmexpiry_test.go` would, and say so in the report.

- [ ] **Step 2: Run them and see them fail**

Run: `go test ./internal/hooks/ -run 'ActorUpkeep|ActorTick' -count=1`
Expected: build failure, `ActorUpkeep` and `ActorTick` undefined.

- [ ] **Step 3: Create `internal/hooks/NewRound_RoomRoundTick.go`**

The room half of `UserRoundTick`, unchanged in behaviour:

```go
package hooks

import (
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

//
// Room Round Tick: rooms with players in them (slice 6 keeps them only).
//

func RoomRoundTick(e events.Event) events.ListenerReturn {

	for _, roomId := range rooms.GetRoomsWithPlayers() {
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}

		room.RoundTick()

		chanceIn100 := 5
		if room.RoomId == -1 {
			chanceIn100 = 20
		}

		var idleMsgs []string
		if len(room.IdleMessages) > 0 {
			idleMsgs = room.IdleMessages
		} else {
			if zCfg := rooms.GetZoneConfig(room.Zone); zCfg != nil {
				if len(zCfg.IdleMessages) > 0 {
					idleMsgs = zCfg.IdleMessages
				}
			}
		}

		idleMsgCt := len(idleMsgs)
		if idleMsgCt > 0 && util.Rand(100) < chanceIn100 {

			if targetRoomId, err := strconv.Atoi(idleMsgs[0]); err == nil {
				idleMsgCt = 0
				if tgtRoom := rooms.LoadRoom(targetRoomId); tgtRoom != nil {
					idleMsgs = tgtRoom.IdleMessages
					idleMsgCt = len(idleMsgs)
				}
			}

			if idleMsgCt > 0 {
				// pick a random message
				idleMsgIndex := uint8(util.Rand(idleMsgCt))

				// If it's a repeating message, treat it as a non-message
				// (Unless it's the only one)
				if idleMsgIndex != room.LastIdleMessage || idleMsgCt == 1 {

					room.LastIdleMessage = idleMsgIndex

					msg := idleMsgs[idleMsgIndex]
					if msg != `` {
						room.SendText(msg)
					}
				}
			}
		}
	}

	return events.Continue
}
```

Before writing it, re-read `NewRound_UserRoundTick.go:21-77` and confirm the copy matches it line for line apart from the loop shape (`continue` on a nil room instead of the `if` block).

- [ ] **Step 4: Create `internal/hooks/NewRound_ActorUpkeep.go`**

Move `runCharmExpiredCommand` (with its doc comment) here unchanged from `NewRound_MobRoundTick.go:13-51`, then:

```go
// ActorUpkeep is every actor's round upkeep, before combat (slice 6): the
// same cooldowns, charm countdown, buff triggers and stat refresh for players
// and NPCs. An account also decays its alignment and reports changed vitals;
// an NPC's charm can expire, and an NPC at 0 health dies. Actors go in ID
// order.
func ActorUpkeep(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	//
	// Reduce existing hostility (if any)
	//
	mobs.ReduceHostility()

	for _, a := range actors.All() {
		if actors.Get(a.Id) == nil {
			continue // ended earlier in this pass
		}

		char := a.Character
		user := a.Controller().Account()
		mob := mobs.GetInstance(a.Id) // nil for a player

		// Roundtick any cooldowns
		char.Cooldowns.RoundTick()

		if user != nil {
			decayAlignment(user, evt.RoundNumber)
		}

		if char.Charmed != nil && char.Charmed.RoundsRemaining > 0 {
			char.Charmed.RoundsRemaining--
		}

		if triggeredBuffs := char.Buffs.Trigger(); len(triggeredBuffs) > 0 {
			triggeredBuffIds := []int{}
			for _, buff := range triggeredBuffs {
				triggeredBuffIds = append(triggeredBuffIds, buff.BuffId)
			}
			events.AddToQueue(events.BuffsTriggered{ActorId: a.Id, BuffIds: triggeredBuffIds})
		}

		// Do charm cleanup: an NPC's charm expires; a player's only counts down.
		if user == nil && mob != nil && char.IsCharmed() && char.Charmed.RoundsRemaining == 0 {
			cmd := char.Charmed.ExpiredCommand
			if charmedBy := char.RemoveCharm(); !charmedBy.IsZero() {
				if charmedUser := accounts.GetByBodyId(charmedBy); charmedUser != nil {
					charmedUser.Character.TrackCharmed(a.Id, false)
				}
			}
			runCharmExpiredCommand(mob, cmd)
			if mobs.GetInstance(a.Id) == nil {
				continue // rested or ended: nothing left to refresh
			}
		}

		// Recalculate all stats at the end of the round tick
		hBefore, mBefore := char.Health, char.Mana
		char.Validate()

		if user != nil {
			if char.Health != hBefore || char.Mana != mBefore {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
			}
			continue
		}

		if mob != nil && char.Health <= 0 {
			// Mob died
			mob.Command(`suicide`)
		}
	}

	return events.Continue
}

// decayAlignment moves an account's alignment toward neutral every
// AlignmentDecayRounds rounds and tells the player when its name changes.
// Accounts only: NPC alignment is a Phase 5 rule.
func decayAlignment(user *accounts.Account, round uint64) {
	decayRounds := int(configs.GetGamePlayConfig().AlignmentDecayRounds)
	if decayRounds <= 0 || round%uint64(decayRounds) != 0 {
		return
	}
	alignmentBefore := user.Character.AlignmentName()
	user.Character.DecayAlignment()
	if alignmentAfter := user.Character.AlignmentName(); alignmentAfter != alignmentBefore {
		before := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
		after := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
		user.SendText(fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, before, after))
		events.AddToQueue(events.AlignmentChanged{
			UserId:       user.UserId,
			AlignmentOld: alignmentBefore,
			AlignmentNew: alignmentAfter,
		})
	}
}
```

Imports: `fmt`, `strings`, `accounts`, `actors`, `behaviour`, `configs`, `events`, `mobs`, `mudlog`, `usercommands`. Re-read `NewRound_UserRoundTick.go:83-128` and `NewRound_MobRoundTick.go:54-111` before writing and confirm every step is carried over; report any you could not place.

- [ ] **Step 5: Create `internal/hooks/NewRound_ActorTick.go`**

```go
package hooks

import (
	"time"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// ActorTick runs every actor's controller for the round, after combat and
// healing (slice 6): an NPC's behaviour (idle, wander, its path) runs from
// BehaviourController.Tick; a player's controller does nothing. Actors go in
// ID order. actors.All is a copy, so a tick may rest or end an NPC.
func ActorTick(e events.Event) events.ListenerReturn {
	tStart := time.Now()
	for _, a := range actors.All() {
		if actors.Get(a.Id) == nil {
			continue // ended earlier in this pass
		}
		a.Controller().Tick()
	}
	util.TrackTime(`ActorTick()`, time.Since(tStart).Seconds())
	return events.Continue
}
```

- [ ] **Step 6: Wire it and delete the old hooks**

In `internal/hooks/hooks.go` replace lines 20-21 (`UserRoundTick`, `MobRoundTick`) with:

```go
	events.RegisterListener(events.NewRound{}, RoomRoundTick)
	events.RegisterListener(events.NewRound{}, ActorUpkeep)
```

and lines 31-32 (`IdleMobs`, the `MobIdle` listener) with:

```go
	events.RegisterListener(events.NewRound{}, ActorTick)
```

Delete with git: `git rm internal/hooks/NewRound_UserRoundTick.go internal/hooks/NewRound_MobRoundTick.go internal/hooks/NewRound_IdleMobs.go internal/hooks/MobIdle_HandleIdleMobs.go`.

In `main.go`, after `actors.Init()` (line 106) add `behaviour.Init()` and import `github.com/GoMudEngine/GoMud/internal/behaviour`.

In `modules/follow/follow.go`: replace line 68 with

```go
	behaviour.OnIdle.Register(f.idleMobHandler)
```

and the handler (lines 246-254) with

```go
// idleMobHandler keeps a following NPC from idling, so it stays with its leader.
func (f *FollowModule) idleMobHandler(c behaviour.IdleCheck) behaviour.IdleCheck {
	if f.isFollowing(followId{mobId: c.Mob.Character.Id}) {
		c.Skip = true
	}
	return c
}
```

Read lines 240-256 first and keep the existing doc comment if there is one. Add the `behaviour` import.

- [ ] **Step 7: Run the tests**

Run: `go build ./... && go vet ./internal/hooks/ ./modules/... && go test ./internal/hooks/ ./internal/behaviour/ ./internal/actors/ ./modules/... -count=1`
Expected: ok. `TestCharmExpired*` still pass (they call `runCharmExpiredCommand`, which moved).

- [ ] **Step 8: Commit**

```bash
git add internal/hooks/NewRound_RoomRoundTick.go internal/hooks/NewRound_ActorUpkeep.go internal/hooks/NewRound_ActorTick.go internal/hooks/actortick_test.go internal/hooks/hooks.go main.go modules/follow/follow.go
git add internal/hooks/NewRound_UserRoundTick.go internal/hooks/NewRound_MobRoundTick.go internal/hooks/NewRound_IdleMobs.go internal/hooks/MobIdle_HandleIdleMobs.go
git commit -F - <<'EOF'
hooks: one actor upkeep before combat and one actor tick after it; NPC behaviour runs from Controller.Tick; follow vetoes idle through behaviour.OnIdle

Deletes UserRoundTick, MobRoundTick, IdleMobs and HandleIdleMobs.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

Run `git status` and `git diff --cached --stat` before committing; the four deletions must be in the commit.

### Task 5: Delete the `MobIdle` event and update the module docs

**Files:** `internal/events/eventtypes.go:135-140`, `_datafiles/html/admin/docs-modules.html:1123, 1557-1563, 1583`.

- [ ] **Step 1:** Grep `MobIdle` across `*.go`, `*.html`, `*.md`, `*.js`, `*.yaml` (exclude `docs/completed/` and the inventory). Expected hits: `eventtypes.go`, `docs-modules.html`, the spec. Any other hit: stop and report.
- [ ] **Step 2:** Delete the `MobIdle` type, its comment and its `Type` method from `eventtypes.go`.
- [ ] **Step 3:** In `docs-modules.html`: delete the `MobIdle` table row (line 1123); in the follow example change the comment and listener lines (1557-1558, 1563) to

```go
    // behaviour.OnIdle lets a module keep an NPC from idling this round.
    behaviour.OnIdle.Register(f.idleMobHandler)
```

(keep the other `RegisterListener` lines, and keep the example's order); replace the callout text at 1583 with: `<p>Register on <code>behaviour.OnIdle</code> and set <code>Skip</code> to keep an NPC from idling (emotes, wandering, the way home) this round. It fires once a round for each NPC that is not held back, fighting or walking a path.</p>`
- [ ] **Step 4:** `go build ./... && go test ./internal/events/ -count=1` → ok.
- [ ] **Step 5: Commit** (`git add internal/events/eventtypes.go _datafiles/html/admin/docs-modules.html`, message `events: drop MobIdle; module docs point at behaviour.OnIdle`).

### Task 6: Delete boredom and the dead idle fields; fix the admin wander prompt

**Files:** `internal/mobs/mobs.go:62-63`, `internal/mobs/saved.go:58`, `internal/mobs/saved_test.go:33,70-71`, `internal/rooms/rooms.go:2066-2073`, `internal/configs/config.memory.go`, `_datafiles/config.yaml:94-102`, `_datafiles/html/admin/config.html:211-212`, `_datafiles/html/admin/config-wizard.html:1344-1357,1652`, `internal/accounts/userrecord.go:462-464`, `internal/usercommands/admin.mob.go:364`.

- [ ] **Step 1:** Delete the `LastIdleCommand` and `BoredomCounter` fields from `mobs.Mob`.
- [ ] **Step 2:** `saved.go:58` becomes `m.WanderCount = 0`. In `saved_test.go` drop `BoredomCounter: 9, LastIdleCommand: 2` from the literal and the two `assert.Zero` lines for them.
- [ ] **Step 3:** In `rooms.go` delete the block "If any players are in the room / Update all mobs in the room that they've seen a player" (`:2066-2073`). If `mobs` becomes an unused import, the build says so; remove it then.
- [ ] **Step 4:** In `config.memory.go` delete the `MaxMobBoredom` and `MobUnloadThreshold` fields and their two `Validate` blocks; the struct comment becomes `// Room memory unload thresholds`. In `_datafiles/config.yaml` delete the two keys and their comment lines (94-102, read first). In `config.html` delete lines 211-212. In `config-wizard.html` delete the two entries (1344-1357) and remove the two names from the `fields` list at 1652.
- [ ] **Step 5:** Delete `Account.RoundTick` (`userrecord.go:462-464`).
- [ ] **Step 5b:** In `internal/behaviour/behaviour.go`, the `Rest` doc comment's "Boredom and charm expiry use it." becomes "Charm expiry uses it."
- [ ] **Step 6:** In `admin.mob.go:364` change the prompt to `How far can this mob wander (-1 = unlimited. 0 = never)?`.
- [ ] **Step 7:** Grep `BoredomCounter|LastIdleCommand|MaxMobBoredom|MobUnloadThreshold|RoundTick\(\)` across the repo (excluding `docs/completed/`, the inventory and the spec). Expected: only `Cooldowns.RoundTick` and `Room.RoundTick` and their callers. Report the grep.
- [ ] **Step 8:** `go build ./... && go test ./internal/mobs/ ./internal/rooms/ ./internal/configs/ ./internal/accounts/ ./internal/usercommands/ -count=1` → ok.
- [ ] **Step 9: Commit** with the named paths above; message `Delete dead NPC boredom unload (Calabe 2026-10-28) and its config keys; drop LastIdleCommand and Account.RoundTick; admin wander prompt reads -1 unlimited, 0 never`.

### Task 7: Guard

**Files:** Create `internal/guards/tick_test.go`.

- [ ] **Step 1: Write the guard**

```go
package guards

import (
	"go/ast"
	"testing"
)

// TestOneActorTick enforces the phase 4b slice 6 ruling: every actor gets the
// same round upkeep (ActorUpkeep) and one tick through Controller.Tick
// (ActorTick). The per-kind round ticks, the idle hooks, the MobIdle event and
// the dead boredom unload are gone and stay gone.
func TestOneActorTick(t *testing.T) {
	banned := map[string]bool{
		"UserRoundTick":      true,
		"MobRoundTick":       true,
		"IdleMobs":           true,
		"HandleIdleMobs":     true,
		"MobIdle":            true,
		"BoredomCounter":     true,
		"MaxMobBoredom":      true,
		"MobUnloadThreshold": true,
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
				t.Errorf("%s: uses %s; one actor upkeep and one actor tick (slice 6)", path, id.Name)
			}
			return true
		})
	}

	// Positive control: the two passes and the controller tick exist, so the
	// walk reached the files the ban is about.
	want := map[string]bool{
		"internal/hooks/NewRound_ActorUpkeep.go ActorUpkeep": false,
		"internal/hooks/NewRound_ActorTick.go ActorTick":     false,
		"internal/actors/controller.go Tick":                 false,
	}
	for path, f := range goFiles(t) {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				key := path + " " + fd.Name.Name
				if _, ok := want[key]; ok {
					want[key] = true
				}
			}
		}
	}
	for key, found := range want {
		if !found {
			t.Errorf("control: %s not found; the guard may be scanning the wrong place", key)
		}
	}
}
```

- [ ] **Step 2:** Run `go test ./internal/guards/ -count=1` → ok.
- [ ] **Step 3: Prove the guard bites:** add `var MobIdle int` to `internal/hooks/hooks.go`, run the guard, expect FAIL naming `internal/hooks/hooks.go: uses MobIdle`, then remove exactly that line. Report both runs.
- [ ] **Step 4: Commit** (`git add internal/guards/tick_test.go`, message `guards: one actor tick; per-kind ticks, idle hooks, MobIdle and boredom stay gone`).

### Task 8: Verify, smoke, docs

- [ ] **Step 1:** Follow the `geomud-verifying` skill: `go build ./...`, `go vet ./...`, `go test ./... -count=1`. Compare with the baseline (39 packages ok, 3 known Windows failures); report the package counts and the names of any failures.
- [ ] **Step 2:** Smoke: `go run ./cmd/smoke`, `go run ./cmd/smoke -bots 2`, `go run ./cmd/smoke -bots 10`. Report pass or fail for each with the last lines of output.
- [ ] **Step 3:** Golden combat test unchanged: `go test ./internal/hooks/ -run Golden -count=1` passes without `-update`.
- [ ] **Step 4: Docs.**
  - `PATCH_NOTES.md`: a slice 6 line at the top of the list, in the style of the slice 5 line: one upkeep and one tick for every actor; NPC behaviour from `Controller.Tick`; `behaviour.OnIdle` replaces `MobIdle` for modules; rulings (boredom deleted, input order kept per kind); fixes (maxwander allows exactly that many moves; an NPC home over its limit idles again; charm-expiry rest no longer followed by a refresh and a `suicide`; admin wander prompt); the seeded rat test; the guard.
  - The spec's "Notes for later slices": add a "Slice 6 (built 2026-10-28) leaves for later" entry: behaviour state stays on the `Mob` (possession moves it); input ordering kept per kind on `Account()` (Calabe); NPC reactions to perceptions are still none, and when added queue through `mob.Command`; `Room.RoundTick` still runs only in rooms with players; wander state resets on copyover and restore; `GetRandomExit` and `findPath` pick in map order, so a seed does not fix wandering where a room has two or more exits; actor order within a round is ID order. In the slice 4 entry remove the "Slice 6: ..." sentence about input ordering and idle calls, and in the slice 3 entry change "so slice 6 queues NPC reactions" to "so NPC reactions, when added, are queued", since slice 6 settles them.
  - `status.md`: "Now" says slice 6 built on the branch, PR open; "Next" item 1 becomes slice 7.
- [ ] **Step 5: Commit** the docs with named paths.

### Task 9: Whole-branch review and PR

- [ ] **Step 1:** Opus whole-branch review of `git diff master...HEAD` against this plan and the spec row: correctness of the tick order, nothing from the two deleted ticks dropped, the follow veto, locks (`actors.All`, `util.Hook` re-entry), tests that would pass on the old code. Fix confirmed findings in new commits.
- [ ] **Step 2:** Leave this plan and the inventory at the repo root (slice 7's first commit moves them to `docs/completed/`, as this slice did for slice 5). Push `strip/phase-4b-6-one-tick` to `origin`, open the PR with `gh pr create --repo pruuk/GeOMUD --base master`, the body ending with the Claude Code line. Merge once CI is green.
