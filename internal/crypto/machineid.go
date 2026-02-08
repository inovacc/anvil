package crypto

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

// MachineID returns a stable machine identifier.
// It tries the OS-specific method first, then falls back to hostname.
func MachineID() (string, error) {
	id, err := platformMachineID()
	if err == nil && id != "" {
		return id, nil
	}

	return fallbackID()
}

// MachineIDHash returns a SHA-256 hash of the machine ID.
func MachineIDHash() ([]byte, error) {
	id, err := MachineID()
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256([]byte(id))
	return h[:], nil
}

func fallbackID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	return strings.TrimSpace(hostname), nil
}
