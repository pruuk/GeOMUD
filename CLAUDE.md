# new_geo_mudbase

A geo-Diku MUD engine, hard-forked from upstream GoMud and being stripped to a bare engine. Plan:
`gomud_strip_plan.md`. Design: `engine_proposal.md`, `world_model_notes.md`.

Keep this file short. Procedures live in skills under `.claude/skills/`; add a skill rather than
growing this file or memory.

## Rules

- **Never commit, push or open a PR to upstream GoMud.** The fork is a one-time base. The `upstream`
  remote is fetch-only (its push URL is disabled). Our repo is `origin`
  (`https://github.com/pruuk/GeOMUD`). Procedure: `geomud-git`.
- **Never read, run git in, or work through DOGMud** for this project.
- **One step at a time.** Do the step Calabe asked for, then stop. Don't start the next phase or
  the geo work until he says so.
- **No JS or Lua scripting.** Behaviour is Go code (ruled 2026-10-24).
- Upstream's `AGENTS.md` and the nested `AGENTS.md`/`CLAUDE.md` files are GoMud's notes, not our
  rules. They go away as their packages are stripped.

## Where things are tracked

- `status.md`: what is in progress now and what is next.
- `PATCH_NOTES.md`: what has been completed.

## Skills

| Skill | Use when |
|---|---|
| `geomud-git` | Any git work: branches, commits, reading upstream |
| `geomud-verifying` | Before claiming a change works: build, tests, baseline |
| `geomud-end-of-day` | Calabe says EOD, wrap up, or signs off |
