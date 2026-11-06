// Package skills holds the data-driven skill definitions.
//
// Skills are identified by lowercase string ids (e.g. "cast", "dual-wield").
// The YAML datafiles in _datafiles/world/default/skills are the source of truth
// for which skills exist at runtime. Code that gates behavior on a specific
// skill references the id as a plain string literal.
package skills

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

// DefaultMaxLevel is the fallback skill cap when none is specified.
const DefaultMaxLevel = 4

var skillIdRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Skill is a data-driven skill definition loaded from a YAML datafile.
type Skill struct {
	SkillId     string `yaml:"skillid"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	MaxLevel    int    `yaml:"maxlevel"` // Validate() defaults to DefaultMaxLevel
}

var allSkills = map[string]*Skill{}

func (s *Skill) Id() string {
	return s.SkillId
}

func (s *Skill) Validate() error {
	s.SkillId = strings.ToLower(strings.TrimSpace(s.SkillId))
	if !skillIdRegex.MatchString(s.SkillId) {
		return fmt.Errorf("invalid skill id %q: must match %s", s.SkillId, skillIdRegex.String())
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("skill %q has no name", s.SkillId)
	}
	if strings.TrimSpace(s.Description) == "" {
		return fmt.Errorf("skill %q has no description", s.SkillId)
	}
	if s.MaxLevel < 1 {
		s.MaxLevel = DefaultMaxLevel
	}
	return nil
}

// Filename uses the raw skill id (the id regex is already filename-safe).
// Deliberately NOT util.ConvertForFilename — it would mangle "dual-wield" to "dual_wield".
func (s *Skill) Filename() string {
	return s.SkillId + `.yaml`
}

func (s *Skill) Filepath() string {
	return s.Filename()
}

// LoadDataFiles loads all skill definitions from disk into the in-memory cache.
func LoadDataFiles() {

	start := time.Now()

	tmpSkills, err := fileloader.LoadAllFlatFiles[string, *Skill](configs.GetFilePathsConfig().DataFiles.String() + `/skills`)
	if err != nil {
		panic(err)
	}

	allSkills = tmpSkills

	mudlog.Info("skills.LoadDataFiles()", "loadedCount", len(allSkills), "Time Taken", time.Since(start))
}

// GetSkill returns the skill definition for the given id, or nil if unknown.
func GetSkill(skillId string) *Skill {
	return allSkills[strings.ToLower(strings.TrimSpace(skillId))]
}

// GetAllSkills returns all loaded skills sorted by SkillId.
func GetAllSkills() []Skill {
	ret := make([]Skill, 0, len(allSkills))
	for _, s := range allSkills {
		ret = append(ret, *s)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].SkillId < ret[j].SkillId
	})
	return ret
}

// GetAllSkillNames returns all loaded skill ids sorted alphabetically.
func GetAllSkillNames() []string {
	ret := make([]string, 0, len(allSkills))
	for id := range allSkills {
		ret = append(ret, id)
	}
	sort.Strings(ret)
	return ret
}

// SkillExists reports whether a skill with the given id is loaded.
func SkillExists(skillId string) bool {
	_, ok := allSkills[strings.ToLower(strings.TrimSpace(skillId))]
	return ok
}

// MaxSkillLevel returns the configured cap for a skill, or DefaultMaxLevel for unknown skills.
func MaxSkillLevel(skillId string) int {
	if s := GetSkill(skillId); s != nil {
		return s.MaxLevel
	}
	return DefaultMaxLevel
}
