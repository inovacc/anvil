package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type shareScreenModel struct {
	profileInput    textinput.Model
	passphraseInput textinput.Model
	fileInput       textinput.Model
	focusIndex      int
	message         string
}

type shareActionMsg struct {
	message string
	err     error
}

func newShareScreenModel() shareScreenModel {
	pi := textinput.New()
	pi.Placeholder = "profile name"
	pi.CharLimit = 128
	pi.Width = 40
	pi.PromptStyle = focusedInputStyle
	pi.TextStyle = focusedInputStyle

	pp := textinput.New()
	pp.Placeholder = "passphrase"
	pp.EchoMode = textinput.EchoPassword
	pp.CharLimit = 256
	pp.Width = 40
	pp.PromptStyle = blurredInputStyle
	pp.TextStyle = blurredInputStyle

	fi := textinput.New()
	fi.Placeholder = "file path"
	fi.CharLimit = 512
	fi.Width = 40
	fi.PromptStyle = blurredInputStyle
	fi.TextStyle = blurredInputStyle

	return shareScreenModel{
		profileInput:    pi,
		passphraseInput: pp,
		fileInput:       fi,
	}
}

func (s shareScreenModel) Update(msg tea.Msg) (shareScreenModel, tea.Cmd) {
	if msg, ok := msg.(shareActionMsg); ok {
		if msg.err != nil {
			s.message = errorStyle.Render(msg.err.Error())
		} else {
			s.message = successStyle.Render(msg.message)
		}

		return s, nil
	}

	var cmd tea.Cmd

	switch s.focusIndex {
	case 0:
		s.profileInput, cmd = s.profileInput.Update(msg)
	case 1:
		s.passphraseInput, cmd = s.passphraseInput.Update(msg)
	case 2:
		s.fileInput, cmd = s.fileInput.Update(msg)
	}

	return s, cmd
}

func (s shareScreenModel) handleKey(msg tea.KeyMsg, v *vault.Vault) (shareScreenModel, tea.Cmd) {
	inputs := []*textinput.Model{&s.profileInput, &s.passphraseInput, &s.fileInput}

	switch msg.String() {
	case "tab":
		inputs[s.focusIndex].Blur()
		inputs[s.focusIndex].PromptStyle = blurredInputStyle
		inputs[s.focusIndex].TextStyle = blurredInputStyle
		s.focusIndex = (s.focusIndex + 1) % 3
		inputs[s.focusIndex].PromptStyle = focusedInputStyle
		inputs[s.focusIndex].TextStyle = focusedInputStyle

		return s, inputs[s.focusIndex].Focus()
	case "e":
		profile := s.profileInput.Value()
		passphrase := s.passphraseInput.Value()

		file := s.fileInput.Value()
		if passphrase == "" || file == "" {
			s.message = errorStyle.Render("Passphrase and file are required.")
			return s, nil
		}

		return s, func() tea.Msg {
			encrypted, err := v.ShareExport(profile, passphrase)
			if err != nil {
				return shareActionMsg{err: err}
			}

			if err := os.WriteFile(file, encrypted, 0600); err != nil {
				return shareActionMsg{err: fmt.Errorf("write file: %w", err)}
			}

			return shareActionMsg{message: "Export written to " + file}
		}
	case "i":
		profile := s.profileInput.Value()
		passphrase := s.passphraseInput.Value()

		file := s.fileInput.Value()
		if passphrase == "" || file == "" {
			s.message = errorStyle.Render("Passphrase and file are required.")
			return s, nil
		}

		return s, func() tea.Msg {
			data, err := os.ReadFile(file)
			if err != nil {
				return shareActionMsg{err: fmt.Errorf("read file: %w", err)}
			}

			export, err := v.ShareImport(data, passphrase, profile)
			if err != nil {
				return shareActionMsg{err: err}
			}

			return shareActionMsg{message: fmt.Sprintf("Imported %d secrets from %q", len(export.Secrets), export.ProfileName)}
		}
	}

	var cmd tea.Cmd

	switch s.focusIndex {
	case 0:
		s.profileInput, cmd = s.profileInput.Update(msg)
	case 1:
		s.passphraseInput, cmd = s.passphraseInput.Update(msg)
	case 2:
		s.fileInput, cmd = s.fileInput.Update(msg)
	}

	return s, cmd
}

func (s shareScreenModel) View(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Share  "))
	b.WriteString("\n\n")

	b.WriteString("  " + inputLabelStyle.Render("Profile") + "\n")
	b.WriteString("  " + s.profileInput.View() + "\n\n")
	b.WriteString("  " + inputLabelStyle.Render("Passphrase") + "\n")
	b.WriteString("  " + s.passphraseInput.View() + "\n\n")
	b.WriteString("  " + inputLabelStyle.Render("File Path") + "\n")
	b.WriteString("  " + s.fileInput.View() + "\n\n")

	if s.message != "" {
		b.WriteString("  " + s.message + "\n\n")
	}

	b.WriteString(helpStyle.Render("  e: export • i: import • tab: switch field • esc: sidebar"))

	return b.String()
}
