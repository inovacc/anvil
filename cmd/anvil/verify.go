package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		keyName, _ := cmd.Flags().GetString("key")
		filePath, _ := cmd.Flags().GetString("file")
		text, _ := cmd.Flags().GetString("string")
		sigStr, _ := cmd.Flags().GetString("signature")
		sigFile, _ := cmd.Flags().GetString("signature-file")

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

		switch {
		case sigFile != "":
			raw, err := os.ReadFile(sigFile)
			if err != nil {
				return fmt.Errorf("read signature file: %w", err)
			}

			sigStr = strings.TrimSpace(string(raw))
		case sigStr != "":
			// already set
		default:
			return fmt.Errorf("provide --signature or --signature-file")
		}

		result, err := v.Verify(keyName, data, sigStr)
		if err != nil {
			return err
		}

		outputResult(cmd, result, func() {
			if result.Valid {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Signature is valid.")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Signature is INVALID.")
			}
		})

		if !result.Valid {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	verifyCmd.Flags().StringP("key", "k", "", "Key name (required)")
	verifyCmd.Flags().StringP("file", "f", "", "File that was signed")
	verifyCmd.Flags().StringP("string", "s", "", "String that was signed")
	verifyCmd.Flags().String("signature", "", "Base64-encoded signature")
	verifyCmd.Flags().String("signature-file", "", "File containing the signature")
	_ = verifyCmd.MarkFlagRequired("key")
	registerCommand(verifyCmd)
}
