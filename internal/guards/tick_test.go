package guards

import (
	"go/ast"
	"testing"
)

// TestOneActorTick enforces the phase 4b slice 6 ruling: every actor gets the
// same round upkeep (ActorUpkeep) and one tick through Controller.Tick
// (ActorTick). The per-kind round ticks, the idle hooks, the MobIdle event and
// the dead boredom unload are gone and stay gone.
func TestOneActorTick(t *testing.T) {
	banned := map[string]bool{
		"UserRoundTick":      true,
		"MobRoundTick":       true,
		"IdleMobs":           true,
		"HandleIdleMobs":     true,
		"MobIdle":            true,
		"BoredomCounter":     true,
		"MaxMobBoredom":      true,
		"MobUnloadThreshold": true,
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
				t.Errorf("%s: uses %s; one actor upkeep and one actor tick (slice 6)", path, id.Name)
			}
			return true
		})
	}

	// Positive control: the two passes and the controller tick exist, so the
	// walk reached the files the ban is about.
	want := map[string]bool{
		"internal/hooks/NewRound_ActorUpkeep.go ActorUpkeep": false,
		"internal/hooks/NewRound_ActorTick.go ActorTick":     false,
		"internal/actors/controller.go Tick":                 false,
	}
	for path, f := range goFiles(t) {
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok {
				key := path + " " + fd.Name.Name
				if _, ok := want[key]; ok {
					want[key] = true
				}
			}
		}
	}
	for key, found := range want {
		if !found {
			t.Errorf("control: %s not found; the guard may be scanning the wrong place", key)
		}
	}
}
