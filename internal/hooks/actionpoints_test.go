package hooks

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init()
	os.Exit(m.Run())
}

func TestActionPointsRegenerateForNPCs(t *testing.T) {
	actors.Reset()
	m := mobs.NewTestInstance(mobs.Mob{})
	a := actors.Get(m.Character.Id)
	require.NotNil(t, a)
	require.Nil(t, a.Controller().Account())

	a.Character.ActionPoints = 0
	a.Character.ActionPointsMax.Value = 100

	ActionPoints(events.NewTurn{})
	assert.Equal(t, 1, a.Character.ActionPoints)

	for i := 0; i < 200; i++ {
		ActionPoints(events.NewTurn{})
	}
	assert.Equal(t, 100, a.Character.ActionPoints)
}

func TestActionPointsRegenerateForPlayers(t *testing.T) {
	actors.Reset()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.ActionPoints = 0
	c.ActionPointsMax.Value = 100
	u := &accounts.Account{UserId: 5, Character: c}
	actors.Spawn(c, actors.NewConnectionController(u))

	ActionPoints(events.NewTurn{})
	assert.Equal(t, 1, c.ActionPoints)
}
