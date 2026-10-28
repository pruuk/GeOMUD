package hooks

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Flee sets a target-less Flee aggro. Only the player half of the round
// resolves it, so the NPC half must clear it or the NPC is blocked for good
// (IdleMobs and inFight both treat any aggro as fighting).
func TestFledNPCStopsFighting(t *testing.T) {
	mob, _ := charmTestNPC(t)
	mob.Character.Aggro = &characters.Aggro{Type: characters.Flee}

	var changed []events.AggroChanged
	id := events.RegisterListener(events.AggroChanged{}, func(e events.Event) events.ListenerReturn {
		changed = append(changed, e.(events.AggroChanged))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.AggroChanged{}, id) })
	events.ProcessEvents()
	changed = nil

	handleMobCombat(events.NewRound{})
	events.ProcessEvents()

	require.Nil(t, mob.Character.Aggro, "a fled NPC stops fighting")
	require.Len(t, changed, 1)
	assert.Equal(t, mob.Character.Id, changed[0].ActorId)
}
