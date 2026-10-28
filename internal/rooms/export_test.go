package rooms

import "testing"

// Seams for the external tests in this directory (package rooms_test), which
// can import packages that import rooms, such as behaviour.

// PersistSetup points the NPC store at a temp dir and empties the room manager
// and the actors.
func PersistSetup(t *testing.T) string { return persistSetup(t) }

// EphemeralRoomIdMinimum is the lowest ephemeral room ID.
func EphemeralRoomIdMinimum() int { return ephemeralRoomIdMinimum }
