package vault_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func TestUserErrorMessage(t *testing.T) {
	err := vault.ErrNotInitialized
	if err.Error() != "vault is not initialized" {
		t.Errorf("Error() = %q, want %q", err.Error(), "vault is not initialized")
	}
}

func TestIsUserError(t *testing.T) {
	t.Run("direct UserError", func(t *testing.T) {
		ue, ok := vault.IsUserError(vault.ErrProfileNotFound)
		if !ok {
			t.Fatal("expected ok")
		}

		if ue.Message != "profile not found" {
			t.Errorf("Message = %q", ue.Message)
		}

		if ue.Hint == "" {
			t.Error("expected non-empty hint")
		}
	})

	t.Run("wrapped UserError", func(t *testing.T) {
		wrapped := fmt.Errorf("outer: %w", vault.ErrPasswordMismatch)

		ue, ok := vault.IsUserError(wrapped)
		if !ok {
			t.Fatal("expected ok for wrapped error")
		}

		if ue.Message != "incorrect password" {
			t.Errorf("Message = %q", ue.Message)
		}
	})

	t.Run("non-UserError", func(t *testing.T) {
		_, ok := vault.IsUserError(errors.New("generic error"))
		if ok {
			t.Error("expected not ok for generic error")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		_, ok := vault.IsUserError(nil)
		if ok {
			t.Error("expected not ok for nil error")
		}
	})
}

func TestDefaultDBPath(t *testing.T) {
	p := vault.DefaultDBPath()
	if p == "" {
		t.Error("expected non-empty DefaultDBPath")
	}
}
