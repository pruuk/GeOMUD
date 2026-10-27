package accounts

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/copyover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyoverRoundTripKeepsOnlineAccounts(t *testing.T) {
	resetUserManager()
	onDisk := map[int]*Account{}
	restoreLoad = func(userId int) (*Account, error) {
		u, ok := onDisk[userId]
		if !ok {
			return nil, fmt.Errorf("no user %d", userId)
		}
		c := *u
		c.session = nil // as decoded from the user file
		if u.Character != nil {
			ch := *u.Character // a copy, so the test does not pass by sharing a pointer
			c.Character = &ch
		}
		return &c, nil
	}
	t.Cleanup(func() { restoreLoad = loadUserById })

	active := NewAccount(61, 0)
	active.Username = "active"
	gone := NewAccount(62, 0)
	gone.Username = "gone"
	require.True(t, giveBodyId(active))
	require.True(t, giveBodyId(gone))
	activeBodyId, goneBodyId := active.Character.Id, gone.Character.Id
	onDisk[61], onDisk[62] = active, gone
	userManagerMu.Lock()
	activateLocked(active, connections.ConnectionId(5))
	activateLocked(gone, connections.ConnectionId(6))
	gone.sess().isLinkDead = true
	userManager.LinkDeadConnections[6] = 123
	userManagerMu.Unlock()

	copyover.ResetRegistry()
	copyover.Register(CopyoverContributor())
	var buf bytes.Buffer
	require.NoError(t, copyover.Save(&buf))

	resetUserManager()
	var activatedIds []int
	OnActivate.Register(func(u *Account) *Account { activatedIds = append(activatedIds, u.UserId); return u })
	require.NoError(t, copyover.RestoreFrom(&buf))

	require.NotNil(t, GetByUserId(61))
	require.NotNil(t, GetByUserId(62))
	assert.Equal(t, activeBodyId, GetByUserId(61).Character.Id, "the body ID survives copyover")
	assert.Equal(t, goneBodyId, GetByUserId(62).Character.Id, "the body ID survives copyover")
	assert.Equal(t, uint64(5), GetByUserId(61).ConnectionId())
	assert.Equal(t, uint64(6), GetByUserId(62).ConnectionId())
	assert.Same(t, GetByUserId(61), GetByConnectionId(5))
	assert.True(t, IsLinkDeadConnection(6))
	assert.Equal(t, uint64(123), userManager.LinkDeadConnections[6])

	var activeIds []int
	for _, u := range GetAllActiveUsers() {
		activeIds = append(activeIds, u.UserId)
	}
	assert.Equal(t, []int{61}, activeIds, "a link-dead account stays out of the active list after copyover")
	assert.Subset(t, activatedIds, []int{61, 62}, "restored accounts go online through OnActivate, so they get actors")
}

// TestEverySessionFieldHasACopyoverRule makes adding a Session field a
// decision about copyover: name here what CopyoverRestore does with it, and
// make CopyoverRestore do it.
func TestEverySessionFieldHasACopyoverRule(t *testing.T) {
	rules := map[string]string{
		"connectionId":   "restored from the saved entry",
		"isLinkDead":     "restored from the saved entry",
		"connectionTime": "reset to now",
		"lastInputRound": "reset to the current round",
		"unsentText":     "dropped; the client redraws its input line",
		"activePrompt":   "dropped; an open prompt does not survive copyover",
		"inputBlocked":   "dropped; input starts unblocked",
	}
	typ := reflect.TypeOf(Session{})
	for i := 0; i < typ.NumField(); i++ {
		if _, ok := rules[typ.Field(i).Name]; !ok {
			t.Errorf("Session.%s has no copyover rule; add one here and handle it in CopyoverRestore", typ.Field(i).Name)
		}
	}
	assert.Len(t, rules, typ.NumField(), "a rule names a field Session no longer has")
}
