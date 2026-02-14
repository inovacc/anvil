package cmd

import (
	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

func init() {
	// Make the built-in completion command visible.
	rootCmd.CompletionOptions.HiddenDefaultCmd = false

	// Register dynamic completions for commands that take profile names.
	for _, c := range []*cobra.Command{vaultProfileDeleteCmd, vaultProfileUseCmd} {
		c.ValidArgsFunction = completeProfileNames
	}

	// Register dynamic completions for commands that take secret keys.
	for _, c := range []*cobra.Command{vaultGetCmd, vaultDeleteCmd} {
		c.ValidArgsFunction = completeSecretKeys
	}
}

func completeProfileNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	v, err := vault.Open(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = v.Close() }()

	profiles, err := v.ListProfiles()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names := make([]string, 0, len(profiles))
	for _, p := range profiles {
		names = append(names, p.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func completeSecretKeys(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	v, err := vault.Open(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer func() { _ = v.Close() }()

	profileName, _ := cmd.Flags().GetString("profile")

	secrets, err := v.List(profileName)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	keys := make([]string, 0, len(secrets))
	for _, s := range secrets {
		keys = append(keys, s.Key)
	}

	return keys, cobra.ShellCompDirectiveNoFileComp
}
