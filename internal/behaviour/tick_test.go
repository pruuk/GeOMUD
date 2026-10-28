package behaviour

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// busyMob is an NPC that idles every round it may: activity 100, one emote.
func busyMob(t *testing.T, roomId, homeId, maxWander int) *mobs.Mob {
	t.Helper()
	m := testMob(t, roomId, homeId, maxWander)
	m.ActivityLevel = 100
	m.IdleCommands = []string{"emote hums."}
	return m
}

// tickRounds runs n rounds of Tick and returns the input queued.
func tickRounds(t *testing.T, mob *mobs.Mob, n int) []string {
	t.Helper()
	return queuedInput(t, mob, func() {
		for i := 0; i < n; i++ {
			Tick(mob)
		}
	})
}

func TestTickIdlesAtItsActivityLevel(t *testing.T) {
	testRoom(t, 9201, testZone, nil)
	mob := busyMob(t, 9201, 9201, 3)
	util.SeedRandForTest(1)

	assert.Contains(t, tickRounds(t, mob, 3), "emote hums.")
}

func TestTickWithActivityZeroNeverIdles(t *testing.T) {
	testRoom(t, 9202, testZone, nil)
	mob := busyMob(t, 9202, 9202, 3)
	mob.ActivityLevel = 0

	assert.Empty(t, tickRounds(t, mob, 50))
}

func TestTickPreventIdleSkipsOneRound(t *testing.T) {
	testRoom(t, 9203, testZone, nil)
	mob := busyMob(t, 9203, 9203, 3)
	mob.PreventIdle = true

	assert.Empty(t, tickRounds(t, mob, 1))
	assert.False(t, mob.PreventIdle, "the hold lasts one round")
}

func TestTickWhileFightingDoesNotIdle(t *testing.T) {
	testRoom(t, 9204, testZone, nil)
	mob := busyMob(t, 9204, 9204, 3)
	mob.Character.Aggro = &characters.Aggro{}

	assert.Empty(t, tickRounds(t, mob, 5))
}

func TestTickOnIdleSkipStopsTheIdle(t *testing.T) {
	testRoom(t, 9205, testZone, nil)
	mob := busyMob(t, 9205, 9205, 3)
	OnIdle.Register(func(c IdleCheck) IdleCheck {
		if c.Mob == mob {
			c.Skip = true
		}
		return c
	})

	assert.Empty(t, tickRounds(t, mob, 5))
}

func TestTickAtHomeOverTheWanderLimitStartsAgain(t *testing.T) {
	testRoom(t, 9206, testZone, nil)
	mob := busyMob(t, 9206, 9206, 2)
	mob.ActivityLevel = 0
	mob.WanderCount = 5

	tickRounds(t, mob, 1)

	assert.Equal(t, 0, mob.WanderCount, "an NPC home again may wander again")
}

func TestTickAwayAtTheWanderLimitPlansTheWayHome(t *testing.T) {
	testRoom(t, 9207, testZone, map[string]int{"north": 9208})
	testRoom(t, 9208, testZone, map[string]int{"south": 9207})
	mob := busyMob(t, 9208, 9207, 2)
	mob.WanderCount = 2

	got := tickRounds(t, mob, 1)

	assert.Empty(t, got, "it plans; it does not idle")
	assert.Equal(t, 1, mob.Path.Len())
}

func TestTickWalksItsPath(t *testing.T) {
	testRoom(t, 9209, testZone, map[string]int{"east": 9210})
	testRoom(t, 9210, testZone, nil)
	mob := busyMob(t, 9209, 9209, 3)
	PathTo(mob, "9210")

	got := tickRounds(t, mob, 1)

	require.NotEmpty(t, got)
	assert.Equal(t, "east", got[0])
}
