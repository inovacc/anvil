package crypto

import (
	"fmt"
	"os"
	"strings"
)

func platformMachineID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		// Try D-Bus machine-id as fallback
		data, err = os.ReadFile("/var/lib/dbus/machine-id")
		if err != nil {
			return "", fmt.Errorf("failed to read machine-id: %w", err)
		}
	}

	return strings.TrimSpace(string(data)), nil
}
