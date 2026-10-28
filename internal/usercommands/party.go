package usercommands

import (
	"fmt"
	"math"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Party(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	args := util.SplitButRespectQuotes(rest)

	partyCommand := `list`
	if len(args) > 0 {
		partyCommand = strings.ToLower(args[0])
		rest, _ = strings.CutPrefix(rest, args[0])
		rest = strings.TrimSpace(rest)
	}

	currentParty := parties.Get(user.UserId)

	if partyCommand == `create` || partyCommand == `new` || partyCommand == `start` {

		// check if they are already part of a party
		if currentParty != nil {
			if currentParty.Invited(user.UserId) {
				Tell(actor, `You already have a pending party invite. Try <ansi fg="command">party accept/decline</ansi> first`)
			} else if currentParty.IsLeader(user.UserId) {
				Tell(actor, `You already own a party Type <ansi fg="command">party list</ansi> for more info.`)
			} else {
				Tell(actor, `You are already party of a party.`)
			}
			return true, nil
		}

		if currentParty = parties.New(user.UserId); currentParty != nil {
			user.EventLog.Add(`party`, `Started a new party`)
			Tell(actor, `You started a new party!`)

			//
			// User started a new party
			//
			events.AddToQueue(events.PartyUpdated{
				Action:  `created`,
				UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
			})

		} else {
			Tell(actor, `Something went wrong.`)
		}

		return true, nil
	}
	// Done with create

	//
	// Everything after this point requires a party or an invitation to a party
	//

	if partyCommand == `invite` {

		if rest == `` {
			Tell(actor, `Invite who?`)
			return true, nil
		}

		// Not in a party? Create one.
		if currentParty == nil {
			currentParty = parties.New(user.UserId)
		}

		if !currentParty.IsLeader(user.UserId) {
			Tell(actor, `You are not the leader of your party.`)
			return true, nil
		}

		partyCfg := configs.GetGamePlayConfig().Party

		if maxCount := int(partyCfg.MaxPlayerCount); maxCount > 0 {
			if len(currentParty.UserIds)+len(currentParty.InviteUserIds) >= maxCount {
				Tell(actor, fmt.Sprintf(`Your party is full (%d/%d).`, len(currentParty.UserIds), maxCount))
				return true, nil
			}
		}

		var invitePlayerId int
		if bool(partyCfg.SameRoomOnly) {
			invitePlayerId, _ = room.FindByName(rest)
		} else {
			if u := accounts.GetByCharacterName(rest); u != nil {
				invitePlayerId = u.UserId
			}
		}

		if invitePlayerId == 0 {
			Tell(actor, fmt.Sprintf(`%s not found.`, rest))
			return true, nil
		}

		if invitedParty := parties.Get(invitePlayerId); invitedParty != nil {
			Tell(actor, `That player is already in a party.`)
			return true, nil
		}

		invitedUser := accounts.GetByUserId(invitePlayerId)

		if invitedUser != nil && currentParty.InvitePlayer(invitePlayerId) {
			Tell(actor, fmt.Sprintf(`You invited <ansi fg="username">%s</ansi> to your party.`, invitedUser.Character.Name))
			messaging.ToAccount(invitedUser, fmt.Sprintf(`<ansi fg="username">%s</ansi> invited you to their party. Type <ansi fg="command">party accept</ansi> or <ansi fg="command">party decline</ansi> to respond.`, user.Character.Name))
		} else {
			Tell(actor, `Something went wrong.`)
		}

		//
		// A new user was invited to the party
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `invited`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})

		return true, nil
	}

	//
	// what follows doesn't mamke sense unless they are in a party
	//

	if currentParty == nil {
		Tell(actor, `You are not attached to a party.`)
		return true, nil
	}

	if partyCommand == `accept` || partyCommand == `join` {

		if !currentParty.Invited(user.UserId) {
			Tell(actor, `You haven't accepted an invitation to the party.`)
			return true, nil
		}

		if bool(configs.GetGamePlayConfig().Party.SameRoomOnly) {
			leader := accounts.GetByUserId(currentParty.LeaderUserId)
			if leader == nil || leader.Character.RoomId != user.Character.RoomId {
				Tell(actor, `You must be in the same room as the party leader to join.`)
				return true, nil
			}
		}

		if currentParty.AcceptInvite(user.UserId) {

			user.EventLog.Add(`party`, `Joined a party`)
			Tell(actor, `You joined the party!`)
			for _, uid := range currentParty.UserIds {
				if uid == user.UserId {
					continue
				}
				if u := accounts.GetByUserId(uid); u != nil {
					messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> joined the party!`, user.Character.Name))
				}
			}

			//
			// A user joined the party
			//
			events.AddToQueue(events.PartyUpdated{
				Action:  `joined`,
				UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
			})

		} else {
			Tell(actor, `Something went wrong.`)
		}
		return true, nil
	}

	if partyCommand == `decline` {

		//
		// User declined invitation
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `declined`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})

		if currentParty.DeclineInvite(user.UserId) {

			if u := accounts.GetByUserId(currentParty.LeaderUserId); u != nil {
				messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> declined the invitation.`, user.Character.Name))
			}
			Tell(actor, `You decline the invitation.`)

		} else {
			Tell(actor, `Something went wrong.`)
		}
		return true, nil
	}

	if partyCommand == `list` {

		//headers := []string{"Name", "Status", "Lvl", "Health", "%", "Location", "Position"}
		headers := []string{"Name", "Status", "Lvl", "Health", "Location", "Position"}
		formatting := [][]string{}

		rows := [][]string{}

		if currentParty != nil {
			isInvited := currentParty.Invited(user.UserId)
			leaderId := currentParty.LeaderUserId

			charmedMobIds := []ids.ID{}

			for _, uid := range currentParty.UserIds {
				uStatus := "In Party"
				if leaderId == uid {
					uStatus = "Leader"
				}

				u := accounts.GetByUserId(uid)
				if u == nil {
					continue
				}
				uLevel := fmt.Sprintf(`%d`, u.Character.Level)
				uRoom := rooms.LoadRoom(u.Character.RoomId)
				uLoc := `-`
				if uRoom != nil {
					uLoc = uRoom.Title
				}
				uHealthPct := int(math.Floor((float64(u.Character.Health) / float64(u.Character.HealthMax.Value)) * 100))
				uHealthPctStr := fmt.Sprintf(`%d%%`, uHealthPct)
				rank := currentParty.GetRank(u.UserId)
				healthClass := util.HealthClass(u.Character.Health, u.Character.HealthMax.Value)

				if isInvited {
					uLevel = `-`
					//uHealth = `-`
					uLoc = `-`
					uHealthPctStr = `-`
					rank = `-`
					healthClass = `black-bold`
				}

				rows = append(rows, []string{
					u.Character.Name,
					uStatus,
					uLevel,
					//uHealth,
					uHealthPctStr,
					uLoc,
					rank,
				})

				rowFormat := []string{`<ansi fg="username">%s</ansi>`,
					`<ansi fg="white-bold">%s</ansi>`,
					`<ansi fg="yellow">%s</ansi>`,
					//`<ansi fg="cyan-bold">%s</ansi>`,
					`<ansi fg="` + healthClass + `">%s</ansi>`,
					`<ansi fg="magenta-bold">%s</ansi>`,
					`<ansi fg="white-bold">%s</ansi>`}

				formatting = append(formatting, rowFormat)

				charmedMobIds = append(charmedMobIds, u.Character.GetCharmIds()...)
			}

			for _, charmedMobId := range charmedMobIds {
				m := mobs.GetInstance(charmedMobId)
				if m == nil {
					continue
				}
				mRoom := rooms.LoadRoom(m.Character.RoomId)
				mLoc := `-`
				if mRoom != nil {
					mLoc = mRoom.Title
				}
				mHealthPct := int(math.Floor((float64(m.Character.Health) / float64(m.Character.HealthMax.Value)) * 100))
				rows = append(rows, []string{
					m.Character.Name,
					`♥friend`,
					fmt.Sprintf(`%d`, m.Character.Level),
					//fmt.Sprintf(`%d/%d`, m.Character.Health, m.Character.HealthMax.Value),
					fmt.Sprintf(`%d%%`, mHealthPct),
					mLoc,
					`-`,
				})

				rowFormat := []string{`<ansi fg="username">%s</ansi>`,
					`<ansi fg="white-bold">%s</ansi>`,
					`<ansi fg="yellow">%s</ansi>`,
					//`<ansi fg="cyan-bold">%s</ansi>`,
					`<ansi fg="` + util.HealthClass(m.Character.Health, m.Character.HealthMax.Value) + `">%s</ansi>`,
					`<ansi fg="magenta-bold">%s</ansi>`,
					`<ansi fg="white-bold">%s</ansi>`}

				formatting = append(formatting, rowFormat)
			}

			for _, uid := range currentParty.InviteUserIds {
				u := accounts.GetByUserId(uid)
				if u == nil {
					continue
				}
				rows = append(rows, []string{
					u.Character.Name,
					`Invited`,
					`-`,
					`-`,
					//`-`,
					`-`,
					`-`,
				})

				rowFormat := []string{`<ansi fg="username">%s</ansi>`,
					`<ansi fg="white-bold">%s</ansi>`,
					`<ansi fg="yellow">%s</ansi>`,
					//`<ansi fg="cyan-bold">%s</ansi>`,
					`<ansi fg="black-bold">%s</ansi>`,
					`<ansi fg="magenta-bold">%s</ansi>`,
					`<ansi fg="white-bold">%s</ansi>`}

				formatting = append(formatting, rowFormat)

			}

			partyTableData := templates.GetTable(`Party Members`, headers, rows, formatting...)
			partyTxt, _ := templates.Process("tables/generic", partyTableData, user.UserId)
			Tell(actor, partyTxt)

			if isInvited {
				Tell(actor, `Type <ansi fg="command">party accept/decline</ansi> to finalize your party membership.`)
			}
		}
	}

	if currentParty.Invited(user.UserId) {
		Tell(actor, `You haven't accepted an invitation to the party.`)
		return true, nil
	}

	//
	// Everything after this point you must be in a party
	//
	if partyCommand == `autoattack` {
		autoAttackOn := false
		if rest == `on` {
			autoAttackOn = true
		} else if rest == `off` {
			autoAttackOn = false
		} else {
			Tell(actor, `Usage: <ansi fg="command">party autoattack [on/off]</ansi>`)
			return true, nil
		}

		wasOnBefore := currentParty.SetAutoAttack(user.UserId, autoAttackOn)

		if autoAttackOn {
			if wasOnBefore {
				Tell(actor, `You already have auto-attack enabled.`)
			} else {
				Tell(actor, `You are now auto-attacking with your party.`)
			}
		} else {
			if wasOnBefore {
				Tell(actor, `You are no longer auto-attacking with your party.`)
			} else {
				Tell(actor, `You already have auto-attacking disabled.`)
			}
		}

		//
		// User party behavior changed
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `behavior`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})
	}

	if partyCommand == `leave` || partyCommand == `quit` {

		if currentParty.IsLeader(user.UserId) {

			if len(currentParty.UserIds) <= 1 {

				//
				// Party is disbanded
				//
				events.AddToQueue(events.PartyUpdated{
					Action:  `disbanded`,
					UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
				})

				user.EventLog.Add(`party`, `Disbanded your party`)
				Tell(actor, `You disbanded the party.`)
				currentParty.Disband()

				return true, nil
			}

			currentParty.LeaderUserId = 0

			// promote someone else to leader
			for _, uid := range currentParty.UserIds {
				if uid == user.UserId {
					continue
				}

				newLeaderUser := accounts.GetByUserId(uid)

				if newLeaderUser == nil {
					continue
				}

				currentParty.LeaderUserId = uid

				break
			}

			if currentParty.LeaderUserId > 0 {
				newLeaderUser := accounts.GetByUserId(currentParty.LeaderUserId)
				for _, uid := range currentParty.UserIds {
					if u := accounts.GetByUserId(uid); u != nil {
						if currentParty.LeaderUserId == uid {
							u.EventLog.Add(`party`, `Promoted to party leader`)
							messaging.ToAccount(u, `You are now the leader of the party.`)
						} else {
							messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now the leader of the party.`, newLeaderUser.Character.Name))
						}
					}
				}

				//
				// New user is promoted to leader
				//
				events.AddToQueue(events.PartyUpdated{
					Action:  `promotion`,
					UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
				})
			}

			currentParty.Leave(user.UserId)
			user.EventLog.Add(`party`, `Left the party`)
			Tell(actor, `You left the party.`)

			return true, nil
		}

		//
		// User is leaving the party
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `left`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})

		currentParty.Leave(user.UserId)
		user.EventLog.Add(`party`, `Left the party`)
		Tell(actor, `You left the party.`)

		for _, uid := range currentParty.UserIds {
			if uid == user.UserId {
				continue
			}
			if u := accounts.GetByUserId(uid); u != nil {
				messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> left the party.`, user.Character.Name))
			}
		}

	}

	if partyCommand == `disband` || partyCommand == `stop` {

		if !currentParty.IsLeader(user.UserId) {
			Tell(actor, `You are not the leader of your party.`)
			return true, nil
		}

		for _, uid := range currentParty.UserIds {
			if uid == user.UserId {
				continue
			}
			if u := accounts.GetByUserId(uid); u != nil {
				messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> disbanded the party.`, user.Character.Name))
			}
		}
		for _, uid := range currentParty.InviteUserIds {
			if u := accounts.GetByUserId(uid); u != nil {
				messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> disbanded the party.`, user.Character.Name))
			}
		}

		//
		// The party is being disbanded
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `disbanded`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})

		currentParty.Disband()
		user.EventLog.Add(`party`, `Disbanded the party`)
		Tell(actor, `You disbanded the party.`)

		return true, nil
	}

	if partyCommand == `kick` {

		if !currentParty.IsLeader(user.UserId) {
			Tell(actor, `You are not the leader of your party.`)
			return true, nil
		}

		allMembers := []string{}
		memberIds := map[string]int{}
		for _, uid := range currentParty.GetMembers() {
			u := accounts.GetByUserId(uid)
			if u == nil {
				continue
			}
			allMembers = append(allMembers, u.Character.Name)
			memberIds[u.Character.Name] = uid
		}

		matchUser, closeMatchUser := util.FindMatchIn(rest, allMembers...)
		if matchUser == `` {
			matchUser = closeMatchUser
		}

		if matchUser == `` {
			Tell(actor, fmt.Sprintf(`%s not found.`, rest))
			return true, nil
		}

		//
		// The user was kicked from the party
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `left`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})

		kickUserId := memberIds[matchUser]

		currentParty.Leave(kickUserId)

		if u := accounts.GetByUserId(kickUserId); u != nil {
			messaging.ToAccount(u, `You were kicked from the party.`)
		}

		for _, uid := range currentParty.UserIds {
			if u := accounts.GetByUserId(uid); u != nil {
				messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> was kicked from the party.`, matchUser))
			}
		}
	}

	if partyCommand == `promote` {

		if !currentParty.IsLeader(user.UserId) {
			Tell(actor, `You are not the leader of your party.`)
			return true, nil
		}

		allMembers := []string{}
		memberIds := map[string]int{}
		for _, uid := range currentParty.GetMembers() {
			u := accounts.GetByUserId(uid)
			if u == nil {
				continue
			}
			allMembers = append(allMembers, u.Character.Name)
			memberIds[u.Character.Name] = uid
		}

		matchUser, closeMatchUser := util.FindMatchIn(rest, allMembers...)
		if matchUser == `` {
			matchUser = closeMatchUser
		}

		if matchUser == `` {
			Tell(actor, fmt.Sprintf(`%s not found.`, rest))
			return true, nil
		}

		//
		// User was promoted to leader
		//
		events.AddToQueue(events.PartyUpdated{
			Action:  `promotion`,
			UserIds: append(currentParty.GetMembers(), currentParty.GetInvited()...),
		})

		promoteUserId := memberIds[matchUser]

		currentParty.LeaderUserId = promoteUserId

		if u := accounts.GetByUserId(promoteUserId); u != nil {
			u.EventLog.Add(`party`, `Promoted to party leader`)
			messaging.ToAccount(u, `You have been promoted to party leader.`)
		}

		for _, uid := range currentParty.UserIds {
			if uid != promoteUserId {
				if u := accounts.GetByUserId(uid); u != nil {
					messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="username">%s</ansi> is now the party leader.`, matchUser))
				}
			}
		}

	}

	if partyCommand == `chat` || partyCommand == `say` {

		if len(rest) == 0 {
			Tell(actor, `What do you want to say?`)
			return true, nil
		}

		for _, uId := range currentParty.GetMembers() {
			if uId == user.UserId {
				continue
			}
			if u := accounts.GetByUserId(uId); u != nil {
				messaging.ToAccount(u, fmt.Sprintf(`<ansi fg="magenta">(party)</ansi> <ansi fg="username">%s</ansi> says, "<ansi fg="yellow">%s</ansi>`, user.Character.Name, rest))
			}
		}

		Tell(actor, fmt.Sprintf(`<ansi fg="magenta">(party)</ansi> You say, "<ansi fg="yellow">%s</ansi>"`, rest))

		events.AddToQueue(events.Communication{
			SourceActorId: user.Character.Id,
			CommType:      `party`,
			Name:          user.Character.Name,
			Message:       rest,
		})
	}

	return true, nil
}
