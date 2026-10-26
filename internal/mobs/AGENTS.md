# Mobs Package Guide

## Scope

- Use this file for mob lifecycle, mob instances, persistence/loading behavior, and mob-side interactions in `internal/mobs`.
- Mob changes often affect combat, rooms, and AI behavior.

## Working Rules

- Preserve the distinction between mob templates/specs and live mob instances.
- Be careful with spawn/despawn, room placement, and instance lookup changes. Those assumptions are reused across many systems.
- If a change affects charmed mobs, inspect the connected package behavior rather than patching only one call site.
- Avoid folding AI policy into this package when it belongs in mob commands, hooks, or modules.

## Verification

- Run targeted mob-package tests when changing lifecycle or lookup behavior.
- Use a higher-level check for changes that affect combat or room movement.
- Call out untested instance-lifecycle or persistence edges when behavior changed.

## Documentation

- Keep only durable mob-lifecycle rules here.
