# Known issues

Engine and world issues found and not fixed, most inherited from upstream GoMud. One line each, with
where. Phase 4b notes live in `2026-10-27-phase-4b-npcs-are-players-design.md` ("Notes for later
slices"). Remove a line when its fix lands (with a `PATCH_NOTES.md` line).

## Engine

- A line sent within 50 ms of the previous one, or of the login prompt, is silently dropped
  (`docs/phase-0-baseline.md`; the smoke test spaces commands to avoid it).
- Concurrent signups colliding on user ID is likely fixed since Phase 4b slice 1 (`CreateUser` holds
  `lifecycleMu` across `GetUniqueUserId` and `AddUser`). Unverified: a 50-bot smoke run without
  `signupMu` (`cmd/smoke/script.go`) would prove it before that workaround goes.
- A web-created account (`apiV1CreateUser`, connection 0) stays online with an actor, and that user
  then cannot log in by telnet ("already logged in").
- With `CorpseItems: true` (off by default), looting a corpse duplicates items under one ID: a
  player corpse keeps a copy of the character still wearing its items (`usercommands/suicide.go:265,280`,
  `get.go:134-148,185-216`), and an NPC corpse lists dropped worn items without taking them off the
  body (`mobcommands/suicide.go:349-366,407`). Phase 5's death rule owns corpses.
- Admin room template saves can lose live items: `BuildRoom` and `ConnectRoom` put a plain template
  load in memory (`roommanager.go:862-865, 925-926`); `SaveRoomTemplate` on a room not in memory
  ignores its instance file; a template container with items replaces the live container's contents.
  Live floor items also reach template files (without IDs) through `mapper.go:1397` and `BuildRoom`,
  so later template loads copy them.
- The web room PATCH (`web/api_v1_rooms.go:238-262`) and the raw room instance PUT accept item IDs
  as sent, so a hand-written API call can give two items one ID. The rooms page never sends them.
- Stale permission keys (`quests.write`, `pets.write` and others) in saved mod accounts load, but a
  full permission-list save that carries one is rejected with a 400.

## Docs and world

- `_datafiles/html/admin/docs-modules.html` code examples at about lines 462, 621, 910, 1399 and 1492
  still use upstream's `*users.UserRecord` (now `*accounts.Account`).
- `empty` world: the `wildfire` mutator names missing buff 22; item 30001 names missing buff 5; the
  `exits` command file was never registered (upstream too).
- Login splash globe (`templates/login/connect-splash.template`): Calabe called it weak, fine for now
  (2026-10-27); redo it when the splash is next touched.
