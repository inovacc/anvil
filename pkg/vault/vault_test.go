package vault_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inovacc/anvil/pkg/vault"
)

func initAndOpen(t *testing.T) (*vault.Vault, string) {
	t.Helper()
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
	for i := 0; i < 3; i++ {
		if err := v.Set("K"+string(rune('A'+i)), "val", "p1", ""); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
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
