package characters

import (
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
)

// InstantiateItems gives c, a body just copied from an NPC template, its own
// items: each held and worn item is made new from the template's
// (items.FromTemplate), so no two NPCs share an item or its ID (Phase 4b
// slice 2c). The item list is a new slice, so the template's is untouched.
func (c *Character) InstantiateItems() {
	if c.Items != nil {
		held := make([]items.Item, len(c.Items))
		for i, itm := range c.Items {
			held[i] = items.FromTemplate(itm)
		}
		c.Items = held
	}
	for _, slot := range AllSlots() {
		if worn := c.Equipment.Get(slot); worn != nil {
			*worn = items.FromTemplate(*worn)
		}
	}
}

// KeepItemIdentity is for an edit that names items only by ItemId (the admin
// user form): c holds the edited items, was the character before the edit.
// Each edited item that stands for an item was had (the same slot first for
// equipment, then the first unused held or worn item of the same ItemId) is
// replaced by that item whole, so it keeps its ID, uses and adjectives; a
// worn item keeps the edit's CanNeverBeRemoved. Any other edited item is new:
// its ID is cleared and Validate gives it one. No two items end with one ID.
func (c *Character) KeepItemIdentity(was *Character) {
	pool := append(append([]items.Item(nil), was.Items...), was.Equipment.GetAllItems()...)
	used := make([]bool, len(pool))

	take := func(edited items.Item, prefer ids.ID) items.Item {
		if edited.ItemId < 1 {
			return edited
		}
		pick := -1
		for i, old := range pool {
			if !used[i] && old.ItemId == edited.ItemId && !old.Id.IsZero() && old.Id == prefer {
				pick = i
				break
			}
		}
		if pick < 0 {
			for i, old := range pool {
				if !used[i] && old.ItemId == edited.ItemId {
					pick = i
					break
				}
			}
		}
		if pick < 0 {
			return edited.WithoutId()
		}
		used[pick] = true
		return pool[pick]
	}

	for _, slot := range AllSlots() {
		worn := c.Equipment.Get(slot)
		if worn == nil {
			continue
		}
		lock := worn.CanNeverBeRemoved
		*worn = take(*worn, was.Equipment.Get(slot).Id)
		if worn.ItemId > 0 {
			worn.CanNeverBeRemoved = lock
		}
	}
	for i := range c.Items {
		c.Items[i] = take(c.Items[i], c.Items[i].Id)
	}
}

// StripItemIds clears the IDs of c's held and worn items, for writing c to a
// template file: an item in a template is not a thing yet. Call it on a copy;
// the item list is replaced, so the original's is untouched.
func (c *Character) StripItemIds() {
	c.Items = items.WithoutIds(c.Items)
	for _, slot := range AllSlots() {
		if worn := c.Equipment.Get(slot); worn != nil {
			*worn = worn.WithoutId()
		}
	}
}
