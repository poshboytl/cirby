# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cirby is a Go CLI tool that merges AI coding agent configuration files (CLAUDE.md, GEMINI.md, .cursorrules, etc.) into a unified `AGENTS.md` format, then symlinks the originals back to it. Go standard library only — no external dependencies.

## Build, Test, and Lint

```bash
go build -o cirby .              # Build
go test ./...                    # Test all
go test ./internal/cirby -run '^TestName$' -v  # Single test
go vet ./...                     # Lint
gofmt -w .                       # Format
```

After behavior changes, smoke-test: `./cirby --help`, `./cirby --version`, `./cirby --dry-run --verbose`

## Architecture

Two files compose the entire codebase:

- **`main.go`** — CLI entry point: parses flags/args, delegates to `cirby.Run(opts)`, handles exit codes.
- **`internal/cirby/cirby.go`** — All core logic:
  - `Run()` orchestrates the full flow: scan → git safety check → filter symlinks → select agent → build prompt → execute agent → create symlinks
  - `scanConfigs()` discovers config files via glob patterns defined in `agentPatterns`
  - `selectAgent()` auto-detects or validates CLI-specified merge agent from `supportedAgents`
  - `executeAgent()` shells out to the selected agent's CLI
  - `buildMergePrompt()` / `buildMergeIntoExistingPrompt()` construct LLM prompts for merging
  - `createSymlink()` replaces original config files with relative symlinks to AGENTS.md

## Key Invariants

1. **Idempotent** — repeated runs produce no extra changes
2. **Git safety** — blocks edits when agent configs are uncommitted (bypass with `--force`)
3. **Dry-run is side-effect free** — must never modify files
4. **Deterministic output** — sorted ordering, stable comparisons

## Adding New Agent Support

Add to `agentPatterns` in `internal/cirby/cirby.go` for detection, and to `supportedAgents` if usable as a merge agent. Update `README.md` and help text in `main.go`.

## Code Style

- `gofmt` for formatting, no hand-formatting
- Return errors with context (`fmt.Errorf("context: %w", err)`), no `panic`
- Early returns over nesting
- Verb-first function names (`scanConfigs`, `createSymlink`)
- Use `Lstat` when symlink identity matters; preserve relative symlink targets
- Keep output format consistent: `[ok]`, `[skip]`, `[error]`

## Change Checklist

1. Implement the smallest change that solves the request
2. Run `gofmt -w` on edited files
3. Run `go test ./...` and `go vet ./...`
4. If CLI behavior changed, run smoke checks and update `README.md`
