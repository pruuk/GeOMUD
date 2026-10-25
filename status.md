# Status (updated 2026-10-24)

## Now
- Nothing in flight. Strip plan (`gomud_strip_plan.md`) not started

## Next
1. Disable Dependabot: delete `.github/dependabot.yml` (version updates are the only Dependabot feature on; alerts and security fixes are already off). Ask Calabe about `update-go-toolchain.yml` too: it also opens PRs, monthly on the 25th
2. Phase 0: build the smoke test, record the memory baseline, delete upstream's root `AGENTS.md`
3. Phase 1: delete leaf packages
4. Phase 2: delete game commands and hooks
5. Phase 3: trim world data

## Waiting on Calabe
- Go-ahead to start Phase 0 (`gomud_strip_plan.md`)

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
