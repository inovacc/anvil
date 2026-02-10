package cmd

import (
	"fmt"
	"os"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "profile",
	Short: "Machine-bound encrypted vault and profile manager",
	Long:  "CLI tool for managing encrypted secrets organized by profiles, bound to this machine.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if envInline, _ := cmd.Flags().GetString("env-inline"); envInline != "" {
			return envInlineHandler(envInline, cmd)
		}

		return cmd.Help()
	},
}

func envInlineHandler(key string, cmd *cobra.Command) error {
	v, err := vault.Open(nil)
	if err != nil {
		return err
	}

	defer func() { _ = v.Close() }()

	profileName, _ := cmd.Flags().GetString("profile")

	value, err := v.EnvInlineGet(key, profileName)
	if err != nil {
		return err
	}

	outputResult(cmd, struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{Key: key, Value: value}, func() {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), value)
	})

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.Flags().String("env-inline", "", "Get a single secret value inline (requires active release)")
	rootCmd.Flags().StringP("profile", "p", "", "Target profile for --env-inline")
}
