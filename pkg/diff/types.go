package diff

import "github.com/goccy/go-yaml/ast"

// DiffType represents the type of change in a diff
type DiffType int

const (
	Added DiffType = iota
	Deleted
	Modified
)

// Diff represents a single difference between two YAML nodes
type Diff struct {
	leftNode  ast.Node
	rightNode ast.Node
}

// newDiff creates a new Diff instance
func newDiff(left, right ast.Node) *Diff {
	return &Diff{
		leftNode:  left,
		rightNode: right,
	}
}

// Type returns the type of the diff (Added, Deleted, or Modified)
func (d *Diff) Type() DiffType {
	if d.leftNode == nil {
		return Added
	}
	if d.rightNode == nil {
		return Deleted
	}
	return Modified
}

// Path returns the YAML path for this diff
func (d *Diff) Path() string {
	switch d.Type() {
	case Added:
		return getNodePath(d.rightNode)
	default:
		return getNodePath(d.leftNode)
	}
}

// LeftNode returns the left node of the diff
func (d *Diff) LeftNode() ast.Node {
	return d.leftNode
}

// RightNode returns the right node of the diff
func (d *Diff) RightNode() ast.Node {
	return d.rightNode
}

// Format formats the diff using the provided options
func (d *Diff) Format(opts ...FormatOption) string {
	formatter := newFormatter(opts...)
	return formatter.formatDiff(d)
}

// DocDiffs represents a collection of diffs for a single YAML document
type DocDiffs []*Diff

// Len implements sort.Interface
func (d DocDiffs) Len() int {
	return len(d)
}

// Swap implements sort.Interface
func (d DocDiffs) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less implements sort.Interface
func (d DocDiffs) Less(i, j int) bool {
	diffA := d[i]
	diffB := d[j]

	nodeA := diffA.leftNode
	if nodeA == nil {
		nodeA = diffA.rightNode
	}

	nodeB := diffB.leftNode
	if nodeB == nil {
		nodeB = diffB.rightNode
	}

	// Sort by line number first
	lineDelta := nodeA.GetToken().Position.Line - nodeB.GetToken().Position.Line
	if lineDelta != 0 {
		return lineDelta < 0
	}

	// If same line, sort by diff type for consistent ordering: Deleted < Modified < Added
	// This ensures deletions appear before modifications, and modifications before additions.
	// We use > because DiffType enum is: Added(0) < Deleted(1) < Modified(2)
	typeA := diffA.Type()
	typeB := diffB.Type()
	return typeA > typeB
}

// Format formats the document diffs using the provided options
func (d DocDiffs) Format(opts ...FormatOption) string {
	formatter := newFormatter(opts...)
	return formatter.formatDocDiffs(d)
}

// FileDiffs represents diffs for multiple YAML documents in a file
type FileDiffs []DocDiffs

// HasDiff returns true if there are any differences
func (f FileDiffs) HasDiff() bool {
	for _, docDiffs := range f {
		if len(docDiffs) > 0 {
			return true
		}
	}
	return false
}

// Format formats the file diffs using the provided options
func (f FileDiffs) Format(opts ...FormatOption) string {
	formatter := newFormatter(opts...)
	return formatter.formatFileDiffs(f)
}
