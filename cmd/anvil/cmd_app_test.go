package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppRegisterCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "app", "register", "--name", "myapp", "--description", "test app", "--service-id", "svc-001")
	if !strings.Contains(stdout, "myapp") || !strings.Contains(stdout, "registered") {
		t.Errorf("unexpected output: %q", stdout)
	}
}

func TestAppRegisterJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "app", "register", "--name", "jsonapp", "--json")
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal: %v (output: %q)", err, stdout)
	}

	if result["name"] != "jsonapp" {
		t.Errorf("expected name=jsonapp, got %v", result["name"])
	}
}

func TestAppListCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "listapp1")
	execCmd(t, "app", "register", "--name", "listapp2")

	stdout, _ := execCmd(t, "app", "list")
	if !strings.Contains(stdout, "listapp1") || !strings.Contains(stdout, "listapp2") {
		t.Errorf("expected apps in output: %q", stdout)
	}
}

func TestAppListJSONCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "ljapp")

	stdout, _ := execCmd(t, "app", "list", "--json")
	var result []map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 app, got %d", len(result))
	}
}

func TestAppInfoCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "infoapp", "--description", "info test", "--service-id", "svc-info")

	stdout, _ := execCmd(t, "app", "info", "infoapp")
	if !strings.Contains(stdout, "infoapp") || !strings.Contains(stdout, "svc-info") {
		t.Errorf("unexpected info output: %q", stdout)
	}
}

func TestAppInfoJSONCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "ijapp")

	stdout, _ := execCmd(t, "app", "info", "ijapp", "--json")
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result["name"] != "ijapp" {
		t.Errorf("expected name=ijapp, got %v", result["name"])
	}
}

func TestAppDisableEnableCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "toggleapp")

	stdout, _ := execCmd(t, "app", "disable", "toggleapp")
	if !strings.Contains(stdout, "disabled") {
		t.Errorf("expected 'disabled' in output: %q", stdout)
	}

	stdout, _ = execCmd(t, "app", "enable", "toggleapp")
	if !strings.Contains(stdout, "enabled") {
		t.Errorf("expected 'enabled' in output: %q", stdout)
	}
}

func TestAppRemoveCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "rmapp")

	stdout, _ := execCmd(t, "app", "remove", "rmapp")
	if !strings.Contains(stdout, "removed") {
		t.Errorf("expected 'removed' in output: %q", stdout)
	}
}

func TestAppSecretCRUDCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "crudapp")

	// Set secret.
	stdout, _ := execCmd(t, "app", "set", "crudapp", "DB_HOST", "localhost", "-d", "database host")
	if !strings.Contains(stdout, "DB_HOST") {
		t.Errorf("expected key in output: %q", stdout)
	}

	// Get secret.
	stdout, _ = execCmd(t, "app", "get", "crudapp", "DB_HOST")
	if !strings.Contains(stdout, "localhost") {
		t.Errorf("expected value in output: %q", stdout)
	}

	// List secrets.
	stdout, _ = execCmd(t, "app", "list-secrets", "crudapp")
	if !strings.Contains(stdout, "DB_HOST") {
		t.Errorf("expected key in list: %q", stdout)
	}

	// Delete secret.
	stdout, _ = execCmd(t, "app", "delete", "crudapp", "DB_HOST")
	if !strings.Contains(stdout, "deleted") {
		t.Errorf("expected 'deleted' in output: %q", stdout)
	}
}

func TestAppExportEnvCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "expapp")
	execCmd(t, "app", "set", "expapp", "KEY1", "val1")
	execCmd(t, "app", "set", "expapp", "KEY2", "val2")

	stdout, _ := execCmd(t, "app", "export", "expapp")
	if !strings.Contains(stdout, "KEY1=val1") || !strings.Contains(stdout, "KEY2=val2") {
		t.Errorf("unexpected export output: %q", stdout)
	}
}

func TestAppExportJSONCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "ejapp")
	execCmd(t, "app", "set", "ejapp", "K", "V")

	stdout, _ := execCmd(t, "app", "export", "ejapp", "--format", "json")
	var entries []map[string]any
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("unmarshal: %v (output: %q)", err, stdout)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestAppImportEnvCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "impapp")

	// Write env file.
	envFile := filepath.Join(t.TempDir(), "secrets.env")
	content := "API_KEY=secret123\nDB_PASS=dbpass\n# comment\n\nBAD_LINE\n"
	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "app", "import", "impapp", envFile)
	if !strings.Contains(stdout, "Imported") && !strings.Contains(stdout, "2") {
		t.Errorf("unexpected import output: %q", stdout)
	}

	// Verify imported.
	stdout, _ = execCmd(t, "app", "get", "impapp", "API_KEY")
	if !strings.Contains(stdout, "secret123") {
		t.Errorf("expected imported value: %q", stdout)
	}
}

func TestAppImportJSONCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "app", "register", "--name", "ijmapp")

	// Write JSON file.
	jsonFile := filepath.Join(t.TempDir(), "secrets.json")
	entries := []map[string]string{
		{"key": "K1", "value": "V1"},
		{"key": "K2", "value": "V2"},
	}
	data, _ := json.Marshal(entries)
	if err := os.WriteFile(jsonFile, data, 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "app", "import", "ijmapp", jsonFile, "--format", "json")
	if !strings.Contains(stdout, "Imported") {
		t.Errorf("unexpected import output: %q", stdout)
	}
}
