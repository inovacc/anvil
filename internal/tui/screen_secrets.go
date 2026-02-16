package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type secretsModel struct {
	profileName string
	secrets     []vault.SecretInfo
	cursor      int
	loaded      bool
	confirm     string // confirming delete of this key
	message     string
	revealKey   string // key whose value is currently shown
	revealValue string
}

type secretsLoadedMsg struct {
	secrets []vault.SecretInfo
	err     error
}

type secretRevealMsg struct {
	key   string
	value string
	err   error
}

type secretActionMsg struct {
	message string
	err     error
}

func (s secretsModel) loadData(v *vault.Vault, profile string) tea.Cmd {
	return func() tea.Msg {
		secrets, err := v.List(profile)
		if err != nil {
			return secretsLoadedMsg{err: err}
		}
		return secretsLoadedMsg{secrets: secrets}
	}
}

func (s secretsModel) Update(msg tea.Msg) (secretsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case secretsLoadedMsg:
		s.loaded = true
		if msg.err == nil {
			s.secrets = msg.secrets
			if s.cursor >= len(s.secrets) {
				s.cursor = max(0, len(s.secrets)-1)
			}
		}
	case secretRevealMsg:
		if msg.err != nil {
			s.message = errorStyle.Render(msg.err.Error())
		} else {
			s.revealKey = msg.key
			s.revealValue = msg.value
		}
	case secretActionMsg:
		if msg.err != nil {
			s.message = errorStyle.Render(msg.err.Error())
		} else {
			s.message = successStyle.Render(msg.message)
		}
		s.confirm = ""
	}
	return s, nil
}

func (s secretsModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("  Secrets [%s]  ", s.profileName)))
	b.WriteString("\n\n")

	if !s.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(s.secrets) == 0 {
		b.WriteString(dimStyle.Render("  No secrets. Press n to create one."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  n: new • esc: back"))
		return b.String()
	}

	if s.confirm != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Delete secret %q? (y/n)", s.confirm)))
		return b.String()
	}

	for i, sec := range s.secrets {
		cursor := "  "
		style := normalStyle
		if i == s.cursor {
			cursor = "> "
			style = selectedStyle
		}

		line := fmt.Sprintf("%s%s", cursor, sec.Key)
		b.WriteString(style.Render(line))
		if sec.Description != "" {
			b.WriteString(dimStyle.Render(fmt.Sprintf(" — %s", sec.Description)))
		}
		b.WriteString("\n")

		// Show revealed value inline
		if s.revealKey == sec.Key && i == s.cursor {
			b.WriteString(fmt.Sprintf("    %s\n", valueStyle.Render(s.revealValue)))
		}
	}

	if s.message != "" {
		b.WriteString("\n  " + s.message)
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter: reveal • n: new • d: delete • esc: back"))

	return b.String()
}

func (s secretsModel) selectedKey() string {
	if len(s.secrets) == 0 || s.cursor >= len(s.secrets) {
		return ""
	}
	return s.secrets[s.cursor].Key
}
