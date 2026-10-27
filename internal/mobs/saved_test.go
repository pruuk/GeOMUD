package mobs

import (
	"fmt"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// withTemplate registers a mob template for the test, as LoadDataFiles would.
func withTemplate(t *testing.T, mobId int, buffIds []int) {
	t.Helper()
	old, had := mobs[mobId]
	mobs[mobId] = &Mob{MobId: MobId(mobId), BuffIds: buffIds}
	t.Cleanup(func() {
		if had {
			mobs[mobId] = old
		} else {
			delete(mobs, mobId)
		}
	})
}

func TestSnapshotRoundTripsWhatDecision13Keeps(t *testing.T) {
	withTemplate(t, 7, []int{5, 6})
	m := Mob{MobId: 7, HomeRoomId: 2, IsElite: true, MaxWander: 3, Hostile: true,
		IdleCommands: []string{"emote sniffs"}, WanderCount: 5, BoredomCounter: 9, LastIdleCommand: 2}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.Name = "rat"
	m.Character.Level = 4
	m.Character.Health = 3
	m.Character.Gold = 11
	m.Character.SetPermaBuffs([]int{5, 6}) // as NewMobById sets from the template
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
	assert.Equal(t, []int{5, 6}, got.Character.PermaBuffIds(), "the template's permanent buffs")
	require.NotNil(t, got.Character.Aggro)
	assert.Equal(t, target, got.Character.Aggro.Target)
	require.NotNil(t, got.Character.Charmed)
	assert.Equal(t, by, got.Character.Charmed.By)
	// what it was doing resets
	assert.Zero(t, got.WanderCount)
	assert.Zero(t, got.BoredomCounter)
	assert.Zero(t, got.LastIdleCommand)
}

func TestMobResetsWhatItWasDoingAsADiskRoundTripWould(t *testing.T) {
	m := Mob{MobId: 7, PreventIdle: true, lastCommandTurn: 99}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.SetTempData("chasing", "someone")
	m.Character.PlayerDamage = map[int]int{3: 40}
	m.Character.LastPlayerDamage = 12
	m.Character.KilledBy = ids.New(ids.NoCell)
	m.Character.KillerMobIsElite = true
	m.Character.KillerMobName = "wolf"

	got := Snapshot(&m).Mob() // an in-memory rest, no disk
	assert.Nil(t, got.GetTempData("chasing"))
	assert.False(t, got.PreventIdle)
	assert.Zero(t, got.lastCommandTurn)
	assert.Empty(t, got.Character.PlayerDamage)
	assert.Zero(t, got.Character.LastPlayerDamage)
	assert.True(t, got.Character.KilledBy.IsZero())
	assert.False(t, got.Character.KillerMobIsElite)
	assert.Empty(t, got.Character.KillerMobName)
	assert.Equal(t, "someone", m.GetTempData("chasing"), "the live NPC it was taken from is untouched")
	assert.Equal(t, 40, m.Character.PlayerDamage[3])
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

func TestRestoreThroughDiskKeepsWhatItHas(t *testing.T) {
	for _, id := range []int{901, 902, 903} {
		t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{BuffId: id, Name: fmt.Sprintf("test buff %d", id), TriggerRate: "1 round", TriggerCount: 5}))
	}
	withTemplate(t, 8, []int{901, 902})

	m := Mob{MobId: 8}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.Level = 20
	m.Character.HealthMax.Base, m.Character.ManaMax.Base = 0, 0 // as a mob loaded from YAML has (they are not saved)
	m.Character.Items = []items.Item{{ItemId: 30001}}
	m.Character.Equipment.Body = items.Item{ItemId: 20001}
	m.Character.SetAdjective("elite", true)
	m.Character.SetPermaBuffs([]int{901, 902}) // the template's, as NewMobById sets them
	m.Character.Validate(true)
	m.Character.AddBuff(903, false) // a buff that runs out
	m.Character.Validate()
	require.Greater(t, m.Character.ManaMax.Value, 1)
	m.Character.Mana = m.Character.ManaMax.Value - 1
	mana := m.Character.Mana
	require.True(t, m.Character.HasBuff(901) && m.Character.HasBuff(902) && m.Character.HasBuff(903))

	b, err := yaml.Marshal(Snapshot(&m))
	require.NoError(t, err)
	var back SavedNPC
	require.NoError(t, yaml.Unmarshal(b, &back))

	r, err := Restore(back)
	require.NoError(t, err)
	t.Cleanup(func() { Rest(r.Character.Id) })

	assert.Equal(t, m.Character.Id, r.Character.Id)
	require.Len(t, r.Character.Items, 1)
	assert.Equal(t, 30001, r.Character.Items[0].ItemId)
	assert.Equal(t, 20001, r.Character.Equipment.Body.ItemId)
	assert.Contains(t, r.Character.Adjectives, "elite")
	assert.Equal(t, mana, r.Character.Mana)
	assert.Equal(t, []int{901, 902}, r.Character.PermaBuffIds())
	assert.True(t, r.Character.HasBuff(901), "template permabuffs survive Validate(true)")
	assert.True(t, r.Character.HasBuff(902))
	assert.True(t, r.Character.HasBuff(903), "and so does a buff that runs out")
}

func TestRestoreCarriesWhatAFreshSpawnOfItsTemplateCarries(t *testing.T) {
	for _, id := range []int{901, 902} {
		t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{BuffId: id, Name: fmt.Sprintf("test buff %d", id), TriggerRate: "1 round", TriggerCount: 5}))
	}
	withTemplate(t, 8, []int{901, 902})
	m := Mob{MobId: 8}
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.SetPermaBuffs([]int{901, 902})
	m.Character.Validate(true)
	saved := Snapshot(&m)

	// The template now has only 901: a fresh spawn would carry 901 alone.
	withTemplate(t, 8, []int{901})
	r, err := Restore(saved)
	require.NoError(t, err)
	t.Cleanup(func() { Rest(r.Character.Id) })
	assert.Equal(t, []int{901}, r.Character.PermaBuffIds())
	r.Character.Buffs.Prune() // as the next turn's prune does
	assert.True(t, r.Character.HasBuff(901))
	assert.False(t, r.Character.HasBuff(902), "a permabuff its template no longer gives goes")
}

func TestRestoreWhoseTemplateIsGoneHasNoTemplatePermabuffs(t *testing.T) {
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{BuffId: 901, Name: "test buff 901", TriggerRate: "1 round", TriggerCount: 5}))
	m := Mob{MobId: 987650} // no such template
	m.Character = *characters.New()
	m.Character.Id = ids.New(ids.NoCell)
	m.Character.SetPermaBuffs([]int{901})
	m.Character.Validate(true)
	require.True(t, m.Character.HasBuff(901))

	r, err := Restore(Snapshot(&m))
	require.NoError(t, err)
	t.Cleanup(func() { Rest(r.Character.Id) })
	assert.Empty(t, r.Character.PermaBuffIds())
	r.Character.Buffs.Prune() // as the next turn's prune does
	assert.False(t, r.Character.HasBuff(901))
}

func TestRestoreRefusesAZeroId(t *testing.T) {
	m := Mob{MobId: 9}
	m.Character = *characters.New()
	_, err := Restore(Snapshot(&m))
	assert.Error(t, err)
}
