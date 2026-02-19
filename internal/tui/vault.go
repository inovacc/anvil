package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

// withVault opens the vault, calls fn, then closes it.
// Returns a tea.Cmd that executes the operation asynchronously.
func withVault(fn func(v *vault.Vault) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		v, err := vault.Open(nil)
		if err != nil {
			return vaultErrorMsg{err: err}
		}

		defer func() { _ = v.Close() }()

		return fn(v)
	}
}

type vaultErrorMsg struct{ err error }
