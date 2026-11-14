# Phase 5 slice 4.2: progression cleanup (spec)

Date: 2026-11-14. Written at `strip/phase-5-4-2-progression` `59cfa1f8` (branched from `master`
`703f4c30`, slice 4.1 merged as #37). Builds on `gomud_strip_plan.md` decisions 6, 12 and 13-15,
the phase design `2026-11-06-phase-5-rules-seam-design.md` (slice 4 row, "Notes for later slices")
and slice 4.1's spec (`2026-11-12-phase-5-4-1-leftovers-spec.md`).
**Not yet reviewed by Calabe.**

**Goal.** A small cleanup slice. The reference ruleset keeps levels, XP and stats (decision 13), so
4.2 deletes only alignment, the training points nobody can spend, and dead progression code; makes
players spend stat points the way NPCs do; and books a player's death by the killing blow. Moving
level, XP and stats behind the ruleset seam stays in slice 5.

## Rulings this spec implements

Earlier, still binding: decision 6 (the engine knows health only; slice 5 does the move), decision 12
(code the shrink deletes goes straight to its bare form, with a golden recorded first), the one-actor
ruling, rulesets return messages and never send.

Calabe, 2026-11-14 (brainstorm):

1. **The reference ruleset keeps progression and stats** (strip plan decision 13, revising decision
   9). Levels, XP, kill XP, the stat block and level-scaled creatures stay. 4.2 and 4.3 delete only
   what is dead.
2. **Alignment is deleted** (decision 14): kill shifts, decay, the displays, race default alignment,
   the alignment help and the follow module's hate check.
3. **Players spend stat points the way NPCs do; training points are deleted** (decision 15).
4. **4.2 is cleanup only.** Moving level, XP and stats behind the seam is slice 5's, with gold.
5. **The killing blow books a player's death.** It is a PvP death only when a player's attack took
   the victim down; that player gets the kill and is named in the death notice. Players who only
   hurt the victim earlier get nothing.

## Facts verified against source (2026-11-14, `59cfa1f8`)

Read or grepped by the controller today.

| Fact | Value | Where |
|---|---|---|
| Character fields | `TrainingPoints` (`trainingpoints`), `StatPoints` (`statpoints`) and `Alignment int8` (`alignment`) are saved fields | `characters/character.go:57`, `:58`, `:62` |
| Level-up | `LevelUp` adds a level, then training points and stat points by the progression config, then `Validate` | `character.go:947-963` |
| Who spends stat points | only `AutoTrain` (round-robin over the six stats' `Training`, `:1128-1149`), called for NPCs at spawn (`mobs/mobs.go:165`) and by the simulator (`dikubasic/simulate.go:66`); nothing spends a player's | grep `StatPoints` |
| Who spends training points | nothing: read only by the `{tp}` prompt token (`accounts/userrecord.prompt.go:259-260`), GMCP `Char.Info` (`modules/gmcp/gmcp.Char.go:434`, field `:768`) and the level-up event (`accounts/userrecord.go:191-227`) | grep `TrainingPoints` |
| No `train` or `stats` command | no file in `internal/usercommands` matches `train`, `score` or `stat`; no command string `train` in Go | `ls`, grep |
| Player stat growth | a player's stats also rise per level through race gains (`stats/stats.go:35-51`, via `RecalculateStats`) | inventory, read |
| Prompt tokens | `{a}`, `{A}` (alignment name, coloured by it), `{tp}`, `{sp}` | `userrecord.prompt.go:248-263` |
| `online` | an Alignment column, coloured by the alignment name | `usercommands/online.go:20`, `:66`, `:74`; `accounts/onlineinfo.go:7`, filled at `userrecord.go:587` |
| GMCP | `Char.Info` sends `alignment` (`gmcp.Char.go:430`, `:764`), `skillpoints` (= `StatPoints`, `:433`) and `trainingpoints` (`:434`, `:768`); `Game` online list sends `alignment` (`gmcp.Game.go:64`, `:77`); listeners for `AlignmentChanged` (`gmcp.Char.go:54`, `:120`) and `CharacterTrained` (`:46`, `:315`) | read |
| `CharacterTrained` | declared (`events/eventtypes.go:276-280`), listened to (GMCP), never sent | grep `CharacterTrained{` |
| `AlignmentChanged` | sent by decay (`hooks/NewRound_ActorUpkeep.go:84`) and by kill shifts (`dikubasic/death.go:189`); declared `eventtypes.go:326-332` | grep |
| Alignment decay | accounts only (`NewRound_ActorUpkeep.go:41`, `:70-88`), every `AlignmentDecayRounds` (`configs/config.gameplay.go:23`, default `:121-122`) | read |
| Kill shift | `killAwards` gives each killer `shiftAlignment` (`dikubasic/death.go:135`, `:157`, `:177-201`), from `AlignmentChange` (`dikubasic/calculations.go:293`) | read |
| Alignment elsewhere | `Validate` clamps it (`character.go:1205-1210`); `UpdateAlignment`, `DecayAlignment`, `AlignmentName` (`:1293-1347`); bands in `characters/alignment.go`; NPCs take the race default at spawn (`mobs.go:182-186`) and in the simulator (`simulate.go:77-80`); players at `start` (`usercommands/start.go:99`); `races.Race.DefaultAlignment` (`races/races.go:34`, clamp `:78-81`); `mobs.HatesAlignment` (`mobs.go:305-325`), read only by `follow` (`modules/follow/follow.go:538`): a creature whose alignment is far from yours "won't let you follow them" | grep `Alignment` over Go |
| Alignment world data | `2-guard.yaml` (`alignment: 30`), three races (`defaultalignment:`); one user file | grep `_datafiles/world/empty` |
| Alignment colours | `ansi-aliases.yaml` names each band (`holy` `:47` to `unholy` `:55`) | grep |
| YAML loading | no `UnmarshalStrict` or `KnownFields` in non-test Go: a saved `alignment:` or `trainingpoints:` key loads and is ignored once the field goes | grep |
| Trainer rooms | `Room.SkillTraining` (`rooms/rooms.go:83`, type `:103`) shows "You can train here! Type `train`..." (`rooms/roomdetails.go:109-110`), and GMCP `Room`/`World` send a `trainer` detail (`gmcp.Room.go:513-514`, `gmcp.World.go:175-176`); admin: `rooms/admin.go:156`, `web/api_v1_rooms.go:244`, `admin/rooms.html:747`, `:1894`; no world file sets `skilltraining:` | grep |
| `MercHirePricePerLevel` | config only (`config.gameplay.go:14`, default `:89-90`; `config.yaml:220`; admin `config.html:231`, `config-wizard.html:1124`); no reader | grep |
| Death booking | `recordPlayerDeath`: `len(PlayerDamage) > 0` books a PvP death, else a mob death (`death/death.go:136-140`); every user in `PlayerDamage` gets `AddPlayerKill` and the victim `AddPlayerDeath` (`:144-155`); the world notice names those users (`:161-168`); `PlayerDeath.KilledByUsers` lists them (`:187`) | read |
| `PlayerDamage` | every player swing at the body, misses included, kept until 30 rounds pass with none (`character.go:139-151`; `NewRound_DoCombat.go:83`) | read |
| Killing blow | combat records `KilledBy` (the attacker's actor ID) and the killer NPC's details when a player is taken out of Up (`NewRound_DoCombat.go:209-219`); a non-combat death leaves it zero | read |
| `PlayerDeath` listeners | `follow` (drops the victim's followers, reads `UserId` only, `follow.go:368-377`); GMCP refreshes kill counts for `KilledByUsers` (`gmcp.Char.go:57`, `:80`) | grep |

## Design

### Alignment goes

Deleted: the `Alignment` field and its clamp; `UpdateAlignment`, `DecayAlignment`, `AlignmentName`,
`characters/alignment.go`; decay in actor upkeep and `AlignmentDecayRounds`; the kill shift
(`shiftAlignment`, `AlignmentChange` and their tests); the `AlignmentChanged` event and its GMCP
listener; the prompt tokens `{a}` and `{A}`; the `online` Alignment column (in-game and
`online.html`) and `OnlineInfo.Alignment`; GMCP `alignment` in `Char.Info` and `Game`; race
`DefaultAlignment` and its use at spawn, in the simulator and at `start`; `mobs.HatesAlignment` and
the follow check that reads it; the alignment help page; the alignment colour aliases (if nothing else
names them; the plan checks); the admin user, mob and race page fields, the web client's character
window field; config docs for the removed key.

Player-visible: no alignment anywhere; kills no longer print "Your alignment has shifted"; a creature
never refuses to be followed. Old saves, mob, race and user files that hold `alignment:` or
`defaultalignment:` load as before (the key is ignored).

### Players spend stat points; training points go

`LevelUp` still adds stat points by the progression config, then spends them at once the way
`AutoTrain` does for NPCs (round-robin over the six stats). One routine serves both. A player loaded
with unspent stat points (every player today who has levelled) has them spent once at load.

Training points are deleted: the field, the level-up grant, the `TrainingPointsPerLevel` and
`TrainingPointsEveryNLevels` config keys (and their admin progression page fields), the `{tp}` and
`{sp}` prompt tokens, GMCP `trainingpoints` and `skillpoints`, and the level-up event's point
counts. The `StatPointsPerLevel` keys stay: they still set how many points a level gives.

Player-visible: a level-up raises stats a little more than today (by the points it used to bank).
`{tp}` and `{sp}` in a saved prompt print as literal text (the plan checks what an unknown token does
and records it).

### Dead progression code goes

- `CharacterTrained` and its GMCP listener (never sent).
- `MercHirePricePerLevel` (no reader).
- Trainer rooms: `Room.SkillTraining`, `TrainingRange`, the "You can train here!" alert, the GMCP
  `trainer` detail, and their admin fields. No world file sets them and no command trains.
- Help text that sends players to the missing `stats` and `train` commands.

### The killing blow books a player's death

When a player dies, the attacker whose attack took it out of Up (combat's `KilledBy`) decides:

- a player: a PvP death; that player alone gets `AddPlayerKill` and the victim `AddPlayerDeath` for
  them; the world notice names that player (to readers who perceive them, as today);
  `PlayerDeath.KilledByUsers` lists that player;
- an NPC: a mob death, no player credit, and a notice that names no killer;
- no killing blow (suicide, a timed effect): a mob death, no player credit, no killer named (today
  such a death books PvP and credits every player in `PlayerDamage`).

Combat records the killer's user ID with the blow, beside the NPC details it already records, so the
books do not need the killer still online. Kill XP for an NPC's death is unchanged: it is still split
among every player who damaged it.

## Testing and evidence

**Commit 1 records goldens before anything changes** (decision 12): a combat fixture where Bob hits
Alice, the fight ends, and a rat later kills Alice within 30 rounds. On today's code the death is
booked PvP, Bob gets the kill and the notice names Bob. A reads fixture that levels a player up and
prints its prompt with `{tp}` and `{sp}` and its stats.

**Goldens that change, each diff read and summarised in its commit and the PR** as "decided" or
"unexpected" (an unexpected change stops the work):

- alignment: the reads goldens that print it (`H-outside` alignment shift, `G-character-info`
  `online`), and any combat golden that prints an alignment line;
- training: the commit 1 reads fixture (stats rise, `{tp}`/`{sp}` lines change);
- death booking: the commit 1 combat fixture (mob death, no Bob credit, no killer named);
  `death_pvp` unchanged (Alice's blow killed Bob);
- smoke transcript fixtures under `cmd/smoke/testdata` that show the `online` table;
- every other golden byte-identical.

**Unit tests:** `LevelUp` spends its stat points (stats rise by the points, none left); a loaded
player with unspent points has them spent once; the death booking for each killing-blow case (player,
NPC, none) and that an earlier attacker gets nothing; an old save with `alignment:` and
`trainingpoints:` loads. Every new test shown to fail under a prescribed mutation.

**Smoke:** unchanged steps; passes with 1, 5 and 12 bots.

## Commits, in order

1. Goldens on today's code (test files and testdata only).
2. Killing-blow booking. The commit 1 combat fixture changes.
3. Players auto-train; training points gone. The commit 1 reads fixture changes.
4. Alignment gone. Alignment goldens change.
5. Dead progression code gone. Goldens unchanged.
6. Docs: patch notes, status, the phase design's slice 4 row and notes (decision 13 rewrites 4.3's
   list too), help.

## What this locks in

- No alignment in the reference ruleset (decision 14).
- Players never choose where stat points go until a ruleset adds `train` (decision 15).
- No assist credit for player kills: one killer per death (ruling 5).
- Level, XP and stats stay engine-named until slice 5 moves them.

## Out of scope

- Moving level, XP, stats and gold behind the seam (slice 5).
- 4.3's subjects. Decision 13 changes 4.3: stats stay, so 4.3's brainstorm reopens its list (mana,
  `statmods`, race fields, skills, the hit roll) and decision 10 (no movement cost or carry limit),
  which rested on "no stats".
- The kill XP split and its level scaler (unchanged).
- The `PlayerDamage` 30-round window itself (still used for kill XP).
- The web theme CSS tokens for alignment and trainer badges (slice 7's web sweep).

## Decisions taken without asking

| Decision | Why | If wrong |
|---|---|---|
| A player's unspent stat points are spent once at load | They are owed (ruling 3 spends them on level-up); otherwise today's players keep a bank they can never use | Leave existing banks unspent |
| One routine spends stat points for players and NPCs | One-actor ruling; `AutoTrain` already does it | none |
| `{tp}` and `{sp}` go rather than printing 0 | Ruling 3 names them; `{sp}` would always read 0 | Keep `{sp}` printing 0 |
| GMCP `skillpoints` goes with `trainingpoints` | It sends `StatPoints`, always 0 after this slice | Keep it at 0 for client compatibility |
| Trainer rooms are deleted here, not in 4.3 | Training is this slice's subject; no command trains and no world file sets them | Leave them for 4.3 with skills |
| Combat records the killer's user ID with the blow | The books then hold whether or not the killer is still online at the death, as the NPC details already do | Look the killer up at death and treat a missing one as a mob death |
| A death with no killing blow books as a mob death with no player credit | Ruling 5 credits the blow; there is none | Credit `PlayerDamage` as today in that case only |
| Kill XP keeps its split among every damaging player | Ruling 5 is about player deaths; NPC kill XP is progression that stays (ruling 1) | Give the XP to the killing blow too |
| The alignment colour aliases go only if nothing else names them | They were for alignment names; another user would break | Keep them |
| The web theme CSS tokens stay for slice 7 | Inert styling across 20 theme files; slice 7 sweeps the web client | Delete them here |
| Slice name 4.2 and branch `strip/phase-5-4-2-progression` | Continues 4.1's naming | Rename |
