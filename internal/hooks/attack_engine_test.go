package hooks

import (
	"slices"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/stretchr/testify/require"
)

// The defender's wimpy check runs after damage: with wimpy above any health
// percentage, Alice queues a flee once the guard hurts her.
func TestAttackRunsTheDefendersWimpyCheck(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	alice.SetConfigOption(`wimpy`, 101)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)

	require.Less(t, alice.Character.Health, 12, "the fixture must damage Alice")
	require.Contains(t, w.out.String(), `Input Alice: "flee"`)
}

// harmless is the active ruleset where every attack connects for 0.
type harmless struct{ rules.Ruleset }

func (harmless) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
	return rules.AttackResult{Hit: true}
}

// The wimpy check runs only on damage: a 0-damage round queues no flee.
func TestAttackSkipsTheWimpyCheckWithoutDamage(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	alice.SetConfigOption(`wimpy`, 101)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(harmless{rules.Get()}))
	w.start()
	w.rounds(3)

	require.NotContains(t, w.out.String(), `"flee"`)
}

// "Your rage subsides" is the ruleset's call: an attacker whose target is no
// longer up when its turn comes stops. Under oneShot the first of the two to
// swing takes the rat out before the other's turn. Which goes first is map
// order (fighters, NewRound_DoCombat.go:47-60), so the test finds out.
func TestRageSubsidesWhenTheTargetIsNotUp(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: the kill's XP may level the killer up and save her
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 30, 30, 20, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "rage-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	bob.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(oneShot{rules.Get()}))
	w.start()
	w.rounds(1)

	first, second := alice, bob
	if !slices.Contains(w.reads[alice.UserId], "STUB HIT") {
		first, second = bob, alice
	}
	require.Contains(t, w.reads[first.UserId], "STUB HIT", "one of them swung first")
	require.Contains(t, w.reads[second.UserId], "Your rage subsides.")
	require.NotContains(t, w.reads[second.UserId], "STUB HIT", "the other did not swing at a body that is not up")
}
