# Cirby

Merge AI coding agent configs into the unified [AGENTS.md](https://agents.md) standard.

> **Cirby** = Code + Kirby — absorbs all your agent configs and unifies them into one.

## The Problem

Different AI coding agents use different configuration files. This creates friction when your team uses different tools.

## Supported Config Files

Cirby can detect and merge these agent configuration files:

| Agent | Config File |
|-------|-------------|
| Claude Code | `CLAUDE.md` |
| Cursor | `.cursorrules`, `.cursor/rules/*.mdc` |
| Windsurf | `.windsurfrules` |
| GitHub Copilot | `.github/copilot-instructions.md` |
| Gemini CLI | `GEMINI.md` |
| Codex | `CODEX.md` |
| OpenCode, AMP | `AGENTS.md` (already standard) |

## Supported Merge Agents

Cirby uses your installed coding agent to intelligently merge configs:

| Agent | CLI Command | Install |
|-------|-------------|---------|
| Claude Code | `claude` | [claude.ai/code](https://claude.ai/code) |
| OpenCode | `opencode` | [opencode.ai](https://opencode.ai) |
| Gemini CLI | `gemini` | [ai.google.dev](https://ai.google.dev/gemini-cli) |
| Cursor CLI | `cursor-agent` | [cursor.com/cli](https://cursor.com/cli) |
| Codex | `codex` | [openai.com/codex](https://openai.com/codex) |
| Aider | `aider` | [aider.chat](https://aider.chat) |

Cirby auto-detects which agents are installed. If multiple are available, you can choose or specify one.

## The Solution

`cirby` uses AI to intelligently merge your agent configs into a unified `AGENTS.md`, then creates symlinks so each tool still finds its expected file.

```bash
$ cirby

Scanning for agent config files...
  Found: CLAUDE.md, GEMINI.md, .cursorrules

Using claude for merge...
[ok] Created AGENTS.md

[ok] Symlinked CLAUDE.md -> AGENTS.md
[ok] Symlinked GEMINI.md -> AGENTS.md
[ok] Symlinked .cursorrules -> AGENTS.md

Done!
```

## Installation

### Go Install

```bash
go install github.com/poshboytl/cirby@latest
```

### From Source

```bash
git clone https://github.com/poshboytl/cirby.git
cd cirby
go build -o cirby .
```

### Download Binary

Check the [Releases](https://github.com/poshboytl/cirby/releases) page for pre-built binaries.

## Usage

```bash
cirby              # Auto-detect available agent
cirby claude       # Use Claude Code for merge
cirby gemini       # Use Gemini CLI for merge
cirby --dry-run    # Preview what would be done
cirby --force      # Skip git safety check
cirby --verbose    # Detailed output
```

## How It Works

1. **Scan** - Find all agent config files in your project
2. **Merge** - Use an AI agent to intelligently combine content (dedup, unify language)
3. **Symlink** - Replace original files with symlinks to `AGENTS.md`

After running cirby:
```
project/
├── AGENTS.md           <- The source of truth
├── CLAUDE.md           -> symlink to AGENTS.md
├── .cursorrules        -> symlink to AGENTS.md
└── GEMINI.md           -> symlink to AGENTS.md
```

## Safety Features

### Git Protection

Cirby requires agent config files to be committed before modifying. This ensures you can always rollback via git.

### If AGENTS.md Already Exists

Cirby skips the merge step and only creates symlinks. Your existing `AGENTS.md` is preserved.

## Contributing

Issues and PRs welcome!

## License

MIT

## Credits

Named after [Kirby](https://en.wikipedia.org/wiki/Kirby_(character)), the Nintendo character who absorbs abilities from others.
