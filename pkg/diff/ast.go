package diff

import (
	"github.com/goccy/go-yaml/ast"
)

func compareNodes(ln, rn ast.Node, opts CompareOptions) []*Diff {
	if ln == nil || rn == nil || ln.Type() != rn.Type() {
		return []*Diff{
			{
				leftNode:  ln,
				rightNode: rn,
			},
		}
	}

	//todo: handle all types
	switch ln.Type() {
	case ast.MappingType:
		return compareMappingNodes(ln.(*ast.MappingNode), rn.(*ast.MappingNode), opts)
	case ast.SequenceType:
		return compareSequenceNodes(ln.(*ast.SequenceNode), rn.(*ast.SequenceNode), opts)
	case ast.StringType:
		leftStringNode := ln.(*ast.StringNode)
		rightStringNode := rn.(*ast.StringNode)
		if leftStringNode.Value != rightStringNode.Value {
			return []*Diff{{leftNode: ln, rightNode: rn}}
		}
	case ast.IntegerType:
		leftIntegerNode := ln.(*ast.IntegerNode)
		rightIntegerNode := rn.(*ast.IntegerNode)
		if leftIntegerNode.Value != rightIntegerNode.Value {
			return []*Diff{{leftNode: ln, rightNode: rn}}
		}
	case ast.FloatType:
		leftFloatNode := ln.(*ast.FloatNode)
		rightFloatNode := rn.(*ast.FloatNode)
		if leftFloatNode.Value != rightFloatNode.Value {
			return []*Diff{{leftNode: ln, rightNode: rn}}
		}
	case ast.BoolType:
		leftBoolNode := ln.(*ast.BoolNode)
		rightBoolNode := rn.(*ast.BoolNode)
		if leftBoolNode.Value != rightBoolNode.Value {
			return []*Diff{{leftNode: ln, rightNode: rn}}
		}
	}
	return nil
}

func compareMappingNodes(leftNode, rightNode *ast.MappingNode, opts CompareOptions) []*Diff {
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

func compareSequenceNodes(leftNode, rightNode *ast.SequenceNode, opts CompareOptions) []*Diff {
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

func ignoreIndexes(diffs []*Diff, opts CompareOptions) []*Diff {
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
