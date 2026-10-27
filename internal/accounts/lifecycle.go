package accounts

import (
	"sync"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	// OnActivate fires after an account is recorded as online: login,
	// link-dead reconnect, copyover reconnect and restore, and account
	// creation. It can fire again for an account that is already online.
	OnActivate util.Hook[*Account]
	// OnDeactivate fires after an online account is logged out.
	OnDeactivate util.Hook[*Account]

	// lifecycleMu makes each online or offline change and its hook one step,
	// so OnActivate and OnDeactivate fire in the order the online map
	// changes. Lock order: lifecycleMu before userManagerMu. Hook handlers
	// must not call a function that takes lifecycleMu.
	lifecycleMu sync.Mutex
)

// activateLocked records u as online on connectionId. The caller holds
// userManagerMu and calls activated(u) after unlocking.
func activateLocked(u *Account, connectionId connections.ConnectionId) {
	u.sess().connectionId = connectionId
	userManager.Users[u.UserId] = u
	userManager.Usernames[u.Username] = u.UserId
	userManager.Connections[connectionId] = u.UserId
	userManager.UserConnections[u.UserId] = connectionId
}

// activated fires OnActivate. Call it without holding userManagerMu.
func activated(u *Account) { OnActivate.Fire(u) }

// deactivated fires OnDeactivate. Call it without holding userManagerMu.
func deactivated(u *Account) { OnDeactivate.Fire(u) }
