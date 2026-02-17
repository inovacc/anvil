package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type secretFormModel struct {
	profileName string
	inputs      []textinput.Model
	focusIndex  int
	submitted   bool
	message     string
}

const (
	formFieldKey = iota
	formFieldValue
	formFieldDesc
	formFieldCount
)

func newSecretFormModel(profile string) secretFormModel {
	inputs := make([]textinput.Model, formFieldCount)

	keyInput := textinput.New()
	keyInput.Placeholder = "secret key"
	keyInput.CharLimit = 256
	keyInput.Width = 40
	keyInput.Focus()
	keyInput.PromptStyle = focusedInputStyle
	keyInput.TextStyle = focusedInputStyle
	inputs[formFieldKey] = keyInput

	valInput := textinput.New()
	valInput.Placeholder = "secret value"
	valInput.CharLimit = 4096
	valInput.Width = 40
	valInput.PromptStyle = blurredInputStyle
	valInput.TextStyle = blurredInputStyle
	inputs[formFieldValue] = valInput

	descInput := textinput.New()
	descInput.Placeholder = "description (optional)"
	descInput.CharLimit = 256
	descInput.Width = 40
	descInput.PromptStyle = blurredInputStyle
	descInput.TextStyle = blurredInputStyle
	inputs[formFieldDesc] = descInput

	return secretFormModel{
		profileName: profile,
		inputs:      inputs,
	}
}

type secretFormSubmitMsg struct {
	err error
}

func (f secretFormModel) Update(msg tea.Msg) (secretFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if msg.String() == "tab" {
				f.focusIndex = (f.focusIndex + 1) % formFieldCount
			} else {
				f.focusIndex = (f.focusIndex - 1 + formFieldCount) % formFieldCount
			}

			cmds := make([]tea.Cmd, formFieldCount)

			for i := range f.inputs {
				if i == f.focusIndex {
					cmds[i] = f.inputs[i].Focus()
					f.inputs[i].PromptStyle = focusedInputStyle
					f.inputs[i].TextStyle = focusedInputStyle
				} else {
					f.inputs[i].Blur()
					f.inputs[i].PromptStyle = blurredInputStyle
					f.inputs[i].TextStyle = blurredInputStyle
				}
			}

			return f, tea.Batch(cmds...)
		}
	case secretFormSubmitMsg:
		if msg.err != nil {
			f.message = errorStyle.Render(msg.err.Error())
		} else {
			f.submitted = true
		}

		return f, nil
	}

	// Update the focused input
	var cmd tea.Cmd

	f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)

	return f, cmd
}

func (f secretFormModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("  New Secret [%s]  ", f.profileName)))
	b.WriteString("\n\n")

	for i, label := range []string{"Key", "Value", "Description"} {
		b.WriteString(fmt.Sprintf("  %s\n", labelStyle.Render(label)))
		b.WriteString(fmt.Sprintf("  %s\n\n", f.inputs[i].View()))
	}

	if f.message != "" {
		b.WriteString("  " + f.message + "\n")
	}

	b.WriteString(helpStyle.Render("  tab: next field • enter: save • esc: cancel"))

	return b.String()
}

func (f secretFormModel) key() string   { return strings.TrimSpace(f.inputs[formFieldKey].Value()) }
func (f secretFormModel) value() string { return f.inputs[formFieldValue].Value() }
func (f secretFormModel) desc() string  { return strings.TrimSpace(f.inputs[formFieldDesc].Value()) }
