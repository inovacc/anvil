package sentinel

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/inovacc/anvil/internal/crypto"
)

// ReleaseState represents an active env release session.
type ReleaseState struct {
	Active      bool          `json:"active"`
	ProfileName string        `json:"profile_name"`
	ExpiresAt   time.Time     `json:"expires_at"`
	SessionID   string        `json:"session_id"`
	Remaining   time.Duration `json:"remaining"`
}

// ErrNoActive is returned when there is no active release session.
var ErrNoActive = errors.New("no active release session")

const sentinelFile = "anvil-sentinel.enc"

// cacheDir returns the cache directory for sentinel files.
var cacheDir = os.UserCacheDir

type sentinelData struct {
	ProfileName string    `json:"profile_name"`
	ExpiresAt   time.Time `json:"expires_at"`
	SessionID   string    `json:"session_id"`
}

func sentinelPath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", fmt.Errorf("get cache dir: %w", err)
	}

	return filepath.Join(dir, sentinelFile), nil
}

// Release creates a new time-limited release session.
func Release(masterKey []byte, profileName string, ttl time.Duration) (*ReleaseState, error) {
	sessionID := fmt.Sprintf("%x", time.Now().UnixNano())

	data := sentinelData{
		ProfileName: profileName,
		ExpiresAt:   time.Now().Add(ttl),
		SessionID:   sessionID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal sentinel: %w", err)
	}

	ciphertext, nonce, err := crypto.Encrypt(masterKey, jsonData)
	if err != nil {
		return nil, fmt.Errorf("encrypt sentinel: %w", err)
	}

	path, err := sentinelPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	payload, err := json.Marshal(map[string][]byte{"ciphertext": ciphertext, "nonce": nonce})
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// Remove disabled marker if present.
	dp, _ := disabledPath()
	_ = os.Remove(dp)

	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return nil, fmt.Errorf("write sentinel: %w", err)
	}

	remaining := time.Until(data.ExpiresAt)

	return &ReleaseState{
		Active:      true,
		ProfileName: profileName,
		ExpiresAt:   data.ExpiresAt,
		SessionID:   sessionID,
		Remaining:   remaining,
	}, nil
}

// Check reads the sentinel file and checks if the session is still valid.
func Check(masterKey []byte) (*ReleaseState, error) {
	path, err := sentinelPath()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &ReleaseState{Active: false}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read sentinel: %w", err)
	}

	var payload map[string][]byte
	if err := json.Unmarshal(raw, &payload); err != nil {
		// Corrupted or short file — treat as inactive.
		_ = os.Remove(path)
		return &ReleaseState{Active: false}, nil
	}

	plaintext, err := crypto.Decrypt(masterKey, payload["ciphertext"], payload["nonce"])
	if err != nil {
		return nil, fmt.Errorf("decrypt sentinel: %w", err)
	}

	var data sentinelData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sentinel: %w", err)
	}

	if time.Now().After(data.ExpiresAt) {
		dp, _ := disabledPath()
		_ = os.Rename(path, dp)
		return &ReleaseState{Active: false, ProfileName: data.ProfileName}, nil
	}

	remaining := time.Until(data.ExpiresAt)

	return &ReleaseState{
		Active:      true,
		ProfileName: data.ProfileName,
		ExpiresAt:   data.ExpiresAt,
		SessionID:   data.SessionID,
		Remaining:   remaining,
	}, nil
}

// IsReleased checks if an active (non-expired) release session exists.
func IsReleased(masterKey []byte) (bool, error) {
	state, err := Check(masterKey)
	if err != nil {
		return false, err
	}

	return state.Active, nil
}

// enabledPath returns the path to the active sentinel file.
func enabledPath() (string, error) {
	return sentinelPath()
}

// disabledPath returns the path to the disabled sentinel file.
func disabledPath() (string, error) {
	p, err := sentinelPath()
	if err != nil {
		return "", err
	}

	return p + ".disabled", nil
}

// Revoke removes the active sentinel session and creates a disabled marker.
func Revoke() error {
	path, err := sentinelPath()
	if err != nil {
		return err
	}

	dp, err := disabledPath()
	if err != nil {
		return err
	}

	if err := os.Rename(path, dp); os.IsNotExist(err) {
		return ErrNoActive
	} else if err != nil {
		return fmt.Errorf("revoke sentinel: %w", err)
	}

	return nil
}
