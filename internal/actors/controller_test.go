package actors

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBehaviourControllerFiresTheStructuredPerception(t *testing.T) {
	m := bodyMob()
	var got []BehaviourPerception
	OnBehaviourPerceive.Register(func(bp BehaviourPerception) BehaviourPerception {
		if bp.Mob == m {
			got = append(got, bp)
		}
		return bp
	})

	p := perception.Perception{Role: perception.Observer, SenseUsed: perception.Sound,
		Message: perception.Message{Kind: perception.WorldEvent, Senses: perception.Sound}}
	NewBehaviourController(m).Perceive(p)
	require.Len(t, got, 1)
	assert.Equal(t, p, got[0].Perception)

	NewBehaviourController(m).Perceive(perception.Perception{Message: perception.Message{Kind: perception.Interface}})
	assert.Len(t, got, 1, "a behaviour drops interface text")
}

func TestConnectionControllerIgnoresEmptyText(t *testing.T) {
	acc := &accounts.Account{UserId: 9}
	assert.NotPanics(t, func() {
		NewConnectionController(acc).Perceive(perception.Perception{})
	})
}
