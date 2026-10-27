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
	path := roomNPCPath(zone, roomId)
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
		for _, s := range f.Resting {
			index[s.Template.Character.Id] = roomId
		}
		return nil
	})
	if err != nil {
		return err
	}

	npcStoreMu.Lock()
	restingIndex = index
	npcStoreMu.Unlock()
	return nil
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
