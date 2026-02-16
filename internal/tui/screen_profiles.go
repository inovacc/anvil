package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type profilesModel struct {
	profiles []vault.ProfileInfo
	cursor   int
	loaded   bool
	confirm  string // non-empty = confirming delete of this profile
	message  string // success/error feedback
}

type profilesLoadedMsg struct {
	profiles []vault.ProfileInfo
	err      error
}

type profileActionMsg struct {
	action  string
	message string
	err     error
}

func (p profilesModel) loadData(v *vault.Vault) tea.Cmd {
	return func() tea.Msg {
		profiles, err := v.ListProfiles()
		if err != nil {
			return profilesLoadedMsg{err: err}
		}
		return profilesLoadedMsg{profiles: profiles}
	}
}

func (p profilesModel) Update(msg tea.Msg) (profilesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case profilesLoadedMsg:
		p.loaded = true
		if msg.err == nil {
			p.profiles = msg.profiles
			if p.cursor >= len(p.profiles) {
				p.cursor = max(0, len(p.profiles)-1)
			}
		}
	case profileActionMsg:
		if msg.err != nil {
			p.message = errorStyle.Render(msg.err.Error())
		} else {
			p.message = successStyle.Render(msg.message)
		}
		p.confirm = ""
	}
	return p, nil
}

func (p profilesModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Profiles  "))
	b.WriteString("\n\n")

	if !p.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(p.profiles) == 0 {
		b.WriteString(dimStyle.Render("  No profiles found. Press c to create one."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  c: create • esc: back • q: quit"))
		return b.String()
	}

	if p.confirm != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Delete profile %q? (y/n)", p.confirm)))
		return b.String()
	}

	for i, prof := range p.profiles {
		cursor := "  "
		style := normalStyle
		if i == p.cursor {
			cursor = "> "
			style = selectedStyle
		}

		marker := ""
		if prof.IsDefault {
			marker = " *"
		}

		line := fmt.Sprintf("%s%s%s (%d secrets)", cursor, prof.Name, marker, prof.SecretCount)
		b.WriteString(style.Render(line))
		if prof.Description != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf(" — %s", prof.Description)))
		}
		b.WriteString("\n")
	}

	if p.message != "" {
		b.WriteString("\n  " + p.message)
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter: secrets • c: create • d: delete • u: set default • esc: back"))

	return b.String()
}

func (p profilesModel) selectedProfile() string {
	if len(p.profiles) == 0 || p.cursor >= len(p.profiles) {
		return ""
	}
	return p.profiles[p.cursor].Name
}
