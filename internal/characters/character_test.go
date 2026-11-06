package characters

import (
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"maps"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggroHoldsOneTarget(t *testing.T) {
	c := New()
	target := ids.New(ids.NoCell)
	c.SetAggro(target, DefaultAttack)
	require.NotNil(t, c.Aggro)
	assert.Equal(t, target, c.Aggro.Target)
	assert.True(t, c.IsAggro(target))
	assert.False(t, c.IsAggro(ids.New(ids.NoCell)))
}

func TestRecoveringHealthClearsTheKiller(t *testing.T) {
	type way struct{ partial, full func(c *Character) }
	ways := map[string]way{
		"Heal":              {func(c *Character) { c.Heal(0, 0) }, func(c *Character) { c.Heal(5, 0) }},
		"ApplyHealthChange": {func(c *Character) { c.ApplyHealthChange(0) }, func(c *Character) { c.ApplyHealthChange(5) }},
	}
	for name, wy := range ways {
		t.Run(name, func(t *testing.T) {
			c := New()
			c.HealthMax.Value = 10
			c.Health = 0
			c.KilledBy = ids.New(ids.NoCell)
			c.KillerMobIsElite = true
			c.KillerMobName = "wolf"
			c.KillerMobId = 7

			wy.partial(c)
			assert.False(t, c.KilledBy.IsZero(), "still down at %d: the killer stays", c.Health)

			wy.full(c)
			assert.GreaterOrEqual(t, c.Health, 1)
			assert.True(t, c.KilledBy.IsZero())
			assert.False(t, c.KillerMobIsElite)
			assert.Empty(t, c.KillerMobName)
			assert.Zero(t, c.KillerMobId)
		})
	}
}

func TestMain(m *testing.M) {
	// Seed the skills cache so skill-gated helpers (GetSkillLevel)
	// behave deterministically without loading datafiles from disk.
	seedSkills := []*skills.Skill{}
	for _, id := range []string{
		`cast`, `dual-wield`, `map`, `enchant`, `peep`,
		`inspect`, `portal`, `search`, `track`, `skulduggery`,
		`brawling`, `scribe`, `protection`, `tame`, `trading`,
		`changeform`,
	} {
		seedSkills = append(seedSkills, &skills.Skill{SkillId: id, Name: id, Description: id, MaxLevel: 4})
	}
	skills.SetTestData(seedSkills)

	os.Exit(m.Run())
}

func TestCharacter_XPTLClampsToMaxInt(t *testing.T) {
	c := New()

	got := c.XPTL(200_000_000)

	assert.Equal(t, math.MaxInt, got)
}

func TestCharacter_CanDualWield(t *testing.T) {
	tests := []struct {
		name       string
		skillLevel int
		want       bool
	}{
		{
			name:       "No DualWield skill",
			skillLevel: 0,
			want:       false,
		},
		{
			name:       "DualWield skill at level 1",
			skillLevel: 1,
			want:       true,
		},
		{
			name:       "DualWield skill at level 2",
			skillLevel: 2,
			want:       true,
		},
		{
			name:       "DualWield skill at level 4",
			skillLevel: 4,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.Skills[`dual-wield`] = tt.skillLevel
			got := c.CanDualWield()
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestCharacter_GetMiscDataKeys(t *testing.T) {
	tests := []struct {
		name        string
		miscData    map[string]any
		prefixMatch []string
		want        []string
	}{
		{
			name:        "No misc data, no prefix",
			miscData:    nil,
			prefixMatch: nil,
			want:        []string{},
		},
		{
			name:        "No misc data, with prefix",
			miscData:    nil,
			prefixMatch: []string{"foo"},
			want:        []string{},
		},
		{
			name: "Misc data, no prefix",
			miscData: map[string]any{
				"foo1": 1,
				"bar2": 2,
				"baz3": 3,
			},
			prefixMatch: nil,
			want:        []string{"foo1", "bar2", "baz3"},
		},
		{
			name: "Misc data, prefix matches one key",
			miscData: map[string]any{
				"foo1": 1,
				"bar2": 2,
				"baz3": 3,
			},
			prefixMatch: []string{"foo"},
			want:        []string{"1"},
		},
		{
			name: "Misc data, prefix matches multiple keys",
			miscData: map[string]any{
				"foo1": 1,
				"foo2": 2,
				"bar3": 3,
			},
			prefixMatch: []string{"foo"},
			want:        []string{"1", "2"},
		},
		{
			name: "Misc data, prefix matches no keys",
			miscData: map[string]any{
				"foo1": 1,
				"bar2": 2,
			},
			prefixMatch: []string{"baz"},
			want:        []string{},
		},
		{
			name: "Misc data, multiple prefixes",
			miscData: map[string]any{
				"foo1": 1,
				"bar2": 2,
				"baz3": 3,
				"foo4": 4,
				"bar5": 5,
			},
			prefixMatch: []string{"foo", "bar"},
			want:        []string{"1", "4", "2", "5"},
		},
		{
			name: "Misc data, prefix is full key",
			miscData: map[string]any{
				"foo":    1,
				"foobar": 2,
			},
			prefixMatch: []string{"foo"},
			want:        []string{"", "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.MiscData = tt.miscData
			got := c.GetMiscDataKeys(tt.prefixMatch...)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
func TestCharacter_CarryCapacity(t *testing.T) {
	tests := []struct {
		name        string
		strengthAdj int
		expectedCap int
	}{
		{
			name:        "Strength 0",
			strengthAdj: 0,
			expectedCap: 5,
		},
		{
			name:        "Strength 2",
			strengthAdj: 2,
			expectedCap: 5,
		},
		{
			name:        "Strength 3",
			strengthAdj: 3,
			expectedCap: 6,
		},
		{
			name:        "Strength 6",
			strengthAdj: 6,
			expectedCap: 7,
		},
		{
			name:        "Strength 9",
			strengthAdj: 9,
			expectedCap: 8,
		},
		{
			name:        "Strength 12",
			strengthAdj: 12,
			expectedCap: 9,
		},
		{
			name:        "Strength 30",
			strengthAdj: 30,
			expectedCap: 15,
		},
		{
			name:        "Strength 100",
			strengthAdj: 100,
			expectedCap: 38,
		},
		{
			name:        "Negative Strength",
			strengthAdj: -3,
			expectedCap: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.Stats.Strength.ValueAdj = tt.strengthAdj
			got := c.CarryCapacity()
			assert.Equal(t, tt.expectedCap, got)
		})
	}
}

func TestCharacter_DeductActionPoints(t *testing.T) {
	tests := []struct {
		name         string
		startAP      int
		deductAmount int
		wantResult   bool
		wantFinalAP  int
	}{
		{
			name:         "Enough action points",
			startAP:      10,
			deductAmount: 5,
			wantResult:   true,
			wantFinalAP:  5,
		},
		{
			name:         "Exact action points",
			startAP:      5,
			deductAmount: 5,
			wantResult:   true,
			wantFinalAP:  0,
		},
		{
			name:         "Not enough action points",
			startAP:      3,
			deductAmount: 5,
			wantResult:   false,
			wantFinalAP:  3,
		},
		{
			name:         "Deduct more than available, negative AP clamp",
			startAP:      2,
			deductAmount: 3,
			wantResult:   false,
			wantFinalAP:  2,
		},
		{
			name:         "Deduct zero action points",
			startAP:      7,
			deductAmount: 0,
			wantResult:   true,
			wantFinalAP:  7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.ActionPoints = tt.startAP
			got := c.DeductActionPoints(tt.deductAmount)
			assert.Equal(t, tt.wantResult, got)
			assert.Equal(t, tt.wantFinalAP, c.ActionPoints)
		})
	}
}
func TestCharacter_SetUserId(t *testing.T) {
	tests := []struct {
		name   string
		userId int
	}{
		{
			name:   "Set positive userId",
			userId: 42,
		},
		{
			name:   "Set zero userId",
			userId: 0,
		},
		{
			name:   "Set negative userId",
			userId: -7,
		},
		{
			name:   "Set large userId",
			userId: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.SetUserId(tt.userId)
			// userId is unexported, so we need to use reflection to check it
			userIdField := reflect.ValueOf(c).Elem().FieldByName("userId")
			assert.True(t, userIdField.IsValid(), "userId field should exist")
			assert.Equal(t, tt.userId, int(userIdField.Int()))
		})
	}
}
func TestCharacter_SetMiscData(t *testing.T) {
	tests := []struct {
		name         string
		initialData  map[string]any
		key          string
		value        any
		expectExists bool
		expectValue  any
	}{
		{
			name:         "Set new key-value pair",
			initialData:  nil,
			key:          "foo",
			value:        123,
			expectExists: true,
			expectValue:  123,
		},
		{
			name:         "Overwrite existing key",
			initialData:  map[string]any{"bar": "baz"},
			key:          "bar",
			value:        "qux",
			expectExists: true,
			expectValue:  "qux",
		},
		{
			name:         "Delete existing key by setting value to nil",
			initialData:  map[string]any{"del": 42},
			key:          "del",
			value:        nil,
			expectExists: false,
			expectValue:  nil,
		},
		{
			name:         "Delete non-existing key by setting value to nil",
			initialData:  map[string]any{"other": 1},
			key:          "missing",
			value:        nil,
			expectExists: false,
			expectValue:  nil,
		},
		{
			name:         "Set key with value zero",
			initialData:  nil,
			key:          "zero",
			value:        0,
			expectExists: true,
			expectValue:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initialData != nil {
				c.MiscData = make(map[string]any)
				for k, v := range tt.initialData {
					c.MiscData[k] = v
				}
			}
			c.SetMiscData(tt.key, tt.value)
			val, ok := c.MiscData[tt.key]
			assert.Equal(t, tt.expectExists, ok)
			assert.Equal(t, tt.expectValue, val)
		})
	}
}
func TestCharacter_GetMiscData(t *testing.T) {
	tests := []struct {
		name        string
		initialData map[string]any
		key         string
		want        any
	}{
		{
			name:        "Get existing key",
			initialData: map[string]any{"foo": 123, "bar": "baz"},
			key:         "foo",
			want:        123,
		},
		{
			name:        "Get another existing key",
			initialData: map[string]any{"foo": 123, "bar": "baz"},
			key:         "bar",
			want:        "baz",
		},
		{
			name:        "Get non-existing key returns nil",
			initialData: map[string]any{"foo": 123},
			key:         "missing",
			want:        nil,
		},
		{
			name:        "Get key from nil map initializes map and returns nil",
			initialData: nil,
			key:         "anything",
			want:        nil,
		},
		{
			name:        "Get key with value zero",
			initialData: map[string]any{"zero": 0},
			key:         "zero",
			want:        0,
		},
		{
			name:        "Get key with value false",
			initialData: map[string]any{"flag": false},
			key:         "flag",
			want:        false,
		},
		{
			name:        "Get key with value nil explicitly set",
			initialData: map[string]any{"nilval": nil},
			key:         "nilval",
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initialData != nil {
				c.MiscData = make(map[string]any)
				for k, v := range tt.initialData {
					c.MiscData[k] = v
				}
			} else {
				c.MiscData = nil
			}
			got := c.GetMiscData(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestCharacter_KeyCount(t *testing.T) {
	tests := []struct {
		name    string
		keyRing map[string]string
		want    int
	}{
		{
			name:    "Nil KeyRing",
			keyRing: nil,
			want:    0,
		},
		{
			name:    "Empty KeyRing",
			keyRing: map[string]string{},
			want:    0,
		},
		{
			name:    "One key in KeyRing",
			keyRing: map[string]string{"lock1": "SEQ1"},
			want:    1,
		},
		{
			name:    "Multiple keys in KeyRing",
			keyRing: map[string]string{"lock1": "SEQ1", "lock2": "SEQ2", "lock3": "SEQ3"},
			want:    3,
		},
		{
			name:    "KeyRing with empty string key",
			keyRing: map[string]string{"": "SEQ"},
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.KeyRing = tt.keyRing
			got := c.KeyCount()
			assert.Equal(t, tt.want, got)
			// Also check that KeyRing is not nil after call
			assert.NotNil(t, c.KeyRing)
		})
	}
}
func TestCharacter_GetKey(t *testing.T) {
	tests := []struct {
		name      string
		keyRing   map[string]string
		lockId    string
		wantValue string
	}{
		{
			name:      "Nil KeyRing returns empty string",
			keyRing:   nil,
			lockId:    "lock1",
			wantValue: "",
		},
		{
			name:      "Empty KeyRing returns empty string",
			keyRing:   map[string]string{},
			lockId:    "lock1",
			wantValue: "",
		},
		{
			name:      "Key exists with exact case",
			keyRing:   map[string]string{"lock1": "SEQ1"},
			lockId:    "lock1",
			wantValue: "SEQ1",
		},
		{
			name:      "Key exists with different case",
			keyRing:   map[string]string{"lock1": "SEQ1"},
			lockId:    "LOCK1",
			wantValue: "SEQ1",
		},
		{
			name:      "Multiple keys, get correct one",
			keyRing:   map[string]string{"lock1": "SEQ1", "lock2": "SEQ2"},
			lockId:    "lock2",
			wantValue: "SEQ2",
		},
		{
			name:      "Key does not exist returns empty string",
			keyRing:   map[string]string{"lock1": "SEQ1"},
			lockId:    "lock3",
			wantValue: "",
		},
		{
			name:      "Key exists with empty string value",
			keyRing:   map[string]string{"lock1": ""},
			lockId:    "lock1",
			wantValue: "",
		},
		{
			name:      "Key exists with empty string key",
			keyRing:   map[string]string{"": "EMPTY"},
			lockId:    "",
			wantValue: "EMPTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.KeyRing = tt.keyRing
			got := c.GetKey(tt.lockId)
			assert.Equal(t, tt.wantValue, got)
			assert.NotNil(t, c.KeyRing)
		})
	}
}
func TestCharacter_SetKey(t *testing.T) {
	tests := []struct {
		name         string
		initialKeys  map[string]string
		lockId       string
		sequence     string
		expectKey    string
		expectExists bool
	}{
		{
			name:         "Set new key with non-empty sequence",
			initialKeys:  nil,
			lockId:       "LockA",
			sequence:     "abc123",
			expectKey:    "ABC123",
			expectExists: true,
		},
		{
			name:         "Set new key with empty sequence (should not exist)",
			initialKeys:  nil,
			lockId:       "LockB",
			sequence:     "",
			expectKey:    "",
			expectExists: false,
		},
		{
			name:         "Overwrite existing key with new sequence",
			initialKeys:  map[string]string{"lockc": "OLDSEQ"},
			lockId:       "LockC",
			sequence:     "newseq",
			expectKey:    "NEWSEQ",
			expectExists: true,
		},
		{
			name:         "Delete existing key by setting empty sequence",
			initialKeys:  map[string]string{"lockd": "SOMETHING"},
			lockId:       "LockD",
			sequence:     "",
			expectKey:    "",
			expectExists: false,
		},
		{
			name:         "Set key with mixed case lockId, should store as lower",
			initialKeys:  nil,
			lockId:       "MiXeDcAsE",
			sequence:     "seq",
			expectKey:    "SEQ",
			expectExists: true,
		},
		{
			name:         "Set key with empty lockId",
			initialKeys:  nil,
			lockId:       "",
			sequence:     "emptykey",
			expectKey:    "EMPTYKEY",
			expectExists: true,
		},
		{
			name:         "Delete non-existing key with empty sequence",
			initialKeys:  map[string]string{"other": "VAL"},
			lockId:       "notfound",
			sequence:     "",
			expectKey:    "",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initialKeys != nil {
				c.KeyRing = make(map[string]string)
				for k, v := range tt.initialKeys {
					c.KeyRing[k] = v
				}
			} else {
				c.KeyRing = nil
			}
			c.SetKey(tt.lockId, tt.sequence)
			got, ok := c.KeyRing[strings.ToLower(tt.lockId)]
			assert.Equal(t, tt.expectExists, ok)
			assert.Equal(t, tt.expectKey, got)
		})
	}
}

func TestCharacter_HasAdjective(t *testing.T) {
	tests := []struct {
		name       string
		adjectives []string
		checkAdj   string
		want       bool
	}{
		{
			name:       "No adjectives, check any",
			adjectives: nil,
			checkAdj:   "sleeping",
			want:       false,
		},
		{
			name:       "Empty adjectives slice, check any",
			adjectives: []string{},
			checkAdj:   "dead",
			want:       false,
		},
		{
			name:       "Single adjective, present",
			adjectives: []string{"wounded"},
			checkAdj:   "wounded",
			want:       true,
		},
		{
			name:       "Single adjective, not present",
			adjectives: []string{"wounded"},
			checkAdj:   "sleeping",
			want:       false,
		},
		{
			name:       "Multiple adjectives, present",
			adjectives: []string{"sleeping", "dead", "wounded"},
			checkAdj:   "dead",
			want:       true,
		},
		{
			name:       "Multiple adjectives, not present",
			adjectives: []string{"sleeping", "dead", "wounded"},
			checkAdj:   "awake",
			want:       false,
		},
		{
			name:       "Adjective is empty string, present",
			adjectives: []string{"", "foo"},
			checkAdj:   "",
			want:       true,
		},
		{
			name:       "Adjective is empty string, not present",
			adjectives: []string{"foo", "bar"},
			checkAdj:   "",
			want:       false,
		},
		{
			name:       "Case sensitive match",
			adjectives: []string{"Sleeping"},
			checkAdj:   "sleeping",
			want:       false,
		},
		{
			name:       "Case sensitive exact match",
			adjectives: []string{"sleeping"},
			checkAdj:   "sleeping",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.Adjectives = tt.adjectives
			got := c.HasAdjective(tt.checkAdj)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestCharacter_SetAdjective(t *testing.T) {
	tests := []struct {
		name      string
		initial   []string
		adj       string
		addToList bool
		want      []string
	}{
		{
			name:      "Add adjective to empty list",
			initial:   nil,
			adj:       "sleeping",
			addToList: true,
			want:      []string{"sleeping"},
		},
		{
			name:      "Add adjective to non-empty list",
			initial:   []string{"wounded"},
			adj:       "sleeping",
			addToList: true,
			want:      []string{"wounded", "sleeping"},
		},
		{
			name:      "Add duplicate adjective (should not duplicate)",
			initial:   []string{"sleeping"},
			adj:       "sleeping",
			addToList: true,
			want:      []string{"sleeping"},
		},
		{
			name:      "Remove adjective from list",
			initial:   []string{"sleeping", "wounded"},
			adj:       "sleeping",
			addToList: false,
			want:      []string{"wounded"},
		},
		{
			name:      "Remove adjective not present (should do nothing)",
			initial:   []string{"wounded"},
			adj:       "sleeping",
			addToList: false,
			want:      []string{"wounded"},
		},
		{
			name:      "Remove last adjective (should result in empty list)",
			initial:   []string{"sleeping"},
			adj:       "sleeping",
			addToList: false,
			want:      []string{},
		},
		{
			name:      "Add multiple adjectives sequentially",
			initial:   []string{},
			adj:       "sleeping",
			addToList: true,
			want:      []string{"sleeping"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initial != nil {
				c.Adjectives = append([]string{}, tt.initial...)
			} else {
				c.Adjectives = nil
			}
			c.SetAdjective(tt.adj, tt.addToList)
			assert.Equal(t, tt.want, c.Adjectives)
		})
	}

	t.Run("Add then remove adjective", func(t *testing.T) {
		c := New()
		c.SetAdjective("sleeping", true)
		assert.Equal(t, []string{"sleeping"}, c.Adjectives)
		c.SetAdjective("sleeping", false)
		assert.Equal(t, []string{}, c.Adjectives)
	})

	t.Run("Remove from nil adjectives", func(t *testing.T) {
		c := New()
		c.Adjectives = nil
		c.SetAdjective("sleeping", false)
		assert.Equal(t, []string{}, c.Adjectives)
	})

	t.Run("Add to nil adjectives", func(t *testing.T) {
		c := New()
		c.Adjectives = nil
		c.SetAdjective("sleeping", true)
		assert.Equal(t, []string{"sleeping"}, c.Adjectives)
	})
}
func TestCharacter_PruneCooldowns(t *testing.T) {
	type cooldownsCase struct {
		name         string
		initial      Cooldowns
		expectPruned Cooldowns
	}
	tests := []cooldownsCase{
		{
			name:         "Nil Cooldowns map does nothing",
			initial:      nil,
			expectPruned: nil,
		},
		{
			name:         "Empty Cooldowns map does nothing",
			initial:      Cooldowns{},
			expectPruned: Cooldowns{},
		},
		{
			name: "Cooldowns with zero and positive values",
			initial: Cooldowns{
				"foo": 0,
				"bar": 2,
				"baz": 0,
			},
			expectPruned: Cooldowns{
				"bar": 2,
			},
		},
		{
			name: "Cooldowns with all zero values",
			initial: Cooldowns{
				"a": 0,
				"b": 0,
			},
			expectPruned: Cooldowns{},
		},
		{
			name: "Cooldowns with all positive values",
			initial: Cooldowns{
				"x": 1,
				"y": 2,
			},
			expectPruned: Cooldowns{
				"x": 1,
				"y": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initial != nil {
				c.Cooldowns = make(Cooldowns)
				for k, v := range tt.initial {
					c.Cooldowns[k] = v
				}
			} else {
				c.Cooldowns = nil
			}
			c.PruneCooldowns()
			if tt.expectPruned == nil {
				assert.Nil(t, c.Cooldowns)
			} else {
				assert.Equal(t, tt.expectPruned, c.Cooldowns)
			}
		})
	}
}
func TestCharacter_GetCooldown(t *testing.T) {
	tests := []struct {
		name         string
		cooldowns    map[string]int
		trackingTag  string
		wantCooldown int
	}{
		{
			name:         "Nil Cooldowns map returns 0 and initializes map",
			cooldowns:    nil,
			trackingTag:  "foo",
			wantCooldown: 0,
		},
		{
			name:         "Empty Cooldowns map returns 0",
			cooldowns:    map[string]int{},
			trackingTag:  "bar",
			wantCooldown: 0,
		},
		{
			name:         "Cooldown exists for tag",
			cooldowns:    map[string]int{"baz": 7, "qux": 3},
			trackingTag:  "baz",
			wantCooldown: 7,
		},
		{
			name:         "Cooldown exists for another tag",
			cooldowns:    map[string]int{"baz": 7, "qux": 3},
			trackingTag:  "qux",
			wantCooldown: 3,
		},
		{
			name:         "Cooldown does not exist for tag returns 0",
			cooldowns:    map[string]int{"baz": 7},
			trackingTag:  "notfound",
			wantCooldown: 0,
		},
		{
			name:         "Cooldown with zero value",
			cooldowns:    map[string]int{"zero": 0},
			trackingTag:  "zero",
			wantCooldown: 0,
		},
		{
			name:         "Cooldown with negative value",
			cooldowns:    map[string]int{"neg": -5},
			trackingTag:  "neg",
			wantCooldown: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.cooldowns != nil {
				c.Cooldowns = make(Cooldowns)
				for k, v := range tt.cooldowns {
					c.Cooldowns[k] = v
				}
			} else {
				c.Cooldowns = nil
			}
			got := c.GetCooldown(tt.trackingTag)
			assert.Equal(t, tt.wantCooldown, got)
			assert.NotNil(t, c.Cooldowns, "Cooldowns map should be initialized")
		})
	}
}
func TestCharacter_GetAllCooldowns(t *testing.T) {
	tests := []struct {
		name      string
		cooldowns map[string]int
		want      map[string]int
	}{
		{
			name:      "Nil Cooldowns returns empty map",
			cooldowns: nil,
			want:      map[string]int{},
		},
		{
			name:      "Empty Cooldowns returns empty map",
			cooldowns: map[string]int{},
			want:      map[string]int{},
		},
		{
			name:      "Single cooldown",
			cooldowns: map[string]int{"attack": 3},
			want:      map[string]int{"attack": 3},
		},
		{
			name:      "Multiple cooldowns",
			cooldowns: map[string]int{"attack": 3, "cast": 5, "move": 1},
			want:      map[string]int{"attack": 3, "cast": 5, "move": 1},
		},
		{
			name:      "Cooldown with zero value",
			cooldowns: map[string]int{"wait": 0},
			want:      map[string]int{"wait": 0},
		},
		{
			name:      "Cooldown with negative value",
			cooldowns: map[string]int{"buggy": -2},
			want:      map[string]int{"buggy": -2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.cooldowns != nil {
				c.Cooldowns = make(Cooldowns)
				for k, v := range tt.cooldowns {
					c.Cooldowns[k] = v
				}
			} else {
				c.Cooldowns = nil
			}
			got := c.GetAllCooldowns()
			assert.Equal(t, tt.want, got)
			// Ensure returned map is a copy, not the same reference
			if c.Cooldowns != nil {
				got["newcd"] = 99
				_, exists := c.Cooldowns["newcd"]
				assert.False(t, exists, "GetAllCooldowns should return a copy, not the original map")
			}
		})
	}
}
func TestCharacter_SetSetting(t *testing.T) {
	tests := []struct {
		name         string
		initial      map[string]string
		settingName  string
		settingValue string
		want         map[string]string
	}{
		{
			name:         "Set new setting",
			initial:      nil,
			settingName:  "color",
			settingValue: "blue",
			want:         map[string]string{"color": "blue"},
		},
		{
			name:         "Overwrite existing setting",
			initial:      map[string]string{"color": "red"},
			settingName:  "color",
			settingValue: "green",
			want:         map[string]string{"color": "green"},
		},
		{
			name:         "Add another setting",
			initial:      map[string]string{"volume": "high"},
			settingName:  "brightness",
			settingValue: "low",
			want:         map[string]string{"volume": "high", "brightness": "low"},
		},
		{
			name:         "Delete existing setting by setting value to empty",
			initial:      map[string]string{"foo": "bar"},
			settingName:  "foo",
			settingValue: "",
			want:         map[string]string{},
		},
		{
			name:         "Delete non-existing setting by setting value to empty",
			initial:      map[string]string{"foo": "bar"},
			settingName:  "baz",
			settingValue: "",
			want:         map[string]string{"foo": "bar"},
		},
		{
			name:         "Set setting with empty name",
			initial:      nil,
			settingName:  "",
			settingValue: "someval",
			want:         map[string]string{"": "someval"},
		},
		{
			name:         "Delete setting with empty name",
			initial:      map[string]string{"": "someval", "other": "val"},
			settingName:  "",
			settingValue: "",
			want:         map[string]string{"other": "val"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initial != nil {
				c.Settings = make(map[string]string)
				for k, v := range tt.initial {
					c.Settings[k] = v
				}
			} else {
				c.Settings = nil
			}
			c.SetSetting(tt.settingName, tt.settingValue)
			assert.Equal(t, tt.want, c.Settings)
		})
	}
}
func TestCharacter_GetSetting(t *testing.T) {
	tests := []struct {
		name        string
		initial     map[string]string
		settingName string
		want        string
	}{
		{
			name:        "Nil Settings returns empty string",
			initial:     nil,
			settingName: "foo",
			want:        "",
		},
		{
			name:        "Empty Settings returns empty string",
			initial:     map[string]string{},
			settingName: "foo",
			want:        "",
		},
		{
			name:        "Setting exists returns value",
			initial:     map[string]string{"foo": "bar"},
			settingName: "foo",
			want:        "bar",
		},
		{
			name:        "Setting does not exist returns empty string",
			initial:     map[string]string{"foo": "bar"},
			settingName: "baz",
			want:        "",
		},
		{
			name:        "Multiple settings, get correct one",
			initial:     map[string]string{"foo": "bar", "baz": "qux"},
			settingName: "baz",
			want:        "qux",
		},
		{
			name:        "Setting exists with empty value",
			initial:     map[string]string{"empty": ""},
			settingName: "empty",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.initial != nil {
				c.Settings = make(map[string]string)
				for k, v := range tt.initial {
					c.Settings[k] = v
				}
			} else {
				c.Settings = nil
			}
			got := c.GetSetting(tt.settingName)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestCharacter_IsDisabled(t *testing.T) {
	tests := []struct {
		name   string
		health int
		want   bool
	}{
		{
			name:   "Health is positive",
			health: 10,
			want:   false,
		},
		{
			name:   "Health is zero",
			health: 0,
			want:   true,
		},
		{
			name:   "Health is negative",
			health: -1,
			want:   true,
		},
		{
			name:   "Health is large positive",
			health: 100,
			want:   false,
		},
		{
			name:   "Health is large negative",
			health: -100,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.Health = tt.health
			got := c.IsDisabled()
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestCharacter_EndAggro(t *testing.T) {
	tests := []struct {
		name       string
		setupAggro bool
	}{
		{
			name:       "Aggro is nil, call EndAggro",
			setupAggro: false,
		},
		{
			name:       "Aggro is set, call EndAggro",
			setupAggro: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.setupAggro {
				c.Aggro = &Aggro{
					Target:        ids.New(ids.NoCell),
					Type:          DefaultAttack,
					RoundsWaiting: 3,
				}
			} else {
				c.Aggro = nil
			}
			c.EndAggro()
			assert.Nil(t, c.Aggro)
		})
	}
}
func TestCharacter_GetSkillLevel(t *testing.T) {
	type args struct {
		skillsMap map[string]int
		skillTag  string
	}
	tests := []struct {
		name     string
		args     args
		expected int
	}{
		{
			name: "Skill exists with positive value",
			args: args{
				skillsMap: map[string]int{`dual-wield`: 3},
				skillTag:  `dual-wield`,
			},
			expected: 3,
		},
		{
			name: "Skill exists with zero value",
			args: args{
				skillsMap: map[string]int{`cast`: 0},
				skillTag:  `cast`,
			},
			expected: 0,
		},
		{
			name: "Skill does not exist",
			args: args{
				skillsMap: map[string]int{`tame`: 2},
				skillTag:  `cast`,
			},
			expected: 0,
		},
		{
			name: "Nil Skills map",
			args: args{
				skillsMap: nil,
				skillTag:  `map`,
			},
			expected: 0,
		},
		{
			name: "Multiple skills, get correct one",
			args: args{
				skillsMap: map[string]int{
					`cast`:       2,
					`dual-wield`: 1,
					`map`:        4,
				},
				skillTag: `map`,
			},
			expected: 4,
		},
		{
			name: "Skill exists with negative value",
			args: args{
				skillsMap: map[string]int{`cast`: -2},
				skillTag:  `cast`,
			},
			expected: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			if tt.args.skillsMap != nil {
				c.Skills = make(map[string]int)
				maps.Copy(c.Skills, tt.args.skillsMap)
			} else {
				c.Skills = nil
			}
			got := c.GetSkillLevel(tt.args.skillTag)
			assert.Equal(t, tt.expected, got)
		})
	}
}
