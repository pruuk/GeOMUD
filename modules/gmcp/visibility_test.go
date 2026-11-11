package gmcp

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	actors.Init()
	os.Exit(m.Run())
}

const (
	visRoom      = 95001
	visHiddenBuf = 95901
	visSeeBuf    = 95902
)

// visWorld is one room for the GMCP visibility tests; events queued while
// setting it up are dropped.
func visWorld(t *testing.T) *rooms.Room {
	t.Helper()
	actors.Reset()
	accounts.ResetActiveUsers()
	r := rooms.NewEmptyRoom()
	r.RoomId, r.Zone, r.Title = visRoom, "gmcpzone", "GMCP Room"
	rooms.AddTestRoom(r)
	t.Cleanup(func() {
		rooms.RemoveTestRoom(visRoom)
		accounts.ResetActiveUsers()
		actors.Reset()
	})
	return r
}

func visPlayer(t *testing.T, r *rooms.Room, userId int, name string) *accounts.Account {
	t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name, c.RoomId, c.Zone = name, r.RoomId, r.Zone
	u := &accounts.Account{UserId: userId, Username: name, Character: c, Role: accounts.RoleUser}
	require.NotNil(t, actors.Spawn(c, actors.NewConnectionController(u)))
	accounts.SetTestUser(u)
	r.AddPlayer(userId)
	return u
}

func visMob(t *testing.T, r *rooms.Room, name string) *mobs.Mob {
	t.Helper()
	m := mobs.Mob{MobId: 95201}
	m.Character = *characters.New()
	m.Character.Name = name
	inst := mobs.NewTestInstance(m)
	id := inst.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	r.AddMob(id)
	return inst
}

// visFlag gives c a test buff carrying one flag.
func visFlag(t *testing.T, c *characters.Character, buffId int, flag string) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000, Flags: []string{flag},
	}))
	require.NoError(t, c.AddBuff(buffId, true))
	require.True(t, c.HasBuffFlag(flag))
}

// outs collects the GMCPOut and GMCPRoomUpdate events fn queues.
func outs(t *testing.T, fn func()) (sent []GMCPOut, updates []GMCPRoomUpdate) {
	t.Helper()
	events.ProcessEvents()
	a := events.RegisterListener(GMCPOut{}, func(e events.Event) events.ListenerReturn {
		sent = append(sent, e.(GMCPOut))
		return events.Cancel // not on to a connection
	}, events.First)
	b := events.RegisterListener(GMCPRoomUpdate{}, func(e events.Event) events.ListenerReturn {
		updates = append(updates, e.(GMCPRoomUpdate))
		return events.Cancel
	}, events.First)
	defer events.UnregisterListener(GMCPOut{}, a)
	defer events.UnregisterListener(GMCPRoomUpdate{}, b)
	fn()
	events.ProcessEvents()
	return sent, updates
}

func names(list []GMCPRoomModule_Payload_Contents_Character) []string {
	out := []string{}
	for _, c := range list {
		out = append(out, c.Name)
	}
	return out
}

// The room contents list a hidden body only to a viewer who perceives it
// (rooms.Perceives): see-hidden, an admin, never anyone else.
func TestRoomContentsListOnlyWhatTheViewerPerceives(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	bob := visPlayer(t, r, 9502, "Bob")
	ada := visPlayer(t, r, 9503, "Ada")
	ada.Role = accounts.RoleAdmin
	rat := visMob(t, r, "rat")
	visFlag(t, bob.Character, visHiddenBuf, "hidden")
	visFlag(t, &rat.Character, visHiddenBuf, "hidden")
	g := &GMCPRoomModule{}

	players, _ := g.GetRoomNode(ann, `Room.Info.Contents.Players`)
	npcs, _ := g.GetRoomNode(ann, `Room.Info.Contents.Npcs`)
	assert.Equal(t, []string{"Ada"}, names(players.([]GMCPRoomModule_Payload_Contents_Character)), "Ann does not perceive hidden Bob")
	assert.Empty(t, npcs.([]GMCPRoomModule_Payload_Contents_Character), "nor the hidden rat")

	players, _ = g.GetRoomNode(ada, `Room.Info.Contents.Players`)
	npcs, _ = g.GetRoomNode(ada, `Room.Info.Contents.Npcs`)
	assert.Equal(t, []string{"Ann", "Bob"}, names(players.([]GMCPRoomModule_Payload_Contents_Character)), "an admin perceives Bob")
	assert.Equal(t, []string{"rat"}, names(npcs.([]GMCPRoomModule_Payload_Contents_Character)), "and the rat")

	visFlag(t, ann.Character, visSeeBuf, "see-hidden")
	npcs, _ = g.GetRoomNode(ann, `Room.Info.Contents.Npcs`)
	assert.Equal(t, []string{"rat"}, names(npcs.([]GMCPRoomModule_Payload_Contents_Character)), "see-hidden perceives it")
}

// Char.Enemies lists a hidden enemy only to a viewer who perceives it.
func TestEnemiesListOnlyWhatTheViewerPerceives(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	rat := visMob(t, r, "rat")
	rat.Character.SetAggro(ann.Character.Id, characters.DefaultAttack)
	visFlag(t, &rat.Character, visHiddenBuf, "hidden")
	g := &GMCPCharModule{}

	enemies, _ := g.GetCharNode(ann, `Char.Enemies`)
	assert.Empty(t, enemies.([]GMCPCharModule_Enemy), "Ann does not perceive the hidden rat fighting her")

	ann.Role = accounts.RoleAdmin
	enemies, _ = g.GetCharNode(ann, `Char.Enemies`)
	require.Len(t, enemies.([]GMCPCharModule_Enemy), 1, "an admin does")
	assert.Equal(t, "rat", enemies.([]GMCPCharModule_Enemy)[0].Name)
}

// A say names its speaker on Comm.Channel only to a receiver who perceives
// it; to anyone else the sender reads "Someone", as the text does.
func TestSayNamesItsSpeakerOnlyToThoseWhoPerceiveIt(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	sneak := visPlayer(t, r, 9502, "Sneak")
	ada := visPlayer(t, r, 9503, "Ada")
	ada.Role = accounts.RoleAdmin
	visFlag(t, sneak.Character, visHiddenBuf, "hidden")

	sent, _ := outs(t, func() {
		(&GMCPCommModule{}).onComm(events.Communication{SourceActorId: sneak.Character.Id, CommType: `say`, Name: "Sneak", Message: "hi"})
	})
	senders := map[int]GMCPCommModule_Payload{}
	for _, o := range sent {
		senders[o.UserId] = o.Payload.(GMCPCommModule_Payload)
	}
	require.Len(t, senders, 3)
	assert.Equal(t, "Someone", senders[ann.UserId].Sender, "Ann does not perceive Sneak")
	assert.Empty(t, senders[ann.UserId].Source, "nor learns whether it is a player")
	assert.Equal(t, "Sneak", senders[ada.UserId].Sender, "an admin does")
	assert.Equal(t, "player", senders[ada.UserId].Source)
	assert.Equal(t, "Sneak", senders[sneak.UserId].Sender, "Sneak perceives itself")
}

// Leaving the game and moving name or announce a body only to viewers who
// perceive it.
func TestRoomEventsReachOnlyThoseWhoPerceiveTheBody(t *testing.T) {
	r := visWorld(t)
	ann := visPlayer(t, r, 9501, "Ann")
	ada := visPlayer(t, r, 9502, "Ada")
	ada.Role = accounts.RoleAdmin
	sneak := visPlayer(t, r, 9503, "Sneak")
	visFlag(t, sneak.Character, visHiddenBuf, "hidden")
	g := &GMCPRoomModule{}

	sent, _ := outs(t, func() {
		g.despawnHandler(events.PlayerDespawn{UserId: sneak.UserId, RoomId: r.RoomId, CharacterName: "Sneak"})
	})
	to := []int{}
	for _, o := range sent {
		to = append(to, o.UserId)
	}
	assert.Equal(t, []int{ada.UserId}, to, "only the admin reads Room.RemovePlayer for the hidden Sneak")

	_, updates := outs(t, func() {
		g.roomChangeHandler(events.RoomChange{ActorId: sneak.Character.Id, ToRoomId: r.RoomId})
	})
	to = []int{}
	for _, u := range updates {
		if u.Identifier == `Room.Info.Contents.Players` {
			to = append(to, u.UserId)
		}
	}
	assert.Equal(t, []int{ada.UserId}, to, "only the admin's room list is refreshed for the hidden mover")
	assert.NotContains(t, to, ann.UserId)

	// The room left behind: the same rule (its own loop).
	_, updates = outs(t, func() {
		g.roomChangeHandler(events.RoomChange{ActorId: sneak.Character.Id, FromRoomId: r.RoomId})
	})
	to = []int{}
	for _, u := range updates {
		if u.Identifier == `Room.Info.Contents.Players` { // the mover gets its own Room.Info
			to = append(to, u.UserId)
		}
	}
	assert.Equal(t, []int{ada.UserId}, to, "only the admin's list is refreshed in the room left")

	// A fight change: the same rule.
	_, updates = outs(t, func() {
		g.aggroChangedHandler(events.AggroChanged{ActorId: sneak.Character.Id, RoomId: r.RoomId})
	})
	to = []int{}
	for _, u := range updates {
		to = append(to, u.UserId)
	}
	assert.ElementsMatch(t, []int{ada.UserId, sneak.UserId}, to, "the admin and Sneak itself, not Ann")
}
