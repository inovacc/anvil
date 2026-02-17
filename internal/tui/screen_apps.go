package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type appsModel struct {
	apps    []vault.AppInfo
	table   table.Model
	loaded  bool
	confirm string
	message string
}

type appsLoadedMsg struct {
	apps []vault.AppInfo
	err  error
}

type appActionMsg struct {
	action  string
	message string
	err     error
}

func newAppsModel(width, height int) appsModel {
	columns := appTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())

	return appsModel{table: t}
}

func appTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}

	w := width - 4

	return []table.Column{
		{Title: "Name", Width: w * 20 / 100},
		{Title: "UUID", Width: w * 20 / 100},
		{Title: "Service ID", Width: w * 15 / 100},
		{Title: "Status", Width: w * 10 / 100},
		{Title: "Secrets", Width: w * 10 / 100},
		{Title: "Last Accessed", Width: w * 25 / 100},
	}
}

func (a appsModel) loadData() tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		apps, err := v.ListApps()
		if err != nil {
			return appsLoadedMsg{err: err}
		}

		return appsLoadedMsg{apps: apps}
	})
}

func (a appsModel) Update(msg tea.Msg) (appsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case appsLoadedMsg:
		a.loaded = true
		if msg.err == nil {
			a.apps = msg.apps

			rows := make([]table.Row, len(a.apps))
			for i, app := range a.apps {
				lastAccessed := "-"
				if !app.LastAccessedAt.IsZero() {
					lastAccessed = app.LastAccessedAt.Format("2006-01-02 15:04")
				}

				rows[i] = table.Row{
					app.Name,
					app.UUID,
					app.ServiceID,
					app.Status,
					fmt.Sprintf("%d", app.SecretCount),
					lastAccessed,
				}
			}

			a.table.SetRows(rows)
		}
	case appActionMsg:
		if msg.err != nil {
			a.message = errorStyle.Render(msg.err.Error())
		} else {
			a.message = successStyle.Render(msg.message)
		}

		a.confirm = ""
	}

	return a, nil
}

func (a appsModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Apps  "))
	b.WriteString("\n\n")

	if !a.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(a.apps) == 0 {
		b.WriteString(dimStyle.Render("  No apps registered. Use 'anvil app register' to create one."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: back"))

		return b.String()
	}

	if a.confirm != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Remove app %q? (y/n)", a.confirm)))
		return b.String()
	}

	b.WriteString(a.table.View())
	b.WriteString("\n")

	if a.message != "" {
		b.WriteString("  " + a.message + "\n")
	}

	b.WriteString(helpStyle.Render("  enter: view secrets • d: remove • esc: back"))

	return b.String()
}

func (a appsModel) selectedApp() string {
	row := a.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}

	return row[0] // name
}
