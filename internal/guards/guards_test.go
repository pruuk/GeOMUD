package guards

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is two levels up from this package.
const repoRoot = "../.."

// goFiles parses every non-test Go file in the repo outside .git and returns
// them keyed by slash path relative to the repo root.
func goFiles(t *testing.T) map[string]*ast.File {
	t.Helper()
	fset := token.NewFileSet()
	out := map[string]*ast.File{}
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "_datafiles") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repoRoot, path)
		out[filepath.ToSlash(rel)] = f
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 100 {
		t.Fatalf("scanned only %d files; the walk is not reaching the repo", len(out))
	}
	return out
}

// TestCharacterPointerIsNotReassigned enforces: once an account or mob has a
// character, that *Character is shared with its actor and is never replaced.
// Edits copy into it (*x.Character = value). Allowed: building a detached copy
// before decoding into it.
func TestCharacterPointerIsNotReassigned(t *testing.T) {
	allowed := map[string]string{
		"internal/web/api_v1_users.go": "decodes the PATCH body into a copy; UpdateOnlineUser copies it into the live pointer",
		"internal/accounts/users.go":   "UpdateOnlineUser points the edited copy at the live pointer",
		"internal/actors/registry.go":  "bindAccount points the actor at its account's live character when a new record of the account comes online; no account's or mob's own pointer changes",
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range as.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == "Character" {
					if _, ok := allowed[path]; !ok {
						t.Errorf("%s: assigns .Character; copy into the pointer instead (*x.Character = v)", path)
					}
				}
			}
			return true
		})
	}
}

// TestNoControllerTypeAssertions enforces "ask the actor, never its type":
// outside internal/actors nothing asserts or switches on a concrete
// controller type.
func TestNoControllerTypeAssertions(t *testing.T) {
	concrete := map[string]bool{"ConnectionController": true, "BehaviourController": true}
	names := func(e ast.Expr) string {
		if star, ok := e.(*ast.StarExpr); ok {
			e = star.X
		}
		if sel, ok := e.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "actors" && concrete[sel.Sel.Name] {
				return sel.Sel.Name
			}
		}
		return ""
	}
	for path, f := range goFiles(t) {
		if strings.HasPrefix(path, "internal/actors/") {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeAssertExpr:
				if name := names(x.Type); name != "" {
					t.Errorf("%s: asserts actors.%s; ask the controller instead", path, name)
				}
			case *ast.CaseClause:
				for _, e := range x.List {
					if name := names(e); name != "" {
						t.Errorf("%s: type switch on actors.%s; ask the controller instead", path, name)
					}
				}
			}
			return true
		})
	}
}

// TestAccountsGoOnlineOneWay enforces that an account is recorded as online
// only by activateLocked, so OnActivate (and so its actor) can never be
// skipped. testhelpers.go is the test-only exception.
func TestAccountsGoOnlineOneWay(t *testing.T) {
	allowed := map[string]bool{
		"internal/accounts/lifecycle.go":   true,
		"internal/accounts/testhelpers.go": true,
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range as.Lhs {
				idx, ok := lhs.(*ast.IndexExpr)
				if !ok {
					continue
				}
				sel, ok := idx.X.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Users" {
					continue
				}
				if x, ok := sel.X.(*ast.Ident); ok && x.Name == "userManager" && !allowed[path] {
					t.Errorf("%s: writes userManager.Users directly; use activateLocked", path)
				}
			}
			return true
		})
	}
}

// TestNoMobInstanceIds: one ID space means the old mob instance ID is gone,
// in names and fields alike.
func TestNoMobInstanceIds(t *testing.T) {
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && strings.Contains(id.Name, "MobInstanceId") {
				t.Errorf("%s: %s; a body is named by its ids.ID", path, id.Name)
			}
			return true
		})
	}
}

// TestIdsAreMadeOnlyAtCreation enforces decision 11: an ID is made once, when
// a thing is created or instantiated. Each allowed file names its creation site.
func TestIdsAreMadeOnlyAtCreation(t *testing.T) {
	allowed := map[string]string{
		"internal/accounts/migration.go":   "giveBodyId: a new character, or one saved before body IDs existed",
		"internal/mobs/mobs.go":            "addInstance: a new NPC instantiation",
		"internal/usercommands/suicide.go": "permadeath reroll: a new character",
		"cmd/newid/main.go":                "a builder creating a thing by hand",
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "New" {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "ids" {
				if _, ok := allowed[path]; !ok {
					t.Errorf("%s: makes an ids.ID outside a creation site", path)
				}
			}
			return true
		})
	}
}

// TestNoNumericShorthandIds forbids actor shorthand IDs (#N and @N): people
// name things by name; clients send ID tokens. The allowlist holds the
// documented exceptions, keyed by file and the literal's Go source text, so a
// new #%d elsewhere in the same file still fails.
func TestNoNumericShorthandIds(t *testing.T) {
	allowed := map[string]string{
		"internal/rooms/rooms.go|`%s#%d`":           "the positional name form name#2, kept by ruling (decision 10 note: name forms unchanged)",
		"internal/usercommands/keyring.go|`#%d %s`": "a room ID label; rooms keep integer IDs until the geo room key",
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING &&
				(strings.Contains(lit.Value, "#%d") || strings.Contains(lit.Value, "@%d")) {
				if allowed[path+"|"+lit.Value] != "" {
					return true
				}
				t.Errorf("%s: formats a #N or @N id (%s)", path, lit.Value)
			}
			return true
		})
	}
}
