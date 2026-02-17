package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type gatherScreenModel struct {
	dirInput textinput.Model
	message  string
}

type gatherActionMsg struct {
	message string
	err     error
}

func newGatherScreenModel() gatherScreenModel {
	di := textinput.New()
	di.Placeholder = ". (current directory)"
	di.CharLimit = 512
	di.Width = 40
	di.PromptStyle = focusedInputStyle
	di.TextStyle = focusedInputStyle
	return gatherScreenModel{dirInput: di}
}

func (g gatherScreenModel) Update(msg tea.Msg) (gatherScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case gatherActionMsg:
		if msg.err != nil {
			g.message = errorStyle.Render(msg.err.Error())
		} else {
			g.message = successStyle.Render(msg.message)
		}
		return g, nil
	}
	var cmd tea.Cmd
	g.dirInput, cmd = g.dirInput.Update(msg)
	return g, cmd
}

func (g gatherScreenModel) handleKey(msg tea.KeyMsg, _ *vault.Vault) (gatherScreenModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		dir := g.dirInput.Value()
		if dir == "" {
			dir = "."
		}
		g.message = dimStyle.Render("Gather requires CLI mode: anvil gather " + dir)
		return g, nil
	}
	var cmd tea.Cmd
	g.dirInput, cmd = g.dirInput.Update(msg)
	return g, cmd
}

func (g gatherScreenModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Gather Secrets  "))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("  Discover secrets from .env, JSON, and YAML files."))
	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("Directory") + "\n")
	b.WriteString("  " + g.dirInput.View() + "\n\n")

	if g.message != "" {
		b.WriteString("  " + g.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  enter: scan • esc: sidebar"))

	return b.String()
}
