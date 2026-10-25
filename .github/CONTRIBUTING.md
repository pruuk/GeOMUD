# Contributing to GeOMUD

Thank you for building on GeOMUD. This page covers how work gets from your branch or fork into
`pruuk/GeOMUD`, and the engine rules every change has to keep.

GeOMUD is a hard fork of [GoMud](https://github.com/GoMudEngine/GoMud). Open pull requests against
`pruuk/GeOMUD`, never against `GoMudEngine/GoMud`, and don't report GeOMUD issues upstream. Pass
`--repo pruuk/GeOMUD` on every `gh pr` and `gh issue` command.

## The short version

1. Talk about a new system before you build it (a one-page spec).
2. One system per pull request.
3. The pull request is green before you open it.
4. The engine stays rules-agnostic, and behaviour is Go.
5. Anything that calls an outside service ships switched off.
6. Run your own review first and attach the notes.

## 1. Spec before code for a new system

A new system is anything that adds a package, a player command family, a persistent store, a new
kind of room, grid or generator behaviour, or a call to an outside service. Before writing it, open
a GitHub issue or a draft PR holding a one-page spec that covers:

- What the player or builder sees and does, in a few sentences.
- Whether it is engine or ruleset, and why.
- What existing systems it touches, and which existing mechanisms it reuses. Search for how the
  codebase already does the job before inventing a new way.
- What it persists, and where (authored YAML, the database, or memory only).
- Whether it changes room identity or generated output. Either one can orphan every override and
  saved location keyed to a room, so say how existing keys survive.
- Any outside service it calls, and what player data it would send.

The owner answers with yes, no, or changes. That is much cheaper than reviewing thousands of lines
built in a direction the engine is not taking. Bug fixes, content inside an existing system, and
small improvements do not need a spec.

## 2. One system per pull request

- A pull request carries one system or one coherent change. Stack dependent work as separate pull
  requests and say which comes first.
- Aim for a diff a reviewer can hold in their head: as a guide, under roughly 5,000 lines of added
  or changed Go. Deletions from the GoMud strip-down (`gomud_strip_plan.md`) do not count toward
  that, but keep each strip pull request to one phase or one package group. Generated content does
  not count either; say how it was generated and include the generator.
- Rebase onto current `master` before opening, and keep the branch rebased while it is in review.
- A system that needs a hook in shared engine code (`internal/rooms`, `internal/events`,
  `internal/usercommands`, `world.go`, `main.go`) lands that hook in its own small pull request
  first, or calls it out at the top of the description.

## 3. Green before you open

Run the full local gate from the repo root and paste the result into the pull request description:

```
gofmt -l $(git ls-files '*.go')               # must print nothing
go vet -composites=false ./...                # same flags as make validate
go generate ./... && git diff --exit-code     # generated files are committed and current
go build ./...
go test ./... -count=1
```

Then boot the server once (`go run .`) and confirm it reaches the login prompt with no panic. CI
(`.github/workflows/ci.yml`) runs the format, vet, generate and test checks on Linux when Go files
change; a pull request with red checks is not ready for review.

Known exceptions: on Windows, three tests fail on the unmodified base and pass on Linux CI:
`scripts` `TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride`, `internal/mobs`
`TestLoadPluginMobs_MergesValidMob` and `cmd/reset-admin-pw` `TestPromptForPasswordMismatch`. Any
other failure is yours. `.gitattributes` keeps Go files LF so the gofmt check works on Windows.

Guard tests that enforce project rules (such as "no ruleset stat name in a core package") name the
rule and the fix when they fail. Fix the code or register the new site as the message says; never
delete or loosen the guard.

## 4. Engine rules

- **Rules-agnostic core.** Stat names, races, classes, skills, spells, combat math and XP curves
  belong to a ruleset, never to engine packages. The engine asks the ruleset what happens.
- **Behaviour is Go.** No JavaScript or Lua scripting. Hand-built content that needs behaviour
  references a named Go behaviour from its YAML.
- **Stable identity.** Room keys come from geography, never from generation order. Generated content
  is a function of room key, seed and time; a change to what a generator outputs is a versioned
  change and says so in its pull request.
- **Light footprint.** One binary, no required external services. A new dependency says what it
  costs in memory and binary size.

## 5. Outside services ship switched off

Every feature that calls an outside service, or sends any player text or data anywhere:

- Is off by default, in both the Go default and the shipped `_datafiles/config.yaml`. The server
  operator turns it on.
- Asks players to opt in. Any consent box starts unticked, and consent given for one feature never
  carries over to a new one.
- Treats what comes back as untrusted input. It never reaches a command, a file path or a template
  without validation.
- Never turns text one player can influence into world text other players see unless it passes the
  same checks and is attributable.
- Bounds its spend, and the bound survives a restart.

## 6. Review your own work first

Before opening, review the diff adversarially, by hand or with your own tools: rules leaking into
core packages, item or currency duplication, permission bypasses, input that reaches commands or
files, a room key or generator output that changed silently, places where you rebuilt something the
engine already has, and hardcoded tuning numbers. Put the notes, and what you fixed, in the
description. Review then checks yours rather than starting from nothing.

## Writing the pull request

The template asks for a description, the changes, the gate result and a checklist. Also include:

- Which spec or issue it implements.
- New config keys with their shipped values.
- New commands players or builders can type.
- Its line for `PATCH_NOTES.md` (added in the pull request), or "none (internal)".
- Anything you are unsure of. Saying so is welcome.

## If you work with Claude Code

This repo ships project instructions in `CLAUDE.md` and procedures in `.claude/skills/`
(`geomud-git`, `geomud-verifying`, `geomud-end-of-day`). They load automatically in a checkout of
this repo and are the same rules review applies.

## Reviews

Review comments say whether they block: a comment prefixed `SUGGESTION:` or `CONSIDER:` does not.
Do not review draft pull requests. Reviews are direct and respectful in both directions.

## Getting help

Ask in the pull request or an issue, or reach the owner directly.
