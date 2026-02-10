package cirby

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Options holds CLI options
type Options struct {
	DryRun  bool
	Force   bool
	Verbose bool
}

// AgentConfig represents a discovered agent configuration file
type AgentConfig struct {
	Path     string
	Agent    string
	Content  string
	Strategy string // "merge", "symlink", "keep"
}

// Agent patterns to scan for
var agentPatterns = []struct {
	Name     string
	Patterns []string
	Strategy string
}{
	{"Claude Code", []string{"CLAUDE.md", "**/CLAUDE.md"}, "symlink"},
	{"Cursor (legacy)", []string{".cursorrules"}, "symlink"},
	{"Cursor (rules)", []string{".cursor/rules/*.mdc"}, "merge"},
	{"Windsurf", []string{".windsurfrules"}, "symlink"},
	{"GitHub Copilot", []string{".github/copilot-instructions.md"}, "symlink"},
	{"Gemini CLI", []string{"GEMINI.md"}, "symlink"},
	{"Codex", []string{"CODEX.md"}, "merge"},
	{"OpenCode/AMP", []string{"AGENTS.md"}, "keep"},
}

// Run executes the main cirby logic
func Run(opts Options) error {
	// Check git status unless --force
	if !opts.Force {
		if err := checkGitStatus(opts); err != nil {
			return err
		}
	}

	// Scan for config files
	configs, err := scanConfigs(opts)
	if err != nil {
		return fmt.Errorf("scanning configs: %w", err)
	}

	if len(configs) == 0 {
		fmt.Println("No agent configuration files found.")
		return nil
	}

	// Check if AGENTS.md already exists
	existingAgentsMD := ""
	if content, err := os.ReadFile("AGENTS.md"); err == nil {
		existingAgentsMD = string(content)
	}

	// Separate configs by strategy
	var toMerge []AgentConfig
	var toSymlink []AgentConfig

	for _, cfg := range configs {
		switch cfg.Strategy {
		case "merge":
			toMerge = append(toMerge, cfg)
		case "symlink":
			// Only symlink if not already a symlink to AGENTS.md
			if !isSymlinkToAgentsMD(cfg.Path) {
				toSymlink = append(toSymlink, cfg)
			}
		case "keep":
			// AGENTS.md itself, nothing to do
			if opts.Verbose {
				fmt.Printf("  [ok] %s (already standard)\n", cfg.Path)
			}
		}
	}

	// Build merged content
	mergedContent := buildMergedContent(existingAgentsMD, toMerge, toSymlink, opts)

	// Check if anything changed
	if existingAgentsMD != "" {
		existingHash := hashContent(existingAgentsMD)
		newHash := hashContent(mergedContent)
		if existingHash == newHash && len(toSymlink) == 0 {
			fmt.Println("[ok] Already in sync. Nothing to do.")
			return nil
		}
	}

	// Dry run - just show what would happen
	if opts.DryRun {
		fmt.Println("\n[Dry Run] Would perform these actions:\n")

		if mergedContent != existingAgentsMD {
			fmt.Println("  • Create/update AGENTS.md with merged content")
			if opts.Verbose {
				fmt.Println("\n--- AGENTS.md content preview ---")
				fmt.Println(mergedContent)
				fmt.Println("--- end preview ---")
			}
		}

		for _, cfg := range toSymlink {
			fmt.Printf("  • Create symlink: %s -> AGENTS.md\n", cfg.Path)
		}

		fmt.Println("\nRun without --dry-run to apply changes.")
		return nil
	}

	// Write AGENTS.md
	if mergedContent != existingAgentsMD {
		if err := os.WriteFile("AGENTS.md", []byte(mergedContent), 0644); err != nil {
			return fmt.Errorf("writing AGENTS.md: %w", err)
		}
		fmt.Println("[ok] Created/updated AGENTS.md")
	}

	// Create symlinks
	for _, cfg := range toSymlink {
		if err := createSymlink(cfg.Path, opts); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", cfg.Path, err)
		}
		fmt.Printf("[ok] Symlinked %s -> AGENTS.md\n", cfg.Path)
	}

	fmt.Println("\nDone!")
	return nil
}

func checkGitStatus(opts Options) error {
	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		// Not a git repo, skip check
		if opts.Verbose {
			fmt.Println("Not a git repository, skipping git check.")
		}
		return nil
	}

	// Check for uncommitted changes in relevant files
	cmd = exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("checking git status: %w", err)
	}

	if len(output) == 0 {
		return nil
	}

	// Check if any agent config files have uncommitted changes
	lines := strings.Split(string(output), "\n")
	var uncommitted []string

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		file := strings.TrimSpace(line[3:])
		if isAgentConfigFile(file) {
			uncommitted = append(uncommitted, file)
		}
	}

	if len(uncommitted) > 0 {
		return fmt.Errorf(`uncommitted changes detected in agent config files:
%s

Please commit first so you can rollback if needed:
  git add %s
  git commit -m "backup before cirby"

Or use --force to skip this check (not recommended)`,
			"  - "+strings.Join(uncommitted, "\n  - "),
			strings.Join(uncommitted, " "))
	}

	return nil
}

func isAgentConfigFile(path string) bool {
	base := filepath.Base(path)
	agentFiles := []string{
		"CLAUDE.md", "AGENTS.md", "CODEX.md", "GEMINI.md",
		".cursorrules", ".windsurfrules",
		"copilot-instructions.md",
	}
	for _, af := range agentFiles {
		if base == af {
			return true
		}
	}
	if strings.HasSuffix(path, ".mdc") && strings.Contains(path, ".cursor/rules/") {
		return true
	}
	return false
}

func scanConfigs(opts Options) ([]AgentConfig, error) {
	var configs []AgentConfig

	if opts.Verbose {
		fmt.Println("Scanning for agent configuration files...")
	}

	for _, agent := range agentPatterns {
		for _, pattern := range agent.Patterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				continue
			}

			for _, match := range matches {
				// Skip if it's a symlink pointing to AGENTS.md
				if isSymlinkToAgentsMD(match) {
					if opts.Verbose {
						fmt.Printf("  [skip] %s (symlink to AGENTS.md, skipping)\n", match)
					}
					continue
				}

				content, err := os.ReadFile(match)
				if err != nil {
					if opts.Verbose {
						fmt.Printf("  [error] %s (error reading: %v)\n", match, err)
					}
					continue
				}

				if opts.Verbose {
					fmt.Printf("  [ok] %s (%s)\n", match, agent.Name)
				}

				configs = append(configs, AgentConfig{
					Path:     match,
					Agent:    agent.Name,
					Content:  string(content),
					Strategy: agent.Strategy,
				})
			}
		}
	}

	// Sort for consistent output
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Path < configs[j].Path
	})

	return configs, nil
}

func isSymlinkToAgentsMD(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	target, err := os.Readlink(path)
	if err != nil {
		return false
	}
	return target == "AGENTS.md" || filepath.Base(target) == "AGENTS.md"
}

func buildMergedContent(existing string, toMerge []AgentConfig, toSymlink []AgentConfig, opts Options) string {
	var parts []string

	// If AGENTS.md already exists, use it as base
	if existing != "" {
		parts = append(parts, strings.TrimSpace(existing))
	}

	// Add content from files to merge (like .cursor/rules/*.mdc, CODEX.md)
	for _, cfg := range toMerge {
		// Skip if content is already in existing
		if existing != "" && strings.Contains(existing, strings.TrimSpace(cfg.Content)) {
			if opts.Verbose {
				fmt.Printf("  [skip] %s content already in AGENTS.md, skipping\n", cfg.Path)
			}
			continue
		}

		header := fmt.Sprintf("\n\n---\n\n<!-- Merged from %s (%s) -->\n\n", cfg.Path, cfg.Agent)
		parts = append(parts, header+strings.TrimSpace(cfg.Content))
	}

	// For symlink files, extract their content into AGENTS.md first (if no existing AGENTS.md)
	if existing == "" && len(toSymlink) > 0 {
		// Use the first file as the base
		base := toSymlink[0]
		parts = append(parts, fmt.Sprintf("# AGENTS.md\n\n%s", strings.TrimSpace(base.Content)))

		// Merge others
		for _, cfg := range toSymlink[1:] {
			if strings.TrimSpace(cfg.Content) == strings.TrimSpace(base.Content) {
				continue // Same content, skip
			}
			header := fmt.Sprintf("\n\n---\n\n<!-- Merged from %s (%s) -->\n\n", cfg.Path, cfg.Agent)
			parts = append(parts, header+strings.TrimSpace(cfg.Content))
		}
	}

	// If still empty, create a minimal AGENTS.md
	if len(parts) == 0 {
		return `# AGENTS.md

This file provides guidance to AI coding agents working with this project.

## Project Overview

<!-- Add your project description here -->

## Build & Test

<!-- Add build and test commands here -->

## Code Style

<!-- Add code style guidelines here -->
`
	}

	return strings.TrimSpace(strings.Join(parts, "")) + "\n"
}

func createSymlink(path string, opts Options) error {
	// Remove existing file (we've already checked git status)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing existing file: %w", err)
	}

	// Calculate relative path to AGENTS.md from the symlink location
	dir := filepath.Dir(path)
	var target string
	if dir == "." {
		target = "AGENTS.md"
	} else {
		// For nested files like .github/copilot-instructions.md
		relPath, err := filepath.Rel(dir, "AGENTS.md")
		if err != nil {
			target = "AGENTS.md"
		} else {
			target = relPath
		}
	}

	return os.Symlink(target, path)
}

func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}
