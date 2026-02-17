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

			tw := tableWriter(w)
			_, _ = fmt.Fprintln(tw, "NAME\tDEFAULT\tSECRETS\tDESCRIPTION\tCREATED")
			for _, p := range profiles {
				def := ""
				if p.IsDefault {
					def = "*"
				}
				created := "-"
				if !p.CreatedAt.IsZero() {
					created = p.CreatedAt.Format("2006-01-02 15:04:05")
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n",
					p.Name,
					def,
					p.SecretCount,
					p.Description,
					created,
				)
			}
			_ = tw.Flush()
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
