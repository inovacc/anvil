package vault

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/inovacc/anvil/internal/crypto"
)

// GenerateKey generates a new asymmetric key pair and stores it encrypted in the vault.
func (v *Vault) GenerateKey(name string, algorithm KeyAlgorithm, description string) (*KeyInfo, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	if err := validateAlgorithm(algorithm); err != nil {
		return nil, err
	}

	exists, err := v.store.KeyExists(name, algorithm)
	if err != nil {
		return nil, fmt.Errorf("check key: %w", err)
	}

	if exists {
		return nil, ErrKeyAlreadyExists
	}

	kp, err := crypto.GenerateKeyPair(crypto.KeyAlgorithm(algorithm))
	if err != nil {
		return nil, fmt.Errorf("generate key pair: %w", err)
	}

	defer crypto.ZeroBytes(kp.PrivateKey)

	encPrivKey, nonce, err := crypto.Encrypt(v.masterKey, kp.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt private key: %w", err)
	}

	fingerprint := crypto.Fingerprint(kp.PublicKey)

	if err := v.store.InsertKey(name, algorithm, encPrivKey, nonce, kp.PublicKey, fingerprint, description); err != nil {
		return nil, fmt.Errorf("store key: %w", err)
	}

	v.logAudit("key.generate", "", name, fmt.Sprintf("algorithm=%s fingerprint=%s", algorithm, fingerprint))

	return &KeyInfo{
		Name:        name,
		Algorithm:   algorithm,
		Fingerprint: fingerprint,
		Description: description,
	}, nil
}

// ListKeys returns metadata for all stored asymmetric keys.
func (v *Vault) ListKeys() ([]KeyInfo, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	rows, err := v.store.ListKeys()
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	keys := make([]KeyInfo, 0, len(rows))
	for _, row := range rows {
		info := KeyInfo{
			Name:        row.Name,
			Algorithm:   row.Algorithm,
			Fingerprint: row.Fingerprint,
			CreatedAt:   row.CreatedAt,
		}
		if row.Description != nil {
			info.Description = *row.Description
		}

		keys = append(keys, info)
	}

	return keys, nil
}

// DeleteKey deletes an asymmetric key. If algorithm is empty, deletes all keys with that name.
func (v *Vault) DeleteKey(name, algorithm string) error {
	if err := v.checkSealed(); err != nil {
		return err
	}

	if algorithm != "" {
		exists, err := v.store.KeyExists(name, algorithm)
		if err != nil {
			return fmt.Errorf("check key: %w", err)
		}

		if !exists {
			return ErrKeyNotFound
		}

		if err := v.store.DeleteKey(name, algorithm); err != nil {
			return fmt.Errorf("delete key: %w", err)
		}
	} else {
		exists, err := v.store.KeyExistsByName(name)
		if err != nil {
			return fmt.Errorf("check key: %w", err)
		}

		if !exists {
			return ErrKeyNotFound
		}

		if err := v.store.DeleteKeyByName(name); err != nil {
			return fmt.Errorf("delete key: %w", err)
		}
	}

	v.logAudit("key.delete", "", name, algorithm)

	return nil
}

// ExportKeyPEM exports a key in PEM format. If private is true, exports the private key (decrypted).
func (v *Vault) ExportKeyPEM(name string, private bool) ([]byte, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	row, err := v.store.GetKeyByName(name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKeyNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get key: %w", err)
	}

	if !private {
		return crypto.MarshalPublicKeyPEM(row.PublicKey), nil
	}

	privKeyBytes, err := crypto.Decrypt(v.masterKey, row.EncryptedPrivateKey, row.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt private key: %w", err)
	}

	defer crypto.ZeroBytes(privKeyBytes)

	v.logAudit("key.export", "", name, "private=true")

	return crypto.MarshalPrivateKeyPEM(privKeyBytes), nil
}

// ImportKeyPEM imports a PEM-encoded key into the vault.
func (v *Vault) ImportKeyPEM(name string, pemData []byte, description string) (*KeyInfo, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	keyBytes, isPrivate, algorithm, err := crypto.ParsePEMKey(pemData)
	if err != nil {
		return nil, fmt.Errorf("parse PEM: %w", err)
	}

	algStr := KeyAlgorithm(algorithm)
	if err := validateAlgorithm(algStr); err != nil {
		return nil, err
	}

	exists, err := v.store.KeyExists(name, algStr)
	if err != nil {
		return nil, fmt.Errorf("check key: %w", err)
	}

	if exists {
		return nil, ErrKeyAlreadyExists
	}

	var privKeyBytes, pubKeyBytes []byte

	if isPrivate {
		privKeyBytes = keyBytes

		pubKeyBytes, err = crypto.ExtractPublicKeyFromPrivate(keyBytes, algorithm)
		if err != nil {
			return nil, fmt.Errorf("extract public key: %w", err)
		}
	} else {
		return nil, &UserError{
			Message: "cannot import public key only",
			Hint:    "Import a PEM file containing a private key (PRIVATE KEY block)",
		}
	}

	defer crypto.ZeroBytes(privKeyBytes)

	encPrivKey, nonce, err := crypto.Encrypt(v.masterKey, privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt private key: %w", err)
	}

	fingerprint := crypto.Fingerprint(pubKeyBytes)

	if err := v.store.InsertKey(name, algStr, encPrivKey, nonce, pubKeyBytes, fingerprint, description); err != nil {
		return nil, fmt.Errorf("store key: %w", err)
	}

	v.logAudit("key.import", "", name, fmt.Sprintf("algorithm=%s fingerprint=%s", algStr, fingerprint))

	return &KeyInfo{
		Name:        name,
		Algorithm:   algStr,
		Fingerprint: fingerprint,
		Description: description,
	}, nil
}

func validateAlgorithm(algorithm KeyAlgorithm) error {
	switch algorithm {
	case AlgorithmEd25519, AlgorithmECDSAP256:
		return nil
	default:
		return ErrUnsupportedAlgorithm
	}
}
