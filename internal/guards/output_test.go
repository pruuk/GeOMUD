package guards

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"
)

// TestOutputGoesThroughMessaging enforces the Phase 4b output rule (spec
// section 2, "Guard test"; slice 7 plan decisions 2, 3 and 7): every byte of
// game text reaches a connection through internal/messaging, and GMCP is the
// only other writer. internal/messaging, internal/connections (the
// transport) and internal/perception (the vocabulary) are outside the guard.
// Everywhere else it fails on:
//
//   - a call to anything named SendText. Account.SendText is the transport
//     primitive (decision 3); only messaging and ConnectionController.Perceive
//     call it.
//   - a composite literal events.Message, events.Broadcast or
//     events.WebClientCommand: these events are the queued transport
//     (decision 2) and only messaging builds them. A literal passed straight
//     to events.RegisterListener or events.UnregisterListener is a type key,
//     not a message, and is allowed.
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
// Every allowlist entry must still match something, so a stale entry fails.
func TestOutputGoesThroughMessaging(t *testing.T) {
	inside := func(path string) bool {
		return strings.HasPrefix(path, "internal/messaging/") ||
			strings.HasPrefix(path, "internal/connections/") ||
			strings.HasPrefix(path, "internal/perception/")
	}
	// rule|path -> reason
	allowed := map[string]string{
		"SendText|internal/actors/controller.go": "ConnectionController.Perceive delivers a perception to the account (decision 3)",
		"event|internal/accounts/userrecord.go":  "Account.SendText, the transport primitive, queues the events.Message (decision 3)",
		"connections|modules/gmcp/gmcp.go":       "GMCP keeps its own path for good (spec section 2)",
		"netconn|cmd/smoke/client.go":            "the smoke test's telnet client writes to the server, as a player types",
	}
	matched := map[string]int{}
	report := func(rule, path, msg string) {
		key := rule + "|" + path
		if _, ok := allowed[key]; ok {
			matched[key]++
			return
		}
		t.Errorf("%s: %s; output goes through internal/messaging", path, msg)
	}
	transportEvents := map[string]bool{"Message": true, "Broadcast": true, "WebClientCommand": true}
	connWriters := map[string]bool{"SendTo": true, "SendRawTo": true, "Broadcast": true}

	files := goFiles(t)
	for path, f := range files {
		if inside(path) {
			continue
		}
		keys := listenerKeys(f)
		conns := netConnNames(f)
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CompositeLit:
				if name := pkgSel(x.Type, "events"); transportEvents[name] && !keys[x] {
					report("event", path, "builds an events."+name)
				}
			case *ast.CallExpr:
				switch fn := x.Fun.(type) {
				case *ast.Ident:
					if fn.Name == "SendText" {
						report("SendText", path, "calls SendText")
					}
				case *ast.SelectorExpr:
					if fn.Sel.Name == "SendText" {
						report("SendText", path, "calls SendText")
					}
					if name := pkgSel(fn, "connections"); connWriters[name] {
						report("connections", path, "writes to a connection (connections."+name+")")
					}
					if fn.Sel.Name == "Write" && isNetConn(fn.X, conns) {
						report("netconn", path, "writes to a net.Conn (.Write)")
					}
				}
			}
			return true
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

// pkgSel returns Sel's name when e is the selector pkg.Sel, else "".
func pkgSel(e ast.Expr, pkg string) string {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if x, ok := sel.X.(*ast.Ident); ok && x.Name == pkg {
		return sel.Sel.Name
	}
	return ""
}

// listenerKeys are the composite literals passed straight to
// events.RegisterListener or events.UnregisterListener.
func listenerKeys(f *ast.File) map[*ast.CompositeLit]bool {
	out := map[*ast.CompositeLit]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if name := pkgSel(call.Fun, "events"); name == "RegisterListener" || name == "UnregisterListener" {
			for _, arg := range call.Args {
				if lit, ok := arg.(*ast.CompositeLit); ok {
					out[lit] = true
				}
			}
		}
		return true
	})
	return out
}

// netConnNames are the names a file declares as net.Conn (fields, params,
// results, vars) or assigns from a .Accept() call.
func netConnNames(f *ast.File) map[string]bool {
	out := map[string]bool{}
	isConnType := func(e ast.Expr) bool { return pkgSel(e, "net") == "Conn" }
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
	conns := netConnNames(f)
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
