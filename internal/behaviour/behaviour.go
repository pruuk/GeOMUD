// Package behaviour holds the things an NPC does by itself: wandering, planning
// a path, and resting out of memory. They were mob commands (Phase 4b slice 4);
// they are not commands any more because nobody types them. The body still
// moves through the shared command set: these functions choose, and issue
// "go <exit>" through the NPC's input queue.
package behaviour

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Wander picks an exit for mob as the wander command did and issues
// "go <exit>"; past MaxWander it plans the way home instead.
func Wander(mob *mobs.Mob) {

	// Friendly (charmed) mobs don't wander.
	if mob.Character.IsCharmed() {
		return
	}

	// If they aren't home and have used up their MaxWander moves, go home.
	// -1 means they can wander forever and never go home. 0 means they never wander.
	if mob.Character.RoomId != mob.HomeRoomId {
		if mob.MaxWander > 0 && mob.WanderCount >= mob.MaxWander {
			PathTo(mob, `home`)
			return
		}
	}

	// If MaxWander is zero, they don't wander.
	if mob.MaxWander == 0 {
		return
	}

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return
	}

	if exitName, roomId := room.GetRandomExit(); exitName != `` {
		// Stay in the mob's zone.
		if r := rooms.LoadRoom(roomId); r != nil && r.Zone == mob.Character.Zone {
			mob.WanderCount++
			mob.Command(fmt.Sprintf("go %s", exitName))
		}
	}
}

// PathTo plans mob's path to the listed rooms, or "home", as the pathto
// command did. It does not move: Tick walks the path.
// A path that cannot be found is logged (Decision 3) and marks the NPC lost,
// as before.
func PathTo(mob *mobs.Mob, args string) {

	// If only going home, check whether a path home was already tried and marked as impossible.
	if args == `home` {
		cantGoHome := mob.GetTempData(`home-impossible`)
		if v, ok := cantGoHome.(bool); ok && v {
			// If can't go home, slowly lose health (10%)
			// This helps to clean up mobs that get stuck in a weird location, which can
			// happen for any number of reasons, like players dragging them through portals
			mob.Character.Health -= int(math.Ceil(float64(mob.Character.HealthMax.Value) / 10))
			return
		}
	}

	toRoomIds := []int{}

	for _, roomIdStr := range util.SplitButRespectQuotes(strings.ToLower(args)) {

		if roomIdStr == `home` {
			toRoomIds = append(toRoomIds, mob.HomeRoomId)
			continue
		}

		if roomIdInt, err := strconv.Atoi(roomIdStr); err == nil {
			toRoomIds = append(toRoomIds, roomIdInt)
		}
	}

	if len(toRoomIds) < 1 || toRoomIds[0] == mob.Character.RoomId {
		return
	}

	path, err := mapper.GetPath(mob.Character.RoomId, toRoomIds...)
	if err != nil {
		mudlog.Warn("pathto", "mobname", mob.Character.Name, "from", mob.Character.RoomId, "to", args, "error", err)
		if args == `home` {
			mob.SetTempData(`home-impossible`, true)
			mob.Character.SetAdjective(`lost`, true)
		}
		return
	}

	if args == `home` && len(path) == 0 {
		mob.SetTempData(`home-impossible`, true)
		mob.Character.SetAdjective(`lost`, true)
		return
	}

	newPath := make([]mobs.PathRoom, len(path))

	// Copy everything over
	for idx, p := range path {
		newPath[idx] = p
	}

	mob.Path.SetPath(newPath)
}

// Rest takes mob out of memory without ending it (slice 2b decision 15), as
// the rest command did: an NPC in an ephemeral room ends there; any other
// rests in its room. It comes back only when that room is next loaded from
// disk. Boredom and charm expiry use it. In an ephemeral room, where nothing
// is saved, it cannot rest and ends as a death would: a slot that owns it
// refills after its respawn rate.
func Rest(mob *mobs.Mob, reason string) {

	mudlog.Info("rest", "mobname", mob.Character.Name, "reason", reason)

	room := rooms.LoadRoom(mob.Character.RoomId)
	if room == nil {
		return
	}

	if room.IsEphemeral() {
		room.EndEphemeralNPC(mob)
		if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
			r.CleanupMobSpawns()
		}
		return
	}

	// A failed save is logged by RestNPC; the NPC stays live.
	room.RestNPC(mob)
}

// RunIdle runs one idle command from data: "wander" and "pathto ..." are
// behaviour actions; anything else is queued as the NPC's input.
func RunIdle(mob *mobs.Mob, command string) {
	verb, args, _ := strings.Cut(strings.TrimSpace(command), ` `)
	switch strings.ToLower(verb) {
	case `wander`:
		Wander(mob)
	case `pathto`:
		PathTo(mob, strings.TrimSpace(args))
	default:
		mob.Command(command)
	}
}
