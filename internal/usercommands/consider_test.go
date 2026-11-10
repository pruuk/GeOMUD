package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A target found by name whose actor is gone is not found, not a crash.
func TestConsiderATargetWithNoActorIsNotFound(t *testing.T) {
	npc, player, room := goSetup(t) // the NPC is "Rat"
	actors.Remove(npc.Id)

	told := playerTexts(t, player, func() {
		_, err := Consider("Rat", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{`<ansi fg="mobname">Rat</ansi> can't be found.` + "\n"}, told)
}
