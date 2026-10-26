package scripting

import (
	"github.com/GoMudEngine/GoMud/internal/parties"
)

type ScriptParty struct {
	actor          *ScriptActor
	includeSelf    bool
	includePresent bool
	includeMissing bool
}

// GetMembers() is the core feature of this, everything works off of it.
func (p ScriptParty) GetMembers() []ScriptActor {

	partyMembers := []ScriptActor{}
	addedUsers := map[int]struct{}{} // Track so we never accidentally add someone twice
	addedMobs := map[int]struct{}{}  // Track so we never accidentally add someone twice

	// Usually we would include self, but just in case...
	if p.includeSelf {

		partyMembers = append(partyMembers, *p.actor)

		if p.actor.characterRecord != nil {
			addedUsers[p.actor.userId] = struct{}{}
		} else if p.actor.mobRecord != nil {
			addedMobs[p.actor.mobInstanceId] = struct{}{}
		}
	}

	partyUserId := p.actor.UserId()
	sourceRoomId := p.actor.GetRoomId()

	if p.actor.userRecord == nil {
		if p.actor.mobRecord.Character.Charmed == nil {
			return partyMembers
		}

		partyUserId = p.actor.mobRecord.Character.Charmed.UserId
	}

	if partyUserId < 1 {
		return partyMembers
	}

	// If in a party, give to all party members.
	if party := parties.Get(partyUserId); party != nil {
		for _, userId := range party.GetMembers() {

			if _, ok := addedUsers[userId]; ok {
				continue
			}

			if a := GetActor(userId, 0); a != nil {

				if a.GetRoomId() == sourceRoomId {
					if !p.includePresent {
						continue
					}
				} else {
					if !p.includeMissing {
						continue
					}
				}

				partyMembers = append(partyMembers, *a)
				addedUsers[userId] = struct{}{}

			}

		}
	}

	mobPartyMembers := []ScriptActor{}

	// Add all charmed mobs for all party members, too.
	for _, char := range partyMembers {
		for _, mobInstId := range char.characterRecord.GetCharmIds() {

			if _, ok := addedMobs[mobInstId]; ok {
				continue
			}

			if a := GetActor(0, mobInstId); a != nil {

				if a.GetRoomId() == sourceRoomId {
					if !p.includePresent {
						continue
					}
				} else {
					if !p.includeMissing {
						continue
					}
				}

				mobPartyMembers = append(mobPartyMembers, *a)
				addedMobs[mobInstId] = struct{}{}

			}
		}
	}

	return append(partyMembers, mobPartyMembers...)
}

//
// Simple helper to loop through all members and apply a function
//

func (p ScriptParty) each(fn func(ScriptActor)) {
	allParty := p.GetMembers()
	for _, a := range allParty {
		fn(a)
	}
}

//
// What follows are many functions graduated from the actor level to the party level
//

func (p ScriptParty) SendText(msg string) {
	p.each(func(a ScriptActor) {
		a.SendText(msg)
	})
}

func (p ScriptParty) SetResetRoomId(roomId int) {
	p.each(func(a ScriptActor) {
		if a.userRecord == nil {
			return
		}
		a.userRecord.Character.RoomIdOnReset = roomId
	})
}

func (p ScriptParty) AddGold(amt int, bankAmt ...int) {
	p.each(func(a ScriptActor) {
		a.AddGold(amt, bankAmt...)
	})
}
func (p ScriptParty) AddHealth(amt int) {
	p.each(func(a ScriptActor) {
		a.AddHealth(amt)
	})
}
func (p ScriptParty) AddMana(amt int) {
	p.each(func(a ScriptActor) {
		a.AddMana(amt)
	})
}
func (p ScriptParty) Command(cmd string, waitSeconds ...float64) {
	p.each(func(a ScriptActor) {
		a.Command(cmd, waitSeconds...)
	})
}
func (p ScriptParty) TrainSkill(skillName string, skillLevel int) {
	p.each(func(a ScriptActor) {
		a.TrainSkill(skillName, skillLevel)
	})
}
func (p ScriptParty) MoveRoom(destRoomId int) {
	p.each(func(a ScriptActor) {
		a.MoveRoom(destRoomId)
	})
}
func (p ScriptParty) AddEventLog(category string, message string) {
	p.each(func(a ScriptActor) {
		a.AddEventLog(category, message)
	})
}

func (p ScriptParty) GiveBuff(buffId int, source string) {
	p.each(func(a ScriptActor) {
		a.GiveBuff(buffId, source)
	})
}
func (p ScriptParty) CancelBuffWithFlag(buffFlag string) {
	p.each(func(a ScriptActor) {
		a.CancelBuffWithFlag(buffFlag)
	})
}
func (p ScriptParty) RemoveBuff(buffId int) {
	p.each(func(a ScriptActor) {
		a.RemoveBuff(buffId)
	})
}
func (p ScriptParty) ChangeAlignment(alignmentChange int) {
	p.each(func(a ScriptActor) {
		a.ChangeAlignment(alignmentChange)
	})
}
func (p ScriptParty) LearnSpell(spellId string) {
	p.each(func(a ScriptActor) {
		a.LearnSpell(spellId)
	})
}
func (p ScriptParty) SetHealth(amt int) {
	p.each(func(a ScriptActor) {
		a.SetHealth(amt)
	})
}
func (p ScriptParty) SetAdjective(adj string, addIt bool) {
	p.each(func(a ScriptActor) {
		a.SetAdjective(adj, addIt)
	})
}
func (p ScriptParty) GiveTrainingPoints(ct int) {
	p.each(func(a ScriptActor) {
		a.GiveTrainingPoints(ct)
	})
}
func (p ScriptParty) GiveStatPoints(ct int) {
	p.each(func(a ScriptActor) {
		a.GiveStatPoints(ct)
	})
}
func (p ScriptParty) GiveExtraLife() {
	p.each(func(a ScriptActor) {
		a.GiveExtraLife()
	})
}

func (p ScriptParty) GrantXP(xpAmt int, reason string) {
	p.each(func(a ScriptActor) {
		a.GrantXP(xpAmt, reason)
	})
}
func (p ScriptParty) TimerSet(name string, period string) {
	p.each(func(a ScriptActor) {
		a.TimerSet(name, period)
	})
}

func (p ScriptParty) MarkVisitedRoom(roomIds ...int) {
	p.each(func(a ScriptActor) {
		a.MarkVisitedRoom(roomIds...)
	})
}

func (p ScriptParty) MarkVisitedZone(zoneName string) {
	p.each(func(a ScriptActor) {
		a.MarkVisitedZone(zoneName)
	})
}
