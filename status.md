# Status (updated 2026-10-27)

## Now
- Phase 4b (NPCs are players), spec `2026-10-27-phase-4b-npcs-are-players-design.md`. Slices 1
  (actors, PR #16) and 2 (one ID space, PR #18) merged. Slice 2b (NPC and companion persistence,
  decisions 12-15) plan `2026-10-27-phase-4b-2b-npc-persistence-plan.md` with its inventory, in
  Calabe's review
- Strip Phases 0-4 and GeOMUD player branding (PR #14) are merged (`gomud_strip_plan.md`)

## Next
1. Phase 4b slice 2b: execute the plan once approved; then slice 2c (item IDs) or 3 (messaging)
2. Phase 5: rules seam. Takes over the death stand-in (revive at 1 health, portal out of room 75;
   `docs/completed/2026-10-26-phase-4-scripting-surface.md`), `suicide`, the combat commands, the
   `professions`/`skills`/`spells` world folders, `set shopstyle`, `help spell` and the loot goblin
   config. Spells and scripted buffs do nothing until then
3. Phase 6: measure and tag. Also measure the Phase 0 engine (`2737986c`) on the trimmed world, since
   the Phase 0 baseline ran on the deleted `default` world. Docs sweep: recheck the "early in
   development" line in `help about` (Calabe kept it 2026-10-27, knowing it goes stale)
4. Security pass before any public server: upstream CodeQL debt (unsalted SHA-256 passwords at
   `internal/util/util.go:384`, user-provided file paths at `util.go:706-752`, integer narrowing)

## Waiting on Calabe
- Review of the Phase 4b slice 2b plan

## Watch
- Phase 4b slice 2 permadeath reroll (`internal/usercommands/suicide.go`): body re-keying and event
  ordering are unverified live, because permadeath is off in the `empty` world. Check it with
  permadeath on (and `LivesStart: 0`) before relying on it. A reroll also orphans the old body's
  charmed mobs until the charm expires (charm has no live callers today)
- `_datafiles/html/admin/docs-modules.html` code examples at about lines 462, 621, 910, 1399 and
  1492 still use upstream's `*users.UserRecord` (now `*accounts.Account`)
- A web-created account (`apiV1CreateUser`, connection 0) stays online with an actor, and that user
  then cannot log in by telnet ("already logged in"). Upstream behaviour
- For the Phase 4b slice 2 plan (from the slice 1 design review, 2026-10-27):
  - Nothing outside `actors` can reach the mob behind an NPC actor without a type assertion (which
    the guard forbids). Put the body's identity on the `Actor` (for example `Actor.Mob`, nil for
    players) and key `byMobInstanceId` on it, not on the controller
  - Possession breaks controller-keyed indexes: an account must remember its own body's actor, and
    logout returns control to it instead of removing it. A single `MoveController(from, to)` is
    cleaner than two `SetController` calls
  - `Actor.Character` is read without the registry lock while a copyover rebind can write it from a
    login goroutine: lock it behind a getter, or move account hooks onto the world loop
  - `accounts.SetTestUser` skips `OnActivate`, so tests that use it get no actor
  - Web-created accounts (`CreateUser` with connection 0) count as online forever and keep an actor
    until restart (upstream behaviour)
- Login goroutines (`LoginUser`, `CopyoverReconnectUser`) change accounts without the MUD lock the
  world loop holds (`util.LockMud`). Web admin edits are safe: their routes run under
  `RunWithMUDLocked` (`internal/web/api_routes.go`, `web.go`), except internal requests. Since 4b
  slice 1 an edit copies into the live character instead of swapping it
- Phase 4b slices 5 and 7 text parity against `cmd/smoke/testdata/pre-4b/bot0.txt`: mask the run
  tag (`smk<tag>u0`), game-clock times and HP/MP prompts, and wandering-NPC lines ("guard enters
  from the north"); strip `\r` or pin `cmd/smoke/testdata/** text eol=lf` in `.gitattributes`
  (only `*.go`, `go.mod`, `go.sum` are pinned; a fresh Windows checkout gets CRLF)
- Login splash globe (`templates/login/connect-splash.template`): Calabe called it weak, fine for
  now (2026-10-27). Redo it when the splash is next touched
- Phase 5 seam: `buffs` has 11 importers; `characters` imports six rules packages (buffs, races,
  skills, spells, statmods, stats)
- 3 Go tests fail on Windows only (`geomud-verifying` lists them); Linux CI is the reference
- Concurrent signups colliding on user ID is likely fixed by Phase 4b slice 1 (`CreateUser` now
  holds `lifecycleMu` across `GetUniqueUserId` and `AddUser`). Unverified: prove it with a 50-bot
  smoke run without `signupMu` (`cmd/smoke/script.go`) before removing that workaround
- Upstream engine bugs, not fixed (`docs/phase-0-baseline.md`):
  a line sent within 50 ms of the previous one, or of the login prompt, is silently dropped
- Stale permission keys (`quests.write`, `pets.write` and others) in saved mod accounts: they load,
  but a full permission-list save that carries one is rejected with a 400
- `empty` world loose ends: the `wildfire` mutator names missing buff 22; item 30001 names missing
  buff 5; the `exits` command file was never registered (upstream too)
