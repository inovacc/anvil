package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type templatesModel struct {
	templates []vault.TemplateInfo
	table     table.Model
	loaded    bool
	message   string
}

type templatesLoadedMsg struct {
	templates []vault.TemplateInfo
	err       error
}

func newTemplatesModel(width, height int) templatesModel {
	columns := templateTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())
	return templatesModel{table: t}
}

func templateTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}
	w := width - 4
	return []table.Column{
		{Title: "Name", Width: w * 30 / 100},
		{Title: "Vars", Width: w * 15 / 100},
		{Title: "Secrets", Width: w * 15 / 100},
		{Title: "Description", Width: w * 40 / 100},
	}
}

func (t templatesModel) loadData(v *vault.Vault) tea.Cmd {
	return func() tea.Msg {
		templates, err := v.ListTemplates()
		if err != nil {
			return templatesLoadedMsg{err: err}
		}
		return templatesLoadedMsg{templates: templates}
	}
}

func (t templatesModel) Update(msg tea.Msg) (templatesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case templatesLoadedMsg:
		t.loaded = true
		if msg.err == nil {
			t.templates = msg.templates
			rows := make([]table.Row, len(t.templates))
			for i, tmpl := range t.templates {
				rows[i] = table.Row{
					tmpl.Name,
					fmt.Sprintf("%d", tmpl.VarCount),
					fmt.Sprintf("%d", tmpl.SecretCount),
					tmpl.Description,
				}
			}
			t.table.SetRows(rows)
		} else {
			t.message = errorStyle.Render(msg.err.Error())
		}
	}
	return t, nil
}

func (t templatesModel) selectedName() string {
	row := t.table.SelectedRow()
	if len(row) == 0 {
		return ""
	}
	return row[0]
}

func (t templatesModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Templates  "))
	b.WriteString("\n\n")

	if !t.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(t.templates) == 0 {
		b.WriteString(dimStyle.Render("  No templates found."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: sidebar"))
		return b.String()
	}

	b.WriteString(t.table.View())
	b.WriteString("\n")

	if t.message != "" {
		b.WriteString("  " + t.message + "\n")
	}

	b.WriteString(helpStyle.Render("  d: delete • esc: sidebar"))

	return b.String()
}
