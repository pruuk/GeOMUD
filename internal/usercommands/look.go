package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/mapper"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
)

func Look(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	secretLook := flags.Has(events.CmdSecretly)

	visibility := room.GetVisibility()

	if visibility < 1 {
		if !actor.Character.HasBuffFlag("nightvision") {
			tell(actor, `You can't see anything!`)
			return true, nil
		}
	}

	isSneaking := actor.Character.HasBuffFlag("hidden")

	// trim off some fluff
	if len(rest) > 2 {
		if rest[0:3] == `at ` {
			rest = rest[3:]
		}
	}
	if len(rest) > 3 {
		if rest[0:4] == `the ` {
			rest = rest[4:]
		}
	}

	lookAt := rest

	// Handle an ordinary look with no target
	if len(lookAt) == 0 {

		if !secretLook && !isSneaking {
			room.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking around.`, nameColour(actor), actor.Character.Name),
				excludeId(actor),
			)

			// Make it a "secret looks" now because we don't want another look message sent out by the lookRoom() func
			secretLook = true
		}
		lookRoom(actor, room.RoomId, secretLook || isSneaking)
		return true, nil
	}

	//
	// look for any mobs, players, npcs
	//

	playerId, mobId := room.FindByName(lookAt)

	if playerId > 0 || !mobId.IsZero() {

		statusTxt := ""
		invTxt := ""

		if playerId > 0 {

			uPtr := accounts.GetByUserId(playerId)
			if uPtr == nil {
				return true, nil
			}
			u := *uPtr

			if !isSneaking {
				u.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking at you.`, nameColour(actor), actor.Character.Name),
				)

				room.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking at <ansi fg="username">%s</ansi>.`, nameColour(actor), actor.Character.Name, u.Character.Name),
					u.UserId)
			}

			tell(actor, buildDescriptionPanel(u.Character))

			itemNames := []string{}
			for _, item := range u.Character.Items {
				itemNames = append(itemNames, item.DisplayName())
			}

			tell(actor, buildInventoryLookPanel(&u.Character.Equipment, itemNames))

		} else if !mobId.IsZero() {

			m := mobs.GetInstance(mobId)
			if m == nil {
				return true, nil
			}

			if !isSneaking {
				targetName := m.Character.GetMobName(ids.ID{}).String()
				room.SendText(
					fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking at %s.`, nameColour(actor), actor.Character.Name, targetName),
					excludeId(actor),
				)
			}

			tell(actor, buildDescriptionPanel(&m.Character))

			itemNames := []string{}
			for _, item := range m.Character.Items {
				itemNames = append(itemNames, item.DisplayName())
			}

			tell(actor, buildInventoryLookPanel(&m.Character.Equipment, itemNames))
		}

		tell(actor, statusTxt)
		tell(actor, invTxt)

		return true, nil

	}

	containerName := room.FindContainerByName(lookAt)
	if containerName != `` {

		itemNames := []string{}
		itemNamesFormatted := []string{}

		container := room.Containers[containerName]

		if container.Lock.IsLocked() {
			tell(actor, ``)
			tell(actor, fmt.Sprintf(`The <ansi fg="container">%s</ansi> is locked.`, containerName))
			tell(actor, ``)
			return true, nil
		}

		if container.Gold > 0 {
			itemNames = append(itemNames, fmt.Sprintf(`%d gold`, container.Gold))
			itemNamesFormatted = append(itemNamesFormatted, fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, container.Gold))
		}

		for _, item := range container.Items {
			if !item.IsValid() {
				room.RemoveItem(item, false)
				continue
			}

			itemNames = append(itemNames, item.Name())
			itemNamesFormatted = append(itemNamesFormatted, fmt.Sprintf(`<ansi fg="itemname">%s</ansi>`, item.DisplayName()))
		}

		if len(container.Recipes) > 0 {

			tell(actor, ``)
			tell(actor, fmt.Sprintf(`You can <ansi fg="command">use</ansi> the <ansi fg="container">%s</ansi> if you put the following objects inside:`, containerName))

			for finalItemId, recipeList := range container.Recipes {

				neededItems := map[int]int{}

				for _, inputItemId := range recipeList {
					neededItems[inputItemId] += 1
				}

				tell(actor, ``)

				tell(actor, fmt.Sprintf(`    <ansi fg="230">To receive 1 <ansi fg="itemname">%s</ansi>:</ansi> `, specDisplayName(finalItemId)))

				for inputItemId, qtyNeeded := range neededItems {
					totalContained := container.Count(inputItemId)
					colorClass := "8" // None fulfilled
					if totalContained == qtyNeeded {
						colorClass = "10"
					} else if totalContained > 0 {
						colorClass = "3"
					}
					tell(actor, fmt.Sprintf(`        <ansi fg="%s">[%d/%d]</ansi> <ansi fg="itemname">%s</ansi>`, colorClass, totalContained, qtyNeeded, specDisplayName(inputItemId)))
				}

			}

		}

		tell(actor, ``)
		tell(actor, buildInsideContainerPanel(itemNames, itemNamesFormatted))
		tell(actor, ``)

		return true, nil
	}

	//
	// Check room exits
	//
	exitName, lookRoomId := room.FindExitByName(lookAt)
	// If nothing found, consider directional aliases
	if exitName == `` {

		if alias := keywords.TryDirectionAlias(lookAt); alias != lookAt {
			exitName, lookRoomId = room.FindExitByName(alias)
			if exitName != `` {
				lookAt = alias
			}
		}
	}

	if exitName != `` {

		if visibility < 2 {

			if !actor.Character.HasBuffFlag("nightvision") {
				biome := room.GetBiome()
				if !biome.IsLit() {
					tell(actor, `It's too dark to see anything in that direction.`)
					return true, nil
				}
			}

		}

		exitInfo, _ := room.GetExitInfo(exitName)
		if exitInfo.Lock.IsLocked() {
			tell(actor, fmt.Sprintf("The %s exit is locked.", exitName))
			return true, nil
		}

		tell(actor, fmt.Sprintf("You peer toward the %s.", exitName))
		if !isSneaking {
			room.SendText(fmt.Sprintf(`<ansi fg="%s">%s</ansi> peers toward the %s.`, nameColour(actor), actor.Character.Name, exitName), excludeId(actor))
		}

		lookRoom(actor, lookRoomId, secretLook || isSneaking)

		return true, nil

	}

	//
	// Check for anything in their backpack they might want to look at
	//
	lookItem, foundItem := actor.Character.FindInBackpack(lookAt)
	lookDestination := `in your backpack`
	if !foundItem {
		// Check for any equipment they are wearing they might want to look at
		lookItem, foundItem = actor.Character.FindOnBody(lookAt)
		lookDestination = `you are wearing`
	}

	if foundItem {

		tell(actor, ``)

		tell(actor,
			fmt.Sprintf(`You look at the <ansi fg="item">%s</ansi> %s:`, lookItem.DisplayName(), lookDestination),
		)

		tell(actor, ``)

		if !isSneaking {
			room.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> is admiring their <ansi fg="item">%s</ansi>.`, nameColour(actor), actor.Character.Name, lookItem.DisplayName()),
				excludeId(actor),
			)
		}

		tell(actor,
			lookItem.GetLongDescription(),
		)

		tell(actor, ``)

		return true, nil
	}

	//
	// Look for any nouns in the room info
	//
	foundNoun, foundDesc := room.FindNoun(lookAt)
	if len(foundNoun) > 0 {

		// Nouns may be multi-line
		finalNounDesc := []string{}
		for _, descLine := range strings.Split(foundDesc, "\n") {
			finalNounDesc = append(finalNounDesc, descLine)
		}

		tell(actor, ``)

		tell(actor,
			fmt.Sprintf(`You look at the <ansi fg="noun">%s</ansi>:`, foundNoun),
		)

		tell(actor, ``)

		if !isSneaking {

			renderNouns := false
			if user := actor.Controller().Account(); user != nil {
				renderNouns = user.HasRolePermission(`room.nouns`)
			}
			if actor.Character.HasBuffFlag("see-nouns") {
				renderNouns = true
			}

			if renderNouns && len(room.Nouns) > 0 {
				for idx, descLine := range finalNounDesc {
					for noun, _ := range room.Nouns {
						finalNounDesc[idx] = strings.Replace(descLine, noun, `<ansi fg="noun">`+noun+`</ansi>`, 1)
					}
				}
			}

			room.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> is examining the <ansi fg="noun">%s</ansi>.`, nameColour(actor), actor.Character.Name, foundNoun),
				excludeId(actor),
			)
		}

		for _, outLine := range finalNounDesc {
			tell(actor, outLine)
		}

		tell(actor, ``)

		return true, nil
	}

	if len(room.Corpses) > 0 {

		mobCorpseLookup := map[string]int{}
		mobCorpses := []string{}

		playerCorpseLookup := map[string]int{}
		playerCorpses := []string{}
		for idx, c := range room.Corpses {
			if c.Prunable {
				continue
			}

			if c.MobId > 0 {
				name := c.Character.Name + ` corpse`
				if _, ok := mobCorpseLookup[name]; !ok {
					mobCorpseLookup[name] = idx
					mobCorpses = append(mobCorpses, name)
				}
			}

			if c.UserId > 0 {
				name := c.Character.Name + ` corpse`
				if _, ok := playerCorpseLookup[name]; !ok {
					playerCorpseLookup[name] = idx
					playerCorpses = append(playerCorpses, name)
				}
			}
		}

		if corpse, corpseFound := room.FindCorpse(rest); corpseFound {

			corpseColor := `mob-corpse`
			if corpse.UserId > 0 {
				corpseColor = `user-corpse`
			}

			tell(actor, fmt.Sprintf(`You look at the <ansi fg="%s">%s corpse</ansi>.`, corpseColor, corpse.Character.Name))
			room.SendText(fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking at the <ansi fg="%s">%s corpse</ansi>.`, nameColour(actor), actor.Character.Name, corpseColor, corpse.Character.Name), excludeId(actor))

			tell(actor, buildCorpseDescriptionPanel(&corpse.Character))

			if configs.GetGamePlayConfig().Death.CorpseItems {
				tell(actor, buildCorpseInventoryPanel(&corpse))
			}

			return true, nil

		}

	}

	// Nothing found
	tell(actor, "Look at what???")

	return true, nil

}

// specDisplayName is the display name of an item kind, read from its spec so
// no item (and no item ID) is made. An unknown ID reads as the empty slot,
// as it did when a throwaway items.New was used for this.
func specDisplayName(itemId int) string {
	if items.GetItemSpec(itemId) == nil {
		itemId = 0
	}
	itm := items.Item{ItemId: itemId}
	return itm.DisplayName()
}

func lookRoom(actor *actors.Actor, roomId int, secretLook bool) {

	room := rooms.LoadRoom(roomId)

	if room == nil {
		return
	}

	// Make sure to prepare the room before anyone looks in if this is the first time someone has dealt with it in a while.
	if room.PlayerCt() < 1 {
		room.Prepare(true)
	}

	if !secretLook {
		// Find the exit back
		lookFromName := room.FindExitTo(actor.Character.RoomId)
		if lookFromName == "" {
			room.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking into the room from somewhere...`, nameColour(actor), actor.Character.Name),
				excludeId(actor),
			)
		} else {
			room.SendText(
				fmt.Sprintf(`<ansi fg="%s">%s</ansi> is looking into the room from the <ansi fg="exit">%s</ansi> exit`, nameColour(actor), actor.Character.Name, lookFromName),
				excludeId(actor),
			)
		}
	}

	// The room view reads account settings and is text for the caller only:
	// an actor without an account gets none.
	user := actor.Controller().Account()
	if user == nil {
		return
	}

	var details rooms.RoomTemplateDetails

	tinyMapOn := user.GetConfigOption(`tinymap`)
	if tinyMapOn == nil {
		tinyMapOn = true
	}

	if user.ScreenReader {
		tinyMapOn = false
	}

	if tinyMapOn.(bool) && roomId > 0 {

		zMapper := mapper.GetMapper(room.RoomId)
		if zMapper == nil {

			mudlog.Error("Map", "error", "Could not find mapper for zone:"+room.Zone)

			details = rooms.GetDetails(room, user)

		} else {

			c := mapper.Config{
				ZoomLevel: 1,
				Width:     5,
				Height:    5,
				UserId:    excludeId(actor),
			}

			c.OverrideSymbol(roomId, '@', `You`)

			output := zMapper.GetLimitedMap(room.RoomId, c)
			tinyMap := []string{}
			tinyMap = append(tinyMap, `╔═════╗`)
			// placeholder rows filled by the render loop below
			for range output.Render {
				tinyMap = append(tinyMap, ``)
			}
			tinyMap = append(tinyMap, `╚═════╝`)
			// This additional check is for ephemeral room copies,
			// which can slightly mess with the map render of the @
			if len(output.Render) > 2 && output.Render[2][2].Symbol != '@' {
				output.Render[2][2] = mapper.MapCell{Symbol: '@'}
			}

			legend := output.GetLegend(keywords.GetAllLegendAliases(room.Zone))

			for i, row := range output.Render {
				var sb strings.Builder
				for _, cell := range row {
					if cell.Symbol == ' ' {
						sb.WriteRune(' ')
						continue
					}

					if cell.FGColor > 0 || cell.BGColor > 0 {
						if cell.FGColor > 0 && cell.BGColor > 0 {
							fmt.Fprintf(&sb, `<ansi fg="%d" bg="%d">%c</ansi>`, cell.FGColor, cell.BGColor, cell.Symbol)
						} else if cell.BGColor > 0 {
							fmt.Fprintf(&sb, `<ansi fg="map-room" bg="%d">%c</ansi>`, cell.BGColor, cell.Symbol)
						} else {
							fmt.Fprintf(&sb, `<ansi fg="%d">%c</ansi>`, cell.FGColor, cell.Symbol)
						}
					} else if _, ok := legend[cell.Symbol]; ok {
						fmt.Fprintf(&sb, `<ansi fg="map-room">%c</ansi>`, cell.Symbol)
					} else {
						sb.WriteRune(cell.Symbol)
					}
				}
				tinyMap[i+1] = `║` + sb.String() + `║`
			}

			details = rooms.GetDetails(room, user, tinyMap)

		}

	} else {
		details = rooms.GetDetails(room, user)
	}

	textOut, _ := templates.Process("descriptions/room-title", details, excludeId(actor))
	tell(actor, textOut)

	tell(actor, buildRoomDescPanel(details))

	signCt := 0
	privateSigns := room.GetPrivateSigns()
	for _, sign := range privateSigns {
		if sign.VisibleUserId == excludeId(actor) {
			signCt++
			textOut, _ = templates.Process("descriptions/rune", sign, excludeId(actor))
			tell(actor, textOut)
		}
	}

	publicSigns := room.GetPublicSigns()
	for _, sign := range publicSigns {
		signCt++
		textOut, _ = templates.Process("descriptions/sign", sign, excludeId(actor))
		tell(actor, textOut)
	}

	if signCt > 0 {
		tell(actor, "")
	}

	textOut, _ = templates.Process("descriptions/who", details, excludeId(actor))
	if len(textOut) > 0 {
		tell(actor, textOut)
	}

	groundStuff := []string{}
	for containerName, container := range room.Containers {

		chestName := fmt.Sprintf(`<ansi fg="container">%s</ansi>`, containerName)

		if container.HasLock() {
			if container.Lock.IsLocked() {
				chestName += ` <ansi fg="white">(locked)</ansi>`
			} else {
				chestName += ` <ansi fg="white">(unlocked)</ansi>`
			}
		}

		groundStuff = append(groundStuff, chestName)

	}

	if room.Gold > 0 {
		groundStuff = append(groundStuff, fmt.Sprintf(`<ansi fg="gold">%d gold</ansi>`, room.Gold))
	}

	for _, item := range room.Items {
		if !item.IsValid() {
			room.RemoveItem(item, false)
			continue
		}
		groundStuff = append(groundStuff, item.DisplayName())
	}

	// Find stashed items
	for _, item := range room.Stash {
		if !item.IsValid() {
			room.RemoveItem(item, true)
		}
		if item.StashedBy != excludeId(actor) {
			continue
		}
		name := item.DisplayName() + ` <ansi fg="item-stashed">(stashed)</ansi>`
		groundStuff = append(groundStuff, name)
	}

	groundStuff = append(groundStuff, details.VisibleCorpses...)

	groundDetails := map[string]any{
		`GroundStuff`: groundStuff,
		`IsDark`:      room.GetBiome().IsDark(),
		`IsNight`:     gametime.IsNight(),
	}
	textOut, _ = templates.Process("descriptions/ontheground", groundDetails, excludeId(actor))
	if len(textOut) > 0 {
		tell(actor, textOut)
	}

	textOut, _ = templates.Process("descriptions/exits", details, excludeId(actor))
	tell(actor, textOut)

}
