package main

import (
	"fmt"
	"os"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var keyExportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export a key in PEM format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		private, _ := cmd.Flags().GetBool("private")
		output, _ := cmd.Flags().GetString("output")

		pemData, err := v.ExportKeyPEM(args[0], private)
		if err != nil {
			return err
		}

		if output != "" {
			if err := os.WriteFile(output, pemData, 0o600); err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			outputResult(cmd, struct {
				File string `json:"file"`
				Name string `json:"name"`
			}{output, args[0]}, func() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Key exported to %s\n", output)
			})

			return nil
		}

		_, _ = fmt.Fprint(cmd.OutOrStdout(), string(pemData))

		return nil
	},
}

func init() {
	keyExportCmd.Flags().Bool("private", false, "Export private key (default: public only)")
	keyExportCmd.Flags().StringP("output", "o", "", "Output file path")
	keyCmd.AddCommand(keyExportCmd)
}
