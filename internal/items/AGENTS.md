# Items Package Guide

## Scope

- Use this file for item specs, item instances, admin helpers, attack-message data, and item-file persistence in `internal/items`.
- Changes here often affect both gameplay and admin editing flows.

## Working Rules

- Preserve the separation between immutable item specifications and mutable item instances.
- Be careful with file-backed item spec operations. Admin helpers and on-disk data must stay in sync.
- Prefer existing helpers for finding, creating, saving, loading, and deleting item specs instead of open-coding file logic.
- Keep item search and matching behavior compatible unless the task explicitly changes user-facing matching semantics.
- If a change affects item events or admin item editing, consider the corresponding package or web surface together.

## Verification

- Run targeted `internal/items` tests for item-loading, matching, or admin-helper changes.
- If the change touches admin item editing, verify the exact create/update/delete path involved.
- Use broader checks if the change crosses into buffs or admin APIs.

## Documentation

- Keep this file about item-package guardrails, not a full item-system reference.
- Add local rules here only when they prevent repeated mistakes in item persistence or editing.
