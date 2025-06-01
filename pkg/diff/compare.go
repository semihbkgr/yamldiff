package diff

import (
	"sort"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// CompareOptions specifies options for customizing the behavior of the comparison.
type CompareOptions struct {
	// IgnoreSeqOrder, when true, treats arrays as equal regardless of the order of their items.
	// For instance, the arrays [1, 2] and [2, 1] will be considered equal.
	IgnoreSeqOrder bool
}

var DefaultCompareOptions = CompareOptions{
	IgnoreSeqOrder: false,
}

// Compare compares two yaml provided as bytes and returns the differences as FileDiffs,
// or an error if there's an issue parsing the files.
func Compare(left []byte, right []byte, opts CompareOptions) (FileDiffs, error) {
	leftAst, err := parser.ParseBytes(left, 0)
	if err != nil {
		return nil, err
	}

	rightAst, err := parser.ParseBytes(right, 0)
	if err != nil {
		return nil, err
	}

	return CompareAst(leftAst, rightAst, opts), nil
}

// CompareFile compares two yaml files specified by file paths and returns the differences as FileDiffs,
// or an error if there's an issue reading or parsing the files.
func CompareFile(leftFile string, rightFile string, opts CompareOptions) (FileDiffs, error) {
	leftAst, err := parser.ParseFile(leftFile, 0)
	if err != nil {
		return nil, err
	}

	rightAst, err := parser.ParseFile(rightFile, 0)
	if err != nil {
		return nil, err
	}

	return CompareAst(leftAst, rightAst, opts), nil
}

// CompareAst compares two yaml documents represented as ASTs and returns the differences as FileDiffs.
func CompareAst(left *ast.File, right *ast.File, opts CompareOptions) FileDiffs {
	var docDiffs = make(FileDiffs, max(len(left.Docs), len(left.Docs)))
	for i := 0; i < len(docDiffs); i++ {
		var l, r ast.Node
		if len(left.Docs) > i {
			l = left.Docs[i].Body
		}
		if len(right.Docs) > i {
			r = right.Docs[i].Body
		}
		docDiff := DocDiffs(compareNodes(l, r, opts))
		sort.Sort(docDiff)
		docDiffs[i] = docDiff
	}
	return docDiffs
}
