package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setPvP and addFoundPlayer are in pvp_test.go (Task 3).

// giveBow puts a shooting weapon in a's hands.
func giveBow(a *actors.Actor) {
	spec := items.ItemSpec{ItemId: 92101, Name: "testbow", NameSimple: "testbow",
		Type: items.Weapon, Subtype: items.Shooting, Hands: items.OneHanded,
		Damage: items.Damage{Attacks: 1, DiceCount: 1, SideCount: 6}}
	a.Character.Equipment.Weapon = items.Item{ItemId: 92101, Id: ids.New(ids.NoCell), Spec: &spec}
}

// shootSetup is goSetup with a bow on the player and a second player,
// "Target", in the room north.
func shootSetup(t *testing.T) (shooter, target *actors.Actor, here, north *rooms.Room) {
	t.Helper()
	_, shooter, here = goSetup(t)
	giveBow(shooter)
	north = rooms.LoadRoom(testNextRoomId)
	target = addFoundPlayer(t, 9203, "Target", north)
	return shooter, target, here, north
}

func TestShootAtAPlayerRefusedWhenPvPDisabled(t *testing.T) {
	shooter, _, here, _ := shootSetup(t)
	setPvP(t, configs.PVPDisabled)

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"PVP is disabled.\n"}, told)
	assert.Nil(t, shooter.Character.Aggro, "a refused shot starts no fight")
}

// Spec ruling 1: with PvP limited, both rooms must be PvP rooms.
func TestShootIntoASafeRoomIsRefused(t *testing.T) {
	shooter, _, here, north := shootSetup(t)
	setPvP(t, configs.PVPLimited)
	here.Pvp, north.Pvp = true, false
	t.Cleanup(func() { here.Pvp, north.Pvp = false, false })

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"This is not a PVP area.\n"}, told)
	assert.Nil(t, shooter.Character.Aggro)
}

// Spec ruling 1, the other half: no shooting out of a safe room either.
func TestShootOutOfASafeRoomIsRefused(t *testing.T) {
	shooter, _, here, north := shootSetup(t)
	setPvP(t, configs.PVPLimited)
	here.Pvp, north.Pvp = false, true
	t.Cleanup(func() { here.Pvp, north.Pvp = false, false })

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"This is not a PVP area.\n"}, told)
	assert.Nil(t, shooter.Character.Aggro)
}

func TestShootAtAPlayerQueuesAggroChanged(t *testing.T) {
	shooter, target, here, _ := shootSetup(t)
	setPvP(t, configs.PVPEnabled)

	got := capture[events.AggroChanged](t, func() {
		_, err := Shoot("Target north", shooter, here, 0)
		require.NoError(t, err)
	})
	require.Len(t, got, 1)
	assert.Equal(t, shooter.Id, got[0].ActorId)
	assert.Equal(t, here.RoomId, got[0].RoomId, "GMCP picks the room to update from this")
	require.NotNil(t, shooter.Character.Aggro)
	assert.Equal(t, target.Id, shooter.Character.Aggro.Target)
	assert.Equal(t, "north", shooter.Character.Aggro.ExitName)
	assert.Equal(t, characters.TypedShot, shooter.Character.Aggro.Shot, "shoot arms one typed shot")
}

// Shooting an NPC queues AggroChanged, and PvP settings do not refuse it
// (disabled here, to pin CanAttack's argument order in shoot's NPC branch).
func TestShootAtAnNPCQueuesAggroChanged(t *testing.T) {
	npc, shooter, here := goSetup(t) // "Rat" starts here: move it north
	giveBow(shooter)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	setPvP(t, configs.PVPDisabled)

	got := capture[events.AggroChanged](t, func() {
		_, err := Shoot("Rat north", shooter, here, 0)
		require.NoError(t, err)
	})
	require.Len(t, got, 1)
	assert.Equal(t, shooter.Id, got[0].ActorId)
	assert.Equal(t, here.RoomId, got[0].RoomId)
	require.NotNil(t, shooter.Character.Aggro)
	assert.Equal(t, characters.TypedShot, shooter.Character.Aggro.Shot, "shoot arms one typed shot")
}

// An NPC shooting a player is not PvP (Phase 5 slice 4.1, plan Q7): it
// shoots even with PvP disabled.
func TestNPCShootingAPlayerAllowedWhenPvPDisabled(t *testing.T) {
	npc, _, here := goSetup(t) // "Rat" is in this room
	giveBow(npc)
	target := addFoundPlayer(t, 9203, "Target", rooms.LoadRoom(testNextRoomId))
	setPvP(t, configs.PVPDisabled)

	_, err := Shoot("Target north", npc, here, 0)
	require.NoError(t, err)
	require.NotNil(t, npc.Character.Aggro, "an NPC shot is not PvP")
	assert.Equal(t, target.Id, npc.Character.Aggro.Target)
}

// The test above with PvP enabled: the NPC's shot sets the aggro either way
// (an NPC shooting a player is not PvP, ruling 10).
func TestNPCShootingAPlayerAllowedWhenPvPEnabled(t *testing.T) {
	npc, _, here := goSetup(t) // "Rat" is in this room
	giveBow(npc)
	target := addFoundPlayer(t, 9203, "Target", rooms.LoadRoom(testNextRoomId))
	setPvP(t, configs.PVPEnabled)

	_, err := Shoot("Target north", npc, here, 0)
	require.NoError(t, err)
	require.NotNil(t, npc.Character.Aggro)
	assert.Equal(t, target.Id, npc.Character.Aggro.Target)
}

// shoot asks the ruleset before shooting an NPC too. (pacifist is in
// ruleset_stub_test.go, Task 5).
func TestShootAtAnNPCAsksTheRuleset(t *testing.T) {
	npc, shooter, here := goSetup(t)
	giveBow(shooter)
	require.NoError(t, rooms.MoveToRoom(npc.Id, testNextRoomId))
	t.Cleanup(rules.SetForTest(pacifist{rules.Get()}))

	told := playerTexts(t, shooter, func() {
		_, err := Shoot("Rat north", shooter, here, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"STUB REFUSES\n"}, told)
	assert.Nil(t, shooter.Character.Aggro)
}
