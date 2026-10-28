package guards

import (
	"go/ast"
	"strconv"
	"strings"
	"testing"
)

// TestNoSecondCommandSet enforces Phase 4b slice 4: one command registry
// (internal/usercommands) serves players and NPCs, and one input path reaches it.
// internal/mobcommands is deleted and processMobInput with it; nothing may bring
// either back.
func TestNoSecondCommandSet(t *testing.T) {
	for path, f := range goFiles(t) {
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasSuffix(p, "/internal/mobcommands") || strings.Contains(p, "/internal/mobcommands/") {
				t.Errorf("%s imports %s: there is one command set, internal/usercommands", path, p)
			}
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "processMobInput" {
				t.Errorf("%s defines processMobInput: every actor's input takes the one path", path)
			}
		}
	}
}

// TestNoAddMobCommand enforces slice 4 decision 16: the plugin API has one
// command registration, AddUserCommand, which takes a usercommands.Command.
func TestNoAddMobCommand(t *testing.T) {
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if x.Sel.Name == "AddMobCommand" {
					t.Errorf("%s: AddMobCommand is gone, use AddUserCommand", path)
				}
			case *ast.FuncDecl:
				if x.Name.Name == "AddMobCommand" {
					t.Errorf("%s defines AddMobCommand", path)
				}
			}
			return true
		})
	}
}

// exprString renders the simple type expressions the command shape uses
// (pkg.Name, *pkg.Name, string).
func exprString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return "*" + exprString(x.X)
	case *ast.SelectorExpr:
		return exprString(x.X) + "." + x.Sel.Name
	}
	return "?"
}

// TestCommandsTakeTheActor enforces slice 4 decision 11: every command takes
// the actor. A command is an exported function in internal/usercommands or one
// registered in userCommands. A command with the Command shape (a string, then
// parameters, then *rooms.Room and events.EventFlag) that has a parameter
// `user *accounts.Account` is written for accounts only. Unexported helpers
// that share the shape (the admin sub-handlers a command calls once it has
// found the account, such as mob_Create) are not commands.
func TestCommandsTakeTheActor(t *testing.T) {
	files := goFiles(t)

	// Names registered in the userCommands map literal.
	registered := map[string]bool{}
	for path, f := range files {
		if !strings.HasPrefix(path, "internal/usercommands/") {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			vs, ok := n.(*ast.ValueSpec)
			if !ok || len(vs.Names) != 1 || vs.Names[0].Name != "userCommands" {
				return true
			}
			ast.Inspect(vs, func(m ast.Node) bool {
				if kv, ok := m.(*ast.KeyValueExpr); ok {
					if lit, ok := kv.Value.(*ast.CompositeLit); ok && len(lit.Elts) > 0 {
						first := lit.Elts[0]
						if inner, ok := first.(*ast.KeyValueExpr); ok {
							first = inner.Value
						}
						if id, ok := first.(*ast.Ident); ok {
							registered[id.Name] = true
						}
					}
				}
				return true
			})
			return false
		})
	}
	if len(registered) < 50 {
		t.Fatalf("found only %d registered commands; the registry scan is not reaching userCommands", len(registered))
	}

	seen := 0
	for path, f := range files {
		if !strings.HasPrefix(path, "internal/usercommands/") {
			continue
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if !fn.Name.IsExported() && !registered[fn.Name.Name] {
				continue
			}
			var types, names []string
			for _, field := range fn.Type.Params.List {
				n := len(field.Names)
				if n == 0 {
					n = 1
				}
				for i := 0; i < n; i++ {
					types = append(types, exprString(field.Type))
					if len(field.Names) > 0 {
						names = append(names, field.Names[i].Name)
					} else {
						names = append(names, "")
					}
				}
			}
			if len(types) < 3 || types[0] != "string" ||
				types[len(types)-2] != "*rooms.Room" || types[len(types)-1] != "events.EventFlag" {
				continue
			}
			seen++
			for i, ty := range types {
				if ty == "*accounts.Account" && names[i] == "user" {
					t.Errorf("%s: %s is a command with the Command shape that takes `user *accounts.Account`: a command takes the actor", path, fn.Name.Name)
				}
			}
		}
	}
	if seen < 50 {
		t.Fatalf("matched only %d functions with the Command shape; the shape match is not reaching the commands", seen)
	}
}
