// Package perception is the vocabulary of Phase 4b messaging (spec section
// 2): what a message is, and what reaches a receiver. It imports nothing of
// the world, so actors.Controller can name a Perception while
// internal/messaging, which delivers them, imports actors.
package perception

import "github.com/GoMudEngine/GoMud/internal/ids"

// Kind says whether a message is something happening in the world or text
// for the actor alone (help, score, prompts, errors).
type Kind uint8

const (
	WorldEvent Kind = iota // perceived by whoever can perceive it
	Interface              // for the actor only; no sense check; behaviours drop it
)

// Sense is a set of senses that can carry a message.
type Sense uint8

const (
	Sight Sense = 1 << iota
	Sound
	Touch
	Smell
)

// senseOrder is the order senses are tried in, and SenseUsed picks the first.
var senseOrder = []Sense{Sight, Sound, Touch, Smell}

// Has reports whether s includes every sense in o.
func (s Sense) Has(o Sense) bool { return s&o == o }

// Each lists the senses in s in sense order.
func (s Sense) Each() []Sense {
	out := []Sense{}
	for _, one := range senseOrder {
		if s.Has(one) {
			out = append(out, one)
		}
	}
	return out
}

// Range is how far a message carries.
type Range uint8

const (
	Room   Range = iota // the source rooms
	Nearby              // the source rooms and the rooms next to them
	Zone                // every room in the source rooms' zones
	World               // everywhere
)

// Audience says who in range receives a message.
type Audience uint8

const (
	InRange   Audience = iota // everyone in range
	ActorOnly                 // the actor alone
	Named                     // the actor and the bodies in Message.Named
)

// Message is one thing to be perceived. Text fields may hold the tokens
// {actor}, {target} and {direction}, resolved per receiver by messaging.
type Message struct {
	Kind     Kind
	Actor    ids.ID // the body doing it; zero for a world event (weather, a door creaking)
	Target   ids.ID // optional
	Room     int    // the source room of a world event with no actor
	Senses   Sense  // which senses can carry it
	Faint    bool   // needs keen senses (today's IsQuiet)
	Range    Range
	Audience Audience
	Named    []ids.ID // for Audience Named

	ActorText    string // what the actor reads
	TargetText   string // what the target reads; ObserverText when empty
	ObserverText string // what everyone else in the source rooms reads
	// FarText is what receivers outside the source rooms read; ObserverText
	// when empty. Only a Nearby far receiver has a {direction}: Zone and
	// World far receivers have none, so FarText for those ranges must not use it.
	FarText string
	// TargetRoomText is what observers in the target's room read when the
	// target is in another room than the actor; ObserverText when empty.
	TargetRoomText string
}

// Role is how a receiver stands to a message.
type Role uint8

const (
	Actor Role = iota
	Target
	Observer
)

// Perception is what reaches one receiver: the message, how it reached
// them, and the text for them with tokens resolved. A connection controller
// writes Text; a behaviour reads the rest.
type Perception struct {
	Role       Role
	Message    Message
	SenseUsed  Sense  // zero for the actor's own text and interface text
	ActorSeen  bool   // the receiver could see the actor (true when there is none)
	TargetSeen bool   // the receiver could see the target (true when there is none)
	Far        bool   // the receiver is outside the source rooms
	Direction  string // for a far receiver next to a source room: the exit back toward it
	Text       string // the receiver's text; empty when there is none for its role
}
