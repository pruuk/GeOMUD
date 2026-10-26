# Status (updated 2026-10-26)

## Now
- Nothing in flight. Strip plan (`gomud_strip_plan.md`) not started

## Next
1. Phase 0: build the smoke test, record the memory baseline, delete upstream's root `AGENTS.md`
2. Phase 1: delete leaf packages
3. Phase 2: delete game commands and hooks
4. Phase 3: trim world data

## Waiting on Calabe
- Go-ahead to start Phase 0 (`gomud_strip_plan.md`)

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
