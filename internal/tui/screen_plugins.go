package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type pluginsModel struct {
	hooks     []vault.HookConfig
	providers []vault.ProviderConfig
	table     table.Model
	loaded    bool
	message   string
}

type pluginsLoadedMsg struct {
	hooks     []vault.HookConfig
	providers []vault.ProviderConfig
	err       error
}

func newPluginsModel(width, height int) pluginsModel {
	columns := pluginTableColumns(width)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(nil),
		table.WithFocused(true),
		table.WithHeight(max(1, height-8)),
	)
	t.SetStyles(tableStyles())

	return pluginsModel{table: t}
}

func pluginTableColumns(width int) []table.Column {
	if width <= 0 {
		width = 80
	}

	w := width - 4

	return []table.Column{
		{Title: "Type", Width: w * 15 / 100},
		{Title: "Event/Name", Width: w * 25 / 100},
		{Title: "Command", Width: w * 35 / 100},
		{Title: "Detail", Width: w * 25 / 100},
	}
}

func (p pluginsModel) loadData(v *vault.Vault) tea.Cmd {
	return func() tea.Msg {
		dbPath, err := vault.ResolveDBPath(nil)
		if err != nil {
			return pluginsLoadedMsg{err: err}
		}

		configPath := filepath.Join(filepath.Dir(dbPath), "plugins.json")
		pm := vault.NewPluginManager(configPath)
		cfg := pm.Config()

		return pluginsLoadedMsg{hooks: cfg.Hooks, providers: cfg.Providers}
	}
}

func (p pluginsModel) Update(msg tea.Msg) (pluginsModel, tea.Cmd) {
	if msg, ok := msg.(pluginsLoadedMsg); ok {
		p.loaded = true
		if msg.err == nil {
			p.hooks = msg.hooks
			p.providers = msg.providers

			rows := make([]table.Row, 0, len(p.hooks)+len(p.providers))
			for _, h := range p.hooks {
				rows = append(rows, table.Row{
					"hook",
					string(h.Event),
					h.Command,
					fmt.Sprintf("%v", h.Args),
				})
			}

			for _, pr := range p.providers {
				rows = append(rows, table.Row{
					"provider",
					pr.Name,
					pr.Command,
					"prefix: " + pr.Prefix,
				})
			}

			p.table.SetRows(rows)
		} else {
			p.message = errorStyle.Render(msg.err.Error())
		}
	}

	return p, nil
}

func (p pluginsModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Plugins  "))
	b.WriteString("\n\n")

	if !p.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(p.hooks) == 0 && len(p.providers) == 0 {
		b.WriteString(dimStyle.Render("  No plugins configured."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: sidebar"))

		return b.String()
	}

	b.WriteString(p.table.View())
	b.WriteString("\n")

	if p.message != "" {
		b.WriteString("  " + p.message + "\n")
	}

	b.WriteString(helpStyle.Render("  esc: sidebar"))

	return b.String()
}
