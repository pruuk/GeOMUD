package mutators

// SetTestSpec adds a mutator spec in memory, for other packages' tests, and
// returns a func that takes it out again. Nothing is saved to disk.
func SetTestSpec(spec *MutatorSpec) (remove func()) {
	old, had := allMutators[spec.MutatorId]
	allMutators[spec.MutatorId] = spec
	return func() {
		if had {
			allMutators[spec.MutatorId] = old
			return
		}
		delete(allMutators, spec.MutatorId)
	}
}
