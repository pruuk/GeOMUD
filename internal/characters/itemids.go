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
// An edited item that carries the ID of an item was had (with the same ItemId)
// is matched to it first, over every slot and held item. Each other edited
// item that stands for an item was had (the same slot first for
// equipment, then the first unused held or worn item of the same ItemId) is
// replaced by that item whole, so it keeps its ID, uses and adjectives; a
// worn item keeps the edit's CanNeverBeRemoved. Any other edited item is new:
// its ID is cleared and Validate gives it one. No two items end with one ID.
func (c *Character) KeepItemIdentity(was *Character) {
	pool := append(append([]items.Item(nil), was.Items...), was.Equipment.GetAllItems()...)
	used := make([]bool, len(pool))

	// claim marks and returns the first unused pool entry that matches.
	claim := func(match func(old items.Item) bool) (items.Item, bool) {
		for i, old := range pool {
			if !used[i] && match(old) {
				used[i] = true
				return old, true
			}
		}
		return items.Item{}, false
	}

	// Pass 1: an edited item that carries the ID of an item the character had
	// (and the same ItemId) is that item. This runs over every slot and held
	// item before any fallback, so a fallback never takes an item that another
	// edited item names by ID.
	slots := AllSlots()
	slotDone := make([]bool, len(slots))
	itemDone := make([]bool, len(c.Items))
	exact := func(edited items.Item) (items.Item, bool) {
		if edited.ItemId < 1 || edited.Id.IsZero() {
			return items.Item{}, false
		}
		return claim(func(old items.Item) bool { return old.ItemId == edited.ItemId && old.Id == edited.Id })
	}
	for si, slot := range slots {
		worn := c.Equipment.Get(slot)
		if worn == nil {
			continue
		}
		lock := worn.CanNeverBeRemoved
		if got, ok := exact(*worn); ok {
			*worn = got
			worn.CanNeverBeRemoved = lock
			slotDone[si] = true
		}
	}
	for i := range c.Items {
		if got, ok := exact(c.Items[i]); ok {
			c.Items[i] = got
			itemDone[i] = true
		}
	}

	// Pass 2: the rest take the same slot's old item first (equipment), then the
	// first unused item of the same ItemId; otherwise they are new.
	take := func(edited items.Item, prefer ids.ID) items.Item {
		if edited.ItemId < 1 {
			return edited.WithoutId()
		}
		if !prefer.IsZero() {
			if got, ok := claim(func(old items.Item) bool { return old.ItemId == edited.ItemId && old.Id == prefer }); ok {
				return got
			}
		}
		if got, ok := claim(func(old items.Item) bool { return old.ItemId == edited.ItemId }); ok {
			return got
		}
		return edited.WithoutId()
	}
	for si, slot := range slots {
		worn := c.Equipment.Get(slot)
		if worn == nil || slotDone[si] {
			continue
		}
		lock := worn.CanNeverBeRemoved
		*worn = take(*worn, was.Equipment.Get(slot).Id)
		if worn.ItemId > 0 {
			worn.CanNeverBeRemoved = lock
		}
	}
	for i := range c.Items {
		if !itemDone[i] {
			c.Items[i] = take(c.Items[i], ids.ID{})
		}
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
