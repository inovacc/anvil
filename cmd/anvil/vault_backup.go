package main

import (
	"fmt"
	"os"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultBackupCmd = &cobra.Command{
	Use:   "backup <file>",
	Short: "Create an encrypted vault backup",
	Long:  "Exports the entire vault (all profiles, secrets, version history, and password hash) as an encrypted archive file. Requires password verification.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			var err error

			password, err = readPassword("Password: ")
			if err != nil {
				return err
			}
		}

		encrypted, err := v.Backup(password)
		if err != nil {
			return err
		}

		if err := os.WriteFile(args[0], encrypted, 0600); err != nil {
			return fmt.Errorf("write backup file: %w", err)
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
			File    string `json:"file"`
		}{
			Message: fmt.Sprintf("Backup written to %s", args[0]),
			File:    args[0],
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Backup written to %s\n", args[0])
		})

		return nil
	},
}

var vaultRestoreCmd = &cobra.Command{
	Use:   "restore <file>",
	Short: "Restore vault from an encrypted backup",
	Long:  "Imports a backup archive into the vault, restoring all profiles, secrets, version history, and password hash. The vault must be initialized but empty.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			var err error

			password, err = readPassword("Password: ")
			if err != nil {
				return err
			}
		}

		encrypted, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read backup file: %w", err)
		}

		if err := v.Restore(encrypted, password); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
			File    string `json:"file"`
		}{
			Message: fmt.Sprintf("Vault restored from %s", args[0]),
			File:    args[0],
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Vault restored from %s\n", args[0])
		})

		return nil
	},
}

func init() {
	vaultBackupCmd.Flags().String("password", "", "Password (non-interactive mode)")
	vaultRestoreCmd.Flags().String("password", "", "Password (non-interactive mode)")
	registerCommand(vaultBackupCmd)
	registerCommand(vaultRestoreCmd)
}
