package vault

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/sealbox"
)

// SharedExport represents encrypted secrets for machine-to-machine transfer.
type SharedExport struct {
	Version     int           `json:"version"`
	ProfileName string        `json:"profile_name"`
	CreatedAt   time.Time     `json:"created_at"`
	Secrets     []SecretEntry `json:"secrets"`
}

// ShareExport encrypts a profile's secrets with a passphrase for sharing.
func (v *Vault) ShareExport(profileName, passphrase string) ([]byte, error) {
	entries, err := v.Export(profileName)
	if err != nil {
		return nil, err
	}

	export := SharedExport{
		Version:     1,
		ProfileName: profileName,
		CreatedAt:   time.Now(),
		Secrets:     entries,
	}

	data, err := json.Marshal(export)
	if err != nil {
		return nil, fmt.Errorf("marshal export: %w", err)
	}

	key, err := crypto.DeriveKey(passphrase, []byte("anvil-share-v1"))
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	encrypted, err := sealbox.Encrypt(key, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt export: %w", err)
	}

	v.logAudit("vault.share.export", profileName, "", fmt.Sprintf("%d secrets", len(entries)))

	return encrypted, nil
}

// ShareImport decrypts a shared export and imports the secrets into a profile.
func (v *Vault) ShareImport(encrypted []byte, passphrase, targetProfile string) (*SharedExport, error) {
	key, err := crypto.DeriveKey(passphrase, []byte("anvil-share-v1"))
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	data, err := sealbox.Decrypt(key, encrypted)
	if err != nil {
		return nil, &UserError{
			Message: "failed to decrypt shared export",
			Hint:    "Check that the passphrase is correct",
		}
	}

	var export SharedExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("unmarshal export: %w", err)
	}

	// Use target profile if specified, otherwise use the original profile name.
	profile := targetProfile
	if profile == "" {
		profile = export.ProfileName
	}

	if err := v.Import(export.Secrets, profile); err != nil {
		return nil, err
	}

	v.logAudit("vault.share.import", profile, "", fmt.Sprintf("%d secrets from %q", len(export.Secrets), export.ProfileName))

	return &export, nil
}
