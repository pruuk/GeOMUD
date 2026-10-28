package usercommands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRoomId = 92001

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init()
	messaging.Init()
	// TryCommand reads the keyword aliases, which need a data folder.
	dataDir, err := filepath.Abs("../../_datafiles/world/empty")
	if err != nil {
		panic(err)
	}
	if err := configs.AddOverlayOverrides(map[string]any{"FilePaths.DataFiles": dataDir}); err != nil {
		panic(err)
	}
	keywords.LoadAliases()
	os.Exit(m.Run())
}

// setup puts one room in memory, an NPC actor and a player actor in it, and
// registers a recording command `zz_test` and an admin-only `zz_admin`.
func setup(t *testing.T) (npc, player *actors.Actor, ran *[]string) {
	t.Helper()
	actors.Reset()
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = testRoomId, "cmdzone"
	rooms.AddTestRoom(r)

	m := mobs.NewTestInstance(mobs.Mob{})
	m.Character.RoomId = testRoomId
	npc = actors.Get(m.Character.Id)
	require.NotNil(t, npc)

	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.RoomId = testRoomId
	u := &accounts.Account{UserId: 9201, Character: c, Password: "$2a$test"} // hashed, so the plaintext-password gate is not hit
	player = actors.Spawn(c, actors.NewConnectionController(u))

	got := []string{}
	ran = &got
	userCommands["zz_test"] = CommandAccess{Func: func(rest string, a *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
		got = append(got, "zz_test:"+a.Id.String()+":"+rest)
		return true, nil
	}, AllowedWhenDowned: true}
	userCommands["zz_admin"] = CommandAccess{Func: func(rest string, a *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
		got = append(got, "zz_admin")
		return true, nil
	}, AllowedWhenDowned: true, AdminOnly: true}

	t.Cleanup(func() {
		delete(userCommands, "zz_test")
		delete(userCommands, "zz_admin")
		rooms.RemoveTestRoom(testRoomId)
		actors.Reset()
	})
	return npc, player, ran
}

func TestNPCRunsNonAdminCommand(t *testing.T) {
	npc, _, ran := setup(t)
	handled, err := TryCommand("zz_test", " hello ", npc.Id, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, []string{"zz_test:" + npc.Id.String() + ":hello"}, *ran)
}

func TestNPCRefusedAdminCommand(t *testing.T) {
	npc, _, ran := setup(t)
	handled, err := TryCommand("zz_admin", "", npc.Id, 0)
	require.NoError(t, err)
	assert.True(t, handled, "refused silently, as a non-admin player is")
	assert.Empty(t, *ran)
}

func TestPlayerWithoutPermissionRefusedAdminCommand(t *testing.T) {
	_, player, ran := setup(t)
	require.False(t, player.Controller().Account().HasRolePermission("zz_admin", true))
	handled, err := TryCommand("zz_admin", "", player.Id, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Empty(t, *ran)
}

func TestPlayerRunsNonAdminCommand(t *testing.T) {
	_, player, ran := setup(t)
	handled, err := TryCommand("zz_test", "x", player.Id, 0)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, []string{"zz_test:" + player.Id.String() + ":x"}, *ran)
}

func TestTryCommandUnknownActor(t *testing.T) {
	setup(t)
	_, err := TryCommand("zz_test", "", ids.New(ids.NoCell), 0)
	assert.Error(t, err)
}

func TestTellAndUserIdOf(t *testing.T) {
	npc, player, _ := setup(t)
	assert.NotPanics(t, func() { Tell(npc, "nobody hears this") })
	assert.Equal(t, 0, UserIdOf(npc))
	assert.Equal(t, 9201, UserIdOf(player))
}
