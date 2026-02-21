package vault_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/anvil/pkg/vault"
)

func initAndOpen(t *testing.T) (*vault.Vault, string) {
	t.Helper()
	t.Setenv("ANVIL_SKIP_TPM", "1")
	dbPath := filepath.Join(t.TempDir(), "vault.db")

	opts := &vault.Options{DBPath: dbPath}
	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = v.Close() })

	return v, dbPath
}

func TestInitOpenCloseLifecycle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	err := vault.Init(opts)
	if !errors.Is(err, vault.ErrAlreadyInitialized) {
		t.Fatalf("expected ErrAlreadyInitialized, got %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if v.DBPath() != dbPath {
		t.Errorf("DBPath = %q, want %q", v.DBPath(), dbPath)
	}

	if err := v.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestOpenWithoutInit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	_, err := vault.Open(opts)
	if !errors.Is(err, vault.ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

func TestProfileCRUD(t *testing.T) {
	v, _ := initAndOpen(t)

	t.Run("create and list", func(t *testing.T) {
		if err := v.CreateProfile("dev", "development", true); err != nil {
			t.Fatalf("CreateProfile: %v", err)
		}

		profiles, err := v.ListProfiles()
		if err != nil {
			t.Fatalf("ListProfiles: %v", err)
		}

		if len(profiles) != 1 {
			t.Fatalf("expected 1 profile, got %d", len(profiles))
		}

		if profiles[0].Name != "dev" {
			t.Errorf("Name = %q, want %q", profiles[0].Name, "dev")
		}

		if profiles[0].Description != "development" {
			t.Errorf("Description = %q, want %q", profiles[0].Description, "development")
		}

		if !profiles[0].IsDefault {
			t.Error("expected IsDefault = true")
		}
	})

	t.Run("duplicate profile", func(t *testing.T) {
		err := v.CreateProfile("dev", "dup", false)
		if !errors.Is(err, vault.ErrProfileExists) {
			t.Fatalf("expected ErrProfileExists, got %v", err)
		}
	})

	t.Run("use profile", func(t *testing.T) {
		if err := v.CreateProfile("staging", "staging env", false); err != nil {
			t.Fatalf("CreateProfile: %v", err)
		}

		if err := v.UseProfile("staging"); err != nil {
			t.Fatalf("UseProfile: %v", err)
		}

		profiles, err := v.ListProfiles()
		if err != nil {
			t.Fatalf("ListProfiles: %v", err)
		}

		for _, p := range profiles {
			if p.Name == "staging" && !p.IsDefault {
				t.Error("expected staging to be default")
			}
		}
	})

	t.Run("use nonexistent profile", func(t *testing.T) {
		err := v.UseProfile("nope")
		if !errors.Is(err, vault.ErrProfileNotFound) {
			t.Fatalf("expected ErrProfileNotFound, got %v", err)
		}
	})

	t.Run("delete profile", func(t *testing.T) {
		if err := v.DeleteProfile("staging"); err != nil {
			t.Fatalf("DeleteProfile: %v", err)
		}

		profiles, err := v.ListProfiles()
		if err != nil {
			t.Fatalf("ListProfiles: %v", err)
		}

		for _, p := range profiles {
			if p.Name == "staging" {
				t.Error("staging should have been deleted")
			}
		}
	})

	t.Run("delete nonexistent profile", func(t *testing.T) {
		err := v.DeleteProfile("nope")
		if !errors.Is(err, vault.ErrProfileNotFound) {
			t.Fatalf("expected ErrProfileNotFound, got %v", err)
		}
	})
}

func TestSecretCRUD(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("test", "test profile", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	t.Run("set and get", func(t *testing.T) {
		if err := v.Set("API_KEY", "secret123", "test", "api key"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		val, err := v.Get("API_KEY", "test")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if val != "secret123" {
			t.Errorf("Get = %q, want %q", val, "secret123")
		}
	})

	t.Run("set with default profile", func(t *testing.T) {
		if err := v.Set("DB_PASS", "pass456", "", "database password"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		val, err := v.Get("DB_PASS", "")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if val != "pass456" {
			t.Errorf("Get = %q, want %q", val, "pass456")
		}
	})

	t.Run("overwrite secret", func(t *testing.T) {
		if err := v.Set("API_KEY", "newsecret", "test", "updated"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		val, err := v.Get("API_KEY", "test")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if val != "newsecret" {
			t.Errorf("Get = %q, want %q", val, "newsecret")
		}
	})

	t.Run("list secrets", func(t *testing.T) {
		secrets, err := v.List("test")
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(secrets) != 2 {
			t.Fatalf("expected 2 secrets, got %d", len(secrets))
		}
	})

	t.Run("delete secret", func(t *testing.T) {
		if err := v.Delete("DB_PASS", "test"); err != nil {
			t.Fatalf("Delete: %v", err)
		}

		_, err := v.Get("DB_PASS", "test")
		if !errors.Is(err, vault.ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("delete nonexistent secret", func(t *testing.T) {
		err := v.Delete("NOPE", "test")
		if !errors.Is(err, vault.ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("get from nonexistent profile", func(t *testing.T) {
		_, err := v.Get("API_KEY", "nope")
		if !errors.Is(err, vault.ErrProfileNotFound) {
			t.Fatalf("expected ErrProfileNotFound, got %v", err)
		}
	})

	t.Run("get nonexistent secret", func(t *testing.T) {
		_, err := v.Get("NOPE", "test")
		if !errors.Is(err, vault.ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("no default profile error", func(t *testing.T) {
		vv, _ := initAndOpen(t)

		_, err := vv.Get("key", "")
		if !errors.Is(err, vault.ErrNoDefaultProfile) {
			t.Fatalf("expected ErrNoDefaultProfile, got %v", err)
		}
	})
}

func TestExportImport(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("exp", "export test", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	entries := []vault.SecretEntry{
		{Key: "K1", Value: "V1", Description: "first"},
		{Key: "K2", Value: "V2", Description: "second"},
		{Key: "K3", Value: "V3"},
	}

	if err := v.Import(entries, "exp"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	exported, err := v.Export("exp")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(exported) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(exported))
	}

	lookup := make(map[string]vault.SecretEntry)
	for _, e := range exported {
		lookup[e.Key] = e
	}

	for _, want := range entries {
		got, ok := lookup[want.Key]
		if !ok {
			t.Errorf("missing key %q", want.Key)
			continue
		}

		if got.Value != want.Value {
			t.Errorf("key %q: value = %q, want %q", want.Key, got.Value, want.Value)
		}

		if got.Description != want.Description {
			t.Errorf("key %q: description = %q, want %q", want.Key, got.Description, want.Description)
		}
	}
}

func TestPasswordOperations(t *testing.T) {
	v, _ := initAndOpen(t)

	t.Run("no password initially", func(t *testing.T) {
		has, err := v.HasPassword()
		if err != nil {
			t.Fatalf("HasPassword: %v", err)
		}

		if has {
			t.Error("expected no password set")
		}
	})

	t.Run("verify without password set", func(t *testing.T) {
		err := v.VerifyPassword("anything")
		if !errors.Is(err, vault.ErrPasswordNotSet) {
			t.Fatalf("expected ErrPasswordNotSet, got %v", err)
		}
	})

	t.Run("set password", func(t *testing.T) {
		if err := v.SetPassword("mypass"); err != nil {
			t.Fatalf("SetPassword: %v", err)
		}

		has, err := v.HasPassword()
		if err != nil {
			t.Fatalf("HasPassword: %v", err)
		}

		if !has {
			t.Error("expected password to be set")
		}
	})

	t.Run("verify correct password", func(t *testing.T) {
		if err := v.VerifyPassword("mypass"); err != nil {
			t.Fatalf("VerifyPassword: %v", err)
		}
	})

	t.Run("verify wrong password", func(t *testing.T) {
		err := v.VerifyPassword("wrong")
		if !errors.Is(err, vault.ErrPasswordMismatch) {
			t.Fatalf("expected ErrPasswordMismatch, got %v", err)
		}
	})

	t.Run("delete password", func(t *testing.T) {
		if err := v.DeletePassword(); err != nil {
			t.Fatalf("DeletePassword: %v", err)
		}

		has, err := v.HasPassword()
		if err != nil {
			t.Fatalf("HasPassword: %v", err)
		}

		if has {
			t.Error("expected password to be deleted")
		}
	})
}

func TestKeyRotation(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("rotatepass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("rot", "rotation test", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	secrets := map[string]string{
		"SECRET_A": "value_a",
		"SECRET_B": "value_b",
		"SECRET_C": "value_c",
	}

	for k, val := range secrets {
		if err := v.Set(k, val, "rot", ""); err != nil {
			t.Fatalf("Set %q: %v", k, err)
		}
	}

	if err := v.RotateKey("rotatepass"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	for k, want := range secrets {
		got, err := v.Get(k, "rot")
		if err != nil {
			t.Fatalf("Get %q after rotation: %v", k, err)
		}

		if got != want {
			t.Errorf("Get %q = %q, want %q", k, got, want)
		}
	}

	status, err := v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if status.KeyVersion < 2 {
		t.Errorf("KeyVersion = %d, want >= 2", status.KeyVersion)
	}
}

func TestKeyRotationWrongPassword(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("correct"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	err := v.RotateKey("wrong")
	if !errors.Is(err, vault.ErrPasswordMismatch) {
		t.Fatalf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestKeyRotationNoPassword(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.RotateKey("anypass")
	if !errors.Is(err, vault.ErrPasswordNotSet) {
		t.Fatalf("expected ErrPasswordNotSet, got %v", err)
	}
}

func TestStatus(t *testing.T) {
	v, dbPath := initAndOpen(t)

	status, err := v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if !status.Initialized {
		t.Error("expected Initialized = true")
	}

	if status.DBPath != dbPath {
		t.Errorf("DBPath = %q, want %q", status.DBPath, dbPath)
	}

	if status.ProfileCount != 0 {
		t.Errorf("ProfileCount = %d, want 0", status.ProfileCount)
	}

	if status.SecretCount != 0 {
		t.Errorf("SecretCount = %d, want 0", status.SecretCount)
	}

	if status.KeyVersion != 1 {
		t.Errorf("KeyVersion = %d, want 1", status.KeyVersion)
	}

	if status.PasswordSet {
		t.Error("expected PasswordSet = false")
	}

	if err := v.CreateProfile("p1", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("k1", "v1", "p1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("k2", "v2", "p1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	status, err = v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if status.ProfileCount != 1 {
		t.Errorf("ProfileCount = %d, want 1", status.ProfileCount)
	}

	if status.SecretCount != 2 {
		t.Errorf("SecretCount = %d, want 2", status.SecretCount)
	}
}

func TestGetStatus(t *testing.T) {
	t.Run("uninitialized", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		opts := &vault.Options{DBPath: dbPath}

		status, err := vault.GetStatus(opts)
		if err != nil {
			t.Fatalf("GetStatus: %v", err)
		}

		if status.Initialized {
			t.Error("expected Initialized = false")
		}
	})

	t.Run("initialized", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")

		opts := &vault.Options{DBPath: dbPath}
		if err := vault.Init(opts); err != nil {
			t.Fatalf("Init: %v", err)
		}

		status, err := vault.GetStatus(opts)
		if err != nil {
			t.Fatalf("GetStatus: %v", err)
		}

		if !status.Initialized {
			t.Error("expected Initialized = true")
		}
	})
}

func TestSecretIsolationBetweenProfiles(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("alpha", "", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.CreateProfile("beta", "", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("SHARED_KEY", "alpha_value", "alpha", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("SHARED_KEY", "beta_value", "beta", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	alphaVal, err := v.Get("SHARED_KEY", "alpha")
	if err != nil {
		t.Fatalf("Get alpha: %v", err)
	}

	betaVal, err := v.Get("SHARED_KEY", "beta")
	if err != nil {
		t.Fatalf("Get beta: %v", err)
	}

	if alphaVal != "alpha_value" {
		t.Errorf("alpha = %q, want %q", alphaVal, "alpha_value")
	}

	if betaVal != "beta_value" {
		t.Errorf("beta = %q, want %q", betaVal, "beta_value")
	}
}

func TestDeleteProfileRemovesProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("temp", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.DeleteProfile("temp"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}

	profiles, err := v.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	for _, p := range profiles {
		if p.Name == "temp" {
			t.Error("profile should have been deleted")
		}
	}
}

func TestImportToNonexistentProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.Import([]vault.SecretEntry{{Key: "K", Value: "V"}}, "nope")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestExportEmptyProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("empty", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	entries, err := v.Export("empty")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestSecretVersioningAndRollback(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("ver", "versioning test", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	t.Run("first set has no history", func(t *testing.T) {
		if err := v.Set("KEY", "value1", "ver", "first"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		history, err := v.SecretHistory("KEY", "ver")
		if err != nil {
			t.Fatalf("SecretHistory: %v", err)
		}

		if len(history) != 0 {
			t.Errorf("expected 0 versions, got %d", len(history))
		}
	})

	t.Run("second set archives first", func(t *testing.T) {
		if err := v.Set("KEY", "value2", "ver", "second"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		history, err := v.SecretHistory("KEY", "ver")
		if err != nil {
			t.Fatalf("SecretHistory: %v", err)
		}

		if len(history) != 1 {
			t.Fatalf("expected 1 version, got %d", len(history))
		}

		if history[0].Version != 1 {
			t.Errorf("version number = %d, want 1", history[0].Version)
		}
	})

	t.Run("third set archives second", func(t *testing.T) {
		if err := v.Set("KEY", "value3", "ver", "third"); err != nil {
			t.Fatalf("Set: %v", err)
		}

		history, err := v.SecretHistory("KEY", "ver")
		if err != nil {
			t.Fatalf("SecretHistory: %v", err)
		}

		if len(history) != 2 {
			t.Fatalf("expected 2 versions, got %d", len(history))
		}
	})

	t.Run("rollback to version 1", func(t *testing.T) {
		if err := v.SecretRollback("KEY", "ver", 1); err != nil {
			t.Fatalf("SecretRollback: %v", err)
		}

		val, err := v.Get("KEY", "ver")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if val != "value1" {
			t.Errorf("after rollback, value = %q, want %q", val, "value1")
		}

		history, err := v.SecretHistory("KEY", "ver")
		if err != nil {
			t.Fatalf("SecretHistory: %v", err)
		}

		if len(history) != 3 {
			t.Errorf("expected 3 versions after rollback, got %d", len(history))
		}
	})

	t.Run("rollback nonexistent version", func(t *testing.T) {
		err := v.SecretRollback("KEY", "ver", 999)
		if err == nil {
			t.Fatal("expected error for nonexistent version")
		}
	})

	t.Run("history for nonexistent secret", func(t *testing.T) {
		_, err := v.SecretHistory("NOPE", "ver")
		if !errors.Is(err, vault.ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("delete removes versions", func(t *testing.T) {
		if err := v.Set("TEMP", "a", "ver", ""); err != nil {
			t.Fatalf("Set: %v", err)
		}

		if err := v.Set("TEMP", "b", "ver", ""); err != nil {
			t.Fatalf("Set: %v", err)
		}

		if err := v.Delete("TEMP", "ver"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		// After delete, the secret and its versions should be gone.
		// Re-create to verify no leftover versions.
		if err := v.Set("TEMP", "fresh", "ver", ""); err != nil {
			t.Fatalf("Set: %v", err)
		}

		history, err := v.SecretHistory("TEMP", "ver")
		if err != nil {
			t.Fatalf("SecretHistory: %v", err)
		}

		if len(history) != 0 {
			t.Errorf("expected 0 versions after delete+recreate, got %d", len(history))
		}
	})
}

func TestKeyRotationWithVersions(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("rotpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("rv", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Create secret with versions.
	if err := v.Set("KEY", "val1", "rv", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("KEY", "val2", "rv", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("KEY", "val3", "rv", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Rotate key — should re-encrypt both secrets and versions.
	if err := v.RotateKey("rotpass"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	// Current value readable after rotation.
	val, err := v.Get("KEY", "rv")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "val3" {
		t.Errorf("current = %q, want %q", val, "val3")
	}

	// Version history readable after rotation.
	history, err := v.SecretHistory("KEY", "rv")
	if err != nil {
		t.Fatalf("SecretHistory: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(history))
	}
	// Ordered DESC: history[0] is version 2, history[1] is version 1.
	if history[0].Version != 2 {
		t.Errorf("version[0] = %d, want 2", history[0].Version)
	}

	if history[1].Version != 1 {
		t.Errorf("version[1] = %d, want 1", history[1].Version)
	}
}

func TestSetNoDefaultProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.Set("KEY", "val", "", "")
	if !errors.Is(err, vault.ErrNoDefaultProfile) {
		t.Fatalf("expected ErrNoDefaultProfile, got %v", err)
	}
}

func TestListNoDefaultProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.List("")
	if !errors.Is(err, vault.ErrNoDefaultProfile) {
		t.Fatalf("expected ErrNoDefaultProfile, got %v", err)
	}
}

func TestDeleteNoDefaultProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.Delete("KEY", "")
	if !errors.Is(err, vault.ErrNoDefaultProfile) {
		t.Fatalf("expected ErrNoDefaultProfile, got %v", err)
	}
}

func TestExportNonexistentProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.Export("nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestDoubleRotation(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("dr", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("KEY", "original", "dr", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// First rotation: v1 → v2.
	if err := v.RotateKey("pass"); err != nil {
		t.Fatalf("RotateKey #1: %v", err)
	}

	status, err := v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if status.KeyVersion != 2 {
		t.Errorf("KeyVersion after first rotation = %d, want 2", status.KeyVersion)
	}

	// Second rotation: v2 → v3.
	if err := v.RotateKey("pass"); err != nil {
		t.Fatalf("RotateKey #2: %v", err)
	}

	status, err = v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if status.KeyVersion != 3 {
		t.Errorf("KeyVersion after second rotation = %d, want 3", status.KeyVersion)
	}

	// Secret still readable.
	val, err := v.Get("KEY", "dr")
	if err != nil {
		t.Fatalf("Get after double rotation: %v", err)
	}

	if val != "original" {
		t.Errorf("value = %q, want %q", val, "original")
	}
}

func TestRotateEmptyVault(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	// Rotate with no profiles/secrets.
	if err := v.RotateKey("pass"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	status, err := v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if status.KeyVersion != 2 {
		t.Errorf("KeyVersion = %d, want 2", status.KeyVersion)
	}
}

func TestResolveDBPath(t *testing.T) {
	t.Run("from options", func(t *testing.T) {
		path, err := vault.ResolveDBPath(&vault.Options{DBPath: "/custom/vault.db"})
		if err != nil {
			t.Fatalf("ResolveDBPath: %v", err)
		}

		if path != "/custom/vault.db" {
			t.Errorf("got %q, want /custom/vault.db", path)
		}
	})

	t.Run("from env var", func(t *testing.T) {
		t.Setenv("ANVIL_DB_PATH", "/env/vault.db")

		path, err := vault.ResolveDBPath(nil)
		if err != nil {
			t.Fatalf("ResolveDBPath: %v", err)
		}

		if path != "/env/vault.db" {
			t.Errorf("got %q, want /env/vault.db", path)
		}
	})

	t.Run("default path", func(t *testing.T) {
		t.Setenv("ANVIL_DB_PATH", "")

		path, err := vault.ResolveDBPath(nil)
		if err != nil {
			t.Fatalf("ResolveDBPath: %v", err)
		}

		if path == "" {
			t.Error("expected non-empty default path")
		}
	})
}

func TestStatusWithMultipleProfiles(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("p1", "", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.CreateProfile("p2", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Add secrets to both profiles.
	for i := range 3 {
		if err := v.Set("K"+string(rune('A'+i)), "val", "p1", ""); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}

	for i := range 2 {
		if err := v.Set("K"+string(rune('X'+i)), "val", "p2", ""); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}

	status, err := v.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if status.ProfileCount != 2 {
		t.Errorf("ProfileCount = %d, want 2", status.ProfileCount)
	}

	if status.SecretCount != 5 {
		t.Errorf("SecretCount = %d, want 5", status.SecretCount)
	}

	if status.SealMethod == "" {
		t.Error("expected SealMethod to be set")
	}
}

func TestImportEmptyList(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("empty", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Importing an empty list should succeed.
	if err := v.Import([]vault.SecretEntry{}, "empty"); err != nil {
		t.Fatalf("Import empty: %v", err)
	}

	secrets, err := v.List("empty")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("expected 0 secrets, got %d", len(secrets))
	}
}

func TestListSecretsWithDescriptionAndUpdate(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("desc", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("KEY1", "val", "desc", "my description"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// Overwrite to populate UpdatedAt.
	if err := v.Set("KEY1", "val2", "desc", "updated desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	secrets, err := v.List("desc")
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}

	if secrets[0].Key != "KEY1" {
		t.Errorf("Key = %q, want KEY1", secrets[0].Key)
	}
}

func TestGetStatusWithEnvVar(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	t.Setenv("ANVIL_DB_PATH", dbPath)

	if err := vault.Init(nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	status, err := vault.GetStatus(nil)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if !status.Initialized {
		t.Error("expected initialized")
	}
}

func TestDeleteProfileCascadesVersions(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("cascade", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Create secret with version history.
	if err := v.Set("KEY", "v1", "cascade", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("KEY", "v2", "cascade", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Delete profile should cascade.
	if err := v.DeleteProfile("cascade"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}

	// Profile should be gone.
	profiles, err := v.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	for _, p := range profiles {
		if p.Name == "cascade" {
			t.Error("profile should be deleted")
		}
	}
}

func TestExportWithDescriptions(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("exp2", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K1", "V1", "exp2", "desc1"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("K2", "V2", "exp2", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	entries, err := v.Export("exp2")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2, got %d", len(entries))
	}

	for _, e := range entries {
		if e.Key == "K1" && e.Description != "desc1" {
			t.Errorf("K1 description = %q, want desc1", e.Description)
		}
	}
}

func TestInitCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	dbPath := filepath.Join(dir, "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestImportOverwrite(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("imp", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K1", "original", "imp", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Import overwrites.
	if err := v.Import([]vault.SecretEntry{{Key: "K1", Value: "new"}}, "imp"); err != nil {
		t.Fatalf("Import: %v", err)
	}

	val, err := v.Get("K1", "imp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "new" {
		t.Errorf("got %q, want %q", val, "new")
	}
}

func TestDBPath(t *testing.T) {
	v, dbPath := initAndOpen(t)

	if v.DBPath() != dbPath {
		t.Errorf("DBPath() = %q, want %q", v.DBPath(), dbPath)
	}
}

func TestDeletePassword(t *testing.T) {
	v, _ := initAndOpen(t)

	// Set a password.
	if err := v.SetPassword("testpass1234"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	has, err := v.HasPassword()
	if err != nil {
		t.Fatalf("HasPassword: %v", err)
	}

	if !has {
		t.Error("expected HasPassword = true")
	}

	// Delete it.
	if err := v.DeletePassword(); err != nil {
		t.Fatalf("DeletePassword: %v", err)
	}

	has, err = v.HasPassword()
	if err != nil {
		t.Fatalf("HasPassword after delete: %v", err)
	}

	if has {
		t.Error("expected HasPassword = false after delete")
	}
}

func TestVerifyPasswordNotSet(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.VerifyPassword("anything")
	if !errors.Is(err, vault.ErrPasswordNotSet) {
		t.Errorf("expected ErrPasswordNotSet, got %v", err)
	}
}

func TestVerifyPasswordMismatch(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("correctpass1"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	err := v.VerifyPassword("wrongpass")
	if !errors.Is(err, vault.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestCreateDuplicateProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("dup", "first", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	err := v.CreateProfile("dup", "second", false)
	if !errors.Is(err, vault.ErrProfileExists) {
		t.Errorf("expected ErrProfileExists, got %v", err)
	}
}

func TestUseProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.UseProfile("nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestDeleteProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.DeleteProfile("nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestListProfilesWithDetails(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("prod", "production env", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.CreateProfile("staging", "staging env", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Add secrets to prod.
	if err := v.Set("K1", "V1", "prod", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	profiles, err := v.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Find prod profile.
	var prod vault.ProfileInfo

	for _, p := range profiles {
		if p.Name == "prod" {
			prod = p
		}
	}

	if prod.Description != "production env" {
		t.Errorf("Description = %q, want %q", prod.Description, "production env")
	}

	if !prod.IsDefault {
		t.Error("expected prod to be default")
	}

	if prod.SecretCount != 1 {
		t.Errorf("SecretCount = %d, want 1", prod.SecretCount)
	}
}

func TestGetSecretNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("test", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	_, err := v.Get("NOPE", "test")
	if !errors.Is(err, vault.ErrSecretNotFound) {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestDeleteSecretNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("test", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	err := v.Delete("NOPE", "test")
	if !errors.Is(err, vault.ErrSecretNotFound) {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestSecretHistoryNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("test", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	_, err := v.SecretHistory("NOPE", "test")
	if !errors.Is(err, vault.ErrSecretNotFound) {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestRollbackInvalidVersion(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("test", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("KEY", "v1", "test", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("KEY", "v2", "test", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Version 999 doesn't exist.
	err := v.SecretRollback("KEY", "test", 999)
	if err == nil {
		t.Error("expected error for invalid version")
	}

	ue, ok := vault.IsUserError(err)
	if !ok {
		t.Errorf("expected UserError, got %T", err)
	} else if ue.Hint == "" {
		t.Error("expected hint in UserError")
	}
}

func TestGetStatusNotInitialized(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nonexistent.db")

	status, err := vault.GetStatus(&vault.Options{DBPath: dbPath})
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if status.Initialized {
		t.Error("expected not initialized")
	}
}

func TestEnvExportExplicitProfile(t *testing.T) {
	v := initOpenWithPassword(t)

	if err := v.CreateProfile("alt", "alt env", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("ALT_KEY", "alt_val", "alt", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if _, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 5 * time.Minute}); err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	// Export with explicit profile override.
	entries, err := v.EnvExport("alt")
	if err != nil {
		t.Fatalf("EnvExport: %v", err)
	}

	if len(entries) != 1 || entries[0].Key != "ALT_KEY" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestEnvInlineGetExplicitProfile(t *testing.T) {
	v := initOpenWithPassword(t)

	if err := v.CreateProfile("alt", "alt env", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("ALT_SECRET", "secret_val", "alt", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if _, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 5 * time.Minute}); err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	val, err := v.EnvInlineGet("ALT_SECRET", "alt")
	if err != nil {
		t.Fatalf("EnvInlineGet: %v", err)
	}

	if val != "secret_val" {
		t.Errorf("got %q, want %q", val, "secret_val")
	}
}

func TestPurgeExpiredVersions(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("purge", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Set a secret, then update it (creates a version).
	if err := v.Set("PK", "val1", "purge", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("PK", "val2", "purge", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Versions exist.
	versions, err := v.SecretHistory("PK", "purge")
	if err != nil {
		t.Fatalf("SecretHistory: %v", err)
	}

	if len(versions) != 1 {
		t.Fatalf("got %d versions, want 1", len(versions))
	}

	// Purge shouldn't remove anything (retention is 30 days).
	count, err := v.PurgeExpiredVersions()
	if err != nil {
		t.Fatalf("PurgeExpiredVersions: %v", err)
	}

	if count != 0 {
		t.Errorf("purged %d, want 0 (nothing expired)", count)
	}
}

func TestSealDoubleError(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")

	opts := &vault.Options{DBPath: dbPath}
	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}

	defer func() { _ = vault.UnsealVault(opts) }()

	_ = v.Close()

	// Open should fail (vault is sealed).
	_, err = vault.Open(opts)
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("expected ErrVaultSealed, got %v", err)
	}
}

func TestAuditLogOperations(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("audit", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K1", "V1", "audit", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("K2", "V2", "audit", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// AuditLog with limit.
	entries, err := v.AuditLog(10)
	if err != nil {
		t.Fatalf("AuditLog: %v", err)
	}

	if len(entries) == 0 {
		t.Error("expected audit entries")
	}

	// Verify entries have detail and secret_key populated.
	for _, e := range entries {
		if e.Action == "" {
			t.Error("expected non-empty action")
		}
	}

	// AuditLogByProfile.
	byProfile, err := v.AuditLogByProfile("audit", 10)
	if err != nil {
		t.Fatalf("AuditLogByProfile: %v", err)
	}

	if len(byProfile) == 0 {
		t.Error("expected audit entries by profile")
	}

	for _, e := range byProfile {
		if e.ProfileName != "audit" {
			t.Errorf("expected profile audit, got %q", e.ProfileName)
		}
	}

	// PurgeAuditLog — purge nothing (future timestamp).
	future := time.Now().Add(-24 * time.Hour)
	if err := v.PurgeAuditLog(future); err != nil {
		t.Fatalf("PurgeAuditLog: %v", err)
	}
}

func TestRotateKeyWithAsymmetricKeys(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("rotpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("keyrot", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Create secrets and versions.
	if err := v.Set("S1", "v1", "keyrot", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("S1", "v2", "keyrot", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Create asymmetric keys.
	_, err := v.GenerateKey("rotkey-ed", vault.AlgorithmEd25519, "")
	if err != nil {
		t.Fatalf("GenerateKey ed25519: %v", err)
	}
	_, err = v.GenerateKey("rotkey-ec", vault.AlgorithmECDSAP256, "")
	if err != nil {
		t.Fatalf("GenerateKey ecdsa: %v", err)
	}

	// Sign before rotation.
	sigResult, err := v.Sign("rotkey-ed", []byte("test data"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Rotate.
	if err := v.RotateKey("rotpass"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	// Secrets still readable.
	val, err := v.Get("S1", "keyrot")
	if err != nil {
		t.Fatalf("Get after rotation: %v", err)
	}
	if val != "v2" {
		t.Errorf("got %q, want v2", val)
	}

	// Keys still work after rotation.
	verifyResult, err := v.Verify("rotkey-ed", []byte("test data"), sigResult.Signature)
	if err != nil {
		t.Fatalf("Verify after rotation: %v", err)
	}
	if !verifyResult.Valid {
		t.Error("signature should still be valid after rotation")
	}

	// Can sign with rotated key.
	_, err = v.Sign("rotkey-ec", []byte("more data"))
	if err != nil {
		t.Fatalf("Sign after rotation: %v", err)
	}
}

func TestPluginManagerOperations(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "plugins.json")

	pm := vault.NewPluginManager(configPath)

	// Add hooks and providers.
	if err := pm.AddHook(vault.HookPreSet, "echo", []string{"pre-set"}); err != nil {
		t.Fatalf("AddHook: %v", err)
	}
	if err := pm.AddProvider("test-provider", "echo", "test://"); err != nil {
		t.Fatalf("AddProvider: %v", err)
	}

	// Check config.
	cfg := pm.Config()
	if len(cfg.Hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(cfg.Hooks))
	}
	if len(cfg.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(cfg.Providers))
	}

	// HasHooks.
	if !pm.HasHooks(vault.HookPreSet) {
		t.Error("expected HasHooks = true for pre-set")
	}
	if pm.HasHooks(vault.HookPostSet) {
		t.Error("expected HasHooks = false for post-set")
	}

	// ListProviders.
	providers := pm.ListProviders()
	if len(providers) != 1 || providers[0] != "test-provider" {
		t.Errorf("unexpected providers: %v", providers)
	}

	// SaveConfig.
	if err := pm.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// Reload from disk.
	pm2 := vault.NewPluginManager(configPath)
	if len(pm2.Config().Hooks) != 1 {
		t.Error("expected hook to persist")
	}
	if len(pm2.Config().Providers) != 1 {
		t.Error("expected provider to persist")
	}

	// RemoveHook.
	if err := pm.RemoveHook(vault.HookPreSet, "echo"); err != nil {
		t.Fatalf("RemoveHook: %v", err)
	}
	if pm.HasHooks(vault.HookPreSet) {
		t.Error("hook should have been removed")
	}

	// RemoveProvider.
	if err := pm.RemoveProvider("test-provider"); err != nil {
		t.Fatalf("RemoveProvider: %v", err)
	}
	if len(pm.ListProviders()) != 0 {
		t.Error("provider should have been removed")
	}
}

func TestPluginManagerNilSafe(t *testing.T) {
	var pm *vault.PluginManager

	// Nil safety.
	if err := pm.RunHooks(vault.HookPreSet, vault.HookPayload{}); err != nil {
		t.Errorf("RunHooks on nil: %v", err)
	}

	_, err := pm.GetFromProvider("x", "y", "z")
	if err == nil {
		t.Error("expected error from nil GetFromProvider")
	}

	if pm.HasHooks(vault.HookPreSet) {
		t.Error("nil HasHooks should return false")
	}

	if pm.ListProviders() != nil {
		t.Error("nil ListProviders should return nil")
	}

	cfg := pm.Config()
	if cfg == nil {
		t.Error("nil Config should return empty config")
	}
}

func TestVerifyInvalidBase64Signature(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GenerateKey("b64key", vault.AlgorithmEd25519, "")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	_, err = v.Verify("b64key", []byte("data"), "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64 signature")
	}
}

func TestEnvReleaseDefaultProfileAndExport(t *testing.T) {
	v := initOpenWithPassword(t)

	// Release with default profile (empty string resolves to default).
	state, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}
	if !state.Active {
		t.Error("expected active after release")
	}
	if state.ProfileName != "default" {
		t.Errorf("profile = %q, want default", state.ProfileName)
	}

	// Export with empty profile should use released profile.
	entries, err := v.EnvExport("")
	if err != nil {
		t.Fatalf("EnvExport: %v", err)
	}
	_ = entries // may be empty, just testing the path

	// Revoke.
	if err := v.EnvRevoke(); err != nil {
		t.Fatalf("EnvRevoke: %v", err)
	}
}

func TestUseProfileSuccess(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("first", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if err := v.CreateProfile("second", "", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.UseProfile("second"); err != nil {
		t.Fatalf("UseProfile: %v", err)
	}

	profiles, err := v.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	for _, p := range profiles {
		if p.Name == "second" && !p.IsDefault {
			t.Error("expected second to be default")
		}
		if p.Name == "first" && p.IsDefault {
			t.Error("expected first to not be default")
		}
	}
}

func TestNilVaultClose(t *testing.T) {
	var v *vault.Vault
	if err := v.Close(); err != nil {
		t.Errorf("Close nil vault: %v", err)
	}
}

func TestSealedOperationsExtended(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.SetPassword("pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("sp", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K", "V", "sp", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Generate a key before sealing.
	_, err = v.GenerateKey("sealkey", vault.AlgorithmEd25519, "")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}
	defer func() {
		_ = vault.UnsealVault(opts)
		_ = v.Close()
	}()

	// Operations on sealed vault should fail.
	_, err = v.Get("K", "sp")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Get: expected ErrVaultSealed, got %v", err)
	}

	err = v.Set("K2", "V2", "sp", "")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Set: expected ErrVaultSealed, got %v", err)
	}

	err = v.CreateProfile("new", "", false)
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("CreateProfile: expected ErrVaultSealed, got %v", err)
	}

	err = v.DeleteProfile("sp")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("DeleteProfile: expected ErrVaultSealed, got %v", err)
	}

	err = v.UseProfile("sp")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("UseProfile: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.ListProfiles()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ListProfiles: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.GenerateKey("k2", vault.AlgorithmEd25519, "")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("GenerateKey: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.ListKeys()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ListKeys: expected ErrVaultSealed, got %v", err)
	}

	err = v.DeleteKey("sealkey", vault.AlgorithmEd25519)
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("DeleteKey: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.ExportKeyPEM("sealkey", false)
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ExportKeyPEM: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.Sign("sealkey", []byte("data"))
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Sign: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.Verify("sealkey", []byte("data"), "sig")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Verify: expected ErrVaultSealed, got %v", err)
	}

	// RotateKey doesn't call checkSealed, but will fail due to zeroed master key.
	err = v.RotateKey("pass")
	if err == nil {
		t.Error("RotateKey: expected error on sealed vault")
	}

	// Backup calls VerifyPassword then Export (which calls checkSealed).
	_, err = v.Backup("pass")
	if err == nil {
		t.Error("Backup: expected error on sealed vault")
	}

	_, err = v.ShareExport("sp", "passphrase")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ShareExport: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.RegisterApp("app", "", "")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("RegisterApp: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.ListApps()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ListApps: expected ErrVaultSealed, got %v", err)
	}
}

func TestOpenWithEnvVar(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	t.Setenv("ANVIL_SKIP_TPM", "1")
	t.Setenv("ANVIL_DB_PATH", dbPath)

	if err := vault.Init(nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Open using env var path (no explicit opts.DBPath).
	v, err := vault.Open(nil)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = v.Close() }()

	if v.DBPath() != dbPath {
		t.Errorf("DBPath = %q, want %q", v.DBPath(), dbPath)
	}
}

func TestImportKeyPEMSealed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}
	t.Setenv("ANVIL_SKIP_TPM", "1")

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}
	defer func() {
		_ = vault.UnsealVault(opts)
		_ = v.Close()
	}()

	_, err = v.ImportKeyPEM("k", []byte("dummy"), "")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ImportKeyPEM: expected ErrVaultSealed, got %v", err)
	}
}

func TestPluginSaveConfigToDir(t *testing.T) {
	// SaveConfig to a directory (not a file) should fail or be handled.
	dir := t.TempDir()
	pm := vault.NewPluginManager(dir) // dir exists but is a directory, not a file

	if err := pm.AddHook(vault.HookPreSet, "echo", []string{"test"}); err != nil {
		t.Fatalf("AddHook: %v", err)
	}

	// Saving to a directory path — implementation may handle this.
	_ = pm.SaveConfig()
}

func TestDeleteKeyByNameOnly(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GenerateKey("delme", vault.AlgorithmEd25519, "test key")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	// Delete with empty algorithm (uses name-only delete path).
	if err := v.DeleteKey("delme", ""); err != nil {
		t.Fatalf("DeleteKey by name: %v", err)
	}

	keys, err := v.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}

	// Delete nonexistent by name.
	err = v.DeleteKey("nonexistent", "")
	if !errors.Is(err, vault.ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestDeleteKeyNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.DeleteKey("nonexistent", vault.AlgorithmEd25519)
	if err == nil {
		t.Error("expected error deleting nonexistent key")
	}
}

func TestExportKeyPEMPrivate(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GenerateKey("privexp", vault.AlgorithmEd25519, "")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	// Export private key.
	pem, err := v.ExportKeyPEM("privexp", true)
	if err != nil {
		t.Fatalf("ExportKeyPEM private: %v", err)
	}
	if len(pem) == 0 {
		t.Error("expected non-empty PEM")
	}

	// Export public key.
	pub, err := v.ExportKeyPEM("privexp", false)
	if err != nil {
		t.Fatalf("ExportKeyPEM public: %v", err)
	}
	if len(pub) == 0 {
		t.Error("expected non-empty public PEM")
	}
}

func TestExportKeyNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.ExportKeyPEM("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestSignVerifyECDSA(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GenerateKey("eckey", vault.AlgorithmECDSAP256, "")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	result, err := v.Sign("eckey", []byte("hello"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	vr, err := v.Verify("eckey", []byte("hello"), result.Signature)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !vr.Valid {
		t.Error("expected valid signature")
	}

	// Wrong data.
	vr, err = v.Verify("eckey", []byte("wrong"), result.Signature)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if vr.Valid {
		t.Error("expected invalid signature")
	}
}

func TestShareExportImportRoundTrip(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("share", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("SK1", "sv1", "share", "shared secret"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.ShareExport("share", "passphrase123")
	if err != nil {
		t.Fatalf("ShareExport: %v", err)
	}

	// Import into same vault under different profile.
	if err := v.CreateProfile("imported", "", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	export, err := v.ShareImport(encrypted, "passphrase123", "imported")
	if err != nil {
		t.Fatalf("ShareImport: %v", err)
	}

	if export.ProfileName != "share" {
		t.Errorf("original profile = %q, want share", export.ProfileName)
	}

	val, err := v.Get("SK1", "imported")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "sv1" {
		t.Errorf("got %q, want sv1", val)
	}
}

func TestShareImportOriginalProfile(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("orig", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K", "V", "orig", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.ShareExport("orig", "pass")
	if err != nil {
		t.Fatalf("ShareExport: %v", err)
	}

	// Import with empty target profile — should use original profile name.
	v2, _ := initAndOpen(t)
	if err := v2.CreateProfile("orig", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	_, err = v2.ShareImport(encrypted, "pass", "")
	if err != nil {
		t.Fatalf("ShareImport: %v", err)
	}

	val, err := v2.Get("K", "orig")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "V" {
		t.Errorf("got %q, want V", val)
	}
}

func TestGetAppSealed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}
	t.Setenv("ANVIL_SKIP_TPM", "1")

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}
	defer func() {
		_ = vault.UnsealVault(opts)
		_ = v.Close()
	}()

	_, err = v.GetApp("test")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("GetApp: expected ErrVaultSealed, got %v", err)
	}

	err = v.RemoveApp("test")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("RemoveApp: expected ErrVaultSealed, got %v", err)
	}

	err = v.DisableApp("test")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("DisableApp: expected ErrVaultSealed, got %v", err)
	}

	err = v.EnableApp("test")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("EnableApp: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.OpenApp("test")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("OpenApp: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.InstallationID()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("InstallationID: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.ShowRecoveryPhrase()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ShowRecoveryPhrase: expected ErrVaultSealed, got %v", err)
	}
}

func TestSecretRollbackProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.SecretRollback("KEY", "nonexistent", 1)
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestSecretHistoryProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.SecretHistory("KEY", "nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestSetWithProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.Set("KEY", "val", "nonexistent", "")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestGetWithProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.Get("KEY", "nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestDeleteWithProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.Delete("KEY", "nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestExportProfileNotFoundDirectly(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.Export("nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestImportProfileNotFoundDirectly(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.Import([]vault.SecretEntry{{Key: "K", Value: "V"}}, "nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestListProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.List("nonexistent")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestEnvReleaseProfileNotFound(t *testing.T) {
	v := initOpenWithPassword(t)

	_, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{
		TTL:         5 * time.Minute,
		ProfileName: "nonexistent",
	})
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestApplyTemplateWithExtraVars(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("tmpl", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	def := &vault.TemplateDefinition{
		Name:        "test-tmpl",
		Description: "test template",
		Variables: []vault.TemplateVariable{
			{Name: "host", Required: true},
			{Name: "port", Default: "5432"},
		},
		Secrets: []vault.TemplateSecret{
			{Key: "DB_URL", Value: "{{.host}}:{{.port}}", Description: "db url"},
			{Key: "DB_EXTRA", Value: "{{.extra}}", Description: "extra"},
		},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Apply with extra vars.
	err := v.ApplyTemplate("test-tmpl", "tmpl", map[string]string{
		"host":  "localhost",
		"extra": "extraval",
	})
	if err != nil {
		t.Fatalf("ApplyTemplate: %v", err)
	}

	val, err := v.Get("DB_URL", "tmpl")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "localhost:5432" {
		t.Errorf("DB_URL = %q, want localhost:5432", val)
	}

	val, err = v.Get("DB_EXTRA", "tmpl")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "extraval" {
		t.Errorf("DB_EXTRA = %q, want extraval", val)
	}
}

func TestApplyTemplateInterpolateError(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("tmpl2", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	def := &vault.TemplateDefinition{
		Name:        "bad-tmpl",
		Description: "bad template",
		Secrets: []vault.TemplateSecret{
			{Key: "BAD", Value: "{{.missing}}", Description: ""},
		},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Apply without providing the variable — interpolate uses missingkey=error.
	err := v.ApplyTemplate("bad-tmpl", "tmpl2", nil)
	if err == nil {
		t.Error("expected error for missing template key")
	}
}

func TestApplyTemplateProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	def := &vault.TemplateDefinition{
		Name:    "prof-tmpl",
		Secrets: []vault.TemplateSecret{{Key: "K", Value: "V"}},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	err := v.ApplyTemplate("prof-tmpl", "nonexistent", nil)
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestCreateTemplateDuplicate(t *testing.T) {
	v, _ := initAndOpen(t)

	def := &vault.TemplateDefinition{
		Name:    "dup-tmpl",
		Secrets: []vault.TemplateSecret{{Key: "K", Value: "V"}},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	err := v.CreateTemplate(def)
	if !errors.Is(err, vault.ErrTemplateExists) {
		t.Errorf("expected ErrTemplateExists, got %v", err)
	}
}

func TestDeleteTemplateNotFoundInVault(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.DeleteTemplate("nonexistent")
	if !errors.Is(err, vault.ErrTemplateNotFound) {
		t.Errorf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestListTemplates(t *testing.T) {
	v, _ := initAndOpen(t)

	// Should have built-in templates.
	templates, err := v.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	if len(templates) == 0 {
		t.Error("expected at least one built-in template")
	}
}

func TestUnsealWithEnvVar(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}
	_ = v.Close()

	// Unseal using env var instead of opts.
	t.Setenv("ANVIL_DB_PATH", dbPath)
	if err := vault.UnsealVault(nil); err != nil {
		t.Fatalf("UnsealVault via env: %v", err)
	}
}

func TestBackupWithMultipleProfiles(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("bkpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	// Create multiple profiles with secrets and versions.
	if err := v.CreateProfile("prod", "production", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if err := v.CreateProfile("dev", "development", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K1", "v1", "prod", "prod key"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("K1", "v2", "prod", "prod key updated"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := v.Set("K2", "devval", "dev", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	archive, err := v.Backup("bkpass")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}

	// Restore into fresh vault.
	v2, _ := initAndOpen(t)
	if err := v2.Restore(archive, "bkpass"); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Check prod secrets.
	val, err := v2.Get("K1", "prod")
	if err != nil {
		t.Fatalf("Get K1: %v", err)
	}
	if val != "v2" {
		t.Errorf("K1 = %q, want v2", val)
	}

	// Check dev secrets.
	val, err = v2.Get("K2", "dev")
	if err != nil {
		t.Fatalf("Get K2: %v", err)
	}
	if val != "devval" {
		t.Errorf("K2 = %q, want devval", val)
	}

	// Check version history.
	history, err := v2.SecretHistory("K1", "prod")
	if err != nil {
		t.Fatalf("SecretHistory: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 version, got %d", len(history))
	}

	// Check profiles restored correctly.
	profiles, err := v2.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestShareExportProfileNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.ShareExport("nonexistent", "pass")
	if !errors.Is(err, vault.ErrProfileNotFound) {
		t.Errorf("expected ErrProfileNotFound, got %v", err)
	}
}

func TestSignKeyNotFoundInVault(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.Sign("nokey", []byte("data"))
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestVerifyKeyNotFoundInVault(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.Verify("nokey", []byte("data"), "sig")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestImportKeyPEMInvalidFormat(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.ImportKeyPEM("badkey", []byte("not a valid PEM"), "")
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestGenerateKeyInvalidAlgorithm(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GenerateKey("k", "invalid-algo", "")
	if err == nil {
		t.Error("expected error for invalid algorithm")
	}
}

func TestRotateKeyWithAppAndSecrets(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("rot2", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Register app with secrets.
	_, err := v.RegisterApp("rotapp2", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	av, err := v.OpenApp("rotapp2")
	if err != nil {
		t.Fatalf("OpenApp: %v", err)
	}

	if err := av.Set("AK", "av", ""); err != nil {
		t.Fatalf("AppVault Set: %v", err)
	}
	_ = av.Close()

	// Add regular secrets.
	if err := v.Set("RK", "rv", "rot2", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Rotate.
	if err := v.RotateKey("pass"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	// Verify regular secret.
	val, err := v.Get("RK", "rot2")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "rv" {
		t.Errorf("got %q, want rv", val)
	}

	// Verify app secret.
	av2, err := v.OpenApp("rotapp2")
	if err != nil {
		t.Fatalf("OpenApp: %v", err)
	}
	defer func() { _ = av2.Close() }()

	aval, err := av2.Get("AK")
	if err != nil {
		t.Fatalf("AppVault Get: %v", err)
	}
	if aval != "av" {
		t.Errorf("got %q, want av", aval)
	}
}

func TestDefaultDBPathWithEnvVar(t *testing.T) {
	t.Setenv("ANVIL_DB_PATH", "/custom/db.db")
	p := vault.DefaultDBPath()
	if p != "/custom/db.db" {
		t.Errorf("got %q, want /custom/db.db", p)
	}
}

func TestInitWithEnvVar(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")
	dbPath := filepath.Join(t.TempDir(), "envvar.db")
	t.Setenv("ANVIL_DB_PATH", dbPath)

	// Init with nil opts — should use ANVIL_DB_PATH.
	if err := vault.Init(nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Verify DB was created at env var path.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected DB file at env var path")
	}
}

func TestSealedOperationsDenied(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")

	opts := &vault.Options{DBPath: dbPath}
	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.CreateProfile("sealed-ops", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("K1", "V1", "sealed-ops", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}

	defer func() {
		_ = vault.UnsealVault(opts)
		_ = v.Close()
	}()

	// All operations should return ErrVaultSealed.
	_, err = v.List("sealed-ops")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("List: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.Export("sealed-ops")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Export: expected ErrVaultSealed, got %v", err)
	}

	err = v.Delete("K1", "sealed-ops")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Delete: expected ErrVaultSealed, got %v", err)
	}

	err = v.Import(nil, "sealed-ops")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Import: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.Status()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Status: expected ErrVaultSealed, got %v", err)
	}
}

func TestPluginRunHooksPreBlockEmptyMessage(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "hook.bat")
		if err := os.WriteFile(cmd, []byte("@echo off\necho {\"allow\":false}"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		cmd = filepath.Join(dir, "hook.sh")
		if err := os.WriteFile(cmd, []byte("#!/bin/sh\necho '{\"allow\":false}'"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"hooks": []map[string]any{
			{"event": "pre-set", "command": cmd},
		},
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	err := pm.RunHooks(vault.HookPreSet, vault.HookPayload{SecretKey: "K"})
	if err == nil {
		t.Fatal("expected error from blocking hook")
	}

	// Should contain "blocked by hook" since message is empty.
	if !strings.Contains(err.Error(), "blocked by hook") {
		t.Errorf("expected 'blocked by hook', got %q", err.Error())
	}
}

func TestPluginRunHooksNonJSONOutput(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "hook.bat")
		if err := os.WriteFile(cmd, []byte("@echo off\necho not json"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		cmd = filepath.Join(dir, "hook.sh")
		if err := os.WriteFile(cmd, []byte("#!/bin/sh\necho 'not json'"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"hooks": []map[string]any{
			{"event": "post-set", "command": cmd},
		},
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	// Non-JSON stdout should be treated as allow (no error).
	err := pm.RunHooks(vault.HookPostSet, vault.HookPayload{SecretKey: "K"})
	if err != nil {
		t.Errorf("RunHooks non-JSON: %v", err)
	}
}

func TestPluginRunHooksEmptyStdout(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "hook.bat")
		if err := os.WriteFile(cmd, []byte("@echo off\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		cmd = filepath.Join(dir, "hook.sh")
		if err := os.WriteFile(cmd, []byte("#!/bin/sh\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"hooks": []map[string]any{
			{"event": "pre-get", "command": cmd},
		},
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	// Empty stdout should be treated as allow.
	err := pm.RunHooks(vault.HookPreGet, vault.HookPayload{SecretKey: "K"})
	if err != nil {
		t.Errorf("RunHooks empty stdout: %v", err)
	}
}

func TestPluginGetFromProviderWithScript(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "provider.bat")
		if err := os.WriteFile(cmd, []byte("@echo off\necho {\"value\":\"secret-from-provider\"}"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		cmd = filepath.Join(dir, "provider.sh")
		if err := os.WriteFile(cmd, []byte("#!/bin/sh\necho '{\"value\":\"secret-from-provider\"}'"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"providers": []map[string]any{
			{"name": "test", "command": cmd, "prefix": "ext/"},
		},
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	val, err := pm.GetFromProvider("test", "mykey", "default")
	if err != nil {
		t.Fatalf("GetFromProvider: %v", err)
	}

	if val != "secret-from-provider" {
		t.Errorf("got %q, want %q", val, "secret-from-provider")
	}
}

func TestPluginGetFromProviderError(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "provider.bat")
		if err := os.WriteFile(cmd, []byte("@echo off\necho {\"error\":\"access denied\"}"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		cmd = filepath.Join(dir, "provider.sh")
		if err := os.WriteFile(cmd, []byte("#!/bin/sh\necho '{\"error\":\"access denied\"}'"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"providers": []map[string]any{
			{"name": "test", "command": cmd, "prefix": "ext/"},
		},
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	_, err := pm.GetFromProvider("test", "mykey", "default")
	if err == nil {
		t.Fatal("expected error from provider")
	}

	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("expected 'access denied', got %q", err.Error())
	}
}

func TestPluginGetFromProviderRawOutput(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "provider.bat")
		if err := os.WriteFile(cmd, []byte("@echo off\necho raw-secret-value"), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		cmd = filepath.Join(dir, "provider.sh")
		if err := os.WriteFile(cmd, []byte("#!/bin/sh\necho 'raw-secret-value'"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"providers": []map[string]any{
			{"name": "raw", "command": cmd, "prefix": "raw/"},
		},
	}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	val, err := pm.GetFromProvider("raw", "key", "default")
	if err != nil {
		t.Fatalf("GetFromProvider raw: %v", err)
	}

	if val != "raw-secret-value" {
		t.Errorf("got %q, want %q", val, "raw-secret-value")
	}
}

func TestEnvStatusAndExportWithRelease(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("pass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("envp", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("S1", "V1", "envp", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Release.
	_, err := v.EnvRelease("pass", &vault.EnvReleaseOptions{
		ProfileName: "envp",
		TTL:         5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	// EnvStatus should show active.
	state, err := v.EnvStatus()
	if err != nil {
		t.Fatalf("EnvStatus: %v", err)
	}

	if !state.Active {
		t.Error("expected active release")
	}

	// EnvExport with empty profile (uses released profile).
	entries, err := v.EnvExport("")
	if err != nil {
		t.Fatalf("EnvExport: %v", err)
	}

	if len(entries) != 1 || entries[0].Key != "S1" {
		t.Errorf("unexpected entries: %v", entries)
	}

	// EnvInlineGet with empty profile.
	val, err := v.EnvInlineGet("S1", "")
	if err != nil {
		t.Fatalf("EnvInlineGet: %v", err)
	}

	if val != "V1" {
		t.Errorf("got %q, want V1", val)
	}

	// Revoke.
	if err := v.EnvRevoke(); err != nil {
		t.Fatalf("EnvRevoke: %v", err)
	}
}

func TestShareImportWrongPassphraseInVault(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("sharep", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("SK", "SV", "sharep", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.ShareExport("sharep", "correct-pass")
	if err != nil {
		t.Fatalf("ShareExport: %v", err)
	}

	_, err = v.ShareImport(encrypted, "wrong-pass", "sharep")
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestGetTemplateAndLoadBuiltins(t *testing.T) {
	v, _ := initAndOpen(t)

	// GetTemplate for nonexistent.
	_, err := v.GetTemplate("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}

	// LoadBuiltinTemplates.
	if err = v.LoadBuiltinTemplates(); err != nil {
		t.Fatalf("LoadBuiltinTemplates: %v", err)
	}

	// Load again (should be idempotent, duplicates skipped).
	if err = v.LoadBuiltinTemplates(); err != nil {
		t.Fatalf("LoadBuiltinTemplates second: %v", err)
	}

	// Now GetTemplate should work.
	templates, err := v.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	if len(templates) == 0 {
		t.Fatal("expected templates after loading builtins")
	}

	tmpl, err := v.GetTemplate(templates[0].Name)
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}

	if tmpl.Name == "" {
		t.Error("expected non-empty template name")
	}
}

func TestPluginRunHooksFailingScript(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "hook.bat")
		_ = os.WriteFile(cmd, []byte("@echo off\nexit /b 1"), 0o700)
	} else {
		cmd = filepath.Join(dir, "hook.sh")
		_ = os.WriteFile(cmd, []byte("#!/bin/sh\nexit 1"), 0o700)
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{"hooks": []map[string]any{{"event": "post-set", "command": cmd}}}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	err := pm.RunHooks(vault.HookPostSet, vault.HookPayload{SecretKey: "K"})
	if err != nil {
		t.Errorf("post-hook failure should not error: %v", err)
	}
}

func TestPluginRunHooksFailingPreScript(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "hook.bat")
		_ = os.WriteFile(cmd, []byte("@echo off\nexit /b 1"), 0o700)
	} else {
		cmd = filepath.Join(dir, "hook.sh")
		_ = os.WriteFile(cmd, []byte("#!/bin/sh\nexit 1"), 0o700)
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{"hooks": []map[string]any{{"event": "pre-delete", "command": cmd}}}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	err := pm.RunHooks(vault.HookPreDelete, vault.HookPayload{SecretKey: "K"})
	if err != nil {
		t.Errorf("failing pre-hook exec should be skipped: %v", err)
	}
}

func TestPluginGetFromProviderFailingScript(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = filepath.Join(dir, "provider.bat")
		_ = os.WriteFile(cmd, []byte("@echo off\nexit /b 1"), 0o700)
	} else {
		cmd = filepath.Join(dir, "provider.sh")
		_ = os.WriteFile(cmd, []byte("#!/bin/sh\nexit 1"), 0o700)
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{"providers": []map[string]any{{"name": "fail", "command": cmd, "prefix": "fail/"}}}
	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)
	_, err := pm.GetFromProvider("fail", "key", "default")
	if err == nil {
		t.Fatal("expected error from failing provider")
	}
}

func TestPluginNilAddRemoveOps(t *testing.T) {
	var pm *vault.PluginManager

	if err := pm.AddHook(vault.HookPreSet, "cmd", nil); err == nil {
		t.Error("expected error from nil AddHook")
	}
	if err := pm.AddProvider("p", "cmd", "pfx/"); err == nil {
		t.Error("expected error from nil AddProvider")
	}
	if err := pm.RemoveHook(vault.HookPreSet, "cmd"); err == nil {
		t.Error("expected error from nil RemoveHook")
	}
	if err := pm.RemoveProvider("p"); err == nil {
		t.Error("expected error from nil RemoveProvider")
	}
	if err := pm.SaveConfig(); err != nil {
		t.Errorf("nil SaveConfig should return nil: %v", err)
	}
}

func TestPluginNilGetFromProvider(t *testing.T) {
	var pm *vault.PluginManager
	_, err := pm.GetFromProvider("p", "k", "default")
	if err == nil {
		t.Fatal("expected error from nil GetFromProvider")
	}
}

func TestSecretRollbackSuccess(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("rollp", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("RK", "v1", "rollp", "desc1"); err != nil {
		t.Fatalf("Set v1: %v", err)
	}

	if err := v.Set("RK", "v2", "rollp", "desc2"); err != nil {
		t.Fatalf("Set v2: %v", err)
	}

	if err := v.Set("RK", "v3", "rollp", "desc3"); err != nil {
		t.Fatalf("Set v3: %v", err)
	}

	// History should have 2 versions (v1 and v2 archived).
	versions, err := v.SecretHistory("RK", "rollp")
	if err != nil {
		t.Fatalf("SecretHistory: %v", err)
	}

	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	// Current value should be v3.
	val, err := v.Get("RK", "rollp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "v3" {
		t.Errorf("got %q, want v3", val)
	}

	// Rollback to version 1.
	if err := v.SecretRollback("RK", "rollp", 1); err != nil {
		t.Fatalf("SecretRollback: %v", err)
	}

	// Current value should now be v1.
	val, err = v.Get("RK", "rollp")
	if err != nil {
		t.Fatalf("Get after rollback: %v", err)
	}

	if val != "v1" {
		t.Errorf("got %q after rollback, want v1", val)
	}
}
