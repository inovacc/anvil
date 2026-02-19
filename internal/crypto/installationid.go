package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

// InstallationID derives a deterministic installation identifier from the
// machine ID hash and sealed key data. The result is a 64-char lowercase
// hex string. Same machine + same sealed key = same ID.
func InstallationID(machineIDHash, sealedData []byte) string {
	h := sha256.New()
	h.Write(machineIDHash)
	h.Write(sealedData)
	return hex.EncodeToString(h.Sum(nil))
}
