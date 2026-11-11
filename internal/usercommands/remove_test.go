package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A curse holds only a body the ruleset calls up; a body that is not up can
// shed a cursed item (the old player death removed worn gear this way).
func TestCursedItemComesOffOnlyABodyThatIsNotUp(t *testing.T) {
	_, player, room := goSetup(t)
	player.Character.Equipment.Ring = testItem(92720, "testring", items.ItemSpec{Type: items.Ring, Subtype: items.Wearable, Cursed: true})

	told := playerTexts(t, player, func() {
		_, err := Remove("testring", player, room, 0)
		require.NoError(t, err)
	})
	require.Len(t, told, 1)
	assert.Contains(t, told[0], "CURSED!")
	assert.Equal(t, 92720, player.Character.Equipment.Ring.ItemId, "up: the curse holds")

	player.Character.Health = 0 // dead under the reference ruleset
	_, err := Remove("testring", player, room, 0)
	require.NoError(t, err)
	assert.Zero(t, player.Character.Equipment.Ring.ItemId, "not up: it comes off")
}
