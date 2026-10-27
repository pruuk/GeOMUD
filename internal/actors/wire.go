package actors

import (
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var initOnce sync.Once

// Init subscribes the registry to account and mob lifecycles. main calls it
// once at startup, before copyover restore and before any room spawns mobs.
// accounts and mobs cannot import this package (accounts imports mobs, and
// this package imports both), so they announce through hooks instead.
func Init() {
	initOnce.Do(func() {
		accounts.OnActivate.Register(func(u *accounts.Account) *accounts.Account { accountOnline(u); return u })
		accounts.OnDeactivate.Register(func(u *accounts.Account) *accounts.Account { accountOffline(u); return u })
		mobs.OnInstanceCreated.Register(func(m *mobs.Mob) *mobs.Mob { mobSpawned(m); return m })
		mobs.OnInstanceDestroyed.Register(func(id ids.ID) ids.ID { mobDestroyed(id); return id })
	})
}

func accountOnline(u *accounts.Account) {
	a, spawned, rebound := bindAccount(u)
	switch {
	case spawned:
		mudlog.Debug("actor", "event", "spawned", "kind", "player", "actorId", a.Id, "userId", u.UserId)
	case rebound:
		mudlog.Debug("actor", "event", "rebound", "kind", "player", "actorId", a.Id, "userId", u.UserId)
	}
}

func accountOffline(u *accounts.Account) {
	if a := unbindAccount(u); a != nil {
		mudlog.Debug("actor", "event", "removed", "kind", "player", "actorId", a.Id, "userId", u.UserId)
	}
}

func mobSpawned(m *mobs.Mob) {
	a := Spawn(&m.Character, NewBehaviourController(m))
	mudlog.Debug("actor", "event", "spawned", "kind", "npc", "actorId", a.Id, "mobInstanceId", m.Character.Id)
}

func mobDestroyed(instanceId ids.ID) {
	if a := unbindMob(instanceId); a != nil {
		mudlog.Debug("actor", "event", "removed", "kind", "npc", "actorId", a.Id, "mobInstanceId", instanceId)
	}
}
