package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func setupTestVault(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	t.Setenv("ANVIL_DB_PATH", dbPath)

	// Initialize the vault.
	if err := vault.Init(nil); err != nil {
		t.Fatalf("vault.Init: %v", err)
	}

	return dbPath
}

func execCmd(t *testing.T, args ...string) (stdout, stderr string) {
	t.Helper()

	cmd := GetRootCmd()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		handleError(cmd, err)
	}

	return outBuf.String(), errBuf.String()
}

func TestVaultInitCLI(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	t.Setenv("ANVIL_DB_PATH", dbPath)

	stdout, _ := execCmd(t, "vault", "init")
	if !strings.Contains(stdout, "initialized successfully") {
		t.Errorf("unexpected init output: %q", stdout)
	}

	// Double init should report already initialized.
	stdout, _ = execCmd(t, "vault", "init")
	if !strings.Contains(stdout, "already initialized") {
		t.Errorf("expected already initialized message, got: %q", stdout)
	}
}

func TestProfileCRUDCLI(t *testing.T) {
	setupTestVault(t)

	// Create profile.
	stdout, stderr := execCmd(t, "vault", "profile", "create", "dev", "--default", "-d", "development")
	if stderr != "" {
		t.Fatalf("create profile error: %s", stderr)
	}
	if !strings.Contains(stdout, "created") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// List profiles.
	stdout, _ = execCmd(t, "vault", "profile", "list")
	if !strings.Contains(stdout, "dev") {
		t.Errorf("expected 'dev' in list output: %q", stdout)
	}

	// Create another and use it.
	execCmd(t, "vault", "profile", "create", "staging")
	stdout, stderr = execCmd(t, "vault", "profile", "use", "staging")
	if stderr != "" {
		t.Fatalf("use profile error: %s", stderr)
	}
	if !strings.Contains(stdout, "staging") {
		t.Errorf("unexpected use output: %q", stdout)
	}

	// Delete profile.
	stdout, stderr = execCmd(t, "vault", "profile", "delete", "staging")
	if stderr != "" {
		t.Fatalf("delete profile error: %s", stderr)
	}
	if !strings.Contains(stdout, "deleted") {
		t.Errorf("unexpected delete output: %q", stdout)
	}
}

func TestSecretCRUDCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	// Set secret.
	stdout, stderr := execCmd(t, "vault", "set", "API_KEY", "secret123", "-p", "test")
	if stderr != "" {
		t.Fatalf("set error: %s", stderr)
	}
	if !strings.Contains(stdout, "set") {
		t.Errorf("unexpected set output: %q", stdout)
	}

	// Get secret.
	stdout, stderr = execCmd(t, "vault", "get", "API_KEY", "-p", "test")
	if stderr != "" {
		t.Fatalf("get error: %s", stderr)
	}
	if !strings.Contains(stdout, "secret123") {
		t.Errorf("expected secret value, got: %q", stdout)
	}

	// List secrets.
	stdout, _ = execCmd(t, "vault", "list", "-p", "test")
	if !strings.Contains(stdout, "API_KEY") {
		t.Errorf("expected API_KEY in list output: %q", stdout)
	}

	// Delete secret.
	stdout, stderr = execCmd(t, "vault", "delete", "API_KEY", "-p", "test")
	if stderr != "" {
		t.Fatalf("delete error: %s", stderr)
	}
	if !strings.Contains(stdout, "deleted") {
		t.Errorf("unexpected delete output: %q", stdout)
	}
}

func TestVaultStatusCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "vault", "status")
	if stderr != "" {
		t.Fatalf("status error: %s", stderr)
	}
	if !strings.Contains(stdout, "Initialized") || !strings.Contains(stdout, "yes") {
		t.Errorf("unexpected status output: %q", stdout)
	}
}

func TestJSONOutputCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "KEY", "VAL", "-p", "test")

	// Status with --json.
	stdout, _ := execCmd(t, "vault", "status", "--json")
	var status map[string]any
	if err := json.Unmarshal([]byte(stdout), &status); err != nil {
		t.Fatalf("status JSON parse error: %v\noutput: %q", err, stdout)
	}
	if status["initialized"] != true {
		t.Errorf("expected initialized=true, got %v", status["initialized"])
	}

	// Profile list with --json.
	stdout, _ = execCmd(t, "vault", "profile", "list", "--json")
	var profiles []map[string]any
	if err := json.Unmarshal([]byte(stdout), &profiles); err != nil {
		t.Fatalf("profiles JSON parse error: %v\noutput: %q", err, stdout)
	}
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestAuditLogCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "KEY1", "val1", "-p", "test")

	// Audit log should contain the set action.
	stdout, stderr := execCmd(t, "vault", "audit")
	if stderr != "" {
		t.Fatalf("audit error: %s", stderr)
	}
	if !strings.Contains(stdout, "secret.set") {
		t.Errorf("expected audit entry for secret.set, got: %q", stdout)
	}

	// JSON output.
	stdout, _ = execCmd(t, "vault", "audit", "--json")
	var entries []map[string]any
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("audit JSON parse error: %v\noutput: %q", err, stdout)
	}
	if len(entries) == 0 {
		t.Error("expected audit entries")
	}
}

func TestSecretVersioningCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	// Set secret twice to create version history.
	execCmd(t, "vault", "set", "DB_PASS", "v1", "-p", "test")
	execCmd(t, "vault", "set", "DB_PASS", "v2", "-p", "test")

	// Check history.
	stdout, stderr := execCmd(t, "vault", "history", "DB_PASS", "-p", "test")
	if stderr != "" {
		t.Fatalf("history error: %s", stderr)
	}
	if !strings.Contains(stdout, "v1") {
		t.Errorf("expected v1 in history, got: %q", stdout)
	}

	// Rollback.
	stdout, stderr = execCmd(t, "vault", "rollback", "DB_PASS", "1", "-p", "test")
	if stderr != "" {
		t.Fatalf("rollback error: %s", stderr)
	}
	if !strings.Contains(stdout, "rolled back") {
		t.Errorf("expected rollback message, got: %q", stdout)
	}

	// Verify value is v1 again.
	stdout, _ = execCmd(t, "vault", "get", "DB_PASS", "-p", "test")
	if !strings.Contains(stdout, "v1") {
		t.Errorf("expected v1 after rollback, got: %q", stdout)
	}
}

func TestExportImportCLI(t *testing.T) {
	dbPath := setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "A", "1", "-p", "test")
	execCmd(t, "vault", "set", "B", "2", "-p", "test")

	// Export as env format.
	stdout, stderr := execCmd(t, "vault", "export", "-p", "test", "-f", "env")
	if stderr != "" {
		t.Fatalf("export error: %s", stderr)
	}
	if !strings.Contains(stdout, "A=1") || !strings.Contains(stdout, "B=2") {
		t.Errorf("unexpected export output: %q", stdout)
	}

	// Export as JSON.
	stdout, _ = execCmd(t, "vault", "export", "-p", "test", "-f", "json")
	var entries []map[string]any
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("export JSON parse: %v\noutput: %q", err, stdout)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Write env file and import into new profile.
	envFile := filepath.Join(filepath.Dir(dbPath), "import.env")
	if err := os.WriteFile(envFile, []byte("X=10\nY=20\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	execCmd(t, "vault", "profile", "create", "imported")
	stdout, stderr = execCmd(t, "vault", "import", envFile, "-p", "imported", "-f", "env")
	if stderr != "" {
		t.Fatalf("import error: %s", stderr)
	}
	if !strings.Contains(stdout, "Imported 2") {
		t.Errorf("expected import count, got: %q", stdout)
	}

	// Verify imported secrets.
	stdout, _ = execCmd(t, "vault", "get", "X", "-p", "imported")
	if !strings.Contains(stdout, "10") {
		t.Errorf("expected imported value 10, got: %q", stdout)
	}
}

func TestTemplateListCLI(t *testing.T) {
	setupTestVault(t)

	// Built-in templates should be available.
	stdout, stderr := execCmd(t, "vault", "template", "list")
	if stderr != "" {
		t.Fatalf("template list error: %s", stderr)
	}
	if !strings.Contains(stdout, "postgres") {
		t.Errorf("expected postgres template, got: %q", stdout)
	}

	// JSON output.
	stdout, _ = execCmd(t, "vault", "template", "list", "--json")
	var templates []map[string]any
	if err := json.Unmarshal([]byte(stdout), &templates); err != nil {
		t.Fatalf("template list JSON parse: %v\noutput: %q", err, stdout)
	}
	if len(templates) < 5 {
		t.Errorf("expected at least 5 built-in templates, got %d", len(templates))
	}
}

func TestTemplateShowCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "vault", "template", "show", "postgres")
	if stderr != "" {
		t.Fatalf("template show error: %s", stderr)
	}
	if !strings.Contains(stdout, "postgres") {
		t.Errorf("expected template details, got: %q", stdout)
	}
}

func TestTemplateApplyCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	stdout, stderr := execCmd(t, "vault", "template", "apply", "redis",
		"-p", "test", "--set", "host=localhost", "--set", "port=6379", "--set", "password=secret")
	if stderr != "" {
		t.Fatalf("template apply error: %s", stderr)
	}
	if !strings.Contains(stdout, "applied") {
		t.Errorf("expected applied message, got: %q", stdout)
	}
}

func TestPluginListCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "vault", "plugin", "list")
	if stderr != "" {
		t.Fatalf("plugin list error: %s", stderr)
	}
	// Output may be JSON or text depending on persistent flag state.
	if !strings.Contains(stdout, "Hooks") && !strings.Contains(stdout, "hooks") {
		t.Errorf("expected hooks section, got: %q", stdout)
	}
}

func TestPluginHookAddRemoveCLI(t *testing.T) {
	setupTestVault(t)

	var hookCmd string
	if runtime.GOOS == "windows" {
		hookCmd = "cmd.exe"
	} else {
		hookCmd = "/bin/echo"
	}

	// Add hook.
	stdout, stderr := execCmd(t, "vault", "plugin", "hook-add", "post-set", hookCmd)
	if stderr != "" {
		t.Fatalf("hook-add error: %s", stderr)
	}
	if !strings.Contains(stdout, "Hook added") && !strings.Contains(stdout, "ok") {
		t.Errorf("expected success output, got: %q", stdout)
	}

	// List should show the hook.
	stdout, _ = execCmd(t, "vault", "plugin", "list")
	if !strings.Contains(stdout, "post-set") {
		t.Errorf("expected hook in list, got: %q", stdout)
	}

	// Remove hook.
	stdout, stderr = execCmd(t, "vault", "plugin", "hook-remove", "post-set", hookCmd)
	if stderr != "" {
		t.Fatalf("hook-remove error: %s", stderr)
	}
	if !strings.Contains(stdout, "Hook removed") && !strings.Contains(stdout, "ok") {
		t.Errorf("expected success output, got: %q", stdout)
	}
}

func TestPluginProviderAddRemoveCLI(t *testing.T) {
	setupTestVault(t)

	// Add provider.
	stdout, stderr := execCmd(t, "vault", "plugin", "provider-add", "aws", "/usr/bin/aws-provider", "aws/")
	if stderr != "" {
		t.Fatalf("provider-add error: %s", stderr)
	}
	if !strings.Contains(stdout, "Provider added") && !strings.Contains(stdout, "ok") {
		t.Errorf("expected success output, got: %q", stdout)
	}

	// Remove provider.
	stdout, stderr = execCmd(t, "vault", "plugin", "provider-remove", "aws")
	if stderr != "" {
		t.Fatalf("provider-remove error: %s", stderr)
	}
	if !strings.Contains(stdout, "Provider removed") && !strings.Contains(stdout, "ok") {
		t.Errorf("expected success output, got: %q", stdout)
	}
}

func TestErrorHandlingCLI(t *testing.T) {
	setupTestVault(t)

	// Get non-existent secret should produce error.
	_, stderr := execCmd(t, "vault", "get", "NONEXISTENT", "-p", "default")
	if stderr == "" {
		t.Error("expected error for non-existent secret")
	}

	// JSON error output.
	_, stderr = execCmd(t, "vault", "get", "NONEXISTENT", "-p", "default", "--json")
	if stderr == "" {
		t.Error("expected JSON error for non-existent secret")
	}
	var errResp map[string]any
	if err := json.Unmarshal([]byte(stderr), &errResp); err != nil {
		t.Fatalf("error JSON parse: %v\noutput: %q", err, stderr)
	}
	if _, ok := errResp["error"]; !ok {
		t.Errorf("expected 'error' field in JSON error: %v", errResp)
	}
}

func TestPluginIntegrationE2E(t *testing.T) {
	dbPath := setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	dir := filepath.Dir(dbPath)
	markerFile := filepath.Join(dir, "hook-fired.txt")

	// Create a hook script that writes a marker file.
	var hookScript, hookCmd string
	if runtime.GOOS == "windows" {
		hookScript = filepath.Join(dir, "hook.bat")
		hookCmd = hookScript
		if err := os.WriteFile(hookScript, []byte(
			"@echo off\r\necho {\"allow\":true}\r\necho hook-fired > \""+markerFile+"\"\r\n",
		), 0o700); err != nil {
			t.Fatalf("write hook script: %v", err)
		}
	} else {
		hookScript = filepath.Join(dir, "hook.sh")
		hookCmd = hookScript
		if err := os.WriteFile(hookScript, []byte(
			"#!/bin/sh\necho '{\"allow\":true}'\necho hook-fired > '"+markerFile+"'\n",
		), 0o700); err != nil {
			t.Fatalf("write hook script: %v", err)
		}
	}

	// Add the hook.
	execCmd(t, "vault", "plugin", "hook-add", "post-set", hookCmd)

	// Set a secret — should trigger the hook.
	stdout, stderr := execCmd(t, "vault", "set", "TRIGGER_KEY", "value", "-p", "test")
	if stderr != "" {
		t.Fatalf("set error: %s", stderr)
	}
	if !strings.Contains(stdout, "set") {
		t.Errorf("unexpected set output: %q", stdout)
	}

	// Verify the hook fired by checking the marker file.
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("hook did not fire — marker file not created")
	}
}
