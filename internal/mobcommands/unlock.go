package mobcommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Unlock(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) < 1 {
		return true, nil
	}

	containerName := room.FindContainerByName(args[0])
	exitName, _ := room.FindExitByName(args[0])

	if containerName != `` {

		container := room.Containers[containerName]

		if !container.Lock.IsLocked() {
			return true, nil
		}

		container.Lock.SetUnlocked()
		room.Containers[containerName] = container

		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> unlocks the <ansi fg="container">%s</ansi>.`, mob.Character.Name, containerName))

		return true, nil

	} else if exitName != `` {

		exitInfo, _ := room.GetExitInfo(exitName)

		if !exitInfo.Lock.IsLocked() {
			return true, nil
		}

		exitInfo.Lock.SetUnlocked()
		room.SetExitLock(exitName, false)

		room.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> unlocks the <ansi fg="exit">%s</ansi> lock`, mob.Character.Name, exitName))

		return true, nil

	}

	return true, nil

}
