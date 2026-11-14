package characters

// IsUp reports whether the ruleset counts c as up (rules.Condition Up).
// internal/rules sets it (rules.Set, rules.SetForTest); characters cannot
// import rules, which imports actors, which imports characters. It is the
// one route by which characters asks the ruleset about a body (Phase 5
// slice 4.1); it compares no health with a number itself.
var IsUp func(c *Character) bool

// isUp is IsUp, failing loudly when nothing set it, as rules.Get does: a
// boot or a test that chose no ruleset must not guess a body's condition.
func isUp(c *Character) bool {
	if IsUp == nil {
		panic("characters: IsUp is not set; rules.Set wires it at boot")
	}
	return IsUp(c)
}
