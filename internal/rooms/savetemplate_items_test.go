package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// liveRoomWithItems is room 4101 in memory with one floor and one stashed
// item, both things with IDs.
func liveRoomWithItems(t *testing.T) (*Room, items.Item, items.Item) {
	t.Helper()
	saveTemplateSetup(t)
	live := memRoom(4101)
	floor := items.Item{ItemId: 1, Spec: &items.ItemSpec{ItemId: 1, Name: "rock"}}
	stash := items.Item{ItemId: 2, StashedBy: 7, Spec: &items.ItemSpec{ItemId: 2, Name: "coin purse"}}
	live.AddItem(floor, false)
	live.AddItem(stash, true)
	require.False(t, live.Items[0].Id.IsZero())
	require.False(t, live.Stash[0].Id.IsZero())
	return live, live.Items[0], live.Stash[0]
}

func assertHoldsOnly(t *testing.T, floor, stash items.Item) {
	t.Helper()
	got := getRoomFromMemory(4101)
	require.NotNil(t, got)
	require.Len(t, got.Items, 1, "one floor item, not two with one ID")
	require.Len(t, got.Stash, 1, "one stashed item, not two with one ID")
	assert.Equal(t, floor.Id, got.Items[0].Id)
	assert.Equal(t, stash.Id, got.Stash[0].Id)
}

// TestSaveRoomTemplateOfTheLiveRoomKeepsEachItemOnce: most admin room commands
// pass the live room itself, whose item lists are the live lists.
func TestSaveRoomTemplateOfTheLiveRoomKeepsEachItemOnce(t *testing.T) {
	live, floor, stash := liveRoomWithItems(t)
	require.NoError(t, SaveRoomTemplate(*live))
	assertHoldsOnly(t, floor, stash)
}

// TestSaveRoomTemplateFromATemplateKeepsTheLiveItems: a room loaded from its
// template file (web PATCH, room build) holds the template's items; the room in
// play keeps the items it has, and gains no new ones from the template.
func TestSaveRoomTemplateFromATemplateKeepsTheLiveItems(t *testing.T) {
	_, floor, stash := liveRoomWithItems(t)
	tpl := Room{RoomId: 4101, Zone: persistZone, Title: "edited",
		Items: []items.Item{{ItemId: 3, Spec: &items.ItemSpec{ItemId: 3, Name: "template rock"}}}}
	require.NoError(t, SaveRoomTemplate(tpl))
	assertHoldsOnly(t, floor, stash)
	assert.Equal(t, "edited", getRoomFromMemory(4101).Title)
}
