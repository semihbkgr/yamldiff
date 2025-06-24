package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/semihbkgr/yamldiff/pkg/diff"
	"github.com/spf13/cobra"
)

// config holds all CLI configuration options
type config struct {
	exitOnDifference bool
	ignoreSeqOrder   bool
	color            bool
	pathsOnly        bool
	metadata         bool
	counts           bool
}

// newConfig creates a new config with default values
func newConfig() *config {
	return &config{
		exitOnDifference: false,
		ignoreSeqOrder:   false,
		color:            true, // color is enabled by default
		pathsOnly:        false,
		metadata:         false,
		counts:           false,
	}
}

// compareOptions converts config to diff.CompareOption slice
func (c *config) compareOptions() []diff.CompareOption {
	var opts []diff.CompareOption
	if c.ignoreSeqOrder {
		opts = append(opts, diff.IgnoreSeqOrder)
	}
	return opts
}

// formatOptions converts config to diff.FormatOption slice
func (c *config) formatOptions() []diff.FormatOption {
	var opts []diff.FormatOption
	if !c.color {
		opts = append(opts, diff.Plain)
	}
	if c.pathsOnly {
		opts = append(opts, diff.PathsOnly)
	}
	if c.metadata {
		opts = append(opts, diff.IncludeMetadata)
	}
	if c.counts {
		opts = append(opts, diff.IncludeCounts)
	}
	return opts
}

// newRootCommand creates the root cobra command
func newRootCommand() *cobra.Command {
	cfg := newConfig()

	cmd := &cobra.Command{
		Use:                   "yamldiff [flags] <file-left> <file-right>",
		Short:                 "structural comparison on two yaml files",
		Args:                  cobra.ExactArgs(2),
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
		Version:               version(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(cmd, args, cfg)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&cfg.exitOnDifference, "exit-code", "e", cfg.exitOnDifference, "Exit with non-zero status code when differences are found")
	cmd.Flags().BoolVarP(&cfg.ignoreSeqOrder, "ignore-order", "i", cfg.ignoreSeqOrder, "Ignore sequence order when comparing")
	cmd.Flags().BoolVar(&cfg.color, "color", cfg.color, "Enable color on output")
	cmd.Flags().BoolVarP(&cfg.pathsOnly, "paths-only", "p", cfg.pathsOnly, "Show only paths of differences without displaying the values")
	cmd.Flags().BoolVarP(&cfg.metadata, "metadata", "m", cfg.metadata, "Include additional metadata such as line numbers and node types in the output")
	cmd.Flags().BoolVarP(&cfg.counts, "counts", "c", cfg.counts, "Display a summary count of total added, deleted, and modified items")

	return cmd
}

// runCommand executes the main comparison logic
func runCommand(cmd *cobra.Command, args []string, cfg *config) error {
	diffs, err := diff.CompareFile(args[0], args[1], cfg.compareOptions()...)
	if err != nil {
		return fmt.Errorf("failed to compare files: %w", err)
	}

	output := diffs.Format(cfg.formatOptions()...)
	if output != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", output)
	}

	if cfg.exitOnDifference && diffs.HasDiff() {
		return errors.New("differences found between yaml files")
	}

	return nil
}

// Execute runs the root command
func Execute() {
	cmd := newRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
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
