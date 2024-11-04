package compare

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
)

func compareNodes(leftNode, rightNode ast.Node, opts DiffOptions) []*Diff {
	if leftNode == nil {
		return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
	}

	if rightNode == nil {
		return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
	}

	// When the map's key size is one, it is just represented by MappingValueNode instead of MappingNode in AST.
	// Wrap MappingValueNode by MappingNode if needed.
	if leftNode.Type() == ast.MappingValueType {
		path := leftNode.GetPath()
		leftNode = ast.Mapping(leftNode.GetToken(), false, leftNode.(*ast.MappingValueNode))
		leftNode.SetPath(path)
	}
	if rightNode.Type() == ast.MappingValueType {
		path := rightNode.GetPath()
		rightNode = ast.Mapping(rightNode.GetToken(), false, rightNode.(*ast.MappingValueNode))
		rightNode.SetPath(path)
	}

	if leftNode.Type() != rightNode.Type() {
		return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
	}

	switch leftNode.Type() {
	case ast.MappingType:
		return compareMappingNodes(leftNode.(*ast.MappingNode), rightNode.(*ast.MappingNode), opts)
	case ast.SequenceType:
		return compareSequenceNodes(leftNode.(*ast.SequenceNode), rightNode.(*ast.SequenceNode), opts)
	case ast.StringType:
		leftStringNode := leftNode.(*ast.StringNode)
		rightStringNode := rightNode.(*ast.StringNode)
		if leftStringNode.Value != rightStringNode.Value {
			return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
		}
	case ast.IntegerType:
		leftIntegerNode := leftNode.(*ast.IntegerNode)
		rightIntegerNode := rightNode.(*ast.IntegerNode)
		if leftIntegerNode.Value != rightIntegerNode.Value {
			return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
		}
	case ast.FloatType:
		leftFloatNode := leftNode.(*ast.FloatNode)
		rightFloatNode := rightNode.(*ast.FloatNode)
		if leftFloatNode.Value != rightFloatNode.Value {
			return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
		}
	case ast.BoolType:
		leftBoolNode := leftNode.(*ast.BoolNode)
		rightBoolNode := rightNode.(*ast.BoolNode)
		if leftBoolNode.Value != rightBoolNode.Value {
			return []*Diff{{leftNode: leftNode, rightNode: rightNode}}
		}
	}
	return nil
}

func compareMappingNodes(leftNode, rightNode *ast.MappingNode, opts DiffOptions) []*Diff {
	leftKeyValueMap := mappingValueNodesIntoMap(leftNode)
	rightKeyValueMap := mappingValueNodesIntoMap(rightNode)
	keyDiffsMap := make(map[string][]*Diff)
	for k, leftValue := range leftKeyValueMap {
		rightValue, ok := rightKeyValueMap[k]
		if !ok {
			node := leftValue.Value
			// wrap the MappingValueNode by MappingNode
			// todo: extract this logic into a function
			if node.Type() == ast.MappingValueType {
				path := node.GetPath()
				node = ast.Mapping(node.GetToken(), false, node.(*ast.MappingValueNode))
				node.SetPath(path)
			}
			keyDiffsMap[k] = []*Diff{{leftNode: node, rightNode: nil}}
			continue
		}
		keyDiffsMap[k] = compareNodes(leftValue.Value, rightValue.Value, opts)
	}
	for k, rightValue := range rightKeyValueMap {
		_, ok := keyDiffsMap[k]
		if ok {
			continue
		}
		node := rightValue.Value
		// wrap the MappingValueNode by MappingNode
		if node.Type() == ast.MappingValueType {
			path := node.GetPath()
			node = ast.Mapping(node.GetToken(), false, node.(*ast.MappingValueNode))
			node.SetPath(path)
		}
		keyDiffsMap[k] = []*Diff{{leftNode: nil, rightNode: node}}
	}

	allDiffs := make([]*Diff, 0)
	for _, diffs := range keyDiffsMap {
		if diffs != nil {
			allDiffs = append(allDiffs, diffs...)
		}
	}

	return allDiffs
}

func mappingValueNodesIntoMap(n *ast.MappingNode) map[string]*ast.MappingValueNode {
	keyValueMap := make(map[string]*ast.MappingValueNode)
	for _, values := range n.Values {
		keyValueMap[values.Key.String()] = values
	}
	return keyValueMap
}

func compareSequenceNodes(leftNode, rightNode *ast.SequenceNode, opts DiffOptions) []*Diff {
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
		diffs = append(diffs, compareNodes(leftValue, rightValue, opts)...)
	}

	if opts.IgnoreSeqOrder {
		diffs = ignoreIndexes(diffs, opts)
	}

	return diffs
}

func ignoreIndexes(diffs []*Diff, opts DiffOptions) []*Diff {
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
			if len(compareNodes(leftNode, rightNode, opts)) == 0 {
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
		resultDiffs = append(resultDiffs, &Diff{leftNode, rightNode})
	}

	return resultDiffs
}

func nodePathString(n ast.Node) string {
	path := n.GetPath()[2:]
	// Path of the MappingNode points to the first key in the map.
	if n.Type() == ast.MappingType {
		path = path[:strings.LastIndex(path, ".")]
	}
	return path
}

func nodeValueString(n ast.Node) string {
	switch n.Type() {
	case ast.MappingType, ast.SequenceType:
		indent := n.GetToken().Position.IndentNum
		s := n.String()
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			lines[i] = fmt.Sprintf("  %s", line[indent:])
		}
		s = strings.Join(lines, "\n")
		return fmt.Sprintf("\n%s", s)
	default:
		return n.String()
	}
}

func nodeMetadata(n ast.Node) string {
	return fmt.Sprintf("[line:%d <%s>]", n.GetToken().Position.Line, n.Type())
}
