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

func TestSpawnAndLookup(t *testing.T) {
	Reset()
	acc := &accounts.Account{UserId: 7}
	char := characters.New()
	a := Spawn(char, NewConnectionController(acc))

	require.NotNil(t, a)
	assert.NotZero(t, a.Id)
	assert.Same(t, a, Get(a.Id))
	assert.Same(t, a, ByUserId(7))
	assert.Same(t, char, a.Character)
	assert.Same(t, acc, a.Controller().Account())
	assert.Len(t, All(), 1)
}

func TestOneIdSpaceForPlayersAndNPCs(t *testing.T) {
	Reset()
	p := Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 1}))
	m := bodyMob()
	n := Spawn(&m.Character, NewBehaviourController(m))

	assert.NotEqual(t, p.Id, n.Id)
	assert.Same(t, p, ByUserId(1))
	assert.Same(t, n, ByMobInstanceId(m.Character.Id))
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
	assert.Nil(t, ByMobInstanceId(m.Character.Id))
	assert.Same(t, acc, Get(id).Controller().Account())
	assert.Error(t, SetController(id+100, NewConnectionController(acc)))
}

func TestRemove(t *testing.T) {
	Reset()
	a := Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 4}))
	Remove(a.Id)
	assert.Nil(t, Get(a.Id))
	assert.Nil(t, ByUserId(4))
	Remove(a.Id) // removing twice is harmless
}

func TestRemovingAnOlderActorKeepsTheNewerIndex(t *testing.T) {
	Reset()
	a := Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 5}))
	b := Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 5}))
	Remove(a.Id)
	assert.Same(t, b, ByUserId(5))

	m := bodyMob()
	x := Spawn(&m.Character, NewBehaviourController(m))
	y := Spawn(&m.Character, NewBehaviourController(m))
	Remove(x.Id)
	assert.Same(t, y, ByMobInstanceId(m.Character.Id))
}

func TestControllerSwapRoundTrip(t *testing.T) {
	Reset()
	m := bodyMob()
	a := Spawn(&m.Character, NewBehaviourController(m))
	acc := &accounts.Account{UserId: 12}

	require.NoError(t, SetController(a.Id, NewConnectionController(acc)))
	require.NoError(t, SetController(a.Id, NewBehaviourController(m)))

	assert.Same(t, a, ByMobInstanceId(m.Character.Id))
	assert.Nil(t, ByUserId(12))
	assert.Nil(t, a.Controller().Account())
}
