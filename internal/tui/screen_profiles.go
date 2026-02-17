package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type profilesModel struct {
	profiles []vault.ProfileInfo
	table    table.Model
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

func newProfilesModel(width, height int) profilesModel {
	columns := profileTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())
	return profilesModel{table: t}
}

func profileTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}
	w := width - 4
	return []table.Column{
		{Title: "Name", Width: w * 25 / 100},
		{Title: "Default", Width: w * 10 / 100},
		{Title: "Secrets", Width: w * 10 / 100},
		{Title: "Description", Width: w * 30 / 100},
		{Title: "Created", Width: w * 25 / 100},
	}
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
			rows := make([]table.Row, len(p.profiles))
			for i, prof := range p.profiles {
				def := ""
				if prof.IsDefault {
					def = "*"
				}
				created := "-"
				if !prof.CreatedAt.IsZero() {
					created = prof.CreatedAt.Format("2006-01-02 15:04")
				}
				rows[i] = table.Row{
					prof.Name,
					def,
					fmt.Sprintf("%d", prof.SecretCount),
					prof.Description,
					created,
				}
			}
			p.table.SetRows(rows)
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

	b.WriteString(p.table.View())
	b.WriteString("\n")

	if p.message != "" {
		b.WriteString("  " + p.message + "\n")
	}

	b.WriteString(helpStyle.Render("  enter: secrets • c: create • d: delete • u: set default • esc: back"))

	return b.String()
}

func (p profilesModel) selectedProfile() string {
	row := p.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	return row[0]
}
