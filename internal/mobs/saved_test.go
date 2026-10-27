package mobs

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestSnapshotRoundTripsWhatDecision13Keeps(t *testing.T) {
	m := Mob{MobId: 7, HomeRoomId: 2, IsElite: true, MaxWander: 3, Hostile: true,
		IdleCommands: []string{"emote sniffs"}, WanderCount: 5, BoredomCounter: 9, LastIdleCommand: 2}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.Name = "rat"
	m.Character.Level = 4
	m.Character.Health = 3
	m.Character.Gold = 11
	m.Character.SetPermaBuffs([]int{5, 6})
	target := ids.New(ids.NoCell)
	m.Character.Aggro = &characters.Aggro{Target: target}
	by := ids.New(ids.NoCell)
	m.Character.Charmed = characters.NewCharm(by, 12, "")

	b, err := yaml.Marshal(Snapshot(&m))
	require.NoError(t, err)
	var back SavedNPC
	require.NoError(t, yaml.Unmarshal(b, &back))

	got := back.Mob()
	assert.Equal(t, m.Character.Id, got.Character.Id)
	assert.Equal(t, MobId(7), got.MobId)
	assert.Equal(t, 2, got.HomeRoomId)
	assert.True(t, got.IsElite)
	assert.Equal(t, 3, got.MaxWander)
	assert.True(t, got.Hostile)
	assert.Equal(t, []string{"emote sniffs"}, got.IdleCommands)
	assert.Equal(t, "rat", got.Character.Name)
	assert.Equal(t, 4, got.Character.Level)
	assert.Equal(t, 3, got.Character.Health)
	assert.Equal(t, 11, got.Character.Gold)
	assert.Equal(t, []int{5, 6}, got.Character.PermaBuffIds())
	require.NotNil(t, got.Character.Aggro)
	assert.Equal(t, target, got.Character.Aggro.Target)
	require.NotNil(t, got.Character.Charmed)
	assert.Equal(t, by, got.Character.Charmed.By)
	// what it was doing resets
	assert.Zero(t, got.WanderCount)
	assert.Zero(t, got.BoredomCounter)
	assert.Zero(t, got.LastIdleCommand)
}

func TestSpawnSlotIsSavedButNotInTemplateYAML(t *testing.T) {
	m := Mob{MobId: 7, SpawnSlot: 3}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)

	tpl, err := yaml.Marshal(&m) // what Mob.Save writes
	require.NoError(t, err)
	assert.NotContains(t, string(tpl), "spawnslot", "the template YAML never carries a slot")

	b, err := yaml.Marshal(Snapshot(&m))
	require.NoError(t, err)
	var back SavedNPC
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, 3, back.Mob().SpawnSlot)
}

func TestRestoreKeepsTheIdAndFires(t *testing.T) {
	var created []*Mob
	OnInstanceCreated.Register(func(m *Mob) *Mob { created = append(created, m); return m })
	m := Mob{MobId: 8}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	r, err := Restore(Snapshot(&m))
	require.NoError(t, err)
	assert.Equal(t, m.Character.Id, r.Character.Id)
	assert.Same(t, r, GetInstance(m.Character.Id))
	assert.Contains(t, created, r)
	_, err = Restore(Snapshot(&m))
	assert.Error(t, err, "restoring an NPC that is already live is refused")
	Rest(m.Character.Id)
	assert.Nil(t, GetInstance(m.Character.Id))
}

func TestRestoreRefusesAZeroId(t *testing.T) {
	m := Mob{MobId: 9}
	m.Character = *characters.New()
	_, err := Restore(Snapshot(&m))
	assert.Error(t, err)
}
