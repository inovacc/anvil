package cmd

import (
	"fmt"
	"os"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultShareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share secrets between machines",
}

var vaultShareExportCmd = &cobra.Command{
	Use:   "export <file>",
	Short: "Export encrypted secrets for sharing",
	Long:  "Encrypts a profile's secrets with a passphrase and writes the result to a file for machine-to-machine transfer.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		passphrase, _ := cmd.Flags().GetString("passphrase")
		if passphrase == "" {
			var err error
			passphrase, err = readPassword("Passphrase: ")
			if err != nil {
				return err
			}
		}

		encrypted, err := v.ShareExport(profileName, passphrase)
		if err != nil {
			return err
		}

		if err := os.WriteFile(args[0], encrypted, 0600); err != nil {
			return fmt.Errorf("write share file: %w", err)
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
			File    string `json:"file"`
			Profile string `json:"profile"`
		}{
			Message: fmt.Sprintf("Shared export written to %s", args[0]),
			File:    args[0],
			Profile: profileName,
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Shared export written to %s\n", args[0])
		})

		return nil
	},
}

var vaultShareImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import shared secrets from another machine",
	Long:  "Decrypts a shared export file and imports the secrets into the specified profile (or the original profile if not specified).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		passphrase, _ := cmd.Flags().GetString("passphrase")
		if passphrase == "" {
			var err error
			passphrase, err = readPassword("Passphrase: ")
			if err != nil {
				return err
			}
		}

		encrypted, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read share file: %w", err)
		}

		export, err := v.ShareImport(encrypted, passphrase, profileName)
		if err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message     string `json:"message"`
			File        string `json:"file"`
			FromProfile string `json:"from_profile"`
			Secrets     int    `json:"secrets_imported"`
		}{
			Message:     fmt.Sprintf("Imported %d secrets from %q", len(export.Secrets), export.ProfileName),
			File:        args[0],
			FromProfile: export.ProfileName,
			Secrets:     len(export.Secrets),
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d secrets from %q\n", len(export.Secrets), export.ProfileName)
		})

		return nil
	},
}

func init() {
	vaultShareExportCmd.Flags().String("profile", "default", "Profile to export")
	vaultShareExportCmd.Flags().String("passphrase", "", "Encryption passphrase (non-interactive mode)")
	vaultShareImportCmd.Flags().String("profile", "", "Target profile (uses original if empty)")
	vaultShareImportCmd.Flags().String("passphrase", "", "Decryption passphrase (non-interactive mode)")
	vaultShareCmd.AddCommand(vaultShareExportCmd)
	vaultShareCmd.AddCommand(vaultShareImportCmd)
	vaultCmd.AddCommand(vaultShareCmd)
}
