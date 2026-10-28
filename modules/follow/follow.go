package follow

import (
	"embed"
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/behaviour"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
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
	f := FollowModule{
		plug:         plugins.New(`follow`, `1.0`),
		followed:     make(map[followId][]followId),
		followers:    make(map[followId]followId),
		followLimits: make(map[followId]uint64),
	}

	//
	// Add the embedded filesystem
	//
	if err := f.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	f.plug.Web.AdminPage("Config", "follow-config", "html/admin/follow-config.html", true, "Modules", "Follow", "Configure follow behavior and allowed follow distances.", "Follow system allowing players and NPCs to trail behind another character.", nil)
	f.plug.Web.AdminPage("About", "follow-about", "html/admin/follow-about.html", true, "Modules", "Follow", "Information and version details for the Follow module.", "", nil) //
	// Register the command (players and NPCs run the same one)
	//
	f.plug.AddUserCommand(`follow`, f.followCommand, true, false)

	events.RegisterListener(events.RoomChange{}, f.roomChangeHandler)
	events.RegisterListener(events.PlayerDespawn{}, f.playerDespawnHandler)
	events.RegisterListener(events.MobDeath{}, f.onMobDeath)
	events.RegisterListener(events.PlayerDeath{}, f.onPlayerDeath)
	behaviour.OnIdle.Register(f.idleMobHandler)
	events.RegisterListener(events.PartyUpdated{}, f.onPartyChange)
	events.RegisterListener(events.NewRound{}, f.onNewRound)
}

//////////////////////////////////////////////////////////////////////
// NOTE: What follows is all custom code. For this module.
//////////////////////////////////////////////////////////////////////

type followId struct {
	userId int
	mobId  ids.ID
}

// Using a struct gives a way to store longer term data.
type FollowModule struct {
	// Keep a reference to the plugin when we create it so that we can call ReadBytes() and WriteBytes() on it.
	plug *plugins.Plugin

	followed     map[followId][]followId // key => who's followed. value ([]followId{}) => who's following them
	followers    map[followId]followId   // key => who's following someone. value => who's being followed
	followLimits map[followId]uint64     // Key => follower Id, value => round the follow forcibly ends
}

// Get all followeres attached to a target
func (f *FollowModule) isFollowing(followCheck followId) bool {
	_, ok := f.followers[followCheck]
	return ok
}

// Get all followeres attached to a target
func (f *FollowModule) getFollowers(followTarget followId) []followId {

	if _, ok := f.followed[followTarget]; !ok {
		return []followId{}
	}

	followerResults := make([]followId, len(f.followed[followTarget]))
	copy(followerResults, f.followed[followTarget])

	return followerResults
}

// Add a single follower to a target
func (f *FollowModule) startFollow(followTarget followId, followSource followId, followCutoffRoundId ...uint64) {

	// Make sure they no longer follow whoever they were before.
	f.stopFollowing(followSource)

	f.followers[followSource] = followTarget
	if _, ok := f.followed[followTarget]; !ok {
		f.followed[followTarget] = []followId{}
	}

	f.followed[followTarget] = append(f.followed[followTarget], followSource)

	if len(followCutoffRoundId) > 0 && followCutoffRoundId[0] > 0 {
		f.followLimits[followSource] = followCutoffRoundId[0]
	}
}

// Remove a single follower from whoever they are following (if any)
func (f *FollowModule) stopFollowing(followSource followId) followId {

	wasFollowing := followId{}

	if followTarget, ok := f.followers[followSource]; ok {
		delete(f.followers, followSource)

		wasFollowing = followTarget

		for idx, fId := range f.followed[followTarget] {
			if fId == followSource {
				f.followed[followTarget] = append(f.followed[followTarget][0:idx], f.followed[followTarget][idx+1:]...)

				if len(f.followed[followTarget]) == 0 {
					delete(f.followed, followTarget)
				}

				break
			}
		}
	}

	// If there was a limit, delete it.
	delete(f.followLimits, followSource)

	return wasFollowing
}

// Remove all followers from a target
func (f *FollowModule) loseFollowers(followTarget followId) []followId {
	allFollowers := f.getFollowers(followTarget)
	for _, followSource := range allFollowers {
		f.stopFollowing(followSource)
	}
	return allFollowers
}

//
// Event Handlers
//

func (f followId) getFollowIdInstance() (user *accounts.Account, mob *mobs.Mob) {
	if f.userId > 0 {
		return accounts.GetByUserId(f.userId), nil
	}
	if !f.mobId.IsZero() {
		return nil, mobs.GetInstance(f.mobId)
	}
	return nil, nil
}

// Does a cleanup check every round for any follows that have expired.
func (f *FollowModule) onNewRound(e events.Event) events.ListenerReturn {

	evt := e.(events.NewRound)

	for fId, rNum := range f.followLimits {
		if rNum > evt.RoundNumber {
			continue
		}

		wasFollowing := f.stopFollowing(fId)

		followTargetUser, followTargetMob := wasFollowing.getFollowIdInstance()
		followSourceUser, followSourceMob := fId.getFollowIdInstance()

		// user being followed?
		if followTargetUser != nil {

			// user doing the following? Tell both users
			if followSourceUser != nil {
				messaging.Tell(followTargetUser.Character.Id, fmt.Sprintf(`<ansi fg="username">%s</ansi> stopped following you.`, followSourceUser.Character.Name))
				messaging.Tell(followSourceUser.Character.Id, fmt.Sprintf(`You are no longer following <ansi fg="username">%s</ansi>.`, followTargetUser.Character.Name))
				continue
			}

			// mob doing the following? tell the target user
			if followSourceMob != nil {
				messaging.Tell(followTargetUser.Character.Id, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> stopped following you.`, followSourceMob.Character.Name))
				continue
			}

			continue
		}

		// mob being followed?
		if followTargetMob != nil {

			// user doing the following? Tell the following user
			if followSourceUser != nil {
				messaging.Tell(followSourceUser.Character.Id, fmt.Sprintf(`You are no longer following <ansi fg="mobname">%s</ansi>.`, followTargetMob.Character.Name))
			}

			continue
		}

	}

	return events.Continue
}

// If players make changes (into/out of party)
// Just make sure they aren't following anyone.
// This is just basic cleanup/precaution
func (f *FollowModule) onPartyChange(e events.Event) events.ListenerReturn {

	evt := e.(events.PartyUpdated)

	for _, uId := range evt.UserIds {
		f.stopFollowing(followId{userId: uId})
	}

	return events.Continue
}

// idleMobHandler keeps a following NPC from idling, so it stays with its leader.
func (f *FollowModule) idleMobHandler(c behaviour.IdleCheck) behaviour.IdleCheck {
	if f.isFollowing(followId{mobId: c.Mob.Character.Id}) {
		c.Skip = true
	}
	return c
}

func (f *FollowModule) roomChangeHandler(e events.Event) events.ListenerReturn {
	evt := e.(events.RoomChange)

	// A player's follow key is the account's UserId; a mob's is its ID.
	moverUser := actors.AccountOf(evt.ActorId)
	moverId := followId{mobId: evt.ActorId}
	if moverUser != nil {
		moverId = followId{userId: moverUser.UserId}
	}

	allFollowers := f.getFollowers(moverId)
	if len(allFollowers) == 0 {
		return events.Continue
	}

	fromRoom := rooms.LoadRoom(evt.FromRoomId)
	if fromRoom == nil {
		return events.Continue
	}

	followExitName := ``
	for exitName, exitInfo := range fromRoom.Exits {
		if exitInfo.RoomId == evt.ToRoomId {
			followExitName = exitName
			break
		}
	}

	if followExitName == `` {
		for exitName, exitInfo := range fromRoom.ExitsTemp {
			if exitInfo.RoomId == evt.ToRoomId {
				followExitName = exitName
				break
			}
		}
	}

	// The exit they went through is gone/missing? (Teleported?)
	// End the follow
	if followExitName == `` {
		if moverUser != nil {
			moverUser.Command(`follow lose`)
		}
	} else {

		for _, fId := range allFollowers {

			if !fId.mobId.IsZero() {

				if mob := mobs.GetInstance(fId.mobId); mob != nil {
					if fromRoom.RoomId == mob.Character.RoomId {
						mob.Command(followExitName, .25)
						// Count follows as wandering
						// This way if following ends, many/most mobs will head home.
						mob.WanderCount++
						continue
					}

					mob.Command(`follow stop`)
				}
				f.stopFollowing(fId)

			} else if fId.userId > 0 {

				if user := accounts.GetByUserId(fId.userId); user != nil {
					if fromRoom.RoomId == user.Character.RoomId {
						user.Command(followExitName, .25)
						continue
					}

					user.Command(`follow stop`)
				}
				f.stopFollowing(fId)

			}

		}

	}

	return events.Continue
}

func (f *FollowModule) playerDespawnHandler(e events.Event) events.ListenerReturn {
	// Don't really care about the event data for this
	evt, typeOk := e.(events.PlayerDespawn)
	if !typeOk {
		return events.Cancel
	}

	f.loseFollowers(followId{userId: evt.UserId})

	return events.Continue
}

func (f *FollowModule) onMobDeath(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(events.MobDeath)
	if !typeOk {
		return events.Cancel
	}

	f.loseFollowers(followId{mobId: evt.ActorId})

	return events.Continue
}

func (f *FollowModule) onPlayerDeath(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(events.PlayerDeath)
	if !typeOk {
		return events.Cancel
	}

	f.loseFollowers(followId{userId: evt.UserId})

	return events.Continue
}

//
// Commands
//

// followCommand is the one follow command for players and NPCs. A follow key
// is the account's UserId for a player and the body's ID for an NPC.
func (f *FollowModule) followCommand(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	user := actor.Controller().Account() // nil for an NPC

	if rest == "" {
		usercommands.Tell(actor, `Follow whom? Try <ansi fg="command">help command</ansi>`)
		return true, nil
	}

	// Parties belong to accounts.
	if user != nil && parties.Get(user.UserId) != nil {
		usercommands.Tell(actor, `You can't use this command while in a party.`)
		return true, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))

	followTargetName := args[0]
	followAction := `follow`
	followEndRound := uint64(0)

	if rest == `stop` || rest == `lose` {
		followAction = rest
		followTargetName = ``
	}

	gd := gametime.GetDate(util.GetRoundCount())

	if len(args) > 1 {
		followEndRound = gd.AddPeriod(strings.Join(args[1:], ` `))
	} else if followPeriod, ok := f.plug.Config.Get(`DefaultFollowPeriod`).(string); ok {
		followEndRound = gd.AddPeriod(followPeriod)
	}

	// in case something went wrong, we still want to cap it.
	if followEndRound <= util.GetRoundCount() {
		followEndRound = gd.AddPeriod(`5 real minutes`)
	}

	userId, targetMobId := 0, ids.ID{}
	if len(followTargetName) > 0 {
		userId, targetMobId = room.FindByName(followTargetName)
	}

	followCommandTarget := followId{userId: userId, mobId: targetMobId}
	followCommandSource := followId{mobId: actor.Id}
	if user != nil {
		followCommandSource = followId{userId: user.UserId}
	}

	if (user != nil && followCommandTarget.userId == followCommandSource.userId) ||
		(user == nil && followCommandTarget.mobId == followCommandSource.mobId) {
		usercommands.Tell(actor, `You can't target yourself.`)
		return true, nil
	}

	// Lose any followers
	if followAction == `lose` {

		if lostFollowers := f.loseFollowers(followCommandSource); len(lostFollowers) > 0 {

			// Tell all the followers they
			for _, fId := range lostFollowers {
				if fId.userId == 0 {
					continue
				}

				if followerUser := accounts.GetByUserId(fId.userId); followerUser != nil {
					messaging.Tell(followerUser.Character.Id, fmt.Sprintf(`You are no longer following <ansi fg="%s">%s</ansi>.`, usercommands.NameColour(actor), actor.Character.Name))
				}
			}

		}

		usercommands.Tell(actor, `Nobody is following you.`)

		return true, nil
	}

	// Stop following someone?
	if followAction == `stop` {

		wasFollowing := f.stopFollowing(followCommandSource)

		if wasFollowing.userId > 0 {

			if followUser := accounts.GetByUserId(wasFollowing.userId); followUser != nil {
				messaging.Tell(followUser.Character.Id, fmt.Sprintf(`<ansi fg="%s">%s</ansi> stopped following you.`, usercommands.NameColour(actor), actor.Character.Name))
				usercommands.Tell(actor, fmt.Sprintf(`You are no longer following <ansi fg="username">%s</ansi>.`, followUser.Character.Name))
				return true, nil
			}

		}

		if !wasFollowing.mobId.IsZero() {

			if followMob := mobs.GetInstance(wasFollowing.mobId); followMob != nil {
				usercommands.Tell(actor, fmt.Sprintf(`You are no longer following <ansi fg="mobname">%s</ansi>.`, followMob.Character.Name))
				return true, nil
			}

		}

		usercommands.Tell(actor, `You aren't following anyone.`)

		return true, nil
	}

	// Default behavior is follow

	// An NPC on a path drops it. The follow takes priority.
	if mob := mobs.GetInstance(actor.Id); mob != nil {
		mob.Path.Clear()
	}

	if followCommandTarget.userId > 0 {

		f.startFollow(followCommandTarget, followCommandSource, followEndRound)

		targetUser := accounts.GetByUserId(followCommandTarget.userId)

		usercommands.Tell(actor, fmt.Sprintf(`You start following <ansi fg="username">%s</ansi>.`, targetUser.Character.Name))

		messaging.Tell(targetUser.Character.Id, fmt.Sprintf(`<ansi fg="%s">%s</ansi> is following you.`, usercommands.NameColour(actor), actor.Character.Name))

		return true, nil
	}

	if !followCommandTarget.mobId.IsZero() {

		targetMob := mobs.GetInstance(followCommandTarget.mobId)

		if targetMob.HatesAlignment(actor.Character.Alignment) {
			usercommands.Tell(actor, fmt.Sprintf(`<ansi fg="mobname">%s</ansi> won't let you follow them.`, targetMob.Character.Name))
		} else {
			f.startFollow(followCommandTarget, followCommandSource, followEndRound)

			usercommands.Tell(actor, fmt.Sprintf(`You start following <ansi fg="mobname">%s</ansi>.`, targetMob.Character.Name))
		}

		return true, nil
	}

	usercommands.Tell(actor, `Follow whom?`)

	return true, nil
}
