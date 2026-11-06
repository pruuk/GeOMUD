package configs

import "math"

type Timing struct {
	TurnMs            ConfigInt `yaml:"TurnMs"`
	RoundSeconds      ConfigInt `yaml:"RoundSeconds"`
	RoundsPerAutoSave ConfigInt `yaml:"RoundsPerAutoSave"`

	// Protected values
	turnsPerRound   int     // calculated and cached when data is validated.
	turnsPerSave    int     // calculated and cached when data is validated.
	roundsPerMinute float64 // calculated and cached when data is validated.
}

func (e *Timing) Validate() {

	if e.TurnMs < 10 {
		e.TurnMs = 100 // default
	}

	if e.RoundSeconds < 1 {
		e.RoundSeconds = 4 // default
	}

	if e.RoundsPerAutoSave < 1 {
		e.RoundsPerAutoSave = 900 // default of 15 minutes worth of rounds
	}

	// Pre-calculate and cache useful values
	e.turnsPerRound = int((e.RoundSeconds * 1000) / e.TurnMs)
	e.turnsPerSave = int(e.RoundsPerAutoSave) * e.turnsPerRound
	e.roundsPerMinute = 60 / float64(e.RoundSeconds)

}

func (e Timing) TurnsPerRound() int {
	return e.turnsPerRound
}

func (e Timing) TurnsPerAutoSave() int {
	return e.turnsPerSave
}

func (e Timing) MinutesToRounds(minutes int) int {
	return int(math.Ceil(e.roundsPerMinute * float64(minutes)))
}

func (e Timing) SecondsToRounds(seconds int) int {
	return int(math.Ceil(float64(seconds) / float64(e.RoundSeconds)))
}

func (e Timing) MinutesToTurns(minutes int) int {
	return int(math.Ceil(float64(minutes*60*1000) / float64(e.TurnMs)))
}

func (e Timing) SecondsToTurns(seconds int) int {
	return int(math.Ceil(float64(seconds*1000) / float64(e.TurnMs)))
}

func (e Timing) RoundsToSeconds(rounds int) int {
	return int(math.Ceil(float64(rounds) * float64(e.RoundSeconds)))
}

func GetTimingConfig() Timing {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Timing
}
