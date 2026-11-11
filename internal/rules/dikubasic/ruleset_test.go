package dikubasic

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init() // NPC instances get actors through its hooks
	os.Exit(m.Run())
}

// npc is a test NPC actor with the given Perception and Speed.
func npc(t *testing.T, perception, speed int) *actors.Actor {
	t.Helper()
	m := mobs.NewTestInstance(mobs.Mob{})
	id := m.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	a := actors.Get(id)
	require.NotNil(t, a, "actors.Init wires NPC instances to actors")
	a.Character.Stats.Perception.ValueAdj = perception
	a.Character.Stats.Speed.ValueAdj = speed
	return a
}

// player is a test actor with an account.
func player(t *testing.T, userId int) *actors.Actor {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	u := &accounts.Account{UserId: userId, Character: c, Password: "$2a$test"}
	a := actors.Spawn(c, actors.NewConnectionController(u))
	t.Cleanup(func() { actors.Remove(a.Id) })
	return a
}

func TestTurnOrderPutsAccountsFirstKeepingOrder(t *testing.T) {
	n1, p1, n2, p2 := npc(t, 0, 0), player(t, 9601), npc(t, 0, 0), player(t, 9602)
	got := New().TurnOrder([]*actors.Actor{n1, p1, n2, p2})
	assert.Equal(t, []*actors.Actor{p1, p2, n1, n2}, got)
}

// matchesOneRoll checks, over many seeds, that roll() answers exactly
// util.Rand(100) < chance from the same seed and makes exactly one draw (the
// next draw after it is the second draw of the stream).
func matchesOneRoll(t *testing.T, chance int, roll func() bool) {
	t.Helper()
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) })
	for seed := int64(1); seed <= 1000; seed++ {
		util.SeedRandForTest(seed)
		want := util.Rand(100) < chance
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		require.Equal(t, want, roll(), "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: exactly one draw", seed)
	}
}

func TestFindsWeaponRollsUnderPerception(t *testing.T) {
	a := npc(t, 50, 0)
	matchesOneRoll(t, 50, func() bool { return New().FindsWeapon(a) })
}

func TestFleeEscapesOnSpeedShareOf70Plus30(t *testing.T) {
	// Unequal Speed, so swapping fleer and blocker changes the chance:
	// int(10/(10+30)*70) = 17, plus 30.
	fleer, blocker := npc(t, 0, 10), npc(t, 0, 30)
	matchesOneRoll(t, 47, func() bool { return New().FleeEscapes(fleer, blocker) })
}

func TestHatredRoundsIsTwoMinutesLessPerception(t *testing.T) {
	a := player(t, 9603)
	a.Character.Stats.Perception.ValueAdj = 7
	assert.Equal(t, configs.GetConfig().Timing.MinutesToRounds(2)-7, New().HatredRounds(a))
}

func TestCanAttackAllowsAnyNPCTarget(t *testing.T) {
	assert.NoError(t, New().CanAttack(player(t, 9604), npc(t, 0, 0)))
}

// A safe room protects whoever stands in it, and that check comes before
// the PvP settings. The attacker stands in a loaded room and the recovery
// room is a loaded, non-zero room (room 0 aliases to the start room; new
// bodies start in unloaded room -1). PvP is disabled, so a missing check, or
// one moved after CanPvp, reads "PVP is disabled." instead.
func TestCanAttackRefusesATargetInTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94001, 94002
	// RestoreOverrides lays keys over the live config and resets nothing
	// else, so the cleanup names each key's old value (as setConfig does).
	old := map[string]any{
		"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		"GamePlay.PVP.Enabled":           string(configs.GetPVPConfig().Enabled),
		"GamePlay.PVP.MinimumLevel":      int(configs.GetPVPConfig().MinimumLevel),
	}
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["SpecialRooms.DeathRecoveryRoom"] = recovery
	next["GamePlay.PVP.Enabled"] = "disabled"
	next["GamePlay.PVP.MinimumLevel"] = 0
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		for k, v := range old {
			back[k] = v
		}
		require.NoError(t, configs.RestoreOverrides(back))
		require.Equal(t, old["SpecialRooms.DeathRecoveryRoom"], int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "config restored")
	})
	require.Equal(t, recovery, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom))

	for _, id := range []int{here, recovery} {
		r := rooms.NewEmptyRoom()
		r.RoomId, r.Zone = id, "rulesetzone"
		rooms.AddTestRoom(r)
		id := id
		t.Cleanup(func() { rooms.RemoveTestRoom(id) })
	}
	attacker, target := player(t, 9607), player(t, 9608)
	attacker.Character.RoomId, target.Character.RoomId = here, recovery
	assert.EqualError(t, New().CanAttack(attacker, target), "Fighting is not allowed here.")
}

// setConfig sets config keys for one test and puts back the old values
// named in old (RestoreOverrides resets nothing it is not given; as
// setConfig in internal/hooks).
func setConfig(t *testing.T, set, old map[string]any) {
	t.Helper()
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	for k, v := range set {
		next[k] = v
	}
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		for k, v := range old {
			back[k] = v
		}
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

// testRoom puts an empty room in memory for one test.
func testRoom(t *testing.T, id int, zone string) {
	t.Helper()
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = id, zone
	rooms.AddTestRoom(r)
	t.Cleanup(func() { rooms.RemoveTestRoom(id) })
}

// recoveryCopy makes roomId (a room testRoom put in memory, in zone) the
// death recovery room, saves its template into a temporary data folder and
// makes a private copy of it, as a death does. It returns the copy's ID.
func recoveryCopy(t *testing.T, roomId int, zone string) int {
	t.Helper()
	tmp := t.TempDir()
	setConfig(t,
		map[string]any{"FilePaths.DataFiles": tmp, "SpecialRooms.DeathRecoveryRoom": roomId},
		map[string]any{
			"FilePaths.DataFiles":            configs.GetFilePathsConfig().DataFiles.String(),
			"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		})
	require.Equal(t, roomId, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "the override took")
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "rooms", zone), 0o755))
	require.NoError(t, rooms.SaveRoomTemplate(*rooms.LoadRoom(roomId)))
	copies, err := rooms.CreateEphemeralRoomIds(roomId)
	require.NoError(t, err)
	id := copies[roomId]
	require.True(t, rooms.IsEphemeralRoomId(id))
	// Free the copy's ephemeral chunk (one of 100; ephemeral.go:14): removing
	// only the room would hold the chunk for the rest of the run.
	t.Cleanup(func() { rooms.TryEphemeralCleanup(id) })
	return id
}

// A target in a private copy of the recovery room is protected as one in
// the room itself. PvP is disabled, so a missed copy reads "PVP is
// disabled." instead.
func TestCanAttackRefusesATargetInACopyOfTheRecoveryRoom(t *testing.T) {
	const here, recovery = 94003, 94004
	testRoom(t, here, "rulesetzone")
	testRoom(t, recovery, "rulesetzone")
	copyId := recoveryCopy(t, recovery, "rulesetzone")
	setConfig(t,
		map[string]any{"GamePlay.PVP.Enabled": "disabled", "GamePlay.PVP.MinimumLevel": 0},
		map[string]any{
			"GamePlay.PVP.Enabled":      string(configs.GetPVPConfig().Enabled),
			"GamePlay.PVP.MinimumLevel": int(configs.GetPVPConfig().MinimumLevel),
		})
	attacker, target := player(t, 9609), player(t, 9610)
	attacker.Character.RoomId, target.Character.RoomId = here, copyId
	assert.EqualError(t, New().CanAttack(attacker, target), "Fighting is not allowed here.")
}

func TestConditionIsDeadAtZeroOrBelow(t *testing.T) {
	a := player(t, 9620)
	for health, want := range map[int]rules.Condition{1: rules.Up, 0: rules.Dead, -1: rules.Dead} {
		a.Character.Health = health
		assert.Equal(t, want, New().Condition(a), "health %d", health)
	}
}

// Parity with the old AutoHeal (NewRound_AutoHeal.go before slice 3).
func TestRegenOutsideTheRecoveryRoomIsTheOldAutoHeal(t *testing.T) {
	const here, recovery = 94005, 94006
	setConfig(t,
		map[string]any{"SpecialRooms.DeathRecoveryRoom": recovery},
		map[string]any{"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)})
	none := rules.RegenResult{}

	a := player(t, 9621)
	a.Character.RoomId, a.Character.Health = here, 5
	full := rules.RegenResult{Health: a.Character.HealthPerRound(), Mana: a.Character.ManaPerRound()}
	require.NotZero(t, full.Health)
	assert.Equal(t, full, New().Regen(a, 3), "every 3rd round")
	assert.Equal(t, full, New().Regen(a, 6))
	assert.Equal(t, none, New().Regen(a, 4), "only every 3rd round")
	assert.Equal(t, none, New().Regen(a, 5))

	a.Character.Aggro = &characters.Aggro{Type: characters.DefaultAttack}
	assert.Equal(t, none, New().Regen(a, 3), "never in a fight")
	a.Character.Aggro = nil

	a.Character.Health = 0
	assert.Equal(t, none, New().Regen(a, 3), "never while down")

	n := npc(t, 0, 0)
	n.Character.RoomId, n.Character.Health = here, 5
	assert.Equal(t, none, New().Regen(n, 3), "accounts only")
}

// reseed puts the shared random source back on a clock seed after a test.
func reseed(t *testing.T) {
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) })
}

// The recovery room (and each private copy of it) heals 1d10 health and
// 1d10 mana every round, health first, drawn from the shared source.
func TestRegenInTheRecoveryRoomRollsD10HealthAndMana(t *testing.T) {
	const recovery = 94007
	testRoom(t, recovery, "regenzone")
	copyId := recoveryCopy(t, recovery, "regenzone")
	a := player(t, 9622)
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	reseed(t)
	for _, roomId := range []int{recovery, copyId} {
		a.Character.RoomId = roomId
		for seed := int64(1); seed <= 200; seed++ {
			a.Character.Health, a.Character.Mana = 1, 0
			util.SeedRandForTest(seed)
			h := util.Rand(10) + 1
			m := util.Rand(10) + 1
			next := util.Rand(1000)

			util.SeedRandForTest(seed)
			got := New().Regen(a, uint64(seed)) // every round, not only every 3rd
			require.Equal(t, h, got.Health, "room %d seed %d", roomId, seed)
			require.Equal(t, m, got.Mana, "room %d seed %d", roomId, seed)
			require.Equal(t, next, util.Rand(1000), "room %d seed %d: two draws", roomId, seed)
			require.Equal(t, 1, a.Character.Health, "the ruleset changes no health")

			// At full health and mana nothing heals, and the two dice are
			// still drawn, as upstream's buff drew them (Q7h).
			a.Character.Health, a.Character.Mana = 100, 100
			util.SeedRandForTest(seed)
			require.Equal(t, rules.RegenResult{}, New().Regen(a, uint64(seed)), "room %d seed %d: full", roomId, seed)
			require.Equal(t, next, util.Rand(1000), "room %d seed %d: full, still two draws", roomId, seed)
		}
	}

	// Any actor there heals, an NPC too (Q7h; outside it, accounts only).
	n := npc(t, 0, 0)
	n.Character.RoomId = copyId
	n.Character.HealthMax.Value, n.Character.ManaMax.Value = 100, 100
	n.Character.Health, n.Character.Mana = 1, 0
	util.SeedRandForTest(1)
	require.True(t, New().Regen(n, 1).Changes(), "an NPC in the recovery room heals")
}

// The amounts are what Heal will apply (capped at the maximums), and the
// lines are upstream's buff's: one for the actor, one for the room.
func TestRegenInTheRecoveryRoomSaysWhatItHealed(t *testing.T) {
	const recovery = 94008
	testRoom(t, recovery, "regenzone")
	recoveryCopy(t, recovery, "regenzone")
	a := player(t, 9623)
	a.Character.RoomId = recovery
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	reseed(t)
	util.SeedRandForTest(1)

	a.Character.Health, a.Character.Mana = 1, 0
	got := New().Regen(a, 1)
	require.Len(t, got.Messages, 2)
	assert.Equal(t, perception.Interface, got.Messages[0].Kind)
	assert.Equal(t, perception.ActorOnly, got.Messages[0].Audience)
	assert.Equal(t, a.Id, got.Messages[0].Actor)
	assert.Equal(t, fmt.Sprintf(`The shadow realm heals you for <ansi fg="healing">%d damage</ansi> and restores <ansi fg="mana-100">%d mana</ansi>!`, got.Health, got.Mana), got.Messages[0].ActorText)
	assert.Equal(t, perception.WorldEvent, got.Messages[1].Kind)
	assert.Equal(t, a.Id, got.Messages[1].Actor)
	assert.Equal(t, perception.Sight, got.Messages[1].Senses, "the room sees it")
	assert.Equal(t, perception.Room, got.Messages[1].Range, "the room, not further")
	assert.Equal(t, `{actor} is recovering from a recent death.`, got.Messages[1].ObserverText)

	// Capped at the maximums: one point short heals exactly one point, and
	// says so, whatever the dice (over 20 seeds, most of the d10s roll above 1).
	for seed := int64(1); seed <= 20; seed++ {
		util.SeedRandForTest(seed)
		a.Character.Health, a.Character.Mana = 99, 100
		got = New().Regen(a, 2)
		require.Equal(t, 1, got.Health, "seed %d: capped at the health maximum", seed)
		require.Zero(t, got.Mana, "seed %d: mana is full", seed)
		require.NotEmpty(t, got.Messages)
		require.Equal(t, `The shadow realm heals you for <ansi fg="healing">1 damage</ansi>!`, got.Messages[0].ActorText)

		util.SeedRandForTest(seed)
		a.Character.Health, a.Character.Mana = 100, 99
		got = New().Regen(a, 3)
		require.Zero(t, got.Health, "seed %d: health is full", seed)
		require.Equal(t, 1, got.Mana, "seed %d: capped at the mana maximum", seed)
		require.NotEmpty(t, got.Messages)
		require.Equal(t, `The shadow realm restores <ansi fg="mana-100">1 mana</ansi>!`, got.Messages[0].ActorText)
	}

	a.Character.Health, a.Character.Mana = 100, 100
	got = New().Regen(a, 4)
	assert.Equal(t, rules.RegenResult{}, got, "nothing to heal: no lines")
}

// The recovery room's line is only the body's presence (round 5): a viewer
// there who does not perceive the hidden body is excluded from it; an admin
// is not, nor is anyone while the body is not hidden.
func TestRecoveryLineExcludesThoseWhoDoNotPerceive(t *testing.T) {
	const recovery = 94009
	testRoom(t, recovery, "regenzone")
	recoveryCopy(t, recovery, "regenzone")
	r := rooms.LoadRoom(recovery)
	a, dan, eve := player(t, 9651), player(t, 9652), player(t, 9653)
	for userId, p := range map[int]*actors.Actor{9651: a, 9652: dan, 9653: eve} {
		p.Character.RoomId = recovery
		r.AddPlayer(userId)
	}
	actors.AccountOf(eve.Id).Role = accounts.RoleAdmin
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	reseed(t)

	a.Character.Health, a.Character.Mana = 1, 0
	got := New().Regen(a, 1)
	require.Len(t, got.Messages, 2)
	assert.Empty(t, got.Messages[1].Exclude, "unhidden: everyone perceives it")

	const hiddenBuff = 96209
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: hiddenBuff, Name: "test hidden", TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{"hidden"},
	}))
	require.NoError(t, a.Character.AddBuff(hiddenBuff, true))
	require.True(t, a.Character.HasBuffFlag("hidden"))

	// AddBuff runs Validate, which recomputes the maximums: set them again.
	a.Character.HealthMax.Value, a.Character.ManaMax.Value = 100, 100
	a.Character.Health, a.Character.Mana = 1, 0
	got = New().Regen(a, 2)
	require.Len(t, got.Messages, 2)
	assert.Equal(t, []ids.ID{dan.Id}, got.Messages[1].Exclude, "Dan does not perceive the hidden body; Eve, an admin, does")
}

// trinket is an item with its own spec (no item file needed).
func trinket(itemId int, name string, typ items.ItemType) items.Item {
	spec := items.ItemSpec{ItemId: itemId, Name: name, NameSimple: name, Type: typ}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

func idsOf(list []items.Item) []int {
	out := []int{}
	for _, it := range list {
		out = append(out, it.ItemId)
	}
	return out
}

// onlinePlayer is player, registered so accounts.GetByUserId finds it.
func onlinePlayer(t *testing.T, userId int) *actors.Actor {
	t.Helper()
	a := player(t, userId)
	accounts.SetTestUser(a.Controller().Account())
	t.Cleanup(func() { accounts.RemoveTestUser(userId) })
	return a
}

// flagged gives c a test buff carrying one flag. A bare test NPC
// (mobs.NewTestInstance(mobs.Mob{})) has a zero Buffs whose maps are nil,
// and AddBuff writes into them (buffs.go:187), so it gets buffs.New() first,
// as givePermaGear does (usercommands/items_test.go:42).
func flagged(t *testing.T, c *characters.Character, buffId int, flag string) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000, Flags: []string{flag},
	}))
	if c.Buffs.List == nil { // buffs.New() makes List non-nil (buffs.go:43-49)
		c.Buffs = buffs.New()
	}
	require.NoError(t, c.AddBuff(buffId, true))
	require.True(t, c.HasBuffFlag(flag))
}

// Decision 11: a player drops what it carries and its gold, keeps what it
// wears, and comes back at 1 health.
func TestOnDeathAPlayerDropsWhatItCarriesAndKeepsWhatItWears(t *testing.T) {
	a := player(t, 9630)
	a.Character.Items = []items.Item{trinket(96101, "testgem", items.Gemstone)}
	a.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)
	a.Character.Gold = 25

	out := New().OnDeath(a)
	assert.Equal(t, []int{96101}, idsOf(out.DropItems))
	assert.Equal(t, 25, out.DropGold)
	assert.True(t, out.Respawn)
	assert.Equal(t, 1, out.RespawnHealth)
	assert.False(t, out.Revive)
	assert.Empty(t, out.Awards)
}

// Ruling 3: an NPC drops what it carries and its gold always, and each worn
// item at its ItemDropChance, one d100 per item in slot order (Neck before
// Ring, items.AllEquipSlots).
func TestOnDeathAnNPCDropsEachWornItemAtItsDropChance(t *testing.T) {
	n := npc(t, 0, 0)
	mobs.GetInstance(n.Id).ItemDropChance = 50
	n.Character.Items = []items.Item{trinket(96101, "testgem", items.Gemstone)}
	n.Character.Equipment.Neck = trinket(96103, "testamulet", items.Neck)
	n.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)
	n.Character.Gold = 30
	reseed(t)
	for seed := int64(1); seed <= 500; seed++ {
		util.SeedRandForTest(seed)
		want := []int{96101}
		if util.Rand(100) < 50 {
			want = append(want, 96103)
		}
		if util.Rand(100) < 50 {
			want = append(want, 96102)
		}
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		out := New().OnDeath(n)
		require.Equal(t, want, idsOf(out.DropItems), "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: one draw per worn item", seed)
		require.Equal(t, 30, out.DropGold)
		require.False(t, out.Respawn, "an NPC's body ends")
	}
}

// The draw order of the moved code: the XP variation (one d3) before the
// worn-item rolls. Both alignments are 0, so UpdateAlignment draws nothing.
func TestOnDeathDrawsTheXPVariationBeforeTheDrops(t *testing.T) {
	onlinePlayer(t, 9631)
	n := npc(t, 0, 0)
	mobs.GetInstance(n.Id).ItemDropChance = 50
	n.Character.PlayerDamage = map[int]int{9631: 5}
	n.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)

	// One killer within 5 levels of the NPC: no scaler, not elite.
	base := n.Character.XPTL(n.Character.Level-1) / 90
	variation := base / 100
	if variation < 1 {
		variation = 1
	}
	reseed(t)
	for seed := int64(1); seed <= 300; seed++ {
		util.SeedRandForTest(seed)
		xp := base + (util.Rand(3)-1)*variation
		ring := util.Rand(100) < 50
		next := util.Rand(1000)

		util.SeedRandForTest(seed)
		out := New().OnDeath(n)
		require.Len(t, out.Awards, 1)
		require.Equal(t, 9631, out.Awards[0].UserId)
		require.Equal(t, xp, out.Awards[0].XP, "seed %d", seed)
		require.Equal(t, ring, len(out.DropItems) == 1, "seed %d", seed)
		require.Equal(t, next, util.Rand(1000), "seed %d: two draws", seed)
	}
}

// An alignment shift is applied to the killer, returned as its line, and
// queued as AlignmentChanged (the ruleset may queue an event: spec, plan
// review round 1).
func TestOnDeathShiftsTheKillersAlignment(t *testing.T) {
	killer := onlinePlayer(t, 9632)
	killer.Character.Alignment = 19 // neutral; killing an evil NPC makes her lawful (reads golden H)
	n := npc(t, 0, 0)
	n.Character.Alignment = -100
	n.Character.PlayerDamage = map[int]int{9632: 5}
	var changed []events.AlignmentChanged
	id := events.RegisterListener(events.AlignmentChanged{}, func(e events.Event) events.ListenerReturn {
		changed = append(changed, e.(events.AlignmentChanged))
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.AlignmentChanged{}, id) })
	events.ProcessEvents() // drop what setup queued

	out := New().OnDeath(n)
	events.ProcessEvents()
	require.Len(t, out.Awards, 1)
	require.Len(t, out.Awards[0].Messages, 1)
	assert.Contains(t, out.Awards[0].Messages[0].ActorText, "Your alignment has shifted from")
	assert.Equal(t, killer.Id, out.Awards[0].Messages[0].Actor)
	assert.Equal(t, "lawful", killer.Character.AlignmentName())
	require.Len(t, changed, 1)
	assert.Equal(t, events.AlignmentChanged{UserId: 9632, AlignmentOld: "neutral", AlignmentNew: "lawful"}, changed[0])
}

// Party XP, moved as is: a killer in a party puts its share in the party's
// pot, which is split equally among the members, each awarded (a member who
// dealt no damage too), with the elite bonus (10% when EliteXPBonus is
// unset) added to each split for an elite NPC. Both alignments are 0, so
// UpdateAlignment draws nothing.
func TestOnDeathSplitsAPartysShare(t *testing.T) {
	onlinePlayer(t, 9640)
	onlinePlayer(t, 9641)
	p := parties.New(9640)
	require.NotNil(t, p)
	require.True(t, p.InvitePlayer(9641))
	require.True(t, p.AcceptInvite(9641))
	t.Cleanup(p.Disband)
	n := npc(t, 0, 0)
	n.Character.Level = 5
	n.Character.TNLScale = 1 // a bare test NPC has 0, which makes its XP 0 (XPTL, character.go:923-933)
	n.Character.PlayerDamage = map[int]int{9640: 5}

	base := n.Character.XPTL(n.Character.Level-1) / 90
	variation := max(base/100, 1)
	reseed(t)
	for _, elite := range []bool{false, true} {
		mobs.GetInstance(n.Id).IsElite = elite
		util.SeedRandForTest(7)
		split := (base + (util.Rand(3)-1)*variation) / 2
		if elite {
			split += int(math.Ceil(float64(split) * 10 / 100.0))
		}
		require.Positive(t, split, "the fixture needs XP to share")

		util.SeedRandForTest(7)
		out := New().OnDeath(n)
		require.Len(t, out.Awards, 2, "elite %v: one award per member", elite)
		assert.Equal(t, []int{9640, 9641}, []int{out.Awards[0].UserId, out.Awards[1].UserId}, "elite %v: member order", elite)
		assert.Equal(t, split, out.Awards[0].XP, "elite %v", elite)
		assert.Equal(t, split, out.Awards[1].XP, "elite %v", elite)
	}
}

func TestOnDeathPermaGearDropsNothing(t *testing.T) {
	p, n := player(t, 9633), npc(t, 0, 0)
	mobs.GetInstance(n.Id).ItemDropChance = 100
	for _, a := range []*actors.Actor{p, n} {
		flagged(t, a.Character, 96201, "perma-gear")
		a.Character.Items = []items.Item{trinket(96101, "testgem", items.Gemstone)}
		a.Character.Equipment.Ring = trinket(96102, "testring", items.Ring)
		a.Character.Gold = 25
		out := New().OnDeath(a)
		assert.Empty(t, out.DropItems)
		assert.Zero(t, out.DropGold)
	}
}

// revive-on-death revives any actor at full health (an NPC's rule before
// slice 3, now everyone's).
func TestOnDeathReviveOnDeathRevivesAnyActorAtFullHealth(t *testing.T) {
	for name, a := range map[string]*actors.Actor{"account": player(t, 9634), "NPC": npc(t, 0, 0)} {
		flagged(t, a.Character, 96202, "revive-on-death")
		a.Character.HealthMax.Value = 33 // after AddBuff's Validate
		a.Character.Gold = 25
		out := New().OnDeath(a)
		assert.True(t, out.Revive, name)
		assert.Equal(t, 33, out.ReviveHealth, name)
		assert.Empty(t, out.DropItems, name)
		assert.Zero(t, out.DropGold, name)
	}
}
