package diff

import (
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
)

func nodePathString(n ast.Node) string {
	path := n.GetPath()[2:]
	// Path of the MappingNode points to the first key in the map.
	if n.Type() == ast.MappingType {
		path = path[:strings.LastIndex(path, ".")]
	}
	return path
}

func nodeValueString(n ast.Node, indent int) (string, bool) {
	s := n.String()
	lines := strings.Split(s, "\n")
	if len(lines) == 1 {
		return s, false
	}

	nodeIndentLevel := 0
	for i, line := range lines {
		if i == 0 {
			nodeIndentLevel = findIndentLevel(line)
			lines[i] = fmt.Sprintf("%s%s%s", "\n", strings.Repeat(" ", indent), line[nodeIndentLevel:])
		} else {
			lines[i] = fmt.Sprintf("%s%s", strings.Repeat(" ", indent), line[nodeIndentLevel:])
		}
	}

	return strings.Join(lines, "\n"), true
}

func nodeMetadata(n ast.Node) string {
	return fmt.Sprintf("[line:%d <%s>]", n.GetToken().Position.Line, n.Type())
}

func findIndentLevel(s string) int {
	for i, c := range s {
		if c != ' ' {
			return i
		}
	}
	return 0
}
