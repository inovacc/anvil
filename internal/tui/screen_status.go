package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inovacc/anvil/pkg/vault"
)

type statusModel struct {
	status   *vault.Status
	profiles []vault.ProfileInfo
	loaded   bool
}

type statusLoadedMsg struct {
	status   *vault.Status
	profiles []vault.ProfileInfo
	err      error
}

func newStatusModel() statusModel {
	return statusModel{}
}

func (s statusModel) loadData() tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		status, err := v.Status()
		if err != nil {
			return statusLoadedMsg{err: err}
		}

		profiles, err := v.ListProfiles()
		if err != nil {
			return statusLoadedMsg{err: err}
		}

		return statusLoadedMsg{status: status, profiles: profiles}
	})
}

func (s statusModel) Update(msg tea.Msg) (statusModel, tea.Cmd) {
	if msg, ok := msg.(statusLoadedMsg); ok {
		s.loaded = true
		if msg.err == nil {
			s.status = msg.status
			s.profiles = msg.profiles
		}
	}

	return s, nil
}

func (s statusModel) View(width int) string {
	if !s.loaded {
		return "\n  Loading..."
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("  Anvil Vault  "))
	b.WriteString("\n\n")

	st := s.status
	if st == nil {
		b.WriteString(errorStyle.Render("  Vault not initialized."))
		return b.String()
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(min(60, width-4))

	var info strings.Builder

	row := func(label, value string) {
		info.WriteString(fmt.Sprintf("  %s  %s\n", labelStyle.Width(14).Render(label), valueStyle.Render(value)))
	}

	row("Database", st.DBPath)
	row("Seal Method", st.SealMethod)
	row("Key Version", fmt.Sprintf("%d", st.KeyVersion))
	row("Profiles", fmt.Sprintf("%d", st.ProfileCount))
	row("Secrets", fmt.Sprintf("%d", st.SecretCount))
	row("Password", boolToYesNo(st.PasswordSet))

	if !st.CreatedAt.IsZero() {
		row("Created", st.CreatedAt.Format("2006-01-02 15:04"))
	}

	for _, p := range s.profiles {
		if p.IsDefault {
			info.WriteString("\n")
			row("Default", fmt.Sprintf("%s (%d secrets)", p.Name, p.SecretCount))

			break
		}
	}

	b.WriteString(box.Render(info.String()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  tab: navigate • q: quit"))

	return b.String()
}

func boolToYesNo(v bool) string {
	if v {
		return "yes"
	}

	return "no"
}
