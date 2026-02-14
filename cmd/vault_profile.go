package cmd

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage vault profiles",
}

var vaultProfileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new vault profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		description, _ := cmd.Flags().GetString("description")
		isDefault, _ := cmd.Flags().GetBool("default")

		if err := v.CreateProfile(args[0], description, isDefault); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Profile %q created.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Profile %q created.\n", args[0])
		})
		return nil
	},
}

var vaultProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all vault profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profiles, err := v.ListProfiles()
		if err != nil {
			return err
		}

		outputResult(cmd, profiles, func() {
			w := cmd.OutOrStdout()
			if len(profiles) == 0 {
				_, _ = fmt.Fprintln(w, "No profiles found.")
				return
			}

			for _, p := range profiles {
				marker := "  "
				if p.IsDefault {
					marker = "* "
				}
				desc := ""
				if p.Description != "" {
					desc = fmt.Sprintf(" - %s", p.Description)
				}
				_, _ = fmt.Fprintf(w, "%s%s (%d secrets)%s\n", marker, p.Name, p.SecretCount, desc)
			}
		})
		return nil
	},
}

var vaultProfileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a vault profile and all its secrets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		if err := v.DeleteProfile(args[0]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Profile %q deleted.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Profile %q deleted.\n", args[0])
		})
		return nil
	},
}

var vaultProfileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set a profile as the default",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		if err := v.UseProfile(args[0]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Now using profile %q.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Now using profile %q.\n", args[0])
		})
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultProfileCmd)

	vaultProfileCreateCmd.Flags().StringP("description", "d", "", "Profile description")
	vaultProfileCreateCmd.Flags().Bool("default", false, "Set as default profile")

	vaultProfileCmd.AddCommand(vaultProfileCreateCmd)
	vaultProfileCmd.AddCommand(vaultProfileListCmd)
	vaultProfileCmd.AddCommand(vaultProfileDeleteCmd)
	vaultProfileCmd.AddCommand(vaultProfileUseCmd)
}
