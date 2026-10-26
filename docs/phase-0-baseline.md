# Phase 0 baseline (2026-10-26)

The numbers Phase 6 of `gomud_strip_plan.md` re-measures against. Taken on unmodified GoMud (master
`477f84f9`, Go 1.27.3) with the default world, before any strip work.

## Tests

- `go build ./...`, `go vet ./...`: pass.
- `go test ./... -count=1` on Windows: 36 packages ok, 3 fail, all Windows-only (they pass on Linux
  CI); the list is in the `geomud-verifying` skill.

## Smoke test

`go run ./cmd/smoke` (`make smoke`, CI job `Smoke Test`). Each bot creates an account and a human
character, plays the tutorial (look, walk, status, inventory, equip, fight the dummy, get the
graduation cap, drop it, get it again, portal), walks out of Town Square and back, says a line,
quits, logs back in and checks it still carries the cap.

| Run | Result | Wall time | Script time per bot (min / median / max) |
|---|---|---|---|
| Launch mode, 1 bot (Windows) | PASS 1/1 | 36.3 s | 36.3 s |
| Launch mode, 50 bots (Windows) | PASS 50/50 | 52.6 s | 36.5 / 40.5 / 52.6 s |
| Container, 50 bots | PASS 50/50 | 58.1 s | 37.7 / 41.8 / 58.1 s |
| Container, 50 bots, held 3 min | PASS 50/50 | 3 m 54 s | 33.9 / 42.0 / 54.2 s |

## Memory

Server in Docker (`provisioning/compose.smoke.yml`, image built from this commit), Docker Desktop
29.8.2 on Windows 10, 24 logical CPUs. Container memory is `docker stats` MEM USAGE; the Go figures
are the admin `server stats` "Specific Memory" table.

| State | Container | Go HeapAlloc | Go HeapSys | Go Sys | Goroutines | Connections | Rooms loaded |
|---|---|---|---|---|---|---|---|
| Just started | 27.8 MiB | | | | | 0 | |
| Idle, 26 min after start | 30.4 MiB | 26.6 MB | 38.8 MB | 49.1 MB | 10 | 1 (admin) | |
| 50 bots mid-script (tutorial), 12 s in | 82-94 MiB | 58.8 MB | 89.5 MB | 105.2 MB | 60 | 51 | 260 |
| 50 bots held in Town Square | 72.8 MiB | 41.8 MB | 93.5 MB | 109.2 MB | 60 | 51 | 68 |
| After the bots quit | 72.6 MiB | | | | | | |

The container does not shrink after the bots leave: the Go runtime keeps freed heap (HeapSys) for
reuse rather than returning it to the OS straight away.

### How to repeat it

```bash
GO_VERSION=$(awk '/^go /{print $2}' go.mod) docker compose -f provisioning/compose.smoke.yml up -d --build
# wait ~25 min, then: docker stats --no-stream geomud-smoke-server
go run ./cmd/smoke -addr 127.0.0.1:43333 -bots 50 -hold 3m
# about 70 s in (every script done, all bots held): docker stats again, and `server stats` as admin
docker compose -f provisioning/compose.smoke.yml down
```

For `server stats`, log in over telnet as `admin` / `password`; the first login forces a password
change (`password`, then the new password twice). That change lives only in the container.

## Engine issues the smoke test found

Both are upstream GoMud behaviour, left as is for now; the smoke client works around them.

- **Concurrent signups collide.** `users.CreateUser` reads the highest user ID and writes the new
  user without a lock, so simultaneous signups get the same ID (30 concurrent signups saved 7 user
  files, and bots got each other's prompts). The smoke test creates accounts one at a time.
- **A fast reply is dropped silently.** The telnet handler ignores any line arriving within
  `Timing.TurnMs` (50 ms) of the previous one, and the window also opens just after the username
  prompt is sent (`main.go` `handleTelnetConnection`, `lastInput := time.Now()` after the prompt).
  A client that answers the prompt instantly loses its answer. The smoke client spaces every line,
  the first one included, 200 ms apart.
