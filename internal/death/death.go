// Package death is the engine's one death routine (Phase 5 slice 3): every
// actor dies through Die, player or NPC, and what a death means is the
// ruleset's (rules.Ruleset.OnDeath). The engine announces the death, keeps
// the books, moves the drops and respawns or ends the body. hooks and
// usercommands import it; it imports neither.
package death

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Die is victim's death, whatever caused it: a combat round, round upkeep,
// a buff, or the suicide command.
func Die(victim *actors.Actor) {
	if victim == nil || victim.Character == nil {
		return
	}
	c := victim.Character

	outcome := rules.Get().OnDeath(victim)

	if outcome.Revive {
		revive(victim, outcome.ReviveHealth)
		return
	}

	user := accountOf(victim)
	mob := mobs.GetInstance(victim.Id)
	room := rooms.LoadRoom(c.RoomId)

	if mob != nil {
		mudlog.Debug(`Mob Death`, `name`, c.Name)
	}

	// Send a death msg to everyone in the room.
	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        victim.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ObserverText: `{actor} has died.`,
	})

	if user != nil {
		recordPlayerDeath(user)
	}
	if mob != nil {
		recordMobDeath(mob)
	}

	for _, aw := range outcome.Awards {
		award(aw, mob)
	}

	clearKillersAggro(victim)

	// The drops leave the body. With no room to put them in, they stay on it.
	var drops []items.Item
	gold := 0
	corpsesOn := bool(configs.GetGamePlayConfig().Death.CorpsesEnabled)
	if room != nil {
		drops, gold = takeDrops(c, outcome)
		placeDrops(room, mob, c.Zone, drops, gold, corpsesOn)
	}

	corpse := rooms.Corpse{Character: corpseBody(c), RoundCreated: util.GetRoundCount(), Items: drops, Gold: gold}

	if !outcome.Respawn && mob != nil {
		end(mob, room)
		if corpsesOn && room != nil {
			corpse.MobId = int(mob.MobId)
			room.AddCorpse(corpse)
		}
		return
	}

	// Every body with an account comes back: there is no permadeath
	// (decision 11), and ending a body needs an NPC instance (Q7b).
	if corpsesOn && room != nil {
		if user != nil {
			corpse.UserId = user.UserId
		} else if mob != nil {
			corpse.MobId = int(mob.MobId)
		}
		room.AddCorpse(corpse)
	}
	respawn(victim, user, outcome.RespawnHealth)
}

// accountOf is the account driving a, or nil for an NPC.
func accountOf(a *actors.Actor) *accounts.Account {
	if ctrl := a.Controller(); ctrl != nil {
		return ctrl.Account()
	}
	return nil
}

// revive gets the victim up instead of dying.
func revive(victim *actors.Actor, health int) {
	victim.Character.Health = health
	victim.Character.ForgetKillerIfUp()

	messaging.Send(perception.Message{
		Kind:         perception.WorldEvent,
		Actor:        victim.Id,
		Senses:       perception.Sight | perception.Sound,
		Range:        perception.Room,
		ObserverText: `{actor} is suddenly revived in a shower of sparks!`,
		// Only the body's presence: none for a viewer who does not perceive
		// it (Task 5a, rooms.NotPerceiving).
		Exclude: rooms.NotPerceiving(victim.Character),
	})

	victim.Character.CancelBuffsWithFlag("revive-on-death")

	if user := accountOf(victim); user != nil {
		events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
	}
}

// recordPlayerDeath keeps an account's books: the kill and death counts, the
// world notice naming who killed it, the elite death, the PlayerDeath event
// and the event log (suicide.go before slice 3, less permadeath).
func recordPlayerDeath(user *accounts.Account) {
	c := user.Character

	if len(c.PlayerDamage) > 0 {
		c.KD.AddPvpDeath()
	} else {
		c.KD.AddMobDeath()
	}

	killedByUserIds := []int{}
	killers := []*accounts.Account{}
	for uid := range c.PlayerDamage {

		if u := accounts.GetByUserId(uid); u != nil {

			// Update PK stats
			c.KD.AddPlayerDeath(u.UserId, u.Character.Name)
			u.Character.KD.AddPlayerKill(user.UserId, c.Name)
			killers = append(killers, u)
		}

		killedByUserIds = append(killedByUserIds, uid)
	}

	// The notice names, to each reader, only the killers that reader
	// perceives; with none left it names no killer at all ("someone"
	// would tell of an unseen killer). Plan review round 5, M1.
	messaging.NotifyEach(func(reader *accounts.Account) string {
		named := []string{}
		for _, k := range killers {
			if namesKiller(reader, k.Character) {
				named = append(named, `<ansi fg="username">`+k.Character.Name+`</ansi>`)
			}
		}
		return deathNotice(c.Name, named)
	})

	killerMobId := 0
	if !c.KilledBy.IsZero() {
		killerMobId = c.KillerMobId
		if c.KillerMobIsElite {
			c.KD.AddEliteDeath(killerMobId, c.KillerMobName)
		}
		c.KilledBy = ids.ID{}
		c.KillerMobIsElite = false
		c.KillerMobName = ``
		c.KillerMobId = 0
	}

	events.AddToQueue(events.PlayerDeath{
		UserId:        user.UserId,
		RoomId:        c.RoomId,
		Username:      user.Username,
		CharacterName: c.Name,
		KilledByUsers: killedByUserIds,
		KillerMobId:   killerMobId,
	})

	user.EventLog.Add(`death`, fmt.Sprintf(`<ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED</ansi>`, c.Name))
}

// namesKiller reports whether the death notice names killer to reader: the
// one visibility rule (rooms.Perceives), from where the reader is. Seeing a
// hidden body takes being in its room, unless the reader is an admin, who
// perceives hidden bodies anywhere; an unhidden killer is named to all, as
// today.
func namesKiller(reader *accounts.Account, killer *characters.Character) bool {
	if reader.Character.RoomId != killer.RoomId && killer.HasBuffFlag("hidden") {
		return reader.Role == accounts.RoleAdmin
	}
	return rooms.Perceives(reader.Character, killer)
}

// deathNotice is the world notice of name's death, naming the killers given
// ("A", "A and B", "A, B and C"); with none it names no killer.
func deathNotice(name string, killers []string) string {
	if len(killers) == 0 {
		return fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> <ansi fg="magenta-bold">***</ansi>%s`, name, term.CRLFStr)
	}
	killedBy := ``
	for i, k := range killers {
		if i > 0 {
			if i < len(killers)-1 {
				killedBy += `, `
			} else {
				killedBy += ` and `
			}
		}
		killedBy += k
	}
	return fmt.Sprintf(`<ansi fg="magenta-bold">***</ansi> <ansi fg="username">%s</ansi> has <ansi fg="red-bold">DIED!</ansi> (killed by %s) <ansi fg="magenta-bold">***</ansi>%s`, name, killedBy, term.CRLFStr)
}

// recordMobDeath queues the MobDeath event.
func recordMobDeath(mob *mobs.Mob) {
	c := &mob.Character
	killedByUsers := make([]int, 0, len(c.PlayerDamage))
	for uId := range c.PlayerDamage {
		killedByUsers = append(killedByUsers, uId)
	}

	events.AddToQueue(events.MobDeath{
		MobId:         int(mob.MobId),
		ActorId:       c.Id,
		RoomId:        c.RoomId,
		CharacterName: c.Name,
		Level:         c.Level,
		PlayerDamage:  c.PlayerDamage,
		KilledByUsers: killedByUsers,
	})
}

// award carries out one share of a kill (Q1): the kill credit when the
// victim is an NPC, the XP, then the award's lines.
func award(aw rules.Award, mob *mobs.Mob) {
	user := accounts.GetByUserId(aw.UserId)
	if user == nil {
		return
	}
	if mob != nil {
		user.Character.KD.AddMobKill(int(mob.MobId))
		if mob.IsElite {
			user.Character.KD.AddEliteKill(int(mob.MobId), mob.Character.Name)
		}
	}
	user.GrantXP(aw.XP, `combat`)
	for _, m := range aw.Messages {
		messaging.Send(m)
	}
}

// clearKillersAggro ends the fight of every account credited with damage to
// the victim whose body still targets it.
func clearKillersAggro(victim *actors.Actor) {
	for uId := range victim.Character.PlayerDamage {
		u := accounts.GetByUserId(uId)
		if u == nil || u.Character.Aggro == nil || u.Character.Aggro.Target != victim.Id {
			continue
		}
		u.Character.Aggro = nil
		events.AddToQueue(events.AggroChanged{ActorId: u.Character.Id, RoomId: u.Character.RoomId})
	}
}

// takeDrops removes the outcome's drops from the body and returns what left
// it: each item carried or worn (an item the body does not hold is not
// dropped), and the gold, up to what the body has. Each item that leaves
// queues ItemOwnership (lost), as every other item move does (spec, plan
// review round 1), and gold that leaves queues EquipmentChange.
func takeDrops(c *characters.Character, outcome rules.DeathOutcome) ([]items.Item, int) {
	drops := make([]items.Item, 0, len(outcome.DropItems))
	for _, itm := range outcome.DropItems {
		if c.RemoveItem(itm) || c.RemoveFromBody(itm) {
			drops = append(drops, itm)
			events.AddToQueue(events.ItemOwnership{ActorId: c.Id, Item: itm, Gained: false})
		}
	}
	gold := min(max(outcome.DropGold, 0), c.Gold)
	c.Gold -= gold
	if gold > 0 {
		// The client's worth refreshes on it, as for drop (drop.go).
		events.AddToQueue(events.EquipmentChange{ActorId: c.Id, GoldChange: -gold})
	}
	return drops, gold
}

// placeDrops puts the drops on the floor when corpses are off, with the
// lines an NPC's floor drops have always had (Q7d); otherwise the corpse
// holds them. MobItemDrop is queued for each NPC drop either way (Q7e).
func placeDrops(room *rooms.Room, mob *mobs.Mob, zone string, drops []items.Item, gold int, corpsesOn bool) {
	for _, itm := range drops {
		if !corpsesOn {
			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Room:         room.RoomId,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ObserverText: fmt.Sprintf(`<ansi fg="item">%s</ansi> drops to the ground.`, itm.DisplayName()),
			})
		}
		if mob != nil {
			events.AddToQueue(events.MobItemDrop{
				MobId:  int(mob.MobId),
				RoomId: room.RoomId,
				Zone:   zone,
				ItemId: itm.ItemId,
			})
		}
		if !corpsesOn {
			room.AddItem(itm, false)
		}
	}

	if !corpsesOn && gold > 0 {
		messaging.Send(perception.Message{
			Kind:         perception.WorldEvent,
			Room:         room.RoomId,
			Senses:       perception.Sight | perception.Sound,
			Range:        perception.Room,
			ObserverText: fmt.Sprintf(`<ansi fg="yellow-bold">%d gold</ansi> drops to the ground.`, gold),
		})
		room.Gold += gold
	}
}

// corpseBody is the dead body for its corpse, carrying, wearing and holding
// nothing: the corpse holds the drops (Corpse.Items, Corpse.Gold), so get
// and look find nothing the dead kept (Q2).
func corpseBody(c *characters.Character) characters.Character {
	body := *c
	body.Items = nil
	body.Equipment = characters.Worn{}
	body.Gold = 0
	return body
}

// end is an NPC's ending (decision 15): its slot waits its respawn rate.
func end(mob *mobs.Mob, room *rooms.Room) {
	rooms.NPCEnded(mob, false)

	// Destroy any record of this mob.
	mobs.DestroyInstance(mob.Character.Id)

	// Clean up mob from room...
	if r := rooms.LoadRoom(mob.HomeRoomId); r != nil {
		r.CleanupMobSpawns()
	}

	if room != nil {
		room.RemoveMob(mob.Character.Id)
	}
}

// respawn brings the body back in the death recovery room (its own private
// copy, MoveToRoom) with no buffs, the given health and no mana (Q7c), its
// damage record cleared and its fight ended. recordPlayerDeath has already cleared the killer
// record (only an account has one: NewRound_DoCombat records it for a
// defender with an account).
func respawn(victim *actors.Actor, user *accounts.Account, health int) {
	c := victim.Character

	c.CancelBuffsWithFlag(buffs.All)

	c.Health = health
	c.Mana = 0
	if user != nil {
		events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})
	}

	clear(c.PlayerDamage)

	// The body's own fight ends with it: a death in upkeep or from a buff
	// runs no endAggro, and a kept target would leave it "too busy" to quit.
	// The room it died in is told, as endAggro tells it.
	if c.Aggro != nil {
		c.EndAggro()
		events.AddToQueue(events.AggroChanged{ActorId: c.Id, RoomId: c.RoomId})
	}

	if err := rooms.MoveToRoom(victim.Id, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)); err != nil {
		mudlog.Error("death", "error", err)
	}
}
