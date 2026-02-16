package vault_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func initAndOpenScoped(t *testing.T) (*vault.ScopedVault, *vault.Vault, string) {
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

	if err := v.CreateProfile("scoped-test", "test profile", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	profiles, err := v.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	var profileUUID string
	for _, p := range profiles {
		if p.Name == "scoped-test" {
			profileUUID = p.UUID
			break
		}
	}

	if profileUUID == "" {
		t.Fatal("profile UUID is empty")
	}

	_ = v.Close()

	sv, err := vault.OpenScoped(profileUUID, opts)
	if err != nil {
		t.Fatalf("OpenScoped: %v", err)
	}

	t.Cleanup(func() { _ = sv.Close() })

	return sv, nil, dbPath
}

func TestOpenScopedValid(t *testing.T) {
	sv, _, _ := initAndOpenScoped(t)

	info := sv.ProfileInfo()
	if info.Name != "scoped-test" {
		t.Errorf("ProfileInfo().Name = %q, want %q", info.Name, "scoped-test")
	}

	if info.UUID == "" {
		t.Error("ProfileInfo().UUID is empty")
	}
}

func TestOpenScopedInvalidUUID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	_, err := vault.OpenScoped("nonexistent-uuid", opts)
	if !errors.Is(err, vault.ErrProfileUUIDNotFound) {
		t.Fatalf("expected ErrProfileUUIDNotFound, got %v", err)
	}
}

func TestScopedCRUD(t *testing.T) {
	sv, _, _ := initAndOpenScoped(t)

	// Set
	if err := sv.Set("API_KEY", "secret123", "test key"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Get
	val, err := sv.Get("API_KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "secret123" {
		t.Errorf("Get = %q, want %q", val, "secret123")
	}

	// List
	secrets, err := sv.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(secrets) != 1 {
		t.Fatalf("List len = %d, want 1", len(secrets))
	}
	if secrets[0].Key != "API_KEY" {
		t.Errorf("List[0].Key = %q, want %q", secrets[0].Key, "API_KEY")
	}

	// Delete
	if err := sv.Delete("API_KEY"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = sv.Get("API_KEY")
	if !errors.Is(err, vault.ErrSecretNotFound) {
		t.Fatalf("expected ErrSecretNotFound after delete, got %v", err)
	}
}

func TestScopedExportImport(t *testing.T) {
	sv, _, _ := initAndOpenScoped(t)

	if err := sv.Set("KEY1", "val1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := sv.Set("KEY2", "val2", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	entries, err := sv.Export()
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Export len = %d, want 2", len(entries))
	}

	// Delete and re-import
	if err := sv.Delete("KEY1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := sv.Delete("KEY2"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := sv.Import(entries); err != nil {
		t.Fatalf("Import: %v", err)
	}

	val, err := sv.Get("KEY1")
	if err != nil {
		t.Fatalf("Get after import: %v", err)
	}
	if val != "val1" {
		t.Errorf("Get = %q, want %q", val, "val1")
	}
}

func TestScopedIsolation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Create two profiles via full vault
	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.CreateProfile("profile-a", "", true); err != nil {
		t.Fatalf("CreateProfile A: %v", err)
	}
	if err := v.CreateProfile("profile-b", "", false); err != nil {
		t.Fatalf("CreateProfile B: %v", err)
	}

	// Set a secret in profile-a via full vault
	if err := v.Set("SHARED_KEY", "value-a", "profile-a", ""); err != nil {
		t.Fatalf("Set in A: %v", err)
	}

	// Set a different secret in profile-b
	if err := v.Set("SHARED_KEY", "value-b", "profile-b", ""); err != nil {
		t.Fatalf("Set in B: %v", err)
	}

	profiles, err := v.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}

	var uuidA, uuidB string
	for _, p := range profiles {
		switch p.Name {
		case "profile-a":
			uuidA = p.UUID
		case "profile-b":
			uuidB = p.UUID
		}
	}

	_ = v.Close()

	// Open scoped to profile-a
	svA, err := vault.OpenScoped(uuidA, opts)
	if err != nil {
		t.Fatalf("OpenScoped A: %v", err)
	}

	val, err := svA.Get("SHARED_KEY")
	if err != nil {
		t.Fatalf("Get from A: %v", err)
	}
	if val != "value-a" {
		t.Errorf("scoped A got %q, want %q", val, "value-a")
	}

	_ = svA.Close()

	// Open scoped to profile-b
	svB, err := vault.OpenScoped(uuidB, opts)
	if err != nil {
		t.Fatalf("OpenScoped B: %v", err)
	}

	val, err = svB.Get("SHARED_KEY")
	if err != nil {
		t.Fatalf("Get from B: %v", err)
	}
	if val != "value-b" {
		t.Errorf("scoped B got %q, want %q", val, "value-b")
	}

	// Profile B cannot see profile A's list
	secrets, err := svB.List()
	if err != nil {
		t.Fatalf("List B: %v", err)
	}
	if len(secrets) != 1 {
		t.Errorf("scoped B sees %d secrets, want 1", len(secrets))
	}

	_ = svB.Close()
}

func TestScopedProfileInfo(t *testing.T) {
	sv, _, _ := initAndOpenScoped(t)

	if err := sv.Set("KEY1", "val1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	info := sv.ProfileInfo()
	if info.SecretCount != 1 {
		t.Errorf("ProfileInfo().SecretCount = %d, want 1", info.SecretCount)
	}
	if info.Name != "scoped-test" {
		t.Errorf("ProfileInfo().Name = %q, want %q", info.Name, "scoped-test")
	}
}
