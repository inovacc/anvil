package vault

import "errors"

// UserError provides user-friendly error messages with optional hints.
type UserError struct {
	Message string
	Hint    string
}

func (e *UserError) Error() string { return e.Message }

var (
	// ErrNotInitialized is returned when the vault has not been initialized.
	ErrNotInitialized = &UserError{
		Message: "vault is not initialized",
		Hint:    "Run 'profile vault init' to create a new vault",
	}

	// ErrAlreadyInitialized is returned when trying to init an already-initialized vault.
	ErrAlreadyInitialized = &UserError{
		Message: "vault is already initialized",
	}

	// ErrMachineMismatch is returned when the vault was sealed on a different machine.
	ErrMachineMismatch = &UserError{
		Message: "vault was sealed on a different machine",
		Hint:    "This vault cannot be opened on this machine",
	}

	// ErrProfileNotFound is returned when a profile does not exist.
	ErrProfileNotFound = &UserError{
		Message: "profile not found",
		Hint:    "Run 'profile vault profile list' to see available profiles",
	}

	// ErrProfileExists is returned when trying to create a profile that already exists.
	ErrProfileExists = &UserError{
		Message: "profile already exists",
		Hint:    "Choose a different name or delete the existing profile",
	}

	// ErrSecretNotFound is returned when a secret does not exist.
	ErrSecretNotFound = &UserError{
		Message: "secret not found",
		Hint:    "Run 'profile vault list' to see available secrets",
	}

	// ErrNoDefaultProfile is returned when no default profile is set.
	ErrNoDefaultProfile = &UserError{
		Message: "no default profile set",
		Hint:    "Run 'profile vault profile create <name> --default' to create one",
	}

	// ErrPasswordNotSet is returned when env operations require a password but none is set.
	ErrPasswordNotSet = &UserError{
		Message: "vault password is not set",
		Hint:    "Run 'profile vault env password set' to set a password",
	}

	// ErrPasswordMismatch is returned when the provided password does not match.
	ErrPasswordMismatch = &UserError{
		Message: "incorrect password",
	}

	// ErrNotReleased is returned when env export is called without an active release.
	ErrNotReleased = &UserError{
		Message: "secrets have not been released",
		Hint:    "Run 'profile vault env release' to release secrets",
	}

	// ErrReleaseExpired is returned when the release session has expired.
	ErrReleaseExpired = &UserError{
		Message: "release session has expired",
		Hint:    "Run 'profile vault env release' to create a new session",
	}
)

// IsUserError checks if an error is or wraps a UserError.
func IsUserError(err error) (*UserError, bool) {
	var ue *UserError
	if errors.As(err, &ue) {
		return ue, true
	}

	return nil, false
}
