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
			name:      "no plain",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{},
			expected: heredoc.Doc(`
				` + "\x1b[93m~\x1b[0m \x1b[93m.global.scrape_interval\x1b[0m: \x1b[92m15s\x1b[0m → \x1b[92m30s\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.global.scrape_timeout\x1b[0m: \x1b[92m10s\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.rule_files[2]\x1b[0m: \x1b[92m\"additional_rules.yml\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.scrape_configs[0].static_configs[0].targets[1]\x1b[0m: \x1b[92m\"localhost:9200\"\x1b[0m" + `
				` + "\x1b[91m-\x1b[0m \x1b[91m.scrape_configs[1].static_configs[0].labels\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    instance\x1b[0m:\x1b[92m \"app-1\"\x1b[0m" + `
				` + "\x1b[91m-\x1b[0m \x1b[91m.scrape_configs[1].static_configs[1]\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    targets\x1b[0m:" + `
				      -` + "\x1b[92m \"app-2.local:8000\"\x1b[0m\x1b[96m\x1b[0m" + `
				` + "\x1b[96m    labels\x1b[0m:\x1b[96m\x1b[0m" + `
				` + "\x1b[96m      instance\x1b[0m:\x1b[92m \"app-2\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.scrape_configs[2]\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    job_name\x1b[0m:\x1b[92m \"kubernetes\"\x1b[0m\x1b[96m\x1b[0m" + `
				` + "\x1b[96m    kubernetes_sd_configs\x1b[0m:" + `
				      -` + "\x1b[96m role\x1b[0m:\x1b[92m pod\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mrelabel_configs\x1b[0m:" + `
				      -` + "\x1b[96m source_labels\x1b[0m: [\x1b[92m__meta_kubernetes_pod_label_app\x1b[0m]\x1b[96m\x1b[0m" + `
				` + "\x1b[96m        regex\x1b[0m:\x1b[92m my-app\x1b[0m" + `
				` + "\x1b[92m        \x1b[0m\x1b[96maction\x1b[0m:\x1b[92m keep\x1b[0m" + ``),
		},
		{
			name:      "no plain with metadata",
			leftYaml:  readFile(t, "testdata/left.yaml"),
			rightYaml: readFile(t, "testdata/right.yaml"),
			options:   []FormatOption{WithMetadata},
			expected: heredoc.Doc(`
				` + "\x1b[93m~\x1b[0m \x1b[93m.global.scrape_interval\x1b[0m: \x1b[97m[line:2 <String>]\x1b[0m \x1b[92m15s\x1b[0m → \x1b[97m[line:2 <String>]\x1b[0m \x1b[92m30s\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.global.scrape_timeout\x1b[0m: \x1b[97m[line:4 <String>]\x1b[0m \x1b[92m10s\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.rule_files[2]\x1b[0m: \x1b[97m[line:9 <String>]\x1b[0m \x1b[92m\"additional_rules.yml\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.scrape_configs[0].static_configs[0].targets[1]\x1b[0m: \x1b[97m[line:16 <String>]\x1b[0m \x1b[92m\"localhost:9200\"\x1b[0m" + `
				` + "\x1b[91m-\x1b[0m \x1b[91m.scrape_configs[1].static_configs[0].labels\x1b[0m: \x1b[97m[line:21 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    instance\x1b[0m:\x1b[92m \"app-1\"\x1b[0m" + `
				` + "\x1b[91m-\x1b[0m \x1b[91m.scrape_configs[1].static_configs[1]\x1b[0m: \x1b[97m[line:22 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    targets\x1b[0m:" + `
				      -` + "\x1b[92m \"app-2.local:8000\"\x1b[0m\x1b[96m\x1b[0m" + `
				` + "\x1b[96m    labels\x1b[0m:\x1b[96m\x1b[0m" + `
				` + "\x1b[96m      instance\x1b[0m:\x1b[92m \"app-2\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.scrape_configs[2]\x1b[0m: \x1b[97m[line:23 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    job_name\x1b[0m:\x1b[92m \"kubernetes\"\x1b[0m\x1b[96m\x1b[0m" + `
				` + "\x1b[96m    kubernetes_sd_configs\x1b[0m:" + `
				      -` + "\x1b[96m role\x1b[0m:\x1b[92m pod\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mrelabel_configs\x1b[0m:" + `
				      -` + "\x1b[96m source_labels\x1b[0m: [\x1b[92m__meta_kubernetes_pod_label_app\x1b[0m]\x1b[96m\x1b[0m" + `
				` + "\x1b[96m        regex\x1b[0m:\x1b[92m my-app\x1b[0m" + `
				` + "\x1b[92m        \x1b[0m\x1b[96maction\x1b[0m:\x1b[92m keep\x1b[0m" + ``),
		},
		{
			name:      "multi-document plain",
			leftYaml:  readFile(t, "testdata/multi-docs-left.yaml"),
			rightYaml: readFile(t, "testdata/multi-docs-right.yaml"),
			options:   []FormatOption{Plain},
			expected: heredoc.Doc(`
				+ .properties.role: 
				    type: string
				    enum:
				      - admin
				      - viewer
				      - editor
				    description: "The role of the user in the system, defining their permissions"
				+ .required[2]: email
				+ .required[3]: role
				---
				~ .properties.name.maxLength: 32 → 64
				+ .properties.description: 
				    type: string
				    maxLength: 512
				    description: "A brief description of the product"
				---
				- .properties.timestamp: 
				    type: number
				    description: "The unix timestamp of when the order was placed"
				+ .properties.created_at: 
				    type: string
				    format: date-time
				    description: "The time when the order was created"`),
		},
		{
			name:      "multi-document paths-only",
			leftYaml:  readFile(t, "testdata/multi-docs-left.yaml"),
			rightYaml: readFile(t, "testdata/multi-docs-right.yaml"),
			options:   []FormatOption{Plain, PathsOnly},
			expected: heredoc.Doc(`
				+ .properties.role
				+ .required[2]
				+ .required[3]
				---
				~ .properties.name.maxLength
				+ .properties.description
				---
				- .properties.timestamp
				+ .properties.created_at`),
		},
		{
			name:      "multi-document with metadata",
			leftYaml:  readFile(t, "testdata/multi-docs-left.yaml"),
			rightYaml: readFile(t, "testdata/multi-docs-right.yaml"),
			options:   []FormatOption{Plain, WithMetadata},
			expected: heredoc.Doc(`
				+ .properties.role: [line:14 <Mapping>] 
				    type: string
				    enum:
				      - admin
				      - viewer
				      - editor
				    description: "The role of the user in the system, defining their permissions"
				+ .required[2]: [line:23 <String>] email
				+ .required[3]: [line:24 <String>] role
				---
				~ .properties.name.maxLength: [line:27 <Integer>] 32 → [line:36 <Integer>] 64
				+ .properties.description: [line:40 <Mapping>] 
				    type: string
				    maxLength: 512
				    description: "A brief description of the product"
				---
				- .properties.timestamp: [line:49 <Mapping>] 
				    type: number
				    description: "The unix timestamp of when the order was placed"
				+ .properties.created_at: [line:62 <Mapping>] 
				    type: string
				    format: date-time
				    description: "The time when the order was created"`),
		},
		{
			name:      "multi-document with counts",
			leftYaml:  readFile(t, "testdata/multi-docs-left.yaml"),
			rightYaml: readFile(t, "testdata/multi-docs-right.yaml"),
			options:   []FormatOption{Plain, IncludeCounts},
			expected: heredoc.Doc(`
				3 added, 0 deleted, 0 modified
				+ .properties.role: 
				    type: string
				    enum:
				      - admin
				      - viewer
				      - editor
				    description: "The role of the user in the system, defining their permissions"
				+ .required[2]: email
				+ .required[3]: role
				---
				1 added, 0 deleted, 1 modified
				~ .properties.name.maxLength: 32 → 64
				+ .properties.description: 
				    type: string
				    maxLength: 512
				    description: "A brief description of the product"
				---
				1 added, 1 deleted, 0 modified
				- .properties.timestamp: 
				    type: number
				    description: "The unix timestamp of when the order was placed"
				+ .properties.created_at: 
				    type: string
				    format: date-time
				    description: "The time when the order was created"`),
		},
		{
			name:      "multi-document no plain",
			leftYaml:  readFile(t, "testdata/multi-docs-left.yaml"),
			rightYaml: readFile(t, "testdata/multi-docs-right.yaml"),
			options:   []FormatOption{},
			expected: heredoc.Doc(`
				` + "\x1b[92m+\x1b[0m \x1b[92m.properties.role\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m string\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96menum\x1b[0m:" + `
				      -` + "\x1b[92m admin\x1b[0m" + `
				` + "\x1b[92m      \x1b[0m-\x1b[92m viewer\x1b[0m" + `
				` + "\x1b[92m      \x1b[0m-\x1b[92m editor\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"The role of the user in the system, defining their permissions\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.required[2]\x1b[0m: \x1b[92memail\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.required[3]\x1b[0m: \x1b[92mrole\x1b[0m" + `
				---
				` + "\x1b[93m~\x1b[0m \x1b[93m.properties.name.maxLength\x1b[0m: \x1b[95m32\x1b[0m → \x1b[95m64\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.properties.description\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m string\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mmaxLength\x1b[0m:\x1b[95m 512\x1b[0m" + `
				` + "\x1b[95m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"A brief description of the product\"\x1b[0m" + `
				---
				` + "\x1b[91m-\x1b[0m \x1b[91m.properties.timestamp\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m number\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"The unix timestamp of when the order was placed\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.properties.created_at\x1b[0m: \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m string\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mformat\x1b[0m:\x1b[92m date-time\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"The time when the order was created\"\x1b[0m" + ``),
		},
		{
			name:      "multi-document no plain with metadata",
			leftYaml:  readFile(t, "testdata/multi-docs-left.yaml"),
			rightYaml: readFile(t, "testdata/multi-docs-right.yaml"),
			options:   []FormatOption{WithMetadata},
			expected: heredoc.Doc(`
				` + "\x1b[92m+\x1b[0m \x1b[92m.properties.role\x1b[0m: \x1b[97m[line:14 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m string\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96menum\x1b[0m:" + `
				      -` + "\x1b[92m admin\x1b[0m" + `
				` + "\x1b[92m      \x1b[0m-\x1b[92m viewer\x1b[0m" + `
				` + "\x1b[92m      \x1b[0m-\x1b[92m editor\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"The role of the user in the system, defining their permissions\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.required[2]\x1b[0m: \x1b[97m[line:23 <String>]\x1b[0m \x1b[92memail\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.required[3]\x1b[0m: \x1b[97m[line:24 <String>]\x1b[0m \x1b[92mrole\x1b[0m" + `
				---
				` + "\x1b[93m~\x1b[0m \x1b[93m.properties.name.maxLength\x1b[0m: \x1b[97m[line:27 <Integer>]\x1b[0m \x1b[95m32\x1b[0m → \x1b[97m[line:36 <Integer>]\x1b[0m \x1b[95m64\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.properties.description\x1b[0m: \x1b[97m[line:40 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m string\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mmaxLength\x1b[0m:\x1b[95m 512\x1b[0m" + `
				` + "\x1b[95m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"A brief description of the product\"\x1b[0m" + `
				---
				` + "\x1b[91m-\x1b[0m \x1b[91m.properties.timestamp\x1b[0m: \x1b[97m[line:49 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m number\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"The unix timestamp of when the order was placed\"\x1b[0m" + `
				` + "\x1b[92m+\x1b[0m \x1b[92m.properties.created_at\x1b[0m: \x1b[97m[line:62 <Mapping>]\x1b[0m \x1b[96m\x1b[0m" + `
				` + "\x1b[96m    type\x1b[0m:\x1b[92m string\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mformat\x1b[0m:\x1b[92m date-time\x1b[0m" + `
				` + "\x1b[92m    \x1b[0m\x1b[96mdescription\x1b[0m:\x1b[92m \"The time when the order was created\"\x1b[0m" + ``),
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
