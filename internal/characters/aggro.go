package characters

import "github.com/GoMudEngine/GoMud/internal/ids"

type AggroType int

const (
	// Enumerated Aggro Types
	DefaultAttack AggroType = iota // Regular H2H combat, everything can decay to this. Starts at zero
	Shooting
	BackStab
	SpellCast
	Flee
)

type SpellAggroInfo struct {
	SpellId   string
	SpellRest string
}

type Aggro struct {
	Type          AggroType
	Target        ids.ID         // The body being attacked (player or NPC)
	SpellInfo     SpellAggroInfo // If Type is SpellCast, this is the spell info
	ExitName      string         // For example, firing a weapon in a direction
	RoundsWaiting int            // How many rounds must pass before this triggers
}
