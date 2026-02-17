package vault_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func initAndOpenScoped(t *testing.T) *vault.ScopedVault {
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

	return sv
}

func TestOpenScopedValid(t *testing.T) {
	sv := initAndOpenScoped(t)

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

func TestScopedGetReturnsMasked(t *testing.T) {
	sv := initAndOpenScoped(t)

	if err := sv.Set("API_KEY", "supersecretvalue123", "test key"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, err := sv.Get("API_KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	// Must NOT return plaintext
	if val == "supersecretvalue123" {
		t.Fatal("Get returned plaintext — RBAC violation")
	}

	// Must contain masking asterisks
	if !strings.Contains(val, "*") {
		t.Errorf("Get = %q, expected masked value with asterisks", val)
	}

	// Must preserve some visible chars (first 3 + last 3 for long values)
	if !strings.HasPrefix(val, "sup") {
		t.Errorf("masked value should start with 'sup', got %q", val)
	}

	if !strings.HasSuffix(val, "123") {
		t.Errorf("masked value should end with '123', got %q", val)
	}
}

func TestScopedGetNotFoundError(t *testing.T) {
	sv := initAndOpenScoped(t)

	_, err := sv.Get("NONEXISTENT")
	if !errors.Is(err, vault.ErrSecretNotFound) {
		t.Fatalf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestScopedSetAndDelete(t *testing.T) {
	sv := initAndOpenScoped(t)

	if err := sv.Set("KEY1", "value1", "desc"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	secrets, err := sv.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(secrets) != 1 || secrets[0].Key != "KEY1" {
		t.Fatalf("List unexpected: %+v", secrets)
	}

	if err := sv.Delete("KEY1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	secrets, err = sv.List()
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}

	if len(secrets) != 0 {
		t.Errorf("expected 0 secrets after delete, got %d", len(secrets))
	}
}

func TestScopedExportDenied(t *testing.T) {
	sv := initAndOpenScoped(t)

	if err := sv.Set("KEY1", "val1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, err := sv.Export()
	if !errors.Is(err, vault.ErrReadDenied) {
		t.Fatalf("expected ErrReadDenied from Export, got %v", err)
	}
}

func TestScopedImportAllowed(t *testing.T) {
	sv := initAndOpenScoped(t)

	entries := []vault.SecretEntry{
		{Key: "IMPORTED_KEY", Value: "imported_val", Description: "imported"},
	}

	if err := sv.Import(entries); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Verify it was stored (get returns masked)
	val, err := sv.Get("IMPORTED_KEY")
	if err != nil {
		t.Fatalf("Get after import: %v", err)
	}

	if !strings.Contains(val, "*") {
		t.Errorf("expected masked value, got %q", val)
	}
}

func TestScopedIsolation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

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

	if err := v.Set("SHARED_KEY", "value-a", "profile-a", ""); err != nil {
		t.Fatalf("Set in A: %v", err)
	}

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

	// CLI (full Vault) can read plaintext
	plainA, err := v.Get("SHARED_KEY", "profile-a")
	if err != nil {
		t.Fatalf("CLI Get A: %v", err)
	}

	if plainA != "value-a" {
		t.Errorf("CLI got %q, want %q", plainA, "value-a")
	}

	_ = v.Close()

	// Scoped A sees masked only
	svA, err := vault.OpenScoped(uuidA, opts)
	if err != nil {
		t.Fatalf("OpenScoped A: %v", err)
	}

	maskedA, err := svA.Get("SHARED_KEY")
	if err != nil {
		t.Fatalf("Scoped Get A: %v", err)
	}

	if maskedA == "value-a" {
		t.Error("scoped A returned plaintext — RBAC violation")
	}

	if !strings.Contains(maskedA, "*") {
		t.Errorf("scoped A value not masked: %q", maskedA)
	}

	// Scoped B sees only its own secrets
	_ = svA.Close()

	svB, err := vault.OpenScoped(uuidB, opts)
	if err != nil {
		t.Fatalf("OpenScoped B: %v", err)
	}

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
	sv := initAndOpenScoped(t)

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

func TestSealAndUnseal(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}

	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.CreateProfile("test", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("KEY", "value", "test", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Seal the vault
	if err := v.Seal(); err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if !v.IsSealed() {
		t.Error("expected IsSealed() = true after Seal()")
	}

	// All operations should fail
	_, err = v.Get("KEY", "test")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Get on sealed vault: expected ErrVaultSealed, got %v", err)
	}

	err = v.Set("KEY2", "val", "test", "")
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("Set on sealed vault: expected ErrVaultSealed, got %v", err)
	}

	_, err = v.ListProfiles()
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Errorf("ListProfiles on sealed vault: expected ErrVaultSealed, got %v", err)
	}

	_ = v.Close()

	// Open should fail while sealed
	_, err = vault.Open(opts)
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Fatalf("Open on sealed vault: expected ErrVaultSealed, got %v", err)
	}

	// OpenScoped should also fail
	_, err = vault.OpenScoped("any-uuid", opts)
	if !errors.Is(err, vault.ErrVaultSealed) {
		t.Fatalf("OpenScoped on sealed vault: expected ErrVaultSealed, got %v", err)
	}

	// Unseal
	if err := vault.UnsealVault(opts); err != nil {
		t.Fatalf("UnsealVault: %v", err)
	}

	// Unseal again should fail
	err = vault.UnsealVault(opts)
	if !errors.Is(err, vault.ErrVaultNotSealed) {
		t.Fatalf("UnsealVault again: expected ErrVaultNotSealed, got %v", err)
	}

	// Open should work again
	v2, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open after unseal: %v", err)
	}

	defer func() { _ = v2.Close() }()

	val, err := v2.Get("KEY", "test")
	if err != nil {
		t.Fatalf("Get after unseal: %v", err)
	}

	if val != "value" {
		t.Errorf("Get = %q, want %q", val, "value")
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"ab", "**"},
		{"abcd", "****"},
		{"abcde", "a***e"},
		{"abcdefgh", "a******h"},
		{"abcdefghij", "abc****hij"},
		{"supersecretvalue123", "sup*************123"},
	}

	for _, tc := range tests {
		got := vault.MaskValue(tc.input)
		if got != tc.want {
			t.Errorf("MaskValue(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
