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
- **Pagination**: Limit/offset for large vaults
- **Security**: Path traversal protection

## MCP Tools

| Tool | Description |
|------|-------------|
| `list-notes` | List markdown files (supports pagination) |
| `read-note` | Read note content |
| `write-note` | Create/update note |
| `append-note` | Append content to note (quick capture) |
| `delete-note` | Delete note |
| `rename-note` | Rename note and update all links |
| `search-vault` | Content search |
| `list-tasks` | Parse checkboxes with metadata |
| `toggle-task` | Toggle task completion |
| `search-by-tags` | Tag-based search (AND) |
| `discover-mocs` | Find MOC structure |
| `recent-notes` | List recently modified notes |
| `backlinks` | Find notes linking to a given note |
| `query-frontmatter` | Search by YAML properties (e.g., `status=draft`) |
| `get-frontmatter` | Get frontmatter of a note |

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
