# GoMud Strip-Down Plan

Date: 2026-10-24. Companions: `engine_proposal.md` (why GoMud), `world_model_notes.md` (grid, dense layers, explore), `mud_base_requirements.md`.

**Goal:** turn upstream GoMud into a bare engine: networking, the pulse, rooms, entities, commands, scripting, persistence and protocol. Everything game-specific goes, and the rules (stats, combat math, skills, spells) move behind a seam so the engine stops knowing them. New work (geo grid, provider chain, traits, explore) is **not** in this plan; it starts once the strip is done.

---

## Facts verified against source

Read from the clone in this folder today. "Importers" = non-test packages that import it, from `go list -f '{{.Imports}}' ./...`.

| Fact | Value | Where |
|---|---|---|
| Base | Upstream `GoMudEngine/GoMud` at `b3f55cc2` (2026-10-05), checked out as `master`, remote `upstream` | `git log -1` |
| License | MIT, "Copyright (c) 2026 Dylan Squires" | `LICENSE.md` |
| Toolchain | `go 1.24.0` in `go.mod`; local Go is 1.25.7 | `go.mod`, `go version` |
| External services | None. No database driver; deps are goja, gopher-lua, websocket, yaml, lru, crypto, i18n and small utilities | `go.mod` |
| Build baseline | `go build ./...` passes | run today |
| Test baseline | 34 packages pass; 1 fails: `scripts` `TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride` (`https_setup_test.go:311`). Cause not investigated; likely a Windows shell difference | `go test ./...` today |
| Root program | `main.go` 1,745 lines, `world.go` 715 lines; `MainWorker` at `world.go:322` | `wc -l`, grep |
| Biggest packages | `usercommands` 137 files / 22,687 lines; `web` 59 / 8,455; `scripting` 27 / 7,340; `rooms` 14 / 5,862; `modules/gmcp` 12 / 4,411; `mobcommands` 40 / 3,892; `hooks` 40 / 3,848 | line counts, non-test files |
| Rules packages and importers | `stats` 5, `combat` 5, `statmods` 7, `spells` 7, `skills` 9, `races` 10, `buffs` 13 | `go list` |
| Fixed stat block | `Strength, Speed, Smarts, Vitality, Mysticism, Perception` | `internal/stats/stats.go:9-15` |
| Combat math | `AttackPlayerVsMob` and siblings | `internal/combat/combat.go:28` onward |
| Lazy room load / unload | `LoadRoom(roomId int)`; unload by `RoomUnloadRounds` and `RoomUnloadThreshold` | `internal/rooms/save_and_load.go:78`; `internal/rooms/roommanager.go:170-175`; `internal/configs/config.memory.go:7-8` |
| Dead package | `internal/clans`: zero importers | grep for `internal/clans` |
| Scripting | `scripting` 27 files / 7,340 lines, 7 importers; deps goja, gopher-lua, gopher-luar. Scripts: `world/default` 108 JS + 1 Lua (buffs 38, rooms 31, mobs 15, items 11, spells 9, pets 4), `world/empty` 19 JS. Commonest entry points: `onCommand` 48, `onStart` 37, `onEnd` 30, `onTrigger` 26 | `find`, grep `^function on` |
| Bundled worlds | `_datafiles/world/default` 1,574 files; `_datafiles/world/empty` 353 files (zones `startland`, `tutorial`, `shadow_realm`, `endless_trashheap`) | `find` |
| Modules | `cleanup`, `follow`, `gmcp`, `webhelp`, wired by generated `modules/` imports (`cmd/generate`) | `modules/*.go` |
| Upstream agent docs | `CLAUDE.md` is one line (`@AGENTS.md`); `AGENTS.md` (54 lines) plus nested `AGENTS.md`/`CLAUDE.md` in `rooms`, `scripting`, `web`, `usercommands`, `mobcommands` | `wc -l`, `ls` |
| Upstream scratch docs at root | `election-tax-plan.md` (249 lines, for an elections module not in this repo), `notes.md` (65 lines, TODOs) | `head` |

---

## 1. What we keep

Three buckets. **Keep** stays as is (it may shrink as its callers go). **Keep the shape, move the rules** stays as engine code, with game math moved out to a reference ruleset. **Strip** is deleted.

### Keep: engine core

| Package / area | Why we need it |
|---|---|
| Root `main.go`, `world.go`, `copyover.go`, `signals_*.go` | Boot, the `MainWorker` pulse (50 ms turns, 4 s rounds), shutdown, hot reboot |
| `connections`, `term`, `inputhandlers` | Telnet, websocket, login and account flow |
| `events`, `hooks` (generic ones only, see below) | The event queue and listeners everything else hangs off |
| `gametime` | Game clock, day/night; base for seasons and f(T) |
| `configs`, `mudlog`, `util`, `uuid`, `version`, `flags`, `fileloader`, `gamelock` | Plumbing used across the codebase (`configs` 33 importers, `util` and `mudlog` 38 each) |
| `rooms`, `exit` | Rooms, exits, lazy load and unload, template + instance layering. The future provider chain is built from this |
| `rooms` ephemeral chunks, `procedural` | Generated, copied room sets: the seed for `explore` dense layers |
| `mutators` | Timed room changes with spawn and despawn rounds: the seed for decaying deltas |
| `users`, `characters`, `mobs`, `items` | The entity model (characters lose their stat fields in Phase 5) |
| `parties`, `modules/follow` | Grouping and following are Diku-generic; group encounters need them |
| `plugins`, `modmanager`, `modules/` wiring | How engine features and later the ruleset plug in (minus the scripting functions) |
| `templates`, `markdown`, `colorpatterns`, `keywords`, `language`, `prompt` | Text output, tables, help, aliases, i18n, prompts: the base for EvTable/EvMenu-style helpers |
| `mapper` | Map drawing; GMCP uses it. Kept until the grid replaces it |
| `modules/gmcp` | Structured protocol; becomes the first encoder over typed events |
| `web` (admin login, config, rooms, users, websocket client) | Builder and operator UI; its rules pages go with their packages |
| `badinputtracker`, `modules/cleanup`, `modules/webhelp` | Small, generic admin and help features |
| `cmd/generate`, `cmd/reset-admin-pw`, `Makefile`, `.github/` CI | Build tooling |

**Player commands kept** (`internal/usercommands`): look, go, exits, say, emote, shout, whisper, broadcast, who, online, quit, save, help, motd, alias, macros, set, password, history, inventory, get, drop, give, put, show, equip, remove, read, lock, unlock, keyring, party, visited, biome, default, print, noop, start (character start flow).

**Admin commands kept:** room, build, zone, teleport, visit, locate, spawn, reload, server, syslogs, command, mute, grant, rename, redescribe, item, mob, modify, prepare, copyover, badcommands.

**Mob commands kept** (`internal/mobcommands`): go, wander, pathto, look, say, sayto, emote, shout, broadcast, get, drop, give, put, equip, remove, unlock, despawn, noop.

**Hooks kept:** message sending, broadcast, player spawn, despawn and drop, round and turn ticks for users and mobs, idle mobs, respawns, autosave, link-dead cleanup, inactive players, ephemeral room cleanup, new day, redraw prompt, setting caches, log follow, web client commands.

### Keep the shape, move the rules

These are engine features whose current numbers are GoMud's game. The engine keeps the loop or the container; the math goes to a reference ruleset (`diku-basic`) that also serves as the test fixture.

| Today | Engine keeps | Moves to the ruleset |
|---|---|---|
| `combat` (5 importers), `NewRound_DoCombat` hook, attack/flee/shoot/consider commands | The round loop, who is fighting whom, the commands that start and end a fight | Hit, damage, defence, death, XP |
| `stats` (5) + `statmods` (7) | A keyed traits container (built after the strip) | Stat names and formulas |
| `buffs` (13) + `NewTurn_PruneBuffs`, `Buff_ApplyBuffs` | Timed affects: apply, tick, expire | What each affect does |
| `races` (10) | "Entity has a kind" | Race lists and racial stats |
| `skills` (9), `spells` (7) | Nothing yet; a command and cooldown framework if needed later | All of it |
| `NewRound_AutoHeal`, `NewTurn_ActionPoints` hooks | The per-round tick | Regen and action-point amounts |

### Strip

| What | Size / importers | Why it goes |
|---|---|---|
| `clans` | 0 importers | Dead code |
| `scripting` (JS and Lua), every `.js`/`.lua` datafile, `UserScript_HandleUserScript` and `NewRound_PruneVMs` hooks, plugin scripting functions | 7 importers; 7,340 lines | Behaviour is Go (decision 2 below) |
| `integrations/discord` | root only | Not needed |
| `migration` | root only | Upgrades old GoMud data; we have none |
| `telemetry`, admin telemetry pages | 5 | Gameplay analytics for GoMud's game (local, not phone-home) |
| `audio` (MSP sound), `MSP_PlaySound`, `RoomChange_LocationMusicChange` | 4 | Not needed now; easy to re-add |
| `suggestions` | 2 | Tab completion tied to items, races and mobs; re-add later if wanted |
| `conversations` | 5 | Scripted mob-to-mob chatter; content flavour |
| `quests` + quest hooks and commands | 6 | Game content system |
| `pets` + pet hooks and commands | 7 | Game content system |
| Economy commands: bank, buy, sell, list, offer, appraise, `shophooks.go`, jobs (professions), train, experience, killstats, pvp, gearup, share, stash, use, eat, drink, break, zombieact, suicide, conditions, cooldowns, status | part of `usercommands` | GoMud's game |
| Every `skill.*.go` command (backstab, tame, track, portal, enchant, cast and the rest) plus `skills.go`, `spells.go` | part of `usercommands` | Rules |
| Admin: buff, spell, skillset, questtoken, rankings, telemetry, zap, paz, formset | part of `usercommands` | Belong to stripped systems |
| Mob commands: aid, alchemy, attack, backstab, befriend, break, callforhelp, cast, converse, drink, eat, givequest, lookforaid, lookfortrouble, portal, shoot, show, sneak, suicide, throw | part of `mobcommands` | Rules or stripped systems (attack and shoot come back through the ruleset) |
| Hooks: loot goblin, pet daily tick, pet level, level-up notices, item-ownership quests, look hints, character-created broadcast, sunrise/sunset notices, zone mutator update (review), map rebuild (review) | part of `hooks` | Game features; the two marked "review" are checked before deletion |
| `_datafiles/world/default` | 1,574 files | GoMud's sample world |
| Root `election-tax-plan.md`, `notes.md`, `feature-screenshots/` | | Upstream's working notes |

`_datafiles/world/empty` stays as the base world and gets trimmed to one zone with a handful of rooms, which becomes the test fixture.

---

## 2. Roadmap

Each phase ends with `go build ./...` and `go test ./...` green (allowing the known `scripts` failure until Phase 0 settles it) and a **smoke test**: log in over telnet, walk, say, pick up and drop an item, log out, log back in. After Phase 5 the smoke test also runs a fight through the reference ruleset. The compiler is the guide throughout: delete a package, follow the errors, and decide each call site by the buckets above.

| Phase | Work | Done when |
|---|---|---|
| **0. Baseline** | Commit the clone with our design docs. Settle the `scripts` test failure (fix it or record it as Windows-only). Build the smoke test. Record idle memory and memory with ~50 bots on the default world. Delete upstream's root `AGENTS.md` (our `CLAUDE.md` no longer loads it) | Baseline numbers and test status written into the repo |
| **1. Leaf packages** | Delete in order of fewest importers: `clans`, `integrations/discord`, `migration`, `suggestions`, `audio`, `telemetry`, `conversations`, `quests`, `pets`, with their hooks, commands, admin pages, API routes and GMCP pieces | No references remain; smoke test passes |
| **2. Game commands and hooks** | Remove the stripped command, mob command and hook lists from section 1 | `usercommands` and `mobcommands` hold only the kept lists plus the combat commands waiting for Phase 5 |
| **3. World data** | Delete `world/default`; point config at a trimmed `empty` world; strip datafile folders whose systems are gone (pets, quests, professions, conversations, audio) | Server boots on the trimmed world with no missing-file warnings |
| **4. Scripting, web and GMCP surface** | Delete `scripting`, its hooks, the plugin scripting functions and every script datafile. Remove web API routes and GMCP packages that exposed stripped systems; keep the rest. Scripted buff effects go dead here until Phase 5 rebuilds them in Go | `goja` and `gopher-lua` gone from `go.mod`; web admin loads; Mudlet receives GMCP Char, Room and Comm |
| **5. Rules seam** | Add a `Ruleset` interface the core calls (attack resolution, damage, regen, death, starting values). Move `combat` math, `stats`, `statmods`, `races`, `skills`, `spells` behind it as `diku-basic`. Turn `buffs` into engine timed affects with ruleset-defined effects written in Go. Add a guard test that fails if a core package names a `diku-basic` stat | A fight runs entirely through the ruleset; swapping in a stub ruleset changes the outcome with no core edit; the guard test passes |
| **6. Measure and tag** | Re-measure memory and package counts against Phase 0; tag the result as the stripped base | Before/after numbers in the repo; tag `strip-base` |

**Ordering notes**

- Phases 1 to 4 are mechanical and suit Sonnet subagents, one package or command group per task, each followed by build and smoke. Phase 5 needs judgement (Opus), and it is where most of the risk sits: `buffs` alone has 13 importers, and `characters` imports every rules package.
- Phase 5 comes after the cuts on purpose. Every stripped caller of `stats` or `buffs` is one fewer call site to move behind the interface.
- The `RoomId int` to `RoomKey` change from `engine_proposal.md` is **not** in this plan. It belongs to the geo work, and doing it on the smaller post-strip codebase is cheaper. Upstream fixes stop applying cleanly once Phase 1 lands, which is accepted: the fork is a one-time base.

---

## Decisions (Calabe, 2026-10-24)

1. **Rules packages move into `diku-basic`, then shrink.** Keeps a working game to test against at every step of Phase 5.
2. **No JS or Lua scripting.** Every behaviour in the scripts today (room commands and triggers, mob idle and dialogue, item use, buff and spell effects) is either rules, which go to the ruleset in Go, or content we are deleting. Hand-built areas that need behaviour get named Go behaviours referenced from their YAML. Revisit only if builders who don't write Go join the project.
3. **Never commit, push or open a PR to upstream GoMud.** Enforced in git (`geomud-git` skill).
