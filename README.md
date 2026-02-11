# obsidian-go-mcp

[![CI](https://github.com/zach-snell/obsidian-go-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/zach-snell/obsidian-go-mcp/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zach-snell/obsidian-go-mcp)](https://goreportcard.com/report/github.com/zach-snell/obsidian-go-mcp)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-zach--snell.github.io-blue)](https://zach-snell.github.io/obsidian-go-mcp/)

A fast, lightweight [MCP](https://modelcontextprotocol.io/) server for [Obsidian](https://obsidian.md/) vaults. Built in Go for speed and simplicity.

**[Documentation](https://zach-snell.github.io/obsidian-go-mcp/)** | **[Quick Start](#quick-start)** | **[Tools Reference](#tools-reference-69-total)**

## Why This Project?

| Feature | obsidian-go-mcp | Other MCP Servers |
|---------|-----------------|-------------------|
| **No plugins required** | Works directly with vault files | Often require Obsidian REST API plugin |
| **Single binary** | One file, zero dependencies | Node.js/Python runtime needed |
| **Cross-platform** | macOS, Linux, Windows | Often have platform issues |
| **69 tools** | Comprehensive vault operations | Typically 10-20 tools |
| **Fast startup** | ~10ms | Seconds for interpreted languages |

## Quick Start

**1. Install with one command:**

```bash
curl -sSL https://raw.githubusercontent.com/zach-snell/obsidian-go-mcp/main/install.sh | bash
```

This auto-detects your OS/architecture and installs to `/usr/local/bin`.

> **No sudo?** Install to `~/.local/bin` instead:
> ```bash
> curl -sSL https://raw.githubusercontent.com/zach-snell/obsidian-go-mcp/main/install.sh | bash -s -- --user
> ```

<details>
<summary>Manual download</summary>

```bash
# macOS (Apple Silicon)
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-darwin-arm64 -o obsidian-mcp && chmod +x obsidian-mcp

# macOS (Intel)
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-darwin-amd64 -o obsidian-mcp && chmod +x obsidian-mcp

# Linux
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-linux-amd64 -o obsidian-mcp && chmod +x obsidian-mcp
```
</details>

**2. Configure your MCP client:**

<details>
<summary><b>Claude Desktop</b></summary>

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obsidian-mcp",
      "args": ["/path/to/your/vault"]
    }
  }
}
```
</details>

<details>
<summary><b>Claude Code</b></summary>

The server will be auto-discovered, or add to your config:

```json
{
  "mcpServers": {
    "obsidian": {
      "command": "/path/to/obsidian-mcp",
      "args": ["/path/to/your/vault"]
    }
  }
}
```
</details>

<details>
<summary><b>Other MCP Clients</b></summary>

```bash
# Run directly (communicates via stdio)
./obsidian-mcp /path/to/vault
```
</details>

**3. Start using it!** Ask your AI assistant to search your vault, create notes, manage tasks, etc.

## Installation Options

### Pre-built Binaries (Recommended)

Download from [Releases](https://github.com/zach-snell/obsidian-go-mcp/releases/latest):

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `obsidian-mcp-darwin-arm64` |
| macOS (Intel) | `obsidian-mcp-darwin-amd64` |
| Linux (x64) | `obsidian-mcp-linux-amd64` |
| Linux (ARM) | `obsidian-mcp-linux-arm64` |
| Windows | `obsidian-mcp-windows-amd64.exe` |

### Go Install

```bash
go install github.com/zach-snell/obsidian-go-mcp/cmd/server@latest
mv $(go env GOPATH)/bin/server $(go env GOPATH)/bin/obsidian-mcp
```

### Build from Source

```bash
git clone https://github.com/zach-snell/obsidian-go-mcp.git
cd obsidian-go-mcp
go build -o obsidian-mcp ./cmd/server
```

### Upgrade

Just run the install script again - it always fetches the latest version:

```bash
curl -sSL https://raw.githubusercontent.com/zach-snell/obsidian-go-mcp/main/install.sh | bash
```

---

## Tools Reference (69 total)

### Core Operations

| Tool | Description |
|------|-------------|
| `list-notes` | List notes with pagination |
| `read-note` | Read a single note |
| `read-notes` | Read multiple notes at once |
| `write-note` | Create or update a note |
| `edit-note` | Surgical find-and-replace |
| `append-note` | Insert content with position targeting |
| `replace-section` | Replace content under a heading |
| `delete-note` | Delete a note |
| `rename-note` | Rename and update all links |
| `move-note` | Move to new location |
| `duplicate-note` | Copy a note |
| `recent-notes` | Recently modified notes |

### Search

| Tool | Description |
|------|-------------|
| `search-vault` | Full-text content search |
| `search-advanced` | Multi-term with AND/OR |
| `search-regex` | Regular expression search |
| `search-by-date` | Find by modification date |
| `search-by-tags` | Tag-based search |
| `search-headings` | Search heading text |
| `query-frontmatter` | Search YAML properties |
| `discover-mocs` | Find Maps of Content |

### Note Context

| Tool | Description |
|------|-------------|
| `get-note-summary` | Lightweight metadata + preview |
| `get-headings` | List all headings with line ranges |
| `get-section` | Extract a specific section |
| `get-frontmatter` | Get YAML frontmatter |
| `get-inline-fields` | Get Dataview fields |

### Frontmatter & Fields

| Tool | Description |
|------|-------------|
| `set-frontmatter` | Set a property |
| `remove-frontmatter-key` | Remove a property |
| `add-alias` | Add an alias |
| `add-tag-to-frontmatter` | Add a tag |
| `set-inline-field` | Set Dataview field |
| `query-inline-fields` | Query by field values |

### Graph & Links

| Tool | Description |
|------|-------------|
| `backlinks` | Notes linking to this note |
| `forward-links` | Outgoing links |
| `orphan-notes` | Unconnected notes |
| `broken-links` | Links to missing notes |
| `unlinked-mentions` | Text that could be linked |
| `suggest-links` | AI-friendly link suggestions |

### Knowledge Gaps

| Tool | Description |
|------|-------------|
| `find-stubs` | Short notes needing expansion |
| `find-outdated` | Notes not updated recently |

### Tasks

| Tool | Description |
|------|-------------|
| `list-tasks` | All tasks with metadata |
| `toggle-task` | Check/uncheck a task |

### Periodic Notes

| Tool | Description |
|------|-------------|
| `daily-note` | Get or create daily note |
| `weekly-note` | Get or create weekly note |
| `monthly-note` | Get or create monthly note |
| `quarterly-note` | Get or create quarterly note |
| `yearly-note` | Get or create yearly note |
| `list-daily-notes` | List daily notes |
| `list-periodic-notes` | List by type |

### Templates

| Tool | Description |
|------|-------------|
| `list-templates` | Available templates |
| `get-template` | View template + variables |
| `apply-template` | Create note from template |

### Organization

| Tool | Description |
|------|-------------|
| `generate-moc` | Generate Map of Content |
| `update-moc` | Add new notes to MOC |
| `generate-index` | Alphabetical index |
| `split-note` | Split at headings |
| `merge-notes` | Combine multiple notes |
| `extract-section` | Move section to new note |

### Folders

| Tool | Description |
|------|-------------|
| `list-folders` | All folders |
| `create-folder` | Create folder |
| `delete-folder` | Delete folder |

### Bulk Operations

| Tool | Description |
|------|-------------|
| `bulk-tag` | Add/remove tags from many notes |
| `bulk-move` | Move many notes |
| `bulk-set-frontmatter` | Set property on many notes |

### Canvas

| Tool | Description |
|------|-------------|
| `list-canvases` | All canvas files |
| `read-canvas` | Parse canvas content |
| `create-canvas` | New empty canvas |
| `add-canvas-node` | Add node to canvas |
| `add-canvas-edge` | Connect nodes |

### Analytics

| Tool | Description |
|------|-------------|
| `vault-stats` | Notes, words, tasks, top tags |

---

## Usage Examples

### Daily Workflow

```
"Create today's daily note and show me my open tasks"
"What did I work on last week?"
"Find notes I haven't touched in 3 months"
```

### Research & Writing

```
"Search my vault for anything about 'machine learning'"
"Find all notes tagged #project and #active"
"What notes mention 'API design' but aren't linked?"
```

### Vault Maintenance

```
"Find orphan notes with no connections"
"Show me stub notes under 100 words"
"Generate a MOC for my projects folder"
```

### Bulk Operations

```
"Add #archive tag to all notes in the old-projects folder"
"Move all notes tagged #2023 to the archive folder"
"Set status: complete on these 5 project notes"
```

---

## Template Variables

Create templates in your `templates/` folder:

```markdown
---
title: {{title}}
date: {{date}}
status: {{status:draft}}
---

# {{title}}

Created: {{datetime}}
```

### Built-in Variables

| Variable | Example |
|----------|---------|
| `{{date}}` | `2024-01-15` |
| `{{time}}` | `14:30` |
| `{{datetime}}` | `2024-01-15 14:30` |
| `{{year}}` | `2024` |
| `{{month}}` | `01` |
| `{{day}}` | `15` |
| `{{title}}` | Note title |
| `{{filename}}` | `Note.md` |
| `{{timestamp}}` | Unix timestamp |

Use `{{var:default}}` for default values.

---

## Task Format

Compatible with [Obsidian Tasks](https://publish.obsidian.md/tasks/) plugin:

```markdown
- [ ] Open task
- [x] Completed task
- [ ] Has due date 📅 2024-01-15
- [ ] High priority ⏫
- [ ] Medium priority 🔼
- [ ] Low priority 🔽
- [ ] Tagged #project #urgent
```

---

## Security

- **Path traversal protection**: All file operations are sandboxed to your vault
- **Read-only by default**: Write operations require explicit tool calls
- **No network access**: The server only accesses local files

---

## Development

```bash
# Setup (requires Go 1.21+)
git clone https://github.com/zach-snell/obsidian-go-mcp.git
cd obsidian-go-mcp

# With mise (recommended)
mise install && mise run check

# Without mise
go build -o obsidian-mcp ./cmd/server
go test -race -cover ./...
```

### Available Commands

| Command | Description |
|---------|-------------|
| `mise run build` | Build binary |
| `mise run test` | Run tests |
| `mise run lint` | Run linters |
| `mise run check` | All checks |
| `mise run fuzz` | Fuzz tests |

---

## FAQ

**Q: Do I need Obsidian running?**  
A: No. This server works directly with vault files on disk.

**Q: Will this conflict with Obsidian?**  
A: No. Both can access the same files safely.

**Q: What about sync (iCloud, Dropbox, etc)?**  
A: Works fine. The server reads/writes standard markdown files.

**Q: Can I use multiple vaults?**  
A: Run multiple server instances, each pointing to a different vault.

---

## License

Apache 2.0 - see [LICENSE](LICENSE)
