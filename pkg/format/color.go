package format

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
)

// Printers for different line statuses (unchanged, added, deleted)
var (
	// printerUnchanged applies syntax highlighting without background
	printerUnchanged = createPrinter(nil)
	// printerAdded applies syntax highlighting with green background
	printerAdded = createPrinter([]color.Attribute{color.BgHiGreen})
	// printerDeleted applies syntax highlighting with red background
	printerDeleted = createPrinter([]color.Attribute{color.BgHiRed})
)

// createPrinter creates a printer with syntax highlighting and optional background
func createPrinter(bgAttrs []color.Attribute) printer.Printer {
	p := printer.Printer{}

	// Helper to create property with foreground color and optional background
	makeProperty := func(fgAttr color.Attribute) *printer.Property {
		return &printer.Property{
			Prefix: formatColorCodes(append(bgAttrs, fgAttr)...),
			Suffix: formatColorCodes(append(bgAttrs, color.FgWhite)...), // Reset to default fg, keep bg
		}
	}

	// Configure syntax highlighting colors
	p.Bool = func() *printer.Property {
		return makeProperty(color.FgHiMagenta)
	}
	p.Number = func() *printer.Property {
		return makeProperty(color.FgHiMagenta)
	}
	p.MapKey = func() *printer.Property {
		return makeProperty(color.FgHiCyan)
	}
	p.Anchor = func() *printer.Property {
		return makeProperty(color.FgHiYellow)
	}
	p.Alias = func() *printer.Property {
		return makeProperty(color.FgHiYellow)
	}
	p.String = func() *printer.Property {
		if len(bgAttrs) > 0 {
			// For added/deleted lines, use a contrasting color for strings
			return makeProperty(color.FgHiYellow)
		}
		return makeProperty(color.FgHiGreen)
	}

	return p
}

// formatColorCodes formats multiple color attributes into an ANSI escape sequence
func formatColorCodes(attrs ...color.Attribute) string {
	if len(attrs) == 0 {
		return ""
	}
	if len(attrs) == 1 {
		return fmt.Sprintf("\x1b[%dm", attrs[0])
	}
	// Combine multiple attributes: \x1b[attr1;attr2;...m
	codes := ""
	for i, attr := range attrs {
		if i > 0 {
			codes += ";"
		}
		codes += fmt.Sprintf("%d", attr)
	}
	return fmt.Sprintf("\x1b[%sm", codes)
}

// colorize applies YAML syntax highlighting to a string
func colorize(s string) string {
	tokens := lexer.Tokenize(s)
	return printerUnchanged.PrintTokens(tokens)
}

// colorizeWithStatus applies YAML syntax highlighting with a background color based on status
func colorizeWithStatus(s string, status int) string {
	tokens := lexer.Tokenize(s)
	switch status {
	case 1: // added
		return printerAdded.PrintTokens(tokens)
	case 2: // deleted
		return printerDeleted.PrintTokens(tokens)
	default: // unchanged
		return printerUnchanged.PrintTokens(tokens)
	}
}
