package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/require"
)

// taken splits what the recorder holds into the event notes and each
// player's reads.
func (w *combatWorld) taken() (notes []string, reads map[int][]string) {
	reads = map[int][]string{}
	for _, e := range w.rec.take() {
		if e.userId == 0 {
			notes = append(notes, e.text)
			continue
		}
		reads[e.userId] = append(reads[e.userId], e.text)
	}
	return notes, reads
}

// flag gives c a test buff carrying one flag and keeps its health (AddBuff
// runs Validate, which would reset it, as in hide).
func (w *combatWorld) flag(c *characters.Character, buffId int, flag string) {
	w.t.Helper()
	w.t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000, Flags: []string{flag},
	}))
	hpMax, hp := c.HealthMax.Value, c.Health
	require.NoError(w.t, c.AddBuff(buffId, true))
	c.HealthMax.Value, c.Health = hpMax, hp
	require.True(w.t, c.HasBuffFlag(flag))
}

// Round upkeep runs the death of any actor the ruleset calls dead (Q7a): an
// NPC ends, a player respawns.
func TestUpkeepRunsTheDeathOfWhoeverTheRulesetCallsDead(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "upkeep-rats")
	w.start()
	alice.Character.Health, rat.Character.Health = 0, 0

	ActorUpkeep(events.NewRound{RoundNumber: 1})
	events.ProcessEvents()

	notes, _ := w.taken()
	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the rat ended")
	require.Contains(t, notes, "MobDeath rat")
	require.Contains(t, notes, "PlayerDeath Alice")
	require.Equal(t, goldenRoomC, rooms.GetOriginalRoom(alice.Character.RoomId), "Alice is in a copy of the recovery room")
	require.Equal(t, 1, alice.Character.Health)
}

// An immediate buff trigger runs the death of a body the ruleset calls dead
// (Q7a). The death ends the fight of every credited account still targeting
// the body (clearKillersAggro), credits the kill and grants the XP (award).
func TestBuffTriggerDeathEndsTheKillersFightAndPaysTheKill(t *testing.T) {
	const triggerBuff = 93902
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: a level up saves the account
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "buff-rats")
	rat.Character.Level = 5 // enough XP that the grant shows
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	rat.Character.PlayerDamage = map[int]int{alice.UserId: 12}
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: triggerBuff, Name: "test trigger", TriggerNow: true, TriggerRate: "1 round", TriggerCount: 1,
	}))
	w.start()
	xpBefore := alice.Character.Experience
	rat.Character.Health = 0

	ApplyBuffs(events.Buff{ActorId: rat.Character.Id, BuffId: triggerBuff})
	events.ProcessEvents()

	notes, reads := w.taken()
	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the rat died")
	require.Contains(t, notes, "MobDeath rat")
	require.Nil(t, alice.Character.Aggro, "her fight with the rat ended")
	require.Contains(t, notes, "AggroChanged Alice room=93001")
	require.Equal(t, 1, alice.Character.KD.GetMobKills(93201), "the kill is credited")
	require.Greater(t, alice.Character.Experience, xpBefore, "the XP is granted")
	require.Contains(t, strings.Join(reads[alice.UserId], "\n"), "experience points")
}

// With corpses off the drops go on the floor with the floor lines (Q7d), and
// nothing is lost.
func TestDeathWithCorpsesOffDropsOnTheFloor(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	setConfig(t,
		map[string]any{"GamePlay.Death.CorpsesEnabled": false},
		map[string]any{"GamePlay.Death.CorpsesEnabled": bool(configs.GetGamePlayConfig().Death.CorpsesEnabled)})
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	alice.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	require.Equal(t, 25, alice.Character.Gold, "a new character's gold")
	w.start()

	death.Die(actors.Get(alice.Character.Id))
	events.ProcessEvents()

	notes, reads := w.taken()
	require.Contains(t, notes, "ItemOwnership Alice item=93108 gained=false", "the gem left her (spec, plan review round 1)")
	room := rooms.LoadRoom(goldenRoomA)
	require.Equal(t, "93108", itemIds(room.Items), "the gem is on the floor")
	require.Equal(t, 25, room.Gold, "her gold is on the floor")
	require.Empty(t, room.Corpses)
	require.Empty(t, alice.Character.Items)
	require.Zero(t, alice.Character.Gold)
	bobRead := strings.Join(reads[bob.UserId], "\n")
	require.Contains(t, bobRead, "testgem")
	require.Contains(t, bobRead, `<ansi fg="yellow-bold">25 gold</ansi> drops to the ground.`)
}

// The gold that leaves a body queues EquipmentChange, as drop does, so the
// client's worth refreshes (whole-branch review m3). A body with no gold to
// drop queues none.
func TestDeathGoldDropQueuesEquipmentChange(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	require.Equal(t, 25, alice.Character.Gold, "a new character's gold")
	bob.Character.Gold = 0
	w.start()
	var changes []events.EquipmentChange
	id := events.RegisterListener(events.EquipmentChange{}, func(e events.Event) events.ListenerReturn {
		changes = append(changes, e.(events.EquipmentChange))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.EquipmentChange{}, id) })

	death.Die(actors.Get(alice.Character.Id))
	death.Die(actors.Get(bob.Character.Id))
	events.ProcessEvents()

	require.Equal(t, []events.EquipmentChange{{ActorId: alice.Character.Id, GoldChange: -25}}, changes)
}

// revive-on-death gets a body up once: the flag is used up, so the next
// death is a death; an account's vitals are reported (Q7f).
func TestReviveOnDeathWorksOnce(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	w.flag(alice.Character, 93903, "revive-on-death")
	w.start()
	a := actors.Get(alice.Character.Id)

	alice.Character.Health = 0
	death.Die(a)
	events.ProcessEvents()
	notes, _ := w.taken()
	require.Positive(t, alice.Character.Health, "revived")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "not moved")
	require.False(t, alice.Character.HasBuffFlag("revive-on-death"), "the flag is used up")
	require.Contains(t, notes, "CharacterVitalsChanged Alice")
	require.NotContains(t, notes, "PlayerDeath Alice")

	alice.Character.Health = 0
	death.Die(a)
	events.ProcessEvents()
	notes, _ = w.taken()
	require.Contains(t, notes, "PlayerDeath Alice", "the second death is a death")
}

// A revive forgets who downed the body (whole-branch review m2), as every
// other health raise does (ForgetKillerIfUp), so a later death is not
// credited to a stale killer.
func TestReviveForgetsTheKiller(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	w.flag(alice.Character, 93903, "revive-on-death")
	w.start()
	c := alice.Character
	c.KilledBy, c.KillerMobId, c.KillerMobName, c.KillerMobIsElite = guard.Character.Id, 93202, "guard", true

	c.Health = 0
	death.Die(actors.Get(c.Id))
	events.ProcessEvents()

	require.Positive(t, c.Health, "revived")
	require.True(t, c.KilledBy.IsZero(), "the killer is forgotten")
	require.Zero(t, c.KillerMobId)
	require.Empty(t, c.KillerMobName)
	require.False(t, c.KillerMobIsElite)
}

// The revive line is only the body's presence (Task 5a's rule, plan review
// round 5): a viewer who does not perceive the hidden body reads nothing,
// not even "someone"; an admin reads it.
func TestReviveOfAHiddenBodyReachesOnlyThoseWhoPerceive(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	eve := w.addPlayer(9303, "Eve", goldenRoomA, 10, 10, 10, 12)
	eve.Role = accounts.RoleAdmin
	w.flag(alice.Character, 93903, "revive-on-death")
	w.hide(alice.Character)
	w.start()

	alice.Character.Health = 0
	death.Die(actors.Get(alice.Character.Id))
	events.ProcessEvents()
	_, reads := w.taken()
	require.Positive(t, alice.Character.Health, "revived")
	require.Contains(t, strings.Join(reads[eve.UserId], "\n"), "is suddenly revived in a shower of sparks!", "an admin perceives the hidden body")
	require.NotContains(t, strings.Join(reads[bob.UserId], "\n"), "is suddenly revived", "Bob does not perceive Alice: no line, not even as someone")
}

// The world death notice names, to each reader, only the killers that reader
// perceives (plan review round 5, M1): a hidden killer (Alice) only to an
// admin anywhere (Eve), a see-hidden reader in her room (Dan) and herself;
// never "someone". An unhidden killer (Gus, in room B) is named to all.
func TestDeathNoticeNamesOnlyTheKillersEachReaderPerceives(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 12)
	carol := w.addPlayer(9303, "Carol", goldenRoomB, 10, 10, 10, 12)
	dan := w.addPlayer(9304, "Dan", goldenRoomA, 10, 10, 10, 12)
	eve := w.addPlayer(9305, "Eve", goldenRoomB, 10, 10, 10, 12)
	fay := w.addPlayer(9306, "Fay", goldenRoomB, 10, 10, 10, 12)
	gus := w.addPlayer(9307, "Gus", goldenRoomB, 10, 10, 10, 12)
	eve.Role = accounts.RoleAdmin
	w.flag(dan.Character, 93905, "see-hidden")
	w.flag(fay.Character, 93906, "see-hidden")
	w.hide(alice.Character)
	w.start()

	die := func(u *accounts.Account, damage map[int]int) map[int][]string {
		u.Character.PlayerDamage = damage
		u.Character.Health = 0
		death.Die(actors.Get(u.Character.Id))
		events.ProcessEvents()
		_, reads := w.taken()
		return reads
	}
	notice := func(reads map[int][]string, u *accounts.Account) string {
		for _, r := range reads[u.UserId] {
			if strings.Contains(r, "DIED!") {
				return r
			}
		}
		require.Failf(t, "no notice", "%s read no death notice", u.Character.Name)
		return ""
	}
	aliceName, gusName := `<ansi fg="username">Alice</ansi>`, `<ansi fg="username">Gus</ansi>`

	reads := die(bob, map[int]int{alice.UserId: 5, gus.UserId: 3})
	for _, u := range []*accounts.Account{alice, dan, eve} {
		n := notice(reads, u)
		require.Contains(t, n, aliceName, "%s perceives the hidden Alice", u.Character.Name)
		require.Contains(t, n, gusName, u.Character.Name)
	}
	for _, u := range []*accounts.Account{bob, carol, fay, gus} {
		n := notice(reads, u)
		require.NotContains(t, n, aliceName, "%s does not perceive the hidden Alice", u.Character.Name)
		require.Contains(t, n, "(killed by "+gusName+")", "%s: only the killer it perceives", u.Character.Name)
		require.NotContains(t, strings.ToLower(n), "someone", u.Character.Name)
	}

	// With the only killer hidden, a reader who does not perceive her reads
	// no "(killed by ...)" at all; an admin reads her name.
	reads = die(gus, map[int]int{alice.UserId: 5})
	require.NotContains(t, notice(reads, carol), "(killed by", "nothing to name, so no killed by")
	require.Contains(t, notice(reads, eve), "(killed by "+aliceName+")")
}

// A respawn starts clean: no buffs, no mana (Q7c), no damage record, no
// killer record; an elite killer is booked as an elite death first. The
// PlayerDeath event names who killed her, and her event log records it.
func TestRespawnStartsClean(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	w.flag(alice.Character, 93904, "testmark") // any flagged buff: buffs.All cancels flagged buffs
	w.start()
	var deaths []events.PlayerDeath
	id := events.RegisterListener(events.PlayerDeath{}, func(e events.Event) events.ListenerReturn {
		deaths = append(deaths, e.(events.PlayerDeath))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.PlayerDeath{}, id) })
	c := alice.Character
	require.Positive(t, c.ManaMax.Value, "the fixture needs room for mana")
	c.Health, c.Mana = 0, c.ManaMax.Value
	c.PlayerDamage = map[int]int{9302: 3}
	c.KilledBy, c.KillerMobId, c.KillerMobName, c.KillerMobIsElite = guard.Character.Id, 93202, "guard", true

	death.Die(actors.Get(c.Id))
	events.ProcessEvents()

	require.Equal(t, goldenRoomC, rooms.GetOriginalRoom(c.RoomId))
	require.Equal(t, 1, c.Health)
	require.Zero(t, c.Mana, "no mana (Q7c)")
	require.False(t, c.HasBuffFlag("testmark"), "every buff is cancelled")
	require.Empty(t, c.PlayerDamage)
	require.True(t, c.KilledBy.IsZero())
	require.Zero(t, c.KillerMobId)
	require.Equal(t, 1, c.KD.EliteDeaths["93202:guard"], "an elite killer's death is booked")

	require.Len(t, deaths, 1)
	require.Equal(t, 93202, deaths[0].KillerMobId, "the killer NPC")
	require.Equal(t, []int{9302}, deaths[0].KilledByUsers, "the accounts credited with damage")
	require.Equal(t, goldenRoomA, deaths[0].RoomId, "where she died, not where she woke")
	require.NotEmpty(t, alice.EventLog)
	last := alice.EventLog[len(alice.EventLog)-1]
	require.Equal(t, "death", last.Category)
	require.Contains(t, last.What, "DIED", "the event log records the death")
}

// A respawn ends the body's own fight (whole-branch review m1): a death in
// upkeep or from a buff runs no endAggro, and a body that kept its target
// would be "too busy" to quit after waking. The room it died in hears of it.
func TestRespawnEndsTheFight(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "respawn-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	alice.Character.Health = 0

	ActorUpkeep(events.NewRound{RoundNumber: 1})
	events.ProcessEvents()

	notes, _ := w.taken()
	require.Contains(t, notes, "PlayerDeath Alice")
	require.Nil(t, alice.Character.Aggro, "her fight ended with her death")
	require.Contains(t, notes, "AggroChanged Alice room=93001")
}

// An NPC's ending frees its spawn slot in its home room (CleanupMobSpawns),
// so the slot can spawn again.
func TestDeathFreesTheNPCsSpawnSlot(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: the ending ledger and the room's NPC file are written there
	dataFolder := configs.GetFilePathsConfig().DataFiles.String()
	t.Cleanup(func() { // registered after recoveryRoom's cleanups, so it runs first, in the temporary folder
		require.NoError(t, os.RemoveAll(filepath.Join(dataFolder, "rooms.instances")))
		require.NoError(t, rooms.LoadEndings())
	})
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "slot-rats")
	rat.SpawnSlot = 1                   // a slot NPC: rooms.NPCEnded records its ending
	home := rooms.LoadRoom(goldenRoomA) // addNPC makes the room it is put in its home (HomeRoomId)
	home.SpawnInfo = []rooms.SpawnInfo{{MobId: 93201, InstanceId: rat.Character.Id}}
	w.start()

	death.Die(actors.Get(rat.Character.Id))
	events.ProcessEvents()

	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the rat ended")
	require.True(t, home.SpawnInfo[0].InstanceId.IsZero(), "its slot is free (CleanupMobSpawns)")
}
