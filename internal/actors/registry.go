package actors

import (
	"fmt"
	"sort"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var (
	mu       sync.RWMutex
	byId     = map[ActorId]*Actor{}
	byUserId = map[int]ActorId{}
)

// Spawn makes a live actor for char, driven by ctrl. The actor's ID is the
// body's ID. A body without an ID, or a second body claiming an ID already
// registered, is a bug: Spawn logs it and returns nil.
func Spawn(char *characters.Character, ctrl Controller) *Actor {
	mu.Lock()
	defer mu.Unlock()
	return spawnLocked(char, ctrl)
}

// kindOf names what a controller drives, for logs: a player or an NPC.
func kindOf(ctrl Controller) string {
	if ctrl != nil && ctrl.Account() != nil {
		return "player"
	}
	return "npc"
}

// spawnLocked is Spawn with mu held. An actor already under the same ID is
// replaced only when it holds the same body; a different body with that ID
// is refused.
func spawnLocked(char *characters.Character, ctrl Controller) *Actor {
	if char == nil || char.Id.IsZero() {
		mudlog.Error("actor", "event", "no body id", "reason", "refusing to spawn a body without an ID")
		return nil
	}
	if prev, ok := byId[char.Id]; ok {
		if prev.Character != char {
			// A different body claims an ID already registered: refuse it and
			// leave the registered actor untouched.
			mudlog.Error("actor", "event", "body id collision", "actorId", char.Id,
				"registeredKind", kindOf(prev.controller), "claimingKind", kindOf(ctrl),
				"reason", "refusing a second body with the same ID")
			return nil
		}
		unindex(prev)
	}
	a := &Actor{Id: char.Id, Character: char, controller: ctrl}
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

// AccountOf returns the account driving the body with id, or nil when there
// is no such actor or no account drives it (an NPC).
func AccountOf(id ActorId) *accounts.Account {
	a := Get(id)
	if a == nil {
		return nil
	}
	ctrl := a.Controller()
	if ctrl == nil {
		return nil
	}
	return ctrl.Account()
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

// All returns every live actor, in ID text order.
func All() []*Actor {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Actor, 0, len(byId))
	for _, a := range byId {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id.String() < out[j].Id.String() })
	return out
}

// Each calls fn for every live actor, in no particular order, under the
// registry's read lock: no copy and no sort, for per-turn loops. fn must not
// call back into the registry's write paths (Spawn, Remove, SetController,
// Reset or a bind), which would deadlock.
func Each(fn func(*Actor)) {
	mu.RLock()
	defer mu.RUnlock()
	for _, a := range byId {
		fn(a)
	}
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
		return fmt.Errorf("actor %s not found", id)
	}
	unindex(a)
	a.controller = ctrl
	index(a)
	return nil
}

// bindAccount gives u an actor, or points the account's existing actor at u
// when u is a different record for the same account. The lookup and the
// change happen under one lock. It reports what it did, for logging. It
// returns nil when u's body has no ID.
func bindAccount(u *accounts.Account) (a *Actor, spawned, rebound bool) {
	mu.Lock()
	defer mu.Unlock()
	return bindLocked(u)
}

// bindLocked is bindAccount with mu held.
func bindLocked(u *accounts.Account) (a *Actor, spawned, rebound bool) {
	if u.Character == nil || u.Character.Id.IsZero() {
		mudlog.Error("actor", "event", "no body id", "userId", u.UserId)
		return nil, false, false
	}
	if id, ok := byUserId[u.UserId]; ok {
		a = byId[id]
		if a.Id == u.Character.Id {
			if a.controller.Account() != u {
				unindex(a)
				a.Character = u.Character
				a.controller = NewConnectionController(u)
				index(a)
				rebound = true
			}
			return a, false, rebound
		}
		// The account's body has a different ID (a permadeath reroll): the
		// actor under the old ID goes and the new body gets its own.
		unindex(a)
		delete(byId, id)
		a = spawnLocked(u.Character, NewConnectionController(u))
		return a, false, a != nil
	}
	a = spawnLocked(u.Character, NewConnectionController(u))
	return a, a != nil, false
}

// BodyReplaced re-keys a player whose body was replaced (a permadeath reroll):
// under one lock it removes the actor with ID old if u drives it, then binds
// u's new body.
func BodyReplaced(u *accounts.Account, old ActorId) {
	mu.Lock()
	defer mu.Unlock()
	if a, ok := byId[old]; ok && a.controller.Account() == u {
		unindex(a)
		delete(byId, old)
	}
	if a, _, _ := bindLocked(u); a != nil {
		mudlog.Debug("actor", "event", "replaced", "kind", "player", "actorId", a.Id, "oldActorId", old, "userId", u.UserId)
	}
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
	if a == nil || a.controller.Account() != u {
		return nil
	}
	unindex(a)
	delete(byId, id)
	return a
}

// unbindMob removes the NPC actor with the mob's body ID and returns it, or
// nil. A player's actor is never removed this way.
func unbindMob(id ActorId) *Actor {
	mu.Lock()
	defer mu.Unlock()
	a := byId[id]
	if a == nil || a.controller.Account() != nil {
		return nil
	}
	unindex(a)
	delete(byId, id)
	return a
}

// index and unindex keep the account lookup current. The caller holds mu.
func index(a *Actor) {
	if acc := a.controller.Account(); acc != nil {
		byUserId[acc.UserId] = a.Id
	}
}

func unindex(a *Actor) {
	if acc := a.controller.Account(); acc != nil && byUserId[acc.UserId] == a.Id {
		delete(byUserId, acc.UserId)
	}
}

// Reset empties the registry. For tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	byId = map[ActorId]*Actor{}
	byUserId = map[int]ActorId{}
}
