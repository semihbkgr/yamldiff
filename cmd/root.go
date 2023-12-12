package cmd

import (
	"fmt"
	"os"

	"github.com/semihbkgr/yamldiff/diff"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "yamldiff",
	Short:        "yaml diff",
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE:         run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	return nil
}
