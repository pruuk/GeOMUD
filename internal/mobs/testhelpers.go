package mobs

// NewTestInstance adds m as a new instantiation, for other packages' tests.
func NewTestInstance(m Mob) *Mob { return addInstance(m) }
