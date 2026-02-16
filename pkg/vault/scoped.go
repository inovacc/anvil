package vault

import (
	"database/sql"
	"errors"
	"fmt"
)

// ScopedVault provides isolated, single-profile access to the vault.
// External apps use this — they can only read/write secrets in their own profile.
type ScopedVault struct {
	vault       *Vault
	profileUUID string
	profileName string
}

// Compile-time interface check.
var _ VaultScoped = (*ScopedVault)(nil)

// OpenScoped opens a vault scoped to a single profile identified by UUID.
func OpenScoped(profileUUID string, opts *Options) (*ScopedVault, error) {
	v, err := Open(opts)
	if err != nil {
		return nil, err
	}

	profile, err := v.storeRef().GetProfileByUUID(profileUUID)
	if errors.Is(err, sql.ErrNoRows) {
		_ = v.Close()
		return nil, ErrProfileUUIDNotFound
	}

	if err != nil {
		_ = v.Close()
		return nil, fmt.Errorf("get profile by UUID: %w", err)
	}

	return &ScopedVault{
		vault:       v,
		profileUUID: profileUUID,
		profileName: profile.Name,
	}, nil
}

// Get decrypts and returns a secret value from the scoped profile.
func (s *ScopedVault) Get(key string) (string, error) {
	return s.vault.Get(key, s.profileName)
}

// Set encrypts and stores a secret in the scoped profile.
func (s *ScopedVault) Set(key, value, description string) error {
	return s.vault.Set(key, value, s.profileName, description)
}

// Delete removes a secret from the scoped profile.
func (s *ScopedVault) Delete(key string) error {
	return s.vault.Delete(key, s.profileName)
}

// List returns metadata for all secrets in the scoped profile.
func (s *ScopedVault) List() ([]SecretInfo, error) {
	return s.vault.List(s.profileName)
}

// Export decrypts and returns all secrets from the scoped profile.
func (s *ScopedVault) Export() ([]SecretEntry, error) {
	return s.vault.Export(s.profileName)
}

// Import encrypts and stores multiple secrets into the scoped profile.
func (s *ScopedVault) Import(entries []SecretEntry) error {
	return s.vault.Import(entries, s.profileName)
}

// ProfileInfo returns metadata about the scoped profile.
func (s *ScopedVault) ProfileInfo() ProfileInfo {
	count, _ := s.vault.storeRef().CountSecrets(s.profileName)

	return ProfileInfo{
		Name:        s.profileName,
		UUID:        s.profileUUID,
		SecretCount: count,
	}
}

// Close zeros the master key and closes the vault.
func (s *ScopedVault) Close() error {
	return s.vault.Close()
}
