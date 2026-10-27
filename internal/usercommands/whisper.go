package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Whisper(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {
		user.SendText("Whisper to who?")
		return true, nil
	}

	whisperName := args[0]
	if len(rest) < len(whisperName)+1 {
		user.SendText("You need to specify a message.")
		return true, nil
	}

	rest = strings.TrimSpace(rest[len(whisperName)+1:])

	toUser := accounts.GetByCharacterName(whisperName)
	if toUser == nil {
		user.SendText("You can't find anyone by that name.")
		return true, nil
	}

	targetIsMod := toUser.Role != accounts.RoleUser

	if user.Muted && !targetIsMod {
		user.SendText(`You are <ansi fg="alert-5">MUTED</ansi>. You can only send <ansi fg="command">whisper</ansi>'s to Admins and Moderators.`)
		return true, nil
	}

	toUser.SendText(fmt.Sprintf(`<ansi fg="white">***</ansi> <ansi fg="black-bold"><ansi fg="username">%s</ansi> whispers, "%s"</ansi> <ansi fg="white">***</ansi>`, user.Character.Name, rest))

	user.SendText(fmt.Sprintf(`You sent a <ansi fg="command">whisper</ansi> to <ansi fg="username">%s</ansi>`, toUser.Character.Name))

	events.AddToQueue(events.Communication{
		SourceUserId: user.UserId,
		TargetUserId: toUser.UserId,
		CommType:     `whisper`,
		Name:         user.Character.Name,
		Message:      rest,
	})

	return true, nil
}
