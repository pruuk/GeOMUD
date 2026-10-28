package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
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

func TestGetByBodyIdSkipsAnAccountWithoutACharacter(t *testing.T) {
	resetUserManager()
	userManager.Users[86] = &Account{UserId: 86} // no Character
	u := NewAccount(87, 7)
	giveBodyId(u)
	userManager.Users[87] = u
	assert.NotPanics(t, func() {
		assert.Same(t, u, GetByBodyId(u.Character.Id))
		assert.Nil(t, GetByBodyId(ids.New(ids.NoCell)))
	})
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

func TestItemIdsSurviveTheUserFile(t *testing.T) {
	u := NewAccount(88, 0)
	u.Character.Items = []items.Item{{ItemId: 1, Id: ids.New(ids.NoCell)}}
	u.Character.Equipment.Weapon = items.Item{ItemId: 2, Id: ids.New(ids.NoCell)}
	b, err := yaml.Marshal(u)
	require.NoError(t, err)
	var back Account
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, u.Character.Items[0].Id, back.Character.Items[0].Id)
	assert.Equal(t, u.Character.Equipment.Weapon.Id, back.Character.Equipment.Weapon.Id)
}
