package compare

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type Diff struct {
	leftNode  ast.Node
	rightNode ast.Node
}

func (d *Diff) Format(opts FormatOptions) string {
	var b strings.Builder
	if d.leftNode == nil { // Added
		sign := "+"
		path := nodePathString(d.rightNode)
		value := nodeValueString(d.rightNode)
		metadata := nodeMetadata(d.rightNode)

		if !opts.Plain {
			sign = color.HiGreenString(sign)
			path = color.HiGreenString(path)
			value = color.HiWhiteString(value)
			metadata = color.HiCyanString(metadata)
		}

		if opts.Silent {
			b.WriteString(fmt.Sprintf("%s %s", sign, path))
		} else {
			if opts.Metadata {
				b.WriteString(fmt.Sprintf("%s %s: %s %s", sign, path, metadata, value))
			} else {
				b.WriteString(fmt.Sprintf("%s %s: %s", sign, path, value))
			}
		}

	} else if d.rightNode == nil { //Deleted
		sign := "-"
		path := nodePathString(d.leftNode)
		value := nodeValueString(d.leftNode)
		metadata := nodeMetadata(d.leftNode)

		if !opts.Plain {
			sign = color.HiRedString(sign)
			path = color.HiRedString(path)
			value = color.HiWhiteString(value)
			metadata = color.HiCyanString(metadata)
		}

		if opts.Silent {
			b.WriteString(fmt.Sprintf("%s %s", sign, path))
		} else {
			if opts.Metadata {
				b.WriteString(fmt.Sprintf("%s %s: %s %s", sign, path, metadata, value))
			} else {
				b.WriteString(fmt.Sprintf("%s %s: %s", sign, path, value))
			}
		}
	} else { //Modified
		sign := "~"
		path := nodePathString(d.leftNode)
		leftValue := nodeValueString(d.leftNode)
		rightValue := nodeValueString(d.rightNode)
		leftMetadata := nodeMetadata(d.leftNode)
		rightMetadata := nodeMetadata(d.rightNode)

		if !opts.Plain {
			sign = color.HiYellowString(sign)
			path = color.HiYellowString(path)
			leftValue = color.HiWhiteString(leftValue)
			rightValue = color.HiWhiteString(rightValue)
			leftMetadata = color.HiCyanString(leftMetadata)
			rightMetadata = color.HiCyanString(rightMetadata)
		}

		if opts.Silent {
			b.WriteString(fmt.Sprintf("%s %s", sign, path))
		} else {
			if opts.Metadata {
				b.WriteString(fmt.Sprintf("%s %s: %s %s -> %s %s", sign, path, leftMetadata, leftValue, rightMetadata, rightValue))
			} else {
				b.WriteString(fmt.Sprintf("%s %s: %s -> %s", sign, path, leftValue, rightValue))
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

	return nodeA.GetToken().Position.Line < nodeB.GetToken().Position.Line
}

func (d DocDiffs) Format(opts FormatOptions) string {
	diffsStrings := make([]string, 0, len(d))
	for _, diff := range d {
		diffsStrings = append(diffsStrings, diff.Format(opts))
	}
	return strings.Join(diffsStrings, "\n")
}

type FileDiffs []DocDiffs

func (d FileDiffs) Format(opts FormatOptions) string {
	docDiffsStrings := make([]string, 0, len(d))
	for _, docDiffs := range d {
		docDiffsStrings = append(docDiffsStrings, docDiffs.Format(opts))
	}
	return strings.Join(docDiffsStrings, "\n---\n")
}

func (d FileDiffs) HasDiff() bool {
	return len(d) > 0
}

// Compare compares two yaml files provided as bytes and returns the differences as FileDiffs,
// or an error if there's an issue parsing the files.
func Compare(left []byte, right []byte, comments bool, opts DiffOptions) (FileDiffs, error) {
	var parserMode parser.Mode
	if comments {
		parserMode |= parser.ParseComments
	}

	leftAst, err := parser.ParseBytes(left, parserMode)
	if err != nil {
		return nil, err
	}

	rightAst, err := parser.ParseBytes(right, parserMode)
	if err != nil {
		return nil, err
	}

	return CompareAst(leftAst, rightAst, opts), nil
}

// CompareFile compares two yaml files specified by file paths and returns the differences as FileDiffs,
// or an error if there's an issue reading or parsing the files.
func CompareFile(leftFile string, rightFile string, comments bool, opts DiffOptions) (FileDiffs, error) {
	var parserMode parser.Mode
	if comments {
		parserMode |= parser.ParseComments
	}

	leftAst, err := parser.ParseFile(leftFile, parserMode)
	if err != nil {
		return nil, err
	}

	rightAst, err := parser.ParseFile(rightFile, parserMode)
	if err != nil {
		return nil, err
	}

	return CompareAst(leftAst, rightAst, opts), nil
}

// CompareAst compares two yaml documents represented as ASTs and returns the differences as FileDiffs.
func CompareAst(left *ast.File, right *ast.File, opts DiffOptions) FileDiffs {
	var docDiffs = make(FileDiffs, max(len(left.Docs), len(left.Docs)))
	for i := 0; i < len(docDiffs); i++ {
		var l, r *ast.DocumentNode
		if len(left.Docs) > i {
			l = left.Docs[i]
		}
		if len(right.Docs) > i {
			r = right.Docs[i]
		}
		docDiff := DocDiffs(compareNodes(l.Body, r.Body, opts))
		sort.Sort(docDiff)
		docDiffs[i] = docDiff
	}
	return docDiffs
}

// DiffOptions specifies options for customizing the behavior of the comparison.
type DiffOptions struct {
	// IgnoreSeqOrder, when true, treats arrays as equal regardless of the order of their items.
	// For instance, the arrays [1, 2] and [2, 1] will be considered equal.
	IgnoreSeqOrder bool
}

var DefaultDiffOptions = DiffOptions{
	IgnoreSeqOrder: false,
}

// FormatOptions specifies options for formatting the output of the comparison.
type FormatOptions struct {
	// Plain disables colored output when set to true.
	Plain bool

	// Silent suppresses the display of values when set to true.
	Silent bool

	// Metadata includes additional metadata, such as line numbers or types, when set to true.
	Metadata bool
}

var DefaultOutputOptions = FormatOptions{
	Plain:    false,
	Silent:   false,
	Metadata: false,
}
