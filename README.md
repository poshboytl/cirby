# Cirby 🩷

Merge AI coding agent configs into the unified [AGENTS.md](https://agents.md) standard.

> **Cirby** = Code + Kirby — absorbs all your agent configs and unifies them into one.

## The Problem

Different AI coding agents use different configuration files:

| Agent | Config File |
|-------|-------------|
| Claude Code | `CLAUDE.md` |
| Cursor | `.cursorrules`, `.cursor/rules/*.mdc` |
| Windsurf | `.windsurfrules` |
| GitHub Copilot | `.github/copilot-instructions.md` |
| Codex | `CODEX.md` |
| OpenCode, AMP | `AGENTS.md` ✓ |

This creates friction when your team uses different tools, and makes it hard to maintain consistent instructions across agents.

## The Solution

`cirby` scans your project, merges all agent configs into `AGENTS.md`, and creates symlinks so each tool still finds its expected file.

```bash
$ cirby
Scanning for agent configuration files...
  ✓ CLAUDE.md (Claude Code)
  ✓ .cursorrules (Cursor)
  ✓ .windsurfrules (Windsurf)

✓ Created/updated AGENTS.md
✓ Symlinked CLAUDE.md → AGENTS.md
✓ Symlinked .cursorrules → AGENTS.md
✓ Symlinked .windsurfrules → AGENTS.md

🎉 Done!
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
# Merge configs and create symlinks
cirby

# Preview changes without modifying files
cirby --dry-run

# Skip git uncommitted changes check
cirby --force

# Verbose output
cirby --verbose
```

## Safety Features

### Git Protection

Cirby requires agent config files to be committed before merging. This ensures you can always rollback via git:

```bash
$ cirby
Error: uncommitted changes detected in agent config files:
  - CLAUDE.md (modified)
  - .cursorrules (untracked)

Please commit first so you can rollback if needed:
  git add CLAUDE.md .cursorrules
  git commit -m "backup before cirby"
```

Use `--force` to skip this check (not recommended).

### Idempotency

Running `cirby` multiple times is safe. It detects existing symlinks and unchanged content, only updating when necessary.

## How It Works

1. **Scan** — Find all agent config files in your project
2. **Merge** — Combine content into `AGENTS.md` (with deduplication)
3. **Symlink** — Replace original files with symlinks to `AGENTS.md`

After running cirby:
```
project/
├── AGENTS.md           ← The source of truth
├── CLAUDE.md           → symlink to AGENTS.md
├── .cursorrules        → symlink to AGENTS.md
├── .windsurfrules      → symlink to AGENTS.md
└── .github/
    └── copilot-instructions.md → symlink to ../AGENTS.md
```

## Contributing

Issues and PRs welcome! This project uses the AGENTS.md standard (of course).

## License

MIT

## Credits

Named after [Kirby](https://en.wikipedia.org/wiki/Kirby_(character)), the Nintendo character who absorbs abilities from others. Cirby absorbs your agent configs and unifies them. 🩷
