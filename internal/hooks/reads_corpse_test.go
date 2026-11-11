package hooks

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readsSeeHiddenBuf is the test buff carrying see-hidden.
const readsSeeHiddenBuf = 93910

// A hidden body is not in the list a viewer's names are matched against
// (Calabe, 2026-11-11): no ordinal, no name, not by ID either. A viewer
// with see-hidden perceives it, as does an admin; a body always perceives
// itself.
func TestHiddenBodiesAreNotInTheViewersNames(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, cid, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	w.hide(&rat.Character)
	w.hide(bob.Character)

	playerId, mobId := room.FindByName(ann.Character, "rat")
	assert.Zero(t, playerId)
	assert.True(t, mobId.IsZero(), "the hidden rat cannot be named")
	playerId, _ = room.FindByName(ann.Character, "bob")
	assert.Zero(t, playerId, "nor the hidden Bob")
	_, mobId = room.FindByName(ann.Character, rat.Character.Id.String())
	assert.True(t, mobId.IsZero(), "not by its ID either")
	playerId, _ = room.FindByName(bob.Character, "bob")
	assert.Equal(t, bob.UserId, playerId, "Bob perceives himself")

	w.admin(cid) // an admin perceives the hidden (reads_golden_test.go:381)
	_, mobId = room.FindByName(cid.Character, "rat")
	assert.Equal(t, rat.Character.Id, mobId, "an admin perceives it")

	w.flag(ann.Character, readsSeeHiddenBuf, "see-hidden")
	_, mobId = room.FindByName(ann.Character, "rat")
	assert.Equal(t, rat.Character.Id, mobId, "see-hidden perceives it, as the room view does")
}

// Following does not cancel hiding (Calabe, 2026-11-11): a hidden follower
// the target does not perceive is not announced, not even as "Someone",
// starting or stopping; the follower still reads its own lines. An admin
// target perceives it and reads them as before (reads golden H's follow
// section shows the unhidden lines).
func TestHiddenFollowerIsNotAnnounced(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, _ := w.trio()
	w.hide(bob.Character)

	w.run("Bob", "follow Ann")
	w.run("Bob", "follow stop")
	out := w.out.String()
	assert.NotContains(t, out, `Ann: <ansi fg="username">Bob</ansi> is following you.`)
	assert.NotContains(t, out, `following you.`, "no line to Ann at all, not even as someone")
	assert.Contains(t, out, `Bob: You start following <ansi fg="username">Ann</ansi>.`)
	assert.Contains(t, out, `Bob: You are no longer following <ansi fg="username">Ann</ansi>.`)

	w.admin(ann)
	w.run("Bob", "follow Ann")
	w.run("Bob", "follow stop")
	out = w.out.String()
	assert.Contains(t, out, `Ann: <ansi fg="username">Bob</ansi> is following you.`, "an admin perceives Bob")
	assert.Contains(t, out, `Ann: <ansi fg="username">Bob</ansi> stopped following you.`)
}

// The room view lists a hidden player or NPC only to a viewer who perceives
// it (rooms.Perceives): see-hidden or an admin.
func TestRoomViewListsOnlyWhatTheViewerPerceives(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	w.hide(bob.Character)
	w.hide(&rat.Character)

	d := rooms.GetDetails(room, ann)
	assert.NotContains(t, strings.Join(d.VisiblePlayers, ","), "Bob", "Ann does not perceive hidden Bob")
	assert.NotContains(t, strings.Join(d.VisibleMobs, ","), "rat", "nor the hidden rat")

	w.admin(ann)
	d = rooms.GetDetails(room, ann)
	assert.Contains(t, strings.Join(d.VisiblePlayers, ","), "Bob", "an admin does")
	assert.Contains(t, strings.Join(d.VisibleMobs, ","), "rat")
}

// follow lose (also run by a leader's teleport) tells each follower its
// follow ended: a leader the follower does not perceive is "someone".
func TestFollowLoseNamesOnlyAPerceivedLeader(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, _, _, _ := w.trio()
	w.run("Bob", "follow Ann")
	w.hide(ann.Character)
	w.run("Ann", "follow lose")
	out := w.out.String()
	assert.Contains(t, out, `Bob: You are no longer following someone.`)
	assert.NotContains(t, out, `Bob: You are no longer following <ansi fg="username">Ann</ansi>.`)
}

// When a follow runs out: an NPC follower's "stopped following you" is only
// its presence, so a target who does not perceive it reads nothing; a player
// following an NPC it does not perceive reads "someone".
func TestFollowExpiryFollowsTheRule(t *testing.T) {
	w := newReadsWorld(t, 1)
	_, _, _, rat := w.trio()
	w.run("rat", "follow Ann")
	w.run("Bob", "follow rat")
	w.hide(&rat.Character)
	w.do("NewRound far in the future", func() { events.AddToQueue(events.NewRound{RoundNumber: 1 << 40}) })
	out := w.out.String()
	assert.NotContains(t, out, `Ann: <ansi fg="mobname">rat</ansi> stopped following you.`, "Ann does not perceive the hidden rat")
	assert.NotContains(t, out, `stopped following you`)
	assert.Contains(t, out, `Bob: You are no longer following someone.`)
}

// A line that is only a body's presence (starting to quit, leaving the
// game) reaches only viewers who perceive it (rooms.NotPerceiving).
func TestPresenceLinesReachOnlyThoseWhoPerceive(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, rat := w.trio()
	dan := w.addPlayer("Dan", readsRoomA)
	w.admin(bob)
	w.hide(ann.Character)

	assert.ElementsMatch(t, []ids.ID{dan.Character.Id, rat.Character.Id}, rooms.NotPerceiving(ann.Character),
		"Dan and the rat do not perceive Ann; Bob, an admin, does")

	w.run("Ann", "quit")
	out := w.out.String()
	assert.Contains(t, out, `Bob: <ansi fg="username">Ann</ansi> sits down and begins to meditate.`)
	assert.NotContains(t, out, `Dan: `, "nothing to Dan, not even someone")

	w.do("Ann leaves the game", func() { HandleLeave(events.PlayerDespawn{UserId: ann.UserId, RoomId: readsRoomA}) })
	out = w.out.String()
	assert.Contains(t, out, `Bob: Suddenly and without warning`)
	assert.NotContains(t, out, `Dan: `, "still nothing to Dan")
}

// linesTo is what the record shows the named player reading.
func linesTo(out, who string) string {
	got := []string{}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, who+": ") {
			got = append(got, line)
		}
	}
	return strings.Join(got, "\n")
}

// Each presence-only line, one call site per case (round 5): a hidden body's
// line reaches an admin (Eve) and never a viewer who does not perceive it
// (Dan), not even as "someone".
func TestEachPresenceLineReachesOnlyThoseWhoPerceive(t *testing.T) {
	for _, tc := range []struct {
		name, line string
		run        func(w *readsWorld, ann, zed *characters.Character)
	}{
		{"teleport", "appears in a flash of light", func(w *readsWorld, ann, _ *characters.Character) {
			w.run("Ann", "teleport 93102") // Ann, an admin, into room B
		}},
		{"teleport party", "appears in a flash of light", func(w *readsWorld, _, _ *characters.Character) {
			w.run("Ann", "teleport 93102") // the leader moves Zed, the hidden member, too
		}},
		{"flee", "flees to the", func(w *readsWorld, ann, _ *characters.Character) {
			w.do("Ann flees", func() { flee(actors.Get(ann.Id), rooms.LoadRoom(readsRoomA)) })
		}},
		{"drop", `<ansi fg="red">drops to the ground!</ansi>`, func(w *readsWorld, ann, _ *characters.Character) {
			w.do("Ann drops", func() { HandlePlayerDrop(events.PlayerDrop{UserId: w.players["Ann"].UserId, RoomId: readsRoomA}) })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newReadsWorld(t, 1)
			ann, _, _, rat := w.trio()
			zed := w.addPlayer("Zed", readsRoomA)
			w.admin(ann)
			hidden := ann.Character
			watchRoom := readsRoomA
			switch tc.name {
			case "teleport", "teleport party":
				watchRoom = readsRoomB
			case "flee":
				ann.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
			}
			if tc.name == "teleport party" {
				p := parties.New(ann.UserId)
				require.NotNil(t, p)
				require.True(t, p.InvitePlayer(zed.UserId))
				require.True(t, p.AcceptInvite(zed.UserId))
				t.Cleanup(p.Disband)
				hidden = zed.Character
			}
			dan := w.addPlayer("Dan", watchRoom)
			eve := w.addPlayer("Eve", watchRoom)
			w.admin(eve)
			_, _ = dan, eve
			w.hide(hidden)

			tc.run(w, ann.Character, zed.Character)
			out := w.out.String()
			assert.Contains(t, linesTo(out, "Eve"), hidden.Name+`</ansi> `+tc.line, "an admin perceives the hidden body")
			for _, line := range strings.Split(linesTo(out, "Dan"), "\n") {
				if !strings.Contains(line, tc.line) {
					continue
				}
				assert.NotContains(t, line, hidden.Name, "Dan does not read the hidden body's line")
				assert.NotContains(t, line, "Someone", "not even as someone")
			}
		})
	}
}
