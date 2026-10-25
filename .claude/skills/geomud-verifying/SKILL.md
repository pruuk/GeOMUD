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
- `go test ./...`: 34 packages pass, 1 fails: `scripts`,
  `TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride` (`https_setup_test.go:311`). Cause
  not yet investigated. Any other failure is new and is ours.

When the baseline changes (the failure is fixed, or a package is stripped), update this section in
the same commit.

## Smoke test

Not built yet: it is Phase 0 work in `gomud_strip_plan.md` (telnet login, walk, say, get and drop an
item, quit, log back in). Until it exists, say "no smoke test run" in any report.

## Reporting

State what ran and its result: package counts from the test run, and failures by name. A strip phase
is done only when its "Done when" line in `gomud_strip_plan.md` holds.
