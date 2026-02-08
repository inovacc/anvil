package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type aiCommandInfo struct {
	Name        string       `json:"name"`
	FullPath    string       `json:"full_path"`
	Description string       `json:"description"`
	Usage       string       `json:"usage"`
	Flags       []aiFlagInfo `json:"flags,omitempty"`
	Category    string       `json:"category"`
}

type aiFlagInfo struct {
	Name         string `json:"name"`
	Shorthand    string `json:"shorthand,omitempty"`
	Description  string `json:"description"`
	DefaultValue string `json:"default,omitempty"`
}

var aicontextCmd = &cobra.Command{
	Use:   "aicontext",
	Short: "Generate AI-readable documentation",
	Long:  "Generates structured documentation about all CLI commands for AI consumption.",
	RunE: func(cmd *cobra.Command, args []string) error {
		compact, _ := cmd.Flags().GetBool("compact")
		category, _ := cmd.Flags().GetString("category")

		commands := collectCommands(rootCmd, "", "")
		if category != "" {
			commands = filterByCategory(commands, category)
		}

		outputResult(cmd, commands, func() {
			if compact {
				printCompactMarkdown(commands)
			} else {
				printFullMarkdown(commands)
			}
		})
		return nil
	},
}

func collectCommands(cmd *cobra.Command, parentPath, category string) []aiCommandInfo {
	fullPath := cmd.Name()
	if parentPath != "" {
		fullPath = parentPath + " " + cmd.Name()
	}

	// Determine category from top-level parent
	currentCategory := category
	if currentCategory == "" && parentPath != "" {
		currentCategory = cmd.Name()
	}
	if parentPath == "" {
		currentCategory = "root"
	}

	var result []aiCommandInfo

	// Only include leaf commands or commands with their own RunE
	if cmd.RunE != nil || cmd.Run != nil {
		info := aiCommandInfo{
			Name:        cmd.Name(),
			FullPath:    fullPath,
			Description: cmd.Short,
			Usage:       cmd.UseLine(),
			Category:    currentCategory,
		}

		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" || f.Name == "json" {
				return
			}
			fi := aiFlagInfo{
				Name:        "--" + f.Name,
				Description: f.Usage,
			}
			if f.Shorthand != "" {
				fi.Shorthand = "-" + f.Shorthand
			}
			if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
				fi.DefaultValue = f.DefValue
			}
			info.Flags = append(info.Flags, fi)
		})

		result = append(result, info)
	}

	for _, child := range visibleSubcommands(cmd) {
		result = append(result, collectCommands(child, fullPath, currentCategory)...)
	}

	return result
}

func filterByCategory(commands []aiCommandInfo, category string) []aiCommandInfo {
	cat := strings.ToLower(category)
	var filtered []aiCommandInfo
	for _, c := range commands {
		if strings.ToLower(c.Category) == cat {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func printFullMarkdown(commands []aiCommandInfo) {
	fmt.Println("# profile CLI")
	fmt.Println()
	fmt.Println("Machine-bound encrypted vault and profile manager.")
	fmt.Println()
	fmt.Println("## Commands")

	for _, c := range commands {
		fmt.Println()
		fmt.Printf("### %s\n", c.FullPath)
		fmt.Println()
		fmt.Println(c.Description)
		fmt.Println()
		fmt.Printf("**Usage:** `%s`\n", c.Usage)

		if len(c.Flags) > 0 {
			fmt.Println()
			fmt.Println("**Flags:**")
			for _, f := range c.Flags {
				flag := f.Name
				if f.Shorthand != "" {
					flag = f.Shorthand + ", " + f.Name
				}
				def := ""
				if f.DefaultValue != "" {
					def = fmt.Sprintf(" (default: %s)", f.DefaultValue)
				}
				fmt.Printf("- `%s` — %s%s\n", flag, f.Description, def)
			}
		}
	}
}

func printCompactMarkdown(commands []aiCommandInfo) {
	fmt.Println("# profile CLI")
	fmt.Println()
	for _, c := range commands {
		flags := ""
		if len(c.Flags) > 0 {
			var names []string
			for _, f := range c.Flags {
				names = append(names, f.Name)
			}
			flags = " [" + strings.Join(names, ", ") + "]"
		}
		fmt.Printf("- `%s` — %s%s\n", c.FullPath, c.Description, flags)
	}
}

func init() {
	aicontextCmd.Flags().Bool("compact", false, "Shorter markdown output")
	aicontextCmd.Flags().String("category", "", "Filter by command group (vault, env, profile)")
	rootCmd.AddCommand(aicontextCmd)
}
