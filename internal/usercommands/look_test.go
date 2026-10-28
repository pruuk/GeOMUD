package usercommands

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNPCLookTellsTheRoomWithMobnameAndNoOneElse(t *testing.T) {
	t.Cleanup(rooms.AddTestDefaultBiome())
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"

	msgs := captureMessages(t, func() {
		handled, err := Look("", npc, room, 0)
		require.NoError(t, err)
		assert.True(t, handled)
	})
	assert.Equal(t,
		[]string{`<ansi fg="mobname">Rat</ansi> is looking around.` + "\n"},
		messagesFor(msgs, testRoomId))
}

func TestNPCSecretLookQueuesNoMessage(t *testing.T) {
	t.Cleanup(rooms.AddTestDefaultBiome())
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)

	msgs := captureMessages(t, func() {
		handled, err := Look("", npc, room, events.CmdSecretly)
		require.NoError(t, err)
		assert.True(t, handled)
	})
	assert.Empty(t, msgs)
}
