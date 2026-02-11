# obsidian-go-mcp

A fast, lightweight MCP (Model Context Protocol) server for Obsidian vaults written in Go.

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/zach-snell/obsidian-go-mcp/cmd/server@latest
```

The binary will be installed as `server`. You may want to rename it:

```bash
mv $(go env GOPATH)/bin/server $(go env GOPATH)/bin/obsidian-mcp
```

### Option 2: Download Binary

```bash
# Linux (amd64)
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-linux-amd64 -o obsidian-mcp
chmod +x obsidian-mcp

# Linux (arm64)
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-linux-arm64 -o obsidian-mcp
chmod +x obsidian-mcp

# macOS (Apple Silicon)
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-darwin-arm64 -o obsidian-mcp
chmod +x obsidian-mcp

# macOS (Intel)
curl -sSL https://github.com/zach-snell/obsidian-go-mcp/releases/latest/download/obsidian-mcp-darwin-amd64 -o obsidian-mcp
chmod +x obsidian-mcp
```

### Option 3: Build from Source

```bash
git clone https://github.com/zach-snell/obsidian-go-mcp.git
cd obsidian-go-mcp
go build -o obsidian-mcp ./cmd/server
```

## Configuration

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

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

### OpenCode

Add to `opencode.json`:

```json
{
  "mcp": {
    "obsidian": {
      "type": "local",
      "command": ["/path/to/obsidian-mcp", "/path/to/your/vault"],
      "enabled": true
    }
  }
}
```

### Generic MCP Client

```bash
# Run the server (communicates via stdio)
./obsidian-mcp /path/to/vault
```

## Features

- **CRUD Operations**: List, read, write, delete, append notes
- **Search**: Case-insensitive content search
- **Task Parsing**: Extract checkboxes with due dates, priorities, tags
- **Tag Search**: Find notes by tags (AND operation)
- **MOC Discovery**: Find Maps of Content (#moc tagged notes)
- **Backlinks**: Find all notes linking to a given note
- **Frontmatter Queries**: Search notes by YAML properties
- **Rename with Link Updates**: Refactor notes safely
- **Daily Notes**: Get/create daily notes with configurable format
- **Templates**: Create notes from templates with variable substitution
- **Vault Statistics**: Word counts, task completion %, top tags
- **Pagination**: Limit/offset for large vaults
- **Security**: Path traversal protection

## MCP Tools (21 total)

### Core Note Operations
| Tool | Description |
|------|-------------|
| `list-notes` | List markdown files (supports pagination) |
| `read-note` | Read note content |
| `write-note` | Create/update note |
| `append-note` | Append content to note (quick capture) |
| `delete-note` | Delete note |
| `rename-note` | Rename note and update all links |
| `recent-notes` | List recently modified notes |

### Search & Discovery
| Tool | Description |
|------|-------------|
| `search-vault` | Content search |
| `search-by-tags` | Tag-based search (AND) |
| `discover-mocs` | Find MOC structure |
| `backlinks` | Find notes linking to a given note |
| `query-frontmatter` | Search by YAML properties (e.g., `status=draft`) |
| `get-frontmatter` | Get frontmatter of a note |

### Tasks
| Tool | Description |
|------|-------------|
| `list-tasks` | Parse checkboxes with metadata |
| `toggle-task` | Toggle task completion |

### Daily Notes
| Tool | Description |
|------|-------------|
| `daily-note` | Get or create today's daily note |
| `list-daily-notes` | List all daily notes |

### Templates
| Tool | Description |
|------|-------------|
| `list-templates` | List available templates |
| `get-template` | Show template content and variables |
| `apply-template` | Create note from template with variable substitution |

### Analytics
| Tool | Description |
|------|-------------|
| `vault-stats` | Vault statistics (notes, words, tasks, top tags)

## Template Variables

Templates support `{{variable}}` and `{{variable:default}}` syntax:

```markdown
---
title: {{title}}
date: {{date}}
author: {{author:Anonymous}}
---

# {{title}}

Created on {{date}} at {{time}}.
```

### Built-in Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{date}}` | Current date | `2024-01-15` |
| `{{time}}` | Current time | `14:30` |
| `{{datetime}}` | Date and time | `2024-01-15 14:30` |
| `{{year}}` | Current year | `2024` |
| `{{month}}` | Current month | `01` |
| `{{day}}` | Current day | `15` |
| `{{title}}` | Note title (without .md) | `My Note` |
| `{{filename}}` | Full filename | `My Note.md` |
| `{{folder}}` | Target folder | `notes` |
| `{{timestamp}}` | Unix timestamp | `1705332600` |

Pass custom variables: `variables="author=John,project=Alpha"`

## Task Format

Compatible with Obsidian Tasks plugin:

```markdown
- [ ] Open task
- [x] Completed task
- [ ] Due date 📅 2024-01-15
- [ ] High priority ⏫
- [ ] Medium priority 🔼
- [ ] Low priority 🔽
- [ ] With tags #project #urgent
```

## Development

Requires Go 1.21+ and [mise](https://mise.jdx.dev/) (optional but recommended).

```bash
# With mise
mise install           # Install Go + tools
mise run build         # Build binary
mise run test          # Run tests
mise run lint          # Run linters
mise run check         # All checks (lint + test + vuln)
mise run fuzz          # Run fuzz tests

# Without mise
go build -o obsidian-mcp ./cmd/server
go test -race -cover ./...
```

## License

Apache 2.0 - see [LICENSE](LICENSE)
