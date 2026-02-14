package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker secrets integration",
}

var vaultDockerExportCmd = &cobra.Command{
	Use:   "export <directory>",
	Short: "Export secrets as Docker-compatible secret files",
	Long:  "Writes each secret as an individual file in the target directory, mimicking Docker's /run/secrets/ layout. Each file is named after the secret key and contains the plaintext value.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		entries, err := v.Export(profileName)
		if err != nil {
			return err
		}

		dir := args[0]
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}

		for _, e := range entries {
			secretPath := filepath.Join(dir, e.Key)
			if err := os.WriteFile(secretPath, []byte(e.Value), 0o600); err != nil {
				return fmt.Errorf("write secret %q: %w", e.Key, err)
			}
		}

		outputResult(cmd, struct {
			Directory string `json:"directory"`
			Count     int    `json:"count"`
			Message   string `json:"message"`
		}{
			Directory: dir,
			Count:     len(entries),
			Message:   fmt.Sprintf("Exported %d secrets to %s", len(entries), dir),
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d secrets to %s\n", len(entries), dir)
		})

		return nil
	},
}

var vaultDockerCleanCmd = &cobra.Command{
	Use:   "clean <directory>",
	Short: "Remove exported Docker secret files",
	Long:  "Removes all secret files from the target directory that match vault secret keys.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		secrets, err := v.List(profileName)
		if err != nil {
			return err
		}

		dir := args[0]
		var removed int

		for _, s := range secrets {
			secretPath := filepath.Join(dir, s.Key)
			if err := os.Remove(secretPath); err == nil {
				removed++
			}
		}

		outputResult(cmd, struct {
			Directory string `json:"directory"`
			Removed   int    `json:"removed"`
			Message   string `json:"message"`
		}{
			Directory: dir,
			Removed:   removed,
			Message:   fmt.Sprintf("Removed %d secret files from %s", removed, dir),
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %d secret files from %s\n", removed, dir)
		})

		return nil
	},
}

var vaultDockerComposeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Generate Docker Compose secrets snippet",
	Long:  "Outputs a Docker Compose YAML snippet mapping vault secrets as external file-based secrets.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		secretsDir, _ := cmd.Flags().GetString("dir")

		secrets, err := v.List(profileName)
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintln(w, "secrets:")

		for _, s := range secrets {
			_, _ = fmt.Fprintf(w, "  %s:\n", s.Key)
			_, _ = fmt.Fprintf(w, "    file: %s\n", filepath.Join(secretsDir, s.Key))
		}

		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultDockerCmd)

	vaultDockerExportCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultDockerCmd.AddCommand(vaultDockerExportCmd)

	vaultDockerCleanCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultDockerCmd.AddCommand(vaultDockerCleanCmd)

	vaultDockerComposeCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultDockerComposeCmd.Flags().String("dir", "./secrets", "Directory path for secret files")
	vaultDockerCmd.AddCommand(vaultDockerComposeCmd)
}
