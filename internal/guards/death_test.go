package guards

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// deathScanned is the engine's round and command code, where the slice 3
// death guards apply. characters, rooms and the rulesets are outside it
// (slice 4 moves the characters checks; the import cycle keeps them there
// for now).
var deathScanned = []string{"internal/hooks/", "internal/usercommands/", "internal/death/"}

// itemIsDisabled lists the scanned files that call items.Item.IsDisabled
// (a disabled equipment slot, not a body's condition), with why. A file on
// it may still not call .Character.IsDisabled().
var itemIsDisabled = map[string]string{
	"internal/usercommands/look.panels.go":      "the equipment and corpse panels skip a disabled slot",
	"internal/usercommands/inventory.panels.go": "the inventory panel skips a disabled slot",
}

// plantedDeath is the death guards' own control: source with exactly what
// each guard must catch. It is parsed, never compiled.
const plantedDeath = `package planted

func f(a, c, m, u, itm any) bool {
	m.Command("suicide")
	u.CommandFlagged(` + "`suicide`" + `, 0)
	events.AddToQueue(events.Input{ActorId: a, InputText: "suicide"})
	_ = itm.IsDisabled()
	return a.Character.IsDisabled() || c.Health <= -10 || 1 > c.Health || c.Health == (0) || c.Health > c.HealthMax
}
`

func parsePlanted(t *testing.T) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "planted.go", plantedDeath, 0)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// deathPrefix is the deathScanned prefix path is under, or "".
func deathPrefix(path string) string {
	for _, p := range deathScanned {
		if strings.HasPrefix(path, p) {
			return p
		}
	}
	return ""
}

// TestEngineAsksTheRulesetWhoIsDead enforces Phase 5 slice 3: the engine
// compares no health with a number and never asks a body whether it is
// disabled; it asks rules.Get().Condition.
func TestEngineAsksTheRulesetWhoIsDead(t *testing.T) {
	files, healthCompares := 0, 0
	reached := map[string]int{}
	itemCalls := map[string]bool{}
	for path, f := range goFiles(t) {
		p := deathPrefix(path)
		if p == "" {
			continue
		}
		reached[p]++
		files++
		literal, compares := healthLiteralCompares(f)
		healthCompares += compares
		for range literal {
			t.Errorf("%s: compares .Health with a number; ask rules.Get().Condition (slice 3)", path)
		}
		for _, onCharacter := range isDisabledCalls(f) {
			if onCharacter {
				t.Errorf("%s: calls .Character.IsDisabled(); ask rules.Get().Condition (slice 3)", path)
				continue
			}
			if _, ok := itemIsDisabled[path]; !ok {
				t.Errorf("%s: calls .IsDisabled(); a body's: ask rules.Get().Condition (slice 3); an item's: add the file to itemIsDisabled with why", path)
				continue
			}
			itemCalls[path] = true
		}
	}
	for path, why := range itemIsDisabled {
		if !itemCalls[path] {
			t.Errorf("stale allowlist: %s (%s) no longer calls IsDisabled; remove it from itemIsDisabled", path, why)
		}
	}
	// Positive controls: the walk reached the packages and saw a .Health
	// comparison (with a variable: the regen pass's change check) ...
	if files < 50 {
		t.Errorf("control: scanned only %d files under %v; the guard may be scanning the wrong place", files, deathScanned)
	}
	// ... every prefix reached a file (a typo in one, or a package that
	// moved, would otherwise go unseen behind the other two) ...
	for _, p := range deathScanned {
		if reached[p] == 0 {
			t.Errorf("control: no file under %s; a scope prefix reaches nothing", p)
		}
	}
	if healthCompares == 0 {
		t.Error("control: no .Health comparison seen in the scanned packages; the guard may not be matching")
	}
	// ... and the matchers find exactly what is planted: three literal
	// compares (signed, reversed, in brackets) and not the compare with a
	// variable; one body IsDisabled call and one other.
	planted := parsePlanted(t)
	if literal, compares := healthLiteralCompares(planted); len(literal) != 3 || compares != 4 {
		t.Errorf("control: found %d literal health compares of 3 and %d health compares of 4 in the planted source; the literal check is broken", len(literal), compares)
	}
	onCharacter, other := 0, 0
	for _, c := range isDisabledCalls(planted) {
		if c {
			onCharacter++
		} else {
			other++
		}
	}
	if onCharacter != 1 || other != 1 {
		t.Errorf("control: found %d .Character.IsDisabled calls of 1 and %d other IsDisabled calls of 1 in the planted source", onCharacter, other)
	}
}

// TestNothingQueuesSuicide enforces Phase 5 slice 3: a death runs
// death.Die; no code queues the suicide command as input.
func TestNothingQueuesSuicide(t *testing.T) {
	inputs, commands := 0, 0
	for path, f := range goFiles(t) {
		found, in, cmd := suicideQueues(f)
		inputs += in
		commands += cmd
		for range found {
			t.Errorf("%s: queues the suicide command (an events.Input or a Command call); a death runs death.Die", path)
		}
	}
	// Positive controls: inputs and command calls exist (Account.Command
	// builds an events.Input; DoCombat queues "equip") ...
	if inputs == 0 {
		t.Error("control: no events.Input literal seen; the guard may not be matching")
	}
	if commands == 0 {
		t.Error("control: no Command call seen; the guard may not be matching")
	}
	// ... and the matcher finds the three planted suicides (Command,
	// CommandFlagged with a raw string, an events.Input).
	if found, _, _ := suicideQueues(parsePlanted(t)); len(found) != 3 {
		t.Errorf("control: found %d of the 3 planted suicides; the suicide check is broken", len(found))
	}
}

// healthLiteralCompares returns where f compares a .Health selector with a
// number literal, and how many .Health comparisons it has in all.
func healthLiteralCompares(f *ast.File) (literal []token.Pos, compares int) {
	ast.Inspect(f, func(n ast.Node) bool {
		be, ok := n.(*ast.BinaryExpr)
		if !ok || !isComparison(be.Op) {
			return true
		}
		l, r := isHealth(be.X), isHealth(be.Y)
		if !l && !r {
			return true
		}
		compares++
		if (l && isNumber(be.Y)) || (r && isNumber(be.X)) {
			literal = append(literal, be.Pos())
		}
		return true
	})
	return literal, compares
}

// isDisabledCalls returns one entry per X.IsDisabled() call in f: true when
// X is a .Character selector (a body's IsDisabled, characters.Character).
// The guard sees syntax, not types: a body reached through another name
// (char := a.Character; char.IsDisabled()) reads as "other", which is
// allowed only in the itemIsDisabled files.
func isDisabledCalls(f *ast.File) []bool {
	out := []bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "IsDisabled" {
			return true
		}
		recv, isSel := sel.X.(*ast.SelectorExpr)
		out = append(out, isSel && recv.Sel.Name == "Character")
		return true
	})
	return out
}

// suicideQueues returns where f queues the suicide command, and how many
// events.Input literals and Command calls it has in all.
func suicideQueues(f *ast.File) (found []token.Pos, inputs, commands int) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CompositeLit:
			sel, ok := x.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Input" {
				return true
			}
			if pkg, ok := sel.X.(*ast.Ident); !ok || pkg.Name != "events" {
				return true
			}
			inputs++
			for _, el := range x.Elts {
				kv, ok := el.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "InputText" && isSuicide(kv.Value) {
					found = append(found, x.Pos())
				}
			}
		case *ast.CallExpr:
			sel, ok := x.Fun.(*ast.SelectorExpr)
			if !ok || (sel.Sel.Name != "Command" && sel.Sel.Name != "CommandFlagged") || len(x.Args) == 0 {
				return true
			}
			commands++
			if isSuicide(x.Args[0]) {
				found = append(found, x.Pos())
			}
		}
		return true
	})
	return found, inputs, commands
}

func isComparison(op token.Token) bool {
	switch op {
	case token.LSS, token.LEQ, token.GTR, token.GEQ, token.EQL, token.NEQ:
		return true
	}
	return false
}

// isHealth reports whether e is a .Health selector (a body's health).
func isHealth(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Health"
}

// isNumber reports whether e is a number literal, signed or in brackets.
func isNumber(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.BasicLit:
		return x.Kind == token.INT || x.Kind == token.FLOAT
	case *ast.UnaryExpr:
		return (x.Op == token.SUB || x.Op == token.ADD) && isNumber(x.X)
	case *ast.ParenExpr:
		return isNumber(x.X)
	}
	return false
}

// isSuicide reports whether e is the string literal suicide.
func isSuicide(e ast.Expr) bool {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	s, err := strconv.Unquote(lit.Value)
	return err == nil && strings.TrimSpace(s) == "suicide"
}
