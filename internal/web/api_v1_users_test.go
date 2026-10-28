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

func TestDecodeUserPatchShopDoesNotAliasTheLiveRecord(t *testing.T) {
	u := patchTestUser()
	u.Character.Shop = make(characters.Shop, 0, 4)
	u.Character.Shop = append(u.Character.Shop, characters.ShopItem{ItemId: 1, Price: 10, Quantity: 5}, characters.ShopItem{ItemId: 2, Price: 20})
	before := append(characters.Shop(nil), u.Character.Shop...)

	got, err := decodeUserPatch(u, []byte(`{"character":{"Shop":[{"ItemId":9}]}}`))
	require.NoError(t, err)
	assert.Equal(t, before, u.Character.Shop, "the live shop is untouched")
	require.Len(t, got.Character.Shop, 1)
	assert.Equal(t, characters.ShopItem{ItemId: 9}, got.Character.Shop[0], "no positional merge into the old entry")

	got, err = decodeUserPatch(u, []byte(`{"Muted":true}`))
	require.NoError(t, err)
	assert.Equal(t, before, got.Character.Shop, "an absent shop stays")
	got.Character.Shop[0].Price = 99
	assert.Equal(t, before, u.Character.Shop, "and is a copy")
}

func TestDecodeUserPatchPartialEquipmentKeepsOtherSlots(t *testing.T) {
	u := patchTestUser()
	u.Character.Equipment.Weapon = items.Item{ItemId: 3, Id: ids.New(ids.NoCell), Uses: 3}
	u.Character.Equipment.Head = items.Item{ItemId: 4, Id: ids.New(ids.NoCell), Uses: 6}

	got, err := decodeUserPatch(u, []byte(`{"character":{"Equipment":{}}}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Equipment, got.Character.Equipment, "an empty Equipment object strips nothing")

	got, err = decodeUserPatch(u, []byte(`{"character":{"Equipment":{"weapon":{"ItemId":3}}}}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Equipment.Weapon.Id, got.Character.Equipment.Weapon.Id)
	assert.Equal(t, u.Character.Equipment.Head, got.Character.Equipment.Head, "a slot left out stays")

	got, err = decodeUserPatch(u, []byte(`{"character":{"Equipment":{"Weapon":{"ItemId":0}}}}`))
	require.NoError(t, err)
	assert.Equal(t, 0, got.Character.Equipment.Weapon.ItemId, "a slot sent empty is cleared")
	assert.Equal(t, u.Character.Equipment.Head, got.Character.Equipment.Head)
}

func TestDecodeUserPatchNullMeansAbsent(t *testing.T) {
	u := patchTestUser()
	u.Character.Shop = characters.Shop{{ItemId: 1}}
	u.Character.Equipment.Weapon = items.Item{ItemId: 3, Id: ids.New(ids.NoCell)}
	u.Character.Equipment.Head = items.Item{ItemId: 4, Id: ids.New(ids.NoCell)}

	got, err := decodeUserPatch(u, []byte(`{"character":{"Items":null,"Equipment":null,"Shop":null}}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Items, got.Character.Items)
	assert.Equal(t, u.Character.Equipment, got.Character.Equipment)
	assert.Equal(t, u.Character.Shop, got.Character.Shop)

	got, err = decodeUserPatch(u, []byte(`{"character":{"Equipment":{"Weapon":null,"Head":{"ItemId":4}}}}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Equipment.Weapon, got.Character.Equipment.Weapon, "a null slot stays")
	assert.Equal(t, u.Character.Equipment.Head.Id, got.Character.Equipment.Head.Id)
}

func TestDecodeUserPatchNullCharacterKeepsTheCharacter(t *testing.T) {
	u := patchTestUser()
	u.Character.Skills = map[string]int{"x": 2}
	got, err := decodeUserPatch(u, []byte(`{"character":null}`))
	require.NoError(t, err)
	require.NotNil(t, got.Character)
	assert.Equal(t, u.Character.Items, got.Character.Items)
	assert.Equal(t, u.Character.Skills, got.Character.Skills)
}

func TestDecodeUserPatchMalformedBodyChangesNothing(t *testing.T) {
	u := patchTestUser()
	u.Macros = map[string]string{"a": "b"}

	_, err := decodeUserPatch(u, []byte(`{"Macros":{"x":"y"},"character":5}`))
	assert.Error(t, err)
	assert.Equal(t, map[string]string{"a": "b"}, u.Macros)
}

func TestDecodeUserPatchHeldItemKeepsIdOverWeaponOfSameKind(t *testing.T) {
	u := patchTestUser()
	a := items.Item{ItemId: 3, Id: ids.New(ids.NoCell), Uses: 9, Adjectives: []string{"bk"}}
	u.Character.Items = []items.Item{a}

	got, err := decodeUserPatch(u, []byte(`{"character":{"Equipment":{"Weapon":{"ItemId":3}}}}`))
	require.NoError(t, err)
	require.Len(t, got.Character.Items, 1)
	assert.Equal(t, a.Id, got.Character.Items[0].Id)
	assert.Equal(t, 9, got.Character.Items[0].Uses)
	assert.True(t, got.Character.Equipment.Weapon.Id.IsZero())
	assert.Nil(t, got.Character.Equipment.Weapon.Adjectives)
}

func TestDecodeUserPatchWithoutItemsKeepsThem(t *testing.T) {
	u := patchTestUser()
	got, err := decodeUserPatch(u, []byte(`{"Muted":true}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Items, got.Character.Items)
}
