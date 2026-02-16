package vault

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/sealbox"
)

// BackupData represents a full vault backup in plaintext (before encryption).
type BackupData struct {
	Version   int                  `json:"version"`
	CreatedAt time.Time            `json:"created_at"`
	Profiles  []BackupProfile      `json:"profiles"`
	Password  []byte               `json:"password_hash,omitempty"`
	Versions  []BackupSecretVer    `json:"secret_versions,omitempty"`
}

// BackupProfile represents a profile and its secrets in a backup.
type BackupProfile struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	IsDefault   bool          `json:"is_default"`
	Secrets     []SecretEntry `json:"secrets"`
}

// BackupSecretVer represents a versioned secret in a backup.
type BackupSecretVer struct {
	ProfileName string `json:"profile_name"`
	Key         string `json:"key"`
	Version     int64  `json:"version"`
	Value       string `json:"value"`
}

// Backup exports the entire vault as an encrypted archive.
// The archive is encrypted with a key derived from the provided password.
func (v *Vault) Backup(password string) ([]byte, error) {
	if err := v.VerifyPassword(password); err != nil {
		return nil, err
	}

	profiles, err := v.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	backup := BackupData{
		Version:   1,
		CreatedAt: time.Now(),
	}

	for _, p := range profiles {
		bp := BackupProfile{
			Name:        p.Name,
			Description: p.Description,
			IsDefault:   p.IsDefault,
		}

		entries, err := v.Export(p.Name)
		if err != nil {
			return nil, fmt.Errorf("export profile %q: %w", p.Name, err)
		}

		bp.Secrets = entries

		// Export version history for each secret (internal decrypted access for backup only).
		for _, e := range entries {
			versions, err := v.secretVersionsDecrypted(e.Key, p.Name)
			if err != nil {
				return nil, fmt.Errorf("history for %q/%q: %w", p.Name, e.Key, err)
			}

			for _, ver := range versions {
				backup.Versions = append(backup.Versions, BackupSecretVer{
					ProfileName: p.Name,
					Key:         e.Key,
					Version:     ver.Version,
					Value:       ver.Value,
				})
			}
		}

		backup.Profiles = append(backup.Profiles, bp)
	}

	// Include password hash so it can be restored.
	has, err := v.HasPassword()
	if err != nil {
		return nil, fmt.Errorf("check password: %w", err)
	}

	if has {
		pw, err := v.store.GetPassword()
		if err != nil {
			return nil, fmt.Errorf("get password: %w", err)
		}

		backup.Password = pw.PasswordHash
	}

	data, err := json.Marshal(backup)
	if err != nil {
		return nil, fmt.Errorf("marshal backup: %w", err)
	}

	// Derive encryption key from password.
	key, err := crypto.DeriveKey(password, []byte("anvil-backup-v1"))
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	encrypted, err := sealbox.Encrypt(key, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt backup: %w", err)
	}

	v.logAudit("vault.backup", "", "", fmt.Sprintf("%d profiles", len(profiles)))

	return encrypted, nil
}

// Restore imports a backup archive into the vault, replacing all existing data.
// The vault must be initialized but empty (no profiles).
func (v *Vault) Restore(encrypted []byte, password string) error {
	key, err := crypto.DeriveKey(password, []byte("anvil-backup-v1"))
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}

	data, err := sealbox.Decrypt(key, encrypted)
	if err != nil {
		return &UserError{
			Message: "failed to decrypt backup",
			Hint:    "Check that the password is correct",
		}
	}

	var backup BackupData
	if err := json.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("unmarshal backup: %w", err)
	}

	// Restore profiles and secrets.
	for _, bp := range backup.Profiles {
		if err := v.CreateProfile(bp.Name, bp.Description, bp.IsDefault); err != nil {
			return fmt.Errorf("create profile %q: %w", bp.Name, err)
		}

		if err := v.Import(bp.Secrets, bp.Name); err != nil {
			return fmt.Errorf("import secrets for %q: %w", bp.Name, err)
		}
	}

	// Restore version history.
	for _, ver := range backup.Versions {
		ciphertext, nonce, err := crypto.Encrypt(v.masterKey, []byte(ver.Value))
		if err != nil {
			return fmt.Errorf("encrypt version %d of %q: %w", ver.Version, ver.Key, err)
		}

		expiresAt := time.Now().Add(DefaultVersionRetention)
		if err := v.store.InsertSecretVersion(ver.ProfileName, ver.Key, ver.Version, ciphertext, nonce, expiresAt); err != nil {
			return fmt.Errorf("insert version %d of %q: %w", ver.Version, ver.Key, err)
		}
	}

	// Restore password hash.
	if len(backup.Password) > 0 {
		if err := v.store.UpsertPassword(backup.Password); err != nil {
			return fmt.Errorf("restore password: %w", err)
		}
	}

	v.logAudit("vault.restore", "", "", fmt.Sprintf("%d profiles", len(backup.Profiles)))

	return nil
}
