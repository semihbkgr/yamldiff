//go:build !js || !wasm

package diff

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
)

// colorWrapper wraps fatih/color.Color for CLI usage
type colorWrapper struct {
	c *color.Color
}

func (cw *colorWrapper) Sprint(a ...interface{}) string {
	return cw.c.Sprint(a...)
}

func (cw *colorWrapper) EnableColor() {
	cw.c.EnableColor()
}

func (cw *colorWrapper) DisableColor() {
	cw.c.DisableColor()
}

// newColor creates a new colorWrapper with the given attributes
func newColorAdded() colorSprinter {
	return &colorWrapper{c: color.New(color.FgHiGreen)}
}

func newColorDeleted() colorSprinter {
	return &colorWrapper{c: color.New(color.FgHiRed)}
}

func newColorModified() colorSprinter {
	return &colorWrapper{c: color.New(color.FgHiYellow)}
}

func newColorMetadata() colorSprinter {
	return &colorWrapper{c: color.New(color.FgHiWhite)}
}

// colorPrinter is a global printer instance configured with syntax highlighting
var colorPrinter printer.Printer = initializeColorPrinter()

// initializeColorPrinter creates a printer with syntax highlighting
func initializeColorPrinter() printer.Printer {
	p := printer.Printer{}

	// Configure syntax highlighting colors
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiMagenta),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiMagenta),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiCyan),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.Anchor = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiYellow),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.Alias = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiYellow),
			Suffix: formatColorCode(color.Reset),
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: formatColorCode(color.FgHiGreen),
			Suffix: formatColorCode(color.Reset),
		}
	}

	return p
}

// formatColorCode formats a color attribute into an ANSI escape sequence
func formatColorCode(attr color.Attribute) string {
	return fmt.Sprintf("\x1b[%dm", attr)
}

// colorize applies YAML syntax highlighting to a string
func colorize(s string) string {
	tokens := lexer.Tokenize(s)
	return colorPrinter.PrintTokens(tokens)
}
