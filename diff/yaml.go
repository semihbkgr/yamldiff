package diff

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

//todo: sort output diffs
//todo: fix diff node type

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

func NewDiffContextBytes(left, right []byte) (*DiffContext, error) {
	yamlLeft, err := parser.ParseBytes(left, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	yamlRight, err := parser.ParseBytes(right, parser.ParseComments)
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
	if leftNode == nil {
		return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
	}

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

	if conf.IgnoreIndex {
		diffs = ignoreIndexes(diffs, conf)
	}

	return diffs
}

func ignoreIndexes(diffs []*Diff, conf *DiffConfig) []*Diff {
	leftNodes := make([]ast.Node, len(diffs))
	rightNodes := make([]ast.Node, len(diffs))
	for i, diff := range diffs {
		leftNodes[i] = diff.NodeLeft
		rightNodes[i] = diff.NodeRight
	}

	for il, leftNode := range leftNodes {
		if leftNode == nil {
			continue
		}
		for ir, rightNode := range rightNodes {
			if rightNode == nil {
				continue
			}
			if len(compareNodes(leftNode, rightNode, conf)) == 0 {
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

func (d DocDiffs) String() string {
	b := strings.Builder{}
	for _, diff := range d {
		if diff.NodeLeft == nil { // Added
			path := nodePathString(diff.NodeRight)
			nodeType := diff.NodeRight.Type()
			value := diff.NodeRight.String()
			b.WriteString(fmt.Sprintf("+ %s: <%s> %s", path, nodeType, value))
		} else if diff.NodeRight == nil { //Deleted
			path := nodePathString(diff.NodeLeft)
			nodeType := diff.NodeLeft.Type()
			value := diff.NodeLeft.String()
			b.WriteString(fmt.Sprintf("- %s: <%s> %s", path, nodeType, value))
		} else { //Modified
			path := nodePathString(diff.NodeLeft)
			leftValue := diff.NodeLeft.String()
			rightValue := diff.NodeRight.String()
			b.WriteString(fmt.Sprintf("~ %s: <%s> %s -> <%s> %s", path, diff.NodeLeft.Type(), leftValue, diff.NodeRight.Type(), rightValue))
		}
		b.WriteRune('\n')
	}
	return b.String()
}

func nodePathString(n ast.Node) string {
	return n.GetPath()[2:]
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
