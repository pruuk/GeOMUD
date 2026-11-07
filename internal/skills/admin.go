package skills

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// GetSkillsMap returns a copy of all loaded skills keyed by SkillId.
func GetSkillsMap() map[string]*Skill {
	result := make(map[string]*Skill, len(allSkills))
	for k, v := range allSkills {
		cp := *v
		result[k] = &cp
	}
	return result
}

func skillSaveModes() []fileloader.SaveOption {
	saveModes := []fileloader.SaveOption{}
	if configs.GetFilePathsConfig().CarefulSaveFiles {
		saveModes = append(saveModes, fileloader.SaveCareful)
	}
	return saveModes
}

// CreateSkill validates and persists a new skill, rejecting duplicate ids.
func CreateSkill(s *Skill) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if _, ok := allSkills[s.SkillId]; ok {
		return fmt.Errorf("skill %q already exists", s.SkillId)
	}
	return saveSkill(s)
}

// SaveSkill validates, persists a skill to disk, and updates the in-memory cache.
func SaveSkill(s *Skill) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return saveSkill(s)
}

func saveSkill(s *Skill) error {
	if err := fileloader.SaveFlatFile[*Skill](configs.GetFilePathsConfig().DataFiles.String()+`/skills`, s, skillSaveModes()...); err != nil {
		return err
	}
	allSkills[s.SkillId] = s
	return nil
}

// DeleteSkill removes a skill from disk and the in-memory cache.
func DeleteSkill(skillId string) error {
	skillId = strings.ToLower(strings.TrimSpace(skillId))
	s, ok := allSkills[skillId]
	if !ok {
		return fmt.Errorf("skill %q not found", skillId)
	}

	path := util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/`, `skills`, `/`, s.Filename())
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing skill file: %w", err)
	}
	delete(allSkills, skillId)
	return nil
}
