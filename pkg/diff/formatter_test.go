package diff

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/stretchr/testify/require"
)

func Test_newFormatter(t *testing.T) {
	tests := []struct {
		name     string
		options  []FormatOption
		expected formatOptions
	}{
		{
			name:     "default options",
			options:  []FormatOption{},
			expected: formatOptions{},
		},
		{
			name: "with plain",
			options: []FormatOption{
				Plain,
			},
			expected: formatOptions{
				plain: true,
			},
		},
		{
			name: "with pathsOnly",
			options: []FormatOption{
				PathsOnly,
			},
			expected: formatOptions{
				pathsOnly: true,
			},
		},
		{
			name: "with metadata",
			options: []FormatOption{
				WithMetadata,
			},
			expected: formatOptions{
				metadata: true,
			},
		},
		{
			name: "with counts",
			options: []FormatOption{
				IncludeCounts,
			},
			expected: formatOptions{
				counts: true,
			},
		},
		{
			name: "with all options",
			options: []FormatOption{
				Plain,
				PathsOnly,
				WithMetadata,
				IncludeCounts,
			},
			expected: formatOptions{
				plain:     true,
				pathsOnly: true,
				metadata:  true,
				counts:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := newFormatter(tt.options...)
			require.Equal(t, tt.expected, formatter.options, "Formatter options do not match expected values")
		})
	}
}

func Test_formatDiff(t *testing.T) {
	tests := []struct {
		name      string
		leftYaml  string
		rightYaml string
		options   []FormatOption
		expected  string
	}{
		{
			name:      "added whole document",
			leftYaml:  "",
			rightYaml: "name: Alice",
			options:   []FormatOption{Plain},
			expected:  "+ name: Alice",
		},
		{
			name:      "deleted whole document",
			leftYaml:  "name: Alice",
			rightYaml: "",
			options:   []FormatOption{Plain},
			expected:  "- name: Alice",
		},
		{
			name:      "modified field",
			leftYaml:  "name: Alice",
			rightYaml: "name: Bob",
			options:   []FormatOption{Plain},
			expected:  "~ .name: Alice → Bob",
		},
		{
			name:     "added field",
			leftYaml: "age: 30",
			rightYaml: heredoc.Doc(`
				age: 30
				name: Alice
			`),
			options:  []FormatOption{Plain},
			expected: "+ .name: Alice",
		},
		{
			name:     "added field with paths only",
			leftYaml: "age: 30",
			rightYaml: heredoc.Doc(`
				age: 30
				name: Alice
			`),
			options:  []FormatOption{Plain, PathsOnly},
			expected: "+ .name",
		},
		{
			name:      "modified field with metadata",
			leftYaml:  "name: Alice",
			rightYaml: "name: Bob",
			options:   []FormatOption{Plain, WithMetadata},
			expected:  "~ .name: [line:1 <String>] Alice → [line:1 <String>] Bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := createRealisticDiff(t, tt.leftYaml, tt.rightYaml)
			formatter := newFormatter(tt.options...)
			result := formatter.formatDiff(diff)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatDocDiffs(t *testing.T) {
	tests := []struct {
		name     string
		options  []FormatOption
		setupFn  func() DocDiffs
		expected string
	}{
		{
			name:    "multiple diffs",
			options: []FormatOption{Plain},
			setupFn: func() DocDiffs {
				diff1 := createRealisticDiff(t, "name: Alice", "name: Bob")
				diff2 := createRealisticDiff(t, "age: 30", heredoc.Doc(`
					age: 30
					city: NYC
				`))
				return DocDiffs{diff1, diff2}
			},
			expected: "~ .name: Alice → Bob\n+ .city: NYC",
		},
		{
			name:    "multiple diffs with counts",
			options: []FormatOption{Plain, IncludeCounts},
			setupFn: func() DocDiffs {
				diff1 := createRealisticDiff(t, "name: Alice", "name: Bob")
				diff2 := createRealisticDiff(t, "age: 30", heredoc.Doc(`
					age: 30
					city: NYC
				`))
				diff3 := createRealisticDiff(t, "status: active", "")
				return DocDiffs{diff1, diff2, diff3}
			},
			expected: "1 added, 1 deleted, 1 modified\n~ .name: Alice → Bob\n+ .city: NYC\n- status: active",
		},
		{
			name:    "empty diffs",
			options: []FormatOption{Plain},
			setupFn: func() DocDiffs {
				return DocDiffs{}
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docDiffs := tt.setupFn()
			formatter := newFormatter(tt.options...)
			result := formatter.formatDocDiffs(docDiffs)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatFileDiffs(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func() FileDiffs
		options  []FormatOption
		expected string
	}{
		{
			name: "single document",
			setupFn: func() FileDiffs {
				diff1 := createRealisticDiff(t, "name: Alice", "name: Bob")
				diff2 := createRealisticDiff(t, "age: 30", heredoc.Doc(`
					age: 30
					city: NYC
				`))
				return FileDiffs{DocDiffs{diff1, diff2}}
			},
			options:  []FormatOption{Plain},
			expected: "~ .name: Alice → Bob\n+ .city: NYC",
		},
		{
			name: "multiple documents",
			setupFn: func() FileDiffs {
				diff1 := createRealisticDiff(t, "name: Alice", "name: Bob")
				diff2 := createRealisticDiff(t, "age: 30", heredoc.Doc(`
					age: 30
					city: NYC
				`))
				return FileDiffs{DocDiffs{diff1}, DocDiffs{diff2}}
			},
			options:  []FormatOption{Plain},
			expected: "~ .name: Alice → Bob\n---\n+ .city: NYC",
		},
		{
			name: "empty file diffs",
			setupFn: func() FileDiffs {
				return FileDiffs{}
			},
			options:  []FormatOption{Plain},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileDiffs := tt.setupFn()
			formatter := newFormatter(tt.options...)
			result := formatter.formatFileDiffs(fileDiffs)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_countDiffs(t *testing.T) {
	tests := []struct {
		name     string
		setupFn  func() DocDiffs
		expected diffCount
	}{
		{
			name: "mixed diff types",
			setupFn: func() DocDiffs {
				diff1 := createRealisticDiff(t, "name: Alice", "name: Bob") // modified
				diff2 := createRealisticDiff(t, "age: 30", heredoc.Doc(`
					age: 30
					city: NYC
				`)) // added
				diff3 := createRealisticDiff(t, "status: active", "") // deleted
				return DocDiffs{diff1, diff2, diff3}
			},
			expected: diffCount{added: 1, deleted: 1, modified: 1},
		},
		{
			name: "empty diffs",
			setupFn: func() DocDiffs {
				return DocDiffs{}
			},
			expected: diffCount{added: 0, deleted: 0, modified: 0},
		},
		{
			name: "only added",
			setupFn: func() DocDiffs {
				diff1 := createRealisticDiff(t, "", "name: Alice")
				diff2 := createRealisticDiff(t, "age: 30", heredoc.Doc(`
					age: 30
					city: NYC
				`))
				return DocDiffs{diff1, diff2}
			},
			expected: diffCount{added: 2, deleted: 0, modified: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docDiffs := tt.setupFn()
			formatter := newFormatter()
			result := formatter.countDiffs(docDiffs)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatValueYaml(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		options  []FormatOption
		expected string
	}{
		{
			name:     "plain mode - no coloring",
			value:    "test: value",
			options:  []FormatOption{Plain},
			expected: "test: value",
		},
		{
			name:     "color mode - with coloring",
			value:    "simple",
			options:  []FormatOption{},
			expected: "simple", // Note: actual coloring would add ANSI codes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := newFormatter(tt.options...)
			result := formatter.formatValueYaml(tt.value)
			// For plain mode, result should match exactly
			if formatter.options.plain {
				require.Equal(t, tt.expected, result)
			} else {
				// For colored mode, just ensure the original text is contained
				require.Contains(t, result, "simple")
			}
		})
	}
}

func Test_formatAdded(t *testing.T) {
	tests := []struct {
		name      string
		rightYaml string
		options   []FormatOption
		expected  string
	}{
		{
			name: "simple field addition",
			rightYaml: heredoc.Doc(`
				age: 30
				name: Alice
			`),
			options:  []FormatOption{Plain},
			expected: "+ .name: Alice",
		},
		{
			name:      "whole document addition",
			rightYaml: "name: Alice",
			options:   []FormatOption{Plain},
			expected:  "+ name: Alice",
		},
		{
			name: "added with paths only",
			rightYaml: heredoc.Doc(`
				age: 30
				name: Alice
			`),
			options:  []FormatOption{Plain, PathsOnly},
			expected: "+ .name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create realistic diff based on the scenario
			var diff *Diff
			if tt.name == "whole document addition" {
				diff = createRealisticDiff(t, "", tt.rightYaml)
			} else {
				diff = createRealisticDiff(t, "age: 30", tt.rightYaml)
			}
			formatter := newFormatter(tt.options...)
			result := formatter.formatAdded(diff)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatDeleted(t *testing.T) {
	tests := []struct {
		name     string
		leftYaml string
		options  []FormatOption
		expected string
	}{
		{
			name: "simple field deletion",
			leftYaml: heredoc.Doc(`
				age: 30
				name: Alice
			`),
			options:  []FormatOption{Plain},
			expected: "- .name: Alice",
		},
		{
			name:     "whole document deletion",
			leftYaml: "name: Alice",
			options:  []FormatOption{Plain},
			expected: "- name: Alice",
		},
		{
			name: "deleted with paths only",
			leftYaml: heredoc.Doc(`
				age: 30
				name: Alice
			`),
			options:  []FormatOption{Plain, PathsOnly},
			expected: "- .name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create realistic diff based on the scenario
			var diff *Diff
			if tt.name == "whole document deletion" {
				diff = createRealisticDiff(t, tt.leftYaml, "")
			} else {
				diff = createRealisticDiff(t, tt.leftYaml, "age: 30")
			}
			formatter := newFormatter(tt.options...)
			result := formatter.formatDeleted(diff)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatModified(t *testing.T) {
	tests := []struct {
		name      string
		leftYaml  string
		rightYaml string
		options   []FormatOption
		expected  string
	}{
		{
			name:      "simple modified value",
			leftYaml:  "name: Alice",
			rightYaml: "name: Bob",
			options:   []FormatOption{Plain},
			expected:  "~ .name: Alice → Bob",
		},
		{
			name:      "modified with paths only",
			leftYaml:  "name: Alice",
			rightYaml: "name: Bob",
			options:   []FormatOption{Plain, PathsOnly},
			expected:  "~ .name",
		},
		{
			name:      "modified with metadata",
			leftYaml:  "name: Alice",
			rightYaml: "name: Bob",
			options:   []FormatOption{Plain, WithMetadata},
			expected:  "~ .name: [line:1 <String>] Alice → [line:1 <String>] Bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := createRealisticDiff(t, tt.leftYaml, tt.rightYaml)
			formatter := newFormatter(tt.options...)
			result := formatter.formatModified(diff)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_getModifiedSymbol(t *testing.T) {
	tests := []struct {
		name           string
		leftMultiLine  bool
		rightMultiLine bool
		path           string
		leftValue      string
		rightValue     string
		expectedSymbol string
		expectedLeft   string
		expectedRight  string
	}{
		{
			name:           "single line values",
			leftMultiLine:  false,
			rightMultiLine: false,
			path:           ".name",
			leftValue:      "Alice",
			rightValue:     "Bob",
			expectedSymbol: "→",
			expectedLeft:   "Alice",
			expectedRight:  "Bob",
		},
		{
			name:           "left multiline, right single line with path",
			leftMultiLine:  true,
			rightMultiLine: false,
			path:           ".config",
			leftValue: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
			rightValue:     "simple",
			expectedSymbol: "\n    ↓",
			expectedLeft: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
			expectedRight: "\n    simple",
		},
		{
			name:           "left single line, right multiline with path",
			leftMultiLine:  false,
			rightMultiLine: true,
			path:           ".config",
			leftValue:      "simple",
			rightValue: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
			expectedSymbol: "\n    ↓",
			expectedLeft:   "\n    simple",
			expectedRight: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
		},
		{
			name:           "both multiline with path",
			leftMultiLine:  true,
			rightMultiLine: true,
			path:           ".config",
			leftValue: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
			rightValue: heredoc.Doc(`
				debug: false
				retries: 3
			`),
			expectedSymbol: "\n    ↓",
			expectedLeft: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
			expectedRight: heredoc.Doc(`
				debug: false
				retries: 3
			`),
		},
		{
			name:           "left single line, right multiline without path",
			leftMultiLine:  false,
			rightMultiLine: true,
			path:           "",
			leftValue:      "simple",
			rightValue: heredoc.Doc(`
				debug: true
				timeout: 100
			`),
			expectedSymbol: "\n    ↓",
			expectedLeft:   "simple",
			expectedRight:  "\n  debug: true\ntimeout: 100\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := newFormatter()
			leftValue := tt.leftValue
			rightValue := tt.rightValue
			symbol := formatter.getModifiedSymbol(tt.leftMultiLine, tt.rightMultiLine, tt.path, &leftValue, &rightValue)
			require.Equal(t, tt.expectedSymbol, symbol)
			require.Equal(t, tt.expectedLeft, leftValue)
			require.Equal(t, tt.expectedRight, rightValue)
		})
	}
}

func Test_buildOutput(t *testing.T) {
	tests := []struct {
		name     string
		sign     string
		path     string
		value    string
		metadata string
		options  []FormatOption
		expected string
	}{
		{
			name:     "basic output with path",
			sign:     "+",
			path:     ".name",
			value:    "Alice",
			metadata: "[line:1 <String>]",
			options:  []FormatOption{},
			expected: "+ .name: Alice",
		},
		{
			name:     "output with paths only",
			sign:     "+",
			path:     ".name",
			value:    "Alice",
			metadata: "[line:1 <String>]",
			options:  []FormatOption{PathsOnly},
			expected: "+ .name",
		},
		{
			name:     "output with metadata and path",
			sign:     "+",
			path:     ".name",
			value:    "Alice",
			metadata: "[line:1 <String>]",
			options:  []FormatOption{WithMetadata},
			expected: "+ .name: [line:1 <String>] Alice",
		},
		{
			name:     "output with paths only and ignored metadata",
			sign:     "+",
			path:     ".name",
			value:    "Alice",
			metadata: "[line:1 <String>]",
			options:  []FormatOption{PathsOnly, WithMetadata},
			expected: "+ .name",
		},
		{
			name:     "output without path",
			sign:     "+",
			path:     "",
			value:    "Alice",
			metadata: "[line:1 <String>]",
			options:  []FormatOption{},
			expected: "+ Alice",
		},
		{
			name:     "output without path with metadata",
			sign:     "+",
			path:     "",
			value:    "Alice",
			metadata: "[line:1 <String>]",
			options:  []FormatOption{WithMetadata},
			expected: "+ [line:1 <String>] Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := newFormatter(tt.options...)
			result := formatter.buildOutput(tt.sign, tt.path, tt.value, tt.metadata)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_buildModifiedOutput(t *testing.T) {
	tests := []struct {
		name          string
		sign          string
		path          string
		leftValue     string
		rightValue    string
		symbol        string
		leftMetadata  string
		rightMetadata string
		options       []FormatOption
		expected      string
	}{
		{
			name:          "basic modified output with path",
			sign:          "~",
			path:          ".name",
			leftValue:     "Alice",
			rightValue:    "Bob",
			symbol:        "→",
			leftMetadata:  "[line:1 <String>]",
			rightMetadata: "[line:1 <String>]",
			options:       []FormatOption{},
			expected:      "~ .name: Alice → Bob",
		},
		{
			name:          "modified output with paths only",
			sign:          "~",
			path:          ".name",
			leftValue:     "Alice",
			rightValue:    "Bob",
			symbol:        "→",
			leftMetadata:  "[line:1 <String>]",
			rightMetadata: "[line:1 <String>]",
			options:       []FormatOption{PathsOnly},
			expected:      "~ .name",
		},
		{
			name:          "modified output with metadata and path",
			sign:          "~",
			path:          ".name",
			leftValue:     "Alice",
			rightValue:    "Bob",
			symbol:        "→",
			leftMetadata:  "[line:1 <String>]",
			rightMetadata: "[line:1 <String>]",
			options:       []FormatOption{WithMetadata},
			expected:      "~ .name: [line:1 <String>] Alice → [line:1 <String>] Bob",
		},
		{
			name:          "modified output without path",
			sign:          "~",
			path:          "",
			leftValue:     "Alice",
			rightValue:    "Bob",
			symbol:        "→",
			leftMetadata:  "[line:1 <String>]",
			rightMetadata: "[line:1 <String>]",
			options:       []FormatOption{},
			expected:      "~ Alice → Bob",
		},
		{
			name:          "modified output without path with metadata",
			sign:          "~",
			path:          "",
			leftValue:     "Alice",
			rightValue:    "Bob",
			symbol:        "→",
			leftMetadata:  "[line:1 <String>]",
			rightMetadata: "[line:1 <String>]",
			options:       []FormatOption{WithMetadata},
			expected:      "~ [line:1 <String>] Alice → [line:1 <String>] Bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := newFormatter(tt.options...)
			result := formatter.buildModifiedOutput(tt.sign, tt.path, tt.leftValue, tt.rightValue, tt.symbol, tt.leftMetadata, tt.rightMetadata)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_diffCount_String(t *testing.T) {
	tests := []struct {
		name     string
		count    diffCount
		expected string
	}{
		{
			name:     "all zeros",
			count:    diffCount{added: 0, deleted: 0, modified: 0},
			expected: "0 added, 0 deleted, 0 modified\n",
		},
		{
			name:     "mixed counts",
			count:    diffCount{added: 2, deleted: 1, modified: 3},
			expected: "2 added, 1 deleted, 3 modified\n",
		},
		{
			name:     "only added",
			count:    diffCount{added: 5, deleted: 0, modified: 0},
			expected: "5 added, 0 deleted, 0 modified\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.count.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_calculateIndentLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "no indentation",
			input:    "text",
			expected: 0,
		},
		{
			name:     "two spaces",
			input:    "  text",
			expected: 2,
		},
		{
			name:     "four spaces",
			input:    "    text",
			expected: 4,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only spaces",
			input:    "    ",
			expected: 0, // function returns 0 for strings with only spaces
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateIndentLevel(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create diffs using the actual Compare function for realistic testing
func createRealisticDiff(t *testing.T, leftYaml, rightYaml string) *Diff {
	t.Helper()
	var left, right []byte
	if leftYaml != "" {
		left = []byte(leftYaml)
	}
	if rightYaml != "" {
		right = []byte(rightYaml)
	}

	fileDiffs, err := Compare(left, right)
	require.NoError(t, err)
	require.NotEmpty(t, fileDiffs)
	require.NotEmpty(t, fileDiffs[0])
	return fileDiffs[0][0]
}
