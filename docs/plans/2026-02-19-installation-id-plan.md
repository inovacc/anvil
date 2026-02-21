# Installation ID Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a deterministic machine-bound installation ID derived from `SHA-256(machine_id_hash || sealed_data)`, exposed via `pkg/vault` and queryable via `anvil id`.

**Architecture:** Pure crypto function in `internal/crypto`, vault method in `pkg/vault`, CLI command in `cmd/anvil`. No schema changes — ID is computed from existing `vault_sealed_key` data.

**Tech Stack:** Go stdlib `crypto/sha256`, `encoding/hex`

---

### Task 1: Crypto — InstallationID function

**Files:**
- Create: `internal/crypto/installationid.go`
- Create: `internal/crypto/installationid_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/crypto/ -run TestInstallationID -v`
Expected: FAIL — `InstallationID` not defined

**Step 3: Write minimal implementation**

```go
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/crypto/ -run TestInstallationID -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/crypto/installationid.go internal/crypto/installationid_test.go
git commit -m "feat: add InstallationID crypto function"
```

---

### Task 2: Vault — InstallationID method and interface

**Files:**
- Create: `pkg/vault/identity.go`
- Modify: `pkg/vault/iface.go:109-129` (add interface + compile-time check)

**Step 1: Write the interface in iface.go**

Add after `VaultRecovery` (line 113):

```go
// VaultIdentity provides installation identity operations.
type VaultIdentity interface {
	InstallationID() (string, error)
}
```

Add compile-time check in the `var` block (after line 128):

```go
_ VaultIdentity = (*Vault)(nil)
```

**Step 2: Run build to verify it fails**

Run: `go build ./pkg/vault/`
Expected: FAIL — `Vault` does not implement `VaultIdentity`

**Step 3: Create identity.go**

```go
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
```

**Step 4: Run build to verify it passes**

Run: `go build ./pkg/vault/`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/vault/identity.go pkg/vault/iface.go
git commit -m "feat: add InstallationID to vault module"
```

---

### Task 3: CLI — `anvil id` command

**Files:**
- Create: `cmd/anvil/id.go`

**Step 1: Create the command**

```go
package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Show the machine-bound installation ID",
	Long:  "Display a deterministic identifier derived from the machine identity and vault sealed key.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return handleError(cmd, err)
		}
		defer func() { _ = v.Close() }()

		id, err := v.InstallationID()
		if err != nil {
			return handleError(cmd, err)
		}

		outputResult(cmd, struct {
			InstallationID string `json:"installation_id"`
		}{id}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
		})

		return nil
	},
}

func init() {
	registerCommand(idCmd)
}
```

**Step 2: Build and verify**

Run: `go build ./cmd/anvil/`
Expected: PASS

**Step 3: Commit**

```bash
git add cmd/anvil/id.go
git commit -m "feat: add anvil id command"
```

---

### Task 4: Documentation updates

**Files:**
- Modify: `CLAUDE.md` — add `id.go` to project structure, `VaultIdentity` to interfaces
- Modify: `README.md` — add `anvil id` to command reference if present

**Step 1: Update CLAUDE.md**

Add `id.go` entry under `cmd/anvil/` in project structure:
```
│   ├── id.go       # Machine-bound installation ID command
```

Add `VaultIdentity` to interfaces list in iface.go description:
```
VaultIdentity         Installation identity (InstallationID)
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add installation ID to project documentation"
```
