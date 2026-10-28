package rooms

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestTemplateDataHasNoItemIds(t *testing.T) {
	r := NewEmptyRoom()
	r.Items = []items.Item{{ItemId: 1, Id: ids.New(ids.NoCell)}}
	r.Stash = []items.Item{{ItemId: 2, Id: ids.New(ids.NoCell)}}
	r.Containers = map[string]Container{"chest": {Items: []items.Item{{ItemId: 3, Id: ids.New(ids.NoCell)}}}}

	tpl := r.templateData()
	assert.True(t, tpl.Items[0].Id.IsZero())
	assert.True(t, tpl.Stash[0].Id.IsZero())
	assert.True(t, tpl.Containers["chest"].Items[0].Id.IsZero())

	assert.False(t, r.Items[0].Id.IsZero(), "the live room keeps its item IDs")
	assert.False(t, r.Stash[0].Id.IsZero())
	assert.False(t, r.Containers["chest"].Items[0].Id.IsZero())

	b, err := yaml.Marshal(&tpl)
	require.NoError(t, err)
	assert.NotContains(t, string(b), "id: 0x", "no item ID reaches the template file")
}

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init()
	// DataFiles is an empty temp dir, so a test that saves a room template
	// never writes into the repo (the default is relative to the package).
	dataDir, err := os.MkdirTemp("", "rooms-datafiles-")
	if err != nil {
		panic(err)
	}
	if err := configs.AddOverlayOverrides(map[string]any{"FilePaths.DataFiles": dataDir}); err != nil {
		panic(err)
	}
	code := m.Run()
	os.RemoveAll(dataDir)
	os.Exit(code)
}

// TestRoomHoldsOneBodyList checks that a room keeps players and mobs in one
// body list and tells them apart by whether an account drives the body.
func TestRoomHoldsOneBodyList(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)

	playerId := ids.New(ids.NoCell)
	acc := &accounts.Account{UserId: 4242, Character: &characters.Character{Id: playerId, Name: "Tester"}}
	require.NotNil(t, actors.Spawn(acc.Character, actors.NewConnectionController(acc)))

	mobId := ids.New(ids.NoCell)
	mob := &mobs.Mob{Character: characters.Character{Id: mobId, Name: "rat"}}
	require.NotNil(t, actors.Spawn(&mob.Character, actors.NewBehaviourController(mob)))

	r := NewEmptyRoom()
	assert.Equal(t, 1, r.AddPlayer(acc.UserId))
	r.bodies = append(r.bodies, mobId) // as Prepare spawns a mob: AddMob needs a live mob instance

	assert.Equal(t, []ids.ID{playerId, mobId}, r.GetBodies())
	assert.Equal(t, []int{acc.UserId}, r.GetPlayers())
	assert.Equal(t, []ids.ID{mobId}, r.GetMobs())
	assert.Equal(t, 1, r.PlayerCt())
	assert.Equal(t, 1, r.MobCt())

	uId, mId := r.FindByName(playerId.String())
	assert.Equal(t, acc.UserId, uId)
	assert.True(t, mId.IsZero())

	uId, mId = r.FindByName(mobId.String())
	assert.Equal(t, 0, uId)
	assert.Equal(t, mobId, mId)

	for _, old := range []string{"#1", "@1", "@4242"} {
		uId, mId = r.FindByName(old)
		assert.Equal(t, 0, uId, old)
		assert.True(t, mId.IsZero(), old)
	}

	n, ok := r.RemovePlayer(acc.UserId)
	assert.True(t, ok)
	assert.Equal(t, 0, n)
	assert.Equal(t, []ids.ID{mobId}, r.GetBodies())
	assert.Empty(t, r.GetPlayers())
}

// TestFindByNameSkipsBodyWithNoActor covers a disconnect race: a body ID stays
// in the room list with no actor and no mob instance. Name search must skip it.
func TestFindByNameSkipsBodyWithNoActor(t *testing.T) {
	actors.Reset()
	t.Cleanup(actors.Reset)

	r := NewEmptyRoom()
	r.bodies = append(r.bodies, ids.New(ids.NoCell))

	assert.NotPanics(t, func() {
		uId, mId := r.FindByName("rat")
		assert.Equal(t, 0, uId)
		assert.True(t, mId.IsZero())
	})
}

func TestFloorHasItemId(t *testing.T) {
	r := NewEmptyRoom()
	assert.False(t, r.floorHasItemId(10))
	r.Items = append(r.Items, items.Item{ItemId: 10})
	assert.True(t, r.floorHasItemId(10))
	assert.False(t, r.floorHasItemId(11))
	r.Stash = append(r.Stash, items.Item{ItemId: 11})
	assert.False(t, r.floorHasItemId(11), "the stash is not the floor (as FindOnFloor(_, false))")
}

func TestRoom_AddCorpse(t *testing.T) {
	r := &Room{}
	assert.Empty(t, r.Corpses, "Expected no corpses initially")

	corpse := Corpse{
		MobId:        0,
		UserId:       123,
		Character:    characters.Character{Name: "TestPlayer"},
		RoundCreated: 10,
	}
	r.AddCorpse(corpse)
	assert.Len(t, r.Corpses, 1, "Expected exactly one corpse after adding")
	assert.Equal(t, corpse, r.Corpses[0], "Expected the added corpse to match")
}

func TestRoom_RemoveCorpse(t *testing.T) {
	r := &Room{}
	corpse1 := Corpse{
		MobId:        0,
		UserId:       123,
		Character:    characters.Character{Name: "PlayerOne"},
		RoundCreated: 10,
	}
	corpse2 := Corpse{
		MobId:        456,
		UserId:       0,
		Character:    characters.Character{Name: "MobOne"},
		RoundCreated: 11,
	}

	r.AddCorpse(corpse1)
	r.AddCorpse(corpse2)

	// Removing existing corpse
	removed := r.RemoveCorpse(corpse1)
	assert.True(t, removed, "Expected to remove an existing corpse successfully")
	assert.Len(t, r.Corpses, 1, "Expected exactly one corpse remaining")

	// Try removing a corpse that does not exist
	nonExistent := Corpse{
		MobId:        999,
		UserId:       999,
		Character:    characters.Character{Name: "Ghost"},
		RoundCreated: 99,
	}
	removed = r.RemoveCorpse(nonExistent)
	assert.False(t, removed, "Expected removal to fail for non-existent corpse")
	assert.Len(t, r.Corpses, 1, "Expected no change in corpses")
}

func TestRoom_FindCorpse(t *testing.T) {
	r := &Room{}

	playerCorpse := Corpse{
		UserId:       123,
		Character:    characters.Character{Name: "PlayerOne"},
		RoundCreated: 5,
		Prunable:     false,
	}
	mobCorpse := Corpse{
		MobId:        456,
		Character:    characters.Character{Name: "MobOne"},
		RoundCreated: 6,
		Prunable:     false,
	}
	r.AddCorpse(playerCorpse)
	r.AddCorpse(mobCorpse)

	// Exact search
	found, ok := r.FindCorpse("PlayerOne corpse")
	assert.True(t, ok, "Expected to find player corpse by exact name")
	assert.Equal(t, "PlayerOne", found.Character.Name, "Expected found corpse to match the correct character")

	// Searching for mob
	found, ok = r.FindCorpse("MobOne corpse")
	assert.True(t, ok, "Expected to find mob corpse by exact name")
	assert.Equal(t, "MobOne", found.Character.Name, "Expected found corpse to match the correct character")

	// Searching partial name (depends on your util.FindMatchIn logic)
	found, ok = r.FindCorpse("player")
	assert.True(t, ok, "Expected to find a close match for player corpse")
	assert.Equal(t, "PlayerOne", found.Character.Name, "Expected found corpse to be the player's")

	// Non-existent
	found, ok = r.FindCorpse("NonExistent")
	assert.False(t, ok, "Expected not to find a missing corpse")
}

func TestFindNoun(t *testing.T) {
	// Create a room with various noun mappings (including aliases).
	r := &Room{
		Nouns: map[string]string{
			"torch":            "a fiery torch",
			"torchAlias":       ":torch", // alias -> :torch means "torch"
			"lamp":             "an illuminating lamp",
			"lampAlias":        ":lamp", // alias -> :lamp means "lamp"
			"candles":          "some wax candles",
			"pony":             "a small horse",
			"mystery":          "just a riddle",
			"secret":           ":mystery",                   // chain alias -> :mystery
			"projector screen": "a wide, matte-white screen", // multi-word matching
			"island":           "a tropical paradise",
			"rocks":            ":island",
			"rocky island":     ":island",
			"feet":             ":hands",
			"hands":            ":feet",
		},
	}

	// Table-driven tests.
	tests := []struct {
		name          string
		inputNoun     string
		wantFoundNoun string
		wantDesc      string
	}{
		{
			name:          "Direct match (torch)",
			inputNoun:     "torch",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "Direct match (candles)",
			inputNoun:     "candles",
			wantFoundNoun: "candles",
			wantDesc:      "some wax candles",
		},
		{
			name:          "Direct match (pony)",
			inputNoun:     "pony",
			wantFoundNoun: "pony",
			wantDesc:      "a small horse",
		},
		{
			name:          "Alias match (torchAlias -> :torch)",
			inputNoun:     "torchAlias",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "Alias match (lampAlias -> :lamp)",
			inputNoun:     "lampAlias",
			wantFoundNoun: "lamp",
			wantDesc:      "an illuminating lamp",
		},
		{
			name:          "Chain alias (secret -> :mystery)",
			inputNoun:     "secret",
			wantFoundNoun: "mystery",
			wantDesc:      "just a riddle",
		},
		{
			name:          "Plural form by adding s (pony -> ponies)",
			inputNoun:     "ponies",
			wantFoundNoun: "pony",
			wantDesc:      "a small horse",
		},
		{
			name:          "Plural form by adding es (torches)",
			inputNoun:     "torches",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},

		{
			name:          "Multi-word input, second word valid (red torches)",
			inputNoun:     "red torches",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "Multi-word input, multi-word match",
			inputNoun:     "projector screen",
			wantFoundNoun: "projector screen",
			wantDesc:      "a wide, matte-white screen",
		},
		{
			name:          "Single-word input, multi-word match 1",
			inputNoun:     "projector",
			wantFoundNoun: "projector screen",
			wantDesc:      "a wide, matte-white screen",
		},
		{
			name:          "Single-word input, multi-word match 2",
			inputNoun:     "screen",
			wantFoundNoun: "projector screen",
			wantDesc:      "a wide, matte-white screen",
		},
		{
			name:          "Multi-word input, first word valid (torch something)",
			inputNoun:     "torch something",
			wantFoundNoun: "torch",
			wantDesc:      "a fiery torch",
		},
		{
			name:          "No match",
			inputNoun:     "gibberish",
			wantFoundNoun: "",
			wantDesc:      "",
		},
		{
			name:          "Multi-word no match (foo bar)",
			inputNoun:     "foo bar",
			wantFoundNoun: "",
			wantDesc:      "",
		},
		{
			name:          "Multi-word input, first word valid (island)",
			inputNoun:     "island",
			wantFoundNoun: "island",
			wantDesc:      "a tropical paradise",
		},
		// Issue #356: Fix Noun aliases
		{
			name:          "Multi-word input, first word valid (rocks)",
			inputNoun:     "rocks",
			wantFoundNoun: "island",
			wantDesc:      "a tropical paradise",
		},
		{
			name:          "Multi-word input, first word valid (rocky island)",
			inputNoun:     "rocky island",
			wantFoundNoun: "island",
			wantDesc:      "a tropical paradise",
		},
		{
			name:          "Multi-word input, first word valid (rocky)",
			inputNoun:     "rocky",
			wantFoundNoun: "island",
			wantDesc:      "a tropical paradise",
		},
		{
			name:          "circular references (hands)",
			inputNoun:     "hands",
			wantFoundNoun: "",
			wantDesc:      "",
		},
		{
			name:          "circular references (feet)",
			inputNoun:     "feet",
			wantFoundNoun: "",
			wantDesc:      "",
		},
	}

	// Run each sub-test.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFound, gotDesc := r.FindNoun(tt.inputNoun)
			if gotFound != tt.wantFoundNoun || gotDesc != tt.wantDesc {
				t.Errorf("FindNoun(%q) = (%q, %q), want (%q, %q)",
					tt.inputNoun, gotFound, gotDesc, tt.wantFoundNoun, tt.wantDesc)
			}
		})
	}
}

func TestInstanceItemIdsReplaceTheTemplates(t *testing.T) {
	tpl := NewEmptyRoom()
	tpl.Items = []items.Item{{ItemId: 1}} // a template item, no ID
	live := NewEmptyRoom()
	live.Items = []items.Item{{ItemId: 1, Id: ids.New(ids.NoCell)}, {ItemId: 2, Id: ids.New(ids.NoCell)}}

	b, err := yaml.Marshal(map[string]any{"items": live.Items}) // as saveRoomInstanceData writes a changed field
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(b, tpl)) // as LoadRoomInstance applies it

	require.Len(t, tpl.Items, 2)
	assert.Equal(t, live.Items[0].Id, tpl.Items[0].Id)
	assert.Equal(t, live.Items[1].Id, tpl.Items[1].Id)
}
