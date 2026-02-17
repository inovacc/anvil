package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type importExportModel struct {
	mode        int // 0 = export, 1 = import
	formatIndex int
	formats     []string
	fileInput   textinput.Model
	message     string
}

type importExportActionMsg struct {
	message string
	err     error
}

func newImportExportModel() importExportModel {
	fi := textinput.New()
	fi.Placeholder = "file path"
	fi.CharLimit = 512
	fi.Width = 40
	fi.PromptStyle = focusedInputStyle
	fi.TextStyle = focusedInputStyle

	return importExportModel{
		formats:   []string{"json", "env"},
		fileInput: fi,
	}
}

func (ie importExportModel) Update(msg tea.Msg) (importExportModel, tea.Cmd) {
	if msg, ok := msg.(importExportActionMsg); ok {
		if msg.err != nil {
			ie.message = errorStyle.Render(msg.err.Error())
		} else {
			ie.message = successStyle.Render(msg.message)
		}

		return ie, nil
	}

	var cmd tea.Cmd

	ie.fileInput, cmd = ie.fileInput.Update(msg)

	return ie, cmd
}

func (ie importExportModel) handleKey(msg tea.KeyMsg, v *vault.Vault) (importExportModel, tea.Cmd) {
	switch msg.String() {
	case "m":
		ie.mode = (ie.mode + 1) % 2
	case "f":
		ie.formatIndex = (ie.formatIndex + 1) % len(ie.formats)
	case "enter":
		file := ie.fileInput.Value()
		if file == "" {
			ie.message = errorStyle.Render("File path is required.")
			return ie, nil
		}

		format := ie.formats[ie.formatIndex]
		if ie.mode == 0 {
			// Export
			return ie, func() tea.Msg {
				entries, err := v.Export("")
				if err != nil {
					return importExportActionMsg{err: err}
				}

				var data []byte

				switch format {
				case "json":
					data, err = json.MarshalIndent(entries, "", "  ")
					if err != nil {
						return importExportActionMsg{err: err}
					}
				case "env":
					var sb strings.Builder
					for _, e := range entries {
						_, _ = fmt.Fprintf(&sb, "%s=%s\n", e.Key, e.Value)
					}

					data = []byte(sb.String())
				}

				if err := os.WriteFile(file, data, 0600); err != nil {
					return importExportActionMsg{err: fmt.Errorf("write: %w", err)}
				}

				return importExportActionMsg{message: fmt.Sprintf("Exported %d secrets to %s", len(entries), file)}
			}
		}
		// Import
		return ie, func() tea.Msg {
			data, err := os.ReadFile(file)
			if err != nil {
				return importExportActionMsg{err: fmt.Errorf("read: %w", err)}
			}

			var entries []vault.SecretEntry

			switch format {
			case "json":
				if err := json.Unmarshal(data, &entries); err != nil {
					return importExportActionMsg{err: fmt.Errorf("parse json: %w", err)}
				}
			case "env":
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}

					k, val, ok := strings.Cut(line, "=")
					if ok {
						entries = append(entries, vault.SecretEntry{Key: k, Value: val})
					}
				}
			}

			if err := v.Import(entries, ""); err != nil {
				return importExportActionMsg{err: err}
			}

			return importExportActionMsg{message: fmt.Sprintf("Imported %d secrets.", len(entries))}
		}
	}

	var cmd tea.Cmd

	ie.fileInput, cmd = ie.fileInput.Update(msg)

	return ie, cmd
}

func (ie importExportModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Import / Export  "))
	b.WriteString("\n\n")

	modeLabel := "Export"
	if ie.mode == 1 {
		modeLabel = "Import"
	}

	b.WriteString("  " + labelStyle.Render("Mode: ") + selectedStyle.Render(modeLabel) + dimStyle.Render("  (m to toggle)"))
	b.WriteString("\n")
	b.WriteString("  " + labelStyle.Render("Format: ") + selectedStyle.Render(ie.formats[ie.formatIndex]) + dimStyle.Render("  (f to toggle)"))
	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("File Path") + "\n")
	b.WriteString("  " + ie.fileInput.View() + "\n\n")

	if ie.message != "" {
		b.WriteString("  " + ie.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  m: mode • f: format • enter: execute • esc: sidebar"))

	return b.String()
}
