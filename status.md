# Status (updated 2026-10-26)

## Now
- Phase 1 (leaf packages) done on branch `strip/phase-1-leaf-packages`, waiting on its PR

## Next
1. Phase 2: delete game commands and hooks
2. Phase 3: trim world data (the smoke script plays the stock tutorial; replace those steps here).
   Also removes world data left by Phase 1: quests, pets, conversations folders, `audio.yaml`,
   `musicfile` keys, pet and quest help templates and keywords
3. Phase 4: scripting, web and GMCP surface. Also decide on Mudlet "Discord Rich Presence" GMCP
   (`modules/gmcp/gmcp.Mudlet.go`, about 450 lines), kept through Phase 1

## Waiting on Calabe
- Go-ahead to start Phase 2 (`gomud_strip_plan.md`)

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Two upstream engine bugs found by the smoke test, not fixed (`docs/phase-0-baseline.md`): concurrent
  signups collide on user ID; a line sent within 50 ms of the previous one, or of the login prompt, is
  silently dropped
- Stale permission keys (`quests.write`, `pets.write`, `conversations.write` and others) stay in saved
  mod accounts; they load fine, but a full permission-list save that still carries one is rejected
  with a 400. Clear them when the permission editor is next touched
