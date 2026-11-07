# Phase 5 inventory (2026-11-06, branch strip/phase-5-rules-seam at 9131eac2)

Three read-only inventories written for the Phase 5 (rules seam) plan: M, rules data and the character model; N, combat, the round, death and progression; O, buffs, effects and surfaces. Each lists the files it read. Their open questions go to Calabe (player-visible) or to the plan as defaults. This file moves to `docs/completed/` with the plan.

# Inventory M: rules data and the character model (Phase 5)

Read 67 of 67 files

Repo `new_geo_mudbase`, `master` at `9131eac2` (checked with `git rev-parse --short HEAD`). Read-only: no repo file changed. Module path `github.com/GoMudEngine/GoMud`. "Non-test" means `_test.go` files excluded. Every grep below was run over `internal modules cmd *.go` unless stated.

## Files read

| # | File | Lines read | Full/partial |
|---|---|---|---|
| 1 | `gomud_strip_plan.md` | 1-146 | full |
| 2 | `status.md` | 1-29 | full |
| 3 | `2026-10-27-phase-4b-npcs-are-players-design.md` | 262-369 | partial (the "Notes for later slices" section only) |
| 4 | `internal/stats/stats.go` | 1-87 | full |
| 5 | `internal/statmods/statmods.go` | 1-101 | full |
| 6 | `internal/races/races.go` | 1-192 | full |
| 7 | `internal/races/admin.go` | 1-57 | full |
| 8 | `internal/races/memory.go` | 1-15 | full |
| 9 | `internal/skills/skill.go` | 1-115 | full |
| 10 | `internal/skills/skills.go` | 1-129 | full |
| 11 | `internal/skills/profession.go` | 1-113 | full |
| 12 | `internal/skills/testhelper.go` | 1-18 | full |
| 13 | `internal/skills/admin.go` | 1-155 | full |
| 14 | `internal/spells/spells.go` | 1-179 | full |
| 15 | `internal/spells/admin.go` | 1-49 | full |
| 16 | `internal/spells/memory.go` | 1-15 | full |
| 17 | `internal/spells/summon.go` | 1-8 | full |
| 18 | `internal/characters/character.go` | 1-2128 (chunks 1-400, 401-800, 801-1200, 1201-1600, 1601-2000, 2001-2128) | full |
| 19 | `internal/characters/aggro.go` | 1-30 | full |
| 20 | `internal/characters/alignment.go` | 1-69 | full |
| 21 | `internal/characters/charminfo.go` | 1-25 | full |
| 22 | `internal/characters/cooldowns.go` | 1-44 | full |
| 23 | `internal/characters/formattedname.go` | 1-124 | full |
| 24 | `internal/characters/mobmastery.go` | 1-42 | full |
| 25 | `internal/characters/itemids.go` | 1-127 | full |
| 26 | `internal/characters/kdstats.go` | 1-129 | full |
| 27 | `internal/characters/roombitset.go` | 1-149 | full |
| 28 | `internal/characters/shop.go` | 1-192 | full |
| 29 | `internal/characters/worn.go` | 1-154 | full |
| 30 | `internal/mobs/mobs.go` | 1-652 (chunks 1-400, 401-652) | full |
| 31 | `internal/mobs/saved.go` | 1-93 | full |
| 32 | `internal/mobs/newmobfile.go` | 1-46 | full |
| 33 | `internal/mobs/admin.go` | 1-103 | full |
| 34 | `internal/mobs/memory.go` | 1-21 | full |
| 35 | `internal/mobs/mobs_path.go` | 1-55 | full |
| 36 | `internal/mobs/plugin.go` | 1-100 | full |
| 37 | `internal/mobs/testhelpers.go` | 1-4 | full |
| 38 | `_datafiles/world/empty/races/0-ghostly_spirit.yaml` | 1-22 | full |
| 39 | `_datafiles/world/empty/races/10-rodent.yaml` | 1-12 | full |
| 40 | `_datafiles/world/empty/races/1-human.yaml` | 1-28 | full |
| 41 | `_datafiles/world/empty/mobs/startland/1-rat.yaml` | 1-20 | full |
| 42 | `_datafiles/world/empty/mobs/startland/2-guard.yaml` | 1-24 | full |
| 43 | `_datafiles/world/empty/users/1.yaml` | 1-16 | full |
| 44 | `_datafiles/world/empty/spells/heal.yaml` | 1-9 | full |
| 45 | `_datafiles/world/empty/skills/cast.yaml` | 1-4 | full |
| 46 | `_datafiles/world/empty/professions/warrior.yaml` | 1-4 | full |
| 47 | `_datafiles/world/empty/items/armor-20000/gloves/20016-torn_gloves.yaml` | 1-10 | full |
| 48 | `_datafiles/world/empty/templates/help/races.template` | 1-35 | full |
| 49 | `internal/hooks/NewTurn_ActionPoints.go` | 1-23 | full |
| 50 | `internal/combat/calculations.go` | 355-380 | partial |
| 51 | `internal/combat/combat.go` | 100-110, 228-256 | partial |
| 52 | `internal/usercommands/start.go` | 30-110 | partial |
| 53 | `internal/usercommands/help.go` | 90-130 | partial |
| 54 | `modules/gmcp/gmcp.Char.go` | 675-700 | partial |
| 55 | `internal/accounts/userrecord.go` | 60-75, 170-262, 598-625 | partial |
| 56 | `internal/events/eventtypes.go` | 186-215, 283-290 | partial |
| 57 | `_datafiles/config.yaml` | 364-450 (non-comment lines printed) | partial |
| 58 | `internal/templates/templatesfunctions.go` | 140-165, 350-372 | partial |
| 59 | `internal/rooms/npcstore.go` | 15-45 | partial |
| 60 | `internal/hooks/NewRound_ActorUpkeep.go` | 60-150 | partial |
| 61 | `internal/behaviour/behaviour.go` | 60-76 | partial |
| 62 | `internal/usercommands/go.go` | 40-70 | partial |
| 63 | `internal/usercommands/equip.go` | 100-150 | partial |
| 64 | `internal/rooms/rooms.go` | 2325-2350 | partial |
| 65 | `internal/rooms/roomdetails.go` | 236-246 | partial |
| 66 | `internal/rooms/roommanager.go` | 410-420 | partial |
| 67 | `internal/web/api_v1_users.go` | 92-130 | partial |

Not read (claims about them come from grep lines only, quoted with file:line): every `_test.go` in the scope packages, the rest of `combat`, `hooks`, `usercommands`, `web`, `modules/gmcp`. For partially read files I claim nothing outside the listed lines except grep hits, which are cited.

---

## 1. Packages: files, lines, exports, imports, importers

Importers from `go list -f '{{.ImportPath}} {{join .Imports " "}}' ./...` (60 packages listed), matching ` github.com/GoMudEngine/GoMud/internal/<pkg>( |$)`. The same command finds 11 importers for `buffs` and 3 for `combat`, matching `status.md` ("buffs has 11 importing packages"), so the match works.

**Importer counts: all six expected numbers confirmed.**

| Package | Non-test importers | Expected | Importers |
|---|---|---|---|
| `stats` | **5** | 5 | accounts, characters, combat, events, races |
| `statmods` | **5** | 5 | buffs, characters, combat, items, web |
| `races` | **7** | 7 | root, characters, combat, mobs, usercommands, web, modules/gmcp |
| `skills` | **7** | 7 | root, accounts, characters, templates, usercommands, web, modules/gmcp |
| `spells` | **4** | 4 | root, characters, usercommands, web |
| `characters` | **12** | 12 | root, accounts, actors, combat, hooks, messaging, mobs, rooms, templates, usercommands, web, modules/gmcp |
| `mobs` (context) | 12 | | root, accounts, actors, behaviour, combat, hooks, rooms, templates, usercommands, web, modules/follow, modules/gmcp |

Files per package that import `characters` (non-test, grep of the import path): root 1, accounts 1, actors 2, combat 5, hooks 2, messaging 1, mobs 3, rooms 3, templates 1, usercommands 7, web 3, modules/gmcp 1.

### 1.1 `internal/stats` (1 non-test file, 87 lines; `stats_test.go` 143 lines not read)
- Imports: `math`, `internal/configs` (go list).
- Exports: `Statistics` struct with six fixed fields `Strength, Speed, Smarts, Vitality, Mysticism, Perception` (`stats.go:9-16`); `StatInfo` {`Training` (saved), `Value`, `ValueAdj`, `Racial`, `Base` (saved), `Mods`, `NoCap`} (`stats.go:20-28`); `(*StatInfo).SetMod` (`:30`), `GainsForLevel(level)` (`:46`, formula from `configs.GetProgressionConfig()` `BaseModFactor/BaseModExponent/NaturalGainsModFactor/NaturalGainsExponent`), `Recalculate(level)` (`:61`, stat cap compression from `StatCapThreshold/Anchor/Exponent/Scale/ExemptBonus`).
- Symbol use outside the package (grep `\bstats\.[A-Z]`): `accounts/userrecord.go` `stats.Statistics` x1 (`:214`), `characters/character.go` `StatInfo` x11 + `Statistics` x6, `combat/mob_rank.go` and `combat/weapon_rank.go` (StatInfo x6, Statistics x1 each), `events/eventtypes.go` x1 (`LevelUp.StatsDelta`, `:205`), `races/races.go` x1 (`Race.Stats`, `:44`), `web/api_v1_progression.go` `StatInfo` x2.
- `SetMod`: grep `\.SetMod\b` finds no caller outside `stats` (dead in production; `Recalculate` is the live path).

### 1.2 `internal/statmods` (1 non-test file, 101 lines; test 60 lines not read)
- Imports: none (go list prints an empty list).
- Exports: `StatMods map[string]int` (`:8`) with `Get(names...)` (`:74`) and `Add` (`:91`); `StatName string` (`:9`); name vars `Picklock, Tame, RacialBonusPrefix, Casting, CastingPrefix, XPScale, HealthRecovery, ManaRecovery, Attacks, Damage, Strength, Speed, Smarts, Vitality, Mysticism, Perception, HealthMax, ManaMax` (`:11-39`); `RegisterStatMod` (`:43`); `GetStatMods` (`:47`, help text per name).
- Users: `buffs/buffspec.go` and `items/itemspec.go` (the `StatMods` map type on buff and item specs); `characters/character.go` (13 names, see 2.2); `combat/armor_rank.go` (12 names), `combat/combat.go` (`RacialBonusPrefix`); `web/api_v1_statmods.go` (`GetStatMods`, `StatName`).
- The type is a neutral keyed container already (`map[string]int`). Only the name list is rules.

### 1.3 `internal/races` (3 files, 264 lines)
- Imports: `errors fmt strings time`, `configs fileloader items mudlog stats util`, `yaml.v2`.
- Exports: `Size` + `Small/Medium/Large` (`races.go:18-28`); `Race` struct (`:30-46`: `RaceId, Name, Description, DefaultAlignment int8, BuffIds []int, Size, TNLScale float32, UnarmedName, Tameable, Damage items.Damage, Selectable, AngryCommands, KnowsFirstAid, Stats stats.Statistics, DisabledSlots`); `GetRaces` (`:48`), `GetRace` (`:56`), `FindRace` (`:60`), `(*Race).Id/Validate/Filename/Filepath/Save` (`:81-176`), `(Race).GetEnabledSlots` (`:132`), `LoadDataFiles` (`:179`); `admin.go`: `GetRacesMap` (`:12`), `SaveRace` (`:22`), `DeleteRace` (`:34`), `NewRace` (`:48`); `memory.go`: `GetMemoryUsage` (`:5`).
- `Validate` runs `Stats.X.Recalculate(1)` for all six stats (`races.go:107-112`), so races depends on the stat formula at load.
- `FindRace`: grep `\bFindRace\b` across the tree finds only its definition (dead).

### 1.4 `internal/skills` (5 non-test files, 530 lines; `skills_test.go` 198 lines not read)
- Imports: `fmt os regexp sort strings time`, `configs fileloader mudlog util`.
- Exports: `DefaultMaxLevel = 4` (`skill.go:16`), `Skill` {SkillId, Name, Description, MaxLevel} (`:21-26`), `LoadDataFiles` (`:62`), `GetSkill` (`:77`), `GetAllSkills` (`:82`), `GetAllSkillNames` (`:94`), `SkillExists` (`:104`), `MaxSkillLevel` (`:110`); `skills.go`: `ProfessionRank` (`:12`), `GetProfessionRanks` (`:21`), `GetProfession` (`:61`, returns a title like "apprentice warrior/ranger" or `scrub`/`demigod`), `GetExperienceLevel` (`:110`); `profession.go`: `Profession` (`:18`), `LoadProfessionDataFiles` (`:75`), `GetProfessionSpec` (`:99`), `GetAllProfessions` (`:104`); `admin.go`: `GetSkillsMap, GetProfessionsMap, CreateSkill, SaveSkill, DeleteSkill, CreateProfession, SaveProfession, DeleteProfession` (`:14-155`); `testhelper.go`: `SetTestData` (`:6`, non-test file, for other packages' tests).
- It is a data registry only: no cooldown, no command, no effect code.

### 1.5 `internal/spells` (4 files, 251 lines)
- Imports: `fmt strings time`, `configs fileloader mudlog util`.
- Exports: `SpellType`, `SpellSchool` (`spells.go:14-15`), `SpellData` {SpellId, Name, Description, Type, School, Cost, WaitRounds, Difficulty} (`:17-26`), constants `WaitRoundsDefault` and 7 types and 3 schools (`:28-42`), `HelpOrHarmString` (`:48`), `TargetTypeString` (`:60`), `FindSpell` (`:94`), `GetSpell` (`:106`), `FindSpellByName` (`:113`), `GetAllSpells` (`:134`), `(*SpellData).Id/Filepath/Validate/GetDifficulty` (`:142-164`), `LoadSpellFiles` (`:166`); `admin.go`: `SaveSpellSpec` (`:12`), `DeleteSpellSpec` (`:34`); `memory.go`: `GetMemoryUsage`; `summon.go`: `Summon` (`:3`, a stub that returns `false, nil`).
- No cast path exists in the package.

### 1.6 `internal/characters` (12 non-test files, 3,213 lines; 10 test files, 4,371 lines, not read)
- Imports (go list): `fmt maps math math/bits slices sort strconv strings time`, `buffs colorpatterns configs events gametime ids items mudlog races skills spells statmods stats util`, `yaml.v2`. Six rules packages: buffs, races, skills, spells, statmods, stats (matches `status.md`).
- Files: `character.go` 2128, `aggro.go` 30, `alignment.go` 69, `charminfo.go` 25, `cooldowns.go` 44, `formattedname.go` 124, `itemids.go` 127, `kdstats.go` 129, `mobmastery.go` 42, `roombitset.go` 149, `shop.go` 192, `worn.go` 154.
- Rules-bearing files: `character.go` (stat block, vitals, XP, skills, spells, race reads), `alignment.go` (good/evil bands and `AlignmentAggroThreshold = 190`, `:22`), `mobmastery.go` (tame proficiency), `kdstats.go` (kill/death tallies), `charminfo.go` (charm), `aggro.go` (`SpellCast`, `BackStab` aggro types, `SpellAggroInfo`, `:11-19`). Engine-only files: `itemids.go`, `roombitset.go`, `worn.go` (except `StatMod`), `formattedname.go`, `cooldowns.go` (generic round counters), `shop.go` (economy, out of scope).

### 1.7 `internal/mobs` (rules data only)
`Mob` (`mobs.go:54-77`) holds rules data: `ItemDropChance`, `Hostile`, `Groups`, `Hates`, `AngryCommands`, `CombatCommands`, `BuffIds`, `EliteChance`, `IsElite`, plus the embedded `Character`. Rules code in the package: `NewMobById` spawn math (`:140-203`), `IsTameable` (`:288`), `HatesRace` (`:399`), `HatesAlignment` (`:409`), `HatesMob` (`:431`), `GetAngryCommand` (`:462`), `ConsidersAnAlly` (`:496`), `GetSellPrice` (`:332`), `Despawns` (`:325`), the hostility table `ReduceHostility/IsHostile/MakeHostile` (`:566-606`). See 2.4.

---

## 2. `characters.Character`: rules fields, rules methods, call sites

Classes: **(a)** engine needs a neutral form; **(b)** pure rules, moves to the ruleset; **(c)** the caller is itself rules code that moves with it; **(d)** dead (no non-test caller). Call sites are non-test, outside `internal/{characters,stats,statmods,races,skills,spells}`, from `grep -nE '\.<Name>\b'`. "N" or "O" marks a caller in a sibling inventory's scope.

### 2.1 Fields (all in `character.go:46-99`)

| Field (line) | Saved as | Call sites outside rules pkgs (by package) | Class |
|---|---|---|---|
| `RaceId` (`:54`) | `raceid` | usercommands: `start.go:40,47,98`, `admin.mob.go:240,241,288,293,400`, `help.go:102` (Race.RaceId); web `api_v1_races.go:60`; gmcp `gmcp.Char.go:685` (race kill tally) | (a) engine keeps "entity has a kind" id; race data is (b) |
| `FormRaceId` (`:55`) | `formraceid` | none (grep `\.FormRaceId\b` empty; grep matched `\.RaceId\b` 11 times, so the pattern style works) | (d) for writers: only `ApplyFormChange` sets it and that has no caller (2.2) |
| `Stats` (`:56`) | `stats:` key (struct has no yaml tag), inner `training`, `base` only (`stats.go:21-25`) | combat 15 lines (N); hooks `NewRound_DoCombat.go:217,307,309,407` (N: hostility rounds from Perception, look-for-weapon roll, flee chance from Speed); usercommands `equip.go:118-145` (stat change message after equip), `suicide_npc.go:179,276` (tame skill-up roll); gmcp `gmcp.Char.go:528-546` (Char.Stats, O); accounts `userrecord.go:229-234` (level-up delta) | equip message (a, needs a ruleset-supplied trait list and diff); gmcp (a, O); the rest (c) |
| `Level` (`:57`) | `level` | accounts `userrecord.go:213,226,251,607`, `userrecord.prompt.go:272` (`{lvl}`); combat 8 (N); mobs `mobs.go:142-170` (spawn level, elite bonus); rooms `rooms.go:537-538` (spawn `Level` from room data), `rooms.go:2343` (`CanPvp` minimum level); usercommands `party.go:239,296`, `suicide.go:97,241-249` (N), `suicide_npc.go` 9 lines (N); gmcp `gmcp.Char.go:433,466,569`, `gmcp.Party.go:220,276`, `gmcp.Game.go:77` (OnlineInfo.Level) | display and PvP gate (a); spawn and death math (c). Open question (i)1 |
| `Experience` (`:58`) | `experience` | combat `simulate.go:63`; mobs `mobs.go:169`; usercommands `suicide.go:247-265` (N); gmcp `gmcp.Char.go:570` | (c); gmcp (a, O) |
| `TrainingPoints` (`:59`) | `trainingpoints` | accounts `userrecord.go:198,237,242`, `userrecord.prompt.go:260` (`{tp}`); gmcp `gmcp.Char.go:436` | (c); prompt and gmcp (a) |
| `StatPoints` (`:60`) | `statpoints` | accounts `userrecord.go:199,238,243`, `userrecord.prompt.go:263` (`{sp}`); combat `simulate.go:53,58`; mobs `mobs.go:159,164`; gmcp `gmcp.Char.go:435` (sent as `SkillPoints`) | (c); prompt and gmcp (a) |
| `Health` (`:61`) | `health` | accounts 8 (prompt `{hp} {hp%}` `userrecord.prompt.go:150-179`, wimpy `userrecord.go:620`); behaviour `behaviour.go:71` (stuck NPC loses 10% a round); combat 16 (N); hooks 17 (`Buff_ApplyBuffs.go:64`, `NewRound_ActorUpkeep.go:115-125`, `NewRound_AutoHeal.go:41-77`, `NewRound_DoCombat.go:148-514`; N); mobs `mobs.go:174`; rooms `rooms.go:1196,1282` (`FindDowned`); usercommands `party.go:245-307`, `remove.go:51`, `suicide.go:304`, `suicide_npc.go:33`; gmcp `gmcp.Char.go:467,553`, `gmcp.Party.go:213,269` | (a): "is down", a health fraction for display, and the prompt/GMCP vitals. Arithmetic is (b) |
| `Mana` (`:62`) | `mana` | accounts prompt `userrecord.prompt.go:185-212`; combat `simulate.go:68`; hooks `NewRound_ActorUpkeep.go:115,119`, `NewRound_AutoHeal.go:42,77`; mobs `mobs.go:175`; usercommands `suicide.go:305`; gmcp `gmcp.Char.go:555` | display (a); the rest (b)/(c) |
| `ActionPoints` (`:63`) | `actionpoints` | accounts prompt `{ap}` `userrecord.prompt.go:217`; hooks `NewTurn_ActionPoints.go:16-18` (+1 per turn to max, every actor); mobs `mobs.go:176`; usercommands `go.go:64` | (a): engine keeps the pool and the per-turn tick (plan section 1); the +1 and the max are (b) |
| `Alignment` (`:64`) | `alignment` | combat `simulate.go:77,80`; mobs `mobs.go:190-193` (race default), `:412-426` (`HatesAlignment`); usercommands `start.go:99`, `suicide_npc.go:156-160,253-257`; modules/follow `follow.go:510` (`HatesAlignment`); `online.go:67,76` and gmcp `gmcp.Game.go:78` read `OnlineInfo.Alignment`, a string from `AlignmentName` | (b); the display string is (a) |
| `Gold` (`:65`) | `gold` | 68 lines in 17 files; on `Character`: accounts prompt `{g}` `userrecord.prompt.go:257`; combat `mob_rank.go:108`; usercommands `drop.go:30-65`, `get.go:118,189,263,329`, `give.go:47-163`, `put.go:85,91`, `suicide.go:206-212`, `suicide_npc.go:375-386`; web `api_v1_users.go:207-208`; gmcp `gmcp.Char.go:567` | (a): the kept commands get/drop/give/put move gold. Open question (i)3 |
| `Bank` (`:66`) | `bank` | web `api_v1_users.go:210-211`; gmcp `gmcp.Char.go:568` | (d) for play (bank command stripped in Phase 2); kept in save, admin and GMCP Worth |
| `SpellBook` (`:68`) | `spellbook` | web `api_v1_mobs.go:113`, `api_v1_users.go:99,126-127` (admin edit keeps it) | (b); no game reader (2.2) |
| `Charmed`, `CharmedMobs` (`:69-70`) | `Charmed` not saved on players (`yaml:"-"`), saved for NPCs in `SavedNPC.Charmed` (`mobs/saved.go:21`); `charmedmobs` | `IsCharmed` 15 sites in behaviour, hooks, rooms, usercommands; `GetCharmIds` `accounts/users.go:259`, `usercommands/party.go:279`, `gmcp.Party.go:241`; `TrackCharmed` `accounts/users.go:261`, `hooks/NewRound_ActorUpkeep.go:105`, `usercommands/suicide_npc.go:53`; `RemoveCharm` `ActorUpkeep.go:103`, `suicide_npc.go:51`; `GetCharmedBy` `combat/combat.go:67`; `.Charmed` `accounts/users.go:358-359`, `ActorUpkeep.go:88-102`, `PlayerDespawn_HandleLeave.go:54` | (a) "is following/controlled by X" is engine-shaped; how a charm starts is (b). Nothing starts one: `Charm()` has test callers only (grep `\.Charm\b` non-test returns one comment line, `behaviour.go:123`). Handoff O |
| `Buffs` (`:72`) | `buffs` | O's scope | O |
| `TNLScale` (`:74`) | not saved | none outside (grep `\.TNLScale\b` empty; read internally at `:1272`, set at `:1391`) | (b) |
| `HealthMax`, `ManaMax`, `ActionPointsMax` (`:75-77`) | not saved | HealthMax: accounts 8, behaviour 1, combat 4, mobs 1, usercommands 8, gmcp 4 (lines in the Health row); ManaMax: accounts 7, combat 2, mobs 1, usercommands 2, gmcp 1; ActionPointsMax: hooks `NewTurn_ActionPoints.go:17-18`, mobs `mobs.go:176` | max for display and clamps (a); the formula (b) |
| `Skills` (`:79`) | `skills` | web `api_v1_users.go:98,123-124`; reads through `GetSkillLevel`, `GetSkills`, `GetAllSkillRanks` (2.2) | (b) |
| `Cooldowns` (`:80`) | `cooldowns` | hooks `NewRound_ActorUpkeep.go:82` (`RoundTick` every actor) | (a) as a framework, but nothing sets one: `TryCooldown` has no caller (2.2), so it is unused |
| `KD` (`:83`) | `kd` | usercommands `suicide.go:59-72,103`, `suicide_npc.go:120-122,243-245` (N); gmcp `gmcp.Char.go:681-724` (O) | (c) N |
| `ExtraLives` (`:85`) | `extralives` | accounts `userrecord.go:71,205,223,246-250`; usercommands `start.go:182`, `suicide.go:116-126` (N) | (c) N (permadeath) |
| `MobMastery` (`:86`) | `mobmastery` | combat `calculations.go:368`; usercommands `suicide_npc.go:193-195,290-292` | (c) tame |
| `PlayerDamage`, `LastPlayerDamage`, `KilledBy`, `KillerMob*` (`:91-96`) | not saved | combat `simulate.go:52`, `combat.go:40`; mobs `mobs.go:138`; hooks `NewRound_DoCombat.go:197-205`; usercommands `suicide.go:56-108,308`, `suicide_npc.go:68-104` | (c) N (credit and death) |
| `permaBuffIds` (`:97`) | not saved | via `SetPermaBuffs` `mobs/mobs.go:182`, `mobs/saved.go:55` | O (affects) |

### 2.2 Methods that compute rules (all `character.go`)

| Method (line) | What it computes | Call sites outside rules pkgs | Class |
|---|---|---|---|
| `GetDescription` (`:144`) | swaps true race name for form race name | rooms `roomdetails.go:95`, `rooms.go:1362,2194` are `Room.GetDescription` (different type); on Character: usercommands `admin.mob.go:359`, `look.panels.go:26,45` | (a) display; the form swap is (d) (no form change happens) |
| `TrackPlayerDamage` (`:160`) | per-attacker damage ledger | combat `combat.go:40` | (c) N |
| `GetBaseCastSuccessChance` (`:179`) | cast % from spell difficulty, `cast` skill, casts, Mysticism, `casting*` mods | none, tests none | (d) |
| `CarryCapacity` (`:233`) | `5 + Strength/3` | accounts `userrecord.prompt.go:269`; usercommands `go.go:53`, `inventory.panels.go:49`; gmcp `gmcp.Char.go:490,506` | (a) engine asks "how much can it carry"; formula (b) |
| `DeductActionPoints` (`:237`) | spend AP | usercommands `go.go:58` | (a) engine pool |
| `GetDefaultDiceRoll` (`:357`) | race dice plus Speed/Strength/Perception bonuses | combat `calculations.go:511`, `combat.go:240`, `mob_rank.go:143` | (c) N. Nil deref if the race is missing (`:359-361`, no nil check) |
| `GetSpells`, `HasSpell`, `DisableSpell`, `EnableSpell`, `TrackSpellCast`, `LearnSpell` (`:381-428`) | spellbook ops | none non-test (tests: 2,1,1,1,1,1 refs) | (d) |
| `UnLearnSpell` (`:430`) | | none, tests none | (d) |
| `GrantXP` (`:438`) | XP after server scale and `xpscale` mod | accounts `userrecord.go:178` only (the other `GrantXP` hits are `Account.GrantXP`: `admin.grant.go:60`, `suicide_npc.go:152,249`, `modules/cleanup/cleanup.go:123`) | (c) N. Accounts-only path |
| `TrackCharmed`, `GetCharmIds`, `Charm`, `GetCharmedBy`, `IsCharmed`, `RemoveCharm` (`:465-526`) | charm bookkeeping | see `Charmed` row | (a) except `Charm` (d, tests only) |
| `KnowsFirstAid` (`:489`) | race flag | none (grep hits are the field `races.go:43` and the method body) | (d); the race flag is shown in `races.template:24` |
| `GetHealthAppearance` (`:536`) | health band text | usercommands `look.panels.go:27` | (a) display, bands are rules-flavoured; divides by `HealthMax.Value` |
| `GetAllSkillRanks` (`:560`) | copy of skills | accounts `userrecord.go:609`; templates `templatesfunctions.go:149`; gmcp `gmcp.Char.go:430,656` | (b) feeding profession title (a display) |
| `GetDefense` (`:567`) | sum of armour, x1.5 with a shield, cap 100 | combat `calculations.go:493`, `combat.go:297`, `mob_rank.go:162` | (c) N |
| `GetAdjectives`, `getFormattedName` (`:621`, `:658`) | `downed` when Health < 1; health % with `RenderHealth` | gmcp `gmcp.Room.go:353,381`; `GetMobName`/`GetPlayerName` in rooms `roomdetails.go:263,283`, usercommands `look.go:123` | (a); needs `IsDown` and a health fraction |
| `GetSkills` (`:875`) | copy | gmcp `gmcp.Char.go:639` | (b), O for gmcp |
| `SetSkill` (`:886`), `TrainSkill` (`:911`) | set/raise with cap | none non-test (only `mudlog` strings inside the methods) | (d) |
| `GetSkillLevel` (`:952`) | level, 0 for unknown skill ids | combat `calculations.go:449,468`, `combat.go:203` (`dual-wield`, N); rooms `roomdetails.go:241` (`peep` shows health % on names); usercommands `remove.go:52` (`enchant`, removing a cursed item), `suicide_npc.go:191,288` (`tame`) | (b); `peep` and `enchant` are rules inside engine commands |
| `GetSkillLevelCost` (`:970`), `GetMaxCharmedCreatures` (`:974`), `GetMapSprawlCapacity` (`:987`) | | none, tests none | (d) |
| `GetMemoryCapacity` (`:979`) via `RememberRoom` (`:1026`) | room history length from `map` skill and Smarts | rooms `roommanager.go:417` calls `RememberRoom` on every move | (d) in effect: `roomHistory` is written only (grep `roomHistory`: `character.go:90,134,1006,1028-1032`, no reader). Rules math on the move path for nothing |
| `ApplyHealthChange` (`:1219`) | clamp to -5 on first drop, -10 floor, cancel `cancel-on-combat` buffs | combat `combat.go:36`, `simulate.go:133,152` | (c) N; the -5/-10 thresholds are death rules |
| `ApplyManaChange` (`:1242`), `BarterPrice` (`:1253`), `MovementCost` (`:1371`) | | none, tests none | (d) |
| `XPTNL`, `XPTL`, `XPTNLActual` (`:1261-1293`) | XP curve from `XPBase/XPLevelFactor/XPLevelPower` and `TNLScale` | combat `simulate.go:63`; mobs `mobs.go:169`; usercommands `suicide.go:248`, `suicide_npc.go:66`; gmcp `gmcp.Char.go:569`; accounts prompt `userrecord.prompt.go:221-239` (`{xp} {XP} {xpn} {xp%}`); templates `templatesfunctions.go:360` (`tnl` func) | (c); prompt and gmcp (a) |
| `LevelUp` (`:1295`) | level, TP/SP grants, full restore | accounts `userrecord.go:201,240` only | (c) N; accounts-only, so NPCs never level |
| `Heal` (`:1331`) | adds hp to Health **and hp to Mana** (`:1339` `c.Mana += hp`; the `mana` argument is unused) | hooks `NewRound_AutoHeal.go:69` (N) | (b). Bug, open question (i)5 |
| `HealthPerRound`, `ManaPerRound` (`:1348`, `:1359`) | `1 + mod` | hooks `NewRound_AutoHeal.go:70-71` (N) | (b) |
| `StatMod` (`:1378`) | equipment + buffs sum by key | combat `calculations.go:167`, `combat.go:195` (N); gmcp `gmcp.Char.go:532-542` (O) | (a): a neutral keyed-mod sum; the keys are rules |
| `RecalculateStats` (`:1383`) | racial base, mods, curve, HP/MP max, AP max fixed at 200 (`:1439`), fires `CharacterStatsChanged` for accounts | combat `mob_rank.go:73`, `weapon_rank.go:72,116`; through `Validate` every round for every actor (`NewRound_ActorUpkeep.go:116`) | (b); the engine needs one "refresh derived values" call |
| `AutoTrain` (`:1486`) | spend stat points round-robin | combat `simulate.go:66`; mobs `mobs.go:173` | (c) |
| `CanDualWield` (`:1510`) | `dual-wield` skill > 0 | internal only, `Wear` `:1878` | (b) inside the engine's `Wear` |
| `Validate` (`:1515`) | race fallback to id 1 (`:1521-1523`), level/xp floors, recalc, clamps, alignment clamp, race disabled slots (`:1589-1618`), hands check (`:1620-1639`) | engine-wide (e.g. `NewRound_ActorUpkeep.go:116`, `start.go:100`, `mobs.go:202-203,547`, `saved.go:77-78`) | mixed: (a) shape, (b) contents |
| `HandsRequired` (`:772`) | Large race 1 hand, Small race +1 hand, shooting weapons ignore size | internal only: `Validate` `:1621,1624`, `BestUpgrades` `:1811-1825`, `Wear` `:1867-1939` | (b) inside engine equip. Nil deref if race missing (`:786-787`) |
| `GetRaceId` (`:1648`), `Race` (`:1655`) | form or true race; name, fallback "Ghostly Spirit" | `GetRaceId`: combat `calculations.go:366`, `combat.go:107,235`, `simulate.go:78`; mobs `mobs.go:191,292,436,470`. `Race()`: combat `combat.go:252`; gmcp `gmcp.Char.go:431` | (a) "kind id/name" |
| `RaceSize` (`:1662`), `ApplyFormChange` (`:1673`) | | none, tests none | (d) |
| `IsFormChanged` (`:1669`), `RevertFormChange` (`:1688`) | | only from `ApplyFormChange` (`:1678-1679`) | (d) |
| `UpdateAlignment` (`:1695`) | resisted move away from neutral | usercommands `suicide_npc.go:157,254` (N) | (c) |
| `DecayAlignment` (`:1728`) | quadratic drift to neutral | hooks `NewRound_ActorUpkeep.go:143` (accounts only, `:136` comment "NPC alignment is a Phase 5 rule") | (b) |
| `AlignmentName` (`:1747`) | band name | accounts `userrecord.go:608`, `userrecord.prompt.go:249,253` (`{a} {A}`); hooks `NewRound_ActorUpkeep.go:142,144`; usercommands `suicide_npc.go:155-158,252-255`; gmcp `gmcp.Char.go:432` | (a) display string; bands (b) |
| `GetCooldown`, `TryCooldown`, `GetAllCooldowns`, `PruneCooldowns` (`:692-726`) | | none non-test (tests 1,0,2,1) | (d); `Cooldowns.RoundTick` and `Prune` live |
| `IsDisabled` (`:1138`) | `Health <= 0` | usercommands `go.go:46`, `usercommands.go:281` (other `IsDisabled` hits are `Item.IsDisabled`) | (a) "is down" |
| `Wear`, `RemoveFromBody` (`:1857`, `:2037`) | equip rules: `Health > 0` gates bound items (`:1915-2023`, `:2041`) | engine commands | (a); the "downed can remove bound items" rule is (b) |

Negative-check method for the (d) rows: `grep -rnwE "<Method>"` over non-test Go with the definition line removed, plus test files and `_datafiles` counted separately. The same command returned 5 non-test refs for `CarryCapacity` and 7 for `GetDefense`, so it can find callers.

### 2.3 Accounts-only rules paths (handoff N, flagged for the one-actor ruling)
- XP and levels: `Account.GrantXP` (`accounts/userrecord.go:176-257`) wraps `Character.GrantXP` and `Character.LevelUp`. No NPC path.
- Alignment decay: `decayAlignment(user, ...)` accounts only (`NewRound_ActorUpkeep.go:84-86,137`).
- `CharacterStatsChanged` fires only when `c.userId != 0` (`character.go:1457-1481`).

### 2.4 `mobs.Mob` rules data and methods

| Item (line) | Call sites outside `mobs` | Class |
|---|---|---|
| `NewMobById` level/elite/stat points/xp/AutoTrain/full vitals/race alignment (`mobs.go:140-196`) | spawn path | (b): "starting values" for a new body; Ruleset.NewBody |
| `EliteChance` (`:72`), `IsElite` (`:73`) | `IsElite`: hooks `NewRound_DoCombat.go:203`, usercommands `suicide_npc.go:121,142,228,244`; `EliteChance`: none outside | (b) |
| `ItemDropChance` (`:57`) | combat `mob_rank.go:109,129`; usercommands `suicide_npc.go:349-351` | (c) N (loot) |
| `Hostile` (`:61`) | combat `mob_rank.go:107,216`; rooms `rooms.go:576,1171,1186`; usercommands `admin.mob.go:382-404`, `go.go:309` | (a) behaviour flag |
| `CombatCommands` (`:66`) | hooks `NewRound_DoCombat.go:268,277` | behaviour (a) |
| `AngryCommands` (`:65`) via `GetAngryCommand` (`:462`, falls back to race `AngryCommands`) | hooks `NewRound_DoCombat.go:239` | behaviour (a); race fallback (b). Nil deref if race missing (`:470-471`) |
| `IsTameable` (`:288`, race `Tameable`) | usercommands `suicide_npc.go:191,288` | (c) |
| `HatesAlignment` (`:409`, `AlignmentAggroThreshold`) | modules/follow `follow.go:510` | (b) |
| `HatesRace` (`:399`), `HatesMob` (`:431`), `ConsidersAnAlly` (`:496`), `GetSellPrice` (`:332`), `Despawns` (`:325`) | none anywhere, tests included (grep `-w` over the whole tree) | (d) |
| `MakeHostile`/`IsHostile`/`ReduceHostility` (`:566-606`) | hooks `NewRound_DoCombat.go:217` (duration minus Perception), usercommands `go.go:312`, hooks `NewRound_ActorUpkeep.go:70` | (a) a group-hostility table keyed by userId (player-only key; one-actor ruling issue, N) |
| `BuffIds` (`:71`) | `SetPermaBuffs` `mobs.go:182`, `saved.go:45` | O |

---

## 3. Hard-coded stat names outside the rules packages

Script: `grep -rn <flags> <pattern>` over Go (non-test and test separately), `_datafiles/world/empty/templates`, world YAML and `_datafiles/config.yaml`. Counts are matching lines.

Patterns and one known hit proving each could match:
- P1 full names, case-insensitive, whole word: `strength|speed|smarts|vitality|mysticism|perception`. Known hit: `internal/usercommands/equip.go:120` `Strength: c.Stats.Strength.ValueAdj`.
- P2 abbreviations, case-sensitive: `Str|Spd|Smt|Vit|Mys|Per`. Known hit: `races.template:26` `Str:`.
- P3 vitals identifiers, case-sensitive: `Health|HealthMax|Mana|ManaMax|HealthPerRound|ManaPerRound|ActionPoints|ActionPointsMax`. Known hit: `modules/gmcp/gmcp.Char.go:553` `Hp: user.Character.Health`.
- P4 vitals lowercase: `health|mana|healthmax|manamax|healthrecovery|manarecovery|hp|mp`. Known hit: `internal/usercommands/equip.go:143` `{"health", ...}`.

### 3.1 P1 full stat names, non-test Go, **after removing the `perception` package**
Raw P1 over-counts because `internal/perception` (Phase 4b messaging senses) shares the name: usercommands had 233 raw hits, almost all `perception.` selectors. Filter: drop lines matching `\bperception\.[A-Z]` or the import path. Per word, per package (outside rules packages):

| Package | strength | speed | smarts | vitality | mysticism | perception |
|---|---|---|---|---|---|---|
| accounts | 1 | 1 | 1 | 1 | 1 | 1 |
| actors | | | | | | 3 (comments and a field `Perception: p`, senses, false positives) |
| combat (N) | 8 | 11 | 10 | 4 | 3 | 12 |
| configs | 1 | 2 | 1 | | | 2 |
| hooks | | 1 | | | | 3 |
| messaging | | | | | | 1 (comment, false positive) |
| perception | | | | | | 5 (package name, false positive) |
| term | | 1 (`TELNET_OPT_TERM_SPD` comment, false positive) | | | | |
| usercommands | 3 | 3 | 5 | 3 | 3 | 5 |
| web | | | | 2 | 2 | |
| modules/gmcp | 3 | 3 | 3 | 3 | 3 | 3 |

True hits outside combat (each line read in the grep output):
- `accounts/userrecord.go:229-234` (level-up `StatsDelta` per stat).
- `configs/config.gameplay.go:42,46,50,54,58,62` (comments naming the driving stat for combat curves).
- `hooks/NewRound_DoCombat.go:217,307,309,407` (Perception, Speed reads; N).
- `usercommands/equip.go:108-113,120-125,137-142` (stat snapshot struct, reads, message names); `suicide_npc.go:179,276`.
- `web/api_v1_progression.go:168-169` (comments; the file also uses `stats.StatInfo`).
- `modules/gmcp/gmcp.Char.go:531-542` (values and `StatMod("strength")` style string keys), `:880-890` (JSON payload fields `strength`...`perception`).
- `statmods` names are also used as string literals: `combat/calculations.go:167` `StatMod(\`attacks\`)`, `combat/combat.go:195` `StatMod(\`damage\`)`, gmcp `:532-542`.

Test files (P1 raw, incl. perception package refs): actors 8, characters 10, combat 6, guards 3, hooks 8, items 9, messaging 42, perception 1, statmods 22, templates 10.

### 3.2 P2 abbreviations
Non-test Go: 1 (`internal/mapper/mapper.node.go:11` "Per-symbol", false positive). Tests: `internal/templates` 29 (`panel_test.go:106-108` uses `Str:` as a sample label). Templates: `races.template:26-31` (6). `keywords.yaml:85-90` defines help aliases `str, vit, spd/spe, mys/myst, smt/sma, per/percep/percept` (caught by P1 via the full names on the same lines).

### 3.3 P3 and P4 vitals names, non-test Go per package

| Package | P3 identifiers | P4 lowercase |
|---|---|---|
| accounts | 20 | 15 |
| behaviour | 1 | 1 |
| combat (N) | 24 | 6 |
| configs | 2 | 1 |
| events | 2 | 4 |
| hooks | 26 | 3 |
| mobs | 3 | |
| rooms | 2 | 1 |
| usercommands | 18 | 3 |
| util | | 8 (`util.go:741-744` `HealthClass`, health colour classes) |
| templates | | 2 |
| web | 3 | 8 |
| modules/gmcp | 8 | 3 |
| cmd/smoke | | 1 |

Inside the rules packages for reference: characters 89/11, statmods 4/4, stats 1/2.

### 3.4 Datafiles
- Templates (P1): `help/attack.md` 1, `encumbrance.template` 1, `health.template` 1, `mysticism.template` 3, `perception.template` 4, `races.template` 6, `smarts.template` 2, `speed.template` 3, `stats.template` 8, `strength.template` 2, `vitality.template` 3. One help page per stat exists (`strength/speed/smarts/vitality/mysticism/perception/stats/health/mana.template`).
- Templates (P3/P4): `health.template` 2/15, `mana.template` 1/13, `set-prompt.template` 14/13, `mysticism.template` 1/2, `vitality.template` 1/1, `stats.template` 0/2, `set-wimpy.md` 0/2.
- World YAML (P1): `races/0-ghostly_spirit.yaml:12-18`, `races/1-human.yaml:17-23` (`stats:` keys), `mobs/startland/1-rat.yaml:19` (`vitality: training: -2`), `items/armor-20000/gloves/20016-torn_gloves.yaml:9-10` (`statmods: strength: 1, perception: -1`), `keywords.yaml:85-90`, `skills/enchant.yaml:4`, `skills/search.yaml:4` (descriptions).
- World YAML (P4): `ansi-aliases.yaml` 23 (colour classes like `health-10`, `mana-50`), `buffs-flags` 1, `keywords.yaml` 2, `users` 1 (`users/1.yaml:15` `health: 10`).
- `_datafiles/config.yaml`: P1 7 lines (`:216` English "speed", `:308-323` combat curve comments), P3 3, P4 1. Progression keys that name stats or vitals: `HPBase, HPPerLevel, HPPerVitality, ManaBase, ManaPerLevel, ManaPerMysticism` (`config.yaml:390-396`); default prompt `{hp}{HP}{mp}{MP}` (`config.yaml:453`).

### 3.5 Notes for the guard test (O builds it)
- A word grep for `perception` hits the `internal/perception` package (engine). The guard should look at selectors on the stat container (`.Stats.<Name>`) and string literals equal to ruleset stat keys, not bare words. `go/ast` fits (the existing `internal/guards` tests, e.g. `TestOutputGuardResolvesImports` `output_test.go:110`, show the style).
- `Health`/`Mana` and `{hp}` prompt tokens are the hard call: they are in 14 non-test packages (3.3). Open question (i)2 decides whether they are engine words.

---

## 4. Races: what core reads off a race, and where

| Race field | Readers outside `races` | Purpose |
|---|---|---|
| `Name` | characters `GetDescription` `:146-151`, `GetAdjectives` `:647-648` (form), `Race()` `:1655-1660`; combat `combat.go:252` (racial bonus key `racial-bonus-<name>`); mobs `HatesMob` `:436-437` (dead); usercommands `start.go:44-101`, `help.go:98-130`; gmcp `gmcp.Char.go:431,685-686`; `races.template:19` | display, matching |
| `Description`, `Selectable` | `start.go:45-49,96`; `help.go:116-118`; `races.template:20` | start flow and help |
| `DefaultAlignment` | `start.go:99`; `mobs.go:190-195`; `combat/simulate.go:77-80` | starting value (b) |
| `Size` | characters `HandsRequired` `:786-793`; combat `calculations.go:366-372` (tame chance, `tameSizeModifier` `:307-311`); `races.template:22`; help text `races.template:8-12` | hands per weapon (b), tame (c) |
| `Damage` (dice, attacks, crit buffs) | characters `GetDefaultDiceRoll` `:359-376`; `races.template:23` | unarmed damage (c, N) |
| `UnarmedName` | combat `combat.go:107,236`; `races.template:23` | attack text (N). `combat.go:107` derefs `GetRace(...)` with no nil check |
| `TNLScale` | characters `RecalculateStats` `:1390-1395`; `races.template:21` | XP curve (b) |
| `Stats.<six>.Base` | characters `RecalculateStats` `:1396-1403`; `races.template:26-31` | stat growth (b) |
| `BuffIds` | characters `reapplyPermabuffs` `:2071-2075`; `races.template:32` | racial permanent affects (O) |
| `DisabledSlots`, `GetEnabledSlots` | characters `Validate` `:1589-1618`; `races.template:16,25`; admin JS `applyRaceEquipDisabled` (`users.html:600`, `mobs.html:790`) | can't-equip slots (a: equipment is engine, the slot list per kind is data) |
| `KnowsFirstAid` | `KnowsFirstAid()` (dead); `races.template:24` | display only |
| `Tameable` | mobs `IsTameable` `:292-296` | (c) |
| `AngryCommands` | mobs `GetAngryCommand` `:469-474` | behaviour fallback |
| existence | characters `Validate` `:1521-1523` sets `RaceId = 1` if unknown; `ApplyFormChange` `:1674` | fallback is world-specific (race 1 is `human` in `empty`) |

Admin and loaders: `main.go:1512` (`races.LoadDataFiles`), web `api_v1_races.go` (CRUD), `api_v1_yaml.go:158-165`, usercommands `admin.mob.go` (`GetRace`, `GetRaces`).

World: three races (`0-ghostly_spirit`, `1-human`, `10-rodent`). The `rodent` file has no `stats:` block (`10-rodent.yaml:1-12`); `ghostly_spirit` and `human` omit `speed` and `mysticism` (base 0 by default).

---

## 5. Skills and spells after the Phase 2 strip

**Registries that remain:** `skills` (skill and profession registries, `skill.go:28`, `profession.go:25`), `spells` (`spells.go:44-46`). Loaded at boot `main.go:1507,1512-1514`; reload `usercommands/admin.reload.go:43-44`. World folders: `skills/` 16 files, `professions/` 10, `spells/` 8 (listing of `_datafiles/world/empty`).

**Who reads skills at runtime (non-test):**
- `GetSkillLevel(<literal id>)`: `dual-wield` (combat `calculations.go:449,468`, `combat.go:203`; N), `peep` (rooms `roomdetails.go:241`, shows health % on names in `look`), `enchant` (usercommands `remove.go:52`, cursed-item removal), `tame` (usercommands `suicide_npc.go:191,288`). Inside characters: `cast` (`:195`, dead method), `dual-wield` (`:1511`, live via `Wear`), `map` (`:980,988`, no effect, see 2.2), `tame` (`:975`, dead).
- Profession title: `skills.GetProfession` in accounts `userrecord.go:609` (who/online), gmcp `gmcp.Char.go:430` (Char.Info `Class`); `GetProfessionRanks` gmcp `:657`; `MaxSkillLevel` gmcp `:643`, usercommands `remove.go:52`; template func `profession` (`templatesfunctions.go:147-151`).
- Web: `api_v1_skills.go`, `api_v1_professions.go`, `api_v1_yaml.go:186,204`, routes `api_routes.go:153-163`, admin pages (`admin_items.go:33-34,164`). O's scope.

**Who reads spells (non-test):** usercommands `help.go:168-175` (`help spell <name>`), web `api_v2_spells.go`, `api_v1_yaml.go`. Characters `GetBaseCastSuccessChance` (`:181`, dead). `aggro.go` still has `SpellCast` aggro type and `SpellAggroInfo` (`:12,16-19`); N reports any reader.

**No cast path, no spell effect, no skill command:** grep `GetBaseCastSuccessChance|LearnSpell|TrackSpellCast` non-test returns nothing (2.2 table). `spells.Summon` is a stub (`summon.go:3-8`).

**Cooldowns:** `Cooldowns` map ticked every round for every actor (`NewRound_ActorUpkeep.go:82`) and pruned in `Validate` (`character.go:1569`); nothing adds one (`TryCooldown` has no caller). Engine framework with no user.

**Dead after the strip (no non-test caller, verified per 2.2):** `GetBaseCastSuccessChance`, all six spellbook methods plus `UnLearnSpell`, `SetSkill`, `TrainSkill`, `GetSkillLevelCost`, `GetMaxCharmedCreatures`, `GetMapSprawlCapacity`, effectively `GetMemoryCapacity`/`RememberRoom` (write-only history), `KnowsFirstAid`, `Charm`, form change (`ApplyFormChange`, `RevertFormChange`, `IsFormChanged`, `RaceSize`), cooldown accessors, `ApplyManaChange`, `BarterPrice`, `MovementCost`; `races.FindRace`; `stats.SetMod`; mobs `HatesRace`, `HatesMob`, `ConsidersAnAlly`, `GetSellPrice`, `Despawns`. Template funcs `tnl` (`templatesfunctions.go:111`) and `profession` (`:147`) have no template user in `empty` (grep `\bprofession\b|\bTNL\b` over `templates` empty; the same grep style finds `buffname` at `races.template:32`).

---

## 6. Persistence

**Where a `Character` is saved**
- User file: `users/<id>.yaml`, `character:` block (`users/1.yaml:7-16`), via `accounts.SaveUser` (`accounts/users.go:382,402,442`, `userrecord.go:255`).
- Mob template: `mobs/<zone>/<id>-<name>.yaml` `character:` block (`1-rat.yaml:13-20`, `2-guard.yaml:9-22`), via `writeTemplateFile` (`mobs/admin.go:25-34`).
- Resting NPCs: room sidecar `rooms.instances/<zone>/<id>.npcs.yaml`, `resting:` list of `SavedNPC` whose `mob:` holds the whole `Mob` including `Character` (`rooms/npcstore.go:26-36`, `mobs/saved.go:15-22`). No sidecar files are in the repo (`find _datafiles -name '*.npcs.yaml'` empty).
- Web admin JSON: `Character` goes over the admin API with Go field names (no json tags): `users.html:599-646` and `mobs.html:789-840` read `ch.RaceId, ch.Level, ch.Experience, ch.Gold, ch.TrainingPoints, ch.StatPoints, ch.ExtraLives, ch.Alignment, ch.Stats, ch.Skills, ch.SpellBook`.

**Rules fields that are saved** (yaml tags, `character.go:54-88`): `raceid`, `formraceid`, `stats` (only `training` and `base` per stat, `stats.go:21,25`), `level`, `experience`, `trainingpoints`, `statpoints`, `health`, `mana`, `actionpoints`, `alignment`, `gold`, `bank`, `spellbook`, `charmedmobs`, `buffs`, `skills`, `cooldowns`, `kd`, `extralives`, `mobmastery`. Not saved: `TNLScale`, `HealthMax`, `ManaMax`, `ActionPointsMax`, `Aggro`, `Charmed` (players), damage ledger, killer fields (`yaml:"-"`, `:69,74-78,91-96`). NPC sidecars also save `aggro` and `charmed` (`saved.go:20-21`).

**Field names in use in `_datafiles/world/empty`:**
- `users/1.yaml`: `raceid: 1`, `level: 1`, `health: 10`.
- `mobs/startland/1-rat.yaml`: `level: 1`, `raceid: 10`, `stats: vitality: training: -2`.
- `mobs/startland/2-guard.yaml`: `raceid: 1`, `level: 10`, `alignment: 30`, `gold: 3`.
- `races/*.yaml`: `raceid, name, description, defaultalignment, size, angrycommands, unarmedname, tnlscale, tameable, selectable, knowsfirstaid, stats.<stat>.base, damage.diceroll, disabledslots`.
- `items/.../20016-torn_gloves.yaml:8-10`: `statmods: {strength: 1, perception: -1}`.
- `skills/*.yaml`: `skillid, name, description, maxlevel` (`cast.yaml`). `spells/*.yaml`: `spellid, name, description, type, school, cost, waitrounds, difficulty` (`heal.yaml`). `professions/*.yaml`: `professionid, name, description, skills` (`warrior.yaml`).
- `config.yaml` `GamePlay.Progression` (`:368-441`): stat curve, HP/Mana per Vitality/Mysticism, TP/SP per level, XP curve, MaxLevel, stat cap.

**What a seam move does to save files**
- YAML: `stats.Statistics` is a struct with lowercase yaml keys (`stats.go:10-15`). A `map[string]StatInfo`-style keyed container with the same lowercase keys and the same `training`/`base` inner keys decodes the same YAML. So `stats:` blocks in users, mob templates, sidecars and race files need no rewrite if the container keeps the key `stats` and string keys `strength` etc. Moving any other field (level, experience, alignment, skills, spellbook...) into a ruleset-owned sub-block changes its path and needs a load-time migration (`accounts/migration.go:25-92` already runs user migrations at boot).
- JSON (web admin): struct field names `Strength`..`Perception` become map keys `strength`..`perception`; `users.html:623` and `mobs.html:840` (`const st = ch.Stats`) and the race admin pages need edits (O).
- `events.LevelUp.StatsDelta stats.Statistics` (`eventtypes.go:205`): its only consumers are the builder in accounts; gmcp registers for `LevelUp` (`gmcp.Char.go:46,332`, `gmcp.Game.go:32`) but grep `StatsDelta` outside events/accounts finds nothing, so the field can go or change type freely.

---

## 7. Proposal: the data half of `Ruleset` (PROPOSAL, not decided)

Grounded in the (a) rows: what engine code asks of a body today.

```go
// internal/rules: the seam. Engine packages import this, never diku-basic.
package rules

type Ruleset interface {
    // Bodies
    NewBody(c *characters.Character, spawn SpawnInfo)   // starting values: mobs.go:140-196, start.go:96-100, accounts New (userrecord.go:71)
    Refresh(c *characters.Character) (changed bool)      // derived values: replaces RecalculateStats (character.go:1383), called from Validate
    ClampVitals(c *characters.Character)                 // Validate's clamps (character.go:1554-1567)

    // Questions the engine asks (each maps to a live call site)
    IsDown(c) bool                         // IsDisabled, GetAdjectives "downed", FindDowned (rooms.go:1196,1282), Wear bound items
    Pools(c) []Pool                        // {Key "health", Label "HP", Cur, Max}: prompt {hp}{mp}, GMCP Char.Vitals, party %, wimpy
    CarryCapacity(c) int                   // go.go:53, inventory.panels.go:49, prompt, GMCP
    MoveCost(c, encumbered bool) int       // go.go:51-55 (10 or 50 today)
    ActionPointTick(c) (gain, max int)     // NewTurn_ActionPoints.go:16-18, RecalculateStats :1439
    HandsRequired(c, it items.Item) int    // Validate, Wear, BestUpgrades
    EnabledSlots(c) []items.ItemType       // race DisabledSlots (Validate :1589-1618)
    PermanentAffects(c) []int              // race BuffIds (reapplyPermabuffs :2071-2075)
    Summary(c) Summary                     // {Level, Class, Alignment string}: who/online (userrecord.go:604-615), GMCP Char.Info, party, PvP min level (rooms.go:2343)
    PromptToken(c, token string) (string, bool) // {lvl} {xp} {tp} {sp} {a} {A} {g}: userrecord.prompt.go:216-272
    TraitList(c) []TraitView               // equip.go stat-change line, GMCP Char.Stats, help pages
    Kinds() []KindInfo                     // races for start.go and help.go (Name, Description, Selectable)
}
```

The combat half (attack, damage, death, XP, regen amounts) is N's; buffs effects are O's.

**Keyed traits container (proposal).** Replace `stats.Statistics` and the three `stats.StatInfo` maxima with one engine type:

```go
package traits
type Trait struct {          // today's StatInfo, same yaml keys
    Training int  `yaml:"training,omitempty"`
    Base     int  `yaml:"base,omitempty"`
    Mods, Value, Adjusted int `yaml:"-"`
}
type Set map[string]*Trait   // keys registered by the ruleset ("strength", ...)
```

`Character.Stats` becomes `traits.Set` with the yaml key `stats` unchanged (see section 6). Mods keep coming from the existing keyed sums `Worn.StatMod` (`worn.go:130`) and `Buffs.StatMod` (`buffs.go:85`), since `statmods.StatMods` is already `map[string]int`. The engine never names a key; diku-basic registers six. `HealthMax`/`ManaMax`/`ActionPointsMax` move into the set or into `Pools`.

## 8. Open questions

### (i) Player-visible (needs Calabe)
1. **Is Level an engine concept?** Today it gates PvP (`rooms.go:2343`), shows in who/party/GMCP, sets spawn level from room data (`rooms.go:537`) and is the `{lvl}` prompt token. Engine field or ruleset trait?
2. **Are health and mana engine words?** 14 non-test packages name them (3.3), the default prompt uses `{hp}{HP}{mp}{MP}` (`config.yaml:453`), GMCP sends `Hp/HpMax/Sp/SpMax` (`gmcp.Char.go:553-556`), colour classes `health-N`/`mana-N` live in `ansi-aliases.yaml`. Keep "health" as the engine's down/alive pool and make mana a ruleset pool, or make both ruleset pools?
3. **Gold:** the kept commands get/drop/give/put/look move it (section 2.1). Engine field, or ruleset currency? Bank is unused in play.
4. **Regen bug:** `Heal` adds the health amount to mana and ignores the mana amount (`character.go:1339`). Fix it in diku-basic (mana regen changes from `1+healthrecovery` to `1+manarecovery`), or keep parity until the seam lands?
5. **NPCs and XP/levels/alignment decay:** all are accounts-only today (2.3). Under the one-actor ruling, does the ruleset decide for every actor (NPCs could then level), or is "players level, NPCs don't" a rule diku-basic keeps?
6. **Dead rules data on characters** (spellbook, form change, cast chance, barter, tame mastery, skills `map/peep/enchant`): delete before the move, or carry into diku-basic so the world's `skills/spells/professions` folders keep meaning? `peep` and `enchant` are live in kept commands (`roomdetails.go:241`, `remove.go:52`).
7. **Race fallback:** an unknown race becomes race 1 (`character.go:1521-1523`, `human` in `empty`). Keep, or fall back to race 0 (ghostly spirit, the character-creation form)?

### (ii) Engineering defaults (my pick)
1. Seam packages: `internal/rules` (interface, registry, `Active()`), `internal/rules/dikubasic` (moves `stats`, `statmods` names, `races`, `skills`, `spells`). Pick: yes.
2. Traits container keeps the yaml key `stats` and lowercase keys so no save-file migration is needed (section 6). Pick: yes; admin JSON pages change (O).
3. Delete the dead code in section 5 in a first slice, before moving anything, to shrink the seam. Pick: yes, except items Calabe keeps under (i)6.
4. Ruleset returns safe defaults where code derefs a missing race today (`combat.go:107`, `character.go:359-361,786-787`, `mobs.go:436-437,470-471`). Pick: yes.
5. Drop `events.LevelUp.StatsDelta` (no consumer outside the builder). Pick: drop.
6. `CharacterStatsChanged` fires from `Refresh` returning `changed`, for every actor with an account (same as today). Pick: yes.
7. Progression config (`config.yaml:368-441`) stays where it is for now and is read only by diku-basic. Pick: yes; move the section when the ruleset gets its own config.
8. `RememberRoom`/`roomHistory`: delete (write-only). Pick: delete.
9. The mob hostility table keyed by `userId` (`mobs.go:22-26,566-606`) is a player-only key; re-key by body ID when N moves hostility. Pick: N's call, flagged.
10. Guard test (O): AST-based, matching `.Stats.<Name>` selectors and string literals equal to registered trait keys, so the `internal/perception` package does not trip it (3.5). Pick: yes.

### Handoffs
- N: `GrantXP`/`LevelUp` accounts-only; `ApplyHealthChange` -5/-10 thresholds; `GetDefaultDiceRoll`, `GetDefense`, `TrackPlayerDamage`, KD, ExtraLives, MobMastery; `DoCombat` Perception/Speed reads (`:217,307,309,407`); `simulate.go`, `mob_rank.go`, `weapon_rank.go` build Characters with `stats.Statistics`; `Heal` bug lives in AutoHeal's path.
- O: buffs, `permaBuffIds`, race `BuffIds`, charm start (nothing calls `Charm`), GMCP `Char.Stats/Vitals/Worth/Skills/Jobs/Kills` (`gmcp.Char.go:430-724`), web `api_v1_statmods/progression/races/skills/professions/spells/users`, admin HTML stat fields, help templates per stat, `keywords.yaml:85-90`, the guard test.


# Inventory N: combat, the round, death and progression (Phase 5)

Read 54 of 54 files.

Master `9131eac2`, read 2026-11-06. Read-only: no repo file changed (`git status --short` empty after the test run below). Every fact below is from the lines listed here; nothing outside them is claimed. "grep" rows mean only the matched lines were read.

Tests run as evidence of today's state: `go test ./internal/hooks/ -run "CombatGolden|KilledBy|ShootWait|TestReadsE|TestReadsH" -count=1` ok; `go test ./internal/combat/ ./internal/guards/ -count=1` ok.

| # | File | Lines read |
|---|---|---|
| 1 | `gomud_strip_plan.md` | full (1-146) |
| 2 | `status.md` | full (1-29) |
| 3 | `docs/completed/2026-10-26-phase-4-scripting-surface.md` | full (1-79) |
| 4 | `2026-10-27-phase-4b-npcs-are-players-design.md` | partial 262-369 |
| 5 | `docs/completed/2026-10-28-phase-4b-5-one-combat-path-plan.md` | partial 30-109, plus grep |
| 6 | `docs/completed/2026-10-28-phase-4b-5-inventory.md` | grep only (337, 353-357, 423-424, 432, 498) |
| 7 | `docs/completed/2026-10-28-phase-4b-7-inventory.md` | grep only (583, 770, 853, 1050, 1055) |
| 8 | `internal/combat/combat.go` | full (1-345) |
| 9 | `internal/combat/attackresult.go` | full (1-16) |
| 10 | `internal/combat/calculations.go` | full (1-572) |
| 11 | `internal/combat/mob_rank.go` | partial 1-60, plus symbol grep |
| 12 | `internal/combat/weapon_rank.go` | partial 1-30, plus symbol grep |
| 13 | `internal/combat/armor_rank.go` | partial 1-30, plus symbol grep |
| 14 | `internal/combat/simulate.go` | partial 90-188, plus symbol grep |
| 15 | `internal/hooks/hooks.go` | full (1-73) |
| 16 | `internal/hooks/NewRound_DoCombat.go` | full (1-521) |
| 17 | `internal/hooks/NewRound_AutoHeal.go` | full (1-89) |
| 18 | `internal/hooks/NewTurn_ActionPoints.go` | full (1-23) |
| 19 | `internal/hooks/NewRound_ActorUpkeep.go` | full (1-154) |
| 20 | `internal/hooks/NewRound_ActorTick.go` | full (1-25) |
| 21 | `internal/hooks/NewRound_RoomRoundTick.go` | full (1-80) |
| 22 | `internal/hooks/NewRound_HandleRespawns.go` | full (1-26) |
| 23 | `internal/hooks/PlayerDrop_HandlePlayerDrop.go` | full (1-48) |
| 24 | `internal/hooks/Buff_ApplyBuffs.go` | partial 55-75 |
| 25 | `internal/hooks/combat_golden_test.go` | full (1-561) |
| 26 | `internal/hooks/reads_golden_test.go` | partial 1-40, 716-766, 831-908 |
| 27 | `internal/usercommands/attack.go` | full (1-289) |
| 28 | `internal/usercommands/flee.go` | full (1-27) |
| 29 | `internal/usercommands/consider.go` | full (1-84) |
| 30 | `internal/usercommands/shoot.go` | full (1-140) |
| 31 | `internal/usercommands/suicide.go` | full (1-313) |
| 32 | `internal/usercommands/suicide_npc.go` | full (1-443) |
| 33 | `internal/usercommands/usercommands.go` | partial 175-200, plus grep (26-29, 132-188, 220-317) |
| 34 | `internal/usercommands/go.go` | partial 30-75, 290-345, 365-390 |
| 35 | `internal/characters/character.go` | partial 155-180, 230-250, 355-380, 434-475, 565-580, 990-1025, 1095-1140, 1215-1382, 1430-1460, 1690-1745 |
| 36 | `internal/characters/alignment.go` | grep only (13-15) |
| 37 | `internal/accounts/userrecord.go` | partial 170-260, 610-650 |
| 38 | `internal/rooms/rooms.go` | partial 2295-2356 |
| 39 | `internal/rooms/roommanager.go` | partial 290-330, 435-470 |
| 40 | `internal/rooms/ephemeral.go` | partial 205-240, plus grep (20, 89, 153, 284) |
| 41 | `internal/configs/config.gameplay.go` | full (1-287) |
| 42 | `internal/configs/config.lootgoblin.go` | full (1-40) |
| 43 | `internal/configs/config.timing.go` | partial 51-57 |
| 44 | `_datafiles/config.yaml` | grep only (lines quoted in section 7) |
| 45 | `internal/util/util.go` | partial 168-200, 540-560 |
| 46 | `internal/items/attack_messages.go` | partial 40-52 |
| 47 | `internal/items/items.go` | partial 100-112 |
| 48 | `internal/behaviour/behaviour.go` | partial 55-85 |
| 49 | `internal/mobs/mobs.go` | partial 575-615, plus grep |
| 50 | `internal/guards/combat_test.go` | partial 1-70 |
| 51 | `cmd/smoke/fight.go` | full (1-189) |
| 52 | `cmd/smoke/script.go` | partial 215-240 |
| 53 | `cmd/smoke/main.go` | grep only (37-46) |
| 54 | `cmd/smoke/testdata/pre-5/bot0.txt` | head 1-5, grep only |

Also grepped (matched lines only, not counted as read): `_datafiles/world/empty/races/{0-ghostly_spirit,1-human,10-rodent}.yaml` (angrycommands), `internal/events/eventtypes.go:328-334`, the testdata dirs (`ls`).

---

## 0. Stale notes in the spec (read this first)

The spec's "Slice 5 inherits" note (`2026-10-27-phase-4b-npcs-are-players-design.md:296-299`) predates slice 5. Slice 5 fixed most of it (plan `docs/completed/2026-10-28-phase-4b-5-one-combat-path-plan.md:39,74-79`). Current state is in section 6. Phase 5 should drop that note rather than plan around it.

---

## 1. `internal/combat`

### Files (non-test 6, test 3, plus upstream `AGENTS.md`/`CLAUDE.md`)

| File | Lines | Exported | Unexported |
|---|---|---|---|
| `combat.go` | 345 | `Attack` (29), `CreditAccount` (63), `GetWaitMessages` (97) | `accountOf` (53), `combatMessage` (78), `buildCombatMessages` (132), `calculateCombat` (188) |
| `attackresult.go` | 16 | `AttackResult` (6): `Hit`, `Crit`, `BuffTarget`, `DamageToTarget`, `Messages` | |
| `calculations.go` | 572 | `Hits` (112), `Crits` (200), `Dodges` (253), `ChanceToTame` (365), `AlignmentChange` (378), `CombatOdds` (546) | `statDelta` (29), `statDeltaProportional` (53), `resolveAttackWeapons` (68), `damageBonus` (85), `hitChance` (99), `extraAttackCount` (134), `weaponlessAttackCount` (147), `combatAttackCount` (158), `critChance` (176), `critMultiplier` (215), `critDamageBonus` (228), `dodgeChance` (241), `dualWieldHitPenalty` (262), `dualWieldActiveWeaponCount` (271), `applyDefenseReduction` (286), `damagePercentOfMax` (298), `tameSizeModifier` (307), `tameHealthBonus` (320), `chanceToTame` (325), `expectedDPS` (437) |
| `mob_rank.go` | 236 | `MobRank` (17), `RankMobs` (83), `FormatMobRankings` (204) | `baselineCharacterAtLevel` (52) |
| `weapon_rank.go` | 220 | `WeaponRank` (14), `RankWeapons` (87), `FormatWeaponRankings` (193) | `baselineCharacter` (58) |
| `armor_rank.go` | 266 | `ArmorRank` (20), `RankArmor` (156), `FormatArmorRankings` (235) | `armorBaseHP` (17), `wornBuffCount` (75), `wornBuffValue` (86), `armorSlotSet` (99), `statWeight` (109) |
| `simulate.go` | 188 | `SimResult` (13), `SimResult.String` (28), `SimulateCombat` (98) | `newSimMob` (41), `applySimBuffs` (184) |
| tests | `calculations_test.go` 436, `simulate_test.go` 93, `weapon_rank_test.go` 68 | | |

Non-test total: 1,843 lines (`wc -l`).

Package imports (`combat.go:3-20`, `calculations.go:3-13`, rank file headers): `accounts`, `actors`, `characters`, `configs`, `ids`, `items`, `mudlog`, `perception`, `races`, `rooms`, `statmods`, `util`, `mobs`, `stats`, `buffs`.

### Importers: 3, confirmed

`go list -f '{{.ImportPath}} {{join .Imports " "}}' ./... | grep internal/combat` gives `hooks`, `usercommands`, `web` (plus `combat` itself as the ImportPath). Test-only importers: none outside the package. Files: `grep -rln 'internal/combat"'`:

| Importer file | Symbols used |
|---|---|
| `internal/hooks/NewRound_DoCombat.go` | `GetWaitMessages` (167), `Attack` (182), `CreditAccount` (214) |
| `internal/usercommands/consider.go` | `CombatOdds` (72, 77) |
| `internal/usercommands/suicide_npc.go` | `AlignmentChange` (156, 253) |
| `internal/web/api_v1_items.go` | `RankWeapons` (18), `RankArmor` (34) |
| `internal/web/api_v1_mobs.go` | `RankMobs` (25) |

`internal/util/util_test.go:181-185` names a local `combat` variable, not the package (grep hit, false positive).

### Dead or test-only exports (grep of the name over `*.go`; each grep matched its own definition, so it could match)

- `ChanceToTame` (`calculations.go:365`): only its own comment line 364. Dead.
- `SimulateCombat`: only `simulate_test.go` (23, 55, 72, 86).
- `FormatMobRankings`, `FormatArmorRankings`: only their definitions. `FormatWeaponRankings`: `weapon_rank_test.go:66` only.
- `Hits`, `Crits`, `Dodges`: only `combat.go:273, 287, 275`. Exported but package-internal in use.

### Guard tie-in (handoff to O)

`internal/guards/combat_test.go:41-53` hard-codes `internal/combat/combat.go` func `Attack` and `internal/hooks/NewRound_DoCombat.go` func `DoCombat` as positive controls; `TestCombatPackageDoesNotSend` (60-70) scans `internal/combat/`. Moving the math into a ruleset package breaks these controls on purpose; Phase 5 must update them in the same commit.

---

## 2. The round, step by step

Registration order (`internal/hooks/hooks.go:18-38`). NewRound: `InactivePlayers`, `UpdateZoneMutators`, `CheckNewDay`, `RoomRoundTick`, `ActorUpkeep`, `HandleRespawns`, `DoCombat` (27), `AutoHeal` (31), `ActorTick` (32). NewTurn: `CleanupLinkDead`, `AutoSave`, `PruneBuffs`, `ActionPoints` (38). Pulse: 50 ms turns, 4 s rounds (`_datafiles/config.yaml:151,155`).

Not mine and engine-only, read for completeness: `RoomRoundTick` (idle room messages, 5% or 20% roll, `NewRound_RoomRoundTick.go:27-30,44`: flavour, not combat), `HandleRespawns` (`room.Prepare`, `NewRound_HandleRespawns.go:14-23`), `ActorTick` (`Controller().Tick()`, `NewRound_ActorTick.go:15-25`). `PruneBuffs` is O's.

### 2a. `ActorUpkeep` (`NewRound_ActorUpkeep.go:63-132`), before combat

| Lines | Step | Class | Ruleset seam |
|---|---|---|---|
| 70 | `mobs.ReduceHostility()` (group hatred countdown) | engine | |
| 82 | cooldowns tick | engine | |
| 84-86, 137-154 | alignment decay, accounts only, every `AlignmentDecayRounds` | rules (decay amount `character.go:1728-1745`; period) | `Ruleset.RoundUpkeep(actor, round)` replaces 84-86; drop the accounts-only guard (spec note: "alignment decay is accounts-only (Phase 5)", design:316) |
| 88-90, 100-112 | charm countdown and expiry | engine | |
| 92-98 | `Buffs.Trigger()` then `BuffsTriggered` event | engine timer (O's) | |
| 115-123 | `Validate()` (stat refresh), vitals event for accounts | rules inside `Validate` (M's) | |
| 125-128 | NPC with `Health <= 0` queues `suicide` | engine dispatch; threshold is rules | `Ruleset.Condition(actor)` |

### 2b. `DoCombat` (`NewRound_DoCombat.go:28-43`)

| Lines | Step | Class | Ruleset seam |
|---|---|---|---|
| 33, 49-62 | `fightersWithAccountsFirst`: accounts (map order) then NPCs (map order) | rules (initiative; plan decision 5 parks it for Phase 5, slice-5 plan :70-71) | `Ruleset.TurnOrder([]*actors.Actor) []*actors.Actor` at 33 |
| 82-84 | nil and no-aggro skip | engine | |
| 87-89 | `no-combat` buff flag skips the turn | engine (flag); what grants it is O's | |
| 91-95 | no room: end aggro | engine | |
| 98, 146 | cancel `cancel-on-combat` buffs on both | engine | |
| 102-106, 267-296 | NPC `CombatCommands`: `util.Rand(100) >= ActivityLevel` then a random command | engine (NPC behaviour data) | keep in engine/behaviour; note it draws RNG before the attack |
| 108-111, 389-473 | flee resolution | engine except the chance | see 2c |
| 121-138 | target lookup; same room or through `Aggro.ExitName` | engine (targeting, range) | |
| 140-144 | defender room missing: end aggro | engine | |
| 148-152 | target `Health < 1`: "Your rage subsides", end aggro | engine; threshold is rules | `Ruleset.Condition` |
| 158-160, 300-322 | NPC with no weapon may equip one: `roll < Perception.ValueAdj` (307-309) | rules (names a stat; the guard test will catch it) | move the roll behind `Ruleset` or behaviour asks the ruleset for a chance |
| 162-172 | `RoundsWaiting` countdown; wait message | engine loop; the wait count is rules (set in `SetAggro`, `character.go:1106-1114`, weapon plus offhand `WaitRounds`, reset every round by 252) | `Ruleset.AttackDelay(actor) int` replaces `character.go:1113` |
| 175-178 | hidden target: skip ("can't seem to find") | engine (senses) | |
| 180-182 | `combat.Attack(a, d)` | rules (whole resolution) | `Ruleset.ResolveAttack(a, d) AttackResult` at 182; the health apply, credit and wimpy in `combat.go:36-47` stay engine |
| 184-186 | crit buffs queued as `events.Buff` | engine apply; which buffs is rules (`BuffTarget`) | in `AttackResult` |
| 188-190 | send messages in order | engine | |
| 196-207 | combat downs a player: record `KilledBy`, `KillerMob*` | engine bookkeeping | |
| 209-220 | group hatred: `MakeHostile(group, acct, MinutesToRounds(2) - a.Perception)` | engine (who hates whom); duration is rules (names Perception, 217) | `Ruleset.HatredRounds(attacker) int` at 217 |
| 222-225, 326-373 | offhand break on hit; crit doubles chance | rules (chance, `items.go:100-112`); item swap and messages engine | `Ruleset.OnHit` returns "break offhand" or keep `BreakTest` as item data |
| 227-229 | vitals event | engine | |
| 231-246 | attacked NPC with no aggro: `PreventIdle`, walks toward a remote attacker, angry command | engine (behaviour) | |
| 248-254 | either side `Health <= 0`: end both aggros; else re-`SetAggro` (resets wait rounds, drops `ExitName`) | engine; threshold rules | `Ruleset.Condition` |
| 38, 477-520 | `handleAffected`: account `Health <= -10` queues `suicide` (495-497), `< 1` queues `PlayerDrop` (499-501); NPC `< 1` queues `suicide` (513-517) | engine dispatch; thresholds rules | `Ruleset.Condition(actor) {Up, Downed, Dead}` |

### 2c. Flee (`NewRound_DoCombat.go:389-473`)

- 394: revert to `DefaultAttack` (engine).
- 397-418: every body in the room targeting the fleer may block. Chance: `Speed_a / (Speed_a + Speed_b) * 70 + 30` (407-408), `util.Rand(100)` (410). Rules. Seam: `Ruleset.FleeEscapes(fleer, blocker) bool` replaces 406-414.
- 420-441: blocked or no exit resumes `FleeFrom` (`resumeAfterFlee` 377-384). Engine (ruled 2026-10-28).
- 435: `GetRandomExit` (engine, map order).
- 453-472: move, charmed NPCs `flee` too, look for accounts. Engine.

### 2d. `combat.Attack` internals (`combat.go:29-50`, `calculateCombat` 188-345): all rules except messaging

| Lines | Step | RNG draws |
|---|---|---|
| 192, `calculations.go:158-171` | attack count: weaponless or claws get `1 + extra(Speed delta) + StatMod("attacks")` | none |
| 195, `calculations.go:85-94` | damage bonus: `StatMod("damage") + Strength delta` | none |
| 201-215, `calculations.go:271-282` | dual wield: active weapon count by `dual-wield` skill; trim at random | `Rand(100)` at level 2; `Rand(len)` per trim |
| 219-224 | backstab: free crit once, aggro back to default | none |
| 229-232, `calculations.go:262-267` | offhand hit penalty -35/-25 | none |
| 235-253 | dice: race default (`character.go:357-380`, Speed/Strength/Perception scaled) or weapon; racial bonus statmod | none |
| 259-262 | message seed: `ConsistentAttackMessages` ? `ItemId` : 0 | |
| 273, `calculations.go:112-129` | hit: Speed proportional, clamped `[ToHitMin, ToHitMax]` | `Rand(100)` |
| 275-281, `calculations.go:241-258` | dodge: Perception delta | `Rand(100)` |
| 283 | damage `RollDice(dCount, dSides) + dBonus` | `dCount` draws |
| 287-294, `calculations.go:176-235` | crit: Smarts proportional, `accuracy`/`blink` flags, Perception multiplier | `Rand(100)` |
| 297, `calculations.go:286-294` | defence: `Rand(GetDefense())` percent off (`character.go:567-580`) | `Rand(n)` only if `n >= 1` (`util.go:187-190`) |
| 299-307, `items/attack_messages.go:40-52` | message pick by damage percent | `Rand(len)` when seed is 0 |
| 309-335 | crit and backstab prefixes, "[you blocked N]" | none |
| 36 | `ApplyHealthChange(-dmg)` (`character.go:1219-1238`: from above 0 a hit stops at -5; otherwise floor -10) | engine apply; clamp is rules |
| 39-41 | `TrackPlayerDamage` credit every swing (ruled) | engine |
| 43-47 | `WimpyCheck` for an account defender (`userrecord.go:618-625`: health percent below the account's `wimpy` option queues `flee`) | engine (account option) |

`GetWaitMessages` (`combat.go:97-126`) is messaging only, but `items.GetPreAttackMessage(...).Get(seed)` draws RNG when the seed is 0.

### 2e. `AutoHeal` (`NewRound_AutoHeal.go:17-89`), after combat

| Lines | Step | Class | Seam |
|---|---|---|---|
| 22 | every 3rd round | rules (regen period) | `Ruleset.Regen` |
| 28-30 | **accounts only** | engine filter; NPCs never regenerate (one-actor ruling issue) | iterate `actors.Each` |
| 33-35 | skip anyone with an aggro | rules (no regen in combat) | |
| 37-39 | skip `DeathRecoveryRoom` (75). The player is in an ephemeral copy (`roommanager.go:302-308`, ids from `ephemeralRoomIdMinimum`, `ephemeral.go:20,153`), so this compare never matches the copy. Reasoned, not tested; consistent with the Phase 4 note that the revived player "regenerates normally" | engine | dies with the stand-in |
| 44-48 | `Health <= -10` queues `suicide` (the only live player death from combat) | rules threshold | `Ruleset.Condition` |
| 50-64 | bleed 1 hp, "bleeding out" lines | rules amount; text engine | `Ruleset.Regen` returns -1 |
| 66-73 | `Heal(HealthPerRound(), ManaPerRound())` | rules (`character.go:1348-1367`: `1 + StatMod(HealthRecovery/ManaRecovery)`) | `Ruleset.Regen(actor) (hp, mana int)` |

Bug: `Character.Heal` adds `hp` to mana (`character.go:1338`, `c.Mana += hp`); the mana argument is ignored. Today both amounts are `1 + statmod`, so it shows only when a mana-recovery statmod differs from health recovery.

### 2f. `ActionPoints` (`NewTurn_ActionPoints.go:10-23`)

Every actor +1 AP per turn, capped at `ActionPointsMax` (engine loop; amount is rules). Max: `ActionPointsMax.Mods = 200` "hard coded for now" then `Recalculate(level)`, floor 50 (`character.go:1439-1455`). Spent only by `go`: cost 10, or 50 when items exceed `CarryCapacity()` (`go.go:51-58`; `character.go:233-235` is `5 + Strength/3`). Nothing in combat spends AP (grep `DeductActionPoints`: only `go.go:58`). Seams: `Ruleset.ActionPointsPerTurn(actor)`, `Ruleset.MoveCost(actor) int` at `go.go:51-56`.

---

## 3. Death

### Every path that ends in a death

All deaths run the `suicide` command (`usercommands.go:186-188`, `AllowedWhenDowned: true`), which branches at `suicide.go:29-31`: an NPC instance goes to `npcDeath` (`suicide_npc.go:26`), an account to the player body.

| Trigger | Who | Where |
|---|---|---|
| Bleed-out reaches -10 | accounts | `NewRound_AutoHeal.go:46-48` |
| Combat round, `<= -10` | accounts | `NewRound_DoCombat.go:495-497`, unreachable (section 6) |
| Combat round, `< 1` | NPCs | `NewRound_DoCombat.go:513-517` |
| Upkeep, `<= 0` | NPCs | `NewRound_ActorUpkeep.go:125-128` (catches non-combat damage, such as the lost-NPC drain `behaviour.go:66-72`, 10% of max per `pathto home` when home is impossible) |
| Buff with `TriggerNow`, NPC `<= 0` | NPCs | `Buff_ApplyBuffs.go:63-69` (queues `events.Input` "suicide") |
| Typed `suicide` | players (and any actor) | registered as a player command, allowed when downed |
| Admin kill | none | no admin or web path writes health (grep `\.Health.*=` over `usercommands`, `web`, `modules`: only `suicide.go:304`, `suicide_npc.go:33`) |

A downed player (`Health < 1`) can't act (`usercommands.go:281,314-317`); `PlayerDrop` prints "drops to the ground" (`PlayerDrop_HandlePlayerDrop.go:38-45`). No `aid` command exists (grep of the registry for `` `aid` ``, `` `heal` ``: none; the same grep found `` `attack` `` at 132), yet the bleed line says "Somebody needs to provide aid!" (`AutoHeal.go:59`). So every downed player dies.

### Player death (`suicide.go:27-313`)

| Lines | What | Class |
|---|---|---|
| 41-44 | refuse if zone is the literal `Shadow Realm` | engine (hard-coded zone name) |
| 47-53 | room line "has died" | engine |
| 56-62 | KD: PvP death if any `PlayerDamage`, else mob death | engine stats; a pet kill counts PvP (credit goes to the charmer) |
| 64-86 | KD per killer account, "killed by" list | engine |
| 93-95 | world notice | engine |
| 97 | penalties only above `Death.ProtectionLevels` | rules |
| 99-109 | elite death from `KilledBy`/`KillerMob*`, then clear | engine |
| 111-119 | `PlayerDeath` event (`Permanent`, `KilledByUsers`, `KillerMobId`) | engine |
| 122-170 | permadeath: spend `ExtraLives`, else unequip, drop all, reroll a new body (`characters.New`, new ID, `actors.BodyReplaced`), move to start (-1) | rules decide; mechanics engine. Unverified live (design:300-302) |
| 177-286 | penalties unless `perma-gear` flag: equipment drop at `EquipmentDropChance`, gold, backpack, XP loss `none`/`level`/percent, corpse (with items when `CorpseItems`) | rules (chances, XP loss); corpse and item moves engine |
| 288-298 | protected: corpse only | engine |
| 300 | cancel all buffs | engine |
| 302-306 | **stand-in**: `Health = 1`, `Mana = 0` | rules (revive values) |
| 308 | clear `PlayerDamage` | engine |
| 310 | move to `DeathRecoveryRoom`, which becomes a private ephemeral copy (`roommanager.go:302-308`) with a fixed `portal` to room 1 (Phase 4 doc :27-29) | engine |

### NPC death (`suicide_npc.go:26-443`)

| Lines | What | Class |
|---|---|---|
| 31-46 | `revive-on-death` flag: full health, cancel the flag | rules |
| 51-55 | remove charm | engine |
| 58-64 | "has died" | engine |
| 66, 83-95 | XP pool: `XPTL(level-1) / 90`, split by number of credited accounts, variation `(Rand(3)-1) * max(1, pool/100)` | rules |
| 73-81 | `MobDeath` event | engine |
| 104-112 | clear each credited account's aggro on the NPC | engine |
| 117-124, 242-247 | KD mob kill and elite kill, skipped in literal zone `Training` | engine (hard-coded zone name) |
| 126-152 | solo XP: level scaler when level gap > 5 (clamp 0.25-1.5); elite bonus `EliteXPBonus` (0 or less falls back to 10, 143-146) | rules |
| 154-172, 251-269 | alignment `AlignmentChange` then `UpdateAlignment` (resisted, RNG, `character.go:1695-1720`) | rules |
| 174-204, 271-302 | tame mastery: `Rand(1000) < levelDelta + skillsDelta` | rules (M/O handoff: tame is a skill) |
| 209-212, 220-308 | party XP: pooled per leader, split evenly over all members (any room), elite bonus, **no level scaler** | rules |
| 312-418 | unless `perma-gear`: carried items always drop, worn items at `ItemDropChance`, gold, to floor or corpse; `NPCEnded`, `DestroyInstance`, cleanup, corpse | rules (drop chances); mechanics engine |
| 420-442 | `perma-gear`: no loot, corpse | engine |

### Where NPC and player death diverge (one-actor ruling)

- Separate code paths entirely (`suicide.go:29-31`); the spec ruled this a stand-in until "the death rule" (design:342-343).
- Only accounts earn XP, KD, alignment, tame mastery; an NPC killed by an uncharmed NPC gives nothing (`suicide_npc.go:92-104` loops `PlayerDamage`, filled only through `CreditAccount`, `combat.go:39-41`).
- A player killer of a player gets no XP and no alignment change (nothing in `suicide.go` grants either).
- Players bleed; NPCs die at `< 1` (DoCombat) or `<= 0` (upkeep). No downed state for NPCs.
- Players regenerate (AutoHeal); NPCs never do.
- Player permadeath, lives, XP loss, drop chances are config-driven; NPC drops use `ItemDropChance` on the mob.

---

## 4. XP, levels, alignment, gold and loot on kill

| What | Computed | Stored | Read by |
|---|---|---|---|
| Kill XP | `suicide_npc.go:66,83-152,209-249` | `Character.Experience` via `Account.GrantXP` (`userrecord.go:176`) then `Character.GrantXP` (`character.go:438-466`: `XPScale` config, `statmods.XPScale`) | `GainExperience` event: `gmcp.Char.go:49`; prompt and info screens (grep: `userrecord.prompt.go`, `online.go`, `party.go`) |
| XP to level | `XPTL` (`character.go:1266-1277`): `(XPBase + L^XPLevelPower * XPLevelFactor * XPBase) * TNLScale` | | `LevelUp`, `XPTNLActual`, NPC XP pool |
| Level up | `Character.LevelUp` (`character.go:1295-1329`): +level, training and stat points per config, `Validate`, full health and mana | `Level`, `TrainingPoints`, `StatPoints` | `LevelUp` event: `gmcp.Char.go:46`, `gmcp.Game.go:32` |
| Lives | `userrecord.go:221-223,247-249` (`LivesOnLevelUp` when permadeath, cap `LivesMax`); start `LivesStart` (`userrecord.go:71`, `start.go:182`) | `ExtraLives` | `suicide.go:116,124-126` |
| Death XP loss | `suicide.go:241-273` | `Experience` | |
| Alignment on kill | `combat.AlignmentChange` (`calculations.go:378-434`) then `UpdateAlignment` | `Alignment` (int8, neutral band -19..19, `alignment.go:13-15`) | `AlignmentChanged` event: `gmcp.Char.go:55` |
| Alignment decay | `ActorUpkeep.go:137-154`, `DecayAlignment` (`character.go:1728-1745`) | `Alignment` | same |
| Kill credit | `TrackPlayerDamage` (`character.go:160-173`; forgets after 30 rounds idle) | `PlayerDamage map[int]int` (account UserId keys) | `suicide.go:56-86`, `suicide_npc.go:68-104`, `MobDeath`/`PlayerDeath` listeners `follow.go:68-69`, `gmcp.Char.go:57-58` |
| Killer of a player | `NewRound_DoCombat.go:196-207` | `KilledBy`, `KillerMobIsElite/Name/Id` (transient, `character.go:93`), cleared on heal (`character.go:1012-1023`) | `suicide.go:99-109` |
| NPC gold and items | `suicide_npc.go:312-418` | room floor (`room.Gold`, `AddItem`) or corpse | `MobItemDrop` event (no listener found in the grep of `RegisterListener(events.MobItemDrop`) |
| Player gold and items on death | `suicide.go:177-286` | floor or corpse | |
| Elite level | `mobs.go:149` (`EliteLevelBonus`) | NPC `Level` | |

Bug (GMCP-visible only): `Account.GrantXP` snapshots `tpBefore`/`spBefore` at 196-197, calls `LevelUp` at 201, then inside the loop re-snapshots (235-236) before the next `LevelUp` (238) and adds that delta (240-241). The first level's training and stat points never enter `LevelUp.TrainingPoints`/`StatPoints`; a single level-up reports 0. The character's real points are right.

---

## 5. Commands that start and end a fight

Registry `usercommands.go`: `attack` (132), `consider` (139, allowed downed), `flee` (144), `shoot` (181), `suicide` (188, allowed downed). No `kill`, `aid`, `backstab`, `cast` (grep).

| Command | Checks | Rules among them |
|---|---|---|
| `attack` (`attack.go:31-289`) | no arg: whoever fights you or your party (40-111); `*`/`*mob`/`*user` random (113-160, `util.Rand`); by name (163); not self (166-171); "attack the darkness" (173-176); NPC charmed by you is "your friend" (191-194); party leader's auto-attack members join (196-214, 252-265); **player target only**: `room.CanPvp` (239-242), party member refused (244-251). Sets aggro (216, 268) with `AggroChanged` (218, 270). No range (room only), no AP, no level gap, no downed-target check | `CanPvp`'s level minimum (`rooms.go:2343-2345`) |
| `shoot` (`shoot.go:37-140`) | needs a `Shooting` weapon (39-42); `shoot <target> <exit>` (51-57); exit exists and unlocked (61-76); target in the adjacent room (70-72); charmed friend (97-100); party member (118-125). `SetAggroRemote` (102, 127). **No `CanPvp`** (contrast `attack.go:239`), **no `AggroChanged`** (contrast `attack.go:218`), no AP | weapon subtype gate is item data |
| `flee` (`flee.go:11-27`) | sets `Aggro{Type: Flee, FleeFrom}` and `AggroChanged`; resolved in the round (2c). Not allowed when downed | chance in the round |
| `consider` (`consider.go:42-111`) | target in room; `CombatOdds` (`calculations.go:546-572`, `expectedDPS` 437-542) mapped to labels (85-101). No nil check on `u`/`m` (71-72, 76-77) | all of the odds |
| `suicide` | death (section 3) | |
| `go`, `quit` | refuse while `inFight` (`go.go:25`, `quit.go:16`; `go.go:373-390`): own aggro or someone in the room targeting you | engine (ruled: no escape except flee) |
| `go` arriving | hostile NPCs only print "notices you as you enter!" (`go.go:309-321`); nothing attacks | NPC aggression is Phase 5 |

`CanPvp` (`rooms.go:2327-2356`): refuses in room -1 and `DeathRecoveryRoom` (compared by raw ID, so not in the ephemeral copy; reasoned), `PVP.Enabled` disabled/limited, `PVP.MinimumLevel` on both sides. Only caller: `attack.go:239` (grep `CanPvp`).

NPCs never start fights: every `SetAggro` caller is a command, the round's re-target, flee resume, or `simulate.go` (grep of `SetAggro|.Aggro = &` over non-test Go). World data angry commands are emotes and shouts only (`races/0-ghostly_spirit.yaml:7`, `1-human.yaml:10-13`, `10-rodent.yaml:8`); no `combatcommands` in `world/empty` (grep).

---

## 6. Inherited bugs from the spec notes, checked against today's code

| Note (design line) | Status | Evidence | Phase 5 |
|---|---|---|---|
| Shadowed `userId` in the flee check (296-297) | **Fixed in slice 5** | `flee` loops bodies with `b.Character.Aggro.Target != a.Id` (`NewRound_DoCombat.go:397-404`); plan decision 2 | drop the note |
| `affectedPlayerIds` gets 0 in a mob fight (297) | **Fixed** | `affected []ids.ID` gets `a.Id` (113) and `d.Id` (154) | drop |
| `user` dereferenced before nil check (297-298) | **Fixed** | nil check at 82 precedes `HasBuffFlag` at 87 | drop |
| Spell fizzle clears aggro without an event (298) | **Gone** | `grep -i fizzle` over Go: no hit (same grep matched other terms, so it could match) | drop |
| `KilledBy` zero for a PvP kill (298-299) | **Fixed** | set for any attacker when the defender has an account (196-207); `TestKilledByOutlivesTheKiller` passes | drop |
| `-10` suicide threshold in the round unreachable (330-331) | **Still there** | `handleAffected` 495-497. A hit from above 0 stops at -5 (`character.go:1226-1227`); a target below 1 is never hit (148-152). Only AutoHeal reaches -10 | goes naturally when `Ruleset.Condition` replaces both thresholds |
| Turn order accounts first (329) | **Still there** | 49-62 | removed by `Ruleset.TurnOrder` |
| NPCs never start or join fights (330) | **Still there** | section 5 | Phase 5 must add NPC aggression (hostile flag, group hatred on arrival, assist) or rule it out of scope |
| Pet's group-hatred duration uses the pet's Perception (331-332) | **Still there** | `MinutesToRounds(2) - a.Character.Stats.Perception.ValueAdj` (217), `a` is the pet | removed when `HatredRounds` gets the credited account's body or a rule |
| Pet that downs a player is the killer (332) | **Still there** | 197-206 records the pet; `PlayerDamage` credits the charmer | rules call: who is "the killer" |
| NPC death via `suicide_npc.go` (342-343) | **Still there** | `suicide.go:29-31` | removed by one death path |
| Alignment decay accounts-only (316) | **Still there** | `ActorUpkeep.go:84-86` | removed by `RoundUpkeep` for all actors |

New ones found in this pass (not in the spec):

1. `shoot` at a player skips `CanPvp` (`shoot.go:112-135` vs `attack.go:239`). PvP through an exit ignores PvP settings. Player-visible.
2. `shoot` fires no `AggroChanged` (`shoot.go:102,127`); slice-5 decision 7 says every aggro change does. GMCP aggro flags lag one round.
3. `AutoHeal` and `CanPvp` compare the raw `DeathRecoveryRoom` ID, but the player is in an ephemeral copy (`roommanager.go:302-308`). Reasoned, not tested.
4. `Character.Heal` uses `hp` for mana (`character.go:1338`).
5. NPCs never regenerate (`AutoHeal.go:28-30`).
6. `LevelUp` event undercounts points (section 4).
7. Party XP has no level scaler and reaches members in any room (`suicide_npc.go:209-249`).
8. `attack` with no argument dereferences `mobs.GetInstance(mId)` without a nil check (`attack.go:48-49`); `consider` likewise (71-77).
9. Dead knobs and code: `PricePerLife` (no consumer outside `configs`, grep), `ChanceToTame` (no caller), `GetRoomWithMostItems` (no caller, `roommanager.go:441`).

---

## 7. Rules config knobs and shipped values

All from `_datafiles/config.yaml` (only config file under `_datafiles`; `find -iname "*config*"`); validation defaults from `config.gameplay.go`.

| Knob | Shipped | yaml line | Consumer | Default if invalid |
|---|---|---|---|---|
| `GamePlay.Death.EquipmentDropChance` | 0.25 | 232 | `suicide.go:182-188,226-229` | 0.0 (`:99-101`) |
| `Death.AlwaysDropBackpack` | false | 236 | `suicide.go:216` | |
| `Death.XPPenalty` | none | 242 | `suicide.go:243-271` | none (`:103-116`) |
| `Death.ProtectionLevels` | 5 | 245 | `suicide.go:97` | 0 |
| `Death.PermaDeath` | false | 248 | `suicide.go:116,122`; `userrecord.go:221` | |
| `Death.CorpsesEnabled` | true | 251 | `suicide.go`, `suicide_npc.go` | |
| `Death.CorpseDecayTime` | 1 hour | 255 | `rooms.go:250` | 1 hour |
| `Death.CorpseItems` | false | 261 | same | |
| `LivesStart` / `LivesMax` / `LivesOnLevelUp` | 3 / 3 / 1 | 264, 267, 270 | `userrecord.go:71,221-249`, `start.go:182` | |
| `PricePerLife` | 1000000 | 273 | none (dead) | 1 |
| `Combat.ConsistentAttackMessages` | true | 307 | `combat.go:103,260` | |
| `Combat.DamageBonusMin/Max` | 0 / 10 | 309-310 | `calculations.go:85-94` | 0 / 10 |
| `Combat.ToHitMin/Max` | 25 / 100 | 312-313 | `calculations.go:99-129` | 25 / 100 |
| `Combat.ExtraAttacksMin/Max` | 0 / 3 | 315-316 | `calculations.go:134-143` | 0 / 3 |
| `Combat.CritChanceMin/Max` | 5 / 30 | 318-319 | `calculations.go:176-197` | 5 / 30 |
| `Combat.CritMultMin/Max` | 1.5 / 3.0 | 321-322 | `calculations.go:215-224` | 1.5 / 3.0 |
| `Combat.DodgeChanceMin/Max` | 5 / 30 | 324-325 | `calculations.go:241-250` | 5 / 30 |
| `PVP.Enabled` | limited | 333 | `rooms.go:2335-2352` | enabled unless `off` |
| `PVP.MinimumLevel` | 15 | 336 | `rooms.go:2343` | 0 |
| `XPScale` | 100 | 340 | `character.go:444` | 100 |
| `AlignmentDecayRounds` | 100 | 344 | `ActorUpkeep.go:138` | 100 |
| `EliteLevelBonus` | 20 | 348 | `mobs.go:149` | 20 |
| `EliteXPBonus` | 10 | 352 | `suicide_npc.go:143,229` | 10 |
| `Progression.XPBase/XPLevelFactor/XPLevelPower` | 1000 / 0.75 / 2.0 | 409, 412, 416 | `character.go:1266-1277` | (M) |
| `Progression.TrainingPoints*/StatPoints*` | 1, 1, 1, 1 | 401-404 | `character.go:1303-1309` | (M) |
| `Progression.HP*/Mana*` | 5, 1.0, 4.0 / 4, 1.0, 3.0 | 390-396 | `RecalculateStats` (M) | |
| `SpecialRooms.DeathRecoveryRoom` | 75 | 581 | `suicide.go:310`, `roommanager.go:302`, `AutoHeal.go:26`, `rooms.go:2329`, `save_and_load.go:437` | |
| `SpecialRooms.StartRoom` | 1 | 576 | `roommanager.go:320-322` | |
| `LootGoblin.RoomId` | 0 | 116 | `admin.mob.go:132` (`mob loot goblin`, skipped when 0); `roommanager.go:443` (no caller) | |
| `LootGoblin.RoundCount/MinimumItems/MinimumGold/IncludeRecentRooms` | 900 / 3 / 100 / true | 121, 124, 127, 132 | none live (the spawn hook was stripped, plan :101) | 10 / 2 / 100 |

Hard-coded rules values (not config): regen every 3 rounds, bleed 1 (`AutoHeal.go:22,51`); -5 and -10 floors (`character.go:1226-1230`); flee 30 + 70 (`NewRound_DoCombat.go:407-408`); hatred 2 minutes (217); AP +1 per turn (`NewTurn_ActionPoints.go:16`), AP max mods 200 floor 50 (`character.go:1439,1453-1455`); move cost 10/50 (`go.go:51-55`); dual wield -35/-25 (`calculations.go:262-267`); revive 1 health, 0 mana (`suicide.go:304-305`); `PlayerDamage` forget after 30 rounds (`character.go:166`); NPC XP `/90` (`suicide_npc.go:83`); zone names `Shadow Realm` (`suicide.go:41`) and `Training` (`suicide_npc.go:119,242`).

---

## 8. Parity tools

### `util.SeedRandForTest`

`util.go:173-185`: one package-level `*rand.Rand` behind `randMu`, clock-seeded; `SeedRandForTest(seed)` replaces it. `Rand(n)` returns 0 **without drawing** when `n < 1` (187-190). `RollDice` draws once per die (543-556). Other `math/rand` users (`events.go:181`, `gametime`, `procedural/gridmaze.go`) do not share it.

### Combat golden (`internal/hooks/combat_golden_test.go`)

Each fixture builds two rooms (93001, 93002), seeds, calls `DoCombat` then `events.ProcessEvents` per round until no aggro (366-388), and records per round: what each player read (markup, via `readsRecorder`), the events `AggroChanged`, `PlayerDrop`, `CharacterVitalsChanged`, `ItemOwnership`, `Buff`, `MobDeath`, `PlayerDeath` and queued inputs (104-140), then every actor's room, hp, aggro (target, type, wait, exit), `PlayerDamage`, killer, weapon, offhand, items, group hatred, room floor (296-341). Queued commands are recorded, **not run**: death itself (XP, loot, revive) is not in this golden.

| Fixture | Seed | Rounds in golden (`grep -c "== round"`) |
|---|---|---|
| `pvm_death` | 111 | 4 |
| `pvp` (shield break chance 90) | 202 | 4 |
| `mvp` | 303 | 6 |
| `mvm_charmed` | 404 | 3 |
| `flee_blocked` | 505 | 3 |
| `flee_success` | 606 | 1 |
| `shoot_exit` | 707 | 2 |
| `observer` (wait 1 weapon) | 808 | 6 |
| `hidden_attacker` | 909 | 6 |

Plus `TestShootWaitLineDoesNotReachTargetRoom` and `TestKilledByOutlivesTheKiller` (asserts, no file). Stats set through `Training` and `Validate`, health max set directly (188-195); no config loaded by the test (no `configs.` call in the file, grep), so it runs with validated defaults (for example `ConsistentAttackMessages` false, which makes every message pick draw RNG). Not verified by printing the config inside the test.

### Reads goldens (`reads_golden_test.go`)

Section E (716-766) runs `consider rat`, `attack rat`, one `DoCombat`, `flee`, one `DoCombat` (seed 1105). Section H (831-908, seed 1108) runs NPC death via `suicide` (tame and alignment lines), alignment decay at round 100, `GrantXP 5000` (level up), and player death with the recovery room set to a fixture room. These are the only tests that pin death, XP and alignment text.

### Smoke fight (`cmd/smoke/fight.go`)

Real server, clock seed. Bot 0 sends `attack rat` until it reads "rat has died." and "You gained … experience points … (combat)" (85-99), retrying when the rat wandered off; bot 1 must read a swing line naming the fighter and the rat and no attacker-only line (139-158). Death step: `suicide`, "has DIED!", recovery room by title, `portal` out (`script.go:221-228`). `cmd/smoke/testdata/pre-5/bot{0,1}.txt` (431, 440 lines) are hand-compared transcripts from `-transcript` (`main.go:44`); no Go code reads them (grep `pre-5` over Go: none; phase-4b-7 inventory :583, :770). They are CRLF on this machine (4b-7 inventory :1050).

### What must stay byte-identical when the math moves

1. The 9 combat goldens, unchanged, with no `-update`. That pins the **RNG draw sequence**, not just the formulas. Per round, per actor: NPC combat-command roll (`DoCombat:273`), look-for-weapon roll (305), dual-wield draws, then per swing: hit, dodge (only on a hit), dice, crit, defence (only if defence ≥ 1), message pick (only when the seed is 0), then offhand break (`items.go:105`), angry command pick (`mobs.go:466,473`). Flee: one roll per blocker. Any reorder, extra draw, or a ruleset that picks messages separately from the math changes the record.
2. Reads goldens E and H (consider odds text, death, XP, alignment, tame lines).
3. `internal/combat` unit tests moved with the code (`calculations_test.go`, `simulate_test.go`, `weapon_rank_test.go`).
4. Smoke strings: "rat has died.", "You gained", "(combat)", "has DIED!", "You prepare to fight", "You attack the darkness!", "Your target can't be found.".

The low-risk route: the `diku-basic` `ResolveAttack` is today's `calculateCombat` moved verbatim, returning messages as now, so the draw order is untouched.

---

## 9. Proposed `Ruleset` shape (combat, round, death half). PROPOSAL, not built

Grounded in the call sites above. One actor type (`*actors.Actor`); the ruleset never sends text, it returns lines (the existing `TestCombatPackageDoesNotSend` principle).

```go
// package rules (engine side). PROPOSAL.
type Condition int
const ( Up Condition = iota; Downed; Dead )

type Ruleset interface {
    // Round
    TurnOrder(fighters []*actors.Actor) []*actors.Actor             // DoCombat:33 (replaces fightersWithAccountsFirst 49-62)
    AttackDelay(a *actors.Actor) int                                // character.go:1113 (SetAggro wait rounds)
    ResolveAttack(a, d *actors.Actor) AttackResult                  // DoCombat:182 (today calculateCombat, combat.go:188-345)
    HatredRounds(attacker *actors.Actor) int                        // DoCombat:217
    BreaksOnHit(d *actors.Actor, crit bool) bool                    // DoCombat:222-225, 333-340
    FleeEscapes(fleer, blocker *actors.Actor) bool                  // DoCombat:406-414
    Condition(a *actors.Actor) Condition                            // DoCombat:148,248,495-516; ActorUpkeep:125; AutoHeal:44-46; Buff_ApplyBuffs:63
    ClampDamage(a *actors.Actor, delta int) int                     // character.go:1219-1238 (-5/-10 floors)
    Consider(a, d *actors.Actor) float64                            // consider.go:72,77 (CombatOdds)
    CanAttack(a, d *actors.Actor, room *rooms.Room) error           // attack.go:239 and the missing shoot.go check; CanPvp's level gate

    // Upkeep and regen
    RoundUpkeep(a *actors.Actor, round uint64) []perception.Message // ActorUpkeep:84-86 (alignment decay, all actors)
    Regen(a *actors.Actor, round uint64) (hp, mana int)             // AutoHeal:22,33,50-73 (bleed is a negative hp)
    ActionPointsPerTurn(a *actors.Actor) int                        // NewTurn_ActionPoints:16
    MoveCost(a *actors.Actor) int                                   // go.go:51-55

    // Death and progression
    OnDeath(victim *actors.Actor, credited []*accounts.Account, killer *actors.Actor) DeathOutcome // suicide.go:97-306, suicide_npc.go:31-418
    XPForLevel(c *characters.Character, level int) int              // character.go:1266 (M owns the stat side)
}

type DeathOutcome struct {
    Revive         bool           // revive-on-death, or a player's revive
    ReviveHealth   int            // stand-in: 1
    ReviveMana     int            // stand-in: 0
    Permanent      bool           // permadeath reroll
    XPLoss         int
    DropWorn       []items.Item   // engine moves them (floor or corpse)
    DropCarried    []items.Item
    DropGold       int
    Awards         []Award        // per account: XP, alignment delta, tame mastery
    Lines          []perception.Message
    ToRecoveryRoom bool
}
```

Engine keeps: the round loop and listener order, target lookup and range, aggro set and clear with events, flee movement, hidden-target skip, messaging, credit tracking, `KilledBy`, wimpy, corpse creation, item and gold moves, the recovery-room move, reroll mechanics, KD bookkeeping, events. The ruleset is set once at boot (`rules.Set(diku.New())`) and read through `rules.Get()`; tests swap it with `t.Cleanup` restore.

Calls that `DoCombat` makes become, in order: `TurnOrder` → (per actor) `Condition(d)` → `ResolveAttack` → apply via `ClampDamage` → `HatredRounds` → `BreaksOnHit` → `Condition(a|d)` → re-aggro with `AttackDelay`; then `handleAffected` reads `Condition` only.

### "A stub ruleset changes the outcome with no core edit": test sketch

```go
// internal/hooks/ruleset_stub_test.go. PROPOSAL.
type oneShot struct{ rules.Ruleset }              // embed diku-basic, override one method
func (oneShot) ResolveAttack(a, d *actors.Actor) rules.AttackResult {
    return rules.AttackResult{Hit: true, DamageToTarget: 1000,
        Messages: []perception.Message{{Kind: perception.WorldEvent, Actor: a.Id, Target: d.Id,
            Senses: perception.Sight, Range: perception.Room, ActorText: "STUB HIT"}}}
}
func TestStubRulesetChangesTheFight(t *testing.T) {
    w := newCombatWorld(t, 111)                   // same fixture as pvm_death
    alice := w.addPlayer(9301, "Alice", goldenRoomA, 5, 30, 20, 10)
    rat := w.addNPC(93201, "rat", 10, goldenRoomA, 10, 5, 5, 12, "stub-rats")
    alice.Character.SetAggro(rat.Character.Id, characters.DefaultAttack)
    t.Cleanup(rules.SetForTest(oneShot{rules.Get()}))
    w.start(); w.rounds(20)
    require.Contains(t, w.out.String(), "== round 1")
    require.NotContains(t, w.out.String(), "== round 2")       // golden pvm_death takes 4
    require.Contains(t, w.out.String(), `Input rat: "suicide"`)
    require.Contains(t, w.reads[alice.UserId], "STUB HIT")
}
```

A second stub on `Regen` (always +5) against `AutoHeal`, and one on `OnDeath` (fixed 7 XP) against reads section H's NPC death, cover the other two halves. The proof of "no core edit" is that the test file is the only diff.

---

## 10. Open questions

### (i) Player-visible: need Calabe

1. **NPC aggression.** Today no NPC starts or joins a fight; `hostile: true` NPCs only print "notices you". Does Phase 5's "a fight runs through the ruleset" include NPCs attacking (hostile on arrival, group hatred, assisting allies), or stay player-initiated?
2. **Death and recovery.** Replace the stand-in (revive at 1 health, portal out of 75) with what: GoMud's -10 plus recovery buff, revive at a fraction of max, or something else? Keep the private recovery room?
3. **Downed state with no aid.** Every downed player dies (no `aid` exists) after bleeding from up to -5 to -10, one point per 3 rounds. Keep downed-and-bleed, or die at 0?
4. **One death path for NPCs and players.** Should NPCs get a downed state, regenerate, and earn XP/alignment for kills (and players for PvP kills), per the one-actor ruling? Today all of that is accounts-only.
5. **Who is the killer when a pet downs a player:** the pet (today) or its charmer?
6. **`shoot` ignores PvP settings.** Fix by routing both `attack` and `shoot` through one `CanAttack` (my pick), or treat as a Phase 5 non-goal?
7. **Turn order:** keep accounts-first in `diku-basic` for golden parity, or move to an initiative rule now (changes close fights)?
8. **Party XP:** no level scaler, and members in other rooms share. Keep upstream's rule?

### (ii) Engineering defaults (my pick in bold)

1. Ruleset location: **`internal/rules` (interface, `Get`/`Set`) plus `internal/rules/dikubasic`**, with `internal/combat` moved into `dikubasic` and the rank/simulate tools moved with it (web imports `dikubasic` for rankings).
2. Granularity of attack resolution: **one `ResolveAttack` returning lines, moved verbatim**, so the RNG draw order and all 9 combat goldens stay byte-identical. Split it finer only after parity holds.
3. Thresholds: **one `Condition(actor)`** replacing the five scattered `< 1`/`<= 0`/`<= -10` checks; delete `handleAffected`'s unreachable -10 branch in the same change (golden unaffected because it never fires).
4. Death: **one engine death routine** calling `OnDeath`, replacing `suicide.go`/`suicide_npc.go` branching; the `suicide` command stays as a player command that calls it.
5. Fix in passing (engine bugs, no rules change): `shoot` `AggroChanged`, `Heal` mana argument, `LevelUp` event point counts, nil checks in `attack`/`consider`, recovery-room comparisons via `GetOriginalRoom`. **Each in its own commit with its golden diff named.** (`Heal` changes mana regen only when the two statmods differ; nothing in goldens should move. Unverified.)
6. Dead code: **delete** `ChanceToTame`, `GetRoomWithMostItems`, `PricePerLife`, and the `LootGoblin` config with the admin `mob loot goblin` branch (status.md:12 hands Phase 5 the loot goblin config; nothing spawns a goblin).
7. Hard-coded zone names `Shadow Realm` and `Training`: **replace** with "already in the recovery room" (via `GetOriginalRoom`) and drop the Training exemption (no such zone in `world/empty`; not grepped, verify before deleting).
8. Guard updates: **edit `internal/guards/combat_test.go` controls** to name the new files in the same commit; O's stat-name guard covers `lookForWeapon`'s Perception roll and `HatredRounds`, so those must move.
9. Golden for death: **add a combat golden fixture that runs the queued `suicide`** (today recorded, not run) before moving death, so death math is pinned by seed, not only by reads section H.

Handoffs: M owns `Validate`/`RecalculateStats`, HP/mana/AP max formulas, `GetDefaultDiceRoll`, `GetDefense`, `XPTL` inputs, starting values. O owns `no-combat`, `cancel-on-combat`, `hidden`, `revive-on-death`, `perma-gear`, crit buffs (`BuffTarget`), buff 24 (death recovery), the stat-name guard, and GMCP `Char` readers of XP/level/alignment.


# Inventory O: buffs, effects and surfaces (Phase 5)

Read 92 of 92 files (54 full, 38 partial; listed below). Files touched only by grep are listed separately and are not counted as read; no claim below rests on unread lines of them beyond the grep line quoted.

Base: `master` at `9131eac2` (`git rev-parse --short HEAD`). Read-only: no repo file changed. Module `github.com/GoMudEngine/GoMud`.

## Files read

| File | Lines read | Full/partial |
|---|---|---|
| internal/buffs/buffspec.go | 1-180 | full |
| internal/buffs/buffs.go | 1-351 | full |
| internal/buffs/flags.go | 1-117 | full |
| internal/buffs/flags_admin.go | 1-113 | full |
| internal/buffs/flags_plugin.go | 1-83 | full |
| internal/buffs/plugin.go | 1-83 | full |
| internal/buffs/admin.go | 1-62 | full |
| internal/buffs/memory.go | 1-16 | full |
| internal/buffs/testhelpers.go | 1-15 | full |
| internal/buffs/CLAUDE.md | 1 | full |
| internal/hooks/Buff_ApplyBuffs.go | 1-81 | full |
| internal/hooks/NewTurn_PruneBuffs.go | 1-78 | full |
| internal/messaging/perceive.go | 1-66 | full |
| internal/perception/perception.go | 1-117 | full |
| internal/guards/doc.go | 1-3 | full |
| internal/guards/guards_test.go | 1-275 | full |
| internal/guards/combat_test.go | 1-103 | full |
| internal/guards/tick_test.go | 1-83 | full |
| internal/guards/commands_test.go | 1-210 | full |
| internal/guards/output_test.go | 1-343 | full |
| internal/characters/charminfo.go | 1-25 | full |
| internal/usercommands/quit.go | 1-23 | full |
| internal/web/api_v1_buffs.go | 1-103 | full |
| internal/web/api_v1_statmods.go | 1-15 | full |
| internal/spells/summon.go | 1-8 | full |
| status.md | 1-29 | full |
| docs/completed/2026-10-26-phase-4-scripting-surface.md | 1-79 | full |
| _datafiles/world/empty/templates/help/stats.template | 1-33 | full |
| _datafiles/world/empty/buffs/{0-meditating,1-illumination,24-death_recovery}.yaml | all (3 files) | full |
| _datafiles/world/empty/buffs-flags/*.yaml | all (17 files) | full |
| _datafiles/world/empty/mutators/{death_recovery,pvp_enabled,wildfire}.yaml | all (3 files) | full |
| `57255ee4^:_datafiles/world/empty/buffs/{0-meditating,1-illumination,24-death_recovery}.js` (deleted, via `git show`) | all (3 files) | full |
| gomud_strip_plan.md | 60-140 | partial |
| 2026-10-27-phase-4b-npcs-are-players-design.md | 262-369 | partial |
| internal/characters/character.go | 620-660, 970-985, 1135-1230, 1370-1385, 2050-2115 | partial |
| internal/hooks/NewRound_ActorUpkeep.go | 60-120 | partial |
| internal/hooks/NewRound_DoCombat.go | 80-100, 140-190 | partial |
| internal/events/eventtypes.go | 1-60 | partial |
| internal/accounts/userrecord.go | 315-340 | partial |
| internal/accounts/users.go | 335-350 | partial |
| internal/accounts/userrecord.prompt.go | 236-250 | partial |
| internal/mobs/mobs.go | 235-255 | partial |
| internal/mobs/saved.go | 1-80 (of 93) | partial |
| modules/gmcp/gmcp.Char.go | 40-200, 420-670, 775-945 | partial |
| internal/usercommands/help.go | 140-207 | partial |
| internal/usercommands/set.go | 75-90, 290-335 | partial |
| internal/usercommands/usercommands.go | 55-70, 280-290 | partial |
| internal/usercommands/suicide.go | 170-182, 295-302 | partial |
| internal/usercommands/suicide_npc.go | 25-58, 305-318 | partial |
| internal/usercommands/look.go | 20-45 | partial |
| internal/usercommands/go.go | 25-45 | partial |
| internal/usercommands/admin.reload.go | 25-55 | partial |
| internal/usercommands/equip.go | 100-150 | partial |
| internal/spells/spells.go | 1-80 (of 179) | partial |
| _datafiles/world/empty/keywords.yaml | 1-130 (of 152) | partial |
| _datafiles/world/empty/templates/help/set.md | 25-36 | partial |
| _datafiles/world/empty/panel-layouts/character/inventory.yaml | 1-30 | partial |
| main.go | 222-245, 1440-1530 | partial |
| internal/mutators/mutators.go | 55-75 | partial |
| internal/rooms/rooms.go | 164-184, 380-452, 2000-2025 | partial |
| internal/rooms/roommanager.go | 350-385 | partial |
| internal/rooms/roomdetails.go | 360-380, 448-460 | partial |
| internal/templates/templatesfunctions.go | 100-165 | partial |
| internal/combat/calculations.go | 195-215 | partial |
| internal/combat/simulate.go | 178-190 | partial |
| internal/items/items.go | 255-265 | partial |
| internal/items/itemspec.go | 368-415 | partial |
| internal/messaging/wire.go | 60-95 | partial |
| internal/web/api_v1_mobs.go | 170-185 | partial |
| internal/characters/shop.go | 15-25 | partial |

Grep only (not read, not counted): `_datafiles/html/admin/*.html` (which pages name buffs/stats/skills), `_datafiles/html/public/static/js/windows/window-character.js` (lines 8-25, 596, 631-632, 665, 755-828 seen in grep output), `internal/web/api_routes.go`, `internal/web/admin_items.go`, `internal/web/admin_routes.go`, `internal/web/admin_nav.go`, `internal/web/api_v1_permissions.go`, `internal/stats/stats.go` (lines 9-27 in grep), `internal/statmods/statmods.go` (31-36), `internal/configs/config.lootgoblin.go`, `internal/buffs/*_test.go` (test names), the test files named in section 2.

---

## 1. `internal/buffs`

### Files

Non-test: `buffspec.go` 180, `buffs.go` 351, `flags.go` 117, `flags_admin.go` 113, `flags_plugin.go` 83, `plugin.go` 83, `admin.go` 62, `memory.go` 16, `testhelpers.go` 15 (1,020 lines). Tests: `buffs_test.go` 477, `buffspec_test.go` 444, `flags_test.go` 91, `plugin_test.go` 133 (27 `func Test`). Notes: `AGENTS.md` 28, `CLAUDE.md` 1 (`@AGENTS.md`). (`wc -l`.)

### Exported symbols

- Spec (`buffspec.go`): `BuffSpec` (28-39), `(*BuffSpec).GetValue` (42), `VisibleNameDesc` (60), `Id` (119), `Validate` (124), `Filename` (154), `Filepath` (159); `BuffMessage`/`BuffMessages` (67-76); `GetBuffSpec` (78, negates a negative id), `GetAllBuffIds` (90), `SearchBuffs` (101), `LoadDataFiles` (164).
- Instance (`buffs.go`): consts `TriggersLeftExpired = 0`, `TriggersLeftUnlimited = 1000000000` (8-9); `Buff` (12-21); `(*Buff).StatMod` (23), `Expired` (33), `Name` (90); `Buffs` (38-42, unexported maps `buffFlags`, `buffIds`); `New` (44); `(*Buffs).Validate` (52), `StatMod` (82), `RemoveBuff` (97), `TriggersLeft` (105), `GetBuffIdsWithFlag` (112), `HasFlag` (123), `HasBuff` (168), `Started` (175), `AddBuff` (181), `Trigger` (227), `GetBuffs` (267), `Prune` (288); `GetDurations` (330).
- Flags (`flags.go`): `All = ""` (22), `FlagSpec` (29-34), `GetFlagSpec` (65), `GetAllFlagSpecs` (70), `GetAllFlagSpecsSorted` (80), `IsValidFlag` (93), `LoadFlagDataFiles` (103). `flags_admin.go`: `ValidateFlagId` (17), `CreateFlagSpec` (37), `SaveFlagSpec` (62), `DeleteFlagSpec` (92). `flags_plugin.go`: `RegisterFlagFS` (18).
- Plugins (`plugin.go`): `RegisterFS` (18). Admin (`admin.go`): `GetAllBuffSpecs` (13), `SaveBuffSpec` (23), `DeleteBuffSpec` (47). `memory.go`: `GetMemoryUsage` (5). `testhelpers.go`: `SetTestSpec` (5).

### Buff spec fields (`buffspec.go:28-39`)

`BuffId`, `Name`, `Description`, `Secret`, `TriggerNow`, `TriggerRate` (time string), `RoundInterval` (derived in `Validate` from `TriggerRate`, `:135`), `TriggerCount`, `StatMods statmods.StatMods`, `Flags []string`. No script or effect field remains (Phase 4.2 removed `GetScript`/`GetScriptPath`, `git show 57255ee4 -- internal/buffs/buffspec.go`).

Instance fields (`buffs.go:12-21`): `BuffId`, `Source`, `OnStartWaiting`, `PermaBuff`, `RoundCounter`, `TriggersLeft`, `TriggersInitial`.

### Lifecycle

- **Load:** `LoadDataFiles` reads `<DataFiles>/buffs` with `fileloader.LoadAllFlatFiles`, panics on error, merges plugin buffs (disk wins) (`buffspec.go:164-180`; `plugin.go:26-83`). `LoadFlagDataFiles` does the same for `buffs-flags` (`flags.go:103-117`). Boot order: flags then buffs then items (`main.go:1509-1511`). Plugin FS registered at `main.go:183-184`. In-game `reload buffs-flags` exists; there is no `reload buffs` (`admin.reload.go:38-40`, cases items, biomes, buffs-flags, skills, translations).
- **Validate spec:** buff 0's `TriggerCount` is forced to `LogoutRounds` (`buffspec.go:127-129`); `RoundInterval < 1` or `TriggerCount < 1` rejects; unknown flags warn only (`:144-149`).
- **Apply:** `events.Buff{ActorId, BuffId, Source}` (`eventtypes.go:21-27`) is queued by `Account.AddBuff` (`userrecord.go:324-332`), `Mob.AddBuff` (`mobs.go:241-249`), `MoveToRoom` mutators (`roommanager.go:364-378`) and the combat round's crit buffs (`NewRound_DoCombat.go:184-186`). Listener `hooks.ApplyBuffs` (registered `hooks.go:12`) resolves the body (account else mob, `Buff_ApplyBuffs.go:33-49`), a negative id removes (`:51-54`), else `Character.AddBuff(id, false)` (`:57`). `TriggerNow` only queues `suicide` for a mob at 0 health (`:62-72`); no effect runs. Then `BuffsTriggered` always, `BuffsAdded` only for accounts (`:74-78`). `events.Buff.Source` is never read (grep `evt.Source` over `internal` and `modules` matched only GMCP Comm lines).
- **Stacking:** `Buffs.AddBuff` re-applying an existing id refreshes `TriggersLeft`, `TriggersInitial`, `PermaBuff` and keeps the counter; no stacks (`buffs.go:203-208`). Permanent: `TriggersLeft = TriggersLeftUnlimited` (`:197-201`). `Character.AddBuff` takes `abs(id)` and calls `Validate()` (`character.go:1158-1165`).
- **Permabuffs:** `reapplyPermabuffs` ref-counts template (`permaBuffIds`), race `BuffIds`, and worn items' `WornBuffIds`, then adds or removes (`character.go:2062-2112`); called from `Validate(true)` (`:1641-1642`) and equip/unequip paths (`:1895`, `:2032`, `:2045`). Template list set by `SetPermaBuffs` (`:2055`), from `mobs.go:182`, `saved.go:55`, `simulate.go:70`.
- **Tick:** once per round for every actor in `ActorUpkeep`: `char.Buffs.Trigger()` then queue `BuffsTriggered` (`NewRound_ActorUpkeep.go:92-98`). `Trigger` increments `RoundCounter`, fires when `RoundCounter % RoundInterval == 0`, decrements `TriggersLeft` unless unlimited (`buffs.go:227-265`). The `buffId...` filter in `Trigger` has no effect: `continue` only continues the inner loop (`:233-239`); no caller passes an id (grep `Buffs.Trigger(` gives only `ActorUpkeep.go:92`).
- **Expire:** per **turn**, `hooks.PruneBuffs` (registered `hooks.go:37`). Players: only rooms with players, `GetPlayers(FindBuffed)`, `Prune`, `Validate`, queue `CharacterVitalsChanged`, `BuffsTriggered`, `BuffsRemoved`; pruning buff 0 logs the user off unless zombie (`NewTurn_PruneBuffs.go:25-63`). Mobs: every mob, `Prune` then `Validate`, no events (`:66-74`). `Prune` drops expired buffs and buffs whose spec is gone, then rebuilds lookups (`buffs.go:288-328`).
- **Cancel:** `Character.CancelBuffsWithFlag(flag)` calls `Buffs.HasFlag(flag, true)`, which marks matches expired (buff 0 is spliced out at once) and then `Validate(true)` (`character.go:1146-1152`; `buffs.go:123-166`).
- **Save:** instances persist in the character YAML (`Buffs buffs.Buffs yaml:"buffs,omitempty"`, `character.go:72`); NPCs carry it inside `SavedNPC.Template` (`saved.go:15-22`), with template permabuffs re-set on restore (`saved.go:39-46`, `:55`, `:78`). Specs are saved by `SaveBuffSpec` (`admin.go:23-45`); flags by `CreateFlagSpec`/`SaveFlagSpec` (locked flags refuse edit and delete, `flags_admin.go:74-76`, `:100-102`).

### Engine vs rules

| Part | Where | Bucket |
|---|---|---|
| Spec load, plugin merge, admin save/delete, memory reporter | `buffspec.go:164`, `plugin.go`, `admin.go`, `memory.go` | engine |
| Timer: `TriggerRate` to `RoundInterval`, `Trigger`, `Prune`, `Expired`, `GetDurations` | `buffspec.go:135`, `buffs.go:227-350` | engine |
| Stacking/refresh, permanent instances, per-flag index | `buffs.go:181-224`, `:52-80` | engine |
| Permabuff ref counting from template, race, worn item | `character.go:2062-2112` | engine (the race list itself is rules, M) |
| Persistence in character and NPC saves | `character.go:72`, `saved.go` | engine |
| Flag registry and cancel-by-flag | `flags*.go`, `buffs.go:123` | engine (the flag set is mixed, section 3) |
| `StatMods` field and `StatMod` summing | `buffspec.go:37`, `buffs.go:23-31`, `:82-88`, `character.go:1378-1380` | rules (stat names) |
| `GetValue` (item price weight) | `buffspec.go:42-58` | rules (economy) |
| Buff 0 = logout countdown | `buffspec.go:127`, `NewTurn_PruneBuffs.go:42-46`, `quit.go:20` | engine (quit flow) using a buff id |
| Effect behaviour (onStart/onTrigger/onEnd) | none now (section 4) | rules, missing |

---

## 2. Importers and calls

`go list -f '{{.ImportPath}} {{join .Imports " "}}' ./... | grep internal/buffs` gives 11 importing packages (the `internal/buffs` line itself matched on its own path): root `github.com/GoMudEngine/GoMud`, `characters`, `combat`, `hooks`, `items`, `mobs`, `mutators`, `templates`, `usercommands`, `web`, `modules/gmcp`. 17 non-test files. Test-only imports (`TestImports`/`XTestImports`): `hooks`, `mobs`, `usercommands` (files: `hooks/combat_golden_test.go:26`, `hooks/reads_golden_test.go:47`, `mobs/saved_test.go:7`, `usercommands/combat_text_test.go:8`, `usercommands/items_test.go:8`; `combat/simulate_test.go:8` calls `buffs.LoadDataFiles` in a package that already imports it).

Classes: **E** engine timed-affect use, **F** reads a flag (rules or engine flag, section 3), **R** pure rules, **S** outward surface, **D** dead.

| Package | File:line | Call | Class |
|---|---|---|---|
| root | `main.go:183-184` | `RegisterFS`, `RegisterFlagFS` | E |
| root | `main.go:1509-1510` | `LoadFlagDataFiles`, `LoadDataFiles` | E |
| characters | `character.go:72`, `:131` | `Buffs` field, `buffs.New()` | E |
| characters | `character.go:1142-1178` | `HasBuffFlag`, `CancelBuffsWithFlag`, `HasBuff`, `AddBuff`, `TrackBuffStarted`, `GetBuffs`, `RemoveBuff` wrappers | E |
| characters | `character.go:1379` | `Buffs.StatMod` in `Character.StatMod` | R |
| characters | `character.go:634-644` | adjectives from `lightsource`, `hidden`, `poison` | F |
| characters | `character.go:1223` | `cancel-on-combat` when health goes below 0 | F |
| characters | `character.go:1547`, `:2062-2112` | `Buffs.Validate`, permabuff ref count | E |
| characters | `character.go:1167-1169` | `TrackBuffStarted` (no non-test caller; grep) | D |
| combat | `armor_rank.go:78`, `:89` | `GetBuffSpec`, `GetValue` for armor rank | R |
| combat | `simulate.go:71`, `:186` | `buffs.New`, `AddBuff` crit buffs in simulator | R |
| hooks | `Buff_ApplyBuffs.go:26`, `:52`, `:57` | apply/remove | E |
| items | `itemspec.go:380`, `:402`, `:409` | `GetBuffSpec`/`GetValue` in `GetBuffSummary` and `AutoCalculateValue` | R |
| mobs | `mobs.go:184` | `buffs.New()` on spawn | E |
| mutators | `admin.go:29` | `ValidateBuffIds` on mutator save | E |
| templates | `templatesfunctions.go:118-142`, `:153` | `buffname`, `buffduration`, `roundstotime` | S |
| usercommands | `admin.reload.go:40` | `reload buffs-flags` | E (admin) |
| usercommands | `admin.room.go:774-775`, `:1430-1431` | trap buff picker for container and exit locks | D (section 5) |
| usercommands | `suicide.go:300` | `CancelBuffsWithFlag(buffs.All)` on player death | R (death rule) |
| web | `api_v1_buffs.go:13-103` | GET list, GET, PATCH, DELETE specs | S |
| web | `api_v1_buffs_flags.go:13-94` | flag CRUD | S |
| web | `api_v1_mobs.go:179` | shop entry `BuffId` validation | D (no `buy`) |
| web | `api_v1_yaml.go:75` | buff YAML view | S |
| modules/gmcp | `gmcp.Char.go:587`, `:595` | `Char.Affects` builder | S |

Callers through `Character` methods in packages that do not import `buffs` (grep of `HasBuffFlag(|CancelBuffsWithFlag(|.AddBuff(|.RemoveBuff(|.HasBuff(|.GetBuffs(|Buffs.(Trigger|Prune|List|StatMod|Validate)|hasFlag(` outside `internal/buffs`, non-test, definitions included): usercommands 32, rooms 25, characters 15, hooks 9, messaging 5, combat 5, modules/gmcp 3, accounts 2, modules/cleanup 1 (97). The flag readers are itemised in section 3. Other non-flag uses:

| File:line | Use | Class |
|---|---|---|
| `NewRound_ActorUpkeep.go:92-98` | `Buffs.Trigger` for every actor | E |
| `NewTurn_PruneBuffs.go:37`, `:70` | `Buffs.Prune`, split players/mobs | E (player/mob split) |
| `accounts/users.go:344` | `RemoveBuff(0)` on link-dead | E (quit flow) |
| `usercommands/quit.go:20` | `AddBuff(0, "quitting")` | E (quit flow) |
| `rooms/rooms.go:384-450` | `ApplyBuffIdToPlayers/Mobs/NativeMobs` | E, but split by kind |
| `rooms/rooms.go:2014-2019` | room tick applies mutator `PlayerBuffIds`/`MobBuffIds`/`NativeBuffIds` | E |
| `rooms/roommanager.go:364-378` | `MoveToRoom` applies `PlayerBuffIds` to any body | E |
| `rooms/rooms.go:1201`, `:1287` | `FindBuffed` (`len(Buffs.List) > 0`) | E |
| `rooms/roomdetails.go:373`, `:389`, `:397`, `:416`, `:432`, `:440` | `RemoveBuff(26)` in tracking display | D (section 5) |
| `combat/combat.go:291`, `attackresult.go:9` | crit sets `BuffTarget` from `CritBuffIds` | R |
| `NewRound_DoCombat.go:184-186` | queue crit buffs on defender | R |

Buff events: `Buff`, `BuffsTriggered`, `BuffsAdded`, `BuffsRemoved` (`eventtypes.go:21-48`). Listeners: `Buff` to `ApplyBuffs` (`hooks.go:12`); the other three only to GMCP (`gmcp.Char.go:52-54`), which refreshes `Char.Affects` or `Char` for accounts (`:101-176`). Nothing in the engine reacts to a trigger now.

---

## 3. Buff flags

Defined as YAML in `_datafiles/world/empty/buffs-flags/` (17 files, all `locked: true` except `tripping`). No flag is a Go constant except `All = ""` (`flags.go:22`). Readers by string literal (grep `["`]<flag>["`]` over non-test Go; counts include hits excluded from the listing as noted):

| Flag | Readers | Effect | Bucket |
|---|---|---|---|
| `hidden` | `character.go:638` (adjective), `accounts/userrecord.prompt.go:244` (`{h}` prompt), `NewRound_DoCombat.go:175` (can't target), `messaging/perceive.go:51` (`canSee` false), `rooms/roomdetails.go:257`, `:275` (hide from look unless `see-hidden`), `rooms/roommanager.go:432`, `rooms/rooms.go:704` (`Unseen` on enter/spawn), `usercommands/go.go:38`, `:350`, `look.go:37`, `read.go:32`, `modules/cleanup/cleanup.go:89` (sneaking quiet lines), `modules/gmcp/gmcp.Room.go:346`, `:374` (omit from room list); cancelled by `drop.go:62,91`, `equip.go:58`, `get.go:186,225,260,287,326,367`, `give.go:70,152`, `remove.go:65`, `show.go:48,82` | concealment | engine (perception) |
| `see-hidden` | `roomdetails.go:258`, `:276` | see hidden bodies in look | engine (perception); not honoured by `messaging.canSee` (`perceive.go:44-55`) |
| `nightvision` | `perceive.go:38`, `look.go:31`, `:239` | see in dark | engine (perception) |
| `lightsource` | `character.go:634` (adjective), `rooms/rooms.go:1165`, `:1257` (`FindHasLight`), used by `GetVisibility` at `rooms.go:203` | light in room | engine (perception) |
| `superhearing` | `perceive.go:19` (faint sound), `messaging/wire.go:79` (`IsQuiet` messages) | hear faint | engine (perception) |
| `see-nouns` | `roomdetails.go:127`, `look.go:338` | show room nouns | engine (presentation) |
| `no-go` | `go.go:31` | blocks movement | engine gate |
| `no-combat` | `NewRound_DoCombat.go:87` | skips combat turn | engine gate |
| `cancel-on-action` | `usercommands.go:286` | ends buff on any command | engine lifecycle |
| `cancel-on-combat` | `character.go:1223`, `NewRound_DoCombat.go:98`, `:146`, `combat/simulate.go:118-119` | ends buff on combat | engine lifecycle |
| `perma-gear` | `usercommands.go:64` (refuse unequip), `suicide.go:177` (no player death drops), `suicide_npc.go:312` (no NPC loot) | gear stuck, no death penalty | rules |
| `revive-on-death` | `suicide_npc.go:31`, `:43` (NPC only) | revive once | rules (player path has no check: one-actor gap) |
| `accuracy` | `combat/calculations.go:204` | crit chance | rules |
| `blink` | `combat/calculations.go:205` (`emote.go:18` is an emote word, not the flag) | crit dodge | rules |
| `poison` | `character.go:642` (adjective only) | label; damage was script | rules |
| `drunk` | `say.go:22`, `shout.go:21` | slurred speech | rules |
| `tripping` | `roomdetails.go:453-458` | warped room text | rules |

Not flags: `darkvision` 0 hits, `charm` 0 hits (grep of the same pattern; the pattern finds `hidden` 31 times, so it can match). Charm is a `CharmInfo` on the character (section 5). Shipped buffs use only `cancel-on-action`, `cancel-on-combat` (buff 0) and `lightsource` (buff 1) (`buffs/*.yaml`).

Risk: `Buffs.HasFlag` dereferences `GetBuffSpec(b.BuffId).Flags` without a nil check (`buffs.go:136-137`); for `All` there is no early return (`:125-132`), so a saved buff whose spec was deleted panics `CancelBuffsWithFlag(buffs.All)` (`suicide.go:300`) until the next turn's `Prune`.

---

## 4. What went dead in Phase 4

Phase 4.2 (`57255ee4`) deleted 3 buff scripts and 8 spell scripts in `empty` plus 7 sample spell scripts (`git log --diff-filter=D`), and removed `scripting.TryBuffScriptEvent` calls: `onStart` in `Buff_ApplyBuffs` (with `TrackBuffStarted`), `onTrigger` in `Buff_ApplyBuffs` (TriggerNow) and the old round ticks, `onEnd` in `NewTurn_PruneBuffs` for players and mobs (`git show 57255ee4 -- internal/hooks/...`).

| Buff | YAML | Lost JS behaviour (deleted file) |
|---|---|---|
| 0 Meditating | `buffs/0-meditating.yaml` | `onStart`: "You sit down and begin your meditation." plus room line; `onTrigger`: "You continue your meditation. *N rounds left*" plus room line. Quit is now silent. Logout on expiry still works (Go). |
| 1 Illumination | `buffs/1-illumination.yaml` | `onStart`/`onEnd`: glow start and fade lines to user and room. Light itself still works (flag). |
| 24 Death Recovery | `buffs/24-death_recovery.yaml` | `onTrigger`: heal 1d10 health and 1d10 mana each round with messages. Now no flags, no mods: does nothing (phase 4 doc lines 20-32). Applied by mutator `death-recovery` (`mutators/death_recovery.yaml:16`) in room 75. |

No shipped buff YAML names a script or effect field (`BuffSpec` has none; the three YAML files hold only id, name, description, rate, count, flags). Every spell lost its whole effect: `spells/*.yaml` 8 files remain with `SpellData` fields only (`spells.go:17-26`), and no `cast` command exists (Phase 2 stripped it; `help.go:168-179` is the only in-game reader).

---

## 5. Charm, companions, spells as effects

- **Charm state:** `Character.Charmed *CharmInfo` (`yaml:"-"`) and `CharmedMobs []ids.ID` (saved) (`character.go:69-70`); `CharmInfo{By, RoundsRemaining, ExpiredCommand}` with `CharmPermanent`, `CharmExpiredRest`, `CharmExpiredRevert` (`charminfo.go:5-25`). NPC saves carry `Charmed` (`saved.go:21`, `:32`, `:57`).
- **Who starts a charm:** nobody. `grep "\.Charm\("` finds only tests (`behaviour_test.go:139`, `character_test.go:31`, `charmexpiry_test.go:100`, `combat_golden_test.go:442`, `:474`); `TrackCharmed(..., true)` only in `accounts/charmed_test.go`. The `tame` skill and charm spells were stripped (strip plan section 1).
- **Plumbing still live:** countdown and NPC expiry with `ExpiredCommand` (`NewRound_ActorUpkeep.go:88-110`; player charm only counts down), follow on move (`go.go:275-305`), flee together (`NewRound_DoCombat.go:459-463`), kill credit to charmer (`combat/combat.go:61-67`), no attacking own charmed (`attack.go:191`, `shoot.go:97`), wander suppression (`behaviour/behaviour.go:25-26`, `tick.go:140-150`), party listing (`party.go:227-283`, `gmcp.Party.go:207-264`), teleport with party (`admin.teleport.go:150-152`), despawn expiry (`PlayerDespawn_HandleLeave.go:51-54`), NPC death cleanup (`suicide_npc.go:51-53`), `FindCharmed` (`rooms.go:49`, `:1176-1186`, `:1262-1272`), login pruning of gone companions (`accounts/users.go:246-261`, `main.go:230-231`), newbie guide special case (`users.go:356-359`), adjective `charmed` (`formattedname.go:26`).
- **Dead:** `GetMaxCharmedCreatures` (`character.go:974-977`, reads skill `tame`): no caller (grep finds only the definition).
- **Spells as effects:** `spells.Summon` returns `false, nil` (`summon.go:3-8`). `SpellBook` stays on characters (`character.go:68`, M's scope).
- **Other buff appliers with no path now:** `Item.AddWornBuff` (`items.go:257`, no caller); `ItemSpec.BuffIds` "on use" (`itemspec.go:198`) only feeds price (`:401-405`; no use/eat/drink command); `gamelock.Lock.TrapBuffIds` (`gamelock.go:16`) is set by `admin.room.go:873`, `:1529`, `rooms.go:465` and applied nowhere (grep `TrapBuffIds` finds no reader); `ShopItem.BuffId` (`shop.go:21`) validated by `api_v1_mobs.go:179`, applied nowhere; `RemoveBuff(26)` in `roomdetails.go:366-440` runs only if misc data `tracking-mob`/`tracking-user` is set, and nothing sets it (grep finds only the two reads at `:366`, `:410`).
- **Mutators:** `MutatorSpec.PlayerBuffIds`, `MobBuffIds`, `NativeBuffIds` (`mutators.go:64-66`). Room tick applies them by kind (`rooms.go:2014-2019`); `MoveToRoom` applies `PlayerBuffIds` to any body, NPCs included (`roommanager.go:364-378`, spec note at design doc line 343). Under one actor model the Player/Mob split is a kind-specific path. `wildfire.yaml:12-13` names buff 22 and `items/consumables-30000/30001-small_red_potion.yaml:8-9` names buff 5; neither spec ships (`buffs/` holds 0, 1, 24; no plugin ships `buffs/`: `find modules -path "*buffs*"` empty while `find modules -path "*gmcp*"` matches). `ApplyBuffs` silently cancels unknown ids (`Buff_ApplyBuffs.go:26-29`).

Handoffs: death and recovery (buff 24, `suicide.go` stand-in) to N; `CritBuffIds` and crit buffs to N; `SpellBook`, skills, races' `BuffIds` to M.

---

## 6. World and data files

`FilePaths.DataFiles` is `_datafiles/world/empty` (`config.yaml:182`). All rules folders are loaded at boot (`main.go:1500-1520`) and required (`main.go:1473-1490`), so none is unread at load time; "unused" below means nothing in play reads the data.

| Path | Files | Loader | Read in play by | State |
|---|---|---|---|---|
| `buffs/` | 3 | `buffs.LoadDataFiles` | apply/tick/expire, GMCP, admin | 0 and 1 work (flags only); 24 does nothing |
| `buffs-flags/` | 17 | `buffs.LoadFlagDataFiles` | `IsValidFlag` warnings, admin | descriptive only; code uses literals |
| `spells/` | 8 YAML | `spells.LoadSpellFiles` | `help spell` (broken, section 7), admin, `SpellBook` | no effect anywhere |
| `skills/` | 16 | `skills.LoadDataFiles` | GMCP `Char.Skills`, skill-level reads (M) | no skill command |
| `professions/` | 10 | `skills.LoadProfessionDataFiles` | GMCP `Char.Info.Class`, `Char.Jobs`, template `profession` | display only |
| `races/` | 3 | `races.LoadDataFiles` | stats, permabuffs (none ship `buffids`), `help races` | M |
| `mutators/` | 3 | `mutators.LoadDataFiles` | rooms | `death-recovery` (buff 24, dead), `wildfire` (buff 22 missing), `pvp-enabled` |
| `combat-messages/` | 8 | `items.LoadDataFiles` | combat | N |
| `items/` | 12 | `items.LoadDataFiles` | everywhere | `torn_gloves` has `statmods` (strength 1, perception -1); potion `buffids: [5]` missing |
| `templates/character/skillup.template` | 1 | templates | no Go reference (grep `skillup` finds none; `character/permadeath` and `character/visited` do resolve, so the grep form works) | unread |
| `config.yaml` `LootGoblin` | block at 112+ | `configs` | `roommanager.go:443`, `admin.mob.go:129`; `RoomId: 0` disables | owner N or a later slice |

No `help/spell*` template exists (`find _datafiles -iname "*spell*"` lists only `html/admin/spells*.html` and `world/empty/spells`).

---

## 7. Surfaces

### Web admin (pages from `admin_items.go`/`admin_routes.go`; API from `api_routes.go`)

| Surface | Route / file | Proposal |
|---|---|---|
| Buffs page, API page | `/admin/buffs`, `/admin/buffs-api` (`admin_routes.go:25-27`; `buffs.html`, `buffs-api.html`) | keep as engine "Affects" editor; stat mods and effect key fields come from the ruleset |
| Buff flags page | `/admin/buffs-flags` (`buffs-flags.html`) | keep; engine flags read-only, ruleset flags listed |
| Buffs API | `GET/PATCH/DELETE /admin/api/v1/buffs...` (`api_routes.go:48-52`; no POST) | keep (engine) |
| Buff flags API | `/admin/api/v1/buff-flags...` (`api_routes.go:55-60`) | keep (engine) |
| StatMods API | `GET /admin/api/v1/statmods` (`api_routes.go:28`, `api_v1_statmods.go:10-15`) | move behind ruleset (trait list) |
| Races, Skills, Professions, Spells pages and APIs | `api_routes.go:145-163`, `:180-185`; pages `races*.html`, `skills*.html`, `spells*.html` | move behind ruleset (M decides what survives) |
| Permissions | `buffs.write`, `skills.write`, `spells.write` (`api_v1_permissions.go:77`, `:89`, `:90`) | follow the pages |
| Items API buff summary | `api_v1_items.go:91` (`GetBuffSummary`) | keep count; value is rules |
| Mob shop `BuffId` | `api_v1_mobs.go:179` | delete with shop buffs |
| Pages that show character stats/buffs (grep only): `users.html` 98 hits, `rooms.html` 88, `mobs.html` 70, `items.html` 55, `progression.html` 13, `items-rank-*.html`, `mob-rankings.html` | | ruleset-driven fields; detail is M/N |

### GMCP (`modules/gmcp/gmcp.Char.go`)

| Package | Fields | Proposal |
|---|---|---|
| `Char.Affects` | name, description, duration_max/left (seconds), type (always `unknown`: `Buff.Source` is never set), affects = `StatMods` map (`:578-632`, `:918-925`) | keep as engine; fix source; mods map keyed by ruleset trait names |
| `Char.Stats` | six named stats plus `StatMod` per stat (`:528-547`, `:879-892`) | move behind ruleset (ruleset supplies the map) |
| `Char.Vitals` | hp, hp_max, sp, sp_max (`:550-561`) | keep hp as engine; sp/mana is rules |
| `Char.Info` | class (profession), race, alignment, level, skillpoints, trainingpoints (`:426-441`) | split: name/account/role engine, rest ruleset |
| `Char.Worth` | gold, bank, tnl, xp (`:564-575`) | rules (N for XP) |
| `Char.Skills`, `Char.Jobs` | (`:635-667`) | move behind ruleset or delete with skills (M) |
| `Char.Kills`, `Char.Enemies` | (`:670-731`, `:444-479`) | N |
| `Room` players/mobs omit `hidden` (`gmcp.Room.go:346`, `:374`) | | engine perception |
| `Party` lists charmed companions (`gmcp.Party.go:207-264`) | | engine |

Webclient `window-character.js` (grep only) hard-codes `Char.Stats` six keys (line 596 shows `mysticism`), Skills and Jobs tabs (631-632), Effects tab from `Char.Affects` (10, 18).

### Templates and help

| Item | Where | Proposal |
|---|---|---|
| Template funcs `buffname`, `buffduration`, `roundstotime`, `profession` | `templatesfunctions.go:118-151`, `:152-157` | keep buff funcs (engine); `profession` moves |
| `help races` shows race buffs | `templates/help/races.template:32` | ruleset help |
| Stat help: `stats`, `strength`, `speed`, `smarts`, `vitality`, `mysticism`, `perception`, `health`, `mana`, `armor`, `damage`, `encumbrance`, `alignment` templates; aliases `keywords.yaml:84-102`; `stats` listed under character (`:19`) | `templates/help/` | ruleset-provided help (diku-basic ships these) |
| `help spell <x>` | `help.go:168-179` | needs `help/spell` or `help/spells` template; neither exists, so the command renders a template error. Delete with spells, or ship a template with the ruleset |
| `set shopstyle` | `set.go:82-87` (shown in `set` overview), `:297-330`; help `set.md:35-36`; golden `reads-golden/G-character-info.txt:39` | written, never read for output (grep `shopstyle` finds no renderer; the shop `list` command was stripped in Phase 2). Delete |
| Webhelp module | `modules/webhelp` renders `keywords` topics (`webhelp.go:68-83`, grep) | keep; content follows the help files |

---

## 8. Guard test

### How guards are built today

Package `internal/guards` has no runtime code (`doc.go:1-3`). Every test walks the repo's non-test Go files via `goFiles(t)` (`guards_test.go:18-48`): skips `.git`, `_datafiles`, `.claude`; parses with `go/parser`; keys by slash path; fails if fewer than 100 files were scanned (positive control). Patterns used:

- Banned identifiers anywhere (`TestOneCombatPath`, `combat_test.go:12-30`; `TestOneActorTick`, `tick_test.go:12-33`), each with a positive control that the real symbols exist (`combat_test.go:32-54`, `tick_test.go:35-56`).
- Package-scoped bans with a scanned-count control (`TestCombatPackageDoesNotSend`, `combat_test.go:60-102`).
- Allowlists keyed by path, or `rule|path|Recv.Func`, with a reason string, and a stale-entry failure: every entry must still match (`guards_test.go:247-251`; `output_test.go:61-88`).
- Import-path resolution so aliases are seen and look-alike local names are not (`importNames`, `pkgSel`, `output_test.go:219-249`), plus a parsed in-memory control source (`TestOutputGuardResolvesImports`, `output_test.go:110-153`).

Representative shape (`tick_test.go:12-33`):

```go
for path, f := range goFiles(t) {
    ast.Inspect(f, func(n ast.Node) bool {
        if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
            t.Errorf("%s: uses %s; ...", path, id.Name)
        }
        return true
    })
}
```

### Proposed `TestCoreNamesNoRulesetStat` (proposal)

**Core** = every non-test Go file except: the ruleset tree (proposed `internal/rules/dikubasic/`), and, while Phase 5 is moving them, `internal/stats`, `internal/statmods`, `internal/races`, `internal/skills`, `internal/spells` as path-prefix allowlist entries that must each still match (so an entry fails once the package is gone, forcing its removal).

**Rules:**
1. **Import ban:** a core file may not import the ruleset package or `internal/stats`/`internal/statmods` (after the move). This is the structural rule; the token rules are the backstop.
2. **Selector rule:** an `*ast.SelectorExpr` whose `Sel.Name` is one of `Strength, Speed, Smarts, Vitality, Mysticism, Perception`, **unless** `pkgSel(sel, imp, ".../internal/perception")` resolves (the type `perception.Perception`).
3. **Literal rule:** a `token.STRING` literal whose unquoted value, lowercased, equals one of `strength, speed, smarts, vitality, mysticism, perception`. Exact match only, so prose such as help text that mentions a stat inside a longer string is not flagged.
4. Positive controls: an in-memory source like `output_test.go:111-131` with `c.Stats.Strength`, `` `smarts` ``, and `perception.Perception` (must fire twice, not three times); and a check that the ruleset package itself contains the six names (so the token list stays in sync with the ruleset).

**Grep that the tokens match today's code** (non-test, `grep -rnoE "[A-Za-z_]*\.(Strength|Speed|Smarts|Vitality|Mysticism|Perception)\b"` minus `perception.Perception`): 157 selector hits: `internal/characters` 73, `combat` 28, `usercommands` 26, `accounts` 12, `modules/gmcp` 6, `races` 6, `hooks` 6. The excluded `perception.Perception` hits are `actors/controller.go:18`, `:40`, `:65`, `:74` and `messaging/messaging.go:72`; without rule 2's exception they would be false positives. Exact string literals: 18 hits, `statmods/statmods.go:31-36`, `usercommands/equip.go:137-142`, `modules/gmcp/gmcp.Char.go:532-542`. Core hits outside the moving packages today: `hooks/NewRound_DoCombat.go:217`, `:307`, `:309`, `:407`; `accounts/userrecord.go:229-234`; `usercommands/equip.go:107-142`, `suicide_npc.go:179`, `:276`; `modules/gmcp/gmcp.Char.go:531-542`; plus `characters` and `combat` (M and N move those).

**False positives avoided:** perception package type via import resolution; struct field declarations are not selectors (the `Perception` field at `actors/controller.go:65` is not flagged by rule 2; only its type selector is, and that is excluded); case-sensitive selector match avoids `perception.Sight`; exact literal match avoids help prose (32 non-exact literals contain a stat word).

**Second tier (decide):** `Mana`/`ManaMax` (32 and 29 selector hits in 8 and 7 files), `TrainingPoints` (7 in 4), `StatPoints` (14 in 6). Mana is diku-basic but sits in `Char.Vitals` and the prompt; see question ii-6.

---

## 9. Proposed timed-affects API and ruleset hooks (proposal)

Grounded in: apply at `Buff_ApplyBuffs.go:57`, tick at `NewRound_ActorUpkeep.go:92`, expire at `NewTurn_PruneBuffs.go:37`, `:70`, cancel at `character.go:1146`, the three lost JS hooks `onStart`/`onTrigger`/`onEnd`, and the stat-mod read at `character.go:1379`.

```go
// package affects (engine; today internal/buffs, renamed in a mechanical slice)
type Spec struct {
    Id          int
    Name, Description string
    Secret      bool
    Rate        string         // today TriggerRate; RoundInterval derived
    Count       int            // today TriggerCount
    TriggerNow  bool
    Flags       []Flag         // engine flags + ruleset-registered flags
    Effect      string         // key of a Go effect the ruleset registered; "" = none
    Mods        map[string]int // opaque trait mods; keys validated by the ruleset (today StatMods)
}
type Instance struct {
    SpecId    int
    Source    string // finally stored (spell, item, area, combat, quitting)
    Permanent bool
    Counter, Left, Initial int
}
type EndReason uint8 // Expired, Cancelled (flag), Removed, Death, SpecGone

type Set struct{ /* today Buffs */ }
func (s *Set) Apply(id int, src string, permanent bool, count int) (*Instance, bool /*new*/)
func (s *Set) Remove(id int, why EndReason) *Instance
func (s *Set) Cancel(f Flag) []*Instance        // today HasFlag(f, true)
func (s *Set) Tick() []*Instance                 // today Trigger
func (s *Set) Expire() []*Instance               // today Prune
func (s *Set) Has(id int) bool
func (s *Set) HasFlag(f Flag) bool
func (s *Set) Mod(key string) int                // today StatMod
func Durations(i *Instance, sp *Spec) (left, total int)

// Engine flags (Go constants; engine reads them):
const (Hidden, SeeHidden, NightVision, LightSource, SuperHearing, SeeNouns,
       NoMove /*no-go*/, NoCombat, CancelOnAction, CancelOnCombat Flag = ...)

// Engine driver: one pass for every actor, in ActorUpkeep (replaces PruneBuffs' split).
func ApplyTo(a *actors.Actor, id int, src string)   // Set.Apply, OnApply, events
func Upkeep(a *actors.Actor)                        // Tick -> OnTick; Expire -> OnEnd
```

```go
// Ruleset side (package rules; diku-basic implements)
type Effect struct {
    OnApply func(a *actors.Actor, i *affects.Instance, sp *affects.Spec)
    OnTick  func(a *actors.Actor, i *affects.Instance, sp *affects.Spec)
    OnEnd   func(a *actors.Actor, i *affects.Instance, sp *affects.Spec, why affects.EndReason)
}
type Ruleset interface {
    // ... attack, damage, regen, death, starting values (sibling inventories)
    AffectEffects() map[string]Effect      // diku-basic: "meditate", "illumination", "death-recovery"
    AffectFlags() []affects.FlagSpec       // accuracy, blink, poison, drunk, tripping, perma-gear, revive-on-death
    TraitKeys() []string                   // validates Spec.Mods keys
    CanPerceive(r *actors.Actor, room *rooms.Room, m perception.Message, s perception.Sense) bool // today messaging.CanPerceive var
}
```

Notes on the shape: buff 0 (logout) stays an engine affect with a reserved id and an account-only `OnEnd` (logoff) through the controller, not a player loop. `BuffsAdded`/`BuffsRemoved` fire for every actor; GMCP already filters to accounts (`gmcp.Char.go:147-150`, `:167-170`). Rules flags (`accuracy`, `blink`) are read by diku-basic's combat math, which moves with N; `drunk` and `tripping` need render hooks the engine lacks (design doc line 288: no per-receiver render hook).

---

## 10. Open questions

### (i) Player-visible, for Calabe

1. **Death recovery:** rebuild buff 24 as diku-basic Go (1d10 health and mana per round in room 75, as the JS did), or drop it with the death stand-in when N rebuilds death?
2. **Lost buff text:** meditation (quit) start and per-round lines, illumination glow start and fade lines are gone; quit is silent. Restore them as Go text?
3. **Flags no content uses:** `poison` (label only, no damage), `drunk`, `tripping`, `accuracy`, `blink`, `see-nouns`. Keep them in diku-basic, or delete until content needs them?
4. **Spells:** nothing casts (no `cast`), `help spell` renders a template error, 8 spell files load. Delete spells from the reference ruleset now, or rebuild a minimal cast?
5. **Charm and companions:** nothing can start a charm, but follow, party, flee, persistence and expiry are live. Keep the plumbing for a later tame/charm, or delete it now?
6. **`set shopstyle`:** shown in `set` and its help but read by nothing. Delete?
7. **GMCP stats:** keep `Char.Stats` with six fixed diku-basic keys (Mudlet and the webclient window hard-code them), or make it a ruleset-supplied map (wire format changes)?
8. **Missing buff references:** `wildfire` mutator names buff 22, the small red potion names buff 5; neither ships. Remove the references or add the buffs?
9. **See-hidden in messaging:** look honours `see-hidden`, but messaging always says "Someone" for a hidden actor. Should `see-hidden` show the name in messages too?
10. **Revive-on-death for players:** only the NPC death path checks it. Under one actor model, should a player with the flag revive too?

### (ii) Engineering defaults (my pick)

1. Merge `PruneBuffs` into `ActorUpkeep` for every actor in ID order; expiry per round, not per turn. Pick: yes (removes the player/mob split and the rooms-with-players limit).
2. Mutator `PlayerBuffIds`/`MobBuffIds`/`NativeBuffIds`: pick `ActorBuffIds` plus `NativeBuffIds` (native is a spawn fact, not a kind), applied by one function at tick and on move.
3. Effects are named Go functions referenced from YAML by `effect:` key; unknown key warns at load like flags. Pick: yes.
4. Engine flags become Go constants; ruleset registers its own; YAML flag files stay as admin descriptions, validated against the registered set. Pick: yes.
5. `StatMods` becomes an opaque `mods` map validated by `Ruleset.TraitKeys()`. Pick: yes.
6. Mana in core: treat `Mana`/`ManaMax` as diku-basic and add them to the guard's second tier once N and M move regen and vitals. Pick: tier two, enabled at the end of Phase 5.
7. Delete dead code in the move: `OnStartWaiting`/`Started`/`TrackBuffStarted`, `BuffMessage(s)`, the `Trigger(buffId...)` parameter, `TriggersLeft`, `GetBuffIdsWithFlag`, `SearchBuffs` (test-only callers), `Item.AddWornBuff`, `TrapBuffIds`, `ShopItem.BuffId`, the `tracking-mob`/`tracking-user` block with `RemoveBuff(26)`, `GetMaxCharmedCreatures`. Pick: yes, each with a grep in the plan.
8. Store `events.Buff.Source` on the instance so `Char.Affects.type` stops reading `unknown`. Pick: yes.
9. Nil-check the spec in `Buffs.HasFlag` (`buffs.go:136`). Pick: yes.
10. Rename `buffs` to `affects`: pick a separate mechanical (Sonnet) slice after behaviour moves, so diffs stay readable.
11. Add `reload buffs` beside `reload buffs-flags`. Pick: yes, small.
12. `messaging.CanPerceive` becomes `Ruleset.CanPerceive` with today's body as the engine default. Pick: yes; senses keep `Sight|Sound` reach as the design doc requires until a ruling changes it.
