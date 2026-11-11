package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/rules"
)

// AutoHeal is the engine's regen pass (Phase 5 slice 3): every round, for
// every actor with a body, the ruleset says what it regenerates; the engine
// applies it, sends the ruleset's lines and reports an account's changed
// vitals. A ruleset that wants bleeding returns a negative Regen.
func AutoHeal(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	for _, a := range fighters() {
		healthStart := a.Character.Health
		manaStart := a.Character.Mana

		r := rules.Get().Regen(a, evt.RoundNumber)
		if r.Changes() {
			a.Character.Heal(r.Health, r.Mana)
		}
		for _, m := range r.Messages {
			messaging.Send(m)
		}

		user := accountOf(a)
		if user == nil {
			continue
		}

		// If it has changed, send an update
		if a.Character.Health != healthStart || a.Character.Mana != manaStart {

			// Trigger a redraw, but only if the users prompt has changed.
			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId, OnlyIfChanged: true}, 100)

			events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})

		}
	}

	return events.Continue
}
