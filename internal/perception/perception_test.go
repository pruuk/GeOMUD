package perception

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSenseSet(t *testing.T) {
	s := Sight | Sound
	assert.True(t, s.Has(Sight))
	assert.True(t, s.Has(Sound))
	assert.False(t, s.Has(Touch))
	assert.Equal(t, []Sense{Sight, Sound}, s.Each())
	assert.Equal(t, []Sense{Sight, Sound, Touch, Smell}, (Smell | Touch | Sound | Sight).Each(), "always in sense order")
	assert.Empty(t, Sense(0).Each())
}
