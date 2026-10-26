package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const sampleHTTPSConfig = `FilePaths:
  DataFiles: _datafiles/world/empty
  WebDomain: "localhost"
  HttpsCertFile: "server.crt"
  HttpsKeyFile: "server.key"
  HttpsEmail: ""
Network:
  HttpPort: 8080
  HttpsPort: 8443
  HttpsRedirect: false
`

func TestHTTPSSetupManualModePrintsOverrideSnippet(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "Override target: _datafiles/world/empty/config-overrides.yaml") {
		t.Fatalf("https-setup output did not show override target:\n%s", output)
	}
	if !strings.Contains(output, "PATCH a running GoMud server via /admin/api/v1/config (recommended)") {
		t.Fatalf("https-setup output did not mark PATCH as recommended:\n%s", output)
	}
	if strings.Contains(output, "This helper no longer edits the bundled base config directly.") {
		t.Fatalf("https-setup output still showed removed intro text:\n%s", output)
	}
	if !strings.Contains(output, "Save the following override snippet") {
		t.Fatalf("https-setup output did not switch to snippet mode:\n%s", output)
	}
	if !strings.Contains(output, "HttpsCertFile: 'server.crt'") {
		t.Fatalf("https-setup output did not include cert override:\n%s", output)
	}

	updated := readHTTPSSetupConfig(t, configPath)
	if updated != sampleHTTPSConfig {
		t.Fatalf("https-setup changed bundled config unexpectedly:\n%s", updated)
	}
}

func TestHTTPSSetupHTTPOnlySnippetClearsHTTPS(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"3",
		"8080",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "HttpsCertFile: ''") {
		t.Fatalf("https-setup output did not clear cert override:\n%s", output)
	}
	if !strings.Contains(output, "HttpsPort: 0") {
		t.Fatalf("https-setup output did not disable HTTPS in snippet:\n%s", output)
	}
}

func TestHTTPSSetupUsesConfigPathAsOverrideTarget(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	overridePath := filepath.Join(t.TempDir(), "custom-overrides.yaml")

	input := strings.Join([]string{
		"3",
		"8080",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CONFIG_PATH": overridePath,
	})
	if !strings.Contains(output, "Override target: "+overridePath) {
		t.Fatalf("https-setup output did not prefer CONFIG_PATH:\n%s", output)
	}
}

func TestHTTPSSetupUsesExistingOverrideValuesAsDefaults(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	overridePath := filepath.Join(t.TempDir(), "custom-overrides.yaml")
	overrideConfig := `FilePaths:
  HttpsCertFile: 'existing.crt'
  HttpsKeyFile: 'existing.key'
Network:
  HttpPort: 9080
  HttpsPort: 9443
  HttpsRedirect: true
`
	if err := os.WriteFile(overridePath, []byte(overrideConfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CONFIG_PATH": overridePath,
	})
	if !strings.Contains(output, "HttpsCertFile: existing.crt") {
		t.Fatalf("https-setup output did not preserve override cert default:\n%s", output)
	}
	if !strings.Contains(output, "HttpPort: 9080") {
		t.Fatalf("https-setup output did not preserve override HTTP port default:\n%s", output)
	}
	if !strings.Contains(output, "HttpsPort: 9443") {
		t.Fatalf("https-setup output did not preserve override HTTPS port default:\n%s", output)
	}
	if !strings.Contains(output, "HttpsRedirect: true") {
		t.Fatalf("https-setup output did not preserve override redirect default:\n%s", output)
	}
}

func TestHTTPSSetupSnippetKeepsWindowsPathsYAMLSafe(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"1",
		`C:\certs\server.crt`,
		`C:\certs\server.key`,
		"",
		"",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "HttpsCertFile: 'C:\\certs\\server.crt'") {
		t.Fatalf("https-setup output did not emit YAML-safe Windows cert path:\n%s", output)
	}
	if !strings.Contains(output, "HttpsKeyFile: 'C:\\certs\\server.key'") {
		t.Fatalf("https-setup output did not emit YAML-safe Windows key path:\n%s", output)
	}
}

func TestHTTPSSetupIgnoresConfigPathWhenItTargetsBundledConfig(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"3",
		"8080",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CONFIG_PATH": configPath,
	})
	if !strings.Contains(output, "Ignoring CONFIG_PATH="+configPath) {
		t.Fatalf("https-setup output did not warn about bundled config target:\n%s", output)
	}
	if !strings.Contains(output, "Override target: _datafiles/world/empty/config-overrides.yaml") {
		t.Fatalf("https-setup output did not fall back to config-overrides target:\n%s", output)
	}
}

func TestHTTPSSetupCanPatchRunningServer(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	curlDir := t.TempDir()
	curlLog := filepath.Join(curlDir, "curl.log")
	curlStub := filepath.Join(curlDir, "curl-stub.sh")
	curlScript := "#!/usr/bin/env sh\nprintf '%s\\n' \"$@\" >\"" + curlLog + "\"\nprintf '200'\n"
	if err := os.WriteFile(curlStub, []byte(curlScript), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"1",
		"http://localhost/",
		"admin",
		"password",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CURL_BIN": curlStub,
	})
	if !strings.Contains(output, "HTTPS setup applied through the admin API.") {
		t.Fatalf("https-setup output did not confirm API patch:\n%s", output)
	}
	if !strings.Contains(output, "Restart GoMud so it rebinds the updated HTTP/HTTPS listeners.") {
		t.Fatalf("https-setup output did not require restart after API patch:\n%s", output)
	}

	logBytes, err := os.ReadFile(curlLog)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	curlArgs := string(logBytes)
	if !strings.Contains(curlArgs, "http://localhost/admin/api/v1/config") {
		t.Fatalf("curl stub did not receive config endpoint:\n%s", curlArgs)
	}
	if !strings.Contains(curlArgs, "FilePaths.HttpsCertFile") {
		t.Fatalf("curl stub did not receive config payload:\n%s", curlArgs)
	}

	updated := readHTTPSSetupConfig(t, configPath)
	if updated != sampleHTTPSConfig {
		t.Fatalf("https-setup changed bundled config unexpectedly:\n%s", updated)
	}
}

func TestHTTPSSetupAutoModeAPIApplyRequiresRestart(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	curlDir := t.TempDir()
	curlLog := filepath.Join(curlDir, "curl.log")
	curlStub := filepath.Join(curlDir, "curl-stub.sh")
	curlScript := "#!/usr/bin/env sh\nprintf '%s\\n' \"$@\" >\"" + curlLog + "\"\nprintf '200'\n"
	if err := os.WriteFile(curlStub, []byte(curlScript), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"2",
		"play.example.com",
		"ops@example.com",
		"1",
		"http://localhost/",
		"admin",
		"password",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CURL_BIN": curlStub,
	})
	if strings.Contains(output, "Admin base URL [http://localhost]") {
		t.Fatalf("https-setup auto API output still showed localhost helper text:\n%s", output)
	}
	if !strings.Contains(output, "Admin base URL [http://127.0.0.1:8080]") {
		t.Fatalf("https-setup auto API output did not show loopback helper text based on current HTTP port:\n%s", output)
	}
	if !strings.Contains(output, "Restart GoMud so it rebinds the updated HTTP/HTTPS listeners.") {
		t.Fatalf("https-setup auto API output did not require restart:\n%s", output)
	}
	if !strings.Contains(output, "Review /admin/https/ if certificate issuance needs troubleshooting.") {
		t.Fatalf("https-setup auto API output did not keep HTTPS troubleshooting guidance:\n%s", output)
	}
}

func TestHTTPSSetupAutoModeLoadsDefaultAdminURLFromActiveOverride(t *testing.T) {
	overrideDir := t.TempDir()
	configPath := writeHTTPSSetupTempConfig(t, "FilePaths:\n  DataFiles: "+overrideDir+"\nNetwork:\n  HttpPort: 8080\n")
	overridePath := filepath.Join(overrideDir, "config-overrides.yaml")
	overrideConfig := `FilePaths:
  WebDomain: "play.example.com"
Network:
  HttpPort: 9090
`
	if err := os.WriteFile(overridePath, []byte(overrideConfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	curlDir := t.TempDir()
	curlStub := filepath.Join(curlDir, "curl-stub.sh")
	curlScript := "#!/usr/bin/env sh\nprintf '200'\n"
	if err := os.WriteFile(curlStub, []byte(curlScript), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"2",
		"",
		"",
		"1",
		"",
		"",
		"password",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CURL_BIN": curlStub,
	})
	if !strings.Contains(output, "Override target: "+overridePath) {
		t.Fatalf("https-setup output did not show the active override target:\n%s", output)
	}
	if !strings.Contains(output, "Admin base URL [http://127.0.0.1:9090]") {
		t.Fatalf("https-setup output did not derive the admin URL from the active override:\n%s", output)
	}
	if !strings.Contains(output, "WebDomain: play.example.com") {
		t.Fatalf("https-setup output did not load hostname from the active override:\n%s", output)
	}
}

func TestHTTPSSetupAPIApplyFailureExplainsRetryAndFallback(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	curlDir := t.TempDir()
	curlStub := filepath.Join(curlDir, "curl-stub.sh")
	curlScript := "#!/usr/bin/env sh\nexit 7\n"
	if err := os.WriteFile(curlStub, []byte(curlScript), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"2",
		"play.example.com",
		"",
		"1",
		"",
		"",
		"password",
		"",
	}, "\n")

	output, err := runHTTPSSetupExpectError(t, configPath, input, map[string]string{
		"CURL_BIN": curlStub,
	})
	if err == nil {
		t.Fatalf("https-setup.sh unexpectedly succeeded for failed API apply")
	}
	if !strings.Contains(output, "Failed to apply settings through the admin API.") {
		t.Fatalf("https-setup output did not report API apply failure:\n%s", output)
	}
	if !strings.Contains(output, "GoMud is not reachable at http://127.0.0.1:8080.") {
		t.Fatalf("https-setup output did not identify unreachable admin URL:\n%s", output)
	}
	if !strings.Contains(output, "If the server is already running, enter its current admin URL and try again.") {
		t.Fatalf("https-setup output did not explain retry guidance:\n%s", output)
	}
	if !strings.Contains(output, "Otherwise, save the override snippet below and restart GoMud.") {
		t.Fatalf("https-setup output did not explain fallback guidance:\n%s", output)
	}
}

func TestHTTPSSetupAutoModePrintsLetsEncryptOverrideSnippet(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"2",
		"play.example.com",
		"ops@example.com",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "WebDomain: play.example.com") {
		t.Fatalf("https-setup output did not show selected hostname:\n%s", output)
	}
	if !strings.Contains(output, "HttpsEmail: ops@example.com") {
		t.Fatalf("https-setup output did not show selected email:\n%s", output)
	}
	if !strings.Contains(output, "WebDomain: \"play.example.com\"") {
		t.Fatalf("https-setup output did not include WebDomain override:\n%s", output)
	}
	if !strings.Contains(output, "HttpPort: 80") || !strings.Contains(output, "HttpsPort: 443") {
		t.Fatalf("https-setup output did not set standard ports for auto mode:\n%s", output)
	}
	if !strings.Contains(output, "HttpsRedirect: true") {
		t.Fatalf("https-setup output did not enable redirect for auto mode:\n%s", output)
	}
	if !strings.Contains(output, "HttpsCertFile: <empty>") || !strings.Contains(output, "HttpsKeyFile: <empty>") {
		t.Fatalf("https-setup output did not automatically blank manual cert paths for auto mode:\n%s", output)
	}
}

func TestHTTPSSetupAutoModeRejectsLocalhost(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"2",
		"localhost",
		"",
	}, "\n")

	output, err := runHTTPSSetupExpectError(t, configPath, input, nil)
	if err == nil {
		t.Fatalf("https-setup.sh unexpectedly succeeded for localhost")
	}
	if !strings.Contains(output, "Automatic HTTPS requires a public hostname") {
		t.Fatalf("https-setup.sh output did not explain invalid hostname:\n%s", output)
	}

	updated := readHTTPSSetupConfig(t, configPath)
	if updated != sampleHTTPSConfig {
		t.Fatalf("https-setup config changed after invalid auto mode input:\n%s", updated)
	}
}

func TestHTTPSSetupAutoModeNormalizesPastedHostname(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"2",
		"https://Play.Example.com/",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "WebDomain: \"play.example.com\"") {
		t.Fatalf("https-setup output did not normalize pasted hostname:\n%s", output)
	}
}

func TestHTTPSSetupPreservesBundledConfigPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not stable on Windows")
	}

	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	if err := os.Chmod(configPath, 0o640); err != nil {
		t.Fatalf("os.Chmod() error = %v", err)
	}

	input := strings.Join([]string{
		"3",
		"8080",
		"2",
		"",
	}, "\n")

	runHTTPSSetup(t, configPath, input, nil)

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("config mode = %o, want %o", got, want)
	}
}

func writeHTTPSSetupTempConfig(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func readHTTPSSetupConfig(t *testing.T, configPath string) string {
	t.Helper()

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	return string(data)
}

func runHTTPSSetup(t *testing.T, configPath string, input string, extraEnv map[string]string) string {
	t.Helper()

	output, err := runHTTPSSetupCommand(t, configPath, input, extraEnv)
	if err != nil {
		t.Fatalf("https-setup.sh failed: %v\n%s", err, output)
	}

	return output
}

func runHTTPSSetupExpectError(t *testing.T, configPath string, input string, extraEnv map[string]string) (string, error) {
	t.Helper()

	return runHTTPSSetupCommand(t, configPath, input, extraEnv)
}

func runHTTPSSetupCommand(t *testing.T, configPath string, input string, extraEnv map[string]string) (string, error) {
	t.Helper()

	scriptDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	tempDir := t.TempDir()
	cmd := exec.Command("sh", "./https-setup.sh")
	cmd.Dir = scriptDir
	cmd.Env = append(os.Environ(),
		"CONFIG_FILE="+configPath,
		"TMPDIR="+tempDir,
	)
	for key, value := range extraEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	return string(output), err
}
