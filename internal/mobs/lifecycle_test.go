package mobs

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
)

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
