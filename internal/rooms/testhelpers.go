package rooms

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// AddTestRoom puts r in memory as LoadRoom finds it, for other packages'
// tests, and panics when a room with that ID is already in memory. It does
// not advance the next room ID (addRoomToMemory's SetNextRoomId can write
// config-overrides.yaml). RemoveTestRoom takes it out. The room manager is
// global, so tests using these seams must not call t.Parallel.
func AddTestRoom(r *Room) {
	if _, ok := roomManager.rooms[r.RoomId]; ok {
		panic(fmt.Sprintf("rooms.AddTestRoom: room %d is already in memory", r.RoomId))
	}
	r.lastVisited = util.GetRoundCount()
	roomManager.rooms[r.RoomId] = r
	if _, ok := roomManager.roomIdToFileCache[r.RoomId]; !ok {
		roomManager.roomIdToFileCache[r.RoomId] = r.Filepath()
	}
	zoneInfo, ok := roomManager.zones[r.Zone]
	if !ok {
		zoneInfo = &ZoneConfig{Name: r.Zone, RoomId: r.RoomId, RoomIds: make(map[int]struct{})}
	}
	zoneInfo.RoomIds[r.RoomId] = struct{}{}
	roomManager.zones[r.Zone] = zoneInfo
}

// RemoveTestRoom takes a test room out of memory, with the zone and file
// cache entries AddTestRoom set.
func RemoveTestRoom(roomId int) { ClearRoomCache(roomId) }

// AddTestDefaultBiome registers a lit `default` biome (what a room with no
// biome falls back to) when none is loaded, for other packages' tests, and
// returns the function that removes it again. Tests that read a room's
// visibility need one: biomes are loaded from files at boot only.
func AddTestDefaultBiome() (remove func()) {
	if _, ok := biomes[`default`]; ok {
		return func() {}
	}
	biomes[`default`] = &BiomeInfo{BiomeId: `default`, Name: `Default`, LitArea: true}
	return func() { delete(biomes, `default`) }
}

// AddTestBody puts a body in r without the move and its events, for other
// packages' tests.
func (r *Room) AddTestBody(id ids.ID) { r.bodies = append(r.bodies, id) }
