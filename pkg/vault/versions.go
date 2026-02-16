package vault

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inovacc/anvil/internal/crypto"
)

// DefaultVersionRetention is the default duration archived versions are kept.
const DefaultVersionRetention = 30 * 24 * time.Hour // 30 days

// archiveCurrentSecret saves the current value of a secret as a version with retention expiry.
func (v *Vault) archiveCurrentSecret(profileName, key string) error {
	exists, err := v.store.SecretExists(profileName, key)
	if err != nil {
		return fmt.Errorf("check secret: %w", err)
	}

	if !exists {
		return nil
	}

	row, err := v.store.GetSecret(profileName, key)
	if err != nil {
		return fmt.Errorf("get current secret: %w", err)
	}

	latestVersion, err := v.store.GetLatestVersionNumber(profileName, key)
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}

	expiresAt := time.Now().Add(DefaultVersionRetention)

	return v.store.InsertSecretVersion(profileName, key, latestVersion+1, row.EncryptedValue, row.Nonce, expiresAt)
}

// SecretHistory returns metadata for all archived versions of a secret.
// Values are never exposed — only version number, creation time, and expiry are returned.
func (v *Vault) SecretHistory(key, profileName string) ([]SecretVersion, error) {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return nil, err
	}

	exists, err := v.store.SecretExists(profile, key)
	if err != nil {
		return nil, fmt.Errorf("check secret: %w", err)
	}

	if !exists {
		return nil, ErrSecretNotFound
	}

	rows, err := v.store.ListSecretVersions(profile, key)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}

	versions := make([]SecretVersion, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, SecretVersion{
			Version:   row.Version,
			CreatedAt: row.CreatedAt,
		})
	}

	return versions, nil
}

// secretVersionsDecrypted returns decrypted version data for internal use (backup only).
// This is NOT exposed through any public API.
func (v *Vault) secretVersionsDecrypted(key, profileName string) ([]backupVersion, error) {
	rows, err := v.store.ListSecretVersions(profileName, key)
	if err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}

	versions := make([]backupVersion, 0, len(rows))
	for _, row := range rows {
		plaintext, err := crypto.Decrypt(v.masterKey, row.EncryptedValue, row.Nonce)
		if err != nil {
			return nil, fmt.Errorf("decrypt version %d: %w", row.Version, err)
		}

		versions = append(versions, backupVersion{
			Version: row.Version,
			Value:   string(plaintext),
		})
	}

	return versions, nil
}

// backupVersion is an internal type for backup data only.
type backupVersion struct {
	Version int64
	Value   string
}

// SecretRollback restores a specific version as the current secret value.
// The current value is archived as a new version before restoring.
// This is the ONLY way to access an archived version's data.
func (v *Vault) SecretRollback(key, profileName string, version int64) error {
	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return err
	}

	versionRow, err := v.store.GetSecretVersion(profile, key, version)
	if errors.Is(err, sql.ErrNoRows) {
		return &UserError{
			Message: fmt.Sprintf("version %d not found for secret %q", version, key),
			Hint:    "Run 'anvil vault history <key>' to see available versions",
		}
	}

	if err != nil {
		return fmt.Errorf("get version: %w", err)
	}

	plaintext, err := crypto.Decrypt(v.masterKey, versionRow.EncryptedValue, versionRow.Nonce)
	if err != nil {
		return fmt.Errorf("decrypt version: %w", err)
	}

	if err := v.archiveCurrentSecret(profile, key); err != nil {
		return err
	}

	ciphertext, nonce, err := crypto.Encrypt(v.masterKey, plaintext)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	row, err := v.store.GetSecret(profile, key)
	if err != nil {
		return fmt.Errorf("get current secret: %w", err)
	}

	desc := ""
	if row.Description != nil {
		desc = *row.Description
	}

	if err := v.store.UpsertSecret(profile, key, ciphertext, nonce, desc); err != nil {
		return err
	}

	v.logAudit("secret.rollback", profile, key, fmt.Sprintf("rolled back to version %d", version))

	return nil
}

// PurgeExpiredVersions removes archived versions past their retention period.
func (v *Vault) PurgeExpiredVersions() (int64, error) {
	count, err := v.store.PurgeExpiredVersions()
	if err != nil {
		return 0, fmt.Errorf("purge expired versions: %w", err)
	}

	if count > 0 {
		v.logAudit("versions.purge", "", "", fmt.Sprintf("purged %d expired versions", count))
	}

	return count, nil
}
