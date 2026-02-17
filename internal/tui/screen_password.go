package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type passwordScreenModel struct {
	hasPassword bool
	input       textinput.Model
	message     string
	loaded      bool
}

type passwordStatusMsg struct {
	hasPassword bool
	err         error
}

type passwordActionMsg struct {
	message string
	err     error
}

func newPasswordScreenModel() passwordScreenModel {
	ti := textinput.New()
	ti.Placeholder = "password (min 8 chars)"
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 256
	ti.Width = 40
	ti.PromptStyle = focusedInputStyle
	ti.TextStyle = focusedInputStyle

	return passwordScreenModel{input: ti}
}

func (p passwordScreenModel) loadData(v *vault.Vault) tea.Cmd {
	return func() tea.Msg {
		has, err := v.HasPassword()
		return passwordStatusMsg{hasPassword: has, err: err}
	}
}

func (p passwordScreenModel) Update(msg tea.Msg) (passwordScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case passwordStatusMsg:
		p.loaded = true
		if msg.err == nil {
			p.hasPassword = msg.hasPassword
		}
	case passwordActionMsg:
		if msg.err != nil {
			p.message = errorStyle.Render(msg.err.Error())
		} else {
			p.message = successStyle.Render(msg.message)
		}

		return p, p.loadData(nil) // will be called with vault in handleKey
	}

	var cmd tea.Cmd

	p.input, cmd = p.input.Update(msg)

	return p, cmd
}

func (p passwordScreenModel) handleKey(msg tea.KeyMsg, v *vault.Vault) (passwordScreenModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		pw := p.input.Value()
		if len(pw) < 8 {
			p.message = errorStyle.Render("Password must be at least 8 characters.")
			return p, nil
		}

		p.input.SetValue("")

		return p, func() tea.Msg {
			if err := v.SetPassword(pw); err != nil {
				return passwordActionMsg{err: err}
			}

			return passwordActionMsg{message: "Password set successfully."}
		}
	case "d":
		if p.hasPassword {
			return p, func() tea.Msg {
				if err := v.DeletePassword(); err != nil {
					return passwordActionMsg{err: err}
				}

				_ = v.EnvRevoke()

				return passwordActionMsg{message: "Password removed."}
			}
		}
	}

	var cmd tea.Cmd

	p.input, cmd = p.input.Update(msg)

	return p, cmd
}

func (p passwordScreenModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Password  "))
	b.WriteString("\n\n")

	if !p.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if p.hasPassword {
		b.WriteString(successStyle.Render("  Password is set."))
	} else {
		b.WriteString(dimStyle.Render("  No password configured."))
	}

	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("New Password") + "\n")
	b.WriteString("  " + p.input.View() + "\n\n")

	if p.message != "" {
		b.WriteString("  " + p.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter: set password • d: delete • esc: sidebar"))

	return b.String()
}
