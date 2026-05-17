# AGENTS.md

This repository follows the AGENTS.md convention. `AGENTS.md` is the canonical
agent guide; `CLAUDE.md` carries the same content for Claude Code compatibility.

## Project Overview

`yamldiff` is a Go CLI and library for structural YAML comparison. It parses YAML
with `github.com/goccy/go-yaml`, compares AST nodes, and formats additions,
deletions, and modifications for terminal output or the WASM playground.

Keep README-facing content in `README.md`. Keep this file focused on instructions
that help coding agents build, test, and change the project safely.

## Repository Layout

- `main.go` wires the executable to `cmd.Execute()`.
- `cmd/` contains Cobra CLI setup, flag validation, exit behavior, and CLI tests.
- `pkg/diff/` contains the public diff library, comparison logic, formatting,
  color helpers, and unit/example tests.
- `cmd/wasm/` exposes the diff library to JavaScript through WebAssembly.
- `web/` is the static playground served by GitHub Pages.
- `examples/` contains sample YAML files for manual CLI checks.
- `.github/workflows/` defines CI, Pages deployment, and release automation.

## Development Commands

Prefer the Taskfile commands when `task` is available:

- Build the CLI: `task build`
- Run tests: `task test`
- Run tests with coverage: `task test COVER=true`
- Format Go code: `task fmt`
- Run lint: `task lint`
- Build WASM assets: `task wasm:build`
- Serve the playground locally: `task wasm:serve`

Equivalent direct commands are:

- `go build .`
- `go test ./...`
- `go test -race -coverprofile=coverage.txt ./...`
- `go fmt ./...`
- `go vet ./...`
- `golangci-lint run ./...`
- `GOOS=js GOARCH=wasm go build -o web/yamldiff.wasm ./cmd/wasm`

## Required Checks

Before finishing Go changes, run:

1. `go fmt ./...` or `task fmt`
2. `go test ./...` or `task test`
3. `golangci-lint run ./...` or `task lint`, when `golangci-lint` is available

If the change touches CLI behavior, also run at least one manual example, such as:

```bash
go run . --metadata examples/pod-v1.yaml examples/pod-v2.yaml
```

If the change touches WASM or `web/`, run `task wasm:build` and, when practical,
serve `web/` locally to smoke test the playground.

## Code Style

- Use idiomatic Go and keep all Go files `gofmt` formatted.
- Keep exported functions, types, and options documented with Go doc comments.
- Return errors with enough context for CLI users and library callers.
- Preserve the separation between `cmd/` and `pkg/diff/`: CLI flag parsing belongs
  in `cmd/`, reusable comparison and formatting behavior belongs in `pkg/diff/`.
- Prefer extending the existing option patterns: `CompareOption` for comparison
  behavior and `FormatOption` for formatting behavior.
- Use `goccy/go-yaml` AST APIs for YAML parsing and position-aware behavior.
- Use the existing color abstraction in `pkg/diff` instead of adding direct color
  handling in callers.
- Do not change public API names, output format, or exit-code behavior casually;
  treat those as user-facing contracts.

## Testing Guidance

- Add or update focused tests for every behavior change.
- Put core comparison tests in `pkg/diff/compare_test.go`.
- Put formatter/output tests in `pkg/diff/formatter_test.go`.
- Put CLI flag and exit-behavior tests in `cmd/root_test.go`.
- Use `pkg/diff/testdata/` for reusable YAML fixtures, especially multi-document
  YAML cases.
- Include tests for edge cases such as empty documents, null values, sequence
  ordering, path-only output, metadata output, and mutually exclusive CLI flags
  when relevant.
- Keep example tests in sync with public library usage.

## Diff Semantics And Gotchas

- The tool performs structural YAML comparison, not textual diffing.
- Mapping keys are matched by key string, then their values are compared.
- Sequence items are compared positionally unless `IgnoreSeqOrder` or
  `--ignore-order` is used.
- `--ignore-order` matches whole sequence elements first, then compares unmatched
  elements, so changes inside complex list items may still appear as modifications.
- Multiple YAML documents are compared by document index and formatted with `---`
  separators.
- Comments are not part of the structural comparison.
- YAML directives, tags, anchors, aliases, and merge keys have intentionally
  limited comparison support; check `compareNodes` before changing this behavior.
- `--metadata`, `--path-only`, and `--stat` have mutual-exclusion rules enforced
  in `cmd/root.go`.
- Exit codes are user-facing: normal success is `0`, `--exit-code` with
  differences exits non-zero, and operational errors exit non-zero.
- Color is controlled by `--color=always|never|auto`; non-TTY output should remain
  usable.

## Dependency And Release Notes

- Run `go mod tidy` after adding, removing, or upgrading dependencies.
- Avoid adding dependencies unless they clearly simplify core behavior.
- Releases are handled by GoReleaser from tags matching `v*.*.*`.
- The GitHub Pages workflow builds `cmd/wasm` and publishes the `web/` directory.
- Do not commit local build artifacts unless the project explicitly tracks them.

## Documentation Updates

Update `README.md` when a change affects:

- CLI flags, usage, examples, or exit behavior
- Public library APIs or examples
- Playground behavior visible to users
- Installation or release instructions

## PR Guidance

- Keep changes scoped to the requested behavior.
- Mention tests and manual checks performed.
- Call out any skipped check and why it was skipped.
- For behavior changes, describe the user-visible effect rather than only the code
  mechanics.
