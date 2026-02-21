package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestVault(t *testing.T) *Vault {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	t.Setenv("ANVIL_DB_PATH", dbPath)
	t.Setenv("ANVIL_SKIP_TPM", "1")

	if err := Init(&Options{DBPath: dbPath}); err != nil {
		t.Fatalf("init vault: %v", err)
	}

	v, err := Open(&Options{DBPath: dbPath})
	if err != nil {
		t.Fatalf("open vault: %v", err)
	}

	t.Cleanup(func() { _ = v.Close() })

	return v
}

func TestGenerateKey(t *testing.T) {
	v := setupTestVault(t)

	tests := []struct {
		name      string
		keyName   string
		algorithm string
		wantErr   bool
	}{
		{"ed25519", "mykey", AlgorithmEd25519, false},
		{"ecdsa-p256", "eckey", AlgorithmECDSAP256, false},
		{"duplicate", "mykey", AlgorithmEd25519, true},
		{"bad algorithm", "badkey", "rsa-4096", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := v.GenerateKey(tt.keyName, tt.algorithm, "test key")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if info.Name != tt.keyName {
				t.Fatalf("name: got %s, want %s", info.Name, tt.keyName)
			}

			if info.Algorithm != tt.algorithm {
				t.Fatalf("algorithm: got %s, want %s", info.Algorithm, tt.algorithm)
			}

			if info.Fingerprint == "" {
				t.Fatal("empty fingerprint")
			}
		})
	}
}

func TestListKeys(t *testing.T) {
	v := setupTestVault(t)

	keys, err := v.ListKeys()
	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}

	_, _ = v.GenerateKey("a", AlgorithmEd25519, "")
	_, _ = v.GenerateKey("b", AlgorithmECDSAP256, "")

	keys, err = v.ListKeys()
	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestDeleteKey(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("delme", AlgorithmEd25519, "")

	if err := v.DeleteKey("delme", AlgorithmEd25519); err != nil {
		t.Fatal(err)
	}

	// Should be gone
	if err := v.DeleteKey("delme", AlgorithmEd25519); err == nil {
		t.Fatal("expected error deleting nonexistent key")
	}
}

func TestDeleteKeyByName(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("multi", AlgorithmEd25519, "")

	if err := v.DeleteKey("multi", ""); err != nil {
		t.Fatal(err)
	}

	keys, _ := v.ListKeys()
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys after delete, got %d", len(keys))
	}
}

func TestExportImportKeyPEM(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("exportme", AlgorithmEd25519, "test")

	// Export public
	pubPEM, err := v.ExportKeyPEM("exportme", false)
	if err != nil {
		t.Fatal(err)
	}

	if len(pubPEM) == 0 {
		t.Fatal("empty public PEM")
	}

	// Export private
	privPEM, err := v.ExportKeyPEM("exportme", true)
	if err != nil {
		t.Fatal(err)
	}

	if len(privPEM) == 0 {
		t.Fatal("empty private PEM")
	}

	// Import the private key under a new name
	info, err := v.ImportKeyPEM("imported", privPEM, "imported key")
	if err != nil {
		t.Fatal(err)
	}

	if info.Algorithm != AlgorithmEd25519 {
		t.Fatalf("algorithm: got %s, want ed25519", info.Algorithm)
	}

	// Cannot import public-only
	_, err = v.ImportKeyPEM("pubonly", pubPEM, "")
	if err == nil {
		t.Fatal("expected error importing public key only")
	}
}

func TestExportKeyNotFound(t *testing.T) {
	v := setupTestVault(t)

	_, err := v.ExportKeyPEM("nonexistent", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestImportKeyDuplicate(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("dup", AlgorithmEd25519, "")

	privPEM, _ := v.ExportKeyPEM("dup", true)

	_, err := v.ImportKeyPEM("dup", privPEM, "")
	if err == nil {
		t.Fatal("expected error importing duplicate")
	}
}

func TestImportKeyFromFile(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("filetest", AlgorithmECDSAP256, "")

	privPEM, _ := v.ExportKeyPEM("filetest", true)

	tmpFile := filepath.Join(t.TempDir(), "key.pem")
	if err := os.WriteFile(tmpFile, privPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	info, err := v.ImportKeyPEM("from-file", data, "from file")
	if err != nil {
		t.Fatal(err)
	}

	if info.Algorithm != AlgorithmECDSAP256 {
		t.Fatalf("algorithm: got %s, want ecdsa-p256", info.Algorithm)
	}
}
