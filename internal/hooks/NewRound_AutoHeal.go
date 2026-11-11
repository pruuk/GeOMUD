package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
)

// AutoHeal is the engine's regen pass (Phase 5 slice 3): every round, for
// every actor with a body, the ruleset says what it regenerates; the engine
// applies it, sends the ruleset's lines and reports an account's changed
// vitals. The bleed-out stays engine code until slice 3's death commit
// removes it.
func AutoHeal(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	for _, a := range fighters() {
		user := accountOf(a)

		healthStart := a.Character.Health
		manaStart := a.Character.Mana

		if user != nil && bleeds(a, evt.RoundNumber) {

			if a.Character.Health <= -10 {

				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

			} else {
				a.Character.Health--
				if room := rooms.LoadRoom(a.Character.RoomId); room != nil {
					messaging.Send(perception.Message{
						Kind:         perception.WorldEvent,
						Actor:        a.Character.Id,
						Senses:       perception.Sight | perception.Sound,
						Range:        perception.Room,
						ActorText:    `<ansi fg="red">you are bleeding out!</ansi>`,
						ObserverText: `{actor} is <ansi fg="red">bleeding out</ansi>! Somebody needs to provide aid!`,
					})
				} else {
					messaging.Tell(a.Character.Id, `<ansi fg="red">you are bleeding out!</ansi>`)
				}
			}

		} else {

			r := rules.Get().Regen(a, evt.RoundNumber)
			if r.Changes() {
				a.Character.Heal(r.Health, r.Mana)
			}
			for _, m := range r.Messages {
				messaging.Send(m)
			}
		}

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

// bleeds reports whether a downed account bleeds this round: the old
// AutoHeal's gates (every 3rd round, out of a fight, outside the death
// recovery room by raw ID) and a body the ruleset does not call up.
func bleeds(a *actors.Actor, round uint64) bool {
	return round%3 == 0 &&
		a.Character.Aggro == nil &&
		a.Character.RoomId != int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom) &&
		rules.Get().Condition(a) != rules.Up
}
