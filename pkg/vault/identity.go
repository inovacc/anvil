package vault

import (
	"fmt"

	"github.com/inovacc/anvil/internal/crypto"
)

// InstallationID returns a deterministic machine-bound identifier for this
// vault installation. It is derived from SHA-256(machine_id_hash || sealed_data).
func (v *Vault) InstallationID() (string, error) {
	if err := v.checkSealed(); err != nil {
		return "", err
	}

	sk, err := v.store.GetSealedKey()
	if err != nil {
		return "", fmt.Errorf("get sealed key: %w", err)
	}

	return crypto.InstallationID(sk.MachineIDHash, sk.SealedData), nil
}
