package combat

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/statmods"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Attack performs one combat round of attacker's attacks on defender, for
// any pair of actors. It applies the damage to the defender, credits the
// attacker's account (or its charmer's) in the defender's PlayerDamage on
// every attack, hit or miss (plan Decision 3), and runs the defender's
// wimpy check when the defender has an account and took damage (Decision
// 9). The result's Messages are the round's lines, one per swing, for the
// caller to send in order.
func Attack(attacker, defender *actors.Actor) AttackResult {
	if attacker == nil || defender == nil {
		return AttackResult{}
	}

	attackResult := calculateCombat(*attacker.Character, *defender.Character)

	defender.Character.ApplyHealthChange(attackResult.DamageToTarget * -1)

	// Remember who has hit him
	if acct := CreditAccount(attacker); acct != nil {
		defender.Character.TrackPlayerDamage(acct.UserId, attackResult.DamageToTarget)
	}

	if attackResult.DamageToTarget != 0 {
		if acct := accountOf(defender); acct != nil {
			acct.WimpyCheck()
		}
	}

	return attackResult
}

// accountOf is the account driving a, or nil for an NPC.
func accountOf(a *actors.Actor) *accounts.Account {
	if ctrl := a.Controller(); ctrl != nil {
		return ctrl.Account()
	}
	return nil
}

// CreditAccount is the account a's attacks are credited to: its own, else
// the account of whoever charmed it, else none. Kill credit and group
// hatred both follow it (plan Decisions 3 and 8).
func CreditAccount(a *actors.Actor) *accounts.Account {
	if acct := accountOf(a); acct != nil {
		return acct
	}
	if by := a.Character.GetCharmedBy(); !by.IsZero() {
		return actors.AccountOf(by)
	}
	return nil
}

// combatMessage is one combat line set as a world event: seen or heard by
// the source rooms (the attacker's, and the defender's when it differs).
// Sound carries it too because combat lines have always reached players in
// a dark room (room sends ignore light); in the dark, messaging renders the
// unseen names as "someone".
func combatMessage(attacker, defender ids.ID, actorText, targetText, observerText, targetRoomText string) perception.Message {
	return perception.Message{
		Kind:           perception.WorldEvent,
		Actor:          attacker,
		Target:         defender,
		Senses:         perception.Sight | perception.Sound,
		Range:          perception.Room,
		ActorText:      actorText,
		TargetText:     targetText,
		ObserverText:   observerText,
		TargetRoomText: targetRoomText,
		// An empty defender-room line means nothing for a remote target's
		// room; messaging only applies this when the rooms differ.
		NoTargetRoomText: targetRoomText == ``,
	}
}

// GetWaitMessages returns the lines for a round the attacker spends waiting
// (stepType items.Wait) or preparing, as at most one message.
func GetWaitMessages(stepType items.Intensity, sourceChar *characters.Character, targetChar *characters.Character) []perception.Message {

	msgs := items.GetPreAttackMessage(sourceChar.Equipment.Weapon.GetSpec().Subtype, stepType)

	// zero means randomly selected, otherwise use the ItemId to consistently choose a message
	msgSeed := 0
	if configs.GetCombatConfig().ConsistentAttackMessages {
		msgSeed = sourceChar.Equipment.Weapon.ItemId
	}

	weaponName := races.GetRace(sourceChar.GetRaceId()).UnarmedName
	if sourceChar.Equipment.Weapon.ItemId > 0 {
		weaponName = sourceChar.Equipment.Weapon.DisplayName()
	}

	toAttackerMsg, toDefenderMsg, toAttackerRoomMsg, toDefenderRoomMsg := buildCombatMessages(
		sourceChar, targetChar,
		weaponName, `[Invalid]`, msgSeed,
		msgs.Together.ToAttacker, msgs.Together.ToDefender, msgs.Together.ToRoom, items.MessageOptions(nil),
		msgs.Separate.ToAttacker, msgs.Separate.ToDefender, msgs.Separate.ToAttackerRoom, msgs.Separate.ToDefenderRoom,
	)

	if toAttackerMsg == `` && toDefenderMsg == `` && toAttackerRoomMsg == `` && toDefenderRoomMsg == `` {
		return nil
	}

	return []perception.Message{
		combatMessage(sourceChar.Id, targetChar.Id, toAttackerMsg, toDefenderMsg, toAttackerRoomMsg, toDefenderRoomMsg),
	}
}

// buildCombatMessages selects the together/separate message variants for a
// single combat message event and fills every token but the names. {actor}
// and {target} are left for messaging, which resolves them per receiver.
// Pass nil (or empty MessageOptions) for togetherToDefenderRoom to suppress that message slot.
func buildCombatMessages(
	sourceChar *characters.Character, targetChar *characters.Character,
	weaponName string, damageStr string, msgSeed int,
	togetherToAttacker, togetherToDefender, togetherToRoom, togetherToDefenderRoom items.MessageOptions,
	separateToAttacker, separateToDefender, separateToAttackerRoom, separateToDefenderRoom items.MessageOptions,
) (toAttackerMsg, toDefenderMsg, toAttackerRoomMsg, toDefenderRoomMsg string) {

	exitName, entranceName := `unknown`, `unknown`

	var atk, def, atkRoom, defRoom items.ItemMessage

	if sourceChar.RoomId == targetChar.RoomId {

		atk = togetherToAttacker.Get(msgSeed)
		def = togetherToDefender.Get(msgSeed)
		atkRoom = togetherToRoom.Get(msgSeed)
		defRoom = togetherToDefenderRoom.Get(msgSeed)

	} else {

		atk = separateToAttacker.Get(msgSeed)
		def = separateToDefender.Get(msgSeed)
		atkRoom = separateToAttackerRoom.Get(msgSeed)
		defRoom = separateToDefenderRoom.Get(msgSeed)

		// Find the exit that leads to the target from the source (if any)
		if r := rooms.LoadRoom(sourceChar.RoomId); r != nil {
			for name, exit := range r.Exits {
				if exit.RoomId == targetChar.RoomId {
					exitName = name
					break
				}
			}
		}
		// Find the exit that leads to the source from the target (if any)
		if r := rooms.LoadRoom(targetChar.RoomId); r != nil {
			for name, exit := range r.Exits {
				if exit.RoomId == sourceChar.RoomId {
					entranceName = name
					break
				}
			}
		}
	}

	fill := strings.NewReplacer(
		string(items.TokenItemName), weaponName,
		string(items.TokenUsesLeft), `[Invalid]`,
		string(items.TokenDamage), damageStr,
		string(items.TokenEntranceName), entranceName,
		string(items.TokenExitName), exitName,
	)

	return fill.Replace(string(atk)), fill.Replace(string(def)), fill.Replace(string(atkRoom)), fill.Replace(string(defRoom))
}

func calculateCombat(sourceChar characters.Character, targetChar characters.Character) AttackResult {

	attackResult := AttackResult{}

	atkCount := combatAttackCount(sourceChar, targetChar)

	// Statmods can add a damage bonus plus the stat-driven damage bonus.
	statModDBonus := sourceChar.StatMod(`damage`) + damageBonus(sourceChar.Stats.Strength.ValueAdj, targetChar.Stats.Strength.ValueAdj)

	for i := 0; i < atkCount; i++ {

		mudlog.Debug(`calculateCombat`, `Atk`, fmt.Sprintf(`%d/%d`, i+1, atkCount), `Source`, sourceChar.Name, `Target`, targetChar.Name)

		attackWeapons := resolveAttackWeapons(sourceChar)

		dualWieldLevel := sourceChar.GetSkillLevel(`dual-wield`)

		if len(attackWeapons) > 1 {
			bothClaws := sourceChar.Equipment.Weapon.GetSpec().Subtype == items.Claws &&
				sourceChar.Equipment.Offhand.GetSpec().Subtype == items.Claws
			maxWeapons := dualWieldActiveWeaponCount(dualWieldLevel, bothClaws)
			util.LogRoll(`Both Weapons`, maxWeapons, 2)

			for len(attackWeapons) > maxWeapons {
				rnd := util.Rand(len(attackWeapons))
				attackWeapons = append(attackWeapons[:rnd], attackWeapons[rnd+1:]...)
			}
		}

		attackMessagePrefix := ``
		// If they are backstabbing it's a free crit
		if sourceChar.Aggro.Type == characters.BackStab {
			attackResult.Crit = true
			attackMessagePrefix = `<ansi fg="magenta-bold">*[BACKSTAB]*</ansi> `
			// Failover to the default attack
			sourceChar.SetAggro(sourceChar.Aggro.Target, characters.DefaultAttack)
		}

		for wIdx, weapon := range attackWeapons {

			// Only the offhand weapon (index > 0) incurs a hit penalty for dual-wielding.
			penalty := 0
			if wIdx > 0 {
				penalty = dualWieldHitPenalty(dualWieldLevel)
			}

			// Set the default weapon info
			raceInfo := races.GetRace(sourceChar.GetRaceId())
			weaponName := raceInfo.UnarmedName
			weaponSubType := items.Generic

			// Get default racial dice rolls
			attacks, dCount, dSides, dBonus, critBuffs := sourceChar.GetDefaultDiceRoll()

			if weapon.ItemId > 0 {

				itemSpec := weapon.GetSpec()

				weaponName = weapon.DisplayName()

				weaponSubType = itemSpec.Subtype
				attacks, dCount, dSides, dBonus, critBuffs = weapon.GetDiceRoll()

				// If there is a bonus vs. a specific race, apply it
				dBonus += weapon.StatMod(string(statmods.RacialBonusPrefix) + strings.ToLower(targetChar.Race()))
			}

			// Apply damage stat modifier after weapon selection so it is never overwritten.
			dBonus += statModDBonus

			// zero means randomly selected, otherwise use the ItemId to consistently choose a message
			msgSeed := 0
			if configs.GetCombatConfig().ConsistentAttackMessages {
				msgSeed = weapon.ItemId
			}

			mudlog.Debug("DiceRolls", "attacks", attacks, "dCount", dCount, "dSides", dSides, "dBonus", dBonus, "critBuffs", critBuffs)

			// Individual weapons may get multiple attacks
			for j := 0; j < attacks; j++ {

				attackTargetDamage := 0
				attackTargetReduction := 0
				isCrit := false

				if Hits(sourceChar.Stats.Speed.ValueAdj, targetChar.Stats.Speed.ValueAdj, penalty) {
					// Check dodge before applying damage.
					if Dodges(targetChar.Stats.Perception.ValueAdj, sourceChar.Stats.Perception.ValueAdj) {
						attackResult.Messages = append(attackResult.Messages, combatMessage(sourceChar.Id, targetChar.Id,
							`<ansi fg="cyan">{target} dodges your attack!</ansi>`,
							`<ansi fg="cyan">You dodge the attack!</ansi>`,
							``, ``))
						continue
					}
					attackResult.Hit = true
					attackTargetDamage = util.RollDice(dCount, dSides) + dBonus

					// Backstab sets attackResult.Crit for the first hit only; subsequent
					// hits use a fresh per-attack roll so crits don't cascade.
					isCrit = attackResult.Crit || Crits(sourceChar, targetChar)
					attackResult.Crit = false // consume the backstab flag after one use
					if isCrit {
						attackResult.Crit = true // record that at least one crit occurred this round
						attackResult.BuffTarget = critBuffs
						attackTargetDamage += critDamageBonus(dCount, dSides, dBonus,
							sourceChar.Stats.Perception.ValueAdj, targetChar.Stats.Perception.ValueAdj)
					}
				}

				attackTargetDamage, attackTargetReduction = applyDefenseReduction(attackTargetDamage, targetChar.GetDefense())

				pct := damagePercentOfMax(attackTargetDamage, dCount, dSides, dBonus)
				msgs := items.GetAttackMessage(weaponSubType, pct)

				toAttackerMsg, toDefenderMsg, toAttackerRoomMsg, toDefenderRoomMsg := buildCombatMessages(
					&sourceChar, &targetChar,
					weaponName, strconv.Itoa(attackTargetDamage), msgSeed,
					msgs.Together.ToAttacker, msgs.Together.ToDefender, msgs.Together.ToRoom, items.MessageOptions(nil),
					msgs.Separate.ToAttacker, msgs.Separate.ToDefender, msgs.Separate.ToAttackerRoom, msgs.Separate.ToDefenderRoom,
				)

				if isCrit {
					toAttackerMsg = `<ansi fg="yellow-bold">***</ansi> ` + toAttackerMsg + ` <ansi fg="yellow-bold">***</ansi>`
					toDefenderMsg = `<ansi fg="yellow-bold">***</ansi> ` + toDefenderMsg + ` <ansi fg="yellow-bold">***</ansi>`
					toAttackerRoomMsg = `<ansi fg="yellow-bold">***</ansi> ` + toAttackerRoomMsg + ` <ansi fg="yellow-bold">***</ansi>`
					if len(toDefenderRoomMsg) > 0 {
						toDefenderRoomMsg = `<ansi fg="yellow-bold">***</ansi> ` + toDefenderRoomMsg + ` <ansi fg="yellow-bold">***</ansi>`
					}
				}

				if len(attackMessagePrefix) > 0 {
					toAttackerMsg = attackMessagePrefix + toAttackerMsg
					toDefenderMsg = attackMessagePrefix + toDefenderMsg
					toAttackerRoomMsg = attackMessagePrefix + toAttackerRoomMsg
					if len(toDefenderRoomMsg) > 0 {
						toDefenderRoomMsg = attackMessagePrefix + toDefenderRoomMsg
					}
				}

				// The victim's line notes what their defense blocked.
				if attackTargetDamage > 0 && attackTargetReduction > 0 {
					toDefenderMsg += fmt.Sprintf(` <ansi fg="red">[you blocked %d]</ansi>`, attackTargetReduction)
				}

				// One message per swing: the attacker's line, the victim's, the
				// attacker's room's, and the defender's room's when it is another room.
				attackResult.Messages = append(attackResult.Messages, combatMessage(sourceChar.Id, targetChar.Id,
					toAttackerMsg, toDefenderMsg, toAttackerRoomMsg, toDefenderRoomMsg))

				attackResult.DamageToTarget += attackTargetDamage
			}

		}
	}

	return attackResult

}
