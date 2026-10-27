# Phase 4b slice 1: actors (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** every logged-in player and every spawned mob is an `actors.Actor` with a runtime `ActorId`
and a swappable `Controller`; accounts are renamed and lose their connection state to a `Session`.

**Architecture:** a new `internal/actors` package holds `Actor`, `ActorId`, `Controller` (with
`ConnectionController` and `BehaviourController`) and a registry. `internal/users` becomes
`internal/accounts` (`UserRecord` becomes `Account`). Accounts and mobs announce their lifecycle
through `util.Hook`s; `actors.Init()` subscribes, because `accounts` already imports `mobs` and so
neither can import `actors`. Old IDs keep routing; nothing outside `actors` consumes actors yet.

**Tech stack:** Go 1.27.3, testify, `gopls` v0.21.1 for the rename.

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md`, section 4, slice 1.

---

## Facts verified against source (2026-10-27, branch `strip/phase-4b-npcs-are-players` at `9c6cec99`; engine code identical to master `c4314b7f`)

| Fact | Value | Where / how |
|---|---|---|
| `internal/users` | 21 files, 3,909 lines (tests included); `UserRecord` at `userrecord.go:31`, `NewUserRecord` at `userrecord.go:58`, package clause `userrecord.go:1` | `wc -l`, source |
| Importers of `internal/users` | 127 Go files, 4 of them tests | `grep -rl 'GoMud/internal/users"'` |
| Connection state on `UserRecord` | `connectionId`, `unsentText`, `connectionTime`, `lastInputRound`, `activePrompt`, `isLinkDead`, `inputBlocked` (`userrecord.go:48-55`). `tempDataStore` is runtime but not connection state and stays | source |
| Direct uses of those fields | `users.go` 28, `userrecord.go` 20, `copyover.go` 5, `userrecord.prompt.go` 2 (`activePrompt`), `users_test.go` 1 | `grep` |
| `ConnectionId` type | `type ConnectionId = uint64` (alias) | `internal/connections/connections.go:16` |
| Where accounts go online (writes to `userManager.Users`) | `CopyoverReconnectUser` (`users.go:221`), `LoginUser` link-dead branch (`:266`), `LoginUser` new login (`:301`), `CreateUser` (`:421`), `CopyoverRestore` (`copyover.go:73`); `SetTestUser` (`testhelpers.go:13`, tests only) | source |
| Where accounts go offline | `LogOutUserByConnectionId` (`users.go:389`) is the only path for an online account. `DeleteUser` (`:722`) refuses online accounts. Both logout callers: `hooks/PlayerDespawn_HandleLeave.go:65`, `world.go:277` | source |
| `CreateUser` callers | `inputhandlers/login.go:142` (signup), `web/api_v1_users.go:181` (web-created account with `NewUserRecord(0, 0)`, so connection 0). The web path puts an offline account in the online map; this plan keeps that behaviour | source |
| Character pointer swaps on accounts | `usercommands/suicide.go:133` (`user.Character = characters.New()`, permadeath); `UpdateOnlineUser` (`users.go:588-608`, `*u = updated`) fed by `web/api_v1_users.go:93,126` and `web/api_v1_permissions.go:212,220`; `ReplaceCharacter` (`userrecord.go:494`) has no callers | `grep '\.Character = \|ReplaceCharacter(\|UpdateOnlineUser('` |
| Mob instances | `instanceCounter` and `mobInstances` (`mobs.go:22-26`); created only in `NewMobById` (`mobs.go:138-221`, store at `:216`); removed only in `DestroyInstance` (`mobs.go:248`). `NewMobById` callers: `rooms/rooms.go:590`, `usercommands/admin.mob.go:141`. `DestroyInstance` callers: 6 (`mobcommands/despawn.go:14`, `mobcommands/suicide.go:51,388,415`, `rooms/roommanager.go:502,510`). No test spawns a mob | source |
| Import direction | `users` imports `mobs` (`LoginUser` calls `mobs.MobInstanceExists`); `mobs` imports neither `users` nor anything that would import `actors` | `grep` |
| Hook type | `util.Hook[T]` with `Register(func(T) T)` and `Fire(T) T`, synchronous, in order | `internal/util/hook.go:26-49` |
| Logging | Default log level is Debug (`mudlog.SetLogLevel("")`); the smoke server inherits it. Tests that log call `mudlog.SetupLogger(nil, "", "", false)` first | `internal/mudlog/mudlog.go:23-37`; `internal/combat/weapon_rank_test.go:21` |
| Startup | `flags.HandleFlags(VERSION)` at `main.go:101`; copyover contributors registered at `:103-108`; `copyover.Restore` at `:250` | source |
| Source-scanning tests | None exist yet | `grep -rl 'go/parser' --include=*_test.go` |
| `gopls rename` | On a plain export (not a git tree) `gopls rename -w` renames `UserRecord` and the package clause and rewrites every import path, but cannot move the directory on Windows ("The system cannot find the path specified"). After a manual `mv internal/users internal/accounts` the export builds. `-d` mode fails on the move | tried 2026-10-27 in the session scratchpad |
| Doc references to `internal/users` | Only historical (`docs/completed/`, `engine_proposal.md` facts, the 4b spec facts) and upstream `AGENTS.md` notes. None are rewritten | `grep` |
| Smoke client | `client.read()` collects ANSI-stripped text (`cmd/smoke/client.go:70-99`); `send()` writes a line (`:130-139`); `newClient` starts the reader immediately (`:53-60`); `dial` callers: `admin.go:28`, `script.go:130,176`; `newClient` in `client_test.go:21`; `play()` at `script.go:129`; `runBots` at `main.go:89` | source |

## Rules for every task

- Edit files with the Edit and Write tools only. No `sed -i`, no shell redirects into the repo, no
  Python write-opens. Generated files are written to the session scratchpad and landed with `cp` or
  `git mv` (Task 2 does this). `~/.claude/hooks/guard.py` enforces it.
- `git add <named paths>` only. Commit messages through a quoted heredoc ending with the
  `Co-Authored-By` line.
- Every task ends with `go build ./...` and the targeted tests named in it; Tasks 2 and 10 also run
  the full suite and smoke.
- `<session scratchpad>` in a command means the scratchpad directory named in your system prompt,
  written as its full path. Never `C:/tmp`, `$TMP` or `/tmp`.
- A step whose expected output does not match: stop and report the output. Do not change the
  expectation to fit.
- Full suite: `go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'`. The only allowed
  failures are the 3 Windows-only ones in `geomud-verifying`.

## File structure

| File | Responsibility |
|---|---|
| `cmd/smoke/recorder.go` (new) | One bot's transcript across its connections |
| `cmd/smoke/client.go`, `script.go`, `main.go`, `admin.go` | Pass the recorder; `-transcript` flag |
| `cmd/smoke/testdata/pre-4b/bot0.txt` (new) | Pre-4b transcript baseline for slices 5 and 7 |
| `internal/accounts/` (renamed from `internal/users/`) | Accounts |
| `internal/accounts/session.go` (new) | `Session`: connection state of an online account |
| `internal/accounts/lifecycle.go` (new) | `OnActivate`/`OnDeactivate` hooks and the shared activate helper |
| `internal/mobs/mobs.go` | `OnInstanceCreated`/`OnInstanceDestroyed` hooks, `addInstance` |
| `internal/actors/actor.go` (new) | `ActorId`, `Actor` |
| `internal/actors/controller.go` (new) | `Controller`, `ConnectionController`, `BehaviourController` |
| `internal/actors/registry.go` (new) | Spawn, lookups, remove, controller swap |
| `internal/actors/wire.go` (new) | `Init()`: hook subscriptions |
| `internal/guards/` (new) | Source-scanning guard tests for the 4b rules |
| `main.go` | Calls `actors.Init()` |

---

### Task 0: Branch

- [ ] **Step 1:** After Calabe approves this plan, the spec and plan merge to `master` as their own PR
  (branch `strip/phase-4b-npcs-are-players`). Then:

```bash
cd "C:/Users/Calabe Davis/workspace/new_geo_mudbase"
git switch master && git pull --ff-only origin master
git switch -c strip/phase-4b-1-actors
```

---

### Task 1: Smoke transcript recording and the pre-4b baseline

**Files:**
- Create: `cmd/smoke/recorder.go`
- Modify: `cmd/smoke/client.go:18-60,130-139`, `cmd/smoke/script.go:129-213`, `cmd/smoke/main.go:35-158`, `cmd/smoke/admin.go:28`
- Test: `cmd/smoke/client_test.go`
- Create: `cmd/smoke/testdata/pre-4b/bot0.txt`

- [ ] **Step 1: Write the failing test** (append to `cmd/smoke/client_test.go`; change `fakeServer`'s
  `newClient(near, time.Second)` to `newClient(near, time.Second, nil)`)

```go
func TestRecorderKeepsTextAndSentLines(t *testing.T) {
	rec := &recorder{}
	near, far := net.Pipe()
	c := newClient(near, time.Second, rec)
	t.Cleanup(func() { c.Close(); far.Close() })

	far.Write([]byte("\x1b[33mWelcome\x1b[0m\r\n"))
	require.NoError(t, c.wait(context.Background(), "welcome", "Welcome"))
	go func() {
		buf := make([]byte, 64)
		far.Read(buf)
		far.Write([]byte("You say, \"hi\"\r\n"))
	}()
	require.NoError(t, c.do(context.Background(), "say", "say hi", `You say, "hi"`))

	assert.Equal(t, "Welcome\n\n> say hi\nYou say, \"hi\"\n", rec.String())
}

func TestNilRecorderIsSafe(t *testing.T) {
	var rec *recorder
	rec.text("x")
	rec.sent("y")
	assert.Equal(t, "", rec.String())
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./cmd/smoke/ -run 'Recorder' -count=1`
Expected: FAIL, `undefined: recorder` and `too many arguments in call to newClient`.

- [ ] **Step 3: Create `cmd/smoke/recorder.go`**

```go
package main

import (
	"strings"
	"sync"
)

// recorder keeps one bot's whole transcript across its connections: the
// server's text as the waits see it (ANSI and carriage returns stripped) and
// each line the bot sent, marked "> ". A nil recorder records nothing.
type recorder struct {
	mu sync.Mutex
	b  strings.Builder
}

func (r *recorder) text(s string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.b.WriteString(s)
	r.mu.Unlock()
}

func (r *recorder) sent(line string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.b.WriteString("\n> " + line + "\n")
	r.mu.Unlock()
}

func (r *recorder) String() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.b.String()
}
```

- [ ] **Step 4: Thread the recorder through the client**

In `cmd/smoke/client.go`: add the field `rec *recorder // the bot's transcript; nil records nothing`
to `client` after `gmcp`. Change `dial` and `newClient`:

```go
func dial(ctx context.Context, addr string, timeout time.Duration, rec *recorder) (*client, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return newClient(conn, timeout, rec), nil
}

func newClient(conn net.Conn, timeout time.Duration, rec *recorder) *client {
	// lastSend starts at connect: the server's drop window also opens when it
	// sends the first prompt (main.go handleTelnetConnection sets lastInput
	// after the prompt), so an instant first reply is lost too.
	c := &client{conn: conn, timeout: timeout, changed: make(chan struct{}, 1), gmcp: map[string]int{}, lastSend: time.Now(), rec: rec}
	go c.read()
	return c
}
```

In `read()`, after `c.text.WriteString(clean)` and before `c.mu.Unlock()`, add `c.rec.text(clean)`.
In `send()`, after `c.lastSend = time.Now()`, add `c.rec.sent(line)`.

In `cmd/smoke/admin.go:28` change `dial(ctx, addr, timeout)` to `dial(ctx, addr, timeout, nil)`.

- [ ] **Step 5: Run the recorder tests, expect PASS**

Run: `go test ./cmd/smoke/ -count=1`
Expected: `ok`.

- [ ] **Step 6: Pass a recorder through `play` and write transcripts from `runBots`**

`cmd/smoke/script.go`: change the signature to
`func play(ctx context.Context, addr string, b bot, timeout, hold time.Duration, rec *recorder) (map[string]int, error)`
and both `dial(ctx, addr, timeout)` calls (lines 130 and 176) to `dial(ctx, addr, timeout, rec)`.
Add to the doc comment: `rec, when not nil, records the bot's transcript across both connections.`

`cmd/smoke/main.go`: add the flag after `adminPass`:

```go
	transcript := flag.String("transcript", "", "directory to write each bot's transcript to (bot<N>.txt); empty writes none")
```

Pass `*transcript` as a new last argument to both `runBots` calls, and change `runBots`:

```go
func runBots(ctx context.Context, addr string, n int, timeout, hold time.Duration, adm adminLogin, transcriptDir string) bool {
```

Inside the bot goroutine replace `pkgs, err := play(ctx, addr, b, timeout, hold)` with:

```go
			var rec *recorder
			if transcriptDir != "" {
				rec = &recorder{}
			}
			pkgs, err := play(ctx, addr, b, timeout, hold, rec)
			if rec != nil {
				name := filepath.Join(transcriptDir, fmt.Sprintf("bot%d.txt", i))
				if werr := os.WriteFile(name, []byte(rec.String()), 0o644); werr != nil && err == nil {
					err = fmt.Errorf("write transcript: %w", werr)
				}
			}
```

Add `"path/filepath"` to `main.go`'s imports if missing. Before the goroutines start, when
`transcriptDir != ""`, run `if err := os.MkdirAll(transcriptDir, 0o755); err != nil { fmt.Printf("smoke: FAIL transcript dir: %v\n", err); return false }`.

- [ ] **Step 7: Build, test, record the baseline into the scratchpad, land it**

```bash
cd "C:/Users/Calabe Davis/workspace/new_geo_mudbase"
go build ./... && go test ./cmd/smoke/ -count=1
SCR="<session scratchpad>/pre-4b"
go run ./cmd/smoke -transcript "$SCR" 2>&1 | tail -2
mkdir -p cmd/smoke/testdata/pre-4b && cp "$SCR/bot0.txt" cmd/smoke/testdata/pre-4b/bot0.txt
grep -c "^> " cmd/smoke/testdata/pre-4b/bot0.txt
```

Expected: `smoke: PASS 1/1`; the transcript has one `> ` line per bot command (about 30). Read the
file: it must start at the login splash ("Welcome to GeOMUD") and end with the final quit.

- [ ] **Step 8: Commit**

```bash
git add cmd/smoke/recorder.go cmd/smoke/client.go cmd/smoke/client_test.go cmd/smoke/script.go cmd/smoke/main.go cmd/smoke/admin.go cmd/smoke/testdata/pre-4b/bot0.txt
git commit -F - <<'EOF'
smoke: -transcript records each bot's session; pre-4b baseline

The baseline is the engine at master c4314b7f, for the text-parity
checks in Phase 4b slices 5 and 7.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 2: Rename `users` to `accounts` and `UserRecord` to `Account`

Mechanical, generated in a plain export of the tree and landed with `git mv` and `cp`.

**Files:** every importer of `internal/users` (127) and the package itself.

- [ ] **Step 1: Export and rename in the scratchpad**

```bash
cd "C:/Users/Calabe Davis/workspace/new_geo_mudbase"
EXP="<session scratchpad>/rename"
rm -rf "$EXP" && mkdir -p "$EXP" && git archive HEAD | tar -x -C "$EXP"
cd "$EXP"
gopls rename -w internal/users/userrecord.go:31:6 Account
gopls rename -w internal/users/userrecord.go:58:6 NewAccount
gopls rename -w internal/users/userrecord.go:1:9 accounts   # prints an AGENTS.md path error: expected, it cannot move the folder
mv internal/users internal/accounts
go build ./... && go vet ./internal/accounts/
grep -rl 'internal/users"' --include=*.go . | wc -l        # expect 0
grep -rn '\bUserRecord\b' --include=*.go .
```

Expected: build and vet print nothing; the import count is 0. The last grep lists only comments (read
each hit; a code hit means a rename was missed). `IndexUserRecord` (the user index's record type) and
test names such as `TestUserRecord_SetPassword_StoresBcryptHash` keep their names and do not match
`\bUserRecord\b`.

- [ ] **Step 2: Land it in the repo**

```bash
cd "C:/Users/Calabe Davis/workspace/new_geo_mudbase"
git mv internal/users internal/accounts
git ls-files '*.go' | while read -r f; do cmp -s "$EXP/$f" "$f" || cp "$EXP/$f" "$f"; done
git status --short | wc -l
go build ./...
```

Expected: build clean. `git status` lists the renamed folder plus the changed importers.

- [ ] **Step 3: Full verification**

```bash
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'
git diff --name-only -- '*.go' | xargs -r gofmt -l
go run ./cmd/smoke 2>&1 | tail -1
```

Expected: only the 3 known Windows failures; gofmt prints nothing; `smoke: PASS 1/1`.

- [ ] **Step 4: Commit**

```bash
git add internal/accounts $(git diff --name-only)
git status --short     # expect only staged renames (R) and modifications (M), nothing unstaged or untracked
git diff --cached --stat | tail -1
git commit -F - <<'EOF'
Rename users to accounts, UserRecord to Account

Phase 4b slice 1. Generated with gopls rename in an export of the tree
and landed with git mv and cp. No behaviour change.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

`git diff --name-only` lists tracked, modified files only; `internal/accounts` covers the moved
folder.

---

### Task 3: `Session` holds the account's connection state

**Files:**
- Create: `internal/accounts/session.go`, `internal/accounts/session_test.go`
- Modify: `internal/accounts/userrecord.go`, `users.go`, `copyover.go`, `userrecord.prompt.go`, `users_test.go`

- [ ] **Step 1: Write the failing test** (`internal/accounts/session_test.go`)

```go
package accounts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountHasSession(t *testing.T) {
	u := NewAccount(5, 77)
	require.NotNil(t, u.session)
	assert.Equal(t, uint64(77), u.ConnectionId())
}

func TestLoadedAccountGetsSessionOnDemand(t *testing.T) {
	u := &Account{UserId: 3} // as decoded from YAML: no session
	assert.Equal(t, uint64(0), u.ConnectionId())
	u.BlockInput()
	assert.True(t, u.InputBlocked())
}

func TestUpdateOnlineUserKeepsSessionAndCharacterPointer(t *testing.T) {
	resetUserManager()
	live := NewAccount(10, 42)
	live.Username = "live"
	live.Character.Name = "Before"
	live.SetUnsentText("typing")
	userManager.Users[10] = live
	char := live.Character

	edited := *live
	editedChar := *live.Character
	editedChar.Name = "After"
	edited.Character = &editedChar
	edited.session = nil // what the web API decodes: no session
	edited.Role = RoleMod

	UpdateOnlineUser(edited)

	got := userManager.Users[10]
	assert.Same(t, char, got.Character, "the live character pointer must survive an edit")
	assert.Equal(t, "After", got.Character.Name)
	assert.Equal(t, RoleMod, got.Role)
	assert.Equal(t, uint64(42), got.ConnectionId())
	assert.Equal(t, "typing", got.GetUnsentText())
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./internal/accounts/ -run 'Session|UpdateOnlineUser' -count=1`
Expected: FAIL, `u.session undefined`.

- [ ] **Step 3: Create `internal/accounts/session.go`**

```go
package accounts

import (
	"time"

	"github.com/GoMudEngine/GoMud/internal/prompt"
)

// Session is the connection state of an online account. It is never saved.
// It travels with the account: when a later slice lets an account drive a
// different actor (possession), the account and its session move together.
type Session struct {
	connectionId   uint64
	unsentText     string
	connectionTime time.Time
	lastInputRound uint64
	activePrompt   *prompt.Prompt
	isLinkDead     bool // are they a link-dead connection currently?
	inputBlocked   bool // Whether input is currently intentionally turned off (for a certain category of commands)
}

// sess returns the account's session, creating an empty one for an account
// decoded from disk that has not gone online yet.
func (u *Account) sess() *Session {
	if u.session == nil {
		u.session = &Session{}
	}
	return u.session
}
```

- [ ] **Step 4: Move the fields**

In `internal/accounts/userrecord.go` replace the seven fields at lines 48-55 (all but `tempDataStore`)
with one field `session *Session`, keeping `tempDataStore map[string]any`. In `NewAccount` replace
`connectionId: connectionId,` and `connectionTime: time.Now(),` with
`session: &Session{connectionId: connectionId, connectionTime: time.Now()},`.

Then, in `userrecord.go`, `users.go`, `copyover.go` and `userrecord.prompt.go`, rewrite every direct
use `X.connectionId`, `X.unsentText`, `X.connectionTime`, `X.lastInputRound`, `X.activePrompt`,
`X.isLinkDead`, `X.inputBlocked` (X is `u` or `user`) as `X.sess().<field>`. The method names and
signatures on `Account` stay the same, so no caller outside the package changes.

Replace `UpdateOnlineUser`'s body (`users.go`) with:

```go
func UpdateOnlineUser(updated Account) {
	userManagerMu.Lock()
	defer userManagerMu.Unlock()

	u, ok := userManager.Users[updated.UserId]
	if !ok {
		return
	}

	// Runtime state is kept from the live record.
	updated.session = u.session
	updated.tempDataStore = u.tempDataStore
	updated.EventLog = u.EventLog

	// The live character pointer is shared with the account's actor, so it
	// must not change: copy the edited character into it.
	if updated.Character != nil && u.Character != nil {
		*u.Character = *updated.Character
		updated.Character = u.Character
	}

	*u = updated
}
```

In `users_test.go` the literal `connectionId: connId,` (in `TestLogOutUserByConnectionId_ValidUser_CleansUpMaps`)
becomes `session: &Session{connectionId: connId},` and `u.connectionId` becomes `u.sess().connectionId`.

- [ ] **Step 5: Verify no direct field use is left**

```bash
grep -n '\.\(connectionId\|unsentText\|connectionTime\|lastInputRound\|activePrompt\|isLinkDead\|inputBlocked\)\b' internal/accounts/*.go | grep -v 'sess()\.\|session\.go'
```

Expected: no output.

- [ ] **Step 6: Run the tests, expect PASS**

Run: `go build ./... && go test ./internal/accounts/ ./internal/web/ -count=1`
Expected: `ok` for both.

- [ ] **Step 7: Commit**

```bash
git add internal/accounts/session.go internal/accounts/session_test.go internal/accounts/userrecord.go internal/accounts/users.go internal/accounts/copyover.go internal/accounts/userrecord.prompt.go internal/accounts/users_test.go
git commit -F - <<'EOF'
accounts: connection state moves to Session

UpdateOnlineUser keeps the live session and copies edits into the live
character pointer instead of replacing it.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 4: Account lifecycle hooks

**Files:**
- Create: `internal/accounts/lifecycle.go`, `internal/accounts/lifecycle_test.go`
- Modify: `internal/accounts/users.go` (`CopyoverReconnectUser`, `LoginUser`, `CreateUser`, `LogOutUserByConnectionId`), `internal/accounts/copyover.go` (`CopyoverRestore`)

- [ ] **Step 1: Write the failing test** (`internal/accounts/lifecycle_test.go`)

```go
package accounts

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/stretchr/testify/assert"
)

func TestActivateRecordsAndFires(t *testing.T) {
	resetUserManager()
	var got []*Account
	OnActivate.Register(func(u *Account) *Account { got = append(got, u); return u })

	u := NewAccount(21, 0)
	u.Username = "hooked"
	userManagerMu.Lock()
	activateLocked(u, connections.ConnectionId(9))
	userManagerMu.Unlock()
	activated(u)

	assert.Same(t, u, userManager.Users[21])
	assert.Equal(t, 21, userManager.Usernames["hooked"])
	assert.Equal(t, 21, userManager.Connections[9])
	assert.Equal(t, connections.ConnectionId(9), userManager.UserConnections[21])
	assert.Equal(t, uint64(9), u.ConnectionId())
	assert.Contains(t, got, u)
}

func TestDeactivatedFires(t *testing.T) {
	var got []*Account
	OnDeactivate.Register(func(u *Account) *Account { got = append(got, u); return u })
	u := NewAccount(22, 0)
	deactivated(u)
	assert.Contains(t, got, u)
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./internal/accounts/ -run 'Activate|Deactivated' -count=1`
Expected: FAIL, `undefined: OnActivate`.

- [ ] **Step 3: Create `internal/accounts/lifecycle.go`**

```go
package accounts

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/util"
)

var (
	// OnActivate fires after an account is recorded as online: login,
	// link-dead reconnect, copyover reconnect and restore, and account
	// creation. It can fire again for an account that is already online.
	OnActivate util.Hook[*Account]
	// OnDeactivate fires after an online account is logged out.
	OnDeactivate util.Hook[*Account]
)

// activateLocked records u as online on connectionId. The caller holds
// userManagerMu and calls activated(u) after unlocking.
func activateLocked(u *Account, connectionId connections.ConnectionId) {
	u.sess().connectionId = connectionId
	userManager.Users[u.UserId] = u
	userManager.Usernames[u.Username] = u.UserId
	userManager.Connections[connectionId] = u.UserId
	userManager.UserConnections[u.UserId] = connectionId
}

// activated fires OnActivate. Call it without holding userManagerMu.
func activated(u *Account) { OnActivate.Fire(u) }

// deactivated fires OnDeactivate. Call it without holding userManagerMu.
func deactivated(u *Account) { OnDeactivate.Fire(u) }
```

- [ ] **Step 4: Use the helpers at every online and offline point**

Each of these blocks of four map writes becomes `activateLocked(user, <connection id>)`, and
`activated(user)` is called once the lock is released, before the function returns success:

| Function | Map writes now | Connection id argument | Where `activated` goes |
|---|---|---|---|
| `CopyoverReconnectUser` | `users.go:221-224` (also drop the `user.sess().connectionId = connectionId` line above them) | `connectionId` | after `userManagerMu.Unlock()` |
| `LoginUser`, link-dead branch | `:266-269` (drop `user.sess().connectionId = connectionId` at `:264`) | `connectionId` | after `userManagerMu.Unlock()` at `:271` |
| `LoginUser`, new login | `:301-304` (drop `user.sess().connectionId = connectionId` at `:299`) | `connectionId` | after `userManagerMu.Unlock()` at `:306` |
| `CreateUser` | `:421-424` | `u.ConnectionId()` | after `userManagerMu.Unlock()` |
| `CopyoverRestore` (`copyover.go`) | `:73-76` (drop `user.sess().connectionId = entry.ConnectionId`) | `entry.ConnectionId` | after the link-dead check, inside the loop |

In `LogOutUserByConnectionId` the function holds the lock with `defer`. Change it to call
`deactivated(u)` after the lock is released: replace `defer userManagerMu.Unlock()` with explicit
`userManagerMu.Unlock()` calls before each `return`, and call `deactivated(u)` after the unlock on the
`u != nil` path. (Line numbers are from before Task 3; Task 3 does not move these blocks.)

- [ ] **Step 5: Verify every write site uses the helper**

```bash
grep -n 'userManager.Users\[.*\] = ' internal/accounts/*.go
```

Expected: only `lifecycle.go` and `testhelpers.go`.

- [ ] **Step 6: Run the tests, expect PASS**

Run: `go build ./... && go test ./internal/accounts/ ./internal/inputhandlers/ ./internal/web/ -count=1`
Expected: `ok`.

- [ ] **Step 7: Commit**

```bash
git add internal/accounts/lifecycle.go internal/accounts/lifecycle_test.go internal/accounts/users.go internal/accounts/copyover.go
git commit -F - <<'EOF'
accounts: OnActivate/OnDeactivate hooks; one activate helper

The five places an account goes online share activateLocked; logout
fires OnDeactivate after releasing the lock.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 5: An account's character pointer never changes

**Files:**
- Modify: `internal/usercommands/suicide.go:133`, `internal/accounts/userrecord.go:493-496`
- Create: `internal/guards/doc.go`, `internal/guards/guards_test.go`

- [ ] **Step 1: Write the failing guard test**

`internal/guards/doc.go`:

```go
// Package guards holds source-scanning tests that enforce Phase 4b rules
// (2026-10-27-phase-4b-npcs-are-players-design.md). It has no runtime code.
package guards
```

`internal/guards/guards_test.go`:

```go
package guards

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is two levels up from this package.
const repoRoot = "../.."

// goFiles parses every non-test Go file in the repo outside .git and returns
// them keyed by slash path relative to the repo root.
func goFiles(t *testing.T) map[string]*ast.File {
	t.Helper()
	fset := token.NewFileSet()
	out := map[string]*ast.File{}
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "_datafiles") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repoRoot, path)
		out[filepath.ToSlash(rel)] = f
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 100 {
		t.Fatalf("scanned only %d files; the walk is not reaching the repo", len(out))
	}
	return out
}

// TestCharacterPointerIsNotReassigned enforces: once an account or mob has a
// character, that *Character is shared with its actor and is never replaced.
// Edits copy into it (*x.Character = value). Allowed: building a detached copy
// before decoding into it.
func TestCharacterPointerIsNotReassigned(t *testing.T) {
	allowed := map[string]string{
		"internal/web/api_v1_users.go": "decodes the PATCH body into a copy; UpdateOnlineUser copies it into the live pointer",
		"internal/accounts/users.go":   "UpdateOnlineUser points the edited copy at the live pointer",
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range as.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == "Character" {
					if _, ok := allowed[path]; !ok {
						t.Errorf("%s: assigns .Character; copy into the pointer instead (*x.Character = v)", path)
					}
				}
			}
			return true
		})
	}
}
```

- [ ] **Step 2: Run it, expect FAIL naming suicide.go and userrecord.go**

Run: `go test ./internal/guards/ -count=1`
Expected: FAIL with `internal/usercommands/suicide.go: assigns .Character` and
`internal/accounts/userrecord.go: assigns .Character`. If it also names a file not in the facts
table, stop and report it.

- [ ] **Step 3: Fix the two sites**

`internal/usercommands/suicide.go:133`: `user.Character = characters.New()` becomes
`*user.Character = *characters.New()`.

`internal/accounts/userrecord.go`: delete `ReplaceCharacter` (lines 493-496 with its comment); it has
no callers (`grep -rn 'ReplaceCharacter(' --include=*.go .` prints only the definition).

- [ ] **Step 4: Run the guard and the affected packages, expect PASS**

Run: `go build ./... && go test ./internal/guards/ ./internal/usercommands/ ./internal/accounts/ -count=1`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/guards/doc.go internal/guards/guards_test.go internal/usercommands/suicide.go internal/accounts/userrecord.go
git commit -F - <<'EOF'
guards: an account's character pointer is never replaced

New internal/guards package for Phase 4b source rules. Permadeath copies
a fresh character into the pointer; ReplaceCharacter (no callers) goes.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 6: Mob lifecycle hooks

**Files:**
- Modify: `internal/mobs/mobs.go:22-26,138-221,248-251`
- Create: `internal/mobs/lifecycle_test.go`

- [ ] **Step 1: Write the failing test**

```go
package mobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddInstanceAssignsIdStoresAndFires(t *testing.T) {
	var created []*Mob
	OnInstanceCreated.Register(func(m *Mob) *Mob { created = append(created, m); return m })

	m := addInstance(Mob{MobId: 9001})
	assert.NotZero(t, m.InstanceId)
	assert.Same(t, m, GetInstance(m.InstanceId))
	assert.Contains(t, created, m)

	n := addInstance(Mob{MobId: 9001})
	assert.NotEqual(t, m.InstanceId, n.InstanceId)
	DestroyInstance(m.InstanceId)
	DestroyInstance(n.InstanceId)
}

func TestDestroyInstanceFires(t *testing.T) {
	var destroyed []int
	OnInstanceDestroyed.Register(func(id int) int { destroyed = append(destroyed, id); return id })

	m := addInstance(Mob{MobId: 9002})
	DestroyInstance(m.InstanceId)
	assert.Nil(t, GetInstance(m.InstanceId))
	assert.Contains(t, destroyed, m.InstanceId)
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./internal/mobs/ -run 'Instance' -count=1`
Expected: FAIL, `undefined: OnInstanceCreated`.

- [ ] **Step 3: Implement**

Add after the `var (...)` block at `mobs.go:22-31`:

```go
var (
	// OnInstanceCreated fires after a mob instance is stored, from NewMobById.
	OnInstanceCreated util.Hook[*Mob]
	// OnInstanceDestroyed fires with the instance id after DestroyInstance removes it.
	OnInstanceDestroyed util.Hook[int]
)

// addInstance gives mob the next instance id, stores it and fires
// OnInstanceCreated. It is the only place instances are added.
func addInstance(mob Mob) *Mob {
	instanceCounter++
	mob.InstanceId = instanceCounter
	mobInstances[mob.InstanceId] = &mob
	stored := mobInstances[mob.InstanceId]
	OnInstanceCreated.Fire(stored)
	return stored
}
```

In `NewMobById`: delete `instanceCounter++` (line 142) and `mob.InstanceId = instanceCounter`
(line 148), and replace the last two statements of the `if` block
(`mobInstances[mob.InstanceId] = &mob` and `return mobInstances[mob.InstanceId]`, lines 215-218 with
the `// Save the mob instance` comment) with `return addInstance(mob)`. Nothing between line 148 and
the store reads `mob.InstanceId` (check with `sed -n 138,221p internal/mobs/mobs.go | grep InstanceId`,
which must print only the lines being removed).

`DestroyInstance` becomes:

```go
func DestroyInstance(instanceId int) {

	delete(mobInstances, instanceId)
	OnInstanceDestroyed.Fire(instanceId)
}
```

- [ ] **Step 4: Run the tests, expect PASS**

Run: `go build ./... && go test ./internal/mobs/ -count=1`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/mobs/mobs.go internal/mobs/lifecycle_test.go
git commit -F - <<'EOF'
mobs: OnInstanceCreated/OnInstanceDestroyed hooks

addInstance is the one place an instance is numbered and stored.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 7: The `actors` package

**Files:**
- Create: `internal/actors/actor.go`, `controller.go`, `registry.go`, `registry_test.go`

- [ ] **Step 1: Write the failing test** (`internal/actors/registry_test.go`)

```go
package actors

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpawnAndLookup(t *testing.T) {
	Reset()
	acc := &accounts.Account{UserId: 7}
	char := characters.New()
	a := Spawn(char, NewConnectionController(acc))

	require.NotNil(t, a)
	assert.NotZero(t, a.Id)
	assert.Same(t, a, Get(a.Id))
	assert.Same(t, a, ByUserId(7))
	assert.Same(t, char, a.Character)
	assert.Same(t, acc, a.Controller().Account())
	assert.Len(t, All(), 1)
}

func TestOneIdSpaceForPlayersAndNPCs(t *testing.T) {
	Reset()
	p := Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 1}))
	m := &mobs.Mob{InstanceId: 1}
	n := Spawn(&m.Character, NewBehaviourController(m))

	assert.NotEqual(t, p.Id, n.Id)
	assert.Same(t, p, ByUserId(1))
	assert.Same(t, n, ByMobInstanceId(1))
	assert.Nil(t, n.Controller().Account())
	assert.Same(t, &m.Character, n.Character)
}

func TestSetControllerKeepsTheId(t *testing.T) {
	Reset()
	m := &mobs.Mob{InstanceId: 3}
	a := Spawn(&m.Character, NewBehaviourController(m))
	id := a.Id
	acc := &accounts.Account{UserId: 9}

	require.NoError(t, SetController(id, NewConnectionController(acc)))

	assert.Equal(t, id, a.Id)
	assert.Same(t, a, ByUserId(9))
	assert.Nil(t, ByMobInstanceId(3))
	assert.Same(t, acc, Get(id).Controller().Account())
	assert.Error(t, SetController(id+100, NewConnectionController(acc)))
}

func TestRemove(t *testing.T) {
	Reset()
	a := Spawn(characters.New(), NewConnectionController(&accounts.Account{UserId: 4}))
	Remove(a.Id)
	assert.Nil(t, Get(a.Id))
	assert.Nil(t, ByUserId(4))
	Remove(a.Id) // removing twice is harmless
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./internal/actors/ -count=1`
Expected: FAIL, `no non-test Go files` or `undefined: Reset`.

- [ ] **Step 3: Create the package**

`internal/actors/actor.go`:

```go
// Package actors is the one actor model (Phase 4b, decision 4): every live
// body in the world, player or NPC, is an Actor driven by a Controller.
// Spec: 2026-10-27-phase-4b-npcs-are-players-design.md.
package actors

import "github.com/GoMudEngine/GoMud/internal/characters"

// ActorId identifies a live actor. It is runtime only and never saved: a
// player's ActorId differs between logins. Zero means no actor.
type ActorId int

// Actor is one live body in the world.
type Actor struct {
	Id        ActorId
	Character *characters.Character

	controller Controller
}

// Controller returns what drives the actor. Swap it with SetController.
func (a *Actor) Controller() Controller { return a.controller }
```

`internal/actors/controller.go`:

```go
package actors

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/mobs"
)

// Controller is what drives an actor: a connection with its account, or an
// NPC behaviour. Code outside this package asks the controller questions and
// never asserts its concrete type (internal/guards enforces this). Later 4b
// slices add Perceive (slice 3) and Tick (slice 6).
type Controller interface {
	// Account is the account driving the actor, or nil when none does.
	Account() *accounts.Account
}

// ConnectionController drives an actor from an account's connection. The
// connection state is the account's Session, so moving this controller to
// another actor moves the account and its connection together.
type ConnectionController struct {
	account *accounts.Account
}

func NewConnectionController(a *accounts.Account) *ConnectionController {
	return &ConnectionController{account: a}
}

func (c *ConnectionController) Account() *accounts.Account { return c.account }

// BehaviourController drives an NPC. In slice 1 it only holds the mob; slice
// 6 moves the mob's runtime behaviour state here and runs it from Tick.
type BehaviourController struct {
	mob *mobs.Mob
}

func NewBehaviourController(m *mobs.Mob) *BehaviourController {
	return &BehaviourController{mob: m}
}

func (b *BehaviourController) Account() *accounts.Account { return nil }
```

`internal/actors/registry.go`:

```go
package actors

import (
	"fmt"
	"sort"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/characters"
)

var (
	mu              sync.RWMutex
	lastId          ActorId
	byId            = map[ActorId]*Actor{}
	byUserId        = map[int]ActorId{}
	byMobInstanceId = map[int]ActorId{}
)

// Spawn makes a live actor for char, driven by ctrl.
func Spawn(char *characters.Character, ctrl Controller) *Actor {
	mu.Lock()
	defer mu.Unlock()
	lastId++
	a := &Actor{Id: lastId, Character: char, controller: ctrl}
	byId[a.Id] = a
	index(a)
	return a
}

// Get returns the actor with id, or nil.
func Get(id ActorId) *Actor {
	mu.RLock()
	defer mu.RUnlock()
	return byId[id]
}

// ByUserId returns the actor an account drives, or nil. It routes old user
// IDs until slice 2 moves callers to ActorId.
func ByUserId(userId int) *Actor {
	mu.RLock()
	defer mu.RUnlock()
	if id, ok := byUserId[userId]; ok {
		return byId[id]
	}
	return nil
}

// ByMobInstanceId returns the actor for a mob instance, or nil. It routes old
// mob instance IDs until slice 2 removes them.
func ByMobInstanceId(instanceId int) *Actor {
	mu.RLock()
	defer mu.RUnlock()
	if id, ok := byMobInstanceId[instanceId]; ok {
		return byId[id]
	}
	return nil
}

// All returns every live actor, in ActorId order.
func All() []*Actor {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]*Actor, 0, len(byId))
	for _, a := range byId {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}

// Remove drops the actor. Removing an unknown id does nothing.
func Remove(id ActorId) {
	mu.Lock()
	defer mu.Unlock()
	if a, ok := byId[id]; ok {
		unindex(a)
		delete(byId, id)
	}
}

// SetController swaps what drives the actor. The ActorId does not change.
func SetController(id ActorId, ctrl Controller) error {
	mu.Lock()
	defer mu.Unlock()
	a, ok := byId[id]
	if !ok {
		return fmt.Errorf("actor %d not found", id)
	}
	unindex(a)
	a.controller = ctrl
	index(a)
	return nil
}

// rebind points an existing actor at a new character and controller, for an
// account that comes back online as a different record.
func rebind(id ActorId, char *characters.Character, ctrl Controller) {
	mu.Lock()
	defer mu.Unlock()
	if a, ok := byId[id]; ok {
		unindex(a)
		a.Character = char
		a.controller = ctrl
		index(a)
	}
}

// index and unindex keep the old-ID lookups current. The caller holds mu.
func index(a *Actor) {
	if acc := a.controller.Account(); acc != nil {
		byUserId[acc.UserId] = a.Id
	}
	if b, ok := a.controller.(*BehaviourController); ok && b.mob != nil {
		byMobInstanceId[b.mob.InstanceId] = a.Id
	}
}

func unindex(a *Actor) {
	if acc := a.controller.Account(); acc != nil && byUserId[acc.UserId] == a.Id {
		delete(byUserId, acc.UserId)
	}
	if b, ok := a.controller.(*BehaviourController); ok && b.mob != nil && byMobInstanceId[b.mob.InstanceId] == a.Id {
		delete(byMobInstanceId, b.mob.InstanceId)
	}
}

// Reset empties the registry. For tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	lastId = 0
	byId = map[ActorId]*Actor{}
	byUserId = map[int]ActorId{}
	byMobInstanceId = map[int]ActorId{}
}
```

- [ ] **Step 4: Run the tests, expect PASS**

Run: `go build ./... && go test ./internal/actors/ -count=1`
Expected: `ok`. A build error `import cycle not allowed` means a dependency in the facts table is
wrong: stop and report it.

- [ ] **Step 5: Commit**

```bash
git add internal/actors/actor.go internal/actors/controller.go internal/actors/registry.go internal/actors/registry_test.go
git commit -F - <<'EOF'
actors: Actor, ActorId, Controller and the registry

One ID space for players and NPCs; swapping a controller keeps the
ActorId. Old user and mob instance IDs route through lookups until
slice 2.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 8: Wire actors to the lifecycles

**Files:**
- Create: `internal/actors/wire.go`, `internal/actors/wire_test.go`
- Modify: `main.go:101` and its imports

- [ ] **Step 1: Write the failing test** (`internal/actors/wire_test.go`)

```go
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
	m := &mobs.Mob{InstanceId: 41}

	mobs.OnInstanceCreated.Fire(m)
	a := ByMobInstanceId(41)
	require.NotNil(t, a)
	assert.Same(t, &m.Character, a.Character)
	assert.Nil(t, a.Controller().Account())

	mobs.OnInstanceDestroyed.Fire(41)
	assert.Nil(t, ByMobInstanceId(41))
}

func TestInitIsIdempotent(t *testing.T) {
	Reset()
	Init()
	Init()
	mobs.OnInstanceCreated.Fire(&mobs.Mob{InstanceId: 51})
	assert.Len(t, All(), 1, "a second Init must not double-register")
}
```

- [ ] **Step 2: Run it, expect a build failure**

Run: `go test ./internal/actors/ -count=1`
Expected: FAIL, `undefined: Init`.

- [ ] **Step 3: Create `internal/actors/wire.go`**

```go
package actors

import (
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

var initOnce sync.Once

// Init subscribes the registry to account and mob lifecycles. main calls it
// once at startup, before copyover restore and before any room spawns mobs.
// accounts and mobs cannot import this package (accounts imports mobs, and
// this package imports both), so they announce through hooks instead.
func Init() {
	initOnce.Do(func() {
		accounts.OnActivate.Register(func(u *accounts.Account) *accounts.Account { accountOnline(u); return u })
		accounts.OnDeactivate.Register(func(u *accounts.Account) *accounts.Account { accountOffline(u); return u })
		mobs.OnInstanceCreated.Register(func(m *mobs.Mob) *mobs.Mob { mobSpawned(m); return m })
		mobs.OnInstanceDestroyed.Register(func(id int) int { mobDestroyed(id); return id })
	})
}

func accountOnline(u *accounts.Account) {
	if a := ByUserId(u.UserId); a != nil {
		if a.Controller().Account() != u {
			rebind(a.Id, u.Character, NewConnectionController(u))
			mudlog.Debug("actor", "event", "rebound", "kind", "player", "actorId", a.Id, "userId", u.UserId)
		}
		return
	}
	a := Spawn(u.Character, NewConnectionController(u))
	mudlog.Debug("actor", "event", "spawned", "kind", "player", "actorId", a.Id, "userId", u.UserId)
}

func accountOffline(u *accounts.Account) {
	if a := ByUserId(u.UserId); a != nil {
		Remove(a.Id)
		mudlog.Debug("actor", "event", "removed", "kind", "player", "actorId", a.Id, "userId", u.UserId)
	}
}

func mobSpawned(m *mobs.Mob) {
	a := Spawn(&m.Character, NewBehaviourController(m))
	mudlog.Debug("actor", "event", "spawned", "kind", "npc", "actorId", a.Id, "mobInstanceId", m.InstanceId)
}

func mobDestroyed(instanceId int) {
	if a := ByMobInstanceId(instanceId); a != nil {
		Remove(a.Id)
		mudlog.Debug("actor", "event", "removed", "kind", "npc", "actorId", a.Id, "mobInstanceId", instanceId)
	}
}
```

- [ ] **Step 4: Call `Init` from `main.go`**

After `flags.HandleFlags(VERSION)` (`main.go:101`) add:

```go
	// Actors track every player and mob from here on (Phase 4b slice 1).
	actors.Init()
```

and add `"github.com/GoMudEngine/GoMud/internal/actors"` to the imports.

- [ ] **Step 5: Run the tests, expect PASS**

Run: `go build ./... && go test ./internal/actors/ . -count=1`
Expected: `ok` for `internal/actors`; the root package keeps its baseline result.

- [ ] **Step 6: Commit**

```bash
git add internal/actors/wire.go internal/actors/wire_test.go main.go
git commit -F - <<'EOF'
actors: Init wires the registry to account and mob lifecycles

Players get an actor when their account goes online and lose it on
logout; mobs get one on spawn and lose it on destroy.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 9: Guard: no type assertion on a controller outside `actors`

**Files:**
- Modify: `internal/guards/guards_test.go`

- [ ] **Step 1: Add the test**

```go
// TestNoControllerTypeAssertions enforces "ask the actor, never its type":
// outside internal/actors nothing asserts or switches on a concrete
// controller type.
func TestNoControllerTypeAssertions(t *testing.T) {
	concrete := map[string]bool{"ConnectionController": true, "BehaviourController": true}
	names := func(e ast.Expr) string {
		if star, ok := e.(*ast.StarExpr); ok {
			e = star.X
		}
		if sel, ok := e.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "actors" && concrete[sel.Sel.Name] {
				return sel.Sel.Name
			}
		}
		return ""
	}
	for path, f := range goFiles(t) {
		if strings.HasPrefix(path, "internal/actors/") {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeAssertExpr:
				if name := names(x.Type); name != "" {
					t.Errorf("%s: asserts actors.%s; ask the controller instead", path, name)
				}
			case *ast.CaseClause:
				for _, e := range x.List {
					if name := names(e); name != "" {
						t.Errorf("%s: type switch on actors.%s; ask the controller instead", path, name)
					}
				}
			}
			return true
		})
	}
}
```

- [ ] **Step 2: Prove it can fail, then pass**

Temporarily add to `main.go`, inside `main()` after `actors.Init()`:
`if _, ok := actors.Get(0).Controller().(*actors.BehaviourController); ok { panic("guard probe") }`.
Run `go test ./internal/guards/ -count=1`: expected FAIL with
`main.go: asserts actors.BehaviourController`. Remove the probe line with Edit, run again: expected `ok`.

- [ ] **Step 3: Commit**

```bash
git add internal/guards/guards_test.go
git commit -F - <<'EOF'
guards: no type assertion on an actor's controller outside actors

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 10: Verify the slice, record it, open the PR

- [ ] **Step 1: Full verification**

```bash
cd "C:/Users/Calabe Davis/workspace/new_geo_mudbase"
go build ./...
go test ./... -count=1 2>&1 | grep -vE '^ok|no test files'
git diff master --name-only -- '*.go' | xargs -r gofmt -l
LOG="<session scratchpad>/slice1-server.log"
go run ./cmd/smoke -server-log "$LOG" 2>&1 | tail -1
go run ./cmd/smoke -bots 3 2>&1 | tail -1
```

Expected: build clean; only the 3 known Windows failures; gofmt prints nothing; both smoke runs
`PASS`. The package count is the baseline's 32 plus `internal/actors`, `internal/guards` (and
`internal/accounts` replacing `internal/users`): record the real number from the run.

- [ ] **Step 2: Evidence that actors follow the lifecycles**

```bash
grep -c 'msg=actor event=spawned kind=player' "$LOG"
grep -c 'msg=actor event=removed kind=player' "$LOG"
grep -c 'msg=actor event=spawned kind=npc' "$LOG"
```

Expected with one bot: 3 player spawns and 3 player removals (the admin once, the bot at signup and
at relogin), and at least 1 NPC spawn. If the counts differ, read the log around each line and
explain the difference before going on; do not adjust the expectation to fit.

- [ ] **Step 3: Update `PATCH_NOTES.md`, `status.md`, `geomud-verifying`**

- `PATCH_NOTES.md`, under `## 2026-10-27` (or the day the slice lands): one line, for example
  "Phase 4b slice 1: every player and mob is an actor (`internal/actors`: one `ActorId` space,
  swappable controller); `users` renamed `accounts`, connection state in `Session`; guard tests in
  `internal/guards`; smoke `-transcript` with the pre-4b baseline".
- `status.md`: slice 1 done; next is slice 2's plan.
- `.claude/skills/geomud-verifying/SKILL.md`: the package count line, with the number from Step 1.
- Move this plan with `git mv` to `docs/completed/` and repoint references, per `geomud-end-of-day`
  section 4.

- [ ] **Step 4: Commit and open the PR**

```bash
git add PATCH_NOTES.md status.md .claude/skills/geomud-verifying/SKILL.md docs/completed/
git commit -F - <<'EOF'
Phase 4b slice 1 done: notes, status, baseline count

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
git remote -v | grep upstream   # push URL must be the NO-PUSH string
git push -u origin strip/phase-4b-1-actors
gh pr create --repo pruuk/GeOMUD --base master --title "Phase 4b slice 1: actors" --body-file <scratchpad body file>
```

The PR body lists what landed, the verification output from Steps 1 and 2, and ends with the Claude
Code attribution line.

---

## Self-review against the spec

| Spec item (section 4, slice 1) | Task |
|---|---|
| `actors`: `Actor`, `ActorId`, `Controller` (both implementations), registry | 7 |
| `users` renamed `accounts` | 2 |
| Connection state moves from the account to `ConnectionController` | 3 (Session), 7 (the controller holds the account and so its Session). Refinement: the state lives in `accounts.Session`, which the account points to, because `accounts` cannot import `actors`; moving a `ConnectionController` still moves the connection, as the spec requires |
| Logged-in players and spawned mobs become actors | 4, 6, 8 |
| Old IDs still route, through registry lookups | 7 (`ByUserId`, `ByMobInstanceId`) |
| Smoke transcript recording; pre-4b transcript from `c4314b7f` | 1 |
| Deletes: connection fields on the account | 3 |
| Done when: registry tests; controller swap keeps the `ActorId` | 7 |
| Done when: guard test, no `Controller` type switch outside `actors` | 9 |
| Done when: smoke passes | 2, 10 |
