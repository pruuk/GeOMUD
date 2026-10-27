package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"

	"github.com/GoMudEngine/GoMud/internal/accounts"
)

/*
* Role Permissions:
* modify 				(All)
* modify.role			(Change user roles)
 */
func Modify(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error) {

	args := util.SplitButRespectQuotes(rest)

	if len(args) < 3 {
		infoOutput, _ := templates.Process("admincommands/help/command.modify", nil, user.UserId)
		user.SendText(infoOutput)
		return true, nil
	}

	// modify role <username> <newrole>
	if args[0] == `role` {

		if !user.HasRolePermission(`modify.role`) {
			user.SendText(`you do not have <ansi fg="command">modify.role</ansi> permission`)
			return true, nil
		}

		searchUser := args[1]
		newRole := args[2]

		if newRole != accounts.RoleAdmin && newRole != accounts.RoleUser && newRole != accounts.RoleMod {
			user.SendText(`<ansi fg="alert-4">Invalid role. Valid roles: user, mod, admin</ansi>`)
			return true, nil
		}

		foundUsername := ``
		foundCharacterName := ``

		for _, u := range accounts.GetAllActiveUsers() {
			if strings.EqualFold(searchUser, u.Username) {

				if u.Role == accounts.RoleAdmin {
					user.SendText(`<ansi fg="alert-4">Admin role cannot be removed this way.</ansi>`)
					return true, nil
				}

				if u.Role == newRole {
					user.SendText(`<ansi fg="alert-4">That role is already set for this user.</ansi>`)
					return true, nil
				}

				foundCharacterName = u.Character.Name
				foundUsername = u.Username

				u.Role = newRole

				accounts.SaveUser(*u)

				u.SendText(`<ansi fg="alert-3">Your role has been set to: ` + newRole + `</ansi>`)
				break
			}
		}

		if len(foundUsername) == 0 {
			accounts.SearchOfflineUsers(func(u *accounts.Account) bool {

				if strings.EqualFold(searchUser, u.Username) {

					if u.Role == accounts.RoleAdmin {
						user.SendText(`<ansi fg="alert-4">Admin role cannot be removed this way.</ansi>`)
						return false
					}

					if u.Role == newRole {
						user.SendText(`<ansi fg="alert-4">That role is already set for this user.</ansi>`)
						return false
					}

					foundCharacterName = u.Character.Name
					foundUsername = u.Username

					u.Role = newRole

					accounts.SaveUser(*u)

					return false
				}

				return true
			})
		}

		if len(foundUsername) > 0 {
			user.SendText(fmt.Sprintf(`Role changed for user <ansi fg="username">%s</ansi> (Character name: <ansi fg="username">%s</ansi>).`, foundUsername, foundCharacterName))
			return true, nil
		}

	}

	return true, nil
}
