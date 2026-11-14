package characters

import "github.com/GoMudEngine/GoMud/internal/ids"

type AggroType int

const (
	// Enumerated Aggro Types
	DefaultAttack AggroType = iota // Regular H2H combat, everything can decay to this. Starts at zero
	Shooting
	Flee
)

// ShotMark is the shot an aggro holds (Phase 5 slice 4.1, ruling 13: one
// shot per action). A body never holds more than one.
type ShotMark int

const (
	NoShot     ShotMark = iota
	TypedShot           // armed by the shoot command, a player's or an NPC's
	EarnedShot          // earned by being shot by a typed shoot; earns nothing back
)

type Aggro struct {
	Type          AggroType
	Target        ids.ID // The body being attacked (player or NPC)
	ExitName      string // For example, firing a weapon in a direction
	RoundsWaiting int    // How many rounds must pass before this triggers
	// FleeFrom is, for a Flee, the target the actor had when it fled (zero
	// if none). A blocked or exitless flee resumes the fight with it.
	FleeFrom ids.ID
	// Since is the round this aggro began on its target: set when the target
	// changes, kept when the same target is re-armed (SetAggro). The longest
	// attacker is the one a body turns to next (Phase 5 slice 4.1, ruling 7).
	// An NPC sidecar saved before 4.1 loads 0, which reads as longest.
	Since uint64
	// OutOfReach counts this aggro's consecutive turns that found its target
	// out of reach (ruling 9's wait); a turn that finds it reachable, in the
	// room or through an exit, resets it. At the configured limit the aggro
	// ends (ruling 12, GamePlay.OutOfReachRounds). NPC sidecars save it with
	// the aggro, so a restored NPC resumes its count; a sidecar saved before
	// 4.1 loads 0.
	OutOfReach int
	// OutOfReachTold records that this wait has told the body "Your target
	// is out of reach.", so the line comes once per wait (plan review round
	// 6). Only a turn that finds the target reachable and perceived clears
	// it, not being attacked by the target (which resets OutOfReach). Saved
	// in NPC sidecars with OutOfReach.
	OutOfReachTold bool
	// Shot is the one shot this aggro holds through an exit (ruling 13):
	// armed by the shoot command (TypedShot), or earned by being shot by a
	// typed shoot (EarnedShot). The turn fires it once (lookup step 3) and
	// the re-arm after the swing spends it; one that cannot be fired is
	// dropped. Only a fired TypedShot earns its target a shot back. Saved in
	// NPC sidecars; a sidecar saved before 4.1 loads NoShot.
	Shot ShotMark
}
