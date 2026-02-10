package main

import (
	"fmt"
	"os"

	"github.com/poshboytl/cirby/internal/cirby"
)

const version = "0.2.0"

func main() {
	args := os.Args[1:]

	// Parse flags and agent
	opts := cirby.Options{
		DryRun:  false,
		Force:   false,
		Verbose: false,
		Agent:   "",
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
			if arg[0] == '-' {
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
				printHelp()
				os.Exit(1)
			}
			// Positional argument = agent name
			opts.Agent = arg
		}
	}

	if err := cirby.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`cirby - Merge AI coding agent configs into AGENTS.md

Usage: cirby [agent] [options]

Arguments:
  agent              Agent to use for smart merge:
                     claude, opencode, gemini, cursor, codex, aider
                     If not specified, auto-detects available agents

Options:
  --dry-run, -n      Preview changes without modifying files
  --force, -f        Skip git uncommitted changes check
  --verbose, -v      Show detailed output
  --version          Show version
  --help, -h         Show this help

Examples:
  cirby              # Auto-detect agent for merge
  cirby claude       # Use Claude Code for merge
  cirby gemini       # Use Gemini CLI for merge
  cirby --dry-run    # Preview what would be done

How it works:
  1. Scans for agent config files (CLAUDE.md, GEMINI.md, .cursorrules, etc.)
  2. Uses an AI agent to intelligently merge content into AGENTS.md
  3. Creates symlinks so each tool finds its expected file

Learn more: https://github.com/poshboytl/cirby`)
}
