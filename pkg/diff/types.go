package diff

import (
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// DiffType represents the type of change in a diff
type DiffType int

const (
	Added DiffType = iota
	Deleted
	Modified
)

// String returns the string representation of the DiffType
func (dt DiffType) String() string {
	switch dt {
	case Added:
		return "Added"
	case Deleted:
		return "Deleted"
	case Modified:
		return "Modified"
	default:
		return "Unknown"
	}
}

// Position represents a byte offset range in source text.
// Zero value means no range (nil node).
type Position struct {
	Start int // 0-based byte offset, inclusive
	End   int // 0-based byte offset, exclusive
}

// Diff represents a single difference between two YAML nodes
type Diff struct {
	leftNode  ast.Node
	rightNode ast.Node
	// leftSourceNode and rightSourceNode represent the source-level nodes
	// that this diff originates from. When set, LeftPosition/RightPosition use
	// these for range calculation instead of leftNode/rightNode. This
	// allows the highlighted region to be broader than the node used for
	// formatting (e.g. the full key: value pair when a key is added or deleted).
	leftSourceNode  ast.Node
	rightSourceNode ast.Node
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

// LeftPosition returns the source text range of the left side.
// Uses the source node if set, otherwise falls back to the diff node.
// Returns zero Position if the left node is nil.
func (d *Diff) LeftPosition() Position {
	if d.leftSourceNode != nil {
		return nodePosition(d.leftSourceNode)
	}
	return nodePosition(d.leftNode)
}

// RightPosition returns the source text range of the right side.
// Uses the source node if set, otherwise falls back to the diff node.
// Returns zero Position if the right node is nil.
func (d *Diff) RightPosition() Position {
	if d.rightSourceNode != nil {
		return nodePosition(d.rightSourceNode)
	}
	return nodePosition(d.rightNode)
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

func nodePosition(node ast.Node) Position {
	if node == nil {
		return Position{}
	}
	return Position{
		Start: nodeStartOffset(node),
		End:   nodeEndOffset(node),
	}
}

func nodeStartOffset(node ast.Node) int {
	switch n := node.(type) {
	case *ast.MappingNode:
		if len(n.Values) > 0 {
			return n.Values[0].Key.GetToken().Position.Offset - 1
		}
		return n.GetToken().Position.Offset - 1
	case *ast.MappingValueNode:
		return n.Key.GetToken().Position.Offset - 1
	default:
		return node.GetToken().Position.Offset - 1
	}
}

func nodeEndOffset(node ast.Node) int {
	switch n := node.(type) {
	case *ast.StringNode:
		return tokenEndOffset(n.Token)
	case *ast.IntegerNode:
		return tokenEndOffset(n.Token)
	case *ast.FloatNode:
		return tokenEndOffset(n.Token)
	case *ast.BoolNode:
		return tokenEndOffset(n.Token)
	case *ast.InfinityNode:
		return tokenEndOffset(n.Token)
	case *ast.NanNode:
		return tokenEndOffset(n.Token)
	case *ast.NullNode:
		return tokenEndOffset(n.Token)
	case *ast.LiteralNode:
		return tokenEndOffset(n.Value.Token)
	case *ast.MappingNode:
		if n.End != nil {
			return tokenEndOffset(n.End)
		}
		if len(n.Values) > 0 {
			return nodeEndOffset(n.Values[len(n.Values)-1])
		}
		return tokenEndOffset(n.GetToken())
	case *ast.SequenceNode:
		if n.End != nil {
			return tokenEndOffset(n.End)
		}
		if len(n.Values) > 0 {
			return nodeEndOffset(n.Values[len(n.Values)-1])
		}
		return tokenEndOffset(n.GetToken())
	case *ast.MappingValueNode:
		return nodeEndOffset(n.Value)
	default:
		tok := node.GetToken()
		return tok.Position.Offset - 1 + len(tok.Value)
	}
}

func tokenEndOffset(tok *token.Token) int {
	return tok.Position.Offset - 1 + len(strings.Trim(tok.Origin, " \t\n\r"))
}
