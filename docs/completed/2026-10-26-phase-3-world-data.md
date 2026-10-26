# Phase 3: trim world data

Date: 2026-10-26. Parent: `gomud_strip_plan.md` Phase 3 ("Delete `world/default`; point config at a
trimmed `empty` world; strip datafile folders whose systems are gone (pets, quests, professions,
conversations, audio). Done when: server boots on the trimmed world with no missing-file
warnings"), and section 1: "`_datafiles/world/empty` stays as the base world and gets trimmed to one
zone with a handful of rooms, which becomes the test fixture". Branch `strip/phase-3-world-data`.

## Facts verified against source (2026-10-26, master `d1302c71`)

| Fact | Value | Where |
|---|---|---|
| World folder setting | `FilePaths.DataFiles: _datafiles/world/default`; code default the same | `_datafiles/config.yaml:191`, `internal/configs/config.filepaths.go:26` (+ its test `:10`, `configs_bench_test.go:22`) |
| World validation | `util.ValidateWorldFiles("_datafiles/world/default", <configured world>)` exits if the configured world lacks any top-level folder that `default` has | `main.go:189`, `internal/util/util.go:883` |
| Other `world/default` paths | `Makefile:160` (clean-instances), `_datafiles/config.yaml:495` (translation `LanguagePaths`), `scripts/https-setup.sh:106` and `scripts/https_setup_test.go:13,39,181`, admin `config-wizard.html:980` placeholder, `cmd/smoke` launch mode copies `_datafiles/world/default` | grep |
| Docs naming `world/default` | admin `scripting-*.html` examples (go with scripting in Phase 4), `internal/skills` README/AGENTS (Phase 5), `_datafiles/guides/*` | grep |
| `empty` world folders (files) | biomes 19, buffs 6, combat-messages 8, conversations 2, items 14, localize 2, mobs 7, mutators 4, panel-layouts 3, pets 1, plugin-data 1, professions 10, quests 2, races 5, rooms 40, skills 16, spells 16, templates 191, users 1 | `find` |
| `empty` zones | `startland` (rooms 1-3, room 1 "Town Square" is `StartRoom`), `tutorial` (900-903, 4 JS), `shadow_realm` (death recovery room 75, 2 JS), `endless_trashheap` (22 rooms, room 139, loot goblin mob 13) | `rooms/*/zone-config.yaml` |
| Tutorial in `empty` | Same rooms and flow as `default`; each room's JS has about 25 more lines, YAML differs in 2 lines each. Graduation cap item 20043 and mobs 57 (orb), 58 (training dummy) exist in `empty` | `git diff --no-index` |
| Special rooms | `StartRoom: 1`, `DeathRecoveryRoom: 75`, `TutorialRooms: [900, 901, 902, 903]` | `_datafiles/config.yaml:616-631` |

## Decisions in this plan

- **The tutorial stays until Phase 4.** It is all room scripts, and Phase 4 deletes every script
  datafile, so the smoke test's per-player item source has to be replaced there anyway. Phase 3 keeps
  the `empty` world's tutorial and moves the smoke test onto it.
- **Zones kept:** `startland` (start room), `tutorial` (until Phase 4), `shadow_realm` (death
  recovery room, while death stays until Phase 5). `endless_trashheap` goes (its loot goblin is gone).
- **Folder validation** checks the folders the engine still loads, from a list in Go, not a
  reference world on disk.
- **Phase 6 note:** the Phase 0 memory baseline was taken on `default` (1,574 files). Phase 6 should
  also measure the Phase 0 engine (`2737986c`) on the trimmed world, so the before/after compares
  engines, not world sizes.

## Tasks (in order)

1. **Engine onto `empty`** (Sonnet): replace the reference-world validation with a required-folder
   list; point config, code defaults, tests, Makefile, https-setup script and test, translation
   paths, the config-wizard placeholder and `cmd/smoke` at `_datafiles/world/empty`; adapt the smoke
   script to the `empty` tutorial and start zone; delete `_datafiles/world/default`.
2. **Trim `empty`** (Sonnet): delete folders and files for systems that are gone (`pets`, `quests`,
   `conversations`, `professions` if nothing loads it, `audio.yaml`, `musicfile` keys), the
   `endless_trashheap` zone and its mobs, help templates and keyword aliases for removed commands
   and systems, panel layouts for removed panels, ANSI aliases for removed systems, and mob idle or
   angry commands that name a removed command. Remove upstream's root working notes
   (`election-tax-plan.md`, `notes.md`, `feature-screenshots/`).

## Rules for both tasks

- Every removal must leave the server booting without missing-file or load warnings: read the server
  log of a smoke run (launch mode keeps it in its temp dir; `-hold` keeps the server up while you
  read it) and account for every WARN and ERROR line.
- Edit files only with Edit and Write; `git rm` for deletions. No public listeners; kill only your
  own PIDs. Throwaway files in the scratchpad.
- Done per task: `go build ./...`, `go vet ./...`, `gofmt -l` on touched files, `go test ./... -count=1`
  `ok` list diffed by name against the previous run, `go run ./cmd/smoke -bots 5` passes.

## Done when

The Phase 3 "Done when" holds (server log on the trimmed world has no missing-file warnings), `go run
./cmd/smoke -bots 50` passes, and the admin pages load in the smoke container (controller).
