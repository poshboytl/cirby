package main

import (
	"fmt"
	"os"

	"github.com/poshboytl/cirby/internal/cirby"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]

	// Parse flags
	opts := cirby.Options{
		DryRun:  false,
		Force:   false,
		Verbose: false,
	}

	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			opts.DryRun = true
		case "--force", "-f":
			opts.Force = true
		case "--verbose", "-v":
			opts.Verbose = true
		case "--version":
			fmt.Printf("cirby v%s\n", version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
			printHelp()
			os.Exit(1)
		}
	}

	if err := cirby.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`cirby - Merge AI coding agent configs into AGENTS.md

Usage: cirby [options]

Options:
  --dry-run, -n    Preview changes without modifying files
  --force, -f      Skip git uncommitted changes check
  --verbose, -v    Show detailed output
  --version        Show version
  --help, -h       Show this help

Examples:
  cirby              # Merge configs and create symlinks
  cirby --dry-run    # Preview what would be done
  cirby --force      # Skip git safety check

Supported agents:
  - Claude Code (CLAUDE.md)
  - Cursor (.cursorrules, .cursor/rules/*.mdc)
  - Windsurf (.windsurfrules)
  - GitHub Copilot (.github/copilot-instructions.md)
  - Gemini CLI (GEMINI.md)
  - Codex (CODEX.md)
  - OpenCode, AMP (AGENTS.md - already standard)

Learn more: https://github.com/poshboytl/cirby`)
}
