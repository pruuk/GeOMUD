package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Both ordinal forms, one parser (Calabe, 2026-11-11): "2.rat" and "rat#2"
// are the second rat. A dot counts only between digits and a name that does
// not start with a digit; anything else stays as typed.
func TestGetMatchNumberReadsBothOrdinalForms(t *testing.T) {
	for _, tt := range []struct {
		input string
		name  string
		num   int
	}{
		{"2.rat", "rat", 2},
		{"rat#2", "rat", 2},
		{"0.rat", "rat", 1}, // as rat#0: below 1 is 1
		{"rat", "rat", 1},
		{"1.5", "1.5", 1},   // a number, not the fifth "5"
		{".rat", ".rat", 1}, // no digits before the dot
		{"2.", "2.", 1},     // no name after the dot
		{"12.Rat Corpse", "rat corpse", 12},
		{"2.rat#3", "2.rat", 3}, // "#N" is read first
		{"a2.rat", "a2.rat", 1}, // not all digits before the dot
	} {
		name, num := GetMatchNumber(tt.input)
		assert.Equal(t, tt.name, name, "%q: name", tt.input)
		assert.Equal(t, tt.num, num, "%q: number", tt.input)
	}
}
