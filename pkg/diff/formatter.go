package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
)

// formatOptions holds all formatting configuration
type formatOptions struct {
	plain     bool // disables colored output
	pathsOnly bool // suppresses the display of values
	metadata  bool // includes additional metadata
	counts    bool // includes diff count summary
}

// FormatOption is a function that modifies format options
type FormatOption func(*formatOptions)

// Plain disables color output and formats values as plain text
func Plain(opts *formatOptions) {
	opts.plain = true
}

// PathsOnly suppresses the display of values
func PathsOnly(opts *formatOptions) {
	opts.pathsOnly = true
}

// WithMetadata includes additional metadata in the output
func WithMetadata(opts *formatOptions) {
	opts.metadata = true
}

// IncludeCounts includes diff count summary in the output
func IncludeCounts(opts *formatOptions) {
	opts.counts = true
}

// formatter handles the formatting of diffs
type formatter struct {
	options formatOptions
}

// newFormatter creates a new formatter with the given options
func newFormatter(opts ...FormatOption) *formatter {
	var options formatOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &formatter{options: options}
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

	result := strings.Join(diffStrings, "\n")

	if f.options.counts {
		count := f.countDiffs(diffs)
		result = count.String() + result
	}

	return result
}

// formatFileDiffs formats a collection of file diffs
func (f *formatter) formatFileDiffs(fileDiffs FileDiffs) string {
	docDiffStrings := make([]string, 0, len(fileDiffs))
	for _, docDiffs := range fileDiffs {
		docDiffStrings = append(docDiffStrings, f.formatDocDiffs(docDiffs))
	}
	return strings.Join(docDiffStrings, "\n---\n")
}

// countDiffs counts the number of each type of diff
func (f *formatter) countDiffs(diffs DocDiffs) diffCount {
	var count diffCount
	for _, diff := range diffs {
		switch diff.Type() {
		case Added:
			count.added++
		case Deleted:
			count.deleted++
		case Modified:
			count.modified++
		}
	}
	return count
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
	sign := "+"
	path := getNodePath(diff.rightNode)
	value, _ := f.getNodeValue(diff.rightNode, path)
	metadata := getNodeMetadata(diff.rightNode)

	if !f.options.plain {
		sign = color.HiGreenString(sign)
		if path != "" {
			path = color.HiGreenString(path)
		}
		metadata = color.HiWhiteString(metadata)
	}
	value = f.formatValueYaml(value)

	return f.buildOutput(sign, path, value, metadata)
}

// formatDeleted formats a deleted diff
func (f *formatter) formatDeleted(diff *Diff) string {
	sign := "-"
	path := getNodePath(diff.leftNode)
	value, _ := f.getNodeValue(diff.leftNode, path)
	metadata := getNodeMetadata(diff.leftNode)

	if !f.options.plain {
		sign = color.HiRedString(sign)
		if path != "" {
			path = color.HiRedString(path)
		}
		metadata = color.HiWhiteString(metadata)
	}
	value = f.formatValueYaml(value)

	return f.buildOutput(sign, path, value, metadata)
}

// formatModified formats a modified diff
func (f *formatter) formatModified(diff *Diff) string {
	sign := "~"
	path := getNodePath(diff.leftNode)

	leftValue, leftMultiLine := f.getNodeValue(diff.leftNode, path)
	rightValue, rightMultiLine := f.getNodeValue(diff.rightNode, path)
	leftMetadata := getNodeMetadata(diff.leftNode)
	rightMetadata := getNodeMetadata(diff.rightNode)

	symbol := f.getModifiedSymbol(leftMultiLine, rightMultiLine, path, &leftValue, &rightValue)

	if !f.options.plain {
		sign = color.HiYellowString(sign)
		if path != "" {
			path = color.HiYellowString(path)
		}
		leftMetadata = color.HiWhiteString(leftMetadata)
		rightMetadata = color.HiWhiteString(rightMetadata)
	}
	leftValue = f.formatValueYaml(leftValue)
	rightValue = f.formatValueYaml(rightValue)

	return f.buildModifiedOutput(sign, path, leftValue, rightValue, symbol, leftMetadata, rightMetadata)
}

// getNodeValue extracts and formats the value of a node
func (f *formatter) getNodeValue(node ast.Node, path string) (string, bool) {
	indent := 4
	newLine := true
	if path == "" {
		indent = 0
		newLine = false
	}
	return formatNodeValue(node, indent, newLine)
}

// getModifiedSymbol determines the symbol to use for modified diffs and adjusts values
func (f *formatter) getModifiedSymbol(leftMultiLine, rightMultiLine bool, path string, leftValue, rightValue *string) string {
	symbol := "→"
	multiline := leftMultiLine || rightMultiLine

	if multiline {
		symbol = "\n    ↓"

		if !leftMultiLine && path != "" {
			*leftValue = fmt.Sprintf("\n    %s", *leftValue)
		} else if !leftMultiLine && path == "" {
			*rightValue = fmt.Sprintf("\n  %s", *rightValue)
		} else if !rightMultiLine && path != "" {
			*rightValue = fmt.Sprintf("\n    %s", *rightValue)
		} else if !rightMultiLine && path == "" {
			*rightValue = fmt.Sprintf("\n  %s", *rightValue)
		} else if path == "" {
			*rightValue = fmt.Sprintf("\n  %s", *rightValue)
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

// diffCount tracks the count of different types of diffs
type diffCount struct {
	added    int
	deleted  int
	modified int
}

// String returns a string representation of the diff count
func (dc *diffCount) String() string {
	return fmt.Sprintf("%d added, %d deleted, %d modified\n", dc.added, dc.deleted, dc.modified)
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
			prefix := "  " // Default 2-space prefix for continuation
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

// colorPrinter is a global printer instance configured with syntax highlighting
var colorPrinter printer.Printer = initializeColorPrinter()

// initializeColorPrinter creates a printer with syntax highlighting
func initializeColorPrinter() printer.Printer {
	p := printer.Printer{}

	// Configure syntax highlighting colors
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiMagenta),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiMagenta),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiCyan),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.Anchor = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiYellow),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.Alias = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiYellow),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiGreen),
			Suffix: formatColorCode(color.Reset),
		}
	}

	return p
}

// formatColorCode formats a color attribute into an ANSI escape sequence
func formatColorCode(attr color.Attribute) string {
	return fmt.Sprintf("\x1b[%dm", attr)
}

// colorize applies YAML syntax highlighting to a string
func colorize(s string) string {
	tokens := lexer.Tokenize(s)
	return colorPrinter.PrintTokens(tokens)
}
