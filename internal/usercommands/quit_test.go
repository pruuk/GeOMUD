package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerTargetedByNPCInRoomIsRefusedQuit(t *testing.T) {
	npc, player, room := goSetup(t)
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	require.Nil(t, player.Character.Aggro, "the player has not attacked back")

	var handled bool
	told := playerTexts(t, player, func() {
		var err error
		handled, err = Quit("", player, room, 0)
		require.NoError(t, err)
	})
	assert.True(t, handled)
	assert.Equal(t, []string{"You're too busy to quit right now!\n"}, told)
}
