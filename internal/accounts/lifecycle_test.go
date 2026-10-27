package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/stretchr/testify/assert"
)

func TestActivateRecordsAndFires(t *testing.T) {
	resetUserManager()
	var got []*Account
	OnActivate.Register(func(u *Account) *Account { got = append(got, u); return u })

	u := NewAccount(21, 0)
	u.Username = "hooked"
	userManagerMu.Lock()
	activateLocked(u, connections.ConnectionId(9))
	userManagerMu.Unlock()
	activated(u)

	assert.Same(t, u, userManager.Users[21])
	assert.Equal(t, 21, userManager.Usernames["hooked"])
	assert.Equal(t, 21, userManager.Connections[9])
	assert.Equal(t, connections.ConnectionId(9), userManager.UserConnections[21])
	assert.Equal(t, uint64(9), u.ConnectionId())
	assert.Contains(t, got, u)
}

func TestDeactivatedFires(t *testing.T) {
	var got []*Account
	OnDeactivate.Register(func(u *Account) *Account { got = append(got, u); return u })
	u := NewAccount(22, 0)
	deactivated(u)
	assert.Contains(t, got, u)
}
