package hooks

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActorUpkeepTicksEveryActorsCooldowns(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Cooldowns = characters.Cooldowns{"kick": 3}
	actors.Spawn(c, actors.NewConnectionController(&accounts.Account{UserId: 61, Character: c}))
	m := mobs.NewTestInstance(mobs.Mob{})
	t.Cleanup(func() { mobs.Rest(m.Character.Id) })
	m.Character.Cooldowns = characters.Cooldowns{"bite": 3}

	ActorUpkeep(events.NewRound{RoundNumber: 1})

	assert.Equal(t, 2, c.Cooldowns["kick"], "a player's cooldowns tick")
	assert.Equal(t, 2, m.Character.Cooldowns["bite"], "an NPC's cooldowns tick")
}

func TestActorTickRunsEachNPCsBehaviourOnce(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	actors.Spawn(c, actors.NewConnectionController(&accounts.Account{UserId: 62, Character: c}))
	m := mobs.NewTestInstance(mobs.Mob{})
	t.Cleanup(func() { mobs.Rest(m.Character.Id) })
	require.NotNil(t, actors.Get(m.Character.Id))

	ticks := 0
	actors.OnBehaviourTick.Register(func(x *mobs.Mob) *mobs.Mob {
		if x == m {
			ticks++
		}
		return x
	})

	ActorTick(events.NewRound{RoundNumber: 1})

	assert.Equal(t, 1, ticks)
}
