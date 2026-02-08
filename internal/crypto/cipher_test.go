package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{name: "short text", plaintext: []byte("hello")},
		{name: "empty", plaintext: []byte("")},
		{name: "long text", plaintext: bytes.Repeat([]byte("a"), 10000)},
		{name: "binary data", plaintext: []byte{0x00, 0xFF, 0x01, 0xFE}},
	}

	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, nonce, err := Encrypt(key, tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}

			if bytes.Equal(ciphertext, tt.plaintext) && len(tt.plaintext) > 0 {
				t.Error("ciphertext should differ from plaintext")
			}

			decrypted, err := Decrypt(key, ciphertext, nonce)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	ciphertext, nonce, err := Encrypt(key1, []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(key2, ciphertext, nonce)
	if err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	key, _ := GenerateKey()

	ciphertext, nonce, err := Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	ciphertext[0] ^= 0xFF

	_, err = Decrypt(key, ciphertext, nonce)
	if err == nil {
		t.Error("expected error decrypting tampered ciphertext")
	}
}

func TestDeriveKey(t *testing.T) {
	salt1, _ := GenerateSalt()
	salt2, _ := GenerateSalt()

	key1, err := DeriveKey("machine-1", salt1)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}

	key2, err := DeriveKey("machine-1", salt1)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("same inputs should produce same key")
	}

	key3, err := DeriveKey("machine-1", salt2)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}

	if bytes.Equal(key1, key3) {
		t.Error("different salts should produce different keys")
	}

	key4, err := DeriveKey("machine-2", salt1)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}

	if bytes.Equal(key1, key4) {
		t.Error("different machine IDs should produce different keys")
	}

	if len(key1) != 32 {
		t.Errorf("key length = %d, want 32", len(key1))
	}
}

func TestSealUnsealMasterKey(t *testing.T) {
	masterKey, _ := GenerateKey()
	salt, _ := GenerateSalt()
	machineID := "test-machine-id"

	sealed, nonce, err := SealMasterKey(masterKey, machineID, salt)
	if err != nil {
		t.Fatalf("SealMasterKey: %v", err)
	}

	unsealed, err := UnsealMasterKey(sealed, nonce, machineID, salt)
	if err != nil {
		t.Fatalf("UnsealMasterKey: %v", err)
	}

	if !bytes.Equal(masterKey, unsealed) {
		t.Error("unsealed key should match original master key")
	}
}

func TestSealUnsealWrongMachine(t *testing.T) {
	masterKey, _ := GenerateKey()
	salt, _ := GenerateSalt()

	sealed, nonce, err := SealMasterKey(masterKey, "machine-A", salt)
	if err != nil {
		t.Fatalf("SealMasterKey: %v", err)
	}

	_, err = UnsealMasterKey(sealed, nonce, "machine-B", salt)
	if err == nil {
		t.Error("expected error unsealing with wrong machine ID")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt: %v", err)
	}

	if len(salt1) != 32 {
		t.Errorf("salt length = %d, want 32", len(salt1))
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("two salts should not be equal")
	}
}

func TestMachineID(t *testing.T) {
	id, err := MachineID()
	if err != nil {
		t.Fatalf("MachineID: %v", err)
	}

	if id == "" {
		t.Error("machine ID should not be empty")
	}

	id2, err := MachineID()
	if err != nil {
		t.Fatalf("MachineID: %v", err)
	}

	if id != id2 {
		t.Error("machine ID should be stable across calls")
	}
}

func TestMachineIDHash(t *testing.T) {
	hash, err := MachineIDHash()
	if err != nil {
		t.Fatalf("MachineIDHash: %v", err)
	}

	if len(hash) != 32 {
		t.Errorf("hash length = %d, want 32", len(hash))
	}
}
