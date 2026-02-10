package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"crypto/sha256"

	"golang.org/x/crypto/hkdf"
)

const (
	keySize   = 32 // AES-256
	nonceSize = 12 // GCM standard nonce
	saltSize  = 32
)

// DeriveKey derives a 256-bit key from the machine ID and a salt using HKDF-SHA256.
func DeriveKey(machineID string, salt []byte) ([]byte, error) {
	hk := hkdf.New(sha256.New, []byte(machineID), salt, []byte("profile-vault-v1"))
	key := make([]byte, keySize)
	if _, err := io.ReadFull(hk, key); err != nil {
		return nil, fmt.Errorf("hkdf derive: %w", err)
	}
	return key, nil
}

// GenerateSalt returns a random salt for key derivation.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}

// GenerateKey returns a random 256-bit key suitable for AES-256.
func GenerateKey() ([]byte, error) {
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with the given key.
// Returns (ciphertext, nonce, error). The nonce is randomly generated.
func Encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("new gcm: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the given key and nonce.
func Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}

// SealMasterKey encrypts the master key with a machine-derived key.
func SealMasterKey(masterKey []byte, machineID string, salt []byte) (sealed, nonce []byte, err error) {
	machineKey, err := DeriveKey(machineID, salt)
	if err != nil {
		return nil, nil, fmt.Errorf("derive machine key: %w", err)
	}

	sealed, nonce, err = Encrypt(machineKey, masterKey)
	if err != nil {
		return nil, nil, fmt.Errorf("seal master key: %w", err)
	}

	return sealed, nonce, nil
}

// UnsealMasterKey decrypts the master key using a machine-derived key.
func UnsealMasterKey(sealed, nonce []byte, machineID string, salt []byte) ([]byte, error) {
	machineKey, err := DeriveKey(machineID, salt)
	if err != nil {
		return nil, fmt.Errorf("derive machine key: %w", err)
	}

	masterKey, err := Decrypt(machineKey, sealed, nonce)
	if err != nil {
		return nil, fmt.Errorf("unseal master key: %w", err)
	}

	return masterKey, nil
}
