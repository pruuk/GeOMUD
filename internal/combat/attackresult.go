package combat

import "github.com/GoMudEngine/GoMud/internal/perception"

// AttackResult is what one Attack did.
type AttackResult struct {
	Hit            bool  // at least one attack connected
	Crit           bool  // at least one attack was a critical hit
	BuffTarget     []int // buffs a crit puts on the defender
	DamageToTarget int   // total damage dealt to the defender

	// Messages are the round's lines, one per swing (a dodge included), in
	// the order they happened; the caller sends them in order through
	// messaging, which resolves {actor} and {target} per receiver.
	Messages []perception.Message
}
