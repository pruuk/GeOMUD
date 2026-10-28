package behaviour

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestTheRatIdlesAndWandersWithinMaxWander is slice 6's done test: the rat,
// with its own data, idles at its activitylevel odds and never makes more
// than maxwander wander moves before it walks home.
//
// The rooms are a ring with one exit each, so the only random choices are the
// rat's own (a room with two or more exits picks in map order, which a seed
// does not fix). The ring is longer than maxwander, so the rat never wanders
// home by chance: every trip uses up its moves and paths home.
func TestTheRatIdlesAndWandersWithinMaxWander(t *testing.T) {
	b, err := os.ReadFile("../../_datafiles/world/empty/mobs/startland/1-rat.yaml")
	require.NoError(t, err)
	var rat mobs.Mob
	require.NoError(t, yaml.Unmarshal(b, &rat))
	require.Equal(t, 10, rat.ActivityLevel)
	require.Equal(t, 8, rat.MaxWander)
	require.Equal(t, []string{"emote wiggles its nose", "wander", ""}, rat.IdleCommands)

	const home, ringSize = 9301, 12
	next := func(id int) int { return home + (id-home+1)%ringSize }
	for id := home; id < home+ringSize; id++ {
		testRoom(t, id, testZone, map[string]int{"east": next(id)})
	}
	mob := testMob(t, home, home, rat.MaxWander)
	mob.ActivityLevel = rat.ActivityLevel
	mob.IdleCommands = rat.IdleCommands

	// Rounds in which the rat rolls its activity odds.
	rolls := 0
	OnIdle.Register(func(c IdleCheck) IdleCheck {
		m := c.Mob
		if m == mob && !(m.MaxWander > 0 && m.WanderCount >= m.MaxWander && m.Character.RoomId != m.HomeRoomId) {
			rolls++
		}
		return c
	})

	var inputs []string
	lid := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		if in := e.(events.Input); in.ActorId == mob.Character.Id {
			inputs = append(inputs, in.InputText)
		}
		return events.Continue
	})
	t.Cleanup(func() { events.UnregisterListener(events.Input{}, lid) })
	events.ProcessEvents()
	inputs = nil

	util.SeedRandForTest(6)
	emotes, wanders, trips, wandersSinceHome, mostWanders := 0, 0, 0, 0, 0
	for round := 0; round < 20000; round++ {
		Tick(mob)
		events.ProcessEvents()
		for _, in := range inputs {
			switch in {
			case "go east": // a wander move
				wanders++
				wandersSinceHome++
				mob.Character.RoomId = next(mob.Character.RoomId)
			case "east": // a step on the path home
				mob.Character.RoomId = next(mob.Character.RoomId)
				if mob.Character.RoomId == home {
					trips++
				}
			case "emote wiggles its nose":
				emotes++
			case "noop":
			default:
				t.Fatalf("round %d: unexpected input %q", round, in)
			}
		}
		inputs = inputs[:0]
		if mob.Character.RoomId == home {
			wandersSinceHome = 0
		}
		if wandersSinceHome > mostWanders {
			mostWanders = wandersSinceHome
		}
		require.LessOrEqual(t, wandersSinceHome, rat.MaxWander, "round %d: wandered past maxwander", round)
	}

	assert.Equal(t, rat.MaxWander, mostWanders, "the rat uses all of its maxwander moves")
	assert.Greater(t, trips, 10, "the rat walks home after its moves")

	// activitylevel 10, then 99 in 100 an idle command, one in three of each:
	// 3.3% of rolls emote and 3.3% wander.
	require.Greater(t, rolls, 15000)
	emoteRate := float64(emotes) / float64(rolls)
	wanderRate := float64(wanders) / float64(rolls)
	assert.InDelta(t, 0.033, emoteRate, 0.008, "emotes %d of %d rolls", emotes, rolls)
	assert.InDelta(t, 0.033, wanderRate, 0.008, "wanders %d of %d rolls", wanders, rolls)
}
