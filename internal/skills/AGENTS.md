# Skills Package Guide

## Scope

- Use this file for skill definitions, progression behavior, and skill-side helpers in `internal/skills`.
- Skill changes are player-facing and often couple to commands, combat, and spells.

## Data-driven model

- Skills are YAML datafiles (`_datafiles/world/default/skills/`), loaded via `fileloader.LoadAllFlatFiles` into the `allSkills` map. See `skill.go`, `admin.go`.
- Skills are keyed by lowercase string ids that match `Character.Skills` keys, so there is no save migration. `maxlevel` is per-skill (default 4); progression math generalizes to `max*(max+1)/2` points per skill.
- Validation policy: zero reads. `GetSkillLevel` returns 0 for unknown skill ids. Orphaned save entries stay inert.
- Skills are referenced as plain lowercase string ids (e.g. `"cast"`, `"dual-wield"`). There are no skill-name constants. Do not reintroduce a `SkillTag` type or skill constants.
- Tests seed the caches via `skills.SetTestData(...)` instead of loading from disk.
- Skill files use `SkillId + ".yaml"` for their filename (NOT `util.ConvertForFilename`, which would mangle `dual-wield`).

## Working Rules

- Preserve compatibility of existing skill identifiers and progression semantics unless the task explicitly changes gameplay balance.
- Be careful with skill-level math and thresholds. Small numeric changes can have broad balance impact.
- If a change affects a skill used by commands or combat, review the caller path as well rather than assuming the package boundary is enough.
- Avoid mixing unrelated gameplay refactors into a skill-table or progression fix.

## Verification

- Run targeted package tests for skill lookup or progression changes.
- Use at least one caller-level check if the change affects combat, usercommands, or spell gating.
- Note balance-sensitive changes explicitly if broad gameplay validation was not run.

## Documentation

- Keep only durable skill-system rules here.
