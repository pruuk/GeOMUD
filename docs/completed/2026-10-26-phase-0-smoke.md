# Phase 0: smoke test and memory baseline

Date: 2026-10-26. Parent: `gomud_strip_plan.md` Phase 0. Decided with Calabe today: a Go program
`cmd/smoke` that can launch the server itself or target a host:port; the memory baseline runs the
server in a Docker container with 50 bots aimed at it, read by `docker stats`.

## Facts verified against source (2026-10-26)

| Fact | Value | Where |
|---|---|---|
| Base config | Always read from `_datafiles/config.yaml`, relative to the working directory | `internal/configs/configs.go` `ReloadConfig` |
| Override file | `CONFIG_PATH` env var, else `<DataFiles>/config-overrides.yaml`; YAML map, keys resolved by `findFullPathFrom` (partial dotted keys allowed) | `internal/configs/configs.go:390`, `:427-447` |
| World folder | `FilePaths.DataFiles: _datafiles/world/default` | `_datafiles/config.yaml:191` |
| Ports | `Network.TelnetPort: [33333, 44444]`, `LocalPort: 9999` (localhost only, no connection cap), `HttpPort: 80`, `HttpsPort: 0`, `SSHPort: 33332` (needs a host key), `AI.Port: 0`; `MaxTelnetConnections: 100` | `_datafiles/config.yaml:520-570` |
| Logs | `LOG_PATH`, `LOG_LEVEL`, `LOG_NOCOLOR` env vars | `main.go:107-109` |
| Server writes | users under `<DataFiles>/users`, room instances under `<DataFiles>/rooms.instances` | `Makefile` `clean-instances`, `users/1.yaml` |
| Admin account | `admin` / `password` in both bundled worlds | `_datafiles/world/*/users/1.yaml` |
| New account prompts | username (`new`), username-new, password-new, password-new-verify, email-new, screen-reader-new, confirm_create | `internal/inputhandlers/login.go:166-268` |
| Character start | `start` asks race then name | `internal/usercommands/start.go:51,118` |
| New players start in | tutorial rooms `[900, 901, 902, 903]`; `StartRoom: 1` (Town Square) | `_datafiles/config.yaml:616-631` |
| Logged-in signal | GMCP `Char.Info` / `Room.Info` arrive once in the world | harness `internal/session/session.go` |
| In-server memory report | admin `server` command uses `util.GetMemoryReport()` | `internal/usercommands/admin.server.go:197`, `internal/util/memory.go` |
| Docker | `compose.yml` service `server`, container `go-mud-server`, ports `33333:33333`, `80:80`, image built by `provisioning/Dockerfile` (copies all of `_datafiles`) | `compose.yml`, `provisioning/Dockerfile` |
| Host port 80 | Taken on this machine by another project's container (`critbench-frontend-1`) | `docker ps` |
| Reusable client code | `gomud-playtest-harness` `internal/telnet` (parser, ANSI strip, GMCP framing, 252 lines with tests), MIT, (c) 2026 Calabe Davis, commit `277d6aa`; its `internal/session/login.go` matches GoMud's prompts case-insensitively | `C:\Users\Calabe Davis\workspace\gomud-playtest-harness` |

## Build

1. `internal/smoke/telnet`: copy the harness's `internal/telnet` (code and tests) with a header
   comment naming the source repo and commit.
2. `cmd/smoke`, one binary, two jobs:
   - **Launch mode (default):** copy `_datafiles/world/default` to a temp dir, write an overrides
     file (DataFiles = the copy; telnet on a free port; HTTP, HTTPS, SSH, AI and LocalPort off where
     the config allows 0), start `go run .` or a built binary with `CONFIG_PATH` set, wait for the
     telnet port, run the script, stop the server (its own PID only), delete the temp dir. The
     repo's `_datafiles` must be byte-identical after a run.
   - **Target mode:** `-addr host:port` skips launching.
   - **Script (one bot):** create a fresh account (unique name per bot), finish `start`, reach a
     room with an item (leave the tutorial if needed), walk, say, get an item, drop it, get it
     again, quit; reconnect, log in, check the character is back and still holds the item. Each
     step checks the server's reply and fails with the step name and the last screen of text.
   - **`-bots N`:** run N scripts concurrently; with `-hold D` the bots stay logged in for D after
     the script so memory can be read under load. Exit 0 only if every bot passed.
3. Docker for the memory baseline: a compose override or separate compose file that runs the
   server container under a GeOMUD name, publishes only the telnet port on a non-default host port,
   and does not publish 80. Do not change what `compose.yml` already does for `make run-docker`.
4. `make smoke` (launch mode) and a CI job in `.github/workflows/ci.yml` that runs it on PRs touching
   Go code.

## Done when

- `go run ./cmd/smoke` passes locally, and `go run ./cmd/smoke -bots 50` passes.
- `go build ./...`, `go vet ./...`, `go test ./... -count=1` show only the 3 known Windows failures.
- Memory baseline (idle, then 50 bots held) recorded in the repo: container memory from
  `docker stats`, plus the admin `server` memory report.
- `geomud-verifying` smoke section updated to name the command.
