package vault_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/inovacc/anvil/pkg/vault"
)

func initOpenWithPassword(t *testing.T) *vault.Vault {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "vault.db")
	opts := &vault.Options{DBPath: dbPath}
	if err := vault.Init(opts); err != nil {
		t.Fatalf("Init: %v", err)
	}
	v, err := vault.Open(opts)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	if err := v.SetPassword("testpass"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if err := v.CreateProfile("default", "test", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	return v
}

func TestEnvReleaseAndStatus(t *testing.T) {
	v := initOpenWithPassword(t)

	state, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{
		TTL: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	if !state.Active {
		t.Error("expected active release")
	}
	if state.ProfileName != "default" {
		t.Errorf("ProfileName = %q, want %q", state.ProfileName, "default")
	}
	if state.SessionID == "" {
		t.Error("expected non-empty session ID")
	}

	status, err := v.EnvStatus()
	if err != nil {
		t.Fatalf("EnvStatus: %v", err)
	}
	if !status.Active {
		t.Error("expected active status")
	}
}

func TestEnvReleaseWrongPassword(t *testing.T) {
	v := initOpenWithPassword(t)

	_, err := v.EnvRelease("wrongpass", nil)
	if !errors.Is(err, vault.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestEnvReleaseTTLBounds(t *testing.T) {
	v := initOpenWithPassword(t)

	_, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 30 * time.Second})
	if err == nil {
		t.Error("expected error for TTL < 1m")
	}

	_, err = v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 25 * time.Hour})
	if err == nil {
		t.Error("expected error for TTL > 24h")
	}
}

func TestEnvRevokeAndStatus(t *testing.T) {
	v := initOpenWithPassword(t)

	if _, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 5 * time.Minute}); err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	if err := v.EnvRevoke(); err != nil {
		t.Fatalf("EnvRevoke: %v", err)
	}

	status, err := v.EnvStatus()
	if err != nil {
		t.Fatalf("EnvStatus: %v", err)
	}
	if status.Active {
		t.Error("expected inactive after revoke")
	}
}

func TestEnvExportRequiresRelease(t *testing.T) {
	v := initOpenWithPassword(t)

	// Clean up any sentinel file from other tests.
	_ = v.EnvRevoke()

	_, err := v.EnvExport("")
	if !errors.Is(err, vault.ErrNotReleased) {
		t.Errorf("expected ErrNotReleased, got %v", err)
	}
}

func TestEnvExportWithRelease(t *testing.T) {
	v := initOpenWithPassword(t)

	if err := v.Set("API_KEY", "secret123", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if _, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 5 * time.Minute}); err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	entries, err := v.EnvExport("")
	if err != nil {
		t.Fatalf("EnvExport: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "API_KEY" || entries[0].Value != "secret123" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestEnvInlineGetRequiresRelease(t *testing.T) {
	v := initOpenWithPassword(t)

	_ = v.EnvRevoke()

	_, err := v.EnvInlineGet("KEY", "")
	if !errors.Is(err, vault.ErrNotReleased) {
		t.Errorf("expected ErrNotReleased, got %v", err)
	}
}

func TestEnvInlineGetWithRelease(t *testing.T) {
	v := initOpenWithPassword(t)

	if err := v.Set("MY_SECRET", "value42", "default", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if _, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 5 * time.Minute}); err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	val, err := v.EnvInlineGet("MY_SECRET", "")
	if err != nil {
		t.Fatalf("EnvInlineGet: %v", err)
	}
	if val != "value42" {
		t.Errorf("got %q, want %q", val, "value42")
	}
}

func TestEnvReleaseTTLExactBounds(t *testing.T) {
	v := initOpenWithPassword(t)

	// Exactly 1 minute — should succeed.
	state, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 1 * time.Minute})
	if err != nil {
		t.Fatalf("EnvRelease at min TTL: %v", err)
	}
	if !state.Active {
		t.Error("expected active release at min TTL")
	}

	// Revoke for clean state.
	_ = v.EnvRevoke()

	// Exactly 24 hours — should succeed.
	state, err = v.EnvRelease("testpass", &vault.EnvReleaseOptions{TTL: 24 * time.Hour})
	if err != nil {
		t.Fatalf("EnvRelease at max TTL: %v", err)
	}
	if !state.Active {
		t.Error("expected active release at max TTL")
	}
}

func TestEnvStatusNoSession(t *testing.T) {
	v := initOpenWithPassword(t)

	// Ensure no active session.
	_ = v.EnvRevoke()

	status, err := v.EnvStatus()
	if err != nil {
		t.Fatalf("EnvStatus: %v", err)
	}
	if status.Active {
		t.Error("expected inactive status when no session")
	}
}

func TestEnvReleaseWithExplicitProfile(t *testing.T) {
	v := initOpenWithPassword(t)

	if err := v.CreateProfile("staging", "staging env", false); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	state, err := v.EnvRelease("testpass", &vault.EnvReleaseOptions{
		ProfileName: "staging",
		TTL:         5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("EnvRelease: %v", err)
	}

	if state.ProfileName != "staging" {
		t.Errorf("ProfileName = %q, want %q", state.ProfileName, "staging")
	}
}
