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
	SilenceUsage: false,
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
var enableComments = false
var diffOptions = diff.DefaultDiffOptions
var outputOptions = diff.DefaultOutputOptions

func run(cmd *cobra.Command, args []string) error {
	diffCtx, err := diff.NewDiffContext(args[0], args[1], enableComments)
	if err != nil {
		return err
	}

	diffs := diffCtx.Diffs(diffOptions)
	fmt.Fprintf(cmd.OutOrStdout(), "%s", diffs.OutputString(outputOptions))

	if exitOnDifference && diffs.HasDifference() {
		return errors.New("yaml files have difference(s)")
	}

	return nil
}

func init() {
	rootCmd.Flags().BoolVarP(&exitOnDifference, "exit", "e", false, "returns non-zero exit status if there is a difference between yaml files")
	rootCmd.Flags().BoolVarP(&diffOptions.IgnoreIndex, "ignore", "i", diffOptions.IgnoreIndex, "ignore indexes in array")
	rootCmd.Flags().BoolVarP(&outputOptions.Plain, "plain", "p", outputOptions.Plain, "uncolored output")
	rootCmd.Flags().BoolVarP(&outputOptions.Silent, "silent", "s", outputOptions.Silent, "print output in silent ignoring values")
	rootCmd.Flags().BoolVarP(&outputOptions.Metadata, "metadata", "m", outputOptions.Metadata, "include metadata in the output (not work with silent flag)")
	rootCmd.Flags().BoolVarP(&enableComments, "comment", "c", enableComments, "display comments in the output")
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
