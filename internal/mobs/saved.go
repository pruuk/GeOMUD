package mobs

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
)

// SavedNPC is a live NPC as saved to disk (Phase 4b decision 13): what it is,
// has and is tangled with. It is separate from Mob so template saves
// (Mob.Save) never gain runtime fields. What the NPC was doing (path, wander
// count, boredom, last idle command, temp data) is not saved.
type SavedNPC struct {
	Template     Mob                   `yaml:"mob"` // the instance's exported fields, including Character
	HomeRoomId   int                   `yaml:"homeroomid"`
	IsElite      bool                  `yaml:"elite,omitempty"`
	PermaBuffIds []int                 `yaml:"permabuffids,omitempty"` // spawn overrides replace the template list
	Aggro        *characters.Aggro     `yaml:"aggro,omitempty"`
	Charmed      *characters.CharmInfo `yaml:"charmed,omitempty"`
}

// Snapshot saves m as it is now.
func Snapshot(m *Mob) SavedNPC {
	return SavedNPC{
		Template:     *m,
		HomeRoomId:   m.HomeRoomId,
		IsElite:      m.IsElite,
		PermaBuffIds: m.Character.PermaBuffIds(),
		Aggro:        m.Character.Aggro,
		Charmed:      m.Character.Charmed,
	}
}

// Mob rebuilds the NPC from its saved form, with what it was doing reset.
func (s SavedNPC) Mob() Mob {
	m := s.Template
	m.HomeRoomId = s.HomeRoomId
	m.IsElite = s.IsElite
	m.Character.SetPermaBuffs(s.PermaBuffIds)
	m.Character.Aggro = s.Aggro
	m.Character.Charmed = s.Charmed
	m.WanderCount, m.BoredomCounter, m.LastIdleCommand = 0, 0, 0
	m.Path = PathQueue{}
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
	m.Character.Validate(true) // permaBuffIds is set, so permabuffs survive
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
