package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type envReleaseModel struct {
	passwordInput textinput.Model
	ttlInput      textinput.Model
	focusIndex    int
	message       string
}

type envReleaseActionMsg struct {
	message string
	err     error
}

func newEnvReleaseModel() envReleaseModel {
	pw := textinput.New()
	pw.Placeholder = "vault password"
	pw.EchoMode = textinput.EchoPassword
	pw.CharLimit = 256
	pw.Width = 40
	pw.PromptStyle = focusedInputStyle
	pw.TextStyle = focusedInputStyle

	ttl := textinput.New()
	ttl.Placeholder = "30m (e.g. 1h, 15m)"
	ttl.CharLimit = 20
	ttl.Width = 40
	ttl.PromptStyle = blurredInputStyle
	ttl.TextStyle = blurredInputStyle

	return envReleaseModel{passwordInput: pw, ttlInput: ttl}
}

func (e envReleaseModel) Update(msg tea.Msg) (envReleaseModel, tea.Cmd) {
	switch msg := msg.(type) {
	case envReleaseActionMsg:
		if msg.err != nil {
			e.message = errorStyle.Render(msg.err.Error())
		} else {
			e.message = successStyle.Render(msg.message)
		}
		return e, nil
	}
	var cmd tea.Cmd
	if e.focusIndex == 0 {
		e.passwordInput, cmd = e.passwordInput.Update(msg)
	} else {
		e.ttlInput, cmd = e.ttlInput.Update(msg)
	}
	return e, cmd
}

func (e envReleaseModel) handleKey(msg tea.KeyMsg, v *vault.Vault) (envReleaseModel, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if e.focusIndex == 0 {
			e.focusIndex = 1
			e.passwordInput.Blur()
			e.passwordInput.PromptStyle = blurredInputStyle
			e.passwordInput.TextStyle = blurredInputStyle
			return e, e.ttlInput.Focus()
		}
		e.focusIndex = 0
		e.ttlInput.Blur()
		e.ttlInput.PromptStyle = blurredInputStyle
		e.ttlInput.TextStyle = blurredInputStyle
		return e, e.passwordInput.Focus()
	case "enter":
		pw := e.passwordInput.Value()
		if pw == "" {
			e.message = errorStyle.Render("Password is required.")
			return e, nil
		}
		ttlStr := e.ttlInput.Value()
		if ttlStr == "" {
			ttlStr = "30m"
		}
		ttl, err := time.ParseDuration(ttlStr)
		if err != nil {
			e.message = errorStyle.Render("Invalid TTL: " + err.Error())
			return e, nil
		}
		e.passwordInput.SetValue("")
		return e, func() tea.Msg {
			state, err := v.EnvRelease(pw, &vault.EnvReleaseOptions{TTL: ttl})
			if err != nil {
				return envReleaseActionMsg{err: err}
			}
			return envReleaseActionMsg{message: fmt.Sprintf("Released for %q. Expires in %s", state.ProfileName, state.Remaining.Truncate(time.Second))}
		}
	case "r":
		return e, func() tea.Msg {
			if err := v.EnvRevoke(); err != nil {
				return envReleaseActionMsg{err: err}
			}
			return envReleaseActionMsg{message: "Release revoked."}
		}
	}
	var cmd tea.Cmd
	if e.focusIndex == 0 {
		e.passwordInput, cmd = e.passwordInput.Update(msg)
	} else {
		e.ttlInput, cmd = e.ttlInput.Update(msg)
	}
	return e, cmd
}

func (e envReleaseModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Env Release  "))
	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("Password") + "\n")
	b.WriteString("  " + e.passwordInput.View() + "\n\n")
	b.WriteString("  " + inputLabelStyle.Render("TTL") + "\n")
	b.WriteString("  " + e.ttlInput.View() + "\n\n")

	if e.message != "" {
		b.WriteString("  " + e.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter: release • r: revoke • tab: switch field • esc: sidebar"))

	return b.String()
}
