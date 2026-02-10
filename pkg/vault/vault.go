package vault

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/inovacc/profile/internal/application"
	"github.com/inovacc/profile/internal/crypto"
	"github.com/inovacc/profile/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// Vault provides encrypted secret storage organized by profiles.
type Vault struct {
	store     *store.Store
	masterKey []byte
	dbPath    string
}

// Init initializes a new vault, generating a master key sealed to this machine.
func Init(opts *Options) error {
	dbPath, err := application.GetApplicationDirectory()
	if err != nil {
		return fmt.Errorf("get application directory: %w", err)
	}

	if opts != nil && opts.DBPath != "" {
		dbPath = opts.DBPath
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
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

	if has {
		return ErrAlreadyInitialized
	}

	machineIDHash, err := crypto.MachineIDHash()
	if err != nil {
		return fmt.Errorf("hash machine ID: %w", err)
	}

	masterKey, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("generate master key: %w", err)
	}

	defer crypto.ZeroBytes(masterKey)

	// Try TPM-first, fall back to software sealing.
	if crypto.IsTPMAvailable() {
		sealedJSON, tpmErr := crypto.SealMasterKeyTPM(masterKey)
		if tpmErr == nil {
			if err := vaultStore.UpsertSealedKey(sealedJSON, nil, nil, machineIDHash, 1, crypto.SealMethodTPM); err != nil {
				return fmt.Errorf("save sealed key: %w", err)
			}

			return nil
		}
		// TPM seal failed — fall through to software path.
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

	if err := vaultStore.UpsertSealedKey(sealed, nonce, salt, machineIDHash, 1, crypto.SealMethodSoftware); err != nil {
		return fmt.Errorf("save sealed key: %w", err)
	}

	return nil
}

// Open opens an existing vault, unsealing the master key with the machine identity.
func Open(opts *Options) (*Vault, error) {
	dbPath, err := application.GetApplicationDirectory()
	if err != nil {
		return nil, fmt.Errorf("get application directory: %w", err)
	}

	if opts != nil && opts.DBPath != "" {
		dbPath = opts.DBPath
	}

	vaultStore, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	has, err := vaultStore.HasSealedKey()
	if err != nil {
		_ = vaultStore.Close()
		return nil, fmt.Errorf("check sealed key: %w", err)
	}

	if !has {
		_ = vaultStore.Close()
		return nil, ErrNotInitialized
	}

	sk, err := vaultStore.GetSealedKey()
	if err != nil {
		_ = vaultStore.Close()
		return nil, fmt.Errorf("get sealed key: %w", err)
	}

	var masterKey []byte

	switch sk.SealMethod {
	case crypto.SealMethodTPM:
		var err error

		masterKey, err = crypto.UnsealMasterKeyTPM(sk.SealedData)
		if err != nil {
			_ = vaultStore.Close()
			return nil, fmt.Errorf("unseal master key (TPM): %w", err)
		}

	default: // "software" or legacy rows without seal_method
		machineIDHash, err := crypto.MachineIDHash()
		if err != nil {
			_ = vaultStore.Close()
			return nil, fmt.Errorf("hash machine ID: %w", err)
		}

		if !bytes.Equal(sk.MachineIDHash, machineIDHash) {
			_ = vaultStore.Close()
			return nil, ErrMachineMismatch
		}

		machineID, err := crypto.MachineID()
		if err != nil {
			_ = vaultStore.Close()
			return nil, fmt.Errorf("get machine ID: %w", err)
		}

		masterKey, err = crypto.UnsealMasterKey(sk.SealedData, sk.Nonce, machineID, sk.KeySalt)
		if err != nil {
			_ = vaultStore.Close()
			return nil, fmt.Errorf("unseal master key: %w", err)
		}
	}

	return &Vault{
		store:     vaultStore,
		masterKey: masterKey,
		dbPath:    dbPath,
	}, nil
}

// Close zeros the master key and closes the vault.
func (v *Vault) Close() error {
	crypto.ZeroBytes(v.masterKey)

	return v.store.Close()
}

// DBPath returns the database file path.
func (v *Vault) DBPath() string {
	return v.dbPath
}

// === Profile operations ===

// CreateProfile creates a new vault profile.
func (v *Vault) CreateProfile(name, description string, isDefault bool) error {
	exists, err := v.store.ProfileExists(name)
	if err != nil {
		return fmt.Errorf("check profile: %w", err)
	}

	if exists {
		return ErrProfileExists
	}

	return v.store.CreateProfile(name, description, isDefault)
}

// DeleteProfile deletes a profile and all its secrets.
func (v *Vault) DeleteProfile(name string) error {
	exists, err := v.store.ProfileExists(name)
	if err != nil {
		return fmt.Errorf("check profile: %w", err)
	}

	if !exists {
		return ErrProfileNotFound
	}

	return v.store.DeleteProfile(name)
}

// UseProfile sets a profile as the default.
func (v *Vault) UseProfile(name string) error {
	exists, err := v.store.ProfileExists(name)
	if err != nil {
		return fmt.Errorf("check profile: %w", err)
	}

	if !exists {
		return ErrProfileNotFound
	}

	return v.store.SetDefaultProfile(name)
}

// ListProfiles returns all vault profiles with secret counts.
func (v *Vault) ListProfiles() ([]ProfileInfo, error) {
	rows, err := v.store.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	profiles := make([]ProfileInfo, 0, len(rows))
	for _, row := range rows {
		count, err := v.store.CountSecrets(row.Name)
		if err != nil {
			return nil, fmt.Errorf("count secrets: %w", err)
		}

		info := ProfileInfo{
			Name:        row.Name,
			SecretCount: count,
			CreatedAt:   row.CreatedAt,
		}
		if row.Description != nil {
			info.Description = *row.Description
		}

		if row.IsDefault != nil {
			info.IsDefault = *row.IsDefault == 1
		}

		profiles = append(profiles, info)
	}

	return profiles, nil
}

// resolveProfile returns the profile name to use: explicit or default.
func (v *Vault) resolveProfile(profileName string) (string, error) {
	if profileName != "" {
		exists, err := v.store.ProfileExists(profileName)
		if err != nil {
			return "", fmt.Errorf("check profile: %w", err)
		}

		if !exists {
			return "", ErrProfileNotFound
		}

		return profileName, nil
	}

	row, err := v.store.GetDefaultProfile()
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoDefaultProfile
	}

	if err != nil {
		return "", fmt.Errorf("get default profile: %w", err)
	}

	return row.Name, nil
}

// === Secret operations ===

// Set encrypts and stores a secret in the given profile (or default).
func (v *Vault) Set(key, value, profileName, description string) error {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return err
	}

	ciphertext, nonce, err := crypto.Encrypt(v.masterKey, []byte(value))
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	return v.store.UpsertSecret(profile, key, ciphertext, nonce, description)
}

// Get decrypts and returns a secret value.
func (v *Vault) Get(key, profileName string) (string, error) {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return "", err
	}

	row, err := v.store.GetSecret(profile, key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrSecretNotFound
	}

	if err != nil {
		return "", fmt.Errorf("get secret: %w", err)
	}

	plaintext, err := crypto.Decrypt(v.masterKey, row.EncryptedValue, row.Nonce)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Delete removes a secret.
func (v *Vault) Delete(key, profileName string) error {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return err
	}

	exists, err := v.store.SecretExists(profile, key)
	if err != nil {
		return fmt.Errorf("check secret: %w", err)
	}

	if !exists {
		return ErrSecretNotFound
	}

	return v.store.DeleteSecret(profile, key)
}

// List returns metadata for all secrets in a profile.
func (v *Vault) List(profileName string) ([]SecretInfo, error) {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return nil, err
	}

	rows, err := v.store.ListSecrets(profile)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	secrets := make([]SecretInfo, 0, len(rows))
	for _, row := range rows {
		info := SecretInfo{
			Key:       row.Key,
			CreatedAt: row.CreatedAt,
		}
		if row.Description != nil {
			info.Description = *row.Description
		}

		if row.UpdatedAt != nil {
			info.UpdatedAt = *row.UpdatedAt
		}

		secrets = append(secrets, info)
	}

	return secrets, nil
}

// Export decrypts and returns all secrets from a profile.
func (v *Vault) Export(profileName string) ([]SecretEntry, error) {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return nil, err
	}

	rows, err := v.store.ListSecrets(profile)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	entries := make([]SecretEntry, 0, len(rows))
	for _, row := range rows {
		plaintext, err := crypto.Decrypt(v.masterKey, row.EncryptedValue, row.Nonce)
		if err != nil {
			return nil, fmt.Errorf("decrypt %q: %w", row.Key, err)
		}

		entry := SecretEntry{
			Key:   row.Key,
			Value: string(plaintext),
		}
		if row.Description != nil {
			entry.Description = *row.Description
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Import encrypts and stores multiple secrets into a profile.
func (v *Vault) Import(entries []SecretEntry, profileName string) error {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		ciphertext, nonce, err := crypto.Encrypt(v.masterKey, []byte(entry.Value))
		if err != nil {
			return fmt.Errorf("encrypt %q: %w", entry.Key, err)
		}

		if err := v.store.UpsertSecret(profile, entry.Key, ciphertext, nonce, entry.Description); err != nil {
			return fmt.Errorf("save %q: %w", entry.Key, err)
		}
	}

	return nil
}

// Status returns vault status information.
func (v *Vault) Status() (*Status, error) {
	profiles, err := v.store.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	var totalSecrets int64

	for _, p := range profiles {
		count, err := v.store.CountSecrets(p.Name)
		if err != nil {
			return nil, fmt.Errorf("count secrets: %w", err)
		}

		totalSecrets += count
	}

	sk, err := v.store.GetSealedKey()
	if err != nil {
		return nil, fmt.Errorf("get sealed key: %w", err)
	}

	var keyVersion int64 = 1
	if sk.Version != nil {
		keyVersion = *sk.Version
	}

	passwordSet, err := v.store.HasPassword()
	if err != nil {
		return nil, fmt.Errorf("check password: %w", err)
	}

	return &Status{
		Initialized:  true,
		DBPath:       v.dbPath,
		ProfileCount: len(profiles),
		SecretCount:  totalSecrets,
		KeyVersion:   keyVersion,
		SealMethod:   sk.SealMethod,
		PasswordSet:  passwordSet,
		CreatedAt:    sk.CreatedAt,
	}, nil
}

// === Password operations ===

// SetPassword hashes and stores a password for env release operations.
func (v *Vault) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return v.store.UpsertPassword(hash)
}

// VerifyPassword checks if the provided password matches the stored hash.
func (v *Vault) VerifyPassword(password string) error {
	has, err := v.store.HasPassword()
	if err != nil {
		return fmt.Errorf("check password: %w", err)
	}

	if !has {
		return ErrPasswordNotSet
	}

	pw, err := v.store.GetPassword()
	if err != nil {
		return fmt.Errorf("get password: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(pw.PasswordHash, []byte(password)); err != nil {
		return ErrPasswordMismatch
	}

	return nil
}

// HasPassword returns whether a password has been set.
func (v *Vault) HasPassword() (bool, error) {
	return v.store.HasPassword()
}

// DeletePassword removes the stored password.
func (v *Vault) DeletePassword() error {
	return v.store.DeletePassword()
}

// MasterKey returns the master key for sentinel operations.
func (v *Vault) MasterKey() []byte {
	return v.masterKey
}

// GetStatus returns vault status without requiring Open (checks if initialized).
func GetStatus(opts *Options) (*Status, error) {
	dbPath, err := application.GetApplicationDirectory()
	if err != nil {
		return nil, fmt.Errorf("get application directory: %w", err)
	}

	if opts != nil && opts.DBPath != "" {
		dbPath = opts.DBPath
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return &Status{
			Initialized: false,
			DBPath:      dbPath,
		}, nil
	}

	v, err := Open(opts)
	if err != nil {
		return &Status{
			Initialized: false,
			DBPath:      dbPath,
		}, nil
	}

	defer func() { _ = v.Close() }()

	return v.Status()
}
