// Round ticks for players
package hooks

import (
	"fmt"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

//
// Player Round Tick
//

func UserRoundTick(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	roomsWithPlayers := rooms.GetRoomsWithPlayers()
	for _, roomId := range roomsWithPlayers {
		// Get rooom
		if room := rooms.LoadRoom(roomId); room != nil {
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

			for _, uId := range room.GetPlayers() {

				user := users.GetByUserId(uId)
				if user == nil {
					continue
				}

				// Roundtick any cooldowns
				user.Character.Cooldowns.RoundTick()

				// Decay alignment toward neutral at the configured interval
				if decayRounds := int(configs.GetGamePlayConfig().AlignmentDecayRounds); decayRounds > 0 && evt.RoundNumber%uint64(decayRounds) == 0 {
					alignmentBefore := user.Character.AlignmentName()
					user.Character.DecayAlignment()
					if alignmentAfter := user.Character.AlignmentName(); alignmentAfter != alignmentBefore {
						before := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
						after := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
						user.SendText(fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, before, after))
						events.AddToQueue(events.AlignmentChanged{
							UserId:       uId,
							AlignmentOld: alignmentBefore,
							AlignmentNew: alignmentAfter,
						})
					}
				}

				if user.Character.Charmed != nil && user.Character.Charmed.RoundsRemaining > 0 {
					user.Character.Charmed.RoundsRemaining--
				}

				if triggeredBuffs := user.Character.Buffs.Trigger(); len(triggeredBuffs) > 0 {

					triggeredBuffIds := []int{}
					for _, buff := range triggeredBuffs {
						triggeredBuffIds = append(triggeredBuffIds, buff.BuffId)
					}

					events.AddToQueue(events.BuffsTriggered{UserId: user.UserId, BuffIds: triggeredBuffIds})
				}

				// Recalculate all stats at the end of the round tick
				hBefore := user.Character.Health
				mBefore := user.Character.Mana
				user.Character.Validate()
				if user.Character.Health != hBefore || user.Character.Mana != mBefore {
					events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
				}

			}

		}

	}

	return events.Continue
}
