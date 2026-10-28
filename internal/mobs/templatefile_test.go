package mobs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// templateWithItems is a mob template whose held and worn items have IDs, as a
// template has in memory after Validate.
func templateWithItems(name string) Mob {
	m := Mob{Zone: "templatefiletest"}
	m.Character = *characters.New()
	m.Character.Name = name
	m.Character.Items = []items.Item{{ItemId: 1, Spec: &items.ItemSpec{ItemId: 1, Name: "tail"}}}
	m.Character.Equipment.Weapon = items.Item{ItemId: 2, Spec: &items.ItemSpec{ItemId: 2, Name: "bow", Type: items.Weapon, Subtype: items.Shooting, Hands: 2}}
	m.Character.Validate()
	return m
}

// readMobFile reads the one template file a writer left for m.
func readMobFile(t *testing.T, m *Mob) string {
	t.Helper()
	path := filepath.Join(configs.GetFilePathsConfig().DataFiles.String(), "mobs", filepath.FromSlash(m.Filepath()))
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(b)
}

func forgetMob(t *testing.T, mobId MobId, name string) {
	t.Cleanup(func() {
		delete(mobs, int(mobId))
		delete(mobNameCache, mobId)
		for i, n := range allMobNames {
			if n == name {
				allMobNames = append(allMobNames[:i], allMobNames[i+1:]...)
				break
			}
		}
		os.RemoveAll(filepath.Join(configs.GetFilePathsConfig().DataFiles.String(), "mobs"))
	})
}

// TestMobTemplateWritersWriteNoItemIds: a template's items are not things, so
// neither writer of a mob template file writes an item ID (slice 2c).
func TestMobTemplateWritersWriteNoItemIds(t *testing.T) {
	// getNextMobId finds no free ID with no templates loaded.
	mobs[9100] = &Mob{MobId: 9100}
	t.Cleanup(func() { delete(mobs, 9100) })

	t.Run("CreateNewMobFile", func(t *testing.T) {
		m := templateWithItems("filetestcreate")
		require.False(t, m.Character.Items[0].Id.IsZero(), "the template held item IDs before the write")
		mobId, err := CreateNewMobFile(m)
		require.NoError(t, err)
		forgetMob(t, mobId, m.Character.Name)

		saved := GetMobSpec(mobId)
		require.NotNil(t, saved)
		data := readMobFile(t, saved)
		assert.Contains(t, data, "itemid: 1", "the item is in the file")
		assert.NotContains(t, data, "id: 0x", "no item ID reaches the template file")
	})

	t.Run("SaveMobSpec", func(t *testing.T) {
		m := templateWithItems("filetestsave")
		m.MobId = getNextMobId()
		forgetMob(t, m.MobId, m.Character.Name)
		require.NoError(t, SaveMobSpec(&m))

		data := readMobFile(t, &m)
		assert.Contains(t, data, "itemid: 1", "the item is in the file")
		assert.NotContains(t, data, "id: 0x", "no item ID reaches the template file")
		assert.False(t, m.Character.Items[0].Id.IsZero(), "the in-memory template keeps its item IDs")
	})
}
