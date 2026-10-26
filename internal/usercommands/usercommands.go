package usercommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

type CommandHelpItem struct {
	Command   string
	Type      string // command/skill
	Category  string
	AdminOnly bool
}

type CommandAccess struct {
	Func              UserCommand
	AllowedWhenDowned bool
	AdminOnly         bool
}

// Signature of user command
type UserCommand func(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error)

type FunctionExporter interface {
	GetExportedFunction(funcName string) (any, bool)
}

func AddFunctionExporter(f FunctionExporter) {
	functionExporters = append(functionExporters, f)
}

// roomTagProvider is set by main at startup via SetRoomTagProvider.
// It returns a map of plugin name to the room tags that plugin has reserved.
var roomTagProvider func() map[string][]string = func() map[string][]string { return nil }

// SetRoomTagProvider registers the function that returns registered room tags.
func SetRoomTagProvider(f func() map[string][]string) {
	roomTagProvider = f
}

var (
	functionExporters = []FunctionExporter{}

	userCommands map[string]CommandAccess = map[string]CommandAccess{
		`alias`:       {Alias, true, false},
		`attack`:      {Attack, false, false},
		`badcommands`: {BadCommands, true, true}, // Admin only
		`biome`:       {Biome, true, false},
		`broadcast`:   {Broadcast, true, false},
		`build`:       {Build, false, true},   // Admin only
		`command`:     {Command, false, true}, // Admin only
		`copyover`:    {Copyover, true, true}, // Admin only
		`consider`:    {Consider, true, false},
		`default`:     {Default, false, false},
		`drop`:        {Drop, true, false},
		`emote`:       {Emote, true, false},
		`equip`:       {Equip, false, false},
		`flee`:        {Flee, false, false},
		`get`:         {Get, false, false},
		`give`:        {Give, false, false},
		`go`:          {Go, false, false},
		`grant`:       {Grant, true, true}, // Admin only
		`help`:        {Help, true, false},
		`keyring`:     {KeyRing, true, false},
		`history`:     {History, true, false},
		`inventory`:   {Inventory, true, false},
		`item`:        {Item, true, true},   // Admin only
		`locate`:      {Locate, true, true}, // Admin only
		`lock`:        {Lock, false, false},
		`look`:        {Look, true, false},
		`macros`:      {Macros, true, false},
		`mob`:         {Mob, true, true},    // Admin only
		`modify`:      {Modify, true, true}, // Admin only
		`motd`:        {Motd, true, false},
		`mute`:        {Mute, true, true},
		`noop`:        {Noop, true, false},
		`online`:      {Online, true, false},
		`party`:       {Party, true, false},
		`password`:    {Password, true, false},
		`prepare`:     {Prepare, true, true}, // Admin only
		`print`:       {Print, true, false},
		`printline`:   {PrintLine, true, false},
		`put`:         {Put, false, false},
		`quit`:        {Quit, true, false},
		`read`:        {Read, false, false},
		`reload`:      {Reload, true, true}, // Admin only
		`remove`:      {Remove, false, false},
		`rename`:      {Rename, false, true},     // Admin only
		`redescribe`:  {Redescribe, false, true}, // Admin only
		`room`:        {Room, false, true},       // Admin only
		`save`:        {Save, true, false},
		`say`:         {Say, true, false},
		`server`:      {Server, false, true}, // Admin only
		`set`:         {Set, true, false},
		`shoot`:       {Shoot, false, false},
		`shout`:       {Shout, true, false},
		`show`:        {Show, true, false},
		`spawn`:       {Spawn, false, true}, // Admin only

		// Kept until Phase 5: NewRound_AutoHeal and NewRound_DoCombat dispatch it as player death.
		`suicide`:  {Suicide, true, false},
		`syslogs`:  {SysLogs, true, true},  // Admin only
		`teleport`: {Teleport, true, true}, // Admin only
		`unlock`:   {Unlock, false, false},
		`unmute`:   {UnMute, true, true}, // Admin only
		`whisper`:  {Whisper, true, false},
		`who`:      {Who, true, false},
		`visit`:    {Visit, true, true}, // Admin only
		`visited`:  {Visited, true, false},
		`zone`:     {Zone, false, true}, // Admin only
		// Special command only used upon creating a new account
		`start`: {Start, false, false},
	}

	selfKeywords = []string{
		`me`,
		`self`,
		`myself`,
	}
)

func GetExportedFunction(fName string) (any, bool) {
	for _, x := range functionExporters {
		if f, ok := x.GetExportedFunction(fName); ok {
			return f, ok
		}
	}
	return nil, false
}

func TryCommand(cmd string, rest string, userId int, flags events.EventFlag) (bool, error) {

	userDisabled := false
	user := users.GetByUserId(userId)
	if user == nil {
		return false, fmt.Errorf(`user %d not found`, userId)
	}

	room := rooms.LoadRoom(user.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, user.Character.RoomId)
	}

	rest = strings.TrimSpace(rest)

	// Figure out whether it was an exit entered
	exitName, _ := room.FindExitByName(cmd)
	if exitName != `` {
		rest = cmd
		cmd = "go"
	} else {

		if alias := user.TryCommandAlias(cmd); alias != cmd {
			// If it's a multi-word aliase, we need to extract the first word to replace the command
			// The rest will be combined with any "rest" the player provided.
			if strings.Contains(alias, ` `) {
				parts := strings.Split(alias, ` `)
				// grab the first word as the new cmd
				cmd = parts[0]
				// Add the "rest" to the end if any
				if len(rest) > 0 {
					rest = strings.TrimPrefix(alias, cmd+` `) + ` ` + rest
				} else {
					rest = strings.TrimPrefix(alias, cmd+` `)
				}
			} else {
				cmd = alias
			}
		}

		if alias := keywords.TryCommandAlias(cmd); alias != cmd {
			// If it's a multi-word aliase, we need to extract the first word to replace the command
			// The rest will be combined with any "rest" the player provided.
			if strings.Contains(alias, ` `) {
				parts := strings.Split(alias, ` `)
				// grab the first word as the new cmd
				cmd = parts[0]
				// Add the "rest" to the end if any
				if len(rest) > 0 {
					rest = strings.TrimPrefix(alias, cmd+` `) + ` ` + rest
				} else {
					rest = strings.TrimPrefix(alias, cmd+` `)
				}
			} else {
				cmd = alias
			}
		}

		userDisabled = user.Character.IsDisabled()

	}

	// Cancel any buffs they have that get cancelled based on them doing anything at all
	user.Character.CancelBuffsWithFlag("cancel-on-action")

	// Experimental, not sure if will have unexpected consequences.
	// Turn keywords for targetting self into actual string of self
	if cmd == `look` {
		for _, selfWord := range selfKeywords {
			wordLen := len(selfWord)
			if rest == selfWord {
				rest = user.Character.Name
				break
			} else if len(rest) >= wordLen+1 && rest[len(rest)-(wordLen+1):] == ` `+selfWord {
				rest = rest[:len(rest)-(wordLen+1)] + ` ` + user.Character.Name
				break
			}
		}
	}

	if user.HasPlaintextPassword() && cmd != `password` {
		user.SendText(`<ansi fg="alert-5">You must change your password before doing anything else.</ansi>`)
		user.SendText(`<ansi fg="yellow">Type <ansi fg="yellow-bold">password</ansi> to set a new password.</ansi>`)
		return true, nil
	}

	if cmdInfo, ok := userCommands[cmd]; ok {

		if !cmdInfo.AllowedWhenDowned {

			// If actually downed, prevent it (unless admin)
			if userDisabled && !cmdInfo.AdminOnly {
				user.SendText("You are unable to do that while downed.")
				return true, nil
			}

			// Disabled input affects commands which can't be performed when downed.
			if user.InputBlocked() {
				return true, nil
			}
		}

		if !cmdInfo.AdminOnly || user.HasRolePermission(cmd, true) {

			start := time.Now()
			defer func() {
				util.TrackTime(`usr-cmd[`+cmd+`]`, time.Since(start).Seconds())
			}()

			if cmdInfo.AdminOnly {
				mudlog.Info("Admin Command", "cmd", cmd, "rest", rest, "userId", user.UserId)
			}

			// Run the command here
			handled, err := cmdInfo.Func(rest, user, room, flags)
			return handled, err

		} else if cmdInfo.AdminOnly {
			user.SendText(fmt.Sprintf(`You don't have permission to use <ansi fg="command">%s</ansi>.`, cmd))
			return true, nil
		}
	}

	if _, ok := emoteAliases[cmd]; ok {
		handled, err := Emote(cmd, user, room, flags)
		return handled, err
	}

	// "go" attempt
	start := time.Now()
	defer func() {
		util.TrackTime(`usr-cmd[go]`, time.Since(start).Seconds())
	}()

	if handled, err := Go(cmd, user, room, flags); handled {
		return handled, err
	}
	// end "go" attempt

	return false, nil
}

// Register mob commands from outside of the package
func RegisterCommand(command string, handlerFunc UserCommand, disabledWhenDowned bool, isAdminOnly bool) {
	userCommands[command] = CommandAccess{
		handlerFunc,
		disabledWhenDowned,
		isAdminOnly,
	}
}
