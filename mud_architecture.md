# Geo-Diku MUD Foundation — Architecture Notes

A working reference for a Diku-timed MUD foundation built on real geography, designed to be rules-agnostic first (Earthdawn comes later).

> **One-line summary:** Rust computes the world, Elixir runs it, YAML defines it, the database remembers what players changed, and clients just speak the protocol.

---

## Goals

- **Diku-style timing** — a steady pulse, combat rounds, twitchy and responsive feel.
- **Real geography** — land derived from GIS data (QGIS / GeoPackage / DEM), water filled in procedurally where data is thin.
- **Lazy, dynamic world** — only rooms with players (or active state) are live; rooms are computed on load, not stored.
- **Time-aware world** — rooms and NPCs reflect elapsed time when loaded.
- **Hand-built content overrides generated content**, gradually, without migrations.
- **Light resource footprint.**
- **Rules-agnostic foundation** — the rules engine (Earthdawn) plugs in later.

---

## Layer Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Clients: telnet (GMCP) · web (JS/WASM) · Godot             │
└──────────────────────────┬──────────────────────────────────┘
                           │ structured protocol (JSON / GMCP)
┌──────────────────────────▼──────────────────────────────────┐
│  ELIXIR RUNTIME (owns state + time)                         │
│   • World clock / pulse                                     │
│   • Connection processes (Foreman-style)                    │
│   • Room & zone processes (lazy start/stop)                 │
│   • Region simulation processes (weather, spread effects)   │
│   • Components: Genesis-style events + Bevy-style sweeps    │
└──────────────┬───────────────────────────────┬──────────────┘
               │ NIF calls (Rustler)           │ reads/writes
┌──────────────▼──────────────┐   ┌────────────▼──────────────┐
│  RUST WORLD CORE            │   │  STORAGE                  │
│  (owns heavy math)          │   │   • YAML (authored, git)  │
│   • GIS ingest & sampling   │   │   • DB (persistent state  │
│   • terrain_z_at providers  │   │     + world deltas)       │
│   • procedural water        │   │   • Memory (transient)    │
│   • room ID scheme          │   │   • GIS source files      │
│   • pathfinding             │   │                           │
└─────────────────────────────┘   └───────────────────────────┘
```

**Rule of thumb:** *Elixir owns state and time. Rust owns heavy math.* Anything moves to Rust only when profiling says it must.

---

## 1. Rust World Core

**Owns:** deterministic, mostly stateless, compute-heavy work.

- GIS ingest (GeoPackage, DEM rasters, GEBCO/ETOPO bathymetry)
- Point sampling — `terrain_z_at(lat, lon)`-style provider interfaces
- Procedural water: low-frequency data base + procedural detail + placed features (marks layer)
- Coastline reconciliation (GIS coast authoritative; procedural seabed constrained to it)
- Room ID scheme — IDs derived from geography, never from generation order
- Room generation from geometry (settlements → room clusters, roads/rivers → edges)
- Pathfinding over the generated graph

**Why Rust:**
- Mature geo crates (`geo`, `gdal`, `proj`, `geozero`) — no reinventing GIS.
- Precise control over determinism (floating point behavior, fixed-point where needed).
- Fast enough to answer point queries thousands of times per second.
- Can compile to WASM, so a builder/studio tool sees the *exact* same world as the server.

**Pros**
- Heavy math never stalls the game loop's logic.
- Pure functions are trivially testable; determinism can be fingerprinted in CI.
- Reusable outside the MUD (combat sim, map tools, Godot client via WASM or GDExtension).

**Cons**
- Two-language project: build tooling, FFI boundary, two mental models.
- NIF safety: a long-running NIF blocks a BEAM scheduler; a native crash can take down the whole VM.
- Iteration on world-gen logic is slower than in Elixir.

**Mitigations**
- Use **Rustler** (catches Rust panics instead of crashing the VM).
- Run anything non-trivial on **dirty schedulers**.
- Keep the NIF surface small and coarse-grained (one call returns a whole room, not 200 tiny calls).
- Fallback option: run the core as a **Port / sidecar process** if NIF risk ever feels too high (costs latency, buys isolation).

**Alternatives considered**
- *Go core:* reuses DOGMud code, but weaker determinism story and cgo for GDAL.
- *Goose core:* compelling memory model and verified binary chunk format; too new, small ecosystem, C NIF wrapping needed.
- *All-Elixir:* simplest stack, but BEAM is poor at numeric/raster work.

---

## 2. Elixir Runtime Shell

**Owns:** connections, rooms, zones, characters, process lifecycle. Base: **fork of Kalevala** (MIT).

- **Foreman-style process per connection** — input, output, events.
- **Conn token** — batches renders and events per input/event.
- **Controllers as state machines** — login, command mode, pager, chargen menus.
- **Command router DSL** — pattern-matched parsing.
- **Views** — text and out-of-band events (GMCP / JSON).
- **Room processes** — started on demand, hydrated from generator + deltas, stopped when idle.
- **Channels** — pub/sub for room speech and global channels.
- **NPC brains** — behavior trees.

**Why Elixir / BEAM:**
- Process-per-room maps directly onto lazy loading.
- Preemptive soft-real-time scheduler → consistent latency under load (good for twitchy feel).
- Supervision: a crash in one room restarts that room, not the server.
- Hot code reload for iterating without kicking players.
- Clustering available later if ever needed.

**Pros**
- Concurrency and fault isolation are the language's core strengths, not bolt-ons.
- Kalevala already solves the boring parts (telnet, TLS, web client, GMCP, controllers).
- Phoenix available for web client hosting and admin tooling.

**Cons**
- Kalevala appears dormant (ExVenture last pushed early 2023) — you own the fork.
- Dynamic typing; large rules codebases need discipline (typespecs, Dialyzer, tests).
- Everything routes through room processes → cross-room effects need explicit messaging.
- Less familiar than Go; some ramp-up.

**Alternatives considered**
- *GoMud fork:* known codebase, MIT, Diku timing built in — but room-graph world model and no process isolation.
- *Evennia:* best toolbox (traits, EvMenu, EvTable), but single-threaded Twisted fights Diku pulse timing and resource goals.
- *Custom Go server (DOGMud-style):* full control, but you rebuild supervision/isolation yourself.

---

## 3. Components & Rules Layer (Elixir)

**Owns:** behavior. Two complementary patterns:

### 3a. Genesis / Caves-of-Qud style — discrete actions
Components hold data *and* behavior; events flow through an entity's components, each able to modify and pass along (`{:cont, event}`).

- Attacks, spellcasting, skill use, item interactions.
- Later: talents, thread items, karma, wound modifiers as components that intercept action events.

**Pros:** new abilities are new components, not edits to core combat code; models "this adds +2 when…" rules cleanly; very data-driven.
**Cons:** event ordering becomes a design concern; debugging long pipelines needs good tracing.
**Note:** Genesis itself is AGPL-3.0 — borrow the *pattern*, write your own implementation (AGPL would force publishing server source).

### 3b. Bevy style — per-tick sweeps
Systems iterate over component tables each pulse.

- Regen, affect/duration timers, hunger/thirst, wound recovery.

**Pros:** natural fit for "every entity with X, every tick"; simple to reason about.
**Cons:** on the BEAM there's no cache-locality win — choose it for clarity, not speed. Large sweeps may need to move to Rust.

---

## 4. Region Simulation Tier (Elixir)

**Owns:** effects that span many rooms, including unloaded ones.

- Weather (region-resolution, like the GoMud weather director over a geography graph)
- Spreading effects: fire, flood, taint/corruption
- Seasonal/environmental state

Rooms *read* region state when they load; regions never need individual rooms to be live.

**Why a separate tier:**
- Pure lazy loading can't propagate effects across rooms nobody is standing in.
- Coarse resolution keeps it cheap enough to run always.

**Pros:** world stays coherent while 99% of rooms are unloaded; rooms stay simple.
**Cons:** two levels of resolution to keep consistent; region boundaries need care.
**Escape hatch:** heavy region math (e.g., hydrology) can call into Rust.

---

## 5. World Clock & Diku Timing (Elixir)

**Owns:** authoritative game time and the pulse.

- One **world clock** process — monotonic, decoupled from wall-clock.
- Pulses broadcast only to **active zones/regions**.
- Combat runs on **per-fight round timers**.
- Regen/affects tick off zone pulses.

**Pros:** single source of truth for time-as-a-function (see §8); consistent Diku feel; inactive areas cost nothing.
**Cons:** pausing, time acceleration, or DB rollback shift *every* time-derived value — must be deliberate.

---

## 6. Clients & Protocol

**Owns:** presentation only. Clients never own game state.

- Telnet with GMCP for classic MUD clients.
- Web client (JS, or Rust→WASM) — can be hosted by Phoenix.
- Optional Godot client (reuse DOGMud learnings).

**Key decision:** define a **structured, typed protocol early** (as with DOGMud's typed JSON over WebSocket). The same event structs feed GMCP and JSON.

**Pros:** clients are swappable; graphical/map clients become possible without server changes.
**Cons:** protocol versioning becomes a maintenance burden once multiple clients exist.

---

## 7. Storage Tiers

| Tier | Holds | Lives in | Survives restart? |
|---|---|---|---|
| **Authored** | prototypes, hand-built rooms, patches/overrides, (later) rules definitions | YAML in git | Yes |
| **Generated** | rooms/terrain derived from GIS + seeds + time | Nowhere — computed on demand | Recomputed |
| **Persistent runtime** | players, inventories, world deltas (dropped items, burned bridge, dead shopkeeper) | Database | Yes |
| **Transient** | active combats, live room process state, caches | Memory (process state / ETS) | No — rebuilt from other tiers |
| **Source data** | GeoPackages, DEM & bathymetry rasters | Files beside the YAML | Yes |

**Why this split:**
- YAML: diffable, reviewable, version-controlled, easy to bulk-edit with Claude Code.
- Generated: storage stays tiny no matter how big the world is.
- DB: runtime state changes constantly and needs transactions/queries — flat files get awkward here.
- Memory: transient things shouldn't pay persistence cost.

**Pros:** each kind of data lives where it's cheapest and safest; huge world, small footprint.
**Cons:** four places to look when debugging "why is this room like this?" — build an inspector command that shows which tier contributed what.

**DB choice:** open question — SQLite (simple, embedded) vs Postgres (Ecto's home turf, better concurrency).

---

## 8. Time-as-Function & Lazy Loading

**Model:** `room_state(T) = baseline(room_id, seed, T) + active_deltas(T)`

- **Baseline** is a pure function of time — seasons, weathering, NPC schedules. Same answer for every observer regardless of visit frequency.
- **Deltas** are player-caused changes, timestamped, each with a **decay rule** back toward baseline (corpses rot, trails regrow) — the YAML decay-chain idea generalized.
- Condition-keyed descriptions: ExtendedRoom-style variants for crafted rooms; biome fragment pools for generated rooms (seeded by room ID + season so they don't flicker).

**Why compute-at-T rather than step-forward-on-load:**
Stepping forward makes the world depend on how often rooms are visited. Direct computation gives one consistent history.

**Pros:** consistent, cheap, scalable; NPCs and rooms "age" naturally.
**Cons:** anything not expressible as f(T) must be a delta; designing decay rules takes thought.

---

## 9. Content Layering & Overrides

Precedence, lowest → highest:

1. **Generated base** — GIS land, GEBCO + procedural water
2. **Patches** — YAML keyed by room ID, overriding selected fields (desc, exits, spawns)
3. **Replacements** — fully hand-built rooms at a generated ID
4. **Anchored areas** — whole hand-built zones attached via **seam rooms** (dock, gate, ferry); local coordinates authoritative inside (concept from Worldbuilder by Aetos)

**Pros:** world is complete on day one and improves incrementally; no migrations ever.
**Cons:** depends entirely on stable, geography-derived IDs; generator changes must be versioned.

---

## Cross-Cutting Risks

| Risk | Why it matters | Mitigation |
|---|---|---|
| **Generator drift** | Any generator change silently changes the world, breaking overrides and saved locations | Pin seeds + data versions; fingerprint regression tests over sample regions in CI |
| **ID instability** | IDs from generation order break every override on change | IDs derived purely from geography/coordinates |
| **Coast seam** | Land (GIS) and sea (procedural) disagreeing | GIS coastline authoritative; automated "no rooms underwater / no land at sea" check |
| **NIF failure** | Native crash or long NIF stalls/kills the VM | Rustler, dirty schedulers, coarse calls, Port fallback |
| **Owning a dormant fork** | No upstream fixes for Kalevala | Treat as a foundation, not a dependency; keep changes well-documented |
| **Two-language overhead** | Context switching, tooling | Keep the boundary narrow and stable |
| **Debuggability** | State assembled from four tiers + time | Admin inspector showing per-tier contributions |

---

## Concepts Borrowed (Ideas, Not Code)

| Source | Idea |
|---|---|
| Kalevala / ExVenture | Foreman, Conn, controllers, command router, views, room processes, crash containment |
| Genesis / Caves of Qud | Event-pipeline components with data + behavior |
| Bevy | System sweeps over component tables |
| Worldbuilder by Aetos | Anchors & seams; provider interface (`terrain_z_at`); determinism as a design constraint; marks layer for hazards; procedural detail as texture, features as placed objects |
| Evennia | Traits, EvMenu, EvTable/EvForm, ExtendedRoom state descriptions, prototypes/spawner |
| Ataxia | Portal/engine split (only relevant if hot reload ever needs process restarts) |
| GoMud / own work | Weather director over geography graph; YAML decay chains |

---

## Open Questions

- **Room density:** uniform grid, or variable (dense near settlements/roads, sparse in wilderness)? Drives the ID scheme and seams.
- **Movement model in wilderness:** room-to-room on a generated graph, or continuous coordinates with rooms as views?
- **DB:** SQLite vs Postgres.
- **Rust integration:** NIF (Rustler) vs Port/sidecar.
- **Protocol schema:** hand-written structs, or a schema language (JSON Schema, protobuf) shared across Elixir, Rust, and clients?
- **First hand-built anchored area** to prove the seam model.
- **How much of Kalevala to keep** vs. rewrite once the fork is yours.