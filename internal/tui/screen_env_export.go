package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type envExportModel struct {
	formatIndex int
	formats     []string
	output      string
	message     string
}

type envExportResultMsg struct {
	output string
	err    error
}

func newEnvExportModel() envExportModel {
	return envExportModel{
		formats: []string{"env", "export", "json", "powershell"},
	}
}

func (e envExportModel) Update(msg tea.Msg) (envExportModel, tea.Cmd) {
	if msg, ok := msg.(envExportResultMsg); ok {
		if msg.err != nil {
			e.message = errorStyle.Render(msg.err.Error())
			e.output = ""
		} else {
			e.output = msg.output
			e.message = ""
		}
	}

	return e, nil
}

func (e envExportModel) handleKey(msg tea.KeyMsg) (envExportModel, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if e.formatIndex < len(e.formats)-1 {
			e.formatIndex++
		}
	case "k", "up":
		if e.formatIndex > 0 {
			e.formatIndex--
		}
	case "enter":
		format := e.formats[e.formatIndex]

		return e, withVault(func(v *vault.Vault) tea.Msg {
			entries, err := v.EnvExport("")
			if err != nil {
				return envExportResultMsg{err: err}
			}

			var out strings.Builder

			for _, en := range entries {
				switch format {
				case "env":
					_, _ = fmt.Fprintf(&out, "%s=%s\n", en.Key, en.Value)
				case "export":
					_, _ = fmt.Fprintf(&out, "export %s=%q\n", en.Key, en.Value)
				case "powershell":
					_, _ = fmt.Fprintf(&out, "$env:%s=%q\n", en.Key, en.Value)
				case "json":
					_, _ = fmt.Fprintf(&out, "  %q: %q,\n", en.Key, en.Value)
				}
			}

			return envExportResultMsg{output: out.String()}
		})
	}

	return e, nil
}

func (e envExportModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Env Export  "))
	b.WriteString("\n\n")

	b.WriteString("  " + labelStyle.Render("Format:") + "\n")

	for i, f := range e.formats {
		marker := "  "
		if i == e.formatIndex {
			marker = "> "
			b.WriteString(selectedStyle.Render("  " + marker + f))
		} else {
			b.WriteString(normalStyle.Render("  " + marker + f))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")

	if e.output != "" {
		b.WriteString(borderStyle.Width(min(width-6, 70)).Render(e.output))
		b.WriteString("\n\n")
	}

	if e.message != "" {
		b.WriteString("  " + e.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  j/k: select format • enter: export • esc: sidebar"))

	return b.String()
}
