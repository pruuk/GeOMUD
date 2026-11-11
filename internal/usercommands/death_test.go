package usercommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/death"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recoveryRoom makes roomId the death recovery room for one test, with its
// template in a temporary data folder (a death moves the body into a
// private copy made from it), and returns the ID of a copy of it. Every
// copy of the room is freed afterwards (freeCopies).
func recoveryRoom(t *testing.T, roomId int) (copyId int) {
	t.Helper()
	oldData := configs.GetFilePathsConfig().DataFiles.String()
	oldRecovery := int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)
	tmp := t.TempDir()
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["FilePaths.DataFiles"] = tmp
	next["SpecialRooms.DeathRecoveryRoom"] = roomId
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["FilePaths.DataFiles"] = oldData
		back["SpecialRooms.DeathRecoveryRoom"] = oldRecovery
		require.NoError(t, configs.RestoreOverrides(back))
	})
	r := rooms.LoadRoom(roomId)
	require.NotNil(t, r)
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "rooms", filepath.FromSlash(rooms.ZoneToFolder(r.Zone))), 0o755))
	require.NoError(t, rooms.SaveRoomTemplate(*r))
	copies, err := rooms.CreateEphemeralRoomIds(roomId)
	require.NoError(t, err)
	copyId = copies[roomId]
	require.True(t, rooms.IsEphemeralRoomId(copyId))
	t.Cleanup(func() { freeCopies(roomId) })
	return copyId
}

// freeCopies frees every private copy of roomId. Each copy holds one of the
// 100 ephemeral chunks (rooms/ephemeral.go) until freed, and a chunk is
// freed only when no player is in it, so the players leave first.
func freeCopies(roomId int) {
	for _, id := range rooms.FindEphemeralRoomIds(roomId) {
		if r := rooms.LoadRoom(id); r != nil {
			for _, userId := range r.GetPlayers() {
				r.RemovePlayer(userId)
			}
		}
		rooms.TryEphemeralCleanup(id)
	}
}

// corpsesOn turns corpses on for one test (the test binary loads no
// config.yaml, so they are off), as setPvP sets PvP.
func corpsesOn(t *testing.T) {
	t.Helper()
	old := bool(configs.GetGamePlayConfig().Death.CorpsesEnabled)
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	next["GamePlay.Death.CorpsesEnabled"] = true
	require.NoError(t, configs.RestoreOverrides(next))
	require.True(t, bool(configs.GetGamePlayConfig().Death.CorpsesEnabled), "the override took")
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		back["GamePlay.Death.CorpsesEnabled"] = old
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

func itemIdsOf(list []items.Item) []int {
	out := []int{}
	for _, it := range list {
		out = append(out, it.ItemId)
	}
	return out
}

// A player's death leaves a corpse found by the player's name (it records
// the account: Corpse.UserId) holding what the player carried and its gold.
// "get all <name>" and "get all from <name>" take it all (spec, plan review
// round 1); "get <item> <name>" takes one item.
func TestGetFromAPlayersCorpse(t *testing.T) {
	for _, tc := range []struct {
		command   string
		wantItems []int
		wantGold  int
	}{
		{"all bob", []int{92730}, 25},
		{"all from bob", []int{92730}, 25},
		{"testgem bob", []int{92730}, 0},
	} {
		t.Run(tc.command, func(t *testing.T) {
			npc, player, room := goSetup(t)
			recoveryRoom(t, testNextRoomId)
			corpsesOn(t)
			player.Character.Name = "Bob"
			player.Character.Items = []items.Item{testItem(92730, "testgem", items.ItemSpec{Type: items.Gemstone})}
			require.Equal(t, 25, player.Character.Gold, "a new character's gold")
			death.Die(player)
			require.Len(t, room.Corpses, 1, "Bob's corpse is in the room")
			require.Equal(t, 9201, room.Corpses[0].UserId, "a player's corpse records the account")

			_, err := Get(tc.command, npc, room, 0)
			require.NoError(t, err)
			assert.Equal(t, tc.wantItems, itemIdsOf(npc.Character.Items), "what the taker carries")
			assert.Equal(t, tc.wantGold, npc.Character.Gold, "the taker's gold")
			assert.Empty(t, room.Corpses[0].Items, "the item left the corpse")
		})
	}
}

// Any other "get all ..." keeps today's room-wide meaning (spec, plan review
// round 1): a word that names no corpse, or more words than "all [from]
// <corpse>" ("all junk bob" names Bob's corpse last), takes everything on
// the floor and leaves Bob's corpse alone. So does a word that also names a
// container ("bob" begins "bobchest"): the container wins, as for every get.
func TestGetAllWithoutACorpseTakesTheFloor(t *testing.T) {
	for _, tc := range []struct{ command, container string }{
		{"all ground", ""}, {"all junk", ""}, {"all from junk", ""}, {"all junk bob", ""},
		{"all bob", "bobchest"},
	} {
		t.Run(tc.command+" "+tc.container, func(t *testing.T) {
			npc, player, room := goSetup(t)
			recoveryRoom(t, testNextRoomId)
			corpsesOn(t)
			player.Character.Name = "Bob"
			player.Character.Items = []items.Item{testItem(92730, "testgem", items.ItemSpec{Type: items.Gemstone})}
			death.Die(player)
			require.Len(t, room.Corpses, 1, "Bob's corpse is in the room")
			room.AddItem(testItem(92732, "teststick", items.ItemSpec{Type: items.Object}), false)
			room.Gold = 5
			if tc.container != "" {
				room.Containers = map[string]rooms.Container{tc.container: {}}
				t.Cleanup(func() { room.Containers = nil })
			}

			_, err := Get(tc.command, npc, room, 0)
			require.NoError(t, err)
			assert.Equal(t, []int{92732}, itemIdsOf(npc.Character.Items), "the floor's item, not the corpse's")
			assert.Equal(t, 5, npc.Character.Gold, "the floor's gold, not the corpse's")
			assert.Empty(t, room.Items, "the floor is empty")
			assert.Equal(t, []int{92730}, itemIdsOf(room.Corpses[0].Items), "the corpse is untouched")
			assert.Equal(t, 25, room.Corpses[0].Gold)
		})
	}
}

// The corpse panel shows what a corpse holds (look no longer gates it on a
// config setting).
func TestLookAtACorpseShowsWhatItHolds(t *testing.T) {
	npc, player, room := goSetup(t)
	corpsesOn(t)
	mobs.GetInstance(npc.Id).MobId = 92801 // a real NPC's template ID, which its corpse records (death.Die)
	npc.Character.Items = []items.Item{testItem(92731, "testpebble", items.ItemSpec{Type: items.Gemstone})}
	death.Die(npc)
	require.Len(t, room.Corpses, 1)

	told := playerTexts(t, player, func() {
		_, err := Look("rat corpse", player, room, 0)
		require.NoError(t, err)
	})
	assert.Contains(t, strings.Join(told, ""), "testpebble", "the panel lists the corpse's item")
}

// A floor item named by the whole phrase wins over a corpse its last word
// names (whole-branch review M1): "get iron key" beside a Keymaster corpse
// picks up the floor's iron key, with or without a living Keymaster in the
// room. A phrase no floor item answers to still reaches the corpse.
func TestGetFloorItemWhoseLastWordNamesACorpse(t *testing.T) {
	for _, living := range []bool{false, true} {
		t.Run(map[bool]string{false: "no living keymaster", true: "living keymaster"}[living], func(t *testing.T) {
			npc, player, room := goSetup(t)
			if living {
				npc.Character.Name = "Keymaster"
			}
			room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "Keymaster"}, RoundCreated: 5,
				Items: []items.Item{testItem(92733, "testgem", items.ItemSpec{Type: items.Gemstone})}})
			room.AddItem(testItem(92734, "iron key", items.ItemSpec{Type: items.Object}), false)

			_, err := Get("iron key", player, room, 0)
			require.NoError(t, err)
			assert.Equal(t, []int{92734}, itemIdsOf(player.Character.Items), "the floor's iron key")
			assert.Empty(t, room.Items, "the floor is empty")
			assert.Equal(t, []int{92733}, itemIdsOf(room.Corpses[0].Items), "the corpse is untouched")

			if !living {
				_, err = Get("testgem keymaster", player, room, 0)
				require.NoError(t, err)
				assert.Equal(t, []int{92734, 92733}, itemIdsOf(player.Character.Items), "the corpse's item")
			}
		})
	}
}

// One numbering per name through get (Task 5b): goSetup's live NPC "Rat"
// is 1, so "get all rat" is refused rather than skipped ahead; 2 is the
// newest rat corpse and 3 the older, in both ordinal forms.
func TestGetAllCountsCorpsesPastTheLiving(t *testing.T) {
	_, player, room := goSetup(t)
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 5, Gold: 1})
	room.AddCorpse(rooms.Corpse{MobId: 7, Character: characters.Character{Name: "rat"}, RoundCreated: 7, Gold: 2})
	gold := player.Character.Gold

	told := playerTexts(t, player, func() {
		_, err := Get("all rat", player, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, []string{"The Rat is still alive.\n"}, told, "1 is the living Rat")
	assert.Equal(t, gold, player.Character.Gold, "nothing taken")

	_, err := Get("all 2.rat", player, room, 0)
	require.NoError(t, err)
	assert.Equal(t, gold+2, player.Character.Gold, "2 is the newest corpse")
	_, err = Get("all rat#3", player, room, 0)
	require.NoError(t, err)
	assert.Equal(t, gold+3, player.Character.Gold, "3 is the older")
}
