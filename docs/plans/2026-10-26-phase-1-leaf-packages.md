# Phase 1: delete leaf packages

Date: 2026-10-26. Parent: `gomud_strip_plan.md` Phase 1 ("Delete in order of fewest importers ...
with their hooks, commands, admin pages, API routes and GMCP pieces. Done when: no references
remain; smoke test passes"). Branch `strip/phase-1-leaf-packages`, one commit per task.

## Facts verified against source (2026-10-26, master `2737986c`)

Importers are non-test Go files outside the package (`grep -rl 'internal/<pkg>"'`).

| Package | Size (non-test) | Importing files |
|---|---|---|
| `clans` | 1 file, 40 lines | none |
| `integrations/discord` | 3 files, 580 lines | `main.go`; config `Integrations.Discord` (`internal/configs/config.integrations.go:7`, `_datafiles/config.yaml:461`) |
| `migration` | 3 files, 366 lines | `main.go` |
| `suggestions` | 2 files, 609 lines | `main.go`, `world.go`, `modules/gmcp/gmcp.Suggestion.go` |
| `audio` | 1 file, 139 lines | `main.go`, `internal/rooms/rooms.go`, `internal/users/userrecord.go`, `internal/web/api_v1_audio.go` |
| `telemetry` | 4 files, 628 lines | `main.go`, `internal/hooks/NewTurn_AutoSave.go`, `internal/inputhandlers/login.go`, `internal/usercommands/admin.telemetry.go`, `internal/usercommands/help.go`, `internal/web/admin_telemetry.go` |
| `conversations` | 5 files, 475 lines | `main.go`, `internal/hooks/MobIdle_HandleIdleMobs.go`, `internal/mobcommands/converse.go`, `internal/mobs/mobs.go`, `internal/web/api_v1_conversations.go`, `internal/web/api_v1_yaml.go` |
| `quests` | 4 files, 360 lines | `main.go`, `internal/characters/character.go`, `internal/hooks/Quest_HandleQuestUpdate.go`, `internal/usercommands/{admin.item,admin.questtoken,quests}.go`, `internal/web/{api_v1_quests,api_v1_yaml}.go`, `modules/gmcp/gmcp.Char.go` |
| `pets` | 7 files, 802 lines | `main.go`, `internal/characters/character.go`, `internal/combat/pet_rank.go`, `internal/hooks/PetLevelChange_SendPetLevelNotifications.go`, `internal/scripting/{pet,pet_func}.go`, `internal/usercommands/{buy,list}.go`, `internal/web/{api_v1_pets,api_v1_yaml}.go` |

Related files that do not import the package but belong to it:

- Commands: `usercommands/{quests,quests.panels,pet}.go`, `mobcommands/{converse,givequest}.go`.
- Hooks: `DayNightCycle_PetDailyTick.go`, `ItemOwnership_CheckItemQuests.go`, `RoomChange_LocationMusicChange.go`.
- Web: `internal/web/admin_audio.go`; admin pages `_datafiles/html/admin/{audio,audio-api,pets,pets-api,pets-ranks,quests,quests-api,scripting-pets,telemetry,telemetry-api}.html`.
- Events: `Quest` (`internal/events/eventtypes.go:54`), `MSP` (`:144`), `PetLevelChange` (`:279`), `PetFed` (`:291`), `PetItemChange` (`:298`).
- Scripting functions (scripting itself goes in Phase 4): `actor_func.go` `HasQuest`, `GiveQuest`, `GetPet`, `PlaySound`, `PlayMusic`, `IsQuestDone`, `ClearQuestToken`; `item_func.go` `GetQuestToken`; `party_func.go` `GiveQuest`; `pet.go`, `pet_func.go`.
- The tutorial (rooms 900-903, which the smoke test plays) references none of these systems; `901.js` only says the word "quests" in dialogue.

## Rules for every task

- Follow the compiler: delete the package, then fix each error by the buckets in
  `gomud_strip_plan.md` section 1. A caller in kept code loses only the stripped call; a caller
  that exists only for the stripped system goes with it.
- Remove with the package: its commands (and their registrations and help), hooks (and their
  registrations), event types nothing else uses, web routes and admin pages (and nav links), GMCP
  pieces, config struct fields and their `_datafiles/config.yaml` sections, and scripting functions.
- Leave world datafiles (`_datafiles/world/*/quests`, `pets`, `conversations`, `audio.yaml`, JS that
  calls removed functions) for Phase 3 and 4. A JS script calling a removed function fails only when
  it runs; the tutorial calls none.
- Saved characters keep loading: removing a struct field from `characters`/`users` must not make an
  existing user YAML fail to load (check the loader is not strict; `_datafiles/world/default/users/1.yaml`
  is the test).
- Done per task: `go build ./...`, `go vet ./...`, `gofmt -l` clean on touched files,
  `go test ./... -count=1` shows only the 3 known Windows failures (fewer packages is expected),
  `go run ./cmd/smoke -bots 5` passes, `grep -rni <name>` over Go, config and admin HTML finds
  nothing left except world datafiles.

## Tasks (in order, one at a time)

1. `clans`, `integrations/discord`, `migration`, `suggestions` (Sonnet)
2. `audio`, `telemetry` (Sonnet)
3. `conversations` (Sonnet)
4. `quests` (Opus: touches `characters`)
5. `pets` (Opus: touches `characters`, `combat`)

## Done when

Phase 1 "Done when" in `gomud_strip_plan.md` holds, `go run ./cmd/smoke -bots 50` passes, and the
web admin still serves its pages. Check that in the smoke container, not with a local server: the
web listener binds every interface (`internal/web/web.go:578`), which raises a Windows firewall
prompt. The controller does this check.
