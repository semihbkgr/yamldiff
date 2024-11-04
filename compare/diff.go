package compare

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

type DiffContext struct {
	left  *ast.File
	right *ast.File
}

func NewDiffContext(filenameLeft, filenameRight string, comments bool) (*DiffContext, error) {
	var parserMode parser.Mode
	if comments {
		parserMode += parser.ParseComments
	}

	yamlLeft, err := parser.ParseFile(filenameLeft, parserMode)
	if err != nil {
		return nil, err
	}
	yamlRight, err := parser.ParseFile(filenameRight, parserMode)
	if err != nil {
		return nil, err
	}
	return &DiffContext{
		left:  yamlLeft,
		right: yamlRight,
	}, nil
}

func NewDiffContextBytes(left, right []byte, comments bool) (*DiffContext, error) {
	var parserMode parser.Mode
	if comments {
		parserMode += parser.ParseComments
	}

	yamlLeft, err := parser.ParseBytes(left, parserMode)
	if err != nil {
		return nil, err
	}
	yamlRight, err := parser.ParseBytes(right, parserMode)
	if err != nil {
		return nil, err
	}
	return &DiffContext{
		left:  yamlLeft,
		right: yamlRight,
	}, nil
}

func (c *DiffContext) Diffs(conf *DiffOptions) FileDiffs {
	return NewFileDiffs(c.left, c.right, conf)
}

type Diff struct {
	NodeLeft  ast.Node
	NodeRight ast.Node
}

type DocDiffs []*Diff

func NewDocDiffs(ln, rn *ast.DocumentNode, conf *DiffOptions) DocDiffs {
	return compareNodes(ln.Body, rn.Body, conf)
}

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

func compareNodes(leftNode, rightNode ast.Node, conf *DiffOptions) []*Diff {
	if leftNode == nil {
		return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
	}

	if rightNode == nil {
		return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
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
	case ast.BoolType:
		leftBoolNode := leftNode.(*ast.BoolNode)
		rightBoolNode := rightNode.(*ast.BoolNode)
		if leftBoolNode.Value != rightBoolNode.Value {
			return []*Diff{{NodeLeft: leftNode, NodeRight: rightNode}}
		}
	}
	return nil
}

func compareMappingNodes(leftNode, rightNode *ast.MappingNode, conf *DiffOptions) []*Diff {
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
			keyDiffsMap[k] = []*Diff{{NodeLeft: node, NodeRight: nil}}
			continue
		}
		keyDiffsMap[k] = compareNodes(leftValue.Value, rightValue.Value, conf)
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
		keyDiffsMap[k] = []*Diff{{NodeLeft: nil, NodeRight: node}}
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

func compareSequenceNodes(leftNode, rightNode *ast.SequenceNode, conf *DiffOptions) []*Diff {
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

func ignoreIndexes(diffs []*Diff, conf *DiffOptions) []*Diff {
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

type FileDiffs []DocDiffs

func (d FileDiffs) OutputString(opts *OutputOptions) string {
	docDiffsStrings := make([]string, 0, len(d))
	for _, docDiffs := range d {
		docDiffsStrings = append(docDiffsStrings, docDiffs.OutputString(opts))
	}
	return strings.Join(docDiffsStrings, "\n---\n")
}

func (d FileDiffs) HasDifference() bool {
	return len(d) > 0
}

func NewFileDiffs(ln, rn *ast.File, conf *DiffOptions) FileDiffs {
	var docDiffs = make(FileDiffs, max(len(ln.Docs), len(rn.Docs)))
	for i := 0; i < len(docDiffs); i++ {
		var l, r *ast.DocumentNode
		if len(ln.Docs) > i {
			l = ln.Docs[i]
		}
		if len(rn.Docs) > i {
			r = rn.Docs[i]
		}
		docDiff := NewDocDiffs(l, r, conf)
		sort.Sort(docDiff)
		docDiffs[i] = docDiff
	}
	return docDiffs
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
