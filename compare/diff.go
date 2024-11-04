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
	NodeLeft  ast.Node
	NodeRight ast.Node
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

	if diffA.NodeLeft != nil {
		nodeA = diffA.NodeLeft
	} else {
		nodeA = diffA.NodeRight
	}

	if diffB.NodeLeft != nil {
		nodeB = diffB.NodeLeft
	} else {
		nodeB = diffB.NodeRight
	}

	return nodeA.GetToken().Position.Line < nodeB.GetToken().Position.Line
}

func (d DocDiffs) OutputString(opts *OutputOptions) string {
	b := strings.Builder{}
	for _, diff := range d {
		// todo: move this logic to Diff.OutputString
		if diff.NodeLeft == nil { // Added
			sign := "+"
			path := nodePathString(diff.NodeRight)
			value := nodeValueString(diff.NodeRight)
			metadata := nodeMetadata(diff.NodeRight)

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

		} else if diff.NodeRight == nil { //Deleted
			sign := "-"
			path := nodePathString(diff.NodeLeft)
			value := nodeValueString(diff.NodeLeft)
			metadata := nodeMetadata(diff.NodeLeft)

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
			path := nodePathString(diff.NodeLeft)
			leftValue := nodeValueString(diff.NodeLeft)
			rightValue := nodeValueString(diff.NodeRight)
			leftMetadata := nodeMetadata(diff.NodeLeft)
			rightMetadata := nodeMetadata(diff.NodeRight)

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
		b.WriteRune('\n')
	}
	return b.String()
}

type FileDiffs []DocDiffs

func (d FileDiffs) OutputString(opts *OutputOptions) string {
	docDiffsStrings := make([]string, 0, len(d))
	for _, docDiffs := range d {
		docDiffsStrings = append(docDiffsStrings, docDiffs.OutputString(opts))
	}
	return strings.Join(docDiffsStrings, "\n---\n")
}

func (d FileDiffs) HasDiff() bool {
	return len(d) > 0
}

type DiffOptions struct {
	IgnoreIndex bool
}

var DefaultDiffOptions = &DiffOptions{
	IgnoreIndex: false,
}

type OutputOptions struct {
	Plain    bool
	Silent   bool
	Metadata bool
}

var DefaultOutputOptions = &OutputOptions{
	Plain:    false,
	Silent:   false,
	Metadata: false,
}

func Compare(left []byte, right []byte, comments bool, opts *DiffOptions) (FileDiffs, error) {
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

	return compareAst(leftAst, rightAst, opts), nil
}

func CompareFile(leftFile string, rightFile string, comments bool, opts *DiffOptions) (FileDiffs, error) {
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

	return compareAst(leftAst, rightAst, opts), nil
}

func compareAst(left *ast.File, right *ast.File, opts *DiffOptions) FileDiffs {
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
