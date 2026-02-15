package vault

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
)

// ErrTemplateNotFound is returned when a template does not exist.
var ErrTemplateNotFound = &UserError{
	Message: "template not found",
	Hint:    "Run 'anvil vault template list' to see available templates",
}

// ErrTemplateExists is returned when trying to create a template that already exists.
var ErrTemplateExists = &UserError{
	Message: "template already exists",
	Hint:    "Choose a different name or delete the existing template",
}

// ErrBuiltinTemplate is returned when trying to delete a built-in template.
var ErrBuiltinTemplate = &UserError{
	Message: "cannot delete built-in template",
	Hint:    "Built-in templates are read-only",
}

// CreateTemplate stores a new template definition.
func (v *Vault) CreateTemplate(def *TemplateDefinition) error {
	exists, err := v.store.TemplateExists(def.Name)
	if err != nil {
		return fmt.Errorf("check template: %w", err)
	}

	if exists {
		return ErrTemplateExists
	}

	data, err := json.Marshal(def)
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}

	if err := v.store.CreateTemplate(def.Name, def.Description, string(data), false); err != nil {
		return fmt.Errorf("save template: %w", err)
	}

	v.logAudit("template.create", "", "", def.Name)

	return nil
}

// GetTemplate retrieves a template definition by name.
func (v *Vault) GetTemplate(name string) (*TemplateDefinition, error) {
	row, err := v.store.GetTemplate(name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTemplateNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	var def TemplateDefinition
	if err := json.Unmarshal([]byte(row.TemplateData), &def); err != nil {
		return nil, fmt.Errorf("unmarshal template: %w", err)
	}

	return &def, nil
}

// ListTemplates returns summary info for all templates.
func (v *Vault) ListTemplates() ([]TemplateInfo, error) {
	rows, err := v.store.ListTemplates()
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	templates := make([]TemplateInfo, 0, len(rows))
	for _, row := range rows {
		var def TemplateDefinition
		if err := json.Unmarshal([]byte(row.TemplateData), &def); err != nil {
			return nil, fmt.Errorf("unmarshal template %q: %w", row.Name, err)
		}

		builtin := false
		if row.Builtin != nil && *row.Builtin == 1 {
			builtin = true
		}

		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}

		templates = append(templates, TemplateInfo{
			Name:        row.Name,
			Description: desc,
			VarCount:    len(def.Variables),
			SecretCount: len(def.Secrets),
			Builtin:     builtin,
		})
	}

	return templates, nil
}

// DeleteTemplate removes a template. Built-in templates cannot be deleted.
func (v *Vault) DeleteTemplate(name string) error {
	row, err := v.store.GetTemplate(name)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTemplateNotFound
	}

	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	if row.Builtin != nil && *row.Builtin == 1 {
		return ErrBuiltinTemplate
	}

	if err := v.store.DeleteTemplate(name); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}

	v.logAudit("template.delete", "", "", name)

	return nil
}

// ApplyTemplate applies a template to a profile, creating secrets from the template definition.
// Variables in secret values are interpolated using Go text/template syntax.
func (v *Vault) ApplyTemplate(name, profileName string, vars map[string]string) error {
	def, err := v.GetTemplate(name)
	if err != nil {
		return err
	}

	// Validate required variables and apply defaults.
	templateVars := make(map[string]string)
	for _, variable := range def.Variables {
		val, ok := vars[variable.Name]
		if !ok || val == "" {
			if variable.Required && variable.Default == "" {
				return &UserError{
					Message: fmt.Sprintf("required variable %q is not set", variable.Name),
					Hint:    fmt.Sprintf("Use --set %s=<value> to provide it", variable.Name),
				}
			}

			val = variable.Default
		}

		templateVars[variable.Name] = val
	}

	// Also include any extra vars passed by the user.
	for k, val := range vars {
		if _, exists := templateVars[k]; !exists {
			templateVars[k] = val
		}
	}

	profile, err := v.resolveProfile(profileName)
	if err != nil {
		return err
	}

	// Interpolate and set each secret.
	for _, secret := range def.Secrets {
		value, err := interpolate(secret.Value, templateVars)
		if err != nil {
			return fmt.Errorf("interpolate secret %q: %w", secret.Key, err)
		}

		if err := v.Set(secret.Key, value, profile, secret.Description); err != nil {
			return fmt.Errorf("set secret %q: %w", secret.Key, err)
		}
	}

	v.logAudit("template.apply", profile, "", fmt.Sprintf("template %q, %d secrets", name, len(def.Secrets)))

	return nil
}

// interpolate renders a Go template string with the given variables.
func interpolate(tmplStr string, vars map[string]string) (string, error) {
	tmpl, err := template.New("secret").Option("missingkey=error").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// LoadBuiltinTemplates inserts built-in templates into the store if they don't already exist.
func (v *Vault) LoadBuiltinTemplates() error {
	for _, def := range builtinTemplates {
		exists, err := v.store.TemplateExists(def.Name)
		if err != nil {
			return fmt.Errorf("check template %q: %w", def.Name, err)
		}

		if exists {
			continue
		}

		data, err := json.Marshal(def)
		if err != nil {
			return fmt.Errorf("marshal template %q: %w", def.Name, err)
		}

		if err := v.store.CreateTemplate(def.Name, def.Description, string(data), true); err != nil {
			return fmt.Errorf("save template %q: %w", def.Name, err)
		}
	}

	return nil
}
