package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/semihbkgr/yamldiff/pkg/diff"
	"github.com/spf13/cobra"
)

// config holds all CLI configuration options
type config struct {
	exitOnDifference bool
	ignoreSeqOrder   bool
	color            string
	pathOnly         bool
	metadata         bool
	stat             bool
	list             bool // use list format (original behavior) instead of unified
}

// newConfig creates a new config with default values
func newConfig() *config {
	return &config{
		exitOnDifference: false,
		ignoreSeqOrder:   false,
		color:            "auto",
		pathOnly:         false,
		metadata:         false,
		stat:             false,
		list:             false,
	}
}

// shouldUseColor determines if color output should be enabled
func (c *config) shouldUseColor() bool {
	switch c.color {
	case "always":
		return true
	case "never":
		return false
	case "auto":
		return !color.NoColor
	default:
		return false // fallback for invalid values
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
	if !c.shouldUseColor() {
		opts = append(opts, diff.Plain)
	}
	if c.pathOnly {
		opts = append(opts, diff.PathsOnly)
	}
	if c.metadata {
		opts = append(opts, diff.WithMetadata)
	}
	return opts
}

// validateColorFlag validates the color flag value
func (c *config) validateColorFlag() error {
	switch c.color {
	case "always", "never", "auto":
		return nil
	default:
		return fmt.Errorf("invalid color value %q: must be one of 'always', 'never', or 'auto'", c.color)
	}
}

// validateMutuallyExclusiveFlags checks for mutually exclusive flags
func (c *config) validateMutuallyExclusiveFlags() error {
	if c.pathOnly && c.metadata {
		return errors.New("flags --path-only and --metadata are mutually exclusive")
	}
	if c.stat && (c.pathOnly || c.metadata) {
		return errors.New("flag --stat cannot be used with --path-only or --metadata")
	}
	// --metadata and --path-only only work with --list format
	if !c.list && (c.metadata || c.pathOnly) {
		return errors.New("flags --metadata and --path-only require --list flag")
	}
	return nil
}

// unifiedOptions converts config to diff.UnifiedOption slice
func (c *config) unifiedOptions() []diff.UnifiedOption {
	var opts []diff.UnifiedOption
	if !c.shouldUseColor() {
		opts = append(opts, diff.UnifiedPlain)
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
	cmd.Flags().StringVar(&cfg.color, "color", cfg.color, "When to use color output. It can be one of always, never, or auto.")
	cmd.Flags().BoolVarP(&cfg.list, "list", "l", cfg.list, "Use list format showing paths and values (original behavior)")
	cmd.Flags().BoolVarP(&cfg.pathOnly, "path-only", "p", cfg.pathOnly, "Show only paths of differences without values (requires --list)")
	cmd.Flags().BoolVarP(&cfg.metadata, "metadata", "m", cfg.metadata, "Include line numbers and node types (requires --list)")
	cmd.Flags().BoolVarP(&cfg.stat, "stat", "s", cfg.stat, "Show only diffstat summary (added, deleted, modified counts)")

	return cmd
}

// runCommand executes the main comparison logic
func runCommand(cmd *cobra.Command, args []string, cfg *config) error {
	if err := cfg.validateColorFlag(); err != nil {
		return err
	}

	if err := cfg.validateMutuallyExclusiveFlags(); err != nil {
		return err
	}

	var output string
	var hasDiff bool

	if cfg.list || cfg.stat {
		// Use original comparison for list format or stat
		diffs, err := diff.CompareFile(args[0], args[1], cfg.compareOptions()...)
		if err != nil {
			return fmt.Errorf("failed to compare files: %w", err)
		}
		hasDiff = diffs.HasDiff()

		if cfg.stat {
			output = formatStat(diffs)
		} else {
			output = diffs.Format(cfg.formatOptions()...)
		}
	} else {
		// Use unified format (default)
		result, err := diff.CompareFileWithAST(args[0], args[1], cfg.compareOptions()...)
		if err != nil {
			return fmt.Errorf("failed to compare files: %w", err)
		}
		hasDiff = result.Diffs.HasDiff()

		if hasDiff {
			output = result.Diffs.FormatUnified(result.LeftAST, result.RightAST, cfg.unifiedOptions()...)
		}
	}

	if output != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", output)
	}

	if cfg.exitOnDifference && hasDiff {
		return errors.New("differences found between yaml files")
	}

	return nil
}

// formatStat returns a summary line with diff statistics
func formatStat(diffs diff.FileDiffs) string {
	var added, deleted, modified int
	for _, docDiffs := range diffs {
		for _, d := range docDiffs {
			switch d.Type() {
			case diff.Added:
				added++
			case diff.Deleted:
				deleted++
			case diff.Modified:
				modified++
			}
		}
	}
	return fmt.Sprintf("%d added, %d deleted, %d modified", added, deleted, modified)
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
