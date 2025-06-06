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
	// trim the dolor sign at the beginning of the path
	path := n.GetPath()[1:]
	return path
}

func nodeValueString(n ast.Node, indent int, startNewLine bool) (string, bool) {
	s := n.String()
	lines := strings.Split(s, "\n")
	if len(lines) == 1 && n.Type() != ast.MappingType {
		return s, false
	}

	nodeIndentLevel := findIndentLevel(lines[0])
	for i, line := range lines {
		if i == 0 {
			lines[i] = fmt.Sprintf("%s%s", strings.Repeat(" ", indent), line[nodeIndentLevel:])
		} else {
			if startNewLine {
				lines[i] = fmt.Sprintf("%s%s", strings.Repeat(" ", indent), line[nodeIndentLevel:])
			} else {
				lines[i] = fmt.Sprintf("%s%s%s", "  ", strings.Repeat(" ", indent), line[nodeIndentLevel:])
			}
		}
	}

	builder := strings.Builder{}
	if startNewLine {
		builder.WriteString("\n")
	}
	builder.WriteString(strings.Join(lines, "\n"))
	return builder.String(), true
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

var p printer.Printer = newDefaultPrinter()

func newDefaultPrinter() printer.Printer {
	if color.NoColor {
		// no color mode
		return printer.Printer{}
	}

	p := printer.Printer{}

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
	return p
}

const escape = "\x1b"

func format(attr color.Attribute) string {
	return fmt.Sprintf("%s[%dm", escape, attr)
}

func colorize(s string) string {
	tokens := lexer.Tokenize(s)
	return p.PrintTokens(tokens)
}
