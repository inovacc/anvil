package crypto

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func platformMachineID() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("failed to open registry key: %w", err)
	}
	defer func() { _ = k.Close() }()

	val, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", fmt.Errorf("failed to read MachineGuid: %w", err)
	}

	return strings.TrimSpace(val), nil
}
