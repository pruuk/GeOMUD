package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// reseedFromClock puts the shared source back on a clock seed, as production runs.
func reseedFromClock() { SeedRandForTest(time.Now().UnixNano()) }

func tenRolls() []int {
	out := make([]int, 10)
	for i := range out {
		out[i] = Rand(1000)
	}
	return out
}

func TestSeedRandForTestRepeats(t *testing.T) {
	t.Cleanup(reseedFromClock)

	SeedRandForTest(42)
	first := tenRolls()
	SeedRandForTest(42)
	assert.Equal(t, first, tenRolls(), "the same seed gives the same rolls")

	SeedRandForTest(43)
	assert.NotEqual(t, first, tenRolls(), "a different seed gives different rolls")
}

func TestSeedRandForTestCoversRollDice(t *testing.T) {
	t.Cleanup(reseedFromClock)

	SeedRandForTest(7)
	a := []int{RollDice(3, 6), RollDice(2, 20), RollDice(1, 100)}
	SeedRandForTest(7)
	b := []int{RollDice(3, 6), RollDice(2, 20), RollDice(1, 100)}
	assert.Equal(t, a, b)
}
