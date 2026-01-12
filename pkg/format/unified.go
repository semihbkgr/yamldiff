package format

import (
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
	"github.com/semihbkgr/yamldiff/pkg/diff"
)

// UnifiedOption is a function that modifies unified format options
type UnifiedOption func(*unifiedOptions)

// unifiedOptions holds all unified formatting configuration
type unifiedOptions struct {
	plain bool // disables colored output, uses +/- prefixes instead
}

// UnifiedPlain disables color output and uses +/- prefixes for distinction
func UnifiedPlain(opts *unifiedOptions) {
	opts.plain = true
}

// lineStatus represents whether a line is unchanged, added, or deleted
type lineStatus int

const (
	statusUnchanged lineStatus = iota
	statusAdded
	statusDeleted
)

// unifiedFormatter handles the formatting of unified diffs
type unifiedFormatter struct {
	options      unifiedOptions
	colorAdded   *color.Color
	colorDeleted *color.Color
}

// newUnifiedFormatter creates a new unified formatter with the given options
func newUnifiedFormatter(opts ...UnifiedOption) *unifiedFormatter {
	var options unifiedOptions
	for _, opt := range opts {
		opt(&options)
	}

	f := &unifiedFormatter{
		options:      options,
		colorAdded:   color.New(color.BgHiGreen, color.FgBlack, color.Bold),
		colorDeleted: color.New(color.BgHiRed, color.FgWhite, color.Bold),
	}

	if !options.plain {
		f.colorAdded.EnableColor()
		f.colorDeleted.EnableColor()
	} else {
		f.colorAdded.DisableColor()
		f.colorDeleted.DisableColor()
	}

	return f
}

// Unified formats the comparison result as a unified diff
func Unified(result *diff.CompareResult, opts ...UnifiedOption) string {
	f := newUnifiedFormatter(opts...)
	return f.format(result)
}

// format formats the entire comparison result
func (f *unifiedFormatter) format(result *diff.CompareResult) string {
	var output strings.Builder

	numDocs := max(len(result.LeftAST.Docs), len(result.RightAST.Docs))

	for i := 0; i < numDocs; i++ {
		if i > 0 {
			output.WriteString("---\n")
		}

		var leftDoc, rightDoc ast.Node
		if i < len(result.LeftAST.Docs) {
			leftDoc = result.LeftAST.Docs[i].Body
		}
		if i < len(result.RightAST.Docs) {
			rightDoc = result.RightAST.Docs[i].Body
		}

		var docDiffs diff.DocDiffs
		if i < len(result.Diffs) {
			docDiffs = result.Diffs[i]
		}

		docOutput := f.formatDocument(leftDoc, rightDoc, docDiffs)
		output.WriteString(docOutput)
	}

	return strings.TrimSuffix(output.String(), "\n")
}

// formatDocument formats a single document with unified diff style
func (f *unifiedFormatter) formatDocument(leftDoc, rightDoc ast.Node, diffs diff.DocDiffs) string {
	// Get all lines from both documents
	var leftLines, rightLines []string
	if leftDoc != nil {
		leftLines = strings.Split(leftDoc.String(), "\n")
	}
	if rightDoc != nil {
		rightLines = strings.Split(rightDoc.String(), "\n")
	}

	// Build status arrays for each line based on diffs
	leftStatus := make([]lineStatus, len(leftLines))
	rightStatus := make([]lineStatus, len(rightLines))

	// Mark deleted and modified lines in left
	for _, d := range diffs {
		if d.Type() == diff.Deleted || d.Type() == diff.Modified {
			node := d.LeftNode()
			if node != nil {
				markNodeLines(node, leftLines, leftStatus, statusDeleted)
			}
		}
	}

	// Mark added and modified lines in right
	for _, d := range diffs {
		if d.Type() == diff.Added || d.Type() == diff.Modified {
			node := d.RightNode()
			if node != nil {
				markNodeLines(node, rightLines, rightStatus, statusAdded)
			}
		}
	}

	// Use LCS-based diff to properly align the documents
	return f.diffLines(leftLines, leftStatus, rightLines, rightStatus)
}

// markNodeLines marks the lines belonging to a node with the given status
func markNodeLines(node ast.Node, lines []string, status []lineStatus, mark lineStatus) {
	startLine := node.GetToken().Position.Line - 1
	nodeStr := node.String()
	nodeLines := strings.Split(nodeStr, "\n")

	// Check if the line before startLine is a key line (ends with ":")
	// This handles cases like "volumeMounts:" where the key is separate from the value
	if startLine > 0 && len(nodeLines) > 1 {
		prevLine := strings.TrimSpace(lines[startLine-1])
		if strings.HasSuffix(prevLine, ":") {
			status[startLine-1] = mark
		}
	}

	// Mark the node content lines
	for j := 0; j < len(nodeLines) && startLine+j < len(status); j++ {
		status[startLine+j] = mark
	}
}

// diffLines performs a proper diff between left and right lines
func (f *unifiedFormatter) diffLines(leftLines []string, leftStatus []lineStatus, rightLines []string, rightStatus []lineStatus) string {
	var output strings.Builder

	// Find LCS (Longest Common Subsequence) of unchanged lines
	// This helps us properly align the documents
	lcs := findLCS(leftLines, leftStatus, rightLines, rightStatus)

	li, ri := 0, 0
	lcsIdx := 0

	for li < len(leftLines) || ri < len(rightLines) {
		// Check if we've reached an LCS match point
		if lcsIdx < len(lcs) {
			leftIdx, rightIdx := lcs[lcsIdx][0], lcs[lcsIdx][1]

			// Output all deleted lines before the match point
			for li < leftIdx {
				if leftStatus[li] == statusDeleted {
					f.writeLine(&output, leftLines[li], statusDeleted)
				}
				li++
			}

			// Output all added lines before the match point
			for ri < rightIdx {
				if rightStatus[ri] == statusAdded {
					f.writeLine(&output, rightLines[ri], statusAdded)
				}
				ri++
			}

			// Output the matching unchanged line
			f.writeLine(&output, rightLines[ri], statusUnchanged)
			li++
			ri++
			lcsIdx++
		} else {
			// No more LCS matches, output remaining lines
			for li < len(leftLines) {
				if leftStatus[li] == statusDeleted {
					f.writeLine(&output, leftLines[li], statusDeleted)
				}
				li++
			}
			for ri < len(rightLines) {
				if rightStatus[ri] == statusAdded {
					f.writeLine(&output, rightLines[ri], statusAdded)
				}
				ri++
			}
		}
	}

	return output.String()
}

// findLCS finds the longest common subsequence of unchanged lines
// Returns a list of [leftIndex, rightIndex] pairs
func findLCS(leftLines []string, leftStatus []lineStatus, rightLines []string, rightStatus []lineStatus) [][2]int {
	// Build list of unchanged lines with their indices
	type indexedLine struct {
		index int
		line  string
	}

	var leftUnchanged, rightUnchanged []indexedLine
	for i, line := range leftLines {
		if leftStatus[i] == statusUnchanged {
			leftUnchanged = append(leftUnchanged, indexedLine{i, line})
		}
	}
	for i, line := range rightLines {
		if rightStatus[i] == statusUnchanged {
			rightUnchanged = append(rightUnchanged, indexedLine{i, line})
		}
	}

	// Find LCS using dynamic programming
	m, n := len(leftUnchanged), len(rightUnchanged)
	if m == 0 || n == 0 {
		return nil
	}

	// DP table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if leftUnchanged[i-1].line == rightUnchanged[j-1].line {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to find the LCS
	var result [][2]int
	i, j := m, n
	for i > 0 && j > 0 {
		if leftUnchanged[i-1].line == rightUnchanged[j-1].line {
			result = append([][2]int{{leftUnchanged[i-1].index, rightUnchanged[j-1].index}}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return result
}

// writeLine writes a single line with appropriate formatting
func (f *unifiedFormatter) writeLine(output *strings.Builder, line string, status lineStatus) {
	if f.options.plain {
		// Plain mode: use +/- prefixes without colors
		switch status {
		case statusAdded:
			output.WriteString("+ ")
		case statusDeleted:
			output.WriteString("- ")
		default:
			output.WriteString("  ")
		}
		output.WriteString(line)
	} else {
		// Color mode: use background colors with syntax highlighting
		coloredLine := colorizeWithStatus(line, int(status))
		// Reset all attributes at end of line
		output.WriteString(coloredLine)
		output.WriteString("\x1b[0m")
	}
	output.WriteString("\n")
}
