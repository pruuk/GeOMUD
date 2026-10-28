package behaviour

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
)

// logCapture is the in-game log tee: it records every log line's level and text.
type logCapture struct {
	mu    sync.Mutex
	lines []string
}

func (l *logCapture) Println(level string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	parts := []string{level}
	for _, x := range v {
		if s, ok := x.(string); ok {
			parts = append(parts, s)
		}
	}
	l.lines = append(l.lines, strings.Join(parts, " "))
}

// has reports whether a line at the given level (WARN, INFO...) mentions sub.
func (l *logCapture) has(level, sub string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, line := range l.lines {
		if strings.HasPrefix(line, level+" ") && strings.Contains(line, sub) {
			return true
		}
	}
	return false
}

var logs = &logCapture{}

func TestMain(m *testing.M) {
	mudlog.SetupLogger(logs, "", "", false)
	actors.Init()
	os.Exit(m.Run())
}

const testZone = "Behaviour Zone"

// testRoom puts a room in memory for the test and takes it out afterwards.
func testRoom(t *testing.T, id int, zone string, exits map[string]int) *rooms.Room {
	t.Helper()
	r := rooms.NewEmptyRoom()
	r.RoomId = id
	r.Zone = zone
	for name, to := range exits {
		r.Exits[name] = exit.RoomExit{RoomId: to}
	}
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(id) })
	return r
}

func testMob(t *testing.T, roomId, homeId, maxWander int) *mobs.Mob {
	t.Helper()
	actors.Reset()
	t.Cleanup(actors.Reset)
	m := mobs.Mob{MobId: 7, HomeRoomId: homeId, MaxWander: maxWander}
	m.Character = *characters.New()
	m.Character.Name = "rat"
	m.Character.RoomId = roomId
	m.Character.Zone = testZone
	return mobs.NewTestInstance(m)
}

// queuedInput collects the input queued for the NPC while f runs.
func queuedInput(t *testing.T, mob *mobs.Mob, f func()) []string {
	t.Helper()
	var got []string
	id := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		if in := e.(events.Input); in.ActorId == mob.Character.Id {
			got = append(got, in.InputText)
		}
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.Input{}, id) })
	events.ProcessEvents() // drain anything left over
	got = nil
	f()
	events.ProcessEvents()
	return got
}

func TestRunIdleWanderQueuesGoForTheOnlyExit(t *testing.T) {
	testRoom(t, 9101, testZone, map[string]int{"north": 9102})
	testRoom(t, 9102, testZone, nil)
	mob := testMob(t, 9101, 9101, 3)

	got := queuedInput(t, mob, func() { RunIdle(mob, "wander") })

	assert.Equal(t, []string{"go north"}, got)
	assert.Equal(t, 1, mob.WanderCount)
}

func TestWanderStaysInTheMobsZone(t *testing.T) {
	testRoom(t, 9111, testZone, map[string]int{"north": 9112})
	testRoom(t, 9112, "Another Zone", nil)
	mob := testMob(t, 9111, 9111, 3)

	got := queuedInput(t, mob, func() { Wander(mob) })

	assert.Empty(t, got)
	assert.Equal(t, 0, mob.WanderCount)
}

func TestWanderWithMaxWanderZeroQueuesNothing(t *testing.T) {
	testRoom(t, 9121, testZone, map[string]int{"north": 9122})
	testRoom(t, 9122, testZone, nil)
	mob := testMob(t, 9121, 9121, 0)

	got := queuedInput(t, mob, func() { RunIdle(mob, "wander") })

	assert.Empty(t, got)
}

func TestWanderCharmedDoesNothing(t *testing.T) {
	testRoom(t, 9131, testZone, map[string]int{"north": 9132})
	testRoom(t, 9132, testZone, nil)
	mob := testMob(t, 9131, 9131, 3)
	mob.Character.Charm(mob.Character.Id, characters.CharmPermanent, "")

	got := queuedInput(t, mob, func() { Wander(mob) })

	assert.Empty(t, got)
}

func TestWanderPastMaxWanderPlansTheWayHome(t *testing.T) {
	testRoom(t, 9141, testZone, map[string]int{"north": 9142})
	testRoom(t, 9142, testZone, map[string]int{"south": 9141})
	mob := testMob(t, 9142, 9141, 1)
	mob.WanderCount = 2

	got := queuedInput(t, mob, func() { Wander(mob) })

	assert.Empty(t, got, "pathto plans; it does not issue a move")
	assert.Equal(t, 1, mob.Path.Len(), "a one-step path home")
}

func TestWanderAtMaxWanderPlansTheWayHome(t *testing.T) {
	testRoom(t, 9143, testZone, map[string]int{"north": 9144})
	testRoom(t, 9144, testZone, map[string]int{"south": 9143})
	mob := testMob(t, 9144, 9143, 2)
	mob.WanderCount = 2

	got := queuedInput(t, mob, func() { Wander(mob) })

	assert.Empty(t, got, "maxwander 2 allows two moves, not three")
	assert.Equal(t, 1, mob.Path.Len())
}

func TestPathToUnknownRoomLogsAndMarksLost(t *testing.T) {
	testRoom(t, 9151, testZone, nil)
	mob := testMob(t, 9151, 9199, 3) // home is a room that does not exist

	got := queuedInput(t, mob, func() { PathTo(mob, "home") })

	assert.Empty(t, got, "the NPC no longer emotes confusion")
	assert.True(t, logs.has("WARN", "pathto"), "the failure is a warning in the server log")
	assert.True(t, mob.Character.HasAdjective("lost"))
	v, _ := mob.GetTempData("home-impossible").(bool)
	assert.True(t, v)
}

func TestPathToPlansAPathAndDoesNotMove(t *testing.T) {
	testRoom(t, 9161, testZone, map[string]int{"east": 9162})
	testRoom(t, 9162, testZone, nil)
	mob := testMob(t, 9161, 9161, 3)

	got := queuedInput(t, mob, func() { PathTo(mob, "9162") })

	assert.Empty(t, got)
	assert.Equal(t, 1, mob.Path.Len())
}

func TestRunIdleOtherCommandsAreQueuedAsInput(t *testing.T) {
	testRoom(t, 9171, testZone, nil)
	mob := testMob(t, 9171, 9171, 3)

	got := queuedInput(t, mob, func() { RunIdle(mob, "emote hums.") })

	assert.Equal(t, []string{"emote hums."}, got)
}

func TestRunIdlePathToRunsTheAction(t *testing.T) {
	testRoom(t, 9181, testZone, map[string]int{"east": 9182})
	testRoom(t, 9182, testZone, nil)
	mob := testMob(t, 9181, 9181, 3)

	got := queuedInput(t, mob, func() { RunIdle(mob, "pathto 9182") })

	assert.Empty(t, got)
	assert.Equal(t, 1, mob.Path.Len())
}
