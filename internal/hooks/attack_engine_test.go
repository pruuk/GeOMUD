package hooks

import (
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
