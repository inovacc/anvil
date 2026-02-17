package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// MasterKeyToMnemonic converts a 32-byte master key to a 24-word BIP-39 mnemonic.
func MasterKeyToMnemonic(masterKey []byte) (string, error) {
	if len(masterKey) != keySize {
		return "", fmt.Errorf("master key must be %d bytes, got %d", keySize, len(masterKey))
	}

	mnemonic, err := bip39.NewMnemonic(masterKey)
	if err != nil {
		return "", fmt.Errorf("generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// MnemonicToMasterKey converts a 24-word BIP-39 mnemonic back to a 32-byte master key.
func MnemonicToMasterKey(mnemonic string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic phrase")
	}

	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("extract entropy: %w", err)
	}

	if len(entropy) != keySize {
		return nil, fmt.Errorf("unexpected entropy size: got %d, want %d", len(entropy), keySize)
	}

	return entropy, nil
}

// MnemonicHash returns the SHA-256 hash of a mnemonic string for verification.
func MnemonicHash(mnemonic string) []byte {
	h := sha256.Sum256([]byte(mnemonic))
	return h[:]
}
