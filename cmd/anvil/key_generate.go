package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var keyGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate a new key pair",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		algorithm, _ := cmd.Flags().GetString("algorithm")
		description, _ := cmd.Flags().GetString("description")

		info, err := v.GenerateKey(args[0], algorithm, description)
		if err != nil {
			return err
		}

		outputResult(cmd, info, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Key %q generated.\n", info.Name)
			_, _ = fmt.Fprintf(w, "  Algorithm:   %s\n", info.Algorithm)
			_, _ = fmt.Fprintf(w, "  Fingerprint: %s\n", info.Fingerprint)
		})

		return nil
	},
}

func init() {
	keyGenerateCmd.Flags().StringP("algorithm", "a", "ed25519", "Key algorithm (ed25519, ecdsa-p256)")
	keyGenerateCmd.Flags().StringP("description", "d", "", "Key description")
	keyCmd.AddCommand(keyGenerateCmd)
}
