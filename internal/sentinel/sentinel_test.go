package sentinel

import (
	"os"
	"testing"
	"time"

	"github.com/inovacc/anvil/internal/crypto"
)

func setupTestDir(t *testing.T) (cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()

	origCacheDir := cacheDir
	cacheDir = func() (string, error) {
		return tmpDir, nil
	}

	return func() {
		cacheDir = origCacheDir
	}
}

func testKey(t *testing.T) []byte {
	t.Helper()

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	return key
}

func TestRelease(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		ttl     time.Duration
	}{
		{"short TTL", "dev", 1 * time.Minute},
		{"medium TTL", "staging", 30 * time.Minute},
		{"long TTL", "production", 24 * time.Hour},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupTestDir(t)
			defer cleanup()

			key := testKey(t)

			state, err := Release(key, tc.profile, tc.ttl)
			if err != nil {
				t.Fatalf("Release() error = %v", err)
			}

			if !state.Active {
				t.Error("expected Active to be true")
			}

			if state.ProfileName != tc.profile {
				t.Errorf("ProfileName = %q, want %q", state.ProfileName, tc.profile)
			}

			if state.SessionID == "" {
				t.Error("expected non-empty SessionID")
			}

			if state.Remaining <= 0 {
				t.Error("expected positive Remaining duration")
			}
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, key []byte)
		wantActive bool
		wantErr    bool
	}{
		{
			name:       "no sentinel file",
			setup:      func(t *testing.T, key []byte) {},
			wantActive: false,
		},
		{
			name: "active release",
			setup: func(t *testing.T, key []byte) {
				if _, err := Release(key, "test", 10*time.Minute); err != nil {
					t.Fatalf("Release() error = %v", err)
				}
			},
			wantActive: true,
		},
		{
			name: "expired release",
			setup: func(t *testing.T, key []byte) {
				if _, err := Release(key, "test", 1*time.Millisecond); err != nil {
					t.Fatalf("Release() error = %v", err)
				}

				time.Sleep(5 * time.Millisecond)
			},
			wantActive: false,
		},
		{
			name: "wrong key",
			setup: func(t *testing.T, key []byte) {
				if _, err := Release(key, "test", 10*time.Minute); err != nil {
					t.Fatalf("Release() error = %v", err)
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupTestDir(t)
			defer cleanup()

			key := testKey(t)
			tc.setup(t, key)

			checkKey := key
			if tc.wantErr {
				checkKey = testKey(t) // different key
			}

			state, err := Check(checkKey)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}

			if state.Active != tc.wantActive {
				t.Errorf("Active = %v, want %v", state.Active, tc.wantActive)
			}
		})
	}
}

func TestRevoke(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, key []byte)
		wantErr error
	}{
		{
			name:    "no active session",
			setup:   func(t *testing.T, key []byte) {},
			wantErr: ErrNoActive,
		},
		{
			name: "revoke active session",
			setup: func(t *testing.T, key []byte) {
				if _, err := Release(key, "test", 10*time.Minute); err != nil {
					t.Fatalf("Release() error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupTestDir(t)
			defer cleanup()

			key := testKey(t)
			tc.setup(t, key)

			err := Revoke()
			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("Revoke() error = %v", err)
			}

			// After revoke, Check should show inactive
			state, err := Check(key)
			if err != nil {
				t.Fatalf("Check() after revoke error = %v", err)
			}

			if state.Active {
				t.Error("expected Active to be false after revoke")
			}
		})
	}
}

func TestIsReleased(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	key := testKey(t)

	// Initially not released
	released, err := IsReleased(key)
	if err != nil {
		t.Fatalf("IsReleased() error = %v", err)
	}

	if released {
		t.Error("expected not released initially")
	}

	// Release
	if _, err := Release(key, "test", 10*time.Minute); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	released, err = IsReleased(key)
	if err != nil {
		t.Fatalf("IsReleased() error = %v", err)
	}

	if !released {
		t.Error("expected released after Release()")
	}

	// Revoke
	if err := Revoke(); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	released, err = IsReleased(key)
	if err != nil {
		t.Fatalf("IsReleased() error = %v", err)
	}

	if released {
		t.Error("expected not released after Revoke()")
	}
}

func TestReleaseOverwrite(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	key := testKey(t)

	// Release for "dev"
	state1, err := Release(key, "dev", 10*time.Minute)
	if err != nil {
		t.Fatalf("first Release() error = %v", err)
	}

	// Release again for "prod" — should overwrite
	state2, err := Release(key, "prod", 30*time.Minute)
	if err != nil {
		t.Fatalf("second Release() error = %v", err)
	}

	if state1.ProfileName == state2.ProfileName {
		t.Error("expected different profile names")
	}

	// Check should show "prod"
	state, err := Check(key)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if state.ProfileName != "prod" {
		t.Errorf("ProfileName = %q, want %q", state.ProfileName, "prod")
	}
}

func TestDisabledFileCleaned(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	key := testKey(t)

	// Release, then revoke (creates disabled file)
	if _, err := Release(key, "test", 10*time.Minute); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	if err := Revoke(); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	dp, _ := disabledPath()
	if _, err := os.Stat(dp); os.IsNotExist(err) {
		t.Fatal("expected disabled file to exist after revoke")
	}

	// New release should remove disabled file
	if _, err := Release(key, "test", 10*time.Minute); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	if _, err := os.Stat(dp); !os.IsNotExist(err) {
		t.Error("expected disabled file to be removed after new release")
	}
}

func TestRevokeAlreadyRevoked(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	key := testKey(t)

	if _, err := Release(key, "test", 10*time.Minute); err != nil {
		t.Fatalf("Release: %v", err)
	}

	if err := Revoke(); err != nil {
		t.Fatalf("first Revoke: %v", err)
	}

	// Second revoke should return ErrNoActive.
	err := Revoke()
	if err != ErrNoActive {
		t.Errorf("expected ErrNoActive, got %v", err)
	}
}

func TestCheckShortFile(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	// Write a file too short to be valid.
	ep, err := enabledPath()
	if err != nil {
		t.Fatalf("enabledPath: %v", err)
	}

	if err := os.WriteFile(ep, []byte("short"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	key := testKey(t)

	state, err := Check(key)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if state.Active {
		t.Error("expected inactive for short file")
	}
}

func TestCheckExpiredAutoRevokes(t *testing.T) {
	cleanup := setupTestDir(t)
	defer cleanup()

	key := testKey(t)

	// Create a session with very short TTL.
	if _, err := Release(key, "test", 1*time.Millisecond); err != nil {
		t.Fatalf("Release: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	state, err := Check(key)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if state.Active {
		t.Error("expected inactive for expired session")
	}

	if state.ProfileName != "test" {
		t.Errorf("ProfileName = %q, want %q", state.ProfileName, "test")
	}

	// After auto-revoke, the enabled file should be gone (renamed to disabled).
	ep, _ := enabledPath()
	if _, err := os.Stat(ep); !os.IsNotExist(err) {
		t.Error("expected enabled file to be removed after auto-revoke")
	}
}
