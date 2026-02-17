package vault_test

import (
	"errors"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func TestTemplateCRUD(t *testing.T) {
	v, _ := initAndOpen(t)

	def := &vault.TemplateDefinition{
		Name:        "test-template",
		Description: "A test template",
		Variables: []vault.TemplateVariable{
			{Name: "host", Required: true, Default: "localhost"},
			{Name: "port", Default: "5432"},
		},
		Secrets: []vault.TemplateSecret{
			{Key: "DB_HOST", Value: "{{.host}}"},
			{Key: "DB_PORT", Value: "{{.port}}"},
		},
	}

	// Create.
	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Duplicate create.
	err := v.CreateTemplate(def)
	if !errors.Is(err, vault.ErrTemplateExists) {
		t.Fatalf("expected ErrTemplateExists, got %v", err)
	}

	// Get.
	got, err := v.GetTemplate("test-template")
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}

	if got.Name != "test-template" {
		t.Errorf("Name = %q, want %q", got.Name, "test-template")
	}

	if len(got.Variables) != 2 {
		t.Errorf("Variables count = %d, want 2", len(got.Variables))
	}

	if len(got.Secrets) != 2 {
		t.Errorf("Secrets count = %d, want 2", len(got.Secrets))
	}

	// List.
	templates, err := v.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	found := false

	for _, ti := range templates {
		if ti.Name == "test-template" {
			found = true

			if ti.VarCount != 2 {
				t.Errorf("VarCount = %d, want 2", ti.VarCount)
			}

			if ti.SecretCount != 2 {
				t.Errorf("SecretCount = %d, want 2", ti.SecretCount)
			}

			if ti.Builtin {
				t.Error("expected Builtin = false")
			}
		}
	}

	if !found {
		t.Error("test-template not found in list")
	}

	// Delete.
	if err := v.DeleteTemplate("test-template"); err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}

	// Get after delete.
	_, err = v.GetTemplate("test-template")
	if !errors.Is(err, vault.ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplateApply(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("tmpl", "template test", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	def := &vault.TemplateDefinition{
		Name: "db-template",
		Variables: []vault.TemplateVariable{
			{Name: "host", Required: true, Default: "localhost"},
			{Name: "port", Default: "5432"},
			{Name: "password", Required: true},
		},
		Secrets: []vault.TemplateSecret{
			{Key: "DB_HOST", Value: "{{.host}}", Description: "host"},
			{Key: "DB_URL", Value: "postgres://user:{{.password}}@{{.host}}:{{.port}}/db", Description: "url"},
		},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Apply with variables.
	vars := map[string]string{
		"host":     "db.example.com",
		"password": "secret123",
	}
	if err := v.ApplyTemplate("db-template", "tmpl", vars); err != nil {
		t.Fatalf("ApplyTemplate: %v", err)
	}

	// Verify secrets were created.
	val, err := v.Get("DB_HOST", "tmpl")
	if err != nil {
		t.Fatalf("Get DB_HOST: %v", err)
	}

	if val != "db.example.com" {
		t.Errorf("DB_HOST = %q, want %q", val, "db.example.com")
	}

	val, err = v.Get("DB_URL", "tmpl")
	if err != nil {
		t.Fatalf("Get DB_URL: %v", err)
	}

	if val != "postgres://user:secret123@db.example.com:5432/db" {
		t.Errorf("DB_URL = %q", val)
	}
}

func TestTemplateApplyMissingRequiredVar(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("tmpl2", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	def := &vault.TemplateDefinition{
		Name: "req-template",
		Variables: []vault.TemplateVariable{
			{Name: "token", Required: true},
		},
		Secrets: []vault.TemplateSecret{
			{Key: "TOKEN", Value: "{{.token}}"},
		},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Apply without required var.
	err := v.ApplyTemplate("req-template", "tmpl2", map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}

	ue, ok := vault.IsUserError(err)
	if !ok {
		t.Fatalf("expected UserError, got %T: %v", err, err)
	}

	if ue.Hint == "" {
		t.Error("expected hint on missing var error")
	}
}

func TestDeleteBuiltinTemplate(t *testing.T) {
	v, _ := initAndOpen(t)

	// Built-in templates are loaded on Open.
	err := v.DeleteTemplate("postgres")
	if !errors.Is(err, vault.ErrBuiltinTemplate) {
		t.Fatalf("expected ErrBuiltinTemplate, got %v", err)
	}
}

func TestTemplateApplyDefaults(t *testing.T) {
	v, _ := initAndOpen(t)

	if err := v.CreateProfile("def", "", true); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	def := &vault.TemplateDefinition{
		Name: "default-template",
		Variables: []vault.TemplateVariable{
			{Name: "region", Default: "us-east-1"},
		},
		Secrets: []vault.TemplateSecret{
			{Key: "REGION", Value: "{{.region}}"},
		},
	}

	if err := v.CreateTemplate(def); err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Apply without setting region — should use default.
	if err := v.ApplyTemplate("default-template", "def", map[string]string{}); err != nil {
		t.Fatalf("ApplyTemplate: %v", err)
	}

	val, err := v.Get("REGION", "def")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "us-east-1" {
		t.Errorf("REGION = %q, want %q", val, "us-east-1")
	}
}

func TestBuiltinTemplatesLoaded(t *testing.T) {
	v, _ := initAndOpen(t)

	templates, err := v.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}

	builtins := map[string]bool{
		"postgres":     false,
		"mysql":        false,
		"redis":        false,
		"aws":          false,
		"github-token": false,
	}

	for _, ti := range templates {
		if _, ok := builtins[ti.Name]; ok {
			builtins[ti.Name] = true
			if !ti.Builtin {
				t.Errorf("expected %q to be builtin", ti.Name)
			}
		}
	}

	for name, found := range builtins {
		if !found {
			t.Errorf("builtin template %q not found", name)
		}
	}
}

func TestGetTemplateNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GetTemplate("nonexistent")
	if !errors.Is(err, vault.ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound, got %v", err)
	}
}

func TestDeleteTemplateNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	err := v.DeleteTemplate("nonexistent")
	if !errors.Is(err, vault.ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound, got %v", err)
	}
}
