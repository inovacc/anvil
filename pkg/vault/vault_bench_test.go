package vault_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func benchVault(b *testing.B) *vault.Vault {
	b.Helper()
	dbPath := filepath.Join(b.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}
	if err := vault.Init(opts); err != nil {
		b.Fatalf("Init: %v", err)
	}
	v, err := vault.Open(opts)
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	b.Cleanup(func() { _ = v.Close() })
	return v
}

func BenchmarkVaultSetGet(b *testing.B) {
	v := benchVault(b)
	if err := v.CreateProfile("bench", "", true); err != nil {
		b.Fatalf("CreateProfile: %v", err)
	}

	b.Run("Set", func(b *testing.B) {
		for i := range b.N {
			key := fmt.Sprintf("KEY_%d", i)
			if err := v.Set(key, "benchmark-secret-value", "bench", ""); err != nil {
				b.Fatalf("Set: %v", err)
			}
		}
	})

	// Pre-populate a key for Get benchmark.
	if err := v.Set("BENCH_KEY", "benchmark-value", "bench", ""); err != nil {
		b.Fatalf("Set: %v", err)
	}

	b.Run("Get", func(b *testing.B) {
		for b.Loop() {
			_, err := v.Get("BENCH_KEY", "bench")
			if err != nil {
				b.Fatalf("Get: %v", err)
			}
		}
	})
}

func BenchmarkVaultList(b *testing.B) {
	v := benchVault(b)
	if err := v.CreateProfile("bench", "", true); err != nil {
		b.Fatalf("CreateProfile: %v", err)
	}

	// Populate 100 secrets.
	for i := range 100 {
		if err := v.Set(fmt.Sprintf("KEY_%03d", i), "value", "bench", ""); err != nil {
			b.Fatalf("Set: %v", err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := v.List("bench")
		if err != nil {
			b.Fatalf("List: %v", err)
		}
	}
}

func BenchmarkVaultExport(b *testing.B) {
	v := benchVault(b)
	if err := v.CreateProfile("bench", "", true); err != nil {
		b.Fatalf("CreateProfile: %v", err)
	}

	for i := range 50 {
		if err := v.Set(fmt.Sprintf("KEY_%03d", i), "secret-value", "bench", ""); err != nil {
			b.Fatalf("Set: %v", err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := v.Export("bench")
		if err != nil {
			b.Fatalf("Export: %v", err)
		}
	}
}
