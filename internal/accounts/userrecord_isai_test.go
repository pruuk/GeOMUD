package accounts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestUserRecordIsAISerializes(t *testing.T) {
	u := Account{Username: "tester", IsAI: true}

	out, err := yaml.Marshal(u)
	require.NoError(t, err)
	assert.Contains(t, string(out), "isai: true")

	var back Account
	require.NoError(t, yaml.Unmarshal(out, &back))
	assert.True(t, back.IsAI, "IsAI should round-trip true")
}

func TestUserRecordIsAIOmittedWhenFalse(t *testing.T) {
	u := Account{Username: "human"}

	out, err := yaml.Marshal(u)
	require.NoError(t, err)
	assert.NotContains(t, string(out), "isai:", "isai should be omitted when false")
}
