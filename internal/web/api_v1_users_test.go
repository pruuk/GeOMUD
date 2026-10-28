package web

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func patchTestUser() *accounts.Account {
	u := &accounts.Account{UserId: 5, Username: "u"}
	u.Character = characters.New()
	a := items.Item{ItemId: 1, Id: ids.New(ids.NoCell), Uses: 4}
	b := items.Item{ItemId: 2, Id: ids.New(ids.NoCell), Uses: 7}
	u.Character.Items = make([]items.Item, 0, 4) // spare capacity, as a live slice has
	u.Character.Items = append(u.Character.Items, a, b)
	return u
}

func TestDecodeUserPatchKeepsItemsAndLeavesTheLiveRecord(t *testing.T) {
	u := patchTestUser()
	before := append([]items.Item(nil), u.Character.Items...)

	got, err := decodeUserPatch(u, []byte(`{"character":{"Items":[{"ItemId":2},{"ItemId":3}]}}`))
	require.NoError(t, err)

	assert.Equal(t, before, u.Character.Items, "the live record is untouched")
	require.Len(t, got.Character.Items, 2)
	assert.Equal(t, before[1].Id, got.Character.Items[0].Id, "the kept item keeps its ID")
	assert.Equal(t, 7, got.Character.Items[0].Uses, "and its state")
	assert.True(t, got.Character.Items[1].Id.IsZero(), "the added item is new")
	assert.Equal(t, 0, got.Character.Items[1].Uses, "and takes nothing from the removed one")
}

func TestDecodeUserPatchWithoutItemsKeepsThem(t *testing.T) {
	u := patchTestUser()
	got, err := decodeUserPatch(u, []byte(`{"Muted":true}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Items, got.Character.Items)
}
