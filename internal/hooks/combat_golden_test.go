package hooks

// The golden combat test: the parity tool for Phase 4b slice 5 (plan
// Decision 4). Each fixture builds a small fight from scratch, seeds the
// shared random source, runs DoCombat for a fixed number of rounds, and
// records, round by round, what each player read and what state changed. The
// record is compared with testdata/combat-golden/<fixture>.txt; run with
// -update to rewrite the files.
//
// Messages are expanded per receiver by the shared readsRecorder
// (reads_golden_test.go), the way the writers deliver them, so the record is
// "what each player read". Text is kept as markup (before AnsiParse), so a
// colour change shows. Queued commands (events.Input) are recorded, not run:
// a "suicide" in the record means the round queued it. The death fixtures
// (runInputs, Phase 5 slice 3) also run them after recording, as the game
// loop would, and a private copy of a room reads copy-of-<room>.

import (
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
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/rules"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/require"
)

const (
	goldenRoomA     = 93001
	goldenRoomB     = 93002
	goldenRoomC     = 93003 // the death recovery room (recoveryRoom); no exit leads to it
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
	rec     *readsRecorder
	groups  []string // NPC groups whose hatred of each player is recorded

	// what the current round produced
	reads  map[int][]string
	events []string

	// deaths turns on the death fixtures' extras (runInputs): the queued
	// inputs are run, and state lines show gold, experience, worn items and
	// the kill and death books.
	deaths  bool
	pending []events.Input

	out strings.Builder
}

func newCombatWorld(t *testing.T, seed int64) *combatWorld {
	t.Helper()
	loadCombatData.Do(func() {
		// Attack message text and race dice come from the
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
		t:   t,
		rec: newReadsRecorder(t),
	}

	// The recorder holds the reads and the queued inputs; the other events
	// go into the same log, so their order with the inputs is kept.
	note := w.rec.note
	listeners := []events.ListenerId{}
	listeners = append(listeners,
		events.RegisterListener(events.AggroChanged{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.AggroChanged)
			note(fmt.Sprintf("AggroChanged %s room=%s", w.name(ev.ActorId), roomName(ev.RoomId)))
			return events.Continue
		}),
		events.RegisterListener(events.PlayerDrop{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.PlayerDrop)
			note(fmt.Sprintf("PlayerDrop %s room=%s", w.user(ev.UserId), roomName(ev.RoomId)))
			return events.Continue
		}),
		events.RegisterListener(events.CharacterVitalsChanged{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.CharacterVitalsChanged)
			note(fmt.Sprintf("CharacterVitalsChanged %s", w.user(ev.UserId)))
			return events.Continue
		}),
		events.RegisterListener(events.ItemOwnership{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.ItemOwnership)
			note(fmt.Sprintf("ItemOwnership %s item=%d gained=%v", w.name(ev.ActorId), ev.Item.ItemId, ev.Gained))
			return events.Continue
		}),
		events.RegisterListener(events.Buff{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.Buff)
			note(fmt.Sprintf("Buff %s buff=%d source=%s", w.name(ev.ActorId), ev.BuffId, ev.Source))
			return events.Continue
		}),
		events.RegisterListener(events.MobDeath{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.MobDeath)
			note(fmt.Sprintf("MobDeath %s", w.name(ev.ActorId)))
			return events.Continue
		}),
		events.RegisterListener(events.PlayerDeath{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.PlayerDeath)
			note(fmt.Sprintf("PlayerDeath %s", w.user(ev.UserId)))
			return events.Continue
		}),
		events.RegisterListener(events.MobItemDrop{}, func(e events.Event) events.ListenerReturn {
			ev := e.(events.MobItemDrop)
			note(fmt.Sprintf("MobItemDrop mob=%d room=%s zone=%s item=%d", ev.MobId, roomName(ev.RoomId), ev.Zone, ev.ItemId))
			return events.Continue
		}),
	)
	t.Cleanup(func() {
		zeros := []events.Event{events.AggroChanged{}, events.PlayerDrop{},
			events.CharacterVitalsChanged{}, events.ItemOwnership{}, events.Buff{}, events.MobDeath{}, events.PlayerDeath{},
			events.MobItemDrop{}}
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

func (w *combatWorld) name(id ids.ID) string { return w.rec.name(id) }

func (w *combatWorld) user(userId int) string { return w.rec.user(userId) }

// collect sorts what the recorder holds into this round's reads per player
// and events.
func (w *combatWorld) collect() {
	for _, e := range w.rec.take() {
		if e.userId == 0 {
			w.events = append(w.events, e.text)
			continue
		}
		// A world notice ends in CRLF (the player-death notice, suicide.go:88-90);
		// the golden side drops its \r (compareGolden), so the record does too.
		text := strings.TrimSuffix(strings.TrimSuffix(e.text, "\n"), "\r")
		if n := strings.Count(text, "\n"); n > 0 {
			// Combat lines are one line each; a multi-line text is a rendered
			// screen (the look after a flee), recorded by its size only so the
			// goldens do not pin the room template.
			text = fmt.Sprintf("(multi-line text, %d lines)", n+1)
		}
		w.reads[e.userId] = append(w.reads[e.userId], text)
	}
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
	w.rec.names[c.Id] = name
	w.rec.users[userId] = name
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
	w.rec.names[id] = name
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

func shield(itemId int, name string) items.Item {
	spec := items.ItemSpec{
		ItemId: itemId, Name: name, NameSimple: name, Type: items.Offhand, Subtype: items.Wearable,
		DamageReduction: 10,
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
		characters.Flee: "Flee",
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
	s := fmt.Sprintf("room=%s hp=%d/%d aggro=[%s] playerdamage=%s", roomName(c.RoomId), c.Health, c.HealthMax.Value, w.aggro(c), w.damage(c))
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
	if w.deaths {
		s += fmt.Sprintf(" gold=%d xp=%d worn=[%s] kd=%s", c.Gold, c.Experience, itemIds(c.GetAllWornItems()), kd(c.KD))
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
		r := rooms.LoadRoom(rid)
		if r == nil {
			continue
		}
		if len(r.Items) > 0 {
			held := []string{}
			for _, it := range r.Items {
				held = append(held, fmt.Sprint(it.ItemId))
			}
			fmt.Fprintf(&w.out, "  room %d floor: %s\n", rid, strings.Join(held, ","))
		}
		if r.Gold > 0 {
			fmt.Fprintf(&w.out, "  room %d floor gold: %d\n", rid, r.Gold)
		}
		for _, cp := range r.Corpses {
			fmt.Fprintf(&w.out, "  room %d corpse %s: items=[%s] gold=%d worn=[%s]\n",
				rid, cp.Character.Name, itemIds(cp.Items), cp.Gold, itemIds(cp.Character.GetAllWornItems()))
		}
	}
}

// start drops what setup queued and records the starting state.
func (w *combatWorld) start() {
	events.ProcessEvents()
	w.rec.take()
	w.pending = nil
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
		w.drainInputs()
		w.collect()
		w.record(fmt.Sprintf("round %d", i))
	}
}

// record writes what the step just collected under "== <title>": each
// player's reads, the events, then the state.
func (w *combatWorld) record(title string) {
	fmt.Fprintf(&w.out, "== %s\n", title)
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

// check compares the record with the golden file, or writes it with -update.
func (w *combatWorld) check(fixture string) {
	w.t.Helper()
	compareGolden(w.t, "combat-golden", fixture, w.out.String())
}

// roomName is a room ID for the record. A private copy (a death moves the
// body into its own copy of the recovery room) reads copy-of-<room>: its ID
// depends on which copies earlier tests made.
func roomName(id int) string {
	if rooms.IsEphemeralRoomId(id) {
		return fmt.Sprintf("copy-of-%d", rooms.GetOriginalRoom(id))
	}
	return fmt.Sprint(id)
}

// itemIds lists the item IDs, comma separated.
func itemIds(list []items.Item) string {
	out := []string{}
	for _, it := range list {
		out = append(out, fmt.Sprint(it.ItemId))
	}
	return strings.Join(out, ",")
}

// kd is the kill and death books for the record: mob kills/mob deaths/PvP
// kills/PvP deaths/elite kills/elite deaths, then whom the PvP kills and
// deaths were against (userid:name=count, sorted).
func kd(k characters.KDStats) string {
	sum := func(m map[string]int) int {
		n := 0
		for _, v := range m {
			n += v
		}
		return n
	}
	counts := func(m map[string]int) string {
		keys := []string{}
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out := []string{}
		for _, key := range keys {
			out = append(out, fmt.Sprintf("%s=%d", key, m[key]))
		}
		return strings.Join(out, ",")
	}
	return fmt.Sprintf("%d/%d/%d/%d/%d/%d pk=[%s] pd=[%s]", k.TotalKills, k.TotalDeaths, k.TotalPvpKills,
		k.TotalPvpDeaths, sum(k.EliteKills), sum(k.EliteDeaths), counts(k.PlayerKills), counts(k.PlayerDeaths))
}

// freeCopies frees every private copy of roomId. Each copy a death makes
// holds one of the 100 ephemeral chunks (rooms/ephemeral.go) until it is
// freed, and a chunk is freed only when no player is in it, so the players
// leave first; without this, repeated runs (-count=N) run out of chunks.
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

// trinket is an item with its own spec, like weapon: a ring, a gem.
func trinket(itemId int, name string, typ items.ItemType, subtype items.ItemSubType) items.Item {
	spec := items.ItemSpec{ItemId: itemId, Name: name, NameSimple: name, Type: typ, Subtype: subtype}
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &spec}
}

// runInputs makes the record run what a step queues as input (events.Input)
// as a command, after recording it, as the game loop would. The death
// fixtures use it: until Phase 5 slice 3 a death is a queued "suicide". It
// also turns on the state lines' gold, experience, worn items and books. Call it
// before start.
func (w *combatWorld) runInputs() {
	w.t.Helper()
	w.deaths = true
	id := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		w.pending = append(w.pending, e.(events.Input))
		return events.Continue
	})
	w.t.Cleanup(func() { events.UnregisterListener(events.Input{}, id) })
}

// drainInputs runs the inputs runInputs collected, in order, each followed
// by the events it queues (which may queue more). Without runInputs nothing
// is collected and this does nothing.
func (w *combatWorld) drainInputs() {
	for len(w.pending) > 0 {
		in := w.pending[0]
		w.pending = w.pending[1:]
		w.rec.note(fmt.Sprintf("Ran %s: %q", w.name(in.ActorId), in.InputText))
		verb, rest, _ := strings.Cut(in.InputText, " ")
		if _, err := usercommands.TryCommand(verb, rest, in.ActorId, in.Flags); err != nil {
			w.rec.note(fmt.Sprintf("(error: %s)", bodyIdPattern.ReplaceAllString(err.Error(), "<id>")))
		}
		events.ProcessEvents()
	}
}

// heal runs the regen pass (AutoHeal) as on rounds 3, 6, 9 and so on (the
// old pass acts only on every third round), recording each as "== heal N",
// until done reports true, at most max times. done is asked before each
// pass, so a fixture whose death needs no pass records none.
func (w *combatWorld) heal(max int, done func() bool) {
	for i := 1; i <= max && !done(); i++ {
		w.reads = map[int][]string{}
		w.events = nil
		AutoHeal(events.NewRound{RoundNumber: uint64(3 * i)})
		events.ProcessEvents()
		w.drainInputs()
		w.collect()
		w.record(fmt.Sprintf("heal %d", i))
	}
}

// recoveryRoom makes room C the death recovery room for this fixture. A
// death moves the body into a private copy made from the room's template
// file, so the template is saved into a temporary data folder; the repo's
// data folder is never written (a level up saves the account there too).
// Every copy of the room a death made is freed afterwards (freeCopies).
// Call it before start.
func (w *combatWorld) recoveryRoom() {
	w.t.Helper()
	c := rooms.NewEmptyRoom()
	c.RoomId, c.Zone, c.Title = goldenRoomC, "goldenzone", "Room C"
	rooms.AddTestRoom(c)
	w.t.Cleanup(func() { rooms.RemoveTestRoom(goldenRoomC) })
	w.t.Cleanup(func() { freeCopies(goldenRoomC) })
	tmp := w.t.TempDir()
	setConfig(w.t,
		map[string]any{"FilePaths.DataFiles": tmp, "SpecialRooms.DeathRecoveryRoom": goldenRoomC},
		map[string]any{
			"FilePaths.DataFiles":            configs.GetFilePathsConfig().DataFiles.String(),
			"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom),
		})
	require.Equal(w.t, goldenRoomC, int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom), "the override took")
	require.NoError(w.t, os.MkdirAll(filepath.Join(tmp, "rooms", "goldenzone"), 0o755))
	require.NoError(w.t, rooms.SaveRoomTemplate(*rooms.LoadRoom(goldenRoomC)))
}

// deathConfig turns corpses on for one fixture (the test binary loads no
// config.yaml, so they are off otherwise).
func deathConfig(t *testing.T) {
	t.Helper()
	gp := configs.GetGamePlayConfig()
	setConfig(t,
		map[string]any{"GamePlay.Death.CorpsesEnabled": true},
		map[string]any{"GamePlay.Death.CorpsesEnabled": bool(gp.Death.CorpsesEnabled)})
	require.True(t, bool(configs.GetGamePlayConfig().Death.CorpsesEnabled), "the override took")
}

// A player kills an NPC.
func TestCombatGoldenPvMToDeath(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom()
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
// blocks some of the damage.
func TestCombatGoldenPvP(t *testing.T) {
	w := newCombatWorld(t, 202)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 15)
	bob.Character.Equipment.Offhand = shield(93102, "testshield")
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("pvp")
	require.True(t, rooms.IsEphemeralRoomId(bob.Character.RoomId), "the fixture must reach Bob's death")
}

// An NPC attacks a player (nothing in data does this today; the aggro is set
// directly).
func TestCombatGoldenMvP(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("mvp")
	require.True(t, rooms.IsEphemeralRoomId(alice.Character.RoomId), "the fixture must reach Alice's death")
}

// An NPC attacks an NPC (the aggro is set directly); nobody is credited.
func TestCombatGoldenMvM(t *testing.T) {
	w := newCombatWorld(t, 404)
	w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 10)
	dog := w.addNPC(93203, "dog", 1, goldenRoomA, 30, 30, 20, 10)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "golden-rats-mvm")
	dog.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("mvm")
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

// A player's flee succeeds: they leave through the only exit and the NPC
// that fought them loses its target.
func TestCombatGoldenFleeSuccess(t *testing.T) {
	w := newCombatWorld(t, 606)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 100, 10, 10, 10)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 1, 5, 5, 10)
	w.addNPC(93203, "dog", 1, goldenRoomA, 10, 10, 10, 10) // an idle bystander
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
	w.recoveryRoom()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93105, "testdagger", items.Stabbing, 1, 4, 1)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 10)
	w.addPlayer(9303, "Carol", goldenRoomA, 10, 10, 10, 10)
	w.hide(alice.Character)
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(6)
	w.check("hidden_attacker")
	require.True(t, rooms.IsEphemeralRoomId(bob.Character.RoomId), "the fixture must reach Bob's death")
}

// A player kills an elite NPC that carries an item and gold and wears two
// (Phase 5 slice 3, recorded on the death before it moved): XP with the
// elite bonus, alignment, the kill on the books, drops by the NPC's
// ItemDropChance (MobItemDrop each), corpse, and the NPC's ending (it leaves
// the room; its slot's ending is written to the endings ledger).
func TestCombatGoldenDeathNPC(t *testing.T) {
	w := newCombatWorld(t, 111)
	w.recoveryRoom() // the temporary data folder: a level up saves the account
	deathConfig(t)
	w.runInputs()
	dataFolder := configs.GetFilePathsConfig().DataFiles.String()
	t.Cleanup(func() { // registered after recoveryRoom's cleanups, so it runs first, in the temporary folder
		require.NoError(t, os.RemoveAll(filepath.Join(dataFolder, "rooms.instances")))
		require.NoError(t, rooms.LoadEndings()) // no ledger file: the ledger in memory is empty again
	})
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	alice.Character.Alignment = 19 // neutral; killing the evil rat shifts her to lawful
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "golden-rats-death")
	rat.ItemDropChance = 50
	rat.IsElite = true // the elite XP bonus and the elite kill
	rat.SpawnSlot = 1  // a slot NPC: rooms.NPCEnded writes its ending to the ledger
	rat.Character.Alignment = -100
	rat.Character.Equipment.Weapon = weapon(93106, "testclub", items.Bludgeoning, 1, 4, 0)
	rat.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	rat.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	rat.Character.Gold = 30
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	w.check("death_npc")
	require.Nil(t, mobs.GetInstance(rat.Character.Id), "the fixture must reach the rat's death")
	require.NotContains(t, rooms.LoadRoom(goldenRoomA).GetMobs(), rat.Character.Id, "the rat left the room")
	ledger, err := os.ReadFile(filepath.Join(dataFolder, "rooms.instances", "npcs.endings.yaml"))
	require.NoError(t, err, "the rat's ending was recorded (rooms.NPCEnded)")
	require.Contains(t, string(ledger), rat.Character.Id.String())
}

// An elite NPC kills a player (Phase 5 slice 3; recorded on the old death
// first, Task 1). She dies in the round her health reaches 0; the gem she
// carries and her gold go into her corpse; she keeps what she wears and her
// XP; the elite death is booked; she wakes in the recovery room at 1 health.
func TestCombatGoldenDeathPlayer(t *testing.T) {
	w := newCombatWorld(t, 303)
	w.recoveryRoom()
	deathConfig(t)
	w.runInputs()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	alice.Character.Level, alice.Character.Experience = 2, 1000 // the record shows no XP is lost
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	alice.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	alice.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.IsElite = true // her death is booked as an elite death
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	dead := func() bool { return rooms.IsEphemeralRoomId(alice.Character.RoomId) }
	w.heal(20, dead)
	w.check("death_player")
	require.True(t, dead(), "the fixture must reach Alice's death")
}

// A player kills a player (Phase 5 slice 3; recorded on the old death first,
// Task 1): the kill credit and the "killed by" notice; the gem Bob carries
// and his gold go into his corpse; he keeps what he wears and his XP; he
// wakes in the recovery room at 1 health.
func TestCombatGoldenDeathPvP(t *testing.T) {
	w := newCombatWorld(t, 202)
	w.recoveryRoom()
	deathConfig(t)
	w.runInputs()
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	bob := w.addPlayer(9302, "Bob", goldenRoomA, 10, 10, 10, 15)
	bob.Character.Level, bob.Character.Experience = 2, 1000 // the record shows no XP is lost
	bob.Character.Equipment.Ring = trinket(93107, "testring", items.Ring, items.Wearable)
	bob.Character.Items = []items.Item{trinket(93108, "testgem", items.Gemstone, items.Mundane)}
	alice.Character.SetAggro(bob.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(20)
	dead := func() bool { return rooms.IsEphemeralRoomId(bob.Character.RoomId) }
	w.heal(20, dead)
	w.check("death_pvp")
	require.True(t, dead(), "the fixture must reach Bob's death")
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

// downedAtZero is the active ruleset with a downed state: at 0 health or
// below an actor is downed, never dead.
type downedAtZero struct{ rules.Ruleset }

func (downedAtZero) Condition(a *actors.Actor) rules.Condition {
	if a.Character.Health <= 0 {
		return rules.Downed
	}
	return rules.Up
}

// The killer is recorded when combat takes a player out of Up, so under a
// ruleset with a downed state, destroying the killer before the death does
// not lose its MobId; healing clears it. (Under the reference rules death
// is immediate and clears the record itself.)
func TestKilledByOutlivesTheKiller(t *testing.T) {
	w := newCombatWorld(t, 303)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 10, 10, 10, 12)
	guard := w.addNPC(93202, "guard", 1, goldenRoomA, 30, 30, 20, 10)
	guard.Character.SetAggro(alice.Character.Id, characters.DefaultAttack)
	t.Cleanup(rules.SetForTest(downedAtZero{rules.Get()}))
	w.start()
	w.rounds(20)
	require.Less(t, alice.Character.Health, 1, "the fixture must reach Alice going down")
	require.Equal(t, goldenRoomA, alice.Character.RoomId, "downed, not dead")

	mobs.DestroyInstance(guard.Character.Id)
	require.Nil(t, mobs.GetInstance(guard.Character.Id))
	require.False(t, alice.Character.KilledBy.IsZero())
	require.Equal(t, 93202, alice.Character.KillerMobId, "the MobId was recorded while the guard existed")

	alice.Character.Heal(alice.Character.HealthMax.Value, 0)
	require.True(t, alice.Character.KilledBy.IsZero(), "back on her feet: no stale killer")
	require.Zero(t, alice.Character.KillerMobId)
}

// A player wounds an NPC, then the fight ends (both aggros are cleared, as
// a "break" would leave them) and the regen pass runs four times (Phase 5
// slice 4.1: recorded before NPC regen; since then the rat heals each pass).
// The rat has no groups, so the fixture leaves no hatred behind.
func TestCombatGoldenNPCRegen(t *testing.T) {
	w := newCombatWorld(t, 1010)
	alice := w.addPlayer(9301, "Alice", goldenRoomA, 30, 30, 20, 10)
	alice.Character.Equipment.Weapon = weapon(93101, "testsword", items.Slashing, 1, 6, 0)
	rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 40)
	alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
	w.start()
	w.rounds(2)
	alice.Character.EndAggro()
	rat.Character.EndAggro()
	wounded := rat.Character.Health
	w.heal(4, func() bool { return false })
	w.check("npc_regen")
	require.Less(t, wounded, rat.Character.HealthMax.Value, "the fixture must wound the rat")
	require.Greater(t, wounded, 0, "the rat must survive the fight")
}
