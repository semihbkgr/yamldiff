package diff

import (
	"sort"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type CompareOption func(o *compareOptions)

// IgnoreSeqOrder indicates whether to ignore the order of sequence items when comparing.
// For example, the sequences [1, 2] and [2, 1] will be considered equal.
func IgnoreSeqOrder(o *compareOptions) {
	o.ignoreSeqOrder = true
}

// Compare compares two yaml provided as bytes and returns the differences as FileDiffs,
// or an error if there's an issue parsing the files.
func Compare(left []byte, right []byte, opts ...CompareOption) (FileDiffs, error) {
	leftAst, err := parser.ParseBytes(left, 0)
	if err != nil {
		return nil, err
	}

	rightAst, err := parser.ParseBytes(right, 0)
	if err != nil {
		return nil, err
	}

	return CompareAst(leftAst, rightAst, opts...), nil
}

// CompareFile compares two yaml files specified by file paths and returns the differences as FileDiffs,
// or an error if there's an issue reading or parsing the files.
func CompareFile(leftFile string, rightFile string, opts ...CompareOption) (FileDiffs, error) {
	leftAst, err := parser.ParseFile(leftFile, 0)
	if err != nil {
		return nil, err
	}

	rightAst, err := parser.ParseFile(rightFile, 0)
	if err != nil {
		return nil, err
	}

	return CompareAst(leftAst, rightAst, opts...), nil
}

// CompareAst compares two yaml documents represented as ASTs and returns the differences as FileDiffs.
func CompareAst(left *ast.File, right *ast.File, opts ...CompareOption) FileDiffs {
	options := &compareOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var docDiffs = make(FileDiffs, max(len(left.Docs), len(right.Docs)))
	for i := 0; i < len(docDiffs); i++ {
		var l, r ast.Node
		if len(left.Docs) > i {
			l = left.Docs[i].Body
		}
		if len(right.Docs) > i {
			r = right.Docs[i].Body
		}
		docDiff := DocDiffs(compareNodes(l, r, options))
		sort.Sort(docDiff)
		docDiffs[i] = docDiff
	}
	return docDiffs
}

type compareOptions struct {
	ignoreSeqOrder bool
}

func compareNodes(ln, rn ast.Node, options *compareOptions) []*Diff {
	// If both nodes are nil in the AST representation, return nil (no differences)
	// A nil node is different than a node in null type in AST.
	// The node is nil when:
	// - the yaml document has no content
	// - the yaml document does not exist (when comparing multiple documents)
	// see:
	// - https://github.com/semihbkgr/yamldiff/issues/26
	// - https://github.com/goccy/go-yaml/issues/753
	if ln == nil && rn == nil {
		return nil
	}

	if ln == nil || rn == nil || !comparableNodes(ln, rn) {
		return []*Diff{newDiff(ln, rn)}
	}

	// Handle all value types from the YAML AST.
	// Note: Structural types (DocumentType, MappingKeyType, MappingValueType, SequenceEntryType,
	// MergeKeyType, UnknownNodeType) are handled through recursive traversal of their parent
	// collection types (MappingType, SequenceType) and don't need explicit comparison here.
	switch ln.Type() {
	case ast.MappingType:
		return compareMappingNodes(ln.(*ast.MappingNode), rn.(*ast.MappingNode), options)
	case ast.SequenceType:
		return compareSequenceNodes(ln.(*ast.SequenceNode), rn.(*ast.SequenceNode), options)
	case ast.StringType, ast.LiteralType:
		leftNodeString := stringNodeValue(ln)
		rightNodeString := stringNodeValue(rn)
		if leftNodeString != rightNodeString {
			return []*Diff{newDiff(ln, rn)}
		}
	case ast.IntegerType:
		leftIntegerNode := ln.(*ast.IntegerNode)
		rightIntegerNode := rn.(*ast.IntegerNode)
		if leftIntegerNode.Value != rightIntegerNode.Value {
			return []*Diff{newDiff(ln, rn)}
		}
	case ast.FloatType:
		leftFloatNode := ln.(*ast.FloatNode)
		rightFloatNode := rn.(*ast.FloatNode)
		if leftFloatNode.Value != rightFloatNode.Value {
			return []*Diff{newDiff(ln, rn)}
		}
	case ast.BoolType:
		leftBoolNode := ln.(*ast.BoolNode)
		rightBoolNode := rn.(*ast.BoolNode)
		if leftBoolNode.Value != rightBoolNode.Value {
			return []*Diff{newDiff(ln, rn)}
		}
	case ast.InfinityType:
		leftInfinityNode := ln.(*ast.InfinityNode)
		rightInfinityNode := rn.(*ast.InfinityNode)
		if leftInfinityNode.Value != rightInfinityNode.Value {
			return []*Diff{newDiff(ln, rn)}
		}
	case ast.NullType, ast.NanType:
		return nil
	}
	return nil
}

func comparableNodes(leftNode, rightNode ast.Node) bool {
	switch leftNode.Type() {
	case ast.StringType, ast.LiteralType:
		return rightNode.Type() == ast.StringType || rightNode.Type() == ast.LiteralType
	default:
		return leftNode.Type() == rightNode.Type()
	}
}

func compareMappingNodes(leftNode, rightNode *ast.MappingNode, options *compareOptions) []*Diff {
	leftKeyValueMap := mappingValueNodesIntoMap(leftNode)
	rightKeyValueMap := mappingValueNodesIntoMap(rightNode)
	keyDiffsMap := make(map[string][]*Diff)
	for k, leftValue := range leftKeyValueMap {
		rightValue, ok := rightKeyValueMap[k]
		if !ok {
			node := wrapMappingValueNode(leftValue.Value)
			keyDiffsMap[k] = []*Diff{newDiff(node, nil)}
			continue
		}
		keyDiffsMap[k] = compareNodes(leftValue.Value, rightValue.Value, options)
	}
	for k, rightValue := range rightKeyValueMap {
		_, ok := keyDiffsMap[k]
		if ok {
			continue
		}
		node := wrapMappingValueNode(rightValue.Value)
		keyDiffsMap[k] = []*Diff{newDiff(nil, node)}
	}

	allDiffs := make([]*Diff, 0)
	for _, diffs := range keyDiffsMap {
		if diffs != nil {
			allDiffs = append(allDiffs, diffs...)
		}
	}

	return allDiffs
}

// wrapMappingValueNode wraps a MappingValueNode in a MappingNode while preserving its path.
// This is necessary for proper diff representation when a mapping value is added or deleted.
func wrapMappingValueNode(node ast.Node) ast.Node {
	if node.Type() == ast.MappingValueType {
		path := node.GetPath()
		node = ast.Mapping(node.GetToken(), false, node.(*ast.MappingValueNode))
		node.SetPath(path)
	}
	return node
}

func mappingValueNodesIntoMap(n *ast.MappingNode) map[string]*ast.MappingValueNode {
	keyValueMap := make(map[string]*ast.MappingValueNode)
	for _, values := range n.Values {
		keyValueMap[values.Key.String()] = values
	}
	return keyValueMap
}

func compareSequenceNodes(leftNode, rightNode *ast.SequenceNode, options *compareOptions) []*Diff {
	diffs := make([]*Diff, 0)
	l := max(len(leftNode.Values), len(rightNode.Values))
	for i := 0; i < l; i++ {
		var leftValue, rightValue ast.Node
		if len(leftNode.Values) > i {
			leftValue = leftNode.Values[i]
		}
		if len(rightNode.Values) > i {
			rightValue = rightNode.Values[i]
		}
		diffs = append(diffs, compareNodes(leftValue, rightValue, options)...)
	}

	if options.ignoreSeqOrder {
		diffs = ignoreIndexes(diffs, options)
	}

	return diffs
}

func ignoreIndexes(diffs []*Diff, options *compareOptions) []*Diff {
	leftNodes := make([]ast.Node, len(diffs))
	rightNodes := make([]ast.Node, len(diffs))
	for i, diff := range diffs {
		leftNodes[i] = diff.leftNode
		rightNodes[i] = diff.rightNode
	}

	for il, leftNode := range leftNodes {
		if leftNode == nil {
			continue
		}
		for ir, rightNode := range rightNodes {
			if rightNode == nil {
				continue
			}
			if len(compareNodes(leftNode, rightNode, options)) == 0 {
				leftNodes[il] = nil
				rightNodes[ir] = nil
				break
			}
		}
	}

	resultDiffs := make([]*Diff, 0)
	for i := range diffs {
		leftNode := leftNodes[i]
		rightNode := rightNodes[i]
		if leftNode == nil && rightNode == nil {
			continue
		}
		resultDiffs = append(resultDiffs, newDiff(leftNode, rightNode))
	}

	return resultDiffs
}

func stringNodeValue(node ast.Node) string {
	switch t := node.(type) {
	case *ast.StringNode:
		return t.Value
	case *ast.LiteralNode:
		return t.Value.Value
	default:
		return ""
	}
}
