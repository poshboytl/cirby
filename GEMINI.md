# GEMINI.md - Cirby Context

This file provides instructional context for Gemini CLI when working on the Cirby project.

## Project Overview

**Cirby** (Code + Kirby) is a Go-based CLI utility that unifies various AI coding agent configurations into a single `AGENTS.md` file. It aims to reduce friction in teams using multiple AI tools by maintaining a single source of truth for agent instructions.

### Core Functionality
1.  **Scan:** Detects configuration files from supported agents (Claude, Cursor, Windsurf, Copilot, Gemini, Codex).
2.  **Merge:** Uses an available AI agent to intelligently combine these configurations into a unified `AGENTS.md`, removing duplicates and using agent-agnostic language.
3.  **Symlink:** Replaces the original configuration files with symlinks to `AGENTS.md`.

### Main Technologies
- **Language:** Go (v1.25.7+)
- **Integration:** Interacts with other AI CLIs (e.g., `claude`, `gemini`, `aider`) to perform the smart merge.
- **Version Control:** Uses `git` for safety checks to ensure no uncommitted changes are lost.

## Architecture

- `main.go`: CLI entry point, versioning, and flag parsing.
- `internal/cirby/`: Contains the core implementation logic.
    - `cirby.go`: Implements scanning, agent execution, prompt building, and symlink management.

## Building and Running

### Key Commands
- **Build:** `go build -o cirby .`
- **Run:** `./cirby` (auto-detects agent) or `./cirby [agent]` (e.g., `cirby gemini`).
- **Dry Run:** `cirby --dry-run` to preview actions.
- **Verbose Mode:** `cirby --verbose` for detailed output.
- **Test:** (TODO) No unit tests are currently implemented.

## Development Conventions

### Coding Style
- Follow standard Go idioms and `gofmt` formatting.
- Core logic resides in the `internal/` directory to prevent external usage as a library.
- Error handling should be explicit and wrapped with context using `fmt.Errorf`.

### Safety First
- **Git Check:** By default, Cirby refuses to run if there are uncommitted changes in agent configuration files. This is a crucial safety feature.
- **Symlink Awareness:** The tool detects if a file is already a symlink to `AGENTS.md` to avoid redundant work.

### Merging Logic
- When merging, prioritize agent-agnostic language.
- The default structure for `AGENTS.md` should include:
    - Project Overview
    - Build & Test Commands
    - Code Style Guidelines
    - Architecture Notes

## Key Files
- `main.go`: CLI wrapper.
- `internal/cirby/cirby.go`: The "brain" of the application.
- `AGENTS.md`: The unified configuration file (generated/maintained by the tool).
- `go.mod`: Project dependencies.
