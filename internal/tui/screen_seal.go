package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type sealScreenModel struct {
	sealed  bool
	loaded  bool
	message string
}

type sealStatusMsg struct {
	sealed bool
}

type sealActionMsg struct {
	message string
	err     error
}

func newSealScreenModel() sealScreenModel {
	return sealScreenModel{}
}

func (s sealScreenModel) loadData() tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		return sealStatusMsg{sealed: v.IsSealed()}
	})
}

func (s sealScreenModel) Update(msg tea.Msg) (sealScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case sealStatusMsg:
		s.loaded = true
		s.sealed = msg.sealed
	case sealActionMsg:
		if msg.err != nil {
			s.message = errorStyle.Render(msg.err.Error())
		} else {
			s.message = successStyle.Render(msg.message)
		}
	}

	return s, nil
}

func (s sealScreenModel) handleKey(msg tea.KeyMsg) (sealScreenModel, tea.Cmd) {
	switch msg.String() {
	case "s":
		return s, withVault(func(v *vault.Vault) tea.Msg {
			if err := v.Seal(); err != nil {
				return sealActionMsg{err: err}
			}

			return sealActionMsg{message: "Vault sealed."}
		})
	case "u":
		return s, func() tea.Msg {
			if err := vault.UnsealVault(nil); err != nil {
				return sealActionMsg{err: err}
			}

			return sealActionMsg{message: "Vault unsealed."}
		}
	}

	return s, nil
}

func (s sealScreenModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Seal / Unseal  "))
	b.WriteString("\n\n")

	if !s.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if s.sealed {
		b.WriteString(warningStyle.Render("  Vault is SEALED"))
	} else {
		b.WriteString(successStyle.Render("  Vault is UNSEALED"))
	}

	b.WriteString("\n\n")

	if s.message != "" {
		b.WriteString("  " + s.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  s: seal • u: unseal • esc: sidebar"))

	return b.String()
}
