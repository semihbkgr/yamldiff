package diff

import (
	"os"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/stretchr/testify/require"
)

var testFilesDiffPaths = []string{
	".global.scrape_interval",
	".global.scrape_timeout",
	".rule_files[2]",
	".scrape_configs[0].static_configs[0].targets[1]",
	".scrape_configs[1].static_configs[0].labels",
	".scrape_configs[1].static_configs[1]",
	".scrape_configs[2]",
}

var testFilesPathDiffTypes = map[string]DiffType{
	".global.scrape_interval":                         Modified,
	".global.scrape_timeout":                          Added,
	".rule_files[2]":                                  Added,
	".scrape_configs[0].static_configs[0].targets[1]": Added,
	".scrape_configs[1].static_configs[0].labels":     Deleted,
	".scrape_configs[1].static_configs[1]":            Deleted,
	".scrape_configs[2]":                              Added,
}

var testMultiDocsFilesDiffPath = [][]string{
	{
		".properties.role",
		".required[2]",
		".required[3]",
	},
	{
		".properties.name.maxLength",
		".properties.description",
	},
	{
		".properties.timestamp",
		".properties.created_at",
	},
}

var testMultiDocsFilesPathDiffTypes = []map[string]DiffType{
	{
		".properties.role": Added,
		".required[2]":     Added,
		".required[3]":     Added,
	},
	{
		".properties.name.maxLength": Modified,
		".properties.description":    Added,
	},
	{
		".properties.timestamp":  Deleted,
		".properties.created_at": Added,
	},
}

func TestCompare(t *testing.T) {
	left := readFile(t, "testdata/left.yaml")
	right := readFile(t, "testdata/right.yaml")

	diffs, err := Compare(left, right)
	require.NoError(t, err)
	require.Len(t, diffs, 1)
	require.Len(t, diffs[0], len(testFilesDiffPaths))

	for i, diff := range diffs[0] {
		require.Equal(t, diff.Path(), testFilesDiffPaths[i])
		require.Equal(t, diff.Type(), testFilesPathDiffTypes[diff.Path()])
	}
}

func TestCompareMultiDocs(t *testing.T) {
	left := readFile(t, "testdata/multi-docs-left.yaml")
	right := readFile(t, "testdata/multi-docs-right.yaml")

	diffs, err := Compare(left, right)
	require.NoError(t, err)
	require.Len(t, diffs, len(testMultiDocsFilesDiffPath))

	for i, docDiff := range diffs {
		require.Len(t, docDiff, len(testMultiDocsFilesDiffPath[i]))

		for j, diff := range docDiff {
			require.Equal(t, diff.Path(), testMultiDocsFilesDiffPath[i][j])
			require.Equal(t, diff.Type(), testMultiDocsFilesPathDiffTypes[i][diff.Path()])
		}
	}
}

func TestCompareFile(t *testing.T) {
	diffs, err := CompareFile("testdata/left.yaml", "testdata/right.yaml")
	require.NoError(t, err)
	require.Len(t, diffs, 1)
	require.Len(t, diffs[0], len(testFilesDiffPaths))

	for i, diff := range diffs[0] {
		require.Equal(t, diff.Path(), testFilesDiffPaths[i])
		require.Equal(t, diff.Type(), testFilesPathDiffTypes[diff.Path()])
	}
}

func TestCompareFileMultiDocs(t *testing.T) {
	diffs, err := CompareFile("testdata/multi-docs-left.yaml", "testdata/multi-docs-right.yaml")
	require.NoError(t, err)
	require.Len(t, diffs, len(testMultiDocsFilesDiffPath))

	for i, docDiff := range diffs {
		require.Len(t, docDiff, len(testMultiDocsFilesDiffPath[i]))

		for j, diff := range docDiff {
			require.Equal(t, diff.Path(), testMultiDocsFilesDiffPath[i][j])
			require.Equal(t, diff.Type(), testMultiDocsFilesPathDiffTypes[i][diff.Path()])
		}
	}
}

func TestCompareFileNotExist(t *testing.T) {
	_, err := CompareFile("testdata/left.yaml", "testdata/not-exist.yaml")
	require.Error(t, err)
	require.True(t, os.IsNotExist(err), "Expected a file not exist error.")

	_, err = CompareFile("testdata/not-exist.yaml", "testdata/right.yaml")
	require.Error(t, err)
	require.True(t, os.IsNotExist(err), "Expected a file not exist error.")
}

func TestCompareAst(t *testing.T) {
	leftAst, err := parser.ParseFile("testdata/left.yaml", 0)
	require.NoError(t, err)

	rightAst, err := parser.ParseFile("testdata/right.yaml", 0)
	require.NoError(t, err)

	diffs := CompareAst(leftAst, rightAst)
	require.Len(t, diffs, 1)
	require.Len(t, diffs[0], len(testFilesDiffPaths))

	for i, diff := range diffs[0] {
		require.Equal(t, diff.Path(), testFilesDiffPaths[i])
		require.Equal(t, diff.Type(), testFilesPathDiffTypes[diff.Path()])
	}
}

func TestCompareAstMultiDocs(t *testing.T) {
	leftAst, err := parser.ParseFile("testdata/multi-docs-left.yaml", 0)
	require.NoError(t, err)

	rightAst, err := parser.ParseFile("testdata/multi-docs-right.yaml", 0)
	require.NoError(t, err)

	diffs := CompareAst(leftAst, rightAst)
	require.Len(t, diffs, len(testMultiDocsFilesDiffPath))

	for i, docDiff := range diffs {
		require.Len(t, docDiff, len(testMultiDocsFilesDiffPath[i]))

		for j, diff := range docDiff {
			require.Equal(t, diff.Path(), testMultiDocsFilesDiffPath[i][j])
			require.Equal(t, diff.Type(), testMultiDocsFilesPathDiffTypes[i][diff.Path()])
		}
	}
}

func TestCompareEmptyDocument(t *testing.T) {
	tests := []struct {
		name         string
		left         string
		right        string
		expectedDiff bool
		expectedType DiffType
	}{
		{
			name:         "empty and null",
			left:         "",
			right:        "null",
			expectedDiff: true,
			expectedType: Added,
		},
		{
			name:         "null and empty",
			left:         "null",
			right:        "",
			expectedDiff: true,
			expectedType: Deleted,
		},
		{
			name:         "empty and empty",
			left:         "",
			right:        "",
			expectedDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := []byte(tt.left)
			right := []byte(tt.right)

			fileDiffs, err := Compare(left, right)
			require.NoError(t, err)

			require.Len(t, fileDiffs, 1)
			docDiffs := fileDiffs[0]

			if !tt.expectedDiff {
				require.Empty(t, docDiffs)
				return
			}

			require.Len(t, docDiffs, 1)
			diff := docDiffs[0]
			require.Equal(t, diff.Type(), tt.expectedType)
			require.Empty(t, diff.Path())
		})
	}
}

func TestCompareMultiDocsUnmatchedDocumentNumber(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs []map[string]DiffType
	}{
		{
			name: "left has more documents",
			left: heredoc.Doc(`
				foo: bar
				---
				baz: qux
				---
				quux: corge
			`),
			right: heredoc.Doc(`
				foo: bar
				---
				baz: qux
			`),
			expectedDiffs: []map[string]DiffType{
				{},
				{},
				{
					"": Deleted,
				},
			},
		},
		{
			name: "right has more documents",
			left: heredoc.Doc(`
				foo: bar
				---
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: bar
				---
				baz: qux
				---
				quux: corge
			`),
			expectedDiffs: []map[string]DiffType{
				{},
				{},
				{
					"": Added,
				},
			},
		},
		// see: https://github.com/semihbkgr/yamldiff/issues/33
		// {
		// 	name: "empty file and empty documents",
		// 	left: heredoc.Doc(``),
		// 	right: heredoc.Doc(`
		// 		---
		// 		---
		// 		---
		// 	`),
		// 	expectedDiffs: []map[string]DiffType{
		// 		{}, {}, {},
		// 	},
		// },
		// {
		// 	name: "empty documents and empty file",
		// 	left: heredoc.Doc(`
		// 		---
		// 		---
		// 		---
		// 	`),
		// 	right: heredoc.Doc(``),
		// 	expectedDiffs: []map[string]DiffType{
		// 		{}, {}, {},
		// 	},
		// },
		// {
		// 	name: "some empty documents",
		// 	left: heredoc.Doc(`
		// 		foo: bar
		// 		---
		// 		---
		// 		baz: qux
		// 	`),
		// 	right: heredoc.Doc(`
		// 		foo: bar
		// 		---
		// 		baz: qux
		// 		---
		// 	`),
		// 	expectedDiffs: []map[string]DiffType{
		// 		{},
		// 		{"": Added},
		// 		{"": Deleted},
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := []byte(tt.left)
			right := []byte(tt.right)

			fileDiffs, err := Compare(left, right)
			require.NoError(t, err)

			require.Len(t, fileDiffs, len(tt.expectedDiffs))

			for i, docDiffs := range fileDiffs {
				expectedDiffs := tt.expectedDiffs[i]
				diffPathTypeMap := make(map[string]DiffType)
				for _, diff := range docDiffs {
					diffPathTypeMap[diff.Path()] = diff.Type()
				}

				require.Len(t, docDiffs, len(expectedDiffs))
				for path, diffType := range expectedDiffs {
					actualDiffType, exists := diffPathTypeMap[path]
					require.True(t, exists, "expected diff for path %s not found", path)
					require.Equal(t, diffType, actualDiffType, "diff type mismatch for path %s", path)
				}
			}
		})
	}
}

func ExampleCompare() {
	left := []byte(`
name: Alice
city: New York
items:
    - foo
    - bar
`)

	right := []byte(`
name: Bob
age: 30
items:
    - foo
    - baz
`)

	diffs, err := Compare(left, right)
	if err != nil {
		panic(err)
	}

	_ = diffs
}

func Test_compareNodesScalarTypes(t *testing.T) {
	tests := []struct {
		name         string
		left         string
		right        string
		expectedDiff bool
	}{
		{
			name:         "same integers",
			left:         "42",
			right:        "42",
			expectedDiff: false,
		},
		{
			name:         "different integers",
			left:         "42",
			right:        "43",
			expectedDiff: true,
		},
		{
			name:         "same floats",
			left:         "3.14",
			right:        "3.14",
			expectedDiff: false,
		},
		{
			name:         "different floats",
			left:         "3.14",
			right:        "2.71",
			expectedDiff: true,
		},
		{
			name:         "same booleans",
			left:         "true",
			right:        "true",
			expectedDiff: false,
		},
		{
			name:         "different booleans",
			left:         "true",
			right:        "false",
			expectedDiff: true,
		},
		{
			name:         "same strings",
			left:         "foo",
			right:        "foo",
			expectedDiff: false,
		},
		{
			name:         "different strings",
			left:         "foo",
			right:        "bar",
			expectedDiff: true,
		},
		{
			name:         "two null",
			left:         "null",
			right:        "null",
			expectedDiff: false,
		},
		{
			name:         "null and integer",
			left:         "null",
			right:        "42",
			expectedDiff: true,
		},
		{
			name:         "null and empty string",
			left:         "null",
			right:        "\"\"",
			expectedDiff: true,
		},
		{
			name:         "null and boolean",
			left:         "null",
			right:        "true",
			expectedDiff: true,
		},
		{
			name:         "string and boolean",
			left:         "true",
			right:        "\"true\"",
			expectedDiff: true,
		},
		{
			name:         "string and integer",
			left:         "42",
			right:        "\"42\"",
			expectedDiff: true,
		},
		{
			name:         "integer and float",
			left:         "42",
			right:        "42.0",
			expectedDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{})

			if tt.expectedDiff {
				require.Len(t, diffs, 1)
				diff := diffs[0]
				require.Equal(t, diff.Type(), Modified)
				require.Empty(t, diff.Path())
			} else {
				require.Empty(t, diffs)
			}
		})
	}
}

func Test_compareNodesMappingTypes(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs map[string]DiffType
	}{
		{
			name: "same mappings",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			expectedDiffs: nil,
		},
		{
			name: "mappings with modified keys",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: bar
				baz: quux
			`),
			expectedDiffs: map[string]DiffType{
				".baz": Modified,
			},
		},
		{
			name: "mappings with added key",
			left: heredoc.Doc(`
				foo: bar
			`),
			right: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			expectedDiffs: map[string]DiffType{
				".baz": Added,
			},
		},
		{
			name: "mappings with deleted key",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: bar
			`),
			expectedDiffs: map[string]DiffType{
				".baz": Deleted,
			},
		},
		{
			name: "mappings with all types of changes",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: barr
				quux: quux
			`),
			expectedDiffs: map[string]DiffType{
				".foo":  Modified,
				".baz":  Deleted,
				".quux": Added,
			},
		},
		{
			name: "mappings with same nested mappings",
			left: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
			`),
			right: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
			`),
		},
		{
			name: "mappings with modified nested mappings",
			left: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
			`),
			right: heredoc.Doc(`
				foo:
				  bar:
				    baz: quux
			`),
			expectedDiffs: map[string]DiffType{
				".foo.bar.baz": Modified,
			},
		}, {
			name: "mappings with added nested mappings",
			left: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
			`),
			right: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
				    quux: quux
			`),
			expectedDiffs: map[string]DiffType{
				".foo.bar.quux": Added,
			},
		},
		{
			name: "mappings with deleted nested mappings",
			left: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
				    quux: quux
			`),
			right: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
			`),
			expectedDiffs: map[string]DiffType{
				".foo.bar.quux": Deleted,
			},
		},
		{
			name: "mappings with all types of changes in nested mappings",
			left: heredoc.Doc(`
				foo:
				  bar:
				    baz: qux
				    quux: quux
			`),
			right: heredoc.Doc(`
				foo:
				  bar:
				    baz: quux
				    corge: grault
			`),
			expectedDiffs: map[string]DiffType{
				".foo.bar.baz":   Modified,
				".foo.bar.quux":  Deleted,
				".foo.bar.corge": Added,
			},
		},
		{
			name: "mappings with empty value",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: bar
				baz:
			`),
			expectedDiffs: map[string]DiffType{
				".baz": Modified,
			},
		},
		{
			name: "mappings with complex changes",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
				fizz:
				  buzz:
				    quux:
				      corge:
				        alpha: beta
				    blip:
				      drop:
				        grault: garply
			`),
			right: heredoc.Doc(`
				foo: barr
				waz: qux
				fizz:
				  buzz:
				    quux: yay
				    bing:
				      drop:
				        grault: garply
			`),
			expectedDiffs: map[string]DiffType{
				".foo":            Modified,
				".baz":            Deleted,
				".waz":            Added,
				".fizz.buzz.quux": Modified,
				".fizz.buzz.blip": Deleted,
				".fizz.buzz.bing": Added,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{})

			if len(tt.expectedDiffs) == 0 {
				require.Empty(t, diffs)
				return
			}

			diffPathTypeMap := make(map[string]DiffType)
			for _, diff := range diffs {
				diffPathTypeMap[diff.Path()] = diff.Type()
			}

			require.Len(t, diffs, len(tt.expectedDiffs))
			for path, diffType := range tt.expectedDiffs {
				actualDiffType, exists := diffPathTypeMap[path]
				require.True(t, exists, "expected diff for path %s not found", path)
				require.Equal(t, diffType, actualDiffType, "diff type mismatch for path %s", path)
			}
		})
	}
}

func Test_compareNodesSequenceTypes(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs map[string]DiffType
	}{
		{
			name: "same sequences",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			expectedDiffs: nil,
		},
		{
			name: "sequences with modified elements",
			left: heredoc.Doc(`
				- foo
				- barr
				- baz
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- qux
			`),
			expectedDiffs: map[string]DiffType{
				"[1]": Modified,
				"[2]": Modified,
			},
		},
		{
			name: "sequences with added element",
			left: heredoc.Doc(`
				- foo
				- bar
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			expectedDiffs: map[string]DiffType{
				"[2]": Added,
			},
		},
		{
			name: "sequences with deleted element",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- foo
				- bar
			`),
			expectedDiffs: map[string]DiffType{
				"[2]": Deleted,
			},
		},
		{
			name: "sequences with all types of changes",
			left: heredoc.Doc(`
				- foo
				- bar
			`),
			right: heredoc.Doc(`
				- foo
				- fizz
				- buzz
			`),
			expectedDiffs: map[string]DiffType{
				"[1]": Modified,
				"[2]": Added,
			},
		},
		{
			name: "sequences with empty elements",
			left: heredoc.Doc(`
				- foo
				- 
			`),
			right: heredoc.Doc(`
				- 
				- bar
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Modified,
				"[1]": Modified,
			},
		},
		{
			name: "sequences reversed order",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
				- qux
			`),
			right: heredoc.Doc(`
				- qux
				- baz
				- bar
				- foo
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Modified,
				"[1]": Modified,
				"[2]": Modified,
				"[3]": Modified,
			},
		},
		{
			name: "sequences with popped elements",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- bar
				- baz
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Modified,
				"[1]": Modified,
				"[2]": Deleted,
			},
		},
		{
			name: "sequences with pushed elements",
			left: heredoc.Doc(`
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Modified,
				"[1]": Modified,
				"[2]": Added,
			},
		},
		{
			name: "sequences with nested sequences",
			left: heredoc.Doc(`
				- foo
				- bar
				- - baz
				  - qux
				  - corge
				- - drop
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- - baz
				  - grault
				- - drop
				  - corge
			`),
			expectedDiffs: map[string]DiffType{
				"[2][1]": Modified,
				"[2][2]": Deleted,
				"[3][1]": Added,
			},
		},
		{
			name: "sequences with complex changes",
			left: heredoc.Doc(`
				- foo
				- bar
				- - baz
				  - corge
				  - - drop
				    - grault
				    - - blip
				      - alpha
			`),
			right: heredoc.Doc(`
				- foo
				- baz
				- - qux
				  - corge
				  - - drop
				    - bing
				- - blip
				  - alpha
			`),
			expectedDiffs: map[string]DiffType{
				"[1]":       Modified,
				"[2][0]":    Modified,
				"[2][2][1]": Modified,
				"[2][2][2]": Deleted,
				"[3]":       Added,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{})

			if len(tt.expectedDiffs) == 0 {
				require.Empty(t, diffs)
				return
			}

			diffPathTypeMap := make(map[string]DiffType)
			for _, diff := range diffs {
				diffPathTypeMap[diff.Path()] = diff.Type()
			}

			require.Len(t, diffs, len(tt.expectedDiffs))
			for path, diffType := range tt.expectedDiffs {
				actualDiffType, exists := diffPathTypeMap[path]
				require.True(t, exists, "expected diff for path %s not found", path)
				require.Equal(t, diffType, actualDiffType, "diff type mismatch for path %s", path)
			}
		})
	}
}

func Test_compareNodesSequenceTypesIgnoreSeqOrder(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs map[string]DiffType
	}{
		{
			name: "same sequences",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			expectedDiffs: nil,
		},
		{
			name: "sequences with popped elements",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- bar
				- baz
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Deleted,
			},
		},
		{
			name: "sequences with pushed elements",
			left: heredoc.Doc(`
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Added,
			},
		},
		{
			name: "sequences with reversed order",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
				- qux
			`),
			right: heredoc.Doc(`
				- qux
				- baz
				- bar
				- foo
			`),
			expectedDiffs: nil,
		},
		{
			name: "sequences with modified elements",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right: heredoc.Doc(`
				- barr
				- qux
				- foo
			`),
			expectedDiffs: map[string]DiffType{
				"[0]": Added,
				"[1]": Modified,
				"[2]": Deleted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{ignoreSeqOrder: true})

			if len(tt.expectedDiffs) == 0 {
				require.Empty(t, diffs)
				return
			}

			diffPathTypeMap := make(map[string]DiffType)
			for _, diff := range diffs {
				diffPathTypeMap[diff.Path()] = diff.Type()
			}

			require.Len(t, diffs, len(tt.expectedDiffs))
			for path, diffType := range tt.expectedDiffs {
				actualDiffType, exists := diffPathTypeMap[path]
				require.True(t, exists, "expected diff for path %s not found", path)
				require.Equal(t, diffType, actualDiffType, "diff type mismatch for path %s", path)
			}
		})
	}
}

func Test_compareNodesCollectionTypes(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs map[string]DiffType
	}{
		{
			name: "mapping with sequence values - same",
			left: heredoc.Doc(`
				items:
				  - foo
				  - bar
				tags:
				  - alpha
				  - beta
			`),
			right: heredoc.Doc(`
				items:
				  - foo
				  - bar
				tags:
				  - alpha
				  - beta
			`),
			expectedDiffs: nil,
		},
		{
			name: "mapping with sequence values - modified sequence",
			left: heredoc.Doc(`
				items:
				  - foo
				  - bar
				tags:
				  - alpha
				  - beta
			`),
			right: heredoc.Doc(`
				items:
				  - foo
				  - baz
				tags:
				  - alpha
				  - gamma
			`),
			expectedDiffs: map[string]DiffType{
				".items[1]": Modified,
				".tags[1]":  Modified,
			},
		},
		{
			name: "mapping with sequence values - added/deleted elements",
			left: heredoc.Doc(`
				items:
				  - foo
				  - bar
				config:
				  - setting1
			`),
			right: heredoc.Doc(`
				items:
				  - foo
				  - bar
				  - baz
				config: []
			`),
			expectedDiffs: map[string]DiffType{
				".items[2]":  Added,
				".config[0]": Deleted,
			},
		},
		{
			name: "sequence of mappings - same",
			left: heredoc.Doc(`
				- name: alice
				  age: 30
				- name: bob
				  age: 25
			`),
			right: heredoc.Doc(`
				- name: alice
				  age: 30
				- name: bob
				  age: 25
			`),
			expectedDiffs: nil,
		},
		{
			name: "sequence of mappings - modified mapping values",
			left: heredoc.Doc(`
				- name: alice
				  age: 30
				- name: bob
				  age: 25
			`),
			right: heredoc.Doc(`
				- name: alice
				  age: 31
				- name: charlie
				  age: 25
			`),
			expectedDiffs: map[string]DiffType{
				"[0].age":  Modified,
				"[1].name": Modified,
			},
		},
		{
			name: "sequence of mappings - added/deleted keys",
			left: heredoc.Doc(`
				- name: alice
				  age: 30
				  city: nyc
				- name: bob
			`),
			right: heredoc.Doc(`
				- name: alice
				  age: 30
				- name: bob
				  email: bob@example.com
			`),
			expectedDiffs: map[string]DiffType{
				"[0].city":  Deleted,
				"[1].email": Added,
			},
		},
		{
			name: "nested mapping with sequences containing mappings",
			left: heredoc.Doc(`
				users:
				  - name: alice
				    roles:
				      - admin
				      - user
				  - name: bob
				    roles:
				      - user
				groups:
				  admins:
				    - alice
			`),
			right: heredoc.Doc(`
				users:
				  - name: alice
				    roles:
				      - admin
				      - moderator
				  - name: charlie
				    roles:
				      - user
				groups:
				  admins:
				    - alice
				    - charlie
			`),
			expectedDiffs: map[string]DiffType{
				".users[0].roles[1]": Modified,
				".users[1].name":     Modified,
				".groups.admins[1]":  Added,
			},
		},
		{
			name: "complex nested structures - sequences in mappings in sequences",
			left: heredoc.Doc(`
				- config:
				    settings:
				      - debug: true
				        level: info
				      - debug: false
				        level: warn
				  metadata:
				    version: 1.0
				- config:
				    settings:
				      - debug: true
				        level: error
			`),
			right: heredoc.Doc(`
				- config:
				    settings:
				      - debug: false
				        level: info
				      - debug: false
				        level: warn
				        timeout: 30
				  metadata:
				    version: 1.1
				    author: admin
				- config:
				    settings:
				      - debug: true
				        level: error
				    features:
				      - logging
			`),
			expectedDiffs: map[string]DiffType{
				"[0].config.settings[0].debug":   Modified,
				"[0].config.settings[1].timeout": Added,
				"[0].metadata.version":           Modified,
				"[0].metadata.author":            Added,
				"[1].config.features":            Added,
			},
		},
		{
			name: "mapping to sequence type change",
			left: heredoc.Doc(`
				data:
				  key1: value1
				  key2: value2
			`),
			right: heredoc.Doc(`
				data:
				  - value1
				  - value2
			`),
			expectedDiffs: map[string]DiffType{
				".data": Modified,
			},
		},
		{
			name: "sequence to mapping type change",
			left: heredoc.Doc(`
				items:
				  - first
				  - second
			`),
			right: heredoc.Doc(`
				items:
				  first: 1
				  second: 2
			`),
			expectedDiffs: map[string]DiffType{
				".items": Modified,
			},
		},
		{
			name: "empty collections",
			left: heredoc.Doc(`
				empty_map: {}
				empty_list: []
			`),
			right: heredoc.Doc(`
				empty_map: {}
				empty_list: []
			`),
			expectedDiffs: nil,
		},
		{
			name: "empty to populated collections",
			left: heredoc.Doc(`
				data: {}
				items: []
			`),
			right: heredoc.Doc(`
				data:
				  key: value
				items:
				  - item1
			`),
			expectedDiffs: map[string]DiffType{
				".data.key": Added,
				".items[0]": Added,
			},
		},
		{
			name: "mixed collection operations",
			left: heredoc.Doc(`
				config:
				  servers:
				    - name: server1
				      ports: [80, 443]
				    - name: server2
				      ports: [8080]
				  defaults:
				    timeout: 30
			`),
			right: heredoc.Doc(`
				config:
				  servers:
				    - name: server1
				      ports: [80, 443, 8443]
				      ssl: true
				  defaults:
				    timeout: 60
				    retries: 3
				  backup:
				    - server3
			`),
			expectedDiffs: map[string]DiffType{
				".config.servers[0].ports[2]": Added,
				".config.servers[0].ssl":      Added,
				".config.servers[1]":          Deleted,
				".config.defaults.timeout":    Modified,
				".config.defaults.retries":    Added,
				".config.backup":              Added,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{})

			if len(tt.expectedDiffs) == 0 {
				require.Empty(t, diffs)
				return
			}

			diffPathTypeMap := make(map[string]DiffType)
			for _, diff := range diffs {
				diffPathTypeMap[diff.Path()] = diff.Type()
			}

			require.Len(t, diffs, len(tt.expectedDiffs))
			for path, diffType := range tt.expectedDiffs {
				actualDiffType, exists := diffPathTypeMap[path]
				require.True(t, exists, "expected diff for path %s not found", path)
				require.Equal(t, diffType, actualDiffType, "diff type mismatch for path %s", path)
			}
		})
	}
}

func Test_compareNodesNestedCollectionPath(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs map[string]DiffType
	}{
		{
			name: "nested mapping with modified scalar values",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        five: value-1
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        five: value-2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.five": Modified,
			},
		},
		{
			name: "nested mapping with added scalar value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        five: value-1
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        five: value-1
				        six: value-2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.six": Added,
			},
		},
		{
			name: "nested mapping with deleted scalar value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        five: value-1
				        six: value-2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        five: value-1
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.six": Deleted,
			},
		},
		{
			name: "nested mapping with modified scalar sequence value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item3
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[1]": Modified,
			},
		},
		{
			name: "nested mapping with modified mapping sequence value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value3
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[1].value": Modified,
			},
		},
		{
			name: "nested mapping with modified sequence value type from scalar to mapping",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[0]": Modified,
				".one.two.three.four.items[1]": Modified,
			},
		},
		{
			name: "nested mapping with modified sequence value type from mapping to scalar",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[0]": Modified,
				".one.two.three.four.items[1]": Modified,
			},
		},
		{
			name: "nested mapping with added scalar sequence value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
				          - item3
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[2]": Added,
			},
		},
		{
			name: "nested mapping with added mapping sequence value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
				          - name: item3
				            value: value3
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[2]": Added,
			},
		},
		{
			name: "nested mapping with deleted scalar sequence value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
				          - item3
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - item1
				          - item2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[2]": Deleted,
			},
		},
		{
			name: "nested mapping with deleted mapping sequence value",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
				          - name: item3
				            value: value3
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        items:
				          - name: item1
				            value: value1
				          - name: item2
				            value: value2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.items[2]": Deleted,
			},
		},
		{
			name: "nested mapping with modified mapping field name",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item:
				          name: item1
				          value: value1
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        new-item:
				          name: item1
				          value: value1
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.item":     Deleted,
				".one.two.three.four.new-item": Added,
			},
		},
		{
			name: "nested mapping with modified mapping field value type",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item:
				          name: item1
				          value: value1
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item: item1
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.item": Modified,
			},
		},
		{
			name: "nested mapping with added mapping field",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item:
				          name: item1
				          value: value1
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item:
				          name: item1
				          value: value1
				        new-item:
				          name: item2
				          value: value2
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.new-item": Added,
			},
		},
		{
			name: "nested mapping with deleted mapping field",
			left: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item:
				          name: item1
				          value: value1
				        new-item:
				          name: item2
				          value: value2
			`),
			right: heredoc.Doc(`
				one:
				  two:
				    three:
				      four:
				        item:
				          name: item1
				          value: value1
			`),
			expectedDiffs: map[string]DiffType{
				".one.two.three.four.new-item": Deleted,
			},
		},
		{
			name: "nested sequence with modified scalar value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item3
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][0]": Modified,
			},
		},
		{
			name: "nested sequence with added scalar value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - item3
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][1]": Added,
			},
		},
		{
			name: "nested sequence with deleted scalar value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - item3
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][1]": Deleted,
			},
		},
		{
			name: "nested sequence with modified mapping sequence value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item
				          value: value1
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item
				          value: value2
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][1].value": Modified,
			},
		},
		{
			name: "nested sequence with modified value type from scalar to mapping",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - item3
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item3
				          value: value3
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][1]": Modified,
			},
		},
		{
			name: "nested sequence with modified value type from mapping to scalar",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item3
				          value: value3
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - item3
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][1]": Modified,
			},
		},
		{
			name: "nested sequence with added mapping sequence value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item3
				          value: value3
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item3
				          value: value3
				        - name: item4
				          value: value4
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][2]": Added,
			},
		},
		{
			name: "nested sequence with deleted mapping sequence value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item3
				          value: value3
				        - name: item4
				          value: value4
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - name: item3
				          value: value3
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][0][2]": Deleted,
			},
		},
		{
			name: "nested sequence with modified sequence value type",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				        - item3
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - item2
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0]": Modified,
			},
		},
		{
			name: "nested sequence with added sequence value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				      - - - item3
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][1]": Added,
			},
		},
		{
			name: "nested sequence with deleted sequence value",
			left: heredoc.Doc(`
				- - - - item1
				  - - - - item2
				      - - - item3
			`),
			right: heredoc.Doc(`
				- - - - item1
				  - - - - item2
			`),
			expectedDiffs: map[string]DiffType{
				"[0][1][0][1]": Deleted,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{})

			if len(tt.expectedDiffs) == 0 {
				require.Empty(t, diffs)
				return
			}

			diffPathTypeMap := make(map[string]DiffType)
			for _, diff := range diffs {
				diffPathTypeMap[diff.Path()] = diff.Type()
			}

			require.Len(t, diffs, len(tt.expectedDiffs))
			for path, diffType := range tt.expectedDiffs {
				actualDiffType, exists := diffPathTypeMap[path]
				require.True(t, exists, "expected diff for path %s not found", path)
				require.Equal(t, diffType, actualDiffType, "diff type mismatch for path %s", path)
			}
		})
	}
}

func Test_compareNodesScalarAndCollectionTypes(t *testing.T) {
	tests := []struct {
		name         string
		left         string
		right        string
		expectedType DiffType
	}{
		// String to collection types
		{
			name:         "string to mapping",
			left:         heredoc.Doc(`"hello"`),
			right:        "key: value",
			expectedType: Modified,
		},
		{
			name: "string to sequence",
			left: heredoc.Doc(`"hello"`),
			right: heredoc.Doc(`
				- item1
				- item2
			`),
			expectedType: Modified,
		},
		{
			name:         "mapping to string",
			left:         "key: value",
			right:        heredoc.Doc(`"hello"`),
			expectedType: Modified,
		},
		{
			name: "sequence to string",
			left: heredoc.Doc(`
				- item1
				- item2
			`),
			right:        heredoc.Doc(`"hello"`),
			expectedType: Modified,
		},

		// Integer to collection types
		{
			name:         "integer to mapping",
			left:         "42",
			right:        "count: 42",
			expectedType: Modified,
		},
		{
			name: "integer to sequence",
			left: "42",
			right: heredoc.Doc(`
				- 1
				- 2
				- 3
			`),
			expectedType: Modified,
		},
		{
			name:         "mapping to integer",
			left:         "count: 42",
			right:        "42",
			expectedType: Modified,
		},
		{
			name: "sequence to integer",
			left: heredoc.Doc(`
				- 1
				- 2
				- 3
			`),
			right:        "42",
			expectedType: Modified,
		},

		// Float to collection types
		{
			name:         "float to mapping",
			left:         "3.14",
			right:        "pi: 3.14",
			expectedType: Modified,
		},
		{
			name: "float to sequence",
			left: "3.14",
			right: heredoc.Doc(`
				- 1.0
				- 2.0
				- 3.14
			`),
			expectedType: Modified,
		},
		{
			name:         "mapping to float",
			left:         "pi: 3.14",
			right:        "3.14",
			expectedType: Modified,
		},
		{
			name: "sequence to float",
			left: heredoc.Doc(`
				- 1.0
				- 2.0
				- 3.14
			`),
			right:        "3.14",
			expectedType: Modified,
		},

		// Boolean to collection types
		{
			name:         "boolean to mapping",
			left:         "true",
			right:        "enabled: true",
			expectedType: Modified,
		},
		{
			name: "boolean to sequence",
			left: "false",
			right: heredoc.Doc(`
				- true
				- false
			`),
			expectedType: Modified,
		},
		{
			name:         "mapping to boolean",
			left:         "enabled: true",
			right:        "true",
			expectedType: Modified,
		},
		{
			name: "sequence to boolean",
			left: heredoc.Doc(`
				- true
				- false
			`),
			right:        "false",
			expectedType: Modified,
		},

		// Null to collection types
		{
			name:         "null to mapping",
			left:         "null",
			right:        "data: null",
			expectedType: Modified,
		},
		{
			name: "null to sequence",
			left: "null",
			right: heredoc.Doc(`
				- null
				- value
			`),
			expectedType: Modified,
		},
		{
			name:         "mapping to null",
			left:         "data: null",
			right:        "null",
			expectedType: Modified,
		},
		{
			name: "sequence to null",
			left: heredoc.Doc(`
				- null
				- value
			`),
			right:        "null",
			expectedType: Modified,
		},

		// Empty collections to scalars
		{
			name:         "empty mapping to string",
			left:         "{}",
			right:        heredoc.Doc(`"empty"`),
			expectedType: Modified,
		},
		{
			name:         "empty sequence to string",
			left:         "[]",
			right:        heredoc.Doc(`"empty"`),
			expectedType: Modified,
		},
		{
			name:         "string to empty mapping",
			left:         heredoc.Doc(`"empty"`),
			right:        "{}",
			expectedType: Modified,
		},
		{
			name:         "string to empty sequence",
			left:         heredoc.Doc(`"empty"`),
			right:        "[]",
			expectedType: Modified,
		},

		// Complex scenarios with nested structures
		{
			name: "complex mapping to scalar",
			left: heredoc.Doc(`
				config:
				  database:
				    host: localhost
				    port: 5432
				  cache:
				    enabled: true
			`),
			right:        heredoc.Doc(`"configuration disabled"`),
			expectedType: Modified,
		},
		{
			name: "complex sequence to scalar",
			left: heredoc.Doc(`
				- name: server1
				  config:
				    host: localhost
				- name: server2
				  config:
				    host: remote
			`),
			right:        "42",
			expectedType: Modified,
		},
		{
			name: "scalar to complex mapping",
			left: heredoc.Doc(`"simple value"`),
			right: heredoc.Doc(`
				users:
				  - alice
				  - bob
				settings:
				  timeout: 30
			`),
			expectedType: Modified,
		},
		{
			name: "scalar to complex sequence",
			left: "false",
			right: heredoc.Doc(`
				- config:
				    debug: true
				- config:
				    debug: false
			`),
			expectedType: Modified,
		},

		// Edge cases with special values
		{
			name:         "quoted number string to mapping",
			left:         heredoc.Doc(`"123"`),
			right:        "number: 123",
			expectedType: Modified,
		},
		{
			name: "quoted boolean string to sequence",
			left: heredoc.Doc(`"true"`),
			right: heredoc.Doc(`
				- true
				- false
			`),
			expectedType: Modified,
		},
		{
			name: "multiline string to mapping",
			left: heredoc.Doc(`
				"line1
				line2
				line3"
			`),
			right: heredoc.Doc(`
				lines:
				  - line1
				  - line2
				  - line3
			`),
			expectedType: Modified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := parseAstNode(t, tt.left)
			right := parseAstNode(t, tt.right)

			diffs := compareNodes(left, right, &compareOptions{})

			require.Len(t, diffs, 1)
			diff := diffs[0]
			require.Equal(t, tt.expectedType, diff.Type())
			require.Empty(t, diff.Path())
		})
	}
}

func Test_compareNodesNilNodes(t *testing.T) {
	tests := []struct {
		name         string
		left         ast.Node
		right        ast.Node
		expectedDiff bool
		expectedType DiffType
	}{
		{
			name:         "both nil",
			left:         nil,
			right:        nil,
			expectedDiff: false,
		},
		{
			name:         "left nil",
			left:         nil,
			right:        parseAstNode(t, "foo: bar"),
			expectedDiff: true,
			expectedType: Added,
		},
		{
			name:         "right nil",
			left:         parseAstNode(t, "foo: bar"),
			right:        nil,
			expectedDiff: true,
			expectedType: Deleted,
		},
		{
			name:         "nil and null",
			left:         nil,
			right:        parseAstNode(t, "null"),
			expectedDiff: true,
			expectedType: Added,
		},
		{
			name:         "null and nil",
			left:         parseAstNode(t, "null"),
			right:        nil,
			expectedDiff: true,
			expectedType: Deleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := compareNodes(tt.left, tt.right, &compareOptions{})

			if tt.expectedDiff {
				require.Len(t, diffs, 1)
				diff := diffs[0]
				require.Equal(t, diff.Type(), tt.expectedType)
				require.Empty(t, diff.Path())
			} else {
				require.Empty(t, diffs)
			}
		})
	}
}

func Test_comparableNodes(t *testing.T) {
	tests := []struct {
		name       string
		left       string
		right      string
		comparable bool
	}{
		// Same types
		{
			name:       "integer and integer",
			left:       "42",
			right:      "84",
			comparable: true,
		},
		{
			name:       "float and float",
			left:       "3.14",
			right:      "2.71",
			comparable: true,
		},
		{
			name:       "boolean and boolean",
			left:       "true",
			right:      "false",
			comparable: true,
		},
		{
			name:       "null and null",
			left:       "null",
			right:      "null",
			comparable: true,
		},
		{
			name:       "mapping and mapping",
			left:       "foo: bar",
			right:      "baz: qux",
			comparable: true,
		},
		{
			name:       "sequence and sequence",
			left:       "- foo\n- bar",
			right:      "- baz\n- qux",
			comparable: true,
		},
		// String and Literal type compatibility
		{
			name:       "string and string",
			left:       "foo",
			right:      "bar",
			comparable: true,
		},
		{
			name:       "string and literal",
			left:       "foo",
			right:      "|\n  bar",
			comparable: true,
		},
		{
			name:       "literal and string",
			left:       "|\n  foo",
			right:      "bar",
			comparable: true,
		},
		{
			name:       "literal and literal",
			left:       "|\n  foo",
			right:      ">\n  bar",
			comparable: true,
		},
		// Different incompatible types
		{
			name:       "string and integer",
			left:       "foo",
			right:      "42",
			comparable: false,
		},
		{
			name:       "string and float",
			left:       "foo",
			right:      "3.14",
			comparable: false,
		},
		{
			name:       "string and boolean",
			left:       "foo",
			right:      "true",
			comparable: false,
		},
		{
			name:       "string and null",
			left:       "foo",
			right:      "null",
			comparable: false,
		},
		{
			name:       "string and mapping",
			left:       "foo",
			right:      "key: value",
			comparable: false,
		},
		{
			name:       "string and sequence",
			left:       "foo",
			right:      "- item",
			comparable: false,
		},
		{
			name:       "integer and float",
			left:       "42",
			right:      "3.14",
			comparable: false,
		},
		{
			name:       "integer and boolean",
			left:       "42",
			right:      "true",
			comparable: false,
		},
		{
			name:       "integer and null",
			left:       "42",
			right:      "null",
			comparable: false,
		},
		{
			name:       "integer and mapping",
			left:       "42",
			right:      "key: value",
			comparable: false,
		},
		{
			name:       "integer and sequence",
			left:       "42",
			right:      "- item",
			comparable: false,
		},
		{
			name:       "float and boolean",
			left:       "3.14",
			right:      "true",
			comparable: false,
		},
		{
			name:       "float and null",
			left:       "3.14",
			right:      "null",
			comparable: false,
		},
		{
			name:       "float and mapping",
			left:       "3.14",
			right:      "key: value",
			comparable: false,
		},
		{
			name:       "float and sequence",
			left:       "3.14",
			right:      "- item",
			comparable: false,
		},
		{
			name:       "boolean and null",
			left:       "true",
			right:      "null",
			comparable: false,
		},
		{
			name:       "boolean and mapping",
			left:       "true",
			right:      "key: value",
			comparable: false,
		},
		{
			name:       "boolean and sequence",
			left:       "true",
			right:      "- item",
			comparable: false,
		},
		{
			name:       "null and mapping",
			left:       "null",
			right:      "key: value",
			comparable: false,
		},
		{
			name:       "null and sequence",
			left:       "null",
			right:      "- item",
			comparable: false,
		},
		{
			name:       "mapping and sequence",
			left:       "key: value",
			right:      "- item",
			comparable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leftNode := parseAstNode(t, tt.left)
			rightNode := parseAstNode(t, tt.right)

			result := comparableNodes(leftNode, rightNode)
			require.Equal(t, tt.comparable, result, "comparableNodes(%s, %s) = %v, want %v",
				tt.left, tt.right, result, tt.comparable)
		})
	}
}

func Test_stringNodeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string node with simple value",
			input:    `"hello"`,
			expected: "hello",
		},
		{
			name:     "string node with empty value",
			input:    `""`,
			expected: "",
		},
		{
			name:     "string node with special characters",
			input:    `"hello\nworld"`,
			expected: "hello\nworld",
		},
		{
			name:     "string node with spaces",
			input:    `"hello world"`,
			expected: "hello world",
		},
		{
			name:     "unquoted string",
			input:    `hello`,
			expected: "hello",
		},
		{
			name:     "string node with single quotes",
			input:    `'hello'`,
			expected: "hello",
		},
		{
			name: "string node with multi-line content using quotes",
			input: heredoc.Doc(`
				"line1\nline2\nline3"
			`),
			expected: "line1\nline2\nline3",
		},
		{
			name: "string node with yaml special characters",
			input: heredoc.Doc(`
				"key: value, [1, 2, 3]"
			`),
			expected: "key: value, [1, 2, 3]",
		},
		{
			name: "literal node with simple value",
			input: heredoc.Doc(`
				|
				  hello
				  world
			`),
			expected: "hello\nworld\n",
		},
		{
			name: "literal node with folded style",
			input: heredoc.Doc(`
				>
				  hello
				  world
			`),
			expected: "hello world\n",
		},
		{
			name: "literal node with complex content",
			input: heredoc.Doc(`
				|
				  This is a multi-line
				  literal string with
				  preserved line breaks
				  and   spacing
			`),
			expected: "This is a multi-line\nliteral string with\npreserved line breaks\nand   spacing\n",
		},
		{
			name: "literal node with empty content",
			input: heredoc.Doc(`
				|
			`),
			expected: "",
		},
		{
			name: "literal node with keep final newlines (|+)",
			input: heredoc.Doc(`
				|+
				  hello
				  world


			`),
			expected: "hello\nworld\n\n\n",
		},
		{
			name: "literal node with strip final newlines (|-)",
			input: heredoc.Doc(`
				|-
				  hello
				  world


			`),
			expected: "hello\nworld",
		},
		{
			name: "folded node with keep final newlines (>+)",
			input: heredoc.Doc(`
				>+
				  hello
				  world


			`),
			expected: "hello world\n\n\n",
		},
		{
			name: "folded node with strip final newlines (>-)",
			input: heredoc.Doc(`
				>-
				  hello
				  world


			`),
			expected: "hello world",
		},
		{
			name: "literal node with indentation and keep newlines (|+)",
			input: heredoc.Doc(`
				|+
				  line 1
				    indented line 2
				  line 3

			`),
			expected: "line 1\n  indented line 2\nline 3\n\n",
		},
		{
			name: "literal node with indentation and strip newlines (|-)",
			input: heredoc.Doc(`
				|-
				  line 1
				    indented line 2
				  line 3

			`),
			expected: "line 1\n  indented line 2\nline 3",
		},
		{
			name: "folded node with long lines and keep newlines (>+)",
			input: heredoc.Doc(`
				>+
				  This is a very long line that should be folded
				  into a single line with spaces between words.
				  
				  This is a second paragraph.

			`),
			expected: "This is a very long line that should be folded into a single line with spaces between words.\nThis is a second paragraph.\n\n",
		},
		{
			name: "folded node with long lines and strip newlines (>-)",
			input: heredoc.Doc(`
				>-
				  This is a very long line that should be folded
				  into a single line with spaces between words.
				  
				  This is a second paragraph.

			`),
			expected: "This is a very long line that should be folded into a single line with spaces between words.\nThis is a second paragraph.",
		},
		{
			name: "literal node empty with keep newlines (|+)",
			input: heredoc.Doc(`
				|+


			`),
			expected: "\n\n",
		},
		{
			name: "literal node empty with strip newlines (|-)",
			input: heredoc.Doc(`
				|-


			`),
			expected: "",
		},
		{
			name: "folded node empty with keep newlines (>+)",
			input: heredoc.Doc(`
				>+


			`),
			expected: "\n\n",
		},
		{
			name: "folded node empty with strip newlines (>-)",
			input: heredoc.Doc(`
				>-


			`),
			expected: "",
		},
		{
			name:     "integer node returns empty string",
			input:    `42`,
			expected: "",
		},
		{
			name:     "float node returns empty string",
			input:    `3.14`,
			expected: "",
		},
		{
			name:     "boolean node returns empty string",
			input:    `true`,
			expected: "",
		},
		{
			name:     "null node returns empty string",
			input:    `null`,
			expected: "",
		},
		{
			name: "array node returns empty string",
			input: heredoc.Doc(`
				- 1
				- 2
				- 3
			`),
			expected: "",
		},
		{
			name: "object node returns empty string",
			input: heredoc.Doc(`
				key: value
			`),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseAstNode(t, tt.input)
			result := stringNodeValue(node)
			require.Equal(t, tt.expected, result)
		})
	}
}

func parseAstNode(t *testing.T, s string) ast.Node {
	t.Helper()
	node, err := parser.ParseBytes([]byte(s), 0)
	require.NoError(t, err)
	require.Len(t, node.Docs, 1)
	return node.Docs[0].Body
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file %s", path)
	return data
}
