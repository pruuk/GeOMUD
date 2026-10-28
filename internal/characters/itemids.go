package characters

import "github.com/GoMudEngine/GoMud/internal/items"

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
