package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Lock(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		Tell(actor, "Unlock what?")
		return true, nil
	}

	containerName := room.FindContainerByName(args[0])
	exitName, _ := room.FindExitByName(args[0])

	if containerName != `` {

		container := room.Containers[containerName]

		if container.Lock.IsLocked() {
			Tell(actor, "That's already locked.")
			return true, nil
		}

		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, containerName)
		hasKey, _ := actor.Character.HasKey(lockId, int(container.Lock.Difficulty))

		var backpackKeyItm items.Item = items.Item{}
		var hasBackpackKey bool = false
		if !hasKey {
			backpackKeyItm, hasBackpackKey = actor.Character.FindKeyInBackpack(lockId)
		}

		if hasKey {
			container.Lock.SetLocked()
			room.Containers[containerName] = container

			itemRoomEvent(actor,
				fmt.Sprintf(`You use a key to relock the <ansi fg="container">%s</ansi>.`, containerName),
				fmt.Sprintf(`{actor} uses a key to relock the <ansi fg="container">%s</ansi>.`, containerName),
			)
		} else if hasBackpackKey {

			itmSpec := backpackKeyItm.GetSpec()

			container.Lock.SetLocked()
			room.Containers[containerName] = container

			// Key entries look like:
			// "key-<roomid>-<exitname>": "<itemid>"
			actor.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
			actor.Character.RemoveItem(backpackKeyItm)

			events.AddToQueue(events.ItemOwnership{
				ActorId: actor.Character.Id,
				Item:    backpackKeyItm,
				Gained:  false,
			})

			itemRoomEvent(actor,
				fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to lock the <ansi fg="container">%s</ansi>, and add it to your key ring for the future.`, itmSpec.Name, containerName),
				fmt.Sprintf(`{actor} uses a key to lock the <ansi fg="container">%s</ansi>.`, containerName),
			)
		} else {
			Tell(actor, `You do not have the key for that.`)
		}

		return true, nil

	} else if exitName != `` {

		exitInfo, _ := room.GetExitInfo(exitName)

		if exitInfo.Lock.IsLocked() {
			Tell(actor, "That's already locked.")
			return true, nil
		}

		lockId := fmt.Sprintf(`%d-%s`, room.RoomId, exitName)
		hasKey, _ := actor.Character.HasKey(lockId, int(exitInfo.Lock.Difficulty))

		var backpackKeyItm items.Item = items.Item{}
		var hasBackpackKey bool = false
		if !hasKey {
			backpackKeyItm, hasBackpackKey = actor.Character.FindKeyInBackpack(lockId)
		}

		if hasKey {
			exitInfo.Lock.SetLocked()
			room.SetExitLock(exitName, true)

			itemRoomEvent(actor,
				fmt.Sprintf(`You use a key to relock the <ansi fg="exit">%s</ansi> lock.`, exitName),
				fmt.Sprintf(`{actor} uses a key to relock the <ansi fg="exit">%s</ansi> lock`, exitName),
			)
		} else if hasBackpackKey {

			itmSpec := backpackKeyItm.GetSpec()

			exitInfo.Lock.SetLocked()
			room.SetExitLock(exitName, true)

			// Key entries look like:
			// "key-<roomid>-<exitname>": "<itemid>"
			actor.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
			actor.Character.RemoveItem(backpackKeyItm)

			events.AddToQueue(events.ItemOwnership{
				ActorId: actor.Character.Id,
				Item:    backpackKeyItm,
				Gained:  false,
			})

			itemRoomEvent(actor,
				fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to lock the <ansi fg="exit">%s</ansi> exit, and add it to your key ring for the future.`, itmSpec.Name, exitName),
				fmt.Sprintf(`{actor} uses a key to lock the <ansi fg="exit">%s</ansi> exit.`, exitName),
			)
		} else {
			Tell(actor, `You do not have the key for that.`)
		}

		return true, nil

	}

	Tell(actor, "There is no such exit or container.")
	return true, nil

}
