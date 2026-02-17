package vault

import (
	"context"
	"fmt"

	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/anvil/internal/sentinel"
	"github.com/inovacc/anvil/internal/store"
	"github.com/inovacc/anvil/internal/store/sqlc"
)

// RotateKey generates a new master key, re-encrypts all secrets, and re-seals
// the new key. Any active sentinel session is revoked. Requires password verification.
func (v *Vault) RotateKey(password string) error {
	if err := v.VerifyPassword(password); err != nil {
		return err
	}

	// Revoke active sentinel session (ignore "no active" errors).
	if err := sentinel.Revoke(); err != nil && err != sentinel.ErrNoActive {
		return fmt.Errorf("revoke sentinel: %w", err)
	}

	// Generate new master key.
	newKey, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("generate new master key: %w", err)
	}

	newVersion, sealMethod, err := v.rotateKeyTx(newKey)
	if err != nil {
		crypto.ZeroBytes(newKey)
		return err
	}

	// Re-encrypt all app database secrets with the new key.
	if err := v.rotateAppSecrets(newKey); err != nil {
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("rotate app secrets: %w", err)
	}

	crypto.ZeroBytes(v.masterKey)
	v.masterKey = newKey

	v.logAudit("key.rotate", "", "", fmt.Sprintf("version %d, method %s", newVersion, sealMethod))

	return nil
}

func (v *Vault) rotateKeyTx(newKey []byte) (int64, string, error) {
	tx, qtx, err := v.store.BeginTx()
	if err != nil {
		return 0, "", fmt.Errorf("begin transaction: %w", err)
	}

	defer v.store.EndTx()

	ctx := context.Background()

	// Fetch and re-encrypt all secrets.
	secrets, err := qtx.ListAllSecrets(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("list secrets: %w", err)
	}

	for _, s := range secrets {
		plaintext, err := crypto.Decrypt(v.masterKey, s.EncryptedValue, s.Nonce)
		if err != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("decrypt secret %q: %w", s.Key, err)
		}

		newCiphertext, newNonce, err := crypto.Encrypt(newKey, plaintext)
		if err != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("re-encrypt secret %q: %w", s.Key, err)
		}

		desc := ""
		if s.Description != nil {
			desc = *s.Description
		}

		if err := qtx.UpsertSecret(ctx, sqlc.UpsertSecretParams{
			ProfileName:    s.ProfileName,
			Key:            s.Key,
			EncryptedValue: newCiphertext,
			Nonce:          newNonce,
			Description:    &desc,
		}); err != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("update secret %q: %w", s.Key, err)
		}
	}

	// Re-encrypt all secret versions.
	secretVersions, err := qtx.ListAllSecretVersions(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("list secret versions: %w", err)
	}

	for _, sv := range secretVersions {
		plaintext, err := crypto.Decrypt(v.masterKey, sv.EncryptedValue, sv.Nonce)
		if err != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("decrypt version %d of %q: %w", sv.Version, sv.Key, err)
		}

		newCiphertext, newNonce, err := crypto.Encrypt(newKey, plaintext)
		if err != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("re-encrypt version %d of %q: %w", sv.Version, sv.Key, err)
		}

		if err := qtx.UpdateSecretVersionData(ctx, sqlc.UpdateSecretVersionDataParams{
			EncryptedValue: newCiphertext,
			Nonce:          newNonce,
			ID:             sv.ID,
		}); err != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("update version %d of %q: %w", sv.Version, sv.Key, err)
		}
	}

	// Get current sealed key for version bump.
	sk, err := qtx.GetSealedKey(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("get sealed key: %w", err)
	}

	var newVersion int64 = 1
	if sk.Version != nil {
		newVersion = *sk.Version + 1
	}

	// Seal the new key.
	machineIDHash, err := crypto.MachineIDHash()
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("machine ID hash: %w", err)
	}

	sealMethod := crypto.SealMethodSoftware

	if crypto.IsTPMAvailable() {
		sealedJSON, tpmErr := crypto.SealMasterKeyTPM(newKey)
		if tpmErr == nil {
			if err := qtx.UpsertSealedKey(ctx, sqlc.UpsertSealedKeyParams{
				SealedData:    sealedJSON,
				Nonce:         []byte{},
				KeySalt:       []byte{},
				Version:       &newVersion,
				MachineIDHash: machineIDHash,
				SealMethod:    crypto.SealMethodTPM,
			}); err != nil {
				_ = tx.Rollback()
				return 0, "", fmt.Errorf("save sealed key: %w", err)
			}

			if err := tx.Commit(); err != nil {
				return 0, "", fmt.Errorf("commit: %w", err)
			}

			return newVersion, crypto.SealMethodTPM, nil
		}
	}

	// Software sealing fallback.
	machineID, err := crypto.MachineID()
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("get machine ID: %w", err)
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("generate salt: %w", err)
	}

	sealed, nonce, err := crypto.SealMasterKey(newKey, machineID, salt)
	if err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("seal master key: %w", err)
	}

	if err := qtx.UpsertSealedKey(ctx, sqlc.UpsertSealedKeyParams{
		SealedData:    sealed,
		Nonce:         nonce,
		KeySalt:       salt,
		Version:       &newVersion,
		MachineIDHash: machineIDHash,
		SealMethod:    sealMethod,
	}); err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("save sealed key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, "", fmt.Errorf("commit: %w", err)
	}

	return newVersion, sealMethod, nil
}

// rotateAppSecrets re-encrypts all secrets in all registered app databases.
func (v *Vault) rotateAppSecrets(newKey []byte) error {
	apps, err := v.store.ListActiveApps()
	if err != nil {
		return fmt.Errorf("list active apps: %w", err)
	}

	for _, app := range apps {
		appStore, err := store.OpenApp(app.DbPath)
		if err != nil {
			return fmt.Errorf("open app %q: %w", app.Name, err)
		}

		secrets, err := appStore.ListAll()
		if err != nil {
			_ = appStore.Close()
			return fmt.Errorf("list secrets in app %q: %w", app.Name, err)
		}

		for _, sec := range secrets {
			plaintext, err := crypto.Decrypt(v.masterKey, sec.EncryptedValue, sec.Nonce)
			if err != nil {
				_ = appStore.Close()
				return fmt.Errorf("decrypt secret %q in app %q: %w", sec.Key, app.Name, err)
			}

			newCiphertext, newNonce, err := crypto.Encrypt(newKey, plaintext)
			if err != nil {
				_ = appStore.Close()
				return fmt.Errorf("re-encrypt secret %q in app %q: %w", sec.Key, app.Name, err)
			}

			if err := appStore.UpdateSecret(sec.ID, newCiphertext, newNonce); err != nil {
				_ = appStore.Close()
				return fmt.Errorf("update secret %q in app %q: %w", sec.Key, app.Name, err)
			}
		}

		_ = appStore.Close()
	}

	return nil
}
