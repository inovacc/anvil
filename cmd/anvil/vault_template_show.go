package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultTemplateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		def, err := v.GetTemplate(args[0])
		if err != nil {
			return err
		}

		outputResult(cmd, def, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Template: %s\n", def.Name)
			if def.Description != "" {
				_, _ = fmt.Fprintf(w, "Description: %s\n", def.Description)
			}

			if len(def.Variables) > 0 {
				_, _ = fmt.Fprintln(w, "\nVariables:")
				for _, v := range def.Variables {
					req := ""
					if v.Required {
						req = " (required)"
					}
					def := ""
					if v.Default != "" {
						def = fmt.Sprintf(" [default: %s]", v.Default)
					}
					_, _ = fmt.Fprintf(w, "  %s%s%s - %s\n", v.Name, req, def, v.Description)
				}
			}

			_, _ = fmt.Fprintln(w, "\nSecrets:")
			for _, s := range def.Secrets {
				_, _ = fmt.Fprintf(w, "  %s = %s\n", s.Key, s.Value)
			}
		})
		return nil
	},
}

func init() {
	vaultTemplateCmd.AddCommand(vaultTemplateShowCmd)
}
