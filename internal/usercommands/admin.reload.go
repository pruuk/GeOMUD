package usercommands

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/language"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

/*
* Role Permissions:
* reload 				(All)
 */
func Reload(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	if rest == "" {
		infoOutput, _ := templates.Process("admincommands/help/command.reload", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	switch strings.ToLower(rest) {
	case `items`:
		items.LoadDataFiles()
		user.SendText(`Items reloaded.`)
	case `biomes`:
		rooms.LoadBiomeDataFiles()
		user.SendText(`Biomes reloaded.`)
	case `buffs-flags`:
		buffs.LoadFlagDataFiles()
		user.SendText(`Buff flags reloaded.`)
	case `skills`:
		skills.LoadDataFiles()
		skills.LoadProfessionDataFiles()
		user.SendText(`Skills and professions reloaded.`)
	case `translations`:
		ok := language.ReloadTranslation()
		if !ok {
			user.SendText(`Translations reload failed.`)
		} else {
			user.SendText(`Translations reloaded.`)
		}
	default:
		user.SendText(`Unknown reload command.`)
	}
	return true, nil
}
