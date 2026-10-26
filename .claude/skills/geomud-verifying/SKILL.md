---
name: geomud-verifying
description: Use before claiming any change in new_geo_mudbase works or any strip phase is done. Covers the build, the test suite, the known baseline failure, and how to report results.
---

# Verifying a change

Run from the repo root. Show the output; never claim a pass you did not run.

```bash
go build ./...
go test ./... 2>&1 | grep -vE '^ok|no test files'   # prints only failures
git diff --name-only -- '*.go' | xargs -r gofmt -l  # must print nothing
```

## Baseline (upstream `b3f55cc2`, 2026-10-24)

- `go build ./...` passes.
- `go test ./... -count=1` on Windows: 32 packages ok (2026-10-26, after Phase 4's scripting removal; it was 33 before, and `internal/scripting` took its tests with it. Phase 1 took it from 36 by removing `telemetry`, `conversations`, `quests` and `pets`), 3 fail, all of which pass on Linux CI (PR #1 run,
  2026-10-24). Any other failure is new and is ours.
  - `scripts` `TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride` (`https_setup_test.go:311`): cause not investigated
  - `internal/mobs` `TestLoadPluginMobs_MergesValidMob`: "filepath mismatch", `/` vs `\` in the expected path suffix
  - `cmd/reset-admin-pw` `TestPromptForPasswordMismatch`: temp dir cleanup fails, prompt input file still held open
- Always pass `-count=1`. Without it Go can replay cached results; the first baseline run here
  showed only 1 failure that way.
- The full pre-PR gate is in `.github/CONTRIBUTING.md` section 3.

When the baseline changes (the failure is fixed, or a package is stripped), update this section in
the same commit.

## Smoke test

`go run ./cmd/smoke` from the repo root (`make smoke`; CI runs it as `Smoke Test`). It builds the
server to a fixed path, runs it on a temp copy of the empty world bound to 127.0.0.1 only, then
stops the server by PID. The repo's `_datafiles` are never written. Flow:

1. An admin logs in (`-admin-user`/`-admin-pass`, default the stock `admin`/`password`), answers
   GoMud's forced first-login password change (sets it to the same text; if the account was
   already changed it sees the "current password" question and cancels), goes to the start room
   and runs `item spawn copper ring` once per bot, checking each lands (`look`). Then it quits.
2. Each bot signs up, runs `start` (race, name; no tutorial question, the world has no tutorial
   rooms), lands in Town Square, then get, drop, get the ring (checking inventory and the floor
   each time), walks north and back, says something, quits, logs back in, checks it still holds
   the ring, dies (`suicide`), checks it is in the recovery room ("Waiting room"), walks out
   through `portal` to Town Square and checks it still holds the ring, holds (`-hold`), and quits.
   A drop the floor-item limit refuses is retried. `quit` is silent until the server hangs up (its
   meditation messages were a JS script, gone since Phase 4); the step waits for the disconnect.
   The death steps exercise the Phase 4 death stand-in (revive at 1 health, fixed exit from
   room 75), recorded in `docs/completed/2026-10-26-phase-4-scripting-surface.md`.

Each bot's client also records the GMCP package names it receives (it lists Char, Room in
`Core.Supports.Set`; the server sends `Comm.Channel` without being asked) and the bot fails unless
it got at least one `Char.*`, `Room.*` and `Comm.*` package (the Mudlet check). The summary prints
every package name with its message count.

Report its summary line. `-bots 50` is the load check; `-addr host:port` targets a running server
(the Docker one in `provisioning/compose.smoke.yml`); `-server-log file` saves the launched
server's log. Baseline numbers: `docs/phase-0-baseline.md`.

Strip phases that remove items, the start room or the admin account must keep these steps in
`cmd/smoke` working in the same change.

## Reporting

State what ran and its result: package counts from the test run, and failures by name. A strip phase
is done only when its "Done when" line in `gomud_strip_plan.md` holds.
