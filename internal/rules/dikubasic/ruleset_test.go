package dikubasic

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init() // NPC instances get actors through its hooks
	os.Exit(m.Run())
}

// npc is a test NPC actor with the given Perception and Speed.
func npc(t *testing.T, perception, speed int) *actors.Actor {
	t.Helper()
	m := mobs.NewTestInstance(mobs.Mob{})
	id := m.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	a := actors.Get(id)
	require.NotNil(t, a, "actors.Init wires NPC instances to actors")
	a.Character.Stats.Perception.ValueAdj = perception
	a.Character.Stats.Speed.ValueAdj = speed
	return a
}

// player is a test actor with an account.
func player(t *testing.T, userId int) *actors.Actor {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	u := &accounts.Account{UserId: userId, Character: c, Password: "$2a$test"}
	a := actors.Spawn(c, actors.NewConnectionController(u))
	t.Cleanup(func() { actors.Remove(a.Id) })
	return a
}

func TestTurnOrderPutsAccountsFirstKeepingOrder(t *testing.T) {
	n1, p1, n2, p2 := npc(t, 0, 0), player(t, 9601), npc(t, 0, 0), player(t, 9602)
	got := New().TurnOrder([]*actors.Actor{n1, p1, n2, p2})
	assert.Equal(t, []*actors.Actor{p1, p2, n1, n2}, got)
}

// matchesOneRoll checks, over many seeds, that roll() answers exactly
// util.Rand(100) < chance from the same seed and makes exactly one draw (the
// next draw after it is the second draw of the stream).
func matchesOneRoll(t *testing.T, chance int, roll func() bool) {
	t.Helper()
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) })
	for seed := int64(1); seed <= 1000; seed++ {
		util.SeedRandForTest(seed)
		want := util.Rand(100) < chance
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		require.Equal(t, want, roll(), "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: exactly one draw", seed)
	}
}

func TestFindsWeaponRollsUnderPerception(t *testing.T) {
	a := npc(t, 50, 0)
	matchesOneRoll(t, 50, func() bool { return New().FindsWeapon(a) })
}

func TestFleeEscapesOnSpeedShareOf70Plus30(t *testing.T) {
	// Unequal Speed, so swapping fleer and blocker changes the chance:
	// int(10/(10+30)*70) = 17, plus 30.
	fleer, blocker := npc(t, 0, 10), npc(t, 0, 30)
	matchesOneRoll(t, 47, func() bool { return New().FleeEscapes(fleer, blocker) })
}

func TestHatredRoundsIsTwoMinutesLessPerception(t *testing.T) {
	a := player(t, 9603)
	a.Character.Stats.Perception.ValueAdj = 7
	assert.Equal(t, configs.GetConfig().Timing.MinutesToRounds(2)-7, New().HatredRounds(a))
}

func TestCanAttackAllowsAnyNPCTarget(t *testing.T) {
	assert.NoError(t, New().CanAttack(player(t, 9604), npc(t, 0, 0)))
}

// A safe room protects whoever stands in it, and that check comes before
// the PvP settings. The attacker stands in a loaded room and the recovery
// room is a loaded, non-zero room (room 0 aliases to the start room; new
// bodies start in unloaded room -1). PvP is disabled, so a missing check, or
// one moved after CanPvp, reads "PVP is disabled." instead.
func TestCanAttackRefusesATargetInTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94001, 94002
	// RestoreOverrides lays keys over the live config and resets nothing
	// else, so the cleanup names each key's old value (as setConfig does).
	old := map[string]any{
		"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		"GamePlay.PVP.Enabled":           string(configs.GetPVPConfig().Enabled),
		"GamePlay.PVP.MinimumLevel":      int(configs.GetPVPConfig().MinimumLevel),
	}
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["SpecialRooms.DeathRecoveryRoom"] = recovery
	next["GamePlay.PVP.Enabled"] = "disabled"
	next["GamePlay.PVP.MinimumLevel"] = 0
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		for k, v := range old {
			back[k] = v
		}
		require.NoError(t, configs.RestoreOverrides(back))
		require.Equal(t, old["SpecialRooms.DeathRecoveryRoom"], int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "config restored")
	})
	require.Equal(t, recovery, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom))

	for _, id := range []int{here, recovery} {
		r := rooms.NewEmptyRoom()
		r.RoomId, r.Zone = id, "rulesetzone"
		rooms.AddTestRoom(r)
		id := id
		t.Cleanup(func() { rooms.RemoveTestRoom(id) })
	}
	attacker, target := player(t, 9607), player(t, 9608)
	attacker.Character.RoomId, target.Character.RoomId = here, recovery
	assert.EqualError(t, New().CanAttack(attacker, target), "Fighting is not allowed here.")
}

// setConfig sets config keys for one test and puts back the old values
// named in old (RestoreOverrides resets nothing it is not given; as
// setConfig in internal/hooks).
func setConfig(t *testing.T, set, old map[string]any) {
	t.Helper()
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	for k, v := range set {
		next[k] = v
	}
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		for k, v := range old {
			back[k] = v
		}
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

// testRoom puts an empty room in memory for one test.
func testRoom(t *testing.T, id int, zone string) {
	t.Helper()
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = id, zone
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(id) })
}

// recoveryCopy makes roomId (a room testRoom put in memory, in zone) the
// death recovery room, saves its template into a temporary data folder and
// makes a private copy of it, as a death does. It returns the copy's ID.
func recoveryCopy(t *testing.T, roomId int, zone string) int {
	t.Helper()
	tmp := t.TempDir()
	setConfig(t,
		map[string]any{"FilePaths.DataFiles": tmp, "SpecialRooms.DeathRecoveryRoom": roomId},
		map[string]any{
			"FilePaths.DataFiles":            configs.GetFilePathsConfig().DataFiles.String(),
			"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		})
	require.Equal(t, roomId, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "the override took")
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "rooms", zone), 0o755))
	require.NoError(t, rooms.SaveRoomTemplate(*rooms.LoadRoom(roomId)))
	copies, err := rooms.CreateEphemeralRoomIds(roomId)
	require.NoError(t, err)
	id := copies[roomId]
	require.True(t, rooms.IsEphemeralRoomId(id))
	// Free the copy's ephemeral chunk (one of 100; ephemeral.go:14): removing
	// only the room would hold the chunk for the rest of the run.
	t.Cleanup(func() { rooms.TryEphemeralCleanup(id) })
	return id
}

// A target in a private copy of the recovery room is protected as one in
// the room itself. PvP is disabled, so a missed copy reads "PVP is
// disabled." instead.
func TestCanAttackRefusesATargetInACopyOfTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94003, 94004
	testRoom(t, here, "rulesetzone")
	testRoom(t, recovery, "rulesetzone")
	copyId := recoveryCopy(t, recovery, "rulesetzone")
	setConfig(t,
		map[string]any{"GamePlay.PVP.Enabled": "disabled", "GamePlay.PVP.MinimumLevel": 0},
		map[string]any{
			"GamePlay.PVP.Enabled":      string(configs.GetPVPConfig().Enabled),
			"GamePlay.PVP.MinimumLevel": int(configs.GetPVPConfig().MinimumLevel),
		})
	attacker, target := player(t, 9609), player(t, 9610)
	attacker.Character.RoomId, target.Character.RoomId = here, copyId
	assert.EqualError(t, New().CanAttack(attacker, target), "Fighting is not allowed here.")
}
