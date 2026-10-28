package messaging

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
)

// nameFor is how receiver reads subject's name: the plain coloured name
// today's room text uses, or "someone" when unseen.
func nameFor(subject *actors.Actor, seen bool) string {
	if subject == nil {
		return ``
	}
	if !seen {
		return `someone`
	}
	style := `mobname`
	if ctrl := subject.Controller(); ctrl != nil && ctrl.Account() != nil {
		style = `username`
	}
	return fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, style, subject.Character.Name)
}

// Literal escapes typed text for a world event's text: every "{" is
// doubled, so the text prints as typed and is never read as a token.
func Literal(s string) string {
	return strings.ReplaceAll(s, `{`, `{{`)
}

// leadIn is the length of the decoration before a text's first word:
// whitespace, <ansi ...> and </ansi> tags, and the enter and leave
// wrappers' ">>>" or "<<<" arrows (config.yaml EnterRoomMessageWrapper,
// ExitRoomMessageWrapper).
func leadIn(text string) int {
	i := 0
	for i < len(text) {
		rest := text[i:]
		switch {
		case rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\r' || rest[0] == '\n':
			i++
		case strings.HasPrefix(rest, `</ansi>`):
			i += len(`</ansi>`)
		case strings.HasPrefix(rest, `<ansi`):
			end := strings.IndexByte(rest, '>')
			if end < 0 {
				return i
			}
			i += end + 1
		case strings.HasPrefix(rest, `>>>`), strings.HasPrefix(rest, `<<<`):
			i += 3
		default:
			return i
		}
	}
	return i
}

// resolveTokens fills {actor}, {target} and {direction}. A "someone" that
// opens the sentence is capitalised: only the decoration leadIn skips may
// come before it. An escaped "{{" (Literal) becomes "{" and
// is never read as the start of a token.
func resolveTokens(text, actorName, targetName, direction string) string {
	if direction != `` {
		direction = fmt.Sprintf(`<ansi fg="exit">%s</ansi>`, direction)
	}
	tokens := []struct{ token, value string }{
		{`{actor}`, actorName}, {`{target}`, targetName}, {`{direction}`, direction},
	}
	// Capitalise an opening someone first, then replace every token in one
	// pass so a value holding a token (a name like "{target}") is not
	// substituted again.
	lead := leadIn(text)
	for _, tok := range tokens {
		if strings.HasPrefix(text[lead:], tok.token) && strings.HasPrefix(tok.value, `someone`) {
			text = text[:lead] + `Someone` + tok.value[len(`someone`):] + text[lead+len(tok.token):]
			break
		}
	}
	// The escape pair goes first: the replacer tries pairs in argument order
	// at each position, so "{{target}" is "{" then plain "target}".
	pairs := make([]string, 0, len(tokens)*2+2)
	pairs = append(pairs, `{{`, `{`)
	for _, tok := range tokens {
		pairs = append(pairs, tok.token, tok.value)
	}
	return strings.NewReplacer(pairs...).Replace(text)
}
