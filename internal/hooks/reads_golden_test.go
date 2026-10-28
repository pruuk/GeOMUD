package hooks

// The reads goldens: the parity tool for Phase 4b slice 7 (plan Task 1,
// Decision 12). Each group file runs a sequence of commands in a small fixed
// world and records what every player read, in delivery order, so the output
// conversion can prove its text unchanged apart from the ruled differences.
// The record is compared with testdata/reads-golden/<file>.txt; run with
// -update to rewrite the files.
//
// readsRecorder is shared with the combat golden. It listens to
// events.Message, events.Broadcast and events.Input, and expands each event to
// the players who read it exactly as the writers deliver it:
//   - a UserId message to that user;
//   - a RoomId message to the room's players at drain time, minus the UserId
//     and the ExcludeUserIds, honouring IsQuiet (superhearing only);
//   - a Broadcast to every active account (in UserId order), the screen-reader
//     text for a screen-reader user when there is one.
//
// Text is kept as markup (before AnsiParse), so a colour change shows. Queued
// commands (events.Input) are recorded, not run. Carriage returns are dropped
// from the record: templates are checked out with CRLF on Windows and LF
// elsewhere, and the record must be the same on both.
//
// Record format (reads goldens): "== <title>" starts a section; "> <name>:
// <line>" is a command the fixture ran; "<reader>: <markup>" is the first line
// of one delivered text and "<reader>| <markup>" each further line of the same
// text; "Input <name>: "<text>"" is a queued command; "mask: <what>" near the
// top names a value the record hides (an ID or a path separator that differs
// per run or per OS). A section title in parentheses names a substitution:
// what the plan asked for was missing or nondeterministic in the empty world,
// and the section runs the nearest command that works.

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/colorpatterns"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/exit"
	"github.com/GoMudEngine/GoMud/internal/gamelock"
	"github.com/GoMudEngine/GoMud/internal/gametime"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/language"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/skills"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/stretchr/testify/require"
	textLang "golang.org/x/text/language"

	// The modules whose lines group H records; plugins.Load registers their
	// commands.
	_ "github.com/GoMudEngine/GoMud/modules/cleanup"
	_ "github.com/GoMudEngine/GoMud/modules/follow"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/combat-golden/*.txt and testdata/reads-golden/*.txt from this run")

// readEntry is one delivered text (userId > 0) or one recorded event
// (userId 0, text already formatted).
type readEntry struct {
	userId int
	text   string
}

// readsRecorder records what each player read and the queued inputs.
type readsRecorder struct {
	t     *testing.T
	names map[ids.ID]string
	users map[int]string
	log   []readEntry
}

func newReadsRecorder(t *testing.T) *readsRecorder {
	t.Helper()
	r := &readsRecorder{t: t, names: map[ids.ID]string{}, users: map[int]string{}}
	msgId := events.RegisterListener(events.Message{}, r.onMessage)
	bcId := events.RegisterListener(events.Broadcast{}, r.onBroadcast)
	inId := events.RegisterListener(events.Input{}, func(e events.Event) events.ListenerReturn {
		in := e.(events.Input)
		r.note(fmt.Sprintf("Input %s: %q", r.name(in.ActorId), in.InputText))
		return events.Continue
	})
	t.Cleanup(func() {
		events.UnregisterListener(events.Message{}, msgId)
		events.UnregisterListener(events.Broadcast{}, bcId)
		events.UnregisterListener(events.Input{}, inId)
	})
	return r
}

// note records an event line in delivery order.
func (r *readsRecorder) note(s string) { r.log = append(r.log, readEntry{text: s}) }

// take returns what was recorded since the last take and clears it.
func (r *readsRecorder) take() []readEntry {
	out := r.log
	r.log = nil
	return out
}

func (r *readsRecorder) name(id ids.ID) string {
	if n, ok := r.names[id]; ok {
		return n
	}
	if id.IsZero() {
		return "(none)"
	}
	return "(unknown)"
}

func (r *readsRecorder) user(userId int) string {
	if n, ok := r.users[userId]; ok {
		return n
	}
	return fmt.Sprintf("(user %d)", userId)
}

func (r *readsRecorder) read(userId int, text string) {
	r.log = append(r.log, readEntry{userId: userId, text: text})
}

// onMessage expands one message to its receivers as Message_SendMessage does.
func (r *readsRecorder) onMessage(e events.Event) events.ListenerReturn {
	m := e.(events.Message)
	if m.UserId > 0 {
		if accounts.GetByUserId(m.UserId) != nil {
			r.read(m.UserId, m.Text)
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
			r.read(userId, m.Text)
		}
	}
	return events.Continue
}

// onBroadcast expands one broadcast to every active account as
// Broadcast_SendToAll does. The writer sends without adding a newline.
func (r *readsRecorder) onBroadcast(e events.Event) events.ListenerReturn {
	b := e.(events.Broadcast)
	all := accounts.GetAllActiveUsers()
	sort.Slice(all, func(i, j int) bool { return all[i].UserId < all[j].UserId })
	for _, u := range all {
		if u.ScreenReader && len(b.TextScreenReader) > 0 {
			r.read(u.UserId, b.TextScreenReader)
			continue
		}
		r.read(u.UserId, b.Text)
	}
	return events.Continue
}

// compareGolden compares got with testdata/<dir>/<name>.txt, or writes it
// with -update.
func compareGolden(t *testing.T, dir, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", dir, name+".txt")
	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o644))
		return
	}
	wantBytes, err := os.ReadFile(path)
	require.NoError(t, err, "no golden %s: run go test ./internal/hooks/ -run 'Golden|Reads' -update", path)
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
			t.Fatalf("%s differs from the golden at line %d\n  want: %s\n  got:  %s", path, i+1, x, g)
		}
	}
}

// ---------------------------------------------------------------------------
// The reads world

const (
	readsRoomA     = 93101
	readsRoomB     = 93102
	readsRoomC     = 93103
	readsHiddenBuf = 93902
	readsHearBuf   = 93903
)

var loadReadsData sync.Once

// readsWorld is one group file's world and record.
type readsWorld struct {
	t   *testing.T
	rec *readsRecorder
	out strings.Builder

	nextUserId int
	players    map[string]*accounts.Account
	npcs       map[string]*mobs.Mob
	masks      []func(string) string
}

// newReadsWorld builds rooms A (north to B, east to C), B (south to A) and C
// (west to A) in zone readszone, a recorder, and seeds the random source.
func newReadsWorld(t *testing.T, seed int64) *readsWorld {
	t.Helper()
	loadReadsData.Do(func() {
		// What main loads at boot that these commands read (rooms stay the
		// fixture's own).
		gametime.LoadGameTimeConfigs()
		buffs.LoadFlagDataFiles() // the buff files need the server's timing config; tests set their buffs
		items.LoadDataFiles()
		races.LoadDataFiles()
		skills.LoadDataFiles()
		skills.LoadProfessionDataFiles()
		mobs.LoadDataFiles()
		// As main does. With no file system registered, templates.Process
		// takes a missing .md for an empty one.
		templates.RegisterFS(plugins.GetPluginRegistry())
		templates.LoadAliases(plugins.GetPluginRegistry())
		plugins.Load(configs.GetFilePathsConfig().DataFiles.String())
		colorpatterns.LoadColorPatterns()
		characters.CompileAdjectiveSwaps()
		language.InitTranslation(language.BundleCfg{
			DefaultLanguage: textLang.Make("en"),
			Language:        textLang.Make("en"),
			LanguagePaths:   []string{"../../_datafiles/localize", "../../_datafiles/world/empty/localize"},
		})
	})

	// Commands save users and room templates into the data folder, so each
	// file runs on a copy of the empty world.
	dataDir := configs.GetFilePathsConfig().DataFiles.String()
	tmp := t.TempDir()
	require.NoError(t, os.CopyFS(tmp, os.DirFS(dataDir)))
	// Templates are checked out with CRLF on Windows (core.autocrlf) and LF
	// on Linux CI; the copy gets LF so both render the same (markdown line
	// breaks depend on it).
	require.NoError(t, filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || (!strings.HasSuffix(path, ".template") && !strings.HasSuffix(path, ".md")) {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(b), "\r\n") {
			return err
		}
		return os.WriteFile(path, []byte(strings.ReplaceAll(string(b), "\r\n", "\n")), 0o644)
	}))
	setConfig(t, map[string]any{"FilePaths.DataFiles": tmp}, map[string]any{"FilePaths.DataFiles": dataDir})
	require.Equal(t, tmp, configs.GetFilePathsConfig().DataFiles.String())

	actors.Reset()
	accounts.ResetActiveUsers()
	t.Cleanup(func() {
		accounts.ResetActiveUsers()
		actors.Reset()
	})
	t.Cleanup(rooms.AddTestDefaultBiome())

	a := rooms.NewEmptyRoom()
	a.RoomId, a.Zone, a.Title = readsRoomA, "readszone", "Room A"
	a.Exits = map[string]exit.RoomExit{"north": {RoomId: readsRoomB}, "east": {RoomId: readsRoomC}}
	b := rooms.NewEmptyRoom()
	b.RoomId, b.Zone, b.Title = readsRoomB, "readszone", "Room B"
	b.Exits = map[string]exit.RoomExit{"south": {RoomId: readsRoomA}}
	c := rooms.NewEmptyRoom()
	c.RoomId, c.Zone, c.Title = readsRoomC, "readszone", "Room C"
	c.Exits = map[string]exit.RoomExit{"west": {RoomId: readsRoomA}}
	for _, r := range []*rooms.Room{a, b, c} {
		rooms.AddTestRoom(r)
	}
	t.Cleanup(func() {
		for _, id := range []int{readsRoomA, readsRoomB, readsRoomC} {
			rooms.RemoveTestRoom(id)
		}
	})

	w := &readsWorld{
		t:          t,
		rec:        newReadsRecorder(t),
		nextUserId: 9311,
		players:    map[string]*accounts.Account{},
		npcs:       map[string]*mobs.Mob{},
	}
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("record so far:\n%s", w.out.String())
		}
	})
	// Parties are package state keyed by UserId, and every file reuses the
	// same UserIds.
	t.Cleanup(func() {
		for _, u := range w.players {
			if p := parties.Get(u.UserId); p != nil {
				p.Disband()
			}
		}
	})

	util.SeedRandForTest(seed)
	t.Cleanup(func() { util.SeedRandForTest(time.Now().UnixNano()) })
	fmt.Fprintf(&w.out, "seed %d\n", seed)
	return w
}

// addPlayer adds a player with an account, a character, an actor and an
// active-user entry, in roomId.
func (w *readsWorld) addPlayer(name string, roomId int) *accounts.Account {
	w.t.Helper()
	userId := w.nextUserId
	w.nextUserId++
	c := characters.New()
	c.Id = ids.New(ids.NoCell)
	c.Name = name
	c.RaceId = 1 // human
	c.RoomId = roomId
	c.Zone = "readszone"
	statted(c, 10, 10, 10, 20)
	u := &accounts.Account{UserId: userId, Username: strings.ToLower(name), Character: c, Password: "$2a$test", Role: accounts.RoleUser}
	actors.Spawn(c, actors.NewConnectionController(u))
	accounts.SetTestUser(u)
	rooms.LoadRoom(roomId).AddPlayer(userId)
	w.players[name] = u
	w.rec.names[c.Id] = name
	w.rec.users[userId] = name
	return u
}

// admin makes the account an admin, the way the role is set on an account.
func (w *readsWorld) admin(u *accounts.Account) { u.Role = accounts.RoleAdmin }

// addNPC adds the empty world's rat (mob 1) named name in roomId.
func (w *readsWorld) addNPC(name string, roomId int) *mobs.Mob {
	w.t.Helper()
	m := mobs.Mob{MobId: 1, HomeRoomId: roomId}
	m.Character = *characters.New()
	m.Character.Name = name
	m.Character.RaceId = 10
	m.Character.RoomId = roomId
	m.Character.Zone = "readszone"
	inst := mobs.NewTestInstance(m)
	id := inst.Character.Id
	w.t.Cleanup(func() { mobs.Rest(id) })
	rooms.LoadRoom(roomId).AddMob(id)
	statted(&inst.Character, 10, 5, 5, 12)
	w.npcs[name] = inst
	w.rec.names[id] = name
	return inst
}

// trio is the usual cast: Ann (actor) and Bob (observer) in A, Cid in C, the
// rat in A.
func (w *readsWorld) trio() (ann, bob, cid *accounts.Account, rat *mobs.Mob) {
	ann = w.addPlayer("Ann", readsRoomA)
	bob = w.addPlayer("Bob", readsRoomA)
	cid = w.addPlayer("Cid", readsRoomC)
	rat = w.addNPC("rat", readsRoomA)
	events.ProcessEvents()
	w.rec.take()
	return
}

// mask registers a replacement applied to every recorded line from here on
// (for an ID or clock value that leaks into text) and names it in the record
// as "mask: <what>", so the file says what it hides.
func (w *readsWorld) mask(what string, f func(string) string) {
	fmt.Fprintf(&w.out, "mask: %s\n", what)
	w.masks = append(w.masks, f)
}

var bodyIdPattern = regexp.MustCompile(`0x[0-9a-f]{16}:0x[0-9a-f]{16}`)

// maskBodyIds hides body IDs (ids.ID.String(), random per run) as <id>.
func (w *readsWorld) maskBodyIds() {
	w.mask("body IDs (0x...:0x..., random per run) read <id>", func(s string) string {
		return bodyIdPattern.ReplaceAllString(s, "<id>")
	})
}

// alone takes every player but keep out of the active users while fn runs
// (for a command that lists active accounts in map order).
func (w *readsWorld) alone(keep string, fn func()) {
	others := []*accounts.Account{}
	for name, u := range w.players {
		if name != keep {
			others = append(others, u)
			accounts.RemoveTestUser(u.UserId)
		}
	}
	defer func() {
		for _, u := range others {
			accounts.SetTestUser(u)
		}
	}()
	fn()
}

func (w *readsWorld) section(title string) {
	events.ProcessEvents()
	w.flush()
	fmt.Fprintf(&w.out, "== %s\n", title)
}

// run runs one command line as the named body (a player or an NPC), then
// drains the event queue.
func (w *readsWorld) run(who, line string) {
	w.t.Helper()
	var id ids.ID
	if u, ok := w.players[who]; ok {
		id = u.Character.Id
	} else if m, ok := w.npcs[who]; ok {
		id = m.Character.Id
	} else {
		w.t.Fatalf("no body named %s", who)
	}
	fmt.Fprintf(&w.out, "> %s: %s\n", who, line)
	verb, rest, _ := strings.Cut(line, " ")
	handled, err := usercommands.TryCommand(verb, rest, id, 0)
	if err != nil {
		fmt.Fprintf(&w.out, "(error: %s)\n", w.masked(err.Error()))
	}
	if !handled {
		fmt.Fprintf(&w.out, "(not handled)\n")
	}
	w.drain()
}

// do runs fn as a step named label, then drains the event queue.
func (w *readsWorld) do(label string, fn func()) {
	fmt.Fprintf(&w.out, "> %s\n", label)
	fn()
	w.drain()
}

func (w *readsWorld) drain() {
	events.ProcessEvents()
	w.flush()
}

func (w *readsWorld) masked(s string) string {
	for _, f := range w.masks {
		s = f(s)
	}
	return s
}

// flush writes what the recorder holds to the record.
func (w *readsWorld) flush() {
	for _, e := range w.rec.take() {
		if e.userId == 0 {
			fmt.Fprintf(&w.out, "%s\n", w.masked(e.text))
			continue
		}
		who := w.rec.user(e.userId)
		text := strings.TrimSuffix(e.text, "\n")
		for i, line := range strings.Split(text, "\n") {
			sep := ":"
			if i > 0 {
				sep = "|"
			}
			line = strings.ReplaceAll(w.masked(line), "\r", "")
			if line == "" {
				fmt.Fprintf(&w.out, "%s%s\n", who, sep)
			} else {
				fmt.Fprintf(&w.out, "%s%s %s\n", who, sep, line)
			}
		}
	}
}

func (w *readsWorld) compare(name string) {
	w.t.Helper()
	events.ProcessEvents()
	w.flush()
	compareGolden(w.t, "reads-golden", name, w.out.String())
}

// hide gives the character the hidden flag through a test buff (as the
// combat golden's hidden fixture does); the returned function removes it.
func (w *readsWorld) hide(c *characters.Character) (unhide func()) {
	return w.flag(c, readsHiddenBuf, "hidden")
}

// flag gives the character a test buff carrying one buff flag; the returned
// function removes it.
func (w *readsWorld) flag(c *characters.Character, buffId int, flag string) (remove func()) {
	w.t.Helper()
	w.t.Cleanup(buffs.SetTestSpec(&buffs.BuffSpec{
		BuffId: buffId, Name: "test " + flag, TriggerRate: "1 round", TriggerCount: 1000,
		Flags: []string{flag},
	}))
	hpMax, hp := c.HealthMax.Value, c.Health
	require.NoError(w.t, c.AddBuff(buffId, true)) // runs Validate
	c.HealthMax.Value, c.Health = hpMax, hp
	require.True(w.t, c.HasBuffFlag(flag))
	return func() {
		c.RemoveBuff(buffId)
		c.HealthMax.Value, c.Health = hpMax, hp
		require.False(w.t, c.HasBuffFlag(flag))
	}
}

// saveTemplate writes the room's template file into the copied data folder
// (room set and the death recovery instance read it).
func (w *readsWorld) saveTemplate(roomId int) {
	w.t.Helper()
	dir := filepath.Join(configs.GetFilePathsConfig().DataFiles.String(), "rooms", "readszone")
	require.NoError(w.t, os.MkdirAll(dir, 0o755))
	require.NoError(w.t, rooms.SaveRoomTemplate(*rooms.LoadRoom(roomId)))
}

// setConfig sets config values (dot paths) for the rest of the test and puts
// the overrides and the restore values back afterwards. AddOverlayOverrides
// does not replace a key that is already set, so this goes through
// RestoreOverrides.
func setConfig(t *testing.T, set, restore map[string]any) {
	t.Helper()
	before := configs.Flatten(configs.GetOverrides())
	next := map[string]any{}
	for k, v := range before {
		next[k] = v
	}
	for k, v := range set {
		next[k] = v
	}
	require.NoError(t, configs.RestoreOverrides(next))
	t.Cleanup(func() {
		back := map[string]any{}
		for k, v := range before {
			back[k] = v
		}
		for k, v := range restore {
			back[k] = v
		}
		require.NoError(t, configs.RestoreOverrides(back))
	})
}

// ---------------------------------------------------------------------------
// The group files

// A: admin room commands.
func TestReadsA_AdminRoom(t *testing.T) {
	w := newReadsWorld(t, 1101)
	w.maskBodyIds() // room info lists the mobs here by ID
	w.mask(`room info's Filepath separator reads / (it is \ on Windows)`, func(s string) string {
		return strings.ReplaceAll(s, `readszone\`, `readszone/`)
	})
	ann, _, _, _ := w.trio()
	w.admin(ann)
	// room set and room exit edit template files, so the rooms need them.
	for _, id := range []int{readsRoomA, readsRoomB, readsRoomC} {
		w.saveTemplate(id)
	}

	w.section("room info")
	w.run("Ann", "room info")
	w.section("room exits (no such subcommand: the reply)")
	w.run("Ann", "room exits")
	w.section("room exit south 93103 (the exit subcommand, substituted for room exits)")
	w.run("Ann", fmt.Sprintf("room exit south %d", readsRoomC))
	w.section("room set title")
	w.run("Ann", "room set title The Test Hall")
	w.compare("A-admin-room")
}

// B: admin item, mob and server commands.
func TestReadsB_AdminItemMob(t *testing.T) {
	w := newReadsWorld(t, 1102)
	ann, _, _, _ := w.trio()
	w.admin(ann)

	w.section("item spawn")
	w.run("Ann", "item spawn 10001")
	w.section("mob spawn rat")
	w.run("Ann", "mob spawn rat")
	w.section("server with no args")
	w.run("Ann", "server")
	w.compare("B-admin-item-mob")
}

// C: other admin commands.
func TestReadsC_AdminOther(t *testing.T) {
	w := newReadsWorld(t, 1103)
	ann, _, _, _ := w.trio()
	w.admin(ann)

	w.section("party setup for teleport")
	w.run("Ann", "party invite Bob")
	w.run("Bob", "party accept")
	w.section("teleport Ann with Bob in her party to C")
	w.run("Ann", fmt.Sprintf("teleport %d", readsRoomC))
	w.section("locate Bob")
	w.run("Ann", "locate Bob")
	w.section("modify role bob mod")
	w.run("Ann", "modify role bob mod")
	w.section("mute Bob")
	w.run("Ann", "mute Bob")
	w.section("grant Bob 10 xp")
	w.run("Ann", "grant Bob 10 xp")
	w.section("spawn gold 5")
	w.run("Ann", "spawn gold 5")
	w.section("syslogs with no args")
	w.run("Ann", "syslogs")
	w.compare("C-admin-other")
}

// D: item commands, once with Ann visible and once hidden.
func TestReadsD_Items(t *testing.T) {
	w := newReadsWorld(t, 1104)
	ann, _, _, _ := w.trio()
	room := rooms.LoadRoom(readsRoomA)
	room.Containers = map[string]rooms.Container{"chest": {}}
	ex := room.Exits["east"]
	ex.Lock = gamelock.Lock{Difficulty: 5}
	room.Exits["east"] = ex
	ann.Character.SetKey(fmt.Sprintf("key-%d-east", readsRoomA), "1")

	pass := func(label string) {
		// Each pass starts from the same floor, chest and backpack.
		room.Containers = map[string]rooms.Container{"chest": {}}
		room.Items = nil
		room.Gold = 0
		c := ann.Character
		c.Items = nil
		c.StoreItem(items.New(1))     // note
		c.StoreItem(items.New(1))     // note
		c.StoreItem(items.New(1))     // note
		c.StoreItem(items.New(1))     // note
		c.StoreItem(items.New(20008)) // cotton shirt
		c.StoreItem(items.New(10001)) // sharp stick
		c.Gold = 20

		w.section(label + ": drop and get")
		w.run("Ann", "drop note")
		w.run("Ann", "get note")
		w.section(label + ": put into a container")
		w.run("Ann", "put note chest")
		w.section(label + ": give to Bob")
		w.run("Ann", "give note Bob")
		w.section(label + ": give to the rat")
		w.run("Ann", "give note rat")
		w.section(label + ": equip and remove")
		w.run("Ann", "equip shirt")
		w.run("Ann", "remove shirt")
		w.section(label + ": read")
		w.run("Ann", "read note")
		w.section(label + ": unlock and lock a door exit")
		w.run("Ann", "unlock east")
		w.run("Ann", "lock east")
		w.section(label + ": show to Bob")
		w.run("Ann", "show stick Bob")
		w.section(label + ": drop all and get all")
		w.run("Ann", "drop all")
		w.run("Ann", "get all")
	}
	pass("visible")
	unhide := w.hide(ann.Character)
	pass("hidden")
	unhide()
	w.compare("D-items")
}

// E: moving and looking.
func TestReadsE_MoveLook(t *testing.T) {
	w := newReadsWorld(t, 1105)
	ann, _, cid, rat := w.trio()
	ann.Character.StoreItem(items.New(10001))
	ann.Character.ActionPoints = 1000
	rooms.LoadRoom(readsRoomA).AddItem(items.New(20008), false)
	// "You hear someone moving around" goes quiet to the rooms next to the
	// destination: Cid in C hears Ann arrive in A, with superhearing.
	w.flag(cid.Character, readsHearBuf, "superhearing")

	w.section("north then south (leave, arrive; Cid in C, superhearing, hears the arrival in A)")
	w.run("Ann", "north")
	w.run("Ann", "south")
	w.section("look")
	w.run("Ann", "look")
	w.section("look Bob")
	w.run("Ann", "look Bob")
	w.section("look rat")
	w.run("Ann", "look rat")
	w.section("look north (through an exit)")
	w.run("Ann", "look north")
	w.section("look at an item on the floor (look finds backpack and worn items only: the reply)")
	w.run("Ann", "look shirt")
	w.section("look at an item in the backpack (substituted for the floor item)")
	w.run("Ann", "look stick")
	w.section("consider rat")
	w.run("Ann", "consider rat")
	w.section("attack rat, one round")
	w.run("Ann", "attack rat")
	w.do("DoCombat", func() { DoCombat(events.NewRound{}) })
	// GetRandomExit picks over the exit map, so A keeps one exit for the flee.
	w.section("flee (A's east exit taken away: the exit pick is in map order)")
	roomA := rooms.LoadRoom(readsRoomA)
	east := roomA.Exits["east"]
	delete(roomA.Exits, "east")
	w.run("Ann", "flee")
	w.do("DoCombat", func() { DoCombat(events.NewRound{}) })
	roomA.Exits["east"] = east
	ann.Character.Aggro = nil
	rat.Character.Aggro = nil
	require.NoError(t, rooms.MoveToRoom(ann.Character.Id, readsRoomA)) // back from the flee
	events.ProcessEvents()
	w.rec.take()
	w.section("hidden: look Bob")
	unhide := w.hide(ann.Character)
	w.run("Ann", "look Bob")
	unhide()
	w.compare("E-move-look")
}

// F: speech, broadcast and parties.
func TestReadsF_CommsParty(t *testing.T) {
	w := newReadsWorld(t, 1106)
	ann, _, cid, _ := w.trio()

	w.section("say hi")
	w.run("Ann", "say hi")
	w.section("say {target} (literal braces)")
	w.run("Ann", "say {target}")
	w.section("shout hi (far text in C)")
	w.run("Ann", "shout hi")
	w.section("emote waves")
	w.run("Ann", "emote waves")
	w.section("hidden: emote waves")
	unhide := w.hide(ann.Character)
	w.run("Ann", "emote waves")
	w.section("hidden: say hi")
	w.run("Ann", "say hi")
	unhide()
	w.section("whisper Bob psst")
	w.run("Ann", "whisper Bob psst")
	w.section("broadcast hello")
	w.run("Ann", "broadcast hello")
	w.section("party invite Bob")
	w.run("Ann", "party invite Bob")
	w.section("party accept (Bob)")
	w.run("Bob", "party accept")
	w.section("party chat hi")
	w.run("Ann", "party chat hi")
	w.section("party leave (Bob)")
	w.run("Bob", "party leave")
	// say's "You hear someone talking" goes quiet to the rooms next to the
	// speaker: Cid in C hears it with superhearing, from the exit towards A.
	w.section("say hi (Cid in C, superhearing, hears it from the next room)")
	w.flag(cid.Character, readsHearBuf, "superhearing")
	w.run("Ann", "say hi")
	w.compare("F-comms-party")
}

// G: character information commands.
func TestReadsG_CharacterInfo(t *testing.T) {
	w := newReadsWorld(t, 1107)
	w.trio()
	// The bare help index differs run to run: keywords.yaml lists "online"
	// under two categories and the loader keeps whichever its map order
	// visits last. A topic page is fixed.
	w.section("help say (substituted for help: the bare index is in map order)")
	w.run("Ann", "help say")
	for _, line := range []string{"set", "who", "inventory", "history",
		"macros", "visited", "keyring", "save", "motd", "biome", "print hello"} {
		w.section(line)
		w.run("Ann", line)
	}
	// The bare alias listing groups the built-in aliases by command in map
	// order, so it differs run to run; adding an alias is the fixed reply.
	w.section("alias zz=look (substituted for alias: the bare listing is in map order)")
	w.run("Ann", "alias zz=look")
	// online lists the active accounts in map order, so it runs with Ann the
	// only one.
	w.section("online (Ann the only active account: the list is in map order)")
	w.alone("Ann", func() { w.run("Ann", "online") })
	w.compare("G-character-info")
}

// H: text sent from outside usercommands.
func TestReadsH_Outside(t *testing.T) {
	w := newReadsWorld(t, 1108)
	ann, bob, _, rat := w.trio()
	gp := configs.GetGamePlayConfig()
	setConfig(t,
		map[string]any{"GamePlay.Death.CorpsesEnabled": true, "GamePlay.AlignmentDecayRounds": 100},
		map[string]any{"GamePlay.Death.CorpsesEnabled": bool(gp.Death.CorpsesEnabled), "GamePlay.AlignmentDecayRounds": int(gp.AlignmentDecayRounds)},
	)

	w.section("NPC death: suicide on the rat after damage by Ann (tame and alignment lines)")
	// The tame roll is util.Rand(1000) < level delta + skills delta; a level
	// delta of 1000 makes it certain. Neutral 19 killing an evil rat moves
	// Ann over the line to lawful.
	ann.Character.SetSkill("tame", 1)
	level := ann.Character.Level
	ann.Character.Level = 1000
	ann.Character.Alignment = 19
	rat.Character.Alignment = -100
	rat.Character.PlayerDamage = map[int]int{ann.UserId: 5}
	rat.Character.Health = 0
	w.run("rat", "suicide")
	ann.Character.Level = level

	w.section("cleanup module: bury rat")
	w.run("Ann", "bury rat")
	w.section("cleanup module: trash note")
	ann.Character.StoreItem(items.New(1))
	ann.Character.StoreItem(items.New(1))
	w.run("Ann", "trash note")
	w.section("cleanup module: hidden, trash note")
	unhide := w.hide(ann.Character)
	w.run("Ann", "trash note")
	unhide()

	w.section("a corpse crumbles (a rat corpse placed in A)")
	rooms.LoadRoom(readsRoomA).AddCorpse(rooms.Corpse{MobId: 1, Character: rat.Character})
	w.do("UpdateCorpses far in the future", func() { rooms.LoadRoom(readsRoomA).UpdateCorpses(1 << 40) })

	w.section("room idle message")
	// RoomRoundTick visits the rooms a move has marked as holding players.
	require.NoError(t, rooms.MoveToRoom(ann.Character.Id, readsRoomB))
	require.NoError(t, rooms.MoveToRoom(ann.Character.Id, readsRoomA))
	events.ProcessEvents()
	w.rec.take()
	rooms.LoadRoom(readsRoomA).IdleMessages = []string{"A cold draught blows through Room A."}
	n := 0
	for ; n < 500 && len(w.rec.log) == 0; n++ {
		RoomRoundTick(events.NewRound{})
		events.ProcessEvents()
	}
	fmt.Fprintf(&w.out, "> RoomRoundTick x%d\n", n)
	w.flush()

	w.section("alignment decay")
	ann.Character.Alignment = 20 // lawful; one step down is neutral
	w.do("ActorUpkeep round 100", func() { ActorUpkeep(events.NewRound{RoundNumber: 100}) })

	// A level up queues events.LevelUp, whose only listeners are GMCP's; the
	// gain line is the only text.
	w.section("GrantXP (the level up has no text outside GMCP)")
	w.do("Ann GrantXP 5000", func() { ann.GrantXP(5000, "test") })

	w.section("follow module: Bob follows Ann, Ann goes north")
	ann.Character.ActionPoints = 1000
	w.run("Bob", "follow Ann")
	w.run("Ann", "north")
	w.section("follow module: Bob stops following")
	w.run("Bob", "follow stop")
	w.run("Ann", "south")

	w.section("player death: Bob downed and suicide (recovery room: an instance of B)")
	setConfig(t, map[string]any{"SpecialRooms.DeathRecoveryRoom": readsRoomB},
		map[string]any{"SpecialRooms.DeathRecoveryRoom": int(configs.GetSpecialRoomsConfig().DeathRecoveryRoom)})
	w.saveTemplate(readsRoomB)
	bob.Character.Health = -10
	w.run("Bob", "suicide")
	w.compare("H-outside")
}
