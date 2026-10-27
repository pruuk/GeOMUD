package mobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddInstanceAssignsIdStoresAndFires(t *testing.T) {
	var created []*Mob
	OnInstanceCreated.Register(func(m *Mob) *Mob { created = append(created, m); return m })

	m := addInstance(Mob{MobId: 9001})
	assert.NotZero(t, m.InstanceId)
	assert.Same(t, m, GetInstance(m.InstanceId))
	assert.Contains(t, created, m)

	n := addInstance(Mob{MobId: 9001})
	assert.NotEqual(t, m.InstanceId, n.InstanceId)
	DestroyInstance(m.InstanceId)
	DestroyInstance(n.InstanceId)
}

func TestDestroyInstanceFires(t *testing.T) {
	var destroyed []int
	OnInstanceDestroyed.Register(func(id int) int { destroyed = append(destroyed, id); return id })

	m := addInstance(Mob{MobId: 9002})
	DestroyInstance(m.InstanceId)
	assert.Nil(t, GetInstance(m.InstanceId))
	assert.Contains(t, destroyed, m.InstanceId)
}
