package vault

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/inovacc/anvil/internal/application"
	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/anvil/internal/store"
)

// InitWithRecovery initializes a new vault and returns the BIP-39 recovery phrase.
// The caller must display the mnemonic to the user.
func InitWithRecovery(opts *Options) (string, error) {
	dbPath, err := application.GetApplicationDirectory()
	if err != nil {
		return "", fmt.Errorf("get application directory: %w", err)
	}

	if opts != nil && opts.DBPath != "" {
		dbPath = opts.DBPath
	} else if envPath := os.Getenv("ANVIL_DB_PATH"); envPath != "" {
		dbPath = envPath
	} else {
		dbPath = filepath.Join(dbPath, dbFileName)
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	vaultStore, err := store.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("open database: %w", err)
	}

	defer func() { _ = vaultStore.Close() }()

	has, err := vaultStore.HasSealedKey()
	if err != nil {
		return "", fmt.Errorf("check sealed key: %w", err)
	}

	if has {
		return "", ErrAlreadyInitialized
	}

	machineIDHash, err := crypto.MachineIDHash()
	if err != nil {
		return "", fmt.Errorf("hash machine ID: %w", err)
	}

	masterKey, err := crypto.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("generate master key: %w", err)
	}

	defer crypto.ZeroBytes(masterKey)

	// Generate BIP-39 mnemonic from master key.
	mnemonic, err := crypto.MasterKeyToMnemonic(masterKey)
	if err != nil {
		return "", fmt.Errorf("generate recovery phrase: %w", err)
	}

	// Store mnemonic hash for verification.
	mnemonicHash := crypto.MnemonicHash(mnemonic)
	if err := vaultStore.UpsertRecovery(mnemonicHash, true); err != nil {
		return "", fmt.Errorf("save recovery hash: %w", err)
	}

	// Try TPM-first, fall back to software sealing.
	if crypto.IsTPMAvailable() {
		sealedJSON, tpmErr := crypto.SealMasterKeyTPM(masterKey)
		if tpmErr == nil {
			if err := vaultStore.UpsertSealedKey(sealedJSON, []byte{}, []byte{}, machineIDHash, 1, crypto.SealMethodTPM); err != nil {
				return "", fmt.Errorf("save sealed key: %w", err)
			}

			return mnemonic, nil
		}
	}

	machineID, err := crypto.MachineID()
	if err != nil {
		return "", fmt.Errorf("get machine ID: %w", err)
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	sealed, nonce, err := crypto.SealMasterKey(masterKey, machineID, salt)
	if err != nil {
		return "", fmt.Errorf("seal master key: %w", err)
	}

	if err := vaultStore.UpsertSealedKey(sealed, nonce, salt, machineIDHash, 1, crypto.SealMethodSoftware); err != nil {
		return "", fmt.Errorf("save sealed key: %w", err)
	}

	return mnemonic, nil
}

// RecoverFromMnemonic recovers a vault on a new machine using a BIP-39 mnemonic.
// It converts the mnemonic back to the master key, verifies it against the stored hash,
// and re-seals the key to the current machine.
func RecoverFromMnemonic(mnemonic string, opts *Options) error {
	dbPath, err := resolveDBPath(opts)
	if err != nil {
		return err
	}

	vaultStore, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	defer func() { _ = vaultStore.Close() }()

	has, err := vaultStore.HasSealedKey()
	if err != nil {
		return fmt.Errorf("check sealed key: %w", err)
	}

	if !has {
		return ErrNotInitialized
	}

	// Check recovery is configured.
	hasRecovery, err := vaultStore.HasRecovery()
	if err != nil {
		return fmt.Errorf("check recovery: %w", err)
	}

	if !hasRecovery {
		return ErrNoRecovery
	}

	// Verify mnemonic hash matches.
	recovery, err := vaultStore.GetRecovery()
	if err != nil {
		return fmt.Errorf("get recovery: %w", err)
	}

	mnemonicHash := crypto.MnemonicHash(mnemonic)
	if !bytes.Equal(mnemonicHash, recovery.MnemonicHash) {
		return ErrRecoveryMismatch
	}

	// Convert mnemonic back to master key.
	masterKey, err := crypto.MnemonicToMasterKey(mnemonic)
	if err != nil {
		return ErrInvalidMnemonic
	}

	defer crypto.ZeroBytes(masterKey)

	// Re-seal to the current machine.
	machineIDHash, err := crypto.MachineIDHash()
	if err != nil {
		return fmt.Errorf("hash machine ID: %w", err)
	}

	sk, err := vaultStore.GetSealedKey()
	if err != nil {
		return fmt.Errorf("get sealed key: %w", err)
	}

	var keyVersion int64 = 1
	if sk.Version != nil {
		keyVersion = *sk.Version
	}

	if crypto.IsTPMAvailable() {
		sealedJSON, tpmErr := crypto.SealMasterKeyTPM(masterKey)
		if tpmErr == nil {
			return vaultStore.UpsertSealedKey(sealedJSON, []byte{}, []byte{}, machineIDHash, keyVersion, crypto.SealMethodTPM)
		}
	}

	machineID, err := crypto.MachineID()
	if err != nil {
		return fmt.Errorf("get machine ID: %w", err)
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	sealed, nonce, err := crypto.SealMasterKey(masterKey, machineID, salt)
	if err != nil {
		return fmt.Errorf("seal master key: %w", err)
	}

	return vaultStore.UpsertSealedKey(sealed, nonce, salt, machineIDHash, keyVersion, crypto.SealMethodSoftware)
}

// ShowRecoveryPhrase converts the live master key to a BIP-39 mnemonic.
func (v *Vault) ShowRecoveryPhrase() (string, error) {
	if err := v.checkSealed(); err != nil {
		return "", err
	}

	hasRecovery, err := v.store.HasRecovery()
	if err != nil {
		return "", fmt.Errorf("check recovery: %w", err)
	}

	if !hasRecovery {
		return "", ErrNoRecovery
	}

	mnemonic, err := crypto.MasterKeyToMnemonic(v.masterKey)
	if err != nil {
		return "", fmt.Errorf("generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// HasRecovery returns whether a recovery phrase has been configured.
func (v *Vault) HasRecovery() (bool, error) {
	return v.store.HasRecovery()
}
