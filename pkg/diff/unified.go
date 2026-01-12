package diff

import (
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
)

// lineStatus represents the status of a line in unified diff output
type lineStatus int

const (
	statusUnchanged lineStatus = iota
	statusAdded
	statusDeleted
)

// annotatedLine represents a single line with its diff status
type annotatedLine struct {
	status  lineStatus
	content string
}

// UnifiedOption is a function that modifies unified format options
type UnifiedOption func(*unifiedOptions)

type unifiedOptions struct {
	plain bool // disables colored output
}

// UnifiedPlain disables color output for unified format
func UnifiedPlain(opts *unifiedOptions) {
	opts.plain = true
}

// unifiedFormatter handles unified diff formatting
type unifiedFormatter struct {
	options      unifiedOptions
	colorAdded   *color.Color
	colorDeleted *color.Color
	colorContext *color.Color
}

// newUnifiedFormatter creates a new unified formatter
func newUnifiedFormatter(opts ...UnifiedOption) *unifiedFormatter {
	var options unifiedOptions
	for _, opt := range opts {
		opt(&options)
	}

	f := &unifiedFormatter{
		options:      options,
		colorAdded:   color.New(color.FgHiGreen),
		colorDeleted: color.New(color.FgHiRed),
		colorContext: color.New(color.Reset),
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

// FormatUnified formats diffs in unified diff style showing the full document
func FormatUnified(leftAST, rightAST *ast.File, diffs FileDiffs, opts ...UnifiedOption) string {
	f := newUnifiedFormatter(opts...)
	return f.format(leftAST, rightAST, diffs)
}

// format generates the unified diff output
func (f *unifiedFormatter) format(leftAST, rightAST *ast.File, diffs FileDiffs) string {
	var result strings.Builder

	numDocs := max(len(leftAST.Docs), len(rightAST.Docs))
	for i := 0; i < numDocs; i++ {
		if i > 0 {
			result.WriteString("---\n")
		}

		var leftDoc, rightDoc ast.Node
		if i < len(leftAST.Docs) {
			leftDoc = leftAST.Docs[i].Body
		}
		if i < len(rightAST.Docs) {
			rightDoc = rightAST.Docs[i].Body
		}

		var docDiffs DocDiffs
		if i < len(diffs) {
			docDiffs = diffs[i]
		}

		lines := f.buildUnifiedLines(leftDoc, rightDoc, docDiffs)
		for _, line := range lines {
			result.WriteString(f.formatLine(line))
			result.WriteString("\n")
		}
	}

	// Remove trailing newline to match existing format behavior
	output := result.String()
	if len(output) > 0 && output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}

	return output
}

// formatLine formats a single annotated line with appropriate prefix and color
func (f *unifiedFormatter) formatLine(line annotatedLine) string {
	switch line.status {
	case statusAdded:
		return f.colorAdded.Sprintf("+%s", line.content)
	case statusDeleted:
		return f.colorDeleted.Sprintf("-%s", line.content)
	default:
		return " " + line.content
	}
}

// buildUnifiedLines builds the unified diff output for a document
func (f *unifiedFormatter) buildUnifiedLines(leftDoc, rightDoc ast.Node, docDiffs DocDiffs) []annotatedLine {
	// Build a map of paths to diffs for quick lookup
	diffMap := make(map[string]*Diff)
	for _, d := range docDiffs {
		diffMap[d.Path()] = d
	}

	// If both docs are nil, return empty
	if leftDoc == nil && rightDoc == nil {
		return nil
	}

	// If only one side exists, show it as all added or all deleted
	if leftDoc == nil {
		return f.annotateNode(rightDoc, statusAdded, 0)
	}
	if rightDoc == nil {
		return f.annotateNode(leftDoc, statusDeleted, 0)
	}

	// Both documents exist - merge them with diff annotations
	return f.mergeNodes(leftDoc, rightDoc, diffMap, 0)
}

// mergeNodes merges two nodes and produces unified diff lines
func (f *unifiedFormatter) mergeNodes(leftNode, rightNode ast.Node, diffMap map[string]*Diff, indent int) []annotatedLine {
	// Check if this exact path has a diff
	path := getNodePath(rightNode)
	if path == "" {
		path = getNodePath(leftNode)
	}

	if diff, ok := diffMap[path]; ok {
		// This node has a diff
		switch diff.Type() {
		case Added:
			return f.annotateNode(rightNode, statusAdded, indent)
		case Deleted:
			return f.annotateNode(leftNode, statusDeleted, indent)
		case Modified:
			// Show old value as deleted, new value as added
			deleted := f.annotateNode(leftNode, statusDeleted, indent)
			added := f.annotateNode(rightNode, statusAdded, indent)
			return append(deleted, added...)
		}
	}

	// No diff at this level - check if nodes are structurally compatible
	if leftNode == nil {
		return f.annotateNode(rightNode, statusAdded, indent)
	}
	if rightNode == nil {
		return f.annotateNode(leftNode, statusDeleted, indent)
	}

	// Handle different node types
	switch leftNode.Type() {
	case ast.MappingType:
		if rightNode.Type() == ast.MappingType {
			return f.mergeMappingNodes(leftNode.(*ast.MappingNode), rightNode.(*ast.MappingNode), diffMap, indent)
		}
	case ast.SequenceType:
		if rightNode.Type() == ast.SequenceType {
			return f.mergeSequenceNodes(leftNode.(*ast.SequenceNode), rightNode.(*ast.SequenceNode), diffMap, indent)
		}
	}

	// For scalar types or incompatible types, if there's no diff, they're the same
	// Just show as unchanged
	return f.annotateNode(rightNode, statusUnchanged, indent)
}

// mergeMappingNodes merges two mapping nodes
func (f *unifiedFormatter) mergeMappingNodes(leftNode, rightNode *ast.MappingNode, diffMap map[string]*Diff, indent int) []annotatedLine {
	var lines []annotatedLine

	// Build key maps
	leftKeys := make(map[string]*ast.MappingValueNode)
	rightKeys := make(map[string]*ast.MappingValueNode)
	var allKeys []string // To maintain order
	keysSeen := make(map[string]bool)

	for _, v := range leftNode.Values {
		key := v.Key.String()
		leftKeys[key] = v
		if !keysSeen[key] {
			allKeys = append(allKeys, key)
			keysSeen[key] = true
		}
	}
	for _, v := range rightNode.Values {
		key := v.Key.String()
		rightKeys[key] = v
		if !keysSeen[key] {
			allKeys = append(allKeys, key)
			keysSeen[key] = true
		}
	}

	// Process each key
	for _, key := range allKeys {
		leftVal, leftOk := leftKeys[key]
		rightVal, rightOk := rightKeys[key]

		if leftOk && rightOk {
			// Key exists in both - merge the values
			keyPath := getNodePath(rightVal.Value)
			if diff, ok := diffMap[keyPath]; ok {
				// Value has a diff
				switch diff.Type() {
				case Modified:
					// Show key with old value as deleted
					deleted := f.formatMappingValue(leftVal, statusDeleted, indent)
					lines = append(lines, deleted...)
					// Show key with new value as added
					added := f.formatMappingValue(rightVal, statusAdded, indent)
					lines = append(lines, added...)
				case Deleted:
					deleted := f.formatMappingValue(leftVal, statusDeleted, indent)
					lines = append(lines, deleted...)
				case Added:
					added := f.formatMappingValue(rightVal, statusAdded, indent)
					lines = append(lines, added...)
				}
			} else {
				// No diff at value level - check for nested diffs
				if hasNestedDiffs(keyPath, diffMap) {
					// Has nested diffs - need to recurse
					keyLine := annotatedLine{
						status:  statusUnchanged,
						content: strings.Repeat(" ", indent) + key + ":",
					}
					lines = append(lines, keyLine)
					nested := f.mergeNodes(leftVal.Value, rightVal.Value, diffMap, indent+2)
					lines = append(lines, nested...)
				} else {
					// No nested diffs - show as unchanged
					unchanged := f.formatMappingValue(rightVal, statusUnchanged, indent)
					lines = append(lines, unchanged...)
				}
			}
		} else if leftOk {
			// Key only in left - deleted
			deleted := f.formatMappingValue(leftVal, statusDeleted, indent)
			lines = append(lines, deleted...)
		} else {
			// Key only in right - added
			added := f.formatMappingValue(rightVal, statusAdded, indent)
			lines = append(lines, added...)
		}
	}

	return lines
}

// mergeSequenceNodes merges two sequence nodes
func (f *unifiedFormatter) mergeSequenceNodes(leftNode, rightNode *ast.SequenceNode, diffMap map[string]*Diff, indent int) []annotatedLine {
	var lines []annotatedLine

	maxLen := max(len(leftNode.Values), len(rightNode.Values))
	for i := 0; i < maxLen; i++ {
		var leftVal, rightVal ast.Node
		if i < len(leftNode.Values) {
			leftVal = leftNode.Values[i]
		}
		if i < len(rightNode.Values) {
			rightVal = rightNode.Values[i]
		}

		if leftVal != nil && rightVal != nil {
			// Both have value at this index - merge them
			path := getNodePath(rightVal)
			if diff, ok := diffMap[path]; ok {
				switch diff.Type() {
				case Modified:
					deleted := f.annotateSequenceItem(leftVal, statusDeleted, indent)
					added := f.annotateSequenceItem(rightVal, statusAdded, indent)
					lines = append(lines, deleted...)
					lines = append(lines, added...)
				case Deleted:
					deleted := f.annotateSequenceItem(leftVal, statusDeleted, indent)
					lines = append(lines, deleted...)
				case Added:
					added := f.annotateSequenceItem(rightVal, statusAdded, indent)
					lines = append(lines, added...)
				}
			} else if hasNestedDiffs(path, diffMap) {
				// Has nested diffs
				itemLines := f.mergeSequenceItem(leftVal, rightVal, diffMap, indent)
				lines = append(lines, itemLines...)
			} else {
				// Unchanged
				unchanged := f.annotateSequenceItem(rightVal, statusUnchanged, indent)
				lines = append(lines, unchanged...)
			}
		} else if leftVal != nil {
			// Only in left - deleted
			deleted := f.annotateSequenceItem(leftVal, statusDeleted, indent)
			lines = append(lines, deleted...)
		} else {
			// Only in right - added
			added := f.annotateSequenceItem(rightVal, statusAdded, indent)
			lines = append(lines, added...)
		}
	}

	return lines
}

// mergeSequenceItem merges a sequence item that has nested diffs
func (f *unifiedFormatter) mergeSequenceItem(leftVal, rightVal ast.Node, diffMap map[string]*Diff, indent int) []annotatedLine {
	var lines []annotatedLine

	// For complex items (mappings), we need to handle the "- " prefix specially
	if leftVal.Type() == ast.MappingType && rightVal.Type() == ast.MappingType {
		// Add the "- " prefix line as unchanged context, then merge the contents
		leftMapping := leftVal.(*ast.MappingNode)
		rightMapping := rightVal.(*ast.MappingNode)

		// Merge the mapping values with indent+2 (for "- " prefix)
		nested := f.mergeMappingNodes(leftMapping, rightMapping, diffMap, indent+2)

		// Transform the first line to include "- " prefix
		if len(nested) > 0 {
			first := nested[0]
			// Replace leading spaces with "- "
			content := strings.TrimPrefix(first.content, strings.Repeat(" ", indent+2))
			first.content = strings.Repeat(" ", indent) + "- " + content
			nested[0] = first
		}
		lines = append(lines, nested...)
	} else {
		// Simple item - show as unchanged
		unchanged := f.annotateSequenceItem(rightVal, statusUnchanged, indent)
		lines = append(lines, unchanged...)
	}

	return lines
}

// formatMappingValue formats a mapping key-value pair
func (f *unifiedFormatter) formatMappingValue(mv *ast.MappingValueNode, status lineStatus, indent int) []annotatedLine {
	var lines []annotatedLine

	key := mv.Key.String()
	value := mv.Value

	// Check if value is a complex type
	switch value.Type() {
	case ast.MappingType, ast.SequenceType:
		// Complex value - key on its own line
		keyLine := annotatedLine{
			status:  status,
			content: strings.Repeat(" ", indent) + key + ":",
		}
		lines = append(lines, keyLine)
		valueLines := f.annotateNode(value, status, indent+2)
		lines = append(lines, valueLines...)
	default:
		// Simple value - key: value on same line
		valueStr := value.String()
		line := annotatedLine{
			status:  status,
			content: strings.Repeat(" ", indent) + key + ": " + valueStr,
		}
		lines = append(lines, line)
	}

	return lines
}

// annotateSequenceItem annotates a sequence item
func (f *unifiedFormatter) annotateSequenceItem(node ast.Node, status lineStatus, indent int) []annotatedLine {
	var lines []annotatedLine

	switch node.Type() {
	case ast.MappingType:
		// Complex item - needs "- " prefix on first line
		mapping := node.(*ast.MappingNode)
		if len(mapping.Values) > 0 {
			// First key gets "- " prefix
			first := mapping.Values[0]
			firstLines := f.formatMappingValue(first, status, indent+2)
			if len(firstLines) > 0 {
				// Transform first line to include "- "
				content := strings.TrimPrefix(firstLines[0].content, strings.Repeat(" ", indent+2))
				firstLines[0].content = strings.Repeat(" ", indent) + "- " + content
				lines = append(lines, firstLines...)
			}
			// Remaining keys
			for i := 1; i < len(mapping.Values); i++ {
				kvLines := f.formatMappingValue(mapping.Values[i], status, indent+2)
				lines = append(lines, kvLines...)
			}
		}
	case ast.SequenceType:
		// Nested sequence
		seq := node.(*ast.SequenceNode)
		for i, v := range seq.Values {
			if i == 0 {
				// First item gets the "- " prefix
				itemLines := f.annotateSequenceItem(v, status, indent+2)
				if len(itemLines) > 0 {
					content := strings.TrimPrefix(itemLines[0].content, strings.Repeat(" ", indent+2))
					if strings.HasPrefix(content, "- ") {
						itemLines[0].content = strings.Repeat(" ", indent) + "- " + content
					} else {
						itemLines[0].content = strings.Repeat(" ", indent) + "- " + content
					}
					lines = append(lines, itemLines...)
				}
			} else {
				itemLines := f.annotateSequenceItem(v, status, indent)
				lines = append(lines, itemLines...)
			}
		}
	default:
		// Simple value
		line := annotatedLine{
			status:  status,
			content: strings.Repeat(" ", indent) + "- " + node.String(),
		}
		lines = append(lines, line)
	}

	return lines
}

// annotateNode annotates all lines of a node with a status
func (f *unifiedFormatter) annotateNode(node ast.Node, status lineStatus, indent int) []annotatedLine {
	if node == nil {
		return nil
	}

	var lines []annotatedLine

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, mv := range n.Values {
			mvLines := f.formatMappingValue(mv, status, indent)
			lines = append(lines, mvLines...)
		}
	case *ast.SequenceNode:
		for _, v := range n.Values {
			itemLines := f.annotateSequenceItem(v, status, indent)
			lines = append(lines, itemLines...)
		}
	default:
		// Scalar value
		nodeStr := node.String()
		for _, lineStr := range strings.Split(nodeStr, "\n") {
			line := annotatedLine{
				status:  status,
				content: strings.Repeat(" ", indent) + lineStr,
			}
			lines = append(lines, line)
		}
	}

	return lines
}

// hasNestedDiffs checks if there are any diffs nested under the given path
func hasNestedDiffs(basePath string, diffMap map[string]*Diff) bool {
	for path := range diffMap {
		if strings.HasPrefix(path, basePath+".") || strings.HasPrefix(path, basePath+"[") {
			return true
		}
	}
	return false
}
