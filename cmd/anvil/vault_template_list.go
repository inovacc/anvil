package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultTemplateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		templates, err := v.ListTemplates()
		if err != nil {
			return err
		}

		outputResult(cmd, templates, func() {
			w := cmd.OutOrStdout()
			if len(templates) == 0 {
				_, _ = fmt.Fprintln(w, "No templates found.")
				return
			}

			for _, t := range templates {
				marker := "  "
				if t.Builtin {
					marker = "* "
				}

				_, _ = fmt.Fprintf(w, "%s%s (%d vars, %d secrets) - %s\n", marker, t.Name, t.VarCount, t.SecretCount, t.Description)
			}
		})

		return nil
	},
}

func init() {
	vaultTemplateCmd.AddCommand(vaultTemplateListCmd)
}
