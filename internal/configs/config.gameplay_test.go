package configs

import "testing"

// Phase 5 slice 4.1 (ruling 12): the out-of-reach wait limit defaults to 5
// rounds when it is 0 or less.
func TestOutOfReachRoundsDefaultsToFive(t *testing.T) {
	for set, want := range map[ConfigInt]ConfigInt{0: 5, -2: 5, 1: 1, 3: 3} {
		g := GamePlay{OutOfReachRounds: set}
		g.Validate()
		if g.OutOfReachRounds != want {
			t.Errorf("OutOfReachRounds %d validates to %d, want %d", set, g.OutOfReachRounds, want)
		}
	}
}
