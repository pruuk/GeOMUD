package accounts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountHasSession(t *testing.T) {
	u := NewAccount(5, 77)
	require.NotNil(t, u.session)
	assert.Equal(t, uint64(77), u.ConnectionId())
}

func TestLoadedAccountGetsSessionOnDemand(t *testing.T) {
	u := &Account{UserId: 3} // as decoded from YAML: no session
	assert.Equal(t, uint64(0), u.ConnectionId())
	u.BlockInput()
	assert.True(t, u.InputBlocked())
}

func TestUpdateOnlineUserKeepsSessionAndCharacterPointer(t *testing.T) {
	resetUserManager()
	live := NewAccount(10, 42)
	live.Username = "live"
	live.Character.Name = "Before"
	live.SetUnsentText("typing")
	userManager.Users[10] = live
	char := live.Character

	edited := *live
	editedChar := *live.Character
	editedChar.Name = "After"
	edited.Character = &editedChar
	edited.session = nil // what the web API decodes: no session
	edited.Role = RoleMod

	UpdateOnlineUser(edited)

	got := userManager.Users[10]
	assert.Same(t, char, got.Character, "the live character pointer must survive an edit")
	assert.Equal(t, "After", got.Character.Name)
	assert.Equal(t, RoleMod, got.Role)
	assert.Equal(t, uint64(42), got.ConnectionId())
	assert.Equal(t, "typing", got.GetUnsentText())
}
