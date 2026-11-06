package characters

import "github.com/GoMudEngine/GoMud/internal/ids"

type AggroType int

const (
	// Enumerated Aggro Types
	DefaultAttack AggroType = iota // Regular H2H combat, everything can decay to this. Starts at zero
	Shooting
	BackStab
	Flee
)

type Aggro struct {
	Type          AggroType
	Target        ids.ID // The body being attacked (player or NPC)
	ExitName      string // For example, firing a weapon in a direction
	RoundsWaiting int    // How many rounds must pass before this triggers
	// FleeFrom is, for a Flee, the target the actor had when it fled (zero
	// if none). A blocked or exitless flee resumes the fight with it.
	FleeFrom ids.ID
}
