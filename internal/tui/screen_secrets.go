package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type secretsModel struct {
	profileName string
	secrets     []vault.SecretInfo
	table       table.Model
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

func newSecretsModel(profileName string, width, height int) secretsModel {
	columns := tableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())

	return secretsModel{
		profileName: profileName,
		table:       t,
	}
}

func tableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}

	w := width - 4 // padding

	return []table.Column{
		{Title: "Key", Width: w * 30 / 100},
		{Title: "Description", Width: w * 30 / 100},
		{Title: "Created", Width: w * 20 / 100},
		{Title: "Updated", Width: w * 20 / 100},
	}
}

func (s secretsModel) loadData(profile string) tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		secrets, err := v.List(profile)
		if err != nil {
			return secretsLoadedMsg{err: err}
		}

		return secretsLoadedMsg{secrets: secrets}
	})
}

func (s secretsModel) Update(msg tea.Msg) (secretsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case secretsLoadedMsg:
		s.loaded = true
		if msg.err == nil {
			s.secrets = msg.secrets

			rows := make([]table.Row, len(s.secrets))
			for i, sec := range s.secrets {
				updated := "-"
				if !sec.UpdatedAt.IsZero() {
					updated = sec.UpdatedAt.Format("2006-01-02 15:04")
				}

				rows[i] = table.Row{
					sec.Key,
					sec.Description,
					sec.CreatedAt.Format("2006-01-02 15:04"),
					updated,
				}
			}

			s.table.SetRows(rows)
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

	b.WriteString(s.table.View())
	b.WriteString("\n")

	if s.revealKey != "" {
		b.WriteString(fmt.Sprintf("  %s: %s\n", labelStyle.Render(s.revealKey), valueStyle.Render(s.revealValue)))
	}

	if s.message != "" {
		b.WriteString("  " + s.message + "\n")
	}

	b.WriteString(helpStyle.Render("  enter: reveal • h: history • n: new • d: delete • esc: back"))

	return b.String()
}

func (s secretsModel) selectedKey() string {
	row := s.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}

	return row[0]
}
