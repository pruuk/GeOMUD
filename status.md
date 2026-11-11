# Status (updated 2026-11-10)

## Now
- Phase 5 slice 3 (round, regen and death through the seam): spec approved, plan drafted, blind plan
  review round 2 done (`strip/phase-5-3-round-death`, pushed). Round 2's open items are in the spec
  ("Plan review round 2"); not yet applied to the plan

## Next
1. Slice 3: Calabe answers the two round 2 questions, the plan takes round 2's fixes, blind review
   round 3, then Calabe reviews the plan (`geomud-workflow`)
2. Phase 5 slices 4-7 (`2026-11-06-phase-5-rules-seam-design.md`): shrink to the bare ruleset (strip
   plan decisions 5-11, plus offhand breakage and `BackStab`); engine knows health only; timed
   affects; close out
3. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world; docs
   sweep, including the "early in development" line in `help about`
4. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- `bury` on a corpse that holds items or gold: refuse, spill to the floor, or owner only
  (slice 3 spec, "Plan review round 2")
- Two corpses with the same name: fix the lookup (ordinals, exact match first) or accept (same place)

## Watch
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- `hooks` `TestCombatGoldenFleeSuccess` fails under `-shuffle=on` and on `-count` repeats, on master
  too: order-dependent shared state, not yet diagnosed
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Moving a player to the recovery room panics when `DeathRecoveryRoom` is 0 (`ephemeral.go:188`;
  package tests load no config; shipped room is 75)
- Known engine and world issues, most from upstream: `docs/known-issues.md`
