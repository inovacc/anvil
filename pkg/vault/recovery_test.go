package vault_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func TestInitWithRecovery(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")

	t.Run("happy path", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")

		mnemonic, err := vault.InitWithRecovery(&vault.Options{DBPath: dbPath})
		if err != nil {
			t.Fatalf("InitWithRecovery: %v", err)
		}

		words := strings.Fields(mnemonic)
		if len(words) != 24 {
			t.Fatalf("expected 24-word mnemonic, got %d words", len(words))
		}
	})

	t.Run("already initialized", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")

		_, err := vault.InitWithRecovery(&vault.Options{DBPath: dbPath})
		if err != nil {
			t.Fatalf("first init: %v", err)
		}

		_, err = vault.InitWithRecovery(&vault.Options{DBPath: dbPath})
		if err == nil {
			t.Fatal("expected error on second init")
		}
	})

	t.Run("uses env var", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		t.Setenv("ANVIL_DB_PATH", dbPath)

		mnemonic, err := vault.InitWithRecovery(nil)
		if err != nil {
			t.Fatalf("InitWithRecovery: %v", err)
		}

		if len(strings.Fields(mnemonic)) != 24 {
			t.Fatal("expected 24-word mnemonic")
		}
	})
}

func TestRecoverFromMnemonic(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")

	t.Run("happy path", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		opts := &vault.Options{DBPath: dbPath}

		mnemonic, err := vault.InitWithRecovery(opts)
		if err != nil {
			t.Fatalf("init: %v", err)
		}

		// Recover should succeed with correct mnemonic.
		if err := vault.RecoverFromMnemonic(mnemonic, opts); err != nil {
			t.Fatalf("RecoverFromMnemonic: %v", err)
		}

		// Vault should still open after recovery.
		v, err := vault.Open(opts)
		if err != nil {
			t.Fatalf("Open after recovery: %v", err)
		}

		_ = v.Close()
	})

	t.Run("wrong mnemonic", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		opts := &vault.Options{DBPath: dbPath}

		_, err := vault.InitWithRecovery(opts)
		if err != nil {
			t.Fatalf("init: %v", err)
		}

		// Use a different valid mnemonic.
		wrongMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
		err = vault.RecoverFromMnemonic(wrongMnemonic, opts)
		if err == nil {
			t.Fatal("expected error with wrong mnemonic")
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		opts := &vault.Options{DBPath: dbPath}

		err := vault.RecoverFromMnemonic("some words", opts)
		if err == nil {
			t.Fatal("expected error for uninitialized vault")
		}
	})
}

func TestShowRecoveryPhrase(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")

	t.Run("happy path", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		opts := &vault.Options{DBPath: dbPath}

		mnemonic, err := vault.InitWithRecovery(opts)
		if err != nil {
			t.Fatalf("init: %v", err)
		}

		v, err := vault.Open(opts)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}

		defer func() { _ = v.Close() }()

		shown, err := v.ShowRecoveryPhrase()
		if err != nil {
			t.Fatalf("ShowRecoveryPhrase: %v", err)
		}

		if shown != mnemonic {
			t.Errorf("shown mnemonic differs from init mnemonic")
		}
	})

	t.Run("no recovery configured", func(t *testing.T) {
		// Init without recovery (regular Init).
		v, _ := initAndOpen(t)

		_, err := v.ShowRecoveryPhrase()
		if err == nil {
			t.Fatal("expected error when no recovery configured")
		}
	})
}

func TestHasRecovery(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")

	t.Run("with recovery", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "vault.db")
		opts := &vault.Options{DBPath: dbPath}

		_, err := vault.InitWithRecovery(opts)
		if err != nil {
			t.Fatalf("init: %v", err)
		}

		v, err := vault.Open(opts)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}

		defer func() { _ = v.Close() }()

		has, err := v.HasRecovery()
		if err != nil {
			t.Fatalf("HasRecovery: %v", err)
		}

		if !has {
			t.Error("expected HasRecovery = true")
		}
	})

	t.Run("without recovery", func(t *testing.T) {
		v, _ := initAndOpen(t)

		has, err := v.HasRecovery()
		if err != nil {
			t.Fatalf("HasRecovery: %v", err)
		}

		if has {
			t.Error("expected HasRecovery = false")
		}
	})
}

func TestInstallationID(t *testing.T) {
	v, _ := initAndOpen(t)

	id, err := v.InstallationID()
	if err != nil {
		t.Fatalf("InstallationID: %v", err)
	}

	if len(id) != 64 {
		t.Fatalf("expected 64-char hex ID, got %d chars", len(id))
	}

	// Should be deterministic.
	id2, err := v.InstallationID()
	if err != nil {
		t.Fatalf("InstallationID second call: %v", err)
	}

	if id != id2 {
		t.Error("InstallationID should be deterministic")
	}
}

func TestKeyRotationWithAsymmetricKeys(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.SetPassword("rotpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("rot", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	// Create secrets and asymmetric keys.
	if err := v.Set("SEC", "val", "rot", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, err := v.GenerateKey("rotkey", vault.AlgorithmEd25519, "")
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	// Sign something before rotation.
	sigResult, err := v.Sign("rotkey", []byte("test data"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Rotate.
	if err := v.RotateKey("rotpass"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	// Secret still readable.
	val, err := v.Get("SEC", "rot")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "val" {
		t.Errorf("secret = %q, want %q", val, "val")
	}

	// Signing still works after rotation.
	sigResult2, err := v.Sign("rotkey", []byte("test data"))
	if err != nil {
		t.Fatalf("Sign after rotation: %v", err)
	}

	if sigResult2.Signature == "" {
		t.Fatal("empty signature after rotation")
	}

	// Verify old signature still works (public key unchanged).
	verifyResult, err := v.Verify("rotkey", []byte("test data"), sigResult.Signature)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if !verifyResult.Valid {
		t.Error("old signature should still be valid after rotation")
	}
}

func TestInitWithRecoverySecretsSurvive(t *testing.T) {
	t.Setenv("ANVIL_SKIP_TPM", "1")

	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}

	mnemonic, err := vault.InitWithRecovery(opts)
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Open, create profile and secret, close.
	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if err := v.CreateProfile("p", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	if err := v.Set("KEY", "VALUE", "p", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_ = v.Close()

	// Recover from mnemonic.
	if err := vault.RecoverFromMnemonic(mnemonic, opts); err != nil {
		t.Fatalf("RecoverFromMnemonic: %v", err)
	}

	// Reopen and verify secret.
	v, err = vault.Open(opts)
	if err != nil {
		t.Fatalf("Open after recovery: %v", err)
	}

	defer func() { _ = v.Close() }()

	val, err := v.Get("KEY", "p")
	if err != nil {
		t.Fatalf("Get after recovery: %v", err)
	}

	if val != "VALUE" {
		t.Errorf("got %q, want VALUE", val)
	}
}
