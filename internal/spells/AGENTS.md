# Spells Package Guide

## Scope

- Use this file for spell metadata, spell lookup, and spell-side persistence/helpers in `internal/spells`.
- Spell behavior usually crosses into combat, buffs, and commands. Spell effects were JS scripts until Phase 4; they are dead until Phase 5 rebuilds them in Go.

## Working Rules

- Preserve spell identifiers and lookup behavior unless the task explicitly changes spell contracts.
- Be careful with changes that affect castability, targeting, or cooldown metadata.
- Keep spell metadata changes separate from unrelated combat or command refactors when possible.

## Verification

- Run targeted package tests for spell lookup or persistence changes.
- Use a higher-level validation path if the change affects casting flow or combat.
- Call out compatibility risk when changing how existing spells are loaded or resolved.

## Documentation

- Keep this file about spell contract and integration guardrails.
