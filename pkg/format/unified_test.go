package format

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/semihbkgr/yamldiff/pkg/diff"
	"github.com/stretchr/testify/require"
)

func TestUnified_SimpleChange(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Bob")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	require.Contains(t, output, "- name: Alice")
	require.Contains(t, output, "+ name: Bob")
}

func TestUnified_AddedField(t *testing.T) {
	left := []byte(heredoc.Doc(`
		name: Alice
	`))
	right := []byte(heredoc.Doc(`
		name: Alice
		age: 30
	`))

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	require.Contains(t, output, "name: Alice")
	require.Contains(t, output, "+ age: 30")
}

func TestUnified_DeletedField(t *testing.T) {
	left := []byte(heredoc.Doc(`
		name: Alice
		city: NYC
	`))
	right := []byte(heredoc.Doc(`
		name: Alice
	`))

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	require.Contains(t, output, "name: Alice")
	require.Contains(t, output, "- city: NYC")
}

func TestUnified_ModifiedValue(t *testing.T) {
	left := []byte(heredoc.Doc(`
		config:
		  port: 8080
		  host: localhost
	`))
	right := []byte(heredoc.Doc(`
		config:
		  port: 3000
		  host: localhost
	`))

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	require.Contains(t, output, "config:")
	require.Contains(t, output, "- ")
	require.Contains(t, output, "+ ")
	require.Contains(t, output, "host: localhost")
}

func TestUnified_NoDiff(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Alice")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	// When no diffs, should just show unchanged content
	require.Contains(t, output, "name: Alice")
	require.NotContains(t, output, "-")
	require.NotContains(t, output, "+")
}

func TestUnified_MultiDocument(t *testing.T) {
	left := []byte("name: Alice\n---\nage: 30")
	right := []byte("name: Bob\n---\nage: 40")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	require.Contains(t, output, "---")
}

func TestUnified_PlainMode(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Bob")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := Unified(result, UnifiedPlain)
	// Plain mode should use +/- prefixes
	require.Contains(t, output, "- ")
	require.Contains(t, output, "+ ")
}

func TestUnified_ColorMode(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Bob")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	// Without UnifiedPlain, should use colors (contains ANSI codes)
	output := Unified(result)
	// Color mode uses background colors, not +/- prefixes
	// The output will contain ANSI escape codes
	require.NotEmpty(t, output)
}
