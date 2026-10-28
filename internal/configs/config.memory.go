package configs

type Memory struct {
	// Room memory unload thresholds
	RoomUnloadRounds    ConfigInt `yaml:"RoomUnloadRounds"`
	RoomUnloadThreshold ConfigInt `yaml:"RoomUnloadThreshold"`
}

func (m *Memory) Validate() {

	if m.RoomUnloadRounds < 5 {
		m.RoomUnloadRounds = 5
	}

	if m.RoomUnloadThreshold < 0 {
		m.RoomUnloadThreshold = 0
	}

}

func GetMemoryConfig() Memory {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.Memory
}
