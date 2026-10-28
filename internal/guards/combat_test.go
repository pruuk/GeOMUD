package guards

import (
	"go/ast"
	"strings"
	"testing"
)

// TestOneCombatPath enforces the phase 4b slice 5 ruling: combat.Attack is the
// one attack and DoCombat the one round. The per-kind attack variants, the two
// per-kind round handlers and the room-message fields are gone and stay gone.
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
				t.Errorf("%s: uses %s; combat has one Attack and one round (slice 5)", path, id.Name)
			}
			return true
		})
	}

	// Positive control: the one Attack and the one round exist, so the walk
	// reached the files the ban is about.
	foundAttack, foundRound := false, false
	for path, f := range goFiles(t) {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if path == "internal/combat/combat.go" && fd.Name.Name == "Attack" && fd.Recv == nil {
				foundAttack = true
			}
			if path == "internal/hooks/NewRound_DoCombat.go" && fd.Name.Name == "DoCombat" {
				foundRound = true
			}
		}
	}
	if !foundAttack {
		t.Error("control: func Attack not found in internal/combat/combat.go; the guard may be scanning the wrong place")
	}
	if !foundRound {
		t.Error("control: func DoCombat not found in internal/hooks/NewRound_DoCombat.go; the guard may be scanning the wrong place")
	}
}

// TestCombatPackageDoesNotSend enforces: internal/combat returns messages and
// the round sends them. Nothing in the package sends text or builds an
// events.Message.
func TestCombatPackageDoesNotSend(t *testing.T) {
	banned := map[string]bool{
		"SendText":              true,
		"Tell":                  true,
		"SendTextToExits":       true,
		"SendTextCommunication": true,
	}
	scanned := 0
	for path, f := range goFiles(t) {
		if !strings.HasPrefix(path, "internal/combat/") {
			continue
		}
		scanned++
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				switch fn := x.Fun.(type) {
				case *ast.Ident:
					if banned[fn.Name] {
						t.Errorf("%s: calls %s; combat returns messages, the round sends them", path, fn.Name)
					}
				case *ast.SelectorExpr:
					if banned[fn.Sel.Name] {
						t.Errorf("%s: calls %s; combat returns messages, the round sends them", path, fn.Sel.Name)
					}
					if pkg, ok := fn.X.(*ast.Ident); ok && pkg.Name == "messaging" && fn.Sel.Name == "Send" {
						t.Errorf("%s: calls messaging.Send; combat returns messages, the round sends them", path)
					}
				}
			case *ast.CompositeLit:
				if sel, ok := x.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == "Message" {
					if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "events" {
						t.Errorf("%s: constructs events.Message; combat returns messages, the round sends them", path)
					}
				}
			}
			return true
		})
	}
	// Positive control: the combat package was parsed.
	if scanned < 5 {
		t.Errorf("control: scanned only %d internal/combat files; the guard may be scanning the wrong place", scanned)
	}
}
