package vault_test

import (
	"errors"
	"path/filepath"
	"testing"

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
		if history[0].Value != "value1" {
			t.Errorf("version value = %q, want %q", history[0].Value, "value1")
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
