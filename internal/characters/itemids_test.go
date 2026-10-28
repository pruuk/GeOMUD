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
