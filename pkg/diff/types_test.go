package diff

import (
	"fmt"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/stretchr/testify/require"
)

func TestNewDiff(t *testing.T) {
	leftYaml := "name: Alice"
	rightYaml := "name: Bob"

	leftAst, err := parser.ParseBytes([]byte(leftYaml), 0)
	require.NoError(t, err, "Failed to parse left YAML")

	rightAst, err := parser.ParseBytes([]byte(rightYaml), 0)
	require.NoError(t, err, "Failed to parse right YAML")

	leftNode := leftAst.Docs[0].Body
	rightNode := rightAst.Docs[0].Body

	diff := newDiff(leftNode, rightNode)

	require.NotNil(t, diff, "newDiff returned nil")
	require.Equal(t, leftNode, diff.leftNode, "Left node not set correctly")
	require.Equal(t, rightNode, diff.rightNode, "Right node not set correctly")
}

func TestDiff_Type(t *testing.T) {
	yaml1 := "name: Alice"
	yaml2 := "name: Bob"

	ast1, err := parser.ParseBytes([]byte(yaml1), 0)
	require.NoError(t, err, "Failed to parse YAML")

	ast2, err := parser.ParseBytes([]byte(yaml2), 0)
	require.NoError(t, err, "Failed to parse YAML")

	node1 := ast1.Docs[0].Body
	node2 := ast2.Docs[0].Body

	tests := []struct {
		name      string
		leftNode  ast.Node
		rightNode ast.Node
		expected  DiffType
	}{
		{"Added - nil left", nil, node1, Added},
		{"Deleted - nil right", node1, nil, Deleted},
		{"Modified - both nodes", node1, node2, Modified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := newDiff(tt.leftNode, tt.rightNode)
			require.Equal(t, tt.expected, diff.Type(), "Expected type %v, got %v", tt.expected, diff.Type())
		})
	}
}

func TestDocDiffs_Len(t *testing.T) {
	yaml1 := "name: Alice"
	yaml2 := "name: Bob"

	ast1, err := parser.ParseBytes([]byte(yaml1), 0)
	require.NoError(t, err, "Failed to parse YAML")

	ast2, err := parser.ParseBytes([]byte(yaml2), 0)
	require.NoError(t, err, "Failed to parse YAML")

	node1 := ast1.Docs[0].Body
	node2 := ast2.Docs[0].Body

	docDiffs := DocDiffs{
		newDiff(node1, node2),
		newDiff(nil, node1),
		newDiff(node2, nil),
	}

	require.Equal(t, 3, docDiffs.Len(), "Expected length 3, got %d", docDiffs.Len())
}

func TestDocDiffs_Swap(t *testing.T) {
	yaml1 := "name: Alice"
	yaml2 := "name: Bob"

	ast1, err := parser.ParseBytes([]byte(yaml1), 0)
	require.NoError(t, err, "Failed to parse YAML")

	ast2, err := parser.ParseBytes([]byte(yaml2), 0)
	require.NoError(t, err, "Failed to parse YAML")

	node1 := ast1.Docs[0].Body
	node2 := ast2.Docs[0].Body

	diff1 := newDiff(node1, node2)
	diff2 := newDiff(nil, node1)

	docDiffs := DocDiffs{diff1, diff2}

	// Swap elements
	docDiffs.Swap(0, 1)

	require.Equal(t, diff2, docDiffs[0], "Swap did not work correctly - element 0")
	require.Equal(t, diff1, docDiffs[1], "Swap did not work correctly - element 1")
}

func TestDocDiffs_Less(t *testing.T) {
	// Create YAML with multiple lines to test line number comparison
	leftYaml := `name: Alice
age: 25
city: New York`

	rightYaml := `name: Bob
age: 30
city: Boston`

	leftAst, err := parser.ParseBytes([]byte(leftYaml), 0)
	require.NoError(t, err, "Failed to parse left YAML")

	rightAst, err := parser.ParseBytes([]byte(rightYaml), 0)
	require.NoError(t, err, "Failed to parse right YAML")

	// Get the mapping node which contains all key-value pairs
	leftMapping := leftAst.Docs[0].Body.(*ast.MappingNode)
	rightMapping := rightAst.Docs[0].Body.(*ast.MappingNode)

	// Create diffs with different line positions
	diff1 := newDiff(leftMapping.Values[0].Key, rightMapping.Values[0].Key) // line 1: name
	diff2 := newDiff(leftMapping.Values[1].Key, rightMapping.Values[1].Key) // line 2: age
	diff3 := newDiff(leftMapping.Values[2].Key, rightMapping.Values[2].Key) // line 3: city

	docDiffs := DocDiffs{diff1, diff2, diff3}

	// Test that diff1 (line 1) comes before diff2 (line 2)
	require.True(t, docDiffs.Less(0, 1), "Expected diff at line 1 to be less than diff at line 2")

	// Test that diff2 (line 2) comes before diff3 (line 3)
	require.True(t, docDiffs.Less(1, 2), "Expected diff at line 2 to be less than diff at line 3")

	// Test reverse order
	require.False(t, docDiffs.Less(1, 0), "Expected diff at line 2 to NOT be less than diff at line 1")
}

func TestFileDiffs_HasDiff(t *testing.T) {
	yaml1 := "name: Alice"
	yaml2 := "name: Bob"

	ast1, err := parser.ParseBytes([]byte(yaml1), 0)
	require.NoError(t, err, "Failed to parse YAML")

	ast2, err := parser.ParseBytes([]byte(yaml2), 0)
	require.NoError(t, err, "Failed to parse YAML")

	node1 := ast1.Docs[0].Body
	node2 := ast2.Docs[0].Body

	tests := []struct {
		name      string
		fileDiffs FileDiffs
		expected  bool
	}{
		{
			name:      "No diffs",
			fileDiffs: FileDiffs{},
			expected:  false,
		},
		{
			name:      "Empty doc diffs",
			fileDiffs: FileDiffs{DocDiffs{}},
			expected:  false,
		},
		{
			name: "Has diffs",
			fileDiffs: FileDiffs{
				DocDiffs{newDiff(node1, node2)},
			},
			expected: true,
		},
		{
			name: "Multiple docs with diffs",
			fileDiffs: FileDiffs{
				DocDiffs{},
				DocDiffs{newDiff(node1, node2)},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.fileDiffs.HasDiff(), "Expected HasDiff() to return %v, got %v", tt.expected, tt.fileDiffs.HasDiff())
		})
	}
}

func TestDiff_Path(t *testing.T) {
	leftYaml := `name: Alice
age: 25`

	rightYaml := `name: Bob
city: Boston`

	leftAst, err := parser.ParseBytes([]byte(leftYaml), 0)
	require.NoError(t, err, "Failed to parse left YAML")

	rightAst, err := parser.ParseBytes([]byte(rightYaml), 0)
	require.NoError(t, err, "Failed to parse right YAML")

	// Get mapping nodes
	leftMapping := leftAst.Docs[0].Body.(*ast.MappingNode)
	rightMapping := rightAst.Docs[0].Body.(*ast.MappingNode)

	tests := []struct {
		name      string
		leftNode  ast.Node
		rightNode ast.Node
		expected  string
	}{
		{
			name:      "Added - uses right node path",
			leftNode:  nil,
			rightNode: rightMapping.Values[1].Key, // city key
			expected:  ".city",
		},
		{
			name:      "Deleted - uses left node path",
			leftNode:  leftMapping.Values[1].Key, // age key
			rightNode: nil,
			expected:  ".age",
		},
		{
			name:      "Modified - uses left node path",
			leftNode:  leftMapping.Values[0].Key,  // name key
			rightNode: rightMapping.Values[0].Key, // name key
			expected:  ".name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := newDiff(tt.leftNode, tt.rightNode)
			path := diff.Path()
			require.Equal(t, tt.expected, path, "Expected path %q, got %q", tt.expected, path)
		})
	}
}

func ExampleDocDiffs_Format() {
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

	s := diffs.Format(Plain)
	fmt.Println(s)

	// Output:
	// ~ .name: Alice → Bob
	// - .city: New York
	// + .age: 30
	// ~ .items[1]: bar → baz
}

func ExampleDocDiffs_Format_includeCounts() {
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

	s := diffs.Format(Plain, IncludeCounts)
	fmt.Println(s)

	// Output:
	// 1 added, 1 deleted, 2 modified
	// ~ .name: Alice → Bob
	// - .city: New York
	// + .age: 30
	// ~ .items[1]: bar → baz
}

func ExampleDocDiffs_Format_withMetadata() {
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

	s := diffs.Format(Plain, WithMetadata)
	fmt.Println(s)

	// Output:
	// ~ .name: [line:2 <String>] Alice → [line:2 <String>] Bob
	// - .city: [line:3 <String>] New York
	// + .age: [line:3 <Integer>] 30
	// ~ .items[1]: [line:6 <String>] bar → [line:6 <String>] baz
}

func ExampleDocDiffs_Format_pathsOnly() {
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

	s := diffs.Format(Plain, PathsOnly)
	fmt.Println(s)

	// Output:
	// 	~ .name
	// - .city
	// + .age
	// ~ .items[1]
}

func TestDocDiffs_Format(t *testing.T) {
	tests := []struct {
		name      string
		leftYaml  []byte
		rightYaml []byte
		options   []FormatOption
		expected  string
	}{
		{
			name:      "plain",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain},
			expected: heredoc.Doc(`
				~ .global.scrape_interval: 15s → 30s
				+ .global.scrape_timeout: 10s
				+ .rule_files[2]: "additional_rules.yml"
				+ .scrape_configs[0].static_configs[0].targets[1]: "localhost:9200"
				- .scrape_configs[1].static_configs[0].labels: 
				    instance: "app-1"
				- .scrape_configs[1].static_configs[1]: 
				    targets:
				      - "app-2.local:8000"
				    labels:
				      instance: "app-2"
				+ .scrape_configs[2]: 
				    job_name: "kubernetes"
				    kubernetes_sd_configs:
				      - role: pod
				    relabel_configs:
				      - source_labels: [__meta_kubernetes_pod_label_app]
				        regex: my-app
				        action: keep`),
		},
		{
			name:      "paths only",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain, PathsOnly},
			expected: heredoc.Doc(`
				~ .global.scrape_interval
				+ .global.scrape_timeout
				+ .rule_files[2]
				+ .scrape_configs[0].static_configs[0].targets[1]
				- .scrape_configs[1].static_configs[0].labels
				- .scrape_configs[1].static_configs[1]
				+ .scrape_configs[2]`),
		},
		{
			name:      "with metadata",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain, WithMetadata},
			expected: heredoc.Doc(`
				~ .global.scrape_interval: [line:2 <String>] 15s → [line:2 <String>] 30s
				+ .global.scrape_timeout: [line:4 <String>] 10s
				+ .rule_files[2]: [line:9 <String>] "additional_rules.yml"
				+ .scrape_configs[0].static_configs[0].targets[1]: [line:16 <String>] "localhost:9200"
				- .scrape_configs[1].static_configs[0].labels: [line:21 <Mapping>] 
				    instance: "app-1"
				- .scrape_configs[1].static_configs[1]: [line:22 <Mapping>] 
				    targets:
				      - "app-2.local:8000"
				    labels:
				      instance: "app-2"
				+ .scrape_configs[2]: [line:23 <Mapping>] 
				    job_name: "kubernetes"
				    kubernetes_sd_configs:
				      - role: pod
				    relabel_configs:
				      - source_labels: [__meta_kubernetes_pod_label_app]
				        regex: my-app
				        action: keep`),
		},
		{
			name:      "with counts",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain, IncludeCounts},
			expected: heredoc.Doc(`
				4 added, 2 deleted, 1 modified
				~ .global.scrape_interval: 15s → 30s
				+ .global.scrape_timeout: 10s
				+ .rule_files[2]: "additional_rules.yml"
				+ .scrape_configs[0].static_configs[0].targets[1]: "localhost:9200"
				- .scrape_configs[1].static_configs[0].labels: 
				    instance: "app-1"
				- .scrape_configs[1].static_configs[1]: 
				    targets:
				      - "app-2.local:8000"
				    labels:
				      instance: "app-2"
				+ .scrape_configs[2]: 
				    job_name: "kubernetes"
				    kubernetes_sd_configs:
				      - role: pod
				    relabel_configs:
				      - source_labels: [__meta_kubernetes_pod_label_app]
				        regex: my-app
				        action: keep`),
		},
		{
			name:      "paths only with counts",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain, PathsOnly, IncludeCounts},
			expected: heredoc.Doc(`
				4 added, 2 deleted, 1 modified
				~ .global.scrape_interval
				+ .global.scrape_timeout
				+ .rule_files[2]
				+ .scrape_configs[0].static_configs[0].targets[1]
				- .scrape_configs[1].static_configs[0].labels
				- .scrape_configs[1].static_configs[1]
				+ .scrape_configs[2]`),
		},
		{
			name:      "metadata with counts",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain, WithMetadata, IncludeCounts},
			expected: heredoc.Doc(`
				4 added, 2 deleted, 1 modified
				~ .global.scrape_interval: [line:2 <String>] 15s → [line:2 <String>] 30s
				+ .global.scrape_timeout: [line:4 <String>] 10s
				+ .rule_files[2]: [line:9 <String>] "additional_rules.yml"
				+ .scrape_configs[0].static_configs[0].targets[1]: [line:16 <String>] "localhost:9200"
				- .scrape_configs[1].static_configs[0].labels: [line:21 <Mapping>] 
				    instance: "app-1"
				- .scrape_configs[1].static_configs[1]: [line:22 <Mapping>] 
				    targets:
				      - "app-2.local:8000"
				    labels:
				      instance: "app-2"
				+ .scrape_configs[2]: [line:23 <Mapping>] 
				    job_name: "kubernetes"
				    kubernetes_sd_configs:
				      - role: pod
				    relabel_configs:
				      - source_labels: [__meta_kubernetes_pod_label_app]
				        regex: my-app
				        action: keep`),
		},
		{
			name:      "all options combined",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{Plain, PathsOnly, WithMetadata, IncludeCounts},
			expected: heredoc.Doc(`
				4 added, 2 deleted, 1 modified
				~ .global.scrape_interval
				+ .global.scrape_timeout
				+ .rule_files[2]
				+ .scrape_configs[0].static_configs[0].targets[1]
				- .scrape_configs[1].static_configs[0].labels
				- .scrape_configs[1].static_configs[1]
				+ .scrape_configs[2]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docDiffs, err := Compare(tt.leftYaml, tt.rightYaml)
			require.NoError(t, err, "Failed to compare YAML files")

			output := docDiffs.Format(tt.options...)
			require.Equal(t, tt.expected, output, "Expected formatted output to match")
		})
	}
}
