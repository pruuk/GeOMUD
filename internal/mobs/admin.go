package mobs

import (
	"fmt"
	"os"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func GetAllMobSpecs() []Mob {
	result := make([]Mob, 0, len(mobs))
	for _, m := range mobs {
		result = append(result, *m)
	}
	return result
}

// writeTemplateFile writes spec to its mob template file. A template's items
// are not things, so the file holds a copy whose items have no IDs (Phase 4b
// slice 2c); spec itself is untouched. Every mob template file write goes
// through here.
func writeTemplateFile(spec *Mob) error {
	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}

	fileSpec := *spec
	fileSpec.Character.StripItemIds()
	return fileloader.SaveFlatFile[*Mob](configs.GetFilePathsConfig().DataFiles.String()+`/mobs`, &fileSpec, saveModes...)
}

func SaveMobSpec(spec *Mob) error {
	if spec.MobId < 1 {
		return fmt.Errorf("cannot save mob spec with invalid MobId %d", spec.MobId)
	}

	if err := spec.Validate(); err != nil {
		return err
	}

	if err := writeTemplateFile(spec); err != nil {
		return err
	}

	oldName, hadOldName := mobNameCache[spec.MobId]

	mobs[spec.Id()] = spec
	mobNameCache[spec.MobId] = spec.Character.Name

	if hadOldName && oldName != spec.Character.Name {
		for i, n := range allMobNames {
			if n == oldName {
				allMobNames[i] = spec.Character.Name
				break
			}
		}
	} else if !hadOldName {
		allMobNames = append(allMobNames, spec.Character.Name)
	}

	return nil
}

func DeleteMobSpec(mobId MobId) error {
	spec := GetMobSpec(mobId)
	if spec == nil {
		return fmt.Errorf("mob %d not found", mobId)
	}

	basePath := util.FilePath(configs.GetFilePathsConfig().DataFiles.String() + `/mobs/`)

	yamlPath := basePath + spec.Filepath()
	if err := os.Remove(yamlPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing mob yaml: %w", err)
	}

	name := mobNameCache[mobId]
	for i, n := range allMobNames {
		if n == name {
			allMobNames = append(allMobNames[:i], allMobNames[i+1:]...)
			break
		}
	}
	delete(mobNameCache, mobId)
	delete(mobs, int(mobId))

	return nil
}

func StockMobShop(mobId MobId, entry characters.ShopItem) error {
	spec, ok := mobs[int(mobId)]
	if !ok {
		return fmt.Errorf("mob %d not found", mobId)
	}

	spec.Character.Shop = append(spec.Character.Shop, entry)

	return SaveMobSpec(spec)
}
