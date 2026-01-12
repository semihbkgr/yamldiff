//go:build js && wasm

package diff

// noopColor is a no-op color implementation for WASM
type noopColor struct{}

func (c *noopColor) Sprint(a ...interface{}) string {
	if len(a) == 0 {
		return ""
	}
	if s, ok := a[0].(string); ok {
		return s
	}
	return ""
}

func (c *noopColor) EnableColor() {}

func (c *noopColor) DisableColor() {}

// newColorAdded creates a no-op color for WASM
func newColorAdded() colorSprinter {
	return &noopColor{}
}

func newColorDeleted() colorSprinter {
	return &noopColor{}
}

func newColorModified() colorSprinter {
	return &noopColor{}
}

func newColorMetadata() colorSprinter {
	return &noopColor{}
}

// colorize returns the string unchanged in WASM (no syntax highlighting)
func colorize(s string) string {
	return s
}
