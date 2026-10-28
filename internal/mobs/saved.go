package mobs

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

// SavedNPC is a live NPC as saved to disk (Phase 4b decision 13): what it is,
// has and is tangled with. It is separate from Mob so template saves
// (SaveMobSpec) never gain runtime fields. What the NPC was doing (path, wander
// count, boredom, last idle command, temp data) is not saved.
type SavedNPC struct {
	Template   Mob                   `yaml:"mob"` // the instance's exported fields, including Character
	HomeRoomId int                   `yaml:"homeroomid"`
	SpawnSlot  int                   `yaml:"spawnslot,omitempty"` // index plus one of the home room slot it fills; zero for none
	IsElite    bool                  `yaml:"elite,omitempty"`
	Aggro      *characters.Aggro     `yaml:"aggro,omitempty"`
	Charmed    *characters.CharmInfo `yaml:"charmed,omitempty"`
}

// Snapshot saves m as it is now.
func Snapshot(m *Mob) SavedNPC {
	return SavedNPC{
		Template:   *m,
		HomeRoomId: m.HomeRoomId,
		SpawnSlot:  m.SpawnSlot,
		IsElite:    m.IsElite,
		Aggro:      m.Character.Aggro,
		Charmed:    m.Character.Charmed,
	}
}

// templatePermaBuffs is the permanent buff list a fresh spawn of this mob
// template gets (decision 17: only the template adds permanent buffs; race
// and worn items add theirs in Validate). A template that is gone gives none.
func templatePermaBuffs(mobId MobId, npc ids.ID) []int {
	spec, ok := mobs[int(mobId)]
	if !ok {
		mudlog.Error("mobs: restored NPC's template is gone; it keeps no template permanent buffs", "mobid", int(mobId), "npc", npc.String())
		return nil
	}
	return append([]int(nil), spec.BuffIds...)
}

// Mob rebuilds the NPC from its saved form, with what it was doing reset.
// Its permanent buffs are its template's, as a fresh spawn's are.
func (s SavedNPC) Mob() Mob {
	m := s.Template
	m.HomeRoomId = s.HomeRoomId
	m.SpawnSlot = s.SpawnSlot
	m.IsElite = s.IsElite
	m.Character.SetPermaBuffs(templatePermaBuffs(m.MobId, m.Character.Id))
	m.Character.Aggro = s.Aggro
	m.Character.Charmed = s.Charmed
	m.WanderCount = 0
	m.Path = PathQueue{}
	m.PreventIdle = false
	m.lastCommandTurn = 0
	m.tempDataStore = nil
	m.Character.ForgetTransient()
	return m
}

// Restore puts a saved NPC back in memory under its saved ID and fires
// OnInstanceCreated. It never makes an ID (decision 11).
func Restore(s SavedNPC) (*Mob, error) {
	m := s.Mob()
	if m.Character.Id.IsZero() {
		return nil, fmt.Errorf("mobs: saved NPC %d has no body ID", m.MobId)
	}
	if _, live := mobInstances[m.Character.Id]; live {
		return nil, fmt.Errorf("mobs: NPC %s is already live", m.Character.Id)
	}
	m.Validate()
	m.Character.Validate(true) // the template's permaBuffIds is set, so its permabuffs apply as on a fresh spawn
	mobInstances[m.Character.Id] = &m
	stored := mobInstances[m.Character.Id]
	OnInstanceCreated.Fire(stored)
	return stored, nil
}

// Rest takes a live NPC out of memory without ending it (decision 15). The
// caller has saved it first. Actors drop it through OnInstanceDestroyed.
func Rest(id ids.ID) {
	if _, ok := mobInstances[id]; !ok {
		return
	}
	delete(mobInstances, id)
	OnInstanceDestroyed.Fire(id)
}
