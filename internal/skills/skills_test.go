package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      Skill
		wantErr bool
		wantId  string
		wantMax int
	}{
		{
			name:    "normalizes id and defaults maxlevel",
			in:      Skill{SkillId: "  Brawling ", Name: "Brawling", Description: "desc"},
			wantId:  "brawling",
			wantMax: DefaultMaxLevel,
		},
		{
			name:    "keeps explicit maxlevel",
			in:      Skill{SkillId: "cast", Name: "Cast", Description: "desc", MaxLevel: 6},
			wantId:  "cast",
			wantMax: 6,
		},
		{
			name:    "allows hyphenated id",
			in:      Skill{SkillId: "dual-wield", Name: "Dual Wield", Description: "desc"},
			wantId:  "dual-wield",
			wantMax: DefaultMaxLevel,
		},
		{
			name:    "rejects id with space",
			in:      Skill{SkillId: "dual wield", Name: "Dual Wield", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "rejects id starting with digit",
			in:      Skill{SkillId: "1skill", Name: "Skill", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "rejects empty name",
			in:      Skill{SkillId: "cast", Name: "", Description: "desc"},
			wantErr: true,
		},
		{
			name:    "rejects empty description",
			in:      Skill{SkillId: "cast", Name: "Cast", Description: ""},
			wantErr: true,
		},
		{
			name:    "negative maxlevel defaults",
			in:      Skill{SkillId: "cast", Name: "Cast", Description: "desc", MaxLevel: -3},
			wantId:  "cast",
			wantMax: DefaultMaxLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.in
			err := s.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantId, s.SkillId)
			assert.Equal(t, tt.wantMax, s.MaxLevel)
		})
	}
}

func TestFilename(t *testing.T) {
	s := &Skill{SkillId: "dual-wield"}
	assert.Equal(t, "dual-wield.yaml", s.Filename())
}

func seedStandard(t *testing.T) {
	t.Helper()
	SetTestData(
		[]*Skill{
			{SkillId: "brawling", Name: "Brawling", Description: "d", MaxLevel: 4},
			{SkillId: "dual-wield", Name: "Dual Wield", Description: "d", MaxLevel: 4},
			{SkillId: "cast", Name: "Cast", Description: "d", MaxLevel: 4},
			{SkillId: "enchant", Name: "Enchant", Description: "d", MaxLevel: 4},
			{SkillId: "deepskill", Name: "Deep Skill", Description: "d", MaxLevel: 6},
		},
	)
}

func TestSkillExistsAndMaxLevel(t *testing.T) {
	seedStandard(t)
	assert.True(t, SkillExists("brawling"))
	assert.True(t, SkillExists("BRAWLING"))
	assert.False(t, SkillExists("nonexistent"))

	assert.Equal(t, 4, MaxSkillLevel("brawling"))
	assert.Equal(t, 6, MaxSkillLevel("deepskill"))
	// Unknown skill falls back to default.
	assert.Equal(t, DefaultMaxLevel, MaxSkillLevel("nonexistent"))
}
