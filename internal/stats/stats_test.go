package stats

import (
	"math"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/stretchr/testify/assert"
)

// defaultProgression sets all ProgressionConfig fields to their documented
// defaults so tests run independently of any loaded config file.
func defaultProgression() {
	_ = configs.SetVal("GamePlay.Progression.BaseModFactor", "0.3333333334")
	_ = configs.SetVal("GamePlay.Progression.BaseModExponent", "1.0")
	_ = configs.SetVal("GamePlay.Progression.NaturalGainsModFactor", "0.5")
	_ = configs.SetVal("GamePlay.Progression.NaturalGainsExponent", "1.0")
	_ = configs.SetVal("GamePlay.Progression.HPBase", "5")
	_ = configs.SetVal("GamePlay.Progression.HPPerLevel", "1.0")
	_ = configs.SetVal("GamePlay.Progression.HPPerVitality", "4.0")
	_ = configs.SetVal("GamePlay.Progression.ManaBase", "4")
	_ = configs.SetVal("GamePlay.Progression.ManaPerLevel", "1.0")
	_ = configs.SetVal("GamePlay.Progression.ManaPerMysticism", "3.0")
	_ = configs.SetVal("GamePlay.Progression.TrainingPointsPerLevel", "1")
	_ = configs.SetVal("GamePlay.Progression.TrainingPointsEveryNLevels", "1")
	_ = configs.SetVal("GamePlay.Progression.StatPointsPerLevel", "1")
	_ = configs.SetVal("GamePlay.Progression.StatPointsEveryNLevels", "1")
	_ = configs.SetVal("GamePlay.Progression.XPBase", "1000")
	_ = configs.SetVal("GamePlay.Progression.XPLevelFactor", "0.75")
	_ = configs.SetVal("GamePlay.Progression.XPLevelPower", "2.0")
	_ = configs.SetVal("GamePlay.Progression.MaxLevel", "100")
	_ = configs.SetVal("GamePlay.Progression.StatCapThreshold", "105")
	_ = configs.SetVal("GamePlay.Progression.StatCapAnchor", "100")
	_ = configs.SetVal("GamePlay.Progression.StatCapExponent", "0.5")
	_ = configs.SetVal("GamePlay.Progression.StatCapScale", "2.0")
	_ = configs.SetVal("GamePlay.Progression.StatCapExemptBonus", "false")
}

func TestStatInfo_GainsForLevel(t *testing.T) {
	defaultProgression()

	// With exponent=1.0 (linear), the formula reduces to the previous constants:
	//   basePoints = int((level-1) * 0.3333 * base)
	//   freePoints = int(level * 0.5)
	tests := []struct {
		name     string
		base     int
		level    int
		expected int
	}{
		{name: "level 1 base 10", base: 10, level: 1, expected: 0},
		{name: "level 1 base 0", base: 0, level: 1, expected: 0},
		{name: "level 0 clamps to 1", base: 10, level: 0, expected: 0},
		{name: "negative level clamps to 1", base: 10, level: -5, expected: 0},
		// level 2, base 10: basePoints=int(1*0.3333*10)=3, free=int(2*0.5)=1 => 4
		{name: "level 2 base 10", base: 10, level: 2, expected: 4},
		// level 10, base 10: basePoints=int(9*0.3333333334*10)=int(30.0)=30, free=int(10*0.5)=5 => 35
		{name: "level 10 base 10", base: 10, level: 10, expected: 35},
		// level 10, base 0: basePoints=0, free=5 => 5
		{name: "level 10 base 0", base: 0, level: 10, expected: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := StatInfo{Base: tt.base}
			assert.Equal(t, tt.expected, si.GainsForLevel(tt.level))
		})
	}
}

func TestStatInfo_Recalculate_BelowCap(t *testing.T) {
	defaultProgression()
	si := StatInfo{Base: 10, Training: 5}
	si.Recalculate(10)

	expectedRacial := si.GainsForLevel(10)
	assert.Equal(t, expectedRacial, si.Racial)
	assert.Equal(t, expectedRacial+5, si.Value)
	assert.Equal(t, si.Value, si.ValueAdj, "ValueAdj should equal Value when below cap")
}

func TestStatInfo_Recalculate_WithMods(t *testing.T) {
	defaultProgression()
	si := StatInfo{Base: 10, Training: 3, Mods: 7}
	si.Recalculate(5)

	assert.Equal(t, si.Racial+3+7, si.Value)
}

func TestStatInfo_Recalculate_AboveCap(t *testing.T) {
	defaultProgression()
	// Force a Value above the cap threshold by using high training.
	// With defaults: threshold=105, anchor=100, exponent=0.5, scale=2.0
	// Value = 200, overage = 200-100 = 100, ValueAdj = 100 + round(sqrt(100)*2) = 120
	si := StatInfo{Base: 10, Training: 200}
	si.Recalculate(1)

	// Value = Racial + 200; Racial at level 1 = 0, so Value = 200
	assert.Equal(t, 200, si.Value)

	// Use the actual config defaults to compute expected ValueAdj.
	cfg := configs.GetProgressionConfig()
	overage := 200 - int(cfg.StatCapAnchor)
	expectedAdj := int(cfg.StatCapAnchor) + int(math.Round(math.Pow(float64(overage), float64(cfg.StatCapExponent))*float64(cfg.StatCapScale)))
	assert.Equal(t, expectedAdj, si.ValueAdj)
	assert.True(t, si.ValueAdj < si.Value, "ValueAdj should be less than Value when above cap")
}

func TestStatInfo_Recalculate_ExactlyAtCap(t *testing.T) {
	defaultProgression()
	// Value == threshold triggers the cap branch.
	// With defaults: threshold=105, anchor=100, exponent=0.5, scale=2.0
	// Value=105, overage=5, ValueAdj = 100 + round(sqrt(5)*2) = 100 + round(4.47) = 104
	si := StatInfo{Base: 0, Training: 105}
	si.Recalculate(1)

	assert.Equal(t, 105, si.Value)
	cfg := configs.GetProgressionConfig()
	overage := 105 - int(cfg.StatCapAnchor)
	expectedAdj := int(cfg.StatCapAnchor) + int(math.Round(math.Pow(float64(overage), float64(cfg.StatCapExponent))*float64(cfg.StatCapScale)))
	assert.Equal(t, expectedAdj, si.ValueAdj)
}
