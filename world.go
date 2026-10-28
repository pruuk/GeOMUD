package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/badinputtracker"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/prompt"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/GoMudEngine/GoMud/internal/web"
)

type WorldInput struct {
	FromId    int
	InputText string
}

func (wi WorldInput) Id() int {
	return wi.FromId
}

type World struct {
	worldInput         chan WorldInput
	ignoreInput        map[ids.ID]uint64 // actor ID -> turn its input was blocked
	enterWorldUserId   chan [2]int
	leaveWorldUserId   chan int
	logoutConnectionId chan connections.ConnectionId
	linkDeadFlag       chan [2]int
	//
	// inputTracker holds the actors that ran (or, without an account, hold)
	// a queued command in this event-loop pass. EventLoop clears it.
	inputTracker map[ids.ID]struct{}
}

func NewWorld(osSignalChan chan os.Signal) *World {

	w := &World{
		worldInput:         make(chan WorldInput),
		ignoreInput:        make(map[ids.ID]uint64),
		enterWorldUserId:   make(chan [2]int),
		leaveWorldUserId:   make(chan int),
		logoutConnectionId: make(chan connections.ConnectionId),
		linkDeadFlag:       make(chan [2]int),
		//
		inputTracker: map[ids.ID]struct{}{},
	}

	// System commands
	events.RegisterListener(events.System{}, w.HandleSystemEvents)
	events.RegisterListener(events.Input{}, w.HandleInputEvents)

	connections.SetShutdownChan(osSignalChan)

	return w
}

func (w *World) HandleInputEvents(e events.Event) events.ListenerReturn {

	input, typeOk := e.(events.Input)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "Input", "Actual Type", e.Type())
		return events.Continue
	}

	actor := actors.Get(input.ActorId)
	if actor == nil {
		// A body with no live actor: a player who logged out, an NPC that
		// died, a rerolled body (Decision 14).
		mudlog.Debug("input", "event", "dropped", "actorId", input.ActorId.String(), "reason", "no live actor")
		return events.Continue
	}
	turnCt := util.GetTurnCount()

	// Decision 13: an actor without an account runs its commands strictly in
	// order, so a delayed command holds back the ones after it; an actor with
	// one keeps the player rules. The limit is one queued command per actor
	// per event-loop pass, not per turn.
	user := actor.Controller().Account()
	if user == nil {
		// Checked first, so a pending delayed command holds back even an
		// immediate one.
		if _, ok := w.inputTracker[actor.Id]; ok {
			return events.CancelAndRequeue
		}
		// Ready turn 0 runs at once and does not count towards the limit.
		if input.ReadyTurn == 0 {
			w.processInput(actor, input.InputText, input.Flags)
			return events.Continue
		}
		// Set before the ready check: a command still waiting holds back every
		// later command for this actor.
		w.inputTracker[actor.Id] = struct{}{}
		if input.ReadyTurn > turnCt {
			return events.CancelAndRequeue
		}
		w.processInput(actor, input.InputText, input.Flags)
		return events.Continue
	}

	// Ready turn 0 runs at once, bypasses the limit and does not count towards it.
	if input.ReadyTurn == 0 {
		w.unblockIfDone(actor.Id, user, input.Flags)
		w.processInput(actor, input.InputText, input.Flags)
		return events.Continue
	}

	// One queued command per actor per event-loop pass.
	if _, ok := w.inputTracker[actor.Id]; ok {
		return events.CancelAndRequeue
	}

	// Not ready yet. The tracker is not set, so a later command that is ready
	// may run first: an actor with an account is not strictly in order.
	if input.ReadyTurn > turnCt {
		// A multi-turn wait flagged to block input records when it began.
		if input.Flags.Has(events.CmdBlockInput) {
			if _, ok := w.ignoreInput[actor.Id]; !ok {
				w.ignoreInput[actor.Id] = turnCt
			}
		}
		return events.CancelAndRequeue
	}

	// Ready to run.
	w.unblockIfDone(actor.Id, user, input.Flags)
	w.processInput(actor, input.InputText, input.Flags)
	w.inputTracker[actor.Id] = struct{}{}

	return events.Continue
}

// unblockIfDone unblocks the account's input when a command flagged to
// unblock it runs and its blocking wait was recorded.
func (w *World) unblockIfDone(actorId ids.ID, user *accounts.Account, flags events.EventFlag) {
	if !flags.Has(events.CmdUnBlockInput) {
		return
	}
	if _, ok := w.ignoreInput[actorId]; ok {
		delete(w.ignoreInput, actorId)
		user.UnblockInput()
	}
}

// Checks whether their level is too high for a guide
func (w *World) HandleSystemEvents(e events.Event) events.ListenerReturn {

	sys, typeOk := e.(events.System)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "System", "Actual Type", e.Type())
		return events.Continue
	}

	if sys.Command == `reload` {

		events.AddToQueue(events.Broadcast{
			Text: `Reloading flat files...`,
		})

		loadAllDataFiles(true)

		events.AddToQueue(events.Broadcast{
			Text:            `Done.` + term.CRLFStr,
			SkipLineRefresh: true,
		})

	} else if sys.Command == `kick` {
		w.Kick(sys.Data.(int), sys.Description)
	} else if sys.Command == `leaveworld` {

		if userInfo := accounts.GetByUserId(sys.Data.(int)); userInfo != nil {
			events.AddToQueue(events.PlayerDespawn{
				UserId:        userInfo.UserId,
				RoomId:        userInfo.Character.RoomId,
				Username:      userInfo.Username,
				CharacterName: userInfo.Character.Name,
				TimeOnline:    userInfo.GetOnlineInfo().OnlineTimeStr,
			})
		}

	} else if sys.Command == `logoff` {

		if user := accounts.GetByUserId(sys.Data.(int)); user != nil {

			user.EventLog.Add(`conn`, `Logged off`)

			events.AddToQueue(events.PlayerDespawn{
				UserId:        user.UserId,
				RoomId:        user.Character.RoomId,
				Username:      user.Username,
				CharacterName: user.Character.Name,
				TimeOnline:    user.GetOnlineInfo().OnlineTimeStr,
			})

		}

	}

	return events.Continue
}

// Send input to the world.
// Just sends via a channel. Will block until read.
func (w *World) SendInput(i WorldInput) {
	w.worldInput <- i
}

func (w *World) SendEnterWorld(userId int, roomId int) {
	w.enterWorldUserId <- [2]int{userId, roomId}
}

func (w *World) SendLeaveWorld(userId int) {
	if !serverAlive.Load() {
		return
	}
	w.leaveWorldUserId <- userId
}

func (w *World) SendLogoutConnectionId(connId connections.ConnectionId) {
	if !serverAlive.Load() {
		return
	}
	w.logoutConnectionId <- connId
}

func (w *World) SendSetLinkDead(userId int, on bool) {
	if !serverAlive.Load() {
		return
	}
	if on {
		w.linkDeadFlag <- [2]int{userId, 1}
	} else {
		w.linkDeadFlag <- [2]int{userId, 0}
	}
}

func (w *World) logOutUserByConnectionId(connectionId connections.ConnectionId) {

	if err := accounts.LogOutUserByConnectionId(connectionId); err != nil {
		mudlog.Error("Log Out Error", "connectionId", connectionId, "error", err)
	}
}

func (w *World) enterWorld(userId int, roomId int) {

	userInfo := accounts.GetByUserId(userId)
	if userInfo != nil {
		events.AddToQueue(events.PlayerSpawn{
			UserId:        userInfo.UserId,
			ConnectionId:  userInfo.ConnectionId(),
			RoomId:        userInfo.Character.RoomId,
			Username:      userInfo.Username,
			CharacterName: userInfo.Character.Name,
		})
	}

	w.UpdateStats()

	// Put htme in the room
	if userInfo != nil {
		rooms.MoveToRoom(userInfo.Character.Id, roomId, true)
	}
}

/*
users can be:
Disconnected	+ OutWorld (no presence)	No record in connections.netConnections or users.LinkDeadConnections	| user object in room
Connected		+ OutWorld (logging in) 	Has record in connections.netConnections 							| user object in room
Connected		+ InWorld  (non-link-dead)	No record in users.LinkDeadConnections								| no link-dead flag		| user object in room
Disconnected	+ InWorld  (link-dead)			Has record in users.LinkDeadConnections 								| has link-dead flag		| user object in room
*/

const (
	// Used in GameTickWorker()
	// Used in MaintenanceWorker()
	roomMaintenancePeriod = time.Second * 3  // Every 3 seconds run room maintenance.
	serverStatsLogPeriod  = time.Second * 60 // Every 60 seconds log server stats.
	ansiAliasReloadPeriod = time.Second * 4  // Every 4 seconds reload ansi aliases.
)

func (w *World) MainWorker(shutdown chan bool, wg *sync.WaitGroup) {

	wg.Add(1)

	mudlog.Info("MainWorker", "state", "Started")
	defer func() {
		mudlog.Warn("MainWorker", "state", "Stopped")
		wg.Done()
	}()

	c := configs.GetConfig()

	roomUpdateTimer := time.NewTimer(roomMaintenancePeriod)
	ansiAliasTimer := time.NewTimer(ansiAliasReloadPeriod)
	eventLoopTimer := time.NewTimer(time.Millisecond)
	turnTimer := time.NewTimer(time.Duration(c.Timing.TurnMs) * time.Millisecond)
	statsTimer := time.NewTimer(time.Duration(10) * time.Second)

loop:
	for {

		// The reason for
		// util.LockGame() / util.UnlockGame()
		// In each of these cases is to lock down the
		// logic for when other processes need to query data
		// such as the webserver

		select {
		case <-shutdown:

			mudlog.Warn(`MainWorker`, `action`, `shutdown received`)

			util.LockMud()
			if err := rooms.SaveAllRooms(); err != nil {
				mudlog.Error("rooms.SaveAllRooms()", "error", err.Error())
			}
			accounts.SaveAllUsers() // Save all user data too.
			util.UnlockMud()

			break loop
		case <-statsTimer.C:

			// TODO: Move this to events
			util.LockMud()

			w.UpdateStats()
			// save the round counter.
			util.SaveRoundCount(c.FilePaths.DataFiles.String() + `/` + util.RoundCountFilename)

			util.UnlockMud()

			statsTimer.Reset(time.Duration(10) * time.Second)

		case <-roomUpdateTimer.C:

			// TODO: Move this to events
			util.LockMud()
			rooms.RoomMaintenance()
			rooms.EphemeralRoomMaintenance()
			util.UnlockMud()

			roomUpdateTimer.Reset(roomMaintenancePeriod)

		case <-ansiAliasTimer.C:

			// TODO: Move this to events
			util.LockMud()
			templates.LoadAliases()
			util.UnlockMud()

			ansiAliasTimer.Reset(ansiAliasReloadPeriod)

		case <-eventLoopTimer.C:

			eventLoopTimer.Reset(time.Millisecond)

			util.LockMud()
			w.EventLoop()
			util.UnlockMud()

		case <-turnTimer.C:

			util.LockMud()
			turnTimer.Reset(time.Duration(c.Timing.TurnMs) * time.Millisecond)

			turnCt := util.IncrementTurnCount()

			events.AddToQueue(events.NewTurn{TurnNumber: turnCt, TimeNow: time.Now()})

			// After a full round of turns, we can do a round tick.
			if turnCt%uint64(c.Timing.TurnsPerRound()) == 0 {

				roundNumber := util.IncrementRoundCount()

				events.AddToQueue(events.NewRound{RoundNumber: roundNumber, TimeNow: time.Now()})
			}

			util.UnlockMud()

		case enterWorldUserId := <-w.enterWorldUserId: // [2]int

			util.LockMud()
			w.enterWorld(enterWorldUserId[0], enterWorldUserId[1])
			util.UnlockMud()

		case leaveWorldUserId := <-w.leaveWorldUserId: // int

			util.LockMud()
			if userInfo := accounts.GetByUserId(leaveWorldUserId); userInfo != nil {
				events.AddToQueue(events.PlayerDespawn{
					UserId:        userInfo.UserId,
					RoomId:        userInfo.Character.RoomId,
					Username:      userInfo.Username,
					CharacterName: userInfo.Character.Name,
					TimeOnline:    userInfo.GetOnlineInfo().OnlineTimeStr,
				})
			}
			util.UnlockMud()

		case logoutConnectionId := <-w.logoutConnectionId: //  connections.ConnectionId

			util.LockMud()
			w.logOutUserByConnectionId(logoutConnectionId)
			util.UnlockMud()

		case linkDeadFlag := <-w.linkDeadFlag: //  [2]int
			if linkDeadFlag[1] == 1 {

				util.LockMud()
				accounts.SetLinkDeadUser(linkDeadFlag[0])
				events.AddToQueue(events.PlayerChanged{
					UserId: linkDeadFlag[0],
				})
				util.UnlockMud()

			}
		}
		c = configs.GetConfig()
	}

}

// Should be goroutine/threadsafe
// Only reads from world channel
func (w *World) InputWorker(shutdown chan bool, wg *sync.WaitGroup) {
	wg.Add(1)

	mudlog.Info("InputWorker", "state", "Started")
	defer func() {
		mudlog.Warn("InputWorker", "state", "Stopped")
		wg.Done()
	}()

loop:
	for {
		select {
		case <-shutdown:
			mudlog.Warn(`InputWorker`, `action`, `shutdown received`)
			break loop
		case wi := <-w.worldInput:

			actor := actors.ByUserId(wi.FromId)
			if actor == nil {
				mudlog.Error("InputWorker", "error", "no actor for user, input dropped", "userId", wi.FromId)
				continue
			}

			events.AddToQueue(events.Input{
				ActorId:   actor.Id,
				InputText: wi.InputText,
				ReadyTurn: util.GetTurnCount(),
			})

		}
	}
}

// processInput runs one line of input for any actor. The steps that only make
// sense with an account (prompts, macros, the last-input round, bad-input
// tracking, the "not recognized" text and the prompt redraw) run when the
// actor has one (Decision 14).
func (w *World) processInput(actor *actors.Actor, inputText string, flags events.EventFlag) {

	user := actor.Controller().Account() // nil for an NPC

	var activeQuestion *prompt.Question = nil
	hadPrompt := false
	if user != nil {
		if cmdPrompt := user.GetPrompt(); cmdPrompt != nil {
			hadPrompt = true
			if activeQuestion = cmdPrompt.GetNextQuestion(); activeQuestion != nil {

				activeQuestion.Answer(string(inputText))
				inputText = ``

				// set the input buffer to invoke the command prompt it was relevant to
				if cmdPrompt.Command != `` {
					inputText = cmdPrompt.Command + " " + cmdPrompt.Rest
				}
			} else {
				// If a prompt was found, but no pending questions, clear it.
				user.ClearPrompt()
			}

		}
	}

	command := ``
	remains := ``

	var err error
	handled := false

	inputText = strings.TrimSpace(inputText)

	if len(inputText) > 0 {

		if user != nil {
			// Update their last input
			// Must be actual text, blank space doesn't count.
			user.SetLastInputRound(util.GetRoundCount())
		}

		// Check for macros
		if user != nil && user.Macros != nil && len(inputText) == 2 {
			if macro, ok := user.Macros[inputText]; ok {
				handled = true
				readyTurn := util.GetTurnCount()
				for _, newCmd := range strings.Split(macro, `;`) {
					if newCmd == `` {
						continue
					}

					events.AddToQueue(events.Input{
						ActorId:   user.Character.Id,
						InputText: newCmd,
						ReadyTurn: readyTurn,
					})

					readyTurn++
				}
			}
		}

		if !handled {

			// Lets users use gossip/say shortcuts without a space
			if len(inputText) > 1 {
				if inputText[0] == '`' || inputText[0] == '.' {
					inputText = fmt.Sprintf(`%s %s`, string(inputText[0]), string(inputText[1:]))
				}
			}

			if index := strings.Index(inputText, " "); index != -1 {
				command, remains = strings.ToLower(inputText[0:index]), inputText[index+1:]
			} else {
				command = inputText
			}

			handled, err = usercommands.TryCommand(command, remains, actor.Id, flags)
			if err != nil {
				mudlog.Warn("user-TryCommand", "command", command, "remains", remains, "error", err.Error())
			}
		}

	} else if user != nil {
		connId := user.ConnectionId()
		connections.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt())), connId)
	}

	if !handled && len(command) > 0 {
		if user == nil {
			// Decision 3: an NPC's unknown command goes to the log; the room
			// sees nothing.
			if usercommands.IsCommand(command) {
				// A real command that declines an actor without an account.
				mudlog.Warn("input", "event", "needs an account", "actorId", actor.Id.String(), "name", actor.Character.Name, "command", command, "rest", remains)
				return
			}
			mudlog.Warn("input", "event", "unknown command", "actorId", actor.Id.String(), "name", actor.Character.Name, "command", command, "rest", remains)
			return
		}

		badinputtracker.TrackBadCommand(command, remains)

		user.SendText(fmt.Sprintf(`<ansi fg="command">%s</ansi> not recognized. Type <ansi fg="command">help</ansi> for commands.`, command))
		user.Command(`emote @looks a little confused`)
	}

	if user == nil {
		return
	}

	// If they had an input prompt, but now they don't, lets make sure to resend a status prompt
	if hadPrompt || (!hadPrompt && user.GetPrompt() != nil) {
		connId := user.ConnectionId()
		connections.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt())), connId)
	}
	// Removing this as possibly redundant.
	// Leaving in case I need to remember that I did it...
	//connId := user.ConnectionId()
	//connections.SendTo([]byte(templates.AnsiParse(user.GetCommandPrompt(true))), connId)

}

func (w *World) UpdateStats() {
	s := web.GetStats()
	s.Reset()

	c := configs.GetNetworkConfig()

	for _, u := range accounts.GetAllActiveUsers() {
		info := u.GetOnlineInfo()
		s.OnlineUsers = append(s.OnlineUsers, info)
		switch info.ConnType {
		case "websocket":
			s.WebSocketConnections++
		case "ssh":
			s.SSHConnections++
		default:
			s.TelnetConnections++
		}
	}

	sort.Slice(s.OnlineUsers, func(i, j int) bool {
		if s.OnlineUsers[i].Role == accounts.RoleAdmin {
			return true
		}
		if s.OnlineUsers[j].Role == accounts.RoleAdmin {
			return false
		}
		return s.OnlineUsers[i].OnlineTime > s.OnlineUsers[j].OnlineTime
	})

	for _, t := range c.TelnetPort {
		p, _ := strconv.Atoi(t)
		if p > 0 {
			s.TelnetPorts = append(s.TelnetPorts, p)
		}
	}

	s.WebSocketPort = int(c.HttpPort)
	s.SSHPort = int(c.SSHPort)

	web.UpdateStats(s)
}

// Force disconnect a user (Makes them linkdead)
func (w *World) Kick(userId int, reason string) {

	user := accounts.GetByUserId(userId)
	if user == nil {
		return
	}

	accounts.SetLinkDeadUser(userId)
	user.EventLog.Add(`conn`, fmt.Sprintf(`Kicked (%s)`, reason))

	connections.Kick(user.ConnectionId(), reason)
}

// Should only handle sending messages out to users
func (w *World) EventLoop() {

	events.ProcessEvents()

	clear(w.inputTracker)
}
