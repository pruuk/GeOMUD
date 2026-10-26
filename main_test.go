package main

import (
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/util"
)

// The base world must carry every folder the engine requires at boot.
func TestBaseWorldHasRequiredFolders(t *testing.T) {
	world := filepath.Join("_datafiles", "world", "empty")
	if err := util.ValidateWorldFiles(world, requiredWorldFolders); err != nil {
		t.Fatal(err)
	}
}
