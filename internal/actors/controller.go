package actors

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// Controller is what drives an actor: a connection with its account, or an
// NPC behaviour. Code outside this package asks the controller questions and
// never asserts its concrete type (internal/guards enforces this). Slice 6
// adds Tick.
type Controller interface {
	// Account is the account driving the actor, or nil when none does.
	Account() *accounts.Account
	// Perceive is called by internal/messaging for each message that
	// reaches the actor.
	Perceive(p perception.Perception)
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

// Perceive writes the receiver's text to the account's output, in order
// with everything else it is sent. A perception with no text for this
// receiver writes nothing.
func (c *ConnectionController) Perceive(p perception.Perception) {
	if p.Text == "" {
		return
	}
	c.account.SendText(p.Text)
}

// BehaviourController drives an NPC. In slice 1 it only holds the mob; slice
// 6 moves the mob's runtime behaviour state here and runs it from Tick.
type BehaviourController struct {
	mob *mobs.Mob
}

func NewBehaviourController(m *mobs.Mob) *BehaviourController {
	return &BehaviourController{mob: m}
}

func (b *BehaviourController) Account() *accounts.Account { return nil }

// BehaviourPerception is one perception reaching an NPC.
type BehaviourPerception struct {
	Mob        *mobs.Mob
	Perception perception.Perception
}

// OnBehaviourPerceive fires for each world event an NPC perceives. Slice 6
// attaches behaviour to it.
var OnBehaviourPerceive util.Hook[BehaviourPerception]

// Perceive hands a world event to the NPC's behaviour as a structured
// perception. Interface text is not for behaviours and is dropped.
func (b *BehaviourController) Perceive(p perception.Perception) {
	if p.Message.Kind == perception.Interface {
		return
	}
	OnBehaviourPerceive.Fire(BehaviourPerception{Mob: b.mob, Perception: p})
}
