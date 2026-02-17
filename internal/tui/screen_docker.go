package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type dockerScreenModel struct {
	dirInput textinput.Model
	message  string
}

type dockerActionMsg struct {
	message string
	err     error
}

func newDockerScreenModel() dockerScreenModel {
	di := textinput.New()
	di.Placeholder = "./secrets"
	di.CharLimit = 512
	di.Width = 40
	di.PromptStyle = focusedInputStyle
	di.TextStyle = focusedInputStyle

	return dockerScreenModel{dirInput: di}
}

func (d dockerScreenModel) Update(msg tea.Msg) (dockerScreenModel, tea.Cmd) {
	if msg, ok := msg.(dockerActionMsg); ok {
		if msg.err != nil {
			d.message = errorStyle.Render(msg.err.Error())
		} else {
			d.message = successStyle.Render(msg.message)
		}

		return d, nil
	}

	var cmd tea.Cmd

	d.dirInput, cmd = d.dirInput.Update(msg)

	return d, cmd
}

func (d dockerScreenModel) handleKey(msg tea.KeyMsg) (dockerScreenModel, tea.Cmd) {
	switch msg.String() {
	case "e":
		dir := d.dirInput.Value()
		if dir == "" {
			dir = "./secrets"
		}

		return d, withVault(func(v *vault.Vault) tea.Msg {
			entries, err := v.Export("")
			if err != nil {
				return dockerActionMsg{err: err}
			}

			if err := os.MkdirAll(dir, 0o700); err != nil {
				return dockerActionMsg{err: fmt.Errorf("create dir: %w", err)}
			}

			for _, e := range entries {
				if err := os.WriteFile(filepath.Join(dir, e.Key), []byte(e.Value), 0o600); err != nil {
					return dockerActionMsg{err: fmt.Errorf("write %q: %w", e.Key, err)}
				}
			}

			return dockerActionMsg{message: fmt.Sprintf("Exported %d secrets to %s", len(entries), dir)}
		})
	case "c":
		dir := d.dirInput.Value()
		if dir == "" {
			dir = "./secrets"
		}

		return d, withVault(func(v *vault.Vault) tea.Msg {
			secrets, err := v.List("")
			if err != nil {
				return dockerActionMsg{err: err}
			}

			var out strings.Builder
			out.WriteString("secrets:\n")

			for _, s := range secrets {
				_, _ = fmt.Fprintf(&out, "  %s:\n    file: %s\n", s.Key, filepath.Join(dir, s.Key))
			}

			return dockerActionMsg{message: "Compose snippet:\n" + out.String()}
		})
	case "x":
		dir := d.dirInput.Value()
		if dir == "" {
			dir = "./secrets"
		}

		return d, withVault(func(v *vault.Vault) tea.Msg {
			secrets, err := v.List("")
			if err != nil {
				return dockerActionMsg{err: err}
			}

			var removed int

			for _, s := range secrets {
				if err := os.Remove(filepath.Join(dir, s.Key)); err == nil {
					removed++
				}
			}

			return dockerActionMsg{message: fmt.Sprintf("Removed %d secret files from %s", removed, dir)}
		})
	}

	var cmd tea.Cmd

	d.dirInput, cmd = d.dirInput.Update(msg)

	return d, cmd
}

func (d dockerScreenModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Docker  "))
	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("Directory") + "\n")
	b.WriteString("  " + d.dirInput.View() + "\n\n")

	if d.message != "" {
		b.WriteString("  " + d.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  e: export • c: compose • x: clean • esc: sidebar"))

	return b.String()
}
