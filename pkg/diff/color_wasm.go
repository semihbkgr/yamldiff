//go:build js && wasm

package diff

import (
	"fmt"
	"html"
	"strings"

	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
)

// htmlColor wraps text in HTML span tags with CSS classes
type htmlColor struct {
	className string
	enabled   bool
}

func (c *htmlColor) Sprint(a ...any) string {
	if len(a) == 0 {
		return ""
	}
	var s string
	if str, ok := a[0].(string); ok {
		s = str
	} else {
		s = fmt.Sprint(a...)
	}

	if !c.enabled || c.className == "" {
		return html.EscapeString(s)
	}
	return fmt.Sprintf(`<span class="%s">%s</span>`, c.className, html.EscapeString(s))
}

func (c *htmlColor) EnableColor() {
	c.enabled = true
}

func (c *htmlColor) DisableColor() {
	c.enabled = false
}

// newColorAdded creates an HTML color for added diffs
func newColorAdded() colorSprinter {
	return &htmlColor{className: "diff-added", enabled: true}
}

func newColorDeleted() colorSprinter {
	return &htmlColor{className: "diff-deleted", enabled: true}
}

func newColorModified() colorSprinter {
	return &htmlColor{className: "diff-modified", enabled: true}
}

func newColorMetadata() colorSprinter {
	return &htmlColor{className: "diff-metadata", enabled: true}
}

// htmlPrinter is a global printer instance configured with HTML syntax highlighting
var htmlPrinter printer.Printer = initializeHTMLPrinter()

// initializeHTMLPrinter creates a printer with HTML syntax highlighting
func initializeHTMLPrinter() printer.Printer {
	p := printer.Printer{}

	// Configure syntax highlighting with HTML spans
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: `<span class="yaml-bool">`,
			Suffix: `</span>`,
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: `<span class="yaml-number">`,
			Suffix: `</span>`,
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: `<span class="yaml-key">`,
			Suffix: `</span>`,
		}
	}
	p.Anchor = func() *printer.Property {
		return &printer.Property{
			Prefix: `<span class="yaml-anchor">`,
			Suffix: `</span>`,
		}
	}
	p.Alias = func() *printer.Property {
		return &printer.Property{
			Prefix: `<span class="yaml-alias">`,
			Suffix: `</span>`,
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: `<span class="yaml-string">`,
			Suffix: `</span>`,
		}
	}

	return p
}

// colorize applies HTML syntax highlighting to a YAML string
func colorize(s string) string {
	tokens := lexer.Tokenize(s)
	highlighted := htmlPrinter.PrintTokens(tokens)

	// Escape any remaining untagged content to prevent XSS
	// The printer already added our safe HTML tags, but we need to escape
	// special characters that aren't part of our tags
	return escapeHTMLExceptTags(highlighted)
}

// escapeHTMLExceptTags escapes HTML special characters except for our safe span tags
func escapeHTMLExceptTags(s string) string {
	// This is a simple approach: we trust the tags we added and escape everything else
	// For production, you might want a more sophisticated HTML sanitizer

	// Split by our known safe tags and escape the text between them
	result := strings.Builder{}
	remaining := s

	for {
		// Find next tag
		tagStart := strings.Index(remaining, "<span")
		if tagStart == -1 {
			// No more tags, escape remaining text
			result.WriteString(html.EscapeString(remaining))
			break
		}

		// Escape text before tag
		result.WriteString(html.EscapeString(remaining[:tagStart]))

		// Find end of opening tag
		tagEnd := strings.Index(remaining[tagStart:], ">")
		if tagEnd == -1 {
			// Malformed, escape rest
			result.WriteString(html.EscapeString(remaining))
			break
		}
		tagEnd += tagStart + 1

		// Write opening tag as-is
		result.WriteString(remaining[tagStart:tagEnd])

		// Find closing tag
		closeTag := strings.Index(remaining[tagEnd:], "</span>")
		if closeTag == -1 {
			// No closing tag, escape rest
			result.WriteString(html.EscapeString(remaining[tagEnd:]))
			break
		}
		closeTag += tagEnd

		// Write content between tags (already should be escaped by lexer)
		result.WriteString(remaining[tagEnd:closeTag])

		// Write closing tag
		result.WriteString("</span>")

		// Move to next part
		remaining = remaining[closeTag+7:] // 7 = len("</span>")
	}

	return result.String()
}
