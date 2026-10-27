package rooms

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/ids"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
	"gopkg.in/yaml.v2"
)

// The NPC store (Phase 4b slice 2b): the per-room sidecar of resting NPCs and
// spawn slot state, the in-memory index of resting NPCs, and the ledger of
// NPCs that ended. The index and ledger are guarded by one mutex because
// autosave runs on the world goroutine and copyover saves from a command.

const (
	npcSidecarSuffix = ".npcs.yaml"
	npcEndingsFile   = "npcs.endings.yaml"
)

// roomNPCFile is a room's sidecar, rooms.instances/<zone>/<id>.npcs.yaml:
// the NPCs resting in the room and the room's spawn slot state.
type roomNPCFile struct {
	Resting []mobs.SavedNPC `yaml:"resting,omitempty"`
	Slots   []slotState     `yaml:"slots,omitempty"`
}

// slotState is one spawn slot's memory: which NPC it owns. Matched to the
// template's SpawnInfo by Index and MobId.
type slotState struct {
	Index int    `yaml:"index"`
	MobId int    `yaml:"mobid"`
	NPC   ids.ID `yaml:"npc"`
}

// ending is an NPC that ended (decision 15): its round, and whether its slot
// may respawn at once (an admin's despawn) or waits its respawn rate (death).
type ending struct {
	Round     uint64 `yaml:"round"`
	Immediate bool   `yaml:"immediate,omitempty"`
}

var (
	// npcStoreRoot returns the rooms.instances directory, the same base
	// SaveRoomInstance uses. Tests replace it.
	npcStoreRoot = func() string {
		return util.FilePath(configs.GetFilePathsConfig().DataFiles.String(), `/rooms.instances/`)
	}

	npcStoreMu    sync.Mutex
	restingIndex  = map[ids.ID]int{}    // resting NPC ID to its room ID
	endingsLedger = map[ids.ID]ending{} // ended NPC ID to its ending

	// npcSidecarOf is the sidecar each NPC was last written to and
	// sidecarNPCs the NPCs each sidecar holds (keyed by clean path). Writing
	// an NPC to one sidecar takes it out of the one it was in before, so an
	// NPC is in at most one sidecar: one that wandered from a saved room to
	// another that then saved is not on disk twice.
	npcSidecarOf = map[ids.ID]string{}
	sidecarNPCs  = map[string][]ids.ID{}
)

// roomNPCPath is the sidecar path for the room with this zone and ID, built
// the way SaveRoomInstance builds the instance path.
func roomNPCPath(zone string, roomId int) string {
	return filepath.Join(npcStoreRoot(), filepath.FromSlash(ZoneToFolder(zone)), fmt.Sprintf("%d%s", roomId, npcSidecarSuffix))
}

func endingsPath() string {
	return filepath.Join(npcStoreRoot(), npcEndingsFile)
}

// readRoomNPCFile reads the sidecar of the room with this zone and ID. A
// missing or unparseable file reads as empty (the parse failure is logged).
func readRoomNPCFile(zone string, roomId int) roomNPCFile {
	return readNPCFileAt(roomNPCPath(zone, roomId))
}

func readNPCFileAt(path string) roomNPCFile {
	var f roomNPCFile
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			mudlog.Error("readRoomNPCFile", "path", path, "error", err)
		}
		return f
	}
	if err := yaml.Unmarshal(data, &f); err != nil {
		mudlog.Error("readRoomNPCFile", "path", path, "error", err)
		return roomNPCFile{}
	}
	return f
}

// writeRoomNPCFile writes the sidecar of the room with this zone and ID. An
// empty file (nothing resting, no slots) deletes the sidecar instead. Each NPC
// written is taken out of any other sidecar it was last written to.
func writeRoomNPCFile(zone string, roomId int, f roomNPCFile) error {
	path := filepath.Clean(roomNPCPath(zone, roomId))
	if err := writeNPCFileAt(path, f); err != nil {
		return err
	}
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	trackSidecarLocked(path, f)
	return nil
}

// trackSidecarLocked records that the sidecar at path now holds f's NPCs and
// takes each of them out of the sidecar it was in before. The caller holds
// npcStoreMu.
func trackSidecarLocked(path string, f roomNPCFile) {
	for _, id := range sidecarNPCs[path] {
		if npcSidecarOf[id] == path {
			delete(npcSidecarOf, id)
		}
	}
	held := make([]ids.ID, 0, len(f.Resting))
	for _, s := range f.Resting {
		id := s.Template.Character.Id
		if prev, ok := npcSidecarOf[id]; ok && prev != path {
			dropFromSidecarLocked(prev, id)
		}
		npcSidecarOf[id] = path
		held = append(held, id)
	}
	if len(held) == 0 {
		delete(sidecarNPCs, path)
	} else {
		sidecarNPCs[path] = held
	}
}

// dropFromSidecarLocked removes one NPC from the sidecar at path. The caller
// holds npcStoreMu.
func dropFromSidecarLocked(path string, id ids.ID) {
	f := readNPCFileAt(path)
	kept := f.Resting[:0]
	for _, s := range f.Resting {
		if s.Template.Character.Id != id {
			kept = append(kept, s)
		}
	}
	if len(kept) != len(f.Resting) {
		f.Resting = kept
		if err := writeNPCFileAt(path, f); err != nil {
			mudlog.Error("npc store: could not take NPC out of old sidecar", "path", path, "npc", id.String(), "error", err)
		}
	}
	held := sidecarNPCs[path][:0]
	for _, h := range sidecarNPCs[path] {
		if h != id {
			held = append(held, h)
		}
	}
	if len(held) == 0 {
		delete(sidecarNPCs, path)
	} else {
		sidecarNPCs[path] = held
	}
}

// deleteRoomNPCFile removes a room's sidecar (the room is being deleted) and
// forgets what it held.
func deleteRoomNPCFile(zone string, roomId int) error {
	path := filepath.Clean(roomNPCPath(zone, roomId))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	trackSidecarLocked(path, roomNPCFile{})
	return nil
}

// moveRoomNPCFile moves a room's sidecar to another zone's folder (the room
// moved zone), keeping the record of which NPC is in which sidecar.
func moveRoomNPCFile(oldZone, newZone string, roomId int) error {
	oldPath := filepath.Clean(roomNPCPath(oldZone, roomId))
	newPath := filepath.Clean(roomNPCPath(newZone, roomId))
	if err := os.MkdirAll(filepath.Dir(newPath), 0777); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	retrackPathLocked(oldPath, newPath)
	return nil
}

// renameZoneNPCFiles keeps the record of which NPC is in which sidecar after
// a zone's rooms.instances folder was renamed (the files moved with it).
func renameZoneNPCFiles(oldZone, newZone string) {
	oldDir := filepath.Clean(filepath.Join(npcStoreRoot(), filepath.FromSlash(ZoneToFolder(oldZone)))) + string(filepath.Separator)
	newDir := filepath.Clean(filepath.Join(npcStoreRoot(), filepath.FromSlash(ZoneToFolder(newZone)))) + string(filepath.Separator)
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	for path := range sidecarNPCs {
		if strings.HasPrefix(path, oldDir) {
			retrackPathLocked(path, newDir+strings.TrimPrefix(path, oldDir))
		}
	}
}

// retrackPathLocked records that the sidecar at oldPath is now at newPath.
// The caller holds npcStoreMu.
func retrackPathLocked(oldPath, newPath string) {
	held, ok := sidecarNPCs[oldPath]
	if !ok {
		return
	}
	delete(sidecarNPCs, oldPath)
	sidecarNPCs[newPath] = held
	for _, id := range held {
		if npcSidecarOf[id] == oldPath {
			npcSidecarOf[id] = newPath
		}
	}
}

// endRestingIn ends the NPCs resting in a room being deleted: each leaves the
// resting index and gets an ending (a death's, so a slot elsewhere that owns
// it refills after its respawn rate). It returns the NPCs ended.
func endRestingIn(roomId int, round uint64) []ids.ID {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	ended := []ids.ID{}
	for id, r := range restingIndex {
		if r != roomId {
			continue
		}
		delete(restingIndex, id)
		endingsLedger[id] = ending{Round: round}
		ended = append(ended, id)
	}
	if len(ended) > 0 {
		if err := saveEndingsLocked(); err != nil {
			mudlog.Error("npc store: endings not saved", "room", roomId, "error", err)
		}
	}
	return ended
}

// endDeletedRoomNPCs settles a room being deleted: its resting NPCs end and
// its sidecar goes. NPCs live in it stay live, as a room delete has always
// left them.
func endDeletedRoomNPCs(zone string, roomId int) {
	if ended := endRestingIn(roomId, util.GetRoundCount()); len(ended) > 0 {
		mudlog.Info("npc store: room deleted, its resting NPCs ended", "room", roomId, "count", len(ended))
	}
	if err := deleteRoomNPCFile(zone, roomId); err != nil {
		mudlog.Error("npc store: sidecar of deleted room not removed", "room", roomId, "error", err)
	}
}

func writeNPCFileAt(path string, f roomNPCFile) error {
	if len(f.Resting) == 0 && len(f.Slots) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0777)
}

// BuildRestingIndex scans every *.npcs.yaml under rooms.instances and fills
// the resting index. A file that does not parse is logged and skipped.
func BuildRestingIndex() error {
	index := map[ids.ID]int{}
	sidecarOf := map[ids.ID]string{}
	held := map[string][]ids.ID{}
	root := npcStoreRoot()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, npcSidecarSuffix) {
			return nil
		}
		roomId, convErr := strconv.Atoi(strings.TrimSuffix(name, npcSidecarSuffix))
		if convErr != nil {
			mudlog.Error("BuildRestingIndex", "path", path, "error", "file name is not a room ID")
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			mudlog.Error("BuildRestingIndex", "path", path, "error", readErr)
			return nil
		}
		var f roomNPCFile
		if yamlErr := yaml.Unmarshal(data, &f); yamlErr != nil {
			mudlog.Error("BuildRestingIndex", "path", path, "error", yamlErr)
			return nil
		}
		clean := filepath.Clean(path)
		for _, s := range f.Resting {
			id := s.Template.Character.Id
			index[id] = roomId
			sidecarOf[id] = clean
			held[clean] = append(held[clean], id)
		}
		return nil
	})
	if err != nil {
		return err
	}

	npcStoreMu.Lock()
	restingIndex = index
	npcSidecarOf = sidecarOf
	sidecarNPCs = held
	npcStoreMu.Unlock()
	return nil
}

// slotNPC says what a slot's NPC is doing (decision 14): live anywhere,
// resting anywhere, or ended; an NPC that is none of these is lost. An ended
// NPC's ending is consumed (and the ledger saved).
func slotNPC(id ids.ID) (state string, end ending) {
	if mobs.GetInstance(id) != nil {
		return "live", ending{}
	}
	if IsResting(id) {
		return "resting", ending{}
	}
	if e, ok := takeEnding(id); ok {
		return "ended", e
	}
	return "lost", ending{}
}

// settleSlot applies slotNPC to spawn slot idx. A live or resting NPC keeps
// its slot. An ended one frees it: at once after an admin's despawn, else
// from the round it ended, plus the respawn rate. A lost one is logged and
// frees it from roundNow. It returns the state seen ("" for an empty slot).
func (r *Room) settleSlot(idx int, roundNow uint64) string {
	si := r.SpawnInfo[idx]
	if si.InstanceId.IsZero() {
		return ""
	}
	state, end := slotNPC(si.InstanceId)
	switch state {
	case "live", "resting":
		return state
	case "ended":
		si.DespawnedRound = end.Round
		if end.Immediate {
			si.DespawnedRound = 0
		}
	default:
		mudlog.Error("npc store: slot lost its NPC", "room", r.RoomId, "npc", si.InstanceId.String())
		si.DespawnedRound = roundNow
	}
	si.InstanceId = ids.ID{}
	r.SpawnInfo[idx] = si
	return state
}

// settleSlotsSaved rewrites the room's sidecar after settleSlot freed a slot,
// so a consumed ending is never left on disk as a slot that still owns the
// NPC (which a restart would call lost).
func (r *Room) settleSlotsSaved() {
	if err := saveRoomNPCs(r); err != nil {
		mudlog.Error("npc store: sidecar not rewritten after a slot freed", "room", r.RoomId, "error", err)
	}
}

// indexResting records that the NPC rests in roomId, unless the index already
// places it somewhere.
func indexResting(id ids.ID, roomId int) {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	if _, ok := restingIndex[id]; !ok {
		restingIndex[id] = roomId
	}
}

// unindexResting takes the NPC out of the index if the index places it in
// roomId (it was restored there).
func unindexResting(id ids.ID, roomId int) {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	if r, ok := restingIndex[id]; ok && r == roomId {
		delete(restingIndex, id)
	}
}

// roomNPCFileOf is the sidecar a room saves now: its resting NPCs, a snapshot
// of each live NPC in it (one entry per ID, so an NPC both resting and live
// is written once), and every spawn slot that owns an NPC.
func roomNPCFileOf(r *Room) roomNPCFile {
	var f roomNPCFile
	seen := map[ids.ID]bool{}
	for _, s := range r.resting {
		id := s.Template.Character.Id
		if seen[id] {
			continue
		}
		seen[id] = true
		f.Resting = append(f.Resting, s)
	}
	for _, id := range r.mobIds() {
		if seen[id] {
			continue
		}
		if m := mobs.GetInstance(id); m != nil {
			seen[id] = true
			f.Resting = append(f.Resting, mobs.Snapshot(m))
		}
	}
	for idx, si := range r.SpawnInfo {
		if !si.InstanceId.IsZero() {
			f.Slots = append(f.Slots, slotState{Index: idx, MobId: si.MobId, NPC: si.InstanceId})
		}
	}
	return f
}

// saveRoomNPCs writes a room's sidecar. Ephemeral rooms have none.
func saveRoomNPCs(r *Room) error {
	if r.IsEphemeral() {
		return nil
	}
	return writeRoomNPCFile(r.Zone, r.RoomId, roomNPCFileOf(r))
}

// loadRoomNPCs reads a room's sidecar into it as it comes into memory: its
// resting NPCs, and slot state merged onto the template's spawn slots by
// index and MobId (a slot a builder changed starts fresh).
func loadRoomNPCs(r *Room) {
	if r.IsEphemeral() {
		return
	}
	f := readRoomNPCFile(r.Zone, r.RoomId)
	r.resting = f.Resting
	for _, s := range f.Resting {
		indexResting(s.Template.Character.Id, r.RoomId)
	}
	for _, s := range f.Slots {
		if s.Index < len(r.SpawnInfo) && r.SpawnInfo[s.Index].MobId == s.MobId {
			r.SpawnInfo[s.Index].InstanceId = s.NPC
			continue
		}
		mudlog.Error("npc store: slot state dropped", "room", r.RoomId, "index", s.Index, "mobid", s.MobId, "npc", s.NPC.String())
	}
}

// carryNPCState copies a room's resting NPCs and slot owners onto the room
// replacing it in a template edit, matching slots by index and MobId.
func carryNPCState(from, to *Room) {
	to.resting = append([]mobs.SavedNPC(nil), from.resting...)
	for idx := range to.SpawnInfo {
		if idx < len(from.SpawnInfo) && to.SpawnInfo[idx].InstanceId.IsZero() && from.SpawnInfo[idx].MobId == to.SpawnInfo[idx].MobId {
			to.SpawnInfo[idx].InstanceId = from.SpawnInfo[idx].InstanceId
		}
	}
}

// restoreResting brings a room's resting NPCs back into memory and into the
// room, then rewrites its sidecar. An NPC that cannot be restored (it is
// already live, or the save is bad) is logged and dropped.
func (r *Room) restoreResting() {
	if len(r.resting) == 0 {
		return
	}
	for _, s := range r.resting {
		id := s.Template.Character.Id
		unindexResting(id, r.RoomId)
		m, err := mobs.Restore(s)
		if err != nil {
			mudlog.Error("npc store: restore failed", "room", r.RoomId, "npc", id.String(), "error", err)
			continue
		}
		m.Character.RoomId = r.RoomId
		m.Character.Zone = r.Zone
		r.removeBody(id) // never listed twice
		r.bodies = append(r.bodies, id)
	}
	r.resting = nil
	if r.MobCt() > 0 {
		roomManager.roomsWithMobs[r.RoomId] = r.MobCt()
	}
	if err := saveRoomNPCs(r); err != nil {
		mudlog.Error("npc store: sidecar not rewritten after restore", "room", r.RoomId, "error", err)
	}
}

// IsResting reports whether the NPC is saved to disk and out of memory.
func IsResting(id ids.ID) bool {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	_, ok := restingIndex[id]
	return ok
}

// LoadEndings reads the endings ledger. A missing file is an empty ledger.
func LoadEndings() error {
	data, err := os.ReadFile(endingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			npcStoreMu.Lock()
			endingsLedger = map[ids.ID]ending{}
			npcStoreMu.Unlock()
			return nil
		}
		return err
	}
	ledger := map[ids.ID]ending{}
	if err := yaml.Unmarshal(data, &ledger); err != nil {
		return err
	}
	npcStoreMu.Lock()
	endingsLedger = ledger
	npcStoreMu.Unlock()
	return nil
}

// saveEndings writes the ledger. The caller holds npcStoreMu.
func saveEndingsLocked() error {
	path := endingsPath()
	if len(endingsLedger) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := yaml.Marshal(endingsLedger)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0777)
}

// saveEndings writes the ledger to disk.
func saveEndings() error {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	return saveEndingsLocked()
}

// RecordEnding adds an ended NPC to the ledger and saves it.
func RecordEnding(id ids.ID, round uint64, immediate bool) {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	endingsLedger[id] = ending{Round: round, Immediate: immediate}
	if err := saveEndingsLocked(); err != nil {
		mudlog.Error("RecordEnding", "id", id.String(), "error", err)
	}
}

// takeEnding reads and removes an NPC's ending (a spawn slot consumes it) and
// saves the ledger.
func takeEnding(id ids.ID) (ending, bool) {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	e, ok := endingsLedger[id]
	if !ok {
		return ending{}, false
	}
	delete(endingsLedger, id)
	if err := saveEndingsLocked(); err != nil {
		mudlog.Error("takeEnding", "id", id.String(), "error", err)
	}
	return e, true
}
