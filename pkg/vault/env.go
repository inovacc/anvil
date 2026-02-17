package vault

import (
	"fmt"
	"time"

	"github.com/inovacc/anvil/internal/sentinel"
)

const (
	minTTL = 1 * time.Minute
	maxTTL = 24 * time.Hour
)

// toReleaseState converts an internal sentinel.ReleaseState to the public type.
func toReleaseState(s *sentinel.ReleaseState) *ReleaseState {
	if s == nil {
		return &ReleaseState{}
	}

	return &ReleaseState{
		Active:      s.Active,
		ProfileName: s.ProfileName,
		ExpiresAt:   s.ExpiresAt,
		SessionID:   s.SessionID,
		Remaining:   s.Remaining,
	}
}

// EnvRelease verifies the password and creates a time-limited release session.
func (v *Vault) EnvRelease(password string, opts *EnvReleaseOptions) (*ReleaseState, error) {
	if err := v.VerifyPassword(password); err != nil {
		return nil, err
	}

	profileName := ""
	ttl := 30 * time.Minute

	if opts != nil {
		if opts.ProfileName != "" {
			profileName = opts.ProfileName
		}

		if opts.TTL > 0 {
			ttl = opts.TTL
		}
	}

	if ttl < minTTL {
		return nil, fmt.Errorf("TTL must be at least %s", minTTL)
	}

	if ttl > maxTTL {
		return nil, fmt.Errorf("TTL must be at most %s", maxTTL)
	}

	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return nil, err
	}

	state, err := sentinel.Release(v.masterKey, profile, ttl)
	if err != nil {
		return nil, err
	}

	v.logAudit("env.release", profile, "", fmt.Sprintf("ttl %s", ttl))

	return toReleaseState(state), nil
}

// EnvRevoke revokes the active release session.
func (v *Vault) EnvRevoke() error {
	if err := sentinel.Revoke(); err != nil {
		return err
	}

	v.logAudit("env.revoke", "", "", "")

	return nil
}

// EnvStatus returns the current release state.
func (v *Vault) EnvStatus() (*ReleaseState, error) {
	state, err := sentinel.Check(v.masterKey)
	if err != nil {
		return nil, err
	}

	return toReleaseState(state), nil
}

// EnvExport checks the sentinel and exports secrets for the released profile.
// Secrets are decrypted on the fly and never cached on disk.
func (v *Vault) EnvExport(profileName string) ([]SecretEntry, error) {
	state, err := sentinel.Check(v.masterKey)
	if err != nil {
		return nil, fmt.Errorf("check release: %w", err)
	}

	if !state.Active {
		return nil, ErrNotReleased
	}

	// Use the released profile if no explicit profile given
	profile := profileName
	if profile == "" {
		profile = state.ProfileName
	}

	return v.Export(profile)
}

// EnvInlineGet checks the sentinel and returns a single secret value.
// This is for inline usage like: MY_KEY=$(profile --env-inline MY_KEY)
func (v *Vault) EnvInlineGet(key, profileName string) (string, error) {
	state, err := sentinel.Check(v.masterKey)
	if err != nil {
		return "", fmt.Errorf("check release: %w", err)
	}

	if !state.Active {
		return "", ErrNotReleased
	}

	profile := profileName
	if profile == "" {
		profile = state.ProfileName
	}

	return v.Get(key, profile)
}
