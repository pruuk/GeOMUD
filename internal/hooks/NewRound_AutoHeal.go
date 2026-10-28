package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

//
// Watches the rounds go by
// Applies autohealing where appropriate
//

func AutoHeal(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	// Every 3 rounds. Else, pass it along.
	if evt.RoundNumber%3 != 0 {
		return events.Continue
	}

	deathRecoveryRoomId := int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)

	onlineIds := accounts.GetOnlineUserIds()
	for _, userId := range onlineIds {
		user := accounts.GetByUserId(userId)

		// Only heal if not in combat
		if user.Character.Aggro != nil {
			continue
		}

		if user.Character.RoomId == deathRecoveryRoomId {
			continue
		}

		healthStart := user.Character.Health
		manaStart := user.Character.Mana

		if user.Character.Health < 1 {

			if user.Character.Health <= -10 {

				user.Command(`suicide`) // suicide drops all money/items and transports to land of the dead.

			} else {
				user.Character.Health--
				if room := rooms.LoadRoom(user.Character.RoomId); room != nil {
					messaging.Send(perception.Message{
						Kind:         perception.WorldEvent,
						Actor:        user.Character.Id,
						Senses:       perception.Sight | perception.Sound,
						Range:        perception.Room,
						ActorText:    `<ansi fg="red">you are bleeding out!</ansi>`,
						ObserverText: `{actor} is <ansi fg="red">bleeding out</ansi>! Somebody needs to provide aid!`,
					})
				} else {
					messaging.Tell(user.Character.Id, `<ansi fg="red">you are bleeding out!</ansi>`)
				}
			}

		} else {

			if user.Character.Health > 0 {
				user.Character.Heal(
					user.Character.HealthPerRound(),
					user.Character.ManaPerRound(),
				)
			}
		}

		// If it has changed, send an update
		if user.Character.Health != healthStart || user.Character.Mana != manaStart {

			// Trigger a redraw, but only if the users prompt has changed.
			events.AddToQueue(events.RedrawPrompt{UserId: user.UserId, OnlyIfChanged: true}, 100)

			events.AddToQueue(events.CharacterVitalsChanged{UserId: user.UserId})

		}

	}

	return events.Continue
}
