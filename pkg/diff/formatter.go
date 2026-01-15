package diff

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
)

// colorSprinter is an interface for color formatting
type colorSprinter interface {
	Sprint(a ...interface{}) string
	EnableColor()
	DisableColor()
}

const (
	// defaultIndent is the indentation level for diff values with paths
	defaultIndent = 4
	// rootIndent is the indentation level for root-level values without paths
	rootIndent = 0
	// continuationPrefix is the spacing prefix for continuation lines
	continuationPrefix = 2
)

// formatOptions holds all formatting configuration
type formatOptions struct {
	plain     bool // disables colored output
	pathsOnly bool // only shows paths, no values
	metadata  bool // includes additional metadata about values, it is mutually exclusive with pathsOnly
}

// FormatOption is a function that modifies format options
type FormatOption func(*formatOptions)

// Plain disables color output and formats values as plain text
func Plain(opts *formatOptions) {
	opts.plain = true
}

// PathsOnly suppresses the display of values
// It is mutually exclusive with WithMetadata
// If this option is set, WithMetadata will be ignored.
func PathsOnly(opts *formatOptions) {
	opts.pathsOnly = true
}

// WithMetadata includes additional metadata in the output
// It is mutually exclusive with PathsOnly
// If PathsOnly is set, this option will be ignored.
func WithMetadata(opts *formatOptions) {
	opts.metadata = true
}

// formatter handles the formatting of diffs
type formatter struct {
	options       formatOptions
	colorAdded    colorSprinter
	colorDeleted  colorSprinter
	colorModified colorSprinter
	colorMetadata colorSprinter
}

// newFormatter creates a new formatter with the given options
func newFormatter(opts ...FormatOption) *formatter {
	var options formatOptions
	for _, opt := range opts {
		opt(&options)
	}
	f := &formatter{
		options:       options,
		colorAdded:    newColorAdded(),
		colorDeleted:  newColorDeleted(),
		colorModified: newColorModified(),
		colorMetadata: newColorMetadata(),
	}

	if !options.plain {
		f.colorAdded.EnableColor()
		f.colorDeleted.EnableColor()
		f.colorModified.EnableColor()
		f.colorMetadata.EnableColor()
	} else {
		f.colorAdded.DisableColor()
		f.colorDeleted.DisableColor()
		f.colorModified.DisableColor()
		f.colorMetadata.DisableColor()
	}
	return f
}

// FormatDiff formats a single diff
func (f *formatter) formatDiff(diff *Diff) string {
	switch diff.Type() {
	case Added:
		return f.formatAdded(diff)
	case Deleted:
		return f.formatDeleted(diff)
	case Modified:
		return f.formatModified(diff)
	}
	return ""
}

// formatDocDiffs formats a collection of document diffs
func (f *formatter) formatDocDiffs(diffs DocDiffs) string {
	diffStrings := make([]string, 0, len(diffs))
	for _, diff := range diffs {
		diffStrings = append(diffStrings, f.formatDiff(diff))
	}
	return strings.Join(diffStrings, "\n")
}

// formatFileDiffs formats a collection of file diffs
func (f *formatter) formatFileDiffs(fileDiffs FileDiffs) string {
	docDiffStrings := make([]string, 0, len(fileDiffs))
	for _, docDiffs := range fileDiffs {
		docDiffStrings = append(docDiffStrings, f.formatDocDiffs(docDiffs))
	}
	return strings.Join(docDiffStrings, "\n---\n")
}

// formatValueYaml formats a value for YAML, applying color if not in plain mode
func (f *formatter) formatValueYaml(value string) string {
	if f.options.plain {
		return value
	}
	return colorize(value)
}

// formatAdded formats an added diff
func (f *formatter) formatAdded(diff *Diff) string {
	sign := f.colorAdded.Sprint("+")

	path := getNodePath(diff.rightNode)
	if path != "" {
		path = f.colorAdded.Sprint(path)
	}

	value, _ := f.getNodeValue(diff.rightNode, path)
	value = f.formatValueYaml(value)

	metadata := getNodeMetadata(diff.rightNode)
	metadata = f.colorMetadata.Sprint(metadata)

	return f.buildOutput(sign, path, value, metadata)
}

// formatDeleted formats a deleted diff
func (f *formatter) formatDeleted(diff *Diff) string {
	sign := f.colorDeleted.Sprint("-")

	path := getNodePath(diff.leftNode)
	if path != "" {
		path = f.colorDeleted.Sprint(path)
	}

	value, _ := f.getNodeValue(diff.leftNode, path)
	value = f.formatValueYaml(value)

	metadata := getNodeMetadata(diff.leftNode)
	metadata = f.colorMetadata.Sprint(metadata)

	return f.buildOutput(sign, path, value, metadata)
}

// formatModified formats a modified diff
func (f *formatter) formatModified(diff *Diff) string {
	sign := f.colorModified.Sprint("~")

	path := getNodePath(diff.leftNode)
	if path != "" {
		path = f.colorModified.Sprint(path)
	}

	leftValue, leftMultiLine := f.getNodeValue(diff.leftNode, path)
	leftValue = f.formatValueYaml(leftValue)

	rightValue, rightMultiLine := f.getNodeValue(diff.rightNode, path)
	rightValue = f.formatValueYaml(rightValue)

	leftMetadata := getNodeMetadata(diff.leftNode)
	leftMetadata = f.colorMetadata.Sprint(leftMetadata)

	rightMetadata := getNodeMetadata(diff.rightNode)
	rightMetadata = f.colorMetadata.Sprint(rightMetadata)

	symbol := f.getModifiedSymbol(leftMultiLine, rightMultiLine, path, &leftValue, &rightValue)

	return f.buildModifiedOutput(sign, path, leftValue, rightValue, symbol, leftMetadata, rightMetadata)
}

// getNodeValue extracts and formats the value of a node
func (f *formatter) getNodeValue(node ast.Node, path string) (string, bool) {
	indent := defaultIndent
	newLine := true
	if path == "" {
		indent = rootIndent
		newLine = false
	}
	return formatNodeValue(node, indent, newLine)
}

// getModifiedSymbol determines the symbol to use for modified diffs and adjusts values
func (f *formatter) getModifiedSymbol(leftMultiLine, rightMultiLine bool, path string, leftValue, rightValue *string) string {
	symbol := "→"
	if !leftMultiLine && !rightMultiLine {
		return symbol
	}

	// For multiline diffs, use vertical arrow symbol with indentation
	defaultIndentSpaces := strings.Repeat(" ", defaultIndent)
	continuationSpaces := strings.Repeat(" ", continuationPrefix)
	symbol = "\n" + defaultIndentSpaces + "↓"

	// Determine indentation: use default indent for paths, continuation for root
	indent := continuationSpaces
	if path != "" {
		indent = defaultIndentSpaces
	}

	// Add newline and indentation to the single-line value
	// Special case: single-line left value with a path gets formatted on the left
	if !leftMultiLine && path != "" {
		*leftValue = fmt.Sprintf("\n%s%s", indent, *leftValue)
	} else {
		// All other cases: format the right value
		// This includes: single-line right, or single-line left without path
		if !rightMultiLine || (path == "") {
			*rightValue = fmt.Sprintf("\n%s%s", indent, *rightValue)
		}
	}

	return symbol
}

// buildOutput builds the output string for single-value diffs
func (f *formatter) buildOutput(sign, path, value, metadata string) string {
	if f.options.pathsOnly {
		return fmt.Sprintf("%s %s", sign, path)
	}

	if path == "" {
		if f.options.metadata {
			return fmt.Sprintf("%s %s %s", sign, metadata, value)
		}
		return fmt.Sprintf("%s %s", sign, value)
	}

	if f.options.metadata {
		return fmt.Sprintf("%s %s: %s %s", sign, path, metadata, value)
	}
	return fmt.Sprintf("%s %s: %s", sign, path, value)
}

// buildModifiedOutput builds the output string for modified diffs
func (f *formatter) buildModifiedOutput(sign, path, leftValue, rightValue, symbol, leftMetadata, rightMetadata string) string {
	if f.options.pathsOnly {
		return fmt.Sprintf("%s %s", sign, path)
	}

	if path == "" {
		if f.options.metadata {
			return fmt.Sprintf("%s %s %s %s %s %s", sign, leftMetadata, leftValue, symbol, rightMetadata, rightValue)
		}
		return fmt.Sprintf("%s %s %s %s", sign, leftValue, symbol, rightValue)
	}

	if f.options.metadata {
		return fmt.Sprintf("%s %s: %s %s %s %s %s", sign, path, leftMetadata, leftValue, symbol, rightMetadata, rightValue)
	}
	return fmt.Sprintf("%s %s: %s %s %s", sign, path, leftValue, symbol, rightValue)
}

// getNodePath extracts the YAML path from a node
func getNodePath(node ast.Node) string {
	if node == nil {
		return ""
	}
	path := node.GetPath()
	if len(path) > 0 && path[0] == '$' {
		return path[1:] // remove leading dollar sign
	}
	return path
}

// formatNodeValue formats a node's value with proper indentation
func formatNodeValue(node ast.Node, indent int, startNewLine bool) (string, bool) {
	if node == nil {
		return "", false
	}

	s := node.String()
	lines := strings.Split(s, "\n")

	// Single line values (except collection types) don't need special formatting
	if len(lines) == 1 && node.Type() != ast.MappingType && node.Type() != ast.SequenceType {
		return s, false
	}

	// Calculate base indentation from the first line
	baseIndent := calculateIndentLevel(lines[0])

	// Reformat each line with proper indentation
	for i, line := range lines {
		if i == 0 {
			// First line: apply base indentation
			lines[i] = strings.Repeat(" ", indent) + strings.TrimPrefix(line, strings.Repeat(" ", baseIndent))
		} else {
			// Subsequent lines: handle different newline scenarios
			prefix := strings.Repeat(" ", continuationPrefix) // Default continuation prefix
			if startNewLine {
				prefix = "" // No extra prefix when starting with newline
			}
			lines[i] = prefix + strings.Repeat(" ", indent) + strings.TrimPrefix(line, strings.Repeat(" ", baseIndent))
		}
	}

	// Build the final formatted string
	var result strings.Builder
	if startNewLine {
		result.WriteString("\n")
	}
	result.WriteString(strings.Join(lines, "\n"))

	return result.String(), true
}

// getNodeMetadata returns metadata information for a node
func getNodeMetadata(node ast.Node) string {
	if node == nil {
		return ""
	}
	return fmt.Sprintf("[line:%d <%s>]", node.GetToken().Position.Line, node.Type())
}

// calculateIndentLevel calculates the indentation level of a string
func calculateIndentLevel(s string) int {
	for i, char := range s {
		if char != ' ' {
			return i
		}
	}
	return 0
}
