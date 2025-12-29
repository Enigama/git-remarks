package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-remarks",
	Short: "Personal developer notes attached to Git commits",
	Long: `git-remarks is a tool for managing personal developer notes attached to Git commits.

Notes are scoped to branches and survive rebases. They are stored using git notes
and remain local by default.

Examples:
  git remarks add "This is a test helper, remove before PR"
  git remarks list
  git remarks resolve a1b2c3d4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to list command when no subcommand is provided
		return listCmd.RunE(cmd, args)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(resolveCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(migrateBranchCmd)
	rootCmd.AddCommand(migrateRewritesCmd)
}

