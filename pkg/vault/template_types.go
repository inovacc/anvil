package vault

// TemplateVariable defines a variable that can be set when applying a template.
type TemplateVariable struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
	Default     string `json:"default,omitempty" yaml:"default,omitempty"`
	Secret      bool   `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// TemplateSecret defines a secret that will be created when a template is applied.
type TemplateSecret struct {
	Key         string `json:"key" yaml:"key"`
	Value       string `json:"value" yaml:"value"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// TemplateDefinition is the full definition of a secret template.
type TemplateDefinition struct {
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   []TemplateVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
	Secrets     []TemplateSecret   `json:"secrets" yaml:"secrets"`
}

// TemplateInfo represents summary information about a template.
type TemplateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	VarCount    int    `json:"var_count"`
	SecretCount int    `json:"secret_count"`
	Builtin     bool   `json:"builtin"`
}
