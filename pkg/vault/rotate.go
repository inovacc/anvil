package vault

import (
	"context"
	"fmt"

	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/anvil/internal/sentinel"
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

	// Begin transaction for atomic re-encryption.
	tx, qtx, err := v.store.BeginTx()
	if err != nil {
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer v.store.EndTx()

	ctx := context.Background()

	// Fetch all secrets.
	secrets, err := qtx.ListAllSecrets(ctx)
	if err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("list secrets: %w", err)
	}

	// Re-encrypt each secret with the new key.
	for _, s := range secrets {
		plaintext, err := crypto.Decrypt(v.masterKey, s.EncryptedValue, s.Nonce)
		if err != nil {
			_ = tx.Rollback()
			crypto.ZeroBytes(newKey)
			return fmt.Errorf("decrypt secret %q: %w", s.Key, err)
		}

		newCiphertext, newNonce, err := crypto.Encrypt(newKey, plaintext)
		if err != nil {
			_ = tx.Rollback()
			crypto.ZeroBytes(newKey)
			return fmt.Errorf("re-encrypt secret %q: %w", s.Key, err)
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
			crypto.ZeroBytes(newKey)
			return fmt.Errorf("update secret %q: %w", s.Key, err)
		}
	}

	// Get current sealed key for version bump.
	sk, err := qtx.GetSealedKey(ctx)
	if err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("get sealed key: %w", err)
	}

	var newVersion int64 = 1
	if sk.Version != nil {
		newVersion = *sk.Version + 1
	}

	// Seal the new key.
	machineIDHash, err := crypto.MachineIDHash()
	if err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("machine ID hash: %w", err)
	}

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
				crypto.ZeroBytes(newKey)
				return fmt.Errorf("save sealed key: %w", err)
			}

			if err := tx.Commit(); err != nil {
				crypto.ZeroBytes(newKey)
				return fmt.Errorf("commit: %w", err)
			}

			crypto.ZeroBytes(v.masterKey)
			v.masterKey = newKey

			return nil
		}
	}

	// Software sealing fallback.
	machineID, err := crypto.MachineID()
	if err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("get machine ID: %w", err)
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("generate salt: %w", err)
	}

	sealed, nonce, err := crypto.SealMasterKey(newKey, machineID, salt)
	if err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("seal master key: %w", err)
	}

	if err := qtx.UpsertSealedKey(ctx, sqlc.UpsertSealedKeyParams{
		SealedData:    sealed,
		Nonce:         nonce,
		KeySalt:       salt,
		Version:       &newVersion,
		MachineIDHash: machineIDHash,
		SealMethod:    crypto.SealMethodSoftware,
	}); err != nil {
		_ = tx.Rollback()
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("save sealed key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		crypto.ZeroBytes(newKey)
		return fmt.Errorf("commit: %w", err)
	}

	crypto.ZeroBytes(v.masterKey)
	v.masterKey = newKey

	return nil
}
