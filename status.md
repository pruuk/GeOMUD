# Status (updated 2026-10-26)

## Now
- Phase 3 (world data) done on branch `strip/phase-3-world-data`, waiting on its PR

## Next
1. Phase 4: scripting, web and GMCP surface. The tutorial is all room scripts, so the smoke test needs
   a new per-player item source here. Also decide on Mudlet "Discord Rich Presence" GMCP
   (`modules/gmcp/gmcp.Mudlet.go`, about 450 lines), kept through Phase 1
2. Phase 4b: NPCs are players (decision 4, 2026-10-26). Design spec for Calabe first
3. Phase 5: rules seam. `suicide` (player and mob death) and the combat commands move here, with
   the `professions`, `skills`, `spells` world folders, `set shopstyle`/`set wimpy`, the `help spell`
   branch and the loot goblin config code
4. Phase 6: measure. Also measure the Phase 0 engine (`2737986c`) on the trimmed world, since the
   Phase 0 baseline ran on the deleted `default` world

## Waiting on Calabe
- Go-ahead to start Phase 4 (`gomud_strip_plan.md`)
- Whether to rewrite `templates/about.template` (still GoMud's text, credits and Discord) for GeOMUD

## Watch
- Phase 5 rules seam: `buffs` has 13 importers and `characters` imports every rules package
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Upstream engine bugs, not fixed (`docs/phase-0-baseline.md`): concurrent signups collide on user ID;
  a line sent within 50 ms of the previous one, or of the login prompt, is silently dropped; "Mob not
  found" errors when an ephemeral room is cleaned up while a script still has delayed commands queued
  for its mob (`mobs.DestroyInstance` does not mark the mob as recently dead; goes with Phase 4)
- Stale permission keys (`quests.write`, `pets.write`, `conversations.write` and others) stay in saved
  mod accounts; they load fine, but a full permission-list save that still carries one is rejected
  with a 400. Clear them when the permission editor is next touched
- `empty` world references missing data: buff 22 (wildfire mutator), item 30001 (small red potion)
  points at missing buff 5; the sample admin `users/1.yaml` names zone Frostfang
- `exits` command file exists but was never registered (upstream too); the plan lists it as kept
- Upstream security debt (CodeQL, 30 open alerts on master, all from upstream): passwords are hashed
  with unsalted SHA-256 (`internal/util/util.go:384`); file reads and writes take user-provided paths
  (`util.go:706-752`); unchecked integer narrowing (`mobs.go:91,97`, `main.go:884`). A PR that edits
  these files shows them as "new" alerts. Needs its own pass before any public server
