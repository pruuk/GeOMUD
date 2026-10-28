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
		Tell(actor, infoOutput)
		return true, nil
	}

	switch strings.ToLower(rest) {
	case `items`:
		items.LoadDataFiles()
		Tell(actor, `Items reloaded.`)
	case `biomes`:
		rooms.LoadBiomeDataFiles()
		Tell(actor, `Biomes reloaded.`)
	case `buffs-flags`:
		buffs.LoadFlagDataFiles()
		Tell(actor, `Buff flags reloaded.`)
	case `skills`:
		skills.LoadDataFiles()
		skills.LoadProfessionDataFiles()
		Tell(actor, `Skills and professions reloaded.`)
	case `translations`:
		ok := language.ReloadTranslation()
		if !ok {
			Tell(actor, `Translations reload failed.`)
		} else {
			Tell(actor, `Translations reloaded.`)
		}
	default:
		Tell(actor, `Unknown reload command.`)
	}
	return true, nil
}
