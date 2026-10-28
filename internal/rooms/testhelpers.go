package rooms

import "github.com/GoMudEngine/GoMud/internal/ids"

// AddTestRoom puts r in memory as LoadRoom finds it, for other packages'
// tests. RemoveTestRoom takes it out.
func AddTestRoom(r *Room) { addRoomToMemory(r, true) }

// RemoveTestRoom takes a test room out of memory, with the zone and file
// cache entries AddTestRoom set.
func RemoveTestRoom(roomId int) { ClearRoomCache(roomId) }

// AddTestBody puts a body in r without the move and its events, for other
// packages' tests.
func (r *Room) AddTestBody(id ids.ID) { r.bodies = append(r.bodies, id) }
