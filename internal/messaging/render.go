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

// resolveTokens fills {actor}, {target} and {direction}. A "someone" that
// opens the text is capitalised.
func resolveTokens(text, actorName, targetName, direction string) string {
	if direction != `` {
		direction = fmt.Sprintf(`<ansi fg="exit">%s</ansi>`, direction)
	}
	for _, tok := range []struct{ token, value string }{
		{`{actor}`, actorName}, {`{target}`, targetName}, {`{direction}`, direction},
	} {
		if strings.HasPrefix(text, tok.token) && strings.HasPrefix(tok.value, `someone`) {
			text = `Someone` + tok.value[len(`someone`):] + text[len(tok.token):]
		}
		text = strings.ReplaceAll(text, tok.token, tok.value)
	}
	return text
}
