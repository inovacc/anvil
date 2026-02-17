package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type templateApplyModel struct {
	templates     []vault.TemplateInfo
	templateIndex int
	profileInput  textinput.Model
	message       string
	loaded        bool
}

type templateApplyLoadedMsg struct {
	templates []vault.TemplateInfo
	err       error
}

type templateApplyActionMsg struct {
	message string
	err     error
}

func newTemplateApplyModel() templateApplyModel {
	pi := textinput.New()
	pi.Placeholder = "profile name (optional)"
	pi.CharLimit = 128
	pi.Width = 40
	pi.PromptStyle = focusedInputStyle
	pi.TextStyle = focusedInputStyle

	return templateApplyModel{profileInput: pi}
}

func (t templateApplyModel) loadData() tea.Cmd {
	return withVault(func(v *vault.Vault) tea.Msg {
		templates, err := v.ListTemplates()
		if err != nil {
			return templateApplyLoadedMsg{err: err}
		}

		return templateApplyLoadedMsg{templates: templates}
	})
}

func (t templateApplyModel) Update(msg tea.Msg) (templateApplyModel, tea.Cmd) {
	switch msg := msg.(type) {
	case templateApplyLoadedMsg:
		t.loaded = true
		if msg.err == nil {
			t.templates = msg.templates
		} else {
			t.message = errorStyle.Render(msg.err.Error())
		}
	case templateApplyActionMsg:
		if msg.err != nil {
			t.message = errorStyle.Render(msg.err.Error())
		} else {
			t.message = successStyle.Render(msg.message)
		}
	}

	var cmd tea.Cmd

	t.profileInput, cmd = t.profileInput.Update(msg)

	return t, cmd
}

func (t templateApplyModel) handleKey(msg tea.KeyMsg) (templateApplyModel, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if t.templateIndex < len(t.templates)-1 {
			t.templateIndex++
		}
	case "k", "up":
		if t.templateIndex > 0 {
			t.templateIndex--
		}
	case "enter":
		if len(t.templates) == 0 {
			return t, nil
		}

		name := t.templates[t.templateIndex].Name
		profile := t.profileInput.Value()

		return t, withVault(func(v *vault.Vault) tea.Msg {
			if err := v.ApplyTemplate(name, profile, nil); err != nil {
				return templateApplyActionMsg{err: err}
			}

			return templateApplyActionMsg{message: "Template \"" + name + "\" applied."}
		})
	}

	var cmd tea.Cmd

	t.profileInput, cmd = t.profileInput.Update(msg)

	return t, cmd
}

func (t templateApplyModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Apply Template  "))
	b.WriteString("\n\n")

	if !t.loaded {
		b.WriteString("  Loading...")
		return b.String()
	}

	if len(t.templates) == 0 {
		b.WriteString(dimStyle.Render("  No templates available."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("  esc: sidebar"))

		return b.String()
	}

	b.WriteString("  " + labelStyle.Render("Template:") + "\n")

	for i, tmpl := range t.templates {
		marker := "  "
		if i == t.templateIndex {
			marker = "> "
			b.WriteString(selectedStyle.Render("  " + marker + tmpl.Name))
		} else {
			b.WriteString(normalStyle.Render("  " + marker + tmpl.Name))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")

	b.WriteString("  " + inputLabelStyle.Render("Profile") + "\n")
	b.WriteString("  " + t.profileInput.View() + "\n\n")

	if t.message != "" {
		b.WriteString("  " + t.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  j/k: select • enter: apply • esc: sidebar"))

	return b.String()
}
