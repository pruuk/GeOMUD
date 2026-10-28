# GeOMUD

A Diku-style MUD engine, in development, for worlds built on real geography. The goal: rooms generated from GIS data on a uniform grid, worlds from one small island up to a whole planet, hand-built areas that override generated ones, and only the rooms players are in kept loaded.

**Status:** early. The code is currently upstream GoMud, about to be stripped down to a bare engine. Nothing geographic is built yet. See [`status.md`](status.md).

## Origins

GeOMUD is a hard fork of [GoMud](https://github.com/GoMudEngine/GoMud) by Dylan Squires and contributors, taken at commit `b3f55cc2` (2026-10-05) under the MIT license. GoMud's history is kept in this repo. This is an independent project: changes here are not sent back to GoMud, so please don't report GeOMUD issues there.

## Documents

| File | What it covers |
|---|---|
| [`mud_base_requirements.md`](mud_base_requirements.md) | Requirements |
| [`mud_architecture.md`](mud_architecture.md) | The original architecture concept |
| [`engine_proposal.md`](engine_proposal.md) | Why GoMud was chosen as the base |
| [`world_model_notes.md`](world_model_notes.md) | World model: grid, dense layers, explore, encounters |
| [`gomud_strip_plan.md`](gomud_strip_plan.md) | What we keep from GoMud, and the strip-down roadmap |
| [`2026-10-27-phase-4b-npcs-are-players-design.md`](2026-10-27-phase-4b-npcs-are-players-design.md) | Phase 4b: one actor model, IDs, NPC persistence, messaging; its slices and rulings |
| [`docs/known-issues.md`](docs/known-issues.md) | Engine and world issues found and not yet fixed |
| [`status.md`](status.md) | What is in progress and what is next |
| [`PATCH_NOTES.md`](PATCH_NOTES.md) | What has been completed |

## Building

Requires Go 1.27.3 or later (pinned in `go.mod`).

```
go build ./...
go test ./...
go run .
```

The Makefile targets (`make build`, `make run`, `make test`) come from GoMud; `make test` also runs code generation and the JS linter. With the stock config, the server listens for telnet on ports 33333 and 44444 and serves the web client and admin over HTTP on port 80.

## License

MIT. See [`LICENSE.md`](LICENSE.md).
