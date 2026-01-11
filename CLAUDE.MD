# yamldiff - Claude Code Guide

## Project Overview

`yamldiff` is a Go-based CLI tool and library for comparing YAML files and highlighting structural changes. It provides rich visual output showing additions, deletions, and modifications in YAML structures.

## Architecture

The project follows a standard Go CLI architecture:

```
yamldiff/
├── main.go              # Entry point
├── cmd/                 # CLI command implementation
│   ├── root.go          # Cobra root command and flags
│   └── root_test.go     # CLI tests
├── pkg/diff/            # Core diff library
│   ├── compare.go       # YAML comparison logic
│   ├── formatter.go     # Output formatting (plain, metadata, paths-only)
│   ├── types.go         # Core types (Diff, DiffType, etc.)
│   └── *_test.go        # Unit tests
└── examples/            # Example YAML files for testing
```

## Key Components

### Core Library (`pkg/diff/`)

- **compare.go** - Main comparison logic using `goccy/go-yaml` for parsing
- **types.go** - Core types:
  - `Diff` - Represents a single difference with path, type, and values
  - `DiffType` - Added, Deleted, Modified
  - `Diffs` - Collection of differences with formatting methods
- **formatter.go** - Output formatting with color support using `fatih/color`:
  - `Plain` - Default colored output
  - `Metadata` - Includes line numbers and node types
  - `PathsOnly` - Shows only the paths of differences

### CLI (`cmd/`)

- **root.go** - Cobra command with flags:
  - `--metadata` / `-m` - Include line numbers and node types
  - `--path-only` / `-p` - Show only paths
  - `--stat` / `-s` - Show only diffstat summary
  - `--ignore-order` / `-i` - Ignore sequence order
  - `--exit-code` / `-e` - Exit with non-zero on differences
  - `--color` - Control color output (always/never/auto)

## Dependencies

- `github.com/goccy/go-yaml` - YAML parsing with position tracking
- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Terminal color output
- `github.com/stretchr/testify` - Testing utilities

## Development Workflow

### Building

```bash
go build -o yamldiff .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/diff/
```

### Installation

```bash
go install github.com/semihbkgr/yamldiff@latest
```

## Code Conventions

1. **Error handling** - All errors are properly propagated and wrapped with context
2. **Testing** - Each package has comprehensive unit tests (`*_test.go`)
3. **Color output** - Use `fatih/color` package, respect `--color` flag
4. **YAML parsing** - Use `goccy/go-yaml` for position tracking support
5. **Documentation** - Exported functions and types have Go doc comments

## Making Changes

### Adding New Features

1. **Core diff logic** - Modify `pkg/diff/compare.go`
2. **Output formatting** - Update `pkg/diff/formatter.go`
3. **CLI flags** - Add to `cmd/root.go`
4. **Always add tests** - Include unit tests in corresponding `*_test.go` files

### After Each Code Change

Always run these commands after making code changes:

```bash
go fmt ./...
golangci-lint run ./...
```

### Testing Changes

- Run tests before committing: `go test ./...`
- Ensure coverage doesn't decrease significantly
- Test CLI manually with example files in `examples/`

### Common Tasks

**Add a new output format:**
1. Define format constant in `pkg/diff/types.go`
2. Implement formatting logic in `pkg/diff/formatter.go`
3. Add CLI flag in `cmd/root.go` if needed
4. Add tests in `pkg/diff/formatter_test.go`

**Modify comparison logic:**
1. Update `pkg/diff/compare.go`
2. Add test cases in `pkg/diff/compare_test.go`
3. Test with real YAML files in `examples/`

## CI/CD

The project uses GitHub Actions:
- **Go CI** - Runs tests and checks on push
- **Codecov** - Tracks test coverage

## Important Notes

- The tool uses structural comparison, not textual diff
- Color output is automatically disabled for non-TTY environments
- The `--ignore-order` flag affects array/sequence comparison
- Exit codes: 0 (no diff), 1 (has diff with --exit-code), 2 (error)
