package diff

import (
	"fmt"
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

// empty yaml is evaluated as nil node in ast representation
// see: https://github.com/goccy/go-yaml/issues/753
func TestCompareEmptyYaml(t *testing.T) {
	tests := []struct {
		name         string
		left         string
		right        string
		expectedType DiffType
	}{
		{
			name:         "empty and null",
			left:         "",
			right:        "null",
			expectedType: Added,
		},
		{
			name:         "null and empty",
			left:         "null",
			right:        "",
			expectedType: Deleted,
		},
		{
			name:         "empty and integer",
			left:         "",
			right:        "42",
			expectedType: Added,
		},
		{
			name:         "integer and empty",
			left:         "42",
			right:        "",
			expectedType: Deleted,
		},
		{
			name: "empty and mapping",
			left: "",
			right: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			expectedType: Added,
		},
		{
			name: "mapping and empty",
			left: heredoc.Doc(`
				foo: bar
				baz: qux
			`),
			right:        "",
			expectedType: Deleted,
		},
		{
			name: "empty and sequence",
			left: "",
			right: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			expectedType: Added,
		},
		{
			name: "sequence and empty",
			left: heredoc.Doc(`
				- foo
				- bar
				- baz
			`),
			right:        "",
			expectedType: Deleted,
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

			require.Len(t, docDiffs, 1)
			diff := docDiffs[0]
			require.Equal(t, diff.Type(), tt.expectedType)
			require.Empty(t, diff.Path())
		})
	}
}

func TestCompareMultiDocsYaml(t *testing.T) {
	tests := []struct {
		name          string
		left          string
		right         string
		expectedDiffs []map[string]DiffType
	}{
		{
			name: "multi-docs with missing doc",
			left: heredoc.Doc(`
				foo: bar
				---
				baz: qux
			`),
			right: heredoc.Doc(`
				foo: bar
			`),
			expectedDiffs: []map[string]DiffType{
				{},
				{
					"": Deleted,
				},
			},
		},
	}

	//todo: test additional multi-docs
	//todo: test multiple missing multi-docs
	//todo: test multiple additional multi-docs
	//todo: test empty yaml in multi-docs

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

	output := diffs.Format(Plain)
	fmt.Println(output)

	// Output:
	// ~ .name: Alice -> Bob
	// - .city: New York
	// + .age: 30
	// ~ .items[1]: bar -> baz
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

//todo: func Test_compareNodesCollectionTypes(t *testing.T) {}

//todo: func Test_compareNodesScalarAndCollectionTypes(t *testing.T) {}

//todo: Test_compareNodesNullNodes(t *testing.T) {}

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
