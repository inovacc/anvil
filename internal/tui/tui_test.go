package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/pkg/vault"
)

func TestBoolToYesNo(t *testing.T) {
	tests := []struct {
		in   bool
		want string
	}{
		{true, "yes"},
		{false, "no"},
	}
	for _, tt := range tests {
		if got := boolToYesNo(tt.in); got != tt.want {
			t.Errorf("boolToYesNo(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTableColumns(t *testing.T) {
	tests := []struct {
		width    int
		wantCols int
	}{
		{0, 4},   // defaults to 80
		{80, 4},  // normal
		{200, 4}, // wide
	}
	for _, tt := range tests {
		cols := tableColumns(tt.width)
		if len(cols) != tt.wantCols {
			t.Errorf("tableColumns(%d): got %d cols, want %d", tt.width, len(cols), tt.wantCols)
		}
		if cols[0].Title != "Key" || cols[1].Title != "Description" || cols[2].Title != "Created" || cols[3].Title != "Updated" {
			t.Errorf("tableColumns(%d): wrong column titles", tt.width)
		}
	}
}

func TestTableColumnsWidthProportions(t *testing.T) {
	cols := tableColumns(100)
	total := 0
	for _, c := range cols {
		total += c.Width
	}
	if total < 90 || total > 100 {
		t.Errorf("total column width = %d, expected ~96", total)
	}
}

func TestProfileTableColumns(t *testing.T) {
	cols := profileTableColumns(80)
	if len(cols) != 5 {
		t.Errorf("got %d cols, want 5", len(cols))
	}
	if cols[0].Title != "Name" {
		t.Errorf("first col = %q, want Name", cols[0].Title)
	}
}

func TestAuditTableColumns(t *testing.T) {
	cols := auditTableColumns(80)
	if len(cols) != 5 {
		t.Errorf("got %d cols, want 5", len(cols))
	}
	if cols[0].Title != "Timestamp" {
		t.Errorf("first col = %q, want Timestamp", cols[0].Title)
	}
}

func TestHistoryTableColumns(t *testing.T) {
	cols := historyTableColumns(80)
	if len(cols) != 2 {
		t.Errorf("got %d cols, want 2", len(cols))
	}
	if cols[0].Title != "Version" {
		t.Errorf("first col = %q, want Version", cols[0].Title)
	}
}

func TestTableStyles(t *testing.T) {
	s := tableStyles()
	_ = s.Header
	_ = s.Selected
	_ = s.Cell
}

func TestNewSecretsModel(t *testing.T) {
	m := newSecretsModel("test-profile", 120, 40)
	if m.profileName != "test-profile" {
		t.Errorf("profileName = %q, want %q", m.profileName, "test-profile")
	}
	if m.loaded {
		t.Error("expected loaded = false")
	}
	if m.confirm != "" {
		t.Error("expected empty confirm")
	}
}

func TestNewSecretFormModel(t *testing.T) {
	m := newSecretFormModel("myprofile")
	if m.profileName != "myprofile" {
		t.Errorf("profileName = %q, want %q", m.profileName, "myprofile")
	}
	if len(m.inputs) != formFieldCount {
		t.Errorf("inputs count = %d, want %d", len(m.inputs), formFieldCount)
	}
	if m.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", m.focusIndex)
	}
}

func TestNewProfileFormModel(t *testing.T) {
	m := newProfileFormModel()
	if m.profileName != "__create_profile__" {
		t.Errorf("profileName = %q, want __create_profile__", m.profileName)
	}
	if len(m.inputs) != formFieldCount {
		t.Errorf("inputs count = %d, want %d", len(m.inputs), formFieldCount)
	}
}

func TestSecretFormAccessors(t *testing.T) {
	m := newSecretFormModel("p")
	if m.key() != "" {
		t.Errorf("key() = %q, want empty", m.key())
	}
	if m.value() != "" {
		t.Errorf("value() = %q, want empty", m.value())
	}
	if m.desc() != "" {
		t.Errorf("desc() = %q, want empty", m.desc())
	}
}

func TestProfilesSelectedProfile(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		if got := p.selectedProfile(); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("with profiles", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		p.profiles = []vault.ProfileInfo{
			{Name: "alpha"},
			{Name: "beta"},
		}
		p, _ = p.Update(profilesLoadedMsg{profiles: p.profiles})
		if got := p.selectedProfile(); got != "alpha" {
			t.Errorf("got %q, want %q", got, "alpha")
		}
	})
}

func TestSecretsSelectedKey(t *testing.T) {
	t.Run("no rows", func(t *testing.T) {
		m := newSecretsModel("p", 80, 40)
		if got := m.selectedKey(); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestStatusModelUpdate(t *testing.T) {
	t.Run("loaded success", func(t *testing.T) {
		s := statusModel{}
		status := &vault.Status{Initialized: true, ProfileCount: 2}
		profiles := []vault.ProfileInfo{{Name: "p1"}}

		s, _ = s.Update(statusLoadedMsg{status: status, profiles: profiles})
		if !s.loaded {
			t.Error("expected loaded = true")
		}
		if s.status != status {
			t.Error("status not set")
		}
		if len(s.profiles) != 1 {
			t.Error("profiles not set")
		}
	})

	t.Run("loaded error", func(t *testing.T) {
		s := statusModel{}
		s, _ = s.Update(statusLoadedMsg{err: fmt.Errorf("fail")})
		if !s.loaded {
			t.Error("expected loaded = true even on error")
		}
		if s.status != nil {
			t.Error("status should be nil on error")
		}
	})
}

func TestProfilesModelUpdate(t *testing.T) {
	t.Run("loaded", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		profiles := []vault.ProfileInfo{{Name: "a"}, {Name: "b"}}
		p, _ = p.Update(profilesLoadedMsg{profiles: profiles})
		if !p.loaded {
			t.Error("expected loaded")
		}
		if len(p.profiles) != 2 {
			t.Errorf("profiles count = %d, want 2", len(p.profiles))
		}
	})

	t.Run("action success", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		p.confirm = "delme"
		p, _ = p.Update(profileActionMsg{message: "done"})
		if p.confirm != "" {
			t.Error("confirm should be reset")
		}
		if !strings.Contains(p.message, "done") {
			t.Errorf("message = %q, want contains 'done'", p.message)
		}
	})

	t.Run("action error", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		p.confirm = "x"
		p, _ = p.Update(profileActionMsg{err: fmt.Errorf("boom")})
		if p.confirm != "" {
			t.Error("confirm should be reset")
		}
		if !strings.Contains(p.message, "boom") {
			t.Errorf("message = %q, want contains 'boom'", p.message)
		}
	})
}

func TestSecretsModelUpdate(t *testing.T) {
	now := time.Now()

	t.Run("loaded success", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		secrets := []vault.SecretInfo{
			{Key: "K1", Description: "d1", CreatedAt: now, UpdatedAt: now},
			{Key: "K2", Description: "", CreatedAt: now},
		}
		s, _ = s.Update(secretsLoadedMsg{secrets: secrets})
		if !s.loaded {
			t.Error("expected loaded")
		}
		if len(s.secrets) != 2 {
			t.Errorf("secrets count = %d, want 2", len(s.secrets))
		}
	})

	t.Run("loaded zero UpdatedAt", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		secrets := []vault.SecretInfo{
			{Key: "K1", CreatedAt: now},
		}
		s, _ = s.Update(secretsLoadedMsg{secrets: secrets})
		rows := s.table.Rows()
		if len(rows) != 1 {
			t.Fatalf("rows = %d, want 1", len(rows))
		}
		if rows[0][3] != "-" {
			t.Errorf("updated column = %q, want '-'", rows[0][3])
		}
	})

	t.Run("reveal success", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s, _ = s.Update(secretRevealMsg{key: "K1", value: "secret"})
		if s.revealKey != "K1" || s.revealValue != "secret" {
			t.Errorf("reveal = (%q, %q), want (K1, secret)", s.revealKey, s.revealValue)
		}
	})

	t.Run("reveal error", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s, _ = s.Update(secretRevealMsg{key: "K1", err: fmt.Errorf("fail")})
		if s.revealKey != "" {
			t.Error("revealKey should not be set on error")
		}
		if !strings.Contains(s.message, "fail") {
			t.Errorf("message = %q, want contains 'fail'", s.message)
		}
	})

	t.Run("action success", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s.confirm = "pending"
		s, _ = s.Update(secretActionMsg{message: "deleted"})
		if s.confirm != "" {
			t.Error("confirm should be reset")
		}
		if !strings.Contains(s.message, "deleted") {
			t.Errorf("message = %q, want contains 'deleted'", s.message)
		}
	})
}

func TestSecretFormModelUpdateTab(t *testing.T) {
	f := newSecretFormModel("p")
	f, _ = f.Update(secretFormSubmitMsg{})
	if f.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", f.focusIndex)
	}
}

func TestStatusView(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		s := statusModel{}
		v := s.View(80)
		if !strings.Contains(v, "Loading") {
			t.Errorf("expected Loading, got: %q", v)
		}
	})

	t.Run("nil status", func(t *testing.T) {
		s := statusModel{loaded: true}
		v := s.View(80)
		if !strings.Contains(v, "not initialized") {
			t.Errorf("expected 'not initialized', got: %q", v)
		}
	})

	t.Run("with status", func(t *testing.T) {
		s := statusModel{
			loaded: true,
			status: &vault.Status{
				Initialized:  true,
				DBPath:       "/tmp/vault.db",
				SealMethod:   "software",
				KeyVersion:   1,
				ProfileCount: 2,
				SecretCount:  5,
				PasswordSet:  true,
				CreatedAt:    time.Now(),
			},
			profiles: []vault.ProfileInfo{
				{Name: "default", IsDefault: true, SecretCount: 3},
			},
		}
		v := s.View(80)
		if !strings.Contains(v, "Anvil Vault") {
			t.Error("expected title")
		}
		if !strings.Contains(v, "software") {
			t.Error("expected seal method")
		}
		if !strings.Contains(v, "yes") {
			t.Error("expected password yes")
		}
	})
}

func TestProfilesView(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		v := p.View(80)
		if !strings.Contains(v, "Loading") {
			t.Error("expected Loading")
		}
	})

	t.Run("empty profiles", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		p.loaded = true
		v := p.View(80)
		if !strings.Contains(v, "No profiles") {
			t.Error("expected No profiles")
		}
	})

	t.Run("confirm dialog", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		p.loaded = true
		p.confirm = "devprofile"
		p.profiles = []vault.ProfileInfo{{Name: "devprofile"}}
		v := p.View(80)
		if !strings.Contains(v, "Delete profile") {
			t.Error("expected delete confirmation")
		}
	})

	t.Run("with profiles", func(t *testing.T) {
		p := newProfilesModel(80, 40)
		p.profiles = []vault.ProfileInfo{
			{Name: "prod", IsDefault: true, SecretCount: 5, Description: "production"},
			{Name: "dev", SecretCount: 2},
		}
		p, _ = p.Update(profilesLoadedMsg{profiles: p.profiles})
		v := p.View(80)
		if !strings.Contains(v, "prod") || !strings.Contains(v, "dev") {
			t.Error("expected profile names")
		}
		if !strings.Contains(v, "*") {
			t.Error("expected default marker")
		}
	})
}

func TestSecretsView(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		v := s.View(80)
		if !strings.Contains(v, "Loading") {
			t.Error("expected Loading")
		}
	})

	t.Run("empty", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s.loaded = true
		v := s.View(80)
		if !strings.Contains(v, "No secrets") {
			t.Error("expected No secrets")
		}
	})

	t.Run("confirm dialog", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s.loaded = true
		s.secrets = []vault.SecretInfo{{Key: "K1"}}
		s.confirm = "K1"
		v := s.View(80)
		if !strings.Contains(v, "Delete secret") {
			t.Error("expected delete confirmation")
		}
	})

	t.Run("with reveal", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s.loaded = true
		now := time.Now()
		s.secrets = []vault.SecretInfo{{Key: "K1", CreatedAt: now}}
		s, _ = s.Update(secretsLoadedMsg{secrets: s.secrets})
		s.revealKey = "K1"
		s.revealValue = "mysecret"
		v := s.View(80)
		if !strings.Contains(v, "mysecret") {
			t.Error("expected revealed value")
		}
	})

	t.Run("with message", func(t *testing.T) {
		s := newSecretsModel("p", 80, 40)
		s.loaded = true
		now := time.Now()
		s.secrets = []vault.SecretInfo{{Key: "K1", CreatedAt: now}}
		s, _ = s.Update(secretsLoadedMsg{secrets: s.secrets})
		s.message = "Something happened"
		v := s.View(80)
		if !strings.Contains(v, "Something happened") {
			t.Error("expected message")
		}
	})
}

func TestSecretFormView(t *testing.T) {
	f := newSecretFormModel("myprofile")
	v := f.View(80)
	if !strings.Contains(v, "myprofile") {
		t.Error("expected profile name in title")
	}
	if !strings.Contains(v, "Key") || !strings.Contains(v, "Value") || !strings.Contains(v, "Description") {
		t.Error("expected field labels")
	}
	if !strings.Contains(v, "tab:") {
		t.Error("expected help text")
	}
}

func TestSecretFormViewWithMessage(t *testing.T) {
	f := newSecretFormModel("p")
	f.message = "Error: something"
	v := f.View(80)
	if !strings.Contains(v, "Error: something") {
		t.Error("expected error message in view")
	}
}

func TestModelView(t *testing.T) {
	m := NewModel(nil)
	m.width = 80
	m.height = 40

	// Default screen is status, should render without panic
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view")
	}
}

// --- Key helpers ---

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func specialKeyMsg(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// --- WindowSizeMsg ---

func TestModelUpdateWindowSize(t *testing.T) {
	m := NewModel(nil)
	m.width = 80
	m.height = 40
	m.currentScreen = screenSecrets
	m.secrets = newSecretsModel("p", 80, 40)

	result, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	rm := result.(model)
	if rm.width != 120 || rm.height != 50 {
		t.Errorf("size = (%d, %d), want (120, 50)", rm.width, rm.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd from WindowSizeMsg")
	}
}

func TestModelUpdateWindowSizeProfiles(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenProfiles
	m.profiles = newProfilesModel(80, 40)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	rm := result.(model)
	if rm.width != 100 {
		t.Errorf("width = %d, want 100", rm.width)
	}
}

func TestModelUpdateWindowSizeAudit(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenAudit
	m.audit = newAuditModel(80, 40)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	rm := result.(model)
	if rm.width != 100 {
		t.Errorf("width = %d, want 100", rm.width)
	}
}

func TestModelUpdateWindowSizeHistory(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenHistory
	m.history = newHistoryModel("key", "p", 80, 40)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	rm := result.(model)
	if rm.width != 100 {
		t.Errorf("width = %d, want 100", rm.width)
	}
}

func TestModelUpdateWindowSizeStatus(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenStatus
	result, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	rm := result.(model)
	if rm.width != 100 {
		t.Errorf("width = %d, want 100", rm.width)
	}
}

// --- Ctrl+C global quit ---

func TestModelUpdateCtrlC(t *testing.T) {
	m := NewModel(nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit cmd from ctrl+c")
	}
}

// --- Sidebar navigation ---

func TestSidebarNavigation(t *testing.T) {
	m := NewModel(nil)
	m.width = 100
	m.height = 40

	// Default focus is sidebar
	if m.focus != focusSidebar {
		t.Errorf("initial focus = %d, want sidebar", m.focus)
	}

	// Reset cursor to 0 for predictable testing
	m.sidebar.cursor = 0

	// Move down
	result, _ := m.Update(keyMsg("j"))
	rm := result.(model)
	if rm.sidebar.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", rm.sidebar.cursor)
	}

	// Move up
	result, _ = rm.Update(keyMsg("k"))
	rm = result.(model)
	if rm.sidebar.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after k", rm.sidebar.cursor)
	}
}

func TestSidebarQuit(t *testing.T) {
	m := NewModel(nil)
	m.focus = focusSidebar
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd from q in sidebar")
	}
}

func TestTabTogglesFocus(t *testing.T) {
	m := NewModel(nil)
	m.width = 100
	m.height = 40
	m.focus = focusSidebar

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	rm := result.(model)
	if rm.focus != focusContent {
		t.Error("tab should switch to content focus")
	}

	result, _ = rm.Update(tea.KeyMsg{Type: tea.KeyTab})
	rm = result.(model)
	if rm.focus != focusSidebar {
		t.Error("tab should switch back to sidebar focus")
	}
}

func TestEscFromContentReturnsSidebar(t *testing.T) {
	m := NewModel(nil)
	m.width = 100
	m.height = 40
	m.currentScreen = screenStatus
	m.focus = focusContent

	result, _ := m.Update(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.focus != focusSidebar {
		t.Error("esc from content should return to sidebar")
	}
}

// --- updateProfiles key tests ---

func testProfilesModel() model {
	m := NewModel(nil)
	m.width = 80
	m.height = 40
	m.currentScreen = screenProfiles
	m.focus = focusContent
	m.profiles = newProfilesModel(80, 40)
	profiles := []vault.ProfileInfo{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "gamma"},
	}
	m.profiles, _ = m.profiles.Update(profilesLoadedMsg{profiles: profiles})
	return m
}

func TestUpdateProfilesQuit(t *testing.T) {
	m := testProfilesModel()
	_, cmd := m.updateProfiles(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestUpdateProfilesEnter(t *testing.T) {
	m := testProfilesModel()
	result, _ := m.updateProfiles(specialKeyMsg(tea.KeyEnter))
	rm := result.(model)
	if rm.currentScreen != screenSecrets {
		t.Errorf("screen = %d, want secrets", rm.currentScreen)
	}
	if rm.secrets.profileName != "alpha" {
		t.Errorf("profileName = %q, want alpha", rm.secrets.profileName)
	}
}

func TestUpdateProfilesEnterBlockedDuringConfirm(t *testing.T) {
	m := testProfilesModel()
	m.profiles.confirm = "alpha"
	result, _ := m.updateProfiles(specialKeyMsg(tea.KeyEnter))
	rm := result.(model)
	if rm.currentScreen != screenProfiles {
		t.Error("should stay on profiles during confirm")
	}
}

func TestUpdateProfilesCreate(t *testing.T) {
	m := testProfilesModel()
	result, cmd := m.updateProfiles(keyMsg("c"))
	rm := result.(model)
	if rm.currentScreen != screenSecretForm {
		t.Errorf("screen = %d, want secret form", rm.currentScreen)
	}
	if rm.secretForm.profileName != "__create_profile__" {
		t.Error("expected profile form")
	}
	if cmd == nil {
		t.Error("expected focus cmd")
	}
}

func TestUpdateProfilesNewCancelConfirm(t *testing.T) {
	m := testProfilesModel()
	m.profiles.confirm = "alpha"
	result, _ := m.updateProfiles(keyMsg("n"))
	rm := result.(model)
	if rm.profiles.confirm != "" {
		t.Error("n should cancel confirm")
	}
}

func TestUpdateProfilesNewNoConfirm(t *testing.T) {
	m := testProfilesModel()
	result, cmd := m.updateProfiles(keyMsg("n"))
	rm := result.(model)
	if rm.currentScreen != screenSecretForm {
		t.Error("n without confirm should go to form")
	}
	if cmd == nil {
		t.Error("expected focus cmd")
	}
}

func TestUpdateProfilesDelete(t *testing.T) {
	m := testProfilesModel()
	result, _ := m.updateProfiles(keyMsg("d"))
	rm := result.(model)
	if rm.profiles.confirm != "alpha" {
		t.Errorf("confirm = %q, want alpha", rm.profiles.confirm)
	}
}

func TestUpdateProfilesConfirmDelete(t *testing.T) {
	m := testProfilesModel()
	m.profiles.confirm = "beta"
	_, cmd := m.updateProfiles(keyMsg("y"))
	if cmd == nil {
		t.Error("expected delete cmd")
	}
}

func TestUpdateProfilesUseProfile(t *testing.T) {
	m := testProfilesModel()
	_, cmd := m.updateProfiles(keyMsg("u"))
	if cmd == nil {
		t.Error("expected use cmd")
	}
}

func TestUpdateProfilesDefaultKey(t *testing.T) {
	m := testProfilesModel()
	m.profiles.message = "old message"
	result, _ := m.updateProfiles(keyMsg("x"))
	rm := result.(model)
	if rm.profiles.message != "" {
		t.Error("default key should clear message")
	}
}

// --- updateSecrets key tests ---

func testSecretsModel() model {
	m := NewModel(nil)
	m.currentScreen = screenSecrets
	m.focus = focusContent
	m.width = 80
	m.height = 40
	m.secrets = newSecretsModel("myprofile", 80, 40)
	m.secrets.loaded = true
	now := time.Now()
	m.secrets.secrets = []vault.SecretInfo{
		{Key: "K1", CreatedAt: now},
		{Key: "K2", CreatedAt: now},
	}
	m.secrets, _ = m.secrets.Update(secretsLoadedMsg{secrets: m.secrets.secrets})
	return m
}

func TestUpdateSecretsQuit(t *testing.T) {
	m := testSecretsModel()
	_, cmd := m.updateSecrets(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestUpdateSecretsEnterReveal(t *testing.T) {
	m := testSecretsModel()
	_, cmd := m.updateSecrets(specialKeyMsg(tea.KeyEnter))
	if cmd == nil {
		t.Error("expected reveal cmd")
	}
}

func TestUpdateSecretsEnterToggleReveal(t *testing.T) {
	m := testSecretsModel()
	m.secrets.revealKey = "K1"
	m.secrets.revealValue = "secret"
	result, _ := m.updateSecrets(specialKeyMsg(tea.KeyEnter))
	rm := result.(model)
	if rm.secrets.revealKey != "" {
		t.Error("expected reveal cleared on toggle")
	}
}

func TestUpdateSecretsEnterBlockedDuringConfirm(t *testing.T) {
	m := testSecretsModel()
	m.secrets.confirm = "K1"
	_, cmd := m.updateSecrets(specialKeyMsg(tea.KeyEnter))
	if cmd != nil {
		t.Error("enter should be blocked during confirm")
	}
}

func TestUpdateSecretsHistory(t *testing.T) {
	m := testSecretsModel()
	result, _ := m.updateSecrets(keyMsg("h"))
	rm := result.(model)
	if rm.currentScreen != screenHistory {
		t.Errorf("screen = %d, want history", rm.currentScreen)
	}
	if rm.history.key != "K1" {
		t.Errorf("history key = %q, want K1", rm.history.key)
	}
}

func TestUpdateSecretsCreate(t *testing.T) {
	m := testSecretsModel()
	result, cmd := m.updateSecrets(keyMsg("c"))
	rm := result.(model)
	if rm.currentScreen != screenSecretForm {
		t.Error("expected form screen")
	}
	if rm.secretForm.profileName != "myprofile" {
		t.Errorf("profileName = %q", rm.secretForm.profileName)
	}
	if cmd == nil {
		t.Error("expected focus cmd")
	}
}

func TestUpdateSecretsNewCancelConfirm(t *testing.T) {
	m := testSecretsModel()
	m.secrets.confirm = "K1"
	result, _ := m.updateSecrets(keyMsg("n"))
	rm := result.(model)
	if rm.secrets.confirm != "" {
		t.Error("n should cancel confirm")
	}
}

func TestUpdateSecretsNewNoConfirm(t *testing.T) {
	m := testSecretsModel()
	result, cmd := m.updateSecrets(keyMsg("n"))
	rm := result.(model)
	if rm.currentScreen != screenSecretForm {
		t.Error("expected form screen")
	}
	if cmd == nil {
		t.Error("expected focus cmd")
	}
}

func TestUpdateSecretsDeleteRequest(t *testing.T) {
	m := testSecretsModel()
	result, _ := m.updateSecrets(keyMsg("d"))
	rm := result.(model)
	if rm.secrets.confirm == "" {
		t.Error("expected confirm to be set")
	}
}

func TestUpdateSecretsConfirmDelete(t *testing.T) {
	m := testSecretsModel()
	m.secrets.confirm = "K1"
	_, cmd := m.updateSecrets(keyMsg("y"))
	if cmd == nil {
		t.Error("expected delete cmd")
	}
}

func TestUpdateSecretsDefaultKey(t *testing.T) {
	m := testSecretsModel()
	result, _ := m.updateSecrets(keyMsg("x"))
	rm := result.(model)
	if rm.secrets.revealKey != "" {
		t.Error("default key should clear reveal")
	}
}

func TestUpdateSecretsDefaultKeyBlockedDuringConfirm(t *testing.T) {
	m := testSecretsModel()
	m.secrets.confirm = "K1"
	m.secrets.revealKey = "K2"
	result, _ := m.updateSecrets(keyMsg("x"))
	rm := result.(model)
	if rm.secrets.revealKey != "K2" {
		t.Error("reveal should not be cleared during confirm")
	}
}

// --- updateAuditScreen key tests ---

func TestUpdateAuditQuit(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenAudit
	m.audit = newAuditModel(80, 40)
	_, cmd := m.updateAuditScreen(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

// --- updateHistoryScreen key tests ---

func TestUpdateHistoryEsc(t *testing.T) {
	m := NewModel(nil)
	m.width = 100
	m.height = 40
	m.currentScreen = screenHistory
	m.focus = focusContent
	m.history = newHistoryModel("key", "p", 80, 40)
	// Esc from history returns to secrets
	result, _ := m.Update(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.currentScreen != screenSecrets {
		t.Errorf("screen = %d, want secrets", rm.currentScreen)
	}
}

func TestUpdateHistoryQuit(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenHistory
	m.history = newHistoryModel("key", "p", 80, 40)
	_, cmd := m.updateHistoryScreen(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

// --- updateSecretForm key tests ---

func TestUpdateSecretFormEscFromProfile(t *testing.T) {
	m := NewModel(nil)
	m.width = 100
	m.height = 40
	m.currentScreen = screenSecretForm
	m.focus = focusContent
	m.secretForm = newProfileFormModel()
	result, _ := m.updateSecretForm(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.currentScreen != screenProfiles {
		t.Errorf("screen = %d, want profiles", rm.currentScreen)
	}
}

func TestUpdateSecretFormEscFromSecret(t *testing.T) {
	m := NewModel(nil)
	m.width = 100
	m.height = 40
	m.currentScreen = screenSecretForm
	m.focus = focusContent
	m.secretForm = newSecretFormModel("myprofile")
	result, _ := m.updateSecretForm(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.currentScreen != screenSecrets {
		t.Errorf("screen = %d, want secrets", rm.currentScreen)
	}
}

func TestUpdateSecretFormEnterCreateProfileEmpty(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecretForm
	m.secretForm = newProfileFormModel()
	result, cmd := m.updateSecretForm(specialKeyMsg(tea.KeyEnter))
	rm := result.(model)
	if !strings.Contains(rm.secretForm.message, "required") {
		t.Errorf("message = %q, want contains 'required'", rm.secretForm.message)
	}
	if cmd != nil {
		t.Error("expected nil cmd for validation error")
	}
}

func TestUpdateSecretFormEnterCreateSecretEmpty(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecretForm
	m.secretForm = newSecretFormModel("myprofile")
	result, cmd := m.updateSecretForm(specialKeyMsg(tea.KeyEnter))
	rm := result.(model)
	if !strings.Contains(rm.secretForm.message, "required") {
		t.Errorf("message = %q, want contains 'required'", rm.secretForm.message)
	}
	if cmd != nil {
		t.Error("expected nil cmd for validation error")
	}
}

func TestUpdateSecretFormDelegateToForm(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecretForm
	m.focus = focusContent
	m.secretForm = newSecretFormModel("p")
	// Tab within secret form delegates to form's tab handler (not sidebar toggle)
	result, _ := m.Update(specialKeyMsg(tea.KeyTab))
	rm := result.(model)
	if rm.secretForm.focusIndex != 1 {
		t.Errorf("focusIndex = %d, want 1 after tab", rm.secretForm.focusIndex)
	}
}

// --- secretFormModel tab navigation ---

func TestSecretFormTabForward(t *testing.T) {
	f := newSecretFormModel("p")
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if f.focusIndex != 1 {
		t.Errorf("focusIndex = %d, want 1", f.focusIndex)
	}
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if f.focusIndex != 2 {
		t.Errorf("focusIndex = %d, want 2", f.focusIndex)
	}
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if f.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0 (wrap)", f.focusIndex)
	}
}

func TestSecretFormShiftTab(t *testing.T) {
	f := newSecretFormModel("p")
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if f.focusIndex != 2 {
		t.Errorf("focusIndex = %d, want 2 (wrap back)", f.focusIndex)
	}
}

func TestSecretFormSubmitError(t *testing.T) {
	f := newSecretFormModel("p")
	f, _ = f.Update(secretFormSubmitMsg{err: fmt.Errorf("oops")})
	if !strings.Contains(f.message, "oops") {
		t.Errorf("message = %q, want contains 'oops'", f.message)
	}
	if f.submitted {
		t.Error("should not be submitted on error")
	}
}

func TestSecretFormSubmitSuccess(t *testing.T) {
	f := newSecretFormModel("p")
	f, _ = f.Update(secretFormSubmitMsg{})
	if !f.submitted {
		t.Error("expected submitted = true")
	}
}

// --- Model.Update non-key message delegation ---

func TestModelUpdateProfileActionMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenProfiles
	m.profiles.confirm = "test"

	result, _ := m.Update(profileActionMsg{message: "done"})
	rm := result.(model)
	if rm.profiles.confirm != "" {
		t.Error("confirm should be reset")
	}
}

func TestModelUpdateSecretActionMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecrets
	m.secrets = newSecretsModel("p", 80, 40)
	m.secrets.confirm = "K1"

	result, _ := m.Update(secretActionMsg{message: "deleted"})
	rm := result.(model)
	if rm.secrets.confirm != "" {
		t.Error("confirm should be reset")
	}
}

func TestModelUpdateDelegatesStatusLoadedMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenStatus
	status := &vault.Status{Initialized: true}

	result, _ := m.Update(statusLoadedMsg{status: status})
	rm := result.(model)
	if !rm.status.loaded {
		t.Error("status should be loaded")
	}
}

func TestModelUpdateDelegatesSecretsLoadedMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecrets
	m.secrets = newSecretsModel("p", 80, 40)

	result, _ := m.Update(secretsLoadedMsg{secrets: []vault.SecretInfo{{Key: "K1", CreatedAt: time.Now()}}})
	rm := result.(model)
	if !rm.secrets.loaded {
		t.Error("secrets should be loaded")
	}
}

func TestModelUpdateDelegatesProfilesLoadedMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenProfiles

	result, _ := m.Update(profilesLoadedMsg{profiles: []vault.ProfileInfo{{Name: "p1"}}})
	rm := result.(model)
	if !rm.profiles.loaded {
		t.Error("profiles should be loaded")
	}
}

func TestModelUpdateDelegatesAuditLoadedMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenAudit
	m.audit = newAuditModel(80, 40)

	result, _ := m.Update(auditLoadedMsg{entries: []vault.AuditEntry{{Action: "set"}}})
	rm := result.(model)
	if !rm.audit.loaded {
		t.Error("audit should be loaded")
	}
}

func TestModelUpdateDelegatesHistoryLoadedMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenHistory
	m.history = newHistoryModel("key", "p", 80, 40)

	result, _ := m.Update(historyLoadedMsg{versions: []vault.SecretVersion{{Version: 1, CreatedAt: time.Now()}}})
	rm := result.(model)
	if !rm.history.loaded {
		t.Error("history should be loaded")
	}
}

// --- Model.View for all screens ---

func TestModelViewAllScreens(t *testing.T) {
	screens := []struct {
		name   string
		screen screen
	}{
		{"status", screenStatus},
		{"profiles", screenProfiles},
		{"secrets", screenSecrets},
		{"secretForm", screenSecretForm},
		{"audit", screenAudit},
		{"history", screenHistory},
		{"password", screenPassword},
		{"keyRotation", screenKeyRotation},
		{"seal", screenSeal},
		{"envRelease", screenEnvRelease},
		{"envStatus", screenEnvStatus},
		{"envExport", screenEnvExport},
		{"templates", screenTemplates},
		{"templateApply", screenTemplateApply},
		{"backup", screenBackup},
		{"share", screenShare},
		{"docker", screenDocker},
		{"plugins", screenPlugins},
		{"gather", screenGather},
		{"importExport", screenImportExport},
	}
	for _, tt := range screens {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(nil)
			m.width = 80
			m.height = 40
			m.currentScreen = tt.screen
			switch tt.screen {
			case screenSecrets:
				m.secrets = newSecretsModel("p", 80, 40)
			case screenSecretForm:
				m.secretForm = newSecretFormModel("p")
			case screenAudit:
				m.audit = newAuditModel(80, 40)
			case screenHistory:
				m.history = newHistoryModel("key", "p", 80, 40)
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
				m.templates = newTemplatesModel(80, 40)
			case screenTemplateApply:
				m.templateApply = newTemplateApplyModel()
			case screenBackup:
				m.backup = newBackupScreenModel()
			case screenShare:
				m.share = newShareScreenModel()
			case screenDocker:
				m.docker = newDockerScreenModel()
			case screenPlugins:
				m.plugins = newPluginsModel(80, 40)
			case screenGather:
				m.gather = newGatherScreenModel()
			case screenImportExport:
				m.importExport = newImportExportModel()
			}
			v := m.View()
			if v == "" {
				t.Error("expected non-empty view")
			}
		})
	}
}

func TestModelViewUnknownScreen(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screen(99)
	v := m.View()
	// View includes sidebar so won't be fully empty
	if v == "" {
		t.Error("expected non-empty view (sidebar)")
	}
}

// --- Profiles view with message ---

func TestProfilesViewWithMessage(t *testing.T) {
	p := newProfilesModel(80, 40)
	p.loaded = true
	p.profiles = []vault.ProfileInfo{{Name: "test"}}
	p, _ = p.Update(profilesLoadedMsg{profiles: p.profiles})
	p.message = "Profile deleted."
	v := p.View(80)
	if !strings.Contains(v, "Profile deleted.") {
		t.Error("expected message in view")
	}
}

// --- Audit view tests ---

func TestAuditView(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		a := newAuditModel(80, 40)
		v := a.View(80)
		if !strings.Contains(v, "Loading") {
			t.Error("expected Loading")
		}
	})

	t.Run("empty", func(t *testing.T) {
		a := newAuditModel(80, 40)
		a.loaded = true
		v := a.View(80)
		if !strings.Contains(v, "No audit") {
			t.Error("expected No audit")
		}
	})

	t.Run("with entries", func(t *testing.T) {
		a := newAuditModel(80, 40)
		entries := []vault.AuditEntry{
			{Action: "set", ProfileName: "default", SecretKey: "API_KEY", CreatedAt: time.Now()},
		}
		a, _ = a.Update(auditLoadedMsg{entries: entries})
		v := a.View(80)
		if !strings.Contains(v, "set") {
			t.Error("expected action in view")
		}
	})
}

// --- History view tests ---

func TestHistoryView(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		h := newHistoryModel("key", "p", 80, 40)
		v := h.View(80)
		if !strings.Contains(v, "Loading") {
			t.Error("expected Loading")
		}
	})

	t.Run("empty", func(t *testing.T) {
		h := newHistoryModel("key", "p", 80, 40)
		h.loaded = true
		v := h.View(80)
		if !strings.Contains(v, "No version") {
			t.Error("expected No version")
		}
	})

	t.Run("with versions", func(t *testing.T) {
		h := newHistoryModel("key", "p", 80, 40)
		versions := []vault.SecretVersion{
			{Version: 1, CreatedAt: time.Now()},
		}
		h, _ = h.Update(historyLoadedMsg{versions: versions})
		v := h.View(80)
		if !strings.Contains(v, "v1") {
			t.Error("expected version in view")
		}
	})
}

// --- Error loading tests ---

func TestSecretsModelUpdateLoadError(t *testing.T) {
	s := newSecretsModel("p", 80, 40)
	s, _ = s.Update(secretsLoadedMsg{err: fmt.Errorf("load failed")})
	if !s.loaded {
		t.Error("expected loaded even on error")
	}
	if len(s.secrets) != 0 {
		t.Error("secrets should be empty on error")
	}
}

func TestSecretsModelUpdateActionError(t *testing.T) {
	s := newSecretsModel("p", 80, 40)
	s.confirm = "K1"
	s, _ = s.Update(secretActionMsg{err: fmt.Errorf("delete failed")})
	if s.confirm != "" {
		t.Error("confirm should be reset")
	}
	if !strings.Contains(s.message, "delete failed") {
		t.Errorf("message = %q, want contains 'delete failed'", s.message)
	}
}

func TestProfilesModelUpdateLoadError(t *testing.T) {
	p := newProfilesModel(80, 40)
	p, _ = p.Update(profilesLoadedMsg{err: fmt.Errorf("load failed")})
	if !p.loaded {
		t.Error("expected loaded even on error")
	}
	if len(p.profiles) != 0 {
		t.Error("profiles should be empty on error")
	}
}

func TestAuditModelUpdateLoadError(t *testing.T) {
	a := newAuditModel(80, 40)
	a, _ = a.Update(auditLoadedMsg{err: fmt.Errorf("audit failed")})
	if !a.loaded {
		t.Error("expected loaded even on error")
	}
	if len(a.entries) != 0 {
		t.Error("entries should be empty on error")
	}
}

func TestHistoryModelUpdateLoadError(t *testing.T) {
	h := newHistoryModel("key", "p", 80, 40)
	h, _ = h.Update(historyLoadedMsg{err: fmt.Errorf("history failed")})
	if !h.loaded {
		t.Error("expected loaded even on error")
	}
	if len(h.versions) != 0 {
		t.Error("versions should be empty on error")
	}
}

// --- Sidebar model tests ---

func TestNewSidebarModel(t *testing.T) {
	s := newSidebarModel()
	if len(s.categories) == 0 {
		t.Error("expected categories")
	}
	if s.totalItems() == 0 {
		t.Error("expected items")
	}
}

func TestSidebarMoveUpDown(t *testing.T) {
	s := newSidebarModel()
	s.moveDown()
	if s.cursor != 1 {
		t.Errorf("cursor = %d, want 1", s.cursor)
	}
	s.moveUp()
	if s.cursor != 0 {
		t.Errorf("cursor = %d, want 0", s.cursor)
	}
	s.moveUp() // shouldn't go negative
	if s.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (clamped)", s.cursor)
	}
}

func TestSidebarSelectedScreen(t *testing.T) {
	s := newSidebarModel()
	// First item is Profiles
	if s.selectedScreen() != screenProfiles {
		t.Errorf("selectedScreen = %d, want screenProfiles", s.selectedScreen())
	}
}

func TestSidebarMoveCursorToScreen(t *testing.T) {
	s := newSidebarModel()
	s.moveCursorToScreen(screenAudit)
	if s.selectedScreen() != screenAudit {
		t.Errorf("selectedScreen = %d, want screenAudit", s.selectedScreen())
	}
}

func TestSidebarView(t *testing.T) {
	s := newSidebarModel()
	v := s.View(30)
	if !strings.Contains(v, "Secrets") {
		t.Error("expected category label in sidebar view")
	}
}

// --- New screen view tests ---

func TestPasswordScreenView(t *testing.T) {
	p := newPasswordScreenModel()
	v := p.View(80)
	if !strings.Contains(v, "Password") {
		t.Error("expected title")
	}
}

func TestKeyRotationView(t *testing.T) {
	k := newKeyRotationModel()
	v := k.View(80)
	if !strings.Contains(v, "Key Rotation") {
		t.Error("expected title")
	}
}

func TestSealScreenView(t *testing.T) {
	s := newSealScreenModel()
	v := s.View(80)
	if !strings.Contains(v, "Seal") {
		t.Error("expected title")
	}
}

func TestEnvReleaseView(t *testing.T) {
	e := newEnvReleaseModel()
	v := e.View(80)
	if !strings.Contains(v, "Env Release") {
		t.Error("expected title")
	}
}

func TestEnvStatusView(t *testing.T) {
	e := newEnvStatusModel()
	v := e.View(80)
	if !strings.Contains(v, "Loading") {
		t.Error("expected loading")
	}
}

func TestEnvExportView(t *testing.T) {
	e := newEnvExportModel()
	v := e.View(80)
	if !strings.Contains(v, "Env Export") {
		t.Error("expected title")
	}
}

func TestTemplatesView(t *testing.T) {
	tm := newTemplatesModel(80, 40)
	v := tm.View(80)
	if !strings.Contains(v, "Loading") {
		t.Error("expected loading")
	}
}

func TestTemplateApplyView(t *testing.T) {
	ta := newTemplateApplyModel()
	v := ta.View(80)
	if !strings.Contains(v, "Loading") {
		t.Error("expected loading")
	}
}

func TestBackupScreenView(t *testing.T) {
	b := newBackupScreenModel()
	v := b.View(80)
	if !strings.Contains(v, "Backup") {
		t.Error("expected title")
	}
}

func TestShareScreenView(t *testing.T) {
	s := newShareScreenModel()
	v := s.View(80)
	if !strings.Contains(v, "Share") {
		t.Error("expected title")
	}
}

func TestDockerScreenView(t *testing.T) {
	d := newDockerScreenModel()
	v := d.View(80)
	if !strings.Contains(v, "Docker") {
		t.Error("expected title")
	}
}

func TestPluginsView(t *testing.T) {
	p := newPluginsModel(80, 40)
	v := p.View(80)
	if !strings.Contains(v, "Loading") {
		t.Error("expected loading")
	}
}

func TestGatherScreenView(t *testing.T) {
	g := newGatherScreenModel()
	v := g.View(80)
	if !strings.Contains(v, "Gather") {
		t.Error("expected title")
	}
}

func TestImportExportView(t *testing.T) {
	ie := newImportExportModel()
	v := ie.View(80)
	if !strings.Contains(v, "Import") {
		t.Error("expected title")
	}
}
