# AGENTS.md

Guidance for agentic coding assistants working in `cirby`.

## Project Snapshot
- Language: Go (`go.mod` declares Go `1.25.7`)
- Binary: `cirby`
- Goal: unify AI agent config files into `AGENTS.md`
- Dependency policy: Go standard library only

## Repository Layout
```text
cirby/
├── main.go                 # CLI entrypoint + flags + exit codes
├── internal/cirby/cirby.go # scan, merge, safety checks, symlinks
├── README.md               # user-facing docs
└── AGENTS.md               # agent guidance
```

## Core Behavior to Preserve
1. Idempotent runs: repeated execution should not create extra changes.
2. Git safety by default: block edits when agent config files are uncommitted.
3. Strategy handling: keep `keep`, merge `merge`, and symlink `symlink` behavior intact.
4. Deterministic output: preserve sorted ordering and stable content comparisons.
5. Dry-run is side-effect free.

## Build, Lint, and Test Commands
Run from repo root.

```bash
# Build
go build -o cirby .
go build -race -o cirby .

# Test all
go test ./...
go test -v ./...
go test -race ./...

# Single package
go test ./internal/cirby

# Single test (exact name)
go test ./internal/cirby -run '^TestName$' -v

# Subset by regex
go test ./internal/cirby -run 'TestScan|TestMerge' -v

# Benchmarks
go test ./internal/cirby -bench '^BenchmarkName$' -run '^$'

# Coverage
go test ./... -cover
go test ./... -coverprofile=coverage.out

# Lint-equivalent (stdlib only)
go vet ./...

# Format / formatting check
gofmt -w .
gofmt -l .
```

Notes:
- There is no external linter config (for example, no `golangci-lint`).
- If no tests exist yet, still keep single-test commands in this format.

## Manual Smoke Checks
Use these after behavior changes:

```bash
./cirby --help
./cirby --version
./cirby --dry-run --verbose
```

## Cross-Compilation
```bash
GOOS=darwin GOARCH=arm64 go build -o cirby-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o cirby-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o cirby-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o cirby-windows-amd64.exe .
```

## Code Style Guidelines

### Formatting and File Hygiene
- Always run `gofmt`; do not hand-format spacing/alignment.
- Default to ASCII unless the file already requires Unicode.
- Keep functions focused; split when responsibilities diverge.
- Prefer deterministic output to avoid noisy diffs.

### Imports
- Follow Go default import grouping/order (`gofmt` output).
- Keep stdlib imports first; add module-local imports only when needed.
- Do not add third-party dependencies unless explicitly requested.

### Types and Data Modeling
- Use structs for explicit payloads (`Options`, `AgentConfig` pattern).
- Prefer descriptive field names aligned to domain meaning.
- Add doc comments for new exported types/functions.
- Prefer concrete types; introduce interfaces only for real behavior boundaries.

### Naming
- Use `camelCase` for unexported identifiers and `PascalCase` for exported ones.
- Use verb-first function names for actions (`scanConfigs`, `createSymlink`).
- Keep booleans clear (`DryRun`, `Force`, `Verbose`).
- Avoid cryptic abbreviations.

### Error Handling
- Return errors; avoid `panic` in normal control flow.
- Wrap with context: `fmt.Errorf("context: %w", err)`.
- Keep user-facing errors actionable.
- In `main`, print to stderr and exit non-zero on failure.

### Control Flow and State
- Prefer early returns to reduce nesting.
- Handle no-op states explicitly (`nothing found`, `already in sync`).
- Keep dry-run logic and mutating logic clearly separated.

### Filesystem and Symlink Semantics
- Preserve safety checks before removing/replacing files.
- Use `Lstat` when symlink identity matters.
- Preserve relative symlink targets for nested files.
- Do not silently ignore read/write failures.

### CLI and UX
- Keep `--help` text aligned with real flags and behavior.
- Keep README examples aligned with actual output semantics.
- Keep output concise and consistent (`[ok]`, `[skip]`, `[error]`).

## Testing Expectations for New Changes
- Add `_test.go` coverage for merge, symlink, and safety logic changes.
- Prefer table-driven tests for pattern/strategy behavior.
- Cover both normal and dry-run execution paths.
- Include edge cases (pre-existing symlinks, empty content, duplicate content).

## Agent Change Checklist
1. Implement the smallest change that solves the request.
2. Run `gofmt -w` on edited Go files.
3. Run `go test ./...` and `go vet ./...`.
4. If CLI behavior changed, run `./cirby --help` and `./cirby --dry-run --verbose`.
5. Update `README.md` if flags, output, or supported files changed.

## Adding New Agent Support
Update `agentPatterns` in `internal/cirby/cirby.go`:

```go
{"New Agent", []string{"NEWAGENT.md", ".newagent"}, "symlink"},
```

Strategies:
- `symlink`: replace original file with symlink to `AGENTS.md`
- `merge`: append content into `AGENTS.md`, keep source file
- `keep`: treat file as already standardized and do nothing

Also update user-facing docs (`README.md` and help text in `main.go`).

## Cursor and Copilot Rule Files
Current repository scan:
- No `.cursorrules` file
- No `.cursor/rules/` entries
- No `.github/copilot-instructions.md`

If these files are added later, merge their guidance into this document.
