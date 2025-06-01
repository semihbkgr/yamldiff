package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/semihbkgr/yamldiff/pkg/diff"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                   "yamldiff [flags] <file-left> <file-right>",
	Short:                 "structural comparison on two yaml files",
	Args:                  cobra.ExactArgs(2),
	SilenceUsage:          false,
	DisableFlagsInUseLine: true,
	RunE:                  run,
	Version:               version(),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type options struct {
	exitOnDifference bool
	ignoreSeqOrder   bool
	formatPlain      bool
	formatSilent     bool
	formatMetadata   bool
}

func (o *options) compareOptions() []diff.CompareOption {
	var opts []diff.CompareOption
	if o.ignoreSeqOrder {
		opts = append(opts, diff.IgnoreSeqOrder)
	}
	return opts
}

func (o *options) formatOptions() []diff.FormatOption {
	var opts []diff.FormatOption
	if o.formatPlain {
		opts = append(opts, diff.Plain)
	}
	if o.formatSilent {
		opts = append(opts, diff.Silent)
	}
	if o.formatMetadata {
		opts = append(opts, diff.IncludeMetadata)
	}
	return opts
}

var opts = &options{
	exitOnDifference: false,
	ignoreSeqOrder:   false,
	formatPlain:      false,
	formatSilent:     false,
	formatMetadata:   false,
}

func run(cmd *cobra.Command, args []string) error {
	diffs, err := diff.CompareFile(args[0], args[1], opts.compareOptions()...)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", diffs.Format(opts.formatOptions()...))

	if opts.exitOnDifference && diffs.HasDiff() {
		return errors.New("differences found between yaml files")
	}

	return nil
}

func init() {
	rootCmd.Flags().BoolVarP(&opts.exitOnDifference, "exit", "e", opts.exitOnDifference, "Exit with a non-zero status code if differences are found between yaml files.")
	rootCmd.Flags().BoolVarP(&opts.ignoreSeqOrder, "unordered", "u", opts.ignoreSeqOrder, "Ignore the order of items in arrays during comparison.")
	rootCmd.Flags().BoolVarP(&opts.formatPlain, "plain", "p", opts.formatPlain, "Output without any color formatting.")
	rootCmd.Flags().BoolVarP(&opts.formatSilent, "silent", "s", opts.formatSilent, "Suppress output of values, showing only differences.")
	rootCmd.Flags().BoolVarP(&opts.formatMetadata, "metadata", "m", opts.formatMetadata, "Include additional metadata in the output (not applicable with the silent flag).")
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
