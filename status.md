# Status (updated 2026-11-06)

## Now
- Nothing in flight. Phase 5 slice 1 (delete dead rules systems) merged (#29); plan in
  `docs/completed/2026-11-06-phase-5-1-dead-systems-plan.md`

## Next
1. Phase 5 slice 2: the seam and the fight (`internal/rules`, `internal/rules/dikubasic`, combat
   moved verbatim, 9 combat goldens byte-identical, stub-ruleset test, `shoot` through `CanAttack`).
   Plan first, with a verified facts table; read the design's "Notes for later slices" and inventory
   N (`2026-11-06-phase-5-inventory.md`)
2. Phase 5 slices 3-7 (`2026-11-06-phase-5-rules-seam-design.md`): round and death through the seam;
   shrink to the bare ruleset (strip plan decisions 5-11); engine knows health only; timed affects;
   close out
3. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world; docs
   sweep, including the "early in development" line in `help about`
4. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- Go for slice 2

## Watch
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Permadeath reroll is tested but never run live (permadeath off); slice 4 deletes permadeath
- Phase 5 seam: `buffs` has 11 importing packages; `characters` imports five rules packages (buffs,
  races, skills, statmods, stats)
- Known engine and world issues, most from upstream: `docs/known-issues.md`
