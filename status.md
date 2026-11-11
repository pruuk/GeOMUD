# Status (updated 2026-11-11)

## Now
- Phase 5 slice 3 (round, regen and death through the seam) done on `strip/phase-5-3-round-death`;
  PR to merge once CI is green. Spec and plan in
  `docs/completed/2026-11-10-phase-5-3-round-death-{spec,plan}.md`

## Next
1. Phase 5 slice 4 (`2026-11-06-phase-5-rules-seam-design.md`): shrink to the bare ruleset (strip
   plan decisions 5-11, plus offhand breakage and `BackStab`). Read the design's "Notes for later
   slices" first: slice 3 handed slice 4 NPC regen, the -5/-10 floors, XP, levels and alignment, the
   `characters` health checks and `FindDowned`, and `actors.BodyReplaced`
2. Phase 5 slices 5-7: engine knows health only; timed affects; close out
3. Hidden actors' own action lines (move, look, read, trash) shown to viewers who perceive them:
   issue pruuk/GeOMUD#33
4. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world; docs
   sweep, including the "early in development" line in `help about`
5. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- Nothing

## Watch
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- `hooks` `TestCombatGoldenFleeSuccess` fails under `-shuffle=on` and on `-count` repeats, on master
  too: order-dependent shared state, not yet diagnosed
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Moving a player to the recovery room panics when `DeathRecoveryRoom` is 0 (`ephemeral.go:188`;
  package tests load no config; shipped room is 75)
- Help pages written in markdown (`get.md`) join every usage line into one when read from a Windows
  checkout (CRLF): `internal/markdown` needs two spaces right before LF. Git stores LF, so servers
  and CI render them correctly
- Known engine and world issues, most from upstream: `docs/known-issues.md`
