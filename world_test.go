package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests drive World.HandleInputEvents on a World built in the test
// (only the two input maps; NewWorld would register a second listener) with
// real actors, a real room and a recording command in the shared registry.

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

var testLogs = &logCapture{}

const worldTestRoomId = 93001

// ran records each run of the zz_world test command as "<actorId>:<rest>".
var ran []string

func TestMain(m *testing.M) {
	mudlog.SetupLogger(testLogs, "", "", false)
	actors.Init()
	dataDir, err := filepath.Abs("_datafiles/world/empty")
	if err != nil {
		panic(err)
	}
	if err := configs.AddOverlayOverrides(map[string]any{"FilePaths.DataFiles": dataDir}); err != nil {
		panic(err)
	}
	keywords.LoadAliases()
	usercommands.RegisterCommand("zz_world", func(rest string, a *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
		ran = append(ran, a.Id.String()+":"+rest)
		return true, nil
	}, true, false)
	// Turn 0 is the "run at once" ready turn; start the clock past it.
	for i := 0; i < 10; i++ {
		util.IncrementTurnCount()
	}
	os.Exit(m.Run())
}

func newTestWorld() *World {
	return &World{
		ignoreInput:  map[ids.ID]uint64{},
		inputTracker: map[ids.ID]struct{}{},
	}
}

// setupInput puts a room in memory with an NPC actor and a player actor in it.
func setupInput(t *testing.T) (npc, player *actors.Actor) {
	t.Helper()
	actors.Reset()
	ran = nil
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = worldTestRoomId, "worldzone"
	rooms.AddTestRoom(r)

	m := mobs.NewTestInstance(mobs.Mob{})
	m.Character.RoomId = worldTestRoomId
	m.Character.Name = "testrat"
	npc = actors.Get(m.Character.Id)
	require.NotNil(t, npc)

	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.RoomId = worldTestRoomId
	u := &accounts.Account{UserId: 9301, Character: c, Password: "$2a$test"} // hashed: no plaintext-password gate
	player = actors.Spawn(c, actors.NewConnectionController(u))

	t.Cleanup(func() {
		rooms.RemoveTestRoom(worldTestRoomId)
		actors.Reset()
		ran = nil
	})
	return npc, player
}

func input(actor *actors.Actor, text string, readyTurn uint64, flags ...events.EventFlag) events.Input {
	in := events.Input{ActorId: actor.Id, InputText: text, ReadyTurn: readyTurn}
	for _, f := range flags {
		in.Flags.Add(f)
	}
	return in
}

func TestNPCInputReachesTheSharedRegistry(t *testing.T) {
	npc, _ := setupInput(t)
	w := newTestWorld()

	assert.Equal(t, events.Continue, w.HandleInputEvents(input(npc, "zz_world  hello ", 0)))
	assert.Equal(t, []string{npc.Id.String() + ":hello"}, ran)
}

func TestInputForUnknownActorIsDropped(t *testing.T) {
	setupInput(t)
	w := newTestWorld()

	in := events.Input{ActorId: ids.New(ids.NoCell), InputText: "zz_world x", ReadyTurn: util.GetTurnCount()}
	assert.NotPanics(t, func() {
		assert.Equal(t, events.Continue, w.HandleInputEvents(in))
	})
	assert.Empty(t, ran)
	assert.Empty(t, w.inputTracker, "a dropped input holds nothing back")
}

func TestNPCUnknownCommandIsLoggedAndShowsNothing(t *testing.T) {
	npc, _ := setupInput(t)
	w := newTestWorld()

	var queued []events.Event
	msgId := events.RegisterListener(events.Message{}, func(e events.Event) events.ListenerReturn {
		queued = append(queued, e)
		return events.Continue
	})
	defer events.UnregisterListener(events.Message{}, msgId)
	inId := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		queued = append(queued, e)
		return events.Cancel // keep it from the global world's router
	}, events.First)
	defer events.UnregisterListener(events.Input{}, inId)
	events.ProcessEvents() // drop anything queued earlier
	queued = nil

	w.HandleInputEvents(input(npc, "zzbogus some words", 0))
	events.ProcessEvents()

	assert.Empty(t, queued, "no room text and no confused emote")
	assert.True(t, testLogs.has("WARN", "unknown command"))
	assert.True(t, testLogs.has("WARN", "zzbogus"))
}

func TestNPCDelayedCommandHoldsBackTheNext(t *testing.T) {
	npc, _ := setupInput(t)
	w := newTestWorld()
	now := util.GetTurnCount()

	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(input(npc, "zz_world first", now+5)))
	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(input(npc, "zz_world second", 0)),
		"an immediate command waits behind the delayed one")
	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(input(npc, "zz_world third", now)),
		"a ready command waits behind the delayed one")
	assert.Empty(t, ran)

	// Next pass, once the first is ready: it runs, and only it.
	clear(w.inputTracker)
	for i := 0; i < 5; i++ {
		util.IncrementTurnCount()
	}
	assert.Equal(t, events.Continue, w.HandleInputEvents(input(npc, "zz_world first", now+5)))
	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(input(npc, "zz_world second", 0)))
	assert.Equal(t, []string{npc.Id.String() + ":first"}, ran)
}

func TestPlayerReadyCommandOvertakesADelayedOne(t *testing.T) {
	_, player := setupInput(t)
	w := newTestWorld()
	now := util.GetTurnCount()

	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(input(player, "zz_world later", now+5)))
	assert.Equal(t, events.Continue, w.HandleInputEvents(input(player, "zz_world ready", now)))
	assert.Equal(t, []string{player.Id.String() + ":ready"}, ran)

	// One queued command per pass: the next ready one waits for the next pass.
	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(input(player, "zz_world again", now)))
	// Ready turn 0 bypasses the limit.
	assert.Equal(t, events.Continue, w.HandleInputEvents(input(player, "zz_world now", 0)))
	assert.Equal(t, []string{player.Id.String() + ":ready", player.Id.String() + ":now"}, ran)
}

// The Inv C 1c edge: a CmdBlockInputUntilComplete command first seen already
// ready never records its block, so the account would stay blocked. The only
// producer (go's exit message) waits one second, which is at least one turn at
// any TurnMs the config accepts, so it is first seen waiting and the block is
// recorded and lifted.
func TestBlockInputUntilCompleteUnblocksAfterAOneSecondWait(t *testing.T) {
	_, player := setupInput(t)
	w := newTestWorld()
	user := player.Controller().Account()

	wait := uint64(configs.GetTimingConfig().SecondsToTurns(1))
	require.GreaterOrEqual(t, wait, uint64(1))

	now := util.GetTurnCount()
	user.BlockInput() // what queueFlagged does when it queues the command
	in := input(player, "zz_world north", now+wait, events.CmdBlockInputUntilComplete)

	assert.Equal(t, events.CancelAndRequeue, w.HandleInputEvents(in))
	assert.Contains(t, w.ignoreInput, player.Id)
	assert.True(t, user.InputBlocked())

	clear(w.inputTracker)
	for i := uint64(0); i < wait; i++ {
		util.IncrementTurnCount()
	}
	assert.Equal(t, events.Continue, w.HandleInputEvents(in))
	assert.False(t, user.InputBlocked())
	assert.NotContains(t, w.ignoreInput, player.Id)
	assert.Equal(t, []string{player.Id.String() + ":north"}, ran)
}
