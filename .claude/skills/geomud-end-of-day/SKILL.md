---
name: geomud-end-of-day
description: Use at the end of a new_geo_mudbase working day or session, when Calabe says EOD, wrap up, button up, or signs off. Covers closing out the day's commits, rewriting status.md, checking PATCH_NOTES.md, the docs sweep, branches and worktrees, the root tidy, the disk check and box hygiene, and the memory handoff.
---

# new_geo_mudbase end of day

Run every step, in order, even when one has nothing to do. Adapted from the DOGMud and CritBench EOD
skills (2026-10-24). Git rules are in `geomud-git`.

## 1. Close out the day's work
- Finished work is committed on its branch with its `PATCH_NOTES.md` line. Work in flight is
  committed on its branch as WIP, with where it stands written in `status.md`.
- Push every branch with today's commits to `origin` (`pruuk/GeOMUD`); the remote must hold
  everything overnight. Never push to `upstream` (check `git remote -v` per `geomud-git`).
- Verification run today? Its result goes in `status.md`, read from the output (`geomud-verifying`).

## 2. Rewrite `status.md`
Short: about 80 lines at most, one line per item. Built from git and the plan, never from memory.

```
# Status (updated <YYYY-MM-DD>)

## Now
- <in progress> (<branch>)

## Next
1. <item>          (at most five, in order)

## Waiting on Calabe
- <open decision, one line, with the doc that explains it>

## Watch
- <risk that isn't a work item yet>          (at most five)
```

- **Finished items leave status today.** Before removing one, check `PATCH_NOTES.md` has its line;
  add it if missing.
- No history, no narrative, no em dashes.

## 3. Check `PATCH_NOTES.md`
Newest day on top, one line per finished item, under a `## <YYYY-MM-DD>` heading. Each line says
what changed and where (phase, package, skill), not how it was done.

## 4. Docs sweep
Live plans and design docs sit at the repo root. When a plan's work is all done (prove it from
`git log`, never from memory), `git mv` it to `docs/completed/`, then grep its basename across
`*.md`, `.claude/` and Go comments and repoint every reference. Record the hit count before the move;
afterwards the old path reads 0 and the new path reads the recorded count. Unsure whether a doc is
live: it is.

## 5. Branches and worktrees
```bash
git branch --merged master      # delete merged branches with git branch -d
git worktree list               # remove worktrees whose branch merged
```
Never remove another session's worktree or branch without asking it (ListAgents).

## 6. Root tidy
- Screenshots and loose images older than today move to `_archive/screenshots/` (gitignored). Today's
  stay: Calabe drops fresh ones in the root for Claude to read.
- Never move or delete a binary (`*.exe`): it may be the server Calabe is running. List stale ones.
- Anything untracked and unrecognised: report it with one line of context and ask.

## 7. Disk check and box hygiene
Run the global `box-hygiene` skill. Then, in PowerShell (read-only):
```powershell
"C: free {0:N1} GB" -f ((Get-PSDrive C).Free/1GB)
foreach ($p in @("$env:LOCALAPPDATA\go-build", "$env:LOCALAPPDATA\Temp\claude\C--Users-Calabe-Davis-workspace-new-geo-mudbase")) {
  if (Test-Path $p) { $s = (Get-ChildItem $p -Recurse -Force -File -ErrorAction SilentlyContinue | Measure-Object Length -Sum).Sum; "{0,8:N1} GB  {1}" -f ($s/1GB), $p }
}
```
- **Under 60 GB free: say so at the top of the report**, with the sizes.
- Go build cache over 20 GB: `go clean -cache`, unless another session is building Go right now
  (ListAgents). It only slows the next build.
- This project's session scratchpads: delete each folder not written in two days, never the current
  session's, never one a running process writes into. Never touch another project's scratchpads.
- Keep deletions and size reports in separate calls. Never kill a process to free a lock.

## 8. Memory handoff
Memory stays clean: rules belong in `CLAUDE.md` or a skill, status in `status.md`, finished work in
`PATCH_NOTES.md`. Memory holds only what fits none of those.
- Rewrite `handoff.md` in memory: this machine's state only (local branches and worktrees still
  alive, anything in `C:/tmp` to keep and why, the one thing the next session picks up first).
- `MEMORY.md` is one-line pointers: no line over 220 characters, file well under 12 KB.
- A memory that has become a rule or a status item moves there and is deleted from memory.

## Report
Short summary for Calabe: what was committed today and on which branch, what is in flight, disk
free, anything left untouched and why, and the top of Next.
