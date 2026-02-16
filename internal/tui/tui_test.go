package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

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
