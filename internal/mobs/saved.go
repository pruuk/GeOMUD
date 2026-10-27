package mobs

import "github.com/GoMudEngine/GoMud/internal/characters"

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
