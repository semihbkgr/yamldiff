package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
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
		return colorize(s), false
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

	return colorize(strings.Join(lines, "\n")), true
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

var p printer.Printer

func init() {
	p = printer.Printer{}
	//p.LineNumber = true
	//p.LineNumberFormat = func(num int) string {
	//	fn := color.New(color.Bold, color.FgHiWhite).SprintFunc()
	//	return fn(fmt.Sprintf("%2d | ", num))
	//}
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiMagenta),
			Suffix: format(color.Reset),
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiMagenta),
			Suffix: format(color.Reset),
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiCyan),
			Suffix: format(color.Reset),
		}
	}
	p.Anchor = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiYellow),
			Suffix: format(color.Reset),
		}
	}
	p.Alias = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiYellow),
			Suffix: format(color.Reset),
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiGreen),
			Suffix: format(color.Reset),
		}
	}
}

const escape = "\x1b"

func format(attr color.Attribute) string {
	return fmt.Sprintf("%s[%dm", escape, attr)
}

func colorize(s string) string {
	tokens := lexer.Tokenize(s)
	return p.PrintTokens(tokens)
}
