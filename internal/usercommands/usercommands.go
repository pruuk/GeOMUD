package usercommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

type CommandHelpItem struct {
	Command   string
	Type      string // command/skill
	Category  string
	AdminOnly bool
}

type CommandAccess struct {
	Func              Command
	AllowedWhenDowned bool
	AdminOnly         bool
}

// Command is one command any actor can run (Phase 4b slice 4): the rest of
// the input line, the actor, its room and the input flags.
type Command func(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error)

// Signature of a command not yet converted to the actor signature (removed
// when Task 4 ends).
type UserCommand func(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error)

// accountCommand adapts a command not yet converted to the actor signature:
// it runs only for an actor with an account. Removed when Task 4 ends.
func accountCommand(f UserCommand) Command {
	return func(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
		user := actor.Controller().Account()
		if user == nil {
			return false, nil
		}
		return f(rest, user, room, flags)
	}
}

// AdaptUserCommand is accountCommand for callers outside the package (the
// plugin registration path). Removed with accountCommand.
func AdaptUserCommand(f UserCommand) Command {
	return accountCommand(f)
}

// tell writes text to the actor: to its account when it has one, nowhere
// otherwise (Decision 11; slice 7 points this at messaging).
func tell(actor *actors.Actor, text string) {
	if user := actor.Controller().Account(); user != nil {
		user.SendText(text)
	}
}

// excludeId is the ID a room send excludes for actor: its account's UserId,
// or 0 for an NPC, which never receives room text (Decision 11).
func excludeId(actor *actors.Actor) int {
	if user := actor.Controller().Account(); user != nil {
		return user.UserId
	}
	return 0
}

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
		`alias`:       {accountCommand(Alias), true, false},
		`attack`:      {accountCommand(Attack), false, false},
		`badcommands`: {BadCommands, true, true}, // Admin only
		`biome`:       {accountCommand(Biome), true, false},
		`broadcast`:   {accountCommand(Broadcast), true, false},
		`build`:       {Build, false, true},        // Admin only
		`command`:     {AdminCommand, false, true}, // Admin only
		`copyover`:    {Copyover, true, true},      // Admin only
		`consider`:    {accountCommand(Consider), true, false},
		`default`:     {accountCommand(Default), false, false},
		`drop`:        {accountCommand(Drop), true, false},
		`emote`:       {accountCommand(Emote), true, false},
		`equip`:       {accountCommand(Equip), false, false},
		`flee`:        {accountCommand(Flee), false, false},
		`get`:         {accountCommand(Get), false, false},
		`give`:        {accountCommand(Give), false, false},
		`go`:          {accountCommand(Go), false, false},
		`grant`:       {Grant, true, true}, // Admin only
		`help`:        {accountCommand(Help), true, false},
		`keyring`:     {accountCommand(KeyRing), true, false},
		`history`:     {accountCommand(History), true, false},
		`inventory`:   {accountCommand(Inventory), true, false},
		`item`:        {Item, true, true},   // Admin only
		`locate`:      {Locate, true, true}, // Admin only
		`lock`:        {accountCommand(Lock), false, false},
		`look`:        {accountCommand(Look), true, false},
		`macros`:      {accountCommand(Macros), true, false},
		`mob`:         {Mob, true, true},    // Admin only
		`modify`:      {Modify, true, true}, // Admin only
		`motd`:        {accountCommand(Motd), true, false},
		`mute`:        {Mute, true, true},
		`noop`:        {accountCommand(Noop), true, false},
		`online`:      {accountCommand(Online), true, false},
		`party`:       {accountCommand(Party), true, false},
		`password`:    {accountCommand(Password), true, false},
		`prepare`:     {Prepare, true, true}, // Admin only
		`print`:       {accountCommand(Print), true, false},
		`printline`:   {accountCommand(PrintLine), true, false},
		`put`:         {accountCommand(Put), false, false},
		`quit`:        {accountCommand(Quit), true, false},
		`read`:        {accountCommand(Read), false, false},
		`reload`:      {Reload, true, true}, // Admin only
		`remove`:      {accountCommand(Remove), false, false},
		`rename`:      {Rename, false, true},     // Admin only
		`redescribe`:  {Redescribe, false, true}, // Admin only
		`room`:        {Room, false, true},       // Admin only
		`save`:        {accountCommand(Save), true, false},
		`say`:         {accountCommand(Say), true, false},
		`server`:      {Server, false, true}, // Admin only
		`set`:         {accountCommand(Set), true, false},
		`shoot`:       {accountCommand(Shoot), false, false},
		`shout`:       {accountCommand(Shout), true, false},
		`show`:        {accountCommand(Show), true, false},
		`spawn`:       {Spawn, false, true}, // Admin only

		// Kept until Phase 5: NewRound_AutoHeal and NewRound_DoCombat dispatch it as player death.
		`suicide`:  {accountCommand(Suicide), true, false},
		`syslogs`:  {SysLogs, true, true},  // Admin only
		`teleport`: {Teleport, true, true}, // Admin only
		`unlock`:   {accountCommand(Unlock), false, false},
		`unmute`:   {UnMute, true, true}, // Admin only
		`whisper`:  {accountCommand(Whisper), true, false},
		`who`:      {accountCommand(Who), true, false},
		`visit`:    {Visit, true, true}, // Admin only
		`visited`:  {accountCommand(Visited), true, false},
		`zone`:     {Zone, false, true}, // Admin only
		// Special command only used upon creating a new account
		`start`: {accountCommand(Start), false, false},
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

func TryCommand(cmd string, rest string, actorId ids.ID, flags events.EventFlag) (bool, error) {

	userDisabled := false
	actor := actors.Get(actorId)
	if actor == nil || actor.Character == nil {
		return false, fmt.Errorf(`actor %s not found`, actorId)
	}
	user := actor.Controller().Account() // nil for an NPC

	room := rooms.LoadRoom(actor.Character.RoomId)
	if room == nil {
		return false, fmt.Errorf(`room %d not found`, actor.Character.RoomId)
	}

	rest = strings.TrimSpace(rest)

	// Figure out whether it was an exit entered
	exitName, _ := room.FindExitByName(cmd)
	if exitName != `` {
		rest = cmd
		cmd = "go"
	} else {

		personalAlias := cmd
		if user != nil {
			personalAlias = user.TryCommandAlias(cmd)
		}
		if alias := personalAlias; alias != cmd {
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

		userDisabled = actor.Character.IsDisabled()

	}

	// Cancel any buffs they have that get cancelled based on them doing anything at all
	actor.Character.CancelBuffsWithFlag("cancel-on-action")

	// Experimental, not sure if will have unexpected consequences.
	// Turn keywords for targetting self into actual string of self
	if cmd == `look` {
		for _, selfWord := range selfKeywords {
			wordLen := len(selfWord)
			if rest == selfWord {
				rest = actor.Character.Name
				break
			} else if len(rest) >= wordLen+1 && rest[len(rest)-(wordLen+1):] == ` `+selfWord {
				rest = rest[:len(rest)-(wordLen+1)] + ` ` + actor.Character.Name
				break
			}
		}
	}

	if user != nil && user.HasPlaintextPassword() && cmd != `password` {
		user.SendText(`<ansi fg="alert-5">You must change your password before doing anything else.</ansi>`)
		user.SendText(`<ansi fg="yellow">Type <ansi fg="yellow-bold">password</ansi> to set a new password.</ansi>`)
		return true, nil
	}

	if cmdInfo, ok := userCommands[cmd]; ok {

		if !cmdInfo.AllowedWhenDowned {

			// If actually downed, prevent it (unless admin)
			if userDisabled && !cmdInfo.AdminOnly {
				tell(actor, "You are unable to do that while downed.")
				return true, nil
			}

			// Disabled input affects commands which can't be performed when downed.
			if user != nil && user.InputBlocked() {
				return true, nil
			}
		}

		// An actor without an account holds no role, so it is never an admin.
		if !cmdInfo.AdminOnly || (user != nil && user.HasRolePermission(cmd, true)) {

			start := time.Now()
			defer func() {
				util.TrackTime(`usr-cmd[`+cmd+`]`, time.Since(start).Seconds())
			}()

			if cmdInfo.AdminOnly {
				mudlog.Info("Admin Command", "cmd", cmd, "rest", rest, "userId", user.UserId)
			}

			// Run the command here
			handled, err := cmdInfo.Func(rest, actor, room, flags)
			return handled, err

		} else if cmdInfo.AdminOnly {
			if user == nil {
				// Refused silently: an NPC is never told anything about admin commands.
				return true, nil
			}
			user.SendText(fmt.Sprintf(`You don't have permission to use <ansi fg="command">%s</ansi>.`, cmd))
			return true, nil
		}
	}

	if _, ok := emoteAliases[cmd]; ok {
		handled, err := accountCommand(Emote)(cmd, actor, room, flags)
		return handled, err
	}

	// "go" attempt
	start := time.Now()
	defer func() {
		util.TrackTime(`usr-cmd[go]`, time.Since(start).Seconds())
	}()

	if handled, err := accountCommand(Go)(cmd, actor, room, flags); handled {
		return handled, err
	}
	// end "go" attempt

	return false, nil
}

// Register mob commands from outside of the package
func RegisterCommand(command string, handlerFunc Command, disabledWhenDowned bool, isAdminOnly bool) {
	userCommands[command] = CommandAccess{
		handlerFunc,
		disabledWhenDowned,
		isAdminOnly,
	}
}
