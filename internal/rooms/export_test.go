package rooms

import (
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

// RestNPC rests one live NPC in this room the way an unloading room does, for
// tests of the resting store: saved into the room's sidecar, out of the room
// and out of memory, listed in the resting index. No ending. It reports false,
// leaving the NPC live, for an ephemeral room or a failed save.
func (r *Room) RestNPC(m *mobs.Mob) bool {
	if r.IsEphemeral() {
		return false
	}
	id := m.Character.Id
	saved := mobs.Snapshot(m)
	// Live in the room, the save writes it; one not listed here is written
	// from its resting entry.
	before := len(r.resting)
	inRoom := false
	for _, b := range r.mobIds() {
		if b == id {
			inRoom = true
			break
		}
	}
	if !inRoom {
		r.resting = append(r.resting, saved)
	}
	if err := saveRoomNPCs(r); err != nil {
		mudlog.Error("npc store: NPC not rested, its room could not be saved", "room", r.RoomId, "npc", id.String(), "error", err)
		r.resting = r.resting[:before]
		return false
	}
	r.RemoveMob(id)
	mobs.Rest(id)
	if inRoom {
		r.resting = append(r.resting, saved)
	}
	indexResting(id, r.RoomId)
	return true
}
