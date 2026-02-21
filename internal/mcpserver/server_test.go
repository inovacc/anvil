package mcpserver_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/inovacc/anvil/internal/mcpserver"
	"github.com/inovacc/anvil/pkg/vault"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func setupTestServer(t *testing.T) *mcp.ClientSession {
	t.Helper()
	t.Setenv("ANVIL_SKIP_TPM", "1")

	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := vault.Init(&vault.Options{DBPath: dbPath}); err != nil {
		t.Fatalf("vault.Init: %v", err)
	}

	srv, err := mcpserver.New(&vault.Options{DBPath: dbPath})
	if err != nil {
		t.Fatalf("mcpserver.New: %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })

	st, ct := mcp.NewInMemoryTransports()
	ctx := context.Background()

	if _, err := srv.MCP.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server.Connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	return cs
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return result
}

func resultJSON(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("empty content")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(tc.Text), &m); err != nil {
		t.Fatalf("unmarshal: %v (text: %s)", err, tc.Text)
	}
	return m
}

func TestListTools(t *testing.T) {
	cs := setupTestServer(t)

	tools, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	expected := []string{
		"secret_get", "secret_set", "secret_delete", "secret_list",
		"profile_list", "profile_create", "profile_delete",
		"key_generate", "key_list", "key_delete", "key_export",
		"sign", "verify",
		"audit_log", "vault_status", "installation_id",
	}

	found := make(map[string]bool)
	for _, tool := range tools.Tools {
		found[tool.Name] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestSecretCRUD(t *testing.T) {
	cs := setupTestServer(t)

	// Create profile first.
	result := callTool(t, cs, "profile_create", map[string]any{
		"name": "test", "is_default": true,
	})
	if result.IsError {
		t.Fatalf("profile_create error: %v", result.Content)
	}

	// Set secret.
	result = callTool(t, cs, "secret_set", map[string]any{
		"key": "API_KEY", "value": "secret123", "profile": "test",
	})
	m := resultJSON(t, result)
	if m["status"] != "created" {
		t.Errorf("expected status created, got %v", m["status"])
	}

	// Get secret.
	result = callTool(t, cs, "secret_get", map[string]any{
		"key": "API_KEY", "profile": "test",
	})
	m = resultJSON(t, result)
	if m["value"] != "secret123" {
		t.Errorf("expected secret123, got %v", m["value"])
	}

	// List secrets.
	result = callTool(t, cs, "secret_list", map[string]any{"profile": "test"})
	m = resultJSON(t, result)
	secrets, ok := m["secrets"].([]any)
	if !ok || len(secrets) != 1 {
		t.Errorf("expected 1 secret, got %v", m["secrets"])
	}

	// Delete secret.
	result = callTool(t, cs, "secret_delete", map[string]any{
		"key": "API_KEY", "profile": "test",
	})
	m = resultJSON(t, result)
	if m["status"] != "deleted" {
		t.Errorf("expected status deleted, got %v", m["status"])
	}
}

func TestProfileTools(t *testing.T) {
	cs := setupTestServer(t)

	// Create.
	result := callTool(t, cs, "profile_create", map[string]any{"name": "staging"})
	m := resultJSON(t, result)
	if m["name"] != "staging" {
		t.Errorf("expected name staging, got %v", m["name"])
	}

	// List (should have the created profile).
	result = callTool(t, cs, "profile_list", nil)
	m = resultJSON(t, result)
	profiles := m["profiles"].([]any)
	if len(profiles) < 1 {
		t.Error("expected at least 1 profile")
	}

	// Delete.
	result = callTool(t, cs, "profile_delete", map[string]any{"name": "staging"})
	if result.IsError {
		t.Errorf("profile_delete error")
	}
}

func TestKeyTools(t *testing.T) {
	cs := setupTestServer(t)

	// Generate.
	result := callTool(t, cs, "key_generate", map[string]any{"name": "mykey"})
	m := resultJSON(t, result)
	if m["name"] != "mykey" {
		t.Errorf("expected name mykey, got %v", m["name"])
	}
	if m["algorithm"] != "ed25519" {
		t.Errorf("expected ed25519, got %v", m["algorithm"])
	}

	// List.
	result = callTool(t, cs, "key_list", nil)
	m = resultJSON(t, result)
	keys := m["keys"].([]any)
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}

	// Export public.
	result = callTool(t, cs, "key_export", map[string]any{"name": "mykey"})
	m = resultJSON(t, result)
	pem, _ := m["pem"].(string)
	if len(pem) == 0 {
		t.Error("expected non-empty PEM")
	}

	// Delete.
	result = callTool(t, cs, "key_delete", map[string]any{
		"name": "mykey", "algorithm": "ed25519",
	})
	m = resultJSON(t, result)
	if m["status"] != "deleted" {
		t.Errorf("expected status deleted, got %v", m["status"])
	}
}

func TestSignVerify(t *testing.T) {
	cs := setupTestServer(t)

	// Generate key.
	callTool(t, cs, "key_generate", map[string]any{"name": "sigkey"})

	// Sign.
	result := callTool(t, cs, "sign", map[string]any{
		"key_name": "sigkey", "data": "hello world",
	})
	m := resultJSON(t, result)
	sig, _ := m["signature"].(string)
	if sig == "" {
		t.Fatal("empty signature")
	}

	// Verify valid.
	result = callTool(t, cs, "verify", map[string]any{
		"key_name": "sigkey", "data": "hello world", "signature": sig,
	})
	m = resultJSON(t, result)
	if m["valid"] != true {
		t.Error("expected valid = true")
	}

	// Verify invalid.
	result = callTool(t, cs, "verify", map[string]any{
		"key_name": "sigkey", "data": "tampered", "signature": sig,
	})
	m = resultJSON(t, result)
	if m["valid"] != false {
		t.Error("expected valid = false")
	}
}

func TestAuditLog(t *testing.T) {
	cs := setupTestServer(t)

	// Create profile and secret to generate audit entries.
	callTool(t, cs, "profile_create", map[string]any{"name": "auditprof", "is_default": true})
	callTool(t, cs, "secret_set", map[string]any{
		"key": "K", "value": "V", "profile": "auditprof",
	})

	result := callTool(t, cs, "audit_log", map[string]any{"limit": 10})
	m := resultJSON(t, result)
	entries, ok := m["entries"].([]any)
	if !ok {
		t.Fatalf("expected entries array, got %T", m["entries"])
	}
	if len(entries) == 0 {
		t.Error("expected audit entries")
	}

	// With profile filter.
	result = callTool(t, cs, "audit_log", map[string]any{
		"limit": 10, "profile": "auditprof",
	})
	m = resultJSON(t, result)
	entries = m["entries"].([]any)
	if len(entries) == 0 {
		t.Error("expected entries for profile filter")
	}
}

func TestVaultStatus(t *testing.T) {
	cs := setupTestServer(t)

	result := callTool(t, cs, "vault_status", nil)
	m := resultJSON(t, result)
	if m["initialized"] != true {
		t.Error("expected initialized = true")
	}
}

func TestInstallationID(t *testing.T) {
	cs := setupTestServer(t)

	result := callTool(t, cs, "installation_id", nil)
	m := resultJSON(t, result)
	id, _ := m["id"].(string)
	if len(id) != 64 {
		t.Errorf("expected 64-char ID, got %d chars", len(id))
	}
}

func TestReadResource(t *testing.T) {
	cs := setupTestServer(t)

	result, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "anvil://status",
	})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(result.Contents) == 0 {
		t.Fatal("expected resource contents")
	}
	if result.Contents[0].MIMEType != "application/json" {
		t.Errorf("expected application/json, got %s", result.Contents[0].MIMEType)
	}

	var status map[string]any
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &status); err != nil {
		t.Fatalf("unmarshal status: %v", err)
	}
	if status["initialized"] != true {
		t.Error("expected initialized = true in resource")
	}
}

func TestSecretGetError(t *testing.T) {
	cs := setupTestServer(t)

	result := callTool(t, cs, "secret_get", map[string]any{
		"key": "nonexistent", "profile": "default",
	})
	if !result.IsError {
		t.Error("expected error for nonexistent secret")
	}
}

func TestKeyGenerateECDSA(t *testing.T) {
	cs := setupTestServer(t)

	result := callTool(t, cs, "key_generate", map[string]any{
		"name": "eckey", "algorithm": "ecdsa-p256",
	})
	m := resultJSON(t, result)
	if m["algorithm"] != "ecdsa-p256" {
		t.Errorf("expected ecdsa-p256, got %v", m["algorithm"])
	}
}
