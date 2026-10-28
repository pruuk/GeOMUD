package actors

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
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

func TestConnectionControllerQueuesTheTextForTheAccount(t *testing.T) {
	var got []events.Message
	id := events.RegisterListener(events.Message{}, func(e events.Event) events.ListenerReturn {
		if m, ok := e.(events.Message); ok && m.UserId == 4242 {
			got = append(got, m)
		}
		return events.Continue
	})
	defer events.UnregisterListener(events.Message{}, id)

	NewConnectionController(&accounts.Account{UserId: 4242}).Perceive(perception.Perception{Text: "hello"})
	events.ProcessEvents()
	require.Len(t, got, 1)
	assert.Equal(t, "hello\n", got[0].Text)
}

func TestConnectionControllerWithNoAccountDoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		NewConnectionController(nil).Perceive(perception.Perception{Text: "hello"})
	})
}

func TestConnectionControllerIgnoresEmptyText(t *testing.T) {
	acc := &accounts.Account{UserId: 9}
	assert.NotPanics(t, func() {
		NewConnectionController(acc).Perceive(perception.Perception{})
	})
}

// bodyMob is the helper in registry_test.go.
func TestBehaviourControllerTickFiresTheBehaviourHook(t *testing.T) {
	m := bodyMob()
	var got []*mobs.Mob
	OnBehaviourTick.Register(func(x *mobs.Mob) *mobs.Mob {
		if x == m {
			got = append(got, x)
		}
		return x
	})

	NewBehaviourController(m).Tick()

	assert.Equal(t, []*mobs.Mob{m}, got)
}

func TestConnectionControllerTickDoesNothing(t *testing.T) {
	assert.NotPanics(t, func() { NewConnectionController(nil).Tick() })
}
