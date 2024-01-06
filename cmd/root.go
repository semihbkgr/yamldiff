package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/semihbkgr/yamldiff/diff"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "yamldiff",
	Short:        "yaml diff",
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE:         run,
	Version:      version(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var exitOnDifference = false
var plainOutput = false
var diffConfig = diff.DefaultDiffConfig

func run(cmd *cobra.Command, args []string) error {
	diffCtx, err := diff.NewDiffContext(args[0], args[1])
	if err != nil {
		return err
	}

	diffs := diffCtx.Diffs(diffConfig)
	fmt.Fprintf(cmd.OutOrStdout(), "%s", diffs.OutputString(!plainOutput))

	if exitOnDifference && diffs.HasDifference() {
		return errors.New("yaml files have difference(s)")
	}

	return nil
}

func init() {
	rootCmd.Flags().BoolVarP(&exitOnDifference, "exit", "e", false, "returns non-zero exit status if there is a difference between yaml files")
	rootCmd.Flags().BoolVarP(&diffConfig.IgnoreIndex, "ignore", "i", diffConfig.IgnoreIndex, "ignore indexes in array")
	rootCmd.Flags().BoolVarP(&plainOutput, "plain", "p", plainOutput, "uncolored output")
}

// buildVersion is set by ldflags
var buildVersion string

func version() string {
	if buildVersion != "" {
		return buildVersion
	}

	i, ok := debug.ReadBuildInfo()
	if ok {
		return i.Main.Version
	}

	return "undefined"
}
