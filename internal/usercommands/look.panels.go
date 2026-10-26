package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
)

// buildDescriptionPanel renders the character/description template as a panel.
// It accepts a *characters.Character and the viewing user's id.
func buildDescriptionPanel(c *characters.Character) string {
	layout, err := templates.LoadPanelLayout("character/description")
	if err != nil {
		layout = templates.NewPanelLayout("open", "single", 1, 1)
		layout.AddPanelsToSlot(layout.AddSlot(), "desc")
		layout.Panel("desc").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Description</ansi> `).SetWidth(78)
	}

	panel := layout.Panel("desc")

	panel.Add(``, ``, c.GetDescription())
	panel.Add(``, ``, c.GetHealthAppearance())

	return layout.Render() + term.CRLFStr
}

// buildCorpseDescriptionPanel renders the character/description-corpse template as a panel.
func buildCorpseDescriptionPanel(c *characters.Character) string {
	layout, err := templates.LoadPanelLayout("character/description")
	if err != nil {
		layout = templates.NewPanelLayout("open", "single", 1, 1)
		layout.AddPanelsToSlot(layout.AddSlot(), "desc")
		layout.Panel("desc").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Description</ansi> `).SetWidth(78)
	}

	skulls := `<ansi fg="red-bold">☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠ ☠</ansi>`

	panel := layout.Panel("desc")
	panel.Add(``, ``, skulls)
	panel.Add(``, ``, fmt.Sprintf(`<ansi fg="8">%s</ansi>`, c.GetDescription()))
	panel.Add(``, ``, `<ansi fg="8">This is a corpse. They are dead.</ansi>`)
	panel.Add(``, ``, skulls)

	return layout.Render() + term.CRLFStr
}

// buildInventoryLookPanel renders the character/inventory-look template as a panel.
// equipment is *characters.Worn, itemNames is the list of carried item display names.
func buildInventoryLookPanel(equipment *characters.Worn, itemNames []string) string {
	layout, err := templates.LoadPanelLayout("character/equipment-look")
	if err != nil {
		layout = templates.NewPanelLayout("open", "single", 1, 1)
		layout.AddPanelsToSlot(layout.AddSlot(), "equip")
		layout.Panel("equip").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Equipment</ansi> `).SetWidth(78)
	}
	layout.Panel("equip").SetLabelWidth(9)

	panel := layout.Panel("equip")

	slots := []struct {
		label string
		item  func() *items.Item
	}{
		{`Weapon`, func() *items.Item { return &equipment.Weapon }},
		{`Offhand`, func() *items.Item { return &equipment.Offhand }},
		{`Head`, func() *items.Item { return &equipment.Head }},
		{`Neck`, func() *items.Item { return &equipment.Neck }},
		{`Body`, func() *items.Item { return &equipment.Body }},
		{`Belt`, func() *items.Item { return &equipment.Belt }},
		{`Gloves`, func() *items.Item { return &equipment.Gloves }},
		{`Ring`, func() *items.Item { return &equipment.Ring }},
		{`Legs`, func() *items.Item { return &equipment.Legs }},
		{`Feet`, func() *items.Item { return &equipment.Feet }},
	}

	for _, s := range slots {
		item := s.item()
		if item.IsDisabled() {
			continue
		}
		panel.Add(
			fmt.Sprintf(`<ansi fg="yellow">%s</ansi>`, s.label),
			fmt.Sprintf(`<ansi fg="yellow">%s</ansi>`, s.label),
			fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, item.NameSimple()),
		)
	}

	var sb strings.Builder
	itemCt := len(itemNames)
	switch {
	case itemCt == 0:
		sb.WriteString(`no`)
	case itemCt < 4:
		sb.WriteString(`a few`)
	case itemCt < 7:
		sb.WriteString(`several`)
	default:
		sb.WriteString(`lots of`)
	}
	sb.WriteString(` objects`)

	return layout.Render() + term.CRLFStr + ` Carrying: ` + sb.String() + term.CRLFStr
}

// buildCorpseInventoryPanel renders the lootable contents of a corpse when CorpseItems is enabled.
func buildCorpseInventoryPanel(c *rooms.Corpse) string {
	layout, err := templates.LoadPanelLayout("character/equipment-look")
	if err != nil {
		layout = templates.NewPanelLayout("open", "single", 1, 1)
		layout.AddPanelsToSlot(layout.AddSlot(), "inv")
		layout.Panel("inv").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Inventory</ansi> `).SetWidth(78)
	}
	layout.Panel("equip").SetTitle(` <ansi fg="black-bold">.:</ansi><ansi fg="20">Inventory</ansi> `)
	layout.Panel("equip").SetLabelWidth(9)

	panel := layout.Panel("equip")

	hasContent := false

	for _, slot := range characters.AllSlots() {
		itm := c.Character.Equipment.Get(slot)
		if itm == nil || itm.ItemId == 0 {
			continue
		}
		if itm.IsDisabled() {
			continue
		}
		label := characters.SlotLabel(slot)
		panel.Add(
			fmt.Sprintf(`<ansi fg="yellow">%s</ansi>`, label),
			fmt.Sprintf(`<ansi fg="yellow">%s</ansi>`, label),
			fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, itm.DisplayName()),
		)
		hasContent = true
	}

	for _, itm := range c.Items {
		panel.Add(``, ``, fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, itm.DisplayName()))
		hasContent = true
	}

	if c.Gold > 0 {
		panel.Add(``, ``, fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, c.Gold))
		hasContent = true
	}

	if !hasContent {
		panel.Add(``, ``, `<ansi fg="8">Nothing of value remains.</ansi>`)
	}

	return layout.Render() + term.CRLFStr
}
