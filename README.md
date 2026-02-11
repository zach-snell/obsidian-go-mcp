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

- **CRUD Operations**: List, read, write, delete, append, move notes
- **Batch Operations**: Read multiple notes, bulk tagging, bulk moves
- **Search**: Content search, regex search, date search, tag search, frontmatter queries
- **Task Parsing**: Extract checkboxes with due dates, priorities, tags
- **Graph Analysis**: Backlinks, forward links, orphan detection, broken links, suggest links
- **Knowledge Gap Analysis**: Find stubs, outdated notes, unlinked mentions
- **Periodic Notes**: Daily, weekly, monthly, quarterly, yearly notes
- **Templates**: Create notes from templates with variable substitution
- **Canvas Support**: Read, create, and modify Obsidian canvas files
- **Folder Management**: List, create, delete folders
- **MOC Generation**: Generate and update Maps of Content
- **Note Refactoring**: Split, merge, extract sections, duplicate notes
- **Inline Fields**: Dataview-style field support (key:: value)
- **Vault Statistics**: Word counts, task completion %, top tags
- **Security**: Path traversal protection

## MCP Tools (67 total)

### Core Note Operations
| Tool | Description |
|------|-------------|
| `list-notes` | List markdown files (supports pagination) |
| `read-note` | Read note content |
| `read-notes` | Read multiple notes in one call |
| `write-note` | Create/update note |
| `append-note` | Append content to note (quick capture) |
| `delete-note` | Delete note |
| `rename-note` | Rename note and update all links |
| `move-note` | Move note to new location with link updates |
| `recent-notes` | List recently modified notes |
| `duplicate-note` | Create a copy of a note |

### Batch Operations
| Tool | Description |
|------|-------------|
| `get-note-summary` | Get lightweight summary (frontmatter, stats, preview) |
| `get-section` | Extract a specific heading section |
| `get-headings` | List all headings in a note |
| `search-headings` | Search across all headings in vault |

### Search & Discovery
| Tool | Description |
|------|-------------|
| `search-vault` | Content search |
| `search-advanced` | Multi-term search with AND/OR operators |
| `search-by-date` | Find notes by modification date range |
| `search-regex` | Search using regular expressions |
| `search-by-tags` | Tag-based search (AND) |
| `discover-mocs` | Find MOC structure |
| `query-frontmatter` | Search by YAML properties (e.g., `status=draft`) |
| `get-frontmatter` | Get frontmatter of a note |

### Frontmatter Manipulation
| Tool | Description |
|------|-------------|
| `set-frontmatter` | Set or update a frontmatter property |
| `remove-frontmatter-key` | Remove a property from frontmatter |
| `add-alias` | Add an alias to a note |
| `add-tag-to-frontmatter` | Add a tag to frontmatter tags array |

### Inline Fields (Dataview)
| Tool | Description |
|------|-------------|
| `get-inline-fields` | Extract Dataview-style fields (key:: value) |
| `set-inline-field` | Set or update an inline field |
| `query-inline-fields` | Search notes by inline field values |

### Graph Analysis
| Tool | Description |
|------|-------------|
| `backlinks` | Find notes linking TO a given note |
| `forward-links` | Show outgoing links FROM a note |
| `orphan-notes` | Find notes with no links to/from them |
| `broken-links` | Find wikilinks pointing to non-existent notes |

### Knowledge Gap Analysis
| Tool | Description |
|------|-------------|
| `find-stubs` | Find short notes that may need expansion |
| `find-outdated` | Find notes not modified recently |
| `unlinked-mentions` | Find text mentions not linked |
| `suggest-links` | Suggest notes that should be linked |

### Tasks
| Tool | Description |
|------|-------------|
| `list-tasks` | Parse checkboxes with metadata |
| `toggle-task` | Toggle task completion |

### Periodic Notes
| Tool | Description |
|------|-------------|
| `daily-note` | Get or create daily note |
| `list-daily-notes` | List daily notes |
| `weekly-note` | Get or create weekly note |
| `monthly-note` | Get or create monthly note |
| `quarterly-note` | Get or create quarterly note |
| `yearly-note` | Get or create yearly note |
| `list-periodic-notes` | List periodic notes by type |

### Templates
| Tool | Description |
|------|-------------|
| `list-templates` | List available templates |
| `get-template` | Show template content and variables |
| `apply-template` | Create note from template with variable substitution |

### MOC & Index Generation
| Tool | Description |
|------|-------------|
| `generate-moc` | Generate a Map of Content from directory |
| `update-moc` | Update existing MOC with new notes |
| `generate-index` | Generate alphabetical index of all notes |

### Note Refactoring
| Tool | Description |
|------|-------------|
| `split-note` | Split a note at heading boundaries |
| `merge-notes` | Merge multiple notes into one |
| `extract-section` | Extract a section to a new note |

### Folders
| Tool | Description |
|------|-------------|
| `list-folders` | List all folders in the vault |
| `create-folder` | Create a new folder |
| `delete-folder` | Delete a folder |

### Bulk Operations
| Tool | Description |
|------|-------------|
| `bulk-tag` | Add/remove tags from multiple notes |
| `bulk-move` | Move multiple notes to a folder |
| `bulk-set-frontmatter` | Set frontmatter property on multiple notes |

### Canvas
| Tool | Description |
|------|-------------|
| `list-canvases` | List all canvas files |
| `read-canvas` | Read and parse a canvas file |
| `create-canvas` | Create a new empty canvas |
| `add-canvas-node` | Add a node (text, file, link, group) |
| `add-canvas-edge` | Add an edge between nodes |

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
