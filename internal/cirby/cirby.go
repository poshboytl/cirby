package cirby

import (
	"bufio"
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
	Agent   string
}

// AgentConfig represents a discovered agent configuration file
type AgentConfig struct {
	Path    string
	Agent   string
	Content string
}

// SupportedAgent represents a coding agent that can be used for merging
type SupportedAgent struct {
	Name    string
	Command string
	Args    func(prompt string) []string
}

// Agent patterns to scan for
var agentPatterns = []struct {
	Name     string
	Patterns []string
}{
	{"Claude Code", []string{"CLAUDE.md"}},
	{"Cursor (legacy)", []string{".cursorrules"}},
	{"Cursor (rules)", []string{".cursor/rules/*.mdc"}},
	{"Windsurf", []string{".windsurfrules"}},
	{"GitHub Copilot", []string{".github/copilot-instructions.md"}},
	{"Gemini CLI", []string{"GEMINI.md"}},
	{"Codex", []string{"CODEX.md"}},
}

// Supported agents for merging
var supportedAgents = []SupportedAgent{
	{
		Name:    "claude",
		Command: "claude",
		Args:    func(prompt string) []string { return []string{"-p", prompt, "--allowedTools", "Edit,Write,Read"} },
	},
	{
		Name:    "opencode",
		Command: "opencode",
		Args:    func(prompt string) []string { return []string{"-p", prompt} },
	},
	{
		Name:    "gemini",
		Command: "gemini",
		Args:    func(prompt string) []string { return []string{"-p", prompt} },
	},
	{
		Name:    "cursor",
		Command: "cursor-agent",
		Args:    func(prompt string) []string { return []string{"chat", prompt} },
	},
	{
		Name:    "codex",
		Command: "codex",
		Args:    func(prompt string) []string { return []string{prompt} },
	},
	{
		Name:    "aider",
		Command: "aider",
		Args:    func(prompt string) []string { return []string{"--message", prompt, "--yes"} },
	},
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
	agentsMDExists := false
	var agentsMDContent string
	if content, err := os.ReadFile("AGENTS.md"); err == nil {
		agentsMDExists = true
		agentsMDContent = string(content)
	}

	// Filter out files that are already symlinks to AGENTS.md
	var toProcess []AgentConfig
	for _, cfg := range configs {
		if isSymlinkToAgentsMD(cfg.Path) {
			if opts.Verbose {
				fmt.Printf("  [skip] %s (already symlinked)\n", cfg.Path)
			}
			continue
		}
		if cfg.Path == "AGENTS.md" {
			continue
		}
		toProcess = append(toProcess, cfg)
	}

	if len(toProcess) == 0 {
		fmt.Println("[ok] Already in sync. Nothing to do.")
		return nil
	}

	// If there are non-symlink files, we need to merge them (even if AGENTS.md exists)
	// Detect or use specified agent
	agent, err := selectAgent(opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		fmt.Println("\n[Dry Run] Would perform these actions:\n")
		if agentsMDExists {
			fmt.Printf("  - Use %s to merge %d new files INTO existing AGENTS.md\n", agent.Name, len(toProcess))
		} else {
			fmt.Printf("  - Use %s to merge %d files into new AGENTS.md\n", agent.Name, len(toProcess))
		}
		for _, cfg := range toProcess {
			fmt.Printf("  - Create symlink: %s -> AGENTS.md\n", cfg.Path)
		}
		fmt.Println("\nRun without --dry-run to apply changes.")
		return nil
	}

	// Build the merge prompt
	var prompt string
	if agentsMDExists {
		prompt = buildMergeIntoExistingPrompt(agentsMDContent, toProcess)
		fmt.Printf("Merging %d new files into existing AGENTS.md with %s...\n", len(toProcess), agent.Name)
	} else {
		prompt = buildMergePrompt(toProcess)
		fmt.Printf("Merging with %s...\n", agent.Name)
	}

	if opts.Verbose {
		fmt.Printf("Prompt:\n%s\n", prompt)
	}

	// Execute the agent
	if err := executeAgent(agent, prompt, opts); err != nil {
		return fmt.Errorf("agent merge failed: %w", err)
	}

	// Verify AGENTS.md exists
	if _, err := os.Stat("AGENTS.md"); os.IsNotExist(err) {
		return fmt.Errorf("agent did not create/update AGENTS.md")
	}

	if agentsMDExists {
		fmt.Println("[ok] Updated AGENTS.md")
	} else {
		fmt.Println("[ok] Created AGENTS.md")
	}

	// Create symlinks
	for _, cfg := range toProcess {
		if err := createSymlink(cfg.Path, opts); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", cfg.Path, err)
		}
		fmt.Printf("[ok] Symlinked %s -> AGENTS.md\n", cfg.Path)
	}

	fmt.Println("\nDone!")
	return nil
}

func selectAgent(opts Options) (SupportedAgent, error) {
	// If agent specified, find it
	if opts.Agent != "" {
		for _, a := range supportedAgents {
			if a.Name == opts.Agent {
				// Check if it's installed
				if _, err := exec.LookPath(a.Command); err != nil {
					return SupportedAgent{}, fmt.Errorf("%s is not installed or not in PATH", a.Name)
				}
				return a, nil
			}
		}
		return SupportedAgent{}, fmt.Errorf("unknown agent: %s (supported: claude, opencode, gemini, cursor, codex, aider)", opts.Agent)
	}

	// Auto-detect available agents
	var available []SupportedAgent
	for _, a := range supportedAgents {
		if _, err := exec.LookPath(a.Command); err == nil {
			available = append(available, a)
		}
	}

	if len(available) == 0 {
		return SupportedAgent{}, fmt.Errorf("no supported agent found. Please install one of: claude, opencode, gemini, cursor, codex, aider")
	}

	if len(available) == 1 {
		fmt.Printf("Using %s to merge config files...\n", available[0].Name)
		return available[0], nil
	}

	// Multiple agents available, let user choose
	fmt.Println("Cirby needs an AI agent to intelligently merge your config files.")
	fmt.Println("Multiple agents detected on your system:\n")
	for i, a := range available {
		fmt.Printf("  %d) %s\n", i+1, a.Name)
	}
	fmt.Printf("\nWhich agent would you like to use? [1]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return available[0], nil
	}

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(available) {
		return available[0], nil
	}

	return available[choice-1], nil
}

func buildMergePrompt(configs []AgentConfig) string {
	var files []string
	for _, cfg := range configs {
		files = append(files, cfg.Path)
	}

	return fmt.Sprintf(`Read the following AI agent configuration files in this project:
%s

These files contain instructions for different AI coding agents. Please:
1. Analyze the content of each file
2. Create a unified AGENTS.md file that combines the best instructions from all files
3. Remove duplicate information
4. Use agent-agnostic language (don't say "Claude should..." or "Gemini should...")
5. Keep the merged content concise and well-organized
6. Write the result to AGENTS.md in the current directory

The AGENTS.md file should follow this structure:
- Project Overview
- Build & Test Commands
- Code Style Guidelines
- Architecture Notes
- Any other relevant sections

Please create the AGENTS.md file now.`, strings.Join(files, "\n"))
}

func buildMergeIntoExistingPrompt(existingContent string, configs []AgentConfig) string {
	var files []string
	for _, cfg := range configs {
		files = append(files, cfg.Path)
	}

	return fmt.Sprintf(`The project already has an AGENTS.md file with the following content:

---
%s
---

New agent configuration files have been found that need to be merged:
%s

Please:
1. Read the new configuration files
2. Analyze what information they contain that is NOT already in AGENTS.md
3. Merge any new, unique information into AGENTS.md
4. Remove any duplicates
5. Use agent-agnostic language (don't say "Claude should..." or "Gemini should...")
6. Keep the content well-organized
7. Update the AGENTS.md file with the merged content

Important: Preserve the existing structure and content of AGENTS.md, only ADD new information that wasn't there before.

Please update the AGENTS.md file now.`, existingContent, strings.Join(files, "\n"))
}

func executeAgent(agent SupportedAgent, prompt string, opts Options) error {
	args := agent.Args(prompt)
	cmd := exec.Command(agent.Command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if opts.Verbose {
		fmt.Printf("Running: %s %s\n", agent.Command, strings.Join(args, " "))
	}

	return cmd.Run()
}

func checkGitStatus(opts Options) error {
	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
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

Or use --force to skip this check`,
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
					Path:    match,
					Agent:   agent.Name,
					Content: string(content),
				})
			}
		}
	}

	// Also check for AGENTS.md
	if _, err := os.Stat("AGENTS.md"); err == nil {
		if opts.Verbose {
			fmt.Println("  [ok] AGENTS.md (standard)")
		}
		configs = append(configs, AgentConfig{
			Path:  "AGENTS.md",
			Agent: "AGENTS.md",
		})
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

func createSymlink(path string, opts Options) error {
	// Remove existing file
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing existing file: %w", err)
	}

	// Calculate relative path to AGENTS.md from the symlink location
	dir := filepath.Dir(path)
	var target string
	if dir == "." {
		target = "AGENTS.md"
	} else {
		relPath, err := filepath.Rel(dir, "AGENTS.md")
		if err != nil {
			target = "AGENTS.md"
		} else {
			target = relPath
		}
	}

	return os.Symlink(target, path)
}
