# GEMINI.md - Project Context

## Project Overview

**Cirby** (Code + Kirby) is a Go-based CLI tool designed to unify AI coding agent configurations. It scans a project for various agent-specific instruction files (like `CLAUDE.md`, `.cursorrules`, `.windsurfrules`, etc.), merges their content into a single `AGENTS.md` file, and replaces the original files with symlinks to the unified `AGENTS.md`.

This ensures that all AI agents used by a team receive consistent instructions while maintaining compatibility with the specific filenames each tool expects.

### Main Technologies
- **Language:** Go (v1.25.7)
- **Tooling:** Git (for safety checks)
- **Standards:** [AGENTS.md](https://agents.md)

### Architecture
- `main.go`: Handles CLI argument parsing and entry point logic.
- `internal/cirby/`: Contains the core logic for:
    - Scanning for configuration files based on predefined patterns.
    - Merging content with deduplication.
    - Managing symlinks (including relative path calculations).
    - Validating git status to prevent accidental data loss.

---

## Building and Running

### Commands
- **Build:** `go build -o cirby .`
- **Run:** `./cirby`
- **Install:** `go install github.com/poshboytl/cirby@latest`

### Usage Options
- `cirby`: Merge configs and create symlinks.
- `cirby --dry-run`: Preview changes without modifying the filesystem.
- `cirby --force`: Skip the git uncommitted changes safety check.
- `cirby --verbose`: Show detailed scanning and merging information.

---

## Development Conventions

- **Safety First:** The tool requires agent configuration files to be committed in Git before it will modify them (unless `--force` is used).
- **Idempotency:** Running `cirby` multiple times is intended to be safe; it detects existing symlinks and identical content to avoid redundant operations.
- **Structure:** Core business logic should reside in the `internal/cirby` package to keep `main.go` focused on the CLI interface.
- **Idiomatic Go:** Follow standard Go formatting and naming conventions.

## TODOs / Future Work
- [ ] Add unit tests for merging logic and symlink creation.
- [ ] Support more AI agent configuration formats as they emerge.
