package items

import (
	"io/fs"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/fileloader"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"gopkg.in/yaml.v2"
)

var (
	pluginFileSystems []fileloader.ReadableGroupFS
)

// RegisterFS registers a plugin file system to be searched when loading item
// data files. Must be called before LoadDataFiles().
func RegisterFS(f ...fileloader.ReadableGroupFS) {
	pluginFileSystems = append(pluginFileSystems, f...)
}

// loadPluginItems walks every sub-filesystem of every registered plugin FS,
// reading item YAML files from an "items/" prefix and merging them into dst.
func loadPluginItems(dst map[int]*ItemSpec) {
	for _, groupFS := range pluginFileSystems {
		for subFS := range groupFS.AllFileSubSystems {
			loadItemsFromFS(subFS, dst)
		}
	}
}

func loadItemsFromFS(subFS fs.ReadFileFS, dst map[int]*ItemSpec) {
	// PluginFiles does not support directory traversal, so use KnownPaths
	// when available to enumerate files directly.
	if pl, ok := subFS.(fileloader.PathLister); ok {
		for _, path := range pl.KnownPaths() {
			if !strings.HasPrefix(path, `items/`) || !strings.HasSuffix(path, `.yaml`) {
				continue
			}
			loadItemFileFromFS(subFS, path, dst)
		}
		return
	}

	// Fallback: standard directory walk for FSes that support it.
	_ = fs.WalkDir(subFS, `items`, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, `.yaml`) {
			return nil
		}
		loadItemFileFromFS(subFS, path, dst)
		return nil
	})
}

func loadItemFileFromFS(subFS fs.ReadFileFS, path string, dst map[int]*ItemSpec) {
	b, err := subFS.ReadFile(path)
	if err != nil {
		mudlog.Error("items.loadItemsFromFS", "path", path, "error", err)
		return
	}

	var spec ItemSpec
	if err := yaml.Unmarshal(b, &spec); err != nil {
		mudlog.Error("items.loadItemsFromFS", "path", path, "error", err)
		return
	}

	// Validate the Filepath() claim matches the actual path so the same
	// rules as the disk loader apply.
	if !strings.HasSuffix(path, spec.Filepath()) {
		mudlog.Error("items.loadItemsFromFS", "path", path, "expected suffix", spec.Filepath(), "error", "filepath mismatch")
		return
	}

	if err := spec.Validate(); err != nil {
		mudlog.Error("items.loadItemsFromFS", "path", path, "error", err)
		return
	}

	if _, exists := dst[spec.ItemId]; exists {
		mudlog.Error("items.loadItemsFromFS", "itemId", spec.ItemId, "path", path, "error", "duplicate item id")
		return
	}

	dst[spec.ItemId] = &spec
}
