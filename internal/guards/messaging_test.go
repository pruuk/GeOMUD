package guards

import (
	"go/ast"
	"strings"
	"testing"
)

// TestOnlyMessagingWritesToConnections enforces the slice 3 guard (spec
// section 2): text reaches a connection only through internal/messaging.
// connectionWriters lists today's other writers; slice 7 moves them and the
// list may only shrink (an entry that no longer writes fails the test). GMCP
// keeps its own path for good. main.go also writes "server full" notices
// straight to a net.Conn before a connection is registered (main.go,
// "Server is full"); this guard does not see those, and slice 7 moves them.
func TestOnlyMessagingWritesToConnections(t *testing.T) {
	writers := map[string]bool{"SendTo": true, "SendRawTo": true, "Broadcast": true}
	permanent := map[string]string{
		"modules/gmcp/gmcp.go": "GMCP keeps its own path (spec section 2)",
	}
	connectionWriters := map[string]string{
		"main.go": "slice 7: connections.Broadcast (copyover complete), no messaging path yet",
	}
	inside := func(path string) bool {
		return strings.HasPrefix(path, "internal/connections/") ||
			strings.HasPrefix(path, "internal/messaging/") ||
			strings.HasPrefix(path, "internal/perception/")
	}

	used := map[string]bool{}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok || !writers[sel.Sel.Name] {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); !ok || x.Name != "connections" {
				return true
			}
			switch {
			case inside(path):
			case permanent[path] != "":
			case connectionWriters[path] != "":
				used[path] = true
			default:
				t.Errorf("%s: writes to a connection (connections.%s) outside internal/messaging", path, sel.Sel.Name)
			}
			return true
		})
	}
	for path := range connectionWriters {
		if !used[path] {
			t.Errorf("%s no longer writes to a connection: remove it from connectionWriters", path)
		}
	}
}
