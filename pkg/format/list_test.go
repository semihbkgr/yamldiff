package format

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/semihbkgr/yamldiff/pkg/diff"
	"github.com/stretchr/testify/require"
)

func TestList_SimpleChange(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Bob")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain)
	expected := "~ .name: Alice → Bob"
	require.Equal(t, expected, output)
}

func TestList_AddedField(t *testing.T) {
	left := []byte("age: 30")
	right := []byte(heredoc.Doc(`
		age: 30
		name: Alice
	`))

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain)
	expected := "+ .name: Alice"
	require.Equal(t, expected, output)
}

func TestList_DeletedField(t *testing.T) {
	left := []byte(heredoc.Doc(`
		age: 30
		name: Alice
	`))
	right := []byte("age: 30")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain)
	expected := "- .name: Alice"
	require.Equal(t, expected, output)
}

func TestList_PathsOnly(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Bob")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain, PathsOnly)
	expected := "~ .name"
	require.Equal(t, expected, output)
}

func TestList_WithMetadata(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Bob")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain, WithMetadata)
	expected := "~ .name: [line:1 <String>] Alice → [line:1 <String>] Bob"
	require.Equal(t, expected, output)
}

func TestList_MultipleChanges(t *testing.T) {
	left := []byte(heredoc.Doc(`
		name: Alice
		city: NYC
	`))
	right := []byte(heredoc.Doc(`
		name: Bob
		age: 30
	`))

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain)
	require.Contains(t, output, "~ .name: Alice → Bob")
	require.Contains(t, output, "- .city: NYC")
	require.Contains(t, output, "+ .age: 30")
}

func TestList_MultiDocument(t *testing.T) {
	left := []byte("name: Alice\n---\nage: 30")
	right := []byte("name: Bob\n---\nage: 40")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain)
	require.Contains(t, output, "~ .name: Alice → Bob")
	require.Contains(t, output, "---")
	require.Contains(t, output, "~ .age: 30 → 40")
}

func TestList_EmptyDiffs(t *testing.T) {
	left := []byte("name: Alice")
	right := []byte("name: Alice")

	result, err := diff.Compare(left, right)
	require.NoError(t, err)

	output := List(result.Diffs, Plain)
	require.Empty(t, output)
}
