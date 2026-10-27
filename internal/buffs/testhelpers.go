package buffs

// SetTestSpec adds a buff spec in memory, for other packages' tests, and
// returns a func that takes it out again.
func SetTestSpec(spec *BuffSpec) (remove func()) {
	old, had := buffs[spec.BuffId]
	buffs[spec.BuffId] = spec
	return func() {
		if had {
			buffs[spec.BuffId] = old
			return
		}
		delete(buffs, spec.BuffId)
	}
}
