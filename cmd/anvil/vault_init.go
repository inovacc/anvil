package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the encrypted vault",
	Long:  "Creates a new vault with a machine-bound master key and BIP-39 recovery phrase.",
	RunE: func(cmd *cobra.Command, args []string) error {
		mnemonic, err := vault.InitWithRecovery(nil)
		if err != nil {
			if errors.Is(err, vault.ErrAlreadyInitialized) {
				outputResult(cmd, struct {
					Initialized bool   `json:"initialized"`
					DBPath      string `json:"db_path"`
					Message     string `json:"message"`
				}{true, vault.DefaultDBPath(), "Vault is already initialized."}, func() {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vault is already initialized.")
				})

				return nil
			}

			return err
		}

		words := strings.Fields(mnemonic)

		outputResult(cmd, struct {
			Initialized bool     `json:"initialized"`
			DBPath      string   `json:"db_path"`
			Message     string   `json:"message"`
			Mnemonic    string   `json:"mnemonic"`
			Words       []string `json:"words"`
		}{true, vault.DefaultDBPath(), "Vault initialized successfully.", mnemonic, words}, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "Vault initialized successfully.")
			_, _ = fmt.Fprintf(w, "Database: %s\n", vault.DefaultDBPath())
			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "Recovery Phrase (24 words) — write this down and store securely:")

			_, _ = fmt.Fprintln(w, "")
			for i, word := range words {
				_, _ = fmt.Fprintf(w, "  %2d. %s\n", i+1, word)
			}

			_, _ = fmt.Fprintln(w, "")
			_, _ = fmt.Fprintln(w, "WARNING: This is the ONLY time the recovery phrase is shown.")
			_, _ = fmt.Fprintln(w, "         If you lose it, vault recovery on a new machine is impossible.")
		})

		return nil
	},
}

func init() {
	registerCommand(vaultInitCmd)
}
