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

func TestSpawnRefusesABodyWithoutAnId(t *testing.T) {
	Reset()
	assert.Nil(t, Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 93})))
	assert.Empty(t, All())
}
