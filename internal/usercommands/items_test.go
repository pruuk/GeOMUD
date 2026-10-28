package usercommands

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const permaGearBuffId = 92901

// testItem is an item with its own spec, so no item data files are needed.
func testItem(itemId int, name string, spec items.ItemSpec) items.Item {
	spec.ItemId = itemId
	spec.Name = name
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

func testSword() items.Item {
	// Shooting, so HandsRequired does not look up race data the test has not loaded.
	return testItem(92701, "testsword", items.ItemSpec{Type: items.Weapon, Subtype: items.Shooting, Hands: items.OneHanded})
}

func testRock() items.Item {
	return testItem(92702, "testrock", items.ItemSpec{Type: items.Object})
}

// givePermaGear gives the actor a buff carrying the perma-gear flag.
func givePermaGear(t *testing.T, a *actors.Actor) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: permaGearBuffId, Name: "test perma-gear", TriggerRate: "1 round", TriggerCount: 5,
		Flags: []string{"perma-gear"},
	}))
	a.Character.Buffs = buffs.New() // a bare test NPC has no buff maps yet
	require.NoError(t, a.Character.AddBuff(permaGearBuffId, true))
	require.True(t, a.Character.HasBuffFlag("perma-gear"))
}

// capture records every event of one type queued while fn runs.
func capture[T events.Event](t *testing.T, fn func()) []T {
	t.Helper()
	var got []T
	var zero T
	id := events.RegisterListener(zero, func(e events.Event) events.ListenerReturn {
		if v, ok := e.(T); ok {
			got = append(got, v)
		}
		return events.Continue
	})
	defer events.UnregisterListener(zero, id)
	events.ProcessEvents()
	got = nil
	fn()
	events.ProcessEvents()
	return got
}

func inputTexts(in []events.Input) []string {
	out := []string{}
	for _, i := range in {
		out = append(out, i.InputText)
	}
	return out
}

func countIn(list []items.Item, it items.Item) int {
	n := 0
	for _, x := range list {
		if x.Equals(it) {
			n++
		}
	}
	return n
}

// The @ prefix keeps Emote from echoing the line back to the actor.
const struggle = `emote @struggles with their gear for a while, then gives up.`

func TestPermaGearStruggleEmoteHasNoSelfEcho(t *testing.T) {
	_, player, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	givePermaGear(t, player)

	var queued []events.Input
	told := playerTexts(t, player, func() {
		queued = capture[events.Input](t, func() {
			require.True(t, permaGearRefused(player))
		})
	})
	require.Equal(t, []string{struggle}, inputTexts(queued))
	require.Len(t, told, 1)
	assert.Contains(t, told[0], `Your gear won't budge.`)

	// Run the queued line as the input path would.
	rest := strings.TrimPrefix(queued[0].InputText, "emote ")
	msgs := capture[events.Message](t, func() {
		_, err := Emote(rest, player, room, 0)
		require.NoError(t, err)
	})
	for _, m := range msgs {
		if m.RoomId == 0 {
			assert.NotContains(t, m.Text, "You Emote:")
			assert.NotEqual(t, player.Controller().Account().UserId, m.UserId, "the player is told nothing")
		}
	}
	room_ := messagesFor(msgs, testRoomId)
	require.NotEmpty(t, room_, "the room still sees it")
	assert.Contains(t, room_[0], "struggles with their gear")
}

func TestNPCGetWithFullBackpackKeepsTheItem(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	for i := 0; i < npc.Character.CarryCapacity(); i++ {
		npc.Character.StoreItem(testItem(92710+i, "filler", items.ItemSpec{Type: items.Object}))
	}
	rock := testRock()
	room.AddItem(rock, false)

	_, err := Get("testrock", npc, room, 0)
	require.NoError(t, err)

	// The mob command removed the item from the floor before it checked the
	// backpack took it. Wherever the item ends up, it must exist exactly once.
	assert.Equal(t, 1, countIn(room.Items, rock)+countIn(npc.Character.Items, rock))
}

func TestNPCPutGoldItLacksIsRefused(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	room.Containers = map[string]rooms.Container{"chest": {}}
	npc.Character.Gold = 3

	_, err := Put("50 gold chest", npc, room, 0)
	require.NoError(t, err)

	assert.Equal(t, 0, room.Containers["chest"].Gold, "no gold minted into the container")
	assert.Equal(t, 3, npc.Character.Gold)
}

func TestPermaGearBlocksItemDropForPlayer(t *testing.T) {
	_, player, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	rock := testRock()
	player.Character.StoreItem(rock)
	givePermaGear(t, player)

	var queued []events.Input
	queued = capture[events.Input](t, func() {
		_, err := Drop("testrock", player, room, 0)
		require.NoError(t, err)
	})

	assert.Equal(t, 1, countIn(player.Character.Items, rock), "item stays in the backpack")
	assert.Empty(t, room.Items)
	assert.Equal(t, []string{struggle}, inputTexts(queued), "the room sees the struggle emote")
}

func TestPermaGearBlocksItemDropForNPCButNotGold(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	rock := testRock()
	npc.Character.StoreItem(rock)
	npc.Character.Gold = 10
	givePermaGear(t, npc)

	_, err := Drop("4 gold", npc, room, 0)
	require.NoError(t, err)
	assert.Equal(t, 4, room.Gold, "gold still drops")

	_, err = Drop("testrock", npc, room, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, countIn(npc.Character.Items, rock))
	assert.Empty(t, room.Items)
}

func TestPermaGearDropAllRefusesItemsOnce(t *testing.T) {
	_, player, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	player.Character.StoreItem(testRock())
	player.Character.StoreItem(testItem(92703, "testpebble", items.ItemSpec{Type: items.Object}))
	player.Character.Gold = 5
	givePermaGear(t, player)

	queued := capture[events.Input](t, func() {
		_, err := Drop("all", player, room, 0)
		require.NoError(t, err)
	})

	assert.Equal(t, 5, room.Gold)
	assert.Len(t, player.Character.Items, 2)
	assert.Equal(t, []string{struggle}, inputTexts(queued))
}

func TestPermaGearBlocksEquipAndRemove(t *testing.T) {
	for name, who := range map[string]func(npc, player *actors.Actor) *actors.Actor{
		"npc":    func(npc, _ *actors.Actor) *actors.Actor { return npc },
		"player": func(_, player *actors.Actor) *actors.Actor { return player },
	} {
		t.Run(name, func(t *testing.T) {
			npc, player, _ := setup(t)
			a := who(npc, player)
			room := rooms.LoadRoom(testRoomId)
			sword := testSword()
			a.Character.StoreItem(sword)
			givePermaGear(t, a)

			queued := capture[events.Input](t, func() {
				_, err := Equip("testsword", a, room, 0)
				require.NoError(t, err)
			})
			assert.Equal(t, 1, countIn(a.Character.Items, sword))
			assert.Equal(t, 0, a.Character.Equipment.Weapon.ItemId)
			assert.Equal(t, []string{struggle}, inputTexts(queued))

			// A worn item stays on.
			a.Character.Equipment.Weapon = testItem(92704, "wornblade", items.ItemSpec{Type: items.Weapon, Subtype: items.Shooting, Hands: items.OneHanded})
			queued = capture[events.Input](t, func() {
				_, err := Remove("wornblade", a, room, 0)
				require.NoError(t, err)
			})
			assert.Equal(t, 92704, a.Character.Equipment.Weapon.ItemId)
			assert.Equal(t, []string{struggle}, inputTexts(queued))
		})
	}
}

func TestReEquipWornItemIsSilentForAnyActor(t *testing.T) {
	for name, who := range map[string]func(npc, player *actors.Actor) *actors.Actor{
		"npc":    func(npc, _ *actors.Actor) *actors.Actor { return npc },
		"player": func(_, player *actors.Actor) *actors.Actor { return player },
	} {
		t.Run(name, func(t *testing.T) {
			npc, player, _ := setup(t)
			a := who(npc, player)
			room := rooms.LoadRoom(testRoomId)
			sword := testSword()
			a.Character.Equipment.Weapon = sword
			a.Character.StoreItem(sword) // the same item, also in the backpack

			var changes []events.EquipmentChange
			var msgs []events.Message
			changes = capture[events.EquipmentChange](t, func() {
				msgs = captureMessages(t, func() {
					_, err := Equip("testsword", a, room, 0)
					require.NoError(t, err)
				})
			})

			assert.Empty(t, changes, "no EquipmentChange")
			assert.Empty(t, messagesFor(msgs, testRoomId), "no room text")
			assert.True(t, a.Character.Equipment.Weapon.Equals(sword), "still wielded")
			assert.Equal(t, 1, countIn(a.Character.Items, sword), "backpack copy kept")
		})
	}
}

func TestEquipNewItemStillAnnouncesWithMobname(t *testing.T) {
	npc, _, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	npc.Character.Name = "Rat"
	sword := testSword()
	npc.Character.StoreItem(sword)

	var changes []events.EquipmentChange
	var msgs []events.Message
	changes = capture[events.EquipmentChange](t, func() {
		msgs = captureMessages(t, func() {
			_, err := Equip("testsword", npc, room, 0)
			require.NoError(t, err)
		})
	})

	assert.Len(t, changes, 1)
	require.Len(t, messagesFor(msgs, testRoomId), 1)
	text := messagesFor(msgs, testRoomId)[0]
	assert.True(t, strings.Contains(text, `<ansi fg="mobname">Rat</ansi> wields their`), text)
}

func TestGiveToPlayerWithoutAccountRecordDoesNotPanic(t *testing.T) {
	npc, player, _ := setup(t)
	room := rooms.LoadRoom(testRoomId)
	room.AddPlayer(player.Controller().Account().UserId)
	rock := testRock()
	npc.Character.StoreItem(rock)

	// The player is in the room by ID but has no record in the account manager.
	assert.NotPanics(t, func() {
		_, err := Give("testrock "+player.Id.String(), npc, room, 0)
		require.NoError(t, err)
	})
	assert.Equal(t, 1, countIn(npc.Character.Items, rock), "nothing moved")
}
