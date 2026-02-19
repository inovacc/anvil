package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type appSecretsModel struct {
	appName     string
	secrets     []vault.SecretInfo
	table       table.Model
	loaded      bool
	confirm     string
	message     string
	revealKey   string
	revealValue string
}

type appSecretsLoadedMsg struct {
	secrets []vault.SecretInfo
	err     error
}

type appSecretRevealMsg struct {
	key   string
	value string
	err   error
}

type appSecretActionMsg struct {
	message string
	err     error
}

func newAppSecretsModel(appName string, width, height int) appSecretsModel {
	columns := appSecretTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())

	return appSecretsModel{
		appName: appName,
		table:   t,
	}
}

func appSecretTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}

	w := width - 4

	return []table.Column{
		{Title: "Key", Width: w * 30 / 100},
		{Title: "Description", Width: w * 30 / 100},
		{Title: "Created", Width: w * 20 / 100},
		{Title: "Updated", Width: w * 20 / 100},
	}
}

func (s appSecretsModel) loadData() tea.Cmd {
	appName := s.appName

	return withVault(func(v *vault.Vault) tea.Msg {
		av, err := v.OpenApp(appName)
		if err != nil {
			return appSecretsLoadedMsg{err: err}
		}

		defer func() { _ = av.Close() }()

		secrets, err := av.List()
		if err != nil {
			return appSecretsLoadedMsg{err: err}
		}

		return appSecretsLoadedMsg{secrets: secrets}
	})
}

func (s appSecretsModel) Update(msg tea.Msg) (appSecretsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case appSecretsLoadedMsg:
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
	case appSecretRevealMsg:
		if msg.err != nil {
			s.message = errorStyle.Render(msg.err.Error())
		} else {
			s.revealKey = msg.key
			s.revealValue = msg.value
		}
	case appSecretActionMsg:
		if msg.err != nil {
			s.message = errorStyle.Render(msg.err.Error())
		} else {
			s.message = successStyle.Render(msg.message)
		}

		s.confirm = ""
	}

	return s, nil
}

func (s appSecretsModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("  App Secrets [%s]  ", s.appName)))
	b.WriteString("\n\n")

	if !s.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(s.secrets) == 0 {
		b.WriteString(dimStyle.Render("  No secrets. Use 'anvil app set' to add one."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: back"))

		return b.String()
	}

	if s.confirm != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Delete secret %q? (y/n)", s.confirm)))
		return b.String()
	}

	b.WriteString(s.table.View())
	b.WriteString("\n")

	if s.revealKey != "" {
		fmt.Fprintf(&b, "  %s: %s\n", labelStyle.Render(s.revealKey), valueStyle.Render(s.revealValue))
	}

	if s.message != "" {
		b.WriteString("  " + s.message + "\n")
	}

	b.WriteString(helpStyle.Render("  enter: reveal • d: delete • esc: back"))

	return b.String()
}

func (s appSecretsModel) selectedKey() string {
	row := s.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}

	return row[0]
}
