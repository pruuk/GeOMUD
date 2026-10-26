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
- `go test ./... -count=1` on Windows: 32 packages ok (2026-10-26, after Phase 1; it was 36 before, and `telemetry`, `conversations`, `quests` and `pets` took their own tests with them), 3 fail, all of which pass on Linux CI (PR #1 run,
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
server to a fixed path, runs it on a temp copy of the default world bound to 127.0.0.1 only, plays a
bot through signup, the tutorial, get/drop, say, quit and re-login, then stops the server by PID.
The repo's `_datafiles` are never written. Report its summary line. `-bots 50` is the load check;
`-addr host:port` targets a running server (the Docker one in `provisioning/compose.smoke.yml`).
Baseline numbers: `docs/phase-0-baseline.md`.

The script plays the stock tutorial, so strip phases that remove tutorial content or scripts
(Phase 3 world data, Phase 4 scripting) must replace those steps in `cmd/smoke/script.go` in the
same change.

## Reporting

State what ran and its result: package counts from the test run, and failures by name. A strip phase
is done only when its "Done when" line in `gomud_strip_plan.md` holds.
