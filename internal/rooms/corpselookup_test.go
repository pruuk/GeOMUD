package rooms

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// corpseOf is a corpse for the lookup tests; gold tells them apart.
func corpseOf(name string, userId, mobId int, round uint64, gold int) Corpse {
	return Corpse{UserId: userId, MobId: mobId, Character: characters.Character{Name: name}, RoundCreated: round, Gold: gold}
}

// With nobody living in the room, a name numbers its corpses newest first
// (RoundCreated; of two from one round, the later added), crumbled ones
// left out, and both ordinal forms work (Calabe, 2026-11-11). FindCorpse
// finds the same corpse.
func TestFindCorpseByRefNumbersCorpsesNewestFirst(t *testing.T) {
	r := &Room{}
	crumbled := corpseOf("rat", 0, 7, 9, 9)
	crumbled.Prunable = true
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 1))
	r.AddCorpse(corpseOf("rat", 0, 7, 8, 3)) // the newest
	r.AddCorpse(crumbled)
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 2)) // same round as gold 1, added later

	for search, gold := range map[string]int{
		"rat": 3, "rat#1": 3, "1.rat": 3,
		"rat#2": 2, "2.rat": 2,
		"rat#3": 1, "3.rat": 1,
		"rat corpse": 3, "rat corpse#2": 2, "2.rat corpse": 2,
	} {
		c, refusal := r.FindCorpseByRef(nil, search)
		require.NotNil(t, c, search)
		assert.Empty(t, refusal, search)
		assert.Equal(t, gold, c.Gold, search)
		v, ok := r.FindCorpse(nil, search)
		require.True(t, ok, search)
		assert.Equal(t, gold, v.Gold, "FindCorpse %s", search)
	}
	c, refusal := r.FindCorpseByRef(nil, "4.rat")
	assert.Nil(t, c, "three live rat corpses, not four (the crumbled one is not a candidate)")
	assert.Empty(t, refusal)
}

// Players' and NPCs' corpses share one numbering (no player-first ranking),
// and a corpse named exactly wins over one the name only begins.
func TestFindCorpseByRefOneListExactFirst(t *testing.T) {
	r := &Room{}
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 1))
	r.AddCorpse(corpseOf("Ratty", 9, 0, 6, 2))          // a player's, newer
	r.AddCorpse(corpseOf("rat corpsewing", 0, 8, 7, 3)) // newest; "rat corpsewing corpse" begins with "rat corpse"

	c, _ := r.FindCorpseByRef(nil, "rat")
	require.NotNil(t, c)
	assert.Equal(t, 3, c.Gold, "the newest corpse the name begins")
	c, _ = r.FindCorpseByRef(nil, "2.rat")
	require.NotNil(t, c)
	assert.Equal(t, 2, c.Gold, "the player's corpse is the second: one list")
	c, _ = r.FindCorpseByRef(nil, "rat corpse")
	require.NotNil(t, c)
	assert.Equal(t, 1, c.Gold, "exact over a prefix that comes first")
}

// RemoveCorpseRef removes the very corpse it is given, though another has
// the same owner, name and round (RemoveCorpse takes the first of those).
func TestRemoveCorpseRefRemovesTheCorpseFound(t *testing.T) {
	r := &Room{}
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 1))
	r.AddCorpse(corpseOf("rat", 0, 7, 5, 2))

	// "rat" is the newest: of two from one round, the one added later.
	// RemoveCorpse, matching owner, name and round, would take the first.
	c, _ := r.FindCorpseByRef(nil, "rat")
	require.NotNil(t, c)
	require.Equal(t, 2, c.Gold)
	require.True(t, r.RemoveCorpseRef(c))
	require.Len(t, r.Corpses, 1)
	assert.Equal(t, 1, r.Corpses[0].Gold, "the one named went, the other stayed")
	assert.False(t, r.RemoveCorpseRef(&Corpse{}), "a corpse not in the room")
}
