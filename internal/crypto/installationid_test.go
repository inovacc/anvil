package crypto

import (
	"testing"
)

func TestInstallationID_Deterministic(t *testing.T) {
	machineIDHash := []byte("test-machine-id-hash-value-32byt")
	sealedData := []byte("test-sealed-data-blob")

	id1 := InstallationID(machineIDHash, sealedData)
	id2 := InstallationID(machineIDHash, sealedData)

	if id1 != id2 {
		t.Errorf("expected deterministic output, got %s and %s", id1, id2)
	}

	if len(id1) != 64 {
		t.Errorf("expected 64-char hex string, got %d chars: %s", len(id1), id1)
	}
}

func TestInstallationID_DifferentInputs(t *testing.T) {
	machineA := []byte("machine-a-hash-value-000000032bt")
	machineB := []byte("machine-b-hash-value-000000032bt")
	sealed := []byte("same-sealed-data")

	idA := InstallationID(machineA, sealed)
	idB := InstallationID(machineB, sealed)

	if idA == idB {
		t.Error("expected different IDs for different machine hashes")
	}
}
