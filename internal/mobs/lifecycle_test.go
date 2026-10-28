package mobs

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
)

func TestNPCsFromOneTemplateHaveTheirOwnItems(t *testing.T) {
	tpl := &Mob{MobId: 9003}
	tpl.Character = *characters.New()
	tpl.Character.Items = []items.Item{{ItemId: 1, Spec: &items.ItemSpec{ItemId: 1, Name: "tail"}}}
	// A bow: HandsRequired returns a Shooting weapon's hands from its spec before
	// it looks up the race, so this test needs no race data.
	tpl.Character.Equipment.Weapon = items.Item{ItemId: 2, Spec: &items.ItemSpec{ItemId: 2, Name: "bow", Type: items.Weapon, Subtype: items.Shooting, Hands: 2}}
	tpl.Validate() // as at template load: the template's items get IDs
	mobs[9003] = tpl
	t.Cleanup(func() { delete(mobs, 9003) })

	a := NewMobById(9003, 1)
	b := NewMobById(9003, 1)
	t.Cleanup(func() { DestroyInstance(a.Character.Id); DestroyInstance(b.Character.Id) })

	assert.NotEqual(t, a.Character.Items[0].Id, b.Character.Items[0].Id)
	assert.NotEqual(t, tpl.Character.Items[0].Id, a.Character.Items[0].Id)
	assert.NotEqual(t, a.Character.Equipment.Weapon.Id, b.Character.Equipment.Weapon.Id)
	assert.False(t, a.Character.Items[0].Id.IsZero())
}

func TestAddInstanceAssignsIdStoresAndFires(t *testing.T) {
	var created []*Mob
	OnInstanceCreated.Register(func(m *Mob) *Mob { created = append(created, m); return m })

	m := addInstance(Mob{MobId: 9001})
	assert.False(t, m.Character.Id.IsZero())
	assert.Same(t, m, GetInstance(m.Character.Id))
	assert.Contains(t, created, m)

	n := addInstance(Mob{MobId: 9001})
	assert.NotEqual(t, m.Character.Id, n.Character.Id)
	DestroyInstance(m.Character.Id)
	DestroyInstance(n.Character.Id)
}

func TestDestroyInstanceFires(t *testing.T) {
	var destroyed []ids.ID
	OnInstanceDestroyed.Register(func(id ids.ID) ids.ID { destroyed = append(destroyed, id); return id })

	m := addInstance(Mob{MobId: 9002})
	DestroyInstance(m.Character.Id)
	assert.Nil(t, GetInstance(m.Character.Id))
	assert.Contains(t, destroyed, m.Character.Id)
}
