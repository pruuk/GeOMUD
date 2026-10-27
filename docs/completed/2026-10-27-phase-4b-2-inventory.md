# Phase 4b slice 2 ("one ID space"): fact inventory

Repo `new_geo_mudbase`, `master` at `df1b04c6`. Read-only pass, 2026-10-27.

## Read-files line

Read 86 of 86 files opened (79 fully, 7 partially). Every file was read with the Read tool in chunks of at most 400 lines.

Fully read (79):
`2026-10-27-phase-4b-npcs-are-players-design.md`, `world.go`, `copyover.go`,
`internal/actors/{registry,wire,actor,controller}.go`,
`internal/accounts/{users,copyover}.go`,
`internal/characters/{aggro,character,charminfo,kdstats}.go`,
`internal/events/eventtypes.go`,
`internal/mobs/mobs.go`,
`internal/rooms/{rooms,roommanager,roomdetails}.go`,
`internal/hooks/{NewRound_DoCombat,Buff_ApplyBuffs,Message_SendMessages,MobIdle_HandleIdleMobs,NewRound_IdleMobs,NewRound_MobRoundTick,NewRound_UserRoundTick,NewTurn_PruneBuffs,PlayerDespawn_HandleLeave,PlayerSpawn_HandleJoin,hooks,RoomChange_CleanupEphemeralRooms,PlayerDrop_HandlePlayerDrop,RedrawPrompt_SendRedraw,UserSettingChanged_ClearSettingCaches}.go`,
`internal/combat/{combat,simulate,calculations}.go`,
`internal/mobcommands/{mobcommands,despawn,drop,emote,equip,get,give,go,look,put,remove,say,sayto,wander,suicide}.go`,
`internal/usercommands/{admin.build,admin.command,admin.grant,admin.locate,admin.mute,admin.mob,admin.teleport,attack,consider,give,go,look,party,shoot,show,start,suicide}.go`,
`internal/mapper/mapper.config.go`, `internal/templates/templatesfunctions.go`,
`internal/web/{api_v1_mobs,api_v1_characters}.go`,
`modules/cleanup/cleanup.go`, `modules/follow/follow.go`,
`modules/gmcp/{gmcp.Char,gmcp.Comm,gmcp.Party,gmcp.Room,gmcp.Game}.go`.

Partially read (7), and why:
- `internal/mapper/mapper.go` lines 600-729: its only pattern hit (line 633) is `Config.HasVisited`, a different method from `Room.HasVisited`; I read the surrounding function and `mapper.config.go` in full to confirm.
- `internal/accounts/userrecord.go` lines 140-369: only the event producers (`ShorthandId`, `GrantXP`, `Command`, `CommandFlagged`, `AddBuff`, `SendText`) were classified.
- `internal/rooms/save_and_load.go` lines 200-334: room copy (`players`/`mobs` copied) and `SaveRoomInstance` (persistence check).
- `internal/rooms/ephemeral.go` lines 215-239, `internal/rooms/admin.go` lines 238-257 and 580-659: the `len(r.players)` readers found by a follow-up grep.
- `modules/gmcp/gmcp.go` lines 170-319: `sendGMCPEvent`, `handlePlayerJoin`, web GMCP request dispatch.
- `internal/web/api_v1_users.go` lines 1-134: GET/PATCH user JSON.

Cited from grep only (file:line given, file not read, not classified on its own): event producers in `internal/usercommands/{say,whisper,broadcast,drop,equip,get,put,remove,lock,unlock,flee,print,set}.go`, `internal/hooks/{NewRound_AutoHeal,Broadcast_SendToAll,NewRound_InactivePlayers,NewTurn_CleanupLinkDead}.go`, `internal/inputhandlers/systemcommands.go`, `internal/parties/parties.go`, `internal/plugins/plugins.go`, `modules/gmcp/gmcp.Mudlet.go`, `internal/rooms/{spawninfo,corpse}.go` (struct definitions via grep -A), `_datafiles/html/public/static/js/windows/window-room.js` (lines 497-521), `_datafiles/html/admin/docs-modules.html` (lines 1124-1228).

**Count check.** Non-test `MobInstanceId` = 157 uses in 37 files (matches the brief). The spec's facts table said 148 at `c4314b7f`; the difference is the 9 uses slice 1 added in `internal/actors/registry.go`. Test files that use these names: `internal/actors/{registry,wire}_test.go`, `internal/characters/character_test.go`, `internal/mobs/lifecycle_test.go`.

**Pattern noise in the counts below** (checked by reading):
- `Visitors(` also matches `PruneVisitors(` and `HasRecentVisitors(`. Real `Room.Visitors(` calls: `roomdetails.go:372,415`, `rooms.go:1896,1940`.
- `GetMobs(` in `internal/web/api_v1_mobs.go` is `apiV1GetMobs`, a false positive. That file serves mob templates; see D.
- `HasVisited(` in `internal/mapper/*` is `Config.HasVisited`, a false positive.
- `MobInstanceId` in `usercommands/party.go` and `gmcp.Party.go` is a local variable name (`charmedMobInstanceIds`).
- Every `\.InstanceId` hit is a mob instance id: either `Mob.InstanceId`, `SpawnInfo.InstanceId` (a mob instance id) or `MobDeath.InstanceId`.

## Table A: per-file counts and fate

Abbreviations: MII = `MobInstanceId`, IID = `.InstanceId`, GI = `mobs.GetInstance(`. Other columns list only the patterns with hits.

"Fate" names the later slice that deletes or rewrites the file. Slice 2's Done-when line is "`MobInstanceId` absent from non-test code", so **slice 2 must still edit every file in this table that contains MII, including the ones slices 4, 5 and 6 delete later**.

| File | MII | IID | GI | Other patterns | Fate |
|---|---|---|---|---|---|
| internal/accounts/users.go | | | 1 | MobInstanceExists 3, CharmedMobs 1, GetCharmIds 3, TrackCharmed 3 | survives |
| internal/actors/registry.go | 9 | 3 | | | survives (`byMobInstanceId` index goes in slice 2) |
| internal/actors/wire.go | | 1 | | | survives |
| internal/characters/aggro.go | 2 | | | | survives |
| internal/characters/character.go | 7 | | | KillerMII 1, CharmedMobs 6, GetCharmIds 1, TrackCharmed 1, Charmed.UserId 3, GetCharmedUserId 1, IsCharmed 1, Aggro.UserId 3, Aggro.MII 1, SetAggro 2, SetAggroRemote 1, IsAggro 1 | survives |
| internal/combat/calculations.go | | | | IsAggro 1 (`ChanceToTame`, :373) | survives |
| internal/combat/combat.go | 1 | | | GetCharmedUserId 1, Aggro.UserId 1, Aggro.MII 1, SetAggro 1 | slice 5 rewrites the 4 `Attack*` variants; the MII hit (:257) is in `calculateCombat`, which survives |
| internal/combat/simulate.go | | 2 | | SetAggro 2 | survives |
| internal/events/eventtypes.go | 11 | | | | survives |
| internal/hooks/Buff_ApplyBuffs.go | 6 | | 1 | | survives |
| internal/hooks/Message_SendMessages.go | | | | GetPlayers 1 | survives (slice 3/7 messaging) |
| internal/hooks/MobIdle_HandleIdleMobs.go | 1 | | 1 | IsCharmed 1 | slice 6 deletes |
| internal/hooks/NewRound_DoCombat.go | 38 | 18 | 10 | AllIds 1, KillerMII 1, IsCharmed 1, Aggro.UserId 10, Aggro.MII 10, SetAggro 5, GetPlayers 1, GetMobs 2, MoveToRoom 1 | slice 5 rewrites |
| internal/hooks/NewRound_IdleMobs.go | 3 | 1 | 1 | AllIds 1, Aggro.UserId 2 | slice 6 deletes |
| internal/hooks/NewRound_MobRoundTick.go | 2 | 1 | 1 | AllIds 1, TrackCharmed 1, IsCharmed 1 | slice 6 deletes |
| internal/hooks/NewRound_UserRoundTick.go | | | | GetPlayers 1 | slice 6 rewrites (UserRoundTick/MobRoundTick split) |
| internal/hooks/NewTurn_PruneBuffs.go | 1 | | 1 | AllIds 1, GetPlayers 1 | survives |
| internal/hooks/PlayerDespawn_HandleLeave.go | | | 1 | IsCharmed 1, GetMobs 1, RemovePlayer 1 | survives |
| internal/hooks/PlayerSpawn_HandleJoin.go | | | | MoveToRoom 1 | survives |
| internal/mapper/mapper.config.go | | | | HasVisited 1 (false positive) | n/a |
| internal/mapper/mapper.go | | | | HasVisited 1 (false positive) | n/a |
| internal/mobcommands/despawn.go | | 2 | | RemoveMob 1 | slice 4 deletes |
| internal/mobcommands/drop.go | 1 | 1 | | | slice 4 deletes |
| internal/mobcommands/emote.go | | | | PlayerCt 1 | slice 4 deletes |
| internal/mobcommands/equip.go | 1 | 1 | | | slice 4 deletes |
| internal/mobcommands/get.go | 1 | 1 | | | slice 4 deletes |
| internal/mobcommands/give.go | 3 | 3 | 1 | FindByName 1 | slice 4 deletes |
| internal/mobcommands/go.go | | 4 | | GetPlayers 1, AddMob 2, RemoveMob 2 | slice 4 deletes |
| internal/mobcommands/look.go | | | 1 | FindByName 1, PlayerCt 1 | slice 4 deletes |
| internal/mobcommands/mobcommands.go | | | 1 | | slice 4 deletes |
| internal/mobcommands/put.go | 1 | 1 | | | slice 4 deletes |
| internal/mobcommands/remove.go | 2 | 2 | | | slice 4 deletes |
| internal/mobcommands/say.go | 1 | 1 | | PlayerCt 1 | slice 4 deletes |
| internal/mobcommands/sayto.go | 4 | 4 | 2 | FindByName 3, PlayerCt 3 | slice 4 deletes |
| internal/mobcommands/suicide.go | 1 | 10 | | TrackCharmed 1, Charmed.UserId 1, Aggro.MII 1, RemoveMob 3 | slice 4 deletes the file, but mob death, XP and loot logic must move somewhere (see E13) |
| internal/mobcommands/wander.go | | | | IsCharmed 1, PlayerCt 1 | slice 4 deletes (wander becomes a behaviour action) |
| internal/mobs/mobs.go | 4 | 6 | | AllIds 1 | survives (`instanceCounter` goes in slice 2) |
| internal/rooms/roomdetails.go | | | 2 | IsCharmed 1, Visitors 2 | survives |
| internal/rooms/roommanager.go | | 3 | 1 | GetPlayers 2, AddPlayer 1, RemovePlayer 1, MarkVisited 2, Visitors 3 (noise) | survives |
| internal/rooms/rooms.go | 3 | 11 | 12 | IsCharmed 3, Aggro.UserId 4, Aggro.MII 2, GetPlayers 6, GetMobs 6, FindByName 1, AddPlayer 1, RemovePlayer 1, AddMob 1, RemoveMob 1, ArePlayersAttacking 2, AreMobsAttacking 2, MarkVisited 3, Visitors 6 (2 real), HasVisited 1, PlayerCt 2, ApplyBuffIdToPlayers 2, ApplyBuffIdToMobs 2 | survives |
| internal/templates/templatesfunctions.go | | | 1 | | survives (`idToMobCharacter` template func) |
| internal/usercommands/admin.build.go | | | | MoveToRoom 2 | survives |
| internal/usercommands/admin.command.go | 1 | | | FindByName 1 | survives (slice 4 changes the Input it sends) |
| internal/usercommands/admin.grant.go | 3 | | | FindByName 1 | survives |
| internal/usercommands/admin.locate.go | 1 | | 1 | AllIds 1 | survives |
| internal/usercommands/admin.mob.go | | 1 | | AddMob 1 | survives |
| internal/usercommands/admin.mute.go | | | | FindByName 2 | survives |
| internal/usercommands/admin.teleport.go | | 2 | 1 | IsCharmed 1, GetPlayers 1, GetMobs 1, AddMob 1, RemoveMob 1, MoveToRoom 2 | survives |
| internal/usercommands/attack.go | 15 | 2 | 2 | IsCharmed 1, Aggro.UserId 6, Aggro.MII 2, SetAggro 2, GetPlayers 3, GetMobs 3, FindByName 1 | survives |
| internal/usercommands/consider.go | | | 1 | FindByName 1 | survives |
| internal/usercommands/give.go | 1 | 1 | 1 | FindByName 1 | survives |
| internal/usercommands/go.go | | | 3 | IsCharmed 2, Aggro.UserId 1, GetPlayers 1, GetMobs 3, MoveToRoom 1 | survives |
| internal/usercommands/look.go | | | 1 | FindByName 1, PlayerCt 1 | survives |
| internal/usercommands/party.go | 3 (var name) | | 1 | GetCharmIds 1, FindByName 1 | survives |
| internal/usercommands/shoot.go | 6 | | 1 | IsCharmed 1, SetAggroRemote 2, FindByName 1 | survives |
| internal/usercommands/show.go | | | 1 | FindByName 1 | survives |
| internal/usercommands/start.go | | | | MoveToRoom 2 | survives |
| internal/usercommands/suicide.go | 3 | | 1 | KillerMII 3, MoveToRoom 2 | survives |
| internal/web/api_v1_mobs.go | | | | GetMobs 1 (false positive) | n/a, but see D5 |
| modules/cleanup/cleanup.go | 1 | 1 | | | survives (its `mobTrashCommand` is a mob command, so slice 4 converts it) |
| modules/follow/follow.go | 2 | 1 | 4 | FindByName 2 | survives (it registers a mob command, slice 4; and a `MobIdle` listener, slice 6) |
| modules/gmcp/gmcp.Char.go | 4 | 1 | 1 | Aggro.MII 2, GetMobs 1 | survives |
| modules/gmcp/gmcp.Comm.go | 3 | | 1 | GetPlayers 1 | survives |
| modules/gmcp/gmcp.Party.go | 4 (1 field, 3 var) | | 1 | GetCharmIds 1 | survives |
| modules/gmcp/gmcp.Room.go | 1 | | 1 | GetPlayers 6, GetMobs 1 | survives |
| world.go | 6 (1 in a comment) | | 1 | MoveToRoom 1 | `HandleInputEvents` mob branch (:89-116) and `processMobInput` (:598-638): slice 4 merges and deletes them. The rest survives |

Totals across all files: MII 157, IID 85, GI 61, MobInstanceExists 3, GetAllMobInstanceIds 6, KillerMobInstanceId 5, CharmedMobs 7, GetCharmIds 6, TrackCharmed 6, Charmed.UserId 4, GetCharmedUserId 2, IsCharmed 15, Aggro.UserId 27, Aggro.MobInstanceId 19, SetAggro 12, SetAggroRemote 3, IsAggro 2, GetPlayers 25, GetMobs 19, FindByName 19, AddPlayer 2, RemovePlayer 3, AddMob 5, RemoveMob 8, ArePlayersAttacking 2, AreMobsAttacking 2, MarkVisited 5, Visitors 11 (4 real calls), HasVisited 3 (1 real), PlayerCt 10, ApplyBuffIdToPlayers 2, ApplyBuffIdToMobs 2, rooms.MoveToRoom 12.

MII by fate: 59 sit in files slices 4, 5 and 6 delete or rewrite (mobcommands 15, DoCombat 38, IdleMobs 3, MobRoundTick 2, MobIdle 1). The other 98 survive, including 5 in `world.go` that slice 4 merges.

## Table B: event types with id fields

Producers and consumers come from `grep -rnE "events\.<Type>(\{|\)|\b[^A-Za-z])"`. That pattern matches both construction (`events.X{`) and type assertion (`e.(events.X)`), so an empty result means none exist.

"Body" means the field refers to a live player or mob, so it converts to `ActorId`. "Account" means it stays `UserId`. A connection-only target (prompt, GMCP, direct text) is listed as account, because it routes to the connection.

| Event | Field | Class | Evidence (consumer use) | Producers | Consumers |
|---|---|---|---|---|---|
| Buff | UserId, MobInstanceId | body (merge into one ActorId) | `ApplyBuffs` resolves either one to a Character and applies the buff (Buff_ApplyBuffs.go:33-50) | userrecord.go:330 (`Account.AddBuff`); mobs.go:272 (`Mob.AddBuff`) | hooks/Buff_ApplyBuffs.go:18 |
| BuffsTriggered | UserId, MobInstanceId | body; the only consumer needs the connection | gmcp.Char.go:101-113 ignores MobInstanceId and sends Char.Affects to UserId | Buff_ApplyBuffs.go:75; NewRound_MobRoundTick.go:43; NewRound_UserRoundTick.go:114; NewTurn_PruneBuffs.go:52 | gmcp.Char.go:101 |
| BuffsAdded | UserId, MobInstanceId | body, GMCP-only consumer; MobInstanceId is never set | gmcp.Char.go:139-152 | Buff_ApplyBuffs.go:78 (UserId only) | gmcp.Char.go:139 |
| BuffsRemoved | UserId, MobInstanceId | as BuffsAdded | gmcp.Char.go:158-171 | NewTurn_PruneBuffs.go:53 (UserId only) | gmcp.Char.go:158 |
| RoomAction | SourceUserId, SourceMobId | dead type | none | none | none |
| Input | UserId, MobInstanceId | body (spec) | world.go:89 branches on MobInstanceId; :124-178 track UserId for input blocking and the per-turn tracker | userrecord.go:288 (`Command`), :319 (`CommandFlagged`); Buff_ApplyBuffs.go:67; PlayerSpawn_HandleJoin.go:76; mobs.go:323 (`Mob.Command`); admin.build.go:47,151; admin.command.go:48,56; world.go:477 (InputWorker, `WorldInput.FromId` = UserId from the connection), :539 (macros) | world.go:78 |
| SkillUsed | UserId | dead type | none | none | none |
| Message | UserId | account (direct text to a connection) | Message_SendMessages.go:23-35 sends to `user.ConnectionId()` | userrecord.go:340 (`SendText`) | hooks/Message_SendMessages.go:17 |
| Message | ExcludeUserIds | body today (compared with room player ids) | Message_SendMessages.go:45-60 loops `room.GetPlayers()` and skips excluded ids | rooms.go:261,273,304 (`SendTextCommunication`, `SendText`, `SendTextToExits`); usercommands/print.go:16 | same |
| Communication | SourceUserId, SourceMobInstanceId | body (find the source's room, label "player"/"mob"); SourceUserId also keys `parties.Get` (account) | gmcp.Comm.go:54-93 | mobcommands/say.go:27; sayto.go:40,52,72,114; usercommands/broadcast.go:28, party.go:640, say.go:38, whisper.go:47 | gmcp.Comm.go:42 |
| Communication | TargetUserId | account/connection (the whisper recipient) | gmcp.Comm.go:99-103 | sayto.go:40,114; whisper.go:47 | gmcp.Comm.go:42 |
| RoomChange | UserId, MobInstanceId | body | Every consumer asks "is the mover a player": CleanupEphemeralRooms.go:17; gmcp.Room.go:167-209 (Players vs Npcs, self-exclusion, Room.Info to the mover); gmcp.Party.go:85-89; follow.go:257 (`followId` key) | roommanager.go:384 (`MoveToRoom`); rooms.go:746 (`AddMob`) | hooks/RoomChange_CleanupEphemeralRooms.go:14; follow.go:255; gmcp.Party.go:79; gmcp.Room.go:159 |
| MobIdle | MobInstanceId | body; slice 6 deletes the event | MobIdle_HandleIdleMobs.go:17; follow.go:247 | NewRound_IdleMobs.go:125 | MobIdle_HandleIdleMobs.go:15; follow.go:245 (`events.First`) |
| EquipmentChange | UserId, MobInstanceId | body, GMCP-only consumer (UserId only) | gmcp.Char.go:268-298 | mobcommands/equip.go:93, remove.go:26,52; usercommands/drop.go:62, equip.go:84, get.go:105,170,241,312, give.go:114,119,161, put.go:93, remove.go:25,81 | gmcp.Char.go:268 |
| ItemOwnership | UserId, MobInstanceId | body, consumers use UserId only | gmcp.Char.go:236-245 (queues an update even when UserId is 0; dropped at :366); gmcp.Room.go:58-84 (players only, finds the room via UserId) | DoCombat.go:398,410,852,864,978,989; mobcommands/drop.go:85, get.go:92, give.go:71,77,118,124, put.go:82; usercommands/drop.go:105, get.go:120,138,208,269,361, give.go:87,93,186,192, go.go:121, lock.go:62,114, put.go:108, unlock.go:62,114; cleanup.go:94,135 | gmcp.Char.go:236; gmcp.Room.go:58 |
| PlayerSpawn | UserId, ConnectionId | account/session (entering the world) | HandleJoin: GetByUserId, RemoveLinkDeadUser, login Input (PlayerSpawn_HandleJoin.go:25-85); gmcp.go:186-199 uses ConnectionId | world.go:285 | PlayerSpawn_HandleJoin.go:19; gmcp.Char.go:340; gmcp.Game.go:30; gmcp.go:186; gmcp.Mudlet.go:211 |
| PlayerDespawn | UserId | account. Its consumers touch body state: charm expiry `IsCharmed(evt.UserId)` (HandleLeave.go:49-55), `room.RemovePlayer(evt.UserId)` (:57), follow `followId{userId}` (follow.go:341), and gmcp.Room.go:136-138 compares room player ids with it | see left | world.go:210,225,425 | PlayerDespawn_HandleLeave.go:24; follow.go:336; gmcp.Game.go:29; gmcp.Room.go:117 |
| PlayerChanged | UserId | account | gmcp.Game.go:109 ignores the fields | start.go:180; world.go:446 | gmcp.Game.go:31 |
| GainExperience | UserId | account (brief) | gmcp.Char.go:216-231 | userrecord.go:196 | gmcp.Char.go:216 |
| LevelUp | UserId | account (brief) | gmcp.Char.go:320 | userrecord.go:211 | gmcp.Char.go:320; gmcp.Game.go:32 |
| PlayerDrop | UserId | body state (health < 1), players only | PlayerDrop_HandlePlayerDrop.go:24-39 texts the user and excludes them from room text | DoCombat.go:1049 | PlayerDrop_HandlePlayerDrop.go:18 |
| PlayerDeath | UserId | account (KD); follow uses it as a body key (follow.go:363) | gmcp.Char.go:80-81 | usercommands/suicide.go:96 | follow.go:358; gmcp.Char.go:80 |
| PlayerDeath | KilledByUsers | account (the keys of PlayerDamage, kill credit) | gmcp.Char.go:79-93 | suicide.go:102 | gmcp.Char.go:78 |
| PlayerDeath | KillerMobId | template MobId; keep | | suicide.go:103 | |
| MobDeath | InstanceId | body | follow.go:352 should use it but uses `evt.MobId` (bug, E8a) | mobcommands/suicide.go:86 | follow.go:347; gmcp.Char.go:78 |
| MobDeath | PlayerDamage, KilledByUsers | account (XP and kill credit) | gmcp.Char.go:79 | suicide.go:92-93 | |
| CharacterChanged | UserId | account; no producer left (alt module stripped) | gmcp.Char.go:177 | none | gmcp.Char.go:177; gmcp.Game.go:33 |
| UserSettingChanged | UserId | account | ClearSettingCaches.go:20-26 | usercommands/set.go:103,126,150,174,211,248,284,321,337,401 | hooks/UserSettingChanged_ClearSettingCaches.go:13 |
| CharacterVitalsChanged | UserId | GMCP-only (player vitals to a connection) | gmcp.Char.go:196-211 | NewRound_AutoHeal.go:74; DoCombat.go:421,424,569,876; NewRound_UserRoundTick.go:122; NewTurn_PruneBuffs.go:51; usercommands/suicide.go:272 | gmcp.Char.go:196 |
| CharacterTrained | UserId | GMCP-only; no producer | gmcp.Char.go:304 | none | gmcp.Char.go:304 |
| CharacterStatsChanged | UserId | account (`Character.userId`, private, players only) | gmcp.Char.go:251 | character.go:1486 | gmcp.Char.go:251 |
| PartyUpdated | UserIds | account (parties are keyed by UserId) | follow.go:236-238 stops `followId{userId}`; gmcp.Party.go:131-180 | party.go:52,126,174,190,382,397,444,460,504,547,598; gmcp.go:275 | follow.go:234; gmcp.Party.go:131 |
| Party | LeaderUserId, UserIds, InviteUserIds, AutoAttackers | dead event type (the live struct is `parties.Party`, keyed by UserId) | none | none | none |
| RedrawPrompt | UserId | account/connection | RedrawPrompt_SendRedraw.go:20-39 | Broadcast_SendToAll.go:33; Message_SendMessages.go:33,82; NewRound_AutoHeal.go:72 | RedrawPrompt_SendRedraw.go:14 |
| AggroChanged | UserId, MobInstanceId | body | gmcp.Party.go:61-72 (UserId leads to the party); gmcp.Room.go:100-103 (UserId>0 picks Players, else Npcs) | DoCombat.go:143,230,290,297,306,481,574,576,579,620,736,743,751,881,883,886,899,909,1002,1004,1007; NewRound_IdleMobs.go:71; mobcommands/suicide.go:123; usercommands/attack.go:190,236, flee.go:17 | gmcp.Party.go:55; gmcp.Room.go:90 |
| CLIRequest | UserId | dead type | none | none | none |
| AlignmentChanged | UserId | account (GMCP Char.Info) | gmcp.Char.go:120-133 | NewRound_UserRoundTick.go:95; mobcommands/suicide.go:180,276 | gmcp.Char.go:120 |
| Purchase | SellerUserId, SellerMobId (template) | dead type | none | none | none |
| MobItemDrop | MobId (template) | no consumer | none | mobcommands/suicide.go:336,367 | none |
| System | Data (any) = UserId for logoff, kick, leaveworld | account | world.go:205-233 casts `Data.(int)` and passes it to GetByUserId | NewTurn_PruneBuffs.go:57; NewRound_InactivePlayers.go:51; NewTurn_CleanupLinkDead.go:38; inputhandlers/systemcommands.go:136 | world.go:184 |

Module-local events, all connection-addressed (keep UserId): `GMCPOut{UserId}` (15 producers, dispatch gmcp.go:492-504 via `accounts.GetConnectionId`); `GMCPCharUpdate`, `GMCPRoomUpdate`, `GMCPGameRequest`, `GMCPWorldUpdate`, `GMCPHelpRequest` (`UserId`); `PartyUpdateVitals{LeaderId}` (a party leader UserId). `gmcp.go:538,560,587` special-cases `UserId == 1` for console debug output.

## Section C: room internals

**Fields** (rooms.go:93-97), all unexported and runtime only:
- `players []int` holds user IDs. `NewRoom`/`NewEmptyRoom` initialise it (:113,129).
- `mobs []int` holds mob instance IDs.
- `visitors map[VisitorType]map[int]uint64` maps a visitor type (`VisitorUser = "user"`, `VisitorMob = "mob"`, :57-58) to id → expiry turn.
- `lastVisited uint64` is a round number.

`RoomManager.roomsWithUsers` and `roomsWithMobs` (roommanager.go:27-28, 48-49) count players and mobs per room. They drive `GetRoomsWithPlayers` (:445-474, which also prunes empty rooms) and `GetRoomsWithMobs` (:476).

**Writers:**
- `AddPlayer` (rooms.go:2017-2028), `RemovePlayer` (:2031-2040). Only caller pairs: `MoveToRoom` (roommanager.go:319,381) and `HandleLeave` (PlayerDespawn_HandleLeave.go:57).
- `AddMob` (rooms.go:737-759) fires `RoomChange{MobInstanceId}`, sets the mob's RoomId and Zone, and updates `roomsWithMobs`. `RemoveMob` (:761-776).
- `Prepare` spawning appends to `r.mobs` directly, with no RoomChange (rooms.go:622).
- `GetDetails` deletes stale ids from `r.mobs` while ranging over it (roomdetails.go:289; bug, E8d).
- The room-replace copy duplicates both lists (save_and_load.go:227-232).
- `MarkVisited(id, vType, subtract...)` (rooms.go:1045-1067). Callers: `MoveToRoom` (roommanager.go:318,324, VisitorUser), `AddMob` and `RemoveMob` (rooms.go:744,763, VisitorMob). `PruneVisitors` refreshes the entries of current occupants and drops expired ones (rooms.go:1811-1851); `RoomMaintenance` calls it every 3 s (roommanager.go:180).

**Readers of `players`** (each needs "actors with an account" once the lists merge):
- `PlayerCt` (rooms.go:1074-1077), used for "skip if no players" in mob say, emote, sayto and look (mobcommands), in `lookRoom` Prepare (usercommands/look.go:406), in wander toward players, and in `Prepare` of adjacent rooms (rooms.go:703).
- `GetPlayers(flags...)` (:1181-1264).
- `ArePlayersAttacking` (:1270), `IsCalm` (:1266).
- `findPlayerByName` (:1405-1447, the `@<userId>` syntax at :1411-1424).
- `isInRoom` (:1866), `RoundTick` boredom reset (:2003).
- `ApplyBuffIdToPlayers` (:417).
- `RepeatSpawnItem` (:2105).
- `GetDetails` (roomdetails.go:248).
- `MoveToRoom` Prepare (roommanager.go:312), `GetRoomsWithPlayers` (:457), `removeRoomFromMemory` refuses to unload while players are present (:497).
- `TryEphemeralCleanup` (ephemeral.go:227), `DeleteRoomForAdmin` / `DeleteZoneForAdmin` (admin.go:246,646).

**Readers of `mobs`:**
- `MobCt` (:1069), `GetMobs(flags...)` (:1079-1179).
- `AreMobsAttacking` (:1286).
- `findMobByName` (:1449-1509, the `#<instanceId>` syntax at :1455-1468; otherwise names are numbered `name#1, name#2` by position, charmed mobs last).
- `isInRoom` (:1856), `RoundTick` (:2004), `ApplyBuffIdToMobs` and `ApplyBuffIdToNativeMobs` (:441,465).
- `GetVisibility` light check (:194).
- `GetDetails` (roomdetails.go:270).
- `removeRoomFromMemory` destroys every mob instance in the room (roommanager.go:501-503).

**Filter semantics to keep:** `GetMobs` and `GetPlayers` filter by Aggro target kind. `FindFightingPlayer` means `Aggro.UserId != 0`; `FindFightingMob` means `Aggro.MobInstanceId != 0` (rooms.go:1114-1123, 1216-1225). With a single `Aggro` ActorId target, these flags need the target's kind (does it have an `Account()`).

`FindAll = 0b111111111` does not include `FindNative` (bit 10), so `GetMobs()` with no flag returns early with every mob (rooms.go:55,1096). This is unchanged behaviour, noted for anyone rewriting the filter.

**Readers of `visitors`:**
- `Visitors(vType)` (rooms.go:1301), `HasVisited` (:1316, no live callers outside rooms), `HasRecentVisitors` (:1345, loot-goblin room choice, roommanager.go:422,430).
- `findMobExit` (:1879-1921) and `findUserExit` (:1923-1965). Both are called only from `GetDetails` tracking (roomdetails.go:386,429). `findUserExit` reads `Visitors(VisitorMob)` and looks those ids up as user ids (bug, E8b).
- `GetDetails` tracking also reads `Visitors(VisitorMob)` and `Visitors(VisitorUser)` directly (roomdetails.go:372,415).

**Persistence: none of these is saved.**
- `SaveRoomInstance` skips unexported fields (`field.PkgPath != ""`, save_and_load.go:264) and `yaml:"-"` fields. `SaveRoomTemplate` marshals the struct, which skips unexported fields.
- `SpawnInfo.InstanceId` is `yaml:"-"` (spawninfo.go:5).
- `Room.Corpses` is `yaml:"-"`. `Corpse.UserId` is an account ref and `Corpse.MobId` a template id (corpse.go:10-11).
- Copyover writes only `{UserId, ConnectionId, IsLinkDead, LinkDeadTurn}` per user (accounts/copyover.go:16-21). Rooms are rebuilt and mobs respawn in the new process; `triggerCopyover` saves rooms and users first (copyover.go:20-21).
- Saved values that are account refs (keep): `Sign.VisibleUserId` (rooms.go:1354-1393), `Item.StashedBy` (look.go:571; save_and_load.go:212).

## Section D: id allocation and every id that leaves the process

**Allocation:**
1. `mobs.addInstance` (mobs.go:42-49): `instanceCounter++` (package var, :23, starts at 0, so the first id is 1). It stores the mob in `mobInstances` and then fires `OnInstanceCreated`. It is the only add path; `NewMobById` (:156-233) is its only caller. `DestroyInstance` (:260-264) deletes and fires `OnInstanceDestroyed(id)`. `combat.newSimMob` (simulate.go:41-92) builds unregistered mobs with `InstanceId == 0`.
2. `actors.Spawn` (registry.go:21-29) and `bindAccount` (:113-117): `lastId++`, one counter for players and NPCs already. `mobSpawned` (wire.go:42-45) calls `Spawn` from the `OnInstanceCreated` hook, so a mob's ActorId is a different number from its InstanceId today. `actors.Init` subscribes before copyover restore and before any room spawns mobs (wire.go:13-24).
3. `accounts.GetUniqueUserId` (users.go:662-698) allocates persistent user ids (index or file scan, highest + 1).
4. Ids compared or used as map keys:
   - By mob instance id: `mobs.mobInstances` (mobs.go:26), `recentlyDied` (:30,105-123; read by world.go:603), `actors.byMobInstanceId` (registry.go:17), `world.mobInputEventTracker` (world.go:49,94,107), DoCombat `mobHealthBefore` and `mobsHandled` (:195,1056), `follow.followId.mobInstanceId` (follow.go:75-78), `rooms.visitors[VisitorMob]`, `SpawnInfo.InstanceId` (respawn bookkeeping, rooms.go:541-627,719-734,1988-1998; roommanager.go:505-515).
   - By user id, for body state: `rooms.visitors[VisitorUser]`, `world.userInputEventTracker` and `ignoreInput` (world.go:41,48), `follow.followId.userId`, `mobs.mobsHatePlayers[group][userId]` (mobs.go:27,634-674), `Mob.playersAttacked` (mobs.go:83,280-293; no callers).
   - By user id, for credit: `Character.PlayerDamage[userId]` (character.go:89,157-171).
   - No instance id is used as a slice index (grep of `[...InstanceId]`-style indexing found only map keys).

**Ids that leave the process (client-visible):**
1. **GMCP `Char.Enemies[].id`** is `"#<InstanceId>"` (gmcp.Char.go:452 via `Mob.ShorthandId`, mobs.go:266-268). `engaged` compares `mob.InstanceId` with the viewer's `Aggro.MobInstanceId` (:436-457).
2. **GMCP `Room.Info.Contents.Npcs[].id`** is `"#<InstanceId>"` (gmcp.Room.go:376). **`Players[].id`** is `"@<UserId>"` (gmcp.Room.go:348 via `Account.ShorthandId`, userrecord.go:161-163).
3. **The web client echoes those ids back as commands.** `window-room.js:497-521` builds `look <id>`, `attack <id>` and `list <id>` menu commands from the Npcs and Players ids. These come back as input and are parsed by `Room.FindByName` (`@N` → user id, `#N` → instance id, rooms.go:1405-1469). The id value and prefix are a round-trip contract with the shipped web client.
4. **Typed and generated commands.** Players and admins can type `#N` or `@N` (`attack`, `look`, `give`, `show`, `consider`, admin `command`, `grant`, `mute`, `follow`, `shoot`). Party auto-attack generates `attack #%d` (instance id) and `attack @%d` (user id) (attack.go:179,227).
5. **Web admin JSON.**
   - `GET /admin/api/v1/users/{id}` (api_v1_users.go:38-53) returns `*accounts.Account`. `GET /admin/api/v1/characters/{name}` (api_v1_characters.go:47-79) returns `*characters.Character`.
   - `Character` has no json tags on `Aggro` (`{Type, MobInstanceId, UserId, SpellInfo, ExitName, RoundsWaiting}`), `Charmed` (`{UserId,...}`), `CharmedMobs`, `KillerMobInstanceId`, `PlayerDamage`, so these leave by Go field name for online users.
   - `PATCH /admin/api/v1/users/{id}` decodes JSON over a copy of the live Character, and `UpdateOnlineUser` copies the whole Character into the live one (users.go:589-611). A client can therefore write these runtime fields.
   - `/admin/api/v1/mobs` (api_v1_mobs.go:16-131) serialises `mobs.Mob` templates, including a JSON `"InstanceId"` key (always 0 for templates) and the Character fields above. PATCH decodes it back.
   - The admin HTML under `_datafiles/html` references none of these field names. The only hit is the docs page below.
6. **Docs page.** `_datafiles/html/admin/docs-modules.html:1124-1228` documents `MobInstanceId` on MobIdle, RoomChange, AggroChanged, Buff*, Input and `InstanceId` on MobDeath. It needs editing in slice 2.
7. **Logs.** Debug logs carry `actorId`, `userId` and `mobInstanceId` (wire.go:30-49); `world.go:604` logs a `mobId` that is really an instance id. No tooling parses them: a grep of `cmd/smoke` for these keys found nothing, and the pre-4b transcript (`cmd/smoke/testdata/pre-4b/bot0.txt`) contains no `#N`/`@N` ids.
8. **Saved files.** No instance id or actor id is written to disk (Section C). Saved account refs stay: `KD.PlayerKills`/`PlayerDeaths` keyed `"userid:name"` (kdstats.go:14-15,52-70), `Sign.VisibleUserId`, `Item.StashedBy`, `users/<UserId>.yaml`.

## Section E: what makes "mob instance id == ActorId" unsafe or surprising

1. **Two counters today.** `mobs.instanceCounter` and `actors.lastId` are independent. For the numbers to match, the instance id must come from the shared leaf counter inside `addInstance`, before the `mobInstances` insert and the hook, and `actors.Spawn` must accept that id instead of allocating. `mobs` and `accounts` cannot import `actors` (wire.go:15-16), so the counter has to sit in a leaf package both import.
2. **Players' visible numbers change.** `@N` is a UserId today. With one ID space a player's id is an ActorId, which:
   - differs per login (actor.go:8-9),
   - is interleaved with mob ids,
   - resets on copyover. Today UserIds survive copyover; mob ids already reset.
   Web-client windows holding pre-copyover `@N` or `#N` ids go stale until the next Room.Info.
3. **The kind lives in the prefix and the return shape.**
   - `FindByName` returns a `(playerId, mobInstanceId)` pair (rooms.go:1396-1403). 18 non-definition callers branch on which half is non-zero.
   - `#`/`@` select the kind, and `@N` deliberately fails mob lookup (:1452-1454) and `#N` player lookup (:1408-1410).
   - One ID space needs one lookup plus an "is this a player" test (`Account() != nil`).
4. **Body references that survive a logout.**
   - Today a mob's `Aggro.UserId` keeps pointing at the account. `IdleMobs` drops aggro only when `GetByUserId` fails or the player left the room (NewRound_IdleMobs.go:66-72).
   - With an ActorId, a logout and re-login makes a new actor, so the old reference dangles.
   - Link-dead rebinding keeps the ActorId (registry.go:102-111), so reconnects are safe.
   - The same applies to `visitors[VisitorUser]` tracking (roomdetails.go:415-417 looks visitors up as accounts), `follow` keys, and `mobsHatePlayers`.
5. **"Is a player here" is a list length today.** See the `players` readers in C: `PlayerCt`, `len(r.players)` in Prepare, unload, ephemeral cleanup, admin delete, and the `roomsWithUsers` index that drives `UserRoundTick` and `PruneBuffs`. One `actors` list needs an account-aware count, or a kept per-kind index.
6. **Exclude lists are user ids.** `Message.ExcludeUserIds` and every `room.SendText(txt, userIds...)` (the spec counts 145 `room.SendText(` sites in 43 files) are matched against `GetPlayers()` (Message_SendMessages.go:45-60). `SendTextToExits(..., room.GetPlayers(rooms.FindAll)...)` passes a player list as the exclude list (usercommands/go.go:205, mobcommands/go.go:128). If `GetPlayers` starts returning ActorIds, every exclude argument has to change in the same slice, or the consumer has to translate.
7. **Name rendering compares the viewer with the aggro target.** `GetMobName`/`GetPlayerName(viewingUserId)` add the "aggro" suffix when `Aggro.UserId == viewingUserId` (character.go:589-595,684). Callers pass `user.UserId` (roomdetails.go:262,281) or 0 (combat.go:169,173; look.go:111; mobcommands/look.go:120).
8. **Pre-existing bugs the conversion will touch.** Converting can silently change outcomes; the spec says outcomes stay identical.
   - (a) `follow.onMobDeath` uses the template `evt.MobId` as an instance id (follow.go:352).
   - (b) `findUserExit` iterates `Visitors(VisitorMob)` and resolves the ids with `accounts.GetByUserId` (rooms.go:1940-1946). With one ID space it would start resolving actor ids, changing tracking output.
   - (c) DoCombat.go:97-99: the inner `userId` shadows the outer one, so the flee-blocked-by-player check compares `u.Character.Aggro.UserId` with `u`'s own id.
   - (d) `GetDetails` mutates `r.mobs` while ranging over it (roomdetails.go:270-290).
   - (e) DoCombat.go:513 appends `user.Character.Aggro.UserId` (0 in a mob fight) to `affectedPlayerIds`.
   - (f) DoCombat.go:50-54 dereferences `user` before its nil check.
   - (g) `handleMobCombat` sets `mob.Character.Aggro = nil` on a spell fizzle without an `AggroChanged` event (:646), unlike other paths.
9. **Dead code that can be deleted rather than converted** (callers checked by grep):
   - `Character.SetCast` has no callers (character.go:1100). So `SpellAggroInfo.TargetUserIds`/`TargetMobInstanceIds` (aggro.go:14-19) are never populated, and the SpellCast branches never run: DoCombat.go:171-245 and 633-680, `IsAggro` :1126-1142.
   - `Mob.PlayerAttacked`/`HasAttackedPlayer` have no callers (mobs.go:280-293).
   - Dead event types: `RoomAction`, `SkillUsed`, `CLIRequest`, `Purchase`, `Party`. No producer: `CharacterTrained`, `CharacterChanged`. No consumer: `MobItemDrop`.
10. **Zero means none everywhere.** That fits `ActorId` (actor.go:9) and ids starting at 1. Two cases to know:
    - `combat.SimulateCombat` sets aggro with both ids 0 (simulate.go:114-115). It only reads `Aggro.Type`, but `calculateCombat` dereferences `sourceChar.Aggro` unconditionally (combat.go:253).
    - `mapper.Config.UserId < 0` means debug (mapper.go:682,693). That is an account field and stays.
11. **Charm mixes body and account in one field.** `CharmInfo.UserId` is used:
    - as a body: clearing aggro on its target (character.go:481), `IsCharmed(user.UserId)` for follow and assist (DoCombat.go:150; go.go:236; teleport.go:134; attack.go:165; shoot.go:79; HandleLeave.go:51);
    - as an account: damage credit `TrackPlayerDamage(charmedUserId)` (combat.go:87-90), `SetTempData` on the account (mobcommands/suicide.go:72-76), and `TrackCharmed` on the charmer's Character (MobRoundTick.go:49-52; suicide.go:41-45).
    `CharmedMobs` lives on the player's Character (`yaml:"-"`) and is pruned against `MobInstanceExists` at login and reconnect (users.go:226-230,270-274,305-309).
12. **Killer reference crosses a queued command.** `KillerMobInstanceId` is set in `handleAffected` (DoCombat.go:1038-1044) and read in `usercommands/suicide.go:84-94` through `mobs.GetInstance`. `suicide` is queued (`user.Command`), so the mob may already be gone; `killerMobId` is then 0. Identical under ActorId.
13. **Mob death lives in a deleted package.** Mob death XP, alignment, tame mastery, KD and loot live in `mobcommands/suicide.go`, which slice 4 deletes. Slice 2 still edits it: 1 MII, 10 IID.
14. **Input bookkeeping splits by kind.** world.go:89-181 handles mobs first, with a blocking tracker even for `ReadyTurn<=0`. Players skip the tracker when `ReadyTurn<=0` and have input blocking keyed by UserId (connection state). Merging the trackers is safe only if the per-kind ordering is kept until slice 4 merges the paths.

## Section F: open questions I could not settle by reading

1. **`#N`/`@N` and GMCP ids.** Should both become `ActorId`, and with one prefix or two? Or does `@<UserId>` stay for players, so the web client and habits keep working? Changing them in lockstep (GMCP id and `FindByName`) keeps the window-room.js round trip intact; changing only one side breaks it.
2. **Room message exclude lists.** Does slice 2 convert `ExcludeUserIds` and every `SendText(..., userIds...)` call to ActorIds? Or does `Message_SendMessages` translate, leaving the conversion to slice 7's messaging work?
3. **GMCP-only events.** `BuffsAdded`, `BuffsRemoved`, `BuffsTriggered`, `CharacterVitalsChanged`, `EquipmentChange`, `ItemOwnership` describe a body, but their only consumers need a connection. Convert them to `ActorId` with consumers calling `Account()`, or leave them `UserId`?
4. **Parties.** Party membership (`parties.Party` UserIds, AutoAttackers, leader), `Communication` party routing and `PartyUpdated` stay UserId under "account-only stays UserId"? Teleport moves charmed mobs with the party leader (admin.teleport.go:132-138), and `attack` auto-assist issues commands to members by UserId.
5. **`mobsHatePlayers[group][userId]`** (mobs.go:651-674): does hostility attach to the account (survives re-login) or the body?
6. **Visitor tracking.** One map keyed by ActorId, plus a kind test, or keep the `VisitorUser`/`VisitorMob` split? Tracking a player by name after they re-log needs the account.
7. **Web admin JSON field names** (`Aggro.MobInstanceId`, `CharmedMobs`, `KillerMobInstanceId`, `Mob.InstanceId`): may they change? Nothing in `_datafiles/html` reads them except the docs page, but external API users might.
8. **Pre-existing bugs (E8).** Preserve or fix them in slice 2? (a) and (b) change behaviour as soon as ids unify.
9. **Leaf package API.** The brief says "a leaf package defines ActorId and its counter". Does `actors.Spawn` then take an id, or get a separate `SpawnWithId`? And does `mobs.addInstance` call the leaf allocator directly?
10. **Files that later slices delete.** Edit them mechanically in slice 2 (needed for the Done-when line), or move the deletions earlier? Slice 4 needs slice 2, so moving them earlier is not possible as the slices are ordered; this question is about effort only.
