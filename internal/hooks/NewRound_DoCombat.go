package hooks

import (
	"fmt"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/combat"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func DoCombat(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	//
	// Combat rounds
	//
	affectedPlayers1, affectedMobs1 := handlePlayerCombat(evt)

	affectedPlayers2, affectedMobs2 := handleMobCombat(evt)

	// Do any resolution or extra checks based on everyone that has been involved in combat this round.
	handleAffected(append(affectedPlayers1, affectedPlayers2...), append(affectedMobs1, affectedMobs2...))

	return events.Continue
}

func handlePlayerCombat(evt events.NewRound) (affectedPlayerIds []int, affectedMobIds []ids.ID) {

	c := configs.GetConfig()

	tStart := time.Now()

	for _, userId := range accounts.GetOnlineUserIds() {

		user := accounts.GetByUserId(userId)

		// If has a buff that prevents combat, skip the player
		if user.Character.HasBuffFlag("no-combat") {
			continue
		}

		if user == nil || user.Character.Aggro == nil {
			continue
		}

		// Disable any buffs that are cancelled by combat
		user.Character.CancelBuffsWithFlag("cancel-on-combat")

		roomId := user.Character.RoomId

		uRoom := rooms.LoadRoom(roomId)
		if uRoom == nil {
			continue
		}

		if user.Character.Aggro.Type == characters.Flee {

			// Revert to Default combat regardless of outcome
			user.Character.SetAggro(user.Character.Aggro.Target, characters.DefaultAttack)

			blockedByMob := ``
			for _, mobInstId := range uRoom.GetMobs(rooms.FindFighting) {
				if mob := mobs.GetInstance(mobInstId); mob != nil {
					if mob.Character.Aggro == nil || mob.Character.Aggro.Target != user.Character.Id {
						continue
					}

					// Stat comparison accounts for up to 70% of chance to flee.
					chanceIn100 := int(float64(user.Character.Stats.Speed.ValueAdj) / (float64(user.Character.Stats.Speed.ValueAdj) + float64(mob.Character.Stats.Speed.ValueAdj)) * 70)
					chanceIn100 += 30

					roll := util.Rand(100)

					util.LogRoll(`Flee`, roll, chanceIn100)

					if roll >= chanceIn100 {
						blockedByMob = mob.Character.Name
						break
					}
				}
			}

			blockedByPlayer := ``
			blockedByPlayerId := 0
			for _, userId := range uRoom.GetPlayers(rooms.FindFighting) {
				if u := accounts.GetByUserId(userId); u != nil {
					// Pre-existing bug E8c, kept for slice 5: the shadowed userId makes this compare u's target with u itself.
					if u.Character.Aggro == nil || u.Character.Aggro.Target != u.Character.Id {
						continue
					}

					// if equal, 25% chance of fleeing... at best, 50% chance. Then add 50% on top.
					chanceIn100 := int(float64(user.Character.Stats.Speed.ValueAdj) / (float64(user.Character.Stats.Speed.ValueAdj) + float64(u.Character.Stats.Speed.ValueAdj)) * 70)
					chanceIn100 += 30

					roll := util.Rand(100)

					util.LogRoll(`Flee`, roll, chanceIn100)

					if roll >= chanceIn100 {
						blockedByPlayer = u.Character.Name
						blockedByPlayerId = u.UserId
						break
					}
				}
			}

			if blockedByMob != `` {
				user.SendText(fmt.Sprintf(`<ansi fg="red-bold"><ansi fg="mobname">%s</ansi> blocks you from fleeing!</ansi>`, blockedByMob))
				uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is blocked from fleeing by <ansi fg="mobname">%s</ansi>!`, user.Character.Name, blockedByMob), user.UserId)
				continue
			}

			if blockedByPlayer != `` {
				user.SendText(fmt.Sprintf(`<ansi fg="red-bold"><ansi fg="username">%s</ansi> blocks you from fleeing!</ansi>`, blockedByPlayer))
				uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is blocked from fleeing by <ansi fg="username">%s</ansi>!`, user.Character.Name, blockedByPlayer), user.UserId, blockedByPlayerId)
				continue
			}

			// Success!
			exitName, exitRoomId := uRoom.GetRandomExit()

			if exitName == `` {
				user.SendText(`You can't find an exit!`)
				continue
			}

			user.SendText(fmt.Sprintf(`You flee to the <ansi fg="exit">%s</ansi> exit!`, exitName))
			uRoom.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> flees to the <ansi fg="exit">%s</ansi> exit!`, user.Character.Name, exitName), user.UserId)

			user.Character.Aggro = nil
			events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})

			if err := rooms.MoveToRoom(user.UserId, exitRoomId); err == nil {

				for _, instId := range uRoom.GetMobs(rooms.FindCharmed) {
					if mob := mobs.GetInstance(instId); mob != nil {
						// Charmed mobs assist
						if mob.Character.IsCharmed(user.Character.Id) {
							mob.Command(exitName)
						}
					}
				}

				newRoom := rooms.LoadRoom(exitRoomId)

				usercommands.Look(``, user, newRoom, events.CmdSecretly)

			}

			continue
		}

		/**************************
		*
		* START HANDLING PHYSICAL COMBAT
		*
		**************************/

		// In combat with another player
		if user.Character.Aggro != nil && accounts.GetByBodyId(user.Character.Aggro.Target) != nil {

			defUser := accounts.GetByBodyId(user.Character.Aggro.Target)

			uRoom := rooms.LoadRoom(roomId)

			if uRoom == nil {
				user.Character.Aggro = nil
				continue
			}

			targetFound := true
			if defUser == nil {
				targetFound = false
			} else if defUser.Character.RoomId != user.Character.RoomId {

				if user.Character.Aggro.ExitName == `` {
					targetFound = false
				} else {
					// If the exitId doesn't match the target room id, can't find em
					if _, exitRoomId := uRoom.FindExitByName(user.Character.Aggro.ExitName); exitRoomId != defUser.Character.RoomId {
						targetFound = false
					}
				}

			}

			if !targetFound {
				user.SendText(`Your target can't be found.`)
				user.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
				continue
			}

			defRoom := rooms.LoadRoom(defUser.Character.RoomId)
			if defRoom == nil {
				user.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
				continue
			}

			defUser.Character.CancelBuffsWithFlag("cancel-on-combat")

			if defUser.Character.Health < 1 {
				user.SendText(`Your rage subsides.`)
				user.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
				continue
			}

			if user.Character.Aggro.RoundsWaiting > 0 {
				mudlog.Debug(`RoundsWaiting`, `User`, user.Character.Name, `Rounds`, user.Character.Aggro.RoundsWaiting)

				user.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, user.Character, defUser.Character, combat.User, combat.User)

				for _, msg := range roundResult.MessagesToSource {
					user.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToTarget {
					defUser.SendText(msg)
				}

				if len(roundResult.MessagesToSourceRoom) > 0 {
					for _, msg := range roundResult.MessagesToSourceRoom {
						uRoom.SendText(msg, user.UserId, defUser.UserId)
					}
				}

				if len(roundResult.MessagesToTargetRoom) > 0 {
					for _, msg := range roundResult.MessagesToTargetRoom {
						defRoom.SendText(msg, user.UserId, defUser.UserId)
					}
				}

				continue
			}

			// Can't see them, can't fight them.
			if defUser.Character.HasBuffFlag("hidden") {
				user.SendText("You can't seem to find your target.")
				continue
			}

			affectedPlayerIds = append(affectedPlayerIds, defUser.UserId)

			roundResult := combat.AttackPlayerVsPlayer(user, defUser)

			room := rooms.LoadRoom(roomId)

			for _, buffId := range roundResult.BuffSource {
				user.AddBuff(buffId, `combat`)
			}

			for _, buffId := range roundResult.BuffTarget {
				defUser.AddBuff(buffId, `combat`)
			}

			for _, msg := range roundResult.MessagesToSource {
				user.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToTarget {
				defUser.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				uRoom.SendText(msg, user.UserId, defUser.UserId)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg, user.UserId, defUser.UserId)
			}

			// If the attack connected, check for damage to equipment.
			if roundResult.Hit {

				defUser.Character.TrackPlayerDamage(user.UserId, roundResult.DamageToTarget)

				// For now, only focus on offhand items.
				if defUser.Character.Equipment.Offhand.ItemId > 0 {

					modifier := 0
					if roundResult.Crit { // Crits double the chance of breakage for offhand items.
						modifier = int(defUser.Character.Equipment.Offhand.GetSpec().BreakChance)
					}

					if defUser.Character.Equipment.Offhand.BreakTest(modifier) {
						// Send message about the break

						defUser.SendText(`<ansi fg="202">***</ansi>`)
						defUser.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple()))
						defUser.SendText(`<ansi fg="202">***</ansi>`)

						defRoom.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple(), defUser.Character.Name), defUser.UserId)

						events.AddToQueue(events.ItemOwnership{
							UserId: defUser.UserId,
							Item:   defUser.Character.Equipment.Offhand,
							Gained: false,
						})

						defUser.Character.RemoveFromBody(defUser.Character.Equipment.Offhand)

						itm := items.New(20) // Broken item
						if !defUser.Character.StoreItem(itm) {
							room.AddItem(itm, false)

							events.AddToQueue(events.ItemOwnership{
								UserId: defUser.UserId,
								Item:   itm,
								Gained: true,
							})
						}
					}
				}
			}

			if roundResult.DamageToSource != 0 {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
			}
			if roundResult.DamageToTarget != 0 {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: defUser.UserId})
			}

			if user.Character.Health <= 0 || defUser.Character.Health <= 0 {
				defUser.Character.EndAggro()
				user.Character.EndAggro()
			} else {
				user.Character.SetAggro(defUser.Character.Id, characters.DefaultAttack)
			}

		}

		// In combat with a mob
		// (any target that is not a player, so a vanished target is cleared here)
		if user.Character.Aggro != nil && !user.Character.Aggro.Target.IsZero() && accounts.GetByBodyId(user.Character.Aggro.Target) == nil {

			affectedMobIds = append(affectedMobIds, user.Character.Aggro.Target)

			defMob := mobs.GetInstance(user.Character.Aggro.Target)

			targetFound := true
			if defMob == nil {
				targetFound = false
			} else if defMob.Character.RoomId != user.Character.RoomId {

				if user.Character.Aggro.ExitName == `` {
					targetFound = false
				} else {
					// Make sure the target is still at the exit

					uRoom := rooms.LoadRoom(roomId)
					if uRoom == nil {
						user.Character.Aggro = nil
						continue
					}

					// If the exitId doesn't match the target room id, can't find em
					if _, exitRoomId := uRoom.FindExitByName(user.Character.Aggro.ExitName); exitRoomId != defMob.Character.RoomId {
						targetFound = false
					}

				}

			}

			if !targetFound {
				user.SendText("Your target can't be found.")
				user.Character.Aggro = nil
				continue
			}

			defRoom := rooms.LoadRoom(defMob.Character.RoomId)

			defMob.Character.CancelBuffsWithFlag("cancel-on-combat")

			if defMob.Character.Health < 1 {
				user.SendText("Your rage subsides.")
				user.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
				continue
			}

			if user.Character.Aggro.RoundsWaiting > 0 {
				mudlog.Debug(`RoundsWaiting`, `User`, user.Character.Name, `Rounds`, user.Character.Aggro.RoundsWaiting)

				user.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, user.Character, &defMob.Character, combat.User, combat.Mob)

				for _, msg := range roundResult.MessagesToSource {
					user.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToSourceRoom {
					uRoom.SendText(msg, user.UserId)
				}

				for _, msg := range roundResult.MessagesToTargetRoom {
					defRoom.SendText(msg, user.UserId)
				}

				continue
			}

			// Can't see them, can't fight them.
			if defMob.Character.HasBuffFlag("hidden") {
				user.SendText("You can't seem to find your target.")
				continue
			}

			// Pre-existing bug E8e, kept for slice 5: this appended the aggro target's user ID, always 0 in a mob fight.
			affectedPlayerIds = append(affectedPlayerIds, 0)

			var roundResult combat.AttackResult

			roundResult = combat.AttackPlayerVsMob(user, defMob)

			for _, buffId := range roundResult.BuffSource {
				user.AddBuff(buffId, `combat`)
			}

			for _, buffId := range roundResult.BuffTarget {
				defMob.AddBuff(buffId, `combat`)
			}

			for _, msg := range roundResult.MessagesToSource {
				user.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				uRoom.SendText(msg, user.UserId)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg, user.UserId)
			}

			//
			// Special mob-only reaction/behavior
			//
			// Hostility default to 5 minutes
			for _, groupName := range defMob.Groups {
				mobs.MakeHostile(groupName, user.UserId, c.Timing.MinutesToRounds(2)-user.Character.Stats.Perception.ValueAdj)
			}

			// Mobs get aggro when attacked
			if defMob.Character.Aggro == nil {
				defMob.PreventIdle = true
				// If not in the same room,
				// find an exit to the room of the player to move to
				if user.Character.RoomId != defMob.Character.RoomId {
					if mobRoom := rooms.LoadRoom(defMob.Character.RoomId); mobRoom != nil {
						for exitName, exitInfo := range mobRoom.Exits {
							if exitInfo.RoomId == user.Character.RoomId {
								defMob.Command(fmt.Sprintf(`go %s`, exitName))
								if actionStr := defMob.GetAngryCommand(); actionStr != `` {
									defMob.Command(actionStr)
								}
								break
							}
						}
					}
				}

			}

			if roundResult.DamageToSource != 0 {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
			}

			if user.Character.Health <= 0 || defMob.Character.Health <= 0 {
				defMob.Character.EndAggro()
				events.AddToQueue(events.AggroChanged{MobInstanceId: defMob.Character.Id, RoomId: defMob.Character.RoomId})
				user.Character.EndAggro()
				events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
			} else {
				user.Character.SetAggro(defMob.Character.Id, characters.DefaultAttack)
				events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
			}

		}

		/**************************
		*
		* END HANDLING PHYSICAL COMBAT
		*
		**************************/

	}

	util.TrackTime(`DoCombat::handlePlayerCombat()`, time.Since(tStart).Seconds())

	return affectedPlayerIds, affectedMobIds
}

func handleMobCombat(evt events.NewRound) (affectedPlayerIds []int, affectedMobIds []ids.ID) {

	tStart := time.Now()

	// Handle mob round of combat
	for _, mobId := range mobs.GetAllMobInstanceIds() {

		mob := mobs.GetInstance(mobId)

		// Only handling combat functions here, so ditch out if not in combat
		if mob == nil || mob.Character.Aggro == nil {
			continue
		}

		// If has a buff that prevents combat, skip the player
		if mob.Character.HasBuffFlag("no-combat") {
			continue
		}

		mobRoom := rooms.LoadRoom(mob.Character.RoomId)

		if mobRoom == nil {
			mob.Character.Aggro = nil
			events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
			continue
		}

		// Disable any buffs that are cancelled by combat
		mob.Character.CancelBuffsWithFlag("cancel-on-combat")

		/**************************
		*
		* START HANDLING PHYSICAL COMBAT
		*
		**************************/
		c := configs.GetConfig()

		// H2H is the base level combat, can do combat commands then
		if mob.Character.Aggro.Type == characters.DefaultAttack {

			// If they have idle commands, maybe do one of them?
			cmdCt := len(mob.CombatCommands)
			if cmdCt > 0 {

				// Each mob has a 10% chance of doing an idle action.
				if util.Rand(100) < mob.ActivityLevel {

					combatAction := mob.CombatCommands[util.Rand(cmdCt)]

					if combatAction == `` { // blank is a no-op
						continue
					}

					var waitTime float64 = 0.0
					allCmds := strings.Split(combatAction, `;`)
					if len(allCmds) >= c.Timing.TurnsPerRound() {
						mob.Command(`say I have a CombatAction that is too long. Please notify an admin.`)
					} else {
						for _, action := range strings.Split(combatAction, `;`) {
							mob.Command(action, waitTime)
							waitTime += 0.1
						}
					}
					continue
				}
			}

		}
		roomId := mob.Character.RoomId

		affectedMobIds = append(affectedMobIds, mob.Character.Id)

		// mob attacks player
		if mob.Character.Aggro != nil && accounts.GetByBodyId(mob.Character.Aggro.Target) != nil {

			defUser := accounts.GetByBodyId(mob.Character.Aggro.Target)
			if defUser == nil || mob.Character.RoomId != defUser.Character.RoomId {
				mob.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				continue
			}

			defRoom := rooms.LoadRoom(defUser.Character.RoomId)
			if defRoom == nil {
				mob.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				continue
			}

			defUser.Character.CancelBuffsWithFlag("cancel-on-combat")

			if defUser.Character.Health < 1 {
				mob.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				continue
			}
			if defUser.Character.HasBuffFlag("hidden") {
				continue
			}

			affectedPlayerIds = append(affectedPlayerIds, defUser.UserId)

			// If no weapon but has stuff in the backpack, look for a weapon
			// Especially useful for when they get disarmed
			if mob.Character.Equipment.Weapon.ItemId == 0 && len(mob.Character.Items) > 0 {

				roll := util.Rand(100)

				util.LogRoll(`Look for weapon`, roll, mob.Character.Stats.Perception.ValueAdj)

				if roll < mob.Character.Stats.Perception.ValueAdj {
					possibleWeapons := []string{}
					for _, itm := range mob.Character.Items {
						iSpec := itm.GetSpec()
						if iSpec.Type == items.Weapon {
							possibleWeapons = append(possibleWeapons, itm.DisplayName())
						}
					}

					if len(possibleWeapons) > 0 {
						mob.Command(fmt.Sprintf("equip %s", possibleWeapons[util.Rand(len(possibleWeapons))]))
					}

				}
			}

			if mob.Character.Aggro.RoundsWaiting > 0 {
				mudlog.Debug(`RoundsWaiting`, `User`, mob.Character.Name, `Rounds`, mob.Character.Aggro.RoundsWaiting)

				mob.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, &mob.Character, defUser.Character, combat.Mob, combat.User)

				for _, msg := range roundResult.MessagesToTarget {
					defUser.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToSourceRoom {
					mobRoom.SendText(msg, defUser.UserId)
				}

				for _, msg := range roundResult.MessagesToTargetRoom {
					defRoom.SendText(msg, defUser.UserId)
				}

				continue
			}

			var roundResult combat.AttackResult

			roundResult = combat.AttackMobVsPlayer(mob, defUser)

			room := rooms.LoadRoom(roomId)

			for _, buffId := range roundResult.BuffSource {
				mob.AddBuff(buffId, `combat`)
			}

			for _, buffId := range roundResult.BuffTarget {
				defUser.AddBuff(buffId, `combat`)
			}

			for _, msg := range roundResult.MessagesToTarget {
				defUser.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				mobRoom.SendText(msg, defUser.UserId)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg, defUser.UserId)
			}

			// If the attack connected, check for damage to equipment.
			if roundResult.Hit {

				// For now, only focus on offhand items.
				if defUser.Character.Equipment.Offhand.ItemId > 0 {

					modifier := 0
					if roundResult.Crit { // Crits double the chance of breakage for offhand items.
						modifier = int(defUser.Character.Equipment.Offhand.GetSpec().BreakChance)
					}

					if defUser.Character.Equipment.Offhand.BreakTest(modifier) {
						// Send message about the break

						defUser.SendText(`<ansi fg="202">***</ansi>`)
						defUser.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple()))
						defUser.SendText(`<ansi fg="202">***</ansi>`)

						defRoom.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="username">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defUser.Character.Equipment.Offhand.NameSimple(), defUser.Character.Name), defUser.UserId)

						events.AddToQueue(events.ItemOwnership{
							UserId: defUser.UserId,
							Item:   defUser.Character.Equipment.Offhand,
							Gained: false,
						})

						defUser.Character.RemoveFromBody(defUser.Character.Equipment.Offhand)

						itm := items.New(20) // Broken item
						if !defUser.Character.StoreItem(itm) {
							room.AddItem(itm, false)

							events.AddToQueue(events.ItemOwnership{
								UserId: defUser.UserId,
								Item:   itm,
								Gained: true,
							})

						}
					}
				}
			}

			if roundResult.DamageToTarget != 0 {
				events.AddToQueue(events.CharacterVitalsChanged{UserId: defUser.UserId})
			}

			if mob.Character.Health <= 0 || defUser.Character.Health <= 0 {
				mob.Character.EndAggro()
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				defUser.Character.EndAggro()
				events.AddToQueue(events.AggroChanged{UserId: defUser.UserId, RoomId: defUser.Character.RoomId})
			} else {
				mob.Character.SetAggro(defUser.Character.Id, characters.DefaultAttack)
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
			}
		}

		// mob attacks mob
		// (any target that is not a player, so a vanished target is cleared here)
		if mob.Character.Aggro != nil && !mob.Character.Aggro.Target.IsZero() && accounts.GetByBodyId(mob.Character.Aggro.Target) == nil {

			affectedMobIds = append(affectedMobIds, mob.Character.Aggro.Target)

			defMob := mobs.GetInstance(mob.Character.Aggro.Target)

			if defMob == nil || mob.Character.RoomId != defMob.Character.RoomId {
				mob.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				continue
			}

			defRoom := rooms.LoadRoom(defMob.Character.RoomId)

			defMob.Character.CancelBuffsWithFlag("cancel-on-combat")

			if defMob.Character.Health < 1 {
				mob.Character.Aggro = nil
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				continue
			}

			if mob.Character.Aggro.RoundsWaiting > 0 {
				mudlog.Debug(`RoundsWaiting`, `User`, mob.Character.Name, `Rounds`, mob.Character.Aggro.RoundsWaiting)

				mob.Character.Aggro.RoundsWaiting--

				roundResult := combat.GetWaitMessages(items.Wait, &mob.Character, &defMob.Character, combat.Mob, combat.Mob)

				for _, msg := range roundResult.MessagesToSourceRoom {
					mobRoom.SendText(msg)
				}

				for _, msg := range roundResult.MessagesToTargetRoom {
					defRoom.SendText(msg)
				}

				continue
			}

			// Can't see them, can't fight them.
			if defMob.Character.HasBuffFlag("hidden") {
				continue
			}

			var roundResult combat.AttackResult

			roundResult = combat.AttackMobVsMob(mob, defMob)

			for _, buffId := range roundResult.BuffSource {
				mob.AddBuff(buffId, `combat`)
			}

			for _, buffId := range roundResult.BuffTarget {
				defMob.AddBuff(buffId, `combat`)
			}

			for _, msg := range roundResult.MessagesToSourceRoom {
				mobRoom.SendText(msg)
			}

			for _, msg := range roundResult.MessagesToTargetRoom {
				defRoom.SendText(msg)
			}

			// Mobs get aggro when attacked
			if defMob.Character.Aggro == nil {
				defMob.PreventIdle = true
			}

			// If the attack connected, check for damage to equipment.
			if roundResult.Hit {
				// For now, only focus on offhand items.
				if defMob.Character.Equipment.Offhand.ItemId > 0 {

					modifier := 0
					if roundResult.Crit { // Crits double the chance of breakage for offhand items.
						modifier = int(defMob.Character.Equipment.Offhand.GetSpec().BreakChance)
					}

					if defMob.Character.Equipment.Offhand.BreakTest(modifier) {
						// Send message about the break

						if defRoom := rooms.LoadRoom(defMob.Character.RoomId); defRoom != nil {

							defRoom.SendText(fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> <ansi fg="mobname">%s</ansi> was carrying breaks! <ansi fg="202">***</ansi></ansi>`, defMob.Character.Equipment.Offhand.NameSimple(), defMob.Character.Name))

							events.AddToQueue(events.ItemOwnership{
								MobInstanceId: defMob.Character.Id,
								Item:          defMob.Character.Equipment.Offhand,
								Gained:        false,
							})

							defMob.Character.RemoveFromBody(defMob.Character.Equipment.Offhand)
							itm := items.New(20) // Broken item
							if !defMob.Character.StoreItem(itm) {
								defRoom.AddItem(itm, false)

								events.AddToQueue(events.ItemOwnership{
									MobInstanceId: defMob.Character.Id,
									Item:          itm,
									Gained:        true,
								})
							}
						}
					}
				}
			}

			if mob.Character.Health <= 0 || defMob.Character.Health <= 0 {
				mob.Character.EndAggro()
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
				defMob.Character.EndAggro()
				events.AddToQueue(events.AggroChanged{MobInstanceId: defMob.Character.Id, RoomId: defMob.Character.RoomId})
			} else {
				mob.Character.SetAggro(defMob.Character.Id, characters.DefaultAttack)
				events.AddToQueue(events.AggroChanged{MobInstanceId: mob.Character.Id, RoomId: mob.Character.RoomId})
			}

		}

		/**************************
		*
		* END HANDLING PHYSICAL COMBAT
		*
		**************************/

	}

	util.TrackTime(`World::handleMobCombat()`, time.Since(tStart).Seconds())

	return affectedPlayerIds, affectedMobIds
}

func handleAffected(affectedPlayerIds []int, affectedMobIds []ids.ID) {

	playersHandled := map[int]struct{}{}
	for _, userId := range affectedPlayerIds {
		if _, ok := playersHandled[userId]; ok {
			continue
		}
		playersHandled[userId] = struct{}{}

		if user := accounts.GetByUserId(userId); user != nil {

			if user.Character.Health <= -10 {

				if user.Character.Aggro != nil && !user.Character.Aggro.Target.IsZero() && accounts.GetByBodyId(user.Character.Aggro.Target) == nil {
					user.Character.KilledBy = user.Character.Aggro.Target
					if killerMob := mobs.GetInstance(user.Character.Aggro.Target); killerMob != nil {
						user.Character.KillerMobIsElite = killerMob.IsElite
						user.Character.KillerMobName = killerMob.Character.Name
					}
				}
				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

			} else if user.Character.Health < 1 {

				events.AddToQueue(events.PlayerDrop{UserId: user.UserId, RoomId: user.Character.RoomId})

			}

		}
	}

	mobsHandled := map[ids.ID]struct{}{}
	for _, mobId := range affectedMobIds {
		if _, ok := mobsHandled[mobId]; ok {
			continue
		}
		mobsHandled[mobId] = struct{}{}

		if mob := mobs.GetInstance(mobId); mob != nil {
			if mob.Character.Health < 1 {

				mob.Command(`suicide`)

			}
		}

	}

}
