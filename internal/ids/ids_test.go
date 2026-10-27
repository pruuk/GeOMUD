package ids

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var format = regexp.MustCompile(`^0x[0-9a-f]{16}:0x[0-9a-f]{16}$`)

func TestNewIsUniqueAndKeepsTheCell(t *testing.T) {
	a, b := New(NoCell), New(0x89c25a31ec0f0e5f)
	assert.False(t, a.IsZero())
	assert.NotEqual(t, a, b)
	assert.NotEqual(t, New(NoCell), New(NoCell))
	assert.Equal(t, uint64(0), a.Cell())
	assert.Equal(t, uint64(0x89c25a31ec0f0e5f), b.Cell())
	assert.Regexp(t, format, a.String())
	assert.Equal(t, "0x89c25a31ec0f0e5f:", b.String()[:19])
}

func TestZero(t *testing.T) {
	var z ID
	assert.True(t, z.IsZero())
	assert.Equal(t, "", z.String())
}

func TestParseRoundTrip(t *testing.T) {
	id := New(0x1234)
	got, err := Parse(id.String())
	require.NoError(t, err)
	assert.Equal(t, id, got)

	for _, bad := range []string{"", "0x1:0x2", "1234567890123456:1234567890123456", "0x89c25a31ec0f0e5f-0xd134e199a405a163",
		"0x89c25a31ec0f0e5f:0xD134E199A405A163", "0x89c25a31ec0f0e5f:0x0000000000000000", "#12", "@5"} {
		_, err := Parse(bad)
		assert.Error(t, err, bad)
	}
}

type holder struct {
	Id    ID `yaml:"id,omitempty"`
	Other int
}

func TestTextMarshalling(t *testing.T) {
	id := New(0x42)

	j, err := json.Marshal(holder{Id: id})
	require.NoError(t, err)
	var jh holder
	require.NoError(t, json.Unmarshal(j, &jh))
	assert.Equal(t, id, jh.Id)

	y, err := yaml.Marshal(holder{Id: id})
	require.NoError(t, err)
	assert.Contains(t, string(y), "id: "+id.String())
	var yh holder
	require.NoError(t, yaml.Unmarshal(y, &yh))
	assert.Equal(t, id, yh.Id)

	empty, err := yaml.Marshal(holder{Other: 1})
	require.NoError(t, err)
	assert.NotContains(t, string(empty), "id:", "a zero ID is omitted")
}
