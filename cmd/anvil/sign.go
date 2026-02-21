package main

import (
	"fmt"
	"os"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign data with a key",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		keyName, _ := cmd.Flags().GetString("key")
		filePath, _ := cmd.Flags().GetString("file")
		text, _ := cmd.Flags().GetString("string")
		output, _ := cmd.Flags().GetString("output")

		var data []byte

		switch {
		case filePath != "":
			data, err = os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
		case text != "":
			data = []byte(text)
		default:
			return fmt.Errorf("provide --file or --string")
		}

		result, err := v.Sign(keyName, data)
		if err != nil {
			return err
		}

		if output != "" {
			if err := os.WriteFile(output, []byte(result.Signature), 0o644); err != nil {
				return fmt.Errorf("write signature: %w", err)
			}

			outputResult(cmd, result, func() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Signature written to %s\n", output)
			})

			return nil
		}

		outputResult(cmd, result, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), result.Signature)
		})

		return nil
	},
}

func init() {
	signCmd.Flags().StringP("key", "k", "", "Key name (required)")
	signCmd.Flags().StringP("file", "f", "", "File to sign")
	signCmd.Flags().StringP("string", "s", "", "String to sign")
	signCmd.Flags().StringP("output", "o", "", "Output file for signature")
	_ = signCmd.MarkFlagRequired("key")
	registerCommand(signCmd)
}
