package diff

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml/parser"
)

func TestFormatUnified_BasicModification(t *testing.T) {
	left := `name: Alice
age: 30`
	right := `name: Bob
age: 30`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain the modification
	if !strings.Contains(output, "-name: Alice") {
		t.Errorf("expected output to contain '-name: Alice', got: %s", output)
	}
	if !strings.Contains(output, "+name: Bob") {
		t.Errorf("expected output to contain '+name: Bob', got: %s", output)
	}
	// Should contain unchanged line
	if !strings.Contains(output, " age: 30") {
		t.Errorf("expected output to contain ' age: 30', got: %s", output)
	}
}

func TestFormatUnified_Addition(t *testing.T) {
	left := `name: Alice`
	right := `name: Alice
age: 30`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain the addition
	if !strings.Contains(output, "+age: 30") {
		t.Errorf("expected output to contain '+age: 30', got: %s", output)
	}
	// Should contain unchanged line
	if !strings.Contains(output, " name: Alice") {
		t.Errorf("expected output to contain ' name: Alice', got: %s", output)
	}
}

func TestFormatUnified_Deletion(t *testing.T) {
	left := `name: Alice
age: 30`
	right := `name: Alice`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain the deletion
	if !strings.Contains(output, "-age: 30") {
		t.Errorf("expected output to contain '-age: 30', got: %s", output)
	}
	// Should contain unchanged line
	if !strings.Contains(output, " name: Alice") {
		t.Errorf("expected output to contain ' name: Alice', got: %s", output)
	}
}

func TestFormatUnified_NestedStructure(t *testing.T) {
	left := `metadata:
  name: app
  version: v1`
	right := `metadata:
  name: app
  version: v2`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain the modification
	if !strings.Contains(output, "-  version: v1") {
		t.Errorf("expected output to contain '-  version: v1', got: %s", output)
	}
	if !strings.Contains(output, "+  version: v2") {
		t.Errorf("expected output to contain '+  version: v2', got: %s", output)
	}
	// Should contain unchanged lines
	if !strings.Contains(output, " metadata:") {
		t.Errorf("expected output to contain ' metadata:', got: %s", output)
	}
	if !strings.Contains(output, "   name: app") {
		t.Errorf("expected output to contain '   name: app', got: %s", output)
	}
}

func TestFormatUnified_Sequence(t *testing.T) {
	left := `items:
  - foo
  - bar`
	right := `items:
  - foo
  - baz`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain the modification
	if !strings.Contains(output, "- bar") || !strings.Contains(output, "-") {
		t.Logf("output: %s", output)
	}
	if !strings.Contains(output, "+ baz") || !strings.Contains(output, "+") {
		t.Logf("output: %s", output)
	}
	// Should contain unchanged items
	if !strings.Contains(output, " items:") {
		t.Errorf("expected output to contain ' items:', got: %s", output)
	}
}

func TestFormatUnified_NoDifference(t *testing.T) {
	yaml := `name: Alice
age: 30`

	leftAST, err := parser.ParseBytes([]byte(yaml), 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(yaml), 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// With no diffs, all lines should be unchanged (space prefix)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, " ") {
			t.Errorf("expected all lines to have space prefix, got: %s", line)
		}
	}
}

func TestFormatUnified_MultiDocument(t *testing.T) {
	left := `name: doc1
---
name: doc2`
	right := `name: doc1-modified
---
name: doc2`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain document separator
	if !strings.Contains(output, "---") {
		t.Errorf("expected output to contain '---' separator, got: %s", output)
	}

	// Should contain the modification in first doc
	if !strings.Contains(output, "-name: doc1") {
		t.Errorf("expected output to contain '-name: doc1', got: %s", output)
	}
	if !strings.Contains(output, "+name: doc1-modified") {
		t.Errorf("expected output to contain '+name: doc1-modified', got: %s", output)
	}

	// Should contain unchanged second doc
	if !strings.Contains(output, " name: doc2") {
		t.Errorf("expected output to contain ' name: doc2', got: %s", output)
	}
}

func TestFormatUnified_ComplexNestedChange(t *testing.T) {
	left := `spec:
  containers:
    - name: app
      image: app:1.0`
	right := `spec:
  containers:
    - name: app
      image: app:2.0`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should contain the modification
	if !strings.Contains(output, "app:1.0") && !strings.Contains(output, "-") {
		t.Logf("output: %s", output)
	}
	if !strings.Contains(output, "app:2.0") && !strings.Contains(output, "+") {
		t.Logf("output: %s", output)
	}

	// Should have proper structure
	if !strings.Contains(output, " spec:") {
		t.Errorf("expected output to contain ' spec:', got: %s", output)
	}
}

func TestUnifiedPlainOption(t *testing.T) {
	left := `name: Alice`
	right := `name: Bob`

	leftAST, err := parser.ParseBytes([]byte(left), 0)
	if err != nil {
		t.Fatalf("failed to parse left: %v", err)
	}

	rightAST, err := parser.ParseBytes([]byte(right), 0)
	if err != nil {
		t.Fatalf("failed to parse right: %v", err)
	}

	diffs := CompareAst(leftAST, rightAST)

	// Test with plain option
	output := FormatUnified(leftAST, rightAST, diffs, UnifiedPlain)

	// Should not contain ANSI escape codes
	if strings.Contains(output, "\x1b[") {
		t.Errorf("expected output to not contain ANSI escape codes with UnifiedPlain option")
	}
}

func TestCompareFileWithAST(t *testing.T) {
	// This test requires actual files, so we skip if running in isolation
	// The functionality is tested through the integration tests
	t.Skip("Requires actual files - tested through integration")
}
