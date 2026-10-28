package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
)

func TestStorageAddItemGivesAnIdWhenMissing(t *testing.T) {
	var s Storage
	assert.True(t, s.AddItem(items.Item{ItemId: 5}))
	assert.False(t, s.Items[0].Id.IsZero())

	kept := s.Items[0].Id
	assert.True(t, s.AddItem(s.Items[0]))
	assert.Equal(t, kept, s.Items[1].Id, "an existing ID is kept")
}
