package guards

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

const (
	eventsPath      = "github.com/GoMudEngine/GoMud/internal/events"
	connectionsPath = "github.com/GoMudEngine/GoMud/internal/connections"
	netPath         = "net"
)

// TestOutputGoesThroughMessaging enforces the Phase 4b output rule (spec
// section 2, "Guard test"; slice 7 plan decisions 2, 3 and 7): every byte of
// game text reaches a connection through internal/messaging, and GMCP is the
// only other writer. internal/messaging, internal/connections (the
// transport) and internal/perception (the vocabulary) are outside the guard.
// Package selectors are resolved through each file's imports, by import
// path, so an aliased import is seen and a local name that happens to read
// "events" is not. Everywhere else it fails on:
//
//   - a call to anything named SendText. Account.SendText is the transport
//     primitive (decision 3); only messaging and ConnectionController.Perceive
//     call it.
//   - any use of the type events.Message, events.Broadcast or
//     events.WebClientCommand (a composite literal, a var or field type,
//     new(...), a conversion, a type assertion): these events are the queued
//     transport (decision 2) and only messaging handles them. A composite
//     literal passed straight to events.RegisterListener or
//     events.UnregisterListener is a type key, not a message, and is allowed.
//   - connections.SendTo, connections.SendRawTo or connections.Broadcast.
//   - X.Write(...) on a net.Conn. The rule is syntactic, file by file: X is
//     an identifier declared in that file with type net.Conn (a parameter,
//     result, var or struct field), or assigned from a .Accept() call (how
//     main.go gets its connections), or a selector whose field name is
//     declared net.Conn in that file. Limit: a net.Conn reached any other way
//     (a getter, an embedded field, another package's struct, a name
//     re-declared with another type in the same file) is not seen. main.go's
//     raw writes go through messaging.RawConn.
//
// ConnectionDetails' write and writeRaw are unexported, and
// plugins.NetConnection has no Write, so the compiler stops a direct write to
// a connection from outside internal/connections; this test does not need a
// rule for it.
//
// An allowlist key is "rule|path", or "rule|path|Recv.Func" to allow only
// inside that function. Every entry must still match something, so a stale
// entry fails.
func TestOutputGoesThroughMessaging(t *testing.T) {
	inside := func(path string) bool {
		return strings.HasPrefix(path, "internal/messaging/") ||
			strings.HasPrefix(path, "internal/connections/") ||
			strings.HasPrefix(path, "internal/perception/")
	}
	// key -> reason
	allowed := map[string]string{
		"SendText|internal/actors/controller.go":                 "ConnectionController.Perceive delivers a perception to the account (decision 3)",
		"event|internal/accounts/userrecord.go|Account.SendText": "Account.SendText, the transport primitive, queues the events.Message (decision 3)",
		"connections|modules/gmcp/gmcp.go":                       "GMCP keeps its own path for good (spec section 2)",
		"netconn|cmd/smoke/client.go":                            "the smoke test's telnet client writes to the server, as a player types",
	}
	matched := map[string]int{}

	files := goFiles(t)
	for path, f := range files {
		if inside(path) {
			continue
		}
		scanOutput(f, func(rule, scope, msg string) {
			for _, key := range []string{rule + "|" + path + "|" + scope, rule + "|" + path} {
				if _, ok := allowed[key]; ok {
					matched[key]++
					return
				}
			}
			t.Errorf("%s (%s): %s; output goes through internal/messaging", path, scope, msg)
		})
	}
	for key, reason := range allowed {
		if matched[key] == 0 {
			t.Errorf("allowlist entry %q (%s) matched nothing; remove it or fix the key", key, reason)
		}
	}

	// Positive controls: the guard is scanning the right files, and each
	// detector can fire.
	if w := files["internal/messaging/wire.go"]; w == nil || !hasFunc(w, "WriteMessage") {
		t.Error("control: messaging.WriteMessage not found in internal/messaging/wire.go")
	}
	if matched["SendText|internal/actors/controller.go"] == 0 {
		t.Error("control: internal/actors/controller.go (ConnectionController.Perceive) no longer calls SendText")
	}
	if matched["connections|modules/gmcp/gmcp.go"] == 0 {
		t.Error("control: modules/gmcp/gmcp.go no longer calls connections.SendTo")
	}
	if i := files["internal/messaging/interface.go"]; i == nil || !writesNetConn(i) {
		t.Error("control: the net.Conn rule does not see messaging.RawConn's c.Write; the rule is broken")
	}
}

// TestOutputGuardResolvesImports is the positive control for import
// resolution and the type rule: an aliased events import is seen in a var
// type, new(...), a conversion and a literal; a listener type key is not
// flagged; an import named "events" that is not internal/events is ignored.
func TestOutputGuardResolvesImports(t *testing.T) {
	src := `package p

import (
	ev "github.com/GoMudEngine/GoMud/internal/events"
	conn "github.com/GoMudEngine/GoMud/internal/connections"
	events "strings"
)

var _ ev.Message
var _ events.Message

type T struct{ M ev.WebClientCommand }

func f(x any) {
	_ = new(ev.Broadcast)
	_ = x.(ev.Message)
	_ = ev.Broadcast(ev.Broadcast{})
	ev.RegisterListener(ev.Message{}, nil)
	conn.SendTo(nil, 1)
}
`
	f, err := parser.ParseFile(token.NewFileSet(), "control.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	scanOutput(f, func(rule, scope, msg string) { got[rule+"|"+scope]++ })
	want := map[string]int{
		"event|":        2, // var _ ev.Message, field M ev.WebClientCommand
		"event|f":       4, // new, assertion, conversion, the converted literal
		"connections|f": 1,
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("control: %s fired %d times, want %d (got %v)", k, got[k], n, got)
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("control: unexpected %s (%d)", k, got[k])
		}
	}
}

// scanOutput reports every output-rule hit in f. scope is the enclosing
// function ("Func" or "Recv.Func"), or "" at package level.
func scanOutput(f *ast.File, report func(rule, scope, msg string)) {
	imp := importNames(f)
	transportEvents := map[string]bool{"Message": true, "Broadcast": true, "WebClientCommand": true}
	connWriters := map[string]bool{"SendTo": true, "SendRawTo": true, "Broadcast": true}
	keys := listenerKeys(f, imp)
	conns := netConnNames(f, imp)
	for _, d := range f.Decls {
		scope := declScope(d)
		ast.Inspect(d, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if name := pkgSel(x, imp, eventsPath); transportEvents[name] && !keys[x] {
					report("event", scope, "uses the type events."+name)
				}
			case *ast.CallExpr:
				switch fn := x.Fun.(type) {
				case *ast.Ident:
					if fn.Name == "SendText" {
						report("SendText", scope, "calls SendText")
					}
				case *ast.SelectorExpr:
					if fn.Sel.Name == "SendText" {
						report("SendText", scope, "calls SendText")
					}
					if name := pkgSel(fn, imp, connectionsPath); connWriters[name] {
						report("connections", scope, "writes to a connection (connections."+name+")")
					}
					if fn.Sel.Name == "Write" && isNetConn(fn.X, conns) {
						report("netconn", scope, "writes to a net.Conn (.Write)")
					}
				}
			}
			return true
		})
	}
}

// declScope names a declaration's function: "Func", "Recv.Func", or "".
func declScope(d ast.Decl) string {
	fd, ok := d.(*ast.FuncDecl)
	if !ok {
		return ""
	}
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	recv := fd.Recv.List[0].Type
	if star, ok := recv.(*ast.StarExpr); ok {
		recv = star.X
	}
	if idx, ok := recv.(*ast.IndexExpr); ok {
		recv = idx.X
	}
	if id, ok := recv.(*ast.Ident); ok {
		return id.Name + "." + fd.Name.Name
	}
	return fd.Name.Name
}

// importNames maps each name a file uses for an import to its path: the
// alias when there is one, else the path's last element. Blank and dot
// imports are left out.
func importNames(f *ast.File) map[string]string {
	out := map[string]string{}
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := path[strings.LastIndex(path, "/")+1:]
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "_" || name == "." {
			continue
		}
		out[name] = path
	}
	return out
}

// pkgSel returns Sel's name when e is the selector X.Sel and X names the
// import with path pkgPath (an unresolved identifier, so not a local), else "".
func pkgSel(e ast.Expr, imp map[string]string, pkgPath string) string {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if x, ok := sel.X.(*ast.Ident); ok && x.Obj == nil && imp[x.Name] == pkgPath {
		return sel.Sel.Name
	}
	return ""
}

// listenerKeys are the type selectors of composite literals passed straight
// to events.RegisterListener or events.UnregisterListener.
func listenerKeys(f *ast.File, imp map[string]string) map[*ast.SelectorExpr]bool {
	out := map[*ast.SelectorExpr]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if name := pkgSel(call.Fun, imp, eventsPath); name == "RegisterListener" || name == "UnregisterListener" {
			for _, arg := range call.Args {
				if lit, ok := arg.(*ast.CompositeLit); ok {
					if sel, ok := lit.Type.(*ast.SelectorExpr); ok {
						out[sel] = true
					}
				}
			}
		}
		return true
	})
	return out
}

// netConnNames are the names a file declares as net.Conn (fields, params,
// results, vars) or assigns from a .Accept() call.
func netConnNames(f *ast.File, imp map[string]string) map[string]bool {
	out := map[string]bool{}
	isConnType := func(e ast.Expr) bool { return pkgSel(e, imp, netPath) == "Conn" }
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Field:
			if isConnType(x.Type) {
				for _, name := range x.Names {
					out[name.Name] = true
				}
			}
		case *ast.ValueSpec:
			if isConnType(x.Type) {
				for _, name := range x.Names {
					out[name.Name] = true
				}
			}
		case *ast.AssignStmt:
			if (x.Tok == token.DEFINE || x.Tok == token.ASSIGN) && len(x.Rhs) == 1 && len(x.Lhs) >= 1 {
				if call, ok := x.Rhs[0].(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Accept" {
						if id, ok := x.Lhs[0].(*ast.Ident); ok {
							out[id.Name] = true
						}
					}
				}
			}
		}
		return true
	})
	return out
}

// isNetConn reports whether e names a net.Conn by netConnNames.
func isNetConn(e ast.Expr, conns map[string]bool) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return conns[x.Name]
	case *ast.SelectorExpr:
		return conns[x.Sel.Name]
	}
	return false
}

// writesNetConn reports whether f has an X.Write call on a net.Conn.
func writesNetConn(f *ast.File) bool {
	conns := netConnNames(f, importNames(f))
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Write" && isNetConn(sel.X, conns) {
				found = true
			}
		}
		return true
	})
	return found
}

// hasFunc reports whether f declares a function or method named name.
func hasFunc(f *ast.File, name string) bool {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == name {
			return true
		}
	}
	return false
}
