# Phase 4b slice 5 inventory (2026-10-28, master dc94a7f7)

Three read-only inventories written for the slice 5 plan (`2026-10-28-phase-4b-5-one-combat-path-plan.md`): D, the combat package; E, the combat round; F, combat output and text parity. Each lists the files it read. Their open questions were settled by Calabe or by the plan's defaults (plan, "Decisions"). This file moves to `docs/completed/` with the plan.

# Inventory D: `internal/combat` for Phase 4b slice 5 (one attack path)

Branch `strip/phase-4b-5-one-combat-path` at `58a0cbc0` (master `dc94a7f7` + docs commit). Read-only; no repo
file or git state changed. Line numbers are as of that commit.

## Read 24 of 24 files

Full (Read tool, chunks of at most 400 lines, or `cat`/`sed` of the whole file):

1. `2026-10-27-phase-4b-npcs-are-players-design.md` (330 lines, full)
2. `internal/combat/combat.go` (392, full)
3. `internal/combat/attackresult.go` (37, full)
4. `internal/combat/calculations.go` (572, full: 1-400, 400-572)
5. `internal/combat/simulate.go` (192, full)
6. `internal/combat/mob_rank.go` (236, full)
7. `internal/combat/weapon_rank.go` (220, full)
8. `internal/combat/armor_rank.go` (266, full)
9. `internal/combat/calculations_test.go` (436, full: 1-400, 400-436)
10. `internal/combat/simulate_test.go` (93, full)
11. `internal/combat/weapon_rank_test.go` (68, full)
12. `internal/items/attack_messages.go` (123, full)
13. `internal/messaging/messaging.go` (222, full)
14. `internal/messaging/render.go` (49, full)
15. `internal/messaging/perceive.go` (66, full)
16. `internal/perception/perception.go` (110, full)
17. `internal/actors/actor.go` (full, `cat`)
18. `internal/actors/controller.go` (full, `sed 1,120`; the file ends inside that range)
19. `internal/characters/formattedname.go` (124, full)

Partial:

20. `internal/hooks/NewRound_DoCombat.go` (947): lines 40-897 read (`handlePlayerCombat`, `handleMobCombat`); 1-39 and 898-947 (`DoCombat`, `handleAffected`) not read. The full handler diff is another inventory's job; here only the parts that consume `AttackResult`.
21. `internal/items/itemspec.go`: 150-219 (intensities, tokens, `Damage`, `ItemMessage`).
22. `internal/characters/character.go`: 155-175 (`TrackPlayerDamage`), 356-380 (`GetDefaultDiceRoll`), 566-688 (`GetDefense`, `GetMobName`, `GetPlayerName`, `GetAdjectives`, `getFormattedName`), 1070-1125 (`SetAggroRemote`, `SetAggro`, `EndAggro`, `HasBuffFlag`), 1201-1225 (`ApplyHealthChange`).
23. `internal/accounts/userrecord.go`: 334-341 (`SendText`), 617-650 (`WimpyCheck`).
24. `internal/util/util.go`: 170-185 (`Rand`, `LogRoll`), 527-540 (`RollDice` head).

Data, parsed rather than read line by line: the 8 files in `_datafiles/world/empty/combat-messages/` (`generic.yaml` 1-140 read; all 8 parsed with a read-only Python `yaml.safe_load` script for token and empty-slot tallies). Greps only (not read): `internal/usercommands/attack.go:170-200`, `shoot.go`, `rooms/rooms.go:262-275`, `actors/registry.go`/`wire.go` function lists, `_datafiles/config.yaml:300-330`, `cmd/smoke/testdata/pre-4b/bot0.txt`.

## Headline findings

1. **The four variants differ only in health application and damage tracking.** All four call the same `calculateCombat(source, target Character, sourceType, targetType)` (`combat.go:31,49,67,82`). Every roll, hit, crit, dodge, damage, and message selection happens there, identically. The `SourceTarget` args affect only name rendering (`combat.go:160-175`) and the `{sourcetype}`/`{targettype}` colour token.
2. **`DamageToSource` and `BuffSource` are never set** by anything (grep: only reads at `combat.go:33,51,69,84`, `simulate.go:134,137,155,158`, `NewRound_DoCombat.go:269,337,444,493,685,814`). So the attacker-side `ApplyHealthChange` and attacker `WimpyCheck` are dead in all four variants.
3. **`hidden` is checked only in `GetWaitMessages`** (`combat.go:124`), not in `calculateCombat`. A hidden attacker's real attack lines name it to the target and the room today. Only the wait-round lines are suppressed.
4. **NPC attackers never use the "separate" (different rooms) messages.** `handleMobCombat` drops aggro unless attacker and defender share a room (`NewRound_DoCombat.go:607,771`). Only a player can attack across rooms, and only via `shoot` (`SetAggroRemote`, `usercommands/shoot.go:86,116`), which needs a `Shooting` weapon (`shoot.go:21`). `empty` has one weapon, `10001-sharp_stick.yaml`, `subtype: bludgeoning`. So ranged combat cannot happen in the test world.
5. **The smoke parity transcript has no combat in it.** `cmd/smoke/testdata/pre-4b/bot0.txt` has no attack, damage or miss line. Its only rat line is `197: Also here: rat, rat and rat`. So the slice 4b text-parity check cannot catch a combat text change. Combat parity needs its own test.
6. **No test exercises `calculateCombat`, `GetWaitMessages`, `buildCombatMessages`, or any `Attack*` variant** (grep of `*_test.go` repo-wide).
7. **Pre-existing crit cascade bug** (`combat.go:320-323`). After the first crit in a round, `attackResult.Crit` stays true, so every later hit that round is a crit without rolling `Crits` (`||` short-circuits). The comment at 318-319 says the opposite. This changes both outcomes and RNG consumption, so parity has to keep it unless a ruling says otherwise.
8. **The spec's facts-table call lines are stale.** It gives `NewRound_DoCombat.go` lines 348, 517, 808, 938 (as of `c4314b7f`). Today the calls are at 265 (PvP), 442 (PvM), 681 (MvP), 812 (MvM).
9. **`messaging.Send` has no non-test caller yet** (grep: 0). Slice 5 is its first production use.

## 1. Exported API of `internal/combat`, with non-test callers

Files: `armor_rank.go`, `attackresult.go`, `calculations.go`, `combat.go`, `mob_rank.go`, `simulate.go`, `weapon_rank.go` (+ 3 tests, + upstream `AGENTS.md`/`CLAUDE.md`).
Importers (grep of the import path): `hooks/NewRound_DoCombat.go:11`, `usercommands/consider.go:8`, `usercommands/suicide_npc.go:9`, `web/api_v1_items.go:8`, `web/api_v1_mobs.go:10`. The grep for `combat\.[A-Z]` outside the package found hits in all five, so it was able to match.

| Symbol | File:line | What it does | Non-test callers |
|---|---|---|---|
| `type SourceTarget string`, `User`, `Mob` | `combat.go:21-26` | Pairing kind: picks the name renderer and the `{sourcetype}`/`{targettype}` colour (`user`+`name`, `mob`+`name`) | `NewRound_DoCombat.go:232,414,662,792` (args to `GetWaitMessages`) |
| `AttackPlayerVsMob(*accounts.Account, *mobs.Mob)` | `combat.go:29` | see section 2 | `NewRound_DoCombat.go:442` |
| `AttackPlayerVsPlayer(*Account, *Account)` | `combat.go:47` | see section 2 | `NewRound_DoCombat.go:265` |
| `AttackMobVsPlayer(*mobs.Mob, *Account)` | `combat.go:65` | see section 2 | `NewRound_DoCombat.go:681` |
| `AttackMobVsMob(*mobs.Mob, *mobs.Mob)` | `combat.go:80` | see section 2 | `NewRound_DoCombat.go:812` |
| `GetWaitMessages(items.Intensity, src, tgt *Character, srcType, tgtType)` | `combat.go:96` | Builds wait-round lines (no damage). Hides target and room lines when the source is `hidden` | `NewRound_DoCombat.go:232,414,662,792` (all pass `items.Wait`) |
| `type AttackResult` + `SendToSource/Target/SourceRoom/TargetRoom/RoomOld` | `attackresult.go:3-37` | see section 3 | through the variants and `GetWaitMessages` |
| `Hits(atkSpd, defSpd, mod) bool` | `calculations.go:112` | Hit roll, `Rand(100) < clamp(hitChance+mod)` | none outside the package (used `combat.go:308`) |
| `Crits(src, tgt Character) bool` | `calculations.go:200` | Crit roll; `accuracy` flag doubles, target `blink` halves | none outside (used `combat.go:320`) |
| `Dodges(defPerc, atkPerc) bool` | `calculations.go:253` | Dodge roll | none outside (used `combat.go:310`) |
| `ChanceToTame(*Account, *mobs.Mob) int` | `calculations.go:365` | Taming odds | **none** (dead) |
| `AlignmentChange(killer, killed int8) int` | `calculations.go:378` | Alignment delta on a kill | `usercommands/suicide_npc.go:144,240` |
| `CombatOdds(atk, def Character) float64` | `calculations.go:546` | Ratio of rounds-to-kill from `expectedDPS` | `usercommands/consider.go:45,50` |
| `type SimResult`, `(SimResult) String()` | `simulate.go:13,28` | Mob-vs-mob sim result | none |
| `SimulateCombat(mobIdA, mobIdB, levelA, levelB, maxRounds)` | `simulate.go:98` | Instant mob-vs-mob fight on unregistered copies | **none** (its tests are all commented out) |
| `type MobRank`, `RankMobs()`, `FormatMobRankings()` | `mob_rank.go:17,83,204` | Balance ranking of mob specs | `RankMobs`: `web/api_v1_mobs.go:25`; `FormatMobRankings`: none |
| `type WeaponRank`, `RankWeapons()`, `FormatWeaponRankings()` | `weapon_rank.go:14,87,193` | Weapon balance ranking | `RankWeapons`: `web/api_v1_items.go:18`; `Format…`: none |
| `type ArmorRank`, `RankArmor()`, `FormatArmorRankings()` | `armor_rank.go:20,156,235` | Armor balance ranking | `RankArmor`: `web/api_v1_items.go:34`; `Format…`: none |

## 2. The four `Attack*` variants side by side

### What is shared (all in `calculateCombat`, `combat.go:223-392`)

`calculateCombat` takes both characters **by value** (`combat.go:223`). Each variant passes a copy (`*user.Character`, `mob.Character`), so nothing inside mutates a live character. The one mutation it tries, `sourceChar.SetAggro(...)` at 258 for backstab, lands on the copy. Note that `Aggro` is a pointer and `SetAggro` replaces it (`character.go:1105`), so the live aggro is untouched. In practice the hook re-sets aggro after the round (`NewRound_DoCombat.go:348,503,758,880`).

Inputs read, all from `characters.Character`, for every pairing:
- Source: `Stats.Speed/Strength/Smarts/Perception.ValueAdj`, `StatMod("damage")`, `StatMod("attacks")`, `Equipment.Weapon/Offhand`, `GetSkillLevel("dual-wield")`, `Aggro.Type`, `GetRaceId()` (race `UnarmedName`, race dice via `GetDefaultDiceRoll`, `character.go:356`), `HasBuffFlag("accuracy")`, `Name`, `RoomId`.
- Target: `Stats.Speed/Strength/Smarts/Perception.ValueAdj`, `Race()` (racial weapon bonus, `combat.go:287`), `GetDefense()` (`character.go:566`), `HasBuffFlag("blink")`, `Name`, `RoomId`.
- No account field and no mob field is read in `calculateCombat`. The only kind-dependent input is the `SourceTarget` pair.

Randomness, in order (all `util.Rand`/`math/rand`; `Rand(n<1)` returns 0 without drawing, `util.go:173-178`):

1. `combatAttackCount` (`calculations.go:158`): no draw. Unarmed or a single claw gives `1 + extraAttackCount + StatMod(attacks)`; armed gives 1.
2. For each of `atkCount` iterations:
   1. If more than one weapon: `dualWieldActiveWeaponCount` draws `Rand(100)` only when dual-wield level is 2 and not both claws (`calculations.go:276`). Then one `Rand(len)` per weapon dropped (`combat.go:247`).
   2. Backstab check (no draw). `BackStab` is never set anywhere outside `aggro.go:11` (grep), so this is dead code.
   3. For each weapon, for each of its `attacks`:
      - `Hits`: `Rand(100)` (`calculations.go:126`).
      - If hit: `Dodges` draws `Rand(100)` (`:255`). If dodged: push the dodge lines and `continue`, which skips everything below (no defense roll, no message draws).
      - If hit and not dodged: `RollDice(dCount, dSides)` (`combat.go:316`), then `Crits` draws `Rand(100)` **only if `attackResult.Crit` is false** (`combat.go:320`, short-circuit `||`, the cascade bug).
      - `applyDefenseReduction`: `Rand(defenseRating)`, no draw when defense is 0 (`calculations.go:287`). It runs on misses too, with damage 0.
      - Message selection: `buildCombatMessages` calls `MessageOptions.Get(msgSeed)` once per slot, in the order attacker, defender, room (together: 3 slots, the 4th is nil) or attacker, defender, attacker room, defender room (separate) (`combat.go:179-189`). `Get` draws `Rand(len)` only when the seed is 0 and the list is non-empty (`attack_messages.go:49-50`).
         - `msgSeed` is the weapon `ItemId` when `ConsistentAttackMessages` is on, which it is (`_datafiles/config.yaml:316`; a per-world override was not found in `_datafiles/world/empty/*.yaml`). So an armed attack draws nothing and all slots take the same index. Unarmed (`ItemId` 0, for example the rat) draws once per non-empty slot, so the attacker, defender and room lines are **independently** chosen variants.
3. `buildCombatMessages` across rooms ranges over the room `Exits` map (`combat.go:193,202`), which is Go map order and not `util.Rand`.

Damage: `RollDice + weapon/race bonus + racial-bonus statmod + StatMod(damage) + damageBonus(Str delta)`. A crit adds `critDamageBonus` (Perception multiplier). Then defense reduction applies. `DamageToTarget` and `DamageToTargetReduction` accumulate (`combat.go:383-384`).

Crits: `Crit` is true if any crit happened (with the cascade bug). `BuffTarget = critBuffs` of the weapon or race, **overwritten** on each crit rather than appended (`combat.go:324`).

Messages pushed per swing (`combat.go:311-381`):

| Case | Source | Target | SourceRoom | TargetRoom |
|---|---|---|---|---|
| Dodge | `<ansi fg="cyan">%s dodges your attack!</ansi>` with **plain `targetChar.Name`**, no name colour (`:311`) | `You dodge the attack!` (`:312`) | none | none |
| Hit or miss, same room | `Together.ToAttacker` | `Together.ToDefender` + ` [you blocked N]` when damage>0 and reduction>0 (`:364-366`) | `Together.ToRoom` | not pushed (slot is nil, `:338`, so `len==0` at `:377`) |
| Hit or miss, different rooms | `Separate.ToAttacker` | `Separate.ToDefender` (+blocked) | `Separate.ToAttackerRoom` | `Separate.ToDefenderRoom` if non-empty |

A crit wraps every non-empty slot in `<ansi fg="yellow-bold">***</ansi> … ***` (`:342-349`). A backstab prefixes `*[BACKSTAB]*` (dead). Source, target and source-room lines are pushed even when empty (`:360,368,371`). `Account.SendText("")` queues `"\n"` (`userrecord.go:334-339`), a blank line. That case is reachable only across rooms with a non-generic, non-shooting subtype (section 4).

The re-substitution at `combat.go:372-373,379` is dead: `buildCombatMessages` already replaced `{target}`/`{targettype}` in those strings (`:211-217`). If it ran, it would set `{targettype}` to `user`/`mob` without the `name` suffix.

### What differs: the variant bodies (`combat.go:28-94`)

| | `AttackPlayerVsMob` `:29` | `AttackPlayerVsPlayer` `:47` | `AttackMobVsPlayer` `:65` | `AttackMobVsMob` `:80` | Real rule or incidental |
|---|---|---|---|---|---|
| Kinds passed | `User, Mob` `:31` | `User, User` `:49` | `Mob, User` `:67` | `Mob, Mob` `:82` | Incidental: only selects the name renderer and colour (below) |
| Apply `DamageToSource` to attacker | if `!=0`, then `WimpyCheck` `:33-36` | if `!=0`, then `WimpyCheck` `:51-54` | always (no guard), no wimpy `:69` | always, no wimpy `:84` | Dead: `DamageToSource` is never set |
| Apply `DamageToTarget` to defender | always, no wimpy `:38` | if `!=0`, then defender `WimpyCheck` `:56-59` | if `!=0`, then defender `WimpyCheck` `:71-74` | always, no wimpy `:85` | The `!=0` guard is incidental: `ApplyHealthChange(0)` changes nothing unless health is already below 0 (it re-cancels `cancel-on-combat` buffs) or above max (it clamps), `character.go:1201-1219`. **`WimpyCheck` is a real rule keyed on the account**: it reads the account config option `wimpy` and issues `flee` (`userrecord.go:617-623`). It runs for a player defender whenever damage was non-zero |
| Damage tracking | `mob.TrackPlayerDamage(user.UserId, DamageToTarget)` **always**, even on 0 damage or a total miss `:41` | none here; the hook does `defUser.TrackPlayerDamage(user.UserId, …)` **only if `Hit`** (`NewRound_DoCombat.go:294-296`) | none | if the attacker is charmed by an online account (`accounts.GetByBodyId(GetCharmedBy())`), `defMob.TrackPlayerDamage(charmer.UserId, …)` always `:88-91` | **Real rule**: the defender records damage per attacking account (or the charmer of an attacking NPC). It feeds kill credit and XP split (the reader is outside this package; unverified here). The always-vs-only-on-hit split between PvM and PvP is a difference to keep or rule on. MvP records nothing, which is correct since there is no account to credit |
| Account fields read | `user.UserId`, `wimpy` option | both `UserId`s (hook), both `wimpy` | defender `wimpy` | charmer `UserId` | `UserId` and `wimpy` come through `Account()` |
| Mob fields read | none beyond `Character` | n/a | none | none | none |
| XP, stat or cooldown effects | none in `combat` | none | none | none | XP, alignment and kills happen at death, outside this package |

Name rendering (`buildCombatMessages`, `combat.go:157-175`):
- A player is `Character.Name`, plain. The template wraps it in `<ansi fg="username">`.
- A mob is `GetMobName(ids.ID{}).String()` (`character.go:589`; `formattedname.go:49-85`): `<ansi fg="mobname[-downed]">Name</ansi>` plus ` (adj|adj)` when it has adjectives (dynamic `downed`/`shop`/`lit`/`hidden`/`poisoned`/form, plus stored `Adjectives`), and a title from the `OnGetFormattedName` hook (no handler registered; grep). The template wraps this again, so a mob reads `<ansi fg="mobname"><ansi fg="mobname">rat</ansi></ansi>`.
- The viewer is the zero ID, so the mob name does **not** vary per receiver today. The `aggro` suffix needs a viewer.

### `GetWaitMessages` (`combat.go:96-143`) vs the attack

| | Wait (`GetWaitMessages`) | Attack (`calculateCombat`) |
|---|---|---|
| Weapon for message table | main weapon subtype only (`:100`) | each active weapon (`:333`) |
| `{damage}` | `[Invalid]` (`:115`) | the damage number |
| Hidden source | target, source-room and target-room lines suppressed (`:124-140`) | not checked |
| Empty lines | skipped (`:120,126,130,135`) | pushed |
| Target-room line | only if rooms differ (`:134`) | only if the separate slot is non-empty |

## 3. `AttackResult` (`attackresult.go:3-17`)

| Field | Set by | Read by (non-test) | Slice 5 |
|---|---|---|---|
| `Hit bool` | `combat.go:315` | `NewRound_DoCombat.go:294,706,836` (offhand break test; PvP damage tracking) | keep |
| `Crit bool` | `combat.go:255,321,323` | `NewRound_DoCombat.go:302,712,841` (break chance modifier) | keep |
| `BuffSource []int` | **never** | `NewRound_DoCombat.go:269,444,685,814`; `simulate.go:137,158` | dead; candidate to drop |
| `BuffTarget []int` | `combat.go:324` | `NewRound_DoCombat.go:273,448,689,818`; `simulate.go:138,159` | keep |
| `DamageToTarget int` | `combat.go:383` | variants `:38,41,57,72,85,90`; hook `:296,340,748`; `simulate.go:133-136,154-157` | keep |
| `DamageToTargetReduction int` | `combat.go:384` | **nobody** | dead |
| `DamageToSource int` | **never** | variants `:33-34,51-52,69,84`; hook `:337,493`; `simulate.go:134,155` | dead |
| `DamageToSourceReduction int` | never | nobody | dead |
| `MessagesToSource []string` | `SendToSource`: `combat.go:121,311,360` | hook `:234,277,416,452` (players only; NPC-attacker lines are built then dropped) | message field |
| `MessagesToTarget []string` | `SendToTarget`: `combat.go:127,312,368` | hook `:238,281,664,693` (player targets only; PvM and MvM target lines dropped) | message field |
| `MessagesToSourceRoom []string` | `SendToSourceRoom`: `combat.go:131,371` | hook `:242-245,285,420,456,668,697,794,822` | **room-message field: delete** |
| `MessagesToTargetRoom []string` | `SendToTargetRoom`: `combat.go:136,378` | hook `:248-251,289,424,460,672,701,798,826` | **room-message field: delete** |
| `MessagesToRoomOld []string` | `SendToRoomOld` `:35`: **never called** | nobody | dead room-message field: delete |

The room fields the spec deletes are `MessagesToSourceRoom` and `MessagesToTargetRoom` (spec section 2: "This replaces `MessagesToSourceRoom` and `MessagesToTargetRoom`"), plus the unused `MessagesToRoomOld`. `MessagesToSource` and `MessagesToTarget` also become redundant once the result carries `perception.Message`(s). The spec row says only "room-message fields", so whether those two go too is a question (Q1).

How the hook delivers today (exclusions matter for parity):

| Pairing | Source lines | Target lines | Source-room | Target-room |
|---|---|---|---|---|
| PvP `:277-291` (wait `:234-252`) | `user.SendText` | `defUser.SendText` | `uRoom.SendText(msg, user, defUser)` | `defRoom.SendText(msg, user, defUser)` |
| PvM `:452-462` (wait `:416-426`) | `user.SendText` | **not sent** | `uRoom.SendText(msg, user)` | `defRoom.SendText(msg, user)` |
| MvP `:693-703` (wait `:664-674`) | **not sent** | `defUser.SendText` | `mobRoom.SendText(msg, defUser)` | `defRoom.SendText(msg, defUser)` (never filled: same room) |
| MvM `:822-828` (wait `:794-800`) | not sent | not sent | `mobRoom.SendText(msg)` | `defRoom.SendText(msg)` (never filled) |

Because mobs never receive text, all four amount to: the attacker gets the attacker lines, the target gets the target lines, everyone else in the source rooms gets the room lines. That is exactly what messaging's role split does. New in messaging: the NPC actor or target receives a structured `Perception` through `OnBehaviourPerceive`, which has no handler today (grep). Order today: buffs are applied before lines are sent (`:269-275` before `:277`), and `AddBuff` may itself send text.

## 4. Attack message data and the mapping to messaging tokens

Source: `_datafiles/world/empty/combat-messages/{bludgeoning,claws,cleaving,generic,shooting,slashing,stabbing,whipping}.yaml`. Loaded by `items/itemspec.go:544`; also edited by the admin page (`items/admin.go:187`, `web/api_v1_yaml.go:385`). Shape: `options[intensity].together{toattacker,todefender,toroom}` and `.separate{toattacker,todefender,toattackerroom,todefenderroom}` (`attack_messages.go:19-37`). The intensity is picked from the percent of max damage (`attack_messages.go:99-123`): `>=101` critical, `>=75` heavy, `>=30` normal, `>=1` weak, else miss. Wait rounds use `wait`. `prepare` is never requested: a grep for `items.Prepare` outside tests finds nothing, and the same grep pattern style found `items.Wait` at the four `GetWaitMessages` calls. `Prepare` is used only in `Validate` (`attack_messages.go:72`).

Fallback: a subtype that has the intensity key returns its own options even when the `separate` lists are empty (`:115-119`); `generic` is used only when the subtype or intensity is missing. **Only `generic.yaml` and `shooting.yaml` have any `separate` lines**; the other six have every separate slot empty, and `shooting` lacks `todefenderroom` for prepare and wait (parse script).

Tokens are `{name}` placeholders replaced by `strings.Replace` (`attack_messages.go:41-43`); there are no `%s` in the data (grep count 0). Tally over all 8 files:

| Token | Count | Filled with (`combat.go:157-208`) | Per receiver? |
|---|---|---|---|
| `{source}` | 379 | player `Name`; mob `GetMobName(zero).String()` | no (today) |
| `{sourcetype}` | 379 | `username` / `mobname` | no |
| `{target}` | 401 | as `{source}` for the target | no |
| `{targettype}` | 400 | `username` / `mobname` | no |
| `{itemname}` | 332 | weapon `DisplayName()` or race `UnarmedName` | no |
| `{damage}` | 240 | damage int (`[Invalid]` on wait) | no |
| `{exitname}` | 54 | exit from the attacker's room to the target's room | no |
| `{entrancename}` | 43 | exit from the target's room back to the attacker's room | no |

By slot: together `toattacker` names only the target; `todefender` names only the source; `toroom` names both. Separate `toattacker` names the target and `{exitname}`; `todefender` **never names the source** (only `{entrancename}`, `{damage}`, `{itemname}`); `toattackerroom` names the source and `{exitname}` (1 line names the target); `todefenderroom` names the target and `{entrancename}`.

Every non-comment occurrence is `<ansi fg="{sourcetype}">{source}</ansi>` or `<ansi fg="{targettype}">{target}</ansi>`, with two exceptions:
- `cleaving.yaml:69`: a bare `{target}`, with no colour today.
- 34 lines (together `todefender` 19, `toroom` 15) put the possessive inside the tag: `<ansi fg="{sourcetype}">{source}''s</ansi>`.

Mapping to messaging (`perception.Message`, `messaging/render.go`):
- `nameFor` renders `<ansi fg="username|mobname">Name</ansi>` (chosen by `Account() != nil`), or `someone` when unseen (`render.go:12-24`). So `<ansi fg="{sourcetype}">{source}</ansi>` becomes `{actor}` and the `{target}` form becomes `{target}`, with the same colour for players.
- `{itemname}`, `{damage}`, `{exitname}`, `{entrancename}` are the same for every receiver, so they are pre-substituted before `Send`. `{direction}` is not used: it exists only for Nearby far receivers (`messaging.go:166-181`), and both rooms of a ranged attack are source rooms.
- Slot to field:
  - together: `toattacker` → `ActorText`, `todefender` (+ blocked suffix) → `TargetText`, `toroom` → `ObserverText`.
  - separate: `toattacker` → `ActorText`, `todefender` → `TargetText`, `toattackerroom` → `ObserverText`, `todefenderroom` → `TargetRoomText`.
  - `Range: Room` covers both rooms (`messaging.go:113-135`).
  - Dodge → `ActorText` + `TargetText`, `ObserverText` empty (the connection writes nothing for empty text, `controller.go:38-41`).

Things the tokens cannot express, or that change bytes:

1. **Mob adjectives and title.** Today a mob's combat name carries `GetMobName` adjectives (for example `(elite|poisoned)`) and the `-downed` suffix. `nameFor` uses bare `Character.Name`. In `empty`, neither mob YAML has adjectives (grep `1-rat.yaml`, `2-guard.yaml`), so this is invisible there. It matches the spec's note that "text that varies per receiver beyond the three tokens … has no form yet".
2. **Double-wrapped mob colour.** Today a mob is `<ansi fg="mobname"><ansi fg="mobname">rat</ansi></ansi>`; with `{actor}` it is single-wrapped. The rendered colour should match. Byte parity before rendering would differ. Unverified how the smoke transcript records it (moot today, finding 5).
3. **Possessive inside the tag** (34 lines). `{actor}'s` puts `'s` outside the colour. To keep it, either edit those lines so `'s` is uncoloured (a visible colour change), or add a possessive token. Data edits touch content that the admin page edits too.
4. **"Someone" capitalisation** works only when the text starts with the token (`render.go:38-43`). The crit wrap `<ansi fg="yellow-bold">***</ansi> ` and the backstab prefix come first, so an unseen crit attacker reads lowercase `someone`. Also, today's lines open with `<ansi fg=…>{source}`. The capital test only fires once that wrapper is replaced by `{actor}`.
5. **Hidden attacker.** Today: wait lines are suppressed to target and room, and attack lines name it. Messaging: `canSee` is false for a `hidden` subject (`perceive.go:51`), so the target and observers read "someone" on every line, and nothing is suppressed. Hidden cannot arise in `empty`: its 3 buffs are `0-meditating`, `1-illumination`, `24-death_recovery`, none flagged `hidden` (grep). See Q3.
6. **Dodge line** uses an uncoloured target name today (`combat.go:311`); `{target}` would colour it.
7. **Darkness.** Today combat lines ignore light. A `Senses: Sight` message would be blocked for observers in a dark room, and the names would become "someone" (`canSee`). `Sight|Sound` passes Sound to everyone (`perceive.go:18-19`), but `ActorSeen` still follows sight, so names still become "someone" in the dark. The spec lists that as an intended difference. No dark rooms in `empty` (spec facts table).
8. **Several lines per round.** One round yields `atkCount × weapons × attacks` swings, each with its own lines, and dodges are interleaved. A single `perception.Message` holds one text per role. See Q2.
9. **Cross-room observer in the target room.** With messaging, the attacker is `canSee`-checked from the target's room (same-room rules, `far=false`). The separate data never names the source in `todefender`/`todefenderroom`, so there is no visible effect.
10. **Empty slots.** Today an empty attack slot still sends a blank line (`SendText("")` queues `"\n"`); messaging drops it. Reachable only for ranged non-shooting attacks, which `shoot` forbids.

## 5. `simulate.go`, `mob_rank.go`, `weapon_rank.go`, `armor_rank.go`

- `simulate.go`: does **not** call the variants. It calls `calculateCombat(*charA, *charB, Mob, Mob)` directly (`:132,153`) and applies health and buffs itself (`:133-138,154-159`). It reads `DamageToSource` and `BuffSource` (dead) and builds the message slices only to throw them away. `SimulateCombat` has no caller and its tests are commented out.
  - If `calculateCombat` starts building `perception.Message`s that need live `ActorId`s, the sim (which runs on unregistered copies, `newSimMob`) needs a path that skips messages, or a pure-maths core. Options: split into `resolveAttack` (rolls and damage, returning per-swing outcome structs) and message building, or delete `SimulateCombat` as dead.
- `mob_rank.go`, `weapon_rank.go`: no variant and no `calculateCombat`. They use `expectedDPS`, `combatAttackCount`, `newSimMob` (mob_rank only) and `baselineCharacter*`. They are unaffected unless `newSimMob` moves with simulate.
- `armor_rank.go`: no combat-path calls (`buffs`, `configs`, `items`, `statmods` only). Unaffected.

## 6. Tests in `internal/combat`

- `calculations_test.go`: 17 live tests pinning the pure formulas at the default config: `AlignmentChange` table, `statDelta`, `damageBonus`, `hitChance`, `extraAttackCount`, `weaponlessAttackCount`, `critChance` (accuracy/blink), `critMultiplier`, `critDamageBonus`, `dodgeChance`, `dualWieldHitPenalty`, `dualWieldActiveWeaponCount` (incl. the level-2 coin flip), `applyDefenseReduction`, `damagePercentOfMax`, `tameSizeModifier`, `tameHealthBonus`, `chanceToTame`. None touch the attack path.
- `weapon_rank_test.go`: `TestRankWeapons_Compiles` (live; loads config and items from the repo root, checks the three sorts are non-increasing). It also defines `loadTestData`. `TestFormatWeaponRankings_Output` is commented out.
- `simulate_test.go`: the whole file is commented out (4 tests).
- Nothing pins `calculateCombat`, `buildCombatMessages`, `GetWaitMessages`, the variants, or `AttackResult` (repo-wide grep of `*_test.go`). Slice 5 needs a seeded parity test, for example `rand.Seed` plus fixed characters: the old variant vs `Attack`, comparing damage, `Hit`, `Crit`, `BuffTarget`, health after, `PlayerDamage`, and per-role text. Note that `util.Rand` uses the global `math/rand` (`util.go:177`), so seeding works if nothing else draws during the test.

## Open questions for Calabe

**Q1. Which `AttackResult` fields go?** The spec row says "room-message fields".
- (a) Delete only `MessagesToSourceRoom`, `MessagesToTargetRoom` and the unused `MessagesToRoomOld`. Keep `MessagesToSource/Target` as strings. This leaves two message paths.
- (b) Delete all five message slices; `AttackResult` carries `Messages []perception.Message`. Also delete the never-set `DamageToSource`, `DamageToSourceReduction`, `BuffSource` and the unread `DamageToTargetReduction`, with their dead readers.
- Recommendation: (b). The slice rule is "no slice leaves two live paths behind", and the dead fields only keep dead code alive in four handlers and the sim.

**Q2. "Its messages are one `messaging.Message`", but a round has many swings.**
- (a) One `Message` per swing (and one per dodge), in order: `AttackResult.Messages []perception.Message`. Text per receiver stays line for line as today.
- (b) One `Message` per round, joining each role's lines with `\n`. This needs dodge lines in `ActorText`/`TargetText` only, and an empty observer slot for those swings. Joining changes nothing visible if `SendText` of a joined string equals N calls (unverified).
- Recommendation: (a). It keeps per-swing ordering and "someone" capitalisation per line, and a behaviour gets one perception per blow.

**Q3. Hidden attacker** (the spec defers this to slice 5).
- (a) Messaging rules: target and observers read "someone" on wait and attack lines; nothing suppressed.
- (b) Keep today exactly: suppress target and room on wait lines only, and name the attacker on attack lines. This needs a bypass of `canSee`.
- (c) Suppress target and room lines on wait **and** attack lines while hidden.
- Recommendation: (a). It is one rule for every message, and the current split (hidden while waiting, named while striking) looks accidental. Hidden cannot occur in `empty`, so parity is unaffected there.

**Q4. Mob names in combat lose their adjectives and title** under `{actor}`/`{target}`.
- (a) Accept: combat uses plain names, like every other messaging line.
- (b) Give `nameFor` a mob formatter (`GetMobName(viewer)`), which would also add the per-viewer `aggro` suffix (new).
- (c) Defer to slice 7's per-receiver render hook.
- Recommendation: (c), and accept (a) for slice 5. No mob in `empty` has adjectives, and slice 7 needs the hook anyway.

**Q5. The 34 possessive lines and `cleaving.yaml:69`** (colour moves or appears).
- (a) Edit the data to `{actor}'s` and accept that `'s` loses colour, and colour the cleaving target.
- (b) Add an `{actor's}`-style token to messaging.
- (c) Keep a combat-side pre-pass that maps `<ansi fg="{sourcetype}">{source}''s</ansi>` to `{actor}<ansi fg="…">'s</ansi>`.
- Recommendation: (a). It is a colour-only change, and the data files are ours to edit. Record it in the text-parity intended-differences list.

**Q6. Damage tracking differences.** Today PvM tracks on every attack, even a total miss for 0 (`combat.go:41`). PvP tracks only on `Hit` (`NewRound_DoCombat.go:296`). MvM tracks to the charmer. MvP tracks nothing.
- (a) Keep each exactly, keyed on `Account()` and charm.
- (b) Unify as "track when the defender takes damage from an account-driven or charmed attacker". This changes PvM: a 0-damage miss no longer refreshes `LastPlayerDamage` (the 30-round reset, `character.go:163`).
- Recommendation: (a) for slice 5 ("outcomes stay identical"), with the difference noted for Phase 5.

**Q7. The crit cascade bug** (`combat.go:320-323`).
- (a) Keep it for parity and mark it like the slice 2 bugs.
- (b) Fix it now, which changes crit rates and RNG order.
- Recommendation: (a). It is a balance change and belongs with Phase 5's ruleset.

**Q8. Combat text parity has no baseline.** The pre-4b transcript has no fight.
- (a) Record a new baseline fight from master before the change (admin spawns a rat, the bot fights it), with rolls masked.
- (b) Rely on a seeded unit test that compares the old variants against `Attack` per role before deleting the variants.
- Recommendation: both. (b) is exact and cheap. (a) is what the slice's Done-when line implies, but it needs a fixed RNG or masking of whole combat lines, since unarmed rats pick message variants at random.

**Q9. Dead code next to the slice**: `SimulateCombat` (no caller, tests commented out), `ChanceToTame` (no caller), `Format*Rankings` (no caller), backstab (`BackStab` never set).
- (a) Delete them in slice 5 where they touch the attack path (`SimulateCombat` uses `calculateCombat`).
- (b) Leave them for the strip phases.
- Recommendation: delete `SimulateCombat` (or move it to a messages-free core) in slice 5, because it is the only other `calculateCombat` caller. Leave the rest for the strip plan.


# Inventory E: the combat round (Phase 4b slice 5)

Branch `strip/phase-4b-5-one-combat-path`, read 2026-10-28. Read-only; no repo files or git state touched.

Read 29 of 29 files (lines grouped below; `accounts/users.go` + `userrecord.go` and the three `gmcp` files share a bullet):
- `internal/hooks/NewRound_DoCombat.go` (full, 947 lines, chunks 1-400, 400-799, 800-947)
- `2026-10-27-phase-4b-npcs-are-players-design.md` (lines 1-320 of the file; all of sections 3-4 and the "Notes for later slices" entries for slices 2c-5)
- `internal/combat/combat.go` (full, 392 lines)
- `internal/usercommands/suicide.go` (full, 309 lines)
- `internal/usercommands/suicide_npc.go` (partial, 1-400: the whole of `npcDeath` up to the no-loot tail)
- `internal/usercommands/flee.go` (full)
- `internal/usercommands/attack.go` (full)
- `internal/usercommands/shoot.go` (full)
- `internal/usercommands/go.go` (partial: 1-89, 225-348 incl. `inFight`)
- `internal/hooks/hooks.go` (full)
- `internal/hooks/mobflee_test.go` (full)
- `internal/hooks/NewRound_IdleMobs.go` (full)
- `internal/hooks/NewRound_AutoHeal.go` (full)
- `internal/hooks/NewRound_MobRoundTick.go` (full)
- `internal/hooks/MobIdle_HandleIdleMobs.go` (full)
- `internal/hooks/PlayerDrop_HandlePlayerDrop.go` (full)
- `internal/rooms/rooms.go` (partial: 866-906 `GetRandomExit`, 1140-1314 `GetMobs`/`GetPlayers`, 2365-2400 `CanPvp`)
- `internal/rooms/roommanager.go` (partial: 275-325 `MoveToRoom`)
- `internal/characters/character.go` (partial: 490-519 charm, 589-618 names, 1080-1130 aggro, 1201-1221 `ApplyHealthChange`)
- `internal/mobs/mobs.go` (partial: 220-275, 455-655)
- `internal/accounts/users.go` (partial: 130-199), `internal/accounts/userrecord.go` (partial: 277-350, 610-624)
- `internal/actors/registry.go` (partial: 90-119)
- `internal/events/eventtypes.go` (partial: 215-344)
- `modules/gmcp/gmcp.Char.go` (partial: 70-99, 440-481), `gmcp.Room.go` (83-110), `gmcp.Party.go` (53-78)
- `internal/behaviour/behaviour.go` (partial: 1-80)
- `docs/completed/2026-10-27-phase-4b-2-inventory.md` (partial: 275-311, the E8 bug list)

Other facts below come from greps, named where used. Line numbers are `NewRound_DoCombat.go` unless a file is named.

---

## 1. Structure of `NewRound_DoCombat.go`

| Function | Lines | What it does |
|---|---|---|
| `DoCombat` | 23-38 | `handlePlayerCombat`, then `handleMobCombat`, then `handleAffected` on the concatenated affected lists (player list then mob list) |
| `handlePlayerCombat` | 40-520 | loop over `accounts.GetOnlineUserIds()` (46) |
| `handleMobCombat` | 522-897 | loop over `mobs.GetAllMobIds()` (527) |
| `handleAffected` | 899-947 | dedups; players: `Health <= -10` sets `KilledBy` then `user.Command("suicide")` (910-919), else `Health < 1` queues `PlayerDrop` (921-923); mobs: `Health < 1` queues `mob.Command("suicide")` (938-940) |

Per-player prologue (46-68): `no-combat` buff skip (51), nil user/aggro skip (55), cancel `cancel-on-combat` buffs (60), load room or skip (64-67).
Flee (69-168): players only (section 3).
Branches, entered in sequence, not `else` (an actor whose aggro survives the first `if` is re-tested by the second):
- **PvP** (player attacks player): 177-351, test `accounts.GetByBodyId(target) != nil`.
- **PvM** (player attacks NPC): 355-507, test `!Target.IsZero() && GetByBodyId(target) == nil`. Comment 354: "any target that is not a player, so a vanished target is cleared here".

Per-NPC prologue (527-559): nil mob/aggro skip (532), `no-combat` skip (537), room nil clears aggro with event (541-547), Flee with zero target clears aggro with event (552-556; slice 4 stand-in, test `mobflee_test.go:15-34`), cancel `cancel-on-combat` (559), `CombatCommands` (569-598, section 2), append self to `affectedMobIds` (601).
- **MvP** (NPC attacks player): 604-761.
- **MvM** (NPC attacks NPC): 765-884, same "not a player" test as PvM (764-765).

### Kept bugs (comments in the file, and the unmarked ones from the same list)

| Id | Line | Text / fact |
|---|---|---|
| E8c | 100 | `// Pre-existing bug E8c, kept for slice 5: the shadowed userId makes this compare u's target with u itself.` Code 98-101: `for _, userId := range uRoom.GetPlayers(rooms.FindFighting)` then `u.Character.Aggro.Target != u.Character.Id`. A player cannot target itself (`attack.go:164-166`), so **no player ever blocks a flee today**. |
| E8e | 437 | `// Pre-existing bug E8e, kept for slice 5: this appended the aggro target's user ID, always 0 in a mob fight.` Code 438 `affectedPlayerIds = append(affectedPlayerIds, 0)`. Harmless: `GetByUserId(0)` at 908 finds nobody (unverified that no account has UserId 0). Side effect: the attacking player is never in the affected list in PvM, nor in PvP (263 appends only the defender). |
| E8f (unmarked) | 51 vs 55 | `user.Character.HasBuffFlag("no-combat")` dereferences `user` before the `user == nil` check. Listed in `docs/completed/2026-10-27-phase-4b-2-inventory.md:283`; spec note line 282 ("`user` dereferenced before its nil check"). No comment marks it. |
| E8g | gone | "spell fizzle clearing aggro without an event" (inventory :284). `grep -i fizzle internal` returns nothing and the file has no SpellCast branch; it went with the dead SpellCast code. The spec note at line 282 is stale on this item. |
| KilledBy zero for PvP | 912 | `KilledBy` is set only when the aggro target is not a player. See also section 2 "death": in practice it is never set by combat at all. |

---

## 2. Difference table across the four branches

Key: **R** = real rule keyed on player vs NPC (with the actor question that replaces the type check); **I** = incidental duplication or drift; **B** = looks like a bug (keep or fix needs a ruling).

### Entry, target finding and aggro clearing

| Aspect | PvP | PvM | MvP | MvM | Verdict |
|---|---|---|---|---|---|
| `no-combat` / nil order | 51 before 55 (E8f) | same | 532 nil then 537 | same | I (E8f is the bug) |
| `cancel-on-combat` on attacker | 60, before flee | 60 | 559, after the flee clear | 559 | I (order only matters for a fleeing actor: a player's flee cancels them, an NPC's flee does not) |
| Attacker room nil | 65 skip, aggro kept; again 183-186 clears aggro, no event | 371-374 clears, no event (only inside the remote case) | 543-546 clears, event | same | I |
| Target in another room | allowed only with `Aggro.ExitName` whose exit leads to the target's room (193-200) | same (366-381) | not allowed: clears aggro, event (607-611) | not allowed (771-775) | **R-ish**: today only players reach `ExitName`, because only `shoot` sets it (`shoot.go:86,116` via `SetAggroRemote`, `character.go:1081-1084`). An NPC that now runs `shoot` (slice 4 gave NPCs the user command set) has its remote aggro cleared silently. Replace with: one rule for all actors ("remote target allowed through `ExitName`"). Ruling needed (Q3). |
| Target not found message | "Your target can't be found." (205) | same (387) | none | none | R: text only to an actor with a connection. Messaging handles it (ActorOnly interface text drops for a behaviour). No type check needed. |
| Event on target not found | always (207) | only if `defMob == nil` (392-394); a target that moved rooms clears aggro with **no** `AggroChanged` | always (609) | always (773) | B/I: PvM's missing event when the target moved. |
| Target room nil | clears, event (211-215) | not checked: `defRoom` (398) used at 425, 461 unguarded | clears, event (613-617) | not checked (777, used 799, 827) | I (nil-deref risk on two branches) |
| Cancel target's `cancel-on-combat` | 218 | 400 | 620 | 779 | same everywhere |
| Target downed (`Health < 1`) | "Your rage subsides." + clear + event (220-224) | same (402-406) | silent clear + event (622-625) | same (781-784) | R: the text goes to whoever has a connection; same rule otherwise |
| Target hidden | after RoundsWaiting; "You can't seem to find your target." and skip, aggro kept (258-261) | same (432-435) | **before** weapon retry and RoundsWaiting, silent skip (627-629) | after RoundsWaiting, silent skip (806-808) | I: position differs. In MvP a hidden target freezes `RoundsWaiting` (not decremented); in the other three it still counts down. Merged path must pick one order (Q5). |
| Same-round re-target | `SetAggro(def, DefaultAttack)` if both alive (348), **no event** | same, with event (503-504) | same, event (758-759) | same, event (880-881) | I: PvP missing event |
| `SetAggro` side effects | `SetAggro` builds a new `Aggro` (`character.go:1104-1108`): drops `ExitName`, resets `RoundsWaiting` to weapon+offhand `WaitRounds`, turns DefaultAttack into Shooting for a shooting weapon | same | same | same | Shared fact: after the first remote round, `ExitName` is lost, so next round the remote target "can't be found" (193-194/366-367). Shooting through an exit lasts one attack round today. B or intended? (Q3) |

### RoundsWaiting (weapon wait) messages

| | PvP 227-255 | PvM 409-429 | MvP 657-677 | MvM 787-803 |
|---|---|---|---|---|
| `GetWaitMessages` kinds | User,User | User,Mob | Mob,User | Mob,Mob |
| Sent | Source, Target, both rooms excluding both | Source, both rooms excluding attacker | Target, both rooms excluding target | both rooms, no exclude |

I: every branch sends each part to whoever can receive text. One `messaging.Message` with roles covers all four. Hidden attacker: `GetWaitMessages` suppresses target and room text when the source has `hidden` (`combat.go:124-140`); `calculateCombat` has no such check (`combat.go:223-392`), so a hidden attacker's actual hits are seen by everyone. Spec note line 304-306 asks slice 5 to rule on this (Q6).

### Attack, damage, messages

| Aspect | PvP | PvM | MvP | MvM | Verdict |
|---|---|---|---|---|---|
| Call | `AttackPlayerVsPlayer` (265) | `AttackPlayerVsMob` (442) | `AttackMobVsPlayer` (681) | `AttackMobVsMob` (812) | all four call `calculateCombat(src, tgt, srcType, tgtType)` (`combat.go:31,49,67,82`) |
| Kind args to `calculateCombat` | used only for name tokens: `TokenSourceType`/`TokenTargetType` = `username`/`mobname` colour, and a Mob side's name is `GetMobName(ids.ID{})` (`combat.go:160-175`) | | | | R: replace with "does the actor have an account" for colour, or a name-render question on the actor. Messaging tokens may take this over |
| Apply damage to target | if nonzero (`combat.go:56-58`) | always (`combat.go:38`) | if nonzero (71-73) | always (85) | I. `ApplyHealthChange(0)` is not a no-op: it clamps `Health` to `HealthMax` and cancels `cancel-on-combat` when `Health < 0` (`character.go:1201-1221`). Tiny outcome difference only |
| `DamageToSource` / `BuffSource` | handled 51-54, 337-339, 269-271 | 33-36, 493-495, 444-446 | 69, 685-687 | 84, 814-816 | **dead**: `calculateCombat` never sets either (grep: only `simulate.go` and these readers). Merged path can drop them, or keep for Phase 5 |
| `WimpyCheck` | both sides if damaged (`combat.go:53,58`) | source only (35) | target (73) | none | R: account config option (`userrecord.go:617-623`, issues `flee` with wait -1). Replace with `Account() != nil` |
| Damage credit `TrackPlayerDamage` | in the hook, **only on Hit** (296) | in `combat.go:41`, **always** (even 0 damage, no hit) | none | to the charmer's UserId if the attacker is charmed by a player, always (`combat.go:88-91`) | **R**: this map is the only input to kill credit, XP split, alignment, tame mastery (`suicide_npc.go:56-293`) and PvP stats (`suicide.go:51-81`). Actor question: "which account, if any, gets credit for this attacker": `Account()` of the attacker, else the account of `GetCharmedBy()`. Gaps: a charmed NPC hitting a **player** gives no credit; PvP needs a Hit, PvM does not (Q2) |
| Buffs from crits | `AddBuff` both (269-275) | same | same | same | I (both `AddBuff` queue the same `events.Buff`: `userrecord.go:324-331`, `mobs.go:243-251`) |
| Attack messages | Source, Target, rooms excl. both (277-291) | Source, rooms excl. attacker (452-462) | Target, rooms excl. target (693-703) | rooms only (822-828) | I: same as wait messages |
| Offhand break check | on Hit, defender's offhand (299-334) | **none** | on Hit (709-745) | on Hit (838-871) | B/I: an NPC's offhand never breaks when a player hits it, but does when an NPC hits it (Q4) |
| Broken item placement | full pack: into `room` = **attacker's** room (267, 325) | n/a | attacker's room (683, 735) | defender's room (848, 861) | I/B for ranged hits |
| Break message | 3 lines to defender + defRoom excl. defender (309-313) | n/a | same (719-723) | defRoom only (850) | I (messaging) |
| `CharacterVitalsChanged` | both sides if damaged (337-342) | source only (493-495, dead because `DamageToSource` is 0) | target (748-750) | none | R: `UserId` event, GMCP only (inventory question 3). Replace with `Account() != nil`. Note PvM never fires it for the attacker, which is correct because the attacker takes no damage |
| `MakeHostile` (group hatred) | none | each `defMob.Groups`, `MakeHostile(group, user.UserId, MinutesToRounds(2) - Perception)` every round (468-470; comment says 5 minutes, code is 2) | none | none | R: keyed on attacker account (`mobs.go:598-608`, read by `go.go:271-282`). Replace: "target is an NPC with groups and attacker has an `Account()`". What a charmed NPC attacker should do is open (Q2) |
| Defender reaction | none | if `defMob.Aggro == nil`: `PreventIdle = true`; if the attacker is in another room, `go <exit>` toward them and an angry command (473-490) | none | if `defMob.Aggro == nil`: `PreventIdle = true` only (831-833) | R/I: "target has a mob" question (`Actor.Mob` or `mobs.GetInstance`). The move only happens in PvM because only players reach a remote attack today. Nothing in the round makes the defender fight back (only `SetAggro` callers: `grep \.SetAggro\(`: attack.go, the round itself, simulate.go); `PATCH_NOTES.md:24` "mobs no longer start or join fights until Phase 5" |
| End of fight | either dead: `EndAggro` both, **no events** (344-346) | events both (497-501) | events both (752-756) | events both (874-878) | I: PvP missing `AggroChanged` |

### NPC-only steps

| Step | Lines | RNG | Actor question |
|---|---|---|---|
| `CombatCommands` | 569-598: only when `Aggro.Type == DefaultAttack`; `Rand(100) < ActivityLevel` then `Rand(cmdCt)`; blank entry is a no-op; too-long entry says a canned line; queues each `;` part 0.1 s apart. **Replaces the attack** that round (`continue` 581, 594), even when the target is gone, and skips the self-append at 601 | 2 | behaviour data on the mob template. Belongs to `BehaviourController` / `Actor.Mob`, not a type check |
| Weapon retry | 635-655, MvP only (not MvM): no weapon and a non-empty pack, `Rand(100) < Perception` then `Rand(len)` and `equip <name>`; happens before RoundsWaiting and is queued (takes effect later) | 1-2 | NPC behaviour; MvM lacks it (I) |
| Angry command | 482-484 (PvM remote case only), `GetAngryCommand` `Rand` (`mobs.go:464-478`) | 1 | NPC behaviour |

### Death and kill credit (not in the round)

The round only decides **who is checked** and queues `suicide`; every credit rule lives in the `suicide` command:

- `handleAffected` players (902-928): `Health <= -10` sets `KilledBy`, `KillerMobIsElite`, `KillerMobName` if the aggro target is an NPC, then queues `suicide`. Otherwise `Health < 1` queues `PlayerDrop` (handled by `PlayerDrop_HandlePlayerDrop.go:16-42`: two lines of text).
  - **Reasoned, not tested: the `-10` branch is unreachable from combat.** `ApplyHealthChange` caps a hit from positive health at -5 (`character.go:1209-1210`); every branch stops attacking a target with `Health < 1` (220, 402, 622, 781); and a dying player's aggro was already ended at 344-346/752-756 before `handleAffected` reads it (912). So `KilledBy` is never set by combat, `PlayerDeath.KillerMobId` is 0 and `AddEliteDeath` never runs (`suicide.go:94-105`). A downed player bleeds in `AutoHeal` (`NewRound_AutoHeal.go:44-56`, every 3 rounds, only with no aggro) and dies there (`:46-48`), which sets no killer either.
- `handleAffected` mobs (931-945): `Health < 1` queues `suicide`. `MobRoundTick.go:103-105` also queues it next round if `Health <= 0`, so two can be queued (unverified what the second does).
- `Suicide` branches on `mobs.GetInstance(actor.Id) != nil` (`suicide.go:27-29`) into `npcDeath`:
  - `KD.Kills` via `AddMobKill(MobId)` per player in `PlayerDamage`, skipped in zone `Training` (`suicide_npc.go:107-112`, party path 229-234); elite kills.
  - XP split by `len(PlayerDamage)`, `Rand(3)` variation (`:80-83`), level scaler, elite bonus, parties.
  - Alignment `combat.AlignmentChange` + `AlignmentChanged` event (`:142-160`, `:238-256`).
  - Tame mastery `Rand(1000)` (`:178-190`, `:274-289`), keyed on template `MobId`.
  - Loot: pack always drops, worn items on `Rand(100) < ItemDropChance` (`:299-360`), corpses.
  - `MobDeath` event with `KilledByUsers` (`:61-69`), and clears the aggro of every player in `PlayerDamage` aimed at it, with events (`:95-100`).
- Player death (`suicide.go:44-308`): `PlayerKills`/`PlayerDeaths` from the dead player's `PlayerDamage` (`:59-81`), which only the PvP branch fills (296). No alignment change for PvP.
- `playersAttacked`: **gone.** Deleted in slice 2 (`docs/completed/2026-10-27-phase-4b-2-ids-plan.md:509`); `grep -i playersattacked` over the repo finds only docs. The spec's list at line 209 is stale on it.
- Charm: no charmed-assist code exists anywhere (every `IsCharmed` call site, by grep: flee followers 153, `go.go:246,267`, `attack.go:191`, `shoot.go:81`, teleport, `HandleLeave`, `MobRoundTick` expiry, `behaviour.go:26`, `HandleIdleMobs.go:23`, rooms finders). In the round, charm appears only in flee followers and the charmer damage credit.
- PvP checks: **none in the round.** `CanPvp` (`rooms.go:2372-2400`) is called only by `attack` (`attack.go:235`); `shoot` against a player checks the party but not `CanPvp` (`shoot.go:101-129`). A fight that walks into a non-PvP room continues.

So the merged round keeps outcomes by preserving: (a) which actor's account is credited in `PlayerDamage` and when (Hit vs always); (b) who goes into the affected lists; (c) `MakeHostile`. Everything else downstream reads those.

---

## 3. Flee

### Player half (69-168)

1. `flee` command (`flee.go:10-21`) sets `Aggro{Type: Flee}` with a zero target and queues `AggroChanged`. It needs no fight: a player not fighting also flees.
2. Round: `SetAggro(zeroTarget, DefaultAttack)` first (72), "regardless of outcome".
3. Mob blockers (74-94): for each NPC in the room with `FindFighting` whose target is the fleer, in `GetMobs` order: chance = `Speed_fleer / (Speed_fleer + Speed_blocker) * 70 + 30`; `roll := Rand(100)`; `roll >= chance` blocks, first blocker wins (break).
4. Player blockers (96-119): same formula over `GetPlayers(FindFighting)`, but E8c means no player ever qualifies, so no roll happens.
5. Blocked: text to fleer and room (121-131); `continue`. **The aggro is left as `{zero target, DefaultAttack}`**, which neither branch acts on (PvP needs `GetByBodyId(zero)`, which returns nil, `users.go:185-187`; PvM needs a non-zero target). Reasoned, not tested: the player then stays "in a fight" for `inFight` (`go.go:331-334`, any non-nil aggro) and cannot walk, and `AutoHeal` skips them (`AutoHeal.go:33`), until they `flee` again or `attack`.
6. Exit: `GetRandomExit` (`rooms.go:868-906`), non-secret unlocked exits plus mutator exits, one `Rand(len)` over a map (order random). No exit: text, `continue` (same stuck aggro).
7. Success: texts (141-142), `Aggro = nil` + event (144-145), `rooms.MoveToRoom` directly (147), which bypasses `go`, `inFight` and action points.
8. Charmed followers (149-157): every `FindCharmed` NPC in the **old** room charmed by the fleer gets `mob.Command("flee")` (queued input). It takes effect at input processing, and the next round's NPC half clears it (552-556) without moving: today a pet "flees" by stopping its fight and stays behind.
9. `usercommands.Look("", actor, newRoom, CmdSecretly)` (161-163).
10. Blockers do not get `MakeHostile`; no hatred is involved in flee (grep: `MakeHostile` is called only at 469).

### What NPC flee needs

- Who blocks: every actor whose aggro target is the fleer. Today only NPCs block players. Under one loop the question is whether players block too (fixing E8c) and whether NPCs block NPCs (Q1).
- `MoveToRoom` already moves NPCs (`roommanager.go:282-286` accepts an actor with no account and a mob instance).
- Wording: room text uses `username` colour (123, 129, 142); an NPC needs `mobname`, or a messaging `{actor}` token.
- No `Look` for an NPC (slice 4 note: an NPC's move or look prepares nothing) or a call that its controller drops.
- The 552-556 stand-in and `TestFledNPCStopsFighting` (`mobflee_test.go:15-34`) go; the test becomes "a fled NPC leaves".
- Charmed followers of an NPC: nothing charms an NPC to an NPC today (charm is by a player body); the follower loop works for any `IsCharmed(fleerId)`.
- `MaxWander`/home: a fled NPC is displaced; `HandleIdleMobs.go:28-32` paths it home if `MaxWander == 0`. Nothing else needed.
- Wimpy: account-only (`WimpyCheck`), so NPCs flee only by command (combat commands, behaviour).

---

## 4. Outside the file

| Dependency | Where | Used at |
|---|---|---|
| `combat.AttackPlayerVsMob/PlayerVsPlayer/MobVsPlayer/MobVsMob` | `combat.go:29-94` | 265, 442, 681, 812 |
| `combat.GetWaitMessages`, `combat.User/Mob` | `combat.go:21-26, 96-143` | 232, 414, 662, 792 |
| `calculateCombat` RNG: dual wield `Rand` (247), `Hits` (308), `Dodges` (310, `calculations.go:253-258`), `RollDice` (316), `Crits` (320, `calculations.go:200-210`) | `combat.go:223-392` | via the four attacks. Note `calculateCombat` takes characters **by value**, so the BackStab `SetAggro` at `combat.go:258` changes a copy; the hook's own `SetAggro` (348/503/758/880) resets the real one |
| `rooms.FindFighting`, `FindCharmed`; `GetMobs`/`GetPlayers` | `rooms.go:43-56, 1142-1314`; `FindFighting` = target is a live account or a live mob (1178-1186, 1281-1289) | 75, 98, 149; also GMCP `Char.Enemies` (`gmcp.Char.go:457`), `attack` with no argument (`attack.go:45,65`) |
| `Room.FindExitByName`, `GetRandomExit`, `Exits` | `rooms.go:1762`, `868` | 197, 378, 134, 479 |
| `rooms.MoveToRoom` | `roommanager.go:275` | 147 |
| `rooms.LoadRoom` | | throughout |
| `characters` aggro: `SetAggro`, `EndAggro`, `Aggro.{Target,Type,ExitName,RoundsWaiting}`, `IsCharmed`, `GetCharmedBy`, `TrackPlayerDamage`, `CancelBuffsWithFlag`, `HasBuffFlag`, `ApplyHealthChange` | `character.go:159-171, 496-514, 1081-1130, 1201-1221` | throughout |
| `mobs.MakeHostile`, `GetAngryCommand`, `PreventIdle`, `CombatCommands`, `ActivityLevel`, `Groups`, `Command` | `mobs.go:598-608, 464-478, 72, 253-275` | 469, 482, 474/832, 572-593 |
| `accounts.GetOnlineUserIds/GetByUserId/GetByBodyId`, `WimpyCheck`, `Account.Command` | `users.go:140-149, 184-199`; `userrecord.go:277-290, 617-624` | throughout |
| `usercommands.Look`, `actors.ByUserId` | `look.go:22`; `registry.go:84` | 161-162 |
| `items.New(20)` broken item, `BreakTest` (`items.go:100`, one `Rand(100)`) | | 306, 716, 845 |

Events fired by the round and their consumers:

| Event | Fired | Consumers |
|---|---|---|
| `AggroChanged{ActorId, RoomId}` | 23 sites (grep) | `gmcp.Room.go:85-110` (Room.Info.Contents Players or Npcs, keyed on `actors.AccountOf`), `gmcp.Party.go:55-78` (party vitals), `mobflee_test.go` |
| `CharacterVitalsChanged{UserId}` | 338, 341, 494, 749 | GMCP (inventory question 3; not re-read) |
| `ItemOwnership` | 315, 327, 725, 737, 852, 863 | not re-read |
| `Buff` (via `AddBuff`) | 270, 274, ... | `ApplyBuffs` (`hooks.go:11`) |
| `PlayerDrop` | 923 | `HandlePlayerDrop` (`hooks.go:52`) |
| `Message` (all `SendText`) | throughout | `Message_SendMessage` |
| Input (`Command`: `suicide`, `flee`, `equip`, `go`, combat/angry commands) | 154, 481-483, 587-591, 651, 919, 940 | world input loop |
| `MobDeath`, `PlayerDeath` | not in the round: `suicide_npc.go:61`, `suicide.go:107` | `gmcp.Char.go:57-58,75-99` (Char.Kills to `KilledByUsers`), `follow.go:66-67,352,363` |

GMCP `Char.Enemies` (`gmcp.Char.go:444-479`) is built on request from the room's `GetMobs(FindFighting)`: it lists only NPCs, and `Engaged` only when the player's target is a mob. It reads state, not round events.

Tick placement (`hooks.go:17-31`): `UserRoundTick`, `MobRoundTick`, `HandleRespawns`, then `DoCombat`, then `AutoHeal`, `IdleMobs`. `IdleMobs` consumes `PreventIdle` set by the round in the same round (`NewRound_IdleMobs.go:60-63`).

Tests touching the round: only `mobflee_test.go:28` calls `handleMobCombat` (grep of `*_test.go` for the four function names). No smoke fight exists yet (grep `cmd/smoke` for `attack`/`flee`: none). The `empty` world's two mobs (`_datafiles/world/empty/mobs/startland/1-rat.yaml`, `2-guard.yaml`) have no combat commands and no attack commands (grep `attack|combatcommands|hostile: true` over `empty` mobs and races: none), so in `empty` the MvP and MvM branches run only if someone makes an NPC `attack`.

---

## 5. Order and randomness

- **Order today:** all players, then all NPCs, then `handleAffected` (players list, then mobs list). Within each phase the order is Go map iteration: `GetOnlineUserIds` ranges `userManager.Users` (`users.go:145`), `GetAllMobIds` ranges `mobInstances` (`mobs.go:231`). So the within-kind order is already random per round.
- **RNG:** `util.Rand` is the global `math/rand` `Intn` (`util.go:173-178`); no seeding hook found in `internal/util` (grep `rand.Seed|rand.New|func Seed`: none). Exact roll streams cannot be reproduced today, so "identical outcomes" can only mean identical rules and odds, not an identical roll sequence.
- **Rolls per actor**, in order: player: flee (one `Rand(100)` per NPC blocker until one blocks; `GetRandomExit` one `Rand`), or attack (`calculateCombat` rolls, then `BreakTest` on hit). NPC: `CombatCommands` (`Rand(100)`, `Rand(n)`), weapon retry (`Rand(100)`, `Rand(n)`), attack rolls, `BreakTest`; PvM defender angry command `Rand`.
- **What a merged loop must keep:**
  1. **Accounts before NPCs.** It is an outcome, not an accident: a player always swings before any NPC in the round, and an NPC killed by a player's swing has its aggro ended (498) before its own turn. Interleaving (for example `actors.All()`, which sorts by ID text, `registry.go:93-103`) would let an NPC drop a player before the player swings. Keep two passes over one actor list (account-driven first), or sort by `Account() != nil`. Ruling needed (Q7).
  2. Affected lists in the same order (players then mobs) so queued `suicide` inputs and `PlayerDrop` events keep their order.
  3. The checks that decide credit and affected membership (section 2 death).
  4. `DoCombat` stays between `MobRoundTick` and `IdleMobs` in `hooks.go`.
- **Incidental:** within-kind order; which branch body each check lives in; the per-kind message fan-out; `DamageToSource`/`BuffSource` handling; redundant room loads (181, 267, 371, 683).
- `actors.Each` holds the registry read lock while it runs (`registry.go:105-115`); the round calls `MoveToRoom` (`actors.Get`, a read) and queues `suicide` rather than calling it, so `Each` may be safe, but `BodyReplaced` (permadeath reroll) takes the write lock and runs from `suicide`, which is queued. Unverified that nothing in the round reaches a write path synchronously; `actors.All()` (a copy) avoids the question.

---

## Open questions for Calabe

**Q1. Who blocks a flee (E8c).** Today only NPCs block a fleeing player; the player-blocker loop never fires.
- (a) Keep exactly: NPCs block players, nobody blocks NPCs. Needs a kind test in the blocker loop.
- (b) One rule: any actor whose target is the fleer may block, players and NPCs alike. Fixes E8c; a player fighting another player can now stop their flee, and players can stop a fleeing NPC.
- (c) Keep E8c for players, NPCs block anyone.
Recommend (b): it is the only one with no kind test, and the bug was kept only to hold outcomes through slice 2. It changes PvP flee odds (rare today).

**Q2. Damage credit (`PlayerDamage`), which drives kills, XP, alignment, mastery and PvP stats.** Today PvM credits every round even with no hit, PvP only on a hit, charmed NPC credits its charmer only against NPCs.
- (a) Keep each case exactly.
- (b) One rule: credit `Account()` of the attacker, else the account of its charmer, on every attack against any target. A pet's kill of a player then counts as its owner's PvP kill; a PvP miss starts counting toward the kill list.
- (c) One rule, credit on hit only. Changes PvM: a player who only missed no longer shares the XP.
Recommend (b) with "always" (matches PvM, which is the common case and the smoke fight), and record the PvP and pet changes in PATCH_NOTES. Same rule decides `MakeHostile`: attacker account, else charmer's account.

**Q3. Attacks through an exit.** Only players reach it today (`shoot`), and `SetAggro` at the end of the round drops `ExitName`, so a ranged attack lasts one round. NPCs with `shoot` lose the aggro silently.
- (a) Keep: remote for accounts only, one round.
- (b) Remote for every actor, still one round (keep the `SetAggro` reset).
- (c) Remote for every actor, and keep `ExitName` across rounds.
Recommend (b): removes the kind test without changing player outcomes. (c) is a behaviour change worth its own slice.

**Q4. Offhand breakage and drift bugs.** NPC offhands never break under a player's hits; broken items land in the attacker's room; PvP end of fight and PvP re-target fire no `AggroChanged`; PvM fires none when the target moved; two branches use an unchecked `defRoom`.
- (a) Keep each exactly.
- (b) One rule for all, fixing the four.
Recommend (b): each fix is the behaviour three of the four branches already have, and the event fixes only make GMCP more accurate. Breakage under player hits does change item loss for NPCs; list it.

**Q5. Hidden target and wait order.** In MvP a hidden target stops the attacker's `RoundsWaiting` countdown and the weapon retry; elsewhere the countdown runs.
- (a) Keep per kind. (b) One order: weapon retry, RoundsWaiting, then hidden check (the PvP/PvM/MvM order).
Recommend (b).

**Q6. Hidden attacker** (spec note): wait messages hide a hidden attacker from target and room; real hits do not.
- (a) Keep both as today. (b) Messaging shows "Someone" for both. (c) Suppress both.
Recommend (b), which is what the messaging design already does for unseen actors, and it is the listed intended text-parity difference.

**Q7. Turn order in the merged loop.**
- (a) Accounts first, then NPCs (today's order, kept with an `Account() != nil` partition).
- (b) One order for everyone (ID order or random).
Recommend (a) for this slice: it is the one ordering fact that changes who wins a close fight, and initiative belongs to the Phase 5 ruleset. It means the loop keeps one question about the controller, which the spec allows (`Account()` is the sanctioned question).

**Q8. Stuck aggro after a blocked or exitless flee.** The player keeps a target-less `DefaultAttack` aggro, cannot walk and does not bleed or heal until they act. NPCs would inherit this under one flee path.
- (a) Keep the target-less aggro for every actor.
- (b) Clear the aggro (with event) when the flee is blocked or finds no exit.
- (c) Restore the pre-flee target (lost today: `flee.go:15` overwrites it), so a blocked fleer keeps fighting.
Recommend (c): a blocked flee leaves you in the fight you tried to leave, which matches the "no escaping a fight except by fleeing" ruling, and it ends the stuck state for players and NPCs alike. It is a behaviour change (today a blocked fleer stops swinging), so it goes in PATCH_NOTES.


# Inventory F: slice 5 combat output, hidden attacker, parity tooling, smoke fight, GMCP

Branch `strip/phase-4b-5-one-combat-path` at `58a0cbc0` (= master `dc94a7f7` + docs commit), read 2026-10-28.

Read 37 of 37 files:

Full: `2026-10-27-phase-4b-npcs-are-players-design.md` (330 lines), `internal/messaging/messaging.go`, `perceive.go`, `render.go`, `messaging_test.go` (two chunks), `main_test.go`, `render_test.go`, `internal/perception/perception.go`, `internal/combat/combat.go`, `attackresult.go`, `internal/hooks/NewRound_DoCombat.go` (three chunks), `internal/hooks/Message_SendMessages.go`, `internal/hooks/Broadcast_SendToAll.go`, `internal/usercommands/attack.go`, `flee.go`, `shoot.go`, `consider.go`, `suicide.go`, `suicide_npc.go` (two chunks), `cmd/smoke/main.go`, `script.go`, `client.go`, `recorder.go`, `admin.go`, `copyover.go`, `server.go`, `cmd/smoke/testdata/pre-4b/bot0.txt`, `_datafiles/world/empty/mobs/startland/1-rat.yaml`, `_datafiles/world/empty/races/1-human.yaml`, `races/10-rodent.yaml`.

Partial: `internal/combat/calculations.go` (1-330 of 572), `internal/combat/simulate.go` (1-80 of 192), `internal/accounts/userrecord.go` (170-229, 334-341), `internal/characters/character.go` (1078-1117, 1235-1274, 1362-1487), `internal/stats/stats.go` (55-88), `internal/util/util.go` (160-194, 520-544), `internal/usercommands/admin.mob.go` (50-164), `internal/usercommands/go.go` (255-294), `internal/hooks/NewRound_IdleMobs.go` (25-104), `modules/gmcp/gmcp.Char.go` (440-481), `internal/actors/controller.go` (38-74, via grep -A), `internal/rooms/rooms.go` (262-317, 1190-1235, 1463-1503).

Grep only (no output sites found): `internal/combat/armor_rank.go`, `mob_rank.go`, `weapon_rank.go`, rest of `simulate.go`, rest of `calculations.go` (grep `SendTo|SendText|Tell\(|events\.Message` over `internal/combat` returned only combat.go and attackresult.go hits, so the pattern could match).

## 0. Facts that shape everything below

| Fact | Where |
|---|---|
| `messaging.Send` has **no production caller** yet; only `messaging_test.go` calls it | grep `messaging\.Send\(` |
| `OnBehaviourPerceive` has no production handler | grep; `internal/actors/controller.go:65` |
| `ConnectionController.Perceive` drops empty text, else `account.SendText(p.Text)` | `internal/actors/controller.go:38-43` |
| `Account.SendText` queues `events.Message{UserId, Text: txt+"\n"}`; so messaging output is still queued, one event per receiver | `internal/accounts/userrecord.go:334-341` |
| `Room.SendText(txt, excludeUserIds...)` queues one `events.Message{RoomId,...}`; receivers are resolved **at delivery** from `room.GetPlayers()` (players only, never NPCs) | `internal/rooms/rooms.go:274-283`; `internal/hooks/Message_SendMessages.go:38-85` |
| messaging resolves receivers **at Send time** from `r.GetBodies()` (players and NPCs) | `internal/messaging/messaging.go:157` |
| `Room.SendTextToExits` prefixes `(From <exit>) ` and has `isQuiet` | `rooms.go:285-317`. No combat-path site calls it (grep of the six files and `combat/`) |
| An actor always gets `ActorText` (no sense check); target gets `TargetText`, else `ObserverText` | `messaging.go:38-50, 83-95` |
| A message needs a non-zero `Senses` or it reaches only the actor | `messaging.go:34-36, 57-61`; test `messaging_test.go:495-506` |
| `canSee` checks `hidden` on the subject only; `see-hidden` is not consulted | `internal/messaging/perceive.go:44-55`; `see-hidden` is honoured by `rooms/roomdetails.go:258,276` |
| "Someone" capitalises only when the text **starts** with `{actor}`/`{target}` | `internal/messaging/render.go:38-43` |
| `AttackResult.MessagesToRoomOld` / `SendToRoomOld`: no reader or writer outside `attackresult.go` | grep over repo |

## 1. Every text output in the combat path

Legend for mapping: K = Kind (W = WorldEvent, I = Interface), S = Senses, R = Range, A = Audience. Text slots: AT ActorText, TT TargetText, OT ObserverText, TRT TargetRoomText, FT FarText.

### 1a. `internal/combat/combat.go` (texts built into `AttackResult`, sent by the round)

| Site | Builds | Today's receivers | Mapping | Fit |
|---|---|---|---|---|
| `combat.go:121` `SendToSource(toAttackerMsg)` (wait/"prepare" message, `GetWaitMessages`) | attacker | attacker | AT | ok |
| `combat.go:127` `SendToTarget` (wait) | defender | defender, **only if attacker not `hidden`** (`:124`) | TT | hidden rule, see section 2 |
| `combat.go:131` `SendToSourceRoom` (wait) | attacker's room | attacker's room, not if hidden | OT | hidden rule |
| `combat.go:136` `SendToTargetRoom` (wait), only when rooms differ (`:134`) | defender's room | not if hidden | TRT | hidden rule |
| `combat.go:311` dodge: `"%s dodges your attack!"` with raw `targetChar.Name`, no ansi name colour | attacker | attacker | AT | ok, but no observer text exists: must send OT empty (observers get a Perception with `Text==""`; connections drop it, a behaviour still receives it), or use `Audience: Named{target}`. Not suppressed for a hidden attacker |
| `combat.go:312` dodge: `You dodge the attack!` | defender | defender | TT | ok |
| `combat.go:360` per-attack `toAttackerMsg` (hit or miss; one per attack, several per round for multi-attack) | attacker | attacker | AT | ok; one `Message` per attack |
| `combat.go:368` per-attack `defenderMsg` + ` [you blocked N]` suffix (`:364-366`) | defender | defender | TT | ok (suffix is target-only text, fits TT) |
| `combat.go:371-374` per-attack `toAttackerRoomMsg` | attacker's room | attacker's room | OT | ok |
| `combat.go:377-381` per-attack `toDefenderRoomMsg`, only when non-empty (rooms differ; together case passes `nil` for that slot, `:338`) | defender's room | defender's room | TRT | ok |

Text construction facts that matter for tokens:

- Names are baked in by `buildCombatMessages` (`combat.go:149-221`): `{source}`/`{target}` are the combat-message YAML tokens (`internal/items/itemspec.go:168,170`), wrapped in the YAML as `<ansi fg="{sourcetype}">{source}</ansi>` (`_datafiles/world/empty/combat-messages/generic.yaml`, header and `prepare` block). **The items token `{target}` is the same string as messaging's `{target}`.** Handing a combat text to messaging with `{target}` unreplaced would let messaging fill it (with its own ansi wrap); the YAML's own `<ansi fg="{targettype}">` wrapper would then nest around it.
- If the slice wants "Someone" for an unseen attacker, `{source}` must become `{actor}`; but the text then starts with `<ansi fg=...>` (and crits start with `<ansi fg="yellow-bold">***</ansi> `, `combat.go:343-348`; backstab with `*[BACKSTAB]*`, `:352-357`), so `render.go:39` will not capitalise "someone". The spec note ("start a text with `{actor}`") cannot hold for these texts without stripping the YAML's ansi wrapper.
- Mob names use `GetMobName(ids.ID{})` with a zero viewer (`combat.go:170,174`): not per receiver today.
- `combat.go:372-373, 379` re-apply `TokenTarget`/`TokenTargetType` after `buildCombatMessages` already replaced every token (`:211-218`): a no-op.
- Message choice is random: `MessageOptions.Get(seed)` picks `mo[util.Rand(ct)]` when seed is 0 (`internal/items/attack_messages.go:45-58`). `ConsistentAttackMessages: true` (`_datafiles/config.yaml:316`) seeds by `weapon.ItemId`, but an unarmed attacker has `ItemId 0` (`calculations.go:76-78`), so **a bot's and a rat's messages are random every attack**. Intensity also follows the damage roll (`attack_messages.go:99-112`).

### 1b. `internal/hooks/NewRound_DoCombat.go`

Player round (`handlePlayerCombat`, `:40-520`):

| Site | Text | Receivers | Mapping | Fit |
|---|---|---|---|---|
| `:122` | `<mobname> blocks you from fleeing!` | fleeing user | AT of one flee-blocked message (Actor=user, Target=blocker) | see `:123` |
| `:123` | `<user> is blocked from fleeing by <mob>!` | user's room, excl. user | OT | ok for a mob blocker (target mob gets a Perception with OT; no connection) |
| `:128` | `<player> blocks you from fleeing!` | fleeing user | AT | see `:129` |
| `:129` | `<user> is blocked from fleeing by <player>!` | room, excl. user **and the blocking player** | OT | **does not fit**: the blocker is the Target and an empty TT falls back to OT (`messaging.go:87-88, 94`), so the blocker would start receiving this line. Needs a TT of its own, or a way to give the target nothing |
| `:137` | `You can't find an exit!` | user | I or W ActorOnly | ok |
| `:141` | `You flee to the <exit> exit!` | user | AT | ok |
| `:142` | `<user> flees to the <exit> exit!` | room, excl. user | OT (sent before `MoveToRoom` at `:147`; both paths resolve the old room) | ok |
| `:154` | `mob.Command("flee")` for charmed mobs | (input, not text) | n/a | n/a |
| `:162` | `usercommands.Look(..., CmdSecretly)` | user | look output, slice 7 | out of slice |
| `:205`, `:387` | `Your target can't be found.` | attacker | I | ok |
| `:221`, `:403` | `Your rage subsides.` | attacker | I (or W ActorOnly) | ok |
| `:234-252` PvP wait: `MessagesToSource` to user, `MessagesToTarget` to defUser, `MessagesToSourceRoom` to uRoom **excl. user and defUser**, `MessagesToTargetRoom` to defRoom excl. both | | one Message (AT/TT/OT/TRT) | ok: actor and target are their own roles |
| `:259`, `:433` | `You can't seem to find your target.` (defender `hidden`) | attacker | I | ok |
| `:277-291` PvP attack: same four sends and exclusions as `:234-252` | | one Message per attack | ok |
| `:309-311` | three lines to defUser: `***`, `Your <item> breaks! ***`, `***` | defender | AT of an "item breaks" message with Actor=defender | three sends become one text with newlines, or three messages |
| `:313` | `The <item> <defUser> was carrying breaks!` | defRoom, **excl. defUser only** (the attacker sees it) | OT | ok (attacker is an observer of this message) |
| `:416-426` PvE wait: source to user; source room excl. user; target room excl. user. **No `MessagesToTarget` send** (mob) | | one Message | ok (rat gets a Perception, no text) |
| `:452-462` PvE attack: same as `:416-426`; `MessagesToTarget` dropped | | one Message per attack | ok |
| `:481, :483` | `defMob.Command("go <exit>")`, `GetAngryCommand()` (only when the mob is in another room) | input | n/a | n/a |

NPC round (`handleMobCombat`, `:522-897`):

| Site | Text | Receivers | Mapping | Fit |
|---|---|---|---|---|
| `:587` | `mob.Command("say I have a CombatAction that is too long...")` | input → `say` | n/a (say is slice 7) | n/a |
| `:578-592` | `mob.Command(combatAction)` | input | n/a | n/a |
| `:651` | `mob.Command("equip ...")` | input | n/a | n/a |
| `:664-674` mob→player wait: target to defUser; source room excl. defUser; target room excl. defUser. No source send (mob) | | one Message | ok |
| `:693-703` mob→player attack: same | | one Message per attack | ok |
| `:719-723` | offhand break, as `:309-313` | | as above | as above |
| `:794-800` mob→mob wait: source room, target room, **no exclusions** | | OT/TRT | ok |
| `:822-828` mob→mob attack: same | | | ok |
| `:850` | `The <item> <mobname> was carrying breaks!` (no exclusions) | defRoom | OT, Actor=defMob | ok |

`handleAffected` (`:899-947`) sends no text; it queues `user.Command("suicide")` (`:919`), `events.PlayerDrop` (`:923`, handler `internal/hooks/PlayerDrop_HandlePlayerDrop.go`, **not read**: unverified whether it prints), and `mob.Command("suicide")` (`:940`).

### 1c. Combat commands

| Site | Text | Receivers | Mapping | Fit |
|---|---|---|---|---|
| `attack.go:172` | `You attack the darkness!` | actor (`Tell`) | I | ok |
| `attack.go:192` | `<mob> is your friend!` | actor | I | ok |
| `attack.go:206`, `:256` | `partyUser.Command("attack <id>")` | input | n/a | n/a |
| `attack.go:220` | `You prepare to fight <mobname>!` | actor | AT | ok |
| `attack.go:223-226` | `<actor> prepares to fight <mob>.` | room excl. actor; **not sent when actor `hidden`** (`:222`) | OT | hidden rule |
| `attack.go:236` | `room.CanPvp` error text | actor | I | ok |
| `attack.go:243` | `<player> is in your party!` | actor | I | ok |
| `attack.go:268` | `You prepare to fight <ansi fg="mobname">player</ansi>!` (player target painted `mobname`: upstream colour bug) | actor | AT | ok; the bug is visible text, a parity point if fixed |
| `attack.go:272-274` | `<actor> prepares to fight you!` | target player; not if hidden | TT | hidden rule |
| `attack.go:276-278` | `<actor> prepares to fight <player>.` | room excl. actor and target; not if hidden | OT | ok (two exclusions are exactly Actor and Target) |
| `flee.go:13` | `You attempt to flee...` | actor | I or AT (no observer text today) | ok |
| `shoot.go:22, 34, 48, 56, 62, 82, 110` | refusals / syntax | actor | I | ok |
| `shoot.go:88-90`, `:118-120` | `You prepare to shoot at <x> through the <exit> exit.` | actor | AT | ok |
| `shoot.go:93-96`, `:124-127` | `<actor> prepares to shoot at <x> through the <exit> exit.` (actor always painted `username`, even an NPC) | **actor's room only**, excl. actor; not if hidden. The target in the next room gets nothing | OT | **does not fit as is**: with a Target set, `Room` range covers the target's room too (`messaging.go:113-135`), so the target would get OT and its room's observers OT/TRT. Either leave Target unset, or the slice decides the target's room now perceives it |
| `consider.go:32, 34, 77-82` | consider text | actor | I | ok (interface, no world event) |

### 1d. Death text

| Site | Text | Receivers | Mapping | Fit |
|---|---|---|---|---|
| `suicide.go:40` | `You're already dead!` | user | I | ok |
| `suicide.go:45-48` | `<user> has died.` | room excl. user | OT, Actor=user, AT empty | ok |
| `suicide.go:83-90` | `*** <user> has DIED! *** (killed by a, b and c)` + `term.CRLFStr`, as `events.Broadcast` | every active user incl. the dead one (`Broadcast_SendToAll.go:31`), no sense check, no `\n` added | W, R World, AT=OT=FT | **does not fit cleanly**: (1) messaging goes through `SendText`, which appends `"\n"` after the CRLF the text already carries (blank line); (2) `Broadcast` has `TextScreenReader`/`SkipLineRefresh` fields with no messaging equivalent; (3) the "killed by" list is several players, which is text, not a role. The smoke script waits on this line (`script.go:207`) |
| `suicide.go:126` | `user.EventLog.Add(...)` | log, not output | n/a | n/a |
| `suicide.go:129-130` | permadeath template, `user.SendText` | user | I | ok |
| `suicide.go:137, 140, 186, 195, 208, 219, 230` | `Remove(...)` / `Drop(...)` command calls: their own texts are slice 7 | | n/a | flag: these print through the drop/remove commands |
| `suicide.go:247, 263` | `You lost N experience points.` | user | I | ok |
| `suicide_npc.go:33` | `<mob> is suddenly revived in a shower of sparks!` | room, no exclusion | OT, Actor=mob | ok |
| `suicide_npc.go:50-52` | `<mob> has died.` | room, **no exclusion** (the killer sees it too) | OT, Actor=mob | ok |
| `suicide_npc.go:140` → `accounts.GrantXP` → `userrecord.go:181` or `:187` | `You gained N experience points! (combat)` (or with `(%d%% scale)`) | each killer, own text | I (one per user) | ok; N is random (below) |
| `suicide_npc.go:154`, `:250` | `Your alignment has shifted from X to Y!` | user | I | ok |
| `suicide_npc.go:185, 187, 282, 284` | tame-learning lines | user | I | ok |
| `suicide_npc.go:236` | party share via `GrantXP` | each member | I | **sends to a party**: one Interface message per member |
| `suicide_npc.go:309-310`, `:340-341` | `<item> drops to the ground.` | room, no exclusion | W, Room, Room=room.RoomId, no actor (or Actor=dead mob) | ok |
| `suicide_npc.go:356-357` | `<N gold> drops to the ground.` | room | same | ok |

GrantXP also level-up text (`userrecord.go:201+`, not read past `:229`): unverified.

## 2. Hidden / sneak rules in the combat path

| Rule | Site |
|---|---|
| Wait/"prepare" messages: attacker `hidden` suppresses target, attacker-room and target-room text; attacker still gets its own | `internal/combat/combat.go:124-140` |
| **Per-attack hit/miss/dodge messages are never suppressed for a hidden attacker** (no `hidden` check in `calculateCombat`) | `combat.go:223-392` (grep: only `:124`) |
| `attack`: no room "prepares to fight" (mob or player) and no "prepares to fight you!" to a player target when the actor is hidden | `internal/usercommands/attack.go:176, 222, 270` |
| `shoot`: no room "prepares to shoot" when hidden | `internal/usercommands/shoot.go:65, 92, 122` |
| Defender hidden: player attacker gets "You can't seem to find your target." and skips the attack | `NewRound_DoCombat.go:258-261` (PvP), `:432-435` (PvE) |
| Defender hidden: NPC attacker silently skips | `NewRound_DoCombat.go:627-629` (vs player), `:806-808` (vs NPC) |
| Attacker hidden is **not** checked in either round loop | grep `hidden` in `NewRound_DoCombat.go`: only the four defender checks |
| Combat cancels `cancel-on-combat` buffs on attacker and defender each round | `NewRound_DoCombat.go:60, 218, 400, 559, 620, 779` |
| messaging renders an unseen (hidden) subject as "someone" and ignores `see-hidden` | `internal/messaging/perceive.go:51`; `render.go:16-18` |
| No `hidden` buff exists in `empty`: its buffs are `0-meditating`, `1-illumination`, `24-death_recovery` (3 files, flags `lightsource`, `cancel-on-action`, `cancel-on-combat`); `buffs-flags/hidden.yaml` defines the flag only | `ls _datafiles/world/empty/buffs`, grep of `flags` |
| `BackStab` aggro type is read (`combat.go:254`) but nothing sets it | grep `BackStab` |

So today, a hidden attacker hides only its opening line and wait lines; every swing still names it to the target and room. Through messaging with `{actor}` tokens, the swing lines would read "someone"/"Someone" (an intended difference per the spec's parity paragraph), and the opening lines would either keep being suppressed (code keeps the `if`) or start arriving as "someone ...".

## 3. Smoke and parity tooling

### What `pre-4b/bot0.txt` covers

399 lines, one bot, recorded on master `c4314b7f` (spec line 229). It contains **no combat and no NPC action by the bot**. In order: splash (1-18), `new` account signup (19-44), the `online`-style who table and `inbox`/`mudletmap not recognized` (45-53), the Void and `start` (55-77), race `human` and name (77-100), Town Square with `Also here: guard` and the ring (105-119), get/inventory/drop/look/get/inventory (120-183), `north` to End of the Line (`Also here: rat, rat and rat`, 184-201), `south` (202-219), `say` (220-223), `quit` with an NPC line `guard leaves towards the north exit.` (224-228), the logout art (229-245), relogin (247-291), `online`, `inventory` (292-318), `suicide` → `*** Smk89146fc0 has DIED! ***` and `HP:1/6 MP:0/5` (319-322), Waiting room (323-342), `portal` back with `On the Ground: Smk89146fc0 corpse` (343-360), `inventory`, `quit` with `guard enters from the north.` and two prompts on one line (`[...4:07AM...]:[...4:09AM HP:2/6 MP:1/5]:`, 382), logout art.

The only death text in it is the player's own Broadcast line. It has no copyover lines (Windows run). It predates slice 4's `npcSpeaks` (`cmd/smoke/admin.go:53, 67-71`, commit `4a5e7106`), which now leaves an extra rat in Town Square, so `Also here:` lines in Town Square differ on today's master whenever that rat has not wandered off.

### Recording

- Flag `-transcript <dir>` writes `bot<N>.txt` per bot (`cmd/smoke/main.go:44, 96-101, 136-146`).
- The recorder stores server text after `telnet.StripAnsi` and `\r` removal, and each sent line as `"\n> " + line + "\n"` (`client.go:79-82`, `recorder.go:16-32`). The admin connection records nothing (`admin.go:76`: `dial(..., nil)`).
- File on disk: index LF, working tree CRLF (`git ls-files --eol`: `i/lf w/crlf attr/`; `core.autocrlf=true`; `.gitattributes` covers only `*.go`, `go.mod`, `go.sum`). A fresh recording is LF (`\r` stripped at `client.go:79`), the checked-out baseline is CRLF, so a byte compare fails on every line unless `\r` is stripped or `cmd/smoke/testdata/** text eol=lf` is added.
- No comparison tool exists (grep for `pre-4b`/`testdata` in `cmd`/`internal` Go code: nothing in smoke).

### What a comparison must mask (observed in bot0.txt and the code)

| Varies | Example | Source |
|---|---|---|
| Run tag | `smk89146fu0`, `smoke89146f`, `Smk89146fc0` | `main.go:102, 127-131, 197-201` |
| Prompt line (game clock, HP/MP), including two prompts glued on one line | `[☀️ 4:00AM HP:6/6 MP:5/5]:` (76), (382) | prompt redraw per message, `Message_SendMessages.go:33, 82` |
| Who-table column widths follow the name length | lines 46-50 vs 272-276 | |
| `nameless-<n>` | line 49 | |
| Wandering NPC lines and `Also here:` contents | 113, 197, 226, 380 | rat `wander`, `maxwander 8`; guard `maxwander 20` (`1-rat.yaml:5,10`; `2-guard.yaml`) |
| Echoed command text (server echo) appears after `> cmd` | 78, 90 | |
| Text/prompt interleaving | line 100 (prompt before "You will be known as") | queue timing |

For a combat transcript add: the attack message variant (random pick, unarmed seed 0), the hit/miss/crit/dodge outcome, damage numbers and `[you blocked N]` (none vs a rat: rat defence 0, `applyDefenseReduction` rolls `util.Rand(0)` = 0, `calculations.go:286-293`, `util.go:173-178`; unverified that the rat wears nothing with defence), the number of rounds (so the number of combat lines), the XP amount, and `rat corpse` in later `On the Ground:` lines.

### Seeding `util.Rand`

- `util.Rand` calls the global `math/rand.Intn` (`internal/util/util.go:12, 173-178`); `RollDice` uses it (`util.go:527-541`). There is **no seed hook**: grep for `rand.Seed|rand.New(` outside tests finds only `internal/gametime/gametimedata.go:185` and `internal/procedural/gridmaze.go:84,93`, which use their own sources.
- Toolchain is Go 1.27.3 (`go.mod:3`, `go version`). Since Go 1.24 the top-level `rand.Seed` is a no-op unless `GODEBUG=randseednop=0` (from Go release notes; **unverified in 1.27**). Seeding therefore needs a code seam (for example `util` holding a `*rand.Rand` that a flag or env var can seed), and even then a live server's roll order depends on goroutine and round timing (bots' send timing, idle NPC rolls in `NewRound_IdleMobs.go`, `HandleRespawns`), so a seed alone does not make a transcript byte-identical. Masking is the practical route; a seam is useful for unit tests of `combat.Attack`.
- Rolls in one attack: weapon drop for dual wield (`combat.go:247`, `calculations.go:276`), hit (`calculations.go:126`), dodge (`:255`), damage dice (`combat.go:316`), crit (`calculations.go:207`), defence (`:287`), message pick (`attack_messages.go:50`). Flee (`NewRound_DoCombat.go:85,109`), NPC combat commands (`:576,578`), XP variation (`suicide_npc.go:83`), tame (`:178`), item drop (`:329`).

### A combat baseline

It does not exist: the pre-4b recording has no fight. It has to be recorded on current master before any slice 5 change, with the same smoke fight the Done-when names, so the smoke script change (fight step, observer bot) lands first, on its own commit, and the baseline is recorded from that commit. Bot 0's file covers the attacker text; bot 1's file covers the observer text, which is the Done-when check.

## 4. The Done-when smoke fight

**Command.** `attack <name>` (`usercommands.go:133`; `attack.go:160-161`: `room.FindByName(rest)`, mob preferred over player at `:186`). `attack` with no argument only finds a mob already fighting you (`:45-62`). A body ID token is accepted (`rooms.go:1482-1491` for players; the mob side, `findMobByName`, not read), but bots do not know IDs (the client keeps only GMCP package names, `client.go:87-90`).

**The rat does not fight back.** No code sets an NPC's aggro onto a player: grep of `SetAggro(` outside `attack.go`/`NewRound_DoCombat.go` finds only `character.go`, `combat.go:258` (backstab reset) and `simulate.go`; `NewRound_DoCombat.go:473-491` sets only `PreventIdle` (and a `go` when in another room); `go.go:284` prints "notices you as you enter!" and nothing else; `PATCH_NOTES.md:24` "mobs no longer start or join fights until Phase 5". So the fight is one-sided: the bot cannot be hurt or killed, the rat has no `combatcommands` (`1-rat.yaml`) and never flees, and the mob half of the round never runs for it (its `Aggro` stays nil, `NewRound_DoCombat.go:532`). The guard's `hates: rats` has no caller (`HatesMob`/`HatesRace` unused outside `mobs.go`; `HatesAlignment` only in `modules/follow/follow.go:511`).

**How it ends.** The player round applies damage (`combat.go:38`), `handleAffected` queues `mob.Command("suicide")` when `Health < 1` (`NewRound_DoCombat.go:937-941`), and `npcDeath` (`suicide_npc.go:24`) prints `rat has died.` to the whole room (`:50-52`), queues `MobDeath` (`:61`), clears each damaging player's aggro (`:95-99`), grants XP (`:140`), may shift alignment (`:150-160`), drops nothing for a stock rat (no items or gold in `1-rat.yaml`; worn items: none, rodent slots are mostly disabled, `10-rodent.yaml:11`), and leaves a corpse (`CorpsesEnabled: true`, `config.yaml:260`; `suicide_npc.go:376-387`). `TrackPlayerDamage` runs on every attack, even a miss (`combat.go:41`), so the bot is always in `PlayerDamage`.

**XP text.** `mobXP = XPTL(0)` → level clamps to 1 → `(1000 + 1^2*0.75*1000)*1 = 1750` (`character.go:1247-1258`, `config.yaml:418-425`; rat TNLScale 1, `10-rodent.yaml:7`); `xpVal = 1750/90 = 19`; variation `max(19/100,1) = 1`; `±1` from `util.Rand(3)-1` → 18, 19 or 20; level delta 0 → scaler 1.0 (`suicide_npc.go:71-128`). So `You gained 18|19|20 experience points! (combat)` (`userrecord.go:187`, assuming `xpScale == 100`; unverified). 19 XP of 1750 needed: no level-up.

**Length (unverified estimate).** Rat HP: `HPBase 5 + level*1 + Vitality.ValueAdj*4` (`character.go:1406-1409`, `config.yaml:399-401`), floor 1 (`:1429`); the rat's vitality is `training: -2` plus race base 0 plus spawn-time training I did not trace (`NewMobById` not read; `simulate.go:54-66` auto-trains stat points), so somewhere from 1 to about 10 HP. The bot: human, fists 1d3 (`1-human.yaml:26`), plus `Speed/50` dice, `Strength/12` sides, `Perception/25` bonus (`character.go:366-368`), at level 1 most likely plain 1d3; hit chance `floor(atk/(atk+def)*100)` floored at 25, 50% at equal speed (`calculations.go:53-63, 99-108`); dodge 5-30% (`config.yaml:333-334`); crit 5-30%. Expected about 1 HP per round; rounds are 4 s (`RoundSeconds: 4`, `config.yaml:164`). A fight likely takes 1-10 rounds (4-40 s) with a long tail. The default step timeout is 30 s (`main.go:40`), so the fight step needs `doFor` with a longer limit (`script.go:36-41`), perhaps 2 minutes. Measuring the rat's HP once (admin `mob` info or a live attack) would replace this estimate.

**What the second bot sees today** (same room, current code):
1. `Smk<tag>c0 prepares to fight rat.` (`attack.go:223-226`, room excl. attacker).
2. One line per attack per round from the combat YAML `toroom` list (`NewRound_DoCombat.go:456-458`, excl. attacker), e.g. a miss or hit variant naming both.
3. `rat has died.` (`suicide_npc.go:50`).
It does not see the attacker's `You prepare to fight rat!`, the `toattacker` lines, or the XP line. Bot 0 sees its own lines plus `rat has died.` (no exclusion) and the XP line. So the Done-when check is: bot 1's transcript holds the `toroom` variants and none of the `toattacker` variants.

**Admin spawn.** `mob spawn <name|id>` (permission `mob.spawn`, `admin.mob.go:56-63`) → `mobs.NewMobById(mobId, room.RoomId)` in the admin's room, admin gets `You wave your hands around and rat appears in the air and falls to the ground.` and the room gets `<admin> waves their hands around and rat appears in the air ...` (`admin.mob.go:139-157`). The smoke admin already does this once, before the bots log in, and leaves the rat in Town Square (`admin.go:53, 67-71`).

**Flakiness sources.**

| Risk | Detail |
|---|---|
| Several rats named `rat` | Three spawn in room 2 (`rooms/startland/2.yaml:11,14,17`), one is the admin's; `attack rat` hits whichever `findMobByName` returns first. The kill is still a kill, but which rat, and whether one is in the room, varies |
| The rat wanders off before the attack | Rat idles at `activitylevel 10`, one of three idle commands is `wander` (`1-rat.yaml:8-12`), `maxwander 8`. The empty `startland` has two rooms, so rats drift between them (`script.go:179-180`). If none is in Town Square, `You attack the darkness!` (`attack.go:172`) |
| Wander after the attack | Once the bot's aggro targets the rat, `inFight` refuses the rat's `go` ("another body in its room has the actor as its aggro target", `go.go:22-23, 328-345`), and `PreventIdle` is set each round (`NewRound_DoCombat.go:474`) and consumed by `IdleMobs` the same round (DoCombat registers before IdleMobs, `hooks.go:26,31`). A `wander` queued before the attack could still run between the attack command and the next round: unverified whether `inFight` sees the aggro then (it should, since `SetAggro` runs inside `attack`) |
| Target moved: `Your target can't be found.` and aggro cleared | `NewRound_DoCombat.go:386-396` |
| Misses / long fights | random, above |
| Bots dying | not possible from the rat (it never attacks) |
| Several bots | With `-bots N` every bot runs `play` concurrently (`main.go:126-153`); if all fight, they need N rats or share one (shared kill splits XP by `len(PlayerDamage)`, `suicide_npc.go:82`). The observer must be in Town Square for the whole fight, which needs a sync point like the copyover `WaitGroup` (`copyover.go:18-29`). Other bots' `say`, `get`/`drop` and walk lines land in each other's transcripts, so multi-bot transcripts are not comparable line for line |
| Other bots' rings | Rings sit on the Town Square floor (`admin.go:44-52`); the corpse adds `rat corpse` to `On the Ground:`; `waitField` checks only that the ring is in that line (`client.go:169-180`), so it still passes |
| Corpse later | `rat corpse` shows in every later Town Square `look` until it decays (decay not read) |

## 5. GMCP combat packages

- `Char.Enemies` (`modules/gmcp/gmcp.Char.go:444-481`) reads the user's `Aggro.Target`, the room's `GetMobs(rooms.FindFighting)`, and each mob's `Id`, `Name`, `Level`, `Health`, `HealthMax`. It is built only on a `Char*` request or an "all" build (`gmcp.go:257-268`; no listener asks for `Char.Enemies` by name, grep). A rat that never takes aggro is not "fighting" (FindFighting definition not read past `rooms.go:1190`; unverified), so today the bot's `Char.Enemies` is likely empty during the rat fight.
- `Char.Vitals` and the rest come from `CharacterVitalsChanged` (`gmcp.Char.go:45, 201`), queued by the round (`NewRound_DoCombat.go:338-342, 494, 749`) and death (`suicide.go:302`). `MobDeath`/`PlayerDeath` → kills (`gmcp.Char.go:57-58`). `GainExperience` (`:49`) from `GrantXP` (`userrecord.go:192`). `AlignmentChanged` (`:55`).
- `Room` and `Party` packages read `AggroChanged` and `Character.Aggro != nil` (`gmcp.Room.go:40, 87, 354, 382`; `gmcp.Party.go:38, 57, 223`).
- **None of these read `AttackResult` messages, `events.Message`, `events.Broadcast` or any text.** They depend on the events and state the round produces (aggro set/cleared, vitals, death, XP). The slice touches them only if the merged round changes which `AggroChanged`/`CharacterVitalsChanged` events fire or when: for example `:338-342` (PvP) queues `CharacterVitalsChanged` without `AggroChanged` on the continue branch (`:348`), while PvE queues both (`:503-504`). Those are outcome differences for the handler diff, not messaging.

## Open questions for Calabe

1. **Hidden attacker.** Today a hidden attacker hides only its opening and wait lines (`combat.go:124`, `attack.go:222,270`, `shoot.go:92,122`); its swings still name it (`combat.go:360-381`). Options: (a) keep today's suppression exactly and keep names baked in, so no "someone" in combat; (b) drop the suppression and use `{actor}` everywhere, so every line reaches the target and room as "someone ..."; (c) keep the suppression for opening lines and use `{actor}` for swings, so swings read "someone". (a) is pure parity and changes nothing visible; (b) unifies on messaging's rule but tells the room a fight started; (c) is closest to today's intent with the spec's "Someone" rule. Recommendation: (c). Hidden cannot be reached in `empty` anyway (no buff grants it), so this is a rule for later content.

2. **Combat YAML tokens vs messaging tokens.** The YAML wraps names as `<ansi fg="{sourcetype}">{source}</ansi>`, and items' `{target}` collides with messaging's `{target}`. Options: (a) combat keeps resolving names itself and hands messaging finished text (no "someone", parity exact); (b) combat maps `{source}`→`{actor}` and `{target}`→`{target}` and strips the YAML ansi wrapper so messaging colours names (texts that open with a name then capitalise "Someone"; crit/backstab prefixes still block it); (c) rewrite the combat-message YAML to messaging tokens. (a) defers the unseen rule for combat; (b) gives it now at the cost of a token-mapping step; (c) touches content. Recommendation: (b), with the parity mask covering the crit-prefixed lowercase "someone".

3. **Which baseline.** The pre-4b recording has no fight, and slices 2-4 changed Town Square's `Also here:` (the admin's rat). Options: (a) record a new baseline on master before slice 5 (fight step and observer included) and compare against it, keeping pre-4b as the slice 7 anchor; (b) compare only the non-combat part against pre-4b and the combat part against a fresh recording; (c) pre-4b only. Recommendation: (a), and record the proof-of-catch (a deliberate one-word change in a combat line must fail the compare) with it.

4. **Who fights in the smoke run.** Options: (a) bot 0 fights, bot 1 waits in the room as observer, bots 2+ skip the fight; (b) every bot fights its own rat; (c) only when `-bots >= 2`. (a) is the Done-when literally and keeps 50-bot runs cheap; (b) loads combat but needs N rats and shared-room sync. Recommendation: (a), with a sync point so the observer is in Town Square before the attack and stays until `rat has died.`.

5. **Which rat.** Every rat is named `rat` and wanders. Options: (a) admin spawns a fresh rat right before the attack and the bot attacks `rat`, accepting that it may hit another rat (any kill passes); (b) add a smoke-only test mob to `empty` (unique name, no idle commands, `activitylevel 0`) so the target and its HP are fixed; (c) admin uses an existing admin command to freeze or name the rat (none found). (b) makes the fight deterministic except for rolls and removes wander lines from the fight; (a) touches no content. Recommendation: (b), if adding test content to `empty` is acceptable; otherwise (a) with a retry when the room has no rat.

6. **Text that does not fit `Message` as built.** The flee-blocked-by-player line (`NewRound_DoCombat.go:129`) excludes the blocker, but an empty `TargetText` falls back to observer text; `shoot`'s opening line reaches only the shooter's room today; the death Broadcast (`suicide.go:83-90`) has a trailing CRLF and screen-reader/refresh fields. Options per case: (a) let the target and the target's room start receiving them (a visible change, listed as intended); (b) add a messaging option (for example a "target gets nothing" flag) in this slice; (c) leave these few sites on the old path for slice 7. Recommendation: (b) for the flee line (small, and slice 7 will meet the same shape), (a) for shoot (the target's room learning of a shot fits `Room` range with a target elsewhere, spec section 2), and keep the death Broadcast as a `World` message with the CRLF trimmed.

7. **Upstream text bugs met on the way** (`attack.go:268` paints a player target `mobname`; `shoot.go:94,125` paints an NPC shooter `username`; `combat.go:311` dodge names the target uncoloured). Options: fix in slice 5 (parity mask lists them as intended differences) or keep for parity and log in `docs/known-issues.md`. Recommendation: fix them, since `nameFor` colours by `Account()` and the fix is free once names go through tokens.
