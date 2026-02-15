package vault_test

import (
	"testing"
	"time"

	"github.com/inovacc/anvil/pkg/vault"
)

func TestAuditLogOnSecretOperations(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("audit", "audit test", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("KEY1", "val1", "audit", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, err := v.Get("KEY1", "audit")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if err := v.Delete("KEY1", "audit"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	entries, err := v.AuditLog(100)
	if err != nil {
		t.Fatalf("AuditLog: %v", err)
	}

	actions := make(map[string]int)
	for _, e := range entries {
		actions[e.Action]++
	}

	tests := []struct {
		action string
		want   int
	}{
		{"profile.create", 1},
		{"secret.set", 1},
		{"secret.get", 1},
		{"secret.delete", 1},
	}

	for _, tt := range tests {
		if actions[tt.action] != tt.want {
			t.Errorf("action %q: got %d, want %d", tt.action, actions[tt.action], tt.want)
		}
	}
}

func TestAuditLogByProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("a", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if err := v.CreateProfile("b", "", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K", "V", "a", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("K", "V", "b", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	entries, err := v.AuditLogByProfile("a", 100)
	if err != nil {
		t.Fatalf("AuditLogByProfile: %v", err)
	}

	for _, e := range entries {
		if e.ProfileName != "a" {
			t.Errorf("expected profile a, got %q", e.ProfileName)
		}
	}
}

func TestAuditLogOnImportExport(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("ie", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Import([]vault.SecretEntry{{Key: "K1", Value: "V1"}}, "ie"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	_, err := v.Export("ie")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	entries, err := v.AuditLog(100)
	if err != nil {
		t.Fatalf("AuditLog: %v", err)
	}

	actions := make(map[string]bool)
	for _, e := range entries {
		actions[e.Action] = true
	}

	if !actions["secret.import"] {
		t.Error("missing secret.import audit entry")
	}
	if !actions["secret.export"] {
		t.Error("missing secret.export audit entry")
	}
}

func TestPurgeAuditLog(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("purge", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if err := v.Set("K", "V", "purge", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	entries, err := v.AuditLog(10)
	if err != nil {
		t.Fatalf("AuditLog: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected audit entries before purge")
	}

	// Purge everything older than 24 hours from now (i.e., everything).
	if err := v.PurgeAuditLog(time.Now().Add(24 * time.Hour)); err != nil {
		t.Fatalf("PurgeAuditLog: %v", err)
	}

	entries, err = v.AuditLog(10)
	if err != nil {
		t.Fatalf("AuditLog: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after purge, got %d", len(entries))
	}
}
