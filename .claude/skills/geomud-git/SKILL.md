---
name: geomud-git
description: Use for any git work in new_geo_mudbase - branching, committing, checking what upstream GoMud changed, or taking an upstream fix. Covers the never-push-upstream rule and how it is enforced.
---

# Git in new_geo_mudbase

## Upstream is read-only, forever

We never commit to, push to, or open a PR against `GoMudEngine/GoMud` (Calabe, 2026-10-24). How it
is enforced:

- `upstream` remote: fetch URL `https://github.com/GoMudEngine/GoMud.git`, push URL set to an
  invalid string, so `git push upstream` fails.
- Branches track `origin` (our repo), never `upstream`.
- `pruuk/GeOMUD` is not a GitHub fork of GoMud, so GitHub never offers upstream as a PR base. Keep it
  that way.
- Never run `gh pr create`, `gh issue create` or `gh repo fork` against GoMudEngine/GoMud. Every `gh`
  command names `--repo pruuk/GeOMUD`.

Check it before any push, and restore it if it drifted:
```bash
git remote -v                              # upstream (push) must be the NO-PUSH string
git config --get branch.master.remote      # must print origin (or nothing before the first push)
```

## Reading upstream

- `git fetch upstream` is fine at any time.
- See what changed since our base: `git log --oneline b3f55cc2..upstream/master`.
- To take one upstream fix, `git cherry-pick <sha>` onto a branch of ours, with the upstream sha in
  the commit message. Expect conflicts once the strip starts; take only fixes for code we kept.

## Our own commits

- Our repo is `origin`, `https://github.com/pruuk/GeOMUD` (public). Push our branches there.
- Commit only when Calabe asks or a skill says to. Branch off `master` for each phase of work
  (`strip/phase-1-leaf-packages` and so on).
- Named pathspecs only (`git add <paths>`); messages through a quoted heredoc.
- A commit that finishes something adds its line to `PATCH_NOTES.md` in the same commit.
