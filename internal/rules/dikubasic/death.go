package dikubasic

import (
	"fmt"
	"math"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// OnDeath is a death under the reference rules (decisions 7 and 11, slice 3
// rulings 1 and 3). A body with revive-on-death gets up at full health. An
// NPC's killers share its XP and shift alignment, as GoMud did. An account
// comes back in the recovery room at 1 health; an NPC's body ends. Drops: an
// account drops what it carries and its gold and keeps what it wears; an NPC
// drops what it carries, its gold, and each worn item at its ItemDropChance;
// nothing under perma-gear.
//
// Random draws, in the order usercommands/suicide_npc.go made them: the XP
// variation, any alignment floors, then one d100 per worn NPC item.
func (Ruleset) OnDeath(victim *actors.Actor) rules.DeathOutcome {
	c := victim.Character

	if c.HasBuffFlag("revive-on-death") {
		return rules.DeathOutcome{Revive: true, ReviveHealth: c.HealthMax.Value}
	}

	mob := mobs.GetInstance(victim.Id)
	if mob == nil {
		out := rules.DeathOutcome{Respawn: true, RespawnHealth: 1}
		if !c.HasBuffFlag("perma-gear") {
			out.DropItems = append([]items.Item{}, c.Items...)
			out.DropGold = c.Gold
		}
		return out
	}

	out := rules.DeathOutcome{Awards: killAwards(mob)}
	if c.HasBuffFlag("perma-gear") {
		return out
	}
	out.DropItems = append([]items.Item{}, c.Items...)
	for _, item := range c.Equipment.GetAllItems() {
		if item.IsRemoveLocked() {
			continue
		}

		roll := util.Rand(100)

		util.LogRoll(`Drop Item`, roll, mob.ItemDropChance)

		if roll >= mob.ItemDropChance {
			continue
		}
		out.DropItems = append(out.DropItems, item)
	}
	out.DropGold = c.Gold
	return out
}

// killAwards is the NPC's XP and alignment for its killers, moved from
// usercommands/suicide_npc.go (slice 3) with the engine's parts left out:
// the kill credit, the XP grant and every send are internal/death's.
// Alignment is applied here, in the order it was, because UpdateAlignment
// can draw a random number.
func killAwards(mob *mobs.Mob) []rules.Award {

	mobXP := mob.Character.XPTL(mob.Character.Level - 1)

	xpVal := mobXP / 90

	xpVariation := xpVal / 100
	if xpVariation < 1 {
		xpVariation = 1
	}

	awards := []rules.Award{}
	partyTracker := map[int]int{} // key is party leader ID, value is how much will be shared.

	if len(mob.Character.PlayerDamage) > 0 {

		xpVal = xpVal / len(mob.Character.PlayerDamage) // Div by number of players that beat him up
		xpVal += ((util.Rand(3) - 1) * xpVariation)     // a little bit of variation

		totalPlayerLevels := 0
		for uId := range mob.Character.PlayerDamage {
			if user := accounts.GetByUserId(uId); user != nil {
				totalPlayerLevels += user.Character.Level
			}
		}

		for uId := range mob.Character.PlayerDamage {
			user := accounts.GetByUserId(uId)
			if user == nil {
				continue
			}

			// In a party: the party shares it, below.
			if p := parties.Get(user.UserId); p != nil {
				partyTracker[p.LeaderUserId] += xpVal
				continue
			}

			xpScaler := 1.0

			// If there's a level delta of more than 5, apply a scaler
			if math.Abs(float64(mob.Character.Level)-float64(totalPlayerLevels)) > 5 {

				xpScaler = float64(mob.Character.Level) / float64(totalPlayerLevels) // How much of the mobs level is the player?
				if xpScaler > 1.5 {
					xpScaler = 1.5
				} else if xpScaler < 0.25 {
					xpScaler = 0.25
				}

			}

			finalXPVal := int(math.Ceil(float64(xpVal) * xpScaler))

			if mob.IsElite {
				finalXPVal = finalXPVal + int(math.Ceil(float64(finalXPVal)*float64(eliteXPBonus())/100.0))
			}

			mudlog.Debug("XP Calculation", "MobLevel", mob.Character.Level, "XPBase", mobXP, "xpVal", xpVal, "xpVariation", xpVariation, "xpScaler", xpScaler, "finalXPVal", finalXPVal)

			awards = append(awards, rules.Award{UserId: user.UserId, XP: finalXPVal, Messages: shiftAlignment(user, mob)})
		}
	}

	// Party XP: an equal share for every member.
	for leaderId, xp := range partyTracker {
		p := parties.Get(leaderId)
		if p == nil {
			continue
		}

		allMembers := p.GetMembers()
		xpSplit := xp / len(allMembers)

		if mob.IsElite {
			xpSplit = xpSplit + int(math.Ceil(float64(xpSplit)*float64(eliteXPBonus())/100.0))
		}

		mudlog.Info(`Party XP`, `totalXP`, xp, `splitXP`, xpSplit, `memberCt`, len(allMembers))

		for _, memberId := range allMembers {
			if user := accounts.GetByUserId(memberId); user != nil {
				awards = append(awards, rules.Award{UserId: user.UserId, XP: xpSplit, Messages: shiftAlignment(user, mob)})
			}
		}
	}

	return awards
}

// eliteXPBonus is the percent XP bonus for killing an elite NPC (10 when
// unset).
func eliteXPBonus() int {
	eliteBonus := int(configs.GetGamePlayConfig().EliteXPBonus)
	if eliteBonus <= 0 {
		eliteBonus = 10
	}
	return eliteBonus
}

// shiftAlignment applies the alignment change for user killing mob and
// returns the line that tells the player, if the alignment's name changed.
func shiftAlignment(user *accounts.Account, mob *mobs.Mob) []perception.Message {
	alignmentBefore := user.Character.AlignmentName()
	alignmentAdj := AlignmentChange(user.Character.Alignment, mob.Character.Alignment)
	user.Character.UpdateAlignment(alignmentAdj)
	alignmentAfter := user.Character.AlignmentName()

	mudlog.Debug("Alignment", "user Alignment", user.Character.Alignment, "mob Alignment", mob.Character.Alignment, `alignmentAdj`, alignmentAdj, `alignmentBefore`, alignmentBefore, `alignmentAfter`, alignmentAfter)

	if alignmentBefore == alignmentAfter {
		return nil
	}

	events.AddToQueue(events.AlignmentChanged{
		UserId:       user.UserId,
		AlignmentOld: alignmentBefore,
		AlignmentNew: alignmentAfter,
	})

	before := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentBefore, alignmentBefore)
	after := fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, alignmentAfter, alignmentAfter)
	return []perception.Message{{
		Kind:      perception.Interface,
		Actor:     user.Character.Id,
		Audience:  perception.ActorOnly,
		ActorText: fmt.Sprintf(`<ansi fg="231">Your alignment has shifted from %s to %s!</ansi>`, before, after),
	}}
}
