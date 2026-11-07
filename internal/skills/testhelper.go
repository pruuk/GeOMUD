package skills

// SetTestData replaces the in-memory skill cache. It is intended
// for use by tests in this and other packages that need a deterministic, on-disk-free
// skill set seeded before exercising skill-dependent behavior.
func SetTestData(skillList []*Skill) {
	allSkills = map[string]*Skill{}
	for _, s := range skillList {
		cp := *s
		allSkills[cp.SkillId] = &cp
	}
}
