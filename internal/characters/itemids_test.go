package characters

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
)

func testItem(itemId int) items.Item {
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &items.ItemSpec{ItemId: itemId, Name: "thing"}}
}

func TestInstantiateItemsGivesEveryItemANewId(t *testing.T) {
	tpl := New()
	tpl.Items = []items.Item{testItem(1), testItem(2)}
	tpl.Equipment.Weapon = testItem(3)

	a, b := *tpl, *tpl // as NewMobById copies the template
	a.InstantiateItems()
	b.InstantiateItems()

	for i := range tpl.Items {
		assert.NotEqual(t, tpl.Items[i].Id, a.Items[i].Id)
		assert.NotEqual(t, a.Items[i].Id, b.Items[i].Id)
		assert.Equal(t, tpl.Items[i].ItemId, a.Items[i].ItemId)
	}
	assert.NotEqual(t, tpl.Equipment.Weapon.Id, a.Equipment.Weapon.Id)
	assert.NotEqual(t, a.Equipment.Weapon.Id, b.Equipment.Weapon.Id)
	assert.True(t, a.Equipment.Head.Id.IsZero(), "an empty slot stays empty")

	a.Items[0].Uses = 99
	assert.NotEqual(t, 99, tpl.Items[0].Uses, "the template's item list is not shared")
}

func TestStripItemIdsLeavesTheOriginal(t *testing.T) {
	live := New()
	live.Items = []items.Item{testItem(1)}
	live.Equipment.Body = testItem(2)

	file := *live
	file.StripItemIds()

	assert.True(t, file.Items[0].Id.IsZero())
	assert.True(t, file.Equipment.Body.Id.IsZero())
	assert.False(t, live.Items[0].Id.IsZero(), "the original keeps its item IDs")
	assert.False(t, live.Equipment.Body.Id.IsZero())
}

func TestKeepItemIdentity(t *testing.T) {
	was := New()
	ring1, ring2, sword := testItem(1), testItem(1), testItem(2)
	ring1.Uses, ring2.Uses = 4, 7
	sword.Adjectives = []string{"rusty"}
	was.Items = []items.Item{ring1, ring2, sword}
	was.Equipment.Weapon = testItem(3)
	was.Equipment.Body = testItem(4)

	// The admin removed one of the two rings, kept the sword, added an item 5,
	// kept the weapon (and bound it), and swapped the body for item 6. The
	// form cannot say which ring went, so the kept ring is the first one.
	edited := New()
	edited.Items = []items.Item{{ItemId: 2}, {ItemId: 1}, {ItemId: 5}}
	edited.Equipment.Weapon = items.Item{ItemId: 3, CanNeverBeRemoved: true}
	edited.Equipment.Body = items.Item{ItemId: 6}

	edited.KeepItemIdentity(was)

	assert.Equal(t, sword.Id, edited.Items[0].Id)
	assert.Equal(t, []string{"rusty"}, edited.Items[0].Adjectives)
	assert.Equal(t, ring1.Id, edited.Items[1].Id, "the first unused item of that kind")
	assert.Equal(t, 4, edited.Items[1].Uses)
	assert.True(t, edited.Items[2].Id.IsZero(), "a new item gets its ID in Validate")

	assert.Equal(t, was.Equipment.Weapon.Id, edited.Equipment.Weapon.Id)
	assert.True(t, edited.Equipment.Weapon.CanNeverBeRemoved, "the admin's lock applies")
	assert.True(t, edited.Equipment.Body.Id.IsZero(), "a swapped slot holds a new item")
}

func TestKeepItemIdentityNeverCopiesAnId(t *testing.T) {
	was := New()
	was.Items = []items.Item{testItem(1)}

	// A posted ID that matches nothing the character had is dropped, so a
	// form cannot make a second item with an existing ID.
	stray := ids.New(ids.NoCell)
	edited := New()
	edited.Items = []items.Item{{ItemId: 1, Id: was.Items[0].Id}, {ItemId: 1, Id: was.Items[0].Id}, {ItemId: 9, Id: stray}}

	edited.KeepItemIdentity(was)

	assert.Equal(t, was.Items[0].Id, edited.Items[0].Id)
	assert.True(t, edited.Items[1].Id.IsZero(), "one item, one ID")
	assert.True(t, edited.Items[2].Id.IsZero())
}
