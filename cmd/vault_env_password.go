package cmd

import (
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultEnvPasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Manage vault env password",
}

var vaultEnvPasswordSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set or update the vault env password",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		has, err := v.HasPassword()
		if err != nil {
			return err
		}

		if has {
			current, err := readPassword("Current password: ")
			if err != nil {
				return err
			}
			if err := v.VerifyPassword(current); err != nil {
				return err
			}
		}

		password, err := readPassword("New password: ")
		if err != nil {
			return err
		}
		if len(password) < 8 {
			return fmt.Errorf("password must be at least 8 characters")
		}

		confirm, err := readPassword("Confirm password: ")
		if err != nil {
			return err
		}
		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}

		if err := v.SetPassword(password); err != nil {
			return err
		}

		fmt.Println("Password set successfully.")
		return nil
	},
}

var vaultEnvPasswordResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Remove the vault env password",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		has, err := v.HasPassword()
		if err != nil {
			return err
		}
		if !has {
			fmt.Println("No password is set.")
			return nil
		}

		current, err := readPassword("Current password: ")
		if err != nil {
			return err
		}
		if err := v.VerifyPassword(current); err != nil {
			return err
		}

		if err := v.DeletePassword(); err != nil {
			return err
		}

		// Also revoke any active release
		_ = v.EnvRevoke()

		fmt.Println("Password removed.")
		return nil
	},
}

func init() {
	vaultEnvPasswordCmd.AddCommand(vaultEnvPasswordSetCmd)
	vaultEnvPasswordCmd.AddCommand(vaultEnvPasswordResetCmd)
	vaultEnvCmd.AddCommand(vaultEnvPasswordCmd)
}
