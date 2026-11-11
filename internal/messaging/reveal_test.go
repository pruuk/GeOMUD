package messaging

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/stretchr/testify/assert"
)

// One visibility rule (Calabe, 2026-11-11): a hidden speaker reads by name
// to an admin and to a receiver with see-hidden, and as "Someone" to anyone
// else. Darkness still hides it from them all: the reveal is of hidden, not
// of light (canSee keeps its light and distance rules).
func TestHiddenIsRevealedToAdminsAndSeeHidden(t *testing.T) {
	w := newWorld(t)
	square := w.room(1, "town", nil)
	sneak, _ := w.body(square, "Sneak", true)
	w.flags[sneak]["hidden"] = true
	_, bobRec := w.body(square, "Bob", true)
	_, adminRec := w.body(square, "Ada", true)
	adminRec.acc.Role = accounts.RoleAdmin
	seer, seerRec := w.body(square, "Seer", true)
	w.flags[seer]["see-hidden"] = true

	Send(say(sneak))
	assert.Equal(t, []string{`Someone says, "hi"`}, texts(bobRec))
	assert.Equal(t, []string{`<ansi fg="username">Sneak</ansi> says, "hi"`}, texts(adminRec), "an admin perceives the hidden")
	assert.Equal(t, []string{`<ansi fg="username">Sneak</ansi> says, "hi"`}, texts(seerRec), "see-hidden perceives the hidden")

	cave := w.room(2, "caves", nil)
	w.light[2] = 0
	dark, _ := w.body(cave, "Dark", true)
	w.flags[dark]["hidden"] = true
	_, caveAdminRec := w.body(cave, "Cad", true)
	caveAdminRec.acc.Role = accounts.RoleAdmin
	Send(say(dark))
	assert.Equal(t, []string{`Someone says, "hi"`}, texts(caveAdminRec), "darkness still hides it from an admin")
}
