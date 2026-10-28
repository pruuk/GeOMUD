package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNPCAttackNamesItselfInMobnameColour(t *testing.T) {
	npc, _, room := goSetup(t)
	bat := mobs.NewTestInstance(mobs.Mob{})
	bat.Character.Name = "Bat"
	bat.Character.RoomId = testRoomId
	bat.Character.Health = 10
	room.AddMob(bat.Character.Id)

	msgs := captureMessages(t, func() {
		handled, err := Attack("Bat", npc, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})
	require.NotNil(t, npc.Character.Aggro)
	assert.Equal(t, bat.Character.Id, npc.Character.Aggro.Target)
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> prepares to fight <ansi fg="mobname">Bat</ansi>.` + "\n"},
		messagesFor(msgs, testRoomId))
}

func TestNPCNeverAttacksItself(t *testing.T) {
	npc, _, room := goSetup(t) // Rat is the only mob in the room
	for _, rest := range []string{"*mob", "*", "Rat"} {
		for i := 0; i < 30; i++ { // the pick is random
			npc.Character.Aggro = nil
			_, err := Attack(rest, npc, room, 0)
			require.NoError(t, err)
			if npc.Character.Aggro != nil {
				assert.NotEqual(t, npc.Id, npc.Character.Aggro.Target, "attack %q picked the attacker", rest)
			}
		}
	}
}
