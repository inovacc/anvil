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

func TestBackupRestoreWithVersionHistory(t *testing.T) {
	v := setupWithProfile(t)
	if err := v.SetPassword("testpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	// Create versions by overwriting.
	if err := v.Set("KEY", "v1", "default", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("KEY", "v2", "default", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("KEY", "v3", "default", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.Backup("testpass")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	// Restore into fresh vault.
	v2, _ := initAndOpen(t)
	if err := v2.Restore(encrypted, "testpass"); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Current value should be v3.
	val, err := v2.Get("KEY", "default")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "v3" {
		t.Errorf("current value = %q, want %q", val, "v3")
	}

	// Version history should be restored.
	history, err := v2.SecretHistory("KEY", "default")
	if err != nil {
		t.Fatalf("SecretHistory: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(history))
	}
	// Ordered DESC: history[0] is version 2 (v2), history[1] is version 1 (v1).
	if history[0].Value != "v2" {
		t.Errorf("version 2 value = %q, want %q", history[0].Value, "v2")
	}
}

func TestBackupRestoreWithPassword(t *testing.T) {
	v := setupWithProfile(t)
	if err := v.SetPassword("testpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if err := v.Set("K", "V", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.Backup("testpass")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	v2, _ := initAndOpen(t)
	if err := v2.Restore(encrypted, "testpass"); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Password hash should be restored — verify works.
	if err := v2.VerifyPassword("testpass"); err != nil {
		t.Fatalf("VerifyPassword after restore: %v", err)
	}

	has, err := v2.HasPassword()
	if err != nil {
		t.Fatalf("HasPassword: %v", err)
	}
	if !has {
		t.Error("expected password to be restored")
	}
}

func TestRestoreInvalidData(t *testing.T) {
	v, _ := initAndOpen(t)

	// Garbage data should fail decryption.
	err := v.Restore([]byte("not-encrypted-data"), "somepass")
	if err == nil {
		t.Fatal("expected error for invalid encrypted data")
	}
}

func TestBackupNoPassword(t *testing.T) {
	v := setupWithProfile(t)

	// Backup requires password verification — no password set should fail.
	_, err := v.Backup("anypass")
	if err == nil {
		t.Fatal("expected error when no password set")
	}
}
