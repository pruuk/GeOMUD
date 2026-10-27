package actors

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	m := &mobs.Mob{InstanceId: 1}
	n := Spawn(&m.Character, NewBehaviourController(m))

	assert.NotEqual(t, p.Id, n.Id)
	assert.Same(t, p, ByUserId(1))
	assert.Same(t, n, ByMobInstanceId(1))
	assert.Nil(t, n.Controller().Account())
	assert.Same(t, &m.Character, n.Character)
}

func TestSetControllerKeepsTheId(t *testing.T) {
	Reset()
	m := &mobs.Mob{InstanceId: 3}
	a := Spawn(&m.Character, NewBehaviourController(m))
	id := a.Id
	acc := &accounts.Account{UserId: 9}

	require.NoError(t, SetController(id, NewConnectionController(acc)))

	assert.Equal(t, id, a.Id)
	assert.Same(t, a, ByUserId(9))
	assert.Nil(t, ByMobInstanceId(3))
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
