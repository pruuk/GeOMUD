package hooks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/stretchr/testify/require"
)

// oneShot is the active ruleset with one rule changed: every attack hits for
// 1000.
type oneShot struct{ rules.Ruleset }

func (oneShot) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
	return rules.AttackResult{Hit: true, DamageToTarget: 1000, Messages: []perception.Message{{
		Kind: perception.WorldEvent, Actor: a.Id, Target: d.Id,
		Senses: perception.Sight, Range: perception.Room, ActorText: "STUB HIT",
	}}}
}

// The pvm_death fixture under a ruleset where every hit kills: the rat dies
// in round 1 instead of round 4, with no change to the engine.
func TestStubRulesetChangesTheFight(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "stub-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(oneShot{rules.Get()}))
	w.start()
	w.rounds(20)

	out := w.out.String()
	require.Contains(t, out, "== round 1")
	require.NotContains(t, out, "== round 2", "the reference ruleset takes 4 rounds (golden pvm_death)")
	require.Contains(t, out, "MobDeath rat")
	require.Contains(t, out, "STUB HIT")
}

// reversed is the active ruleset with the turn order reversed (NPCs
// first). It also records who resolves an attack, in order, and otherwise
// delegates, so the random draws are the reference ruleset's.
type reversed struct {
	rules.Ruleset
	attackers *[]ids.ID
}

func (r reversed) TurnOrder(f []*actors.Actor) []*actors.Actor {
	in := r.Ruleset.TurnOrder(f)
	out := make([]*actors.Actor, 0, len(in))
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return out
}

func (r reversed) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
	*r.attackers = append(*r.attackers, a.Id)
	return r.Ruleset.ResolveAttack(a, d)
}

// Turn order is the ruleset's: reversed, the guard attacks before Alice in
// round 1 (the reference ruleset has Alice first).
func TestStubRulesetChangesTurnOrder(t *testing.T) {
	w := newCombatWorld(t, 202)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 50)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 50)
	alice.Character.SetAggro(guard.Character.Id, characters.DefaultAttack)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	var attackers []ids.ID
	t.Cleanup(rules.SetForTest(reversed{Ruleset: rules.Get(), attackers: &attackers}))
	w.start()
	w.rounds(1)

	require.Len(t, attackers, 2)
	require.Equal(t, []ids.ID{guard.Character.Id, alice.Character.Id}, attackers)
}

// armed is the active ruleset with the weapon search forced to an answer.
// It records whom it is asked about.
type armed struct {
	rules.Ruleset
	finds bool
	asked *[]ids.ID
}

func (r armed) FindsWeapon(npc *actors.Actor) bool {
	*r.asked = append(*r.asked, npc.Id)
	return r.finds
}

// The weapon search is the ruleset's: an unarmed guard carrying a weapon
// queues "equip" when the ruleset says it finds one, and not otherwise. Each
// case gives the guard the Perception that makes the reference ruleset
// answer the opposite, so each fails if the engine ignores the stub.
func TestStubRulesetDecidesTheWeaponSearch(t *testing.T) {
	for _, finds := range []bool{true, false} {
		t.Run(fmt.Sprint(finds), func(t *testing.T) {
			w := newCombatWorld(t, 303)
			alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 50)
			guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 50)
			guard.Character.Items = append(guard.Character.Items, weapon(93106, "testclub", items.Bludgeoning, 1, 4, 0))
			if finds {
				guard.Character.Stats.Perception.ValueAdj = 0 // the reference never finds
			} else {
				guard.Character.Stats.Perception.ValueAdj = 100 // the reference always finds
			}
			guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
			var asked []ids.ID
			t.Cleanup(rules.SetForTest(armed{Ruleset: rules.Get(), finds: finds, asked: &asked}))
			w.start()
			w.rounds(1)

			if finds {
				require.Contains(t, w.out.String(), `Input guard: "equip testclub"`)
			} else {
				require.NotContains(t, w.out.String(), `"equip `)
			}
			require.Equal(t, []ids.ID{guard.Character.Id}, asked, "asked about the NPC, not its target")
		})
	}
}

// hateful is the active ruleset with group hatred lasting 2 rounds. It
// records whom it is asked about.
type hateful struct {
	rules.Ruleset
	asked *[]ids.ID
}

func (h hateful) HatredRounds(a *actors.Actor) int {
	*h.asked = append(*h.asked, a.Id)
	return 2
}

// Hatred length is the ruleset's, asked about the attacker: the rat's group
// hates Alice for exactly two passes of ReduceHostility.
func TestStubRulesetSetsHatredLength(t *testing.T) {
	w := newCombatWorld(t, 111)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 500, "hate-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	var asked []ids.ID
	t.Cleanup(rules.SetForTest(hateful{Ruleset: rules.Get(), asked: &asked}))
	w.start()
	w.rounds(1)

	require.Equal(t, []ids.ID{alice.Character.Id}, asked)
	group := rat.Groups[0] // addNPC suffixes the name per run
	require.True(t, mobs.IsHostile(group, alice.UserId))
	mobs.ReduceHostility()
	require.True(t, mobs.IsHostile(group, alice.UserId), "1 round left")
	mobs.ReduceHostility()
	require.False(t, mobs.IsHostile(group, alice.UserId), "2 rounds are up")
}

// Downed is generic engine handling: under a ruleset with a downed state
// (downedAtZero, combat_golden_test.go) the round downs Alice instead of
// killing her, the room reads her drop, and nobody dies.
func TestStubRulesetDownedStateIsNotDeath(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	id := events.RegisterListener(events.PlayerDrop{}, HandlePlayerDrop)
	t.Cleanup(func() { events.UnregisterListener(events.PlayerDrop{}, id) })
	t.Cleanup(rules.SetForTest(downedAtZero{rules.Get()}))
	w.start()
	w.rounds(20)

	out := w.out.String()
	require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "not moved to the recovery room")
	require.Contains(t, out, "PlayerDrop Alice")
	require.Contains(t, out, `you drop to the ground!`)
	require.NotContains(t, out, "PlayerDeath")
}

// reviver is the active ruleset where every death is a revive at 7 health.
type reviver struct{ rules.Ruleset }

func (reviver) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{Revive: true, ReviveHealth: 7}
}

func TestStubRulesetRevivesInsteadOfDeath(t *testing.T) {
	w := newCombatWorld(t, 111)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "revive-rats")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(reviver{rules.Get()}))
	w.start()
	w.rounds(20)

	out := w.out.String()
	require.Equal(t, 7, rat.Character.Health)
	require.NotNil(t, mobs.GetInstance(rat.Character.Id), "the rat lives on")
	require.Contains(t, out, "is suddenly revived in a shower of sparks!")
	require.NotContains(t, out, "MobDeath")
}

// fixedDrops is the active ruleset where a death drops exactly drop and 3
// gold, and the body ends.
type fixedDrops struct {
	rules.Ruleset
	drop items.Item
}

func (f fixedDrops) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{DropItems: []items.Item{f.drop}, DropGold: 3}
}

// The corpse holds exactly the ruleset's drops, and its body nothing (Q2):
// the rat keeps a second carried item, its ring and 27 gold, and none of
// them may show on the corpse's body copy.
func TestStubRulesetDecidesTheDrops(t *testing.T) {
	w := newCombatWorld(t, 111)
	deathConfig(t) // corpses on
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "drop-rats")
	gem := trinket(93108, "testgem", items.Gemstone, items.Mundane)
	pebble := trinket(93109, "testpebble", items.Gemstone, items.Mundane) // carried, not dropped
	rat.Character.Items = []items.Item{gem, pebble}
	rat.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	rat.Character.Gold = 30
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(fixedDrops{Ruleset: rules.Get(), drop: gem}))
	w.start()
	w.rounds(20)

	room := rooms.LoadRoom(goldenRoomA)
	require.Len(t, room.Corpses, 1)
	c := room.Corpses[0]
	require.Equal(t, "93108", itemIds(c.Items))
	require.Equal(t, 3, c.Gold)
	require.Equal(t, "93109", itemIds(rat.Character.Items), "the rat's body kept the pebble (the fixture makes the next asserts mean something)")
	require.Equal(t, 27, rat.Character.Gold, "and 27 gold")
	require.Empty(t, c.Character.GetAllWornItems(), "the corpse's body wears nothing")
	require.Empty(t, c.Character.Items, "the corpse's body carries nothing")
	require.Zero(t, c.Character.Gold, "the corpse's body holds no gold")
	require.Empty(t, room.Items, "nothing on the floor")
	require.Zero(t, room.Gold)
}

// plusFive is the active ruleset where an account regenerates 5 health
// every round, with a line.
type plusFive struct{ rules.Ruleset }

func (plusFive) Regen(a *actors.Actor, round uint64) rules.RegenResult {
	if a.Controller() == nil || a.Controller().Account() == nil {
		return rules.RegenResult{}
	}
	return rules.RegenResult{Health: 5, Messages: []perception.Message{{
		Kind: perception.Interface, Actor: a.Id, Audience: perception.ActorOnly, ActorText: "STUB REGEN",
	}}}
}

// The regen pass applies the ruleset's amounts and sends its lines.
func TestStubRulesetDecidesRegen(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 20)
	alice.Character.Health = 3
	t.Cleanup(rules.SetForTest(plusFive{rules.Get()}))
	w.start()

	AutoHeal(events.NewRound{RoundNumber: 1}) // not a 3rd round: the reference ruleset gives nothing
	events.ProcessEvents()

	require.Equal(t, 8, alice.Character.Health)
	var reads, notes []string
	for _, e := range w.rec.take() {
		if e.userId == 0 {
			notes = append(notes, e.text)
		} else {
			reads = append(reads, e.text)
		}
	}
	require.Contains(t, strings.Join(reads, ""), "STUB REGEN")
	require.Contains(t, notes, "CharacterVitalsChanged Alice")
}

// respawner is the active ruleset where every death respawns the body at 4
// health.
type respawner struct{ rules.Ruleset }

func (respawner) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{Respawn: true, RespawnHealth: 4}
}

// The engine respawns at the ruleset's health (not a fixed 1), and respawns
// an NPC too when the ruleset says so: its body does not end.
func TestStubRulesetDecidesTheRespawn(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "respawn-rats")
	t.Cleanup(rules.SetForTest(respawner{rules.Get()}))
	w.start()

	for _, c := range []*characters.Character{alice.Character, &rat.Character} {
		c.Health = 0
		death.Die(actors.Get(c.Id))
		events.ProcessEvents()
		require.Equal(t, 4, c.Health, "%s respawns at the ruleset's health", c.Name)
		require.Equal(t, goldenRoomC, rooms.GetOriginalRoom(c.RoomId), "%s is in a copy of the recovery room", c.Name)
	}
	require.NotNil(t, mobs.GetInstance(rat.Character.Id), "the rat's body did not end")
}

// overDrops is the active ruleset whose drops ask for more than the body
// has: an item it does not hold, and gold.
type overDrops struct {
	rules.Ruleset
	drops []items.Item
	gold  int
}

func (o overDrops) OnDeath(v *actors.Actor) rules.DeathOutcome {
	return rules.DeathOutcome{DropItems: o.drops, DropGold: o.gold}
}

// The engine drops only what the body has: an item it does not hold is not
// made from nothing, and the gold is at most the body's and never below 0.
func TestStubRulesetCannotDropWhatTheBodyLacks(t *testing.T) {
	gem := trinket(93108, "testgem", items.Gemstone, items.Mundane)
	phantom := trinket(93109, "testpebble", items.Gemstone, items.Mundane) // never held
	for _, tc := range []struct{ gold, wantCorpse, wantBody int }{
		{1000, 30, 0}, // more than it has: all of it
		{-5, 0, 30},   // below 0: none
	} {
		t.Run(fmt.Sprint(tc.gold), func(t *testing.T) {
			w := newCombatWorld(t, 111)
			deathConfig(t) // corpses on
			rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "overdrop-rats")
			rat.Character.Items = []items.Item{gem}
			rat.Character.Gold = 30
			t.Cleanup(rules.SetForTest(overDrops{Ruleset: rules.Get(), drops: []items.Item{gem, phantom}, gold: tc.gold}))
			w.start()

			death.Die(actors.Get(rat.Character.Id))
			events.ProcessEvents()

			room := rooms.LoadRoom(goldenRoomA)
			require.Len(t, room.Corpses, 1)
			require.Equal(t, "93108", itemIds(room.Corpses[0].Items), "the held gem, not the phantom")
			require.Equal(t, tc.wantCorpse, room.Corpses[0].Gold)
			require.Equal(t, tc.wantBody, rat.Character.Gold)
		})
	}
}
