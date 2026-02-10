# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cirby is a Go CLI tool that merges AI coding agent configuration files (CLAUDE.md, .cursorrules, .windsurfrules, etc.) into the unified AGENTS.md standard, then creates symlinks so each tool still finds its expected file.

## Build & Test

```bash
go build -o cirby .       # Build
go test ./...              # Run all tests
```

Cross-compile:
```bash
GOOS=darwin GOARCH=arm64 go build -o cirby-darwin-arm64 .
GOOS=linux GOARCH=amd64 go build -o cirby-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o cirby-windows-amd64.exe .
```

## Architecture

Two-file structure with clear separation:

- `main.go` — CLI entry point. Parses flags (`--dry-run`, `--force`, `--verbose`, `--version`, `--help`) into an `Options` struct, then calls `cirby.Run(opts)`.
- `internal/cirby/cirby.go` — All core logic. The pipeline is: `checkGitStatus()` (safety gate) -> `scanConfigs()` (glob-based file discovery) -> `buildMergedContent()` (content assembly with SHA256 deduplication) -> write AGENTS.md -> `createSymlink()` for each agent file.

### Agent Pattern Registry

New agent support is added by appending to the `agentPatterns` slice in `internal/cirby/cirby.go`. Each entry has a name, glob patterns, and a strategy:

- **symlink** — Replace the file with a symlink to AGENTS.md
- **merge** — Merge content into AGENTS.md, keep original file (used for `.cursor/rules/*.mdc`, `CODEX.md`)
- **keep** — Already AGENTS.md standard, no action needed

## Key Constraints

- **Zero external dependencies** — stdlib only, single static binary
- **Idempotent** — Running multiple times produces identical results
- **Git-safe** — Requires agent config files to be committed before modifying (unless `--force`)
- Use `gofmt` for formatting
- Errors should be wrapped with context: `fmt.Errorf("operation: %w", err)`
