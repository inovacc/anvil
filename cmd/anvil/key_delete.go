package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var keyDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		algorithm, _ := cmd.Flags().GetString("algorithm")

		if err := v.DeleteKey(args[0], algorithm); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Key %q deleted.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Key %q deleted.\n", args[0])
		})

		return nil
	},
}

func init() {
	keyDeleteCmd.Flags().StringP("algorithm", "a", "", "Key algorithm (delete specific algorithm only)")
	keyCmd.AddCommand(keyDeleteCmd)
}
