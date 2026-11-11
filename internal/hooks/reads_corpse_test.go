package hooks

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
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

// ratCorpses puts two rat corpses in room A: the older (round 5) holds 1
// gold, the newer (round 7) 2.
func ratCorpses(room *rooms.Room) {
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 1})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 7, Gold: 2})
}

// One numbering per name, the living first (Calabe, 2026-11-11). Room A
// holds the live rat, a live player Ratty and two rat corpses. Players and
// NPCs are numbered apart (FindByName) and a command takes the player when
// both match, so 1 reaches the living (Ratty, then the rat) and the corpses
// start at 2: 2 is the newest, 3 the older. A player's name takes no "The".
// The person typing is not in their own name's count: Ann's "ann" is her
// newest corpse; Bob's "ann" is Ann.
func TestCorpseNumbersCountPastTheLiving(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, bob, _, _ := w.trio()
	w.addPlayer("Ratty", readsRoomA)
	room := rooms.LoadRoom(readsRoomA)
	ratCorpses(room)
	room.AddCorpse(rooms.Corpse{UserId: ann.UserId, Character: characters.Character{Name: "Ann"}, RoundCreated: 6, Gold: 3})

	c, refusal := room.FindCorpseByRef(ann.Character, "rat")
	assert.Nil(t, c)
	assert.Equal(t, "Ratty is still alive.", refusal, "1 is the living; the player when a player and an NPC both match")
	for search, gold := range map[string]int{"2.rat": 2, "rat#2": 2, "3.rat": 1, "rat corpse": 2, "ann": 3} {
		c, refusal := room.FindCorpseByRef(ann.Character, search)
		require.NotNil(t, c, search)
		assert.Empty(t, refusal, search)
		assert.Equal(t, gold, c.Gold, search)
	}
	c, refusal = room.FindCorpseByRef(ann.Character, "4.rat")
	assert.Nil(t, c, "two rat corpses only")
	assert.Empty(t, refusal)

	c, refusal = room.FindCorpseByRef(bob.Character, "ann")
	assert.Nil(t, c)
	assert.Equal(t, "Ann is still alive.", refusal, "to Bob, Ann is alive and first")
	c, _ = room.FindCorpseByRef(bob.Character, "2.ann")
	require.NotNil(t, c)
	assert.Equal(t, 3, c.Gold)

	c, refusal = room.FindCorpseByRef(ann.Character, "bob")
	assert.Nil(t, c, "no corpse answers to bob: nothing, not a refusal")
	assert.Empty(t, refusal)
}

// A hidden body takes no ordinal in a corpse lookup either, and nothing
// says it is there; a viewer with see-hidden counts it.
func TestHiddenBodiesTakeNoCorpseOrdinal(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, _, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	ratCorpses(room)
	w.hide(&rat.Character)

	c, refusal := room.FindCorpseByRef(ann.Character, "rat")
	require.NotNil(t, c, "the hidden rat takes no ordinal: 1 is the newest corpse")
	assert.Empty(t, refusal, "and nothing says it is there")
	assert.Equal(t, 2, c.Gold)

	w.flag(ann.Character, readsSeeHiddenBuf, "see-hidden")
	c, refusal = room.FindCorpseByRef(ann.Character, "rat")
	assert.Nil(t, c)
	assert.Equal(t, "The rat is still alive.", refusal)
}

// bury numbers as the lookup does: an ordinal on the living is refused, not
// skipped; "2.rat" buries the newest rat corpse, here the later of two from
// the same round (which RemoveCorpse, matching owner, name and round, would
// mistake for the first).
func TestBuryBuriesTheCorpseItNames(t *testing.T) {
	w := newReadsWorld(t, 1)
	w.trio() // the rat is alive in room A
	room := rooms.LoadRoom(readsRoomA)
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 1})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 2})

	w.run("Ann", "bury rat")
	require.Len(t, room.Corpses, 2, "the live rat is not a corpse")
	w.run("Ann", "bury 2.rat")

	require.Len(t, room.Corpses, 1)
	require.Equal(t, 1, room.Corpses[0].Gold, "the newest rat corpse was buried, the older stays")
	out := w.out.String()
	require.Contains(t, out, `Ann: The rat is still alive.`)
	require.Contains(t, out, `Ann: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`)
}

// Bob looking at "bob" sees his newest corpse when one is here (the person
// typing is not in their own name's count for a corpse); with none, he looks
// at himself, as before.
func TestLookAtYourOwnNameFindsYourCorpse(t *testing.T) {
	w := newReadsWorld(t, 1)
	_, bob, _, _ := w.trio()
	room := rooms.LoadRoom(readsRoomA)

	w.run("Bob", "look bob")
	before := w.out.String()
	require.NotContains(t, before, "Bob: You look at the", "no corpse yet: Bob looks at himself")

	room.AddCorpse(rooms.Corpse{UserId: bob.UserId, Character: characters.Character{Name: "Bob"}, RoundCreated: 5})
	w.run("Bob", "look bob")
	require.Contains(t, w.out.String()[len(before):], `Bob: You look at the <ansi fg="user-corpse">Bob corpse</ansi>.`)
}

// bury refuses a player's corpse that holds any item or gold, whoever asks,
// its owner too: no single command destroys a player's belongings (Calabe,
// 2026-11-11). An empty player corpse, and an NPC corpse with its loot,
// bury as before.
func TestBuryRefusesAPlayersCorpseThatHoldsAnything(t *testing.T) {
	w := newReadsWorld(t, 1)
	ann, _, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	room.RemoveMob(rat.Character.Id) // no live rat: "rat" is the corpse
	room.AddCorpse(rooms.Corpse{UserId: ann.UserId, Character: characters.Character{Name: "Ann"}, Gold: 5})
	room.AddCorpse(rooms.Corpse{UserId: 9398, Character: characters.Character{Name: "Zed"}, Items: []items.Item{items.New(1)}})
	room.AddCorpse(rooms.Corpse{UserId: 9399, Character: characters.Character{Name: "Yan"}})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, Gold: 25})

	w.run("Ann", "bury ann") // her own name reaches her corpse (Task 5b)
	w.run("Ann", "bury zed")
	w.run("Ann", "bury yan")
	w.run("Ann", "bury rat")

	out := w.out.String()
	require.Contains(t, out, `Ann: You can't bury the <ansi fg="user-corpse">Ann corpse</ansi> while it still holds something.`, "her own, holding gold")
	require.Contains(t, out, `Ann: You can't bury the <ansi fg="user-corpse">Zed corpse</ansi> while it still holds something.`, "another's, holding an item")
	require.Contains(t, out, `Ann: You bury the <ansi fg="user-corpse">Yan corpse</ansi>.`, "an empty player corpse")
	require.Contains(t, out, `Ann: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`, "an NPC corpse, loot and all")
	left := []string{}
	for _, c := range room.Corpses {
		left = append(left, c.Character.Name)
	}
	require.Equal(t, []string{"Ann", "Zed"}, left, "the two holding something are still there")
}

// Two players bury in turn: the second bury resolves its name against what
// is left, so it takes the next match or reads the not-found line. A corpse
// that has crumbled (Prunable) is never found. Commands and the round tick
// run one at a time (MainWorker, util.LockMud), so this order is the only one.
func TestBuryAfterTheCorpseIsGone(t *testing.T) {
	w := newReadsWorld(t, 1)
	_, _, _, rat := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	room.RemoveMob(rat.Character.Id) // no live rat: "rat" is a corpse
	room.AddCorpse(rooms.Corpse{UserId: 9399, Character: characters.Character{Name: "Yan"}})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 3})
	room.AddCorpse(rooms.Corpse{MobId: 8, Character: characters.Character{Name: "rat"}, RoundCreated: 5})
	room.AddCorpse(rooms.Corpse{UserId: 9398, Character: characters.Character{Name: "Xan"}, Prunable: true})

	w.run("Ann", "bury yan")
	w.run("Bob", "bury yan") // already gone
	w.run("Ann", "bury rat") // the newest rat (round 5)
	w.run("Bob", "bury rat") // the other rat (round 3)
	w.run("Ann", "bury rat") // none left
	w.run("Bob", "bury xan") // crumbled, never a candidate

	out := w.out.String()
	require.Contains(t, out, `Ann: You bury the <ansi fg="user-corpse">Yan corpse</ansi>.`)
	require.Contains(t, out, `Bob: You don't see a yan around for burying.`, "the second bury finds it gone")
	require.Contains(t, out, `Ann: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`)
	require.Contains(t, out, `Bob: You bury the <ansi fg="mob-corpse">rat corpse</ansi>.`, "the next rat, resolved afresh")
	require.Contains(t, out, `Ann: You don't see a rat around for burying.`, "no rat left")
	require.Contains(t, out, `Bob: You don't see a xan around for burying.`, "a crumbled corpse is not found")
	left := []string{}
	for _, c := range room.Corpses {
		left = append(left, c.Character.Name)
	}
	require.Equal(t, []string{"Xan"}, left, "only the crumbled corpse remains, for UpdateCorpses to remove")
}
