// Package actors is the one actor model (Phase 4b, decision 4): every live
// body in the world, player or NPC, is an Actor driven by a Controller.
// Spec: 2026-10-27-phase-4b-npcs-are-players-design.md.
package actors

import "github.com/GoMudEngine/GoMud/internal/characters"

// ActorId identifies a live actor. It is runtime only and never saved: a
// player's ActorId differs between logins. Zero means no actor.
type ActorId int

// Actor is one live body in the world.
type Actor struct {
	Id        ActorId
	Character *characters.Character

	controller Controller
}

// Controller returns what drives the actor. Swap it with SetController.
func (a *Actor) Controller() Controller {
	mu.RLock()
	defer mu.RUnlock()
	return a.controller
}
