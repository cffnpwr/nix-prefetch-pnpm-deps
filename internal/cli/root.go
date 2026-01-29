package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "nix-prefetch-pnpm-deps",
	Short: "prefetch dependencies for pnpm",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	fetcherVersionFlag.Register(rootCmd)
	pnpmPathFlag.Register(rootCmd)
	workspaceFlag.Register(rootCmd)
	pnpmFlagFlag.Register(rootCmd)
	hashFlag.Register(rootCmd)
	quietFlag.Register(rootCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
