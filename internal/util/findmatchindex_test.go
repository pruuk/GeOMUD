package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FindMatchIndexIn tells repeated items apart by position (two corpses of
// one name): "#N" picks the Nth, and an exact match wins over a prefix that
// comes before it.
func TestFindMatchIndexInTellsRepeatsApart(t *testing.T) {
	list := []string{"rat corpsewing corpse", "rat corpse", "rat corpse"}
	for _, tt := range []struct {
		search       string
		match, close int
	}{
		{"rat", -1, 0},
		{"rat#1", -1, 0},
		{"rat#2", -1, 1},
		{"rat#3", -1, 2},
		{"rat#4", -1, -1},
		{"rat corpse", 1, 0},
		{"rat corpse#2", 2, 1},
		{"", -1, -1},
	} {
		match, close := FindMatchIndexIn(tt.search, list...)
		assert.Equal(t, tt.match, match, "%q: match", tt.search)
		assert.Equal(t, tt.close, close, "%q: close match", tt.search)
	}
}
