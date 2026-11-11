package usercommands

import (
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const hiddenTestBuffId = 92910

// hideBody gives a the hidden flag through a test buff.
func hideBody(t *testing.T, a *actors.Actor) {
	t.Helper()
	t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: hiddenTestBuffId, Name: "test hidden", TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{"hidden"},
	}))
	if a.Character.Buffs.List == nil { // a bare test NPC has no buff maps yet
		a.Character.Buffs = buffs.New()
	}
	require.NoError(t, a.Character.AddBuff(hiddenTestBuffId, true))
	require.True(t, a.Character.HasBuffFlag("hidden"))
}

// Walking in, a player is told only of the hostile NPCs it perceives: a
// hidden one is not named ("notices you"), unless the player is an admin.
func TestEnteringNamesOnlyTheHostilesTheMoverPerceives(t *testing.T) {
	_, player, here := goSetup(t)
	lurker := mobs.NewTestInstance(mobs.Mob{Hostile: true})
	lurker.Character.Name = "Lurker"
	lurker.Character.Health = 10
	id := lurker.Character.Id
	t.Cleanup(func() { mobs.Rest(id) })
	rooms.LoadRoom(testNextRoomId).AddMob(id)
	hideBody(t, actors.Get(id))

	walk := func(admin bool) string {
		player.Controller().Account().Role = accounts.RoleUser
		if admin {
			player.Controller().Account().Role = accounts.RoleAdmin
		}
		told := playerTexts(t, player, func() {
			_, err := Go("north", player, here, 0)
			require.NoError(t, err)
		})
		_, err := Go("south", player, rooms.LoadRoom(testNextRoomId), 0)
		require.NoError(t, err)
		player.Character.ActionPoints = 100
		return strings.Join(told, "")
	}
	assert.NotContains(t, walk(false), "Lurker", "the hidden Lurker is not named")
	assert.Contains(t, walk(true), "Lurker", "an admin perceives it (the control: the line exists)")
}

// An untargeted attack never picks a body the attacker does not perceive
// (the one visibility rule): the hidden Rat fighting the player is not
// picked, nor picked at random; an admin picks it.
func TestUntargetedAttackPicksOnlyWhatTheAttackerPerceives(t *testing.T) {
	npc, player, room := goSetup(t)
	user := player.Controller().Account()
	accounts.SetTestUser(user) // the room's fight lists find the player's account by body
	t.Cleanup(func() { accounts.RemoveTestUser(user.UserId) })
	npc.Character.SetAggro(player.Id, characters.DefaultAttack)
	hideBody(t, npc)

	for _, rest := range []string{``, `*`, `*mob`} {
		told := playerTexts(t, player, func() {
			_, err := Attack(rest, player, room, 0)
			require.NoError(t, err)
		})
		assert.Nil(t, player.Character.Aggro, "%q: no target picked", rest)
		if rest == `` {
			assert.Equal(t, []string{"You attack the darkness!\n"}, told)
		}
	}

	// The control: the same pick with the Rat perceived (an admin perceives
	// it) does attack it, so the refusals above are the rule's.
	user.Role = accounts.RoleAdmin
	_, err := Attack(``, player, room, 0)
	require.NoError(t, err)
	require.NotNil(t, player.Character.Aggro, "an admin perceives the Rat and attacks it")
	assert.Equal(t, npc.Id, player.Character.Aggro.Target)
}
