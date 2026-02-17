package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var vaultTemplateCreateCmd = &cobra.Command{
	Use:   "create <file>",
	Short: "Create a template from a YAML or JSON file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var def vault.TemplateDefinition

		// Try YAML first, then JSON.
		if err := yaml.Unmarshal(data, &def); err != nil {
			if err := json.Unmarshal(data, &def); err != nil {
				return fmt.Errorf("parse template file (tried YAML and JSON): %w", err)
			}
		}

		if def.Name == "" {
			return &vault.UserError{
				Message: "template name is required",
				Hint:    "Add a 'name' field to your template file",
			}
		}

		if len(def.Secrets) == 0 {
			return &vault.UserError{
				Message: "template must define at least one secret",
				Hint:    "Add a 'secrets' section to your template file",
			}
		}

		if err := v.CreateTemplate(&def); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}{def.Name, fmt.Sprintf("Template %q created.", def.Name)}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Template %q created.\n", def.Name)
		})
		return nil
	},
}

func init() {
	vaultTemplateCmd.AddCommand(vaultTemplateCreateCmd)
}
