# Patch Notes

Newest day first. One line per finished item.

## 2026-10-27
- `help about` rewritten for GeOMUD: what it is, its status, credit to GoMud (Dylan Squires and contributors, MIT) and both repo links; GoMud's history, Discord advert and thanks list removed. `help geomud` added as an alias (`help gomud` kept); the `help` index points to `help about`

## 2026-10-26
- Go 1.26.0 to 1.27.3 (`go.mod`); no dependency versions changed, Windows baseline unchanged (same 3 failures)
- Dependabot off (`.github/dependabot.yml` deleted); Go toolchain workflow is manual-only (schedule removed)
- Phase 0: telnet smoke test `cmd/smoke` (`make smoke`, CI job), Docker server for load runs (`provisioning/compose.smoke.yml`), baseline in `docs/phase-0-baseline.md` (idle 30.4 MiB, 50 bots held 72.8 MiB); upstream root `AGENTS.md` deleted
- Phase 1: deleted leaf packages `clans`, `integrations/discord`, `migration`, `suggestions` (tab completion), `audio` (MSP), `telemetry`, `conversations`, `quests`, `pets`, with their commands, hooks, events, scripting functions, GMCP pieces, config and admin pages (about 21,000 lines); world data for them stays until Phase 3
- Ruling: NPCs are players under the hood; new Phase 4b before the rules seam (`gomud_strip_plan.md` decision 4)
- Phase 2: deleted game commands (economy, character info, every skill and spell command, stripped admin commands), 18 mob commands and 5 game hooks; `suicide` (the death routine) and the combat commands stay until Phase 5; mobs no longer start or join fights until Phase 5
- Phase 3: server runs on the trimmed `empty` world (376 to 252 files: zones startland, tutorial, shadow_realm); `world/default` deleted (1,574 files); world validation checks the folders the engine loads; restored buff 32 that `empty` lacked (the tutorial fight ran without its bonus); upstream root notes removed
- Phase 4: JS/Lua scripting deleted (`goja`, `gopher-lua` out of `go.mod`; about 72,500 lines); tutorial zone deleted; smoke test now has an admin spawn items, runs a death and walk-out, and asserts GMCP Char, Room and Comm; death stand-in until Phase 5 (revive at 1 health, portal out of the recovery room); Mudlet Discord Rich Presence removed

## 2026-10-24
- Cloned upstream GoMud (`b3f55cc2`, MIT) into this repo; `upstream` remote made fetch-only
- Wrote `gomud_strip_plan.md`: what we keep from GoMud and the strip roadmap
- Replaced GoMud's README with a GeOMUD README (origins, docs, building); repo published at `pruuk/GeOMUD`
- Added project `CLAUDE.md`, `status.md`, this file, and skills `geomud-git`, `geomud-verifying`, `geomud-end-of-day`
- Contributing rules adapted from DOGMud (`.github/CONTRIBUTING.md`, PR template, CODEOWNERS to `@pruuk`)
- Removed GoMud's release, Docker publish and Discord workflows; CI now runs Go tests on master pushes; dependabot covers Go modules and Actions only
- `.gitattributes` keeps Go files LF so gofmt checks pass on Windows; Windows test baseline corrected to 3 known failures
- First EOD run of `geomud-end-of-day`; the skill gained the EOD PR step, local-date rule, remote branch cleanup and a scratchpad-age fix
- Decisions: rules move into a `diku-basic` reference ruleset, then shrink; JS and Lua scripting is dropped
