package vault_test

import (
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func setupWithProfile(t *testing.T) *vault.Vault {
	t.Helper()
	v, _ := initAndOpen(t)
	if err := v.CreateProfile("default", "default profile", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	return v
}

func TestBackupAndRestore(t *testing.T) {
	v := setupWithProfile(t)

	if err := v.SetPassword("testpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if err := v.Set("db-url", "postgres://localhost/test", "default", "database URL"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("api-key", "sk-12345", "default", "API key"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.Backup("testpass")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("Backup returned empty data")
	}

	// Restore into a fresh vault.
	v2, _ := initAndOpen(t)
	if err := v2.Restore(encrypted, "testpass"); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	val, err := v2.Get("db-url", "default")
	if err != nil {
		t.Fatalf("Get after restore: %v", err)
	}
	if val != "postgres://localhost/test" {
		t.Errorf("got %q, want %q", val, "postgres://localhost/test")
	}

	val, err = v2.Get("api-key", "default")
	if err != nil {
		t.Fatalf("Get after restore: %v", err)
	}
	if val != "sk-12345" {
		t.Errorf("got %q, want %q", val, "sk-12345")
	}
}

func TestBackupWrongPassword(t *testing.T) {
	v := setupWithProfile(t)
	if err := v.SetPassword("testpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	_, err := v.Backup("wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestRestoreWrongPassword(t *testing.T) {
	v := setupWithProfile(t)
	if err := v.SetPassword("testpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if err := v.Set("key", "val", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.Backup("testpass")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	v2, _ := initAndOpen(t)
	err = v2.Restore(encrypted, "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password on restore")
	}

	var ue *vault.UserError
	if ok := isUserError(err, &ue); ok {
		if ue.Hint == "" {
			t.Error("expected hint on UserError")
		}
	}
}

func isUserError(err error, target **vault.UserError) bool {
	ue, ok := err.(*vault.UserError)
	if ok {
		*target = ue
	}
	return ok
}
