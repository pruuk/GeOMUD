package accounts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// Charm and companions were deleted in Phase 5.1. A user file saved before
// that still carries `charmedmobs:` under the character and must load as it
// does in LoadUser (a plain yaml.Unmarshal into an Account).
func TestOldUserFileWithCharmedMobsStillLoads(t *testing.T) {
	const old = `username: oldtimer
character:
  name: Zed
  level: 3
  charmedmobs:
  - 5
`
	loaded := &Account{}
	require.NoError(t, yaml.Unmarshal([]byte(old), loaded))
	assert.Equal(t, "oldtimer", loaded.Username)
	require.NotNil(t, loaded.Character)
	assert.Equal(t, "Zed", loaded.Character.Name)
	assert.Equal(t, 3, loaded.Character.Level)
}
