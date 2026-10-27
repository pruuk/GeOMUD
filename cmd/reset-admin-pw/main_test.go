package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func TestFindAdminUser(t *testing.T) {
	chdirToRepoRoot(t)
	overridePath := writeTestConfig(t)
	t.Setenv("CONFIG_PATH", overridePath)
	mudlog.SetupLogger(nil, "", "", false)

	if err := configs.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig() error = %v", err)
	}
	createUserIndex(t)

	adminUser, err := findAdminUser()
	if err != nil {
		t.Fatalf("findAdminUser() error = %v", err)
	}

	if adminUser.Username != "admin" {
		t.Fatalf("findAdminUser() username = %q, want %q", adminUser.Username, "admin")
	}
	if adminUser.Role != accounts.RoleAdmin {
		t.Fatalf("findAdminUser() role = %q, want %q", adminUser.Role, accounts.RoleAdmin)
	}
}

func TestPromptForPasswordMismatch(t *testing.T) {
	chdirToRepoRoot(t)
	inputFile := writePromptInput(t, "first\nsecond\n")
	outputFile := tempOutputFile(t)

	_, err := promptForPassword(inputFile, outputFile, -1)
	if err != errPasswordMismatch {
		t.Fatalf("promptForPassword() error = %v, want %v", err, errPasswordMismatch)
	}
}

func TestResetAdminPasswordPersistsHash(t *testing.T) {
	chdirToRepoRoot(t)
	overridePath := writeTestConfig(t)
	t.Setenv("CONFIG_PATH", overridePath)
	mudlog.SetupLogger(nil, "", "", false)

	if err := configs.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig() error = %v", err)
	}
	createUserIndex(t)

	adminUser, err := findAdminUser()
	if err != nil {
		t.Fatalf("findAdminUser() error = %v", err)
	}

	const newPassword = "new-secret"
	if err := adminUser.SetPassword(newPassword); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}
	if err := accounts.SaveUser(*adminUser); err != nil {
		t.Fatalf("SaveUser() error = %v", err)
	}

	reloadedUser, err := accounts.LoadUser(adminUser.Username, true)
	if err != nil {
		t.Fatalf("LoadUser() error = %v", err)
	}

	if !reloadedUser.PasswordMatches(newPassword) {
		t.Fatal("PasswordMatches() returned false for updated password")
	}
	if reloadedUser.Password == newPassword {
		t.Fatal("password was stored in plaintext")
	}
}

func writeTestConfig(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	dataFiles := filepath.Join(root, "world")
	usersDir := filepath.Join(dataFiles, "users")
	if err := os.MkdirAll(usersDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	adminUser := strings.TrimSpace(`
userid: 1
role: admin
username: admin
password: password
character:
  name: AdminAnt
`)
	if err := os.WriteFile(filepath.Join(usersDir, "1.yaml"), []byte(adminUser+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(admin) error = %v", err)
	}

	overridePath := filepath.Join(root, "config-overrides.yaml")
	override := "FilePaths.DataFiles: " + dataFiles + "\n"
	if err := os.WriteFile(overridePath, []byte(override), 0o600); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}

	return overridePath
}

func createUserIndex(t *testing.T) {
	t.Helper()

	index := accounts.InitUserIndex()
	if err := index.Create(); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := index.AddUser(1, "admin"); err != nil {
		t.Fatalf("AddUser(admin) error = %v", err)
	}
}

func writePromptInput(t *testing.T, contents string) *os.File {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "prompt-input")
	if err != nil {
		t.Fatalf("CreateTemp(input) error = %v", err)
	}
	if _, err := file.WriteString(contents); err != nil {
		t.Fatalf("WriteString(input) error = %v", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatalf("Seek(input) error = %v", err)
	}
	return file
}

func tempOutputFile(t *testing.T) *os.File {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "prompt-output")
	if err != nil {
		t.Fatalf("CreateTemp(output) error = %v", err)
	}
	return file
}

func chdirToRepoRoot(t *testing.T) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("Chdir(%q) error = %v", repoRoot, err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})
}
