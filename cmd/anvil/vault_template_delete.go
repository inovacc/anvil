package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultTemplateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		if err := v.DeleteTemplate(args[0]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Template %q deleted.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Template %q deleted.\n", args[0])
		})

		return nil
	},
}

func init() {
	vaultTemplateCmd.AddCommand(vaultTemplateDeleteCmd)
}
