# Phase 4b slice 2c: item IDs (implementation plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to
> implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** every item has one `ids.ID` (`0x<cell>:0x<random>`), made once when the item is created or
instantiated and kept for good: saved with it, restored with it, and never shared by two items.

**Architecture:** `items.Item.UUID` (upstream's `internal/uuid`, never saved) becomes
`Item.Id ids.ID`, saved as `id`. Items get an ID where they are made (`items.New`, `Validate` for an
item that has none, and `items.FromTemplate` when an NPC is made from its template). Template files
(room and mob templates) never hold item IDs. Clients address an item by its ID token, as they do
NPCs and players since slice 2; the `!N` and `!N:<uuid>` forms go. `internal/uuid` is deleted.

**Tech stack:** Go, testify, `gopkg.in/yaml.v2`. No new dependency.

Parent spec: `2026-10-27-phase-4b-npcs-are-players-design.md`, decisions 10 and 11, section 4
slice 2c. Branch: `strip/phase-4b-2c-item-ids` off `master`.

---

## Facts verified against source (2026-10-28, master `82b5f53b`)

| Fact | Value | Where |
|---|---|---|
| `internal/uuid` users | One non-test importer, `internal/items/items.go`, plus `items_test.go`. No other package and no `go.mod` entry | `grep 'internal/uuid"'` |
| `uuid.` call sites outside the package | 6, all in `items.go`: `UUIDItem` const (`:33`), field (`:39`), `New` (`:57`), `Validate` (`:155-156`), `FindMatchIn` (`:580-593`) | `grep 'uuid\.'` |
| **Item IDs are never saved today** | `UUID uuid.UUID \`yaml:"-"\` // \`yaml:"uuid,omitempty"\`` (`items.go:39`), so every load gives every item a new UUID through `Validate`. Tagged so since upstream's UUID PR `162842e0` (2025-04-21). No file under `_datafiles` contains `uuid` | `items.go:39`, `git log -S`, `grep -ril uuid _datafiles` (empty) |
| Uses of item identity | `Item.Equals` (`ItemId` and `UUID`, `items.go:383-385`) in 11 places: `accounts/storage.go:42`, `characters/character.go:763,801,817,2019`, `mobcommands/equip.go:64`, `rooms/container.go:26`, `rooms/corpse.go:45`, `rooms/rooms.go:911,918` | `grep '\.Equals('` |
| `ShorthandId` | Only `Item.ShorthandId` is left (`items.go:478-484`, `!<ItemId>:<uuid>`). Callers: GMCP `Room.Info.Contents.Items[].id` (`gmcp.Room.go:319`) and `Char.Inventory` items (`gmcp.Char.go:857`) | `grep ShorthandId` |
| Client round trip | `window-room.js:534` sends `get <itm.id>` with the GMCP id as an opaque string; no other web client window reads an item id | `grep` over `static/js/windows` |
| `!` item forms | `FindMatchIn` reads `!<ItemId>` and `!<ItemId>:<uuid>` (`items.go:576-607`). The only Go producers of `!%d` are the room spawn checks `rooms.go:640` (floor) and `:663` (container). No help file or admin page documents `!N` | `grep '!%d'`, `grep` over `_datafiles` |
| Where items are made | `items.New` (15 non-test calls in 7 files); `Item.Validate` gives an ID to any item without one (`items.go:149-166`), called on load for characters (`character.go:1561-1565`), rooms (`rooms.go:2253-2264`), mob templates (`Mob.Validate` calls `Character.Validate`, `mobs.go:575`) and mob instances (`mobs.go:208-210`) | `grep` |
| **NPCs from one template share item IDs today** | Mob templates are validated at load, so their items already have UUIDs. `NewMobById` copies the template with `mob := *m` (`mobs.go:158`): the `Items` slice shares the template's backing array and `Equipment` (`Worn`, a struct of `Item` values) copies the template's UUIDs. Every NPC of one template holds the same item IDs. `newSimMob` (`combat/simulate.go:42`, via `GetMobSpec`'s `*m` copy, `mobs.go:233-239`) does the same | read |
| Template writers | `SaveRoomTemplate` (`save_and_load.go:147`) marshals the room it is given; callers include live rooms (`roommanager.go:864` `fromRoom`, `mapper.go:1397`), so live floor items reach template files. `SaveMobSpec` (`mobs/admin.go:21`) writes the spec. `Mob.Save` (`mobs.go:594`) has no callers | `grep SaveRoomTemplate`, `grep '\.Save()'` |
| Room template copies | Ephemeral rooms load the template file once per copy (`ephemeral.go:148`), so a template file holding item IDs would give every copy the same IDs | read |
| Room instance saves | `LoadRoomInstance` loads the template, then unmarshals the instance file over it (`save_and_load.go:102-124`); `saveRoomInstanceData` writes the fields that differ from a fresh template load (`:267-331`). Copyover runs `rooms.SaveAllRooms()` then `accounts.SaveAllUsers()` (`copyover.go:20-21`) | read |
| NPC store | `mobs.Snapshot` saves the whole `Mob` with its `Character`; `Restore` keeps saved items and only runs `Validate`, which keeps a non-zero ID (`mobs/saved.go:25-83`) | read |
| **Admin "Save User" scrambles items** | The admin page posts items as `{ItemId}` only and equipment as `{ItemId, CanNeverBeRemoved}` (`users.html:828-831, 1148-1154`). `apiV1PatchUser` decodes over a copy of the character whose `Items` slice shares the live record's backing array (`api_v1_users.go:80-99`). `encoding/json` decodes into existing slice elements, so posted item *k* keeps every other field (UUID, uses, adjectives) of the old item *k*, and the live record's items change before validation. Checked with a scratch program (go1.25.7): old `[{1 A 5} {2 B 7}]`, posted `[{2} {3} {9}]` gives `[{2 A 5} {3 B 7} {9 "" 0}]`, live record changed to `[{2 A 5} {3 B 7}]` | read, scratch run |
| ID guard | `TestIdsAreMadeOnlyAtCreation` allows `ids.New` in four files (`guards_test.go:162-187`); `TestNoNumericShorthandIds` forbids `#%d`/`@%d` literals (`:193-210`). Both skip test files | read |
| ID text form | `ids.ID` marshals through `MarshalText`/`UnmarshalText` for YAML and JSON; the zero ID is `""` and `omitempty` drops it | `internal/ids/ids.go` |
| Smoke | Bots get, drop and get a `copper ring`, then check they still hold it after relogin, death and (Linux) copyover (`cmd/smoke/script.go:172-228`). The client counts GMCP package names only, not payloads (`client.go:87-90`) | read |
| Test baseline | `go build ./...` passes; `go test ./... -count=1`: 35 ok, the 3 known Windows failures (`TestPromptForPasswordMismatch`, `TestLoadPluginMobs_MergesValidMob`, `TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride`) | run 2026-10-28 |

## Decisions this plan applies

From the spec (Calabe, 2026-10-27):

- Decision 10: IDs are `0x<16 hex>:0x<16 hex>`; things created where there is no geography use cell
  `0x0`. People never type IDs: they name things by name, and GMCP carries IDs as opaque tokens.
- Decision 11: an ID is assigned once, at creation or instantiation, and kept for good. An item is one
  of the listed instances. Only creating or instantiating something new makes a new ID.

Defaults set by the controller from those rulings (recorded so they are not re-asked; Calabe reviews
them with this plan):

- **"Saved items load with their identity kept"** (the slice's Done-when) is read as: from this
  slice on, an item's ID is saved and every load keeps it. Items saved before this slice have no
  saved identity to keep (fact above); each gets its ID at its first load, as bodies saved before
  slice 2 did (`giveBodyId`), and keeps it after its next save.
- **A template item is not a thing.** Room and mob template files never hold item IDs. Making an NPC
  from its template gives each of its items a new ID (`items.FromTemplate`), the same as a new
  instantiation of the NPC gets a new body ID. Loading a room from its template file gives its items
  IDs through `Validate`, and the room instance save keeps them from then on.
- **The `!N` forms go.** `!<ItemId>` and `!<ItemId>:<uuid>` are IDs people could type (decision 10);
  clients send the ID token instead. The two spawn checks that used `!%d` look up by `ItemId`
  directly. A guard forbids `!%d` literals as it does `#%d`/`@%d`.
- **Admin "Save User" keeps the identity of the items it keeps.** Without this, saving IDs would make
  the positional merge above hand item *k*'s ID to whatever item the admin put at position *k*. The
  handler decodes onto empty item lists, then matches each posted item to an existing item of the same
  `ItemId` (same slot first for equipment) and keeps that item whole; an unmatched one is new. This
  also fixes the upstream merge that moved uses, adjectives and enchantments between items, and stops
  a failed decode from changing the live record. No admin page change.
- `Mob.Save` (no callers) is deleted rather than taught to strip IDs.
- No smoke change: the smoke client does not parse GMCP payloads, and the bots' rings change hands, so
  an ID check there would be fragile. Persistence is proved by Go tests (Task 5); smoke must still
  pass.

## File structure

| File | Change |
|---|---|
| `internal/items/items.go` | `Id ids.ID` replaces `UUID`; `New`, `Validate`, `Equals`, `FindMatchIn` use it; `ShorthandId` and `UUIDItem` go; new `FromTemplate`, `WithoutId`, `WithoutIds` |
| `internal/items/items_test.go` | uuid tests become ID tests |
| `internal/uuid/` | deleted |
| `modules/gmcp/gmcp.Room.go`, `gmcp.Char.go` | item `id` is `itm.Id.String()` |
| `internal/rooms/rooms.go` | spawn checks look up by `ItemId`; new `floorHasItemId` |
| `internal/characters/itemids.go` (new) | `InstantiateItems`, `StripItemIds`, `KeepItemIdentity` |
| `internal/characters/itemids_test.go` (new) | tests for the three |
| `internal/mobs/mobs.go` | `NewMobById` instantiates the template's items; `Mob.Save` deleted |
| `internal/mobs/admin.go` | `SaveMobSpec` writes a copy without item IDs |
| `internal/mobs/saved.go` | comment no longer names `Mob.Save` |
| `internal/combat/simulate.go` | `newSimMob` instantiates the template's items |
| `internal/rooms/save_and_load.go` | `SaveRoomTemplate` writes `templateData()`; new `templateData` |
| `internal/web/api_v1_users.go` | decode split into `decodeUserPatch`, with replace semantics and `KeepItemIdentity` |
| `internal/web/api_v1_users_test.go` (new) | tests for `decodeUserPatch` |
| `internal/guards/guards_test.go` | `items.go` allowed to make IDs; `!%d` forbidden; `FromTemplate` and `InstantiateItems` callers fenced |
| tests in `internal/accounts`, `internal/rooms`, `internal/mobs` | persistence tests (Task 5) |
| `PATCH_NOTES.md`, `status.md`, spec, `geomud-verifying` | Task 6 |

---

### Task 0: Branch

- [ ] **Step 1:** From a clean `master` equal to `origin/master`:

```bash
git switch -c strip/phase-4b-2c-item-ids
```

---

### Task 1: `Item.Id` replaces the UUID; the `!` forms go

**Files:**
- Modify: `internal/items/items.go:3-12, 31-34, 39, 52-67, 149-166, 383-385, 478-484, 574-608`
- Modify: `internal/items/items_test.go`
- Modify: `modules/gmcp/gmcp.Room.go:319`, `modules/gmcp/gmcp.Char.go:857`
- Delete: `internal/uuid/`
- Modify: `internal/guards/guards_test.go:163-168`

- [ ] **Step 1: Rewrite the identity tests in `items_test.go`.** Replace the import of
  `internal/uuid` with `internal/ids`, change `makeItem` and replace the three `TestFindMatchIn_Shorthand*`
  tests and `TestItem_Equals`:

```go
// makeItem builds a test Item with an inline Spec so no file loading is needed.
func makeItem(id int, name string) Item {
	return Item{
		ItemId: id,
		Id:     ids.New(ids.NoCell),
		Spec:   &ItemSpec{ItemId: id, Name: name, NameSimple: name},
	}
}
```

```go
func TestFindMatchIn_ByIdToken(t *testing.T) {
	sword := makeItem(10, "sword")
	shield := makeItem(20, "shield")

	_, full := FindMatchIn(sword.Id.String(), shield, sword)
	assert.Equal(t, sword.Id, full.Id)
}

func TestFindMatchIn_IdTokenOfAnItemNotInTheList(t *testing.T) {
	sword := makeItem(10, "sword")
	other := makeItem(10, "sword")

	partial, full := FindMatchIn(sword.Id.String(), other)
	assert.Equal(t, Item{}, partial)
	assert.Equal(t, Item{}, full)
}

// The !N forms are gone (decision 10: people never type IDs). "!10" is not a
// name, so it matches nothing.
func TestFindMatchIn_BangFormsMatchNothing(t *testing.T) {
	sword := makeItem(10, "sword")

	for _, s := range []string{"!10", "!10:" + sword.Id.String()} {
		partial, full := FindMatchIn(s, sword)
		assert.Equal(t, Item{}, partial, s)
		assert.Equal(t, Item{}, full, s)
	}
}

func TestItem_Equals(t *testing.T) {
	a := makeItem(1, "sword")
	b := makeItem(1, "sword")

	assert.False(t, a.Equals(b), "different IDs should not be equal")
	assert.True(t, a.Equals(a), "same item should equal itself")
}

func TestValidateGivesAnIdOnlyWhenMissing(t *testing.T) {
	itm := makeItem(1, "sword")
	itm.Id = ids.ID{}
	itm.Validate()
	first := itm.Id
	assert.False(t, first.IsZero())
	itm.Validate()
	assert.Equal(t, first, itm.Id, "an existing ID is never replaced")

	empty := Item{ItemId: 0}
	empty.Validate()
	assert.True(t, empty.Id.IsZero(), "an empty slot is not an item")
}

func TestIdSurvivesYAML(t *testing.T) {
	itm := makeItem(1, "sword")
	b, err := yaml.Marshal(itm)
	require.NoError(t, err)
	assert.Contains(t, string(b), "id: "+itm.Id.String())
	var back Item
	require.NoError(t, yaml.Unmarshal(b, &back))
	assert.Equal(t, itm.Id, back.Id)
}

func TestFromTemplateMakesANewItem(t *testing.T) {
	tpl := makeItem(1, "sword")
	tpl.Adjectives = []string{"rusty"}
	tpl.Uses = 3

	a, b := FromTemplate(tpl), FromTemplate(tpl)
	assert.NotEqual(t, tpl.Id, a.Id)
	assert.NotEqual(t, a.Id, b.Id)
	assert.False(t, a.Id.IsZero())
	assert.Equal(t, 3, a.Uses)
	assert.Equal(t, []string{"rusty"}, a.Adjectives)

	a.Adjectives[0] = "shiny"
	a.Spec.Name = "changed"
	assert.Equal(t, "rusty", tpl.Adjectives[0], "the template's adjectives are not shared")
	assert.Equal(t, "sword", tpl.Spec.Name, "the template's overrides are not shared")

	assert.Equal(t, Item{}, FromTemplate(Item{}), "an empty slot stays empty")
	assert.Equal(t, ItemDisabledSlot, FromTemplate(ItemDisabledSlot))
}

func TestWithoutIds(t *testing.T) {
	list := []Item{makeItem(1, "sword"), makeItem(2, "shield")}
	got := WithoutIds(list)
	for _, itm := range got {
		assert.True(t, itm.Id.IsZero())
	}
	assert.False(t, list[0].Id.IsZero(), "the given list is not changed")
	assert.Nil(t, WithoutIds(nil))
}
```

  Add `"github.com/stretchr/testify/require"` and `"gopkg.in/yaml.v2"` to the test imports. Leave
  `TestFindMatchIn_EmptyList` and the name-matching tests as they are.

- [ ] **Step 2: Run them to see them fail**

Run: `go test ./internal/items/ -count=1`
Expected: build failure (`unknown field Id in struct literal`, `undefined: FromTemplate`).

- [ ] **Step 3: Change `items.go`.**

Imports: drop `internal/uuid`, add `"github.com/GoMudEngine/GoMud/internal/ids"`. Delete the
`const ( UUIDItem = ... )` block (`:31-34`) with its TODO comment.

Field (`:39`):

```go
	Id                ids.ID         `yaml:"id,omitempty"`            // This item, for good (Phase 4b decision 11)
```

`New` (`:52-67`): replace `newItm.UUID = uuid.New(UUIDItem)` with

```go
		newItm.Id = ids.New(ids.NoCell) // rooms carry no cell until the geo room key
```

`Validate` (`:154-157`):

```go
	// An item without an ID came from a template file or was saved before
	// item IDs existed: it becomes a thing now, and keeps this ID from here on.
	if i.Id.IsZero() {
		i.Id = ids.New(ids.NoCell)
	}
```

`Equals` (`:383-385`):

```go
func (i *Item) Equals(b Item) bool {
	return i.ItemId == b.ItemId && i.Id == b.Id
}
```

Delete `ShorthandId` (`:478-484`). Add after `Equals`:

```go
// FromTemplate makes a new item from an item in a template (an NPC's template
// holdings): the same state, its own ID, and no slices or overrides shared
// with the template. An empty or disabled slot is returned as it is.
func FromTemplate(tpl Item) Item {
	if tpl.ItemId < 1 {
		return tpl
	}
	itm := tpl
	itm.Id = ids.New(ids.NoCell) // rooms carry no cell until the geo room key
	itm.Adjectives = append([]string(nil), tpl.Adjectives...)
	if tpl.Spec != nil {
		spec := *tpl.Spec
		itm.Spec = &spec
	}
	itm.tempDataStore = nil
	return itm
}

// WithoutId is i as written to a template file: an item in a template is not
// a thing yet, so it has no ID.
func (i Item) WithoutId() Item {
	i.Id = ids.ID{}
	return i
}

// WithoutIds copies list with every item's ID cleared (see WithoutId).
func WithoutIds(list []Item) []Item {
	if list == nil {
		return nil
	}
	out := make([]Item, len(list))
	for idx, itm := range list {
		out[idx] = itm.WithoutId()
	}
	return out
}
```

`FindMatchIn` (`:574-608`): replace the whole `if len(itemName) > 1 { if itemName[0] == '!' { ... } }`
block with

```go
	// Clients name an item by its ID token (GMCP ids); people name it by name.
	if id, err := ids.Parse(itemName); err == nil {
		for _, itm := range items {
			if itm.Id == id {
				return itm, itm
			}
		}
		return Item{}, Item{}
	}
```

  and update the doc comment above `FindMatchIn` to add: "An ID token matches only the item with
  that ID." Drop the `strconv` import if nothing else in the file uses it (`go build` will say).

- [ ] **Step 4: GMCP.** In `modules/gmcp/gmcp.Room.go:319` and `modules/gmcp/gmcp.Char.go:857`
  replace `itm.ShorthandId()` with `itm.Id.String()`.

- [ ] **Step 5: Delete `internal/uuid`.**

```bash
git rm -r internal/uuid
```

- [ ] **Step 6: Allow `items.go` to make IDs.** In `TestIdsAreMadeOnlyAtCreation`'s `allowed` map
  (`guards_test.go:163-168`) add:

```go
		"internal/items/items.go":          "New: a new item; Validate: an item that has none (from a template file, or saved before item IDs existed); FromTemplate: an NPC's template item instantiated",
```

- [ ] **Step 7: Build and run**

```bash
go build ./...
go test ./internal/items/ ./internal/guards/ ./modules/gmcp/ -count=1
grep -rn --include=*.go -E 'uuid|UUID|ShorthandId' . | grep -v _test.go
```

Expected: build passes; tests pass; the grep prints nothing.

- [ ] **Step 8: Commit**

```bash
git add internal/items/items.go internal/items/items_test.go modules/gmcp/gmcp.Room.go modules/gmcp/gmcp.Char.go internal/guards/guards_test.go internal/uuid
git commit -F - <<'EOF'
items: Item.Id (ids.ID, saved) replaces the never-saved UUID

Items are named by their ID token (GMCP) or by name; the !N and
!N:<uuid> forms go. internal/uuid had no other user and is deleted.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 2: Room spawn checks look up by `ItemId`; `!%d` is forbidden

**Files:**
- Modify: `internal/rooms/rooms.go:640, 663` and add `floorHasItemId`
- Test: `internal/rooms/rooms_test.go`
- Modify: `internal/guards/guards_test.go:189-210`

- [ ] **Step 1: Write the failing test** in `internal/rooms/rooms_test.go`:

```go
func TestFloorHasItemId(t *testing.T) {
	r := NewEmptyRoom()
	assert.False(t, r.floorHasItemId(10))
	r.Items = append(r.Items, items.Item{ItemId: 10})
	assert.True(t, r.floorHasItemId(10))
	assert.False(t, r.floorHasItemId(11))
	r.Stash = append(r.Stash, items.Item{ItemId: 11})
	assert.False(t, r.floorHasItemId(11), "the stash is not the floor (as FindOnFloor(_, false))")
}
```

  (add the `items` import if the file lacks it).

- [ ] **Step 2:** Run `go test ./internal/rooms/ -run TestFloorHasItemId -count=1`.
  Expected: FAIL, `r.floorHasItemId undefined`.

- [ ] **Step 3: Implement.** Add near `FindOnFloor` in `rooms.go`:

```go
// floorHasItemId reports whether an item made from itemId lies on the floor
// (not in the stash).
func (r *Room) floorHasItemId(itemId int) bool {
	for _, itm := range r.Items {
		if itm.ItemId == itemId {
			return true
		}
	}
	return false
}
```

  At `rooms.go:640` replace
  `if _, alreadyExists := r.FindOnFloor(fmt.Sprintf(\`!%d\`, spawnInfo.ItemId), false); !alreadyExists {`
  with `if !r.floorHasItemId(spawnInfo.ItemId) {`.
  At `rooms.go:663` replace
  `if _, alreadyExists := container.FindItem(fmt.Sprintf(\`!%d\`, spawnInfo.ItemId)); !alreadyExists {`
  with `if _, alreadyExists := container.FindItemById(spawnInfo.ItemId); !alreadyExists {`
  (`container.go:49`).

- [ ] **Step 4: Forbid `!%d`.** In `TestNoNumericShorthandIds` change the doc comment's first line to
  "forbids shorthand IDs (#N, @N and the item form !N): people" and the condition to

```go
				(strings.Contains(lit.Value, "#%d") || strings.Contains(lit.Value, "@%d") || strings.Contains(lit.Value, "!%d")) {
```

  and the error to `"%s: formats a #N, @N or !N id (%s)"`.

- [ ] **Step 5: Run**

```bash
go build ./...
go test ./internal/rooms/ ./internal/guards/ -count=1
```

Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add internal/rooms/rooms.go internal/rooms/rooms_test.go internal/guards/guards_test.go
git commit -F - <<'EOF'
rooms: item spawn checks look up by ItemId; guard forbids !%d

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 3: Templates hold no item IDs; an NPC made from a template gets new items

**Files:**
- Create: `internal/characters/itemids.go`, `internal/characters/itemids_test.go`
- Modify: `internal/mobs/mobs.go:158` (after `mob := *m`), delete `Mob.Save` (`:594-...`)
- Modify: `internal/mobs/saved.go:12-13` (comment)
- Modify: `internal/mobs/admin.go:35`
- Modify: `internal/combat/simulate.go:42-46`
- Modify: `internal/rooms/save_and_load.go:163`
- Test: `internal/mobs/lifecycle_test.go`, `internal/rooms/rooms_test.go`
- Modify: `internal/guards/guards_test.go`

- [ ] **Step 1: Write the failing character tests** in `internal/characters/itemids_test.go`:

```go
package characters

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
)

func testItem(itemId int) items.Item {
	return items.Item{ItemId: itemId, Id: ids.New(ids.NoCell), Spec: &items.ItemSpec{ItemId: itemId, Name: "thing"}}
}

func TestInstantiateItemsGivesEveryItemANewId(t *testing.T) {
	tpl := New()
	tpl.Items = []items.Item{testItem(1), testItem(2)}
	tpl.Equipment.Weapon = testItem(3)

	a, b := *tpl, *tpl // as NewMobById copies the template
	a.InstantiateItems()
	b.InstantiateItems()

	for i := range tpl.Items {
		assert.NotEqual(t, tpl.Items[i].Id, a.Items[i].Id)
		assert.NotEqual(t, a.Items[i].Id, b.Items[i].Id)
		assert.Equal(t, tpl.Items[i].ItemId, a.Items[i].ItemId)
	}
	assert.NotEqual(t, tpl.Equipment.Weapon.Id, a.Equipment.Weapon.Id)
	assert.NotEqual(t, a.Equipment.Weapon.Id, b.Equipment.Weapon.Id)
	assert.True(t, a.Equipment.Head.Id.IsZero(), "an empty slot stays empty")

	a.Items[0].Uses = 99
	assert.NotEqual(t, 99, tpl.Items[0].Uses, "the template's item list is not shared")
}

func TestStripItemIdsLeavesTheOriginal(t *testing.T) {
	live := New()
	live.Items = []items.Item{testItem(1)}
	live.Equipment.Body = testItem(2)

	file := *live
	file.StripItemIds()

	assert.True(t, file.Items[0].Id.IsZero())
	assert.True(t, file.Equipment.Body.Id.IsZero())
	assert.False(t, live.Items[0].Id.IsZero(), "the original keeps its item IDs")
	assert.False(t, live.Equipment.Body.Id.IsZero())
}
```

- [ ] **Step 2:** Run `go test ./internal/characters/ -run 'InstantiateItems|StripItemIds' -count=1`.
  Expected: FAIL, `InstantiateItems undefined`.

- [ ] **Step 3: Implement** `internal/characters/itemids.go`:

```go
package characters

import "github.com/GoMudEngine/GoMud/internal/items"

// InstantiateItems gives c, a body just copied from an NPC template, its own
// items: each held and worn item is made new from the template's
// (items.FromTemplate), so no two NPCs share an item or its ID (Phase 4b
// slice 2c). The item list is a new slice, so the template's is untouched.
func (c *Character) InstantiateItems() {
	if c.Items != nil {
		held := make([]items.Item, len(c.Items))
		for i, itm := range c.Items {
			held[i] = items.FromTemplate(itm)
		}
		c.Items = held
	}
	for _, slot := range AllSlots() {
		if worn := c.Equipment.Get(slot); worn != nil {
			*worn = items.FromTemplate(*worn)
		}
	}
}

// StripItemIds clears the IDs of c's held and worn items, for writing c to a
// template file: an item in a template is not a thing yet. Call it on a copy;
// the item list is replaced, so the original's is untouched.
func (c *Character) StripItemIds() {
	c.Items = items.WithoutIds(c.Items)
	for _, slot := range AllSlots() {
		if worn := c.Equipment.Get(slot); worn != nil {
			*worn = worn.WithoutId()
		}
	}
}
```

- [ ] **Step 4:** Run the Step 2 command again. Expected: PASS.

- [ ] **Step 5: NPCs.** Write the failing test in `internal/mobs/lifecycle_test.go`:

```go
func TestNPCsFromOneTemplateHaveTheirOwnItems(t *testing.T) {
	tpl := &Mob{MobId: 9003}
	tpl.Character = *characters.New()
	tpl.Character.Items = []items.Item{{ItemId: 1, Spec: &items.ItemSpec{ItemId: 1, Name: "tail"}}}
	tpl.Character.Equipment.Weapon = items.Item{ItemId: 2, Spec: &items.ItemSpec{ItemId: 2, Name: "claw"}}
	tpl.Validate() // as at template load: the template's items get IDs
	mobs[9003] = tpl
	t.Cleanup(func() { delete(mobs, 9003) })

	a := NewMobById(9003, 1)
	b := NewMobById(9003, 1)
	t.Cleanup(func() { DestroyInstance(a.Character.Id); DestroyInstance(b.Character.Id) })

	assert.NotEqual(t, a.Character.Items[0].Id, b.Character.Items[0].Id)
	assert.NotEqual(t, tpl.Character.Items[0].Id, a.Character.Items[0].Id)
	assert.NotEqual(t, a.Character.Equipment.Weapon.Id, b.Character.Equipment.Weapon.Id)
	assert.False(t, a.Character.Items[0].Id.IsZero())
}
```

  (imports `characters`, `items`). Run `go test ./internal/mobs/ -run TestNPCsFromOneTemplate -count=1`.
  Expected: FAIL on the first `NotEqual` (the two NPCs share the template's item ID). If
  `NewMobById` needs configs or races the test cannot reach, report it and stop rather than changing
  the test's intent.

- [ ] **Step 6: Implement.** In `mobs.go` `NewMobById`, directly after `mob := *m // Make a copy of the mob`:

```go
		mob.Character.InstantiateItems() // a new NPC's items are new items (slice 2c)
```

  In `combat/simulate.go` `newSimMob`, after the `if mob == nil` block:

```go
	mob.Character.InstantiateItems() // the copy must not share the template's items
```

  Delete `func (r *Mob) Save() error` in `mobs.go` and any import only it used. In `saved.go:12-13`
  change "so template saves (Mob.Save) never gain runtime fields" to "so template saves (SaveMobSpec)
  never gain runtime fields".

- [ ] **Step 7:** Run `go test ./internal/mobs/ ./internal/combat/ -count=1`. Expected: PASS apart from
  the known `TestLoadPluginMobs_MergesValidMob` on Windows.

- [ ] **Step 8: Template writers.** In `mobs/admin.go` `SaveMobSpec`, replace the `SaveFlatFile` call's
  `spec` argument with a copy:

```go
	fileSpec := *spec
	fileSpec.Character.StripItemIds() // a template's items have no IDs (slice 2c)
	if err := fileloader.SaveFlatFile[*Mob](configs.GetFilePathsConfig().DataFiles.String()+`/mobs`, &fileSpec, saveModes...); err != nil {
```

  In `rooms/save_and_load.go` add:

```go
// templateData is r as written to its template file: an item in a template is
// not a thing yet, so it has no ID (Phase 4b slice 2c). The item lists and the
// container map are copies, so the live room keeps its item IDs.
func (r Room) templateData() Room {
	r.Items = items.WithoutIds(r.Items)
	r.Stash = items.WithoutIds(r.Stash)
	if r.Containers != nil {
		containers := make(map[string]Container, len(r.Containers))
		for name, c := range r.Containers {
			c.Items = items.WithoutIds(c.Items)
			containers[name] = c
		}
		r.Containers = containers
	}
	return r
}
```

  and in `SaveRoomTemplate` change `data, err := yaml.Marshal(&roomTpl)` to

```go
	fileData := roomTpl.templateData()
	data, err := yaml.Marshal(&fileData)
```

  If `go vet` reports a copied lock in `templateData`, take `*Room` and copy the exported item fields
  into a new `Room` value the same way `SaveRoomTemplate` already receives one by value; report which
  you did.

- [ ] **Step 9: Test the room template data** in `rooms_test.go`:

```go
func TestTemplateDataHasNoItemIds(t *testing.T) {
	r := NewEmptyRoom()
	r.Items = []items.Item{{ItemId: 1, Id: ids.New(ids.NoCell)}}
	r.Stash = []items.Item{{ItemId: 2, Id: ids.New(ids.NoCell)}}
	r.Containers = map[string]Container{"chest": {Items: []items.Item{{ItemId: 3, Id: ids.New(ids.NoCell)}}}}

	tpl := r.templateData()
	assert.True(t, tpl.Items[0].Id.IsZero())
	assert.True(t, tpl.Stash[0].Id.IsZero())
	assert.True(t, tpl.Containers["chest"].Items[0].Id.IsZero())

	assert.False(t, r.Items[0].Id.IsZero(), "the live room keeps its item IDs")
	assert.False(t, r.Stash[0].Id.IsZero())
	assert.False(t, r.Containers["chest"].Items[0].Id.IsZero())

	b, err := yaml.Marshal(&tpl)
	require.NoError(t, err)
	assert.NotContains(t, string(b), "id: 0x", "no item ID reaches the template file")
}
```

  (imports `ids`, `yaml "gopkg.in/yaml.v2"` as the package already uses it, `require`). If the room's
  own YAML has another `id:` key that starts `0x`, narrow the assertion to the item lists' YAML and say
  so.

- [ ] **Step 10: Fence the instantiation calls.** Add to `guards_test.go`:

```go
// TestItemsAreInstantiatedOnlyFromTemplates enforces decision 11 for template
// items: items.FromTemplate makes new IDs, so only the NPC-from-template path
// may call it, and only the places that make an NPC from its template may call
// Character.InstantiateItems.
func TestItemsAreInstantiatedOnlyFromTemplates(t *testing.T) {
	fromTemplate := map[string]string{
		"internal/characters/itemids.go": "InstantiateItems: an NPC's template items made new",
	}
	instantiate := map[string]string{
		"internal/mobs/mobs.go":        "NewMobById: a new NPC instantiation",
		"internal/combat/simulate.go":  "newSimMob: a simulated NPC instantiation",
	}
	for path, f := range goFiles(t) {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); ok && x.Name == "items" && sel.Sel.Name == "FromTemplate" {
				if _, ok := fromTemplate[path]; !ok {
					t.Errorf("%s: calls items.FromTemplate outside the NPC-from-template path", path)
				}
			}
			if sel.Sel.Name == "InstantiateItems" {
				if _, ok := instantiate[path]; !ok {
					t.Errorf("%s: calls InstantiateItems outside an NPC instantiation", path)
				}
			}
			return true
		})
	}
}
```

  (`goFiles` keys files by slash-separated path from the repo root and skips `_test.go`,
  `guards_test.go:22-37`.)

- [ ] **Step 11: Run**

```bash
go build ./...
go vet ./internal/rooms/ ./internal/mobs/ ./internal/characters/
go test ./internal/characters/ ./internal/mobs/ ./internal/rooms/ ./internal/combat/ ./internal/guards/ -count=1
```

Expected: pass, apart from the known `TestLoadPluginMobs_MergesValidMob`.

- [ ] **Step 12: Commit**

```bash
git add internal/characters/itemids.go internal/characters/itemids_test.go internal/mobs/mobs.go internal/mobs/saved.go internal/mobs/admin.go internal/mobs/lifecycle_test.go internal/combat/simulate.go internal/rooms/save_and_load.go internal/rooms/rooms_test.go internal/guards/guards_test.go
git commit -F - <<'EOF'
items: templates hold no item IDs; a new NPC's items are new items

NPCs made from one template shared its item IDs (a shallow copy of the
template). NewMobById and newSimMob now make each item new; room and mob
template writers drop item IDs. Mob.Save had no callers and is deleted.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 4: Admin "Save User" keeps the identity of the items it keeps

**Files:**
- Modify: `internal/characters/itemids.go`, `internal/characters/itemids_test.go`
- Modify: `internal/web/api_v1_users.go:80-99`
- Create: `internal/web/api_v1_users_test.go`

- [ ] **Step 1: Write the failing test** in `itemids_test.go`:

```go
func TestKeepItemIdentity(t *testing.T) {
	was := New()
	ring1, ring2, sword := testItem(1), testItem(1), testItem(2)
	ring1.Uses, ring2.Uses = 4, 7
	sword.Adjectives = []string{"rusty"}
	was.Items = []items.Item{ring1, ring2, sword}
	was.Equipment.Weapon = testItem(3)
	was.Equipment.Body = testItem(4)

	// The admin removed one of the two rings, kept the sword, added an item 5,
	// kept the weapon (and bound it), and swapped the body for item 6. The
	// form cannot say which ring went, so the kept ring is the first one.
	edited := New()
	edited.Items = []items.Item{{ItemId: 2}, {ItemId: 1}, {ItemId: 5}}
	edited.Equipment.Weapon = items.Item{ItemId: 3, CanNeverBeRemoved: true}
	edited.Equipment.Body = items.Item{ItemId: 6}

	edited.KeepItemIdentity(was)

	assert.Equal(t, sword.Id, edited.Items[0].Id)
	assert.Equal(t, []string{"rusty"}, edited.Items[0].Adjectives)
	assert.Equal(t, ring1.Id, edited.Items[1].Id, "the first unused item of that kind")
	assert.Equal(t, 4, edited.Items[1].Uses)
	assert.True(t, edited.Items[2].Id.IsZero(), "a new item gets its ID in Validate")

	assert.Equal(t, was.Equipment.Weapon.Id, edited.Equipment.Weapon.Id)
	assert.True(t, edited.Equipment.Weapon.CanNeverBeRemoved, "the admin's lock applies")
	assert.True(t, edited.Equipment.Body.Id.IsZero(), "a swapped slot holds a new item")
}

func TestKeepItemIdentityNeverCopiesAnId(t *testing.T) {
	was := New()
	was.Items = []items.Item{testItem(1)}

	// A posted ID that matches nothing the character had is dropped, so a
	// form cannot make a second item with an existing ID.
	stray := ids.New(ids.NoCell)
	edited := New()
	edited.Items = []items.Item{{ItemId: 1, Id: was.Items[0].Id}, {ItemId: 1, Id: was.Items[0].Id}, {ItemId: 9, Id: stray}}

	edited.KeepItemIdentity(was)

	assert.Equal(t, was.Items[0].Id, edited.Items[0].Id)
	assert.True(t, edited.Items[1].Id.IsZero(), "one item, one ID")
	assert.True(t, edited.Items[2].Id.IsZero())
}
```

- [ ] **Step 2:** Run `go test ./internal/characters/ -run KeepItemIdentity -count=1`. Expected: FAIL,
  `KeepItemIdentity undefined`.

- [ ] **Step 3: Implement** in `itemids.go` (add the `ids` import):

```go
// KeepItemIdentity is for an edit that names items only by ItemId (the admin
// user form): c holds the edited items, was the character before the edit.
// Each edited item that stands for an item was had (the same slot first for
// equipment, then the first unused held or worn item of the same ItemId) is
// replaced by that item whole, so it keeps its ID, uses and adjectives; a
// worn item keeps the edit's CanNeverBeRemoved. Any other edited item is new:
// its ID is cleared and Validate gives it one. No two items end with one ID.
func (c *Character) KeepItemIdentity(was *Character) {
	pool := append(append([]items.Item(nil), was.Items...), was.Equipment.GetAllItems()...)
	used := make([]bool, len(pool))

	take := func(edited items.Item, prefer ids.ID) items.Item {
		if edited.ItemId < 1 {
			return edited
		}
		pick := -1
		for i, old := range pool {
			if !used[i] && old.ItemId == edited.ItemId && !old.Id.IsZero() && old.Id == prefer {
				pick = i
				break
			}
		}
		if pick < 0 {
			for i, old := range pool {
				if !used[i] && old.ItemId == edited.ItemId {
					pick = i
					break
				}
			}
		}
		if pick < 0 {
			return edited.WithoutId()
		}
		used[pick] = true
		return pool[pick]
	}

	for _, slot := range AllSlots() {
		worn := c.Equipment.Get(slot)
		if worn == nil {
			continue
		}
		lock := worn.CanNeverBeRemoved
		*worn = take(*worn, was.Equipment.Get(slot).Id)
		if worn.ItemId > 0 {
			worn.CanNeverBeRemoved = lock
		}
	}
	for i := range c.Items {
		c.Items[i] = take(c.Items[i], c.Items[i].Id)
	}
}
```

  `Worn.GetAllItems` (`worn.go:146-154`) returns only slots with `ItemId > 0`, so the pool holds only
  real items. Equipment is matched before held items so a kept weapon takes its own slot's item.

- [ ] **Step 4:** Run the Step 2 command again. Expected: PASS.

- [ ] **Step 5: Use it in the handler.** In `api_v1_users.go`, move the decode into a function and
  call it from `apiV1PatchUser` in place of the lines from `updated := *u` through the
  `json.Unmarshal` error check:

```go
// decodeUserPatch applies an admin PATCH body to a copy of u. The form sends
// the complete desired skills, spellbook, items and equipment, and
// json.Unmarshal merges into existing maps and slice elements, so those start
// empty on the copy: a removed entry stays removed, an item never takes
// another item's fields, and the live record is not touched. Items the form
// kept keep their identity (characters.KeepItemIdentity). A body that leaves
// out the items or the equipment leaves them as they were.
func decodeUserPatch(u *accounts.Account, body []byte) (accounts.Account, error) {
	updated := *u
	if u.Character != nil {
		charCopy := *u.Character
		charCopy.Skills = nil
		charCopy.SpellBook = nil
		charCopy.Items = nil
		charCopy.Equipment = characters.Worn{}
		updated.Character = &charCopy
	}
	if err := json.Unmarshal(body, &updated); err != nil {
		return accounts.Account{}, err
	}
	if u.Character == nil || updated.Character == nil {
		return updated, nil
	}

	var present struct {
		Character *struct {
			Items     json.RawMessage
			Equipment json.RawMessage
		}
	}
	_ = json.Unmarshal(body, &present) // the body already decoded above
	if present.Character == nil || present.Character.Items == nil {
		updated.Character.Items = append([]items.Item(nil), u.Character.Items...)
	}
	if present.Character == nil || present.Character.Equipment == nil {
		updated.Character.Equipment = u.Character.Equipment
	}
	updated.Character.KeepItemIdentity(u.Character)
	return updated, nil
}
```

  `Account.Character` has only a YAML tag (`userrecord.go:40`), so its JSON key is `Character`;
  `encoding/json` matches keys case-insensitively, which is how the form's `character`
  (`users.html:1162`) decodes today, and how the untagged `present` struct above matches it too.
  `KeepItemIdentity` on unchanged lists hands every item back to itself, so leaving them out is safe.

  In the handler:

```go
	updated, err := decodeUserPatch(u, body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "malformed request body: "+err.Error())
		return
	}
```

  Delete the old comment block about detaching the character (its content is now on
  `decodeUserPatch`). Add the `characters` and `items` imports.

- [ ] **Step 6: Test the decode** in `internal/web/api_v1_users_test.go`:

```go
package web

import (
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func patchTestUser() *accounts.Account {
	u := &accounts.Account{UserId: 5, Username: "u"}
	u.Character = characters.New()
	a := items.Item{ItemId: 1, Id: ids.New(ids.NoCell), Uses: 4}
	b := items.Item{ItemId: 2, Id: ids.New(ids.NoCell), Uses: 7}
	u.Character.Items = make([]items.Item, 0, 4) // spare capacity, as a live slice has
	u.Character.Items = append(u.Character.Items, a, b)
	return u
}

func TestDecodeUserPatchKeepsItemsAndLeavesTheLiveRecord(t *testing.T) {
	u := patchTestUser()
	before := append([]items.Item(nil), u.Character.Items...)

	got, err := decodeUserPatch(u, []byte(`{"character":{"Items":[{"ItemId":2},{"ItemId":3}]}}`))
	require.NoError(t, err)

	assert.Equal(t, before, u.Character.Items, "the live record is untouched")
	require.Len(t, got.Character.Items, 2)
	assert.Equal(t, before[1].Id, got.Character.Items[0].Id, "the kept item keeps its ID")
	assert.Equal(t, 7, got.Character.Items[0].Uses, "and its state")
	assert.True(t, got.Character.Items[1].Id.IsZero(), "the added item is new")
	assert.Equal(t, 0, got.Character.Items[1].Uses, "and takes nothing from the removed one")
}

func TestDecodeUserPatchWithoutItemsKeepsThem(t *testing.T) {
	u := patchTestUser()
	got, err := decodeUserPatch(u, []byte(`{"Muted":true}`))
	require.NoError(t, err)
	assert.Equal(t, u.Character.Items, got.Character.Items)
}
```

  Run `go test ./internal/web/ -run DecodeUserPatch -count=1`.
  Expected: PASS.

- [ ] **Step 7: Run**

```bash
go build ./...
go test ./internal/characters/ ./internal/web/ ./internal/accounts/ -count=1
```

Expected: pass.

- [ ] **Step 8: Commit**

```bash
git add internal/characters/itemids.go internal/characters/itemids_test.go internal/web/api_v1_users.go internal/web/api_v1_users_test.go
git commit -F - <<'EOF'
web: admin Save User keeps the identity of the items it keeps

The form names items by ItemId only, and json.Unmarshal merged each
posted item into the old item at its position (and into the live
record). Items and equipment now decode onto empty lists and are
matched to the items the character had.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 5: Persistence tests

Each test below should pass on first run (Tasks 1-3 did the work). A test that fails is a finding:
report it with its output, do not weaken it.

**Files:**
- Test: `internal/accounts/bodyid_test.go`, `internal/rooms/rooms_test.go`,
  `internal/rooms/npcpersist_test.go`

- [ ] **Step 1: The user file.** In `bodyid_test.go`:

```go
func TestItemIdsSurviveTheUserFile(t *testing.T) {
	u := NewAccount(88, 0)
	u.Character.Items = []items.Item{{ItemId: 1, Id: ids.New(ids.NoCell)}}
	u.Character.Equipment.Weapon = items.Item{ItemId: 2, Id: ids.New(ids.NoCell)}
	b, err := yaml.Marshal(u)
	require.NoError(t, err)
	var back Account
	require.NoError(t, yaml.Unmarshal(b, &back))
	back.Character.Validate()
	assert.Equal(t, u.Character.Items[0].Id, back.Character.Items[0].Id)
	assert.Equal(t, u.Character.Equipment.Weapon.Id, back.Character.Equipment.Weapon.Id)
}
```

  If `Character.Validate` needs item specs or races to be loaded and fails without them, drop the
  `Validate` call and assert on the unmarshalled values; say so in the report.

- [ ] **Step 2: The room instance overlay.** In `rooms_test.go`, the instance save's YAML unmarshalled
  over a template keeps the instance's item IDs and drops the template's items:

```go
func TestInstanceItemIdsReplaceTheTemplates(t *testing.T) {
	tpl := NewEmptyRoom()
	tpl.Items = []items.Item{{ItemId: 1}} // a template item, no ID
	live := NewEmptyRoom()
	live.Items = []items.Item{{ItemId: 1, Id: ids.New(ids.NoCell)}, {ItemId: 2, Id: ids.New(ids.NoCell)}}

	b, err := yaml.Marshal(map[string]any{"items": live.Items}) // as saveRoomInstanceData writes a changed field
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(b, tpl)) // as LoadRoomInstance applies it

	require.Len(t, tpl.Items, 2)
	assert.Equal(t, live.Items[0].Id, tpl.Items[0].Id)
	assert.Equal(t, live.Items[1].Id, tpl.Items[1].Id)
}
```

- [ ] **Step 3: The NPC store.** In `npcpersist_test.go` (same flow as
  `TestUnloadRestsNPCsAndPrepareRestoresThem`, `:56-94`; add the `items` import):

```go
func TestRestoredNPCKeepsItsItemIds(t *testing.T) {
	persistSetup(t)
	id := liveTestNPC(t, 5, 2, 13)
	m := mobs.GetInstance(id)
	held := items.Item{ItemId: 1, Id: ids.New(ids.NoCell)}
	worn := items.Item{ItemId: 2, Id: ids.New(ids.NoCell)}
	m.Character.Items = append(m.Character.Items, held)
	m.Character.Equipment.Weapon = worn

	r := memRoom(4100)
	r.bodies = append(r.bodies, id)
	removeRoomFromMemory(r)
	require.Nil(t, mobs.GetInstance(id))

	r2 := NewEmptyRoom()
	r2.RoomId, r2.Zone = 4100, persistZone
	loadRoomNPCs(r2)
	addRoomToMemory(r2)
	r2.Prepare(false)

	back := mobs.GetInstance(id)
	require.NotNil(t, back)
	require.Len(t, back.Character.Items, 1)
	assert.Equal(t, held.Id, back.Character.Items[0].Id)
	assert.Equal(t, worn.Id, back.Character.Equipment.Weapon.Id)
}
```

  If restore's `Validate` drops items whose `ItemId` has no loaded spec, use the spec-carrying form
  (`Spec: &items.ItemSpec{ItemId: 1, Name: "tail"}`) for both items and say so.

- [ ] **Step 4: Run**

```bash
go test ./internal/accounts/ ./internal/rooms/ -count=1
```

Expected: pass.

- [ ] **Step 5: Commit**

```bash
git add internal/accounts/bodyid_test.go internal/rooms/rooms_test.go internal/rooms/npcpersist_test.go
git commit -F - <<'EOF'
items: tests that item IDs survive the user file, room saves and the NPC store

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

---

### Task 6: Verify and record (controller)

- [ ] **Step 1: Full checks** (`geomud-verifying`):

```bash
go build ./...
go test ./... -count=1 2>&1 | grep -vE 'no test files'
git diff --name-only master -- '*.go' | xargs -r gofmt -l
go run ./cmd/smoke
grep -rn --include=*.go -E 'uuid\.|\.UUID\b|ShorthandId|`!%d' . | grep -v _test.go
```

Expected: build passes; 34 packages ok (35 less the deleted `internal/uuid`) and only the 3 known
Windows failures; gofmt prints nothing; smoke passes (copyover step skipped on Windows); the grep
prints nothing.

- [ ] **Step 2: Records**, in one commit:
  - `PATCH_NOTES.md`: a line for slice 2c (item IDs saved and kept; NPCs from one template no longer
    share item IDs; admin Save User keeps item identity; `!N` forms and `internal/uuid` gone).
  - `status.md`: slice 2c done; next is slice 3 (messaging core).
  - Spec `2026-10-27-phase-4b-npcs-are-players-design.md`: in section 4 note 2c is built; add to
    "Notes for later slices": the admin page still sends items as `ItemId` only (identity is kept by
    matching, so two identical items may swap identity on an admin save; harmless while items of one
    `ItemId` differ only in uses and adjectives, which `KeepItemIdentity` keeps per matched item); a
    hand-written `id:` on an item in a room or mob template file would be shared by every copy
    (template writers strip IDs; nothing strips them from hand edits).
  - `.claude/skills/geomud-verifying/SKILL.md`: baseline 34 packages ok, with the reason.
  - Move this plan to `docs/completed/` when the PR merges.

```bash
git add PATCH_NOTES.md status.md 2026-10-27-phase-4b-npcs-are-players-design.md .claude/skills/geomud-verifying/SKILL.md
git commit -F - <<'EOF'
Phase 4b slice 2c done: notes, status, baseline

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
EOF
```

- [ ] **Step 3:** Push the branch to `origin` and open the PR with `gh pr create --repo pruuk/GeOMUD`
  only when Calabe says to.

## Done when

Every item ID is `ids.ID` (no `uuid` left in the tree); an item's ID survives the user file, room
instance saves and the NPC store (Task 5 tests); no two NPCs from one template share an item ID; no
template writer writes an item ID; admin Save User keeps the IDs of items it keeps; the guards pass;
build, tests at the new baseline, and smoke pass.
