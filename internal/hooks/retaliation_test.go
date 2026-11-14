package hooks

import (
	"errors"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/gamelock"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mutators"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// told is what userId read since the last call.
func (w *combatWorld) told(userId int) []string {
	events.ProcessEvents()
	out := []string{}
	for _, e := range w.rec.take() {
		if e.userId == userId {
			out = append(out, strings.TrimSuffix(e.text, "\n"))
		}
	}
	return out
}

// aggroOn gives c an aggro on target that began on round since.
func aggroOn(c *characters.Character, target ids.ID, since uint64) {
	c.SetAggro(target, characters.DefaultAttack)
	c.Aggro.Since = since
}

// Phase 5 slice 4.1, ruling 7: the next opponent is the one in the room,
// up and perceived, whose aggro on the body began earliest; ties go to turn
// order (accounts first, then NPCs, each in room order).
func TestNextOpponentIsTheLongestAttacker(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	rat := w.addNPC(93205, "rat", 10, goldenRoomA, 5, 5, 5, 30)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30) // last in the room, first in turn order
	a := actors.Get(alice.Character.Id)
	next := func() string {
		if n := death.NextOpponent(a); n != nil {
			return n.Character.Name
		}
		return "none"
	}

	assert.Equal(t, "none", next(), "nobody attacks her")

	aggroOn(&cat.Character, alice.Character.Id, 9)
	aggroOn(&rat.Character, alice.Character.Id, 5)
	assert.Equal(t, "rat", next(), "earliest Since, though the cat is first in turn order")

	rat.Character.Aggro.Since = 9
	assert.Equal(t, "cat", next(), "a tie goes to turn order")

	aggroOn(bob.Character, alice.Character.Id, 9)
	assert.Equal(t, "Bob", next(), "accounts come first in turn order")

	bob.Character.Health = 0
	assert.Equal(t, "cat", next(), "a body that is not up is skipped")

	w.hide(&cat.Character)
	assert.Equal(t, "rat", next(), "a body she cannot perceive is skipped")

	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomB))
	assert.Equal(t, "none", next(), "a body in another room is skipped")
}

// Phase 5 slice 4.1, rulings 7 and 9: settleTarget's five steps.
func TestSettleTargetSteps(t *testing.T) {
	setup := func(t *testing.T) (*combatWorld, *actors.Actor, *rooms.Room) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		return w, actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	}

	t.Run("1 gone, nobody else: ends with today's line", func(t *testing.T) {
		w, a, room := setup(t)
		a.Character.SetAggro(ids.New(ids.NoCell), characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Your target can't be found.`}, w.told(9301))
	})

	t.Run("1 gone, an attacker here: switches", func(t *testing.T) {
		w, a, room := setup(t)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.SetAggro(ids.New(ids.NoCell), characters.DefaultAttack)
		w.told(9301)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id, "the turn goes on to the new opponent")
		assert.Equal(t, cat.Character.Id, a.Character.Aggro.Target)
		assert.Equal(t, []string{`You turn to fight <ansi fg="mobname">cat</ansi>!`}, w.told(9301))
	})

	t.Run("2 not up, nobody else: rage subsides", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
		rat.Character.Health = 0
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Your rage subsides.`}, w.told(9301))
	})

	t.Run("3 in another room, a shot, a bow, an attacker here: kept for the shot", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		typeShoot(a.Character, "north", rat.Character.Id)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id, "a target it is about to shoot is reachable: kept")
		assert.Equal(t, characters.TypedShot, a.Character.Aggro.Shot, "the lookup does not spend the shot")
	})

	t.Run("4 in another room, a bow but no shot, an attacker here: switches", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggroRemote("north", rat.Character.Id, characters.Shooting) // its shot spent
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id, "a target it will not shoot again gives way to the attacker here")
	})

	t.Run("3 a shot and no shooting weapon: dropped, waits", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		typeShoot(a.Character, "north", rat.Character.Id) // owed, but she holds no bow
		assert.Nil(t, settleTarget(a, room))
		require.NotNil(t, a.Character.Aggro, "the aggro waits")
		assert.Equal(t, characters.NoShot, a.Character.Aggro.Shot, "a shot that cannot be fired is dropped")
	})

	t.Run("4 in another room, an attacker here: switches", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id)
	})

	t.Run("3 in another room, a shot, a bow and an unlocked exit: shoots", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		a.Character.Aggro.Shot = characters.EarnedShot // owed, with no exit on the aggro yet
		events.ProcessEvents()
		w.rec.take()
		aggroChanged := func() int {
			events.ProcessEvents()
			n := 0
			for _, e := range w.rec.take() {
				if e.userId == 0 && strings.HasPrefix(e.text, "AggroChanged Alice ") {
					n++
				}
			}
			return n
		}
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id)
		assert.Equal(t, characters.Shooting, a.Character.Aggro.Type)
		assert.Equal(t, "north", a.Character.Aggro.ExitName)
		assert.Equal(t, 1, aggroChanged(), "turning to shoot queues one AggroChanged")
		require.NotNil(t, settleTarget(a, room))
		assert.Equal(t, 0, aggroChanged(), "already shooting through that exit: none")
	})

	t.Run("3 the exit is locked: the shot is dropped, waits", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		room.Exits["north"] = exit.RoomExit{RoomId: goldenRoomB, Lock: gamelock.Lock{Difficulty: 5}}
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		a.Character.Aggro.Shot = characters.TypedShot
		assert.Nil(t, settleTarget(a, room))
		require.NotNil(t, a.Character.Aggro)
		assert.Equal(t, characters.NoShot, a.Character.Aggro.Shot)
	})

	t.Run("5 waits, says so once, swings when they share a room", func(t *testing.T) {
		w, a, room := setup(t)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, settleTarget(a, room))
		assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "told once")
		require.NotNil(t, a.Character.Aggro, "the aggro waits")
		rat.Character.RoomId = goldenRoomA
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id)
	})
}

// Phase 5 slice 4.1, rulings 6, 8 and 9: a defender with no aggro takes one
// on its attacker after an attack, hit or miss.
func TestRetaliate(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	a := actors.Get(alice.Character.Id)

	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
	retaliate(actors.Get(rat.Character.Id), a, false)
	require.NotNil(t, rat.Character.Aggro, "same room: a plain attack")
	assert.Equal(t, alice.Character.Id, rat.Character.Aggro.Target)
	assert.Empty(t, rat.Character.Aggro.ExitName)

	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	aggroOn(&cat.Character, rat.Character.Id, 1)
	retaliate(actors.Get(cat.Character.Id), a, false)
	assert.Equal(t, rat.Character.Id, cat.Character.Aggro.Target, "a body already fighting keeps its target")

	archer := w.addNPC(93206, "archer", 1, goldenRoomB, 5, 5, 5, 30)
	archer.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	retaliate(actors.Get(archer.Character.Id), a, true)
	require.NotNil(t, archer.Character.Aggro)
	assert.Equal(t, characters.Shooting, archer.Character.Aggro.Type, "another room, a typed shot, a bow: shoots back")
	assert.Equal(t, "south", archer.Character.Aggro.ExitName)
	assert.Equal(t, characters.EarnedShot, archer.Character.Aggro.Shot, "one shot back")

	bowman := w.addNPC(93209, "bowman", 1, goldenRoomB, 5, 5, 5, 30)
	bowman.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	retaliate(actors.Get(bowman.Character.Id), a, false)
	require.NotNil(t, bowman.Character.Aggro, "a shot back earns nothing: an aggro that waits")
	assert.Equal(t, characters.NoShot, bowman.Character.Aggro.Shot, "a shot back earns nothing")
	assert.Empty(t, bowman.Character.Aggro.ExitName)

	dog := w.addNPC(93203, "dog", 1, goldenRoomB, 5, 5, 5, 30)
	retaliate(actors.Get(dog.Character.Id), a, true)
	require.NotNil(t, dog.Character.Aggro, "another room, no bow: an aggro that waits")
	assert.Empty(t, dog.Character.Aggro.ExitName)
	assert.Equal(t, characters.NoShot, dog.Character.Aggro.Shot)

	down := w.addNPC(93207, "mouse", 10, goldenRoomA, 5, 5, 5, 30)
	down.Character.Health = 0
	retaliate(actors.Get(down.Character.Id), a, false)
	assert.Nil(t, down.Character.Aggro, "a body that is not up does not retaliate")

	t.Cleanup(rules.SetForTest(refusesAll{rules.Get()}))
	owl := w.addNPC(93208, "owl", 10, goldenRoomA, 5, 5, 5, 30)
	retaliate(actors.Get(owl.Character.Id), a, false)
	assert.Nil(t, owl.Character.Aggro, "the ruleset's CanAttack refuses")
}

// Phase 5 slice 4.1, ruling 13: one shot per action. A typed shot fires
// once and earns its target one shot back; a shot back earns nothing; a
// second typed shoot earns a second shot back; a body never holds more than
// one shot; a shot from someone the body is not fighting earns nothing.
func TestTypedShotEarnsOneShotBack(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 500)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomB, 30, 30, 20, 500)
	bob.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	a, b := actors.Get(alice.Character.Id), actors.Get(bob.Character.Id)
	typeShoot(alice.Character, "north", bob.Character.Id)

	combatTurn(a) // Alice fires her typed shot
	require.NotNil(t, alice.Character.Aggro)
	assert.Equal(t, characters.NoShot, alice.Character.Aggro.Shot, "fired: spent")
	assert.Equal(t, "north", alice.Character.Aggro.ExitName, "the re-arm keeps the exit")
	assert.True(t, alice.Character.Aggro.OutOfReachTold, "the shot sets the told mark")
	require.NotNil(t, bob.Character.Aggro, "Bob retaliates")
	assert.Equal(t, characters.EarnedShot, bob.Character.Aggro.Shot, "a typed shot earns one back")

	combatTurn(b) // Bob fires back
	assert.Equal(t, characters.NoShot, bob.Character.Aggro.Shot, "fired: spent")
	assert.Equal(t, characters.NoShot, alice.Character.Aggro.Shot, "a shot back earns nothing")

	w.told(9301)
	w.told(9302)
	combatTurn(a)
	combatTurn(b)
	assert.Empty(t, w.told(9301), "Alice neither fires nor is told her target is out of reach")
	assert.Equal(t, characters.NoShot, bob.Character.Aggro.Shot)

	typeShoot(alice.Character, "north", bob.Character.Id) // Alice types shoot again
	combatTurn(a)
	assert.Equal(t, characters.EarnedShot, bob.Character.Aggro.Shot, "a second typed shoot earns a second shot back")

	bob.Character.Aggro.Shot = characters.TypedShot // as if Bob had typed shoot too
	typeShoot(alice.Character, "north", bob.Character.Id)
	combatTurn(a)
	assert.Equal(t, characters.TypedShot, bob.Character.Aggro.Shot, "never more than one: his own typed shot stays")

	carol := w.addPlayer(9303, "Carol", goldenRoomA, 30, 30, 20, 500)
	carol.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	bob.Character.Aggro.Shot = characters.NoShot
	typeShoot(carol.Character, "north", bob.Character.Id)
	combatTurn(actors.Get(carol.Character.Id))
	assert.Equal(t, alice.Character.Id, bob.Character.Aggro.Target, "Bob keeps his target")
	assert.Equal(t, characters.NoShot, bob.Character.Aggro.Shot, "a shot from someone he is not fighting earns nothing")
}

// refusesAll is the active ruleset with every attack refused.
type refusesAll struct{ rules.Ruleset }

func (refusesAll) CanAttack(attacker, defender *actors.Actor) error {
	return errors.New("Fighting is not allowed here.")
}

// neverEscapes is the active ruleset with every flee blocked.
type neverEscapes struct{ rules.Ruleset }

func (neverEscapes) FleeEscapes(fleer, blocker *actors.Actor) bool { return false }

// Phase 5 slice 4.1 (plan Q15): a blocked flee resumes the fight with the
// Since it had, so the fleer keeps its place.
func TestFailedFleeKeepsSince(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	aggroOn(alice.Character, guard.Character.Id, 5)
	aggroOn(&guard.Character, alice.Character.Id, 5)
	t.Cleanup(rules.SetForTest(neverEscapes{rules.Get()}))

	_, err := usercommands.Flee("", a, rooms.LoadRoom(goldenRoomA), 0)
	require.NoError(t, err)
	require.Equal(t, characters.Flee, alice.Character.Aggro.Type)
	flee(a, rooms.LoadRoom(goldenRoomA))

	require.NotNil(t, alice.Character.Aggro)
	assert.Equal(t, guard.Character.Id, alice.Character.Aggro.Target, "the fight resumes")
	assert.Equal(t, uint64(5), alice.Character.Aggro.Since, "in its place")
}

// Phase 5 slice 4.1 (plan review round 2): before the turn attacks its
// target, in the room or through an exit, the ruleset must still allow it
// (the fight may have moved into a room where PvP is off). Refused, the body
// switches to an attacker in its room, or its aggro ends with the refusal.
func TestSettleTargetAsksCanAttackBeforeAttacking(t *testing.T) {
	setup := func(t *testing.T) (*combatWorld, *actors.Actor, *rooms.Room) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		t.Cleanup(rules.SetForTest(refusesAll{rules.Get()}))
		return w, actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	}

	t.Run("in the room, refused, nobody else: ends with the refusal", func(t *testing.T) {
		w, a, room := setup(t)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		a.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Fighting is not allowed here.`}, w.told(9301))
	})

	t.Run("through an exit, refused: ends with the refusal", func(t *testing.T) {
		w, a, room := setup(t)
		bob := w.addPlayer(9302, "Bob", goldenRoomB, 10, 10, 10, 30)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		a.Character.Aggro.Shot = characters.TypedShot // a shot to fire through the exit
		w.told(9301)
		assert.Nil(t, settleTarget(a, room))
		assert.Nil(t, a.Character.Aggro)
		assert.Equal(t, []string{`Fighting is not allowed here.`}, w.told(9301))
	})

	t.Run("in the room, refused, an attacker it may fight here: switches", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
		t.Cleanup(rules.SetForTest(refuses{Ruleset: rules.Get(), target: bob.Character.Id}))
		a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
		aggroOn(&cat.Character, a.Id, 1)
		a.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		d := settleTarget(a, room)
		require.NotNil(t, d)
		assert.Equal(t, cat.Character.Id, d.Id)
	})
}

// refuses is the active ruleset refusing every attack on one body.
type refuses struct {
	rules.Ruleset
	target ids.ID
}

func (r refuses) CanAttack(attacker, defender *actors.Actor) error {
	if defender.Id == r.target {
		return errors.New("Fighting is not allowed here.")
	}
	return r.Ruleset.CanAttack(attacker, defender)
}

// Phase 5 slice 4.1 (plan review round 3, ruling 10): two players carrying
// aggros on each other into a room where PvP is off stop fighting; neither
// switches to the other, since the switch also asks CanAttack.
func TestPlayersWithMutualAggrosStopWherePvPIsOff(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
	setConfig(t,
		map[string]any{"GamePlay.PVP.Enabled": "limited", "GamePlay.PVP.MinimumLevel": 0},
		map[string]any{
			"GamePlay.PVP.Enabled":      string(configs.GetPVPConfig().Enabled),
			"GamePlay.PVP.MinimumLevel": int(configs.GetPVPConfig().MinimumLevel),
		})
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	aggroOn(alice.Character, bob.Character.Id, 1)
	aggroOn(bob.Character, alice.Character.Id, 1)
	w.told(9301)
	assert.Nil(t, settleTarget(a, room), "no swing at Bob, and no switch to him")
	assert.Nil(t, a.Character.Aggro)
	assert.Equal(t, []string{`This is not a PVP area.`}, w.told(9301))
}

// noCombatBuff is a test buff whose flag takes a body's swing.
const noCombatBuff = 93903

// Phase 5 slice 4.1 (plan review round 3): a body whose swing is taken
// (no-combat, an NPC's combat command) still settles its target every
// round, so its wait runs out at the limit.
func TestTakenTurnsStillCountTheWait(t *testing.T) {
	t.Run("no-combat", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
			BuffId: noCombatBuff, Name: "test calm", TriggerRate: "1 round", TriggerCount: 1000,
			Flags: []string{"no-combat"},
		}))
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		require.NoError(t, alice.Character.AddBuff(noCombatBuff, true))
		require.True(t, alice.Character.HasBuffFlag("no-combat"))
		a := actors.Get(alice.Character.Id)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		for turn := 1; turn <= 5; turn++ {
			require.NotNil(t, a.Character.Aggro, "turn %d", turn)
			combatTurn(a)
		}
		assert.Nil(t, a.Character.Aggro, "a no-combat body loses track at the limit")
	})

	t.Run("an NPC's combat command", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 30)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		dog.CombatCommands = []string{"emote growls."}
		dog.ActivityLevel = 100 // the combat command takes every turn it can
		d := actors.Get(dog.Character.Id)
		d.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		for turn := 1; turn <= 5; turn++ {
			require.NotNil(t, d.Character.Aggro, "turn %d", turn)
			combatTurn(d)
		}
		assert.Nil(t, d.Character.Aggro, "an NPC whose combat commands take its turns still loses track")
	})
}

// Phase 5 slice 4.1 (plan review round 3): a switch re-arms the aggro as
// attack <name> does, so a weapon with a delay waits its rounds before
// swinging at the new opponent.
func TestSwitchKeepsTheWeaponDelay(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	alice.Character.Equipment.Weapon = weapon(93104, "testmace", items.Bludgeoning, 1, 4, 1)
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	aggroOn(&cat.Character, a.Id, 1)
	a.Character.SetAggro(ids.New(ids.NoCell), characters.DefaultAttack) // a target that is gone

	combatTurn(a)
	require.NotNil(t, a.Character.Aggro)
	assert.Equal(t, cat.Character.Id, a.Character.Aggro.Target, "switched")
	assert.Equal(t, 0, a.Character.Aggro.RoundsWaiting, "the mace's one round of delay was spent this turn")
	_, swung := cat.Character.PlayerDamage[9301]
	assert.False(t, swung, "no swing at the cat this turn")
}

// Phase 5 slice 4.1 (plan review round 2): a tie on Since goes to turn
// order as the round builds it (players by UserId, then NPCs by MobId),
// not to the order the bodies came into the room.
func TestNextOpponentTieGoesToTurnOrderNotRoomOrder(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93205, "rat", 10, goldenRoomA, 5, 5, 5, 30) // in the room before the cat
	cat := w.addNPC(93204, "cat", 10, goldenRoomA, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	aggroOn(&rat.Character, alice.Character.Id, 9)
	aggroOn(&cat.Character, alice.Character.Id, 9)
	require.Equal(t, "cat", death.NextOpponent(a).Character.Name, "NPCs by MobId")

	carol := w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 30) // in the room before Bob
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
	aggroOn(carol.Character, alice.Character.Id, 9)
	aggroOn(bob.Character, alice.Character.Id, 9)
	assert.Equal(t, "Bob", death.NextOpponent(a).Character.Name, "players by UserId, before NPCs")
}

// Phase 5 slice 4.1 (ruling 12): a wait ends on its 5th turn in a row out of
// reach (the default limit); the body loses track of its target and those
// in its room see it give up.
func TestWaitRunsOutAtTheLimit(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.told(9301)

	for turn := 1; turn <= 4; turn++ {
		assert.Nil(t, settleTarget(a, room))
		require.NotNil(t, a.Character.Aggro, "turn %d: still waiting", turn)
	}
	assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "told once")
	assert.Nil(t, settleTarget(a, room))
	assert.Nil(t, a.Character.Aggro, "the 5th turn out of reach ends the wait")
	events.ProcessEvents()
	reads := map[int][]string{}
	for _, e := range w.rec.take() {
		reads[e.userId] = append(reads[e.userId], strings.TrimSuffix(e.text, "\n"))
	}
	assert.Equal(t, []string{`You lose track of <ansi fg="mobname">rat</ansi>.`}, reads[9301])
	assert.Equal(t, []string{`<ansi fg="username">Alice</ansi> mumbles about losing their quarry.`}, reads[9303])
}

// Phase 5 slice 4.1 (ruling 12): any turn that finds the target reachable,
// in the room or through an exit (shooting), resets the count.
func TestWaitCountResetsWhenTheTargetIsReachable(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	wait := func(turns int) {
		for i := 0; i < turns; i++ {
			require.NotNil(t, a.Character.Aggro, "the wait has not run out")
			require.Nil(t, settleTarget(a, room))
		}
	}

	wait(4)
	rat.Character.RoomId = goldenRoomA
	require.NotNil(t, settleTarget(a, room), "in the room: reachable")
	rat.Character.RoomId = goldenRoomB
	wait(4)
	require.NotNil(t, a.Character.Aggro, "the count started again: 4 of 5")

	a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	a.Character.Aggro.Shot = characters.TypedShot
	require.NotNil(t, settleTarget(a, room), "through the exit: reachable by shooting")
	a.Character.Equipment.Weapon = items.Item{} // the held shot is dropped on the next turn
	wait(4)
	assert.NotNil(t, a.Character.Aggro, "shooting reset the count too: 4 of 5")
}

// Phase 5 slice 4.1 (ruling 12, replacing behaviour.loseLostQuarry): an
// NPC whose player target left the game ends its fight at once (step 1);
// one waiting on an NPC in another room ends at the limit like anyone.
func TestNPCWaitsEndForOfflinePlayersAndNPCs(t *testing.T) {
	w := newCombatWorld(t, 1)
	bob := w.addPlayer(9302, "Bob", goldenRoomB, 10, 10, 10, 30)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 5, 5, 5, 30)
	d, room := actors.Get(dog.Character.Id), rooms.LoadRoom(goldenRoomA)

	d.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	actors.Remove(bob.Character.Id) // Bob leaves the game
	assert.Nil(t, settleTarget(d, room))
	assert.Nil(t, d.Character.Aggro, "an offline player target ends the fight at once")

	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	d.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	for turn := 1; turn <= 4; turn++ {
		require.Nil(t, settleTarget(d, room))
		require.NotNil(t, d.Character.Aggro, "turn %d", turn)
	}
	assert.Nil(t, settleTarget(d, room))
	assert.Nil(t, d.Character.Aggro, "an NPC target out of reach ends at the limit too")
}

// Phase 5 slice 4.1 (plan review round 4; ruling 13): a typed shot fires
// once. Its round re-arms through the same exit with one AggroChanged for
// the shooter; the rounds after it the aggro waits, firing nothing, queuing
// nothing and telling nothing (the shot set the told mark). Typing shoot
// again fires again.
func TestOneShotPerTypedShoot(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 200)
	typeShoot(alice.Character, "north", rat.Character.Id)
	events.ProcessEvents()
	w.rec.take()
	round := func(n int, wantShot bool) {
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		changed, read := 0, 0
		for _, e := range w.rec.take() {
			if e.userId == 0 && strings.HasPrefix(e.text, "AggroChanged Alice ") {
				changed++
			}
			if e.userId == 9301 {
				read++
			}
		}
		require.NotNil(t, alice.Character.Aggro, "round %d", n)
		assert.Equal(t, "north", alice.Character.Aggro.ExitName, "round %d: the exit is kept", n)
		if wantShot {
			assert.Equal(t, 1, changed, "round %d: one AggroChanged for the shot", n)
			assert.NotZero(t, read, "round %d: Alice reads her shot", n)
		} else {
			assert.Equal(t, 0, changed, "round %d: no second shot", n)
			assert.Zero(t, read, "round %d: no second shot, and no out-of-reach line after a shot", n)
		}
	}
	round(1, true)
	round(2, false)
	round(3, false)
	typeShoot(alice.Character, "north", rat.Character.Id)
	round(4, true)
	round(5, false)
}

// Phase 5 slice 4.1 (plan review round 4): a target in reach that the body
// cannot perceive counts toward the wait limit. A hidden attacker who stops
// attacking is lost at the limit; one who keeps hitting resets the count with
// each attack (plan review round 6), so the body never loses track of it.
func TestUnseenTargetCountsTowardTheWait(t *testing.T) {
	t.Run("the hidden attacker stops", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 60)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 60)
		w.hide(alice.Character)
		b := actors.Get(bob.Character.Id)
		b.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
		for turn := 1; turn <= 5; turn++ {
			require.NotNil(t, b.Character.Aggro, "turn %d", turn)
			combatTurn(b)
		}
		assert.Nil(t, b.Character.Aggro, "Bob loses track of hidden Alice at the limit")
	})

	t.Run("the hidden attacker keeps hitting", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 1, 20, 60)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 500)
		w.hide(alice.Character)
		alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
		w.told(9302)
		for round := 1; round <= 12; round++ {
			DoCombat(events.NewRound{})
			events.ProcessEvents()
			require.NotNil(t, bob.Character.Aggro, "round %d: each attack resets Bob's count", round)
			assert.Equal(t, alice.Character.Id, bob.Character.Aggro.Target)
		}
		for _, line := range w.told(9302) {
			assert.NotContains(t, line, "You lose track of", "Bob never loses track while she attacks")
		}
	})
}

// Phase 5 slice 4.1 (plan review round 6): being attacked by its target, hit
// or miss, resets a body's wait count. A body under steady fire from the next
// room (its shooter typing shoot every round, ruling 13) keeps its aggro;
// once the fire stops it loses track at the limit.
func TestBeingAttackedByItsTargetResetsTheWait(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 1, 20, 60)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	dan := w.addPlayer(9304, "Dan", goldenRoomB, 10, 10, 10, 500)
	w.told(9304)
	for round := 1; round <= 12; round++ {
		typeShoot(alice.Character, "north", dan.Character.Id) // she types shoot every round
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		require.NotNil(t, dan.Character.Aggro, "round %d: Dan, shot every round, keeps his aggro", round)
		assert.Equal(t, alice.Character.Id, dan.Character.Aggro.Target)
	}
	outOfReach := 0
	for _, line := range w.told(9304) {
		assert.NotContains(t, line, "You lose track of", "steady fire never runs the wait out")
		if line == `Your target is out of reach.` {
			outOfReach++
		}
	}
	assert.Equal(t, 1, outOfReach, "one out-of-reach line for the whole wait, though each shot resets the count")

	// Alice's fire stops; Bob, whom Dan is not fighting, shoots him from
	// Alice's room instead, which does not reset Dan's wait on Alice. Dan's
	// turn in the round of Alice's last shot was his 1st out of reach (she
	// acts first), so the 4th round after it is his 5th.
	alice.Character.Aggro = nil
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 30, 1, 20, 60)
	bob.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 2, 0)
	for round := 1; round <= 3; round++ {
		typeShoot(bob.Character, "north", dan.Character.Id)
		DoCombat(events.NewRound{})
		events.ProcessEvents()
		require.NotNil(t, dan.Character.Aggro, "round %d after the fire stopped: still waiting", round)
	}
	DoCombat(events.NewRound{})
	events.ProcessEvents()
	assert.Nil(t, dan.Character.Aggro, "his 5th turn out of reach ends the wait")
	assert.Contains(t, w.told(9304), `You lose track of <ansi fg="username">Alice</ansi>.`)
}

// Phase 5 slice 4.1 (plan review round 6): "Your target is out of reach."
// comes once per wait. A turn that finds the target reachable ends the wait,
// so the next time it is out of reach the body is told again.
func TestOutOfReachLineOncePerWait(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a := actors.Get(alice.Character.Id)
	room := rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.told(9301)

	assert.Nil(t, settleTarget(a, room))
	assert.Nil(t, settleTarget(a, room))
	assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "the first wait: told once")

	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomA))
	require.NotNil(t, settleTarget(a, room), "reachable: the wait ends")
	require.NoError(t, rooms.MoveToRoom(rat.Character.Id, goldenRoomB))
	assert.Nil(t, settleTarget(a, room))
	assert.Equal(t, []string{`Your target is out of reach.`}, w.told(9301), "a new wait: told again")
}

// Phase 5 slice 4.1 (plan review round 5): a shot goes through any exit the
// shoot command could use, temporary and mutator exits included, and a
// mutator's lock on an exit stops it.
func TestShootingExitUsesEveryExitTheRoomHas(t *testing.T) {
	const mutId = "test shooting exit"
	setup := func(t *testing.T) (*actors.Actor, *rooms.Room) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
		a := actors.Get(alice.Character.Id)
		a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		a.Character.Aggro.Shot = characters.TypedShot // a shot to fire through some exit
		return a, rooms.LoadRoom(goldenRoomA)
	}
	mutator := func(t *testing.T, room *rooms.Room, exits map[string]exit.RoomExit) {
		t.Cleanup(mutators.SetTestSpec(&mutators.MutatorSpec{MutatorId: mutId, Exits: exits}))
		require.True(t, room.Mutators.Add(mutId))
	}

	t.Run("a temporary exit", func(t *testing.T) {
		a, room := setup(t)
		delete(room.Exits, "north")
		room.ExitsTemp = map[string]exit.TemporaryRoomExit{
			"shimmering portal": {RoomId: goldenRoomB, Title: "shimmering portal"},
		}
		require.NotNil(t, settleTarget(a, room), "shoots through the portal")
		assert.Equal(t, "shimmering portal", a.Character.Aggro.ExitName)
	})

	t.Run("a mutator's exit", func(t *testing.T) {
		a, room := setup(t)
		delete(room.Exits, "north")
		mutator(t, room, map[string]exit.RoomExit{"trapdoor": {RoomId: goldenRoomB}})
		require.NotNil(t, settleTarget(a, room), "shoots through the trapdoor")
		assert.Equal(t, "trapdoor", a.Character.Aggro.ExitName)
	})

	t.Run("a mutator locks the exit", func(t *testing.T) {
		a, room := setup(t)
		mutator(t, room, map[string]exit.RoomExit{
			"north": {RoomId: goldenRoomB, Lock: gamelock.Lock{Difficulty: 5}},
		})
		assert.Nil(t, settleTarget(a, room), "the locked exit stops the shot: waits")
		require.NotNil(t, a.Character.Aggro)
		assert.Empty(t, a.Character.Aggro.ExitName)
	})
}

// Phase 5 slice 4.1 (plan review round 5, step 4): a body whose target is in
// the room but unseen switches to an attacker here it perceives; with no
// such attacker it keeps the target and the turn counts toward the wait.
func TestUnseenTargetGivesWayToASeenAttacker(t *testing.T) {
	setup := func(t *testing.T) (*combatWorld, *actors.Actor, *mobs.Mob) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		w.hide(alice.Character)
		b := actors.Get(bob.Character.Id)
		b.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
		rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
		return w, b, rat
	}

	t.Run("a perceived attacker here: switches", func(t *testing.T) {
		_, b, rat := setup(t)
		aggroOn(&rat.Character, b.Id, 1)
		d := settleTarget(b, rooms.LoadRoom(goldenRoomA))
		require.NotNil(t, d)
		assert.Equal(t, rat.Character.Id, d.Id, "the turn swings at the rat")
		assert.Equal(t, rat.Character.Id, b.Character.Aggro.Target)
	})

	t.Run("a seen target keeps the turn over another attacker", func(t *testing.T) {
		w := newCombatWorld(t, 1)
		alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
		bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 30)
		b := actors.Get(bob.Character.Id)
		b.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
		rat := w.addNPC(93201, "rat", 10, goldenRoomA, 5, 5, 5, 30)
		aggroOn(&rat.Character, b.Id, 1)
		d := settleTarget(b, rooms.LoadRoom(goldenRoomA))
		require.NotNil(t, d)
		assert.Equal(t, alice.Character.Id, d.Id, "no switch from a target in reach")
	})

	t.Run("nobody attacks: keeps the target and waits", func(t *testing.T) {
		_, b, _ := setup(t)
		target := b.Character.Aggro.Target
		d := settleTarget(b, rooms.LoadRoom(goldenRoomA))
		require.NotNil(t, d, "the turn goes at the unseen target")
		assert.Equal(t, target, d.Id)
		require.NotNil(t, b.Character.Aggro)
		assert.Equal(t, target, b.Character.Aggro.Target)
		assert.Equal(t, 1, b.Character.Aggro.OutOfReach, "the unseen turn counts")
	})
}

// Phase 5 slice 4.1 (spec "Death"; blind review round 7): a death releases
// every body targeting the victim, a body waiting on it from another room
// too, so a dead player respawned in the recovery room is never waited on.
func TestDeathReleasesAnAttackerInAnotherRoom(t *testing.T) {
	w := newCombatWorld(t, 1)
	w.recoveryRoom()
	deathConfig(t)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 60)
	bob := w.addPlayer(9302, "Bob", goldenRoomB, 1, 1, 1, 1)
	carol := w.addPlayer(9303, "Carol", goldenRoomB, 30, 30, 20, 60)
	carol.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 4, 6, 0)
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack) // she waits on Bob from room A
	carol.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	dead := func() bool { return rooms.IsEphemeralRoomId(bob.Character.RoomId) }
	for round := 1; round <= 4 && !dead(); round++ {
		require.NotNil(t, alice.Character.Aggro, "round %d: she still waits on him", round)
		DoCombat(events.NewRound{})
		events.ProcessEvents()
	}
	require.True(t, dead(), "the fixture must reach Bob's death within her wait limit")
	assert.Nil(t, alice.Character.Aggro, "his death ends her wait, though she is in another room")
}

// Phase 5 slice 4.1 (ruling 12; blind review round 7): the wait limit is
// GamePlay.OutOfReachRounds, not a fixed 5.
func TestWaitLimitComesFromTheConfig(t *testing.T) {
	w := newCombatWorld(t, 1)
	setConfig(t,
		map[string]any{"GamePlay.OutOfReachRounds": 2},
		map[string]any{"GamePlay.OutOfReachRounds": int(configs.GetGamePlayConfig().OutOfReachRounds)})
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	assert.Nil(t, settleTarget(a, room))
	require.NotNil(t, a.Character.Aggro, "1 turn of 2")
	assert.Nil(t, settleTarget(a, room))
	assert.Nil(t, a.Character.Aggro, "a limit of 2 ends the wait on its 2nd turn")
}

// Phase 5 slice 4.1 (plan review round 5; blind review round 7): with two
// exits to the target's room, a shot goes through the aggro's own exit when
// it still leads there, otherwise through the first by name.
func TestShootingExitPrefersItsOwnExitThenTheFirstByName(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	for _, name := range []string{"east", "up", "west", "down", "out", "gate"} {
		room.Exits[name] = exit.RoomExit{RoomId: goldenRoomB} // north leads there too
	}
	for i := 0; i < 20; i++ { // map order differs between calls
		a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
		a.Character.Aggro.Shot = characters.TypedShot
		require.NotNil(t, settleTarget(a, room))
		require.Equal(t, "down", a.Character.Aggro.ExitName, "no exit of its own: the first by name")
	}
	typeShoot(a.Character, "north", rat.Character.Id)
	require.NotNil(t, settleTarget(a, room))
	assert.Equal(t, "north", a.Character.Aggro.ExitName, "its own exit still leads there: kept")
}

// Phase 5 slice 4.1 (ruling 12; blind review round 7): a body nobody in its
// room perceives loses track unseen: they read no mumble line.
func TestLoseTrackIsNotSeenByThoseWhoCannotPerceiveTheBody(t *testing.T) {
	w := newCombatWorld(t, 1)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 30)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 30)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 5, 5, 5, 30)
	w.hide(alice.Character)
	a, room := actors.Get(alice.Character.Id), rooms.LoadRoom(goldenRoomA)
	a.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.told(9301)
	for turn := 1; turn <= 5; turn++ {
		assert.Nil(t, settleTarget(a, room))
	}
	require.Nil(t, a.Character.Aggro, "she lost track at the limit")
	events.ProcessEvents()
	reads := map[int][]string{}
	for _, e := range w.rec.take() {
		reads[e.userId] = append(reads[e.userId], strings.TrimSuffix(e.text, "\n"))
	}
	assert.Contains(t, reads[9301], `You lose track of <ansi fg="mobname">rat</ansi>.`)
	assert.Empty(t, reads[9303], "Carol cannot perceive hidden Alice: no mumble line")
}

// Phase 5 slice 4.1 (spec "What stays as today": a body with no-combat takes
// no turn; blind review round 7): no-combat still stops a flee, now that the
// lookup runs before it.
func TestNoCombatStillStopsAFlee(t *testing.T) {
	w := newCombatWorld(t, 1)
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: noCombatBuff, Name: "test calm", TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{"no-combat"},
	}))
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 100, 10, 10, 30)
	require.NoError(t, alice.Character.AddBuff(noCombatBuff, true))
	fleeAggro(alice.Character)
	a := actors.Get(alice.Character.Id)

	combatTurn(a)
	assert.Equal(t, goldenRoomA, alice.Character.RoomId, "no-combat: no flee")
	require.NotNil(t, alice.Character.Aggro)
	assert.Equal(t, characters.Flee, alice.Character.Aggro.Type, "the flee stays pending")

	alice.Character.RemoveBuff(noCombatBuff)
	combatTurn(a)
	assert.Equal(t, goldenRoomB, alice.Character.RoomId, "control: without it she flees")
}
