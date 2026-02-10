# AGENTS.md

This file provides guidance to AI coding agents working with cirby.

## Project Overview

Cirby is a CLI tool that merges AI coding agent configuration files into the unified AGENTS.md standard. It's written in Go for cross-platform support and zero-dependency distribution.

## Build & Test

```bash
# Build
go build -o cirby .

# Run tests
go test ./...

# Cross-compile
GOOS=darwin GOARCH=arm64 go build -o cirby-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o cirby-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o cirby-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o cirby-windows-amd64.exe .
```

## Code Style

- Use `gofmt` for formatting
- Keep functions focused and small
- Error messages should be actionable
- No external dependencies (stdlib only)

## Project Structure

```
cirby/
├── main.go                    # CLI entry point
├── internal/cirby/cirby.go    # Core logic
├── AGENTS.md                  # This file
└── README.md                  # User documentation
```

## Key Design Decisions

1. **Idempotent** — Running multiple times produces the same result
2. **Git-safe** — Requires files to be committed before modifying
3. **Symlink-based** — Original files become symlinks, not copies
4. **Zero dependencies** — Only Go stdlib, single binary output

## Adding New Agent Support

To add support for a new coding agent, update `agentPatterns` in `internal/cirby/cirby.go`:

```go
{"New Agent", []string{"NEWAGENT.md", ".newagent"}, "symlink"},
```

Strategies:
- `symlink` — Replace with symlink to AGENTS.md
- `merge` — Merge content into AGENTS.md, keep original
- `keep` — Already AGENTS.md standard, do nothing
