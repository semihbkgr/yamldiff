package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestNewConfig(t *testing.T) {
	cfg := newConfig()

	if cfg.exitOnDifference != false {
		t.Errorf("expected exitOnDifference to be false, got %v", cfg.exitOnDifference)
	}
	if cfg.ignoreSeqOrder != false {
		t.Errorf("expected ignoreSeqOrder to be false, got %v", cfg.ignoreSeqOrder)
	}
	if cfg.color != "auto" {
		t.Errorf("expected color to be 'auto', got %s", cfg.color)
	}
	if cfg.pathOnly != false {
		t.Errorf("expected pathOnly to be false, got %v", cfg.pathOnly)
	}
	if cfg.metadata != false {
		t.Errorf("expected metadata to be false, got %v", cfg.metadata)
	}
	if cfg.stat != false {
		t.Errorf("expected stat to be false, got %v", cfg.stat)
	}
}

func TestConfigShouldUseColor(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		noColor  bool
		expected bool
	}{
		{"always", "always", false, true},
		{"always with NoColor", "always", true, true},
		{"never", "never", false, false},
		{"never with NoColor", "never", true, false},
		{"auto without NoColor", "auto", false, true},
		{"auto with NoColor", "auto", true, false},
		{"invalid value", "invalid", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original NoColor value
			originalNoColor := color.NoColor
			defer func() { color.NoColor = originalNoColor }()

			color.NoColor = tt.noColor

			cfg := &config{color: tt.color}
			result := cfg.shouldUseColor()

			if result != tt.expected {
				t.Errorf("shouldUseColor() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigCompareOptions(t *testing.T) {
	tests := []struct {
		name           string
		ignoreSeqOrder bool
		expectedCount  int
	}{
		{"no options", false, 0},
		{"with ignoreSeqOrder", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{ignoreSeqOrder: tt.ignoreSeqOrder}
			opts := cfg.compareOptions()

			if len(opts) != tt.expectedCount {
				t.Errorf("compareOptions() returned %d options, want %d", len(opts), tt.expectedCount)
			}
		})
	}
}

func TestConfigFormatOptions(t *testing.T) {
	tests := []struct {
		name         string
		color        string
		pathOnly     bool
		metadata     bool
		stat         bool
		expectedOpts int
	}{
		{"no options", "always", false, false, false, 0},
		{"plain only", "never", false, false, false, 1},
		{"paths only", "always", true, false, false, 1},
		{"metadata only", "always", false, true, false, 1},
		{"stat only", "always", false, false, true, 0}, // stat is handled separately, not as FormatOption
		{"all options", "never", true, true, true, 3},  // stat not included in FormatOptions
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{
				color:    tt.color,
				pathOnly: tt.pathOnly,
				metadata: tt.metadata,
				stat:     tt.stat,
			}
			opts := cfg.formatOptions()

			if len(opts) != tt.expectedOpts {
				t.Errorf("formatOptions() returned %d options, want %d", len(opts), tt.expectedOpts)
			}
		})
	}
}

func TestConfigValidateColorFlag(t *testing.T) {
	tests := []struct {
		name    string
		color   string
		wantErr bool
	}{
		{"valid always", "always", false},
		{"valid never", "never", false},
		{"valid auto", "auto", false},
		{"invalid value", "invalid", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{color: tt.color}
			err := cfg.validateColorFlag()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateColorFlag() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "invalid color value") {
					t.Errorf("validateColorFlag() error = %v, expected error to contain 'invalid color value'", err)
				}
			}
		})
	}
}

func TestConfigValidateMutuallyExclusiveFlags(t *testing.T) {
	tests := []struct {
		name     string
		pathOnly bool
		metadata bool
		stat     bool
		wantErr  bool
	}{
		{"no flags", false, false, false, false},
		{"path only", true, false, false, false},
		{"metadata only", false, true, false, false},
		{"stat only", false, false, true, false},
		{"path-only and metadata", true, true, false, true},
		{"stat and path-only", false, true, true, true},
		{"stat and metadata", true, false, true, true},
		{"all flags", true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{
				pathOnly: tt.pathOnly,
				metadata: tt.metadata,
				stat:     tt.stat,
			}
			err := cfg.validateMutuallyExclusiveFlags()

			if (err != nil) != tt.wantErr {
				t.Errorf("validateMutuallyExclusiveFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRootCommand(t *testing.T) {
	cmd := newRootCommand()

	if cmd.Use != "yamldiff [flags] <file-left> <file-right>" {
		t.Errorf("expected Use to be 'yamldiff [flags] <file-left> <file-right>', got %s", cmd.Use)
	}

	if cmd.Short != "structural comparison on two yaml files" {
		t.Errorf("expected Short to be 'structural comparison on two yaml files', got %s", cmd.Short)
	}

	if cmd.Args == nil {
		t.Error("expected Args to be set")
	}

	if cmd.SilenceUsage != false {
		t.Errorf("expected SilenceUsage to be false, got %v", cmd.SilenceUsage)
	}

	if cmd.DisableFlagsInUseLine != true {
		t.Errorf("expected DisableFlagsInUseLine to be true, got %v", cmd.DisableFlagsInUseLine)
	}

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	// Check that all expected flags are present
	expectedFlags := []string{"exit-code", "ignore-order", "color", "path-only", "metadata", "stat"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("expected flag %s to be present", flagName)
		}
	}
}

func TestRunCommandWithInvalidArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"no args", []string{}, true},
		{"one arg", []string{"file1.yaml"}, true},
		{"three args", []string{"file1.yaml", "file2.yaml", "file3.yaml"}, true},
		{"two args", []string{"file1.yaml", "file2.yaml"}, false}, // This will fail at file read, but args validation passes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCommand()
			err := cmd.Args(cmd, tt.args)

			if (err != nil) != tt.expectError {
				t.Errorf("Args validation error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestRunCommandWithInvalidColorFlag(t *testing.T) {
	cmd := newRootCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set invalid color flag
	cfg := &config{color: "invalid"}

	err := runCommand(cmd, []string{"file1.yaml", "file2.yaml"}, cfg)

	if err == nil {
		t.Error("expected error for invalid color flag")
	}

	if !strings.Contains(err.Error(), "invalid color value") {
		t.Errorf("expected error to contain 'invalid color value', got %v", err)
	}
}

func TestRunCommandWithMutuallyExclusiveFlags(t *testing.T) {
	cmd := newRootCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set mutually exclusive flags
	cfg := &config{
		color:    "auto",
		pathOnly: true,
		metadata: true,
	}

	err := runCommand(cmd, []string{"file1.yaml", "file2.yaml"}, cfg)

	if err == nil {
		t.Error("expected error for mutually exclusive flags")
	}

	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected error to contain 'mutually exclusive', got %v", err)
	}
}

func TestRunCommandWithValidFiles(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2.yaml")

	yaml1 := `name: Alice
age: 30`
	yaml2 := `name: Bob
age: 25`

	if err := os.WriteFile(file1, []byte(yaml1), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2, []byte(yaml2), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := newRootCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cfg := newConfig()
	err := runCommand(cmd, []string{file1, file2}, cfg)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name") || !strings.Contains(output, "age") {
		t.Errorf("expected output to contain diff information, got: %s", output)
	}
}

func TestRunCommandWithExitOnDifference(t *testing.T) {
	// Create temporary test files with differences
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2.yaml")

	yaml1 := `name: Alice`
	yaml2 := `name: Bob`

	if err := os.WriteFile(file1, []byte(yaml1), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2, []byte(yaml2), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := newRootCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cfg := &config{
		exitOnDifference: true,
		color:            "auto",
	}

	err := runCommand(cmd, []string{file1, file2}, cfg)

	if err == nil {
		t.Error("expected error when differences found and exitOnDifference is true")
	}

	if !strings.Contains(err.Error(), "differences found") {
		t.Errorf("expected error to contain 'differences found', got %v", err)
	}
}

func TestRunCommandWithNoDifferences(t *testing.T) {
	// Create temporary test files with no differences
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.yaml")
	file2 := filepath.Join(tmpDir, "file2.yaml")

	yaml := `name: Alice
age: 30`

	if err := os.WriteFile(file1, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd := newRootCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cfg := &config{
		exitOnDifference: true,
		color:            "auto",
	}

	err := runCommand(cmd, []string{file1, file2}, cfg)

	if err != nil {
		t.Errorf("unexpected error when no differences: %v", err)
	}

	// Output should be empty or just whitespace when no differences
	output := strings.TrimSpace(buf.String())
	if output != "" {
		t.Errorf("expected empty output when no differences, got: %s", output)
	}
}

func TestRunCommandWithNonExistentFiles(t *testing.T) {
	cmd := newRootCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cfg := newConfig()
	err := runCommand(cmd, []string{"nonexistent1.yaml", "nonexistent2.yaml"}, cfg)

	if err == nil {
		t.Error("expected error for non-existent files")
	}

	if !strings.Contains(err.Error(), "failed to compare files") {
		t.Errorf("expected error to contain 'failed to compare files', got %v", err)
	}
}

// TestExecute tests the Execute function by checking if it creates and runs the command
func TestExecute(t *testing.T) {
	// We can't easily test Execute() function directly because it calls os.Exit(1) on error
	// Instead, we test that newRootCommand() creates a valid command that Execute() would use
	cmd := newRootCommand()

	if cmd == nil {
		t.Error("newRootCommand() should not return nil")
		return
	}

	// Test that the command has the basic structure we expect
	if cmd.Use == "" {
		t.Error("command should have a Use field set")
	}
}
