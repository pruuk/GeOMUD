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
