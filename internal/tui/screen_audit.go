package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type auditModel struct {
	entries []vault.AuditEntry
	table   table.Model
	loaded  bool
}

type auditLoadedMsg struct {
	entries []vault.AuditEntry
	err     error
}

func newAuditModel(width, height int) auditModel {
	columns := auditTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())

	return auditModel{table: t}
}

func auditTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}

	w := width - 4

	return []table.Column{
		{Title: "Timestamp", Width: w * 25 / 100},
		{Title: "Action", Width: w * 15 / 100},
		{Title: "Profile", Width: w * 15 / 100},
		{Title: "Key", Width: w * 20 / 100},
		{Title: "Detail", Width: w * 25 / 100},
	}
}

func (a auditModel) loadData() tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		entries, err := v.AuditLog(100)
		if err != nil {
			return auditLoadedMsg{err: err}
		}

		return auditLoadedMsg{entries: entries}
	})
}

func (a auditModel) Update(msg tea.Msg) (auditModel, tea.Cmd) {
	if msg, ok := msg.(auditLoadedMsg); ok {
		a.loaded = true
		if msg.err == nil {
			a.entries = msg.entries

			rows := make([]table.Row, len(a.entries))
			for i, e := range a.entries {
				rows[i] = table.Row{
					e.CreatedAt.Format("2006-01-02 15:04:05"),
					e.Action,
					e.ProfileName,
					e.SecretKey,
					e.Detail,
				}
			}

			a.table.SetRows(rows)
		}
	}

	return a, nil
}

func (a auditModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Audit Log  "))
	b.WriteString("\n\n")

	if !a.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(a.entries) == 0 {
		b.WriteString(dimStyle.Render("  No audit log entries."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: back • q: quit"))

		return b.String()
	}

	b.WriteString(a.table.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  esc: back • q: quit"))

	return b.String()
}
