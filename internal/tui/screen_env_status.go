package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type envStatusModel struct {
	state   *vault.ReleaseState
	loaded  bool
	message string
}

type envStatusLoadedMsg struct {
	state *vault.ReleaseState
	err   error
}

type envStatusActionMsg struct {
	message string
	err     error
}

func newEnvStatusModel() envStatusModel {
	return envStatusModel{}
}

func (e envStatusModel) loadData() tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		state, err := v.EnvStatus()
		if err != nil {
			return envStatusLoadedMsg{err: err}
		}

		return envStatusLoadedMsg{state: state}
	})
}

func (e envStatusModel) Update(msg tea.Msg) (envStatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case envStatusLoadedMsg:
		e.loaded = true
		if msg.err == nil {
			e.state = msg.state
		} else {
			e.message = errorStyle.Render(msg.err.Error())
		}
	case envStatusActionMsg:
		if msg.err != nil {
			e.message = errorStyle.Render(msg.err.Error())
		} else {
			e.message = successStyle.Render(msg.message)
		}
	}

	return e, nil
}

func (e envStatusModel) handleKey(msg tea.KeyMsg) (envStatusModel, tea.Cmd) {
	if msg.String() == "r" {
		return e, withVault(func(v *vault.Vault) tea.Msg {
			if err := v.EnvRevoke(); err != nil {
				return envStatusActionMsg{err: err}
			}

			state, err := v.EnvStatus()
			if err != nil {
				return envStatusLoadedMsg{err: err}
			}

			return envStatusLoadedMsg{state: state}
		})
	}

	return e, nil
}

func (e envStatusModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Env Status  "))
	b.WriteString("\n\n")

	if !e.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if e.state == nil || !e.state.Active {
		b.WriteString(dimStyle.Render("  Status: inactive"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(successStyle.Render("  Status: active"))
		b.WriteString("\n")
		fmt.Fprintf(&b, "  Profile:   %s\n", e.state.ProfileName)
		fmt.Fprintf(&b, "  Session:   %s\n", e.state.SessionID)
		fmt.Fprintf(&b, "  Expires:   %s\n", e.state.ExpiresAt.Format(time.RFC3339))
		fmt.Fprintf(&b, "  Remaining: %s\n", e.state.Remaining.Truncate(time.Second))
		b.WriteString("\n")
	}

	if e.message != "" {
		b.WriteString("  " + e.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  r: revoke • esc: sidebar"))

	return b.String()
}
