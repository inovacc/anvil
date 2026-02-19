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

	// ErrProfileUUIDNotFound is returned when a profile UUID does not exist.
	ErrProfileUUIDNotFound = &UserError{
		Message: "profile UUID not found",
		Hint:    "Verify the UUID and try again",
	}

	// ErrVaultSealed is returned when the vault is temporarily sealed.
	ErrVaultSealed = &UserError{
		Message: "vault is sealed",
		Hint:    "Run 'anvil vault unseal' to unlock the vault",
	}

	// ErrVaultNotSealed is returned when trying to unseal a vault that is not sealed.
	ErrVaultNotSealed = &UserError{
		Message: "vault is not sealed",
	}

	// ErrReadDenied is returned when scoped access attempts to read plaintext secrets.
	ErrReadDenied = &UserError{
		Message: "plaintext read denied for scoped access",
		Hint:    "Scoped vault only allows masked reads and writes; use the CLI for full access",
	}

	// ErrReleaseExpired is returned when the release session has expired.
	ErrReleaseExpired = &UserError{
		Message: "release session has expired",
		Hint:    "Run 'profile vault env release' to create a new session",
	}

	// ErrNoRecovery is returned when the vault has no recovery phrase configured.
	ErrNoRecovery = &UserError{
		Message: "no recovery phrase configured",
		Hint:    "Recovery phrase is set during vault initialization",
	}

	// ErrInvalidMnemonic is returned when a provided mnemonic is invalid.
	ErrInvalidMnemonic = &UserError{
		Message: "invalid recovery phrase",
		Hint:    "Provide a valid 24-word BIP-39 mnemonic",
	}

	// ErrKeyNotFound is returned when an asymmetric key does not exist.
	ErrKeyNotFound = &UserError{
		Message: "key not found",
		Hint:    "Run 'anvil key list' to see available keys",
	}

	// ErrKeyAlreadyExists is returned when trying to create a key that already exists.
	ErrKeyAlreadyExists = &UserError{
		Message: "key already exists",
		Hint:    "Choose a different name or delete the existing key",
	}

	// ErrUnsupportedAlgorithm is returned for an unsupported key algorithm.
	ErrUnsupportedAlgorithm = &UserError{
		Message: "unsupported algorithm",
		Hint:    "Supported algorithms: ed25519, ecdsa-p256",
	}

	// ErrSignatureInvalid is returned when signature verification fails.
	ErrSignatureInvalid = &UserError{
		Message: "invalid signature",
	}

	// ErrRecoveryMismatch is returned when the mnemonic does not match the vault.
	ErrRecoveryMismatch = &UserError{
		Message: "recovery phrase does not match this vault",
		Hint:    "Verify you are using the correct 24-word phrase for this vault",
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
