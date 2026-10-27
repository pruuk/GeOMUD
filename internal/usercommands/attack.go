package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Attack(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	attackPlayerId := 0
	attackMobId := ids.ID{}

	if rest == `` {
		partyInfo := parties.Get(user.UserId)

		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
			m := mobs.GetInstance(mId)
			if m.Character.Aggro == nil {
				continue
			}

			if m.Character.Aggro.Target == user.Character.Id {
				attackMobId = m.Character.Id
				break
			}

			if partyInfo != nil {
				if tu := accounts.GetByBodyId(m.Character.Aggro.Target); tu != nil && partyInfo.IsMember(tu.UserId) {
					attackMobId = m.Character.Id
					break
				}
			}
		}

		if attackMobId.IsZero() {
			for _, uId := range room.GetPlayers(rooms.FindFightingPlayer) {
				u := accounts.GetByUserId(uId)
				if u == nil || u.Character.Aggro == nil {
					continue
				}

				if u.Character.Aggro.Target == user.Character.Id {
					attackPlayerId = u.UserId
					break
				}

				if partyInfo != nil {
					if tu := accounts.GetByBodyId(u.Character.Aggro.Target); tu != nil && partyInfo.IsMember(tu.UserId) {
						attackPlayerId = u.UserId
						break
					}
				}
			}
		}

		// Finally, if still no targets, check if any party members are aggroed and just glom onto that
		if attackMobId.IsZero() && attackPlayerId == 0 {
			if partyInfo != nil {
				for uId := range partyInfo.GetMembers() {
					if partyUser := accounts.GetByUserId(uId); partyUser != nil {
						if partyUser.Character.Aggro == nil {
							continue
						}

						target := partyUser.Character.Aggro.Target

						if mobs.GetInstance(target) != nil {
							attackMobId = target
							break
						}

						if tu := accounts.GetByBodyId(target); tu != nil {
							attackPlayerId = tu.UserId
							break
						}

					}
				}
			}
		}

	} else if rest[0] == '*' { // choose a target at random. Friend or foe.

		if rest == `*` { // * ANYONE

			allMobs := room.GetMobs()
			allPlayers := []int{}
			for _, userId := range room.GetPlayers() {
				if userId == user.UserId {
					continue
				}
				allPlayers = append(allPlayers, userId)
			}

			totalTargets := len(allMobs) + len(allPlayers)
			if totalTargets == 0 {
				return true, nil
			}

			randomSelection := util.Rand(totalTargets)

			if randomSelection < len(allMobs) {
				attackMobId = allMobs[randomSelection]
			} else {
				randomSelection -= len(allMobs)
				attackPlayerId = allPlayers[randomSelection]
			}

		} else if rest == `*mob` { // *mob ANY MOB

			if allMobs := room.GetMobs(); len(allMobs) > 0 {
				attackMobId = allMobs[util.Rand(len(allMobs))]
			}

		} else { // *user etc. ANY PLAYER

			allPlayers := []int{}
			for _, userId := range room.GetPlayers() {
				if userId == user.UserId {
					continue
				}
				allPlayers = append(allPlayers, userId)
			}

			if len(allPlayers) > 0 {
				attackPlayerId = allPlayers[util.Rand(len(allPlayers))]
			}

		}

	} else {
		attackPlayerId, attackMobId = room.FindByName(rest)
	}

	if attackPlayerId == user.UserId { // Can't attack self!
		attackPlayerId = 0
	}

	if attackMobId.IsZero() && attackPlayerId == 0 {
		user.SendText("You attack the darkness!")
		return true, nil
	}

	isSneaking := user.Character.HasBuffFlag("hidden")

	/*
		combatAddlWaitRounds := user.Character.Equipment.Weapon.GetSpec().WaitRounds + user.Character.Equipment.Weapon.GetSpec().WaitRounds
		attkType := characters.DefaultAttack
		if user.Character.Equipment.Weapon.GetSpec().Subtype == items.Shooting {
			attkType = characters.Shooting
		}
	*/

	if !attackMobId.IsZero() {

		m := mobs.GetInstance(attackMobId)

		if m != nil {
			if m.Character.IsCharmed(user.Character.Id) {
				user.SendText(fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
				return true, nil
			}

			if party := parties.Get(user.UserId); party != nil {
				if party.IsLeader(user.UserId) {
					for _, id := range party.GetAutoAttackUserIds() {
						if id == user.UserId {
							continue
						}
						if partyUser := accounts.GetByUserId(id); partyUser != nil {
							if partyUser.Character.RoomId == user.Character.RoomId {

								partyUser.Command(fmt.Sprintf(`attack %s`, attackMobId)) // a body ID token names this exact mob

							}
						}

					}
				}
			}

			user.Character.SetAggro(attackMobId, characters.DefaultAttack)

			events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})

			user.SendText(fmt.Sprintf(`You prepare to fight <ansi fg="mobname">%s</ansi>!`, m.Character.Name))

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>.`, user.Character.Name, m.Character.Name),
					user.UserId,
				)
			}

		}

	} else if attackPlayerId > 0 {

		if p := accounts.GetByUserId(attackPlayerId); p != nil {

			if pvpErr := room.CanPvp(user, p); pvpErr != nil {
				user.SendText(pvpErr.Error())
				return true, nil
			}

			if partyInfo := parties.Get(user.UserId); partyInfo != nil {
				if partyInfo.IsMember(attackPlayerId) {
					user.SendText(fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
					return true, nil
				}
			}

			if party := parties.Get(user.UserId); party != nil {
				if party.IsLeader(user.UserId) {
					for _, id := range party.GetAutoAttackUserIds() {
						if id == user.UserId {
							continue
						}
						if partyUser := accounts.GetByUserId(id); partyUser != nil {
							if partyUser.Character.RoomId == user.Character.RoomId {
								partyUser.Command(fmt.Sprintf(`attack @%d`, attackPlayerId)) // # denotes a specific mob instanceId
							}
						}
					}
				}
			}

			user.Character.SetAggro(p.Character.Id, characters.DefaultAttack)

			events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})

			user.SendText(fmt.Sprintf(`You prepare to fight <ansi fg="mobname">%s</ansi>!`, p.Character.Name))

			if !isSneaking {

				p.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight you!`, user.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="username">%s</ansi> prepares to fight <ansi fg="username">%s</ansi>.`, user.Character.Name, p.Character.Name),
					user.UserId, attackPlayerId)
			}

		}

	}

	return true, nil
}
