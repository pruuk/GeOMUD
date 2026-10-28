package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestPruneKeepsRestingNPCAndDropsUnknown(t *testing.T) {
	resting := ids.New(ids.NoCell)
	unknown := ids.New(ids.NoCell)

	old := NPCExists
	NPCExists = func(id ids.ID) bool { return id == resting }
	defer func() { NPCExists = old }()

	u := NewAccount(91, 0)
	u.Character.TrackCharmed(resting, true)
	u.Character.TrackCharmed(unknown, true)

	pruneCharmed(u)

	assert.Equal(t, []ids.ID{resting}, u.Character.GetCharmIds())
}

func TestCharmedMobsSurviveTheUserFile(t *testing.T) {
	u := NewAccount(92, 0)
	a, b := ids.New(ids.NoCell), ids.New(ids.NoCell)
	u.Character.TrackCharmed(a, true)
	u.Character.TrackCharmed(b, true)

	data, err := yaml.Marshal(u)
	require.NoError(t, err)
	var back Account
	require.NoError(t, yaml.Unmarshal(data, &back))

	assert.Equal(t, []ids.ID{a, b}, back.Character.GetCharmIds())
}
