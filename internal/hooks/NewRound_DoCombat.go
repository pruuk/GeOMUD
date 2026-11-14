package hooks

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// DoCombat runs one combat round: one turn for every actor with an aggro,
// player or NPC, by the same rules, then the death checks for everyone the
// round touched.
func DoCombat(e events.Event) events.ListenerReturn {

	tStart := time.Now()

	var affected []ids.ID
	for _, a := range rules.Get().TurnOrder(fighters()) {
		affected = append(affected, combatTurn(a)...)
	}

	// Do any resolution or extra checks based on everyone that has been involved in combat this round.
	handleAffected(affected)

	util.TrackTime(`DoCombat()`, time.Since(tStart).Seconds())

	return events.Continue
}

// fighters lists every actor with a body that may take a combat turn, in
// death.SortBodies order (players by UserId, then NPCs by MobId, then body
// ID). The ruleset puts them in turn order.
func fighters() []*actors.Actor {
	out := []*actors.Actor{}
	for _, userId := range accounts.GetOnlineUserIds() {
		if a := actors.ByUserId(userId); a != nil {
			out = append(out, a)
		}
	}
	for _, mobId := range mobs.GetAllMobIds() {
		if a := actors.Get(mobId); a != nil {
			out = append(out, a)
		}
	}
	return death.SortBodies(out)
}

// accountOf is the account driving a, or nil for an NPC.
func accountOf(a *actors.Actor) *accounts.Account {
	if ctrl := a.Controller(); ctrl != nil {
		return ctrl.Account()
	}
	return nil
}

// attack runs one round of a's attacks on d: the ruleset works it out, the
// engine applies the damage, credits a's account on every attack, hit or
// miss, and runs the defender's wimpy check when it took damage.
func attack(a, d *actors.Actor) rules.AttackResult {
	result := rules.Get().ResolveAttack(a, d)

	d.Character.ApplyHealthChange(result.DamageToTarget * -1)

	// Remember who has hit him
	if acct := accountOf(a); acct != nil {
		d.Character.TrackPlayerDamage(acct.UserId, result.DamageToTarget)
	}

	if result.DamageToTarget != 0 {
		if acct := accountOf(d); acct != nil {
			acct.WimpyCheck()
		}
	}

	return result
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

	aRoom := rooms.LoadRoom(a.Character.RoomId)
	if aRoom == nil {
		endAggro(a)
		return nil
	}

	// A buff that prevents combat takes the turn (the swing below, or a flee).
	noCombat := a.Character.HasBuffFlag("no-combat")

	if a.Character.Aggro.Type == characters.Flee {
		if noCombat {
			return nil
		}
		a.Character.CancelBuffsWithFlag("cancel-on-combat")
		flee(a, aRoom)
		return nil
	}

	/**************************
	*
	* START HANDLING PHYSICAL COMBAT
	*
	**************************/

	// Every body with an aggro settles its target every round, before
	// anything may take its turn (no-combat, an NPC's combat command): the
	// lookup runs, a switch or a lost track happens, the wait count advances
	// (Phase 5 slice 4.1, ruling 12). Only the swing is skipped.
	d := settleTarget(a, aRoom)
	if d == nil {
		return []ids.ID{a.Id}
	}

	if noCombat {
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

	affected = append(affected, a.Id)

	defRoom := rooms.LoadRoom(d.Character.RoomId)
	if defRoom == nil {
		endAggro(a)
		return affected
	}

	d.Character.CancelBuffsWithFlag("cancel-on-combat")

	affected = append(affected, d.Id)

	// An NPC with no weapon but something in its pack may look for a weapon
	// (especially useful when disarmed).
	if mob := mobs.GetInstance(a.Id); mob != nil {
		lookForWeapon(a, mob)
	}

	if a.Character.Aggro.RoundsWaiting > 0 {
		mudlog.Debug(`RoundsWaiting`, `Actor`, a.Character.Name, `Rounds`, a.Character.Aggro.RoundsWaiting)

		a.Character.Aggro.RoundsWaiting--

		for _, m := range rules.Get().WaitLines(a, d) {
			messaging.Send(m)
		}

		return affected
	}

	// Can't see them, can't fight them (the one visibility rule).
	if !rooms.Perceives(a.Character, d.Character) {
		usercommands.Tell(a, "You can't seem to find your target.")
		return affected
	}

	roundResult := attack(a, d)

	for _, buffId := range roundResult.BuffTarget {
		events.AddToQueue(events.Buff{ActorId: d.Id, BuffId: buffId, Source: `combat`})
	}

	for _, m := range roundResult.Messages {
		messaging.Send(m)
	}

	defAcct := accountOf(d)

	// Combat took a player out of Up: remember who did it (plan Decision 7),
	// for the death that follows. The defender was up before this attack:
	// the rage check above returns for one who is not.
	if defAcct != nil && rules.Get().Condition(d) != rules.Up {
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
	// account. How long is the ruleset's.
	if defMob != nil {
		if acct := accountOf(a); acct != nil {
			hatred := rules.Get().HatredRounds(a)
			for _, groupName := range defMob.Groups {
				mobs.MakeHostile(groupName, acct.UserId, hatred)
			}
		}
	}

	if defAcct != nil && roundResult.DamageToTarget != 0 {
		events.AddToQueue(events.CharacterVitalsChanged{UserId: defAcct.UserId})
	}

	// An NPC that is attacked and not fighting stays put; attacked from
	// another room, it heads that way (and retaliates below).
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

	// A target in another room was shot through an exit (lookup step 3).
	// Only a typed shot earns the target one shot back (ruling 13).
	shot := d.Character.RoomId != a.Character.RoomId
	typedShot := shot && a.Character.Aggro.Shot == characters.TypedShot

	// A defender whose aggro targets its attacker is still in that fight:
	// the attack, hit or miss, resets its wait count (plan review round 6),
	// and a typed shot earns it one shot back, never more than one.
	if d.Character.Aggro != nil && d.Character.Aggro.Target == a.Id {
		d.Character.Aggro.OutOfReach = 0
		if typedShot && d.Character.Aggro.Shot == characters.NoShot {
			d.Character.Aggro.Shot = characters.EarnedShot
		}
	}
	retaliate(d, a, typedShot)

	// A defender taken out of Up leaves its own fight. The attacker keeps
	// its aggro: the death that follows (death.Die) turns it to its next
	// opponent, or ends it; under a ruleset with a downed state its next
	// turn does (settleTarget).
	if rules.Get().Condition(d) != rules.Up {
		endAggro(d)
	} else if shot {
		// The shot is spent (ruling 13). The re-arm keeps the exit it shot
		// through, so the aggro waits on the target (lookup step 5), with
		// the told mark set: the body just shot the target there, so it is
		// not told "Your target is out of reach.".
		a.Character.SetAggroRemote(a.Character.Aggro.ExitName, d.Id, characters.Shooting)
		a.Character.Aggro.OutOfReachTold = true
		events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
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

// settleTarget finds the target a swings at this turn (Phase 5 slice 4.1,
// rulings 7-9), or returns nil when the turn ends here:
//  1. the target no longer exists: switch to the next opponent, or end
//     ("Your target can't be found.");
//  2. the target is not up: switch, or end ("Your rage subsides.");
//  3. the target is in another room, the aggro holds a shot (ruling 13), a
//     wields a shooting weapon and an unlocked exit leads there: shoot
//     through that exit, once (a target a is about to shoot is reachable,
//     so it is kept even if someone here attacks a). A shot that cannot be
//     fired is dropped, and the lookup goes on to step 4;
//  4. the target is in another room (its shot spent or never held), or
//     here but a cannot perceive it, and someone here whom a perceives
//     attacks a: switch;
//  5. otherwise the aggro waits, telling a once "Your target is out of
//     reach."; a swings again once they share a room. After
//     GamePlay.OutOfReachRounds such turns in a row (5 by default) a loses
//     track of the target and the aggro ends (ruling 12). A target in reach
//     that a cannot perceive counts the same way (inReach). Being attacked
//     by the target resets the count but not the told mark, so the line
//     comes once per wait.
//
// A target it would attack (in the room, or through an exit) must still be
// one the ruleset lets it attack (rules.CanAttack: the fight may have moved
// into a room where PvP is off); if not, it switches as in step 4 or ends,
// told the refusal. The switch picks only opponents the ruleset lets it
// attack (death.NextOpponent). A switch re-arms the aggro through SetAggro,
// as attack <name> does, and the turn goes on to the new opponent: it swings
// this turn, unless its weapon's delay makes it wait first.
func settleTarget(a *actors.Actor, aRoom *rooms.Room) *actors.Actor {
	d := actors.Get(a.Character.Aggro.Target)
	if d == nil || d.Character == nil {
		return switchOrEnd(a, `Your target can't be found.`)
	}
	if rules.Get().Condition(d) != rules.Up {
		return switchOrEnd(a, `Your rage subsides.`)
	}
	if d.Character.RoomId == a.Character.RoomId {
		if err := rules.Get().CanAttack(a, d); err != nil {
			return switchOrEnd(a, err.Error())
		}
		// Step 4 for a target here that a cannot perceive: a perceived
		// attacker in the room comes first.
		if !rooms.Perceives(a.Character, d.Character) && death.Switch(a) {
			return actors.Get(a.Character.Aggro.Target)
		}
		return inReach(a, d)
	}
	if a.Character.Aggro.Shot != characters.NoShot {
		if exitName := shootingExit(a, aRoom, d.Character.RoomId); exitName != `` {
			if err := rules.Get().CanAttack(a, d); err != nil {
				return switchOrEnd(a, err.Error())
			}
			if a.Character.Aggro.Type != characters.Shooting || a.Character.Aggro.ExitName != exitName {
				a.Character.Aggro.Type = characters.Shooting
				a.Character.Aggro.ExitName = exitName
				events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
			}
			return inReach(a, d)
		}
		// A shot that cannot be fired (no shooting weapon, no unlocked
		// exit) is dropped (ruling 13).
		a.Character.Aggro.Shot = characters.NoShot
	}
	if death.NextOpponent(a) != nil {
		death.Switch(a)
		return actors.Get(a.Character.Aggro.Target)
	}
	a.Character.Aggro.OutOfReach++
	if !a.Character.Aggro.OutOfReachTold {
		a.Character.Aggro.OutOfReachTold = true
		usercommands.Tell(a, `Your target is out of reach.`)
	}
	if a.Character.Aggro.OutOfReach >= outOfReachRounds() {
		loseTrack(a, d)
	}
	return nil
}

// inReach settles a target a can reach (in its room, or through an exit).
// One a perceives resets the wait count and the told mark (OutOfReachTold). One a cannot perceive counts
// toward the wait limit like a target out of reach (ruling 12; the turn
// then says "You can't seem to find your target."), and a loses track of it
// at the limit.
func inReach(a, d *actors.Actor) *actors.Actor {
	if rooms.Perceives(a.Character, d.Character) {
		a.Character.Aggro.OutOfReach = 0
		a.Character.Aggro.OutOfReachTold = false
		return d
	}
	a.Character.Aggro.OutOfReach++
	if a.Character.Aggro.OutOfReach >= outOfReachRounds() {
		loseTrack(a, d)
		return nil
	}
	return d
}

// outOfReachRounds is how many consecutive turns an aggro may wait on a
// target out of reach (ruling 12): GamePlay.OutOfReachRounds, which
// GamePlay.Validate makes 5 when it is 0 or less (GetGamePlayConfig always
// returns a validated config).
func outOfReachRounds() int {
	return int(configs.GetGamePlayConfig().OutOfReachRounds)
}

// loseTrack ends a's wait on d (ruling 12): a is told it lost track of d,
// and those in its room who perceive it see it give up.
func loseTrack(a, d *actors.Actor) {
	messaging.Send(perception.Message{
		Kind:      perception.WorldEvent,
		Actor:     a.Id,
		Target:    d.Id,
		Audience:  perception.ActorOnly,
		ActorText: `You lose track of {target}.`,
	})
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        a.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ObserverText: `{actor} mumbles about losing their quarry.`,
		Exclude:      rooms.NotPerceiving(a.Character),
	})
	endAggro(a)
}

// switchOrEnd turns a to its next opponent and returns it, or ends a's
// aggro with line and returns nil.
func switchOrEnd(a *actors.Actor, line string) *actors.Actor {
	if death.Switch(a) {
		return actors.Get(a.Character.Aggro.Target)
	}
	usercommands.Tell(a, line)
	endAggro(a)
	return nil
}

// shootingExit is the exit from aRoom, unlocked and leading to roomId, that a
// can shoot through: none unless a wields a shooting weapon. Exits are looked
// up as the shoot command does (room.FindExitByName for where an exit leads,
// room.GetExitInfo for its lock), so temporary and mutator exits and mutator
// locks count. The aggro's own exit is preferred; otherwise the first by
// name of the room's exits, temporary exits and active mutators' exits.
func shootingExit(a *actors.Actor, aRoom *rooms.Room, roomId int) string {
	if a.Character.Equipment.Weapon.GetSpec().Subtype != items.Shooting {
		return ``
	}
	names := []string{}
	if a.Character.Aggro != nil && a.Character.Aggro.ExitName != `` {
		names = append(names, a.Character.Aggro.ExitName)
	}
	all := []string{}
	for name := range aRoom.Exits {
		all = append(all, name)
	}
	for name := range aRoom.ExitsTemp {
		all = append(all, name)
	}
	for mut := range aRoom.ActiveMutators {
		if spec := mut.GetSpec(); spec != nil {
			for name := range spec.Exits {
				all = append(all, name)
			}
		}
	}
	sort.Strings(all)
	names = append(names, all...)
	for _, name := range names {
		exitName, toRoom := aRoom.FindExitByName(name)
		if toRoom != roomId {
			continue
		}
		if exitInfo, ok := aRoom.GetExitInfo(exitName); ok && exitInfo.Lock.IsLocked() {
			continue
		}
		return exitName // the name shoot <direction> would store
	}
	return ``
}

// retaliate gives a defender that is not fighting an aggro on its attacker
// (Phase 5 slice 4.1, rulings 6, 8 and 9), after every attack, hit or miss,
// when it is up and the ruleset lets it attack back: a plain attack when they
// share a room; when the attacker is in another room and its shot was typed
// (typedShot) and the defender can shoot there, an aggro through the exit
// holding one shot back (ruling 13); otherwise an aggro that waits. An
// attacker the defender cannot perceive is taken all the same (its turn then
// says "You can't seem to find your target.").
func retaliate(d, a *actors.Actor, typedShot bool) {
	if d.Character.Aggro != nil || rules.Get().Condition(d) != rules.Up {
		return
	}
	if rules.Get().CanAttack(d, a) != nil {
		return
	}
	if a.Character.RoomId == d.Character.RoomId {
		d.Character.SetAggro(a.Id, characters.DefaultAttack)
	} else if dRoom := rooms.LoadRoom(d.Character.RoomId); typedShot && dRoom != nil && shootingExit(d, dRoom, a.Character.RoomId) != `` {
		d.Character.SetAggroRemote(shootingExit(d, dRoom, a.Character.RoomId), a.Id, characters.Shooting)
		d.Character.Aggro.Shot = characters.EarnedShot
	} else {
		d.Character.SetAggro(a.Id, characters.DefaultAttack)
	}
	events.AddToQueue(events.AggroChanged{ActorId: d.Id, RoomId: d.Character.RoomId})
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
func lookForWeapon(a *actors.Actor, mob *mobs.Mob) {
	if mob.Character.Equipment.Weapon.ItemId != 0 || len(mob.Character.Items) == 0 {
		return
	}

	if rules.Get().FindsWeapon(a) {
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

// resumeAfterFlee puts a fleer whose flee failed back in the fight it tried
// to leave (plan Decision 2), or out of combat if it fled no target.
func resumeAfterFlee(a *actors.Actor, fleeFrom ids.ID, since uint64) {
	if fleeFrom.IsZero() {
		a.Character.EndAggro()
	} else {
		a.Character.SetAggro(fleeFrom, characters.DefaultAttack)
		a.Character.Aggro.Since = since // the fight resumes in its place
	}
	events.AddToQueue(events.AggroChanged{ActorId: a.Id, RoomId: a.Character.RoomId})
}

// flee resolves a flee for any actor (plan Decisions 2 and 13). Anyone in
// the room fighting the fleer may block it; a blocked or exitless fleer
// resumes the fight it fled; a successful one leaves by a random exit.
func flee(a *actors.Actor, aRoom *rooms.Room) {

	fleeFrom := a.Character.Aggro.FleeFrom
	since := a.Character.Aggro.Since

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

		if !rules.Get().FleeEscapes(a, b) {
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
		resumeAfterFlee(a, fleeFrom, since)
		return
	}

	exitName, exitRoomId := aRoom.GetRandomExit()

	if exitName == `` {
		usercommands.Tell(a, `You can't find an exit!`)
		resumeAfterFlee(a, fleeFrom, since)
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
		Exclude:      rooms.NotPerceiving(a.Character), // a leaving: only its presence
	})

	endAggro(a)

	if err := rooms.MoveToRoom(a.Id, exitRoomId); err != nil {
		return
	}

	// Only a player sees where they fled to; an NPC's look prepares nothing.
	if accountOf(a) != nil {
		usercommands.Look(``, a, rooms.LoadRoom(exitRoomId), events.CmdSecretly)
	}
}

// handleAffected settles everyone the round touched, actors with an account
// first, then NPCs, each once: the ruleset says who is dead (death.Die) or,
// under a ruleset with a downed state, down (an account's PlayerDrop).
func handleAffected(affected []ids.ID) {

	handled := map[ids.ID]struct{}{}
	npcs := []*actors.Actor{}

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
			switch rules.Get().Condition(a) {
			case rules.Dead:
				death.Die(a)
			case rules.Downed:
				events.AddToQueue(events.PlayerDrop{UserId: user.UserId, RoomId: user.Character.RoomId})
			}
			continue
		}

		if mobs.GetInstance(id) != nil {
			npcs = append(npcs, a)
		}
	}

	for _, a := range npcs {
		if rules.Get().Condition(a) == rules.Dead {
			death.Die(a)
		}
	}

}
