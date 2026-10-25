# Geo-Diku MUD Foundation — Requirements Synopsis

A short summary of the requirements as currently understood. Companion to `mud-architecture.md`.

---

## Purpose

Build a reusable MUD **foundation** first, then layer a **rules-as-faithful-as-possible Earthdawn** game on top of it. The foundation should not assume Earthdawn rules. This is a separate project from DOGMud; nothing is being ported.

---

## Core Requirements

### Gameplay feel
- **Diku-style, "twitchy" real-time play**: steady pulse, combat rounds, responsive output. Not turn-based.
- Timing must stay responsive under load.

### Resources
- **Light resource footprint.** This was the main reason for choosing GoMud over Evennia for DOGMud, and it remains a hard requirement here.

### World & geography
- **Real geography from GIS sources** (GIS files, ideally authored or read from QGIS), imported as needed.
- **Land rooms generated deterministically** from GIS data.
- **Water and maritime areas generated procedurally** where GIS data is thin or unavailable, using concepts from Worldbuilder by Aetos.
- **Hand-built content overrides generated content**, added gradually over time without migrating or rewriting what exists.

### Dynamic, time-aware world
- **Only rooms with active players (or active state) are loaded.**
- **Rooms change subtly over time**, with the changes applied when they load.
- **Crafted areas support alternate descriptions** for weather, season, and other conditions (ExtendedRoom style).
- **NPCs change over time** according to how much time has passed since they were last seen.

### Builder & developer conveniences (Evennia-inspired, rebuilt)
- **Traits-style stat system** (used heavily in past work).
- **Menu system** (EvMenu-style state-machine dialogs).
- **Table and form text layout helpers** (EvTable/EvForm-style).
- **Solid persistence.** Evennia's built-in DB was valued, but YAML may be preferable for content; see Storage.

### Clients
- **Client-agnostic server.** A web client can be written in JS, Rust, or anything else.
- A structured protocol so classic and graphical clients can coexist.

---

## Constraints & Principles

- **Borrow ideas, not code**, from Evennia, Worldbuilder by Aetos, Kalevala/ExVenture, Genesis, and GoMud.
- **Foundation before rules.** Earthdawn mechanics plug in later and must not be baked into the core.
- **Fit the language to the job.** Using more than one language is acceptable if each one earns its place.

---

## Current Direction (proposed, not final)

- **Rust**: world math core (GIS ingest and sampling, procedural water, ID scheme, generation, pathfinding).
- **Elixir/BEAM**: runtime shell (connections, rooms, characters, world clock, region simulation). Likely a fork of Kalevala, with Genesis/Qud-style event components for actions and Bevy-style sweeps for per-tick systems.
- **Storage**: YAML in git for authored content; generated content computed on demand rather than stored; a database for persistent runtime state such as players and world deltas; memory for transient state.

---

## Explicitly Out of Scope (for now)

- A particular ruleset's implementation

---

## Open Questions
- Room density: uniform grid, or denser near settlements and roads?
- Wilderness movement: room-to-room on a generated graph, or continuous coordinates?
- Database: SQLite or Postgres?
- Rust integration: NIF (Rustler) or a Port/sidecar process?
- Protocol schema format shared across server and clients.
- First hand-built anchored area to prove the override/seam model.