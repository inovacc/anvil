package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultRecoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover vault using a 24-word recovery phrase",
	Long:  "Recover the vault master key on a new machine using the BIP-39 mnemonic shown during init.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mnemonic, _ := cmd.Flags().GetString("mnemonic")

		if mnemonic == "" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Enter your 24-word recovery phrase (all on one line):")

			scanner := bufio.NewScanner(cmd.InOrStdin())
			if scanner.Scan() {
				mnemonic = strings.TrimSpace(scanner.Text())
			}

			if mnemonic == "" {
				return fmt.Errorf("no recovery phrase provided")
			}
		}

		if err := vault.RecoverFromMnemonic(mnemonic, nil); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{"Vault recovered and re-sealed to this machine."}, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "Vault recovered and re-sealed to this machine.")
			_, _ = fmt.Fprintf(w, "Database: %s\n", vault.DefaultDBPath())
		})

		return nil
	},
}

func init() {
	vaultRecoverCmd.Flags().String("mnemonic", "", "24-word recovery phrase (for scripting)")
	registerCommand(vaultRecoverCmd)
}
