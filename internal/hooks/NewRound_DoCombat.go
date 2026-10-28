package hooks

import (
	"fmt"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/combat"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// DoCombat runs one combat round: one turn for every actor with an aggro,
// player or NPC, by the same rules, then the death checks for everyone the
// round touched.
func DoCombat(e events.Event) events.ListenerReturn {

	tStart := time.Now()

	var affected []ids.ID
	for _, a := range fightersWithAccountsFirst() {
		affected = append(affected, combatTurn(a)...)
	}

	// Do any resolution or extra checks based on everyone that has been involved in combat this round.
	handleAffected(affected)

	util.TrackTime(`DoCombat()`, time.Since(tStart).Seconds())

	return events.Continue
}

// fightersWithAccountsFirst lists the actors that may take a combat turn:
// those with an account first, then NPCs (plan Decision 5: a player still
// swings before any NPC in the round). Each group comes from the same ID
// source as before, so its order is map order, as it always was.
func fightersWithAccountsFirst() []*actors.Actor {
	fighters := []*actors.Actor{}
	for _, userId := range accounts.GetOnlineUserIds() {
		if a := actors.ByUserId(userId); a != nil {
			fighters = append(fighters, a)
		}
	}
	for _, mobId := range mobs.GetAllMobIds() {
		if a := actors.Get(mobId); a != nil {
			fighters = append(fighters, a)
		}
	}
	return fighters
}

// accountOf is the account driving a, or nil for an NPC.
func accountOf(a *actors.Actor) *accounts.Account {
	if ctrl := a.Controller(); ctrl != nil {
		return ctrl.Account()
	}
	return nil
}

// endAggro ends a's fight and says so.
func endAggro(a *actors.Actor) {
	a.Character.EndAggro()
	events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
}

// combatTurn is one actor's turn in the round. It returns the actors whose
// health the round must check afterwards.
func combatTurn(a *actors.Actor) (affected []ids.ID) {

	if a == nil || a.Character == nil || a.Character.Aggro == nil {
		return nil
	}

	// If has a buff that prevents combat, skip the actor
	if a.Character.HasBuffFlag("no-combat") {
		return nil
	}

	aRoom := rooms.LoadRoom(a.Character.RoomId)
	if aRoom == nil {
		endAggro(a)
		return nil
	}

	// Disable any buffs that are cancelled by combat
	a.Character.CancelBuffsWithFlag("cancel-on-combat")

	// An NPC's combat commands (behaviour data on its template) may take its
	// turn instead of the attack.
	if mob := mobs.GetInstance(a.Id); mob != nil && a.Character.Aggro.Type == characters.DefaultAttack {
		if combatCommand(mob) {
			return nil
		}
	}

	if a.Character.Aggro.Type == characters.Flee {
		flee(a, aRoom)
		return nil
	}

	affected = append(affected, a.Id)

	/**************************
	*
	* START HANDLING PHYSICAL COMBAT
	*
	**************************/

	d := actors.Get(a.Character.Aggro.Target)

	// The target is in the room, or (plan Decision 6) through the exit the
	// aggro names.
	targetFound := d != nil && d.Character != nil
	if targetFound && d.Character.RoomId != a.Character.RoomId {
		if a.Character.Aggro.ExitName == `` {
			targetFound = false
		} else if _, exitRoomId := aRoom.FindExitByName(a.Character.Aggro.ExitName); exitRoomId != d.Character.RoomId {
			targetFound = false
		}
	}

	if !targetFound {
		usercommands.Tell(a, `Your target can't be found.`)
		endAggro(a)
		return affected
	}

	defRoom := rooms.LoadRoom(d.Character.RoomId)
	if defRoom == nil {
		endAggro(a)
		return affected
	}

	d.Character.CancelBuffsWithFlag("cancel-on-combat")

	if d.Character.Health < 1 {
		usercommands.Tell(a, `Your rage subsides.`)
		endAggro(a)
		return affected
	}

	affected = append(affected, d.Id)

	// An NPC with no weapon but something in its pack may look for a weapon
	// (especially useful when disarmed).
	if mob := mobs.GetInstance(a.Id); mob != nil {
		lookForWeapon(mob)
	}

	if a.Character.Aggro.RoundsWaiting > 0 {
		mudlog.Debug(`RoundsWaiting`, `Actor`, a.Character.Name, `Rounds`, a.Character.Aggro.RoundsWaiting)

		a.Character.Aggro.RoundsWaiting--

		for _, m := range combat.GetWaitMessages(items.Wait, a.Character, d.Character) {
			messaging.Send(m)
		}

		return affected
	}

	// Can't see them, can't fight them.
	if d.Character.HasBuffFlag("hidden") {
		usercommands.Tell(a, "You can't seem to find your target.")
		return affected
	}

	healthBefore := d.Character.Health

	roundResult := combat.Attack(a, d)

	for _, buffId := range roundResult.BuffTarget {
		events.AddToQueue(events.Buff{ActorId: d.Id, BuffId: buffId, Source: `combat`})
	}

	for _, m := range roundResult.Messages {
		messaging.Send(m)
	}

	defAcct := accountOf(d)

	// Combat damage downed a player: remember who did it (plan Decision 7),
	// for the death that follows.
	if defAcct != nil && healthBefore >= 1 && d.Character.Health < 1 {
		d.Character.KilledBy = a.Id
		d.Character.KillerMobIsElite = false
		d.Character.KillerMobName = ``
		d.Character.KillerMobId = 0
		// Recorded now, so the death does not need the killer to exist still.
		if killerMob := mobs.GetInstance(a.Id); killerMob != nil {
			d.Character.KillerMobIsElite = killerMob.IsElite
			d.Character.KillerMobName = killerMob.Character.Name
			d.Character.KillerMobId = int(killerMob.MobId)
		}
	}

	defMob := mobs.GetInstance(d.Id)

	// Group hatred follows kill credit (plan Decision 8): the attacker's
	// account, or its charmer's.
	if defMob != nil {
		if acct := combat.CreditAccount(a); acct != nil {
			c := configs.GetConfig()
			for _, groupName := range defMob.Groups {
				mobs.MakeHostile(groupName, acct.UserId, c.Timing.MinutesToRounds(2)-a.Character.Stats.Perception.ValueAdj)
			}
		}
	}

	// If the attack connected, check for damage to equipment.
	if roundResult.Hit {
		offhandBreakCheck(d, defRoom, roundResult.Crit)
	}

	if defAcct != nil && roundResult.DamageToTarget != 0 {
		events.AddToQueue(events.CharacterVitalsChanged{UserId: defAcct.UserId})
	}

	// An NPC that is attacked and not fighting stays put; attacked from
	// another room, it heads that way.
	if defMob != nil && d.Character.Aggro == nil {
		defMob.PreventIdle = true
		if a.Character.RoomId != d.Character.RoomId {
			for exitName, exitInfo := range defRoom.Exits {
				if exitInfo.RoomId == a.Character.RoomId {
					defMob.Command(fmt.Sprintf(`go %s`, exitName))
					if actionStr := defMob.GetAngryCommand(); actionStr != `` {
						defMob.Command(actionStr)
					}
					break
				}
			}
		}
	}

	if a.Character.Health <= 0 || d.Character.Health <= 0 {
		endAggro(a)
		endAggro(d)
	} else {
		a.Character.SetAggro(d.Id, characters.DefaultAttack)
		events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
	}

	/**************************
	*
	* END HANDLING PHYSICAL COMBAT
	*
	**************************/

	return affected
}

// combatCommand may run one of an NPC's combat commands, in place of its
// attack this round. It reports whether it took the turn.
func combatCommand(mob *mobs.Mob) bool {
	cmdCt := len(mob.CombatCommands)
	if cmdCt == 0 {
		return false
	}

	if util.Rand(100) >= mob.ActivityLevel {
		return false
	}

	combatAction := mob.CombatCommands[util.Rand(cmdCt)]

	if combatAction == `` { // blank is a no-op
		return true
	}

	c := configs.GetConfig()

	var waitTime float64 = 0.0
	allCmds := strings.Split(combatAction, `;`)
	if len(allCmds) >= c.Timing.TurnsPerRound() {
		mob.Command(`say I have a CombatAction that is too long. Please notify an admin.`)
	} else {
		for _, action := range allCmds {
			mob.Command(action, waitTime)
			waitTime += 0.1
		}
	}
	return true
}

// lookForWeapon has an NPC with no weapon and a non-empty pack maybe equip a
// weapon from it (queued, so it takes effect later).
func lookForWeapon(mob *mobs.Mob) {
	if mob.Character.Equipment.Weapon.ItemId != 0 || len(mob.Character.Items) == 0 {
		return
	}

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

// offhandBreakCheck may break the defender's offhand item after a hit. A
// broken item it cannot carry lands in its own room.
func offhandBreakCheck(d *actors.Actor, defRoom *rooms.Room, crit bool) {

	// For now, only focus on offhand items.
	if d.Character.Equipment.Offhand.ItemId <= 0 {
		return
	}

	modifier := 0
	if crit { // Crits double the chance of breakage for offhand items.
		modifier = int(d.Character.Equipment.Offhand.GetSpec().BreakChance)
	}

	if !d.Character.Equipment.Offhand.BreakTest(modifier) {
		return
	}

	itemName := d.Character.Equipment.Offhand.NameSimple()

	usercommands.Tell(d, `<ansi fg="202">***</ansi>`)
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        d.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ActorText:    fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> Your <ansi fg="item">%s</ansi> breaks! <ansi fg="202">***</ansi></ansi>`, itemName),
		ObserverText: fmt.Sprintf(`<ansi fg="214"><ansi fg="202">***</ansi> The <ansi fg="item">%s</ansi> {actor} was carrying breaks! <ansi fg="202">***</ansi></ansi>`, itemName),
	})
	usercommands.Tell(d, `<ansi fg="202">***</ansi>`)

	events.AddToQueue(events.ItemOwnership{
		ActorId: d.Id,
		Item:    d.Character.Equipment.Offhand,
		Gained:  false,
	})

	d.Character.RemoveFromBody(d.Character.Equipment.Offhand)

	itm := items.New(20) // Broken item
	if !d.Character.StoreItem(itm) {
		defRoom.AddItem(itm, false)

		events.AddToQueue(events.ItemOwnership{
			ActorId: d.Id,
			Item:    itm,
			Gained:  true,
		})
	}
}

// resumeAfterFlee puts a fleer whose flee failed back in the fight it tried
// to leave (plan Decision 2), or out of combat if it fled no target.
func resumeAfterFlee(a *actors.Actor, fleeFrom ids.ID) {
	if fleeFrom.IsZero() {
		a.Character.EndAggro()
	} else {
		a.Character.SetAggro(fleeFrom, characters.DefaultAttack)
	}
	events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
}

// flee resolves a flee for any actor (plan Decisions 2 and 13). Anyone in
// the room fighting the fleer may block it; a blocked or exitless fleer
// resumes the fight it fled; a successful one leaves by a random exit.
func flee(a *actors.Actor, aRoom *rooms.Room) {

	fleeFrom := a.Character.Aggro.FleeFrom

	// Revert to Default combat regardless of outcome
	a.Character.SetAggro(a.Character.Aggro.Target, characters.DefaultAttack)

	var blocker *actors.Actor
	for _, bodyId := range aRoom.GetBodies() {
		b := actors.Get(bodyId)
		if b == nil || b.Id == a.Id || b.Character == nil {
			continue
		}
		if b.Character.Aggro == nil || b.Character.Aggro.Target != a.Id {
			continue
		}

		// Stat comparison accounts for up to 70% of chance to flee.
		chanceIn100 := int(float64(a.Character.Stats.Speed.ValueAdj) / (float64(a.Character.Stats.Speed.ValueAdj) + float64(b.Character.Stats.Speed.ValueAdj)) * 70)
		chanceIn100 += 30

		roll := util.Rand(100)

		util.LogRoll(`Flee`, roll, chanceIn100)

		if roll >= chanceIn100 {
			blocker = b
			break
		}
	}

	if blocker != nil {
		messaging.Send(perception.Message{
			Kind:         perception.WorldEvent,
			Actor:        a.Id,
			Target:       blocker.Id,
			Senses:       perception.Sight | perception.Sound,
			Range:        perception.Room,
			ActorText:    `<ansi fg="red-bold">{target} blocks you from fleeing!</ansi>`,
			TargetText:   `You block {actor} from fleeing!`,
			ObserverText: `{actor} is blocked from fleeing by {target}!`,
		})
		resumeAfterFlee(a, fleeFrom)
		return
	}

	exitName, exitRoomId := aRoom.GetRandomExit()

	if exitName == `` {
		usercommands.Tell(a, `You can't find an exit!`)
		resumeAfterFlee(a, fleeFrom)
		return
	}

	// Success!
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        a.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ActorText:    fmt.Sprintf(`You flee to the <ansi fg="exit">%s</ansi> exit!`, exitName),
		ObserverText: fmt.Sprintf(`{actor} flees to the <ansi fg="exit">%s</ansi> exit!`, exitName),
	})

	endAggro(a)

	if err := rooms.MoveToRoom(a.Id, exitRoomId); err != nil {
		return
	}

	// A charmed NPC fighting beside the fleer is in the fight too: it flees
	// on its own roll.
	for _, instId := range aRoom.GetMobs(rooms.FindCharmed) {
		if mob := mobs.GetInstance(instId); mob != nil {
			if mob.Character.IsCharmed(a.Id) {
				mob.Command(`flee`)
			}
		}
	}

	// Only a player sees where they fled to; an NPC's look prepares nothing.
	if accountOf(a) != nil {
		usercommands.Look(``, a, rooms.LoadRoom(exitRoomId), events.CmdSecretly)
	}
}

// handleAffected queues the death or the drop of everyone the round left
// below 1 health: actors with an account first, then NPCs, each once.
func handleAffected(affected []ids.ID) {

	handled := map[ids.ID]struct{}{}
	npcs := []*mobs.Mob{}

	for _, id := range affected {
		if _, ok := handled[id]; ok {
			continue
		}
		handled[id] = struct{}{}

		a := actors.Get(id)
		if a == nil {
			continue
		}

		if user := accountOf(a); user != nil {

			if user.Character.Health <= -10 {

				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

			} else if user.Character.Health < 1 {

				events.AddToQueue(events.PlayerDrop{UserId: user.UserId, RoomId: user.Character.RoomId})

			}

			continue
		}

		if mob := mobs.GetInstance(id); mob != nil {
			npcs = append(npcs, mob)
		}
	}

	for _, mob := range npcs {
		if mob.Character.Health < 1 {

			mob.Command(`suicide`)

		}
	}

}
