# Phase 4b slice 4 inventory (2026-10-28, master 48a7d92b)

Three read-only inventories written for the slice 4 plan (`2026-10-28-phase-4b-4-one-command-set-plan.md`): A, mob commands against user commands; B, the user command surface; C, the two input paths. Each lists the files it read. Their open questions were settled by Calabe or by the plan's defaults (plan, "Decisions"). This file moves to `docs/completed/` with the plan.

# Inventory A: mob commands vs user commands (Phase 4b slice 4)

Master `48a7d92b`, read 2026-10-28. Read only; no repo file edited, no git state changed.

Read 64 of 64 files in full:
- `internal/mobcommands/`: mobcommands.go, broadcast.go, despawn.go, drop.go, emote.go, equip.go, get.go, give.go, go.go, look.go, noop.go, pathto.go, put.go, remove.go, rest.go, say.go, sayto.go, shout.go, suicide.go, unlock.go, wander.go (21)
- `internal/usercommands/`: usercommands.go, broadcast.go, drop.go, emote.go, equip.go, get.go, give.go, go.go, look.go, noop.go, put.go, remove.go, say.go, shout.go, suicide.go, unlock.go, whisper.go, admin.command.go (18)
- `internal/hooks/`: MobIdle_HandleIdleMobs.go, NewRound_IdleMobs.go, NewRound_MobRoundTick.go, NewRound_DoCombat.go, Buff_ApplyBuffs.go, Broadcast_SendToAll.go, NewTurn_ActionPoints.go (7)
- `world.go`, `internal/mobs/mobs.go`, `internal/characters/charminfo.go`, `internal/actors/actor.go` (4)
- `internal/plugins/plugins.go`, `internal/plugins/plugincallbacks.go`, `modules/all-modules.go`, `modules/cleanup/cleanup.go`, `modules/follow/follow.go`, `modules/gmcp/gmcp.Comm.go` (6)
- `2026-10-27-phase-4b-npcs-are-players-design.md` (1)
- Data: `_datafiles/world/empty/mobs/startland/1-rat.yaml`, `.../2-guard.yaml`, `races/0-ghostly_spirit.yaml`, `races/1-human.yaml`, `races/10-rodent.yaml`, `rooms/startland/1.yaml`, `templates/admincommands/help/command.command.template` (7)

Read in part only (the cited line ranges; not read in full): `internal/rooms/rooms.go` (261-300, 585-610, 720-770), `internal/rooms/roommanager.go` (265-400), `internal/characters/character.go` (236-243, 470-495, grep hits 63/77/1418-1433), `_datafiles/world/empty/keywords.yaml` (105-147), `_datafiles/config.yaml` (grep 187-197), `_datafiles/html/admin/docs-modules.html` (625-645, grep 367/1505/1554). Not read: `internal/mobcommands/README.md`, `internal/mobcommands/AGENTS.md` (non-Go, go with the package).

## Facts verified against source

| Fact | Value | Where |
|---|---|---|
| Mob registry | 21 entries, not 20: the spec's list plus `rest` (added for decision 15) | `mobcommands/mobcommands.go:24-46`; spec line 20 lists 20 |
| Mob signature | `MobCommand func(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error)`; no flags | `mobcommands.go:17` |
| User signature | `UserCommand func(rest, *accounts.Account, *rooms.Room, events.EventFlag) (bool, error)` | `usercommands.go:30` |
| Unregistered mob file | only `unlock.go` (`Unlock` at `unlock.go:12`); every other `.go` file in the package is in the map | `ls`, `mobcommands.go:24-46` |
| Runtime mob commands from plugins | `trash`, `bury` (cleanup), `follow` (follow); all four modules are compiled in | `modules/cleanup/cleanup.go:58-59`, `modules/follow/follow.go:62`, `modules/all-modules.go:9-12` |
| Mob dispatch | `world.go:640` -> `mobcommands.TryCommand`; unhandled -> `mob.Command("emote looks a little confused (%s %s).")` | `world.go:613-653` |
| Mob dispatch order | alias (twice) -> registry -> `Go` fallthrough (if not downed) -> emote aliases | `mobcommands.go:63, 87-104, 106-122, 124-136, 137-140` |
| User dispatch order | exit name -> `go`; per-user alias; global alias; cancel `cancel-on-action` buffs; `look` self-keywords; plaintext-password gate; registry (downed message, `InputBlocked`, admin gate); emote aliases; `Go` fallthrough | `usercommands.go:155-158, 161-195, 202, 206-217, 219-223, 225-260, 262-265, 273-275` |
| Emote alias maps | byte-identical in both packages (diff of `mobcommands/emote.go:12-92` vs `usercommands/emote.go:13-93`) | `diff` |
| `mob.Command` | splits on `;`, queues `events.Input{ActorId, InputText, ReadyTurn}` one turn apart after an optional delay | `mobs/mobs.go:279-310` |
| `IsCommunication` / `SourceIsMod` | set, read nowhere (only `rooms.go:268`, `usercommands/broadcast.go:24-25`, `events/eventtypes.go:64,77` mention them). So `SendText` vs `SendTextCommunication` has no behavioural difference today | grep whole repo |
| `events.Communication` | one listener: GMCP `Comm.Channel`. `say` goes to every player in the source's room regardless of `TargetUserId` | `modules/gmcp/gmcp.Comm.go:32, 69-85` |
| Action points | regenerate only for active users | `hooks/NewTurn_ActionPoints.go:23-28` |
| Mob movement | `room.RemoveMob` / `destRoom.AddMob`; `AddMob` fires `RoomChange` | `rooms.go:725-757` |
| Player movement | `rooms.MoveToRoom(userId, ...)`: death-room instancing, start-room alias, mutator buffs, visited marks, ephemeral reset room, `RoomChange` | `roommanager.go:269-396` |

## 1. Per-command diff

Legend for `AllowedWhenDowned` column: mob / user. "Downed" for a user also prints "You are unable to do that while downed." and honours `InputBlocked` (`usercommands.go:227-239`); a downed mob is silently handled (`mobcommands.go:107-110`).

| Mob command | User command (file) | Downed mob / user | Behavioural differences (mob side / user side) | Recommendation |
|---|---|---|---|---|
| `broadcast` | `broadcast` (`usercommands/broadcast.go`) | false / true (`mobcommands.go:25`, `usercommands.go:57`) | (a) Mute check user only (`u/broadcast.go:15-18`). (b) Mob fires no `events.Communication`; user fires `CommType broadcast` (`u/broadcast.go:28-33`), so mob broadcasts never reach GMCP. (c) Colours: mob `black-bold`/`mobname`/`yellow` (`m/broadcast.go:16`), user `broadcast-prefix`/`username`/`broadcast-body` (`u/broadcast.go:20`). (d) User sets `IsCommunication`, `SourceIsMod` (unread). | Take the user version. Mute is an account check (refuse-on-nil rule makes it a no-op for NPCs). Firing Communication for NPCs is a gain. Name colour (`mobname` vs `username`) is a presentation question; keep per-actor name colour via a helper. Downed flag: user `true` wins. **No producer** issues mob `broadcast` (section 3). |
| `despawn` | none | false / - | Logs, `rooms.NPCEnded(mob, true)`, `DestroyInstance`, `CleanupMobSpawns` on home room, `room.RemoveMob` (`m/despawn.go:9-28`). | Engine call (spec). **No producer** outside admin `command` (grep: only the comment `charminfo.go:16`). |
| `drop` | `drop` (`usercommands/drop.go`) | false / true | (a) Mob: `perma-gear` buff blocks item drops (gold still drops) (`m/drop.go:63-65`); user has no such check. (b) User cancels `hidden` (`u/drop.go:57, 85`); mob does not. (c) Gold: user rejects `> Gold` with message and fires `EquipmentChange{GoldChange}` (`u/drop.go:52-55, 62-65`); mob silently ignores over-balance and fires nothing (`m/drop.go:51-60`). (d) Mob adds item to floor before removing from inventory (`m/drop.go:82-83`), user removes then adds (`u/drop.go:103, 126`). (e) Grenade stub user only (`u/drop.go:120-124`). (f) All error texts user only. | Take user version; keep `perma-gear` as a general rule (it is a buff flag, so it applies to whoever has it), or justify dropping it: suicide also honours it for mob loot (`m/suicide.go:327`) and user death (`u/suicide.go:164`), so the flag means "gear does not leave this body". Recommend: keep, move into the shared `drop`. **No producer** in code or data. |
| `emote` | `emote` (`usercommands/emote.go`) | true / true | (a) Mob returns early when the room has no players (`m/emote.go:98-100`). (b) User: emote-alias key lookup (`u/emote.go:109-116`); mob gets the already-expanded text from dispatch (`mobcommands.go:137-140`), user gets the key (`usercommands.go:262-264`). (c) User mute check for free text (`u/emote.go:118-121`). (d) User `@` prefix suppresses self echo (`u/emote.go:123-127`). (e) User self echo "You Emote: ..."; mob none. (f) Name colour `mobname` vs `username`. | Take user version. Drop the no-players early return (messaging with no receivers is cheap; slice 6 behaviours may perceive emotes). All other differences are actor-only text or account checks. Most-used mob command: idle data, angry commands, hooks, `equip`/`remove`/`go` internals (section 3). Text parity: mob emote lines must still read `<name> <text>`. |
| `equip` | `equip` (`usercommands/equip.go`) | false / false | (a) Mob `perma-gear` refuses with `emote struggles with their gear...` (`m/equip.go:15-18`). (b) Mob accepts `all` and `random` (`m/equip.go:20-31, 36-41`); user neither. (c) Mob re-wearing the same item is a silent no-op (`m/equip.go:64-67`); user always messages. (d) Mob cancels `hidden` only on a real change (`m/equip.go:70`); user on any success (`u/equip.go:44`). (e) Room text: mob "puts on X" / "wields X" (`m/equip.go:85, 88`); user "puts on their X" / "wields their X" (`u/equip.go:71, 79`). (f) Mob `Validate()` vs user `Validate(true)` (`m/equip.go:91`, `u/equip.go:62`). (g) User stat-diff and failure messages (`u/equip.go:37, 64, 90-97`). (h) Mob does not fire `EquipmentChange` when re-wearing the same item. | Take user version. Keep `perma-gear` refusal as a general rule (same reasoning as drop). `all`/`random`: **no producer** (only `equip <DisplayName>` at `NewRound_DoCombat.go:638`); drop unless Calabe wants them as player features. Keep (c) the same-item no-op (it prevents room spam from blind retries; DoCombat retries at random when weaponless, `NewRound_DoCombat.go:622-641`). Room text: the user wording ("their") becomes the one wording; a text-parity change for NPC lines, to list. Unverified: whether `FindInBackpack(itm.DisplayName())` matches when DisplayName carries ANSI or adjectives (`NewRound_DoCombat.go:633, 638`). |
| `get` | `get` (`usercommands/get.go`) | false / false | (a) Mob has no container, corpse or noun handling (`u/get.go:64-84, 88-222, 224-296, 397-404`). (b) **Mob ignores `StoreItem`'s result and removes the item from the room first** (`m/get.go:89-90`): a full backpack loses the item. User checks (`u/get.go:356-389`). (c) Mob `gold` exact; user accepts prefixes `g`/`go`/`gol` (`u/get.go:300`). (d) Gold: user fires `EquipmentChange` (`u/get.go:312-315`), mob none. (e) Stash: user auto-finds own stash and clears `StashedBy` (`u/get.go:332-340, 352-354`); mob does neither. (f) `exploding` adjective refusal user only (`u/get.go:344-347`). (g) Room text: mob "picks up <gold>N gold</gold>" (`m/get.go:51`) vs user "picks up some gold" (`u/get.go:321`); stash pickups user "digs around" (`u/get.go:372`). | Take user version; (b) is a mob bug and goes. Stash and corpse looting extend to NPCs, which is fine (no producer; nothing in data issues `get`). |
| `give` | `give` (`usercommands/give.go`) | false / false | (a) Gold parse: mob needs a space before `gold` (`m/give.go:33`); user `len>4 && ends "gold"`, then slices `[0:len-5]` (`u/give.go:34-36`). (b) Negative gold: user refuses (`u/give.go:39-42`); mob lets it through, then does nothing (`m/give.go:87, 133`). (c) Give to player: mob sends only the target text (`m/give.go:83-85, 92-94`); user sends giver, target and room texts (`u/give.go:76-85, 124-133`). (d) Gold: user fires two `EquipmentChange` (`u/give.go:114-122, 161-164`); mob none. (e) Self-give gold user only (`u/give.go:101-109`). (f) **User dereferences `targetUser` without a nil check** (`u/give.go:69-73`); mob checks (`m/give.go:61-64`). (g) User giving to an NPC makes the NPC emote (`u/give.go:200-204`); mob-to-mob does not. (h) Room text to NPC target uses `mobname` both sides. | Take user version, keeping the mob nil check (f). (g): after the fold an NPC giving to an NPC would trigger the receiver's emote, which is a new behaviour; recommend keep (one rule for all givers). **No producer.** |
| `go` | `go` (`usercommands/go.go`) | false / false | (a) **Aggro**: user refuses "You are in combat!" (`u/go.go:19-22`); mob has no check. Mobs that pursue a fleeing player have aggro and are told to move (`u/go.go:245-273`, `NewRound_DoCombat.go:477`). (b) **Action points**: user deducts 10, or 50 when encumbered (`u/go.go:45-62`); mobs never regain AP (`NewTurn_ActionPoints.go:23-28` loops users only). (c) Numeric room ID: mob may `go <roomId>` to an adjacent exit or teleport to any room ("runs off suddenly", `m/go.go:20-61`); user none. (d) `go home` -> `pathto home` mob only (`m/go.go:68-71`). (e) Locked exit: mob emotes "tries to go ... but it's locked" and stops (`m/go.go:74-79`); user tries lockpick sequence, key ring, backpack key (`u/go.go:68-144`). (f) Far side locked: mob refuses silently (`m/go.go:99-103`); user unlocks it (`u/go.go:167-170`). (g) `ExitMessage` requeue user only (`u/go.go:146-150`). (h) Sneaking: user hides leave/enter text (`u/go.go:180-206`); mob always announces even when `hidden`. (i) Mover: `RemoveMob`/`AddMob` vs `MoveToRoom` (mutator buffs, visited marks, death-room instancing). (j) After moving, user only: party follow, charmed NPCs follow, pursuing NPCs roll to follow, hostile NPCs "notice you", secret `look` (`u/go.go:208-313`). (k) Mob only: waypoint `noop` 1 s delay (`m/go.go:130-135`). (l) No exit: user "bumping into walls" for compass words (`u/go.go:319-342`); mob returns false, so dispatch tries emote aliases, then the "looks a little confused" emote. (m) Dest room nil: both return `(false, err)`. (n) Enter text: user wraps the already-tagged `enterFromExit` in a second `exit` tag (`u/go.go:201`), mob does not (`m/go.go:125`). | Needs rulings on (a), (b), (c), (e)/(f), (h). See open questions 1-3. Keep (k) as a behaviour-side rule (the behaviour that issued `go` adds the delay, not `go`). (d): move to the behaviour (`pathto` already handles `home`). (i), (j): take the user path for all actors; (j) makes NPCs lead charmed followers and pursuers too, which is the actor model's intent. (l): take user ("bumping into walls" room text for NPCs too). (n) cosmetic, take mob form. Producers: section 3 (wander, path steps, follow module, flee, pursuit). |
| `look` | `look` (`usercommands/look.go`) | false / true | (a) Secret look: mob `secretly` text prefix (`m/look.go:15-19`); user `events.CmdSecretly` flag (`u/look.go:23`). (b) Order: mob exits, backpack, people, body, nouns (`m/look.go:56-158`); user people, containers, exits (+direction aliases), backpack/body, nouns, corpses (`u/look.go:70-387`). (c) Darkness/nightvision user only (`u/look.go:25-32, 222-232`). (d) Mob sends no text to itself; user sends panels and the room view (`u/look.go:95-102, 119-126`, `lookRoom` 436-601). (e) Mob locked exit silent; user message (`u/look.go:235-238`). (f) Mob noun text uses `username` colour for the mob (`m/look.go:154`). (g) User renders noun highlighting from `HasRolePermission` (`u/look.go:310`), an account call. (h) User `look` self-keywords rewrite in dispatch (`usercommands.go:206-217`). | Take user version; actor-only text drops for a `BehaviourController` once it goes through messaging (slice 7). Until then the folded `look` writes with `user.SendText`, so the plan must decide what an NPC actor's send does in slice 4 (see open question 5). Drop the `secretly` prefix (no producer; the flag replaces it). **No producer** issues mob `look`. |
| `noop` | `noop` (`usercommands/noop.go`) | true / true | Identical (`m/noop.go:9-11`, `u/noop.go:10-12`). | Spec makes it a behaviour action. But `mob.Sleep` issues `noop` with a delay (`mobs.go:275-277`) and `go` issues it at waypoints (`m/go.go:134`): the delay is the point, and the input queue (`world.go:101-118`) is what blocks the actor. A shared `noop` command already exists and is identical, so recommend keep `noop` as the one user command and have behaviours issue it, rather than inventing a separate wait. `wake` aliases to `noop` (`keywords.yaml:125`). |
| `pathto` | none | false / - | See section 2. | Behaviour action (spec). |
| `put` | `put` (`usercommands/put.go`) | false / false | (a) **Mob gold is not checked against balance and not subtracted** (`m/put.go:73-76`): a mob can mint gold into a container. User checks and subtracts (`u/put.go:85-98`). (b) Locked container: user refuses (`u/put.go:47-52`); mob ignores the lock. (c) User fires `EquipmentChange` (`u/put.go:93-96`; note it passes `+goldAmt`, sign differs from drop/get). (d) Self texts user only. | Take user version; (a) and (b) are mob bugs. **No producer.** |
| `remove` | `remove` (`usercommands/remove.go`) | false / false | (a) Mob `perma-gear` refusal emote (`m/remove.go:14-17`). (b) User skips remove-locked items in `all` and refuses them singly (`u/remove.go:18-20, 40-45`); mob ignores `IsRemoveLocked`. (c) Cursed items: user refuses unless max `enchant` (`u/remove.go:47-59`); mob ignores curses. (d) Mob fires `EquipmentChange` even when `RemoveFromBody` fails (`m/remove.go:50-55`); user only on success (`u/remove.go:81-84`). (e) User cancels `hidden` before trying (`u/remove.go:61`), mob only on success (`m/remove.go:41`). (f) Stat diff and texts user only. | Take user version; keep `perma-gear` as for equip/drop. **No producer.** |
| `replyto` | none (nearest: `whisper`, `usercommands/whisper.go`) | true / - | Player target: "replies to you" to target, room text excluding target, sneak form "someone replies" to target only (`m/sayto.go:138-154`). NPC target: room text unless sneaking (`m/sayto.go:155-168`). No `events.Communication`. `whisper` is different: global by character name, players only, mute rule, `CommType whisper` (`u/whisper.go:15-55`). | **No producer** in code or data. Needs ruling (open question 4). Recommend drop. |
| `rest` | none | false / - | Logs; ephemeral room -> `EndEphemeralNPC` + home `CleanupMobSpawns`; else `room.RestNPC(mob)` (`m/rest.go:14-30`). Not in the spec's list (added by slice 2b, decision 15). | Engine call, like `despawn`. Producers: boredom (`NewRound_IdleMobs.go:47`) and `CharmExpiredRest` (`charminfo.go:7`, which nothing sets: `Charm(` has no non-test caller, `character.go:480` is the only definition). Needs ruling only on whether charm-expiry text keeps a `rest` command form (open question 6). |
| `say` | `say` (`usercommands/say.go`) | true / true | (a) Mob returns early with no players (`m/say.go:14-17`). (b) Mute and drunk user only (`u/say.go:15-26`). (c) Colour `saytext-mob` vs `saytext` (`m/say.go:22,24`, `u/say.go:29,31`). (d) Self echo user only. (e) Event order differs (Communication before vs after exits text); same event fields. | Take user version; drop (a). Keep a per-actor say colour if text parity matters (`saytext-mob` is an ANSI alias; check `ansi-aliases.yaml`, unverified whether it differs from `saytext`). Producers: guard spawn idle (`rooms/startland/1.yaml:16`), `NewRound_DoCombat.go:574`. |
| `sayto` | none | true / - | Player target: "says to you", room text excluding target, sneak form to target only; Communication `say` (sneak form sets `TargetUserId`, but GMCP sends `say` to the whole room anyway, `gmcp.Comm.go:69-85`, so a hidden NPC's name leaks over GMCP) (`m/sayto.go:14-83`). NPC target: room text plus Communication unless sneaking. | **No producer.** Needs ruling (open question 4). |
| `saytoonly` | none | true / - | Target-only text plus Communication `say` with `TargetUserId` (`m/sayto.go:85-124`). GMCP still sends it to the whole room (`gmcp.Comm.go:69-85`), so "only" does not hold on GMCP. NPC targets ignored. | **No producer.** Needs ruling (open question 4). |
| `shout` | `shout` (`usercommands/shout.go`) | true / true | (a) Mob text not upper-cased (`u/shout.go:22`). (b) Mute, drunk user only. (c) Far text: mob "Someone is shouting from the X direction." with no words, `Exits` only (`m/shout.go:20-26`); user includes the words and covers `Exits`, `ExitsTemp` and mutator exits (`u/shout.go:35-63`). (d) Colour `saytext-mob` vs `yellow`. (e) Self echo user only. | Take user version (the spec's `FarText` already follows it, spec line 185-186). Text-parity change for NPC shouts (human race angry commands, `races/1-human.yaml:12-14`). |
| `suicide` | `suicide` (`usercommands/suicide.go`) | true / true | Different functions with one name. Mob: `revive-on-death` buff, `vanish` form, recent-death tracking, charm cleanup, "The Guide" (MobId 38) special case, `MobDeath` event, XP split, party XP, alignment, tame chance, KD kills, loot to floor or corpse by `ItemDropChance`, `NPCEnded`, `DestroyInstance`, corpse (`m/suicide.go:19-443`). User: Shadow Realm guard, broadcast, KD deaths, `PlayerDeath` event, permadeath reroll, penalties, corpse, revive at 1 HP, move to death room (`u/suicide.go:25-300`). | Spec: death branches on "does this actor return from a template". Recommend one engine death call (not a command) with the two bodies as its branches; `suicide` stays a player command until Phase 5 (`usercommands.go:108-109`). Producers: `NewRound_MobRoundTick.go:70`, `NewRound_DoCombat.go:927`, `Buff_ApplyBuffs.go:66-69` (raw `events.Input`). `vanish`: no producer. "The Guide" (MobId 38): no such mob in `empty` (only mobs 1 and 2 exist); recommend drop. Open question 7. |
| `wander` | none | false / - | See section 2. | Behaviour action (spec). |
| `unlock` (unregistered) | `unlock` (`usercommands/unlock.go`) | - / false | Mob version unlocks any container or exit with no key (`m/unlock.go:12-57`). Unreachable: not in the map (`mobcommands.go:24-46`), so `unlock`/`open` (`keywords.yaml:122`) from a mob falls to "looks a little confused". | Delete (spec). After the fold NPCs gain the user `unlock` (needs a key); no producer issues it. |

Plugin mob commands (not in the map; registered at load, `plugins.go:485-487`):

| Mob command | User command | Downed mob / user | Differences | Recommendation |
|---|---|---|---|---|
| `trash` (cleanup) | `trash` (`cleanup.go:79-125`) | false / false | Mob: silent removal plus `ItemOwnership` (`cleanup.go:127-144`). User: not-found text, messages unless sneaking, XP grant via `user.GrantXP` (account call) (`cleanup.go:84-120`). | Keep the user handler; XP grant needs `Account()` (nil for NPCs, so no XP). Drop `AddMobCommand` from the plugin API, or keep it as an alias of `AddUserCommand` for one release (open question 8). |
| `bury` (cleanup) | `bury` (`cleanup.go:146-176`) | false / false | Texts only (`cleanup.go:178-204`). | User handler. |
| `follow` (follow) | `follow` (`follow.go:377-510`) | true / true | User: refuses in a party, refuses self, NPC target that `HatesAlignment` refuses (`follow.go:384-387, 421-424, 496-498`). Mob: clears `Path` before following (`follow.go:596`); returns `false` when no target is found (`follow.go:616`), which makes the NPC emote "looks a little confused". Both `stop` branches print the follower's own name in the "stopped following you" text (`follow.go:457, 584`, bug). | One handler; keep `Path.Clear` for NPC followers (a behaviour concern, so via the actor's behaviour hook). Producers: `follow.go:297, 314, 326` (`follow lose`/`follow stop`). |

## 2. The four behaviour actions

### Who issues them

| Text | Producer | Line |
|---|---|---|
| `wander` | rat idle commands | `_datafiles/world/empty/mobs/startland/1-rat.yaml:10` |
| `wander` | guard spawn override idle commands (spawn slot replaces the template list, `rooms.go:600-602`) | `_datafiles/world/empty/rooms/startland/1.yaml:20` |
| (idle pick) | `HandleIdleMobs` runs `mob.GetIdleCommand()` at `ActivityLevel` odds | `hooks/MobIdle_HandleIdleMobs.go:45-49`; `mobs.go:506-522` (1 in 100 always empty) |
| `pathto home` | displaced non-wanderer | `hooks/MobIdle_HandleIdleMobs.go:27-31` |
| `pathto home` | wandered past `MaxWander` | `hooks/MobIdle_HandleIdleMobs.go:37-41` |
| `pathto <ids...>` | path step out of place, re-plan to waypoints | `hooks/NewRound_IdleMobs.go:86-96` |
| `pathto home` | re-plan failed | `hooks/NewRound_IdleMobs.go:99` |
| `pathto home` | `wander` past limit | `mobcommands/wander.go:19-28` |
| `pathto home` | `go home` | `mobcommands/go.go:68-71` |
| `<exitName>` (path step) | `IdleMobs` walks `mob.Path` | `hooks/NewRound_IdleMobs.go:105-116` |
| `noop` (delay N s) | `mob.Sleep`, which has no caller (grep `\.Sleep\(` over `*.go` finds only `time.Sleep`) | `mobs.go:275-277` |
| `noop` (1 s) | waypoint reached | `mobcommands/go.go:130-135` |
| `despawn` | none outside admin `command` | grep |

### What each does today

- **`wander`** (`mobcommands/wander.go:12-87`): charmed -> `(true, error "friendly mobs don't wander")` (logged by `world.go:641-643`). Away from home and `WanderCount > MaxWander` (with `MaxWander > -1`) -> `pathto home`. `MaxWander == 0` -> nothing. `wander loot` / `wander players` restrict to adjacent rooms with items/gold or players (no producer of either argument). Otherwise `room.GetRandomExit()`. Destination must be in the mob's zone (`restrictZone` is always true, `wander.go:36`). On a pick: `WanderCount++`, then `go <exit>`. Note: the zone check is on `r.Zone == mob.Character.Zone`; a failed check does nothing.
- **`pathto`** (`mobcommands/pathto.go:14-71`): `home` with `home-impossible` temp data set -> lose 10% max health per call (the stuck-NPC reaper), return. Parses `home` and integer room IDs; nothing, or first target is the current room -> nothing. `mapper.GetPath` error -> for `home`, set `home-impossible` and the `lost` adjective; returns `(false, err)`, so dispatch makes the NPC emote "looks a little confused (pathto home)" (`world.go:647-651`). Empty path for `home` -> same flags, `(true, nil)`. Otherwise `mob.Path.SetPath`. It does not move: `IdleMobs` issues one exit per round from the path (`NewRound_IdleMobs.go:80-124`) and clears the path, resetting `WanderCount` when home.
- **`despawn`**: section 1.
- **`noop`**: section 1. The delay lives in `mob.Command`'s scheduling and the world input tracker (`mobs.go:279-310`, `world.go:101-118`), not in `noop`.

Note for the spec line "wander and pathto pick an exit and issue `go <exit>`": today `pathto` only plans; the round tick issues the step as a bare exit name (`NewRound_IdleMobs.go:110`), which reaches `Go` through the fallthrough (`mobcommands.go:124-135`). Slice 4 can keep that shape (behaviour plans, tick issues `go <exit>`), with slice 6 moving the tick.

## 3. Every producer of mob command text (non-test)

`grep '\.Command\(|\.CommandFlagged\('` over `*.go`, mob receivers only, plus raw `events.Input` to mobs:

| File:line | Text | In mob registry? |
|---|---|---|
| `world.go:649` | `emote looks a little confused (%s %s).` | yes |
| `modules/follow/follow.go:307` | `<exitName>` (0.25 s) | `go` fallthrough |
| `modules/follow/follow.go:314` | `follow stop` | plugin |
| `internal/hooks/MobIdle_HandleIdleMobs.go:29, 40` | `pathto home` | yes |
| `internal/hooks/MobIdle_HandleIdleMobs.go:47` | idle command from data | see data rows |
| `internal/hooks/NewRound_IdleMobs.go:47` | `rest depression %d/%d` | yes |
| `internal/hooks/NewRound_IdleMobs.go:70` | `emote mumbles about losing their quarry.` | yes |
| `internal/hooks/NewRound_IdleMobs.go:94` | `pathto <id> <id>...` | yes |
| `internal/hooks/NewRound_IdleMobs.go:99` | `pathto home` | yes |
| `internal/hooks/NewRound_IdleMobs.go:110` | `<exitName>` | `go` fallthrough |
| `internal/hooks/NewRound_MobRoundTick.go:59` | charm `ExpiredCommand`, split on `;` (nothing sets a charm today) | depends |
| `internal/hooks/NewRound_MobRoundTick.go:70` | `suicide` | yes |
| `internal/hooks/NewRound_DoCombat.go:152` | `<exitName>` (charmed NPC follows a fleeing player) | `go` fallthrough |
| `internal/hooks/NewRound_DoCombat.go:477` | `go <exitName>` (attacked from another room) | yes |
| `internal/hooks/NewRound_DoCombat.go:479` | angry command (template, else race) | see data rows |
| `internal/hooks/NewRound_DoCombat.go:574` | `say I have a CombatAction that is too long. Please notify an admin.` | yes |
| `internal/hooks/NewRound_DoCombat.go:577` | each `;` part of a `CombatCommands` entry (none in `empty`) | depends |
| `internal/hooks/NewRound_DoCombat.go:638` | `equip <DisplayName>` | yes |
| `internal/hooks/NewRound_DoCombat.go:927` | `suicide` | yes |
| `internal/hooks/Buff_ApplyBuffs.go:66-69` | `suicide` (raw `events.Input`, ReadyTurn 0) | yes |
| `internal/mobs/mobs.go:276` | `noop` (delayed) | yes |
| `internal/mobcommands/wander.go:23` | `pathto home` | yes |
| `internal/mobcommands/wander.go:71, 79` | `go <exit>` | yes |
| `internal/mobcommands/go.go:69` | `pathto home` | yes |
| `internal/mobcommands/go.go:76` | `emote tries to go the <exit> exit, but it's locked.` | yes |
| `internal/mobcommands/go.go:134` | `noop` (1 s) | yes |
| `internal/mobcommands/equip.go:16`, `remove.go:15` | `emote struggles with their gear for a while, then gives up.` | yes |
| `internal/usercommands/give.go:201` | `emote counts his gold coins and chuckles a bit.` | yes |
| `internal/usercommands/give.go:203` | `emote considers the <item> for a moment.` | yes |
| `internal/usercommands/go.go:237` | `<rest>` (charmed NPC follows the player's exit) | `go` fallthrough |
| `internal/usercommands/go.go:272` | `<rest>` (pursuing NPC follows; has aggro) | `go` fallthrough |
| `internal/usercommands/admin.command.go:48-52` | anything an admin types (`command <mob> <text>`) | any |

Data strings (`_datafiles/world/empty`, the configured `DataFiles`, `config.yaml:191`):

| File:line | Field | Text |
|---|---|---|
| `mobs/startland/1-rat.yaml:9-11` | idlecommands | `emote wiggles its nose`, `wander`, `''` |
| `rooms/startland/1.yaml:16-20` | spawn idlecommands (guard) | `say did you know there's a sign in the Townsquare with a map of the area?`, `""` x3, `wander` |
| `races/0-ghostly_spirit.yaml:7` | angrycommands | `emote drifts aggressively.` |
| `races/1-human.yaml:11-14` | angrycommands | `emote is looking for trouble.`, `shout it's time to die!`, `shout oh man, it is on now!`, `shout i've been waiting for this moment.` |
| `races/10-rodent.yaml:8` | angrycommands | `emote squeeks angrily.` |
| `mobs/startland/2-guard.yaml` | none | no command fields |

No `combatcommands` anywhere in data (grep -l over `_datafiles`, `modules`, `internal`: only `1-rat.yaml`, three races, `rooms/startland/1.yaml` carry command fields).

Strings that must still parse after the fold: `emote <text>`, `say <text>`, `shout <text>`, `go <exit>`, bare `<exit>`, `equip <name>`, `suicide`, `rest <reason>`, `pathto home`, `pathto <ids>`, `wander`, `noop`, `follow stop`, and whatever an admin types.

Commands a mob is told to run that are not in the mob registry: **none found** in code or data. Reachable only through admin `command`: `unlock`/`open` (alias, `keywords.yaml:122`) and every user-only command (`attack`, `whisper`, `flee`, ...), which today fall to "looks a little confused". No producer issues `attack` to a mob (grep of `.Command(` shows `attack` only for party users, `attack.go:182, 230`).

## 4. Plugins and modules

- API: `Plugin.AddMobCommand(command, mobcommands.MobCommand, allowWhenDowned)` (`internal/plugins/plugins.go:311-322`), stored in `PluginCallbacks.mobCommands` (`plugincallbacks.go:12, 24`), registered at `Load` via `mobcommands.RegisterCommand` (`plugins.go:485-487`). `mobcommands.RegisterCommand`'s third parameter is named `isBlockable` but fills `AllowedWhenDowned` (`mobcommands.go:147-152`).
- Callers: `modules/cleanup/cleanup.go:58-59` (`trash`, `bury`), `modules/follow/follow.go:62` (`follow`). `gmcp` and `webhelp` register user commands only (`gmcp.Mudlet.go:73-75` by grep; webhelp none by grep).
- Docs to update: `_datafiles/html/admin/docs-modules.html:631-643` ("Mob Commands" API table), `:1554` (follow example), `:367`, `:1505` (prose).

## 5. Admin and help listings

- `internal/usercommands/admin.command.go:29-33`: with fewer than two args, lists `mobcommands.GetAllMobCommands()` through `templates/admincommands/help/command.command.template:13-16` ("Mob Commands:" block). Lines 6-8 of the template say mobs lack the full command range. Both go or change with the fold.
- `admin.command.go:41-66` dispatches to mob or player by queuing `events.Input{ActorId}`; it already routes by actor, so it survives the merge as is.
- No other help or keyword listing of mob commands found (grep `GetAllMobCommands`, `mobcommands` outside the package: `plugins`, `admin.command.go`, `world.go` only).

## 6. Other notes for the plan

- Mob dispatch calls `keywords.TryCommandAlias` twice (`mobcommands.go:63` without the multi-word split, then `:87`), so a multi-word alias (e.g. `about` -> `help about`) becomes the single "command" `help about` for mobs. Gone with the merge.
- Users check exits before commands (`usercommands.go:155`); mobs check commands before exits (`mobcommands.go:106-135`). An exit named like a command or emote resolves differently. Take the user order.
- Users cancel `cancel-on-action` buffs on every command (`usercommands.go:202`); mobs never. After the fold NPCs will; no NPC in `empty` has such a buff (unverified across buff data).
- Spec facts row 20 says 20 mob commands; the registry has 21 (`rest`). The plan's facts table should say 21 plus 3 plugin commands plus 1 unregistered file.

## Open questions for Calabe

1. **NPCs moving while in combat.** Today an NPC with an aggro target can walk (pursuing a fleeing player, `usercommands/go.go:245-273`; charmed followers, `NewRound_DoCombat.go:148-155`); a player cannot (`usercommands/go.go:19-22`). Options: (a) one rule, nobody in combat walks: pursuit and charmed follow-on-flee stop working; (b) keep the NPC exemption as an engine move that does not go through `go`; (c) `go` takes a flag (for example a pursuit flag on `events.EventFlag`) that skips the combat check, set by the pursuit and follow code for any actor. Recommend (c): one `go`, the exception named at the call sites, and players could get pursuit later by the same flag.
2. **Action points and encumbrance for NPCs.** User `go` charges AP; AP regenerates for players only (`NewTurn_ActionPoints.go:23-28`), so NPCs (YAML sets no `actionpoints`) would never move. Options: (a) extend AP regeneration to every actor and charge NPCs too: NPC speed becomes AP-limited, a gameplay change; (b) charge only actors with an account: a hidden actor-kind branch inside `go`; (c) regenerate for all actors but set NPC movement cost from the behaviour (cost 0 for now). Recommend (a) with regeneration for all actors, checked against wander rates in the slice 6 seeded test; if Calabe wants zero gameplay change in slice 4, (c).
3. **Locks, sneaking and room-ID moves for NPCs.** Mob `go` refuses locked exits with an emote and never unlocks the far side, ignores `hidden` for leave/enter text, and accepts `go <roomId>` teleports (`mobcommands/go.go:20-61, 74-79, 99-103`). Options: (a) take the user rules for everyone (NPCs use keys if they carry them, hidden NPCs move silently, no room-ID form); (b) keep each mob rule. Recommend (a); the room-ID form has no producer and keeps a teleport outside the movement path.
4. **`sayto`, `saytoonly`, `replyto`.** No producer in code or data; they existed for upstream's scripted conversations. Options: (a) drop all three; (b) promote `sayto` to a user command (directed speech, everyone hears it) and drop the other two; (c) keep all three as user commands. Recommend (a) now; directed speech belongs to the comms-channel work. If (b) or (c): `saytoonly` and sneaking `sayto` leak to the whole room over GMCP today (`gmcp.Comm.go:69-85`), so they need `Named` audience delivery.
5. **What an NPC's actor-only text does in slice 4.** The folded commands send actor text with `user.SendText`; slice 7 moves output to messaging. Options: (a) slice 4 gives the command an actor whose send is a no-op for NPCs (Account nil); (b) convert each folded command's sends to messaging inside slice 4. Recommend (a), keeping slice 4's scope; slice 7 converts.
6. **`rest` as a command.** It has no user counterpart; boredom issues it (`NewRound_IdleMobs.go:47`) and the charm-expiry constant chains it after an emote (`charminfo.go:7`). Options: (a) engine call like `despawn`, and charm expiry runs the emote then calls the engine; (b) keep `rest` as an NPC-only command. Recommend (a) (spec rule: no NPC-only commands).
7. **NPC death.** Options: (a) one engine death call branching on "returns from a template", with `suicide` left as the player command until Phase 5; (b) one `suicide` command that branches. Recommend (a). Also: drop the `vanish` form (no producer) and the "The Guide" MobId 38 case (no such mob in `empty`)?
8. **Plugin API `AddMobCommand`.** Options: (a) remove it and change cleanup and follow to one handler each; (b) keep it one slice as a deprecated wrapper. Recommend (a); modules are in-tree.
9. **`perma-gear`.** It blocks NPC equip/remove/drop today and NPC and player death loot. Options: (a) apply it to every actor in the shared `equip`, `remove`, `drop`; (b) drop it from those commands. Recommend (a).


# Inventory B: usercommands for Phase 4b slice 4 (master `48a7d92b`)

Read-only. No repo file edited, no git state changed.

## Read 96 of 96 files

Read with the Read tool, in full (chunks of at most 400 lines):

- `internal/usercommands/` all 68 `.go` files: admin.badcommands.go, admin.build.go, admin.command.go, admin.copyover.go, admin.grant.go, admin.item.go, admin.locate.go, admin.mob.go, admin.modify.go, admin.mute.go, admin.prepare.go, admin.redescribe.go, admin.reload.go, admin.rename.go, admin.room.go, admin.server.go, admin.spawn.go, admin.syslogs.go, admin.teleport.go, admin.visit.go, admin.zone.go, alias.go, attack.go, biome.go, broadcast.go, consider.go, default.go, descriptions.panels.go, drop.go, emote.go, equip.go, exits.go, flee.go, get.go, give.go, go.go, help.go, history.go, inventory.go, inventory.panels.go, keyring.go, lock.go, look.go, look.panels.go, macros.go, motd.go, noop.go, online.go, party.go, password.go, print.go, put.go, quit.go, read.go, remove.go, save.go, say.go, set.go, shoot.go, shout.go, show.go, start.go, suicide.go, unlock.go, usercommands.go, visited.go, whisper.go, who.go
- `internal/actors/actor.go`, `controller.go`, `registry.go`, `wire.go`
- `internal/messaging/messaging.go`, `render.go`; `internal/perception/perception.go`
- `internal/plugins/plugincallbacks.go`; `internal/hooks/Message_SendMessages.go`; `internal/mobcommands/mobcommands.go`; `internal/guards/messaging_test.go`

Read with the Read tool, in part:

- `2026-10-27-phase-4b-npcs-are-players-design.md` lines 40-315 (lines 9-39 printed with `sed`)
- `internal/plugins/plugins.go` 285-324, 465-494
- `modules/cleanup/cleanup.go` 50-204; `modules/follow/follow.go` 370-519; `modules/gmcp/gmcp.Mudlet.go` 60-84, 215-347
- `world.go` 85-189, 420-500, 490-688
- `internal/hooks/NewRound_DoCombat.go` 140-169
- `internal/accounts/userrecord.go` 75-184, 270-369
- `internal/rooms/rooms.go` 2371-2410; `internal/rooms/roomdetails.go` 80-139

Printed in part with `sed`/`grep -A` (Bash), not the Read tool:

- `main.go` 175-186, 355-362; `modules/gmcp/gmcp.Help.go` 60-70; `modules/webhelp/webhelp.go` 115-125; `internal/templates/templates.go` 125-175; `internal/rooms/rooms.go` 255-300; `internal/rooms/roomdetails.go` 60-80; `internal/rooms/roommanager.go` 265-275; `internal/accounts/userrecord.go` 605-630; `internal/accounts/users.go` 184-196; `internal/accounts/userrecord.prompt.go` 116-140; `modules/gmcp/gmcp.Mudlet.go` 103-111

## Facts verified against source

| Fact | Value | Where |
|---|---|---|
| Files in `internal/usercommands` | 68 `.go`, 0 `_test.go` (plus AGENTS.md, CLAUDE.md, README.md). Not 71: 71 is the command count with plugins (65 + 6) | `ls` |
| Lines | 11,655 | `wc -l *.go` |
| Built-in registry | 65 entries | `usercommands.go:52-121` |
| Plugin user commands | 6 (`trash`, `bury`, `follow`, `mudletmap`, `mudletui`, `checkclient`) | section 4 |
| Mob registry | 21 entries (spec facts table says 20; `rest` is the extra one) | `mobcommands.go:25-47` |
| Signature today | `func(rest string, user *accounts.Account, room *rooms.Room, flags events.EventFlag) (bool, error)` | `usercommands.go:30` |
| `actors.Actor` | `Id ActorId`, `Character *characters.Character`, unexported `controller`; `Controller()` under `mu.RLock` | `actors/actor.go:16-28` |
| `Controller` | `Account() *accounts.Account`, `Perceive(perception.Perception)`; `BehaviourController.Account()` returns nil | `actors/controller.go:14-20, 55` |
| Lookups | `actors.Get(id)`, `actors.AccountOf(id)`, `actors.ByUserId(userId)`, `actors.All()` | `actors/registry.go:62-103` |
| Interface delivery | `Kind == Interface` or `Audience == ActorOnly` goes to the actor only (`messaging.go:38-43`); `ConnectionController.Perceive` calls `account.SendText(p.Text)` (`controller.go:38-43`); `BehaviourController.Perceive` drops Interface (`controller.go:69-72`) | |
| Token resolution | Every delivery, Interface included, runs `resolveTokens` on `{actor}`, `{target}`, `{direction}` | `messaging.go:75-76`, `render.go:28-49` |
| `Account.SendText` | queues `events.Message{UserId, Text: txt+"\n"}` | `accounts/userrecord.go:334-341` |
| `room.SendText` | queues `events.Message{RoomId, ExcludeUserIds}`; the hook sends only to `room.GetPlayers()`, so NPCs never get room text today | `rooms/rooms.go:273-282`; `hooks/Message_SendMessages.go:38-85` |
| `Account.Command` / `CommandFlagged` | queue `events.Input{ActorId: u.Character.Id}`; `CommandFlagged` also calls `BlockInput()` (session) when `CmdBlockInput` is set | `userrecord.go:277-322` |
| Connection writes from usercommands | none; guard `TestOnlyMessagingWritesToConnections` does not list any usercommands file | `guards/messaging_test.go:21-37` |

## 1. Per-file inventory

Column key: **Char** = count of `user.Character` occurrences; **ST** = count of `user.SendText(` calls; neither package uses `SendWebClientCommand` (0 hits). **Account items** lists every other `user.` use with line numbers. For `user.UserId` I tag the purpose: `tpl` = `templates.Process(..., user.UserId)` (picks the screen-reader template per account, `templates.go:142-158`); `excl` = exclusion argument to `room.SendText`/`SendTextCommunication`; `move` = `rooms.MoveToRoom(user.UserId, ...)` (looks the user up by account, `roommanager.go:269-273`); `party` = `parties.*` keyed by UserId; `self` = compares a `FindByName` player id with the caller's UserId; `evt` = an event field.

**Class** values: **any** = works for any actor (account uses have an obvious nil answer such as Muted=false, Role=not mod, no permission); **any (output only)** = has no effect except text to the caller, so an NPC runs it and the text is dropped; **account** = needs an account, refuse for NPCs; **mixed** = core works for any actor, named parts need the account.

### Admin files (all `AdminOnly: true`, so `TryCommand` gates them on `HasRolePermission`, `usercommands.go:241`)

| File | Commands | Char | ST | Account items | Class |
|---|---|---|---|---|---|
| admin.badcommands.go | badcommands | 0 | 1 | UserId tpl 35 | account |
| admin.build.go | build | 6 | 16 | UserId tpl 29; move 43, 146 | account |
| admin.command.go | command | 0 | 1 | UserId tpl 31. Queues `events.Input` for a target mob or player body (48-61) | account |
| admin.copyover.go | copyover | 0 | 2 | none | account |
| admin.grant.go | grant | 0 | 5 | UserId tpl 24; self target 48; `u.GrantXP` on the target account 55 | account |
| admin.item.go | item (item_List, item_Spawn, item_Create) | 1 | 35 | HasRolePermission 38, 49, 60; ClientSettings 108 (screen width, session); StartPrompt 158; ClearPrompt 261, 278, 308, 413; UserId tpl 30, 195, 216, 301, 328; excl 130 | account |
| admin.locate.go | locate | 0 | 5 | UserId tpl 23, 153, 166; sends to target account `locateUser.SendText` 40 | account |
| admin.mob.go | mob (mob_List, mob_Spawn, mob_Create) | 1 | 38 | HasRolePermission 42, 53, 64; ClientSettings 112; StartPrompt 187; ClearPrompt 204, 245, 318, 397; UserId tpl 34, 236, 284, 311, 339; excl 150; `Help(..., user, ...)` 264 | account |
| admin.modify.go | modify | 0 | 8 | HasRolePermission 33; UserId tpl 25; writes other accounts' `Role` 65, 92 and `SaveUser` 67, 94 | account |
| admin.mute.go | mute, unmute | 0 | 6 | UserId tpl 20, 47; writes target `u.Muted` 31, 58; reads `u.Username` 33, 60 | account |
| admin.prepare.go | prepare | 0 | 2 | UserId tpl 17 | account |
| admin.redescribe.go | redescribe | 4 | 3 | UserId tpl 24; excl 52 | account |
| admin.reload.go | reload | 0 | 8 | UserId tpl 23 | account |
| admin.rename.go | rename | 4 | 3 | UserId tpl 24; excl 56 | account |
| admin.room.go | room (room_Edit_Containers, room_Edit_Exits, room_Edit_Mutators) | 0 | 137 | HasRolePermission 73, 80, 90, 100, 118, 188, 235, 280, 307, 364, 388, 531; StartPrompt 604, 1201, 1584; ClearPrompt 638, 658, 684, 1136, 1235, 1255, 1281, 1572, 1704; UserId tpl 47, 302, 608, 790, 825, 858, 966, 1205, 1446, 1481, 1514, 1621, 1654, 1685; helper `room_Edit_Containers_SendRecipes(user *accounts.Account, ...)` 1141, called 1029, 1047, 1080 | account |
| admin.server.go | server (server_Config) | 0 | 27 | StartPrompt 317; ClearPrompt 337, 346, 368, 398; UserId tpl 36, 180, 307, 322, 418 | account |
| admin.spawn.go | spawn | 3 | 4 | UserId tpl 25; excl 54, 80, 93 | account |
| admin.syslogs.go | syslogs | 0 | 3 | UserId tpl 20; ConnectionId 29, 35 (`events.Log{FollowAdd/FollowRemove}`) | account |
| admin.teleport.go | teleport | 1 | 8 | UserId tpl 30; HasRolePermission 53, 73, 84; `targetUser := user` 36, then `targetUser.UserId` move 92, excl 105, party 108, 111; `Look(``, partyUser, ...)` 130 and `Look(``, targetUser, ...)` 147 | account |
| admin.visit.go | visit (visit_List, visit_SetUnset, visit_resolveTarget, visit_applyZone) | 0 | 9 | UserId tpl 24, 40, 103, 112; `visit_resolveTarget(args, invoker *accounts.Account)` returns the invoker account as target 47-49; `SaveUser(*targetUser)` 138, 159 behind `isOffline`, which `visit_resolveTarget` never sets true (49, 55, 58) | account |
| admin.zone.go | zone (zone_Edit) | 0 | 23 | UserId tpl 32, 179, 208, 239; StartPrompt 142; ClearPrompt 278 | account |

### Player files

| File | Commands | Char | ST | Account items | Class |
|---|---|---|---|---|---|
| alias.go | alias | 0 | 6 | AddCommandAlias 26; Aliases 69, 74; UserId tpl 66, 79 | account |
| attack.go | attack | 18 | 6 | UserId party 22, 173, 174, 215, 222, 223; self 98, 128, 144, 176, 225; excl 200, 251; `room.CanPvp(user, p)` 210 (takes two accounts but reads only `.Character`, `rooms.go:2371-2398`); `partyUser.Command(...)` 182, 230; `p.SendText` 245 | mixed: aggro setting works for any actor; party target pick (22-89) and party auto-attack (173-189, 222-235) need the account; self-exclusion and PvP need an actor-keyed form |
| biome.go | biome | 0 | 2 | none | any (output only) |
| broadcast.go | broadcast | 3 | 1 | Muted 15; Role 25 (`SourceIsMod`) | any (NPC: not muted, not mod). A mob `broadcast` also exists |
| consider.go | consider | 2 | 4 | UserId self 25 | any (output only) |
| default.go | default | 0 | 0 | `user.Command(`storage`)` 17; `Look(``, user, ...)` 22. No `storage` command is registered anywhere (grep finds `storage` only in default.go) | any |
| descriptions.panels.go | none (panel helpers) | 0 | 0 | none | n/a |
| drop.go | drop | 14 | 8 | UserId excl 72, 97, 116; recursive `Drop(..., user, ...)` 31, 37 | any |
| emote.go | emote; also every `emoteAliases` key via `usercommands.go:262-265` | 5 | 4 | UserId excl 102, 113, 131 (131 is `SendTextCommunication`); Muted 118 (free-text emotes only) | any |
| equip.go | equip | 12 | 7 | UserId excl 55, 72, 80 | any |
| exits.go | none: `Exits` is not registered and has no caller (grep `Exits(` finds only its definition) | 0 | 1 | `rooms.GetDetails(room, user)` 12; UserId tpl 14 | dead code |
| flee.go | flee | 6 | 1 | none | any |
| get.go | get | 35 | 27 | UserId excl 111, 126, 144, 176, 219, 251, 284, 322, 373, 381, 401; stash owner `stashItemMatch.StashedBy == user.UserId` 335; recursive `Get(..., user, ...)` 26, 33 | any (stash recovery is UserId-keyed, so an NPC cannot reclaim a stash) |
| give.go | give | 19 | 12 | UserId excl 84, 108, 132, 171, 183; self 101; target sends `targetUser.SendText` 79, 127 | any |
| go.go | go; also the exit-name rewrite (`usercommands.go:155-158`) and the fallback `Go(cmd, ...)` (`usercommands.go:273`) | 25 | 15 | UserId excl 90, 99, 114, 196, 203, 337; GMCP `gmcpSendFunc(user.UserId, ...)` 136, 327; `user.CommandFlagged(rest, flags\|CmdIsRequeue\|CmdBlockInputUntilComplete, 1)` 148; move 175; party 208, 210, 213; `partyUser.Command(rest)` 219; `mobs.IsHostile(groupName, user.UserId)` 295; `Look(``, user, destRoom, CmdSecretly)` 313 | mixed: the move itself needs a non-account path for NPCs (`MoveToRoom` is account-only); party follow, GMCP and the exit-message requeue (`CommandFlagged` blocks input on the session) need the account |
| help.go | help (and exported `GetHelpContents` 135) | 0 | 2 | HasRolePermission 56 (admin topics listed only if permitted); UserId tpl 76 | any (output only) |
| history.go | history | 0 | 1 | EventLog 26; UserId tpl 42 | account |
| inventory.go | inventory | 1 | 1 | `buildInventoryPanel(user, ...)` 105 | any (output only) |
| inventory.panels.go | none (`buildInventoryPanel(user *accounts.Account, ...)` 18) | 1 | 0 | uses only `user.Character` (19) | n/a (helper can take `*characters.Character`) |
| keyring.go | keyring | 2 | 1 | UserId tpl 164 | any (output only) |
| lock.go | lock | 14 | 10 | UserId excl 49, 71, 101, 123 | any |
| look.go | look; also called by default, go, start, teleport and `hooks/NewRound_DoCombat.go:159` | 17 | 43 | UserId excl 56, 115, 242, 275, 325, 375, 426, 431; `u.SendText` 86; HasRolePermission 310; in `lookRoom`: GetConfigOption 438 (`tinymap`), ScreenReader 443, `mapper.Config{UserId}` 462, `sign.VisibleUserId == user.UserId` 524, `item.StashedBy != user.UserId` 581, UserId tpl 516, 526, 534, 542, 595, 600; `rooms.GetDetails(room, user)` 454, 508, 513 (`GetDetails` takes `*accounts.Account` and reads `user.UserId` 96, `HasRolePermission` 126 and `GetConfigOption` 245, `roomdetails.go`) | mixed: the observer lines ("X is looking at...") work for any actor; the rendering of the room for the caller reads account settings and is dropped for an NPC anyway |
| look.panels.go | none (panel helpers on `*characters.Character`) | 0 | 0 | none | n/a |
| macros.go | macros | 0 | 8 | Macros 15, 20, 22, 29 | account |
| motd.go | motd | 0 | 1 | none | any (output only) |
| noop.go | noop | 0 | 0 | none | any (spec turns mob `noop` into a behaviour action) |
| online.go | online | 0 | 1 | Role 25, 77; UserId tpl 97 | any (output only) |
| party.go | party | 10 | 38 | UserId party 30, 36, 38, 46, 79, 82, 146, 159, 196, 218, 345, 364, 391, 451, 466, 483, 519, 570 and skip-self 164, 414, 471, 489, 631; EventLog 47, 161, 403, 452, 467, 511; UserId tpl 336; sends to other accounts `u.SendText`/`invitedUser.SendText` (14 + 1 in file) | account |
| password.go | password | 0 | 4 | StartPrompt 12; HasPlaintextPassword 14; PasswordMatches 20; ClearPrompt 22, 43, 49; SetPassword 47; `accounts.SaveUser(*user)` 53 | account |
| print.go | print, printline | 0 | 1 | `events.Message{UserId: user.UserId, Text: rest}` 16-19 (no trailing newline, unlike `SendText`) | any (output only) |
| put.go | put | 9 | 10 | UserId excl 100, 115 | any |
| quit.go | quit | 1 | 1 | `user.AddBuff(0, `quitting`)` 15 (queues `events.Buff{ActorId: u.Character.Id}`, `userrecord.go:324-331`) | account (leaving the game has no NPC meaning) |
| read.go | read | 3 | 3 | UserId excl 44 | any |
| remove.go | remove | 13 | 6 | UserId excl 76; recursive `Remove(..., user, ...)` 21 | any |
| save.go | save | 0 | 2 | `accounts.SaveUser(*user)` 12 | account |
| say.go | say | 5 | 2 | Muted 15; UserId excl 29, 31 (`SendTextCommunication`); `room.SendTextToExits` 36 | any. A mob `say` also exists |
| set.go | set | 2 | 51 | ScreenReader 27, 330, 335; GetConfigOption 33, 41, 49, 57, 64, 71, 78, 112, 136, 160, 186, 223, 260, 298; SetConfigOption 124, 148, 172, 199, 200, 202, 203, 205, 206, 236, 237, 239, 240, 242, 243, 274, 279, 314, 316; Macros 356, 357, 366, 396; `events.UserSettingChanged{UserId}` 104, 127, 151, 175, 212, 249, 285, 322, 338, 402 | account (only `set description`, 93-109, is character data; it still emits a UserId event) |
| shoot.go | shoot | 10 | 9 | UserId excl 94, 123; party 106 | any (the party check reads nil for an NPC) |
| shout.go | shout | 3 | 2 | Muted 14; UserId excl 30, 32, 38, 46, 59 (all `SendTextCommunication`) | any. A mob `shout` also exists |
| show.go | show | 6 | 7 | UserId excl 70, 99; `targetUser.SendText` 58, 62 | any |
| start.go | start | 14 | 21 | StartPrompt 26; UserId tpl 53, 106; TempName 115; Username 115, 122; ClearPrompt 162, 195, 219; SetCharacterName 166; EventLog 177; `events.PlayerChanged{UserId}` 180; move 201, 251; excl 209; `Help(...)` 75; `Start(...)` 163; `Look(...)` 212, 254 | account |
| suicide.go | suicide | 62 | 4 | UserId excl 38; `KD.AddPlayerKill(user.UserId, ...)` 58; `events.PlayerDeath{UserId, Username}` 99, 101; EventLog 117, 161, 198, 211, 222, 240, 256; UserId tpl 120; `oldRoom.RemovePlayer(user.UserId)` 138; `actors.BodyReplaced(user, oldBodyId)` 145; `oldRoom.AddPlayer(user.UserId)` 151; move 154, 297; `rooms.Corpse{UserId}` 264, 279; `events.CharacterVitalsChanged{UserId}` 293 | account (player death stand-in; spec section 3 keeps NPC death separate. A mob `suicide` also exists) |
| unlock.go | unlock | 14 | 10 | UserId excl 49, 71, 101, 123 | any |
| usercommands.go | registry, `TryCommand`, `RegisterCommand` | 6 | 4 | TryCommandAlias 161; HasPlaintextPassword 219; InputBlocked 236; HasRolePermission 241; UserId 249 (log) | see section 2 |
| visited.go | visited | 1 | 1 | UserId tpl 63 | any (output only) |
| whisper.go | whisper | 3 | 5 | Muted 38; `toUser.SendText` 43; `events.Communication{TargetUserId: toUser.UserId}` 49 | any (target resolved by character name to an account, 30) |
| who.go | who | 0 | 1 | `rooms.GetDetails(room, user)` 12; UserId tpl 14 | any (output only) |

Counts: 31 files need an account (21 admin files gated by role, plus alias, history, macros, party, password, quit, save, set, start, suicide). The spec's estimate was "about 18 of the 68" (spec line 143); the 10 non-admin files are the recount the spec asked for. Mixed: attack, go, look. Every other command file is "any".

The `user` value is also passed whole to helpers that take `*accounts.Account`: `rooms.GetDetails` (exits.go:12, look.go:454/508/513, who.go:12), `room.CanPvp` (attack.go:210), `actors.BodyReplaced` (suicide.go:145), `buildInventoryPanel` (inventory.go:105), `room_Edit_Containers_SendRecipes`, `visit_*`, and the recursive or cross calls `Look`, `Help`, `Drop`, `Get`, `Remove`, `Start`, `Emote`, `Go` (full list from grep `[(, ]user[,)]`).

## 2. `TryCommand` and `CommandAccess` (`usercommands.go`, read in full)

`CommandAccess{Func UserCommand; AllowedWhenDowned bool; AdminOnly bool}` (23-27). The positional registry literals are `{Func, AllowedWhenDowned, AdminOnly}`. `RegisterCommand(command, handlerFunc, disabledWhenDowned, isAdminOnly)` (282-288) stores its third argument into `AllowedWhenDowned`; the parameter name says the opposite, but the only caller passes `info.AllowedWhenDowned` (`plugins.go:482`), so behaviour is consistent.

`TryCommand(cmd, rest string, userId int, flags)` in order:

| Lines | Step | Needs from the account |
|---|---|---|
| 142-145 | `accounts.GetByUserId(userId)`; error if nil | the account is the input key |
| 147-150 | `rooms.LoadRoom(user.Character.RoomId)` | character only |
| 152 | trim `rest` | none |
| 155-158 | exit name typed as a command: `rest = cmd; cmd = "go"`. `userDisabled` stays false on this branch | none |
| 161-177 | personal aliases: `user.TryCommandAlias(cmd)`, multi-word split | `Aliases` (account) |
| 179-195 | global aliases: `keywords.TryCommandAlias(cmd)` | none |
| 197 | `userDisabled = user.Character.IsDisabled()` (only on the non-exit branch) | character |
| 202 | `user.Character.CancelBuffsWithFlag("cancel-on-action")` | character |
| 206-217 | `look me/self/myself` rewritten to the character's name | character |
| 219-223 | plaintext password gate: anything but `password` is refused | `Password` (account) |
| 225-260 | registry hit: | |
| 227-239 | if not `AllowedWhenDowned`: refuse "unable to do that while downed" when `userDisabled && !AdminOnly` (so admin-only commands skip the downed check for everyone, before the permission check); then silently refuse when `user.InputBlocked()` | `InputBlocked` is session state on the account (`userrecord.go:300-302`) |
| 241 | run if `!AdminOnly \|\| user.HasRolePermission(cmd, true)` | `Role`, `Permissions` |
| 243-246 | `util.TrackTime("usr-cmd[cmd]")` | none |
| 248-250 | `mudlog.Info("Admin Command", ..., "userId", user.UserId)` | `UserId` |
| 253 | `cmdInfo.Func(rest, user, room, flags)` | |
| 256-259 | admin-only without permission: "You don't have permission to use ..." | |
| 262-265 | `emoteAliases[cmd]` hit: `Emote(cmd, ...)` with no downed or input-blocked check | none |
| 268-275 | fallback `Go(cmd, ...)` with timing; no downed check here (`go.go:40` checks `IsDisabled` itself) | none |
| 278 | `false, nil` (world then prints "not recognized" and queues `emote @looks a little confused`, `world.go:591-598`) | |

Checks the brief asked about that `TryCommand` does **not** do: no plugin lookup (plugins only add entries to the same map via `RegisterCommand`, `plugins.go:481-483`); no function-exporter use (`GetExportedFunction` is used only in `go.go:134, 325` for GMCP); no `CmdSecretly` handling (only `look.go:23` reads it); no cooldowns (grep `Cooldown|cooldown` in the package: 0 hits; the same grep found the `CmdSecretly` lines, so it could match).

The input path around it is also account-keyed and is what slice 4 merges: `world.go:93` branches on `actors.AccountOf(input.ActorId)`; nil goes to `processMobInput` (96-123), otherwise `userId := inputUser.UserId` (125) with `ignoreInput`/`UnblockInput` per UserId (131-139, 159-161, 174-181). `processInput` (502-611) uses `GetPrompt`/`ClearPrompt` (512-526), `SetLastInputRound` (542), `Macros` (545-563), `ConnectionId` + `GetCommandPrompt` for prompt redraws (587-588, 602-605). `processMobInput` (613-653) calls `mobcommands.TryCommand`, whose order differs: global alias, downed check, registry, then `Go`, then emote aliases (`mobcommands.go:92-142`); no personal alias, input-block, password or permission checks.

## 3. Output and callers

Totals in `internal/usercommands` (grep `-o`, all files):

- `user.SendText(`: **688**
- `room.SendText(`: **66**; plus `destRoom.SendText(` 2 (go.go:199, start.go:205) and `gotoRoom.SendText(` 1 (admin.teleport.go:103): 69 room-text sends
- `SendTextToExits(`: **2** (go.go:205, say.go:36)
- `SendTextCommunication(`: **8** (emote.go:129; say.go:29, 31; shout.go:30, 32 on `room`; shout.go:38, 46, 59 on `otherRoom`)
- sends to other accounts: `u.SendText(` 14, `targetUser.SendText(` 4, `partyUser.SendText(` 2, `toUser`, `p`, `locateUser`, `invitedUser` 1 each (24)
- `events.Message{` 1 (print.go:16); `SendWebClientCommand` 0

Callers of `usercommands` outside the package (grep `usercommands\.` over every `.go`, tests included; no test file matched):

| File:line | Use |
|---|---|
| `world.go:580` | `usercommands.TryCommand(command, remains, userId, flags)` |
| `internal/hooks/NewRound_DoCombat.go:159` | `usercommands.Look(``, user, newRoom, events.CmdSecretly)` after a flee |
| `modules/gmcp/gmcp.Mudlet.go:344` | `usercommands.Help("mudletui", user, room, flags)` |
| `modules/gmcp/gmcp.Help.go:65` | `usercommands.GetHelpContents(evt.Topic)` (no account) |
| `modules/webhelp/webhelp.go:121` | `usercommands.GetHelpContents(searchTerm)` (no account) |
| `internal/plugins/plugins.go:297, 300, 303` | `AddUserCommand(..., handlerFunc usercommands.UserCommand, ...)`, builds `usercommands.CommandAccess` |
| `internal/plugins/plugins.go:482` | `usercommands.RegisterCommand(cmd, info.Func, info.AllowedWhenDowned, info.AdminOnly)` |
| `internal/plugins/plugincallbacks.go:11, 23` | `map[string]usercommands.CommandAccess` |
| `main.go:181` | `usercommands.AddFunctionExporter(plugins.GetPluginRegistry())` |
| `main.go:183` | `usercommands.SetRoomTagProvider(plugins.GetRegisteredRoomTags)` |
| `main.go:359` | `usercommands.SetCopyoverFunc(triggerCopyover)` |

## 4. Commands registered from outside the package

All via `Plugin.AddUserCommand(command string, handlerFunc usercommands.UserCommand, allowWhenDowned bool, isAdminOnly bool)` (`plugins.go:297-308`), so every handler has today's `UserCommand` signature.

| Command | Handler | Registered | Downed / admin | Account use in the handler |
|---|---|---|---|---|
| `trash` | `(*CleanupModule).userTrashCommand` | `modules/cleanup/cleanup.go:55` (def 79) | false / false | `user.GrantXP` 118 (Account method: sends text, writes `EventLog`, `userrecord.go:176-184`); UserId excl 106 |
| `bury` | `(*CleanupModule).userBuryCommand` | `cleanup.go:56` (def 146) | false / false | UserId excl 165 |
| `follow` | `(*FollowModule).followUserCommand` | `modules/follow/follow.go:61` (def 377) | true / false | `parties.Get(user.UserId)` 384; follow source keyed `followId{userId: user.UserId}` 419; target sends via `accounts.GetByUserId` 437-487 |
| `mudletmap` | `(*GMCPMudletModule).sendMapCommand` | `modules/gmcp/gmcp.Mudlet.go:73` (def 322) | true / false | ConnectionId 324; UserId 326 |
| `mudletui` | `(*GMCPMudletModule).sendUICommand` | `gmcp.Mudlet.go:74` (def 252) | false / false | ConnectionId 254; `getUserBoolOption` (GetConfigOption, 103-111) 268; SetConfigOption 291 and via `handleToggleCommand` 243; UserId 290, 295, 299 |
| `checkclient` | `(*GMCPMudletModule).checkClientCommand` | `gmcp.Mudlet.go:75` (def 333) | true / false | ConnectionId 335; `getUserBoolOption` 338; `usercommands.Help` 344 |

The same modules register mob commands through `AddMobCommand` with `mobcommands.MobCommand func(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error)`: `trash` (`cleanup.go:58`, def 127), `bury` (`cleanup.go:59`, def 178), `follow` (`follow.go:62`, def 512). `modules/webhelp` registers no command. `modules/all-modules.go` and `modules.lock.yaml` were not read (unverified whether they register anything; the grep for `AddUserCommand`/`RegisterCommand(` over all `.go` found only the sites above).

## 5. `Command(` call sites queuing input for a player's body (non-test)

Grep `\.(Command|CommandFlagged)\(` over every non-test `.go`, then dropped 33 mob receivers (`mob.`, `m.`, `defMob.`, ...; each checked) and `exec.Command`.

| File:line | Receiver | Text |
|---|---|---|
| `internal/accounts/userrecord.go:621` | `u` (WimpyCheck) | `flee`, wait -1 |
| `internal/hooks/NewRound_AutoHeal.go:48` | `user` | `suicide` |
| `internal/hooks/NewRound_DoCombat.go:906` | `user` | `suicide` |
| `internal/hooks/PlayerSpawn_HandleJoin.go:93` | `user.CommandFlagged` | `look`, `CmdSecretly` |
| `internal/usercommands/attack.go:182` | `partyUser` | `attack <mob body ID>` |
| `internal/usercommands/attack.go:230` | `partyUser` | `attack <player body ID>` |
| `internal/usercommands/default.go:17` | `user` | `storage` (not registered anywhere) |
| `internal/usercommands/go.go:148` | `user.CommandFlagged` | `rest` (the exit), `flags\|CmdIsRequeue\|CmdBlockInputUntilComplete`, wait 1 |
| `internal/usercommands/go.go:219` | `partyUser` | `rest` (follow the leader) |
| `modules/follow/follow.go:297` | `moverUser` | `follow lose` |
| `modules/follow/follow.go:322` | `user` | the exit name, wait .25 |
| `modules/follow/follow.go:326` | `user` | `follow stop` |
| `world.go:597` | `user` | `emote @looks a little confused` |

Both methods only queue `events.Input{ActorId: u.Character.Id}` (`userrecord.go:284-288, 315-320`), so an actor-level equivalent needs no account except for the `BlockInput` in `CommandFlagged`.

## 6. Tests

None. `internal/usercommands` has no `_test.go` (`ls *_test.go`: no such file), and no test anywhere references `usercommands.` (the grep in section 3 included test files). Related note from the spec (line 276): `accounts.SetTestUser` skips `OnActivate`, so a test that uses it gets no actor.

## Other findings

- `exits.go` `Exits` is dead: not registered, no caller.
- `default.go:17` queues `storage`, which nothing registers; on a `storage`-tagged room the player gets "storage not recognized" plus the confused emote (`world.go:591-598`). Inferred from code, not run.
- `admin.visit.go`: the offline branch (`SaveUser`, 138, 159) is unreachable because `visit_resolveTarget` never returns `isOffline = true`.
- `admin.room.go` checks permission keys `room.tag` (118) and `room.exits` (307, 364) that its header comment (23-34) does not list.
- Interface text through `messaging.Send` would resolve `{actor}`, `{target}`, `{direction}` (`messaging.go:75-76`). No template in `_datafiles` contains those tokens (grep: only `combat-messages/*.yaml`, which slice 5 owns, and admin HTML), but text the player typed and gets echoed (say, emote, alias and macro listings, item names and descriptions) could.

## Open questions for Calabe

1. **How command output reaches the actor in this slice.** Today every command calls `user.SendText` 688 times. Player-facing effect: none if done right; an NPC running `look` or `inventory` produces text that must go nowhere.
   - A: a helper in usercommands, `tell(actor, text)`, that calls `Account().SendText` when the account is non-nil and does nothing otherwise. Pros: byte-identical output, small diff, slice 7 later swaps the helper body to `messaging`. Cons: one more indirection that slice 7 must remove.
   - B: convert all 688 sites to `messaging.Send(Interface, ActorOnly)` now. Pros: the end state. Cons: this is slice 7's work, and Interface text gets token resolution, so echoed player text containing `{actor}` would change.
   - C: write `actor.Controller().Account().SendText(...)` inline. Cons: panics for NPCs.
   - Locks in: A keeps slice 4 to signatures and dispatch; B pulls output conversion forward and needs a ruling on token resolution for Interface text.
   - Recommendation: A.

2. **Room text exclusion lists.** 69 room sends and 8 communication sends pass `user.UserId` to exclude the speaker; the hook delivers room text to players only (`Message_SendMessages.go:45`). Options: A, pass the account's UserId when there is one and 0 for an NPC (harmless, NPCs never receive room text today); B, convert these to `messaging` now. Recommendation: A, convert in slice 7 with the text-parity check.

3. **Movement for NPCs in the shared `go`.** `rooms.MoveToRoom` takes a UserId and looks up an account (`roommanager.go:269-273`); mob `go` moves with `RemoveMob`/`AddMob` (`mobcommands/go.go:42-43, 110-111`). Options: A, `go` branches on `Account()` and keeps each path's move; B, add one actor-level move in `rooms` now; C, keep mob `go` separate until later. A keeps outcomes identical with the least new mechanism; B is cleaner but touches room bookkeeping that slice 2 just settled. Recommendation: A, with the mob-variant diff the spec already requires (section 3) deciding the text.

4. **What `TryCommand` does for an NPC actor.** Account-only steps: personal aliases (161), plaintext-password gate (219), `InputBlocked` (236), role permission (241). Options: A, skip the first three when `Account()` is nil and treat the permission check as "no permission" (NPCs cannot run admin commands); B, also let NPCs run admin commands (some upstream content had mobs run them; nothing in `empty` does, unverified for other worlds). The two dispatch orders also differ (user: exit, alias, registry, emote alias, go; mob: alias, registry, go, emote alias). Recommendation: A, and adopt the user order for both, listing in the plan any command name that collides with an exit name.

5. **Party, GMCP, stash, signs and screen-reader templates are keyed by UserId** (`parties.Get`, `gmcpSendFunc`, `StashedBy`, `VisibleUserId`, `templates.Process(..., userId)`, `mobs.IsHostile(group, userId)`, `rooms.GetDetails(room, *Account)`). Options: A, keep them account-keyed and skip them when `Account()` is nil; B, re-key them by ActorId in this slice. Recommendation: A; re-keying is not needed for "one command set" and belongs to the slices that own those systems.

6. **Plugin API.** `AddUserCommand` and `AddMobCommand` both change if there is one command set. Options: A, change `AddUserCommand` to the actor signature and remove `AddMobCommand`, folding `trash`, `bury` and `follow` mob variants into the user handlers; B, keep both registration calls during slice 4. Recommendation: A, since slice 4 deletes `mobcommands` (spec section 4 table).


# Inventory C: one input path (Phase 4b slice 4), master `48a7d92b`

Read 33 of 33 files:

Printed with the Read tool (chunks of at most 400 lines):
1. `world.go` (in full, 725 lines, two chunks 1-400 and 400-725)
2. `internal/events/eventtypes.go` (in part, 1-120)
3. `internal/events/const.go` (in full)
4. `internal/events/events.go` (in full)
5. `internal/accounts/userrecord.go` (in part, 140-339 and 535-574)
6. `internal/mobs/mobs.go` (in part, 60-359)
7. `internal/mobcommands/mobcommands.go` (in full)
8. `internal/usercommands/usercommands.go` (in full)
9. `internal/actors/actor.go` (in full)
10. `internal/actors/controller.go` (in full)
11. `internal/actors/registry.go` (in full)
12. `internal/actors/wire.go` (in full)
13. `internal/accounts/users.go` (in part, 160-459)
14. `internal/accounts/copyover.go` (in full)
15. `main.go` (in part, 330-549, 880-1139, 1395-1444)
16. `internal/hooks/PlayerSpawn_HandleJoin.go` (in full)
17. `internal/hooks/Buff_ApplyBuffs.go` (in full)
18. `internal/usercommands/admin.command.go` (in full)
19. `internal/usercommands/admin.build.go` (in full)
20. `internal/mobcommands/emote.go` (in full)
21. `internal/inputhandlers/signals.go` (in full)
22. `internal/hooks/NewRound_IdleMobs.go` (in full)
23. `internal/hooks/MobIdle_HandleIdleMobs.go` (in full)
24. `internal/hooks/NewRound_MobRoundTick.go` (in full)
25. `internal/accounts/session.go` (in full)
26. `internal/usercommands/emote.go` (in part, 105-136)
27. `2026-10-27-phase-4b-npcs-are-players-design.md` (in full, 315 lines, one chunk)

Printed with `sed`/`cat`/`grep` in Bash (in part):
28. `internal/web/web.go` (750-770)
29. `internal/usercommands/go.go` (120-160)
30. the `internal/hooks` file defining `CleanupLinkDead` (first 60 lines, whole function)
31. `internal/keywords/keywords.go` (214-227)
32. `_datafiles/world/empty/keywords.yaml` (command-alias block, 108-152)
33. `internal/util/util.go` (76-82), `internal/guards/messaging_test.go` (grep hits only)

Everything else below is from greps whose hits are cited. Anything not checked says "unverified".

---

## Facts verified against source

| Fact | Value | Where |
|---|---|---|
| `events.Input` fields | `ActorId ids.ID`, `InputText string`, `ReadyTurn uint64`, `Flags EventFlag` | `internal/events/eventtypes.go:51-56` |
| Input listeners | exactly one: `World.HandleInputEvents` | `world.go:71` (grep `events.Input{}` finds no other `RegisterListener`) |
| Router | `HandleInputEvents` | `world.go:78-190` |
| Player/NPC split | `actors.AccountOf(input.ActorId) == nil` means mob path | `world.go:93-96`; `internal/actors/registry.go:70-80` |
| `processInput` | `world.go:502-611`, signature `(userId int, inputText string, flags events.EventFlag)` | |
| `processMobInput` | `world.go:613-653`, signature `(mobId ids.ID, inputText string)`, no flags | |
| `InputWorker` | `world.go:469-500` | |
| Per-pass trackers | `userInputEventTracker map[int]struct{}`, `mobInputEventTracker map[ids.ID]struct{}`, cleared every `EventLoop` | `world.go:50-51, 722-723` |
| `ignoreInput map[int]uint64` | written/deleted only in `HandleInputEvents`, never read to drop input | `world.go:43, 133-134, 159-160, 176-177` (grep) |
| `WorldInput.ReadyTurn` | never set or read | `world.go:34`; no `ReadyTurn:` in `main.go` (grep) |
| `EventFlag` values | `CmdSecretly 0b10`, `CmdIsRequeue 0b100`, `CmdBlockInput 0b1000`, `CmdUnBlockInput 0b10000`, `CmdBlockInputUntilComplete` = both | `internal/events/const.go:21-26` |
| `EventFlag.Add` is broken | body is `*f = (*f) &^ flag` (clears, same as `Remove`); no callers | `internal/events/const.go:9-11`; grep `Flags.Add(`/`flags.Add(` returns nothing |
| Mob registry size | 21 entries (spec facts table says 20 and omits `rest`) | `internal/mobcommands/mobcommands.go:25-47` |
| Spec facts row "Input" is stale | says `events.Input{UserId, MobInstanceId, ...}`, `world.go:89-113, 487, 598` | spec line 23 vs `eventtypes.go:51-56` |
| Plugin mob commands | `cleanup`: `trash`, `bury`; `follow`: `follow` (each also has a user command) | `modules/cleanup/cleanup.go:55-59`; `modules/follow/follow.go:61-62` |
| Direct `mobcommands` importers | `world.go`, `internal/usercommands/admin.command.go`, `internal/plugins/plugins.go`, `internal/plugins/plugincallbacks.go` | grep |
| Tests on the router | none: no `_test.go` names `HandleInputEvents`, `processInput`, `processMobInput` or `mobcommands` | grep |
| `world.go` on the connection-writer allowlist | yes, "slice 7" | `internal/guards/messaging_test.go:36` |

---

## 1. Input to command, today

### 1a. Producers of `events.Input` (non-test, all 12)

| # | Site | ActorId | ReadyTurn | Flags | Priority | Driver |
|---|---|---|---|---|---|---|
| 1 | `world.go:492` `InputWorker` | `actors.ByUserId(wi.FromId).Id` | `util.GetTurnCount()` (now) | none | 0 | player keyboard |
| 2 | `world.go:554` macro expansion in `processInput` | `user.Character.Id` | now, then +1 per `;` part | none | 0 | player |
| 3 | `internal/accounts/userrecord.go:284` `Account.Command` | `u.Character.Id` | now + `SecondsToTurns(1)*wait` | none | 0 | any code |
| 4 | `internal/accounts/userrecord.go:315` `Account.CommandFlagged` | `u.Character.Id` | same as #3 | caller's flags; calls `BlockInput()` first if `CmdBlockInput` (`:311-313`) | 0 | any code |
| 5 | `internal/mobs/mobs.go:302` `Mob.Command` | `m.Character.Id` | `m.lastCommandTurn` (see section 3) | none | 0 | any code |
| 6 | `internal/hooks/PlayerSpawn_HandleJoin.go:76` | `user.Character.Id` | `0` (immediate) | none | 0 | `Server.OnLoginCommands` config |
| 7 | `internal/hooks/Buff_ApplyBuffs.go:66` | `evt.ActorId` (mob only, `isMob && Health <= 0`) | unset = `0` | none | 0 | `suicide` on a buff kill |
| 8 | `internal/usercommands/admin.command.go:48` | `mobId` from `room.FindByName` | now, +1 per `;` part | none | 0 | admin `command` |
| 9 | `internal/usercommands/admin.command.go:57` | `targetUser.Character.Id` | same | none | 0 | admin `command` on a player |
| 10 | `internal/usercommands/admin.build.go:47` | `user.Character.Id` | unset = `0` | none | **-1** | `build zone`: `look` |
| 11 | `internal/usercommands/admin.build.go:151` | `user.Character.Id` | unset = `0` | none | **-1** | `build room`: `look` |

Rows 1-11 are 11 literal sites; grep `events\.Input\{` returns 12 lines, the 12th being the `RegisterListener` at `world.go:71`.

Notable callers of #3/#4/#5 (grep of `.Command(`): `user.CommandFlagged("look", CmdSecretly)` at `PlayerSpawn_HandleJoin.go:93`; the exit-message requeue `user.CommandFlagged(rest, flags|CmdIsRequeue|CmdBlockInputUntilComplete, 1)` at `internal/usercommands/go.go:148`; `processInput`'s own `user.Command("emote @looks a little confused")` at `world.go:597`; follow moves `user.Command(exit, .25)` / `mob.Command(exit, .25)` at `modules/follow/follow.go:307, 322`; party follow `partyUser.Command(rest)` at `usercommands/go.go:219`; charmed-mob follow `mob.Command(rest)` at `usercommands/go.go:237, 272`; mob AI in `NewRound_DoCombat.go:152, 477, 479, 574, 577, 638, 927`, `NewRound_IdleMobs.go:47, 70, 94, 99, 110`, `MobIdle_HandleIdleMobs.go:29, 40, 47`, `NewRound_MobRoundTick.go:59, 70`, `mobcommands/go.go:69, 76, 134`, `wander.go:23, 71, 79`, `equip.go:16`, `remove.go:15`, `usercommands/give.go:201, 203`; player death `user.Command("suicide")` at `NewRound_DoCombat.go:906`.

### 1b. Network to `events.Input` (players only)

1. Connection goroutines build `WorldInput{FromId: userId, InputText}` and call `worldManager.SendInput` (unbuffered channel, blocks until read: `world.go:251-253`, channel made at `world.go:57`):
   - telnet `handleTelnetConnection`: `main.go:913-919`;
   - copyover-resumed telnet `resumeRestoredConnection`: `main.go:518-522`;
   - websocket `HandleWebSocketConnection`: `main.go:1116-1122`;
   - SSH `handleSSHConnection`: `main.go:1431-1435`.
2. Client-side rate limit: telnet, resumed telnet and SSH drop a line typed within `Timing.TurnMs` of the previous one (`main.go:514-516, 883-896, 1427-1429`). **Websocket has no such throttle** (`main.go:1112-1122`). AI connections (`ConnType() == ConnAI`) additionally pass `AICommandAllowed(round, Network.AI.CommandsPerRound)` or get "Command dropped — AI rate limit" (`main.go:900-911`). The AI path is otherwise the same telnet loop.
3. `if connDetails.InputDisabled() { continue }` before handlers (`main.go:494-496`; telnet and SSH equivalents unverified line-by-line).
4. Input handlers (`TelnetIAC`, `Ansi`, `Cleanser`, `TextPrefix`, `Echo`, `History`, admin-only `SystemCommandInputHandler`, `SignalHandler`; `main.go:444-455`) transform or swallow bytes. None produce `events.Input` or call `processInput`: grep in `internal/inputhandlers` for `processInput|TryCommand|events.Input|worldInput` finds nothing; `systemcommands.go:136` queues `events.System`, `:172` a `Broadcast`. `SignalHandler` rewrites ctrl-Q/W/X to `/quit`, `/who`, `/shutdown 0` (`signals.go:42-61`), which `SystemCommandInputHandler` handles (admin only).
5. `InputWorker` (goroutine started at `main.go:355`) reads `w.worldInput`, looks up `actors.ByUserId(wi.FromId)`; on nil logs "no actor for user, input dropped" and drops it (`world.go:486-490`); else queues `events.Input{ActorId: actor.Id, InputText, ReadyTurn: util.GetTurnCount()}` (`world.go:492-496`).

Web client: `events.WebClientCommand` is outbound only (server to browser), produced at `main.go:1065` and `userrecord.go:345`, consumed by `hooks/WebClientCommand_SendWebClientCommand.go`. No web API produces input: grep of `internal/web` for `.Command(`, `CommandFlagged(`, `events.Input` finds nothing.

### 1c. The consumer: `World.HandleInputEvents` (`world.go:78-190`)

Runs inside `events.ProcessEvents` (one listener, `world.go:71`). Classifies with `actors.AccountOf(input.ActorId)` (`world.go:93`).

**Mob branch** (`AccountOf == nil`, `world.go:96-123`). Also taken by any ActorId with no live actor (dead mob, logged-out player, stale ID).
1. If `mobInputEventTracker[ActorId]` set: `CancelAndRequeue` (`:101-103`). This check is **before** the `ReadyTurn == 0` check, so a pending delayed mob command also holds back immediate ones.
2. `ReadyTurn <= 0` (i.e. `== 0`, it is `uint64`): `processMobInput` now, does not set the tracker (`:106-109`).
3. Set the tracker **before** the readiness check (`:114`), so a not-yet-ready command blocks every later command for that mob this pass. Result: strict FIFO per mob.
4. `ReadyTurn > turnCt`: `CancelAndRequeue` (`:116-118`).
5. `processMobInput(ActorId, InputText)` (`:120`). Flags are dropped.

**Player branch** (`world.go:125-189`), `userId := inputUser.UserId`.
1. `ReadyTurn == 0`: if `CmdUnBlockInput` and `ignoreInput[userId]` set, delete it and `user.UnblockInput()`; then `processInput(userId, text, flags)`; return (`:128-145`). Bypasses the tracker and does not set it.
2. If `userInputEventTracker[userId]` set: requeue (`:148-150`).
3. `ReadyTurn > turnCt`: if `CmdBlockInput`, record `ignoreInput[userId] = turnCt` once; `input.Flags.Remove(CmdBlockInput)` acts on the local copy only, the requeued event is `pe.event` (`events.go:225`), so it has no effect (harmless: the map write is idempotent). Requeue (`:154-167`). The tracker is **not** set here, so a later, ready command for the same player can run before an earlier delayed one: players are not FIFO, mobs are.
4. Ready: unblock as in step 1 (`:174-183`), `processInput` (`:185`), then set the tracker (`:187`).

**Requeue mechanics.** `CancelAndRequeue` puts the event on `requeues` (`events.go:224-226`), which is re-added at the top of the next `ProcessEvents` in the same relative order with fresh order counters (`events.go:197-200`, `reAddToQueue` `:127-157`). Events queued by a handler during a pass run in the same pass (`events.go:175, 203-230`). `World.eventRequeue` (`world.go:49, 714-720`) is never appended to (grep), so it is dead.

**Rate in practice.** The tracker is cleared every `EventLoop` (`world.go:722-723`), which runs every 1 ms (`eventLoopTimer`, `world.go:339, 397-403`). The comment "If an event was already processed for this user this turn" (`:147`) is wrong: the limit is one non-immediate command per actor per event-loop pass, not per turn. Per-turn pacing comes from the client throttle (1b.2) for typed input and from `Mob.Command`'s one-turn spacing (section 3) for mobs.

**Input blocking.** `Account.BlockInput()` sets `Session.inputBlocked` at enqueue time (`userrecord.go:292-294, 311-313`). `usercommands.TryCommand` refuses silently (`return true, nil`) any command that is not `AllowedWhenDowned` while `user.InputBlocked()` (`usercommands.go:227-238`). Unblock is only via `CmdUnBlockInput` when `ignoreInput[userId]` exists (`world.go:131-138, 174-182`). Edge: a `CmdBlockInputUntilComplete` command whose first sighting in `HandleInputEvents` is already ready (wait under one turn and first seen with no tracker set) never sets `ignoreInput`, so `UnblockInput` never runs and the player stays blocked. The only producer, `go.go:148`, waits 1 s; whether `SecondsToTurns(1)` is always at least 1 is unverified. Mobs have no blocking (`processMobInput` takes no flags).

**`CmdSecretly`** is only read by `usercommands/look.go:23`; passed through `processInput` to `TryCommand` to the command (`world.go:580`, `usercommands.go:253`). `CmdIsRequeue` is read by `usercommands/go.go:146`. Mobs lose all flags.

**Link-dead and disconnect.** Read error with `LinkDeadSeconds > 0`: `SendSetLinkDead` (`main.go:481-483, 982-985`) then `accounts.SetLinkDeadUser` on the world loop (`world.go:450-458`; `users.go:335-371`): account stays in `userManager.Users`, so its actor stays and `AccountOf` still returns it; queued and delayed inputs keep running as player input. Prompt redraw `connections.SendTo(..., user.ConnectionId())` then targets a removed connection id (outcome unverified). Expiry: `CleanupLinkDead` on `NewTurn` queues `System{leaveworld}` (`hooks` file, item 30), which queues `PlayerDespawn` (`world.go:216-226`). Logout: `LogOutUserByConnectionId` (`users.go:389-423`) fires `OnDeactivate` (`:410`), which `unbindAccount`s the actor (`wire.go:21, 40-44`; `registry.go:186-200`). Any input still queued for that body then has no actor, takes the **mob** branch, `mobs.GetInstance` is nil, `RecentlyDied` is false, so it logs `"Mob not found"` at error level (`world.go:616-621`) and drops. Without link-dead (`LinkDeadSeconds == 0`) this happens at once on disconnect (`main.go:484-486`).

**Copyover.** Registered contributors: connections, accounts, util, gametime, tokens (`main.go:109-113`); none saves the event queue, so queued and delayed input (including delayed mob commands) is lost on copyover. Accounts restore with `SetLastInputRound(now)` (`copyover.go:72-73`) and link-dead state (`:78-81`), then `activated(user)` binds actors (`:84`). `Mob.lastCommandTurn` is reset by `mobs/saved.go:61` on restore (grep; `saved_test.go:75-88`).

---

## 2. `processInput` and `processMobInput`, step by step

### `processInput` (`world.go:502-611`). [A] = account-only (needs `Account`/`Session`).

1. `accounts.GetByUserId(userId)`; nil logs "User not found" and returns (`:504-508`).
2. [A] Prompt answering: `user.GetPrompt()` (Session `activePrompt`, `userrecord.go:558-561`). If a pending `prompt.Question` exists, `Answer(inputText)`, then replace the input with `cmdPrompt.Command + " " + cmdPrompt.Rest` (re-invokes the command that owns the prompt), else empty. If a prompt has no pending question, `ClearPrompt()` (`:510-528`). Prompts are started only by `StartPrompt` in `admin.item/mob/room/server/zone.go`, `password.go`, `start.go` (grep, 10 call lines in 8 files incl. `userrecord.go`).
3. `strings.TrimSpace(inputText)` (`:536`). Mobs do not trim.
4. If non-empty:
   - [A] `user.SetLastInputRound(util.GetRoundCount())` (`:542`). This fires for **every** player input, including internally queued ones (macros, follow moves, `OnLoginCommands`, admin `command`, the "confused" emote), so AFK (`userrecord.go:590`) and `InactivePlayers` (`hooks/NewRound_InactivePlayers.go:40`) are reset by non-typed input today.
   - [A] Macros: if `user.Macros != nil && len(inputText) == 2` and it matches, split on `;`, queue each as `events.Input{ActorId: user.Character.Id, ReadyTurn: now+i}`, `handled = true` (`:545-563`).
   - Shortcut split: if `len > 1` and first char is backtick or `.`, insert a space ("`x" to "` x") (`:568-572`). The keyword aliases then map backtick to `broadcast` and `.` to `say` (`keywords.yaml:108-109`). Mechanically actor-neutral; mobs don't get it today.
   - Command/rest split at the first space, command lowercased (`:574-578`). Same as mob (`:632-636`).
   - `usercommands.TryCommand(command, remains, userId, flags)`; error logged as warn `user-TryCommand` (`:580-583`).
5. Empty input: [A] `connections.SendTo(AnsiParse(user.GetCommandPrompt()), user.ConnectionId())` (`:586-589`).
6. Not handled and `command != ""`: `badinputtracker.TrackBadCommand(command, remains)` (`:594`; only caller, grep), `user.SendText("<cmd> not recognized. Type help for commands.")` (`:596`, [A] interface text), `user.Command("emote @looks a little confused")` (`:597`; the `@` makes `usercommands.Emote` skip the "You Emote:" echo, `usercommands/emote.go:123-124`).
7. [A] If `hadPrompt` or a prompt now exists, redraw the prompt with `connections.SendTo` (`:602-605`). (`world.go` is on the guard allowlist for slice 7.)

### `processMobInput` (`world.go:613-653`)

1. `mobs.GetInstance(mobId)`; nil: if not `mobs.RecentlyDied(mobId)` log "Mob not found" error; return (`:616-622`). `RecentlyDied` keeps entries 15 rounds, pruned only when more than 30 are tracked (`mobs.go:106-120`); `TrackRecentDeath` is called only from `mobcommands/suicide.go:36` (grep).
2. No trim, no prompts, no macros, no shortcut split, no last-input, no bad-input tracking, no prompt redraw.
3. Split, lowercase command (`:632-636`), `mobcommands.TryCommand(command, remains, mobId)`; error logged as warn `mob-TryCommand` (`:640-643`).
4. Not handled and `command != ""`: `mob.Command("emote looks a little confused (<cmd> <rest>).")` (`:647-651`). Room observers see the raw failed command text, a debugging aid that leaks to players.

---

## 3. `Mob.Command`, `Account.Command`, `mobcommands.TryCommand`

### `Mob.Command(inputTxt string, waitSeconds ...float64)` (`internal/mobs/mobs.go:279-310`)

- `readyTurn = max(now, m.lastCommandTurn)` (`:281-291`).
- `turnDelay = SecondsToTurns(1) * waitSeconds[0]` (`:293-295`).
- Splits on `;`; for part `i`, `m.lastCommandTurn = readyTurn + turnDelay + i`, queue `Input{ActorId: m.Character.Id, InputText: cmd, ReadyTurn: m.lastCommandTurn}` (`:297-308`). Parts are not trimmed.
- Effect: every mob command is scheduled at or after the previous one, so a mob has a per-actor command queue in time order. Two calls in the same turn with no wait land on turns T and T (first: `readyTurn = now`, `lastCommandTurn = now`; second: `readyTurn = lastCommandTurn = now`), so they are not spaced by a turn across calls, only within one `;` string; FIFO comes from the router tracker (1c mob step 3).
- `lastCommandTurn` field: `mobs.go:80`; reset on restore `saved.go:61`; no other readers (grep).
- `Mob.Sleep(seconds)` = `Command("noop", seconds)` (`mobs.go:275-277`); with the tracker this stalls every later mob command until it runs. It has no callers (grep `\.Sleep(` excluding `time.Sleep` returns nothing).
- Idle and boredom: `IdleMobs` does not consult pending commands; it issues `Command` for boredom `rest depression` (`NewRound_IdleMobs.go:44-55`), lost quarry (`:65-76`), path steps (`:80-124`), else queues `MobIdle` (`:126`), whose handler issues `pathto home` or an idle command at `ActivityLevel` odds (`MobIdle_HandleIdleMobs.go:27-49`). `PreventIdle` skips one round (`NewRound_IdleMobs.go:58-61`), set by combat (`NewRound_DoCombat.go:470, 819`). `BoredomCounter` reset when a player is seen (`rooms.go:2070`). Because of the FIFO tracker, idle commands queue behind any pending delayed mob command.

### `Account.Command(inputTxt string, waitSeconds ...float64)` (`userrecord.go:277-290`) and `CommandFlagged` (`:304-322`)

- `ReadyTurn = now + SecondsToTurns(1)*wait`. No `;` split, no `lastCommandTurn`, no per-account queue. Two delayed commands are ordered only by `ReadyTurn`, and an immediate one can overtake a delayed one (router player step 3).
- `CommandFlagged` calls `BlockInput()` up front when `CmdBlockInput` is set.

### `mobcommands.TryCommand(cmd, rest string, mobId ids.ID)` (`internal/mobcommands/mobcommands.go:60-145`)

1. Lowercase `cmd`, trim `rest` (`:62-63`).
2. `cmd = keywords.TryCommandAlias(cmd)` (`:65`), a plain map lookup (`keywords.go:214-220`).
3. Mob lookup (nil: error "mob instance doesn't exist"), `rooms.LoadRoom` (nil: error) (`:69-77`).
4. `mobDisabled = mob.Character.IsDisabled()` (`:79`). Room props block commented out (`:82-90`).
5. Alias again with multi-word split (`:92-108`). Because step 2 already replaced `cmd`, this only fires on an alias of an alias; a multi-word alias (for example `pchat` to `party chat`, `keywords.yaml:127`) stays as the whole string `party chat` and matches nothing. Harmless today (no mob command is multi-word).
6. Registry hit: if disabled and not `AllowedWhenDowned`, return `true, nil` silently (`:110-114`); else time and run (`:116-122`).
7. No registry hit and not disabled: try `Go(cmd, mob, room)` as an exit name (`:126-138`).
8. `emoteAliases[cmd]` (mob-only table of 80 pre-written emotes, `mobcommands/emote.go:11-92`): `Emote(text)` (`:139-142`). Mob `Emote` sends nothing when `room.PlayerCt() < 1` (`emote.go:98-100`).
9. Else `false, nil` (`:144`), leading to the "confused" emote in `processMobInput`.
- Plugin commands: `RegisterCommand(cmd, fn, isBlockable)` writes the same map (`:148-153`); called from `plugins.go:486` (plugins register via `AddMobCommand`, `plugins.go:311-320`).
- No admin, no permission, no plaintext-password, no `cancel-on-action` buff check, no `look me` self-keyword rewrite, no exit-name-first check.

### `usercommands.TryCommand(cmd, rest, userId, flags)` (`internal/usercommands/usercommands.go:139-279`), for comparison

Order: account lookup, room (`:142-150`); exit name first: `cmd` becomes `go` (`:155-158`); else [A] account alias `user.TryCommandAlias` (`:161-177`), keyword alias with multi-word split (`:179-195`), `IsDisabled` (`:197`); `CancelBuffsWithFlag("cancel-on-action")` (`:202`); `look` self-keywords (`:206-217`); [A] plaintext-password gate (`:219-223`); registry: not `AllowedWhenDowned` and disabled and not admin-only gives "You are unable to do that while downed." (`:230-233`); [A] `InputBlocked` silent refusal (`:236-238`); [A] admin-only via `HasRolePermission` (`:241-259`); `emoteAliases` (`:262-265`); `Go` fallback (`:268-275`); `false, nil`.

Differences the merge must settle: order (exit-first vs registry-first), downed refusal text (players get a message, mobs silence), the `cancel-on-action` buff cancel (players only today), and that disabled mobs skip the `Go` fallback while disabled players reach `Go` and are refused inside it (unverified inside `Go`).

### Name overlap of the two registries

- In both (14): `broadcast, drop, emote, equip, get, give, go, look, noop, put, remove, say, shout, suicide` (`mobcommands.go:26-46`, `usercommands.go:52-121`). `AllowedWhenDowned` differs for `broadcast` (user true / mob false), `drop` (true/false), `look` (true/false); same for the rest checked (`emote` true/true, `noop` true/true, `say` true/true, `shout` true/true, `suicide` true/true, `give/go/get/equip/put/remove` false/false).
- Mob-only (7): `despawn, pathto, replyto, rest, sayto, saytoonly, wander`. The spec names `wander`, `pathto`, `despawn`, `noop` as behaviour actions; `rest` (boredom unload, decision 15) is not named there.
- Same name, different meaning: `suicide` (mob death and loot vs player death stand-in), `rest` (mob unload; no user `rest`).
- Plugin pairs (user + mob variant): `trash`, `bury`, `follow`.
- `admin.command.go:29` lists `mobcommands.GetAllMobCommands()` in its help; it needs a new source after the merge.

---

## 4. Everything else keyed on the input source

- **`actors`**: `AccountOf` (`registry.go:70-80`) is the router's only classifier. `Actor` has `Id`, `Character`, `controller`; no `Mob` field yet (`actor.go:16-21`). The spec's note (spec line 266-268) says to put the body on the `Actor` for slices 4-5 rather than asserting the controller type. Today the merged path can still reach the mob by `mobs.GetInstance(actor.Id)`.
- **Lifecycle hooks**: `OnActivate`/`OnDeactivate` bind and unbind player actors; `OnInstanceCreated`/`OnInstanceDestroyed` bind and unbind NPC actors (`wire.go:18-58`). `DestroyInstance` deletes from `mobInstances` then fires `OnInstanceDestroyed` (`mobs.go:258-262`; also `saved.go:92` for resting NPCs).
- **`RecentlyDied`**: only used to silence the "Mob not found" log (`world.go:618`).
- **Copyover**: see 1c. No queued input survives.
- **AI/bot connections**: same telnet path, extra per-round limit (`main.go:900-911`); account must be `IsAI` on the AI port (`main.go:857-859`).
- **Web client**: no inbound command path besides the websocket game socket (1b).
- **`inputhandlers`**: nothing calls `processInput` or queues `events.Input` (1b.4).
- **`InputWorker`**: player-only by construction (`actors.ByUserId`).
- **Permadeath reroll**: `actors.BodyReplaced` changes the player's ActorId (`registry.go:171-181`); input still queued under the old body ID takes the mob branch and logs "Mob not found" (inferred from `world.go:93-96, 616-621`; live path unverified, spec line 284-286).
- **`Buff_ApplyBuffs`** and **`RoomChange_CleanupEphemeralRooms`**, `gmcp`, `follow` also classify by `actors.AccountOf` (grep), but not for input.

---

## 5. Concurrency

- `HandleInputEvents`, `processInput`, `processMobInput` and every command run on the **MainWorker** goroutine (`main.go:356`), inside `w.EventLoop()` under `util.LockMud()` (`world.go:397-403`; `LockMud` is a plain `sync.Mutex`, `util.go:76-82`). The other `LockMud` holders are the MainWorker's own timer cases (`world.go:357-458`) and web handlers via `RunWithMUDLocked` (`web.go:754-766`).
- `InputWorker` runs on its own goroutine (`main.go:355`) without `LockMud`; it takes `actors.mu.RLock` (`registry.go:85-86`) and `events.qLock` (`events.go:92-93`).
- Locks taken inside the router: `actors.mu.RLock` twice (`registry.go:63`, `actor.go:25`), released before return; `accounts.userManagerMu.RLock` in `GetByUserId` (`users.go:172`); `qLock` per `AddToQueue`. `ProcessEvents` drops `qLock` around each listener (`events.go:221-228`), so re-entrant `AddToQueue` is safe. None of these are nested with each other in the router today.
- Login goroutines take `lifecycleMu` then `userManagerMu`, release it, then fire `OnActivate`, which takes `actors.mu` (`users.go:267-333`; `wire.go:28`; `registry.go:133-136`). Nothing takes `userManagerMu` while holding `actors.mu` (`controller.Account()` is a field read, `controller.go:33`).
- `mobInstances`, `Mob.lastCommandTurn`, `w.ignoreInput` and both trackers are unsynchronized maps/fields that rely on `LockMud`. All `Mob.Command` callers found are on the world loop (grep: hooks, commands, `follow`; no off-loop callers in `main.go`/`inputhandlers`).
- **Merging changes no lock order** provided the merged handler keeps the pattern "one `actors` lookup, release, then `accounts` lookup" and does not call `accounts.GetByUserId` while holding `actors.mu`. Reaching the account through `AccountOf(id)` (already returned) rather than `GetByUserId` removes the `userManagerMu` read entirely. The existing slice-1 hazard stays: `Actor.Character` and the controller can be rebound from a login goroutine without `LockMud` (spec line 272-275).

---

## Open questions for Calabe (each needs a ruling)

1. **Ordering of one actor's commands.** Today mobs are strict FIFO (a delayed command holds back everything after it) and players are not (an immediate command overtakes a delayed one; `ReadyTurn 0` bypasses everything). Player-facing effect: with FIFO, a player who types while a 1 s exit message is pending (the `go.go:148` requeue) waits for it; without FIFO, a mob's idle emote can run before its delayed combat action.
   - A: one rule, FIFO for every actor (mob behaviour). Changes player feel for delayed commands; `BlockInput` becomes mostly redundant.
   - B: one rule, player behaviour for every actor. NPC delayed actions (`noop` sleeps, combat waits at `NewRound_DoCombat.go:577`) stop holding back the queue; mob behaviour changes.
   - C: keep both, keyed on whether the actor has an account. Keeps parity but the router still branches on player vs NPC.
   - Recommendation: C for slice 4 (behaviour parity, decision 2), keyed on a controller question, not a type; record the asymmetry for slice 6.

2. **Account-only steps in the merged path** (prompts, macros, last-input round, bad-input tracker, "not recognized" text, prompt redraw). These only make sense with an `Account`.
   - A: one function that runs them when `actor.Controller().Account() != nil`.
   - B: move them onto `ConnectionController` (an input-preprocess method on `Controller`), so the router has no account checks.
   - Recommendation: A in slice 4 (smallest change, spec says "call `Account()` and handle nil"); B when possession lands.

3. **Unknown command from an NPC.** Players get "not recognized" plus `emote @looks a little confused`; mobs show observers `looks a little confused (<cmd> <rest>).`
   - A: NPCs get the player's emote (no raw text leaked).
   - B: keep the mob's version.
   - C: NPCs log at warn and do nothing visible.
   - Recommendation: C. The leaked text is a debugging aid; a broken NPC behaviour should be in the log, not the room.

4. **Downed refusal for NPCs.** Players get "You are unable to do that while downed."; mobs are refused silently. With one registry the text goes through `SendText`/messaging, which a `BehaviourController` drops as Interface text anyway. Recommendation: one rule (send the interface text); the NPC never sees it.

5. **`AllowedWhenDowned` conflicts** for `broadcast`, `drop`, `look` (user true, mob false). Recommendation: take the user value; a downed NPC rarely issues these.

6. **`rest` and `suicide` name clashes.** Mob `rest` is the boredom unload (decision 15) and mob `suicide` is NPC death; the user `suicide` is player death.
   - A: make both engine calls (like `despawn`), not commands.
   - B: one `suicide` command branching on "does this actor return from a template" (spec section 3 death rule).
   - Recommendation: A for `rest` (it is the memory valve, not an action); B for `suicide` until slice 5 settles death, since `NewRound_MobRoundTick.go:70`, `NewRound_DoCombat.go:906, 927` and `Buff_ApplyBuffs.go:66` all dispatch it as input.

7. **Input for a body with no actor** (logged-out player, dead mob, rerolled body). Today it takes the mob branch and logs "Mob not found" at error level unless `RecentlyDied`. Recommendation: drop it at debug level in the router when `actors.Get` is nil, and delete `RecentlyDied` if nothing else needs it (it has one reader, `world.go:618`).

8. **Last-input round from non-typed input.** Today every player input, including follow moves and macros, resets AFK. Recommendation: keep as is in slice 4 (parity), note it for later; fixing it means `InputWorker` stamping the round instead.

9. **Shortcuts and `TrimSpace` for NPCs.** Harmless to apply to all actors. Recommendation: apply to all (one path), no behaviour change for existing mob scripts (none start with backtick or dot; unverified across `_datafiles`).

10. **Small bugs found, fix in slice 4 or leave?** `EventFlag.Add` clears instead of sets (`const.go:9-11`, no callers); `input.Flags.Remove` on a copy (`world.go:163`); dead `World.eventRequeue` and `WorldInput.ReadyTurn`; the "per turn" comment that is really per 1 ms pass; double keyword alias in `mobcommands.TryCommand` (goes with the package); possible stuck `inputBlocked` for a sub-turn `CmdBlockInputUntilComplete`. Recommendation: fix `Add` and delete the dead fields in slice 4 (they sit in the code being merged); leave the stuck-block edge as a noted risk with a test.
