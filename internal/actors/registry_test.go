package actors

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bodyMob is a mob instance with a body ID, as mobs.addInstance makes one.
func bodyMob() *mobs.Mob {
	m := &mobs.Mob{}
	m.Character.Id = ids.New(ids.NoCell)
	return m
}

// bodyChar is a character with a body ID, as a saved or newly made player has.
func bodyChar() *characters.Character {
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	return c
}

func TestSpawnAndLookup(t *testing.T) {
	Reset()
	acc := &accounts.Account{UserId: 7}
	char := bodyChar()
	a := Spawn(char, NewConnectionController(acc))

	require.NotNil(t, a)
	assert.Equal(t, char.Id, a.Id)
	assert.Same(t, a, Get(a.Id))
	assert.Same(t, a, ByUserId(7))
	assert.Same(t, char, a.Character)
	assert.Same(t, acc, a.Controller().Account())
	assert.Len(t, All(), 1)
}

func TestOneIdSpaceForPlayersAndNPCs(t *testing.T) {
	Reset()
	p := Spawn(bodyChar(), NewConnectionController(&accounts.Account{UserId: 1}))
	m := bodyMob()
	n := Spawn(&m.Character, NewBehaviourController(m))

	assert.NotEqual(t, p.Id, n.Id)
	assert.Same(t, p, ByUserId(1))
	assert.Same(t, n, Get(m.Character.Id))
	assert.Nil(t, n.Controller().Account())
	assert.Same(t, &m.Character, n.Character)
}

func TestAccountOfNamesTheDriver(t *testing.T) {
	Reset()
	acc := &accounts.Account{UserId: 3}
	p := Spawn(bodyChar(), NewConnectionController(acc))
	m := bodyMob()
	Spawn(&m.Character, NewBehaviourController(m))

	assert.Same(t, acc, AccountOf(p.Id))
	assert.Nil(t, AccountOf(m.Character.Id), "an NPC has no account")
	assert.Nil(t, AccountOf(ids.New(ids.NoCell)), "no actor")
	assert.Nil(t, AccountOf(ids.ID{}), "zero ID")
}

func TestSetControllerKeepsTheId(t *testing.T) {
	Reset()
	m := bodyMob()
	a := Spawn(&m.Character, NewBehaviourController(m))
	id := a.Id
	acc := &accounts.Account{UserId: 9}

	require.NoError(t, SetController(id, NewConnectionController(acc)))

	assert.Equal(t, id, a.Id)
	assert.Same(t, a, ByUserId(9))
	assert.Same(t, acc, Get(id).Controller().Account())
	assert.Error(t, SetController(ids.New(ids.NoCell), NewConnectionController(acc)))

	// Now driven by an account, it is no longer treated as an NPC.
	assert.Nil(t, unbindMob(id))
	assert.Same(t, a, Get(id))
}

func TestSpawnRefusesASecondBodyWithTheSameId(t *testing.T) {
	Reset()
	acc := &accounts.Account{UserId: 21}
	char := bodyChar()
	first := Spawn(char, NewConnectionController(acc))
	require.NotNil(t, first)

	m := &mobs.Mob{}
	m.Character.Id = char.Id // a different body claiming the same ID
	assert.Nil(t, Spawn(&m.Character, NewBehaviourController(m)))

	assert.Same(t, first, Get(char.Id))
	assert.Same(t, char, first.Character)
	assert.Same(t, acc, first.Controller().Account())
	assert.Same(t, first, ByUserId(21))
	assert.Len(t, All(), 1)
}

func TestLoginWithANewBodyIdRekeysTheActor(t *testing.T) {
	Reset()
	u := &accounts.Account{UserId: 94, Character: characters.New()}
	u.Character.Id = ids.New(ids.NoCell)
	accounts.OnActivate.Fire(u)
	old := u.Character.Id
	u.Character.Id = ids.New(ids.NoCell) // no BodyReplaced call
	accounts.OnActivate.Fire(u)
	assert.Nil(t, Get(old))
	require.NotNil(t, Get(u.Character.Id))
	assert.Same(t, u, Get(u.Character.Id).Controller().Account())
	assert.Len(t, All(), 1)
}

func TestRemove(t *testing.T) {
	Reset()
	a := Spawn(bodyChar(), NewConnectionController(&accounts.Account{UserId: 4}))
	Remove(a.Id)
	assert.Nil(t, Get(a.Id))
	assert.Nil(t, ByUserId(4))
	Remove(a.Id) // removing twice is harmless
}

func TestRemovingAnOlderActorKeepsTheNewerIndex(t *testing.T) {
	Reset()
	a := Spawn(bodyChar(), NewConnectionController(&accounts.Account{UserId: 5}))
	b := Spawn(bodyChar(), NewConnectionController(&accounts.Account{UserId: 5}))
	Remove(a.Id)
	assert.Same(t, b, ByUserId(5))
}

func TestSpawningTheSameBodyTwiceKeepsTheNewerActor(t *testing.T) {
	Reset()
	m := bodyMob()
	x := Spawn(&m.Character, NewBehaviourController(m))
	y := Spawn(&m.Character, NewBehaviourController(m))
	require.NotNil(t, x)
	require.NotNil(t, y)
	assert.Same(t, y, Get(m.Character.Id))
	assert.Len(t, All(), 1)
}

func TestControllerSwapRoundTrip(t *testing.T) {
	Reset()
	m := bodyMob()
	a := Spawn(&m.Character, NewBehaviourController(m))
	acc := &accounts.Account{UserId: 12}

	require.NoError(t, SetController(a.Id, NewConnectionController(acc)))
	require.NoError(t, SetController(a.Id, NewBehaviourController(m)))

	assert.Same(t, a, Get(m.Character.Id))
	assert.Nil(t, ByUserId(12))
	assert.Nil(t, a.Controller().Account())
}

func TestNPCActorIdIsItsBodyId(t *testing.T) {
	Reset()
	m := &mobs.Mob{}
	m.Character.Id = ids.New(ids.NoCell)
	mobs.OnInstanceCreated.Fire(m)
	a := Get(m.Character.Id)
	require.NotNil(t, a)
	assert.Equal(t, m.Character.Id, a.Id)
}

func TestPlayerKeepsItsActorIdAcrossLogins(t *testing.T) {
	Reset()
	body := ids.New(ids.NoCell)
	first := &accounts.Account{UserId: 91, Character: characters.New()}
	first.Character.Id = body
	accounts.OnActivate.Fire(first)
	accounts.OnDeactivate.Fire(first)
	again := &accounts.Account{UserId: 91, Character: characters.New()} // the next login loads the saved file
	again.Character.Id = body
	accounts.OnActivate.Fire(again)
	assert.Equal(t, body, ByUserId(91).Id)
}

func TestBodyReplacedRekeysThePlayer(t *testing.T) {
	Reset()
	u := &accounts.Account{UserId: 92, Character: characters.New()}
	u.Character.Id = ids.New(ids.NoCell)
	accounts.OnActivate.Fire(u)
	old := u.Character.Id
	u.Character.Id = ids.New(ids.NoCell)
	BodyReplaced(u, old)
	assert.Nil(t, Get(old))
	assert.Same(t, u, Get(u.Character.Id).Controller().Account())
	assert.Len(t, All(), 1)
}

func TestEachVisitsEveryActorOnce(t *testing.T) {
	Reset()
	want := map[ActorId]bool{}
	for i := 0; i < 5; i++ {
		c := bodyChar()
		Spawn(c, NewConnectionController(&accounts.Account{UserId: 200 + i}))
		want[c.Id] = true
	}
	seen := map[ActorId]int{}
	Each(func(a *Actor) { seen[a.Id]++ })
	require.Len(t, seen, len(want))
	for id := range want {
		assert.Equal(t, 1, seen[id])
	}
}

func TestSpawnRefusesABodyWithoutAnId(t *testing.T) {
	Reset()
	assert.Nil(t, Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 93})))
	assert.Empty(t, All())
}
