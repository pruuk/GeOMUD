package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapitalisedSomeoneOnlyAtTheStart(t *testing.T) {
	assert.Equal(t, `Someone waves at someone.`, resolveTokens(`{actor} waves at {target}.`, `someone`, `someone`, ``))
	assert.Equal(t, `<ansi fg="exit">north</ansi>`, resolveTokens(`{direction}`, ``, ``, `north`))
	assert.Equal(t, `no tokens`, resolveTokens(`no tokens`, `a`, `b`, `c`))
}
