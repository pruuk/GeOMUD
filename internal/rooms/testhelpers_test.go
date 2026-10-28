package rooms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddTestRoomRefusesADuplicateID(t *testing.T) {
	const id = 987001
	first := NewEmptyRoom()
	first.RoomId, first.Zone = id, "testhelpers-zone"
	AddTestRoom(first)
	defer RemoveTestRoom(id)

	assert.True(t, IsRoomLoaded(id))
	second := NewEmptyRoom()
	second.RoomId, second.Zone = id, "testhelpers-zone"
	assert.Panics(t, func() { AddTestRoom(second) })
	assert.Same(t, first, LoadRoom(id), "the first room stays in memory")
}
