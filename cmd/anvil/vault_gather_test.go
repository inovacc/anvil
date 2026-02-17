package main

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func TestDiscoverEnvFiles(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("DB_HOST=localhost\nDB_PASS=secret123\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".env.production"), []byte("API_KEY=prod-key\n"), 0o644)

	files, err := discoverSecrets(dir, 5, defaultExcludes)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	total := 0
	for _, f := range files {
		total += len(f.Secrets)
	}

	if total != 3 {
		t.Fatalf("expected 3 total secrets, got %d", total)
	}
}

func TestDiscoverJSONSecrets(t *testing.T) {
	dir := t.TempDir()

	cfg := map[string]any{
		"database": map[string]any{
			"host":     "localhost",
			"password": "dbpass",
		},
		"api_key": "mykey",
		"name":    "app",
	}
	data, _ := json.Marshal(cfg) //nolint:errchkjson // test helper
	_ = os.WriteFile(filepath.Join(dir, "config.json"), data, 0o644)

	files, err := discoverSecrets(dir, 5, defaultExcludes)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	keys := make(map[string]bool)
	for _, s := range files[0].Secrets {
		keys[s.Key] = true
	}

	if !keys["database.password"] {
		t.Error("expected database.password key")
	}

	if !keys["api_key"] {
		t.Error("expected api_key key")
	}

	if keys["name"] {
		t.Error("name should not be a secret")
	}

	if keys["database.host"] {
		t.Error("database.host should not be a secret")
	}
}

func TestDiscoverYAMLSecrets(t *testing.T) {
	dir := t.TempDir()

	yamlContent := `
app:
  name: myapp
  secret_key: supersecret
  database:
    host: localhost
    password: dbpass
`
	_ = os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yamlContent), 0o644)

	files, err := discoverSecrets(dir, 5, defaultExcludes)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	keys := make(map[string]bool)
	for _, s := range files[0].Secrets {
		keys[s.Key] = true
	}

	if !keys["app.secret_key"] {
		t.Error("expected app.secret_key")
	}

	if !keys["app.database.password"] {
		t.Error("expected app.database.password")
	}

	if keys["app.name"] {
		t.Error("app.name should not be a secret")
	}
}

func TestExcludedDirectories(t *testing.T) {
	dir := t.TempDir()

	nodeDir := filepath.Join(dir, "node_modules")
	_ = os.MkdirAll(nodeDir, 0o755)
	_ = os.WriteFile(filepath.Join(nodeDir, ".env"), []byte("SECRET=hidden\n"), 0o644)

	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("VISIBLE=yes\n"), 0o644)

	files, err := discoverSecrets(dir, 5, defaultExcludes)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].Path != filepath.Join(dir, ".env") {
		t.Errorf("expected root .env, got %s", files[0].Path)
	}
}

func TestDepthLimit(t *testing.T) {
	dir := t.TempDir()

	deep := filepath.Join(dir, "a", "b", "c")
	_ = os.MkdirAll(deep, 0o755)
	_ = os.WriteFile(filepath.Join(deep, ".env"), []byte("DEEP=val\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("SHALLOW=val\n"), 0o644)

	files, err := discoverSecrets(dir, 1, defaultExcludes)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file (depth limited), got %d", len(files))
	}
}

func TestKeyDedup(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY=first\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".env.local"), []byte("KEY=second\n"), 0o644)

	files, err := discoverSecrets(dir, 5, defaultExcludes)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]vault.SecretEntry)

	var order []string

	for _, f := range files {
		for _, s := range f.Secrets {
			if _, exists := seen[s.Key]; !exists {
				order = append(order, s.Key)
			}

			seen[s.Key] = s
		}
	}

	entries := make([]vault.SecretEntry, 0, len(order))
	for _, k := range order {
		entries = append(entries, seen[k])
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 deduped entry, got %d", len(entries))
	}
	// Last-file-wins: value should be "second" (walk order is alphabetical).
	if entries[0].Value != "second" {
		t.Errorf("expected last-file-wins value 'second', got %q", entries[0].Value)
	}
}

func TestFlattenMap(t *testing.T) {
	input := map[string]any{
		"a": map[string]any{
			"b": "val1",
			"c": map[string]any{
				"d": "val2",
			},
		},
		"e": "val3",
	}

	out := make(map[string]string)
	flattenMap("", input, out)

	expected := map[string]string{
		"a.b":   "val1",
		"a.c.d": "val2",
		"e":     "val3",
	}

	if len(out) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(out))
	}

	for k, v := range expected {
		if out[k] != v {
			t.Errorf("key %s: expected %q, got %q", k, v, out[k])
		}
	}
}

func TestFilterSecretKeys(t *testing.T) {
	flat := map[string]string{
		"database.host":     "localhost",
		"database.password": "secret",
		"api_key":           "key123",
		"app.name":          "myapp",
		"auth.token":        "tok",
		"credential.file":   "/path",
	}

	entries := filterSecretKeys(flat)

	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, e.Key)
	}

	sort.Strings(keys)

	expected := []string{"api_key", "auth.token", "credential.file", "database.password"}
	sort.Strings(expected)

	if len(keys) != len(expected) {
		t.Fatalf("expected %d secret keys, got %d: %v", len(expected), len(keys), keys)
	}

	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("expected key %q at index %d, got %q", k, i, keys[i])
		}
	}
}

func TestExtraExcludes(t *testing.T) {
	dir := t.TempDir()

	custom := filepath.Join(dir, "mydir")
	_ = os.MkdirAll(custom, 0o755)
	_ = os.WriteFile(filepath.Join(custom, ".env"), []byte("KEY=val\n"), 0o644)

	excludes := make(map[string]bool)
	maps.Copy(excludes, defaultExcludes)

	excludes["mydir"] = true

	files, err := discoverSecrets(dir, 5, excludes)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatalf("expected 0 files with custom exclude, got %d", len(files))
	}
}
