package diff

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type DiffContext struct {
	left  *ast.File
	right *ast.File
}

func NewDiffContext(filenameLeft, filenameRight string) (*DiffContext, error) {
	yamlLeft, err := parser.ParseFile(filenameLeft, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	yamlRight, err := parser.ParseFile(filenameRight, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &DiffContext{
		left:  yamlLeft,
		right: yamlRight,
	}, nil
}

func (c *DiffContext) Diffs(conf *DiffConfig) FileDiffs {
	return NewFileDiffs(c.left, c.right, conf)
}

type Diff struct {
	NodeLeft  ast.Node
	NodeRight ast.Node
}

type DocDiffs []*Diff

func NewDocDiffs(ln, rn *ast.DocumentNode, conf *DiffConfig) DocDiffs {
	return compareNodes(ln.Body, rn.Body, conf)
}

func compareNodes(leftNode, rightNode ast.Node, conf *DiffConfig) []*Diff {
	// TODO: return diffs for all children nodes
	if leftNode == nil {
		return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
	}

	// TODO: return diffs for all children nodes
	if rightNode == nil {
		return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
	}

	if leftNode.Type() != rightNode.Type() {
		return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
	}

	switch leftNode.Type() {
	case ast.MappingType:
		return compareMappingNodes(leftNode.(*ast.MappingNode), rightNode.(*ast.MappingNode), conf)
	case ast.SequenceType:
		return compareSequenceNodes(leftNode.(*ast.SequenceNode), rightNode.(*ast.SequenceNode), conf)
	case ast.StringType:
		leftStringNode := leftNode.(*ast.StringNode)
		rightStringNode := rightNode.(*ast.StringNode)
		if leftStringNode.Value != rightStringNode.Value {
			return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
		}
	case ast.IntegerType:
		leftIntegerNode := leftNode.(*ast.IntegerNode)
		rightIntegerNode := rightNode.(*ast.IntegerNode)
		if leftIntegerNode.Value != rightIntegerNode.Value {
			return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
		}
	case ast.FloatType:
		leftFloatNode := leftNode.(*ast.FloatNode)
		rightFloatNode := rightNode.(*ast.FloatNode)
		if leftFloatNode.Value != rightFloatNode.Value {
			return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
		}
	case ast.MappingValueType: // when MappingNode's value size is one
		leftMappingValueNode := leftNode.(*ast.MappingValueNode)
		rightMappingValueNode := rightNode.(*ast.MappingValueNode)
		if leftMappingValueNode.Key.String() != rightMappingValueNode.Key.String() {
			// TODO: return diffs for all children nodes
			return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
		}
		return compareNodes(leftMappingValueNode.Value, rightMappingValueNode.Value, conf)
	}
	return nil
}

func compareMappingNodes(leftNode, rightNode *ast.MappingNode, conf *DiffConfig) []*Diff {
	leftKeyValueMap := mappingValueNodesIntoMap(leftNode)
	rightKeyValueMap := mappingValueNodesIntoMap(rightNode)
	keyDiffsMap := make(map[string][]*Diff)
	for k, leftValue := range leftKeyValueMap {
		rightValue, ok := rightKeyValueMap[k]
		if !ok {
			// TODO: return diffs for all children nodes
			keyDiffsMap[k] = []*Diff{{NodeLeft: leftValue, NodeRight: nil}}
			continue
		}
		keyDiffsMap[k] = compareNodes(leftValue.Value, rightValue.Value, conf)
	}
	for k, rightValue := range rightKeyValueMap {
		_, ok := keyDiffsMap[k]
		if ok {
			continue
		}
		// TODO: return diffs for all children nodes
		keyDiffsMap[k] = []*Diff{{NodeLeft: nil, NodeRight: rightValue}}
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

func compareSequenceNodes(leftNode, rightNode *ast.SequenceNode, conf *DiffConfig) []*Diff {
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
		diffs = append(diffs, compareNodes(leftValue, rightValue, conf)...)
	}
	return diffs
}

func (d DocDiffs) String() string {
	b := strings.Builder{}
	for _, diff := range d {
		if diff.NodeLeft == nil {
			b.WriteString(fmt.Sprintf("Add: %s", diff.NodeRight.GetPath()))
		} else if diff.NodeRight == nil {
			b.WriteString(fmt.Sprintf("Delete: %s", diff.NodeLeft.GetPath()))
		} else {
			b.WriteString(fmt.Sprintf("Update: %s %s=>%s", diff.NodeLeft.GetPath(), diff.NodeLeft.String(), diff.NodeRight.String()))
		}
		b.WriteRune('\n')
	}
	return b.String()
}

type FileDiffs []DocDiffs

func (d FileDiffs) String() string {
	docDiffsStrings := make([]string, 0, len(d))
	for _, docDiffs := range d {
		docDiffsStrings = append(docDiffsStrings, docDiffs.String())
	}
	return strings.Join(docDiffsStrings, "\n---\n")
}

func (d FileDiffs) HasDifference() bool {
	return len(d) > 0
}

func NewFileDiffs(ln, rn *ast.File, conf *DiffConfig) FileDiffs {
	var docDiffs = make(FileDiffs, max(len(ln.Docs), len(rn.Docs)))
	for i := 0; i < len(docDiffs); i++ {
		var l, r *ast.DocumentNode
		if len(ln.Docs) > i {
			l = ln.Docs[i]
		}
		if len(rn.Docs) > i {
			r = rn.Docs[i]
		}
		docDiffs[i] = NewDocDiffs(l, r, conf)
	}
	return docDiffs
}

type DiffConfig struct {
	IgnoreIndex bool
}

var DefaultDiffConfig = &DiffConfig{
	IgnoreIndex: false,
}
