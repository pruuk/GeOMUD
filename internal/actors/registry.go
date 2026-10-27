package actors

import (
	"fmt"
	"sort"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
)

var (
	mu              sync.RWMutex
	lastId          ActorId
	byId            = map[ActorId]*Actor{}
	byUserId        = map[int]ActorId{}
	byMobInstanceId = map[int]ActorId{}
)

// Spawn makes a live actor for char, driven by ctrl.
func Spawn(char *characters.Character, ctrl Controller) *Actor {
	mu.Lock()
	defer mu.Unlock()
	lastId++
	a := &Actor{Id: lastId, Character: char, controller: ctrl}
	byId[a.Id] = a
	index(a)
	return a
}

// Get returns the actor with id, or nil.
func Get(id ActorId) *Actor {
	mu.RLock()
	defer mu.RUnlock()
	return byId[id]
}

// ByUserId returns the actor an account drives, or nil. It routes old user
// IDs until slice 2 moves callers to ActorId.
func ByUserId(userId int) *Actor {
	mu.RLock()
	defer mu.RUnlock()
	if id, ok := byUserId[userId]; ok {
		return byId[id]
	}
	return nil
}

// ByMobInstanceId returns the actor for a mob instance, or nil. It routes old
// mob instance IDs until slice 2 removes them.
func ByMobInstanceId(instanceId int) *Actor {
	mu.RLock()
	defer mu.RUnlock()
	if id, ok := byMobInstanceId[instanceId]; ok {
		return byId[id]
	}
	return nil
}

// All returns every live actor, in ActorId order.
func All() []*Actor {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Actor, 0, len(byId))
	for _, a := range byId {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}

// Remove drops the actor. Removing an unknown id does nothing.
func Remove(id ActorId) {
	mu.Lock()
	defer mu.Unlock()
	if a, ok := byId[id]; ok {
		unindex(a)
		delete(byId, id)
	}
}

// SetController swaps what drives the actor. The ActorId does not change.
func SetController(id ActorId, ctrl Controller) error {
	mu.Lock()
	defer mu.Unlock()
	a, ok := byId[id]
	if !ok {
		return fmt.Errorf("actor %d not found", id)
	}
	unindex(a)
	a.controller = ctrl
	index(a)
	return nil
}

// bindAccount gives u an actor, or points the account's existing actor at u
// when u is a different record for the same account. The lookup and the
// change happen under one lock. It reports what it did, for logging.
func bindAccount(u *accounts.Account) (a *Actor, spawned, rebound bool) {
	mu.Lock()
	defer mu.Unlock()
	if id, ok := byUserId[u.UserId]; ok {
		a = byId[id]
		if a.controller.Account() != u {
			unindex(a)
			a.Character = u.Character
			a.controller = NewConnectionController(u)
			index(a)
			rebound = true
		}
		return a, false, rebound
	}
	lastId++
	a = &Actor{Id: lastId, Character: u.Character, controller: NewConnectionController(u)}
	byId[a.Id] = a
	index(a)
	return a, true, false
}

// unbindAccount removes the actor u drives, but only while u is still the
// record driving it: a late logout of an older record leaves a newer login's
// actor alone. It returns the removed actor, or nil.
func unbindAccount(u *accounts.Account) *Actor {
	mu.Lock()
	defer mu.Unlock()
	id, ok := byUserId[u.UserId]
	if !ok {
		return nil
	}
	a := byId[id]
	if a.controller.Account() != u {
		return nil
	}
	unindex(a)
	delete(byId, id)
	return a
}

// unbindMob removes the actor for a mob instance and returns it, or nil.
func unbindMob(instanceId int) *Actor {
	mu.Lock()
	defer mu.Unlock()
	id, ok := byMobInstanceId[instanceId]
	if !ok {
		return nil
	}
	a := byId[id]
	unindex(a)
	delete(byId, id)
	return a
}

// index and unindex keep the old-ID lookups current. The caller holds mu.
func index(a *Actor) {
	if acc := a.controller.Account(); acc != nil {
		byUserId[acc.UserId] = a.Id
	}
	if b, ok := a.controller.(*BehaviourController); ok && b.mob != nil {
		byMobInstanceId[b.mob.InstanceId] = a.Id
	}
}

func unindex(a *Actor) {
	if acc := a.controller.Account(); acc != nil && byUserId[acc.UserId] == a.Id {
		delete(byUserId, acc.UserId)
	}
	if b, ok := a.controller.(*BehaviourController); ok && b.mob != nil && byMobInstanceId[b.mob.InstanceId] == a.Id {
		delete(byMobInstanceId, b.mob.InstanceId)
	}
}

// Reset empties the registry. For tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	lastId = 0
	byId = map[ActorId]*Actor{}
	byUserId = map[int]ActorId{}
	byMobInstanceId = map[int]ActorId{}
}
