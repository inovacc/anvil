package main

import (
	"fmt"
	"os"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var keyImportCmd = &cobra.Command{
	Use:   "import <name> <pem-file>",
	Short: "Import a PEM-encoded key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		pemData, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("read PEM file: %w", err)
		}

		description, _ := cmd.Flags().GetString("description")

		info, err := v.ImportKeyPEM(args[0], pemData, description)
		if err != nil {
			return err
		}

		outputResult(cmd, info, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Key %q imported.\n", info.Name)
			_, _ = fmt.Fprintf(w, "  Algorithm:   %s\n", info.Algorithm)
			_, _ = fmt.Fprintf(w, "  Fingerprint: %s\n", info.Fingerprint)
		})

		return nil
	},
}

func init() {
	keyImportCmd.Flags().StringP("description", "d", "", "Key description")
	keyCmd.AddCommand(keyImportCmd)
}
