package messaging

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recorder is a controller that keeps what it perceives. With an account it
// stands for a player (names render as username), without one for an NPC.
type recorder struct {
	acc *accounts.Account
	got []perception.Perception
}

func (r *recorder) Account() *accounts.Account       { return r.acc }
func (r *recorder) Perceive(p perception.Perception) { r.got = append(r.got, p) }

// world builds rooms in memory and bodies in them for one test.
type world struct {
	t     *testing.T
	rooms []int
	flags map[ids.ID]map[string]bool
	light map[int]int
}

func newWorld(t *testing.T) *world {
	t.Helper()
	actors.Reset()
	w := &world{t: t, flags: map[ids.ID]map[string]bool{}, light: map[int]int{}}
	oldFlag, oldLight := hasFlag, roomVisibility
	hasFlag = func(c *characters.Character, flag string) bool { return w.flags[c.Id][flag] }
	roomVisibility = func(r *rooms.Room) int {
		if v, ok := w.light[r.RoomId]; ok {
			return v
		}
		return 2
	}
	t.Cleanup(func() {
		hasFlag, roomVisibility = oldFlag, oldLight
		for _, id := range w.rooms {
			rooms.RemoveTestRoom(id)
		}
		actors.Reset()
	})
	return w
}

// room adds a room in zone; exits maps an exit name to the room it leads to.
func (w *world) room(id int, zone string, exits map[string]int) *rooms.Room {
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone = id, zone
	r.Exits = map[string]exit.RoomExit{}
	for name, to := range exits {
		r.Exits[name] = exit.RoomExit{RoomId: to}
	}
	rooms.AddTestRoom(r)
	w.rooms = append(w.rooms, id)
	return r
}

// body puts a named body in r. player says whether an account drives it.
func (w *world) body(r *rooms.Room, name string, player bool) (ids.ID, *recorder) {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = name
	c.RoomId = r.RoomId
	rec := &recorder{}
	if player {
		rec.acc = &accounts.Account{UserId: len(w.flags) + 100, Character: c}
	}
	require.NotNil(w.t, actors.Spawn(c, rec))
	r.AddTestBody(c.Id)
	w.flags[c.Id] = map[string]bool{}
	return c.Id, rec
}

func texts(r *recorder) []string {
	out := []string{}
	for _, p := range r.got {
		out = append(out, p.Text)
	}
	return out
}

func say(actor ids.ID) perception.Message {
	return perception.Message{
		Actor: actor, Senses: perception.Sound, Range: perception.Room,
		ActorText:    `You say, "hi"`,
		ObserverText: `{actor} says, "hi"`,
	}
}

func TestRoomRangeRolesAndNames(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, fredRec := w.body(square, "Fred", true)
	steve, steveRec := w.body(square, "Steve", true)
	_, ratRec := w.body(square, "rat", false)

	Send(perception.Message{
		Actor: fred, Target: steve, Senses: perception.Sight | perception.Sound, Range: perception.Room,
		ActorText:    `You attack {target}!`,
		TargetText:   `{actor} attacks you!`,
		ObserverText: `{actor} attacks {target}!`,
	})

	assert.Equal(t, []string{`You attack <ansi fg="username">Steve</ansi>!`}, texts(fredRec))
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> attacks you!`}, texts(steveRec))
	require.Len(t, ratRec.got, 1)
	assert.Equal(t, perception.Observer, ratRec.got[0].Role)
	assert.Equal(t, perception.Sight, ratRec.got[0].SenseUsed, "sight is tried first")
	assert.Equal(t, perception.Actor, fredRec.got[0].Role)
	assert.Equal(t, perception.Target, steveRec.got[0].Role)
}

func TestNPCNamesRenderAsMobname(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	rat, _ := w.body(square, "rat", false)
	_, fredRec := w.body(square, "Fred", true)

	Send(say(rat))
	assert.Equal(t, []string{`<ansi fg="mobname">rat</ansi> says, "hi"`}, texts(fredRec))
}

func TestTargetFallsBackToObserverText(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, _ := w.body(square, "Fred", true)
	steve, steveRec := w.body(square, "Steve", true)

	m := say(fred)
	m.Target = steve
	Send(m)
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> says, "hi"`}, texts(steveRec))
}

func TestDarkRoomBlocksSight(t *testing.T) {
	w := newWorld(t)
	cave := w.room(1, "caves", nil)
	w.light[1] = 0
	fred, _ := w.body(cave, "Fred", true)
	_, steveRec := w.body(cave, "Steve", true)
	owl, owlRec := w.body(cave, "Owl", true)
	w.flags[owl]["nightvision"] = true

	Send(perception.Message{Actor: fred, Senses: perception.Sight, Range: perception.Room,
		ActorText: `You wave.`, ObserverText: `{actor} waves.`})

	assert.Empty(t, steveRec.got, "a sight-only message does not reach a receiver in the dark")
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> waves.`}, texts(owlRec), "nightvision sees")
}

func TestSomeoneWhenTheActorIsUnseen(t *testing.T) {
	w := newWorld(t)
	cave := w.room(1, "caves", nil)
	w.light[1] = 0
	fred, fredRec := w.body(cave, "Fred", true)
	_, steveRec := w.body(cave, "Steve", true)

	Send(say(fred)) // heard in the dark, speaker unseen
	assert.Equal(t, []string{`Someone says, "hi"`}, texts(steveRec))
	assert.False(t, steveRec.got[0].ActorSeen)
	assert.Equal(t, perception.Sound, steveRec.got[0].SenseUsed)
	assert.Equal(t, []string{`You say, "hi"`}, texts(fredRec), "the actor always gets its text")

	square := w.room(2, "town", nil)
	sneak, _ := w.body(square, "Sneak", true)
	w.flags[sneak]["hidden"] = true
	_, bobRec := w.body(square, "Bob", true)
	m := say(sneak)
	m.ObserverText = `You hear {actor} whisper.`
	Send(m)
	assert.Equal(t, []string{`You hear someone whisper.`}, texts(bobRec), "lower case inside a sentence")
}

func TestFaintNeedsSuperhearing(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, _ := w.body(square, "Fred", true)
	_, steveRec := w.body(square, "Steve", true)
	keen, keenRec := w.body(square, "Keen", true)
	w.flags[keen]["superhearing"] = true

	m := say(fred)
	m.Faint = true
	Send(m)
	assert.Empty(t, steveRec.got)
	assert.Len(t, keenRec.got, 1)
}

func TestNearbyReachesNeighboursWithTheirDirection(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", map[string]int{"north": 2, "east": 3})
	w.room(2, "town", map[string]int{"south": 1})
	w.room(3, "town", nil)                              // no exit back: not reached, as shout today
	far := w.room(4, "town", map[string]int{"west": 1}) // an exit to the square, but the square has none to it
	fred, _ := w.body(square, "Fred", true)
	_, nRec := w.body(roomOf(t, 2), "Nora", true)
	_, eRec := w.body(roomOf(t, 3), "Ed", true)
	_, fRec := w.body(far, "Fay", true)
	_, sqRec := w.body(square, "Sam", true)

	Send(perception.Message{Actor: fred, Senses: perception.Sound, Range: perception.Nearby,
		ActorText:    `You shout, "HI"`,
		ObserverText: `{actor} shouts, "HI"`,
		FarText:      `{actor} shouts from the {direction} direction, "HI"`})

	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> shouts, "HI"`}, texts(sqRec))
	assert.Equal(t, []string{`Someone shouts from the <ansi fg="exit">south</ansi> direction, "HI"`}, texts(nRec))
	assert.True(t, nRec.got[0].Far)
	assert.Equal(t, "south", nRec.got[0].Direction)
	assert.Empty(t, eRec.got)
	assert.Empty(t, fRec.got, "neighbours are the rooms the source room leads to")
}

func TestRoomRangeWithATargetElsewhereCoversBothRooms(t *testing.T) {
	w := newWorld(t)
	a := w.room(1, "town", map[string]int{"north": 2})
	b := w.room(2, "town", map[string]int{"south": 1})
	archer, _ := w.body(a, "Archer", true)
	_, aRec := w.body(a, "Ann", true)
	deer, deerRec := w.body(b, "deer", false)
	_, bRec := w.body(b, "Bea", true)

	Send(perception.Message{Actor: archer, Target: deer, Senses: perception.Sight, Range: perception.Room,
		ActorText: `You shoot {target}.`, TargetText: `An arrow hits you.`, ObserverText: `{actor} shoots {target}.`})

	assert.Equal(t, []string{`<ansi fg="username">Archer</ansi> shoots <ansi fg="mobname">deer</ansi>.`}, texts(aRec))
	assert.Equal(t, []string{`<ansi fg="username">Archer</ansi> shoots <ansi fg="mobname">deer</ansi>.`}, texts(bRec),
		"the target's room is a source room, so its observers are not far")
	assert.Equal(t, []string{`An arrow hits you.`}, texts(deerRec))
}

func TestZoneAndWorldRanges(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	gate := w.room(2, "town", nil)
	cave := w.room(3, "caves", nil)
	crier, _ := w.body(square, "Crier", true)
	_, gateRec := w.body(gate, "Gus", true)
	_, caveRec := w.body(cave, "Cass", true)

	m := perception.Message{Actor: crier, Senses: perception.Sound, Range: perception.Zone,
		ActorText: `You call out.`, ObserverText: `{actor} calls out.`, FarText: `A call echoes through town.`}
	Send(m)
	assert.Equal(t, []string{`A call echoes through town.`}, texts(gateRec))
	assert.Empty(t, caveRec.got)

	m.Range = perception.World
	m.FarText = ``
	Send(m)
	assert.Equal(t, []string{`Someone calls out.`}, texts(caveRec), "far receivers fall back to observer text and do not see the actor")
}

func TestActorOnlyAndInterface(t *testing.T) {
	w := newWorld(t)
	cave := w.room(1, "caves", nil)
	w.light[1] = 0
	fred, fredRec := w.body(cave, "Fred", true)
	_, steveRec := w.body(cave, "Steve", true)

	Send(perception.Message{Kind: perception.Interface, Actor: fred, Audience: perception.ActorOnly,
		Range: perception.World, ActorText: `Your score: 3`})
	assert.Equal(t, []string{`Your score: 3`}, texts(fredRec), "interface text needs no sense and no light")
	assert.Empty(t, steveRec.got)
}

func TestNamedAudience(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	other := w.room(2, "town", nil)
	fred, fredRec := w.body(square, "Fred", true)
	steve, steveRec := w.body(square, "Steve", true)
	_, samRec := w.body(square, "Sam", true)
	away, awayRec := w.body(other, "Away", true)

	Send(perception.Message{Actor: fred, Senses: perception.Sound, Range: perception.Room,
		Audience: perception.Named, Named: []ids.ID{steve, away},
		ActorText: `You whisper.`, ObserverText: `{actor} whispers to you.`})

	assert.Len(t, fredRec.got, 1)
	assert.Equal(t, []string{`<ansi fg="username">Fred</ansi> whispers to you.`}, texts(steveRec))
	assert.Empty(t, samRec.got, "not named")
	assert.Empty(t, awayRec.got, "named but out of range")
}

func TestWorldEventWithoutAnActor(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	_, fredRec := w.body(square, "Fred", true)

	Send(perception.Message{Room: 1, Senses: perception.Sound, Range: perception.Room,
		ObserverText: `A door creaks.`})
	require.Len(t, fredRec.got, 1)
	assert.Equal(t, `A door creaks.`, fredRec.got[0].Text)
	assert.True(t, fredRec.got[0].ActorSeen)
}

func TestBehaviourReceivesTheStructuredPerception(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	fred, _ := w.body(square, "Fred", true)
	_, ratRec := w.body(square, "rat", false)

	Send(say(fred))
	require.Len(t, ratRec.got, 1)
	p := ratRec.got[0]
	assert.Equal(t, perception.Observer, p.Role)
	assert.Equal(t, fred, p.Message.Actor)
	assert.Equal(t, perception.Sound, p.SenseUsed)
	assert.True(t, p.ActorSeen)
}

func TestReceiversAreDeduplicated(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", map[string]int{"north": 2})
	w.room(2, "town", map[string]int{"south": 1})
	fred, _ := w.body(square, "Fred", true)
	_, samRec := w.body(square, "Sam", true)

	Send(perception.Message{Actor: fred, Senses: perception.Sound, Range: perception.World,
		ActorText: `x`, ObserverText: `y`})
	assert.Len(t, samRec.got, 1)
}

func roomOf(t *testing.T, id int) *rooms.Room {
	t.Helper()
	r := rooms.LoadRoom(id)
	require.NotNil(t, r)
	return r
}
