package usercommands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"
)

/*
* Role Permissions:
* visit 				(All)
 */
func Visit(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) == 0 {
		infoOutput, _ := templates.Process("admincommands/help/command.visit", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	subCmd := strings.ToLower(args[0])
	args = args[1:]

	switch subCmd {
	case "list":
		return visit_List(args, user)
	case "set":
		return visit_SetUnset(args, user, true)
	case "unset":
		return visit_SetUnset(args, user, false)
	default:
		infoOutput, _ := templates.Process("admincommands/help/command.visit", nil, user.UserId)
		user.SendText(infoOutput)
	}

	return true, nil
}

func visit_resolveTarget(args []string, invoker *accounts.Account) (*accounts.Account, bool, error) {
	if len(args) == 0 {
		return invoker, false, nil
	}

	username := args[0]

	if online := accounts.GetByCharacterName(username); online != nil {
		return online, false, nil
	}

	return nil, false, fmt.Errorf(`user "%s" not found`, username)
}

func visit_List(args []string, user *accounts.Account) (bool, error) {

	targetUser, isOffline, err := visit_resolveTarget(args, user)
	if err != nil {
		user.SendText(err.Error())
		return true, nil
	}

	allZoneNames := rooms.GetAllZoneNames()
	sort.Strings(allZoneNames)

	headers := []string{"Zone", "Visited", "Unvisited", "% Complete"}
	rows := [][]string{}

	for _, zoneName := range allZoneNames {
		zCfg := rooms.GetZoneConfig(zoneName)
		if zCfg == nil {
			continue
		}

		visited, total := targetUser.Character.ZoneVisitProgress(zoneName, zCfg.RoomIds)
		unvisited := total - visited

		pct := 0
		if total > 0 {
			pct = (visited * 100) / total
		}

		rows = append(rows, []string{
			zoneName,
			fmt.Sprintf(`%d`, visited),
			fmt.Sprintf(`%d`, unvisited),
			fmt.Sprintf(`%d%%`, pct),
		})
	}

	title := fmt.Sprintf(`Visit Progress for %s`, targetUser.Character.Name)
	if isOffline {
		title += ` (offline)`
	}

	tableData := templates.GetTable(title, headers, rows)
	tplTxt, _ := templates.Process("tables/generic", tableData, user.UserId)
	user.SendText(tplTxt)

	return true, nil
}

func visit_SetUnset(args []string, user *accounts.Account, markVisited bool) (bool, error) {

	if len(args) == 0 {
		infoOutput, _ := templates.Process("admincommands/help/command.visit", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	zonePart := args[0]
	userArgs := args[1:]

	targetUser, isOffline, err := visit_resolveTarget(userArgs, user)
	if err != nil {
		user.SendText(err.Error())
		return true, nil
	}

	action := "marked visited"
	if !markVisited {
		action = "reset to unvisited"
	}

	if strings.ToLower(zonePart) == "all" {
		allZoneNames := rooms.GetAllZoneNames()
		for _, zoneName := range allZoneNames {
			visit_applyZone(targetUser, zoneName, markVisited)
		}

		if isOffline {
			accounts.SaveUser(*targetUser)
		}

		user.SendText(fmt.Sprintf(`All zones %s for <ansi fg="username">%s</ansi>.`, action, targetUser.Character.Name))
		return true, nil
	}

	allZoneNames := rooms.GetAllZoneNames()
	exactZone, closeZone := util.FindMatchIn(zonePart, allZoneNames...)
	resolvedZone := exactZone
	if resolvedZone == `` {
		resolvedZone = closeZone
	}
	if resolvedZone == `` {
		user.SendText(fmt.Sprintf(`Zone "%s" not found.`, zonePart))
		return true, nil
	}

	visit_applyZone(targetUser, resolvedZone, markVisited)

	if isOffline {
		accounts.SaveUser(*targetUser)
	}

	user.SendText(fmt.Sprintf(`Zone <ansi fg="yellow">%s</ansi> %s for <ansi fg="username">%s</ansi>.`, resolvedZone, action, targetUser.Character.Name))

	return true, nil
}

func visit_applyZone(targetUser *accounts.Account, zoneName string, markVisited bool) {
	zCfg := rooms.GetZoneConfig(zoneName)
	if zCfg == nil {
		return
	}

	if markVisited {
		for roomId := range zCfg.RoomIds {
			targetUser.Character.MarkVisitedRoom(roomId, zoneName, nil)
		}
	} else {
		if targetUser.Character.ZonesVisited != nil {
			delete(targetUser.Character.ZonesVisited, zoneName)
		}
	}
}
