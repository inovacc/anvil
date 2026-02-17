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
		// Verify column titles
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
	// Should use roughly 96 (100-4 padding)
	if total < 90 || total > 100 {
		t.Errorf("total column width = %d, expected ~96", total)
	}
}

func TestTableStyles(t *testing.T) {
	s := tableStyles()
	// Just verify it doesn't panic and returns something
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
	// Initially empty
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
		p := profilesModel{}
		if got := p.selectedProfile(); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("valid cursor", func(t *testing.T) {
		p := profilesModel{
			profiles: []vault.ProfileInfo{
				{Name: "alpha"},
				{Name: "beta"},
			},
			cursor: 1,
		}
		if got := p.selectedProfile(); got != "beta" {
			t.Errorf("got %q, want %q", got, "beta")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		p := profilesModel{
			profiles: []vault.ProfileInfo{{Name: "only"}},
			cursor:   5,
		}
		if got := p.selectedProfile(); got != "" {
			t.Errorf("got %q, want empty", got)
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

func TestDashboardModelUpdate(t *testing.T) {
	t.Run("loaded success", func(t *testing.T) {
		d := dashboardModel{}
		status := &vault.Status{Initialized: true, ProfileCount: 2}
		profiles := []vault.ProfileInfo{{Name: "p1"}}

		d, _ = d.Update(dashboardLoadedMsg{status: status, profiles: profiles})
		if !d.loaded {
			t.Error("expected loaded = true")
		}
		if d.status != status {
			t.Error("status not set")
		}
		if len(d.profiles) != 1 {
			t.Error("profiles not set")
		}
	})

	t.Run("loaded error", func(t *testing.T) {
		d := dashboardModel{}
		d, _ = d.Update(dashboardLoadedMsg{err: fmt.Errorf("fail")})
		if !d.loaded {
			t.Error("expected loaded = true even on error")
		}
		if d.status != nil {
			t.Error("status should be nil on error")
		}
	})
}

func TestProfilesModelUpdate(t *testing.T) {
	t.Run("loaded", func(t *testing.T) {
		p := profilesModel{cursor: 5}
		profiles := []vault.ProfileInfo{{Name: "a"}, {Name: "b"}}
		p, _ = p.Update(profilesLoadedMsg{profiles: profiles})
		if !p.loaded {
			t.Error("expected loaded")
		}
		// cursor should be clamped
		if p.cursor != 1 {
			t.Errorf("cursor = %d, want 1 (clamped)", p.cursor)
		}
	})

	t.Run("action success", func(t *testing.T) {
		p := profilesModel{confirm: "delme"}
		p, _ = p.Update(profileActionMsg{message: "done"})
		if p.confirm != "" {
			t.Error("confirm should be reset")
		}
		if !strings.Contains(p.message, "done") {
			t.Errorf("message = %q, want contains 'done'", p.message)
		}
	})

	t.Run("action error", func(t *testing.T) {
		p := profilesModel{confirm: "x"}
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
		// The zero UpdatedAt should produce "-"
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

	// Tab forward
	f, _ = f.Update(secretFormSubmitMsg{}) // unrelated msg, no change
	if f.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0", f.focusIndex)
	}
}

func TestDashboardView(t *testing.T) {
	t.Run("not loaded", func(t *testing.T) {
		d := dashboardModel{}
		v := d.View(80)
		if !strings.Contains(v, "Loading") {
			t.Errorf("expected Loading, got: %q", v)
		}
	})

	t.Run("nil status", func(t *testing.T) {
		d := dashboardModel{loaded: true}
		v := d.View(80)
		if !strings.Contains(v, "not initialized") {
			t.Errorf("expected 'not initialized', got: %q", v)
		}
	})

	t.Run("with status", func(t *testing.T) {
		d := dashboardModel{
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
		v := d.View(80)
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
		p := profilesModel{}
		v := p.View(80)
		if !strings.Contains(v, "Loading") {
			t.Error("expected Loading")
		}
	})

	t.Run("empty profiles", func(t *testing.T) {
		p := profilesModel{loaded: true}
		v := p.View(80)
		if !strings.Contains(v, "No profiles") {
			t.Error("expected No profiles")
		}
	})

	t.Run("confirm dialog", func(t *testing.T) {
		p := profilesModel{
			loaded:   true,
			confirm:  "devprofile",
			profiles: []vault.ProfileInfo{{Name: "devprofile"}},
		}
		v := p.View(80)
		if !strings.Contains(v, "Delete profile") {
			t.Error("expected delete confirmation")
		}
	})

	t.Run("with profiles", func(t *testing.T) {
		p := profilesModel{
			loaded: true,
			profiles: []vault.ProfileInfo{
				{Name: "prod", IsDefault: true, SecretCount: 5, Description: "production"},
				{Name: "dev", SecretCount: 2},
			},
			cursor: 0,
		}
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

func TestDashboardInit(t *testing.T) {
	d := dashboardModel{}
	cmd := d.Init()
	if cmd != nil {
		t.Error("expected nil cmd from Init()")
	}
}

func TestModelView(t *testing.T) {
	m := NewModel(nil)
	m.width = 80
	m.height = 40

	// Default screen is dashboard, should render without panic
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

func TestModelUpdateWindowSizeNonSecrets(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenDashboard
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

// --- updateDashboard key tests ---

func TestUpdateDashboardQuit(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenDashboard
	_, cmd := m.updateDashboard(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd from q")
	}
}

func TestUpdateDashboardNavigateToProfiles(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenDashboard
	result, _ := m.updateDashboard(keyMsg("p"))
	rm := result.(model)
	if rm.currentScreen != screenProfiles {
		t.Errorf("screen = %d, want %d", rm.currentScreen, screenProfiles)
	}
}

func TestUpdateDashboardUnhandledKey(t *testing.T) {
	m := NewModel(nil)
	_, cmd := m.updateDashboard(keyMsg("x"))
	if cmd != nil {
		t.Error("expected nil cmd for unhandled key")
	}
}

// --- updateProfiles key tests ---

func testProfilesModel() model {
	m := NewModel(nil)
	m.currentScreen = screenProfiles
	m.profiles = profilesModel{
		loaded: true,
		profiles: []vault.ProfileInfo{
			{Name: "alpha"},
			{Name: "beta"},
			{Name: "gamma"},
		},
		cursor: 1,
	}
	return m
}

func TestUpdateProfilesEsc(t *testing.T) {
	m := testProfilesModel()
	result, _ := m.updateProfiles(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.currentScreen != screenDashboard {
		t.Errorf("screen = %d, want dashboard", rm.currentScreen)
	}
	if rm.profiles.message != "" {
		t.Error("message should be cleared")
	}
}

func TestUpdateProfilesQuit(t *testing.T) {
	m := testProfilesModel()
	_, cmd := m.updateProfiles(keyMsg("q"))
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestUpdateProfilesNavigateUp(t *testing.T) {
	m := testProfilesModel()
	m.profiles.cursor = 2
	result, _ := m.updateProfiles(keyMsg("k"))
	rm := result.(model)
	if rm.profiles.cursor != 1 {
		t.Errorf("cursor = %d, want 1", rm.profiles.cursor)
	}
}

func TestUpdateProfilesNavigateUpAtTop(t *testing.T) {
	m := testProfilesModel()
	m.profiles.cursor = 0
	result, _ := m.updateProfiles(specialKeyMsg(tea.KeyUp))
	rm := result.(model)
	if rm.profiles.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (clamped)", rm.profiles.cursor)
	}
}

func TestUpdateProfilesNavigateDown(t *testing.T) {
	m := testProfilesModel()
	m.profiles.cursor = 0
	result, _ := m.updateProfiles(keyMsg("j"))
	rm := result.(model)
	if rm.profiles.cursor != 1 {
		t.Errorf("cursor = %d, want 1", rm.profiles.cursor)
	}
}

func TestUpdateProfilesNavigateDownAtBottom(t *testing.T) {
	m := testProfilesModel()
	m.profiles.cursor = 2 // last index
	result, _ := m.updateProfiles(specialKeyMsg(tea.KeyDown))
	rm := result.(model)
	if rm.profiles.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (clamped)", rm.profiles.cursor)
	}
}

func TestUpdateProfilesNavigateBlockedDuringConfirm(t *testing.T) {
	m := testProfilesModel()
	m.profiles.confirm = "alpha"
	m.profiles.cursor = 1
	result, _ := m.updateProfiles(keyMsg("k"))
	rm := result.(model)
	if rm.profiles.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (blocked by confirm)", rm.profiles.cursor)
	}
}

func TestUpdateProfilesEnter(t *testing.T) {
	m := testProfilesModel()
	m.width = 80
	m.height = 40
	result, _ := m.updateProfiles(specialKeyMsg(tea.KeyEnter))
	rm := result.(model)
	if rm.currentScreen != screenSecrets {
		t.Errorf("screen = %d, want secrets", rm.currentScreen)
	}
	if rm.secrets.profileName != "beta" {
		t.Errorf("profileName = %q, want beta", rm.secrets.profileName)
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
	if rm.profiles.confirm != "beta" {
		t.Errorf("confirm = %q, want beta", rm.profiles.confirm)
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

func TestUpdateProfilesNavigationClearsMessage(t *testing.T) {
	m := testProfilesModel()
	m.profiles.message = "old message"
	m.profiles.cursor = 1
	result, _ := m.updateProfiles(keyMsg("j"))
	rm := result.(model)
	if rm.profiles.message != "" {
		t.Error("navigation should clear message")
	}
}

// --- updateSecrets key tests ---

func testSecretsModel() model {
	m := NewModel(nil)
	m.currentScreen = screenSecrets
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

func TestUpdateSecretsEsc(t *testing.T) {
	m := testSecretsModel()
	result, _ := m.updateSecrets(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.currentScreen != screenProfiles {
		t.Errorf("screen = %d, want profiles", rm.currentScreen)
	}
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
	// Simulate enter on K1 again — should toggle off
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
	// default key clears reveal
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
	// default branch skipped during confirm
	if rm.secrets.revealKey != "K2" {
		t.Error("reveal should not be cleared during confirm")
	}
}

// --- updateSecretForm key tests ---

func TestUpdateSecretFormEscFromProfile(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecretForm
	m.secretForm = newProfileFormModel()
	result, _ := m.updateSecretForm(specialKeyMsg(tea.KeyEsc))
	rm := result.(model)
	if rm.currentScreen != screenProfiles {
		t.Errorf("screen = %d, want profiles", rm.currentScreen)
	}
}

func TestUpdateSecretFormEscFromSecret(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenSecretForm
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
	// key() is empty → validation error
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
	// key() and value() both empty → validation error
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
	m.secretForm = newSecretFormModel("p")
	// Tab should delegate to form's Update
	result, _ := m.updateSecretForm(specialKeyMsg(tea.KeyTab))
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

func TestModelUpdateDelegatesDashboardLoadedMsg(t *testing.T) {
	m := NewModel(nil)
	m.currentScreen = screenDashboard
	status := &vault.Status{Initialized: true}

	result, _ := m.Update(dashboardLoadedMsg{status: status})
	rm := result.(model)
	if !rm.dashboard.loaded {
		t.Error("dashboard should be loaded")
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

// --- Model.View for all screens ---

func TestModelViewAllScreens(t *testing.T) {
	screens := []struct {
		name   string
		screen screen
	}{
		{"dashboard", screenDashboard},
		{"profiles", screenProfiles},
		{"secrets", screenSecrets},
		{"secretForm", screenSecretForm},
	}
	for _, tt := range screens {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(nil)
			m.width = 80
			m.height = 40
			m.currentScreen = tt.screen
			if tt.screen == screenSecrets {
				m.secrets = newSecretsModel("p", 80, 40)
			}
			if tt.screen == screenSecretForm {
				m.secretForm = newSecretFormModel("p")
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
	if v != "" {
		t.Errorf("expected empty view for unknown screen, got %q", v)
	}
}

// --- Profiles view with message ---

func TestProfilesViewWithMessage(t *testing.T) {
	p := profilesModel{
		loaded:   true,
		profiles: []vault.ProfileInfo{{Name: "test"}},
		message:  "Profile deleted.",
	}
	v := p.View(80)
	if !strings.Contains(v, "Profile deleted.") {
		t.Error("expected message in view")
	}
}

// --- secretsLoadedMsg with error ---

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

// --- secretActionMsg error ---

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

// --- profilesLoadedMsg with error ---

func TestProfilesModelUpdateLoadError(t *testing.T) {
	p := profilesModel{}
	p, _ = p.Update(profilesLoadedMsg{err: fmt.Errorf("load failed")})
	if !p.loaded {
		t.Error("expected loaded even on error")
	}
	if len(p.profiles) != 0 {
		t.Error("profiles should be empty on error")
	}
}
