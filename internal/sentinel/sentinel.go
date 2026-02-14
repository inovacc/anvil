package sentinel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/inovacc/anvil/internal/application"
	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/sealbox"
)

const (
	enabledFile  = "PROFILE_ENV_RELEASE_ENABLED"
	disabledFile = "PROFILE_ENV_RELEASE_DISABLED"
)

var (
	ErrExpired  = errors.New("release session has expired")
	ErrNoActive = errors.New("no active release session")
)

// SessionPayload is the encrypted data stored in the sentinel file.
type SessionPayload struct {
	ProfileName string    `json:"profile_name"`
	ExpiresAt   time.Time `json:"expires_at"`
	SessionID   string    `json:"session_id"`
	MachineHash string    `json:"machine_hash"`
	CreatedAt   time.Time `json:"created_at"`
}

// ReleaseState represents the current state of an env release.
type ReleaseState struct {
	Active      bool
	ProfileName string
	ExpiresAt   time.Time
	SessionID   string
	Remaining   time.Duration
}

// cacheDir is a function variable to allow overriding in tests.
var cacheDir = defaultCacheDir

func defaultCacheDir() (string, error) {
	return application.GetApplicationDirectory()
}

func enabledPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, enabledFile), nil
}

func disabledPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, disabledFile), nil
}

// Release creates a new env release session. The sentinel file is encrypted
// with the vault master key, so it is opaque without vault access.
func Release(masterKey []byte, profileName string, ttl time.Duration) (*ReleaseState, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	machineHash, err := crypto.MachineIDHash()
	if err != nil {
		return nil, fmt.Errorf("get machine hash: %w", err)
	}

	sessionID := fmt.Sprintf("%x", machineHash[:8])
	now := time.Now()

	payload := SessionPayload{
		ProfileName: profileName,
		ExpiresAt:   now.Add(ttl),
		SessionID:   sessionID,
		MachineHash: fmt.Sprintf("%x", machineHash),
		CreatedAt:   now,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// sealbox.Encrypt packs nonce (12 bytes) || ciphertext into a single blob.
	fileData, err := sealbox.Encrypt(masterKey, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt payload: %w", err)
	}

	ep, err := enabledPath()
	if err != nil {
		return nil, err
	}

	// Write to a temp file, then rename it for atomicity
	tmpFile := fmt.Sprintf("%s.tmp", ep)
	if err := os.WriteFile(tmpFile, fileData, 0o600); err != nil {
		return nil, fmt.Errorf("write sentinel: %w", err)
	}

	if err := os.Rename(tmpFile, ep); err != nil {
		_ = os.Remove(tmpFile)
		return nil, fmt.Errorf("rename sentinel: %w", err)
	}

	// Remove a disabled file if it exists
	dp, err := disabledPath()
	if err != nil {
		return nil, err
	}

	_ = os.Remove(dp)

	return &ReleaseState{
		Active:      true,
		ProfileName: profileName,
		ExpiresAt:   payload.ExpiresAt,
		SessionID:   sessionID,
		Remaining:   time.Until(payload.ExpiresAt),
	}, nil
}

// Revoke revokes the active release by renaming the enabled file to disable.
func Revoke() error {
	ep, err := enabledPath()
	if err != nil {
		return err
	}

	dp, err := disabledPath()
	if err != nil {
		return err
	}

	if err := os.Rename(ep, dp); err != nil {
		if os.IsNotExist(err) {
			return ErrNoActive
		}

		return fmt.Errorf("revoke sentinel: %w", err)
	}

	return nil
}

// Check reads and validates the sentinel file. If expired, it auto-revokes.
func Check(masterKey []byte) (*ReleaseState, error) {
	ep, err := enabledPath()
	if err != nil {
		return nil, err
	}

	fileData, err := os.ReadFile(ep)
	if err != nil {
		if os.IsNotExist(err) {
			return &ReleaseState{Active: false}, nil
		}

		return nil, fmt.Errorf("read sentinel: %w", err)
	}

	// sealbox.Decrypt expects nonce (12 bytes) || ciphertext.
	if len(fileData) < sealbox.NonceSize {
		return &ReleaseState{Active: false}, nil
	}

	data, err := sealbox.Decrypt(masterKey, fileData)
	if err != nil {
		return nil, fmt.Errorf("decrypt sentinel: %w", err)
	}

	var payload SessionPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	remaining := time.Until(payload.ExpiresAt)
	if remaining <= 0 {
		// Auto-revoke an expired session
		_ = Revoke()

		return &ReleaseState{
			Active:      false,
			ProfileName: payload.ProfileName,
			ExpiresAt:   payload.ExpiresAt,
			SessionID:   payload.SessionID,
		}, nil
	}

	return &ReleaseState{
		Active:      true,
		ProfileName: payload.ProfileName,
		ExpiresAt:   payload.ExpiresAt,
		SessionID:   payload.SessionID,
		Remaining:   remaining,
	}, nil
}

// IsReleased returns true if there is an active, non-expired release session.
func IsReleased(masterKey []byte) (bool, error) {
	state, err := Check(masterKey)
	if err != nil {
		return false, err
	}

	return state.Active, nil
}
