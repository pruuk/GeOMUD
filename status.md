# Status (updated 2026-11-14)

## Now
- Phase 5 slice 4.1 (round and death leftovers) is executed on `strip/phase-5-4-1-leftovers`; PR #37,
  waiting on CI

## Next
1. Phase 5 slice 4.2 (progression: levels, XP, alignment, training and stat points, the PvP minimum
   level and their displays), then slice 4.3 (the body: the stat block, `statmods`, mana, race
   fields, action points and carry capacity, skills, the simple hit roll with weapon damage, which
   also decides the crit cascade). Brainstorm each with Calabe first (`geomud-workflow`); read the
   design's "Notes for later slices"
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
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference. Last
  run 2026-11-14 on `5d6bbe71`: 41 packages ok plus those 3; smoke passed with
  1, 5 and 12 bots
- `hooks` `TestCombatGoldenFleeSuccess` fails under `-shuffle=on` and on `-count` repeats, on master
  too: order-dependent shared state, not yet diagnosed
- Smoke -addr mode (a server someone else started) can time out its fight step: against that
  world's own rats, spawned at level 1 to 5, about 1.7% (level 4) and 2.4% (level 5) of fights run past
  60 rounds, the 4-minute fightStepTimeout at 4-second rounds (simulated). Launch mode, which CI runs,
  no longer has this: since slice 4.1 its world copy pins the rats at levels 1 and 5 (ruling 11), and
  its fights end within 17 rounds. Pre-existing. Fix candidate for -addr mode: raise the timeout
- 12-bot smoke on Windows sometimes stalls: the server answers no input for 10-15 s and bots time
  out on simple steps (drop, sign-up), not in the fight. Seen 2026-11-14 on the slice 4.1 tree (4 of
  5 runs) and on `b8ce7284` before its retaliation commit (1 of 4); a later sequential A/B passed 4
  of 4 on both. Cause not found
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Moving a player to the recovery room panics when `DeathRecoveryRoom` is 0 (`ephemeral.go:188`;
  package tests load no config; shipped room is 75). Help pages in markdown join their usage lines
  in a Windows (CRLF) checkout; git stores LF, so servers render them correctly
- Known engine and world issues, most from upstream: `docs/known-issues.md`
