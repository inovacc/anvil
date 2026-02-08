package vault

import "errors"

var (
	// ErrNotInitialized is returned when the vault has not been initialized.
	ErrNotInitialized = errors.New("vault not initialized, run 'profile vault init'")

	// ErrAlreadyInitialized is returned when trying to init an already-initialized vault.
	ErrAlreadyInitialized = errors.New("vault already initialized")

	// ErrMachineMismatch is returned when the vault was sealed on a different machine.
	ErrMachineMismatch = errors.New("vault was sealed on a different machine")

	// ErrProfileNotFound is returned when a profile does not exist.
	ErrProfileNotFound = errors.New("profile not found")

	// ErrProfileExists is returned when trying to create a profile that already exists.
	ErrProfileExists = errors.New("profile already exists")

	// ErrSecretNotFound is returned when a secret does not exist.
	ErrSecretNotFound = errors.New("secret not found")

	// ErrNoDefaultProfile is returned when no default profile is set.
	ErrNoDefaultProfile = errors.New("no default profile set, create one with 'profile vault profile create <name> --default'")

	// ErrPasswordNotSet is returned when env operations require a password but none is set.
	ErrPasswordNotSet = errors.New("vault password not set, run 'profile vault env password set'")

	// ErrPasswordMismatch is returned when the provided password does not match.
	ErrPasswordMismatch = errors.New("incorrect password")

	// ErrNotReleased is returned when env export is called without an active release.
	ErrNotReleased = errors.New("secrets not released, run 'profile vault env release'")

	// ErrReleaseExpired is returned when the release session has expired.
	ErrReleaseExpired = errors.New("release session has expired")
)
