# Phase 5 rulings for Calabe (2026-11-06)

From the three inventories in `2026-11-06-phase-5-inventory.md` (M: rules data, N: combat and death,
O: buffs and surfaces). The inventories raised 25 player-visible questions. Most of them answer
themselves once ruling 1 is made, so four rulings are asked here; the rest are listed under
"Defaults I'll take" with the reason, and any of them can be overruled.

Spot-checked against source today: `Heal` adds the health amount to mana
(`internal/characters/character.go:1339`); `attack` calls `CanPvp` (`attack.go:239`) and `shoot`
does not; an unknown race becomes race 1 (`character.go:1522`); a hostile NPC only prints "notices
you as you enter!" (`usercommands/go.go:322`); `revive-on-death` is read only in
`suicide_npc.go:31`; no `aid` command exists; no `help/spell` template exists; the default prompt is
`HP:{hp}/{HP} MP:{mp}/{MP}` (`config.textformats.go:18`); GMCP `Char.Stats` has fixed fields such as
`strength` (`modules/gmcp/gmcp.Char.go:880`).

## Answers (Calabe, 2026-11-06)

Recorded as decisions 5-8 in `gomud_strip_plan.md`.

1. Neither option: strip the rules to a bare minimum so an Earthdawn ruleset can be layered on later,
   while leaving a working barebones MUD for forks. With the 2026-10-24 decision 1 ("move into
   diku-basic, then shrink"), this means the move keeps parity and a shrink step follows. The
   parity defaults below apply to the move only.
2. Health only.
3. Die at 0, no downed state. The engine keeps a ruleset "is dead" call so Earthdawn can add its
   unconsciousness rating later.
4. Delete, restore the quit line.

5. Gold only (decision 9).

---

## Ruling 1. What Phase 5 does to the game

**Player-facing effect:** whether fights, death and progression play the same after Phase 5.

- **A. Parity (recommended).** Phase 5 moves today's rules behind the seam unchanged: same numbers,
  same random draws, all 9 combat goldens byte-identical. Clear bugs are fixed in their own commits
  (golden diff shown per fix). Dead systems are deleted. New game design (NPCs that start fights,
  initiative instead of players-first turn order, NPCs that level, party XP scaling) waits for the
  real ruleset after the strip.
  - Pro: the goldens prove the move is mechanical; a stub ruleset test proves the seam. Smallest risk
    in the riskiest phase.
  - Con: diku-basic keeps GoMud's quirks (NPCs never attack, players act first) until someone
    redesigns them.
  - Locks in: diku-basic is a test fixture and a faithful copy of GoMud's rules, not the game.
- **B. Redesign while moving.** Settle NPC aggression, initiative, NPC levelling and party XP now.
  - Pro: one pass through the code.
  - Con: every golden changes for design reasons, so the goldens stop proving the move; each design
    question needs its own spec.

The questions this settles under A: NPC aggression (N1) stays as today; turn order (N7) stays
players-first; party XP (N8) unchanged; XP and levelling stay accounts-only as a diku-basic rule (M5,
N4; the engine path is one, the ruleset decides who gains XP); unknown race still falls back to race
1 (M7); pet kills still name the pet as killer (N5).

## Ruling 2. What the engine knows, and what only the ruleset knows

**Player-facing effect:** none if done right. The prompt, `who`, party list and GMCP keep their
current text. This decides what a future ruleset (the geo game) can rename or drop.

Today 14 packages name health and mana; level gates PvP (`rooms.go:2343`) and shows in `who`, party,
GMCP and the `{lvl}` prompt token; gold moves through the kept get/drop/give/put commands.

- **A. Engine knows health only (recommended).** The core needs "alive, downed, dead" and a health
  bar, so health and max health stay engine fields. Level, XP, mana, gold and the six stats become
  ruleset values in a keyed traits container. Prompt tokens, `who`/party columns and GMCP vitals and
  stats are filled by the ruleset under today's names, so Mudlet and the web client see no change.
  PvP's minimum-level check moves into the ruleset.
  - Pro: a ruleset without mana or levels is possible with no core edit; the guard test covers mana
    and level as well as the stats.
  - Con: more call sites move (gold through four kept commands; mana through regen and prompt).
  - Locks in: the engine's only built-in pool is health.
- **B. Engine knows the Diku core: health, mana, level, gold.** Only the six stats and the formulas
  move.
  - Pro: smaller Phase 5.
  - Con: the geo ruleset inherits mana, levels and gold as fixed engine fields, or edits the core to
    drop them.
- **C. Engine knows nothing, health included.** Purest seam, but every "is this actor alive" check
  becomes a ruleset call; not worth it.

GMCP `Char.Stats` (O7) under A: the ruleset supplies a map, and diku-basic fills it with today's six
field names, so clients see the same JSON.

## Ruling 3. Death

**Player-facing effect:** what a player sees when they die, and how long they lie bleeding.

Today: at 0 health a player is downed and bleeds 1 health a round (`NewRound_AutoHeal.go:44-59`)
while the room is told "Somebody needs to provide aid!"; at -10
they die, revive at 1 health in a private copy of room 75 and walk out through a portal (the Phase 4
stand-in). No `aid` command exists, so every downed player dies. The Death Recovery buff (24) used to
heal 1d10 health and mana a round in room 75; its JS is gone, so it does nothing. NPCs die at 0 and
skip all of this. `revive-on-death` works only for NPCs.

- **A. Make the stand-in the diku-basic death rule (recommended).** Keep downed and bleed-out, change
  the bleed line so it no longer promises aid, rebuild the Death Recovery regen in Go, and route
  players and NPCs through one engine death routine that calls the ruleset (so `revive-on-death`
  works for anyone who has it).
  - Pro: closest to parity; one death path, as the one-actor ruling asks.
  - Con: a downed player still just waits to die until someone builds `aid`.
- **B. Same, but die at 0 (no downed state)** until `aid` exists.
  - Pro: no pointless wait. Con: drops a Diku staple and changes the reads goldens.
- **C. Same as A, and also build `aid`.** New feature inside the strip; not recommended under
  ruling 1A.

## Ruling 4. Dead systems: delete or carry

**Player-facing effect:** none today (each of these already does nothing or errors).

- Spells: no `cast` command; `help spell` shows a template error. Spell data in `world/empty/spells`.
- Charm: nothing can start one; follow, party, flee, save and expiry code still runs for it.
- Skills dead code on `Character`: spellbook, form change, cast chance, barter, tame mastery, the
  `map` skill. (`peep` and `enchant` skill checks are live in kept commands and stay.)
- `professions` world folder; `set shopstyle` (saved, never read).
- Buff flags no shipped buff uses: poison, drunk, tripping, accuracy, blink, see-nouns.
- References to buffs that don't ship: the `wildfire` mutator (buff 22) and a potion (buff 5).
- Quit and illumination lost their messages in Phase 4; quit is silent.

Options:
- **A. Delete all of it; restore the quit message as plain Go text (recommended).** Charm,
  companions and spells come back as Go behaviour when the game wants them. The quit message is one
  line a player notices; the illumination glow goes with the dead flags.
- **B. Delete all of it, quit stays silent.**
- **C. Carry spells and charm into diku-basic** as Go, dormant. More code under the seam with no
  player path to it.

## Ruling 5. What the bare ruleset keeps after the shrink

**Player-facing effect:** what a fresh fork of the base plays like.

The floor is fixed by rulings 2-4: health, a hit roll and weapon damage, health regen, death at 0
and a respawn, and races as names only. Anything kept above that floor is code and world data an
Earthdawn ruleset replaces later, and that a forker gets as a starting point.

Candidates to keep:
- **Levels and XP.** A forker gets progression out of the box. Earthdawn has no levels in this form
  (circles and legend points), so it replaces them.
- **A stat block** (GoMud's six stats, with items and races modifying them). Earthdawn has its own
  six attributes and step dice, so it replaces them.
- **Races with stat differences.** Needs the stat block.
- **Gold.** The kept get, drop, give and put commands already handle coins; nearly any MUD has a
  currency.

Recommendation: keep **gold only**. The base plays as "walk, talk, pick things up, fight, die,
respawn". Everything else is Earthdawn's to define, and the guard test then forbids the core from
naming any ruleset value (gold, and anything a ruleset registers), not just six stat names.

---

## Defaults I'll take (overrule any)

Bug fixes, each its own commit with its golden diff:
- `Heal` adds health to mana and ignores the mana amount: fix (mana regen numbers change).
- `shoot` skips `CanPvp` and sends no aggro event: fix to match `attack`.
- `see-hidden` is honoured by `look` but not by messaging: make messaging honour it. Your 4b ruling
  ("a hidden actor reads Someone everywhere") stands for viewers without `see-hidden`.
- Level-up event under-reports training and stat points; missing nil checks in `attack` and
  `consider`; two checks compare the raw room 75 ID and miss the private copy; a deleted buff spec
  can panic `CancelBuffsWithFlag` on death.

Engineering:
- Layout: `internal/rules` (interface) and `internal/rules/dikubasic` (moved packages).
- Dead code goes in a first slice, before the move.
- Buff expiry joins `ActorUpkeep` for every actor (per round, not per turn); mutators get one buff
  list for every actor; buff YAML names a Go effect with an `effect:` key. Rename `buffs` to
  `affects` afterwards, mechanically.
- Save files keep the `stats` key and lowercase names, so they load unchanged.
- Guard test built on the Go syntax tree, not a word grep (`internal/perception` shares the word
  Perception).
- Before moving death, add a combat golden that runs the queued `suicide`.
- Loot goblin config, `ChanceToTame`, `PricePerLife`, `GetRoomWithMostItems`, `RememberRoom` and
  `LevelUp.StatsDelta`: delete (no readers; greps in the inventory).
