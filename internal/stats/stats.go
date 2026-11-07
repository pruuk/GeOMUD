package stats

import (
	"math"

	"github.com/GoMudEngine/GoMud/internal/configs"
)

type Statistics struct {
	Strength   StatInfo `yaml:"strength,omitempty"`   // Muscular strength (damage?)
	Speed      StatInfo `yaml:"speed,omitempty"`      // Speed and agility (dodging)
	Smarts     StatInfo `yaml:"smarts,omitempty"`     // Intelligence and wisdom (magic power, memory, deduction, etc)
	Vitality   StatInfo `yaml:"vitality,omitempty"`   // Health and stamina (health capacity)
	Mysticism  StatInfo `yaml:"mysticism,omitempty"`  // Magic and mana (magic capacity)
	Perception StatInfo `yaml:"perception,omitempty"` // How well you notice things
}

// When saving to a file, we don't need to write all the properties that we calculate.
// Just keep track of "Training" because that's not calculated.
type StatInfo struct {
	Training int  `yaml:"training,omitempty"` // How much it's been trained with Training Points spending
	Value    int  `yaml:"-"`                  // Final calculated value
	ValueAdj int  `yaml:"-"`                  // Final calculated value (Adjusted)
	Racial   int  `yaml:"-"`                  // Value provided by racial benefits
	Base     int  `yaml:"base,omitempty"`     // Base stat value
	Mods     int  `yaml:"-"`                  // How much it's modded by equipment, spells, etc.
	NoCap    bool `yaml:"-"`                  // When true, skip the stat cap compression in Recalculate
}

// GainsForLevel returns the racial stat value at the given level, using the
// configured progression formula:
//
//	racial = floor(base * BaseModFactor * (level-1)^BaseModExponent)
//	       + floor(NaturalGainsModFactor * level^NaturalGainsExponent)
func (si *StatInfo) GainsForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	cfg := configs.GetProgressionConfig()

	basePoints := int(math.Pow(float64(level-1), float64(cfg.BaseModExponent)) *
		float64(cfg.BaseModFactor) * float64(si.Base))

	freePoints := int(math.Pow(float64(level), float64(cfg.NaturalGainsExponent)) *
		float64(cfg.NaturalGainsModFactor))

	return basePoints + freePoints
}

func (si *StatInfo) Recalculate(level int) {
	si.Racial = si.GainsForLevel(level)
	si.Value = si.Racial + si.Training + si.Mods
	si.ValueAdj = si.Value
	if si.NoCap {
		return
	}
	cfg := configs.GetProgressionConfig()
	if bool(cfg.StatCapExemptBonus) {
		// Compress only the racial portion; training and mods are added uncapped.
		compressedRacial := si.Racial
		if si.Racial >= int(cfg.StatCapThreshold) {
			overage := si.Racial - int(cfg.StatCapAnchor)
			if overage < 0 {
				overage = 0
			}
			compressedRacial = int(cfg.StatCapAnchor) + int(math.Round(math.Pow(float64(overage), float64(cfg.StatCapExponent))*float64(cfg.StatCapScale)))
		}
		si.ValueAdj = compressedRacial + si.Training + si.Mods
	} else if si.ValueAdj >= int(cfg.StatCapThreshold) {
		overage := si.ValueAdj - int(cfg.StatCapAnchor)
		if overage < 0 {
			overage = 0
		}
		si.ValueAdj = int(cfg.StatCapAnchor) + int(math.Round(math.Pow(float64(overage), float64(cfg.StatCapExponent))*float64(cfg.StatCapScale)))
	}
}
