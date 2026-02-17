package cmd

import (
	"fmt"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultTemplateApplyCmd = &cobra.Command{
	Use:   "apply <name>",
	Short: "Apply a template to create secrets",
	Long:  "Applies a template, interpolating variables and creating secrets in the target profile.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		setFlags, _ := cmd.Flags().GetStringSlice("set")

		vars := make(map[string]string)
		for _, s := range setFlags {
			key, value, ok := strings.Cut(s, "=")
			if !ok {
				return fmt.Errorf("invalid --set format %q, expected KEY=VALUE", s)
			}
			vars[key] = value
		}

		if err := v.ApplyTemplate(args[0], profileName, vars); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Template string `json:"template"`
			Message  string `json:"message"`
		}{args[0], fmt.Sprintf("Template %q applied.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Template %q applied.\n", args[0])
		})
		return nil
	},
}

func init() {
	vaultTemplateApplyCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultTemplateApplyCmd.Flags().StringSlice("set", nil, "Set template variable (KEY=VALUE), can be repeated")
	vaultTemplateCmd.AddCommand(vaultTemplateApplyCmd)
}
