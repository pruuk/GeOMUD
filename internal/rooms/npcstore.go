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
	"github.com/GoMudEngine/GoMud/internal/gametime"
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

// slotState is one spawn slot's memory: which NPC it owns, or, for a slot
// freed by a death, the round its respawn wait counts from. Matched to the
// template's SpawnInfo by Index and MobId.
type slotState struct {
	Index     int    `yaml:"index"`
	MobId     int    `yaml:"mobid"`
	NPC       ids.ID `yaml:"npc,omitempty"`
	Despawned uint64 `yaml:"despawned,omitempty"`
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

	// npcWriteFile and npcRemove are every sidecar and ledger write; crash
	// tests replace them to stop the disk after a given write.
	npcWriteFile = os.WriteFile
	npcRemove    = os.Remove

	npcStoreMu    sync.Mutex
	restingIndex  = map[ids.ID]int{}    // resting NPC ID to its room ID
	endingsLedger = map[ids.ID]ending{} // ended NPC ID to its ending

	// npcSidecars is the sidecars each NPC is in on disk and sidecarNPCs the
	// NPCs each sidecar holds (keyed by clean path). Writing an NPC to one
	// sidecar takes it out of every other, and a sidecar rewritten without an
	// NPC keeps it while no other sidecar holds it, so an NPC is on disk once:
	// one that wandered from a saved room is neither written twice nor lost
	// when that room rewrites before its new room saves. An NPC is in two
	// sidecars only after a crash between those two writes.
	npcSidecars = map[ids.ID][]string{}
	sidecarNPCs = map[string][]ids.ID{}
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
// empty file (nothing resting, no slots) deletes the sidecar instead.
func writeRoomNPCFile(zone string, roomId int, f roomNPCFile) error {
	path := filepath.Clean(roomNPCPath(zone, roomId))
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	return writeSidecarLocked(path, f)
}

// writeSidecarLocked writes the sidecar at path. An NPC the sidecar held that
// f leaves out stays in it while no other sidecar holds it (it wandered off
// and its new room has not saved yet), so dropping it never leaves it on disk
// nowhere. Once the file is written, each NPC in f is taken out of every
// other sidecar. The caller holds npcStoreMu.
func writeSidecarLocked(path string, f roomNPCFile) error {
	written := map[ids.ID]bool{}
	resting := append([]mobs.SavedNPC(nil), f.Resting...)
	for _, s := range resting {
		written[s.Template.Character.Id] = true
	}
	var old *roomNPCFile
	for _, id := range sidecarNPCs[path] {
		if written[id] || !onlyInSidecarLocked(id, path) {
			continue
		}
		if old == nil {
			o := readNPCFileAt(path)
			old = &o
		}
		for _, s := range old.Resting {
			if s.Template.Character.Id == id {
				resting = append(resting, s)
				break
			}
		}
	}
	f.Resting = resting

	if err := writeNPCFileAt(path, f); err != nil {
		return err
	}
	setSidecarHeldLocked(path, restingIds(f))

	for _, s := range f.Resting {
		id := s.Template.Character.Id
		if !written[id] {
			continue
		}
		for _, other := range append([]string(nil), npcSidecars[id]...) {
			if other != path {
				dropFromSidecarLocked(other, id)
			}
		}
	}
	return nil
}

// onlyInSidecarLocked reports whether the sidecar at path is the only one
// holding the NPC. The caller holds npcStoreMu.
func onlyInSidecarLocked(id ids.ID, path string) bool {
	in := npcSidecars[id]
	return len(in) == 1 && in[0] == path
}

// setSidecarHeldLocked records that the sidecar at path now holds exactly
// held. The caller holds npcStoreMu.
func setSidecarHeldLocked(path string, held []ids.ID) {
	for _, id := range sidecarNPCs[path] {
		kept := []string{}
		for _, p := range npcSidecars[id] {
			if p != path {
				kept = append(kept, p)
			}
		}
		if len(kept) == 0 {
			delete(npcSidecars, id)
		} else {
			npcSidecars[id] = kept
		}
	}
	if len(held) == 0 {
		delete(sidecarNPCs, path)
		return
	}
	sidecarNPCs[path] = held
	for _, id := range held {
		npcSidecars[id] = append(npcSidecars[id], path)
	}
}

// dropFromSidecarLocked removes one NPC from the sidecar at path. A failed
// write is logged and the record left as it is, since the file still holds
// the NPC. The caller holds npcStoreMu.
func dropFromSidecarLocked(path string, id ids.ID) {
	f := readNPCFileAt(path)
	kept := []mobs.SavedNPC{}
	for _, s := range f.Resting {
		if s.Template.Character.Id != id {
			kept = append(kept, s)
		}
	}
	if len(kept) != len(f.Resting) {
		f.Resting = kept
		if err := writeNPCFileAt(path, f); err != nil {
			mudlog.Error("npc store: could not take NPC out of old sidecar", "path", path, "npc", id.String(), "error", err)
			return
		}
	}
	held := []ids.ID{}
	for _, h := range sidecarNPCs[path] {
		if h != id {
			held = append(held, h)
		}
	}
	setSidecarHeldLocked(path, held)
}

// forgetNPCLocked takes an NPC out of every sidecar holding it (it ended).
// The caller holds npcStoreMu.
func forgetNPCLocked(id ids.ID) {
	for _, path := range append([]string(nil), npcSidecars[id]...) {
		dropFromSidecarLocked(path, id)
	}
}

// restingIds lists the NPCs in a sidecar, in file order.
func restingIds(f roomNPCFile) []ids.ID {
	out := []ids.ID{}
	for _, s := range f.Resting {
		out = append(out, s.Template.Character.Id)
	}
	return out
}

// deleteRoomNPCFile removes a room's sidecar (the room is being deleted) and
// forgets what it held.
func deleteRoomNPCFile(zone string, roomId int) error {
	path := filepath.Clean(roomNPCPath(zone, roomId))
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	if err := npcRemove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	setSidecarHeldLocked(path, nil)
	return nil
}

// moveRoomNPCFile moves a room's sidecar to another zone's folder (the room
// moved zone), keeping the record of which NPC is in which sidecar.
func moveRoomNPCFile(oldZone, newZone string, roomId int) error {
	oldPath := filepath.Clean(roomNPCPath(oldZone, roomId))
	newPath := filepath.Clean(roomNPCPath(newZone, roomId))
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(newPath), 0777); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	retrackPathLocked(oldPath, newPath)
	return nil
}

// renameZoneNPCFiles renames a zone's rooms.instances folder (instance saves
// and sidecars move with it; a missing folder is not an error) and keeps the
// record of which NPC is in which sidecar, all under the store lock.
func renameZoneNPCFiles(oldZone, newZone string) error {
	oldDir := filepath.Clean(filepath.Join(npcStoreRoot(), filepath.FromSlash(ZoneToFolder(oldZone))))
	newDir := filepath.Clean(filepath.Join(npcStoreRoot(), filepath.FromSlash(ZoneToFolder(newZone))))
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	if err := os.Rename(oldDir, newDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	oldDir += string(filepath.Separator)
	newDir += string(filepath.Separator)
	for path := range sidecarNPCs {
		if strings.HasPrefix(path, oldDir) {
			retrackPathLocked(path, newDir+strings.TrimPrefix(path, oldDir))
		}
	}
	return nil
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
		for i, p := range npcSidecars[id] {
			if p == oldPath {
				npcSidecars[id][i] = newPath
			}
		}
	}
}

// endRestingIn ends the NPCs resting in a room being deleted: each leaves the
// resting index and gets an ending (a death's, so a slot elsewhere that owns
// it refills after its respawn rate). saved is what the room held (its sidecar
// and resting list); an NPC saved with no spawn slot ends without a ledger
// entry, since no slot will read it. It returns the NPCs ended.
func endRestingIn(roomId int, saved []mobs.SavedNPC, round uint64) []ids.ID {
	slotless := map[ids.ID]bool{}
	for _, s := range saved {
		if s.SpawnSlot == 0 {
			slotless[s.Template.Character.Id] = true
		}
	}
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	ended := []ids.ID{}
	recorded := false
	for id, r := range restingIndex {
		if r != roomId {
			continue
		}
		delete(restingIndex, id)
		ended = append(ended, id)
		if !slotless[id] {
			endingsLedger[id] = ending{Round: round}
			recorded = true
		}
	}
	if recorded {
		if err := saveEndingsLocked(); err != nil {
			mudlog.Error("npc store: endings not saved", "room", roomId, "error", err)
		}
	}
	return ended
}

// endDeletedRoomNPCs settles a room being deleted: its resting NPCs end and
// its sidecar goes. NPCs live in it stay live, as a room delete has always
// left them. The sidecar goes before the endings are written, so a crash in
// between leaves slots that log their NPC lost, never NPCs that come back.
// An old copy it holds of an NPC now live in another room is first saved by
// that room, so deleting the file never leaves the NPC on disk nowhere.
func endDeletedRoomNPCs(zone string, roomId int) {
	f := readRoomNPCFile(zone, roomId)
	for _, s := range f.Resting {
		m := mobs.GetInstance(s.Template.Character.Id)
		if m == nil || m.Character.RoomId == roomId {
			continue
		}
		if other := roomManager.rooms[m.Character.RoomId]; other != nil {
			if err := saveRoomNPCs(other); err != nil {
				mudlog.Error("npc store: room not saved before a delete", "room", other.RoomId, "error", err)
			}
		}
	}
	saved := f.Resting
	if r := roomManager.rooms[roomId]; r != nil {
		saved = append(saved, r.resting...)
	}
	if err := deleteRoomNPCFile(zone, roomId); err != nil {
		mudlog.Error("npc store: sidecar of deleted room not removed", "room", roomId, "error", err)
	}
	if ended := endRestingIn(roomId, saved, util.GetRoundCount()); len(ended) > 0 {
		mudlog.Info("npc store: room deleted, its resting NPCs ended", "room", roomId, "count", len(ended))
	}
}

func writeNPCFileAt(path string, f roomNPCFile) error {
	if len(f.Resting) == 0 && len(f.Slots) == 0 {
		if err := npcRemove(path); err != nil && !os.IsNotExist(err) {
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
	return npcWriteFile(path, data, 0777)
}

// BuildRestingIndex scans every *.npcs.yaml under rooms.instances and fills
// the resting index. A file that does not parse is logged and skipped.
func BuildRestingIndex() error {
	index := map[ids.ID]int{}
	sidecarsOf := map[ids.ID][]string{}
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
			sidecarsOf[id] = append(sidecarsOf[id], clean)
			held[clean] = append(held[clean], id)
		}
		return nil
	})
	if err != nil {
		return err
	}

	npcStoreMu.Lock()
	restingIndex = index
	npcSidecars = sidecarsOf
	sidecarNPCs = held
	npcStoreMu.Unlock()
	return nil
}

// slotNPC says what a slot's NPC is doing (decision 14): live anywhere,
// resting anywhere, or ended; an NPC that is none of these is lost. An ended
// NPC's ending is read, not consumed: settleSlotsSaved consumes it once the
// freed slot is on disk.
func slotNPC(id ids.ID) (state string, end ending) {
	if mobs.GetInstance(id) != nil {
		return "live", ending{}
	}
	if IsResting(id) {
		return "resting", ending{}
	}
	if e, ok := peekEnding(id); ok {
		return "ended", e
	}
	return "lost", ending{}
}

// slotSettling collects what a pass of settleSlot over a room's slots freed.
type slotSettling struct {
	freed bool     // a slot was freed, so the sidecar is rewritten
	ended []ids.ID // endings to consume once it is
}

// settle applies settleSlot to slot idx and records what it freed.
func (s *slotSettling) settle(r *Room, idx int, roundNow uint64) {
	npc := r.SpawnInfo[idx].InstanceId
	state := r.settleSlot(idx, roundNow)
	if state == "ended" || state == "lost" {
		s.freed = true
	}
	if state == "ended" {
		s.ended = append(s.ended, npc)
	}
}

// save rewrites the room's sidecar if a slot was freed, and only then
// consumes the endings that freed them: a crash in between leaves an ending
// no slot will read, never a slot that owns an NPC that is nowhere.
func (s *slotSettling) save(r *Room) {
	if s.freed {
		r.settleSlotsSaved(s.ended)
	}
}

// settleSlot applies slotNPC to spawn slot idx. A live or resting NPC keeps
// its slot. An ended one frees it: at once after an admin's despawn, else
// from the round it ended, plus the respawn rate. A lost one is logged and
// frees it from roundNow. It returns the state seen ("" for an empty slot).
// Callers go through slotSettling, which consumes the ending.
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

// settleSlotsSaved rewrites the room's sidecar after settleSlot freed a slot
// (with its wait), then consumes the endings that freed them. If the sidecar
// is not written the endings stay, so a restart still reads them.
func (r *Room) settleSlotsSaved(ended []ids.ID) {
	if err := saveRoomNPCs(r); err != nil {
		mudlog.Error("npc store: sidecar not rewritten after a slot freed", "room", r.RoomId, "error", err)
		return
	}
	takeEndings(ended)
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

// unindexNPC takes a live NPC out of the index, wherever it places it.
func unindexNPC(id ids.ID) {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	delete(restingIndex, id)
}

// roomNPCFileOf is the sidecar a room saves now: a snapshot of each live NPC
// in it, its resting NPCs (one entry per ID: an NPC both live and resting is
// written as it is live, and its resting entry is dropped), every spawn slot
// that owns an NPC and every slot still waiting after a death.
func roomNPCFileOf(r *Room) roomNPCFile {
	var f roomNPCFile
	seen := map[ids.ID]bool{}
	for _, id := range r.mobIds() {
		if seen[id] {
			continue
		}
		if m := mobs.GetInstance(id); m != nil {
			seen[id] = true
			f.Resting = append(f.Resting, mobs.Snapshot(m))
		}
	}
	var resting []mobs.SavedNPC
	for _, s := range r.resting {
		id := s.Template.Character.Id
		if seen[id] {
			continue
		}
		seen[id] = true
		f.Resting = append(f.Resting, s)
		resting = append(resting, s)
	}
	if len(resting) != len(r.resting) {
		r.resting = resting
		// SaveRoomInstance saves a copy: drop them from the room itself too.
		if mem := roomManager.rooms[r.RoomId]; mem != nil && mem != r {
			mem.resting = resting
		}
	}
	roundNow := util.GetRoundCount()
	for idx, si := range r.SpawnInfo {
		if !si.InstanceId.IsZero() {
			f.Slots = append(f.Slots, slotState{Index: idx, MobId: si.MobId, NPC: si.InstanceId})
			continue
		}
		if slotWaiting(si, roundNow) {
			f.Slots = append(f.Slots, slotState{Index: idx, MobId: si.MobId, Despawned: si.DespawnedRound})
		}
	}
	return f
}

// slotWaiting reports whether a free NPC slot is still waiting out its
// respawn rate after a death, so its wait must be saved.
func slotWaiting(si SpawnInfo, roundNow uint64) bool {
	if si.MobId == 0 || si.DespawnedRound == 0 {
		return false
	}
	return roundNow < gametime.GetDate(si.DespawnedRound).AddPeriod(si.RespawnRate)
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
	r.resting = nil
	for _, s := range f.Resting {
		id := s.Template.Character.Id
		if mobs.GetInstance(id) != nil {
			// An old copy of an NPC that wandered off and is live elsewhere:
			// it stays on disk until its new room saves, and is not resting.
			continue
		}
		r.resting = append(r.resting, s)
		indexResting(id, r.RoomId)
	}
	for _, s := range f.Slots {
		if s.Index < len(r.SpawnInfo) && r.SpawnInfo[s.Index].MobId == s.MobId {
			r.SpawnInfo[s.Index].InstanceId = s.NPC
			r.SpawnInfo[s.Index].DespawnedRound = s.Despawned
			continue
		}
		mudlog.Error("npc store: slot state dropped", "room", r.RoomId, "index", s.Index, "mobid", s.MobId, "npc", s.NPC.String())
	}
}

// carryNPCState copies a room's resting NPCs, slot owners and slot respawn
// waits onto the room replacing it in a template edit, matching slots by
// index and MobId.
func carryNPCState(from, to *Room) {
	to.resting = append([]mobs.SavedNPC(nil), from.resting...)
	for idx := range to.SpawnInfo {
		if idx < len(from.SpawnInfo) && to.SpawnInfo[idx].InstanceId.IsZero() && from.SpawnInfo[idx].MobId == to.SpawnInfo[idx].MobId {
			to.SpawnInfo[idx].InstanceId = from.SpawnInfo[idx].InstanceId
			if from.SpawnInfo[idx].InstanceId.IsZero() {
				to.SpawnInfo[idx].DespawnedRound = from.SpawnInfo[idx].DespawnedRound
			}
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
		m, err := mobs.Restore(s)
		if err != nil {
			unindexResting(id, r.RoomId)
			mudlog.Error("npc store: restore failed", "room", r.RoomId, "npc", id.String(), "error", err)
			continue
		}
		// Live now: out of the index whatever room it names (a crash can
		// leave it in two sidecars). The sidecar write below takes it out of
		// any other sidecar.
		unindexNPC(id)
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
		if err := npcRemove(path); err != nil && !os.IsNotExist(err) {
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
	return npcWriteFile(path, data, 0777)
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

// NPCEnded is called as an NPC ends (death, or an admin's despawn with
// immediate set), before it leaves memory. An NPC with a spawn slot gets a
// ledger entry for that slot to read; one without ends with none, so the
// ledger holds only entries a slot will consume. Either way it first leaves
// every sidecar holding it, so a crash before the ledger write leaves a slot
// that logs its NPC lost, never an NPC that comes back.
func NPCEnded(m *mobs.Mob, immediate bool) {
	id := m.Character.Id
	round := util.GetRoundCount()
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	forgetNPCLocked(id)
	if m.SpawnSlot > 0 {
		endingsLedger[id] = ending{Round: round, Immediate: immediate}
		if err := saveEndingsLocked(); err != nil {
			mudlog.Error("NPCEnded", "id", id.String(), "error", err)
		}
	}
}

// RestNPC rests one live NPC in this room (decision 15: boredom and charm
// expiry): saved into the room's sidecar, out of the room and out of memory,
// listed in the resting index. No ending. It reports false, leaving the NPC
// live, for an ephemeral room (nothing there is saved) or a failed save.
func (r *Room) RestNPC(m *mobs.Mob) bool {
	if r.IsEphemeral() {
		return false
	}
	id := m.Character.Id
	saved := mobs.Snapshot(m)
	// Live in the room, the save writes it; one not listed here is written
	// from its resting entry.
	before := len(r.resting)
	inRoom := false
	for _, b := range r.mobIds() {
		if b == id {
			inRoom = true
			break
		}
	}
	if !inRoom {
		r.resting = append(r.resting, saved)
	}
	if err := saveRoomNPCs(r); err != nil {
		mudlog.Error("npc store: NPC not rested, its room could not be saved", "room", r.RoomId, "npc", id.String(), "error", err)
		r.resting = r.resting[:before]
		return false
	}
	r.RemoveMob(id)
	mobs.Rest(id)
	if inRoom {
		r.resting = append(r.resting, saved)
	}
	indexResting(id, r.RoomId)
	return true
}

// peekEnding reads an NPC's ending without consuming it.
func peekEnding(id ids.ID) (ending, bool) {
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	e, ok := endingsLedger[id]
	return e, ok
}

// takeEndings removes these NPCs' endings (spawn slots consumed them) and
// saves the ledger once.
func takeEndings(ended []ids.ID) {
	if len(ended) == 0 {
		return
	}
	npcStoreMu.Lock()
	defer npcStoreMu.Unlock()
	changed := false
	for _, id := range ended {
		if _, ok := endingsLedger[id]; ok {
			delete(endingsLedger, id)
			changed = true
		}
	}
	if !changed {
		return
	}
	if err := saveEndingsLocked(); err != nil {
		mudlog.Error("takeEndings", "count", len(ended), "error", err)
	}
}

// EndEphemeralNPC ends an NPC leaving memory in an ephemeral room, where
// nothing is saved so it cannot rest: it ends as a death would (a slot that
// owns it refills after its respawn rate), leaves every sidecar, and is
// destroyed and taken out of the room.
func (r *Room) EndEphemeralNPC(m *mobs.Mob) {
	id := m.Character.Id
	NPCEnded(m, false)
	mobs.DestroyInstance(id)
	r.RemoveMob(id)
}

// takeEnding reads and removes an NPC's ending and saves the ledger.
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
