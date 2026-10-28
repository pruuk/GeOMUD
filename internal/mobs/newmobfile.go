package mobs

import (
	"errors"
)

func CreateNewMobFile(newMobInfo Mob) (MobId, error) {

	newMobInfo.MobId = getNextMobId()

	if newMobInfo.MobId == 0 {
		return 0, errors.New(`Could not find a new mob id to assign.`)
	}

	if err := newMobInfo.Validate(); err != nil {
		return 0, err
	}

	if newMobInfo.MobId == 0 {
		return 0, errors.New(`Could not find a new mob id to assign.`)
	}

	if err := writeTemplateFile(&newMobInfo); err != nil {
		return 0, err
	}

	// Save to in-memory cache
	allMobNames = append(allMobNames, newMobInfo.Character.Name)
	mobNameCache[newMobInfo.MobId] = newMobInfo.Character.Name
	mobs[newMobInfo.Id()] = &newMobInfo

	return newMobInfo.MobId, nil
}

func getNextMobId() MobId {

	lowestFreeId := MobId(0)
	for _, mInfo := range mobs {

		if mInfo.MobId >= lowestFreeId {
			lowestFreeId = mInfo.MobId + 1
		}
	}

	return lowestFreeId
}
