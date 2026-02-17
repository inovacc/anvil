package vault_test

import (
	"testing"
)

func TestShareExportAndImport(t *testing.T) {
	v := setupWithProfile(t)

	if err := v.Set("secret1", "value1", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := v.Set("secret2", "value2", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.ShareExport("default", "mypassphrase")
	if err != nil {
		t.Fatalf("ShareExport: %v", err)
	}

	// Import into fresh vault with a different profile name.
	v2 := setupWithProfile(t)
	if err := v2.CreateProfile("imported", "imported secrets", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	export, err := v2.ShareImport(encrypted, "mypassphrase", "imported")
	if err != nil {
		t.Fatalf("ShareImport: %v", err)
	}

	if export.ProfileName != "default" {
		t.Errorf("expected original profile name %q, got %q", "default", export.ProfileName)
	}

	if len(export.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(export.Secrets))
	}

	val, err := v2.Get("secret1", "imported")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "value1" {
		t.Errorf("got %q, want %q", val, "value1")
	}
}

func TestShareImportWrongPassphrase(t *testing.T) {
	v := setupWithProfile(t)
	if err := v.Set("key", "val", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.ShareExport("default", "correct")
	if err != nil {
		t.Fatalf("ShareExport: %v", err)
	}

	v2 := setupWithProfile(t)

	_, err = v2.ShareImport(encrypted, "wrong", "default")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
}

func TestShareImportUsesOriginalProfile(t *testing.T) {
	v := setupWithProfile(t)
	if err := v.Set("key", "val", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	encrypted, err := v.ShareExport("default", "pass")
	if err != nil {
		t.Fatalf("ShareExport: %v", err)
	}

	// Import with empty target profile — should use "default".
	v2 := setupWithProfile(t)

	export, err := v2.ShareImport(encrypted, "pass", "")
	if err != nil {
		t.Fatalf("ShareImport: %v", err)
	}

	if export.ProfileName != "default" {
		t.Errorf("expected profile %q, got %q", "default", export.ProfileName)
	}

	val, err := v2.Get("key", "default")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "val" {
		t.Errorf("got %q, want %q", val, "val")
	}
}
