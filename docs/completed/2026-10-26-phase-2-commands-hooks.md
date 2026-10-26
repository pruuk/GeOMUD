# Phase 2: delete game commands and hooks

Date: 2026-10-26. Parent: `gomud_strip_plan.md` Phase 2 ("Remove the stripped command, mob command
and hook lists from section 1. Done when: `usercommands` and `mobcommands` hold only the kept lists
plus the combat commands waiting for Phase 5"). Branch `strip/phase-2-commands-hooks`, one commit per
task.

## Facts verified against source (2026-10-26, master `ff4b5896`)

User command registry: `internal/usercommands/usercommands.go` (map of name to `{func, adminOnly,
...}`). Mob command registry: `internal/mobcommands/mobcommands.go:24` `mobCommands`.

**User commands, by bucket** (registry keys today):

| Bucket | Commands |
|---|---|
| Keep (plan list) | look, go, exits, say, emote, shout, whisper, broadcast, who, online, quit, save, help, motd, alias, macros, set, password, history, inventory, get, drop, give, put, show, equip, remove, read, lock, unlock, keyring, party, visited, biome, default, print, noop, start |
| Keep (companions of kept commands) | printline (with print), unmute (with mute) |
| Keep (admin, plan list) | room, build, zone, teleport, visit, locate, spawn, reload, server, syslogs, command, mute, grant, rename, redescribe, item, mob, modify, prepare, copyover, badcommands |
| Keep until Phase 5 (combat) | attack, flee, shoot, consider |
| Keep until Phase 5 (death) | suicide: it is the player death routine (corpse, drops, XP loss, recovery room), issued by `NewRound_AutoHeal.go:48` and `NewRound_DoCombat.go:1129` at -10 HP |
| Strip: economy and character info (task 1) | bank, buy, sell, list, offer, appraise (+ `shophooks.go`), jobs, train, experience, killstats, pvp, gearup, share, stash, use, eat, drink, break, zombieact, conditions, cooldowns, status |
| Strip: not on the keep list (task 1) | ask (orders a hired or charmed mob), picklock (lockpicks plus a skill roll) |
| Strip: skills and spells (task 2) | every `skill.*.go` command: aid, backstab, bump, cast, changeform, disarm, dual-wield, enchant, unenchant, uncurse, inspect, map, peep, pickpocket, portal, pray, rank, recover, scribe, search, sneak, tackle, tame, throw, track; plus `skills`, `spells` |
| Strip: admin for stripped systems (task 2) | buff, spell, skillset, rankings, zap, paz, formset |

`unenchant` and `uncurse` live in `skill.enchant.go`. `look.go` special-cases `cast` (`usercommands.go:311`).

**Mob commands** (registry keys today):

| Bucket | Commands |
|---|---|
| Keep (plan list) | go, wander, pathto, look, say, sayto (+ its `replyto`, `saytoonly` keys), emote, shout, broadcast, get, drop, give, put, equip, remove, unlock, despawn, noop |
| Keep until Phase 5 (death) | suicide: the mob death routine, issued by `NewRound_DoCombat.go:1150` and `NewRound_MobRoundTick.go:75` |
| Strip (task 3) | aid, alchemy, attack, backstab, befriend, break, callforhelp, cast, drink, eat, gearup, lookforaid, lookfortrouble, portal, shoot, show, sneak, throw. Engine code that issues one (for example `usercommands/go.go:318` `lookfortrouble`) loses that call |

Mob `attack` and `shoot` come back through the ruleset in Phase 5. Until then a mob cannot start a
fight; a player can still attack a mob (user `attack` stays). The kept mob commands are temporary:
Phase 4b (NPCs are players, decision 4 in `gomud_strip_plan.md`) has NPCs run the player commands and
deletes `mobcommands`, so task 3 should not improve or refactor mob command code.

**Hooks** (`internal/hooks/`):

| Bucket | Hooks |
|---|---|
| Strip (task 3) | `NewRound_SpawnLootGoblin`, `LevelUp_SendLevelNotifications`, `Looking_HandleLookHints`, `CharacterCreated_BroadcastNewChar`, `DayNightCycle_NotifySunriseSunset` |
| Keep (were "review") | `NewRound_UpdateZoneMutators` (ticks zone mutators; `mutators` is kept), `RebuildMap_HandleMapRebuild` (rebuilds the `mapper` cache on room load; `mapper` is kept) |
| Phase 4 (scripting) | `UserScript_HandleUserScript`, `NewRound_PruneVMs`, `ItemOwnership_ItemScriptEvents` |
| Phase 5 (rules) | `Buff_ApplyBuffs`, `NewTurn_PruneBuffs`, `NewRound_DoCombat`, `NewRound_AutoHeal`, `NewTurn_ActionPoints` |
| Keep | everything else |

**Smoke test dependency.** The tutorial (`_datafiles/world/default/rooms/tutorial/901.js`) teaches
`status`, `inventory`, `experience` and `conditions`, in that order, and `cmd/smoke/script.go` types
them. The room's `onCommand` advances the lesson when the typed command matches and then lets the
command run. `903.js` "portal" is a temporary exit named "swirling portal", not the `portal` skill.

## Rules for every task

- Delete the command file, its registration, its help topic files that live in Go packages (for
  example `skill.*.md`, `*.panels.go`), and code that exists only for it. Callers in kept code lose
  only the stripped call.
- Leave world datafiles (help templates under `_datafiles/world/*/templates/help`, `keywords.yaml`
  aliases, mob and room YAML or JS that type a removed command) for Phase 3 and 4. A script or mob
  that issues a removed command gets the normal unknown-command path at runtime.
- Leave the rules packages (`skills`, `spells`, `buffs`, `combat`, `stats`, `races`) and their admin
  pages for Phase 5, even if a command that used them goes now.
- Done per task: `go build ./...`, `go vet ./...`, `gofmt -l` on touched files, `go test ./... -count=1`
  `ok` list diffed against the previous one (name any package that drops out), `go run ./cmd/smoke
  -bots 5` passes, the registry holds only what the tables above say.

## Tasks (in order, one at a time)

1. Economy and character-info user commands, plus ask and picklock (Sonnet). If the tutorial or the
   smoke script needs a removed command (`status`, `experience`, `conditions`), find out how the
   tutorial behaves now and change `cmd/smoke/script.go` to check what the room still does (the
   teacher's reply, the next lesson, the exit unlocking) rather than the removed command's output.
2. Skill and spell user commands, `skills`, `spells`, and the stripped admin commands (Sonnet).
3. Stripped mob commands and hooks (Sonnet).

## Done when

The Phase 2 "Done when" holds (checked against the tables above), `go run ./cmd/smoke -bots 50`
passes, and the admin pages still load in the smoke container (controller).
