package vault

import (
	"testing"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	v := setupTestVault(t)

	tests := []struct {
		name      string
		algorithm string
		data      []byte
	}{
		{"ed25519/text", AlgorithmEd25519, []byte("hello world")},
		{"ed25519/empty", AlgorithmEd25519, []byte{}},
		{"ecdsa-p256/text", AlgorithmECDSAP256, []byte("hello world")},
		{"ecdsa-p256/empty", AlgorithmECDSAP256, []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyName := "key-" + tt.name

			_, err := v.GenerateKey(keyName, tt.algorithm, "")
			if err != nil {
				t.Fatalf("generate key: %v", err)
			}

			result, err := v.Sign(keyName, tt.data)
			if err != nil {
				t.Fatalf("sign: %v", err)
			}

			if result.Signature == "" {
				t.Fatal("empty signature")
			}

			if result.KeyName != keyName {
				t.Fatalf("key name: got %s, want %s", result.KeyName, keyName)
			}

			verifyResult, err := v.Verify(keyName, tt.data, result.Signature)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}

			if !verifyResult.Valid {
				t.Fatal("signature should be valid")
			}

			// Tampered data
			tampered := append([]byte{}, tt.data...)
			tampered = append(tampered, 0xFF)

			verifyResult, err = v.Verify(keyName, tampered, result.Signature)
			if err != nil {
				t.Fatalf("verify tampered: %v", err)
			}

			if verifyResult.Valid {
				t.Fatal("tampered data should not verify")
			}
		})
	}
}

func TestSignKeyNotFound(t *testing.T) {
	v := setupTestVault(t)

	_, err := v.Sign("nonexistent", []byte("data"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyKeyNotFound(t *testing.T) {
	v := setupTestVault(t)

	_, err := v.Verify("nonexistent", []byte("data"), "c2ln")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyBadSignature(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("badtest", AlgorithmEd25519, "")

	result, err := v.Verify("badtest", []byte("data"), "aW52YWxpZA==")
	if err != nil {
		t.Fatalf("verify should not error for invalid sig: %v", err)
	}

	if result.Valid {
		t.Fatal("invalid signature should not verify")
	}
}

func TestSignVerifyCrossKey(t *testing.T) {
	v := setupTestVault(t)

	_, _ = v.GenerateKey("key1", AlgorithmEd25519, "")
	_, _ = v.GenerateKey("key2", AlgorithmEd25519, "")

	sigResult, _ := v.Sign("key1", []byte("data"))

	verifyResult, err := v.Verify("key2", []byte("data"), sigResult.Signature)
	if err != nil {
		t.Fatal(err)
	}

	if verifyResult.Valid {
		t.Fatal("cross-key verification should fail")
	}
}
