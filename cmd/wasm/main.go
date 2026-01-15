//go:build js && wasm

package main

import (
	"runtime/debug"
	"syscall/js"

	"github.com/semihbkgr/yamldiff/pkg/diff"
)

func main() {
	// Expose functions to JavaScript
	js.Global().Set("yamldiffCompare", js.FuncOf(compare))
	js.Global().Set("yamldiffVersion", js.FuncOf(version))
	js.Global().Set("yamldiffStat", js.FuncOf(stat))

	// Keep the Go program running
	select {}
}

// compare compares two YAML strings and returns the diff
// JavaScript signature: yamldiffCompare(left: string, right: string, options?: {ignoreOrder?: boolean, pathOnly?: boolean, metadata?: boolean}) => {result?: string, error?: string}
func compare(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return map[string]any{"error": "yamldiffCompare requires at least 2 arguments: left, right"}
	}

	left := args[0].String()
	right := args[1].String()

	// Parse options
	var ignoreOrder, pathOnly, metadata bool
	if len(args) >= 3 && !args[2].IsNull() && !args[2].IsUndefined() {
		opts := args[2]
		if v := opts.Get("ignoreOrder"); !v.IsUndefined() {
			ignoreOrder = v.Bool()
		}
		if v := opts.Get("pathOnly"); !v.IsUndefined() {
			pathOnly = v.Bool()
		}
		if v := opts.Get("metadata"); !v.IsUndefined() {
			metadata = v.Bool()
		}
	}

	// Build compare options
	var compareOpts []diff.CompareOption
	if ignoreOrder {
		compareOpts = append(compareOpts, diff.IgnoreSeqOrder)
	}

	// Compare
	diffs, err := diff.Compare([]byte(left), []byte(right), compareOpts...)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	// Build format options (WASM uses HTML colorization)
	var formatOpts []diff.FormatOption
	if pathOnly {
		formatOpts = append(formatOpts, diff.PathsOnly)
	} else if metadata {
		formatOpts = append(formatOpts, diff.WithMetadata)
	}

	result := diffs.Format(formatOpts...)
	return map[string]any{"result": result, "hasDiff": diffs.HasDiff()}
}

// stat returns diff statistics
// JavaScript signature: yamldiffStat(left: string, right: string, options?: {ignoreOrder?: boolean}) => {result?: {added: number, deleted: number, modified: number}, error?: string}
func stat(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return map[string]any{"error": "yamldiffStat requires at least 2 arguments: left, right"}
	}

	left := args[0].String()
	right := args[1].String()

	// Parse options
	var ignoreOrder bool
	if len(args) >= 3 && !args[2].IsNull() && !args[2].IsUndefined() {
		opts := args[2]
		if v := opts.Get("ignoreOrder"); !v.IsUndefined() {
			ignoreOrder = v.Bool()
		}
	}

	// Build compare options
	var compareOpts []diff.CompareOption
	if ignoreOrder {
		compareOpts = append(compareOpts, diff.IgnoreSeqOrder)
	}

	// Compare
	diffs, err := diff.Compare([]byte(left), []byte(right), compareOpts...)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	// Count stats
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

	return map[string]any{
		"result": map[string]any{
			"added":    added,
			"deleted":  deleted,
			"modified": modified,
		},
		"hasDiff": diffs.HasDiff(),
	}
}

// version returns the version of yamldiff
// JavaScript signature: yamldiffVersion() => string
func version(this js.Value, args []js.Value) any {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}
