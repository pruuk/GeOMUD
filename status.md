# Status (updated 2026-10-26)

## Now
- Phase 2 (game commands and hooks) done on branch `strip/phase-2-commands-hooks`, waiting on its PR

## Next
1. Phase 3: trim world data (the smoke script plays the stock tutorial; replace those steps here).
   Also removes world data left by Phases 1-2: quests, pets, conversations folders, `audio.yaml`,
   `musicfile` keys, help templates and keywords for removed commands, and mob idle or angry
   commands naming a removed command (such a mob emotes "looks a little confused" until then)
2. Phase 4: scripting, web and GMCP surface. Also decide on Mudlet "Discord Rich Presence" GMCP
   (`modules/gmcp/gmcp.Mudlet.go`, about 450 lines), kept through Phase 1
3. Phase 4b: NPCs are players (decision 4, 2026-10-26). Design spec for Calabe first
4. Phase 5: rules seam. `suicide` (player and mob death) and the combat commands move here

## Waiting on Calabe
- Go-ahead to start Phase 3 (`gomud_strip_plan.md`)

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Two upstream engine bugs found by the smoke test, not fixed (`docs/phase-0-baseline.md`): concurrent
  signups collide on user ID; a line sent within 50 ms of the previous one, or of the login prompt, is
  silently dropped
- Stale permission keys (`quests.write`, `pets.write`, `conversations.write` and others) stay in saved
  mod accounts; they load fine, but a full permission-list save that still carries one is rejected
  with a 400. Clear them when the permission editor is next touched
