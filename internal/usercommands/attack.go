package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// othersOf is the body IDs without the actor's own: an NPC is among the mobs
// of its room and must never pick itself as a target.
func othersOf(actor *actors.Actor, bodyIds []ids.ID) []ids.ID {
	out := make([]ids.ID, 0, len(bodyIds))
	for _, id := range bodyIds {
		if id != actor.Id {
			out = append(out, id)
		}
	}
	return out
}

func Attack(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// Parties and party auto-attack belong to accounts; user is nil for an NPC.
	user := actor.Controller().Account()
	selfUserId := ExcludeId(actor)

	attackPlayerId := 0
	attackMobId := ids.ID{}

	if rest == `` {
		var partyInfo *parties.Party
		if user != nil {
			partyInfo = parties.Get(user.UserId)
		}

		// If no argument supplied, attack whoever is attacking the player currently.
		for _, mId := range room.GetMobs(rooms.FindFightingPlayer) {
			m := mobs.GetInstance(mId)
			if m.Character.Aggro == nil {
				continue
			}

			if m.Character.Aggro.Target == actor.Character.Id {
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

				if u.Character.Aggro.Target == actor.Character.Id {
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

			allMobs := othersOf(actor, room.GetMobs())
			allPlayers := []int{}
			for _, userId := range room.GetPlayers() {
				if userId == selfUserId {
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

			if allMobs := othersOf(actor, room.GetMobs()); len(allMobs) > 0 {
				attackMobId = allMobs[util.Rand(len(allMobs))]
			}

		} else { // *user etc. ANY PLAYER

			allPlayers := []int{}
			for _, userId := range room.GetPlayers() {
				if userId == selfUserId {
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

	if attackPlayerId == selfUserId { // Can't attack self!
		attackPlayerId = 0
	}
	if attackMobId == actor.Id { // An NPC can't attack itself either.
		attackMobId = ids.ID{}
	}

	if attackMobId.IsZero() && attackPlayerId == 0 {
		Tell(actor, "You attack the darkness!")
		return true, nil
	}

	isSneaking := actor.Character.HasBuffFlag("hidden")

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
			if m.Character.IsCharmed(actor.Character.Id) {
				Tell(actor, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> is your friend!`, m.Character.Name))
				return true, nil
			}

			if user != nil {
				if party := parties.Get(user.UserId); party != nil {
					if party.IsLeader(user.UserId) {
						for _, id := range party.GetAutoAttackUserIds() {
							if id == user.UserId {
								continue
							}
							if partyUser := accounts.GetByUserId(id); partyUser != nil {
								if partyUser.Character.RoomId == actor.Character.RoomId {

									partyUser.Command(fmt.Sprintf(`attack %s`, attackMobId)) // a body ID token names this exact mob

								}
							}

						}
					}
				}
			}

			actor.Character.SetAggro(attackMobId, characters.DefaultAttack)

			events.AddToQueue(events.AggroChanged{ActorId: actor.Character.Id, RoomId: actor.Character.RoomId})

			Tell(actor, fmt.Sprintf(`You prepare to fight <ansi fg="mobname">%s</ansi>!`, m.Character.Name))

			if !isSneaking {
				room.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> prepares to fight <ansi fg="mobname">%s</ansi>.`, NameColour(actor), actor.Character.Name, m.Character.Name),
					selfUserId,
				)
			}

		}

	} else if attackPlayerId > 0 {

		if p := accounts.GetByUserId(attackPlayerId); p != nil {

			if pvpErr := room.CanPvp(actor.Character, p.Character); pvpErr != nil {
				Tell(actor, pvpErr.Error())
				return true, nil
			}

			if user != nil {
				if partyInfo := parties.Get(user.UserId); partyInfo != nil {
					if partyInfo.IsMember(attackPlayerId) {
						Tell(actor, fmt.Sprintf(`<ansi fg="username">%s</ansi> is in your party!`, p.Character.Name))
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
								if partyUser.Character.RoomId == actor.Character.RoomId {
									partyUser.Command(fmt.Sprintf(`attack %s`, p.Character.Id)) // a body ID token names this exact player
								}
							}
						}
					}
				}
			}

			actor.Character.SetAggro(p.Character.Id, characters.DefaultAttack)

			events.AddToQueue(events.AggroChanged{ActorId: actor.Character.Id, RoomId: actor.Character.RoomId})

			Tell(actor, fmt.Sprintf(`You prepare to fight <ansi fg="mobname">%s</ansi>!`, p.Character.Name))

			if !isSneaking {

				p.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> prepares to fight you!`, NameColour(actor), actor.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> prepares to fight <ansi fg="username">%s</ansi>.`, NameColour(actor), actor.Character.Name, p.Character.Name),
					selfUserId, attackPlayerId)
			}

		}

	}

	return true, nil
}
