package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Flee(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	if user.Character.Aggro == nil || user.Character.Aggro.Type != characters.Flee {
		user.SendText(`You attempt to flee...`)

		user.Character.Aggro = &characters.Aggro{}
		user.Character.Aggro.Type = characters.Flee
		events.AddToQueue(events.AggroChanged{UserId: user.UserId, RoomId: user.Character.RoomId})
	}

	return true, nil
}
