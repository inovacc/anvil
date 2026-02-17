package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type keyRotationModel struct {
	input   textinput.Model
	message string
}

type keyRotationMsg struct {
	message string
	err     error
}

func newKeyRotationModel() keyRotationModel {
	ti := textinput.New()
	ti.Placeholder = "vault password"
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 256
	ti.Width = 40
	ti.PromptStyle = focusedInputStyle
	ti.TextStyle = focusedInputStyle

	return keyRotationModel{input: ti}
}

func (k keyRotationModel) Update(msg tea.Msg) (keyRotationModel, tea.Cmd) {
	if msg, ok := msg.(keyRotationMsg); ok {
		if msg.err != nil {
			k.message = errorStyle.Render(msg.err.Error())
		} else {
			k.message = successStyle.Render(msg.message)
		}

		return k, nil
	}

	var cmd tea.Cmd

	k.input, cmd = k.input.Update(msg)

	return k, cmd
}

func (k keyRotationModel) handleKey(msg tea.KeyMsg, v *vault.Vault) (keyRotationModel, tea.Cmd) {
	if msg.String() == "enter" {
		pw := k.input.Value()
		if pw == "" {
			k.message = errorStyle.Render("Password is required.")
			return k, nil
		}

		k.input.SetValue("")

		return k, func() tea.Msg {
			if err := v.RotateKey(pw); err != nil {
				return keyRotationMsg{err: err}
			}

			status, err := v.Status()
			if err != nil {
				return keyRotationMsg{message: "Key rotated (could not fetch version)."}
			}

			return keyRotationMsg{message: fmt.Sprintf("Key rotated to version %d.", status.KeyVersion)}
		}
	}

	var cmd tea.Cmd

	k.input, cmd = k.input.Update(msg)

	return k, cmd
}

func (k keyRotationModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Key Rotation  "))
	b.WriteString("\n\n")

	b.WriteString(warningStyle.Render("  This will re-encrypt all secrets with a new master key."))
	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("Password") + "\n")
	b.WriteString("  " + k.input.View() + "\n\n")

	if k.message != "" {
		b.WriteString("  " + k.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter: rotate • esc: sidebar"))

	return b.String()
}
