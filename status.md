# Status (updated 2026-10-26)

## Now
- Phase 0 (baseline) done on branch `strip/phase-0-baseline`, waiting on its PR

## Next
1. Phase 1: delete leaf packages
2. Phase 2: delete game commands and hooks
3. Phase 3: trim world data (the smoke script plays the stock tutorial; replace those steps here)
4. Phase 4: scripting, web and GMCP surface

## Waiting on Calabe
- Go-ahead to start Phase 1 (`gomud_strip_plan.md`)

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Two upstream engine bugs found by the smoke test, not fixed (`docs/phase-0-baseline.md`): concurrent
  signups collide on user ID; a line sent within 50 ms of the previous one, or of the login prompt, is
  silently dropped
