package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
)

type Diff struct {
	leftNode  ast.Node
	rightNode ast.Node
}

// todo: refactor this function
func (d *Diff) Format(opts ...FormatOption) string {
	options := &formatOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var b strings.Builder
	if d.leftNode == nil { // Added
		sign := "+"
		path := nodePathString(d.rightNode)
		indent := 4
		newLine := true
		if path == "" {
			indent = 0
			newLine = false
		}
		value, _ := nodeValueString(d.rightNode, indent, newLine)
		metadata := nodeMetadata(d.rightNode)

		if !options.plain {
			sign = color.HiGreenString(sign)
			if path != "" {
				path = color.HiGreenString(path)
			}
			value = colorize(value)
			metadata = color.HiWhiteString(metadata)
		}

		if options.silent {
			b.WriteString(fmt.Sprintf("%s %s", sign, path))
		} else {
			if path == "" {
				if options.metadata {
					b.WriteString(fmt.Sprintf("%s %s %s", sign, metadata, value))
				} else {
					b.WriteString(fmt.Sprintf("%s %s", sign, value))
				}
			} else {
				if options.metadata {
					b.WriteString(fmt.Sprintf("%s %s: %s %s", sign, path, metadata, value))
				} else {
					b.WriteString(fmt.Sprintf("%s %s: %s", sign, path, value))
				}
			}
		}

	} else if d.rightNode == nil { //Deleted
		sign := "-"
		path := nodePathString(d.leftNode)
		indent := 4
		newLine := true
		if path == "" {
			indent = 0
			newLine = false
		}
		value, _ := nodeValueString(d.leftNode, indent, newLine)
		metadata := nodeMetadata(d.leftNode)

		if !options.plain {
			sign = color.HiRedString(sign)
			if path != "" {
				path = color.HiRedString(path)
			}
			value = colorize(value)
			metadata = color.HiWhiteString(metadata)
		}

		if options.silent {
			b.WriteString(fmt.Sprintf("%s %s", sign, path))
		} else {
			if path == "" {
				if options.metadata {
					b.WriteString(fmt.Sprintf("%s %s %s", sign, metadata, value))
				} else {
					b.WriteString(fmt.Sprintf("%s %s", sign, value))
				}
			} else {
				if options.metadata {
					b.WriteString(fmt.Sprintf("%s %s: %s %s", sign, path, metadata, value))
				} else {
					b.WriteString(fmt.Sprintf("%s %s: %s", sign, path, value))
				}
			}
		}
	} else { //Modified
		sign := "~"
		path := nodePathString(d.leftNode)
		indent := 4
		newLine := true
		if path == "" {
			indent = 0
			newLine = false
		}
		leftValue, leftMultiLine := nodeValueString(d.leftNode, indent, newLine)
		rightValue, rightMultiLine := nodeValueString(d.rightNode, indent, newLine)
		leftMetadata := nodeMetadata(d.leftNode)
		rightMetadata := nodeMetadata(d.rightNode)

		symbol := "->"

		multiline := leftMultiLine || rightMultiLine
		if multiline {
			symbol = "\n    ↓"
		}

		if multiline && !leftMultiLine {
			if path != "" {
				leftValue = fmt.Sprintf("\n    %s", leftValue)
			} else {
				rightValue = fmt.Sprintf("\n  %s", rightValue)
			}
		} else if multiline && !rightMultiLine {
			if path != "" {
				rightValue = fmt.Sprintf("\n    %s", rightValue)
			} else {
				rightValue = fmt.Sprintf("\n  %s", rightValue)
			}
		} else if multiline {
			if path == "" {
				rightValue = fmt.Sprintf("\n  %s", rightValue)
			}
		}

		if !options.plain {
			sign = color.HiYellowString(sign)
			if path != "" {
				path = color.HiYellowString(path)
			}
			leftValue = colorize(leftValue)
			rightValue = colorize(rightValue)
			leftMetadata = color.HiWhiteString(leftMetadata)
			rightMetadata = color.HiWhiteString(rightMetadata)
		}

		if options.silent {
			b.WriteString(fmt.Sprintf("%s %s", sign, path))
		} else if path == "" {
			if options.metadata {
				b.WriteString(fmt.Sprintf("%s %s %s %s %s %s", sign, leftMetadata, leftValue, symbol, rightMetadata, rightValue))
			} else {
				b.WriteString(fmt.Sprintf("%s %s %s %s", sign, leftValue, symbol, rightValue))
			}
		} else {
			if options.metadata {
				b.WriteString(fmt.Sprintf("%s %s: %s %s %s %s %s", sign, path, leftMetadata, leftValue, symbol, rightMetadata, rightValue))
			} else {
				b.WriteString(fmt.Sprintf("%s %s: %s %s %s", sign, path, leftValue, symbol, rightValue))
			}
		}
	}
	return b.String()
}

type DocDiffs []*Diff

func (a DocDiffs) Len() int {
	return len(a)
}

func (a DocDiffs) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a DocDiffs) Less(i, j int) bool {
	diffA := a[i]
	diffB := a[j]
	var nodeA ast.Node
	var nodeB ast.Node

	if diffA.leftNode != nil {
		nodeA = diffA.leftNode
	} else {
		nodeA = diffA.rightNode
	}

	if diffB.leftNode != nil {
		nodeB = diffB.leftNode
	} else {
		nodeB = diffB.rightNode
	}

	//todo: it is still partial ordered, which can cause the inconsistent order of items after sorting
	if nodeA.GetToken().Position.Line == nodeB.GetToken().Position.Line {
		return diffA.leftNode != nil
	}

	return nodeA.GetToken().Position.Line < nodeB.GetToken().Position.Line
}

func (d DocDiffs) Format(opts ...FormatOption) string {
	diffsStrings := make([]string, 0, len(d))
	for _, diff := range d {
		diffsStrings = append(diffsStrings, diff.Format(opts...))
	}
	return strings.Join(diffsStrings, "\n")
}

type FileDiffs []DocDiffs

func (d FileDiffs) Format(opts ...FormatOption) string {
	docDiffsStrings := make([]string, 0, len(d))
	for _, docDiffs := range d {
		docDiffsStrings = append(docDiffsStrings, docDiffs.Format(opts...))
	}
	return strings.Join(docDiffsStrings, "\n---\n")
}

func (d FileDiffs) HasDiff() bool {
	return len(d) > 0
}

// formatOptions specifies options for formatting the output of the comparison.
type formatOptions struct {
	// plain disables colored output when set to true.
	plain bool

	// silent suppresses the display of values when set to true.
	silent bool

	// metadata includes additional metadata, such as line numbers or types, when set to true.
	metadata bool
}

type FormatOption func(*formatOptions)

func Plain(opts *formatOptions) {
	opts.plain = true
}

func Silent(opts *formatOptions) {
	opts.silent = true
}

func IncludeMetadata(opts *formatOptions) {
	opts.metadata = true
}
