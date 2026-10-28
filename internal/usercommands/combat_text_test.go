package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addPlayer spawns one more player body in room.
func addPlayer(t *testing.T, userId int, name string, room *rooms.Room) *actors.Actor {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = name
	c.RoomId = room.RoomId
	u := &accounts.Account{UserId: userId, Character: c, Password: "$2a$test"}
	a := actors.Spawn(c, actors.NewConnectionController(u))
	room.AddPlayer(userId)
	return a
}

func TestHiddenAttackerOpeningLineReadsSomeoneToObservers(t *testing.T) {
	_, player, room := goSetup(t)
	player.Character.Name = "Hero"
	observer := addPlayer(t, 9202, "Watcher", room)

	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{BuffId: 9100, Name: "Test hidden", TriggerRate: "1 round", TriggerCount: 5, Flags: []string{"hidden"}}))
	require.NoError(t, player.Character.AddBuff(9100, true))
	require.True(t, player.Character.HasBuffFlag("hidden"))

	bat := mobs.NewTestInstance(mobs.Mob{})
	bat.Character.Name = "Bat"
	bat.Character.RoomId = testRoomId
	bat.Character.Health = 10
	room.AddMob(bat.Character.Id)

	var told, seen []string
	seen = playerTexts(t, observer, func() {
		told = playerTexts(t, player, func() {
			handled, err := Attack("Bat", player, room, 0)
			require.NoError(t, err)
			assert.True(t, handled)
		})
	})
	_ = told
	assert.Equal(t, []string{`Someone prepares to fight <ansi fg="mobname">Bat</ansi>.` + "\n"}, seen)
}

func TestAttackerReadsItsOwnLineAndObserversReadTheName(t *testing.T) {
	_, player, room := goSetup(t)
	player.Character.Name = "Hero"
	observer := addPlayer(t, 9202, "Watcher", room)

	bat := mobs.NewTestInstance(mobs.Mob{})
	bat.Character.Name = "Bat"
	bat.Character.RoomId = testRoomId
	bat.Character.Health = 10
	room.AddMob(bat.Character.Id)

	var own []string
	seen := playerTexts(t, observer, func() {
		own = playerTexts(t, player, func() {
			_, err := Attack("Bat", player, room, 0)
			require.NoError(t, err)
		})
	})
	assert.Equal(t, []string{`You prepare to fight <ansi fg="mobname">Bat</ansi>!` + "\n"}, own)
	assert.Equal(t, []string{`<ansi fg="username">Hero</ansi> prepares to fight <ansi fg="mobname">Bat</ansi>.` + "\n"}, seen)
}

func TestNPCDeathRoomLineReachesAPlayerInTheRoom(t *testing.T) {
	npc, player, room := goSetup(t)

	told := playerTexts(t, player, func() {
		handled, err := Suicide("", npc, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})
	assert.Equal(t, []string{`<ansi fg="mobname">Rat</ansi> has died.` + "\n"}, told)
}

func TestShootOpeningLineStaysInTheShootersRoom(t *testing.T) {
	_, player, room := goSetup(t)
	player.Character.Name = "Hero"
	observer := addPlayer(t, 9202, "Watcher", room)
	next := rooms.LoadRoom(testNextRoomId)
	far := addPlayer(t, 9203, "Faraway", next)

	var seen, farSeen []string
	farSeen = playerTexts(t, far, func() {
		seen = playerTexts(t, observer, func() {
			sendShootOpening(player, `<ansi fg="mobname">Bat</ansi>`, "north")
		})
	})
	assert.Equal(t,
		[]string{`<ansi fg="username">Hero</ansi> prepares to shoot at <ansi fg="mobname">Bat</ansi> through the <ansi fg="exit">north</ansi> exit.` + "\n"},
		seen)
	assert.Empty(t, farSeen, "the target's room reads nothing")
}
