# Status (updated 2026-10-24)

## Now
- Strip plan written (`gomud_strip_plan.md`); not started. `master` is upstream `b3f55cc2` plus one commit of docs and skills, pushed to `origin`

## Next
1. Phase 0: first commit, settle the `scripts` test failure, build the smoke test, record memory baseline
2. Phase 1: delete leaf packages
3. Phase 2: delete game commands and hooks
4. Phase 3: trim world data
5. Phase 4: strip scripting, trim web and GMCP

## Waiting on Calabe
- Go-ahead to start Phase 0

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
