# Status (updated 2026-10-27)

## Now
- `help about` rewritten for GeOMUD (branch `docs/about-template`). Strip Phases 0-4 are merged
  (`gomud_strip_plan.md`)

## Next
1. Phase 4b: NPCs are players (decision 4). Starts with a design spec for Calabe's review
2. Still branded GoMud for players: the login splash
   (`_datafiles/world/empty/templates/login/connect-splash.template`: "Welcome to GoMud", GoMud's
   repo link, GoMud ASCII logo) and `Server.MudName: "GoMud"` in `_datafiles/config.yaml`
3. Phase 5: rules seam. Takes over the death stand-in (revive at 1 health, portal out of room 75;
   `docs/completed/2026-10-26-phase-4-scripting-surface.md`), `suicide`, the combat commands, the
   `professions`/`skills`/`spells` world folders, `set shopstyle`, `help spell` and the loot goblin
   config. Spells and scripted buffs do nothing until then
4. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world, since
   the Phase 0 baseline ran on the deleted `default` world. Docs sweep: recheck the "early in
   development" line in `help about` (Calabe kept it 2026-10-27, knowing it goes stale)
5. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- Go-ahead to start the Phase 4b design

## Watch
- Phase 5 seam: `buffs` has 11 importers; `characters` imports six rules packages (buffs, races,
  skills, spells, statmods, stats)
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Upstream engine bugs, not fixed (`docs/phase-0-baseline.md`): concurrent signups collide on user ID;
  a line sent within 50 ms of the previous one, or of the login prompt, is silently dropped
- Stale permission keys (`quests.write`, `pets.write` and others) in saved mod accounts: they load,
  but a full permission-list save that carries one is rejected with a 400
- `empty` world loose ends: the `wildfire` mutator names missing buff 22; item 30001 names missing
  buff 5; the `exits` command file was never registered (upstream too)
