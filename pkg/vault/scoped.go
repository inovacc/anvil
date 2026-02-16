package vault

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ScopedVault provides isolated, single-profile access to the vault.
// External services use this — they can only see masked secret values
// and write/update secrets within their own profile. Plaintext read
// is denied (RBAC enforcement).
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

// Get returns the masked value of a secret. Scoped access never exposes plaintext.
func (s *ScopedVault) Get(key string) (string, error) {
	val, err := s.vault.Get(key, s.profileName)
	if err != nil {
		return "", err
	}

	return MaskValue(val), nil
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

// Export is denied for scoped access — services cannot export plaintext secrets.
func (s *ScopedVault) Export() ([]SecretEntry, error) {
	return nil, ErrReadDenied
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

// MaskValue masks a secret value for display. Shows first and last characters
// with asterisks in between. Short values are fully masked.
func MaskValue(value string) string {
	n := len(value)
	switch {
	case n == 0:
		return "****"
	case n <= 4:
		return strings.Repeat("*", n)
	case n <= 8:
		return string(value[0]) + strings.Repeat("*", n-2) + string(value[n-1])
	default:
		return value[:3] + strings.Repeat("*", n-6) + value[n-3:]
	}
}
