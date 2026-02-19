package vault

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/inovacc/anvil/internal/crypto"
)

// Sign signs data with the named key and returns a base64-encoded signature.
func (v *Vault) Sign(keyName string, data []byte) (*SignResult, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	row, err := v.store.GetKeyByName(keyName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKeyNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get key: %w", err)
	}

	privKeyBytes, err := crypto.Decrypt(v.masterKey, row.EncryptedPrivateKey, row.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt private key: %w", err)
	}

	defer crypto.ZeroBytes(privKeyBytes)

	sig, err := crypto.Sign(privKeyBytes, crypto.KeyAlgorithm(row.Algorithm), data)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}

	v.logAudit("key.sign", "", keyName, fmt.Sprintf("data_len=%d", len(data)))

	return &SignResult{
		Signature: base64.StdEncoding.EncodeToString(sig),
		KeyName:   keyName,
		Algorithm: row.Algorithm,
	}, nil
}

// Verify verifies a base64-encoded signature against data using the named key.
func (v *Vault) Verify(keyName string, data []byte, signatureB64 string) (*VerifyResult, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	row, err := v.store.GetKeyByName(keyName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrKeyNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get key: %w", err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	valid, err := crypto.Verify(row.PublicKey, crypto.KeyAlgorithm(row.Algorithm), data, sigBytes)
	if err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}

	return &VerifyResult{
		Valid:     valid,
		KeyName:   keyName,
		Algorithm: row.Algorithm,
	}, nil
}
