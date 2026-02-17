package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/internal/tui"
	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "anvil",
	Short:         "Machine-bound encrypted vault and secret manager",
	Long:          "Anvil — machine-bound encrypted vault for managing secrets organized by profiles.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if envInline, _ := cmd.Flags().GetString("env-inline"); envInline != "" {
			return envInlineHandler(envInline, cmd)
		}

		// If running in a terminal, launch TUI; otherwise show help
		if term.IsTerminal(int(os.Stdin.Fd())) {
			return launchTUI()
		}

		return cmd.Help()
	},
}

func launchTUI() error {
	v, err := vault.Open(nil)
	if err != nil {
		return err
	}

	defer func() { _ = v.Close() }()

	m := tui.NewModel(v)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
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

// GetRootCmd returns the root cobra command for testing.
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		handleError(rootCmd, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.Flags().String("env-inline", "", "Get a single secret value inline (requires active release)")
	rootCmd.Flags().StringP("profile", "p", "", "Target profile for --env-inline")
}
