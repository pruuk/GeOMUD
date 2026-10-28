# Status (updated 2026-10-28)

## Now
- Phase 4b slice 2c (item IDs): PR #22 open, CI green
  (plan `2026-10-28-phase-4b-2c-item-ids-plan.md`)
- Phase 4b slice 3 (messaging core): built on `strip/phase-4b-3-messaging-core`, branched from the
  2c branch (rebase onto master after #22 merges); awaiting Calabe's go to push and open the PR
  (plan `2026-10-28-phase-4b-3-messaging-core-plan.md`)
- Strip Phases 0-4 and Phase 4b slices 1, 2 and 2b are merged (`gomud_strip_plan.md`; spec
  `2026-10-27-phase-4b-npcs-are-players-design.md`)

## Next
1. Phase 4b slice 4 (one command set and input path); its plan reads the spec's "Notes for later
   slices" first
2. Phase 4b slices 5-7 (one command set, one combat path, one tick, output conversion); each plan
   reads the spec's "Notes for later slices" first
3. Phase 5: rules seam. Takes over the death stand-in (revive at 1 health, portal out of room 75;
   `docs/completed/2026-10-26-phase-4-scripting-surface.md`), `suicide`, the combat commands, the
   `professions`/`skills`/`spells` world folders, `set shopstyle`, `help spell` and the loot goblin
   config. Spells, scripted buffs and charm do nothing until then
4. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world; docs
   sweep, including the "early in development" line in `help about`
5. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- Merge PR #22 (slice 2c); push and PR for slice 3

## Watch
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Permadeath reroll and companion persistence are tested but never run live (permadeath off,
  nothing charms); spec "Notes for later slices"
- Phase 5 seam: `buffs` has 11 importers; `characters` imports six rules packages (buffs, races,
  skills, spells, statmods, stats)
- Known engine and world issues, most from upstream: `docs/known-issues.md`
