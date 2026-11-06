package buffs

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/statmods"
	"github.com/stretchr/testify/assert"
)

func TestBuffSpec_GetValue(t *testing.T) {
	tests := []struct {
		name        string
		spec        BuffSpec
		expectedVal int
	}{
		{
			name: "No statmods, no flags, no triggercount",
			spec: BuffSpec{
				RoundInterval: 5,
			},
			expectedVal: 0,
		},
		{
			name: "Statmods positive and negative, no flags, no triggercount",
			spec: BuffSpec{
				StatMods:      statmods.StatMods{"str": 3, "dex": -2},
				RoundInterval: 5,
			},
			expectedVal: 5, // |3| + |−2| = 5, freqVal = 0, flags = 0
		},
		{
			name: "Statmods, flags, no triggercount",
			spec: BuffSpec{
				StatMods:      statmods.StatMods{"str": 2},
				Flags:         []string{"no-combat", "hidden"},
				RoundInterval: 3,
			},
			expectedVal: 2 + (5 - 3) + 2*5, // 2 + 2 + 10 = 14
		},
		{
			name: "Statmods, flags, triggercount > 0",
			spec: BuffSpec{
				StatMods:      statmods.StatMods{"str": 1, "dex": 2},
				Flags:         []string{"no-combat"},
				RoundInterval: 2,
				TriggerCount:  3,
			},
			expectedVal: (1 + 2 + (5 - 2) + 1*5) * 3, // (3 + 3 + 5) * 3 = 11 * 3 = 33
		},
		{
			name: "Negative RoundInterval",
			spec: BuffSpec{
				StatMods:      statmods.StatMods{"str": 2},
				Flags:         []string{},
				RoundInterval: 10,
			},
			expectedVal: 2, // freqVal = 5-10 = -5 -> 0, so 2+0+0=2
		},
		{
			name: "Zero RoundInterval",
			spec: BuffSpec{
				StatMods:      statmods.StatMods{},
				Flags:         []string{"no-combat"},
				RoundInterval: 0,
			},
			expectedVal: 0 + 5 + 5, // freqVal = 5-0=5, flags=1*5=5, total=10
		},
		{
			name: "TriggerCount is 1 (should not multiply)",
			spec: BuffSpec{
				StatMods:      statmods.StatMods{"str": 2},
				Flags:         []string{},
				RoundInterval: 4,
				TriggerCount:  1,
			},
			expectedVal: (2 + (5 - 4) + 0) * 1, // (2+1+0)*1=3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.spec.GetValue()
			assert.Equal(t, tt.expectedVal, val)
		})
	}
}
func TestBuffSpec_VisibleNameDesc(t *testing.T) {
	tests := []struct {
		name     string
		spec     BuffSpec
		wantName string
		wantDesc string
	}{
		{
			name: "Secret buff returns mysterious values",
			spec: BuffSpec{
				Secret:      true,
				Name:        "Poison",
				Description: "Deals damage over time",
			},
			wantName: "Mysterious Affliction",
			wantDesc: "Unknown",
		},
		{
			name: "Non-secret buff returns actual name and description",
			spec: BuffSpec{
				Secret:      false,
				Name:        "Fast Healing",
				Description: "Increases health recovery",
			},
			wantName: "Fast Healing",
			wantDesc: "Increases health recovery",
		},
		{
			name: "Empty name and description, not secret",
			spec: BuffSpec{
				Secret:      false,
				Name:        "",
				Description: "",
			},
			wantName: "",
			wantDesc: "",
		},
		{
			name: "Empty name and description, secret",
			spec: BuffSpec{
				Secret:      true,
				Name:        "",
				Description: "",
			},
			wantName: "Mysterious Affliction",
			wantDesc: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotDesc := tt.spec.VisibleNameDesc()
			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantDesc, gotDesc)
		})
	}
}
func TestGetBuffSpec(t *testing.T) {
	// Save and restore original buffs map
	origBuffs := buffs
	defer func() { buffs = origBuffs }()

	// Setup test buffs
	buffs = map[int]*BuffSpec{
		1: {BuffId: 1, Name: "Test Buff 1"},
		2: {BuffId: 2, Name: "Test Buff 2"},
	}

	tests := []struct {
		name      string
		inputId   int
		wantBuff  *BuffSpec
		wantFound bool
	}{
		{
			name:      "Existing positive buffId",
			inputId:   1,
			wantBuff:  &BuffSpec{BuffId: 1, Name: "Test Buff 1"},
			wantFound: true,
		},
		{
			name:      "Existing negative buffId (should convert to positive)",
			inputId:   -2,
			wantBuff:  &BuffSpec{BuffId: 2, Name: "Test Buff 2"},
			wantFound: true,
		},
		{
			name:      "Non-existing buffId",
			inputId:   99,
			wantBuff:  nil,
			wantFound: false,
		},
		{
			name:      "Non-existing negative buffId",
			inputId:   -99,
			wantBuff:  nil,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBuffSpec(tt.inputId)
			if tt.wantFound {
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantBuff.BuffId, got.BuffId)
				assert.Equal(t, tt.wantBuff.Name, got.Name)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}
func TestGetAllBuffIds(t *testing.T) {
	origBuffs := buffs
	defer func() { buffs = origBuffs }()

	tests := []struct {
		name       string
		setupBuffs map[int]*BuffSpec
		wantIds    []int
	}{
		{
			name:       "No buffs",
			setupBuffs: map[int]*BuffSpec{},
			wantIds:    []int{},
		},
		{
			name: "Single buff",
			setupBuffs: map[int]*BuffSpec{
				10: {BuffId: 10, Name: "Solo Buff"},
			},
			wantIds: []int{10},
		},
		{
			name: "Multiple buffs",
			setupBuffs: map[int]*BuffSpec{
				1: {BuffId: 1, Name: "Buff One"},
				2: {BuffId: 2, Name: "Buff Two"},
				3: {BuffId: 3, Name: "Buff Three"},
			},
			wantIds: []int{1, 2, 3},
		},
		{
			name: "Buffs with non-sequential IDs",
			setupBuffs: map[int]*BuffSpec{
				100: {BuffId: 100, Name: "Buff 100"},
				5:   {BuffId: 5, Name: "Buff 5"},
				42:  {BuffId: 42, Name: "Buff 42"},
			},
			wantIds: []int{100, 5, 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffs = tt.setupBuffs
			got := GetAllBuffIds()
			assert.ElementsMatch(t, tt.wantIds, got)
		})
	}
}
func TestBuffSpec_Id(t *testing.T) {
	tests := []struct {
		name   string
		spec   BuffSpec
		wantId int
	}{
		{
			name:   "Positive BuffId",
			spec:   BuffSpec{BuffId: 42},
			wantId: 42,
		},
		{
			name:   "Zero BuffId",
			spec:   BuffSpec{BuffId: 0},
			wantId: 0,
		},
		{
			name:   "Negative BuffId",
			spec:   BuffSpec{BuffId: -7},
			wantId: -7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.Id()
			assert.Equal(t, tt.wantId, got)
		})
	}
}
func TestBuffSpec_Filename(t *testing.T) {
	tests := []struct {
		name     string
		spec     BuffSpec
		expected string
	}{
		{
			name:     "Simple name",
			spec:     BuffSpec{BuffId: 1, Name: "Fast Healing"},
			expected: "1-fast_healing.yaml",
		},
		{
			name:     "Name with special characters",
			spec:     BuffSpec{BuffId: 42, Name: "Poison!@#"},
			expected: "42-poison___.yaml",
		},
		{
			name:     "Name with spaces and mixed case",
			spec:     BuffSpec{BuffId: 7, Name: "Night Vision"},
			expected: "7-night_vision.yaml",
		},
		{
			name:     "Name with underscores and dashes",
			spec:     BuffSpec{BuffId: 100, Name: "Hydrated_buff-test"},
			expected: "100-hydrated_buff_test.yaml",
		},
		{
			name:     "Empty name",
			spec:     BuffSpec{BuffId: 5, Name: ""},
			expected: "5-.yaml",
		},
		{
			name:     "Name with leading/trailing spaces",
			spec:     BuffSpec{BuffId: 8, Name: "  Hidden Buff  "},
			expected: "8-__hidden_buff__.yaml",
		},
		{
			name:     "Name with multiple spaces",
			spec:     BuffSpec{BuffId: 9, Name: "Buff    With   Spaces"},
			expected: "9-buff____with___spaces.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.Filename()
			assert.Equal(t, tt.expected, got)
		})
	}
}
func TestBuffSpec_Filepath(t *testing.T) {
	tests := []struct {
		name     string
		spec     BuffSpec
		expected string
	}{
		{
			name:     "Simple name",
			spec:     BuffSpec{BuffId: 1, Name: "Fast Healing"},
			expected: "1-fast_healing.yaml",
		},
		{
			name:     "Name with special characters",
			spec:     BuffSpec{BuffId: 42, Name: "Poison!@#"},
			expected: "42-poison___.yaml",
		},
		{
			name:     "Name with spaces and mixed case",
			spec:     BuffSpec{BuffId: 7, Name: "Night Vision"},
			expected: "7-night_vision.yaml",
		},
		{
			name:     "Name with underscores and dashes",
			spec:     BuffSpec{BuffId: 100, Name: "Hydrated_buff-test"},
			expected: "100-hydrated_buff_test.yaml",
		},
		{
			name:     "Empty name",
			spec:     BuffSpec{BuffId: 5, Name: ""},
			expected: "5-.yaml",
		},
		{
			name:     "Name with leading/trailing spaces",
			spec:     BuffSpec{BuffId: 8, Name: "  Hidden Buff  "},
			expected: "8-__hidden_buff__.yaml",
		},
		{
			name:     "Name with multiple spaces",
			spec:     BuffSpec{BuffId: 9, Name: "Buff    With   Spaces"},
			expected: "9-buff____with___spaces.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.Filepath()
			assert.Equal(t, tt.expected, got)
		})
	}
}
