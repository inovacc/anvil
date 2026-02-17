package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type backupScreenModel struct {
	fileInput     textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	message       string
}

type backupActionMsg struct {
	message string
	err     error
}

func newBackupScreenModel() backupScreenModel {
	fi := textinput.New()
	fi.Placeholder = "backup file path"
	fi.CharLimit = 512
	fi.Width = 40
	fi.PromptStyle = focusedInputStyle
	fi.TextStyle = focusedInputStyle

	pi := textinput.New()
	pi.Placeholder = "password"
	pi.EchoMode = textinput.EchoPassword
	pi.CharLimit = 256
	pi.Width = 40
	pi.PromptStyle = blurredInputStyle
	pi.TextStyle = blurredInputStyle

	return backupScreenModel{fileInput: fi, passwordInput: pi}
}

func (b backupScreenModel) Update(msg tea.Msg) (backupScreenModel, tea.Cmd) {
	switch msg := msg.(type) {
	case backupActionMsg:
		if msg.err != nil {
			b.message = errorStyle.Render(msg.err.Error())
		} else {
			b.message = successStyle.Render(msg.message)
		}
		return b, nil
	}
	var cmd tea.Cmd
	if b.focusIndex == 0 {
		b.fileInput, cmd = b.fileInput.Update(msg)
	} else {
		b.passwordInput, cmd = b.passwordInput.Update(msg)
	}
	return b, cmd
}

func (b backupScreenModel) handleKey(msg tea.KeyMsg, v *vault.Vault) (backupScreenModel, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if b.focusIndex == 0 {
			b.focusIndex = 1
			b.fileInput.Blur()
			b.fileInput.PromptStyle = blurredInputStyle
			b.fileInput.TextStyle = blurredInputStyle
			b.passwordInput.PromptStyle = focusedInputStyle
			b.passwordInput.TextStyle = focusedInputStyle
			return b, b.passwordInput.Focus()
		}
		b.focusIndex = 0
		b.passwordInput.Blur()
		b.passwordInput.PromptStyle = blurredInputStyle
		b.passwordInput.TextStyle = blurredInputStyle
		b.fileInput.PromptStyle = focusedInputStyle
		b.fileInput.TextStyle = focusedInputStyle
		return b, b.fileInput.Focus()
	case "b":
		file := b.fileInput.Value()
		pw := b.passwordInput.Value()
		if file == "" || pw == "" {
			b.message = errorStyle.Render("File path and password are required.")
			return b, nil
		}
		return b, func() tea.Msg {
			encrypted, err := v.Backup(pw)
			if err != nil {
				return backupActionMsg{err: err}
			}
			if err := os.WriteFile(file, encrypted, 0600); err != nil {
				return backupActionMsg{err: fmt.Errorf("write file: %w", err)}
			}
			return backupActionMsg{message: "Backup written to " + file}
		}
	case "r":
		file := b.fileInput.Value()
		pw := b.passwordInput.Value()
		if file == "" || pw == "" {
			b.message = errorStyle.Render("File path and password are required.")
			return b, nil
		}
		return b, func() tea.Msg {
			data, err := os.ReadFile(file)
			if err != nil {
				return backupActionMsg{err: fmt.Errorf("read file: %w", err)}
			}
			if err := v.Restore(data, pw); err != nil {
				return backupActionMsg{err: err}
			}
			return backupActionMsg{message: "Vault restored from " + file}
		}
	}
	var cmd tea.Cmd
	if b.focusIndex == 0 {
		b.fileInput, cmd = b.fileInput.Update(msg)
	} else {
		b.passwordInput, cmd = b.passwordInput.Update(msg)
	}
	return b, cmd
}

func (b backupScreenModel) View(width int) string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("  Backup / Restore  "))
	sb.WriteString("\n\n")

	sb.WriteString("  " + inputLabelStyle.Render("File Path") + "\n")
	sb.WriteString("  " + b.fileInput.View() + "\n\n")
	sb.WriteString("  " + inputLabelStyle.Render("Password") + "\n")
	sb.WriteString("  " + b.passwordInput.View() + "\n\n")

	if b.message != "" {
		sb.WriteString("  " + b.message + "\n\n")
	}

	sb.WriteString(helpStyle.Render("  b: backup • r: restore • tab: switch field • esc: sidebar"))

	return sb.String()
}
