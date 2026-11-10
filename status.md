# Status (updated 2026-11-10)

## Now
- Phase 5 slice 2 (the seam and the fight) done on `strip/phase-5-2-seam-fight`; PR to merge once CI
  is green. Spec and plan move to `docs/completed/` with the merge

## Next
1. Phase 5 slice 3: round, regen and death through the seam (`Condition`, `Regen`, `RoundUpkeep`, one
   death routine calling `OnDeath`). Follows `geomud-workflow`: brainstorm with Calabe first. Read
   the design's "Notes for later slices" (slice 2 added several)
2. Phase 5 slices 4-7 (`2026-11-06-phase-5-rules-seam-design.md`): shrink to the bare ruleset (strip
   plan decisions 5-11, plus offhand breakage and `BackStab`); engine knows health only; timed
   affects; close out
3. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world; docs
   sweep, including the "early in development" line in `help about`
4. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- Go for slice 3 (brainstorm)

## Watch
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- `hooks` `TestCombatGoldenFleeSuccess` fails under `go test -shuffle=on` (and on `-count=3` repeats),
  on the pre-slice-2 baseline too: order-dependent shared state, not yet diagnosed
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Permadeath reroll is tested but never run live (permadeath off); slice 4 deletes permadeath
- Phase 5 seam: `buffs` has 11 importing packages; `characters` imports five rules packages (buffs,
  races, skills, statmods, stats); `characters` cannot call `rules` (cycle), so the weapon delay
  stays engine code until swing speed
- Known engine and world issues, most from upstream: `docs/known-issues.md`
