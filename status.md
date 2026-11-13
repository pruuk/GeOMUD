# Status (updated 2026-11-12)

## Now
- Phase 5 slice 4.1 (round and death leftovers, retaliation and target switching), branch
  `strip/phase-5-4-1-leftovers` (`f8e96841`, WIP). Spec and plan at that branch's root. Spec rulings
  1-13 approved by Calabe; ruling 13 (shooting is one shot per action) is recorded but not yet
  applied to the spec's design text or the plan. Plan through blind review round 6, not yet
  approved; round 6's fixes are applied

## Next
1. Apply ruling 13 to the spec design (lookup step 3, the shooter re-arm) and the plan, settling its
   open detail (does each shot received earn one shot back; does the aggro then wait or end), then
   blind review round 7 and on until a round finds nothing material, then Calabe reviews the plan
2. Execute slice 4.1 subagent-driven, verify, PR
3. Slices 4.2 (progression) and 4.3 (the body): each brainstorm, spec, plan, PR
4. Phase 5 slices 5-7: engine knows health only; timed affects; close out
5. Hidden actors' own action lines (issue pruuk/GeOMUD#33); Phase 6 measure and tag; security pass
   before any public server (upstream CodeQL debt, `internal/util/util.go:384`, `:706-752`)

## Waiting on Calabe
- Ruling 13's open detail, if it is a player-visible choice (spec ruling 13)

## Watch
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference. Last
  run 2026-11-12 on the plan applied to a scratch copy: 41 packages ok plus those 3; smoke passed
  with 1, 5 and 12 bots
- `hooks` `TestCombatGoldenFleeSuccess` fails under `-shuffle=on` and on `-count` repeats, on master
  too: order-dependent shared state, not yet diagnosed
- Copyover runs only on Linux: the smoke copyover step (player and NPC IDs survive) is CI-only
- Moving a player to the recovery room panics when `DeathRecoveryRoom` is 0 (`ephemeral.go:188`;
  package tests load no config; shipped room is 75). Help pages in markdown join their usage lines
  in a Windows (CRLF) checkout; git stores LF, so servers render them correctly
- Known engine and world issues, most from upstream: `docs/known-issues.md`
