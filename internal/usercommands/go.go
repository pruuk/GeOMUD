package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Go(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// nil for an NPC; used only for the account parts (GMCP, party, hostility groups).
	user := actor.Controller().Account()

	if actor.Character.Aggro != nil {
		tell(actor, "You can't do that! You are in combat!")
		return true, nil
	}

	// If has a buff that prevents combat, skip the player
	if actor.Character.HasBuffFlag("no-go") {
		tell(actor, "You can't do that!")
		return true, nil
	}

	c := configs.GetTextFormatsConfig()

	isSneaking := actor.Character.HasBuffFlag("hidden")

	handled := false

	exitName, goRoomId := room.FindExitByName(rest)

	if exitName != `` {

		if actor.Character.IsDisabled() {
			tell(actor, "You are unable to do that while downed.")
			return true, nil
		}

		actionCost := 10
		encumbered := false
		if len(actor.Character.Items) > actor.Character.CarryCapacity() {
			actionCost = 50
			encumbered = true
		}

		if !actor.Character.DeductActionPoints(actionCost) {

			if encumbered {
				tell(actor, "You're too encumbered to move (<ansi fg=\"command\">help encumbrance</ansi>)!")
			} else {
				tell(actor, "You're too tired to move (slow down)!")
				mudlog.Debug("No ActionPoints", "AP", actor.Character.ActionPoints, "Needed", actionCost)
			}

			return true, nil
		}

		originRoomId := actor.Character.RoomId

		exitInfo, _ := room.GetExitInfo(exitName)

		if exitInfo.Lock.IsLocked() {

			lockId := fmt.Sprintf(`%d-%s`, room.RoomId, exitName)

			hasKey, hasSequence := actor.Character.HasKey(lockId, int(room.Exits[exitName].Lock.Difficulty))

			lockpickItm := items.Item{}
			// Only look for a lockpick kit if they know the sequence
			if hasSequence {
				for _, itm := range actor.Character.GetAllBackpackItems() {
					if itm.GetSpec().Type == items.Lockpicks {
						lockpickItm = itm
						break
					}
				}
			}

			if lockpickItm.ItemId > 0 && hasSequence {

				tell(actor, `You know this lock well, you quickly pick it.`)
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> quickly picks the lock on the <ansi fg="exit">%s</ansi> exit.`, actor.Character.Name, exitName),
					excludeId(actor))

				exitInfo.Lock.SetUnlocked()
				room.SetExitLock(exitName, false)

			} else if hasKey {
				tell(actor, fmt.Sprintf(`You use the key on your key ring to unlock the <ansi fg="exit">%s</ansi> exit.`, exitName))
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> exit.`, actor.Character.Name, exitName),
					excludeId(actor))

				exitInfo.Lock.SetUnlocked()
				room.SetExitLock(exitName, false)

			} else {

				// check for a key item on their person
				if backpackKeyItm, hasBackpackKey := actor.Character.FindKeyInBackpack(lockId); hasBackpackKey {

					itmSpec := backpackKeyItm.GetSpec()

					tell(actor, fmt.Sprintf(`You use your <ansi fg="item">%s</ansi> to unlock the <ansi fg="exit">%s</ansi> exit, and add it to your key ring for the future.`, itmSpec.Name, exitName))
					room.SendText(
						fmt.Sprintf(`<ansi fg="username">%s</ansi> uses a key to unlock the <ansi fg="exit">%s</ansi> exit.`, actor.Character.Name, exitName),
						excludeId(actor))

					// Key entries look like:
					// "key-<roomid>-<exitname>": "<itemid>"
					actor.Character.SetKey(`key-`+lockId, fmt.Sprintf(`%d`, backpackKeyItm.ItemId))
					actor.Character.RemoveItem(backpackKeyItm)

					events.AddToQueue(events.ItemOwnership{
						ActorId: actor.Character.Id,
						Item:    backpackKeyItm,
						Gained:  false,
					})

					exitInfo.Lock.SetUnlocked()
					room.SetExitLock(exitName, false)
				}

				if exitInfo.Lock.IsLocked() {
					tell(actor, `There's a lock preventing you from going that way. You'll need a <ansi fg="item">Key</ansi> or to <ansi fg="command">pick</ansi> the lock with <ansi fg="item">lockpicks</ansi>.`)
					// Send GMCP message
					if user != nil {
						if f, ok := GetExportedFunction(`SendGMCPEvent`); ok {
							if gmcpSendFunc, ok := f.(func(int, string, any)); ok { // make sure the func definition is `func(int, string, any)`
								gmcpSendFunc(user.UserId, `Room.WrongDir`, fmt.Sprintf(`"%s"`, exitName))
							}
						}
					}

					return true, nil
				}
			}

		}

		if exitInfo.ExitMessage != `` && !flags.Has(events.CmdIsRequeue) {
			tell(actor, exitInfo.ExitMessage)
			queueFlagged(actor, rest, flags|events.CmdIsRequeue|events.CmdBlockInputUntilComplete, 1)
			return true, nil
		}

		// Load current room details
		destRoom := rooms.LoadRoom(goRoomId)
		if destRoom == nil {
			return false, fmt.Errorf(`room %d not found`, goRoomId)
		}

		// Grab the exit in the target room that leads to this room (if any)
		enterFromExit := destRoom.FindExitTo(room.RoomId)

		if len(enterFromExit) < 1 {
			enterFromExit = "somewhere"
		} else {

			// Entering through the other side unlocks this side
			exitInfo := destRoom.Exits[enterFromExit]
			if exitInfo.Lock.IsLocked() {
				exitInfo.Lock.SetUnlocked()
				destRoom.SetExitLock(enterFromExit, false)
			}

			enterFromExit = fmt.Sprintf(`the <ansi fg="exit">%s</ansi>`, enterFromExit)
		}

		if err := rooms.MoveToRoom(actor.Character.Id, destRoom.RoomId); err != nil {
			tell(actor, "Oops, couldn't move there!")
		} else {

			// Tell the player they are moving
			if isSneaking {
				tell(actor,
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`You <ansi fg="black-bold">sneak</ansi> towards the <ansi fg="exit">%s</ansi> exit.`, exitName),
					))
			} else {
				tell(actor,
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`You head towards the <ansi fg="exit">%s</ansi> exit.`, exitName),
					))

				// Tell the old room they are leaving
				room.SendText(
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`<ansi fg="username">%s</ansi> leaves towards the <ansi fg="exit">%s</ansi> exit.`, actor.Character.Name, exitName),
					),
					excludeId(actor))

				// Tell the new room they have arrived
				destRoom.SendText(
					fmt.Sprintf(string(c.EnterRoomMessageWrapper),
						fmt.Sprintf(`<ansi fg="username">%s</ansi> enters from <ansi fg="exit">%s</ansi>.`, actor.Character.Name, enterFromExit),
					),
					excludeId(actor))

				destRoom.SendTextToExits(`You hear someone moving around.`, true, room.GetPlayers(rooms.FindAll)...)
			}

			var currentParty *parties.Party
			if user != nil {
				currentParty = parties.Get(user.UserId)
			}
			if currentParty != nil {

				if currentParty.IsLeader(user.UserId) {

					for _, partyMemberId := range currentParty.UserIds {
						if partyMemberId == user.UserId {
							continue
						}
						if partyUser := accounts.GetByUserId(partyMemberId); partyUser != nil {
							if partyUser.Character.RoomId == room.RoomId {
								partyUser.SendText(`You follow the party leader.`)
								partyUser.Command(rest)
							}
						}
					}

				}
			}

			for _, instId := range room.GetMobs(rooms.FindCharmed) {
				mob := mobs.GetInstance(instId)
				if mob == nil {
					continue
				}
				// They only follow if they're in the same room as the player
				if mob.Character.RoomId != originRoomId {
					continue
				}
				if mob.Character.IsCharmed(actor.Character.Id) { // Charmed mobs follow
					mob.Command(rest)
				}
			}

			if !isSneaking {
				//
				// When leaving a room, mobs who were attacking may follow
				//
				mobIds := room.GetMobs(rooms.FindFightingPlayer)
				for _, mobId := range mobIds {
					mob := mobs.GetInstance(mobId)
					if mob == nil {
						continue
					}

					if mob.Character.Aggro == nil || mob.Character.Aggro.Target != actor.Character.Id {
						continue
					}

					speedDelta := mob.Character.Stats.Speed.ValueAdj - actor.Character.Stats.Speed.ValueAdj
					if speedDelta < 1 {
						speedDelta = 1
					}

					// Chance that a mob follows the player
					targetVal := 20 + mob.Character.Stats.Perception.ValueAdj + speedDelta

					roll := util.Rand(100)

					util.LogRoll(`Mob Follow`, roll, targetVal)

					if roll >= targetVal {
						continue
					}

					mob.Command(rest)

				}

				//
				// When entering a room, mobs might be waiting to attack
				//
				mobIds = destRoom.GetMobs(rooms.FindAll)
				for _, mobId := range mobIds {
					mob := mobs.GetInstance(mobId)
					if mob == nil {
						continue
					}
					if mob.Character.Aggro != nil {
						continue
					}
					if mob.Character.IsCharmed() {
						continue
					}

					isHostile := mob.Hostile // Is it automatically hostile?
					if !isHostile {
						for _, groupName := range mob.Groups {
							if user != nil && mobs.IsHostile(groupName, user.UserId) {
								isHostile = true
								break
							}
						}
						if !isHostile { // is it still not hostile?
							continue
						}
					}

					tell(actor, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> notices you as you enter!`, mob.Character.Name))

				}

			}

			handled = true

			Look(``, actor, destRoom, events.CmdSecretly)

		}

	}

	if !handled {

		if rest == "north" || rest == "south" || rest == "east" || rest == "west" || rest == "up" || rest == "down" || rest == "northwest" || rest == "northeast" || rest == "southwest" || rest == "southeast" {
			tell(actor, "You're bumping into walls.")

			// Send GMCP message
			if user != nil {
				if f, ok := GetExportedFunction(`SendGMCPEvent`); ok {
					if gmcpSendFunc, ok := f.(func(int, string, any)); ok { // make sure the func definition is `func(int, string, any)`
						gmcpSendFunc(user.UserId, `Room.WrongDir`, fmt.Sprintf(`"%s"`, rest))
					}
				}
			}

			if !actor.Character.HasBuffFlag("hidden") {

				room.SendText(
					fmt.Sprintf(string(c.ExitRoomMessageWrapper),
						fmt.Sprintf(`<ansi fg="username">%s</ansi> is bumping into walls.`, actor.Character.Name),
					),
					excludeId(actor))
			}
			handled = true
		}

	}

	return handled, nil
}
