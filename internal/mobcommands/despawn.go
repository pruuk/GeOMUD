package mobcommands

import (
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Despawn(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	mudlog.Info("despawn", "mobname", mob.Character.Name, "reason", rest)

	// An admin's despawn ends the NPC (decision 15); its slot respawns at once.
	rooms.NPCEnded(mob, true)

	// Destroy any record of this mob.
	mobs.DestroyInstance(mob.Character.Id)

	// Clean up mob from room...
	if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
		r.CleanupMobSpawns()
	}

	// Remove from current room
	room.RemoveMob(mob.Character.Id)

	return true, nil
}
