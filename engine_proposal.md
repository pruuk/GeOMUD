# Geo-Diku MUD Engine: Single-Codebase Proposal

Date: 2026-10-23. Companion to `mud_architecture.md` and `mud_base_requirements.md`.

**Recommendation in one line:** hard-fork upstream **GoMud** (MIT since 2026-05-05), strip it to an engine, put the rules behind a ruleset interface, and replace its room storage with a layered room-provider chain fed by an offline GIS bake. One language (Go), one binary, no Rust or Elixir.

---

## Facts verified against source

Every row was read from source or the GitHub API today. "Upstream" means `GoMudEngine/GoMud` at `b00487046` (2026-10-05), read through the `upstream` remote of the DOGMud checkout.

| Fact | Value | Where it was read |
|---|---|---|
| GoMud license | MIT now; **GPL-3.0 until 2026-05-05**, relicensed in commit `79c7bf97d` | `git log upstream/master -- LICENSE.md`; `git show 855807d3b:LICENSE.md` is GPL-3.0 |
| DOGMud's own `LICENSE.md` | Still GPL-3.0 (forked before the relicense) | `DOGMud/LICENSE.md` |
| Upstream activity | Last commit 2026-10-05; 237 stars; 29 open issues; not archived | GitHub API |
| Upstream size | 558 non-test `.go` files | `git ls-tree -r upstream/master` |
| External services needed | None. No database driver in `go.mod`; deps are websocket, yaml, goja, gopher-lua, lru, crypto and small utilities | `upstream/master:go.mod` |
| Pulse and round | `TurnMs: 50`, `RoundSeconds: 4`; `MainWorker` fires `NewTurn` each turn and `NewRound` every `TurnsPerRound()` turns | `_datafiles/config.yaml:160,164`; `world.go:322` onward |
| Lazy room load | `LoadRoom(roomId int)` checks memory, else loads template YAML plus instance YAML and caches it | `internal/rooms/save_and_load.go:78` |
| Room unload | Rooms unvisited for `RoomUnloadRounds` (450) are unloaded once more than `RoomUnloadThreshold` (200) are in memory | `internal/rooms/roommanager.go:170-228`; `config.yaml:106,109` |
| Override layering already exists (two layers) | Instance YAML is unmarshalled onto template data; only changed fields are saved (`instance:"skip"` tag excludes fields) | Header comment, `save_and_load.go` |
| Boot-time room index | `roomManager.roomIdToFileCache` maps every room ID to a file | `internal/rooms/admin.go:95,467,601` |
| Room identity | `RoomId int`, also the filename; exits are `map[string]exit.RoomExit` | `internal/rooms/rooms.go:66,76` |
| `RoomId` reach | 591 matching lines across `.go` files (tests included) | `git grep -c '\bRoomId\b'` |
| Time-stamped room changes | `Mutator` has `SpawnedRound` and `DespawnedRound` ("useful for decay"); text modifiers prepend/append/replace | `internal/mutators/mutators.go` |
| Biome field | `Biome string` on rooms, "used for weather generation"; no weather system in upstream code | `rooms.go:74`; `git grep -il weather` |
| Procedural generation | Seeded grid mazes only (`internal/procedural/gridmaze.go`); no GIS or coordinates | `internal/procedural/` |
| Ephemeral rooms | Chunks of copied rooms at IDs ≥ 1,000,000 (100 chunks × 1,000 rooms) | `internal/rooms/ephemeral.go` |
| Rules baked into core | Fixed six-stat struct (`Strength, Speed, Smarts, Vitality, Mysticism, Perception`); `AttackPlayerVsMob` and three siblings in `combat.go` | `internal/stats/stats.go`; `internal/combat/combat.go:28-239` |
| Rules coupling (non-test files importing each package) | stats 6, combat 8, statmods 9, spells 14, skills 17, races 21, buffs 27 | `git grep -l` per import path |
| Plugin API | `AddUserCommand`, `AddMobCommand`, `AddScriptingFunction`, `ExportFunction`, `AttachFileSystem`, `WriteStruct/ReadIntoStruct` | `internal/plugins/plugins.go:269-449` |
| Scripting | JavaScript (goja) and Lua (gopher-lua), chosen by file extension | `internal/scripting/engine.go`, `vmwrapper.go`, `luavm.go` |
| Protocols | GMCP module (Char, Comm, Room, World, Gametime, Party, Mudlet...); websocket port; admin web UI and REST API | `modules/gmcp/`; `internal/web/` |
| Player persistence | One YAML file per user, `users/<id>.yaml` | `internal/users/users.go:450,630` |
| Table and prompt helpers | `templates.GetTable(title, headers, rows, ...)`; `prompt.Prompt.Ask(question, options, default...)` | `internal/templates/templates.go:363`; `internal/prompt/prompt.go:58` |
| Kalevala | MIT, last commits 2025-05-04 (removing deprecation warning, telemetry, credo); ExVenture archived, last push 2023-02-14 | GitHub API |
| Evennia | BSD-3, active (commits 2026-08-19) | GitHub API |
| Ranvier core | MIT, last commit 2019-08-10 | GitHub API |
| CoffeeMud | Apache-2.0, Java, pushed 2026-10-23 | GitHub API |
| DikuMUD3 | LGPL-2.1, C++, active (commits 2026-10-21) | GitHub API, `LICENSE` |
| tbaMUD / CircleMUD | CircleMUD license: "You must not use CircleMUD to make money or be compensated in any way", plus the DikuMUD license | `tbamud/doc/license.txt` |
| Rust MUD engines | Largest is `duysqubix/MuOxi` (102 stars, GPL-3.0); `xenith-studios/ataxia` last push 2023-08-30 | GitHub search |

---

## Part 1: Evaluation of the concept

### What is strong, and survives any stack choice

The real value in `mud_architecture.md` is the **data model**, and none of it depends on Rust or Elixir:

1. **Geography-derived room IDs.** This is the load-bearing idea. Overrides, saved locations and deltas all hang off it.
2. **`room_state(T) = baseline(id, seed, T) + deltas(T)`.** Computing at T instead of stepping forward on load gives one consistent history regardless of who visited when. Correct call.
3. **Four-layer content precedence** (generated, patch, replacement, anchored area). Lets the world exist on day one and improve without migrations.
4. **Region tier** for effects that cross unloaded rooms. Pure lazy loading genuinely cannot do weather or fire spread without it.
5. **Storage tiers** split by how the data changes. Sound.
6. **Generator drift and ID stability** named as top risks, with fingerprint tests as the mitigation. That is the right fear.

### What is weak

1. **Three things to own for one developer.** Rust core, Elixir runtime, and a fork of Kalevala. Kalevala's last commits (2025-05-04) are maintenance removals, and ExVenture is archived. The doc's own list of what Kalevala gives you is networking, controllers, GMCP and views. The whole Diku layer (combat rounds, mobs, items, affects, spawns) would be written from scratch on top.
2. **Rust is in the design before profiling asks for it.** The doc's rule is "anything moves to Rust only when profiling says it must," yet Rust is the day-one world core. GIS ingest is a build step, not a runtime job: bake GeoPackage and DEM data offline into a compact grid, and runtime sampling becomes an array lookup that any language does at millions per second. That removes the NIF risk row from the risk table entirely.
3. **Diku zone resets and f(T) are not reconciled.** Diku worlds repopulate on zone reset timers and mobs wander between rooms. A pure f(T) baseline has no wandering. The design needs an explicit rule: spawn tables are part of the baseline (deterministic from room ID, seed and T), and anything that moved is a delta.
4. **The biggest open question is still open.** Room density and wilderness movement drive the ID scheme, the seam model and the generator. Everything else can wait; this cannot. My recommendation is below.
5. **"Borrow ideas, not code"** made sense against GPL and AGPL sources. GoMud is now MIT, so the constraint no longer protects anything, and it costs the most expensive part of the project (the Diku runtime).

### Verdict

Keep the data model wholesale. Drop the three-runtime stack. The requirements ask for Diku timing, a light footprint, lazy rooms, override layering and a client-agnostic protocol; GoMud already ships the first four in a basic form and most of the fifth.

---

## Part 2: Candidate comparison

Scored against the requirements file. ✔ = has it, ◐ = partial, ✘ = absent or blocked.

**Verification status:** the GoMud column, plus the License and Upstream rows for every engine, are verified (facts table). The other engines' feature cells come from general knowledge of those codebases and were **not** read from source today. They don't change the result, because GoMud wins on verified rows alone, but check them before quoting them anywhere.

| Requirement | GoMud (Go) | Kalevala (Elixir) | Evennia (Python) | DikuMUD3 (C++) | CoffeeMud (Java) | tbaMUD / ROM (C) | Ranvier (JS) |
|---|---|---|---|---|---|---|---|
| Diku pulse and combat rounds | ✔ built in | ✘ build it | ◐ Twisted single reactor fights the pulse | ✔ | ✔ | ✔ | ◐ |
| Light footprint | ✔ one binary, no services | ✔ BEAM is light | ✘ the original reason it was rejected | ✔ | ✘ JVM, very large | ✔ | ◐ |
| Lazy room load and unload | ✔ | ◐ room processes, you write hydration | ◐ DB-backed objects cached on access; no generated rooms | ✘ zones loaded at boot | ✘ | ✘ | ✘ |
| Override layering | ◐ template + instance | ✘ | ◐ prototypes | ✘ | ✘ | ✘ | ✘ |
| Rules separable from core | ◐ coupled, extractable (see numbers) | ✔ no rules at all | ✔ | ✘ deeply Diku | ✘ deeply its own | ✘ | ◐ bundles |
| GMCP + websocket | ✔ | ✔ | ✔ | ✔ websocket | ◐ | ◐ patches | ◐ |
| Scripting for content | ✔ JS + Lua | ✘ | ✔ Python | ✔ DIL | ✔ | ✘ | ✔ |
| License fit | ✔ MIT (fork post-2026-05-05) | ✔ MIT | ✔ BSD-3 | ◐ LGPL-2.1 | ✔ Apache-2.0 | ✘ no commercial use | ✔ MIT |
| Upstream alive | ✔ 2026-10-05 | ✘ maintenance only | ✔ | ✔ | ✔ | ✔ | ✘ 2019 |
| Your existing depth | ✔ DOGMud is a GoMud fork | ✘ | ◐ past project | ✘ | ✘ | ✘ | ✘ |

**GoMud wins** on every requirement except rules separation, and that gap is measurable and bounded (coupling counts in the facts table). DikuMUD3 is the runner-up on feel, but its rules are the codebase and C++ plus DIL is a poor base for a rules-agnostic engine. Kalevala is the cleanest blank slate, which is exactly its problem: the engine you would build on it is the engine GoMud already is.

**Fork from upstream, not DOGMud.** DOGMud's `LICENSE.md` is still GPL-3.0 and it carries game-specific code; the requirements say nothing is ported.

---

## Part 3: The proposal

### Scope: what "engine" means

**In the engine**
- Networking: telnet (with GMCP), websocket, TLS, the admin web API.
- Clock: turn pulse, combat rounds, game time, seasons.
- Command dispatch, user and mob command registration, plugins, scripting.
- Entity model: characters, mobs, items, rooms, with a traits container instead of fixed stats.
- A **generic combat round loop** that asks the ruleset what happens.
- Room provider chain, lazy load and unload, deltas with decay, region tier.
- Persistence (players, deltas), authored YAML loading, generated content on demand.
- Builder conveniences: traits, menus, tables and forms, an inspector.
- GIS bake tool and the generator that reads its output.

**Out of the engine** (a game supplies these)
- Stat names, races, classes, skills, spells, combat math, XP curves.
- World content beyond one test region and one anchored test area.

A small **reference ruleset** ("diku-basic": hit roll, damage roll, AC) ships in the repo as an example and as the test fixture that proves the seams work. It is not part of the engine API.

### Architecture after the fork

```
 clients: telnet+GMCP · websocket JSON · (Godot, web)
                │  typed event structs → GMCP and JSON encoders
 ┌──────────────▼───────────────────────────────────────────┐
 │ GoMud core (kept): connections, MainWorker pulse,        │
 │ event queue, commands, plugins, JS/Lua scripting, admin  │
 ├──────────────────────────────────────────────────────────┤
 │ NEW  ruleset interface   ← reference diku-basic ruleset  │
 │ NEW  traits container    (replaces stats.Statistics)     │
 │ NEW  room provider chain (replaces roomIdToFileCache)    │
 │      generated → patch → replacement → anchor → deltas   │
 │ NEW  region tier         (weather, spread, season)       │
 └──────┬──────────────────────────┬────────────────────────┘
        │ reads baked tiles         │ reads/writes
 ┌──────▼──────────────┐    ┌──────▼─────────────────────────┐
 │ geo bake output      │    │ YAML in git (authored)         │
 │ (versioned files)    │    │ SQLite (players, deltas)       │
 └──────▲──────────────┘    └────────────────────────────────┘
        │ offline, at build time
 ┌──────┴──────────────────────────────┐
 │ geobake tool: GeoPackage, DEM,      │
 │ GEBCO → tiles + feature index       │
 └─────────────────────────────────────┘
```

### Key changes, each tied to what exists today

**1. Room identity.** `RoomId int` (591 lines) becomes a `RoomKey` type. Geography-derived keys encode tile, cell and level; authored anchored areas get their own namespace. Do this first and mechanically, while the code still matches upstream, because every later change sits on it.

**2. Room provider chain.** GoMud's template-then-instance merge in `LoadRoomInstance` is already a two-layer override. Generalise it to an ordered list of providers, each able to supply or patch a room: generator, patch YAML, replacement YAML, anchored zone, runtime deltas. The boot-time `roomIdToFileCache` index of every room file goes away for generated space (an infinite-ish world cannot be indexed at boot). The existing unload logic (`RoomUnloadRounds`, `RoomUnloadThreshold`) stays.

**3. Rules extraction.** Introduce a `Ruleset` interface the core calls for: attack resolution, damage, regen per round, death, derived values, what traits an entity starts with. Move `internal/combat` calculations, `internal/stats`, `internal/races` and the stat parts of `skills`, `spells`, `buffs` behind it. The coupling counts (6 to 27 importing files per package) say this is a few weeks of focused work, not a rewrite. `buffs` (27 files) is the widest and should go last, kept as an engine-level "timed affect" with ruleset-defined effects.

**4. Traits.** Replace the fixed six-field `Statistics` struct with a keyed traits container (static, counter, gauge, in the Evennia sense), defined by the ruleset in YAML. Earthdawn's attribute steps, karma and wounds later become trait definitions, with no core change.

**5. Geo bake and generator.** The bake is an offline tool, run when source data changes, producing versioned tile files plus a fingerprint. GDAL does the heavy GIS work at bake time (CLI or a short script), so the server stays pure Go with no cgo. At runtime the generator samples tiles, applies seeded procedural water constrained to the GIS coastline, and returns a whole room in one call. Fingerprint tests over sample regions run in CI to catch generator drift.

**6. Time and deltas.** Baseline is f(room key, seed, T): description fragments by biome and season, spawn tables, NPC schedules. Deltas are timestamped records with a decay rule. GoMud's `Mutator` (`SpawnedRound`, `DespawnedRound`, text prepend/append/replace) is the starting shape; it gains a decay rule and moves to SQLite. NPCs seen again are computed at the current T from their last-seen state, not stepped forward.

**7. Region tier.** New coarse-grid simulation for weather, fire and flood, running every N rounds on the main worker. Rooms read region state on load; `Biome` already exists as the hook.

**8. Persistence.** SQLite through a pure-Go driver, so the single binary survives. Players and deltas move there; authored content stays YAML in git. Postgres is not needed at this scale and would add a service.

**9. Builder conveniences.** Extend `prompt.Prompt` into a proper menu state machine (EvMenu style: nodes, options, goto). Extend `templates.GetTable` with column widths, wrapping and a form layout (EvForm style). Add an `inspect` admin command that shows, for any room, which provider layer supplied each field.

**10. Protocol.** Define the outbound event structs once, with a schema file the clients can generate from; GMCP and JSON are encoders over the same structs. The existing `modules/gmcp` packages become the first encoder.

### Recommended answers to the open questions

| Question | Recommendation | Why |
|---|---|---|
| Room density | Variable: a base grid in wilderness (for example 1 km cells), subdivided near roads, rivers and settlements | Diku feel needs rooms to mean something; uniform dense grids produce thousands of identical rooms |
| Wilderness movement | Room-to-room on the generated graph | Diku combat, follow, flee and tracking all assume discrete rooms; continuous coordinates break all of them |
| Database | SQLite | Embedded, single binary, enough for a MUD's write rate |
| Rust integration | None. Bake offline | Removes the NIF risk; revisit only if profiling demands it |
| Protocol schema | JSON Schema generated from Go structs | One source of truth, consumable by JS, Rust and Godot clients |
| First anchored area | A small port town with a dock seam onto procedural water | Proves land, sea and anchor seams in one place |

### Phases

| Phase | Work | Done when |
|---|---|---|
| 0. Fork and baseline | Fork upstream at current HEAD; confirm MIT; build; measure idle memory and memory with 50 scripted bots | Baseline numbers recorded in the repo |
| 1. Room identity | `RoomId int` → `RoomKey` | Upstream test suite passes unchanged in behaviour |
| 2. Rules extraction | `Ruleset` interface, traits container, diku-basic reference ruleset | A fight in the test region runs entirely through the ruleset; swapping in a stub ruleset changes the outcome with no core edit |
| 3. Provider chain | Generalise template/instance into the chain; drop boot index for generated space | `inspect` shows per-layer contributions on a patched generated room |
| 4. Geo bake and generator | Bake tool, tile format, sampler, procedural water, CI fingerprints | A sample region walks end to end; fingerprints stable across runs |
| 5. Time, deltas, region tier | f(T) baseline, decaying deltas in SQLite, weather region sim | A dropped item and a burned room decay on schedule while unloaded |
| 6. Builder kit and protocol | Menus, tables/forms, event schema | A chargen menu and a GMCP room event built only from engine helpers |
| 7. Anchored test area | Port town with dock seam | Walk from GIS land through the town onto water and back |

Phases 1 and 2 are where most of the risk sits. Phase 1 is mechanical (Sonnet-grade work). Phase 2 needs judgement.

### Risks

| Risk | Mitigation |
|---|---|
| Upstream merges stop once the fork diverges | Accept it: treat the fork as a one-time base. Phase 1 lands first so the few early upstream fixes still cherry-pick cleanly |
| Generation inside the single `MainWorker` loop stalls the pulse | Generate on a worker goroutine with a cache; the room appears on the next turn (50 ms) |
| Rules extraction leaks (a stat name hard-coded somewhere) | A guard test that greps core packages for the reference ruleset's trait names |
| Generator drift | Versioned bake output plus fingerprint tests in CI, as the original doc says |
| Go GIS ecosystem is thinner than Rust's | Only matters at bake time, where GDAL does the work regardless of language |

---

## Decisions for you

**Project: new_geo_mudbase.** Effect for players: none yet; this sets what the engine is built from.

1. **Base codebase.** Upstream GoMud (recommended) or Kalevala as the original doc proposed. GoMud locks in Go as the only server language and a hard fork. Kalevala keeps the Elixir path open but means writing the Diku runtime yourself.
2. **Wilderness movement.** Room-to-room on a variable-density grid (recommended) or continuous coordinates. This locks in the room ID scheme, so it has to be settled before Phase 1.

If you agree with both recommendations, the next step is a Phase 0 and Phase 1 implementation plan.
