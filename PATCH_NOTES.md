# Patch Notes

Newest day first. One line per finished item.

## 2026-10-24
- Cloned upstream GoMud (`b3f55cc2`, MIT) into this repo; `upstream` remote made fetch-only
- Wrote `gomud_strip_plan.md`: what we keep from GoMud and the strip roadmap
- Replaced GoMud's README with a GeOMUD README (origins, docs, building); repo published at `pruuk/GeOMUD`
- Added project `CLAUDE.md`, `status.md`, this file, and skills `geomud-git`, `geomud-verifying`, `geomud-end-of-day`
- Contributing rules adapted from DOGMud (`.github/CONTRIBUTING.md`, PR template, CODEOWNERS to `@pruuk`)
- `.gitattributes` keeps Go files LF so gofmt checks pass on Windows; Windows test baseline corrected to 3 known failures
- Decisions: rules move into a `diku-basic` reference ruleset, then shrink; JS and Lua scripting is dropped
