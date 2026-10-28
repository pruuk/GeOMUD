package items

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// makeItem builds a test Item with an inline Spec so no file loading is needed.
func makeItem(id int, name string) Item {
	return Item{
		ItemId: id,
		Id:     ids.New(ids.NoCell),
		Spec:   &ItemSpec{ItemId: id, Name: name, NameSimple: name},
	}
}

func TestItem_NameMatch(t *testing.T) {
	itm := makeItem(1, "Golden Sword")

	tests := []struct {
		input         string
		allowContains bool
		wantPartial   bool
		wantFull      bool
	}{
		{"golden sword", false, true, true},
		{"golden", false, true, false},
		{"sword", false, false, false},
		{"sword", true, true, false},
		{"golden sword", true, true, true},
		{"xyz", false, false, false},
		{"xyz", true, false, false},
		{"GOLDEN", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			partial, full := itm.NameMatch(tt.input, tt.allowContains)
			assert.Equal(t, tt.wantPartial, partial, "partial")
			assert.Equal(t, tt.wantFull, full, "full")
		})
	}
}

func TestItem_NameMatch_EmptyItem(t *testing.T) {
	itm := Item{ItemId: 0}
	partial, full := itm.NameMatch("anything", true)
	assert.False(t, partial)
	assert.False(t, full)
}

func TestFindMatchIn_ExactMatch(t *testing.T) {
	sword := makeItem(1, "sword")
	shield := makeItem(2, "shield")

	partial, full := FindMatchIn("sword", sword, shield)
	assert.Equal(t, Item{}, partial)
	assert.Equal(t, sword.ItemId, full.ItemId)
}

func TestFindMatchIn_PartialMatch(t *testing.T) {
	longsword := makeItem(1, "longsword")
	shield := makeItem(2, "shield")

	partial, full := FindMatchIn("long", longsword, shield)
	assert.Equal(t, longsword.ItemId, partial.ItemId)
	assert.Equal(t, Item{}, full)
}

func TestFindMatchIn_NoMatch(t *testing.T) {
	sword := makeItem(1, "sword")

	partial, full := FindMatchIn("axe", sword)
	assert.Equal(t, Item{}, partial)
	assert.Equal(t, Item{}, full)
}

func TestFindMatchIn_ContainsFallback(t *testing.T) {
	// "sword" is a prefix of the word "sword" in the two-word name "long sword",
	// so the contains pass finds it when the prefix pass on the full name does not.
	longsword := makeItem(1, "long sword")

	partial, full := FindMatchIn("sword", longsword)
	assert.Equal(t, longsword.ItemId, partial.ItemId)
	assert.Equal(t, Item{}, full)
}

func TestFindMatchIn_NthMatch(t *testing.T) {
	sword1 := makeItem(1, "sword")
	sword2 := makeItem(2, "sword")
	sword3 := makeItem(3, "sword")

	_, full := FindMatchIn("sword#2", sword1, sword2, sword3)
	assert.Equal(t, sword2.ItemId, full.ItemId)

	_, full = FindMatchIn("sword#3", sword1, sword2, sword3)
	assert.Equal(t, sword3.ItemId, full.ItemId)
}

func TestFindMatchIn_ByIdToken(t *testing.T) {
	sword := makeItem(10, "sword")
	shield := makeItem(20, "shield")

	_, full := FindMatchIn(sword.Id.String(), shield, sword)
	assert.Equal(t, sword.Id, full.Id)
}

func TestFindMatchIn_IdTokenOfAnItemNotInTheList(t *testing.T) {
	sword := makeItem(10, "sword")
	other := makeItem(10, "sword")

	partial, full := FindMatchIn(sword.Id.String(), other)
	assert.Equal(t, Item{}, partial)
	assert.Equal(t, Item{}, full)
}

// The !N forms are gone (decision 10: people never type IDs). "!10" is not a
// name, so it matches nothing.
func TestFindMatchIn_BangFormsMatchNothing(t *testing.T) {
	sword := makeItem(10, "sword")

	for _, s := range []string{"!10", "!10:" + sword.Id.String()} {
		partial, full := FindMatchIn(s, sword)
		assert.Equal(t, Item{}, partial, s)
		assert.Equal(t, Item{}, full, s)
	}
}

func TestFindMatchIn_EmptyList(t *testing.T) {
	partial, full := FindMatchIn("sword")
	assert.Equal(t, Item{}, partial)
	assert.Equal(t, Item{}, full)
}

func TestItem_HasAdjective(t *testing.T) {
	itm := makeItem(1, "sword")

	assert.False(t, itm.HasAdjective("exploding"))
	itm.SetAdjective("exploding", true)
	assert.True(t, itm.HasAdjective("exploding"))
}

func TestItem_SetAdjective_Remove(t *testing.T) {
	itm := makeItem(1, "sword")
	itm.SetAdjective("exploding", true)
	itm.SetAdjective("exploding", false)
	assert.False(t, itm.HasAdjective("exploding"))
}

func TestItem_SetAdjective_NoDuplicate(t *testing.T) {
	itm := makeItem(1, "sword")
	itm.SetAdjective("exploding", true)
	itm.SetAdjective("exploding", true)
	assert.Equal(t, 1, len(itm.Adjectives))
}

func TestItem_IsDisabled(t *testing.T) {
	assert.True(t, ItemDisabledSlot.IsDisabled())
	assert.False(t, makeItem(1, "sword").IsDisabled())
}

func TestItem_IsCursed(t *testing.T) {
	itm := makeItem(1, "cursed blade")
	itm.Spec.Cursed = true

	assert.True(t, itm.IsCursed())
	itm.Uncurse()
	assert.False(t, itm.IsCursed())
}

func TestItem_Equals(t *testing.T) {
	a := makeItem(1, "sword")
	b := makeItem(1, "sword")

	assert.False(t, a.Equals(b), "different IDs should not be equal")
	assert.True(t, a.Equals(a), "same item should equal itself")
}

func TestValidateGivesAnIdOnlyWhenMissing(t *testing.T) {
	itm := makeItem(1, "sword")
	itm.Id = ids.ID{}
	itm.Validate()
	first := itm.Id
	assert.False(t, first.IsZero())
	itm.Validate()
	assert.Equal(t, first, itm.Id, "an existing ID is never replaced")

	empty := Item{ItemId: 0}
	empty.Validate()
	assert.True(t, empty.Id.IsZero(), "an empty slot is not an item")
}

func TestIdSurvivesYAML(t *testing.T) {
	itm := makeItem(1, "sword")
	b, err := yaml.Marshal(itm)
	require.NoError(t, err)
	assert.Contains(t, string(b), "id: "+itm.Id.String())
	var back Item
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, itm.Id, back.Id)
}

func TestFromTemplateMakesANewItem(t *testing.T) {
	tpl := makeItem(1, "sword")
	tpl.Adjectives = []string{"rusty"}
	tpl.Uses = 3

	a, b := FromTemplate(tpl), FromTemplate(tpl)
	assert.NotEqual(t, tpl.Id, a.Id)
	assert.NotEqual(t, a.Id, b.Id)
	assert.False(t, a.Id.IsZero())
	assert.Equal(t, 3, a.Uses)
	assert.Equal(t, []string{"rusty"}, a.Adjectives)

	a.Adjectives[0] = "shiny"
	a.Spec.Name = "changed"
	assert.Equal(t, "rusty", tpl.Adjectives[0], "the template's adjectives are not shared")
	assert.Equal(t, "sword", tpl.Spec.Name, "the template's overrides are not shared")

	assert.Equal(t, Item{}, FromTemplate(Item{}), "an empty slot stays empty")
	assert.Equal(t, ItemDisabledSlot, FromTemplate(ItemDisabledSlot))
}

func TestWithoutIds(t *testing.T) {
	list := []Item{makeItem(1, "sword"), makeItem(2, "shield")}
	got := WithoutIds(list)
	for _, itm := range got {
		assert.True(t, itm.Id.IsZero())
	}
	assert.False(t, list[0].Id.IsZero(), "the given list is not changed")
	assert.Nil(t, WithoutIds(nil))
}

func TestItem_TempData(t *testing.T) {
	itm := makeItem(1, "sword")

	assert.Nil(t, itm.GetTempData("key"))
	itm.SetTempData("key", "value")
	assert.Equal(t, "value", itm.GetTempData("key"))
	itm.SetTempData("key", nil)
	assert.Nil(t, itm.GetTempData("key"))
}
