package actors

import (
	"os"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	mudlog.SetupLogger(nil, "", "", false)
	Init()
	os.Exit(m.Run())
}

func TestAccountOnlineAndOffline(t *testing.T) {
	Reset()
	u := &accounts.Account{UserId: 31, Character: characters.New()}

	accounts.OnActivate.Fire(u)
	a := ByUserId(31)
	require.NotNil(t, a)
	assert.Same(t, u.Character, a.Character)

	accounts.OnActivate.Fire(u) // link-dead reconnect: same record
	assert.Same(t, a, ByUserId(31))
	assert.Len(t, All(), 1)

	again := &accounts.Account{UserId: 31, Character: characters.New()} // copyover: new record
	accounts.OnActivate.Fire(again)
	b := ByUserId(31)
	assert.Equal(t, a.Id, b.Id, "the same account keeps its actor")
	assert.Same(t, again, b.Controller().Account())
	assert.Same(t, again.Character, b.Character)

	accounts.OnDeactivate.Fire(again)
	assert.Nil(t, ByUserId(31))
	assert.Empty(t, All())
}

func TestMobSpawnAndDestroy(t *testing.T) {
	Reset()
	m := bodyMob()

	mobs.OnInstanceCreated.Fire(m)
	a := ByMobInstanceId(m.Character.Id)
	require.NotNil(t, a)
	assert.Same(t, &m.Character, a.Character)
	assert.Nil(t, a.Controller().Account())

	mobs.OnInstanceDestroyed.Fire(m.Character.Id)
	assert.Nil(t, ByMobInstanceId(m.Character.Id))
	assert.Empty(t, All())
}

func TestStaleLogoutDoesNotRemoveTheLiveActor(t *testing.T) {
	Reset()
	old := &accounts.Account{UserId: 71, Character: characters.New()}
	accounts.OnActivate.Fire(old)
	id := ByUserId(71).Id

	fresh := &accounts.Account{UserId: 71, Character: characters.New()} // a new login of the same account
	accounts.OnActivate.Fire(fresh)
	accounts.OnDeactivate.Fire(old) // the old record's logout lands late

	a := ByUserId(71)
	require.NotNil(t, a, "the live login keeps its actor")
	assert.Equal(t, id, a.Id)
	assert.Same(t, fresh, a.Controller().Account())

	accounts.OnDeactivate.Fire(fresh)
	assert.Nil(t, ByUserId(71))
	assert.Empty(t, All())
}

func TestInitIsIdempotent(t *testing.T) {
	Reset()
	Init()
	Init()
	mobs.OnInstanceCreated.Fire(bodyMob())
	assert.Len(t, All(), 1, "a second Init must not double-register")
}
