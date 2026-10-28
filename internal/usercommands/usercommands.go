package usercommands

import (
	"fmt"
	"strings"
	"time"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/keywords"
	"github.com/GoMudEngine/GoMud/internal/messaging"
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

// Tell writes interface text to the actor through messaging: to its account
// when it has one, nowhere otherwise; empty text writes a blank line.
func Tell(actor *actors.Actor, text string) {
	messaging.Tell(actor.Id, text)
}

// UserIdOf is the actor's account UserId, or 0 for an NPC: the user ID
// templates and the mapper take.
func UserIdOf(actor *actors.Actor) int {
	if user := actor.Controller().Account(); user != nil {
		return user.UserId
	}
	return 0
}

// ExcludeId is the ID a room send excludes for actor: its account's UserId,
// or 0 for an NPC, which never receives room text (Decision 11).
func ExcludeId(actor *actors.Actor) int {
	if user := actor.Controller().Account(); user != nil {
		return user.UserId
	}
	return 0
}

// NameColour is the ANSI alias an actor's name is written in: `username` for
// an actor with an account, `mobname` otherwise.
func NameColour(actor *actors.Actor) string {
	if actor.Controller().Account() != nil {
		return "username"
	}
	return "mobname"
}

// permaGearRefused reports whether the actor's gear is stuck on it (the
// perma-gear buff flag). It then refuses: the actor is told, and the room sees
// the actor struggle with its gear (Decision 15).
func permaGearRefused(actor *actors.Actor) bool {
	if !actor.Character.HasBuffFlag("perma-gear") {
		return false
	}
	Tell(actor, `Your gear won't budge.`)
	// The @ prefix keeps Emote from echoing "You Emote: ..." back to the actor.
	queue(actor, `emote @struggles with their gear for a while, then gives up.`)
	return true
}

// queue queues input text for the actor's body, ready after the optional wait
// in seconds. It queues what Account.Command queues (Decision 11).
func queue(actor *actors.Actor, text string, wait ...float64) {
	readyTurn := util.GetTurnCount()
	if len(wait) > 0 {
		readyTurn += uint64(float64(configs.GetTimingConfig().SecondsToTurns(1)) * wait[0])
	}

	events.AddToQueue(events.Input{
		ActorId:   actor.Id,
		InputText: text,
		ReadyTurn: readyTurn,
	})
}

// queueFlagged is queue with input flags. It does what Account.CommandFlagged
// does: an account's input is blocked first when the flags ask for it.
func queueFlagged(actor *actors.Actor, text string, flagData events.EventFlag, wait ...float64) {
	readyTurn := util.GetTurnCount()
	if len(wait) > 0 {
		readyTurn += uint64(float64(configs.GetTimingConfig().SecondsToTurns(1)) * wait[0])
	}

	if flagData&events.CmdBlockInput == events.CmdBlockInput {
		if user := actor.Controller().Account(); user != nil {
			user.BlockInput()
		}
	}

	events.AddToQueue(events.Input{
		ActorId:   actor.Id,
		InputText: text,
		ReadyTurn: readyTurn,
		Flags:     flagData,
	})
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
		`alias`:       {Alias, true, false},
		`attack`:      {Attack, false, false},
		`badcommands`: {BadCommands, true, true}, // Admin only
		`biome`:       {Biome, true, false},
		`broadcast`:   {Broadcast, true, false},
		`build`:       {Build, false, true},        // Admin only
		`command`:     {AdminCommand, false, true}, // Admin only
		`copyover`:    {Copyover, true, true},      // Admin only
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

		// Kept until Phase 5: it is player and NPC death. NewRound_AutoHeal, NewRound_DoCombat
		// (players and mobs), NewRound_ActorUpkeep (runCharmExpiredCommand) and Buff_ApplyBuffs dispatch it.
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
				Tell(actor, "You are unable to do that while downed.")
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
		handled, err := Emote(cmd, actor, room, flags)
		return handled, err
	}

	// "go" attempt
	start := time.Now()
	defer func() {
		util.TrackTime(`usr-cmd[go]`, time.Since(start).Seconds())
	}()

	if handled, err := Go(cmd, actor, room, flags); handled {
		return handled, err
	}
	// end "go" attempt

	return false, nil
}

// IsCommand reports whether name is a registered command or emote, resolving
// the global keyword aliases the way TryCommand does (a multi-word alias
// counts by its first word). Personal aliases belong to accounts and are not
// resolved. world.go uses it to tell a command that declined an actor (it needs
// an account) from a word that is no command at all.
func IsCommand(name string) bool {
	if alias := keywords.TryCommandAlias(name); alias != name {
		name = strings.Split(alias, ` `)[0]
	}
	if _, ok := userCommands[name]; ok {
		return true
	}
	_, ok := emoteAliases[name]
	return ok
}

// RegisterCommand adds a command to the one registry from outside of the package.
func RegisterCommand(command string, handlerFunc Command, allowedWhenDowned bool, isAdminOnly bool) {
	userCommands[command] = CommandAccess{
		handlerFunc,
		allowedWhenDowned,
		isAdminOnly,
	}
}
