package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inovacc/anvil/pkg/vault"
)

type dashboardModel struct {
	status   *vault.Status
	profiles []vault.ProfileInfo
	loaded   bool
}

type dashboardLoadedMsg struct {
	status   *vault.Status
	profiles []vault.ProfileInfo
	err      error
}

func (d dashboardModel) Init() tea.Cmd {
	return nil
}

func (d dashboardModel) loadData(v *vault.Vault) tea.Cmd {
	return func() tea.Msg {
		status, err := v.Status()
		if err != nil {
			return dashboardLoadedMsg{err: err}
		}
		profiles, err := v.ListProfiles()
		if err != nil {
			return dashboardLoadedMsg{err: err}
		}
		return dashboardLoadedMsg{status: status, profiles: profiles}
	}
}

func (d dashboardModel) Update(msg tea.Msg) (dashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardLoadedMsg:
		d.loaded = true
		if msg.err == nil {
			d.status = msg.status
			d.profiles = msg.profiles
		}
	}
	return d, nil
}

func (d dashboardModel) View(width int) string {
	if !d.loaded {
		return "\n  Loading..."
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("  Anvil Vault  "))
	b.WriteString("\n\n")

	s := d.status
	if s == nil {
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

	row("Database", s.DBPath)
	row("Seal Method", s.SealMethod)
	row("Key Version", fmt.Sprintf("%d", s.KeyVersion))
	row("Profiles", fmt.Sprintf("%d", s.ProfileCount))
	row("Secrets", fmt.Sprintf("%d", s.SecretCount))
	row("Password", boolToYesNo(s.PasswordSet))
	if !s.CreatedAt.IsZero() {
		row("Created", s.CreatedAt.Format("2006-01-02 15:04"))
	}

	// Show default profile
	for _, p := range d.profiles {
		if p.IsDefault {
			info.WriteString("\n")
			row("Default", fmt.Sprintf("%s (%d secrets)", p.Name, p.SecretCount))
			break
		}
	}

	b.WriteString(box.Render(info.String()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  p: profiles • q: quit"))

	return b.String()
}

func boolToYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
