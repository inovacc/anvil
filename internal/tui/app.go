package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

type screen int

const (
	screenDashboard screen = iota
	screenProfiles
	screenSecrets
	screenSecretForm
)

type model struct {
	vault         *vault.Vault
	currentScreen screen
	dashboard     dashboardModel
	profiles      profilesModel
	secrets       secretsModel
	secretForm    secretFormModel
	width, height int
	err           error
}

// NewModel creates the root TUI model.
func NewModel(v *vault.Vault) model {
	return model{
		vault:         v,
		currentScreen: screenDashboard,
	}
}

func (m model) Init() tea.Cmd {
	return m.dashboard.loadData(m.vault)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle screen-specific input
		switch m.currentScreen {
		case screenDashboard:
			return m.updateDashboard(msg)
		case screenProfiles:
			return m.updateProfiles(msg)
		case screenSecrets:
			return m.updateSecrets(msg)
		case screenSecretForm:
			return m.updateSecretForm(msg)
		}
	}

	// After profile/secret actions, reload the list
	switch msg.(type) {
	case profileActionMsg:
		m.profiles, _ = m.profiles.Update(msg)
		return m, m.profiles.loadData(m.vault)
	case secretActionMsg:
		m.secrets, _ = m.secrets.Update(msg)
		return m, m.secrets.loadData(m.vault, m.secrets.profileName)
	}

	// Delegate non-key messages to current screen
	switch m.currentScreen {
	case screenDashboard:
		var cmd tea.Cmd
		m.dashboard, cmd = m.dashboard.Update(msg)
		return m, cmd
	case screenProfiles:
		var cmd tea.Cmd
		m.profiles, cmd = m.profiles.Update(msg)
		return m, cmd
	case screenSecrets:
		var cmd tea.Cmd
		m.secrets, cmd = m.secrets.Update(msg)
		return m, cmd
	case screenSecretForm:
		var cmd tea.Cmd
		m.secretForm, cmd = m.secretForm.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.currentScreen {
	case screenDashboard:
		return m.dashboard.View(m.width)
	case screenProfiles:
		return m.profiles.View(m.width)
	case screenSecrets:
		return m.secrets.View(m.width)
	case screenSecretForm:
		return m.secretForm.View(m.width)
	}
	return ""
}

func (m model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "p":
		m.currentScreen = screenProfiles
		return m, m.profiles.loadData(m.vault)
	}
	return m, nil
}

func (m model) updateProfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.currentScreen = screenDashboard
		m.profiles.message = ""
		return m, m.dashboard.loadData(m.vault)
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "up" || msg.String() == "k":
		if m.profiles.confirm == "" && m.profiles.cursor > 0 {
			m.profiles.cursor--
			m.profiles.message = ""
		}
	case msg.String() == "down" || msg.String() == "j":
		if m.profiles.confirm == "" && m.profiles.cursor < len(m.profiles.profiles)-1 {
			m.profiles.cursor++
			m.profiles.message = ""
		}
	case msg.String() == "enter":
		if name := m.profiles.selectedProfile(); name != "" && m.profiles.confirm == "" {
			m.currentScreen = screenSecrets
			m.secrets = secretsModel{profileName: name}
			return m, m.secrets.loadData(m.vault, name)
		}
	case msg.String() == "c":
		if m.profiles.confirm == "" {
			m.currentScreen = screenSecretForm
			m.secretForm = newProfileFormModel()
			return m, m.secretForm.inputs[0].Focus()
		}
	case msg.String() == "n":
		if m.profiles.confirm != "" {
			m.profiles.confirm = ""
		} else {
			m.currentScreen = screenSecretForm
			m.secretForm = newProfileFormModel()
			return m, m.secretForm.inputs[0].Focus()
		}
	case msg.String() == "d":
		if m.profiles.confirm == "" {
			if name := m.profiles.selectedProfile(); name != "" {
				m.profiles.confirm = name
			}
		}
	case msg.String() == "y":
		if m.profiles.confirm != "" {
			name := m.profiles.confirm
			v := m.vault
			return m, func() tea.Msg {
				if err := v.DeleteProfile(name); err != nil {
					return profileActionMsg{err: err}
				}
				return profileActionMsg{action: "delete", message: fmt.Sprintf("Profile %q deleted.", name)}
			}
		}
	case msg.String() == "u":
		if m.profiles.confirm == "" {
			if name := m.profiles.selectedProfile(); name != "" {
				v := m.vault
				return m, func() tea.Msg {
					if err := v.UseProfile(name); err != nil {
						return profileActionMsg{err: err}
					}
					return profileActionMsg{action: "use", message: fmt.Sprintf("Now using profile %q.", name)}
				}
			}
		}
	}

	return m, nil
}

// newProfileFormModel reuses the secret form for creating profiles (key=name, value unused, desc=description).
func newProfileFormModel() secretFormModel {
	inputs := make([]textinput.Model, formFieldCount)

	nameInput := textinput.New()
	nameInput.Placeholder = "profile name"
	nameInput.CharLimit = 128
	nameInput.Width = 40
	nameInput.PromptStyle = focusedInputStyle
	nameInput.TextStyle = focusedInputStyle
	inputs[formFieldKey] = nameInput

	// Hidden — not used for profiles but keeps indexing consistent
	valInput := textinput.New()
	valInput.Placeholder = "(not used)"
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
		profileName: "__create_profile__",
		inputs:      inputs,
	}
}

func (m model) updateSecrets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc":
		m.currentScreen = screenProfiles
		m.secrets.message = ""
		return m, m.profiles.loadData(m.vault)
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "up" || msg.String() == "k":
		if m.secrets.confirm == "" && m.secrets.cursor > 0 {
			m.secrets.cursor--
			m.secrets.revealKey = ""
			m.secrets.message = ""
		}
	case msg.String() == "down" || msg.String() == "j":
		if m.secrets.confirm == "" && m.secrets.cursor < len(m.secrets.secrets)-1 {
			m.secrets.cursor++
			m.secrets.revealKey = ""
			m.secrets.message = ""
		}
	case msg.String() == "enter":
		if m.secrets.confirm == "" {
			key := m.secrets.selectedKey()
			if key != "" {
				// Toggle reveal
				if m.secrets.revealKey == key {
					m.secrets.revealKey = ""
					m.secrets.revealValue = ""
					return m, nil
				}
				profile := m.secrets.profileName
				v := m.vault
				return m, func() tea.Msg {
					val, err := v.Get(key, profile)
					if err != nil {
						return secretRevealMsg{key: key, err: err}
					}
					return secretRevealMsg{key: key, value: val}
				}
			}
		}
	case msg.String() == "c":
		if m.secrets.confirm == "" {
			m.currentScreen = screenSecretForm
			m.secretForm = newSecretFormModel(m.secrets.profileName)
			return m, m.secretForm.inputs[0].Focus()
		}
	case msg.String() == "n":
		if m.secrets.confirm != "" {
			m.secrets.confirm = ""
		} else {
			m.currentScreen = screenSecretForm
			m.secretForm = newSecretFormModel(m.secrets.profileName)
			return m, m.secretForm.inputs[0].Focus()
		}
	case msg.String() == "d":
		if m.secrets.confirm == "" {
			if key := m.secrets.selectedKey(); key != "" {
				m.secrets.confirm = key
			}
		}
	case msg.String() == "y":
		if m.secrets.confirm != "" {
			key := m.secrets.confirm
			profile := m.secrets.profileName
			v := m.vault
			return m, func() tea.Msg {
				if err := v.Delete(key, profile); err != nil {
					return secretActionMsg{err: err}
				}
				return secretActionMsg{message: fmt.Sprintf("Secret %q deleted.", key)}
			}
		}
	}

	return m, nil
}

func (m model) updateSecretForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.secretForm.profileName == "__create_profile__" {
			m.currentScreen = screenProfiles
		} else {
			m.currentScreen = screenSecrets
		}
		return m, nil
	case "enter":
		if m.secretForm.profileName == "__create_profile__" {
			// Create profile
			name := m.secretForm.key()
			desc := m.secretForm.desc()
			if name == "" {
				m.secretForm.message = errorStyle.Render("Name is required.")
				return m, nil
			}
			v := m.vault
			m.currentScreen = screenProfiles
			return m, func() tea.Msg {
				if err := v.CreateProfile(name, desc, false); err != nil {
					return profileActionMsg{err: err}
				}
				return profileActionMsg{action: "create", message: fmt.Sprintf("Profile %q created.", name)}
			}
		}
		// Create secret
		key := m.secretForm.key()
		value := m.secretForm.value()
		desc := m.secretForm.desc()
		if key == "" || value == "" {
			m.secretForm.message = errorStyle.Render("Key and value are required.")
			return m, nil
		}
		profile := m.secretForm.profileName
		v := m.vault
		m.currentScreen = screenSecrets
		m.secrets = secretsModel{profileName: profile}
		return m, tea.Batch(
			func() tea.Msg {
				if err := v.Set(key, value, profile, desc); err != nil {
					return secretActionMsg{err: err}
				}
				return secretActionMsg{message: fmt.Sprintf("Secret %q saved.", key)}
			},
			m.secrets.loadData(m.vault, profile),
		)
	}

	// Delegate to form
	var cmd tea.Cmd
	m.secretForm, cmd = m.secretForm.Update(msg)
	return m, cmd
}
