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
	// Reset persistent flags to avoid leakage between tests.
	_ = cmd.PersistentFlags().Set("json", "false")
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

func TestDockerExportCleanCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "DB_HOST", "localhost", "-p", "test")
	execCmd(t, "vault", "set", "DB_PASS", "secret", "-p", "test")

	dir := filepath.Join(t.TempDir(), "docker-secrets")

	// Export.
	stdout, stderr := execCmd(t, "vault", "docker", "export", dir, "-p", "test")
	if stderr != "" {
		t.Fatalf("docker export error: %s", stderr)
	}
	if !strings.Contains(stdout, "Exported 2") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Verify files exist.
	data, err := os.ReadFile(filepath.Join(dir, "DB_HOST"))
	if err != nil {
		t.Fatalf("read secret file: %v", err)
	}
	if string(data) != "localhost" {
		t.Errorf("expected 'localhost', got %q", string(data))
	}

	// Clean.
	stdout, stderr = execCmd(t, "vault", "docker", "clean", dir, "-p", "test")
	if stderr != "" {
		t.Fatalf("docker clean error: %s", stderr)
	}
	if !strings.Contains(stdout, "Removed 2") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Verify files removed.
	if _, err := os.Stat(filepath.Join(dir, "DB_HOST")); !os.IsNotExist(err) {
		t.Error("expected secret file to be removed")
	}
}

func TestDockerComposeCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "API_KEY", "key123", "-p", "test")

	stdout, stderr := execCmd(t, "vault", "docker", "compose", "-p", "test")
	if stderr != "" {
		t.Fatalf("docker compose error: %s", stderr)
	}
	if !strings.Contains(stdout, "secrets:") || !strings.Contains(stdout, "API_KEY") {
		t.Errorf("unexpected compose output: %q", stdout)
	}
}

func TestShareExportImportCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "SECRET", "val123", "-p", "test")

	shareFile := filepath.Join(t.TempDir(), "shared.enc")

	// Export.
	stdout, stderr := execCmd(t, "vault", "share", "export", shareFile, "--profile", "test", "--passphrase", "testpass123")
	if stderr != "" {
		t.Fatalf("share export error: %s", stderr)
	}
	if !strings.Contains(stdout, "Shared export written") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Create target profile and import.
	execCmd(t, "vault", "profile", "create", "imported")
	stdout, stderr = execCmd(t, "vault", "share", "import", shareFile, "--profile", "imported", "--passphrase", "testpass123")
	if stderr != "" {
		t.Fatalf("share import error: %s", stderr)
	}
	if !strings.Contains(stdout, "Imported") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Verify imported.
	stdout, _ = execCmd(t, "vault", "get", "SECRET", "-p", "imported")
	if !strings.Contains(stdout, "val123") {
		t.Errorf("expected imported secret, got: %q", stdout)
	}
}

func TestRotateKeyCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "KEY1", "val1", "-p", "test")

	// Set password first (required for rotate-key).
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")

	// Rotate.
	stdout, stderr := execCmd(t, "vault", "rotate-key", "--password", "mypassword1")
	if stderr != "" {
		t.Fatalf("rotate-key error: %s", stderr)
	}
	if !strings.Contains(stdout, "rotated successfully") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Verify secret still readable.
	stdout, _ = execCmd(t, "vault", "get", "KEY1", "-p", "test")
	if !strings.Contains(stdout, "val1") {
		t.Errorf("expected secret after rotation, got: %q", stdout)
	}
}

func TestBackupRestoreCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "BK_KEY", "bk_val", "-p", "test")

	// Set password (required for backup).
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")

	backupFile := filepath.Join(t.TempDir(), "backup.enc")

	// Backup.
	stdout, stderr := execCmd(t, "vault", "backup", backupFile, "--password", "mypassword1")
	if stderr != "" {
		t.Fatalf("backup error: %s", stderr)
	}
	if !strings.Contains(stdout, "Backup written") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Create fresh vault for restore.
	dbPath2 := filepath.Join(t.TempDir(), "vault2.db")
	t.Setenv("ANVIL_DB_PATH", dbPath2)
	if err := vault.Init(nil); err != nil {
		t.Fatalf("vault.Init for restore: %v", err)
	}

	// Restore.
	stdout, stderr = execCmd(t, "vault", "restore", backupFile, "--password", "mypassword1")
	if stderr != "" {
		t.Fatalf("restore error: %s", stderr)
	}
	if !strings.Contains(stdout, "restored") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Verify secret is accessible.
	stdout, _ = execCmd(t, "vault", "get", "BK_KEY", "-p", "test")
	if !strings.Contains(stdout, "bk_val") {
		t.Errorf("expected restored secret, got: %q", stdout)
	}
}

func TestEnvPasswordSetCLI(t *testing.T) {
	setupTestVault(t)

	// Set password via non-interactive flag.
	stdout, stderr := execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")
	if stderr != "" {
		t.Fatalf("password set error: %s", stderr)
	}
	if !strings.Contains(stdout, "Password set") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Update password.
	stdout, stderr = execCmd(t, "vault", "env", "password", "set", "--password", "newpassword1", "--current", "mypassword1")
	if stderr != "" {
		t.Fatalf("password update error: %s", stderr)
	}
	if !strings.Contains(stdout, "Password set") {
		t.Errorf("unexpected output: %q", stdout)
	}
}

func TestEnvReleaseStatusRevokeCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "K1", "V1", "-p", "test")
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")

	// Clean up any stale sentinel from other tests.
	execCmd(t, "vault", "env", "revoke")

	// Status before release.
	stdout, _ := execCmd(t, "vault", "env", "status")
	if !strings.Contains(stdout, "inactive") {
		t.Errorf("expected inactive status, got: %q", stdout)
	}

	// Release.
	stdout, stderr := execCmd(t, "vault", "env", "release", "-p", "test", "--password", "mypassword1", "--ttl", "5m")
	if stderr != "" {
		t.Fatalf("release error: %s", stderr)
	}
	if !strings.Contains(stdout, "released") {
		t.Errorf("unexpected release output: %q", stdout)
	}

	// Status after release.
	stdout, _ = execCmd(t, "vault", "env", "status")
	if !strings.Contains(stdout, "active") {
		t.Errorf("expected active status, got: %q", stdout)
	}

	// Env export.
	stdout, stderr = execCmd(t, "vault", "env", "export", "-p", "test", "-f", "env")
	if stderr != "" {
		t.Fatalf("env export error: %s", stderr)
	}
	if !strings.Contains(stdout, "K1=V1") {
		t.Errorf("expected K1=V1 in env export, got: %q", stdout)
	}

	// Env export as shell export format.
	stdout, _ = execCmd(t, "vault", "env", "export", "-p", "test", "-f", "export")
	if !strings.Contains(stdout, "export K1=") {
		t.Errorf("expected export format, got: %q", stdout)
	}

	// Revoke.
	stdout, stderr = execCmd(t, "vault", "env", "revoke")
	if stderr != "" {
		t.Fatalf("revoke error: %s", stderr)
	}
	if !strings.Contains(stdout, "revoked") {
		t.Errorf("unexpected revoke output: %q", stdout)
	}

	// Status after revoke.
	stdout, _ = execCmd(t, "vault", "env", "status")
	if !strings.Contains(stdout, "inactive") {
		t.Errorf("expected inactive after revoke, got: %q", stdout)
	}
}

func TestCmdTreeCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "cmdtree")
	if stderr != "" {
		t.Fatalf("cmdtree error: %s", stderr)
	}
	if !strings.Contains(stdout, "vault") || !strings.Contains(stdout, "anvil") {
		t.Errorf("expected command tree, got: %q", stdout)
	}
}

func TestCmdTreeJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "cmdtree", "--json")
	var entries []cmdTreeEntry
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("cmdtree JSON parse: %v\noutput: %q", err, stdout)
	}
	if len(entries) < 10 {
		t.Errorf("expected many commands, got %d", len(entries))
	}
}

func TestTemplateCreateDeleteCLI(t *testing.T) {
	setupTestVault(t)

	// Create a template file.
	tmplFile := filepath.Join(t.TempDir(), "tmpl.yaml")
	tmplContent := `name: mytest
description: test template
secrets:
  - key: "{{.prefix}}_HOST"
    value: "{{.host}}"
  - key: "{{.prefix}}_PORT"
    value: "{{.port}}"
variables:
  - name: prefix
    description: Key prefix
  - name: host
    description: Hostname
  - name: port
    description: Port number
`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	stdout, stderr := execCmd(t, "vault", "template", "create", tmplFile)
	if stderr != "" {
		t.Fatalf("template create error: %s", stderr)
	}
	if !strings.Contains(stdout, "created") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Show.
	stdout, stderr = execCmd(t, "vault", "template", "show", "mytest")
	if stderr != "" {
		t.Fatalf("template show error: %s", stderr)
	}
	if !strings.Contains(stdout, "mytest") {
		t.Errorf("unexpected show output: %q", stdout)
	}

	// Delete.
	stdout, stderr = execCmd(t, "vault", "template", "delete", "mytest")
	if stderr != "" {
		t.Fatalf("template delete error: %s", stderr)
	}
	if !strings.Contains(stdout, "deleted") {
		t.Errorf("unexpected delete output: %q", stdout)
	}
}

func TestDockerExportJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "KEY", "val", "-p", "test")

	dir := filepath.Join(t.TempDir(), "dsecrets")
	stdout, _ := execCmd(t, "vault", "docker", "export", dir, "-p", "test", "--json")
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}
	if result["count"].(float64) != 1 {
		t.Errorf("expected count 1, got %v", result["count"])
	}
}

func TestEnvStatusJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "vault", "env", "status", "--json")
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}
	if result["active"] != false {
		t.Errorf("expected active=false, got %v", result["active"])
	}
}

func TestGatherCLI(t *testing.T) {
	setupTestVault(t)

	// Create a temp project directory with secret files.
	projDir := t.TempDir()

	// .env file
	if err := os.WriteFile(filepath.Join(projDir, ".env"), []byte("DB_HOST=localhost\nDB_PASS=secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// JSON config with secrets
	if err := os.WriteFile(filepath.Join(projDir, "config.json"), []byte(`{"api_key":"abc123","name":"app"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Gather with --yes and --profile.
	stdout, stderr := execCmd(t, "vault", "gather", projDir, "--yes", "--profile", "gathered")
	if stderr != "" {
		t.Fatalf("gather error: %s", stderr)
	}
	if !strings.Contains(stdout, "gathered") {
		t.Errorf("unexpected output: %q", stdout)
	}

	// Verify secrets were imported.
	stdout, _ = execCmd(t, "vault", "get", "DB_HOST", "-p", "gathered")
	if !strings.Contains(stdout, "localhost") {
		t.Errorf("expected DB_HOST=localhost, got: %q", stdout)
	}

	stdout, _ = execCmd(t, "vault", "get", "DB_PASS", "-p", "gathered")
	if !strings.Contains(stdout, "secret") {
		t.Errorf("expected DB_PASS=secret, got: %q", stdout)
	}
}

func TestGatherJSONCLI(t *testing.T) {
	setupTestVault(t)

	projDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projDir, ".env"), []byte("TOKEN=xyz\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := execCmd(t, "vault", "gather", projDir, "--yes", "--profile", "jtest", "--json")
	if stderr != "" {
		t.Fatalf("gather JSON error: %s", stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}
	if result["profile"] != "jtest" {
		t.Errorf("expected profile jtest, got %v", result["profile"])
	}
	if result["count"].(float64) != 1 {
		t.Errorf("expected count 1, got %v", result["count"])
	}
}

func TestGatherNoFilesCLI(t *testing.T) {
	setupTestVault(t)

	emptyDir := t.TempDir()
	stdout, _ := execCmd(t, "vault", "gather", emptyDir, "--yes", "--profile", "empty")
	if !strings.Contains(stdout, "No secret files found") {
		t.Errorf("expected no files message, got: %q", stdout)
	}
}

func TestGatherExcludesCLI(t *testing.T) {
	setupTestVault(t)

	projDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projDir, ".env"), []byte("ROOT_KEY=val\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(projDir, "secrets")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, ".env"), []byte("SUB_KEY=val\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Gather excluding the "secrets" subdirectory.
	stdout, stderr := execCmd(t, "vault", "gather", projDir, "--yes", "--profile", "excl", "--exclude", "secrets")
	if stderr != "" {
		t.Fatalf("gather error: %s", stderr)
	}

	// Should only have 1 secret from root .env.
	stdout, _ = execCmd(t, "vault", "get", "ROOT_KEY", "-p", "excl")
	if !strings.Contains(stdout, "val") {
		t.Errorf("expected ROOT_KEY, got: %q", stdout)
	}

	// SUB_KEY should not exist.
	_, stderr = execCmd(t, "vault", "get", "SUB_KEY", "-p", "excl")
	if stderr == "" {
		t.Error("expected error for excluded SUB_KEY")
	}
}

func TestSealUnsealCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "seal-test", "--default")
	execCmd(t, "vault", "set", "KEY1", "value1", "-p", "seal-test")

	// Seal the vault.
	stdout, _ := execCmd(t, "vault", "seal")
	if !strings.Contains(stdout, "sealed") && !strings.Contains(stdout, "Sealed") {
		t.Errorf("unexpected seal output: %q", stdout)
	}

	// Operations should fail while sealed.
	_, stderr := execCmd(t, "vault", "get", "KEY1", "-p", "seal-test")
	if !strings.Contains(stderr, "sealed") {
		t.Errorf("expected sealed error, got stderr: %q", stderr)
	}

	// Unseal.
	stdout, _ = execCmd(t, "vault", "unseal")
	if !strings.Contains(stdout, "unsealed") && !strings.Contains(stdout, "Unsealed") {
		t.Errorf("unexpected unseal output: %q", stdout)
	}

	// Operations should work again.
	stdout, stderr = execCmd(t, "vault", "get", "KEY1", "-p", "seal-test")
	if stderr != "" {
		t.Errorf("get after unseal error: %s", stderr)
	}
	if !strings.Contains(stdout, "value1") {
		t.Errorf("expected value1, got: %q", stdout)
	}
}

func TestSealUnsealJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "vault", "seal", "--json")
	var result struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v (output: %q)", err, stdout)
	}
	if !strings.Contains(strings.ToLower(result.Message), "sealed") {
		t.Errorf("unexpected message: %q", result.Message)
	}

	stdout, _ = execCmd(t, "vault", "unseal", "--json")
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v (output: %q)", err, stdout)
	}
	if !strings.Contains(strings.ToLower(result.Message), "unsealed") {
		t.Errorf("unexpected message: %q", result.Message)
	}
}

func TestUnsealNotSealedCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "unseal")
	if !strings.Contains(stderr, "not sealed") {
		t.Errorf("expected 'not sealed' error, got: %q", stderr)
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
