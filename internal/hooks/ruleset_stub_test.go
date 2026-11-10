package hooks

import (
	"fmt"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/perception"
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
	require.Contains(t, out, `Input rat: "suicide"`)
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
