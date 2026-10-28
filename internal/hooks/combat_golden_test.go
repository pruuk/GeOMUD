package hooks

// The golden combat test: the parity tool for Phase 4b slice 5 (plan
// Decision 4). Each fixture builds a small fight from scratch, seeds the
// shared random source, runs DoCombat for a fixed number of rounds, and
// records, round by round, what each player read and what state changed. The
// record is compared with testdata/combat-golden/<fixture>.txt; run with
// -update to rewrite the files.
//
// Room messages are expanded per receiver the way Message_SendMessage
// delivers them (a UserId message to that user; a RoomId message to every
// player in the room at delivery time, minus the UserId and ExcludeUserIds,
// honouring IsQuiet), so the record is "what each player read". Text is kept
// as markup (before AnsiParse), so a colour change shows. Queued commands
// (events.Input) are recorded, not run: a "suicide" in the record means the
// round queued it.

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/require"
)

var updateCombatGolden = flag.Bool("update", false, "rewrite testdata/combat-golden/*.txt from this run")

const (
	goldenRoomA     = 93001
	goldenRoomB     = 93002
	goldenHiddenBuf = 93901
)

var (
	loadCombatData sync.Once
	goldenRun      int
)

// combatWorld is one fixture's actors, rooms and record.
type combatWorld struct {
	t       *testing.T
	players []*accounts.Account // record order
	npcs    []*mobs.Mob         // record order
	names   map[ids.ID]string
	users   map[int]string
	groups  []string // NPC groups whose hatred of each player is recorded

	// what the current round produced
	reads  map[int][]string
	events []string

	out strings.Builder
}

func newCombatWorld(t *testing.T, seed int64) *combatWorld {
	t.Helper()
	loadCombatData.Do(func() {
		// Attack message text, the broken-item spec and race dice come from the
		// empty world's data, so a change to that data shows in the goldens.
		races.LoadDataFiles()
		items.LoadDataFiles()
	})

	actors.Reset()
	accounts.ResetActiveUsers()
	t.Cleanup(func() {
		accounts.ResetActiveUsers()
		actors.Reset()
	})
	t.Cleanup(rooms.AddTestDefaultBiome()) // a successful flee ends in a look

	a := rooms.NewEmptyRoom()
	a.RoomId, a.Zone, a.Title = goldenRoomA, "goldenzone", "Room A"
	a.Exits = map[string]exit.RoomExit{"north": {RoomId: goldenRoomB}}
	b := rooms.NewEmptyRoom()
	b.RoomId, b.Zone, b.Title = goldenRoomB, "goldenzone", "Room B"
	b.Exits = map[string]exit.RoomExit{"south": {RoomId: goldenRoomA}}
	rooms.AddTestRoom(a)
	rooms.AddTestRoom(b)
	t.Cleanup(func() {
		rooms.RemoveTestRoom(goldenRoomA)
		rooms.RemoveTestRoom(goldenRoomB)
	})

	w := &combatWorld{
		t:     t,
		names: map[ids.ID]string{},
		users: map[int]string{},
	}

	listeners := []events.ListenerId{}
	listeners = append(listeners,
		events.RegisterListener(events.Message{}, w.onMessage),
		events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
			in := e.(events.Input)
			w.events = append(w.events, fmt.Sprintf("Input %s: %q", w.name(in.ActorId), in.InputText))
			return events.Continue
		}),
		events.RegisterListener(events.AggroChanged{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.AggroChanged)
			w.events = append(w.events, fmt.Sprintf("AggroChanged %s room=%d", w.name(ev.ActorId), ev.RoomId))
			return events.Continue
		}),
		events.RegisterListener(events.PlayerDrop{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.PlayerDrop)
			w.events = append(w.events, fmt.Sprintf("PlayerDrop %s room=%d", w.user(ev.UserId), ev.RoomId))
			return events.Continue
		}),
		events.RegisterListener(events.CharacterVitalsChanged{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.CharacterVitalsChanged)
			w.events = append(w.events, fmt.Sprintf("CharacterVitalsChanged %s", w.user(ev.UserId)))
			return events.Continue
		}),
		events.RegisterListener(events.ItemOwnership{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.ItemOwnership)
			w.events = append(w.events, fmt.Sprintf("ItemOwnership %s item=%d gained=%v", w.name(ev.ActorId), ev.Item.ItemId, ev.Gained))
			return events.Continue
		}),
		events.RegisterListener(events.Buff{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.Buff)
			w.events = append(w.events, fmt.Sprintf("Buff %s buff=%d source=%s", w.name(ev.ActorId), ev.BuffId, ev.Source))
			return events.Continue
		}),
		events.RegisterListener(events.MobDeath{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.MobDeath)
			w.events = append(w.events, fmt.Sprintf("MobDeath %s", w.name(ev.ActorId)))
			return events.Continue
		}),
		events.RegisterListener(events.PlayerDeath{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.PlayerDeath)
			w.events = append(w.events, fmt.Sprintf("PlayerDeath %s", w.user(ev.UserId)))
			return events.Continue
		}),
	)
	t.Cleanup(func() {
		zeros := []events.Event{events.Message{}, events.Input{}, events.AggroChanged{}, events.PlayerDrop{},
			events.CharacterVitalsChanged{}, events.ItemOwnership{}, events.Buff{}, events.MobDeath{}, events.PlayerDeath{}}
		for i, z := range zeros {
			events.UnregisterListener(z, listeners[i])
		}
	})

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("record so far:\n%s", w.out.String())
		}
	})

	util.SeedRandForTest(seed)
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) }) // back on a clock seed, as production runs
	fmt.Fprintf(&w.out, "seed %d\n", seed)
	return w
}

func (w *combatWorld) name(id ids.ID) string {
	if n, ok := w.names[id]; ok {
		return n
	}
	if id.IsZero() {
		return "(none)"
	}
	return "(unknown)"
}

func (w *combatWorld) user(userId int) string {
	if n, ok := w.users[userId]; ok {
		return n
	}
	return fmt.Sprintf("(user %d)", userId)
}

// onMessage expands one message to its receivers as Message_SendMessage does.
func (w *combatWorld) onMessage(e events.Event) events.ListenerReturn {
	m := e.(events.Message)
	text := strings.TrimSuffix(m.Text, "\n")
	if n := strings.Count(text, "\n"); n > 0 {
		// Combat lines are one line each; a multi-line text is a rendered
		// screen (the look after a flee), recorded by its size only so the
		// goldens do not pin the room template.
		text = fmt.Sprintf("(multi-line text, %d lines)", n+1)
	}
	if m.UserId > 0 {
		if accounts.GetByUserId(m.UserId) != nil {
			w.reads[m.UserId] = append(w.reads[m.UserId], text)
		}
	}
	if m.RoomId > 0 {
		room := rooms.LoadRoom(m.RoomId)
		if room == nil {
			return events.Continue
		}
		for _, userId := range room.GetPlayers() {
			if userId == m.UserId {
				continue
			}
			skip := false
			for _, ex := range m.ExcludeUserIds {
				if ex == userId {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			u := accounts.GetByUserId(userId)
			if u == nil {
				continue
			}
			if m.IsQuiet && !u.Character.HasBuffFlag("superhearing") {
				continue
			}
			w.reads[userId] = append(w.reads[userId], text)
		}
	}
	return events.Continue
}

// statted gives a character fixed stats (as training on top of its race's
// base) and full health of hp. The test loads no progression config, so the
// health maximum is set directly; anything that runs Validate afterwards
// would reset it (hide restores it).
func statted(c *characters.Character, speed, strength, perception, hp int) {
	c.Stats.Speed.Training = speed
	c.Stats.Strength.Training = strength
	c.Stats.Perception.Training = perception
	c.Validate()
	c.HealthMax.Value = hp
	c.Health = hp
}

func (w *combatWorld) addPlayer(userId int, name string, roomId int, speed, strength, perception, hp int) *accounts.Account {
	w.t.Helper()
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = name
	c.RaceId = 1 // human
	c.RoomId = roomId
	c.Zone = "goldenzone"
	statted(c, speed, strength, perception, hp)
	u := &accounts.Account{UserId: userId, Character: c, Password: "$2a$test"}
	actors.Spawn(c, actors.NewConnectionController(u))
	accounts.SetTestUser(u)
	rooms.LoadRoom(roomId).AddPlayer(userId)
	w.players = append(w.players, u)
	w.names[c.Id] = name
	w.users[userId] = name
	return u
}

func (w *combatWorld) addNPC(mobId int, name string, raceId int, roomId int, speed, strength, perception, hp int, groups ...string) *mobs.Mob {
	w.t.Helper()
	// Group hatred is package state in mobs with no reset, so each run gets
	// its own group names; the record prints them without the suffix.
	goldenRun++
	for i, g := range groups {
		groups[i] = fmt.Sprintf("%s#%d", g, goldenRun)
	}
	m := mobs.Mob{MobId: mobs.MobId(mobId), HomeRoomId: roomId, Groups: groups}
	m.Character = *characters.New()
	m.Character.Name = name
	m.Character.RaceId = raceId
	inst := mobs.NewTestInstance(m)
	id := inst.Character.Id
	w.t.Cleanup(func() { mobs.Rest(id) })
	rooms.LoadRoom(roomId).AddMob(id)
	statted(&inst.Character, speed, strength, perception, hp)
	w.npcs = append(w.npcs, inst)
	w.names[id] = name
	w.groups = append(w.groups, groups...)
	return inst
}

// weapon is a weapon with its own spec, so no item file is needed for it.
func weapon(itemId int, name string, subtype items.ItemSubType, dice, sides, waitRounds int) items.Item {
	spec := items.ItemSpec{
		ItemId: itemId, Name: name, NameSimple: name, Type: items.Weapon, Subtype: subtype, Hands: items.OneHanded,
		WaitRounds: waitRounds, Damage: items.Damage{Attacks: 1, DiceCount: dice, SideCount: sides},
	}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

func shield(itemId int, name string, breakChance uint8) items.Item {
	spec := items.ItemSpec{
		ItemId: itemId, Name: name, NameSimple: name, Type: items.Offhand, Subtype: items.Wearable,
		DamageReduction: 10, BreakChance: breakChance,
	}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

func (w *combatWorld) hide(c *characters.Character) {
	w.t.Helper()
	w.t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: goldenHiddenBuf, Name: "test hidden", TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{"hidden"},
	}))
	hpMax, hp := c.HealthMax.Value, c.Health
	require.NoError(w.t, c.AddBuff(goldenHiddenBuf, true)) // runs Validate
	c.HealthMax.Value, c.Health = hpMax, hp
	require.True(w.t, c.HasBuffFlag("hidden"))
}

func (w *combatWorld) aggro(c *characters.Character) string {
	if c.Aggro == nil {
		return "none"
	}
	kind := map[characters.AggroType]string{
		characters.DefaultAttack: "DefaultAttack", characters.Shooting: "Shooting", characters.BackStab: "BackStab",
		characters.SpellCast: "SpellCast", characters.Flee: "Flee",
	}[c.Aggro.Type]
	s := fmt.Sprintf("%s %s wait=%d", w.name(c.Aggro.Target), kind, c.Aggro.RoundsWaiting)
	if c.Aggro.ExitName != "" {
		s += " exit=" + c.Aggro.ExitName
	}
	return s
}

func (w *combatWorld) damage(c *characters.Character) string {
	keys := []int{}
	for k := range c.PlayerDamage {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	parts := []string{}
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", w.user(k), c.PlayerDamage[k]))
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func (w *combatWorld) state(c *characters.Character) string {
	s := fmt.Sprintf("room=%d hp=%d/%d aggro=[%s] playerdamage=%s", c.RoomId, c.Health, c.HealthMax.Value, w.aggro(c), w.damage(c))
	if !c.KilledBy.IsZero() {
		s += fmt.Sprintf(" killedby=%s mob=%d elite=%t", w.name(c.KilledBy), c.KillerMobId, c.KillerMobIsElite)
	}
	if c.Equipment.Weapon.ItemId > 0 {
		s += fmt.Sprintf(" weapon=%d", c.Equipment.Weapon.ItemId)
	}
	if c.Equipment.Offhand.ItemId > 0 {
		s += fmt.Sprintf(" offhand=%d", c.Equipment.Offhand.ItemId)
	}
	if len(c.Items) > 0 {
		held := []string{}
		for _, it := range c.Items {
			held = append(held, fmt.Sprint(it.ItemId))
		}
		s += " items=" + strings.Join(held, ",")
	}
	return s
}

func (w *combatWorld) writeState() {
	fmt.Fprintf(&w.out, "-- state\n")
	for _, u := range w.players {
		fmt.Fprintf(&w.out, "  %s: %s\n", u.Character.Name, w.state(u.Character))
	}
	for _, m := range w.npcs {
		fmt.Fprintf(&w.out, "  %s: %s\n", m.Character.Name, w.state(&m.Character))
	}
	for _, g := range w.groups {
		for _, u := range w.players {
			if mobs.IsHostile(g, u.UserId) {
				fmt.Fprintf(&w.out, "  group %s hates %s\n", strings.SplitN(g, "#", 2)[0], u.Character.Name)
			}
		}
	}
	for _, rid := range []int{goldenRoomA, goldenRoomB} {
		if r := rooms.LoadRoom(rid); r != nil && len(r.Items) > 0 {
			held := []string{}
			for _, it := range r.Items {
				held = append(held, fmt.Sprint(it.ItemId))
			}
			fmt.Fprintf(&w.out, "  room %d floor: %s\n", rid, strings.Join(held, ","))
		}
	}
}

// start drops what setup queued and records the starting state.
func (w *combatWorld) start() {
	events.ProcessEvents()
	fmt.Fprintf(&w.out, "== start\n")
	w.writeState()
}

// fighting reports whether any fixture actor still has an aggro.
func (w *combatWorld) fighting() bool {
	for _, u := range w.players {
		if u.Character.Aggro != nil {
			return true
		}
	}
	for _, m := range w.npcs {
		if m.Character.Aggro != nil {
			return true
		}
	}
	return false
}

// rounds runs combat rounds until no fixture actor has an aggro, at most max.
func (w *combatWorld) rounds(max int) {
	for i := 1; i <= max && w.fighting(); i++ {
		w.reads = map[int][]string{}
		w.events = nil
		DoCombat(events.NewRound{})
		events.ProcessEvents()

		fmt.Fprintf(&w.out, "== round %d\n", i)
		for _, u := range w.players {
			fmt.Fprintf(&w.out, "-- %s reads\n", u.Character.Name)
			for _, line := range w.reads[u.UserId] {
				fmt.Fprintf(&w.out, "  %s\n", line)
			}
		}
		fmt.Fprintf(&w.out, "-- events\n")
		for _, e := range w.events {
			fmt.Fprintf(&w.out, "  %s\n", e)
		}
		w.writeState()
	}
}

// check compares the record with the golden file, or writes it with -update.
func (w *combatWorld) check(fixture string) {
	w.t.Helper()
	path := filepath.Join("testdata", "combat-golden", fixture+".txt")
	got := w.out.String()
	if *updateCombatGolden {
		require.NoError(w.t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(w.t, os.WriteFile(path, []byte(got), 0o644))
		return
	}
	wantBytes, err := os.ReadFile(path)
	require.NoError(w.t, err, "no golden for %s: run go test ./internal/hooks/ -run TestCombatGolden -update", fixture)
	want := strings.ReplaceAll(string(wantBytes), "\r\n", "\n")
	if got == want {
		return
	}
	gl, wl := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(gl) || i < len(wl); i++ {
		var g, x string
		if i < len(gl) {
			g = gl[i]
		}
		if i < len(wl) {
			x = wl[i]
		}
		if g != x {
			w.t.Fatalf("%s differs from the golden at line %d\n  want: %s\n  got:  %s", path, i+1, x, g)
		}
	}
}

// A player kills an NPC.
func TestCombatGoldenPvMToDeath(t *testing.T) {
	w := newCombatWorld(t, 111)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "golden-rats-pvm")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("pvm_death")
	require.Less(t, rat.Character.Health, 1, "the fixture must reach the rat's death")
}

// A player attacks a player who does not fight back; the defender's shield
// can break.
func TestCombatGoldenPvP(t *testing.T) {
	w := newCombatWorld(t, 202)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 15)
	bob.Character.Equipment.Offhand = shield(93102, "testshield", 90)
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("pvp")
	require.Less(t, bob.Character.Health, 1, "the fixture must reach Bob going down")
}

// An NPC attacks a player (nothing in data does this today; the aggro is set
// directly).
func TestCombatGoldenMvP(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("mvp")
	require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")
}

// An NPC charmed by a player attacks an NPC; the charmer is credited.
func TestCombatGoldenMvMCharmed(t *testing.T) {
	w := newCombatWorld(t, 404)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 10)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 30, 30, 20, 10)
	dog.Character.Charm(alice.Character.Id, 100, "")
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "golden-rats-mvm")
	dog.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("mvm_charmed")
	require.Less(t, rat.Character.Health, 1, "the fixture must reach the rat's death")
}

// fleeAggro is what the flee command sets.
func fleeAggro(c *characters.Character) { c.Aggro = &characters.Aggro{Type: characters.Flee} }

// A player's flee is blocked by the NPC fighting them.
func TestCombatGoldenFleeBlocked(t *testing.T) {
	w := newCombatWorld(t, 505)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 1, 10, 10, 30)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 20, 5, 5, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	fleeAggro(alice.Character)
	w.start()
	w.rounds(3)
	w.check("flee_blocked")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "the fixture must show a blocked flee")
}

// A player's flee succeeds: they leave through the only exit, their charmed
// NPC is told to flee, and the NPC that fought them loses its target.
func TestCombatGoldenFleeSuccess(t *testing.T) {
	w := newCombatWorld(t, 606)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 100, 10, 10, 10)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 1, 5, 5, 10)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 10, 10, 10, 10)
	dog.Character.Charm(alice.Character.Id, 100, "")
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	fleeAggro(alice.Character)
	w.start()
	w.rounds(2)
	w.check("flee_success")
	require.Equal(t, goldenRoomB, alice.Character.RoomId, "the fixture must show a successful flee")
}

// A player shoots an NPC in the next room through an exit; a player in that
// room watches. The remote aggro lasts one attack round.
func TestCombatGoldenShootThroughExit(t *testing.T) {
	w := newCombatWorld(t, 707)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 0)
	w.addPlayer(9303, "Carol", goldenRoomB, 10, 10, 10, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 10, 5, 5, 12, "golden-rats-shoot")
	alice.Character.SetAggroRemote("north", rat.Character.Id, characters.Shooting)
	w.start()
	w.rounds(2)
	w.check("shoot_exit")
}

// A player fights an NPC with a weapon that waits a round, and a second
// player in the room watches.
func TestCombatGoldenObserver(t *testing.T) {
	w := newCombatWorld(t, 808)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93104, "testmace", items.Bludgeoning, 1, 4, 1)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "golden-rats-observer")
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(6)
	w.check("observer")
}

// A hidden player attacks a player with a weapon that waits a round, a third
// player watching: the hidden attacker reads as "Someone" to both of them in
// every line, waits and swings alike.
func TestCombatGoldenHiddenAttacker(t *testing.T) {
	w := newCombatWorld(t, 909)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93105, "testdagger", items.Stabbing, 1, 4, 1)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 10)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 10)
	w.hide(alice.Character)
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(6)
	w.check("hidden_attacker")
}

// A bow that waits a round, shot through an exit: the shooter's wait line
// (empty defender-room text) must not reach the target's room.
func TestShootWaitLineDoesNotReachTargetRoom(t *testing.T) {
	w := newCombatWorld(t, 707)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93103, "testbow", items.Shooting, 1, 6, 1)
	carol := w.addPlayer(9303, "Carol", goldenRoomB, 10, 10, 10, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomB, 10, 5, 5, 12)
	alice.Character.SetAggroRemote("north", rat.Character.Id, characters.Shooting)
	w.start()
	w.rounds(1)
	require.NotEmpty(t, w.reads[alice.UserId], "the shooter reads the wait line")
	require.Empty(t, w.reads[carol.UserId], "observers in the target's room read nothing on the wait round")
}

// The killer is recorded when combat downs the player, so destroying the
// killer before the death does not lose its MobId; healing clears it.
func TestKilledByOutlivesTheKiller(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")

	mobs.DestroyInstance(guard.Character.Id)
	require.Nil(t, mobs.GetInstance(guard.Character.Id))
	require.False(t, alice.Character.KilledBy.IsZero())
	require.Equal(t, 93202, alice.Character.KillerMobId, "the MobId was recorded while the guard existed")

	alice.Character.Heal(alice.Character.HealthMax.Value, 0)
	require.True(t, alice.Character.KilledBy.IsZero(), "back on her feet: no stale killer")
	require.Zero(t, alice.Character.KillerMobId)
}
