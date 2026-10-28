package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventFlagAddSetsTheFlag(t *testing.T) {
	f := CmdSecretly
	f.Add(CmdBlockInput)
	assert.True(t, f.Has(CmdBlockInput))
	assert.True(t, f.Has(CmdSecretly), "Add keeps the flags already set")

	f.Add(CmdBlockInput)
	assert.Equal(t, CmdSecretly|CmdBlockInput, f, "adding a flag twice changes nothing")
}

func TestEventFlagRemoveClearsOnlyThatFlag(t *testing.T) {
	f := CmdSecretly | CmdBlockInputUntilComplete
	f.Remove(CmdBlockInput)
	assert.False(t, f.Has(CmdBlockInput))
	assert.True(t, f.Has(CmdUnBlockInput))
	assert.True(t, f.Has(CmdSecretly))
}
