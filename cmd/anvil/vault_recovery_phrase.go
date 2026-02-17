package main

import (
	"fmt"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultRecoveryPhraseCmd = &cobra.Command{
	Use:   "recovery-phrase",
	Short: "Show the vault recovery phrase",
	Long:  "Display the 24-word BIP-39 recovery phrase for the current vault. Requires password verification.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		password, _ := cmd.Flags().GetString("password")

		hasPw, err := v.HasPassword()
		if err != nil {
			return err
		}

		if hasPw {
			if password == "" {
				return &vault.UserError{
					Message: "password required",
					Hint:    "Use --password to provide the vault password",
				}
			}

			if err := v.VerifyPassword(password); err != nil {
				return err
			}
		}

		mnemonic, err := v.ShowRecoveryPhrase()
		if err != nil {
			return err
		}

		words := strings.Fields(mnemonic)

		outputResult(cmd, struct {
			Mnemonic string   `json:"mnemonic"`
			Words    []string `json:"words"`
		}{mnemonic, words}, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "Recovery Phrase (24 words):")
			_, _ = fmt.Fprintln(w, "")
			for i, word := range words {
				_, _ = fmt.Fprintf(w, "  %2d. %s\n", i+1, word)
			}
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "WARNING: Store this phrase securely. Anyone with these words can recover your vault.")
		})
		return nil
	},
}

func init() {
	vaultRecoveryPhraseCmd.Flags().String("password", "", "Vault password for verification")
	registerCommand(vaultRecoveryPhraseCmd)
}
