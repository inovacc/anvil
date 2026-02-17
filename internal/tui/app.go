package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/inovacc/anvil/pkg/vault"
)

type screen int

const (
	screenStatus screen = iota
	screenProfiles
	screenSecrets
	screenSecretForm
	screenImportExport
	screenPassword
	screenKeyRotation
	screenSeal
	screenEnvRelease
	screenEnvStatus
	screenEnvExport
	screenTemplates
	screenTemplateApply
	screenAudit
	screenHistory
	screenDocker
	screenShare
	screenPlugins
	screenGather
	screenBackup
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusContent
)

type model struct { //nolint:recvcheck // pointer receiver needed for resizeCurrentScreen
	sidebar       sidebarModel
	focus         focusArea
	currentScreen screen

	// Existing screens
	profiles   profilesModel
	secrets    secretsModel
	secretForm secretFormModel
	audit      auditModel
	history    historyModel

	// New screens
	status        statusModel
	password      passwordScreenModel
	keyRotation   keyRotationModel
	seal          sealScreenModel
	envRelease    envReleaseModel
	envStatus     envStatusModel
	envExport     envExportModel
	templates     templatesModel
	templateApply templateApplyModel
	backup        backupScreenModel
	share         shareScreenModel
	docker        dockerScreenModel
	plugins       pluginsModel
	gather        gatherScreenModel
	importExport  importExportModel

	width, height int
}

// NewModel creates the root TUI model.
func NewModel() model {
	sb := newSidebarModel()
	// Default to Status screen
	sb.moveCursorToScreen(screenStatus)

	m := model{
		sidebar:       sb,
		currentScreen: screenStatus,
		focus:         focusSidebar,
	}
	m.initScreen(screenStatus)

	return m
}

func (m model) Init() tea.Cmd {
	return m.loadScreenData(screenStatus)
}

// initScreen initializes the model for the given screen. Must be called on
// the model that will be returned (not a copy).
func (m *model) initScreen(sc screen) {
	w := m.contentWidth()
	h := m.contentHeight()

	switch sc { //nolint:exhaustive // screenSecretForm initialized elsewhere
	case screenStatus:
		m.status = newStatusModel()
	case screenProfiles:
		m.profiles = newProfilesModel(w, h)
	case screenSecrets:
		m.secrets = newSecretsModel(m.secrets.profileName, w, h)
	case screenAudit:
		m.audit = newAuditModel(w, h)
	case screenHistory:
		m.history = newHistoryModel(m.history.key, m.history.profileName, w, h)
	case screenPassword:
		m.password = newPasswordScreenModel()
	case screenKeyRotation:
		m.keyRotation = newKeyRotationModel()
	case screenSeal:
		m.seal = newSealScreenModel()
	case screenEnvRelease:
		m.envRelease = newEnvReleaseModel()
	case screenEnvStatus:
		m.envStatus = newEnvStatusModel()
	case screenEnvExport:
		m.envExport = newEnvExportModel()
	case screenTemplates:
		m.templates = newTemplatesModel(w, h)
	case screenTemplateApply:
		m.templateApply = newTemplateApplyModel()
	case screenBackup:
		m.backup = newBackupScreenModel()
	case screenShare:
		m.share = newShareScreenModel()
	case screenDocker:
		m.docker = newDockerScreenModel()
	case screenPlugins:
		m.plugins = newPluginsModel(w, h)
	case screenGather:
		m.gather = newGatherScreenModel()
	case screenImportExport:
		m.importExport = newImportExportModel()
	}
}

// loadScreenData returns a command that fetches data for the given screen.
func (m model) loadScreenData(sc screen) tea.Cmd {
	switch sc { //nolint:exhaustive // not all screens need async data
	case screenStatus:
		return m.status.loadData()
	case screenProfiles:
		return m.profiles.loadData()
	case screenSecrets:
		return m.secrets.loadData(m.secrets.profileName)
	case screenAudit:
		return m.audit.loadData()
	case screenHistory:
		return m.history.loadData()
	case screenPassword:
		return m.password.loadData()
	case screenSeal:
		return m.seal.loadData()
	case screenEnvStatus:
		return m.envStatus.loadData()
	case screenTemplates:
		return m.templates.loadData()
	case screenTemplateApply:
		return m.templateApply.loadData()
	case screenPlugins:
		return m.plugins.loadData()
	}

	return nil
}

func (m model) contentWidth() int {
	w := max(
		// sidebar + border + padding
		m.width-sidebarWidth-3, 40)

	return w
}

func (m model) contentHeight() int {
	if m.height < 10 {
		return 20
	}

	return m.height
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeCurrentScreen()

		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Tab toggles focus
		if msg.Type == tea.KeyTab && m.currentScreen != screenSecretForm {
			if m.focus == focusSidebar {
				m.focus = focusContent
				m.sidebar.focused = false
			} else {
				m.focus = focusSidebar
				m.sidebar.focused = true
			}

			return m, nil
		}

		if m.focus == focusSidebar {
			return m.updateSidebar(msg)
		}

		return m.updateContent(msg)

	case vaultErrorMsg:
		// Surface vault open errors as status messages on the current screen
		// For now, just ignore — individual screens handle their own errors
		return m, nil
	}

	// Delegate non-key messages to current screen
	return m.delegateMsg(msg)
}

func (m model) updateSidebar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "j", "down":
		m.sidebar.moveDown()
		return m, nil
	case "k", "up":
		m.sidebar.moveUp()
		return m, nil
	case "enter":
		sc := m.sidebar.selectedScreen()
		if sc != m.currentScreen {
			m.currentScreen = sc
			m.focus = focusContent
			m.sidebar.focused = false
			m.initScreen(sc)

			return m, m.loadScreenData(sc)
		}

		m.focus = focusContent
		m.sidebar.focused = false

		return m, nil
	}

	return m, nil
}

func (m model) updateContent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Esc from content returns focus to sidebar (except in sub-views)
	if msg.String() == "esc" {
		switch m.currentScreen { //nolint:exhaustive // default handles all other screens
		case screenSecretForm:
			// Return to caller screen
			return m.updateSecretForm(msg)
		case screenHistory:
			// Return to secrets
			m.currentScreen = screenSecrets
			return m, nil
		default:
			m.focus = focusSidebar
			m.sidebar.focused = true

			return m, nil
		}
	}

	switch m.currentScreen {
	case screenProfiles:
		return m.updateProfiles(msg)
	case screenSecrets:
		return m.updateSecrets(msg)
	case screenSecretForm:
		return m.updateSecretForm(msg)
	case screenAudit:
		return m.updateAuditScreen(msg)
	case screenHistory:
		return m.updateHistoryScreen(msg)
	case screenStatus:
		return m.updateStatusScreen(msg)
	case screenPassword:
		return m.updatePasswordScreen(msg)
	case screenKeyRotation:
		return m.updateKeyRotationScreen(msg)
	case screenSeal:
		return m.updateSealScreen(msg)
	case screenEnvRelease:
		return m.updateEnvReleaseScreen(msg)
	case screenEnvStatus:
		return m.updateEnvStatusScreen(msg)
	case screenEnvExport:
		return m.updateEnvExportScreen(msg)
	case screenTemplates:
		return m.updateTemplatesScreen(msg)
	case screenTemplateApply:
		return m.updateTemplateApplyScreen(msg)
	case screenBackup:
		return m.updateBackupScreen(msg)
	case screenShare:
		return m.updateShareScreen(msg)
	case screenDocker:
		return m.updateDockerScreen(msg)
	case screenPlugins:
		return m.updatePluginsScreen(msg)
	case screenGather:
		return m.updateGatherScreen(msg)
	case screenImportExport:
		return m.updateImportExportScreen(msg)
	}

	return m, nil
}

func (m model) delegateMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	// After profile/secret actions, reload
	switch msg.(type) {
	case profileActionMsg:
		m.profiles, _ = m.profiles.Update(msg)
		return m, m.profiles.loadData()
	case secretActionMsg:
		m.secrets, _ = m.secrets.Update(msg)
		return m, m.secrets.loadData(m.secrets.profileName)
	}

	switch m.currentScreen {
	case screenStatus:
		var cmd tea.Cmd

		m.status, cmd = m.status.Update(msg)

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
	case screenAudit:
		var cmd tea.Cmd

		m.audit, cmd = m.audit.Update(msg)

		return m, cmd
	case screenHistory:
		var cmd tea.Cmd

		m.history, cmd = m.history.Update(msg)

		return m, cmd
	case screenPassword:
		var cmd tea.Cmd

		m.password, cmd = m.password.Update(msg)

		return m, cmd
	case screenKeyRotation:
		var cmd tea.Cmd

		m.keyRotation, cmd = m.keyRotation.Update(msg)

		return m, cmd
	case screenSeal:
		var cmd tea.Cmd

		m.seal, cmd = m.seal.Update(msg)

		return m, cmd
	case screenEnvRelease:
		var cmd tea.Cmd

		m.envRelease, cmd = m.envRelease.Update(msg)

		return m, cmd
	case screenEnvStatus:
		var cmd tea.Cmd

		m.envStatus, cmd = m.envStatus.Update(msg)

		return m, cmd
	case screenEnvExport:
		var cmd tea.Cmd

		m.envExport, cmd = m.envExport.Update(msg)

		return m, cmd
	case screenTemplates:
		var cmd tea.Cmd

		m.templates, cmd = m.templates.Update(msg)

		return m, cmd
	case screenTemplateApply:
		var cmd tea.Cmd

		m.templateApply, cmd = m.templateApply.Update(msg)

		return m, cmd
	case screenBackup:
		var cmd tea.Cmd

		m.backup, cmd = m.backup.Update(msg)

		return m, cmd
	case screenShare:
		var cmd tea.Cmd

		m.share, cmd = m.share.Update(msg)

		return m, cmd
	case screenDocker:
		var cmd tea.Cmd

		m.docker, cmd = m.docker.Update(msg)

		return m, cmd
	case screenPlugins:
		var cmd tea.Cmd

		m.plugins, cmd = m.plugins.Update(msg)

		return m, cmd
	case screenGather:
		var cmd tea.Cmd

		m.gather, cmd = m.gather.Update(msg)

		return m, cmd
	case screenImportExport:
		var cmd tea.Cmd

		m.importExport, cmd = m.importExport.Update(msg)

		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	sidebarView := m.sidebar.View(m.height)
	contentView := m.renderContent()

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, contentView)
}

func (m model) renderContent() string {
	w := m.contentWidth()
	switch m.currentScreen {
	case screenStatus:
		return m.status.View(w)
	case screenProfiles:
		return m.profiles.View(w)
	case screenSecrets:
		return m.secrets.View(w)
	case screenSecretForm:
		return m.secretForm.View(w)
	case screenAudit:
		return m.audit.View(w)
	case screenHistory:
		return m.history.View(w)
	case screenPassword:
		return m.password.View(w)
	case screenKeyRotation:
		return m.keyRotation.View(w)
	case screenSeal:
		return m.seal.View(w)
	case screenEnvRelease:
		return m.envRelease.View(w)
	case screenEnvStatus:
		return m.envStatus.View(w)
	case screenEnvExport:
		return m.envExport.View(w)
	case screenTemplates:
		return m.templates.View(w)
	case screenTemplateApply:
		return m.templateApply.View(w)
	case screenBackup:
		return m.backup.View(w)
	case screenShare:
		return m.share.View(w)
	case screenDocker:
		return m.docker.View(w)
	case screenPlugins:
		return m.plugins.View(w)
	case screenGather:
		return m.gather.View(w)
	case screenImportExport:
		return m.importExport.View(w)
	}

	return ""
}

func (m *model) resizeCurrentScreen() {
	w := m.contentWidth()

	h := m.contentHeight()
	switch m.currentScreen { //nolint:exhaustive // only table-based screens need resize
	case screenSecrets:
		m.secrets.table.SetWidth(w)
		m.secrets.table.SetHeight(max(1, h-8))
		m.secrets.table.SetColumns(tableColumns(w))
	case screenProfiles:
		m.profiles.table.SetWidth(w)
		m.profiles.table.SetHeight(max(1, h-8))
		m.profiles.table.SetColumns(profileTableColumns(w))
	case screenAudit:
		m.audit.table.SetWidth(w)
		m.audit.table.SetHeight(max(1, h-8))
		m.audit.table.SetColumns(auditTableColumns(w))
	case screenHistory:
		m.history.table.SetWidth(w)
		m.history.table.SetHeight(max(1, h-8))
		m.history.table.SetColumns(historyTableColumns(w))
	}
}

// --- Screen-specific update handlers ---

func (m model) updateProfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "enter":
		if name := m.profiles.selectedProfile(); name != "" && m.profiles.confirm == "" {
			m.currentScreen = screenSecrets
			m.secrets = newSecretsModel(name, m.contentWidth(), m.contentHeight())
			m.sidebar.moveCursorToScreen(screenSecrets)

			return m, m.secrets.loadData(name)
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

			return m, withVault(func(v *vault.Vault) tea.Msg {
				if err := v.DeleteProfile(name); err != nil {
					return profileActionMsg{err: err}
				}

				return profileActionMsg{action: "delete", message: fmt.Sprintf("Profile %q deleted.", name)}
			})
		}
	case msg.String() == "u":
		if m.profiles.confirm == "" {
			if name := m.profiles.selectedProfile(); name != "" {
				return m, withVault(func(v *vault.Vault) tea.Msg {
					if err := v.UseProfile(name); err != nil {
						return profileActionMsg{err: err}
					}

					return profileActionMsg{action: "use", message: fmt.Sprintf("Now using profile %q.", name)}
				})
			}
		}
	default:
		if m.profiles.confirm == "" {
			m.profiles.message = ""

			var cmd tea.Cmd

			m.profiles.table, cmd = m.profiles.table.Update(msg)

			return m, cmd
		}
	}

	return m, nil
}

// newProfileFormModel reuses the secret form for creating profiles.
func newProfileFormModel() secretFormModel {
	inputs := make([]textinput.Model, formFieldCount)

	nameInput := textinput.New()
	nameInput.Placeholder = "profile name"
	nameInput.CharLimit = 128
	nameInput.Width = 40
	nameInput.PromptStyle = focusedInputStyle
	nameInput.TextStyle = focusedInputStyle
	inputs[formFieldKey] = nameInput

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
	case msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "enter":
		if m.secrets.confirm == "" {
			key := m.secrets.selectedKey()
			if key != "" {
				if m.secrets.revealKey == key {
					m.secrets.revealKey = ""
					m.secrets.revealValue = ""

					return m, nil
				}

				profile := m.secrets.profileName

				return m, withVault(func(v *vault.Vault) tea.Msg {
					val, err := v.Get(key, profile)
					if err != nil {
						return secretRevealMsg{key: key, err: err}
					}

					return secretRevealMsg{key: key, value: val}
				})
			}
		}
	case msg.String() == "h":
		if m.secrets.confirm == "" {
			if key := m.secrets.selectedKey(); key != "" {
				m.currentScreen = screenHistory
				m.history = newHistoryModel(key, m.secrets.profileName, m.contentWidth(), m.contentHeight())
				m.sidebar.moveCursorToScreen(screenHistory)

				return m, m.history.loadData()
			}
		}
	case msg.String() == "r":
		// Rollback placeholder — requires history sub-view
		_ = m.secrets.confirm
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

			return m, withVault(func(v *vault.Vault) tea.Msg {
				if err := v.Delete(key, profile); err != nil {
					return secretActionMsg{err: err}
				}

				return secretActionMsg{message: fmt.Sprintf("Secret %q deleted.", key)}
			})
		}
	default:
		if m.secrets.confirm == "" {
			m.secrets.revealKey = ""
			m.secrets.message = ""

			var cmd tea.Cmd

			m.secrets.table, cmd = m.secrets.table.Update(msg)

			return m, cmd
		}
	}

	return m, nil
}

func (m model) updateSecretForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.secretForm.profileName == "__create_profile__" {
			m.currentScreen = screenProfiles
			m.sidebar.moveCursorToScreen(screenProfiles)
		} else {
			m.currentScreen = screenSecrets
			m.sidebar.moveCursorToScreen(screenSecrets)
		}

		return m, nil
	case "enter":
		if m.secretForm.profileName == "__create_profile__" {
			name := m.secretForm.key()

			desc := m.secretForm.desc()
			if name == "" {
				m.secretForm.message = errorStyle.Render("Name is required.")
				return m, nil
			}

			m.currentScreen = screenProfiles
			m.sidebar.moveCursorToScreen(screenProfiles)

			return m, withVault(func(v *vault.Vault) tea.Msg {
				if err := v.CreateProfile(name, desc, false); err != nil {
					return profileActionMsg{err: err}
				}

				return profileActionMsg{action: "create", message: fmt.Sprintf("Profile %q created.", name)}
			})
		}

		key := m.secretForm.key()
		value := m.secretForm.value()

		desc := m.secretForm.desc()
		if key == "" || value == "" {
			m.secretForm.message = errorStyle.Render("Key and value are required.")
			return m, nil
		}

		profile := m.secretForm.profileName
		m.currentScreen = screenSecrets
		m.sidebar.moveCursorToScreen(screenSecrets)
		m.secrets = newSecretsModel(profile, m.contentWidth(), m.contentHeight())

		return m, tea.Batch(
			withVault(func(v *vault.Vault) tea.Msg {
				if err := v.Set(key, value, profile, desc); err != nil {
					return secretActionMsg{err: err}
				}

				return secretActionMsg{message: fmt.Sprintf("Secret %q saved.", key)}
			}),
			m.secrets.loadData(profile),
		)
	}

	var cmd tea.Cmd

	m.secretForm, cmd = m.secretForm.Update(msg)

	return m, cmd
}

func (m model) updateAuditScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	default:
		var cmd tea.Cmd

		m.audit.table, cmd = m.audit.table.Update(msg)

		return m, cmd
	}
}

func (m model) updateHistoryScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "r":
		// Rollback selected version
		if len(m.history.versions) > 0 {
			row := m.history.table.SelectedRow()
			if len(row) > 0 {
				var version int64

				_, _ = fmt.Sscanf(row[0], "v%d", &version)
				if version > 0 {
					key := m.history.key
					profile := m.history.profileName

					return m, withVault(func(v *vault.Vault) tea.Msg {
						if err := v.SecretRollback(key, profile, version); err != nil {
							return historyLoadedMsg{err: err}
						}

						versions, err := v.SecretHistory(key, profile)
						if err != nil {
							return historyLoadedMsg{err: err}
						}

						return historyLoadedMsg{versions: versions}
					})
				}
			}
		}
	default:
		var cmd tea.Cmd

		m.history.table, cmd = m.history.table.Update(msg)

		return m, cmd
	}

	return m, nil
}

// Screen update stubs — implemented by each screen file

func (m model) updateStatusScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "q" {
		return m, tea.Quit
	}

	return m, nil
}

func (m model) updatePasswordScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.password, cmd = m.password.handleKey(msg)

	return m, cmd
}

func (m model) updateKeyRotationScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.keyRotation, cmd = m.keyRotation.handleKey(msg)

	return m, cmd
}

func (m model) updateSealScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.seal, cmd = m.seal.handleKey(msg)

	return m, cmd
}

func (m model) updateEnvReleaseScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.envRelease, cmd = m.envRelease.handleKey(msg)

	return m, cmd
}

func (m model) updateEnvStatusScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.envStatus, cmd = m.envStatus.handleKey(msg)

	return m, cmd
}

func (m model) updateEnvExportScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.envExport, cmd = m.envExport.handleKey(msg)

	return m, cmd
}

func (m model) updateTemplatesScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "d":
		if name := m.templates.selectedName(); name != "" {
			return m, withVault(func(v *vault.Vault) tea.Msg {
				if err := v.DeleteTemplate(name); err != nil {
					return templatesLoadedMsg{err: err}
				}

				templates, err := v.ListTemplates()
				if err != nil {
					return templatesLoadedMsg{err: err}
				}

				return templatesLoadedMsg{templates: templates}
			})
		}
	default:
		var cmd tea.Cmd

		m.templates.table, cmd = m.templates.table.Update(msg)

		return m, cmd
	}

	return m, nil
}

func (m model) updateTemplateApplyScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.templateApply, cmd = m.templateApply.handleKey(msg)

	return m, cmd
}

func (m model) updateBackupScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.backup, cmd = m.backup.handleKey(msg)

	return m, cmd
}

func (m model) updateShareScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.share, cmd = m.share.handleKey(msg)

	return m, cmd
}

func (m model) updateDockerScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.docker, cmd = m.docker.handleKey(msg)

	return m, cmd
}

func (m model) updatePluginsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	default:
		var cmd tea.Cmd

		m.plugins.table, cmd = m.plugins.table.Update(msg)

		return m, cmd
	}
}

func (m model) updateGatherScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.gather, cmd = m.gather.handleKey(msg)

	return m, cmd
}

func (m model) updateImportExportScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.importExport, cmd = m.importExport.handleKey(msg)

	return m, cmd
}
