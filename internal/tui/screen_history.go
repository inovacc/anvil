package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type historyModel struct {
	key         string
	profileName string
	versions    []vault.SecretVersion
	table       table.Model
	loaded      bool
}

type historyLoadedMsg struct {
	versions []vault.SecretVersion
	err      error
}

func newHistoryModel(key, profileName string, width, height int) historyModel {
	columns := historyTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())
	return historyModel{
		key:         key,
		profileName: profileName,
		table:       t,
	}
}

func historyTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}
	w := width - 4
	return []table.Column{
		{Title: "Version", Width: w * 30 / 100},
		{Title: "Created", Width: w * 70 / 100},
	}
}

func (h historyModel) loadData(v *vault.Vault) tea.Cmd {
	key := h.key
	profile := h.profileName
	return func() tea.Msg {
		versions, err := v.SecretHistory(key, profile)
		if err != nil {
			return historyLoadedMsg{err: err}
		}
		return historyLoadedMsg{versions: versions}
	}
}

func (h historyModel) Update(msg tea.Msg) (historyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case historyLoadedMsg:
		h.loaded = true
		if msg.err == nil {
			h.versions = msg.versions
			rows := make([]table.Row, len(h.versions))
			for i, ver := range h.versions {
				rows[i] = table.Row{
					fmt.Sprintf("v%d", ver.Version),
					ver.CreatedAt.Format("2006-01-02 15:04:05"),
				}
			}
			h.table.SetRows(rows)
		}
	}
	return h, nil
}

func (h historyModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("  History [%s]  ", h.key)))
	b.WriteString("\n\n")

	if !h.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(h.versions) == 0 {
		b.WriteString(dimStyle.Render("  No version history."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: back • q: quit"))
		return b.String()
	}

	b.WriteString(h.table.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  esc: back • q: quit"))

	return b.String()
}
