package characters

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/stretchr/testify/assert"
)

func TestNewCharm(t *testing.T) {
	by1, by2, by3 := ids.New(ids.NoCell), ids.New(ids.NoCell), ids.New(ids.NoCell)
	tests := []struct {
		name          string
		by            ids.ID
		rounds        int
		expireCommand string
		want          *CharmInfo
	}{
		{
			name:          "Normal values",
			by:            by1,
			rounds:        5,
			expireCommand: "emote bows",
			want:          &CharmInfo{By: by1, RoundsRemaining: 5, ExpiredCommand: "emote bows"},
		},
		{
			name:          "Permanent charm",
			by:            by2,
			rounds:        CharmPermanent,
			expireCommand: CharmExpiredRest,
			want:          &CharmInfo{By: by2, RoundsRemaining: CharmPermanent, ExpiredCommand: CharmExpiredRest},
		},
		{
			name:          "Empty expire command",
			by:            by3,
			rounds:        10,
			expireCommand: "",
			want:          &CharmInfo{By: by3, RoundsRemaining: 10, ExpiredCommand: ""},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := NewCharm(tt.by, tt.rounds, tt.expireCommand)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestCharmInfo_Expire(t *testing.T) {
	tests := []struct {
		name     string
		initial  *CharmInfo
		expected int
	}{
		{
			name:     "Expire normal charm",
			initial:  &CharmInfo{By: ids.New(ids.NoCell), RoundsRemaining: 5, ExpiredCommand: "emote bows"},
			expected: 0,
		},
		{
			name:     "Expire permanent charm",
			initial:  &CharmInfo{By: ids.New(ids.NoCell), RoundsRemaining: CharmPermanent, ExpiredCommand: CharmExpiredRest},
			expected: 0,
		},
		{
			name:     "Expire already expired charm",
			initial:  &CharmInfo{By: ids.New(ids.NoCell), RoundsRemaining: 0, ExpiredCommand: "emote leaves"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Expire()
			assert.Equal(t, tt.expected, tt.initial.RoundsRemaining)
		})
	}
}
