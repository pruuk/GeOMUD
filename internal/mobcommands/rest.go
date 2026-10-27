package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// Rest takes the NPC out of memory without ending it (Phase 4b decision 15):
// it rests in its current room and comes back when that room is next
// prepared. Boredom and charm expiry use it. In an ephemeral room, where
// nothing is saved, it leaves as a despawn, as before.
func Rest(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	mudlog.Info("rest", "mobname", mob.Character.Name, "reason", rest)

	if room.IsEphemeral() {
		return Despawn(rest, mob, room)
	}

	// A failed save is logged by RestNPC; the NPC stays live.
	room.RestNPC(mob)

	return true, nil
}
