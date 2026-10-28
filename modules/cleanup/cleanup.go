package cleanup

import (
	"embed"
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (

	//////////////////////////////////////////////////////////////////////
	// NOTE: The below //go:embed directive is important!
	// It embeds the relative path into the var below it.
	//////////////////////////////////////////////////////////////////////

	//go:embed files/*
	files embed.FS
)

// ////////////////////////////////////////////////////////////////////
// NOTE: The init function in Go is a special function that is
// automatically executed before the main function within a package.
// It is used to initialize variables, set up configurations, or
// perform any other setup tasks that need to be done before the
// program starts running.
// ////////////////////////////////////////////////////////////////////
func init() {
	//
	// We can use all functions only, but this demonstrates
	// how to use a struct
	//
	c := CleanupModule{
		plug: plugins.New(`cleanup`, `1.0`),
	}

	//
	// Add the embedded filesystem
	//
	if err := c.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	c.plug.Web.AdminPage("Config", "cleanup-config", "html/admin/cleanup-config.html", true, "Modules", "Cleanup", "Configure automated world cleanup schedules and rules.", "Automated cleanup module removing stale corpses, items, and temporary objects.", nil)
	c.plug.Web.AdminPage("About", "cleanup-about", "html/admin/cleanup-about.html", true, "Modules", "Cleanup", "Information and version details for the Cleanup module.", "", nil)
	//
	// Register the commands (players and NPCs run the same ones)
	//
	c.plug.AddUserCommand("trash", c.trashCommand, false, false)
	c.plug.AddUserCommand("bury", c.buryCommand, false, false)

}

// Using a struct gives a way to store longer term data.
type CleanupModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin

	DefaultTrashExperienceValue int
}

func (c *CleanupModule) loadConfig() {

	if experienceValue, ok := c.plug.Config.Get("ExperienceValue").(int); ok {
		c.DefaultTrashExperienceValue = experienceValue
	}

}

func (c *CleanupModule) trashCommand(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	// Check whether the actor has an item in their inventory that matches
	matchItem, found := actor.Character.FindInBackpack(rest)

	if !found {
		usercommands.Tell(actor, fmt.Sprintf(`You don't have a "%s" to trash.`, rest))
	} else {

		c.loadConfig()

		isSneaking := actor.Character.HasBuffFlag("hidden")

		actor.Character.RemoveItem(matchItem)

		events.AddToQueue(events.ItemOwnership{
			ActorId: actor.Character.Id,
			Item:    matchItem,
			Gained:  false,
		})

		trashText := fmt.Sprintf(`You trash the <ansi fg="item">%s</ansi> for good.`, matchItem.DisplayName())
		if isSneaking {
			usercommands.Tell(actor, trashText)
		} else {
			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Actor:        actor.Id,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ActorText:    trashText,
				ObserverText: fmt.Sprintf(`{actor} destroys <ansi fg="item">%s</ansi>...`, matchItem.DisplayName()),
			})
		}

		// Experience goes to an account.
		if user := actor.Controller().Account(); user != nil && c.DefaultTrashExperienceValue > 0 {

			// Grant experience equal to a tenth of the item value
			iSpec := matchItem.GetSpec()

			xpGrant := int(float64(iSpec.Value) / 10)
			if xpGrant < c.DefaultTrashExperienceValue {
				xpGrant = c.DefaultTrashExperienceValue
			}
			user.GrantXP(xpGrant, `trash cleanup`)

		}

	}

	return true, nil
}

func (c *CleanupModule) buryCommand(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	if len(args) == 0 {
		usercommands.Tell(actor, "Bury what?")
		return true, nil
	}

	if corpse, corpseFound := room.FindCorpse(rest); corpseFound {

		if room.RemoveCorpse(corpse) {

			corpseColor := `mob-corpse`
			if corpse.UserId > 0 {
				corpseColor = `user-corpse`
			}

			messaging.Send(perception.Message{
				Kind:         perception.WorldEvent,
				Actor:        actor.Id,
				Senses:       perception.Sight | perception.Sound,
				Range:        perception.Room,
				ActorText:    fmt.Sprintf(`You bury the <ansi fg="%s">%s corpse</ansi>.`, corpseColor, corpse.Character.Name),
				ObserverText: fmt.Sprintf(`{actor} buries the <ansi fg="%s">%s corpse</ansi>.`, corpseColor, corpse.Character.Name),
			})
			return true, nil

		}

		return true, nil
	}

	usercommands.Tell(actor, fmt.Sprintf("You don't see a %s around for burying.", rest))

	return true, nil
}
