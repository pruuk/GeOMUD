package behaviour

import (
	"strconv"
	"strings"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var initOnce sync.Once

// Init attaches Tick to every NPC's controller (actors.OnBehaviourTick).
// main calls it once at startup, after actors.Init.
func Init() {
	initOnce.Do(func() {
		actors.OnBehaviourTick.Register(func(m *mobs.Mob) *mobs.Mob { Tick(m); return m })
	})
}

// IdleCheck asks whether an NPC may idle this round. A handler sets Skip to
// keep it from idling.
type IdleCheck struct {
	Mob  *mobs.Mob
	Skip bool
}

// OnIdle fires each round for an NPC that is about to idle: it is not held
// back, not fighting and not walking a path. A handler that sets Skip stops
// this round's idle, including the way home past its wander limit; the follow
// module uses it so a following NPC stays with its leader. It replaces the
// MobIdle event (slice 6).
var OnIdle util.Hook[IdleCheck]

// Tick is one NPC's round of behaviour, run after combat and healing through
// BehaviourController.Tick.
func Tick(mob *mobs.Mob) {
	if mob == nil {
		return
	}

	// If idle prevented, it's a one round interrupt (until another comes along)
	if mob.PreventIdle {
		mob.PreventIdle = false
		return
	}

	// If they are doing some sort of combat thing,
	// Don't do idle actions
	if mob.Character.Aggro != nil {
		loseLostQuarry(mob)
		return
	}

	if walkPath(mob) {
		return
	}

	if OnIdle.Fire(IdleCheck{Mob: mob}).Skip {
		return
	}

	idle(mob)
}

// loseLostQuarry ends the fight of an NPC whose target is gone: a player
// target who is offline or in another room is lost. A vanished NPC target is
// treated the same (accepted trade-off).
func loseLostQuarry(mob *mobs.Mob) {
	t := mob.Character.Aggro.Target
	if t.IsZero() || mobs.GetInstance(t) != nil {
		return
	}
	if user := accounts.GetByBodyId(t); user == nil || user.Character.RoomId != mob.Character.RoomId {
		mob.Command(`emote mumbles about losing their quarry.`)
		mob.Character.Aggro = nil
		events.AddToQueue(events.AggroChanged{ActorId: mob.Character.Id, RoomId: mob.Character.RoomId})
	}
}

// walkPath takes the NPC's next step on its path, and reports whether the
// round went on the path. A path that ends or breaks is cleared, and the
// round goes on to idle.
func walkPath(mob *mobs.Mob) bool {
	currentStep := mob.Path.Current()
	if currentStep == nil && mob.Path.Len() == 0 {
		return false
	}

	if currentStep != nil {
		// If their currentStep isn't actually the room they are in
		// They've somehow been moved. Reclaculate a new path.
		if currentStep.RoomId() != mob.Character.RoomId {
			reDoWaypoints := mob.Path.Waypoints()
			if len(reDoWaypoints) > 0 {
				newArgs := make([]string, 0, len(reDoWaypoints))
				for _, wpInt := range reDoWaypoints {
					newArgs = append(newArgs, strconv.Itoa(wpInt))
				}
				PathTo(mob, strings.Join(newArgs, ` `))
				return true
			}
			// if we were unable to come up with a new path, send them home.
			PathTo(mob, "home")
			return true
		}
	}

	if nextStep := mob.Path.Next(); nextStep != nil {
		if room := rooms.LoadRoom(mob.Character.RoomId); room != nil {
			if exitInfo, ok := room.Exits[nextStep.ExitName()]; ok {
				if exitInfo.RoomId == nextStep.RoomId() {
					mob.Command(nextStep.ExitName())
					if nextStep.Waypoint() {
						// Anytime a mob reaches a waypoint, introduce a 1 second delay before they can perform any additional commands.
						// This gives a more natural feel to mob behavior, and gives those following a moment to catch up before the mob does something.
						mob.Command("noop", 1)
					}
					return true
				}
			}
		}
	}

	mob.Path.Clear()
	if mob.HomeRoomId == mob.Character.RoomId {
		mob.WanderCount = 0
	}
	return false
}

// idle is the NPC's free round: home if it has strayed, else maybe one of its
// idle commands at its ActivityLevel odds.
func idle(mob *mobs.Mob) {
	isCharmed := mob.Character.IsCharmed()

	// if a mob shouldn't be allowed to leave their area (via wandering)
	// but has somehow been displaced, such as pulling through combat, spells, or otherwise
	// tell them to path back home
	if mob.MaxWander == 0 && mob.Character.RoomId != mob.HomeRoomId && !isCharmed {
		PathTo(mob, "home")
		return
	}

	if isCharmed {
		return
	}

	// MaxWander moves used up: away from home it heads home; at home it may
	// wander again (it used to stay over the limit and never idle again).
	if mob.MaxWander > 0 && mob.WanderCount >= mob.MaxWander {
		if mob.Character.RoomId != mob.HomeRoomId {
			PathTo(mob, "home")
			return
		}
		mob.WanderCount = 0
	}

	// If they have idle commands, maybe do one of them?
	if util.Rand(100) < mob.ActivityLevel {
		if idleCmd := mob.GetIdleCommand(); idleCmd != `` {
			RunIdle(mob, idleCmd)
		}
	}
}
