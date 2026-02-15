package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
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
