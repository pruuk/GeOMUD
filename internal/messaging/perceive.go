package messaging

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

// CanPerceive reports whether receiver, in room, takes in m through sense.
// It is the hook Phase 5's ruleset replaces (spec section 2). In 4b: sight
// needs light or nightvision (look's rule); a faint sound needs
// superhearing (today's quiet messages); touch and smell pass.
var CanPerceive = func(receiver *actors.Actor, room *rooms.Room, m perception.Message, sense perception.Sense) bool {
	switch sense {
	case perception.Sight:
		return canSeeIn(receiver, room)
	case perception.Sound:
		return !m.Faint || hasFlag(receiver.Character, "superhearing")
	default:
		return true
	}
}

// hasFlag and roomVisibility are seams for tests: buff flags need loaded buff
// specs, and visibility reads the game clock and biome.
var (
	hasFlag        = func(c *characters.Character, flag string) bool { return c.HasBuffFlag(flag) }
	roomVisibility = func(r *rooms.Room) int { return r.GetVisibility() }
)

// canSeeIn reports whether receiver can see in room.
func canSeeIn(receiver *actors.Actor, room *rooms.Room) bool {
	if room == nil {
		return false
	}
	return roomVisibility(room) >= 1 || hasFlag(receiver.Character, "nightvision")
}

// canSee reports whether receiver sees the body subject: a body always sees
// itself; a hidden body is unseen; a far receiver, or one whose sight is
// blocked in its own room, sees no one.
func canSee(receiver *actors.Actor, room *rooms.Room, subject *actors.Actor, far bool) bool {
	if subject == nil {
		return true
	}
	if subject.Id == receiver.Id {
		return true
	}
	if far || hasFlag(subject.Character, "hidden") {
		return false
	}
	return canSeeIn(receiver, room)
}

// perceivedBy returns the first sense in m.Senses that carries m to
// receiver, or zero when none does.
func perceivedBy(receiver *actors.Actor, room *rooms.Room, m perception.Message) perception.Sense {
	for _, sense := range m.Senses.Each() {
		if CanPerceive(receiver, room, m, sense) {
			return sense
		}
	}
	return 0
}
