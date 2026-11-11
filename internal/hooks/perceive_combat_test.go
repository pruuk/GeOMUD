package hooks

import (
	"fmt"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/stretchr/testify/require"
)

// A hidden defender cannot be fought by an attacker who does not perceive
// it; an admin perceives it (the one visibility rule, Calabe 2026-11-11).
func TestCombatAsksTheOneVisibilityRule(t *testing.T) {
	for _, admin := range []bool{false, true} {
		t.Run(fmt.Sprintf("admin=%v", admin), func(t *testing.T) {
			w := newCombatWorld(t, 111)
			alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
			rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "seen-rats")
			w.hide(&rat.Character)
			if admin {
				alice.Role = accounts.RoleAdmin
			}
			alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
			w.start()
			w.rounds(1)
			if admin {
				require.NotContains(t, w.reads[alice.UserId], "You can't seem to find your target.", "an admin finds it")
			} else {
				require.Contains(t, w.reads[alice.UserId], "You can't seem to find your target.")
			}
		})
	}
}
