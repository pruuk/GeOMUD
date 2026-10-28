package messaging

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokensInANameAreNotSubstituted(t *testing.T) {
	assert.Equal(t, `{target} hits Bob.`, resolveTokens(`{actor} hits {target}.`, `{target}`, `Bob`, ``))
	assert.Equal(t, `{direction} from <ansi fg="exit">north</ansi>`, resolveTokens(`{actor} from {direction}`, `{direction}`, ``, `north`))
}

func TestLiteralBracesPrintAsTyped(t *testing.T) {
	text := "{actor} says, \"" + Literal("{target} {{x}") + "\""
	assert.Equal(t, `Fred says, "{target} {{x}"`, resolveTokens(text, `Fred`, `Bob`, `north`))
	assert.Equal(t, `no braces`, Literal(`no braces`))
}

func TestCapitalisedSomeoneOnlyAtTheStart(t *testing.T) {
	assert.Equal(t, `Someone waves at someone.`, resolveTokens(`{actor} waves at {target}.`, `someone`, `someone`, ``))
	assert.Equal(t, `<ansi fg="exit">north</ansi>`, resolveTokens(`{direction}`, ``, ``, `north`))
	assert.Equal(t, `no tokens`, resolveTokens(`no tokens`, `a`, `b`, `c`))
}

// The production enter and leave wrappers (_datafiles/config.yaml) put
// whitespace, a tag and ">>> " before the line; {actor} still opens the
// sentence.
func TestCapitalisedSomeoneAfterLeadingMarkup(t *testing.T) {
	const enter = "   <ansi fg=\"enters-message\"> >>> </ansi>%s\n"
	const leave = "   <ansi fg=\"leaves-message\"> >>> </ansi>%s\n"
	cases := []struct{ name, text, want string }{
		{"enter wrapper", fmt.Sprintf(enter, `{actor} enters.`),
			fmt.Sprintf(enter, `Someone enters.`)},
		{"leave wrapper", fmt.Sprintf(leave, `{actor} leaves.`),
			fmt.Sprintf(leave, `Someone leaves.`)},
		{"left arrows", ` <ansi fg="x"><<<</ansi> {actor} leaves.`,
			` <ansi fg="x"><<<</ansi> Someone leaves.`},
		{"plain start", `{actor} waves.`, `Someone waves.`},
		{"mid-sentence", `The rat bites {actor}.`, `The rat bites someone.`},
		{"first word not the token", fmt.Sprintf(enter, `A door opens for {actor}.`),
			fmt.Sprintf(enter, `A door opens for someone.`)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, resolveTokens(c.text, `someone`, `Bob`, ``))
		})
	}
}
