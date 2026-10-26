package configs

import "testing"

func TestFilePathsValidateSetsDefaults(t *testing.T) {
	filePaths := FilePaths{}

	filePaths.Validate()

	if filePaths.DataFiles != "_datafiles/world/empty" {
		t.Fatalf("FilePaths.Validate() DataFiles = %q, want %q", filePaths.DataFiles, "_datafiles/world/empty")
	}
	if filePaths.HttpsCacheDir != "_datafiles/tls" {
		t.Fatalf("FilePaths.Validate() HttpsCacheDir = %q, want %q", filePaths.HttpsCacheDir, "_datafiles/tls")
	}
}
