package actors

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/mobs"
)

// Controller is what drives an actor: a connection with its account, or an
// NPC behaviour. Code outside this package asks the controller questions and
// never asserts its concrete type (internal/guards enforces this). Later 4b
// slices add Perceive (slice 3) and Tick (slice 6).
type Controller interface {
	// Account is the account driving the actor, or nil when none does.
	Account() *accounts.Account
}

// ConnectionController drives an actor from an account's connection. The
// connection state is the account's Session, so moving this controller to
// another actor moves the account and its connection together.
type ConnectionController struct {
	account *accounts.Account
}

func NewConnectionController(a *accounts.Account) *ConnectionController {
	return &ConnectionController{account: a}
}

func (c *ConnectionController) Account() *accounts.Account { return c.account }

// BehaviourController drives an NPC. In slice 1 it only holds the mob; slice
// 6 moves the mob's runtime behaviour state here and runs it from Tick.
type BehaviourController struct {
	mob *mobs.Mob
}

func NewBehaviourController(m *mobs.Mob) *BehaviourController {
	return &BehaviourController{mob: m}
}

func (b *BehaviourController) Account() *accounts.Account { return nil }
