package usercommands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/perception"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/util"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
)

/*
* Role Permissions:
* item 				(All)
* item.create		(Create a new item)
* item.spawn		(Spawn a new item in the room)
 */
func Item(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 1 {
		infoOutput, _ := templates.Process("admincommands/help/command.item", nil, user.UserId)
		Tell(actor, infoOutput)
		return true, nil
	}

	// create a new item
	if args[0] == `create` {

		if !user.HasRolePermission(`item.create`) {
			Tell(actor, `you do not have <ansi fg="command">item.create</ansi> permission`)
			return true, nil
		}

		return item_Create(strings.TrimSpace(rest[6:]), user, room, flags)
	}

	// spawn an existing item
	if args[0] == `spawn` {

		if !user.HasRolePermission(`item.spawn`) {
			Tell(actor, `you do not have <ansi fg="command">item.create</ansi> permission`)
			return true, nil
		}

		return item_Spawn(strings.TrimSpace(rest[5:]), user, room, flags)
	}

	// List existing items
	if args[0] == `list` {

		if !user.HasRolePermission(`item.spawn`) {
			Tell(actor, `you do not have <ansi fg="command">item.create</ansi> permission`)
			return true, nil
		}

		return item_List(strings.TrimSpace(rest[4:]), user, room, flags)
	}

	return true, nil
}

func item_List(rest string, user *accounts.Account, _ *rooms.Room, _ events.EventFlag) (bool, error) {

	itmNames := []templates.NameDescription{}
	longestName := 0

	for _, itm := range items.GetAllItemSpecs() {

		entry := templates.NameDescription{
			Id:   itm.ItemId,
			Name: itm.Name,
		}

		// If searching for matches
		if len(rest) > 0 {
			if !strings.Contains(rest, `*`) {
				rest += `*`
			}
			if !util.StringWildcardMatch(strings.ToLower(entry.Name), rest) {
				continue
			}
		}

		if len(entry.Name) > longestName {
			longestName = len(entry.Name)
		}

		itmNames = append(itmNames, entry)
	}

	sort.SliceStable(itmNames, func(i, j int) bool {
		return strings.ToLower(itmNames[i].Name) < strings.ToLower(itmNames[j].Name)
	})

	numWidth := len(strconv.Itoa(len(itmNames)))
	colWidth := 1 + numWidth + 2 + longestName + 1

	messaging.ToAccount(user, ``)
	sw := user.ClientSettings().Display.GetScreenWidth()
	strOut := templates.DynamicList(itmNames, colWidth, sw, numWidth, longestName)
	messaging.ToAccount(user, strOut)
	messaging.ToAccount(user, ``)

	return true, nil
}

func item_Spawn(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	itemId := items.FindItem(rest)
	if itemId != 0 {

		itm := items.New(itemId)
		if itm.ItemId > 0 {
			room.AddItem(itm, false)

			messaging.ToAccount(user,
				fmt.Sprintf(`You wave your hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, itm.DisplayName()),
			)
			if spawner := actors.ByUserId(user.UserId); spawner != nil {
				messaging.Send(perception.Message{
					Kind:         perception.WorldEvent,
					Actor:        spawner.Id,
					Senses:       perception.Sight | perception.Sound,
					Range:        perception.Room,
					ObserverText: fmt.Sprintf(`{actor} waves their hands around and <ansi fg="item">%s</ansi> appears from thin air and falls to the ground.`, itm.DisplayName()),
				})
			}

			return true, nil
		}

	}

	messaging.ToAccount(user,
		fmt.Sprintf(`Item <ansi fg="itemname">%s</ansi> not found.`, rest),
	)

	return true, nil
}

func item_Create(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	var newItemSpec = items.ItemSpec{}

	if len(rest) > 0 {
		if itemId := items.FindItem(rest); itemId > 0 {
			if spec := items.GetItemSpec(itemId); spec != nil {
				newItemSpec = *spec
			}
		}
	}

	// Get if already exists, otherwise create new
	cmdPrompt, isNew := user.StartPrompt(`item create`, rest)

	if isNew {
		messaging.ToAccount(user, ``)
		messaging.ToAccount(user, fmt.Sprintf(`Lets get a little info first.%s`, term.CRLFStr))
	}

	//
	// Name Selection
	//
	{

		question := cmdPrompt.Ask(`What will the item be called?`, []string{newItemSpec.Name}, newItemSpec.Name)
		if !question.Done {
			return true, nil
		}

		newItemSpec.Name = question.Response

	}

	//
	// Type Selection
	//
	{
		allTypes := items.ItemTypes()

		typeOptions := []templates.NameDescription{}
		for _, t := range allTypes {
			typeOptions = append(typeOptions, templates.NameDescription{
				Name:        t.Type,
				Description: t.Description,
			})
		}

		question := cmdPrompt.Ask(`What Type of item will it be?`, []string{string(newItemSpec.Type)}, string(newItemSpec.Type))
		if !question.Done {
			tplTxt, _ := templates.Process("tables/numbered-list", typeOptions, user.UserId)
			messaging.ToAccount(user, tplTxt)
			return true, nil
		}

		typeNameSelection := question.Response
		if restNum, err := strconv.Atoi(typeNameSelection); err == nil {
			if restNum > 0 && restNum <= len(typeOptions) {
				typeNameSelection = typeOptions[restNum-1].Name
			}
		}

		for _, t := range allTypes {
			if strings.EqualFold(t.Type, typeNameSelection) {
				newItemSpec.Type = items.ItemType(t.Type)
			}
		}

		if newItemSpec.Type == `` {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list", typeOptions, user.UserId)
			messaging.ToAccount(user, tplTxt)

			return true, nil
		}
	}

	//
	// Damage (if weapon)
	//
	if newItemSpec.Type == items.Weapon {

		question := cmdPrompt.Ask(`What damage does this weapon do (Example: 1d4)?`, []string{newItemSpec.Damage.DiceRoll}, newItemSpec.Damage.DiceRoll)
		if !question.Done {
			return true, nil
		}

		if question.Response != `` {
			newItemSpec.Damage.InitDiceRoll(question.Response)
			newItemSpec.Damage.DiceRoll = newItemSpec.Damage.FormatDiceRoll()
		}

	}

	//
	// Target room/exit/container (If key)
	//
	if newItemSpec.Type == items.Key {

		roomIdStr := ``
		roomExitStr := ``

		keyParts := strings.Split(newItemSpec.KeyLockId, `-`)
		if len(keyParts) == 2 {
			roomIdStr = keyParts[0]
			roomExitStr = keyParts[1]
		}

		question := cmdPrompt.Ask(`What Room Id will this key be used in?`, []string{roomIdStr}, roomIdStr)
		if !question.Done {
			return true, nil
		}

		if question.Response == `` {
			messaging.ToAccount(user, "Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		roomId, _ := strconv.Atoi(question.Response)
		if roomId == 0 {
			question.RejectResponse()
			return true, nil
		}

		question = cmdPrompt.Ask(`What exit name or container will this open?`, []string{roomExitStr}, roomExitStr)
		if !question.Done {
			return true, nil
		}

		if question.Response == `` {
			messaging.ToAccount(user, "Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		newItemSpec.KeyLockId = fmt.Sprintf(`%d-%s`, roomId, strings.ToLower(question.Response))
	}

	//
	// SubType Selection
	//
	{
		allSubTypes := items.ItemSubtypes()

		subTypeOptions := []templates.NameDescription{}
		for _, t := range allSubTypes {
			subTypeOptions = append(subTypeOptions, templates.NameDescription{
				Name:        t.Type,
				Description: t.Description,
			})
		}

		question := cmdPrompt.Ask(`What Subtype of item will it be?`, []string{string(newItemSpec.Subtype)}, string(newItemSpec.Subtype))
		if !question.Done {
			tplTxt, _ := templates.Process("tables/numbered-list", subTypeOptions, user.UserId)
			messaging.ToAccount(user, tplTxt)
			return true, nil
		}

		if question.Response == `` {
			messaging.ToAccount(user, "Aborting...")
			user.ClearPrompt()
			return true, nil
		}

		typeNameSelection := question.Response
		if restNum, err := strconv.Atoi(typeNameSelection); err == nil {
			if restNum > 0 && restNum <= len(subTypeOptions) {
				typeNameSelection = subTypeOptions[restNum-1].Name
			}
		}

		for _, t := range allSubTypes {
			if strings.EqualFold(t.Type, typeNameSelection) {
				newItemSpec.Subtype = items.ItemSubType(t.Type)
			}
		}

		if newItemSpec.Subtype == `` {
			question.RejectResponse()

			tplTxt, _ := templates.Process("tables/numbered-list", subTypeOptions, user.UserId)
			messaging.ToAccount(user, tplTxt)

			return true, nil
		}
	}

	//
	// Maximum Uses
	//
	{
		defaultYN := `n`
		if newItemSpec.Uses > 0 {
			defaultYN = `y`
		}
		question := cmdPrompt.Ask(`Will this item have a maximum number of uses?`, []string{`y`, `n`}, defaultYN)
		if !question.Done {
			return true, nil
		}

		if question.Response == `y` {
			question := cmdPrompt.Ask(`How many uses will this item have?`, []string{strconv.Itoa(newItemSpec.Uses)}, strconv.Itoa(newItemSpec.Uses))
			if !question.Done {
				return true, nil
			}

			newItemSpec.Uses, _ = strconv.Atoi(question.Response)
		}

	}

	//
	// Name Selection
	//
	{
		question := cmdPrompt.Ask(`Gold value override:`, []string{strconv.Itoa(newItemSpec.Value)}, strconv.Itoa(newItemSpec.Value))
		if !question.Done {
			return true, nil
		}
		newItemSpec.Value, _ = strconv.Atoi(question.Response)
	}

	//
	// Description
	//
	{
		question := cmdPrompt.Ask(`Enter a description for the item:`, []string{newItemSpec.Description}, newItemSpec.Description)
		if !question.Done {
			return true, nil
		}

		newItemSpec.Description = question.Response
	}

	//
	// Confirm?
	//
	{
		question := cmdPrompt.Ask(`Does this look correct?`, []string{`y`, `n`}, `n`)
		if !question.Done {

			messaging.ToAccount(user, `  <ansi fg="yellow-bold">Name:</ansi>        <ansi fg="white-bold">`+newItemSpec.Name+`</ansi>`)
			messaging.ToAccount(user, `  <ansi fg="yellow-bold">Desc:</ansi>        <ansi fg="white-bold">`+newItemSpec.Description+`</ansi>`)
			messaging.ToAccount(user, `  <ansi fg="yellow-bold">Type:</ansi>        <ansi fg="white-bold">`+string(newItemSpec.Type)+`</ansi>`)

			if newItemSpec.Type == items.Key {
				messaging.ToAccount(user, `  <ansi fg="yellow-bold">KeyId:</ansi>       <ansi fg="white-bold">`+newItemSpec.KeyLockId+`</ansi>`)
			}
			if newItemSpec.Type == items.Weapon {
				messaging.ToAccount(user, `  <ansi fg="yellow-bold">Damage:</ansi>      <ansi fg="white-bold">`+newItemSpec.Damage.DiceRoll+`</ansi>`)
			}

			if newItemSpec.Uses > 0 {
				messaging.ToAccount(user, `  <ansi fg="yellow-bold">Uses:</ansi>        <ansi fg="white-bold">`+strconv.Itoa(newItemSpec.Uses)+`</ansi>`)
			}

			if newItemSpec.Value > 0 {
				messaging.ToAccount(user, `  <ansi fg="yellow-bold">Value:</ansi>       <ansi fg="white-bold">`+strconv.Itoa(newItemSpec.Value)+`</ansi>`)
			}

			messaging.ToAccount(user, `  <ansi fg="yellow-bold">SubType:</ansi>     <ansi fg="white-bold">`+string(newItemSpec.Subtype)+`</ansi>`)

			return true, nil
		}

		user.ClearPrompt()

		if question.Response != `y` {
			messaging.ToAccount(user, "Aborting...")
			return true, nil
		}
	}

	newItemId, err := items.CreateNewItemFile(newItemSpec)

	if err != nil {
		messaging.ToAccount(user, "Error: "+err.Error())
		return true, nil
	}

	itemInst := items.GetItemSpec(newItemId)
	if itemInst == nil {
		messaging.ToAccount(user, "Error: item spec not found after creation.")
		return true, nil
	}

	messaging.ToAccount(user, ``)
	messaging.ToAccount(user, `  <ansi bg="red" fg="white-bold">ITEM CREATED</ansi>`)
	messaging.ToAccount(user, ``)
	messaging.ToAccount(user, `  <ansi fg="yellow-bold">File Path:</ansi>   <ansi fg="white-bold">`+itemInst.Filepath()+`</ansi>`)
	messaging.ToAccount(user, ``)
	messaging.ToAccount(user, `  <ansi fg="black-bold">note: Try <ansi fg="command">item spawn `+newItemSpec.Name+`</ansi> to test it.`)

	return true, nil
}
