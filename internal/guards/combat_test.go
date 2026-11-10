package guards

import (
	"go/ast"
	"strings"
	"testing"
)

// TestOneCombatPath enforces the phase 4b slice 5 ruling, as the Phase 5 seam
// keeps it: one attack (the engine's attack, calling the ruleset's
// ResolveAttack) and one round (DoCombat). The per-kind attack variants, the
// two per-kind round handlers and the room-message fields are gone and stay
// gone.
func TestOneCombatPath(t *testing.T) {
	banned := map[string]bool{
		"AttackPlayerVsMob":    true,
		"AttackPlayerVsPlayer": true,
		"AttackMobVsPlayer":    true,
		"AttackMobVsMob":       true,
		"handlePlayerCombat":   true,
		"handleMobCombat":      true,
		"MessagesToSourceRoom": true,
		"MessagesToTargetRoom": true,
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
				t.Errorf("%s: uses %s; combat has one attack and one round (slice 5, kept by the Phase 5 seam)", path, id.Name)
			}
			return true
		})
	}

	// Positive control: the one attack, the ruleset's resolution and the one
	// round exist, so the walk reached the files the ban is about.
	foundAttack, foundResolve, foundRound := false, false, false
	for path, f := range goFiles(t) {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if path == "internal/hooks/NewRound_DoCombat.go" && fd.Name.Name == "attack" && fd.Recv == nil {
				foundAttack = true
			}
			if path == "internal/rules/dikubasic/ruleset.go" && fd.Name.Name == "ResolveAttack" && fd.Recv != nil {
				foundResolve = true
			}
			if path == "internal/hooks/NewRound_DoCombat.go" && fd.Name.Name == "DoCombat" {
				foundRound = true
			}
		}
	}
	if !foundAttack {
		t.Error("control: func attack not found in internal/hooks/NewRound_DoCombat.go; the guard may be scanning the wrong place")
	}
	if !foundResolve {
		t.Error("control: method ResolveAttack not found in internal/rules/dikubasic/ruleset.go; the guard may be scanning the wrong place")
	}
	if !foundRound {
		t.Error("control: func DoCombat not found in internal/hooks/NewRound_DoCombat.go; the guard may be scanning the wrong place")
	}
}

// TestRulesetDoesNotSend enforces: internal/rules returns messages and the
// engine sends them. Nothing under it sends text or builds an events.Message.
func TestRulesetDoesNotSend(t *testing.T) {
	banned := map[string]bool{
		"SendText":              true,
		"Tell":                  true,
		"SendTextToExits":       true,
		"SendTextCommunication": true,
	}
	scanned := 0
	for path, f := range goFiles(t) {
		if !strings.HasPrefix(path, "internal/rules/") {
			continue
		}
		scanned++
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				switch fn := x.Fun.(type) {
				case *ast.Ident:
					if banned[fn.Name] {
						t.Errorf("%s: calls %s; a ruleset returns messages, the engine sends them", path, fn.Name)
					}
				case *ast.SelectorExpr:
					if banned[fn.Sel.Name] {
						t.Errorf("%s: calls %s; a ruleset returns messages, the engine sends them", path, fn.Sel.Name)
					}
					if pkg, ok := fn.X.(*ast.Ident); ok && pkg.Name == "messaging" && fn.Sel.Name == "Send" {
						t.Errorf("%s: calls messaging.Send; a ruleset returns messages, the engine sends them", path)
					}
				}
			case *ast.CompositeLit:
				if sel, ok := x.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "Message" {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "events" {
						t.Errorf("%s: constructs events.Message; a ruleset returns messages, the engine sends them", path)
					}
				}
			}
			return true
		})
	}
	// Positive control: internal/rules was parsed.
	if scanned < 5 {
		t.Errorf("control: scanned only %d internal/rules files; the guard may be scanning the wrong place", scanned)
	}
}
