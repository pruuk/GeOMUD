# Phase 4: scripting, web and GMCP surface

Date: 2026-10-26. Parent: `gomud_strip_plan.md` Phase 4 ("Delete `scripting`, its hooks, the plugin
scripting functions and every script datafile. Remove web API routes and GMCP packages that exposed
stripped systems; keep the rest. Scripted buff effects go dead here until Phase 5 rebuilds them in Go.
Done when: `goja` and `gopher-lua` gone from `go.mod`; web admin loads; Mudlet receives GMCP Char,
Room and Comm"). Decision 2 (no JS or Lua). Branch `strip/phase-4-scripting-surface`.

## Decisions (Calabe, 2026-10-26)

- **Smoke item source:** the smoke test logs in as admin and runs `item spawn <id>` once per bot in
  the start room before the bots arrive. No engine or world change. Target mode takes the admin
  password as a flag.
- **Mudlet Discord Rich Presence** (`modules/gmcp/gmcp.Mudlet.go`, 788 lines, 111 Discord mentions)
  is removed.

## Facts verified against source (2026-10-26, master `fb0a0bdc`)

| Fact | Value | Where |
|---|---|---|
| `internal/scripting` | 25 non-test files, 6,592 lines | `wc` |
| Importers (non-test Go files outside the package) | 36: `main.go`, `world.go`, `modules/follow/follow.go`, `internal/plugins/plugins.go`; hooks `Buff_ApplyBuffs`, `ItemOwnership_ItemScriptEvents`, `MobIdle_HandleIdleMobs`, `NewRound_DoCombat`, `NewRound_IdleMobs`, `NewRound_MobRoundTick`, `NewRound_PruneVMs`, `NewRound_UserRoundTick`, `NewTurn_PruneBuffs`, `PlayerDrop_HandlePlayerDrop`, `PlayerSpawn_HandleJoin`, `RoomChange_CleanupEphemeralRooms`, `UserScript_HandleUserScript`; mobcommands `go`, `suicide`; usercommands `admin.teleport`, `equip`, `give`, `go`, `show`, `start`, `suicide`, `usercommands`; web `api_v1_{buffs,items,mobs,rooms,scripting,scripting_dts,user_script}`, `api_v2_spells`, `objecttypes_sync_test` | `grep -rl 'internal/scripting"'` |
| goja / gopher-lua / gopher-luar outside `scripting` | none | grep |
| Script datafiles | 26 `.js`, 0 `.lua`: `sample-scripts/spells` 7; `empty`: buffs 3, spells 8, rooms tutorial 4, shadow_realm 2, startland 1, mobs/tutorial/scripts 1 | `find` |
| Plugin scripting API | `Plugin.AddScriptingFunction` (`internal/plugins/plugins.go:307`), registered into `scripting.AddModlueFunction` (`:511`) | grep |
| Admin scripting pages | `scripting{,-api,-buffs,-functions,-items,-mobs,-objects,-rooms,-spells,-users}.html`, `users-script.html`; Go `admin_scripting.go`, `api_v1_scripting.go`, `api_v1_scripting_dts.go`, `api_v1_user_script.go` (+ test) | `ls` |
| Tutorial | rooms 900-903 run entirely on room scripts; `start.go:183` asks "skip the tutorial?" and the "no" path builds ephemeral copies of `TutorialRooms` (`start.go:220`) | read |
| `item spawn` | admin `item spawn <id>` puts the item on the admin's room floor (`admin.item.go:47`, `room.AddItem` at `:123`) | read |

## Tasks (in order)

1. **Smoke off the tutorial** (Sonnet). Remove the tutorial zone and set `TutorialRooms` to empty; `start`
   skips the tutorial question when no tutorial rooms are configured (new characters go to the start
   room). Smoke: an admin session (temp world in launch mode; `-admin-user`/`-admin-pass` flags in
   target mode, defaulting to the stock admin) spawns one item per bot in the start room, handling
   GoMud's forced first-login password change; bots then get, drop, look, get, walk, say, quit, relogin
   and check they hold it. Keep step names meaningful. This lands first so the smoke test stays green
   while scripting is removed.
2. **Delete scripting** (Opus). `internal/scripting`, the two script-only hooks (`UserScript_HandleUserScript`,
   `NewRound_PruneVMs`), script calls in every other importer (remove only the call; the surrounding
   behaviour stays), `ItemOwnership_ItemScriptEvents` if it becomes empty, the plugin scripting API,
   the admin scripting pages and API, the scripting TypeScript declarations and their sync test, every
   script datafile, and `goja`/`gopher-lua`/`gopher-luar` from `go.mod`. Buffs, spells, mobs and rooms
   that had scripts keep their YAML; their scripted effects go dead until Phase 5. Room scripts that
   gated exits or printed signs simply stop.
3. **GMCP and web surface** (Sonnet). Remove Mudlet Discord Rich Presence. Remove any remaining web API
   route, admin page or GMCP package that only exposed a stripped system (check against Phases 1-4).
   Extend the smoke client to record GMCP packages and fail unless it receives `Char.*`, `Room.*` and
   `Comm.*` packages during the script (the plan's Mudlet check, automated).

## Rules

- Same as Phases 1-3: Edit/Write only, `git rm` for deletions, no public listeners, own PIDs only,
  throwaway files in the scratchpad. Server log of a smoke run accounted for after each task.
- Done per task: `go build ./...`, `go vet ./...`, `gofmt -l`, `go test ./... -count=1` `ok` list diffed by
  name (packages that drop out named), `go run ./cmd/smoke -bots 5` passes.

## Done when

`goja` and `gopher-lua` are gone from `go.mod`; the smoke test passes with 50 bots and asserts GMCP
Char, Room and Comm; all admin pages load in the smoke container (controller); the server log is
clean.
