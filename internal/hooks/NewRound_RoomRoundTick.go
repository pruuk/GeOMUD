package hooks

import (
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

//
// Room Round Tick: rooms with players in them (slice 6 keeps them only).
//

func RoomRoundTick(e events.Event) events.ListenerReturn {

	for _, roomId := range rooms.GetRoomsWithPlayers() {
		room := rooms.LoadRoom(roomId)
		if room == nil {
			continue
		}

		room.RoundTick()

		chanceIn100 := 5
		if room.RoomId == -1 {
			chanceIn100 = 20
		}

		var idleMsgs []string
		if len(room.IdleMessages) > 0 {
			idleMsgs = room.IdleMessages
		} else {
			if zCfg := rooms.GetZoneConfig(room.Zone); zCfg != nil {
				if len(zCfg.IdleMessages) > 0 {
					idleMsgs = zCfg.IdleMessages
				}
			}
		}

		idleMsgCt := len(idleMsgs)
		if idleMsgCt > 0 && util.Rand(100) < chanceIn100 {

			if targetRoomId, err := strconv.Atoi(idleMsgs[0]); err == nil {
				idleMsgCt = 0
				if tgtRoom := rooms.LoadRoom(targetRoomId); tgtRoom != nil {
					idleMsgs = tgtRoom.IdleMessages
					idleMsgCt = len(idleMsgs)
				}
			}

			if idleMsgCt > 0 {
				// pick a random message
				idleMsgIndex := uint8(util.Rand(idleMsgCt))

				// If it's a repeating message, treat it as a non-message
				// (Unless it's the only one)
				if idleMsgIndex != room.LastIdleMessage || idleMsgCt == 1 {

					room.LastIdleMessage = idleMsgIndex

					msg := idleMsgs[idleMsgIndex]
					if msg != `` {
						room.SendText(msg)
					}
				}
			}
		}
	}

	return events.Continue
}
