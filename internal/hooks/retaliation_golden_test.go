package hooks

// Combat golden fixtures for retaliation and target switching (Phase 5
// slice 4.1, rulings 6-9). Same harness as combat_golden_test.go.

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/require"
)

// Two NPCs attack a player in turn while she fights a third she chose; she
// kills it and turns to the one that attacked first (the rat, Since 1000),
// not the one first in turn order (the cat, MobId 93204, Since 1001).
func TestCombatGoldenSwitchToLongestAttacker(t *testing.T) {
	w := newCombatWorld(t, 1101)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 8)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	rat := w.addNPC(93205, "rat", 10, goldenRoomA, 5, 5, 5, 30)
	alice.Character.SetAggro(dog.Character.Id, characters.DefaultAttack)
	rat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack) // Since 1000
	util.IncrementRoundCount()
	cat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack) // Since 1001
	w.start()
	w.rounds(8)
	w.check("switch_longest")
	require.Equal(t, rat.Character.Id, alice.Character.Aggro.Target, "she turned to the rat")
}

// A player chooses her target; an NPC attacking her dies at another
// player's hand, and her target does not move.
func TestCombatGoldenChosenTargetStays(t *testing.T) {
	w := newCombatWorld(t, 1102)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	carol := w.addPlayer(9303, "Carol", goldenRoomA, 30, 30, 20, 60)
	carol.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 4)
	alice.Character.SetAggro(dog.Character.Id, characters.DefaultAttack)
	rat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	carol.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(4)
	w.check("chosen_target_stays")
	require.Less(t, rat.Character.Health, 1, "the fixture must reach the rat's death")
	require.Equal(t, dog.Character.Id, alice.Character.Aggro.Target, "her chosen target stays")
}

// The target flees and another attacker is in the room: the body switches.
func TestCombatGoldenTargetFleesSwitch(t *testing.T) {
	w := newCombatWorld(t, 1103)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 100, 5, 5, 30)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	cat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	fleeAggro(&rat.Character)
	w.start()
	w.rounds(3)
	w.check("target_flees_switch")
	require.Equal(t, goldenRoomB, rat.Character.RoomId, "the rat fled")
	require.Equal(t, cat.Character.Id, alice.Character.Aggro.Target, "she turned to the cat")
}

// The target flees and nobody else attacks: the aggro waits, saying so
// once, and swings when the target comes back.
func TestCombatGoldenTargetFleesWait(t *testing.T) {
	w := newCombatWorld(t, 1104)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 100, 5, 5, 30)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	fleeAggro(&rat.Character)
	w.start()
	w.rounds(3)
	require.Equal(t, goldenRoomB, rat.Character.RoomId, "the rat fled")
	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomA))
	w.rounds(1)
	w.check("target_flees_wait")
	require.Equal(t, rat.Character.Id, alice.Character.Aggro.Target)
}

// Shot from another room: with a bow the body shoots back through the
// exit once and then waits; a second typed shoot earns a second shot back;
// neither side fires in the rounds between (ruling 13). Without a bow the
// body waits ("Your target is out of reach."); an NPC walks over (its
// queued "go" is run) and swings on arrival.
func TestCombatGoldenShotFromAnotherRoom(t *testing.T) {
	w := newCombatWorld(t, 1105)
	w.runInputs()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomB, 30, 30, 20, 60)
	bob.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	carol := w.addPlayer(9303, "Carol", goldenRoomB, 30, 30, 20, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 10, 5, 5, 60)
	rat.Character.ActionPoints = 100 // a walk costs action points; the harness runs no turn that refills them
	typeShoot(alice.Character, "north", bob.Character.Id)
	w.start()
	w.rounds(3) // Alice's shot and Bob's shot back, then a round with neither
	typeShoot(alice.Character, "north", bob.Character.Id)
	w.rounds(2) // a second typed shoot: a second shot back, then neither
	typeShoot(alice.Character, "north", carol.Character.Id)
	w.rounds(2)
	typeShoot(alice.Character, "north", rat.Character.Id)
	w.rounds(3)
	w.check("shot_from_another_room")
}

// A hidden attacker: the defender takes an aggro on it and cannot find it.
func TestCombatGoldenHiddenAttackerRetaliation(t *testing.T) {
	w := newCombatWorld(t, 1106)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 5, 20, 60)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 60)
	w.hide(alice.Character)
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(2)
	w.check("hidden_attacker_retaliation")
	require.Equal(t, alice.Character.Id, bob.Character.Aggro.Target)
}

// Three bodies target the rat when the round starts. Carol kills it; the
// dog's own turn then finds it not up and turns to Bob, who attacks the dog;
// the rat's death turns Alice to the cat that attacks her and ends Carol's
// fight (she has nobody to turn to).
func TestCombatGoldenDeathReleasesAttackers(t *testing.T) {
	w := newCombatWorld(t, 1107)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 4, 6, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 5, 5, 5, 60)
	carol := w.addPlayer(9303, "Carol", goldenRoomA, 5, 5, 5, 60)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 60)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 1, 1, 1, 1)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	carol.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	dog.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	cat.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	bob.Character.SetAggro(dog.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(1)
	w.check("death_releases_attackers")
	require.Less(t, rat.Character.Health, 1, "the fixture must reach the rat's death")
	require.Equal(t, cat.Character.Id, alice.Character.Aggro.Target)
	require.Nil(t, carol.Character.Aggro)
	require.Equal(t, bob.Character.Id, dog.Character.Aggro.Target)
}

// A wait runs out (ruling 12): the rat flees, nobody else fights Alice, and
// after 5 turns out of reach she loses track of it; Carol sees her give up.
func TestCombatGoldenWaitRunsOut(t *testing.T) {
	w := newCombatWorld(t, 1108)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 60)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 60)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 100, 5, 5, 30)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	fleeAggro(&rat.Character)
	w.start()
	w.rounds(10)
	w.check("wait_runs_out")
	require.Equal(t, goldenRoomB, rat.Character.RoomId, "the rat fled")
	require.Nil(t, alice.Character.Aggro, "she lost track of it")
}
