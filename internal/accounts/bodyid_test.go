package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestGiveBodyIdOnlyWhenMissing(t *testing.T) {
	u := NewAccount(81, 0)
	require.True(t, u.Character.Id.IsZero())
	assert.True(t, giveBodyId(u))
	first := u.Character.Id
	assert.False(t, first.IsZero())
	assert.False(t, giveBodyId(u), "an existing ID is never replaced")
	assert.Equal(t, first, u.Character.Id)
}

func TestBodyIdSurvivesTheUserFile(t *testing.T) {
	u := NewAccount(82, 0)
	giveBodyId(u)
	b, err := yaml.Marshal(u)
	require.NoError(t, err)
	var back Account
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, u.Character.Id, back.Character.Id)
}

func TestGetByBodyIdFindsTheOnlineAccount(t *testing.T) {
	resetUserManager()
	u := NewAccount(84, 6)
	giveBodyId(u)
	userManager.Users[84] = u
	assert.Same(t, u, GetByBodyId(u.Character.Id))
	assert.Nil(t, GetByBodyId(ids.New(ids.NoCell)))
	assert.Nil(t, GetByBodyId(ids.ID{}))
}

func TestUpdateOnlineUserKeepsTheBodyId(t *testing.T) {
	resetUserManager()
	live := NewAccount(83, 5)
	giveBodyId(live)
	userManager.Users[83] = live
	edited := *live
	ec := *live.Character
	attempted := ids.New(ids.NoCell) // an admin edit trying to change the ID
	ec.Id = attempted
	edited.Character = &ec
	UpdateOnlineUser(edited)
	assert.Equal(t, live.Character.Id, userManager.Users[83].Character.Id)
	assert.NotEqual(t, attempted, userManager.Users[83].Character.Id)
}
